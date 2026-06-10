package debughttp

import (
	"encoding/json"
	"testing"
)

// TestEquipSnapItemIDZeroIsExact locks down the /state equipment contract: the
// item_id key's PRESENCE signals an exact resolution, so item id 0 (the Iron
// Mace) must serialize as "item_id":0 — an omitempty int would strip it and
// make a perfectly resolved weapon read as an ambiguous slot. Ambiguous slots
// (nil pointer) omit the key entirely.
func TestEquipSnapItemIDZeroIsExact(t *testing.T) {
	zero := 0
	b, err := json.Marshal(equipSnap{Slot: 4, SlotName: "weapon", Name: "Iron Mace", ItemID: &zero})
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	v, ok := wire["item_id"]
	if !ok {
		t.Fatalf("exact resolution of item id 0 must carry the item_id key: %s", b)
	}
	if f, _ := v.(float64); f != 0 {
		t.Errorf("item_id = %v, want 0: %s", v, b)
	}

	b, err = json.Marshal(equipSnap{Slot: 4, SlotName: "weapon", Name: "a mace (or similar)"})
	if err != nil {
		t.Fatal(err)
	}
	wire = map[string]any{}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	if _, present := wire["item_id"]; present {
		t.Errorf("ambiguous slot must omit item_id: %s", b)
	}
}
