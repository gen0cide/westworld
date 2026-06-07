package main

// Minimal Voyage AI embeddings client (pure net/http). Out-of-band tooling —
// holds the key, runs offline, NOT on any host path.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const voyageURL = "https://api.voyageai.com/v1/embeddings"

type voyageClient struct {
	key   string
	model string
	hc    *http.Client
}

func newVoyage(key, model string) *voyageClient {
	return &voyageClient{key: key, model: model, hc: &http.Client{Timeout: 180 * time.Second}}
}

type embedReq struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type"` // "document" | "query"
}

type embedResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Detail string `json:"detail"` // set on error
}

// embed returns one vector per input text (one request; caller batches).
func (v *voyageClient) embed(ctx context.Context, texts []string, inputType string) ([][]float32, int, error) {
	body, err := json.Marshal(embedReq{Input: texts, Model: v.model, InputType: inputType})
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+v.key)
	req.Header.Set("content-type", "application/json")

	resp, err := v.hc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var er embedResp
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, 0, fmt.Errorf("voyage decode (status %d): %w; body: %s", resp.StatusCode, err, trunc(string(raw), 300))
	}
	if resp.StatusCode != http.StatusOK || er.Detail != "" {
		return nil, 0, fmt.Errorf("voyage status %d: %s", resp.StatusCode, firstNonEmpty(er.Detail, trunc(string(raw), 300)))
	}
	out := make([][]float32, len(er.Data))
	for i, d := range er.Data {
		out[i] = d.Embedding
	}
	return out, er.Usage.TotalTokens, nil
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
