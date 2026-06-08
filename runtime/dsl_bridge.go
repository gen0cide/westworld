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
	// Parse first (no validate), so we can resolve `extends "..."`
	// paths and merge parent procs + on-handlers into this file
	// *before* the validator runs against the unified declaration set.
	file, err := parser.Parse(path, string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", path, err)
	}
	visited := map[string]bool{absPath: true}
	if err := mergeExtends(file, filepath.Dir(absPath), visited); err != nil {
		return nil, err
	}
	if err := validator.Validate(file); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	// Runtime version targeting is MANDATORY for disk-loaded routines: every
	// .routine must declare `runtime "X.Y"`, and that target must be compatible
	// with the current runtime (see docs/lang/versioning.md).
	if file.Runtime == "" {
		return nil, fmt.Errorf(
			"%s: missing required `runtime \"X.Y\"` directive — every .routine must declare the Routine Runtime version it targets (current is %s; see docs/lang/versioning.md)",
			path, spec.RuntimeVersion)
	}
	if err := spec.CheckTarget(file.Runtime); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if file.Routine == nil {
		return nil, fmt.Errorf("%s: no routine declaration (only procs/handlers)", path)
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

// mergeExtends resolves each `extends "..."` path relative to
// `baseDir`, recursively loads transitive parents, and merges their
// procs / on-handlers / bounds into `file`.
//
// Semantics (v1):
//   - Handlers: additive — parent's `on chat_received` and child's
//     `on chat_received` both fire (in declaration order: parent
//     first, then child).
//   - Procs: child overrides parent — if both define `proc helper()`,
//     the child's wins. (super() to chain is deferred.)
//   - Bounds: additive, parent-first.
//   - Parent files must NOT declare a `routine ...` — they are
//     libraries (procs + handlers only).
//   - Cycles are rejected (`extends` chain forming a loop).
func mergeExtends(file *ast.File, baseDir string, visited map[string]bool) error {
	for _, raw := range file.Extends {
		resolved := raw
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(baseDir, raw)
		}
		abs, err := filepath.Abs(resolved)
		if err != nil {
			return fmt.Errorf("extends %q: %w", raw, err)
		}
		if visited[abs] {
			return fmt.Errorf("extends cycle: %s already in chain", abs)
		}
		visited[abs] = true

		rawBytes, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("extends %q (resolved to %s): %w", raw, abs, err)
		}
		parent, err := parser.Parse(abs, string(rawBytes))
		if err != nil {
			return fmt.Errorf("parse parent %s: %w", abs, err)
		}
		if parent.Routine != nil {
			return fmt.Errorf("%s: parent file %s must not declare a routine — parents are libraries (procs + handlers only)", file.Filename, abs)
		}
		// Recurse so the deepest ancestor's decls land in `file`
		// first, then closer ancestors override them.
		if err := mergeExtends(parent, filepath.Dir(abs), visited); err != nil {
			return err
		}
		// Handlers + bounds: additive, parent before child.
		file.Handlers = append(append([]*ast.OnHandler{}, parent.Handlers...), file.Handlers...)
		file.Bounds = append(append([]*ast.BoundsDecl{}, parent.Bounds...), file.Bounds...)
		// Procs: child overrides on name collision.
		childNames := make(map[string]bool, len(file.Procs))
		for _, p := range file.Procs {
			childNames[p.Name] = true
		}
		var merged []*ast.ProcDecl
		for _, p := range parent.Procs {
			if !childNames[p.Name] {
				merged = append(merged, p)
			}
		}
		merged = append(merged, file.Procs...)
		file.Procs = merged
	}
	return nil
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
	file, err := parser.Parse(logicalName, source)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", logicalName, err)
	}
	if len(file.Extends) > 0 {
		return nil, fmt.Errorf("%s: extends is only supported in disk-loaded routines (ParseRoutineFile); string-loaded routines have no base directory for path resolution", logicalName)
	}
	if err := validator.Validate(file); err != nil {
		return nil, fmt.Errorf("validate %s: %w", logicalName, err)
	}
	// Runtime targeting is OPTIONAL for transient string-loaded routines (REPL
	// fragments, exec()/improvise() snippets); if declared it must be compatible.
	if file.Runtime != "" {
		if err := spec.CheckTarget(file.Runtime); err != nil {
			return nil, fmt.Errorf("%s: %w", logicalName, err)
		}
	}
	if file.Routine == nil {
		return nil, fmt.Errorf("%s: no routine declaration (only procs/handlers)", logicalName)
	}
	return &RoutineFile{Path: logicalName, File: file}, nil
}

// interpOptions configure NewRoutineInterpreter. Zero value = the normal gated
// path (every existing caller).
type interpOptions struct {
	// ungated skips the pearl Gate wrap on PrimaryAction handlers. Used ONLY by
	// the ANALYSIS-mode operator command path, where the operator's directive
	// bypasses pearl/persona/the Act planner by design. The autonomous conductor
	// and the dry-run path keep the gated default so real cognition still runs
	// real policy.
	ungated bool
}

// InterpOption mutates the interpreter construction options.
type InterpOption func(*interpOptions)

// WithoutPearlGate builds an interpreter whose PrimaryAction handlers are NOT
// wrapped by the pearl Gate — the ANALYSIS-mode operator-command bypass.
func WithoutPearlGate() InterpOption { return func(o *interpOptions) { o.ungated = true } }

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
func (h *Host) NewRoutineInterpreter(ctx context.Context, opts ...InterpOption) *interp.Interpreter {
	var io interpOptions
	for _, o := range opts {
		o(&io)
	}
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 64)
	// Install the per-statement line hook (cradle live Routine panel) when the
	// host has one wired. Covers both the plain (RunRoutine/RunRoutineSource)
	// and the detour (StartCoro) paths — every routine run builds its
	// interpreter here.
	if h.OnStmt != nil {
		it.Hooks = &interp.Hooks{OnStmt: h.OnStmt}
	}
	h.startEventTranslator(ctx, it)
	// Bind the routine ctx so namespace-dispatched action callables
	// (trade.*, bank.*, duel.*, magic.cast, prayer.*) inherit the same
	// cancellation/deadline as the flat builtins (which capture ctx in
	// their actionCallable).
	h.routineCtx = ctx
	// Reserved entities — the namespace roots (§6). self/world/
	// inventory/combat plus the promoted top-level subsystem roots
	// trade/bank/duel/magic/prayer.
	it.Reserved["self"] = &selfView{host: h}
	it.Reserved["world"] = &worldView{host: h}
	it.Reserved["inventory"] = &inventoryView{host: h}
	it.Reserved["combat"] = &combatView{host: h}
	it.Reserved["trade"] = &tradeView{host: h}
	it.Reserved["bank"] = &bankView{host: h}
	it.Reserved["duel"] = &duelView{host: h}
	it.Reserved["magic"] = &magicView{host: h}
	it.Reserved["prayer"] = &prayerView{host: h}
	it.Reserved["shop"] = &shopView{host: h}

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
		// Apperception: wrap state-mutating actions with the pearl gate so the
		// host's own policy can veto/substitute before a packet is sent. Only
		// when an engine is wired; read-only actions are never gated. The
		// ANALYSIS-mode operator path opts OUT (io.ungated) so a direct command
		// bypasses pearl by design.
		if h.Pearl != nil && a.Kind == spec.PrimaryAction && !io.ungated {
			fn = h.gateAction(a.Name, fn)
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

// RunRoutineSource parses, validates, and executes a routine from a DSL SOURCE
// string (no file). name is the logical name used in parse errors / logs. This
// is the entry point for mesa-authored routines (Act's WriteRoutine moves) —
// they run through the same interpreter + pearl gate as file routines.
func (h *Host) RunRoutineSource(ctx context.Context, name, source string, args []interp.Value, opts ...InterpOption) (interp.Result, error) {
	if name == "" {
		name = "mesa/authored"
	}
	rf, err := ParseRoutineString(name, source)
	if err != nil {
		return interp.Result{}, err
	}
	it := h.NewRoutineInterpreter(ctx, opts...)
	return it.RunRoutine(ctx, rf.File, args), nil
}
