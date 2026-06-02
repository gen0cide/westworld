// scenery_render renders an rscdump fixture through RenderDumpWith with a real
// facts (scenery defs) + model bundle, so we can visually inspect placed scenery
// (the render_dump CLI uses syntheticFacts which has NO scenery). Usage:
//
//	scenery_render <out.png> <defID> <model> <w> <h> [dir]
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render/orsc"
)

const modelsPath = "/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/models.orsc"

func main() {
	if len(os.Args) < 6 {
		fmt.Println("usage: scenery_render <out.png> <defID> <model> <w> <h> [dir]")
		os.Exit(2)
	}
	out := os.Args[1]
	defID, _ := strconv.Atoi(os.Args[2])
	model := os.Args[3]
	w, _ := strconv.Atoi(os.Args[4])
	h, _ := strconv.Atoi(os.Args[5])
	dir := 0
	if len(os.Args) > 6 {
		dir, _ = strconv.Atoi(os.Args[6])
	}

	const size = 16
	n := size * size
	d := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera:  rscdump.Camera{Yaw: 100 * 4, Distance: 1100, ScreenW: 512, ScreenH: 334, ViewDist: 9, ClipNear: 5, ClipFar: 7000},
		Window:  rscdump.Window{BaseX: 0, BaseY: 0, Plane: 0, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: make([]byte, n), GroundColour: make([]byte, n), TerrainSeed: 0},
		Scenery: []rscdump.Scenery{{X: 8, Y: 8, ID: defID, Dir: dir}},
	}
	b, err := orsc.OpenBundle(modelsPath)
	if err != nil {
		panic(err)
	}
	f := &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
		defID: {ID: defID, Name: model, Model: model, Width: w, Height: h},
	}}
	png, _, err := orsc.RenderDumpWith(d, f, b)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(out, png, 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s (defID=%d model=%s %dx%d dir=%d)\n", out, defID, model, w, h, dir)
}
