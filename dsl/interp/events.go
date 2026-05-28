package interp

import (
	"context"

	"github.com/gen0cide/westworld/dsl/ast"
)

// PendingEvent is one event from the host's event bus that the
// interpreter should dispatch to a matching `on` handler. Host code
// constructs PendingEvents and pushes them onto Interpreter.Events.
//
// `Name` matches the handler signature (e.g. "chat_received").
// `Args` are the positional parameters passed to the handler in
// declaration order — matching the v1 event table in dsl.md.
type PendingEvent struct {
	Name string
	Args []Value
}

// Yielder is implemented by Callables that yield control to the
// event scheduler before they run. Primary actions (walk_to,
// attack, etc.) yield so handlers can interleave between actions —
// per dsl.md "Run **between** primary actions, never mid-action."
//
// Pure functions (procs, stdlib oracles like evaluate) do NOT yield;
// they're treated as pure computation that can't be interrupted.
type Yielder interface {
	Yields() bool
}

// dispatchPendingEvents drains the Events channel without blocking
// and runs any registered handler whose name matches. Called from
// the start of every yielding action invocation, plus once before
// the routine's main body begins.
//
// Handler bodies run synchronously in the calling goroutine, in a
// fresh child of the routine's env so they can read locals but new
// assignments don't leak out of the handler. (dsl.md: handlers
// have read-only access to routine locals.)
func (it *Interpreter) dispatchPendingEvents(ctx context.Context, routineEnv *Env) {
	if it.Events == nil || len(it.OnHandlers) == 0 {
		return
	}
	for {
		var ev PendingEvent
		select {
		case ev = <-it.Events:
		default:
			return
		}
		handlers := it.OnHandlers[ev.Name]
		for _, h := range handlers {
			it.Hooks.fireHandler(ev.Name, ev.Args)
			it.runHandler(ctx, h, routineEnv, ev.Args)
		}
	}
}

// runHandler runs one handler body in a child of the routine's env.
// Handler control-flow signals (return / break / continue) are
// trapped here — handlers cannot return from the routine.
// abortSignal from a handler does propagate, since dsl.md allows
// `on death { abort "died" }` to terminate the parent routine.
func (it *Interpreter) runHandler(ctx context.Context, h *ast.OnHandler, routineEnv *Env, args []Value) {
	handlerEnv := routineEnv.Child()
	// Bind handler parameters.
	for i, p := range h.Params {
		if i < len(args) {
			handlerEnv.Define(p.Name, args[i])
		} else {
			handlerEnv.Define(p.Name, Null{})
		}
	}
	defer func() {
		switch r := recover().(type) {
		case nil:
		case breakSignal, continueSignal, returnSignal:
			// Handlers can't break/continue/return out of the routine.
			// Silently swallow — dsl.md says handlers run to
			// completion and yield back to the routine.
			_ = r
		case abortSignal, *RuntimeError:
			// These should propagate to the routine — re-panic.
			panic(r)
		default:
			panic(r)
		}
	}()
	it.execBlock(ctx, h.Body, handlerEnv)
}

// registerHandlers walks the file's top-level on-handlers (and the
// routine's handlers, if any) and indexes them by event name. Called
// once at routine startup.
func (it *Interpreter) registerHandlers(file *ast.File, r *ast.RoutineDecl) {
	it.OnHandlers = map[string][]*ast.OnHandler{}
	for _, h := range file.Handlers {
		it.OnHandlers[h.Event] = append(it.OnHandlers[h.Event], h)
	}
	if r != nil {
		for _, h := range r.Handlers {
			it.OnHandlers[h.Event] = append(it.OnHandlers[h.Event], h)
		}
	}
}
