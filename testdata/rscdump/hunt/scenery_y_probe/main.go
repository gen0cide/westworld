// scenery_y_probe checks the VERTICAL anchor of placed scenery: a normal object
// sits with its base at Y = -elevation (0 on flat ground), but object id 74
// (windmill sails) is lifted an extra 480 units UP in the deob
// (Mudclient.java:6142-6144: `if (objType==74) model.translate(0,0,-480,true)`),
// so the sails sit atop the mill tower, not at ground level. We render a
// windmillsail at id 74 and a same-model object at a NON-74 id, and compare the
// min transformed-Y (the highest point, since -Y is up) of each. If GO applied
// the +480 lift, the id-74 model's verts would be ~480 units higher (more
// negative Y) than the same model at a non-74 id. GO has no id-74 case, so they
// will be IDENTICAL — that is the bug.
package main

import (
	"fmt"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

const modelsPath = "/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/models.orsc"

func mkDump(defID int) *rscdump.Dump {
	const size = 16
	n := size * size
	return &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera:  rscdump.Camera{Yaw: 0, Distance: 1500, ScreenW: 512, ScreenH: 334, ViewDist: 9, ClipNear: 5, ClipFar: 7000},
		Window:  rscdump.Window{BaseX: 0, BaseY: 0, Plane: 0, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: make([]byte, n), TerrainSeed: 0},
		Scenery: []rscdump.Scenery{{X: 6, Y: 6, ID: defID, Dir: 0}},
	}
}

func minMaxY(faces []render.BuiltFace) (minY, maxY int32, n int) {
	maxModel := 0
	for _, ff := range faces {
		if ff.Model > maxModel {
			maxModel = ff.Model
		}
	}
	first := true
	for _, ff := range faces {
		if ff.Model != maxModel || maxModel == 0 {
			continue
		}
		y := ff.Centroid[1]
		if first {
			minY, maxY = y, y
			first = false
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
		n++
	}
	return
}

func main() {
	b, err := render.OpenBundle(modelsPath)
	if err != nil {
		panic(err)
	}
	// Same model "windmillsail" placed at id 74 (special) and id 900 (no special).
	f := &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
		74:  {ID: 74, Name: "sails-id74", Model: "windmillsail", Width: 1, Height: 3},
		900: {ID: 900, Name: "sails-id900", Model: "windmillsail", Width: 1, Height: 3},
	}}

	for _, id := range []int{74, 900} {
		faces, err := render.RenderDumpFacesWith(mkDump(id), f, b)
		if err != nil {
			panic(err)
		}
		minY, maxY, n := minMaxY(faces)
		fmt.Printf("id=%-4d windmillsail: face-centroid Y range [%d .. %d] (n=%d faces)\n", id, minY, maxY, n)
	}
	fmt.Println("\nDeob lifts id 74 by -480 (UP); GO has no id-74 case, so the two")
	fmt.Println("rows above will be IDENTICAL. A faithful port would show id 74's Y")
	fmt.Println("range shifted ~ -480 vs id 900.")
}
