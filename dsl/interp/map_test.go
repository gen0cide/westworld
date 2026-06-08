package interp_test

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// These tests cover field/index access on *interp.Map — the value shape that
// host actions like search_map / reachable hand back to a routine. The bug they
// guard against: *interp.Map implemented neither Getter nor Indexer, so `hit.x`
// / `hit["reach"]` (and the documented search_map consumption pattern) failed at
// runtime with "map does not support field access" / "does not support indexing".

func mapVal() *interp.Map {
	return &interp.Map{Items: map[string]interp.Value{
		"reach":   interp.String("open"),
		"x":       interp.Int(74),
		"y":       interp.Int(583),
		"payable": interp.Bool(true),
	}}
}

func mustReturn(t *testing.T, res interp.Result) {
	t.Helper()
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind=%v err=%v, want ResultReturned", res.Kind, res.Err)
	}
}

// TestMapFieldAccess: `m.key` reads a map entry.
func TestMapFieldAccess(t *testing.T) {
	res := run(t, `routine r() { return self.reach }`, withReserved("self", mapVal()))
	mustReturn(t, res)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "open" {
		t.Fatalf("self.reach = %v (%T), want String(\"open\")", res.Value, res.Value)
	}
}

// TestMapIndexAccess: `m["key"]` reads a map entry.
func TestMapIndexAccess(t *testing.T) {
	res := run(t, `routine r() { return self["x"] }`, withReserved("self", mapVal()))
	mustReturn(t, res)
	if i, ok := interp.AsInt(res.Value); !ok || i != 74 {
		t.Fatalf(`self["x"] = %v (%T), want 74`, res.Value, res.Value)
	}
}

// TestMapFieldAndIndexAgree: dotted and subscript access return the same value.
func TestMapFieldAndIndexAgree(t *testing.T) {
	res := run(t, `routine r() { return self.payable == self["payable"] }`, withReserved("self", mapVal()))
	mustReturn(t, res)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Fatalf("self.payable == self[\"payable\"] = %v, want true", res.Value)
	}
}

// TestMapMissingFieldErrors: an absent key surfaces a clean runtime error, not a
// silent null — matching every other Getter.
func TestMapMissingFieldErrors(t *testing.T) {
	res := run(t, `routine r() { return self.nope }`, withReserved("self", mapVal()))
	if res.Kind != interp.ResultErrored {
		t.Fatalf("kind=%v, want ResultErrored for a missing map field", res.Kind)
	}
}

// TestListOfMapsFindAndAccess: the search_map consumption pattern — a list of
// maps, filtered with .find on a per-entry field, then a field read off the hit.
func TestListOfMapsFindAndAccess(t *testing.T) {
	list := &interp.List{Items: []interp.Value{
		&interp.Map{Items: map[string]interp.Value{"reach": interp.String("blocked"), "x": interp.Int(10)}},
		&interp.Map{Items: map[string]interp.Value{"reach": interp.String("open"), "x": interp.Int(42)}},
	}}
	src := `routine r() {
		open = self.find(h => h.reach == "open")
		return open.x
	}`
	res := run(t, src, withReserved("self", list))
	mustReturn(t, res)
	if i, ok := interp.AsInt(res.Value); !ok || i != 42 {
		t.Fatalf("open.x = %v (%T), want 42", res.Value, res.Value)
	}
}

// TestListOfMapsIndexThenField: `list[0].field` — indexing the list then reading
// a field off the map entry (the `hits[0].gate` shape in the search_map manual).
func TestListOfMapsIndexThenField(t *testing.T) {
	list := &interp.List{Items: []interp.Value{
		&interp.Map{Items: map[string]interp.Value{"gate": interp.String("Toll gate"), "x": interp.Int(74)}},
	}}
	res := run(t, `routine r() { return self[0].gate }`, withReserved("self", list))
	mustReturn(t, res)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "Toll gate" {
		t.Fatalf("self[0].gate = %v (%T), want String(\"Toll gate\")", res.Value, res.Value)
	}
}

// TestSearchMapConsumptionPattern: the exact documented search_map pattern end
// to end — a builtin returns a RESULT wrapping a list of maps, the routine reads
// .val, .find()s the open hit on a per-entry field, and go_to's its coords. This
// is the consumption the dslmanual example teaches; it could not run before the
// *interp.Map Getter/Indexer fix.
func TestSearchMapConsumptionPattern(t *testing.T) {
	searchMap := callableFunc(func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		return interp.Ok(&interp.List{Items: []interp.Value{
			&interp.Map{Items: map[string]interp.Value{"reach": interp.String("blocked"), "x": interp.Int(10), "y": interp.Int(20), "gate": interp.String("Toll gate")}},
			&interp.Map{Items: map[string]interp.Value{"reach": interp.String("open"), "x": interp.Int(42), "y": interp.Int(43), "gate": interp.String("")}},
		}}), nil
	})
	var wentTo [2]int
	goTo := callableFunc(func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		x, _ := interp.AsInt(args[0])
		y, _ := interp.AsInt(args[1])
		wentTo = [2]int{int(x), int(y)}
		return interp.Ok(interp.Null{}), nil
	})
	note := callableFunc(func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		return interp.Null{}, nil
	})
	src := `routine r() {
		res = search_map("mining-site")
		if res.err == null {
			hits = res.val
			open = hits.find(h => h.reach == "open")
			if open != null { go_to(open.x, open.y) } else { note("gated: " + hits[0].gate) }
		}
	}`
	res := run(t, src,
		withBuiltin("search_map", searchMap),
		withBuiltin("go_to", goTo),
		withBuiltin("note", note))
	if res.Kind == interp.ResultErrored || res.Kind == interp.ResultAborted {
		t.Fatalf("routine failed: kind=%v err=%v", res.Kind, res.Err)
	}
	if wentTo != [2]int{42, 43} {
		t.Fatalf("go_to called with %v, want [42 43] (the open mine)", wentTo)
	}
}

// TestForOverListOfMaps: iterate a list of maps and read a field off each — the
// scan_for / search_map iterate-and-prune loop.
func TestForOverListOfMaps(t *testing.T) {
	list := &interp.List{Items: []interp.Value{
		&interp.Map{Items: map[string]interp.Value{"x": interp.Int(1)}},
		&interp.Map{Items: map[string]interp.Value{"x": interp.Int(2)}},
		&interp.Map{Items: map[string]interp.Value{"x": interp.Int(3)}},
	}}
	src := `routine r() {
		total = 0
		for h in self { total = total + h.x }
		return total
	}`
	res := run(t, src, withReserved("self", list))
	mustReturn(t, res)
	if i, ok := interp.AsInt(res.Value); !ok || i != 6 {
		t.Fatalf("sum of h.x = %v (%T), want 6", res.Value, res.Value)
	}
}
