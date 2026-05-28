package cognition

import (
	"context"
	"time"
)

// Bundle is a packaged context retrieval ready to feed to the brain.
// It is the cognition layer's product: everything the strategist
// needs to reason about a question, pre-assembled in one struct.
//
// Field semantics:
//
//   - Goal: the question or task being reasoned about. Echoes the
//     Retrieval.Goal that produced this Bundle so the brain can see
//     what the cognition layer thought it was retrieving for.
//   - Reflections: long-term reflective memories — generalizations
//     the host has formed over its lifetime. Mesa stores these as
//     small, mostly-cached chunks; in the stub they're canned
//     sentences keyed off goal heuristics.
//   - Episodic: recent specific memories — concrete past events the
//     host remembers. In the real client these come from mesa's
//     vector store, ranked by salience + similarity to Goal.
//   - Persona: shallow snapshot of persona traits relevant to this
//     decision. Strings (not typed) because the brain prompts treat
//     persona as opaque labels.
//   - WorldSnapshot: a serialized, observable world summary. In the
//     real client this is built by snapshotting world.World; in the
//     stub it is a canned sentence so tests stay deterministic.
//   - Timestamp: when the Bundle was assembled. Useful for staleness
//     checks and observability (delos chain-of-thought logs).
type Bundle struct {
	Goal          string
	Reflections   []string
	Episodic      []string
	Persona       map[string]string
	WorldSnapshot string
	Timestamp     time.Time
}

// Retrieval describes what we want the cognition layer to fetch.
// Callers build one of these per brain consultation; the cognition
// client returns a Bundle.
//
// Field semantics:
//
//   - Goal: the question being asked. The retrieval client uses this
//     as the vector-similarity query for episodic/reflective memory,
//     and threads it back into the resulting Bundle.Goal so the
//     brain sees what it was retrieved against.
//   - HostName: identifies which host's memory to retrieve from.
//     In a multi-host process, each host has its own mesa partition.
//     Empty means "no host context" — useful for cradle-wide tests.
//   - MaxItems: soft upper bound on Reflections + Episodic items
//     returned. Zero means "use the client's default" (the stub
//     uses 3). Real client honors this against its top-K parameter.
//   - IncludeWorld: when true, the client should populate
//     WorldSnapshot. Setting this to false is a cheap signal that
//     the caller already has the world state and just needs memory.
type Retrieval struct {
	Goal         string
	HostName     string
	MaxItems     int
	IncludeWorld bool
}

// Client packages memory + context for the brain. All implementations
// are expected to be safe for concurrent use across multiple hosts.
//
// The real Phase-3 implementation will be backed by mesa over gRPC.
// The stub returns deterministic hand-crafted Bundles for tests +
// early integration.
type Client interface {
	// Retrieve assembles a Bundle answering the given Retrieval.
	// Implementations should honor ctx cancellation; the stub
	// honors it as a courtesy but performs no I/O.
	Retrieve(ctx context.Context, r Retrieval) (*Bundle, error)
}
