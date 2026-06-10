# Development Workflow — REPL-Driven, Delores as Collaborator

> **STATUS: CURRENT** (verified against code 2026-06-10, HEAD `0bfa818`).
> The REPL loop described here is BUILT (`runtime/repl.go` on top of
> `dsl/interp/repl.go` Sessions) and was how the DSL surface got built out.
> The now-primary loop is autonomous — conductor + cradle daemon +
> debug-http — described in its own section below. Phase framing is
> historical; the canonical ladder is [`../phases.md`](../phases.md).

> **Host** — an autonomous AI actor. Delores is one of these
> hosts. She's also one of the **build assistants** for
> westworld itself.

## The model

The development team for westworld is three collaborators:

1. **Alex** (human) — decides what we're trying to do, makes
   design calls, drives the REPL session
2. **Claude** (AI, this conversation) — writes Go infrastructure
   (parser, validator, interpreter, runtime), authors routines,
   updates docs
3. **Delores** (AI host, running in OpenRSC) — provides live
   world-state observation, executes actions, surfaces real
   server behavior, discovers gaps

Delores is **not just a test subject**. She's how we find out
what the world actually does, what packets really come back,
what edge cases exist that we wouldn't think to spec upfront.
Every gap she exposes is one less surprise in production.

## The loop

The inner development loop (entry points: `host -username <name>`
drops into the REPL when no routine/goal is given, or force it
with `-repl`; `legacy-cradle -repl` is the older path):

```
1. Alex / Claude form a hypothesis: "we should be able to bank
   a load of iron ore by talking to the banker, then depositing
   each loot slot."
2. Drive delores through it step-by-step in the REPL:
       >>> talk_to(banker)
       >>> .state                       # what dialog options came back?
       >>> answer(2)
       >>> bank.is_open                  # did the bank UI open?
       >>> bank.deposit(inventory.slots[0].item, inventory.slots[0].amount)
       >>> ...
3. Observe what works, what fails. Three outcomes:

   a. Everything works → save the sequence as
      examples/routines/<name>.routine and add a conformance
      test. Done.
   b. A query is missing (e.g. .state didn't show dialog
      options) → Claude adds the accessor to the view family in
      runtime/views_*.go (views_self.go, views_world.go,
      views_inventory.go, views_bank.go, ...), rebuilds, we retry.
   c. An action fails unexpectedly → Claude inspects the wire
      capture, identifies whether it's an action wrapper bug, a
      packet-decoder bug, or a server quirk we need to document.
      Fix, retry.
4. Once stable, commit the Go code AND the working routine.
```

Cycle time per iteration: minutes for query/accessor additions,
sometimes hours for packet-level bugs. Faster than writing a
spec upfront and grinding through it offline.

## The now-primary loop: conductor + daemon + debug-http

The stdin REPL above is how the language surface was built; it
still works, but day-to-day development now runs against an
**autonomous** host and observes/steers it over HTTP:

- **Conductor** (`runtime/conductor.go`) — the host's autonomy
  spine. It repeatedly selects an Intent (a routine file or
  mesa-authored inline DSL), runs it to completion, and observes
  the result before choosing the next. It runs concurrently with
  `Host.Run` (the frame pump that keeps the world mirror live).
  `cmd/host` is the single-host runner: `-routine`/`-routines`
  for fixed sequences, `-goal` + `-mesa` for the mesa Act planner
  (situation → freshly-authored DSL).
- **Cradle daemon** (`cmd/cradle-server` + package `cradle/`) —
  the fleet control plane: loads `*.hostcfg` (file or directory),
  runs and supervises every host in one process over shared
  static deps, and mounts the HTTP/JSON control API + web UI
  (`cradle/api.go`, `cradle/webui.go`, default `localhost:8099`).
  `cmd/cradle-ctl` is its CLI.
- **debug-http** (package `debughttp/`) — the scriptable per-host
  control plane: `POST /eval` (one DSL line, same Session model
  as the REPL), `POST /script` (a whole routine), `GET /state`,
  `GET /events`, `GET /mind`, a `/ws` live thought stream, and a
  browser dashboard at `/`, plus a JSONL event log. Enable with
  `legacy-cradle -debug-http` (default `localhost:8090`) or
  `host -debug-addr localhost:8090`.

The hypothesis→observe→fix cycle is unchanged; what changed is
that delores plays *by herself* and the humans interrogate her
mid-run over `/eval` and `/state` instead of holding her stdin.

## Why this works for westworld specifically

- **The world is the source of truth.** RSC's server behavior is
  not fully documented anywhere. OpenRSC source code is the
  closest thing, but it has its own divergences from authentic
  RSC. Asking delores is faster than reading server source.
- **The DSL exists to be written in.** Every routine we author
  via REPL stress-tests the language and surfaces ergonomic
  issues we wouldn't catch in synthetic tests.
- **Discovery is high-bandwidth.** A 30-minute REPL session
  often surfaces 5–10 gaps. A 30-minute spec session surfaces
  maybe one.
- **No authorization friction.** Pre-launch, the host accounts
  and the world state are sandbox. Delores can attack, trade,
  drop items, die, teleport, set stats — none of it matters.
  **Before going live (the scale rollout in
  [`../phases.md`](../phases.md)), the world is reset and the
  host accounts are recreated.** Until then, experimentation is
  unconstrained.

## What gets persisted, and where

| Artifact | Persisted to | Committed? |
|---|---|---|
| Production-quality routines | `examples/routines/<name>.routine` | yes |
| Conformance test cases | `testdata/conformance/<n>_<name>.routine` + `.expected` | yes |
| Go infrastructure (Host methods, view structs, validator changes) | `runtime/`, `dsl/` | yes |
| Doc updates triggered by discoveries | `docs/`, `docs/lang/` | yes |
| Casual REPL experimentation, scratch routines, half-baked drafts | `local/routines/` | **no — gitignored** |
| Server-side admin-action exploration notes | `local/notes/` | no — gitignored |
| debug-http JSONL event logs | `/tmp/cradle_debug/<username>_events.jsonl` (override with `-debug-log`) | no — outside the repo |

The `local/` convention: anything under it is yours/ours to
shape however we want, never goes to github. Use it for
exploration, half-finished experiments, personal scratchpads.
The repo stays the curated artifact; `local/` is the workshop.

`.gitignore` includes `local/` at repo root.

## REPL session etiquette

A few habits that make sessions productive:

- **Open with state.** First line of every session is `.state` —
  prints position, vitals (hp/prayer), fatigue, combat level, an
  inventory summary, and the visible-NPC count. Reset the mental
  model of where delores is. (debug-http's `GET /state` serves a
  richer JSON snapshot — inventory contents, nearby NPCs/players,
  dialog state, recent server messages.)
- **One change per try.** Don't bundle "open bank AND deposit
  AND withdraw" into a single command sequence. The atomic
  approach makes it obvious where failures land.
- **Save what works.** When a sub-sequence runs clean, dump it
  to `local/routines/<scratch>.routine` immediately. Don't
  rely on memory.
- **Promote when stable.** Once a `local/` routine has been run
  3+ times without modification and survives the conformance
  harness, promote it to `examples/routines/` or to a paired
  `testdata/conformance/<name>.routine` + `.expected`.
- **Log the surprises.** When delores does something
  unexpected, note it. The note might become a doc update, a
  bug fix, or a new conformance case.

## Roles inside a session

The three roles aren't formal — they shift moment to moment.
But a useful default:

- **Alex** sets the destination ("let's get banking working
  end-to-end") and picks which experiment to try next.
- **Claude** translates the hypothesis into REPL input,
  observes results, proposes next steps, edits Go when something
  needs scaffolding.
- **Delores** executes against the live server and reports
  back. Her "reporting" is just the REPL output plus whatever
  query accessors we've built.

When something goes wrong, the fix is almost always Claude
edit-Go-rebuild. Alex sets priorities. Delores reveals truth.

## How this shaped the build order (historical)

This model inverted Phase 2.5's ordering from "build everything,
test at the end" to "build the minimum dev surface, then iterate
live" — and that's what happened: Stage 1 (Result/Error model +
bang variants, filename enforcement + `ParseRoutineString`, the
REPL itself) shipped first, then Stage 2 (query-layer expansion,
`when`, `defer`, `try`/`recover`, `select`) was built accessor by
accessor against live delores, each construct tried in a scratch
routine before its Go implementation was considered done. The
same pattern carried through knowledge ingestion (`recall(...)`
exercised at the prompt), admin tooling (each admin action
registered, exercised in REPL, promoted to a routine), and live
edge-case testing. See [`../phases.md`](../phases.md) for the
canonical phase ladder and what actually shipped when; this doc
explains the *how*, that doc the *what*.

## Production launch boundary

The live-driven model only applies pre-launch. At the scale
rollout (see [`../phases.md`](../phases.md)):

1. The OpenRSC world is reset to a clean state
2. Admin-tagged host accounts are deleted; new non-admin
   accounts are minted for the cohorts
3. All `local/` content stays local (gitignored already)
4. `examples/routines/` and `testdata/conformance/` are the
   canonical routine library for production hosts

Everything before that boundary is workshop. Everything after
is performance.
