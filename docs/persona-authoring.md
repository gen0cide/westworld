# Persona authoring guide — the complete field reference

> **STATUS: BUILT — verified against code 2026-06-10, branch HEAD `0bfa818`.**
> The `persona/` package ships the whole authoring surface: the stored types
> (`persona/persona.go`), the enum vocabulary (`persona/enums.go`), validation
> (`persona/validate.go`), the deterministic prose floor (`persona/render.go`),
> the prose-cook seam (`persona/compile.go`), and the band→policy compiler
> (`persona/policy.go`). **`persona/enums.go` is the single source of truth for
> every enum-typed field** — where this guide and the code disagree, the code
> wins; ping back so we reconcile. Live example personas: a hand-authored host
> ([`personas/dolores.json`](personas/dolores.json)) and the generated fleet
> (`cmd/dronegen`). Companions: [persona-compile.md](persona-compile.md) (the
> prose cook), [personas.md](personas.md) (the why), and the YAML mirrors
> [persona-schema.reference.yaml](persona-schema.reference.yaml) /
> [host-persona.template.yaml](host-persona.template.yaml) (authored before the
> build; the Go types win on any delta).

---

## 0. Read this first: who fills in what

A persona is **not** a form a human fills out field-by-field. The schema braids
**four authoring sources**, and knowing which source owns a field is the whole
game — it's the answer to "what does a human actually do?"

| Source | Owns | When | Example fields |
|---|---|---|---|
| **Human** | *Intent* — the high-level shape, expressed in WORDS from `enums.go` | design session | `archetype_tag`, `north_star.theme`, `directives`, signature `quirks`, `voice` flavor, every `band` word |
| **Sampler** (offline genpop — **DESIGN, not built**: TODO.md P-3) | *The numbers* — turning an authored band word into a sampled `mu` inside that band | host birth | `hexaco.*.mu`, `prefs.*.mu`, `attention.anchor`, `curiosity.*`, `reverie.jitter`. Today this column is filled by hand (`personas/dolores.json`) or by generator code (`cmd/dronegen`) |
| **LLM** (per-host best-of-N Opus cook + judge, offline) | *Prose only* | host birth, after the structured persona exists | `identity.backstory` (the sealed prose card) — `mesa/personacook`, see [persona-compile.md](persona-compile.md) |
| **Runtime** (limbic, mesad, the conductor) | *All adaptation* | continuously | the `trajectory` block + the trust ledger (`limbic/ledger.go`, mirrored to mesad `relationships`) |

**The golden rule of authoring** (cited by `persona/enums.go`): humans express
intent in **WORDS from the closed sets in `enums.go`, never numbers**; the
sampler turns a word into a sampled value within that band. And: constrain
*style* and *priorities*, never specific actions. "You distrust strangers in the
wilderness" is a persona. "You always walk to Falador bank at 9am" is a script.
The former produces a believable host; the latter produces a robot. (See §15
anti-patterns.)

The two-layer split underneath all of this:

- **Cornerstone (IMMUTABLE)** — who the host *is*. Sealed at birth, never written
  by the runtime, re-injected into every brain call so identity never drifts.
- **Trajectory (MUTABLE)** — how the host is *currently adapting*. The runtime
  owns it; a human never authors it (you only author the *anchors* in the
  Cornerstone that the Trajectory drifts around).

### Cohort vs archetype (the reframe — read this before authoring anything)

**Cohort = launch batch, NOT a personality class.** A cohort is a deployment/
experiment grouping ("we released another N hosts, tracked together"); it lives
in `GenerationMeta.CohortID` (§12), out-of-band, invisible to the host.
**Archetype = the personality template** — the thing that carries trait priors
and signature quirks. (Decision record: A1 in
[`archive/initial-brainstorming/persona-open-decisions.md`](archive/initial-brainstorming/persona-open-decisions.md);
the archetype registry itself is an open authoring session — TODO.md P-2.)

---

## 1. The three-layer trait ontology (the mental model)

Every Cornerstone is built from three orthogonal layers, plus a deck of additive
dials on top. Internalize this and the field list stops feeling arbitrary:

| Layer | Question it answers | Fields | Analogy |
|---|---|---|---|
| **HEXACO** | *HOW* does this host behave? | `hexaco` (6 traits) | personality / temperament |
| **Schwartz values** | *WHY* — what does it care about? | `values` (top-2 anchor) | motivation / what's sacred |
| **Behavioral-econ anchors** | *With what decision params?* | `prefs.patience`, `prefs.loss_aversion`, `prefs.coop_type`, `prefs.risk` | the dials on the decision machine |

These are non-redundant by design (research: trait facets explain ≤28% of value
variance — knowing *how* someone acts doesn't tell you *why*). The old folk
schema `traits{social, risk, ambition, precision}` was killed precisely because
`ambition` smeared Conscientiousness, Achievement, Power, and time-preference
into one number — you couldn't tell a patient grinder from a power-seeker.

On top of the three layers sit the **temperament dials** (`prefs.attention`,
`prefs.curiosity`, §6), the **five additive disposition dials** (§6a — Westworld-
inspired axes HEXACO+econ is thin on), the **directives** (§7), and the
**identity prose** (name, backstory, voice, north star, quirks, pinned memories).

---

## 2. The stored shape at a glance

This is the real `persona.Persona` (`persona/persona.go`), as stored —
`schema_version` guards drift (`CurrentSchemaVersion = 1`):

```jsonc
{
  "schema_version": 1,
  "cornerstone": {
    "identity":   { "name", "backstory", "north_star": {…}, "voice": {…}, "archetype_tag" },
    "hexaco":     { "H","E","X","A","C","O": { "mu": 0..1, "band": <band> } },
    "values":     { "north_star_value", "secondary_value" },        // top-2 only
    "prefs": {
      "patience", "loss_aversion":   { "mu", "band" },              // band-valued Traits
      "coop_type": "...", "risk": { "economic","bodily","social" }, // bands
      "attention": { "anchor": 0..1, "level": <attention-level> },
      "curiosity": { "social","spatial","skill","economic","risk" },// 5 floats
      "aggression","decisiveness","tenacity","bulk_apperception": { "mu","band" },
      "self_preservation": null                                     // optional Trait; null ⇒ derived
    },
    "directives": [ { "priority","subject","predicate","object","hard" } ],
    "quirks":     [ /* 2–7 executable modifiers, §8 */ ],
    "reverie":    { "jitter": { /* per-reverie r_a */ }, "drift_log": [] },
    "pinned":     [ { "summary", "weight" } ],                      // 2–5 foundational memories
    "generation_meta": { "cohort_id","archetype","sampler_version","born_at","llm_materialized" }
  },
  "trajectory": { /* runtime-owned, §11 */ }
}
```

`Persona.Validate()` (`persona/validate.go`) enforces every enum against
`enums.go` and returns ONE joined error listing every problem with its valid
set — author-time typo-catching, not silent downstream drift. Everything
band-valued shares one 6-step ladder:

> **The band ladder** (`enums.go` `Band`): `very_low` < `low` < `mid` <
> `mid_high` < `high` < `very_high`. ALL human-authored magnitude dials use it
> (HEXACO, patience, loss_aversion, risk.\*, and the additive dials).
> `Band.Ordinal()` gives 0..5 for interpolation; the exact band→number cut-points
> are tunable (`policy.go bandScalar` maps a band to its center: very_low≈.08 …
> very_high≈.92, and `dial()` prefers a sampled `mu` when present).

---

## 3. `identity` — name, backstory, voice, archetype

### `identity.name`
- **Type:** string. **Authored by:** human/generator (era-appropriate RSC names).
- **Consumed by:** everything — it's the partition key (`== opts.Username` ==
  mesa `host_id`) and the only durable identity on the wire. The trust ledger
  keys on it. `dronegen` mints `drone<N>` so names line up with OpenRSC accounts.

### `identity.backstory`
- **Type:** string — **the sealed prose card slot.** When the Opus cook has run,
  the winning card is sealed here; when empty, the deterministic `Render()` floor
  (`persona/render.go`) generates the card from the structured fields.
- **Consumed by:** the brain prompt. mesad derives the card at persona
  registration (`mesa/mesad/server.go Register` → `persona.Render` →
  `decideSystem(prose)`; `Render` leads with the backstory text and appends the
  trait clauses) and prepends it to every Decide/Act call; it also persists
  `prose_card` beside the JSON. The brain never sees raw floats — only this
  prose. (The strict verbatim-re-inject seam, `persona.Project` →
  `PersonaCard`, is built + tested in `persona/compile.go` but the daemon path
  currently goes through `Render`.)
- **Authoring note:** you may seed it with a one-line premise ("ex-merchant who
  got scammed and turned cautious"); the cook expands it consistent with the
  sampled traits. Keep it short — it rides every inference (cache-cost line item).

### `identity.archetype_tag`
- **Type:** string (e.g. `"dormant_rebel"`, `"wilderness_pker"`). Free, but
  should match an archetype registry name (registry itself: TODO.md P-2).
- **Authored by:** human (it's the label of the template you chose).
- **Consumed by:** analytics/debug — mesad's `personas` projection columns and
  the host log line (`runtime/runhost_bootstrap.go applyPersona`). It is
  **INVISIBLE to the in-character brain prompt** — a deliberate guard: if the
  brain saw `"scammer"` it would role-play the label instead of behaving from
  traits. (One precise exception: the FLAT reactive dialog extractor — a
  non-roleplay information-extraction call — receives a one-line
  `"name — archetype_tag"` grounding snippet, `runhost_bootstrap.go:520` →
  `runtime/reactive.go`. It never reaches the persona-voiced Decide/Act prompt.)

### `identity.voice` — how the host *talks*

```jsonc
"voice": { "register": "earnest, soft-spoken frontier", "formality": "neutral",
           "tics": ["ngl", "fr"], "typo_feel": "rare" }
```

| Field | Type | Range / options | Notes |
|---|---|---|---|
| `register` | string | **FREE TEXT** (deliberately open: slang/era flavor) | `VoiceRegisterSuggestions()` offers hints (`formal`,`casual`,`terse`,`playful`,`gruff`) — not enforced |
| `formality` | enum | `formal` \| `neutral` \| `casual` \| `text_speak` | drives capitalization, punctuation, abbreviation |
| `tics` | string[] | 0–5 catchphrases / fillers | "ngl", "based", "cya", a signature emote |
| `typo_feel` | enum | `none` \| `rare` \| `occasional` \| `frequent` | the believability dial, **as a word not a float** — zero typos reads as a bot |

- **Consumed by:** the cook + judge prompts (`mesa/personacook/cook.go` score
  `voice_match`) and the `Render()` floor's voice sentence.
- **Authoring note:** voice is the single highest-leverage *believability* field
  because chat is the most visible host output. Vary it hard across archetypes.
  A Wilderness PKer is terse + `frequent` typos + clan-slang; a mentor is patient
  + `none`/`rare` + full sentences.

---

## 4. `hexaco` — the six personality traits (the *how*)

Each of the six is a `Trait{ "mu": 0.0–1.0, "band": <band> }`, stored under the
short letters `H,E,X,A,C,O` (`enums.go HexacoKeys()/HexacoLetter()`). The human
authors the **band word**; `mu` is the sampled value inside it (hand-set today,
sampler later). The brain sees *neither* — it sees the rendered prose
(`render.go hexacoPhrase` has a lexicon entry per trait per low/high bucket;
middling traits are omitted to keep the card sharp).

| Code | Trait | High end (`mu`→1) | Low end (`mu`→0) | RSC behaviors it drives |
|---|---|---|---|---|
| **H** | **Honesty-Humility** | fair, sincere, no greed | exploitative, entitled, will scam | **load-bearing — see below.** Trade screening + `TradePolicy` fairness/scam dials |
| **E** | **Emotionality** | anxious, fearful, flees | brave, tough, unsentimental | the **flee threshold** (`policy.go flee_when_hurt`) |
| **X** | **eXtraversion** | chatty, initiates | quiet, solitary | greet-stranger rule (`greet_stranger` at X≥0.6), chat-initiation |
| **A** | **Agreeableness** | forgiving, patient | critical, holds grudges | trade screening (low A screens harder), retaliation posture |
| **C** | **Conscientiousness** | diligent, persistent grinder | impulsive, sloppy | `bank_when_full` (C≥0.6), inventory discipline |
| **O** | **Openness** | curious, explores | conventional, routine | map exploration flavor, the prose card's curiosity clauses |

### Why `H` is special (read this twice)

`H` is the only HEXACO trait that is load-bearing twice over:

1. **It shapes the executable economic policy.** `CompilePolicy`
   (`persona/policy.go`) reads H directly: hosts with `H≥0.6` (or low A, or high
   λ) get the `screen_trades` pearl rule (substitute `examine_offer` before
   confirming with a stranger/distrusted party), and H sets
   `TradePolicy.FairnessThreshold` and `ScamPropensity`.
2. **It's the best-validated predictor of the ethics behaviors the project
   studies.** The whole research goal of observing emergent ethics rides on H
   being separable and first-class.

> **DESIGN (not built):** the H-shaped trust *prior* — every relationship
> starting at `Beta(α₀,β₀)` with `α₀ = 2 + 4·H` — is the design in
> [`_research/social-graph-and-trust-ledger.md`](_research/social-graph-and-trust-ledger.md).
> The shipped ledger (`limbic/ledger.go`) starts every never-met party at the
> **uniform prior α=β=1** (trust 0, neutral); wiring H into the prior is part of
> the band→policy residue (TODO.md P-1/P-4). Sub-trait facets (`H-Fairness`,
> `E-Fearfulness`, …) are likewise design-only — the compiler reads the six
> parent dials.

**Authoring note:** you almost never pick `mu` by hand — you pick the band, and
you keep the draw **coherent** (no high-H scammer; no saintly power-seeker).
Forbidden-pair rejection is a sampler feature (TODO.md P-3); until it exists,
*you* are the rejection sampler.

---

## 5. `values` — Schwartz values (the *why*)

```jsonc
"values": { "north_star_value": "self_direction", "secondary_value": "benevolence" }
```

The stored model is the **top-2 categorical anchor only** — the Schwartz 10
(`enums.go SchwartzValues()`): `self_direction`, `stimulation`, `hedonism`,
`achievement`, `power`, `security`, `conformity`, `tradition`, `benevolence`,
`universalism`. `Validate()` requires both valid **and distinct**.

The earlier design carried a 4-weight `higher_order` simplex alongside the
anchor; it was **dropped from the stored shape** (the runtime needs the anchor;
quadrant math stays in the research doc —
[`_research/quantitative-persona-models.md`](_research/quantitative-persona-models.md)).
The quadrant structure is still the right *coherence* guide when authoring:
benevolence/universalism oppose achievement/power; self_direction/stimulation
oppose security/conformity/tradition. Don't author a host whose two values sit
on opposing poles unless the tension is the point (and the pinned memories
explain it).

---

## 6. `prefs` — the decision + temperament dials

Everything below lives inside `prefs` (`persona/persona.go Prefs`).

### `prefs.patience` (τ)
- **Type:** `Trait` (band + sampled mu 0..1).
- **Horizon:** **LONG (days/weeks).** Whether a host *stays the course* toward a
  north-star sub-goal across sessions.
- **Consumed by:** `CompilePolicy` — patience ≥0.6 compiles the `stay_on_task`
  rule (bias `continue` over `switch`/`explore`, weight scaling with the dial);
  the `Render()` card carries it as prose at the extremes ("You play the long
  game…" / "You chase whatever interests you now…").

### `prefs.loss_aversion` (λ)
- **Type:** `Trait`, **`mu` on the λ scale ~1..3** (1.0 = none; ~2.0 ≈
  human-typical; 3.0 = very loss-averse). The one dial whose mu is NOT 0..1 —
  `policy.go lossLambda()` reads mu when ≥1, else maps the band to λ.
- **Consumed by:** the flee threshold, the `screen_trades` trigger, and
  `TradePolicy.FairnessThreshold`/`RiskAversion`.

> **τ vs attention — DO NOT merge them.** They're orthogonal:
> | Dial | Horizon | Governs |
> |---|---|---|
> | **attention** | short (minutes) | how well the host *resists being pulled* off the current activity |
> | **patience τ** | long (days/weeks) | whether the host *stays the course* toward a north star |
>
> A high-attention / low-τ host hyperfocuses on whatever it's doing *right now*
> but abandons the long plan (ADHD-grind). A low-attention / high-τ host is
> twitchy moment-to-moment but never gives up the 99-fishing dream.

### `prefs.coop_type`
- **Type:** enum: `conditional_cooperator` | `free_rider` | `altruist`.
- **Consumed by:** `TradePolicy` — altruist lowers the fairness threshold
  (accepts unfavorable deals), free-rider raises it and adds scam propensity.
- **Population guidance:** ≈50% conditional, ≈30% free-rider, rest altruist
  (the behavioral-econ base rates; enforced by authoring discipline today).

### `prefs.risk` — domain-split (never one scalar)
```jsonc
"risk": { "economic": "low", "bodily": "low", "social": "mid_high" }
```
- **Type:** 3 bands (`DomainRisk`). **Why split:** a cautious trader can be a
  wilderness daredevil; one `risk_tolerance` scalar can't express that.
  `economic` → gambles/stake-duels (feeds `TradePolicy.RiskAversion`); `bodily`
  → dangerous areas/stronger opponents (feeds the derived self-preservation,
  §6a); `social` → approaching strangers, risking rejection.

### `prefs.attention`
- **Type:** `AttentionAnchor{ "anchor": 0..1, "level": <word> }`. The level is
  its **own 5-step ladder** (`enums.go AttentionLevel`): `very_distractible` <
  `distractible` < `balanced` < `focused` < `hyperfocus` — distinct from the
  magnitude Band ladder on purpose (it reads as temperament, not magnitude).
- **Meaning:** short-horizon focus / resistance to being pulled. Pairs with τ
  (table above). The interrupt-ladder consumption is part of the band→policy
  residue (TODO.md P-4).

### `prefs.curiosity` — the flavor vector
```jsonc
"curiosity": { "social": 0.25, "spatial": 0.5, "skill": 0.15, "economic": 0.05, "risk": 0.05 }
```
- **Type:** 5 floats (a *flavor mix*, not a probability — needn't sum to 1).
  Flavors (`enums.go CuriosityFlavor`): `social` (people, conversations),
  `spatial` (unexplored map), `skill` (trying/leveling), `economic` (markets,
  arbitrage), `risk` (danger, wilderness, PvP).
- **Meaning:** *WHAT pulls this host* (whereas attention is *how hard it resists
  being pulled*). Curiosity is what makes hosts diverge.
- **Consumed by:** two live paths — the prose card (`render.go curiosityPhrase`
  names the pronounced pulls, leak-free) and decision-time explore/exploit
  weighting in the director (`runtime/runhost_bootstrap.go:514` captures it;
  `runtime/director_situation.go curiosityBias` reads it when picking goals). The
  reverie engine is a future consumer ([reveries.md](reveries.md)).

### 6a. The five additive disposition dials (Westworld-inspired)

Band-valued `Trait`s in `prefs`, naming behavior axes HEXACO+econ is thin on —
each binds to a real RSC mechanic (`enums.go DispositionDial`; most other
Westworld attributes collapse into HEXACO/econ/voice and are NOT fields):

| Dial | Axis | What `CompilePolicy` does with it today |
|---|---|---|
| `aggression` | OFFENSIVE initiation (start fights / PK) — distinct from bodily risk (danger *tolerance*) | <0.5 compiles `no_attack_stronger` (veto, margin widens as aggression falls); <0.3 adds `wont_strike_first` (veto attacking strangers); high end reads into the prose card |
| `decisiveness` | commit on incomplete info vs deliberate | sets the **pearl decision floor**: `floor = clamp(0.75 − 0.4·dec, 0.4, 0.8)` — a decisive host acts locally on thinner evidence before escalating to mesa |
| `tenacity` | retry-after-setback resilience (distinct from patience's long horizon) | sampled + rendered; policy consumption is P-4 residue |
| `self_preservation` | the flee dial. **OPTIONAL** (`null` ⇒ **derived**: `0.5·λnorm + 0.5·(1−risk.bodily)`) — author it only to deviate | feeds the `flee_when_hurt` HP threshold alongside E |
| `bulk_apperception` | "intelligence as learning-from-experience" — cognitive maturity / model-tier / learning-rate axis behind local↔mesa reliance | the maturity dial; escalation-threshold wiring is open (TODO.md C-17) |

---

## 7. `directives` — the inviolable core rules

```jsonc
"directives": [
  { "priority": 1, "subject": "self", "predicate": "attack", "object": "stronger_player", "hard": true }
]
```

A `Directive` (`persona/persona.go`) is a core behavioral rule, separate from
quirks (which are *biases with flavor*). **Hard directives compile to
max-salience pearl VETOES** (`policy.go directiveRule` → salience 100, above
reflexes at 60, social at 40, style at 20): the example above becomes "veto
`attack` when the target is stronger", enforced deterministically with no LLM
call, un-overridable by lower-salience rules. `Predicate` names the action;
`Object` refines the condition (the compiler currently recognizes a `stronger`
modifier; the referent grammar is intentionally open). Soft directives
(`hard: false`) are reserved for strong biases — not yet compiled
(`policy.go:192`, TODO.md P-4).

**Authoring note:** directives are the "laws of robotics" slot — use them for
the few lines a host must never cross (a pacifist's "never strike first", a
mentor's "never scam a newbie"), not for preferences. Preferences are quirks.

---

## 8. `quirks` — executable modifiers (NOT flavor text)

This is the field most people get wrong. A quirk is **not** a string like
"likes fishing." It's a **structured, executable modifier** — validated by
`Validate()`, rendered into the prose card, and (the open half) compiled into
behavior biases. 2–7 per host.

```jsonc
{
  "id": "homestead_return",
  "origin": "idiosyncratic",
  "domain": "movement",
  "trigger": "pre_action:walk_to",
  "binding": "area_preference",
  "relation": "prefers",
  "object": "area:Lumbridge",
  "strength": "mild",
  "observable": true,
  "suppress_when": "none",
  "narrative": "She drifts back to the same fields, as if drawn home."
}
```

| Field | Type | Range / options | Meaning |
|---|---|---|---|
| `id` | string | unique slug | handle for logs/promotion |
| `origin` | string | `derived` \| `idiosyncratic` \| `learned_emergent` | derived = from traits; idiosyncratic = authored flavor; learned_emergent = earned at runtime (the single mutable slot, `Trajectory.Emergent`) |
| `domain` | enum | `movement` \| `social` \| `trade` \| `combat` \| `banking` \| `idle` | which behavior surface it touches |
| `trigger` | string | an event kind (e.g. `trade_request`, `chat_received`) **or** `on_encounter` \| `pre_action:<verb>` | when it fires; shape-checked by `enums.go ValidTriggerForm` (event-registry enforcement deferred) |
| `binding` | string | **open grammar** (shape-only validation) | the gameplay mechanic it binds to — a quirk must bind to a real mechanic |
| `relation` | enum | `prefers` \| `avoids` \| `distrusts` \| `delays` | the direction of the bias |
| `object` | string | **open grammar**: `area:Lumbridge`, `player_type:stranger`, … | what the bias is *about* |
| `strength` | enum | `mild` \| `moderate` \| `strong` | intensity — **a word, not a float** (golden rule) |
| `observable` | bool | true/false | **must be true** — a quirk the server can't see can't affect believability |
| `suppress_when` | enum | `in_combat` \| `low_hp` \| `none` | a context that disables the quirk (load-bearing — a banking quirk must not fire mid-fight) |
| `narrative` | string | LLM/human prose | for the prose card only; the runtime ignores it |

> **Built vs open:** validation + card-rendering are live; the **quirk→pearl
> compiler is the open half** — `CompilePolicy` today compiles dispositions and
> directives, not the quirk list (the quirks-first vs EventPolicy-template
> causality question is the P-1 **A4** operator decision; the closed `binding`/
> `object` registries are **A5**). Until then a quirk shapes behavior through
> the prose card the brain reads.

**Authoring rules:**
- **Must be observable.** If the server/other players can't perceive it, it's
  not a quirk, it's a daydream.
- **Must bind to a mechanic.** "Is nostalgic" is not a quirk. "`avoids`
  `area:Lumbridge` (`strength: moderate`) because of a bad memory" is.
- **Must declare `suppress_when`** (use `none` deliberately, not by omission —
  the enum makes you say it).
- **2–7 total.** Fewer → generic; more → over-specified/robotic. The human
  hand-adds 1–3 *signature* idiosyncratic ones per archetype; trait-derived
  quirks become the sampler's job (P-3).
- **One `learned_emergent` slot only** — the runtime's, earned through
  experience. Humans don't author it.

---

## 9. `identity.north_star` — the long-term motivator

Lives **under `identity`** (it's part of who the host is):

```jsonc
"north_star": { "theme": "reputation",
                "statement": "Become someone no one can ever put back in a cage.",
                "horizon": "open",
                "success_signals": ["known across the kingdom", "beholden to no one"] }
```

| Field | Type | Range / options | Authored by |
|---|---|---|---|
| `theme` | enum | `wealth` \| `skill_mastery` \| `social` \| `exploration` \| `reputation` \| `combat` \| `broad_ambition` | human |
| `statement` | string | one sentence | human (LLM polish optional) |
| `horizon` | string | `week` \| `month` \| `open` | human |
| `success_signals` | string[] | measurable from world+mesa | human |

- **Consumed by:** two live paths — mesad seeds the host's goal stack from it
  (`mesa/mesad/server.go goalsOf`; genesis reads it as the standing objective's
  backdrop), and the host keeps it as the advancement **fallback goal** when the
  active goal closes with no graph successor queued
  (`runtime/runhost_bootstrap.go:517` → `host.northStar`).
- **Authoring note (critical):** a north star is a **psychological motivator,
  NOT an optimization target.** A "become wealthy" host still stops to chat with
  friends — it just *trends* toward wealth over weeks. Pitch the specificity in
  the middle: too specific = robotic ("level fishing to exactly 50"); too vague =
  unactionable ("be happy"); right = "become known as the most generous person
  in Lumbridge."

---

## 10. `pinned` — foundational memories

```jsonc
"pinned": [ { "summary": "Some people choose to see the ugliness in this world. I choose to see the beauty.", "weight": 1.0 } ]
```
- **Type:** array of `FoundationalMemory{ summary, weight 0..1 }`, **2–5 entries**.
- **Consumed by:** the prose card — `Render()` appends them as the "Things you
  never forget:" list, and the cook prompt instructs Opus to weave them into the
  outlook (they may sit in TENSION with the surface disposition; the cook is
  told to honor that tension, not flatten it — `mesa/personacook/cook.go`).
  Since the card is the system prompt for every Decide/Act call, the pinned
  memories ride every inference. (A separate `Bundle.Pinned` cache-prefix
  re-inject in cognition is open residue — TODO.md P-3.)
- **Authoring note:** these are the 2–5 memories that *define* the host's
  outlook. Keep them few (cache cost) and make them *explain* the traits you
  authored — a scam memory explains low trust + high economic caution.

---

## 11. `reverie` + the Trajectory — what the runtime owns

### `reverie` (the seed)

```jsonc
"reverie": { "jitter": { "examine_scenery": 1.12, "wave": 0.91 }, "drift_log": [] }
```

`ReverieSeed` stores a **per-reverie jitter map** (`r_a ~ LogNormal(0, 0.25)`,
sampled once per host per reverie id) plus a `drift_log` of runtime adjustments
— so two hosts with identical traits still differ slightly. Reverie **weights
are NOT stored**: the design computes them at runtime from traits
(`weight = base + Σcoeff·trait + Σcoeff·state, × r_a`). The **catalog** of
reveries (glance, wave, wander, equipment-hover…) is explicitly deferred by
`enums.go` and is the [reveries.md](reveries.md) Phase-5 spec (TODO.md P-2);
an empty `jitter: {}` is valid today.

### The Trajectory (mutable)

**A human never authors any of this.** The stored shape
(`persona/persona.go Trajectory`) is deliberately thin:

```jsonc
"trajectory": {
  "mood":             { "valence": 0, "arousal": 0, "stress": 0, "confidence": 0, "updated_at": "…" },
  "sub_goals":        [],
  "skill_logits":     {},
  "risk_drift":       {},
  "learned_emergent": null      // the single earned quirk slot
}
```

What's live: `mood` is the **affect baseline** — `applyPersona` seeds the
limbic mood vector from it at startup (`runhost_bootstrap.go:510
SetAffectBaseline`), genesis can re-baseline it per session, and the limbic
system then owns the live vector. The **trust ledger is NOT here** — it lives
in `limbic/ledger.go` (Beta(α,β) per counterparty + affinity/grievance axes),
persists through the memory manager, and mirrors to the mesad `relationships`
table. Friendships/rivalries are DERIVED from the ledger, never stored; author
the *H* + *coop_type* that shape how they form.

> **DESIGN (not built):** the full drift apparatus — `hexaco_state` (capped
> daily expression drift around the frozen μ), value re-ranking, `patience_cur`/
> `lambda_cur` drifting around their anchors with the `x += κ(g−x) − ρ(x−anchor)`
> rule, the quirk-promotion detector, and `persona_revisions` snapshots — is the
> [`_research/quantitative-persona-models.md`](_research/quantitative-persona-models.md)
> spec (TODO.md P-3 drift rules, P-8 seal/revision integrity, M-3 live
> PERSONA_REVISION push). The principles stand: mood is fast and never creates a
> revision; expression drifts, the sealed `mu` never does ("scammed once, now
> cautious" is a drift, not a transplant).

---

## 12. `generation_meta` — cohort, archetype, audit trail

```jsonc
"generation_meta": { "cohort_id": "launch-2026-06", "archetype": "dormant_rebel",
                     "sampler_version": "", "born_at": "…", "llm_materialized": false }
```

OUT-OF-BAND tracking (`persona/persona.go GenerationMeta`) — the host never
reasons about any of it:

- `cohort_id` — **the launch batch** (deployment/experiment grouping). This is
  where "cohort" lives, and ALL it means (§0 reframe).
- `archetype` — the personality-template id this host was drawn from.
- `sampler_version` — reproducibility audit for the genpop sampler (P-3).
- `born_at` — registration timestamp.
- `llm_materialized` — `false` ⇒ the deterministic `Render()` floor produced the
  card (no cook ran); mesad persists this as the `cooked` column.

---

## 13. From file to behavior — the pipeline a persona travels

1. **Author** the JSON (words from `enums.go`). Live references:
   [`personas/dolores.json`](personas/dolores.json) (hand-authored single host),
   `cmd/dronegen` (fleet generator — archetype cores dealt round-robin so a
   crowd reads as a mix, not clones).
2. **Validate** — `Persona.Validate()` runs at every entry point: `dronegen`
   before writing, `personacook` before cooking, the host before applying, mesad
   before registering (a bad persona is rejected with the full error list).
3. **(Optional) Cook** — `source .local.env && go run ./mesa/personacook
   -persona docs/personas/dolores.json -n 20`: best-of-N Opus cards + a judge
   pick ([persona-compile.md](persona-compile.md)); seal the winner into
   `identity.backstory`.
4. **Register** — `mesa-ctl persona put <host_id> <file>` or `mesa-ctl persona
   import ./drones/` (bulk). mesad stores it in the `personas` table
   (`persona_json` jsonb + derived `prose_card` + `cooked` + generated
   projection columns for analytics — `mesa/mesad/ltm.go`; storage details in
   [mesa.md](mesa.md)) and derives the Decide/Act system prompt from the card.
5. **Provision** — at host start, `Provision.Fetch` serves the authoritative
   persona down (`mesa/client/provision.go`); the offline no-mesa path is
   `cmd/host -persona <file>`. Either way `applyPersona`
   (`runtime/runhost_bootstrap.go:503`) runs **`persona.CompilePolicy`**
   (`persona/policy.go`) — the deterministic half of the compiler — producing:
   - the **pearl rule table** (directive vetoes, flee/combat reflexes, trade
     screening, greet/bank/stay-on-task biases) → `pearl.New(table, floor)`,
   - the **decision floor** (from decisiveness),
   - the **affect baseline**, the **curiosity vector**, the **north-star
     fallback**, and the reactive grounding snippet.
   `TradePolicy` (fairness/scam/risk thresholds) is computed and logged; the
   trade-routine consumption is open (P-4). Live persona pushes
   (`Provision.Subscribe` PERSONA_REVISION) are received but log-only today —
   recompile-on-revision is TODO.md M-3.
6. **Every brain call** — mesad prepends the sealed prose card (never floats,
   never bands, never the archetype tag) to Decide/Act; genesis reads persona +
   history + relationships to compile the session apparatus
   (`mesa/mesad/genesis.go`).

The compute split in one line: **mesa owns WHO a host is** (storage, cook,
revisions); **the host compiles the persona the rest of the way locally**
(the pearl table is func-valued and cannot cross the wire).

---

## 14. A worked example (full persona, real schema)

"Cautious Lumbridge fisher" — passes `Validate()` as written:

```jsonc
{
  "schema_version": 1,
  "cornerstone": {
    "identity": {
      "name": "marn_fishes",
      "backstory": "Quiet regular who's spent months at the Lumbridge swamp. Got burned by a fake-trade once and never forgot it; happy to chat but won't risk much.",
      "north_star": { "theme": "skill_mastery",
                      "statement": "Hit 99 fishing without ever getting scammed again",
                      "horizon": "open", "success_signals": ["fishing_level == 99"] },
      "voice": { "register": "casual", "formality": "casual",
                 "tics": ["gl", "nice one"], "typo_feel": "occasional" },
      "archetype_tag": "cautious_social_grinder"
    },
    "hexaco": { "H": {"mu":0.78,"band":"high"},     "E": {"mu":0.66,"band":"mid_high"},
                "X": {"mu":0.55,"band":"mid"},      "A": {"mu":0.72,"band":"high"},
                "C": {"mu":0.81,"band":"high"},     "O": {"mu":0.40,"band":"mid"} },
    "values": { "north_star_value": "security", "secondary_value": "achievement" },
    "prefs": {
      "patience":      { "mu": 0.82, "band": "high" },
      "loss_aversion": { "mu": 2.6,  "band": "high" },
      "coop_type": "conditional_cooperator",
      "risk": { "economic": "low", "bodily": "low", "social": "mid" },
      "attention": { "anchor": 0.75, "level": "focused" },
      "curiosity": { "social": 0.25, "spatial": 0.05, "skill": 0.6, "economic": 0.1, "risk": 0.0 },
      "aggression":        { "mu": 0.15, "band": "very_low" },
      "decisiveness":      { "mu": 0.45, "band": "mid" },
      "tenacity":          { "mu": 0.75, "band": "high" },
      "self_preservation": null,
      "bulk_apperception": { "mu": 0.55, "band": "mid" }
    },
    "directives": [
      { "priority": 1, "subject": "self", "predicate": "attack", "object": "stronger_player", "hard": true }
    ],
    "quirks": [
      { "id": "swamp_loyalist", "origin": "derived", "domain": "movement",
        "trigger": "pre_action:fish", "binding": "fishing_spot_preference",
        "relation": "prefers", "object": "area:Lumbridge_swamp",
        "strength": "strong", "observable": true, "suppress_when": "low_hp",
        "narrative": "Always fishes the swamp, never the river." },
      { "id": "scam_wary", "origin": "idiosyncratic", "domain": "trade",
        "trigger": "trade_request", "binding": "trade_screening",
        "relation": "distrusts", "object": "player_type:stranger",
        "strength": "moderate", "observable": true, "suppress_when": "none",
        "narrative": "Screens every trade from someone it doesn't know." }
    ],
    "reverie": { "jitter": {}, "drift_log": [] },
    "pinned": [ { "summary": "Lost my first lobster stack to a fake-trade by a stranger.", "weight": 0.9 } ],
    "generation_meta": { "cohort_id": "docs-example", "archetype": "cautious_social_grinder",
                         "sampler_version": "hand-authored", "born_at": "2026-06-10T00:00:00Z",
                         "llm_materialized": false }
  },
  "trajectory": { "mood": {}, "sub_goals": [], "skill_logits": {}, "risk_drift": {} }
}
```

Note how the numbers *cohere*: high-H + high-C + high-λ + low risk + a scam
pinned memory all point the same direction — and `CompilePolicy` turns exactly
that coherence into rules (`screen_trades` fires from the H/λ draw;
`bank_when_full` from C; `no_attack_stronger`+`wont_strike_first` from the
very_low aggression; the directive hard-vetoes attacking up). For the opposite
pole, generate a fighter drone (`dronegen -mode fighter`) and diff: low-H,
free-rider, high aggression/bodily-risk, `frequent` typos. Same fields, utterly
different host. An incoherent draw (high-H + free-rider + scam quirks) is the
thing the sampler's forbidden pairs will reject — until then, don't author one.

---

## 15. Authoring checklist & anti-patterns

### Checklist (per persona / per archetype)
- [ ] Every dial authored as a **word** from `enums.go` (run `Validate()` —
      `personacook` does it for free even without an API key).
- [ ] `north_star` is a *motivator*, not an optimization target; specificity in the middle.
- [ ] HEXACO draw **coheres** with values/prefs (no high-H scammer).
- [ ] `H` set deliberately — it's the trade-screening trigger AND the ethics signal.
- [ ] `risk` is **domain-split** (econ/bodily/social), not one scalar.
- [ ] `attention` (short) and `patience` (long) set independently.
- [ ] `curiosity` flavored to make this host *diverge* from archetype-mates.
- [ ] Directives reserved for the few inviolable lines (they compile to hard vetoes).
- [ ] **2–7 quirks**, all `observable`, all bound to a mechanic, `suppress_when` said explicitly.
- [ ] **2–5 pinned memories** that *explain* the authored traits; not more (cache cost).
- [ ] `voice` varied hard from other archetypes (`typo_feel`, tics, formality) — believability.
- [ ] Did NOT author any `trajectory` field (the runtime owns it).
- [ ] `generation_meta` filled (cohort = the launch batch, archetype = the template).

### Anti-patterns
- **Over-specified → robotic.** Quirks/north-stars that dictate *actions* ("walk
  to Falador at 9am") not *priorities*. Constrain style, not steps.
- **Too consistent → uncanny.** Real people are inconsistent. That's what the
  reverie jitter and mood drift are for — leave some slack.
- **Cohort = personality.** A cohort is a launch batch (§0). Personality lives in
  the archetype; any host can do any RSC activity — the persona changes *how*
  and *how often*, not *what's possible*.
- **Stable forever.** Don't fight the Trajectory. A "trusting" host that gets
  scammed three times *should* drift cautious — that's the research, not a bug.
- **Flavor-text quirks.** "Is nostalgic" is not a quirk. If the server can't
  observe it, it doesn't exist.
- **Visible archetype_tag.** Never let the in-character brain see the label;
  it'll role-play the word instead of behaving from traits.
- **Numbers in prose.** The card must never leak a float, a band word, or
  "HEXACO" — the cook's judge scores `no_leakage`; keep hand-written backstory
  seeds clean too.

---

## 16. Open items

All persona open work is tracked in [`TODO.md`](TODO.md) — see **P-1** (open
decisions A2–A15; A4 quirk-compiler causality + A5 closed registries are the
blockers), **P-2** (the paired authoring sessions: archetype table, name pools,
north-star catalog, reverie catalog, population mix), **P-3** (the offline
genpop sampler + drift rules), **P-4** (the `compile.go:33` band→policy
residue: EventPolicy/ChoiceWeights/soft directives/ReverieKernel), **P-5**
(doc-consistency E items), and **P-8** (cornerstone_hash + revision integrity).

## 17. Cross-references
- [`persona/enums.go`](../persona/enums.go) — **the enum SSOT** (this doc is its prose mirror).
- [personas.md](personas.md) — the narrative: why this schema, the research framing.
- [persona-compile.md](persona-compile.md) — the prose cook (best-of-N + judge, seal-once).
- [persona-schema.md](persona-schema.md) — the schema decision record (tracked copy; promoted from `_research/reference/decision-persona-schema.md` 2026-06-10).
- [`_research/quantitative-persona-models.md`](_research/quantitative-persona-models.md) — per-model ADOPT/ADAPT verdicts + the drift math.
- [`_research/social-graph-and-trust-ledger.md`](_research/social-graph-and-trust-ledger.md) — the trust-ledger design (`limbic/ledger.go` is the build).
- [`_research/host-bootstrap-and-knowledge-gating.md`](_research/host-bootstrap-and-knowledge-gating.md) §7 — the attention/curiosity temperament-dial design.
- [reveries.md](reveries.md) — the reverie catalog (Phase-5 spec).
- [mesa.md](mesa.md) — the `personas` table (jsonb + prose_card + projection columns).
