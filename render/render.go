package render

import (
	"os"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// Bundle holds the decoded asset archives + caches the renderer consumes.
type Bundle struct {
	Models *ModelCache
}

// OpenBundle opens the OpenRSC models.orsc archive.
func OpenBundle(modelsPath string) (*Bundle, error) {
	arc, err := assets.OpenArchive(modelsPath)
	if err != nil {
		return nil, err
	}
	return &Bundle{Models: NewModelCache(arc)}, nil
}

// View are the render parameters.
type View struct {
	X, Y     int // host world TILE coords
	Plane    int
	Rotation int // 0..255 RSC camera angle (yaw = rotation*4)
	Zoom     int // default ~600 -> distance = zoom*2
	W, H     int // output image size

	// Entities are the actors (players/NPCs) the host perceives, in ABSOLUTE
	// world-tile coords. Drawn as depth-scaled billboards. Optional: leave nil
	// to render only the static world (terrain + scenery + boundaries).
	Entities []Entity

	// DrawSelf, when not explicitly disabled, draws the local player ("bernard")
	// as a standing-frame human billboard at the centre of his own view (his own
	// tile). RSC always shows the host in the third-person scene. Defaults to ON;
	// set NoSelf to suppress it.
	NoSelf bool
}

// RenderView assembles the terrain + nearby scenery around the host tile and
// renders a PNG of what that host sees. This is the Phase-1 MVP entrypoint.
func RenderView(land *pathfind.Landscape, f *facts.Facts, b *Bundle, v View) ([]byte, error) {
	if v.Zoom == 0 {
		v.Zoom = 750
	}
	if v.W == 0 {
		v.W = 512
	}
	if v.H == 0 {
		v.H = 334
	}

	// CENTER the 96x96 terrain window ON the host tile. The earlier code
	// mirrored pathfind.BuildGrid's midRegion centring (baseX = mid*48-48),
	// which only guarantees the host lands somewhere in 24..72 of the window —
	// frequently near a window edge, so the camera looked across a cliff
	// edge-on (the dark-wedge artifact). Since pathfind.Landscape.Tile lazily
	// loads ANY sector, we are not bound to sector-aligned windows: put the
	// host at local (48,48), the dead centre, every time.
	baseX := v.X - terrainSize/2
	baseY := v.Y - terrainSize/2

	sc := &Scene{}
	// Build the flattened terrain-height grid once; both the terrain mesh and
	// the boundary feet anchor to it.
	heights := TerrainHeights(land, baseX, baseY, v.Plane)
	if os.Getenv("RENDER_NO_TERRAIN") == "" {
		terrain := BuildTerrain(land, baseX, baseY, v.Plane)
		sc.Add(terrain)
	}

	// boundaries (walls/fences/doors) within the window
	if os.Getenv("RENDER_NO_BOUNDARIES") == "" && f != nil {
		if bd := BuildBoundaries(f, land, baseX, baseY, v.Plane, heights); bd != nil {
			sc.Add(bd)
		}
	}

	// entity billboards (players/NPCs) the host perceives. RSC characters are 2D
	// sprites: the depth-scaled sprite blit happens AFTER RenderTo (see
	// DrawEntitySprites below). The 3D-cross billboard is kept ONLY as a fallback
	// for actors whose sprite composite fails (archives missing / unknown npc id)
	// so they still appear as *something* — successful sprites are not crossed.
	if os.Getenv("RENDER_NO_ENTITIES") == "" && len(v.Entities) > 0 {
		var fallback []Entity
		for _, e := range v.Entities {
			var cs *CompositeSprite
			if e.Kind == EntityNPC {
				cs = compositeNPC(e.NpcID)
			} else {
				cs = compositePlayer()
			}
			if cs == nil {
				fallback = append(fallback, e)
			}
		}
		if len(fallback) > 0 {
			if ents := BuildEntities(fallback, baseX, baseY, v.Plane, heights); ents != nil {
				sc.Add(ents)
			}
		}
	}

	// place scenery within the window
	if os.Getenv("RENDER_NO_SCENERY") == "" && f != nil && b != nil {
		for _, loc := range f.SceneryLocs {
			if loc.X < baseX || loc.X >= baseX+terrainSize ||
				loc.Y < baseY || loc.Y >= baseY+terrainSize {
				continue
			}
			if g := PlaceScenery(b.Models, f, land, baseX, baseY, v.Plane, loc); g != nil {
				sc.Add(g)
			}
		}
	}

	// camera at the host tile, looking with the requested yaw.
	localX := int32(v.X-baseX)*128 + 64
	localZ := int32(v.Y-baseY)*128 + 64
	t := land.Tile(v.X, v.Y, v.Plane)
	elev := int32(t.GroundElevation) * 3
	// setCamera(x, z=height, y, pitch=912, yaw=rotation*4, roll=0, distance)
	sc.Cam = SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(v.Zoom)*2)

	surf := NewSurface(v.W, v.H)
	surf.Clear(skyColour)
	sc.RenderTo(surf, v.W/2, v.H/2)

	// 2D character-sprite pass: composite + blit nearby NPCs/players and the
	// local player as depth-scaled standing billboards over the rendered world.
	// RSC characters are sprites, not 3D models; this replaces the 3D-cross
	// billboards for any actor whose sprite composites successfully (the cross is
	// kept only as a fallback inside BuildEntities when the archives are missing).
	if os.Getenv("RENDER_NO_ENTITIES") == "" {
		DrawEntitySprites(surf, sc.Cam, v, v.Entities, baseX, baseY, heights)
	}

	return surf.PNG()
}

// DrawEntitySprites composites each entity (and the local player) into a 2D
// standing-frame billboard and blits it depth-scaled over the rendered world.
//
// Projection mirrors GameModel.project() EXACTLY, including the axis swap
// Scene.RenderTo performs before projecting (CameraYaw <-> CameraRoll), so a
// character lands at the same screen position the 3D pipeline would place a
// model at that tile. For each actor:
//   - foot world point = (lx*128+64, -heights[lx][ly], ly*128+64)
//   - transform into camera space (CameraX/Y/Z subtract, then yaw/roll/pitch)
//   - skip if camZ < clipNear
//   - viewX = (x<<viewDist)/z, viewY = (y<<viewDist)/z
//   - screen feet at (W/2 + viewX, H/2 + viewY)
//   - screen size = (worldW<<viewDist)/z, (worldH<<viewDist)/z
//   - blit scaled, top-left (sx-screenW/2, feetY-screenH)
//
// The local player is drawn last (closest, at the camera tile) unless v.NoSelf.
func DrawEntitySprites(surf *Surface, cam Camera, v View, ents []Entity, baseX, baseY int, heights [][]int32) {
	// Axis swap at the call site (Scene.RenderTo): project is called with
	// yaw/roll bound swapped. Replicate so screen placement matches the 3D pass.
	cam.CameraYaw, cam.CameraRoll = cam.CameraRoll, cam.CameraYaw

	n := terrainSize
	cx := v.W / 2
	cy := v.H / 2

	blit := func(lx, ly int, worldW, worldH int, cs *CompositeSprite) {
		if cs == nil {
			return
		}
		if lx < 0 || lx >= n || ly < 0 || ly >= n {
			return
		}
		wx := int32(lx)*128 + 64
		wz := int32(ly)*128 + 64
		wy := -heights[lx][ly] // feet on the terrain

		x := wx - cam.CameraX
		y := wy - cam.CameraY
		z := wz - cam.CameraZ
		if cam.CameraYaw != 0 {
			ys := sine11[cam.CameraYaw]
			yc := sine11[cam.CameraYaw+1024]
			X := (y*ys + x*yc) >> 15
			y = (y*yc - x*ys) >> 15
			x = X
		}
		if cam.CameraRoll != 0 {
			rs := sine11[cam.CameraRoll]
			rc := sine11[cam.CameraRoll+1024]
			X := (z*rs + x*rc) >> 15
			z = (z*rc - x*rs) >> 15
			x = X
		}
		if cam.CameraPitch != 0 {
			ps := sine11[cam.CameraPitch]
			pc := sine11[cam.CameraPitch+1024]
			Y := (y*pc - z*ps) >> 15
			z = (y*ps + z*pc) >> 15
			y = Y
		}
		if z < clipNear {
			return
		}
		viewX := (x << uint(viewDist)) / z
		viewY := (y << uint(viewDist)) / z
		sx := int32(cx) + viewX
		feetY := int32(cy) + viewY
		screenW := (int32(worldW) << uint(viewDist)) / z
		screenH := (int32(worldH) << uint(viewDist)) / z
		if screenW <= 0 || screenH <= 0 {
			return
		}
		topLeftX := int(sx - screenW/2)
		topLeftY := int(feetY - screenH)
		surf.BlitSpriteScaled(cs, topLeftX, topLeftY, int(screenW), int(screenH))
	}

	for _, e := range ents {
		lx := e.X - baseX
		ly := e.Y - baseY
		var cs *CompositeSprite
		var worldW, worldH int
		switch e.Kind {
		case EntityNPC:
			cs = compositeNPC(e.NpcID)
			worldW, worldH = npcBillboardSize(e.NpcID)
		default: // EntityPlayer / EntitySelf rendered as the default human
			cs = compositePlayer()
			worldW, worldH = playerBillboardW, playerBillboardH
		}
		blit(lx, ly, worldW, worldH, cs)
	}

	// Local player at the centre of his own view (his own tile).
	if !v.NoSelf {
		if cs := compositePlayer(); cs != nil {
			blit(v.X-baseX, v.Y-baseY, playerBillboardW, playerBillboardH, cs)
		}
	}
}

// skyColour is a flat sky/background fill.
const skyColour = 0x6080a0
