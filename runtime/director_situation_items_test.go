package runtime

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// --- C-37 ground-item perception (the iron-dagger exhibit) -------------------
//
// Ground items were rendered NOWHERE in the Act situation, so no host could be
// opportunistic about the world's free tools (an iron dagger spawned in the
// goblin house and nobody took it; 15 bronze pickaxe spawns stayed invisible
// while hosts failed shop buys). The contract under test: nearby reachable
// items render one tabular-stable line per kind with the never-held cue,
// nothing renders when nothing qualifies, the kind list caps at the nearest
// groundItemLineCap, and the DSL-1 reach gate drops other-component spawns.

// groundItemTestHost builds the standard test host with the REAL static item
// catalogue (live ids/names — fixture defs could drift from what the live
// hint renders) and EMPTY hands, so every kind starts never-held. The host
// stands at (120,504) (newTestHost) with no world oracle, so the reach gate
// is nil and keeps everything — exactly the selectors' oracle-less behavior.
//
// Static-catalogue items used below (id, name):
//
//	156 Bronze Pickaxe · 28 Iron dagger · 0 Iron Mace · 1 Iron Short Sword
//	2 Iron Kite Shield · 3 Iron Square Shield · 4 Wooden Shield
func groundItemTestHost() *Host {
	h := newTestHost()
	h.facts = facts.LoadStaticCatalogs()
	h.world.Inventory.Replace(nil)
	return h
}

// TestGroundItemsHintRendersNearby proves the core perception line: nearby
// reachable spawns render name + distance + direction (RSC bearings: +x =
// west, -y = north), each carries the never-held cue for a kind with no
// first-touch record, the nearest kind leads (tabular-stable ordering), and
// the block actually reaches the channel the model reads (mesad's actPrompt
// renders a fixed key set, so it rides "scene").
func TestGroundItemsHintRendersNearby(t *testing.T) {
	h := groundItemTestHost()
	h.world.GroundItems.Add(128, 504, 156) // Bronze Pickaxe — 8 tiles west
	h.world.GroundItems.Add(120, 501, 28)  // Iron dagger — 3 tiles north
	d := quietDirector()
	hint := d.groundItemsHint(h, 120, 504)
	for _, want := range []string{
		"ON THE GROUND",
		"pick_up(world.ground_items.nearest)",
		"Bronze Pickaxe, 8 tiles west — you have never held one",
		"Iron dagger, 3 tiles north — you have never held one",
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	if strings.Index(hint, "Iron dagger") > strings.Index(hint, "Bronze Pickaxe") {
		t.Fatalf("nearest kind must lead the list:\n%s", hint)
	}
	// End-to-end: the situation the planner sees must carry the block.
	sit := d.situation(h, Outcome{})
	if !strings.Contains(sit.Hints["scene"], "ON THE GROUND") {
		t.Fatalf("ground-item block did not reach the scene hint:\n%s", sit.Hints["scene"])
	}
}

// TestGroundItemsHintNeverHeldCueSuppressed proves both held-evidence paths
// silence the cue while the line itself still renders: a ledger first-touch
// record (perceiveItemGained's heldItemClaim, written on every ItemGained)
// and the current inventory (session-genesis gear never fires ItemGained).
func TestGroundItemsHintNeverHeldCueSuppressed(t *testing.T) {
	h := groundItemTestHost()
	h.world.GroundItems.Add(122, 504, 156) // Bronze Pickaxe
	h.world.GroundItems.Add(120, 502, 28)  // Iron dagger
	h.knowledge.Observe("Bronze Pickaxe", heldItemClaim, true, 1.0)
	h.world.Inventory.Replace([]world.InvSlot{{ItemID: 28, Amount: 1, Wielded: true}})
	d := quietDirector()
	hint := d.groundItemsHint(h, 120, 504)
	for _, want := range []string{"Bronze Pickaxe", "Iron dagger"} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	if strings.Contains(hint, "you have never held one") {
		t.Fatalf("held kinds must not carry the never-held cue:\n%s", hint)
	}
}

// TestGroundItemsHintAbsent proves the empty cases render NOTHING (an empty
// block is silence, not a header): no items at all, items beyond the radius,
// and unresolvable ids (no catalogue — "item#N" noise must not render). The
// radius boundary itself still renders.
func TestGroundItemsHintAbsent(t *testing.T) {
	d := quietDirector()

	h := groundItemTestHost()
	if hint := d.groundItemsHint(h, 120, 504); hint != "" {
		t.Fatalf("no items must render nothing, got:\n%s", hint)
	}
	sit := d.situation(h, Outcome{})
	if strings.Contains(sit.Hints["scene"], "ON THE GROUND") {
		t.Fatalf("scene must not carry an empty ground-item block:\n%s", sit.Hints["scene"])
	}

	h.world.GroundItems.Add(120+groundItemRadius+1, 504, 156) // one tile too far
	if hint := d.groundItemsHint(h, 120, 504); hint != "" {
		t.Fatalf("beyond-radius items must render nothing, got:\n%s", hint)
	}

	h2 := groundItemTestHost()
	h2.world.GroundItems.Add(120+groundItemRadius, 504, 156) // exactly at the radius
	if hint := d.groundItemsHint(h2, 120, 504); !strings.Contains(hint, "Bronze Pickaxe") {
		t.Fatalf("the radius boundary must still render, got:\n%s", hint)
	}

	bare := newTestHost() // h.facts == nil: ids resolve to "item#N"
	bare.world.GroundItems.Add(121, 504, 156)
	if hint := d.groundItemsHint(bare, 120, 504); hint != "" {
		t.Fatalf("unresolvable ids must render nothing, got:\n%s", hint)
	}
}

// TestGroundItemsHintKindCapAndDedup proves the bounds: duplicate spawns of a
// kind fold into ONE line carrying the nearest spawn's distance, the list
// holds the nearest groundItemLineCap kinds in distance order, and the
// farthest kind folds out entirely.
func TestGroundItemsHintKindCapAndDedup(t *testing.T) {
	h := groundItemTestHost()
	h.world.GroundItems.Add(121, 504, 156) // Bronze Pickaxe — 1 tile
	h.world.GroundItems.Add(126, 504, 156) // Bronze Pickaxe again — folds into the 1-tile line
	h.world.GroundItems.Add(122, 504, 0)   // Iron Mace — 2 tiles
	h.world.GroundItems.Add(123, 504, 1)   // Iron Short Sword — 3 tiles
	h.world.GroundItems.Add(124, 504, 2)   // Iron Kite Shield — 4 tiles
	h.world.GroundItems.Add(125, 504, 3)   // Iron Square Shield — 5 tiles
	h.world.GroundItems.Add(120, 511, 4)   // Wooden Shield — 7 tiles, 6th kind: dropped
	d := quietDirector()
	hint := d.groundItemsHint(h, 120, 504)
	if got := strings.Count(hint, "Bronze Pickaxe"); got != 1 {
		t.Fatalf("duplicate spawns must fold into one line, got %d:\n%s", got, hint)
	}
	if !strings.Contains(hint, "Bronze Pickaxe, 1 tiles west") {
		t.Fatalf("the folded line must carry the NEAREST spawn's distance:\n%s", hint)
	}
	if got := strings.Count(hint, "• "); got != groundItemLineCap {
		t.Fatalf("kind lines = %d, want the cap %d:\n%s", got, groundItemLineCap, hint)
	}
	if strings.Contains(hint, "Wooden Shield") {
		t.Fatalf("cap exceeded — the farthest kind must fold out:\n%s", hint)
	}
	// Distance order holds across the survivors.
	if strings.Index(hint, "Iron Mace") > strings.Index(hint, "Iron Short Sword") ||
		strings.Index(hint, "Iron Short Sword") > strings.Index(hint, "Iron Kite Shield") {
		t.Fatalf("kinds must render nearest first:\n%s", hint)
	}
}

// TestGroundItemsHintUnreachableFiltered proves the DSL-1 reach gate applies
// to the render: a spawn in a DIFFERENT walkable component inside the radius
// is never advertised, while a same-component spawn is. Real landscape
// fixture (newOracleHost; skips without the OpenRSC archive). The default
// Lumbridge stand has no negative space within the hint radius, so the test
// RELOCATES the host to a component boundary: some small component's rep tile
// (a walled interior) sits within sight of a neighboring component — the gate
// must do the dropping there, not the distance filter.
func TestGroundItemsHintUnreachableFiltered(t *testing.T) {
	h := newOracleHost(t, 0)
	const maxStands = 200 // bound the search; the first interiors usually hit
	hostX, hostY, negX, negY := -1, -1, -1, -1
	for i, c := range h.worldOracle.Components() {
		if i >= maxStands {
			break
		}
		hc, _, _, ok := h.worldOracle.CompNear(c.RepX, c.RepY)
		if !ok {
			continue
		}
		// Ring-scan the radius for a tile the live gate resolves to a
		// DIFFERENT component (CompNear snaps exactly like reachGate does).
		for r := 1; r <= groundItemRadius && negX < 0; r++ {
			for dx := -r; dx <= r && negX < 0; dx++ {
				for dy := -r; dy <= r; dy++ {
					if dx > -r && dx < r && dy > -r && dy < r {
						continue // ring border only
					}
					x, y := c.RepX+dx, c.RepY+dy
					if oc, _, _, ok := h.worldOracle.CompNear(x, y); ok && oc != hc {
						negX, negY = x, y
						break
					}
				}
			}
		}
		if negX >= 0 {
			hostX, hostY = c.RepX, c.RepY
			break
		}
	}
	if negX < 0 {
		t.Skipf("no component boundary found within %d tiles of the first %d rep tiles", groundItemRadius, maxStands)
	}
	h.world.Self.SetPosition(world.Coord{X: hostX, Y: hostY})

	d := quietDirector()
	h.world.GroundItems.Add(negX, negY, 156) // Bronze Pickaxe in negative space
	if hint := d.groundItemsHint(h, hostX, hostY); hint != "" {
		t.Fatalf("an unreachable-only spawn must render nothing, got:\n%s", hint)
	}
	h.world.GroundItems.Add(hostX, hostY, 156) // and one at the host's own feet
	hint := d.groundItemsHint(h, hostX, hostY)
	if !strings.Contains(hint, "Bronze Pickaxe, at your feet") {
		t.Fatalf("a reachable spawn must render (with the dist-0 wording), got:\n%s", hint)
	}
	if got := strings.Count(hint, "• "); got != 1 {
		t.Fatalf("the unreachable spawn must stay filtered (want 1 line, got %d):\n%s", got, hint)
	}
}

// TestNeverHeldItemEvidence pins the first-touch test itself: untried with no
// evidence, held once either record exists, and a ledger-less host renders no
// cue (never claim an ignorance the evidence cannot back).
func TestNeverHeldItemEvidence(t *testing.T) {
	d := quietDirector()
	h := groundItemTestHost()
	if !d.neverHeldItem(h, "Bronze Pickaxe") {
		t.Fatal("empty hands + empty ledger must read never-held")
	}
	h.knowledge.Observe("bronze pickaxe", heldItemClaim, true, 1.0) // ledger normalizes case
	if d.neverHeldItem(h, "Bronze Pickaxe") {
		t.Fatal("a ledger first-touch record must read held")
	}
	h2 := groundItemTestHost()
	h2.world.Inventory.Replace([]world.InvSlot{{ItemID: 156, Amount: 1}})
	if d.neverHeldItem(h2, "Bronze Pickaxe") {
		t.Fatal("an in-inventory kind must read held")
	}
	h3 := groundItemTestHost()
	h3.knowledge = nil
	if d.neverHeldItem(h3, "Bronze Pickaxe") {
		t.Fatal("no ledger ⇒ no cue: neverHeldItem must answer false")
	}
}
