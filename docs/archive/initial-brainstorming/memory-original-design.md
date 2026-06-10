> **ARCHIVED (initial brainstorming), 2026-06-10.** This design was superseded or manifested
> differently. What shipped: the tiered `memory/` Manager (journal + policy table), `hostkv/`
> (pebble), `limbic/ledger.go` (uniform-prior Beta trust ledger), `runtime/memory_journal.go`,
> and the mesa side (`mesa/mesad/ltm.go` + `cron.go`). The living doc is `docs/memory.md`
> (rewritten same-commit). Still-open items were harvested into `docs/TODO.md`. Kept verbatim
> below for the record.

# Memory — fungible, layered, human-like

> **STATUS: EMPTY PACKAGE / DESIGN ONLY** (verified 2026-05-31). The
> `memory/` directory contains **no Go files**. Nothing in this doc
> is implemented: the four tiers (working / episodic / relational /
> reflective), salience scoring, the importance-heuristic Go snippet,
> compression stages, and the relational-update-on-trade flow are all
> Phase 3-4 design. The `event.Trade` / `event.Death` types the
> importance examples reference *do* exist (in `event/`), but no
> importance scorer consumes them. Statements phrased as preferences
> ("LLM-first, extract heuristics later") are design intent, not
> built behavior. Read this as a specification for future work.

## The problem

Hosts with perfect recall feel as alien as hosts with no memory. Humans remember selectively — recent events vividly, older events as gist, some forgotten entirely, some reinforced through repetition. Memories transform over time, becoming summaries of themselves, sometimes inaccurate.

We want host memory to have these properties: **selective retention, importance weighting, reinforcement on encounter, hierarchical compression, cue-dependent retrieval.** The Stanford Generative Agents paper (Park et al., 2023) is the closest reference; we extend it with explicit fungibility (compression stages) and tighter integration with the brain's tiered routing.

## Four categories, four lifecycles

| Category | Examples | Lifespan | Storage |
|---|---|---|---|
| **Working** | Last few minutes of events: who said what, where I walked, who's nearby | Seconds to minutes | In-memory ring buffer on cradle |
| **Episodic** | Discrete significant events: "I traded with JimBob", "I died at Greater Demons" | Hours to weeks; decay-based | Mesa: `episodes` table |
| **Relational** | Per-person facts: "JimBob is a reliable trader, 14 trades, 700k gp total" | Months to indefinite | Mesa: `relationships` table |
| **Reflective** | Generalizations: "I'm better at fishing than fighting", "Lumbridge is friendlier than Varrock" | Long-lived, revisable | Mesa: `reflections` table |

Plus a fifth category (procedural) that's the **routine library** in mesa — see [dsl.md](dsl.md). That's "remembered how-to" rather than "remembered events."

A second, lighter form of procedural/skill memory sits beside the routine library: the host's **earned verb vocabulary** (the mesa `bot_vocabulary` table — symbols that graduate on first successful use; see [mesa.md](mesa.md) and the progressive-disclosure / knowledge-gating design). Where the routine library is "what procedures I've assembled," the vocabulary is "what interface I've learned to use" — grown by experience, one verb at a time. Also Phase 3-4 design: `bot_vocabulary` is not built.

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

### The trust ledger — the concrete model behind `trust_score`

The per-other `trust_score` above is, concretely, a **Bayesian Beta(α,β) posterior** — it carries a **mean** (how reliable I think they are) *and* a **confidence** (how sure I am), so the host can distinguish "sure he's reliable, 200 trades" from "weak prior, met once." (See [mesa.md](mesa.md) relationships + the social-graph / trust-ledger design.)

- **The prior is shaped by the host's own persona.** At edge creation the prior is seeded from the host's Honesty-Humility H: `α₀ = 2 + 4H`, `β₀ = 2 + 4(1−H)`. High-H hosts extend benefit-of-the-doubt; low-H (scammer prototypes) start cynical. A fresh edge reads as "I just met this person" — middling mean, low confidence.
- **Updates are severity-weighted and LLM-graded.** Each cooperation/defection applies `α += w` or `β += w`. The classification *and* the weight `w ∈ [0.3, 8]` are an LLM determination — the mesa **TrustGrade** job (a Haiku DecisionClass) judges "cooperation or defection, and how severe given context?" and returns `{cooperative, w}`. (A coarse heuristic grade moves the local copy in-band instantly; the TrustGrade result is the authoritative correction.)
- **Updates fire only on an ATTRIBUTED signal.** A trade, chat, or duel carries a counterparty name, so it updates. A melee death has no attacker on the wire → the host **abstains** (no one to blame).
- **Decay is toward the H prior, not toward zero.** An un-reinforced relationship drifts back to "I trust them about as much as a stranger of my temperament," never to "enemy." The row **NEVER deletes** — it is structurally lossless (only the freetext `notes` compress); the counts and the episode persist as record.
- **Rivalries and friendships are DERIVED, not stored as truth.** A rivalry is flagged when `trust < 0.2` co-occurs with a high-severity defection — and it writes a vivid, high-salience episode (a rivalry is a *remembered event*, not just a low number). A friendship requires high trust *with real confidence* (a single fair trade reads as "acquaintance, seems fine," not "friend").

This is still Phase 3-4 design: the α/β columns, the structured tags, the `TrustGrade` DecisionClass, and the local-graph plumbing are all unbuilt.

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

### The scratch cache (working-memory sibling)

Working memory also includes a host-**authored** key→value scratch store — `cache.get` / `cache.set` / `cache.incr` — for rate-limit, dedup, and small flags (e.g. "asked JimBob about steel bars twice"). It is **local-fast read/write** in-band, with **async write-through to mesa `working_scratch`** (see [mesa.md](mesa.md)) so cognition can fold relevant entries into LLM prompts (the brain's reply can be informed by "I already asked Bob twice").

It is a **sibling** to the event ring, not the same thing:
- **vs. the event ring**: the ring is system-populated (events the host *experienced*); the scratch cache is host-populated (values the host's own routines *wrote* via `cache.set`). Both are local-fast and both mirror to mesa.
- **vs. episodic memory**: the cache is ephemeral *working state* (a counter, a flag), not a durable record of something that happened. Rate-limit keys like `asked:JimBob` are session-scoped and TTL'd — they should not outlive the goal; durable facts get promoted to a relational note/tag, never left as a cache key.

This is Phase 3-4 design: there is no `cache.*` verb surface in the DSL today and no `working_scratch` mesa table — both are unbuilt.

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
- **Cross-host relational consistency**: if HostA's record of HostB says "trusted friend" and HostB's record of HostA says "annoying acquaintance," is that fine (asymmetric relationships are realistic) or do we want some weak consistency mechanism?
