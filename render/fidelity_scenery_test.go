package render

import (
	"os"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
)

// modelsArchivePath is the OpenRSC models.orsc used to exercise PlaceScenery
// with real .ob3 geometry. Absent in a clean CI checkout, so the scenery
// fidelity test skips when it is missing (the terrain tests need no archive).
const modelsArchivePath = "/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/models.orsc"

// TestFidelity_WindmillSailsLifted (scenery-id74-no-vertical-lift): object id 74
// (windmill sails) floats 480 world-units UP so the sail assembly sits atop the
// mill tower, not on the ground (Mudclient.java:6229-6231:
//
//	if (objType == 74) model.a(0, 0, -480, true)
//
// 3rd arg -> baseZ -> vertical Y; -Y is up). The same model at a non-74 id gets
// no lift. We render windmillsail at id 74 and id 900 and assert id 74's face
// centroids are ~480 units higher (more negative Y). Before the fix the two were
// identical.
func TestFidelity_WindmillSailsLifted(t *testing.T) {
	if _, err := os.Stat(modelsArchivePath); err != nil {
		t.Skipf("models archive not present (%s); skipping scenery placement check", modelsArchivePath)
	}
	b, err := OpenBundle(modelsArchivePath)
	if err != nil {
		t.Skipf("open models archive: %v", err)
	}
	f := &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
		74:  {ID: 74, Name: "sails-id74", Model: "windmillsail", Width: 1, Height: 3},
		900: {ID: 900, Name: "sails-id900", Model: "windmillsail", Width: 1, Height: 3},
	}}

	mk := func(defID int) *rscdump.Dump {
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

	minY := func(defID int) int32 {
		faces, err := RenderDumpFacesWith(mk(defID), f, b)
		if err != nil {
			t.Fatalf("render id %d: %v", defID, err)
		}
		// the placed model is the highest model id in the set.
		maxModel := 0
		for _, ff := range faces {
			if ff.Model > maxModel {
				maxModel = ff.Model
			}
		}
		var min int32
		first := true
		for _, ff := range faces {
			if ff.Model != maxModel || maxModel == 0 {
				continue
			}
			if first || ff.Centroid[1] < min {
				min, first = ff.Centroid[1], false
			}
		}
		if first {
			t.Fatalf("no placed-model faces for id %d", defID)
		}
		return min
	}

	top74 := minY(74)
	top900 := minY(900)
	lift := top900 - top74 // positive => id 74 is higher (more negative Y)
	if lift < 400 || lift > 560 {
		t.Errorf("windmill sails (id 74) lift wrong: id74 topY=%d id900 topY=%d lift=%d (want ~480)", top74, top900, lift)
	}
}
