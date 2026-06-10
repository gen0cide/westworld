# Actions — The THEN THAT Layer

> **STATUS: CURRENT — verified 2026-06-10 against branch
> `tidy/structure-and-docs`, HEAD `0bfa818`.** The typed error model
> and the action surface below are BUILT and executing. Open action
> work is tracked in [`docs/TODO.md`](../TODO.md) (the SSOT — IDs
> like `DSL-6`/`DSL-24` cited inline); this doc carries no backlog.

> **Host** — an autonomous AI actor. Actions are the verbs a host
> can use to change the world.

## Status

- **Typed error model: BUILT.** `dsl/interp/error.go` defines the
  `ErrorCode` enum, the `Error{Code, Reason, Fatal}` value, the
  `CallResult{Val, Err}` shape every fallible callable returns, the
  `Ok`/`Fail`/`FailFatal` constructors, and `BangCallable` (the `!`
  variant wrapper).
- **Action registry: BUILT.** `runtime/dsl_actions.go` is the hub —
  the central `actionHandlers` name→handler map plus the per-
  namespace verb tables (`tradeVerbs`, `bankVerbs`, `duelVerbs`,
  `magicVerbs`, `prayerVerbs`, `combatVerbs`). Handler bodies live
  in the per-namespace `runtime/actions_*.go` files.
- **Registration is spec-driven.** `dsl/spec/actions.go` is the
  canonical table of every callable (name, kind, arity, params,
  catalog tags, doc line). The bridge
  (`runtime/dsl_bridge.go::NewRoutineInterpreter`) iterates it,
  binds each entry to its handler (or a `NOT_IMPLEMENTED` stub when
  `NotYetImplemented`), and auto-registers the bang variant for
  every bang-eligible kind. `dsl/spec/consistency_test.go` catches
  spec/handler drift.
- **Policy gate: BUILT.** When `Host.Pearl` is wired, every flat
  `PrimaryAction` handler is wrapped by the pearl gate
  (`runtime/dsl_bridge.go:287` → `runtime/pearl.go::gateAction`):
  the host's own compiled policy can allow, substitute, or veto the
  action before a packet is sent. A veto returns a typed
  `POLICY_VETO` result and publishes an `event.PolicyVeto` so
  cognition sees its own refusal. ANALYSIS-mode operator commands
  opt out via `WithoutPearlGate()`. Namespace-dispatched verbs
  (`trade.*`, `bank.*`, …) currently dispatch ungated through
  `runtime/namespace_actions.go`.
- **Still stubs** (registered, return `NOT_IMPLEMENTED`): the skill
  verbs `mine`/`fish`/`chop`/`cook`, the LLM verbs
  `exec`/`improvise`/`reflect_now`, the memory verbs
  `wait_for_chat`/`observe`, and the persona reads
  `mood`/`motivation`.

## The error model

### Go side

Every action's failure mode maps to a typed `ErrorCode`
(`dsl/interp/error.go`). The full enum today:

| Group | Codes |
|---|---|
| Movement | `PATH_BLOCKED`, `OUT_OF_RANGE`, `DOOR_LOCKED` |
| Inventory | `INVENTORY_FULL`, `INVENTORY_EMPTY`, `NO_SUCH_ITEM` |
| Combat / targets | `TARGET_DEAD`, `TARGET_OUT_OF_VIEW`, `RETREAT_TOO_EARLY`, `EAT_IN_COMBAT` |
| Session | `NOT_LOGGED_IN`, `INTERRUPTED` |
| UI / sub-system | `BANK_NOT_OPEN`, `TRADE_NOT_ACTIVE`, `DIALOG_NOT_OPEN`, `SHOP_NOT_OPEN` |
| Misc | `ACTION_TIMEOUT`, `SERVER_REJECTED` (server said no, with prose), `NOT_IMPLEMENTED` (stub), `POLICY_VETO` (the host vetoed itself — the server never saw it) |

Action wrappers construct results with `interp.Ok(value)` /
`interp.Fail(code, reason)` / `interp.FailFatal(code, reason)`.
Go errors coming back from `Host` methods are classified into codes
by `runtime/dsl_actions.go::wrapServerErr` — typed sentinels first
(a wrapped `*DoorLockedError` surfaces as `DOOR_LOCKED` with door
coords + server prose), then message-substring matching, with
`SERVER_REJECTED` as the catch-all.

### DSL side

Actions return a result value with two fields:

| Field | Type | Notes |
|---|---|---|
| `result.val` | Value | The action's return value (`null` for most actions, an item-view for `pick_up()`, etc.) |
| `result.err` | Error or null | `null` on success; otherwise the error |

The `Error` shape:

| Field | Type | Notes |
|---|---|---|
| `error.code` | String | SCREAMING_SNAKE — `"PATH_BLOCKED"`, etc. |
| `error.reason` | String | Human-readable detail |
| `error.fatal` | Bool | Should the routine give up entirely? Set on unrecoverable failures (server disconnect, etc.) |

(The Go type behind `result` is `interp.CallResult` — named that to
avoid colliding with the routine-completion `Result`; DSL code only
ever sees `.val` / `.err`.)

### Three error idioms

```
# 1. Bang form — assert success, abort on failure (most common)
walk_to!(x=120, y=504)
eat!("lobster")
attack!(target)

# 2. Explicit result check — inline handling
result = walk_to(x=120, y=504)
if result.err {                          # truthy if any error
    if result.err.code == "PATH_BLOCKED" {
        try_alternative_route()
    }
    elif result.err.code == "OUT_OF_RANGE" {
        abort "too_far"
    }
}

# 3. try/recover — boundary for a chain of bangs
try {
    walk_to!(x=spot.x, y=spot.y)
    pick_up!(loot)
    eat!("lobster")
} recover err {
    note(f"chain failed: {err.code}")
    if err.code == "PATH_BLOCKED" { try_alternative_route() }
    else { abort err.code }
}
```

The bang isn't a syntactic decorator — it's a separate callable.
`eat` and `eat!` are registered as two different names in
`Interpreter.Builtins` (`runtime/dsl_bridge.go:292`). On failure the
bang panics with the typed abort signal — the routine ends
`ResultAborted` (or `try`/`recover` binds the Error); on success it
unwraps and returns `.val` directly, so `picked = pick_up!(item)`
binds the item-view, not a result shell. The validator knows both
names; tooling (REPL `.help`, the generated manual) treats them as a
pair.

### Which callables get bang variants

The rule is uniform and lives in
`dsl/spec/actions.go::ActionSpec.BangEligible`: **anything that
returns a typed result gets a bang variant.** By `ActionKind`:

- **`PrimaryAction`** — every state-mutating verb (walk_to, attack,
  eat, use, talk_to, say, logout, …). Bang-eligible.
- **`LLMStdlib`** — contemplate_reality, evaluate, decide, exec,
  improvise, reflect_now. Bang-eligible.
- **`MemoryStdlib`** — recall, relation_with, remember, recollect,
  forget, wait_for_chat, observe. Bang-eligible.
- **Namespaced verbs** — `trade.request!`, `bank.deposit!`, etc.
  also work; `runtime/namespace_actions.go` strips the bang and
  wraps the bound callable in a `BangCallable`.

What does **not** get a bang variant:

- **`Primitive`** — wait, wait_until, note, look_around, the
  spatial/bounds helpers, the map-perception verbs, resolve.
  Local-only; cancellation flows through ctx-cancel, not
  `result.err`.
- **`PersonaRead`** — mood, motivation. Pure reads.
- **Procs** — user-defined helpers return whatever they return;
  bang wouldn't make sense.

The validator knows the bang-or-not status of every registered
callable and rejects `note!()` etc. at parse time
(`dsl/validator/validator.go`).

## The action menu

Categorized list of every action the language supports, mirroring
the namespace order in [`api.md`](api.md). `error_codes` is the set
of failure modes the action might emit.

### Movement

```
walk_to(x = int, y = int)
walk_to(position)
walk_to(x = int, y = int, attempt_open_doors = bool)   # default: true
```
- error_codes: `PATH_BLOCKED`, `DOOR_LOCKED`, `OUT_OF_RANGE`, `INTERRUPTED`
- blocking: yes (returns when arrived or fails); local-region
  pathfinding only — for cross-region travel use `go_to`
- **Auto-opens closed doors by default** (`attempt_open_doors=true`,
  `runtime/traverse.go::DefaultWalkOptions`). When the walk stalls
  adjacent to an openable boundary (door / doorframe / gate that has
  an Open action), walk_to interacts with it and re-pathfinds. Same
  path handles the dynamic case where another player closes a door
  in front of you mid-walk.
- **Error code semantics distinguish terrain from doors:**
  - `PATH_BLOCKED` — no route exists. The map itself is in the
    way: a fence with no gate, a wall, deep water, an impassable
    overlay. The routine can't fix this by interacting with
    anything; it needs a different destination or a way around.
  - `DOOR_LOCKED` — an openable boundary on the path was contacted,
    walk_to attempted to open it, and after the retry budget the
    host still can't pass. The error's `.reason` includes the
    door coordinates AND any server message captured at the
    moment of the failed open (e.g. "This door appears to be
    locked", "You need a key to enter"), plus the classified
    refusal kind/precondition (`runtime/traverse.go::DoorLockedError`
    — locked vs toll vs level/quest requirement). Routines branch
    on the prose to decide between retrying, finding a key, or
    giving up:

    ```
    result = walk_to(x=87, y=552)   # known locked-door room
    if result.err.code == "DOOR_LOCKED" {
        if result.err.reason.contains("need a key") {
            return go_get_key()
        }
        if result.err.reason.contains("members only") {
            note("members-only zone, abandoning")
            return "no_access"
        }
    }
    ```
- Plain walls and fences (not openable per the boundary def) are
  never tried; pathfind treats them as impassable and routes
  around them. If no route around exists, `PATH_BLOCKED` fires.
- Set `attempt_open_doors=false` for strict semantics — useful
  for scouts that should report locked doors instead of barging
  through, or for routines that want to reason about door state
  explicitly:

  ```
  result = walk_to(x=128, y=664, attempt_open_doors=false)
  if result.err.code == "PATH_BLOCKED" {
      note("door blocks the path — reporting and giving up")
      return "blocked"
  }
  ```

```
go_to(target)               # coords or a known TOWN/landmark name
go_to(x, y)                 # positional coords form
```
- error_codes: `PATH_BLOCKED`, `DOOR_LOCKED`, `NO_SUCH_ITEM`, `INTERRUPTED`
- blocking: yes (returns when arrived or fails)
- **Travels anywhere in the world** — across regions, beyond the
  local pathfinder window — by stepping reachable waypoints toward
  the goal (opening gated doors en route) and replanning each hop.
- `target` may be coords (`x, y` or a position) or a known **town /
  landmark name** (`"Lumbridge"`, `"Varrock"`), resolved via
  `Gazetteer.PlaceByName`. **It does NOT take a POI type** — a
  string like `"bank"` or `"mining-site"` fails with `NO_SUCH_ITEM`
  and a steering message. That auto-routing masked ignorance and
  walked hosts into gates they couldn't pay; the cognition-first
  path is `search_map("bank")` to SEE the reachable destinations,
  then `go_to` the coords you choose (see Perception below).
- **Greedy**: a real obstacle it must go *around* (a river, a maze
  dead-end) stalls it and returns an error — the host must then
  reason a detour. Contrast with `walk_to`, which pathfinds within
  the local region only.

```
walk_path(corners)          # pre-planned multi-corner walk, list of [x, y] (max 25)
follow(player)              # server-side follow (opcode 165); player view or name
open_boundary(boundary)     # default click on a door/gate/fence/web view
```
There is no `unfollow()` verb — follow ends when the host walks
elsewhere.

### Combat

```
attack(target)              # npc-view or player-view (§9 flat alias of combat.attack)
attack_ranged(target)       # fire from the CURRENT tile, no melee pre-walk (safespot ranging)
combat.set_style(style)     # "controlled" | "aggressive" | "accurate" | "defensive"
combat.retreat()            # break melee by walking one tile away
combat.retreat(wait_rounds=false)
combat.retreat_to(x, y)     # flee to a specific safe tile once allowed
```
- attack error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`, `TARGET_DEAD`, `INTERRUPTED`
- retreat error_codes: `RETREAT_TOO_EARLY` — RSC forbids fleeing
  until the opponent has made 3 hits ("first 3 rounds of combat").
  By default the verb waits the rounds out when it can detect them;
  with `wait_rounds=false` it attempts immediately and hands back
  the typed rejection for poll-and-branch.
- Retreating IS the disengage mechanic in v235 (fleeing is a
  WALK_TO_POINT; the server breaks combat on it) — there is no
  separate `disengage()` verb.

### NPC dialog

```
talk_to(npc)                # Npc view, Int index, OR name string
converse(npc)               # LISTEN: aggregate speech, stop at real choices
answer(option_index)        # 1-based index from current dialog
find_option(needle)         # 1-based index of first option containing needle, 0 if none
wait_for_dialog()           # block until a menu opens (default 5s timeout)
wait_for_dialog(timeout_seconds)
pickpocket(npc)             # fire the NPC's primary command (command1); one attempt per call
```
- talk_to error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`
- converse error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`, `SERVER_REJECTED`
- answer error_codes: `DIALOG_NOT_OPEN`, `NO_SUCH_ITEM` (index < 1 —
  did `find_option` find no match?)
- **`talk_to` arg may be an Npc view, an Int server index, OR a
  name string** — `talk_to("banker")` auto-targets the nearest
  visible NPC of that name (no find/nearest boilerplate). Same for
  `converse` and `pickpocket`.
- **`converse` LISTENS — it does not pick for you.** It opens the
  dialog, aggregates everything the NPC says (speech-aware: it keeps
  waiting while speech bubbles are still streaming), and advances
  only the choices code can resolve — a lone "continue" prompt, an
  all-exit menu, a banker's bank-access option
  (`runtime/actions_dialog.go::resolveKnownDialogChoice`). At any
  REAL multi-option choice it STOPS with the menu open so YOU read
  what was said and decide. There is no topic/pick argument — NPCs
  only speak their pre-authored lines.
  - `.val` returns `{ said: [lines], options: [menu]|null,
    ended: bool, answered: int }`. `options != null` ⇒ a real
    choice is waiting; `ended == true` ⇒ the conversation finished.
  - Pattern: `r = converse(npc); if r.val.options != null {
    answer(find_option("Yes")); converse(npc) }`.
  - Fails (`SERVER_REJECTED`) if the NPC is busy with another
    player.

### Items

```
drop(item)                  # item-view or slot=N
pick_up(ground_item)        # ground-item-view
eat(item)                   # Eat/Drink/Bury — server decides by item def
equip(item)                 # item-view or slot=N; server enforces levels/slots
unequip(item)
use(item)                   # one arg: the item's own inventory command — use("sleeping bag")
use(item, target)           # use item on a target VIEW (boundary/item/scenery/ground_item/npc/player)
use(item, x=X, y=Y)         # use item on the scenery at a tile — cook/smith on scenery
use_inventory_default(item) # option-1 click: Bury bones, Clean herb, Empty bucket...
```
- drop error_codes: `INVENTORY_EMPTY`, `NO_SUCH_ITEM`
- pick_up error_codes: `OUT_OF_RANGE`, `INVENTORY_FULL`
- eat error_codes: `NO_SUCH_ITEM`, `EAT_IN_COMBAT` — RSC rejects
  item actions mid-fight ("You can't do that whilst you are
  fighting"); the typed code lets a routine retreat-then-eat
  instead of seeing a silent no-op
- pick_up `.val` returns the picked-up item-view on success
- `use`'s target kind picks the wire opcode: boundary (key on
  door), item (chisel on gem), scenery (raw food on range)

### Banking

The frozen surface is namespaced (api.md §10), dispatched through
the `bank` view:

```
bank.open(banker)           # walks adjacent + talks + opens
bank.deposit(item, amount)
bank.withdraw(item, amount)
bank.close()
bank.deposit_all()          # bulk verbs (#117/#118)
bank.deposit_all(keep)
bank.withdraw_all(item)
bank.withdraw_x(item, amount)
```
- error_codes: `BANK_NOT_OPEN`, `NO_SUCH_ITEM`, `INVENTORY_FULL`

### Trade

Namespaced `trade.*` (api.md §10) — `request` absorbed the old
`trade_request`+`accept_trade` (same opcode); `accept` is the
offer-screen accept, `confirm` the confirm-screen accept:

```
trade.request(player)
trade.offer(items)          # list of [item_id, amount] pairs (raw Int ids — TODO.md DSL-10)
trade.accept()              # screen 1
trade.confirm()             # screen 2
trade.decline()
```
- error_codes: `TRADE_NOT_ACTIVE`, `OUT_OF_RANGE`
- Handler bodies: `runtime/actions_trade.go`; verb table:
  `runtime/dsl_actions.go::tradeVerbs`.

### Duel

```
duel.request(player)
duel.set_rules(rules)       # 4 bools [retreat, magic, prayer, weapons] (list or named)
duel.stake(items)           # list of [item_id, amount] pairs
duel.accept()               # screen 1
duel.confirm()              # screen 2
duel.decline()
```
- Handler bodies: `runtime/actions_duel.go`.

### Shop

View-dispatched on the `shop` root (also reachable as
`world.shop.*` — `runtime/views_shop.go`):

```
shop.buy(item, qty)
shop.sell(item, qty)
shop.close()
```
- error_codes: `SHOP_NOT_OPEN`, `NO_SUCH_ITEM`
- Reads live beside the verbs: `shop.is_open`, `shop.is_general`,
  `shop.slots`, `shop.stock(item)`, `shop.price(item)`.
- There is no `shop.open(shopkeeper)` yet — open via
  `talk_to` + the shop dialog option (TODO.md `DSL-7`).

### Magic & prayer

```
cast(spell)                 # self-targeted (§9 flat alias of magic.cast)
cast(spell, target)         # NPC/Player/Position/item target selects the opcode
magic.cast(spell, target?)  # same handler, namespaced form
prayer.activate(prayer)
prayer.deactivate(prayer)
```

### Social

```
say(message)                # public chat
whisper(to, message)        # private message
add_friend(name)            # required before PMs can be sent or received
```
There is no `remove_friend` verb yet.

### Session & admin

```
logout()
command(cmd)                # admin command WITHOUT the leading "::" — command("tele 103 532")
```
A validator-time admin gate for `command()` is open work
(TODO.md `DSL-16`).

### Skills (declared, not implemented)

```
mine(rock)      fish(spot)      chop(tree)      cook(item, fire)
```

Registered as stubs returning `NOT_IMPLEMENTED`
(`spec.NotYetImplemented` → `runtime/dsl_actions.go::makeStub`).
The validator still accepts them at parse time — rejecting NYI
verbs is TODO.md `DSL-6`. **The working skilling path today is the
generic verbs**: `interact_at(x=, y=, option=)` fires the def's
command ("Chop" on a tree, "Mine" on a rock, "Net"/"Fish" on a
fishing spot), and `use("raw rat meat", x=216, y=731)` cooks on
scenery. `scan_for("rock")` enumerates the targets.

### Spatial utilities & bounds constructors (non-action)

```
distance_to(target)         # Chebyshev tiles from self.position to any .x/.y view
distance_to_xy(x, y)
nearest_npc()               # closest NPC of any type (vs world.npcs.find = first roster match)
nearest_npc(n => n.type_id == 4)
in_region(x1, y1, x2, y2)   # true iff self.position inside the rectangle
interact_at(position | x=, y=, option=)   # primary/secondary click on a scenery tile
box(x1, y1, x2, y2)         # bounds constructors for `bounds <shape> { ... }`
circle(cx, cy, radius)
near(radius)                # centered on self.position at routine start
```

### Primitives (non-action)

```
wait(seconds)               # number; the `wait 2.8..4.5` STATEMENT form takes a range
wait_until(_ => predicate)  # blocks until the lambda evaluates truthy
wait_until(_ => predicate, timeout_seconds)
note(text)                  # journal write (lightweight, no LLM)
```

- `wait` is chunked into 200ms slices with the pending-event queue
  drained between slices (`dsl/interp/interp.go::runWait`) — `on`
  handlers fire DURING a long wait, not only at its edges. The
  range form (`wait 2.8..4.5`) is the **statement** syntax; it
  draws jitter from the interpreter's seeded `Rand`, so it is
  deterministic given the seed. The call form takes a number.
- `wait_until` takes a single-arg **lambda** (the arg is ignored:
  `wait_until(_ => self.hp > 1, 5)`) because the DSL evaluates call
  arguments eagerly — a bare expression would be computed once.
  Polls at 200ms; returns `true` on satisfied, `false` on timeout.
- `note` publishes an `event.RoutineNote` on the host bus so the
  cradle UI streams it as a live in-character feed.

### Perception (non-action)

Map- and scene-reading primitives. Pure reads — no packets, no
typed failure (except where noted). They translate raw world state
into brain-ready prose and bearings (names, not ids).

```
look_around()               # default radius 10 tiles
look_around(radius)
where_am_i()
where_is(name)
bearing_to(x, y)            # two ints
bearing_to(position)        # position-like value
```

- `look_around([radius])` — returns a brain-ready multi-line TEXT
  summary of the scene: where you are (nearest area + POIs), self
  vitals, nearby NPCs (name + combat level + threat), players
  (combat level + threat + worn gear), ground items, notable
  scenery — all names, not ids. Default radius 10 tiles.
- `where_am_i()` — readable summary of where the host is in the
  world: nearest named area + notable POIs (bank / altar /
  furnace / ...) with bearing + distance. Map perception, not raw
  coords.
- `where_is(name)` — locate a named place (`"Lumbridge"`) or a POI
  type (`"bank"`, `"altar"`, `"furnace"`, `"fishing-point"`,
  `"mining-site"`) relative to the host: distance + bearing +
  coords. Backed by the world-map gazetteer.
- `bearing_to(x, y | position)` — 8-point compass direction
  (N/NE/E/.../NW) from the host to a target tile; `"here"` if
  coincident.

**Map perception** (`runtime/actions_worldmap.go`, backed by the
shared WorldOracle) — the cognition-first way to CHOOSE a
destination. The oracle INFORMS (per-destination reach + the
binding gate, its requirement, what you have); the brain DECIDES.
Each call costs a few in-world "study" seconds:

```
search_map(type)            # ranked list of REAL destinations for a POI type:
                            #   {label, x, y, dist, reach, gate, needs, you_have, payable}
                            #   reach = "open" | "gated" (payable) | "blocked"
reachable(x, y)             # the same explanation for ONE specific tile
survey_map()                # short text overview: what's open vs gated vs blocked around you
```

`search_map` is the one perception verb with a typed failure: it
returns a `NO_SUCH_ITEM` result when no destinations of that type
exist (or no map data is loaded).

**Scene perception** (`runtime/actions_scenery.go`) — enumerate the
LOCAL scene (static facts + live world mirror) instead of
hardcoding tiles. Cheap glance, no study cost:

```
scan_for(type)              # nearest-first list of scenery: {x, y, name, kind, def_id, position}
scan_for(type, radius)      # e.g. for r in scan_for("rock") { interact_at(x=r.x, y=r.y) }
```

Returns an empty list (branch on `.length == 0`) when none are
nearby — never a failure. Depleted objects drop out; a freshly-lit
fire shows up.

**Recognition** (`runtime/actions_resolve.go`, api.md §5) — fenced,
non-GUI primitives routed through the host's recognition faculty
(learned-alias store → conservative fuzzy match → brain fallback);
definitions/ids come from facts, never from the LLM:

```
resolve(text)               # ranked List<Match{def, kind, score}>, [] if none
resolve(text, kind)         # kind filter: "item"/"npc"/"loc"/"spell"/"prayer"
resolve_one(text, kind?)    # best single Match or Null
```

### Stdlib — LLM + memory (touch brain or memory)

These reach the cognition layer. The seams are `Host.Strategist`
(`brain.Strategist`) and `Host.Retriever` (`cognition.Client`) —
default `brain.StubStrategist` / `cognition.StubClient` return
deterministic canned answers; the real implementations are the mesa
RPC adapters (`mesa/client/adapters.go::AsStrategist/AsRetriever`),
wired when a mesa client is configured
(`runtime/runhost.go:105`). Cost discipline is structural,
not a per-call cap: `decide()` consults the host's pearl policy
first (`Pearl.TryDecide` — a hit answers locally with NO LLM call;
a miss can still persona-bias the option order) and memoizes
pearl-miss verdicts in a TTL decision cache
(`runtime/decision_cache.go`). The routine-level budgets (op count,
wall clock — `dsl/interp/caps.go`) bound everything else.

```
contemplate_reality(question = "")   # open-ended decision → .val = short string code
evaluate(situation)                  # → .val = 0-1 float (strategist confidence)
decide(options, context = "")        # → .val = chosen option (pearl → cache → LLM)
recall(query, top = 5)               # → .val = list of strings; Corpus chunks (with
                                     #   provenance) when wired, else Retriever reflections
relation_with(name)                  # → .val = trust grade from the host's OWN ledger when
                                     #   the party is known; falls back to the Retriever
remember(key, value)                 # tiered memory write (scratch→local→mesa by namespace)
recollect(key)                       # tiered read; .val = value, or null on a miss
forget(key)                          # delete from every tier
```

All of the above are BUILT (`runtime/actions_flow.go`,
`runtime/actions_memory.go`). Still `NOT_IMPLEMENTED` stubs:
`exec(prompt)`, `improvise(prompt)`, `reflect_now()` (LLM-authored
DSL fragments / synchronous reflection), `wait_for_chat()`,
`observe(target, ticks)`, and the persona reads `mood()` /
`motivation()`.

### Reputation / scratch-cache / relational-tag verbs — NOT BUILT

An earlier revision of this doc spec'd a Phase-4 verb set here:
`trust(name)`, `trust_confidence(name)`, `reputation(name)`,
`is_rival`/`is_ally`/`is_stranger`, `cache.get/set/incr`,
`rel.tag`/`rel.has_tag`. **None of these exist** — zero spec rows.
What shipped instead: the trust ledger itself (read today only
through `relation_with`), and the tiered-memory verbs
`remember`/`recollect`/`forget`, which cover the scratch-cache role.
The reputation read surface is tracked as TODO.md `DSL-24` (paired
with the trust-ledger remainder `C-24`).

## Argument styles

Most actions accept arguments in two forms:

```
walk_to(x=120, y=504)              # named (preferred for clarity)
walk_to(120, 504)                  # positional
walk_to(spot)                      # single arg with .x/.y fields
```

Single-entity forms work for any action that takes a position —
the runtime extracts `.x` and `.y` from the argument if it's a
Getter (`runtime/dsl_helpers.go::resolvePoint`). So
`walk_to(self.position)` and `walk_to(target)` both work without
explicit field access.

## Yielding semantics

Every bridge-registered callable is bound as a yielding
`actionCallable` (`runtime/dsl_actions.go:245`). Before (and after)
a yielding call, the interpreter drains the pending event queue and
dispatches any matching handlers (`dsl/interp/interp.go::evalCall`).
This is what makes `on` and `when` handlers fire "between actions,
never mid-action" — with the one deliberate exception that `wait`
(statement or call) pumps events every 200ms DURING the sleep, so a
reaction never waits out a long sleep.

## Return values

Action `result.val` is `null` for most actions (the action either
succeeded or it didn't — no useful return). Actions that produce
data return the data on `.val`:

- `pick_up()` → picked-up item-view
- `converse()` → `{said, options, ended, answered}` map
- `contemplate_reality()` / `decide()` → chosen string
- `evaluate()` → 0–1 float
- `recall()` → list of strings
- `relation_with()` → trust-grade / relation string
- `search_map()` / `reachable()` → destination map(s)
- `scan_for()` → list of scenery entries

Reading `.val` on a successful action that returns null gives
you `null`. No surprises.

## Naming consistency rules

See [`syntax.md`](syntax.md) for the full naming convention, but
specifically for actions:

- Snake_case verbs in present tense — `walk_to`, `pick_up`, not
  `walked` or `picking_up`
- Subsystem verbs are namespaced — `bank.deposit`, `trade.request`,
  `combat.set_style`; the only sanctioned flat aliases are `attack`
  and `cast` (api.md §9/§10)
- Bang variant for every bang-eligible callable — auto-generated
- The canonical name table is `dsl/spec/actions.go`; new actions
  add a spec row + an `actionHandlers` entry, and the consistency
  test enforces the pairing

---

Open action-layer work (selector reachability defaults, structured
outcomes / `do_until`, `examine`, shop.open, NYI-verb rejection,
reputation reads, …) lives in [`docs/TODO.md`](../TODO.md) — see
`DSL-1`, `DSL-6`, `DSL-7`, `DSL-10`, `DSL-11`, `DSL-16`, `DSL-21`,
`DSL-24`.
