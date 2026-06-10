package validator_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/validator"
)

// stubCatalog is a tiny in-test world catalog: a few items and a few
// place/POI strings, matching runtime semantics (item EXACT, place
// SUBSTRING). It implements validator.Catalog.
type stubCatalog struct {
	items  map[string]bool
	places []string // lowercased place names / POI types, substring-matched (KnownPlaceOrPOI)
	towns  []string // lowercased TOWN names only, substring-matched (KnownPlace — go_to's target)
}

func newStub() *stubCatalog {
	return &stubCatalog{
		items:  map[string]bool{"cooked meat": true, "bronze sword": true, "bread": true},
		places: []string{"lumbridge", "varrock", "bank", "furnace", "altar", "mining-site", "fishing-point"},
		towns:  []string{"lumbridge", "varrock", "falador"},
	}
}

func (c *stubCatalog) KnownPlace(s string) bool {
	want := strings.ToLower(strings.TrimSpace(s))
	for _, t := range c.towns {
		if strings.Contains(t, want) {
			return true
		}
	}
	return false
}

func (c *stubCatalog) KnownItem(s string) bool {
	return c.items[strings.ToLower(strings.TrimSpace(s))]
}

func (c *stubCatalog) KnownPlaceOrPOI(s string) bool {
	want := strings.ToLower(strings.TrimSpace(s))
	// One-directional, mirroring runtime: query must be a substring of a
	// catalog entry (a superstring like "mining-site-area" is rejected).
	for _, p := range c.places {
		if strings.Contains(p, want) {
			return true
		}
	}
	return false
}

func (c *stubCatalog) Examples(kind string) []string {
	switch kind {
	case spec.CatalogItem:
		return []string{"cooked meat", "bread"}
	case spec.CatalogPlaceOrPOI:
		return []string{"bank", "furnace", "altar"}
	case spec.CatalogPlace:
		return []string{"Lumbridge", "Varrock"}
	}
	return nil
}

func checkArgs(t *testing.T, src string) []error {
	t.Helper()
	f, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return validator.CheckArgLiterals(f, newStub())
}

func wantArgsOK(t *testing.T, src string) {
	t.Helper()
	if errs := checkArgs(t, src); len(errs) > 0 {
		t.Fatalf("expected no arg errors, got %v\n--- src ---\n%s", errs, src)
	}
}

func wantArgErr(t *testing.T, src, wantSubstr string) {
	t.Helper()
	errs := checkArgs(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected an arg error containing %q, got none\n--- src ---\n%s", wantSubstr, src)
	}
	joined := ""
	for _, e := range errs {
		joined += e.Error() + "\n"
	}
	if !strings.Contains(joined, wantSubstr) {
		t.Errorf("errors %q do not contain %q\n--- src ---\n%s", joined, wantSubstr, src)
	}
}

// ----- go_to (place_or_poi), substring semantics -----

func TestGoToBogusPlaceRejected(t *testing.T) {
	wantArgErr(t, `routine r() { go_to("the mine") }`, `"the mine" is not a known town/place`)
}

func TestGoToMiningSiteAreaRejected(t *testing.T) {
	// A free description that is NOT a substring of any place name or POI type.
	wantArgErr(t, `routine r() { go_to("mining-site-area-far-east") }`, "go_to")
}

func TestGoToPOITypeRejected(t *testing.T) {
	// go_to no longer takes a POI TYPE — "bank" must be rejected (use search_map
	// then go_to the coords). Town names + coords are the only literal targets.
	wantArgErr(t, `routine r() { go_to("bank") }`, "go_to")
}

func TestGoToMiningSiteRejected(t *testing.T) {
	// "mining-site" is a POI type, no longer a valid go_to target.
	wantArgErr(t, `routine r() { go_to("mining-site") }`, "go_to")
}

func TestGoToKnownTownPasses(t *testing.T) {
	wantArgsOK(t, `routine r() { go_to("Lumbridge") }`)
}

func TestGoToCoordinatesSkipped(t *testing.T) {
	wantArgsOK(t, `routine r() { go_to(120, 504) }`)
}

func TestGoToBangVariantChecked(t *testing.T) {
	wantArgErr(t, `routine r() { go_to!("totally-not-a-place") }`, "go_to")
}

// ----- eat / use / equip (item), exact semantics -----

func TestEatBogusItemRejected(t *testing.T) {
	wantArgErr(t, `routine r() { eat("totally-not-an-item") }`, `"totally-not-an-item" is not a real item`)
}

func TestEatRealItemPasses(t *testing.T) {
	wantArgsOK(t, `routine r() { eat("cooked meat") }`)
}

func TestUseItemArgChecked(t *testing.T) {
	wantArgErr(t, `routine r() { use("nonexistent thing", x=216, y=731) }`, "use")
}

func TestUseRealItemWithCoordsPasses(t *testing.T) {
	wantArgsOK(t, `routine r() { use("bronze sword", x=216, y=731) }`)
}

func TestEquipBogusItemRejected(t *testing.T) {
	wantArgErr(t, `routine r() { equip("imaginary armour") }`, "equip")
}

// ----- dynamic args are skipped (only literals are checkable) -----

func TestVariableArgSkipped(t *testing.T) {
	wantArgsOK(t, `routine r() { food = "whatever"; eat(food) }`)
}

func TestFStringArgSkipped(t *testing.T) {
	wantArgsOK(t, `routine r() { x = 1; go_to(f"place {x}") }`)
}

func TestViewArgSkipped(t *testing.T) {
	wantArgsOK(t, `routine r() { talk_to(world.npcs.find(n => n.name == "Guide")) }`)
}

func TestNearestNpcSkipped(t *testing.T) {
	wantArgsOK(t, `routine r() { attack(nearest_npc()) }`)
}

// ----- npc args are SOFT: literal npc names are not hard-rejected -----

func TestTalkToLiteralNpcNotRejected(t *testing.T) {
	// Even a clearly-absent NPC name passes the static check (npc is soft).
	wantArgsOK(t, `routine r() { talk_to("Nonexistent NPC") }`)
}

// ----- nil catalog disables the check -----

func TestNilCatalogPasses(t *testing.T) {
	f, err := parser.Parse("t.routine", `routine r() { go_to("the mine") }`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if errs := validator.CheckArgLiterals(f, nil); len(errs) != 0 {
		t.Fatalf("nil catalog should produce no errors, got %v", errs)
	}
}

// ----- the checker descends into nested blocks / loops -----

func TestNestedBlockChecked(t *testing.T) {
	src := `routine r() {
		while inventory.free > 0 {
			if self.hp < 5 { eat("phantom food") }
		}
	}`
	wantArgErr(t, src, `"phantom food" is not a real item`)
}
