# Personas, cohorts, and north stars

## Status

**Deferred design topic.** This is a major paired design session between Alex and Claude that must happen before Phase 3 begins. Phase 0-2 can proceed without it; the brain and reveries cannot.

Alex's words from the architecture conversation:

> "This is where the reveries come into play and I'm going to work deeply with you on this — we will have to LLM generate, but we need to come up with cohorts, personas, 'north stars', etc. Please save this entire LOE as something we can pair on once we've finished answering all the questions, but the reality is it'll be LLM and hand authored as a pairing. I know we'll probably have to start with a handful, but the goal is we have a population of 500 hosts."

## What needs to be designed together

### 1. Cohorts

Clusters of hosts that share core traits but vary individually. Cohort design is the *taxonomy* of the population. Hand-authored examples to start with:

- **Lumbridge regulars**: chatty, social, low-skill, friendly to newcomers
- **Edgeville merchants**: transactional, calculating, banking-focused, tolerant of risk
- **Varrock loiterers**: ambient socializers, idle frequently, light-skill diversity
- **Wilderness PKers**: aggressive, suspicious, high-combat focus
- **Solo grinders**: focused, less social, methodical
- **Newbie tourists**: chaotic, exploratory, occasionally helpless

Open questions for the design session:
- How many cohorts at population scale (10? 20? 50?)
- What's the distribution across cohorts? (skewed — most agents in "normal" cohorts, a few in extreme ones)
- Are cohorts persistent or can hosts transition between cohorts over time?
- How are cohorts assigned at registration?

### 2. Persona templates

Hand-authored archetypes within each cohort. Each template specifies the dimensions of variation, then per-agent LLM-generated content fills them in. Dimensions probably include:

- Cohort
- Approximate age (affects vocabulary, references, attention span)
- Vocabulary style (formal/casual/text-speak/era-appropriate slang)
- Vocal tics (catchphrases, fillers, common typos)
- Social temperament (shy / outgoing / antagonistic / helpful)
- Risk tolerance (cautious / reckless / calculating)
- Ambition profile (lazy / ambitious / fanatical)
- Daily activity preferences
- North star (long-term goal — see below)
- Reverie weights (per-persona overrides of base reverie frequencies)
- **Default event handlers** (see "Persona-level default handlers" below)

5-10 templates × LLM-generated variation = hundreds of distinct personas. The variation is one-time at registration (cached in mesa); the persona is then *revisable* — when the agent's lived experience materially shifts its outlook (gets scammed repeatedly → develops a warier `on chat_received`), the persona advances via a revision tracked in mesa's `persona_revisions` table.

### 2a. Persona-level default handlers

The DSL has a two-tier handler model (see [dsl.md](dsl.md)). Each persona declares a default handler for *every* supported event — these are the host's always-on reflexes that fire regardless of what routine is running. Routines temporarily override them, then defaults resume.

Every persona must define handlers for the full event vocabulary (currently 15 events). This is the place where "must declare every hook" applies — it forces the persona designer (LLM-generated, then optionally hand-curated) to explicitly consider every reactive surface.

Examples by persona archetype:

```
# Persona: "cautious lumby fisher"
on hp_below(0.3) {
    food = inventory.first_food()
    if food { eat(food) }
}
on attacked_by(other) {
    if other.combat_level > self.combat_level + 2 {
        # Cautious — flee from anything above us
        walk_to(nearest_safe_spot())
    }
}
on chat_received(speaker, message) {
    # Friendly default — respond to most messages
    if relation_with(speaker.name) {
        chat.queue(f"hi {speaker.name}")
    } elif message.contains("hi") or message.contains("hey") {
        # Sometimes respond to strangers, modulated by mood
        if mood().anxious < 0.5 {
            chat.queue("hi")
        }
    }
}
on level_up(skill) {
    chat.queue(f"yay {skill} up!")
}
on death() {
    note("died — was it worth it?")
}
# ... all other events explicitly declared
```

```
# Persona: "reckless wilderness pker"
on hp_below(0.3) {
    # Eat but don't flee
    food = inventory.first_food()
    if food { eat(food) }
}
on attacked_by(other) {
    # ALWAYS engage — different from cautious
    if not combat.engaged {
        attack(other)
    }
}
on chat_received(speaker, message) {
    # Generally ignore strangers
    if relation_with(speaker.name) and relation_with(speaker.name).trust_score > 0.5 {
        chat.queue(f"sup {speaker.name}")
    }
}
on level_up(skill) {
    if skill in ["attack", "strength", "defence", "hits"] {
        chat.queue("nice")
    }
    # Skill-up indifferent for non-combat skills
}
on death() {
    note("died — find them again later")
    # Probably: persona's relational record for the killer updates to "rival"
}
# ... all other events explicitly declared
```

Note how the same event names get materially different default behaviors. This is the core mechanism for persona-driven differentiation in reactive behavior. Routines override specific handlers when situation-specific behavior is needed; everything else is the persona's reflex.

### 2b. Persona evolution

Personas are not frozen. The agent's lived experience can revise them over time. Examples:

- A "trusting cautious" persona that gets scammed three times revises its `on chat_received` to be warier with strangers
- An "ambitious grinder" persona that builds a real friend group through chat slowly revises its routine library priorities to include more social activities
- A "reckless pker" persona that dies repeatedly to a specific player develops a relational record for that player and a routine-specific avoidance reflex

Revisions are LLM-generated (Sonnet, rare — once every few days at most) and stored in mesa's `persona_revisions` table with a rationale. The current persona is always the latest revision; history is preserved for research analysis.

### 3. North stars

Every host has a long-term goal that motivates its strategic decisions. Examples:

- "Become wealthy through trading" (transactional)
- "Become the best fisher on the server" (skill-focused, narrow)
- "Build a close friend group" (social-focused)
- "Explore the whole map" (exploratory, lower commitment to any single thing)
- "Become known as a generous person" (reputation-driven)
- "PK the most other players" (combat-focused, antagonistic)
- "Master every skill" (broad ambition)

North stars are NOT optimization targets — they're psychological motivators. A host with "become wealthy" north star will still take time to chat with friends, just like a real player. But over weeks, their strategic choices accumulate in a direction.

Design questions:
- How specific should north stars be? Too specific = robotic ("level fishing to 50"). Too vague = unactionable ("be happy").
- Can north stars evolve? A host who gets scammed early might develop a "become the most distrustful trader" north star that wasn't in the original template.
- How are conflicting drives resolved? Persona has "wealth" north star but also "shy" social temperament — agent should still do social things, just less than an outgoing-wealth-driven persona.

### 4. Reverie catalog (cross-references [reveries.md](reveries.md))

The full catalog of reveries that exist in the system, with default weights per cohort/persona. Probably 15-20 distinct reveries. Each gets per-persona weighting.

Designing this means deciding: what specific small behaviors should hosts exhibit? Glance? Wave? Wander? Sigh-typing-pause? Each adds believability if implemented well.

### 5. LLM persona-generation pipeline

The mechanism for "given a cohort + template + seed, produce a unique persona." Probably:

1. One-time Sonnet call at host registration
2. Inputs: cohort spec, template spec, a random seed, a list of names not yet used
3. Outputs: full persona JSON (name, age, backstory paragraph, vocabulary samples, north star statement, key relationships if any, etc.)
4. Result cached in mesa's `bots.persona` JSONB column

This is the design choice of what's randomized vs deterministic. Names + appearances + small biographical details = LLM. Cohort + template + general traits = fixed at registration.

### 6. Population mix targeting

For 500 hosts, what cohort/persona distribution produces interesting social dynamics?

Hypothesis: skew toward "casual normal" cohorts with a long tail of extremes:
- ~60% "casual social/skill-focused" (Lumbridge regulars, Varrock loiterers)
- ~20% "focused/transactional" (Edgeville merchants, solo grinders)
- ~10% "newbie/exploratory"
- ~10% "extreme" (PKers, fanatical grinders, oddballs)

This produces a Pareto-ish population. Most interactions are between "normal" hosts; "extreme" hosts add color and create observable incidents.

## Why this matters

The persona system is the **substrate** that makes the research goals possible:

- **Long-term strategic accomplishment**: hosts pursue north stars over weeks. Without north stars, hosts are reactive — never building toward anything.
- **Community formation**: hosts of compatible cohorts gravitate together; hosts of clashing cohorts produce friction. The cohort taxonomy is the source of social structure.
- **Ethics observation**: ethical choices correlate with persona traits. The "ambitious-reckless" cohort might steal more than "shy-cautious." Observable patterns require persona variation.
- **Believability**: a population of identical hosts is detectable. Persona variation creates the diversity that makes individual hosts believable as humans.

Get personas wrong → 500 indistinguishable agents. Get them right → a believable society.

## When this design session happens

Before Phase 3 (brain + cognition implementation). Phase 0-2 can proceed without it. The persona system gates Phase 3 because the brain prompt needs persona, the cognition layer needs cohort_id, and reveries need persona traits to do anything meaningful.

Estimated session length: a full working day with Alex, possibly across multiple sessions. The output is concrete enough to feed implementation: persona schema, cohort definitions, north star catalog, reverie weights table.

## Anti-patterns to avoid

- **Too detailed personas**: overspecifying makes hosts robotic ("you always do X"). Personas should constrain *style* and *priorities*, not specific actions.
- **Too consistent personas**: real humans have inconsistencies. Persona templates should include "occasional surprising trait" slots.
- **Cohort = role**: a cohort isn't a job assignment. Hosts of any cohort can do any RSC activity; cohort affects *how* and *how often*, not what's possible.
- **Stable forever**: personas should be revisable as evidence accumulates. A "trusting" persona that gets scammed three times should adapt.
