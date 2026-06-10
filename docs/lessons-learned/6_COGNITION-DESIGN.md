# Cognition design: there is no tick

*Five research papers described a mind with a clock. A 6,109-line validated
architecture got rejected on one paragraph. The "rejected" Pearl and Limbic
shipped five days later. And the live host walked to Bob's Axes for a pickaxe
because nothing in her could tell knowing from guessing. The cognition
subsystem's lessons are about design process as much as design content — read
this before adding anything to the host's mind.*

---

## The problem as experienced

Cognition was designed *onto* a stack that already worked: `Host.Run` pumps RSC
frames → `world.Apply` → synthetic events → `event.Bus.Publish`, and a DSL VM
runs routines that *pull* cognition through builtins (`decide`,
`contemplate_reality`, `recall`). Six commissioned deep-research reports
(gpt55/claude/gemini, architecture + persona each) all assumed a control
structure this code does not have: a 10ms/100Hz subsumption tick (gemini), a
`Reveries.tick()` cadence and an "appraisal loop" (gpt55), a per-action
interceptor (gpt55-persona), a `Reactive` tier doing a synchronous LLM call
inside a flee reflex (claude-arch). The ground truth — verified verbatim in
`docs/_research/reference/decision-cognitive-loop.md` §1 — is that `Host.Run`
spawns exactly one background goroutine (`heartbeatLoop`) and nothing in the
process wakes up on a schedule to think.

The design effort itself then became the problem. The v2 architecture document
grew to **6,109 lines and 44 parser-validated Mermaid diagrams**, built by a
26-agent workflow with two adversarial critic passes
(`docs/_research/reference/REVIEW-GUIDE.v2.md`). It was rejected within a day
of delivery. Its successor's reject table then over-corrected and ruled out
mechanisms (Pearl, the Limbic) that the project built almost immediately
afterward — leaving the resume-point doc asserting "Pearl was REJECTED"
(`docs/_research/SESSION-STATE.md:61`) while `pearl/` sat in the tree, live and
wired. The docs-grounding audit later flagged that line as one of the two most
dangerous inversions in the corpus.

Meanwhile the live host showed what missing cognition actually looks like
(diagnosis of record: `docs/world-knowledge-and-learning.md` §1, 2026-06-08):

- **Delores repeatedly walked to Bob's Axes for a pickaxe** — a name match
  (pick**axe** → Bob's **Axes**), not knowledge. Bob's sells axes. She never
  learned Nurmof in the Dwarven Mine sells the full range, and never recorded
  that Bob's was a dead end.
- **The planner emitted `go_to("mining-site")` as if it were a plan** — blind
  delegation to the gazetteer, conflating "I can name a primitive" with "I know
  how to achieve the goal." No check for a reachable site, prior experience, or
  the pickaxe precondition.
- **She didn't recognize a player she'd played with a lot** — ordinary public
  chat never reached the trust ledger at all (§8.2 of the same doc).
- **She asked the same NPC the same question forever** inside a ~12-tile box
  (`docs/_research/phase-5b-foraging-plan.md:7`), because the ask drive treated
  RSC NPCs as queryable oracles.
- **When stuck, she re-authored the same wrong-terrain script**, because the
  escalation ladder raised call *frequency*, never context
  (`docs/_research/SESSION-STATE-2026-06-09-forage-and-metacognition.md:62`).

## The false leads

**False lead #1: adopt the literature's control structure.** Every paper's
tick/loop/interceptor was seductive because it gave cognition an obvious home.
All of it was rejected in one decision record
(`decision-cognitive-loop.md`, 2026-06-01): *"There is no tick. There never
will be."* The papers' loop was fiction *as a control structure* — but the
**decomposition** it implied (appraise → deliberate → veto → act) was sound,
and the record's entire job was re-mapping that decomposition onto the seams
that exist: appraisal = one bus-subscriber goroutine (sibling to
`heartbeatLoop`), deliberation = pulled by the `decide()` builtin and layered
Pearl-then-LLM, veto = a synchronous check before the Decision returns, action
= the existing verbs, cadence work = mesa crons. Rejecting the structure while
keeping the decomposition is the move; either alone fails.

**False lead #2: doc-scale as design-quality.** v2 had everything process can
buy — 44 validated diagrams, file:line plumbing on every hop, multi-agent
fan-out, two critic passes — and was rejected on a one-paragraph ground: it
*"imported a research-paper component zoo (Pearl/Director/Guard/Limbic/Social)
as in-band components and drew several of them sending packets to the server.
That violates the layering"*
(`docs/_research/cognitive-architecture-design.md:12-16`, the "Why v3"
paragraph; the one rule: **only Session touches the wire**). No amount of
validation rigor checks the one constraint that matters if that constraint
isn't written into the validator. The 543-line v3 that replaced it was built on
verbatim quotes of the real code first (the groundtruth-\*.md file maps in
`docs/_research/reference/`), which is the discipline that killed the
component zoo.

**False lead #3: rejections without expiry conditions.** v3's §6 reject table
threw out "Pearl as a compiled fast-path decision core" ("the fast path already
exists") and "a multi-float affect plane with its own Limbic/sink" ("a
subsystem is unjustified"). Five days later both shipped: the System-1 limbic
loop (`af35919`, 2026-06-06), the conductor + pearl seams (`8c5b146`), the
persona→`pearl.Table`+`TradePolicy` compiler (`4a06e3f`), and the live wiring
(`cb97792`) — all on the same day. They shipped *re-grounded*: Pearl is not an
in-band component but a deterministic table consulted at exactly two seams —
`TryDecide` before the Strategist inside `decide()`
(`runtime/actions_flow.go:197`, `runtime/host.go:96`) and `Gate` wrapping
PrimaryAction handlers (`runtime/dsl_bridge.go:219-221`, `pearl/doc.go`); the
Limbic is one bus-subscriber goroutine maintaining affect + a trust ledger
(`runtime/host.go:111-117`), exactly as `decision-cognitive-loop.md` §3.1 had
re-mapped it. The rejections were right about the *framing* (no in-band
component zoo) and wrong about the *mechanisms* — and because they carried no
expiry condition ("rejected UNTIL someone re-grounds this in the seams"), the
reject table and the session-state doc kept asserting a dead verdict the tree
itself contradicted. The decision records, which had preserved the mechanisms
re-mapped, were the layer of the design corpus that survived contact with the
build; the big synthesis docs were not.

**False lead #4: trusting a synthesized design over the operator's model.**
The transport/storage decision record reasoned carefully about where the LLM
and embedding calls live — and put them in the cradle. Alex's correction is
recorded *in-place* at the top of that doc
(`docs/_research/reference/decision-transport-storage.md` §0, 2026-06-06): the
**compute-locality boundary** is the organizing principle. A host is an
isolated compute unit doing everything compute-local-feasible (world mirror,
routines, conductor, pearl *execution*, Limbic, hostkv) holding **no keys and
making no external calls**; mesa owns LLM inference, RAG, long-term memory,
memory crons, and persona/policy *compilation*. `brain.Strategist.Decide` is a
mesa RPC, not an Anthropic-direct call; the connection is per host, with the
cradle out of the data path. The §0-correction convention matters as much as
the content: the superseding ruling lives at the top of the doc it overturns,
so a future reader cannot absorb the stale model first. (The pre-correction
snapshot of this same doc in `reference.bak/` was judged the most actively
misleading file in the audit and the strongest delete —
`docs/_research/docs-audit/execution-plan.md` Stage 1.)

**False lead #5: earned vocabulary (G2).** The bootstrap design's core
mechanism — *"the grammar is free; the symbol table is earned"*
(`docs/_research/host-bootstrap-and-knowledge-gating.md` §2): hosts would
unlock DSL verbs progressively, validator-enforced, with `punt_count` as the
learning observable. Elegant, prior-art-backed (Soar chunking, Voyager), and
never built — and the evidence ran the opposite way. The DSL-manual analysis
proved **prose presence predicts usage**: the LLM uses exactly and only what
worked examples model, and full disclosure plus worked examples is what made
hosts competent (`docs/brain.md:78`; `docs/TODO.md` C-12). The Plutonium
gap-analysis put the cost of *withholding* surface plainly: our entity views
were richer than the reference bot's, but the field set never reached the LLM —
"we are losing with a better hand because we never show our cards"
(`docs/_research/plutonium/dsl-gap-analysis.md:217`). G2 survives only as a
possible cohort *experiment* (C-12), which is the correct fate for a rejected
design with research value: parked with an explicit re-entry condition, not
deleted and not blocking. (The teaching mechanics live in chapter 5.)

**False lead #6: principled-sounding sequencing gates.** The plan said a
persona/cohort/reverie design session *must gate* the autonomy phase. Events
falsified it: the brain shipped first on one hand-authored persona, and the
persona schema/compiler followed, designed against real runtime needs instead
of speculation (`docs/phases.md:388-393`, `:499-500`; `docs/TODO.md:279`). A
gate that cannot name what concretely breaks without it is a hypothesis, and a
thin vertical slice is how you falsify it cheaply.

**False lead #7: NPCs as queryable oracles.** The ask drive sent questions to
the nearest NPC — but RSC NPCs walk canned dialogue trees; `converse()` is not
free-form Q&A. Facts come from shop stock, exploration, *players*, and curated
knowledge — never NPC interrogation (`docs/_research/phase-5b-foraging-plan.md:7`,
`runtime/forage.go:20`). The fix is structural, not prompt-level: the
interlocutor picker promotes players above NPCs for factual where-to topics
(`runtime/speech.go:467`), tags every tried NPC `source-tried:npc:<name>` so a
spent tree is never re-asked (`runtime/speech.go:478,250`), hard-stops at
`ask-exhausted`, and player answers land as `ProvHearsay 0.5` hypotheses that
route to verification, never auto-closure.

**False lead #8: escalate capability when stuck.** The cheap-loop→Act ladder
escalated *frequency*; the temptation was a bigger model. Alex's correction:
**the gap is CONTEXT, not capability** — a host stuck for lack of terrain
context just re-authors the same wrong-terrain script at any model size. The
design (#32): on a STALL, call the *same* model with a distilled,
decision-ready payload — the host precomputes reachability and passes
*conclusions* ("reachable from here: bed@(a,b); NOT reachable: bed@(c,d)
[behind wall]"), not raw maps — System-1 precomputes, System-2 reasons over
conclusions (`SESSION-STATE-2026-06-09-forage-and-metacognition.md:62`). The
trigger shipped; the rich payload is `docs/TODO.md` C-3, deliberately
re-scoped after the pathing fix because much of "stuck" was the lying
collision grid (chapter 1) — a reminder to re-validate a cognition diagnosis
after the truth-source layer under it changes.

**False lead #9: keyword threat-coding.** The genesis keyword ladder routes
chat like "trade"/"sell"/"free" to a TRADE-INTEREST reflex tier — and the
design explicitly forbids the cheap version: **"free" is not "scam."** A gift
is often genuine; whether an offer is generosity, fair trade, or scam is a
contextual judgment made by disposition + the trust ledger + the actual offer
— never by a word (`docs/cognition-and-autonomy.md:111`, §2). A distrustful
persona is *cautious, not closed*: it screens, it doesn't refuse. Wiring
threat semantics to keywords would have manufactured 500 paranoid hermits and
called it safety.

## The determined fix

The structural decisions that held, and where each lives now:

- **Two paths, no tick.** Continuous, deterministic, zero-LLM appraisal as one
  bus-subscriber goroutine per host; on-demand deliberation pulled by DSL
  builtins, layered local-then-LLM; cadence work (flush/reflection/decay) in
  mesa crons, never in the host (`decision-cognitive-loop.md` TL;DR). The
  invariant is carried as a living code comment where the next maintainer will
  trip over it: the conductor "is deliberately NOT a tick loop"
  (`runtime/conductor.go:17-18`) — the conductor selects Intents and runs them
  to completion on its own goroutine while `Host.Run` stays the frame pump.
- **Freeze the seam, grow additively.** Two method signatures were frozen on
  2026-06-01 — `Strategist.Decide(ctx, Situation) (*Decision, error)` and
  `Client.Retrieve(ctx, Retrieval) (*Bundle, error)`
  (`decision-interfaces-contracts.md` §2.1) — and survived the entire
  build-out unchanged (`brain/brain.go:63`). All growth was additive struct
  fields and new nil-default sibling fields on `Host` (`runtime/host.go:96`:
  nil Pearl = skip the layer, today's behavior preserved, CI green with
  nothing wired). This is why Pearl, the Limbic, the knowledge ledger, and the
  goal graph could each land independently without a single flag-day.
- **The epistemic layer.** A World-Knowledge Ledger of graded, provenanced
  beliefs (`cognition/knowledge`), a goal graph that makes failures generative
  (`cognition/goalgraph`), means-ends interrogation before acting, open
  questions with exhaustion accounting, and the forage drive that makes her
  LEAVE the box — shipped as the commit ladder pinned in the
  `docs/world-knowledge-and-learning.md` header (Phase 0 `106ac72` … 5b-2/3
  `2b31068`, stall→ask/forage `8216967`). The unifying diagnosis — *"the host
  could not tell knowing from guessing, and nothing forced it to check"* (§1.2)
  — is the sentence to test any new planner feature against: provenance gates
  action; closure requires `ProvSystem` (verified-myself), not hearsay.
- **Attribution is a wiring problem before it is a modeling problem.** The
  unrecognized-friend bug was one missing channel: the ledger recorded `Met`
  only on `event.ChatReceived` (the *named* server-message path), but ordinary
  public chat arrives as `event.OtherPlayerChat` — index-based, no name, off
  opcode 234 — which the limbic handler ignored. Fix: resolve
  `PlayerIndex → name` against the world mirror and record the encounter,
  skipping the host's own echoed chat (`3214d23`, 2026-06-08;
  `runtime/limbic.go:199-210`; `docs/world-knowledge-and-learning.md` §8.2).
  Same family as the broader cardinal constraint (chapter 1 of the cognition
  docs, `runtime/limbic.go:195-198`): trust updates fire **only** on signals
  that carry a counterparty name on the wire; ambiguous ⇒ record nothing.
  Better silent than mis-attributed — but also audit that every channel that
  *does* carry identity actually feeds the ledger, or the host literally
  cannot come to know anyone.

## The durable rules

1. **There is no tick.** Import a paper's decomposition, never its control
   structure; re-map every stage onto a seam that already exists before
   accepting a new goroutine, and let the cadence work live in mesa.
2. **Doc-scale is not design-quality.** Validation rigor only checks the
   constraints you encoded; quote the real code (groundtruth maps) before
   designing, and state the governing invariant — "only Session touches the
   wire" — where the validator can see it.
3. **Rejections need expiry conditions.** Reject framings, salvage mechanisms;
   write "rejected until/unless X" or the verdict will outlive its grounds and
   invert a resume doc.
4. **Record authoritative corrections in-place**, as a dated §0 at the top of
   the doc they overturn — and treat any surviving pre-correction copy as the
   most dangerous file in the corpus.
5. **Freeze the seam, grow additively.** Frozen method signatures + additive
   struct fields + nil-default sibling fields are what let five cognition
   subsystems land without a flag-day.
6. **A sequencing gate is a hypothesis** — ship the thin vertical slice on
   hand-authored inputs and let reality decide what actually gates what.
7. **Gate action on provenance.** An agent that cannot model its own ignorance
   is confidently wrong; hearsay opens hypotheses, only verification closes
   them.
8. **Design the learning loop around how the world actually emits facts.**
   NPCs walk canned trees; facts come from shop stock, exploration, and
   players — and a spent source must be remembered as spent.
9. **When an agent is stuck, escalate context before capability** — same
   model, distilled decision-ready conclusions; System-1 precomputes,
   System-2 reasons.
10. **Never keyword-code judgment.** "Free" is not "scam"; threat assessment
    belongs to disposition + relationship + the actual offer. A distrustful
    persona screens, it doesn't refuse.
11. **Full disclosure beat earned vocabulary on evidence** — prose presence
    predicts usage; show your cards, and park the elegant alternative with an
    explicit re-entry condition.
12. **Attribution starves silently.** Abstain on unattributable harm, but
    audit every identity-carrying wire channel into the ledger — the signal
    you never wire up is the friend the host never makes.

## Sources

- `docs/_research/reference/decision-cognitive-loop.md` (2026-06-01) — "There
  is no tick. There never will be."; the two-path re-mapping; the papers'
  fictional loops vs the verified `Host.Run` ground truth.
- `runtime/conductor.go:13-27` — the no-tick invariant as a living code
  comment; the conductor/frame-pump composition.
- `docs/_research/reference/cognitive-architecture-design.v2-rejected.md`
  (6,109 lines) + `REVIEW-GUIDE.v2.md` (44 validated diagrams, 26-agent
  workflow, two critic passes) — the rejected v2.
- `docs/_research/cognitive-architecture-design.md:12-16` ("Why v3", the
  one-paragraph rejection ground) and §6 (the reject table that aged wrong);
  `docs/_research/SESSION-STATE.md:61` (the stale "Pearl was REJECTED"
  inversion); `docs/_research/docs-audit/execution-plan.md` Stage 3 item 45 +
  Stage 1 (the reference.bak delete).
- Commits `af35919` (limbic loop), `8c5b146` (conductor + pearl/memory seams),
  `4a06e3f` (persona→pearl compiler), `cb97792` (live pearl wiring) — all
  2026-06-06, five days after the rejection.
- `runtime/actions_flow.go:197`, `runtime/dsl_bridge.go:219-221`,
  `pearl/doc.go`, `runtime/host.go:96,111-117` — the re-grounded seams.
- `docs/world-knowledge-and-learning.md` §0–§2 (Bob's Axes,
  `go_to("mining-site")`, "knowing from guessing"), §8.1–§8.2 (attribution,
  the public-chat fix `3214d23`), header (the Phase 0→5b commit ladder);
  `runtime/limbic.go:195-210`.
- `docs/_research/phase-5b-foraging-plan.md:7,243-249`; `runtime/forage.go:20`;
  `runtime/speech.go:467,478` — NPCs-are-not-oracles and the exhaustion
  machinery.
- `docs/_research/SESSION-STATE-2026-06-09-forage-and-metacognition.md:62` +
  `docs/TODO.md` C-3 — the stuck-breaker, context-not-capability.
- `docs/cognition-and-autonomy.md` §2 (`:111`) — the keyword ladder and
  "free" ≠ "scam".
- `docs/_research/host-bootstrap-and-knowledge-gating.md` §2 (G2);
  `docs/brain.md:78`; `docs/TODO.md` C-12;
  `docs/_research/plutonium/dsl-gap-analysis.md:217` — the earned-vocabulary
  arc and its evidence-based defeat.
- `docs/_research/reference/decision-transport-storage.md` §0 (2026-06-06) —
  the compute-locality correction.
- `docs/_research/reference/decision-interfaces-contracts.md` §2.1 (the frozen
  seams); `brain/brain.go:63`.
- `docs/phases.md:388-393,499-500`; `docs/TODO.md:279` — the falsified
  persona-gates-the-brain rule.
