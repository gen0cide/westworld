# REPL — Interactive Shell

> **Host** — an autonomous AI actor. The REPL is the primary
> **development surface** for westworld pre-launch. Humans drive
> a live host (typically delores) through DSL fragments
> interactively to author routines, validate behavior, and
> discover gaps. See [`development-workflow.md`](development-workflow.md)
> for the full model.

## Status

**Not yet implemented.** This doc is the spec. **Priority: high
for Phase 2.5 Stage 1** — REPL is the dev surface for the rest
of Phase 2.5, not a final polish step.

## Two modes

### `cradle -repl`

Login, then drop straight into the prompt. No routine runs. Use
for exploration: read the host's state, try DSL fragments,
register watchers, send a few actions by hand.

```
$ cradle -repl -username alex -dwell 30m
[INFO] logged in as alex
[INFO] entering REPL — type .help for commands, .quit to exit

>>> self.position
(120, 504)
>>> world.npcs.length
3
>>> world.npcs[0].name
"Hans"
>>> walk_to(x=124, y=501)
{val: null, err: null}
>>> self.position
(124, 501)
>>> .quit
[INFO] logging out
```

### `cradle -routine x.routine -repl-on-fail`

Run a routine. If it errors or aborts, drop into the REPL with
the routine's state preserved — env, locals, deferred-stack, the
host's connection, the world mirror. Inspect what went wrong,
optionally retry the failed call.

```
$ cradle -routine fishing.routine -repl-on-fail -dwell 30m
[INFO] routine fishing started
[ERR ] aborted at fishing.routine:18: walk_to: PATH_BLOCKED

dropping into REPL with routine state preserved.
type .help for commands, .resume to retry the failed call.

>>> self.position
(124, 503)
>>> world.locs.banks.nearest(self.position)
<placement Bank booth @ (150, 504)>
>>> walk_to(x=150, y=504)
{val: null, err: null}
>>> .resume
[INFO] resuming routine fishing at line 18
...
[INFO] routine fishing returned "ok"
```

`.resume` re-runs the failed call from the REPL state. If the
state has changed (we walked somewhere different, ate a food,
etc.), the retry runs against the new state. Routine continues
from where it aborted.

## Meta commands (prefixed `.`)

| Command | What it does |
|---|---|
| `.help` | Show commands |
| `.help <name>` | Show docs for an action/proc (eventually, when we add docstrings) |
| `.quit` | Logout cleanly and exit |
| `.state` | Pretty-print current host state — vitals, position, inventory summary |
| `.env` | Show defined locals + persistent watchers in current REPL session |
| `.routines` | List available `.routine` files in the routine library |
| `.load <path>` | Load + validate a `.routine` file without running it |
| `.run <path>` | Load + run a `.routine` file inline |
| `.resume` | (only after `-repl-on-fail`) Retry the failed call and continue the routine |
| `.cancel` | (only after `-repl-on-fail`) Give up on the routine; routine ends in aborted state |
| `.watchers` | Show currently-active `when` subscriptions |
| `.log <level>` | Set log level — `debug`, `info`, `warn`, `error` |

## Input model

- One line per submission, parsed as either an expression or a
  statement
- Expressions print their evaluated value via `.Display()`
- Statements execute and print nothing (unless they have side
  effects — e.g. action calls print the Result)
- Multi-line input via continuation prompt when a block is open

```
>>> proc square(x) {
...   return x * x
... }
>>> square(7)
49
```

Persistent env across lines — `x = 5` then `x + 1` works. Procs
defined in the REPL stay defined until quit.

Top-level `on` and `when` are accepted and register live
watchers against the actual host. They persist until quit or
explicit unregister.

## Identity

REPL-authored fragments are loaded via `ParseRoutineString` with
the logical name `<repl-line-N>` (N = monotonically increasing
input counter). Traces, logs, and the conformance system see
that name; no filename match enforcement applies (see
[`syntax.md`](syntax.md) for the loader rules).

## Safety

The REPL runs against the **live** host — every action call
sends real packets to OpenRSC. Mistakes are visible to other
players. Two cautions:

- Use `cradle -repl -username delores` to drop in as a test
  account, not as your main
- The Result of every action is printed by default, so you see
  failures immediately

The REPL respects the same resource caps as routines (op budget,
wall clock, etc.). One pathological input can't hang the host.

## Implementation notes

Build over `interp.Interpreter` + `runtime.Host`:

- Wrap the line input in `parser.Parse` first; if it's a single
  expression, evaluate it; if it's a statement or a block, run
  it
- Persistent `*Env` shared across lines
- Watchers register into the same scope-stack the interpreter
  uses for `when` blocks; `.watchers` walks the stack and prints
- `.resume` requires the interpreter to support "resume from
  pos" — re-enter the routine's `execBlock` at the failed AST
  node. Doable with a small AST-walker refactor.
- Use `liner` or `readline` for line editing + history. History
  file at `~/.westworld/repl_history`.

Estimated effort: 200–300 LOC + readline integration. The
`-repl-on-fail` part needs the resume support, which is the
biggest single piece.
