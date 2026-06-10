# Observability

> **STATUS: BUILT — in a different shape than originally designed.** Verified 2026-06-10
> against branch HEAD `0bfa818`. The original "technician tablets" design (a `cmd/delos`
> web app, an `obs` package, a mesa `brain_calls` ledger) never manifested under those
> names; the intent shipped as the **cradle daemon** (web UI + JSON API), the per-host
> **`debughttp/` control plane**, the **decision stream** (`agent_thought` →
> `decisions.jsonl` + mesa), and **host→mesa telemetry** (the `metrics` table). The
> original design prose is archived verbatim at
> [`archive/initial-brainstorming/observability-original-design.md`](archive/initial-brainstorming/observability-original-design.md).

The Westworld metaphor still holds — technicians watching hosts think, pausing them,
issuing directives — but the tablet is the cradle daemon, not a separate orchestrator.
One process supervises the fleet and serves the API, the UI, and every host's debug
surface on a single port.

## 1. The cradle daemon — fleet UI + API (`cradle/`, `cmd/cradle-server`)

`cmd/cradle-server -config <dir> -listen localhost:8099` runs the fleet and the control
plane. `cradle/api.go` is the HTTP/JSON surface:

- `GET /api/hosts` — the roster (`cradle.HostStatus`: status, goal, mesa health,
  restarts, live position/HP, current routine + executing line, analysis flag).
- `POST /api/hosts/{name}/pause|resume|stop|restart` — lifecycle, via the registry.
- `POST /api/hosts/{name}/analysis/enter|exit|directive`, `GET .../analysis` —
  operator-override ANALYSIS mode (conductor frozen, world replies suppressed, memory
  writes suspended; directives are classified command/answer/hypothetical and return a
  structured `runtime.AnalysisResult` verdict).
- `/api/hosts/{name}/debug/…` — each LIVE host's full `debughttp` surface, proxied under
  a path prefix (`StripPrefix`-mounted on host-live, torn down on exit). One shared
  server, no per-host ports.
- `/debug/pprof/` — process-wide pprof; the goroutine profile names any leaked bus
  subscriber by file:line (the soak's leak instrumentation).

`cradle/web/index.html` (embedded, served at `/`) is the single-page web UI over the
same JSON the CLI uses, so UI and CLI never drift. Per-host panels: **Routine**
(line-numbered DSL source with a live current-line highlight, fed by
`HostStatus.line_trace`), **Narration** (the in-character `note()` stream, kind
`routine_note`), **Thoughts** (per-turn decision cards: trigger, move kind, first-person
reasoning, the authored DSL), **Chat & Server**, **Skills & XP**, **Inventory &
Surroundings**, **Mind** (knowledge · relationships · goal graph), and the **Operator
Console** (ANALYSIS directives, auto-entering analysis mode). Feeds are seeded from the
recorded event ring, then go live over the per-host WebSocket.

`cmd/cradle-ctl` drives the same API headless:
`list status pause resume stop restart bounce logoff state events eval script analysis
tell tail`.

## 2. The per-host debug control plane (`debughttp/`)

A live, scriptable HTTP surface over one logged-in host — the non-blocking sibling of
the stdin REPL, sharing one persistent interpreter session (`debughttp/server.go`):

| endpoint | what |
|---|---|
| `GET /` | browser dashboard: live state + thought stream |
| `GET /ws` | WebSocket of **every** bus event, live |
| `GET /state` | world-mirror snapshot: position, vitals, fatigue, skills/XP, inventory, NPCs/players, dialog, recent server messages + the `bus_subs_star`/`bus_subs_total` leak gauges |
| `GET /mind` | the mind inspector: knowledge facts (claim/confidence/provenance), relationships (trust/affinity/grievance), goal-graph nodes+edges, open questions |
| `GET /events?since=N&kind=K&limit=L` | the recorded event ring |
| `POST /eval` | one DSL line against the persistent session |
| `POST /script` | a whole `.routine` |

Every bus event is recorded to an in-memory ring (100k standalone; 4k per host in the
fleet — `cradle/api.go` `perHostRing`) **and** appended to a JSONL file (default
`/tmp/cradle_debug/<username>_events.jsonl`). In the fleet the JSONL is allowlisted to
analysis-relevant kinds (`agent_thought`, `system_message`, `policy_veto`,
`chat_message`) because the full firehose is unbounded disk at scale; size-capped
rotation is still open ([TODO.md](TODO.md) O-3).

Standalone mounts of the same surface: `cmd/host -debug-addr localhost:8090` and
`cmd/legacy-cradle -debug-http` (default `localhost:8090`, `-debug-log` for the JSONL
path).

## 3. The decision stream

`event.AgentThought` (kind `agent_thought`, `event/agent.go`) is the chain-of-thought
record: turn, trigger, goal, position/HP, perception, first-person reasoning, move kind,
and any DSL she authored. It is published each turn by the autonomous director
(`runtime/mesa_director.go`) and by the forage/speech/hybrid deciders. Persistence is
twofold (`runtime/decisions.go`): every thought is appended to
`<dataDir>/decisions.jsonl` (default data dir `~/.westworld/hosts/<user>`; one JSON
object per line, flushed per write — the overnight-soak trace) **and** mirrored to mesa
as a salience-0.4 observation (kind `"decision"`) on the existing observation stream, so
fleet-level decision queries hit one place.

## 4. Host→mesa telemetry (`runtime/telemetry.go`)

Every 60s each host ships a metrics batch over `Journal.ReportMetrics` into mesa's
append-only Postgres `metrics` table (`mesa/mesad/ltm.go`, indexed
`(host_id, name, at DESC)`): `host.uptime_seconds`, `pearl.vetoes`, `agent.turns`,
`journal.episodes`, `ledger.relationships`, `bus.subscribers_star` / `_total` (THE
leak-acceptance gauge — must plateau, not grow per turn), `go.goroutines`,
`go.heap_inuse_mb`, and `memory.hits` / `misses` / `journal_depth`. This is the
observability input for the 200-drone soak verdicts (O-5) and mesa cron aggregation.

Logs throughout are structured `slog` (`-log-level debug|info|warn|error` on
`cmd/cradle-server` and `cmd/host`).

## 5. What never manifested — and where the intent went

| original design | reality |
|---|---|
| `cmd/delos` web app | the cradle daemon UI (`cradle/web/index.html`) |
| `obs` package event log | the `debughttp` recorder (ring + JSONL) |
| `brain_calls` ledger (model/tokens/cost/latency per call) | **NOT BUILT** — nothing tracks tokens or dollars today; `mesa/llm/anthropic.go` has no usage accounting ([TODO.md](TODO.md) O-9) |
| cohort analytics + cost dashboard | **NOT BUILT** — re-scoped per-archetype over the mesa `metrics` table ([TODO.md](TODO.md) O-7) |
| admin actions: pause/resume/restart/snapshot | shipped (registry + `/state`) |
| "inject event" | shipped as `/eval` + ANALYSIS directives |
| "edit persona" | shipped as `mesa-ctl persona put/set` (admin plane, `mesa/mesad/admin.go`) |
| "force reflection", "adjust cohort" | stale — there is no on-demand reflection-generation trigger (recall surfaces stored reflections; genesis compiles at login), and cohort was reframed as launch batch (archetype carries personality) |

Open work: see [/docs/TODO.md](TODO.md) §6 — notably O-3 (JSONL rotation), O-5 (the
soak run), O-7 (analytics re-scope), O-9 (LLM cost/token ledger), O-13 (pearl hit-rate
telemetry).

## Privacy considerations (relevant for any future open-sourcing)

If we ever publish the swarm transcripts as a research dataset, the captured brain prompts and responses contain agent "thoughts" that could be considered private. Even though no real humans are involved, the prompts may contain content about other agents that — once published — exists in a "social" form.

For now: it's all our data. We make decisions about publication later. The system records everything.
