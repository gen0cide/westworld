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
// view consults, plus the small machinery (boundAction, bang
// wrapping) that turns an actionHandler into a Callable bound to the
// host's routine ctx. The handler FUNCTION BODIES still live in the
// per-namespace actions_*.go files and are shared with the flat
// aliases that survive in spec.Actions (attack, cast).
//
// Adding a namespaced verb = add a row to the relevant table below
// AND a documentation row in dsl/spec/accessors.go (the validator's
// declared surface). The accessor-consistency test asserts each
// path resolves against a stub host.

// boundAction wraps an actionHandler into an actionCallable bound to
// the host and the host's active routine ctx. Reuses the same
// actionCallable shape the flat builtins use, so yield/Result/error
// semantics are identical.
func (h *Host) boundAction(name string, fn actionHandler) *actionCallable {
	ctx := h.routineCtx
	if ctx == nil {
		ctx = context.Background()
	}
	return &actionCallable{name: name, host: h, ctx: ctx, fn: fn}
}

// namespaceAction resolves a namespace verb (with optional trailing
// bang) against the supplied verb table and returns a ready Callable.
// `root` is the namespace name used for display/diagnostics
// (e.g. "trade"); `field` is the requested member (e.g. "request" or
// "request!"). Returns (nil, false) if the verb is unknown so the
// view's Get can fall through to its Views.
func (h *Host) namespaceAction(root, field string, table map[string]actionHandler) (interp.Value, bool) {
	base, isBang := spec.StripBang(field)
	fn, ok := table[base]
	if !ok {
		return nil, false
	}
	call := h.boundAction(root+"."+base, fn)
	if isBang {
		return &interp.BangCallable{Underlying: call, Name: root + "." + base + "!"}, true
	}
	return call, true
}
