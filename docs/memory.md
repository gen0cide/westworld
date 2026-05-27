# Memory — fungible, layered, human-like

## The problem

Bots with perfect recall feel as alien as bots with no memory. Humans remember selectively — recent events vividly, older events as gist, some forgotten entirely, some reinforced through repetition. Memories transform over time, becoming summaries of themselves, sometimes inaccurate.

We want host memory to have these properties: **selective retention, importance weighting, reinforcement on encounter, hierarchical compression, cue-dependent retrieval.** The Stanford Generative Agents paper (Park et al., 2023) is the closest reference; we extend it with explicit fungibility (compression stages) and tighter integration with the brain's tiered routing.

## Four categories, four lifecycles

| Category | Examples | Lifespan | Storage |
|---|---|---|---|
| **Working** | Last few minutes of events: who said what, where I walked, who's nearby | Seconds to minutes | In-memory ring buffer on cradle |
| **Episodic** | Discrete significant events: "I traded with JimBob", "I died at Greater Demons" | Hours to weeks; decay-based | Mesa: `episodes` table |
| **Relational** | Per-person facts: "JimBob is a reliable trader, 14 trades, 700k gp total" | Months to indefinite | Mesa: `relationships` table |
| **Reflective** | Generalizations: "I'm better at fishing than fighting", "Lumbridge is friendlier than Varrock" | Long-lived, revisable | Mesa: `reflections` table |

Plus a fifth category (procedural) that's the **routine library** in mesa — see [dsl.md](dsl.md). That's "remembered how-to" rather than "remembered events."

## Salience as the central concept

Each memory has a salience score that determines retrieval priority:

```
salience(t) = importance × recency_decay(t) + reinforcement_bonus
recency_decay(t) = exp(-(now - last_accessed) / tau)
```

Tunable parameters:
- `importance` ∈ [0, 1]: assigned at write time (heuristic + occasional LLM for borderline cases)
- `tau` (decay half-life): varies per memory type. Episodic ~3 days. Relational ~30 days. Reflective ~6 months.
- `reinforcement_bonus`: increments each time the memory is retrieved or related events occur. "Met JimBob again — strengthen that record."

This gives human-like properties:

- **Recency bias**: recent things are more retrievable (high decay term initially)
- **Importance weighting**: high-stakes events persist; mundane ones fade
- **Reinforcement**: repeated encounters with a person/situation keep the memory vivid
- **Cue-dependent retrieval**: vector search means relevant cues surface relevant memories, not chronological order

## Fungibility — compression stages, not deletion

Memories don't go from "perfect recall" to "gone." They go through a continuum of compression:

```
Stage 1: full event detail
    "On Tuesday at 14:32 I traded JimBob 100 lobsters for 50k gp at the
     Varrock east bank. We chatted about the price of yew logs. He seemed
     friendly."
                              ↓ decay (salience falls)
Stage 2: summarized
    "Traded lobsters with JimBob a few weeks ago, got fair price."
                              ↓ further decay
Stage 3: gist only
    "I think I've traded with JimBob before — seems fine."
                              ↓
Stage 4: dropped (only the relational record persists; the episode is gone)
```

Each stage transition is an LLM compression call (Haiku, cheap, batched during idle). The compression sweep runs hourly per-host on mesa-side background jobs.

When the brain retrieves a stage-3 memory, the prompt marks it `(fuzzy memory)` so the brain knows to treat it as approximate. The host might say "I think we've traded before, can't remember exactly when" — which is exactly how humans recall fading memories.

## Importance scoring at write time

Two-tier:

**Tier 1: heuristic rules** (free, instant, covers ~90% of events)

```go
func ImportanceOf(ev event.Event) float64 {
    switch e := ev.(type) {
    case event.Trade:           return clamp(0.3 + valueScore(e.GpAmount) + relScore(e.Other))
    case event.Death:           return 0.9
    case event.Scammed:         return 1.0
    case event.LevelUp:         return 0.5
    case event.ChatReceived:    return 0.1 + relScore(e.Sender)*0.4
    case event.NpcKill:         return 0.05 + rarityScore(e.NpcId)*0.5
    case event.QuestStart:      return 0.7
    case event.FirstMeeting:    return 0.4
    case event.ItemFound:       return 0.05 + rarityScore(e.ItemId)*0.6
    }
    return 0.1
}
```

**Tier 2: LLM-assigned** for events the heuristic returns "unsure" on (importance between 0.2 and 0.6). Cheap Haiku call: "Rate this event's memorability 0-10."

Per Alex's preference: LLM-first, extract heuristics from labeled data later if cost demands. The above table is illustrative — actual heuristics will emerge from observation.

## Relational memory — the sticky one

Episodic memory decays. Relational memory does not — it's the long-lived social fabric.

When you trade with JimBob:
1. **Trade event fires** from the wire layer
2. **Episodic memory** written with full detail (item list, location, time, snippets)
3. **Relational memory** updated: increment `trade_count`, update `last_traded`, accumulate `value_exchanged`, adjust `trust_score` if outcome was good/bad
4. **Notes field** may get an agent-written entry: "JimBob is reliable, knows the yew log market"

Later, when JimBob walks by your host:
1. Entity update packet identifies JimBob
2. Cognition retrieves the relational record automatically
3. The next strategist prompt includes: "Players nearby: JimBob (you've traded 14 times, ~700k gp, last 3 days ago; notes: 'reliable trader, fair on yew logs')"
4. The brain can decide naturally: greet, ignore, propose another trade, etc.

**The structural data is lossless** (encounter counts, trust scores, trade volumes). Only the `notes` field is subject to compression — a human might forget the conversation but remember "Bob was a fair guy." Same for hosts.

A relational record is **never fully dropped**, even after years. The host always "knows" they've met JimBob before, even if details are gone.

## Reflective memory — the agent's narrative voice

Reflections are higher-order generalizations the agent forms about itself, others, and the world. They're generated periodically by Sonnet from recent episodic content.

Example reflections that might emerge:

- "I make more gp from lobsters than from combat — I should focus there."
- "Players in Edgeville are quicker to engage in conversation than Varrock."
- "JimBob always shows up at the bank around the same time of day."
- "I lose interest in tasks faster when others are watching."

These shape long-term strategy. They're available for retrieval during Strategic-class brain calls. They're revisable when evidence shifts — `reflections.last_revised` tracks when an opinion changed.

The reflection generator runs every ~30 minutes of active play OR after N significant events. Cheap because rare.

## Working memory (local to cradle)

The most recent ~50 events live in an in-memory ring buffer on the cradle host. This is cheap, sub-millisecond access, used for:
- Reactive condition checks ("was I attacked in the last 3 seconds?")
- The "Recent events" section of brain prompts
- Periodic batch-write to mesa as episodes

It's the only memory tier that's truly host-local — everything else is mesa-backed.

## Cost implications

Memory operations are the second-biggest cost driver after brain calls themselves:

| Operation | Per host per day | Notes |
|---|---|---|
| Importance scoring (LLM tier, Haiku) | ~$0.003 | Heuristic handles most |
| Reflection generation (Sonnet, ~6/day) | ~$0.03 | High value per call |
| Compression / decay (Haiku, ~10-30/day) | ~$0.005 | Batched, background |
| Memory infrastructure overhead | ~$0.04 total | |

This buys human-like memory. The alternative (perfect recall) is cheaper but breaks believability.

## Why this matters for research

The fungible memory system is the **substrate** on which the research goals depend:

- **Long-term strategic accomplishment** requires durable memory of plans, attempts, outcomes
- **Community formation** requires recognition: hosts remembering each other across encounters, building reputation over time
- **Ethics observation** requires consequences: a host that scams gets remembered by victims, and that memory shapes future interactions
- **Believability** requires human-like memory failure: hosts that recall *imperfectly*, the way humans do

Without this layer, the project becomes "stateless LLMs in RSC." With it, hosts have actual identity that persists, evolves, and shapes their behavior.

## Open questions

- **Tau values per memory type**: 3 days / 30 days / 6 months are guesses. Need to tune from observation.
- **Compression LLM prompt design**: how to compress an episode into 1-2 sentences while preserving the "feel" of the memory (subjective tone, emotional context, not just facts)?
- **Reinforcement bonus formula**: linear? Exponential with diminishing returns? Bayesian update? Needs experimentation.
- **Cross-bot relational consistency**: if BotA's record of BotB says "trusted friend" and BotB's record of BotA says "annoying acquaintance," is that fine (asymmetric relationships are realistic) or do we want some weak consistency mechanism?
