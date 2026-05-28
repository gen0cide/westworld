// Package cognition is the host's retrieval-orchestration layer: it
// answers "what data does the brain need to think about this?" by
// assembling a Bundle from a host's memory + observable state.
//
// In the eventual real implementation (Phase 3), cognition talks to
// mesa over gRPC — vector-similarity queries against the host's
// episodic memory, batch lookups for relational records, structured
// fact lookups against OpenRSC defs already loaded in facts.Facts.
// The brain layer above never sees those access patterns directly;
// it just consumes Bundles.
//
// This package currently ships a deterministic StubClient that
// returns hand-crafted Bundles. The stub lets the rest of the system
// — runtime, brain stubs, routine tests — wire to a stable
// cognition.Client interface today, before mesa retrieval lands.
//
// # Dependency direction
//
// cognition is a LEAF package — it imports only stdlib. It does NOT
// import brain (brain depends on cognition for the Bundle type, not
// the other way around). It does NOT import runtime, dsl, mesa, or
// facts: those wirings happen at the call site so cognition stays
// trivially testable.
//
// # Conventions
//
//   - All public entry points take a context.Context as the first
//     argument, even where the stub ignores it. The real client
//     issues network calls, so the contract reserves cancellation
//     from day one.
//   - The stub is stateless and goroutine-safe. The real client will
//     also be expected to be safe for concurrent Retrieve calls from
//     many hosts in one process.
//   - Errors are returned, not panicked. The stub presently has no
//     failure modes, but the signature reserves error for the real
//     client (network down, mesa rejected the query, etc.).
package cognition
