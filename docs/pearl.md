# Decision Record: The Policy / Quirk Engine

> **PROMOTED 2026-06-10** — verbatim copy of
> `docs/_research/reference/decision-policy-quirk-engine.md`, the design record for
> the built `pearl` package (rule table, predicate vocabulary, TryDecide + Gate
> seams, POLICY_VETO, compile-from-persona). Promoted to a tracked path per the
> docs-grounding audit: `docs/_research/` is gitignored, and the citation in
> `pearl/doc.go` broke on fresh clones. Verified against code 2026-06-10.

> STATUS: PROPOSED (2026-06-01). Scope = the executable policy/quirk layer that
> GATES and SHAPES host actions on the fast path. Resolves DECISION
> `policy-quirk-engine`.
>
> Codenames adopted: **Pearl** (host-local compiled cognitive core / fast path),
> **Cornerstone** (immutable persona core + directives the gate reads),
> **Reverie** (saliency-weighted bias, downstream consumer of the same policy
> table). The policy/quirk engine is the *interpreter* that runs the
> **Pearl-compiled** policy table.

---

## 1. The decision in one paragraph

Build a single per-host, in-process **`pearl.Engine`** that interprets a compiled,
flat **rule table** (`[]Rule`, where each `Rule = {When predicate, Then effect}`).
Each rule is a deterministic predicate over a small read-only fact view
(world mirror + affect vector + persona facets + relationship snapshot) producing
one of three effect kinds — **bias** (reorder/weight a `decide()` option set),
**inject** (emit a micro-action intent on the fast path), or **veto/substitute**
(deny or replace a proposed action). Rules are compiled ONCE from persona
(`PersonaCompiler`, off the hot path, at host registration / rare revision) into
that `[]Rule` artifact and cached on the Host. The engine plugs into the EXISTING
code at exactly two seams with NO VM change and NO new tick loop:

1. **The decision seam** — `decide()` / `contemplate_reality()` builtins
   (`runtime/actions_ambient.go`) call `pearl.Engine.TryDecide` BEFORE
   `Host.Strategist.Decide`. Hit ⇒ no LLM round-trip (the "Pearl fast path").
   Miss ⇒ bias the option set, then escalate to the LLM, then run the result
   through the veto.
2. **The action seam (Apperception)** — `pearl.Engine.Gate` wraps every
   `actionCallable.fn` at the single registration chokepoint
   (`runtime/dsl_bridge.go:253`). It is a synchronous pure-Go predicate pass that
   can veto/substitute any action — author-written OR brain-injected — in
   microseconds.

This is the same alias→fuzzy→brain escalation shape `cognition/resolve` already
ships, generalized from *recognition* to *decision + action gating*. It is pure
Go, allocation-light, shares one compiled table per host, and adds zero hot-path
network I/O.

---

## 2. Why this shape (rationale + trade-offs)

### Why one engine, two seams (not three subsystems)
gpt55 proposes distinct compiled artifacts (AppraisalPolicy, SocialHeuristics,
ChoiceWeights, SpeechPolicy…) and gemini proposes a separate `Guard` + separate
`Pearl`. At 500 hosts in one process, splitting these into separate interfaces
multiplies wiring and cache copies for no behavioral gain — they all read the same
per-host fact view and all execute the same predicate→effect evaluation. Collapse
them into ONE `pearl.Engine` whose rules carry an effect *kind*. The two *seams*
(decide-time, action-time) are real and distinct because the host code calls them
at distinct points; the *engine* behind both is one thing.

### Why a flat rule table, not a DSL / not a tree / not a SAT solver
- gemini's "symbolic-logic constraint solver / formal verification" is gold-plating.
  Rejected: at 500 hosts the gate must run in microseconds per action. A flat
  `[]Rule` scanned with a salience pre-sort is O(rules), rules capped at ~30/host.
- A second mini-DSL is rejected: we already have `dsl/interp`. The policy table is
  *compiled output* consumed by Go, not authored by humans. Predicates are typed Go
  closures selected from a fixed **predicate vocabulary** (enums), not arbitrary code
  — this keeps it deterministic, testable, and serializable for mesa to ship.
- gpt55's "event_policies as ENUMS → shared DSL handler templates" is the right
  instinct for the *authoring* side, but the runtime representation should be the
  compiled `[]Rule`, not re-parsed DSL per host per turn (token + consistency hazard
  the digest flags).

### Why compile from persona off the hot path
claude-persona + gpt55 agree: persona injected only as prompt text is decorative.
The `PersonaCompiler` runs at registration (and on rare `persona_revisions`),
emitting the `[]Rule` table. This is the claude-persona "external sampling + LLM as
materializer only" discipline applied to behavior: numbers/quirks → deterministic
rules, computed once. The compiled table is what populates / backs the
`cognition.Bundle.Persona` placeholder.

### Why the veto is pure-Go and synchronous
The host loop is single-goroutine per routine; `Strategist.Decide` already blocks
that goroutine. A veto that called the LLM would double the blocking latency and the
cost. The Apperception gate MUST be local Go predicates (gemini's correct core,
minus the SAT solver). It is also the only deterministic, unit-testable place to
enforce operator policy ("never drop items > N gp", "coward never attacks a
higher-combat player") independent of LLM behavior — which the believability
research grade requires.

### Trade-off accepted
The fast-path `TryDecide` can answer a decision "wrong" relative to what the LLM
would say. Mitigation: Pearl only answers decisions whose rules fire with high
confidence (`Rule.Confidence >= floor`); everything else escalates. This is the
explicit cost/quality lever — tune the floor up to spend more on the LLM, down to
spend less. Same knob `cognition/resolve` exposes via `scoreBrain`.

---

## 3. Concrete interfaces

New package `pearl/` (Go, pure, no CGo). Lives beside `brain/`, `cognition/`.

```go
package pearl // import "github.com/gen0cide/westworld/pearl"

// --- the fact view the engine reads (read-only, cheap, no allocation in hot path) ---

// Facts is the read-only snapshot a rule predicate evaluates against. It is
// assembled by the Host once per gate/decide call from already-live state:
// the world mirror, the per-host affect vector, the compiled persona facets,
// and the relationship row for the current counterparty (if any). All fields
// are value copies or stable pointers; predicates must not mutate.
type Facts struct {
    World    WorldView        // thin read-only adapter over *world.World (see §5)
    Affect   Affect           // per-host mood vector (lazily decayed on read)
    Persona  Facets           // compiled categorical persona facets (immutable)
    Counter  *Relationship    // trust/affinity row for the other party; nil if none
    Event    EventCtx         // what triggered this eval (decide question / action name / bus kind)
    Now      time.Time
}

type Affect struct {
    Stress     float64 // 0..1
    Confidence float64 // 0..1
    Valence    float64 // -1..1
}

// Facets is the runtime-readable projection of the persona (NOT the prose the
// LLM sees). Bands are pre-bucketed at compile time so predicates compare ints.
type Facets map[string]int8 // e.g. "bodily_risk":-1, "social_risk":+2, "diligence":+1

type Relationship struct {
    Name      string
    Trust     float64 // -1..1 (derived from mesa Beta(α,β); see relationship-matrix decision)
    Familiar  int     // encounter count
    Tags      []string // "scammer","trusted_partner","rival"
}

// --- the rule representation (the compiled artifact) ---

type EffectKind uint8
const (
    EffectBias       EffectKind = iota // reorder/weight a decide() option set
    EffectInject                       // emit a micro-action intent (reverie/quirk tic)
    EffectVeto                         // deny a proposed action
    EffectSubstitute                   // replace a proposed action's choice/args
)

// Predicate is a compiled, side-effect-free test over Facts. Selected from a
// fixed vocabulary at compile time (PredID), never authored as free code.
type Predicate func(f *Facts) bool

// Effect describes what fires when the predicate holds.
type Effect struct {
    Kind EffectKind

    // Bias: option label -> additive weight (applied to a decide() option set).
    Bias map[string]float64

    // Inject/Substitute: the action to emit or swap to, in DSL-callable form.
    Action string        // e.g. "examine", "wait", "say"
    Args   []interp.Value // resolved positional args (or nil)

    // Veto/Substitute: human/log reason; surfaced to delos + DSL .err.reason.
    Reason string
}

type Rule struct {
    ID         string
    When       Predicate
    Then       Effect
    Salience   int8    // higher = evaluated/applied first (veto > inject > bias tie-break)
    Confidence float64 // [0,1]; TryDecide only answers from rules >= engine floor
    Origin     string  // "cornerstone_directive" | "quirk:<id>" | "operator_policy"
}

// --- the engine (one per host; built from the compiled table) ---

type Engine struct { /* table []Rule, sorted by Salience desc; confidence floor; rng for jitter */ }

func New(table []Rule, floor float64, seed int64) *Engine

// TryDecide is the fast path for decide()/contemplate_reality(). On a confident
// local hit it returns (decision, true) and NO LLM call happens. On miss it
// returns (nil, false) AND a biased option ordering the caller passes to the LLM.
func (e *Engine) TryDecide(f *Facts, options []string) (*brain.Decision, []string, bool)

// Gate is the Apperception veto. Called synchronously before every action
// executes (author-written or brain-injected). Verdict.Allow=false vetoes;
// Substitute (optional) replaces the action.
func (e *Engine) Gate(f *Facts, action string, args []interp.Value) Verdict

// Injections returns any EffectInject intents whose predicates currently hold
// (consumed by the reverie/quirk emission point; see §6). Refractory-gated.
func (e *Engine) Injections(f *Facts) []Effect

type Verdict struct {
    Allow      bool
    Substitute *Effect // non-nil ⇒ run this instead
    Reason     string
    RuleID     string  // for delos audit
}
```

The compiler (off hot path):

```go
package pearl

// PersonaCompiler turns a persona document into the compiled rule table.
// Runs at host registration and on persona_revisions — NEVER per turn.
type PersonaCompiler interface {
    Compile(p persona.Document) (Table, error)
}
type Table struct {
    Rules   []Rule
    Version string // persona_revision_id; for cache invalidation + delos
}
```

These interfaces are **additive** — they touch no existing frozen signature.
`brain.Strategist` / `cognition.Client` / the DSL surface are unchanged.

---

## 4. Data model

### 4.1 Compiled rule table (runtime, in-process)
Lives on the Host as `*pearl.Engine` (a new field, §5). One per host, ~10-30 rules.
Built from the persona's compiled artifact. Not persisted as such — it is a pure
function of the persona document + compiler version, so it is rebuilt on login.

### 4.2 Persona-source → rule mapping (what the compiler reads)
The compiler input is the persona document (owned by the persona/mesa decision).
The fields THIS engine consumes:

| Persona source (from claude-persona / gpt55 schema) | Compiles to |
|---|---|
| `cornerstone.core_directives[] {priority, rule}` | `EffectVeto`/`EffectSubstitute` rules, `Origin=cornerstone_directive`, max Salience |
| `quirks[] {domain, trigger, modifier, strength, observable}` | `EffectBias` (action_bias) + `EffectInject` (reverie_bias) rules, `Origin=quirk:<id>` |
| `decision_style` (bodily/economic/social risk bands) | predicate thresholds + `EffectBias` weights on flee/trade/attack option sets |
| `social_policy` (greeting/help/retaliation enums) | `EffectInject` (greet) + `EffectBias` (help) rules keyed on bus event context |
| operator policy (process-wide, NOT persona) | highest-Salience `EffectVeto` rules merged into every host's table at compile |

### 4.3 Quirk object (the persisted source form, mesa `bots.persona` JSONB)
Adopt gpt55's structured quirk verbatim (it already maps onto our triggers):

```json
{ "id": "trade_double_check",
  "domain": "trade",                  // trade|inventory|navigation|combat|banking|social
  "trigger": "before_trade_confirm",  // maps to action name OR bus event kind (see §6)
  "visibility": "public_indirect",    // gate inject budget on visibility (server-visible only)
  "modifier": { "action_bias": {"confirm_trade": -0.4}, "inject": "examine_offer" },
  "strength": 0.7,                    // -> Rule.Confidence + Bias magnitude
  "stickiness": "permanent" }
```

Capped 2-4 quirks/host (gpt55's 70/25/5 mundane/idiosyncratic/oddball default).
The compiler reads `trigger` to choose the predicate, `modifier` to build the
`Effect`, `strength` → `Confidence`/weight, `visibility` to decide whether an
`EffectInject` is worth a server-visible action at all (reject invisible UI-only
gestures per both digests).

### 4.4 No new persistent store
The rule table is derived, not stored. Quirks live in the existing planned
`bots.persona` JSONB. Operator policy is a static config file the cradle loads
once. Affect + relationship are owned by their own decisions (Limbic /
relationship-matrix); this engine only READS them through `Facts`. **No new
Postgres/Redis/SQLite dependency** is introduced by the policy engine.

---

## 5. Exactly how it integrates with the real code

### 5.1 New Host field (the seam)
`runtime/host.go` — add ONE field next to `Strategist`/`Retriever`, defaulting nil:

```go
Pearl *pearl.Engine // host-local policy/quirk engine; nil ⇒ no gating, no fast path
```

Nil-default keeps every existing test green (no engine ⇒ behavior identical to
today). Production wiring in `cmd/cradle` sets it after construction, exactly like
the `Strategist`/`Retriever` swap. It is per-host (each host has its own compiled
table) but the COMPILER and operator-policy rules are shared.

### 5.2 Decision seam — inside the existing builtins (no VM change)
In `dslDecide` / `dslContemplateReality` (`runtime/actions_ambient.go`), insert
the fast path BEFORE the `Strategist.Decide` call:

```go
// dslDecide, after options are parsed, before h.Strategist.Decide:
if h.Pearl != nil {
    f := h.pearlFacts(ctx, question) // assemble Facts from world + affect + persona + counterparty
    if d, biased, hit := h.Pearl.TryDecide(f, options); hit {
        return interp.Ok(interp.String(d.Choice)), nil // NO LLM CALL
    } else {
        options = biased // LLM still decides, but from a persona-biased ordering
    }
}
decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question, Options: options})
// ... then run decision through the gate (substitute/veto) before returning:
if h.Pearl != nil {
    if v := h.Pearl.Gate(f, "decide:"+decision.Choice, nil); !v.Allow {
        if v.Substitute != nil { decision.Choice = v.Substitute.Action }
        // else: surface a typed failure or fall to a safe default
    }
}
```

This is the `cognition/resolve` escalation pattern (alias hit → return; miss →
brain) applied to decisions. It adds no new builtin name, honors `ctx`, and keeps
the frozen DSL surface intact.

### 5.3 Action seam (Apperception) — wrap fn at the single registration chokepoint
`runtime/dsl_bridge.go:247-258` is the ONE place every action callable is
constructed. Wrap `fn` there (non-invasive; no VM/interp change — the digest
explicitly calls this out as option (b)):

```go
for i := range spec.Actions {
    a := &spec.Actions[i]
    fn, hasHandler := actionHandlers[a.Name]
    if !hasHandler || a.NotYetImplemented { fn = makeStub(a.Name) }
    if h.Pearl != nil && a.Kind == spec.PrimaryAction { // gate only state-mutating actions
        fn = h.gateAction(a.Name, fn) // closure: build Facts, Pearl.Gate, veto/substitute/passthrough
    }
    base := &actionCallable{name: a.Name, host: h, ctx: ctx, fn: fn}
    it.Builtins[a.Name] = base
    if a.BangEligible() { it.Builtins[a.Name+"!"] = &interp.BangCallable{Underlying: base, Name: a.Name + "!"} }
}
```

`gateAction` returns an `actionHandler` that builds `Facts`, calls `Pearl.Gate`,
and on veto returns `interp.Fail(POLICY_VETO, verdict.Reason)` (a new typed
`ErrorCode`, additive) or on substitute calls the substitute action's handler.
This gates BOTH author-written actions AND any future `exec`/`improvise`
brain-authored fragments — they run on the same wrapped `Builtins`. The
`interp.Hooks` surface is NOT used: it is observe-only (no return value), so it
cannot veto (per groundtruth-dsl §4) — wrapping `fn` is the correct lever.

### 5.4 Facts assembly (`Host.pearlFacts`) — reads only live state
`WorldView` is a thin read-only adapter over the existing `*world.World`
accessors (`Self.Position`, `Players.FindByName`, `Inventory.Count`, etc., from
groundtruth-world-facts §2). Affect + Relationship come from the per-host structs
the Limbic/relationship decisions own (this engine is a pure consumer). Persona
`Facets` is the compiled categorical projection cached on the Host. Assembly is
allocation-light and bounded (cap nearby players/npcs via `Near`, per the
world-facts hot-path warning).

### 5.5 Reverie / quirk injection point (no tick)
`EffectInject` intents do NOT need a tick loop. They surface at the same yield
boundaries the DSL already has: the reverie emission decision (Phase 5) runs
`Pearl.Injections(f)` from inside an `on <event>` handler context or a post-action
hook, picks at most one refractory-gated server-visible intent, and emits it via
the normal action path (which is itself gated). This honors the event-driven
reality — injection is pulled at a yield point, never pushed by a timer. (If a
slow cadence is later wanted, it is a sibling goroutine to `heartbeatLoop`, per
host-loop-bus §6.1 — but v1 needs none.)

### 5.6 Populating the Persona placeholder
The compiled `Facets` + a rendered prose block become what backs
`cognition.Bundle.Persona`. The real `cognition.Client` (mesa) flattens the
compiled artifact into the `map[string]string` for the existing prompt surface
and `relation_with` convention — no change to `Bundle`'s shape required for v1
(additive `Bundle` fields are a later, separate decision).

---

## 6. Evaluation order (the precise contract)

For a given seam invocation, with the table pre-sorted by `Salience` desc:

**Decision seam (`decide`/`contemplate_reality`):**
1. Assemble `Facts`.
2. Scan rules; collect firing rules whose predicate holds.
3. If any firing `EffectVeto`/`EffectSubstitute` rule names the *only* / *all*
   options ⇒ resolve immediately (forced choice), confidence = rule confidence.
4. Else collect `EffectBias` weights into the option set; if the top option's
   net weight margin ≥ `decideMargin` AND its backing rule `Confidence ≥ floor`
   ⇒ **Pearl hit**: return that option, NO LLM.
5. Else **miss**: return the bias-reordered option list to the caller, which
   escalates to `Strategist.Decide`.
6. Run the resulting choice through the Gate (step in §5.2) for a final veto.

**Action seam (`Gate`):** evaluated in Salience order, first decisive verdict wins:
1. **Operator-policy vetoes** (highest Salience, `Origin=operator_policy`) — hard,
   non-overridable. e.g. "never drop item > N gp".
2. **Cornerstone-directive vetoes** (`Origin=cornerstone_directive`) — persona's
   inviolable rules. e.g. "coward never attacks higher-combat player".
3. **Quirk substitutes** (`Origin=quirk:*`, `EffectSubstitute`) — e.g. replace
   `confirm_trade` with `examine_offer` once (refractory-gated).
4. Default: **Allow**.

**Injection (`Injections`):** only `EffectInject` rules; filtered to
`visibility != private`, then refractory-gated, then at most one returned
(mode-first per gpt55: pick mode bucket before action). Empty by default.

Determinism: ties broken by `Rule.ID` lexical order. Per-host jitter (reverie
LogNormal multiplier, claude-persona) is the ONLY randomness, seeded once per host
from `opts.Username` — reproducible for tests + delos replay.

---

## 7. Risks / open questions

- **Predicate vocabulary scope creep.** The fixed `PredID` enum must cover the
  triggers we actually want (trade-confirm, low-HP, attacked-by-stronger, stranger
  encounter, bank-arrival). Start with ~12 predicates mapped to existing world
  accessors + bus kinds; resist a general expression language. Risk: under-coverage
  forces quirks to be inert. Mitigation: the compiler emits a no-op + a delos warning
  for any quirk whose trigger has no predicate.
- **Facts assembly cost on the hot path.** `pearlFacts` runs on every gated action
  and every `decide()`. At 500 hosts this is hot. Must be allocation-light and cap
  nearby-entity scans (`world.Near` radius), reuse a per-host scratch `Facts`.
  Open: benchmark a worst-case combat routine (many gated `attack` calls).
- **Veto deadlock with author routines.** If the gate vetoes an action the routine
  has no fallback for, the routine may loop or abort. The `bang` form turns a veto
  into an `abortSignal` (clean), but the non-bang form returns a typed failure the
  author may not handle. Open: do we want a "veto is always recoverable" contract,
  or should cornerstone vetoes hard-abort? Lean: operator-policy veto ⇒ hard abort
  (safety), persona/quirk veto ⇒ recoverable failure.
- **Compiler lives where?** v1: in-process `pearl.PersonaCompiler` run at login from
  the persona doc. Eventually mesa compiles + ships the `Table` (the gemini "Pearl
  compiled by mesa" idea). The interface is identical; only the location moves.
  Don't build the gRPC ship path until mesa splits to its own binary (matches the
  claude-arch over-build critique).
- **Affect/relationship coupling.** This engine READS affect + relationship but does
  not own them. If those decisions (Limbic, relationship-matrix) land later, the
  engine must degrade gracefully: nil `Counter`, zero `Affect` ⇒ rules that depend
  on them simply don't fire. Confirmed safe by design (predicates test concrete
  values; absent ⇒ false).
- **Interaction with `exec`/`improvise` (brain-authored DSL).** When those reserved
  builtins land, brain-authored fragments run on the SAME gated `Builtins` — good
  (free apperception over LLM-authored actions). Open: should the gate be STRICTER
  for brain-authored fragments (lower trust) than author-written routines? Possible
  via an `Facts.Event.Source` flag the gate reads.
- **Operator-policy distribution.** Process-wide vetoes are merged into every host's
  table at compile. If operator policy changes at runtime, every host's table is
  stale until recompile. v1: operator policy is static config (cradle restart to
  change). Acceptable at current phase.
```