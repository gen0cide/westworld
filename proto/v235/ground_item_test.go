package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDecodeGroundItem_MultiEntry guards the opcode-99 fix: the server streams
// EVERY in-view ground item in one packet, so the decoder must read all entries.
// The old code read only the first — which silently hid items like the free
// Bronze Pickaxe (156) on the Barbarian-Village table whenever it wasn't first.
func TestDecodeGroundItem_MultiEntry(t *testing.T) {
	payload := []byte{
		0x00, 0x0A, 0x05, 0x05, // ADD item 10 at offset (+5,+5)            (NOT the pickaxe)
		0x00, 0x9C, 0x03, 0xFE, // ADD Bronze Pickaxe 156 at (+3,-2)        (2nd entry — old decoder dropped it)
		0x80, 0xFA, 0x01, 0x01, // in-range REMOVE 250 (0x80FA) at (+1,+1)
		0xFF, 0xFB, 0x04, // out-of-range CLEAR at (-5,+4)            ([255][x][y], no id)
	}
	ev, err := decodeGroundItem(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	gu, ok := ev.(event.GroundItemUpdates)
	if !ok {
		t.Fatalf("want event.GroundItemUpdates, got %T", ev)
	}
	want := []event.GroundItemDelta{
		{ItemID: 10, OffsetX: 5, OffsetY: 5, Disappear: false},
		{ItemID: 156, OffsetX: 3, OffsetY: -2, Disappear: false},
		{ItemID: 250, OffsetX: 1, OffsetY: 1, Disappear: true},
		{ItemID: 0, OffsetX: -5, OffsetY: 4, Disappear: true},
	}
	if len(gu.Updates) != len(want) {
		t.Fatalf("got %d updates, want %d: %+v", len(gu.Updates), len(want), gu.Updates)
	}
	for i, w := range want {
		if gu.Updates[i] != w {
			t.Errorf("update[%d] = %+v, want %+v", i, gu.Updates[i], w)
		}
	}

	// The decisive assertion: the pickaxe IS decoded even though it is the
	// SECOND entry (the exact case the old single-entry decoder dropped).
	found := false
	for _, u := range gu.Updates {
		if u.ItemID == 156 && !u.Disappear {
			found = true
		}
	}
	if !found {
		t.Fatal("Bronze Pickaxe (156) not decoded — it would still be invisible to perception")
	}
}

// TestDecodeRemoveWorldEntity guards the opcode-211 decode: a sequence of
// [short offX][short offY] SIGNED player-relative points (the offsets exceed
// int8 range — that is why this opcode exists).
func TestDecodeRemoveWorldEntity(t *testing.T) {
	payload := []byte{
		0x00, 0x90, 0xFF, 0x70, // (+144, -144)
		0xFF, 0x38, 0x00, 0xC8, // (-200, +200)
	}
	ev, err := decodeRemoveWorldEntity(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	rw, ok := ev.(event.RemoveWorldEntities)
	if !ok {
		t.Fatalf("want event.RemoveWorldEntities, got %T", ev)
	}
	want := []event.RemovePoint{
		{OffsetX: 144, OffsetY: -144},
		{OffsetX: -200, OffsetY: 200},
	}
	if len(rw.Points) != len(want) {
		t.Fatalf("got %d points, want %d: %+v", len(rw.Points), len(want), rw.Points)
	}
	for i, w := range want {
		if rw.Points[i] != w {
			t.Errorf("point[%d] = %+v, want %+v", i, rw.Points[i], w)
		}
	}
}
