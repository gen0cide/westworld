# Questions and decisions

> **STATUS: ANNOTATED HISTORICAL LEDGER** (verified against code 2026-06-10, HEAD
> `0bfa818`). The decision text below is the original bootstrap-era record,
> preserved verbatim — this doc answers "why did we do it this way?", not "what
> is built?" (that's [`index.md`](index.md)). Each entry now carries a
> **Status** line graded against the code:
> **HELD** (decision stands, manifested as written) ·
> **HELD-DIFFERENTLY** (intent stands, the mechanism changed) ·
> **REVERSED** (the opposite shipped) ·
> **UNBUILT** (still design, nothing manifested) ·
> **RESOLVED** (a pending question that got answered).
> Open follow-ups are not tracked here — the work queue SSOT is
> [`TODO.md`](TODO.md) (IDs cited inline). A final section records two
> decisions the build adopted silently that were never written down.

## Strategic / research direction

### A1. Research goal priority
**Decision**: Primary: (c) Emergent economy, (d) "Believable population", (e) LLM eval — these three are nearly co-equal. (a) Turing-among-bots is the substrate that enables them but is not the end goal itself. (b) Density/scale study is academic, not the goal.

The real goal: through well-crafted reveries and memory systems, measure long-term strategic objective accomplishment, ideally through organically organized group/community structures. Observe morality and ethics under no-external-observer conditions.

**Status (2026-06-10): HELD** — the standing research frame; the live statement
is [`research-goals.md`](research-goals.md). The long-horizon measurement
phases themselves are [`TODO.md`](TODO.md) R-3.

### A2. Deception fidelity bar
**Decision**: Bots fully believe everyone else is human and act accordingly. The world is populated almost entirely by bots; humans are observation-only. Bots must not treat each other differently than they would treat humans — and to do that, they need to *actually believe* everyone is human. Casual to sustained believability; adversarial probing is out of scope.

**Status (2026-06-10): HELD** — the belief is enforced at the prompt layer:
mesa's Act prompt presents nearby players as "real people"
(`mesa/mesad/act.go`, the 👥 NEARBY PLAYERS block). Adversarial probing remains
out of scope.

## Architecture & project setup

### B1. Phase 0 green-light
**Decision**: Yes. Bootstrap `github.com/gen0cide/westworld`, start wire protocol work.

**Status (2026-06-10): HELD** — executed; the repo is the proof.

### B2. Server target
**Decision**: We run the OpenRSC server ourselves at `~/Code/openrsc/` with the `westworld.conf` config (this repo's `inc/westworld.conf`). Localhost-only for v1. Modifications include lifting all anti-abuse limits while preserving authentic content settings.

**Status (2026-06-10): HELD** — `inc/westworld.conf` is live; ops detail in
[`server-config.md`](server-config.md). One drift inside that conf: the F2P
posture reversed (see [F2P out of gate](#f2p-out-of-gate) below). Re-enabling
public-exposure limits for production is [`TODO.md`](TODO.md) O-8.

### B3. Bot starting state
**Decision**: Fresh accounts, level 1 in everything. Hosts start as new players, just like in Westworld where hosts are commissioned into roles and have to "learn" their narratives. Preservation XP rates are brutal — this forces social cooperation, mentor relationships, and meaningful long-term goals.

**Status (2026-06-10): HELD** — authentic rates verified in
`inc/westworld.conf` (`combat_exp_rate: 1`, `skilling_exp_rate: 1`,
`want_fatigue: true`); hosts start fresh (Delores ran tutorial island from
level 1).

## DSL & scripting

### C1. AR scripting model
**Status**: Alex providing examples. DSL design pends his input. Reference archive: OpenRSC APOS Java scripts (https://gitlab.com/openrsc/APOS) as protocol reference, NOT as behavioral templates — those scripts are drone-style and lack social/chat awareness.

**Status (2026-06-10): RESOLVED** — the pend closed long ago: the DSL is
designed and built (`dsl/` — parser, ast, validator, interp; spec in
[`lang/`](lang/README.md)). APOS served its protocol-reference purpose.

### F1. DSL hosting choice
**Decision**: Custom interpreter, AR-inspired. Reasons:
- Token efficiency: simpler DSL = smaller prompts when the brain generates routines
- We can inject `reverie.tick()` at call sites natively
- Total control over sandbox, timeouts, event handling
- Mirrors what felt right to Alex from his AR experience

**Status (2026-06-10): HELD** — the custom interpreter shipped (`dsl/interp`),
with the sandbox/timeout/event control as promised (suspend controller,
watchers, validator gate) and the token-efficiency bet paying off via the
cached DSL-manual prompt prefix (`mesa/mesad/dslmanual.go`, `act.go`). One
rationale bullet aged: the injection point manifested as the `OnAction`
action-boundary seam (`dsl/interp/interp.go`, fired in `events.go`), and the
Phase-5 decision record supersedes the `reverie.tick()` model itself — see
A2' below.

### E3. Routine promotion (was "skill promotion")
**Decision** (revised): **No bootstrap routines.** Hosts learn organically — handing them naive starters biases their development toward our scaffolding rather than their own emergent flow. Routines are private by default. Each agent develops its own.

Inter-host learning is deferred to a later phase: "drone teaches drone" (observation/mimicry/blog pattern). Mimicry can be suboptimal — hosts should continually evaluate alternatives, not lock in the first thing they see.

The hooks in the DSL (where reveries inject, how observers tap, etc.) must be well-defined before routines start being filled in. This relates to the persona/reveries LOE.

**Status (2026-06-10): HELD** — no starter routines were ever shipped. The
routine library that exists is *learned*: `runtime/library.go` promotes a
successful LLM-authored move into a per-host, bbolt-persisted cache keyed by
situation signature (the cheap loop replays it LLM-free). It is private per
host, exactly as decided. Inter-host learning remains deferred —
[`TODO.md`](TODO.md) R-3.

## Cognition / memory

### D2. Memory backend
**Decision**: Mesa service with Postgres + pgvector. Shared across all hosts. Per-host scoping enforced at the API layer (no cross-host memory leakage). Knowledge (wiki RAG) and shared skills live here too.

**Status (2026-06-10): HELD** (first two sentences) — `mesa/mesad/ltm.go`
migrates Postgres tables for episodes (pgvector HNSW cosine), observations,
relationships, knowledge, goal_graphs, personas, kv, goals, metrics; per-host
scoping is enforced at the API layer (`mesa/mesad/auth.go` — the gRPC
interceptor binds bearer-token → host_id and never trusts a host_id from the
request body). **HELD-DIFFERENTLY** on the tail: wiki RAG became out-of-band
offline tooling (`mesa/wikirag`), not a mesa runtime surface, and "shared
skills" never landed mesa-side (the routine library is per-host local — see
E3).

### D3. Embedding model
**Decision**: Voyage 3. Alex has an existing account/key. Cheap, high-quality, no need to run local ONNX yet.

**Status (2026-06-10): HELD** — `mesa/embed/voyage.go` (`DefaultModel =
"voyage-3"`, 1024-dim), wired in `mesa/cmd/mesad/main.go` via
`VOYAGE_AI_KEY`; episodes embed on write, Recall ranks by pgvector cosine
(`mesa/mesad/ltm.go`). No local ONNX, no CGo.

### E5. Wiki ingestion timing
**Decision**: Phase 3.5. Build and test it as a cohort experiment: spawn wiki-RAG-enabled hosts alongside no-wiki hosts and compare. If RAG measurably improves agent performance/believability, it's a default for future cohorts.

Wiki content will be F2P-filtered at ingest (most agents are F2P; P2P content would pollute the knowledge pool).

**Status (2026-06-10): UNBUILT** — the cohort experiment never ran
([`TODO.md`](TODO.md) R-1). Substrate exists (persona `CohortID`,
`mesa/wikirag` extraction + Voyage search tooling), but the delivery shape
changed: hosts are slated to get *curated* knowledge via the persona-compile
worldbrief→cook pipeline ([`mesa/knowledge-pipeline.md`](mesa/knowledge-pipeline.md)),
not live wiki RAG at runtime. The F2P-filter rationale is moot while the world
runs P2P (see the F2P entry — REVERSED).

### D6. Always include nearby player memory in prompts
**Decision**: Yes, always include relational records for nearby players in strategist prompt context. Trickier in dense areas (Varrock with 30 nearby players) — need to summarize or truncate to highest-relevance subset.

**Status (2026-06-10): HELD** — `runtime/mesa_director.go nearbyPlayersHint`
lists players within radius 16, each tagged with `trustNote` (trust grade,
familiarity, notes from the limbic ledger); `mesa/mesad/act.go` renders it as
the 👥 block with relationship-weighing instructions. The dense-area
truncation concern is still open beyond the radius cap — it gets exercised by
the fleet soak ([`TODO.md`](TODO.md) O-5).

### E1. Revisable reflections
**Decision**: Yes, must be revisable. The `reflections` table has a `last_revised` column. Hosts update earlier opinions when evidence accumulates.

**Status (2026-06-10): HELD-DIFFERENTLY** — no `reflections` table or
`last_revised` column exists. Revisability manifested as Beta-mean belief
confidence in the knowledge ledger (`cognition/knowledge` — restating a claim
accumulates α/β evidence) plus the Tier-2 insight cron's contradiction
reconciliation, which demotes the losing belief once, idempotently
(`mesa/mesad/cron_insight.go`).

### E2. Cross-bot reputation propagation
**Decision**: Bots communicate via in-game comms only (chat, PMs). They do NOT have shared mesa-side reputation databases.

If BotA tells BotB "JimBob scammed me," BotB can choose to believe, doubt, or investigate. Reputation propagation is contextual — BotA could be lying, in on the scam, or have a grudge. The brain decides how much weight to give the claim based on relationship context.

**Status (2026-06-10): HELD** — relationships are AuthLocal per host (the
`relationships` table in `mesa/mesad/ltm.go` is a per-host mirror; there is no
cross-host read path). Heard claims arrive via the reactive ExtractDialog
tier with graded confidence and a provenance the host *overrides* from the
speaker's role — a player can't claim system authority
(`mesa/proto/mesa.proto` `DialogClaim`).

## Brain / LLM strategy

### D1. Brain interface shape
**Decision**: Define a `Brain` interface from day one, with `AnthropicBrain` as the single concrete implementation. The interface costs almost nothing (one method signature) and enables: mock brains for testing, easy A/B testing against other providers later, clean swap to local models if cost demands.

**Status (2026-06-10): HELD-DIFFERENTLY** — the interface shipped
(`brain.Strategist`, default `brain.StubStrategist` at
`runtime/host.go:457`), and it paid for itself exactly as argued. But
`AnthropicBrain` never existed: the production implementation is
`mesaclient.AsStrategist` (`mesa/client/adapters.go`), which routes `Decide`
over mesa gRPC. The Anthropic client lives mesa-side only
(`mesa/llm/anthropic.go`) — a host never talks to Anthropic directly.

### D2'. Tiered model routing
**Decision**: Yes. Inside `AnthropicBrain`, route by decision class:
- Sonnet 4.6: strategic planning, script generation, persona-consistency audits
- Haiku 4.5: routine "what next" decisions, chat replies, reactive emergencies, importance scoring
- The router itself is pure Go logic (no LLM needed to choose which LLM to call)

**Status (2026-06-10): HELD-DIFFERENTLY** — the tiering and the pure-Go
router both held; the *home* moved server-side into mesad (there is no
AnthropicBrain — see D1). Three tiers now: `actLLM` (claude-sonnet-4-6, Act /
DSL authoring), `decideLLM` (claude-haiku-4-5, narrow Decide / Chat /
reactive ExtractDialog), and a genesis tier (claude-opus-4-8, rare login
compile) — the model flags in `mesa/cmd/mesad/main.go`. Escalation predicates
are
deterministic Go (e.g. `extractWantsNuance`, `mesa/mesad/act.go`).

### D4. Importance scoring
**Decision**: LLM-first (Haiku per event) initially. Each scored event is labeled data; later we can extract heuristic rules from the labeled corpus if cost demands. Bias toward LLM until cost proves prohibitive.

**Status (2026-06-10): HELD-DIFFERENTLY** — the cost stance held, the
per-event granularity didn't: the consolidation cron batches ~50 observations
per Haiku call (Tier-1, `mesa/mesad/cron.go DefaultCronConfig`), with Tier-0
no-LLM gates (novelty/dedup, salience decay) in front and a rare Tier-2
Sonnet insight pass behind a pre-filtered escalation queue
(`mesa/mesad/cron_insight.go`). Lossy-bus handling of the very highest-
importance events is [`TODO.md`](TODO.md) C-23.

### D5. Reflection cadence
**Decision**: Agent-decided, with a hard cap of 30 minutes between reflections. The agent can request a reflection sooner if significant events occur; mesa enforces the upper bound to prevent runaway costs.

**Status (2026-06-10): HELD-DIFFERENTLY** — "reflection" manifested as
operator-tuned mesa crons, not agent-requested: consolidation ticks every 60s
and the insight loop at 3× that (`mesa/mesad/cron.go`), concurrency-capped so
they can't starve Act/Chat/Decide. The agent does not request reflections;
the 30-minute-cap idea survives only as the general bounded-cost stance.

## Believability / reveries

### F2. Reverie defaults
**Decision**: Critical to get right. Default-on once stable; opt-in during early development to keep debugging clean. Reveries are NOT polish — they're load-bearing for the research goal. Bake them in from Phase 3.

**Status (2026-06-10): UNBUILT** (the engine) — and the timeline slipped:
reveries are now the Phase-5 build ([`TODO.md`](TODO.md) P-6). The
load-bearing framing held and its substrate exists: limbic mood (`limbic/` +
the `runLimbic` goroutine in `runtime/host.go`), per-host `ReverieSeed`
jitter (`persona/persona.go`), and pearl `EffectInject` reverie/quirk-tic
intents (`pearl/engine.go Injections`) — but nothing consumes the intents yet
(`Injections` has zero callers outside the package).

### A2'. Reveries as core (not polish)
**Decision**: The `reverie.tick()` call between every atomic action in the DSL interpreter is mandatory. Reveries can also be invoked explicitly from within a routine. The reverie module reads the host's emotional state (persona traits + recent context + idle time) and dispatches augmentations: idle wander, glance, spontaneous chat, mistake injection, persona-quirk fire.

**Status (2026-06-10): REVERSED (mechanism) + UNBUILT** — the mandatory
`reverie.tick()` interpreter hook was explicitly refuted by the Phase-5
decision record in favor of a sibling-goroutine *scheduler* driving public
effector methods behind a busy-flag ([`TODO.md`](TODO.md) P-6, with the
ActionArbiter prerequisite P-9). What exists today is the interpreter's
`OnAction` action-boundary seam (`dsl/interp/interp.go`, fired at every
action boundary in `events.go`) — built for exactly this job, wired by
nothing in `runtime/` yet. The "reads emotional state, dispatches
augmentations" intent stands as the live spec ([`reveries.md`](reveries.md)).

### E4. Persona generation
**Decision**: Hybrid — LLM-generated variation atop hand-authored templates. Major paired design session required (see [personas.md](personas.md)). Topics:
- Cohorts (clusters of agents sharing traits)
- Persona templates (5-10 archetypes)
- North stars (long-term motivations)
- Reverie design (the augmentations themselves)
- LLM-generation pipeline for per-agent variation

Saved as a pending LOE for paired work before Phase 3 begins.

**Status (2026-06-10): HELD** — the hybrid shipped: typed persona schema +
enum SSOT + `Validate()`/`CompilePolicy` (`persona/`), with the LLM pipeline
as the Opus best-of-N cook + judge (`mesa/personacook`). One reframe since:
**cohort = launch batch** (analytics grouping) and **archetype =
personality** (`persona/persona.go` `CohortID`/`Archetype`). The paired
session is still partially owed — open decisions and the authoring tables
(including the reverie catalog) are [`TODO.md`](TODO.md) P-1/P-2.

### F3. Initial swarm scale targets
**Decision**: Phased rollouts — 10 → 100 → 500. Each scale validates assumptions before the next.

**Status (2026-06-10): HELD (plan), not yet executed** — the fleet substrate
exists (`cmd/cradle-server` runs many hosts in one process); the formal
10/100/500 waves are [`TODO.md`](TODO.md) R-2, with the 200-drone soak (O-5)
as the standing pre-milestone.

## Architectural confirmations

### G1. Mesa as separate binary + library mock for dev
**Decision**: Both. `cmd/mesa` for production. `mesa/client` package has a `MockClient` implementation for local single-host dev (Phase 0-2 don't need mesa running).

**Status (2026-06-10): HELD-DIFFERENTLY** — separate binary yes, but it is
`mesa/cmd/mesad` (binary `mesad`), not `cmd/mesa`. `MockClient` never
existed; the offline seam is `mesa/client.StubClient` (the always-offline
default — Act/Decide return `ErrOffline`, Recall empty, Remember no-op;
`mesa/client/client.go`) plus `brain.StubStrategist`, and a host runs fully
self-contained on local routines + pearl without mesa, as intended.

### G2. Process model for delos swarm
**Decision**: Many hosts per process. Shared HTTP connection pools to mesa, shared Anthropic rate limiter, shared embedding cache. Standalone single-host mode is also supported for testing.

**Status (2026-06-10): HELD-DIFFERENTLY** — "many hosts per process" shipped
as the cradle daemon (`cmd/cradle-server` + `cradle/` — facts + landscape
loaded once, shared by pointer), and standalone single-host mode is
`cmd/host`. The sharing details inverted, though: mesa connections are
deliberately **one gRPC conn per host, no pooling** (`cradle/deps.go
DialMesa`); there is no Anthropic rate limiter anywhere host-side because
hosts never call Anthropic (see D1) — fleet-level LLM throttling in mesad is
open ([`TODO.md`](TODO.md) O-10); and there is no host-side embedding cache
because embedding is mesa-side only (D3).

## Communications architecture

### Async comms (Alex's late observation, baked in)
**Decision**: Chat is a separate channel from primary actions. Real RSC physics: you can't click-target and type at the same time. So:
- **Primary channel** = atomic actions (one at a time, each takes some duration)
- **Secondary channel** = chat queue, drained opportunistically between primary actions

The brain queues "I want to say X." The chat worker decides WHEN to send: between routine action transitions, while idle at a bank, during walks (one short message at a time), never mid-combat or mid-bank-interaction. Produces naturally fragmented, human-feeling chat.

**Status (2026-06-10): HELD-DIFFERENTLY** — the two-channel separation
shipped, but as a *reactive social goroutine*, not an outbox: incoming player
chat is handled off the Act loop in `runtime/runhost.go` via the
cheap mesa `Chat` RPC (Haiku tier, `mesa/mesad/act.go`) + `host.Say`,
rate-limited and with directed-reply routing; `Ask` is the proactive twin
(`mesa/client/client.go`, used by `runtime/speech.go`). No queue-drained-
between-actions worker exists; "never mid-combat" timing discipline would
arrive with the ActionArbiter ([`TODO.md`](TODO.md) P-9).

## Observability

### Technician tablets (delos web UI)
**Decision**: The delos binary serves a web UI in addition to managing the swarm. The UI provides:
- Live state inspection of any host (persona, position, inventory, current routine, recent events)
- Chain-of-thought tap (subscribe to a host's live Brain call stream)
- Cohort analytics (population-level metrics, A/B comparison views)
- Admin controls (pause host, force reflection, inject event, edit persona)

Data plumbing (Brain call ledger in mesa, structured event capture) must be in place from Phase 3 onward. The UI itself is Phase 5.

**Status (2026-06-10): HELD-DIFFERENTLY** — a `delos` binary never existed;
the intent landed across three real surfaces: the **cradle daemon** web UI +
JSON API (`cradle/api.go` — host list/inspect, pause/resume/stop/restart,
analysis mode, and a per-host debug proxy to the debughttp dashboard, whose
`/ws` live thought stream — `agent_thought` events, `debughttp/server.go` —
is the chain-of-thought tap) plus `cmd/cradle-ctl`; the
**mesa Admin plane** (`cmd/mesa-ctl` — persona set, goal push, fleet gen);
and the mesad **metrics** table fed by `Journal.ReportMetrics` (the data
plumbing, in place as demanded). Cohort/archetype analytics remain open —
[`TODO.md`](TODO.md) O-7.

## Server config

### F2P out of gate
**Decision**: `member_world: false` in westworld.conf. Reasons:
- Geographic density: F2P map is ~25-30% of total walkable area; at 500 hosts, social encounters become routine
- Server-enforced confinement via gate locs (verified in OpenRSC source)
- Early-game IS F2P logically; level 1 agents would stay F2P-flavored organically anyway
- Wiki pollution is solvable at ingest time (filter to F2P-relevant content)
- P2P remains available as a future cohort experiment

**Status (2026-06-10): REVERSED** (current posture) — `inc/westworld.conf:46`
reads `member_world: true`, annotated "Pre-launch dev posture; flip to false
when wiring production F2P hosts." The rationale above still governs the
*production* posture; the flip-back (plus gate-loc confinement re-verify) is
[`TODO.md`](TODO.md) O-8.

### Anti-abuse limits removed
**Decision**: All `*_per_ip`, `*_per_second`, flood/suspicion ban timers, registration limits, packet limits set to absurdly high or zero values. Authentic content settings (XP rates, fatigue, location_data, broken-mechanics fixes) preserved verbatim. See `inc/westworld.conf` and [server-config.md](server-config.md).

**Status (2026-06-10): HELD** — verified in `inc/westworld.conf` (per-ip /
per-second limits at 10000/1000, flood + suspicious ban timers 0; authentic
XP/fatigue settings intact). Re-enabling limits for any public exposure is
part of [`TODO.md`](TODO.md) O-8.

## Adopted silently — recorded 2026-06-10

Two infrastructure decisions were made by the build without ever being written
into this ledger. Both trace to the architecture-paper digest
(`_research/reference/digest-claude-arch.md`), whose right-sizing verdicts the
code followed:

### No Redis
**Decision (followed, never recorded)**: No Redis. The hot read path
(persona, goals, working memory) is in-process Go state on each `Host`; the
local-durable tier is bbolt (`memory/` + `hostkv`); Postgres covers everything
durable. The digest's verdict — "Redis collapses Postgres read traffic" only
matters with N separate processes; for a many-hosts-in-one-process cradle,
process memory *is* the L1 cache — is exactly what shipped. Verified: no
Redis dependency in `go.mod` (storage deps are `jackc/pgx/v5` + `go.etcd.io/bbolt`).
Redis would earn reconsideration only as cross-process fan-out if the
single-daemon model ever splits.

### No River job queue
**Decision (followed, never recorded)**: No `riverqueue/river`. The cron
*catalog* was adopted (consolidation, insight, GC/decay) but runs as plain Go
goroutines with `time.Ticker` + a shared concurrency-cap semaphore inside
mesad (`mesa/mesad/cron.go` — `cronSem`, ticker loops), per the digest's
"adopt the job catalog, defer the dependency" verdict.

**Caveat — one digest verdict was reversed**: the same paper advised deferring
gRPC until mesa split to its own binary. Mesa shipped *as* a gRPC service
anyway (`mesa/proto/mesa.proto`, bearer-token auth interceptor in
`mesa/mesad/auth.go`), so the transport half of that staging advice is dead;
only the no-Redis/no-River halves survive as live decisions.
