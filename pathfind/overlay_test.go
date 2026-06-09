package pathfind

import (
	"testing"

	"github.com/gen0cide/westworld/facts"
)

// The live overlay (backlog #1, "combinedTile") is what makes the BFS see the
// server's dynamic world: open doors passable, closed doors solid, depleted
// scenery walkable, spawned blockers solid. These tests exercise
// applyLiveOverlay directly over a hand-built grid + facts (no archive needed),
// asserting the exact Mask bits — the door bits are the deadlock-critical path.

func TestApplyLiveOverlay_OpenDoorClearsEdge(t *testing.T) {
	f := &facts.Facts{BoundaryDefs: map[int]*facts.BoundaryDef{
		1: {ID: 1, DoorType: 1, Unknown: 1}, // openable door (blocks when closed)
	}}
	g := &Grid{} // BaseX/BaseY = 0 -> local == world
	// Static baseline placed a CLOSED door on the north edge of (10,10).
	g.Mask[10][10] |= WallNorth
	g.Mask[10][9] |= WallSouth
	// The server now reports it OPEN (removed override).
	applyLiveOverlay(g, f, &LiveState{Boundaries: []LiveBoundary{
		{X: 10, Y: 10, Dir: 0, ID: -1, Removed: true},
	}})
	if g.Mask[10][10]&WallNorth != 0 {
		t.Fatal("open door must clear WallNorth so the BFS can cross")
	}
	if g.Mask[10][9]&WallSouth != 0 {
		t.Fatal("open door must clear the mirrored WallSouth")
	}
}

func TestApplyLiveOverlay_ClosedDoorBlocks_OpenDefClears(t *testing.T) {
	f := &facts.Facts{BoundaryDefs: map[int]*facts.BoundaryDef{
		1:  {ID: 1, DoorType: 1, Unknown: 0},  // fixed wall/fence: BlocksMovement
		2:  {ID: 2, DoorType: 1, Unknown: 1},  // CLOSED openable door: BlocksWhenClosed
		11: {ID: 11, DoorType: 0, Unknown: 1}, // open doorframe: passable
	}}
	g := &Grid{}
	applyLiveOverlay(g, f, &LiveState{Boundaries: []LiveBoundary{
		{X: 10, Y: 10, Dir: 1, ID: 1}, // fixed wall on the west edge
		{X: 20, Y: 20, Dir: 1, ID: 2}, // CLOSED door on the west edge
	}})
	if g.Mask[10][10]&WallEast == 0 || g.Mask[9][10]&WallWest == 0 {
		t.Fatal("fixed wall must set WallEast + mirrored WallWest")
	}
	// The server's rule (registerObject: doorType==1 blocks): a CLOSED
	// openable door IS a wall until opened — the BFS routes TO it and the
	// traversal flow opens it. (Safe only because the flow lands with this.)
	if g.Mask[20][20]&WallEast == 0 || g.Mask[19][20]&WallWest == 0 {
		t.Fatal("a CLOSED openable door must block the edge (server doorType==1 rule)")
	}

	// Door OPEN on the authentic wire = ID-OVERWRITE to the open doorframe
	// def (11, DoorType=0) at the same edge — never a removal. The swap must
	// clear the edge so the replanned BFS routes through.
	applyLiveOverlay(g, f, &LiveState{Boundaries: []LiveBoundary{
		{X: 20, Y: 20, Dir: 1, ID: 11},
	}})
	if g.Mask[20][20]&WallEast != 0 || g.Mask[19][20]&WallWest != 0 {
		t.Fatal("the open-def ID swap must clear the door's edge")
	}
}

// TestApplyLiveOverlay_SceneryGateSwap guards the def-swap path that the
// cabbage-field pen exposed: a closed gate (def 60, type 2, 1x2) opened on
// the server arrives as an opcode-48 ID swap to the OPEN def (59, type 3,
// non-blocking). The wire carries no direction, so the overlay re-derives
// the footprint from the STATIC loc and must clear BOTH footprint tiles'
// edges.
func TestApplyLiveOverlay_SceneryGateSwap(t *testing.T) {
	f := &facts.Facts{
		SceneryDefs: map[int]*facts.SceneryDef{
			60: {ID: 60, Name: "gate", Command1: "open", Command2: "Examine", Type: 2, Width: 1, Height: 2},
			59: {ID: 59, Name: "gate", Command1: "WalkTo", Command2: "close", Type: 3, Width: 1, Height: 2},
		},
		BoundaryDefs: map[int]*facts.BoundaryDef{},
		SceneryLocs:  []facts.SceneryLoc{{DefID: 60, X: 10, Y: 10, Direction: 2}},
	}
	f.BuildIndex()

	g := &Grid{}
	applyScenery(g, f) // static layer: closed gate, dir 2 → 2x1 swapped footprint
	for _, tile := range [][2]int{{10, 10}, {11, 10}} {
		if g.Mask[tile[0]][tile[1]]&WallSouth == 0 {
			t.Fatalf("static closed gate must block (%d,%d) south edge", tile[0], tile[1])
		}
	}

	// The host opens it: the server swaps the loc to the OPEN def (59).
	applyLiveOverlay(g, f, &LiveState{Scenery: []LiveScenery{
		{X: 10, Y: 10, ID: 59},
	}})
	for x := 9; x <= 12; x++ {
		for y := 9; y <= 12; y++ {
			if g.Mask[x][y] != 0 {
				t.Fatalf("opened gate must clear its whole footprint; residue at (%d,%d)=%d", x, y, g.Mask[x][y])
			}
		}
	}

	// And the re-close 3s later (swap back to 60) restores the blocks.
	applyLiveOverlay(g, f, &LiveState{Scenery: []LiveScenery{
		{X: 10, Y: 10, ID: 60},
	}})
	for _, tile := range [][2]int{{10, 10}, {11, 10}} {
		if g.Mask[tile[0]][tile[1]]&WallSouth == 0 {
			t.Fatalf("re-closed gate must re-block (%d,%d)", tile[0], tile[1])
		}
	}
}

func TestApplyLiveOverlay_DepletedSceneryClears(t *testing.T) {
	f := &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{}}
	g := &Grid{}
	g.Mask[20][20] |= FullBlockC | Object // static baseline: a solid tree/rock
	// Server cleared it (mined-out rock / burned fire).
	applyLiveOverlay(g, f, &LiveState{Scenery: []LiveScenery{
		{X: 20, Y: 20, Removed: true},
	}})
	if g.Mask[20][20]&(FullBlockC|Object) != 0 {
		t.Fatal("depleted scenery must clear its footprint block")
	}
}

func TestApplyLiveOverlay_LiveSolidSceneryBlocks(t *testing.T) {
	f := &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
		97: {ID: 97, Type: 1, Width: 1, Height: 1}, // a lit fire (solid)
	}}
	g := &Grid{}
	applyLiveOverlay(g, f, &LiveState{Scenery: []LiveScenery{
		{X: 20, Y: 20, ID: 97},
	}})
	if g.Mask[20][20]&FullBlockC == 0 {
		t.Fatal("a runtime-spawned solid object must block its tile")
	}
}

func TestApplyLiveOverlay_NilIsNoop(t *testing.T) {
	g := &Grid{}
	g.Mask[5][5] = WallNorth
	applyLiveOverlay(g, nil, nil)
	if g.Mask[5][5] != WallNorth {
		t.Fatal("nil overlay must leave the static grid untouched")
	}
}

// TestStampSceneryFootprint_Type2MultiTile guards the phantom-gap root cause:
// a multi-tile directional (type 2) object — e.g. a 2-wide closed gate — must
// stamp its edge pair on EVERY footprint tile, not just the loc tile. The old
// code stamped one tile, leaving the gate's second half unmodeled free space
// the BFS planned through (the cabbage-field pen).
func TestStampSceneryFootprint_Type2MultiTile(t *testing.T) {
	def := &facts.SceneryDef{ID: 60, Type: 2, Width: 2, Height: 1}
	g := &Grid{}
	stampSceneryFootprint(g, 10, 10, 0, def, true) // dir 0: east edge per tile

	for _, tile := range [][2]int{{10, 10}, {11, 10}} {
		if g.Mask[tile[0]][tile[1]]&WallEast == 0 {
			t.Fatalf("footprint tile (%d,%d) missing WallEast — multi-tile loop broken", tile[0], tile[1])
		}
	}
	if g.Mask[9][10]&WallWest == 0 || g.Mask[10][10]&WallWest == 0 {
		t.Fatal("mirrored WallWest missing on a neighbor")
	}

	// Clearing must remove exactly what stamping set (the overlay's
	// def-swap path relies on symmetric set/clear).
	stampSceneryFootprint(g, 10, 10, 0, def, false)
	for x := 8; x <= 12; x++ {
		if g.Mask[x][10] != 0 {
			t.Fatalf("clear left residue at (%d,10): %d", x, g.Mask[x][10])
		}
	}
}

// TestStampSceneryFootprint_DirSwapAndType1 checks the 90°-rotated dimension
// swap and the type-1 full-block rule.
func TestStampSceneryFootprint_DirSwapAndType1(t *testing.T) {
	def := &facts.SceneryDef{ID: 5, Type: 1, Width: 3, Height: 1}
	g := &Grid{}
	stampSceneryFootprint(g, 20, 20, 2, def, true) // dir 2: w/h swap -> 1x3
	for dy := 0; dy < 3; dy++ {
		if g.Mask[20][20+dy]&(FullBlockC|Object) == 0 {
			t.Fatalf("rotated footprint tile (20,%d) not blocked", 20+dy)
		}
	}
	if g.Mask[21][20] != 0 {
		t.Fatal("unrotated-width tile must NOT be blocked after the dir swap")
	}
}

// TestStampSceneryFootprint_SkipID1147 mirrors server World.java:501-503 /
// Plutonium world.go:345-347: def 1147 never contributes collision.
func TestStampSceneryFootprint_SkipID1147(t *testing.T) {
	def := &facts.SceneryDef{ID: 1147, Type: 1, Width: 1, Height: 1}
	g := &Grid{}
	stampSceneryFootprint(g, 10, 10, 0, def, true)
	if g.Mask[10][10] != 0 {
		t.Fatal("def 1147 must be excluded from collision (server/Plutonium parity)")
	}
}

// TestOverlayBlocks_TileDefs checks the authoritative terrain rule: overlay is
// a 1-based TileDef index, 250 remaps to 2, ObjectType!=0 blocks — and the
// legacy water/lava hardcode survives only as the no-table fallback.
func TestOverlayBlocks_TileDefs(t *testing.T) {
	f := &facts.Facts{TileDefs: []facts.TileDef{
		{ID: 0, ObjectType: 0}, // overlay 1: walkable (path)
		{ID: 1, ObjectType: 1}, // overlay 2: water — blocks
		{ID: 2, ObjectType: 0}, // overlay 3
	}}
	if f.OverlayBlocks(1) {
		t.Fatal("overlay 1 must be walkable")
	}
	if !f.OverlayBlocks(2) {
		t.Fatal("overlay 2 (water) must block")
	}
	if !f.OverlayBlocks(250) {
		t.Fatal("overlay 250 must remap to 2 and block")
	}
	if f.OverlayBlocks(0) || f.OverlayBlocks(99) {
		t.Fatal("overlay 0 / out-of-table must not block")
	}
	empty := &facts.Facts{}
	if !empty.OverlayBlocks(2) || !empty.OverlayBlocks(11) || empty.OverlayBlocks(3) {
		t.Fatal("no-table fallback must keep the legacy 2/11 behavior")
	}
}
