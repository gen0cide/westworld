# Westworld — Documentation Index

> **STATUS: CURRENT** (verified against code 2026-06-10, HEAD `0bfa818`).
> This is the map: subsystem status matrix, binary inventory, reading order,
> doc map, quickstart. If a banner anywhere in docs/ disagrees with this
> matrix, one of them is stale — fix both in the same change.

**Start here.** Westworld is a Go monorepo (`github.com/gen0cide/westworld`) building an
LLM-driven RuneScape Classic **bot swarm** — a believable, autonomous ~500-host
population — as a research substrate for studying long-horizon agent behaviour,
emergent community, ethics-without-observation, and believability at scale.
See [`../README.md`](../README.md) for the project pitch and the four research
questions; [`research-goals.md`](research-goals.md) for the success metrics.

Hosts run as supervised instances inside **one `cmd/cradle-server` process**
(the fleet daemon: facts + landscape loaded once and shared, one mesa
connection per host, HTTP API + web UI), or one-at-a-time under
**`cmd/host`** (the single-host runner: conductor turn-loop, optional mesa
autonomy, debug dashboard). Each host speaks the RSC wire protocol, mirrors
world state, runs a deterministic **routine DSL** for behaviour, and — when
connected to **mesa** (the off-host gRPC service that owns the LLM seams,
persona registry, and long-term memory) — plays autonomously. An optional
**software renderer** reconstructs what a host sees.

---

## Subsystem status matrix

Be precise about what is real. Status legend: **IMPLEMENTED** (working) ·
**PARTIAL** · **STUB** (interface + canned/no-op impl) · **EMPTY** (package/dir
exists, no Go) · **PLANNED** (design only).

| Subsystem | Package(s) | Status | Doc |
|---|---|---|---|
| Wire protocol (RSC Payload235) | `proto/v235`, `session` | **IMPLEMENTED** | [protocol.md](protocol.md), [world-state.md](world-state.md) |
| World-state mirror | `world`, `event` | **IMPLEMENTED** | [world-state.md](world-state.md) |
| Static game data | `facts`, `assets`, `pathfind`, `worldmap` (+ `cmd/defsgen` codegen) | **IMPLEMENTED** | [architecture.md](architecture.md) |
| Action bridge | `action` | **IMPLEMENTED** | [lang/actions.md](lang/actions.md) |
| Routine DSL | `dsl/*` | **IMPLEMENTED** (open surface gaps tracked as `DSL-*` in [TODO.md](TODO.md)) | [dsl.md](dsl.md), [lang/](lang/README.md) |
| Host runtime / net loop / conductor | `runtime`, `limbic` | **IMPLEMENTED** (conductor + director + limbic goroutines: `runtime/conductor.go`, `runtime/mesa_director.go`) | [architecture.md](architecture.md), [layers.md](layers.md) |
| Render engine + spectator | `render`, `cmd/legacy-cradle/spectate.go` | **IMPLEMENTED** (decoupled) | [render-engine.md](render-engine.md) |
| Scenario system (live-test) | `cmd/scenariogen`, `examples/scenarios` | **IMPLEMENTED** | [scenarios.md](scenarios.md) |
| Cognition (retrieval + knowledge) | `cognition` (`resolve/`, `corpus/`, `knowledge/`, `goalgraph/`) | **IMPLEMENTED** (root `Client` backed by mesa Recall via `mesa/client.AsRetriever`; `StubClient` is the offline default) | [cognition.md](cognition.md), [world-knowledge-and-learning.md](world-knowledge-and-learning.md) |
| Brain (LLM strategist) | `brain` (seam) + `mesa/mesad` (LLM) | **IMPLEMENTED** (real strategist = mesa Decide/Act via `mesa/client.AsStrategist`; `StubStrategist` is the offline default) | [brain.md](brain.md) |
| Memory | `memory`, `hostkv` | **IMPLEMENTED** (tiered Manager + journal + policy table; pebble store; mesa Remote via `mesa/client.AsRemote`) | [memory.md](memory.md) |
| Personas / identity | `persona` + `mesa/personacook` | **IMPLEMENTED** (schema, `Validate()`, `CompilePolicy`, prose render; LLM cook mesa-side) | [personas.md](personas.md), [persona-authoring.md](persona-authoring.md), [persona-compile.md](persona-compile.md) |
| Policy engine (pearl) | `pearl` | **IMPLEMENTED** (compiled from persona; decision gate + quirk/reverie injections) | [persona-authoring.md](persona-authoring.md), [architecture.md](architecture.md) |
| Reveries (believable variance) | — (no package) | **PLANNED** (Phase-5 spec; fragments shipped: `limbic` mood, `persona.ReverieSeed`, pearl `EffectInject`) | [reveries.md](reveries.md) |
| Mesa (off-host service) | `mesa/mesad`, `mesa/client`, `mesa/proto` | **IMPLEMENTED** (gRPC: Decide/Act/Genesis LLM seams, persona registry, LTM + distillation crons, Admin plane) | [mesa.md](mesa.md), [mesa/ARCHITECTURE.md](mesa/ARCHITECTURE.md), [mesa/PROTOCOL.md](mesa/PROTOCOL.md), [mesa/admin-control-plane.md](mesa/admin-control-plane.md) |
| Fleet control plane (cradle daemon) | `cradle`, `cmd/cradle-server`, `cmd/cradle-ctl` | **IMPLEMENTED** (hostcfg loader, supervision, HTTP API + web UI) | [observability.md](observability.md), [architecture.md](architecture.md) |
| Observability | `cradle/web` UI, `debughttp`, decision stream (`runtime/decisions.go`) | **IMPLEMENTED** (in-process; there is no separate observability service — the `delos` idea became the cradle daemon) | [observability.md](observability.md) |

## Binary inventory

| Binary | What it is |
|---|---|
| `cmd/cradle-server` | The fleet daemon: loads `*.hostcfg`, runs + supervises all hosts in one process; HTTP API + web UI (default `localhost:8099`). |
| `cmd/cradle-ctl` | CLI over the cradle-server API (`list`/`status`/`pause`/`restart`/`state`/`eval`/`tail`). |
| `cmd/host` | Single-host runner: conductor turn-loop; `-routine`/`-routines` scripted, `-mesa`+`-goal` autonomous, `-debug-addr` live dashboard, `-repl`. |
| `cmd/legacy-cradle` | Protocol-era single-action harness (one login, one verb); still the home of `-spectate` (live browser viewport), `-render-view` (PNG), `-repl`. |
| `cmd/mesad` | The mesa gRPC service (default `:7077`): LLM seams (`ANTHROPIC_API_KEY`), Postgres LTM, crons, Admin plane. |
| `cmd/mesa-ctl` | Operator CLI over mesad's Admin service (`persona put/import/ls/get/set/rm`, `fleet`, `goal`); authenticates with `$ADMIN_TOKEN`. |
| `cmd/orsc-ctl` | CLI over the OpenRSC server's admin HTTP API (health, world, player inspection). |
| `cmd/scenariogen`, `cmd/parsecheck`, `cmd/defsgen`, `cmd/dronegen`, `cmd/varrock-tiles` | Tooling: scenario generation/runners, offline routine validation gate, static-facts codegen, drone-persona generation, landscape tile inspection. |

There is no `cmd/cradle`, `cmd/mesa`, `cmd/delos`, or `cmd/rendertest` — the
first was renamed to `cmd/legacy-cradle` when `cmd/cradle-server` became the
host process, the mesa binary lives at `cmd/mesad`, and delos never
manifested (its role is covered by the cradle daemon + `debughttp`).

---

## Reading order for a new agent

1. [`../README.md`](../README.md) + [`research-goals.md`](research-goals.md) — what & why.
2. **This file** — the map + status.
3. [`architecture.md`](architecture.md) / [`layers.md`](layers.md) — how the pieces fit; the host control loop.
4. [`state.md`](state.md) — where the project actually is right now (refresh-on-arrival).
5. Then the subsystem doc(s) for whatever you're touching (matrix above).
6. [`TODO.md`](TODO.md) — the **single source of truth for open work** (IDs like `MP-1`, `DSL-13`). [`phases.md`](phases.md) / [`tasks.md`](tasks.md) are the historical record of how we got here.
7. [`questions-and-decisions.md`](questions-and-decisions.md) — why things are the way they are.

## Doc map by topic

- **Behaviour scripting (routine DSL):** [`dsl.md`](dsl.md) + the [`lang/`](lang/README.md)
  reference (`overview`, `syntax`, `actions`, `events`, `state`, `api`, `knowledge`,
  `repl`, `writing-routines`). Authoring routines: start at [`lang/writing-routines.md`](lang/writing-routines.md).
- **Live testing:** [`scenarios.md`](scenarios.md) — the scenario system (author →
  generate → run → triage), multi-host orchestration, the YAML schema.
- **Rendering:** [`render-engine.md`](render-engine.md) — the software rasteriser +
  live spectator + authenticity invariants. (Superseded plan archived: `archive/initial-brainstorming/render-port-plan.md`.)
- **Server/protocol (for agents working on the wire or world state):**
  [`protocol.md`](protocol.md), [`world-state.md`](world-state.md), [`server-config.md`](server-config.md),
  [`tutorial-host-snapshot.md`](tutorial-host-snapshot.md); server ops runbook: [`agents/openrsc-steward.md`](agents/openrsc-steward.md).
- **Cognition & autonomy:** [`cognition.md`](cognition.md), [`world-knowledge-and-learning.md`](world-knowledge-and-learning.md),
  [`cognition-and-autonomy.md`](cognition-and-autonomy.md), [`brain.md`](brain.md), [`memory.md`](memory.md),
  [`lang/thought-architecture.md`](lang/thought-architecture.md).
- **Personas:** [`personas.md`](personas.md) (concepts), [`persona-authoring.md`](persona-authoring.md)
  (field-by-field authoring guide), [`persona-compile.md`](persona-compile.md) (the cook),
  [`persona-schema.reference.yaml`](persona-schema.reference.yaml) + [`host-persona.template.yaml`](host-persona.template.yaml),
  example: [`personas/dolores.json`](personas/dolores.json).
- **Mesa (the off-host service):** [`mesa.md`](mesa.md), then [`mesa/ARCHITECTURE.md`](mesa/ARCHITECTURE.md)
  (topology), [`mesa/PROTOCOL.md`](mesa/PROTOCOL.md) (the gRPC surface),
  [`mesa/admin-control-plane.md`](mesa/admin-control-plane.md) (operator plane),
  [`mesa/knowledge-pipeline.md`](mesa/knowledge-pipeline.md) (persona compile + memory RAG).
- **Reveries (design):** [`reveries.md`](reveries.md) — the Phase-5 believable-variance spec.
- **Observability & ops:** [`observability.md`](observability.md) (cradle UI, debug
  dashboard, decision stream), [`agents/`](agents/README.md) (operating charters).
- **Planning & history:** [`TODO.md`](TODO.md) — the one live backlog (SSOT);
  [`phases.md`](phases.md) + [`tasks.md`](tasks.md) — the closed historical ledger;
  [`state.md`](state.md) — current snapshot; [`questions-and-decisions.md`](questions-and-decisions.md) — decision record.
- **Archive:** [`archive/initial-brainstorming/`](archive/initial-brainstorming/) — superseded
  or differently-manifested design docs, each with an ARCHIVED header saying what actually
  shipped. War-story chapters land in `lessons-learned/` (Stage 4 of the 2026-06 docs tidy).

## Quickstart

```bash
cd ~/Code/westworld
go build ./...                     # build everything
go test ./...                      # run tests

# Secrets: WESTWORLD_PASSWORD supplies the OpenRSC test password — see
# server-config.md; never print/commit it:
set -a; . ./.local.env; set +a

# One host, scripted (conductor runs the routine):
go run ./cmd/host -username stubbs -server localhost:43594 -facts ~/Code/openrsc \
  -routine examples/routines/<name>.routine
# add -debug-addr localhost:8090 for the live dashboard (/ws thoughts, /state, /eval)

# One host, autonomous (mesa plans, host executes):
go run ./cmd/mesad            # mesa service on :7077 (Postgres + ANTHROPIC_API_KEY)
go run ./cmd/host -username Delores -mesa localhost:7077 -goal "earn money in Varrock"

# A fleet (the cradle daemon):
go run ./cmd/cradle-server -config cradle/hostcfg/examples/swarm.hostcfg
# web UI + HTTP API on localhost:8099; drive it with cmd/cradle-ctl (list/eval/tail/...)

# Live render viewport (legacy harness): go run ./cmd/legacy-cradle -spectate \
#   -spectate-addr localhost:8089, then open the URL.
# Live-test scenarios: see scenarios.md (cmd/scenariogen + its run_*.sh runners).
```

> **Accuracy convention:** every subsystem doc states its status (IMPLEMENTED /
> PARTIAL / STUB / EMPTY / PLANNED) at the top. If you implement something that a
> doc marks STUB/EMPTY/PLANNED, update that banner + this matrix in the same change.
