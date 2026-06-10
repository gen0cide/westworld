package interp_test

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
)

// pointGetter is a minimal entity-like Value exposing .x/.y/.tag —
// stands in for the npc/player/ground-item views that real routines
// call list.nearest on. tag lets the test identify which element won.
type pointGetter struct {
	x, y int
	tag  string
}

func (p *pointGetter) Kind() string    { return "point" }
func (p *pointGetter) Display() string { return p.tag }
func (p *pointGetter) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(int64(p.x)), true
	case "y":
		return interp.Int(int64(p.y)), true
	case "tag":
		return interp.String(p.tag), true
	}
	return nil, false
}

func runNearest(t *testing.T, src string) interp.Result {
	return runNearestReach(t, src, nil)
}

// runNearestReach is runNearest with an optional Reachable hook — the
// runtime-bridge seam list.nearest's default reachable-only filtering
// hangs off.
func runNearestReach(t *testing.T, src string, reachable func(x, y int) bool) interp.Result {
	t.Helper()
	file, err := parser.Parse("n.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	it := interp.New()
	it.Reachable = reachable
	// `points()` returns a list of point getters at fixed tiles;
	// `here()` / `there()` return reference positions.
	it.Builtins["points"] = &pureCallable{fn: func(_ []interp.Value) interp.Value {
		return &interp.List{Items: []interp.Value{
			&pointGetter{x: 10, y: 10, tag: "far"},
			&pointGetter{x: 3, y: 4, tag: "near"},
			&pointGetter{x: 100, y: 100, tag: "faraway"},
		}}
	}}
	it.Builtins["here"] = &pureCallable{fn: func(_ []interp.Value) interp.Value {
		return &pointGetter{x: 0, y: 0, tag: "origin"}
	}}
	return it.RunRoutine(context.Background(), file, nil)
}

// TestListNearestByPosition — .nearest(pos) picks the Chebyshev-closest
// element, reading .x/.y off each list item and off the reference.
func TestListNearestByPosition(t *testing.T) {
	res := runNearest(t, `routine r() {
		p = points().nearest(here())
		return p.tag
	}`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "near" {
		t.Errorf("nearest(pos): got %v, want String(near)", res.Value)
	}
}

// TestListNearestByXY — the two-int form recenters the search.
func TestListNearestByXY(t *testing.T) {
	// From (100,100) the "faraway" point at (100,100) is closest.
	res := runNearest(t, `routine r() {
		p = points().nearest(100, 100)
		return p.tag
	}`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "faraway" {
		t.Errorf("nearest(x,y): got %v, want String(faraway)", res.Value)
	}
}

// TestListNearestAfterFilter — the documented filter(...).nearest(...)
// chain: narrow then pick closest survivor.
func TestListNearestAfterFilter(t *testing.T) {
	// Exclude the "near" point; closest survivor to origin is "far" (10,10).
	res := runNearest(t, `routine r() {
		p = points().filter(q => q.tag != "near").nearest(here())
		return p.tag
	}`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "far" {
		t.Errorf("filter().nearest(): got %v, want String(far)", res.Value)
	}
}

// TestListNearestEmptyIsNull — nearest on an empty list returns Null
// (so `... .nearest(self.position) != null` guards work).
func TestListNearestEmptyIsNull(t *testing.T) {
	res := runNearest(t, `routine r() {
		p = points().filter(q => false).nearest(here())
		return p
	}`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("empty nearest: got %v, want Null", res.Value)
	}
}

// TestListNearestReachableDefault — with the runtime Reachable hook
// installed, list.nearest skips unreachable elements by default: from
// (100,100) the coincident "faraway" point is gated out, so the
// closest REACHABLE survivor ("far" at (10,10)) wins.
func TestListNearestReachableDefault(t *testing.T) {
	gate := func(x, y int) bool { return x < 50 } // "faraway" (100,100) is negative space
	res := runNearestReach(t, `routine r() {
		p = points().nearest(100, 100)
		return p.tag
	}`, gate)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "far" {
		t.Errorf("gated nearest: got %v, want String(far)", res.Value)
	}
}

// TestListNearestReachableOptOut — reachable=false restores the
// omniscient pick even with the hook installed.
func TestListNearestReachableOptOut(t *testing.T) {
	gate := func(x, y int) bool { return x < 50 }
	res := runNearestReach(t, `routine r() {
		p = points().nearest(100, 100, reachable=false)
		return p.tag
	}`, gate)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "faraway" {
		t.Errorf("opt-out nearest: got %v, want String(faraway)", res.Value)
	}
}

// TestListNearestNilHookUnfiltered — no hook (tests, oracle-less
// hosts) means no filtering: the pre-DSL-1 behavior is preserved.
func TestListNearestNilHookUnfiltered(t *testing.T) {
	res := runNearest(t, `routine r() {
		p = points().nearest(100, 100)
		return p.tag
	}`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "faraway" {
		t.Errorf("nil-hook nearest: got %v, want String(faraway)", res.Value)
	}
}
