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

## Phase 2.5 — Language v2 ✓ (mostly done, two items deferred)

### Stage 1 — Minimum dev surface (sequential) ✓

- `#51` ✓ **Result/Error model + bang variants** — typed
  `ErrorCode` enum (Go) + `Result {val, err}` (DSL) +
  auto-generated `eat!` / `walk_to!` / etc.
- `#53` ✓ **Filename = routine name + `ParseRoutineString`** —
  validator enforces filename↔routine-name match for files;
  string loader for in-memory routines (REPL, exec, improvise).
- `#54` ✓ **REPL** — `cradle -repl` + `cradle -repl-on-fail`
  with `.resume`.

### Stage 2 — Iterate live (delores in the loop) ✓

- `#46` ✓ **Query layer umbrella** (~100 accessors) — all
  per-domain sub-tasks shipped:
  - `#56` Vitals (~12) ✓
  - `#57` Skills (~90) ✓
  - `#58` Equipment (~14) ✓
  - `#59` Prayer (~58) ✓
  - `#60` Magic (~200, 48 spells embedded from SpellDef.xml) ✓
  - `#61` Inventory enhancements (~7) ✓
  - `#62` Entity-view enrichment (~10) ✓
  - `#63` `world.locs.*` additions (~6) ✓
  - `#64` Recent-events buffer (~4) ✓
  - `#65` Combat / bank / trade views (~15) ✓
- `#47` ✓ `when` watchers — block-scoped state-transition handlers
- `#48` ✓ `select` — block-until-one-fires construct (when / on / timeout cases)
- `#49` ✓ `defer` — cleanup on scope exit (LIFO)
- `#50` ✓ `try`/`recover` — bang-error boundary
- `#52` ✓ **`extends "parent.routine"` (file-level v1)** —
  inherits parent procs + on-handlers; cycle-safe;
  string-loaded routines reject extends explicitly. See
  [`lang/syntax.md`](lang/syntax.md) "File structure → extends".
  *Per-handler `extends host` + `super()` chaining split to #93.*
- `#66` ✓ Lambdas — `IDENT => expr` for filter/map/find predicates
- `#67` ✓ Validator cohesion pass

### Stage 2.5 — Live-test discoveries (Phase 2.8 work pulled in early)

These landed during interactive REPL + live OpenRSC sessions
between bernard + delores; documented here so the survey is
complete.

- `#68`–`#71` ✓ REPL polish + locs.searchByName preference +
  walk_to / say docstring fixes
- `#72`–`#74` ✓ Cognition + brain stubs (deterministic) wired
  through `ask_brain` / `contemplate_reality` builtins —
  the LLM bridge in delos picks up where the stubs leave off
- `#75`–`#78` ✓ `use(item, target)` polymorphism —
  scenery / ground_item / npc / player + `interact_at(x, y, opt?)`
- `#79` ✓ Spatial utilities: `distance_to` + `in_region`
- `#80` ✓ `.contains()` on `world.last_*` message records
- `#81` ✓ Dialog option menu — surfaced as `world.dialog.options`
- `#82` ✓ Tier 1 decoder anomaly assertions sweep
- `#83` ✓ `walk_path([(x,y), ...])` — pre-planned multi-corner walks
- `#84` ✓ `is_reachable(x, y)` — pre-flight pathfinder check
- `#86` ✓ `wait_for_dialog(timeout=Ns)` — block on dialog menu
- `#87` ✓ `event.item_gained(id, count)` — inventory growth event
- `#88` ✓ `world.ground_items.by_id(id, radius?)` — nearest by type
- `#89` ✗ `world.npcs.by_type(id)` + `.random()` — NEVER built under these names (the false ✓ propagated from primitives-backlog; typed selection shipped as `world.npcs.by_id`/finders — see lang/api.md)
- `#90` ✗ `last_attacked_npc` / `last_attacked_player` — NEVER built under these names (combat target state shipped as `combat.target`; see lang/api.md)
- `#91` ✓ Trade state machine (bernard ↔ delores) live-tested
- `#92` ✓ Duel state machine (bernard ↔ delores) live-tested,
  including death + respawn detection
- Bounds `{ }` directive ✓ — region-scoped event handlers
  (box / circle / near shapes, nested intersection semantics)

### Stage 2 mop-up (closed late, after Phase 2.5 declared done)

- `#94` ✓ Inventory amount decoder bug — non-stackable slots
  read garbage. Threaded `isStackable` through
  `decodeInventorySlotUpdate` mirroring full-snapshot decoder.
- `#85` ✓ `repeat { body } until <cond> timeout <expr>` —
  shipped as a statement form (do-while bounded by wall-clock).
  Mandatory-timeout + no-handler-body validator guards. See
  [`lang/syntax.md`](lang/syntax.md) "repeat ... until".

### Stage 3 — Deferred (still pending)

- `#95` Live-test combat style toggle — needs alex + live
  OpenRSC up; observe XP split via skill_xp_gained events
  post-toggle.
- `#93` **v2 of #52**: per-handler `on ev() extends host { ... super(...) ... }`
  + `host.defaults.<event>` resolution. Waits for Phase 4
  persona tier.

### Cleanup ✓

- `#55` ✓ `runtime/auto_eat.go` + `runtime/combat_loop.go` deleted
  in favor of routine equivalents.

## Phase 2.9 — BODY API freeze + build-out ✓ (#114–123)

Shipped 2026-05-29 as one concentrated burst. **This phase was not in the
original 2.6→2.8 ladder — it ran *ahead* of it.** After Phase 2.5 the team
chose to freeze the v1 host-facing API and bring the implementation up to that
frozen surface before building any cognition on top. The commit hashes are the
authoritative record (the in-session task IDs `#114–123` were never enumerated
in this doc until now).

- `#114` ✓ **API freeze (v1)** — `docs/lang/api.md` §8 + `build-backlog.md` +
  `protocol.md` ratified and pushed (`38ef5a0`). Amend only by deliberate
  decision, never by drift.
- `#115` ✓ **DSL hub split + namespaced Def/Instance rename** — split the
  `dsl_actions.go` / `dsl_views.go` monoliths into per-namespace
  `actions_<ns>.go` / `views_<ns>.go` behind one central registration table
  (`d0fc067`); then the flat→namespaced surface rename + `ItemDef`-vs-`InvSlot`
  Def/Instance split across spec/impl/scenarios (`d41b7e9`, with
  `namespace_rename_test.go`).
- `#116` ✓ **opcode-234 combat/health decode** onto the world mirror (`faeae0b`).
- `#117` ✓ **Perception accessors** — combat (`909f2ac`), world + inventory
  (`2364403`), bank reads (`97c7924`), magic (`a6d3e45`), `prayer.active`
  (`353d1de`), `self.equipped` / `self.position.plane` (`6883324`).
- `#118` ✓ **Verbs** — `bank.deposit_all`/`close` and the other bulk verbs
  (landed alongside the #117 commits).
- `#119` ✓ **Events** — `world.messages` ring + `on message` / `on xp_gain` /
  `on npc_killed` / `on target_died` (`c342f01`).
- `#120` ✓ **`resolve()` / `resolve_one()` control-plane** + learned-alias
  store (new `cognition/resolve/` package, REAL) (`d310c3b`, `3347572`).
- `#121` ✓ **Content-table fixes** (data, verified vs OpenRSC defs +
  server source — see `build-backlog.md` §"#121 content table").
- `#122` ✓ **Live-test catalog merged** on the frozen namespaced surface
  (`8120f91`, 195 entries). Later refactored (2026-05-31) into a **code-free
  manifest** referencing the `.routine` corpus (now 196 files); `cmd/scenariogen`
  became a validator, old catalog → `scenarios.yaml.bak` — see
  [`scenarios.md`](scenarios.md) §2.1. Ongoing hardening = Phase 2.8 below.
- `#123` ✓ **`protocol.md` re-run** against the namespaced surface.
- ✓ **Two new faculties built end-to-end this cycle:** `shop` (decode + world
  mirror + actions/views, `3678044`) and **fatigue → sleep** (`ba8d043`).

> **ID-numbering note.** In-session tracker IDs `#96–#113` were used during the
> freeze era and are not all back-filled here — notably `#98–100`
> (cognition/brain stubs), `#105` (combat-style builtin, see `#95`), and `#107`
> (`command()` admin builtin, `b876c2d`). When the in-session tracker and this
> doc disagree on a number, the commit hash wins.

## Render engine + live spectator ✓ (out-of-band; actively iterated)

**Not on the phase ladder.** A decoupled, read-only observability capability
that landed in parallel with the body work and is *still being iterated* — by
far the largest single workstream by commit volume. A pure-Go (WASM-capable)
software renderer reconstructs a host's *perceived* RSC scene and serves it live
in a browser via `cradle -spectate` (click-to-walk via the BFS pathfinder,
rotate/zoom, `p`/`c` agent-readable captures).

The engine is now **`render/orsc`** — a from-scratch, line-faithful Go port of
OpenRSC's `orsc/graphics/three` pipeline (World/Scene/RSModel/Shader), reading
OpenRSC's own `.orsc` assets. It renders terrain, textures, scenery, multi-story
buildings + roofs, doorframes, and animated + gliding actor billboards. The
original RSCPlus-matched classic software rasteriser (`Scene`/`GameModel`/`Surface`
in `render/`) was **removed** once the orsc port superseded it (see git history);
`render/` is now just the shared sprite-compositing + `View`/`Entity` types both
the renderer and the cradle use. Tracked in [`render-engine.md`](render-engine.md)
— **not enumerated as numbered tasks here.**

## Phase 2.6 — Knowledge ingestion — ⏳ NOT STARTED (re-sequenced)

> **Status:** deferred and never begun. After Phase 2.5, effort went to the
> BODY build-out (#114–123), the render engine, and the scenario run-to-ground
> campaign instead — **zero** commits touch mesa / embeddings / `recall()`
> wiring as of HEAD (`fd0731c`). `recall()` remains the Phase-2 stub. This work
> now folds into the higher-layer scaffolding handed to the dev-partner agent
> (see Phase 3+ and [`archive/initial-brainstorming/westworld-dev-partner.md`](archive/initial-brainstorming/westworld-dev-partner.md)).

Will file tasks when starting. Anticipated structure:

- Mesa knowledge_chunks schema + embedding worker (Voyage 3)
- rsc.wiki scraper + ingest (one-time tool)
- AutoRune script corpus ingest (Alex's historical scripts)
- `recall()` stdlib wiring (replaces Phase-2 stub)
- `cradle -knowledge-query` admin CLI

## Phase 2.7 — Admin tooling — ✓ substantially delivered (via the scenario campaign)

The original plan (dedicated `admin_*` DSL verbs + an `IsAdmin` gate) was
**superseded in practice**: admin capability arrived through the generic
`command()` builtin (`b876c2d`) driving server `::` commands, and the test
hosts (Drone1–20 + Bernard + Delores + Arnold) are all admin per the DB clone.
The 196-scenario corpus already uses admin `::commands` (`item`, `damage`,
`setstat`/`setcurrentstats`, `teleport`, `wipeinv`) for its preconditions —
those engine fixes landed during the live-test cycle. No separate
`admin_sandbox.routine` / `admin_*` verb surface was needed.

Remaining (only if a future need surfaces): a validator-time `IsAdmin` gate so
production hosts reject `command()` at parse time.

## Phase 2.8 — Live edge-case testing — 🔄 IN PROGRESS (the "run-to-ground" campaign)

> **Status:** active. The individual state-machine "outputs" below (trade,
> duel, death/respawn, banking, dialog, dynamic boundaries) all **shipped** in
> Phase 2.5 Stage 2.5. What remains in-flight is **breadth/correctness
> validation** via the scenario "run-to-ground" campaign: the 196-scenario
> corpus (`examples/scenarios/*/*.routine` — what the runners actually execute)
> is swept against a live OpenRSC server and every failure is run to ground — *scenarios are gap-finders; fixes
> land in the engine, never worked around in content.* Organized as
> worktree-isolated fan-out rounds: rounds 1+2 (herblaw/thieving/magic/bank/
> inverr/misc + fire/gather/stalls) merged at `18ac18b`, then r3 (combat
> `fd0731c`, gathering `5117845`, magic `272ca58`, ranged `40bea3a`). Catalog
> pass-rate trajectory: 16 baseline → ~70/88 (`aab4300`) → still hardening. Not
> declared finished. Full mechanics in [`scenarios.md`](scenarios.md).

These pre-existing tasks were **discovered and closed during live REPL/scenario
sessions** rather than as upfront work — they're outputs of playing the game,
and the state machines among them are now ✓ live (see Phase 2.5 Stage 2.5):

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

## Phase 3+ — the higher layers — ⭐ NEW ACTIVE FRONTIER

With the BODY API frozen and built out, observability shipped, and the scenario
campaign in mature maintenance, the forward direction is to **build UP**
(cognition → brain → persona → memory → reveries → mesa) and **OUT** (steward
the OpenRSC server the swarm runs against). This is the work being **handed to
two external AI agents** — see [`agents/`](agents/README.md):

1. **OpenRSC server stewardship** — own the live game server + DB + the
   authentic-client reference sources. Charter:
   [`agents/openrsc-steward.md`](agents/openrsc-steward.md).
2. **Higher-layer scaffolding** (the dev partner) — turn the cognition/brain
   stubs real and stand up the `memory`/`persona`/`reveries`/`mesa` interface
   scaffolds so personas/reveries can be authored *inside* them. Charter:
   [`archive/initial-brainstorming/westworld-dev-partner.md`](archive/initial-brainstorming/westworld-dev-partner.md). This
   subsumes the deferred Phase 2.6 (knowledge ingestion / mesa) work.

Tasks for these phases get filed at phase start; see [`phases.md`](phases.md)
for the narrative and dependency ordering. Carried-over items that unblock
here: `#93` (per-handler `extends host` + `super()` — waits for the Phase 4
persona tier) and `#95` (combat-style live XP-split verify — folds into the
#117 combat work).

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

After the `dsl/spec/` refactor (2026-05-28) **and the per-namespace hub split
(`d0fc067`, 2026-05-29)**, the language surface (builtin names, events, query
paths) is centralized in `dsl/spec/`, and the *implementation* now lives in
per-namespace files behind one central registration table — **not** the old
`dsl_actions.go` / `dsl_views.go` monoliths (those were split):

- **Add a new builtin**: row in `dsl/spec/actions.go` + Go wrapper in the
  matching `runtime/actions_<ns>.go` (registered via the central registry).
  Validator + bridge pick it up automatically. Consistency tests catch drift.
- **Add a new event**: row in `dsl/spec/events.go` + extension to
  `runtime/dsl_events.go::translateEvent` if it maps to a typed Go event.
- **Add a new query accessor**: row in `dsl/spec/accessors.go` + extension to
  the relevant namespace's `runtime/views_<ns>.go` `Get()` switch.

If you find yourself touching `dsl/validator/validator.go`'s `builtins` map or
the per-namespace register blocks by hand, you're going against the grain —
those derive from spec via the central registry at init time.

## When this doc is out of date

If the in-session tracker has tasks numbered higher than what's
listed here, this doc is stale. The next person to do
structural work updates the doc; status-only changes don't
require it.
