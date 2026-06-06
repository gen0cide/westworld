package world

import (
	"reflect"
	"testing"
)

// TestNpcOrderChurnReappendsAtEnd locks in the opcode-79 fix: when the
// server re-sends an ALREADY-tracked NPC as a "new" record (the in-combat
// remove + re-append-at-end churn), our order mirror must move that index
// to the END — not leave it in place — or the slot->index map desyncs from
// the server's localNpcs and every later positional update is
// misattributed (NPCs silently vanish from world.npcs). Regression for the
// mining-instructor-disappears bug found driving stubbs through tutorial
// island.
func TestNpcOrderChurnReappendsAtEnd(t *testing.T) {
	s := NewNpcsState()
	for _, idx := range []int{10, 11, 12, 13} {
		s.Set(NpcRecord{Index: idx, TypeID: 1, X: 100, Y: 100})
	}
	if got, want := s.Order(), []int{10, 11, 12, 13}; !reflect.DeepEqual(got, want) {
		t.Fatalf("initial order = %v, want %v", got, want)
	}

	// NPC 11 is in combat: server removes it from its slot and re-appends
	// at the end, emitting a fresh "new" record. Our Set must mirror that.
	s.Set(NpcRecord{Index: 11, TypeID: 1, X: 101, Y: 100})
	if got, want := s.Order(), []int{10, 12, 13, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after churn order = %v, want %v (11 must move to end)", got, want)
	}

	// Movement updates must NOT reorder (localNpcs order is stable for a
	// plain move) — only re-append-as-new does.
	s.MoveBy(12, 1, 0, 2)
	if got, want := s.Order(), []int{10, 12, 13, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after move order = %v, want %v (move must not reorder)", got, want)
	}

	// Combat hit-points accumulated before the churn survive the re-add.
	s.SetHits(11, 3, 7, 10)
	s.Set(NpcRecord{Index: 11, TypeID: 1, X: 102, Y: 100})
	if r, _ := s.Get(11); !r.HasHits || r.CurHits != 7 {
		t.Fatalf("churn lost combat state: %+v", r)
	}
}
