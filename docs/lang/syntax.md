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
host.hp < 40 { do_something() }` at the top of the routine can
call a `proc do_something() {...}` declared anywhere in the file.

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
- **One canonical name per concept** — `host.hp` only, no
  `hitpoints` or `health` aliases

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
| `extends` | Handler override (extends a parent like `host`) |
| `super` | Call into the parent handler from within an override |
| `defer` | Cleanup hook for scope exit |
| `try`, `recover` | Bang-error boundary |
| `require` | Routine preconditions block |
| `if`, `elif`, `else` | Conditional |
| `while`, `for`, `in` | Loops |
| `break`, `continue` | Loop control |
| `return` | Return from routine/proc |
| `abort` | Exit the routine with a reason |
| `wait` | Yield the routine for N seconds |
| `and`, `or`, `not` | Logical operators |
| `true`, `false`, `null` | Literals |

The bang `!` is a method-name suffix, not a keyword — see
[`actions.md`](actions.md).

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
