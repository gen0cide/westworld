package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/validator"
)

// This file is the public entry point for running .routine files on
// a Host. Step 6 of the DSL plan.
//
// Workflow:
//
//   1. ParseRoutineFile parses + validates a .routine file from
//      disk.
//   2. NewRoutineInterpreter builds an interp.Interpreter wired to
//      this Host (reserved entities + action callables).
//   3. Run executes the routine against the live OpenRSC connection.
//
// Resource caps (op budget, wall clock, recursion, memory) are
// deferred to step 7 — for now, callers control termination via
// the passed context.

// RoutineFile is a parsed + validated .routine source file ready to
// hand to NewRoutineInterpreter.Run().
type RoutineFile struct {
	Path string
	File *ast.File
}

// ParseRoutineFile reads `path`, parses it, and validates it. Returns
// any parse or validation error with full source position.
func ParseRoutineFile(path string) (*RoutineFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	file, err := parser.Parse(path, string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := validator.Validate(file); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	if file.Routine == nil {
		return nil, fmt.Errorf("%s: no routine declaration (only procs/handlers)", path)
	}
	return &RoutineFile{Path: path, File: file}, nil
}

// NewRoutineInterpreter constructs an interp.Interpreter pre-loaded
// with this Host's reserved entities, action callables, and an
// event-translator goroutine that pipes typed bus events into the
// interpreter's PendingEvent channel.
//
// The returned interpreter is bound to `ctx` — every action
// invocation uses ctx as its deadline / cancellation signal, and
// the translator goroutine exits when ctx is canceled.
//
// Cancelling ctx interrupts any in-flight blocking action (walk_to,
// pick_up, wait) and the routine terminates with ResultCanceled.
func (h *Host) NewRoutineInterpreter(ctx context.Context) *interp.Interpreter {
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 64)
	h.startEventTranslator(ctx, it)
	// Reserved entities.
	it.Reserved["self"] = &selfView{host: h}
	it.Reserved["world"] = &worldView{host: h}
	it.Reserved["inventory"] = &inventoryView{host: h}
	it.Reserved["combat"] = &combatView{host: h}

	// Registration is driven entirely by dsl/spec/actions.go.
	// Every spec entry becomes a registered Callable; the spec's
	// Kind decides bang eligibility. The handler is looked up in
	// runtime/dsl_actions.go::actionHandlers (or replaced by a
	// NOT_IMPLEMENTED stub when spec.NotYetImplemented).
	//
	// To add a builtin: add a row to spec.Actions AND an entry in
	// actionHandlers (the consistency test catches mismatches).
	for i := range spec.Actions {
		a := &spec.Actions[i]
		fn, hasHandler := actionHandlers[a.Name]
		if !hasHandler || a.NotYetImplemented {
			fn = makeStub(a.Name)
		}
		base := &actionCallable{name: a.Name, host: h, ctx: ctx, fn: fn}
		it.Builtins[a.Name] = base
		if a.BangEligible() {
			it.Builtins[a.Name+"!"] = &interp.BangCallable{Underlying: base, Name: a.Name + "!"}
		}
	}

	return it
}

// RunRoutine parses, validates, and executes a routine file from
// disk. `args` are the positional arguments to the routine's entry
// point.
func (h *Host) RunRoutine(ctx context.Context, path string, args []interp.Value) (interp.Result, error) {
	rf, err := ParseRoutineFile(path)
	if err != nil {
		return interp.Result{}, err
	}
	it := h.NewRoutineInterpreter(ctx)
	return it.RunRoutine(ctx, rf.File, args), nil
}
