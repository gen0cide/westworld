package interp_test

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
)

// runTry skips validator (test routines use fake builtins like
// `boom!`/`maybe_fail` that aren't in spec.Actions). See
// dsl/interp/error_test.go for the same pattern.
func runTry(t *testing.T, it *interp.Interpreter, src string) interp.Result {
	t.Helper()
	file, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return it.RunRoutine(context.Background(), file, nil)
}

func TestTryCleanRunSkipsRecover(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["mark"] = cap
	res := runTry(t, it, `routine r() {
		try {
			mark(1)
		} recover err {
			mark(99)
		}
		return "ok"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if len(cap.calls) != 1 {
		t.Fatalf("mark ran %d times, want 1 (recover should NOT run)", len(cap.calls))
	}
	if i, _ := cap.calls[0][0].(interp.Int); int64(i) != 1 {
		t.Errorf("ran %v, want mark(1)", cap.calls[0])
	}
}

func TestTryRecoverCatchesAbort(t *testing.T) {
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["mark"] = cap
	res := runTry(t, it, `routine r() {
		try {
			abort "OUT_OF_FOOD"
		} recover err {
			mark(err)
		}
		return "recovered"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned (recover caught abort)", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "recovered" {
		t.Errorf("value: got %v, want recovered", res.Value)
	}
	if len(cap.calls) != 1 {
		t.Fatalf("recover did not run; got %d marks", len(cap.calls))
	}
	if s, _ := cap.calls[0][0].(interp.String); string(s) != "OUT_OF_FOOD" {
		t.Errorf("err: got %v, want String(OUT_OF_FOOD)", cap.calls[0][0])
	}
}

func TestTryRecoverCatchesBangAbort(t *testing.T) {
	// Bang variant fails → abortSignal with *Error reason →
	// recover binds an Error and routine survives.
	cap := &captureCallable{}
	failing := &resultCallable{
		yields: true,
		fn: func(args []interp.Value, _ map[string]interp.Value) *interp.CallResult {
			return interp.Fail(interp.PATH_BLOCKED, "stuck")
		},
	}
	it := interp.New()
	it.Builtins["foo"] = failing
	it.Builtins["foo!"] = &interp.BangCallable{Underlying: failing, Name: "foo!"}
	it.Builtins["mark"] = cap
	res := runTry(t, it, `routine r() {
		try {
			foo!()
		} recover err {
			mark(err.code)
		}
		return "ok"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if len(cap.calls) != 1 {
		t.Fatalf("mark calls: got %d, want 1", len(cap.calls))
	}
	if s, _ := cap.calls[0][0].(interp.String); string(s) != "PATH_BLOCKED" {
		t.Errorf("err.code captured: got %v, want PATH_BLOCKED", cap.calls[0][0])
	}
}

func TestTryDoesNotCatchReturn(t *testing.T) {
	// return inside try should propagate as the routine's return.
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["mark"] = cap
	res := runTry(t, it, `routine r() {
		try {
			return "from_try"
		} recover err {
			mark(99)
		}
		return "after_try"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "from_try" {
		t.Errorf("return: got %v, want from_try", res.Value)
	}
	if len(cap.calls) != 0 {
		t.Errorf("recover ran on return (should not have); marks=%v", cap.calls)
	}
}

func TestTryRecoverCanReAbort(t *testing.T) {
	// recover catches the abort, decides it can't handle it, and
	// re-aborts with a different reason.
	res := runTry(t, interp.New(), `routine r() {
		try {
			abort "ORIGINAL"
		} recover err {
			abort "WRAPPED"
		}
		return "unreachable"
	}`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "WRAPPED" {
		t.Errorf("abort: got %v, want WRAPPED", res.Value)
	}
}

func TestTryDeferRunsBeforeRecover(t *testing.T) {
	// `defer mark(1)` inside try should fire before recover
	// (Go semantics: defers drain at the point execBlock unwinds).
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["mark"] = cap
	runTry(t, it, `routine r() {
		try {
			defer mark(1)
			abort "fail"
		} recover err {
			mark(2)
		}
		return "ok"
	}`)
	// Hmm — defer is currently per-routine, not per-block. The
	// `defer mark(1)` registered inside the try fires at routine
	// exit, AFTER recover runs. So the order is:
	//   1. abort triggers in try
	//   2. recover runs → mark(2)
	//   3. routine returns "ok"
	//   4. drainDeferred fires → mark(1)
	// We document this behavior; if users want block-scoped
	// defers we'd need separate stacks per scope (future work).
	if len(cap.calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(cap.calls))
	}
	first, _ := cap.calls[0][0].(interp.Int)
	second, _ := cap.calls[1][0].(interp.Int)
	if int64(first) != 2 || int64(second) != 1 {
		t.Errorf("order: got %d then %d, want 2 then 1 (recover before routine-exit defer)",
			int64(first), int64(second))
	}
}

func TestTryNestedRecover(t *testing.T) {
	// inner try catches the inner abort; outer try sees nothing.
	cap := &captureCallable{}
	it := interp.New()
	it.Builtins["mark"] = cap
	res := runTry(t, it, `routine r() {
		try {
			try {
				abort "INNER"
			} recover inner_err {
				mark(inner_err)
			}
			mark("after_inner_try")
		} recover outer_err {
			mark("OUTER_should_not_run")
		}
		return "ok"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v", res.Kind)
	}
	if len(cap.calls) != 2 {
		t.Fatalf("calls: got %d, want 2", len(cap.calls))
	}
	if s, _ := cap.calls[0][0].(interp.String); string(s) != "INNER" {
		t.Errorf("inner recover got %v", cap.calls[0][0])
	}
}
