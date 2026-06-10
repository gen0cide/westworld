# Architecture

> **STATUS: verified against code 2026-06-10, branch `tidy/structure-and-docs`,
> HEAD `0bfa818`.** Every package/binary claim below was checked against the
> tree at that commit. Open work is tracked in [`TODO.md`](TODO.md) (the SSOT —
> IDs like `P-6`/`O-9` below refer to its items); this doc carries no backlog
> of its own.

> **Reading guide.** This doc slices the system by Go package (the
> implementation view). If you're new and want the *AI-perspective*
> view of the host — body → senses → routines → cognition → persona —
> read [`layers.md`](layers.md) first, then come back here for the
> package-level details. One cross-cutting frame from that doc to keep
> in mind: the host's seven layers are its **in-band** consciousness
> (sense → think → act), and around them sits an **out-of-band**
> management plane it never experiences — the cradle daemon
> (`cmd/cradle-server`) and mesa (see [`layers.md`](layers.md) for the
> in-band/out-of-band framing).

## The binaries

Westworld is a Go monorepo at `github.com/gen0cide/westworld`. Three entry
points are the production system; the rest are operator CLIs and offline
tooling.

| Binary | Status | Purpose | Lifespan |
|---|---|---|---|
| `cmd/cradle-server` | BUILT | The host control-plane daemon. Loads a `*.hostcfg` file (or a directory of them) and runs + supervises the whole fleet in **one process** over shared static resources (facts + landscape + world oracle loaded once, shared by pointer; one mesa connection per host). Mounts the HTTP/JSON control API + embedded web UI on `-listen` (default `localhost:8099`) and proxies each live host's debug surface under `/api/hosts/{name}/debug/` — one server, no per-host ports. Daemon core in package `cradle/` (`registry.go`, `api.go`, `webui.go`, `hostcfg/`). | Always-on operations. |
| `cmd/host` | BUILT | Standalone single-host runner: connect to OpenRSC, pump the world mirror, then either the autonomous conductor turn loop (`-goal`), a fixed routine sequence (`-routines`), or an interactive REPL. `-debug-addr` serves the `debughttp` dashboard. The "just bring up a host" entry point; the cradle daemon is the fleet path. Both share the same composition root, `runtime.RunHost`. | One session. |
| `cmd/mesad` | BUILT | The off-host cognition service: a **gRPC** server (`mesa/proto`, default `:7077`) holding the Anthropic key the host is forbidden from having. Three LLM tiers wired by flags: Act/DSL-authoring (Sonnet), Decide option-picks (Haiku), session Genesis (Opus). All durable state in Postgres (+pgvector when a Voyage key is set) — see the mesa section below. | Always-on infrastructure. |

Operator CLIs (thin clients, no business logic): `cmd/cradle-ctl` (HTTP/JSON
over the cradle-server API — same API the web UI uses), `cmd/mesa-ctl` (gRPC
over mesad's operator-only Admin service: persona registry, bulk import, goal
push), `cmd/orsc-ctl` (HTTP over the OpenRSC admin API).

Offline / dev tooling: `cmd/legacy-cradle` (the original single-bot CLI;
still the home of the `-spectate` live software-rendered browser viewport and
a standalone `-debug-http` control plane), `cmd/parsecheck` (static
parse/validate gate over the scenario corpus), `cmd/scenariogen` (scenario
manifest maintenance), `cmd/defsgen` (regenerates the checked-in static world
data in `facts/` from an OpenRSC tree), `cmd/dronegen` (fleet persona
generator), `cmd/varrock-tiles` (teleport-target tile inspector). Mesa-side:
`mesa/personacook` (Opus persona prose cook), `mesa/wikirag` (offline
Voyage wiki embed/search — out-of-band tooling, not host runtime), and
`mesa/wikitest` (one-shot `cognition/corpus` wiki-dump recall smoke check).

There is **no `cmd/delos`** and never was code under that name. The swarm-
orchestrator role earlier docs assigned to "delos" shipped as
`cmd/cradle-server` + the cradle web UI + `cradle-ctl`; the residue
(cohort/archetype analytics, cost dashboard, perturbation toolkit) is
TODO `O-7`.

Packages are shared across binaries by explicit import (e.g. `persona` is
imported by both the host runtime and mesad — the schema and `CompilePolicy`
are one codebase on both sides of the wire). No `internal/` discipline.

## The layer cake (within one host)

Reading bottom-up. Each layer depends only on layers below. Each layer is one
(or, for `dsl`/`cognition`, one tree of) Go package at the top level of the
repo.

**STATUS legend:** BUILT = working code in the tree; NOT BUILT = no package
exists.

| Layer | Package | Status | Responsibility |
|---|---|---|---|
| 1. Wire protocol | `proto/v235` | BUILT | RSC packet encode/decode (Payload235-equivalent). Pure functions, no I/O. ISAAC cipher, RSC string compression, the multi-event coords decoders. |
| 2. Session | `session` | BUILT | TCP connection lifecycle: handshake, RSA login, packet read/write goroutines. |
| 3. World state | `world` | BUILT | State mirror — own player, nearby entities, ground items, inventory, skills/xp. Updated by inbound packets via `Apply`. |
| — Static knowledge | `facts`, `assets` | BUILT | Read-only world data (item/npc/object defs) and the `.orsc` archive decoders (`assets`: jag/bzip/sprites/models). The defs are **generated Go literals** (`facts/static_defs_gen.go`, `static_locs_gen.go`, `idnames_gen.go` via `cmd/defsgen`) so a deployed host needs zero runtime file I/O. One `*facts.Facts` is loaded once per process and shared by pointer across hosts. |
| — Pathfinding | `pathfind` | BUILT | BFS routing over the 96×96 local collision grid, plane-aware (`plane = worldY/944`), with a **live overlay** (`pathfind.LiveState`, fed by `runtime.liveState()`): opened doors clear, closed doors block, depleted scenery clears, spawned blockers block. Feeds `WalkTo`/`go_to`. |
| — World map | `worldmap` | BUILT | The precomputed static-geography **Oracle** behind the `search_map`/`reachable`/`survey_map` perception verbs: global walkability + capability-gated transport. Loaded once per process, shared by pointer (`Options.WorldOracle`). |
| 4. Actions | `action` | BUILT | Atomic operations: `Walk`, `Attack`, `Eat`, trade/bank/shop/magic/prayer/duel, etc. Each builds outbound packets; `runtime.Host` validates preconditions against `world`. |
| 5. Events | `event` | BUILT | Pub/sub bus for things that happen *to* the host. The frame handler publishes typed events (incl. synthetic deltas: `ItemGained`, `XPGain`, `TargetDied`, the player level/equipment-change edges); layers above subscribe. Lossy by design: a slow subscriber drops, never back-pressures the frame pump (`event/bus.go`). |
| 6. Memory (local) | `memory`, `hostkv` | BUILT | Tiered `memory.Manager` (scratch → local → remote policy cascade, write-back journal, negative caching, maturity-dialed remote reads) on `Host.Memory`, backing the `remember`/`recollect`/`forget` DSL verbs. `hostkv` is the substrate: bbolt durable store + LRU scratch. The durable episodic **journal** (`memory/journal.go`) is captured by the `runMemory` bus goroutine and recalled into the per-turn Situation (`runtime/mesa_director.go`). Remote tier = mesa when linked (`mesaclient.AsRemote`), `NopRemote` offline. |
| 7. Routine DSL | `dsl/*` | BUILT | DSL runtime as a package tree: `dsl/token`, `dsl/lex`, `dsl/ast`, `dsl/parser`, `dsl/validator`, `dsl/interp` (interpreter + REPL), `dsl/spec` (action/accessor/event spec), `dsl/conformance`. Routines run in `dsl/interp` on the conductor goroutine. |
| 8. Cognition | `cognition` | BUILT | `cognition/resolve` (recognition: loose text → canonical defs, learned aliases), `cognition/corpus` (shared-knowledge retrieval surface), `cognition/knowledge` (the graded, provenance-tagged world-knowledge ledger), `cognition/goalgraph` (the typed intention graph). The retrieval seam routes to mesa via `mesaclient.AsRetriever`; `cognition.StubClient` is the offline fallback. |
| 9. Brain | `brain` | BUILT (LLM lives in mesa) | The live LLM is `mesa/llm` + `mesa/mesad` (Act/Decide/Chat/Genesis, Sonnet/Haiku/Opus tiering, prompt caching), reached via `mesaclient.AsStrategist` and the `MesaDirector` Act loop. The `brain/` package keeps the `Strategist` interface + `StubStrategist` (offline fallback). See [`brain.md`](brain.md), [`cognition-and-autonomy.md`](cognition-and-autonomy.md). |
| — Limbic | `limbic` | BUILT | System-1 deterministic affect: the mood vector (`limbic.Affect`, feeds pearl's affect predicates) + the Beta(α,β) trust ledger (`limbic.Ledger`, backs `relation_with`). Maintained by the `runLimbic` bus goroutine — no LLM, no tick. |
| — Pearl | `pearl` | BUILT | The onboard policy/quirk engine — the deterministic fast path that decides and gates **without** an LLM. Two seams on the host: `TryDecide` (consulted before the Strategist in `decide()`) and `Gate` (wraps every state-mutating action at the dsl bridge; `POLICY_VETO` on refusal). |
| 10. Persona | `persona` | BUILT | Persona types/schema (HEXACO + Schwartz values + econ + Westworld dials + Directives) + `Validate()` + `CompilePolicy` (persona → `pearl.Table` + affect baseline), shared host↔mesa. Provisioned from mesa at login and compiled onto `Host.Pearl` locally (the table is func-valued — it never crosses the wire). |
| 11. Reveries | — | NOT BUILT | No `reveries` package exists. The believability layer (timing jitter, idle gestures, persona chat) is Phase 5 — TODO `P-6` (engine shape: sibling-goroutine scheduler, **not** an interpreter tick) and `P-9` (ActionArbiter/`is_busy` prerequisite). Design: [`reveries.md`](reveries.md). |
| 12. Runtime | `runtime` | BUILT | The `Host` abstraction: composition of the layers below into one agent. Owns the connection, runs the frame pump (`Connect` → `Run`), decodes frames into world updates + events, and exposes the high-level action methods the DSL builtins call. Also home of the **conductor** (turn loop), the **directors** (MesaDirector/HybridDirector), reactive tiers, and `RunHost` (the per-host composition root shared by `cmd/host` and the cradle daemon). |
| 13. Render | `render` | BUILT | Decoupled, headless software 3D renderer (terrain, scenery, boundaries, entity billboards). Given a host's position + camera params it returns a PNG of what the host sees. `cmd/legacy-cradle -spectate` serves a live viewport. Depends only on `assets`/`facts`/`pathfind` — not on `runtime`. |

Process-level (out-of-band) packages that are not host layers: `cradle`
(fleet registry + control API + web UI), `debughttp` (per-host scriptable
HTTP/WS debug surface), `scenariogen` (manifest library used by the cmd).

## The mesa server (`mesa/`)

> **STATUS: BUILT.** gRPC end to end; the HTTP/REST + "mesa/server,
> store, embed, consolidate" package sketch from earlier drafts never
> existed under those names. Full treatment: [`mesa.md`](mesa.md),
> [`mesa/ARCHITECTURE.md`](mesa/ARCHITECTURE.md),
> [`mesa/PROTOCOL.md`](mesa/PROTOCOL.md).

| Package | Responsibility |
|---|---|
| `mesa/proto` | The gRPC contract (`mesa.proto`, generated stubs checked in). Six services, 28 RPCs: **Game** (Act / Decide / Chat / AnalysisInterpret / ExtractDialog), **Knowledge** (Recall + the `Fetch*` cold-start bootstraps: relationships, goal, knowledge ledger, goal graph), **Journal** (Remember episode stream, RecordObservations firehose, the `Sync*` host→mesa snapshots, ReportMetrics), **KV** (Put/Get/Delete), **Provision** (Fetch persona, Subscribe directive push, Genesis), **Admin** (PutPersonas/Get/List/Delete + PushGoal — operator token, distinct from host tokens). |
| `mesa/mesad` | The server implementation: `act.go` (the Act planner — situation → DSL Move), `genesis.go` (the one heavy Opus-at-login session compile), `admin.go` + `auth.go` (per-host bearer tokens; a host can never rewrite another host's identity), `ltm.go` (the durable store), `cron.go` + `cron_insight.go` (the tiered distillation crons: Tier-0 no-LLM dedup/GC, Tier-1 Haiku batched claim extraction, Tier-2 Sonnet on the rare escalation queue only), `catalog.go` (world name-sets for static arg validation), `dslmanual.go` (the cached DSL manual prompt prefix). |
| `mesa/llm` | The Anthropic client. Three tier instances wired by `cmd/mesad` flags: `-act-model` (Sonnet), `-decide-model` (Haiku), `-genesis-model` (Opus). Prompt caching on the stable prefixes. |
| `mesa/client` | `GRPCClient` + the host-side adapters: `AsStrategist` (decide() → Game.Decide), `AsRetriever` (recall → Knowledge), `AsRemote` (memory tier → Journal/KV), per-host creds/hostkey. |
| `mesa/embed` | Voyage embeddings client. When wired, `Knowledge.Recall` ranks by pgvector cosine similarity; without it, recall degrades to Postgres lexical/recency ranking. |

Durable state is **Postgres** (`-db` / `DATABASE_URL`), schema owned by
`ltm.go` (idempotent migrations): `episodes` (with a pgvector `embedding`
column), `observations`, `relationships`, `knowledge`, `goal_graphs`, `goals`,
`personas` (persona JSON + cooked prose card + generated projection columns),
`kv`, `metrics`. Without a DB-reachable LTM, mesad falls back to a volatile
in-memory episode map.

## Control flow (the host's main loop)

The host is event-driven, not tick-driven. The RSC server runs ~640ms ticks;
the host reacts to inbound packets as they arrive. The frame pump
(`runtime.Host.Run`, `runtime/host.go`) does **one** thing on its goroutine:
pump inbound frames into world state and onto the event bus. It spawns four
background bus/timer goroutines, none of which is a tick and none of which
can block it:

```go
// runtime/host.go (paraphrased from Run)
go h.heartbeatLoop(heartCtx) // periodic keepalive
go h.runLimbic(heartCtx)     // System-1: events → affect + trust ledger (no LLM)
go h.runMemory(heartCtx)     // episodic capture: salient events → durable journal
go h.runMetrics(heartCtx)    // telemetry: counters + snapshots → mesa, 60s cadence
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case frame, ok := <-h.conn.Recv(): // one decoded RSC frame
        if !ok {
            return h.conn.Err()        // session closed → terminal
        }
        h.handleFrame(frame)           // decode → world.Apply → bus.Publish
    }
}
```

`handleFrame` is where the real work happens:

1. **Decode** the frame. Most opcodes go through `v235.DecodeInbound` and produce one typed event; a handful that carry many records (player/NPC coord updates, the update-players/NPCs lists) are special-cased to emit several events.
2. **Apply** each event to the `world` mirror (`world.Apply`).
3. **Derive synthetic events** by diffing pre/post `Apply` snapshots — `ItemGained`, `XPGain`, the `TargetDied`/combat-round edges, and the per-player level/equipment-change edges — so routines can subscribe to gains and deaths rather than poll.
4. **Publish** every event on the `event.Bus`. The conductor's routine watchers, the limbic/memory/metrics goroutines, and the debughttp recorder all subscribe.

**The conductor** (`runtime/conductor.go`) is the autonomy spine — the
"agent driver" earlier drafts marked planned. It runs **concurrently** with
`Run`, on its own goroutine, composed by the caller (`runtime.RunHost`):

```go
go host.Run(ctx)   // frame pump: must be started first
conductor.Run(ctx) // turn loop: select Intent → run routine → observe Outcome
```

Each turn it asks its `Director` for the next `Intent` (a routine file or
inline DSL source), runs it in `dsl/interp` to completion, and feeds the
`Outcome` back into the next selection. It is interruptible at and within
turn boundaries (`Pause` cancels the in-flight routine's context; ANALYSIS
mode rides the same path). There is still **no tick**: turns are driven by
routine completion, not a timer.

**Strategist invocation** happens at two altitudes, neither on a timer:

- **Per turn (Act):** the production Director is a `HybridDirector` wrapping
  a `MesaDirector` (`runtime/runhost_bootstrap.go`). The MesaDirector
  snapshots the host's state into a `mesaclient.Situation`
  (`situation()` in `runtime/mesa_director.go` — world, goal, journal recall,
  knowledge, transcript, stuck/fail-streak detection), calls `Game.Act`
  ("what do I do now?"), and turns the returned DSL Move into a conductor
  Intent. The HybridDirector serves cheap local routine-library turns and
  escalates to the wrapped MesaDirector at decision points.
- **In-routine (`decide()`):** the DSL decision builtins run a local-first
  ladder — `Pearl.TryDecide` (deterministic, sub-ms) → the bounded
  `decisionCache` (LRU+TTL memo of past verdicts) → `Host.Strategist`
  (`Game.Decide`, Haiku). Offline hosts keep the `StubStrategist`.

Reactive behaviors stay inline and deterministic: the fatigue sleep-screen
captcha is answered in `handleFrame` (the hardcoded `sleepWord`); doors/gates
are opened en route inside the `go_to`/`WalkTo` traversal flow
(`runtime/pathing.go`), with server refusals classified
(`runtime/obstacles.go`) and ledgered (`blockedEdges`, TTL-bounded) so a
locked door is tried twice, remembered, and routed around. The speed-2
reactive tier (`runtime/reactive.go`: per-speaker conversation windows +
keyword-ladder triggers) and the speed-3 observation firehose
(`runtime/observation.go`) are bus-fed and bounded.

## Mesa's role in a host's session

All of this traffic is live today (gRPC, per-host bearer auth). The host
runs degraded-but-functional without mesa: local tiers + pearl only, stub
strategist, no persona provisioning.

| When | RPC | What |
|---|---|---|
| Login | `Provision.Fetch` | Persona down; host compiles it locally (`persona.CompilePolicy` → `Host.Pearl` + affect baseline). |
| Login | `Provision.Genesis` | One heavy Opus compile over the host's full history → this session's goal, mood baseline, attention keyword ladder. `-goal` overrides; failure falls back to the persona north-star. |
| Login | `Knowledge.Fetch*` | Cold-start bootstrap of trust ledger / goal / knowledge ledger / goal graph when the local store is empty (mesa is authoritative; bbolt is the warm start). |
| Always-on | `Provision.Subscribe` | Live directive push (e.g. `mesa-ctl`/Admin `PushGoal` → a soft runtime goal override, no restart). |
| Per conductor turn | `Game.Act` | Situation → DSL Move. The cost-dominant call (Sonnet, cached manual prefix). |
| On pearl-miss `decide()` | `Game.Decide` | Narrow option pick (Haiku), memoized by the host's decision cache. |
| On being spoken to | `Game.Chat` / `Game.ExtractDialog` | The cheap reactive social path, off the Act loop. |
| Background | `Journal.Remember` / `RecordObservations` | Episode stream + salience-gated observation firehose (cron fodder). |
| Background | `Journal.Sync*` | Trust-ledger / goal / knowledge / goal-graph snapshots up (host-authoritative mirror). |
| Every 60s | `Journal.ReportMetrics` | Telemetry counters + state snapshots (`runtime/telemetry.go`). |
| Mesa-side crons | — | Consolidation (Haiku, batched) folds observations into the knowledge ledger; the insight cron (Sonnet, escalation queue only) reconciles contradictions and grows the goal graph. |

Working memory, the world mirror, reactive logic, pearl gating, limbic
affect, and the in-progress routine all stay local — the compute-locality
split: host = isolated local compute; mesa = LLM/RAG/LTM/crons/persona.

## The integration pattern: sibling fields, nil-guarded, stub-defaulted

How upper layers attach to the host without touching the frame pump — the
pattern every layer above the body uses (and the one new layers should use):

- **Exported sibling fields on `*runtime.Host`**, interface- or
  pointer-typed: `Strategist`, `Retriever`, `Corpus`, `Resolver`, `Pearl`,
  `Memory`, plus the shared-singleton options (`Facts`, `Landscape`,
  `WorldOracle`). `runtime.New` defaults `Strategist`/`Retriever` to
  deterministic stubs (`brain.StubStrategist`, `cognition.StubClient`) and
  the rest to nil.
- **Nil = feature absent = prior behavior.** Every consumer nil-guards:
  nil `Pearl` means no gating, nil `Memory` means the memory verbs report
  `NOT_IMPLEMENTED`, nil `WorldOracle` means the map verbs report "no map
  data loaded", nil `Resolver` lazily builds a session-only one. `go test
  ./...` stays green with nothing wired.
- **One composition root swaps the fields** after `New` and before
  `Connect`: `runtime.RunHost` (`runhost_bootstrap.go`) builds the per-host
  stores, dials mesa, and sets `host.Strategist = mesaclient.AsStrategist(...)`,
  `host.Retriever = mesaclient.AsRetriever(...)`, `host.Memory = memory.New(...)`,
  `host.Resolver = resolve.New(...)`, and (via persona provisioning)
  `host.Pearl`. Both `cmd/host` and the cradle daemon go through this one
  function.
- **Background consumers are bus subscribers**, spawned as siblings of
  `heartbeatLoop` inside `Run` (`runLimbic`, `runMemory`, `runMetrics`) or
  composed beside it (the conductor, `socialReflex`, the debughttp
  recorder). The bus drops on a full subscriber buffer — background work can
  run slow, but it can never stall the frame pump.
- **Shared vs per-host:** immutable singletons (`*facts.Facts`,
  `*pathfind.Landscape`, `*worldmap.Oracle`) are loaded once per process and
  shared by pointer across the fleet (`runtime.SharedDeps`); stateful pieces
  (bus, stores, mesa connection, pearl engine) are strictly per-host.

## Cross-cutting: observability

What exists (all BUILT; fuller catalog in [`observability.md`](observability.md)):

- **The cradle web UI + HTTP API** (`cradle/api.go`, `webui.go`): fleet
  list/status/pause/resume/stop, live routine source with the executing line
  highlighted, the Thoughts/decision stream, and a per-host debug proxy —
  consumed identically by the browser and `cradle-ctl`.
- **`debughttp`**: per-host scriptable control plane — `GET /state`,
  `GET /events`, `/ws` thought stream, `POST /eval` (one DSL line),
  `POST /script`, `/mind` (affect/relationships/knowledge views) — with every
  bus event recorded to an in-memory ring + a durable JSONL file.
- **The decision stream** (`runtime/decisions.go`): every AgentThought
  (director decisions, detours, stalls, goal lifecycle) appends to
  `decisions.jsonl` in the host's data dir and mirrors to mesa — the
  overnight-soak trace.
- **Host telemetry** (`runtime/telemetry.go`): counters (pearl vetoes, agent
  turns) + state snapshots (memory hit/miss, ledger/journal sizes, uptime)
  reported to mesad's `metrics` table every 60s.
- **Server-message logging**: `runtime` logs every player-facing server
  message at INFO (`logServerMessage`) so a sweep can grep a host's output.
- **The live visual tap**: `cmd/legacy-cradle -spectate` (software-rendered
  browser viewport).

There is **no `obs` package** and no per-call LLM accounting — nothing
tracks tokens or dollars today. Cost/token ledger = TODO `O-9`; analytics
dashboards = TODO `O-7`.

## Cross-cutting: cohorts

Reality today: `cohort_id` lives on the persona Cornerstone
(`persona/persona.go`, `generation_meta` — analytics-only grouping; the
cohort = **launch batch**, archetype = personality), is projected as a
generated column on mesad's `personas` table (`ltm.go`), and is settable at
runtime via `mesa-ctl set`. No cohort-scoped joins, population analytics, or
cohort feature flags exist yet — that is TODO `O-7` (analytics) and `R-1`
(the wiki-RAG cohort experiment, for which the substrate now exists).

## What this isn't

For clarity:

- **Not a Go reimplementation of OpenRSC.** The server is OpenRSC, unchanged. Westworld is the client/orchestration/memory side.
- **Not a Java reimplementation port.** No code is mechanically translated from OpenRSC. The wire protocol is reverse-engineered from `Payload235Parser.java` / `Payload235Generator.java` as reference but rewritten cleanly in idiomatic Go.
- **Not a general bot framework.** The architecture is opinionated for our specific research goals — believability, memory, social emergence. A user wanting a "click these resources" bot would find this overkill.

## Why this shape

The architecture follows from the research goals (see [research-goals.md](research-goals.md)):

- **Memory as a central service** (mesa's Postgres LTM) enables population-scale research queries that per-host local stores would not.
- **Reveries as load-bearing** (not polish) follows directly from the "no drones" principle — still the largest unbuilt slice (TODO `P-6`).
- **Tiered brain routing** is cost-driven: ~500 hosts × 24/7 is only economically feasible with Haiku for the bulk (Decide, the consolidation cron), Sonnet where authoring quality matters (Act, the insight cron), and Opus only at the rare session-genesis compile.
- **Deterministic DSL execution with a separate believability layer** keeps reproducible behavior and humanizing variance in different substrates (the reverie engine drives effector methods from a scheduler — it does not hook the interpreter).
- **Cohort-first design** in the persona schema is what makes the project research rather than a demo.
