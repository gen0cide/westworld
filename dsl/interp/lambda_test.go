package interp_test

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
)

func runLambda(t *testing.T, src string) interp.Result {
	t.Helper()
	file, err := parser.Parse("l.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return interp.New().RunRoutine(context.Background(), file, nil)
}

func TestLambdaSingleArg(t *testing.T) {
	res := runLambda(t, `routine r() {
		double = n => n * 2
		return double(21)
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 42 {
		t.Errorf("got %v, want Int(42)", res.Value)
	}
}

func TestLambdaClosesOverLocals(t *testing.T) {
	// The lambda captures x from the enclosing scope; reassigning
	// x after the lambda is defined doesn't change what the lambda
	// reads (Env walks ancestors, finds x=10 — wait, actually our
	// env doesn't snapshot on define, so reassignment WOULD change
	// what's read. That's lexical-scope-by-reference semantics —
	// closures see live bindings.) Document that this is the model.
	res := runLambda(t, `routine r() {
		x = 5
		add_x = n => n + x
		return add_x(3)
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 8 {
		t.Errorf("got %v, want Int(8)", res.Value)
	}
}

func TestListFilterWithLambda(t *testing.T) {
	res := runLambda(t, `routine r() {
		evens = [1, 2, 3, 4, 5, 6].filter(n => n % 2 == 0)
		return evens.length
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 3 {
		t.Errorf("filter length: got %v, want Int(3)", res.Value)
	}
}

func TestListMapWithLambda(t *testing.T) {
	res := runLambda(t, `routine r() {
		squared = [1, 2, 3].map(n => n * n)
		return squared[2]
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 9 {
		t.Errorf("map[2]: got %v, want Int(9)", res.Value)
	}
}

func TestListFindWithLambda(t *testing.T) {
	res := runLambda(t, `routine r() {
		hit = [10, 20, 30, 40].find(n => n > 25)
		return hit
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 30 {
		t.Errorf("find: got %v, want Int(30)", res.Value)
	}
}

func TestListFindMissingReturnsNull(t *testing.T) {
	res := runLambda(t, `routine r() {
		hit = [1, 2, 3].find(n => n > 100)
		return hit
	}`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("missing find: got %v, want Null", res.Value)
	}
}

func TestListFirstLast(t *testing.T) {
	// .first and .last are field accessors, not methods —
	// (Pick a verbose name like .first_item vs Go-style .First() —
	// going with field-access since lists are otherwise value-like.)
	res := runLambda(t, `routine r() {
		l = [11, 22, 33]
		return l.first + l.last
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 44 {
		t.Errorf("first+last: got %v, want Int(44)", res.Value)
	}
}

func TestListFilterChainedWithMap(t *testing.T) {
	res := runLambda(t, `routine r() {
		# Chain: evens of [1..6] = [2,4,6]; double each = [4,8,12]; sum length
		doubled = [1, 2, 3, 4, 5, 6].filter(n => n % 2 == 0).map(n => n * 2)
		return doubled.length
	}`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 3 {
		t.Errorf("chained: got %v, want Int(3)", res.Value)
	}
}
