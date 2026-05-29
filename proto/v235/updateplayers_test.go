package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// writeZeroQuoted appends a zero-quoted string in the v235 wire form:
// a leading 0x00, the string bytes, then a 0x00 terminator. Mirrors
// what the server's writeZeroQuotedString emits and what
// ReadZeroQuotedString consumes.
func writeZeroQuoted(b *Buffer, s string) {
	b.WriteByte(0)
	b.WriteBytes([]byte(s))
	b.WriteByte(0)
}

// TestDecodeUpdatePlayersAppearanceCombat verifies that the type-5
// appearance record's two trailing combat bytes (combatLevel +
// skullType) are decoded onto OtherPlayerAppearance.
//
// Layout (one update record):
//
//	[short] updateCount = 1
//	[short] playerIndex
//	[byte]  updateType = 5
//	[short] appearanceID
//	[zq]    name
//	[zq]    name (full)
//	[byte]  equipmentCount = 1
//	[byte]  worn sprite
//	[byte×4] colours (hair, top, trouser, skin)
//	[byte]  combatLevel
//	[byte]  skullType
func TestDecodeUpdatePlayersAppearanceCombat(t *testing.T) {
	b := NewBuffer(64)
	b.WriteUint16(1)   // updateCount
	b.WriteUint16(42)  // playerIndex
	b.WriteByte(5)     // updateType
	b.WriteUint16(123) // appearanceID
	writeZeroQuoted(b, "Zezima")
	writeZeroQuoted(b, "Zezima")
	b.WriteByte(1)  // equipment count
	b.WriteByte(7)  // one worn sprite
	b.WriteByte(2)  // hair colour
	b.WriteByte(8)  // top colour
	b.WriteByte(14) // trouser colour
	b.WriteByte(1)  // skin colour
	b.WriteByte(76) // combatLevel
	b.WriteByte(1)  // skullType (skulled / PK-flagged)

	events, err := DecodeUpdatePlayers(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1: %+v", len(events), events)
	}
	ap, ok := events[0].(event.OtherPlayerAppearance)
	if !ok {
		t.Fatalf("got %T, want event.OtherPlayerAppearance", events[0])
	}
	if ap.PlayerIndex != 42 {
		t.Errorf("PlayerIndex: got %d, want 42", ap.PlayerIndex)
	}
	if ap.Name != "Zezima" {
		t.Errorf("Name: got %q, want %q", ap.Name, "Zezima")
	}
	if !ap.HasCombat {
		t.Fatalf("HasCombat: got false, want true")
	}
	if ap.CombatLevel != 76 {
		t.Errorf("CombatLevel: got %d, want 76", ap.CombatLevel)
	}
	if ap.SkullType != 1 {
		t.Errorf("SkullType: got %d, want 1", ap.SkullType)
	}
}

// TestDecodeUpdatePlayersWornEquipment verifies the type-5 equipment
// block decodes the per-slot worn-sprite bytes into WornSprites,
// indexed by equip slot, and that the trailing combat bytes still
// decode after a full 12-slot equipment block.
func TestDecodeUpdatePlayersWornEquipment(t *testing.T) {
	b := NewBuffer(128)
	b.WriteUint16(1)   // updateCount
	b.WriteUint16(77)  // playerIndex
	b.WriteByte(5)     // updateType = appearance
	b.WriteUint16(200) // appearanceID
	writeZeroQuoted(b, "Wieldy")
	writeZeroQuoted(b, "Wieldy")
	// Equipment block: 12 worn-sprite bytes in slot order. Put a weapon
	// sprite in slot 4 and a shield in slot 3 so we can assert by slot.
	b.WriteByte(12) // equipment count
	worn := []byte{0, 0, 0, 21, 16, 0, 0, 0, 0, 0, 0, 0}
	b.WriteBytes(worn)
	b.WriteByte(2)  // hair colour
	b.WriteByte(8)  // top colour
	b.WriteByte(14) // trouser colour
	b.WriteByte(1)  // skin colour
	b.WriteByte(51) // combatLevel
	b.WriteByte(0)  // skullType

	events, err := DecodeUpdatePlayers(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	ap := events[0].(event.OtherPlayerAppearance)
	if !ap.HasWorn {
		t.Fatal("HasWorn: got false, want true")
	}
	if ap.WornCount != 12 {
		t.Errorf("WornCount: got %d, want 12", ap.WornCount)
	}
	if ap.WornSprites[event.EquipSlotShield] != 21 {
		t.Errorf("shield slot: got %d, want 21", ap.WornSprites[event.EquipSlotShield])
	}
	if ap.WornSprites[event.EquipSlotWeapon] != 16 {
		t.Errorf("weapon slot: got %d, want 16", ap.WornSprites[event.EquipSlotWeapon])
	}
	// The combat bytes after a full equipment block must still decode.
	if !ap.HasCombat || ap.CombatLevel != 51 {
		t.Errorf("combat after equip: HasCombat=%v level=%d, want true/51", ap.HasCombat, ap.CombatLevel)
	}
}

// TestDecodeUpdatePlayersAppearanceTruncated verifies that when the
// combat bytes are absent (truncated record), HasCombat stays false
// rather than reporting a bogus combat level of 0.
func TestDecodeUpdatePlayersAppearanceTruncated(t *testing.T) {
	b := NewBuffer(64)
	b.WriteUint16(1)  // updateCount
	b.WriteUint16(7)  // playerIndex
	b.WriteByte(5)    // updateType
	b.WriteUint16(99) // appearanceID
	writeZeroQuoted(b, "Bob")
	writeZeroQuoted(b, "Bob")
	b.WriteByte(0) // equipment count = 0
	b.WriteByte(2) // hair colour
	b.WriteByte(3) // top colour
	b.WriteByte(4) // trouser colour
	b.WriteByte(5) // skin colour
	// NO combat bytes appended.

	events, err := DecodeUpdatePlayers(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	ap := events[0].(event.OtherPlayerAppearance)
	if ap.HasCombat {
		t.Errorf("HasCombat: got true, want false (combat bytes were absent)")
	}
	if ap.CombatLevel != 0 || ap.SkullType != 0 {
		t.Errorf("combat fields: got level=%d skull=%d, want 0/0", ap.CombatLevel, ap.SkullType)
	}
}

// TestDecodeUpdatePlayersDamage verifies the type-2 damage record
// decodes the cur/max hitpoints (the engaged target's health as the
// wire encodes it).
func TestDecodeUpdatePlayersDamage(t *testing.T) {
	b := NewBuffer(16)
	b.WriteUint16(1)  // updateCount
	b.WriteUint16(13) // playerIndex
	b.WriteByte(2)    // updateType = damage
	b.WriteByte(4)    // damage
	b.WriteByte(27)   // curHits
	b.WriteByte(31)   // maxHits

	events, err := DecodeUpdatePlayers(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	dmg, ok := events[0].(event.OtherPlayerDamage)
	if !ok {
		t.Fatalf("got %T, want event.OtherPlayerDamage", events[0])
	}
	if dmg.PlayerIndex != 13 || dmg.Damage != 4 || dmg.CurHits != 27 || dmg.MaxHits != 31 {
		t.Errorf("damage: got %+v, want {idx=13, dmg=4, cur=27, max=31}", dmg)
	}
}

// TestDecodeUpdatePlayersProjectile verifies type-3 (at NPC) and
// type-4 (at player) decode the caster->victim engagement and set
// VictimIsNpc correctly.
func TestDecodeUpdatePlayersProjectile(t *testing.T) {
	b := NewBuffer(32)
	b.WriteUint16(2)  // updateCount
	b.WriteUint16(5)  // caster playerIndex
	b.WriteByte(3)    // updateType = projectile at NPC
	b.WriteUint16(11) // projectile type
	b.WriteUint16(88) // victim NPC index
	b.WriteUint16(6)  // caster playerIndex
	b.WriteByte(4)    // updateType = projectile at player
	b.WriteUint16(12) // projectile type
	b.WriteUint16(90) // victim player index

	events, err := DecodeUpdatePlayers(b.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	p0 := events[0].(event.OtherPlayerProjectile)
	if !p0.VictimIsNpc || p0.CasterIndex != 5 || p0.ProjectileID != 11 || p0.VictimNpcIndex != 88 {
		t.Errorf("npc projectile: got %+v, want {caster=5, proj=11, npc=88, isNpc=true}", p0)
	}
	p1 := events[1].(event.OtherPlayerProjectile)
	if p1.VictimIsNpc || p1.CasterIndex != 6 || p1.ProjectileID != 12 || p1.VictimPlayerIndex != 90 {
		t.Errorf("player projectile: got %+v, want {caster=6, proj=12, player=90, isNpc=false}", p1)
	}
}
