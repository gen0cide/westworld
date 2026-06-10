package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/spec"
)

// namespace_actions.go is the NAMESPACE ACTION DISPATCH HUB.
//
// Per api.md §6 the frozen surface namespaces every Action under its
// subsystem root — trade.request, bank.deposit, duel.stake,
// magic.cast, prayer.activate, combat.attack. These are NOT bare
// builtins (which would defeat the namespacing rule and trip the
// validator's snake_case action-name constraint); they are *view-
// dispatched* callables returned from the namespace views' Get().
//
// This file holds the single per-namespace verb→handler table each
// view consults, plus the small machinery (routineBinding,
// boundAction, bang wrapping) that turns an actionHandler into a
// Callable bound to the OWNING INTERPRETER's routine ctx — the same
// ctx the flat builtins capture. The handler FUNCTION BODIES still live in the
// per-namespace actions_*.go files and are shared with the flat
// aliases that survive in spec.Actions (attack, cast).
//
// Adding a namespaced verb = add a row to the relevant table below
// AND a documentation row in dsl/spec/accessors.go (the validator's
// declared surface). The accessor-consistency test asserts each
// path resolves against a stub host.

// routineBinding carries the routine context for ONE interpreter
// construction. NewRoutineInterpreter builds a single binding per
// interpreter and shares it across all of that interpreter's
// namespace views, so every namespace-dispatched verb binds the SAME
// ctx the flat builtins captured at construction — no host-global,
// no cross-interpreter contamination (a cancelled detour ctx can
// never poison a parked grind's combat.retreat()), and no data race
// when an HTTP-goroutine interpreter (debughttp /script, Activate)
// is built while a conductor routine runs.
//
// nil-safe: views constructed without a binding (worldView sub-view
// literals in tests, selfView's internal combatView) fall back to
// context.Background(), matching the old nil-routineCtx behavior.
type routineBinding struct {
	ctx context.Context
}

// context returns the bound routine ctx, or context.Background()
// when the binding (or its ctx) was never set.
func (b *routineBinding) context() context.Context {
	if b == nil || b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

// boundAction wraps an actionHandler into an actionCallable bound to
// the host and the interpreter's routine ctx (via the per-interpreter
// binding). Reuses the same actionCallable shape the flat builtins
// use, so yield/Result/error semantics are identical.
func (h *Host) boundAction(bind *routineBinding, name string, fn actionHandler) *actionCallable {
	return &actionCallable{name: name, host: h, ctx: bind.context(), fn: fn}
}

// namespaceAction resolves a namespace verb (with optional trailing
// bang) against the supplied verb table and returns a ready Callable
// bound to the calling view's interpreter ctx. `root` is the
// namespace name used for display/diagnostics (e.g. "trade");
// `field` is the requested member (e.g. "request" or "request!").
// Returns (nil, false) if the verb is unknown so the view's Get can
// fall through to its Views.
func (h *Host) namespaceAction(bind *routineBinding, root, field string, table map[string]actionHandler) (interp.Value, bool) {
	base, isBang := spec.StripBang(field)
	fn, ok := table[base]
	if !ok {
		return nil, false
	}
	call := h.boundAction(bind, root+"."+base, fn)
	if isBang {
		return &interp.BangCallable{Underlying: call, Name: root + "." + base + "!"}, true
	}
	return call, true
}
