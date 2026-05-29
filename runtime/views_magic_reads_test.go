package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Tests for the client-computed magic.* reads (#117): magic.level /
// magic.max_level / magic.known / magic.can_cast / magic.has_runes_for.
// All resolve through magicView without any server round-trip.

// magic.level is the current/boosted Magic level (skill index 6);
// magic.max_level is the base. newTestHost sets all skills to 30.
func TestMagicLevelReads(t *testing.T) {
	h := newTestHost()
	// Boost current magic to 40 while leaving the base at 30.
	h.world.Self.SetSkill(magicSkillID, 40, 30, 0)

	lvl := runRoutine(t, h, `routine r() { return magic.level }`)
	if i, ok := lvl.Value.(interp.Int); !ok || int64(i) != 40 {
		t.Errorf("magic.level: got %v, want Int(40)", lvl.Value)
	}
	mx := runRoutine(t, h, `routine r() { return magic.max_level }`)
	if i, ok := mx.Value.(interp.Int); !ok || int64(i) != 30 {
		t.Errorf("magic.max_level: got %v, want Int(30)", mx.Value)
	}
}

// magic.known lists spells whose req_level <= current/boosted magic
// level. With magic boosted to 13, Fire strike (id 6, req 13) is in;
// raising the threshold past it removes it.
func TestMagicKnownUsesBoostedLevel(t *testing.T) {
	h := newTestHost()
	h.world.Self.SetSkill(magicSkillID, 13, 1, 0) // boosted to 13, base 1

	res := runRoutine(t, h, `routine r() {
		return magic.known.find(s => s.name == "Fire strike")
	}`)
	if _, isNull := res.Value.(interp.Null); isNull || res.Value == nil {
		t.Errorf("magic.known should include Fire strike at boosted lvl 13, got %v", res.Value)
	}

	// At current level 12 Fire strike (req 13) drops out → find returns null.
	h.world.Self.SetSkill(magicSkillID, 12, 1, 0)
	res2 := runRoutine(t, h, `routine r() {
		return magic.known.find(s => s.name == "Fire strike")
	}`)
	if _, isNull := res2.Value.(interp.Null); !isNull {
		t.Errorf("magic.known should exclude Fire strike at lvl 12, got %v", res2.Value)
	}
}

// magic.can_cast is a level-only gate against the current level and is
// null-safe for unknown spells.
func TestMagicCanCast(t *testing.T) {
	h := newTestHost()
	h.world.Self.SetSkill(magicSkillID, 13, 13, 0)

	// Fire strike (req 13) by name → castable at lvl 13.
	yes := runRoutine(t, h, `routine r() { return magic.can_cast("Fire strike") }`)
	if b, ok := yes.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("magic.can_cast(Fire strike) at lvl 13: got %v, want true", yes.Value)
	}

	// Below req level → not castable.
	h.world.Self.SetSkill(magicSkillID, 12, 12, 0)
	no := runRoutine(t, h, `routine r() { return magic.can_cast("Fire strike") }`)
	if b, ok := no.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("magic.can_cast(Fire strike) at lvl 12: got %v, want false", no.Value)
	}

	// Unknown spell name → false, never null.
	unk := runRoutine(t, h, `routine r() { return magic.can_cast("No Such Spell") }`)
	if b, ok := unk.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("magic.can_cast(unknown): got %v, want Bool(false)", unk.Value)
	}

	// can_cast ignores runes: lvl-ok but no runes still returns true.
	h.world.Self.SetSkill(magicSkillID, 13, 13, 0)
	h.world.Inventory.Replace(nil)
	noRunes := runRoutine(t, h, `routine r() { return magic.can_cast("Fire strike") }`)
	if b, ok := noRunes.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("magic.can_cast must be level-only (ignore runes): got %v, want true", noRunes.Value)
	}
}

// magic.has_runes_for checks inventory counts and treats an equipped
// elemental staff as supplying that element. Wind strike (id 0)
// requires mind(35)x1 + air(33)x1.
func TestMagicHasRunesForWithStaff(t *testing.T) {
	h := newTestHost()

	// Full runes in inventory → true.
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 35, Amount: 5}, // mind runes
		{ItemID: 33, Amount: 5}, // air runes
	})
	full := runRoutine(t, h, `routine r() { return magic.has_runes_for("Wind strike") }`)
	if b, ok := full.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("has_runes_for(Wind strike) with both runes: got %v, want true", full.Value)
	}

	// Missing the air rune → false.
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 35, Amount: 5}, // only mind
	})
	missing := runRoutine(t, h, `routine r() { return magic.has_runes_for("Wind strike") }`)
	if b, ok := missing.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("has_runes_for(Wind strike) missing air rune: got %v, want false", missing.Value)
	}

	// Equipped air staff (item 101, wielded) supplies the air element;
	// only mind runes remain needed → true.
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 35, Amount: 5},                 // mind runes
		{ItemID: 101, Amount: 1, Wielded: true}, // staff of air (wielded)
	})
	withStaff := runRoutine(t, h, `routine r() { return magic.has_runes_for("Wind strike") }`)
	if b, ok := withStaff.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("has_runes_for(Wind strike) with equipped air staff + mind runes: got %v, want true", withStaff.Value)
	}

	// Same air staff but NOT wielded → does not count; air rune missing.
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 35, Amount: 5},                  // mind runes
		{ItemID: 101, Amount: 1, Wielded: false}, // staff of air, not wielded
	})
	unwielded := runRoutine(t, h, `routine r() { return magic.has_runes_for("Wind strike") }`)
	if b, ok := unwielded.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("has_runes_for(Wind strike) with UN-wielded air staff: got %v, want false", unwielded.Value)
	}

	// Unknown spell → false.
	unk := runRoutine(t, h, `routine r() { return magic.has_runes_for("No Such Spell") }`)
	if b, ok := unk.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("has_runes_for(unknown): got %v, want false", unk.Value)
	}
}
