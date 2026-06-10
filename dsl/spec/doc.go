// Package spec is the single source of truth for the DSL's surface.
//
// Every builtin name, every event, every query-accessor path lives
// in one table here. The validator, the host bridge, the REPL, and
// doc-generation tooling all read from these tables. Adding a new
// surface item is a one-line entry in this package, not a
// scavenger-hunt across validator + bridge + docs.
//
// Tables:
//
//   - Actions ([]ActionSpec)   — every callable: actions, primitives,
//     LLM stdlib, memory stdlib, persona reads. See actions.go.
//   - Events ([]EventSpec)     — every bus event a routine can
//     `on`-handle. See events.go.
//   - Accessors ([]AccessorSpec) — every query-layer attribute path
//     hosts can read. Documentation-and-discovery source; the
//     actual Getter implementations live in the runtime/views_*.go family.
//     See accessors.go.
//
// # Dependency direction
//
// spec is a LEAF package — it imports nothing from dsl/ or runtime/.
// Validator imports spec. Bridge imports spec. This avoids any
// circular-import hazard and keeps spec usable from doc-gen tooling
// without dragging in the whole interpreter.
//
// # Consistency tests
//
// dsl/spec/consistency_test.go asserts that every spec entry has a
// matching implementation downstream (and vice versa). Drift
// between the spec and the running code is a test failure, not a
// silent doc rot. See that file for the exact invariants enforced.
//
// # Adding a new surface item
//
// 1. Action / primitive / stdlib: add a row to Actions in
//    actions.go. The validator picks it up automatically. Add the
//    Go wrapper in runtime/dsl_actions.go and wire it in
//    runtime/dsl_bridge.go.
// 2. Event: add a row to Events in events.go. The validator picks
//    up the arity check automatically. If the event has a Go
//    event.X type, also extend runtime/dsl_events.go translateEvent
//    to map it.
// 3. Accessor: add a row to Accessors in accessors.go. The runtime
//    side requires extending the relevant view's Get() switch in
//    the runtime/views_*.go family. The accessor consistency test will
//    flag spec-without-impl mismatches.
package spec
