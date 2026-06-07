# mesa — gRPC architecture sketch

> mesa is the shared, out-of-process home for everything a host CANNOT do on its
> own local compute: LLM inference, RAG, long-term memory, the memory crons, and
> persona/policy compilation. Hosts are isolated compute units; they phone mesa
> to unlock higher cognition, mirror local state up, and report metrics.
> The host holds NO keys and makes NO external calls — only mesa does.

## Topology (Phase B/C: mesa split out; the connection is host↔mesa)

The cradle is ONLY a management engine — it spawns/supervises N isolated Host
instances and shares static read-only assets (facts, landscape). It is NOT in
the data path: **each host owns its own gRPC connection to mesa.** A cradle-
hosted host and a `cmd/host` standalone host use the identical host→mesa path.

```
 ┌──────────── cradle = lifecycle manager (spawn/supervise/share facts) ────────┐
 │  Host #1            Host #2            ...            Host #N                  │
 │  (isolated)         (isolated)                        (isolated)              │
 │  world·routines·conductor·pearl(EXEC)·Limbic·hostkv  + Provision applier      │
 │     │                  │                                 │                     │
 │  ┌──┴── mesa client ┐  ┌──┴── mesa client ┐           ┌──┴── mesa client ┐    │
 │  │ OWN gRPC conn    │  │ OWN gRPC conn    │   ...     │ OWN gRPC conn    │    │
 │  │ auth as host_id  │  │ auth as host_id  │           │ auth as host_id  │    │
 │  │ unary Cognition  │  │ ...              │           │ ...              │    │
 │  │ unary Memory     │  └────────┬─────────┘           └────────┬─────────┘    │
 │  │ cstream Ingest   │           │                              │              │
 │  │ sstream Directive│           │                              │              │
 │  └────────┬─────────┘           │                              │              │
 └───────────┼─────────────────────┼──────────────────────────────┼─────────────┘
             │  one gRPC/HTTP2 conn PER HOST   (mTLS + host-auth @ Phase C)       │
             └─────────────────────┴──────────────┬───────────────┘
                                                   ▼
 ┌────────────────────────────── mesa service (cmd/mesa) ──────────────────┐
 │  gRPC server:  Cognition │ Memory │ Ingest │ Directive-hub               │
 │      │            │          │          ▲                                 │
 │      ▼            ▼          ▼          │ push                            │
 │  Inferencer   Retriever   Ingestor   (per-process stream, fan-out by host)│
 │  Anthropic    pgvector    batch writes                                   │
 │      │  ▲        │  ▲         │              ▲                            │
 │      │  └────────┴──┴─────────┴── RAG ───────┘                            │
 │      ▼                                                                    │
 │  Postgres + pgvector  (LTM of record)        Cron engine                  │
 │  Redis (hot cache, job/embed queue, locks)   reflection · consolidation · │
 │  Embedder (Voyage HTTP)                       trust-decay · reverie ·      │
 │                                               persona/policy COMPILATION  │
 │                                               └─ emits Directives ────────┘
 └──────────────────────────────────────────────────────────────────────────┘
```

## Four capabilities → gRPC shape

| Capability | RPC | Style | Notes |
|---|---|---|---|
| **Cognition** | `Cognition.Deliberate`, `.Evaluate` | unary, deadline-bounded | pearl-miss escalation. mesa: RAG → prompt → LLM → log brain_call → Decision (+cache hints). |
| **Recall** | `Memory.Retrieve` | unary, 200ms | pgvector hybrid query; Redis read-cache (Phase B). |
| **Mirror** | `Ingest.WriteEpisodes` / `UpsertRelationship` / `LogBrainCall` | client-stream, fire-and-forget | batched, idempotent (`host_id+hash+occurred_at` UNIQUE). The host's upload of episodes / trust deltas / affect / KV mirror / metrics. |
| **Provision** | `Directive.Subscribe` | **server-stream, one per host** | The host subscribes its own stream; `DirectiveEnvelope{id↑, kind, payload}`. Carries compiled Cornerstone, pearl Table (PEARL_REFRESH), goal/persona revisions, decay outputs. |

## Connection model
- **One gRPC connection per HOST** (not per process). Each isolated host owns its
  own `mesa.Client` + connection and authenticates as its own `host_id`. The host
  IS its own self-contained unit, mesa link included.
- The cradle is a **lifecycle manager only** — spawn/supervise/restart hosts,
  share static read-only assets (facts, landscape). It is **not** a connection
  aggregator and **not** in the data path. So `cmd/host` (standalone) and a
  cradle-hosted host take the identical host→mesa path.
- The Directive server-stream is **per host** (each host subscribes its own).
- Trade-off vs the old "1 conn/process, mux by host_id": more connections (N per
  cradle), chosen deliberately for isolation + identical standalone/cradle paths.
  gRPC handles many HTTP/2 conns; mesa scales horizontally (LB + replicas).
- Phase C (multi-cradle): mTLS; each host authenticates its own `host_id`.

## Data stores (in mesa only)
- **Postgres + pgvector** — LTM of record:
  - `episodes(host_id, text, embedding vector(1024), salience, stage, occurred_at, idempotency_key UNIQUE)`
  - `relationships(host_id, name, alpha, beta, encounters, trust_score, total_value_traded, tags)` — lossless, never deleted, only decayed
  - `brain_calls(...)` audit · `personas(doc, compiled_cornerstone, compiled_table, version)` · `goals(...)`
  - hybrid retrieval: HNSW ANN + structured WHERE, `score = salience·0.4 + similarity·0.6`
- **Redis** (Phase B only) — hot read cache (`rag:{host}:{hash}`, `persona:{host}`), embed/ingest job queue (Redis Stream), cron locks. Not before Phase B.
- **Embedder** — Voyage over HTTP, async worker fills `embedding=NULL`.

## Cron engine ("cron jobs across memory")
Scheduled jobs over the corpus, each emitting Directives downstream (Redis lock so one replica runs each):
- **Reflection** (~30m): summarize recent episodes → reflections → maybe `GOAL_REVISION`.
- **Consolidation**: dedupe/merge episodes, decay salience.
- **Trust decay** (hourly/daily): decay `relationships.trust_score` → `TRUST_DECAY`.
- **Reverie rebaseline**: recompute saliency weights → `REVERIE_REBASELINE`.
- **Persona compilation** (birth + revision) and **policy/pearl-table compilation** (birth + experience-driven recompile) → `PERSONA_REVISION` / `PEARL_REFRESH`. **Compilation lives here, not on the host.**

## Degradation / reconnect (host stays alive offline)
- Cognition/Recall timeout or `ErrOffline` → host falls back to pearl/local (`decide!`/`recall!`). Never blocks.
- Mirror down → host buffers in the **hostkv write-back journal** (already built), drains on reconnect.
- Directive stream drop → reconnect, resume from `lastApplied[host_id]` (monotonic id; idempotent, no CRDT).

## Phasing (right-sized)
- **Phase A (now):** mesa is an in-process Go package the cradle embeds; "RPC" = direct Go calls; Postgres via `pgx`; **no gRPC, no Redis.** The four-capability Go interfaces (`mesa.Client`) are the contract; `StubClient` stands in until the real impl. **All host code is shaped for this today.**
- **Phase B:** mesa splits to `cmd/mesa`; the Go interfaces gain a generated `.proto`; one gRPC conn/process; Redis enters.
- **Phase C:** multi-cradle; mTLS + host-auth.

## Package layout
- `mesa/client/` — host-side client lib: `mesa.Client` interface + `StubClient` + adapters (`AsStrategist`/`AsRetriever`/`AsRemote`). **Must graduate to a TRACKED location when `cmd/host`/cradle depends on it** (committed host code can't import gitignored code). Today it's sketched at `mesa/` root.
- `mesa/server/` — `cmd/mesa`: gRPC server, Inferencer, Retriever, Ingestor, Directive hub, Cron engine, Postgres/Redis/Embedder.

## Open decision
Where does the **host-side client lib** live so committed host code can depend on it? Options: (a) un-gitignore `mesa/client/` and track just the client; (b) move the client to a tracked top-level package (e.g. `mind/`), keep `mesa/` = server only; (c) make mesa a separate Go module the host `require`s. Pick before wiring `cmd/host` to a real (non-stub) mesa.
```
