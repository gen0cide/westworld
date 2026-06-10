# Mesa — the off-host cognition, memory, and control service

> **STATUS: BUILT, live** (verified 2026-06-10 against branch HEAD `0bfa818`).
> Mesa is a **gRPC** service (`mesa/proto/mesa.proto` — six services) over
> **Postgres + pgvector** (`mesa/mesad/ltm.go`), run as `cmd/mesad` and
> operated through `cmd/mesa-ctl`. The original REST/SQL/TrustGrade design this
> doc used to carry is archived verbatim at
> `docs/archive/initial-brainstorming/mesa-original-design.md`; the one design
> reversal worth remembering (trust authority) is recorded below. Companion
> docs: `docs/mesa/ARCHITECTURE.md` (topology + degradation ladder),
> `docs/mesa/PROTOCOL.md` (per-RPC contract),
> `docs/mesa/admin-control-plane.md`, `docs/mesa/knowledge-pipeline.md`.

## What mesa is

The compute-locality split: a **host is an isolated local compute unit** — it
holds no API keys and makes no external calls. Mesa is the off-host home for
everything that is *not* compute-local-feasible (`mesa/mesad/server.go` package
doc; host-side contract in `mesa/client/doc.go`):

- **LLM inference** — the Anthropic key lives only here (`mesa/llm`); every
  host LLM call is a mesa RPC, never Anthropic-direct.
- **Long-term memory + RAG** — durable, host-namespaced episodic memory in
  Postgres, embedded with Voyage, recalled by pgvector cosine similarity.
- **The distillation crons** — async System-2 loops that fold the raw
  observation firehose into graded world-knowledge and the intention graph.
- **Persona lifecycle** — storage (the `personas` table), validation, prose
  render/cook, provisioning down to hosts, runtime CRUD.
- **The operator control plane** — the admin-token-gated `Admin` service +
  `mesa-ctl` (persona CRUD, fleet config generation, live goal push).

The cradle daemon (`cmd/cradle-server`) is **not** in this data path — it is a
lifecycle manager that spawns/supervises hosts and shares static assets. Each
host owns its own gRPC connection and authenticates as its own host_id; a
cradle-hosted host and a standalone `cmd/host` take the identical host→mesa
path.

Why a service and not per-host sqlite, in present tense: (1) the keys and the
LLM seams must live off-host; (2) population-scale analytics are SQL over one
store (the `metrics` time series, the generated persona facet columns); (3)
crash recovery — hosts restart and re-bootstrap identity, memory, trust,
knowledge, and intentions from the `Fetch*` cold-start RPCs. The latency
trade-off is absorbed by cadence: mesa is hit on turn cadence (Act) and async
mirrors, never per game tick.

## Binaries and packages

| What | Where |
|---|---|
| The service | `cmd/mesad` (gRPC on `-addr`, default `:7077`) |
| Operator CLI | `cmd/mesa-ctl` (persona put/import/ls/get/set/rm, `fleet gen`, `goal push`) |
| Wire contract | `mesa/proto/mesa.proto` (+ generated `mesa.pb.go` / `mesa_grpc.pb.go`) |
| Server impl | `mesa/mesad` (server.go, act.go, genesis.go, ltm.go, cron.go, cron_insight.go, admin.go, auth.go, catalog.go, dslmanual.go) |
| Host-side client | `mesa/client` — `GRPCClient` (grpc.go) + `StubClient` (offline: every call degrades to `ErrOffline` → local behavior) + adapters backing the brain/cognition/memory seams |
| Anthropic client | `mesa/llm/anthropic.go` — SDK-free Messages client, prompt-cached system blocks |
| Embeddings client | `mesa/embed/voyage.go` — Voyage AI, default `voyage-3` (1024-d) |
| Persona cook | `mesa/personacook` (see `docs/persona-compile.md`) |
| Offline wiki RAG tooling | `mesa/wikirag` (out-of-band; not host runtime — see `docs/mesa/knowledge-pipeline.md`) |

## The gRPC surface

Six services in `mesa/proto/mesa.proto`. Per-method request/response shapes and
semantics live in `docs/mesa/PROTOCOL.md`; this is the map:

| Service | RPCs | What it is |
|---|---|---|
| `Game` | `Act`, `Decide`, `Chat`, `AnalysisInterpret`, `ExtractDialog` | The LLM seams. `Act` is the planner step (situation → DSL Move); `Decide` a narrow in-routine option pick; `Chat` the cheap social reply (plus the proactive `mode="ask"` branch); `AnalysisInterpret` classifies operator-override directives; `ExtractDialog` is the reactive tier's claims+intent extraction. |
| `Knowledge` | `Recall`, `FetchRelationships`, `FetchGoal`, `FetchKnowledge`, `FetchGoalGraph` | RAG over the host's own episodes (`Recall`, ~200ms-deadline contract) + the four cold-start bootstraps a restarting host pulls down. |
| `Journal` | `Remember` (client-stream), `RecordObservations` (client-stream), `SyncRelationships`, `SyncGoal`, `SyncKnowledge`, `SyncGoalGraph`, `ReportMetrics` | The host→mesa mirror path: episodes, the salience-gated perception firehose, and absolute snapshots of the AuthLocal ledgers + telemetry. |
| `KV` | `Put`, `Get`, `Delete` | The generic opaque-state transport under `memory.Manager`'s remote tier — not a host-facing surface; purposeful state uses a semantic RPC instead. |
| `Provision` | `Fetch`, `Subscribe` (server-stream), `Genesis` | Identity down: the compiled persona/prose/goals bundle, the live `Directive` push stream, and the heavy Opus-at-login session compile. |
| `Admin` | `PutPersonas` (client-stream), `GetPersona`, `ListPersonas`, `DeletePersona`, `PushGoal` | The operator plane (separate credential; below). |

**`Game.Act`** is the workhorse (`mesa/mesad/act.go`): the static DSL manual
(`dslmanual.go`, generated tail appended from `dsl/spec` so it never drifts
from the engine) rides as an ephemeral **prompt-cached** system block; the
persona prose card and live situation ride uncached. Mesa parses + validates
the returned move *before* it ships — an author→validate→re-prompt loop
(`maxActAttempts = 3`), including static rejection of hallucinated literal args
(`go_to("mining-site")`) against the world name catalogs (`catalog.go`, loaded
via `-facts` from `facts.LoadStaticCatalogs`). After the last failed attempt it
falls back to a brief idle rather than shipping broken DSL.

**`Provision.Genesis`** (`genesis.go`) is the session compile: one Opus call at
login over persona + episode history + relationships + standing goal — all
gathered mesa-side — returning a history-aware goal, a mood baseline, and the
keyword→tier→action attention ladder.

**`Provision.Subscribe`** (`server.go:609`) is the live push stream: a per-host
buffered channel registry (`subs`), monotonic directive ids, and a
server-driven drain on SIGTERM (`Shutdown()` closes a channel every stream
selects on, so `GracefulStop` can complete with a fleet connected). Today the
only producer is `Admin.PushGoal` (`GOAL_REVISION`); an operator-pushed goal is
re-sent on reconnect, a persona-baseline goal deliberately is not
(`entry.goalPushed`), so a connect-time send never clobbers a richer genesis
goal. Live `PERSONA_REVISION` push is open (TODO **M-3**).

## Identity and auth

- Every RPC is authenticated by per-host bearer token via the
  `UnaryAuth`/`StreamAuth` interceptors (`mesa/mesad/auth.go`): token →
  host_id, bound into the request context. Handlers read
  `hostIDFromContext` and **never trust a host_id from the request body** —
  every row a host touches is scoped to its authenticated identity.
- The token is `mesaclient.HostKey(username)` = SHA-512(username)
  (`mesa/client/hostkey.go`) — an explicitly-flagged **placeholder, not a
  secret**: it wires identity binding through the stack with no out-of-band
  issuance, to be replaced (real per-host secret or mTLS) before exposure
  beyond a trusted link (TODO **M-6**).
- The `Admin` service authenticates separately with the operator credential
  (`$ADMIN_TOKEN`, constant-time compare), **fail-closed**: an unset token
  disables the Admin API entirely. Admin methods are never reachable with a
  host bearer token.

## LLM tiering

Three `mesa/llm` clients, wired from `cmd/mesad` flags (all seams require
`ANTHROPIC_API_KEY`; without it they return `Unavailable` and the host degrades
to local behavior — persona provisioning still works):

| Tier | Flag (default) | Used by |
|---|---|---|
| `actLLM` | `-act-model` (`claude-sonnet-4-6`) | `Act` (high volume; cached manual prefix), the Tier-2 insight cron |
| `decideLLM` | `-decide-model` (`claude-haiku-4-5-20251001`) | `Decide`, `Chat`/`Ask`, the Tier-1 consolidation cron |
| `genesisLLM` | `-genesis-model` (`claude-opus-4-8`) | `Genesis` only (rare, history-rich) |

The cardinal cost rule (enforced by the cron seam selectors in `cron.go` /
`cron_insight.go`): bulk work is Haiku, the rare pre-filtered slice is Sonnet,
**never Opus on bulk**. Note there is currently no token/cost ledger and no
fleet-level throttle in front of these seams — TODO **O-9** / **O-10**.

## Storage: the real schema

Postgres is mesad's system of record (`mesa/mesad/ltm.go`; DSN resolution
`-db` > `$DATABASE_URL` > `postgres://localhost:5432/westworld` — startup is
fatal without it). The pool floor is 16 conns so the crons can't starve the
latency-sensitive RPCs. `migrate()` (ltm.go:113) is idempotent and runs in one
transaction. Embeddings: when `VOYAGE_AI_KEY` is set, episodes are embedded on
write and queries on recall (`mesa/embed`, `voyage-3`, 1024-d), ranked by
pgvector cosine over an **HNSW** index; without the key, `Recall` degrades to
Postgres full-text relevance + recency.

| Table | Key | What |
|---|---|---|
| `episodes` | `(host_id, idem_key)` | kind/body/importance/entity/occurred_at + `embedding vector(1024)`. Written by `Journal.Remember` (deduped by idempotency key), read by `Knowledge.Recall`. |
| `observations` | `(host_id, idem_key)` | The salience-gated perception firehose — cron fodder, **never recalled into the planner**. Capture kinds (npc_dialog/player_chat/server_msg/claim_heard) get a text-hashed idem key so distinct same-second lines don't collapse. |
| `relationships` | `(host_id, name)` | Beta `alpha`/`beta` + encounters/tags/value_traded + the multi-axis `affinity`/`grievance` columns. A **mirror** of the host-local ledger (see decision record). |
| `knowledge` | `(host_id, subject)` | `beliefs_json` (`[]KnowledgeBelief`): graded, provenance-tagged beliefs about things. Written by both the consolidation cron and the host's `SyncKnowledge` — last-writer-wins per subject. |
| `goal_graphs` | `host_id` | The intention-graph snapshot (`{nodes,edges}` jsonb); host `SyncGoalGraph` + the insight cron share the upsert, last-writer-wins per host. |
| `personas` | `host_id` | The compiled identity — see next section. |
| `kv` | `(host_id, key)` | Opaque durable host state (`memory.Manager` remote tier). Also where the crons keep per-host state: `cron:consolidate:cursor`, `cron:escalate:queue`. |
| `goals` | `host_id` | The mutable standing objective + `progress` jsonb (vs the immutable persona north-star) — structured so analytics can ask "who's pursuing X". |
| `metrics` | `(host_id, name, at)` index | Append-only telemetry time series from `ReportMetrics` (rollup/retention is a future cron concern). |

What does **not** exist (vs the archived design): no `bots`, `reflections`,
`routines`/`routine_versions`, `persona_revisions`, `brain_calls`,
`knowledge_chunks`, `bot_vocabulary`, or `working_scratch` tables. Routine
storage is host-local; the cost ledger is open (**O-9**); the shared knowledge
corpus enters via the knowledge pipeline instead
(`docs/mesa/knowledge-pipeline.md`).

## The personas table

(The storage row `docs/persona-authoring.md` §15 points at.)

- **`persona_json jsonb` is canonical** — the exact host wire format
  (`persona.Persona`), provisioned down by `Provision.Fetch`.
- `prose_card text` is the rendered identity card the brain actually reads
  (cook output, or the `persona.Render()` floor); `cooked boolean` records
  which. Both are app-written — prose is derived, never stored in the JSON.
- 33 **`GENERATED ALWAYS AS ... STORED` projection columns** are extracted
  from `persona_json` (ltm.go:268-303): name, archetype, cohort_id,
  north_star/theme, voice facets, the six HEXACO bands, the two values, the
  prefs bands (patience, loss_aversion, aggression, decisiveness, tenacity,
  bulk_apperception, self_preservation, coop_type, the risk triple,
  attention_level), and the five curiosity dials as reals. Always in sync with
  the JSON, zero app-side extraction — population analytics are plain SQL.
  Numeric extractions are NULL-tolerant (regex-guarded `CASE`) so one malformed
  persona row can't brick the migration (the L15 lesson, with a live-DB
  remediation pass for the old hard-cast columns).
- Lifecycle: `Server.Register` (server.go:171) validates → renders prose +
  system prompt → registers live → upserts here; `LoadPersonas`
  (server.go:215) restores everything on boot, so `-host host_id=persona.json`
  flags are a *first-run seed*, not the mechanism. Runtime mutation goes
  through the Admin plane.
- One row per host, last write wins — the original `persona_revisions` history
  table was never built; history/rollback is TODO **M-4**.

## The crons

`mesa/mesad/cron.go` + `cron_insight.go` — the Phase-4 System-2 distillation
(invariants and the cost model: `docs/world-knowledge-and-learning.md`
§3.5/§3.6). Started by `mesad` after `Attach`, stopped before the gRPC drain
(no torn writes); `-cron-disable` turns them off.

- **Tier-0 (no LLM):** novelty/dedup on fold, and GC — stale low-confidence
  subjects pruned past `-cron-knowledge-ttl` (30d), per-host subject cap
  `-cron-max-subjects` (500).
- **Tier-1 (Haiku, the bulk):** the consolidation loop, every
  `-cron-consolidate-every` (60s) — per-host batched claim extraction over the
  observation firehose (`-cron-batch-size`, 50 obs per call), folded into the
  `knowledge` ledger; cursor + poison-batch skip kept in `kv`. Items the
  extractor flags ride into a bounded escalation queue (cap 200).
- **Tier-2 (Sonnet, rare):** the insight loop, every `-cron-insight-every`
  (180s) — drains the escalation queue (≤ `-cron-insight-max-per-host` 6 items
  per call): open-question closure and cross-entity chaining, merged into
  `goal_graphs`.

Starvation guards: crons never hold `s.mu` during an LLM call, iterate hosts
off Postgres (not the in-mem registry), are capped by `cronSem`
(`-cron-concurrency`, 4) and `-cron-max-hosts-per-sweep` (64, most-active
first), and each per-host job has a 60s timeout.

The archived design's five background jobs (reflection generator,
decay/compression sweep, importance scorer, TrustGrade, trust decay) were never
built in that shape — the consolidation+insight pair is what shipped.

## The Admin control plane

`mesa/mesad/admin.go` + the `Admin` service; client `cmd/mesa-ctl`. Design doc:
`docs/mesa/admin-control-plane.md` (built v1 + goal push + fleet gen).

```
mesa-ctl persona put <host_id> <file|->    # register/replace one (live, no restart)
mesa-ctl persona import <dir|->            # bulk: dir of <host_id>.json, or NDJSON on stdin
mesa-ctl persona ls|get|rm <host_id>
mesa-ctl persona set <host_id> <field> <v> # patch one dial/facet; server-side re-validate
mesa-ctl fleet gen [flags]                 # emit a cradle hostcfg from the persona registry
mesa-ctl goal push <goal> [--match glob]   # SOFT runtime goal onto running hosts
```

- `PutPersonas` streams; one bad persona doesn't fail the batch (per-item
  report — a 199-good/1-bad import lands the 199). Handlers reuse
  `Server.Register`, so every change is validate → render → persist → live,
  with **no mesad restart**. `DeletePersona` also revokes the host's derived
  token.
- `fleet gen` reads the registered persona set and emits a matching
  `cradle/hostcfg` YAML (names == host_ids by construction) — the bridge from
  mesa's registry to the cradle daemon's fleet config.
- `PushGoal` fans a `GOAL_REVISION` out over the `Subscribe` registry to hosts
  matching an optional host_id glob, reporting pushed (live) vs matched
  (including offline) counts. It is a *soft* override: the director prefers it
  until replaced.
- Remainder (live persona-revision push, `hosts` roster, recook,
  history/rollback, genesis-via-ctl): TODO **M-3**/**M-4**.

## Decision record: the trust-authority reversal

The archived design made **mesa authoritative for trust**: hosts POSTed
provisionally-graded relational events, and a mesa-side *TrustGrade* job
(batched Haiku, severity weight w ∈ [0.3, 8], an Honesty-Humility-shaped Beta
prior) returned the authoritative grade, with the host converging on its next
read.

That reversed. **Trust is host-local and host-authoritative** (`AuthLocal`):

- The ledger lives in `limbic/ledger.go` — a multi-axis (trust/affinity/
  grievance) Beta(α,β) posterior with a **uniform α=β=1 prior** (not
  H-shaped), updated by the host's own deterministic event handling.
- `memory/policy.go:44` defines `AuthLocal` ("this host's private truth") and
  classifies `relationship` — along with `goal`, `journal`, `knowledge`, and
  `goalgraph` — under it.
- Mesa's `relationships` table is a **mirror**: the host pushes absolute
  snapshots up (`Journal.SyncRelationships` — "best-effort durability; the
  host stays authoritative", server.go:394) and mesa serves them back only for
  cold-start bootstrap (`Knowledge.FetchRelationships`).
- No mesa-side trust-grading LLM job exists. The name `TrustGrade` survives
  only as the host-side bucket enum (`limbic.TrustGrade`:
  hostile/wary/neutral/friendly/trusted).

Mesa is authoritative only for what it computes: the persona compile, and the
cron-distilled beliefs — which share the per-subject last-writer-wins upsert
with the host's own pushes rather than overriding them.

## Open work

All backlog for this layer lives in `docs/TODO.md` (the SSOT) — §5 Mesa
(**M-1**–**M-9**; M-7 carries this doc's former open-questions section) plus
the cross-cutting **O-9** (LLM cost/token ledger — nothing tracks tokens or
dollars today) and **O-10** (fleet LLM throttling/degradation).
