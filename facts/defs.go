package facts

import "strings"

// SceneryDef is the static definition of a scenery type — a thing
// that occupies tiles in the world (tree, well, table, anvil,
// fishing spot, ladder, etc.).
//
// Loaded from GameObjectDef.xml in OpenRSC's def tree.
type SceneryDef struct {
	ID          int
	Name        string // e.g., "Well", "Tree", "Fishing spot"
	Description string // examine text
	Command1    string // first right-click command, e.g., "Chop", "WalkTo", "Mine"
	Command2    string // second right-click command, typically "Examine"
	Type        int    // 1 = solid scenery; 3 = walkable; others used for special cases
	Width       int    // tile footprint width
	Height      int    // tile footprint height
	Model       string // RSC client model name
}

// BoundaryDef is the static definition of a boundary type — a thing
// that blocks movement along a tile edge (wall, fence, gate, door).
//
// Loaded from DoorDef.xml. The Unknown field is named after the XML
// tag in the original Jagex data; it's effectively an "openable" flag
// — doorframes and unlocked doors carry Unknown=1, plain walls/fences
// carry 0. The server treats a boundary as blocking when
// DoorType != 0 AND Unknown == 0 (see openrsc WorldLoader.loadSection).
type BoundaryDef struct {
	ID          int
	Name        string // e.g., "Wall", "Door", "Fence"
	Description string
	Command1    string // typically "Open" for doors, "WalkTo" for walls
	Command2    string // typically "Examine" or "Close"
	// Height is modelVar1 from DoorDef.xml — the wall's vertical extent in
	// world units (RSC GameData.wallObjectHeight; e.g. 192 for a full wall).
	Height int
	// FrontDeco / BackDeco are modelVar2 / modelVar3 — the front/back face
	// decoration index (RSC wallObjectSomething2 / wallObjectTexture). When
	// >= 0 they index a texture; here we use them to pick a wall fill colour.
	FrontDeco int
	BackDeco  int
	DoorType  int // 0 = passable shell, 1+ = blocks per the server's rule
	Unknown   int // 1 = openable (door/doorframe), 0 = fixed (wall/fence)
}

// BlocksMovement reports whether this boundary type is treated as a
// solid wall in the landscape collision grid. Matches the server's
// loadSection rule. Openable doors (Unknown=1) are returned as
// non-blocking here — the bot resolves them by sending an Open
// interaction before walking through.
func (d *BoundaryDef) BlocksMovement() bool {
	if d == nil {
		return false
	}
	return d.DoorType != 0 && d.Unknown == 0
}

// IsOpenable reports whether this boundary is a door/doorframe the bot can
// open (Unknown==1), as opposed to a fixed wall/fence (Unknown==0).
func (d *BoundaryDef) IsOpenable() bool {
	return d != nil && d.Unknown == 1
}

// BlocksWhenClosed reports whether this boundary blocks movement while it is
// CLOSED — the server's dynamic wall-object rule (DoorType==1), matching the
// OpenRSC server's registerWallObject (and Plutonium world.go:392). A closed
// openable door must be pathfound-to as a wall and opened before crossing; an
// OPEN door does not block. Distinct from BlocksMovement (static fences/walls
// that never open). Consult this against LIVE door state — a static loc alone
// cannot tell open from closed, so do not use it to permanently wall a tile.
func (d *BoundaryDef) BlocksWhenClosed() bool {
	return d != nil && d.DoorType == 1
}

// NpcDef is the static definition of an NPC species — a Chicken, a
// Goblin, a Banker, a Greater Demon.
//
// Loaded from NpcDefs.json.
type NpcDef struct {
	ID          int
	Name        string
	Description string
	Command1    string // typically "Attack" or "Talk-to"
	Command2    string
	Hits        int // base HP
	Attack      int
	Defense     int
	Strength    int
	Attackable  bool
	Aggressive  bool
	// CombatLvl is the NPC's combat level (NpcDefs.json "combatlvl") — the
	// at-a-glance "how dangerous" signal. Ranged flags whether the NPC
	// attacks at range (NpcDefs.json "ranged" is a bool). Both come straight
	// from the server data and were previously dropped on load.
	CombatLvl int
	Ranged    bool
	// Rendering fields (OpenRSC NpcDefs.json). Sprites[layer] is the animation
	// index (into authenticAnimDefs) for each of the 12 body-part layers; -1 =
	// none. The colours are raw 24-bit dye values. Camera1/Camera2 are the
	// billboard width/height the client passes to drawSprite.
	Sprites      [12]int
	HairColour   int
	TopColour    int
	BottomColour int
	SkinColour   int
	Camera1      int
	Camera2      int
}

// ItemDef is the static definition of an item type.
//
// Loaded from ItemDefs.json.
type ItemDef struct {
	ID            int
	Name          string
	Description   string
	Command       string // e.g., "Eat", "Drink", "Wield"
	IsMembersOnly bool
	IsStackable   bool
	IsUntradable  bool
	IsWearable    bool
	// BasePrice is the item's catalogue base value (gp). Shop buy/sell
	// prices are derived from it via the shop's price modifiers — see
	// world.ShopState.BuyPrice.
	BasePrice int
	// AppearanceID is the inventory-icon sprite index (OpenRSC ItemDef.spriteID,
	// from ItemDefs.json "appearanceID"). The 2D icon is sprite (spriteItem +
	// AppearanceID) in Authentic_Sprites.orsc. Empirically: item 0 (Iron Mace)
	// AppearanceID=117 -> sprite 2267 = the mace icon.
	//
	// The LOW BYTE of AppearanceID is also what the opcode-234 appearance
	// packet carries per worn slot (PlayerRecord.EquipBySlot), so it
	// doubles as the key for resolving another player's worn equipment
	// back to an item — see Facts.ResolveWorn / facts/worn.go.
	AppearanceID int
	// WearSlot is the equip slot a wearable item occupies (ItemDefs.json
	// "wearSlot"); it matches event.EquipSlot* numbering (3=shield,
	// 4=weapon, 5=hat/helmet, 6=body, 7=legs, 8=gloves, 9=boots,
	// 10=amulet, 11=cape). Meaningful only when IsWearable.
	WearSlot int

	// Combat bonuses this item contributes when worn (ItemDefs.json).
	// These are what the in-game equipment screen sums into the player's
	// "equipment status"; Host.SelfEquipmentBonuses totals them over the
	// currently-wielded items.
	ArmourBonus      int
	WeaponAimBonus   int
	WeaponPowerBonus int
	MagicBonus       int
	PrayerBonus      int
}

// IsOpenableBarrier reports whether this scenery def is a door/gate the host
// can open IN PLACE: the name contains "gate"/"door" and a command is
// open/close — the server's own dispatch rule (DoorAction.blockObjectAction
// matches by name + command, NOT by type). 67 blocking defs match, including
// the wooden-gate family (57/60/137/138/254/346 — Plutonium's GATES table),
// castle double doors (64/142), and the Al-Kharid toll gate (180). Type stays
// the COLLISION rule (a closed type-2 barrier blocks movement until opened);
// this is the OPENABILITY rule — what the traversal flow may interact-open and
// what the worldmap Oracle treats as crossable rather than a component cut.
func (d *SceneryDef) IsOpenableBarrier() bool {
	if d == nil {
		return false
	}
	name := strings.ToLower(d.Name)
	if !strings.Contains(name, "gate") && !strings.Contains(name, "door") {
		return false
	}
	c1 := strings.ToLower(d.Command1)
	c2 := strings.ToLower(d.Command2)
	return c1 == "open" || c1 == "close" || c2 == "open" || c2 == "close"
}

// TileDef is one ground-overlay definition from TileDef.xml. The overlay
// byte in the landscape is a 1-BASED index into this table (overlay-1),
// with the legacy 250 value remapped to 2 (water). ObjectType != 0 means
// the overlay terrain is impassable (water, lava, void) — the authoritative
// replacement for the old hardcoded "overlay 2 or 11 blocks" rule.
type TileDef struct {
	ID         int // 0-based table index; overlay value = ID+1
	Colour     int
	Unknown    int
	ObjectType int // != 0 => terrain blocks movement
}

// OverlayBlocks reports whether a landscape ground-overlay value marks
// impassable terrain, per TileDef.xml (overlay-1 indexing, 250→2 remap —
// mirrors Plutonium world.go:189-195 and the server's WorldLoader). Falls
// back to the legacy well-known water/lava values when the table isn't
// loaded (hand-built test fixtures).
func (f *Facts) OverlayBlocks(overlay int) bool {
	if overlay == 250 {
		overlay = 2
	}
	if overlay <= 0 {
		return false
	}
	if len(f.TileDefs) == 0 {
		return overlay == 2 || overlay == 11 // legacy fallback
	}
	if overlay-1 >= len(f.TileDefs) {
		return false
	}
	return f.TileDefs[overlay-1].ObjectType != 0
}

// ScenerLoc is an instance of a scenery type at a specific tile.
type SceneryLoc struct {
	DefID     int // joins to SceneryDef.ID
	X         int
	Y         int
	Direction int
}

// BoundaryLoc is an instance of a boundary at a specific tile + edge.
type BoundaryLoc struct {
	DefID     int
	X         int
	Y         int
	Direction int // 0 = horizontal (blocks N-S), 1 = vertical (blocks E-W), etc.
}

// NpcLoc is an NPC spawn point + roaming bounds.
type NpcLoc struct {
	DefID          int
	StartX, StartY int
	MinX, MinY     int
	MaxX, MaxY     int
}

// GroundItemLoc is an item that spawns at a fixed location.
type GroundItemLoc struct {
	DefID     int
	X         int
	Y         int
	RespawnMs int // respawn time in milliseconds; 0 if never respawns
}
