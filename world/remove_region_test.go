package world

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestApplyRemoveWorldEntities exercises the opcode-211 bulk region clear:
// every dynamic boundary/scenery/ground-item entry in a swept 8x8 region
// (abs>>3) must drop; entries in other regions must survive. This is the
// server's ONLY eviction channel for far-away dynamic state — without it the
// stores accumulate stale entries that would feed false collision forever.
func TestApplyRemoveWorldEntities(t *testing.T) {
	w := NewWorld()
	w.Self.SetPosition(Coord{X: 200, Y: 500})

	// Region A = (152,496)>>3 = (19,62). Region B = (240,560)>>3 = (30,70).
	w.Boundaries.Set(152, 496, 0, 11) // A: an opened door override
	w.Boundaries.Set(240, 560, 1, 2)  // B
	w.Scenery.Add(153, 497, 60)       // A: a closed gate
	w.Scenery.Remove(154, 498)        // A: a removal mark (depleted rock)
	w.Scenery.Add(241, 561, 97)       // B
	w.GroundItems.Add(155, 499, 156)  // A: a pickaxe
	w.GroundItems.Add(242, 562, 10)   // B

	// The server clears region A: point offset = (152-200, 496-500) = (-48,-4).
	w.Apply(event.RemoveWorldEntities{Points: []event.RemovePoint{
		{OffsetX: -48, OffsetY: -4},
	}})

	if _, ok := w.Boundaries.Get(152, 496, 0); ok {
		t.Error("boundary override in the swept region must drop")
	}
	if _, ok := w.Scenery.At(153, 497); ok {
		t.Error("scenery record in the swept region must drop")
	}
	if w.Scenery.IsRemoved(154, 498) {
		t.Error("scenery removal mark in the swept region must drop")
	}
	if _, ok := w.GroundItems.m[[2]int{155, 499}]; ok {
		t.Error("ground item in the swept region must drop")
	}

	// Region B untouched.
	if _, ok := w.Boundaries.Get(240, 560, 1); !ok {
		t.Error("boundary outside the swept region must survive")
	}
	if _, ok := w.Scenery.At(241, 561); !ok {
		t.Error("scenery outside the swept region must survive")
	}
	if _, ok := w.GroundItems.m[[2]int{242, 562}]; !ok {
		t.Error("ground item outside the swept region must survive")
	}
}

// TestApplyRemoveWorldEntitiesWireOrder guards the offset-anchoring contract:
// the server sends position updates BEFORE the per-tick entity deltas and the
// 211 clear (GameStateUpdater.sendUpdatePackets), so a 211 arriving after a
// move must resolve its offsets against the NEW position under our serial
// Apply. If Apply ever becomes non-serial/batched, this catches it.
func TestApplyRemoveWorldEntitiesWireOrder(t *testing.T) {
	w := NewWorld()
	w.Self.SetPosition(Coord{X: 100, Y: 100})
	w.Scenery.Add(160, 100, 60) // region (20,12)

	// The host moves east, THEN the 211 for the region around the old camp
	// arrives — offsets are relative to the NEW position (the wire order).
	w.Self.SetPosition(Coord{X: 130, Y: 100})
	w.Apply(event.RemoveWorldEntities{Points: []event.RemovePoint{
		{OffsetX: 30, OffsetY: 0}, // 130+30=160 -> region (20,12)
	}})

	if _, ok := w.Scenery.At(160, 100); ok {
		t.Error("211 offsets must resolve against the CURRENT (post-move) position")
	}
}
