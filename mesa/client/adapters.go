package mesaclient

import (
	"context"
	"encoding/json"

	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/memory"
)

// The adapters let one mesaclient.Client back the existing host seams unchanged.
// Act is the NEW primary call (the agent-step) and has no legacy seam — it is
// consumed by the mesa-backed Director/conductor escalation directly.

// --- brain.Strategist via Decide (the narrow in-routine choice) --------------

// AsStrategist returns a brain.Strategist that routes decide() to mesa.Decide.
func AsStrategist(c Client, hostID string) brain.Strategist {
	return &strategistAdapter{c: c, hostID: hostID}
}

type strategistAdapter struct {
	c      Client
	hostID string
}

func (a *strategistAdapter) Decide(ctx context.Context, s brain.Situation) (*brain.Decision, error) {
	d, err := a.c.Decide(ctx, &Choice{HostID: a.hostID, Question: s.Question, Options: s.Options})
	if err != nil {
		return nil, err
	}
	return &brain.Decision{Choice: d.Choice, Reasoning: d.Reasoning, Confidence: d.Confidence}, nil
}

// --- cognition.Client via Recall --------------------------------------------

// AsRetriever returns a cognition.Client that pulls game knowledge from mesa.
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
	k, err := a.c.Recall(ctx, &Query{HostID: a.hostID, Text: r.Goal, Kind: KnowAny, TopK: top})
	if err != nil {
		return nil, err
	}
	b := &cognition.Bundle{Goal: r.Goal}
	for _, it := range k.Items {
		b.Episodic = append(b.Episodic, it.Text)
	}
	return b, nil
}

// --- memory.Remote via Recall (Search) + Remember (Put) ---------------------

// AsRemote returns a memory.Remote backed by mesa: Search → Recall, writes →
// Remember (as a KV-mirror episode). Exact-key Get folds into Recall.
func AsRemote(c Client, hostID string) memory.Remote {
	return &remoteAdapter{c: c, hostID: hostID}
}

type remoteAdapter struct {
	c      Client
	hostID string
}

func (a *remoteAdapter) Get(ctx context.Context, key string) (json.RawMessage, bool, error) {
	v, found, err := a.c.GetKV(ctx, a.hostID, key)
	if err != nil || !found {
		return nil, false, err
	}
	return json.RawMessage(v), true, nil
}

func (a *remoteAdapter) Put(ctx context.Context, key string, val json.RawMessage) error {
	return a.c.PutKV(ctx, a.hostID, key, []byte(val))
}

func (a *remoteAdapter) Delete(ctx context.Context, key string) error {
	return a.c.DeleteKV(ctx, a.hostID, key)
}

func (a *remoteAdapter) Search(ctx context.Context, query string, k int) ([]memory.SearchHit, error) {
	res, err := a.c.Recall(ctx, &Query{HostID: a.hostID, Text: query, Kind: KnowAny, TopK: k})
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
