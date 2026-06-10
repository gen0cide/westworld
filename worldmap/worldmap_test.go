package worldmap

import (
	"os"
	"testing"
	"time"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// openRSCRoot is the OpenRSC source root that holds the real
// Authentic_Landscape archive + def/loc data. Overridable for CI.
func openRSCRoot() string {
	if r := os.Getenv("OPENRSC_ROOT"); r != "" {
		return r
	}
	return "/Users/flint/Code/openrsc"
}

// loadReal loads the real Facts + Landscape and precomputes the Oracle.
// The validation contract requires running against the real archive, so
// a missing root is a HARD FAIL, not a skip.
func loadReal(t *testing.T) (*Oracle, *facts.Facts, *pathfind.Landscape) {
	t.Helper()
	root := openRSCRoot()
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("OpenRSC root %q not available (set OPENRSC_ROOT): %v", root, err)
	}

	f, err := facts.Load(facts.DefaultSources(root))
	if err != nil {
		t.Fatalf("load facts: %v", err)
	}
	lpath := root + "/server/conf/server/data/Authentic_Landscape.orsc"
	ls, err := pathfind.OpenLandscape(lpath)
	if err != nil {
		t.Fatalf("open landscape %q: %v", lpath, err)
	}
	t.Cleanup(func() { ls.Close() })

	start := time.Now()
	o, err := Precompute(f, ls)
	if err != nil {
		t.Fatalf("precompute: %v", err)
	}
	t.Logf("precompute: %v, components=%d, plane-0 tiles=%d",
		time.Since(start), o.NumComponents(), o.dim*o.dim)
	return o, f, ls
}

// standableSum reports the total standable tiles across all components.
func standableSum(o *Oracle) int {
	n := 0
	for _, c := range o.comps {
		n += c.Tiles
	}
	return n
}

func TestPrecomputeShape(t *testing.T) {
	o, _, _ := loadReal(t)

	if o.NumComponents() == 0 {
		t.Fatalf("no components built")
	}
	// One dominant mainland + many pockets is the expected structure.
	var biggest int
	for _, c := range o.comps {
		if c.Tiles > biggest {
			biggest = c.Tiles
		}
	}
	t.Logf("components=%d, standable tiles=%d, largest component=%d",
		o.NumComponents(), standableSum(o), biggest)

	// Every standable tile has exactly one label; void tiles are -1;
	// component sizes sum to the standable-tile count. The oracle now
	// stacks ALL FOUR floors (band-encoded Y), so the scan covers DimY.
	var labeled, standable int
	for x := 0; x < o.dim; x++ {
		for y := 0; y < o.dimY; y++ {
			lbl := o.CompAt(x, y)
			std := o.standable(x, y)
			if std {
				standable++
			}
			if lbl >= 0 {
				labeled++
				if !std {
					t.Fatalf("tile (%d,%d) labeled %d but not standable", x, y, lbl)
				}
			} else if std {
				t.Fatalf("tile (%d,%d) standable but unlabeled", x, y)
			}
		}
	}
	if labeled != standable {
		t.Fatalf("labeled=%d != standable=%d", labeled, standable)
	}
	if got := standableSum(o); got != standable {
		t.Fatalf("component size sum=%d != standable=%d", got, standable)
	}
}

// cheb is the Chebyshev distance between two tiles.
func cheb(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// gridContains reports whether both tiles fit inside one BuildGrid window
// centered on (ax,ay): the goal must lie within [baseX,baseX+96).
func gridContains(g *pathfind.Grid, x, y int) bool {
	_, _, ok := g.World2Local(x, y)
	return ok
}

// findPathReaches reports whether the real client pathfinder finds a
// foot path from a to b inside g. The caller must have verified b is in
// g's window via gridContains.
func findPathReaches(g *pathfind.Grid, ax, ay, bx, by int) bool {
	return g.FindPathToTile(ax, ay, bx, by, false) != nil
}

// TestEngineAgreement validates the flood-fill against the REAL
// pathfinder (pathfind.FindPathToTile is the walking ground truth):
// same-component close pairs must be foot-reachable; different-component
// close pairs must NOT be.
func TestEngineAgreement(t *testing.T) {
	o, f, ls := loadReal(t)

	// The two engines intentionally answer DIFFERENT questions about
	// openable doors/gates: the Oracle treats them as crossable
	// ("reachable with effort" — the traversal flow opens them en route),
	// while the movement grid keeps a CLOSED barrier blocked until it is
	// actually opened. To test the invariant this test exists for —
	// identical GEOMETRY rules in both engines — build the grid over a
	// facts view with openable-barrier scenery filtered out, so both
	// engines model the same (openables-crossable) world.
	fOpen := *f
	fOpen.SceneryLocs = nil
	for _, loc := range f.SceneryLocs {
		if d := f.SceneryDef(loc.DefID); d.IsOpenableBarrier() {
			continue
		}
		fOpen.SceneryLocs = append(fOpen.SceneryLocs, loc)
	}
	f = &fOpen

	// Collect standable tiles in a region we know is populated (the
	// Lumbridge / Al-Kharid corridor) and pair them up. We sample on a
	// coarse stride to keep the test fast but cover many tiles.
	type pt struct{ x, y int }
	var samples []pt
	for x := 60; x < 220; x += 3 {
		for y := 540; y < 700; y += 3 {
			if o.standable(x, y) {
				samples = append(samples, pt{x, y})
			}
		}
	}
	if len(samples) < 50 {
		t.Fatalf("too few standable samples in corridor: %d", len(samples))
	}

	var sameChecked, diffChecked int
	for i := 0; i < len(samples); i++ {
		a := samples[i]
		// Build ONE local grid centered on a. The window is the same
		// 96x96 grid the real client/pathfinder uses; we only test b's
		// that actually fit inside it (gridContains), so the engine's
		// window bound is respected rather than guessed.
		g, err := pathfind.BuildGrid(ls, f, a.x, a.y, 0, nil)
		if err != nil {
			t.Fatalf("build grid at (%d,%d): %v", a.x, a.y, err)
		}
		ca := o.CompAt(a.x, a.y)
		for j := i + 1; j < len(samples); j++ {
			b := samples[j]
			if !gridContains(g, b.x, b.y) {
				continue
			}
			cb := o.CompAt(b.x, b.y)
			reaches := findPathReaches(g, a.x, a.y, b.x, b.y)

			if ca == cb {
				// Same flood component => engine MUST foot-reach — BUT
				// only when the two tiles are close enough that the
				// connecting in-component path is guaranteed to stay
				// inside the bounded 96x96 FindPath window. For far
				// same-component pairs the only route can legitimately
				// detour outside the window, so a local FindPath miss is
				// a window artifact, not a disagreement (the documented
				// "ground truth is short-range only" limitation). Skip
				// the assertion for those rather than over-claim.
				if cheb(a.x, a.y, b.x, b.y) <= 16 {
					if !reaches {
						t.Errorf("same component %d (close) but FindPath cannot reach (%d,%d)->(%d,%d)",
							ca, a.x, a.y, b.x, b.y)
					}
					sameChecked++
				}
			} else {
				// Different flood components => engine must NOT foot-reach.
				// This direction is ALWAYS valid: any in-window path
				// FindPath found would prove the tiles are one component,
				// so a reach here is a true disagreement at any distance.
				if reaches {
					t.Errorf("different components %d vs %d but FindPath reaches (%d,%d)->(%d,%d)",
						ca, cb, a.x, a.y, b.x, b.y)
				}
				diffChecked++
			}
		}
		// Cap the work so the test stays snappy but meaningful.
		if sameChecked >= 2000 && diffChecked >= 100 {
			break
		}
	}
	t.Logf("engine agreement: same-component pairs checked=%d, different-component pairs checked=%d",
		sameChecked, diffChecked)
	if sameChecked == 0 {
		t.Fatalf("no same-component pairs validated")
	}
}

// TestTollGateCorrection locks in the load-bearing correction: the
// Al-Kharid toll gate is NOT a collision cut, so Lumbridge (135,654) and
// the Al-Kharid mine (74,583) are in the SAME walking component. A future
// change that accidentally splits them must fail here.
func TestTollGateCorrection(t *testing.T) {
	o, _, _ := loadReal(t)

	const (
		lumbX, lumbY = 135, 654 // Lumbridge
		mineX, mineY = 74, 583  // Al-Kharid mine
	)

	cl, lx, ly, okL := o.CompNear(lumbX, lumbY)
	if !okL {
		t.Fatalf("could not snap Lumbridge (%d,%d) to a standable tile", lumbX, lumbY)
	}
	cm, mx, my, okM := o.CompNear(mineX, mineY)
	if !okM {
		t.Fatalf("could not snap Al-Kharid mine (%d,%d) to a standable tile", mineX, mineY)
	}

	t.Logf("Lumbridge -> comp %d (snap %d,%d); Al-Kharid mine -> comp %d (snap %d,%d)",
		cl, lx, ly, cm, mx, my)

	if cl != cm {
		t.Fatalf("toll-gate correction FAILED: Lumbridge comp %d != Al-Kharid mine comp %d "+
			"(the gate must NOT be a collision cut; it is openable and lives in game-logic)",
			cl, cm)
	}
}

// TestPOIIndexing checks that gazetteer POIs snap to standable tiles and
// bind to components, and that the component-filter works.
func TestPOIIndexing(t *testing.T) {
	o, _, _ := loadReal(t)

	pois := o.POIs()
	if len(pois) == 0 {
		t.Fatalf("no POIs indexed")
	}
	var bound, unbound int
	compSet := map[int32]bool{}
	for _, p := range pois {
		if p.Comp >= 0 {
			bound++
			compSet[p.Comp] = true
			if !o.standable(p.SnapX, p.SnapY) {
				t.Errorf("POI %s snapped to non-standable (%d,%d)", p.Type, p.SnapX, p.SnapY)
			}
			if o.CompAt(p.SnapX, p.SnapY) != p.Comp {
				t.Errorf("POI %s comp %d != CompAt(snap) %d", p.Type, p.Comp, o.CompAt(p.SnapX, p.SnapY))
			}
		} else {
			unbound++
		}
	}
	t.Logf("POIs: total=%d bound=%d unbound=%d distinct-components=%d", len(pois), bound, unbound, len(compSet))
	if bound == 0 {
		t.Fatalf("no POIs bound to a component")
	}

	// Filter sanity: POIsInComponents over the full set returns exactly
	// the bound POIs (any type).
	got := o.POIsInComponents("", compSet)
	if len(got) != bound {
		t.Fatalf("POIsInComponents(all) returned %d, want %d", len(got), bound)
	}
	// Empty set matches nothing.
	if n := len(o.POIsInComponents("", map[int32]bool{})); n != 0 {
		t.Fatalf("POIsInComponents(empty set) returned %d, want 0", n)
	}
}
