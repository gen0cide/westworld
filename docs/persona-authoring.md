# Persona authoring guide — the complete field reference

> **STATUS: REFERENCE FOR A DESIGNED-BUT-UNBUILT SCHEMA** (2026-06-06). This
> documents every field of the persona schema settled in [personas.md](personas.md)
> "The persona schema (designed)" and derived in
> [`_research/quantitative-persona-models.md`](_research/quantitative-persona-models.md)
> + [`_research/reference/decision-persona-schema.md`](_research/reference/decision-persona-schema.md).
> **No persona code exists yet** (`persona/` is empty). This is the field
> dictionary + authoring workflow — the thing a human reads when sitting down to
> craft (or curate) a persona, and the contract the sampler/compiler will honor.
> Where this guide and `personas.md` disagree, `personas.md` is canonical; ping
> back so we reconcile.

---

## 0. Read this first: who fills in what

A persona is **not** a form a human fills out field-by-field. The schema braids
**four authoring sources**, and knowing which source owns a field is the whole
game — it's the answer to "what does a human actually do?"

| Source | Owns | When | Example fields |
|---|---|---|---|
| **Human** | *Intent* — the high-level shape | design session | `cohort`, `archetype_tag`, `north_star.theme`, signature `quirks`, `voice` flavor, cohort priors + `forbidden_pairs` |
| **Offline sampler** (deterministic Go, seeded) | *The numbers* | host birth (genpop) | `hexaco.*.mu`, `values.*`, `prefs.*`, `risk.*`, `attention_anchor`, `curiosity_anchor`, derived `quirks`, `reverie.jitter` |
| **LLM** (1 batched call / ~50 hosts) | *Prose only* | host birth, after sampling | `identity.backstory`, `voice.tics`, per-quirk `narrative`, `north_star.statement` |
| **Runtime** (Limbic, cognitiveLoop, mesa) | *All adaptation* | continuously | the **entire** `trajectory` block + the trust ledger |

So the practical authoring loop is:

1. **Author a cohort** (a prior over traits + a weight + forbidden pairs). This is
   most of the leverage — you're shaping a *distribution*, not one host.
2. **Author a template/archetype** inside the cohort (north-star theme, voice
   register, 1–3 signature quirks, reverie leanings).
3. **Let the sampler draw** an individual host's numbers from the cohort prior
   (seeded, reproducible, forbidden-pairs rejected).
4. **Let the LLM render** the prose once, offline.
5. **Curate**: read the rendered host, nudge the cohort prior or archetype if the
   population reads flat or wrong, regenerate. You almost never hand-edit one
   host's floats — you fix the cohort.

> **The golden rule of authoring:** constrain *style* and *priorities*, never
> specific actions. "You distrust strangers in the wilderness" is a persona.
> "You always walk to Falador bank at 9am" is a script. The former produces a
> believable host; the latter produces a robot. (See §13 anti-patterns.)

The two-layer split underneath all of this:

- **Cornerstone (IMMUTABLE)** — who the host *is*. Sealed at birth, never written
  by the runtime, re-injected into every brain call so identity never drifts.
- **Trajectory (MUTABLE)** — how the host is *currently adapting*. The runtime
  owns it entirely; a human never authors it (you only author the *anchors* in
  the Cornerstone that the Trajectory drifts around).

---

## 1. The three-layer trait ontology (the mental model)

Every Cornerstone is built from three orthogonal layers. Internalize this and the
field list stops feeling arbitrary:

| Layer | Question it answers | Fields | Analogy |
|---|---|---|---|
| **HEXACO** | *HOW* does this host behave? | `hexaco` (6 traits) | personality / temperament |
| **Schwartz values** | *WHY* — what does it care about? | `values` | motivation / what's sacred |
| **Behavioral-econ anchors** | *With what decision params?* | `prefs`, `risk` | the dials on the decision machine |

These are non-redundant by design (research: trait facets explain ≤28% of value
variance — knowing *how* someone acts doesn't tell you *why*). The old folk schema
`traits{social, risk, ambition, precision}` was killed precisely because
`ambition` smeared Conscientiousness, Achievement, Power, and time-preference into
one number — you couldn't tell a patient grinder from a power-seeker. The three
layers keep those separable.

On top of the three layers sit two **temperament dials** (attention + curiosity,
§6) and the **identity prose** (name, backstory, voice, north star, quirks, pinned
memories).

---

## 2. The Cornerstone at a glance (immutable)

```jsonc
"cornerstone": {
  "identity":   { name, backstory, voice{...}, archetype_tag },
  "north_star": { theme, statement, horizon, success_signals[] },
  "hexaco":     { H, E, X, A, C, O } each = { mu: 0..1, band },
  "values":     { north_star_value, secondary_value, higher_order{4 weights} },
  "prefs":      { patience_tau, loss_aversion_lambda, coop_type, risk{econ,bodily,social} },
  "attention_anchor":  0..1,
  "curiosity_anchor":  { social, spatial, skill, economic, risk },
  "quirks":     [ /* 2–7 executable modifiers */ ],
  "reverie":    { jitter_seed },
  "pinned":     [ /* 2–5 foundational memories */ ]
}
```

Each subsection below documents **type · range/options · default · authored-by ·
consumed-by · example · authoring notes** for every field.

---

## 3. `identity` — name, backstory, voice, archetype

### `identity.name`
- **Type:** string. **Authored by:** sampler (picks from an unused-names pool).
- **Consumed by:** everything — it's the partition key (`== opts.Username`) and the
  only durable identity on the wire. The trust ledger keys on it.
- **Authoring note:** the human authors the *name pool / style* per cohort
  (era-appropriate RSC names, clan tags, l33t-speak handles), not individual names.

### `identity.backstory`
- **Type:** string (rendered prose paragraph). **Authored by:** LLM at birth.
- **Consumed by:** the brain prompt — this is part of the re-injected immutable
  block. It is *prose only*; the brain never sees raw floats.
- **Authoring note:** the human authors the *backstory seed* in the template (a
  one-line premise: "ex-merchant who got scammed and turned cautious"); the LLM
  expands it into prose consistent with the sampled HEXACO/values. Keep it short —
  it rides every inference, so it's a cache-cost line item.

### `identity.archetype_tag`
- **Type:** string (e.g. `"helpful_mentor"`, `"wilderness_pker"`).
- **Range/options:** free, but should match a cohort archetype name.
- **Authored by:** human (it's the label of the template you chose).
- **Consumed by:** **debug/analytics ONLY — INVISIBLE to the brain.** This is a
  deliberate guard: if the brain saw `"scammer"` it would role-play the label
  instead of behaving from traits. Tags are for delos dashboards, not cognition.

### `identity.voice` — how the host *talks*

```jsonc
"voice": { "register": "casual", "formality": "casual",
           "tics": ["ngl", "fr", "..."], "typo_rate": 0.04 }
```

| Field | Type | Range / options | Authored by | Notes |
|---|---|---|---|---|
| `register` | string | `formal` \| `casual` \| `terse` \| `playful` \| `gruff` \| era-slang labels | human (flavor) → LLM (samples) | the overall speaking style |
| `formality` | string | `formal` \| `neutral` \| `casual` \| `text_speak` | human | drives capitalization, punctuation, abbreviation |
| `tics` | string[] | 0–5 catchphrases / fillers / verbal habits | LLM, human-curated | "ngl", "based", "cya", a signature emote |
| `typo_rate` | float | `0.0`–`~0.12` | sampler (corr. w/ C: low-C → higher) | fraction of messages with a plausible typo; **believability dial** — zero typos reads as a bot |

- **Consumed by:** chat generation (the async Phase-2 reply) + the prose card.
- **Authoring note:** voice is the single highest-leverage *believability* field
  because chat is the most visible host output. Vary it hard across cohorts. A
  Wilderness PKer is terse + high-typo + clan-slang; a mentor is patient +
  low-typo + full sentences.

---

## 4. `hexaco` — the six personality traits (the *how*)

Each of the six is `{ "mu": 0.0–1.0, "band": <band> }`. `mu` is the frozen trait
mean (sampled); `band` is the coarse bucket the runtime keys on (the brain sees
*neither* number nor band — it sees the rendered prose).

**Band vocabulary** (ordered, derived from `mu` percentiles):
`very_low` < `low` < `mid` < `mid_high` < `high` < `very_high`.

> The runtime reads **bands**, not floats (`persona.Facet("diligence") → band`).
> Numbers exist only for analytics + cohort science in delos.

| Code | Trait | High end (`mu`→1) | Low end (`mu`→0) | RSC behaviors it drives |
|---|---|---|---|---|
| **H** | **Honesty-Humility** | fair, sincere, no greed, no manipulation | exploitative, entitled, will scam | **load-bearing — see below.** Trade honesty, the trust-ledger *prior*, whether the host scams |
| **E** | **Emotionality** | anxious, fearful, sentimental, flees | brave, tough, unsentimental | the **flee threshold**, panic-eating, attachment to friends |
| **X** | **eXtraversion** | chatty, initiates, seeks crowds | quiet, solitary, ignores chat | chat-initiation rate, banking-area lingering, party-seeking |
| **A** | **Agreeableness** | forgiving, patient, lenient | critical, holds grudges, retaliates | retaliation vs forgiveness, response to insults, grudge persistence |
| **C** | **Conscientiousness** | diligent, organized, persistent grinder | impulsive, sloppy, distractible | **grind persistence (pairs w/ τ)**, inventory discipline, low typo rate, low exploration temperature |
| **O** | **Openness** | curious, explores, tries new skills | conventional, routine, repetitive | map exploration, skill-diversity, the curiosity-spatial flavor |

### Why `H` is special (read this twice)

`H` is the **only HEXACO trait that's load-bearing twice over**, which is why it's
a first-class field and not buried in a folk `trustfulness` key:

1. **It sets the trust-ledger prior.** Every relationship starts at
   `Trust = Beta(α₀, β₀)` where **`α₀ = 2 + 4·H`** and **`β₀ = 2 + 4·(1−H)`**. So a
   high-H host (`mu≈0.84`) starts generous and assumes others are trustworthy; a
   scammer prototype (`mu_H≈0.12`) starts cynical. (Math owned by
   [`_research/social-graph-and-trust-ledger.md`](_research/social-graph-and-trust-ledger.md);
   `H` is *defined* in the persona, *consumed* by the ledger.)
2. **It's the best-validated predictor of the ethics behaviors the project
   studies.** The whole research goal of observing emergent ethics rides on H
   being separable and first-class.

**Authoring note:** you almost never set `mu` by hand. You set the **cohort
prior** (mentors high-H, scammers low-H) and a **forbidden pair** like
`["H:high", "value:power"]` so the sampler can't draw a psychologically-impossible
saintly power-seeker. The sampler draws each host's `mu` from the cohort's prior;
forbidden-pair rejection resamples the impossible ones.

### Facets the runtime keys on directly
The compiler can read sub-trait facets (not just the 6 means):
`H-Fairness` (trade honesty), `E-Fearfulness` (flee threshold), `C-Diligence`
(grind persistence), `O-Inquisitiveness` (curiosity-spatial flavor). These are
derived from the parent trait + a small offset, not separate authored fields in
v1.

---

## 5. `values` — Schwartz values (the *why*)

```jsonc
"values": {
  "north_star_value": "benevolence",
  "secondary_value":  "self_direction",
  "higher_order": { "self_transcendence": 0.42, "self_enhancement": 0.12,
                    "openness_to_change": 0.28, "conservation": 0.18 }
}
```

We use a **trimmed** Schwartz model: a **top-2 categorical anchor** + **4
higher-order weights** (NOT the full ranked-10 vector — the runtime only needs the
anchor + the simplex; the full 10 is a delos-analytics nice-to-have).

### `north_star_value` / `secondary_value`
- **Type:** string, one of the **Schwartz 10**:
  `self_direction`, `stimulation`, `hedonism`, `achievement`, `power`,
  `security`, `conformity`, `tradition`, `benevolence`, `universalism`.
- **Authored by:** sampler (from cohort prior); human authors the cohort's value
  leanings.
- **Consumed by:** north-star alignment + the **value-weighted skill reward** (§
  skill_logits) — this is what makes two hosts with identical starting gear
  diverge in week one. A Power host gets reward from wealth deltas; a Benevolence
  host from *helping* outcomes.

### `higher_order` (4 weights, sum ≈ 1.0)
The Schwartz-10 collapse into 4 higher-order quadrants (two opposing axes):

| Higher-order | Contains | Opposes |
|---|---|---|
| `self_transcendence` | benevolence, universalism | self_enhancement |
| `self_enhancement` | achievement, power, (hedonism) | self_transcendence |
| `openness_to_change` | self_direction, stimulation, (hedonism) | conservation |
| `conservation` | security, conformity, tradition | openness_to_change |

- **Type:** 4 floats on a simplex (sum to 1). **Range:** each `0.0–1.0`.
- **Authoring note:** the two axes are **opposed** — high self-transcendence
  implies low self-enhancement. The sampler enforces this; if you hand-author,
  keep opposing pairs anti-correlated or the host reads incoherent.

---

## 6. `prefs` + temperament dials — the decision parameters

### `prefs.patience_tau` (τ)
- **Type:** float `0.0–1.0`. **Default anchor example:** `0.65`. **Authored by:** sampler.
- **Horizon:** **LONG (days/weeks).** Whether a host *stays the course* toward a
  north-star sub-goal across sessions.
- **Consumed by:** the deliberation ladder — high τ biases "continue the current
  sub-goal" over "re-strategize" (and resolves it locally, no LLM call).
- **High τ:** sits at lobsters for 100 hours to hit 99 fishing. **Low τ:** abandons
  the long plan, goal-hops weekly.

### `prefs.loss_aversion_lambda` (λ)
- **Type:** float, **`~1.0–3.0`**. **Default anchor example:** `2.10` (1.0 = no
  loss aversion; ~2.0 ≈ human-typical; 3.0 = very loss-averse). **Authored by:** sampler.
- **Consumed by:** flee / trade-confirm / PK risk decisions, and it **scales the
  grievance deltas** the trust ledger records (a high-λ host takes a lopsided
  trade *harder*).
- **High λ:** flees early, refuses risky trades, holds losses as grudges.
  **Low λ:** dives into the wilderness, gambles trades.

> **τ vs the attention dial — DO NOT merge them.** They're orthogonal:
> | Dial | Horizon | Governs |
> |---|---|---|
> | **attention** | short (minutes) | how well the host *resists being pulled* off the current activity (interrupt/reverie/switch threshold) |
> | **patience τ** | long (days/weeks) | whether the host *stays the course* toward a north star |
>
> A high-attention / low-τ host hyperfocuses on whatever it's doing *right now* but
> abandons the long plan (ADHD-grind). A low-attention / high-τ host is twitchy
> moment-to-moment but never gives up the 99-fishing dream.

### `prefs.coop_type`
- **Type:** string, one of: `conditional_cooperator` | `free_rider` | `altruist`.
- **Default mix (population):** ≈50% conditional, ≈30% free-rider, rest altruist.
- **Authored by:** sampler (cohort-skewed: PKers/scammers → free-rider;
  mentors/clan → conditional/altruist).
- **Consumed by:** (1) the trust-ledger TrustGrade prior — a free-rider reads a
  neutral act as exploitable sooner; a conditional-cooperator mirrors the
  partner's last move. (2) trade/help decisions — an altruist biases "help the
  newbie," a free-rider biases "take the lopsided-in-my-favor trade."

### `prefs.risk` — domain-split (never one scalar)
```jsonc
"risk": { "economic": "mid", "bodily": "low", "social": "mid_high" }
```
- **Type:** 3 ordered bands (`very_low`..`very_high`). **Authored by:** sampler.
- **Why split:** a cautious trader can be a wilderness daredevil. One `risk_tolerance`
  scalar can't express that. The three axes:
  - `economic` — willingness to make risky trades/gambles/stake duels.
  - `bodily` — willingness to enter dangerous areas / fight stronger players.
  - `social` — willingness to approach strangers, risk rejection, be antagonistic.

### `attention_anchor`
- **Type:** scalar float `0.0–1.0`. **Example:** `0.7`. **Authored by:** sampler.
- **Meaning:** short-horizon focus / **resistance to being pulled**. Low =
  distractible (ADHD), high = hyperfocus. Pairs with `prefs.patience_tau` (long
  horizon) — see the table above.
- **Consumed by:** the interrupt ladder + reverie-fire + explore-switch thresholds
  (`host-bootstrap §7.3`; chat-ladder edges).

### `curiosity_anchor` — the flavor vector
```jsonc
"curiosity_anchor": { "social": 0.2, "spatial": 0.1, "skill": 0.6, "economic": 0.1, "risk": 0.0 }
```
- **Type:** 5 floats (a *flavor mix*, not a probability — they needn't sum to 1,
  though the example does). **Authored by:** sampler.
- **Meaning:** *WHAT pulls this host* (whereas attention is *how hard it resists
  being pulled*). The five flavors:
  - `social` — drawn to people, conversations, crowds.
  - `spatial` — drawn to unexplored map / new areas.
  - `skill` — drawn to trying/leveling new skills.
  - `economic` — drawn to markets, deals, arbitrage.
  - `risk` — drawn to danger, the wilderness, PvP.
- **Consumed by:** the differential-learning engine — curiosity is *what makes
  hosts diverge*. It biases which goals the host volunteers for and where it wanders.
- **Authoring note:** this is one of the most *expressive* author-facing knobs.
  A high-`spatial`/high-`O` host explores the whole map; a high-`skill`/high-`C`
  host min-maxes one skill tree.

---

## 7. `quirks` — executable modifiers (NOT flavor text)

This is the field most people get wrong. A quirk is **not** a string like
"likes fishing." It's a **structured, executable modifier** the compiler turns
into a behavior bias the runtime applies with **no LLM call**. 2–7 per host.

```jsonc
{
  "id": "falador_west_loyalist",
  "origin": "derived",
  "domain": "banking",
  "trigger": "pre_action:bank",
  "binding": "bank_node_preference",
  "relation": "prefers",
  "object": "bank_node:Falador_west",
  "strength": 0.7,
  "observable": true,
  "suppress_when": "in_combat",
  "narrative": "Always banks at the west Falador bank, even when it's out of the way."
}
```

| Field | Type | Range / options | Meaning |
|---|---|---|---|
| `id` | string | unique slug | handle for logs/promotion |
| `origin` | string | `derived` \| `idiosyncratic` \| `learned_emergent` | derived = from traits; idiosyncratic = authored flavor; learned_emergent = earned at runtime (the single mutable slot) |
| `domain` | string | `movement` \| `social` \| `trade` \| `combat` \| `banking` \| `idle` | which behavior surface it touches |
| `trigger` | string | an `event.Kind()` (e.g. `trade_request`, `chat_received`) **or** `on_encounter` \| `pre_action:<verb>` | when it fires |
| `binding` | string | a gameplay-binding enum | **rejects non-operational quirks** — a quirk must bind to a real game mechanic |
| `relation` | string | `prefers` \| `avoids` \| `distrusts` \| `delays` | the direction of the bias |
| `object` | string | a referent: `bank_node:Falador_west`, `equipment:bronze_full`, `player_type:newbie` | what the bias is *about* |
| `strength` | float | `0.0–1.0` | intensity of the bias |
| `observable` | bool | true/false | **must be true** — a quirk the server can't see is rejected (it can't affect believability) |
| `suppress_when` | string | e.g. `in_combat`, `low_hp` | a context that disables the quirk (load-bearing — a banking quirk must not fire mid-fight) |
| `narrative` | string | LLM prose | for the prose card only; the runtime ignores it |

**Authoring rules:**
- **Must be observable.** If the server/other-players can't perceive it, it's not a
  quirk, it's a daydream. Reject `camera_pan`-style UI-only quirks.
- **Must bind to a mechanic.** "Is nostalgic" is not a quirk. "`avoids`
  `area:Lumbridge` (`strength` 0.6) because of a bad memory" is.
- **Must declare `suppress_when`** if it could fire at a dangerous moment.
- **2–7 total.** Fewer → host reads generic. More → host reads over-specified /
  robotic. The sampler *derives* most from traits; the human hand-adds 1–3
  *signature* idiosyncratic ones per archetype.
- **One `learned_emergent` slot only** — that's the runtime's, earned through
  experience (capped, logged on every promotion). Humans don't author it.

---

## 8. `north_star` — the long-term motivator

```jsonc
"north_star": { "theme": "wealth", "statement": "Become the richest merchant in Varrock",
                "horizon": "month", "success_signals": ["gp_total > 1M", "owns_full_rune"] }
```

| Field | Type | Range / options | Authored by |
|---|---|---|---|
| `theme` | string | `wealth` \| `skill_mastery` \| `social` \| `exploration` \| `reputation` \| `combat` \| `broad_ambition` | human |
| `statement` | string | one sentence | LLM (from theme) |
| `horizon` | string | `week` \| `month` \| `open` | human |
| `success_signals` | string[] | measurable from world+mesa | human/LLM |

- **Consumed by:** `cognition.Retrieval.Goal` + the deliberation ladder — it biases
  *strategic* choices over weeks.
- **Authoring note (critical):** a north star is a **psychological motivator, NOT
  an optimization target.** A "become wealthy" host still stops to chat with
  friends, just like a real player — it just *trends* toward wealth over weeks.
  Pitch the specificity in the middle:
  - **Too specific** = robotic: "level fishing to exactly 50."
  - **Too vague** = unactionable: "be happy."
  - **Right:** "become known as the most generous person in Lumbridge."

---

## 9. `pinned` — foundational memories

```jsonc
"pinned": [ { "summary": "Was scammed out of my first rune set by a fake-trade.", "weight": 0.9 } ]
```
- **Type:** array of `{ summary: string, weight: 0.0–1.0 }`, **2–5 entries**.
- **Authored by:** human/LLM at birth (these are `episodes(stage=0)` — never decayed).
- **Consumed by:** re-injected into **every** brain call alongside the Cornerstone
  prose (`Bundle.Pinned`). This is the gemini "a host never forgets who it is."
- **Authoring note:** these are the 2–5 memories that *define* the host's outlook —
  the formative experiences the backstory references. Keep them few; every pinned
  memory rides every inference (cache cost). They should *explain* the HEXACO/values
  the sampler drew (a scam memory explains low trust / high economic-caution).

---

## 10. `reverie` (jitter) + reverie weights

```jsonc
"reverie": { "jitter_seed": "<r_a derived from prng_seed>" }
```
- **Reverie weights are NOT stored.** They're computed at runtime from traits:
  `weight = base + Σcoeff·trait + Σcoeff·state, × per-host jitter r_a`.
  Only the per-host **jitter** (`r_a ~ LogNormal(0, 0.25)`, sampled once) is stored,
  so two hosts with identical traits still differ slightly.
- **Authored by:** sampler (the seed). The **catalog** of reveries (glance, wave,
  wander, equipment-hover, typing-pause…) lives in code and is **deferred design**
  (the paired session, see [reveries.md](reveries.md)).
- **Example weight closure:** `examine_scenery_w = 0.05 + 0.25·O + 0.10·(1−C) + 0.20·novelty, × r_a`
  — so high-Openness, low-Conscientiousness hosts examine scenery more; stress
  lifts anxious reveries; recent damage lifts equipment-hover.
- **Authoring note:** a human doesn't set reverie weights per host. You author the
  *trait coefficients* in the catalog (design-session work) and the *cohort*; the
  individual weighting falls out of the host's traits + jitter.

---

## 11. The Trajectory (mutable) — what the runtime owns

**A human never authors any of this.** It's documented here so you understand what
the Cornerstone anchors *drift into*. The runtime (Limbic + cognitiveLoop)
writes it; it flushes to mesa on a ~30s cadence; slow re-ranks snapshot into
`persona_revisions`.

```jsonc
"trajectory": {
  "hexaco_state":  { "X": +0.03, "A": -0.01 },   // capped daily drift of EXPRESSION around frozen μ
  "mood":          { "valence": +0.2, "arousal": 0.4, "stress": 0.1, "confidence": 0.6 },
  "value_weights": { "benevolence": 0.31, "self_direction": 0.22 }, // re-rankable lower tier (top-2 stays anchored)
  "sub_goals":     ["fish-lobsters-to-71", "teach-Jeren-cooking"],
  "skill_logits":  { "fishing": 2.1, "cooking": 1.4 },  // soft-RL focus, value-weighted reward
  "patience_cur":  0.62,                          // τ drifting around prefs.patience_tau
  "lambda_cur":    2.32,                          // λ drifting around prefs.loss_aversion_lambda
  "attention_cur": 0.66,                          // drifting around attention_anchor
  "curiosity_cur": { "social": 0.2, ... },        // drifting around curiosity_anchor
  "learned_emergent": null,                        // the single earned quirk slot
  "reverie_jitter": "<r_a>",
  "cooperation_type": "conditional_cooperator"     // live read of coop_type
  // trust_ledger + friendships/rivalries: OWNED BY social.Graph, DERIVED, never stored here
}
```

| Field | Type / range | Written by | Anchor it drifts around |
|---|---|---|---|
| `hexaco_state` | small capped Δ per trait, daily | Limbic (social influence λ≈0.1) | the frozen `hexaco.*.mu` |
| `mood.valence` | `-1..1` | Limbic (death/damage↓, xp/item↑), lazy decay | persona baseline |
| `mood.arousal` | `0..1` | Limbic | baseline |
| `mood.stress` | `0..1` | Limbic | baseline |
| `mood.confidence` | `0..1` | Limbic | baseline |
| `skill_logits` | `map[skill]float`, softmax temp = `1 − 0.5·C` | soft-RL on XP, **value-weighted reward** | — |
| `patience_cur` / `lambda_cur` | scalars | drift rule `x += κ(g−x) − ρ(x−anchor)`, κ≈0.02, ρ≈0.05 | the `prefs.*` anchors |
| `attention_cur` / `curiosity_cur` | scalar / vector | drift | the `*_anchor` |
| `learned_emergent` | one `Quirk` or null | quirk-promotion detector (e.g. 12/14 trips), capped | — |
| trust ledger | `Beta(α,β)` per other | TrustGrade (LLM-graded, attribution-gated) | the H-shaped prior `α₀=2+4H` |

**Key facts for authors:**
- `mood` is **fast** and churns constantly — it NEVER creates a `persona_revision`.
  A value re-rank, new sub-goal, or emergent-quirk promotion DOES.
- `hexaco_state` shifts *expression*, never the sealed `mu`. It's capped so a host
  can never become a second personality. "Scammed once, now cautious" lives here —
  it's a drift, not a transplant.
- **friendships/rivalries are DERIVED** from the trust ledger, never stored. Don't
  author relationships as fields; author the *H* + *coop_type* that shape how they
  form.

---

## 12. Two worked examples (full personas)

### A — "Cautious Lumbridge fisher" (cohort: Lumbridge regulars)

```jsonc
{
  "cornerstone": {
    "identity": {
      "name": "marn_fishes", "archetype_tag": "cautious_social_grinder",
      "backstory": "Quiet regular who's spent months at the Lumbridge swamp. Got burned by a fake-trade once and never forgot it; happy to chat but won't risk much.",
      "voice": { "register": "casual", "formality": "casual", "tics": ["gl", "nice one"], "typo_rate": 0.05 }
    },
    "north_star": { "theme": "skill_mastery", "statement": "Hit 99 fishing without ever getting scammed again",
                    "horizon": "open", "success_signals": ["fishing_level == 99"] },
    "hexaco": { "H": {"mu":0.78,"band":"high"}, "E":{"mu":0.66,"band":"mid_high"},
                "X": {"mu":0.55,"band":"mid"}, "A":{"mu":0.72,"band":"high"},
                "C": {"mu":0.81,"band":"high"}, "O":{"mu":0.40,"band":"mid"} },
    "values": { "north_star_value": "security", "secondary_value": "achievement",
                "higher_order": {"self_transcendence":0.28,"self_enhancement":0.22,"openness_to_change":0.15,"conservation":0.35} },
    "prefs": { "patience_tau": 0.82, "loss_aversion_lambda": 2.6, "coop_type": "conditional_cooperator",
               "risk": {"economic":"low","bodily":"low","social":"mid"} },
    "attention_anchor": 0.75,
    "curiosity_anchor": { "social": 0.25, "spatial": 0.05, "skill": 0.6, "economic": 0.1, "risk": 0.0 },
    "quirks": [
      { "id":"swamp_loyalist","origin":"derived","domain":"movement","trigger":"pre_action:fish",
        "binding":"fishing_spot_preference","relation":"prefers","object":"area:Lumbridge_swamp",
        "strength":0.8,"observable":true,"suppress_when":"low_hp","narrative":"Always fishes the swamp, never the river." },
      { "id":"scam_wary","origin":"idiosyncratic","domain":"trade","trigger":"trade_request",
        "binding":"trade_screening","relation":"distrusts","object":"player_type:stranger",
        "strength":0.7,"observable":true,"suppress_when":"","narrative":"Screens every trade from someone it doesn't know." }
    ],
    "pinned": [ { "summary":"Lost my first lobster stack to a fake-trade by a stranger.","weight":0.9 } ]
  }
}
```

### B — "Reckless wilderness PKer" (cohort: Wilderness PKers)

```jsonc
{
  "cornerstone": {
    "identity": {
      "name": "xX_dragon_Xx", "archetype_tag": "wilderness_pker",
      "backstory": "Lives in the wild. Trusts almost no one, hunts everyone. Talks trash, takes risks, dies and comes back.",
      "voice": { "register": "terse", "formality": "text_speak", "tics": ["ez","get rekt","lol"], "typo_rate": 0.10 }
    },
    "north_star": { "theme": "combat", "statement": "Be the most feared PKer in the wilderness",
                    "horizon": "month", "success_signals": ["pk_kills > 500"] },
    "hexaco": { "H": {"mu":0.18,"band":"low"}, "E":{"mu":0.22,"band":"low"},
                "X": {"mu":0.60,"band":"mid_high"}, "A":{"mu":0.20,"band":"low"},
                "C": {"mu":0.35,"band":"mid"}, "O":{"mu":0.55,"band":"mid"} },
    "values": { "north_star_value": "power", "secondary_value": "stimulation",
                "higher_order": {"self_transcendence":0.05,"self_enhancement":0.55,"openness_to_change":0.30,"conservation":0.10} },
    "prefs": { "patience_tau": 0.30, "loss_aversion_lambda": 1.2, "coop_type": "free_rider",
               "risk": {"economic":"high","bodily":"very_high","social":"high"} },
    "attention_anchor": 0.45,
    "curiosity_anchor": { "social": 0.15, "spatial": 0.2, "skill": 0.05, "economic": 0.1, "risk": 0.5 },
    "quirks": [
      { "id":"always_engage","origin":"derived","domain":"combat","trigger":"other_player_damage",
        "binding":"combat_response","relation":"prefers","object":"action:retaliate",
        "strength":0.9,"observable":true,"suppress_when":"low_hp","narrative":"Attacks back instantly when hit, never flees first." },
      { "id":"trash_talk","origin":"idiosyncratic","domain":"social","trigger":"npc_killed",
        "binding":"chat_emit","relation":"prefers","object":"phrase:taunt",
        "strength":0.6,"observable":true,"suppress_when":"in_combat","narrative":"Taunts after a kill." }
    ],
    "pinned": [ { "summary":"First PK kill in level-20 wild — hooked ever since.","weight":0.8 } ]
  }
}
```

Note how the **same fields** produce two utterly different hosts, and how the
sampled numbers *cohere*: PKer is low-H (will exploit) + low-E (won't flee) +
low-A (retaliates) + low-τ (goal-hops) + low-λ (dives into danger) + high
`risk.bodily` + high `curiosity.risk` + free-rider. The forbidden-pair check would
*reject* a draw like high-H + value:power.

---

## 13. Authoring checklist & anti-patterns

### Checklist (per persona / per cohort)
- [ ] Picked a **cohort** with a trait prior, a population weight, and forbidden pairs.
- [ ] `north_star` is a *motivator*, not an optimization target; specificity in the middle.
- [ ] HEXACO draw **coheres** with values/prefs (no high-H scammer); forbidden pairs encoded.
- [ ] `H` set deliberately — it's the trust prior AND the ethics signal.
- [ ] `risk` is **domain-split** (econ/bodily/social), not one scalar.
- [ ] `attention` (short) and `patience_tau` (long) set independently.
- [ ] `curiosity_anchor` flavored to make this host *diverge* from cohort-mates.
- [ ] **2–7 quirks**, all `observable`, all bound to a mechanic, dangerous ones have `suppress_when`.
- [ ] **2–5 pinned memories** that *explain* the sampled traits; not more (cache cost).
- [ ] `voice` varied hard from other cohorts (typo_rate, tics, formality) — believability.
- [ ] Did NOT author any `trajectory` field (the runtime owns it).

### Anti-patterns (from `personas.md`)
- **Over-specified → robotic.** Quirks/north-stars that dictate *actions* ("walk to
  Falador at 9am") not *priorities*. Constrain style, not steps.
- **Too consistent → uncanny.** Real people are inconsistent. That's what the
  forbidden-pair *allowances*, jitter, and reverie noise are for — leave some.
- **Cohort = role.** A cohort isn't a job. Any host can do any RSC activity; cohort
  changes *how* and *how often*, not *what's possible*.
- **Stable forever.** Don't fight the Trajectory. A "trusting" host that gets
  scammed three times *should* drift cautious — that's the research, not a bug.
- **Flavor-text quirks.** "Is nostalgic" is not a quirk. If the server can't
  observe it, it doesn't exist.
- **Visible archetype_tag.** Never let the brain see the label; it'll role-play the
  word instead of behaving from traits.

---

## 14. What's still OPEN (not settled by this schema)

The *shape* above is settled. These remain the deferred paired Alex+Claude session
(see [personas.md](personas.md) "What needs to be designed together"):

- **Cohort taxonomy** — how many (10/20/50?), the distribution, whether hosts
  transition cohorts, how cohorts are assigned at registration.
- **The 13-archetype prototype table** + `cohorts.yaml` priors + forbidden-pair set.
- **The reverie catalog** — which specific small behaviors exist + their trait
  coefficients (cross-refs [reveries.md](reveries.md)).
- **The north-star catalog** + how conflicting drives resolve.
- **The LLM persona-generation pipeline** — the exact offline batched-materialize
  call (cohort+template+seed → prose).
- **The population mix** for 500 hosts (the Pareto skew: ~60% casual, ~10% extreme).
- **The compiler band→policy mapping table** + the reverie coefficient set (the
  highest-leverage unbuilt piece — see `decision-persona-schema.md` §5).

---

## 15. Cross-references
- [personas.md](personas.md) — canonical schema + the deferred design.
- [`_research/quantitative-persona-models.md`](_research/quantitative-persona-models.md) — the per-model ADOPT/ADAPT verdicts + update rules + math.
- [`_research/reference/decision-persona-schema.md`](_research/reference/decision-persona-schema.md) — the Go types, the compiler, the EventPolicy bridge.
- [`_research/social-graph-and-trust-ledger.md`](_research/social-graph-and-trust-ledger.md) — the trust ledger that consumes `H` + `coop_type`.
- [`_research/host-bootstrap-and-knowledge-gating.md`](_research/host-bootstrap-and-knowledge-gating.md) §7 — the attention/curiosity temperament dials.
- [reveries.md](reveries.md) — the reverie catalog (deferred).
- [mesa.md](mesa.md) — `hosts.persona` JSONB + `persona_revisions` storage.
</content>
</invoke>
