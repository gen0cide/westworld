// Package mesaclient is the host-side gateway to everything that is NOT
// compute-local-feasible. A host is an isolated compute unit; for its higher
// cognitive functions it phones mesa, which owns LLM inference, RAG,
// long-term memory, the memory crons, and persona/policy compilation. The host
// holds no keys and makes no external calls — it only talks to mesa.
//
// The surface is four capabilities:
//
//   - Cognition  (req/resp): Deliberate / Evaluate — LLM inference, the
//     escalation target when the local pearl fast path misses.
//   - Recall     (req/resp): semantic retrieval over long-term memory (RAG).
//   - Mirror     (async):    upload local state (episodes, trust deltas, affect,
//     KV mirror) and metrics for durability + the crons.
//   - Provision  (push):     mesa→host stream of compiled artifacts and cron
//     outputs (Cornerstone, pearl Table, goal/persona revisions, decay).
//
// Each isolated host owns its OWN Client + connection to mesa and authenticates
// as its own host_id — the host is self-contained, mesa link included. The
// cradle is only a lifecycle manager (spawn/supervise/share static assets), not
// a connection aggregator, so a standalone cmd/host and a cradle-hosted host
// take the identical host→mesa path. This package is the contract + a StubClient
// (offline: everything degrades to local-only) + adapters that back the existing
// brain.Strategist / cognition.Client / memory.Remote seams unchanged. The real
// gRPC implementation lands when mesa splits to its own binary; until then the
// stub keeps the host fully self-contained.
package mesaclient
