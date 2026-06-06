package memory

import (
	"context"
	"encoding/json"
)

// Remote is the host's remote memory tier — mesa. It is the only tier that can
// serve cross-host facts of record and semantic recall. It is an interface
// because its implementation is pluggable and, for now, deferred: a real
// mesa-backed Remote (gRPC) lands when mesa splits into its own binary.
//
// Implementations must be safe for concurrent use and must honor ctx
// (especially the read deadline the Manager imposes via the maturity dial).
type Remote interface {
	// Get fetches an exact key. ok is false on a clean miss.
	Get(ctx context.Context, key string) (json.RawMessage, bool, error)
	// Put stores an exact key.
	Put(ctx context.Context, key string, val json.RawMessage) error
	// Delete removes an exact key.
	Delete(ctx context.Context, key string) error
	// Search runs semantic recall (capability-exclusive to the remote tier).
	Search(ctx context.Context, query string, k int) ([]SearchHit, error)
	// Healthy reports whether the remote is currently reachable. The Manager
	// skips remote reads/writes when false (degraded/offline mode).
	Healthy() bool
}

// SearchHit is one semantic-recall result.
type SearchHit struct {
	Value      json.RawMessage
	Score      float64
	Provenance string // source identifier (doc/section/event id)
}

// NopRemote is the default remote: always unhealthy, every Get misses, writes
// are no-ops. It lets a host run fully offline (everything served from / kept
// in the local tiers, remote writes accumulate in the journal) until a real
// Remote is wired.
type NopRemote struct{}

func (NopRemote) Get(context.Context, string) (json.RawMessage, bool, error) { return nil, false, nil }
func (NopRemote) Put(context.Context, string, json.RawMessage) error         { return nil }
func (NopRemote) Delete(context.Context, string) error                       { return nil }
func (NopRemote) Search(context.Context, string, int) ([]SearchHit, error)   { return nil, nil }
func (NopRemote) Healthy() bool                                              { return false }
