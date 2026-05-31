# Syntax — Surface Form

This doc covers everything below the level of state/events/actions:
how files are laid out, what counts as an identifier, what
comments look like, what the keywords are.

## File structure

A `.routine` file has three kinds of top-level declarations:

```
# top-level handlers — active for the whole routine run
on chat_received(speaker, message) { ... }

# helper procs — pure, callable from the routine body
proc nearest_spot() { ... }

# the entry point — exactly one per file, required
routine fish_at_port_sarim(rod_type) {
    require { ... }
    # body
    return "ok"
}
```

Order within the file doesn't matter for resolution — the parser
reads the whole file, the validator builds the proc and handler
symbol tables from the entire file, and the interpreter binds
every proc into the env before the routine body runs. So `when
self.hp < 40 { do_something() }` at the top of the routine can
call a `proc do_something() {...}` declared anywhere in the file.

### `extends "parent.routine"` — file-level inheritance

**Status: EXECUTED (disk-loaded routines only).** The merge runs in
`runtime/dsl_bridge.go` `mergeExtends`, before validation. It is
**not** available to string-loaded routines (REPL / `exec()` /
`improvise()`), which have no base directory for path resolution —
`ParseRoutineString` rejects a routine containing `extends`. The
per-handler `extends host` / `super()` override chain (below) is a
**separate, not-yet-implemented** construct — don't conflate the
two.

A routine file can pull in procs + on-handlers from one or more
parent files:

```
# child.routine
extends "common/banking.routine"
extends "common/eating.routine"

routine fish_at_swamp() { ... }
```

Path resolution is relative to the child file's directory.
Semantics:

- **Handlers**: additive. Parent's `on chat_received` and child's
  `on chat_received` both fire (parent first, then child).
- **Procs**: child overrides parent on name collision. (Chaining
  via `super()` is deferred to the per-handler `extends host`
  form — see `events.md`.)
- **Bounds**: additive, parent-first.
- **Parents must be libraries** — a parent file must declare procs
  and/or handlers but no `routine ...`. Loading rejects a parent
  with its own routine declaration.
- **Cycles are rejected**. `A extends B extends A` errors at load.

`extends` is only supported by `ParseRoutineFile` (disk-loaded
routines, where a base directory exists). `ParseRoutineString`
rejects routines containing `extends`.

## Filename ↔ routine name

**Required to match.** When loaded from a file path, the validator
checks that `filepath.Base(path)` without the `.routine` extension
equals the declared routine name. Mismatch is a validation error.

```
# File: fish_at_port_sarim.routine
routine fish_at_port_sarim() { ... }      # ✓ matches

# File: fish_at_port_sarim.routine
routine fishing() { ... }                 # ✗ validation error
```

Two loaders, two rules:

- `ParseRoutineFile(path)` — loads from disk, enforces match
- `ParseRoutineString(logicalName, source)` — loads from memory
  (host-authored transient routines via `exec()` / `improvise()` /
  REPL), uses the caller-supplied logical name as the identity

The host-generated case: when the strategist `exec()`s a fragment,
the runtime mints something like `exec:<short-hash-of-source>` as
the logical name so traces stay legible. REPL fragments get
`<repl-line-N>`.

This pairing is also load-bearing for future namespacing:
`import "lib/banking/deposit_all"` resolves to a file whose
routine is named `deposit_all` at path `lib/banking/deposit_all.routine`.
Path and name move together.

## Comments

Single-line `#` to end-of-line. That's it for now.

```
# a single-line comment

# multi-line notes use chained # lines, Python-style.
# the lexer just consumes everything from # to \n.

routine x() {              # trailing comments work too
    return 1
}
```

No block comments (`/* ... */`). No docstring convention yet — if
we add one later, the rule will be "the contiguous `#` lines
immediately above a declaration are its doc, extractable by
tooling."

## Identifiers

`snake_case`. ASCII letters, digits, underscores. Must start with
a letter or underscore. Convention:

- **Booleans** prefix `is_` — `is_busy`, `is_in_combat`, `is_friend`
- **Maximums** prefix `max_` — `max_hp`, `max_prayer`
- **Counts** no prefix — `free`, `length`, `level`
- **Verbs** in present tense — `eat`, `walk_to`, never `eating`
- **Error codes** SCREAMING_SNAKE — `PATH_BLOCKED`, `INVENTORY_FULL`
- **One canonical name per concept** — `self.hp` only, no
  `hitpoints` or `health` aliases (game-state queries live on
  `self`, not `host`; see [`overview.md`](overview.md))

## Keywords

| Keyword | Purpose |
|---|---|
| `routine` | Entry-point declaration |
| `proc` | Helper function declaration |
| `on` | Persistent event handler (file-top) |
| `when` | Block-scoped state-transition watcher |
| `select` | Block until one of several conditions fires |
| `timeout` | Time-bounded case inside `select` |
| `becomes`, `changes`, `increases`, `added`, `removed` | Subscription qualifiers |
| `extends` | File-level inheritance (`extends "parent.routine"`) — merges parent procs + on-handlers (**implemented**, disk-load only); the per-handler form `on ev() extends host` for handler override chains is **planned** (Phase 4) |
| `super` | (reserved) Call into the parent handler from within an `extends host` override — **not yet implemented**; currently a validate-time error wherever used |
| `defer` | Cleanup hook for scope exit |
| `try`, `recover` | Bang-error boundary |
| `require` | Routine preconditions block |
| `if`, `elif`, `else` | Conditional |
| `while`, `for`, `in` | Loops |
| `repeat`, `until` | Retry-with-timeout block (see below) |
| `break`, `continue` | Loop control |
| `return` | Return from routine/proc |
| `abort` | Exit the routine with a reason |
| `wait` | Yield the routine for N seconds |
| `and`, `or`, `not` | Logical operators |
| `true`, `false`, `null` | Literals |

The bang `!` is a method-name suffix, not a keyword — see
[`actions.md`](actions.md).

## `repeat { ... } until <cond> timeout <expr>` — retry with timeout

A do-while loop bounded by wall-clock time. Use this for "click
the banker, see 'Please wait', try again" patterns or any other
poll-with-backoff shape where blind retries could spin forever.

```
repeat {
    open_bank(banker)
    wait 1
} until world.bank.is_open timeout 10s
```

Semantics:

- **Body runs at least once** (do-while, not while).
- After each iteration the condition is evaluated; if truthy,
  the loop exits.
- If the condition is still falsy when wall-clock elapsed
  exceeds `timeout`, the loop also exits — the caller is
  expected to re-check the predicate to find out whether it
  succeeded or timed out.
- `break` / `continue` inside the body behave normally.

Two validator-enforced guards:

1. **Timeout is mandatory.** A `repeat { ... } until <cond>`
   without an explicit `timeout` is a validation error —
   "accidentally infinite retry" is the worst kind of bug, so
   the language refuses to compile it.
2. **Not allowed inside event handlers.** Like `wait` and
   `wait_until`, `repeat ... until` can yield, and handlers must
   not yield.

Timeout expressions are normal scalar expressions in **seconds**
(use floats for sub-second values: `timeout 0.5`). Time-unit
suffixes (`30s`, `2m`) are not yet supported — write `30` or
`120` and document with a comment if needed.

## Lambdas

**Status: EXECUTED.** Single-arg lambdas (`IDENT => expr`) parse
(`dsl/parser/parser.go` `parseLambda`), validate, and run
(`dsl/interp/interp.go` `lambdaCallable`; tests in
`dsl/interp/lambda_test.go`). They close over the enclosing env and
are the idiomatic predicate for `filter` / `map` / `find` /
`nearest`. The **multi-arg** form `(IDENT, IDENT) => expr` is the
only part still planned — the parser accepts a single bare `IDENT`
before `=>` today.

Single-expression anonymous functions for use with `filter`,
`map`, `find`, and similar collection predicates. Grammar:

```
IDENT => expr                # single arg — EXECUTED
(IDENT, IDENT) => expr       # multi-arg — planned, not yet parsed
```

Examples:

```
weak_npcs = world.npcs.filter(n => n.combat_level <= 30)
my_lobster = inventory.find(s => s.item.name == "Lobster")
nearest_strong_player = world.players
    .filter(p => p.combat_level > self.combat_level)
    .nearest(self.position)
```

A lambda's body is **one expression** — no statements, no
control flow inside. For multi-step logic, use a named `proc`
and pass it by name:

```
proc is_worth_killing(n) {
    if n.combat_level > self.combat_level * 1.5 { return false }
    if n.is_attackable == false { return false }
    return true
}

routine grind() {
    target = world.npcs.filter(is_worth_killing).nearest(self.position)
    ...
}
```

Lambda parameters are fresh locals in the lambda's body; they
don't leak. The body is a normal expression — anything legal
in expression position (including stdlib calls in non-
subscription contexts) is fine.

The validator treats a lambda body as a function-call argument
context — same subscription-safety rules as the calling site.
A lambda passed to `world.npcs.filter(...)` in a `when` predicate
must itself be subscription-safe.

## Runtime type errors

Some operations are dynamic-typed and can fail at runtime when
operand types don't make sense. These are **routine errors**
(turn into ResultErrored), not compile errors:

- Mixed-type comparisons that aren't numeric promotion:
  `"abc" < 5` → runtime error.
- `for x in <non-iterable>`: iterating a non-list, non-range,
  non-string value → runtime error.
- Float ranges in `for ... in low..high` iteration: only int
  ranges are iterable. `for n in 1.5..2.5 { ... }` → runtime
  error. Float ranges work fine in `wait low..high` (random
  pick within range).
- Division/modulo by zero → runtime error.
- Field access on a value that doesn't expose that field
  (`self.banana`) → runtime error.

## Procs and reserved-name access

Procs are **pure helpers**: the validator forbids primary
actions, `wait`, and `wait_until` inside them. But procs CAN
read reserved names (`self`, `host`, `world`, `inventory`,
`combat`) — those are queries, not mutations:

```
proc nearest_attackable_npc() {
    return world.npcs
        .filter(n => n.is_attackable)
        .nearest(self.position)
}
```

This proc reads `world.npcs` and `self.position`. Both are
queries; both are allowed. If the proc tried to call
`walk_to(...)` or `say(...)`, the validator would reject it.

## Literals

- Integers: `42`, `-7`
- Floats: `3.14`, `0.5`
- Strings: `"hi"` with escapes `\n`, `\t`, `\"`, `\\`
- F-strings: `f"hi {name}, you have {gold} gp"`
- Bools: `true`, `false`
- Null: `null`
- Lists: `[1, 2, 3]`
- Range: `2.8..4.5` (used in `wait`) or `1..10` (iteration)

## Operators (low → high precedence)

| Precedence | Operators |
|---|---|
| 1 | `or` |
| 2 | `and` |
| 3 | `not` (unary) |
| 4 | `==`, `!=` |
| 5 | `<`, `<=`, `>`, `>=` |
| 6 | `+`, `-` |
| 7 | `*`, `/`, `%` |
| 8 | unary `-`, `+` |
| 9 | `..` (range) |
| 10 | postfix `()` call, `.field`, `[idx]` |
| 11 | primary (literal, ident, paren, list) |

## Truthiness

Python-style "many falsey": `false`, `null`, `0`, `0.0`, `""`,
`[]` are falsey. Everything else is truthy. Entities (Getter
values) are truthy if non-nil.

## Equality

Structural / value-equal. `==` compares values, not references.
Numeric kinds cross-promote: `1 == 1.0` is `true`. Lists and maps
compare element-wise.
