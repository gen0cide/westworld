package render

// render is the sprite-compositing + view-data layer shared by the live
// renderer (render/orsc, the faithful OpenRSC three/ port). It owns the entity /
// item sprite compositing (entitysprite.go, itemsprite.go, composite_export.go,
// animdefs.go, sprites.go) and the View/Entity value types the renderer + cradle
// pass around. The classic 3D software renderer that used to live here (Scene,
// GameModel, the rasterizer, terrain/roof/boundary/scenery builders) was removed
// once render/orsc became the renderer — see git history for that engine.

// View are the render parameters the live spectator passes per frame. The camera
// fields (Rotation / Zoom / W / H) describe the shot; the Entities + Self* fields
// the actors; the dynamic-state fields the live server overlays.
type View struct {
	X, Y     int // host world TILE coords
	Plane    int
	Rotation int // 0..255 RSC camera angle (yaw = rotation*4)
	Zoom     int // default ~600 -> distance = zoom*2
	W, H     int // output image size

	// AnimFrame advances the model-swap animation (fire/torch/fireplace) so the
	// live -spectate viewport flickers. A static render leaves it 0 (frame 1).
	AnimFrame int

	// Entities are the actors (players/NPCs) the host perceives, in ABSOLUTE
	// world-tile coords. Drawn as depth-scaled billboards. Optional: leave nil
	// to render only the static world (terrain + scenery + boundaries).
	Entities []Entity

	// NoSelf suppresses drawing the local player ("bernard") at his own tile.
	// Default OFF — RSC always shows the host in the third-person scene.
	NoSelf bool

	// SelfHeading is the host's own 8-way facing (0=N..7=NW). Added to the camera
	// term to choose which side of the local player's sprite to draw.
	SelfHeading int

	// Self appearance: the host's own worn-equipment sprite array + colour indices,
	// mirrored from world.Self. SelfHasEquip selects the real-appearance composite;
	// when false the host renders the default human.
	SelfEquipSprites  [12]int
	SelfHairColour    int
	SelfTopColour     int
	SelfTrouserColour int
	SelfSkinColour    int
	SelfHasEquip      bool

	// SelfOffX / SelfOffZ are the host's own sub-tile WORLD-unit offset (128 ==
	// one tile) while he glides between tiles. They shift BOTH the camera AND the
	// host's own billboard by the same amount, so bernard stays centred on screen
	// while the world scrolls smoothly under him. SelfStepPhase/SelfMoving drive
	// his leg-cycle frame, as for other actors.
	SelfOffX, SelfOffZ int
	SelfStepPhase      int
	SelfMoving         bool

	// --- live dynamic server state (all OPTIONAL / nil-guarded) ---

	// BoundaryRemoved reports whether the wall/door edge at absolute (x, y, dir)
	// has been removed by a live server update (door opened, web cut). dir uses
	// the authentic createModel convention (mudclient.java:6769-6780):
	//   0 = (x,y)..(x+1,y) east-west; 1 = (x,y)..(x,y+1) north-south;
	//   2 = '\' diagonal; 3 = '/' diagonal. nil => static walls render unchanged.
	BoundaryRemoved func(x, y, dir int) bool

	// DynamicScenery are live GameObjects the host perceives that are NOT in the
	// static landscape (lit fires, regrown/changed trees, etc.).
	DynamicScenery []DynamicSceneryItem

	// SceneryRemoved reports whether the static scenery at absolute (x, y) was
	// cleared by a server removal (mined rock, burned-out fire). nil => no
	// suppression.
	SceneryRemoved func(x, y int) bool

	// GroundItems are dropped items the host perceives, drawn as small ground
	// markers within the window.
	GroundItems []GroundItemMarker
}

// DynamicSceneryItem is one live GameObject the host perceives, in absolute
// world-tile coords. ID joins to facts.SceneryDef.ID. Direction is the object's
// heading (0 when unknown).
type DynamicSceneryItem struct {
	X, Y      int
	ID        int
	Direction int
}

// GroundItemMarker is one dropped item the host perceives, in absolute world-tile
// coords. ItemID drives the rendered inventory icon (compositeItem).
type GroundItemMarker struct {
	X, Y   int
	ItemID int
}
