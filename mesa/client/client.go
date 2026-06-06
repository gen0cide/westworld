package mesaclient

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrOffline is returned by Cognition calls when no mesa is reachable. Callers
// treat it as "deliberate locally / fall back", never as a hard failure.
var ErrOffline = errors.New("mesa: offline")

// Client is the host's gateway to mesa's four capabilities. Implementations must
// be safe for concurrent use and honor ctx (the host imposes deadlines).
type Client interface {
	// --- Cognition (LLM inference; pearl-miss escalation) ---
	Deliberate(ctx context.Context, req *DeliberateRequest) (*Decision, error)
	Evaluate(ctx context.Context, req *EvaluateRequest) (float64, error)

	// --- Recall (RAG over long-term memory) ---
	Recall(ctx context.Context, req *RecallRequest) (*Context, error)

	// --- Mirror (async upload of local state + metrics) ---
	Mirror(ctx context.Context, batch *MirrorBatch) error

	// --- Provision (mesa→host push of compiled artifacts / cron outputs) ---
	// The returned channel is closed when the subscription ends.
	Subscribe(ctx context.Context, hostID string) (<-chan Directive, error)

	// Healthy reports reachability. When false the host runs local-only:
	// pearl/Limbic answer what they can, Recall returns empty, Mirror queues.
	Healthy() bool
}

// DeliberateRequest asks mesa to make a decision the local pearl couldn't.
// The host ships the context it already has (so mesa needn't re-derive it);
// mesa adds RAG before calling the LLM.
type DeliberateRequest struct {
	HostID   string
	Question string
	Options  []string
	Affect   *Affect           // current mood, modulates model tier/temperature
	Hints    map[string]string // optional extra context (counterparty, goal, ...)
}

// Decision is mesa's wire form of a decision (≙ brain.Decision).
type Decision struct {
	Choice     string
	Reasoning  string
	Confidence float64
	// Cacheable hints to the host's local decision cache (optional).
	CacheKey string
	CacheTTL int64 // seconds; 0 = do not cache
}

// EvaluateRequest asks mesa for a 0..1 confidence assessment of a situation.
type EvaluateRequest struct {
	HostID    string
	Situation string
}

// RecallRequest is a semantic retrieval over long-term memory.
type RecallRequest struct {
	HostID string
	Query  string
	TopK   int
}

// Context is the retrieval result.
type Context struct {
	Items []ContextItem
}

// ContextItem is one recalled chunk with provenance and relevance.
type ContextItem struct {
	Text       string
	Provenance string
	Score      float64
}

// Affect is the wire form of the host's mood vector.
type Affect struct {
	Stress     float64
	Confidence float64
	Valence    float64
}

// MirrorKind tags an uploaded update.
type MirrorKind string

const (
	MirrorEpisode      MirrorKind = "episode"      // an episodic memory
	MirrorRelationship MirrorKind = "relationship" // a trust-ledger delta/snapshot
	MirrorAffect       MirrorKind = "affect"       // an affect snapshot
	MirrorKV           MirrorKind = "kv"           // a local key/value mirror
	MirrorMetric       MirrorKind = "metric"       // an analytics counter/event
)

// Update is one item in a mirror batch.
type Update struct {
	Kind    MirrorKind
	Key     string
	Payload json.RawMessage
}

// MirrorBatch is a fire-and-forget upload of local state for one host.
type MirrorBatch struct {
	HostID  string
	Updates []Update
}

// DirectiveKind enumerates the compiled artifacts / cron outputs mesa pushes.
type DirectiveKind string

const (
	DirectivePearlRefresh      DirectiveKind = "pearl_refresh"      // new compiled policy Table
	DirectiveGoalRevision      DirectiveKind = "goal_revision"      //
	DirectiveTrustDecay        DirectiveKind = "trust_decay"        // cron output
	DirectiveReverieRebaseline DirectiveKind = "reverie_rebaseline" // cron output
	DirectivePersonaRevision   DirectiveKind = "persona_revision"   // new compiled Cornerstone
)

// Directive is one mesa→host push. ID is monotonic per host; the host applies
// strictly increasing IDs and ignores stale ones (idempotent, no CRDT).
type Directive struct {
	HostID  string
	ID      int64
	Kind    DirectiveKind
	Payload json.RawMessage
}

// --- StubClient: the offline default -----------------------------------------

// StubClient is an always-offline Client. Cognition errors with ErrOffline,
// Recall returns empty, Mirror is a no-op, Subscribe yields a closed channel.
// It lets a host run fully self-contained until a real mesa is wired.
type StubClient struct{}

func (StubClient) Deliberate(context.Context, *DeliberateRequest) (*Decision, error) {
	return nil, ErrOffline
}
func (StubClient) Evaluate(context.Context, *EvaluateRequest) (float64, error) { return 0, ErrOffline }
func (StubClient) Recall(context.Context, *RecallRequest) (*Context, error) {
	return &Context{}, nil
}
func (StubClient) Mirror(context.Context, *MirrorBatch) error { return nil }
func (StubClient) Subscribe(context.Context, string) (<-chan Directive, error) {
	ch := make(chan Directive)
	close(ch)
	return ch, nil
}
func (StubClient) Healthy() bool { return false }
