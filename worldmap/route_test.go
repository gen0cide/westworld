package worldmap

import "testing"

// TestRouteSamePlane plans a corridor across Lumbridge open ground and
// asserts waypoints land on standable tiles of the same plane.
func TestRouteSamePlane(t *testing.T) {
	o, _, _ := loadReal(t)
	wps := o.Route(134, 644, 126, 651, nil)
	if len(wps) == 0 {
		t.Fatal("no corridor across open Lumbridge ground")
	}
	for _, wp := range wps {
		if !o.standable(wp[0], wp[1]) {
			t.Fatalf("waypoint (%d,%d) not standable", wp[0], wp[1])
		}
		if wp[1]/planeHeight != 0 {
			t.Fatalf("waypoint (%d,%d) left plane 0", wp[0], wp[1])
		}
	}
}

// TestRouteCrossPlaneNil: cross-plane routing is the ladder router's job —
// Route must refuse rather than invent a vertical path.
func TestRouteCrossPlaneNil(t *testing.T) {
	o, _, _ := loadReal(t)
	if wps := o.Route(134, 644, 134, 644+944, nil); wps != nil {
		t.Fatalf("cross-plane Route must be nil, got %d waypoints", len(wps))
	}
}

// TestPerPlaneComponents guards the "non-congruent fields" fix: tiles on
// plane 1 carry their own components, never plane 0's, and CompNear never
// snaps across a floor band. (Lumbridge castle first floor exists at
// footprint (134,650)+944.)
func TestPerPlaneComponents(t *testing.T) {
	o, _, _ := loadReal(t)
	// Find real standable tiles around Lumbridge castle on BOTH floors
	// (don't guess interiors), then assert the floors never share a
	// component and CompNear never snaps across a band.
	groundComps := map[int32]bool{}
	upComps := map[int32]bool{}
	for x := 120; x <= 140; x++ {
		for fy := 640; fy <= 660; fy++ {
			if c := o.CompAt(x, fy); c >= 0 {
				groundComps[c] = true
			}
			if c := o.CompAt(x, fy+planeHeight); c >= 0 {
				upComps[c] = true
			}
		}
	}
	if len(groundComps) == 0 {
		t.Fatal("no standable plane-0 tiles around Lumbridge castle")
	}
	if len(upComps) == 0 {
		t.Fatal("no standable plane-1 tiles above Lumbridge castle — upper floor missing from the oracle")
	}
	for c := range upComps {
		if groundComps[c] {
			t.Fatalf("component %d spans plane 0 AND plane 1 — floors are connected", c)
		}
	}
	if _, _, sy, ok := o.CompNear(134, planeHeight); ok && sy/planeHeight != 1 {
		t.Fatalf("CompNear snapped across the band: footprintY %d", sy%planeHeight)
	}
}
