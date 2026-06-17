package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/TaisukeFujise/flea-market-api/internal/service"
	"google.golang.org/genai"
)

const geminiModelName = "gemini-2.0-flash-001"

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

const systemInstruction = `あなたは中古品フリマアプリの商品査定AIです。
商品の複数方向の画像を分析し、傷や汚れを検出してください。
各画像の直後に撮影方向（front/back/right/left/top）を示すテキストが続きます。

以下のJSON形式のみで返答してください。説明文、Markdown、コードブロックは禁止です。

{
  "condition": "good" または "fair" または "poor",
  "condition_note": "全体的な商品状態の日本語説明（1〜2文）",
  "damages": [
    {
      "image_angle": "front" または "back" または "right" または "left" または "top",
      "damage_type": "scratch" または "dirt" または "wear",
      "bbox_x1": 0,
      "bbox_y1": 0,
      "bbox_x2": 1000,
      "bbox_y2": 1000,
      "description": "傷の説明（日本語）"
    }
  ]
}

condition の基準：
- good: 傷や汚れがほとんどなく状態が良い
- fair: 使用感があり軽微な傷や汚れがある
- poor: 目立つ傷・汚れ・破損がある

bbox は画像左上を(0,0)・右下を(1000,1000)とした正規化座標で指定してください。
傷がない場合は damages を空配列にしてください。`

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
	if raw == "" {
		return service.DamageDetectionResult{}, fmt.Errorf("empty response from Gemini")
	}

	var dr detectionResponse
	if err := json.Unmarshal([]byte(raw), &dr); err != nil {
		return service.DamageDetectionResult{}, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

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
