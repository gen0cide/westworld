package facts

import (
	"os"
	"testing"
)

// openrscRoot is the path to the running OpenRSC server's defs
// directory. Tests that need real data check for its presence and
// skip if not found, so the package can be built in CI without the
// data.
const openrscRoot = "/Users/flint/Code/openrsc"

func TestLoadRealData(t *testing.T) {
	if _, err := os.Stat(openrscRoot); err != nil {
		t.Skipf("openrsc data not present at %s; skipping", openrscRoot)
	}

	srcs := DefaultSources(openrscRoot)
	f, err := Load(srcs)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	t.Logf("loaded: %s", f.Summary())

	// Sanity checks — exact counts come from the data files.
	if len(f.SceneryDefs) < 100 {
		t.Errorf("scenery defs count looks too low: %d", len(f.SceneryDefs))
	}
	if len(f.SceneryLocs) < 1000 {
		t.Errorf("scenery locs count looks too low: %d", len(f.SceneryLocs))
	}
	if len(f.NpcDefs) < 100 {
		t.Errorf("npc defs count looks too low: %d", len(f.NpcDefs))
	}

	// Lumbridge spawn at (120, 648). Look for "Well" within 30 tiles.
	well := f.NearestByName("Well", 120, 648, 30, "scenery")
	if well == nil {
		t.Logf("no well found within 30 tiles of Lumbridge spawn (may be data variant issue)")
	} else {
		t.Logf("nearest Well: %+v", well)
	}

	// What's at the Lumbridge spawn tile itself?
	atSpawn := f.At(120, 648)
	t.Logf("at Lumbridge spawn (120, 648): %d placements", len(atSpawn))
	for _, p := range atSpawn {
		t.Logf("  %s id=%d %q dir=%d %s", p.Kind, p.DefID, p.Name, p.Direction, p.Extra)
	}

	// NPC spawns near Lumbridge spawn.
	near := f.Near(120, 648, 8)
	npcCount := 0
	for _, p := range near {
		if p.Kind == "npc_spawn" {
			npcCount++
			if npcCount <= 5 {
				t.Logf("  nearby NPC spawn: %s at (%d, %d) — %s", p.Name, p.X, p.Y, p.Extra)
			}
		}
	}
	t.Logf("total NPC spawns within 8 tiles of Lumbridge spawn: %d", npcCount)
}
