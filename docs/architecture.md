# Architecture

> **Reading guide.** This doc slices the system by Go package (the
> implementation view). If you're new and want the *AI-perspective*
> view of the host — body → senses → routines → cognition → persona —
> read [`layers.md`](layers.md) first, then come back here for the
> package-level details. One cross-cutting frame from that doc to keep
> in mind: the host's seven layers are its **in-band** consciousness
> (sense → think → act), and around them sits an **out-of-band**
> management plane it never experiences — the cradle, mesa, and delos
> (see [`layers.md`](layers.md) for the in-band/out-of-band framing).

## The three binaries

Westworld is a Go monorepo at `github.com/gen0cide/westworld` that builds three independent binaries:

| Binary | Status | Purpose | Lifespan |
|---|---|---|---|
| `cmd/cradle` | **IMPLEMENTED** | The per-host runtime. One process = one host. Embeds a packet-level RSC client, a state mirror, a DSL interpreter, the cognition/brain hooks (currently stubs), and a `-spectate` live software-rendered viewport. | Long-running (days/weeks). |
| `cmd/mesa` | **PLANNED** (Phase 2.6/3) — `cmd/mesa/` is empty | The shared memory + RAG service. Holds per-host memory (episodic, relational, reflective), shared knowledge (rsc.wiki, chat corpus), and embeddings (Voyage 3). Postgres + pgvector backed. Exposes HTTP API. | Always-on infrastructure. |
| `cmd/delos` | **PLANNED** (Phase 6) — `cmd/delos/` is empty | The swarm orchestrator. Manages many cradle hosts (lifecycle, supervision, scaling). Also serves the "technician tablets" web UI for observability — live state inspection, chain-of-thought capture, cohort analytics. | Always-on operations. |

Three more binaries exist today as tooling: `cmd/parsecheck` (protocol parse harness), `cmd/rendertest` (renderer harness), and `cmd/scenariogen` (the live-test scenario runner).

Each binary is built from packages that may be shared across binaries (e.g., the `proto` package is used by `cmd/cradle` directly and will be reused by `cmd/delos` for replay/debugging tools). No `internal/` discipline — packages are explicit imports.

## The layer cake (within `cradle`)

Reading bottom-up. Each layer depends only on layers below. Each layer is one (or, for `dsl`/`cognition`, one tree of) Go package at the top level of the repo.

**STATUS legend:** IMPLEMENTED = working; STUB = interface present, only a canned/no-op concrete impl exists; EMPTY = directory exists with no `.go` files (planned).

| Layer | Package | Status | Responsibility |
|---|---|---|---|
| 1. Wire protocol | `proto/v235` | IMPLEMENTED | RSC packet encode/decode (Payload235-equivalent). Pure functions, no I/O. ISAAC cipher, RSC string compression, the multi-event coords decoders. |
| 2. Session | `session` | IMPLEMENTED | TCP connection lifecycle: handshake, RSA login, packet read/write goroutines, heartbeat. |
| 3. World state | `world` | IMPLEMENTED | State mirror — own player, nearby entities, ground items, inventory, skills/xp. Updated by inbound packets via `Apply`. |
| — Static knowledge | `facts`, `assets` | IMPLEMENTED | Read-only world data (item/npc/object defs) and the `.orsc` archive decoders. One `*facts.Facts` is loaded once per process and shared by pointer across hosts. |
| — Pathfinding | `pathfind` | IMPLEMENTED | BFS routing over the landscape walkability grid; feeds `WalkTo`. |
| 4. Actions | `action` | IMPLEMENTED | Atomic operations: `Walk`, `Attack`, `Eat`, trade/bank/magic, etc. Each builds outbound packets; `runtime.Host` validates preconditions against `world`. |
| 5. Events | `event` | IMPLEMENTED | Pub/sub bus for things that happen *to* the host. The frame handler publishes typed events (incl. synthetic deltas: `ItemGained`, `XPGain`, `TargetDied`); layers above subscribe. |
| 6. Memory (local) | `memory` | EMPTY (planned) | Working memory ring buffer. Today the equivalent recent-event state lives on `world.Recent`; the dedicated package is not yet written. |
| 7. Routine DSL | `dsl/*` | IMPLEMENTED | DSL runtime as a package tree: `dsl/token`, `dsl/lex`, `dsl/ast`, `dsl/parser`, `dsl/validator`, `dsl/interp` (interpreter + REPL), `dsl/spec` (action/accessor/event spec), `dsl/conformance`. Routines run in `dsl/interp`. *(This was called the `script` layer in earlier drafts; the package is `dsl`.)* |
| 8. Cognition | `cognition` | STUB + PARTIAL | Retrieval surface to mesa (`cognition.Client`). The interface is real; the only concrete impl is `StubClient` (canned bundles). Two sub-packages are real and used by routines today: `cognition/resolve` (player-text → facts recognition) and `cognition/corpus` (rsc.wiki RAG corpus). *(planned: `PrepareDecision` assembles the brain's `DecisionRequest` incl. an earned, progressively-disclosed DSL surface.)* |
| 9. Brain | `brain` | STUB | LLM strategist behind a `brain.Strategist` interface. Only `StubStrategist` (deterministic canned decisions) exists; the Anthropic-backed, tiered (Sonnet/Haiku) impl lands in Phase 4. *(planned: `DecisionRequest`/`Decision` contract; tiered routing; a `TrustGrade` class.)* |
| 10. Reveries | `reveries` | EMPTY (planned) | Cross-cutting believability augmentations (timing jitter, idle wander, persona chat). Phase 5. *(planned: trait-derived weights + a lightweight mood state.)* |
| 11. Runtime | `runtime` | IMPLEMENTED | The `Host` abstraction: composition of the layers below into one agent. Owns the connection, runs the control loop (`Connect` → `Run`), decodes frames into world updates + events, and exposes the high-level action methods the DSL builtins call. |
| 12. Render | `render` | IMPLEMENTED | Decoupled, headless software 3D renderer (terrain, scenery, boundaries, entity billboards). Given a host's position + camera params it returns a PNG of what the host sees. `cmd/cradle -spectate` serves a live viewport. Depends only on `assets`/`facts`/`pathfind` — not on `runtime`. Shipped 2026-05-30. |
| 13. Persona | `persona` | EMPTY (planned) | Persona types/schemas, split out so external tools can author personas without runtime deps. Design deferred. *(planned: the consolidated HEXACO-based schema + curiosity/attention dials + trust ledger.)* |

## The mesa server (within `cmd/mesa`)

> **STATUS: PLANNED.** None of the packages below exist yet — `mesa/` is an empty
> directory and `cmd/mesa/` has no `.go` files. This section is the target design
> (Phase 2.6 stands up the knowledge corpus; Phase 3 adds per-host memory).

| Package | Responsibility |
|---|---|
| `mesa/server` | HTTP API surface, request validation, auth (bearer tokens per host), routing. |
| `mesa` (root) | Shared request/response types, error codes (used by both server and client SDK). |
| `mesa/client` | Client SDK for the HTTP API. Imported by `cradle`'s `cognition` package. Has a mock impl for local dev. |
| (server-internal) `store` | Postgres + pgvector schema, migrations, query layer. |
| (server-internal) `embed` | Voyage 3 client for embeddings. Batches small. |
| (server-internal) `consolidate` | Background jobs: memory decay, reflection generation, compression of low-salience episodes. |

## Control flow (the host's main loop)

The host is event-driven, not tick-driven. The RSC server runs ~640ms ticks; the host reacts to inbound packets as they arrive. The actual loop (`runtime.Host.Run`, `runtime/host.go`) is deliberately small — it does **one** thing: pump inbound frames into world state and onto the event bus.

```go
// runtime/host.go (paraphrased)
func (h *Host) Run(ctx context.Context) error {
    go h.heartbeatLoop(heartCtx)   // periodic keepalive on its own goroutine
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case frame, ok := <-h.conn.Recv():   // one decoded RSC frame
            if !ok {                          // session closed → terminal
                return h.conn.Err()
            }
            h.handleFrame(frame)              // decode → world.Apply → bus.Publish
        }
    }
}
```

`handleFrame` is where the real work happens:

1. **Decode** the frame. Most opcodes go through `v235.DecodeInbound` and produce one typed event; a handful that carry many records (player/NPC coord updates, the update-players/NPCs lists) are special-cased to emit several events.
2. **Apply** each event to the `world` mirror (`world.Apply`).
3. **Derive synthetic events** by diffing pre/post `Apply` snapshots — `ItemGained`, `XPGain`, and the `TargetDied`/combat-round edges — so routines can subscribe to gains and deaths rather than poll.
4. **Publish** every event on the `event.Bus`. Routines (in `dsl/interp`) and any other subscribers react.

Things the loop does **not** do (contrary to earlier drafts): there is no `script.Step()`/`script.Tick()` channel, no `reveries.MaybeInject()`, no `chat.MaybeSend()`, and no `strategistTick`. Those fields do not exist on `Host`.

**Planned (Phase 4):** two NEW goroutines join the host, both siblings of `heartbeatLoop`, neither a tick — (a) an **agent driver** that, at goal boundaries (a routine returns / is interrupted / the host idles), assembles a decision request via the cognition layer, calls the brain (`RunRoutine`/`WriteRoutine`/`DirectAction`/`Idle`), and runs the result; (b) a **cognitiveLoop** that subscribes to the event bus to appraise significant events and update affect/relational state in the background. These are invoked at boundaries / on events, not polled per tick — the no-tick invariant above still holds. (Design detail in the `_research` notes.)

Where the other layers actually live:

- **Routines** run inside `dsl/interp`, driven by the interpreter; they call the host's high-level action methods (`WalkTo`, attack/trade/bank builtins) and subscribe to bus events (`on item_gained(...)`, `on target_died(...)`). They are not pumped by a channel in `Run`.
- **Reactive behaviors** (e.g. auto-answering the fatigue sleep-screen captcha with the hardcoded word, auto-opening a door that stalls a walk) are handled inline — in `handleFrame` for the sleep-screen, inside `WalkTo` for doors — not via a separate urgent-events channel.
- **The strategist + retriever** are not on a timer at all. They are invoked on demand by DSL cognition builtins (`contemplate_reality`, `decide`, `recall`, …) through the `Host.Strategist` / `Host.Retriever` hooks, which default to the stub implementations.
- **Reveries** are unimplemented (the `reveries` package is empty); the believability layer is Phase 5.

## Mesa's role in a host's tick

> **STATUS: PLANNED.** Mesa is not built, so today none of this traffic happens —
> the host runs entirely locally with stub cognition/brain. This table is the
> target once mesa lands (Phases 2.6 → 3). The one piece that exists today is the
> in-process knowledge corpus (`cognition/corpus`, loaded from the rsc.wiki dump),
> which Phase 2.6 will move behind mesa.

Mesa is intended to be hit on slower cadences than the agent loop:

| Operation | Frequency | Why |
|---|---|---|
| Episode write | per significant event | Single small HTTP POST |
| Working memory → episodic flush | every ~30s | Batch-write recent buffer |
| Relational query (nearby players) | per strategist call | One batch request for all visible players |
| Knowledge retrieval (wiki RAG) | per strategist call when needed | Vector top-K against shared mesa-side index |
| Reflection generation | every ~30 active minutes | Backend job; mesa may queue + LLM-call asynchronously |
| Decay/compression sweep | hourly background | Per-host, run by mesa |

Working memory, current world state, reactive logic, and the in-progress routine all stay local. Anything that's queried during a strategist call (a few times per minute at most) goes through mesa.

## Cross-cutting: observability

> **STATUS: PARTIAL.** The `obs` package is empty (planned), and Brain calls
> aren't recorded to mesa yet (mesa doesn't exist). What works today:

- **Today:** every published event flows through `event.Bus`. The `dsl/interp`
  layer has observability hooks (`dsl/interp/observability.go`), and `runtime`
  logs every player-facing server message at INFO (`logServerMessage`) so the
  scenario sweep can grep a host's stdout. The `cmd/cradle -spectate` viewport is
  the live visual "tap."
- **Planned:** once mesa + delos land, every Brain call will be recorded to mesa
  (prompt, response, model, token counts, cost, latency) for "chain-of-thought"
  replay, and delos will subscribe to a live host's event stream for the "tap" UI.
  A dedicated `obs` package will hold the structured-log surface.

## Cross-cutting: cohorts

> **STATUS: PLANNED.** Cohorts live in mesa, which isn't built. Hosts currently
> run un-cohorted. This is the intended design:

Every host gets a `cohort_id` at registration. Mesa scopes all writes/reads by host but joins to cohort_id on the relationship and event tables. This enables population-level analytics ("did cohort_A trade more than cohort_B?") and cohort-based feature flags ("only cohort_A has wiki RAG access").

## What this isn't

For clarity:

- **Not a Go reimplementation of OpenRSC.** The server is OpenRSC, unchanged. Westworld is the client/orchestration/memory side.
- **Not a Java reimplementation port.** No code is mechanically translated from OpenRSC. The wire protocol is reverse-engineered from `Payload235Parser.java` / `Payload235Generator.java` as reference but rewritten cleanly in idiomatic Go.
- **Not a general bot framework.** The architecture is opinionated for our specific research goals — believability, memory, social emergence. A user wanting a "click these resources" bot would find this overkill.

## Why this shape

The architecture follows from the research goals (see [research-goals.md](research-goals.md)):

- **Memory as a central service** (mesa) enables population-scale research queries that per-host sqlite would not.
- **Reveries as load-bearing** (not polish) follows directly from the "no drones" principle.
- **Tiered brain routing** is cost-driven: 500 hosts × 24/7 is only economically feasible with Haiku for routine decisions.
- **DSL with explicit reverie call sites** keeps both deterministic execution and humanizing variance in one substrate.
- **Cohort-first design** in mesa is what makes the project research rather than a demo.
