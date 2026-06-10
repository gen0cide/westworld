package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDropChurnedRemovals verifies the in-combat remove+readd collapse:
// when an NPC index appears as BOTH a REMOVE and a re-added NEW record
// in one opcode-79 packet (the server's inCombat() remove + re-add with
// the combat sprite), the REMOVE is dropped so the engaged NPC isn't
// pruned every tick — while a REMOVE with no matching re-add (a true
// despawn/death) survives. (#combat-prune)
func TestDropChurnedRemovals(t *testing.T) {
	events := []event.Event{
		event.NpcNearby{Index: 10, Removed: true},           // churn: removed...
		event.NpcNearby{Index: 11, Removed: true},           // true despawn
		event.NpcNearby{Index: 10, TypeID: 19, IsNew: true}, // ...then re-added
		event.NpcNearby{Index: 12, X: 5, Y: 5},              // movement, untouched
	}
	out := dropChurnedRemovals(events)

	var sawRemove10, sawRemove11, sawNew10, sawMove12 bool
	for _, ev := range out {
		n, ok := ev.(event.NpcNearby)
		if !ok {
			continue
		}
		switch {
		case n.Index == 10 && n.Removed:
			sawRemove10 = true
		case n.Index == 11 && n.Removed:
			sawRemove11 = true
		case n.Index == 10 && n.IsNew:
			sawNew10 = true
		case n.Index == 12 && !n.Removed && !n.IsNew:
			sawMove12 = true
		}
	}
	if sawRemove10 {
		t.Error("churned REMOVE for re-added npc 10 should be dropped")
	}
	if !sawRemove11 {
		t.Error("true despawn REMOVE for npc 11 should survive")
	}
	if !sawNew10 {
		t.Error("re-add NEW for npc 10 should survive")
	}
	if !sawMove12 {
		t.Error("movement for npc 12 should survive")
	}
}

// TestDropChurnedRemovalsNoNew is the fast path: with no re-adds, the
// event slice is returned unchanged (every REMOVE is a real despawn).
func TestDropChurnedRemovalsNoNew(t *testing.T) {
	events := []event.Event{
		event.NpcNearby{Index: 1, Removed: true},
		event.NpcNearby{Index: 2, X: 3, Y: 3},
	}
	out := dropChurnedRemovals(events)
	if len(out) != len(events) {
		t.Fatalf("no-new fast path changed length: got %d, want %d", len(out), len(events))
	}
}
