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
	// Drain bus events: deliver to AST on-handlers AND Go-side
	// event listeners (used by select-on cases). Then sweep any
	// active when-watchers so state-transition predicates fire on
	// the same beat as the events that triggered the state change.
	if it.Events != nil {
		for {
			var ev PendingEvent
			select {
			case ev = <-it.Events:
			default:
				goto afterDrain
			}
			// Listener fires first (select-on captures the event
			// even if an AST handler would consume it). Both surfaces
			// see every event — they're not mutually exclusive.
			it.fireEventListeners(ev.Name, ev.Args)
			handlers := it.OnHandlers[ev.Name]
			for _, h := range handlers {
				it.Hooks.fireHandler(ev.Name, ev.Args)
				it.runHandler(ctx, h, routineEnv, ev.Args)
			}
			// Bounded handlers — filtered by location, if the event
			// carries one. Convention: handler arity-3 args[1] and
			// args[2] are (x, y); arity-2 args[0] and args[1] are
			// (x, y). Other arities = no location, skip the filter.
			bounded := it.BoundedHandlers[ev.Name]
			if len(bounded) > 0 {
				x, y, hasLoc := eventLocation(ev.Args)
				for _, bh := range bounded {
					if hasLoc && !bh.Matches(x, y) {
						continue
					}
					it.Hooks.fireHandler(ev.Name, ev.Args)
					it.runHandler(ctx, bh.Handler, routineEnv, ev.Args)
				}
			}
		}
	}
afterDrain:
	// Watcher sweep happens AFTER event delivery so a handler that
	// mutates state (or that we observe via recent-events buffer)
	// causes the watcher to fire on the same boundary.
	it.evalWatchersOnce(ctx)
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
// once at routine startup. Also walks the file's bounds declarations
// and registers their on-handlers with a location filter.
func (it *Interpreter) registerHandlers(ctx context.Context, file *ast.File, r *ast.RoutineDecl) {
	it.OnHandlers = map[string][]*ast.OnHandler{}
	it.BoundedHandlers = map[string][]BoundedHandler{}
	for _, h := range file.Handlers {
		it.OnHandlers[h.Event] = append(it.OnHandlers[h.Event], h)
	}
	if r != nil {
		for _, h := range r.Handlers {
			it.OnHandlers[h.Event] = append(it.OnHandlers[h.Event], h)
		}
	}
	for _, b := range file.Bounds {
		it.registerBoundsDecl(ctx, b, nil)
	}
}

// RegionPredicate is a (x, y) → bool filter returned by bounds
// shape constructors. The dispatcher applies it to the (x, y) args
// of location-bearing events (item_appeared, item_disappeared) to
// decide whether a bounded handler fires.
type RegionPredicate func(x, y int) bool

// BoundedHandler bundles an OnHandler with the predicate stack
// from its enclosing bounds blocks. All predicates must return
// true (logical AND — intersection of nested bounds) for the
// handler to fire.
type BoundedHandler struct {
	Handler *ast.OnHandler
	Filters []RegionPredicate
}

// Matches returns true iff every predicate in the stack accepts
// the (x, y) location. Used by the event dispatcher.
func (b BoundedHandler) Matches(x, y int) bool {
	for _, f := range b.Filters {
		if !f(x, y) {
			return false
		}
	}
	return true
}

// registerBoundsDecl evaluates the bounds shape expression to a
// RegionPredicate, appends it to the inherited predicate stack,
// and walks the contained on-handlers (and nested bounds).
func (it *Interpreter) registerBoundsDecl(ctx context.Context, b *ast.BoundsDecl, inherited []RegionPredicate) {
	// Evaluate the shape expression in the routine env (or a fresh
	// env if registration happens before the routine starts — for
	// now we use the routineEnv set by runDecl).
	env := it.routineEnv
	if env == nil {
		// Bounds registration runs after registerHandlers and BEFORE
		// the routine body. If routineEnv isn't set yet, use the
		// global root env.
		env = NewEnv()
	}
	shapeVal, err := func() (v Value, rerr *RuntimeError) {
		defer func() {
			if r := recover(); r != nil {
				if re, ok := r.(*RuntimeError); ok {
					rerr = re
				}
			}
		}()
		v = it.eval(ctx, b.Shape, env)
		return v, nil
	}()
	if err != nil || shapeVal == nil {
		// Shape evaluation failed — register handlers with a
		// never-match predicate so the validator can still catch
		// the runtime issue.
		neverMatch := RegionPredicate(func(int, int) bool { return false })
		filters := append(append([]RegionPredicate(nil), inherited...), neverMatch)
		it.registerBoundedHandlers(b, filters)
		return
	}
	pred, ok := shapeVal.(*regionPredicateValue)
	if !ok {
		// The shape didn't evaluate to a region predicate. Wrap as
		// never-match.
		neverMatch := RegionPredicate(func(int, int) bool { return false })
		filters := append(append([]RegionPredicate(nil), inherited...), neverMatch)
		it.registerBoundedHandlers(b, filters)
		return
	}
	filters := append(append([]RegionPredicate(nil), inherited...), pred.predicate)
	it.registerBoundedHandlers(b, filters)
	for _, nested := range b.Bounds {
		it.registerBoundsDecl(ctx, nested, filters)
	}
}

// registerBoundedHandlers registers each handler in the bounds
// block under BoundedHandlers with the given filter stack.
func (it *Interpreter) registerBoundedHandlers(b *ast.BoundsDecl, filters []RegionPredicate) {
	for _, h := range b.Handlers {
		it.BoundedHandlers[h.Event] = append(it.BoundedHandlers[h.Event], BoundedHandler{
			Handler: h,
			Filters: filters,
		})
	}
}

// regionPredicateValue is the Value type returned by the built-in
// bounds shape constructors (box, circle, near). The dispatcher
// type-asserts on it to extract the filter function.
type regionPredicateValue struct {
	name      string // for debugging / Display()
	predicate RegionPredicate
}

func (r *regionPredicateValue) Kind() string    { return "region" }
func (r *regionPredicateValue) Display() string { return "<region " + r.name + ">" }

// NewRegionPredicate constructs a Value that bounds blocks recognize
// as a shape. Host-side shape constructors (box, circle, near) return
// these from their Call() methods.
func NewRegionPredicate(name string, pred RegionPredicate) Value {
	return &regionPredicateValue{name: name, predicate: pred}
}

// eventLocation extracts (x, y) from an event's args based on
// per-event-name conventions in dsl/spec/events.go. Returns
// hasLoc=false for events without location params (chat, death,
// trade_*, duel_other_accepted, etc.) — these pass through any
// bounds filter unfiltered.
//
// Convention:
//   item_appeared(item_id, x, y)     → args[1], args[2]
//   item_disappeared(item_id, x, y)  → args[1], args[2]
//
// New location-bearing events should be added here as they're
// wired. Events with no location simply skip the filter (handler
// fires regardless of bounds).
func eventLocation(args []Value) (x, y int, hasLoc bool) {
	if len(args) < 3 {
		return 0, 0, false
	}
	xi, xok := AsInt(args[1])
	yi, yok := AsInt(args[2])
	if !xok || !yok {
		return 0, 0, false
	}
	return int(xi), int(yi), true
}
