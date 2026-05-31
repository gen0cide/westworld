# DSL and Runtime

> **NOTE (2026-05-28):** This is the **original** design doc that
> seeded the language. The DSL has since shipped end-to-end and
> evolved. For the current working spec, see
> [`docs/lang/`](lang/) — that subfolder supersedes this doc on
> specific points: event model, scoped watchers, Error/Result
> shape, filename rules, naming conventions, error codes,
> stdlib/action terminology, and the IFTTT/host-ontology framing.
>
> This doc remains useful for the *why* (design goals, rejected
> alternatives, sandbox philosophy, token economics) and as
> historical context. Where this doc and `docs/lang/` disagree,
> `docs/lang/` wins.

## Purpose

This document specifies the scripting language and the virtual machine that executes it. The DSL is the substrate every host routine runs on, the surface the LLM strategist generates code for, and the integration point for reveries, sandbox enforcement, and observability. Getting this right has more downstream impact than any other architectural decision.

## Design goals (ranked)

1. **Resilient.** LLM-generated code is assumed buggy. The runtime must never let a routine crash a host, exhaust resources, or escape its sandbox. A misbehaving routine is contained and recoverable.
2. **Predictable.** Same input + same world state → same execution trace. Reveries inject controlled randomness via a seeded RNG, never the interpreter itself.
3. **Inspectable.** Admin tools see the current AST node, local variables, frame stack, and pending events. Pausable mid-routine.
4. **Event-aware.** Routines can declare `on event_name { ... }` handlers as first-class syntax. The runtime dispatches inbound events between primary actions, cooperatively.
5. **Reverie-integrated.** `reverie.tick()` runs between every primary action automatically; routines can call it explicitly.
6. **Token-efficient.** Surface is compact enough that LLM-generated routines stay small. Common operations are short.
7. **Familiar.** LLM has seen enough Python-ish syntax that generation is reliable. Structured control flow (if/while/for) instead of goto-chains.

These are stack-ranked. When goals conflict, resilience wins.

## Why a custom interpreter (not Starlark/Lua/JS)

Considered: embedded Starlark (go.starlark.net), gopher-lua, goja (JS), tengo, yaegi.

| Concern | Embed Starlark | Custom interpreter |
|---|---|---|
| First-class event handlers | Bolted on as library | Native syntax |
| Sandbox tightness | Good defaults | Total control |
| Op-counted execution | Built-in (CallableSafe / `Steps`) | Built ourselves |
| Resource ceilings (allocation, recursion, wall-clock) | Partial; needs wrapping | Designed in |
| LLM-friendly syntax | Excellent (Python-subset) | We design it |
| Implementation size | 0 LOC for runtime, ~500 LOC for sandboxing wrappers | ~2-3k LOC for full interpreter |
| Routine serialization for cross-host sharing | Source-level only | AST-level (post-parse, validated) |
| Source position in errors | Yes | Yes — we control it |
| Long-term maintenance | External dependency | Ours |

The deciding factor is **event handler syntax as a first-class declaration**, not a callback registration. AR got this right; Starlark would force it into `on_event_name = lambda: ...` patterns that are syntactically heavier and less reliable to LLM-generate. Combined with wanting tight control over the sandbox and being willing to write ~2-3k LOC, custom wins.

**We can revisit this** if implementation exceeds ~4k LOC or if Starlark gains the event syntax we want. The DSL surface should be small enough that we could plausibly migrate routines via a transpilation step if needed.

## Language specification

### Lexical structure

- **Identifiers**: `[a-zA-Z_][a-zA-Z0-9_]*`. No `%`/`@` prefixes (unlike AR).
- **Keywords**: `routine`, `proc`, `on`, `if`, `elif`, `else`, `while`, `for`, `in`, `break`, `continue`, `return`, `abort`, `require`, `wait`, `true`, `false`, `null`, `and`, `or`, `not`.
- **Comments**: `#` to end of line. No block comments (Python-style minimal).
- **Numbers**: integers (`42`, `-7`), floats (`3.14`, `0.5`), ranges (`2.8..4.5` — used for jittered waits).
- **Strings**: double-quoted, support escape sequences (`\n`, `\t`, `\"`, `\\`). F-strings via `f"hi {name}"` for interpolation.
- **Booleans**: `true` / `false` (lowercase, Go-style).
- **Null**: `null` (sentinel for missing values, e.g., no nearby NPC of a given type).
- **Operators**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `and`, `or`, `not`. Standard precedence.
- **Block delimiters**: braces `{ ... }`. Not whitespace-significant.

### Top-level constructs

A routine file may contain:

1. **Routine-level `on` handlers** — event subscriptions active only while this routine runs (override persona-level defaults; see "Two-tier handler model" below)
2. **`proc` declarations** — pure helper procedures (no primary actions, no side effects beyond return values, no calls into routines)
3. **One `routine` declaration** — the entry point, which the host invokes (or which another routine calls)

```
# fish_at_swamp.routine

on chat_received(speaker, message) {
    # While fishing, respond chattily to greetings from friends
    if speaker.name in self.friends and message.contains("hi") {
        say(f"hey {speaker.name}!")
    }
}

proc nearest_fishing_net() {
    return world.locs.fishing_spots
        .nearest(self.position)
}

routine fish_at_swamp(food_type="shrimp") {
    require {
        wielded != null
        wielded.type == "fishing_rod"
    }
    spot = nearest_fishing_net()
    if not spot {                  # Python-style: null is falsey
        abort "no_spot_found"
    }
    walk_to(spot.x, spot.y)
    while inventory.free > 0 {     # Python-style: 0 is falsey
        if fatigue > 90 {
            abort "tired"
        }
        fish(spot)
        # reverie.tick() injected automatically here
    }
    return_to_bank()               # call another routine as a sub-routine
    return "inventory_banked"
}
```

Note this routine does NOT declare `on hp_below`, `on attacked_by`, `on level_up`, etc. Those fall through to **persona-level default handlers** — see below.

### Routine declarations

```
routine name(param1, param2=default, ...) {
    [require { precondition1; precondition2; ... }]
    [on event_name(params) { ... }]   # routine-scoped handlers
    [on event_name(params) { ... }]
    body
}
```

- Parameters may have default values.
- `require` clauses run before the body. Any falsey condition → routine aborts with `"precondition_failed: <expr>"`.
- Routine-scoped `on` handlers override persona-level defaults for the named events, *only while this routine is running*. See "Two-tier handler model" below.
- The body may contain any statement, including action calls, control flow, explicit `reverie.tick()`, and calls to other routines.
- Routines return via `return "completion_code"` or `abort "error_reason"`. Implicit fall-through returns `"completed"`.

### Routines calling routines

A routine body may invoke another routine as a sub-routine:

```
routine harvest_loop() {
    while true {
        result = fish_at_swamp()
        if result == "tired" {
            sleep_and_recover()
            continue
        }
        if result == "inventory_banked" {
            continue
        }
        abort f"unexpected: {result}"
    }
}

routine fish_at_swamp() {
    # ... fishing logic ...
    return_to_bank()
    return "inventory_banked"
}

routine return_to_bank() {
    bank = world.locs.banks.nearest(self.position)
    walk_to(bank.entrance.x, bank.entrance.y)
    banker = world.npcs.nearest_by_name("Banker")
    open_bank(banker)
    deposit_all()
    close_bank()
    return "banked"
}
```

Sub-routine call semantics:

- **Same execution context** — the sub-routine runs in the same goroutine, blocking the caller until it returns.
- **Own locals** — the sub-routine has its own local scope. Caller's locals are not visible.
- **Shared resource budgets** — the op budget, wall-clock deadline, and recursion depth limit are shared across the entire call tree. A sub-routine cannot "earn back" budget by being deep.
- **Recursion allowed but capped** — 64 total frames across all routine + proc calls. A routine calling itself works but is bounded.
- **`require` runs at every entry** — when a sub-routine is called, its preconditions check first. Failure → sub-routine returns `"precondition_failed"` to caller.
- **`abort` propagates** — an `abort` in a sub-routine propagates up: caller sees its routine return with the abort reason, and unless the caller catches it, the abort continues propagating.
- **Handlers are dynamically scoped (innermost-only)** — see below.

### Two-tier handler model

> **Status (verified against the interpreter):** ASPIRATIONAL except
> for plain handler dispatch. Today, file-level **and** routine-level
> `on` handlers register and fire (`interp/events.go`
> `registerHandlers` indexes both). What is **not** implemented: the
> persona-level default-handler tier, the `extends host` / `super()`
> override chain, and the dynamic-scoping "innermost-only override
> + de-register on sub-routine return" semantics described below.
> Those are gated on the persona tier (Phase 4). For the current,
> accurate picture see `docs/lang/events.md` (its implementation-status
> table is authoritative). The `on hp_below(0.3)` /
> `on attacked_by(other)` threshold-arg handler forms shown here were
> **removed** — use `when self.hp_fraction < 0.3 { ... }` instead.

Event handlers come from two layers:

**Persona-level default handlers** (defined once per host, in the persona; see [personas.md](personas.md)):
Every host's persona declares default handlers for every supported event. These are the "always-on reflexes" — the host's personality reactions that apply regardless of what routine is running. Examples:
- `on hp_below(0.3)` for a "cautious" persona might eat aggressively; a "reckless" persona might fight on
- `on attacked_by(other)` varies wildly by persona temperament
- `on chat_received` defaults vary by how social the persona is

Persona-level handlers are the place where "every event MUST be declared" applies (in the persona schema). It forces the persona author (LLM-generated, then optionally hand-curated) to explicitly consider every reactive surface.

**Routine-scoped handlers** (defined inside a routine, per the syntax above):
Routine-scoped handlers temporarily override persona-level defaults *for the events they name*. Other events still use persona defaults. When the routine returns (or its parent resumes it), routine-scoped handlers de-register and persona defaults take over again.

**Dynamic scoping for nested routine calls (innermost-only)**:
When routine A calls routine B:
- B's handlers register on entry, override A's same-named handlers AND persona defaults
- Events that B doesn't handle fall through to persona defaults (NOT to A's handlers)
- When B returns, B's handlers de-register
- A's handlers re-activate for the rest of A's body

This is dynamic scoping: only the innermost active routine's handlers + persona defaults are in effect. Outer routines' handlers are suspended while a sub-routine runs.

The reasoning: routines compose by *replacing* reactive behavior, not by *stacking* it. If `fish_at_swamp` defines an `on chat_received` that responds chattily, and it calls `return_to_bank`, then while walking to the bank we want either (a) `return_to_bank`'s own handler if it has one, or (b) persona default — *not* `fish_at_swamp`'s "I'm fishing" handler bleeding into a context where it doesn't make sense.

If a routine wants its reactive behavior to persist through sub-routines, it can pass a flag to the sub-routine and the sub-routine can choose to defer, or — more idiomatically — the persona-level handler should already cover the common case.

### Procedure declarations

```
proc name(param1, param2=default, ...) {
    body
}
```

- Procs are pure: no primary actions, no reverie ticks, no waiting.
- Procs are deterministic given inputs; they're used for computation only.
- Procs may call other procs.
- Procs return via `return value`.

The proc/routine distinction prevents bugs where the LLM tries to "reuse" a helper that turns out to take game time. If you want to share an action sequence, that's a routine, not a proc.

### Event handlers

```
on event_name(param1, param2, ...) {
    body
}
```

- Registered when the routine is instantiated.
- Fire when the host's event bus publishes a matching event.
- Run **between** primary actions, never mid-action.
- May call any expression / action / proc.
- Have access to the routine's locals (read-only) and module-level state.
- Cannot `return` from the routine — they execute to completion and yield back.
- May `abort` the routine they're part of (e.g., `on death { abort "died" }`).

Available events (extensible; this is the v1 set):

| Event | Parameters |
|---|---|
| `chat_received` | `speaker, message` |
| `private_message` | `speaker, message` |
| `hp_below(threshold)` | none (threshold is a registration argument) |
| `fatigue_above(threshold)` | none |
| `attacked_by(other)` | `attacker` |
| `damage_taken(amount)` | `amount, source` |
| `npc_appeared(npc)` | `npc` (entity) |
| `npc_moved(npc)` | `npc` (entity) |
| `item_appeared(item)` | `item` (ground item) |
| `inventory_full` | none |
| `level_up(skill)` | `skill, new_level` |
| `trade_request(other)` | `other` |
| `server_message(text)` | `text` |
| `coords_changed(x, y)` | `x, y` |

### Built-in actions (primary)

Actions are *blocking*: they send a packet and wait for server confirmation. Between actions, the interpreter ticks reveries and processes pending event handlers.

| Action | Effect |
|---|---|
| `walk_to(x, y)` | Walk to coord; blocks until arrival or path failure |
| `attack(target)` | Initiate combat with NPC or player |
| `eat(item)` | Eat from inventory slot or by item-ref |
| `drop(item)` | Drop inventory item to ground |
| `pick_up(item)` | Pick up nearby ground item |
| `mine(rock)` | Mine a nearby rock loc |
| `fish(spot)` | Fish at a fishing spot loc |
| `chop(tree)` | Chop tree loc |
| `cook(item, fire)` | Cook item on fire |
| `cast(spell, target)` | Cast spell on target |
| `say(message)` | Public chat (queued through chat subsystem, async) |
| `whisper(target, message)` | Private message (also queued) |
| `talk_to(npc)` | Initiate NPC dialog |
| `answer(option_index)` | Choose NPC dialog option |
| `open_bank(banker)` | Open bank via banker NPC |
| `deposit(item, count)` | Deposit from inventory to bank |
| `withdraw(item, count)` | Withdraw from bank to inventory |
| `close_bank()` | Close bank interface |
| `logout()` | Initiate logout |

### Built-in primitives (non-action)

| Construct | Effect |
|---|---|
| `wait(seconds)` | Block for `seconds`; if a range like `2.8..4.5`, picks uniformly |
| `wait_until(condition)` | Block until expression evaluates truthy (polled at next tick) |
| `reverie.tick()` | Explicitly invoke reverie injection point |
| `note(text)` | Write a note to the agent's journal (non-action; lightweight episodic write) |
| `chat.queue(message)` | Add message to chat queue (sent opportunistically by chat worker; non-blocking) |

### Built-in state accessors

These are read-only properties of the implicit `self` object and `world` view:

- `self.position` (x, y)
- `self.hp`, `self.max_hp`
- `self.prayer`, `self.max_prayer`
- `self.fatigue` (0-100)
- `self.combat_level`
- `self.skills.fishing`, `self.skills.cooking`, etc.
- `self.wielded` (item or null)
- `self.equipped` (map of slot → item)
- `inventory.free` (count of empty slots)
- `inventory.has(item_name)`, `inventory.count(item_name)`, `inventory.slots`
- `world.players` (visible players, list)
- `world.npcs` (visible npcs, list)
- `world.ground_items` (visible ground items)
- `world.locs` (typed accessors: `world.locs.fishing_spots`, `world.locs.banks`, etc.)
- `combat.engaged` (boolean), `combat.target` (entity or null)

### Built-in collection methods

```
list.length
list.contains(x)
list.first()
list.nearest(point)
list.filter(fn)        # fn is a proc identifier or single-arg lambda (`x => expr`) — both SHIPPED, see docs/lang/syntax.md "Lambdas"
list.length == 0
```

### Reserved variables

- `self` — own player state (read-only)
- `world` — world state view (read-only)
- `inventory` — inventory view (read-only; mutations happen via actions)
- `combat` — combat state (read-only)

### What's NOT in the language

- No I/O (no file, no network, no `print` — `note()` writes to journal instead)
- No goroutine / threading primitives
- No reflection
- No imports (single-file routines for v1; routine library handles sharing)
- No exceptions (errors are returned from actions as result codes)
- No classes / OOP — structs are read-only views from the runtime

Note that *arbitrary* eval is unavailable — but the standard library exposes constrained LLM-invocation primitives (`contemplate_reality()`, `exec()`) that provide controlled non-determinism. See "Standard library" below.

## Standard library

The DSL surface is intentionally small. Beyond the built-in actions and state accessors, hosts have access to a curated standard library that exposes higher-order primitives. **These are the bridge between the deterministic routine layer and the non-deterministic strategist layer** — routines write the normal path explicitly and use stdlib calls to punt to the brain when situations exceed what they know how to handle.

The stdlib is organized in five tiers, from constrained-and-cheap to powerful-and-expensive.

### Tier 1 — Decision oracles

LLM-backed decisions where the caller pre-defines the choice space. The LLM picks; the routine branches.

```
contemplate_reality(question: string = "what should I do next?") -> string
```
Sends the full context (routine state, world snapshot, recent events, relevant memory, optional RAG) to the strategist (Sonnet 4.6). Returns a short string code (the strategist's chosen action label). The routine branches on the returned code.

Rate-limited: max 5 invocations per routine instance. Exceeded → routine receives `"contemplation_exhausted"` and must use deterministic fallback. Cost: ~$0.005 per call. Latency: ~1-3s.

```
evaluate(situation: string) -> float
```
Strategist returns a single 0-1 score for a question (e.g., "how dangerous is this?"). Useful for thresholding without committing to a categorical decision. Rate-limited: max 10 per routine. Cost: ~$0.001 per call (uses Haiku).

```
decide(options: list, context: string = "") -> string
```
Strategist picks one of the explicit options provided by the caller. Reasoning is logged but not returned. Returns the chosen option verbatim. Rate-limited: max 10 per routine. Cost: ~$0.002 per call.

### Tier 2 — Generative escape hatches

LLM authors small fragments of DSL code at runtime; these run inline within the calling routine's sandbox.

```
exec(prompt: string) -> string
```
The strategist generates a DSL fragment given the prompt + full context. The fragment is parsed and validated like any routine; failures return `"exec_failed: <reason>"`. On success, the fragment executes within the caller's op budget and wall-clock budget. Returns the fragment's return value (or `"completed"`).

Rate-limited: max 1 invocation per routine instance. Exceeded → `"exec_exhausted"`. Cost: ~$0.01 per call. The generated fragment is bounded by the SAME op budget as the calling routine — no earning back budget by punting.

```
improvise(prompt: string) -> string
```
Like `exec`, but the generated fragment is not retained anywhere — pure one-shot. Useful when the situation is too unique to merit a routine. Rate-limited: max 2 per routine instance. Cost: ~$0.01 per call.

### Tier 3 — Memory & reflection primitives

Direct access to the host's memory via cognition, without going through the strategist.

```
recall(query: string, top: int = 5) -> list
```
Returns the top-K most relevant past episodes (vector search over own episodic memory). Cheap (mesa retrieval, no LLM). 

```
relation_with(name: string) -> map | null
```
Returns the host's relational record for a named player, or `null` if stranger. Cheap.

```
reflect_now() -> string
```
Triggers a synchronous reflection generation pass (normally background). Returns the new reflection's ID. Cost: ~$0.005 (Sonnet). Rate-limited: max 1 per routine.

```
note(text: string) -> void
```
Writes a journal entry without going through episode-with-importance machinery. Lightweight, no LLM. Use for "I noticed this fishing spot has been crowded all morning" style annotations.

### Tier 4 — Social primitives

Higher-level abstractions over the chat subsystem.

```
wait_for_chat(timeout_seconds: float = 30, from: string = "") -> map | null
```
Blocks until a chat message arrives (optionally filtered by sender) or the timeout elapses. Returns `{speaker, message, timestamp}` or `null` on timeout. Counts against wall clock but not op budget.

```
observe(target: entity, ticks: int = 10) -> map
```
Watches a target entity for N game ticks; returns a summary of what happened (positions visited, actions taken, chat sent, etc.). Useful for "watch what this player does before approaching them."

### Tier 5 — Persona / self introspection

Routines can query the host's own state at a higher abstraction than world state.

```
mood() -> map
```
Returns the host's current emotional state: `{calm, anxious, excited, bored, frustrated}` weights summing to 1.0. Computed from persona traits + recent events + idle time.

```
motivation() -> map
```
Returns the host's north star and current sub-goal: `{north_star: "build wealth through trading", current_focus: "amass 100k gp"}`.

```
self.persona.traits -> map
```
Read-only access to the persona's trait values (shyness, ambition, etc.). Routines can branch on persona to behave consistently in character.

### Cost ceiling per routine instance

Across all stdlib calls in a single routine invocation, the maximum LLM-driven cost is bounded:

| Tier | Max calls | Max cost contribution |
|---|---|---|
| Tier 1 contemplate_reality | 5 × $0.005 | $0.025 |
| Tier 1 evaluate | 10 × $0.001 | $0.010 |
| Tier 1 decide | 10 × $0.002 | $0.020 |
| Tier 2 exec | 1 × $0.010 | $0.010 |
| Tier 2 improvise | 2 × $0.010 | $0.020 |
| Tier 3 reflect_now | 1 × $0.005 | $0.005 |
| **Worst case per routine instance** | | **~$0.09** |

A pathological routine that maxes out every stdlib limit costs ~$0.09. At 500 hosts running ~50 routines/day each, that's a worst-case ceiling of $2,250/day. Realistic usage will be 10-20x lower — most routines never invoke stdlib LLM calls at all.

### Punt rate as a research metric

The frequency at which a routine punts to `contemplate_reality()` or `exec()` is itself an observable. Mature routines (well-developed via revision over time) punt rarely. New routines (or routines facing novel situations) punt often. This gives us a measurable signal for "is this host learning":

- Routine V1 might punt on 50% of iterations
- After several revisions, V8 might punt on 5% of iterations
- A "stagnant" routine that never reduces its punt rate is one the host isn't learning from

Logged per-routine-version in mesa's `routine_versions` table, this becomes a real research output: rate-of-learning curves per host, per cohort, per persona type. See [research-goals.md](research-goals.md).

### Mock stdlib for tests

The standard library is implemented behind interfaces. Tests use `MockStdlib` with canned responses, the same way mock brain works. No LLM calls in CI. The conformance suite includes test cases that explicitly call stdlib primitives and assert behavior.

### Why this lives in the language (not as builtin actions)

Stdlib primitives differ from primary actions in three ways:
1. **No game state mutation** — these don't send packets to the server; they consult the brain or memory
2. **No required server confirmation** — they complete based on internal state, not RSC server response
3. **Different resource limits** — wall clock matters but the game-action rate limit doesn't apply

Conceptually they sit between "pure procs" (deterministic, no I/O) and "primary actions" (game-state-mutating). They form a third category: **strategic operations**. Calling them takes wall-clock time and LLM budget, but doesn't change the RSC game state directly.

## Runtime (VM) architecture

### Pipeline

```
Routine source
    │
    ▼
Lexer       — produces tokens with source positions
    │
    ▼
Parser      — produces AST
    │
    ▼
Validator   — identifier resolution, proc/routine arity checks,
              event-handler signature checks, precondition extraction
    │
    ▼
Compiled Routine (immutable, cacheable, hash-keyed)
    │
    ▼ instantiated per invocation
Routine Instance
  - frame stack
  - locals
  - op budget (counter)
  - wall-clock deadline
  - reverie state
  - host context (read-only references to self / world / inventory)
    │
    ▼
Interpreter (AST-walking, in dedicated goroutine)
    │ outbound:   ▲ inbound:
    │ Action      │ ActionResult
    ▼ Channel    │ Channel
       Host Action Layer
       (validates preconditions, sends packet,
        waits for server confirmation, returns result)
```

### One routine instance = one goroutine

Each routine invocation runs in a dedicated goroutine. This is idiomatic Go and avoids the complexity of writing a continuation-passing interpreter.

When the routine calls a primary action:

1. Interpreter constructs an `Action` value
2. Sends to `host.actionOut` channel
3. Blocks on receive from `host.actionIn` channel
4. Host's main loop receives the Action, dispatches it (sends packet, awaits server confirmation)
5. Host's main loop sends the `ActionResult` back via `host.actionIn`
6. Interpreter resumes with the result

Between action send and action receive, the interpreter:
- Calls `reverie.MaybeFire(ctx, host)` — reverie may itself produce actions, which go through the same channel
- Drains pending event handlers — if any inbound events triggered handlers, those handlers run to completion before the next primary action

```go
type Interpreter struct {
    routine     *CompiledRoutine
    locals      map[string]Value
    frames      []*Frame
    opBudget    int
    deadline    time.Time
    seed        int64       // for reverie RNG
    host        host.Reference
    actionOut   chan<- Action
    actionIn    <-chan ActionResult
    eventQueue  []*EventHandlerInvocation
}

func (i *Interpreter) Run(ctx context.Context) (CompletionCode, error) {
    defer func() {
        if r := recover(); r != nil {
            // panic from buggy AST eval — log + return error
            err = fmt.Errorf("routine panic: %v", r)
        }
    }()
    return i.eval(ctx, i.routine.Body)
}

func (i *Interpreter) callAction(ctx context.Context, a Action) (ActionResult, error) {
    // Tick reverie before the action
    i.reverieTick(ctx)
    // Drain any pending event handlers
    if err := i.drainEvents(ctx); err != nil {
        return ActionResult{}, err
    }
    // Send action and wait
    select {
    case i.actionOut <- a:
    case <-ctx.Done():
        return ActionResult{}, ctx.Err()
    }
    select {
    case result := <-i.actionIn:
        return result, nil
    case <-ctx.Done():
        return ActionResult{}, ctx.Err()
    }
}
```

### Cancellation

The interpreter takes a `context.Context` everywhere. Cancellation propagates immediately:
- Cancels any in-flight action
- Interpreter returns `context.Canceled`
- Goroutine exits

The host can cancel for any reason: shutdown, reactive override (HP critical, combat started), admin "pause" command, wall-clock deadline exceeded.

### Op budget

Every AST node evaluation decrements an op counter. When the counter hits zero, the routine errors with `"op_budget_exceeded"`. Default budget: 1,000,000 ops per routine invocation. Tunable per routine via metadata.

Op counting prevents:
- Infinite loops in pure-computation regions (e.g., `proc compute() { while true { } }`)
- Pathological recursive procs
- Buggy LLM output that doesn't terminate

Primary actions don't count against op budget (they have their own wall-clock tracking).

### Wall-clock budget

Every routine instance has a `deadline = startTime + maxWallTime`. Default: 4 hours. Tunable per routine.

Wall-clock budget catches:
- Routines that idle forever waiting for an event that won't come
- Routines that get stuck in repeated action retries
- Long-running grinds that should have completed

When wall-clock exceeded, routine errors with `"wall_clock_exceeded"`.

### Recursion depth

Frame stack depth capped at 64. Calling a proc that calls itself recursively eventually fails with `"recursion_depth_exceeded"`. Tail-call optimization is NOT performed — explicit looping (`while`) is preferred.

### Memory caps

- Strings: max length 4096 chars per string
- Lists: max length 1024 elements
- Number of distinct local variables: max 256
- Total interpreter-allocated bytes: tracked, capped at 1 MB per routine instance

Exceeding these → routine errors with descriptive message. The cap is loose enough that no legitimate routine should hit it; the goal is preventing pathological inputs from chewing memory.

### Action rate limit

Routines can't spam actions. Per host, max action rate is one per `min_action_interval` (default 100ms — RSC tick is 640ms, this gives some safety margin). Calling another action sooner than this blocks the interpreter (waits the remainder of the interval).

This prevents:
- Routines that exploit the action layer's buffering to overwhelm the server
- "Pixel-perfect" timing that detection systems could flag

### Panic recovery

`defer recover()` at the top of every interpreter goroutine. If anything panics:
1. Capture stack trace
2. Log with full context (routine name, current AST node, locals snapshot)
3. Return `RoutinePanic` error
4. Host pauses the routine, marks it failed, surfaces to delos for review

A panicking routine never propagates up to crash the host process.

### Error semantics

Actions return typed result codes. Routines explicitly handle them:

```
result = walk_to(120, 504)
if result.success {
    # ok
} else {
    abort f"walk_failed: {result.reason}"
}
```

Common result codes per action:
- `success` — operation completed
- `path_blocked` — walk couldn't reach destination
- `out_of_range` — target not in range
- `inventory_full` — couldn't pick up / receive
- `target_died` — combat target died before action completed
- `interrupted` — reactive override took precedence
- `server_rejected(reason)` — server said no, with text

Routines that don't check results get sensible defaults: failures cause an abort. Routines can explicitly opt into "try and continue" by checking results.

### Cooperative event dispatch

Event handlers run **between primary actions**, never mid-action. The flow:

1. Routine calls `fish(spot)`
2. Interpreter's `callAction` enters
3. Reverie tick fires (may inject actions, which the routine doesn't see)
4. **Pending event handlers drain**: any events that fired since the last action checkpoint run their handlers now, to completion
5. Action is sent to host
6. Wait for completion
7. Action returns; interpreter resumes
8. Next routine statement

Event handlers cannot recursively schedule themselves — if `on chat_received` fires and the handler itself triggers another `chat_received` (somehow), the second one queues for the next tick rather than running immediately. Prevents runaway dispatch.

Event handlers have a per-invocation op budget (smaller than the main routine's — default 10,000 ops). Exceeded → handler errored, but the main routine continues.

## Reverie integration

Between every primary action call, the interpreter automatically calls `reverie.MaybeFire(ctx, host)`. The reverie module:

1. Inspects emotional state (computed from persona traits + recent events + idle time)
2. Rolls dice (seeded RNG, deterministic given seed)
3. Selects a reverie to fire (or none — most ticks fire nothing)
4. Executes the reverie via the same action channel as routines

This means reveries are invisible to routine authors. A routine like:

```
while inventory.free > 0 {
    fish(spot)
}
```

…is naturally peppered with idle wanders, glances at passing players, spontaneous chats, etc. — all from the reverie module, all between `fish()` calls.

Routines may also call `reverie.tick()` explicitly at points where humans would naturally pause:

```
walk_to(bank_door)
reverie.tick()       # explicit pause and look around moment
open_bank(banker)
```

## Static validation (parse-time)

Before a routine can run, the validator checks:

- All identifiers resolve (no calls to undefined procs/routines)
- All event handlers have valid event names + correct arity
- All `require` clauses are pure expressions (no actions)
- All `proc` bodies contain no primary actions
- All loops have at least one yield point (action, wait, or reverie.tick) — prevents tight infinite loops at parse time
- All paths from routine entry either return, abort, or are guaranteed to reach an action
- Estimate op-budget upper bound from AST size — refuse routines too large to plausibly run

Validation errors include source position and a human-readable explanation. LLM strategist retries on validation failure with the error in context.

## Persistence

Routines are stored in mesa as DSL source text. On load, the interpreter parses + validates + compiles to an in-memory AST. Compiled routines are cached by source hash — calling the same routine repeatedly doesn't re-parse.

The compiled AST is NOT serialized to disk in v1. Source is the only persistent form. If parse/validate becomes a bottleneck (unlikely), we can serialize compiled ASTs later.

## Testing strategy

Four tiers, run on every commit.

### Tier 1: Unit tests
- Lexer: every token type, edge cases, malformed input
- Parser: every grammar production positive/negative cases
- Validator: every static check positive/negative
- Interpreter: every AST node type, locally scoped tests

### Tier 2: Conformance suite
- Directory of `.routine` scripts under `testdata/conformance/`
- Each script has a paired `.expected` file with expected execution trace
- Tests run the script against a mock action API and compare traces
- Covers: every action verb, every event handler type, every control flow construct, every error case

### Tier 3: Property tests (go-quickcheck or rapid)
- Property: any string either parses cleanly or returns a validation error with source position; never panics
- Property: any parsed routine either completes, errors with a typed reason, or hits a resource limit; never panics
- Property: deterministic execution — same seed + same world snapshot → same action sequence
- Property: cancellation always returns within 100ms of context cancel

### Tier 4: Fuzz tests
- `go test -fuzz=` on the lexer and parser inputs
- Random byte sequences should never panic
- Random valid-looking ASTs (via generator) should evaluate without panic

### Differential tests
- Run the same routine with reveries-disabled vs reveries-enabled-deterministic-seed
- Structural properties (e.g., "completes within N actions") should hold in both modes
- Reveries-enabled has *additional* actions but never *missing* core actions

### Mock action API
The test harness includes a `MockActionAPI` that:
- Accepts actions on an in-memory channel
- Returns canned results (configured per test)
- Records the action sequence for trace comparison
- Simulates server-side validation (e.g., walking to an unwalkable tile returns `path_blocked`)

This lets us test routines without standing up a real OpenRSC server. Real-server tests are integration tests, run nightly.

## Observability hooks

The runtime exposes hooks for delos web UI:

- Live current AST node + source line — admin sees "routine X is on line 47, in the `while` loop at index 3"
- Live locals snapshot — admin sees current variable values
- Action history — last N actions sent
- Event handler invocation history
- Reverie fire history
- Op budget remaining + wall clock remaining
- Pause / resume control — interpreter checks a pause flag between AST nodes

When paused, the interpreter holds at the current node. Admin can inspect, modify locals if desired (debug-only feature), then resume.

## Examples

### Minimal routine with persona-fallback handlers

```
# fish_at_port_sarim.routine
# This routine defines no routine-scoped handlers — it relies entirely on
# persona-level defaults for hp_below, attacked_by, level_up, etc.
# Only the routine-specific overrides are declared here.

on chat_received(speaker, message) {
    # While fishing, respond chattily ONLY to friends saying hello
    if relation_with(speaker.name) and message.contains("hi") {
        say(f"hey {speaker.name}")
    }
}

proc nearest_fishing_net() {
    return world.locs.fishing_spots.nearest(self.position)
}

routine fish_at_port_sarim() {
    require {
        self.skills.fishing >= 1
        inventory.has("small_net")
    }

    spot = nearest_fishing_net()
    if not spot { abort "no_fishing_spot" }

    walk_to(spot.x, spot.y)

    while inventory.free > 0 {
        if fatigue > 90 { return "tired" }
        fish(spot)
    }

    return_to_bank()    # call another routine
    return "banked"
}
```

### Routine using strategist escape hatches

```
# wilderness_pk.routine
# A routine that doesn't know exactly what to do in combat — punts to brain.

on attacked_by(other) {
    # Override the persona's default — we're explicitly in PK mode and want
    # to engage rather than flee (the persona default for "cautious" is flee).
    if other.combat_level <= self.combat_level + 3 {
        attack(other)
    } else {
        # Out of our league — use persona default (flee)
    }
}

routine pk_hunt() {
    walk_to(125, 142)    # wilderness lvl 47 area
    
    while not fatigue > 90 {
        targets = world.players.filter_by_combat_range(self.combat_level - 3, self.combat_level + 3)
        
        if targets.length == 0 {
            # No targets — wander or contemplate
            choice = decide(["wander_deeper", "wait_here", "retreat"], "no targets in range")
            if choice == "wander_deeper" {
                walk_to(self.position.x, self.position.y - 10)
                continue
            }
            if choice == "retreat" { return "no_targets" }
            wait(15..30)
            continue
        }
        
        target = targets.first()
        
        # Decide if this target is worth engaging — the LLM weighs context
        # we don't want to hardcode (equipment, reputation, etc.)
        risk = evaluate(f"how risky is attacking {target.name} at combat {target.combat_level}?")
        if risk > 0.7 {
            note(f"skipped {target.name} — too risky")
            continue
        }
        
        attack(target)
        # combat resolves; if we die, persona's on_death handler fires
    }
    
    return "tired"
}
```

These examples show:
- **Persona-level handlers** carry HP-management, generic combat reactivity, level-up celebration, etc. (defined in the persona, not the routine)
- **Routine-scoped handlers** override persona defaults only for the events that the specific activity needs to handle differently
- **Sub-routine calls** (`return_to_bank()`) compose activity rather than inlining everything
- **Decision oracles** (`decide`, `evaluate`) consult the strategist when the routine doesn't want to hardcode a choice
- **Reverie ticks** are invisible between every action — the routines never reference them, but every fishing iteration and every walk step is naturally peppered with glances, idle wanders, persona-driven small chat

## Resolved language design choices

These were open during the spec phase and now have decisions baked in:

- **Pure procs plus single-arg lambdas** *(superseded — this row originally read "pure procs only, no lambdas/closures")*: single-expression lambdas `x => expr` later shipped (Phase 2.5) for `filter`/`map`/`find`/`nearest` predicates and close over the enclosing env; named procs are still the idiom for multi-step logic. See `docs/lang/syntax.md` "Lambdas".
- **Python-style truthiness**: `false`, `null`, `0`, `0.0`, `""`, `[]`, empty maps are falsey; everything else truthy. LLM-friendly, compact in routines.
- **Event handler order**: arrival order. No priority handlers. Safety-critical reactivity (HP-low etc.) is handled by persona-level defaults that always fire.
- **Op/budget uniformity**: same caps for every routine regardless of cohort/persona. Behavioral differentiation comes from persona-driven routine *content*, not artificial budget advantages. This is the substrate for differential learning as a research observable.
- **Two-tier handler model**: persona-level defaults (every event declared, host-wide reflexes) + routine-scoped overrides (innermost-only, override defaults only while this routine runs). Sub-routines push their own handler scope.
- **Routines can call routines**: sub-routine calls share op budget, wall-clock, and recursion depth with caller. Handlers are dynamically scoped (innermost-only).
- **Standard library with LLM escape hatches**: `contemplate_reality()`, `exec()`, `decide()`, `evaluate()`, etc. — bridge between deterministic routines and the strategist for situations exceeding the routine's knowledge.

## Open implementation questions

These get resolved during Phase 2 implementation:

- **Action async-cancel semantics**: if a routine is canceled mid-action, should the action complete on the server side anyway? Probably yes — we send the packet, can't unsend it.
- **Routine versioning at DSL-grammar evolution**: when DSL grammar changes, existing routines stored in mesa need either migration or version-tagged interpretation. Probably tag each routine version with its DSL spec version; interpreter handles old versions until lazy-migrated.
- **exec() generated fragment storage**: do we save successful exec fragments anywhere (for replay/analysis), or are they truly ephemeral? Probably save to `routine_versions` with origin `"exec:<parent_routine>"` for research observability.
- **Persona-level handler updates**: when a persona's reflexes change (e.g., agent learns from being scammed and becomes warier), does the persona's handler set update? Probably yes; this is also versioned in mesa.

## Open language-design questions for Phase 2 paired session

- Range syntax: `2.8..4.5` for inclusive range, or `..=` like Rust? Probably `..` is fine for a small DSL.
- F-string syntax: `f"hi {name}"` or `"hi $name"` shell-style? Python f-strings are cleaner.
- Truthiness rules: only `false` and `null` are falsey? Or also `0`, empty strings, empty lists? Python-style "many things are falsey" is friendlier; strict-true-only is more predictable. Lean toward Python-style.
- Equality: `==` for value equality everywhere? Or object identity vs value equality? Probably value equality only — we don't have user-defined types.
- Pattern matching / switch: do we need it? Probably not for v1.

## When this gets resolved

Phase 2 (DSL runtime build-out). Approximate effort: 2-3k LOC Go for full lexer + parser + validator + interpreter + sandbox + tests. Estimated 2-3 weeks of focused work with agent assistance.

## Implementation order

1. Lexer + parser (~600 LOC) + unit tests
2. Validator (~300 LOC) + tests
3. AST evaluator core: expressions, control flow, locals (~500 LOC) + tests
4. Mock action API + interpreter loop integration (~300 LOC) + tests
5. Op budget + wall clock + memory caps (~200 LOC) + tests
6. Event handler dispatch (~300 LOC) + tests
7. Reverie tick integration (~200 LOC) + tests
8. Conformance suite + property tests (~500 LOC test code)
9. Observability hooks for delos (~200 LOC)

Total: ~3k LOC of runtime + ~1k LOC of tests. Each step is independently testable; we don't need all of it working before validating early steps.

## What this enables

With this runtime in place:

- The brain can generate routines and they execute predictably or fail with clear errors
- Routines are sandboxed: a bug in LLM-generated code is contained
- Reveries weave in invisibly, making "running a routine" indistinguishable from organic play
- Event handlers let routines be situationally aware without polling
- Admin tools can inspect, pause, and resume any routine mid-flight
- The whole runtime can be regression-tested against a conformance suite as we evolve

This is the substrate everything above it (cognition, brain, reveries) depends on. It deserves the engineering investment.
