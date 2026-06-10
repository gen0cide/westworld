> **ARCHIVED (initial brainstorming), 2026-06-10.** The C-backlog here was built (persona/
> package); canonical docs are `docs/persona-authoring.md` / `docs/persona-compile.md`.
> PRIMARY SOURCE preserved: A1's verbatim answer — **cohort = launch batch** (a deployment
> grouping, not a personality class; archetype = personality). Still-open B1-B5 authoring
> sessions + ambiguous A-items were harvested into `docs/TODO.md`. Kept verbatim below.

# Persona — open decisions worksheet (TEMPORARY)

> **How to use this.** One block per decision. Each has: **Problem** (digested),
> **Options + tradeoffs**, **My recommendation**, the **Decision needed** (the
> crisp question), and an **Answer:** slot for you to fill. Pure build/execution
> items (no judgment call) are listed compactly at the end. Gitignored working
> doc — delete when drained into the canonical docs.
>
> Priority: 🔴 blocks code scaffolding · 🟡 needed before Phase-3 runs · 🟢 defer-OK.

---

## A. Architectural / design decisions

### A1. 🔴 Archetype vs cohort — where do trait priors live?
**Problem.** The HEXACO/value/econ priors have to live *somewhere* the sampler draws from. The reference YAML's Appendix A currently puts `hexaco_mu_prior` on the **cohort**. We argued the **archetype** should carry traits and the **cohort** should carry only archetype-mix + flavor (home, name pool, activity bias, voice register) + optional experiment μ-shifts.
**Options + tradeoffs.**

- **(a) Two levels — archetype carries traits, cohort carries mix+flavor.** *Pro:* avoids flattening (within one cohort you still get a full personality spread); matches `decision-persona-schema.md §3` (`Archetype{MuVector,Sigma}` + cohort `{PiMultipliers,MuShifts}`); cohorts stay cheap to author. *Con:* one more layer of config indirection.
- **(b) One level — cohort carries everything.** *Pro:* simplest config. *Con:* cohort = personality ⇒ ~10 personality clusters wearing 500 nametags = the exact flattening the system exists to avoid.
- **(c) One level — drop "cohort," only archetypes.** *Pro:* clean. *Con:* loses the spatial/social grouping ("who hangs out where") that drives community formation.
**My recommendation.** (a). It's the only one that doesn't flatten and it matches the decision record.
**Decision needed.** Confirm two-level (archetype = traits, cohort = mix+flavor)? And: may cohorts apply **μ-shifts** for experiments (e.g. `shy_majority`), or stay pure mix+flavor?
**Answer:** I think there is a misconception between cohort and archetype. Cohort represents host launches and don't really have anything to do to connect personas. Simply, we released another N hosts into the population, and maybe we're doing some expirmentation on it from the persona perspective, but it's simply a way to track that cohort. Archetype is really the underlying personality field here. Cohorts are likely to be pulled from a library of "inspiration" that we'll have built up, or might be copy + tweak of previous host configurations.

---

### A2. 🟡 Cohort taxonomy — count, persistence, assignment
**Problem.** We need to fix how many cohorts exist, whether a host can change cohort over time, and how cohorts are assigned at registration.
**Options + tradeoffs.**

- **Count:** ~6 seed set (easy to balance, less diversity) · ~10–15 (good spread, more authoring) · ~50 (max diversity, hard to balance + analyze).
- **Persistence:** *persistent* (cohort = immutable origin/home; personality change handled by Trajectory drift, not cohort change — cleaner analytics) vs *transitionable* (a host migrates cohorts as it drifts — more emergent, but muddies cohort science + adds machinery).
- **Assignment:** weighted random draw at genpop (simple, hits the population mix) vs designed/seeded clusters.
**My recommendation.** Start **~6–8** seed cohorts, **persistent** (cohort is origin, not current personality), **weighted-random** assignment at genpop. Defer cohort transitions.
**Decision needed.** Count? Persistent or transitionable? Assignment method?
**Answer:** We probably need to rethink this now because of the answer to A1.

---

### A3. 🟡 Band cut-points — absolute thresholds or population percentiles?
**Problem.** `band` ("high") is derived from `mu`. Is the cut-point **absolute** (`mu>0.7 → high`) or a **population percentile** (top quintile → high)? `decision-persona-schema.md` says percentile; everything else reads as absolute. They diverge hard under cohort experiments.
**Options + tradeoffs.**
- **Absolute.** *Pro:* "high honesty" means the same disposition in every population; cohort experiments *correctly* skew the band distribution (a scammer cohort genuinely has few high-H hosts); simpler. *Con:* a homogeneous cohort may have almost no contrast in some traits.
- **Percentile.** *Pro:* always a spread of bands for legibility/contrast. *Con:* "high" means different things in different populations → analytics confound, and the prose for "high X" could describe an absolutely-average host (a scammer pop's "most honest scammer" would render as "high H" — wrong).
**My recommendation.** **Absolute.** Bands feed prose + reflexes; we want stable behavioral meaning, and experiments *should* shift the distribution.
**Decision needed.** Absolute or percentile? If absolute, confirm/edit the ladder, e.g.: `very_low<0.15 · low<0.35 · mid<0.50 · mid_high<0.65 · high<0.85 · very_high≥0.85`.
**Answer:**

---

### A4. 🔴 The compiler band→policy + EventPolicy → DSL handler templates
**Problem.** Persona must be **causal on the no-LLM path**. The planned mechanism is compiled `EventPolicy` enums resolving to versioned DSL "persona-tier default handlers" (the `extends host`/`super()` mechanism, task #93) — which **do not exist in the DSL yet**. `decision-persona-schema.md §5`: this is the highest-leverage unbuilt piece; if it's too coarse, "hosts read identically despite distinct Cornerstones." This is the main flattening risk.
**Sub-decisions + tradeoffs.**
- **(a) Build EventPolicy→templates now, or quirks-first?** Quirks already compile to a behavior bias (`EffectBias`) and, with the trust prior + flee/eat thresholds, give real persona causality *without* #93. The full handler stack is net-new DSL work + a parity-test surface.
- **(b) Enum granularity.** e.g. `on_trade_request: screen_then_engage | decline_strangers | eager_accept` — 3 buckets, or finer? Finer = more differentiation, more templates to author/test.
- **(c) Must-declare-every-event vs sparse-with-fallthrough.** Must-declare forces the designer to consider every reactive surface (heavy); sparse is lighter but leaves gaps to the base reflex.
**My recommendation.** **Quirks-first.** Ship persona causality via quirks + trust prior + survival thresholds (no #93 needed), and make EventPolicy→handler-templates its own dedicated effort. Start enums **coarse (3-ish)**; **sparse** declaration (base reflex fills gaps).
**Decision needed.** Quirks-first or build EventPolicy templates now? Enum granularity? Declare-all or sparse?
**Answer:**

---

### A5. 🔴 Quirk Bucket-C registries — close the silent-no-op hole
**Problem.** `quirk.trigger / object / binding / suppress_when` are semi-open strings today. An unmatched value doesn't error — the quirk **silently never fires**. They need closed, validated vocabularies checked at genpop.
**Options + tradeoffs.** Tight closed registry (safe; rejects non-conforming quirks at birth + logs) vs loose free strings (expressive but silent failures). Not really a contest — the only real question is *what's in the sets*.
**My recommendation.** Closed enums + validate-at-genpop (reject + log). Proposed sets (draftable in full):
- `trigger` ∈ {all `event.Kind()`} ∪ {`pre_action:<verb>` for verbs in the action registry} ∪ {`on_encounter`}
- `object` prefix ∈ {`bank_node` · `area` · `player_type` · `equipment` · `item` · `npc` · `skill` · `phrase`}
- `binding` ∈ a curated gameplay-binding set (e.g. `fishing_spot_preference`, `trade_screening`, `combat_response`, `bank_node_preference`, `chat_emit`, …)
- `suppress_when` ∈ {`in_combat` · `low_hp` · `banking` · `trading` · `none`}
**Decision needed.** Approve closed-enum + validate-at-genpop? Want me to draft the full registries for review?
**Answer:**

---

### A6. 🟡 `learned_emergent` quirk promotion
**Problem.** A quirk can graduate into the single mutable slot through experience. Unspecified: who detects it, the trip threshold, the cap.
**Options + tradeoffs.** Detector home: a **mesa reflection-cadence job** (rare, off hot path, can use a model) vs the **Limbic** (cheap, in-band, but it's a fast subscriber). Threshold tighter (fewer, higher-confidence promotions) vs looser (more drift, more identity-creep risk).
**My recommendation.** Mesa reflection job; threshold start ~**12/14 recent similar trips**; **one slot**, cap `‖Δr_a‖<0.5`, log every promotion + snapshot to `persona_revisions`.
**Decision needed.** Detector home? Threshold? Confirm one-slot + cap.
**Answer:**

---

### A7. 🟡 Persona-revision trigger / cadence
**Problem.** Slow Trajectory re-ranks snapshot into `persona_revisions` (Sonnet, rare). Settled: fast mood churn never triggers. Unsettled: the positive triggers + rate limit.
**My recommendation.** Triggers = {value re-rank crosses a threshold · a new *stable* sub-goal · emergent-quirk promotion}. Rate-limit ≤ once / few days. Store snapshot + rationale.
**Decision needed.** Confirm the trigger set + the rate limit.
**Answer:**

---

### A8. 🟢 `skill_logits` seeding at birth
**Problem.** Initialize the soft-RL skill focus from `curiosity_anchor` (+ C temperature) or start flat?
**Tradeoffs.** Seeded → faster week-one divergence, hosts feel distinct immediately. Flat → divergence is purely earned (purer experiment, slower to read).
**My recommendation.** Light seed from curiosity.
**Decision needed.** Seeded-from-curiosity or flat?
**Answer:**

---

### A9. 🟢 `sub_goals` seeding at birth
**Problem.** Boot with an empty list (first cognitive loop decomposes the north star) or a seeded starter sub-goal?
**Tradeoffs.** Seeded → the Tutorial-Island first think-turn has something concrete to act on. Empty → cleaner, but the very first decision is "figure out what to do."
**My recommendation.** Seed one starter decomposed from the north star.
**Decision needed.** Empty or seeded starter?
**Answer:**

---

### A10. 🟡 `typo_rate` / voice realization mechanism
**Problem.** `typo_rate` + voice are believability dials, but *how* a host's chat actually acquires its voice + typos is unspecified.
**Options + tradeoffs.**
- **(a) Deterministic post-processor** mangles outgoing chat at `typo_rate`. *Pro:* reproducible, cheap, exact rate. *Con:* typos can look mechanical.
- **(b) Instruct the chat-gen LLM** to write in-voice + occasionally typo. *Pro:* natural. *Con:* entropy; LLMs often ignore "make a typo."
- **(c) Hybrid:** LLM writes in-voice content; deterministic post-processor injects plausible typos at `typo_rate`.
**My recommendation.** (c) hybrid.
**Decision needed.** Which approach?
**Answer:**

---

### A11. 🟡 Mood → prompt projection vocabulary
**Problem.** We carved out the per-call mood line as a **deterministic template** (not an LLM call). The actual lookup table — `(valence, arousal, stress, confidence) → short phrase` — is unwritten.
**My recommendation.** Author a small bucketed table (I can draft it). Keep it deterministic.
**Decision needed.** Approve a deterministic table (vs LLM)? Want me to draft it?
**Answer:**

---

### A12. 🟡 Curiosity→behavior & attention→interrupt mappings
**Problem.** `curiosity_cur` (what pulls you) and `attention_cur` (resist being pulled) are named as the differential-learning engine + interruptibility, but the concrete formulas (how curiosity biases goal selection / wander; how attention sets the interrupt-ladder threshold) aren't written.
**My recommendation.** Persona **exposes** the scalars/vector; the exact formulas live in the cognitiveLoop / interrupt-ladder design (cognition.md). Defer the math there, don't duplicate it in persona.
**Decision needed.** OK to defer the formulas to the cognitiveLoop design (persona just exposes fields)? Or specify now?
**Answer:**

---

### A13. 🟡 Cache-prefix re-injection dependency
**Problem.** Believability needs the Cornerstone re-injected every call; cost needs it to be a cached prefix. This relies on the (unbuilt) Anthropic strategist honoring prompt caching.
**My recommendation.** Design the Cornerstone block as a stable, ordered cache-prefix now; treat the caching itself as a brain-integration constraint, not a persona blocker.
**Decision needed.** Acknowledge as a brain-build constraint (nothing to decide now), or do you want a fallback plan if caching underperforms?
**Answer:**

---

### A14. 🟢 `prefs` behavioral-econ dormancy
**Problem.** Only λ has live consumers; the rest of `prefs` is "speculative, unconsumed until an LLM strategist computes a utility function."
**My recommendation.** Keep stored `omitempty` + dormant; build λ consumers only.
**Decision needed.** Park the rest, or build consumers now?
**Answer:**

---

### A15. 🟢 Trust-ledger projection lossiness
**Problem.** `relation_with` reads a flat `-1..1` label — lossy vs the real `Beta(α,β)` ledger.
**My recommendation.** Keep the flat map for backward-compat; route the richer ledger to the brain via `Bundle.Card` once the brain reads structured persona.
**Decision needed.** Confirm flat-now / rich-later?
**Answer:**

---

## B. Authoring decisions (the deferred "paired session" content)

### B1. 🔴 The archetype prototype table
**Problem.** The actual prototypes (μ-vectors + Σ + value/econ priors per archetype) don't exist. The persona paper offers 13.
**Decision needed.** Use **13** (the paper's table) or our own count? Author by hand, LLM-draft-then-curate, or pair on it?
**Answer:**

### B2. 🔴 `cohorts.yaml` — weights, forbidden pairs, name pools
**Problem.** The population skew, the forbidden-pair set, and per-cohort name pools are unspecified.
**My recommendation.** Ratify the Pareto skew (~60% casual / 20% focused / 10% newbie / 10% extreme); I draft the forbidden pairs + name pools for review.
**Decision needed.** Ratify the skew? OK for me to draft the rest?
**Answer:**

### B3. 🟡 North-star catalog + conflict resolution + evolvability
**Problem.** Catalog count is open; two unresolved rules: can a north-star **evolve** (e.g. scammed-early → "most distrustful trader")? How are **conflicting drives** resolved (wealth north-star + shy temperament)?
**My recommendation.** North-stars evolvable only via `persona_revisions` (rare). Conflict = the north-star **biases** but the temperament **gates** (a shy wealth-seeker still trades less than an outgoing one) — i.e. blend, not override.
**Decision needed.** Evolvable? Conflict = blend or priority?
**Answer:**

### B4. 🟡 Reverie catalog
**Problem.** Which small behaviors exist + their trait coefficients + suppress rules — the cross-cutting paired session (`reveries.md`).
**Decision needed.** Scope it now, or after the persona code lands (it only needs the trait-weight *form* + jitter to exist for v1)?
**Answer:**

### B5. 🟡 500-host population mix
**Problem.** The exact distribution across cohorts/archetypes for the target population.
**Decision needed.** Ratify the Pareto skew (ties to B2), or want a different shape?
**Answer:**

---

## C. Build / execution backlog (no decision — just scheduling)
Behind the existing nil-stub seam (`Host.Strategist`, `Host.Retriever`):
- [ ] 🔴 `persona/` package: types (`Persona`/`Cornerstone`/`Trajectory`/`PersonaCard`), `Compiler`, deterministic `Render()`, `Facet()`.
- [ ] 🔴 Offline **sampler** (`cmd/delores genpop`): cohort/archetype priors → forbidden-pair rejection → derive quirks + reverie jitter → seal + hash.
- [ ] 🟡 **Best-of-N compile pipeline** (offline) — per `persona-compile.md`.
- [ ] 🟡 **mesa**: full `hosts.persona` JSONB + `cornerstone_hash`/`prng_seed`/`schema_version`; `episodes.stage=0` + widen `chk_stage` 0..3; `persona_revisions` write path; JSONB-tolerant migration.
- [ ] 🟡 **cognition.Bundle** additive `Card`/`Pinned`/`Affect` + the re-inject cache-prefix.
- [ ] 🟡 **Limbic** bus subscriber (sibling goroutine): mood/skill-logit/drift writes, lazy decay-on-read, ~30s flush.
- [ ] 🟡 **Drift update rules** (κ/ρ regression-to-anchor; Fleeson `hexaco_state`; softmax temp `T=1−0.5·C`).
- [ ] 🟡 Fill reserved `mood()` / `motivation()` DSL builtins (`actions.go` `NotYetImplemented`).
- [ ] 🟢 `reveries/` package: `Catalog()` + `ReverieKernel` closures + the expression hook (Phase 5).

## D. Parked (do NOT build yet)
- Gossip / reputation propagation (`SigGossip*`) — needs the live trust graph + chat content classification first.
- Full prospect-theory evaluator, full 10-value Schwartz vector, Gaussian copula + SBERT dedupe — replaced by trimmed forms; revisit only if delos needs them.

## E. Doc-consistency loose ends (cheap, no decision)
- [ ] Fix Appendix A (archetype/cohort split) in `persona-schema.reference.yaml` — pending **A1**.
- [ ] Audience-tag pass (`LLM`/`CODE`) across the reference YAML + pin the Bucket-C registries — pending **A5**.
- [ ] Write the preseed tiers (Tier 0/1/2) into the reference YAML.
- [ ] Apply remaining §GAP checklist items from `quantitative-persona-models.md §10` to the canonical docs.

---

## Fastest unblock
If you answer only three for code to start: **A1** (archetype/cohort split), **A5** (quirk registries), **A4** (compiler causality path — even just "quirks-first").
