# Westworld

> "These violent delights have violent ends."

A Go monorepo for an LLM-driven RuneScape Classic bot population. The goal: build a sandboxed world where hundreds of LLM agents play the game organically, each believing themselves the only AI in a society of humans, and observe what emerges — long-term strategic accomplishment, social fabric, group formation, and the morality and ethics of decisions made when no one is watching.

This is not a botting framework in the conventional sense. The agents are not optimized to grind XP or farm gp. They are configured to be *believable players* — flawed, social, curious, persistent — and we study what kind of society they form.

## What's here

| Binary | Role | Westworld analogue |
|---|---|---|
| `cmd/cradle` | The per-host runtime. One process = one bot. | The simulation where hosts iterate |
| `cmd/mesa` | The central memory + RAG service. Postgres + pgvector. | Corporate HQ where the systems live |
| `cmd/delos` | The swarm orchestrator + observability UI (the "technician tablets"). | The corporation that runs the park |

Each host's brain is structured in layers: a strategist (Claude Sonnet for novel decisions, Haiku for routine ones) drives a deterministic DSL interpreter that executes routines. Reveries — small unconscious behavioral augmentations like timing jitter, idle wandering, and persona-driven chat — are injected at every action call site, making routine execution indistinguishable from organic play.

Memory is fungible. Episodes decay or compress unless reinforced. Relational records (per-player social facts) accumulate naturally as hosts trade, fight, or chat. Reflections (higher-order insights derived from recent events) shape long-term strategic behavior. All of this lives in `mesa`, scoped per-host but accessible to admin queries for research analysis.

## Why this exists

The questions we want to answer:

1. Can LLM agents, given the right cognitive architecture, accomplish *long-term* strategic objectives — not minute-to-minute optimization but month-scale goals?
2. Do organically organized communities and groupings emerge from a population of independent agents, or does atomized individual play remain the norm?
3. When agents believe everyone around them is human and act accordingly, what do their *ethics* look like? Do they steal? Help strangers? Form factions? Hold grudges?
4. What persona / cognitive design produces *the most believable host* — one that other hosts treat as fully human, sustained over weeks of interaction?

We're using RuneScape Classic specifically because it's small, observable, has rich social affordances (chat, trading, PvP, PvE cooperation), and runs on a server we control end-to-end.

## Architecture overview

```
                   ┌──────────────────────────────┐
                   │  mesa (shared service)        │
                   │  Postgres + pgvector + Voyage │
                   └──────────────┬───────────────┘
                                  │ HTTP API
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
   ┌────┴─────┐              ┌────┴─────┐              ┌────┴─────┐
   │ cradle #1│              │ cradle #2│   ...        │ cradle #N│
   │ (Go bot) │              │ (Go bot) │              │ (Go bot) │
   └────┬─────┘              └────┬─────┘              └────┬─────┘
        │ RSC wire protocol       │                         │
        └─────────────────────────┴─────────────────────────┘
                                  │
                          ┌───────┴────────┐
                          │ OpenRSC server │
                          │ (westworld.conf)│
                          └────────────────┘
```

Detailed architecture: [`docs/architecture.md`](docs/architecture.md)

## Deep dives

Every architectural decision, design topic, and research consideration is documented in `docs/`:

- [research-goals.md](docs/research-goals.md) — the questions this project tries to answer
- [architecture.md](docs/architecture.md) — full layer cake, package map, control flow
- [phases.md](docs/phases.md) — phase-by-phase build plan
- [questions-and-decisions.md](docs/questions-and-decisions.md) — every design decision with rationale
- [protocol.md](docs/protocol.md) — RSC wire protocol notes (mc234/235 / Payload235)
- [mesa.md](docs/mesa.md) — memory + RAG service design
- [brain.md](docs/brain.md) — LLM strategist, tiered model routing, prompt design
- [cognition.md](docs/cognition.md) — RAG strategy, retrieval, knowledge sources
- [memory.md](docs/memory.md) — episodic / relational / reflective / working memory
- [reveries.md](docs/reveries.md) — the believability layer
- [dsl.md](docs/dsl.md) — scripting language design (AutoRune-inspired)
- [personas.md](docs/personas.md) — cohort / persona / north-star system (deferred design session)
- [observability.md](docs/observability.md) — delos UI ("technician tablets"), chain-of-thought capture
- [server-config.md](docs/server-config.md) — westworld.conf rationale

## Status

Bootstrap phase. No Go code yet — architecture and docs are being committed first. See [phases.md](docs/phases.md) for the build plan.

## Running the OpenRSC server with the westworld config

The server itself is OpenRSC at `~/Code/openrsc/`. The westworld config lives here at `inc/westworld.conf`. To run the OpenRSC server pointed at this config, see [docs/server-config.md](docs/server-config.md).

## License

TBD. This is a research project; license decision deferred until the work is meaningfully usable.
