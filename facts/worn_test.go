package facts

import (
	"os"
	"testing"
)

// TestResolveWorn checks the worn-equipment reverse index against worn
// values captured live from players on the running server (the appearance
// packet carries item AppearanceID & 0xFF per slot). EquipSlot numbering:
// 3=shield, 4=weapon, 5=hat/helmet.
func TestResolveWorn(t *testing.T) {
	if _, err := os.Stat(openrscRoot); err != nil {
		t.Skipf("openrsc data not present at %s; skipping", openrscRoot)
	}
	f, err := Load(DefaultSources(openrscRoot))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// (slot, wornValue, wantName) — captured from live players (alex's
	// dragon kit and Stubbs's tutorial gear).
	cases := []struct {
		slot, val int
		want      string
	}{
		{5, 179, "Dragon medium Helmet"}, // hat
		{5, 70, "Medium Bronze Helmet"},  // hat
		{4, 162, "Dragon axe"},           // weapon
		{3, 106, "Wooden Shield"},        // shield
	}
	for _, c := range cases {
		cands := f.ResolveWorn(c.slot, c.val)
		if len(cands) == 0 {
			t.Errorf("ResolveWorn(slot=%d, val=%d): no candidates, want %q", c.slot, c.val, c.want)
			continue
		}
		found := false
		for _, d := range cands {
			if d.Name == c.want {
				found = true
				break
			}
		}
		if !found {
			names := make([]string, len(cands))
			for i, d := range cands {
				names[i] = d.Name
			}
			t.Errorf("ResolveWorn(slot=%d, val=%d) = %v, want to contain %q", c.slot, c.val, names, c.want)
		}
	}

	// Empty slot and unknown slot resolve to nothing.
	if got := f.ResolveWorn(5, 0); got != nil {
		t.Errorf("ResolveWorn empty slot = %v, want nil", got)
	}

	// Same-metal melee weapons share a worn appearance: value 48 in the
	// weapon slot is every bronze sword/dagger, so >1 candidate is correct.
	if cands := f.ResolveWorn(4, 48); len(cands) < 2 {
		t.Errorf("ResolveWorn(weapon, 48) = %d candidates, want several (bronze weapon family)", len(cands))
	}

	// Bidirectional id<->name: name->def->id, and id->def->name round-trips.
	d := f.ItemDefByName("Dragon medium Helmet")
	if d == nil || d.ID != 795 {
		t.Fatalf("ItemDefByName(Dragon medium Helmet) = %v, want id 795", d)
	}
	if back := f.ItemDef(d.ID); back == nil || back.Name != "Dragon medium Helmet" {
		t.Errorf("ItemDef(%d) round-trip = %v, want name Dragon medium Helmet", d.ID, back)
	}
	// Case-insensitive + nil on unknown.
	if f.ItemDefByName("bRoNzE lOnG sWoRd") == nil {
		t.Errorf("ItemDefByName should be case-insensitive")
	}
	if f.ItemDefByName("no such item zzz") != nil {
		t.Errorf("ItemDefByName(unknown) should be nil")
	}
}
