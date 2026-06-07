// Package llm is mesa's Anthropic Messages client — a thin, SDK-free wrapper
// used by the mesa service for its LLM seams (Decide/Act/cook). It lives
// mesa-side (gitignored): the host never imports it and never makes external
// calls. The host holds no keys; mesa does.
package llm

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

// DefaultModel is the model mesa reaches for when none is configured.
const DefaultModel = "claude-opus-4-8"

// Client is a minimal Anthropic Messages client.
type Client struct {
	key   string
	model string
	hc    *http.Client
}

// New builds a client. An empty model falls back to DefaultModel.
func New(key, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{key: key, model: model, hc: &http.Client{Timeout: 120 * time.Second}}
}

// Model returns the configured model id.
func (c *Client) Model() string { return c.model }

type msgRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    any          `json:"system,omitempty"` // string OR []sysBlock (for cache_control)
	Messages  []msgContent `json:"messages"`
	// temperature is intentionally omitted — Opus 4.8 deprecated it and controls
	// sampling itself.
}

// SystemBlock is one segment of a structured system prompt. Cache marks it with
// an ephemeral cache_control breakpoint so Anthropic prompt-caches the prefix —
// the right home for a large, static manual reused across many calls.
type SystemBlock struct {
	Text  string
	Cache bool
}

type sysBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}

type cacheControl struct {
	Type string `json:"type"` // "ephemeral"
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

// Complete sends one Messages request with a plain system string.
func (c *Client) Complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	var sys any
	if system != "" {
		sys = system
	}
	return c.complete(ctx, sys, user, maxTokens)
}

// CompleteSystem sends one Messages request with a structured (optionally
// cached) system prompt. Blocks marked Cache get an ephemeral cache_control
// breakpoint so the prefix is prompt-cached across calls.
func (c *Client) CompleteSystem(ctx context.Context, blocks []SystemBlock, user string, maxTokens int) (string, error) {
	sys := make([]sysBlock, 0, len(blocks))
	for _, b := range blocks {
		sb := sysBlock{Type: "text", Text: b.Text}
		if b.Cache {
			sb.CacheControl = &cacheControl{Type: "ephemeral"}
		}
		sys = append(sys, sb)
	}
	return c.complete(ctx, sys, user, maxTokens)
}

// complete sends one Messages request (system is a string or []sysBlock) and
// returns the concatenated text.
func (c *Client) complete(ctx context.Context, system any, user string, maxTokens int) (string, error) {
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
		return "", fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if mr.Error != nil {
		return "", fmt.Errorf("anthropic error (%s): %s", mr.Error.Type, mr.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic status %d", resp.StatusCode)
	}
	var sb bytes.Buffer
	for _, p := range mr.Content {
		if p.Type == "text" {
			sb.WriteString(p.Text)
		}
	}
	return sb.String(), nil
}
