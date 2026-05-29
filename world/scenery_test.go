package world

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestApplySceneryUpdatesResolvesAbsolute checks that SEND_SCENERY_HANDLER
// deltas (player-relative offsets) land as absolute tiles in
// DynamicScenery, and that the 60000 sentinel removes a tile.
func TestApplySceneryUpdatesResolvesAbsolute(t *testing.T) {
	w := NewWorld()
	w.Self.SetPosition(Coord{X: 220, Y: 3110})

	// A fire (def 97) lit at the player's own tile (offset 0,0), plus a
	// range (def 11) five tiles east.
	w.Apply(event.SceneryUpdates{Updates: []event.SceneryDelta{
		{ID: 97, OffsetX: 0, OffsetY: 0},
		{ID: 11, OffsetX: 5, OffsetY: 0},
	}})

	fire, ok := w.Scenery.At(220, 3110)
	if !ok || fire.ID != 97 {
		t.Fatalf("fire at (220,3110): got %+v ok=%v, want ID 97", fire, ok)
	}
	rng, ok := w.Scenery.At(225, 3110)
	if !ok || rng.ID != 11 {
		t.Fatalf("range at (225,3110): got %+v ok=%v, want ID 11", rng, ok)
	}

	// Now the fire burns out — server sends the removal sentinel at the
	// same tile (still offset 0,0 since we haven't moved).
	w.Apply(event.SceneryUpdates{Updates: []event.SceneryDelta{
		{ID: SceneryRemoveSentinel, OffsetX: 0, OffsetY: 0},
	}})
	if _, ok := w.Scenery.At(220, 3110); ok {
		t.Errorf("fire tile should be cleared after removal sentinel")
	}
	if _, ok := w.Scenery.At(225, 3110); !ok {
		t.Errorf("range tile should be untouched by the fire removal")
	}
}

// TestDynamicSceneryNear checks the radius query used by
// world.scenery.by_id(id, radius=R).
func TestDynamicSceneryNear(t *testing.T) {
	s := NewDynamicScenery()
	s.Add(100, 100, 97)
	s.Add(103, 100, 97) // 3 tiles east
	s.Add(100, 110, 11) // 10 tiles north

	within2 := s.Near(100, 100, 2)
	if len(within2) != 1 {
		t.Fatalf("Near(100,100,2): got %d, want 1 (only the on-tile fire)", len(within2))
	}
	within5 := s.Near(100, 100, 5)
	if len(within5) != 2 {
		t.Fatalf("Near(100,100,5): got %d, want 2", len(within5))
	}
}
