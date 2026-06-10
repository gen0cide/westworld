package facts

//go:generate go run github.com/gen0cide/westworld/cmd/defsgen -out .

// LoadStatic builds the Facts store from the checked-in generated data
// (static_defs_gen.go / static_locs_gen.go, produced by cmd/defsgen from the
// OpenRSC server tree). This is the production path: a deployed host performs
// ZERO file I/O and needs no openrsc checkout. Parity with the legacy
// file-parsing Load is enforced by TestStaticParity against a live tree.
//
// The generated tables are shared, immutable package data — LoadStatic only
// allocates the spatial/name indexes, so calling it is cheap relative to the
// old XML/JSON parse. Same one-copy-per-process rule as Load applies.
func LoadStatic() *Facts {
	f := &Facts{
		SceneryDefs:  sceneryDefsGen,
		BoundaryDefs: boundaryDefsGen,
		NpcDefs:      npcDefsGen,
		ItemDefs:     itemDefsGen,
		TileDefs:     tileDefsGen,

		SceneryLocs:    sceneryLocsGen,
		BoundaryLocs:   boundaryLocsGen,
		NpcLocs:        npcLocsGen,
		GroundItemLocs: groundItemLocsGen,
	}
	f.buildIndex()
	f.buildWornIndex()
	f.gaz = loadGazetteer()
	return f
}

// LoadStaticCatalogs is the generated-data twin of LoadCatalogs: item + npc
// name catalogs and the gazetteer only — no placements, no spatial index.
// For callers that validate names rather than walk the world (mesad).
func LoadStaticCatalogs() *Facts {
	f := &Facts{
		NpcDefs:  npcDefsGen,
		ItemDefs: itemDefsGen,
	}
	f.buildWornIndex()
	f.gaz = loadGazetteer()
	return f
}
