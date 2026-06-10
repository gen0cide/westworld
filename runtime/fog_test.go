package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	"github.com/gen0cide/westworld/world"
)

// fakeFogOracle is a tiny in-memory component grid following the oracle's
// band-encoded convention (plane = y/944): only tiles present in comps are
// standable. Pointer type so it is a comparable fogIndex cache key.
type fakeFogOracle struct {
	dim, dimY int
	comps     map[[2]int]int32
}

func (f *fakeFogOracle) Dim() int  { return f.dim }
func (f *fakeFogOracle) DimY() int { return f.dimY }
func (f *fakeFogOracle) CompAt(x, y int) int32 {
	if x < 0 || y < 0 || x >= f.dim || y >= f.dimY {
		return -1
	}
	if c, ok := f.comps[[2]int{x, y}]; ok {
		return c
	}
	return -1
}

// TestFogSectorRecordingPersistsRoundTrip proves the visited-sector set
// survives a restart through the memory layer (same spine as the knowledge
// ledger): record, flush, then a fresh host sharing the durable store reloads
// the identical set — and the persisted format is the compact sorted-id JSON.
func TestFogSectorRecordingPersistsRoundTrip(t *testing.T) {
	store := hostkv.NewMemory()
	ctx := context.Background()

	h := newTestHost()
	h.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	h.fogObservePosition(120, 504)  // plane 0, sector (2,10)
	h.fogObservePosition(121, 505)  // same sector — must not add an entry
	h.fogObservePosition(120, 1448) // plane 1, same footprint — a DISTINCT sector
	if n := len(h.fog.visited); n != 2 {
		t.Fatalf("expected 2 visited sectors, got %d", n)
	}
	h.flushFog(ctx)

	// Compact persisted format: sorted JSON ids under the stable key.
	raw, ok := store.GetRaw(fogSectorsKey)
	if !ok {
		t.Fatalf("flush did not write %q to the local store", fogSectorsKey)
	}
	want := `[212,653]` // (0*21+10)*21+2 and (1*21+10)*21+2
	if string(raw) != want {
		t.Fatalf("persisted set = %s, want %s", raw, want)
	}

	h2 := newTestHost()
	h2.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	if len(h2.fog.visited) != 0 {
		t.Fatal("precondition: a fresh host should have no visited sectors before load")
	}
	h2.loadFog(ctx)
	if len(h2.fog.visited) != 2 ||
		!h2.fog.visited[fogSectorAt(120, 504)] || !h2.fog.visited[fogSectorAt(120, 1448)] {
		t.Fatalf("after reload visited = %v, want the two recorded sectors", h2.fog.visited)
	}
	// Restored sectors are NOT pre-harvested: the new session re-harvests once.
	if len(h2.fog.harvested) != 0 {
		t.Fatal("restore must not mark sectors harvested (harvest is per-session)")
	}
}

// TestFogCoverageAgainstFakeOracle proves the coverage math: the denominator
// counts only sectors containing standable tiles (ocean/void never count) and
// the numerator only the visited ones among them.
func TestFogCoverageAgainstFakeOracle(t *testing.T) {
	h := newTestHost()
	comps := map[[2]int]int32{
		{60, 10}: 0, // sector (1,0) — walkable via a single tile
		// sectors (0,1)/(1,1) hold no standable tile — void, never counted
	}
	// Sector (0,0) is walkable THROUGHOUT (one standable tile per 8x8
	// sub-cell) so partial sight registers as partial understanding.
	for cx := 0; cx < 48; cx += 8 {
		for cy := 0; cy < 48; cy += 8 {
			comps[[2]int{cx + 4, cy + 4}] = 0
		}
	}
	fake := &fakeFogOracle{dim: 96, dimY: 96, comps: comps}
	h.fog.oracle = fake

	frac, seen, total := h.Coverage()
	if total != 2 || seen != 0 || frac != 0 {
		t.Fatalf("unvisited coverage = (%v, %d, %d), want (0, 0, 2)", frac, seen, total)
	}

	h.fogObservePosition(10, 10) // visit the first walkable sector
	frac, seen, total = h.Coverage()
	// Weighted semantics (sub-cell upgrade): one standing position lights only
	// the sub-cells inside its view radius, so a visited sector contributes
	// strictly less than a full sector-count.
	if total != 2 || seen != 1 || frac <= 0 || frac >= 0.5 {
		t.Fatalf("coverage after one visit = (%v, %d, %d), want seen=1 and 0 < frac < 0.5", frac, seen, total)
	}
	prev := frac

	h.fogObservePosition(10, 60) // a VOID sector: recorded, but never counts
	frac, seen, total = h.Coverage()
	if total != 2 || seen != 1 || frac != prev {
		t.Fatalf("void-sector visit changed coverage: (%v, %d, %d), want unchanged (%v, 1, 2)", frac, seen, total, prev)
	}

	// Walking the whole sector lights every walkable sub-cell — full weight.
	for wx := 0; wx < 48; wx += 6 {
		for wy := 0; wy < 48; wy += 6 {
			h.fogObservePosition(wx, wy)
		}
	}
	frac, _, _ = h.Coverage()
	if frac <= prev || frac != 0.5 {
		t.Fatalf("thorough walk should fully explore sector (0,0) = half the world: got %v (prev %v), want 0.5", frac, prev)
	}
}

// TestFogFirstEntryHarvestWritesLedgerOnce proves the perception harvest: the
// first entry into a sector writes ProvObserved location beliefs for the
// load-bearing POIs (named subject + a claim containing it), dedups repeated
// instances of the same subject, skips non-notable scenery, harvests visible
// ground items — and a re-entry into the same sector writes NOTHING again.
func TestFogFirstEntryHarvestWritesLedgerOnce(t *testing.T) {
	h := newTestHost()
	f := &facts.Facts{
		SceneryDefs: map[int]*facts.SceneryDef{
			118: {ID: 118, Name: "Furnace", Command1: "WalkTo"},
			0:   {ID: 0, Name: "Tree", Command1: "Chop"},
			100: {ID: 100, Name: "Rock", Command1: "Mine", Model: "copperrock1"},
			5:   {ID: 5, Name: "signpost", Command1: "WalkTo"}, // NOT load-bearing
		},
		SceneryLocs: []facts.SceneryLoc{
			{DefID: 118, X: 100, Y: 100},
			{DefID: 0, X: 101, Y: 101},
			{DefID: 0, X: 102, Y: 102}, // second tree: same subject, deduped per sector
			{DefID: 100, X: 103, Y: 103},
			{DefID: 5, X: 104, Y: 104},
		},
		ItemDefs: map[int]*facts.ItemDef{542: {ID: 542, Name: "Cape"}},
	}
	f.BuildIndex()
	h.facts = f
	h.world.GroundItems.Add(105, 105, 542) // a spawn the host SEES in the mirror

	h.fogObservePosition(100, 100) // first entry into sector (2,2)

	furnace := h.knowledge.Get("furnace")
	if furnace.Kind != "location" || len(furnace.Beliefs) != 1 {
		t.Fatalf("furnace = %+v, want one location belief", furnace)
	}
	if b := furnace.Beliefs[0]; b.Provenance != knowledge.ProvObserved ||
		b.Claim != "furnace is at (100,100)" || b.Confidence() < 0.85 {
		t.Fatalf("furnace belief = %+v, want observed 'furnace is at (100,100)' at high confidence", b)
	}
	if tree := h.knowledge.Get("tree"); len(tree.Beliefs) != 1 || tree.Beliefs[0].Claim != "tree is at (101,101)" {
		t.Fatalf("tree = %+v, want ONE deduped belief at the first instance", tree)
	}
	if rock := h.knowledge.Get("copper rock"); len(rock.Beliefs) != 1 || rock.Beliefs[0].Claim != "copper rock is at (103,103)" {
		t.Fatalf("copper rock = %+v, want the ore-named subject from the def model", rock)
	}
	if h.knowledge.Known("signpost") {
		t.Fatal("non-load-bearing scenery must not be harvested")
	}
	cape := h.knowledge.Get("cape")
	if cape.Kind != "item" || len(cape.Beliefs) != 1 ||
		cape.Beliefs[0].Claim != "cape found on the ground at (105,105)" ||
		cape.Beliefs[0].Provenance != knowledge.ProvObserved {
		t.Fatalf("cape = %+v, want one observed ground-spawn belief", cape)
	}
	if got := h.KnownPOICount(); got != 3 { // furnace, tree, copper rock (cape is kind=item)
		t.Fatalf("KnownPOICount = %d, want 3", got)
	}

	// Re-entering the SAME sector must be a no-op (once per sector per session).
	alpha := furnace.Beliefs[0].Alpha
	h.fogObservePosition(110, 110)
	again := h.knowledge.Get("furnace")
	if len(again.Beliefs) != 1 || again.Beliefs[0].Alpha != alpha || again.Familiar != 1 {
		t.Fatalf("re-entry re-harvested: %+v (alpha was %v)", again, alpha)
	}
}

// TestFogFrontierDirections proves the per-quadrant frontier math on a 3x3
// sector grid: with the centre + corners visited, each compass quadrant
// reports the adjacent unexplored sector's representative tile and distance
// (RSC compass: north = smaller y, east = smaller x), and FrontierTarget
// picks the deterministic nearest with a stable label.
func TestFogFrontierDirections(t *testing.T) {
	h := newTestHost()
	fake := &fakeFogOracle{dim: 144, dimY: 144, comps: map[[2]int]int32{}}
	// One standable tile at the centre of each of the 9 sectors, all comp 0.
	for _, c := range [][2]int{
		{24, 24}, {72, 24}, {120, 24},
		{24, 72}, {72, 72}, {120, 72},
		{24, 120}, {72, 120}, {120, 120},
	} {
		fake.comps[c] = 0
	}
	h.fog.oracle = fake
	h.world.Self.SetPosition(world.Coord{X: 72, Y: 72})

	// Visit the centre and the four corners — the four edge sectors remain.
	for _, c := range [][2]int{{72, 72}, {24, 24}, {120, 24}, {24, 120}, {120, 120}} {
		h.fogObservePosition(c[0], c[1])
	}

	fr := h.FrontierDirections(72, 72)
	if len(fr) != 4 {
		t.Fatalf("FrontierDirections = %+v, want all four quadrants", fr)
	}
	want := map[string][2]int{
		"north": {72, 24}, // smaller y
		"east":  {24, 72}, // smaller x (x increases west)
		"south": {72, 120},
		"west":  {120, 72},
	}
	for _, f := range fr {
		w, ok := want[f.Direction]
		if !ok {
			t.Fatalf("unexpected direction %q", f.Direction)
		}
		if f.X != w[0] || f.Y != w[1] || f.Dist != 48 {
			t.Fatalf("%s frontier = (%d,%d) dist %d, want (%d,%d) dist 48", f.Direction, f.X, f.Y, f.Dist, w[0], w[1])
		}
	}

	// FrontierTarget: all four candidates tie at dist 48 — lowest sector id
	// wins (the north sector (1,0)) with a stable label.
	x, y, label, ok := h.FrontierTarget("where to buy a pickaxe")
	if !ok || x != 72 || y != 24 || label != "unexplored area near (72,24)" {
		t.Fatalf("FrontierTarget = (%d,%d,%q,%v), want (72,24,'unexplored area near (72,24)',true)", x, y, label, ok)
	}

	// The forage fallback honours the per-question source-tried tag.
	q := h.goalGraph.Upsert("where-to-buy:pickaxe", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	src, ok := h.frontierForageSource(q)
	if !ok || src.label != label || src.x != 72 || src.y != 24 {
		t.Fatalf("frontierForageSource = (%+v, %v), want the frontier target", src, ok)
	}
	h.goalGraph.Tag(q.ID, "source-tried:place:"+label)
	if _, ok := h.frontierForageSource(q); ok {
		t.Fatal("a source-tried frontier must not be re-offered for the same question")
	}

	// Exploring everything empties the frontier.
	for _, c := range [][2]int{{72, 24}, {24, 72}, {120, 72}, {72, 120}} {
		h.fogObservePosition(c[0], c[1])
	}
	if fr := h.FrontierDirections(72, 72); len(fr) != 0 {
		t.Fatalf("fully-explored world still reports a frontier: %+v", fr)
	}
	if _, _, _, ok := h.FrontierTarget("anything"); ok {
		t.Fatal("fully-explored world must report no frontier target")
	}

	// Coverage on the fake: 9 walkable, 9 visited.
	if frac, seen, total := h.Coverage(); frac != 1 || seen != 9 || total != 9 {
		t.Fatalf("coverage = (%v, %d, %d), want (1, 9, 9)", frac, seen, total)
	}
}

// TestFogHarvestIsSightNotTouch pins the operator's core intention: entering
// a sector's corner must NOT teach the host about POIs on the far side. A
// furnace 40 tiles away (outside the view radius) stays unknown until the
// host actually walks near it.
func TestFogHarvestIsSightNotTouch(t *testing.T) {
	h := newTestHost()
	f := &facts.Facts{
		SceneryDefs: map[int]*facts.SceneryDef{
			118: {ID: 118, Name: "Furnace", Command1: "WalkTo"},
		},
		SceneryLocs: []facts.SceneryLoc{
			{DefID: 118, X: 44, Y: 44}, // far corner of sector (0,0)
		},
	}
	f.BuildIndex()
	h.facts = f

	h.fogObservePosition(2, 2) // near corner: furnace is ~42 tiles away
	if h.knowledge.Known("furnace") {
		t.Fatal("touching the sector corner must not teach the far-side furnace")
	}
	h.fogObservePosition(40, 40) // now the furnace is in view
	if !h.knowledge.Known("furnace") {
		t.Fatal("walking within view of the furnace must harvest it")
	}
}
