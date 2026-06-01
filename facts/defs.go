package facts

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
	AppearanceID int
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
	DefID int
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
