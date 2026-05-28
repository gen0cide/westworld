package interp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// resultCallable returns a *CallResult on call. Lets tests register
// fake actions that produce specific success/failure shapes.
type resultCallable struct {
	yields bool
	fn     func(args []interp.Value, named map[string]interp.Value) *interp.CallResult
}

func (r *resultCallable) Kind() string    { return "test-action" }
func (r *resultCallable) Display() string { return "<test-action>" }
func (r *resultCallable) Yields() bool    { return r.yields }
func (r *resultCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	return r.fn(args, named), nil
}

func TestErrorCodeStringScreamingSnake(t *testing.T) {
	cases := map[interp.ErrorCode]string{
		interp.PATH_BLOCKED:    "PATH_BLOCKED",
		interp.OUT_OF_RANGE:    "OUT_OF_RANGE",
		interp.INVENTORY_FULL:  "INVENTORY_FULL",
		interp.SERVER_REJECTED: "SERVER_REJECTED",
		interp.NOT_IMPLEMENTED: "NOT_IMPLEMENTED",
	}
	for code, want := range cases {
		if got := code.String(); got != want {
			t.Errorf("code %d: got %q, want %q", int(code), got, want)
		}
	}
}

func TestErrorGetFields(t *testing.T) {
	e := interp.NewFatalError(interp.OUT_OF_RANGE, "target too far")
	for field, want := range map[string]string{
		"code":   "OUT_OF_RANGE",
		"reason": "target too far",
	} {
		v, ok := e.Get(field)
		if !ok {
			t.Errorf("missing field %q", field)
			continue
		}
		if got := v.Display(); got != want {
			t.Errorf("field %q: got %q, want %q", field, got, want)
		}
	}
	fatal, _ := e.Get("fatal")
	if b, _ := fatal.(interp.Bool); !bool(b) {
		t.Error("fatal: got false, want true")
	}
}

func TestCallResultGetVal(t *testing.T) {
	r := interp.Ok(interp.Int(42))
	v, ok := r.Get("val")
	if !ok {
		t.Fatal("missing .val")
	}
	if i, _ := v.(interp.Int); int64(i) != 42 {
		t.Errorf("val: got %v, want Int(42)", v)
	}
	e, ok := r.Get("err")
	if !ok {
		t.Fatal("missing .err")
	}
	if _, isNull := e.(interp.Null); !isNull {
		t.Errorf("err: got %v, want Null on success", e)
	}
}

func TestCallResultGetErr(t *testing.T) {
	r := interp.Fail(interp.PATH_BLOCKED, "stuck")
	e, ok := r.Get("err")
	if !ok {
		t.Fatal("missing .err")
	}
	errVal, isErr := e.(*interp.Error)
	if !isErr {
		t.Fatalf("err: got %T, want *Error", e)
	}
	if errVal.Code != interp.PATH_BLOCKED {
		t.Errorf("err.code: got %v, want PATH_BLOCKED", errVal.Code)
	}
}

// ----- BangCallable behavior -----

func TestBangUnwrapsSuccess(t *testing.T) {
	src := `routine r() { return foo!() }`
	it := newInterpForTest()
	it.Builtins["foo"] = &resultCallable{
		yields: true,
		fn: func(args []interp.Value, _ map[string]interp.Value) *interp.CallResult {
			return interp.Ok(interp.String("hello"))
		},
	}
	it.Builtins["foo!"] = &interp.BangCallable{Underlying: it.Builtins["foo"], Name: "foo!"}
	res := runWithInterp(t, it, src)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "hello" {
		t.Errorf("value: got %v, want String(\"hello\")", res.Value)
	}
}

func TestBangAbortsOnFailure(t *testing.T) {
	src := `routine r() { return foo!() }`
	it := newInterpForTest()
	it.Builtins["foo"] = &resultCallable{
		yields: true,
		fn: func(args []interp.Value, _ map[string]interp.Value) *interp.CallResult {
			return interp.Fail(interp.PATH_BLOCKED, "stuck at the door")
		},
	}
	it.Builtins["foo!"] = &interp.BangCallable{Underlying: it.Builtins["foo"], Name: "foo!"}
	res := runWithInterp(t, it, src)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	e, ok := res.Value.(*interp.Error)
	if !ok {
		t.Fatalf("aborted value: got %T (%v), want *Error", res.Value, res.Value)
	}
	if e.Code != interp.PATH_BLOCKED {
		t.Errorf("err.code: got %v, want PATH_BLOCKED", e.Code)
	}
}

// ----- Panic containment -----

// panickingCallable always panics with an arbitrary Go value (not a
// control-flow signal). The interpreter must contain it without
// crashing the process.
type panickingCallable struct{}

func (p *panickingCallable) Kind() string    { return "panic-action" }
func (p *panickingCallable) Display() string { return "<panic-action>" }
func (p *panickingCallable) Yields() bool    { return true }
func (p *panickingCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	panic("intentional nil-style panic")
}

func TestUnknownPanicIsContained(t *testing.T) {
	src := `routine r() { return boom() }`
	it := newInterpForTest()
	it.Builtins["boom"] = &panickingCallable{}
	res := runWithInterp(t, it, src)
	if res.Kind != interp.ResultErrored {
		t.Fatalf("kind: got %v, want errored", res.Kind)
	}
	if res.Err == nil || !strings.Contains(res.Err.Msg, "interpreter panic") {
		t.Errorf("error: got %v, want msg containing \"interpreter panic\"", res.Err)
	}
}

func TestNilPointerStyleGoPanicContained(t *testing.T) {
	// Trigger a real Go nil deref via a callable that does it.
	src := `routine r() { return boom() }`
	it := newInterpForTest()
	it.Builtins["boom"] = &nilDerefCallable{}
	res := runWithInterp(t, it, src)
	if res.Kind != interp.ResultErrored {
		t.Fatalf("kind: got %v, want errored — got %v with %v", res.Kind, res.Value, res.Err)
	}
}

type nilDerefCallable struct{}

func (n *nilDerefCallable) Kind() string    { return "nil-action" }
func (n *nilDerefCallable) Display() string { return "<nil-action>" }
func (n *nilDerefCallable) Yields() bool    { return true }
func (n *nilDerefCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var p *int
	_ = *p // deliberate nil deref
	return interp.Null{}, nil
}

// ----- helpers -----

func newInterpForTest() *interp.Interpreter {
	it := interp.New()
	return it
}

// runWithInterp parses but does NOT validate — tests here register
// fake builtins that the validator's static table doesn't know
// about. The interpreter resolves names at runtime so the calls
// work; bypassing the validator just lets us exercise the bang /
// panic-containment paths without polluting the static builtins
// table with test-only entries.
func runWithInterp(t *testing.T, it *interp.Interpreter, src string) interp.Result {
	t.Helper()
	file, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return it.RunRoutine(context.Background(), file, nil)
}

// silence "imported and not used" if validator stops being referenced.
var _ = validator.Validate
