// Command cmd_genwalls authors the walls_* hunt fixtures: a flat single-tile
// scene (terrainSeed=0) with one wall encoded in a chosen grid, so we can
// isolate GO's wall builder per encoding (H/V/diag, def id, removed-edge).
//
//   go run ./testdata/rscdump/hunt/cmd_genwalls
//
// All fixtures share the single_tile_door camera/window so they line up with
// the existing self-test geometry. Walls are placed at local tile (8,9) to
// match the door self-test footprint.
package main

import (
	"log"
	"path/filepath"

	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	size  = 16
	baseX = 200
	baseY = 200
	// local tile to host the wall (matches single_tile_door footprint).
	wx = 8
	wy = 9
)

func idx(x, y int) int { return x*size + y }

func baseDump() *rscdump.Dump {
	n := size * size
	elev := make([]byte, n)
	col := make([]byte, n)
	for i := range col {
		col[i] = 70 // green grass palette index, matching single_tile_door
	}
	return &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: rscdump.Camera{
			X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0,
			Distance: 1500, ViewDist: 9, ClipNear: 5, ClipFar: 7000,
			ScreenW: 512, ScreenH: 334,
		},
		Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: 0, Size: size},
		Terrain: &rscdump.Terrain{
			Size:         size,
			Elevation:    elev,
			GroundColour: col,
			Overlay:      make([]byte, n),
			Roof:         make([]byte, n),
			WallH:        make([]byte, n),
			WallV:        make([]byte, n),
			WallDiag:     make([]int32, n),
			TerrainSeed:  0,
		},
		Self: &rscdump.Self{X: baseX + 8, Y: baseY + 8, NoSelf: true},
	}
}

func save(d *rscdump.Dump, name string) {
	p := filepath.Join("testdata", "rscdump", "hunt", name)
	if err := d.Save(p); err != nil {
		log.Fatalf("save %s: %v", name, err)
	}
	log.Printf("wrote %s", p)
}

func main() {
	// Control: no walls (baseline twin for structural diffs).
	none := baseDump()
	save(none, "walls_none.json")

	// V wall (east-west edge along X) at (8,9), def id 0 (1-based value 1 = "Wall").
	v := baseDump()
	v.Terrain.WallV[idx(wx, wy)] = 1
	save(v, "walls_v_id1.json")

	// H wall (north-south edge along Y) at (8,9), def id 0.
	h := baseDump()
	h.Terrain.WallH[idx(wx, wy)] = 1
	save(h, "walls_h_id1.json")

	// '\' diagonal wall, def id 0 (value 1).
	dbk := baseDump()
	dbk.Terrain.WallDiag[idx(wx, wy)] = 1
	save(dbk, "walls_diag_back_id1.json")

	// '/' diagonal wall, def id 0 (value 12001).
	dfw := baseDump()
	dfw.Terrain.WallDiag[idx(wx, wy)] = 12001
	save(dfw, "walls_diag_fwd_id1.json")

	// A wall whose def id is a WINDOW-like entry. In the real openrsc DoorDef.xml
	// index 3 = "Window" (unknown=0), index 4 = "Fence" (unknown=0),
	// index 1 = "Doorframe" (unknown=1). We place a V wall with each so the
	// wall-render harness (real defs) can show how GO treats them.
	for _, tc := range []struct {
		id1  int // 1-based stored value
		name string
	}{
		{1, "walls_v_wall.json"},      // def 0 "Wall" unknown=0
		{2, "walls_v_doorframe.json"}, // def 1 "Doorframe" unknown=1 doorType=0
		{3, "walls_v_door.json"},      // def 2 "Door" unknown=1 doorType=1
		{4, "walls_v_window.json"},    // def 3 "Window" unknown=0
		{5, "walls_v_fence.json"},     // def 4 "Fence" unknown=0
	} {
		dd := baseDump()
		dd.Terrain.WallV[idx(wx, wy)] = byte(tc.id1)
		save(dd, tc.name)
	}

	// Highwall: in DoorDef.xml the "Highwall" entry (modelVar1=275) is def index 7
	// (1-based value 8). Build it to test wall HEIGHT.
	hw := baseDump()
	hw.Terrain.WallV[idx(wx, wy)] = 8
	save(hw, "walls_v_highwall.json")

	// battlement: def index 10 (1-based 11), height=70 (a low wall). Height test.
	bat := baseDump()
	bat.Terrain.WallV[idx(wx, wy)] = 11
	save(bat, "walls_v_battlement.json")

	// snowwall: def index 12 (1-based 13). unknown=0 (built by standing-wall
	// loop), mv2=mv3=-31711 (a NEGATIVE flat colour, NOT a texture id). This is
	// the cleanest isolation of the per-object front/back-COLOUR table: method422
	// paints the face with -31711, GO paints flat stone grey.
	snow := baseDump()
	snow.Terrain.WallV[idx(wx, wy)] = 13
	save(snow, "walls_v_snowwall.json")

	// timberwall: def index 14 (1-based 15). unknown=0, mv2=mv3=21 (texture
	// "timbered" = a WOOD-toned wall). GO renders it as grey stone because its
	// name isn't in the wood name-heuristic and no texel buffer exists.
	timber := baseDump()
	timber.Terrain.WallV[idx(wx, wy)] = 15
	save(timber, "walls_v_timberwall.json")

	// blank: def index 16 (1-based 17). unknown=0 (built by standing-wall loop),
	// mv2=mv3=12345678 (TRANSPARENT sentinel). Both deob (build face, rasterizer
	// skips draw) and GO (skip quad entirely) render nothing -> control.
	blank := baseDump()
	blank.Terrain.WallV[idx(wx, wy)] = 17
	save(blank, "walls_v_blank.json")
}
