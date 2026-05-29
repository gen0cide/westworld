# Architecture

> **Reading guide.** This doc slices the system by Go package (the
> implementation view). If you're new and want the *AI-perspective*
> view of the host — body → senses → routines → cognition → persona —
> read [`layers.md`](layers.md) first, then come back here for the
> package-level details.

## The three binaries

Westworld is a Go monorepo at `github.com/gen0cide/westworld` that builds three independent binaries:

| Binary | Purpose | Lifespan |
|---|---|---|
| `cmd/cradle` | The per-host runtime. One process = one host. Embeds a packet-level RSC client, a state mirror, a DSL interpreter, a brain, a memory layer (local + mesa-backed), and a reverie augmentation system. | Long-running (days/weeks). Crash-recoverable via mesa-persisted state. |
| `cmd/mesa` | The shared memory + RAG service. Holds per-host memory (episodic, relational, reflective), shared knowledge (rsc.wiki, chat corpus), and embeddings (Voyage 3). Postgres + pgvector backed. Exposes HTTP API. | Always-on infrastructure. |
| `cmd/delos` | The swarm orchestrator. Manages many cradle hosts (lifecycle, supervision, scaling). Also serves the "technician tablets" web UI for observability — live state inspection, chain-of-thought capture, cohort analytics. | Always-on operations. |

Each binary is built from packages that may be shared across binaries (e.g., the `proto` package is used by `cmd/cradle` directly and by `cmd/delos` for replay/debugging tools). No `internal/` discipline — packages are explicit imports.

## The layer cake (within `cradle`)

Reading bottom-up. Each layer depends only on layers below. Each layer is one Go package at the top level of the repo.

| Layer | Package | Responsibility |
|---|---|---|
| 1. Wire protocol | `proto/v235` | RSC packet encode/decode (Payload235-equivalent). Pure functions, no I/O. |
| 2. Session | `session` | TCP connection lifecycle: handshake, RSA login, packet read/write goroutines, heartbeat, reconnect. |
| 3. World state | `world` | State mirror — own player, nearby entities, ground items, inventory, regional walkability. Updated by inbound packets. |
| 4. Actions | `action` | Atomic operations: `Walk`, `Attack`, `Eat`, `TalkTo`, etc. Each validates preconditions against `world`, sends outbound packets via `session`. |
| 5. Events | `event` | Pub/sub bus for things that happen *to* the host. Inbound packet handlers publish typed events; layers above subscribe. |
| 6. Memory (local) | `memory` | Working memory ring buffer (~50 most recent events). Thin local cache; everything else is in mesa. |
| 7. Script | `script` | DSL runtime: lexer, parser, AST, interpreter, sandbox. Routines run here. |
| 8. Cognition | `cognition` | Retrieval client to mesa: knowledge RAG (wiki), per-host memory queries, structured world facts. |
| 9. Brain | `brain` | LLM strategist. Routes decisions by class (Sonnet for strategic/script-gen, Haiku for routine/tactical/chat). Anthropic-specific concrete impl behind a `Brain` interface. |
| 10. Reveries | `reveries` | Cross-cutting believability augmentations. Injected at every action call site via `reverie.tick()`. Persona-driven, emotional-state-aware. |
| 11. Runtime | `runtime` | The `Host` abstraction: composition of layers 2-10 into a single coherent agent. Owns the goroutine, runs the control loop, exposes `Connect / Run / Stop`. |
| 12. Persona | `persona` | Persona types/schemas. Lives separately so external tools and admins can author personas without depending on runtime internals. |

## The mesa server (within `cmd/mesa`)

| Package | Responsibility |
|---|---|
| `mesa/server` | HTTP API surface, request validation, auth (bearer tokens per host), routing. |
| `mesa` (root) | Shared request/response types, error codes (used by both server and client SDK). |
| `mesa/client` | Client SDK for the HTTP API. Imported by `cradle`'s `cognition` package. Has a mock impl for local dev. |
| (server-internal) `store` | Postgres + pgvector schema, migrations, query layer. |
| (server-internal) `embed` | Voyage 3 client for embeddings. Batches small. |
| (server-internal) `consolidate` | Background jobs: memory decay, reflection generation, compression of low-salience episodes. |

## Control flow (one host, one tick of the agent loop)

The host is fundamentally event-driven, not tick-driven. The RSC server has 640ms ticks; the host responds to inbound packets and timers as they arrive.

```go
for {
    select {
    case <-ctx.Done():
        return ctx.Err()

    case pkt := <-h.conn.Inbound():
        h.world.Apply(pkt)                          // update state mirror
        for _, ev := range eventsFromPacket(pkt) {  // derive typed events
            h.events.Publish(ev)
        }

    case ev := <-h.urgentEvents:                    // reactive conditions
        h.handleUrgent(ctx, ev)                     // HP low → eat, etc.

    case <-h.script.Step():                         // advance running routine
        h.script.Tick(ctx)                          //   includes reverie.tick() between actions

    case <-h.chatDrain.C:                           // opportunistic chat send
        h.chat.MaybeSend()                          //   queued messages drain when idle

    case <-h.strategistTick.C:                      // periodic re-strategize
        if h.brain != nil && h.script.Idle() {
            h.consultBrain(ctx)
        }

    case <-h.reverieTick.C:                         // standalone reverie injection
        h.reveries.MaybeInject(h.actions)
    }
}
```

Key properties:

- **Reactive priority**: urgent events (HP low, combat started, fatigue critical) preempt running routines.
- **Brain is optional**: pure-script hosts skip the strategist case entirely.
- **Reveries hooked twice**: once inside `script.Tick` (between every routine action) and once as a standalone timer (idle wander, spontaneous chat). The first is in-routine flavoring; the second is "you've been idle long enough to do something unprompted."
- **Chat is a separate channel**: the chat worker drains its queue opportunistically. Routines don't block on chat.

## Mesa's role in a host's tick

Mesa is hit on slower cadences than the agent loop:

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

Every Brain call is recorded to mesa with: prompt input, response output, model used, token counts, cost estimate, latency. This is the "chain-of-thought" capture — admin tools can replay any host's reasoning.

Every published event flows through `event.Bus` and is captured in a structured log (via `obs` package). Delos can subscribe to a live host's event stream for the "tap" UI feature.

## Cross-cutting: cohorts

Every host has a `cohort_id` assigned at registration. Mesa scopes all writes/reads by host but joins to cohort_id on the relationship and event tables. This enables population-level analytics ("did cohort_A trade more than cohort_B?") and cohort-based feature flags ("only cohort_A has wiki RAG access").

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
