package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// Tests for the #117 self namespace additions: self.position.plane and
// the self.equipped per-slot accessors (reading world.Self.EquipBySlot).

func TestSelfPositionPlaneGround(t *testing.T) {
	h := newTestHost() // test host is at Y=504 -> plane 0
	res := runRoutine(t, h, `routine r() { return self.position.plane }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Errorf("position.plane (ground): got %v, want Int(0)", res.Value)
	}
}

func TestSelfPositionPlaneUpper(t *testing.T) {
	h := newTestHost()
	// Y = 1*PlaneHeight + 100 -> plane 1 (matches world.Self.Plane()).
	h.world.Self.SetPosition(world.Coord{X: 50, Y: world.PlaneHeight + 100})
	res := runRoutine(t, h, `routine r() { return self.position.plane }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("position.plane (upper): got %v, want Int(1)", res.Value)
	}
	if got := h.world.Self.Plane(); got != 1 {
		t.Errorf("world.Self.Plane(): got %d, want 1 (accessor must match)", got)
	}
}

func TestSelfEquippedListSurfacePreserved(t *testing.T) {
	h := newTestHost()
	// .all returns a real list usable with .filter; the wielded item
	// (542) is present.
	res := runRoutine(t, h, `routine r() { return self.equipped.all.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("equipped.all.length: got %v, want Int(1)", res.Value)
	}
	// Indexing + iteration still resolve over the wielded list.
	res = runRoutine(t, h, `routine r() { return self.equipped[0].id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 542 {
		t.Errorf("equipped[0].id: got %v, want Int(542)", res.Value)
	}
}

func TestSelfEquippedPerSlotSpriteEmptyByDefault(t *testing.T) {
	h := newTestHost()
	// No appearance update landed for self, so every slot reads 0 /
	// is_empty. .id stays null (no sprite->item reverse map).
	res := runRoutine(t, h, `routine r() { return self.equipped.weapon.sprite_id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Errorf("equipped.weapon.sprite_id (default): got %v, want Int(0)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return self.equipped.weapon.is_empty }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("equipped.weapon.is_empty (default): got %v, want Bool(true)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return self.equipped.weapon.id }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("equipped.weapon.id: got %v, want Null (sprite->item map gap)", res.Value)
	}
}

func TestSelfEquippedPerSlotReadsEquipBySlot(t *testing.T) {
	h := newTestHost()
	// Land self's worn-equipment sprite ids via the additive setter
	// (the future self-appearance landing path).
	var worn [event.NumEquipSlots]int
	worn[event.EquipSlotWeapon] = 16
	worn[event.EquipSlotShield] = 21
	h.world.Self.SetWornEquipment(worn)

	res := runRoutine(t, h, `routine r() { return self.equipped.weapon.sprite_id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 16 {
		t.Errorf("equipped.weapon.sprite_id: got %v, want Int(16)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return self.equipped.shield.sprite_id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 21 {
		t.Errorf("equipped.shield.sprite_id: got %v, want Int(21)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return self.equipped.head.is_empty }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("equipped.head.is_empty: got %v, want Bool(true)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return self.equipped.weapon.slot_name }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "weapon" {
		t.Errorf("equipped.weapon.slot_name: got %v, want String(weapon)", res.Value)
	}
}
