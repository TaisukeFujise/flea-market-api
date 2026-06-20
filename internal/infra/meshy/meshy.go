package meshy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.meshy.ai"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type createJobRequest struct {
	ImageURLs     []string `json:"image_urls"`
	AIModel       string   `json:"ai_model"`
	ShouldTexture bool     `json:"should_texture"`
	EnablePBR     bool     `json:"enable_pbr"`
	HDTexture     bool     `json:"hd_texture"`
	ShouldRemesh  bool     `json:"should_remesh"`
	AutoSize      bool     `json:"auto_size"`
	OriginAt      string   `json:"origin_at"`
	TargetFormats []string `json:"target_formats"`
}

type createJobResponse struct {
	Result string `json:"result"`
}

type jobStatusResponse struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Progress  int               `json:"progress"`
	ModelURLs map[string]string `json:"model_urls"`
}

func (c *Client) CreateJob(ctx context.Context, imageURLs []string) (string, error) {
	body, err := json.Marshal(createJobRequest{
		ImageURLs:     imageURLs,
		AIModel:       "meshy-6",
		ShouldTexture: true,
		EnablePBR:     true,
		HDTexture:     true,
		ShouldRemesh:  true,
		AutoSize:      true,
		OriginAt:      "bottom",
		TargetFormats: []string{"glb"},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/multi-image-to-3d", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("meshy create job: status=%d body=%s", resp.StatusCode, b)
	}

	var result createJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Result, nil
}

func (c *Client) GetJobStatus(ctx context.Context, jobID string) (status string, glbURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/multi-image-to-3d/"+jobID, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("meshy get job: status=%d body=%s", resp.StatusCode, b)
	}

	var result jobStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	return result.Status, result.ModelURLs["glb"], nil
}

func (c *Client) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("meshy download: status=%d url=%s", resp.StatusCode, url)
	}
	return resp.Body, nil
}
