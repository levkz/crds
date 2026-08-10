package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is anything that can answer a chat completion. The HTTP client
// implements it; tests and --dry-run substitute a fake.
type Client interface {
	// Complete sends a system+user prompt and returns the model's text reply.
	Complete(ctx context.Context, system, user string) (string, error)
}

// NewClient returns an HTTP client talking to the resolved configuration.
func NewClient(cfg Config) *HTTPClient {
	return &HTTPClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		http:    &http.Client{Timeout: 5 * time.Minute},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// HTTPClient implements Client against any OpenAI-compatible endpoint.
type HTTPClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// Complete posts a chat completion and returns the assistant's text.
func (c *HTTPClient) Complete(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("ai: marshal request: %w", err)
	}

	endpoint := c.baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: request %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("ai: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai: %s returned %d: %s", endpoint, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var parsed chatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("ai: upstream error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("ai: empty reply from %s", endpoint)
	}
	return parsed.Choices[0].Message.Content, nil
}