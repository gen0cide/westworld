package spec

// AccessorSpec documents one query-layer attribute path.
//
// Unlike Actions and Events (where the spec drives behavior), the
// accessor implementations live in runtime/dsl_views.go as Getter
// switch arms. The spec here is the DOCUMENTATION + DISCOVERY +
// CONSISTENCY source: every path listed is what hosts should be
// able to read, and the consistency test asserts each path resolves
// against a stub host.
//
// Adding a query: add a row here AND extend the relevant view's
// Get() method in runtime/dsl_views.go.
type AccessorSpec struct {
	// Path is the dotted accessor as a slice. E.g.
	// `self.hp` is ["self", "hp"]; `self.skills.fishing.level`
	// is ["self", "skills", "fishing", "level"].
	Path []string

	// Kind describes the returned value's type (informally — used
	// by tooling and docs, not enforced by the validator). Examples:
	// "int", "float", "string", "bool", "Position", "Inventory",
	// "list<NpcView>", "Error?".
	Kind string

	// DocSummary is one line about what the accessor returns.
	DocSummary string

	// NotYetImplemented marks an accessor the spec defines but
	// runtime/dsl_views.go doesn't yet expose. The consistency
	// test uses this to distinguish "spec drift" from "planned
	// but not built."
	NotYetImplemented bool
}

// Accessors is the canonical query-layer surface. Use the helper
// constructors below to keep entries terse.
//
// Currently this list is partial — it covers what runtime/dsl_views.go
// implements today. The Phase 2.5 query layer build-out (tasks
// #56–#65) extends this to ~100 entries; that work adds spec rows
// in lock-step with the view implementations.
//
// Naming convention reminder (see docs/lang/syntax.md):
//   - Booleans prefix `is_` (is_busy, is_in_combat)
//   - Maximums prefix `max_` (max_hp, max_prayer)
//   - Plain noun for counts (free, length, level)
var Accessors = []AccessorSpec{
	// ===== self — first-person game state =====
	{Path: []string{"self", "position"}, Kind: "Position",
		DocSummary: "Current tile {x, y, plane}."},
	{Path: []string{"self", "hp"}, Kind: "int",
		DocSummary: "Current hits (boostable)."},
	{Path: []string{"self", "max_hp"}, Kind: "int",
		DocSummary: "Base hits skill level."},
	{Path: []string{"self", "prayer"}, Kind: "int",
		DocSummary: "Current prayer points."},
	{Path: []string{"self", "max_prayer"}, Kind: "int",
		DocSummary: "Base prayer skill level."},
	{Path: []string{"self", "fatigue"}, Kind: "int",
		DocSummary: "Fatigue (0-100; RSC sleep mechanic threshold)."},
	{Path: []string{"self", "combat_level"}, Kind: "int",
		DocSummary: "Derived combat level."},
	{Path: []string{"self", "wielded"}, Kind: "ItemView?",
		DocSummary: "Currently-wielded weapon, or null if unarmed."},
	{Path: []string{"self", "skills"}, Kind: "SkillsView",
		DocSummary: "Skill access: self.skills.<name>.level / .max_level / .xp."},

	// ===== inventory =====
	{Path: []string{"inventory", "free"}, Kind: "int",
		DocSummary: "Number of empty inventory slots."},
	{Path: []string{"inventory", "slots"}, Kind: "list<ItemView>",
		DocSummary: "Every inventory slot (item-view with .item_id, .amount, .is_wielded)."},
	{Path: []string{"inventory", "has"}, Kind: "callable(item)->bool",
		DocSummary: "Returns true iff the item is present (id, name, or item-view)."},
	{Path: []string{"inventory", "count"}, Kind: "callable(item)->int",
		DocSummary: "Returns total amount of the named item across slots."},

	// ===== combat (mostly stubs today) =====
	{Path: []string{"combat", "engaged"}, Kind: "bool",
		DocSummary: "Currently in combat?"},
	{Path: []string{"combat", "target"}, Kind: "Entity?",
		DocSummary: "Current combat target (NPC or player), or null."},

	// ===== world — visible entities =====
	{Path: []string{"world", "players"}, Kind: "list<PlayerView>",
		DocSummary: "Players visible in our view radius."},
	{Path: []string{"world", "npcs"}, Kind: "list<NpcView>",
		DocSummary: "NPCs visible in our view radius."},
	{Path: []string{"world", "ground_items"}, Kind: "list<GroundItemView>",
		DocSummary: "Ground items visible in our view radius."},
	{Path: []string{"world", "locs"}, Kind: "LocsView",
		DocSummary: "Static facts-derived locations: banks, altars, fishing_spots, etc."},

	// ----- planned by docs/lang/state.md Build Plan, not yet
	// wired in runtime/dsl_views.go (tasks #56–#65) -----
	{Path: []string{"self", "hp_fraction"}, Kind: "float", DocSummary: "hp / max_hp."},
	{Path: []string{"self", "quest_points"}, Kind: "int", DocSummary: "QP count."},
	{Path: []string{"self", "is_busy"}, Kind: "bool", DocSummary: "Currently performing an action? (stub: always false until action tracking lands)"},
	{Path: []string{"self", "is_in_combat"}, Kind: "bool", DocSummary: "Engaged with an NPC or player? (stub: always false until combat tracking lands)"},
	{Path: []string{"self", "is_sleeping"}, Kind: "bool", DocSummary: "Sleep screen up due to fatigue? (stub: always false until sleep tracking lands)"},
	{Path: []string{"self", "equipped"}, Kind: "EquippedView", DocSummary: "Equipped slots (.weapon/.shield/.head/...).", NotYetImplemented: true},
	{Path: []string{"self", "prayer", "active"}, Kind: "list<Prayer>", DocSummary: "Currently active prayers.", NotYetImplemented: true},
	{Path: []string{"self", "spells", "known"}, Kind: "list<Spell>", DocSummary: "Spells with sufficient level.", NotYetImplemented: true},
	{Path: []string{"inventory", "weight"}, Kind: "int", DocSummary: "Sum of slot weights.", NotYetImplemented: true},
	{Path: []string{"inventory", "is_full"}, Kind: "bool", DocSummary: "True iff free == 0."},
	{Path: []string{"inventory", "find"}, Kind: "callable(item)->ItemView?", DocSummary: "First matching slot, or null."},
	{Path: []string{"inventory", "slot_of"}, Kind: "callable(item)->int?", DocSummary: "Slot index of first matching item, or null."},
	{Path: []string{"world", "last_chat"}, Kind: "ChatEvent?", DocSummary: "Most recent chat in the ring buffer.", NotYetImplemented: true},
	{Path: []string{"world", "last_pm"}, Kind: "PMEvent?", DocSummary: "Most recent private message.", NotYetImplemented: true},
	{Path: []string{"world", "last_damage"}, Kind: "DamageEvent?", DocSummary: "Most recent damage taken.", NotYetImplemented: true},
	{Path: []string{"bank", "is_open"}, Kind: "bool", DocSummary: "Bank UI currently shown.", NotYetImplemented: true},
	{Path: []string{"bank", "slots"}, Kind: "list<ItemView>", DocSummary: "Bank contents.", NotYetImplemented: true},
	{Path: []string{"trade", "is_active"}, Kind: "bool", DocSummary: "Trade UI currently shown.", NotYetImplemented: true},

	// ===== host — persona / identity =====
	{Path: []string{"host", "name"}, Kind: "string",
		DocSummary: "The host's RSC username."},
	{Path: []string{"host", "persona"}, Kind: "PersonaView",
		DocSummary: "Persona traits map.", NotYetImplemented: true},
	{Path: []string{"host", "mood"}, Kind: "map<string,float>",
		DocSummary: "Emotional-state weights summing to 1.0.", NotYetImplemented: true},
	{Path: []string{"host", "motivation"}, Kind: "Motivation",
		DocSummary: "{north_star, current_focus}.", NotYetImplemented: true},
	{Path: []string{"host", "defaults"}, Kind: "DefaultsView",
		DocSummary: "Persona-tier default handler refs (used by `extends host`).", NotYetImplemented: true},
}
