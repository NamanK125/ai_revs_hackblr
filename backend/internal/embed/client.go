// Package embed provides an HTTP client for the FastEmbed sidecar service.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is an HTTP wrapper for the Python FastEmbed sidecar.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new embedder client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// embedRequest is the request payload sent to the embedder.
type embedRequest struct {
	Texts []string `json:"texts"`
}

// embedResponse is the response payload from the embedder.
type embedResponse struct {
	Vectors [][]float32 `json:"vectors"`
}

// Embed sends texts to the sidecar for embedding.
// Returns a list of float32 vectors, one per text.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embedRequest{
		Texts: texts,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed returned status %d", resp.StatusCode)
	}

	var respBody embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}

	return respBody.Vectors, nil
}

// Health checks the embedder sidecar health endpoint.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health returned status %d", resp.StatusCode)
	}

	return nil
}