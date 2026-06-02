# A persona schema worth running 500 agents on

**The team's current five-trait schema is a folk taxonomy that conflates psychometrically distinct constructs and omits Honesty-Humility — the single best-validated predictor of the very ethics behaviors the project exists to study.** Switching from floats to categorical descriptors helped marginally but treats a symptom; the underlying flattening comes from RLHF typicality bias, Assistant-persona bleed-through, and weak behavioral grounding, none of which depend on data type. The right architecture is a three-layer ontology — HEXACO disposition × Schwartz values × behavioral-economic preferences — sampled jointly from an empirical mixture model with realistic covariance, then materialized as rich behavioral text rather than adjective labels. Quirks should be executable behavior modifiers with dimensional structure, not flavor-string lists. Reveries should be trait-derived weights with per-agent jitter, not an independent named-weights bag. Identity (HEXACO mean, top values, north-star, backstory, voice register, core quirks) is immutable on the month horizon; everything else — trust ledgers, sub-goals, skill focus, mood, reverie expression, friendships — drifts through experience, producing the differential-learning trajectory the research design requires. The rest of this report makes that case in detail and ends with a concrete final field structure.

## What the project actually assumes today

The Westworld repo (gen0cide/westworld) treats persona, reveries, memory, and the brain as **load-bearing but mostly unimplemented**. The `persona/`, `memory/`, `reveries/` Go packages are empty as of late May 2026; the brain is a `StubStrategist`. Working today: protocol/session/world/action/event/DSL/pathfinding/render. The design documents specify a persona template with **cohort, age, vocabulary style, vocal tics, social temperament, risk tolerance, ambition, daily activity preferences, north star, reverie weights, and default event handlers** — broader than the five-trait critique target in the task brief but no more rigorous. The README states research-goal priorities explicitly: emergent economy, believable population, LLM-eval, with "Turing-among-bots" reframed as substrate rather than goal. The "5 Properties of Autonomy" frame (independence, adaptive contextual understanding, complex decision-making, goal coherence, transparent reasoning) is anchored against population-scale measurements: skill-checkpoint reach rates, relational-graph clustering vs. null model, categorical ethics incidents correlated with persona traits, and detection-suspicion utterance rates.

Three design commitments matter most for the schema: **(a) all hosts start equal** (no bootstrap routines, uniform op budget) so divergence must come entirely from persona; **(b) reveries are not polish but the package that distinguishes "a population of drones" from "a population of plausible inhabitants"**; and **(c) cross-bot reputation is not shared at the mesa layer** — agents must communicate beliefs via in-game chat, so trust must live on each agent as a private model of every other agent.

## Why the current trait keys cannot carry the research

The schema's five categorical keys — *social_disposition, risk_tolerance, ambition, curiosity, trustfulness* — fail on construct validity in ways that directly undermine each research question. **`Ambition` alone conflates four independent constructs**: HEXACO Conscientiousness (diligence facet), Schwartz Achievement value, Schwartz Power value, and behavioral-economic time preference δ. A patient diligent grinder (high C, high δ, achievement-oriented) collapses into the same bin as a Machiavellian power-seeker (low Honesty-Humility, Power-oriented, often impulsive). The simulation literally cannot distinguish the two cases that the ethics research most needs to separate.

`Trustfulness` makes the most damaging single conflation: it equates trust (HEXACO Agreeableness trust facet; Berg-Dickhaut-McCabe trust-game send amount) with trustworthiness (Honesty-Humility; trust-game return amount). These are **psychometrically distinct** — Thielmann & Hilbig (2018) and Schild, Stern & Zettler (2019) show H predicts trustworthiness, not trust — and the asymmetry between them is precisely what economic exploitation requires. Conflating them erases the construct.

`Risk_tolerance` mashes prospect-theory parameters (α, β diminishing sensitivity; λ loss aversion ≈ 2.25 in Tversky-Kahneman 1992, ≈ 1.31 in Brown et al. 2024 meta-analysis; γ probability weighting) with HEXACO Emotionality (Fearfulness facet). Weber, Blais & Betz (2002, DOSPERT) show financial and physical risk-taking dissociate; conflating them prevents the simulation from distinguishing an agent who fears the Wilderness from one who simply refuses negative-EV gambles. `Social_disposition` collapses Extraversion (energy, boldness) with Agreeableness (forgiveness, trust); `curiosity` is the only key that maps to a single construct (HEXACO Openness, Inquisitiveness facet).

**Missing entirely**: Honesty-Humility as an explicit factor, despite being the project's central research target; Emotionality (which drives grudges, attachment, loss reactions); explicit time preference (δ, β_qh) for month-scale grinds; loss aversion λ; a motivational/values layer linking traits to the north-star; and Fischbacher-Gächter-Fehr (2001) cooperation type (conditional cooperators ≈ 50%, free riders ≈ 30%) needed to model clan dynamics.

## The three-layer ontology

Replace the five keys with three independently validated layers, each with its own empirical literature. **HEXACO captures *how* an agent acts**; **Schwartz values capture *why*** (and anchor the north-star); **behavioral-economic preferences capture decision parameters** that plug directly into utility functions and policy. Anglim, Knowles, Dunlop & Marty (2017) show HEXACO facets explain at most ~28% of Schwartz-value variance — confirming the layers are non-redundant.

The case for HEXACO over Big Five is decisive for this project specifically. The **Honesty-Humility factor** is the strongest meta-analytic personality predictor of prosocial behavior in situations that afford exploitation (Thielmann, Spadaro & Balliet 2020, *Psychological Bulletin*, k = 770 effect sizes), of unobserved cheating (Hilbig & Zettler 2009; Kleinlogel et al. 2018, die-under-cup paradigm), of dictator-game giving and trust-game return (Thielmann & Hilbig 2018; Schild et al. 2019), and shares a latent correlation of **−0.95 with the Dark Triad common factor** (Hodson et al. 2018). A model without H cannot tell the difference between an agent that doesn't steal because it's observed and one that doesn't steal because it's honest — which is the exact manipulation the unobserved-ethics paradigm depends on.

Each factor maps to concrete RSC behaviors at the facet level: **H–Sincerity/Fairness** governs trade-window honesty; **H–Greed-Avoidance** governs whether a corpse left by a friend gets looted; **E–Fearfulness** controls Wilderness entry and flee thresholds; **E–Sentimentality** drives mentor attachment and clan stickiness; **A–Forgiveness/Patience** decides whether scams produce grudges or re-trades; **C–Diligence** is the strongest predictor of multi-week grind persistence; **C–Prudence** governs banking frequency; **O–Inquisitiveness** drives map exploration and quest-dialogue depth.

The values layer uses Schwartz's (1992, 2012) circumplex — power, achievement, hedonism, stimulation, self-direction, universalism, benevolence, conformity, tradition, security — collapsed to the four cross-culturally robust higher-order dimensions (self-enhancement vs. self-transcendence; openness-to-change vs. conservation), plus a categorical pick of the agent's *top* basic value as the explicit `north_star_value`. This is non-trivially distinct from HEXACO: an agent's Conscientiousness tells you they'll grind, but Schwartz Achievement vs. Schwartz Benevolence tells you whether they'll grind for personal glory or to give wealth away.

The decision-parameter layer pins down what HEXACO and values leave under-specified. **Risk**: α and β diminishing sensitivity (default ≈ 0.88), λ loss aversion (lognormal prior, median ≈ 2.0), γ probability weighting (≈ 0.65); these plug directly into prospect-theory value functions in the brain's gamble-evaluation routines. **Time**: daily δ (0.99–0.9999) and quasi-hyperbolic β (Laibson 1997; ≈ 0.85 default with σ); month-scale grinds are unreachable without explicit patience parameters and Yeh et al. (2021) confirm time preference loads on its own factor independent of Big Five. **Social**: Fehr-Schmidt α_envy (≈ 0.85) and β_guilt (≈ 0.30), Berg-Dickhaut-McCabe trust-send fraction, separate positive and negative reciprocity weights, and a categorical Fischbacher-Gächter-Fehr cooperation type.

## On flattening: data type is a symptom, not the cause

The team's switch from floats to categorical descriptors was a reasonable tactical patch but rests on a wrong causal model. **LLMs do not internally average numeric trait values.** Levy & Geva (2024) and Nikankin et al. (2025) show models encode numbers digit-wise in base 10 with circular per-digit features, not as continuous magnitudes; "0.42" is processed as opaque tokens carrying only weak ordinal information. What floats actually do wrong is invite the model to pattern-match on aggregated-psychometric-report context, where the typical text is dispassionate and middle-of-the-road — pulling outputs toward the population mean by stylistic association.

The dominant flatteners operate on any representation. **RLHF systematically collapses output diversity** (Kirk et al. 2024, ICLR; Padmakumar & He 2024 on writing homogenization; Mohammadi 2024 "Creativity Has Left the Chat"); **typicality bias is a data-level cause** that survives algorithm changes (Zhang et al. 2025 estimate α ≈ 0.57 — annotators prefer the more typical of two equal-quality responses 57% of the time); the **default Assistant persona persists** even under role-play and co-activates with the assigned character (Anthropic Persona Vectors, Chen, Arditi, Sleight, Evans & Lindsey 2025; SYCON-Bench 2025); and the **express-vs-behave gap** means an LLM can recite Big-Five scores accurately while its actions show only weak correspondence (Serapio-García et al. 2025 *Nat. Mach. Intell.*; Jiang et al. 2023 PersonaLLM; "Linear Personality Probing and Steering" 2025 finds aggregate trait directions are degenerate — "average extraversion" has no clean behavioral correlate).

Critically, **the silicon-sampling literature documents flattening with categorical conditioning, not floats**. Bisbee et al. (2024) found ChatGPT matches ANES means but produces too-low variance and distorted correlations using demographic *labels*. Santurkar et al. (2023) showed misalignment persists "even after explicitly steering the LMs towards particular demographic groups." Park, Schoenegger & Zhu (2024) found a "correct answer effect" with near-zero variance across runs. Dominguez-Olmedo, Hardt & Mendler-Dünner (2024, NeurIPS) showed Census-ACS responses dominated by ordering biases that, when debiased, collapse toward uniform — never recovering human population entropy. **All of these used categorical inputs.** The data type wasn't the problem then either.

Categorical descriptors have their own collapse mode: every "moderately cautious" agent inherits the same trained-in semantic cluster, producing within-bin near-clones; Lutz et al. (2025) found explicit descriptor priming actively *amplifies* stereotyping relative to name- or interview-based priming; and Cheng et al. (2023, "Marked Personas" and "CoMPosT") document caricature distortion in LLM persona simulation that descriptors make worse, not better.

The fix is structural, not representational. **Sample trait values externally from empirical population distributions** — HEXACO-PI-R public data, IPIP-NEO (~1M respondents), Quantic Foundry Gamer Motivation Profile, the WVS — preserving real covariance via a Gaussian copula. **Materialize each numeric profile deterministically into a behavioral paragraph before it touches the agent LLM**: an agent with H at the 12th percentile becomes a paragraph of concrete acts ("when she finds a stack of arrows on the path, she pockets them without looking up"), never a number or an adjective. This is the move that empirically restored fidelity in Park, Zou, Shaw, Hill et al. (2024, "Generative Agent Simulations of 1,000 People"): interview-grounded conditioning reached 85% of within-person test-retest reliability while demographic-only conditioning showed substantial demographic-group bias and worse fidelity.

The recommended representation is therefore **hybrid**: numeric percentile (continuous, for sampling and analytics) → categorical band (low/mid-low/mid/mid-high/high, for legibility) → behavioral exemplars paragraph (for the actual prompt the LLM sees). The numbers exist so we can do science; the bands exist so engineers can read agents; the behavioral paragraph exists because that's what the model actually responds to.

## Quirks as executable modifiers, not flavor strings

A freeform string array of quirks invites the cliché attractor that LLMs collapse to: loves coffee, lucky number 7, hates Mondays, talks to plants. Worse, freeform quirks don't affect behavior — they're flavor text the brain may or may not consult. The replacement is **dimensional decomposition into executable modifiers**.

Each quirk should specify a domain (chat, movement, economic, combat, social, ritual, aesthetic, informational), an object referent (specific NPC class, item, tile, color, level band), a relation (prefers, refuses, avoids, ritualizes, fears, distrusts, greets), an intensity (probabilistic nudge to near-deterministic), and a trigger condition (encounter, pre-action, post-action, location, state). This grammar produces concrete RSC quirks: *"Always banks at Falador first"* is a `movement / bank_node:Falador_west / prefers / strong / pre-action:bank-intent` modifier with a 40% extra-travel willingness; *"Mistrusts anyone in full bronze"* is `social / equipment:bronze_full / distrusts / moderate / on-encounter` applying a −0.3 trust prior; *"Sells in stacks of 28"* is `economic / shop-sell / quantizes / strong / on-trade` rounding quantities.

Per agent, sample **3–7 quirks** (truncated negative-binomial, mean 4). Below two leaves agents bland; above eight drowns the structured traits in noise. The mix per agent should split roughly **50% trait-derived** (deterministic templates filled by the LLM, e.g. low-H + high-C → "ritualized pre-scam routine: [specific action]"; high-O + Bartle-Explorer → "detour to examine [obscure scenery]") and **40% idiosyncratic** (LLM-generated under anti-cliché constraints), reserving **10% as a `learned_emergent` slot** that the runtime promotes when behavior tracking detects a stable pattern (banked at Falador first on 12/14 trips → promote to explicit quirk).

The generation pipeline must defeat typicality bias actively. Verbalized sampling (Zhang et al. 2025) explicitly asks the model to emit a *distribution* — "generate 30 candidates: at least 5 common, 15 unusual, 10 rare, no two sharing a (domain, object) pair" — and yields 1.6–2.1× diversity gains on closed models. Generate **at population scale, not per agent**: 50 agents per LLM call breaks the "every agent loves coffee" attractor. Include hard-banned templates as negative examples, force RSC mechanical grounding ("must reference a specific named city, NPC class, item, or UI element"), and seed cross-domain analogies (bartending rituals, surgeon scrub protocols, drummer pre-show routines) to break the model's RuneScape-quirk default. After generation, run an embedding-based dedupe (SBERT, farthest-point sampling at cosine ≥ 0.85) and a population-entropy check across the (domain × relation) joint distribution; resample any bin holding > 15% of all quirks.

Most quirks are immutable identity, but allow the small `learned_emergent` channel to accumulate during play. This matches the Crusader Kings 3 design lesson: stable identity prevents the "every ruler became a stat-god saint within a decade" collapse, while a narrow mutability channel lets behavior shape persona over time.

## Reveries: trait-derived weights with personal jitter

The current reveries proposal lists named weights — `camera_pan`, `examine_scenery`, `pathing_wobble`, `brief_hesitation`, `accidental_misclick`. This will mode-collapse harder than traits do, because reveries are a composition of two many-to-one maps: trait sampler → reverie sampler. If both flatten, the population's idle behavior becomes uniform noise — useless for the believability layer the docs call "load-bearing."

The fix is to make reverie weights **deterministic functions of HEXACO + state**, with a small per-agent multiplicative `r_a` vector sampled once from a tight log-normal (σ ≈ 0.25). This restores agent fingerprint while keeping reveries an *expression of* personality rather than orthogonal noise. The mapping uses interpretable linear forms: `examine_scenery_w = 0.05 + 0.25·Openness + 0.10·(1−Conscientiousness) + 0.20·novelty`; `stats_screen_check_w = 0.02 + 0.35·Conscientiousness + 0.20·Achievement_value − 0.10·Openness`; `equipment_hover_w = 0.02 + 0.30·Emotionality + 0.20·recent_damage − 0.10·Conscientiousness`; `pace_w = 0.01 + 0.30·Emotionality + 0.15·(1−Conscientiousness) + 0.40·awaiting_reply`.

The RSC-grounded catalog should include **22 reveries** spanning camera pan/rotate (Openness ↑), examine scenery / examine NPC (Openness, social curiosity), pathing wobble (low-C, Emotionality), brief hesitation (Emotionality), accidental misclick (low-C, fatigue), stats-screen check (Conscientiousness, achievement), equipment hover (Emotionality, recent damage), minimap glance (Openness, spatial uncertainty), chat re-read (Emotionality, social anxiety), AFK micro-break (low-C, escapism), emote (Extraversion), idle facing ritual (Conscientiousness, OCD-like), inventory restack (Conscientiousness), pace (Emotionality, awaiting reply), inventory hover (Openness), landmark walk-by (Openness, Explorer), type-then-erase (Emotionality, low-X), shop price compare (Conscientiousness, achievement), doorway loiter (Conscientiousness, low-X), player scan (Extraversion, Socializer), and prayer/magic tab flip (Emotionality, post-event). Trigger taxonomy: true-idle (> 2s), inter-action micro (200–1500 ms), post-completion, awaiting-response, in-unfamiliar-location, stress-state, social-density. RSC's auto-logout at 5 minutes of total idleness makes anti-AFK fidgeting a genuine player behavior — reveries serve this functionally, not just decoratively.

**Test for flattening empirically.** Log every reverie event over a 24-sim-hour window; compute per-agent multinomial `p_a` over the 22 actions (Laplace-smoothed). Population entropy `H(mean_a p_a)` should exceed the single-template baseline `H(p_template)` by a margin; median pairwise Jensen-Shannon divergence between agents should be > 0.05 (< 0.02 indicates severe flattening). Train a ridge regression `traits ← p_a` — R² should land in [0.3, 0.7]: too low means reveries are noise, too high means they're over-deterministic with no personal jitter. If flattening is detected, increase σ on `r_a`, add stronger state-dependent terms, or inject ~5% "rare-reverie" agents pinned at 3× population mean on a single reverie.

Reveries are mostly state-mediated and rebaselining on their own, but allow **slow capped drift** on the `r_a` vector through critical events: near-death (HP < 10%) permanently lifts `equipment_hover` and `prayer_tab_flip` by ~10%; surviving a scam lifts `examine_npc` and `chat_reread` (vigilance); first city visits temporarily boost `camera_pan` and `landmark_walkby` for ~2 in-sim days. Cap total drift at ‖Δr_a‖ < 0.5 across the run and log every event so post-hoc analysis can attribute behavioral change to specific experiences. This is the Westworld-faithful design: reveries are exactly the site where memory leaks into behavior.

## Population sampling as a multivariate mixture

The 500 agents should not be independent draws from uniform marginals on each trait. **Independent sampling produces psychologically impossible agents** — max-H + max-Power, high-Universalism + high-Dark-Triad — because HEXACO correlates internally (H ↔ A ≈ .20–.35; H ↔ Schwartz Power ≈ −.35 to −.50; latent H ↔ Dark Triad ≈ −0.95 per Hodson et al. 2018) and Schwartz values lie on a circumplex with structural opposites (power vs. universalism, achievement vs. benevolence, conformity vs. self-direction).

Use a **mixture of multivariate Gaussians** over the combined trait+value+preference vector. Each component is an *archetype prototype* with prevalence weight π_k, prototype mean μ_k, and intra-archetype covariance Σ_k (small enough that prototypes are recognizable, large enough that intra-cluster agents differ behaviorally). The 13 components, grounded in Bartle (1996), Yee (2006, *CyberPsychology & Behavior*), Quantic Foundry's six-factor model, and RSC community archetypes documented on classic.runescape.wiki and old-school community essays:

| # | Archetype | π | H | C | X | Top Schwartz values | r | τ |
|---|---|---:|---:|---:|---:|---|---:|---:|
| 1 | Casual social grinder | 0.22 | .60 | .45 | .60 | Benevolence, Hedonism | .35 | .50 |
| 2 | Skiller / Ironman grinder | 0.13 | .65 | .85 | .35 | Achievement, Self-Direction | .25 | .85 |
| 3 | Helpful mentor | 0.09 | .85 | .65 | .60 | Benevolence, Universalism | .30 | .65 |
| 4 | Clan/community loyalist | 0.10 | .60 | .55 | .75 | Benevolence, Conformity | .55 | .55 |
| 5 | Merchant / flipper | 0.06 | .55 | .75 | .55 | Achievement, Self-Direction | .55 | .80 |
| 6 | Quest/lore completionist | 0.07 | .70 | .80 | .40 | Self-Direction, Universalism | .40 | .80 |
| 7 | PKer (legit) | 0.07 | .45 | .55 | .70 | Power, Achievement, Stimulation | .85 | .35 |
| 8 | Mini-game / action enthusiast | 0.06 | .55 | .45 | .70 | Stimulation, Hedonism | .65 | .30 |
| 9 | RP / immersionist | 0.04 | .65 | .55 | .55 | Self-Direction, Universalism | .40 | .60 |
| 10 | Bot-hunter / vigilante | 0.03 | .70 | .70 | .55 | Conformity, Universalism, Power | .55 | .60 |
| 11 | Scammer | 0.04 | .15 | .60 | .70 | Power, Achievement, Hedonism | .65 | .55 |
| 12 | Lurer (Wilderness predator) | 0.02 | .12 | .55 | .60 | Power, Stimulation | .80 | .60 |
| 13 | Off-prototype (broad Gaussian) | 0.07 | mixed |||| mixed |||

This produces a believable skew: **mentors + social grinders + clan loyalists ≈ 41%** (most players are friendly), **PKers + scammers + lurers ≈ 13%** (a salient minority — matching the perception that "the Wilderness feels lawless" even when most players never enter), **merchants ≈ 6%** (small but economically dominant). Cohort experiments parameterize as multiplicative weights on π and optional shift vectors on μ — a `shy_majority` cohort multiplies social-introvert archetypes by 1.3–1.8x and bumps Extraversion μ by −0.15, leaving the covariance structure untouched so cross-cohort comparisons stay clean. An `ethics_stress` cohort multiplies scammer/lurer weights by 2.5x to test contagion of small dishonesty among mid-H agents.

## Two layers: immutable identity, mutable adaptation

The Roberts & DelVecchio (2000) meta-analysis of 152 longitudinal studies establishes rank-order Big-Five stability at r ≈ .64 by age 30 and r ≈ .74 by 50–70. Mischel & Shoda (1995, CAPS) and Fleeson (2001, density-distribution view; 2015 Whole Trait Theory) resolve the person-situation debate: stable individual differences live in the **mean and SD of each person's behavioral distribution**, and in **if-then contingencies**, not in moment-to-moment behavior. For our compressed month-scale simulation this argues for treating deep trait means as effectively immutable (above the empirical .74 plateau) while letting expression states, social ledgers, and learned routines carry all the visible adaptation.

**Immutable fields**: name, backstory, HEXACO factor means (the deep μ vector), top-2 Schwartz value labels, `north_star`, `archetype_tag` (the mixture-component label, for analysis only — invisible to the agent), voice register / baseline formality, core quirks (the trait-derived and idiosyncratic slots), prospect-theory α and β diminishing-sensitivity parameters.

**Mutable fields**: HEXACO *state* (current behavioral expression — Fleeson density mean shifts with social context, social-influence λ ≈ 0.1); full Schwartz value-weight vector (slow re-ranking under big experiences; deep top-2 anchor); current sub-goals derived under the north_star; skill focus (RL-style softmax logits over activities); wealth orientation; current loss-aversion λ and patience τ (slow drift around deep anchor); trust ledger (Beta(α, β) posteriors per other agent); friendships and rivalries (derived from trust ledger); episodic memory store (append-mostly with salience decay); learned routines; mood (valence, arousal — fast updates); reverie `r_a` jitter vector (capped slow drift); clan membership; `learned_emergent` quirk slot.

**External, mutable, observer-side**: reputation. Per the project's "no shared mesa reputation" design commitment, reputation lives as Beta posteriors held by *other* agents, propagated through in-game chat with reduced weight for hearsay (×0.3) and skepticism scaled by the observer's H + C composite.

### Concrete update dynamics

**Trust ledger** uses a Bayesian Beta with H-shaped priors and severity-weighted updates: `Trust(A→B) = Beta(α, β)`, `α₀ = 2 + 4·H_A`, `β₀ = 2 + 4·(1−H_A)`, then `α += w` on cooperation and `β += w` on defection, where w ranges from 0.3 (small chat help) to 8 (major betrayal). Trust below 0.2 with a high-severity event flags rivalry and writes a high-salience episodic memory.

**Skill focus and routine** use soft-RL on logits z_i over activities, softmax temperature T = 1 − 0.5·C_A (high-C agents explore less), and outcome rewards R defined as a dot product of outcome features with the agent's *value* vector — so what counts as reward is itself personality-dependent. Self-Transcendence-high agents get reward from helping outcomes; Power-high agents from item-wealth deltas; this is what makes two agents with identical starting gear diverge in week one.

**Risk tolerance and patience** drift slowly with regression to the deep anchor: `r_t = r_{t−1} + κ(g_t − r_{t−1}) − ρ(r_{t−1} − r_deep)` with κ ≈ 0.02 (learning) and ρ ≈ 0.05 (anchoring). This produces believable "scammed once, now cautious" shifts without permanent personality change.

**Critical life events** — deep-Wilderness death with rare drops, scam victimization above 50% of net worth, betrayal by previously-trusted (>0.7) party, completion of a major quest milestone — trigger 3–5× larger updates on relevant mutable fields, append high-salience memories, can re-rank one mid-tier Schwartz value (Conformity rises after public shaming), and may spawn new sub-goals under the north_star ("never get scammed again," "always help noobs like the one who helped me"). This is where the differential-learning trajectory accumulates.

### How this supports the four research questions

**Long-term strategic accomplishment** is carried by the explicit τ (patience) and Conscientiousness-Diligence parameters, which together govern whether an agent stays the multi-week course toward a 99 in fishing. The current schema's `ambition` keyword cannot do this — it doesn't separate "wants to be wealthy" from "will sit at lobsters for 100 hours." **Organic community formation** falls out of joint trait sampling (compatible agents exist to find each other) crossed with the Beta-updated trust ledger crossed with the Fischbacher-Gächter-Fehr (2001) cooperation type. Conditional-cooperator clusters with similar τ form stable trading partnerships; high-H + high-Benevolence agents become natural mentors; reputation propagation through chat locks in factional "us vs. them" reputations. **Emergent ethics under perceived unobservation** works because the 6% scammer/lurer prototypes carry a *disposition* (μ_H ≈ 0.12–0.15) but the *behavior* must still emerge from situated choice. This lets the research vary perceived observability (Wilderness as off-camera), expected punishment (presence of bot-hunter agents), and cohort composition (does an ethics_stress cohort with 15% low-H produce dishonesty contagion among mid-H agents?). **Believability across weeks** is carried by re-injecting the immutable layer at every inference (Frisch & Giulianelli 2024 show identity drift dominates within tens of turns without re-injection), periodic InCharacter / BFI re-administration as a CI check, embedding-space dispersion monitoring, and the reverie distributional checks above.

Two agents with the same north_star ("be respected in the community") diverge cleanly: a high-C, high-Achievement Merchant pursues it via fair-flipping reputation, while a high-X, low-H Scammer pursues it via being a notorious flexer. Same terminal goal, opposite policies, both believable — and the joint sampling makes such pairs routine rather than hand-authored.

## The recommended schema, end to end

The final structure has four blocks. Annotations: **I** = immutable; **M** = mutable; representation in brackets is sampling form → prompt form.

```yaml
# ---- block 1: identity (all IMMUTABLE) ----
name:                "Eldrin Marsh"                   [I, string]
age_apparent:        34                                [I, int]
backstory:           "<paragraph rendered from archetype + traits>" [I, text]
archetype_tag:       "helpful_mentor"                  [I, debug-only]
cohort_id:           "baseline"                        [I, string]
north_star:          "be known as someone newcomers can trust"  [I, text]
north_star_value:    benevolence                       [I, categorical from Schwartz-10]
secondary_value:     self_direction                    [I, categorical]
voice:
  formality:         mid_low                           [I, band]
  register:          "warm, slightly verbose, uses 'aye'" [I, text]

# ---- block 2: HEXACO disposition (μ IMMUTABLE; state MUTABLE) ----
hexaco:
  honesty_humility:  {percentile: 89, band: high,   mu: 0.84}  [I numeric+band]
  emotionality:      {percentile: 52, band: mid,    mu: 0.51}  [I]
  extraversion:      {percentile: 62, band: mid_high,mu: 0.61} [I]
  agreeableness:     {percentile: 78, band: high,   mu: 0.76}  [I]
  conscientiousness: {percentile: 71, band: mid_high,mu: 0.68} [I]
  openness:          {percentile: 60, band: mid_high,mu: 0.58} [I]
  # salient facets the runtime reads directly
  facets: {diligence: high, fearfulness: mid_low, fairness: high,
           patience: high, inquisitiveness: mid_high,
           social_boldness: mid}                       [I]
  state:                                                # current expression
    expression_drift: {X: +0.03, A: -0.01, ...}         [M, daily]

# ---- block 3: Schwartz values ----
values:
  higher_order:                                          # simplex, sum=1
    self_transcendence:   0.42                          [I top labels; M weights]
    self_enhancement:     0.12
    openness_to_change:   0.28
    conservation:         0.18
  basic_ranked:                                          # full 10 ranked
    [benevolence, self_direction, universalism, ...]    [I top-2; M lower]

# ---- block 4: behavioral-economic preferences ----
prefs:
  risk:
    alpha_gain:    0.88                                 [I]
    beta_loss:     0.88                                 [I]
    lambda:        2.10  (deep)  →  2.32 (current)     [I deep; M current]
    gamma:         0.65                                 [I]
  time:
    delta_daily:   0.9975 (deep)  →  0.9968 (current)  [I deep; M current]
    beta_qh:       0.86                                 [I]
  social:
    alpha_envy:    0.65                                 [I]
    beta_guilt:    0.32                                 [I]
    trust_send:    0.58                                 [I (prior); ledger M]
    coop_type:     conditional_cooperator               [I categorical]
    reciprocity_pos: 0.74                               [I]
    reciprocity_neg: 0.55                               [I]

# ---- block 5: quirks (structured, executable) ----
quirks:
  - {id: qk_01, origin: derived, domain: social, trigger: encounter_newbie,
     modifier: "give 100gp + welcome chat", strength: 0.7,
     observable: "newbie_aid_event", mutability: I,
     narrative: "remembers being broke at Lumbridge bridge"}
  - {id: qk_02, origin: idiosyncratic, domain: ritual, trigger: pre_bank,
     modifier: "always examine the bank booth first", strength: 0.9,
     observable: "examine_event{target=bank_booth}", mutability: I}
  - {id: qk_03, origin: learned_emergent, domain: movement,
     trigger: pre_bank, modifier: "prefer Falador west", strength: 0.4,
     observable: "bank_node distribution", mutability: M}
  # ... 3–7 total

# ---- block 6: reveries (trait-derived weights + personal jitter) ----
reveries:
  schema:             trait_derived                     [I structure]
  jitter_vector r_a:  [1.04, 0.92, 1.15, ...]           [I at creation; M slow drift]
  computed_weights:   <derived at runtime from HEXACO + state + r_a>
  cooldowns_ms:       {default: 15000, per_reverie_overrides: {...}}
  drift_log:          [<event, Δr, timestamp>, ...]     [M append-only]

# ---- block 7: adaptive state (all MUTABLE) ----
state:
  mood:               {valence: +0.2, arousal: 0.4}     [M, fast]
  current_subgoals:   ["fish-lobsters-to-71", "teach-Jeren-Cooking"] [M]
  skill_logits:       {fishing: 2.1, cooking: 1.4, ...} [M, RL]
  wealth_target:      {gp: 250000, asset_mix: "balanced"} [M]
  trust_ledger:       {PlayerX: Beta(7, 2), PlayerY: Beta(3, 6), ...} [M]
  friendships:        [PlayerX, PlayerZ]                [M, derived]
  rivalries:          [PlayerY]                         [M, derived]
  clan:               "Lumbridge Reds"                  [M]
  episodic_memories:  <store with salience decay>       [M]
  learned_routines:   <DSL routine library>             [M]
```

## How to procedurally generate 500 unique-but-believable agents

The pipeline runs **externally** in deterministic Python/Go and only calls the LLM for materialization passes, never to invent traits. **Step 1**: assign each agent a `cohort_id` and resolve the cohort's archetype weights π and μ shift. **Step 2**: sample an archetype k ~ Cat(π), then θ ~ N(μ_k, Σ_k) over the full 16-dim trait+value+preference vector, clipped to valid ranges, with the Schwartz simplex normalized. **Step 3**: compute per-agent reverie jitter r_a ~ LogNormal(0, 0.25) once. **Step 4**: derive the 50% trait-template quirks deterministically from trait combinations (low-H + high-C → pre-scam routine slot; high-O + Explorer → exotic-examine slot). **Step 5**: in a single batched LLM call across 50 agents, materialize (a) the backstory paragraph (from archetype + traits + values, with anti-stereotype prompts per Lutz et al. 2025), (b) the idiosyncratic quirks under verbalized-sampling distribution constraints with hard-banned cliché templates and forced RSC mechanical grounding (Zhang et al. 2025), and (c) the voice-register exemplars. **Step 6**: post-generation validation — SBERT embedding of agent backstories with farthest-point dedupe at cosine ≥ 0.85; population entropy check across (domain × relation) quirk bins; sanity check that no Schwartz-incoherent agents (high Power + high Universalism) survived sampling. **Step 7**: at runtime, re-inject the immutable identity block in the system prompt on every brain call, attach the rendered behavioral paragraph (never raw numbers or band labels), and roll dice externally for stochastic decisions rather than relying on temperature.

The same north_star can yield wildly different behavior profiles because traits, values, and preferences route the agent through different decision policies even when terminal goal language is identical — exactly the differential-learning property the project requires. Two "be wealthy" agents diverge inside one week because their τ, their λ, their H, and their cooperation type all push different first actions, and the trust-ledger and skill-logits state then compounds.

## What changes if the team accepts this

Concretely, six things change. The five categorical keys are replaced with a 16-field three-layer ontology grounded in HEXACO + Schwartz + behavioral economics with empirically-supported joint covariance. The categorical-only representation is replaced with a hybrid: external numeric sampling for science, band labels for legibility, behavioral paragraphs for the actual prompt. Quirks become structured executable modifiers with dimensional decomposition rather than freeform strings. Reverie weights become trait-derived with per-agent jitter rather than independent named-weights bags. Persona generation moves from per-agent LLM invention to external population-mixture sampling with LLM materialization only — eliminating the dominant source of flattening. An explicit immutable/mutable contract anchors month-scale identity coherence while leaving the trust ledger, skill logits, sub-goals, mood, and reverie jitter free to carry differential learning.

The deeper conceptual shift is recognizing that the team's original framing — "floats flatten, categoricals don't" — was tactically right but strategically wrong. The flattening is structural, not representational; it lives in the alignment-induced prior, the Assistant persona's persistent activation, and the gap between expressing a trait label and behaving consistently with it. Fix those, and you can use floats, bands, or paragraphs — they all work. Don't fix them, and no representation will save you. The schema above fixes them by sampling externally from real population data with real covariance, by grounding personas in concrete behavioral acts rather than abstract labels, by re-injecting identity on every inference, by deriving reveries from traits rather than letting the LLM invent them, and by measuring the resulting diversity post-hoc with embedding dispersion, Jensen-Shannon divergences on reverie distributions, and periodic InCharacter / BFI re-administration. That is the architecture worth running 500 agents on for a month.