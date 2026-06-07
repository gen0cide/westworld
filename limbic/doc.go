// Package limbic holds the state the host's System-1 affect path maintains —
// the deterministic, no-LLM "limbic system" that the runtime's Limbic goroutine
// (a bus subscriber, sibling to heartbeatLoop) drives from game events.
//
// Two models, both pure and concurrency-safe:
//
//   - Affect: the mood vector (stress, confidence, valence) that events nudge
//     and that decays toward a baseline on read (lazy decay). It feeds the
//     pearl engine's affect predicates.
//
//   - Ledger: a Beta(α,β) trust ledger keyed by counterparty. Positive and
//     negative interactions accrue evidence; trust is the Beta mean mapped to
//     [-1,1], familiarity is the encounter count. It backs the relation_with
//     DSL verb and the pearl engine's relationship predicates.
//
// The package is pure data + update rules (no bus, no world, no I/O) so it is
// trivially testable; the runtime owns the goroutine that routes events into it
// and the persistence that flushes it through the memory layer.
package limbic
