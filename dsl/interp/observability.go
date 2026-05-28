package interp

import (
	"time"

	"github.com/gen0cide/westworld/dsl/ast"
)

// Hooks bundle optional callbacks that observe routine execution.
// Used by the conformance suite to capture a trace, by the
// delos/mesa pipeline to emit structured telemetry, and by tests
// to assert specific events.
//
// All fields are optional — leave nil to skip the hook. The
// interpreter calls the hooks on the calling goroutine, so they
// must not block on long-running operations (use buffered channels
// or defer expensive work).
type Hooks struct {
	// OnRoutineStart fires once per RunRoutine call, before any
	// statement executes.
	OnRoutineStart func(routine string)

	// OnRoutineEnd fires once per RunRoutine call, after the body
	// has terminated, with the resolved result.
	OnRoutineEnd func(routine string, r Result, elapsed time.Duration)

	// OnAction fires before each yielding (action) callable invocation.
	// Args are the resolved positional arguments. Named args are not
	// included to keep traces compact; callers needing them can
	// inspect via OnAfterAction.
	OnAction func(name string, args []Value)

	// OnAfterAction fires after each yielding callable returns,
	// with the action's result Value and duration.
	OnAfterAction func(name string, result Value, elapsed time.Duration)

	// OnHandler fires before an `on` handler dispatches.
	OnHandler func(event string, args []Value)

	// OnAbort fires immediately before a routine ends with an
	// abort. Reason is the value passed to `abort`.
	OnAbort func(reason Value, pos ast.Node)
}

// fireRoutineStart is a nil-safe hook invocation helper.
func (h *Hooks) fireRoutineStart(routine string) {
	if h != nil && h.OnRoutineStart != nil {
		h.OnRoutineStart(routine)
	}
}

func (h *Hooks) fireRoutineEnd(routine string, r Result, elapsed time.Duration) {
	if h != nil && h.OnRoutineEnd != nil {
		h.OnRoutineEnd(routine, r, elapsed)
	}
}

func (h *Hooks) fireAction(name string, args []Value) {
	if h != nil && h.OnAction != nil {
		h.OnAction(name, args)
	}
}

func (h *Hooks) fireAfterAction(name string, result Value, elapsed time.Duration) {
	if h != nil && h.OnAfterAction != nil {
		h.OnAfterAction(name, result, elapsed)
	}
}

func (h *Hooks) fireHandler(event string, args []Value) {
	if h != nil && h.OnHandler != nil {
		h.OnHandler(event, args)
	}
}

func (h *Hooks) fireAbort(reason Value, pos ast.Node) {
	if h != nil && h.OnAbort != nil {
		h.OnAbort(reason, pos)
	}
}
