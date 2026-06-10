package runtime

import (
	"context"
	"sync"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Regression tests for the leak-audit #3 routineCtx contamination bug:
// Host.routineCtx was a HOST-GLOBAL written by every
// NewRoutineInterpreter construction and read by namespaceAction/
// boundAction at namespace-verb access time. A detour interpreter's
// construction therefore re-pointed the global at the detour ctx, and
// after runDetour's deferred cancel fired, EVERY namespace verb the
// resumed (parked) grind touched — combat.retreat() included — bound
// a CANCELLED context until the next turn rebuilt the interpreter.
// The fix carries the ctx on a per-interpreter routineBinding shared
// by that interpreter's views, making namespace verbs identical in
// ctx discipline to flat builtins (which capture ctx at construction).

// reservedView pulls a namespace root out of an interpreter's
// Reserved table as a Get-able view.
func reservedView(t *testing.T, it *interp.Interpreter, root string) interface {
	Get(field string) (interp.Value, bool)
} {
	t.Helper()
	v, ok := it.Reserved[root]
	if !ok {
		t.Fatalf("Reserved[%q] missing", root)
	}
	g, ok := v.(interface {
		Get(field string) (interp.Value, bool)
	})
	if !ok {
		t.Fatalf("Reserved[%q] (%T) is not a Get-able view", root, v)
	}
	return g
}

// verbCtx resolves a namespace verb through a view and returns the
// ctx its actionCallable bound. Unwraps bang callables.
func verbCtx(t *testing.T, view interface {
	Get(field string) (interp.Value, bool)
}, field string) context.Context {
	t.Helper()
	v, ok := view.Get(field)
	if !ok {
		t.Fatalf("verb %q did not resolve", field)
	}
	var under any = v
	if bc, isBang := v.(*interp.BangCallable); isBang {
		under = bc.Underlying
	}
	ac, ok := under.(*actionCallable)
	if !ok {
		t.Fatalf("verb %q resolved to %T, want *actionCallable", field, under)
	}
	if ac.ctx == nil {
		// Pre-fix shopView built its callables with NO ctx at all.
		t.Fatalf("verb %q bound a nil ctx", field)
	}
	return ac.ctx
}

// TestNamespaceVerbsSurviveDetourCtxCancel simulates the detour
// park/resume sequence: interpreter A (the grind) is built from a
// live ctx, interpreter B (the detour) is built afterwards from its
// own ctx which is then CANCELLED (runDetour's deferred dcancel).
// A's namespace verbs must still bind A's live ctx — with the
// pre-fix host-global they bound B's cancelled ctx.
func TestNamespaceVerbsSurviveDetourCtxCancel(t *testing.T) {
	h := newTestHost()

	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	itA := h.NewRoutineInterpreter(ctxA)

	// Detour: built second, cancelled before the grind resumes.
	ctxB, cancelB := context.WithCancel(context.Background())
	itB := h.NewRoutineInterpreter(ctxB)
	cancelB()

	// Every namespace root the grind might touch on resume, including
	// the world.* sub-view path (world.trade/duel/bank) and shop.
	roots := []struct {
		root string
		verb string
	}{
		{"combat", "retreat"}, // the detour-resume safety net from the audit
		{"combat", "attack"},
		{"combat", "attack!"}, // bang variant unwraps to the same binding
		{"trade", "request"},
		{"bank", "deposit"},
		{"duel", "stake"},
		{"magic", "cast"},
		{"prayer", "activate"},
		{"shop", "buy"},
	}
	for _, c := range roots {
		got := verbCtx(t, reservedView(t, itA, c.root), c.verb)
		if got.Err() != nil {
			t.Errorf("grind %s.%s bound a dead ctx (%v) after detour cancel — routineCtx contamination", c.root, c.verb, got.Err())
		}
		if got != ctxA {
			t.Errorf("grind %s.%s bound ctx %p, want the grind's own ctx %p", c.root, c.verb, got, ctxA)
		}
	}

	// world.trade / world.duel / world.bank sub-views must propagate
	// the same per-interpreter binding.
	worldA := reservedView(t, itA, "world")
	for _, sub := range []struct {
		root string
		verb string
	}{{"trade", "request"}, {"duel", "stake"}, {"bank", "deposit"}} {
		sv, ok := worldA.Get(sub.root)
		if !ok {
			t.Fatalf("world.%s did not resolve", sub.root)
		}
		view, ok := sv.(interface {
			Get(field string) (interp.Value, bool)
		})
		if !ok {
			t.Fatalf("world.%s (%T) is not a Get-able view", sub.root, sv)
		}
		got := verbCtx(t, view, sub.verb)
		if got != ctxA {
			t.Errorf("world.%s.%s bound ctx %p, want the grind's ctx %p", sub.root, sub.verb, got, ctxA)
		}
	}

	// And the detour's OWN verbs correctly bind the detour's
	// (now-cancelled) ctx — per-interpreter discipline cuts both ways.
	if got := verbCtx(t, reservedView(t, itB, "combat"), "retreat"); got != ctxB {
		t.Errorf("detour combat.retreat bound ctx %p, want the detour's ctx %p", got, ctxB)
	}
}

// TestNamespaceCtxConcurrentConstructionNoRace exercises the (b) data
// race from the audit: debughttp POST /script and Activate construct
// interpreters on HTTP goroutines while a live conductor routine
// resolves namespace verbs. With the host-global routineCtx this was
// a go-race-detectable write/read race; with the per-interpreter
// binding there is no shared mutable state. Run under -race.
func TestNamespaceCtxConcurrentConstructionNoRace(t *testing.T) {
	h := newTestHost()

	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	itA := h.NewRoutineInterpreter(ctxA)
	combat := reservedView(t, itA, "combat")

	const iters = 200
	var wg sync.WaitGroup
	wg.Add(2)

	// HTTP-goroutine analogue: keep constructing fresh interpreters.
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			h.NewRoutineInterpreter(ctx)
			cancel() // detour-style teardown; also reaps the event translator
		}
	}()

	// Conductor-routine analogue: keep resolving namespace verbs on
	// the long-lived interpreter; the binding must stay ctxA.
	// (No t.Fatalf here — only Errorf is goroutine-safe.)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			v, ok := combat.Get("retreat")
			if !ok {
				t.Errorf("iteration %d: combat.retreat did not resolve", i)
				return
			}
			ac, ok := v.(*actionCallable)
			if !ok {
				t.Errorf("iteration %d: combat.retreat resolved to %T, want *actionCallable", i, v)
				return
			}
			if ac.ctx != ctxA {
				t.Errorf("iteration %d: combat.retreat bound foreign ctx %p, want %p", i, ac.ctx, ctxA)
				return
			}
		}
	}()

	wg.Wait()
}
