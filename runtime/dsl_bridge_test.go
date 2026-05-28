package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// These tests exercise the DSL→Host views WITHOUT actually
// connecting to OpenRSC. We build a minimal Host with just the
// world-state mirror populated by hand, then run small `.routine`
// snippets through the bridge to assert that reserved entities
// resolve correctly.

// newTestHost builds a Host with no connection, no facts, but a
// populated world-state mirror so DSL views have something to read.
func newTestHost() *Host {
	h := New(Options{Username: "test"})
	// Pre-populate self position + skills so DSL can read them.
	h.world.Self.SetPosition(world.Coord{X: 120, Y: 504})
	var cur, mx, xp [world.NumSkills]int
	for i := range cur {
		cur[i] = 30
		mx[i] = 30
	}
	cur[3] = 50 // Hits = 50
	mx[3] = 50
	cur[10] = 25 // Fishing = 25
	mx[10] = 25
	h.world.Self.SetAllSkills(cur, mx, xp, 0)
	h.world.Self.SetFatigue(15)
	// Populate inventory with two items.
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 542, Amount: 1, Wielded: true}, // wielded item
		{ItemID: 373, Amount: 5},                 // 5 lobsters
	})
	return h
}

// runRoutine compiles src and runs it on the test host's
// interpreter. Uses ParseRoutineString so tests can declare their
// routine with any name they want (`routine r() { ... }`,
// `routine fish() { ... }`, etc.) without having to match the
// helper's file name. For the disk-loader filename-match path,
// use TestParseRoutineFileEnforcesNameMatch / runRoutineFromFile.
func runRoutine(t *testing.T, h *Host, src string) interp.Result {
	t.Helper()
	rf, err := ParseRoutineString("<test>", src)
	if err != nil {
		t.Fatalf("ParseRoutineString: %v", err)
	}
	it := h.NewRoutineInterpreter(context.Background())
	return it.RunRoutine(context.Background(), rf.File, nil)
}

func TestSelfPositionVisibleToRoutine(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.position.x }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 120 {
		t.Errorf("got %v, want Int(120)", res.Value)
	}
}

func TestSelfHpReadable(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.hp }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 50 {
		t.Errorf("got %v, want Int(50)", res.Value)
	}
}

func TestSelfEquippedList(t *testing.T) {
	h := newTestHost()
	// Test host wields item 542 in slot 0; only that one item is
	// flagged Wielded, so self.equipped is a 1-element list.
	res := runRoutine(t, h, `routine r() { return self.equipped.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("equipped.length: got %v, want Int(1)", res.Value)
	}
}

func TestSelfEquippedItemId(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.equipped.first.id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 542 {
		t.Errorf("equipped[0].id: got %v, want Int(542)", res.Value)
	}
}

func TestSelfHpFraction(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.hp_fraction }`)
	if f, ok := res.Value.(interp.Float); !ok || float64(f) != 1.0 {
		t.Errorf("hp_fraction: got %v, want Float(1.0)", res.Value)
	}
}

func TestSelfQuestPoints(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.quest_points }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Errorf("quest_points: got %v, want Int(0)", res.Value)
	}
}

func TestSelfIsBusyStub(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.is_busy }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("is_busy stub: got %v, want false", res.Value)
	}
}

func TestSelfFatigue(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.fatigue }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 15 {
		t.Errorf("got %v, want Int(15)", res.Value)
	}
}

func TestSkillsLevelLookup(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.skills.fishing.level }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 25 {
		t.Errorf("got %v, want Int(25)", res.Value)
	}
}

func TestSkillsMaxLevel(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.skills.fishing.max_level }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 25 {
		t.Errorf("max_level: got %v, want Int(25)", res.Value)
	}
}

func TestSkillsXPToNext(t *testing.T) {
	h := newTestHost()
	// Test host has xp=0 for fishing (level 25 max). The xp
	// threshold for level 26 is 8740. So xp_to_next_level = 8740.
	res := runRoutine(t, h, `routine r() { return self.skills.fishing.xp_to_next_level }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 8740 {
		t.Errorf("xp_to_next_level: got %v, want Int(8740)", res.Value)
	}
}

func TestSkillsPercentToNext(t *testing.T) {
	h := newTestHost()
	// xp=0, max_level=25 → percent_to_next_level = 0.
	res := runRoutine(t, h, `routine r() { return self.skills.fishing.percent_to_next_level }`)
	if f, ok := res.Value.(interp.Float); !ok || float64(f) != 0.0 {
		t.Errorf("percent_to_next: got %v, want Float(0)", res.Value)
	}
}

func TestSkillsUnknownNameErrors(t *testing.T) {
	// Routine accessing a nonexistent skill name is a programming
	// bug (probably a typo); surface it as ResultErrored rather
	// than silently returning null. Routines can use the explicit
	// 18-skill list if they need to be defensive.
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.skills.spellcraft }`)
	if res.Kind != interp.ResultErrored {
		t.Errorf("unknown skill: got kind %v, want errored (value=%v)", res.Kind, res.Value)
	}
}

func TestInventoryFree(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.free }`)
	// Inventory has 30 slots in RSC, 2 used -> 28 free.
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 28 {
		t.Errorf("got %v, want Int(28) (30 - 2 used)", res.Value)
	}
}

func TestInventoryIsFull(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.is_full }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("is_full: got %v, want false (test host has only 2/30 slots used)", res.Value)
	}
}

func TestInventoryFindReturnsItemView(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.find(373).id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 373 {
		t.Errorf("find(373).id: got %v, want Int(373)", res.Value)
	}
}

func TestInventoryFindMissingReturnsNull(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.find(99999) }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("find(missing): got %v, want Null", res.Value)
	}
}

func TestInventorySlotOf(t *testing.T) {
	h := newTestHost()
	// Slot 0 is the wielded item (542); slot 1 is the 5 lobsters (373).
	res := runRoutine(t, h, `routine r() { return inventory.slot_of(373) }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("slot_of(373): got %v, want Int(1)", res.Value)
	}
}

func TestInventorySlotOfMissingReturnsNull(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.slot_of(99999) }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("slot_of(missing): got %v, want Null", res.Value)
	}
}

func TestInventoryHasItem(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.has(373) }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("got %v, want Bool(true)", res.Value)
	}
}

func TestInventoryHasMissingItem(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.has(999) }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("got %v, want Bool(false)", res.Value)
	}
}

func TestInventoryCount(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.count(373) }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 5 {
		t.Errorf("got %v, want Int(5)", res.Value)
	}
}

func TestWorldLastChatNullWhenEmpty(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return world.last_chat }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("expected Null for empty buffer, got %v", res.Value)
	}
}

func TestWorldLastChatPopulatedByEvent(t *testing.T) {
	h := newTestHost()
	// Apply a ChatReceived event so the world.Recent buffer fills.
	h.world.Apply(event.NewChatReceived(event.MessageChat, "delores", "hi there", ""))
	res := runRoutine(t, h, `routine r() { return world.last_chat.speaker }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "delores" {
		t.Errorf("speaker: got %v, want String(\"delores\")", res.Value)
	}
}

func TestWorldLastPMPopulated(t *testing.T) {
	h := newTestHost()
	h.world.Apply(event.PrivateMessage{Sender: "alex", Message: "hello"})
	res := runRoutine(t, h, `routine r() { return world.last_pm.message }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "hello" {
		t.Errorf("message: got %v, want String(\"hello\")", res.Value)
	}
}

func TestWorldLastServerMessagePopulated(t *testing.T) {
	h := newTestHost()
	h.world.Apply(event.SystemMessage{Message: "You can't go through this door."})
	res := runRoutine(t, h, `routine r() { return world.last_server_message.message }`)
	if s, ok := res.Value.(interp.String); !ok || !strings.Contains(string(s), "door") {
		t.Errorf("message: got %v, want containing 'door'", res.Value)
	}
}

func TestWorldLastDialogTextPopulated(t *testing.T) {
	h := newTestHost()
	h.world.Apply(event.NpcDialogText{Text: "Greetings, traveller."})
	res := runRoutine(t, h, `routine r() { return world.last_dialog_text.text }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "Greetings, traveller." {
		t.Errorf("dialog text: got %v", res.Value)
	}
}

func TestNpcViewIsAttackableStubWithoutFacts(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 1, X: 100, Y: 200, TypeID: 3})
	res := runRoutine(t, h, `routine r() { return world.npcs[0].is_attackable }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("is_attackable without facts: got %v, want false", res.Value)
	}
}

func TestNpcViewCombatLevelNullWithoutFacts(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 1, X: 100, Y: 200, TypeID: 3})
	res := runRoutine(t, h, `routine r() { return world.npcs[0].combat_level }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("combat_level without facts: got %v, want Null", res.Value)
	}
}

func TestPlayerViewIsFriendStub(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetPosition(1, 100, 200)
	h.world.Players.SetName(1, "alex")
	res := runRoutine(t, h, `routine r() { return world.players[0].is_friend }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("is_friend stub: got %v, want false", res.Value)
	}
}

func TestGroundItemViewIsMineStub(t *testing.T) {
	h := newTestHost()
	h.world.GroundItems.Add(120, 504, 373)
	res := runRoutine(t, h, `routine r() { return world.ground_items[0].is_mine }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("is_mine stub: got %v, want false", res.Value)
	}
}

func TestWorldNpcsReturnsList(t *testing.T) {
	h := newTestHost()
	// No NPCs in the test world.
	res := runRoutine(t, h, `routine r() { return world.npcs.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Errorf("got %v, want Int(0)", res.Value)
	}
}

func TestRequirePreconditionPasses(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `
		routine r() {
			require {
				self.hp > 0
				inventory.free > 0
			}
			return "ok"
		}
	`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "ok" {
		t.Errorf("value: got %v, want \"ok\"", res.Value)
	}
}

func TestRequirePreconditionFails(t *testing.T) {
	h := newTestHost()
	// 0 < self.hp, so this fails.
	res := runRoutine(t, h, `
		routine r() {
			require { self.hp > 999 }
			return "ok"
		}
	`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
}

// TestStdlibStubsRouteToStrategist — contemplate_reality used to
// return NOT_IMPLEMENTED; now it routes through Host.Strategist
// (defaulting to brain.StubStrategist) and returns a canned
// String on CallResult.val. Updated when #74 wired the stubs.
//
// Pick a builtin still genuinely stubbed for the original
// NOT_IMPLEMENTED check — `improvise` and `reflect_now` still
// route through makeStub.
func TestStdlibStubsRouteToStrategist(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return contemplate_reality("hmm") }`)
	cr, ok := res.Value.(*interp.CallResult)
	if !ok {
		t.Fatalf("got %T (%v), want *CallResult", res.Value, res.Value)
	}
	if cr.Err != nil {
		t.Fatalf("expected CallResult.Err to be nil after #74; got %v", cr.Err)
	}
	if _, ok := cr.Val.(interp.String); !ok {
		t.Errorf("val: got %T (%v), want String (brain stub returns a Choice string)", cr.Val, cr.Val)
	}
}

func TestStillStubbedActionReturnsNotImplemented(t *testing.T) {
	// improvise is still routed through makeStub — keeps the
	// NOT_IMPLEMENTED contract under test for genuinely-stubbed
	// surfaces.
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return improvise("anything") }`)
	cr, ok := res.Value.(*interp.CallResult)
	if !ok {
		t.Fatalf("got %T (%v), want *CallResult", res.Value, res.Value)
	}
	if cr.Err == nil {
		t.Fatal("expected CallResult.Err to be set for unstubbed builtin")
	}
	if cr.Err.Code != interp.NOT_IMPLEMENTED {
		t.Errorf("err.code: got %v, want NOT_IMPLEMENTED", cr.Err.Code)
	}
}

func TestStillStubbedBangAborts(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { improvise!("anything"); return "unreached" }`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted (bang on unstubbed should abort)", res.Kind)
	}
	e, ok := res.Value.(*interp.Error)
	if !ok {
		t.Fatalf("aborted value: got %T (%v), want *Error", res.Value, res.Value)
	}
	if e.Code != interp.NOT_IMPLEMENTED {
		t.Errorf("err.code: got %v, want NOT_IMPLEMENTED", e.Code)
	}
}

func TestErrFieldAccessFromDSL(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `
		routine r() {
			result = improvise("anything")
			if result.err {
				return result.err.code
			}
			return "no_err"
		}
	`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "NOT_IMPLEMENTED" {
		t.Errorf("got %v, want String(\"NOT_IMPLEMENTED\")", res.Value)
	}
}

func TestParseRoutineFileSurfacesError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.routine")
	if err := os.WriteFile(path, []byte(`routine bad() { x = nonexistent }`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseRoutineFile(path)
	if err == nil {
		t.Fatal("expected validation error for unbound identifier")
	}
}

func TestParseRoutineFileEnforcesNameMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fish_at_swamp.routine")
	if err := os.WriteFile(path, []byte(`routine completely_different() { return "ok" }`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseRoutineFile(path)
	if err == nil {
		t.Fatal("expected error for filename/routine name mismatch")
	}
	if !strings.Contains(err.Error(), `must match declared routine name`) {
		t.Errorf("error message should mention name match; got %q", err.Error())
	}
}

func TestParseRoutineFileAcceptsMatchingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fish_at_swamp.routine")
	if err := os.WriteFile(path, []byte(`routine fish_at_swamp() { return "ok" }`), 0o644); err != nil {
		t.Fatal(err)
	}
	rf, err := ParseRoutineFile(path)
	if err != nil {
		t.Fatalf("unexpected error for matching name: %v", err)
	}
	if rf.File.Routine.Name != "fish_at_swamp" {
		t.Errorf("routine name: got %q, want fish_at_swamp", rf.File.Routine.Name)
	}
	if rf.Path != path {
		t.Errorf("Path: got %q, want %q", rf.Path, path)
	}
}

func TestParseRoutineStringAcceptsAnyName(t *testing.T) {
	// String loader doesn't enforce filename match — caller picks
	// the logical name.
	rf, err := ParseRoutineString("<repl-line-5>", `routine r() { return 42 }`)
	if err != nil {
		t.Fatalf("ParseRoutineString: %v", err)
	}
	if rf.Path != "<repl-line-5>" {
		t.Errorf("Path: got %q, want <repl-line-5>", rf.Path)
	}
	if rf.File.Routine.Name != "r" {
		t.Errorf("routine name: got %q, want r", rf.File.Routine.Name)
	}
}

func TestParseRoutineStringRejectsEmptyLogicalName(t *testing.T) {
	_, err := ParseRoutineString("", `routine r() { return 1 }`)
	if err == nil {
		t.Fatal("expected error for empty logical name")
	}
}

func TestParseRoutineStringSurfacesParseError(t *testing.T) {
	_, err := ParseRoutineString("<test>", `routine r() { x = }`)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "<test>") {
		t.Errorf("error should include logical name; got %q", err.Error())
	}
}

func TestParseRoutineStringRequiresRoutineDecl(t *testing.T) {
	// Only procs + handlers, no routine decl — error.
	_, err := ParseRoutineString("<test>", `proc helper() { return 1 }`)
	if err == nil {
		t.Fatal("expected error for missing routine declaration")
	}
}
