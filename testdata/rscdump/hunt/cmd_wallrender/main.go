// Command cmd_wallrender renders an rscdump/1 fixture through the GO engine with
// CALLER-SUPPLIED boundary defs (so we exercise the per-object wall colour /
// visibility / height tables that method422 reads, instead of the generic
// wood-door synthetic facts). It writes a PNG and prints the built face set so
// we can structurally diff GO-actual vs the deob method422 spec.
//
// Usage:
//
//	go run ./testdata/rscdump/hunt/cmd_wallrender <dump.json> <out.png> [defsMode]
//
// defsMode (optional):
//
//	"real"      load the real openrsc DoorDef.xml (default)
//	"synthetic" mimic render.syntheticFacts (generic wood door, no texture)
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
	"github.com/gen0cide/westworld/render/orsc"
)

const doorDefXML = "/home/free/code/rsc-hacking/openrsc/server/conf/server/defs/DoorDef.xml"

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <dump.json> <out.png> [real|synthetic]", os.Args[0])
	}
	inPath, outPath := os.Args[1], os.Args[2]
	mode := "real"
	if len(os.Args) >= 4 {
		mode = os.Args[3]
	}

	d, err := rscdump.Load(inPath)
	if err != nil {
		log.Fatal(err)
	}

	var f *facts.Facts
	switch mode {
	case "real":
		f, err = loadRealBoundaryDefs()
		if err != nil {
			log.Fatal(err)
		}
	case "synthetic":
		// generic wood door for every referenced wall id
		f = syntheticLike(d)
	default:
		log.Fatalf("unknown defsMode %q", mode)
	}

	png, _, err := orsc.RenderDumpWith(d, f, nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outPath, png, 0o644); err != nil {
		log.Fatal(err)
	}

	faces, err := orsc.RenderDumpFacesWith(d, f, nil)
	if err != nil {
		log.Fatal(err)
	}
	// emit a faces.json sidecar
	type facesDoc struct {
		Schema string             `json:"schema"`
		Faces  []render.BuiltFace `json:"faces"`
	}
	fd := facesDoc{Schema: "rscdump-faces/1", Faces: faces}
	b, _ := json.MarshalIndent(fd, "", "  ")
	_ = os.WriteFile(outPath+".faces.json", b, 0o644)

	fmt.Printf("rendered %s -> %s (mode=%s, %d faces, %d bytes PNG)\n", inPath, outPath, mode, len(faces), len(png))
	// print only the NON-terrain faces (walls): terrain faces are numerous and
	// have flat tile fills; walls live above y=0 at the wall centroid.
	for _, fc := range faces {
		// heuristic: wall faces have 4 verts and a negative Y centroid magnitude
		// far from the ground band; print all 4-vert faces whose |Y|>=24 (a wall
		// half-height ~ -96). Terrain quads sit near y in {0,-something small}.
		if fc.NumVerts == 4 && (fc.Centroid[1] <= -40) {
			fmt.Printf("  WALLFACE c=%v n=%d front=%d back=%d\n", fc.Centroid, fc.NumVerts, fc.FillFront, fc.FillBack)
		}
	}
}

func loadRealBoundaryDefs() (*facts.Facts, error) {
	b, err := os.ReadFile(doorDefXML)
	if err != nil {
		return nil, err
	}
	type doorDefXMLrow struct {
		Name      string `xml:"name"`
		ModelVar1 int    `xml:"modelVar1"`
		ModelVar2 int    `xml:"modelVar2"`
		ModelVar3 int    `xml:"modelVar3"`
		DoorType  int    `xml:"doorType"`
		Unknown   int    `xml:"unknown"`
	}
	type arr struct {
		Defs []doorDefXMLrow `xml:"DoorDef"`
	}
	var a arr
	if err := xml.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	defs := map[int]*facts.BoundaryDef{}
	for i, r := range a.Defs {
		defs[i] = &facts.BoundaryDef{
			ID: i, Name: r.Name, Height: r.ModelVar1,
			FrontDeco: r.ModelVar2, BackDeco: r.ModelVar3,
			DoorType: r.DoorType, Unknown: r.Unknown,
		}
	}
	return &facts.Facts{BoundaryDefs: defs}, nil
}

func syntheticLike(d *rscdump.Dump) *facts.Facts {
	ids := map[int]bool{}
	add := func(g []byte) {
		for _, b := range g {
			if b != 0 {
				ids[int(b)-1] = true
			}
		}
	}
	add(d.Terrain.WallH)
	add(d.Terrain.WallV)
	for _, dv := range d.Terrain.WallDiag {
		if dv <= 0 || dv >= 24000 {
			continue
		}
		if dv < 12000 {
			ids[int(dv)-1] = true
		} else {
			ids[int(dv-12001)] = true
		}
	}
	defs := map[int]*facts.BoundaryDef{}
	for id := range ids {
		defs[id] = &facts.BoundaryDef{ID: id, Name: "Door", Height: 192, FrontDeco: -1, BackDeco: -1, DoorType: 1, Unknown: 1}
	}
	return &facts.Facts{BoundaryDefs: defs}
}
