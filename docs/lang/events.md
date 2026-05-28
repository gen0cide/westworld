# Events — The WHEN THIS Subscription Layer

> **Host** — an autonomous AI actor. This doc covers everything
> a host's routine does *reactively*: handlers, watchers, blocking
> waits, cleanup, and overrides.

## Status

- **Implemented today**: file-level `on event(args) { ... }` for
  five bus events (`chat_received`, `private_message`,
  `server_message`, `coords_changed`, `trade_request`).
  Persistent through the routine's run.
- **Planned**: `when <expr>` block-scoped watchers, `select`,
  `defer`, `super()`, `try`/`recover`, `extends host`.
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
interpreter drains the event queue around yielding action calls).

Handler bodies can call any expression, action, or proc. They
have read-only access to the routine's locals.

Handlers **cannot** `return` from the routine (they'd be
returning from a different frame). They **can** `abort` to
terminate the parent routine — useful for "kill the routine on
this kind of message."

### Category A — true bus events (need persistent `on`)

These have no underlying queryable state. They're transient.

| Event | Args | Source |
|---|---|---|
| `chat_received` | `speaker, message` | Public chat from another player |
| `private_message` | `speaker, message` | Incoming PM |
| `server_message` | `text` | System message from server |
| `damage_taken` | `amount, source` | We were hit (distinct from "my hp is X now") |
| `level_up` | `skill, new_level` | Skill level just increased |
| `trade_request` | `from` | Another player initiated a trade |
| `coords_changed` | `x, y` | Our position updated |
| `death` | (none) | We just died |

That's the canonical list. Anything else is better expressed as
`when` over the query layer.

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

### Re-evaluation cadence

Watchers re-evaluate **between action calls** plus at fixed tick
intervals (~100ms when the routine is idle). Cheap because:

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
        when chat_received(speaker, msg) {     # bus event as a case
            if msg == "stop" { abort "TOLD_TO_STOP" }
        }
        timeout 30s {
            abort "STUCK_IN_COMBAT"
        }
    }
}
```

**Semantics:**

- Block at this point until one case becomes true (rising edge
  for `when`-cases, fires-on-event for bus-event cases).
- Truth at select-entry counts: if `target.hp == 0` is already
  true when select starts, that case fires immediately on the
  first poll.
- Run the winning case's body. Other cases drop without
  firing — even if they would have been true.
- Exit the select. Subsequent statements run normally.

The `timeout Ns` case is special — sugar for "fire after N
seconds elapsed." Prevents wait-forever bugs. Use it on every
select unless you have a strong reason not to.

### Continuous selection

Like Go's `for { select {...} }`, wrap in `while` for repeated
selection:

```
routine combat_loop() {
    kills = 0
    while kills < 10 {
        target = nearest_goblin()
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

That replaces the Go `runtime/combat_loop.go` reactor entirely.

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

The host's persona declares its own `on` handlers in a separate
persona definition (`docs/personas.md` — exact format TBD).
Routines that `extends host` see those as the parent chain.

If no persona-level handler exists for an event, `super()` is a
no-op (logs a debug line, returns null).

### Status

**Parses and validates today.** Routine-level handlers don't
yet dispatch alongside file-level handlers (validator accepts
them, interpreter ignores them). Wire-up needs the persona tier
to register defaults somewhere the interpreter can find them.

## What we removed from dsl.md

- `on hp_below(40)` and `on fatigue_above(85)` — these were
  threshold-arg handlers. Replaced by `when self.hp < 40 { ... }`
  and `when self.fatigue > 85 { ... }`. Same expressive power,
  composes naturally with the rest of the language.
- `on npc_appeared(npc)` and `on npc_moved(npc)` — redundant
  once routines can iterate `world.npcs` between actions. If
  enough routines actually need them, we can revisit, but they
  shouldn't be in the default set.

## Examples

### Auto-eat (replaces `runtime/auto_eat.go`)

```
# auto_eat.routine — registered as a long-running handler bundle
routine auto_eat() {
    when self.hp_fraction < 0.4 {
        food = inventory.find_food()
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
        food = inventory.find_food()
        if food { eat!(food) }
        else { abort "OUT_OF_FOOD" }
    }
    when inventory.is_full {
        abort "INVENTORY_FULL"
    }

    kills = 0
    while kills < max_kills {
        target = world.npcs.nearest_by_name("Goblin")
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

```
on chat_received(speaker, message) extends host {
    # log first, then let the persona's default reaction run
    note(f"public chat: {speaker} said {message}")
    super()
}
```

## Implementation notes

- `on` handlers index by event name, file-level only for v1.
- `when` watchers: each block opens a "watcher scope" with its
  own slice of registered watchers; on scope exit, pop & unregister.
- `select` waits on a shared scheduler that polls watchers
  between ticks; first transition wins; cancel the rest.
- `defer` is a per-routine stack of closures, LIFO unwound in
  the `execBody` recover path.
- `try`/`recover` is a special-case panic boundary: bang aborts
  panic with a typed signal that `try` catches; non-bang errors
  flow through normally.

Total estimated effort for the new constructs: 4-5 days of
focused work + tests. The validator changes are the trickiest part
(detecting subscription-safe expressions).
