// Package embed is mesa's text-embedding client (Voyage AI). It is mesa-side
// only — it holds the VOYAGE_AI_KEY and makes the external call a host never
// does. mesad uses it to embed episodes on write and queries on recall, ranking
// long-term memory by pgvector cosine similarity.
package embed

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

// DefaultModel is Voyage's general-purpose embedding model; 1024 dimensions.
const DefaultModel = "voyage-3"

// dims maps known models to their output dimensionality (for the pgvector
// column). Unknown models fall back to DefaultDim.
var dims = map[string]int{
	"voyage-3":       1024,
	"voyage-3-large": 1024,
	"voyage-3.5":     1024,
	"voyage-3-lite":  512,
}

// DefaultDim is the assumed dimensionality when a model isn't in the table.
const DefaultDim = 1024

// Voyage is a minimal embeddings client (pure net/http).
type Voyage struct {
	key   string
	model string
	hc    *http.Client
}

// NewVoyage builds a client for model (empty → DefaultModel).
func NewVoyage(key, model string) *Voyage {
	if model == "" {
		model = DefaultModel
	}
	return &Voyage{key: key, model: model, hc: &http.Client{Timeout: 30 * time.Second}}
}

// Model returns the configured model name.
func (v *Voyage) Model() string { return v.model }

// Dim returns the embedding dimensionality for the configured model.
func (v *Voyage) Dim() int {
	if d, ok := dims[v.model]; ok {
		return d
	}
	return DefaultDim
}

type embedReq struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type"` // "document" (stored) | "query" (search)
}

type embedResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Detail string `json:"detail"` // set on error
}

// Embed returns one vector per input text. inputType is "document" for stored
// content and "query" for a search string (Voyage tunes each asymmetrically).
func (v *Voyage) Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	body, err := json.Marshal(embedReq{Input: texts, Model: v.model, InputType: inputType})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+v.key)
	req.Header.Set("content-type", "application/json")

	resp, err := v.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var er embedResp
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, fmt.Errorf("voyage decode (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || er.Detail != "" {
		detail := er.Detail
		if detail == "" {
			detail = trunc(string(raw), 200)
		}
		return nil, fmt.Errorf("voyage status %d: %s", resp.StatusCode, detail)
	}
	out := make([][]float32, len(er.Data))
	for i, d := range er.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
