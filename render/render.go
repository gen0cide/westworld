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
	if os.Getenv("RENDER_NO_TERRAIN") == "" {
		terrain := BuildTerrain(land, baseX, baseY, v.Plane)
		sc.Add(terrain)
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
	return surf.PNG()
}

// skyColour is a flat sky/background fill.
const skyColour = 0x6080a0
