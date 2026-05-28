// Package pathfind ports the original RSC client's local-region BFS
// pathfinder to Go. The bot uses this to compute a corner-compressed
// path through walls, doors, and scenery — the same way the real client
// does — instead of trusting the server to interpolate around
// obstacles.
//
// Ported from /Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/
// (World.java, Sector.java, Tile.java, CollisionFlag.java).
package pathfind

// CollisionFlag bits packed into each tile of the 96x96 grid. Cardinal
// WALL_* flags mark walls on individual tile edges; the *_BLOCKED
// values combine a wall with FULL_BLOCK so cardinal-and-diagonal
// movement checks share a single bitwise AND.
//
// Ported verbatim from orsc/graphics/three/CollisionFlag.java — keep
// the integer values identical so the BFS expansion logic in
// findPath() matches the original client.
const (
	WallNorth = 1
	WallEast  = 2
	WallSouth = 4
	WallWest  = 8

	wallNorthEast = WallNorth | WallEast
	wallNorthWest = WallNorth | WallWest
	wallSouthEast = WallSouth | WallEast
	wallSouthWest = WallSouth | WallWest

	FullBlockA = 16
	FullBlockB = 32
	FullBlockC = 64
	fullBlock  = FullBlockA | FullBlockB | FullBlockC

	// Object is the flag the client uses to mark a tile occupied by a
	// solid scenery object (a tree, sign, etc.) for higher-level checks.
	// findPath itself doesn't read it, but we set it so callers can
	// distinguish blocked-by-wall from blocked-by-object.
	Object = 128

	WestBlocked  = fullBlock | WallWest
	SouthBlocked = fullBlock | WallSouth
	NorthBlocked = fullBlock | WallNorth
	EastBlocked  = fullBlock | WallEast

	SouthEastBlocked = fullBlock | wallSouthEast
	SouthWestBlocked = fullBlock | wallSouthWest
	NorthEastBlocked = fullBlock | wallNorthEast
	NorthWestBlocked = fullBlock | wallNorthWest
)

// SOURCE_* are written into the BFS parent-direction array during
// path reconstruction. They share bit values with the WALL_* flags
// but live in a separate array, so the overlap is fine.
//
// During reconstruction the algorithm walks "back" toward the start —
// SourceEast set means the visited tile was reached *from* the tile
// to its east, so we step east to walk backwards.
const (
	sourceSouth = 1
	sourceWest  = 2
	sourceNorth = 4
	sourceEast  = 8

	sourceNorthEast = sourceNorth | sourceEast
	sourceNorthWest = sourceNorth | sourceWest
	sourceSouthEast = sourceSouth | sourceEast
	sourceSouthWest = sourceSouth | sourceWest

	visitedStart = 99 // initial marker the client writes for the start tile
)
