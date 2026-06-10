package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Combat-namespace perception tests (#117). These build a test host
// (no connection) and hand-populate world.Players / world.Npcs the way
// the opcode-104 / opcode-234 decoders would, then assert the combat.*
// views resolve. SelfPlayerIndex (0) is the local player's own slot.

// TestCombatTargetNullWhenIdle: with no engagement and no last-attack,
// combat.target is Null and combat.engaged is false.
func TestCombatTargetNullWhenIdle(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return combat.target == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("combat.target idle: got %v, want true (null)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("combat.engaged idle: got %v, want false", res.Value)
	}
}

// TestCombatTargetResolvesEngagedNpc: when our own slot (index 0) is
// engaged with an NPC, combat.target resolves to that NPC and
// combat.engaged is true. The NPC's health rides on its own record.
func TestCombatTargetResolvesEngagedNpc(t *testing.T) {
	h := newTestHost()
	// NPC index 7 is in view with hp 12/22.
	h.world.Npcs.Set(world.NpcRecord{Index: 7, TypeID: 3, X: 121, Y: 504})
	h.world.Npcs.SetHits(7, 5, 12, 22)
	// Our own slot engages NPC 7 (projectile type-3): npc victim 7,
	// player side -1.
	h.world.Players.SetEngagement(world.SelfPlayerIndex, 1, 7, -1)

	res := runRoutine(t, h, `routine r() { return combat.target.index }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 7 {
		t.Errorf("combat.target.index: got %v, want Int(7)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("combat.engaged: got %v, want true", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return combat.target.hp_fraction }`)
	if f, ok := res.Value.(interp.Float); !ok || float64(f) != 12.0/22.0 {
		t.Errorf("combat.target.hp_fraction: got %v, want 12/22", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return combat.target.health }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 12 {
		t.Errorf("combat.target.health: got %v, want Int(12)", res.Value)
	}
}

// TestCombatEngagedWhenAttackedByPlayer: a defender with no outgoing
// attack is still "engaged" if another player is firing projectiles at
// our slot (their EngagedPlayerIndex == SelfPlayerIndex).
func TestCombatEngagedWhenAttackedByPlayer(t *testing.T) {
	h := newTestHost()
	// Player 9 engages us (player victim == SelfPlayerIndex 0).
	h.world.Players.SetEngagement(9, 2, -1, world.SelfPlayerIndex)
	res := runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("combat.engaged (attacked): got %v, want true", res.Value)
	}
}

// TestCombatTargetMeleeFallback: authentic melee carries no projectile,
// so our own engagement fields stay empty; combat.target falls back to
// the entity we last issued attack() against while it is still visible.
func TestCombatTargetMeleeFallback(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 4, TypeID: 1, X: 120, Y: 505})
	h.lastAttackedNpcIndex.Store(4)
	res := runRoutine(t, h, `routine r() { return combat.target.index }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 4 {
		t.Errorf("combat.target melee fallback: got %v, want Int(4)", res.Value)
	}
}

// TestCombatStyleReadDefault: combat.style defaults to "controlled"
// before any set_style.
func TestCombatStyleReadDefault(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return combat.style }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "controlled" {
		t.Errorf("combat.style default: got %v, want \"controlled\"", res.Value)
	}
}

// TestCombatStyleWriteThroughMirror: setting the host's combatStyle
// field (the write-through done by dslSetCombatStyle) is reflected by
// the read-side combat.style view.
func TestCombatStyleWriteThroughMirror(t *testing.T) {
	h := newTestHost()
	h.combatStyle = 1 // aggressive
	res := runRoutine(t, h, `routine r() { return combat.style }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "aggressive" {
		t.Errorf("combat.style after set: got %v, want \"aggressive\"", res.Value)
	}
}

// TestNpcHpFractionDestubbed: an NPC with no observed hits reports null
// hp_fraction; once a hits update lands it reports the live fraction.
func TestNpcHpFractionDestubbed(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 2, TypeID: 1, X: 120, Y: 504})
	// Before hits: null.
	res := runRoutine(t, h, `routine r() { return world.npcs.first.hp_fraction == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("npc.hp_fraction pre-hits: got %v, want true (null)", res.Value)
	}
	// After hits: 8/40 = 0.2.
	h.world.Npcs.SetHits(2, 3, 8, 40)
	res = runRoutine(t, h, `routine r() { return world.npcs.first.hp_fraction }`)
	if f, ok := res.Value.(interp.Float); !ok || float64(f) != 0.2 {
		t.Errorf("npc.hp_fraction post-hits: got %v, want 0.2", res.Value)
	}
}
