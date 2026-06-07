// Package memory is the host's tiered memory manager — the intelligent storage
// layer behind the location-blind DSL remember/recall verbs.
//
// A routine never knows whether a fact lives in RAM, on local disk, or in mesa.
// The Manager decides, per operation, using a policy that combines a
// per-namespace table, optional per-call hints, and runtime telemetry. It is
// NOT a dumb "try local then remote" fallback: a tier is consulted for one of
// two reasons —
//
//   - it is a faster COPY of the same data (cache logic: read-through cascade,
//     write-back, promote-on-hit, negative caching, per-class authority); or
//   - it is the only place that can ANSWER (router logic: semantic recall and
//     cross-host facts live only in the remote tier).
//
// Tiers (fastest → most capable):
//
//	L0 scratch  hostkv.Scratch  µs    ephemeral working memory
//	L1 local    hostkv.Store    µs    this-host private durable (bbolt)
//	L2 remote   Remote (mesa)   100ms+ facts of record, cross-host, semantic
//
// Scratch and local are used directly; only the remote tier is abstracted
// (Remote interface) because its implementation (mesa) is pluggable and, for
// now, deferred — NopRemote stands in (always unhealthy, always misses), so the
// host runs fully offline until a real remote is wired.
//
// Remote reads obey a MATURITY DIAL: a newborn host (empty caches, no compiled
// intuition) phones home for almost everything and waits patiently (long sync
// deadline); as its local tiers fill, the deadline tightens and reads can flip
// to async. See SetMaturity / Class.RemoteDeadline.
package memory
