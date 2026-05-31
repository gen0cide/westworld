# Current state (read this first on context refresh)

Last refreshed: 2026-05-31. **Two threads since the 2026-05-29 freeze: (1) the BODY API freeze + build-out (below), and (2) a now-SHIPPED render port — a decoupled Go RSC viewport renderer + live browser spectator (merged 2026-05-30, `wt/render` → `main` at `55578dc`).** The render thread was independent of the API freeze and is now landed. The earlier 2026-05-29 framing: **pivoted from scenario-grinding to freezing the BODY API before cognition/brain/persona** — the live-test campaign drove the engine hard, so the host-facing DSL surface was locked to give the upper layers a stable contract.

## SHIPPED 2026-05-30 — render port + live spectator (`render/` + `cradle -spectate`)

A standalone Go reimplementation of the RSC client viewport, so an agent (or a human) can *watch* a host play without the Java/web client. **No CGo, no native window — the browser is the display.** Verified against OpenRSC cache assets.

- **`render/` package** — decodes OpenRSC `.orsc` archives (models, landscape, textures) and rasterizes the host's perceived world to a PNG: terrain + overlays, boundaries/walls + doors/doorframes, roofs (under-roof cull, roof-level pass), scenery, NPCs/players with authentic clothing colours, walk-cycle leg animation + smooth sub-tile motion interpolation (gliding walk), item sprites, face sorting. Pure render-side presentation layer keyed on the server's stable actor index — it touches no world state, so combat/trade/etc. are unaffected.
- **`cradle -spectate`** (`cmd/cradle/spectate.go`) — after login, serves a LIVE browser viewport (default `localhost:8089`) that follows the host: `GET /frame` renders a fresh PNG at ~15fps, arrow keys rotate / `+`/`-` zoom, **click-to-walk** (`/walk`: screen→tile pick → `host.Walk`), and agent-readable captures — `p` = `/shot` (single PNG) and `c` = `/clip` (contact sheet), both persisted to `/tmp/render_out/spectator/` with timestamps so a coding agent can read its own host's screen.
- **`cmd/rendertest/`** — headless renderer harness that proves the archive decode + single-frame render against the OpenRSC cache.

This is the new "the host can see, and so can we" capability. It does not change the body/DSL surface; it is an observability + agent-self-perception tool. (The deeper engine internals + the remaining work live in `docs/render-port-plan.md`, owned by the human — do not duplicate it here.)

**Live-test catalog — ~70/88 effective** (64/81 single-host + 6/7 multi-host), up from a 16 baseline. Engine gaps fixed live this cycle: dialog `answer` off-by-one + clear-between-menus (chained menus); `pickpocket`/`npc_command` verb; inventory-mirror opcode-123 remove-and-shift (`wipeinv`); trade split into `confirm_trade`/`finalize_trade`; `interact_at(view)`; facts-on name resolution; `setstat` level-first + `setcurrentstats`; cooking off the quest-gated range. Keystone realization: **scenarios are gap-finders — fixes land in the engine, not worked around in content.**

**API FROZEN (v1, ratified 2026-05-29 — committed + pushed as `38ef5a0`):**
- [`docs/lang/api.md`](lang/api.md) — the **frozen host-facing surface (v1)**: faculties (View/Action/Event), value types incl. **Def vs Instance** (`ItemDef` vs `InvSlot{idx,def,quantity}`), the **capability boundary** (GUI-player equivalence — perceive only what a player sees, do only what a player can click; encoding ≠ capability, so ids are fine), the **control plane** (flow / `note` / `resolve` recognition / `command` admin — fenced, non-GUI), namespacing rules, and the **full §8 per-namespace reference** (self/inventory/world/combat/magic/prayer/trade/bank/duel/shop/ambient/control, every entry tagged exists/rename/to-build). Amend only by deliberate decision, never by drift.
- [`docs/lang/build-backlog.md`](lang/build-backlog.md) — every gap centralized by layer, tagged DONE/BUILD/RESEARCH/REFACTOR/MIND-OUT/CONTENT, + §10 spec↔impl drift to reconcile.
- [`docs/lang/protocol.md`](lang/protocol.md) — the wire-level shadow of api.md (opcodes / encodings / handler quirks).
- [`docs/lang/writing-routines.md`](lang/writing-routines.md) — the living host-facing scripting guide (still flat-named; migrated to namespaced in the wave-2 rename).

## BODY BUILD-OUT — in progress (the freeze's payoff)

Bringing the implementation up to the frozen surface, structured to **fan out** (we were too serial). Architectural decision driving the parallelism: **split the two DSL hub files** (`dsl_actions.go` / `dsl_views.go`) into per-namespace impl files (`actions_<ns>.go` / `views_<ns>.go`) behind **one central registration table** (the whole surface legible at a glance in one file; bulky impl distributed). The hub monoliths were *why* the body work was serial — split them and per-namespace agents stop colliding.

- **Wave 1 (running, worktree-isolated):** (A) the mechanical hub split — no renames, build+parse-gated → `wt/split-hubs`; (B) **#116** decode opcode-234 combat/health into new world-mirror fields → `wt/wire-combat`; (C) **#120** cognition `resolve()` + learned-alias store (new `cognition/resolve/` package) → `wt/cognition-resolve`. Disjoint files; merge order split → wire → cognition, stitching `resolve`'s builtin registration into the central registry at merge.
- **Wave 2 (after the split merges):** the per-namespace fan-out the split unlocks. *Stage 1* — one agent establishes the **namespaced + Def/Instance surface** (the §10 rename across registry/spec/impl + scenario migration + regen, build+parse-green = the old "#115b"). *Stage 2* — per-namespace agents fan out the additive **#117** perception accessors / **#118** verbs / **#119** events (combat target/health/engaged, `magic.cast` unify, `prayer.active`, `bank.deposit_all`/`close`, `shop.*`, `self.combat_style`/`equipped`/`position.plane`, `world.messages` + `on message(pattern)`, `on xp_gain(skill)`), each owning its own files.
- **Then** **#121** content fixes (data, not engine) + **#122** quarantine / re-merge / sweep the 195 proposed scenarios against the namespaced surface + **#123** protocol.md thorough re-run.

**In parallel ABOVE the body:** cognition/brain (#98–100) + the personas/reveries design — independent of the freeze, openable at will (Alex wants to pair on reveries/persona before Phase 3).

**Commit discipline:** the freeze docs are committed + pushed (`38ef5a0` on `main`). Each merged wave branch gets its own commit after `go build ./...` + `go run ./cmd/parsecheck` green; **grep the staged diff for the password literal before every commit**; the 195-merge in `scenarios.yaml` stays parked as WIP until #122.

Re-prioritized todo: tasks **#114–123**. **#114 done**; #115/#116/#120 in flight. Older pending: `#93` (super/extends → Phase 4), `#95` (combat-style live-verify → folds into #117).

> **Host** — an autonomous AI actor in the system. One host = one
> running `cradle` process = one logged-in OpenRSC character that
> perceives the world, thinks, and acts on its own. Throughout
> this codebase and these docs, the AI agents we build are
> **always** called *hosts* — never "bots," never "agents."

## Where Phase 2.5 actually landed

The full design lives in [`docs/lang/`](lang/) and the closed task
list is in [`docs/tasks.md`](tasks.md). Highlights of what's now
live:

- **Stage 1** ✓ — Result/Error + bang variants (#51), filename ↔ routine-name + ParseRoutineString (#53), REPL with `-repl-on-fail` + `.resume` (#54).
- **Stage 2** ✓ — full query layer (~100 accessors across vitals, all 18 skills, equipment, all 14 prayers, 48 spells from SpellDef.xml, inventory, entity views, locs, recent-events ring, combat/bank/trade); `when` / `select` / `defer` / `try`/`recover` / lambdas / validator cohesion pass.
- **Stage 2.5** — live-test discoveries pulled in from Phase 2.8 ahead of schedule: polymorphic `use(item, target)`, `interact_at`, `distance_to` / `in_region`, `.contains()` on last-message records, `world.dialog.options`, `walk_path`, `is_reachable`, `wait_for_dialog`, `event.item_gained`, `world.ground_items.by_id` / `world.npcs.by_type`, `last_attacked_*`, full trade + duel + death state machines, bounds `{...}` region-scoped event filters, and file-level `extends "parent.routine"` library inheritance (#52 v1).

**Deferred to later phases:**

- `#85` `repeat_until(predicate, timeout=Ns)` — needs lazy-eval predicate grammar; no urgency until a routine actually needs the retry-with-timeout pattern.
- `#93` per-handler `on ev() extends host { super(...) }` — waits for Phase 4 persona-tier defaults to define what `host.defaults.<event>` resolves to.

Delores remains one of the build assistants for westworld itself,
not just a test subject. See [`docs/lang/development-workflow.md`](lang/development-workflow.md).

## Next: Phase 2.6 — knowledge ingestion

Phase 2.6 is the next ticket-worthy chunk. See
[`docs/phases.md`](phases.md) for the narrative. Anticipated:
mesa knowledge_chunks schema + Voyage 3 embed pipeline, rsc.wiki
scraper (the background crawler at 3300+/6045 pages is feeding
into this), AutoRune script corpus ingest, `recall()` stdlib
wiring (replaces the Phase-2 stub), `cradle -knowledge-query`
admin CLI. None of these tasks are filed yet — they get tickets
when work starts.

## Where the project-wide layer cake sits

Two complementary views:

- [`docs/layers.md`](layers.md) — **AI-perspective**, 7 layers,
  body → senses → routines → cognition → persona. Read this if
  you want to understand *what the host is doing while it's
  alive*.
- [`docs/architecture.md`](architecture.md) — **package-perspective**,
  12 layers, the same material sliced by Go package. Read this
  when you need to know which file does what.

Both views describe the same system. Layers.md is intentionally
the broader-audience explainer.

This doc captures where the host actually is so a fresh-context Claude
can pick up productively without re-deriving everything from the
codebase + chat. It's deliberately frank about what's verified live
vs. built-but-untested vs. designed-but-not-built — Phase 2 is far
from done.

## Two-line summary

A host (= single OpenRSC connection + world-state mirror + event
bus) navigates, fights, skills, trades, duels, banks, talks to
NPCs, and dies/respawns — all driven by `.routine` programs in a
complete custom DSL (lexer → parser → validator → interpreter →
host bridge, with `when`/`select`/`defer`/`try`/`extends`/lambdas/
`repeat_until`/`wait_until` and a ~100-accessor query layer). Hosts
can also `recall()` rsc.wiki knowledge. As of 2026-05-30 a host's
perceived world also **renders** to a live browser viewport
(`cradle -spectate`) — the host can see, and so can we. The current
frontier is the **BODY build-out** (bringing the impl up to the
frozen v1 API surface — the namespaced Def/Instance rename + the
additive perception/verb/event fan-out), breadth/correctness
validation via the 88-scenario live-test catalog, and the Phase 2.6
knowledge-ingestion build (embeddings, AutoRune corpus). The
cognition/brain/persona/memory/reveries layers above the body remain
**stubs or design only** (see those docs' STATUS banners).

## What's actually verified live against OpenRSC

These have been run end-to-end and confirmed by server logs:

- **Login / logout cycle** for accounts `alex` (admin) and `delores`
  (host, combat lvl 34, atk/str/def all 30).
- **Walking via pathfinder**: multi-corner walk packets. Verified
  with a 25+ tile walk around Lumbridge castle walls. Direct port
  of `mudclient.walkToArea`'s wire format.
- **Open boundary (door)**: pathfinder routes adjacent to the door,
  sends InteractWithBoundary opcode 14, server confirms `onOpBound`
  trigger.
- **Goblin combat loop**: AttackNpc → server walks delores to mob →
  combat resolves → `onKillNpc` server log → ground item drops
  (bones) → PickUpItem with `onTakeObj` confirmation → ItemCommand
  (Bury) with `Bones.onOpInv: Bury` confirmation. Ran 3 kill cycles
  in a single 5-min session.
- **Server-side player follow** (opcode 165): delores trailing
  alex around Lumbridge.
- **PM exchange** with `alex`: AddFriend → PrivateMessage → alex
  echoes back the exact text. PM smart08_16 length prefix bug
  found and fixed during testing.
- **Inbound chat decode**: opcode 234 update-players, opcode 79 NPC
  coords, opcode 131 server messages, opcode 120 PM with body
  decompression.

## Verified live (added since Phase 2.5 start)

Since the original 2026-05-28 snapshot, the following all moved
from "built but unverified" to "live-tested against OpenRSC":

- **TalkToNpc** — works against both static and wandering NPCs
  (pathfinder retries when the target moves).
- **ChooseDialogOption** (opcode 116) — exercised against real
  dialog trees; `world.dialog.options` exposes option text so
  routines can pick by content via `find_option(text)`.
- **AttackPlayer** (opcode 171) — used during duel flows.
- **Trade handshake** end-to-end with both bernard and delores
  driving — symmetric handshake (both sides send 142), two-stage
  confirm dance, server's MessageType.TRADE notification routed
  to `event.TradeRequestReceived`.
- **Bank deposit / withdraw / close** — bernard's `delores_bank_cycle.routine`
  visits Lumbridge banker, deposits, closes cleanly.
- **DropItem** (opcode 246) — used in `delores_drop_and_pickup`
  and as the seed for bernard's looter reactor.
- **Death / respawn** — `on death` handler fires reliably,
  `self.last_death_at` + `self.death_count` populated. Bernard
  losing duels exercised this path fully.
- **Duels** (gentleman's PvP) — full state machine: opcodes 103,
  8, 176, 77, 197, 33, 210, 253, 172, 225, 6, 30 plus the
  reset-accepts-on-offer server behavior. Members-only gate at
  `PlayerDuelHandler.java:39` discovered + worked around.
- **Skills surface** — all 18 RSC skills wired into the query
  layer (#57) including prayer (#59, 14 prayers from
  PrayerDef.xml) and magic (#60, 48 spells embedded from
  SpellDef.xml at init).
- **Dynamic boundary updates** — `world.Boundaries` overrides map
  + decoded inbound opcode 91 boundary updates; cut webs / opened
  doors now reflect in the pathfinder grid.
- **AutoEat watcher** — port of `runtime/auto_eat.go` lives as
  `examples/routines/auto_eat.routine`; Go version deleted.

## Still NOT verified live

- **Combat style toggle** (3 melee styles) — `set_combat_style()`
  builtin + opcode 29 are wired (#105) but the XP-split effect has
  not been observed live yet (#95).
- **Most of the 88-scenario catalog** — the live-test sweep is
  mid-iteration. Many scenarios run their actions but the in-world
  outcome (skill XP gained, item produced, NPC dialog advanced)
  has not been individually confirmed. See `cmd/scenariogen/`.

(The non-stackable inventory amount decoder bug noted in earlier
revisions of this doc was repaired in #94 — `decodeInventory` and
`decodeInventorySlotUpdate` now consult `isStackable`.)

## The DSL — exact status

**DSL Phase 2.5 is DONE.** All 9 step-1-DSL items + the full
Phase 2.5 stage list shipped. Current language surface:

| Feature | Status |
|---|---|
| Lexer + parser + AST + validator + interpreter (steps 1–5) | ✓ |
| Action channel + Host bridge + resource caps (steps 6–7) | ✓ |
| Event handler dispatch + conformance harness (steps 8–9) | ✓ |
| Result/Error model + bang variants (#51) | ✓ |
| ParseRoutineString + filename-match enforcement (#53) | ✓ |
| REPL with `-repl-on-fail` + `.resume` (#54) | ✓ |
| Full ~100-accessor query layer (#46, #56–#65) | ✓ |
| `when` watchers (#47) | ✓ |
| `select { when / on / timeout }` (#48) | ✓ |
| `defer` (#49) | ✓ |
| `try`/`recover` (#50) | ✓ |
| File-level `extends "parent.routine"` (#52 v1) | ✓ |
| Lambdas `IDENT => expr` (#66) | ✓ |
| Validator cohesion pass (#67) | ✓ |
| Tier-1+2 primitives (#75–#92) — `use(item, target)`, `interact_at`, `walk_path`, `is_reachable`, `wait_for_dialog`, `event.item_gained`, `world.ground_items.by_id`, `world.npcs.by_type`, `last_attacked_*`, etc. | ✓ |
| `bounds { ... }` region-scoped event filter | ✓ |
| Trade + duel + bank + death state machines (#91, #92, #28, #29) | ✓ |
| `repeat { body } until <cond> timeout <expr>` (#85) | ✓ |
| Per-handler `extends host` + `super()` (#93 = v2 of #52) | deferred to Phase 4 |
| Stdlib LLM oracles (`exec`, `improvise`, `contemplate_reality`) | stubs, real bridge in delos / Phase 3 |

Conventions established for the DSL:

- File extension is **`.routine`**, not `.ws` (chose with alex; `.ws`
  was historic WordStar / WSF Windows Script File).
- Packages live under `dsl/`: `dsl/token/`, `dsl/lex/`, `dsl/ast/`,
  `dsl/parser/`, `dsl/validator/`, `dsl/interp/`.
- Hand-written recursive-descent — no parser generator, no regex
  in the lexer.
- F-string lexer alternates between literal-fragment mode and
  placeholder mode using two flags (`inFString`, `inPlaceholder`).
- AST nodes implement `Node` (has `Pos()` + private marker) plus
  `Stmt`/`Expr` sub-interfaces for type-safety in the parser.
- Control flow in the interpreter is propagated via panic with
  sentinel types (`returnSignal`, `breakSignal`, etc.) recovered
  at the top of each block — keeps every recursive call site
  clean of plumbing. **Critical:** the recover() inside
  `execLoopBody` must be called exactly once per panic, captured
  into a variable. Calling recover() a second time returns nil
  and swallows the real signal — bug we hit in step 7.
- Truthiness, equality, and numeric promotion follow Python-style
  rules per dsl.md.
- Reserved names `self` / `world` / `inventory` / `combat` are
  bound from `Interpreter.Reserved` at routine startup; entities
  expose attribute access via the `Getter` interface and indexing
  via `Indexer`.
- Primary actions implement `interp.Yielder` so the interpreter
  drains the event queue around them, letting `on` handlers
  interleave between actions.
- Resource caps live in `dsl/interp/caps.go` — default 1M ops,
  4h wall clock, 64-deep recursion, 1024-element lists, 4096-
  char strings. Tests override via `Interpreter.Caps`.
- Observability via `Interpreter.Hooks` — five callbacks
  (RoutineStart/End, Action/AfterAction, Handler, Abort) used by
  the conformance runner today and by delos telemetry in Phase 3.
- Conformance test format: paired `.routine` + `.expected` files
  in `testdata/conformance/`. Adding a new case is two files; no
  test code changes.

## Repo / package layout

```
westworld/
├── action/        — outbound packet helpers (one file per concern):
│                    walk, talk, follow, combat, items, boundary,
│                    bank, trade, prayer, equip, magic, social
├── brain/         — LLM strategist (STUB: deterministic decisions, no LLM)
├── cmd/cradle/    — single-host CLI driver + REPL + `-spectate` viewport
├── cmd/rendertest/ — headless render harness (proves .orsc decode + frame)
├── cmd/parsecheck/ — parse-gate (build/CI sanity)
├── cmd/scenariogen/ — scenario generator (live-test catalog)
├── cmd/{delos,mesa}/ — EMPTY (no .go yet; delos = planned telemetry /
│                    chain-of-thought, mesa = planned memory service)
├── cognition/     — retrieval client (STUB; mesa wiring in Phase 2.6).
│                    resolve/ (resolve() alias store, REAL) + corpus/
├── render/        — SHIPPED Go RSC viewport renderer (.orsc decode →
│                    PNG; terrain/boundaries/roofs/scenery/actors)
├── docs/          — design docs (architecture, brain, dsl, lang/, etc.)
├── dsl/
│   ├── token/     — token kinds + Position + Token
│   ├── lex/       — lexer (state-machine, hand-written)
│   ├── ast/       — AST node types
│   ├── parser/    — recursive-descent parser
│   ├── validator/ — static validation
│   ├── interp/    — AST interpreter + event dispatch + caps
│   ├── spec/      — language surface tables (actions / events / accessors)
│   └── conformance/ — golden-trace test runner
├── event/         — typed event types + event.Bus pub/sub
├── examples/routines/ — example .routine files + common/ libraries
├── facts/         — static OpenRSC defs + locs (loaded once per process)
│                    includes SpellDef.xml + PrayerDef.xml embedded
├── pathfind/      — BFS, grid, sector loader, multi-corner walk encoding
├── proto/v235/    — wire format: framing, opcodes, ISAAC, RSC compression
├── runtime/       — Host (per-host stateful object) + DSL bridge:
│                    follow.go, combat.go, items.go, boundary.go,
│                    bank.go, trade.go, prayer.go, magic.go,
│                    dsl_bridge.go, dsl_actions.go, dsl_views.go,
│                    dsl_events.go, host.go
├── session/       — TCP/ISAAC session wrapper
└── world/         — per-host world-state mirror (Self, Inventory, Npcs,
                     Players, GroundItems, Bank, Boundaries, Trade, Duel)
```

## Important runtime concepts

### "Reactor patterns" are now routines

Earlier we had Go-coded long-running goroutines (`runtime.CombatLoop`,
`runtime.AutoEat`) that drove Host actions in response to events.
With Phase 2.5 complete, both are deleted (#55) — the equivalent
routines live in `examples/routines/auto_eat.routine` and
`examples/routines/combat_loop.routine`, expressed via `when` +
`on` + `select` with no Go-side reactor needed.

### "Host" = one cradle process

`runtime.Host` is the per-host object. Owns: TCP session, world
mirror, event bus, facts pointer (shared across hosts in a swarm),
pathfind landscape archive (also shared), logger.

### Walk packet quirks

`mudclient.walkToArea` sends `firstStep` + signed-byte deltas of
direction-change corners. `Path.addStep` server-side interpolates
between corners — only the *final* tile of each addStep gets a
`player-blocking` check. Sending tile-by-tile waypoints (one per
inter-corner tile) makes every intermediate get the player-blocking
check, which truncates the walk if any other player is on a tile.
Always send corner-compressed.

### Server-side action handler pitfalls observed

- `GroundItemTake` calls `player.resetPath()` if the item isn't in
  view — sending walk+take in one burst nukes the walk. Two-phase
  required: walk first, poll until within view radius, then take.
- `GameObjectWallAction` (boundaries) only sets a WalkToAction, never
  initiates walking. The host must send a Walk packet to the boundary's
  tile before the interact packet.
- NPC pathing handlers `setFollowing` + per-tick `walkToEntity` only
  emit 1-step paths that the WalkingQueue silently drops if the next
  tile isn't adjacent — so any NPC further than 1 tile needs an
  explicit precursor walk.
- `Path.MAXIMUM_SIZE = 50` total tiles. Multi-corner walks longer
  than this get silently truncated.

## Security note

The OpenRSC test password leaked twice into committed files
(once as a default flag value in `cmd/cradle/main.go`, once into
an earlier version of this doc). Both leaks were scrubbed via
`git filter-repo --replace-text` + `--replace-message` and
force-pushed to origin.

The host reads its password from the `WESTWORLD_PASSWORD` env var
at runtime. **Never** embed the literal value anywhere — code,
docs, commit messages, log lines, none of it. When discussing
the incident, refer to it as "the OpenRSC test password" and
nothing more.

Rotate the OpenRSC-side password if the value has any reuse
risk elsewhere — even after force-push, GitHub caches the old
commit blobs for some time.

Before committing: `git diff --cached | grep -i <known-secret>`.

## Open design questions for next session

From `docs/dsl.md` "Open language-design questions":

- Range syntax `..` vs `..=` — currently `..` per dsl.md.
- F-string syntax `f"hi {name}"` vs shell-style — Python f-strings
  chosen.
- Truthiness rules — Python-style "many falsey" chosen, codify in
  interpreter step.
- Equality — value equality only chosen.

Nothing blocks step 2 (statement parser) on these.

## Quick commands for next session

```bash
# Build everything
cd ~/Code/westworld && go build ./...

# Run all tests
cd ~/Code/westworld && go test ./...

# Run a host with a single action
export WESTWORLD_PASSWORD=...
cd ~/Code/westworld && /tmp/cradle -username delores -look-around -dwell 5s

# Server logs (server-side trigger plugins log to here)
tail -f /Users/flint/Code/openrsc/server/logs/rsc_preservation_1.log

# Latest pcap per player (for wire-level debugging)
ls -lt "/Users/flint/Code/openrsc/server/logs/pcaps/RSC Preservation/<name>/2026-05/" | head

# Decode a pcap
gunzip /tmp/x.pcap.gz && tcpdump -r /tmp/x.pcap -A -x -nn | head
```

## How to pick up cleanly

1. `git pull` on `~/Code/westworld`.
2. `go test ./...` — every package should be green.
3. **Phase 2.5 is closed**; the next ticketable chunk is
   **Phase 2.6 (knowledge ingestion)**. Anticipated order
   (none filed yet — file tickets when starting):
   - Mesa knowledge_chunks schema + Voyage 3 embed pipeline
     (Postgres + pgvector behind a small HTTP surface).
   - rsc.wiki scraper + ingest. The background crawler is
     archiving 6045 pages; chunk by section heading, ~500–800
     tokens/chunk.
   - AutoRune script corpus ingest — Alex's historical AutoRune
     scripts as first-class "how would a script normally do X"
     reference.
   - `recall(query, top=N)` stdlib wiring (replaces the Phase-2
     stub in `cognition/`).
   - `cradle -knowledge-query` admin CLI for sanity-checking
     the corpus.
4. **REPL is the day-to-day driver.** `cradle -repl` or
   `cradle -repl-on-fail` (with `.resume`) is how new routines +
   primitives land. Author live with delores or bernard, save once
   it runs clean, then commit.
5. The `cradle -routine path.routine` flag is live for headless
   runs. Author your test routine, run it, watch the log + tail
   the server side at
   `~/Code/openrsc/server/logs/rsc_preservation_1.log`.
6. Adding a conformance test is two files in
   `testdata/conformance/` — no code changes. Use the existing
   cases as a template.
7. Reveries / persona LOE design session is on the table —
   alex wants to pair on it before Phase 3. See the
   `reveries-persona-LOE` memory.

**Reminder about secrets:** the OpenRSC test password leaked twice
already. Read it from `WESTWORLD_PASSWORD` env var only; do NOT
embed it in any file, doc, or commit message — not even when
documenting incidents. Grep the staged diff for the literal
before committing.
