package runtime

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// --- Sell affordance (TODO C-cluster; retro cause #6) ------------------------
//
// shop.sell was wired for the project's whole life and NEVER invoked; the hint
// these tests pin down is the prose that makes the affordance visible. The
// contract under test: fires for broke+sellable, silent when rich, silent when
// nothing a shop would buy, item list capped, coins never counted as stock.

// sellTestHost builds the standard test host with the REAL static item
// catalogue (live BasePrice/IsUntradable flags + the real Coins id) and the
// given inventory. Real defs keep the sellability rules honest — fixture defs
// could drift from the catalogue the live hint prices against.
//
// Static-catalogue items used below (id, name, base gp):
//
//	10 Coins(1) · 0 Iron Mace(63) · 1 Iron Short Sword(91)
//	2 Iron Kite Shield(238) · 3 Iron Square Shield(168) · 4 Wooden Shield(20)
//	23 Flour(2, UNTRADABLE) · 24 Amulet of GhostSpeak(35, UNTRADABLE)
func sellTestHost(slots ...world.InvSlot) *Host {
	h := newTestHost()
	h.facts = facts.LoadStaticCatalogs()
	h.world.Inventory.Replace(slots)
	return h
}

// TestSellAffordanceHintBrokeAndSellable proves the core affordance: a broke
// host (12gp) carrying shop-buyable items gets ONE compact line that names the
// items most-valuable-first, groups duplicate unstackable slots, prices the
// lot, teaches shop.sell — and that the line actually reaches the channel the
// model reads (mesad's actPrompt renders a fixed key set, so it rides "scene").
func TestSellAffordanceHintBrokeAndSellable(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 10, Amount: 12}, // 12gp — broke
		world.InvSlot{ItemID: 2, Amount: 1},   // Iron Kite Shield, 238gp
		world.InvSlot{ItemID: 0, Amount: 1},   // Iron Mace ×3, one per slot
		world.InvSlot{ItemID: 0, Amount: 1},
		world.InvSlot{ItemID: 0, Amount: 1},
	)
	d := quietDirector()
	hint := d.sellAffordanceHint(h, 120, 504)
	for _, want := range []string{
		"You hold sellable items",
		"Iron Kite Shield", // most valuable, named first
		"3 Iron Mace",      // three slots grouped into one stack
		"~427gp",           // 238 + 3×63; the coins themselves not counted
		"shop.sell(item)",
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	if strings.Contains(hint, "Coins") {
		t.Fatalf("coins must never be listed as sellable stock:\n%s", hint)
	}
	if strings.Index(hint, "Iron Kite Shield") > strings.Index(hint, "Iron Mace") {
		t.Fatalf("most valuable item should lead the list:\n%s", hint)
	}
	// End-to-end: the situation the planner sees must carry the line.
	sit := d.situation(h, Outcome{})
	if !strings.Contains(sit.Hints["scene"], "shop.sell(item)") {
		t.Fatalf("sell affordance did not reach the scene hint:\n%s", sit.Hints["scene"])
	}
}

// TestSellAffordanceHintAbsentWhenRich proves scarcity: at or above the coin
// floor the hint must not render at all — hints stay scarce to stay load-bearing.
func TestSellAffordanceHintAbsentWhenRich(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 10, Amount: 500}, // rich
		world.InvSlot{ItemID: 2, Amount: 1},    // sellable, but irrelevant now
	)
	d := quietDirector()
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("rich host must get no sell hint, got:\n%s", hint)
	}
	// Boundary: exactly the floor counts as "not broke".
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 10, Amount: sellHintCoinFloor},
		{ItemID: 2, Amount: 1},
	})
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("at the coin floor the hint must stay silent, got:\n%s", hint)
	}
	sit := d.situation(h, Outcome{})
	if strings.Contains(sit.Hints["scene"], "shop.sell") {
		t.Fatalf("scene must not carry a sell hint when rich:\n%s", sit.Hints["scene"])
	}
}

// TestSellAffordanceHintAbsentWhenNothingSellable proves the other scarcity
// edge: untradables (even with a positive BasePrice — the tradeability filter,
// not the price one), an empty inventory, and a catalogue-less host all stay
// silent.
func TestSellAffordanceHintAbsentWhenNothingSellable(t *testing.T) {
	d := quietDirector()
	h := sellTestHost(
		world.InvSlot{ItemID: 23, Amount: 1}, // Flour — untradable, BasePrice 2
		world.InvSlot{ItemID: 24, Amount: 1}, // Amulet of GhostSpeak — untradable
	)
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("untradables must not trigger the hint, got:\n%s", hint)
	}
	h.world.Inventory.Replace(nil)
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("an empty inventory must not trigger the hint, got:\n%s", hint)
	}
	bare := newTestHost() // h.facts == nil: nothing can be priced
	if hint := d.sellAffordanceHint(bare, 120, 504); hint != "" {
		t.Fatalf("nil facts must not trigger the hint, got:\n%s", hint)
	}
}

// TestSellAffordanceHintItemCap proves the 3-item cap: with five sellable
// kinds only the three most valuable are named, the cheapest two fold into
// "and 2 more", and the gp figure still counts everything.
func TestSellAffordanceHintItemCap(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 0, Amount: 1}, // Iron Mace 63 — folded
		world.InvSlot{ItemID: 1, Amount: 1}, // Iron Short Sword 91
		world.InvSlot{ItemID: 2, Amount: 1}, // Iron Kite Shield 238
		world.InvSlot{ItemID: 3, Amount: 1}, // Iron Square Shield 168
		world.InvSlot{ItemID: 4, Amount: 1}, // Wooden Shield 20 — folded
	)
	d := quietDirector()
	hint := d.sellAffordanceHint(h, 120, 504)
	for _, want := range []string{
		"Iron Kite Shield", "Iron Square Shield", "Iron Short Sword",
		"and 2 more",
		"~580gp", // 63+91+238+168+20 — the fold still counts toward the total
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	for _, banned := range []string{"Iron Mace", "Wooden Shield"} {
		if strings.Contains(hint, banned) {
			t.Fatalf("cap exceeded — %q should fold into 'and 2 more':\n%s", banned, hint)
		}
	}
}

// TestSellAffordanceHintNearestShopRider proves the proximity rider uses only
// the embedded gazetteer already in RAM: (120,504) sits ~10 tiles from the
// (130,512) general shop so the rider names it; from (700,300) the nearest is
// hundreds of tiles away — no rider (a far shop is search_map's job).
func TestSellAffordanceHintNearestShopRider(t *testing.T) {
	h := sellTestHost(world.InvSlot{ItemID: 0, Amount: 1}) // broke, one mace
	d := quietDirector()
	if hint := d.sellAffordanceHint(h, 120, 504); !strings.Contains(hint, "Nearest: general shop ~10 tiles") {
		t.Fatalf("expected the nearest-shop rider at (120,504):\n%s", hint)
	}
	if hint := d.sellAffordanceHint(h, 700, 300); strings.Contains(hint, "Nearest:") {
		t.Fatalf("no rider expected far from any shop:\n%s", hint)
	}
}

// TestIgnoranceHintsRender pins the three self-aware-ignorance lines: fog
// coverage (weighted, with frontier directions), never-tried skills, and the
// aspiration portfolio — all riding the scene hint (fixed-key constraint).
func TestIgnoranceHintsRender(t *testing.T) {
	h := newTestHost()
	fake := &fakeFogOracle{dim: 96, dimY: 96, comps: map[[2]int]int32{
		{10, 10}: 0, {60, 10}: 0,
	}}
	h.fog.oracle = fake
	h.fogObservePosition(10, 10)

	d := NewMesaDirector(&fakeAskClient{healthy: true}, "Delores", "g", slog.Default())
	hint := d.explorationHint(h, 10, 10)
	if !strings.Contains(hint, "EXPLORATION") || !strings.Contains(hint, "%") {
		t.Fatalf("exploration hint missing coverage: %q", hint)
	}
	if !strings.Contains(hint, "Unknown lands") {
		t.Fatalf("exploration hint missing frontier: %q", hint)
	}

	sk := d.skillIgnoranceHint(h)
	if sk != "" && !strings.Contains(sk, "never tried") {
		t.Fatalf("skill hint malformed: %q", sk)
	}

	// Aspirations: dark without a portfolio, renders with one.
	if asp := d.aspirationHint(h); asp != "" {
		t.Fatalf("aspiration hint should be dark with no portfolio: %q", asp)
	}
}
