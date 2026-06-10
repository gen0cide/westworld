package runtime

import "testing"

// TestNeighborSectorUnderstanding proves the compass-adjacent glance the debug
// fog endpoint serves: fixed N,E,S,W order under the RSC mirror (east = the
// SMALLER-x neighbour), a visited neighbour reads terrain > 0 while an
// unvisited one reads 0, and an off-grid direction is omitted entirely.
func TestNeighborSectorUnderstanding(t *testing.T) {
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

	// Visit only the LARGER-x neighbour of the centre sector — WEST in RSC.
	h.fogObservePosition(120, 72)

	glances := h.NeighborSectorUnderstanding(72, 72)
	if len(glances) != 4 {
		t.Fatalf("centre sector: got %d glances, want 4 (N,E,S,W)", len(glances))
	}
	byDir := map[string]SectorGlance{}
	for i, g := range glances {
		byDir[g.Direction] = g
		want := [...]string{"north", "east", "south", "west"}[i]
		if g.Direction != want {
			t.Errorf("glance %d direction = %q, want fixed order %q", i, g.Direction, want)
		}
	}
	if byDir["west"].Terrain <= 0 {
		t.Errorf("west (larger x, visited) terrain = %v, want > 0", byDir["west"].Terrain)
	}
	if byDir["east"].Terrain != 0 {
		t.Errorf("east (smaller x, unvisited) terrain = %v, want 0", byDir["east"].Terrain)
	}

	// At the map corner, north (y-48 < 0) and east (x-48 < 0) fall off the grid.
	glances = h.NeighborSectorUnderstanding(24, 24)
	if len(glances) != 2 {
		t.Fatalf("corner sector: got %d glances, want 2 (S,W only)", len(glances))
	}
	if glances[0].Direction != "south" || glances[1].Direction != "west" {
		t.Errorf("corner directions = %q,%q, want south,west", glances[0].Direction, glances[1].Direction)
	}

	// In the grid's LAST sector column (largest x), the "west" neighbour
	// (x+48 → sgx ≥ fogGridStride) is off-grid too — it must be omitted, not
	// glanced as a phantom 0%/0% sector indistinguishable from an unexplored
	// one. The low edges above were always guarded; this pins the upper edge.
	edgeX := (fogGridStride-1)*fogSectorSize + 24
	glances = h.NeighborSectorUnderstanding(edgeX, 72)
	if len(glances) != 3 {
		t.Fatalf("last-column sector: got %d glances, want 3 (N,E,S only)", len(glances))
	}
	for i, want := range []string{"north", "east", "south"} {
		if glances[i].Direction != want {
			t.Errorf("last-column glance %d = %q, want %q", i, glances[i].Direction, want)
		}
	}
}
