# `bounds { ... }` — region-scoped event handlers

**Status:** proposal (parser change required)

**Pitch:** a scope construct that filters location-bearing events to
a geometric region. Lets routines say "watch for items appearing in
this area" without manually re-checking coordinates in every handler.

## Syntax

The simplest form is a rectangle:

```
bounds box(x1=120, y1=640, x2=130, y2=660) {
    on item_appeared(item_id, x, y) {
        # only fires when (x, y) is inside the box
        gi = world.ground_items.by_id(item_id)
        pick_up(gi)
    }
}
```

A radius around a center point:

```
bounds circle(cx=self.position.x, cy=self.position.y, radius=12) {
    on item_appeared(item_id, x, y) { ... }
}
```

A named scenery feature (resolved via facts):

```
bounds near(scenery="furnace", radius=4) {
    on item_appeared(item_id, x, y) { ... }
}
```

A polygon (future, ordered points):

```
bounds polygon([(120, 640), (130, 640), (130, 660), (120, 660)]) {
    on item_appeared(item_id, x, y) { ... }
}
```

## Semantics

- A `bounds` block is a NEW lexical scope at the same level as `routine`.
- Inside, only `on event(...)` handlers and `proc` declarations are
  legal. (`while`, `wait`, etc. would conflict with the parent
  routine's main flow.)
- Handler arity matches `dsl/spec/events.go` as today. The filter
  applies to events that carry location params (currently
  `item_appeared`, `item_disappeared`, future `npc_appeared`,
  `damage_taken` at source position, etc.). Events without
  location pass through untouched (e.g., `death`, `chat_received`).
- **Nesting** is allowed and the intersection rule applies: an
  event must be inside all enclosing bounds to fire any contained
  handler.
- The bound expression evaluates ONCE at routine start (capturing
  e.g. `self.position` at that moment). Future iteration could
  add `dynamic` form that re-evaluates per event.

## Why this beats `if not in_region(...) { continue }`

1. **Single source of truth** — change the box and all handlers
   inside automatically respect the new region. With the if-check
   pattern, every handler has its own coord check (drift bug
   surface).
2. **Validator can warn** — handlers inside a bound that ignore
   their (x, y) params are probably a bug; the validator can
   surface "this `on` is bound but doesn't use location".
3. **Readability** — the routine's structure shows WHERE behavior
   applies, not just what.
4. **Composes with future terrain awareness** — `bounds
   walkable_from(120, 640, radius=8)` could use the pathfinder to
   exclude tiles behind walls. Pure if-checks can't easily flow
   pathfinder calls into the predicate without manual work.

## Implementation notes

- New AST node: `BoundsStmt{Shape, Body []Decl}` parallel to
  `RoutineDecl`.
- Lexer: `bounds` becomes a keyword.
- Parser: at routine top level, accept either `on`/`proc`/`when`
  declarations OR a `bounds <shape>(args) { decls }` block.
- Interpreter: `bounds` registers each contained `on` handler with
  a filter predicate. The event dispatcher calls the predicate
  with the event's args before firing.
- Shapes start as built-in callables: `box(...)`, `circle(...)`,
  `near(...)` — each returns a predicate function on (x, y).

## Out of scope (for v1)

- Polygon shape (deferred)
- Dynamic bounds (re-evaluating per event)
- Bounds that reference moving entities (e.g., "within 4 tiles
  of alex") — possible later, but requires extension to the
  filter predicate signature.
