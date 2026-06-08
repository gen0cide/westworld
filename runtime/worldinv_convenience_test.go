package runtime

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// Tests for the #117 convenience reads:
//   world.ground_items.nearest(pos?)  — dual-mode bare field / callable
//   world.ground_items.most_valuable  — value-sorted selector
//   inventory.find_any([id,...])      — first slot matching ANY ref
//
// These exercise runtime dispatch through the interpreter, mirroring
// namespace_rename_test.go's RunRoutine style.

// testFactsWithItems builds a minimal Facts carrying just the named
// item defs so value-sorted / name-resolving reads have something to
// resolve against. Only ItemDefs is populated — enough for these tests.
func testFactsWithItems(defs ...*facts.ItemDef) *facts.Facts {
	m := make(map[int]*facts.ItemDef, len(defs))
	for _, d := range defs {
		m[d.ID] = d
	}
	return &facts.Facts{ItemDefs: m}
}

// --- world.ground_items.nearest ---

func TestGroundItemsNearestBareFieldNullWhenEmpty(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return world.ground_items.nearest == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("ground_items.nearest (none visible): got %v, want Bool(true)", res.Value)
	}
}

func TestGroundItemsNearestBareFieldClosestToSelf(t *testing.T) {
	h := newTestHost()                     // self at (120, 504)
	h.world.GroundItems.Add(120, 510, 100) // 6 away
	h.world.GroundItems.Add(122, 504, 200) // 2 away — nearest
	res := runRoutine(t, h, `routine r() { return world.ground_items.nearest.id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 200 {
		t.Errorf("ground_items.nearest.id: got %v, want Int(200)", res.Value)
	}
}

func TestGroundItemsNearestCalledWithPosition(t *testing.T) {
	h := newTestHost() // self at (120, 504)
	h.world.GroundItems.Add(120, 510, 100)
	h.world.GroundItems.Add(122, 504, 200)
	// Recenter on (120, 511): item 100 (at 120,510) is now nearest.
	res := runRoutine(t, h, `routine r() { return world.ground_items.nearest(120, 511).id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 100 {
		t.Errorf("ground_items.nearest(120, 511).id: got %v, want Int(100)", res.Value)
	}
}

func TestGroundItemsNearestCalledWithPositionView(t *testing.T) {
	h := newTestHost() // self at (120, 504)
	h.world.GroundItems.Add(120, 510, 100)
	h.world.GroundItems.Add(122, 504, 200) // 2 away from self — nearest to self.position
	// Passing self.position (a Getter with .x/.y) exercises the
	// single-view-arg branch of resolvePoint.
	res := runRoutine(t, h, `routine r() { return world.ground_items.nearest(self.position).id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 200 {
		t.Errorf("ground_items.nearest(self.position).id: got %v, want Int(200)", res.Value)
	}
}

func TestGroundItemsNearestCalledNoArgsMatchesBareField(t *testing.T) {
	h := newTestHost()
	h.world.GroundItems.Add(120, 510, 100)
	h.world.GroundItems.Add(122, 504, 200)
	res := runRoutine(t, h, `routine r() { return world.ground_items.nearest().id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 200 {
		t.Errorf("ground_items.nearest().id: got %v, want Int(200)", res.Value)
	}
}

// --- world.ground_items.most_valuable ---

func TestGroundItemsMostValuableNullWhenEmpty(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return world.ground_items.most_valuable == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("ground_items.most_valuable (none visible): got %v, want Bool(true)", res.Value)
	}
}

func TestGroundItemsMostValuableByBasePrice(t *testing.T) {
	h := newTestHost()
	h.facts = testFactsWithItems(
		&facts.ItemDef{ID: 100, Name: "Bones", BasePrice: 1},
		&facts.ItemDef{ID: 200, Name: "Coins", BasePrice: 0},
		&facts.ItemDef{ID: 300, Name: "Rune Plate", BasePrice: 65000},
	)
	h.world.GroundItems.Add(120, 504, 100)
	h.world.GroundItems.Add(121, 504, 200)
	h.world.GroundItems.Add(122, 504, 300) // most valuable
	res := runRoutine(t, h, `routine r() { return world.ground_items.most_valuable.id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 300 {
		t.Errorf("ground_items.most_valuable.id: got %v, want Int(300)", res.Value)
	}
}

// --- inventory.find_any ---

func TestInventoryFindAnyListArgMatchesFirstSlot(t *testing.T) {
	h := newTestHost()
	// Test host inventory: slot 0 = id 542 (wielded), slot 1 = id 373.
	res := runRoutine(t, h, `routine r() { return inventory.find_any([373, 542]).id }`)
	// Slot order wins: slot 0 (542) is earlier than slot 1 (373).
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 542 {
		t.Errorf("inventory.find_any([373,542]).id: got %v, want Int(542)", res.Value)
	}
}

func TestInventoryFindAnyVarargs(t *testing.T) {
	h := newTestHost()
	// Only 373 is present (542 also present but exclude it); first
	// matching slot for the set {373} is slot 1.
	res := runRoutine(t, h, `routine r() { return inventory.find_any(999, 373).id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 373 {
		t.Errorf("inventory.find_any(999, 373).id: got %v, want Int(373)", res.Value)
	}
}

func TestInventoryFindAnyNoMatchIsNull(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.find_any([1, 2, 3]) == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Errorf("inventory.find_any (no match): got %v, want Bool(true)", res.Value)
	}
}

func TestInventoryFindAnyByName(t *testing.T) {
	h := newTestHost()
	h.facts = testFactsWithItems(
		&facts.ItemDef{ID: 542, Name: "Rune Pickaxe"},
		&facts.ItemDef{ID: 373, Name: "Lobster"},
	)
	res := runRoutine(t, h, `routine r() { return inventory.find_any(["Lobster"]).id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 373 {
		t.Errorf("inventory.find_any([\"Lobster\"]).id: got %v, want Int(373)", res.Value)
	}
}

// Guard: an empty find_any is an authoring error, not a silent Null.
func TestInventoryFindAnyEmptyErrors(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.find_any([]) }`)
	if res.Err == nil {
		t.Errorf("inventory.find_any([]) should error, got value %v", res.Value)
	}
}

// resolveSlot now accepts an item NAME (string), not just a slot int or an
// item-view — so the manual's eat("cookedmeat") / equip("bronze short sword")
// forms work, mirroring use()'s resolveItemID resolution. (Test host inventory:
// slot 0 = id 542, slot 1 = id 373.)
func TestResolveSlotByName(t *testing.T) {
	h := newTestHost()
	h.facts = testFactsWithItems(
		&facts.ItemDef{ID: 542, Name: "Bronze Pickaxe"},
		&facts.ItemDef{ID: 373, Name: "cookedmeat"},
		&facts.ItemDef{ID: 999, Name: "Rune Plate Mail"}, // a real item NOT held
	)
	str := func(s string) []interp.Value { return []interp.Value{interp.String(s)} }

	// Exact name → the slot holding that item.
	if slot, err := resolveSlot(h, str("cookedmeat"), nil); err != nil || slot != 1 {
		t.Errorf("resolveSlot(\"cookedmeat\"): got slot=%d err=%v, want slot=1", slot, err)
	}
	// Substring fallback (like resolveItemID / use): "pickaxe" → "Bronze Pickaxe" → slot 0.
	if slot, err := resolveSlot(h, str("pickaxe"), nil); err != nil || slot != 0 {
		t.Errorf("resolveSlot(\"pickaxe\"): got slot=%d err=%v, want slot=0", slot, err)
	}
	// A real item not in inventory → graceful "not in inventory" (callers map to NO_SUCH_ITEM).
	if _, err := resolveSlot(h, str("rune plate mail"), nil); err == nil || !strings.Contains(err.Error(), "not in inventory") {
		t.Errorf("resolveSlot(<real item not held>): want 'not in inventory', got %v", err)
	}
	// An unknown name is also surfaced as "not in inventory", never a hard error.
	if _, err := resolveSlot(h, str("totally-not-an-item"), nil); err == nil || !strings.Contains(err.Error(), "not in inventory") {
		t.Errorf("resolveSlot(<unknown name>): want 'not in inventory', got %v", err)
	}
	// A bare int is still an explicit slot index (unchanged).
	if slot, err := resolveSlot(h, []interp.Value{interp.Int(2)}, nil); err != nil || slot != 2 {
		t.Errorf("resolveSlot(int 2): got slot=%d err=%v, want slot=2", slot, err)
	}
}

// Sanity: confirm the test world's slot layout assumptions hold, so the
// find_any expectations above stay meaningful if newTestHost changes.
func TestWorldInvConvenienceTestHostInvariants(t *testing.T) {
	h := newTestHost()
	slots := h.world.Inventory.Slots()
	if len(slots) < 2 || slots[0].ItemID != 542 || slots[1].ItemID != 373 {
		t.Fatalf("test host inventory layout changed: %+v", slots)
	}
	_ = world.InvSlot{} // keep the world import load-bearing
}
