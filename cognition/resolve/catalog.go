package resolve

import (
	"sort"

	"github.com/gen0cide/westworld/facts"
)

// Kind constants name the catalogs resolve() can search. They double as the
// `kind` filter argument to Resolve and as the Kind field on a Match.
const (
	KindItem   = "item"
	KindNpc    = "npc"
	KindLoc    = "loc"
	KindSpell  = "spell"
	KindPrayer = "prayer"
)

// allKinds is the search order used when no kind filter is supplied. Order is
// deterministic so ties across catalogs break the same way every run.
var allKinds = []string{KindItem, KindNpc, KindLoc, KindSpell, KindPrayer}

// entry is one canonical, searchable name drawn from the facts registry,
// flattened into a uniform shape so the matcher can score every catalog the
// same way.
//
// Def carries the concrete facts definition (*facts.ItemDef, *facts.NpcDef,
// *facts.SceneryDef, *facts.BoundaryDef, *facts.SpellDef, *facts.PrayerDef)
// so a caller can recover the full record and, crucially, the id — ids always
// come from here, never from the LLM. ID is hoisted out for convenience and
// for the (id never invented) guarantee in the api contract.
type entry struct {
	Canonical string
	Kind      string
	ID        int
	Def       any
}

// Catalog is the read-only, name-indexed view of the facts registry that the
// matcher searches. It is built once per resolver from a *facts.Facts plus the
// process-global spell/prayer tables, and is safe for concurrent reads.
type Catalog struct {
	// byKind maps a Kind constant to its flattened, name-sorted entries.
	byKind map[string][]entry
}

// NewCatalog flattens the facts registry into a searchable Catalog.
//
// f may be nil (e.g. a host that has not loaded world defs yet) — in that case
// the item/npc/loc catalogs are empty but spells and prayers, which live in
// the process-global facts tables, are still populated. This keeps resolve()
// useful for the static catalogs even before world data is loaded.
func NewCatalog(f *facts.Facts) *Catalog {
	c := &Catalog{byKind: map[string][]entry{}}

	if f != nil {
		items := make([]entry, 0, len(f.ItemDefs))
		for id, d := range f.ItemDefs {
			if d == nil || d.Name == "" {
				continue
			}
			items = append(items, entry{Canonical: d.Name, Kind: KindItem, ID: id, Def: d})
		}
		c.byKind[KindItem] = sortEntries(items)

		npcs := make([]entry, 0, len(f.NpcDefs))
		for id, d := range f.NpcDefs {
			if d == nil || d.Name == "" {
				continue
			}
			npcs = append(npcs, entry{Canonical: d.Name, Kind: KindNpc, ID: id, Def: d})
		}
		c.byKind[KindNpc] = sortEntries(npcs)

		// "loc" is the union of scenery and boundary defs — the named things
		// a player would refer to as a place/object in the world.
		locs := make([]entry, 0, len(f.SceneryDefs)+len(f.BoundaryDefs))
		for id, d := range f.SceneryDefs {
			if d == nil || d.Name == "" {
				continue
			}
			locs = append(locs, entry{Canonical: d.Name, Kind: KindLoc, ID: id, Def: d})
		}
		for id, d := range f.BoundaryDefs {
			if d == nil || d.Name == "" {
				continue
			}
			locs = append(locs, entry{Canonical: d.Name, Kind: KindLoc, ID: id, Def: d})
		}
		c.byKind[KindLoc] = sortEntries(locs)
	}

	spells := make([]entry, 0, len(facts.Spells))
	for i := range facts.Spells {
		d := &facts.Spells[i]
		if d.Name == "" {
			continue
		}
		spells = append(spells, entry{Canonical: d.Name, Kind: KindSpell, ID: d.ID, Def: d})
	}
	c.byKind[KindSpell] = sortEntries(spells)

	prayers := make([]entry, 0, len(facts.Prayers))
	for i := range facts.Prayers {
		d := &facts.Prayers[i]
		if d.Name == "" {
			continue
		}
		prayers = append(prayers, entry{Canonical: d.Name, Kind: KindPrayer, ID: d.ID, Def: d})
	}
	c.byKind[KindPrayer] = sortEntries(prayers)

	return c
}

// entries returns the searchable entries for kind. kind == "" returns every
// catalog concatenated in allKinds order. The returned slice must not be
// mutated by callers.
func (c *Catalog) entries(kind string) []entry {
	if kind != "" {
		return c.byKind[kind]
	}
	var out []entry
	for _, k := range allKinds {
		out = append(out, c.byKind[k]...)
	}
	return out
}

// find returns the entry for an exact (case-fold/normalized) canonical name
// within kind, used to resolve an alias-store or brain hit back to a def+id.
// kind == "" searches all catalogs in deterministic order.
func (c *Catalog) find(canonical, kind string) (entry, bool) {
	target := normalize(canonical)
	for _, e := range c.entries(kind) {
		if normalize(e.Canonical) == target {
			return e, true
		}
	}
	return entry{}, false
}

// sortEntries returns es sorted by canonical name (case-fold) then id, for
// deterministic iteration and stable tie-breaking in the matcher.
func sortEntries(es []entry) []entry {
	sort.SliceStable(es, func(i, j int) bool {
		ni, nj := normalize(es[i].Canonical), normalize(es[j].Canonical)
		if ni != nj {
			return ni < nj
		}
		return es[i].ID < es[j].ID
	})
	return es
}
