package render

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
)

const doorDefXMLPath = "/home/free/code/rsc-hacking/openrsc/server/conf/server/defs/DoorDef.xml"

// loadRealBoundaryDefs reads the authentic OpenRSC DoorDef.xml so the wall fill
// test exercises the real per-object modelVar2/3 colour table (method422), not
// the generic synthetic wood door. Mirrors testdata/.../cmd_wallrender.
func loadRealBoundaryDefs(t *testing.T) *facts.Facts {
	t.Helper()
	b, err := os.ReadFile(doorDefXMLPath)
	if err != nil {
		t.Skipf("DoorDef.xml not present (%s); skipping real-def wall fill check", doorDefXMLPath)
	}
	type row struct {
		Name      string `xml:"name"`
		ModelVar1 int    `xml:"modelVar1"`
		ModelVar2 int    `xml:"modelVar2"`
		ModelVar3 int    `xml:"modelVar3"`
		DoorType  int    `xml:"doorType"`
		Unknown   int    `xml:"unknown"`
	}
	var a struct {
		Defs []row `xml:"DoorDef"`
	}
	if err := xml.Unmarshal(b, &a); err != nil {
		t.Fatalf("parse DoorDef.xml: %v", err)
	}
	defs := make(map[int]*facts.BoundaryDef, len(a.Defs))
	for i, d := range a.Defs {
		defs[i] = &facts.BoundaryDef{
			ID: i, Name: d.Name, Height: d.ModelVar1,
			FrontDeco: d.ModelVar2, BackDeco: d.ModelVar3,
			DoorType: d.DoorType, Unknown: d.Unknown,
		}
	}
	return &facts.Facts{BoundaryDefs: defs}
}

// TestFidelity_WallColourTable (wall-colour-table-ignored / wall-front-back-
// collapsed): method422 paints each wall face with the def's own modelVar2/3
// (FrontDeco/BackDeco) verbatim — a flat <0 colour straight through, a >=0
// texture id via the baked per-id colour. A SNOWWALL (def 12) carries the flat
// colour -31711 (snow-white); before the fix GO collapsed every wall to a
// hardcoded stone grey (-16914). We render the snowwall fixture with the real
// DoorDef.xml and assert the wall face fill is -31711, not the stone heuristic.
func TestFidelity_WallColourTable(t *testing.T) {
	f := loadRealBoundaryDefs(t)
	p := filepath.Join("..", "testdata", "rscdump", "hunt", "walls_v_snowwall.json")
	d, err := rscdump.Load(p)
	if err != nil {
		t.Fatalf("load snowwall fixture: %v", err)
	}
	faces, err := RenderDumpFacesWith(d, f, nil)
	if err != nil {
		t.Fatalf("render snowwall: %v", err)
	}
	const snowFlat = int32(-31711) // DoorDef snowwall modelVar2/3
	const stoneHeuristic = int32(-16914)
	found := false
	for _, fc := range faces {
		// the snow wall quad sits at the wall centroid, well above the ground band.
		if fc.NumVerts != 4 || fc.Centroid[1] > -40 {
			continue
		}
		if fc.FillFront == stoneHeuristic || fc.FillBack == stoneHeuristic {
			t.Errorf("snowwall wall face still painted the stone heuristic (%d) — per-object colour table ignored", stoneHeuristic)
		}
		if fc.FillFront == snowFlat && fc.FillBack == snowFlat {
			found = true
		}
	}
	if !found {
		t.Errorf("snowwall wall face not painted its def colour %d (want front==back==%d)", snowFlat, snowFlat)
	}
}
