package spec

// ViewSpec documents one ENTITY VIEW — the field surface of the values
// the query layer returns (an NPC from world.npcs, an InvSlot from
// inventory.find, a placement from scan_for, ...). These are not rooted
// at a fixed accessor path (they live ON returned values), so they get
// their own table instead of spec.Accessors rows (DSL-2).
//
// Same honesty contract as Accessors: every field listed here is
// verified against the view's Get() switch in the runtime/views_*.go
// family — the spec documents what EXISTS, never what is planned.
// Stub fields that always return the same value are documented as
// such so routines don't branch on them.
type ViewSpec struct {
	// Kind is the view's Kind() string — what `.Kind()`/error messages
	// call it (e.g. "npc", "ground_item", "placement").
	Kind string

	// Doc is one line: what the view represents and which selectors
	// return it.
	Doc string

	// Fields is the verified field table.
	Fields []ViewFieldSpec
}

// ViewFieldSpec documents one field on an entity view.
type ViewFieldSpec struct {
	Name       string
	Kind       string
	DocSummary string
}

// Views is the canonical entity-view field surface, rendered into the
// planner manual by APIReference. Order matters only for readability.
var Views = []ViewSpec{
	{
		Kind: "npc",
		Doc:  "A live NPC from world.npcs / nearest_npc / combat.target.",
		Fields: []ViewFieldSpec{
			{Name: "index", Kind: "int", DocSummary: "Server roster index (passes to attack/talk_to/converse)."},
			{Name: "type_id", Kind: "int", DocSummary: "NPC species def id."},
			{Name: "x", Kind: "int", DocSummary: "Tile x."},
			{Name: "y", Kind: "int", DocSummary: "Tile y."},
			{Name: "position", Kind: "Position", DocSummary: "{x, y} as a position value."},
			{Name: "name", Kind: "string", DocSummary: "Species name from facts; \"\" while the def is still loading (predicates just see a non-match — no null guard needed)."},
			{Name: "def", Kind: "NpcDef?", DocSummary: "The static catalog entry, or Null if facts not loaded."},
			{Name: "combat_level", Kind: "int?", DocSummary: "Derived combat level ((atk+str+def)/4 + hits/4), or Null without facts."},
			{Name: "relative_level", Kind: "int?", DocSummary: "Their combat level minus yours (positive = stronger than you)."},
			{Name: "threat", Kind: "string?", DocSummary: "Relative danger band (\"trivial\"..\"deadly\") — the level-colour cue the client paints."},
			{Name: "threat_colour", Kind: "string?", DocSummary: "The colour name for the threat band."},
			{Name: "max_hp", Kind: "int?", DocSummary: "Species max hits from facts."},
			{Name: "is_attackable", Kind: "bool", DocSummary: "Def-level attackable flag (false without facts)."},
			{Name: "is_aggressive", Kind: "bool", DocSummary: "Def-level aggression flag (false without facts)."},
			{Name: "hp_fraction", Kind: "float?", DocSummary: "cur/max hits — NULL until this NPC has been fought (only the engaged target gets health updates on the wire); null means unhurt/unknown, not dead."},
			{Name: "health", Kind: "int?", DocSummary: "Current hits, same fought-only gate as hp_fraction."},
			{Name: "in_combat_with", Kind: "Null", DocSummary: "STUB — always Null until per-NPC combat tracking lands. Do not branch on it."},
		},
	},
	{
		Kind: "player",
		Doc:  "A visible player from world.players / combat.target / trade events.",
		Fields: []ViewFieldSpec{
			{Name: "index", Kind: "int", DocSummary: "Server roster index."},
			{Name: "name", Kind: "string", DocSummary: "Player username."},
			{Name: "x", Kind: "int", DocSummary: "Tile x."},
			{Name: "y", Kind: "int", DocSummary: "Tile y."},
			{Name: "position", Kind: "Position", DocSummary: "{x, y} as a position value."},
			{Name: "combat_level", Kind: "int?", DocSummary: "From the appearance packet; Null until decoded."},
			{Name: "relative_level", Kind: "int?", DocSummary: "Their combat level minus yours; Null until appearance decoded."},
			{Name: "threat", Kind: "string?", DocSummary: "Relative danger band (\"trivial\"..\"deadly\"); Null until appearance decoded."},
			{Name: "threat_colour", Kind: "string?", DocSummary: "Colour name for the threat band."},
			{Name: "is_skulled", Kind: "bool", DocSummary: "PK skull showing."},
			{Name: "hp_fraction", Kind: "float?", DocSummary: "cur/max hits — Null unless you have fought them (PVP health arrives only for the engaged opponent)."},
			{Name: "health", Kind: "int?", DocSummary: "Current hits, same fought-only gate."},
			{Name: "equipment", Kind: "EquipmentView", DocSummary: "Worn-equipment perception for the whole set; per-slot reads also hang directly off the player (.helmet / .weapon / .shield / ...)."},
			{Name: "is_friend", Kind: "bool", DocSummary: "STUB — always false until the friend-list packet is mirrored. Do not branch on it."},
			{Name: "in_combat_with", Kind: "Null", DocSummary: "STUB — always Null until per-player combat tracking lands."},
		},
	},
	{
		Kind: "ground_item",
		Doc:  "A ground item from world.ground_items selectors; passes to pick_up / use.",
		Fields: []ViewFieldSpec{
			{Name: "id", Kind: "int", DocSummary: "Item def id (alias: item_id)."},
			{Name: "item_id", Kind: "int", DocSummary: "Alias of id."},
			{Name: "x", Kind: "int", DocSummary: "Tile x."},
			{Name: "y", Kind: "int", DocSummary: "Tile y."},
			{Name: "position", Kind: "Position", DocSummary: "{x, y} as a position value."},
			{Name: "name", Kind: "string", DocSummary: "Item name from facts."},
			{Name: "def", Kind: "ItemDef?", DocSummary: "The static catalog entry, or Null if facts not loaded."},
			{Name: "is_mine", Kind: "bool", DocSummary: "STUB — always false until 3-min loot-ownership windows are tracked; pick_up just returns SERVER_REJECTED when it isn't yours."},
		},
	},
	{
		Kind: "placement",
		Doc:  "A map placement from scan_for / world.scenery selectors / world.locs.*.nearest; passes to interact_at(x=, y=) / use / go_to.",
		Fields: []ViewFieldSpec{
			{Name: "x", Kind: "int", DocSummary: "Tile x."},
			{Name: "y", Kind: "int", DocSummary: "Tile y."},
			{Name: "name", Kind: "string", DocSummary: "Def name (\"Rocks\", \"Tree\", \"Fire\")."},
			{Name: "kind", Kind: "string", DocSummary: "\"scenery\" / \"boundary\" / \"npc_spawn\" — decides the opcode use() fires."},
			{Name: "def_id", Kind: "int", DocSummary: "The underlying def id."},
			{Name: "direction", Kind: "int", DocSummary: "Boundary facing (0 for scenery)."},
			{Name: "position", Kind: "Position", DocSummary: "{x, y} as a position value."},
		},
	},
	{
		Kind: "boundary",
		Doc:  "A door/gate/wall-edge from world.boundaries.at/.near; passes to open_boundary / use(key, door).",
		Fields: []ViewFieldSpec{
			{Name: "x", Kind: "int", DocSummary: "Tile x."},
			{Name: "y", Kind: "int", DocSummary: "Tile y."},
			{Name: "direction", Kind: "int", DocSummary: "Wall-edge facing (alias: dir)."},
			{Name: "position", Kind: "Position", DocSummary: "{x, y} as a position value."},
			{Name: "name", Kind: "string", DocSummary: "Def name from facts (\"door\" when unknown)."},
			{Name: "door_type", Kind: "int", DocSummary: "Raw DoorType from the def (0 when unknown)."},
			{Name: "is_openable", Kind: "bool", DocSummary: "True for doors/doorframes the host can open (assumed true when the def is unknown)."},
			{Name: "blocks_when_closed", Kind: "bool", DocSummary: "True for closeable doors that block while closed — open before crossing (false when unknown)."},
		},
	},
	{
		Kind: "item",
		Doc:  "An InvSlot instance from inventory.slots / find / find_all / find_any / self.wielded.",
		Fields: []ViewFieldSpec{
			{Name: "idx", Kind: "int?", DocSummary: "Slot index; Null when slot context is unknown (self.wielded)."},
			{Name: "quantity", Kind: "int", DocSummary: "Amount in this slot (alias: amount)."},
			{Name: "def", Kind: "ItemDef?", DocSummary: "The static catalog entry, or Null if facts not loaded."},
			{Name: "id", Kind: "int", DocSummary: "Item def id."},
			{Name: "amount", Kind: "int", DocSummary: "Alias of quantity."},
			{Name: "name", Kind: "string", DocSummary: "Item name."},
			{Name: "is_stackable", Kind: "bool", DocSummary: "Def-level stackable flag (false without facts)."},
			{Name: "is_wearable", Kind: "bool", DocSummary: "Def-level wearable flag."},
			{Name: "is_wielded", Kind: "bool", DocSummary: "This slot is currently worn/wielded (false for views built without slot context)."},
			{Name: "is_members_only", Kind: "bool", DocSummary: "Def-level members flag."},
			{Name: "command", Kind: "string", DocSummary: "Default right-click command (Eat/Drink/Bury/Wield/...)."},
		},
	},
	{
		Kind: "item_def",
		Doc:  "An ItemDef DEFINITION (static catalog entry) reached via .def on items/ground items.",
		Fields: []ViewFieldSpec{
			{Name: "id", Kind: "int", DocSummary: "Catalog id."},
			{Name: "name", Kind: "string", DocSummary: "Catalog name."},
			{Name: "description", Kind: "string", DocSummary: "Examine text."},
			{Name: "command", Kind: "string", DocSummary: "Default right-click command."},
			{Name: "is_stackable", Kind: "bool", DocSummary: "Stackable flag (alias: stackable)."},
			{Name: "is_wearable", Kind: "bool", DocSummary: "Wearable flag (alias: wearable)."},
			{Name: "is_members_only", Kind: "bool", DocSummary: "Members flag (alias: members)."},
			{Name: "is_tradable", Kind: "bool", DocSummary: "Tradable flag (alias: tradable)."},
			{Name: "is_edible", Kind: "bool", DocSummary: "Command is Eat/Drink — the closest GUI-visible edibility signal (alias: edible)."},
		},
	},
	{
		Kind: "npc_def",
		Doc:  "An NpcDef DEFINITION (static catalog entry) reached via Npc.def.",
		Fields: []ViewFieldSpec{
			{Name: "id", Kind: "int", DocSummary: "Catalog id."},
			{Name: "name", Kind: "string", DocSummary: "Species name."},
			{Name: "description", Kind: "string", DocSummary: "Examine text."},
			{Name: "combat_level", Kind: "int", DocSummary: "Derived combat level."},
			{Name: "max_hp", Kind: "int", DocSummary: "Species max hits."},
			{Name: "is_attackable", Kind: "bool", DocSummary: "Attackable flag (alias: attackable)."},
			{Name: "is_aggressive", Kind: "bool", DocSummary: "Aggression flag (alias: aggressive)."},
			{Name: "command1", Kind: "string", DocSummary: "Primary interaction command (e.g. \"Pickpocket\")."},
			{Name: "command2", Kind: "string", DocSummary: "Secondary interaction command."},
		},
	},
	{
		Kind: "equip_slot",
		Doc:  "One worn-equipment slot from self.equipped.<slot> (.weapon / .shield / .head / ...).",
		Fields: []ViewFieldSpec{
			{Name: "slot", Kind: "int", DocSummary: "Primary body-animation layer index."},
			{Name: "slot_name", Kind: "string", DocSummary: "Readable slot name."},
			{Name: "sprite_id", Kind: "int", DocSummary: "Worn appearance sprite id (0 = nothing/not observed) — the honest wire datum."},
			{Name: "is_empty", Kind: "bool", DocSummary: "Nothing worn in this slot."},
			{Name: "id", Kind: "int?", DocSummary: "Item id, resolved from your own inventory's wielded slots; Null when not exactly resolvable."},
			{Name: "name", Kind: "string?", DocSummary: "Worn item label; Null when empty."},
			{Name: "def", Kind: "ItemDef?", DocSummary: "Catalog entry when exactly resolved; Null otherwise."},
		},
	},
	{
		Kind: "equipment_bonuses",
		Doc:  "Summed worn-gear combat bonuses from self.equipped.bonuses.",
		Fields: []ViewFieldSpec{
			{Name: "armour", Kind: "int", DocSummary: "Total armour bonus."},
			{Name: "aim", Kind: "int", DocSummary: "Total weapon-aim bonus (alias: weapon_aim)."},
			{Name: "power", Kind: "int", DocSummary: "Total weapon-power bonus (alias: weapon_power)."},
			{Name: "magic", Kind: "int", DocSummary: "Total magic bonus."},
			{Name: "prayer", Kind: "int", DocSummary: "Total prayer bonus."},
		},
	},
	{
		Kind: "message",
		Doc:  "One entry of the world.messages ring.",
		Fields: []ViewFieldSpec{
			{Name: "text", Kind: "string", DocSummary: "Message text (alias: message)."},
			{Name: "kind", Kind: "string", DocSummary: "Message kind (\"server\" unless classified)."},
			{Name: "at", Kind: "string", DocSummary: "HH:MM:SS arrival time."},
			{Name: "contains", Kind: "callable(needle)->bool", DocSummary: "Case-insensitive substring test."},
		},
	},
}
