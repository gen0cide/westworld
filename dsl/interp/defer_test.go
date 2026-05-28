package interp_test

import (
	"context"
	"sync"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// captureCallable records every call's arguments — useful for
// asserting defer order and arg-capture semantics.
type captureCallable struct {
	mu    sync.Mutex
	calls [][]interp.Value
}

func (c *captureCallable) Kind() string    { return "capture" }
func (c *captureCallable) Display() string { return "<capture>" }
func (c *captureCallable) Yields() bool    { return false }
func (c *captureCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, args)
	return interp.Null{}, nil
}

// runDefer parses but does NOT validate — test routines reference
// `cleanup`/`boom` fakes that aren't in spec.Actions. The
// interpreter resolves names at runtime so the calls work; skipping
// the validator just avoids polluting the static spec with
// test-only entries. See dsl/interp/error_test.go for the same
// pattern.
func runDefer(t *testing.T, it *interp.Interpreter, src string) interp.Result {
	t.Helper()
	file, err := parser.Parse("d.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return it.RunRoutine(context.Background(), file, nil)
}

// silence "imported and not used" if validator stops being used.
var _ = validator.Validate

func TestDeferRunsOnNormalReturn(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	runDefer(t, it, `routine r() {
		defer cleanup(1)
		return "ok"
	}`)
	if len(cap.calls) != 1 {
		t.Fatalf("cleanup ran %d times, want 1", len(cap.calls))
	}
}

func TestDeferRunsOnAbort(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	res := runDefer(t, it, `routine r() {
		defer cleanup(99)
		abort "intentional"
	}`)
	if res.Kind != interp.ResultAborted {
		t.Errorf("kind: got %v, want aborted", res.Kind)
	}
	if len(cap.calls) != 1 {
		t.Fatalf("cleanup ran %d times, want 1 (defers fire on abort)", len(cap.calls))
	}
}

func TestDeferRunsLIFO(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	runDefer(t, it, `routine r() {
		defer cleanup(1)
		defer cleanup(2)
		defer cleanup(3)
		return "ok"
	}`)
	if len(cap.calls) != 3 {
		t.Fatalf("got %d calls, want 3", len(cap.calls))
	}
	// LIFO: last-deferred (3) ran first, first-deferred (1) ran last.
	want := []int64{3, 2, 1}
	for i, c := range cap.calls {
		if len(c) != 1 {
			t.Fatalf("call[%d] had %d args, want 1", i, len(c))
		}
		got, ok := c[0].(interp.Int)
		if !ok {
			t.Fatalf("call[%d] arg was %T, want Int", i, c[0])
		}
		if int64(got) != want[i] {
			t.Errorf("call[%d]: got %d, want %d", i, int64(got), want[i])
		}
	}
}

func TestDeferCapturesArgsAtDeferTime(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	// Go-style: x is captured at defer-time. The reassignment
	// after the defer must NOT change what cleanup sees.
	runDefer(t, it, `routine r() {
		x = 5
		defer cleanup(x)
		x = 10
		return "ok"
	}`)
	if len(cap.calls) != 1 {
		t.Fatalf("ran %d times, want 1", len(cap.calls))
	}
	got, _ := cap.calls[0][0].(interp.Int)
	if int64(got) != 5 {
		t.Errorf("arg captured: got %d, want 5 (defer captures at defer-time)", int64(got))
	}
}

func TestDeferRegisteredBeforeAbortRuns(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	// Only the first defer is registered before the abort; the
	// second defer never gets pushed because the abort interrupts
	// the body. Only the first should fire.
	runDefer(t, it, `routine r() {
		defer cleanup(1)
		abort "stop"
		defer cleanup(2)
	}`)
	if len(cap.calls) != 1 {
		t.Fatalf("got %d calls, want 1 (second defer never registered)", len(cap.calls))
	}
	got, _ := cap.calls[0][0].(interp.Int)
	if int64(got) != 1 {
		t.Errorf("arg: got %d, want 1", int64(got))
	}
}

func TestDeferRunsEvenOnRuntimeError(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	res := runDefer(t, it, `routine r() {
		defer cleanup(1)
		x = 1 / 0      # division by zero — RuntimeError
		return "ok"
	}`)
	if res.Kind != interp.ResultErrored {
		t.Errorf("kind: got %v, want errored", res.Kind)
	}
	if len(cap.calls) != 1 {
		t.Errorf("cleanup ran %d times, want 1 (defers fire on RuntimeError too)", len(cap.calls))
	}
}

func TestDeferOfPanickingCallableIsContained(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["cleanup"] = cap
	it.Builtins["boom"] = &panickingCallable{}
	// First-registered defer panics. Second (LIFO-second-to-run)
	// should still execute — defers are independent.
	runDefer(t, it, `routine r() {
		defer cleanup(1)
		defer boom()
		return "ok"
	}`)
	if len(cap.calls) != 1 {
		t.Errorf("cleanup ran %d times, want 1 (boom's panic shouldn't block cleanup)", len(cap.calls))
	}
}

func TestDeferOfNonCallRejectedAtParseOrEval(t *testing.T) {
	// `defer 42` — 42 isn't a call. Should either fail to parse
	// or fail at runtime. Either is acceptable; what matters is
	// the routine doesn't silently accept it.
	it := interp.New()
	res := runDefer(t, it, `routine r() { defer 42 }`)
	if res.Kind == interp.ResultReturned || res.Kind == interp.ResultCompleted {
		t.Errorf("expected error, got %v", res.Kind)
	}
}
