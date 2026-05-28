# Phases

A staged build plan. Each phase has a clear validation milestone —
something that works end-to-end and proves the architecture is
sound before moving on. This doc is the canonical "what's next"
plan; sub-tasks are tracked in the task list (#19+).

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

## Current plan — Phase 2.5 → 2.8

The high-level strategy: get scripting solid, ingest the
knowledge corpus, then use admin-powered live testing to drive
out edge cases across quests and skills. World integration is
done **reactively as edge cases surface**, not as a separate
upfront phase.

### Phase 2.5 — Language v2 (next)

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

**Stage 3: Deferred.**

9. **`super()` / `extends host`** (#52) — **DEFERRED to Phase 4**
   until persona-tier defaults exist.

**Validation**: rewrite `examples/routines/kill_goblins.routine`
to use `when self.hp < 30 { eat!(...) }` instead of relying on
the Go auto-eat. Author the new version interactively in the
REPL, save once it runs clean for 30 minutes. Delete
`runtime/auto_eat.go` + `runtime/combat_loop.go` once their
routine equivalents are live and tested.

### Phase 2.6 — Knowledge ingestion (RAG corpus)

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

### Phase 2.7 — Admin tooling (delores becomes an admin)

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

### Phase 2.8 — Live build & test (edge-case discovery)

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

### Phase 4 — Brain + cognition

**Goal**: a host can be driven by an LLM strategist that uses
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

```
Phase 0 ─► 1 ─► 2 ─► 2.5 ─► 2.6 ─► 2.7 ─► 2.8 ─► 3 ─► 4 ─► 4.5 ─► 5 ─► 6 ─► 7 ─► 8+
                                                              ▲
                                                              │
                                                  persona/reverie design session
                                                  (required gate; also unblocks
                                                  Phase 2.5 super())
```

Order of operations key points:

- **Phase 2.5 is the unlock.** Until the query layer + control
  flow lands, every world-integration routine is hamstrung.
- **2.6 (RAG) before 2.7 (admin tools)** because delores benefits
  from being able to `recall()` while we're testing — turns
  "what should the host do here?" into a recallable hint.
- **2.7 before 2.8.** Without admin, live edge-case testing
  takes orders of magnitude longer because everything has to
  be earned in-game first.
- **2.8 is open-ended.** Phase ends when delores can roundtrip
  3+ quests and 18 skills are exercised; could be 2 weeks or
  2 months.
- **Persona/reverie design is still the gate to Phase 4.** Same
  as the prior plan. The new wrinkle: it also unblocks
  `super()` / `extends host` parked since Phase 2.5.
- **Phase 3 (full mesa) can run in parallel with 2.7/2.8** —
  the knowledge subset is enough to unblock testing; the
  per-host memory schema is independent.
