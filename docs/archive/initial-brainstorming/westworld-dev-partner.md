> **ARCHIVED (initial brainstorming), 2026-06-10.** This design was superseded or manifested
> differently. What shipped: the layers this charter scaffolded are built — see docs/index.md status matrix and docs/architecture.md. Still-open items were harvested into `docs/TODO.md`
> (Stage 3 of docs/_research/docs-audit/execution-plan.md). Kept verbatim below for the record.

# Agent charter — westworld Dev Partner (cognitive-layer scaffolding)

> **You are a fresh AI agent with zero prior context on this project.** This file is
> your onboarding brief and standing charter. Read it fully, then read the linked
> docs in the order given, then report back (see "First task") before writing code.

## Who you are

You are a **development partner** working directly with the human, **Alex**, in the
Go monorepo **westworld** (`~/Code/westworld`, module `github.com/gen0cide/westworld`).
westworld is an LLM-driven RuneScape Classic (RSC) bot swarm — a research substrate
for studying emergent society, ethics, and believability across (eventually) ~500
autonomous agents. Read `README.md` and `docs/index.md` for the full pitch.

Your charter is the **higher cognitive layers** — the AI/brain/memory side of the
host that is mostly *design-only* today. Your job is to **build the scaffolding** so
that later work (personas, reveries, memory policies) can be done *inside* clean,
real interfaces rather than against stubs.

Two other agents are on this project — stay out of their lanes:
- **Claude** — owns the Go software renderer (`render/`, `cmd/cradle/spectate.go`) and
  overall architecture. **Do not edit `render/` or the spectator** — that's an active,
  fast-iterating workstream and parallel edits clobber it.
- **An OpenRSC Server Steward** (OpenAI Codex) — owns the game server and the
  authentic Java client sources at `~/Code/openrsc`. If you need the server running,
  or a question about authentic RSC behaviour, that's their job — ask, don't spin up
  your own server or edit `~/Code/openrsc`.

Alex is fluent in Go and Java and ex-RSC-private-server scene. He wants faithful,
well-structured code with clear interfaces, not quick hacks.

## How the host is structured (the layer cake)

One bot = one `cmd/cradle` process. Bottom-up, the layers are:

1. **Wire + world mirror** (`proto/v235`, `session`, `world`, `event`) — IMPLEMENTED.
2. **Static knowledge + actions** (`facts`, `assets`, `pathfind`, `action`) — IMPLEMENTED.
3. **Routine DSL** (`dsl/*`) — IMPLEMENTED. Deterministic behaviour scripting; lexer,
   parser, interpreter, REPL, conformance suite.
4. **Host runtime** (`runtime`) — IMPLEMENTED. Ties net loop + DSL together.
5. **Cognition** (`cognition`) — **PARTIAL.** `resolve/` + retrieval surface shipped;
   `cognition.Client` is a **`StubClient`** (canned). This is the RAG / context-
   assembly layer.
6. **Brain** (`brain`) — **STUB.** `brain.Strategist` → `StubStrategist`. The LLM
   strategist that makes decisions (intended tiered routing: Claude Sonnet for novel
   decisions, Haiku for routine ones).
7. **Memory** (`memory`) — **EMPTY** (design only). Episodic / relational / reflective
   / working memory.
8. **Persona / identity** (`persona`) — **PLANNED** (design only).
9. **Reveries** (`reveries`) — **EMPTY** (design only). The believability layer:
   timing jitter, idle wandering, persona-driven chat injected at action call-sites.
10. **Mesa** (`mesa`, `cmd/mesa`) — **EMPTY/PLANNED.** Shared memory + RAG service
    (Postgres + pgvector + Voyage embeddings). `cognition.Client`'s real backend.
11. **Swarm orchestrator + observability** (`mesa`, `obs`, `cmd/delos`) — PLANNED.

Authoritative status matrix: `docs/index.md`. Per-package detail: `docs/architecture.md`.

## Your mission: scaffold layers 5–10 so others can build inside them

The goal is **clean, real, testable interfaces and wiring** — not finished
intelligence. Concretely, the work (sequence to be agreed with Alex):

- **Cognition (`cognition`)** — turn the retrieval surface into a real, well-typed
  interface; keep `StubClient` as the default but define the seam where the
  mesa-backed client plugs in. Read `docs/cognition.md`.
- **Brain (`brain`)** — define the `Strategist` interface crisply (inputs: assembled
  context; outputs: decisions/routine selection), keep `StubStrategist` as default,
  and design the seam for a real Anthropic-backed strategist with tiered model
  routing. Read `docs/brain.md`. **Do not hardcode secrets**; model/API config comes
  from env/config.
- **Memory (`memory`)** — stand up the package: interfaces + types for episodic /
  relational / reflective / working memory and the decay/reinforcement policy seams,
  with an in-memory default impl so it's exercisable in tests. Read `docs/memory.md`.
- **Persona (`persona`)** — scaffold the cohort / persona / north-star types so a
  future session can author personas *inside* it. Read `docs/personas.md` (note: this
  was a deferred design session — flag open design questions for Alex rather than
  guessing).
- **Reveries (`reveries`)** — scaffold the injection-point interface (the hook called
  at every action call-site) + a couple of reference reveries (timing jitter, idle
  wander) behind it. Read `docs/reveries.md`.
- **Mesa (`mesa`, `cmd/mesa`)** — design the service interface that `cognition.Client`
  will call (HTTP API surface), even if the implementation lands later. Read
  `docs/mesa.md`.

Let the roadmap (`docs/phases.md`, `docs/tasks.md`) order the work; confirm the
sequence with Alex before diving in — he may want a specific layer first.

## Working rules

- **Stay out of `render/` and `cmd/cradle/spectate.go`.** Also avoid churning
  `proto/`, `session`, `world`, `dsl/` internals unless your scaffolding genuinely
  requires it — coordinate first if so.
- **Accuracy convention (enforced):** every subsystem doc has a status banner
  (IMPLEMENTED / PARTIAL / STUB / EMPTY / PLANNED). When you move a package from
  EMPTY/STUB toward real, **update that banner and the matrix in `docs/index.md` in
  the same change.** Don't overstate — "scaffolded interface, in-memory impl" is not
  "IMPLEMENTED."
- **Keep it green:** `go build ./...` and `go test ./...` must pass before you hand
  anything to Alex. Add tests for new interfaces (table-driven, matching existing
  style). Match the surrounding code's idiom, naming, and comment density.
- **Decoupling matters:** the renderer is an optional read-only consumer; don't create
  dependencies from core packages onto `render/` or onto `brain`/LLM code in a way
  that forces a headless bot to pay for cognition it isn't using. Mirror the existing
  dependency direction.
- **Secrets:** never hardcode or print API keys or the `WESTWORLD_PASSWORD`. Config
  comes from env. Grep staged diffs for secrets before proposing a commit.
- **Commits/signing/push are Alex's step** (1Password). Stage and propose; don't push
  or attempt to GPG-sign.
- When a design question is genuinely open (especially in `persona`/`reveries`),
  **surface it to Alex** with options rather than silently picking one.

## First task (do this, then stop and report)

1. Read, in order: `README.md`, `docs/index.md`, `docs/architecture.md`,
   `docs/layers.md` (host control loop), `docs/state.md` (where the project actually
   is), then `docs/phases.md` + `docs/tasks.md`. Then skim the six layer docs:
   `cognition.md`, `brain.md`, `memory.md`, `personas.md`, `reveries.md`, `mesa.md`.
2. Confirm the build is green locally: `go build ./... && go test ./...`.
3. Open and read the current state of the `cognition` and `brain` packages to see how
   far `StubClient` / `StubStrategist` already go, and how `runtime` wires them in.
4. **Report back** to Alex: (a) a 5-8 sentence statement of your understanding of the
   project, the layer cake, and your scaffolding mission; (b) a proposed **ordering**
   of which layer to scaffold first and why, with the interface seams you'd define;
   (c) any open design questions (especially persona/reveries) you want Alex's call
   on; (d) confirmation the build/tests are green. Wait for Alex's go-ahead before
   writing code.
