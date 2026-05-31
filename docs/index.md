# Westworld — Documentation Index

**Start here.** Westworld is a Go monorepo (`github.com/gen0cide/westworld`) building an
LLM-driven RuneScape Classic **bot swarm** — a believable, autonomous ~500-host
population — as a research substrate for studying long-horizon agent behaviour,
emergent community, ethics-without-observation, and believability at scale.
See [`../README.md`](../README.md) for the project pitch and the four research
questions; [`research-goals.md`](research-goals.md) for the success metrics.

Each host is one `cmd/cradle` process: it speaks the RSC wire protocol, mirrors
world state, and runs a deterministic **routine DSL** for behaviour. An optional
**software renderer** reconstructs what a host sees. The cognition/persona/swarm
layers are mostly design-only today (see status below).

---

## Subsystem status matrix

Be precise about what is real. Status legend: **IMPLEMENTED** (working) ·
**PARTIAL** · **STUB** (interface + canned/no-op impl) · **EMPTY** (package/dir
exists, no Go) · **PLANNED** (design only).

| Subsystem | Package(s) | Status | Doc |
|---|---|---|---|
| Wire protocol (RSC Payload235) | `proto/v235`, `session` | **IMPLEMENTED** | [protocol.md](protocol.md), [world-state.md](world-state.md) |
| World-state mirror | `world`, `event` | **IMPLEMENTED** | [world-state.md](world-state.md) |
| Static game data | `facts`, `assets`, `pathfind` | **IMPLEMENTED** | [architecture.md](architecture.md) |
| Action bridge | `action` | **IMPLEMENTED** | [lang/actions.md](lang/actions.md) |
| Routine DSL | `dsl/*` | **IMPLEMENTED** (some constructs parsed-not-yet-executed — see doc) | [dsl.md](dsl.md), [lang/](lang/README.md) |
| Host runtime / net loop | `runtime` | **IMPLEMENTED** | [architecture.md](architecture.md) |
| **Render engine + spectator** | `render`, `cmd/cradle/spectate.go` | **IMPLEMENTED** (decoupled) | [render-engine.md](render-engine.md) |
| **Scenario system** (live-test) | `cmd/scenariogen`, `examples/scenarios` | **IMPLEMENTED** | [scenarios.md](scenarios.md) |
| Cognition (context assembly) | `cognition` | **PARTIAL** (`resolve/` shipped; `Client`=STUB) | [cognition.md](cognition.md) |
| Brain (LLM strategist) | `brain` | **STUB** (`StubStrategist`) | [brain.md](brain.md) |
| Memory (episodic/reflective) | `memory` | **EMPTY** (design) | [memory.md](memory.md) |
| Personas / identity | `persona` | **PLANNED** (design) | [personas.md](personas.md) |
| Reveries (believable variance) | `reveries` | **EMPTY** (design) | [reveries.md](reveries.md) |
| Swarm orchestrator | `mesa`, `cmd/mesa` | **EMPTY/PLANNED** | [mesa.md](mesa.md) |
| Observability | `obs`, `cmd/delos` | **EMPTY/PLANNED** | [observability.md](observability.md) |

The host binary `cmd/cradle` is **IMPLEMENTED**; `cmd/mesa` and `cmd/delos` are
empty (planned). Tooling binaries `cmd/rendertest`, `cmd/scenariogen`,
`cmd/parsecheck` are real.

---

## Reading order for a new agent

1. [`../README.md`](../README.md) + [`research-goals.md`](research-goals.md) — what & why.
2. **This file** — the map + status.
3. [`architecture.md`](architecture.md) / [`layers.md`](layers.md) — how the pieces fit; the host control loop.
4. [`state.md`](state.md) — where the project actually is right now (refresh-on-arrival).
5. Then the subsystem doc(s) for whatever you're touching (matrix above).
6. [`phases.md`](phases.md) / [`tasks.md`](tasks.md) — the roadmap & task plan.
7. [`questions-and-decisions.md`](questions-and-decisions.md) — why things are the way they are.

## Doc map by topic

- **Behaviour scripting (routine DSL):** [`dsl.md`](dsl.md) + the [`lang/`](lang/README.md)
  reference (`overview`, `syntax`, `actions`, `events`, `state`, `api`, `knowledge`,
  `repl`, `writing-routines`). Authoring routines: start at [`lang/writing-routines.md`](lang/writing-routines.md).
- **Live testing:** [`scenarios.md`](scenarios.md) — the scenario system (author →
  generate → run → triage), multi-host orchestration, the YAML schema.
- **Rendering:** [`render-engine.md`](render-engine.md) — the software rasteriser +
  live spectator + authenticity invariants. (Superseded plan: `render-port-plan.md`.)
- **Server/protocol (for agents working on the wire or world state):**
  [`protocol.md`](protocol.md), [`world-state.md`](world-state.md), [`server-config.md`](server-config.md).
- **The AI/persona layer (mostly design today):** [`cognition.md`](cognition.md),
  [`brain.md`](brain.md), [`memory.md`](memory.md), [`personas.md`](personas.md),
  [`reveries.md`](reveries.md), [`mesa.md`](mesa.md), [`lang/thought-architecture.md`](lang/thought-architecture.md).

## Quickstart

```bash
cd ~/Code/westworld
go build ./...                     # build everything
go test ./...                      # run tests

# Run a host against a local OpenRSC server (the WESTWORLD_PASSWORD env supplies the
# OpenRSC test password — see server-config.md; never print/commit it):
set -a; . ./.local.env; set +a
go run ./cmd/cradle -username bernard -server localhost:43594 -facts ~/Code/openrsc -routine examples/routines/<name>.routine

# Live render viewport (decoupled): add -spectate -spectate-addr localhost:8089, open the URL.
# Live-test scenarios: see scenarios.md (cmd/scenariogen + the runner scripts).
```

> **Accuracy convention:** every subsystem doc states its status (IMPLEMENTED /
> PARTIAL / STUB / EMPTY / PLANNED) at the top. If you implement something that a
> doc marks STUB/EMPTY/PLANNED, update that banner + this matrix in the same change.
