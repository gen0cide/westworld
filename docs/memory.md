# Memory

How a host remembers — the real, shipped system. (The original design exploration is
archived at [`archive/initial-brainstorming/memory-original-design.md`](archive/initial-brainstorming/memory-original-design.md);
open items live in [`TODO.md`](TODO.md).)

## Why memory matters for the research

A host that cannot remember cannot learn, and a host that remembers everything is a
database, not a character. The research bet is **selective retention**: what a host
keeps — and what it lets fade — is part of its personality and its believability.
Memory is therefore not one store but a stack of mechanisms with different
retention/recall characters, and the interesting research surface is the policy
layer that decides what goes where.

## The layers (what actually exists)

### 1. The tiered key/value manager — `memory/`

`memory.Manager` sits behind the DSL's location-blind `remember`/`recollect`/`forget`
verbs. A routine never knows where a fact lives; the Manager routes per operation
using a per-namespace policy table + per-call hints + runtime telemetry
(`memory/policy.go`, `memory/manager.go`).

| tier | backing | latency | role |
|---|---|---|---|
| L0 scratch | `hostkv` in-memory | µs | ephemeral working memory |
| L1 local | `hostkv.Store` (pebble) | µs | this-host private durable state |
| L2 remote | `Remote` interface → mesa | 100ms+ | facts of record, cross-host, semantic recall |

Two distinct reasons to consult a tier: as a faster **copy** (read-through cascade,
write-back, promote-on-hit, negative caching) or because it is the only tier that can
**answer** (semantic recall, cross-host facts → remote only). Remote reads obey a
**maturity dial** (`Manager.SetMaturity`): a newborn host phones home for nearly
everything with patient deadlines; as local tiers fill, deadlines tighten and reads
flip async.

A **write-back journal** (`memory/journal.go`) queues remote-bound writes while mesa
is unreachable and survives restarts in the local store. *Known gap: nothing calls
`Manager.Flush` in production yet and same-key entries are not compacted — tracked
in TODO.md (refactor finding R2).*

### 2. Durable host state — `hostkv/`

One pebble-backed store per host (`~/.westworld/hosts/<name>/<name>.db.pebble`),
flat string keys namespaced by convention (`ledger:`, `goal:`, `knowledge:`), raw
JSON values, typed access via generic `Get`/`Set`. Chosen over bbolt for
write-heavy host workloads (chain-of-thought, ledgers: ~1µs vs ~8ms per write) and
over badger for per-instance footprint (0.2MB vs 85MB at rest). Legacy bbolt files
auto-migrate on first open. Safe against use-after-close (one host's teardown must
never panic the fleet process).

### 3. The episodic journal — `runtime/memory_journal.go`

A bus-subscriber goroutine folds **salient events** into a durable episode journal
(combat, trades, deaths, notable conversations — kind-graded importance, not a raw
event log). Episodes are what genesis reads at login ("memories of a past life")
and what mesa's consolidation crons digest.

### 4. The trust ledger — `limbic/ledger.go`

Relationship memory: a **Beta(α,β) ledger** keyed by counterparty. The prior is
uniform (α=β=1 ⇒ neutral trust for a never-met party); interactions accrue evidence;
**trust** is the Beta mean mapped to [-1,1] (an honesty hit-rate: "will this trade
be honest?"), while **affinity** and **grievance** accumulate separately (liking and
grudges are not probabilities). Persisted via the host store; surfaced to cognition
and the cradle Mind inspector.

### 5. Knowledge & goals — `cognition/knowledge`, `cognition/goalgraph`

Not "memory" in the episodic sense but persisted the same way: the knowledge ledger
(what the host believes, with confidence and provenance) and the goal graph (what it
wants, with open questions) — both serialized into the host store and mirrored to
mesa as observations. See `docs/world-knowledge-and-learning.md`.

### 6. The mesa side — `mesa/mesad/ltm.go`, `cron.go`, `cron_insight.go`

Long-term memory of record (Postgres). Hosts stream observations up; **consolidation
crons** digest them in batches (Haiku-tier calls, ~60s cadence) into beliefs and
relationship summaries; an **insight cron** (~3min) produces higher-tier inferences
pushed down into the host's goal graph. Tier-0 GC prunes stale low-confidence
subjects (TTL + per-host subject cap). Semantic recall is served from here.

## Decision flow (one read)

`recollect("varrock.bank")` → Manager checks negative cache → L0/L1 cached copy?
(serve + telemetry) → else cascade to remote within the maturity deadline → miss
records a negative-cache entry; hit promotes to local. Writes go local-first with
remote write-back through the journal.

## What this is not

- Not a vector DB on the host: semantic recall is a **mesa** capability; the host's
  recall path is lexical/keyed.
- Not total recall: retention is policy-graded by design; the journal keeps
  episodes, not the firehose (the debug event JSONL is observability, not memory).
- Not shared between hosts except through mesa: hosts never read each other's
  local stores.
