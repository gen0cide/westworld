package orsc

import (
	"sync"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

var (
	rvOnce    sync.Once
	rvModels  *assets.Archive
	rvSprites *assets.SpriteArchive
)

// RenderViewCached is RenderView with the models + sprites archives opened once
// (from the canonical Cache/video paths) and reused across calls — the entry the
// live spectator uses each frame (only land + facts + the View vary per frame).
func RenderViewCached(land *pathfind.Landscape, f *facts.Facts, v render.View) ([]byte, error) {
	rvOnce.Do(func() {
		rvModels, _ = assets.OpenArchive(modelsPath)
		rvSprites, _ = assets.OpenSprites(spritesPath)
	})
	return RenderView(land, f, rvModels, rvSprites, v)
}

// RenderView is the orsc analogue of render.RenderView: it renders the live
// spectator View through the faithful three/ engine — terrain + textures +
// scenery + walls/buildings + entity billboards (the host, other players, NPCs
// from v.Entities) — host-centred camera. Assets are passed in (opened once by
// the caller); textures are (re)loaded per call for now (a future optimisation
// caches them). Returns PNG bytes. This is what the spectator calls to drive the
// faithful renderer live (WESTWORLD_RENDERER=orsc).
func RenderView(land *pathfind.Landscape, f *facts.Facts, models *assets.Archive, sprites *assets.SpriteArchive, v render.View) ([]byte, error) {
	w, h := v.W, v.H
	if w == 0 {
		w = 512
	}
	if h == 0 {
		h = 334
	}
	baseX := v.X - windowCentreTile
	baseY := v.Y - windowCentreTile
	plane := v.Plane

	surf := NewSurface(w, h)
	scene := NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)

	scene.AddModel(BuildTerrain(land, baseX, baseY, plane))
	scene.LoadTexturesFromArchive(sprites, textureSlots)
	if f != nil {
		placeSceneryModels(scene, models, land, f, baseX, baseY, plane)
		// Stacked walls + roofs for every visible story (ground + upper floors),
		// threaded through one accumulating elevation cache so upper stories lift
		// onto the lower floors' wall-tops (World.loadSections, stories.go).
		for _, m := range BuildStories(land, f, baseX, baseY, plane) {
			scene.AddModel(m)
		}
		// Diagonal scenery objects / diagonal doors: the 48001..59999 DiagonalWalls
		// band the wall + scenery passes drop (World.addModels, diagobj.go). Without
		// this a diagonal door renders as zero geometry.
		for _, m := range BuildDiagonalObjects(models, land, f, baseX, baseY, plane) {
			scene.AddModel(m)
		}
		addViewEntities(scene, land, f, v, baseX, baseY, plane)
	}

	// Camera look-at = the host tile centre + his sub-tile glide offset (SelfOffX/Z),
	// so the world scrolls SMOOTHLY under the centred host over the 600ms walk tick
	// instead of jumping a whole tile each step (render.go:309-310). The host sprite
	// gets the same offset (addViewEntities), so he stays centred while the world glides.
	cx := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffX)
	cz := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffZ)
	elev := elevationOf(land, baseX, baseY, plane, cx, cz)
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	scene.fogLandscapeDistance = fogLandscape
	scene.fogEntityDistance = fogLandscape
	// render/ camera: yaw = v.Rotation*4 (v.Rotation is 0..255; orsc is 1024-step).
	scene.SetCamera(cx, -elev, cz, cameraPitch, int32(v.Rotation)*4, 0, int32(zoom)*2)
	scene.Render()
	return surf.toImage()
}

// addViewEntities composites + places the View's actors (host + other players +
// NPCs) as billboards, reusing render/'s verified compositing for the pixels.
// Facing = (entity heading + camTerm) & 7, camTerm = (v.Rotation+16)/32 — the
// exact render/ formula.
func addViewEntities(scene *Scene, land *pathfind.Landscape, f *facts.Facts, v render.View, baseX, baseY, plane int) {
	camTerm := (v.Rotation + 16) / 32

	// offX/offZ are the actor's sub-tile glide offset (world units, 128 == one tile)
	// so it slides smoothly between tiles over the walk tick instead of teleporting.
	place := func(cs *render.CompositeSprite, bw, bh, tileX, tileY, offX, offZ int) {
		if cs == nil || cs.W <= 0 || cs.H <= 0 {
			return
		}
		lx, lz := tileX-baseX, tileY-baseY
		if lx < 0 || lx >= worldWindowTiles || lz < 0 || lz >= worldWindowTiles {
			return
		}
		wx := int32(lx*tileWorldUnits + tileWorldUnits/2 + offX)
		wz := int32(lz*tileWorldUnits + tileWorldUnits/2 + offZ)
		wy := -elevationOf(land, baseX, baseY, plane, wx, wz)
		scene.AddEntity(wx, wy, wz, int32(bw), int32(bh), cs.Pix, cs.Opaque, cs.W, cs.H, cs.Flip)
	}

	// NPCs + other players (v.Entities).
	for _, e := range v.Entities {
		facing := (e.Heading + camTerm) & 7
		switch e.Kind {
		case render.EntityNPC:
			w, hh := render.NPCBillboardSize(f, e.NpcID)
			place(render.CompositeNPCSprite(f, e.NpcID, facing, e.StepPhase), w, hh, e.X, e.Y, e.OffX, e.OffZ)
		default: // EntityPlayer
			cs := playerComposite(e.HasEquip, e.EquipSprites, e.HairColour, e.TopColour, e.TrouserColour, e.SkinColour, facing, e.StepPhase)
			place(cs, 145, 220, e.X, e.Y, e.OffX, e.OffZ)
		}
	}

	// The host itself at the centre tile (unless suppressed). His glide offset matches
	// the camera's (SetCamera above), so he stays centred while the world scrolls.
	if !v.NoSelf {
		facing := (v.SelfHeading + camTerm) & 7
		cs := playerComposite(v.SelfHasEquip, v.SelfEquipSprites, v.SelfHairColour, v.SelfTopColour, v.SelfTrouserColour, v.SelfSkinColour, facing, v.SelfStepPhase)
		place(cs, 145, 220, v.X, v.Y, v.SelfOffX, v.SelfOffZ)
	}
}

// playerComposite picks the real-appearance billboard when equipment is known,
// else the default human (mirrors render/'s playerSprite). step is the walk-cycle
// frame (0 = standing; non-zero while gliding) — passing it animates the walk.
func playerComposite(hasEquip bool, equip [12]int, hair, top, trouser, skin, facing, step int) *render.CompositeSprite {
	if hasEquip {
		if cs := render.CompositePlayerAppearanceSprite(equip, hair, top, trouser, skin, facing, step); cs != nil {
			return cs
		}
	}
	return render.CompositePlayerSprite(facing, step)
}
