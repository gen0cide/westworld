# Phases

A staged build plan. Each phase has a clear validation milestone — a thing that works end-to-end that proves the architecture is sound before moving on.

## Phase 0 — Foundation: wire + walk

**Goal**: prove the wire protocol works. One binary, one host, walks to a coord, logs out.

Packages introduced:
- `proto/v235` — login, walk, logout, position update, heartbeat
- `session` — connection lifecycle
- `world` — minimal: own position only
- `action` — `Walk(x, y)` only
- `cmd/cradle` — thin CLI

**Validation**: `westworld cradle -username alex -password REDACTED -walk 120,504` connects to `localhost:43596`, logs in, walks to Varrock center, waits 5s, logs out. ~650 LOC.

**Reference**: OpenRSC's `Payload235Parser.java` and `Payload235Generator.java` are the protocol spec.

## Phase 1 — Full action surface

**Goal**: every atomic action a host might want. Programmatic Go API, no DSL yet.

Packages introduced/expanded:
- `world` — full state mirror: inventory, nearby entities, ground items, regional walkability, fatigue/poison/HP
- `action` — full vocabulary: `Walk`, `Attack`, `Eat`, `TalkTo`, `Drop`, `PickUp`, `OpenBank`, `Deposit`, `Withdraw`, `CastSpell`, `Say`, `WhisperTo`
- `event` — typed events for all inbound packet effects
- `runtime` — the `Host` struct; programmatic Go API for callers

**Validation**: write a hand-coded Go test that logs in, walks to a chicken, kills it, picks up the feathers, walks to a bank, deposits them, logs out. Same hand-coded test against a real OpenRSC server. ~3-4k LOC total.

## Phase 2 — DSL runtime

**Goal**: routines can be authored as DSL scripts and executed by a sandboxed interpreter.

Packages introduced:
- `script` — lexer, parser, AST, interpreter, sandbox (no I/O escape, action timeouts, etc.)

Depends on: AR scripting language description (Alex's homework). DSL design will be AutoRune-inspired; details in [dsl.md](dsl.md).

**Validation**: write a routine in the DSL ("walk to swamp, fish until inventory full, walk to bank, deposit, repeat") and run it via `cmd/cradle -routine my_routine.ws`. Routine completes successfully against the live OpenRSC server.

## Phase 2.5 — Mesa

**Goal**: the shared memory + RAG service exists and is queryable.

Packages introduced:
- `mesa` (root) — shared types
- `mesa/server` — HTTP API, schema migrations, embedding worker, background jobs
- `mesa/client` — SDK + mock impl
- `cmd/mesa` — server binary

**Validation**: stand up mesa locally with Postgres. From a Go test, call `mesa.RegisterBot`, `mesa.WriteEpisode`, `mesa.QueryEpisodes`. Vector retrieval returns relevant entries.

## Phase 3 — Brain + cognition

**Goal**: a host can be driven by an LLM strategist that uses mesa for memory and (optionally) RAG.

Packages introduced/expanded:
- `cognition` — retrieval client wrapping mesa
- `brain` — Anthropic client, tiered model routing, prompt construction, response parsing
- `memory` — local working memory; integration with mesa for episodic/relational/reflective

**Major design session required first**: persona/cohort/reverie design ([personas.md](personas.md), [reveries.md](reveries.md)). Cannot fill in reveries hooks without knowing what reveries are.

**Validation**: a single host runs autonomously for ~1 hour against the live OpenRSC server. The brain makes strategic decisions; routines execute with reveries injected; memory writes flow to mesa; relational records build up.

## Phase 3.5 — Wiki RAG cohort experiment

**Goal**: rsc.wiki ingested into mesa, F2P-filtered. Cohort `wiki_enabled` vs cohort `no_wiki` to test impact.

Components:
- Wiki scraper (one-time tool, possibly its own small binary)
- Voyage 3 embedding of chunked wiki content
- F2P content filtering at ingest

**Validation**: run two cohorts of 5 hosts each for ~1 week. Compare metrics: skill levels, gp earned, deaths, social interactions. Decide if wiki RAG is worth the cost at scale.

## Phase 4 — Reveries (full)

**Goal**: reveries are no longer placeholder hooks — they are a full, persona-driven, emotional-state-aware system.

Packages expanded:
- `reveries` — timing jitter, idle behaviors, spontaneous chat, mistake injection, persona-specific quirks. Hooked at every `reverie.tick()` call site.

**Validation**: side-by-side comparison of routines with reveries-off vs reveries-on. Verify "no drone" behavior — routines visibly weave in idle wander, chat, glances, etc.

## Phase 5 — Delos (orchestrator + UI)

**Goal**: swarm management and the "technician tablets" web UI.

Packages introduced:
- `cmd/delos` — orchestrator + web UI binary
- (delos-internal) supervised host pool, lifecycle, restart-on-crash
- (delos-internal) admin web UI: live host inspection, chain-of-thought tap, cohort analytics

**Validation**: launch 10 hosts via delos. Web UI shows live state of all 10. Admin can click a host and watch its event stream + Brain calls in real time.

## Phase 6 — Scale rollout

| Wave | Size | Goal |
|---|---|---|
| 6a | 10 hosts | Validate end-to-end behavior at small scale; iterate on issues |
| 6b | 100 hosts | Validate mesa performance, cost scaling, log volume |
| 6c | 500 hosts | Target population; research observations begin |

## Phase 7+ — Open research

Once the population is alive and stable, research takes over from engineering:
- Long-horizon goal accomplishment studies
- Community/clustering analysis
- Ethics observation
- Cohort experiments (P2P, persona variation, reverie variation)
- Inter-host learning ("drone teaches drone", observation-and-mimicry, "blogging" pattern)

## Order of dependencies

```
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6
                              ↓         ↑
                          Phase 2.5     │
                          (mesa)        │
                                        │
                       Persona/reverie  │
                       design session  ─┘
                       (required gate)
```

The persona/reverie design session is a **required gate** before Phase 3 begins. Alex and Claude pair-design it. See [personas.md](personas.md) and [reveries.md](reveries.md).
