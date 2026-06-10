package spec

// ReservedRoots is the canonical set of namespace-root identifiers a routine
// can never bind (api.md §6). The validator rejects assignments to these, the
// runtime registers a view for each BUILT root, and the manual documents them
// — all three derive from this one list so they cannot drift.
//
// "host" is reserved AHEAD of being wired: its accessors are spec'd
// (NotYetImplemented) and reserving the name now means no authored routine
// can bind `host = ...` that would break the day the view lands.
func ReservedRoots() []string {
	return []string{
		"self", "world", "inventory", "combat",
		"trade", "bank", "duel", "magic", "prayer", "shop",
		"host",
	}
}
