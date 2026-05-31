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
// For NPCs, NpcID is the config85.jag NPC id (NOT the OpenRSC NpcDefs.json id):
// it selects the body-part sprite layers + clothing colours used to composite
// the 2D billboard (see entitysprite.go). NpcID < 0 (or a missing/failed
// composite) makes the renderer fall back to the 3D-cross billboard.
type Entity struct {
	X, Y  int
	Kind  EntityKind
	NpcID int // config85.jag npc id (sprite source); -1 / 0 if unknown

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

// entity billboard dimensions (world units). RSC actors stand ~2 tiles tall;
// 128 units == one tile. A figure ~1.6 tiles tall and ~0.6 tiles wide reads as
// a person at the renderer's scale without looking like a wall.
const (
	entityHeight   = 200
	entityHalfWide = 38

	// entityShade is the FIXED flat shade index every billboard face is lit at.
	// The ramp is ramp[255-j] = colour*(j*j)/65536, so a LOW index is BRIGHT
	// (index 0 = full colour, 255 = black). 40 reads as near-full figure colour
	// — the same bright-fixed recipe water/shore quads use to avoid going black.
	entityShade = 40
)

// entity fill colours (method305-encoded flat 5:5:5). Players read as a warm
// tunic blue; NPCs as a muted brown/tan so the two are distinguishable at a
// glance. Flat (non-gouraud) shaded so a vertical billboard never gouraud-darkens
// to black like the early terrain cliffs did.
var (
	entityPlayerColour = method305(70, 90, 200)  // blue tunic
	entityNPCColour    = method305(170, 140, 95) // tan/brown
	entitySelfColour   = method305(220, 200, 60) // gold (host marker)
)

func entityColour(k EntityKind) int32 {
	switch k {
	case EntityPlayer:
		return entityPlayerColour
	case EntitySelf:
		return entitySelfColour
	default:
		return entityNPCColour
	}
}

// BuildEntities assembles one camera-agnostic "cross" billboard per in-window
// entity into a single GameModel: two perpendicular vertical quads (an X in
// plan view) rising entityHeight world-units off the terrain at the entity's
// tile centre. The cross means the figure stays visible (and roughly the same
// silhouette) from any camera yaw — a single quad would vanish edge-on. The
// quads are flat-shaded and fed through the same painter-sorted pipeline as
// scenery, so they occlude correctly behind walls/scenery and depth-scale with
// the perspective divide. baseX/baseY anchor the 96x96 window; heights is the
// flattened terrain grid (TerrainHeights) so each figure's feet sit on the
// ground. Returns nil if no entity falls inside the window.
func BuildEntities(ents []Entity, baseX, baseY, plane int, heights [][]int32) *GameModel {
	n := terrainSize

	type placed struct {
		lx, ly int
		colour int32
	}
	var ps []placed
	for _, e := range ents {
		lx := e.X - baseX
		ly := e.Y - baseY
		if lx < 0 || lx >= n || ly < 0 || ly >= n {
			continue
		}
		ps = append(ps, placed{lx, ly, entityColour(e.Kind)})
	}
	if len(ps) == 0 {
		return nil
	}

	// 2 quads per entity, 4 verts each.
	g := NewGameModel(len(ps)*8, len(ps)*2)
	for _, p := range ps {
		// tile centre in world units; feet at the terrain height there.
		cx := int32(p.lx)*128 + 64
		cz := int32(p.ly)*128 + 64
		foot := -heights[p.lx][p.ly]
		top := foot - entityHeight

		// quad A spans along +X (east-west facing pair)
		ax0 := g.AddVertex(cx-entityHalfWide, foot, cz)
		ax1 := g.AddVertex(cx-entityHalfWide, top, cz)
		ax2 := g.AddVertex(cx+entityHalfWide, top, cz)
		ax3 := g.AddVertex(cx+entityHalfWide, foot, cz)
		// quad B spans along +Z (north-south facing pair), perpendicular
		bz0 := g.AddVertex(cx, foot, cz-entityHalfWide)
		bz1 := g.AddVertex(cx, top, cz-entityHalfWide)
		bz2 := g.AddVertex(cx, top, cz+entityHalfWide)
		bz3 := g.AddVertex(cx, foot, cz+entityHalfWide)

		// FIXED-intensity faces (like water/shore): a billboard's normal is
		// near-vertical and ±X/±Z, so the directional light's dot product would
		// otherwise drive these faces to a near-black slab. Pinning a bright
		// shade index makes relight()/light() leave them at full figure colour
		// from every camera yaw. Both quads of the cross, both sides opaque.
		g.AddFixedFace([]int{ax0, ax1, ax2, ax3}, p.colour, p.colour, entityShade)
		g.AddFixedFace([]int{bz0, bz1, bz2, bz3}, p.colour, p.colour, entityShade)
	}
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
