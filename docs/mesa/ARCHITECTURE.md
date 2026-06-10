# mesa — service architecture

> **STATUS: CURRENT** (verified against code 2026-06-10, HEAD `0bfa818`). mesa is
> BUILT: a standalone gRPC daemon (`mesa/cmd/mesad`) backed by Postgres+pgvector,
> with Anthropic LLM seams, Voyage embeddings, two distillation crons, and an
> operator Admin plane (`cmd/mesa-ctl`). The wire contract lives in
> [`PROTOCOL.md`](PROTOCOL.md) and `mesa/proto/mesa.proto` (the SSOT).
> Backlog/open items: [`docs/TODO.md`](../TODO.md) (M-3/M-4/M-6, C-19, O-10).

mesa is the shared, out-of-process home for everything a host CANNOT do on its
own local compute: LLM inference, semantic recall (RAG), long-term memory, the
memory-distillation crons, and the persona lifecycle. Hosts are isolated compute
units; they phone mesa to unlock higher cognition, mirror local state up, and
report metrics. The host holds NO keys and makes NO external calls — only mesa
does (`ANTHROPIC_API_KEY`, `VOYAGE_AI_KEY` live on mesad; see
`mesa/mesad/server.go` package doc, `mesa/embed/voyage.go`).

## Topology (the connection is host↔mesa)

The cradle daemon (`cmd/cradle-server`, package `cradle/`) is ONLY a management
engine — it spawns/supervises N isolated Host instances and shares static
read-only assets (facts, landscape). It is NOT in the data path: **each host
owns its own gRPC connection to mesa** (`cradle/deps.go` `DialMesa` — "one gRPC
connection per host (no pooling)"). A cradle-hosted host and a `cmd/host`
standalone host use the identical host→mesa path (both call
`mesaclient.NewGRPCClient(addr, mesaclient.HostKey(hostID))`).

```
 ┌──────── cradle-server = lifecycle manager (spawn/supervise/share facts) ─────┐
 │  Host #1            Host #2            ...            Host #N                │
 │  (isolated)         (isolated)                        (isolated)             │
 │  world·routines·conductor·pearl(EXEC)·limbic·hostkv·memory.Manager           │
 │     │                  │                                 │                   │
 │  ┌──┴── mesa client ┐  ┌──┴── mesa client ┐           ┌──┴── mesa client ┐   │
 │  │ OWN gRPC conn    │  │ OWN gRPC conn    │   ...     │ OWN gRPC conn    │   │
 │  │ auth as host_id  │  │ auth as host_id  │           │ auth as host_id  │   │
 │  └────────┬─────────┘  └────────┬─────────┘           └────────┬─────────┘   │
 └───────────┼─────────────────────┼──────────────────────────────┼─────────────┘
             │      one gRPC/HTTP2 conn PER HOST (bearer token)   │
             └─────────────────────┴──────────────┬───────────────┘
                                                  ▼
 ┌──────────────────── mesad (mesa/cmd/mesad, :7077) ──────────────────────────┐
 │  gRPC services:  Game │ Knowledge │ Journal │ KV │ Provision │ Admin        │
 │                   │        │          │       │       │          │          │
 │  LLM tiers (mesa/llm, Anthropic):     │       │   per-host    operator      │
 │   actLLM=Sonnet · decideLLM=Haiku ·   │       │   Subscribe   plane         │
 │   genesisLLM=Opus                     │       │   push stream (mesa-ctl,    │
 │                                       ▼       ▼       │        ADMIN_TOKEN) │
 │  Postgres + pgvector (LTM of record, mesa/mesad/ltm.go)                     │
 │  Voyage embedder (mesa/embed, embed-on-write + resend backfill)             │
 │  Cron engine: consolidation (Tier-0/1, Haiku) · insight (Tier-2, Sonnet)    │
 └─────────────────────────────────────────────────────────────────────────────┘
```

## Capabilities → real gRPC services

The four conceptual capabilities (Cognition / Recall / Mirror / Provision)
manifested as six services. Message-level detail: [`PROTOCOL.md`](PROTOCOL.md)
and `mesa/proto/mesa.proto`; server impl: `mesa/mesad/`.

| Service | RPCs | Style | Role |
|---|---|---|---|
| **Game** | `Act`, `Decide`, `Chat`, `AnalysisInterpret`, `ExtractDialog` | unary | The LLM seams. `Act` = the planner agent-step (Situation→DSL Move, `act.go`); `Decide` = narrow pearl-miss option pick; `Chat` = cheap social reflex off the Act loop (`runtime/runhost_bootstrap.go` `socialReflex`). |
| **Knowledge** | `Recall`, `FetchRelationships`, `FetchGoal`, `FetchKnowledge`, `FetchGoalGraph` | unary | Semantic recall over LTM (~200ms deadline) + cold-start bootstrap of the trust ledger / goal / knowledge ledger / goal graph (mesa→host). |
| **Journal** | `Remember` (stream), `RecordObservations` (stream), `SyncRelationships`, `SyncGoal`, `SyncKnowledge`, `SyncGoalGraph`, `ReportMetrics` | client-stream + unary | The host's mirror up: episodes, the salience-gated observation firehose (cron fodder), AuthLocal snapshots, telemetry. Idempotent (`(host_id, idem_key)` PKs). |
| **KV** | `Put`, `Get`, `Delete` | unary | `memory.Manager`'s remote tier — opaque durable host state. |
| **Provision** | `Fetch`, `Subscribe` (server-stream), `Genesis` | unary + stream | Persona down (`persona_json` — the pearl table is func-valued and **never crosses the wire**; the host compiles it locally), live directive pushes, and the heavy Opus session-genesis compile. |
| **Admin** | `PutPersonas` (stream), `GetPersona`, `ListPersonas`, `DeletePersona`, `PushGoal` | operator-only | Persona CRUD + live goal push via `cmd/mesa-ctl`; gated on `$ADMIN_TOKEN`, fail-closed (`mesa/mesad/admin.go`, `auth.go`). See [`admin-control-plane.md`](admin-control-plane.md). |

Host-side, one `mesaclient.Client` backs the existing seams unchanged via
adapters (`mesa/client/adapters.go`): `AsStrategist` (brain.Strategist →
`Decide`), `AsRetriever` (cognition.Client → `Recall`), `AsRemote`
(memory.Remote → `Recall`/`Remember`). `Act` is consumed directly by the
MesaDirector (`runtime/mesa_director.go`).

## Connection model
- **One gRPC connection per HOST** (not per process). Each isolated host owns its
  own `mesaclient.Client` + connection and authenticates as its own `host_id`.
  The host IS its own self-contained unit, mesa link included.
- The cradle is a **lifecycle manager only** — spawn/supervise/restart hosts,
  share static read-only assets (facts, landscape). It is **not** a connection
  aggregator and **not** in the data path. So `cmd/host` (standalone) and a
  cradle-hosted host take the identical host→mesa path.
- The `Provision.Subscribe` server-stream is **per host** (each host subscribes
  its own; mesad keeps a per-host push registry, `server.go` `subs`).
- Trade-off vs "1 conn/process, mux by host_id": more connections (N per
  cradle), chosen deliberately for isolation + identical standalone/cradle paths.
  gRPC handles many HTTP/2 conns fine.
- **Auth today:** per-RPC bearer token, derived as `SHA-512(username)` — an
  explicit PLACEHOLDER for identity *binding*, not secrecy (`mesa/auth/hostkey.go`;
  the shared leaf both client and server import). The interceptor binds the
  authenticated `host_id` into ctx; handlers never trust a request-body host_id
  (`mesa/mesad/auth.go`). The Admin plane uses a separate `$ADMIN_TOKEN`
  (fail-closed: unset ⇒ Admin disabled). Transport is **insecure** — the link is
  trusted/local for now; TLS/mTLS lands with deployment
  (`mesa/client/grpc.go:35`, TODO **M-6**).

## Data stores (in mesa only)

Postgres + pgvector is the **only** store — there is no Redis anywhere in the
system (the old Phase-B Redis plan was never adopted). All DDL lives in
`mesa/mesad/ltm.go` `migrate()` (idempotent, run at `OpenLTM`):

- `episodes(host_id, idem_key, kind, body, importance, entity, occurred_at, created_at, embedding vector(1024))`
  — PK `(host_id, idem_key)`; HNSW cosine index. Embedded **on write** (Voyage,
  `voyage-3`/1024-dim); a NULL embedding (no key / transient error) is
  backfilled by the host's at-least-once journal resend (`ltm.go` `Add`, M12).
- `observations(host_id, idem_key, kind, subject, body, salience, occurred_at, …)`
  — the raw perception firehose; cron fodder, deliberately NOT in episode recall.
- `relationships(host_id, name, alpha, beta, encounters, tags, value_traded, affinity, grievance, …)`
  — the AuthLocal trust-ledger mirror: the host pushes absolute Beta(α,β) state
  up; only `value_traded` is journal-accrued mesa-side (completed-trade episodes).
- `knowledge(host_id, subject, kind, encounters, last_seen, tags, beliefs_json, …)`
  — the distilled world-knowledge ledger; written by both the consolidation cron
  and `SyncKnowledge` (last-writer-wins per subject).
- `goal_graphs(host_id, snapshot jsonb, …)` — the intention graph; host pushes,
  insight cron grows it.
- `personas(host_id, persona_json jsonb, prose_card, cooked, …)` — the host's
  compiled identity, plus GENERATED projection columns surfacing every facet/dial
  as queryable columns derived from the JSON (NULL-tolerant casts, L15).
- `kv(host_id, key, value bytea, …)` — the generic opaque host-state mirror.
- `goals(host_id, objective, progress jsonb, …)` — the mutable standing objective.
- `metrics(host_id, name, value, at)` — append-only telemetry time series.

**Recall ranking** is pure pgvector cosine when an embedder is wired
(`recallCosine`); without one it degrades to Postgres full-text relevance +
recency (`recallLexical`; empty query ⇒ pure recency, the bootstrap path).
Composite ranking (similarity × recency × salience) is open — TODO **C-15/M-2**.

## Cron engine (the System-2 distillation)

Two background loops, both in `mesa/mesad/` (`cron.go`, `cron_insight.go`),
started by mesad after `Attach` (`StartCrons`; `-cron-disable` to opt out):

- **Consolidation** (default 60s): folds the observation firehose into the
  `knowledge` ledger. Tier-0 (no LLM): novelty/dedup, salience-decay GC,
  subject caps. Tier-1 (Haiku, `decideLLM`): batched per-host claim extraction
  (50 obs/call default — the core cost lever). Flags rare high-salience claims
  onto a durable per-host escalation queue (`cron:escalate:queue` in `kv`).
- **Insight** (default 180s): Tier-2, RARE — drains the escalation queue with
  ONE Sonnet call per host per tick (never Opus, never on bulk): contradiction
  reconciliation, cross-entity goal-graph chaining, LLM-judged open-question
  closure. Cursor-gated + idempotent (a re-run is a no-op).

Invariants: crons never hold `s.mu` during an LLM call, iterate hosts off
Postgres (not the in-mem registry), and are concurrency-capped (`cronSem`) so
they cannot starve Act/Chat/Decide. Knobs are mesad flags
(`-cron-consolidate-every`, `-cron-batch-size`, `-cron-concurrency`, …).

**Not built** (the old design listed them as crons; the directive kinds are
reserved in the proto but unused): reflection summaries, trust decay
(TODO **C-24**), reverie rebaseline (TODO **P-6**). Persona/policy compilation
is NOT a cron — it happens at registration/import time (`Server.Register`:
validate → render → persist → register live; prose cook via `mesa/personacook`)
and the pearl table is always compiled host-side from the provisioned persona.

## Degradation / reconnect (host stays alive offline)
- LLM seams down (`ANTHROPIC_API_KEY` unset ⇒ Unavailable; or timeout/error):
  `decide()` tries the pearl fast path and the TTL decision cache BEFORE mesa
  (`runtime/actions_flow.go` `dslDecide`); an `Act` failure idles the
  director ~3s and retries next turn — the loop never dies
  (`runtime/mesa_director.go`, the `d.idle(3)` on planner error). `StubClient`
  (always `ErrOffline`) keeps a mesa-less host fully local.
- Mirror down → the host buffers writes in the **`memory.Manager` write-back
  journal** persisted in hostkv (`memory/manager.go`, `journalPrefix`); `Flush`
  drains oldest-first once `Remote.Healthy()` again.
- Directive stream: the host subscribes once at provision
  (`runtime/runhost_bootstrap.go` `subscribeDirectives`). `GOAL_REVISION` is
  applied live (`SetLiveGoal`); `PEARL_REFRESH`/`PERSONA_REVISION` are
  logged-not-applied today (TODO **M-3**). **DESIGN (not built):** resume-on-
  reconnect via `last_applied_id` — the proto reserves the field
  (`SubscribeRequest`), but the client does not send it and there is no
  reconnect loop yet.

## Phasing — collapsed (historical)

The old A/B/C ladder resolved as: **A and B happened, without Redis.** mesa was
born as the standalone gRPC daemon `mesa/cmd/mesad` (there was never a shipped
in-process embed phase), Redis was never adopted anywhere, and the connection
model landed as one-conn-per-HOST (not the per-process mux B envisioned). What
remains of "Phase C" is deployment hardening — TLS/mTLS and real per-host
secrets (TODO **M-6**) — plus fleet-scale LLM throttling (TODO **O-10**).

## Package layout (all tracked)
- `mesa/proto/` — `mesa.proto` (the wire SSOT) + generated pb/grpc code.
- `mesa/mesad/` — the server: services, auth interceptors, LTM, crons, genesis,
  arg catalog (static validation of hallucinated literals in authored DSL).
- `mesa/cmd/mesad/` — the binary (`:7077`; flags for models, DSN, crons, facts).
- `mesa/client/` — the host-side client lib: `Client` interface, `GRPCClient`,
  `StubClient`, seam adapters. Imported by `cmd/host` and `cradle/` (the old
  "must graduate to a tracked location" open decision is resolved — it did).
- `mesa/auth/` — the shared HostKey derivation leaf (client+server import it;
  neither imports the other).
- `mesa/llm/` — the Anthropic client (three tiers wired in mesad: act=Sonnet,
  decide=Haiku, genesis=Opus).
- `mesa/embed/` — the Voyage embedding client.
- `mesa/personacook/` — offline persona prose-cook tooling.
- `mesa/wikirag/`, `mesa/wikitest/` — out-of-band offline wiki tooling, not in
  the host/mesa runtime path.

Open/backlog items live in [`docs/TODO.md`](../TODO.md) (SSOT) — see **M-3**
(live PERSONA_REVISION push), **M-4** (admin v3 remainder), **M-6** (mTLS),
**C-19** (genesis remainder / unused directive kinds), **O-10** (fleet LLM
throttling).
