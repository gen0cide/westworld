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

// runWithCaps parses, validates, and runs src with custom Caps.
// Returns the Result for assertion.
func runWithCaps(t *testing.T, src string, caps interp.Caps) interp.Result {
	t.Helper()
	file, err := parser.Parse("caps.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}
	it := interp.New()
	it.Caps = caps
	it.Rand = rand.New(rand.NewSource(1))
	return it.RunRoutine(context.Background(), file, nil)
}

func TestDefaultCapsAllowSmallRoutine(t *testing.T) {
	res := runWithCaps(t, `routine r() {
		x = 0
		while x < 100 { x = x + 1 }
		return x
	}`, interp.DefaultCaps())
	if res.Kind != interp.ResultReturned {
		t.Errorf("kind: got %v, want returned", res.Kind)
	}
}

func TestOpBudgetExhaustionAborts(t *testing.T) {
	caps := interp.DefaultCaps()
	caps.OpBudget = 200 // very small
	res := runWithCaps(t, `routine r() {
		x = 0
		while x < 10000 { x = x + 1 }
		return x
	}`, caps)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); !strings.Contains(string(s), "op_budget_exceeded") {
		t.Errorf("abort reason: got %v, want op_budget_exceeded", res.Value)
	}
}

func TestWallClockAborts(t *testing.T) {
	caps := interp.DefaultCaps()
	caps.WallClock = 50 * time.Millisecond
	// A tight pure-Go loop will eventually trip wall-clock or
	// op-budget. Make op budget huge so wall-clock fires first.
	caps.OpBudget = 100_000_000
	res := runWithCaps(t, `routine r() {
		x = 0
		while x < 1000000 { x = x + 1 }
		return x
	}`, caps)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); !strings.Contains(string(s), "wall_clock_exceeded") {
		t.Errorf("abort reason: got %v, want wall_clock_exceeded", res.Value)
	}
}

func TestRecursionDepthCap(t *testing.T) {
	caps := interp.DefaultCaps()
	caps.MaxRecursion = 8
	res := runWithCaps(t, `
		proc rec(n) {
			if n <= 0 { return 0 }
			return rec(n - 1) + 1
		}
		routine r() {
			return rec(100)
		}
	`, caps)
	// Recursion failure surfaces as a routine runtime error
	// (because proc.Call returns the *RuntimeError), which the
	// interpreter funnels into ResultErrored — UNLESS it bubbles
	// through a panic at a deeper layer. Either way, the routine
	// MUST NOT return a clean integer.
	if res.Kind == interp.ResultReturned {
		t.Fatalf("recursion did not trip cap: returned %v", res.Value)
	}
}

func TestListSizeCap(t *testing.T) {
	caps := interp.DefaultCaps()
	caps.MaxListLen = 3
	res := runWithCaps(t, `routine r() {
		return [1, 2, 3, 4, 5]
	}`, caps)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); !strings.Contains(string(s), "list_too_large") {
		t.Errorf("abort reason: got %v, want list_too_large", res.Value)
	}
}

func TestStringSizeCap(t *testing.T) {
	caps := interp.DefaultCaps()
	caps.MaxStringLen = 10
	// f-string interpolation of a long literal.
	res := runWithCaps(t, `routine r() {
		x = "this is a much longer string than the cap"
		return f"hello {x}"
	}`, caps)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); !strings.Contains(string(s), "string_too_large") {
		t.Errorf("abort reason: got %v, want string_too_large", res.Value)
	}
}

func TestCapsDoNotInterfereWithNormalRoutines(t *testing.T) {
	// Realistic routine with default caps should run without issue.
	res := runWithCaps(t, `
		proc helper(x) { return x * 2 }
		routine r() {
			total = 0
			for n in 1..50 {
				total = total + helper(n)
			}
			return total
		}
	`, interp.DefaultCaps())
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v (%v), want returned", res.Kind, res.Value)
	}
	// 2 * (1+2+...+50) = 2 * 1275 = 2550
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 2550 {
		t.Errorf("value: got %v, want Int(2550)", res.Value)
	}
}
