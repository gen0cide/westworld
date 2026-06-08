package runtime

import (
	"os"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/world"
	"github.com/gen0cide/westworld/worldmap"
)

// openRSCRoot is the OpenRSC source root holding the real facts + landscape.
// These map-perception tests need the real archive to precompute the oracle;
// when it is not available we SKIP (unlike worldmap's own tests, which hard-
// fail — there the archive is the contract under test; here it is a fixture).
func openRSCRoot() string {
	if r := os.Getenv("OPENRSC_ROOT"); r != "" {
		return r
	}
	return "/Users/flint/Code/openrsc"
}

// newOracleHost builds a Host wired with the real facts + a precomputed
// WorldOracle, placed at Lumbridge, with `coins` gold in its inventory. It is
// the live-state fixture the search_map / reachable handlers read.
func newOracleHost(t *testing.T, coins int) *Host {
	t.Helper()
	root := openRSCRoot()
	if _, err := os.Stat(root); err != nil {
		t.Skipf("OpenRSC root %q not available (set OPENRSC_ROOT) — skipping map-perception wiring test", root)
	}
	f, err := facts.Load(facts.DefaultSources(root))
	if err != nil {
		t.Fatalf("load facts: %v", err)
	}
	lpath := root + "/server/conf/server/data/Authentic_Landscape.orsc"
	ls, err := pathfind.OpenLandscape(lpath)
	if err != nil {
		t.Fatalf("open landscape: %v", err)
	}
	t.Cleanup(func() { ls.Close() })
	oracle, err := worldmap.Precompute(f, ls)
	if err != nil {
		t.Fatalf("precompute oracle: %v", err)
	}

	h := New(Options{Username: "mapper", Facts: f, Landscape: ls, WorldOracle: oracle})
	t.Cleanup(func() { h.Close() })

	// Stand at Lumbridge (east of the Al-Kharid toll gate).
	h.world.Self.SetPosition(world.Coord{X: 135, Y: 654})

	// Put `coins` coins in the inventory at the live coin id.
	if coins > 0 {
		def := f.ItemDefByName("Coins")
		if def == nil {
			t.Fatalf("facts has no 'Coins' item def")
		}
		h.world.Inventory.Replace([]world.InvSlot{{ItemID: def.ID, Amount: coins}})
	}
	return h
}

// callList runs a handler that returns Ok(*interp.List) and returns the list.
func callList(t *testing.T, v interp.Value, err error) *interp.List {
	t.Helper()
	if err != nil {
		t.Fatalf("handler returned Go error: %v", err)
	}
	res, ok := v.(*interp.CallResult)
	if !ok {
		t.Fatalf("handler returned %T, want *interp.CallResult", v)
	}
	if res.Err != nil {
		t.Fatalf("handler failed: code=%s reason=%s", res.Err.Code, res.Err.Reason)
	}
	list, ok := res.Val.(*interp.List)
	if !ok {
		t.Fatalf("result value is %T, want *interp.List", res.Val)
	}
	return list
}

// hitField reads a string/int/bool field off a search_map result map entry.
func hitField(t *testing.T, m interp.Value, key string) interp.Value {
	t.Helper()
	mp, ok := m.(*interp.Map)
	if !ok {
		t.Fatalf("list entry is %T, want *interp.Map", m)
	}
	v, ok := mp.Items[key]
	if !ok {
		t.Fatalf("result map missing key %q (have %v)", key, mp.Items)
	}
	return v
}

// findMineHit returns the search_map entry whose label/coords match the
// Al-Kharid mine, or nil. The mine snaps near (74,583).
func findMineHit(t *testing.T, list *interp.List) *interp.Map {
	t.Helper()
	for _, it := range list.Items {
		mp := it.(*interp.Map)
		x, _ := interp.AsInt(mp.Items["x"])
		y, _ := interp.AsInt(mp.Items["y"])
		// The Al-Kharid mine sits west of the toll line (x ~ 74, y ~ 583).
		if x >= 60 && x <= 90 && y >= 570 && y <= 600 {
			return mp
		}
	}
	return nil
}

// TestSearchMapTollGateCoinless: a coinless host at Lumbridge searching for
// mining-sites must SEE the Al-Kharid mine as reach="blocked" by the toll gate,
// with its requirement and you_have=0 — the cognition-first contract.
func TestSearchMapTollGateCoinless(t *testing.T) {
	h := newOracleHost(t, 0)

	v, err := dslSearchMap(t.Context(), h, []interp.Value{interp.String("mining-site")}, nil)
	list := callList(t, v, err)
	if len(list.Items) == 0 {
		t.Fatalf("search_map returned no mining-site destinations")
	}
	mine := findMineHit(t, list)
	if mine == nil {
		t.Fatalf("search_map did not include the Al-Kharid mine among %d hits", len(list.Items))
	}
	reach := hitField(t, mine, "reach").Display()
	if reach != "blocked" {
		t.Fatalf("coinless: Al-Kharid mine reach=%q, want blocked", reach)
	}
	if payable, _ := hitField(t, mine, "payable").(interp.Bool); bool(payable) {
		t.Fatalf("coinless: Al-Kharid mine payable=true, want false")
	}
	gate := hitField(t, mine, "gate").Display()
	needs := hitField(t, mine, "needs").Display()
	youHave, _ := interp.AsInt(hitField(t, mine, "you_have"))
	if gate == "" || needs == "" {
		t.Fatalf("coinless: blocked hit lacks gate/needs (gate=%q needs=%q)", gate, needs)
	}
	if youHave != 0 {
		t.Fatalf("coinless: you_have=%d, want 0", youHave)
	}
	t.Logf("coinless Al-Kharid mine => reach=%s gate=%q needs=%q you_have=%d payable=false",
		reach, gate, needs, youHave)
}

// TestSearchMap15Coins: the SAME host with 15 coins must now see the mine as
// reach="gated" (a gate is in the way but it can pay), payable=true,
// you_have=15 — the differential feedback the operator asked for.
func TestSearchMap15Coins(t *testing.T) {
	h := newOracleHost(t, 15)

	v, err := dslSearchMap(t.Context(), h, []interp.Value{interp.String("mining-site")}, nil)
	list := callList(t, v, err)
	mine := findMineHit(t, list)
	if mine == nil {
		t.Fatalf("search_map did not include the Al-Kharid mine")
	}
	reach := hitField(t, mine, "reach").Display()
	if reach != "gated" {
		t.Fatalf("15-coin: Al-Kharid mine reach=%q, want gated", reach)
	}
	if payable, _ := hitField(t, mine, "payable").(interp.Bool); !bool(payable) {
		t.Fatalf("15-coin: Al-Kharid mine payable=false, want true")
	}
	youHave, _ := interp.AsInt(hitField(t, mine, "you_have"))
	if youHave != 15 {
		t.Fatalf("15-coin: you_have=%d, want 15", youHave)
	}
	t.Logf("15-coin Al-Kharid mine => reach=%s you_have=%d payable=true", reach, youHave)
}

// TestReachableHandler: the reachable(x,y) single-tile verdict mirrors the
// search_map entry — coinless => blocked at the mine.
func TestReachableHandler(t *testing.T) {
	h := newOracleHost(t, 0)
	v, err := dslReachable(t.Context(), h, []interp.Value{interp.Int(74), interp.Int(583)}, nil)
	if err != nil {
		t.Fatalf("reachable Go error: %v", err)
	}
	res := v.(*interp.CallResult)
	if res.Err != nil {
		t.Fatalf("reachable failed: %s %s", res.Err.Code, res.Err.Reason)
	}
	mp := res.Val.(*interp.Map)
	if got := mp.Items["reach"].Display(); got != "blocked" {
		t.Fatalf("reachable(74,583) coinless reach=%q, want blocked", got)
	}
}

// TestSurveyMapText: survey_map returns a non-empty text overview that mentions
// at least one gate state for a coinless host near Lumbridge.
func TestSurveyMapText(t *testing.T) {
	h := newOracleHost(t, 0)
	v, err := dslSurveyMap(t.Context(), h, nil, nil)
	if err != nil {
		t.Fatalf("survey_map Go error: %v", err)
	}
	s, ok := v.(interp.String)
	if !ok {
		t.Fatalf("survey_map returned %T, want String", v)
	}
	if len(string(s)) == 0 {
		t.Fatalf("survey_map returned empty text")
	}
	t.Logf("survey_map =>\n%s", string(s))
}

// TestSearchMapNoOracleDegrades: with no oracle wired, search_map fails
// gracefully (NO_SUCH_ITEM) rather than panicking — the nil-degrade contract.
func TestSearchMapNoOracleDegrades(t *testing.T) {
	h := New(Options{Username: "no-map"})
	defer h.Close()
	v, err := dslSearchMap(t.Context(), h, []interp.Value{interp.String("mining-site")}, nil)
	if err != nil {
		t.Fatalf("search_map Go error: %v", err)
	}
	res, ok := v.(*interp.CallResult)
	if !ok || res.Err == nil {
		t.Fatalf("search_map with no oracle should fail gracefully, got %v", v)
	}
}
