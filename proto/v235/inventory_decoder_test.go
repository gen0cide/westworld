package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDecodeInventorySlotUpdateStackable covers the case where the
// item in the slot is stackable — the server includes a smart-
// encoded amount field after the id+wielded short.
func TestDecodeInventorySlotUpdateStackable(t *testing.T) {
	// slot=3, id=619 (coins, stackable), wielded=false, amount=42
	// (2-byte smart form because 42 < 0x80 high bit clear).
	b := NewBuffer(16)
	b.WriteByte(3)
	b.WriteUint16(619)
	b.WriteUint16(42)
	payload := b.Bytes()

	ev, err := decodeInventorySlotUpdate(payload, func(int) bool { return true })
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	su, ok := ev.(event.InventorySlotUpdate)
	if !ok {
		t.Fatalf("got %T, want event.InventorySlotUpdate", ev)
	}
	if su.Slot != 3 {
		t.Errorf("slot: got %d, want 3", su.Slot)
	}
	if su.Item == nil || su.Item.ItemID != 619 || su.Item.Amount != 42 || su.Item.Wielded {
		t.Errorf("item: got %+v, want {619, amount=42, wielded=false}", su.Item)
	}
}

// TestDecodeInventorySlotUpdateNonStackable covers the bug — the
// fix is that a non-stackable item must NOT consume an amount field
// (the server doesn't write one). Prior bug: read whatever garbage
// bytes followed and reported amounts like 10879108.
func TestDecodeInventorySlotUpdateNonStackable(t *testing.T) {
	// slot=0, id=87 (bronze axe, non-stackable), wielded=false.
	// Pad with extra bytes to simulate that the prior bug would
	// happily read them as an amount.
	b := NewBuffer(16)
	b.WriteByte(0)
	b.WriteUint16(87)
	b.WriteUint32(0xDEADBEEF) // garbage following the slot — must NOT be read
	payload := b.Bytes()

	ev, err := decodeInventorySlotUpdate(payload, func(id int) bool {
		// Bronze axe (87) is NOT stackable. Real impl threads
		// facts.ItemDef.IsStackable through; here we hardcode.
		return false
	})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	su, ok := ev.(event.InventorySlotUpdate)
	if !ok {
		t.Fatalf("got %T, want event.InventorySlotUpdate", ev)
	}
	if su.Item == nil {
		t.Fatal("item nil; want non-nil with amount=1")
	}
	if su.Item.ItemID != 87 {
		t.Errorf("item_id: got %d, want 87", su.Item.ItemID)
	}
	if su.Item.Amount != 1 {
		t.Errorf("amount: got %d, want 1 (non-stackable items default to 1, not 0xDEADBEEF)", su.Item.Amount)
	}
}

// TestDecodeInventorySlotUpdateClear covers the slot-cleared case:
// raw id field of 0 means "this slot is now empty"; Item should be
// nil.
func TestDecodeInventorySlotUpdateClear(t *testing.T) {
	b := NewBuffer(16)
	b.WriteByte(5)
	b.WriteUint16(0)
	payload := b.Bytes()

	ev, err := decodeInventorySlotUpdate(payload, func(int) bool { return false })
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	su, ok := ev.(event.InventorySlotUpdate)
	if !ok {
		t.Fatalf("got %T, want event.InventorySlotUpdate", ev)
	}
	if su.Slot != 5 {
		t.Errorf("slot: got %d, want 5", su.Slot)
	}
	if su.Item != nil {
		t.Errorf("item: got %+v, want nil (slot cleared)", su.Item)
	}
}

// TestDecodeInventoryMixed exercises the full-snapshot decoder with
// a mix of stackable + non-stackable items in sequence — the bug
// here was that reading the amount unconditionally realigned every
// subsequent slot. With the fix, slot[N+1] starts where the prior
// slot's stackable-or-not encoding ended.
func TestDecodeInventoryMixed(t *testing.T) {
	b := NewBuffer(16)
	b.WriteByte(3) // 3 slots
	// slot 0: non-stackable axe (87), wielded=false, no amount
	b.WriteUint16(87)
	// slot 1: stackable coins (619), amount=100
	b.WriteUint16(619)
	b.WriteUint16(100)
	// slot 2: non-stackable shield (1234), wielded=true, no amount
	b.WriteUint16(1234 | 0x8000)
	payload := b.Bytes()

	isStackable := func(id int) bool { return id == 619 }
	ev, err := decodeInventory(payload, isStackable)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	snap, ok := ev.(event.InventorySnapshot)
	if !ok {
		t.Fatalf("got %T, want event.InventorySnapshot", ev)
	}
	if len(snap.Items) != 3 {
		t.Fatalf("items: got %d, want 3", len(snap.Items))
	}
	if snap.Items[0].ItemID != 87 || snap.Items[0].Amount != 1 || snap.Items[0].Wielded {
		t.Errorf("slot 0: got %+v, want {87, amount=1, wielded=false}", snap.Items[0])
	}
	if snap.Items[1].ItemID != 619 || snap.Items[1].Amount != 100 || snap.Items[1].Wielded {
		t.Errorf("slot 1: got %+v, want {619, amount=100, wielded=false}", snap.Items[1])
	}
	if snap.Items[2].ItemID != 1234 || snap.Items[2].Amount != 1 || !snap.Items[2].Wielded {
		t.Errorf("slot 2: got %+v, want {1234, amount=1, wielded=true}", snap.Items[2])
	}
}
