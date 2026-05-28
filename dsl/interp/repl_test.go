package interp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

func TestSessionEvaluatesExpression(t *testing.T) {
	it := interp.New()
	sess := it.NewSession(context.Background(), "<test>")
	r := sess.Eval(context.Background(), "1 + 2 * 3")
	if r.Err != nil {
		t.Fatalf("eval: %v", r.Err)
	}
	if !r.IsExpression {
		t.Error("IsExpression: got false, want true")
	}
	if i, ok := r.Value.(interp.Int); !ok || int64(i) != 7 {
		t.Errorf("value: got %v, want Int(7)", r.Value)
	}
}

func TestSessionExecutesStatement(t *testing.T) {
	it := interp.New()
	sess := it.NewSession(context.Background(), "<test>")
	r := sess.Eval(context.Background(), "x = 42")
	if r.Err != nil {
		t.Fatalf("eval: %v", r.Err)
	}
	if r.IsExpression {
		t.Error("IsExpression: got true, want false (was an assignment)")
	}
	// Then read x back as an expression.
	r2 := sess.Eval(context.Background(), "x")
	if r2.Err != nil {
		t.Fatalf("eval x: %v", r2.Err)
	}
	if i, ok := r2.Value.(interp.Int); !ok || int64(i) != 42 {
		t.Errorf("x: got %v, want 42 (persistent env)", r2.Value)
	}
}

func TestSessionPersistentEnvAcrossLines(t *testing.T) {
	it := interp.New()
	sess := it.NewSession(context.Background(), "<test>")
	for i, line := range []string{
		"x = 5",
		"y = 10",
		"z = x + y",
	} {
		r := sess.Eval(context.Background(), line)
		if r.Err != nil {
			t.Fatalf("line %d %q: %v", i, line, r.Err)
		}
	}
	r := sess.Eval(context.Background(), "z")
	if i, ok := r.Value.(interp.Int); !ok || int64(i) != 15 {
		t.Errorf("z: got %v, want 15", r.Value)
	}
}

func TestSessionParseErrorSurvives(t *testing.T) {
	it := interp.New()
	sess := it.NewSession(context.Background(), "<test>")
	r := sess.Eval(context.Background(), "this is not valid syntax !!!")
	if r.Err == nil {
		t.Fatal("expected parse error")
	}
	// Session should still work after the failure.
	r2 := sess.Eval(context.Background(), "42")
	if r2.Err != nil {
		t.Errorf("post-error eval failed: %v", r2.Err)
	}
	if i, _ := r2.Value.(interp.Int); int64(i) != 42 {
		t.Errorf("recovery: got %v, want 42", r2.Value)
	}
}

func TestSessionStrayReturnDoesNotCrash(t *testing.T) {
	it := interp.New()
	sess := it.NewSession(context.Background(), "<test>")
	r := sess.Eval(context.Background(), "return 5")
	if r.Err == nil {
		t.Fatal("expected error for return outside routine")
	}
	if !strings.Contains(r.Err.Error(), "return outside") {
		t.Errorf("error: got %q, want substring \"return outside\"", r.Err.Error())
	}
	// Session continues working.
	r2 := sess.Eval(context.Background(), "1 + 1")
	if r2.Err != nil {
		t.Errorf("post-stray-return: %v", r2.Err)
	}
}

func TestSessionPanicContained(t *testing.T) {
	it := interp.New()
	it.Builtins["boom"] = &panickingCallable{}
	sess := it.NewSession(context.Background(), "<test>")
	r := sess.Eval(context.Background(), "boom()")
	if r.Err == nil {
		t.Fatal("expected session-contained error from panicking callable")
	}
	if !strings.Contains(r.Err.Error(), "session panic") &&
		!strings.Contains(r.Err.Error(), "panic") {
		t.Errorf("error %q should mention panic containment", r.Err.Error())
	}
	// Session survives.
	r2 := sess.Eval(context.Background(), "7 * 6")
	if r2.Err != nil {
		t.Errorf("post-panic: %v", r2.Err)
	}
	if i, _ := r2.Value.(interp.Int); int64(i) != 42 {
		t.Errorf("recovery: got %v, want 42", r2.Value)
	}
}
