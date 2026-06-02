# Personas, cohorts, and north stars

> **STATUS: SCHEMA DESIGNED / EVERYTHING ELSE ASPIRATIONAL** (verified
> 2026-06-01). The `persona/` directory contains **no Go files** — there
> is **no persona code at all** (verified: empty directory). What has
> changed since the last revision: the **persona field SCHEMA** is now
> designed — that slice of the awaited paired Alex+Claude session is
> settled and landed below in "The persona schema (designed)". Nothing
> about it is implemented; it is the designed *output* of one part of the
> session, not built code. Everything *else* (the cohort taxonomy/count/
> distribution/assignment, the persona templates, the north-star catalog,
> the reverie catalog + weights, the LLM persona-generation pipeline, and
> the 500-host population mix) **remains deferred design** awaiting the
> rest of that session. Specifically, the **persona-tier default
> handlers** and `extends host` mechanism shown in the DSL examples do
> **not** exist in the current DSL: the shipped DSL (Phase 2.5) has `on`
> event handlers but no persona-level always-on reflexes (that is task
> `#93`, deferred to Phase 4 — the designed schema *un-parks* it; see the
> key tie (c) below). Read the schema section as "the designed field
> contract"; read everything below "What needs to be designed together"
> as "what we still intend to design together," not "what we have built."

## Status

**Partially designed; mostly still deferred.** This is a major paired design session between Alex and Claude that must happen before Phase 3 begins. Phase 0-2 can proceed without it; the brain and reveries cannot. **One slice of that session is now done:** the persona *field schema* — the immutable/mutable contract and the concrete fields — is designed and recorded in "The persona schema (designed)" below. It is a design artifact, not code: the `persona/` package is still empty. The rest of the session (how cohorts, templates, the population mix, the reverie catalog, and the generation pipeline *fill that schema in*) remains open, captured in "What needs to be designed together."

Alex's words from the architecture conversation:

> "This is where the reveries come into play and I'm going to work deeply with you on this — we will have to LLM generate, but we need to come up with cohorts, personas, 'north stars', etc. Please save this entire LOE as something we can pair on once we've finished answering all the questions, but the reality is it'll be LLM and hand authored as a pairing. I know we'll probably have to start with a handful, but the goal is we have a population of 500 hosts."

## The persona schema (designed)

This is the **one part of the awaited session that is now settled**: the field *schema* a persona is made of. It is a design artifact — **no persona code exists yet** (`persona/` is empty). What this section fixes is the *shape* every persona must have; how cohorts, templates, the population mix, the reverie catalog, and the LLM generation pipeline *fill this shape in* is still the open paired work below. (Full derivation, per-model verdicts, and update rules live in the `_research` notes: `quantitative-persona-models.md`, `host-bootstrap-and-knowledge-gating.md` §7/§7.5, `social-graph-and-trust-ledger.md`.)

The schema replaces the old folk `traits{social, risk, ambition, precision}` sketch, which conflated too much to carry the research — `ambition` alone smears together Conscientiousness-diligence, Schwartz Achievement, Schwartz Power, and time-preference, so the simulation could not distinguish a patient grinder from a Machiavellian power-seeker. The replacement is a three-layer ontology: **HEXACO** (*how* a host behaves), **Schwartz values** (*why*), and **behavioral-economic anchors** (the *decision params*).

### The immutable / mutable contract (Cornerstone vs Trajectory)

A persona is split into two layers with a hard discipline between them:

- **IMMUTABLE — the Cornerstone.** The deep trait *means*, identity, north star, voice, and the temperament anchors. Effectively fixed on the month horizon (rank-order trait stability is high in adults), sealed at genpop, and **never written by the runtime** (enforced by type discipline — the Cornerstone has no setters reachable from the host; operator-only edits go through `persona_revisions`). Crucially, the Cornerstone is **RE-INJECTED on every inference** — the immutable behavioral prose plus pinned foundational memories ride a cache-prefixed system block in every brain call. This is the one discipline the research adds on top of the split: Frisch & Giulianelli (2024) show identity drift dominates within *tens of turns* without re-injection, so coherence is not "set once at the prompt's start" — it is re-asserted continuously, cheaply, via prompt caching.
- **MUTABLE — the Trajectory.** A small in-RAM struct (flushed to mesa on a ~30s cadence) that carries **all** visible adaptation: mood, the drifting expression of traits, skill focus, sub-goals, the current temperament dials, and a pointer into the trust ledger. This layer is what carries the **differential-learning trajectory** the project studies (`research-goals.md` §1a) — "scammed once, now cautious" is a Trajectory drift, not a personality transplant.

### Field schema

```jsonc
// persona.Persona = { Cornerstone (immutable), Trajectory (mutable) }
// stored in mesa hosts.persona JSONB; compiled ONCE at birth (NO LLM) into the PersonaCard the runtime reads.
{
  "cornerstone": {                          // IMMUTABLE — re-injected every inference for identity coherence
    // --- identity (the re-injected prose block) ---
    "identity":   { "name": "...", "backstory": "<rendered prose>",
                    "voice": { "register": "...", "tics": ["...", "..."], "typo_rate": 0.04 } },
    "north_star": { "theme": "...", "statement": "...", "horizon": "month", "success_signals": ["..."] },

    // --- HEXACO disposition: 6 dimension means (μ) + bands. H is DEFINED here, USED by the trust ledger. ---
    "hexaco": { "H": { "mu": 0.84, "band": "high" },        // Honesty-Humility — the load-bearing ethics trait
                "E": { "mu": 0.51, "band": "mid" },         // Emotionality
                "X": { "mu": 0.61, "band": "mid_high" },    // eXtraversion
                "A": { "mu": 0.76, "band": "high" },        // Agreeableness
                "C": { "mu": 0.68, "band": "mid_high" },    // Conscientiousness
                "O": { "mu": 0.58, "band": "mid_high" } },  // Openness

    // --- Schwartz values: TOP-2 anchor + 4 higher-order weights (the ADAPTED form, NOT the full 10-value vector) ---
    "values": { "north_star_value": "benevolence", "secondary_value": "self_direction",
                "higher_order": { "self_transcendence": 0.42, "self_enhancement": 0.12,
                                  "openness_to_change": 0.28, "conservation": 0.18 } },

    // --- behavioral-econ anchors (deep; the live values drift in Trajectory) ---
    "prefs": { "patience_tau": 0.65,                        // long-horizon goal persistence anchor
               "loss_aversion_lambda": 2.10,               // PK / trade / flee risk anchor
               "coop_type": "conditional_cooperator",      // cond-cooperator | free-rider | altruist
               "risk": { "economic": "mid", "bodily": "low", "social": "mid_high" } },

    // --- temperament dials (bootstrap §7): an immutable ANCHOR here, a mutable CURRENT in Trajectory ---
    "attention_anchor": 0.7,                                // scalar 0..1 — short-horizon focus / resist-being-pulled
    "curiosity_anchor": { "social": 0.2, "spatial": 0.1, "skill": 0.6, "economic": 0.1, "risk": 0.0 }, // flavor vector

    "quirks":  [ /* 2-7 core quirks as executable modifiers */ ],
    "pinned":  [ /* 2-5 foundational memories, always re-injected alongside the Cornerstone prose */ ]
  },

  "trajectory": {                            // MUTABLE — all visible adaptation = the differential-learning trajectory
    "hexaco_state":   { "X": +0.03, "A": -0.01 },          // Fleeson density shift around μ (social-influence λ≈0.1, capped daily)
    "mood":           { "valence": +0.2, "arousal": 0.4, "stress": 0.1, "confidence": 0.6 }, // fast, lightweight; weights reveries
    "value_weights":  { "benevolence": 0.31, "self_direction": 0.22 },  // re-rankable lower tier (top-2 stays anchored)
    "sub_goals":      ["fish-lobsters-to-71", "teach-Jeren-cooking"],
    "skill_focus":    { "fishing": 2.1, "cooking": 1.4 },  // softmax LOGITS over activities (value-weighted reward)
    "patience_cur":   0.62,                                // τ — drifts around prefs.patience_tau
    "loss_aversion_cur": 2.32,                             // λ — drifts around prefs.loss_aversion_lambda
    "attention_cur":  0.66,                                // scalar — drifts around attention_anchor
    "curiosity_cur":  { "social": 0.2, "spatial": 0.1, "skill": 0.6, "economic": 0.1, "risk": 0.0 }, // drifts around curiosity_anchor
    "reverie_jitter": "<r_a from prng_seed>",              // per-host reverie jitter; weights are DERIVED, not stored
    "cooperation_type": "conditional_cooperator",          // the live read of coop_type (anchored in Cornerstone)
    "learned_emergent": null,                              // single promotable quirk slot earned through experience
    "trust_ledger":   "-> social.Graph Beta(α,β) per other",  // POINTER to the trust ledger (social-graph-and-trust-ledger.md)
    "friendships_rivalries": "DERIVED from trust_ledger"   // never stored as primary state — computed from the ledger
  }
}
```

### Per-model verdicts (faithful to `quantitative-persona-models.md`)

| Model | Verdict | What we keep |
|---|---|---|
| Immutable/mutable contract (re-inject identity every inference) | **ADOPT** | The spine. The Cornerstone/Trajectory split + the Frisch & Giulianelli re-injection discipline. |
| HEXACO × Schwartz × behavioral-econ ontology + mixture sampling | **ADOPT** schema / **ADAPT** sampling | The three-layer field structure. Sampling is cohort-conditioned priors sampled OFFLINE at host birth (NOT a full Gaussian copula) — forbidden-pair rejection captures the one thing the copula buys. |
| Patience τ + loss-aversion λ | **ADOPT** | τ = long-horizon goal persistence (pairs with attention = short-horizon focus); λ drives PK/trade/flee risk and feeds the trust ledger. Orthogonal dials, kept as separate fields. |
| Skill-focus softmax logits | **ADOPT/ADAPT** | The "what I like doing" that drifts with XP. ADAPT: the reward is **value-weighted** (a dot-product with the host's Schwartz weights) so identical-gear hosts diverge in week one. |
| Cooperation types | **ADAPT** | One sampled categorical (`cond-cooperator | free-rider | altruist`) that biases the trust-ledger prior + trade decisions; the full Fehr-Schmidt reciprocity-parameter soup is dropped. |
| Mood as Fleeson density-distribution | **ADAPT** | Kept **lightweight** — small capped HEXACO-state shifts + a fast valence/arousal mood, reconciled with the cradle-side mood that already weights reveries. NOT a new affect plane. |
| Reputation propagation through chat (gossip) | **DEFER** | Value-add but heavier; needs the live trust graph + chat content classification first. Phase-later. |

**Trims (where the source models over-reach):** the full **10-value Schwartz vector** is cut to a **top-2 anchor + 4 higher-order weights** (the runtime only needs the anchor + the higher-order simplex); the full prospect-theory evaluator (α/β/γ diminishing-sensitivity + probability weighting) is dropped in favor of **λ only**; the mixture copula + SBERT dedupe + population-entropy gate are replaced by cohort-conditioned priors + forbidden-pair rejection.

### The key ties

- **(a) `H` (Honesty-Humility) is load-bearing twice over.** It sets the trust-ledger **prior** — `α₀ = 2 + 4·H`, `β₀ = 2 + 4·(1−H)` (see `social-graph-and-trust-ledger.md`), so a high-H host starts generous/trusting and the scammer prototypes (μ_H ≈ 0.12) start cynical — *and* it is the best-validated predictor of the **ethics behaviors the project studies**, which is exactly why it is a first-class immutable field rather than buried in a folk `trustfulness` key. `H` is defined here; the trust ledger (the `trust_ledger` pointer in Trajectory) consumes it.
- **(b) attention and curiosity each carry an immutable ANCHOR + a mutable CURRENT.** `attention_anchor` / `curiosity_anchor` live in the Cornerstone (the baseline disposition); `attention_cur` / `curiosity_cur` live in Trajectory (the experience-drifted value). Same values and shapes as bootstrap §7 — split to fit the immutable/mutable contract.
- **(c) the schema un-parks the persona-level default handlers (§2a) and `extends host` (#93).** The "persona-level default handlers" below are the **base layer of the planned 3-tier handler stack** (persona base reflex → learned specialization → routine override). The designed Cornerstone (HEXACO, coop_type, λ, voice) is what those base reflexes read to differentiate; having a settled schema for them is what makes the deferred `extends host` / `super()` mechanism (task `#93`) buildable rather than hand-wavy.
- **(d) the mutable Trajectory layer drifting IS persona evolution (§2b).** "Persona evolution" below is not a separate mechanism — it is the Trajectory drifting (mood, skill logits, value re-ranks, the `learned_emergent` quirk slot) and, on slow re-ranks, snapshotting into mesa's `persona_revisions`. Fast mood churn never creates a revision; a value re-rank or a new emergent quirk does.

## What needs to be designed together

> The field schema above is the **designed** part. The items in this section are how that schema gets *filled in* — the cohort taxonomy, the templates, the population mix, the reverie catalog, and the generation pipeline. These **remain the open paired session** and are not settled by the schema work.

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

> **Still OPEN; now schema-anchored.** *What* a template is (a hand-authored archetype with dimensions of variation) and *how many* we author is open paired work. But the dimensions it varies are now the **designed schema fields** above — the folk dimensions in the list below (social temperament / risk tolerance / ambition profile) are superseded by the three-layer schema (HEXACO bands, Schwartz top-2 + higher-order, the behavioral-econ anchors τ/λ/coop_type/domain-split risk, plus the attention scalar + curiosity flavor vector). The list below is preserved as the original design intent; read it through the schema lens.

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
