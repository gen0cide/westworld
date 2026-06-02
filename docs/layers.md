# Layers — the host, explained from the outside in

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

- **The cradle** — the process that hosts the in-band loop. It owns
  the session, pumps frames into the world+events loop, loads the
  persona at startup, schedules the host's think-turns (the planned
  agent-driver loop, §6), and exposes the observability tap
  (`-spectate`, the brain-call ledger).
- **Mesa** — the shared memory + knowledge service and its
  background jobs (importance scoring, reflection, decay/
  compression). The host experiences *"I recall trading with
  JimBob"*; it never experiences the episode write, the embedding,
  or the salience decay underneath.
- **Delos** — swarm orchestration, cohorts, and the technician UI.
  Purely management-plane; never part of any host's experience.

**Reveries and persona (layer 7) are cross-cutting injection, not a
top-of-cake layer in the dependency sense.** The host never invokes
reveries; the management plane applies them across many seams of the
in-band *action* path (timing jitter, path imperfection, banker
selection, async chat, the occasional misclick), weighted by persona
and a lightweight mood state. They *color* the answers the lower
layers return rather than issuing commands of their own. The
governing rule of the whole in-band stack stays simple: **only the
session ultimately touches the wire; every host action flows down
through the runtime.** (Fuller treatment of the in-band/out-of-band
frame and the agent-driver loop is in the `_research` notes.)

## 1. Wire — the body's nerves

The lowest-level surface: bytes on a TCP socket, framed as RSC
packets, encrypted via ISAAC, with RSA at login. Every other
layer sits on top of this. The wire layer is **pure mechanism**
— it doesn't know anything about the game, it just shuttles
opcodes back and forth.

Nothing about cognition or behavior lives here. If you replaced
RSC tomorrow with a different MMO, only this layer would
fundamentally change.

## 2. Session — being logged in

The session layer is "this host is a player named `delores`,
currently connected, authenticated, with an ISAAC keystream and
a heartbeat going." It owns the reconnect logic, the read/write
goroutines, the login dance. Once the session is up, everything
above just sees a typed stream of "messages from the server"
and a typed channel of "things to send."

This is the smallest layer worth naming. It exists because
without it every higher layer would have to deal with login
state and connection drops.

## 3. World — what the host believes is around it

The host keeps a mirror of the world inside itself: where its
body is standing, what's in its inventory, who's nearby, what
ground items are visible, which doors are open and which are
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
  opened, death. These events flow on an internal pub/sub bus;
  any higher layer can subscribe.
- **Interaction** is the outbound side: typed *actions* the
  host can take. Walk to a tile, attack a target, eat food,
  open a door, drop an item, cast a spell, offer a trade. Each
  action is a single atomic verb that turns into the right
  packet on the wire.

These two halves are deliberately separated even though they
sit at the same height. Events are *involuntary* — they happen
to the host whether it wants them or not. Actions are
*voluntary* — the host chooses them. The reason that matters:
the layer above wants to talk about both, but as different
things ("when the host *perceives* X, take *action* Y").

There is a third small thing living at this layer: **reactions**
— low-level pre-cognitive responses wired directly between
events and actions. The host might "flinch" when attacked
(turn to face the attacker, surface the threat upward) without
thinking. Reactions stay narrow on purpose; the moment a
response gets context-dependent or goal-aware, it belongs in
the layer above.

**Planned (Phase 4): a three-tier handler stack.** Today's
narrow reactions are the seed of a planned generalization — a
**handler stack** that keeps reflexes first-class and persistent
across routines instead of dying when a routine ends. Three
ordered tiers, resolved top-down, first-match-wins:

1. **Persona base reflex** — an always-on default the persona
   owns, persisting across every routine. When a host is new this
   is often a raw "punt" up to cognition ("attacked — fight or
   flee?").
2. **Learned specialization** — a context-scoped fast handler the
   host's mind wrote to *replace* that punt in one situation ("in
   this mine, just eat at HP<30, don't ask"). These are earned, not
   given (see §6 on the falling punt rate).
3. **Routine override** — the routine-scoped `on` handlers we have
   today, active only while the routine runs, chained into the
   lower tiers via `extends host` / `super()`.

The stack is arbitrated by an **interrupt-priority ladder** —
which reaction is allowed to preempt which — running roughly:
**survival** (low-HP eat / flee / sleep-captcha — preempts
everything) > **committed/modal** (an open trade, duel, or bank,
*and* active combat — you orient but defer) > **directed-social /
topic** (someone uses your name, or says a topic you're watching —
preempts the grind) > **grind** (the current routine) > **ambient
observe** (passing chatter — attention-gated, never a preemption).
Handlers declare a tier; the runtime does the preempt and
auto-resume, so the host never juggles priority integers.

The planned **chat response is two-phase**, which is what makes
this work without true mid-action preemption. **Phase 1 (orient)**
is an immediate, deterministic reflex with no model call — face the
speaker, emit a tiny ack — fired at the next action boundary
(~1–2s), exactly how a human looks up the instant they're
addressed. **Phase 2 (reply)** is a patient, asynchronous
model-composed reply that lands in the chat queue a beat later. The
instant part is a between-actions reflex; the slow part is paid
off-loop and invisibly. (Full handler-stack, ladder, and
two-phase design in the `_research` notes.)

## 5. Behavior Runtime — the routine VM

This is the layer where the host gains agency. It's a small
virtual machine that runs `.routine` programs — scripts written
in a custom DSL that say things like "while there's a tree
nearby, chop it, then bank when full, then come back." The VM
turns those programs into sequences of actions at layer 4 and
subscribes to events at layer 4 to react.

The DSL is the *language* — `routine`, `on`, `when`, `select`,
`defer`, `try`, `bounds`, `extends`, etc. The VM is the
*interpreter* — a sandboxed AST walker with budgets (op limit,
wall-clock cap, recursion depth, memory) so a routine can't
run away with the host.

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
  a library file. Eventually they can override and chain via
  `super()` once the persona layer below exists to anchor the
  default parent.

This is the layer that *does the thing*. If you imagine the
host as a person, this is "follow this recipe to bake bread."
The recipe is concrete, step-by-step, and can be written down
in advance.

## 6. Cognition — choosing which recipe

The cognition layer is where the host stops following a recipe
and starts deciding which recipe to follow. It has two parts
that, like layer 4, are easier to think about together:

- **Memory** — the host's record of what it has experienced.
  Working memory (the most recent ~50 events, in-process),
  episodic memory (significant past events, stored in mesa),
  relational memory (what this host knows about other hosts
  and players), and reflective memory (compressed summaries of
  long stretches: "I spent yesterday banking in Lumbridge").
  Memory is queryable — the host can ask itself "have I traded
  with this player before?" and get an answer.
- **Reasoning** — the actual decision-making. The host can
  consult its knowledge base (RSC wiki + AutoRune historical
  scripts ingested into mesa), look up its own memory, examine
  the current world state, and decide what to do next.

The reasoning step is intentionally tiered:

- **Routine decisions** (do I have food? am I being attacked?
  where's the nearest banker?) are answered from cached
  retrieval — cheap, fast, deterministic. No LLM call.
- **Tactical decisions** (this player is challenging me to a
  duel, do I accept?) consult a small language model. The host
  is reasoning, but the bar for a confident answer is low.
- **Strategic decisions** (it's morning, my food's running low,
  the nearest bank is far — do I keep training or go home?)
  consult a larger model. These are infrequent — minutes apart,
  not seconds.

This tiering matters because the project goal is 500 hosts
running 24/7. If every "should I eat?" went through a frontier
LLM, the cost would be ruinous. The split keeps the expensive
reasoning rare and the cheap reasoning constant.

Cognition is what makes a host *adapt*. The behavior runtime
runs the same routine the same way every time; cognition
chooses which routine to run, when to stop, and when to start
something different.

**Planned (Phase 4): the agent-driver loop — how cognition actually
gets a turn.** The mechanism behind "cognition chooses which recipe
to run" is that the host's life alternates **act-turns** and
**think-turns**. A routine runs to completion (an act-turn:
thousands of cheap, model-free actions), and at that goal boundary
the cradle gives the host a **think-turn**: cognition assembles a
*decision request* (persona, world snapshot, recent events, relevant
memory and relations, the index of routines the host knows), the
host's mind returns one of a small set of decisions — **run a known
routine**, **write a new routine** then run it, take a single
**direct action** (mostly chat), or **idle** — and the runtime
executes the result, then the loop repeats. One think-turn is one
bounded reasoning call; one act-turn is thousands of actions. That
ratio is the whole economic argument for running hundreds of hosts
at once. (This loop is scheduled by the cradle at boundaries, not
polled on a tick; the runtime view is in `architecture.md`.)

Two forward-pointers for things cognition will lean on, designed but
not yet built:

- **Relational memory is a per-host trust ledger.** What a host
  believes about another player is modeled as a Bayesian Beta
  posterior — a trust *mean* and a *confidence* (evidence mass), so
  "sure he's reliable, 200 trades" reads differently from "weak
  hunch, met once." The prior is seeded from the host's own
  disposition; updates are severity-weighted and graded by a model
  ("was that cooperation or a scam, and how bad?"). It updates only
  on signals that carry a reliable counterparty name (trade, chat,
  duel) and abstains on melee, where the wire gives no attacker.
- **Scripting competence is earned, not given.** A fresh host does
  not start as a scripting god. It begins with a small survival
  vocabulary and a high "punt rate" (how often it reaches up to
  cognition), and earns a wider, progressively-disclosed verb
  vocabulary through successful use — a falling punt curve is the
  research-grade learning signal. The on-ramp is **Tutorial Island**:
  every host plays the game's own tutorial first, learning the
  basics from the game's own in-context instructions with zero
  documentation handed to it.

(Both are pointer-depth here; the trust-ledger math and the
earned-vocabulary mechanism are detailed in the `_research` notes,
and will land in their own docs.)

## 7. Persona & Reveries — who the host is

The top layer is the part the host **doesn't see**. It's the
fingerprint that makes this host *this host* — not a generic
agent, not a copy of every other host. Two pieces:

- **Persona** is configuration. Values, defaults, behavioral
  knobs. "Delores cuts corners when walking, prefers melee,
  speaks tersely, is slow to trust new players." The persona is
  declared once, in a config file, and is treated by everything
  below it as opaque data — neither the routines nor the
  cognition layer can introspect "what is my persona?" beyond
  consulting individual knobs.
- **Reveries** is the runtime mechanism that *injects* the
  persona's influence at every action call site. When the host
  asks "give me a nearby banker," reveries silently picks the
  third-nearest instead of the first; when the host says "wait
  2 seconds," reveries jitters it to 1.8s or 2.3s based on the
  persona's twitchiness; when pathfinding returns three
  equal-cost routes, reveries weights the choice by how much
  the persona hugs walls.

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
script never has to think about. That's how 500 hosts running
the same routine end up feeling like 500 different people.

**Planned (Phase 4): the persona's exploratory temperament.** The
persona will carry, beyond its behavioral knobs, two orthogonal
dispositions that shape *what a host does with its time*:

- **Curiosity** — a *flavored drive* (a weight vector over
  social / spatial / skill / economic / risk) that says what pulls a
  host and how hard. It is the engine of differential learning: a
  socially-curious host drifts toward chatting and following; a
  skill-curious one toward trying new verbs.
- **Attention** — a scalar focus / resist-distraction dial. Low
  attention is easily pulled off-task; high attention hyperfocuses
  and under-explores. It's the threshold any pull (curiosity,
  boredom, an interruption) must clear to actually yank the host off
  what it's doing.

Alongside these sits a **lightweight mood state** (calm / anxious /
excited / bored, plus recent stress and satisfaction) — a small
piece of cradle-side state nudged by events and decaying toward a
persona baseline. It is *not* an in-band subsystem the host runs; it
exists to **weight reveries** (and flavor the host's reasoning) — a
stressed host's reveries get more anxious, its chat curter.

The full consolidated **persona schema** — an immutable-core /
mutable-layers contract, so a host stays *itself* across weeks while
lived experience still advances it — is being designed; the detail
lands in `personas.md` in a later slice. (Quantitative shape and the
curiosity/attention models are in the `_research` notes.)

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

- Package names, file paths, Go interfaces. That's
  [`architecture.md`](architecture.md).
- The detailed control loop / tick model. That's also in
  `architecture.md`.
- The DSL surface (syntax, builtins, events, accessors). That's
  the [`lang/`](lang/) subfolder, starting with
  [`lang/README.md`](lang/README.md).
- The mesa server, the brain's model routing, the
  observability tap. Each has its own doc.
