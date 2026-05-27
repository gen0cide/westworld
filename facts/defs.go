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
// Loaded from DoorDef.xml.
type BoundaryDef struct {
	ID           int
	Name         string // e.g., "Wall", "Door", "Fence"
	Description  string
	Command1     string // typically "Open" for doors, blank for walls
	Command2     string // typically "Examine"
	Unwalkable   bool   // blocks movement
	ModelVisible int    // sprite when not opened
	ModelOpened  int    // sprite when opened
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
