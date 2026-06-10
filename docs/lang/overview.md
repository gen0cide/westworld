# Overview — The IFTTT Mental Model

> **STATUS: CURRENT — verified 2026-06-10 against branch
> `tidy/structure-and-docs`, HEAD `0bfa818`.** Everything on this
> page is BUILT and executable except where marked — the `host.*`
> root is spec-reserved but not wired (see
> [Ontology](#ontology-host-vs-self) below). Open language work
> lives in [`docs/TODO.md`](../TODO.md) (the SSOT).

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

    target = world.npcs
        .filter(n => n.name == "Goblin")
        .nearest(self.position)
    attack!(target)               # THEN THAT — action

    select {                      # block until one fires
        # target.health is null until engaged, 0 at death
        when target.health == 0 { return "killed" }
        when self.hp < 15 { abort "low_hp" }
        timeout 30s { abort "took_too_long" }
    }
}
```

Three things to notice in that snippet:

1. **Queries are everywhere.** `self.hp`, `inventory.has(...)`,
   `n.name`, `target.health` — these are all attribute access,
   either on a reserved root (`self` / `world` / `inventory` /
   `combat` / …; the canonical list is
   `dsl/spec/spec_roots.go`) or on an entity view pulled from
   one. Cheap, side-effect-free, always reflect current world
   state. Note there is no `nearest_by_name(...)` helper — by-name
   selection is the `filter(lambda).nearest(pos)` chain shown
   above (list methods live in `dsl/interp/interp.go`'s
   `evalMember`).

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

> **STATUS: `host.*` is DESIGN, not built.** `host` is a reserved
> root — `dsl/spec/spec_roots.go` reserves the name ahead of
> wiring so no authored routine can bind it — but the runtime
> registers no view for it (`runtime/dsl_bridge.go` wires
> self/world/inventory/combat/trade/bank/duel/magic/prayer/shop
> only), and all five `host.*` accessor rows in
> `dsl/spec/accessors.go` are flagged `NotYetImplemented`. The
> `extends host` form is equally unbuilt: the parser accepts only
> file-path `extends "<file>"` at top level and no extends clause
> on handlers (`dsl/parser/parser.go`), and `super()` is rejected
> by the validator outside the not-yet-existent extends-host
> handler (`dsl/validator/validator.go`). The persona reflex tier
> that wires all of this is `DSL-13` in
> [`docs/TODO.md`](../TODO.md).

For practical purposes: when you want vitals or world state, use
`self` — today that means **always** use `self`; `host` becomes
useful only when DSL-13 lands.

## What's NOT in this layer

Routines are deterministic. They don't make creative decisions —
that's the brain's job, accessed via stdlib functions like
`contemplate_reality()`, `decide()`, `evaluate()` (BUILT — routed
through `Host.Strategist`, the mesa LLM client in production and
a deterministic stub in tests; handlers in
`runtime/dsl_actions.go`). `exec()` / `improvise()` are spec'd
but `NotYetImplemented` — they resolve to NOT_IMPLEMENTED stubs.
See [`thought-architecture.md`](thought-architecture.md) for how
the routine layer interfaces with the LLM-driven brain above and
the world-state / memory layers below.

## Three layers, three docs

The full taxonomy is split across three sibling docs:

- [`state.md`](state.md) — every accessor on the reserved roots
  (`self`/`host`/`world`/`inventory`/`combat`, plus
  `trade`/`bank`/`duel`/`magic`/`prayer`/`shop`), 114 spec'd paths
  total (`dsl/spec/accessors.go`, consistency-tested; 8 flagged
  `NotYetImplemented`)
- [`events.md`](events.md) — `on` / `when` / `select` / `defer` /
  `super()` / `try`/`recover`, plus the categorized event list
- [`actions.md`](actions.md) — every verb in the language plus
  the Result/Error model

Start with state.md — the other two layers stack on top of it.
