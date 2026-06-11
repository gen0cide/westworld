package mesad

import (
	"os"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/facts"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

const testFactsRoot = "/Users/flint/Code/openrsc"

// loadTestCatalog builds the real facts-backed catalog, or skips the test
// when the OpenRSC tree isn't present (keeps CI hermetic).
func loadTestCatalog(t *testing.T) *argCatalog {
	t.Helper()
	if _, err := os.Stat(testFactsRoot); err != nil {
		t.Skipf("OpenRSC facts tree not present at %s; skipping catalog test", testFactsRoot)
	}
	f, err := facts.LoadCatalogs(facts.DefaultSources(testFactsRoot))
	if err != nil {
		t.Fatalf("LoadCatalogs: %v", err)
	}
	cat := NewArgCatalog(f)
	if !cat.loaded() {
		t.Fatalf("catalog loaded but empty")
	}
	return cat
}

func routineMove(src string) *mesapb.Move {
	return &mesapb.Move{Kind: mesapb.MoveKind_WRITE_ROUTINE, DslSource: src}
}

func directMove(verb string, args ...string) *mesapb.Move {
	return &mesapb.Move{Kind: mesapb.MoveKind_DIRECT_ACTION, Verb: verb, ActionArgs: args}
}

// ----- the headline requirement: bad literals rejected, good ones pass -----

func TestValidateMoveRejectsHallucinatedGoTo(t *testing.T) {
	cat := loadTestCatalog(t)
	// A free description that is NOT a real place name or POI type.
	_, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to("the secret mine to the far east") }`), cat)
	if err == nil {
		t.Fatalf("expected go_to(<bogus>) to be rejected")
	}
	if !strings.Contains(err.Error(), "go_to") {
		t.Errorf("rejection should name the verb; got %q", err)
	}
}

func TestValidateMoveRejectsHallucinatedEat(t *testing.T) {
	cat := loadTestCatalog(t)
	_, err := validateMove(routineMove(`runtime "1.0"
routine r() { eat("totally-not-an-item") }`), cat)
	if err == nil {
		t.Fatalf("expected eat(<bogus>) to be rejected")
	}
	if !strings.Contains(err.Error(), "item") {
		t.Errorf("rejection should mention item; got %q", err)
	}
}

func TestValidateMoveRejectsPOIType(t *testing.T) {
	cat := loadTestCatalog(t)
	// go_to no longer takes a POI TYPE — "mining-site"/"bank" must be rejected
	// (use search_map, then go_to the coords). Town names + coords still pass.
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to("mining-site") }`), cat); err == nil {
		t.Fatalf("go_to(\"mining-site\") should be rejected (POI type, not a town)")
	}
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to("bank") }`), cat); err == nil {
		t.Fatalf("go_to(\"bank\") should be rejected (POI type, not a town)")
	}
	// a known TOWN name still resolves and passes.
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to("Lumbridge") }`), cat); err != nil {
		t.Fatalf("go_to(\"Lumbridge\") should pass (town), got %v", err)
	}
}

func TestValidateMoveAcceptsCoordinates(t *testing.T) {
	cat := loadTestCatalog(t)
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to(120, 504) }`), cat); err != nil {
		t.Fatalf("go_to(120, 504) should pass, got %v", err)
	}
}

func TestValidateMoveAcceptsRealItem(t *testing.T) {
	cat := loadTestCatalog(t)
	// "bread" is a real item in the dataset.
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { eat("bread") }`), cat); err != nil {
		t.Fatalf("eat(\"bread\") should pass, got %v", err)
	}
}

func TestValidateMoveAcceptsItemSubstring(t *testing.T) {
	cat := loadTestCatalog(t)
	// resolveItemID does exact-then-SUBSTRING, so use("pickaxe") resolves to
	// "mithril pickaxe" (etc.) at runtime. The static check must mirror that and
	// NOT over-reject the short literal — over-rejecting a valid routine is worse
	// than the hallucination this feature catches.
	for _, src := range []string{
		"runtime \"1.0\"\nroutine r() { use(\"pickaxe\") }",
		"runtime \"1.0\"\nroutine r() { use(\"sword\") }",
	} {
		if _, err := validateMove(routineMove(src), cat); err != nil {
			t.Fatalf("substring item arg should pass (mirrors resolveItemID); got %v for %q", err, src)
		}
	}
}

// ----- direct_action path uses the same catalog -----

func TestValidateDirectActionRejectsHallucinatedGoTo(t *testing.T) {
	cat := loadTestCatalog(t)
	if _, err := validateMove(directMove("go_to", "the secret mine to the far east"), cat); err == nil {
		t.Fatalf("expected direct go_to(<bogus>) to be rejected")
	}
}

func TestValidateDirectActionRejectsPOIType(t *testing.T) {
	cat := loadTestCatalog(t)
	// go_to no longer takes a POI TYPE — a direct go_to("bank") is rejected.
	if _, err := validateMove(directMove("go_to", "bank"), cat); err == nil {
		t.Fatalf("direct go_to(\"bank\") should be rejected (POI type, not a town)")
	}
	// a known town still passes.
	if _, err := validateMove(directMove("go_to", "Varrock"), cat); err != nil {
		t.Fatalf("direct go_to(\"Varrock\") should pass (town), got %v", err)
	}
}

// ----- dynamic args + npc soft-catalog are not rejected -----

func TestValidateMoveSkipsDynamicAndNpcArgs(t *testing.T) {
	cat := loadTestCatalog(t)
	// variable item arg, nearest_npc(), and a literal (situation-dependent)
	// npc name all pass the static check.
	src := `runtime "1.0"
routine r() {
	food = "whatever"
	eat(food)
	attack(nearest_npc())
	talk_to("Some Random Person")
}`
	if _, err := validateMove(routineMove(src), cat); err != nil {
		t.Fatalf("dynamic/npc args should not be rejected, got %v", err)
	}
}

// ----- nil/unloaded catalog disables the extra check (graceful) -----

func TestValidateMoveNilCatalogSkips(t *testing.T) {
	var cat *argCatalog // nil → loaded() false → arg check skipped
	if _, err := validateMove(routineMove(`runtime "1.0"
routine r() { go_to("the secret mine to the far east") }`), cat); err != nil {
		t.Fatalf("nil catalog should skip arg validation, got %v", err)
	}
}
