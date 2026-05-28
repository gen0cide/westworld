package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
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

func TestSkillsLookup(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return self.skills.fishing }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 25 {
		t.Errorf("got %v, want Int(25)", res.Value)
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

func TestStdlibStubReturnsTypedError(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return contemplate_reality("hmm") }`)
	cr, ok := res.Value.(*interp.CallResult)
	if !ok {
		t.Fatalf("got %T (%v), want *CallResult", res.Value, res.Value)
	}
	if cr.Err == nil {
		t.Fatal("expected CallResult.Err to be set for stub")
	}
	if cr.Err.Code != interp.NOT_IMPLEMENTED {
		t.Errorf("err.code: got %v, want NOT_IMPLEMENTED", cr.Err.Code)
	}
	if cr.Err.Reason != "contemplate_reality" {
		t.Errorf("err.reason: got %q, want \"contemplate_reality\"", cr.Err.Reason)
	}
}

func TestStdlibStubBangAbortsRoutine(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { contemplate_reality!("hmm"); return "unreached" }`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted (bang on stub should abort)", res.Kind)
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
			result = contemplate_reality("hmm")
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
