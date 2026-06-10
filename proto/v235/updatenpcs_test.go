package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDecodeUpdateNpcsDamage verifies the type-2 record decodes the
// NPC's OWN cur/max hitpoints (opcode 104) — the field opcode 234 never
// carries for an NPC.
//
// Layout (one update record):
//
//	[short] updateCount = 1
//	[short] npcIndex
//	[byte]  updateType = 2
//	[byte]  damage
//	[byte]  curHits
//	[byte]  maxHits
func TestDecodeUpdateNpcsDamage(t *testing.T) {
	b := NewBuffer(16)
	b.WriteUint16(1)   // updateCount
	b.WriteUint16(305) // npcIndex
	b.WriteByte(2)     // updateType = damage/hits
	b.WriteByte(6)     // damage
	b.WriteByte(8)     // curHits
	b.WriteByte(14)    // maxHits

	events, err := DecodeUpdateNpcs(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1: %+v", len(events), events)
	}
	dmg, ok := events[0].(event.NpcDamage)
	if !ok {
		t.Fatalf("got %T, want event.NpcDamage", events[0])
	}
	if dmg.NpcIndex != 305 || dmg.Damage != 6 || dmg.CurHits != 8 || dmg.MaxHits != 14 {
		t.Errorf("npc damage: got %+v, want {idx=305, dmg=6, cur=8, max=14}", dmg)
	}
}

// TestDecodeUpdateNpcsMultiple verifies two damage records in one
// packet both decode and stay aligned.
func TestDecodeUpdateNpcsMultiple(t *testing.T) {
	b := NewBuffer(32)
	b.WriteUint16(2) // updateCount
	b.WriteUint16(10)
	b.WriteByte(2)
	b.WriteByte(1)
	b.WriteByte(9)
	b.WriteByte(10)
	b.WriteUint16(11)
	b.WriteByte(2)
	b.WriteByte(3)
	b.WriteByte(2)
	b.WriteByte(5)

	events, err := DecodeUpdateNpcs(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	d0 := events[0].(event.NpcDamage)
	d1 := events[1].(event.NpcDamage)
	if d0.NpcIndex != 10 || d0.CurHits != 9 || d0.MaxHits != 10 {
		t.Errorf("record 0: got %+v, want idx=10 cur=9 max=10", d0)
	}
	if d1.NpcIndex != 11 || d1.CurHits != 2 || d1.MaxHits != 5 {
		t.Errorf("record 1: got %+v, want idx=11 cur=2 max=5", d1)
	}
}

// TestDecodeUpdateNpcsChat verifies type-1 NPC chat decodes the
// recipient index and a round-trippable RSC-compressed body.
func TestDecodeUpdateNpcsChat(t *testing.T) {
	msg := "hello adventurer"
	enc := EncipherRSCString(msg)

	b := NewBuffer(64)
	b.WriteUint16(1)      // updateCount
	b.WriteUint16(42)     // npcIndex
	b.WriteByte(1)        // updateType = chat
	b.WriteUint16(0xFFFF) // recipientIndex = -1 (broadcast)
	b.WriteSmart08_16(len(msg))
	b.WriteBytes(enc)

	events, err := DecodeUpdateNpcs(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	chat, ok := events[0].(event.NpcChat)
	if !ok {
		t.Fatalf("got %T, want event.NpcChat", events[0])
	}
	if chat.NpcIndex != 42 {
		t.Errorf("NpcIndex: got %d, want 42", chat.NpcIndex)
	}
	if chat.RecipientIndex != -1 {
		t.Errorf("RecipientIndex: got %d, want -1", chat.RecipientIndex)
	}
	if chat.MessageText != msg {
		t.Errorf("MessageText: got %q, want %q", chat.MessageText, msg)
	}
}
