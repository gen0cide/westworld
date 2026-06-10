# Events — The WHEN THIS Subscription Layer

> **STATUS: CURRENT** (verified against code 2026-06-10, HEAD `0bfa818`).
> The reactive vocabulary — `on` / `when` / `select` / `defer` /
> `try`/`recover` / file-level `extends` — is **BUILT** and executed
> (`dsl/interp/`). The one exception is the per-handler
> `on event() extends host` / `super()` chain, still a validate-time
> error (TODO.md **DSL-13**). Sections marked PLANNED / DESIGN below
> are the live spec for that unbuilt work, not descriptions of code.

> **Host** — an autonomous AI actor. This doc covers everything
> a host's routine does *reactively*: handlers, watchers, blocking
> waits, cleanup, and overrides.

## Status

> **Verified against the interpreter** (`dsl/interp/`,
> `runtime/dsl_bridge.go`). The reactive control-flow vocabulary
> below is **executed**, not just parsed — `when` / `select` /
> `defer` / `try`/`recover` and file-level `extends` all run today.
> The one exception is the per-handler
> `on event() extends host` / `super()` override chain, which is
> a validate-time error if used. The persona schema it was parked
> on now exists (`persona/`), but no default-handler registry is
> wired — TODO.md **DSL-13** tracks it as unblocked.

### Implementation status

| Construct | Status | Where it runs |
|---|---|---|
| file-level `on event(args) { ... }` | **EXECUTED** | `interp/events.go` `dispatchPendingEvents`/`runHandler` |
| routine-level `on` handlers (inside `routine { ... }`) | **EXECUTED** | `interp/events.go` `registerHandlers` indexes `routine.Handlers` alongside file-level |
| `when <expr> [becomes/changes] { ... }` watchers | **EXECUTED** | `interp/watchers.go` `execWhen`/`evalWatchersOnce` |
| `select { when… / on… / timeout… }` | **EXECUTED** | `interp/watchers.go` `execSelect` |
| `defer <call>` | **EXECUTED** | `interp/interp.go` `pushDefer`/`drainDeferred` |
| `try { … } recover err { … }` | **EXECUTED** | `interp/interp.go` `execTry` |
| `bounds { on ... }` location-scoped handlers | **EXECUTED** | `interp/events.go` `registerBoundsDecl`/`registerBoundedHandlers`; dispatch filters on the event's `(x, y)` args via `eventLocation` |
| file-level `extends "parent.routine"` (proc/handler/bounds merge) | **EXECUTED, disk-load only** | `runtime/dsl_bridge.go` `mergeExtends` (via `ParseRoutineFile`); `ParseRoutineString` (REPL / `exec()`) rejects `extends` outright |
| per-handler `on event() extends host` + `super()` | **PARSED-NOT-EXECUTED** | needs the persona default-handler registry; `super` is a validate error today (`dsl/validator/validator.go`; TODO.md DSL-13) |

Subscription qualifiers (`becomes true`/`becomes false`/`changes`)
are executed. `increases` / `added` / `removed` and `by N+`
thresholds shown in the examples below are **not yet implemented** —
the parser/validator accept the bare qualifier set
(`WhenBecomesTrue`/`WhenBecomesFalse`/`WhenChanges`,
`dsl/ast/ast.go`); collection-delta and counter-delta qualifiers
are tracked as TODO.md **DSL-19**.

- **Bus events wired today** (`runtime/dsl_events.go`
  `translateEvents`/`translateEvent`): `chat_received`,
  `private_message`, `server_message`, `message`, `coords_changed`,
  `death`, `damage_taken` (self-only; `source` always empty — v235
  carries no attacker identity), `xp_gain`, `level_up`,
  `equipment_changed`, `player_equipment_changed`,
  `player_level_changed`, `item_gained`, `item_appeared`,
  `item_disappeared`, `target_died` + its `npc_killed` alias,
  `bank_opened` / `bank_slot_update` / `bank_closed`,
  `boundary_changed`, and the full trade/duel lifecycle
  (`trade_request`, `trade_opened`, `trade_other_offer`,
  `trade_other_accepted`, `trade_confirm_shown`, `trade_closed`,
  `duel_request_incoming`, `duel_opened`, `duel_other_offer`,
  `duel_settings_update`, `duel_other_accepted`,
  `duel_confirm_shown`, `duel_closed`). The canonical,
  machine-checked list lives in `dsl/spec/events.go` (**34
  entries**) — that file is the source of truth, not this prose.
  (Its `NotYetImplemented` flag on `damage_taken` is stale — the
  translator does emit it; TODO.md DSL-19 carries the flag flip.)
- **Removed from dsl.md**: `hp_below(threshold)`,
  `fatigue_above(threshold)` — replaced by `when` expressions.

## The five constructs

| Keyword | When it fires | Where it lives |
|---|---|---|
| `on` | Persistent bus events (transient, can't be polled) | File top-level |
| `when` | State transitions on query expressions | Inside a block (routine body, loops, etc.) |
| `select` | Block until ONE of several conditions fires | Inside the routine body |
| `defer` | On scope exit (return, abort, error, cancel) | Inside any block |
| `try`/`recover` | Catches bang aborts from inside a block | Inside the routine body |

Each is documented below.

## `on` — persistent bus handlers

For transient events that have no observable footprint after the
moment (chat messages, damage events, level-ups). These need a
real event bus because you can't poll for them — they happen and
they're gone.

```
on chat_received(speaker, message) {
    if speaker == "alex" and message == "stop" {
        abort "TOLD_TO_STOP"
    }
}

on private_message(speaker, message) {
    response = contemplate_reality(f"how do I respond to {speaker}?")
    whisper!(speaker, response)
}

on damage_taken(amount, source) {
    note(f"took {amount} damage from {source}")
}
```

Lifetime: registered at routine start, active until the routine
ends. Fire between primary actions (never mid-action — the
interpreter drains the event queue around yielding action calls,
and in 200ms slices during `wait`, so handlers fire *during* a
long wait rather than only at its edges — `dsl/interp/interp.go`
`runWait`).

Handler bodies can call any expression, action, or proc. They
have read-only access to the routine's locals.

Handlers **cannot** `return` from the routine (they'd be
returning from a different frame). They **can** `abort` to
terminate the parent routine — useful for "kill the routine on
this kind of message."

**Handlers cannot yield.** The validator rejects `wait`,
`wait_until`, and `select` inside handler bodies. Handlers are
reactive and short by design — a `wait 5` inside a handler
would block the routine's main loop for 5 seconds. If you need
to time-bound a reaction, do the bounded work in the routine
body and let the handler just trigger a state change. (Same
rule as `proc` bodies, which also can't yield.)

Handlers MAY use `defer` and `try`/`recover`. They're
non-yielding control-flow constructs and don't block the
routine's main loop.

### Category A — true bus events (need persistent `on`)

These have no underlying queryable state. They're transient.

| Event | Args | Source |
|---|---|---|
| `chat_received` | `speaker, message` | Public chat from another player |
| `private_message` | `speaker, message` | Incoming PM |
| `server_message` | `text` | System message from server |
| `damage_taken` | `amount, source` | We were hit (distinct from "my hp is X now") |
| `level_up` | `skill, new_level` | One of our OWN skills gained a base level |
| `equipment_changed` | `slot, item` | Our OWN worn/wielded set changed (one fire per changed slot) |
| `trade_request` | `from` | Another player initiated a trade |
| `coords_changed` | `x, y` | Our position updated |
| `death` | (none) | We just died |

That's the original v1 core, **not** the canonical list — the
canonical list is `dsl/spec/events.go` (34 entries), which also
carries the full trade/duel lifecycle (shipped 2026-05-28), the
bank window events, ground-item appear/disappear, `boundary_changed`,
`message`, `xp_gain`, and `target_died`/`npc_killed`. Anything that
*could* be polled as state is better expressed as `when` over the
query layer.

#### `level_up(skill, new_level)` — IMPLEMENTED

Fires when one of the host's **own** skills gains a base level.
`skill` is the lowercase skill name (attack, fishing, …);
`new_level` is the new base level. Filter in-body on `skill`
(e.g. `if skill == "fishing"`). Synthesized by diffing per-skill
base levels across stat updates — the **same edge pattern as
`xp_gain`** (see the change-event note below).

```
on level_up(skill, new_level) {
    if skill == "fishing" and new_level == 40 {
        note("can fish lobsters now")
    }
}
```

#### `equipment_changed(slot, item)` — IMPLEMENTED

Fires when the host's **own** worn/wielded set changes — **once per
changed slot**. `slot` is the human slot name
(helmet/body/legs/weapon/shield/gloves/boots/amulet/cape); `item` is
the new worn item in that slot (`item.is_empty` if the slot was
emptied). A bare `on equipment_changed()` still works — extra args
are ignored — when you only care that *something* changed. For the
full new state, read `self.equipped` / `self.equipped.bonuses`.
Synthesized by diffing the host's worn items across inventory
updates.

```
on equipment_changed(slot, item) {
    note(f"now wearing {item.name} in {slot}")
}

on equipment_changed() {
    # don't care which slot — just re-read the full set
    recompute_bonuses(self.equipped.bonuses)
}
```

### Category A′ — other-player change events (also `on`)

Edge-triggered events about a **visible other player** changing. Like
the self events above, these are synthesized by diffing the world
mirror and fire **only on an actual change** — never on the periodic
appearance re-send.

| Event | Args | Source |
|---|---|---|
| `player_equipment_changed` | `player, slot, item` | A visible player's worn equipment changed (one fire per changed slot) |
| `player_level_changed` | `player, new_level` | A visible player's combat level changed |

#### `player_equipment_changed(player, slot, item)`

Fires when a visible **other** player's worn equipment changes —
**once per changed slot**. `player` is the player view; `slot` is the
human slot name
(helmet/body/legs/weapon/shield/gloves/boots/amulet/cape); `item` is
the new worn item in that slot (`item.is_empty` if removed). Fires
only on an **actual change**, not the periodic appearance re-send.
Synthesized by diffing `PlayerRecord.EquipBySlot` on each appearance
update.

```
on player_equipment_changed(player, slot, item) {
    if slot == "weapon" {
        note(f"{player.name} just drew {item.name}")
    }
}
```

#### `player_level_changed(player, new_level)`

Fires when a visible **other** player's **combat** level changes.
`player` is the player view; `new_level` is their new combat level.

```
on player_level_changed(player, new_level) {
    note(f"{player.name} is now combat {new_level}")
}
```

#### Note — these are edge-triggered change events

`level_up`, `equipment_changed`, `player_equipment_changed`, and
`player_level_changed` are **edge-triggered**: they fire only on a
real change, synthesized in `runtime/frame.go` by diffing the world
mirror — the **same mechanism** as the existing `xp_gain` /
`item_gained` / `target_died` events (all published from
`runtime/frame.go`, translated in `runtime/dsl_events.go`). They are
not raw wire packets; they are deltas the runtime computes from
successive snapshots.

### Planned additions (Phase 4) — surface what's already decoded + the one new shape

> **PARTIALLY SHIPPED; remainder is DESIGN.** The canonical,
> machine-checked event set remains `dsl/spec/events.go` (see the
> Status note above) — that file is the source of truth, *not* this
> prose. The biggest bucket below has since shipped; what remains is
> tracked as TODO.md **DSL-19**.

An event-vocabulary audit found the catalog is **under-populated,
not over-populated**: roughly thirty bus events are already
*decoded into typed events/state* on the wire side, far more than
the nine-row "Category A" table above surfaces. Most gaps are
therefore a **spec-row + translate-case change** (surface what the
translator currently drops), not new wire work. Three buckets:

- **Surface the already-decoded events the translator drops —
  LARGELY SHIPPED (2026-05-28).** The fuller trade/duel *lifecycle*
  (the offer → accept → confirm → scam-watch machine beyond bare
  `trade_request`; the parallel duel stake negotiation) is **BUILT**:
  `trade_opened` / `trade_other_offer` / `trade_other_accepted` /
  `trade_confirm_shown` / `trade_closed` and `duel_request_incoming` /
  `duel_opened` / `duel_other_offer` / `duel_settings_update` /
  `duel_other_accepted` / `duel_confirm_shown` / `duel_closed`, all
  in `dsl/spec/events.go` + `runtime/dsl_events.go`. (`level_up` also
  shipped — see Category A above.) Still unsurfaced (decoded but not
  translated; TODO.md DSL-19): `npc_chat` / NPC dialogue,
  `dialog_opened`, `sleep_captcha`, `npc_appeared`,
  `shop_opened`/`shop_closed`, `item_lost`.
- **The ONE genuinely-new *shape*: `idle` / `tick` scheduler-source
  events.** Modeled on AutoRune's `OnIdle` (stuck-detection) and
  `OnTimer` (periodic beat). Unlike everything above, these are not
  a translation gap — they need a **new source** (a scheduler/timer
  feeding the bus), because nothing on the wire emits them. Highest-
  value gap for running unattended for hours; genuinely new
  mechanism, so still genuinely open work (TODO.md **C-13** holds the
  scheduler-home question; **DSL-12** the related `host.idle_ticks`
  accessor).

**What is NOT an `on` event — the reflex/deliberate split.**
`under_attack`, `followed`, and `can_eat` are **NOT** planned bus
events. They are `when`-predicates over the query layer, because the
**v235 damage packet carries no attacker identity** — "something is
hitting me / who" cannot be a clean push. This is the reflex ↔
deliberate cleavage: fast *reflex* events that have a real transient
footprint (`damage_taken` fires, but its `source` is permanently
empty) belong on `on`; *deliberate judgment* state ("am I under
attack, by a player, within catching range, can I heal this round")
is polled state and belongs on `when`. The reflex tier is therefore
largely a state-watcher tier — see the `when` section below and the
planned predicates `combat.under_attack`,
`combat.attacker_is_player`, `combat.being_followed`,
`combat.can_eat`, `self.poisoned` (none built; TODO.md DSL-19).

## `when <expr>` — state-transition watchers

For reactive behavior triggered by state changes (HP dropping
below a threshold, inventory becoming full, target dying). Reuses
the query expression syntax — the same predicate you'd use in an
`if`.

```
routine fight_goblin() {
    when self.hp < 40 {
        eat!(inventory.find("lobster"))
    }
    when inventory.is_full {
        bank_loot()
    }

    attack!(target)
    wait_until(target.hp == 0)
    return "killed"
}
```

### Block-scoping

`when` is **lexically scoped to its enclosing block**. The
watcher registers on block entry and unregisters on block exit
(any path — return, abort, error, normal fall-through). Different
parts of a routine can have completely different reactive shapes:

```
routine mixed_activity() {
    when self.hp < 30 { eat!(inventory.find("lobster")) }  # whole routine

    {
        # combat-specific watchers, only active in this block
        when self.prayer < 10 {
            for p in self.prayer.active { disable(p) }
        }
        when combat.target.hp_fraction < 0.1 {
            note("target almost dead")
        }
        attack!(target)
        wait_until(target.hp == 0)
    }
    # the two combat-specific watchers stop here

    walk_to!(x = bank.x, y = bank.y)
    deposit_loot()
}
```

### Truth semantics

**Edge-triggered on rising edge.** A handler fires when the
predicate transitions from false to true. Registration with
already-true counts as the rising edge — if you enter a block
where `self.hp` is already 35 and register `when self.hp < 40`,
the handler fires immediately.

- HP currently 35, register `when self.hp < 40 { eat!() }` →
  fires once
- After eat, HP = 50 → no fire (true → false transition)
- HP drops to 35 again → fires (rising edge)
- HP stays at 30 for 10 ticks → no extra fires

The watcher tracks last-known-truth-value. On any re-evaluation,
if last = false and now = true, fire. Update last.

### Subscription qualifiers

For non-numeric predicates, qualifier keywords make the
transition kind explicit:

```
when self.is_in_combat { ... }                # default = becomes-true
when self.is_in_combat becomes true { ... }   # explicit (same as above)
when self.is_in_combat becomes false { ... }  # falling edge
when self.is_in_combat changes { ... }        # either edge
when combat.target changes { ... }            # any non-equal change
when inventory.slots changes { ... }
when world.ground_items added { ... }         # collection grew
when world.players removed { ... }            # collection shrank
when self.skills.fishing.level increases { ... }  # counter went up
when self.skills.fishing.xp increases by 100+ { ... }  # by-threshold
```

> **Implemented today**: `becomes true` (the default), `becomes
> false`, and `changes`. The `increases` / `added` / `removed`
> qualifiers and the `by N+` threshold form shown above are **not
> yet wired** — the AST only carries `WhenBecomesTrue` /
> `WhenBecomesFalse` / `WhenChanges` (`dsl/ast/ast.go`); tracked as
> TODO.md **DSL-19**. Until the delta qualifiers land, express
> "went up" as `when self.skills.fishing.level changes { ... }` and
> check the new value inside the body.

### Re-evaluation cadence

Watchers re-evaluate at every `dispatchPendingEvents` boundary
(`dsl/interp/events.go`): **around every yielding action call**
(before and after), every **200ms slice during `wait`**
(`waitSlice`, `dsl/interp/interp.go`), and every **100ms poll
inside a blocked `select`** (`selectTickInterval`,
`dsl/interp/watchers.go`). Interactive sessions with no routine
loop (REPL, debug-http) pump the same path via
`Session.PumpEvents` (`dsl/interp/repl.go`). Cheap because:

- All watchable expressions are pure (no I/O, no LLM, no actions)
- World state changes happen on inbound packets; a routine with
  no actions and no inbound activity has nothing to recompute

Typical routine has 1–5 active watchers; that's <50 ops per
re-evaluation. Negligible.

### What's NOT watchable

Validator rejects any `when` expression that:
- Calls a primary action (would mutate state)
- Calls an LLM stdlib function (`contemplate_reality`, `decide`,
  `evaluate`, `exec`, `improvise`) — too expensive per tick
- Calls `wait` or `wait_until`

Memory-touching calls (`recall`, `relation_with`) ARE watchable
since they don't go through the LLM — but they hit mesa, so use
sparingly in subscriptions.

## `select` — block until one of these fires

Pattern from Go's `select`. Wait at this point in the routine
until one of the cases becomes true, then run that case's body.

```
routine fight() {
    attack!(target)
    select {
        when target.hp == 0 {
            pick_up_drops_near(target.position)
        }
        when self.hp < 15 {
            flee!()
            abort "LOW_HP"
        }
        on chat_received(speaker, msg) {       # bus event as a case
            if msg == "stop" { abort "TOLD_TO_STOP" }
        }
        timeout 30s {
            abort "STUCK_IN_COMBAT"
        }
    }
}
```

**Semantics:**

- **Three case types**: `when expr { }` (state transition),
  `on event(args) { }` (bus event), `timeout Ns { }`.
- Block at this point until one case becomes ready (rising edge
  for `when` cases, event-arrives for `on` cases, time-elapsed
  for `timeout`).
- **Truth at select-entry counts.** If `target.hp == 0` is
  already true when select starts, that case fires immediately
  on the first poll.
- **First-declared wins** when multiple cases are simultaneously
  ready. Deterministic, unlike Go's pseudo-random select.
- Run the winning case's body. Other cases drop without firing
  even if they would have been true.
- Exit the select. Subsequent statements run normally.

`break` and `continue` inside a select case body **propagate to
the enclosing loop**, not the select. (The select itself exits
automatically once a case fires.) This matches the common Go
pattern of `for { select { ... break } }`:

```
while kills < 10 {
    target = nearest_npc(n => n.name == "Goblin")
    attack!(target)
    select {
        when target.hp == 0 { kills = kills + 1 }
        when self.hp < 10 {
            flee!()
            break           # exits the WHILE loop, not just the select
        }
        timeout 20s {
            continue        # next iteration of the WHILE loop
        }
    }
}
```

If a `select` appears outside a loop, `break`/`continue` inside
its cases is a validator error (same rule as `break`/`continue`
anywhere else outside a loop).

**The `timeout Ns` case** is sugar for "fire after N seconds
elapsed." Prevents wait-forever bugs. The validator **warns**
(not errors) if a `select` has no `timeout` clause — sometimes
you genuinely want to block until an event, but most uses want
a bound. Wall-clock budget is the hard backstop in either case.

Time units accepted by `timeout`: `Ns` (seconds, default if
unit omitted), `Nms` (milliseconds), `Nm` (minutes — note: the
trailing letter disambiguates from int literals).

### Continuous selection

Like Go's `for { select {...} }`, wrap in `while` for repeated
selection:

```
routine combat_loop() {
    kills = 0
    while kills < 10 {
        target = nearest_npc(n => n.name == "Goblin")
        if target == null { wait 2; continue }
        attack!(target)
        select {
            when target.hp == 0 {
                pick_up_drops_near(target.position)
                kills = kills + 1
            }
            when self.hp == 0 { abort "DIED" }
            timeout 20s { abort "STUCK" }
        }
    }
    return kills
}
```

That replaced the Go `runtime/combat_loop.go` reactor entirely —
the file is gone from `runtime/`.

## `defer fn()` — cleanup on scope exit

LIFO stack of cleanup closures, run on any exit path (return,
abort, error, cancel). Inspired by Go's `defer`.

```
routine bank_run() {
    talk_to!(banker)
    answer!(2)                    # "I'd like to use my bank"
    defer close_bank()            # always closes the bank UI

    for slot in inventory.slots {
        if slot.is_loot {
            deposit!(slot.item, slot.amount)
        }
    }
    withdraw!("lobster", 20)
    return "ok"
}
```

If `deposit!` fails partway through, the routine aborts, the
defer fires, the bank closes cleanly. Without `defer`, the bank
UI would stay open after the abort and the next routine would
encounter a stale state.

Common uses: `defer close_bank()`, `defer unwield_torch()`,
`defer cancel_subscription(sub)` if subs are reified, `defer
logout()` for safety nets.

Deferred calls execute in LIFO order — last-deferred runs first.

## `try`/`recover` — bang-error boundary

Bang actions abort the routine on failure. `try`/`recover`
creates a boundary that catches the abort and gives you the
error to inspect.

```
try {
    walk_to!(x = spot.x, y = spot.y)
    fish!(spot)
    eat!("lobster")
} recover err {
    note(f"chain failed: {err.code} — {err.reason}")
    if err.code == "PATH_BLOCKED" {
        try_alternative_route()
    } else {
        abort err.code
    }
}
```

Semantics:

- `try` block runs normally
- If any bang inside fails, control jumps to `recover` with the
  error bound to the named identifier (`err` here)
- `err` shape: same as `Result.err` — `.code`, `.reason`,
  `.fatal` (see [`actions.md`](actions.md))
- `recover` can re-`abort` to propagate, or do something else and
  continue
- Deferred calls in the `try` block still run before `recover`
  executes — Go semantics

### `break` / `continue` semantics summary

- Inside a `while` or `for...in` loop body: behave normally
  (skip / exit the loop).
- Inside a `select` case that's inside a loop: propagate to the
  enclosing loop (the select itself exits naturally).
- Inside a `select` case that's NOT inside a loop: validator
  error.
- Anywhere else (proc body, handler body, routine top level):
  validator error.

## `super()` and `extends host` — handler overrides

This is the two-tier handler model from dsl.md, finally
implementable now that we have the persona layer concept
nailed down (see [`thought-architecture.md`](thought-architecture.md)).

The host's persona registers **default handlers** for common
events. A routine can override those defaults locally:

```
on damage_taken(amount, source) extends host {
    log_combat_event(source, amount)
    super()                       # forward (amount, source)
                                  # to host's default reaction
}
```

`super()` calls the parent handler. With no args, it forwards
the args the override received. With args, it passes those
instead — useful for tweaking before delegating:

```
on damage_taken(amount, source) extends host {
    if relation_with(source.name).is_friend {
        super(amount / 2, source)     # soften before host reacts
    } else {
        super()                       # forward as-is
    }
}
```

Three patterns the host's defaults might want overridden:

- **Augment** — do something first, then call super (logging,
  state capture, mood update)
- **Modify args** — call super with different values (the
  "halve the damage if attacker is a friend" example)
- **Suppress** — don't call super at all (handler runs in place
  of the default rather than alongside it)

### What the persona registers

The host's persona is no longer format-TBD — the schema is BUILT
(`persona/persona.go`; authoring guide `docs/persona-authoring.md`).
What it does **not** yet declare is default `on` handlers: the
schema has no handler field, and no registry feeds the interpreter
defaults. That registry is the missing piece (TODO.md **DSL-13**).
When it lands, routines that `extends host` see the persona's
handlers as the parent chain.

If no persona-level handler exists for an event, `super()` is a
no-op (logs a debug line, returns null).

### `super()` scope rules

- Valid: inside the body of an `on event() extends host`
  handler.
- Validator error: anywhere else (routine body, proc body,
  non-`extends` handler, `when` predicate, `select` predicate,
  `require` block, top-level statement).

### Status

**PARSED-NOT-EXECUTED.** Routine-level `on` handlers *do* dispatch
today (`registerHandlers` indexes both file-level and
`routine.Handlers`) — the open piece is specifically the
`extends host` / `super()` override chain. The persona schema now
exists (`persona/`), but there is no default-handler registry for
the interpreter to source parents from, so `super` is a
**validate-time error** wherever it appears (see
`dsl/validator/validator.go`, "super() is only valid inside an
'on event() extends host' handler body"). TODO.md **DSL-13** tracks
the wire-up, now unblocked by the persona tier. Don't author
routines against `extends host` / `super()` yet.

### Phase-4 generalization — the 3-tier stackable handler stack (PLANNED)

> **STILL A VALIDATE-ERROR TODAY.** This subsection does not change
> the status above: `extends host` / `super()` remain
> PARSED-NOT-EXECUTED, a validate-time error wherever they appear.
> What follows is the live design spec for TODO.md **DSL-13**, not a
> description of anything that runs now.

The parked model above is **two tiers** — routine override →
persona default. The settled Phase-4 design **generalizes that into
three**:

1. **Persona base reflex** — always-on default owned by the persona,
   persisting across every routine. When the host is new this is a
   raw LLM-punt (`on attacked_by(src) { decide("fight or flee?") }`).
2. **Learned specialization** — a context-scoped fast handler the
   brain wrote to *replace the punt in one situation* ("in the
   hobgoblin mine, always eat at HP<30 instead of asking"). Inserted
   between base and routine; context-scoped via the existing `when` /
   `bounds` gates.
3. **Routine-scoped override** — what we already have: routine-level
   `on` handlers, active only while the routine runs, layered onto
   the lower two via `extends host` / `super()`.

The novelty over today is that handlers become **first-class,
individually-addressable, and persisted**: a reflex is a slot keyed
by `(event_name, context)`, recorded in a planned mesa
`handler_versions` table (a near-clone of `routine_versions`,
carrying `origin` / `parent_version` / `rationale` and a per-handler
**`punt_count`**). Because the slot lives in the persona registry —
not inside the routine's VM frame — **a handler can start as an
LLM-punt and be specialized over time without reloading the
routine**: the brain installs a faster implementation into the slot
out-of-band, and the host simply *gets faster* at reacting in that
context. The punt→specialize loop is the same one `routine_versions`
already instruments (`docs/mesa.md`), mirrored at per-handler
granularity, with one discipline the prior art forces: **hysteresis**
— a freshly-specialized handler must not be re-punted inside a
refractory window, or the signal oscillates between "compile" and
"punt again."

`super()` is exactly **the composition operator this stack needs**:
it chains to the next armed lower layer (the augment / modify-args /
suppress patterns above map 1:1 onto fall-through — call `super()` to
delegate down, omit it to replace). The stack is deliberately
**bounded and shallow** (these three tiers): ordered fall-through
scales, but deep mutual suppression past ~5 layers does not — the
subsumption-architecture and behavior-tree-fallback precedent. So
the rule is "a small fixed number of ordered fallback layers," not
an arbitrarily tall tower. The three tiers supply the layers;
`extends host` supplies the edges; `super()` is the fall-through that
powers both the augment/suppress patterns *and* the punt-safety-net
(a specialization that fails in an uncovered context falls through to
the base punt, which re-raises `punt_count` and re-triggers
specialization — the fall-through *is* the safety net).

Again: this is the design for the persona tier; it does not run today.

## What we removed from dsl.md

- `on hp_below(40)` and `on fatigue_above(85)` — these were
  threshold-arg handlers. Replaced by `when self.hp < 40 { ... }`
  and `when self.fatigue > 85 { ... }`. Same expressive power,
  composes naturally with the rest of the language.
- `on npc_appeared(npc)` and `on npc_moved(npc)` — redundant
  once routines can iterate `world.npcs` between actions. If
  enough routines actually need them, we can revisit, but they
  shouldn't be in the default set.

## The interrupt-priority ladder (PLANNED)

> **DESIGN, NOT BUILT.** Like the 3-tier stack above, this is
> design gated on the persona handler tier. Today there is no notion
> of handler tiers, priority, or interruptibility regions — see the
> honest nuance about run-all dispatch below. Nothing here ships.
> (Detour tiers / pause / displacement arbitration shipped on the
> runtime side; the *handler-tier* remainder is TODO.md **C-11**.)

When the persona tier lands, handlers will carry a **TIER declared
at registration** — the host writes a label (e.g. `survival`,
`social`), it never juggles priority integers. A single-threaded
engine arbitrates by the ladder; tie-breaks are by
**context-specificity**, not source position. The ordering, highest
first:

| Tier | What it is | Interrupt behavior |
|---|---|---|
| **T0 survival** | low-HP eat / flee / sleep-captcha | preempts everything; **suppresses** the social interrupt |
| **T1 committed/modal** | open trade / duel / bank + active combat | orients but **DEFERS the reply** ("busy, one sec") until the region clears |
| **T2 directed-social / topic** | name-directed mention, PM, or a watched-topic match | preempts the grind; does **not** preempt T0/T1 |
| **T3 grind** | the current routine | the default activity |
| **(below T3) ambient-observe** | not directed, not a watched topic | attention-gated OBSERVE event, **never** a preemption |

The two **hard suppressors** of a social interrupt are exactly
**T0 (survival)** — your name gets no stop-and-talk while you run for
your life — and **T1 (committed regions)**, which *orient but defer
the reply* rather than yanking you out. (Combat is T1, which is why a
name during combat is *supposed* to wait.)

**Honest nuance — this is a real semantic change, not free reuse.**
TODAY, `on`-handler dispatch runs **ALL** matching handlers for an
event in declaration order — a run-all fan-out within the tier
(`dsl/interp/events.go`). The planned change is **first-match-wins**
(single-threaded, ordered by context-specificity, with `super()`
fall-through) **for the SURVIVAL tier only**, where two handlers both
acting *is* a bug. Run-all **FAN-OUT stays correct for the OBSERVE
tier** (chat-logging, paint, XP/inventory deltas — many independent
consumers). So the migration *splits the two tiers*; it does not
abolish fan-out, and it carries migration risk for any routine that
today relies on two handlers for one event both firing.

**The two-phase chat response.** A directed chat reply is split so
the human-visible reaction time is the action boundary, not the LLM
latency:

- **Phase 1 — ORIENT (immediate, deterministic, NO LLM).** Face the
  speaker + a tiny ack at the next action boundary (~1–2s). This is
  the **persona base layer** of the 3-tier stack — `on addressed →
  face + ack`, the cheapest believable reflex every host ships with.
  Never punts to the brain.
- **Phase 2 — REPLY (patient, async, LLM).** The actual reply is
  composed off the action loop and lands in the **chat queue** when
  ready. Between orient and reply the host keeps grinding or pauses,
  persona-dependent.

This is why between-actions dispatch (no mid-action preemption) is
*enough* for chat: the part that must be instant is a between-actions
reflex; the part that is slow is async and never needed mid-action.

## Content-keyed chat watchers (PLANNED)

> **DESIGN, NOT BUILT.** There is no content-keyed
> watcher syntax, no tier label, and no name-directedness
> classification today; every `on chat_received` handler currently
> fires on *every* public line and filters in its own body. Nothing
> here ships. (The relational verbs the v2 example leans on —
> `rel.has_tag`, scratch `cache` — are TODO.md **DSL-24**;
> directedness classification is part of **C-11**.)

A goal can install a **standing TOPIC-keyed chat watcher** so the
host pursues a goal *opportunistically* across its ordinary activity
instead of idling on it ("I need steel bars" → install a watcher, go
back to mining, pounce when anyone mentions the topic — **ambient
goal pursuit**). The planned shape is a `when`-guard on
`chat_received`, with a sugar form:

```
on chat_received(p, msg) when msg.contains("steel bar") { ... }   # explicit
on chat("steel bar") { ... }                                      # sugar
```

**Matching is layered by cost.** Keyword matching, **normalized**
(case-insensitive + whitespace-collapse + simple plural-stem, so
"steel bar" catches "SELLING STEELBARS") is **BASIC** — most hosts
live here. Semantic "is this message *about* X" is an LLM/embedding
judgment *per line* and is therefore **ADVANCED + gated** (a budget
item, earned via the disclosure gate).

**Directedness sets the tier.** A **name-directed mention or a PM** is
a high-priority **social interrupt (T2)** — you cannot be addressed by
name and keep mining without a glance. **Ambient chat** (not directed,
not a watched topic) is **observe-tier** — attention-gated, never a
preemption. (PM ≥ public-name-directed > ambient.)

**The crude→matured trajectory.** The crude shape is basic to write;
the wisdom is *learned*. A fresh host writes a v1 that pesters
everyone, every time anyone says "steel bar" — including people who
just said they have none, and the same person five times. It gets
ignored, then annoyed at; the failure signal accrues, and over
`routine_versions` it learns its way to a v2 that adds rate-limiting
(a per-player ask counter) and relational checks (don't re-ask
someone tagged "no-steel"), with a polite capped fallback:

```
# steel_bar_hunter v2 — wisdom emerged (memory-aware + rate-limited)
on chat_received(speaker, msg) social when msg.matches_topic("steel bar") {
    if rel.has_tag(speaker, "no-steel") { return }   # learned: don't re-ask
    asks = cache.incr(f"asked:{speaker}")
    if asks > 2 { return }                            # rate-limit
    engage()
    if asks == 1 { ask_to_buy(speaker) }
    else         { ask("do you at least know who sells them?") }
}
```

That is a clean, observable instance of differential learning inside
one handler — recorded in `routine_versions` with `origin =
'self_revision'` and a rationale. (`social`, `matches_topic`,
`rel.has_tag`, and the scratch `cache` verbs are all part of this
Phase-4 design, not the shipped surface.)

## Examples

### Auto-eat (replaces `runtime/auto_eat.go`)

(Historical heading: `runtime/auto_eat.go` was the hardcoded Go
reactor this pattern replaced — that file is long gone from
`runtime/`; auto-eat lives in routines now.)

```
# auto_eat.routine — registered as a long-running handler bundle
routine auto_eat() {
    when self.hp_fraction < 0.4 {
        food = inventory.find_any(["lobster", "salmon", "bread"])
        if food == null {
            abort "OUT_OF_FOOD"
        }
        eat!(food)
    }
    # block forever — the host's controller cancels via ctx
    wait_until(false)
}
```

### Combat loop with safety net

```
routine kill_goblins(max_kills = 10) {
    defer close_all_interfaces()

    when self.hp_fraction < 0.3 {
        food = inventory.find_any(["lobster", "salmon", "bread"])
        if food { eat!(food) }
        else { abort "OUT_OF_FOOD" }
    }
    when inventory.is_full {
        abort "INVENTORY_FULL"
    }

    kills = 0
    while kills < max_kills {
        target = nearest_npc(n => n.name == "Goblin")
        if target == null {
            wait 2
            continue
        }
        attack!(target)
        select {
            when target.hp == 0 {
                pick_up_drops_near(target.position)
                kills = kills + 1
            }
            when self.hp == 0 { abort "DIED" }
            timeout 30s { abort "STUCK" }
        }
    }
    return kills
}
```

### Conversational handler with persona override

(Validate error today — `extends host` / `super()` are
PARSED-NOT-EXECUTED; see the `super()` Status section above.)

```
on chat_received(speaker, message) extends host {
    # log first, then let the persona's default reaction run
    note(f"public chat: {speaker} said {message}")
    super()
}
```

## Implementation notes

- `on` handlers index by event name — file-level, routine-level,
  and `bounds`-scoped (location-filtered) all dispatch
  (`dsl/interp/events.go` `registerHandlers` /
  `registerBoundedHandlers`).
- `when` watchers: each block opens a "watcher scope" with its
  own slice of registered watchers; on scope exit, pop & unregister
  (`dsl/interp/watchers.go` `pushWatcherFrame`).
- `select` polls its cases every 100ms (`selectTickInterval`);
  first transition wins; cancel the rest.
- `defer` is a per-routine stack of closures, LIFO unwound in
  the `execBody` recover path (`dsl/interp/interp.go`
  `drainDeferred`).
- `try`/`recover` is a special-case panic boundary: bang aborts
  panic with a typed signal that `try` catches; non-bang errors
  flow through normally (`dsl/interp/interp.go` `execTry`).

All of the above are **implemented and tested** (`dsl/interp/`
`watchers_test.go`, `try_test.go`, `defer_test.go`,
`events_test.go`). The remaining design work is the per-handler
`extends host` / `super()` chain — TODO.md **DSL-13** — plus the
event-vocabulary remainder in **DSL-19**.
