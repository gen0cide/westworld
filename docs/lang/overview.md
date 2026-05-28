# Overview — The IFTTT Mental Model

> **Host** — an autonomous AI actor in the system. The `.routine`
> language is what hosts use to drive themselves. See
> [README.md](README.md) for the broader terminology and
> [`thought-architecture.md`](thought-architecture.md) for how the
> routine layer fits into the host's full mind.

The language is structured around three layers, mapped onto an
IFTTT framing:

| IFTTT term | DSL layer | What it does |
|---|---|---|
| **IF THIS** | State / query layer | Read-only views over the host's perception of itself and the world. Every UI-visible piece of state is a query. |
| **WHEN THIS** | Event / subscription layer | Reactive handlers that fire on state transitions (or on transient bus events). |
| **THEN THAT** | Action layer | Verbs the host can take. Send a packet, change game state, possibly fail. |

Routines compose these three layers. A small combat routine reads:

```
routine kill_goblin() {
    require {
        self.hp > 10              # IF THIS — query
        inventory.has("lobster")  # IF THIS — query
    }

    when self.hp < 40 {           # WHEN THIS — subscription
        eat!("lobster")           # THEN THAT — action (bang form)
    }

    target = world.npcs.nearest_by_name("Goblin")
    attack!(target)               # THEN THAT — action

    select {                       # block until one fires
        when target.hp == 0 { return "killed" }
        when self.hp < 15 { abort "low_hp" }
        timeout 30s { abort "took_too_long" }
    }
}
```

Three things to notice in that snippet:

1. **Queries are everywhere.** `self.hp`, `inventory.has(...)`,
   `target.hp` — these are all attribute access on the reserved
   entities `host` / `self` / `world` / `inventory` / `combat`.
   Cheap, side-effect-free, always reflect current world state.

2. **Subscriptions look like queries with `when` in front.** The
   `when self.hp < 40` watcher reuses the same expression syntax
   as the require block. The validator ensures the expression is
   pure (no actions, no LLM calls), so re-evaluating it every
   tick is cheap.

3. **Actions are verbs that can fail.** The `!` form (`eat!`,
   `attack!`) asserts success and aborts the routine on failure.
   The non-bang form returns a `Result` the caller branches on.

## Ontology: `host` vs `self`

The host is the autonomous AI actor — the agent inhabiting one
RSC character. Inside a routine, two reserved names give
different views of that same host:

- **`self`** — first-person view from inside the routine.
  `self.hp`, `self.position`, `self.skills.fishing.level`. What
  "I" know about myself right this instant. Mostly used for
  game-state queries.
- **`host`** — the persona / character / identity layer the
  routine runs on top of. `host.persona.shyness`,
  `host.defaults.on_damage_taken`. Used by `extends host`
  overrides and persona-tier default handlers.

Same physical entity, two vantage points. `self` is the
inhabitant; `host` is the substrate the inhabitant runs on.

For practical purposes: when you want vitals or world state, use
`self`. When you want the persona/identity layer, use `host`.
Most routines use `self` heavily and `host` rarely (mostly only
in `extends host` chains).

## What's NOT in this layer

Routines are deterministic. They don't make creative decisions —
that's the brain's job, accessed via stdlib functions like
`contemplate_reality()`, `decide()`, `evaluate()`, `exec()`. See
[`thought-architecture.md`](thought-architecture.md) for how the
routine layer interfaces with the LLM-driven brain above and the
world-state / memory layers below.

## Three layers, three docs

The full taxonomy is split across three sibling docs:

- [`state.md`](state.md) — every accessor on `self`/`host`/`world`/
  `inventory`/`combat`, ~100 fields total
- [`events.md`](events.md) — `on` / `when` / `select` / `defer` /
  `super()` / `try`/`recover`, plus the categorized event list
- [`actions.md`](actions.md) — every verb in the language plus
  the Result/Error model

Start with state.md — the other two layers stack on top of it.
