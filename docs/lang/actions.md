# Actions — The THEN THAT Layer

> **Host** — an autonomous AI actor. Actions are the verbs a host
> can use to change the world.

## Status

- **Implemented today**: most primary actions are wired through
  `runtime/dsl_actions.go` — walk_to, attack, talk_to, answer,
  drop, pick_up, eat, deposit, withdraw, close_bank, open_bank,
  say, whisper, logout, plus primitives wait and note. Stubs for
  mine/fish/chop/cook/cast (no Host method yet) and for the LLM
  stdlib.
- **Today's error model**: stringly-typed. Actions return `Null`
  on success or `String("walk_failed: ...")` on failure. Ugly.
- **Planned**: typed `Result`/`Error` with Go-side enum mapped to
  SCREAMING_SNAKE codes; bang variants for assert-success.

## The error model

### Go side

Every action's failure mode maps to a typed `ErrorCode` enum:

```go
// dsl/errors.go (planned)
type ErrorCode int

const (
    PATH_BLOCKED ErrorCode = iota
    OUT_OF_RANGE
    INVENTORY_FULL
    INVENTORY_EMPTY
    NO_SUCH_ITEM
    TARGET_DEAD
    TARGET_OUT_OF_VIEW
    INTERRUPTED
    NOT_LOGGED_IN
    BANK_NOT_OPEN
    TRADE_NOT_ACTIVE
    DIALOG_NOT_OPEN
    SERVER_REJECTED
    ACTION_TIMEOUT
)

func (e ErrorCode) String() string {
    return [...]string{
        "PATH_BLOCKED", "OUT_OF_RANGE", "INVENTORY_FULL", ...
    }[e]
}
```

Each action wrapper maps the underlying Go error onto one of
these codes plus a human-readable `reason` string.

### DSL side

Actions return a `Result` value with two fields:

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
    fish!(spot)
    eat!("lobster")
} recover err {
    note(f"chain failed: {err.code}")
    if err.code == "PATH_BLOCKED" { try_alternative_route() }
    else { abort err.code }
}
```

The bang isn't a syntactic decorator — it's a separate callable.
`eat` and `eat!` are registered as two different names in
`Interpreter.Builtins`. The bridge auto-generates the bang
variant for every action. Validator knows both names; tooling
(REPL `.help`, doc lookup) treats them as a pair.

### Which callables get bang variants

The rule is uniform: **anything that returns a `Result` gets a
bang variant.** That's:

- **All primary actions** — walk_to, attack, eat, drop, pick_up,
  talk_to, answer, deposit, withdraw, open_bank, close_bank,
  say, whisper, logout, follow, set_combat_style, the future
  skill verbs (mine, fish, chop, cook, cast), and every admin
  action (admin_set_stat, admin_give_item, etc.)
- **All LLM stdlib calls** — contemplate_reality, decide,
  evaluate, exec, improvise, reflect_now. These can fail
  (rate-limit exceeded, model error, exhausted budget) and
  return typed `Result` values.
- **All memory stdlib calls** — recall, relation_with. They
  hit mesa and can fail (network, schema mismatch).

What does **not** get a bang variant:

- **`wait` / `wait_until`** — cancellation flows through
  ctx-cancel, not as a `Result.err`. They can't fail in the
  typed sense.
- **`note`** — local logger write. Doesn't fail.
- **`mood` / `motivation`** — pure persona reads. Don't fail.
- **Procs** — user-defined helpers. They return whatever they
  return; bang wouldn't make sense.

The validator knows the bang-or-not status of every registered
callable and rejects `note!()` etc. at parse time.

## The action menu

Categorized list of every action the language supports. Field
types are arg expectations; `error_codes` is the set of failure
modes the action might emit.

### Movement

```
walk_to(x = int, y = int)
walk_to(position)
walk_to(x = int, y = int, attempt_open_doors = bool)   # default: true
```
- error_codes: `PATH_BLOCKED`, `DOOR_LOCKED`, `OUT_OF_RANGE`, `INTERRUPTED`
- blocking: yes (returns when arrived or fails)
- **Auto-opens closed doors by default** (`attempt_open_doors=true`).
  When the walk stalls adjacent to an openable boundary (door /
  doorframe / gate that has an Open action), walk_to interacts
  with it and re-pathfinds. Same path handles the dynamic case
  where another player closes a door in front of you mid-walk.
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
    locked", "You need a key to enter"). Routines branch on the
    prose to decide between retrying, finding a key, or giving up:

    ```
    result = walk_to(x=87, y=552)   # known locked-door room
    if result.err.code == "DOOR_LOCKED" {
        if contains(result.err.reason, "need a key") {
            return go_get_key()
        }
        if contains(result.err.reason, "members only") {
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
  if result.err.code == PATH_BLOCKED {
      note("door blocks the path — reporting and giving up")
      return "blocked"
  }
  ```

```
go_to(target)               # coords, named place, or POI type
go_to(target, y)            # positional x, y form
```
- error_codes: `PATH_BLOCKED`, `OUT_OF_RANGE`, `INTERRUPTED`
- blocking: yes (returns when arrived or fails)
- **Travels anywhere in the world** — across regions, beyond the
  local pathfinder window — by stepping reachable waypoints toward
  the goal (opening gated doors en route) and replanning each hop.
- `target` may be: coords (`x, y` or a position), a named place
  (`"Lumbridge"`, `"Varrock"`), or a POI type (`"bank"`,
  `"furnace"`, `"fishing-point"`) resolved to the **nearest** via
  the gazetteer.
- **Greedy**: a real obstacle it must go *around* (a river, a maze
  dead-end) stalls it and returns an error — the host must then
  reason a detour. Contrast with `walk_to`, which pathfinds within
  the local region only.

```
follow(player)
unfollow()
```
- error_codes: `OUT_OF_RANGE`, `INTERRUPTED`

### Combat

```
attack(target)              # npc-view or player-view
disengage()                 # break combat manually
```
- attack error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`, `TARGET_DEAD`, `INTERRUPTED`

```
set_combat_style(style)     # "accurate" | "aggressive" | "defensive" | "controlled"
```

### NPC dialog

```
talk_to(npc)                # Npc view, Int index, OR name string
converse(npc)               # talk + auto-answer the whole dialog
converse(npc, pick)         # prefer options containing `pick`
answer(option_index)        # 1-based index from current dialog
```
- talk_to error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`
- converse error_codes: `OUT_OF_RANGE`, `TARGET_OUT_OF_VIEW`, `SERVER_REJECTED`
- answer error_codes: `DIALOG_NOT_OPEN`
- **`talk_to` arg may be an Npc view, an Int server index, OR a
  name string** — `talk_to("banker")` auto-targets the nearest
  visible NPC of that name (no find/nearest boilerplate).
- **`converse` drives the NPC's *whole* dialogue to completion**:
  it auto-answers every menu (preferring an option whose text
  contains the optional `pick` substring, else the first) until no
  more menus appear. Bakes in the talk→answer→repeat pattern for
  NPC interaction (tutorial guides, quests).
  - `npc` may be an Npc view, an Int server index, OR a name
    string (auto-targets the nearest visible NPC of that name).
  - `.val` returns the number of menus answered.
  - Fails (`SERVER_REJECTED`) if the NPC is busy with another
    player. Read `world.last_dialog_text` / `world.messages` for
    what was said.

### Items

```
drop(item)                  # item-view or slot index
pick_up(ground_item)        # ground-item-view
eat(item)                   # any consumable
use(item, target)           # use item on target — covers many skills
```
- drop error_codes: `INVENTORY_EMPTY`, `NO_SUCH_ITEM`
- pick_up error_codes: `OUT_OF_RANGE`, `INVENTORY_FULL`
- eat error_codes: `NO_SUCH_ITEM`
- pick_up.val returns the picked-up item-view on success

### Banking

```
open_bank(banker)           # walks adjacent + talks + opens
deposit(item, amount)
withdraw(item, amount)
close_bank()
```
- error_codes: `BANK_NOT_OPEN`, `NO_SUCH_ITEM`, `INVENTORY_FULL`

### Trade

```
trade_request(player)
trade_accept()
trade_decline()
trade_offer(items)          # list of item-views
trade_confirm()
```
- error_codes: `TRADE_NOT_ACTIVE`, `OUT_OF_RANGE`

### Social

```
say(message)                # public chat
whisper(target, message)    # private message
add_friend(name)
remove_friend(name)
```

### Skills (planned — Host method not yet implemented)

```
mine(rock)
fish(spot)
chop(tree)
cook(item, fire)
cast(spell, target)
```

Currently registered as stubs returning the error code
`NOT_IMPLEMENTED`. Real implementations land as part of the
18-skill integration (task #41 in the project task list).

### System

```
logout()
admin_command(cmd)          # admin-only: "::heal", etc.
```

### Primitives (non-action, non-yielding)

```
wait(seconds)               # accepts int, float, or range like 2.8..4.5
wait_until(predicate)       # blocks until expression becomes truthy
note(text)                  # journal write (lightweight, no LLM)
```

`wait` and `wait_until` use the interpreter's seeded `Rand` for
range jitter — deterministic given the seed.

`wait_until` validates that the predicate is subscription-safe
(same rule as `when` expressions — no actions, no LLM calls).

### Perception (non-action, non-yielding)

Map- and scene-reading primitives. Pure reads — no packets, no
yield, no typed failure. They translate raw world state into
brain-ready prose and bearings (names, not ids).

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

### Stdlib (touch brain or memory — expensive)

These cost LLM dollars and have per-routine call caps. See
[`thought-architecture.md`](thought-architecture.md) for the
broader cognitive context.

```
contemplate_reality(question = "")   # max 5/routine, ~$0.005, Sonnet
evaluate(situation)                   # max 10/routine, ~$0.001, Haiku
decide(options, context = "")         # max 10/routine, ~$0.002, Haiku
exec(prompt)                          # max 1/routine, ~$0.01
improvise(prompt)                     # max 2/routine, ~$0.01
recall(query, top = 5)                # cheap, no LLM (vector search in mesa)
relation_with(name)                   # cheap, no LLM
reflect_now()                         # max 1/routine, ~$0.005
mood()                                # cheap, returns map
motivation()                          # cheap, returns map
```

Stdlib calls today are stubs returning the error code
`NOT_IMPLEMENTED`. Real LLM bridge lives in delos (Phase 3).

### Stdlib — Planned (Phase 4 design)

These are the **planned** stdlib additions from the cognition/social
design work — none are built yet. The cognition, social-graph, and
scratch-cache backends don't exist, so these are spec'd here for
the verb surface, not implemented. The one exception is
`relation_with(name)`, which **already exists today as a STUB**
(returns the relation record; see the action menu / Stdlib list above).

#### Reputation queries (pure, fast, local-copy reads)

A small set of **pure, fast, watchable** reputation reads — no LLM,
no network on the hot path. They read the host's **local trust-ledger
copy** (the per-host hot copy of relationship Edges, hydrated on
spawn/encounter, write-through-mirrored to mesa — see `mesa.md`
relationships + `_research/social-graph-and-trust-ledger.md`). Because
they're pure and local-fast, they're usable inside `when` expressions
and handler bodies. No bang variants (local reads don't fail in the
typed sense).

| Call | Returns | Notes |
|---|---|---|
| `trust(name)` | `float 0..1` | The Beta posterior mean — "how reliable do I think they are" |
| `trust_confidence(name)` | `float 0..1` | Evidence strength — "how SURE am I" (200 trades vs met once) |
| `reputation(name)` | `band` | The DERIVED band: `stranger` \| `acquaintance` \| `friend` \| `rival` \| `enemy` |
| `is_rival(name)` | `bool` | Derived from tag/band |
| `is_ally(name)` | `bool` | Derived from tag/band |
| `is_stranger(name)` | `bool` | True when confidence is low |
| `relation_with(name)` | relation record or null | **EXISTS TODAY AS A STUB** — returns the relation record |

#### Scratch cache (per-host key→value working store)

A small per-host **scratch cache** for rate-limit / dedup work ("have I
asked this guy?"). **Local-fast** read/write on the hot path, with an
**async write-through to mesa `working_scratch`** so cognition can fold
relevant contents into LLM-call prompts (`PrepareDecision`). Distinct
from mesa **episodic memory** (this is fast local scratch, not durable
recall) and distinct from `note(text)` (which is a journal write). See
`_research/chat-interruption-and-engagement.md` §5.3.

| Call | Returns | Notes |
|---|---|---|
| `cache.get(key)` | value or null | Local-fast read, never touches the network on the hot path |
| `cache.set(key, value)` | null | Local write + async write-through to mesa `working_scratch` |
| `cache.incr(key)` | `int` | Atomic increment — the per-player ask counter idiom |

#### Relational tags (structured per-relationship facts)

Structured per-relationship facts the host learns (e.g. `"no-steel"` —
"don't re-ask this person for steel bars"). Backed by mesa
`relationships.tags` (see `mesa.md`; the schema has only freetext
`notes` today, so the structured tag set is part of the same Phase-4
social work).

| Call | Returns | Notes |
|---|---|---|
| `rel.tag(name, tag)` | null | Set a structured tag on the relationship |
| `rel.has_tag(name, tag)` | `bool` | Pure, fast, local — watchable in `when`/handlers |

## Argument styles

Most actions accept arguments in two forms:

```
walk_to(x=120, y=504)              # named (preferred for clarity)
walk_to(120, 504)                  # positional
walk_to(spot)                      # single arg with .x/.y fields
```

Single-entity forms work for any action that takes a position —
the runtime extracts `.x` and `.y` from the argument if it's a
Getter. So `walk_to(self.position)` and `walk_to(target)` both
work without explicit field access.

## Yielding semantics

Every primary action **yields** before it runs. That means the
interpreter drains the pending event queue, dispatches any
matching handlers, THEN executes the action. This is what makes
`on` and `when` handlers fire "between actions, never
mid-action."

`wait` also yields. `note` does not (it's pure / instant).

## Return values

Action `Result.val` is `null` for most actions (the action either
succeeded or it didn't — no useful return). Actions that produce
data return the data on `.val`:

- `pick_up()` → picked-up item-view
- `contemplate_reality()` → strategist's chosen action label (string)
- `evaluate()` → 0–1 float
- `decide()` → chosen option (string)
- `recall()` → list of episode records
- `relation_with()` → relational record or null

Reading `.val` on a successful action that returns null gives
you `null`. No surprises.

## Naming consistency rules

See [`syntax.md`](syntax.md) for the full naming convention, but
specifically for actions:

- Snake_case verbs in present tense — `walk_to`, `pick_up`, not
  `walked` or `picking_up`
- Verb-noun pairs where ambiguous — `set_combat_style`, not
  `set_style`
- Bang variant for every action — auto-generated
- Action names match the dsl.md "Built-in actions" table; new
  actions added there too as we wire them

## Open questions for tomorrow

- **Should non-bang actions ever raise?** I think no — non-bang
  is the "let me handle it" path, always returns a Result.
  Validator/runtime errors (type mismatch, etc.) can still raise.
- **Implicit Result unwrap in conditionals?** Idea:
  `if walk_to(...) { ... }` is true if no error. Maybe — but
  `walk_to(...).err` is one extra token and more explicit.
  Lean toward explicit.
- **Action timeouts** — should every action have a default
  timeout? Or rely on the wall-clock budget? I'd put it in the
  budget for now and add per-call timeouts later if needed.
- **Lambda args** — `world.npcs.filter(n => n.combat_level <= 30)`.
  dsl.md punts on this. We don't need it for v1; for now,
  filter/map use proc references: `filter(weak_enough)` where
  `proc weak_enough(n) { return n.combat_level <= 30 }`.

## Implementation order for the new error model

1. Add `dsl/errors.go` with `ErrorCode` enum + `String()` method
2. Add `interp.Error` type (Code, Reason, Fatal) + `interp.Result`
   (Val, Err)
3. Update `runtime/dsl_actions.go` to construct typed errors
   instead of strings; map each Go error type to a code
4. Add bang-variant auto-registration in `dsl_bridge.go` — for
   every registered action, register `<name>!` that panics with
   abortSignal on failure
5. Update validator to know both `<name>` and `<name>!` as
   builtins; bang is action context, same constraints
6. Migrate existing routines + tests to the new shape

Estimate: 1–2 days, mostly mechanical. Most of the work is the
ErrorCode-to-string switch and the routine migration.
