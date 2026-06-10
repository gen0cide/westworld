# World-Knowledge Ledger & the Learning Drive

> STATUS: BUILT — verified against code 2026-06-10, branch HEAD `0bfa818`;
> file pointers re-checked at `93d3cd1` (the RT-2 split moved the director's
> epistemic/goal code from `runtime/mesa_director.go` into
> `runtime/director_situation.go` / `runtime/director_goals.go`).
> Originated as a stew-on design (2026-06-08, Alex + Claude, mid-load-test); the
> entire phased plan (§9) then shipped as the cognition family-7 commit ladder:
> Phase 0 `106ac72` (+hardening `c5aca73`) · Phase 1 `e493c4e`…`95409a0` ·
> 2a `d129589` / 2b `8a18c24` · 3a `82f82f7` / 3b `2daffee` · 4a `e376aba` /
> 4b `8a2e7d7` · 5a `eb999c4` · 5b-1 `6372684` · 5b-2/3 `2b31068` ·
> stall→ask/forage `8216967`. Phases 5a/5b extended the original plan; their
> specs live in `_research/phase-5b-foraging-plan.md`, `_research/phase-5b-1-impl.md`,
> `_research/phase-5b-23-impl.md`, `_research/phase-5b-revision-backlog.md`.
> Each §3 subsection below carries a BUILT marker with file pointers; §1–§2 are
> the diagnosis of record (past tense — every gap there is closed). Section
> numbers are load-bearing: code comments cite them (`mesa/mesad/cron.go` §3.5/§3.6,
> `cognition/goalgraph/goalgraph.go` §3.7, `runtime/observation.go` §3.5).
> Companions: `docs/mesa/knowledge-pipeline.md` (seed→growth knowledge +
> per-host memory RAG), `docs/_research/social-graph-and-trust-ledger.md`,
> `docs/_research/cognitive-architecture-design.md`. Open work: `docs/TODO.md`
> §3 Cognition (C-1…C-24 — C-5/C-6/C-7/C-8 are the 5b remainder, C-9/C-10 the
> §9 cross-cutting remainder, C-17 the §6 opens, C-21 the §3.8a deferred set).

---

## 0. The shape in one paragraph

When this was written (2026-06-08) a host had a **trust ledger for *people*** but
**no analogous epistemic layer for the *world***: no model of what it knows vs.
doesn't know about items, shops, where-to-get-X, quests, or how things work — and
no *drive* to close those gaps. It was **all exploit, no explore-to-learn**: it
could not represent its own ignorance, so it acted confidently wrong (walked to
*Bob's Axes* for a pickaxe; emitted `go_to("mining-site")` as if that equaled
knowing how to mine). The design adds the missing fourth learning structure — a **World-Knowledge Ledger** (graded
beliefs about things, with confidence + provenance) — plus a **reasoning layer**
that interrogates the host's own knowledge before acting (means-ends +
"open questions"), plus a **learning drive** that turns unresolved, goal-relevant
unknowns into "go find out" actions. The heavy lifting is **two-speed**: the host
(System-1) perceives, reacts, and streams a salience-gated observation firehose up
to mesa; **mesa (System-2)** runs LLM crons that distill that stream into the
knowledge + relationship ledgers — i.e. **memory consolidation / sleep**. Two
structures sit alongside the knowledge ledger and complete the picture: a per-host
**goal graph** (the host's own intentions and how tasks relate — `ore → smelt →
smithing`, `mine ← blocked-by ← bear`, with open questions/open goals as nodes),
which makes actions *purposeful* and failures *generative* and is the anti-stuck
backbone; and **social cognition** that turns player↔player chat from ungrounded
"yapping" into a contextualized two-way channel for learning, teaching, and
knowledge propagation through the community. **All of it manifested** — the
ledger is `cognition/knowledge`, the graph is `cognition/goalgraph`, the
reasoning layer and drives live in `runtime/`, the crons in `mesa/mesad/`.

---

## 1. The problem (diagnosis of record, 2026-06-08 — all three symptoms since fixed)

### 1.1 Symptoms (as observed then)

- **"Where do I buy a pickaxe?"** Delores repeatedly walked to **Bob's Axes**,
  which sells axes (and only a *bronze* pickaxe) — a name-match (pick**axe** →
  Bob's **Axes**), not knowledge. She never learned that **Nurmof** in the Dwarven
  Mine sells the full range, and never recorded that Bob's was a dead end.
  *Fixed by:* the knowledge ledger + perception writers (negative "not sold here"
  knowledge sticks, `runtime/perception.go`), the ASK drive (§3.8a), and the
  forage drive (`runtime/forage.go`, 5b-2/3 `2b31068`).
- **`go_to("mining-site")`** — the planner emitted this as if it were a plan, but
  it was a *blind delegation* to the gazetteer ("route me to the nearest POI tagged
  mining-site"). The host conflated **"I can name a primitive"** with **"I know
  how to achieve the goal."** It never asked: do I know a *reachable* mining site?
  have I mined there before? do I even have a *pickaxe* (the precondition)?
  *Fixed by:* the epistemic `Situation` (§3.2, `d129589`) plus `go_to` dropping
  POI-type targeting for towns-only + catalogued places (`3c0ab33`/`3fde838`).
- **She didn't recognize a player she'd played with a lot.** Root-caused + fixed
  — see §8.2, committed `3214d23` (`runtime/limbic.go:203`) — but it was the same
  family: the host's learned state wasn't capturing what it should.

### 1.2 The unifying failure (then)

> The host **could not tell knowing from guessing**, and nothing forced it to check.

It assumed perfect information each turn, pattern-matched a plausible next move,
and papered over epistemic gaps with confident-looking verbs. There was no
chain-of-thought step that interrogated "what do I need to know / have for this,
and do I actually know/have it?" before committing. "Confidently wrong" was the
signature.

---

## 2. Diagnosis (grounded in the 2026-06-08 codebase — every gap since closed)

What existed for learning, and the hole:

| Captures | Structure | Where |
| --- | --- | --- |
| **WHO** (trust/familiarity w/ people) | `limbic.Ledger` — Beta(α,β); `Met/Observe/Known/Rel/Export/Import` | `limbic/ledger.go`, `runtime/limbic.go` (key `relationship:_ledger`; bbolt + mesa `SyncRelationships`) |
| **WHAT happened** (episodic) | `memory.Journal` — ring of `Episode{kind,text,importance,entity}` | `memory/journal.go`, `runtime/memory_journal.go` (key `journal:_main`; mirrored to mesa LTM via `Remember`) |
| **HOW to talk** (recognition) | `resolve.AliasStore` — learned `text→canonical`, grows w/ play | `cognition/resolve/alias_store.go` (`aliases.json`, learns on brain-resolve) |
| **ABOUT things** (semantic world-knowledge) | was **MISSING** — now `cognition/knowledge` (§3.1) | `cognition/knowledge/knowledge.go`, `runtime/knowledge.go` (key `knowledge:_ledger`) |

Concrete gaps confirmed by the investigation, with how each closed:

1. **Curiosity was decorative.** `persona.Curiosity{Spatial,Skill,Economic,Social,
   Risk}` existed, but was **never read at decision time** (grep of `runtime/` for
   `.Curiosity` was empty) and **never rendered into the LLM character card**
   (`persona/render.go` omitted it). The drive that should power "go learn" did
   nothing. *Closed* (`106ac72`): `persona/render.go` `curiosityPhrase` renders it
   into the prose card, and `curiosityBias` (`runtime/director_situation.go`) reads it at
   decision time for the explore↔exploit cue (§3.3). (`Attention` beyond the
   keyword reflex is still narrow.)
2. **No model of ignorance.** The `Situation` the planner reasoned over was purely
   *observational* ("what I see now") — no "known-unknowns" or confidence-in-facts
   channel (only mood/affect). `Trajectory.SubGoals` was (and remains) a dead
   struct field; there was no means-ends chaining. The only learning-ish signal was
   *reactive* stuck-detection. *Closed* (`d129589`): the epistemic `Situation`
   block in `runtime/director_situation.go` (`epConfFloor=0.5`, `hints["known"]` /
   `hints["unknowns"]`, `sliceBlockers` → `[BLOCKS YOUR GOAL]`, `renderUnknowns`
   anti-bluff block); the goal graph (§3.7) is the real intention structure that
   superseded `SubGoals`.
3. **No semantic-knowledge memory.** Nowhere to store "Nurmof sells pickaxes;
   Bob's doesn't." Worse: the fact wasn't even in our data — OpenRSC shop stock is
   hardcoded in Java plugins and never exported to `facts` (still true, and held
   deliberately: discover-not-seed, §6). The host only saw a shop's stock **after
   opening it** (`world/shop.go`, from `SEND_SHOP_OPEN`) — then forgot it.
   *Closed* (`e493c4e` + `95409a0`): the ledger persists it, and the Tier-0
   perception writers (`runtime/perception.go` — the shop catalogue is the
   flagship writer) record stock, including *negative* knowledge, on sight.
4. **Information-gathering channels existed but didn't write back.** `recall(...)`
   (`runtime/actions_memory.go` `dslRecall`) and `converse/say/whisper` existed,
   but **NPC/player replies were never parsed into stored facts.** *Closed*
   (`8a18c24`): the reactive tier (`runtime/reactive.go`) extracts claims from
   dialog via mesa `ExtractDialog` and writes them to the ledger in <10s
   (`writebackClaims`). Live *global* wiki query stays forbidden by design (no
   firehose) — knowledge is *earned by play* (`docs/mesa/knowledge-pipeline.md`).

---

## 3. The architecture: two-speed cognition

> Everything in §3 is BUILT; each subsection opens with its file pointers. The
> design prose is kept because it is the rationale the code comments cite.

Framing: **memory consolidation.** Lived experience → reflection → semantic
knowledge + refined relationships. The host does fast, cheap, in-the-moment work;
mesa does slow, deliberate consolidation on its own clock ("sleep").

```
HOST  (System-1, real-time, cheap)                MESA (System-2, async, LLM crons)
──────────────────────────────────               ─────────────────────────────────
perceive ─┐                                       ┌─ consolidation cron (Haiku)
          ├─ novelty + salience gate              │    fold observations → ledger
          ├─ emit OBSERVATION stream ───────────► │
deterministic ledger fast-path (familiarity,      ├─ insight cron (Sonnet/Opus)
   trade/duel outcomes)                           │    chain facts, reconcile,
read CURRENT distilled ledger to decide ◄──────── │    close open questions,
in-moment curiosity reaction (optional ask)       │    deferred sentiment
SYNC escalation for core-blocking unknowns ─────► └─ push distilled ledgers down
```

This is **not greenfield**: episodes already mirror to mesa LTM (the stream pipe),
the knowledge-pipeline doc already names a "reflection pipeline" + LTM crons, and
the trust ledger's *local-fast + mesa-mirror* is the exact two-speed template.

### 3.1 The World-Knowledge Ledger (the data layer)

> BUILT (`e493c4e`): `cognition/knowledge/knowledge.go` (`Entry`/`Belief`/`Fact`,
> caps + `Prune`); persistence `runtime/knowledge.go` (key `knowledge:_ledger`,
> warm-start + flush on the limbic cadence, frozen under analysis mode);
> namespace `memory/policy.go` `"knowledge"` (`WriteBack`/`AuthLocal`/
> `CascadeCached`, like `"relationship"`). The shipped shape upgrades the sketch
> below: a `Belief` is **Beta(α,β)-backed** (confidence = the Beta mean
> α/(α+β), evidence-accumulating like the trust ledger) rather than a stored
> scalar, and provenance is the constant set `ProvSystem` / `ProvObserved` /
> `ProvDeduced` / `ProvHearsay`.

A per-host store of **graded beliefs about things** — the sibling to the trust
ledger. Discover, **not seed** (see §6): beliefs are *earned*, and therefore can
be **wrong**.

Original sketch (kept for the rationale; see the package for the real types):

```go
type KnowledgeEntry struct {
    Subject     string         // "rune plate", "pickaxe", "Nurmof", "Varrock east mine"
    Kind        string         // item | shop | npc | place | quest | concept | recipe
    Beliefs     []Belief
    Familiarity int            // encounter/observation count (cf. trust Encounters)
    LastSeen    int64
    Tags        []string
}

type Belief struct {
    Claim      string   // "sells: iron/steel/mithril/rune pickaxe", "is at (x,y)", "requires: champions guild quest"
    Confidence float64  // 0..1
    Provenance string   // "observed" > "did-myself" > "told-by:<player>" > "inferred"
    At         int64
}
```

- **Provenance is load-bearing.** Confidence is gated by *how I know it*:
  saw-it-myself › did-it › was-told › inferred. The crons reconcile contradictions
  and let direct observation override hearsay.
- **Persistence: reuse the trust-ledger spine.** New `knowledge:` namespace in
  `memory/policy.go` (`WriteBack`, `AuthLocal`, `CascadeCached`, like
  `relationship:`); key `knowledge:_ledger`; `load/flush/bootstrap` mirroring
  `runtime/limbic.go` (bbolt local + mesa mirror; cold-start from mesa).
- The reasoning layer (§3.2) **reads** this ledger; the crons (§3.5) **write** it.

### 3.2 The reasoning layer (means-ends + open questions)

> BUILT (`d129589`, Phase 2a): the epistemic block in `runtime/director_situation.go`
> — `epConfFloor=0.5` splits known vs unknown, `hints["known"]`/`hints["unknowns"]`
> feed the Act prompt, `sliceBlockers` labels goal-blocking open questions
> `[BLOCKS YOUR GOAL]`, `renderUnknowns` is the anti-bluff block. Open questions
> are first-class `cognition/goalgraph` nodes (`OpenQuestions()`/`Blockers()`);
> `markGoalBlocked` (`runtime/director_goals.go`) raises them. Success write-back is
> the perception writers (§3.5 speed 1). Phase 5b-1 (`6372684`) added topical
> open questions + the exhaustion-aware picker — spec:
> `_research/phase-5b-1-impl.md`.

The fix for "confidently wrong" is a **pre-action epistemic step**: before
committing to a primitive, decompose the goal and check what's known.

- **Means-ends decomposition.** "To achieve G I need P1…Pn." For each precondition,
  consult the knowledge ledger + world state for a *graded* answer. Resolve the
  **weakest / lowest-confidence** blocker first.
- **Open questions — a first-class primitive.** When the host hits something it
  can't resolve ("what's that armor / how'd they get it", "where do pickaxes
  sell"), that becomes an **open question** in its state — not a passive note. Open
  questions are what the learning drive pursues and what the insight cron tries to
  *close*.
- **Epistemic `Situation`.** Surface goal-relevant unknowns to the planner as
  first-class lines ("You don't yet know where to buy a pickaxe."), alongside a
  confidence dimension (not just mood). Ignorance becomes a planning input, not a
  silent void. This also makes the chain-of-thought **auditable** — the host shows
  "here's what I need, my confidence in each, the gap I'll close first."
- **Self-reinforcing.** A `go_to`/action that *succeeds* writes back "I know the
  Varrock east mine, mined here" (provenance: did-myself, high confidence) — so
  next time it genuinely knows, instead of re-guessing.

### 3.3 Explore ↔ exploit (priority-gated)

> BUILT: `curiosityBias` (`runtime/director_situation.go`) renders the
> decision-time cue from `persona.Curiosity` — core-blocked unknowns demand
> resolution regardless of curiosity; tangents are curiosity-gated detours.
> The drives that *act* on it: ASK (§3.8a) and FORAGE (`runtime/forage.go`).

Both are true: curiosity feeds it **and** there's a hard "you don't know this,
work it out." The arbitration is by **priority**:

```
priority(open_question) = blocks_active_CORE_goal? × persona_curiosity_weight
```

- **Blocks core work** → *hard* "stop and resolve it." The host spends the time;
  may even **synchronously** ask mesa "distill/answer this now" rather than wait
  for a cron.
- **Tangent / whim / sidequest** → curiosity-gated *optional* detour; otherwise
  **note it** (stream up) and return to the main loop — the cron resolves it
  passively over time. (This is Alex's "overridden if it was a whim, but if it's
  core work, spend the time.")

### 3.4 Rich relationships (deterministic spine + LLM overlay)

> BUILT (`2daffee`, Phase 3b): `limbic/ledger.go` is multi-axis — TRUST, AFFINITY
> (`ObserveAffinity`, warmth) and GRIEVANCE (`ObserveGrievance`) are independent,
> with `affinityCap`/`grievanceCap` saturation. Context-gating is live in
> `runtime/limbic.go`: a completed duel bumps familiarity + affinity (sport,
> never a trust hit); a wilderness PK death records grievance + a `ganked-me`
> tag **only when attributable** (`pkKiller()` engaged-target + skull heuristic;
> ambiguous ⇒ record nothing). See §8.1. The cron-maintained **LLM overlay**
> (tone from chat, deferred rumination) is the unbuilt remainder — `docs/TODO.md`
> C-24.

Relationships are **not a single trust scalar** — model them as richly as people
actually are.

- **Deterministic spine** (fast, local, in-the-moment reads): familiarity +
  clear-signal outcomes — trade outcome, duel, **item loss**, etc. (today's
  `limbic.Ledger`).
- **LLM overlay** (cron-maintained, subjective): the nuance LLMs are good at —
  **tone/rudeness** from chat; **separate axes** (I *enjoy dueling* you / I
  *don't trust* you / you *annoy* me — not one number); **deferred rumination**
  ("on reflection, X has been dismissive every time → drop sentiment").
- **Mood loop is free.** Interactions already move affect (`limbic`), so a cron
  can **correlate the mood-trace around interactions with X** — "every duel with
  X I lost items and my mood dropped → I resent X." An *earned* grudge.
- **Duel ≠ PvP (the near-term concrete case).** A duel is a *consensual sport* →
  a "sparring partner / competitor" bond (neutral-to-positive familiarity), **not
  a trust hit**. A *wilderness gank* is betrayal/threat → a big negative + a
  "ganked me" tag. Rule: **combat→trust deltas must be context-gated** (wilderness
  only). Use the ledger's existing `Tags` for typed/contextual relationships. (See
  §8.1; duels never ding trust — that floor held from before the build through it.)

### 3.5 Perception → knowledge: three speeds

> BUILT — all three speeds:
> **Speed 1** (`95409a0`): Tier-0 perception writers, `runtime/perception.go`
> (shop catalogue is the flagship; per-item dedup, NPC attribution, negative
> knowledge).
> **Speed 2** (`8a18c24`, Phase 2b; tuned `0aa9b29`): the reactive tier,
> `runtime/reactive.go` — per-speaker rolling windows with lookback + latch +
> TTL refresh exactly as designed below; `reactiveObserve → triggerHit →
> spawnExtract → mesa ExtractDialog → writebackClaims` lands claims in the
> ledger in <10s, and `closeResolvedQuestions` flips answered open questions
> (`closeQConf=0.6` — authoritative closes, hearsay doesn't).
> **Speed 3** (`e16e039` + `e376aba`/`8a2e7d7`, Phase 4): the firehose —
> `runtime/observation.go` emits salience-gated, fire-and-forget observations
> (capped in-flight; the gate is still the static `observationFloor=0.3` stub —
> `docs/TODO.md` C-20) into mesad `ltm.go` `AddObservation` (deduped), distilled
> by the consolidation cron (`mesa/mesad/cron.go`) and the Tier-2 insight cron
> (`mesa/mesad/cron_insight.go`); push-down back to hosts via
> `FetchKnowledge`/`FetchGoalGraph` (`mesa/client/client.go`).

The host **cannot do all of this in real time** — but some of it it MUST do in
*near*-real-time. Knowledge acquisition runs at **three speeds**, chosen by how
soon the host needs to act on what it just perceived:

1. **Deterministic fast-path** (Tier 0, no LLM, in-loop): structured perception
   the host can parse itself — shop stock, inventory, skills, position — written
   straight to the ledger. Instant, free. (§3.1's shop-stock writer.)
2. **Reactive tier — trigger → high gear** (Tier 1 Haiku/Sonnet, **target <10s**,
   *trigger-driven*): every inbound *signal* — NPC dialog, **player↔player chat**,
   server messages, salient events — is watched by a cheap, **always-on trigger
   detector** (Tier 0, no LLM: the genesis **keyword ladder** — own name / friends
   / trade / goal words, already in `mesa_director.go` — × salience × directed-at-me
   × goal-touch × named-entity). Most signals don't trip it and fall through to
   ambient (speed 3). When one *fires*, that signal **kicks into high gear**: a
   fast extraction-and-intent RPC runs within **~10 seconds** (Haiku, escalating to
   Sonnet for nuance — **not Opus** unless something genuinely needs the depth),
   writing the extracted claims / intent to the ledger and, **when warranted,
   interrupting or steering the current action** (the conductor's existing detour
   path) so the host reacts *now*. The window for "a player just offered a trade /
   warned me of a PKer / asked me a question" is *seconds*, not the minutes-to-
   half-hour a batch cron takes — late is useless. The host stays light: the
   detector is deterministic and the LLM is a mesa RPC; trigger-gated and sparse,
   so cheap. (NPC dialog during questing/tutorial is just the most reliable
   trigger — "the game told me the prerequisite, act on it now.")
3. **Slow firehose → async crons** (Tier 1/2, batched, mesa-scheduled): ambient,
   non-urgent perception, distilled later for the deep, cross-entity, reflective
   work that does *not* need to be ready this turn.

**A trigger opens a conversation *window*, not a line.** The triggering message is
rarely self-contained — "…ok, meet me *there*" needs the lines before it, and the
offer / answer / instruction often arrives in the lines *after*. So the reactive
tier keeps an always-on, **per-speaker rolling buffer** (Tier 0, deterministic,
bounded — the last few lines / ~last minute from each recent speaker), and on a
trigger it: (a) pulls that speaker's **lookback buffer** as pre-context, and
(b) **latches** — for a sustained window it routes that speaker's *subsequent*
lines straight into high-gear extraction (bypassing the trigger gate; the host is
already engaged), the window's TTL **refreshing on each new line** so an active
back-and-forth stays hot and a conversation that goes quiet decays out. The unit
of extraction is the **windowed exchange** (lookback + live tail) — so the host
understands "over these five lines Player X offered a rune scimitar for 30k and
named Varrock west bank as the spot," not five disconnected fragments. Extraction
runs incrementally as the window grows (keeping the reaction <10s) with a final
consolidation when it closes. This extends the director's existing `transcript` +
`lastPlayerMsg` pinning into a per-speaker, trigger-latched conversation buffer.

The slow firehose + crons (speed 3) handle everything that can wait:

- **Emit a salience-gated observation stream.** Typed observations
  (`entity-sighting`, `claim-heard`, `transaction`, `outcome`) emitted cheaply,
  fire-and-forget to mesa (same pattern as episode mirroring). **Most perception
  never streams** — gate by novelty (cheap local "have I seen this before?") ×
  importance × persona curiosity.
- **Consolidation cron** (interval / per-entity threshold; cheap model): folds new
  observations into ledger entries — extract structured claims from chat, merge /
  dedup, sentiment-tag, score salience. Frequent, cheap.
- **Insight cron** (rare; *event-triggered when a new fact touches a pending open
  question*; expensive model): the deep work — **cross-entity chaining**,
  contradiction reconciliation, deferred-sentiment judgment, and **open-question
  closure**. It watches incoming observations for ones that *answer* open
  questions:

  > Saw rune plate on a player → "champions guild quest" (don't know what that
  > is) → **open question**. … Later someone says "the quest needs X, Y, Z" →
  > the insight cron **closes** it into actionable, chained knowledge:
  > `rune plate ← champions guild quest ← prereqs X, Y, Z`.

- **Push distilled ledgers down** to hosts on bootstrap/recall (existing mirror).
- **Knowledge cull cron** = GC: bound ledger size by salience decay.

### 3.6 LLM tiering + cost model

> BUILT — the tier discipline is enforced where the LLM work lives:
> `mesa/mesad/cron.go` (header states the invariants from this section: Tier-0
> no-LLM novelty/dedup/GC; Tier-1 Haiku batched extraction; Tier-2 only on the
> flagged slice; never hold `s.mu` during an LLM call; `cronSem` concurrency
> cap) and `mesa/mesad/cron_insight.go` ("§3.6 is absolute: Sonnet not Opus,
> only on flagged work"). Reactive-tier Sonnet escalation fires only on long
> windows (`0aa9b29`). The **persona-modulated escalation threshold**
> (`BulkApperception`/`Curiosity.*` as the cost-is-personality knob) remains
> design — `docs/TODO.md` C-17.

Tier the work so it's affordable at fleet scale (200+ hosts):

- **Tier 0 — no LLM (deterministic):** novelty detection (seen-set), exact dedup,
  salience decay / GC, recency-frequency scoring. Don't burn even Haiku on "have
  I seen this" or "is this stale."
- **Tier 1 — Haiku/Sonnet (the bulk):** two jobs share the cheap-to-mid tier,
  different latency — (a) the **reactive tier** (§3.5 speed 2): trigger-driven,
  **<10s** extraction-and-intent on signals that trip the attention detector
  (NPC / player / server dialog + salient events), written to the ledger
  immediately and able to interrupt the current action — Haiku, escalating to
  Sonnet for nuance, **not Opus**; and (b) the **consolidation cron**: *batched*,
  deferred per-observation claim extraction, sentiment/tone tagging, merge/dedup,
  and **triage** (flag what needs depth).
- **Tier 2 — Sonnet/Opus (rare, deep):** the insight cron, *only on flagged work*
  — chaining, contradiction reconcile, deferred sentiment, open-question closure.

Routing & multipliers:

- **Escalation is itself a cheap Haiku call.** Most observations die at Tier 1;
  escalate only on a contradiction, an observation that *touches an open
  question*, or a threshold cross. Cost ≈ cheap × volume + expensive × (tiny
  fraction).
- **Batch** (distill an entity's N new observations in one call, not N calls) and
  **prompt-cache the stable prefix** (persona + the entity's current ledger
  entry) — the Act tier already runs cached prefixes; the crons can too.
- **Persona-modulated escalation threshold.** A high-`BulkApperception` /
  high-curiosity host chews on more (escalates readily — thinks harder, costs
  more); a low one lets more pass cheaply. **The cost knob *is* a personality
  knob** — a sharp, reflective host vs. an oblivious one falls out of the same
  dial, and the attention economy becomes a *trait*.

### 3.7 The Goal Graph (intentions & how tasks relate)

> BUILT (`f4ffd0a`): `cognition/goalgraph/goalgraph.go` — typed nodes
> (goal/sub-goal/open-question/state), typed edges, `Blockers`/`OpenQuestions`/
> `OpenGoals` traversal, `ImportMerge` (cron push-down merge) and
> `PruneTerminal` (memory-not-solver bound). Persistence
> `runtime/goalgraph.go` (key `goalgraph:_main`, same spine). Writers:
> `markGoalBlocked` (`runtime/director_goals.go`), death →
> `recordDeathOnGoalGraph` (`runtime/limbic.go` — failures generative, H19),
> question lifecycle in `runtime/reactive.go`/`speech.go`/`forage.go`. Goal
> lifecycle (completion/progress/advancement/spin) is Phase 5a `eb999c4`.
> Mesa grows/prunes it via the insight cron + `FetchGoalGraph` push-down
> (`8a2e7d7`). The fully **goal-graph-aware Act loop** ("the deepest refactor")
> is the open remainder — `docs/TODO.md` C-9.

The third leg of the mind. **Knowledge ledger** = beliefs about the *world*;
**relationship ledger** = *people*; **goal graph** = the host's own *intentions*
and how its tasks relate. At design time there was **none** — which is exactly
why Delores mined but dropped the ore (no link `ore → smelt → smithing`) and died
to the same bear repeatedly (no link `survive-here ← train-combat`). She had
actions, not a structure of *why*.

A per-host directed graph:

- **Nodes:** goals / sub-goals / **open-goals**, **open questions** (the §3.2
  primitive lives here as a node type), and states/resources ("have a pickaxe",
  "smithing L20", "have ore").
- **Edges (typed relations):** `requires` (smith ← bars ← ore), `produces`
  (mining → ore), `enables` (combat → survive-at-mine), `blocked-by`
  (mine ← bear), `serves` (ore → smithing).
- Edges are partly **derived from world knowledge** (recipes/prereqs in the
  knowledge ledger become edges) and partly from the host's chosen objective +
  lived experience.

Why it's the keystone:

- **Makes actions purposeful.** "Keep the ore" falls out of an `ore --serves-->
  smithing` edge existing. Downstream value is just a forward edge.
- **Makes failures generative.** Dying to the bear marks `mine --blocked-by-->
  bear`; traversal spawns an *enabling* sub-goal `train-combat --enables-->
  survive-at-mine`. The host reasons its way out instead of repeating the death.
- **Is the anti-stuck mechanism.** A host stuck in a loop is one with no graph to
  notice "this node advances nothing / keeps failing." Per-node progress + blocked
  state lets the director detect dead ends and re-traverse — and it's
  **inspectable**: you open the graph and *see* why a host is looping. That's how
  we debug a community.
- **Gives priority for free** (§3.3): an open question's / sub-goal's priority ≈
  how central it is + how much it unblocks core nodes. Explore/exploit reads the
  graph.

**Is it too heavy-handed? No — it's the missing backbone, *provided we build it
right*:** a **lightweight, persisted, accreting memory graph** (same trust-ledger
spine; a `goalgraph:` namespace) that the **host reads/traverses** ("what next /
why / what does this failure imply") and that **mesa crons grow and prune** (the
LLM is good at "this failure implies you need X first" and "this resource serves
that goal"). It is emphatically **not** a classical real-time planner the host
solves every tick — *that* would be too heavy. **Memory, not search.** Recallable,
traversable, inspectable, understandable — all four, and none are gold-plating:
traversable+inspectable is how we debug the community, and understandable (it's the
host's *own* stated intentions) is how it stays grounded.

### 3.8 Social cognition — player↔player as a real channel

> BUILT: contextual listening is the reactive tier (§3.5 speed 2) relating
> incoming chat to open questions/goals (`runtime/reactive.go` passes goal +
> questions into `ExtractDialog`; `closeResolvedQuestions`); intent-driven
> speech is §3.8a (`82f82f7`); host↔host propagation rides the teach path +
> `ProvHearsay` extraction described there.

At design time chat was "yapping": persona voice with no informational intent,
ungrounded in goals/knowledge/the conversation — hosts talked *at* each other
instead of listening and learning. Player↔player communication had to become a
**two-way information channel**, because it is how a community *learns* and how
culture emerges.

- **Contextual listening.** Incoming chat is *related to the listener's state*:
  does it **answer an open question**? touch a goal? carry a social cue (tone)? It
  updates the knowledge ledger (the rune-plate example *is* social learning) and
  the relationship overlay (§3.4) — it does not just scroll past. (The first
  substrate was the `OtherPlayerChat → ledger` fix, §8.2 / `3214d23`.)
- **Intent-driven speech.** A host talks because it *wants something*: **ask** to
  close an open question, **answer** when asked what it knows, **coordinate** on a
  shared goal, **teach**. Generation is grounded in the transcript + relationship +
  knowledge state — so it's a conversation, not noise.
- **Host↔host knowledge propagation (emergent culture).** A host that learned
  "Nurmof sells pickaxes" can *tell* another → knowledge, mistakes, and lore
  spread through the community; reputations form. This is the "evolving community"
  payoff, and it rides the same knowledge / relationship / open-question machinery.

This is **"absolutely key"** (Alex) — and it's *why* the ledgers + open questions
come first: you cannot "ask to resolve an open question" or "teach a fact" without
those structures. A grounded-*listening* slice can start early (on the chat fix +
knowledge ledger); intent-driven speech + teaching/propagation deepen after.

#### 3.8a Intent-driven speech (ask / answer / teach) — Phase 3a (BUILT)

Speech is now **goal-serving, not a reflex**. The host talks to **learn** (the ASK
drive) and to **help** (grounded answer + volunteer teach). The project invariant
holds throughout: the **host is the deterministic gate** (no LLM) — it decides
*when* and *who* and supplies *context*; **mesa composes the words** on the cheap
Haiku tier (`Game.Chat`, `ChatTurn.mode`). All of it is **frozen under analysis
mode** like the other learning I/O.

- **The ASK drive (`runtime/speech.go` `tryAsk`).** Fires from `socialReflex` on
  `event.AgentThought` (no standalone ticker — the events are published by the
  director's Act turns, so ask cadence is Act-turn-driven; a genuinely host-owned
  proactive clock is `docs/TODO.md` C-20). Pure predicate, all must hold: not frozen; off the global ask floor; a goal-blocking
  **open question** exists (prefer `goalGraph.Blockers(LiveGoal())`, else newest
  open question) that the host does **not** already know (ledger confidence <
  `epConfFloor`; `how-to-progress:` skips this and relies on cooldown + the attempt
  cap); a **relevant interlocutor** in range (`askRadius=5` Chebyshev; NPC whose
  name the question implicates is boosted, ledger familiarity is a tiebreak, any
  named in-range NPC is a weak fallback, a nearby player a lower hearsay-tier
  fallback); and that target is off its cooldowns. On a full pass it composes via
  `mesa.Ask` (mode=`ask`), **latches** the target, **says** the line (which fans
  into the reactive windows so the Q pairs with the forthcoming A), and records the
  cooldowns + tags the question `asked:<target>`.
- **The loop closes (`runtime/reactive.go` `closeResolvedQuestions`).** The answer
  returns through the existing speed-2 reactive path (`reactiveObserve → triggerHit
  → spawnExtract → ExtractDialog → writebackClaims`). A new tail write in
  `writebackClaims` flips a matching open question to `StatusDone` and un-blocks the
  goal it was blocking — **don't-know → ask → learn → the open question closes.**
  Only **authoritative** claims close it (`closeQConf=0.6`; NPC/server write at
  0.85, player hearsay at 0.5 does not auto-close). It only mutates *existing*
  nodes — no graph growth (honors "memory, not solver").
- **Grounded answer + teach (`runtime/speech.go` `groundReply`).** The reflex reply
  is grounded in the ledger: a confident subject passes a `You KNOW: …` fact (hedge
  if unsure); an unknown subject passes an explicit `you do NOT know — say so, don't
  bluff` line, so **honesty is a host-supplied fact**, not a hope about the LLM. It
  may **volunteer** one high-confidence (`>=0.75`) belief the player's line touches
  (off `teachCooldown`) — the host↔host propagation seed: host A's answer reaches
  host B as a player line → B's reactive tier extracts a `ProvHearsay` claim, so
  knowledge flows host→host through ordinary speech, provenance-tiered weak.

**The `ChatTurn.mode` contract.** `""` / `"reply"` = today's grounded answer path
(backward compatible). `"ask"` = the host proactively asks `from` about `topic`;
mesa branches to an ASK system prompt that forbids bluffing ("you are ASKING, not
telling"), same 80-char cap + silence-on-error. One thin client method `Ask(...)`
wraps it; existing `Chat` callers are unchanged.

**Anti-spam (host-light, all O(1) map lookups, RAM-only, GC'd off the limbic
ticker).** A host that machine-guns questions is worse than a quiet one, so the
gate defaults to silence:

| Constant | Value | Role |
|---|---|---|
| `askGlobalGap` | 45 s | global ask floor (anti machine-gun) |
| `askQuestionCooldown` | 5 min | don't re-ask the same question of anyone |
| `askSameQSameTargetCooldown` | 15 min | don't re-ask the same question of the same target |
| `pesterCooldown` | 90 s | don't fire two different questions at one target back-to-back |
| `maxAskAttempts` | 3 | hard stop per question (tag `ask-exhausted`; cron re-arm = C-21) |
| `askRadius` | 5 tiles | must be in local-chat / `talk_to` range |
| `teachCooldown` | 5 min / player | the host volunteers, it doesn't lecture |
| `closeQConf` | 0.6 | authoritative closes; hearsay doesn't auto-close |

**Deferred set:** tracked as `docs/TODO.md` C-21 (NPC dialog-tree menus for asks;
unprompted first-class teach; LLM-judged fuzzy closure + `ask-exhausted` re-arm +
ask reformulation; multi-turn follow-ups / multi-target coordination;
trust-weighted target choice).

---

## 4. Worked examples

(Design narratives. The machinery behind each is live: pickaxe = ask + forage +
perception writers; rune plate = open question + insight cron; grudge = the
grievance axis; the two goal-graph examples = `serves` edges +
`recordDeathOnGoalGraph`.)

- **Pickaxe (fixed).** Goal needs a pickaxe → means-ends notices "I have none, and
  I don't know where to buy one" → **open question** (core-blocking → hard) → ask
  the nearest miner / recall → "Nurmof, Dwarven Mine" (provenance: told) → go,
  verify, **upgrade to provenance: did-myself** → known forever. And the first
  time she opens Bob's Axes and sees no (real) pickaxe, she records "Bob's: axes
  only" and never returns for one.
- **Rune plate (chaining).** See rune plate → "champions guild quest?" open
  question (low priority, just noted) → later overhear the prereqs → insight cron
  closes the chain → when rune plate becomes goal-relevant, the reasoning layer
  reads a ready dependency chain instead of guessing.
- **Earned grudge.** Repeated item-loss duels vs. X + the mood-trace dropping each
  time → cron concludes "I resent X" → future interactions with X are colored,
  with provenance ("after several losing duels"), not a scripted flag.

- **Downstream value (goal graph).** Mining with `ore --serves--> bars --serves-->
  smithing` in the graph → she *keeps* the ore and heads to a furnace, because
  dropping it would orphan a node she cares about. The action inherits purpose from
  the edge.
- **Failure → enabling goal (goal graph).** A bear kills her at the mine →
  `mine --blocked-by--> bear` is recorded → traversal spawns
  `train-combat --enables--> survive-at-mine` → she goes and trains, then returns —
  instead of dying on the same loop.

---

## 5. What the build reused (grounding)

The "not greenfield" bet held — every reuse named at design time is how it
actually shipped:

- **Two-speed template:** `runtime/limbic.go` (`load`/`flush`/`bootstrap`,
  bbolt + mesa mirror) → cloned for the knowledge ledger (`runtime/knowledge.go`)
  and the goal graph (`runtime/goalgraph.go`).
- **Stream pipe:** `runtime/memory_journal.go` `mirrorEpisode` → `mesaMem.Remember`
  → the observation firehose rides the same shape (`runtime/observation.go`).
- **Namespaces:** `memory/policy.go` `Policy.Classes` was already namespace-routed;
  `"knowledge"` and `"goalgraph"` were added beside `"relationship"`.
- **Recall / RAG:** `mesa` `Knowledge.Recall`, `runtime` `dslRecall`, the corpus
  + per-host memory RAG (`docs/mesa/knowledge-pipeline.md`) — the retrieval seam
  for reading distilled knowledge into prompts.
- **Reflection pipeline + LTM crons:** named in the knowledge-pipeline doc — the
  consolidation/insight crons (`mesa/mesad/cron.go`, `cron_insight.go`) are its
  concrete instantiation.
- **Tags + affect:** `limbic.Ledger` `Tags` (typed relationships — `ganked-me`
  lives there) and the affect vector (the mood-trace for grudges). The goal
  graph reuses the same tag idiom for question state (`asked:<target>`,
  `ask-exhausted`, `source-tried:place:<label>` / `source-spent:place:<label>`).

---

## 6. Decision records (all DECIDED items held; every one manifested as decided)

- **Discover vs. seed — DECIDED: discover.** (Alex.) Knowledge is learned by play,
  not pre-baked. We will *not* export OpenRSC shop stock as seed; the host learns
  it by visiting/asking. (Truer to "she should *learn* it"; slower, more emergent.
  A future hybrid with the worldbrief compile remains possible but is out of scope.)
- **Relationship richness — DECIDED: as rich as possible.** Multi-axis + LLM
  overlay + mood loop (§3.4).
- **Explore/exploit — DECIDED: both, priority-gated** (§3.3).
- **Goal graph — DECIDED: yes, but as a lightweight accreting memory graph**
  (host reads/traverses, mesa crons grow/prune), NOT a real-time planner (§3.7).
- **Social cognition — DECIDED: first-class, "absolutely key"** — contextual
  listening + intent-driven speech + host↔host propagation (§3.8).
- **Information-source resolution (5b) — DECIDED: hybrid exploration + LLM-proposed
  sources.** (Alex, 2026-06-08.) When a goal-relevant fact is missing, the host
  resolves candidate sources two ways, NOT via a hand-curated source-map: (a) **pure
  exploration** — visit unseen shops/locations and let the deterministic perception
  writers (§3.1) harvest what's actually there, incl. *negative* knowledge ("not
  sold here → elsewhere"); and (b) **LLM-proposed sources** — a mesa RPC proposes
  likely source *types/places* for the unknown, which are then **verified against the
  ledger / by visiting** before being trusted (this verification step is the guard
  against the original `go_to("mining-site")` hallucination — a proposal is a
  *hypothesis to test*, never a confident plan). A curated/RAG knowledge seed remains
  a **fallback**, not the primary mechanism. *Shipped as:* ARM-1 pure exploration
  (`runtime/forage.go`, 5b-2/3 `2b31068`); ARM-2 (player-tip verification +
  trust decay) and ARM-3 (`ProposeSources` RPC) are `docs/TODO.md` C-7/C-8.

Remaining opens are tracked in `docs/TODO.md`: C-17 (cron cadence/trigger mix,
sync-escalation build-vs-defer, ledger GC/decay specifics, the persona dials
feeding the escalation threshold) and C-5/C-6 (5b forage termination + dials).

---

## 7. Phasing — the first slice (EXECUTED — all four landed)

The prove-the-spine slice, as designed, with where each landed:

1. **Make curiosity non-decorative** — `106ac72` (`persona/render.go`
   `curiosityPhrase` + `curiosityBias`, now `runtime/director_situation.go`).
2. **Discover + persist shop stock** — `e493c4e` (ledger + `knowledge:`
   namespace) + `95409a0` (the shop-catalogue perception writer). "Bob's = axes
   only" sticks.
3. **One consolidation pass** — `e376aba` (the Tier-1 distillation cron,
   `mesa/mesad/cron.go`).
4. **One open question + means-ends check** — `d129589` (epistemic `Situation`;
   goal-blocking unknowns raised as open questions, "go find out" preferred over
   a blind `go_to`).

It exercised the full vertical — salience emission → mesa cron → distilled
ledger → reasoning-layer read → behavior change — before the build generalized.

---

## 8. Near-term, separable items (both DONE)

### 8.1 Duel ≠ PvP relationship typing — BUILT (`2daffee`, Phase 3b)

Combat→trust is **context-gated** in `runtime/limbic.go`: duel damage/loss =
sparring (familiarity + affinity, **never** a trust penalty; the duel-fight
window `duelFightWindow` gates deaths during a duel out of grievance);
wilderness death-with-an-attacker = `ObserveGrievance` + a `ganked-me` tag. The
v235 Death packet still carries no attacker, so attribution leans on the
engaged-target + skull heuristic (`pkKiller()`); **ambiguous ⇒ record nothing**
(better silent than mis-attributed — the cardinal constraint). The host's own
death while on a goal also writes `blocked_by` + an enabling sub-goal to the
goal graph (`recordDeathOnGoalGraph`).

### 8.2 Trust-ledger public-chat fix — COMMITTED `3214d23`

Root cause of "Delores didn't recognize a player she played with": the limbic
handler recorded `Met` only on `event.ChatReceived` (the *server-message* channel,
which carries a name) — but nearby players' **public chat** arrives as
`event.OtherPlayerChat` (opcode 234 / UpdatePlayers, **index-based, no name**),
which `limbicHandle` ignored. So ordinary conversation never fed the ledger. The
fix: an `OtherPlayerChat` case that resolves `PlayerIndex → name` via
`world.Players.Get` (RLock-safe) and records `Met`, skipping the host's own echoed
chat (`runtime/limbic.go:203` + regression test). Not retroactive.

---

## 9. Development plan — to a living, un-stuck community (ALL PHASES SHIPPED)

**North star (Alex):** get to a community of hosts *playing soon* that aren't stuck
in weird failure loops — visibly *evolving*, *learning*, and *talking to each other
for real*. So the ordering optimized for **observable stability first, depth
after.** Each phase shipped something you can watch. Invariant throughout: the
**host stays light** (read/traverse + emit), **mesa does the heavy growth**, and
**the cradle daemon makes the mind inspectable** (which is *how* we debug the
community).

It was a real cross-cutting refactor — host conductor/director, the `memory`
package, proto/transport, mesa LTM/crons, and the cradle web UI all moved — and
each phase shipped on its own, as planned. Phase bodies below are the plan of
record; each carries its DONE marker.

### Phase 0 — Stop the bleeding — DONE `106ac72` (+hardening `c5aca73`)
*No new heavy infra.*
- The limbic public-chat fix (§8.2) — committed `3214d23`, the listening substrate.
- `Curiosity` made **non-decorative**: `persona/render.go` `curiosityPhrase` +
  decision-time `curiosityBias`.
- **Anti-stuck v0:** the `failStreak` consecutive-failure trigger
  (now `runtime/director_situation.go`). Later deepened by the Phase-5a spin detector
  (`eb999c4`), the world-progress stall detector (`161c1a2`), and
  stall→`markGoalBlocked` (`8216967`) which routes a stall into ask/forage.
- *Outcome held:* fewer weird loops; who's stuck is visible.

### Phase 1 — The mind, made inspectable — DONE `e493c4e`…`95409a0`
*The three persisted structures + the inspector. Observability FIRST.*
- **(host)** Knowledge ledger `e493c4e` (`knowledge:`) and goal graph `f4ffd0a`
  (`goalgraph:`), both on the trust-ledger spine (`load`/`flush`/bootstrap; bbolt
  + mesa mirror). Open questions / open goals are goal-graph nodes.
- **(host→mesa)** Salience-gated **observation stream** `e16e039` (host emit
  `runtime/observation.go` + proto + mesad ingest `ltm.go` `AddObservation`).
- **(cradle daemon)** Mind inspector `3666575`: `debughttp` `/mind` snapshot
  (knowledge / relationships / goal nodes / open questions), proxied by
  `cmd/cradle-server`'s API and rendered as the web-UI **Mind** panel
  (`cradle/web/index.html` `renderMind`); decision stream deepened in `76800d0`.
- **(host)** Live perception writers `95409a0` (`runtime/perception.go`).
- *Outcome held:* each host's mind is watchable — what it knows, who it knows,
  what it's trying to do and why.

### Phase 2 — Reasoning that uses the structures — DONE 2a `d129589` / 2b `8a18c24`
*The "un-confidently-wrong" phase.*
- **(host, 2a)** Means-ends pre-action check; graded-belief consult; open
  questions for goal-blocking unknowns; **failure → enabling sub-goal**
  (`recordDeathOnGoalGraph`); epistemic `Situation` (known/unknowns/blockers) to
  the planner; priority-gated explore/exploit (§3.3).
- **(host, 2b)** The reactive tier (§3.5 speed 2) — dialog answers reach the
  ledger in <10s (Sonnet escalation tuned in `0aa9b29`).
- **(mesa Act)** The prompt carries the goal-graph slice + unknowns, not a flat goal.
- *Outcome held:* hosts consult what they know, flag what they don't, and reason
  out of failures.

### Phase 3 — Social as a real channel — DONE 3a `82f82f7` / 3b `2daffee`
- **(host, 3a)** Intent-driven speech — ask / grounded answer / volunteer teach
  (§3.8a, `runtime/speech.go`); the loop closes via `closeResolvedQuestions`.
- **(host, 3b)** Multi-axis relationships + duel ≠ PvP context-gating
  (§3.4/§8.1, `limbic/ledger.go` + `runtime/limbic.go`).
- **(host↔host)** Propagation rides the teach path: host A's answer reaches host
  B as a player line → B extracts a `ProvHearsay` claim.
- *Outcome held:* conversations that learn and teach.

### Phase 4 — System-2 at scale — DONE 4a `e376aba` / 4b `8a2e7d7`
- **(mesa)** Tier-1 consolidation cron (`mesa/mesad/cron.go`: batched claim
  extraction, dedup, Tier-0 GC, escalation flagging) + Tier-2 insight cron
  (`cron_insight.go`: chaining, contradiction reconcile, open-question closure,
  goal-graph growth/prune) + push-down (`FetchKnowledge`/`FetchGoalGraph`).
  Persona-modulated escalation remains design (C-17).
- *Outcome held:* distillation runs off-host on mesa's clock, tier-disciplined.

### Phases 5a/5b — extensions beyond the original plan
- **5a — goal lifecycle** `eb999c4`: completion / progress / advancement / spin
  detection in the director. Followed by the 54-finding Opus audit ladder
  (`1bcacc4`…`0922bee`), all fixed + regression-tested.
- **5b — directed information-foraging**: 5b-1 `6372684` (topical open
  questions + exhaustion-aware picker), 5b-2/3 `2b31068` (`runtime/forage.go` —
  the tryAsk twin: travel to the nearest untried reachable shop, harvest by
  perception incl. negative knowledge, rotate on a miss), stall→ask/forage
  `8216967`. Specs: `_research/phase-5b-foraging-plan.md`,
  `_research/phase-5b-1-impl.md`, `_research/phase-5b-23-impl.md`,
  `_research/phase-5b-revision-backlog.md`. Remainder: C-5 (termination — the
  open obligation), C-6 (dials), C-7 (ARM-2), C-8 (ARM-3).

### Cross-cutting threads — where they landed
- **Proto/transport:** observation RPC + knowledge/goal-graph sync — shipped
  (`mesa/proto`, `mesa/client`). At-200-hosts batching/backpressure: C-10.
- **Host conductor/director:** the fully goal-graph-aware Act loop ("the deepest
  refactor") is the one designed thread NOT built: C-9. What shipped instead is
  the lighter spine — epistemic Situation + blockers + the ask/forage drives.
- **`memory` package:** `knowledge:` + `goalgraph:` namespaces — shipped.
- **cradle daemon:** inspector UI + plumbing — shipped (`debughttp` +
  `cmd/cradle-server`).
- **mesa LTM/crons:** scheduler, tier routing, raw + distilled storage — shipped.
- **The standing risk** (still the invariant): keep the host **light** and mesa
  **heavy**; the goal graph is **memory, not a real-time planner**; control cost
  with tiering + salience gating. If anything starts making the host *solve*
  rather than *read*, it has drifted into "too heavy."

Open work from this plan lives in `docs/TODO.md` §3 Cognition (the SSOT) — most
directly C-5…C-10, C-17, C-20, C-21.

### Sequencing logic (as it played out)
Phases 0–1 made the community **stable and inspectable**. Phase 2 made hosts
**reason** (stopped confidently-wrong, gave downstream purpose + failure
recovery). Phase 3 made them **social for real**. Phase 4 made it **deep +
affordable at scale**. Phases 5a/5b then closed the loop the field exposed:
goals that *end*, and a host that **leaves to go find out**.

---

## 10. One-line summary

The host has an **epistemic layer** (a knowledge ledger of graded, provenanced
beliefs) and a **reasoning layer** that checks it before acting, fed by a
**salience-gated perception stream** that **mesa LLM crons consolidate while the
host "sleeps"** — so it stops being confidently wrong and starts *learning what it
doesn't know.*
