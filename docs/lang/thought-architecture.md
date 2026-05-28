# Thought Architecture — How Routines Fit Into a Host's Mind

> **Host** — an autonomous AI actor. This doc is about what's
> *inside* the host: the layered cognitive architecture, and where
> the `.routine` layer sits within it.

A host's mind is layered. The routine layer is one piece of a
larger system that also includes perception, working memory,
long-term memory, a brain (LLM strategist), a persona (long-term
identity), and reveries (cross-cutting believability hooks). This
doc maps those layers and shows where routines interface with each.

## The cognitive stack, bottom-up

```
                           ┌──────────────────────────┐
                           │  Persona (identity)      │  who am I, long-term
                           ├──────────────────────────┤
                           │  Reveries (believability)│  cross-cutting hooks
                           ├──────────────────────────┤
                           │  Brain (LLM strategist)  │  decide & generate
                           ├──────────────────────────┤
                           │  Cognition (retrieval)   │  fetch from memory
                           ├──────────────────────────┤
                           │  Routines (procedural)   │  ←── this doc's layer
                           ├──────────────────────────┤
                           │  Memory (working)        │  short-term ring buffer
                           ├──────────────────────────┤
                           │  Events (pub/sub)        │  what's happening
                           ├──────────────────────────┤
                           │  Actions (verbs)         │  do things
                           ├──────────────────────────┤
                           │  World (perception)      │  what I observe
                           ├──────────────────────────┤
                           │  Session / Wire protocol │  packet I/O
                           └──────────────────────────┘
```

Each layer depends only on layers below it. The full enumeration
of Go packages backing each layer is in
[`docs/architecture.md`](../architecture.md) — this doc focuses
on how the **routine** layer (the `.routine` language) interfaces
with what's above and below.

## What the routine layer is

Procedural memory. Skills the host has learned: how to fish at
Port Sarim, how to grind goblins, how to bank a load of iron ore.
Each `.routine` file is one named skill. They're deterministic —
given the same world snapshot and inputs, the same routine takes
the same actions. That determinism is what makes routines
replayable, testable, and conformance-checkable.

Routines are NOT where the host's creativity lives. A routine
that says "if I see a stranger, decide what to say" doesn't
hard-code the response — it calls `contemplate_reality(...)` and
branches on whatever the brain returns. **Routines are the
*known* plans; the brain handles the *novel*.**

## How routines interface with each layer

### Down: World, Events, Actions

These are the layers the routine directly reads from and writes
to. They're documented in detail in the IFTTT layers:

- **World** ↔ **State queries** ([`state.md`](state.md)) —
  `self.hp`, `world.npcs`, `inventory.free`, etc. The world layer
  is the host's perception (mirror of inbound packets); the
  query layer exposes it to the routine.
- **Events** ↔ **Subscriptions** ([`events.md`](events.md)) —
  `on chat_received(...)`, `when self.hp < 40 { ... }`. The
  event bus publishes typed events from inbound packets; the
  subscription layer maps them onto routine handlers.
- **Actions** ↔ **Verbs** ([`actions.md`](actions.md)) — `walk_to`,
  `attack`, `eat`, `say`. The action layer sends outbound
  packets; the verb layer is just the DSL surface for those
  actions plus the Result/Error model.

The query, subscription, and verb layers ARE the routine layer's
interface to the world. Read those three docs to understand the
full surface.

### Up: Memory, Cognition, Brain

A routine can reach UP the stack via the stdlib (documented in
[`actions.md`](actions.md) under "Stdlib calls"). These are
expensive — they go through the brain or memory subsystems — and
each has a per-routine call budget.

| Stdlib call | What it does | Lives in |
|---|---|---|
| `contemplate_reality(question)` | LLM-backed open-ended decision (Sonnet); returns a short string code | Brain |
| `evaluate(situation)` | LLM-backed 0-1 numeric assessment (Haiku) | Brain |
| `decide(options)` | LLM-backed choice from a caller-supplied list | Brain |
| `exec(prompt)` | Brain authors a fresh DSL fragment; runs inline | Brain + Routine |
| `improvise(prompt)` | One-shot DSL fragment, not retained | Brain + Routine |
| `recall(query, top=5)` | Vector retrieval over episodic memory | Memory (via cognition/mesa) |
| `relation_with(name)` | Lookup a relational record | Memory |
| `reflect_now()` | Trigger synchronous reflection (LLM) | Brain |
| `note(text)` | Lightweight journal write | Memory |

The host's persona drives how often a routine should reach up.
A confident, experienced host runs deterministic routines most
of the time. A nervous, novel-situation-heavy host hits
`contemplate_reality` often. The **punt rate** (how often a
routine consults the brain) is a measurable property of routine
maturity (see `docs/research-goals.md`).

### Sideways: Persona, Reveries

These two layers cut across the stack rather than living above or
below.

- **Persona** — the host's identity (name, traits, north-star
  goals, mood). Routines read it via `host.persona.*`. The
  persona is also the parent for routine-level handler overrides:
  `on damage_taken extends host` means "the host has a default
  reaction to being hit; my routine overrides it while running."
  See [`docs/personas.md`](../personas.md).
- **Reveries** — cross-cutting "believability" hooks injected
  between actions. A host's persona-driven inner narrative,
  intrusive thoughts, occasional emotional shifts. Routines
  don't call into reveries directly; the interpreter ticks
  reveries between primary action calls automatically. See
  [`docs/reveries.md`](../reveries.md).

## Why this matters for routine authors

When you (or the brain) writes a routine, you're deciding which
layer to push a problem onto:

- "I can express this as deterministic logic over queries" →
  put it in the routine (cheap, fast, replayable)
- "This needs world knowledge I don't have inline" → `recall()`
  or use the cognition layer's structured world facts
- "This needs creative judgment" → `contemplate_reality()` or
  `decide()` (expensive, slow, non-deterministic)
- "This is the host's personality" → live in the persona, not
  the routine

Pushing problems to the wrong layer is the most common DSL
design mistake. A routine that calls `contemplate_reality()`
every iteration burns budget. A routine that hard-codes a
personality trait that should be in the persona makes the host
inconsistent across routines. The query / subscription / verb
layers are deliberately the cheapest and most expressive — the
routine should *prefer* them whenever possible.

## Determinism vs. non-determinism boundary

Inside the routine layer:
- All language constructs (control flow, expressions, actions)
  are deterministic
- Action results CAN fail (server rejects, path blocked, etc.)
  but the failure modes are typed and stable
- `wait` with a range literal (`wait 2.8..4.5`) uses the
  interpreter's seeded `Rand` for jitter — deterministic given
  the seed

Outside the routine layer (via stdlib):
- `contemplate_reality`, `decide`, `evaluate`, `exec`,
  `improvise`, `reflect_now` — all touch the LLM, all
  non-deterministic
- `recall` and `relation_with` are deterministic given the
  same memory state but the memory state itself evolves

The boundary is enforced statically. The validator marks each
stdlib call as "subscription-safe" or not. Expressions in `on`,
`when`, `select`, and `require` blocks can only use
subscription-safe calls (essentially: pure queries plus
`recall` / `relation_with`). LLM-touching calls (`contemplate_*`,
`decide`, `exec`, etc.) can only appear in action position
within the routine body.

## Concrete example

```
# fish_at_port_sarim.routine

# A handler we want active for the whole routine.
on chat_received(speaker, message) {
    # Reactive social chatter — uses the brain to decide what
    # to say back. Expensive, rate-limited, but in character.
    if speaker.is_friend {
        response = contemplate_reality(f"how do I greet {speaker}?")
        say!(response)
    }
}

proc nearest_spot() {
    return world.locs.fishing_spots.nearest(self.position)
}

routine fish_at_port_sarim(rod_type = "fly") {
    require {
        self.wielded.name == rod_type
        inventory.free > 5
        self.fatigue < 80
    }

    # Combat-style auto-eat — only active for this routine,
    # block-scoped via `when`.
    when self.hp < 40 {
        food = inventory.find("lobster")
        if food { eat!(food) }
    }

    # Cleanup runs on any exit path (return, abort, error).
    defer note("fishing run ended")

    spot = nearest_spot()
    if spot == null {
        abort "NO_FISHING_SPOT"
    }

    walk_to!(x = spot.x, y = spot.y)

    while inventory.free > 0 {
        if self.fatigue > 90 { return "tired" }

        fish!(spot)
        wait 2.8..4.5    # jittered for believability
    }

    return "inventory_full"
}
```

This routine touches every cognitive layer:

- **World / queries** — `self.wielded`, `self.fatigue`, `self.hp`,
  `inventory.free`, `world.locs.fishing_spots`, `spot.x`
- **Events / subscriptions** — `on chat_received`, `when
  self.hp < 40`
- **Actions / verbs** — `eat!`, `walk_to!`, `fish!`, `say!`, `wait`
- **Memory** — `note("fishing run ended")` in the defer
- **Brain** — `contemplate_reality` for chatter response
- **Persona** — implicitly, via the `speaker.is_friend` check
  (relational data) and the persona's chattiness traits informing
  whether to respond at all (the persona itself decides whether
  this routine even runs)

Each call has a different cost. The routine author (human or
brain-generated) chooses which layer to consult for each
sub-problem. That choice is the design work; the language
itself just makes the choice expressible.
