package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
//
// `Path` is the on-disk source if loaded from a file (via
// ParseRoutineFile), or the caller-supplied logical name if loaded
// from a string (via ParseRoutineString) — for instance
// `<repl-line-12>` or `exec:1a2b3c4d`. Either way, it's the
// identity used in traces, log lines, and conformance reports.
type RoutineFile struct {
	Path string
	File *ast.File
}

// ParseRoutineFile reads `path`, parses it, validates it, and
// **enforces that the filename basename (without `.routine`)
// matches the declared routine name** per docs/lang/syntax.md.
//
// For in-memory transient routines (REPL fragments, exec()-
// authored snippets, test cases that don't live on disk), use
// ParseRoutineString instead — that variant takes a logical name
// without filesystem coupling.
func ParseRoutineFile(path string) (*RoutineFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	file, err := parseAndValidate(path, string(raw))
	if err != nil {
		return nil, err
	}
	// Filename ↔ routine-name enforcement. Only meaningful for
	// the file loader; the string loader has no filename to check.
	wantName := strings.TrimSuffix(filepath.Base(path), ".routine")
	if file.Routine.Name != wantName {
		return nil, fmt.Errorf(
			"%s: filename basename %q must match declared routine name %q "+
				"(see docs/lang/syntax.md \"Filename ↔ routine name\")",
			path, wantName, file.Routine.Name)
	}
	return &RoutineFile{Path: path, File: file}, nil
}

// ParseRoutineString parses a routine from an in-memory source
// string with a caller-supplied logical name. The name appears in
// traces, observability hooks, and error messages; it is NOT
// matched against the declared routine name (transient routines
// can use whatever names make sense — `<repl-line-5>`,
// `exec:abc12345`, `test/foo`).
//
// Use this for the REPL, exec() / improvise() LLM-authored
// fragments, and any test code that wants to parse a routine
// from a literal string. The disk-backed `ParseRoutineFile`
// path is the only one that enforces filename matching.
func ParseRoutineString(logicalName, source string) (*RoutineFile, error) {
	if logicalName == "" {
		return nil, fmt.Errorf("ParseRoutineString: logicalName must be non-empty")
	}
	file, err := parseAndValidate(logicalName, source)
	if err != nil {
		return nil, err
	}
	return &RoutineFile{Path: logicalName, File: file}, nil
}

// parseAndValidate is the shared pipeline behind both loaders.
// Returns the validated *ast.File or a wrapped error with the
// loader-supplied identity (filename or logical name) for
// diagnostics. The "must have a routine declaration" check lives
// here too — both loaders require it.
func parseAndValidate(identity, source string) (*ast.File, error) {
	file, err := parser.Parse(identity, source)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", identity, err)
	}
	if err := validator.Validate(file); err != nil {
		return nil, fmt.Errorf("validate %s: %w", identity, err)
	}
	if file.Routine == nil {
		return nil, fmt.Errorf("%s: no routine declaration (only procs/handlers)", identity)
	}
	return file, nil
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
