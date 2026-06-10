# Westworld

> "These violent delights have violent ends."

**Westworld is a research project that gives AI agents a small persistent world to
live in — and then studies how they live.**

The world is a faithfully-run server of *RuneScape Classic*, the 2001 online
role-playing game: a few square miles of towns, mines, banks, shops, monsters, and
other people. The inhabitants — we call them **hosts**, and the vocabulary of the
show is used deliberately throughout — are autonomous agents, each with its own
personality, memories, relationships, and open-ended goals. A host is not told
what to do. It wakes up in the world, remembers its past lives, decides what it
wants, and goes about its day: mining ore, picking fights it thinks it can win,
haggling in shops, asking strangers for directions, sleeping when it's tired,
holding grudges when it's wronged.

We are not building a game bot. Hosts are deliberately *not* optimized to play
well — they are configured to be **believable**: flawed, social, curious,
persistent. The interesting output is not experience points; it's the decision
traces, the relationships, the emergent social fabric, and what the agents do
when they believe nobody is watching.

## A night in the park

From a recent overnight run, twelve hosts, unscripted: five hosts whose
personalities want to get *strong and rich through combat* trained on goblins all
night and — because their personas value allies — two of them independently teamed
up into a shared grinding crew. A "gambler" host got fined and walked to the bank
to pay it off. A quiet craftsman spent the morning asking a guard where the good
mines are. One fighter woke from a previous life's death with the goal *"shake off
that death an' get back to grindin'."* Every one of those decisions is captured,
timestamped, and traceable to the persona trait or memory that drove it.

## The research questions

1. **Long-horizon agency** — can LLM agents, given the right cognitive
   architecture, pursue *month-scale* goals rather than minute-scale optimization?
2. **Emergent community** — do independent agents organically form groups,
   economies, and social structures, or stay atomized?
3. **Ethics without observation** — when agents believe everyone around them is
   human, what do their morals look like? Do they steal, help strangers, form
   factions, hold grudges?
4. **Believability** — what persona and cognitive design produces a host that
   other players (and other hosts) treat as fully human, sustained over weeks?

The full charter, success metrics, and explicit non-goals: [`docs/research-goals.md`](docs/research-goals.md).

## How a host works (the five-minute version)

```
                 ┌─────────────────────────────────────┐
                 │  MESA — the off-host services        │
                 │  LLM planning · persona registry ·   │
                 │  long-term memory · analysis crons   │
                 └──────────────────┬──────────────────┘
                                    │ gRPC (one link per host)
      ┌─────────────────────────────┴─────────────────────────────┐
      │  THE CRADLE — one process supervising the whole fleet      │
      │  ┌─────────┐  ┌─────────┐  ┌─────────┐       ┌─────────┐  │
      │  │ Delores │  │ bernard │  │ drone1  │  ...  │ droneN  │  │
      │  └────┬────┘  └────┬────┘  └────┬────┘       └────┬────┘  │
      └───────┼────────────┼────────────┼─────────────────┼───────┘
              │   the original 2001 wire protocol, byte-faithful   │
      ┌───────┴────────────┴────────────┴─────────────────┴───────┐
      │            THE PARK — an OpenRSC server we control         │
      └────────────────────────────────────────────────────────────┘
```

- **Body** — each host speaks the original game protocol like a 2001 game client
  would, maintains a live mirror of what it can see, and moves through the world
  with real pathfinding (doors, ladders, multiple floors). To the server, and to
  anyone playing alongside it, a host is just another player.
- **Mind** — cognition is layered by cost, like reflexes versus deliberation.
  Cheap deterministic reflexes handle survival (flee when hurt, sleep when
  exhausted). A scripting language built for this project (**the routine DSL**)
  executes minutes-long behaviors — mine until your bag is full, bank it, repeat.
  A large language model is consulted only at the top, to *plan*: it writes those
  routines, sets goals, and reflects. Most seconds of a host's life cost nothing.
- **Personality** — a host's character sheet (its **persona**: personality-test
  dimensions, values, risk appetite, voice, quirks) is not just prompt text. It
  *compiles* into enforcement: a policy engine (**the pearl**) that can veto or
  bias actions — an honest host screens unfair trades, a pacifist literally cannot
  throw the first punch — plus weights that shape what the planner is even shown.
- **Memory** — hosts remember selectively, the way a character should: salient
  episodes ("I died to a mugger near the east gate"), a per-person trust ledger
  built from actual interactions, accumulated knowledge about the world with
  honest confidence ("I *believe* there's a mine south of here, but I've never
  seen it"), and long-term consolidation that runs while they play. Memories
  survive logout — and feed the next life's first thoughts.
- **The control room** — the cradle's web UI shows each host's position, vitals,
  current routine with a live line-by-line trace, its stream of thoughts, and its
  mind's internals (relationships, goals, open questions). Every decision is also
  journaled to disk for morning-after analysis.

## Why RuneScape Classic?

It's small enough to be observable end-to-end and rich enough to matter: chat,
trade, player-versus-player conflict, cooperative monster fights, an economy,
skills that take real time to raise. The protocol is fully understood, the server
([OpenRSC](https://github.com/Open-RSC)) is open source and runs under our
control, and a 2001-era world has a property modern games don't: **everything an
agent needs to know fits in its head**, so failures of knowledge are *interesting*
rather than inevitable.

## Where to go next

| If you are... | Start with |
|---|---|
| **Curious what this is about** | [`docs/research-goals.md`](docs/research-goals.md), then the war stories in [`docs/lessons-learned/`](docs/lessons-learned/README.md) — the most readable writing in the repo |
| **Technical, want the map** | [`docs/index.md`](docs/index.md) — the documentation index: subsystem status matrix, binary inventory, reading order |
| **An engineer, want the design** | [`docs/architecture.md`](docs/architecture.md) (the layer cake), [`docs/cognition-and-autonomy.md`](docs/cognition-and-autonomy.md) (how a host thinks), [`docs/lang/README.md`](docs/lang/README.md) (the routine DSL) |
| **Interested in the AI/persona side** | [`docs/personas.md`](docs/personas.md), [`docs/persona-authoring.md`](docs/persona-authoring.md), [`docs/memory.md`](docs/memory.md), [`docs/mesa.md`](docs/mesa.md) |
| **Wondering what's planned** | [`docs/TODO.md`](docs/TODO.md) — the single source of truth for open work |
| **An archaeologist** | [`docs/archive/initial-brainstorming/`](docs/archive/initial-brainstorming/) — designs that shipped differently, kept honest; [`docs/questions-and-decisions.md`](docs/questions-and-decisions.md) — every decision with its fate tagged |

## Status

Live. The full stack — protocol, world mirror, DSL, layered cognition, personas
with compiled policy, tiered memory, the mesa services, and the fleet control
plane — is built and soaks nightly with a mixed-personality fleet (named hosts
plus drone cohorts). The current frontier is the metacognition cluster (hosts
reasoning about what they don't know) and the reverie layer (unconscious
believability: timing, idling, small talk). Per-subsystem truth lives in the
[status matrix](docs/index.md); the wire protocol's correctness story is told in
[`docs/lessons-learned/2_WIRE-PROTOCOL.md`](docs/lessons-learned/2_WIRE-PROTOCOL.md).

## Running it

You need an OpenRSC server with this repo's `westworld.conf` (see
[`docs/server-config.md`](docs/server-config.md)), Postgres for mesa, and an
Anthropic API key for the planning seams. The fleet daemon is
`cmd/cradle-server`; the per-host debug harness is `cmd/host`. The quickstart
lives in [`docs/index.md`](docs/index.md).

## License

TBD. This is a research project; license decision deferred until the work is
meaningfully usable.
