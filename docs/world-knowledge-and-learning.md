# Design: World-Knowledge Ledger & the Learning Drive

> STATUS: PROPOSED — stew-on (2026-06-08, Alex + Claude). Not built. Raised while
> the 200-drone fight-club load test was running. Companion to
> `docs/mesa/knowledge-pipeline.md` (seed→growth knowledge + per-host memory RAG),
> `docs/_research/social-graph-and-trust-ledger.md` (the relationship/trust
> ledger), and `docs/_research/cognitive-architecture-design.md`. This doc
> captures a design conversation in full so we don't lose it.

---

## 0. The shape in one paragraph

A host today has a **trust ledger for *people*** but **no analogous epistemic
layer for the *world***: no model of what it knows vs. doesn't know about items,
shops, where-to-get-X, quests, or how things work — and no *drive* to close those
gaps. It is **all exploit, no explore-to-learn**: it cannot represent its own
ignorance, so it acts confidently wrong (walks to *Bob's Axes* for a pickaxe;
emits `go_to("mining-site")` as if that equals knowing how to mine). The proposal
adds the missing fourth learning structure — a **World-Knowledge Ledger** (graded
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
knowledge propagation through the community.

---

## 1. The problem

### 1.1 Symptoms (observed)

- **"Where do I buy a pickaxe?"** Delores repeatedly walks to **Bob's Axes**,
  which sells axes (and only a *bronze* pickaxe) — a name-match (pick**axe** →
  Bob's **Axes**), not knowledge. She never learns that **Nurmof** in the Dwarven
  Mine sells the full range, and never records that Bob's was a dead end.
- **`go_to("mining-site")`** — the planner emits this as if it were a plan, but
  it's a *blind delegation* to the gazetteer ("route me to the nearest POI tagged
  mining-site"). The host conflates **"I can name a primitive"** with **"I know
  how to achieve the goal."** It never asks: do I know a *reachable* mining site?
  have I mined there before? do I even have a *pickaxe* (the precondition)?
- **She didn't recognize a player she'd played with a lot.** (Root-caused +
  fixed this session — see §8.2 — but it's the same family: the host's learned
  state isn't capturing what it should.)

### 1.2 The unifying failure

> The host **cannot tell knowing from guessing**, and nothing forces it to check.

It assumes perfect information each turn, pattern-matches a plausible next move,
and papers over epistemic gaps with confident-looking verbs. There is no
chain-of-thought step that interrogates "what do I need to know / have for this,
and do I actually know/have it?" before committing. "Confidently wrong" is the
signature.

---

## 2. Diagnosis (grounded in the codebase)

What exists for learning, and the hole:

| Captures | Structure | Where |
| --- | --- | --- |
| **WHO** (trust/familiarity w/ people) | `limbic.Ledger` — Beta(α,β); `Met/Observe/Known/Rel/Export/Import` | `limbic/ledger.go`, `runtime/limbic.go` (key `relationship:_ledger`; bbolt + mesa `SyncRelationships`) |
| **WHAT happened** (episodic) | `memory.Journal` — ring of `Episode{kind,text,importance,entity}` | `memory/journal.go`, `runtime/memory_journal.go` (key `journal:_main`; mirrored to mesa LTM via `Remember`) |
| **HOW to talk** (recognition) | `resolve.AliasStore` — learned `text→canonical`, grows w/ play | `cognition/resolve/alias_store.go` (`aliases.json`, learns on brain-resolve) |
| **ABOUT things** (semantic world-knowledge) | **MISSING** | — |

Concrete gaps confirmed by investigation:

1. **Curiosity is decorative.** `persona.Curiosity{Spatial,Skill,Economic,Social,
   Risk}` exists, but is **never read at decision time** (grep of `runtime/` for
   `.Curiosity` is empty) and **never rendered into the LLM character card**
   (`persona/render.go` omits it). The drive that should power "go learn" does
   nothing. Same for most of `Attention` (used only for the keyword reflex).
2. **No model of ignorance.** The `Situation` the planner reasons over
   (`runtime/mesa_director.go` `situation()`, `mesa/client` `Situation`) is purely
   *observational* ("what I see now"). There is no "known-unknowns" or
   confidence-in-facts channel (only mood/affect). `Trajectory.SubGoals` is a
   struct field that is **dead code**; there is no means-ends chaining. The only
   learning-ish signal is **stuck-detection** — and it's *reactive recovery*
   ("you haven't moved, try another direction"), not "I have an info gap, fill it."
3. **No semantic-knowledge memory.** Nowhere to store "Nurmof sells pickaxes;
   Bob's doesn't." Worse: **the fact isn't even in our data.** OpenRSC shop stock
   is hardcoded in Java plugins (`.../npcs/lumbridge/BobsAxes.java`,
   `.../npcs/dwarvenmine/NurmofPickaxe.java`) and **never exported** to `facts`
   (`facts/load.go` has no shop defs). The host only sees a shop's stock **after
   opening it** (`world/shop.go`, from `SEND_SHOP_OPEN`) — then forgets it.
4. **Information-gathering channels exist but don't write back.** `recall(...)`
   (own episodic memory / corpus; `runtime/actions_ambient.go`) and
   `converse/say/whisper` exist, but **NPC/player replies are never parsed into
   stored facts.** Live *global* wiki query is forbidden by design (no firehose) —
   knowledge is meant to be *earned by play* (`docs/mesa/knowledge-pipeline.md`).

---

## 3. The architecture: two-speed cognition

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

A per-host store of **graded beliefs about things** — the sibling to the trust
ledger. Discover, **not seed** (see §6): beliefs are *earned*, and therefore can
be **wrong**.

Sketch (a `cognition/knowledge` package, mirroring `limbic.Ledger`):

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
  §8.1; today duels already don't ding trust, which is the correct floor.)

### 3.5 Perception → knowledge: three speeds

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

The third leg of the mind. **Knowledge ledger** = beliefs about the *world*;
**relationship ledger** = *people*; **goal graph** = the host's own *intentions*
and how its tasks relate. Today there is **none** — which is exactly why Delores
mines but drops the ore (no link `ore → smelt → smithing`) and dies to the same
bear repeatedly (no link `survive-here ← train-combat`). She has actions, not a
structure of *why*.

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

Today chat is "yapping": persona voice with no informational intent, ungrounded in
goals/knowledge/the conversation — hosts talk *at* each other instead of listening
and learning. Player↔player communication must become a **two-way information
channel**, because it is how a community *learns* and how culture emerges.

- **Contextual listening.** Incoming chat is *related to the listener's state*:
  does it **answer an open question**? touch a goal? carry a social cue (tone)? It
  updates the knowledge ledger (the rune-plate example *is* social learning) and
  the relationship overlay (§3.4) — it does not just scroll past. (Substrate
  landed this session: the `OtherPlayerChat → ledger` fix, §8.2.)
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
  the `event.AgentThought` tick (a host-owned proactive clock — no ticker). Pure
  predicate, all must hold: not frozen; off the global ask floor; a goal-blocking
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
| `maxAskAttempts` | 3 | hard stop per question (tag `ask-exhausted`; a cron re-arms) |
| `askRadius` | 5 tiles | must be in local-chat / `talk_to` range |
| `teachCooldown` | 5 min / player | the host volunteers, it doesn't lecture |
| `closeQConf` | 0.6 | authoritative closes; hearsay doesn't auto-close |

**Deferred:** relationship/trust-weighted target & teach choice (Phase 3b); driving
real NPC dialog-tree menus (v1 uses local-chat `say` with the name woven in);
unprompted first-class teach; LLM-judged fuzzy closure + re-arming `ask-exhausted`
+ ask reformulation (Phase 4 crons); multi-turn follow-ups / multi-target coordinate.

---

## 4. Worked examples

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

## 5. What to reuse (grounding)

- **Two-speed template:** `runtime/limbic.go` (`load`/`flush`/`bootstrap`,
  bbolt + mesa mirror) → clone for the knowledge ledger.
- **Stream pipe:** `runtime/memory_journal.go` `mirrorEpisode` → `mesaMem.Remember`
  → the observation firehose rides the same shape.
- **Namespaces:** `memory/policy.go` `Policy.Classes` is already namespace-routed;
  add `knowledge:`.
- **Recall / RAG:** `mesa` `Knowledge.Recall`, `runtime` `dslRecall`, the corpus
  + per-host memory RAG (`docs/mesa/knowledge-pipeline.md`) — the retrieval seam
  for reading distilled knowledge into prompts.
- **Reflection pipeline + LTM crons:** already named in the knowledge-pipeline doc
  — the consolidation/insight crons are its concrete instantiation.
- **Tags + affect:** `limbic.Ledger` `Tags` (typed relationships) and the affect
  vector (the mood-trace for grudges) already exist.

---

## 6. Open decisions

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
  a **fallback**, not the primary mechanism.
- **Still open:** exact cron cadence/trigger mix (interval vs. threshold vs.
  event); the open-question schema + how aggressively the reasoning layer pursues
  them; how much of the "sync escalation" path we build vs. defer; ledger GC /
  decay policy specifics; the precise persona dials that feed the escalation
  threshold (`BulkApperception`, `Curiosity.*`).

---

## 7. Phasing — a first slice that touches every layer

Prove the spine end-to-end before the full build:

1. **Make curiosity non-decorative.** Render `Curiosity` into the character card
   and let it weight explore-vs-grind (`persona/render.go` + the Act prompt).
2. **Discover + persist shop stock.** On `SEND_SHOP_OPEN`, write what was seen
   into the knowledge ledger (new `knowledge:` namespace, trust-ledger spine).
   Now "Bob's = axes only" sticks.
3. **One consolidation pass.** A minimal Haiku cron that folds shop observations
   into clean ledger entries with provenance.
4. **One open question + means-ends check.** For "get item X": check the ledger
   for "where sold"; if unknown, raise an open question and prefer "go find out"
   over a blind `go_to`.

This exercises: salience emission → mesa cron → distilled ledger → reasoning layer
read → behavior change, on one narrow vertical (shops/items), before generalizing.

---

## 8. Near-term, separable items

### 8.1 Duel ≠ PvP relationship typing

Make combat→trust **context-gated**: duel damage/loss = sparring (neutral/positive
familiarity + a "duelist/competitor" tag), **never** a trust penalty; wilderness
death-with-an-attacker = large negative + "ganked me" tag. Today duels already
register familiarity-only (no trust hit), which is the correct floor — the work is
(a) not regressing it when we wire combat→relationship signals, and (b) attributing
the wilderness PK, which needs the engaged-player-index + wilderness context
(the v235 damage packet has no attacker — the "unattributed combat" gap).

### 8.2 Trust-ledger public-chat fix (DONE this session, uncommitted)

Root cause of "Delores didn't recognize a player she played with": the limbic
handler recorded `Met` only on `event.ChatReceived` (the *server-message* channel,
which carries a name) — but nearby players' **public chat** arrives as
`event.OtherPlayerChat` (opcode 234 / UpdatePlayers, **index-based, no name**),
which `limbicHandle` ignored. So ordinary conversation never fed the ledger. **Fix:**
added an `OtherPlayerChat` case that resolves `PlayerIndex → name` via
`world.Players.Get` (RLock-safe) and records `Met`, skipping the host's own echoed
chat (`runtime/limbic.go` + regression test). Applies on next host rebuild+restart;
**not retroactive**.

---

## 9. Development plan — to a living, un-stuck community

**North star (Alex):** get to a community of hosts *playing soon* that aren't stuck
in weird failure loops — visibly *evolving*, *learning*, and *talking to each other
for real*. So the ordering optimizes for **observable stability first, depth
after.** Each phase ships something you can watch. Invariant throughout: the
**host stays light** (read/traverse + emit), **mesa does the heavy growth**, and
**cradle makes the mind inspectable** (which is *how* we debug the community).

This is a real cross-cutting refactor — host conductor/director, the `memory`
package, proto/transport, mesa LTM/crons, and the cradle UI all move. The plan
threads it so each phase is shippable on its own.

### Phase 0 — Stop the bleeding (turn on what's dormant; kill the worst loops)
*No new heavy infra. Days, not weeks.*
- **Commit** the limbic public-chat fix (§8.2) — the listening substrate.
- Make `Curiosity` **non-decorative**: render into the character card + weight
  explore↔grind. (`persona/render.go`, the Act prompt.)
- **Anti-stuck v0:** per-(sub)goal *progress + blocked* tracking in the
  director/conductor; when a node isn't advancing, try an alternative / widen
  perception — don't loop. (`runtime/mesa_director.go`, `hybrid_director.go`,
  conductor.)
- **(cradle)** Surface "current goal + stuck?" per host in the web UI.
- *Outcome:* immediately fewer weird loops; we can see who's stuck.

### Phase 1 — The mind, made inspectable (data layer + observability)
*The three persisted structures + the cradle inspector. Observability FIRST so we
can debug the community as it grows.*
- **(host)** Knowledge ledger (`knowledge:`) and goal graph (`goalgraph:`), both on
  the trust-ledger spine (`load`/`flush`/`bootstrap`; bbolt + mesa mirror). Open
  questions / open goals are goal-graph nodes.
- **(host→mesa)** Salience-gated **observation stream** (rides the episode-mirror
  pattern; new proto messages; mind backpressure at 200 hosts).
- **(mesa)** Ingest + store the observation stream; a *minimal synchronous*
  consolidation so the ledgers/graph populate (full crons come in Phase 4).
- **(cradle)** Inspector panels: per-host **knowledge ledger, relationships,
  goal graph (traversable), open questions.** The debugging keystone.
- *Outcome:* we can *watch* each host's mind — what it knows, who it knows, what
  it's trying to do and why.

### Phase 2 — Reasoning that uses the structures (the "un-confidently-wrong" phase)
*The director consults the goal graph + knowledge ledger before acting.*
- **(host)** Means-ends pre-action check; consult the ledger for *graded* beliefs;
  raise open questions for goal-blocking unknowns; **failure → spawn enabling
  sub-goal** (bear → combat); **downstream value** (keep ore → smelt) from forward
  edges.
- **(host)** Epistemic `Situation`: surface unknowns + confidence to the planner;
  priority-gated explore/exploit read from the graph.
- **(mesa Act)** The prompt carries a goal-graph slice + unknowns, not a flat goal.
- *Outcome:* hosts stop being confidently wrong; they pursue chains and reason out
  of failures. Visible "evolution."

### Phase 3 — Social as a real channel (the community phase)
*Contextual listening + intent-driven chat + host↔host propagation.*
- **(host)** Relate incoming chat to open questions/goals/knowledge; intent-driven
  speech (ask / answer / coordinate / teach); write back to the knowledge ledger +
  relationship overlay.
- **(host↔host)** Knowledge propagation; reputation/tone via the relationship
  overlay; coordination on shared goals.
- *Outcome:* conversations that learn and teach; knowledge and lore spread; the
  community visibly evolves. (This is the "absolutely key" payoff.)

### Phase 4 — System-2 at scale (mesa crons + LLM tiering)
*The async distillation that makes it deep AND affordable across 200+ hosts.*
- **(mesa)** Consolidation + insight crons (§3.5): deferred sentiment, knowledge
  chaining, open-question closure, **goal-graph growth/prune**; LLM tiering (§3.6);
  knowledge cull / GC; push-down of distilled ledgers/graphs; persona-modulated
  escalation.
- *Outcome:* emergent grudges / lore / culture; sustainable cost at fleet scale.

### Cross-cutting threads to iron out
- **Proto/transport:** observation-stream RPC; knowledge + goal-graph sync
  messages; backpressure/batching at 200 hosts.
- **Host conductor/director — the deepest refactor:** the Act loop becomes
  *goal-graph-aware* (read/traverse, spawn sub-goals, detect dead ends) instead of
  driving a single flat goal. Touches `mesa_director`, `hybrid_director`, the
  conductor turn loop.
- **`memory` package:** new `knowledge:` + `goalgraph:` namespaces + policies
  (mirror `relationship:`).
- **cradle:** the inspector UI + its data plumbing; fleet-scale stream/cron
  orchestration.
- **mesa LTM/crons:** the cron scheduler, model-tier routing, storage for raw
  observations + distilled structures.
- **The standing risk:** keep the host **light** and mesa **heavy**; the goal graph
  is **memory, not a real-time planner**; control cost with tiering + salience
  gating. If any phase starts making the host *solve* rather than *read*, we've
  drifted into "too heavy."

### Sequencing logic
Phases 0–1 make the community **stable and inspectable** (the near-term ask).
Phase 2 makes hosts **reason** (stops confidently-wrong, gives downstream purpose +
failure-recovery). Phase 3 makes them **social for real** (the community payoff).
Phase 4 makes it **deep + affordable at scale**. Social listening can start as a
thin slice in Phase 1; full teaching/propagation needs Phases 1–2 underneath.

---

## 10. One-line summary

Give the host an **epistemic layer** (a knowledge ledger of graded, provenanced
beliefs) and a **reasoning layer** that checks it before acting, fed by a
**salience-gated perception stream** that **mesa LLM crons consolidate while the
host "sleeps"** — so it stops being confidently wrong and starts *learning what it
doesn't know.*
