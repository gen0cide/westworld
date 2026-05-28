package runtime

import (
	"context"
	"os"
	"path/filepath"
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
// interpreter. Returns the result and any parse/validate error.
func runRoutine(t *testing.T, h *Host, src string) interp.Result {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "t.routine")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	res, err := h.RunRoutine(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("RunRoutine: %v", err)
	}
	return res
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

func TestStdlibStubReturnsErrorString(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return contemplate_reality("hmm") }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "contemplate_reality: not_implemented" {
		t.Errorf("got %v, want stub error string", res.Value)
	}
}

func TestParseRoutineFileSurfacesError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.routine")
	if err := os.WriteFile(path, []byte(`routine r() { x = nonexistent }`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseRoutineFile(path)
	if err == nil {
		t.Fatal("expected validation error for unbound identifier")
	}
}
