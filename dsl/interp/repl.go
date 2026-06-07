package interp

import (
	"context"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/parser"
)

// Session is a long-running interactive interpreter context — the
// model the REPL uses. Unlike RunRoutine (one-shot routine
// execution), a Session keeps its environment, registered procs,
// registered handlers, and budget across multiple Eval() calls.
//
// The cradle binary's `-repl` flag drives a Session: each line of
// REPL input becomes one Eval() call.
//
// Sessions inherit the Interpreter's Caps, Reserved entities, and
// Builtins. The budget tracker is shared across the whole session
// (a malicious one-liner can't reset the wall clock by quitting
// and re-entering).
type Session struct {
	interp *Interpreter
	env    *Env
	name   string // logical name for diagnostics ("<repl>")
}

// NewSession constructs a session against the given Interpreter.
// `name` is the logical identity used in parse-error messages
// (typically "<repl>" or "<test>"). The session starts with the
// interpreter's Reserved entities bound in its root env.
func (it *Interpreter) NewSession(ctx context.Context, name string) *Session {
	if name == "" {
		name = "<session>"
	}
	env := NewEnv()
	for k, v := range it.Reserved {
		env.Define(k, v)
	}
	// Fresh budget for the whole session. Sessions are
	// long-running so the wall-clock budget should generally be
	// raised by the caller via it.Caps if a >4h interactive run
	// is expected.
	it.budget = newBudget(it.Caps)
	it.OnHandlers = map[string][]*ast.OnHandler{}
	it.routineEnv = env
	return &Session{
		interp: it,
		env:    env,
		name:   name,
	}
}

// Env returns the session's root environment. Useful for tooling
// that wants to introspect or pre-bind values.
func (s *Session) Env() *Env { return s.env }

// PumpEvents drains the bus-event queue once and fires any registered
// on-handlers / select listeners / when-watchers against the session
// env. A long-running interactive session (REPL, -debug-http) has no
// routine loop to dispatch passively, so a host can call this on a
// ticker to make top-level `on` handlers fire while idle. Must not run
// concurrently with Eval on the same session (share a lock).
func (s *Session) PumpEvents(ctx context.Context) {
	s.interp.dispatchPendingEvents(ctx, s.env)
}

// Interpreter returns the underlying interpreter. Use for advanced
// operations (registering builtins, inspecting OnHandlers, etc.).
// Most callers should go through Eval / LoadFile.
func (s *Session) Interpreter() *Interpreter { return s.interp }

// LoadFile binds every proc + handler declared in the routine
// file into the session, without invoking the routine's entry
// point. Used by the REPL's `.load <path>` to make a file's
// helpers and reactive watchers available for interactive use.
//
// Procs become first-class callables in the session env; handlers
// extend the interpreter's OnHandlers map (so they fire on
// matching bus events while the session is alive). The routine
// declaration in file.Routine is NOT run here — use the
// session's Eval to invoke it, or use the REPL's `.run` which
// composes LoadFile + invoke.
func (s *Session) LoadFile(file *ast.File) {
	for _, p := range file.Procs {
		s.env.Define(p.Name, &procCallable{interp: s.interp, file: file, proc: p})
	}
	if s.interp.OnHandlers == nil {
		s.interp.OnHandlers = map[string][]*ast.OnHandler{}
	}
	for _, h := range file.Handlers {
		s.interp.OnHandlers[h.Event] = append(s.interp.OnHandlers[h.Event], h)
	}
}

// EvalResult is the outcome of evaluating one Session input. For
// expression inputs, Value is the evaluated expression and
// IsExpression is true. For statement inputs, Value is nil and
// IsExpression is false. Err is set if parsing or evaluation
// failed; the session remains usable.
type EvalResult struct {
	Value        Value
	IsExpression bool
	Err          error
}

// Eval parses and runs one line of DSL source against the
// session's persistent environment. Order of dispatch:
//
//  1. If the input starts with `on ` (a top-level event-handler
//     declaration), parse as an OnHandler and register it in the
//     session's OnHandlers map.
//  2. Otherwise try expression parse — most REPL inputs are
//     queries like `self.hp`.
//  3. Fall back to statement parse for assignments, control
//     flow, and bare action calls.
//
// Control-flow signals that escape the input (a stray `return`
// or `abort` typed at the REPL) get caught and surfaced as Err —
// they don't tear down the session.
func (s *Session) Eval(ctx context.Context, source string) EvalResult {
	res := EvalResult{}

	// 1. Top-level `on` handler declaration → register in session.
	trimmed := strings.TrimLeft(source, " \t\n\r")
	if strings.HasPrefix(trimmed, "on ") || strings.HasPrefix(trimmed, "on\t") {
		h, err := parser.ParseOnHandler(s.name, source)
		if err != nil {
			res.Err = err
			return res
		}
		if s.interp.OnHandlers == nil {
			s.interp.OnHandlers = map[string][]*ast.OnHandler{}
		}
		s.interp.OnHandlers[h.Event] = append(s.interp.OnHandlers[h.Event], h)
		return res
	}

	// Try expression parse first.
	if expr, err := parser.ParseExpr(s.name, source); err == nil && expr != nil {
		// Wrap in a defer/recover that converts any control-flow
		// signal or Go panic into an error so the session survives.
		var caught error
		var val Value
		func() {
			defer func() {
				switch r := recover().(type) {
				case nil:
					// ok
				case returnSignal:
					caught = fmt.Errorf("return outside of routine/proc")
				case abortSignal:
					if reason, ok := r.Reason.(Value); ok {
						caught = fmt.Errorf("abort: %s", reason.Display())
					} else {
						caught = fmt.Errorf("abort (unknown reason)")
					}
				case breakSignal:
					caught = fmt.Errorf("break outside of loop")
				case continueSignal:
					caught = fmt.Errorf("continue outside of loop")
				case *RuntimeError:
					caught = r
				default:
					caught = fmt.Errorf("session panic (contained): %v", r)
				}
			}()
			v, evalErr := s.interp.evalSafe(ctx, expr, s.env)
			if evalErr != nil {
				caught = evalErr
				return
			}
			val = v
		}()
		if caught != nil {
			res.Err = caught
			return res
		}
		res.Value = val
		res.IsExpression = true
		return res
	}

	// Fall back to statement parse.
	stmt, err := parser.ParseStmt(s.name, source)
	if err != nil {
		res.Err = err
		return res
	}
	if stmt == nil {
		// Empty input is fine — no-op.
		return res
	}
	var caught error
	func() {
		defer func() {
			switch r := recover().(type) {
			case nil:
				// ok
			case returnSignal:
				caught = fmt.Errorf("return outside of routine/proc")
			case abortSignal:
				if reason, ok := r.Reason.(Value); ok {
					caught = fmt.Errorf("abort: %s", reason.Display())
				} else {
					caught = fmt.Errorf("abort (unknown reason)")
				}
			case breakSignal:
				caught = fmt.Errorf("break outside of loop")
			case continueSignal:
				caught = fmt.Errorf("continue outside of loop")
			case *RuntimeError:
				caught = r
			default:
				caught = fmt.Errorf("session panic (contained): %v", r)
			}
		}()
		s.interp.execStmt(ctx, stmt, s.env)
	}()
	if caught != nil {
		res.Err = caught
	}
	return res
}
