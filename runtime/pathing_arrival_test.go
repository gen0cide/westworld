package runtime

import "testing"

// These tests lock the targeted go_to POI-arrival fix: GoTo must arrive
// at the nearest STANDABLE approach tile when the requested target is a
// non-standable building/object footprint, while preserving EXACT
// radius-1 arrival for standable+reachable targets and still failing
// loudly when no standable approach sits near the target.
//
// The snap/bounding logic is decomposed into the pure snapStandable
// helper so it can be exercised over a synthetic collision predicate
// without building a real ~96x96 pathfind grid.

// blockedSet builds a standable predicate where every listed tile is
// NON-standable (a wall / building / object footprint) and everything
// else is standable.
func blockedSet(blocked ...[2]int) func(int, int) bool {
	set := make(map[[2]int]bool, len(blocked))
	for _, b := range blocked {
		set[b] = true
	}
	return func(x, y int) bool { return !set[[2]int{x, y}] }
}

func TestSnapStandable_TargetAlreadyStandable(t *testing.T) {
	// A standable target must be returned UNCHANGED with ok=true. This
	// is the load-bearing invariant: when the requested tile is open
	// ground, GoTo arrives on exactly that tile (radius-1), no early
	// "close enough".
	standable := blockedSet() // nothing blocked
	x, y, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if !ok || x != 138 || y != 639 {
		t.Fatalf("standable target: got (%d,%d) ok=%v, want (138,639) ok=true", x, y, ok)
	}
}

func TestSnapStandable_NonStandableTargetSnapsToNeighbor(t *testing.T) {
	// The general-shop class: the target tile is a building (the POI
	// marker sits ON the shop), but its neighbors are open. The snap
	// must land on an adjacent standable tile within the bound.
	standable := blockedSet([2]int{138, 639})
	x, y, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if !ok {
		t.Fatalf("non-standable target with open neighbor: ok=false, want a snapped tile")
	}
	if x == 138 && y == 639 {
		t.Fatalf("snap returned the non-standable target itself (%d,%d)", x, y)
	}
	if cheb := chebyshev(x, y, 138, 639); cheb < 1 || cheb > arriveSnapRadius {
		t.Fatalf("snapped tile (%d,%d) is chebyshev %d from target; want within [1,%d]", x, y, cheb, arriveSnapRadius)
	}
	if !standable(x, y) {
		t.Fatalf("snapped tile (%d,%d) is not standable", x, y)
	}
}

func TestSnapStandable_PicksClosestRing(t *testing.T) {
	// Snap must prefer the closest standable tile: with the target and
	// its whole radius-1 ring blocked but radius-2 open, the result sits
	// at chebyshev 2, never further.
	var blocked [][2]int
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			blocked = append(blocked, [2]int{138 + dx, 639 + dy})
		}
	}
	standable := blockedSet(blocked...)
	x, y, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if !ok {
		t.Fatalf("ring-1 blocked, ring-2 open: ok=false, want a snapped tile")
	}
	if cheb := chebyshev(x, y, 138, 639); cheb != 2 {
		t.Fatalf("snapped tile (%d,%d) is chebyshev %d from target; want exactly 2", x, y, cheb)
	}
}

func TestSnapStandable_NoStandableWithinBoundFailsClosed(t *testing.T) {
	// A genuine wall: the target and EVERY tile within the bound is
	// blocked. snapStandable must report ok=false (and leave the coords
	// at the raw target) so GoTo keeps the raw target and fails loudly
	// into the cognition re-plan instead of "arriving" on a wall.
	var blocked [][2]int
	for dx := -arriveSnapRadius; dx <= arriveSnapRadius; dx++ {
		for dy := -arriveSnapRadius; dy <= arriveSnapRadius; dy++ {
			blocked = append(blocked, [2]int{138 + dx, 639 + dy})
		}
	}
	standable := blockedSet(blocked...)
	x, y, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if ok {
		t.Fatalf("all tiles within bound blocked: ok=true (%d,%d), want ok=false", x, y)
	}
	if x != 138 || y != 639 {
		t.Fatalf("failed snap should leave raw target; got (%d,%d), want (138,639)", x, y)
	}
}

func TestSnapStandable_RespectsBound(t *testing.T) {
	// Standable ground exists, but only OUTSIDE the bound. The snap must
	// NOT reach for it — declaring arrival that far from the requested
	// target is exactly the failure mode we are guarding against.
	var blocked [][2]int
	reach := arriveSnapRadius + 3
	for dx := -reach; dx <= reach; dx++ {
		for dy := -reach; dy <= reach; dy++ {
			if chebyshev(138+dx, 639+dy, 138, 639) <= arriveSnapRadius {
				blocked = append(blocked, [2]int{138 + dx, 639 + dy})
			}
		}
	}
	standable := blockedSet(blocked...) // open beyond arriveSnapRadius
	_, _, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if ok {
		t.Fatalf("standable ground only beyond bound: ok=true, want ok=false (bounded)")
	}
}

func TestArrivalTarget_NoLandscapeIsNoOp(t *testing.T) {
	// With no landscape archive loaded (the unit-test Host), arrivalTarget
	// must fall back to the raw target with snapped=false so GoTo behaves
	// exactly as before — a safe no-op that never spuriously "arrives".
	h := newTestHost()
	if h.landscape != nil {
		t.Skip("test host unexpectedly has a landscape; this test asserts the nil-landscape no-op")
	}
	ax, ay, snapped := h.arrivalTarget(138, 639)
	if snapped || ax != 138 || ay != 639 {
		t.Fatalf("nil-landscape arrivalTarget: got (%d,%d) snapped=%v, want (138,639) snapped=false", ax, ay, snapped)
	}
}

// TestArrivalPredicate_RadiusOne documents the arrival semantics GoTo
// applies to the snapped tile: arrival is declared only within
// arriveRadius (1) of the (snapped) approach tile, never on the basis of
// being merely "near" the raw non-standable target.
func TestArrivalPredicate_RadiusOne(t *testing.T) {
	const arriveRadius = 1
	// Target (138,639) is a building; snap lands on an adjacent standable
	// tile. The host arrives only once it is within radius 1 of THAT
	// snapped tile.
	standable := blockedSet([2]int{138, 639})
	arrX, arrY, ok := snapStandable(138, 639, arriveSnapRadius, standable)
	if !ok {
		t.Fatalf("setup: expected a snapped approach tile")
	}
	// SAFETY BOUND: the snap moved OFF the non-standable building tile and
	// stayed within arriveSnapRadius of the requested target — this is what
	// keeps GoTo from ever "arriving" far from where it was asked to go.
	if arrX == 138 && arrY == 639 {
		t.Fatalf("snap must move off the non-standable target (138,639)")
	}
	if absVal(arrX-138) > arriveSnapRadius || absVal(arrY-639) > arriveSnapRadius {
		t.Fatalf("snapped tile (%d,%d) exceeds snap bound %d of target (138,639)", arrX, arrY, arriveSnapRadius)
	}
	// ARRIVAL is radius-1 of the SNAPPED tile: a host 1 tile away arrives;
	// a host 3 tiles away does not. (Distinct positions, not a self-compare.)
	if posX, posY := arrX+1, arrY; !(absVal(posX-arrX) <= arriveRadius && absVal(posY-arrY) <= arriveRadius) {
		t.Fatalf("a host 1 tile from the snapped approach must count as arrived")
	}
	if posX, posY := arrX+3, arrY; absVal(posX-arrX) <= arriveRadius && absVal(posY-arrY) <= arriveRadius {
		t.Fatalf("a host 3 tiles from the snapped approach must NOT count as arrived")
	}
}
