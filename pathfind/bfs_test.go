package pathfind

import (
	"testing"

	"github.com/gen0cide/westworld/facts"
)

const factsRoot = "/Users/flint/Code/openrsc"

// loadTestWorld loads facts + landscape for the Lumbridge area. Skips
// the test if the OpenRSC source tree isn't checked out next to the
// bot's repo.
func loadTestWorld(t *testing.T) (*facts.Facts, *Landscape) {
	t.Helper()
	f, err := facts.Load(facts.DefaultSources(factsRoot))
	if err != nil {
		t.Skipf("facts not available: %v", err)
	}
	l, err := OpenLandscape(testLandscape)
	if err != nil {
		t.Skipf("landscape not available: %v", err)
	}
	return f, l
}

func TestFindPathStraightOpen(t *testing.T) {
	f, l := loadTestWorld(t)
	defer l.Close()

	// Delores's spawn area at (134, 644). Walk 5 tiles east in open
	// ground (no walls).
	g, err := BuildGrid(l, f, 134, 644, 0, nil)
	if err != nil {
		t.Fatalf("BuildGrid: %v", err)
	}
	corners := g.FindPathToTile(134, 644, 139, 644, false)
	if len(corners) == 0 {
		t.Fatal("no path found for 5-tile straight east walk in open ground")
	}
	t.Logf("straight east 5 tiles: %d corners → %v", len(corners), corners)
	// Last corner should be the goal.
	if corners[len(corners)-1] != (Corner{X: 139, Y: 644}) {
		t.Errorf("last corner = %v, want goal (139, 644)", corners[len(corners)-1])
	}
}

func TestFindPathThroughDoorway(t *testing.T) {
	f, l := loadTestWorld(t)
	defer l.Close()

	// From south of the Lumbridge general store door to a tile
	// adjacent to where the Shopkeeper stands behind the counter.
	// Doorframe is at (132, 641) dir=2; player at (131, 642) is
	// directly south of it.
	g, err := BuildGrid(l, f, 131, 642, 0, nil)
	if err != nil {
		t.Fatalf("BuildGrid: %v", err)
	}
	corners := g.FindPathToTile(131, 642, 134, 641, false)
	if len(corners) == 0 {
		t.Fatal("no path found through Lumbridge shop doorway")
	}
	t.Logf("through doorway to (134, 641): %d corners → %v", len(corners), corners)
}

func TestFindPathToHans(t *testing.T) {
	f, l := loadTestWorld(t)
	defer l.Close()

	// Delores spawn → Hans's spawn area inside the Lumbridge castle
	// courtyard. The courtyard entrance is gated by the CLOSED double
	// doors at (128,658) — scenery def 64 "doors", type 2, 1x2 footprint,
	// command "Open". With correct footprint collision the STATIC grid
	// must block the route (the server does: a closed door is a wall
	// until opened). The old version of this test passed only because
	// the doors' second footprint tile (128,659) was unmodeled — the
	// phantom-gap bug — and the "path" walked through the doors.
	g, err := BuildGrid(l, f, 134, 644, 0, nil)
	if err != nil {
		t.Fatalf("BuildGrid: %v", err)
	}
	if corners := g.FindPathToTile(134, 644, 130, 655, true); len(corners) != 0 {
		t.Fatalf("static path to Hans must be BLOCKED by the closed castle doors, got %v", corners)
	}

	// Sanity: open ground OUTSIDE the courtyard stays reachable (we
	// blocked the door edges, not the surrounding area).
	if corners := g.FindPathToTile(134, 644, 126, 651, false); len(corners) == 0 {
		t.Fatal("open ground outside the castle courtyard must remain reachable")
	}
}
