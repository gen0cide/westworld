# Decision Record — The Persona Schema (Cornerstone / Trajectory / Pearl-compile)

> **PROMOTED 2026-06-10** — verbatim copy of
> `docs/_research/reference/decision-persona-schema.md`, the persona-schema
> decision record (Cornerstone/Trajectory split, PersonaCompiler, "brain sees
> prose only"). Promoted to a tracked path per the docs-grounding audit:
> `docs/_research/` is gitignored, and the links from tracked
> `persona-authoring.md` / `persona-compile.md` / `persona-schema.reference.yaml`
> broke on fresh clones. Verified against code 2026-06-10.

> STATUS: PROPOSED (2026-06-01). Scope: the persona schema only — the data
> model, the immutable-vs-mutable split, and the field→code compilation that
> turns persona into runtime behavior. Replaces the placeholder
> `cognition.Bundle.Persona map[string]string`.
>
> Sources synthesized: claude-persona (immutable/mutable rigor, HEXACO,
> quirks-as-executable, reveries-as-weights, external-sampling), gpt55-persona
> (categorical core + event-policy enums + compile-first + domain risk +
> trust-is-mutable), gemini Cornerstone (immutable core + pinned foundational
> memories). Grounded against the real code: `cognition.Bundle`,
> `brain.Situation`, `event.Bus`, `world.World`, `dsl/spec/actions.go`,
> empty `persona/` + `reveries/`, doc-only `mesa`.

---

## 1. Recommendation (the decision)

Build a **two-tier persona** with a **compile step** between storage and runtime:

1. **Cornerstone (immutable identity).** A sealed, versioned JSONB sub-object —
   sampled offline at host registration, never mutated by the runtime, edited
   only by a deliberate operator action recorded in mesa's `persona_revisions`.
   Holds: `identity`, `hexaco` (6 numeric trait means + bands), `values`
   (Schwartz north-star + secondary), domain-split risk, **derived quirks**,
   **reverie jitter seed**, and **pinned foundational memories**. This is the
   gemini "Cornerstone" — a host never forgets who it is.

2. **Trajectory (mutable adaptation).** Everything that carries visible change:
   trust ledger (the Relationship Matrix), skill-focus logits, sub-goals, mood
   (affect), drifted risk/patience scalars, and a single `learned_emergent`
   quirk slot. Lives in mesa relational/episodic tables + a small per-host
   in-RAM struct; slow re-rankings snapshot into `persona_revisions`.

3. **PersonaCard (the compiled artifact).** A one-time **PersonaCompiler**
   (gpt55's "missing compiler", gemini's "Pearl compiles, cradle runs") turns
   Cornerstone + a Trajectory snapshot into typed, deterministic artifacts the
   cradle executes with **no LLM round-trip**: a behavioral prose card for the
   brain prompt, an `EventPolicy` table, a `QuirkSet` of executable modifiers, a
   `ReverieKernel` of trait-derived weight functions, and `ChoiceWeights`. The
   compiled `PersonaCard` — not raw JSON — is what populates `Bundle.Persona`.

### Representation decision: hybrid, resolving the claude-vs-gpt55 split

- **Stored numerically + categorically** (claude): HEXACO as `{mu float, band string}`,
  risk as floats. Numbers exist for analytics/cohort experiments in delos; bands
  exist for engineer legibility and as the compile input.
- **Compiled to categorical policy + behavioral prose** (gpt55): the cradle's
  deterministic policies key off **bands/enums** (legible, debuggable); the brain
  prompt sees **only the rendered behavioral paragraph**, never raw floats or
  band labels.
- **Sampled externally, materialized by LLM only** (claude): the numeric vector
  is sampled in deterministic Go at registration; the LLM is called **once per
  ~50 hosts, offline** to render backstory/quirk-narrative/voice — never to
  invent traits, never on the hot path.

### Why this and not the alternatives

- **vs. "keep `map[string]string`":** the placeholder is causally inert —
  persona injected as prompt text alone collapses to prompt-only roleplay
  (gpt55's central warning). The compile step is what makes persona *causal* on
  the no-LLM hot path. We keep a flat `map[string]string` *view* for backward
  compat (the frozen `relation_with` reads `Persona["relation:NAME"]`), but it
  is now a **projection of the compiled card**, not the source of truth.
- **vs. claude's full 16-dim Gaussian-mixture + copula:** over-built for v1 and
  needs a numerical dependency. Ship a 13-archetype mixture with diagonal-ish Σ;
  the archetype mean does most of the work. Encode a few forbidden trait pairs
  (max-Honesty + max-Power) rather than fitting a full covariance matrix.
- **vs. gpt55's pure-categorical store:** loses the numeric substrate delos needs
  for flattening metrics and cohort science. Keep both; bands are derived from
  percentiles.
- **vs. claude's behavioral-econ parameter soup (λ, δ, envy/guilt) as required:**
  nothing consumes a prospect-theory evaluator today (the brain is
  `Decide(Situation)→Choice`, not a utility function). Store the `prefs` block in
  JSONB (cheap, future-proof) but **do not build the decision functions** until
  the LLM strategist exists. Flag as speculative.

---

## 2. Concrete interfaces (Go)

New package `persona/` (currently empty). All pure Go, no I/O, no CGo. The
compiled artifacts are **immutable value types shared read-only across the ~500
hosts in a process** (one `*PersonaCard` per host, hydrated at login).

```go
package persona // github.com/gen0cide/westworld/persona

// ---- Stored shape (mirrors mesa bots.persona JSONB; (de)serialized as JSON) ----

type Persona struct {
    SchemaVersion int          `json:"schema_version"` // guards doc/shape drift
    Cornerstone   Cornerstone  `json:"cornerstone"`    // immutable, sealed
    Trajectory    Trajectory   `json:"trajectory"`     // mutable snapshot
}

type Cornerstone struct {
    Identity  Identity            `json:"identity"`
    Hexaco    map[string]Trait    `json:"hexaco"`   // keys H,E,X,A,C,O
    Values    Values              `json:"values"`
    Risk      DomainRisk          `json:"risk"`     // domain-split (gpt55)
    Prefs     *EconPrefs          `json:"prefs,omitempty"` // SPECULATIVE; nil v1
    Quirks    []Quirk             `json:"quirks"`   // derived + idiosyncratic
    Reverie   ReverieSeed         `json:"reverie"`  // jitter vector, not weights
    Pinned    []FoundationalMemory`json:"pinned"`   // gemini Cornerstone
    Cohort    string              `json:"cohort_id"`
    Gen       GenerationMeta      `json:"generation_meta"`
}

type Trait struct {
    Mu   float64 `json:"mu"`   // 0..1 trait mean (frozen)
    Band string  `json:"band"` // "very_low".."very_high" (compile input)
}

type Identity struct {
    Name        string `json:"name"`         // == opts.Username partition key
    Backstory   string `json:"backstory"`    // rendered prose
    NorthStar   NorthStar `json:"north_star"`
    Voice       Voice  `json:"voice"`
    ArchetypeTag string `json:"archetype_tag"` // debug-only; INVISIBLE to brain
}

type NorthStar struct { // structured, feeds cognition.Retrieval.Goal (gpt55)
    Theme         string   `json:"theme"`
    Statement     string   `json:"statement"`
    Horizon       string   `json:"horizon"`        // "week"|"month"|"open"
    SuccessSignals []string `json:"success_signals"`// measurable from world+mesa
}

type Values struct {
    NorthStarValue string   `json:"north_star_value"` // Schwartz-10 categorical
    SecondaryValue string   `json:"secondary_value"`
}

type DomainRisk struct { // gpt55: never a single risk_tolerance
    Economic string `json:"economic"` // ordered band
    Bodily   string `json:"bodily"`
    Social   string `json:"social"`
}

// Quirk is an EXECUTABLE modifier, not flavor text (claude+gpt55).
// It compiles into a QuirkSet the cradle applies WITHOUT an LLM call.
type Quirk struct {
    ID          string  `json:"id"`
    Origin      string  `json:"origin"`   // "derived"|"idiosyncratic"|"learned_emergent"
    Domain      string  `json:"domain"`   // "movement"|"social"|"trade"|"combat"|"banking"|"idle"
    Trigger     string  `json:"trigger"`  // an event Kind() (e.g. "trade_request") or "on_encounter"|"pre_action:bank"
    Binding     string  `json:"binding"`  // GAMEPLAY binding enum; rejects non-operational quirks
    Relation    string  `json:"relation"` // "prefers"|"avoids"|"distrusts"|"delays"
    Object      string  `json:"object"`   // referent (e.g. "bank_node:Falador_west", "equipment:bronze_full")
    Strength    float64 `json:"strength"` // 0..1 intensity
    Observable  bool    `json:"observable"`// reject server-invisible quirks
    SuppressWhen string `json:"suppress_when"`// e.g. "in_combat" (gpt55: load-bearing)
    Narrative   string  `json:"narrative"`// LLM-rendered; for the prose card only
}

type ReverieSeed struct { // claude: weights are NOT stored, jitter is
    Jitter map[string]float64 `json:"jitter"` // per-reverie r_a ~ LogNormal(0,0.25), sampled once
    DriftLog []ReverieDrift    `json:"drift_log"`
}

type FoundationalMemory struct { // gemini Cornerstone, pinned every brain call
    Summary string  `json:"summary"`
    Weight  float64 `json:"weight"`
}

// ---- Mutable Trajectory (snapshot of state that lives authoritatively in mesa) ----

type Trajectory struct {
    Mood        Affect             `json:"mood"`         // valence/arousal/stress/confidence
    SubGoals    []string           `json:"sub_goals"`
    SkillLogits map[string]float64 `json:"skill_logits"` // soft-RL focus
    RiskDrift   map[string]float64 `json:"risk_drift"`   // drift around frozen anchors
    Emergent    *Quirk             `json:"learned_emergent,omitempty"` // single promotable slot
    // Trust ledger / relationships are large + per-counterparty: NOT inlined here.
    // They live in mesa.relationships and are projected per-decision (see §4).
}

type Affect struct { // gemini affect, scaled down: no biochem, lazily decayed
    Valence    float64   `json:"valence"`    // -1..1
    Arousal    float64   `json:"arousal"`    // 0..1
    Stress     float64   `json:"stress"`     // 0..1
    Confidence float64   `json:"confidence"` // 0..1
    UpdatedAt  time.Time `json:"updated_at"` // for lazy decay-on-read
}

// ---- The COMPILED artifact (the thing the runtime actually executes) ----

type PersonaCard struct {
    HostName     string             // == opts.Username
    Prose        string             // behavioral paragraph — the ONLY thing the brain sees
    EventPolicy  map[string]string  // event Kind() -> policy enum (gpt55 compile target)
    Quirks       QuirkSet           // pre-indexed by trigger for O(1) lookup
    Reverie      ReverieKernel      // trait-derived weight closures + cooldowns
    ChoiceWeight map[string]float64 // option-bias hints for decide()
    Directives   []Directive        // hard rules for the future brain.Guard
    Pinned       []string           // foundational memory summaries (always prepended)
    Affect       Affect             // current snapshot (read into the brain call)
    CompiledAt   time.Time
}

type QuirkSet struct{ byTrigger map[string][]Quirk } // built from Cornerstone.Quirks + Trajectory.Emergent
func (q QuirkSet) For(trigger string) []Quirk

// ReverieKernel: claude's linear forms as closures. Pure, in-process, no I/O.
// Weight(id) = base + Σ coeff·trait + Σ coeff·state, × per-host jitter r_a.
type ReverieKernel struct {
    weights map[string]func(state ReverieState) float64
    cooldowns map[string]time.Duration
}
func (k ReverieKernel) Weight(id string, st ReverieState) float64

type Directive struct{ Priority int; Rule string } // for brain.Guard (future)

// ---- The compiler (runs at registration + on rare revision; NOT on hot path) ----

type Compiler interface {
    Compile(p Persona, snap Trajectory) (*PersonaCard, error)
}

// Render produces the behavioral prose from numeric/categorical fields.
// In production this is the offline batched LLM materialization; the default
// impl is a deterministic template (keeps `go test` green with no API key).
func (p Persona) Render() string

// Facet/Band accessors the DSL/runtime read directly (no LLM).
func (p Persona) Facet(name string) string // e.g. "diligence" -> band
```

### EventPolicy enum vocabulary (the persona→DSL bridge — gpt55's strongest idea)

The compiled `EventPolicy` map keys are **real `event.Kind()` strings** (verified
against `event/events.go` + the `on <event>` surface in `docs/lang/api.md`):
`trade_request`, `chat_received`, `private_message`, `other_player_damage`
(attacked), `death`, `xp_gain`/`experience_gain` (level/celebrate), `npc_killed`.
Values are a small enum that resolves to **version-controlled DSL handler
templates** (the unbuilt "persona-tier default handlers"):

```
on_trade_request:    screen_then_engage | decline_strangers | eager_accept
on_chat_received:    greet_back | ignore | greet_if_social
on_other_player_damage: eat_then_flee | retaliate | flee_immediately
on_insult (chat):    ignore_then_note | clap_back | de_escalate
on_xp_gain (level):  pause_and_celebrate_if_social | continue_silent
```

### Reverie package (currently empty `reveries/`)

```go
package reveries // github.com/gen0cide/westworld/reveries
// Catalog lives in CODE (gpt55); persona stores only jitter + bands (claude).
// Reject any reverie not server-visible/socially-inferable (claude/gemini).
type Reverie struct {
    ID, Domain string
    SuppressWhen string            // never mid-combat (gpt55)
    DefaultBand string             // off|trace|rare|occasional|common|signature
}
func Catalog() []Reverie
// Weight is the claude linear form realized via the compiled ReverieKernel:
//   examine_scenery_w = 0.05 + 0.25*O + 0.10*(1-C) + 0.20*novelty, × r_a
// Pure in-process, host-local — NO mesa/brain round-trip.
```

---

## 3. Data model (concrete)

### mesa `bots.persona` JSONB (the doc-only schema, made concrete)

The `persona` JSONB column (mesa.md line 62) holds the full `persona.Persona`
serialization. Immutable vs mutable split *inside* the blob:

| JSON path | tier | mutated by | persistence |
|---|---|---|---|
| `cornerstone.*` | **immutable** | operator only, via `persona_revisions` | sealed; never written at runtime |
| `cornerstone.pinned[]` | immutable | operator only | always prepended to Bundle |
| `trajectory.mood` | mutable (fast) | Limbic (bus subscriber) | in-RAM; flushed to mesa on ~30s cadence |
| `trajectory.skill_logits` | mutable (slow) | soft-RL on XPGain | mesa, lazy |
| `trajectory.risk_drift` | mutable (slow) | drift around anchor | mesa, lazy |
| `trajectory.sub_goals` | mutable | strategist/reflection | `persona_revisions` on re-rank |
| `trajectory.learned_emergent` | mutable (rare) | quirk promotion (12/14 trips) | `persona_revisions` |

The **trust ledger / Relationship Matrix is NOT in the blob** — it is large and
per-counterparty. It maps onto the existing `mesa.relationships` table
(`bot_id, other_username, trust_score REAL -1..1, relationship TEXT, …counts…,
notes`), keyed on player **`Name`** (the only durable identity per
`groundtruth-world-facts`; `PlayerRecord.Index` is volatile). claude's update
math instantiates it: per-other `Beta(α,β)` with H-shaped priors
(`α₀=2+4·H`, `β₀=2+4·(1−H)`), `α += w` on cooperation / `β += w` on defection,
`w∈[0.3,8]` by severity; `trust_score = α/(α+β)·2−1` projected into the existing
`REAL -1..1` column. No new shared-reputation table (honors "each host privately
models others").

### `persona_revisions` (already specced, mesa.md line 150)

Used **only** for: a Cornerstone edit (operator) OR a slow Trajectory re-rank
(value re-ranking, new sub-goal, emergent-quirk promotion). Each writes a full
`persona_snapshot JSONB` + `rationale`. Fast mood/affect churn does NOT create
revisions.

### Population/cohort generation artifact (offline, deterministic Go)

A static config: `[]Archetype{ Name, Pi, MuVector, Sigma(diagonal) }` (13
archetypes) + cohort overrides `{PiMultipliers, MuShifts}`. The generator:
`Cat(π) → N(μ_k, diag Σ_k) → clip/normalize → band → derive template quirks →
sample reverie jitter r_a → ONE batched LLM call per ~50 hosts for
backstory/quirk-narrative/voice → write JSONB`. No Python, no copula, no
embedding-dedupe in v1 (delos backlog).

### `cognition.Bundle` change (additive — frozen-contract-safe)

`Bundle.Persona map[string]string` stays (the `relation_with` builtin reads
`Persona["relation:NAME"]`), but is now **populated from the compiled card**:
`{"prose": card.Prose, "north_star": …, "mood": …, "relation:NAME": …}`.
Add three additive fields (do not change method signatures):

```go
type Bundle struct {
    // … existing fields …
    Card     *persona.PersonaCard // structured runtime-readable persona
    Pinned   []string             // foundational + reverie-resurfaced (gemini)
    Affect   *persona.Affect      // read into brain tier/temperature selection
}
```

---

## 4. Exactly how it integrates with the real code

The real seam is two interface fields on `*runtime.Host` plus the Bundle
(`groundtruth-cognition-brain-seam`). Nothing about the host loop changes; no
tick is added.

1. **Storage → card (compile at login, off hot path).** When a host logs in,
   the real `cognition.Client` (mesa-backed, behind the existing
   `Host.Retriever` field) fetches `bots.persona`, calls
   `persona.Compiler.Compile(p, traj)` ONCE, and caches the `*PersonaCard` in
   the per-host client state keyed by `Retrieval.HostName` (==`opts.Username`).
   The compiler runs in-process; **no LLM** (prose was rendered offline at
   registration and stored). This is the gemini "Pearl: mesa compiles, cradle
   runs" — the `PersonaCard` IS the host-local cognitive artifact.

2. **Card → Bundle (per strategist call).** Each `Retrieve` returns the cached
   card via `Bundle.Card` + flattens it into `Bundle.Persona` (the prose under
   `"prose"`, plus the `relation:NAME` entries from the trust ledger for any
   counterparty named in `Retrieval.Goal`). `Bundle.Pinned` = Cornerstone
   foundational memories (always) + any reverie-resurfaced episode. This is the
   read path that makes persona causal in the brain prompt; the brain sees
   **prose only** (the immutable block goes in a cache-prefixed system prompt —
   ~100% per-host cache hit since it never changes).

3. **EventPolicy / Quirks → DSL (the no-LLM fast path).** The compiled
   `EventPolicy` enums resolve to shared DSL handler templates registered for
   the `on <event>` surface (`dsl/spec/actions.go`). `QuirkSet.For(trigger)`
   is consulted by the runtime action methods (e.g. `WalkTo`/`bank` read a
   `movement/bank_node:Falador_west/prefers` quirk as a +travel-willingness
   bias) — pure Go, no round-trip. This is where persona is *causal* without
   the LLM. The card is read-only and shared-safe across the 500 hosts'
   goroutines.

4. **Trajectory mutation (Limbic bus subscriber — NEW, but no tick).** A new
   `event.Bus.Subscribe("*")` consumer (sibling goroutine to `heartbeatLoop`,
   per `groundtruth-host-loop-bus`) maps events → Trajectory deltas:
   `other_player_damage`→stress↑ + trust↓; `xp_gain`→confidence↑ + skill-logit
   update; trade-completed→trust↑. Mood decays **lazily on read** (timestamp
   diff × decay constant), not via a per-host daemon. Flushes to mesa on the
   documented ~30s cadence. **Mutating Trajectory NEVER touches Cornerstone** —
   the immutable/mutable contract is enforced by the type split (Cornerstone has
   no setters; the Limbic only holds a `*Trajectory`).

5. **Reverie weights (in-process, host-local).** `reveries.Catalog()` +
   `card.Reverie.Weight(id, state)` are pure functions; a future reveries hook
   (Phase 5, sibling goroutine) samples mode-first then expresses via existing
   high-level actions (an extra `WalkTo` to an adjacent tile), NOT tick-precise
   motor control (which doesn't exist).

6. **Directives → brain.Guard (future).** `card.Directives` is the stable rule
   source for the apperception veto gate, when that lands. Schema-ready now;
   not built in this decision.

### What does NOT change (honoring frozen contracts)

- `brain.Strategist.Decide` / `cognition.Client.Retrieve` signatures are
  byte-identical (additive Bundle fields only).
- `Bundle.Persona map[string]string` survives for `relation_with`.
- Stubs stay swappable: `persona` package ships a deterministic template
  `Render()` + a `Compiler` that needs no Anthropic/Postgres key, so `go test`
  stays green (the load-bearing stub contract).
- No new go.mod deps for the schema itself (pure Go + encoding/json). The LLM
  materializer + mesa/pgvector are the real Strategist/Client's deps, behind the
  seam — not the persona schema's.

---

## 5. Risks / open questions

- **Compiler is the highest-leverage unbuilt piece and the easiest to under-spec.**
  gpt55 names it but designs none; this record gives the artifact shapes but the
  exact band→policy mapping table and reverie coefficient set are still to be
  authored. Risk: if quirks/event-policies are too coarse, hosts read
  identically despite distinct Cornerstones (flattening leaks past the schema).
- **`learned_emergent` quirk promotion can drift identity** if uncapped. Enforce
  claude's cap (‖Δr_a‖ < 0.5 over a run; one emergent slot; log every promotion).
  Open: who runs the promotion detector (a reflection-cadence job in mesa?).
- **EventPolicy → DSL templates assumes persona-tier default handlers exist.**
  They don't yet. The enum vocabulary is the contract, but the template library
  is net-new DSL authoring work and must pass the spec↔`actionHandlers` parity
  test.
- **`prefs` (behavioral-econ) is dead weight until the LLM strategist computes a
  utility function.** Stored but unconsumed in v1. Risk of premature schema lock;
  keep it `omitempty` and explicitly speculative.
- **Trust-ledger projection into `Bundle.Persona["relation:NAME"]` is lossy.**
  Beta(α,β) → a single `-1..1` score → a string label. Fine for the current
  `relation_with` surface; the richer ledger should reach the brain via the new
  `Bundle.Card`/structured path, not the flat map, once the brain reads it.
- **Cache-prefixed immutable block depends on the (unbuilt) Anthropic strategist
  honoring prompt caching.** Believability needs re-injection every call; cost
  needs caching. This is a brain-integration concern flagged here, not solved.
- **Offline LLM materialization (1 call / 50 hosts) is a registration-time
  pipeline that does not exist yet.** Until built, `Render()` falls back to the
  deterministic template — acceptable for tests, flatter for believability.
- **Reverie catalog must be validated against `runtime/actions_*.go` + the render
  engine** before commit — any reverie the server can't observe is rejected
  (claude/gemini). Some catalog entries (camera_pan, minimap_glance) are likely
  UI-only and produce no server-visible effect.
- **Schema migration:** `schema_version` guards shape drift, but mesa needs a
  JSONB-tolerant read + per-version migration (mesa.md open question line 334).
