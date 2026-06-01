package render

// EntityKind distinguishes the billboard's colour/size by who it is.
type EntityKind int

const (
	EntityNPC EntityKind = iota
	EntityPlayer
	EntitySelf // the host itself (rarely drawn — camera sits on it)
)

// Entity is one billboarded actor the host perceives, in ABSOLUTE world-tile
// coords (the same space facts.NpcLocs and world.NpcRecord/PlayerRecord use —
// the live decoder converts relative offsets to absolute). The renderer turns
// each into a depth-scaled standing figure.
//
// For NPCs, NpcID is the OpenRSC npc id (the facts.NpcDef key, which IS the id
// the server sends): it selects the body-part sprite layers + clothing colours
// (NpcDef.Sprites/*Colour) used to composite the 2D billboard from
// Authentic_Sprites.orsc (see entitysprite.go). NpcID < 0 (or a missing/failed
// composite) makes the renderer fall back to the 3D-cross billboard.
type Entity struct {
	X, Y  int
	Kind  EntityKind
	NpcID int // OpenRSC npc id (facts.NpcDef key + sprite source); -1 / 0 if unknown

	// Heading is the actor's server-reported facing as an 8-way RSC direction
	// (animationCurrent: 0=N .. 7=NW, the value drawNpc/drawPlayer add to the
	// camera term). Combined with the camera rotation it selects which side of
	// the sprite to draw: facing = (Heading + (cameraRotation+16)/32) & 7. In
	// Phase 3a this stays 0 everywhere (no server-heading plumbing yet) — the
	// facing still varies correctly as the camera pans because the camera term
	// changes; Phase 3b carries the real per-entity heading here.
	Heading int

	// Player appearance (EntityPlayer / EntitySelf only). EquipSprites is the
	// per-slot worn-equipment SPRITE ids (indexed by event.EquipSlot* == the
	// layer index the npcAnimationArray draw order uses); a value v>0 means
	// the layer's animation id is v-1 (mudclient.java:2963 equippedItem[layer]-1).
	// HairColour/TopColour/TrouserColour/SkinColour are the appearance colour
	// indices dyeing those layers. HasEquip selects the real-appearance
	// composite (compositePlayerAppearance) over the default-human one; when
	// false the player renders the default human (compositePlayer).
	EquipSprites  [12]int
	HairColour    int
	TopColour     int
	TrouserColour int
	SkinColour    int
	HasEquip      bool

	// Index is the server's stable per-actor index (NpcRecord.Index /
	// PlayerRecord.Index). The renderer ignores it; the spectator uses it to key
	// per-entity motion interpolation across frames.
	Index int
	// OffX / OffZ are the sub-tile WORLD-unit render offsets (128 == one tile)
	// applied to this actor's foot point so it GLIDES between tiles instead of
	// snapping each server tick. OffX shifts along world X (tile-X axis), OffZ
	// along world Z (tile-Y axis). Both 0 = drawn exactly on the tile (the static
	// behaviour). Set by the spectator's motion interpolation; a one-shot render
	// leaves them 0.
	OffX, OffZ int
	// StepPhase is the walk-cycle index (npcWalkModel = {0,1,2,1}) for the
	// leg-cycle frame while Moving; 0 / Moving=false renders the standing frame.
	StepPhase int
	Moving    bool
}
