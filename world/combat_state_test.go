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

// TestApplyAppearanceColours checks the four appearance colour indices
// (hair/top/trouser/skin) land on the player record from a type-5 update.
func TestApplyAppearanceColours(t *testing.T) {
	w := NewWorld()
	w.Apply(event.OtherPlayerAppearance{
		PlayerIndex:   42,
		Name:          "Zezima",
		HairColour:    2,
		TopColour:     8,
		TrouserColour: 14,
		SkinColour:    1,
		HasColours:    true,
	})
	rec, ok := w.Players.Get(42)
	if !ok {
		t.Fatal("player 42 not recorded")
	}
	if !rec.HasColours {
		t.Fatal("HasColours: got false, want true")
	}
	if rec.HairColour != 2 || rec.TopColour != 8 || rec.TrouserColour != 14 || rec.SkinColour != 1 {
		t.Errorf("colours: got hair=%d top=%d trouser=%d skin=%d, want 2/8/14/1",
			rec.HairColour, rec.TopColour, rec.TrouserColour, rec.SkinColour)
	}
}

// TestApplyAppearanceSelfMirror checks that a type-5 appearance update for our
// OWN player index (SelfPlayerIndex == 0) also lands the worn equipment +
// colours onto world.Self, so the render path can show bernard in his real
// kit. Other players' updates must NOT touch Self.
func TestApplyAppearanceSelfMirror(t *testing.T) {
	w := NewWorld()
	var worn [event.NumEquipSlots]int
	worn[event.EquipSlotWeapon] = 16
	worn[event.EquipSlotBody] = 9
	w.Apply(event.OtherPlayerAppearance{
		PlayerIndex:   SelfPlayerIndex,
		Name:          "bernard",
		WornSprites:   worn,
		WornCount:     12,
		HasWorn:       true,
		HairColour:    3,
		TopColour:     8,
		TrouserColour: 14,
		SkinColour:    0,
		HasColours:    true,
	})
	if !w.Self.HasEquip() {
		t.Fatal("Self.HasEquip: got false, want true after self appearance update")
	}
	if got := w.Self.EquipSprites(); got[event.EquipSlotWeapon] != 16 || got[event.EquipSlotBody] != 9 {
		t.Errorf("Self equip: got weapon=%d body=%d, want 16/9", got[event.EquipSlotWeapon], got[event.EquipSlotBody])
	}
	hair, top, trouser, skin, ok := w.Self.AppearanceColours()
	if !ok || hair != 3 || top != 8 || trouser != 14 || skin != 0 {
		t.Errorf("Self colours: got ok=%v hair=%d top=%d trouser=%d skin=%d, want true/3/8/14/0",
			ok, hair, top, trouser, skin)
	}

	// An appearance update for ANOTHER player must not clobber Self.
	w2 := NewWorld()
	w2.Apply(event.OtherPlayerAppearance{PlayerIndex: 5, Name: "Other", HasWorn: true, HasColours: true, HairColour: 9})
	if w2.Self.HasEquip() {
		t.Error("Self.HasEquip: got true after a NON-self appearance update, want false")
	}
	if _, _, _, _, ok := w2.Self.AppearanceColours(); ok {
		t.Error("Self colours marked present after a NON-self appearance update")
	}
}

// TestDynamicSceneryRemovalMark checks that the 60000 removal sentinel both
// clears the live record (the by_id accessor contract) AND marks the tile as
// actively removed (so the renderer suppresses static baseline scenery there),
// and that re-adding a live object clears the removal mark.
func TestDynamicSceneryRemovalMark(t *testing.T) {
	s := NewDynamicScenery()
	s.Add(100, 200, 97)
	if s.IsRemoved(100, 200) {
		t.Fatal("freshly added tile must not be marked removed")
	}
	s.Remove(100, 200)
	if _, ok := s.At(100, 200); ok {
		t.Error("At must report the tile gone after Remove")
	}
	if !s.IsRemoved(100, 200) {
		t.Error("IsRemoved must report the tile removed after the sentinel")
	}
	s.Add(100, 200, 11) // a new object appears at the same tile
	if s.IsRemoved(100, 200) {
		t.Error("IsRemoved must clear once a live object re-occupies the tile")
	}
}

// TestNpcRemovePrunesMirror checks that an opcode-79 REMOVE_NPC
// (event.NpcNearby{Removed:true}) prunes the NPC from both the record
// map and the ordered local-NPC list, so combat.target / world.npcs
// stop resolving a dead/despawned NPC. (#combat-prune regression)
func TestNpcRemovePrunesMirror(t *testing.T) {
	w := NewWorld()
	// Two NPCs enter view.
	w.Apply(event.NpcNearby{Index: 100, X: 50, Y: 60, TypeID: 19, IsNew: true})
	w.Apply(event.NpcNearby{Index: 101, X: 51, Y: 60, TypeID: 19, IsNew: true})
	if got := len(w.Npcs.All()); got != 2 {
		t.Fatalf("after 2 spawns: got %d npcs, want 2", got)
	}
	if got := w.Npcs.Order(); len(got) != 2 || got[0] != 100 || got[1] != 101 {
		t.Fatalf("order after spawns: got %v, want [100 101]", got)
	}
	// 100 accumulates combat state, then the server removes it (death).
	w.Apply(event.NpcDamage{NpcIndex: 100, Damage: 4, CurHits: 1, MaxHits: 5})
	w.Apply(event.NpcNearby{Index: 100, Removed: true})
	if _, ok := w.Npcs.Get(100); ok {
		t.Error("npc 100 still present after REMOVE")
	}
	if got := w.Npcs.Order(); len(got) != 1 || got[0] != 101 {
		t.Errorf("order after remove: got %v, want [101]", got)
	}
}

// TestNpcMovePreservesCombatState checks that a position-only movement
// update keeps an already-tracked NPC's type + accumulated hits (the
// movement record carries neither). Regression for the in-combat
// remove+readd churn fix relying on Move/Set carry-forward.
func TestNpcMovePreservesCombatState(t *testing.T) {
	w := NewWorld()
	w.Apply(event.NpcNearby{Index: 200, X: 10, Y: 10, TypeID: 19, IsNew: true})
	w.Apply(event.NpcDamage{NpcIndex: 200, Damage: 2, CurHits: 3, MaxHits: 5})
	// Movement update: a RELATIVE one-tile east step (DX:1) from (10,10) ->
	// (11,10), carrying no type/hits on the wire. opcode-79 movement encodes a
	// direction/delta, not an absolute coord.
	w.Apply(event.NpcNearby{Index: 200, DX: 1, DY: 0, IsNew: false})
	rec, ok := w.Npcs.Get(200)
	if !ok {
		t.Fatal("npc 200 lost after movement update")
	}
	if rec.X != 11 || rec.Y != 10 {
		t.Errorf("position: got (%d,%d), want (11,10)", rec.X, rec.Y)
	}
	if rec.TypeID != 19 {
		t.Errorf("TypeID clobbered by movement: got %d, want 19", rec.TypeID)
	}
	if !rec.HasHits || rec.CurHits != 3 {
		t.Errorf("combat state lost: HasHits=%v CurHits=%d, want true/3", rec.HasHits, rec.CurHits)
	}
}
