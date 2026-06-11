# Writing Routines — practical guide + RSC scripting nuances

> **STATUS: CURRENT** (verified against code 2026-06-10, HEAD `0b35f43`).
> Living document: every entry below was learned by running routines live and
> watching them fail. When you discover a new nuance, add it here.
> Open work lives in [`docs/TODO.md`](../TODO.md) (SSOT) — §10 below cites the
> relevant IDs.

A companion to `syntax.md` (grammar) and `actions.md` (verb reference). This
file is the **operational** guide: how to actually make a routine *do the right
thing* against the live OpenRSC server, and the non-obvious nuances that cause
silent failures. It is meant to be loaded into a host's context (or mesa's RAG)
so a bot can write correct routines on the first try.

---

## 1. The model

- One routine = one logged-in character executing one `.routine` program.
- Filename (snake_case) **must** equal the declared `routine` name.
- A routine `return`s a value (success) or `abort`s (failure). Test scenarios
  use `return "PASS: ..."` / `abort "FAIL: ..."`.
- Actions are sent to the server and resolve asynchronously. The world you read
  (`self`, `world`, `inventory`) is a **local mirror** updated from server
  packets — it lags the action by a tick or more. Always *wait for the effect*,
  never assume an action took hold synchronously.

## 2. Reading vs. doing

- `self.*` — your vitals, skills, position, equipment.
- `world.npcs / players / locs / ground_items / scenery / boundaries` — visible
  entities (`locs` = static placements from the landscape; `scenery` = live
  GameObjects, e.g. a fire you just lit — see §7).
- `trade / duel / bank` — top-level interaction-subsystem roots (state + verbs);
  `world.dialog` — dialog-menu state. (`world.trade / duel / bank` stay aliased
  for back-compat but new code uses the top-level roots.)
- `magic / prayer` — top-level subsystem roots: `magic.cast(...)` + the spell
  catalog (`magic.book / known`), `prayer.activate(...)` + the prayer catalog.
- `shop` — top-level root for the open shop window: `shop.is_open / is_general /
  slots / stock(item) / price(item)` reads + `shop.buy(item, qty) /
  sell(item, qty) / close()` verbs (`runtime/views_shop.go`,
  `runtime/actions_shop.go`). The window opens through the shopkeeper's dialog
  (`talk_to` + `answer`) — there is no `shop.open` verb yet (TODO.md DSL-7).
- `inventory.find(id) / find_all(id) / count(id) / used / free / is_full`.
- `world.messages` — server-message log (`List<Message>`; each has
  `.text / .kind / .at / .contains("...")`). Assert on server text with
  `world.messages.last.contains("...")` (or the back-compat
  `world.last_server_message.contains("...")`).
- Verbs: see `dsl/spec/actions.go` + `dsl/spec/accessors.go` (authoritative) —
  `walk_to`, `walk_path`, `go_to` (cross-region travel; takes coords or a town
  NAME, never a POI type), `use`, `interact_at`, `magic.cast`,
  `prayer.activate`, `equip`, `combat.attack` (alias `attack`), `talk_to`,
  `converse` (listen to an NPC — §7a), `pickpocket`, `eat`, `drop`, `scan_for`
  (enumerate nearby scenery by type — §7), `search_map` (rank reachable POI
  destinations), `trade.*` / `duel.*` / `bank.*` / `shop.*` verbs, `note`, …

## 3. Waiting — pick the right primitive

| want | use |
| --- | --- |
| fixed pause | `wait 2`  (seconds; bare number, decimals ok: `wait 0.5`) |
| block until a condition (poll ~200ms) | `wait_until(_ => <cond>, <timeoutSecs>)` |
| retry an action until it works | `repeat { <action>; wait N } until <cond> timeout <secs>` |
| block on one of several events | `select { on <event> {...} timeout 10s {...} }` |

**Timeout literal trap:** inside `select { ... timeout 10s { } }` the `Ns`
duration suffix is required. Inside `repeat ... until <cond> timeout <expr>` the
timeout is a **plain expression in seconds** — write `timeout 45`, **not**
`timeout 45s` (the `s` lexes as a stray identifier → "unbound identifier s").

## 3a. Building strings — `format()` first

`format(template, args...)` is the primary way to build a string — a pure
function, legal anywhere an expression is. `{}` placeholders fill
positionally, left-to-right; `{{`/`}}` escape literal braces; each argument
renders exactly as an f-string placeholder would:

```
command(format("fatigue {} 80", self.name))
note(format("have {} gp", inventory.count("coins")))
```

- Placeholders are bare `{}` **only**. `{name}` inside a format template is
  literal text, not a lookup (the validator warns — it is almost always a
  mistake).
- Literal templates are arity-checked at validation time (`{}` count must
  equal argument count); a computed template defers the check to runtime
  (`FORMAT_MISMATCH`).
- F-strings still work as the inline secondary form
  (`f"fatigue {self.name} 80"`): one expression per `{}`, plain double quotes
  for nested strings, never backslash escapes. When a placeholder gets
  complex, switch to `format()` or bind a local first.

## 4. Admin setup commands (test scenarios)

Run via `command("...")`. These need the account to be a moderator/admin.

- `item <id> <amount>` — spawn to your own inventory (id self-targets).
- `teleport <x> <y>` — move yourself. **Teleport is async + reloads the region**
  — the NPC/scenery mirror repopulates a beat later (see §6).
- `summon <name>` — pull another logged-in player to you.
- `setstat <level> <stat>` — **LEVEL FIRST**, stat second (e.g. `setstat 40
  cooking`). A leading number self-targets. Sets max + xp + base level.
- `setcurrentstats <level> <stat>` — sets the **effective/current** level. Skill
  and spell *level checks* read the current level, so **always pair them**:
  ```
  command("setstat 25 magic"); command("setcurrentstats 25 magic")
  ```
  Setting only `setstat` will pass the catalog requirement but a spell needing
  level 25 can still fail its in-the-moment check.
- `fatigue <name> <pct>` — needs a NAME arg (no numeric self-target). In a body:
  `command(f"fatigue {self.name} 80")`.
- `wipeinv <name>` — needs the name: `command(f"wipeinv {self.name}")`. Bare
  `wipeinv` is a no-op. (The old local-mirror gap is FIXED — the per-slot
  removal burst now empties the mirror correctly; see §10.)
- `heal`, `recharge`, `damage <name> <n>`, `kill <name>`, `spawnnpc <id> <radius> <mins>`.

Skill short-names (for the admin commands): attack, defense, strength, hits,
ranged, prayer, magic, cooking, woodcut, fletching, fishing, firemaking,
crafting, smithing, mining, herblaw, agility, thieving.

⚠ The admin short-name can differ from the **`self.skills.<name>` accessor**.
Notably `setstat 40 woodcut` (command) vs `self.skills.woodcutting.xp`
(accessor — full word). When in doubt, the accessor names are in
`runtime/views_self.go`'s `skillIDs` map.

## 5. Facts must be loaded

The runtime resolves `type_id → NPC name` and `def_id → scenery name` from the
OpenRSC defs ("facts"). Every runner binary takes `-facts <openrsc-root>`
(`cmd/cradle-server`, `cmd/legacy-cradle`, `cmd/host` — default
`/Users/flint/Code/openrsc`); run with the default, NOT `-facts ""`. With facts
off, every `n.name == "..."` and
`world.locs.search("tree")` returns nothing and your routine aborts as "not in
view / not nearby" even though the entity is right there.

## 6. The post-teleport race

After `teleport`/`summon`, the entity mirror loads asynchronously. Binding
immediately gets `null`. Poll first:
```
command("teleport 135 663"); wait 1.5
wait_until(_ => world.npcs.find(n => n.name.lower == "cook") != null, 8)
c = world.npcs.find(n => n.name.lower == "cook")
```
Corpus convention: every setup `command(...)` is followed by `wait 1.5` (the
scenario corpus does this throughout), which covers *item grants* and *stat
sets*; region/NPC loads still want the explicit poll. (`cmd/scenariogen` no
longer generates routine bodies — the `.routine` files under
`examples/scenarios/` are the source of truth; it only validates the manifest.)

## 7. Interactions — pick the RIGHT verb per mechanic

This is the #1 source of silent "nothing happened" failures. The interaction
verb depends on how the server models the action:

- **Gathering — woodcutting / mining / fishing → the scenery COMMAND, not the
  tool.** Use `interact_at(spot)` (sends command1: Chop/Mine/Net/Lure/…). The
  tool just needs to be in your inventory. `use(axe, tree)` returns "Nothing
  interesting happens" unless the server has `GATHER_TOOL_ON_SCENERY` on (it's
  off on preservation). Also `batch_progression` is OFF → one click = one
  attempt, so **reclick in a loop**:
  ```
  rock = world.locs.rocks.nearest(self.position)
  repeat { interact_at(rock); wait 2 } until inventory.find(202) != null timeout 60
  ```
  Prefer `scan_for("rock")` over `world.locs.*` for target-picking: it merges
  the static map with the live view, so depleted/removed objects drop out and a
  freshly-lit fire / regrown tree shows up. Iterate the ranked list instead of
  hardcoding a tile.
- **Cooking / smelting → `use(item, scenery)`.** `use(raw_food, fire_or_range)`,
  `use(ore, furnace)`. Fire cooking needs a fire lit first (firemaking).
- **Firemaking → DROP the logs, then `use(tinderbox, ground_log)`.**
  `use(tinderbox, logs_in_inventory)` only nags "...put the logs down before you
  light them!" (opcode 91). The real light is the ground path (opcode 53):
  ```
  drop(inventory.find(14))                       # only plain Logs (14) light here
  gl = world.ground_items.by_id(14, radius=0)
  use(inventory.find(166), gl)                   # tinderbox on the GROUND log
  wait_until(_ => world.scenery.by_id(97, radius=0) != null, 5)
  fire = world.scenery.by_id(97, radius=0)       # the fire is SCENERY, not a ground item
  ```
  A lit fire is a **scenery GameObject def 97**, delivered via
  `SEND_SCENERY_HANDLER` (opcode 48) — read it from **`world.scenery`**, NOT
  `world.ground_items.by_id(97)` (a fire is not a ground item) and NOT
  `world.locs.search("fire")` (that only knows PRE-PLACED fires from the static
  landscape, never one you just lit). Fires persist ~90s, so re-running on the
  same tile hits "You can't light a fire here" — rotate teleport tiles in a
  `repeat ... until` to dodge a still-burning tile.
- **`interact_at` argument forms** (it takes a VIEW, not bare coords):
  - `interact_at(view)` — primary command (option 1).
  - `interact_at(view, 2)` — secondary command (e.g. stalls "steal from" is opt 2).
  - `interact_at(x=N, y=N)` — explicit coords (option defaults to 1).
  - ✗ `interact_at(rock.x, rock.y)` — WRONG: the 2nd positional is read as the
    option, so you get "option must be 1 or 2, got <y>".
- **Boundaries (doors / webs) → `world.boundaries.at(x, y, dir)`**, not
  `world.locs.doors.nearest()` (that returns a *placement*, and
  `open_boundary`/`use` want a boundary view). Cut a web:
  `use(knife, world.boundaries.at(208, 547, 2))`.
- **Smithing → `use(bar, anvil)` with a HAMMER (id 168) in inventory.** The bar
  is the trigger; the hammer is a required item (not the trigger). The anvil
  opens a **two-level menu** (category → item) — you must answer BOTH menus, not
  one. (See §7a — multi-step dialog navigation.)
- **NPCs:** `talk_to(npc)` opens dialog. `pickpocket(npc)` (the canonical
  NPC-command verb) fires the NPC's primary action command (command1 — e.g.
  "Pickpocket" on a Man).
  Like gathering, the skill action is one-attempt-per-call, so loop:
  `repeat { pickpocket(man); wait 3 } until self.skills.thieving.xp > prev timeout 30`.

## 7a. Dialog menus (talk / smithing / crafting / banking)

`answer(idx)` is **1-based** (`answer(1)` = first option), matching
`find_option(text)` which returns the 1-based index of the first option
containing `text` (0 = no match → answering it errors, by design). After each
`answer()` the menu is cleared, so a follow-up `wait_for_dialog()` blocks until
the NEXT menu arrives. **Multi-level menus** (smithing: category → item) need one
`answer` per level:
```
use(bar, anvil)
wait_for_dialog(8); answer(find_option("Weapon"))   # menu 1
wait_for_dialog(8); answer(find_option("Dagger"))   # menu 2
```

`converse(npc)` wraps the whole talk loop when you want to LISTEN rather than
drive: it opens the dialog, aggregates everything the NPC says, auto-advances
only the unambiguous prompts (a lone "continue", an all-exit menu, a banker's
bank-access option), and STOPS at any real multi-option choice with the menu
still open. It returns `{said, options, ended, answered}` — `options != null`
means a real choice is waiting: read it, `answer(find_option("..."))`, then
call `converse(npc)` again. There is no topic argument — NPCs only speak their
pre-authored lines (`converse("npc", "pickaxe")` will NOT ask about pickaxes).

## 8. Multi-party (trade / duel)

- Trade is two screens, each with its own button:
  - `trade.accept()` — "Accept" on the **offer** screen (idempotent).
  - `trade.confirm()` — "Confirm" on the **final** screen.
  Both parties must accept the offer screen before the final screen opens. Any
  offer change resets both first-accepts. Drive it event-by-event (note the
  event name `trade_request` stays underscore-flat — events are not verbs):
  ```
  trade.request(b)
  select { on trade_opened(o) {} timeout 15s { abort "no open" } }
  trade.offer([[10, 100]])
  trade.accept()
  select { on trade_other_accepted() {} timeout 15s { abort "no accept" } }
  trade.confirm()
  select { on trade_closed(done) { ... } timeout 15s { abort "no close" } }
  ```
- Duel is parallel: `duel.request(b)` → `duel.set_rules(...)` / `duel.stake(...)`
  → `duel.accept()` (offer screen) → `duel.confirm()` (final screen).

## 8a. Quest-gated scenery / NPCs

Some scenery is gated by quest progress. The Lumbridge **Cook's Range**
(`SceneryId.COOKS_RANGE`) is dead to anyone who has *started* Cook's Assistant
(`ObjectCooking.onUseLoc`: if `questStage(COOKS_ASSISTANT) > -1` it silently
refuses — no message if no cook is adjacent). Since test hosts are reused and
may have started quests, prefer **non-gated equivalents** (a plain `Range`, def
11, in Varrock/Al-Kharid) over the quest-flavoured one. General rule: if a
"use"/interact on the right tile produces *zero* server response, suspect a
quest/members/region gate before blaming the cradle.

## 9. Assert REAL outcomes (no false positives)

A test that spawns item X and asserts `inventory.find(X) != null` tests nothing.
Assert the *result of the action*:
- xp gained: capture `prev = self.skills.cooking.xp`, assert `> prev` after.
- new item produced (different id from any input).
- a state/position/vitals change, or a specific `world.messages` entry
  (`world.messages.last.contains("...")`).

## 10. Engine / DSL gaps — resolved

Everything this section used to track as open is FIXED. The earlier fixes
(`pickpocket` as the canonical NPC-command verb, 1-based multi-step dialog
navigation, the DROP-then-light fire-cooking chain) are folded into §7/§7a
above. The later four, for the record:

- **Bulk inventory removal mirrored.** Opcode 123 decodes as a remove-and-shift
  event (`proto/v235/inbound.go` → `event.InventoryRemoveSlot` →
  `world.Inventory.RemoveSlot` in `world/world.go`), so a `wipeinv` burst at
  slot 0 empties the whole local mirror, matching the server's
  `ArrayList.remove(index)` semantics.
- **Shop verbs + view.** Top-level `shop` root with reads + `buy/sell/close` —
  see §2.
- **`self.position.plane`.** Floor/plane accessor (`dsl/spec/accessors.go`
  `self.position.plane`; `runtime/views_self.go` positionView) — 0 ground,
  1+ upper, 3 underground.
- **`use(item, solid scenery)` reach.** `UseItemOnScenery` / `interact_at` walk
  INTO the def's server-exact interact rect before sending
  (`runtime/boundary.go` — `sceneryRectAt` + `approachAndAct`, mirroring the
  server's `GameObject.getObjectBoundary`), awaiting arrival first; type-2/3
  usable-direction objects no longer silently drop out-of-rect clicks.

Open work: see [`docs/TODO.md`](../TODO.md) — §2 "DSL surface, spec & manual"
(e.g. DSL-1 reachable-by-default selectors, DSL-7 `shop.open`, DSL-11
structured skilling outcomes) and §1 MP-11 (`use(item, sceneryView)` residue in
`runtime/actions_use.go`).
