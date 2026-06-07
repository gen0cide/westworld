package main

// Minimal Anthropic Messages API client (pure net/http, no SDK) — enough to run
// the persona cook + judge. This lives mesa-side (gitignored): the host never
// makes external calls.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const anthropicURL = "https://api.anthropic.com/v1/messages"

type anthropicClient struct {
	key   string
	model string
	hc    *http.Client
}

func newAnthropic(key, model string) *anthropicClient {
	return &anthropicClient{key: key, model: model, hc: &http.Client{Timeout: 120 * time.Second}}
}

type msgRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []msgContent `json:"messages"`
	// NOTE: temperature is intentionally omitted — Opus 4.8 deprecated it and
	// controls sampling itself. Best-of-N relies on the model's default-sampling
	// variation across independent calls.
}

type msgContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type msgResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// complete sends one Messages request and returns the concatenated text.
func (c *anthropicClient) complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	reqBody, err := json.Marshal(msgRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []msgContent{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var mr msgResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return "", fmt.Errorf("decode response (status %d): %w; body: %s", resp.StatusCode, err, truncate(string(body), 300))
	}
	if mr.Error != nil {
		return "", fmt.Errorf("anthropic error (%s): %s", mr.Error.Type, mr.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic status %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
	var sb bytes.Buffer
	for _, p := range mr.Content {
		if p.Type == "text" {
			sb.WriteString(p.Text)
		}
	}
	return sb.String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
