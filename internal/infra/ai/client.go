package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"google.golang.org/genai"
)

const geminiModelName = "gemini-2.5-flash"

type VertexAIClient struct {
	client    *genai.Client
	gcsBucket string
}

func NewVertexAIClient(ctx context.Context) (*VertexAIClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: os.Getenv("VERTEX_AI_LOCATION"),
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("genai.NewClient: %w", err)
	}
	bucket := os.Getenv("GCS_BUCKET_NAME")
	if bucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME is not set")
	}
	return &VertexAIClient{client: client, gcsBucket: bucket}, nil
}

const systemInstruction = `
あなたは中古品フリマアプリの商品査定AIです。
商品の複数方向の画像を分析し、傷や汚れを検出してください。
各画像の直前に撮影方向（front/back/right/left/top）を示すテキストが続きます。

出力は response schema に完全に従ってください。
説明文、Markdown、コードブロック、schema外のフィールドは禁止です。

condition の基準：
- good: 傷や汚れがほとんどなく状態が良い
- fair: 使用感があり軽微な傷や汚れがある
- poor: 目立つ傷・汚れ・破損がある

damage_type の基準：
- scratch: 線状の傷、削れ、ひっかき傷
- dirt: 汚れ、シミ、変色、付着物
- wear: 使用に伴う擦れ、角スレ、表面の摩耗

bbox は以下のルールで指定してください：
- 対象の傷・汚れ・使用感が完全に含まれる最小の矩形にしてください
- 不確かな場合は、対象を少し広めに含めてください
- bbox_x1 < bbox_x2、bbox_y1 < bbox_y2 を必ず満たしてください
- 画像左上を(0,0)・右下を(1000,1000)とした正規化座標で、[0,1000] の整数で返してください
- 同じ損傷を複数回返さないでください。

傷がない場合は damages を空配列にしてください。
condition_noteは日本語で1~2文にしてください。
`

var detectionResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"condition": {
			Type: genai.TypeString,
			Enum: []string{"good", "fair", "poor"},
		},
		"condition_note": {
			Type: genai.TypeString,
		},
		"damages": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"image_angle": {
						Type: genai.TypeString,
						Enum: []string{"front", "back", "right", "left", "top"},
					},
					"damage_type": {
						Type: genai.TypeString,
						Enum: []string{"scratch", "dirt", "wear"},
					},
					"bbox_x1":     {Type: genai.TypeInteger},
					"bbox_y1":     {Type: genai.TypeInteger},
					"bbox_x2":     {Type: genai.TypeInteger},
					"bbox_y2":     {Type: genai.TypeInteger},
					"description": {Type: genai.TypeString},
				},
				Required: []string{
					"image_angle",
					"damage_type",
					"bbox_x1",
					"bbox_y1",
					"bbox_x2",
					"bbox_y2",
					"description",
				},
			},
		},
	},
	Required: []string{
		"condition",
		"condition_note",
		"damages",
	},
}

type detectionResponse struct {
	Condition     string           `json:"condition"`
	ConditionNote string           `json:"condition_note"`
	Damages       []detectedDamage `json:"damages"`
}

type detectedDamage struct {
	ImageAngle  string  `json:"image_angle"`
	DamageType  string  `json:"damage_type"`
	BboxX1      *int    `json:"bbox_x1"`
	BboxY1      *int    `json:"bbox_y1"`
	BboxX2      *int    `json:"bbox_x2"`
	BboxY2      *int    `json:"bbox_y2"`
	Description *string `json:"description"`
}

var validConditions = map[string]bool{
	string(domain.ConditionGood): true,
	string(domain.ConditionFair): true,
	string(domain.ConditionPoor): true,
}

var validDamageTypes = map[string]bool{
	string(domain.DamageTypeScratch): true,
	string(domain.DamageTypeDirt):    true,
	string(domain.DamageTypeWear):    true,
}

func (c *VertexAIClient) Detect(ctx context.Context, inputs []service.DetectorInput) (service.DamageDetectionResult, error) {
	parts := make([]*genai.Part, 0, len(inputs)*2+1)
	parts = append(parts, genai.NewPartFromText("以下の商品画像を査定してください"))
	for _, img := range inputs {
		parts = append(parts,
			genai.NewPartFromText(fmt.Sprintf("撮影方向: %s", string(img.Angle))),
			genai.NewPartFromURI(fmt.Sprintf("gs://%s/%s", c.gcsBucket, img.GCSName), img.ContentType),
		)
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
		ResponseMIMEType:  "application/json",
		ResponseSchema:    detectionResponseSchema,
		CandidateCount:    1,
	}

	result, err := c.client.Models.GenerateContent(ctx, geminiModelName, contents, config)
	if err != nil {
		return service.DamageDetectionResult{}, fmt.Errorf("GenerateContent: %w", err)
	}
	if result.PromptFeedback != nil && result.PromptFeedback.BlockReason != genai.BlockedReasonUnspecified {
		return service.DamageDetectionResult{}, fmt.Errorf("prompt blocked: %v", result.PromptFeedback.BlockReason)
	}
	raw := strings.TrimSpace(result.Text())
	slog.Info("Gemini raw response", "response", raw)
	if raw == "" {
		return service.DamageDetectionResult{}, fmt.Errorf("empty response from Gemini")
	}

	var dr detectionResponse
	if err := json.Unmarshal([]byte(raw), &dr); err != nil {
		slog.Error("failed to parse Gemini response", "raw", raw, "error", err)
		return service.DamageDetectionResult{}, fmt.Errorf("failed to parse Gemini response: %w", err)
	}
	slog.Info("Gemini parsed response", "condition", dr.Condition, "condition_note", dr.ConditionNote, "damage_count", len(dr.Damages))

	if !validConditions[dr.Condition] {
		dr.Condition = string(domain.ConditionFair)
	}

	angleToImageID := make(map[domain.ImageAngle]string, len(inputs))
	for _, img := range inputs {
		angleToImageID[img.Angle] = img.ImageID
	}

	damages := make([]domain.DamageCreate, 0, len(dr.Damages))
	for _, dmg := range dr.Damages {
		if !validDamageTypes[dmg.DamageType] {
			continue
		}
		imageID, ok := angleToImageID[domain.ImageAngle(dmg.ImageAngle)]
		if !ok {
			continue
		}
		damages = append(damages, domain.DamageCreate{
			ImageID:     imageID,
			DamageType:  domain.DamageType(dmg.DamageType),
			BboxX1:      dmg.BboxX1,
			BboxY1:      dmg.BboxY1,
			BboxX2:      dmg.BboxX2,
			BboxY2:      dmg.BboxY2,
			Description: dmg.Description,
		})
	}

	return service.DamageDetectionResult{
		Condition:     domain.ProductCondition(dr.Condition),
		ConditionNote: dr.ConditionNote,
		Damages:       damages,
	}, nil
}
