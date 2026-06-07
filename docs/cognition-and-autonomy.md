# Cognition & Autonomy — how a host thinks without calling the LLM every turn

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

## 3. The synchronous loop (current vs. target)

**Today** (the loop we're escaping): the autonomous path is `MesaDirector` calling
`mesa.Act` **every turn**, which returns a tiny routine that finishes in seconds —
so Act fires again. L1 (pearl gate) and in-routine `decide()`→pearl exist, but L3
(GoalDirector driving autonomy) and the routine **library** don't, so nothing
absorbs turns.

**Target:**

```
turn ─► GoalDirector.Next()            ◀ L3 local: goal+state → routine
           │
   ┌───────┴────────┐
   │ library match?  │
   └───────┬────────┘
     yes ┌─┘        └─┐ no / novel / stuck
         ▼            ▼
   RUN library    mesa.Act ◀ Sonnet, ONCE
   routine         └─ WRITE_ROUTINE → run + PROMOTE to library
   (loops; 0 LLM)     (next time it's a local RUN_ROUTINE)
         │
   decide() → pearl.TryDecide + decision-cache ; miss → mesa.Decide (Haiku, rare)
```

The first "mine tin" costs an LLM call; the next thousand don't.

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

## 6. Built today vs. seams vs. gaps

**Built:** gRPC mesa (Game/Knowledge/Journal/Provision) + auth; `MesaDirector` Act
loop; pearl gate + `TryDecide`; limbic; `memory.Manager` (wired, unused by the loop);
debughttp dashboard/WS; persona provisioning; model tiering.

**Seams that exist (delivery channels):** Provision Fetch/Subscribe + Directive
kinds; `decide()`→pearl fast path; `when`/`select` watchers; the social reflex
(reply-only).

**Gaps to build:** runtime memory loop (capture+recall); mesa LTM (Journal persist +
Knowledge.Recall retrieval); session-genesis compiler; handler tiers + directedness
classification + committed-region model; **routine suspend/resume + detour stack**
(today the only preempt is `abort`); routine **library** + RUN_ROUTINE reuse +
WRITE_ROUTINE promotion; bounded LRU decision cache; GoalDirector driving autonomy.

## 7. Build roadmap (sequenced)

1. **Memory loop (local)** — episodic capture + recall into the Situation + durable
   objective/progress. *Foundation; everything else consumes it.*
2. **mesa LTM** — Journal persist + Knowledge.Recall retrieval (Voyage + pgvector,
   host-namespaced) + host mirrors episodes up.
3. **Session genesis** — Opus-at-login compile → goal/handlers/keyword-ladder/mood/
   policy via Provision. *(needs memory)*
4. **Interrupt ladder + detour/resume** — tiers, directedness classification,
   committed regions, two-phase, routine suspend/resume + detour stack.
5. **Cheap loop** — routine library + RUN_ROUTINE reuse + promotion + decision cache
   + GoalDirector driving autonomy (LLM becomes escalation-only).
6. **Quick wins** — Act latency/usage instrumentation; chat-codec regression test;
   reinforce autonomous looping routines; cache persona block; trim DSL manual.

**The arc:** memory → genesis → cheap reactive runtime. Concentrate the expensive
cognition into a rare history-rich boot; run the session on its compiled output.
