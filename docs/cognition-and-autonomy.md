# Cognition & Autonomy — how a host thinks without calling the LLM every turn

> **STATUS: CANONICAL, BUILT** — verified against the code 2026-06-10 (branch HEAD
> `0bfa818`). §0–§2/§4–§5 are the original design narrative, kept verbatim — the
> system manifested faithfully to it (any "today"-style asides in them describe the
> pre-build state). §3 describes the synchronous loop **as built**; §6 maps every
> design element to its shipped implementation (including session genesis's
> **reduced output set**); §7 is the open-work pointer. Open items live in
> [`TODO.md`](TODO.md) §3 (Cognition).

> The canonical model for the host's decision-making: a layered cognitive stack
> where expensive reasoning is **concentrated** (a rare, history-rich compile) and
> the moment-to-moment loop runs **cheap and local**, escalating to an LLM only on
> novelty. This doc folds together the layered-decision, interrupt/detour, session-
> genesis, and memory designs. Companion deep-dives: `layers.md` (interrupt ladder),
> `mesa.md` (the off-host service), `brain.md`/`cognition.md` (the seams),
> `_research/chat-interruption-and-engagement.md` (chat two-phase, full detail).

## 0. The thesis

The failure mode to avoid is **"call the LLM every turn"** — a stateless oracle
re-deriving the world each tick. The fix is a tiered cognition where:

- **Deep reasoning is paid rarely** — once per login (session genesis), plus a few
  boundaries — and *amortized* over a whole session.
- **The per-turn loop is local** — reflexes, a routine VM running real programs
  (loops/watchers), a goal director selecting library routines, a decision cache.
- **The LLM is the escalation target**, not the driver — it's reached on *novelty*
  (no routine for this), *ambiguity* (what does this chat want?), or *replanning*.

Compiler tiers, cheapest first:

```
LOCAL (no LLM)   pearl gate · limbic · routine VM · decision cache · GoalDirector   ← 95% of turns
HAIKU            triage (ambiguous chat), decide() misses, chat replies             ← cheap async
SONNET           mesa.Act — author/select a DSL move when no routine fits            ← escalation
OPUS             session genesis — compile the host's mind from full history         ← rare (login)
```

## 1. The layered cognitive stack

```
                         ┌───────────────────────────────────────────────┐
  SLOW / OFF-HOST        │  L5  mesa LTM + crons (async)                  │  LLM + RAG
                         │      consolidation, reflection, goal/persona    │
                         │      compilation, decay                         │
                         ├───────────────────────────────────────────────┤
                         │  L4  mesa.Act — the PLANNER (System-2)         │  ◀ Sonnet
                         │      "no routine for this; think" →             │
                         │      RUN_ROUTINE | WRITE_ROUTINE | DIRECT       │
                         ╞════════════════ ESCALATION GATE ═══════════════╡
  FAST / ON-HOST         │  L3  GoalDirector — pick a LIBRARY routine     │  local
  (no LLM)               │      from goal + world state                   │
                         ├───────────────────────────────────────────────┤
                         │  L2  Routine VM — a DSL PROGRAM runs           │  local
                         │      loops / repeat-until / when-watchers;      │  (1 routine
                         │      decide() → pearl.TryDecide + cache         │   = many acts)
                         ├───────────────────────────────────────────────┤
                         │  L1  Reflexes — pearl Gate (veto/substitute),  │  local, instant
                         │      limbic affect/trust                        │
                         └───────────────────────────────────────────────┘
            world mirror (frame pump) + memory.Manager feed every layer; the
            interrupt ladder (§4) runs ASYNC across L1–L4 and can preempt.
```

## 2. Session genesis — compile the mind once, run it cheap

The top tier. At **login** (and a few re-genesis boundaries) one heavy **Opus** call
takes the host's *entire* context — persona Cornerstone + full history/LTM + world
snapshot + relationships — and **compiles the session's apparatus**:

```
  LOGIN / re-genesis boundary
        │
        ▼  Opus, full context: "Wake up. Who am I, what have I done, what now,
        │   and how should I react to the world this session?"
        ▼
 ┌──────────────┬───────────────┬──────────────────┬──────────┬───────────────┐
 │ GOAL STACK   │ HANDLER SET   │ KEYWORD→TIER→     │ MOOD     │ POLICY/PEARL  │
 │              │ (session      │ ACTION LADDER     │ BASELINE │ DELTAS +      │
 │              │  on-handlers) │ (chat interrupts) │          │ library picks │
 └──────┬───────┴──────┬────────┴────────┬─────────┴────┬─────┴──────┬────────┘
        ▼              ▼                 ▼              ▼            ▼
   GoalDirector   L2 watchers    interrupt ladder   limbic      pearl + RUN_ROUTINE
   (L3)                          (L1/L2, async)     (L1)        library (L2/L3)
        └──────── delivered via mesa Provision Fetch/Subscribe directives ──────┘
        (PERSONA_REVISION · PEARL_REFRESH · GOAL_REVISION · ROUTINE_UPSERT · REVERIE_REBASELINE)
```

The genesis call is the **producer** for the already-defined Provision/Directive
channel; the host already knows how to apply those. It's the static persona compile
(`CompilePolicy`) *upgraded* to a history-aware session compile.

**Keyword-ladder example** — genesis derives, per host, from history:

```
keyword/pattern        tier              default action (cheap reflex)
─────────────────────  ────────────────  ──────────────────────────────────
"<host name>"          DIRECTED-SOCIAL   orient + triage (detour? reply? continue?)
"<friend> names"       DIRECTED-SOCIAL   orient warmly, lower interrupt threshold
"trade"/"sell"/"buy"/  TRADE-INTEREST    consider it — open to trading (do I need
"offer"/"free"                           this? is it a good deal? a gift to befriend?)
"<her goal topic>"     TOPIC-WATCH       consider engaging (proactive half)
generic chatter        AMBIENT           never preempt
```

(a) the keyword corpus, (b) each one's ladder tier, (c) a default action — all
derived, not hardcoded. Most hits then resolve **locally**; only an ambiguous
directed hit pays the cheap async triage.

**Trade & generosity are NOT keyword-coded as threats.** "free" is not "scam." A
gift is often genuine — someone clearing inventory, or making a friend — and trade is
a first-class economic+social activity hosts *should* pursue (they need resources,
spot good deals, want to give). Whether a specific offer is generosity, a fair trade,
or a scam is a **contextual judgment** made by the host's disposition + the
**relationship** (trust ledger, §5) + the actual offer — never by the word "free."
A distrustful persona is *cautious*, not *closed*; it screens, it doesn't refuse.

**Dependency:** genesis is only as good as the memory feeding it → **build memory
first** (§5). Re-genesis triggers: login, long idle, goal completion/abandonment, a
sustained novel situation, or a periodic cron — *not* every turn.

*(As built, genesis produces a reduced output set — goal + mood + keyword ladder,
delivered by direct RPC rather than Provision directives. See §6.)*

## 3. The synchronous loop — as built

The cheap local loop shipped as **`HybridDirector`** (`runtime/hybrid_director.go`),
which wraps the LLM planner (`MesaDirector` → `mesa.Act`) behind a learned routine
cache. Each turn:

```
turn ─► HybridDirector.Next()                ◀ L3 local (runtime/hybrid_director.go)
           │
   ┌───────┴────────────────┐
   │ situation signature →   │   goal-hash + coarse position + nearby NPC types
   │ RoutineLibrary hit?     │   + inventory bucket + dialog-open flag
   └───────┬────────────────┘
     yes ┌─┘            └─┐ miss · re-validation due · world STALLED
         ▼                ▼
   REPLAY library     MesaDirector → mesa.Act     ◀ Sonnet, ONCE
   routine (0 LLM)     └─ WRITE_ROUTINE | RUN_ROUTINE | DIRECT_ACTION | IDLE
         │                a working authored GRIND is PROMOTED into the library
         │                (next time this situation is a free local replay)
   decide() → pearl.TryDecide → decision cache → mesa.Decide (Haiku, rare)
```

The first "mine tin" costs an LLM call; the next thousand don't.

The cache is **self-healing**: a replayed routine that fails is **evicted** (re-author
next time); a stable signature is re-validated by the LLM every
`maxConsecutiveReuse` (8) turns; and a **world-progress stall detector** (5 turns
with no change to position/fatigue/HP/inventory/XP) evicts the cached routine and
escalates with a STALLED signal (`NoteStall`) so the planner re-plans differently
instead of looping on a "successful but pointless" routine. One-shot moves (a single
say / direct action / idle) are never promoted. The library
(`runtime/library.go`) persists through `memory.Manager`, so a warmed host restarts
mostly-local. Every replay/promote/evict/stall decision streams to the Thoughts
panel (`publishDecision`). Wiring: `runtime/runhost_bootstrap.go` — autonomous mode
is `NewHybridDirector(NewMesaDirector(...), NewRoutineLibrary(host.Memory), goal, log)`.

A deterministic, drive-based **`GoalDirector`** also exists at the same Director
seam (`runtime/director.go`): prioritized `Drive`s with eligibility guards
(`WhenHPBelow`, `WhenFatigueAbove`, `WhenInventoryFull`, …) and an idle fallback —
the no-LLM shape for scripted/guarded behavior.

## 4. The asynchronous axis — interrupts, two-phase, detour/resume

Orthogonal to the turn loop: salient events arrive *while a routine runs*, and the
response is **graded, not binary**. Arbitrated by the **interrupt-priority ladder**
(`layers.md`): handlers declare a tier; the runtime preempts + auto-resumes.

```
   SURVIVAL          low-HP eat / flee / sleep-captcha      ── preempts EVERYTHING (reflex)
   COMMITTED/MODAL   open trade / duel / bank / combat      ── orient but DEFER
   DIRECTED-SOCIAL   your NAME / a watched topic            ── PREEMPTS grind  ◀ needs judgment
   GRIND             the current routine                    ── default
   AMBIENT OBSERVE   passing chatter                        ── attention-gated, never preempts
```

The `attention` persona trait shifts the *edges* (resist vs. yield), never the hard
suppressors. **Two-phase response** is the keystone that makes action-boundary
preemption sufficient (no true mid-action preemption):

```
  addressed ─► PHASE 1 ORIENT (deterministic reflex, NO LLM: face + tiny ack, ~1–2s)
                  │
                  ▼  TRIAGE (cheap, async): given "what I'm doing" + the message →
                  │   GRADED verdict (not binary):
                  ▼
   ignore · orient-only · reply+continue · DETOUR+resume · replan
                                              │              │
                                     suspend→sub→resume   abort→L4 Act
   (Phase-2 LLM reply lands in the chat queue a beat later, off the action loop)
```

**Detour stack** (the "detour then return" mechanic):

```
  chop_trees()  ← GRIND   ── "follow me to the bank" → verdict=DETOUR
     │ suspend @ safe yield point         push
     ▼
  follow_alex_to_bank()  ← detour sub-routine
     ▼ returns                            pop
  resume chop_trees() where it left off
```

This is why the LLM isn't called constantly: ambient chatter never reaches it,
orient is a reflex, triage/reply are cheap+async and only for *directed* events,
full replan is rare.

## 5. Memory — the fuel for all of it

Tiered (`memory.Manager`: scratch → local → remote/mesa). Three kinds:

- **Episodic** — what happened: completed objectives, level-ups, milestones, places,
  people, notable failures. Captured from the bus + completed routines.
- **Semantic** — durable facts she's learned ("the bank is at X", "tin is here").
- **Relational — the trust ledger** (its own first-class store; `limbic.Ledger`
  already exists in RAM). Per *player*: a **Beta(α,β) trust posterior** updated by
  lossless interaction deltas (good trade → +α, scam/betrayal → +β, value traded,
  encounters), plus tags ("generous", "haggles", "scammed me once") and free notes.
  This is the substrate for judging an offer (§2): generosity vs. fair deal vs. scam
  is read off the *relationship + the offer*, not a keyword. It also feeds the
  interrupt ladder (a friend's name lowers her interrupt threshold). Mirrored to
  mesa (`Journal` RelationDelta) for durability + cross-session memory.

**Social presence is proactive, not only reactive.** Today she only notices a player
when they chat *at* her. But every nearby player is in the world mirror
(`world.Players`) and is a legitimate thing to engage **if she wants to** — greet a
passer-by, strike up a trade, offer help, follow a friend. The Situation must surface
nearby players (with bearing) as *interactable presences*, and genesis/persona decide
how outgoing she is — so a sociable host initiates, an aloof one keeps to her grind.
This is the **proactive half** of social engagement (the reactive half is the
interrupt ladder in §4); trade is one of its primary expressions.

Memory feeds **every** layer: the per-turn Situation gets a "what you remember /
already did" + "who's nearby and what I know of them" block; genesis compiles from
the full history; the GoalDirector resumes a plan instead of re-deriving it. Without
memory the loop is stateless and genesis has nothing to reason over — so memory
(episodic + the trust ledger) is the **first** build.

## 6. Where each design element landed (verified 2026-06-10, HEAD `0bfa818`)

Everything in the §1 stack is **BUILT** except the full interrupt-ladder remainder
(TODO **C-11**). The map, layer by layer:

- **L1 pearl gate** — `pearl` package. Every PrimaryAction DSL handler is wrapped by
  `gateAction` (`runtime/pearl.go`): `Pearl.Gate` veto/substitute, with a
  tutorial-island-scoped bypass. `Pearl.TryDecide` is the `decide()` fast path —
  a policy hit answers with no LLM; a miss can still hand back a persona-biased
  option ordering (`runtime/actions_flow.go`, `dslDecide`).
- **L1 limbic** — `limbic` package: the affect vector runs on the `runLimbic`
  bus-subscriber goroutine (`runtime/limbic.go`); the **Beta(α,β) trust ledger**
  (`limbic/ledger.go`) is multi-axis (trust/affinity + tags) and mirrors to mesa
  via `SyncRelationships` / cold-starts from `FetchRelationships`.
- **L2 routine VM** — `dsl/*` through `runtime/dsl_bridge.go`. `decide()` resolves
  pearl → **decision cache** → mesa. The cache is a bounded LRU+TTL
  (`hostkv.Scratch`, 256 entries, 5-min TTL; key = question + option-set + coarse
  HP/fatigue/mood tiers — `runtime/decision_cache.go`); only pearl-*misses* are
  cached, and a miss-of-both pays one Haiku `Decide`.
- **L3 director** — `GoalDirector` (prioritized Drives, `runtime/director.go`) and
  the autonomous loop's `HybridDirector` + `RoutineLibrary` (§3). The library
  persists through `memory.Manager` (`runtime/library.go`).
- **L4 mesa.Act** — `mesa/mesad/act.go`: returns `WRITE_ROUTINE | RUN_ROUTINE |
  DIRECT_ACTION | IDLE` moves; Sonnet by default (`cmd/mesad -act-model`,
  `claude-sonnet-4-6`; Decide = Haiku; genesis = Opus).
- **L5 mesa LTM + crons** — `mesa/mesad/ltm.go`: Postgres schema (episodes with
  pgvector HNSW embeddings via Voyage `mesa/embed`, observations, relationships,
  knowledge, goal_graphs, goals, personas, kv, metrics). `Knowledge.Recall` ranks by
  pgvector cosine similarity when an embedder is wired, degrades to Postgres
  full-text relevance + recency without one. Crons: **consolidation** (`mesa/mesad/cron.go` — Haiku-batched distillation
  of the observation firehose into the knowledge ledger) and **insight**
  (`mesa/mesad/cron_insight.go` — Sonnet, *only* on the pre-flagged escalation
  queue: contradiction reconciliation, cross-entity goal-graph chaining,
  LLM-judged open-question closure).
- **Session genesis — BUILT, reduced output set.** `mesa/mesad/genesis.go` +
  `runtime/runhost_bootstrap.go`. One Opus call at login (default `claude-opus-4-8`)
  reads persona prose + the last 30 episodes + the trust ledger + the standing goal
  (all gathered **mesa-side**) and compiles **three of §2's five channels**: the
  session **goal**, the **mood baseline** (→ `SetAffectBaseline`), and the
  **keyword→tier→action ladder** (→ `SetKeywordLadder` on both the director's
  attention hint and the reactive trigger detector). *Not* produced: session
  handler-sets, pearl/policy deltas, library picks. Delivery is the direct
  `Genesis` RPC, not Provision directives; `Provision.Subscribe` exists and applies
  `GOAL_REVISION` live (`subscribeDirectives`), while `PEARL_REFRESH` /
  `PERSONA_REVISION` are received-and-logged, not yet applied. The only wired
  trigger is `login` (an explicit `-goal` overrides; the persona north-star is the
  no-genesis fallback); idle/goal-completion/cron re-genesis is unwired.
- **Interrupts + detours (§4)** — `runtime/detour.go`: intents run as suspendable
  coroutines (`runtime/coro.go`); an interrupt **parks** the grind, runs the detour
  to completion, and **resumes** exactly where it left off. Shipped tiers, each with
  its own arbiter/hysteresis: **survival** (HP < 35% → eat/flee, re-arm at 55%),
  **fatigue** (≥95% → sleep detour with its own 3-min budget + mid-episode refire),
  **displacement** (single-update ≥8-tile jump = teleport/lure → *abort* + re-plan
  rather than park/resume), **reactive** (an urgent directed signal,
  `runtime/reactive.go` → `maybeInterrupt`), and **forage** (`runtime/forage.go`,
  defers while in a committed dialog). Detours do not nest — a running detour is not
  itself interruptible in this slice.
- **Reactive chat (the §4 reply/triage half)** — two paths. The **reply** reflex:
  `socialReflex` (`runtime/runhost_bootstrap.go`) answers directed chat off the bus
  via mesa `Chat` (Haiku) with rich context (goal + current reasoning + perception)
  and knowledge-grounded honesty (`groundReply`). The **learning** path
  (`runtime/reactive.go`): the deterministic `triggerHit` detector (keyword ladder ×
  salience × directed-at-me × goal-touch, NO LLM) latches a per-speaker
  conversation window → mesa `ExtractDialog` → claims land in the knowledge ledger
  in <10s → an urgent intent raises a conductor detour. Proactive speech is the
  ASK/teach gate (`runtime/speech.go`). The remaining ladder machinery —
  directedness classification beyond the keyword ladder, the full tier set, the
  committed-region model, two-phase ORIENT→REPLY as DSL-level machinery — is
  **C-11**.
- **Memory (§5)** — episodic memory is `memory.Journal` (bounded,
  importance-ranked episodes + the standing objective), captured by the `runMemory`
  bus goroutine (`runtime/memory_journal.go`): warm-starts from local bbolt,
  cold-starts from mesa, flushes on a cadence, and mirrors episodes up via
  `Remember`. The trust ledger shipped exactly as designed (§5). The per-turn
  Situation (`runtime/director_situation.go`, `situation()`) carries the designed
  blocks: durable-memory recall (`memoryHint`), **nearby players annotated with
  trust-ledger notes** (`nearbyPlayersHint` — §5's "today she only notices a player
  when they chat at her" has since shipped), the genesis attention ladder
  (`attentionHint`), and the epistemic layer (goal-graph slice + graded beliefs +
  open questions, `epistemicHints`). DSL verbs `remember`/`recollect`/`forget`
  (`runtime/actions_memory.go`) run over the tiered `memory.Manager`.

## 7. Open work

Open work: see [`/docs/TODO.md`](TODO.md) §3 (Cognition). The one residual from this
doc is **C-11** — the full interrupt ladder (directedness classification beyond the
keyword ladder, full T0–T3 tier set, committed-region model, two-phase ORIENT→REPLY
as DSL-level machinery); C-11's "§7.4" source pointer refers to this doc's former
build-roadmap item 4, which §6 supersedes. Adjacent: C-2 (confidence-scaled routine
commitment), C-9 (goal-graph-aware Act loop).

**The arc held:** memory → genesis → cheap reactive runtime. The expensive cognition
is concentrated in a rare history-rich boot; the session runs on its compiled output.
