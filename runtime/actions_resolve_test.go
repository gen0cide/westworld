package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// resolveTestFacts is a small, hermetic facts registry covering the
// item/npc canonical names the resolve() builtin tests exercise.
func resolveTestFacts() *facts.Facts {
	return &facts.Facts{
		ItemDefs: map[int]*facts.ItemDef{
			4:  {ID: 4, Name: "Rune 2-handed Sword", Command: "Wield"},
			10: {ID: 10, Name: "Attack potion", Command: "Drink"},
		},
		NpcDefs: map[int]*facts.NpcDef{
			3: {ID: 3, Name: "Chicken", Command1: "Attack", Attackable: true},
		},
		SceneryDefs:  map[int]*facts.SceneryDef{1: {ID: 1, Name: "Anvil", Command1: "Smith"}},
		BoundaryDefs: map[int]*facts.BoundaryDef{},
	}
}

// newResolveTestHost builds a Host wired with a facts-backed resolver
// whose alias store has been seeded so a stage-1 (exact alias) hit is
// deterministic. No network, no world mirror needed for these.
func newResolveTestHost(t *testing.T) *Host {
	t.Helper()
	f := resolveTestFacts()
	h := New(Options{Username: "test", Facts: f})
	store := resolve.NewAliasStore()
	store.Seed(map[string]resolve.SeedAlias{
		"r2h": {Canonical: "Rune 2-handed Sword", Kind: resolve.KindItem},
	})
	h.Resolver = resolve.New(f, store, nil)
	return h
}

// TestResolveOneAliasHit is the headline path: a learned alias resolves
// to the right item def + id through the DSL primitive.
func TestResolveOneAliasHit(t *testing.T) {
	h := newResolveTestHost(t)

	res := runRoutine(t, h, `routine r() { return resolve_one("r2h").def.id }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 4 {
		t.Fatalf(`resolve_one("r2h").def.id: got %v, want Int(4)`, res.Value)
	}

	res = runRoutine(t, h, `routine r() { return resolve_one("r2h").canonical }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "Rune 2-handed Sword" {
		t.Fatalf(`resolve_one("r2h").canonical: got %v, want "Rune 2-handed Sword"`, res.Value)
	}

	res = runRoutine(t, h, `routine r() { return resolve_one("r2h").source }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != resolve.SourceAlias {
		t.Fatalf(`resolve_one("r2h").source: got %v, want %q`, res.Value, resolve.SourceAlias)
	}
}

// TestResolveFuzzyRanked exercises stage-2 fuzzy matching through the
// list-returning resolve(): "attack pot" should surface "Attack potion".
func TestResolveFuzzyRanked(t *testing.T) {
	h := newResolveTestHost(t)

	res := runRoutine(t, h, `routine r() { return resolve("attack pot", "item").length }`)
	n, ok := res.Value.(interp.Int)
	if !ok || int64(n) < 1 {
		t.Fatalf(`resolve("attack pot","item").length: got %v, want >=1`, res.Value)
	}

	res = runRoutine(t, h, `routine r() { return resolve("attack pot", "item")[0].canonical }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "Attack potion" {
		t.Fatalf(`resolve("attack pot","item")[0].canonical: got %v, want "Attack potion"`, res.Value)
	}
}

// TestResolveOneMiss confirms resolve_one returns Null (not an error)
// for text that resolves to nothing.
func TestResolveOneMiss(t *testing.T) {
	h := newResolveTestHost(t)

	res := runRoutine(t, h, `routine r() { return resolve_one("zzzzzz nonsense") == null }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Fatalf(`resolve_one(miss) == null: got %v, want Bool(true)`, res.Value)
	}
}

// TestResolveEmptyListMiss confirms the list form returns [] (length 0)
// on a miss rather than Null.
func TestResolveEmptyListMiss(t *testing.T) {
	h := newResolveTestHost(t)

	res := runRoutine(t, h, `routine r() { return resolve("zzzzzz nonsense").length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Fatalf(`resolve(miss).length: got %v, want Int(0)`, res.Value)
	}
}

// TestResolveLazyWithoutExplicitResolver confirms the builtin works on a
// host that never had h.Resolver set — the lazy in-memory resolver is
// built from h.facts, so spells/items still resolve.
func TestResolveLazyWithoutExplicitResolver(t *testing.T) {
	f := resolveTestFacts()
	h := New(Options{Username: "test", Facts: f})
	// No h.Resolver assigned — exercises the lazy fallback.

	res := runRoutine(t, h, `routine r() { return resolve_one("Anvil", "loc").def.is_scenery }`)
	if b, ok := res.Value.(interp.Bool); !ok || !bool(b) {
		t.Fatalf(`resolve_one("Anvil","loc").def.is_scenery: got %v, want Bool(true)`, res.Value)
	}
}
