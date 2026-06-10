# Phases

> **STATUS (verified 2026-06-10 against branch `tidy/structure-and-docs`, HEAD `0bfa818`).**
> This doc is the project's phase ladder: the historical record of what shipped in
> what order, and the **decoder** for phase-number references ("Phase 2.9",
> "Phase 4") used across docs, commits, and code comments. Phases 0–4 are BUILT —
> several delivered out of order or in a different shape than planned; each
> section states exactly how. **Open work lives in one place:
> [`TODO.md`](TODO.md)** — this doc cites its IDs (R-1, P-6, O-7, …) instead of
> carrying a backlog. The closed per-task ledger (#19–#123) is
> [`tasks.md`](tasks.md).
>
> *Phase-number namespaces:* this ladder (0–8+) is not the only one. The
> cognition workstream numbered its own internal phases (1–5b) inside
> [`world-knowledge-and-learning.md`](world-knowledge-and-learning.md), and
> `docs/dsl.md` has its own 9 sub-phases inside Phase 2. `TODO.md`'s header
> documents all the `#N` namespace collisions.

> **Host** — an autonomous AI actor in the system. One host = one `runtime.Host`
> = one logged-in OpenRSC character. Hosts run as a fleet inside the
> `cmd/cradle-server` daemon, or one-off via `cmd/host` (conductor/REPL runner)
> or `cmd/legacy-cradle` (the original flag-driven harness). See
> [`lang/README.md`](lang/README.md) for terminology.

## Completed phases

### Phase 0 — Foundation: wire + walk ✓ (early 2026)

Wire-protocol + walk + logout. One binary, one host, one
deterministic round-trip.

Shipped: `proto/v235`, `session`, minimal `world`, minimal
`action`, `cmd/legacy-cradle` (named `cmd/cradle` until the
daemon split). ~650 LOC.

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
conformance), `runtime/dsl_*.go` bridges, `legacy-cradle -routine`
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

One item deferred: per-handler `extends host` + `super()` (#93 = v2 of #52),
parked on the Phase-4 persona tier. **That tier now exists, so #93 is
unblocked** — tracked as [`TODO.md`](TODO.md) DSL-13. (`repeat_until` #85 has
since **shipped** as `repeat { … } until <cond> timeout <expr>` — see
[`tasks.md`](tasks.md) Stage 2 mop-up.)

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
cognition/brain/persona layers were then built on. (At the time it was framed
as a handoff point to external agents; the upper layers were in fact built
in-repo — the dev-partner charter is archived at
`archive/initial-brainstorming/westworld-dev-partner.md`.)

### Render engine + live spectator ✓ (shipped out-of-band)

A decoupled, read-only Go RSC viewport renderer + live browser spectator
(`legacy-cradle -spectate`) landed **independent of the phase ladder** (~67
commits, 2026-05-29 → 05-31). It reconstructs a host's *perceived* scene
pixel-matched to RSCPlus — host self-perception + human observability. Details:
[`render-engine.md`](render-engine.md). The workstream is in maintenance; the
2026-05-31 `.claude/HANDOFF.md` bug queue is retired, polish residue is
[`TODO.md`](TODO.md) O-14. (Observability was originally slated for Phase 6;
the host-perception renderer arrived early and out of sequence.)

## The 2.6 → 2.8 reconciliation (design record + what actually happened)

**What actually happened vs. the plan below:** after Phase 2.5 the team did the
BODY freeze + build-out (2.9) and a multi-round scenario "run-to-ground"
campaign — with admin already in hand via `command()` — *before* the 2.6
knowledge work. Phase 2.6's Slice 1 then landed in-repo (`ccbc220`:
`cognition/corpus` + `recall()` wiring) and the rest of 2.6 manifested as the
**mesa knowledge stack** (gRPC `Recall`, Postgres+pgvector LTM, `mesa/wikirag`).
The "build UP" that this section once deferred to external agents **happened
in-repo over 2026-06-01 → 06-10**: cognition → brain → persona → memory →
mesa (see Phases 3/4 below and the diagram at the bottom). The 2.6/2.7/2.8
sections below are kept for their design content, annotated with their real
status.

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
3. **REPL** (#54). Two modes: standalone `legacy-cradle -repl` and
   pdb-style `-repl-on-fail` drop-in on routine failure
   with `.resume`. *(The standalone REPL shipped — `legacy-cradle
   -repl`, and `cmd/host` defaults to a REPL when no routine is
   given; `-repl-on-fail` never did — TODO DSL-27.)*

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
   until persona-tier defaults exist. *(Now unblocked — TODO DSL-13.)*

**Validation** (historic; this milestone is met): rewrite the goblin-killing
routine to use `when self.hp < 30 { eat!(...) }` instead of relying on the Go
auto-eat. Author the new version interactively in the REPL, save once it runs
clean for 30 minutes. Delete `runtime/auto_eat.go` + `runtime/combat_loop.go`
once their routine equivalents are live and tested. *(Done — the equivalents
now live as `examples/routines/auto_eat.routine` + `combat_loop.routine`; the
once-referenced `kill_goblins.routine` no longer exists, see
`kill_one_goblin.routine` / `safe_chicken_killer.routine`.)*

### Phase 2.6 — Knowledge ingestion (RAG corpus) — ✓ delivered (Slice 1 in-phase, the rest as the mesa knowledge stack)

> **Status:** Slice 1 SHIPPED (`ccbc220`): `cognition/corpus`
> (`MemoryCorpus` + `LoadWikiDump`, namespace gating + federation),
> `legacy-cradle -wiki-dump` / `-dev` flags, and `recall()` wired —
> `dslRecall` in `runtime/actions_ambient.go` queries the corpus first
> (provenance-prefixed chunks) and falls back to the per-host retriever;
> it fails `NOT_IMPLEMENTED` only when neither is wired. **The "recall()
> is a stub" era is over.** Slices 2+ manifested differently than designed:
> not an HTTP chunk service but the gRPC `Knowledge.Recall` RPC
> (`mesa/proto/mesa.proto:173`) against mesa's LTM (Postgres + pgvector +
> Voyage embeddings — `mesa/mesad/ltm.go`, `mesa/embed/voyage.go`), with
> `mesa/wikirag` as the offline wiki embed/search tool. Slice plan + chunk
> format: [`lang/knowledge.md`](lang/knowledge.md). Original design kept
> below, annotated per item.

**Goal**: hosts can consult external knowledge (rsc.wiki +
historical AutoRune scripts) via `recall()` and stdlib queries
without needing the full brain layer.

Strategy: stand up enough of mesa to serve the knowledge corpus,
defer per-host memory infrastructure to Phase 3.

1. **Mesa knowledge_chunks schema + embedding worker** — Postgres
   + pgvector, Voyage 3 embedding pipeline. Minimal HTTP surface:
   `POST /chunks`, `POST /search`. *(Manifested as the gRPC
   `Recall` RPC + `episodes` LTM with pgvector cosine ranking —
   no HTTP surface, no separate `knowledge_chunks` table.)*
2. **rsc.wiki scraper** (one-time tool) — chunk by section
   heading, ~500–800 tokens/chunk, filter by F2P/P2P metadata.
   Embed and load into mesa. *(Shipped: clean goquery extraction +
   the `mesa/wikirag` Voyage embed/search tool — offline tooling,
   not host-runtime; live hosts load wiki pages via `-wiki-dump`.)*
3. **AutoRune script corpus ingest** — tag with `source='autorune'`.
   *(Never done — TODO M-8; the "discover, don't seed" decision in
   [`world-knowledge-and-learning.md`](world-knowledge-and-learning.md)
   forbids it as live host knowledge. The `dev` namespace gate exists,
   unused.)*
4. **`recall()` stdlib wiring** — replace the Phase-2 stub. *(Done —
   `dsl/spec/actions.go:480`, `runtime/actions_ambient.go` `dslRecall`.)*
5. **`legacy-cradle -knowledge-query "how do I cook a lobster"`** — admin CLI
   surface for sanity-checking the corpus. *(Never built — TODO M-9;
   `mesa/wikirag`'s CLI partially covers it.)*

**Validation** *(met via the corpus path)*: from a routine, `chunks =
recall("how do I identify herbs"); say(f"top hit: {chunks[0]}")` returns
provenance-prefixed wiki chunks (`[source § page § section] text`); the
same content is queryable offline via `mesa/wikirag`.

### Phase 2.7 — Admin tooling (delores becomes an admin) — ✓ delivered (differently)

> **Status:** substantially delivered, but via the generic `command()` builtin
> (`b876c2d`) + server `::` commands rather than the dedicated `admin_*` verbs
> below. The test hosts are admin per the DB clone; the scenario corpus (now
> 263 routines) uses admin `::commands` for preconditions. Only a
> validator-time `IsAdmin` gate (so production hosts reject `command()`)
> remains optional — TODO DSL-16.

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

*(Of the four: item 1 happened; items 2–4 were never needed — raw
`command()` strings in routines proved sufficient.)*

### Phase 2.8 — Live build & test (edge-case discovery) — ✓ campaign mature (corpus open-ended)

> **Status:** the scenario **"run-to-ground" campaign** — the corpus
> (`examples/scenarios/{combat,edges,movement,quests,skills,social}`, now
> **263** `.routine` files) swept against a live server, failures run to
> ground in the engine. Rounds r1+r2 (merged at `18ac18b`) → r3
> (`fd0731c`/`5117845`/`272ca58`/`40bea3a`). The campaign is in mature
> maintenance: TODO S-1 carries the open engine gaps, S-2 the two unmerged
> proposal files, S-5 the catalog CI check. The state machines among the
> "outputs" below (trade/duel/death/bank/dialog/dynamic-boundaries) already
> **shipped** in Phase 2.5. Mechanics: [`scenarios.md`](scenarios.md).

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
without crashing. *(This bar was never formally recorded and is
superseded in spirit by autonomous play — whether it still matters
is an operator call, TODO S-4.)*

## Subsequent phases — what each became

### Phase 3 — Full Mesa (per-host memory + relational records) — ✓ substantially delivered (out of order)

Delivered across the 2026-06 memory-spine + mesa workstreams, in a different
shape than planned:

- **Per-host durable memory**: `hostkv` (pebble-backed local store) + the
  tiered `memory.Manager` (L0 scratch / L1 local / L2 mesa remote) behind the
  location-blind DSL verbs `remember`/`recollect`/`forget`
  (`dsl/spec/actions.go:486-492`, `memory/manager.go`, `memory/policy.go`).
  Full doc: [`memory.md`](memory.md).
- **Episodic events**: `runtime/memory_journal.go` on the host side; mesa LTM
  `episodes` table (Postgres + pgvector + Voyage embeddings,
  `mesa/mesad/ltm.go`) fed by the client-streaming `Journal.Remember` RPC;
  consolidation + insight crons (`mesa/mesad/cron.go`, `cron_insight.go`)
  distill it.
- **Relational records**: the Beta(α,β) trust ledger lives **host-side**
  (`limbic/ledger.go`, host-authoritative per the AuthLocal decision) and is
  mirrored to mesa via `Journal.SyncRelationships` — the reverse of the
  planned mesa-graded design. Remainder (prior choice, decay, LLM overlay):
  TODO C-24.
- The planned schema rename `bots`/`bot_id` → `hosts`/`host_id` is **MOOT** —
  the schema was born `host_id`-native (`mesa/mesad/ltm.go:117`).

The planned validation ("RegisterHost / WriteEpisode / QueryEpisodes work
end-to-end; relational records form across sessions") is met under different
names: `Journal.Remember`/`Knowledge.Recall`/`Provision.Fetch` + the
relationship sync pair. Open tuning questions: TODO C-15, M-2.

### Phase 4 — Brain + cognition ✓ CORE SHIPPED (2026-06-06), leftovers since landed

> **A host is driven autonomously by an LLM.** The `MesaDirector` Act loop
> (`runtime/mesa_director.go`) + the gRPC mesa service (`mesa/mesad`, `mesa/llm`
> Anthropic client with per-seam Act/Decide/Genesis models) + persona
> provisioning are live — **Delores autonomously completed tutorial island**
> (`dacc425`). The items the original banner listed as outstanding have all
> since landed: RAG (`Knowledge.Recall` + LTM), durable memory accretion
> (journal + LTM + crons), relational records / trust ledger
> (`limbic/ledger.go`). Cognition model + forward narrative:
> [`cognition-and-autonomy.md`](cognition-and-autonomy.md); residue:
> [`TODO.md`](TODO.md) §3 (Cognition).

**Original goal**: a host can be driven by an LLM strategist that uses
mesa for memory and the knowledge corpus for RAG.

The plan's claim that a **persona/cohort/reverie design session must gate this
phase** was falsified by events: the brain shipped first, on a hand-authored
persona, and the persona schema/compiler followed (`persona/`,
`mesa/personacook`). The remaining design sessions proceed via the worksheet —
TODO P-1/P-2. The session's other deliverable — unblocking `super()` /
`extends host` (#93) — is now unblocked by the existing persona tier (TODO
DSL-13).

**Validation** *(met by autonomous play)*: a host runs autonomously against
OpenRSC — genesis at login, the Act loop decides, routines execute, memory
accretes via the journal, detours interrupt and resume. Reverie injection is
the one validation clause still open — that is Phase 5.

### Phase 4.5 — Wiki RAG cohort experiment — open (TODO R-1)

Cohort `wiki_enabled` vs `no_wiki` for ~1 week to measure impact. Not started;
the substrate exists (`cohort_id` on the persona — `persona/persona.go:170` —
and in the LTM persona record; the corpus + `Recall` path).

### Phase 5 — Reveries (full) — not started (TODO P-6)

Real persona-driven, emotion-aware reveries. Timing jitter, idle behaviors,
spontaneous chat, mistake injection, persona-specific quirks. Design:
[`reveries.md`](reveries.md) — with one standing correction from the decision
record: the engine ships as a **sibling-goroutine scheduler driving public
effector methods**, not a `reverie.tick()` interpreter hook. Pieces already
manifested: limbic mood (`limbic/affect.go`), persona dials + `ReverieSeed`
(`persona/persona.go`). Prerequisite: ActionArbiter + a real `is_busy`
(TODO P-9).

### Phase 6 — Orchestrator + UI (was "Delos") — ✓ substantially pre-delivered by the cradle daemon

No binary named `delos` exists or is planned. The phase's intent — swarm
management + "technician tablets" + live host inspection + chain-of-thought
capture — was delivered out-of-band by the **cradle daemon** family:

- `cradle/` (registry + supervision + HTTP/JSON API + web UI:
  `cradle/registry.go`, `cradle/api.go`, `cradle/webui.go`, `cradle/web/`),
  run by `cmd/cradle-server` (the whole fleet in ONE process over shared
  facts/landscape, one mesa conn per host), driven by `cmd/cradle-ctl`.
- The per-host `debughttp` control plane (`POST /eval`, `/script`,
  `GET /state`, `/events`, JSONL event log) and the persisted decision stream
  (`runtime/decisions.go`).
- The mesa **Admin plane** + `cmd/mesa-ctl` (persona put/import/ls/get/set/rm,
  goal push, fleet gen) — see [`mesa/admin-control-plane.md`](mesa/admin-control-plane.md).

Residue: cohort/archetype analytics, cost dashboard, perturbation toolkit
(TODO O-7) and the LLM cost/token ledger — nothing tracks tokens or dollars
today (TODO O-9).

### Phase 7 — Scale rollout — open (TODO R-2)

| Wave | Size | Goal |
|---|---|---|
| 7a | 10 hosts | Validate end-to-end at small scale |
| 7b | 100 hosts | Mesa perf, cost scaling, log volume |
| 7c | 500 hosts | Target population; research begins |

Ad-hoc 200-drone load tests have run (`cmd/dronegen` persona families), but
the formal waves have not. The standing instrumented soak — with acceptance
criteria — is TODO O-5.

### Phase 8+ — Open research — open (TODO R-3)

Long-horizon goal accomplishment, community/clustering analysis,
ethics observation, cohort experiments, inter-host learning
patterns. The named learning observable — punt-rate instrumentation — has
zero code; the metrics harness is TODO R-4.

## Order of dependencies

**Planned ladder (original):**
```
Phase 0 ─► 1 ─► 2 ─► 2.5 ─► 2.6 ─► 2.7 ─► 2.8 ─► 3 ─► 4 ─► 4.5 ─► 5 ─► 6 ─► 7 ─► 8+
```

**What actually happened, part 1 (through `fd0731c`, 2026-05-31):**
```
Phase 0 ─► 1 ─► 2 ─► 2.5 ─► 2.9 (BODY freeze + build-out, #114–123) ─► 2.8 (scenario
                                run-to-ground campaign, admin already in hand)
                            └─► render engine + spectator  (out-of-band, parallel)
```

**Part 2 (`fd0731c` → HEAD `0bfa818`, 2026-06-01 → 06-10) — the "build UP" happened:**
```
fd0731c ─► cognition/mesa spine: hostkv + conductor + pearl + tiered memory +
           limbic ledger; mesa gRPC daemon (mesad: Act/Decide/Chat/Recall +
           LTM + crons); persona schema + compiler + personacook
           ─► AUTONOMOUS PLAY — Delores completes tutorial island (dacc425)
        ─► knowledge & learning: knowledge ledger + goal graph + forage drive
           (the w-k-l doc's internal Phases 1–5b)
        ─► plutonium comparative study ─► movement rebuild: live collision
           overlay, plane-aware BFS, door handling (b24012c → 6a329fd) +
           fatigue arbiter / clean-EOF soak trio (0214a0b)
        ─► control plane: cradle daemon + web UI + cradle-ctl (730ae9a);
           mesa Admin plane v1 + mesa-ctl (1c2423e)
        ─► docs-grounding tidy (this branch, HEAD 0bfa818)
```

Order-of-operations notes (reconciled with reality):

- **Phase 2.5 was the unlock**, as planned — the query layer + control flow.
- **2.9 (BODY freeze + build-out) came next, not 2.6.** The team chose to
  freeze + harden the body contract before building cognition on it.
- **Admin (2.7) arrived ahead of schedule** via `command()` (`b876c2d`), so the
  **2.8 live-test campaign ran without waiting on 2.6.** The original
  "2.6 → 2.7 → 2.8" ordering was not followed.
- **2.6 (RAG) split**: Slice 1 shipped in-repo (`ccbc220`); the mesa-backed
  remainder manifested as the `Recall` RPC + LTM + wikirag, not the designed
  HTTP chunk service.
- **2.8 is open-ended** — the corpus stands at 263 scenarios; the campaign is
  mature maintenance (TODO S-1).
- **The "persona design gates Phase 4" rule was falsified** — the brain shipped
  first on a hand-authored persona; #93/`super()` is unblocked (TODO DSL-13).
- **Phase 3 did run out of order** — delivered alongside Phase 4, not before
  it, and host-authoritative rather than mesa-authoritative for relationships.
- **Phase 6 was pre-delivered** out-of-band by the cradle daemon rather than
  built as a phase.

Everything still open is in [`TODO.md`](TODO.md) — the single source of truth
for work items; this doc carries no backlog.
