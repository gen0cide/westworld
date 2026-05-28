package interp_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
)

// mustParseFile parses without validating (these tests inject
// arbitrary test-only builtins like `tick`/`hp` that aren't in
// dsl/spec — same pattern as lambda_test.go).
func mustParseFile(t *testing.T, src string) *ast.File {
	t.Helper()
	f, err := parser.Parse("w.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return f
}

// pureCallable is a non-yielding builtin — used for state queries
// inside watcher predicates (subscription-safe by construction).
type pureCallable struct {
	fn func(args []interp.Value) interp.Value
}

func (p *pureCallable) Kind() string    { return "pure" }
func (p *pureCallable) Display() string { return "<pure>" }
func (p *pureCallable) Yields() bool    { return false }
func (p *pureCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	return p.fn(args), nil
}

// TestWhenFiresOnRisingEdge — register `when hp() < 40 { eat() }`;
// drop HP across the threshold; verify eat called exactly once.
func TestWhenFiresOnRisingEdge(t *testing.T) {
	src := `
routine r() {
    when hp() < 40 {
        eat()
    }
    tick()
    tick()
    tick()
    return "done"
}
`
	f := mustParseFile(t, src)

	var hp int32 = 50
	var eatCount int32

	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Builtins["hp"] = &pureCallable{fn: func(_ []interp.Value) interp.Value {
		return interp.Int(int64(atomic.LoadInt32(&hp)))
	}}
	it.Builtins["eat"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&eatCount, 1)
		return interp.Null{}, nil
	}}
	// `tick` is a yielding no-op that mutates state between ticks
	// to drive the watcher to fire.
	var tickN int32
	it.Builtins["tick"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		n := atomic.AddInt32(&tickN, 1)
		switch n {
		case 1:
			atomic.StoreInt32(&hp, 35) // cross threshold → should fire
		case 2:
			atomic.StoreInt32(&hp, 30) // already below → should NOT re-fire
		case 3:
			atomic.StoreInt32(&hp, 50)
			// next tick after this would re-arm; we stop here
		}
		return interp.Null{}, nil
	}}

	res := runRoutine(t, it, f)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if got := atomic.LoadInt32(&eatCount); got != 1 {
		t.Errorf("eat called %d times, want 1 (rising edge once)", got)
	}
}

// TestWhenFiresImmediatelyIfAlreadyTrue — design says
// "registration with already-true counts as the rising edge".
func TestWhenFiresImmediatelyIfAlreadyTrue(t *testing.T) {
	src := `
routine r() {
    when hp() < 40 {
        eat()
    }
    tick()
    return "done"
}
`
	f := mustParseFile(t, src)

	var hp int32 = 35 // already below threshold at registration
	var eatCount int32
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Builtins["hp"] = &pureCallable{fn: func(_ []interp.Value) interp.Value {
		return interp.Int(int64(atomic.LoadInt32(&hp)))
	}}
	it.Builtins["eat"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&eatCount, 1)
		return interp.Null{}, nil
	}}
	it.Builtins["tick"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		return interp.Null{}, nil
	}}

	res := runRoutine(t, it, f)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if got := atomic.LoadInt32(&eatCount); got != 1 {
		t.Errorf("eat: got %d, want 1 (already-true should fire immediately on register)", got)
	}
}

// TestWhenScopeReleasesOnBlockExit — watchers inside a nested
// block should not fire after that block exits.
func TestWhenScopeReleasesOnBlockExit(t *testing.T) {
	src := `
routine r() {
    {
        when hp() < 40 { eat() }
        tick()
    }
    tick()
    tick()
    return "done"
}
`
	f := mustParseFile(t, src)

	var hp int32 = 50
	var eatCount int32
	var tickN int32
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Builtins["hp"] = &pureCallable{fn: func(_ []interp.Value) interp.Value {
		return interp.Int(int64(atomic.LoadInt32(&hp)))
	}}
	it.Builtins["eat"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&eatCount, 1)
		return interp.Null{}, nil
	}}
	it.Builtins["tick"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		n := atomic.AddInt32(&tickN, 1)
		switch n {
		case 1:
			// inside block — keep above threshold so watcher doesn't fire
		case 2:
			// outside block — drop below, but watcher is unregistered
			atomic.StoreInt32(&hp, 30)
		case 3:
			atomic.StoreInt32(&hp, 20)
		}
		return interp.Null{}, nil
	}}

	res := runRoutine(t, it, f)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if got := atomic.LoadInt32(&eatCount); got != 0 {
		t.Errorf("eat: got %d, want 0 (watcher should be unregistered after block exit)", got)
	}
}

// TestSelectPicksFirstReadyWhenCase — first-declared wins on
// pre-check, even if a later case is also true at entry.
func TestSelectPicksFirstReadyWhenCase(t *testing.T) {
	src := `
routine r() {
    select {
        when first()  { note("first")  }
        when second() { note("second") }
    }
    return "done"
}
`
	f := mustParseFile(t, src)

	var captured string
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Builtins["first"] = &pureCallable{fn: func(_ []interp.Value) interp.Value { return interp.Bool(true) }}
	it.Builtins["second"] = &pureCallable{fn: func(_ []interp.Value) interp.Value { return interp.Bool(true) }}
	it.Builtins["note"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		if len(args) > 0 {
			captured = string(args[0].(interp.String))
		}
		return interp.Null{}, nil
	}}

	res := runRoutine(t, it, f)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if captured != "first" {
		t.Errorf("captured: got %q, want first", captured)
	}
}

// TestSelectTimeout — when no case ready, the timeout case fires
// after the duration elapses.
func TestSelectTimeout(t *testing.T) {
	src := `
routine r() {
    select {
        when never() { note("never") }
        timeout 50ms { note("timeout") }
    }
    return "done"
}
`
	f := mustParseFile(t, src)

	var captured string
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.SelectTickOverride = 5 * time.Millisecond
	it.Builtins["never"] = &pureCallable{fn: func(_ []interp.Value) interp.Value { return interp.Bool(false) }}
	it.Builtins["note"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		if len(args) > 0 {
			captured = string(args[0].(interp.String))
		}
		return interp.Null{}, nil
	}}

	start := time.Now()
	res := runRoutine(t, it, f)
	elapsed := time.Since(start)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if captured != "timeout" {
		t.Errorf("captured: got %q, want timeout", captured)
	}
	if elapsed < 50*time.Millisecond || elapsed > 500*time.Millisecond {
		t.Errorf("elapsed %v: outside expected 50ms..500ms window", elapsed)
	}
}

// TestSelectOnEvent — bus event delivered while select is blocked
// fires the matching on-case with bound params.
func TestSelectOnEvent(t *testing.T) {
	src := `
routine r() {
    select {
        on chat_received(speaker, msg) { note(speaker + ": " + msg) }
        timeout 500ms { note("timeout") }
    }
    return "done"
}
`
	f := mustParseFile(t, src)

	var captured string
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.SelectTickOverride = 5 * time.Millisecond
	it.Builtins["note"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		if len(args) > 0 {
			captured = string(args[0].(interp.String))
		}
		return interp.Null{}, nil
	}}

	// Deliver event ~20ms after start.
	go func() {
		time.Sleep(20 * time.Millisecond)
		it.Events <- interp.PendingEvent{
			Name: "chat_received",
			Args: []interp.Value{interp.String("alex"), interp.String("hi")},
		}
	}()

	res := runRoutine(t, it, f)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: %v", res.Kind)
	}
	if captured != "alex: hi" {
		t.Errorf("captured: got %q, want \"alex: hi\"", captured)
	}
}

// runRoutine is a test helper that calls RunRoutine.
func runRoutine(t *testing.T, it *interp.Interpreter, file *ast.File) interp.Result {
	t.Helper()
	return it.RunRoutine(context.Background(), file, nil)
}
