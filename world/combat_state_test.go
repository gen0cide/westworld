package world

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestApplyAppearanceCombat checks that the combat-state bytes from a
// type-5 appearance update land on the player mirror.
func TestApplyAppearanceCombat(t *testing.T) {
	w := NewWorld()
	w.Apply(event.OtherPlayerAppearance{
		PlayerIndex: 42,
		Name:        "Zezima",
		CombatLevel: 76,
		SkullType:   1,
		HasCombat:   true,
	})
	rec, ok := w.Players.Get(42)
	if !ok {
		t.Fatal("player 42 not recorded")
	}
	if !rec.HasAppearanceCombat {
		t.Fatal("HasAppearanceCombat: got false, want true")
	}
	if rec.CombatLevel != 76 || rec.SkullType != 1 {
		t.Errorf("combat: got level=%d skull=%d, want 76/1", rec.CombatLevel, rec.SkullType)
	}
	if rec.Name != "Zezima" {
		t.Errorf("Name: got %q, want Zezima", rec.Name)
	}
}

// TestApplyAppearanceCombatSkippedWhenAbsent checks that an appearance
// update without combat bytes leaves HasAppearanceCombat false.
func TestApplyAppearanceCombatSkippedWhenAbsent(t *testing.T) {
	w := NewWorld()
	w.Apply(event.OtherPlayerAppearance{PlayerIndex: 7, Name: "Bob", HasCombat: false})
	rec, ok := w.Players.Get(7)
	if !ok {
		t.Fatal("player 7 not recorded")
	}
	if rec.HasAppearanceCombat {
		t.Error("HasAppearanceCombat: got true, want false")
	}
}

// TestApplyPlayerDamageLandsHits checks that a type-2 damage update
// lands the engaged target's cur/max hitpoints on the player mirror.
func TestApplyPlayerDamageLandsHits(t *testing.T) {
	w := NewWorld()
	w.Apply(event.OtherPlayerDamage{PlayerIndex: 13, Damage: 4, CurHits: 27, MaxHits: 31})
	rec, ok := w.Players.Get(13)
	if !ok {
		t.Fatal("player 13 not recorded")
	}
	if !rec.HasHits {
		t.Fatal("HasHits: got false, want true")
	}
	if rec.CurHits != 27 || rec.MaxHits != 31 || rec.LastDamage != 4 {
		t.Errorf("hits: got cur=%d max=%d dmg=%d, want 27/31/4", rec.CurHits, rec.MaxHits, rec.LastDamage)
	}
	if rec.LastDamageAt.IsZero() {
		t.Error("LastDamageAt not stamped")
	}
}

// TestApplyProjectileEngagement checks that a type-3 projectile lands
// the engagement on the caster and the incoming-attack on the victim
// NPC (who-is-fighting-whom).
func TestApplyProjectileEngagement(t *testing.T) {
	w := NewWorld()
	w.Apply(event.OtherPlayerProjectile{
		CasterIndex:    5,
		ProjectileID:   11,
		VictimNpcIndex: 88,
		VictimIsNpc:    true,
	})
	caster, ok := w.Players.Get(5)
	if !ok {
		t.Fatal("caster 5 not recorded")
	}
	if caster.EngagedNpcIndex != 88 || caster.EngagedPlayerIndex != -1 || caster.ProjectileID != 11 {
		t.Errorf("engagement: got npc=%d player=%d proj=%d, want 88/-1/11",
			caster.EngagedNpcIndex, caster.EngagedPlayerIndex, caster.ProjectileID)
	}
	if caster.EngagedAt.IsZero() {
		t.Error("EngagedAt not stamped")
	}
	npc, ok := w.Npcs.Get(88)
	if !ok {
		t.Fatal("victim NPC 88 not recorded")
	}
	if npc.IncomingFromPlayerIndex != 5 || npc.IncomingProjectileID != 11 {
		t.Errorf("npc incoming: got from=%d proj=%d, want 5/11",
			npc.IncomingFromPlayerIndex, npc.IncomingProjectileID)
	}
}

// TestNpcSetPreservesCombatState checks that a per-tick position
// update (NpcsState.Set) does NOT clobber accumulated combat state.
func TestNpcSetPreservesCombatState(t *testing.T) {
	s := NewNpcsState()
	s.SetHits(88, 3, 5, 10)
	s.SetIncomingAttack(88, 5, 11)
	// Now a position/type update arrives (carries no combat fields).
	s.Set(NpcRecord{Index: 88, X: 100, Y: 200, TypeID: 3})
	rec, ok := s.Get(88)
	if !ok {
		t.Fatal("npc 88 not recorded")
	}
	if rec.X != 100 || rec.Y != 200 || rec.TypeID != 3 {
		t.Errorf("position: got x=%d y=%d type=%d, want 100/200/3", rec.X, rec.Y, rec.TypeID)
	}
	if !rec.HasHits || rec.CurHits != 5 || rec.MaxHits != 10 {
		t.Errorf("hits clobbered by position update: got hasHits=%v cur=%d max=%d", rec.HasHits, rec.CurHits, rec.MaxHits)
	}
	if rec.IncomingFromPlayerIndex != 5 {
		t.Errorf("incoming attacker clobbered: got %d, want 5", rec.IncomingFromPlayerIndex)
	}
}

// TestApplyNpcDamageLandsHits checks that an opcode-104 type-2 NPC
// damage event lands the NPC's OWN cur/max hitpoints on the npc mirror
// (un-stubbing Npc.health for any visible NPC).
func TestApplyNpcDamageLandsHits(t *testing.T) {
	w := NewWorld()
	w.Apply(event.NpcDamage{NpcIndex: 305, Damage: 6, CurHits: 8, MaxHits: 14})
	rec, ok := w.Npcs.Get(305)
	if !ok {
		t.Fatal("npc 305 not recorded")
	}
	if !rec.HasHits {
		t.Fatal("HasHits: got false, want true")
	}
	if rec.CurHits != 8 || rec.MaxHits != 14 || rec.LastDamage != 6 {
		t.Errorf("hits: got cur=%d max=%d dmg=%d, want 8/14/6", rec.CurHits, rec.MaxHits, rec.LastDamage)
	}
	if rec.LastDamageAt.IsZero() {
		t.Error("LastDamageAt not stamped")
	}
}

// TestApplyAppearanceWornEquipment checks that the type-5 appearance
// update's per-slot worn sprites land on the player mirror, indexed by
// equip slot.
func TestApplyAppearanceWornEquipment(t *testing.T) {
	w := NewWorld()
	var worn [event.NumEquipSlots]int
	worn[event.EquipSlotWeapon] = 16
	worn[event.EquipSlotShield] = 21
	w.Apply(event.OtherPlayerAppearance{
		PlayerIndex: 77,
		Name:        "Wieldy",
		WornSprites: worn,
		WornCount:   12,
		HasWorn:     true,
	})
	rec, ok := w.Players.Get(77)
	if !ok {
		t.Fatal("player 77 not recorded")
	}
	if !rec.HasEquip {
		t.Fatal("HasEquip: got false, want true")
	}
	if rec.EquipBySlot[event.EquipSlotWeapon] != 16 {
		t.Errorf("weapon slot: got %d, want 16", rec.EquipBySlot[event.EquipSlotWeapon])
	}
	if rec.EquipBySlot[event.EquipSlotShield] != 21 {
		t.Errorf("shield slot: got %d, want 21", rec.EquipBySlot[event.EquipSlotShield])
	}
}
