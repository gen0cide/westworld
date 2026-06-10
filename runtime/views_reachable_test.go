package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Behavior tests for the DSL-1 default-ON reachable= filter on the
// action-intent selectors (views_reachable.go). The oracle-backed
// tests use the real Lumbridge fixture (newOracleHost, skips without
// the OpenRSC archive) and FABRICATE one reachable and one UNREACHABLE
// target — the unreachable one strictly CLOSER — so a selector that
// ignores the gate visibly returns the wrong answer.

// reachAndNegativeTiles finds, for the fixture host: a tile in a
// DIFFERENT walkable component (negative space — the gate must drop
// targets there) and a nearby tile in the host's own component (the
// gate must keep targets there). The negative tile comes straight from
// the oracle's component table (the representative tile of any other
// component), so the test doesn't depend on wall geometry luck around
// the fixture position. Skips when the oracle has a single component.
func reachAndNegativeTiles(t *testing.T, h *Host) (negX, negY, reachX, reachY int) {
	t.Helper()
	pos := h.world.Self.Position()
	hostComp, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y)
	if !ok {
		t.Skip("host tile has no resolvable component")
	}
	negX, negY = -1, -1
	for _, c := range h.worldOracle.Components() {
		if c.ID == hostComp {
			continue
		}
		// CompNear must agree (it snaps; a rep tile adjacent to a
		// bigger component could snap-resolve differently).
		if rc, _, _, ok := h.worldOracle.CompNear(c.RepX, c.RepY); ok && rc != hostComp {
			negX, negY = c.RepX, c.RepY
			break
		}
	}
	if negX < 0 {
		t.Skip("oracle has no second walkable component")
	}
	for r := 0; r <= 10; r++ {
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				if r > 0 && dx > -r && dx < r && dy > -r && dy < r {
					continue // ring border only
				}
				x, y := pos.X+dx, pos.Y+dy
				if c := h.worldOracle.CompAt(x, y); c == hostComp {
					return negX, negY, x, y
				}
			}
		}
	}
	t.Skip("no same-component tile within 10 tiles of the fixture position")
	return
}

// TestGroundItemSelectorsReachableDefault drives the decisive
// only-unreachable case first — every gated selector must answer
// Null/honestly-empty while the raw .all list and the reachable=false
// opt-out still see the item — then adds a reachable item and asserts
// the selectors pick it.
func TestGroundItemSelectorsReachableDefault(t *testing.T) {
	h := newOracleHost(t, 0)
	negX, negY, reachX, reachY := reachAndNegativeTiles(t, h)
	const itemID = 10 // coins
	h.world.GroundItems.Add(negX, negY, itemID)
	gv := &groundItemsView{host: h}
	byID := &groundItemsByIDCallable{host: h}

	// Only an unreachable item visible: gated selectors say nothing-there.
	// .nearest reads as null (falsey, == null) but must STAY CALLABLE so the
	// documented opt-out world.ground_items.nearest(reachable=false) works in
	// exactly the all-unreachable state it exists for (review DSL-1).
	nv, _ := gv.Get("nearest")
	if !interp.IsNullish(nv) || interp.Truthy(nv) || !interp.Equal(nv, interp.Null{}) {
		t.Fatalf("nearest over only-unreachable items = %T, want a null-like read", nv)
	}
	nc, ok := nv.(interp.Callable)
	if !ok {
		t.Fatalf("nearest over only-unreachable items = %T, want it still Callable (the reachable=false escape hatch)", nv)
	}
	esc, escErr := nc.Call(nil, map[string]interp.Value{"reachable": interp.Bool(false)})
	if escErr != nil {
		t.Fatalf("nearest(reachable=false): %v", escErr)
	}
	if gi, ok := esc.(*groundItemView); !ok || gi.record.X != negX || gi.record.Y != negY {
		t.Fatalf("nearest(reachable=false) = %v, want the unreachable item at (%d,%d)", esc, negX, negY)
	}
	// ... while the gated called form still answers Null.
	esc, escErr = nc.Call(nil, nil)
	if escErr != nil || !isNull(esc) {
		t.Fatalf("gated nearest() over only-unreachable items = %v (err %v), want Null", esc, escErr)
	}
	if v, _ := gv.Get("most_valuable"); !isNull(v) {
		t.Fatalf("most_valuable over only-unreachable items = %T, want Null", v)
	}
	got, err := byID.Call([]interp.Value{interp.Int(itemID)}, nil)
	if err != nil {
		t.Fatalf("by_id: %v", err)
	}
	if !isNull(got) {
		t.Fatalf("by_id default over only-unreachable items = %T, want Null", got)
	}
	// ... while the opt-out and the raw list still see it.
	got, err = byID.Call([]interp.Value{interp.Int(itemID)}, map[string]interp.Value{"reachable": interp.Bool(false)})
	if err != nil {
		t.Fatalf("by_id reachable=false: %v", err)
	}
	gi, ok := got.(*groundItemView)
	if !ok || gi.record.X != negX || gi.record.Y != negY {
		t.Fatalf("by_id reachable=false = %v, want the unreachable item at (%d,%d)", got, negX, negY)
	}
	if v, _ := gv.Get("all"); len(v.(*interp.List).Items) != 1 {
		t.Fatalf(".all must stay unfiltered, got %v", v)
	}

	// Add a reachable item: the selectors pick it, ignoring the other.
	h.world.GroundItems.Add(reachX, reachY, itemID)
	v, _ := gv.Get("nearest")
	near, ok := v.(*groundItemsNearestValue)
	if !ok {
		t.Fatalf("nearest with a reachable item = %T, want the dual-mode wrapper", v)
	}
	if near.base.record.X != reachX || near.base.record.Y != reachY {
		t.Fatalf("nearest picked (%d,%d), want the REACHABLE item at (%d,%d)",
			near.base.record.X, near.base.record.Y, reachX, reachY)
	}
	got, err = byID.Call([]interp.Value{interp.Int(itemID)}, nil)
	if err != nil {
		t.Fatalf("by_id: %v", err)
	}
	gi, ok = got.(*groundItemView)
	if !ok || gi.record.X != reachX || gi.record.Y != reachY {
		t.Fatalf("by_id default = %v, want the reachable item at (%d,%d)", got, reachX, reachY)
	}
}

func isNull(v interp.Value) bool {
	_, ok := v.(interp.Null)
	return ok
}

// TestNearestNpcReachableDefault: with only a negative-space NPC in
// view, nearest_npc answers Null; reachable=false restores
// omniscience; a reachable NPC is then preferred.
func TestNearestNpcReachableDefault(t *testing.T) {
	h := newOracleHost(t, 0)
	negX, negY, reachX, reachY := reachAndNegativeTiles(t, h)
	h.world.Npcs.Set(world.NpcRecord{Index: 1, TypeID: 1, X: negX, Y: negY})

	v, err := dslNearestNpc(context.Background(), h, nil, nil)
	if err != nil {
		t.Fatalf("nearest_npc: %v", err)
	}
	if !isNull(v) {
		t.Fatalf("nearest_npc over only-unreachable = %T, want Null", v)
	}

	v, err = dslNearestNpc(context.Background(), h, nil, map[string]interp.Value{"reachable": interp.Bool(false)})
	if err != nil {
		t.Fatalf("nearest_npc reachable=false: %v", err)
	}
	nv, ok := v.(*npcView)
	if !ok || nv.record.Index != 1 {
		t.Fatalf("nearest_npc reachable=false = %v, want the unreachable npc 1", v)
	}

	h.world.Npcs.Set(world.NpcRecord{Index: 2, TypeID: 1, X: reachX, Y: reachY})
	v, err = dslNearestNpc(context.Background(), h, nil, nil)
	if err != nil {
		t.Fatalf("nearest_npc: %v", err)
	}
	nv, ok = v.(*npcView)
	if !ok || nv.record.Index != 2 {
		t.Fatalf("nearest_npc = %v, want the REACHABLE npc 2", v)
	}
}

// truthyPred is a minimal always-true predicate Callable for the
// positional-arg tests (nearest_npc(pred, false)).
type truthyPred struct{}

func (truthyPred) Kind() string    { return "lambda" }
func (truthyPred) Display() string { return "<pred>" }
func (truthyPred) Call([]interp.Value, map[string]interp.Value) (interp.Value, error) {
	return interp.Bool(true), nil
}

// TestPositionalReachableFlag: the spec arity admits the reachable flag as a
// trailing POSITIONAL arg — nearest_npc(pred, false), scan_for(type, radius,
// false) — so both must honor it like the named opt-out instead of silently
// ignoring it (scan_for) or rejecting it at runtime (nearest_npc).
func TestPositionalReachableFlag(t *testing.T) {
	h := newOracleHost(t, 0)
	negX, negY, _, _ := reachAndNegativeTiles(t, h)
	h.world.Npcs.Set(world.NpcRecord{Index: 1, TypeID: 1, X: negX, Y: negY})

	v, err := dslNearestNpc(context.Background(), h, []interp.Value{truthyPred{}, interp.Bool(false)}, nil)
	if err != nil {
		t.Fatalf("nearest_npc(pred, false): %v", err)
	}
	if nv, ok := v.(*npcView); !ok || nv.record.Index != 1 {
		t.Fatalf("nearest_npc(pred, false) = %v, want the unreachable npc 1", v)
	}
	// A truthy positional flag keeps the gate on.
	v, err = dslNearestNpc(context.Background(), h, []interp.Value{truthyPred{}, interp.Bool(true)}, nil)
	if err != nil {
		t.Fatalf("nearest_npc(pred, true): %v", err)
	}
	if !isNull(v) {
		t.Fatalf("nearest_npc(pred, true) over only-unreachable = %T, want Null", v)
	}

	// scan_for: a live scenery record on the negative-space tile appears ONLY
	// under the opt-out — positional and named must agree.
	const fireID = 97
	def := h.facts.SceneryDef(fireID)
	if def == nil || def.Name == "" {
		t.Skip("facts has no name for scenery def 97")
	}
	h.world.Scenery.Add(negX, negY, fireID)
	pos := h.world.Self.Position()
	radius := interp.Int(int64(chebyshev(pos.X, pos.Y, negX, negY) + 1))
	hasNegTile := func(v interp.Value) bool {
		list, ok := v.(*interp.List)
		if !ok {
			return false
		}
		for _, item := range list.Items {
			if pv, ok := item.(*placementView); ok && pv.p.X == negX && pv.p.Y == negY {
				return true
			}
		}
		return false
	}
	typ := interp.String(def.Name)
	if d, err := dslScanFor(context.Background(), h, []interp.Value{typ, radius}, nil); err != nil || hasNegTile(d) {
		t.Fatalf("default scan_for must drop the negative-space record (err %v)", err)
	}
	if p, err := dslScanFor(context.Background(), h, []interp.Value{typ, radius, interp.Bool(false)}, nil); err != nil || !hasNegTile(p) {
		t.Fatalf("scan_for(type, radius, false) must include the negative-space record (err %v)", err)
	}
}

// TestScenerySelectorsReachableDefault: world.scenery.nearest / .by_id
// drop a negative-space dynamic-scenery record; by_id's
// reachable=false opt-out and a reachable record restore answers.
func TestScenerySelectorsReachableDefault(t *testing.T) {
	h := newOracleHost(t, 0)
	negX, negY, reachX, reachY := reachAndNegativeTiles(t, h)
	const fireID = 97
	h.world.Scenery.Add(negX, negY, fireID)
	sv := &sceneryView{host: h}
	byID := &sceneryByIDCallable{host: h}

	if v, _ := sv.Get("nearest"); !isNull(v) {
		t.Fatalf("scenery.nearest over only-unreachable = %T, want Null", v)
	}
	got, err := byID.Call([]interp.Value{interp.Int(fireID)}, nil)
	if err != nil {
		t.Fatalf("scenery.by_id: %v", err)
	}
	if !isNull(got) {
		t.Fatalf("scenery.by_id default over only-unreachable = %T, want Null", got)
	}
	got, err = byID.Call([]interp.Value{interp.Int(fireID)}, map[string]interp.Value{"reachable": interp.Bool(false)})
	if err != nil {
		t.Fatalf("scenery.by_id reachable=false: %v", err)
	}
	pv, ok := got.(*placementView)
	if !ok || pv.p.X != negX || pv.p.Y != negY {
		t.Fatalf("scenery.by_id reachable=false = %v, want the unreachable record at (%d,%d)", got, negX, negY)
	}
	if v, _ := sv.Get("all"); len(v.(*interp.List).Items) != 1 {
		t.Fatalf(".all must stay unfiltered, got %v", v)
	}

	h.world.Scenery.Add(reachX, reachY, fireID)
	v, _ := sv.Get("nearest")
	pv, ok = v.(*placementView)
	if !ok || pv.p.X != reachX || pv.p.Y != reachY {
		t.Fatalf("scenery.nearest = %v, want the reachable record at (%d,%d)", v, reachX, reachY)
	}
}

// TestBoundariesNearReachableDefault: every door the default returns is
// in the host's component, and the reachable=false opt-out returns a
// superset.
func TestBoundariesNearReachableDefault(t *testing.T) {
	h := newOracleHost(t, 0)
	pos := h.world.Self.Position()
	hostComp, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y)
	if !ok {
		t.Skip("host tile has no resolvable component")
	}
	near := &boundaryNearCallable{host: h}
	def, err := near.Call([]interp.Value{interp.Int(16)}, nil)
	if err != nil {
		t.Fatalf("boundaries.near: %v", err)
	}
	all, err := near.Call([]interp.Value{interp.Int(16)}, map[string]interp.Value{"reachable": interp.Bool(false)})
	if err != nil {
		t.Fatalf("boundaries.near reachable=false: %v", err)
	}
	defList, allList := def.(*interp.List), all.(*interp.List)
	if len(defList.Items) > len(allList.Items) {
		t.Fatalf("default near (%d hits) larger than reachable=false (%d) — opt-out must be a superset",
			len(defList.Items), len(allList.Items))
	}
	for _, item := range defList.Items {
		bv := item.(*boundaryView)
		if c, _, _, ok := h.worldOracle.CompNear(bv.x, bv.y); !ok || c != hostComp {
			t.Fatalf("boundaries.near returned door at (%d,%d) comp=%d outside host comp %d", bv.x, bv.y, c, hostComp)
		}
	}
}

// TestSelectorsUnfilteredWithoutOracle: a host with no world oracle
// must not filter at all (reachGate nil) — selectors degrade to the
// old omniscient behavior instead of returning Null everywhere.
func TestSelectorsUnfilteredWithoutOracle(t *testing.T) {
	h := New(Options{})
	defer h.Close()
	if h.reachGate() != nil {
		t.Fatal("reachGate must be nil without a world oracle")
	}
	h.world.GroundItems.Add(5, 5, 10)
	gv := &groundItemsView{host: h}
	v, _ := gv.Get("nearest")
	if _, ok := v.(*groundItemsNearestValue); !ok {
		t.Fatalf("oracle-less nearest = %T, want the item wrapper (no filtering)", v)
	}
	if !h.tileReachable(5, 5) {
		t.Fatal("tileReachable must report true without an oracle")
	}
}
