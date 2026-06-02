# Reveries — the believability layer

> **STATUS: EMPTY PACKAGE / DESIGN ONLY** (verified 2026-05-31). The
> `reveries/` directory contains **no Go files**. The `Augmenter`
> struct, its `Tick()` method, and the `Reverie` interface shown
> below do not exist; the DSL interpreter has no `reverie.tick()`
> hook between actions. The "reveries are load-bearing" claim is a
> research hypothesis driving the design, not a verified result. The
> full reverie catalog is explicitly deferred to a paired design
> session with Alex (see `docs/personas.md`, also design-only). Read
> this as the design rationale + intended shape, not implemented
> behavior.

## What reveries are

In Westworld, reveries were tiny unconscious gestures — a glance, a tic, a hesitation — that the hosts began exhibiting after Ford's update. They had no functional purpose. They were the bug that started consciousness.

For us, reveries are the same conceptual primitive: small behavioral augmentations that have no game-mechanical purpose but make the host's behavior indistinguishable from organic play. They are the substrate of believability.

Reveries are NOT polish. They are NOT the last thing to add. They are **load-bearing** — the difference between a population of drones and a population of plausible inhabitants.

## What gets injected

A non-exhaustive catalog:

| Reverie | Trigger | Example |
|---|---|---|
| Timing jitter | Every action | Walk action takes 0.8-1.4× nominal latency |
| Pathing imperfection | Every walk | Path 1-2 tiles off optimal sometimes |
| Idle wander | After N seconds of route stillness | Walk 2 tiles north and back |
| Glance | Passing a nearby player | Click their name, then move on |
| Spontaneous chat | Passing a friend / chatty mood | "hi" or "wb" or "cool armor" |
| Distraction | Mid-routine | Walk to investigate something nearby, then resume |
| Mistake | Click | Click wrong NPC, recover within 1-2 ticks |
| Stale focus | After long routine | Stop, look around, briefly idle, resume |
| Persona quirks | Persona-specific | Shy personas idle-fidget more; outgoing ones chat-initiate more |
| Fatigue acknowledgment | When fatigue rising | "I should rest soon" thought logged; eventual route to bed |

Each reverie is a small Go function that may or may not fire when called, based on dice rolls weighted by persona traits and emotional state.

## How reveries integrate with routines

The DSL interpreter calls `reverie.tick()` between every atomic action in a running routine. The reverie system inspects the host's current state and *may* inject one of the above behaviors. Most ticks, nothing fires (the dice didn't roll). When something fires, it's a small departure from the routine — a glance, a 2-tile wander, a quick chat — before the routine resumes.

The DSL also allows routines to *explicitly* call `reverie.tick()` at deliberate "pause and look around" points. This is for routine authors (the brain) who want to mark "here is a natural moment to be human."

Conceptually:

```
routine fish_at_swamp {
    walk_to(swamp_x, swamp_y)
    while inventory.free > 0 {
        if hp < max_hp * 0.3 { eat(any_food); continue }
        if fatigue > 90 { abort("tired") }
        fish(nearest_spot)
        // reverie.tick() fires automatically here, between every action
    }
    walk_to(bank_x, bank_y)
    bank_deposit_all()
}
```

The routine looks like a script. In practice, between each `fish(nearest_spot)` and the next, the reverie system has a chance to inject: glance, wander, chat, stop-and-look. The routine never has to know about these — they're invisibly woven in.

## Persona-driven, emotional-state-aware

Reveries are not uniform across hosts. The same situation produces different reveries for different personas:

- A **"shy/quiet"** persona: more idle fidget, less chat-initiation, glances away when noticed
- An **"outgoing"** persona: chat-initiates frequently, less idle, more eye-contact-equivalent (clicking names)
- A **"focused grinder"** persona: less reverie overall, more routine-faithfulness, occasionally annoyed by interruptions
- A **"curious explorer"** persona: more distraction-fire, more investigating-nearby-things, less routine-faithfulness

The "emotional state" is a lightweight runtime construct on the cradle: mood (calm/anxious/excited/bored), recent stress (deaths, attacks, scams), recent satisfaction (successful trades, levelups). Reveries query the state and weight their fire probabilities accordingly.

### Mood as a Fleeson density-distribution (lightweight)

The mood construct above is given a name and a shape: it is modeled as small **Fleeson density shifts** on the host's trait state — momentary expression nudges around an immutable personality baseline, not a separate affect organ. Two pieces, both deliberately small:

- A **fast valence/arousal** plus recent stress/satisfaction, updated out-of-band by events (a death or attack lifts stress; an XP gain or good trade lifts confidence) and **decaying toward the persona baseline** on read — no per-host daemon, just a timestamp-diff decay.
- A small, **capped, slow** shift of the *expression mean* around the immutable trait means, nudged by social context with a tiny social-influence weight (λ≈0.1). It is bounded by construction so it shifts expression, never identity — it can never become a second personality.

This is kept **deliberately lightweight**: it only (a) weights reveries (higher stress → more anxious reveries) and (b) biases the brain-prompt tone. It is **not** a heavy affect subsystem with its own components. Concretely it is a mutable `Trajectory` field of the persona schema (see [personas.md](personas.md)) — the same lightweight cradle-side state, now named.

### Reverie weights are trait-derived

The fire-weights above are **derived** from the persona's traits plus a per-agent jitter seed — **not** an independent, separately-stored named-weights bag. The same trait set that drives a host's behavior also drives its believability variance, so the two never drift apart. A reverie's weight is a small closure over `(traits + mood/expression state + per-host jitter)` evaluated at runtime — e.g. a pacing/fidget reverie lifts with extraversion and with awaiting-a-reply, an equipment-hover reverie lifts with recent damage. Only the per-host jitter is stored; the weights themselves are computed, never persisted as a bag.

### The curiosity + attention dials

The persona differences sketched above (the curious explorer, the focused grinder) are driven by two named temperament dials, each with an **immutable anchor + a mutable current** (see [personas.md](personas.md)):

- **Curiosity** — a *flavored* drive, a weight vector over **social / spatial / skill / economic / risk**. It weights the **explore-reveries**: idle-wander, investigate-nearby, glance, chat-initiate. A spatial-curious host wanders and maps; a social-curious host chat-initiates; a skill-curious host tinkers. Curiosity is *what* pulls the host, and how hard.
- **Attention** — a *scalar* `0..1`. It is the **threshold / gate** that any distraction reverie (and any external interruption) must clear to actually pull the host off its current activity. **Low attention** fires more distraction/chat — this is the "scatterbrain." **High attention** suppresses them — this **is** the existing "focused grinder = less reverie, annoyed by interruptions," now expressed as a dial rather than a hand-wave. Attention is *how well* the host resists being pulled.

## Chat as a special-case reverie

The chat queue is technically a separate subsystem (because it's async with primary actions), but conceptually it's reverie-driven. A "say something" reverie fires, the message goes into the chat queue, the chat worker drains it opportunistically between actions. The host appears to chat naturally while doing other things — but never paragraph-typing during combat.

This is the **slow** half of a two-phase chat response. The **fast** half is the **orient reflex**: when an incoming chat line is directed at the host — its username, or a PM — the host *immediately* and deterministically **turns to face the speaker** (a tiny ack), with no brain call. That orient is a **persona base reflex** every host has, fired at the next action boundary (~1–2s) — the human "look up instantly" beat. The patient, LLM-composed reply is then the async "say something" reverie above, draining through the existing chat-queue path — the "speak a beat later" half. So the special-case chat reverie is really two phases: a deterministic orient (reflex) + the async reverie that composes the reply. (See [chat-interruption-and-engagement.md] for the full ladder; design-only.)

## Mistakes as believability

Real players misclick. They try to attack out of range. They accidentally drop wielded weapons. They walk into trees because they clicked too fast.

The mistake reverie injects controlled errors:
- Click wrong NPC (recover within 1-2 ticks by re-clicking)
- Walk path with a small detour
- Attempt action with bad precondition (eat when at full HP — wastes a food, but humans do this)
- Misread inventory (drop wrong item occasionally)

Frequency is persona-tuned. A "high-precision" persona makes fewer mistakes. A "scatterbrained" persona makes more.

## What reveries are NOT

- They are not anti-detection measures against an adversary. The bar is "indistinguishable from human play" to other agents, not "evades a sophisticated bot-detection system."
- They are not random noise injection. Every reverie is purposeful — it represents a specific human behavior.
- They are not blockers of routine progress. A routine still gets done; reveries just make the journey believable.

## Implementation shape

```go
package reveries

type Augmenter struct {
    persona *persona.Persona
    state   *EmotionalState
    rng     *rand.Rand
    history []Fire   // recent reveries — avoid repeating same one back-to-back
}

func (a *Augmenter) Tick(ctx context.Context, actions action.API, world world.View) {
    candidates := a.candidates(world)        // which reveries are eligible right now
    weighted := a.weight(candidates)          // persona + state weighting
    chosen := a.maybeFire(weighted)           // dice roll; usually nothing fires
    if chosen != nil {
        chosen.Execute(ctx, actions)
    }
}

type Reverie interface {
    Name() string
    EligibleIn(world world.View) bool
    BaseWeight() float64
    Execute(ctx context.Context, actions action.API) error
}
```

Adding a new reverie = implementing the interface. The Augmenter loop never has to change.

## The persona/reveries design session

The full catalog of reveries, their weights per persona, and how they map to cohort design is a **paired design session** between Alex and Claude. See [personas.md](personas.md) for the LOE checkpoint. We do NOT pre-fill the reverie catalog — it gets co-designed as part of persona work.

## Open questions

- How many reveries should exist at v1? Probably 15-20 distinct reveries is enough to produce convincing variation.
- How does the host learn over time? Does its persona's reverie weights drift based on experience? E.g., a host that gets attacked frequently in Varrock develops a "wary glance toward Varrock-direction" reverie that wasn't in the initial template.
- Reverie observation by other hosts: does seeing a familiar reverie on another player ("oh, this player always wanders that way") create relational records? Probably yes — distinct behavioral fingerprints aid recognition.
