package interp

import (
	"context"
	"fmt"

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
// session's persistent environment. Try-parse-as-expression first
// (most REPL inputs are queries like `self.hp`); fall back to
// statement parsing for assignments, control flow, and bare
// action calls.
//
// Control-flow signals that escape the input (a stray `return`
// or `abort` typed at the REPL) get caught and surfaced as Err —
// they don't tear down the session.
func (s *Session) Eval(ctx context.Context, source string) EvalResult {
	res := EvalResult{}

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
