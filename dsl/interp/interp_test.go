package interp_test

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// run parses, validates, and executes src with optional reserved
// values and builtins. Returns the routine's Result.
func run(t *testing.T, src string, opts ...func(*interp.Interpreter)) interp.Result {
	t.Helper()
	file, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}
	it := interp.New()
	it.Rand = rand.New(rand.NewSource(1))
	for _, o := range opts {
		o(it)
	}
	return it.RunRoutine(context.Background(), file, nil)
}

// withBuiltin returns an opt that registers a single named callable.
func withBuiltin(name string, fn callableFunc) func(*interp.Interpreter) {
	return func(i *interp.Interpreter) {
		i.Builtins[name] = fn
	}
}

func withReserved(name string, v interp.Value) func(*interp.Interpreter) {
	return func(i *interp.Interpreter) {
		i.Reserved[name] = v
	}
}

// callableFunc is a function type that satisfies interp.Callable so
// tests can register builtins inline.
type callableFunc func(args []interp.Value, named map[string]interp.Value) (interp.Value, error)

func (c callableFunc) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	return c(args, named)
}

// fakeEntity implements interp.Getter so tests can supply `self` etc.
type fakeEntity struct{ fields map[string]interp.Value }

func (e *fakeEntity) Kind() string    { return "entity" }
func (e *fakeEntity) Display() string { return "<entity>" }
func (e *fakeEntity) Get(field string) (interp.Value, bool) {
	v, ok := e.fields[field]
	return v, ok
}

// ----- arithmetic + types -----

func TestIntArithmetic(t *testing.T) {
	got := run(t, `routine r() { return 1 + 2 * 3 - 4 }`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 3 {
		t.Errorf("got %v (%T), want Int(3)", got, got)
	}
}

func TestFloatPromotion(t *testing.T) {
	got := run(t, `routine r() { return 1 + 2.5 }`).Value
	if f, ok := got.(interp.Float); !ok || float64(f) != 3.5 {
		t.Errorf("got %v, want Float(3.5)", got)
	}
}

func TestIntDivisionPromotesToFloat(t *testing.T) {
	got := run(t, `routine r() { return 7 / 2 }`).Value
	if f, ok := got.(interp.Float); !ok || float64(f) != 3.5 {
		t.Errorf("got %v, want Float(3.5)", got)
	}
}

func TestModuloByZeroIsError(t *testing.T) {
	res := run(t, `routine r() { return 5 % 0 }`)
	if res.Kind != interp.ResultErrored {
		t.Errorf("got kind %v, want errored", res.Kind)
	}
	if res.Err == nil || !strings.Contains(res.Err.Msg, "modulo by zero") {
		t.Errorf("want modulo-by-zero error, got %v", res.Err)
	}
}

// ----- comparisons + booleans -----

func TestComparisonOps(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{`routine r() { return 1 < 2 }`, true},
		{`routine r() { return 2 <= 2 }`, true},
		{`routine r() { return 3 > 2 }`, true},
		{`routine r() { return 2 >= 3 }`, false},
		{`routine r() { return 1 == 1.0 }`, true},
		{`routine r() { return "abc" == "abc" }`, true},
		{`routine r() { return "abc" != "abd" }`, true},
		{`routine r() { return null == null }`, true},
	}
	for _, c := range cases {
		got := run(t, c.src).Value
		if b, ok := got.(interp.Bool); !ok || bool(b) != c.want {
			t.Errorf("%s → got %v, want %v", c.src, got, c.want)
		}
	}
}

func TestLogicalShortCircuit(t *testing.T) {
	res := run(t, `routine r() { return false and (1 / 0) }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("expected completion, got %v", res.Kind)
	}
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("got %v, want false", res.Value)
	}
}

func TestNotOperator(t *testing.T) {
	got := run(t, `routine r() { return not false }`).Value
	if b, ok := got.(interp.Bool); !ok || !bool(b) {
		t.Errorf("got %v, want true", got)
	}
}

func TestTruthinessRules(t *testing.T) {
	cases := map[string]bool{
		`routine r() { if 0 { return "yes" } return "no" }`:     false,
		`routine r() { if 1 { return "yes" } return "no" }`:     true,
		`routine r() { if "" { return "yes" } return "no" }`:    false,
		`routine r() { if "x" { return "yes" } return "no" }`:   true,
		`routine r() { if [] { return "yes" } return "no" }`:    false,
		`routine r() { if [1] { return "yes" } return "no" }`:   true,
		`routine r() { if null { return "yes" } return "no" }`:  false,
		`routine r() { if false { return "yes" } return "no" }`: false,
	}
	for src, wantTruthy := range cases {
		want := "no"
		if wantTruthy {
			want = "yes"
		}
		got := run(t, src).Value
		if s, ok := got.(interp.String); !ok || string(s) != want {
			t.Errorf("%s → got %v, want %q", src, got, want)
		}
	}
}

// ----- strings + f-strings -----

func TestStringConcat(t *testing.T) {
	got := run(t, `routine r() { return "foo" + "bar" }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "foobar" {
		t.Errorf("got %v, want String(foobar)", got)
	}
}

func TestFStringInterpolation(t *testing.T) {
	got := run(t, `routine r(name, gold) { return f"hi {name}, you have {gold} gp" }`)
	// RunRoutine doesn't accept args here; use require defaults via let-assignment.
	_ = got
	got = run(t, `routine r() {
		name = "alex"
		gold = 1234
		return f"hi {name}, you have {gold} gp"
	}`)
	if s, ok := got.Value.(interp.String); !ok || string(s) != "hi alex, you have 1234 gp" {
		t.Errorf("got %v, want fully-interpolated", got.Value)
	}
}

// ----- if / while / for -----

func TestIfElifElse(t *testing.T) {
	got := run(t, `routine r() {
		x = 5
		if x == 1 { return "one" }
		elif x == 5 { return "five" }
		else { return "other" }
	}`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "five" {
		t.Errorf("got %v, want five", got)
	}
}

func TestWhileCounter(t *testing.T) {
	got := run(t, `routine r() {
		i = 0
		while i < 5 { i = i + 1 }
		return i
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 5 {
		t.Errorf("got %v, want 5", got)
	}
}

func TestForOverList(t *testing.T) {
	got := run(t, `routine r() {
		total = 0
		for n in [1, 2, 3, 4] { total = total + n }
		return total
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 10 {
		t.Errorf("got %v, want 10", got)
	}
}

func TestForOverIntRange(t *testing.T) {
	got := run(t, `routine r() {
		total = 0
		for n in 1..5 { total = total + n }
		return total
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 15 {
		t.Errorf("got %v, want 15", got)
	}
}

func TestRepeatUntilExitsOnCondition(t *testing.T) {
	got := run(t, `routine r() {
		i = 0
		repeat {
			i = i + 1
		} until i >= 3 timeout 5
		return i
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 3 {
		t.Errorf("got %v, want 3", got)
	}
}

func TestRepeatUntilRunsBodyAtLeastOnce(t *testing.T) {
	// Even if the condition is true from the start, the body must
	// run once (do-while semantics).
	got := run(t, `routine r() {
		i = 0
		repeat {
			i = i + 1
		} until true timeout 5
		return i
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("got %v, want 1 (body must run at least once)", got)
	}
}

func TestRepeatUntilExitsOnTimeout(t *testing.T) {
	// Condition that never becomes true, short timeout. The body
	// uses `wait` (registered as a real sleep here) so we don't
	// burn op budget faster than the wall-clock timeout — this is
	// the realistic shape: a retry loop with a small backoff.
	waitFn := callableFunc(func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		secs, _ := interp.AsFloat(args[0])
		time.Sleep(time.Duration(secs * float64(time.Second)))
		return interp.Null{}, nil
	})
	start := time.Now()
	got := run(t, `routine r() {
		i = 0
		repeat {
			i = i + 1
			wait 0.05
		} until false timeout 0.3
		return i
	}`, withBuiltin("wait", waitFn)).Value
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("repeat_until didn't honor timeout — elapsed %v", elapsed)
	}
	if elapsed < 250*time.Millisecond {
		t.Errorf("repeat_until exited too fast — elapsed %v, want at least ~300ms", elapsed)
	}
	if _, ok := got.(interp.Int); !ok {
		t.Errorf("got %v, want Int", got)
	}
}

func TestRepeatUntilBreakOut(t *testing.T) {
	got := run(t, `routine r() {
		i = 0
		repeat {
			i = i + 1
			if i == 2 { break }
		} until i >= 100 timeout 5
		return i
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 2 {
		t.Errorf("got %v, want 2", got)
	}
}

func TestBreakOut(t *testing.T) {
	got := run(t, `routine r() {
		i = 0
		while i < 100 {
			if i == 3 { break }
			i = i + 1
		}
		return i
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 3 {
		t.Errorf("got %v, want 3", got)
	}
}

func TestContinueSkips(t *testing.T) {
	got := run(t, `routine r() {
		sum = 0
		for n in 1..10 {
			if n == 5 { continue }
			sum = sum + n
		}
		return sum
	}`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 50 {
		t.Errorf("got %v, want 50 (1..10 sum is 55, minus 5)", got)
	}
}

// ----- return / abort / require -----

func TestReturnValue(t *testing.T) {
	res := run(t, `routine r() { return "ok" }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, ok := res.Value.(interp.String); !ok || string(s) != "ok" {
		t.Errorf("value: got %v, want \"ok\"", res.Value)
	}
}

func TestAbortReason(t *testing.T) {
	res := run(t, `routine r() { abort "panic" }`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, ok := res.Value.(interp.String); !ok || string(s) != "panic" {
		t.Errorf("reason: got %v, want \"panic\"", res.Value)
	}
}

func TestRequirePass(t *testing.T) {
	res := run(t, `routine r() {
		require { 1 > 0 }
		return "ok"
	}`)
	if res.Kind != interp.ResultReturned {
		t.Errorf("kind: got %v, want returned", res.Kind)
	}
}

func TestRequireFail(t *testing.T) {
	res := run(t, `routine r() {
		require { 1 > 2 }
		return "ok"
	}`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); !strings.Contains(string(s), "precondition_failed") {
		t.Errorf("reason: got %v, want precondition_failed", res.Value)
	}
}

// ----- procs -----

func TestProcCall(t *testing.T) {
	got := run(t, `
		proc square(x) { return x * x }
		routine r() { return square(6) }
	`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 36 {
		t.Errorf("got %v, want 36", got)
	}
}

func TestProcWithDefaultArg(t *testing.T) {
	got := run(t, `
		proc greet(name, greeting = "hi") { return greeting + " " + name }
		routine r() { return greet("alex") }
	`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "hi alex" {
		t.Errorf("got %v, want \"hi alex\"", got)
	}
}

func TestProcCallsProc(t *testing.T) {
	got := run(t, `
		proc double(x) { return x * 2 }
		proc quad(x) { return double(double(x)) }
		routine r() { return quad(3) }
	`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 12 {
		t.Errorf("got %v, want 12", got)
	}
}

// ----- member + index access -----

func TestMemberAccessOnReservedEntity(t *testing.T) {
	self := &fakeEntity{fields: map[string]interp.Value{
		"hp":     interp.Int(50),
		"max_hp": interp.Int(100),
	}}
	got := run(t, `routine r() { return self.hp }`, withReserved("self", self)).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 50 {
		t.Errorf("got %v, want 50", got)
	}
}

func TestListIndexAccess(t *testing.T) {
	got := run(t, `routine r() { return [10, 20, 30][1] }`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 20 {
		t.Errorf("got %v, want 20", got)
	}
}

func TestListLength(t *testing.T) {
	got := run(t, `routine r() { return [1, 2, 3, 4].length }`).Value
	if i, ok := got.(interp.Int); !ok || int64(i) != 4 {
		t.Errorf("got %v, want 4", got)
	}
}

func TestStringIndex(t *testing.T) {
	got := run(t, `routine r() { return "hello"[1] }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "e" {
		t.Errorf("got %v, want \"e\"", got)
	}
}

// ----- builtins -----

func TestBuiltinCalled(t *testing.T) {
	var capturedArg interp.Value
	say := callableFunc(func(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
		if len(args) > 0 {
			capturedArg = args[0]
		}
		return interp.Null{}, nil
	})
	run(t, `routine r() { say("hello world") }`, withBuiltin("say", say))
	if s, ok := capturedArg.(interp.String); !ok || string(s) != "hello world" {
		t.Errorf("builtin received %v, want \"hello world\"", capturedArg)
	}
}

func TestBuiltinNamedArgs(t *testing.T) {
	var namedX, namedY interp.Value
	walk := callableFunc(func(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
		namedX = named["x"]
		namedY = named["y"]
		return interp.Null{}, nil
	})
	run(t, `routine r() { walk_to(x = 120, y = 504) }`, withBuiltin("walk_to", walk))
	if i, _ := namedX.(interp.Int); int64(i) != 120 {
		t.Errorf("x: got %v, want 120", namedX)
	}
	if i, _ := namedY.(interp.Int); int64(i) != 504 {
		t.Errorf("y: got %v, want 504", namedY)
	}
}

// ----- end-to-end -----

func TestRealisticFishingRoutineCompletes(t *testing.T) {
	// Mock everything the routine touches.
	self := &fakeEntity{fields: map[string]interp.Value{
		"wielded": interp.String("fishing-rod"),
		"fatigue": interp.Int(20),
	}}
	inventory := &fakeEntity{fields: map[string]interp.Value{
		"free": interp.Int(0), // start full so we break right away
	}}
	world := &fakeEntity{fields: map[string]interp.Value{
		"locs": &fakeEntity{fields: map[string]interp.Value{
			"fishing_spots": &fakeEntity{fields: map[string]interp.Value{
				"x": interp.Int(100),
				"y": interp.Int(200),
			}},
		}},
	}}
	fishCalled := false
	src := `
routine r() {
    require {
        self.wielded != null
        self.fatigue < 90
    }
    spot = world.locs.fishing_spots
    walk_to(x = spot.x, y = spot.y)
    while inventory.free > 0 {
        fish(spot)
    }
    return "banked"
}`
	res := run(t, src,
		withReserved("self", self),
		withReserved("world", world),
		withReserved("inventory", inventory),
		withBuiltin("walk_to", callableFunc(func(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
			return interp.Null{}, nil
		})),
		withBuiltin("fish", callableFunc(func(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
			fishCalled = true
			return interp.Null{}, nil
		})),
	)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "banked" {
		t.Errorf("return value: got %v, want \"banked\"", res.Value)
	}
	if fishCalled {
		t.Error("fish() should not have been called (inventory.free == 0)")
	}
}
