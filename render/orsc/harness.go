package orsc

// harness.go — the render entry point. RenderBridge opens the SAME OpenRSC
// .orsc assets the existing render/ package consumes, builds the terrain window
// centred on a world tile via the World port (world.go's BuildTerrain), drives
// the faithful three/ pipeline (Scene.setCamera -> Scene.endScene), and returns
// PNG bytes. It is the hook we call to render the 104,655 bridge twice — once
// through this port and once through render/ — and diff the two.
//
// The data plumbing mirrors render/render.go's RenderView (assets via
// assets.OpenArchive/.Get + assets.DecodeModel, landscape via
// pathfind.OpenLandscape(...).Tile, sprites via assets.OpenSprites(...).Sprite),
// but the geometry/camera/raster ALGORITHM is the three/ one: terrain built by
// world.go's BuildTerrain (World.generateLandscapeModel) as a *Model, added with
// Scene.AddModel, projected+rasterised by Scene.Render (RSModel.rotate1024 +
// Scene.endScene).
//
// COORDINATE NOTE: world.go builds terrain in WINDOW-LOCAL world units — vertices
// at (x*128, y, z*128) for window tile x,z in 0..95 (World.java:549), and its
// getElevation takes those same window-local fine coords. The camera therefore
// works in window-local units too: the requested world tile lands at the window
// centre (its SW corner is baseX=worldX-48, baseY=worldY-48), so the look-at is
// the local-centre tile.

import (
	"fmt"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

// openRSCRoot is the OpenRSC tree root facts.DefaultSources expects (the parent
// of server/conf/server/defs); the asset paths above live under its Client_Base.
const openRSCRoot = "/Users/flint/Code/openrsc"

// Canonical asset locations (the same .orsc files render/ opens; the mudclient's
// Cache/video directory).
const (
	landscapePath = "/Users/flint/Code/openrsc/Client_Base/Cache/video/Authentic_Landscape.orsc"
	modelsPath    = "/Users/flint/Code/openrsc/Client_Base/Cache/video/models.orsc"
	spritesPath   = "/Users/flint/Code/openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc"
)

// Render-window / camera constants, mirrored from the three/ + mudclient drivers.
const (
	// tileWorldUnits: one landscape tile is 128 world units on a side. World.java
	// places terrain vertices at (x*128, -elevation, z*128) (World.java:549,711).
	tileWorldUnits = 128

	// worldWindowTiles: the terrain window side length, in tiles. world.go's
	// BuildTerrain emits a fixed 96x96 window (World.loadSection's 2x2 sectors);
	// getElevation's interpolation domain is 0..95 (World.java:1180).
	worldWindowTiles = 96

	// windowCentreTile: the window-local tile the requested world tile maps to
	// (the SW corner is baseX = worldX - windowCentreTile), so the look-at sits at
	// the dead centre of the window.
	windowCentreTile = worldWindowTiles / 2

	// cameraPitch is mudclient.DEFAULT_CAMERA_PITCH (mudclient.java:312 == 912):
	// the fixed downward tilt of the spectator/overview camera
	// (mudclient.java:15169 setCamera(..., 912, rotation, 0, ...)).
	cameraPitch = 912

	// fogLandscape is Scene.fogLandscapeDistance for an open overview: the
	// spectator viewport sets 10000 (mudclient.java:15164) so distant terrain
	// stays visible rather than fogging out.
	fogLandscape = 10000
)

// Scene-pool sizes, mirroring the Scene ctor call the mudclient makes for the 3D
// world (Scene.java:81 new Scene(graphics, maxModels, maxPolygons,
// maxSpriteFaces)). Generous bounds for a full terrain window (world.go's
// accumulator is 18688 verts/faces).
const (
	maxModels      = 1024
	maxPolygons    = 65535
	maxSpriteFaces = 5000
)

// RenderBridge renders the world centred on world tile (x, y) on the given plane
// and returns PNG bytes. rot is the camera yaw in 1024-step units (0..1023),
// zoom is the orbit offset BEFORE the *2 the camera applies (mudclient passes
// zoom_distance*2; mudclient.java:15169), and w/h are the output image size.
//
// Pipeline (mirrors mudclient.renderLoginScreenViewports +
// mudclient.drawScene -> Scene.endScene):
//
//  1. Open the landscape + models + sprites archives.
//  2. world.BuildTerrain centred on (x, y): emit the 96x96 terrain window as a
//     *Model and Scene.AddModel it.
//  3. Scene.SetCamera(centerX, -elevation, centerZ, pitch, rot, 0, zoom*2) — the
//     spectator camera (mudclient.java:15169), in window-local world units.
//  4. Set fogLandscapeDistance / fogEntityDistance for the frame.
//  5. Scene.Render() (Scene.endScene): project all models, sort+resolve Polygons,
//     scan + Shader-fill into Surface.Pixels.
//  6. Surface.toImage() -> PNG bytes.
func RenderBridge(x, y, plane, rot, zoom, w, h int) ([]byte, error) {
	// ---- 1. open assets (the same .orsc files render/ reads) ----
	land, err := pathfind.OpenLandscape(landscapePath)
	if err != nil {
		return nil, fmt.Errorf("orsc: open landscape: %w", err)
	}
	defer land.Close()

	// The models + sprites archives are opened so the World/scenery + texture
	// passes can read them; terrain itself needs only the landscape. They are part
	// of the contract plumbing (assets.OpenArchive / assets.OpenSprites) and are
	// validated here so a missing archive fails fast at the entry point.
	models, err := assets.OpenArchive(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("orsc: open models archive: %w", err)
	}

	sprites, err := assets.OpenSprites(spritesPath)
	if err != nil {
		return nil, fmt.Errorf("orsc: open sprites archive: %w", err)
	}
	defer sprites.Close()

	// ---- 2. surface + scene ----
	surf := NewSurface(w, h)
	scene := NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)

	// Window SW corner so the requested tile lands at the centre of the window.
	baseX := x - windowCentreTile
	baseY := y - windowCentreTile

	// Terrain window as one accumulator *Model (World.generateLandscapeModel,
	// world.go BuildTerrain). Added to the scene's render list.
	terrain := BuildTerrain(land, baseX, baseY, plane)
	scene.AddModel(terrain)

	// Load the authentic texture set (water/planks/lava/…) from Authentic_Sprites
	// into the scene's resource database, so textured terrain faces (fill >= 0)
	// have a texel buffer to sample (mirrors mudclient.loadTexturesAuthentic).
	scene.LoadTexturesFromArchive(sprites, textureSlots)

	// Scenery (trees, railings, wells, signs, …) from SceneryLocs — the
	// World.addLoginScreenModels equivalent, driven by the OpenRSC server defs.
	// Missing defs/models are skipped; a facts-load failure just omits scenery.
	if f, ferr := facts.Load(facts.DefaultSources(openRSCRoot)); ferr == nil {
		placeSceneryModels(scene, models, land, f, baseX, baseY, plane)
		// Stacked walls + roofs for every visible story (ground + upper floors),
		// threaded through one accumulating elevation cache so the 2nd/3rd stories
		// lift onto the lower floors' wall-tops (World.loadSections, stories.go).
		for _, m := range BuildStories(land, f, baseX, baseY, plane) {
			scene.AddModel(m)
		}
		addDemoEntities(scene, land, f, baseX, baseY, plane)
	}

	// ---- 3. camera (mudclient.java:15169 spectator setCamera), window-local ----
	// The requested tile sits at window-local (windowCentreTile, windowCentreTile);
	// its CENTRE in world units is tile*128 + 64. Ground height comes from world.go's
	// getElevation at that same window-local fine point — the function the terrain
	// vertices were placed by, so camera and ground stay locked.
	centerX := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2)
	centerZ := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2)
	elev := elevationOf(land, baseX, baseY, plane, centerX, centerZ)
	centerY := -elev

	scene.fogLandscapeDistance = fogLandscape
	scene.fogEntityDistance = fogLandscape

	// setCamera(centerX, centerY, centerZ, xRot=pitch, yRot=rot, zRot=roll, offset)
	// — the exact arg order Scene.setCamera takes (Scene.java:2930). offset is the
	// orbit distance; mudclient passes zoom_distance*2.
	scene.SetCamera(centerX, centerY, centerZ, cameraPitch, int32(rot), 0, int32(zoom)*2)

	// ---- 5. render ----
	scene.Render()

	// ---- 6. encode ----
	return surf.toImage()
}

// elevationOf returns the window-local bilinear ground elevation at fine world
// coords (x, z) for the 96x96 window whose SW corner is (baseX, baseY) — the same
// value world.go's terrain build placed its vertices at, so the camera height
// tracks the deck/terrain exactly. It is a thin wrapper around the World port's
// (unexported) terrainBuilder.getElevation (World.getElevation, World.java:1173)
// so the harness shares the EXACT elevation math the terrain mesh used (no
// second, drifting copy).
func elevationOf(land *pathfind.Landscape, baseX, baseY, plane int, x, z int32) int32 {
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}
	return b.getElevation(x, z)
}

// addDemoEntities places a couple of validation billboards (the local player at
// the centre tile + a goblin two tiles east) so the entity layer can be seen
// without the live spectator. It reuses render/'s verified entity compositing for
// the pixels and the orsc Scene's AddEntity for placement + depth-sorted blit.
// (In the live viewport the spectator will drive real per-actor entities.)
func addDemoEntities(scene *Scene, land *pathfind.Landscape, f *facts.Facts, baseX, baseY, plane int) {
	add := func(cs *render.CompositeSprite, bw, bh, lx, lz int) {
		if cs == nil || cs.W <= 0 || cs.H <= 0 {
			return
		}
		wx := int32(lx*tileWorldUnits + tileWorldUnits/2)
		wz := int32(lz*tileWorldUnits + tileWorldUnits/2)
		wy := -elevationOf(land, baseX, baseY, plane, wx, wz)
		scene.AddEntity(wx, wy, wz, int32(bw), int32(bh), cs.Pix, cs.Opaque, cs.W, cs.H, cs.Flip)
	}
	// Local player at the centre tile (facing south, standing).
	add(render.CompositePlayerSprite(0, 0), 145, 220, windowCentreTile, windowCentreTile)
	// A goblin two tiles east, if present in the defs.
	for id, d := range f.NpcDefs {
		if d.Name == "Goblin" {
			w, h := render.NPCBillboardSize(f, id)
			add(render.CompositeNPCSprite(f, id, 0, 0), w, h, windowCentreTile+2, windowCentreTile)
			break
		}
	}
}
