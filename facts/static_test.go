package facts

import (
	"os"
	"reflect"
	"testing"
)

// TestStaticParity proves the checked-in generated data (LoadStatic) is
// byte-identical to what the legacy file-parsing loader produces from the
// live OpenRSC tree. Skips when the tree isn't present (CI), so the gate
// only fires on machines that can actually regenerate.
func TestStaticParity(t *testing.T) {
	src := DefaultSources(defaultOpenRSCRoot())
	if _, err := os.Stat(src.ItemDefsJSON); err != nil {
		t.Skipf("openrsc tree not present (%v); parity not checkable here", err)
	}
	live, err := Load(src)
	if err != nil {
		t.Fatalf("legacy load: %v", err)
	}
	gen := LoadStatic()

	if !reflect.DeepEqual(live.SceneryDefs, gen.SceneryDefs) {
		t.Error("SceneryDefs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.BoundaryDefs, gen.BoundaryDefs) {
		t.Error("BoundaryDefs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.NpcDefs, gen.NpcDefs) {
		t.Error("NpcDefs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.ItemDefs, gen.ItemDefs) {
		t.Error("ItemDefs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.TileDefs, gen.TileDefs) {
		t.Error("TileDefs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.SceneryLocs, gen.SceneryLocs) {
		t.Error("SceneryLocs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.BoundaryLocs, gen.BoundaryLocs) {
		t.Error("BoundaryLocs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.NpcLocs, gen.NpcLocs) {
		t.Error("NpcLocs diverge — rerun: go generate ./facts")
	}
	if !reflect.DeepEqual(live.GroundItemLocs, gen.GroundItemLocs) {
		t.Error("GroundItemLocs diverge — rerun: go generate ./facts")
	}
}

func defaultOpenRSCRoot() string {
	if v := os.Getenv("OPENRSC_ROOT"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return home + "/Code/openrsc"
}

// TestStaticSanity guards the generated tables' gross shape with no external
// tree needed — a regen that silently parsed nothing fails here in CI.
func TestStaticSanity(t *testing.T) {
	f := LoadStatic()
	checks := []struct {
		name string
		n    int
		min  int
	}{
		{"SceneryDefs", len(f.SceneryDefs), 1000},
		{"BoundaryDefs", len(f.BoundaryDefs), 200},
		{"NpcDefs", len(f.NpcDefs), 700},
		{"ItemDefs", len(f.ItemDefs), 1200},
		{"TileDefs", len(f.TileDefs), 20},
		{"SceneryLocs", len(f.SceneryLocs), 20000},
		{"BoundaryLocs", len(f.BoundaryLocs), 800},
		{"NpcLocs", len(f.NpcLocs), 3000},
		{"GroundItemLocs", len(f.GroundItemLocs), 1000},
		{"ItemIDByName", len(ItemIDByName), 1200},
		{"NpcIDByName", len(NpcIDByName), 700},
		{"SceneryIDByName", len(SceneryIDByName), 1000},
		{"BoundaryIDByName", len(BoundaryIDByName), 200},
		{"QuestIDByName", len(QuestIDByName), 30},
		{"SkillIDByName", len(SkillIDByName), 18},
		{"SpellOrdinalByName", len(SpellOrdinalByName), 40},
		{"AppearanceSlotByName", len(AppearanceSlotByName), 10},
	}
	for _, c := range checks {
		if c.n < c.min {
			t.Errorf("%s: %d entries, want >= %d", c.name, c.n, c.min)
		}
	}
	// Spot-check the symbol layer against well-known wire ids.
	if got := ItemIDByName["BRONZE_PICKAXE"]; got != 156 {
		t.Errorf("BRONZE_PICKAXE = %d, want 156", got)
	}
	if got := SkillIDByName["MINING"]; got != 14 {
		t.Errorf("MINING = %d, want 14", got)
	}
	if got, ok := SkillIDByName["HITPOINTS"]; !ok || got != 3 {
		t.Errorf("HITPOINTS = %d (ok=%v), want 3", got, ok)
	}
}
