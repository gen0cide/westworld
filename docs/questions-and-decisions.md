# Questions and decisions

A record of every architectural and design question that was resolved during the bootstrap design phase. Each entry has the question, the decision, and a brief rationale. This document is the canonical answer when someone asks "why did we do it this way?"

## Strategic / research direction

### A1. Research goal priority
**Decision**: Primary: (c) Emergent economy, (d) "Believable population", (e) LLM eval — these three are nearly co-equal. (a) Turing-among-bots is the substrate that enables them but is not the end goal itself. (b) Density/scale study is academic, not the goal.

The real goal: through well-crafted reveries and memory systems, measure long-term strategic objective accomplishment, ideally through organically organized group/community structures. Observe morality and ethics under no-external-observer conditions.

### A2. Deception fidelity bar
**Decision**: Bots fully believe everyone else is human and act accordingly. The world is populated almost entirely by bots; humans are observation-only. Bots must not treat each other differently than they would treat humans — and to do that, they need to *actually believe* everyone is human. Casual to sustained believability; adversarial probing is out of scope.

## Architecture & project setup

### B1. Phase 0 green-light
**Decision**: Yes. Bootstrap `github.com/gen0cide/westworld`, start wire protocol work.

### B2. Server target
**Decision**: We run the OpenRSC server ourselves at `~/Code/openrsc/` with the `westworld.conf` config (this repo's `inc/westworld.conf`). Localhost-only for v1. Modifications include lifting all anti-abuse limits while preserving authentic content settings.

### B3. Bot starting state
**Decision**: Fresh accounts, level 1 in everything. Hosts start as new players, just like in Westworld where hosts are commissioned into roles and have to "learn" their narratives. Preservation XP rates are brutal — this forces social cooperation, mentor relationships, and meaningful long-term goals.

## DSL & scripting

### C1. AR scripting model
**Status**: Alex providing examples. DSL design pends his input. Reference archive: OpenRSC APOS Java scripts (https://gitlab.com/openrsc/APOS) as protocol reference, NOT as behavioral templates — those scripts are drone-style and lack social/chat awareness.

### F1. DSL hosting choice
**Decision**: Custom interpreter, AR-inspired. Reasons:
- Token efficiency: simpler DSL = smaller prompts when the brain generates routines
- We can inject `reverie.tick()` at call sites natively
- Total control over sandbox, timeouts, event handling
- Mirrors what felt right to Alex from his AR experience

### E3. Routine promotion (was "skill promotion")
**Decision** (revised): **No bootstrap routines.** Hosts learn organically — handing them naive starters biases their development toward our scaffolding rather than their own emergent flow. Routines are private by default. Each agent develops its own.

Inter-host learning is deferred to a later phase: "drone teaches drone" (observation/mimicry/blog pattern). Mimicry can be suboptimal — hosts should continually evaluate alternatives, not lock in the first thing they see.

The hooks in the DSL (where reveries inject, how observers tap, etc.) must be well-defined before routines start being filled in. This relates to the persona/reveries LOE.

## Cognition / memory

### D2. Memory backend
**Decision**: Mesa service with Postgres + pgvector. Shared across all hosts. Per-host scoping enforced at the API layer (no cross-host memory leakage). Knowledge (wiki RAG) and shared skills live here too.

### D3. Embedding model
**Decision**: Voyage 3. Alex has an existing account/key. Cheap, high-quality, no need to run local ONNX yet.

### E5. Wiki ingestion timing
**Decision**: Phase 3.5. Build and test it as a cohort experiment: spawn wiki-RAG-enabled hosts alongside no-wiki hosts and compare. If RAG measurably improves agent performance/believability, it's a default for future cohorts.

Wiki content will be F2P-filtered at ingest (most agents are F2P; P2P content would pollute the knowledge pool).

### D6. Always include nearby player memory in prompts
**Decision**: Yes, always include relational records for nearby players in strategist prompt context. Trickier in dense areas (Varrock with 30 nearby players) — need to summarize or truncate to highest-relevance subset.

### E1. Revisable reflections
**Decision**: Yes, must be revisable. The `reflections` table has a `last_revised` column. Hosts update earlier opinions when evidence accumulates.

### E2. Cross-bot reputation propagation
**Decision**: Bots communicate via in-game comms only (chat, PMs). They do NOT have shared mesa-side reputation databases.

If BotA tells BotB "JimBob scammed me," BotB can choose to believe, doubt, or investigate. Reputation propagation is contextual — BotA could be lying, in on the scam, or have a grudge. The brain decides how much weight to give the claim based on relationship context.

## Brain / LLM strategy

### D1. Brain interface shape
**Decision**: Define a `Brain` interface from day one, with `AnthropicBrain` as the single concrete implementation. The interface costs almost nothing (one method signature) and enables: mock brains for testing, easy A/B testing against other providers later, clean swap to local models if cost demands.

### D2'. Tiered model routing
**Decision**: Yes. Inside `AnthropicBrain`, route by decision class:
- Sonnet 4.6: strategic planning, script generation, persona-consistency audits
- Haiku 4.5: routine "what next" decisions, chat replies, reactive emergencies, importance scoring
- The router itself is pure Go logic (no LLM needed to choose which LLM to call)

### D4. Importance scoring
**Decision**: LLM-first (Haiku per event) initially. Each scored event is labeled data; later we can extract heuristic rules from the labeled corpus if cost demands. Bias toward LLM until cost proves prohibitive.

### D5. Reflection cadence
**Decision**: Agent-decided, with a hard cap of 30 minutes between reflections. The agent can request a reflection sooner if significant events occur; mesa enforces the upper bound to prevent runaway costs.

## Believability / reveries

### F2. Reverie defaults
**Decision**: Critical to get right. Default-on once stable; opt-in during early development to keep debugging clean. Reveries are NOT polish — they're load-bearing for the research goal. Bake them in from Phase 3.

### A2'. Reveries as core (not polish)
**Decision**: The `reverie.tick()` call between every atomic action in the DSL interpreter is mandatory. Reveries can also be invoked explicitly from within a routine. The reverie module reads the host's emotional state (persona traits + recent context + idle time) and dispatches augmentations: idle wander, glance, spontaneous chat, mistake injection, persona-quirk fire.

### E4. Persona generation
**Decision**: Hybrid — LLM-generated variation atop hand-authored templates. Major paired design session required (see [personas.md](personas.md)). Topics:
- Cohorts (clusters of agents sharing traits)
- Persona templates (5-10 archetypes)
- North stars (long-term motivations)
- Reverie design (the augmentations themselves)
- LLM-generation pipeline for per-agent variation

Saved as a pending LOE for paired work before Phase 3 begins.

### F3. Initial swarm scale targets
**Decision**: Phased rollouts — 10 → 100 → 500. Each scale validates assumptions before the next.

## Architectural confirmations

### G1. Mesa as separate binary + library mock for dev
**Decision**: Both. `cmd/mesa` for production. `mesa/client` package has a `MockClient` implementation for local single-host dev (Phase 0-2 don't need mesa running).

### G2. Process model for delos swarm
**Decision**: Many hosts per process. Shared HTTP connection pools to mesa, shared Anthropic rate limiter, shared embedding cache. Standalone single-host mode is also supported for testing.

## Communications architecture

### Async comms (Alex's late observation, baked in)
**Decision**: Chat is a separate channel from primary actions. Real RSC physics: you can't click-target and type at the same time. So:
- **Primary channel** = atomic actions (one at a time, each takes some duration)
- **Secondary channel** = chat queue, drained opportunistically between primary actions

The brain queues "I want to say X." The chat worker decides WHEN to send: between routine action transitions, while idle at a bank, during walks (one short message at a time), never mid-combat or mid-bank-interaction. Produces naturally fragmented, human-feeling chat.

## Observability

### Technician tablets (delos web UI)
**Decision**: The delos binary serves a web UI in addition to managing the swarm. The UI provides:
- Live state inspection of any host (persona, position, inventory, current routine, recent events)
- Chain-of-thought tap (subscribe to a host's live Brain call stream)
- Cohort analytics (population-level metrics, A/B comparison views)
- Admin controls (pause host, force reflection, inject event, edit persona)

Data plumbing (Brain call ledger in mesa, structured event capture) must be in place from Phase 3 onward. The UI itself is Phase 5.

## Server config

### F2P out of gate
**Decision**: `member_world: false` in westworld.conf. Reasons:
- Geographic density: F2P map is ~25-30% of total walkable area; at 500 hosts, social encounters become routine
- Server-enforced confinement via gate locs (verified in OpenRSC source)
- Early-game IS F2P logically; level 1 agents would stay F2P-flavored organically anyway
- Wiki pollution is solvable at ingest time (filter to F2P-relevant content)
- P2P remains available as a future cohort experiment

### Anti-abuse limits removed
**Decision**: All `*_per_ip`, `*_per_second`, flood/suspicion ban timers, registration limits, packet limits set to absurdly high or zero values. Authentic content settings (XP rates, fatigue, location_data, broken-mechanics fixes) preserved verbatim. See `inc/westworld.conf` and [server-config.md](server-config.md).
