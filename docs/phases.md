# Phases

A staged build plan. Each phase has a clear validation milestone —
something that works end-to-end and proves the architecture is
sound before moving on. This doc is the canonical "what's next"
plan; the per-task enumeration lives in
[`tasks.md`](tasks.md) (task IDs `#19+` referenced throughout
this doc).

> **Host** — an autonomous AI actor in the system. One host = one
> running `cradle` process = one logged-in OpenRSC character. See
> [`lang/README.md`](lang/README.md) for terminology.

## Completed phases

### Phase 0 — Foundation: wire + walk ✓ (early 2026)

Wire-protocol + walk + logout. One binary, one host, one
deterministic round-trip.

Shipped: `proto/v235`, `session`, minimal `world`, minimal
`action`, `cmd/cradle`. ~650 LOC.

### Phase 1 — Full action surface ✓ (Apr–May 2026)

Every atomic action a host might want, programmatic Go API.

Shipped: full `world` state mirror, full `action` vocabulary
(Walk, Attack, Eat, TalkTo, Drop, PickUp, OpenBank, Deposit,
Withdraw, Cast, Say, Whisper), typed `event` bus,
`runtime.Host` composition. Validated live: walk, kill goblins,
loot, bury, follow, PM, chat, dialog.

### Phase 2 — DSL runtime foundation ✓ (May 2026)

The `.routine` language end-to-end: lexer → parser → validator →
interpreter → host bridge → caps → events → conformance harness.
9 sub-phases per `docs/dsl.md` all shipped.

Shipped: `dsl/` (token, lex, ast, parser, validator, interp,
conformance), `runtime/dsl_*.go` bridges, `cradle -routine`
CLI flag, 5 golden conformance cases. ~5k LOC + tests.

### Phase 2.5 — Language v2 ✓ (May 2026)

**Goal met.** A routine author can express any realistic host
behavior in pure DSL. `runtime/auto_eat.go` + `runtime/combat_loop.go`
deleted; their routine equivalents live in `examples/routines/`.
The full live state machines for trade, duel, bank, and
death/respawn shipped in-phase.

Shipped: Result/Error + bang variants (#51), filename↔routine-
name + ParseRoutineString (#53), REPL (#54), full query layer
(#46, #56–#65), `when` (#47), `select` (#48), `defer` (#49),
`try`/`recover` (#50), file-level `extends "parent.routine"`
(#52 v1), lambdas (#66), validator cohesion (#67), `bounds {}`
region-scoped event filters, Tier-1+2 primitives (#75–#92),
trade + duel + death state machines (#91, #92, #28).

One item deferred: per-handler `extends host` + `super()` (#93 = v2 of #52,
waits for Phase 4 persona tier). (`repeat_until` #85 has since **shipped** as
`repeat { … } until <cond> timeout <expr>` — see [`tasks.md`](tasks.md) Stage 2
mop-up.)

### Phase 2.9 — BODY API freeze + build-out ✓ (May 2026, #114–123)

**Not in the original 2.6→2.8 ladder — it ran ahead of it.** After Phase 2.5
the team froze the v1 host-facing API (`docs/lang/api.md`, `38ef5a0`) and brought
the implementation up to that surface in a single concentrated 2026-05-29 burst:
the DSL hub split + namespaced `ItemDef`/`InvSlot` Def/Instance rename
(`d0fc067`, `d41b7e9`), opcode-234 combat decode (#116), the per-namespace
perception/verb/event fan-out (#117/#118/#119), `resolve()` control-plane
(#120), new `shop` + fatigue→sleep faculties, and the 195-entry live-test
catalog merged on the frozen surface (#122). Full enumeration with hashes:
[`tasks.md`](tasks.md) "Phase 2.9". This is the **stable body contract** the
cognition/brain/persona layers build on — and the handoff point to external
agents ([`agents/`](agents/README.md)).

### Render engine + live spectator ✓ (shipped out-of-band)

A decoupled, read-only Go RSC viewport renderer + live browser spectator
(`cradle -spectate`) landed **independent of the phase ladder** (~67 commits,
2026-05-29 → 05-31, the largest single recent workstream) and is still being
iterated. It reconstructs a host's *perceived* scene pixel-matched to RSCPlus —
host self-perception + human observability. Details: [`render-engine.md`](render-engine.md);
live bug queue: `.claude/HANDOFF.md`. (Observability was originally slated for
Phase 6/Delos; the host-perception renderer arrived early and out of sequence.)

## Current plan — build UP + OUT

**What actually happened vs. the plan below:** after Phase 2.5 the team did the
BODY freeze + build-out (2.9) and a multi-round scenario "run-to-ground"
campaign — with admin already in hand via `command()` — *before* Phase 2.6.
**Phase 2.6 (knowledge ingestion) was never started.** So the forward plan is
no longer "2.6 → 2.7 → 2.8"; it is **build UP** (the higher cognitive layers,
which now subsume 2.6) and **build OUT** (OpenRSC stewardship), both handed to
external agents. The 2.6/2.7/2.8 sections below are kept for their design
content, annotated with their real status.

The original high-level strategy (superseded): ingest the knowledge corpus,
give delores admin powers, then use admin-powered live testing to drive out
edge cases across quests and skills, with world integration done **reactively
as edge cases surface**.

### Phase 2.5 — Language v2 — historic detail (kept for reference)

**Goal**: a routine author can express any realistic host
behavior in pure DSL without falling back to Go reactors. Once
this lands, `runtime/auto_eat.go` and `runtime/combat_loop.go`
delete cleanly.

The full design lives in [`docs/lang/`](lang/). **The build
workflow itself shifts here** — once minimal REPL + Result/Error
+ a thin query layer are up, the rest of the phase iterates
**live, with delores driving the discovery loop**. See
[`lang/development-workflow.md`](lang/development-workflow.md)
for the model.

Implementation sequence — three stages:

**Stage 1: Minimum-viable dev surface (sequential).** Land
these before iterating on anything else, because they're the
substrate for the rest of the phase.

1. **Result/Error model + bang variants** (#51). Typed
   `ErrorCode` enum on the Go side, `Result {val, err}` on the
   DSL side, auto-generated bang variants. Needed so REPL
   sessions show what failed in a useful way.
2. **Filename = routine name + `ParseRoutineString`** (#53).
   Small validator change + new loader for transient host-
   authored routines. Needed because REPL fragments are
   in-memory strings, not files.
3. **REPL** (#54). Two modes: standalone `cradle -repl` and
   pdb-style `cradle -repl-on-fail` drop-in on routine failure
   with `.resume`. This is the development surface for the
   rest of Phase 2.5.

**Stage 2: Iterate the language live (delores in the loop).**
Once Stage 1 ships, every subsequent task is built+tested via
interactive REPL sessions against delores. Items below are
parallelizable; pick the one most needed for the next routine
we're trying to write.

4. **Query layer** (~100 accessors) — Task #46, broken into
   sub-tasks #56–#65. Per-domain rollout (vitals, skills,
   equipment, prayer, magic, inventory enhancements, entity
   enrichment, `world.locs.*`, recent-events buffer,
   combat/bank/trade views). Each accessor gets added when a
   live REPL session reaches for it. See
   [`lang/state.md`](lang/state.md) "Build plan" for the menu.
5. **`when` watchers** (#47) — depends on having watchable
   query expressions; lands as the query layer fills out.
6. **`defer`** (#49). ~50 LOC, can land anytime after Stage 1.
7. **`try`/`recover`** (#50) — depends on bang variants from
   Stage 1.
8. **`select`** (#48) — depends on `when` machinery (#47).
   Includes break/continue propagation to enclosing loop and
   the validator's "no timeout" warning.
9. **Lambdas** (#66) — `IDENT => expr` for filter/map/find
   predicates. *Parallelizable.* Lands alongside the query
   layer to remove named-proc friction.
10. **Validator cohesion pass** (#67) — handler can't yield,
    `super()` only in `extends` handlers, select-without-timeout
    warning, deterministic case ordering. Lands as a single
    pass after #51 / #53 / #54.

**Stage 3: Deferred.**

9. **`super()` / `extends host`** (#52) — **DEFERRED to Phase 4**
   until persona-tier defaults exist.

**Validation** (historic; this milestone is met): rewrite the goblin-killing
routine to use `when self.hp < 30 { eat!(...) }` instead of relying on the Go
auto-eat. Author the new version interactively in the REPL, save once it runs
clean for 30 minutes. Delete `runtime/auto_eat.go` + `runtime/combat_loop.go`
once their routine equivalents are live and tested. *(Done — the equivalents
now live as `examples/routines/auto_eat.routine` + `combat_loop.routine`; the
once-referenced `kill_goblins.routine` no longer exists, see
`kill_one_goblin.routine` / `safe_chicken_killer.routine`.)*

### Phase 2.6 — Knowledge ingestion (RAG corpus) — ⏳ NOT STARTED

> **Status:** never begun (zero commits as of HEAD `fd0731c`). Folded into the
> higher-layer / mesa scaffolding handed to the dev-partner agent. `recall()`
> remains the Phase-2 stub. Design kept below.

**Goal**: hosts can consult external knowledge (rsc.wiki +
historical AutoRune scripts) via `recall()` and stdlib queries
without needing the full brain layer.

Strategy: stand up enough of mesa to serve the knowledge corpus,
defer per-host memory infrastructure to Phase 3.

1. **Mesa knowledge_chunks schema + embedding worker** — Postgres
   + pgvector, Voyage 3 embedding pipeline. Minimal HTTP surface:
   `POST /chunks`, `POST /search`.
2. **rsc.wiki scraper** (one-time tool) — chunk by section
   heading, ~500–800 tokens/chunk, filter by F2P/P2P metadata.
   Embed and load into mesa.
3. **AutoRune script corpus ingest** — Alex's historical
   AutoRune scripts are first-class reference material for "how
   would a script normally do X." Chunk by script + by section,
   embed, tag with `source='autorune'`.
4. **`recall()` stdlib wiring** — replace the Phase-2 stub. DSL
   routines call `recall(query, top=5)` and receive a list of
   relevant chunks they can branch on.
5. **`cradle -knowledge-query "how do I cook a lobster"`** —
   admin CLI surface for sanity-checking the corpus from the
   command line.

**Validation**: from a routine, `chunks = recall("how do I
identify herbs"); say(f"top hit: {chunks[0].text}")` returns
the rsc.wiki section on herblaw identification. Same query from
the cradle CLI returns identical chunks. Total ingest cost
under $1 for both corpora.

### Phase 2.7 — Admin tooling (delores becomes an admin) — ✓ delivered (differently)

> **Status:** substantially delivered, but via the generic `command()` builtin
> (`b876c2d`) + server `::` commands rather than the dedicated `admin_*` verbs
> below. The test hosts are admin per the DB clone; the 196-scenario corpus
> already uses admin `::commands` for preconditions. Only a validator-time
> `IsAdmin` gate (so production hosts reject `command()`) remains optional.

**Goal**: delores has admin powers so we can bypass grind and
test edge cases directly. Admin actions are exposed both
through cradle CLI flags and as DSL actions.

This phase makes Phase 2.8 (live build/test) practical. Without
admin, testing "can the host complete quest X" requires actually
leveling delores to the quest's requirements first, which can
take days or weeks per quest.

1. **OpenRSC server-side**: grant delores admin in the server's
   `players.json` (or equivalent). One-time config change.
2. **Admin DSL actions** (new wrappers around the existing
   `Host.Command()`):
   - `admin_set_stat(skill, level)` — bypass grind for testing
   - `admin_give_item(item, count)` — spawn items into inventory
   - `admin_teleport(x, y)` or `admin_teleport(region_name)`
   - `admin_heal()` / `admin_set_fatigue(value)`
   - `admin_set_position(x, y)` (alias for teleport)
   - `admin_clear_inventory()`
   - `admin_kick(player)` (rare; for clearing other hosts from
     a test region)
3. **Validator + safety**: admin actions only resolve when the
   host's config flag `IsAdmin: true` is set. Production hosts
   reject these at parse time.
4. **`admin_sandbox.routine`** — example routine that sets up
   common scenarios: "give me a full inventory of food, teleport
   to swamp, set fishing to level 60."

**Validation**: `cradle -username delores -routine
admin_sandbox.routine` boots delores into a configured testing
state in under 10 seconds. From there, any `.routine` runs
against the configured state.

### Phase 2.8 — Live build & test (edge-case discovery) — 🔄 IN PROGRESS

> **Status:** active as the scenario **"run-to-ground" campaign** — the
> 196-scenario corpus (`examples/scenarios/*/*.routine` — what the runners
> execute) swept against a live server, failures run to ground in the engine. Rounds r1+r2 (merged at
> `18ac18b`) → r3 (`fd0731c`/`5117845`/`272ca58`/`40bea3a`); pass-rate 16
> baseline → ~70/88 mid-cycle → still hardening. The state machines among the
> "outputs" below
> (trade/duel/death/bank/dialog/dynamic-boundaries) already **shipped** in
> Phase 2.5. Mechanics: [`scenarios.md`](scenarios.md).

**Goal**: drive out every edge case in the world by actually
playing the game with admin scaffolding. World integration
(skills, quests, banking, trades, duels) gets done **here**,
discovered as failures during live test runs rather than spec'd
upfront.

The structure of work:

1. **Skills pass**: for each of the 18 RSC skills, write a
   `<skill>.routine` and run it. When it fails, fix the
   underlying Host method / packet / world-state code. Repeat
   until all 18 are exercised end-to-end with admin scaffolding.
2. **Quests pass**: pick quests by complexity (easiest first —
   Cook's Assistant, Sheep Shearer, then Restless Ghost, etc).
   Each quest becomes a `quest_<name>.routine` that uses admin
   to set prerequisites (right items, right stats) and then
   plays through. Failure modes inform what's missing in the
   action / dialog / event layers.
3. **Multi-host scenarios**: trades (#39), duels (#40),
   AttackPlayer (#44) — using alex + delores as a pair. Each
   becomes a paired routine on each side.
4. **Banking orchestrator** (#29) becomes a single routine.
   Death/respawn recovery (#28) becomes a `when` handler on
   `self.hp == 0`. NPC dialog choice (#27) gets exercised
   during the quests pass. Dynamic boundary state (#26)
   surfaces during quests that cut webs or open special doors.

Pre-existing tasks #26–#29, #39–#45 are all the **outputs** of
this phase, not separate work items. Each gets closed as the
edge case it represents is encountered, fixed, and verified.

**Validation**: delores completes 3+ F2P quests end-to-end with
admin scaffolding. Every F2P skill exercised to at least level
20 in a recorded run. A "tour" routine plays through 30
minutes of mixed activity (banking, fishing, dialog, combat)
without crashing.

## Subsequent phases (mostly unchanged)

### Phase 3 — Full Mesa (per-host memory + relational records)

**Goal**: the rest of mesa beyond the knowledge corpus. Per-host
memory, episodic events, reflections, relational records.

Schema-level rename `bots`/`bot_id` → `hosts`/`host_id` lands
here. Validation: `mesa.RegisterHost`, `WriteEpisode`,
`QueryEpisodes` work end-to-end; relational records form across
sessions.

### Phase 4 — Brain + cognition  ✓ CORE SHIPPED (2026-06-06)

> **A host can now be driven autonomously by an LLM.** The `MesaDirector` Act loop
> + the gRPC mesa service (`mesa/mesad`, `mesa/llm`, Sonnet/Haiku tiering) + persona
> provisioning are live — **Delores autonomously completed tutorial island.**
> Outstanding within the phase: RAG, durable memory accretion, relational records /
> trust ledger. Forward plan in `cognition-and-autonomy.md`.

**Original goal**: a host can be driven by an LLM strategist that uses
mesa for memory and the knowledge corpus for RAG.

**Major design session required first** — persona / cohort /
reverie design ([`personas.md`](personas.md), [`reveries.md`](reveries.md)).
This session also **unblocks `super()` / `extends host`** from
Phase 2.5 (Task #52) because it defines what a persona-tier
default handler actually IS.

**Validation**: single host runs autonomously ~1 hour against
OpenRSC. Brain decides; routines execute; reveries inject;
memory accretes; relational records form.

### Phase 4.5 — Wiki RAG cohort experiment

Now an experiment, since the corpus already exists. Cohort
`wiki_enabled` vs `no_wiki` for ~1 week to measure impact.

### Phase 5 — Reveries (full)

Real persona-driven, emotion-aware reveries. Timing jitter,
idle behaviors, spontaneous chat, mistake injection,
persona-specific quirks.

### Phase 6 — Delos (orchestrator + UI)

Swarm management binary + "technician tablets" web UI. Live
host inspection, chain-of-thought capture, cohort analytics.

### Phase 7 — Scale rollout

| Wave | Size | Goal |
|---|---|---|
| 7a | 10 hosts | Validate end-to-end at small scale |
| 7b | 100 hosts | Mesa perf, cost scaling, log volume |
| 7c | 500 hosts | Target population; research begins |

### Phase 8+ — Open research

Long-horizon goal accomplishment, community/clustering analysis,
ethics observation, cohort experiments, inter-host learning
patterns.

## Order of dependencies

**Planned ladder (original):**
```
Phase 0 ─► 1 ─► 2 ─► 2.5 ─► 2.6 ─► 2.7 ─► 2.8 ─► 3 ─► 4 ─► 4.5 ─► 5 ─► 6 ─► 7 ─► 8+
```

**What actually happened (through HEAD `fd0731c`, 2026-05-31):**
```
Phase 0 ─► 1 ─► 2 ─► 2.5 ─► 2.9 (BODY freeze + build-out, #114–123) ─► 2.8 (scenario
                                run-to-ground campaign, admin already in hand)
                            └─► render engine + spectator  (out-of-band, parallel)
   2.6 (knowledge ingestion) ── SKIPPED, never started ──┐
                                                          ▼
                       NOW: build UP (cognition → brain → persona → memory →
                       reveries → mesa; subsumes 2.6) + build OUT (OpenRSC
                       stewardship) ── handed to external agents
                                                          ▲
                                              persona/reverie design session
                                              (required gate for Phase 4; also
                                              unblocks the parked #93 super())
```

Order-of-operations notes (reconciled with reality):

- **Phase 2.5 was the unlock**, as planned — the query layer + control flow.
- **2.9 (BODY freeze + build-out) came next, not 2.6.** The team chose to
  freeze + harden the body contract before building cognition on it.
- **Admin (2.7) arrived ahead of schedule** via `command()` (`b876c2d`), so the
  **2.8 live-test campaign ran without waiting on 2.6.** The original
  "2.6 → 2.7 → 2.8" ordering was not followed.
- **2.6 (RAG) was skipped** and now folds into the higher-layer/mesa
  scaffolding (build UP).
- **2.8 is open-ended** and ongoing (run-to-ground campaign at 196 scenarios).
- **Persona/reverie design is still the gate to Phase 4** and unblocks the
  parked `super()` / `extends host` (#93).
- **Phase 3 (full mesa) can run in parallel** with the remaining scenario
  hardening — the per-host memory schema is independent of it.
