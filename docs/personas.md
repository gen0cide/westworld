# Personas, archetypes, and north stars

> **STATUS: BUILT (schema, validation, compiler, storage, provisioning) — population
> authoring still open.** Verified 2026-06-10 against branch HEAD `0bfa818`. The
> `persona/` package is the schema SSOT: `persona.go` (Cornerstone/Trajectory
> structs), `enums.go` (every authoring enum, the words-not-numbers rule),
> `validate.go` (`Persona.Validate()`), `compile.go` (PersonaCard / ProseCook /
> `Project`), `render.go` (deterministic prose floor), `policy.go`
> (`CompilePolicy` → `pearl.Table` + TradePolicy + decision floor). mesad stores
> personas in Postgres (`mesa/mesad/ltm.go:200`) and serves them to hosts at
> startup (`runtime/runhost_bootstrap.go` `provisionPersona`). What is still
> design: the population-authoring paired sessions (archetype prototype table,
> cohorts file, north-star catalog, reverie catalog, 500-host mix) and the DSL
> persona-reflex tier — those sections are labeled **DESIGN (not built)** below.
> Open work lives in [`TODO.md`](TODO.md) (P-1…P-8, DSL-13, M-3/M-4).

Detail docs: [persona-authoring.md](persona-authoring.md) (the per-field
dictionary), [persona-compile.md](persona-compile.md) (the prose cook),
[`persona-schema.reference.yaml`](persona-schema.reference.yaml) +
[`host-persona.template.yaml`](host-persona.template.yaml) (full shapes),
[`personas/dolores.json`](personas/dolores.json) (a real stored persona). The
open-decisions worksheet this doc used to defer to is archived at
[`archive/initial-brainstorming/persona-open-decisions.md`](archive/initial-brainstorming/persona-open-decisions.md),
with A1's verbatim answer (cohort = launch batch) preserved as primary source.

## The pipeline, end to end (BUILT)

1. **Author in words.** A human (or a generator like `cmd/dronegen`, which emits
   validated drone persona files round-robin across sub-archetypes) writes a
   persona JSON using only the closed enum sets in `persona/enums.go` — the
   6-step `Band` ladder for every magnitude dial, Schwartz value names, voice
   formality/typo-feel, attention's own 5-step ladder, curiosity flavors, quirk
   sub-enums. Numbers (`mu`) are sampled, never authored. `Validate()` checks
   every enum and returns one joined error listing each miss with its valid set.
2. **Cook the prose card (optional, offline).** `mesa/personacook` runs a
   best-of-N Opus cook + judge against a stored persona (the full JSON goes into
   the prompt, so quirk narratives and pinned memories reach the card); the
   operator seals the winning second-person card into `identity.backstory`
   (personacook prints candidates + the judge pick — it does not write back;
   automating that loop is TODO P-8). `persona.DeterministicCook` / `Render()`
   is the no-LLM floor — un-cooked hosts still get a faithful-if-flatter card;
   `generation_meta.llm_materialized` records which path produced it. See
   [persona-compile.md](persona-compile.md).
3. **Register with mesa.** `mesa-ctl persona put|import|ls|get|set|rm` manages
   the registry (`cmd/mesa-ctl`); mesad persists to the Postgres `personas`
   table — `host_id` PK, canonical `persona_json` jsonb, derived `prose_card`,
   `cooked` flag, plus ~30 GENERATED facet columns (every band/dial extracted
   from the JSON, NULL-tolerant) so the fleet is queryable by SQL with zero app
   extraction (`mesa/mesad/ltm.go:200-300`).
4. **Provision at host startup.** A mesa-linked host pulls its authoritative
   persona via `Provision.Fetch` (`mesa/mesad/server.go:592`, wire shape
   `mesa/client/provision.go` — persona JSON + prose + goals); offline hosts
   load a local file via `-persona`. Both paths funnel into `applyPersona`
   (`runtime/runhost_bootstrap.go:503`), the sole chokepoint.
5. **Compile locally, then discard.** `persona.CompilePolicy` (deterministic, no
   LLM — the compiled table is func-valued and cannot cross the wire) turns
   disposition into the host's executable policy; `applyPersona` wires the pearl
   table + decision floor onto the host, sets the limbic affect baseline from
   `trajectory.mood`, captures the curiosity vector for decision-time
   explore/exploit weighting, keeps the north-star statement as the goal
   fallback (used when `-goal`/genesis produce nothing, and as the advancement
   fallback when the active goal closes), and keeps a one-line
   `personaSnippet` for the reactive extractor (`runtime/reactive.go:464`).
   The persona struct itself is then discarded — the runtime never writes it.
6. **Re-inject on every inference.** The brain never sees the struct — only the
   prose card. mesad prepends `e.prose` to every LLM seam: Act
   (`mesa/mesad/act.go:28`, system blocks = cached DSL manual + uncached
   `# YOUR CHARACTER` block), Decide (`server.go:741 decideSystem`), Chat/Ask
   (`act.go:370`), genesis (`genesis.go:55`), and the crons (`cron.go:297`,
   `cron_insight.go:223`). This is the Frisch & Giulianelli re-injection
   discipline, manifested. (The persona block is not yet a cached prefix —
   TODO P-1 A13.)

## The schema (BUILT) — Cornerstone vs Trajectory

The immutable/mutable contract survived from the design intact:

- **IMMUTABLE — the Cornerstone.** Identity (name, sealed backstory card, north
  star, voice), HEXACO bands, Schwartz top-2 values, the `prefs` dial block,
  Directives, Quirks, the reverie seed, pinned foundational memories, and
  out-of-band `generation_meta`. Sealed at birth; operator-only edits go through
  `mesa-ctl persona set` (validated). H (honesty-humility) remains the
  load-bearing ethics trait — `CompilePolicy` reads it for trade screening,
  fairness threshold, and scam propensity.
- **MUTABLE — the Trajectory.** The runtime-owned snapshot: `mood` (the
  valence/arousal/stress/confidence `Affect` vector — the live copy is
  `runtime/limbic.go`, seeded from the persona at load), `sub_goals`,
  `skill_logits`, `risk_drift`, and the single `learned_emergent` promotable
  quirk slot.

The real shape (`persona/persona.go`):

```jsonc
{
  "schema_version": 1,                  // CurrentSchemaVersion guard
  "cornerstone": {
    "identity": { "name", "backstory",  // backstory = the sealed prose card
                  "north_star": { "theme", "statement", "horizon", "success_signals" },
                  "voice": { "register" /* free text */, "formality", "tics", "typo_feel" },
                  "archetype_tag" },    // debug-only; invisible to the brain
    "hexaco": { "H|E|X|A|C|O": { "mu", "band" } },
    "values": { "north_star_value", "secondary_value" },   // Schwartz top-2, must differ
    "prefs": {
      "patience": {..}, "loss_aversion": {..},             // band Traits (λ mu ~1..3)
      "coop_type": "conditional_cooperator|free_rider|altruist",
      "risk": { "economic", "bodily", "social" },          // domain-split bands
      "attention": { "anchor", "level" },                  // its own 5-step ladder
      "curiosity": { "social","spatial","skill","economic","risk" },  // flavor vector
      // additive Westworld-inspired dials (all band Traits):
      "aggression": {..}, "decisiveness": {..}, "tenacity": {..},
      "self_preservation": {..},        // OPTIONAL — derived from λ + bodily risk when omitted
      "bulk_apperception": {..}         // cognitive maturity / local↔mesa reliance
    },
    "directives": [ { "priority","subject","predicate","object","hard" } ],
    "quirks":  [ /* typed executable modifiers; closed sub-enums, open object grammar */ ],
    "reverie": { "jitter", "drift_log" },   // per-host reverie seed (engine not built yet)
    "pinned":  [ { "summary", "weight" } ],
    "generation_meta": { "cohort_id","archetype","sampler_version","born_at","llm_materialized" }
  },
  "trajectory": {
    "mood": { "valence","arousal","stress","confidence","updated_at" },
    "sub_goals": [], "skill_logits": {}, "risk_drift": {},
    "learned_emergent": null            // single promotable quirk slot
  }
}
```

### Deltas from the designed sketch (annotated, not silent)

The schema as built diverges from the 2026-06-01 design in this doc's previous
revision. The diffs:

| Designed | Built |
|---|---|
| `values.higher_order` (4 higher-order Schwartz weights) | **Dropped.** `Values` is the top-2 anchor only (`persona.go:70`). |
| Fat Trajectory: `hexaco_state`, `value_weights`, `patience_cur`, `loss_aversion_cur`, `attention_cur`, `curiosity_cur`, `cooperation_type`, `reverie_jitter`, `trust_ledger` pointer, `friendships_rivalries` | **Trimmed** to mood + sub_goals + skill_logits + risk_drift + learned_emergent. The anchor/current drift machinery and the ~30s mesa flush never manifested; mood lives in `runtime/limbic.go`. The trust-ledger pointer never manifested either: relationships are a mesad table, and pearl's `Facts.Counter.Trust` slot is deliberately left nil at assembly (`runtime/pearl.go:20`) — trust-dependent rules simply don't fire yet. Drift update rules remain open (TODO P-3). |
| `attention_anchor`/`curiosity_anchor` as top-level Cornerstone fields | Moved into `prefs` as `Attention{Anchor,Level}` + the `Curiosity` vector. |
| Numbers authored directly (`"mu": 0.84`) | **Words only**: every magnitude dial is authored as a `Band` word; `mu` is sampler output (`enums.go` golden rule). |
| — (absent) | **Added: five additive dials** — `aggression`, `decisiveness`, `tenacity`, `self_preservation` (optional/derived), `bulk_apperception` (`enums.go:93-124`). |
| — (absent) | **Added: `Directives`** — inviolable core rules; hard ones compile to max-salience pearl vetoes (`policy.go:156-161`); soft→bias compile still TODO (`policy.go:192`). |
| — (absent) | **Added: `GenerationMeta`** — cohort/archetype/sampler audit trail, out-of-band, invisible to the host. |
| Stored in "mesa `hosts.persona` / `bots.persona` JSONB" | Stored in mesad's **`personas` table** (`persona_json` jsonb + `prose_card` + `cooked` + generated facet columns). |

The full quantitative derivation behind the schema (per-model verdicts on
HEXACO/Schwartz/the econ anchors, update-rule sketches) lives in the
`_research` notes: `_research/quantitative-persona-models.md`,
`_research/host-bootstrap-and-knowledge-gating.md` §7/§7.5, and
`_research/social-graph-and-trust-ledger.md`.

## Cohort vs archetype — the A1 reframe

The previous revision of this doc used "cohort" to mean a personality cluster
("Lumbridge regulars", "Wilderness PKers"). Alex's A1 answer (preserved verbatim
in the archived worksheet) reframed it:

- **Cohort = launch batch.** A deployment/experiment grouping — "we released
  another N hosts, maybe with an experimental tweak, and want to track them."
  It lives in `generation_meta.cohort_id`, is analytics-only, and is invisible
  to the host.
- **Archetype = personality.** The underlying trait-prior bundle a persona is
  sampled from. It lives in `generation_meta.archetype` (plus the debug-only
  `identity.archetype_tag`); the archetype *registry* (prototype table, priors,
  forbidden pairs) is the open paired session — TODO P-2.

Everywhere the old text said "cohort taxonomy", read "archetype taxonomy".
`cmd/dronegen`'s sub-archetypes (loudmouth / quiet tagalong / haggler /
wide-eyed tourist) are the first working example of the pattern.

## The deterministic policy compile (BUILT)

`persona.CompilePolicy` (`persona/policy.go`) is the CODE-food half of the
compiler — a pure function of the persona, no LLM, run host-side at every load.
It produces:

- **`pearl.Table`** — the always-on reflexes the pearl engine runs at the gate
  and decide seams: `flee_when_hurt` (threshold from emotionality +
  self-preservation), `no_attack_stronger` / `wont_strike_first` (aggression),
  `greet_stranger` (extraversion), `screen_trades` (honesty/agreeableness/λ),
  `bank_when_full` (conscientiousness), `stay_on_task` (patience), plus one
  max-salience veto per hard Directive.
- **`TradePolicy`** — fairness threshold, scam propensity, risk aversion (from
  H, coop type, λ, economic risk).
- **`DecisionFloor`** — the pearl confidence floor from decisiveness (a decisive
  host acts locally on thinner evidence).

Not yet compiled: quirks (stored + validated + fed to the prose cook, but no
quirk→rule mapping — blocked on the A4/A5 operator calls), soft directives,
and the PersonaCard's `EventPolicy`/`ChoiceWeights`/`ReverieKernel` slots
(`compile.go:33` TODO). See TODO P-1/P-4.

## What needs to be designed together

> **DESIGN (not built).** These are the open paired Alex+Claude sessions that
> *fill in* the built schema — tracked as TODO **P-1** (decisions A2–A15; A1 is
> answered) and **P-2** (authoring B1–B5). The sketches below are the live spec
> for those sessions, kept here on purpose.

Alex's charter for this work, from the architecture conversation (primary
source, kept verbatim):

> "This is where the reveries come into play and I'm going to work deeply with you on this — we will have to LLM generate, but we need to come up with cohorts, personas, 'north stars', etc. Please save this entire LOE as something we can pair on once we've finished answering all the questions, but the reality is it'll be LLM and hand authored as a pairing. I know we'll probably have to start with a handful, but the goal is we have a population of 500 hosts."

### 1. Archetype taxonomy (was "cohorts")

Hand-authored personality clusters to seed the prototype table (~13 per the
worksheet), each carrying trait priors the sampler draws from:

- **Lumbridge regulars**: chatty, social, low-skill, friendly to newcomers
- **Edgeville merchants**: transactional, calculating, banking-focused, tolerant of risk
- **Varrock loiterers**: ambient socializers, idle frequently, light-skill diversity
- **Wilderness PKers**: aggressive, suspicious, high-combat focus
- **Solo grinders**: focused, less social, methodical
- **Newbie tourists**: chaotic, exploratory, occasionally helpless

Open: archetype count at population scale, priors + forbidden trait pairs per
archetype, and the cohorts file (archetype mix + flavor: home, name pool,
activity bias) — reframed per A1 so cohorts stay mix+flavor, never personality.

### 2. The genpop sampler

The offline priors→rejection→quirks→seal pipeline that turns an archetype draw
into a unique validated persona at scale (TODO P-3). `cmd/dronegen` is the
hand-rolled precursor; `mesa/personacook` covers the prose half. The
generation pipeline design from the old revision ("one-time LLM call at
registration, cached in mesa") shipped in spirit — cook offline, register via
`mesa-ctl`, mesad is the system of record.

### 2a. Persona-level default handlers

> **DESIGN (not built) — but the first slice ships today via pearl.** The
> designed mechanism: each persona declares default DSL handlers for every
> supported event — always-on reflexes that fire regardless of the running
> routine, which routines temporarily override (`extends host` + `super()`, the
> 3-tier stack: persona base reflex → learned specialization → routine
> override). The DSL parser/validator still reject `extends host`; this is
> **TODO DSL-13** (task #93), un-parked now that the schema it needed exists.

What exists today instead: `CompilePolicy`'s pearl rules *are* persona-level
always-on reflexes — `flee_when_hurt` / `no_attack_stronger` / `greet_stranger`
fire at the gate/decide seams regardless of routine, and materially differ
between a cautious fisher and a reckless PKer (the exact differentiation the
old DSL sketches illustrated). The DSL-visible handler tier remains the design
goal for reflexes that need full routine expressiveness.

### 2b. Persona evolution

> **DESIGN (mostly not built).** The slot exists: `trajectory.learned_emergent`
> holds one promotable quirk earned through experience, and operator edits flow
> through `mesa-ctl persona set`. Not built: the `persona_revisions` table, the
> LLM-generated revision path (rare Sonnet calls with rationale), recompile-on-
> revision (today a persona change requires a host restart — live
> `PERSONA_REVISION` push is TODO M-3), and the seal-integrity guard
> (`cornerstone_hash` exists only as a comment, `persona/policy.go:262` — TODO
> P-8). The principle stands: fast mood churn never creates a revision; a value
> re-rank or a new emergent quirk does.

### 3. North stars

Built as schema: `identity.north_star` {theme (7-value enum), statement,
horizon, success_signals}, and live in the runtime — the statement is the
host's goal fallback and the genesis seed. The *catalog* (which north stars,
per archetype, plus conflict-resolution rules) is open paired work (TODO P-2).

North stars are NOT optimization targets — they're psychological motivators. A
host with a "become wealthy" north star still takes time to chat with friends,
just like a real player. But over weeks, its strategic choices accumulate in a
direction.

Open design questions: specificity (too specific = robotic, too vague =
unactionable); whether north stars can evolve; how conflicting drives resolve
(a "wealth" north star with a shy temperament should still socialize, just
less).

### 4. Reverie catalog (cross-references [reveries.md](reveries.md))

The schema slot is built (`cornerstone.reverie` ReverieSeed: per-reverie jitter
+ drift log) but nothing consumes it — the reverie engine and the catalog
itself (which gestures, default weights per archetype, validated for
server-visibility) are explicitly deferred in `enums.go` and tracked as TODO
P-2/P-6.

### 5. Population mix targeting

For 500 hosts, what archetype distribution produces interesting social
dynamics? Working hypothesis (unchanged from the original design): skew toward
"casual normal" archetypes with a long tail of extremes — ~60% casual
social/skill-focused, ~20% focused/transactional, ~10% newbie/exploratory,
~10% extreme (PKers, fanatics, oddballs). Most interactions happen between
"normal" hosts; "extreme" hosts add color and create observable incidents.
Open until the archetype table exists (TODO P-2).

## Why this matters

The persona system is the **substrate** that makes the research goals possible:

- **Long-term strategic accomplishment**: hosts pursue north stars over weeks.
  Without north stars, hosts are reactive — never building toward anything.
- **Community formation**: hosts of compatible archetypes gravitate together;
  clashing archetypes produce friction. The archetype taxonomy is the source of
  social structure.
- **Ethics observation**: ethical choices correlate with persona traits — H is
  wired straight into trade screening and scam propensity today. Observable
  patterns require persona variation.
- **Believability**: a population of identical hosts is detectable. Persona
  variation creates the diversity that makes individual hosts believable as
  humans.

Get personas wrong → 500 indistinguishable agents. Get them right → a
believable society.

## Anti-patterns to avoid

- **Too detailed personas**: overspecifying makes hosts robotic ("you always do
  X"). Personas should constrain *style* and *priorities*, not specific actions.
- **Too consistent personas**: real humans have inconsistencies. The quirk slots
  exist for "occasional surprising trait" texture.
- **Archetype = role**: an archetype isn't a job assignment. Hosts of any
  archetype can do any RSC activity; archetype affects *how* and *how often*,
  not what's possible.
- **Stable forever**: personas should be revisable as evidence accumulates. A
  "trusting" persona that gets scammed three times should adapt (the evolution
  machinery above is the open slice of this).
- **Leaking CODE-food to the brain**: bands, mus, and the archetype tag never
  reach the model — only the prose card does (`render.go` is deliberately
  number-free). Keep it that way.

---

Open work: see [`TODO.md`](TODO.md) — persona section P-1…P-8 (operator
decisions A2–A15, authoring sessions B1–B5, genpop sampler, `compile.go:33` /
`policy.go:192` code TODOs, seal integrity), DSL-13 (persona reflex tier),
M-3/M-4 (live revision push, recook/rollback tooling).
