package mesad

// World name catalogs for static argument validation. The Act planner
// sometimes emits well-formed DSL with hallucinated string args —
// go_to("mining-site") with no such POI, eat("typo-item") — which parse
// fine and Fail at runtime, wasting a host round-trip. argCatalog holds
// the four name sets (item names, npc names, place names, POI types) so
// mesad can statically reject a bad literal BEFORE it ships, feeding the
// existing Act re-prompt loop.
//
// This is the facts-backed implementation of validator.Catalog; the
// validator stays facts-free (the catalog is injected). Matching mirrors
// runtime resolution exactly so we never reject what the engine accepts:
//   - place_or_poi: case-insensitive SUBSTRING over place names ∪ POI
//     types (runtime dslGoTo → PlaceByName / NearestPOI).
//   - item: case-insensitive EXACT, then SUBSTRING fallback (runtime
//     resolveItemID matches an exact name, else a substring of any item name).

import (
	"sort"
	"strings"

	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/validator"
	"github.com/gen0cide/westworld/facts"
)

// argCatalog is the loaded world name-set, queried by the static arg
// checker. A nil *argCatalog is a valid no-op catalog (every check
// passes) — so validation degrades gracefully when -facts is unset.
type argCatalog struct {
	items     map[string]bool // lowercased item names (exact match)
	itemNames []string        // lowercased item names (substring fallback)
	places    []string        // lowercased place names (substring match)
	poiTypes  []string        // distinct lowercased POI types (substring match)
	itemEx    []string        // a few real item names, for messages
	placeEx   []string        // a few real place names / POI types, for messages
}

// NewArgCatalog builds the name sets from a defs-only Facts (LoadCatalogs).
// f may be nil, in which case the catalog is empty (everything passes).
func NewArgCatalog(f *facts.Facts) *argCatalog {
	c := &argCatalog{items: map[string]bool{}}
	if f == nil {
		return c
	}
	for _, d := range f.ItemDefs {
		if d == nil {
			continue
		}
		if n := strings.ToLower(strings.TrimSpace(d.Name)); n != "" && !c.items[n] {
			c.items[n] = true
			c.itemNames = append(c.itemNames, n)
		}
	}
	g := f.Gazetteer()
	seenPlace := map[string]bool{}
	for _, p := range g.Places {
		if n := strings.ToLower(strings.TrimSpace(p.Name)); n != "" && !seenPlace[n] {
			seenPlace[n] = true
			c.places = append(c.places, n)
		}
	}
	seenType := map[string]bool{}
	for _, p := range g.POIs {
		if t := strings.ToLower(strings.TrimSpace(p.Type)); t != "" && !seenType[t] {
			seenType[t] = true
			c.poiTypes = append(c.poiTypes, t)
		}
	}
	c.buildExamples()
	return c
}

// loaded reports whether the catalog actually holds data — when false,
// validation should be skipped entirely (graceful degradation).
func (c *argCatalog) loaded() bool {
	return c != nil && (len(c.items) > 0 || len(c.places) > 0 || len(c.poiTypes) > 0)
}

// KnownPlaceOrPOI mirrors runtime go_to/where_is string resolution: a
// case-insensitive SUBSTRING match against any place name OR POI type.
func (c *argCatalog) KnownPlaceOrPOI(s string) bool {
	want := strings.ToLower(strings.TrimSpace(s))
	if want == "" || c == nil {
		return true // empty/no catalog → don't reject
	}
	// Mirror runtime exactly: PlaceByName matches when the QUERY is a
	// substring of a place name; NearestPOI matches when the QUERY is a
	// substring of a POI type. The match is ONE-directional (query ⊆
	// catalog entry) — a hallucinated SUPERSTRING like "mining-site-area"
	// is NOT a substring of "mining-site", so it is correctly rejected.
	for _, p := range c.places {
		if strings.Contains(p, want) {
			return true
		}
	}
	for _, t := range c.poiTypes {
		if strings.Contains(t, want) {
			return true
		}
	}
	return false
}

// KnownItem mirrors runtime resolveItemID EXACTLY: a case-insensitive EXACT
// match, then a SUBSTRING fallback (the query is a substring of any real item
// name). resolveItemID resolves use("pickaxe") → "mithril pickaxe" via this
// same fallback, so an exact-only check would over-reject valid args — worse
// than the hallucination it's meant to catch.
func (c *argCatalog) KnownItem(s string) bool {
	want := strings.ToLower(strings.TrimSpace(s))
	if want == "" || c == nil || len(c.items) == 0 {
		return true
	}
	if c.items[want] {
		return true
	}
	for _, n := range c.itemNames {
		if strings.Contains(n, want) {
			return true
		}
	}
	return false
}

// Examples returns a few real, valid sample values for the rejection
// message, so the re-prompt re-teaches valid forms.
func (c *argCatalog) Examples(kind string) []string {
	if c == nil {
		return nil
	}
	switch kind {
	case spec.CatalogItem:
		return c.itemEx
	case spec.CatalogPlaceOrPOI:
		return c.placeEx
	}
	return nil
}

// buildExamples picks a small, stable, recognizable set of samples.
func (c *argCatalog) buildExamples() {
	// Prefer well-known POI types / towns for place_or_poi so the hint is
	// instantly recognizable; fall back to whatever the catalog holds.
	c.placeEx = pickPresent(c.poiTypes, "bank", "furnace", "altar", "fishing-point", "mining-site")
	if len(c.placeEx) < 3 {
		c.placeEx = append(c.placeEx, pickPresent(c.places, "lumbridge", "varrock", "falador")...)
	}
	c.itemEx = pickPresent(itemKeys(c.items), "bread", "cookedmeat", "cooked meat", "bronze short sword", "logs")
	if len(c.itemEx) < 3 {
		// Top up with real arbitrary names so the hint always has examples,
		// regardless of how this dataset happens to spell things.
		for _, k := range itemKeys(c.items) {
			if len(c.itemEx) >= 4 {
				break
			}
			if !contains(c.itemEx, k) {
				c.itemEx = append(c.itemEx, k)
			}
		}
	}
}

// pickPresent returns the wanted values that actually exist in pool,
// preserving the wanted order; capped at 5.
func pickPresent(pool []string, wanted ...string) []string {
	set := map[string]bool{}
	for _, p := range pool {
		set[p] = true
	}
	var out []string
	for _, w := range wanted {
		if set[w] {
			out = append(out, w)
			if len(out) >= 5 {
				break
			}
		}
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func itemKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out) // deterministic fallback samples
	return out
}

// compile-time check: *argCatalog satisfies the validator's Catalog.
var _ validator.Catalog = (*argCatalog)(nil)
