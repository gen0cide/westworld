package runtime

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// scanCall runs dslScanFor and asserts it returned a BARE *interp.List (NOT a
// CallResult — so routines can `for r in scan_for(...)` / index / .first
// directly) with no Go error.
func scanCall(t *testing.T, h *Host, args []interp.Value, named map[string]interp.Value) *interp.List {
	t.Helper()
	v, err := dslScanFor(t.Context(), h, args, named)
	if err != nil {
		t.Fatalf("scan_for returned a Go error: %v", err)
	}
	list, ok := v.(*interp.List)
	if !ok {
		t.Fatalf("scan_for returned %T, want *interp.List", v)
	}
	return list
}

// pvField reads a field off a scan_for entry, asserting the entry is
// field-accessible (a Getter) — the contract that distinguishes scan_for's
// *placementView entries from search_map's non-field-accessible *interp.Map.
func pvField(t *testing.T, item interp.Value, field string) interp.Value {
	t.Helper()
	g, ok := item.(interp.Getter)
	if !ok {
		t.Fatalf("scan_for entry %T is not field-accessible (not a Getter)", item)
	}
	v, ok := g.Get(field)
	if !ok {
		t.Fatalf("scan_for entry has no field %q", field)
	}
	return v
}

func pvInt(t *testing.T, item interp.Value, field string) int {
	t.Helper()
	i, ok := interp.AsInt(pvField(t, item, field))
	if !ok {
		t.Fatalf("scan_for entry field %q is not int", field)
	}
	return int(i)
}

func scanContainsTile(list *interp.List, x, y int) bool {
	for _, item := range list.Items {
		g, ok := item.(interp.Getter)
		if !ok {
			continue
		}
		gx, _ := g.Get("x")
		gy, _ := g.Get("y")
		ix, _ := interp.AsInt(gx)
		iy, _ := interp.AsInt(gy)
		if int(ix) == x && int(iy) == y {
			return true
		}
	}
	return false
}

// firstSceneryTypeNear discovers a scenery type name that actually exists within
// `radius` of the fixture position, so the tests don't hardcode an assumption
// about which scenery sits at Lumbridge.
func firstSceneryTypeNear(h *Host, radius int) (facts.Placement, bool) {
	pos := h.world.Self.Position()
	for _, p := range h.facts.Near(pos.X, pos.Y, radius) {
		if p.Kind == "scenery" && p.Name != "" {
			return p, true
		}
	}
	return facts.Placement{}, false
}

// TestScanForFindsStaticScenery: scan_for surfaces the real static scenery near
// the host, and every entry is field-accessible and of the requested type.
func TestScanForFindsStaticScenery(t *testing.T) {
	h := newOracleHost(t, 0)
	seed, ok := firstSceneryTypeNear(h, 10)
	if !ok {
		t.Skip("no static scenery within 10 tiles of the fixture position")
	}
	list := scanCall(t, h, []interp.Value{interp.String(seed.Name)}, nil)
	if len(list.Items) == 0 {
		t.Fatalf("scan_for(%q) returned empty though facts.Near reported that scenery", seed.Name)
	}
	low := strings.ToLower(seed.Name)
	for _, item := range list.Items {
		name := pvField(t, item, "name").Display()
		if !strings.Contains(strings.ToLower(name), low) {
			t.Fatalf("scan_for(%q) leaked a %q entry — type filter wrong", seed.Name, name)
		}
		_, _ = pvInt(t, item, "x"), pvInt(t, item, "y") // must be field-accessible
		if k := pvField(t, item, "kind").Display(); k != "scenery" {
			t.Fatalf("entry kind=%q, want scenery", k)
		}
	}
	t.Logf("scan_for(%q) => %d hits", seed.Name, len(list.Items))
}

// TestScanForSortedByDistance: the list is ranked nearest-first.
func TestScanForSortedByDistance(t *testing.T) {
	h := newOracleHost(t, 0)
	seed, ok := firstSceneryTypeNear(h, 15)
	if !ok {
		t.Skip("no static scenery near fixture")
	}
	pos := h.world.Self.Position()
	list := scanCall(t, h, []interp.Value{interp.String(seed.Name), interp.Int(15)}, nil)
	prev := -1
	for i, item := range list.Items {
		d := chebyshev(pos.X, pos.Y, pvInt(t, item, "x"), pvInt(t, item, "y"))
		if i > 0 && d < prev {
			t.Fatalf("entry %d dist=%d < previous %d — not nearest-first", i, d, prev)
		}
		prev = d
	}
}

// TestScanForRadiusArg: a small radius is a subset of a large one, positional
// and named radius agree, and every narrow hit is within the radius.
func TestScanForRadiusArg(t *testing.T) {
	h := newOracleHost(t, 0)
	seed, ok := firstSceneryTypeNear(h, 20)
	if !ok {
		t.Skip("no static scenery near fixture")
	}
	pos := h.world.Self.Position()
	wide := scanCall(t, h, []interp.Value{interp.String(seed.Name), interp.Int(20)}, nil)
	narrow := scanCall(t, h, []interp.Value{interp.String(seed.Name), interp.Int(2)}, nil)
	if len(narrow.Items) > len(wide.Items) {
		t.Fatalf("radius=2 returned more hits (%d) than radius=20 (%d)", len(narrow.Items), len(wide.Items))
	}
	for _, item := range narrow.Items {
		if d := chebyshev(pos.X, pos.Y, pvInt(t, item, "x"), pvInt(t, item, "y")); d > 2 {
			t.Fatalf("radius=2 returned a hit %d tiles away", d)
		}
	}
	named := scanCall(t, h, []interp.Value{interp.String(seed.Name)}, map[string]interp.Value{"radius": interp.Int(2)})
	if len(named.Items) != len(narrow.Items) {
		t.Fatalf("named radius=2 (%d hits) != positional radius=2 (%d hits)", len(named.Items), len(narrow.Items))
	}
}

// TestScanForEmptyResult: an unmatched type returns an empty list, not a failure.
func TestScanForEmptyResult(t *testing.T) {
	h := newOracleHost(t, 0)
	list := scanCall(t, h, []interp.Value{interp.String("zzz-no-such-scenery")}, nil)
	if len(list.Items) != 0 {
		t.Fatalf("scan_for(nonsense) returned %d items, want empty list", len(list.Items))
	}
}

// TestScanForNoFactsDegrades: with no facts and an empty live mirror, scan_for
// returns an empty list and never panics (does NOT need the OpenRSC fixture).
func TestScanForNoFactsDegrades(t *testing.T) {
	h := New(Options{Username: "bare"})
	t.Cleanup(func() { h.Close() })
	list := scanCall(t, h, []interp.Value{interp.String("rock")}, nil)
	if len(list.Items) != 0 {
		t.Fatalf("scan_for with no facts returned %d items, want empty", len(list.Items))
	}
}

// TestScanForMergesLiveScenery: a runtime-spawned (live) scenery record is
// included alongside the static map.
func TestScanForMergesLiveScenery(t *testing.T) {
	h := newOracleHost(t, 0)
	pos := h.world.Self.Position()
	var defID int
	var defName string
	for id, def := range h.facts.SceneryDefs {
		if def != nil && def.Name != "" {
			defID, defName = id, def.Name
			break
		}
	}
	if defName == "" {
		t.Skip("facts has no named scenery defs")
	}
	lx, ly := pos.X+2, pos.Y // within the default radius
	h.world.Scenery.Add(lx, ly, defID)
	list := scanCall(t, h, []interp.Value{interp.String(defName)}, nil)
	if !scanContainsTile(list, lx, ly) {
		t.Fatalf("scan_for(%q) did not include the live record at (%d,%d) among %d hits", defName, lx, ly, len(list.Items))
	}
}

// TestScanForRemovedSceneryDropsOut: a static tile the server explicitly cleared
// (depleted rock / burned fire) must not be reported — depletion awareness.
func TestScanForRemovedSceneryDropsOut(t *testing.T) {
	h := newOracleHost(t, 0)
	target, ok := firstSceneryTypeNear(h, 10)
	if !ok {
		t.Skip("no static scenery near fixture")
	}
	before := scanCall(t, h, []interp.Value{interp.String(target.Name)}, nil)
	if !scanContainsTile(before, target.X, target.Y) {
		t.Fatalf("baseline scan_for(%q) missing static tile (%d,%d)", target.Name, target.X, target.Y)
	}
	h.world.Scenery.Remove(target.X, target.Y) // server says it's gone
	after := scanCall(t, h, []interp.Value{interp.String(target.Name)}, nil)
	if scanContainsTile(after, target.X, target.Y) {
		t.Fatalf("scan_for still returned removed/depleted tile (%d,%d)", target.X, target.Y)
	}
}
