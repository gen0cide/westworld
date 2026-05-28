# Task list

Canonical enumeration of every work item, grouped by phase.
**Status here is the *plan*, not real-time state** — for in-flight
status during an active session, check the harness's task tracker
(numbered IDs match this doc). Update this doc when the structure
changes (tasks added/removed/reordered), not for every
in-progress/completed status flip.

Cross-references:
- [`phases.md`](phases.md) — phase narratives + ordering
- [`lang/state.md`](lang/state.md) — query layer build checklist
  (sub-tasks #56–#65)
- [`lang/development-workflow.md`](lang/development-workflow.md) —
  how work actually happens (REPL-driven, delores in the loop)

## Completed

### Phase 0 — Foundation: wire + walk
- `#19` Port Sector/Tile binary format from Authentic_Landscape.orsc
- `#20` Build 96x96 collision grid centered on player
- `#21` Define CollisionFlag constants matching the client
- `#22` Port World.findPath BFS to Go
- `#23` Add multi-corner Walk packet builder + path encoder

### Phase 1 — Full action surface
- `#24` Update Host action methods to pathfind before action
- `#25` E2E: Goblin attack ✓ live (TalkTo Hans partially walked — Hans wanders)

### Phase 2 — DSL runtime foundation
- `#30` DSL step 1: token + lexer + AST + parser skeleton
- `#31` DSL step 2: statement parser
- `#32` DSL step 3: expression parser with precedence climbing
- `#33` DSL step 4: static validator
- `#34` DSL step 5: AST interpreter + locals + control flow
- `#35` DSL step 6: action channel + Host bridge
- `#36` DSL step 7: resource caps
- `#37` DSL step 8: event handler dispatch + two-tier scope
- `#38` DSL step 9: conformance suite + delos observability hooks

## Phase 2.5 — Language v2 (next)

### Stage 1 — Minimum dev surface (sequential)

These three are the bottleneck. Land in this order, then Stage 2
iterates live.

- `#51` **Result/Error model + bang variants** —
  typed `ErrorCode` enum (Go) + `Result {val, err}` (DSL) +
  auto-generated `eat!` / `walk_to!` / etc. Per
  [`lang/actions.md`](lang/actions.md).
- `#53` **Filename = routine name + `ParseRoutineString`** —
  validator enforces filename↔routine-name match for files;
  string loader for in-memory routines (REPL, exec, improvise).
  Per [`lang/syntax.md`](lang/syntax.md).
- `#54` **REPL** — `cradle -repl` + `cradle -repl-on-fail` with
  `.resume`. Per [`lang/repl.md`](lang/repl.md) and
  [`lang/development-workflow.md`](lang/development-workflow.md).

### Stage 2 — Iterate live (delores in the loop)

Parallelizable; pick by what the next routine wants.

- `#46` **Query layer umbrella** (~100 accessors) — broken into
  per-domain sub-tasks below. Per
  [`lang/state.md`](lang/state.md) "Build plan" section.
  - `#56` Vitals (~12) — hp/max_hp/hp_fraction/prayer/max_prayer/fatigue/combat_level/quest_points/position/is_busy/is_in_combat/is_sleeping
  - `#57` Skills (~90) — all 18 × level/max_level/xp/xp_to_next_level/percent_to_next_level + `.list()`
  - `#58` Equipment (~14) — 10 slots + style + total_bonuses
  - `#59` Prayer (~58) — 14 prayers × name/level_req/drain_rate/is_active + active/available lists
  - `#60` Magic (~200) — ~50 spells × level_req/runes_required/can_cast + known/selected/castable
  - `#61` Inventory enhancements (~7) — weight/find/slot_of/is_full/is_stackable/def
  - `#62` Entity-view enrichment (~10) — combat_level/max_hp/hp_fraction/is_attackable/in_combat_with/is_friend/is_mine
  - `#63` `world.locs.*` additions (~6) — scenery/shops/spawn_points + `.within(radius)`
  - `#64` Recent-events buffer (~4) — last_chat/last_pm/last_damage/last_server_message
  - `#65` Combat / bank / trade views (~15) — `combat.target/is_engaged/last_damage_*`, `bank.is_open/slots/has/count`, `trade.is_active/opponent/mine/theirs/both_accepted`
- `#47` `when` watchers — block-scoped state-transition handlers (rising-edge, truth-at-registration counts)
- `#48` `select` — block-until-one-fires construct. Three case
  types (`when` / `on` / `timeout`). Break/continue inside
  cases propagate to the enclosing loop (matches Go's
  `for { select {...} }`). First-declared wins on
  simultaneous-ready. Time units in `timeout`: `Ns`/`Nms`/`Nm`.
- `#49` `defer` — cleanup on scope exit (LIFO, runs on
  return/abort/error/cancel)
- `#50` `try`/`recover` — bang-error boundary. Defers in the
  `try` block run before `recover` executes.
- `#66` Lambdas — `IDENT => expr` (and `(IDENT, IDENT) => expr`)
  for filter/map/find predicates. Single-expression body only;
  for multi-statement, use named procs.
- `#67` Validator cohesion pass — all the Tier-1 rules from the
  design review: handlers can't yield (no wait/wait_until/select),
  `super()` only in extends handlers, select-without-timeout
  warning, bang on non-bang-eligible callables rejected.

### Stage 3 — Deferred

- `#52` `super()` / `extends host` — waits for Phase 4 persona tier

### Cleanup (lands when Stage 2 routine equivalents exist)

- `#55` Delete `runtime/auto_eat.go` + `runtime/combat_loop.go` — once `when` + query layer + Result/Error let the routine versions take over

## Phase 2.6 — Knowledge ingestion (not yet ticketed)

Will file tasks when starting. Anticipated structure:

- Mesa knowledge_chunks schema + embedding worker (Voyage 3)
- rsc.wiki scraper + ingest (one-time tool)
- AutoRune script corpus ingest (Alex's historical scripts)
- `recall()` stdlib wiring (replaces Phase-2 stub)
- `cradle -knowledge-query` admin CLI

## Phase 2.7 — Admin tooling (not yet ticketed)

Will file tasks when starting. Anticipated structure:

- OpenRSC server-side: grant delores admin
- Admin DSL actions: `admin_set_stat`, `admin_give_item`,
  `admin_teleport`, `admin_heal`, `admin_set_fatigue`,
  `admin_set_position`, `admin_clear_inventory`, `admin_kick`
- Validator-time `IsAdmin` flag check
- `admin_sandbox.routine` — common scenario bootstrap

## Phase 2.8 — Live edge-case testing (outputs)

These pre-existing tasks all get **discovered and closed during
live REPL sessions** in Phase 2.8, not as separate upfront work.
They're outputs of playing the game, not specs:

- `#26` Dynamic boundary state (cut webs / opened doors at runtime)
- `#27` NPC dialog choice — live test with a real dialog tree
- `#28` Death/respawn detection + recovery
- `#29` Banking orchestrator (deposit-when-full) — author as routine
- `#39` Trades full state machine
- `#40` Duels
- `#41` 18 RSC skills + their opcodes
- `#42` E2E test: completed trade between alex + delores (full handshake)
- `#43` E2E test: bank deposit cycle live (visit banker, deposit, close)
- `#44` E2E test: AttackPlayer (consensual PVP / dueling zone)
- `#45` E2E test: DropItem from inventory + alex picks it up

## Phase 3+ (mesa, brain, reveries, delos, scale, research)

Tasks for these phases get filed at phase start. See
[`phases.md`](phases.md) for the narrative.

## How to use this doc

- **Starting a session:** read [`state.md`](state.md) first for
  current status, then this doc for the queue.
- **Pick the next thing:** lowest pending ID in the current
  stage's section is usually the right call, but anything
  marked "parallelizable" in phases.md can interleave.
- **Closing a task:** update the in-session task tracker for
  status; only edit this doc when the *structure* changes.
- **Adding a task:** add to the in-session tracker, then add a
  one-line entry here. Cross-reference any doc that describes
  the work.

## Adding language surface (vs adding implementation)

After the `dsl/spec/` refactor (2026-05-28), the language surface
(builtin names, events, query paths) is centralized:

- **Add a new builtin**: row in `dsl/spec/actions.go` + Go
  wrapper in `runtime/dsl_actions.go::actionHandlers`. Validator
  + bridge pick it up automatically. Consistency tests catch
  drift.
- **Add a new event**: row in `dsl/spec/events.go` + extension
  to `runtime/dsl_events.go::translateEvent` if it maps to a
  typed Go event.
- **Add a new query accessor**: row in `dsl/spec/accessors.go`
  + extension to the relevant view's `Get()` switch in
  `runtime/dsl_views.go`.

If you find yourself touching `dsl/validator/validator.go`'s
`builtins` map or `runtime/dsl_bridge.go`'s register block by
hand, you're going against the grain — those maps now derive
from spec at init time.

## When this doc is out of date

If the in-session tracker has tasks numbered higher than what's
listed here, this doc is stale. The next person to do
structural work updates the doc; status-only changes don't
require it.
