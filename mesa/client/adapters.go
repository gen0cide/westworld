package mesaclient

import (
	"context"
	"encoding/json"

	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/memory"
)

// The adapters let one mesa.Client back the three existing host seams unchanged,
// so wiring is just `host.Strategist = mesa.AsStrategist(c, hostID)` etc. Each
// adapter binds a HostID (the mux key) and translates the seam's domain types to
// mesa's wire types.

// --- brain.Strategist via Cognition.Deliberate ------------------------------

// AsStrategist returns a brain.Strategist that escalates decisions to mesa.
func AsStrategist(c Client, hostID string) brain.Strategist {
	return &strategistAdapter{c: c, hostID: hostID}
}

type strategistAdapter struct {
	c      Client
	hostID string
}

func (a *strategistAdapter) Decide(ctx context.Context, s brain.Situation) (*brain.Decision, error) {
	req := &DeliberateRequest{HostID: a.hostID, Question: s.Question, Options: s.Options}
	d, err := a.c.Deliberate(ctx, req)
	if err != nil {
		return nil, err
	}
	return &brain.Decision{Choice: d.Choice, Reasoning: d.Reasoning, Confidence: d.Confidence}, nil
}

// --- cognition.Client via Recall --------------------------------------------

// AsRetriever returns a cognition.Client that pulls context from mesa's RAG.
func AsRetriever(c Client, hostID string) cognition.Client {
	return &retrieverAdapter{c: c, hostID: hostID}
}

type retrieverAdapter struct {
	c      Client
	hostID string
}

func (a *retrieverAdapter) Retrieve(ctx context.Context, r cognition.Retrieval) (*cognition.Bundle, error) {
	top := r.MaxItems
	if top <= 0 {
		top = 5
	}
	res, err := a.c.Recall(ctx, &RecallRequest{HostID: a.hostID, Query: r.Goal, TopK: top})
	if err != nil {
		return nil, err
	}
	b := &cognition.Bundle{Goal: r.Goal}
	for _, it := range res.Items {
		b.Episodic = append(b.Episodic, it.Text)
	}
	return b, nil
}

// --- memory.Remote via Recall (Search) + Mirror (Put) -----------------------

// AsRemote returns a memory.Remote backed by mesa: Search → Recall, writes →
// Mirror. Exact-key Get is a recall miss for now (cross-host facts are modeled
// as recall, not generic KV — see the transport-doc reconciliation note).
func AsRemote(c Client, hostID string) memory.Remote {
	return &remoteAdapter{c: c, hostID: hostID}
}

type remoteAdapter struct {
	c      Client
	hostID string
}

func (a *remoteAdapter) Get(context.Context, string) (json.RawMessage, bool, error) {
	return nil, false, nil // exact-key cross-host reads fold into Search/Recall
}

func (a *remoteAdapter) Put(ctx context.Context, key string, val json.RawMessage) error {
	return a.c.Mirror(ctx, &MirrorBatch{
		HostID:  a.hostID,
		Updates: []Update{{Kind: MirrorKV, Key: key, Payload: val}},
	})
}

func (a *remoteAdapter) Delete(context.Context, string) error { return nil }

func (a *remoteAdapter) Search(ctx context.Context, query string, k int) ([]memory.SearchHit, error) {
	res, err := a.c.Recall(ctx, &RecallRequest{HostID: a.hostID, Query: query, TopK: k})
	if err != nil || res == nil {
		return nil, err
	}
	hits := make([]memory.SearchHit, 0, len(res.Items))
	for _, it := range res.Items {
		raw, _ := json.Marshal(it.Text)
		hits = append(hits, memory.SearchHit{Value: raw, Score: it.Score, Provenance: it.Provenance})
	}
	return hits, nil
}

func (a *remoteAdapter) Healthy() bool { return a.c.Healthy() }
