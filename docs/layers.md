# Layers — the host, explained from the outside in

> **STATUS: CURRENT** — verified against the code on 2026-06-10, branch HEAD
> `0bfa818`. Every layer below is **BUILT and live** unless a block is
> explicitly labelled **DESIGN (not built)**. Open work is tracked in
> [`TODO.md`](TODO.md) (the SSOT — IDs like `DSL-13`/`C-11` are cited inline).

This is the AI-perspective view of how a westworld host is put
together. It's the right doc to read if you want to understand
**what a host is doing while it's alive in the world** — not the
package layout. For that, see
[`architecture.md`](architecture.md), which slices the same
material by Go package.

A host is one OpenRSC character whose body, senses, reflexes,
mind, and personality are all separate machinery that happens
to share a process. Each layer below is a coherent slice of
that machinery. Layers depend only downward — the higher layers
ask things of the lower ones, never the other way around.

### In-band vs out-of-band — what frames the seven layers

The seven layers are the host's **in-band consciousness**: what it
experiences and controls, the loop of *sense → think → act*. The
host believes it is a person perceiving the world, running its
routines, and taking actions. Everything in this doc below is
in-band unless it says otherwise.

Around that consciousness sits an **out-of-band management plane**
the host never experiences as such — the scaffolding that keeps it
alive and remembers for it:

- **The cradle** — the process machinery that hosts the in-band
  loop. `runtime.RunHost` (`runtime/runhost_bootstrap.go`) brings
  one host up end-to-end: session, world mirror, memory tiers, mesa
  link, persona compile, conductor. The **cradle daemon**
  (`cradle/` + `cmd/cradle-server`) runs and supervises a whole
  fleet of those in one process over shared static deps (facts +
  landscape loaded once, `cradle/deps.go`), with crash/restart
  supervision and pause/resume (`cradle/registry.go`), an HTTP
  control API + live web UI (`cradle/api.go`, `cradle/webui.go`),
  and a CLI (`cmd/cradle-ctl`). Think-turns are scheduled by the
  conductor (§6). Observability taps: the cradle UI's live
  routine/decision panels, the per-host `debughttp` dashboard, and
  `cmd/legacy-cradle -spectate` (a rendered browser viewport).
- **Mesa** — the shared cognition + memory service: a gRPC daemon
  (`mesa/mesad`, binary `cmd/mesad`) backed by Postgres
  (+pgvector for semantic recall, `mesa/mesad/ltm.go`), with
  per-tier LLM clients and background crons that fold the raw
  observation firehose into durable knowledge — bulk Haiku
  consolidation, a rare Sonnet insight pass, no-LLM salience-decay
  GC (`mesa/mesad/cron.go`, `cron_insight.go`). The host
  experiences *"I recall trading with JimBob"*; it never
  experiences the episode write, the embedding, or the salience
  decay underneath.
- **Fleet orchestration** — there is no separate swarm-orchestrator
  service (the early design called one "delos"; it never existed).
  The cradle daemon *is* the technician plane, and mesa's Admin
  service (`mesa/mesad/admin.go`, driven by `cmd/mesa-ctl`) carries
  the fleet-level verbs: generate hosts, set personas, push goals.

**Reveries and persona (layer 7) are cross-cutting injection, not a
top-of-cake layer in the dependency sense.** The host never invokes
its persona; the persona compiles into policy that is applied at
seams of the in-band *action* path (the pearl gate, the decide
fast-path, quirk tics — §7), weighted by a lightweight mood state.
It *colors* the answers the lower layers return rather than issuing
commands of its own. The governing rule of the whole in-band stack
stays simple: **only the session ultimately touches the wire; every
host action flows down through the runtime.** (The runtime view of
this frame is in [`architecture.md`](architecture.md); the reverie
catalog itself is the Phase-5 spec in [`reveries.md`](reveries.md).)

## 1. Wire — the body's nerves

The lowest-level surface: bytes on a TCP socket, framed as RSC
packets, encrypted via ISAAC, with RSA at login (`proto/v235`).
Every other layer sits on top of this. The wire layer is **pure
mechanism** — it doesn't know anything about the game, it just
shuttles opcodes back and forth.

Nothing about cognition or behavior lives here. If you replaced
RSC tomorrow with a different MMO, only this layer would
fundamentally change.

## 2. Session — being logged in

The session layer is "this host is a player named `delores`,
currently connected, authenticated, with an ISAAC keystream and
a heartbeat going." It owns the login dance (`session/handshake.go`:
session-id, RSA-sealed login, ISAAC seeding) and the read/write
goroutines (`session/conn.go`). Once the session is up, everything
above just sees a typed stream of "messages from the server"
and a typed channel of "things to send."

This is the smallest layer worth naming. It exists because
without it every higher layer would have to deal with login
state and connection drops. (Crash *recovery* lives a level up
and out-of-band: the cradle registry restarts a dead host with
backoff — `cradle/registry.go`.)

## 3. World — what the host believes is around it

The host keeps a mirror of the world inside itself (`world/`):
where its body is standing, what's in its inventory, who's nearby,
what ground items are visible, which doors are open and which are
closed. Inbound packets update this mirror; the mirror is what
every higher layer reads from.

A few things to understand here:

- **The mirror is the host's reality.** If the server thinks
  delores is at (122, 658) and the mirror says (121, 657),
  delores is at (121, 657) as far as her own decisions go.
  Drift matters; the layer's correctness is in keeping that
  mirror tight.
- **The world is *this host's* view.** Two hosts in the same
  region maintain independent mirrors. There's no shared
  "server state" in the cradle; each host's model of the world
  is its own.
- **It is not omniscient.** The mirror only contains what the
  server has told this host about — entities outside view
  range simply don't exist in its world model. This is
  important: if delores can't perceive bernard, no amount of
  cognition will make her plan around him.

This layer is the seam between protocol and perception. Below
it: bytes. Above it: meaning.

## 4. Perception & Interaction — the senses and the hands

This is the surface the host actually uses to engage with the
world. It comes in two halves that are easier to think about
together than apart:

- **Perception** is the inbound side: typed *events* the host
  receives when something happens to it or around it. Damage
  taken, item gained, chat received, NPC appeared, dialog
  opened, death. These events flow on an internal pub/sub bus
  (`event/bus.go`); any higher layer can subscribe.
- **Interaction** is the outbound side: typed *actions* the
  host can take (`action/`). Walk to a tile, attack a target,
  eat food, open a door, drop an item, cast a spell, offer a
  trade. Each action is a single atomic verb that turns into
  the right packet on the wire.

These two halves are deliberately separated even though they
sit at the same height. Events are *involuntary* — they happen
to the host whether it wants them or not. Actions are
*voluntary* — the host chooses them. The reason that matters:
the layer above wants to talk about both, but as different
things ("when the host *perceives* X, take *action* Y").

There is a third thing living at this layer: **reflexes** —
pre-cognitive responses wired between events and actions, below
deliberation. These are real today: the conductor's detour stack
carries the survival reflex (HP below 35% → park the routine,
eat/flee), the fatigue reflex (≥95% → find a bed or sleeping bag
before XP stops), and the displacement arbiter (an unexplained
jump of more than 8 tiles in one position update → stop and
re-plan); the social reflex (a player speaks to the host → reply
off-loop) runs as its own goroutine. Reflexes stay narrow; the
moment a response gets context-dependent or goal-aware, it
belongs in the layers above.

**The interrupt-priority ladder — BUILT, as the conductor's detour
stack.** The designed ladder (**survival** > **committed/modal** >
**directed-social / topic** > **grind** > **ambient observe**)
shipped in a recognizable shape:

- **survival** — `runtime/detour.go`: the running routine is a
  suspendable coroutine (`runtime/coro.go`); a survival or fatigue
  trigger **parks** it, runs an eat/flee/sleep detour to
  completion, then **resumes** it exactly where it left off, with
  hysteresis and timeouts so one crisis fires exactly one detour.
- **displacement** — same stack, different verb: a teleport/lure/
  stairs jump **aborts** the grind and bounces control back to the
  director for a fresh decision — there is no deterministic detour
  for "I was unexpectedly moved"; the point is to re-plan.
- **directed-social / topic** — `runtime/reactive.go`: a
  deterministic trigger (keyword ladder × salience × directed-at-me
  × goal-touch, no LLM) latches the speaker into a bounded
  conversation window, extracts the exchange via mesa
  (`ExtractDialog` — Haiku, escalating to Sonnet only for long
  multi-turn exchanges), writes claims into the knowledge ledger
  in under 10 seconds, and can raise a conductor interrupt
  (`maybeInterrupt`) so the host re-plans now instead of next turn.
- **grind** — the current routine (an act-turn, §6).
- **ambient observe** — the salience-gated observation firehose
  (`runtime/perception.go`), attention-gated, never a preemption.

The runtime does the preempt and auto-resume; the host never
juggles priority integers. Missing from the built ladder: the
**committed/modal** tier (an open trade, duel, or bank — orient but
defer) and directedness classification beyond the keyword ladder.
Both are the **C-11** remainder in [`TODO.md`](TODO.md).

**DESIGN (not built): a three-tier handler stack.** Today's
reflexes are still the seed of a planned generalization — a
**handler stack** that keeps reflexes first-class and persistent
across routines instead of dying when a routine ends. Three
ordered tiers, resolved top-down, first-match-wins:

1. **Persona base reflex** — an always-on default the persona
   owns, persisting across every routine. When a host is new this
   is often a raw "punt" up to cognition ("attacked — fight or
   flee?").
2. **Learned specialization** — a context-scoped fast handler the
   host's mind wrote to *replace* that punt in one situation ("in
   this mine, just eat at HP<30, don't ask").
3. **Routine override** — the routine-scoped `on` handlers we have
   today, active only while the routine runs, chained into the
   lower tiers via `extends host` / `super()`.

The parser and validator still reject `extends host` / `super()`;
this is **DSL-13** in [`TODO.md`](TODO.md), unblocked now that the
persona tier it was parked on exists.

> **Where this is built / specified:** detour suspend/resume + the
> displacement arbiter (`runtime/detour.go`, `runtime/coro.go`) and the
> reactive tier (`runtime/reactive.go`) are code. The handler tiers, the full
> tier set, the committed-region model, and richer directedness remain design —
> [`_research/chat-interruption-and-engagement.md`](_research/chat-interruption-and-engagement.md)
> synthesized in [`cognition-and-autonomy.md`](cognition-and-autonomy.md) §4,
> tracked as **DSL-13** + **C-11**. (This is the remaining slice of what older
> docs called "Task #15".)

**Chat: the reply half is built; the orient half is not.** A player
speaking to the host gets a model-composed reply off the Act loop:
`socialReflex` (`runtime/runhost_bootstrap.go`) answers on the
cheap mesa Chat path (Haiku), so a conversation never costs a
routine rewrite, and the reactive tier's per-speaker windows keep a
Q&A paired up across lines. The host also speaks *on purpose*: the
intent-driven speech gate (`runtime/speech.go`) asks goal-blocking
questions of nearby interlocutors (the ASK drive) and volunteers
high-confidence knowledge (teach), with deterministic anti-spam
gating — mesa supplies only the words. The designed **phase-1
orient** reflex — face the speaker and emit a tiny deterministic
ack at the next action boundary, no model call — is not built; it
rides with **C-11**.

## 5. Behavior Runtime — the routine VM

This is the layer where the host gains agency. It's a small
virtual machine that runs `.routine` programs — scripts written
in a custom DSL that say things like "while there's a tree
nearby, chop it, then bank when full, then come back." The VM
turns those programs into sequences of actions at layer 4 and
subscribes to events at layer 4 to react.

The DSL is the *language* — `routine`, `on`, `when`, `select`,
`defer`, `try`, `bounds`, `extends`, etc. (`dsl/parser`). The VM
is the *interpreter* — a sandboxed AST walker (`dsl/interp`) with
budgets (op limit, wall-clock cap, recursion depth, memory —
`dsl/interp/caps.go`) so a routine can't run away with the host.

A few load-bearing properties:

- **Routines are pausable.** When a routine calls `walk_to(x,
  y)`, the VM yields, lets the action play out, and resumes when
  it completes. The host stays responsive to events the whole
  time.
- **Events interleave with actions.** Between every action the
  VM drains the event queue, firing any `on event(...) {...}`
  handlers the routine declared. The same machinery underpins
  `when` (scoped watchers), `select` (block-until-one-fires),
  and `bounds {...}` (region-scoped event filters).
- **Routines compose.** A routine file can `extends
  "common/safety.routine"` to inherit handlers and helpers from
  a library file. Override-and-chain via `extends host` /
  `super()` onto a persona-owned default parent is the open
  **DSL-13** slice.

This is the layer that *does the thing*. If you imagine the
host as a person, this is "follow this recipe to bake bread."
The recipe is concrete, step-by-step, and can be written down
in advance.

## 6. Cognition — choosing which recipe

The cognition layer is where the host stops following a recipe
and starts deciding which recipe to follow. It has two parts
that, like layer 4, are easier to think about together:

- **Memory** — the host's record of what it has experienced; now
  several real stores behind one location-blind surface. *Working
  memory*: the director's rolling narrative transcript (capped at
  80 lines; the last ~18 feed each think-turn —
  `runtime/mesa_director.go`) plus a 256-entry `hostkv.Scratch`.
  *Episodic memory*: the journal (`runtime/memory_journal.go`),
  persisted locally and mirrored up to mesa's Postgres episodes.
  *Relational memory*: the limbic trust ledger (below).
  *Semantic memory*: the world-knowledge ledger
  (`cognition/knowledge`) — graded, provenance-tagged beliefs about
  npcs/places/shops/items — and the intention graph
  (`cognition/goalgraph`). Routines read and write it through the
  location-blind `remember`/`recollect`/`forget` verbs, routed by
  the tiered `memory.Manager` (scratch → local durable hostkv
  (pebble) → mesa) — see [`memory.md`](memory.md). *Reflective*
  compression happens out-of-band: mesa's crons distill the
  observation firehose into the knowledge ledger and goal graph
  while the host gets on with its life.
- **Reasoning** — the actual decision-making. The host examines
  its ledgers and the world mirror, consults static game facts
  (`facts/`), and asks mesa to plan. (A keyword-indexed rsc.wiki
  corpus behind the `recall()` builtin exists — `cognition/corpus`
  — but is wired only in `cmd/legacy-cradle`; the daemon-path host
  learns through the knowledge ledger + mesa instead.)

The reasoning step is intentionally tiered — and every tier is
real:

- **Routine decisions** (do I have food? am I being attacked?
  where's the nearest banker?) are answered locally and
  deterministically: the world mirror, facts tables, the pearl
  decide fast-path (`pearl.TryDecide`, consulted before any
  strategist — `runtime/host.go`), and the **cheap loop**:
  `runtime/hybrid_director.go` promotes a repeated, succeeding
  plan to a cached routine and replays it with zero model calls
  (stall/spin detectors evict it and force a re-plan when it
  stops working). No LLM call.
- **Tactical decisions** consult a small model: mesa `Decide`
  (Haiku) for narrow option-picks, `Chat` (Haiku) and
  `ExtractDialog` (Haiku, Sonnet for nuance) for the
  reactive/social path.
- **Strategic decisions** consult a larger model: mesa `Act`
  (Sonnet) plans the next move each think-turn, and session
  **genesis** (Opus, once per login — `mesa/mesad/genesis.go`)
  compiles the host's history into this session's goal, mood
  baseline, and attention keyword ladder.

This tiering matters because the project goal is 500 hosts
running 24/7. If every "should I eat?" went through a frontier
LLM, the cost would be ruinous. The split keeps the expensive
reasoning rare and the cheap reasoning constant.

Cognition is what makes a host *adapt*. The behavior runtime
runs the same routine the same way every time; cognition
chooses which routine to run, when to stop, and when to start
something different.

**BUILT: the agent-driver loop — how cognition actually gets a
turn.** The host's life alternates **act-turns** and
**think-turns**, exactly as designed. The **conductor**
(`runtime/conductor.go`) — deliberately *not* a tick loop; the
frame pump (`Host.Run`) keeps the mirror live on its own goroutine
— runs the turn cycle: ask the director for an Intent, run it to
completion as a (suspendable) routine, observe the Outcome, settle,
repeat. The **MesaDirector** (`runtime/mesa_director.go`) is the
think-turn: it assembles a *Situation* — world snapshot, affect,
rolling transcript, graded beliefs and explicit unknowns,
goal-graph state, curiosity bias, stuck/spin/stall signals — and
asks `mesa.Act`, which returns exactly the designed decision set
(`MoveKind`, `mesa/client/client.go`): **run a known routine**,
**write a new routine** then run it (parsed, validated,
pearl-gated, quarantined until reviewed), take a single **direct
action**, or **idle**. The **HybridDirector**
(`runtime/hybrid_director.go`) wraps it with the cheap local loop
so repeatable grinds replay with no model call (one-shot intents
are never cached). One think-turn is one bounded reasoning call;
one act-turn is thousands of actions. That ratio is still the whole
economic argument for running hundreds of hosts at once.

Two things this section once flagged as "designed but not yet
built" have since resolved in opposite directions:

- **The trust ledger is BUILT — multi-axis, simpler prior.**
  `limbic/ledger.go`: TRUST is the designed Bayesian Beta(α,β)
  posterior — mean mapped to [-1,1], familiarity = evidence count,
  so "sure he's reliable, 200 trades" reads differently from "weak
  hunch, met once" — but it shipped with a uniform Beta(1,1) prior
  rather than a disposition-seeded one, and relationships grew two
  more *independent* axes: AFFINITY (warmth) and GRIEVANCE
  (accumulated wrongs). The runtime's limbic goroutine
  (`runtime/limbic.go`) feeds it from named-counterparty events
  (trades, chat, PMs, duels, projectiles), persists it through the
  memory layer, and mirrors it to mesa (`SyncRelationships`). The
  DSL reads it via `relation_with`. The remainder — LLM severity
  grading, decay-toward-prior, the prior question, richer DSL
  reads — is **C-24** + **DSL-24** in [`TODO.md`](TODO.md).
- **DESIGN (not built): scripting competence earned, not given.**
  The falling-punt-rate / progressively-disclosed-vocabulary
  mechanism never manifested — hosts get the full DSL manual, and
  the dsl-manual analysis found that full disclosure is precisely
  what made hosts competent. It survives as a possible cohort
  experiment (**C-12**, operator decision pending; punt-rate
  instrumentation is **R-3**).

## 7. Persona & Reveries — who the host is

The top layer is the part the host **doesn't see**. It's the
fingerprint that makes this host *this host* — not a generic
agent, not a copy of every other host. Two pieces:

- **Persona is configuration — BUILT.** The `persona/` package is
  the schema SSOT (`persona/persona.go`: identity, disposition,
  drives, quirks, speech style, the Curiosity flavor vector, the
  Attention anchor, a ReverieSeed with per-reverie jitter sampled
  once; `enums.go` for the authored vocabulary; `validate.go`).
  When mesa is linked it is authoritative: the persona provisions
  down at startup and **compiles locally** into an executable
  policy (`persona.CompilePolicy`, `persona/compile.go` → pearl
  rules) — the func-valued policy table never crosses the wire.
  The persona object itself is *discarded after compile*; only the
  compiled policy and a few captured dials (e.g. curiosity) survive
  on the host (`runtime/host.go`), so neither routines nor
  cognition can introspect "what is my persona?" — the DSL spec
  reserves a `host.persona` view the runtime deliberately does not
  implement.
- **Reveries is the injection machinery — the seams exist; the
  catalog is the Phase-5 spec.** What runs today is the **pearl
  engine** at three seams: `Gate` (a synchronous veto/observe on
  every action — `runtime/pearl.go`), `TryDecide` (the no-LLM
  decision fast-path biased by disposition —
  `runtime/actions_flow.go`), and `Injections` (`EffectInject`
  quirk/reverie micro-action tics — `pearl/engine.go`). The full
  catalog the design describes — timing jitter, path imperfection,
  banker selection, the occasional misclick, injected at every
  action call site — is specced in [`reveries.md`](reveries.md)
  and not yet wired at those call sites.

The reason persona is split from reveries is that *persona is
the noun and reveries is the verb*. You can write a persona
without knowing how the runtime will use it. You can change the
runtime's reverie machinery without rewriting personas. And
admins can author personas in external tools that don't drag
in the whole runtime.

The principle here is **the host does not write naturalistic
code**. The routines don't say "wait with jitter" or "pick a
random banker." They say "wait" and "give me a banker," and
the runtime silently varies the answer in a way the host
script never has to think about. Today that variation comes
from the pearl seams above; per-call-site variation arrives
with the reverie engine. That's how 500 hosts running the same
routine end up feeling like 500 different people.

**BUILT: the persona's exploratory temperament.** The persona
carries, beyond its behavioral knobs, two orthogonal dispositions
that shape *what a host does with its time* — and both are live:

- **Curiosity** — a *flavored drive* (a 5-flavor weight vector over
  social / spatial / skill / economic / risk —
  `persona/persona.go`), captured at bootstrap and read at decision
  time: `curiosityBias` (`runtime/director_situation.go`) renders an
  explore↔exploit cue into the Act prompt (a strong dominant flavor
  earns a detour cue when nothing blocks the core goal; a
  low-curiosity persona stays on task), and the mesad insight
  cron's escalation bar is modulated by the persona drive
  (`mesa/mesad/cron.go`).
- **Attention** — the short-horizon focus dial (`AttentionAnchor`,
  `persona/persona.go`) plus the session-genesis **attention
  keyword ladder**: words and people compiled at login that should
  catch the host's eye in others' chat (`SetKeywordLadder`,
  `runtime/mesa_director.go`; consumed by the reactive trigger in
  `runtime/reactive.go`).

Alongside these sits the **lightweight mood state** —
`limbic.Affect`, a three-axis vector (stress, confidence, valence;
not the calm/anxious/excited/bored word-set the early design
sketched). It is nudged by events on the limbic goroutine, decays
lazily toward a persona baseline (re-based by genesis each
session — `SetAffectBaseline`), and is *not* an in-band subsystem
the host runs: it exists to feed the pearl engine's affect
predicates and the per-turn Situation, so a stressed host decides
and speaks differently.

The consolidated **persona schema** the design promised has landed:
see [`personas.md`](personas.md) and
[`persona-authoring.md`](persona-authoring.md).

## How the layers talk to each other

```
   ┌────────────────────────────────────────────────────┐
   │ 7. Persona + Reveries  (invisible to the host)     │
   │    "who am I, baked into how everything I do        │
   │     gets subtly varied"                             │
   └────────────────────────────────────────────────────┘
                          │  silently shapes
                          ▼
   ┌────────────────────────────────────────────────────┐
   │ 6. Cognition  (memory + reasoning)                  │
   │    "which routine should I run, given what I know" │
   └────────────────────────────────────────────────────┘
                          │  picks + adjusts
                          ▼
   ┌────────────────────────────────────────────────────┐
   │ 5. Behavior Runtime  (routine VM)                   │
   │    "execute this recipe, react to events"           │
   └────────────────────────────────────────────────────┘
              │ triggers                ▲ receives
              ▼                         │
   ┌──────────────────────┐  ┌──────────────────────────┐
   │ 4a. Actions          │  │ 4b. Events               │
   │   (interaction)      │  │   (perception)           │
   │   verbs to the world │  │   typed notifications    │
   └──────────────────────┘  └──────────────────────────┘
                          │
                          ▼
   ┌────────────────────────────────────────────────────┐
   │ 3. World  (the host's mirror of reality)            │
   └────────────────────────────────────────────────────┘
                          ▲
                          │  packets in/out
   ┌────────────────────────────────────────────────────┐
   │ 2. Session  (logged-in connection)                  │
   └────────────────────────────────────────────────────┘
                          ▲
                          │
   ┌────────────────────────────────────────────────────┐
   │ 1. Wire  (RSC packets on a socket)                  │
   └────────────────────────────────────────────────────┘
```

Each upward arrow is "I'm telling you about something that
happened." Each downward arrow is "make this happen" or
"shape what I'm about to ask for." The persona layer at the
top is dashed in spirit — it doesn't issue commands, it
*colors* the answers that come back from below.

## Why this shape, in one paragraph

The reason the host is built in layers and not as a monolithic
script is that we want population-scale behavior to be
*emergent*, not *coded*. We don't write a 500-line if-tree that
makes delores feel different from bernard. We write one set of
routines (layer 5), one cognition stack (layer 6), one wire
protocol (layer 1) — and then we let the persona/reverie layer
(layer 7) silently flavor every action, every wait, every
choice, in a way the routines themselves never see. The result
is that the same recipe produces a different meal in different
kitchens, which is what makes 500 hosts read as a town instead
of as a fleet.

## What this doc deliberately leaves out

- Package names beyond pointers, Go interfaces, wiring. That's
  [`architecture.md`](architecture.md).
- The detailed control loop / turn model. That's also in
  `architecture.md`.
- The DSL surface (syntax, builtins, events, accessors). That's
  the [`lang/`](lang/) subfolder, starting with
  [`lang/README.md`](lang/README.md).
- The mesa server ([`mesa.md`](mesa.md),
  [`mesa/ARCHITECTURE.md`](mesa/ARCHITECTURE.md),
  [`mesa/PROTOCOL.md`](mesa/PROTOCOL.md)), the memory stack
  ([`memory.md`](memory.md)), and the observability taps
  ([`observability.md`](observability.md)). Each has its own doc.
