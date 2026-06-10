# REPL — Interactive Shell

> **STATUS: BUILT** (verified 2026-06-10 against branch `tidy/structure-and-docs`,
> HEAD `0bfa818`). Two layers: the persistent interpreter **Session**
> (`dsl/interp/repl.go`) and the stdin shell over it (`runtime/repl.go`).
> `.load` / `.run` are BUILT. The deferred niceties (`.routines`, `.watchers`,
> `.log`, multi-line input, readline+history, `-repl-on-fail`/`.resume`) live in
> [TODO.md](../TODO.md) **DSL-27**. The daemon-era equivalent is the debughttp
> control plane's `POST /eval` — same Session machinery over HTTP, plus an
> idle event pump the stdin REPL lacks.

> **Host** — an autonomous AI actor. The REPL was the primary development
> surface pre-conductor and is still the fastest way to poke a live host by
> hand. See [`development-workflow.md`](development-workflow.md) for the
> surrounding loop.

## Entry points

- **`cmd/host -repl`** — the modern single-host runner. The REPL is also the
  fallback director: a non-headless host with no goal, no routine, and no
  persona-derived goal drops into it (`runtime/runhost.go:233`).
  Daemon hosts set `Headless` and error out instead of blocking on stdin.
- **`cmd/legacy-cradle -repl`** — the original entry (`cmd/legacy-cradle/main.go:105`).
  In `-repl` mode INFO logs are suppressed so the prompt stays clean; `-v`
  brings DEBUG back (`main.go:138-148`).
- **HTTP sibling** — `debughttp` serves the same Session (logical name
  `<debug-http>`) over `POST /eval` (body = one DSL fragment; response =
  `{ok, value, is_expression, error, events_before}`). Standalone via
  `cmd/host -debug-addr` or `cmd/legacy-cradle -debug-http`; under the cradle
  daemon (`cmd/cradle-server`) every live host gets it at
  `POST /api/hosts/{name}/debug/eval` (`cradle/api.go`). Prefer this for
  scripted driving — see `debughttp/server.go` for the full endpoint set.

## Session model

`dsl/interp/repl.go` — a `Session` is a long-running interpreter context:
persistent env, registered procs/handlers, and one budget across all `Eval()`
calls. It inherits the host's reserved roots (`self`, `world`, `inventory`,
`combat`, `trade`, `bank`, `duel`, `magic`, `prayer`, `shop` —
`runtime/dsl_bridge.go:257-266`) and every spec-registered action callable.
Action calls from the REPL go through the **same pearl Gate** as routine
actions (state-mutating `PrimaryAction`s are wrapped when a pearl engine is
wired — `dsl_bridge.go:287-289`), and send real packets.

`Session.Eval` dispatch order per input:

1. Top-level `on event(params) { body }` → parsed via `parser.ParseOnHandler`
   and registered in the session's OnHandlers map (persists until quit).
2. Expression parse (`parser.ParseExpr`) — most inputs are queries like
   `self.hp`; the value prints via `.Display()`.
3. Statement parse (`parser.ParseStmt`) — assignment, control flow, bare
   action calls. Statements print nothing; an action-call *expression* prints
   its CallResult (`{val: ...}` / `{err: ...}`).

Escaping control flow (`return`/`abort`/`break`/`continue` typed at the
prompt) and Go panics are caught and surfaced as errors — the session
survives parse errors, runtime errors, and panics alike.

Budget is session-wide, not per-line: `NewSession` takes one fresh budget from
the interpreter's Caps (`DefaultCaps`: 1,000,000 AST-node evals, 4h wall clock
— `dsl/interp/caps.go:42-43`). One pathological input can't hang the host (op
budget kills runaway loops), but a marathon interactive session will
eventually hit the wall clock unless the caller raises `Caps`.

## Meta commands (prefixed `.`)

Built — dispatched in `runtime/repl.go:handleMeta`, never touch the interpreter:

| Command | What it does |
|---|---|
| `.quit` / `.exit` / `.q` | Exit the REPL |
| `.help` | Command list |
| `.help <name>` | Spec docs for an action or event — kind, arity, params, doc summary, bang-eligibility, NYI marker (`spec.ByName` / `spec.EventByName`) |
| `.state` | Host vitals snapshot — position, hp, prayer, fatigue, combat level, inventory used/free, visible NPC count |
| `.env` | Locals defined in this session (filters `self`/`host`/`world`/`inventory`/`combat`; the newer namespace roots `trade`/`bank`/`duel`/`magic`/`prayer`/`shop` currently leak into the listing — `runtime/repl.go:isReservedDSLName`) |
| `.builtins` | Every registered DSL callable, grouped by spec kind, stubs marked |
| `.accessors` | Known query-layer accessor paths from the spec |
| `.events` | Dump the **live recent-events buffer** — last chat/PM/damage/server-message/NPC-dialog observed |
| `.events spec` | The language's event catalog — `on`-handler signatures |
| `.load <path>` | Parse a `.routine` file and bind its procs + handlers into the session, without invoking the entry point |
| `.run <path>` | `.load` + invoke the routine's entry point; prints the Result kind and value |

Not built: `.routines`, `.watchers`, `.log <level>`, `.resume`/`.cancel` —
[TODO.md](../TODO.md) DSL-27. Note the old `.watchers` blocker ("`when`
watchers don't exist yet", task #47) is dead — when-watchers shipped long ago
(`dsl/interp/watchers.go`); only the listing meta-command is missing.

## Input model

**One line per submission.** The stdin loop is a plain `bufio.Scanner`
(`runtime/repl.go:66-71`) — no continuation prompt, no brace-depth tracking,
no readline, no history file. A `proc f() {` opener errors. Multi-line
fragments reach a session two ways only:

- `.load` / `.run` a `.routine` file, or
- POST a multi-line body to debughttp `/eval` (the HTTP path takes the whole
  body as one fragment, newlines included).

Procs cannot be declared at the prompt at all — `parser.ParseStmt` has no
`proc` case (`dsl/parser/parser.go:359`); declarations are file-level syntax.
Procs bound via `.load` become first-class callables in the session env and
persist until quit. Plain locals persist too: `x = 5` then `x + 1` works.

## Events and watchers at the prompt

- **`on` handlers persist** (registered by `Session.Eval` case 1) — but the
  stdin REPL has **no idle pump**. Pending bus events are dispatched only
  *while an input is evaluating*: around yielding action calls
  (`dsl/interp/interp.go:1163-1196`), during `wait` (chunked, event-pumping —
  `interp.go:595-604`), and inside `select`'s poll loop. Sitting at the prompt,
  nothing fires. The debughttp Session is better here: `Activate` runs a 200ms
  `PumpEvents` ticker (`debughttp/server.go:196-210`), so handlers registered
  over `/eval` fire while idle.
- **Top-level `when` does not work.** `when` is block-scoped by design
  (watchers unregister when the enclosing block exits); typed bare at the
  prompt there is no open watcher frame and the registration is silently
  dropped (`dsl/interp/watchers.go:75-86`). A `when` inside a one-line
  `{ ... }` block dies when the block exits on the same beat. `on` is the
  persistent reactive form for interactive use.

## Example session

```
$ host -username delores -repl
westworld REPL — type .help for commands, .quit to exit
>>> self.position
(120,504)
>>> self.hp
35
>>> walk_to(x=124, y=501)
{val: null}
>>> x = inventory.free
>>> x
12
>>> .quit
[INFO] exiting REPL
```

## Identity

The stdin session's logical name is **`<repl>`** (`runtime/repl.go:45`); the
HTTP session's is **`<debug-http>`** (`debughttp/server.go:195`). That name is
what parse errors, traces, and diagnostics carry. There is no per-line
`<repl-line-N>` counter — `ParseRoutineString` accepts any logical name and no
filename-match enforcement applies to transient fragments
(`runtime/dsl_bridge.go:165-200`).

## Safety

The REPL runs against the **live** host — every action call sends real
packets to OpenRSC, gated by the host's own pearl policy exactly as routine
actions are. Mistakes are visible to other players, so drop in as a test
account, not your main. Every action-call expression prints its CallResult,
so failures surface immediately; the session-wide op budget and wall clock
(above) bound runaway inputs.

## Deferred work

DESIGN, not built: **`-repl-on-fail`** — run a routine and, on error/abort,
drop into the REPL with the routine's env preserved; `.resume` re-enters the
failed call, `.cancel` gives up. No such flag exists in any binary; it needs
resume-from-AST-node support in the interpreter. It remains the live spec for
that feature, tracked with the rest of the REPL niceties.

All open REPL items (multi-line input, `-repl-on-fail`/`.resume`/`.cancel`,
`.routines`, `.watchers`, `.log`, readline + history) are tracked in
[docs/TODO.md](../TODO.md) under **DSL-27** — the daemon-era `/eval` covers
much of the need, so they sit at low priority.
