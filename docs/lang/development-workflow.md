# Development Workflow — REPL-Driven, Delores as Collaborator

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

Once Phase 2.5 Stage 1 lands (Result/Error model, REPL,
ParseRoutineString), the inner development loop is:

```
1. Alex / Claude form a hypothesis: "we should be able to bank
   a load of iron ore by talking to the banker, then depositing
   each loot slot."
2. Drive delores through it step-by-step in the REPL:
       >>> talk_to(banker)
       >>> .state                       # what dialog options came back?
       >>> answer(2)
       >>> bank.is_open                  # did the bank UI open?
       >>> deposit(inventory.slots[0].item, inventory.slots[0].amount)
       >>> ...
3. Observe what works, what fails. Three outcomes:

   a. Everything works → save the sequence as
      examples/routines/<name>.routine and add a conformance
      test. Done.
   b. A query is missing (e.g. .state didn't show dialog
      options) → Claude adds the accessor to runtime/dsl_views.go,
      rebuilds, we retry.
   c. An action fails unexpectedly → Claude inspects the wire
      capture, identifies whether it's an action wrapper bug, a
      packet-decoder bug, or a server quirk we need to document.
      Fix, retry.
4. Once stable, commit the Go code AND the working routine.
```

Cycle time per iteration: minutes for query/accessor additions,
sometimes hours for packet-level bugs. Faster than writing a
spec upfront and grinding through it offline.

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
  **Before going live (Phase 7 scale rollout), the world is
  reset and the host accounts are recreated.** Until then,
  experimentation is unconstrained.

## What gets persisted, and where

| Artifact | Persisted to | Committed? |
|---|---|---|
| Production-quality routines | `examples/routines/<name>.routine` | yes |
| Conformance test cases | `testdata/conformance/<n>_<name>.routine` + `.expected` | yes |
| Go infrastructure (Host methods, view structs, validator changes) | `runtime/`, `dsl/` | yes |
| Doc updates triggered by discoveries | `docs/`, `docs/lang/` | yes |
| Casual REPL experimentation, scratch routines, half-baked drafts | `local/routines/` | **no — gitignored** |
| REPL session history | `~/.westworld/repl_history` | no — outside the repo |
| Server-side admin-action exploration notes | `local/notes/` | no — gitignored |

The `local/` convention: anything under it is yours/ours to
shape however we want, never goes to github. Use it for
exploration, half-finished experiments, personal scratchpads.
The repo stays the curated artifact; `local/` is the workshop.

`.gitignore` includes `local/` at repo root.

## REPL session etiquette

A few habits that make sessions productive:

- **Open with state.** First line of every session is `.state` —
  prints vitals, position, inventory summary. Reset the mental
  model of where delores is.
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

## How this changes Phase 2.5 ordering

In Phase 2.5, the order shifts from "build everything,
test at the end" to "build the minimum dev surface,
then iterate live."

Minimum dev surface (Stage 1, sequential):
1. Result/Error model + bang variants
2. Filename enforcement + `ParseRoutineString` loader
3. REPL itself (both modes)

Once Stage 1 lands, Stage 2 (query layer expansion, `when`,
`defer`, `try`/`recover`, `select`) iterates against live
delores. Every accessor we add gets tested by typing it in the
REPL. Every new control-flow construct gets tried out in a
scratch routine before its Go implementation is considered
done.

See [`../phases.md`](../phases.md) for the canonical phase
plan; this doc explains the *how*, that doc the *what*.

## Phase 2.6 / 2.7 / 2.8 implications

Same pattern applies all the way down:

- **Phase 2.6 (Knowledge ingestion)**: REPL session asks
  `>>> recall("how do I cook a lobster")` against the corpus.
  If results look bad, we know to retune chunking or
  embedding.
- **Phase 2.7 (Admin tooling)**: each new admin action is
  registered, exercised in REPL, promoted to a routine once
  stable.
- **Phase 2.8 (Live edge-case testing)**: delores plays
  through quests and skills with admin scaffolding, in the
  REPL. Failures become Go fixes or doc additions.

## Production launch boundary

The live-driven model only applies pre-launch. At Phase 7
(scale rollout):

1. The OpenRSC world is reset to a clean state
2. Admin-tagged host accounts are deleted; new non-admin
   accounts are minted for the cohorts
3. All `local/` content stays local (gitignored already)
4. `examples/routines/` and `testdata/conformance/` are the
   canonical routine library for production hosts

Everything before that boundary is workshop. Everything after
is performance.
