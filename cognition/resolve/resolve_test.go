package resolve

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/facts"
)

// testFacts builds a small, deterministic facts registry covering the item
// and npc canonical names the recognition tests exercise. We construct the
// exported maps directly rather than loading OpenRSC files so the tests are
// hermetic.
func testFacts() *facts.Facts {
	return &facts.Facts{
		ItemDefs: map[int]*facts.ItemDef{
			4:  {ID: 4, Name: "Rune 2-handed Sword", Command: "Wield"},
			5:  {ID: 5, Name: "Rune Longsword", Command: "Wield"},
			10: {ID: 10, Name: "Attack potion", Command: "Drink"},
			11: {ID: 11, Name: "Strength potion", Command: "Drink"},
			12: {ID: 12, Name: "Defense potion", Command: "Drink"},
			13: {ID: 13, Name: "Cooked chicken", Command: "Eat"},
		},
		NpcDefs: map[int]*facts.NpcDef{
			3: {ID: 3, Name: "Chicken", Command1: "Attack", Attackable: true},
			4: {ID: 4, Name: "Goblin", Command1: "Attack", Attackable: true},
		},
		SceneryDefs:  map[int]*facts.SceneryDef{1: {ID: 1, Name: "Anvil", Command1: "Smith"}},
		BoundaryDefs: map[int]*facts.BoundaryDef{},
	}
}

// TestAliasLearned_r2h is the headline scenario: once the host has learned
// "r2h" → "Rune 2-handed Sword", resolve("r2h") is a single, exact, stage-1
// table hit returning the right item def + id.
func TestAliasLearned_r2h(t *testing.T) {
	f := testFacts()
	r := New(f, NewAliasStore(), nil)

	// Before learning: "r2h" has no tokens in common with the canonical name,
	// so the conservative matcher returns nothing (no false positive).
	if got := r.Resolve("r2h", KindItem); len(got) != 0 {
		t.Fatalf("pre-learn resolve(r2h) = %d matches, want 0: %+v", len(got), got)
	}

	// Teach the host its lingo.
	if err := r.Aliases().Learn("r2h", "Rune 2-handed Sword", KindItem); err != nil {
		t.Fatalf("Learn: %v", err)
	}

	got := r.Resolve("r2h", KindItem)
	if len(got) != 1 {
		t.Fatalf("post-learn resolve(r2h) = %d matches, want exactly 1: %+v", len(got), got)
	}
	m := got[0]
	if m.Canonical != "Rune 2-handed Sword" {
		t.Errorf("Canonical = %q, want %q", m.Canonical, "Rune 2-handed Sword")
	}
	if m.ID != 4 {
		t.Errorf("ID = %d, want 4 (from facts, never invented)", m.ID)
	}
	if m.Kind != KindItem {
		t.Errorf("Kind = %q, want %q", m.Kind, KindItem)
	}
	if m.Score != scoreExact {
		t.Errorf("Score = %v, want exact %v for a learned-alias hit", m.Score, scoreExact)
	}
	if m.Source != SourceAlias {
		t.Errorf("Source = %q, want %q", m.Source, SourceAlias)
	}
	def, ok := m.Def.(*facts.ItemDef)
	if !ok || def == nil {
		t.Fatalf("Def is %T, want *facts.ItemDef", m.Def)
	}
	if def.ID != 4 {
		t.Errorf("Def.ID = %d, want 4", def.ID)
	}
}

// TestFuzzy_rune2handed checks the conservative token matcher: the loose,
// space-and-no-hyphen phrasing "rune 2 handed" recognizes "Rune 2-handed
// Sword" as the top match without any prior learning.
func TestFuzzy_rune2handed(t *testing.T) {
	f := testFacts()
	r := New(f, NewAliasStore(), nil)

	got := r.Resolve("rune 2 handed", KindItem)
	if len(got) == 0 {
		t.Fatalf("resolve(rune 2 handed) returned no matches")
	}
	top := got[0]
	if top.Canonical != "Rune 2-handed Sword" {
		t.Fatalf("top match = %q (score %v), want %q\nfull list: %+v",
			top.Canonical, top.Score, "Rune 2-handed Sword", got)
	}
	if top.ID != 4 {
		t.Errorf("top.ID = %d, want 4", top.ID)
	}
	if top.Source != SourceFuzzy {
		t.Errorf("top.Source = %q, want %q", top.Source, SourceFuzzy)
	}
	if top.Score < scoreAllTokensLoose {
		t.Errorf("top.Score = %v, want a strong all-tokens fuzzy score (>= %v)",
			top.Score, scoreAllTokensLoose)
	}
}

// TestAmbiguous_potion checks that a genuinely ambiguous query returns a
// ranked list of every plausible candidate (all three potions), best-first,
// rather than guessing one.
func TestAmbiguous_potion(t *testing.T) {
	f := testFacts()
	r := New(f, NewAliasStore(), nil)

	got := r.Resolve("potion", KindItem)
	if len(got) < 3 {
		t.Fatalf("resolve(potion) = %d matches, want >= 3 (the ambiguous list): %+v",
			len(got), got)
	}

	// Every returned match must actually contain the query token and be an
	// item (kind filter honored).
	wantNames := map[string]bool{
		"Attack potion":   false,
		"Strength potion": false,
		"Defense potion":  false,
	}
	for _, m := range got {
		if m.Kind != KindItem {
			t.Errorf("match %q has kind %q, want item", m.Canonical, m.Kind)
		}
		if _, ok := wantNames[m.Canonical]; ok {
			wantNames[m.Canonical] = true
		}
	}
	for name, seen := range wantNames {
		if !seen {
			t.Errorf("ambiguous list missing expected candidate %q; got %+v", name, got)
		}
	}

	// Ranked best-first: scores must be non-increasing.
	for i := 1; i < len(got); i++ {
		if got[i].Score > got[i-1].Score {
			t.Errorf("matches not sorted best-first at %d: %v > %v",
				i, got[i].Score, got[i-1].Score)
		}
	}
}

// TestKindFilter confirms the kind argument restricts the searched catalog:
// "chicken" as an item finds the cooked item; as an npc finds the monster.
func TestKindFilter(t *testing.T) {
	f := testFacts()
	r := New(f, NewAliasStore(), nil)

	asItem := r.Resolve("chicken", KindItem)
	if len(asItem) == 0 || asItem[0].Kind != KindItem {
		t.Fatalf("resolve(chicken, item) = %+v, want an item match", asItem)
	}
	if asItem[0].Canonical != "Cooked chicken" {
		t.Errorf("item chicken top = %q, want %q", asItem[0].Canonical, "Cooked chicken")
	}

	asNpc := r.Resolve("chicken", KindNpc)
	if len(asNpc) == 0 || asNpc[0].Kind != KindNpc {
		t.Fatalf("resolve(chicken, npc) = %+v, want an npc match", asNpc)
	}
	if asNpc[0].Canonical != "Chicken" {
		t.Errorf("npc chicken top = %q, want %q", asNpc[0].Canonical, "Chicken")
	}
}

// TestBrainFallback_WritesBack exercises stage 3: when the alias store and
// fuzzy matcher both miss, the brain hook is consulted, its answer is
// validated against the catalog, and the resolved alias is WRITTEN BACK so the
// next identical query is a cheap stage-1 alias hit (no second brain call).
func TestBrainFallback_WritesBack(t *testing.T) {
	f := testFacts()
	brain := NewStubBrain(map[string]Candidate{
		// "grunt" shares no substring or token with "Goblin", so the
		// conservative fuzzy matcher can't catch it; only the brain can.
		"grunt": {Canonical: "Goblin", Kind: KindNpc},
	})
	r := New(f, NewAliasStore(), brain)

	// First resolve: misses stages 1+2, hits the brain.
	got := r.Resolve("grunt", KindNpc)
	if len(got) != 1 {
		t.Fatalf("resolve(grunt) = %d matches, want 1 from brain: %+v", len(got), got)
	}
	m := got[0]
	if m.Canonical != "Goblin" || m.Kind != KindNpc {
		t.Errorf("brain match = %q/%q, want Goblin/npc", m.Canonical, m.Kind)
	}
	if m.ID != 4 {
		t.Errorf("brain match ID = %d, want 4 (from facts, not the model)", m.ID)
	}
	if m.Source != SourceBrain {
		t.Errorf("Source = %q, want %q", m.Source, SourceBrain)
	}
	if m.Score != scoreBrain {
		t.Errorf("Score = %v, want %v", m.Score, scoreBrain)
	}
	if brain.Calls != 1 {
		t.Fatalf("brain consulted %d times on first resolve, want 1", brain.Calls)
	}

	// Write-back: the store now knows "gob".
	if c, ok := r.Aliases().Lookup("grunt", KindNpc); !ok || c != "Goblin" {
		t.Fatalf("after brain hit, alias store lookup(grunt) = (%q,%v), want (Goblin,true)", c, ok)
	}

	// Second resolve: must be a stage-1 alias hit — the brain is NOT consulted
	// again.
	got2 := r.Resolve("grunt", KindNpc)
	if len(got2) != 1 || got2[0].Source != SourceAlias {
		t.Fatalf("second resolve(grunt) = %+v, want a single alias-source hit", got2)
	}
	if brain.Calls != 1 {
		t.Errorf("brain consulted %d times total, want exactly 1 (second was a table hit)", brain.Calls)
	}
}

// TestBrainRejectsHallucination ensures the resolver never trusts a brain
// answer that isn't a real canonical name (no invented ids leak through).
func TestBrainRejectsHallucination(t *testing.T) {
	f := testFacts()
	brain := NewStubBrain(map[string]Candidate{
		"xyzzy": {Canonical: "Dragon Plate of Infinite Power", Kind: KindItem}, // not in catalog
	})
	r := New(f, NewAliasStore(), brain)

	got := r.Resolve("xyzzy", KindItem)
	if len(got) != 0 {
		t.Fatalf("resolve(xyzzy) = %+v, want 0 (hallucinated name rejected)", got)
	}
	if _, ok := r.Aliases().Lookup("xyzzy", KindItem); ok {
		t.Errorf("a rejected hallucination must NOT be written back to the store")
	}
}

// TestSpellsAndPrayers confirms the static process-global catalogs resolve
// even with a nil facts registry, and that exact canonical names score 1.0.
func TestSpellsAndPrayers(t *testing.T) {
	r := New(nil, NewAliasStore(), nil)

	// Prayers are a fixed table; "thick skin" is canonical.
	pr := r.Resolve("thick skin", KindPrayer)
	if len(pr) == 0 {
		t.Fatalf("resolve(thick skin, prayer) returned nothing; prayer catalog empty?")
	}
	if pr[0].Canonical != "Thick skin" {
		t.Errorf("prayer top = %q, want %q", pr[0].Canonical, "Thick skin")
	}
	if pr[0].Score != scoreExact {
		t.Errorf("exact prayer score = %v, want %v", pr[0].Score, scoreExact)
	}
	def, ok := pr[0].Def.(*facts.PrayerDef)
	if !ok || def == nil {
		t.Fatalf("prayer Def is %T, want *facts.PrayerDef", pr[0].Def)
	}
}

// TestPersistence round-trips a JSON-backed store: learn, reopen from the same
// path, and confirm the learned alias survives.
func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "host", "aliases.json")

	store, err := LoadAliasStore(path)
	if err != nil {
		t.Fatalf("LoadAliasStore (fresh): %v", err)
	}
	if store.Len() != 0 {
		t.Fatalf("fresh store Len = %d, want 0", store.Len())
	}
	if err := store.Learn("r2h", "Rune 2-handed Sword", KindItem); err != nil {
		t.Fatalf("Learn (persisted): %v", err)
	}

	// Reopen from disk in a brand-new store.
	reopened, err := LoadAliasStore(path)
	if err != nil {
		t.Fatalf("LoadAliasStore (reopen): %v", err)
	}
	c, ok := reopened.Lookup("r2h", KindItem)
	if !ok || c != "Rune 2-handed Sword" {
		t.Fatalf("reopened lookup(r2h) = (%q,%v), want (Rune 2-handed Sword,true)", c, ok)
	}
}

// TestSeed checks a persona can pre-load its starting vocabulary.
func TestSeed(t *testing.T) {
	f := testFacts()
	store := NewAliasStore().Seed(map[string]SeedAlias{
		"att pot": {Canonical: "Attack potion", Kind: KindItem},
	})
	r := New(f, store, nil)

	got := r.Resolve("att pot", KindItem)
	if len(got) != 1 || got[0].Canonical != "Attack potion" {
		t.Fatalf("resolve(att pot) = %+v, want single seeded match to Attack potion", got)
	}
	if got[0].Source != SourceAlias {
		t.Errorf("Source = %q, want %q (seeded entries are alias hits)", got[0].Source, SourceAlias)
	}
}

// TestUnknownContext is a guard: an empty query and a no-match query never
// panic and return nil.
func TestUnknownContext(t *testing.T) {
	r := New(testFacts(), NewAliasStore(), nil)
	if got := r.ResolveCtx(context.Background(), "   ", KindItem); got != nil {
		t.Errorf("resolve(blank) = %+v, want nil", got)
	}
	if got := r.Resolve("zzzzzzzz no such thing", KindItem); len(got) != 0 {
		t.Errorf("resolve(garbage) = %+v, want 0 matches", got)
	}
}
