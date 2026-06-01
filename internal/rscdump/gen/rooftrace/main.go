// Command rooftrace loads an rscdump/1 fixture and calls render.BuildRoofs on it,
// printing each emitted roof face with its RAW local vertices (lx*128 grid coords,
// -height) so the 7-way cascade + heights can be checked against the deob spec
// without the camera transform getting in the way.
//
//	go run ./internal/rscdump/gen/rooftrace <dump.json>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

const wallObjectHeight = 192

// buildFacts mirrors render.syntheticFacts: every wall id referenced by the
// terrain grids gets a generic def with Height=192.
func buildFacts(d *rscdump.Dump) *facts.Facts {
	if d.Terrain == nil {
		return nil
	}
	ids := map[int]bool{}
	for _, b := range d.Terrain.WallH {
		if b != 0 {
			ids[int(b)-1] = true
		}
	}
	for _, b := range d.Terrain.WallV {
		if b != 0 {
			ids[int(b)-1] = true
		}
	}
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
	f := &facts.Facts{BoundaryDefs: map[int]*facts.BoundaryDef{}}
	for id := range ids {
		f.BoundaryDefs[id] = &facts.BoundaryDef{ID: id, Name: "Door", Height: wallObjectHeight, FrontDeco: -1, BackDeco: -1, DoorType: 1}
	}
	return f
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <dump.json>", os.Args[0])
	}
	d, err := rscdump.Load(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	land := d.Landscape()
	if land == nil {
		log.Fatal("no terrain")
	}
	f := buildFacts(d)
	const terrainSize = 160
	baseX := d.Self.X - terrainSize/2
	baseY := d.Self.Y - terrainSize/2
	plane := d.Window.Plane
	heights := render.TerrainHeights(land, baseX, baseY, plane)
	g := render.BuildRoofs(f, land, baseX, baseY, plane, heights)
	if g == nil {
		fmt.Println("BuildRoofs returned nil (no roof faces)")
		return
	}
	fmt.Printf("baseX=%d baseY=%d plane=%d  verts=%d faces=%d\n", baseX, baseY, plane, g.NumVertices, g.NumFaces)
	for i := 0; i < g.NumFaces; i++ {
		verts := g.FaceVertices[i]
		fmt.Printf("face #%d nv=%d front=%d back=%d  verts:", i, len(verts), g.FaceFillFront[i], g.FaceFillBack[i])
		for _, v := range verts {
			lx := g.VertexX[v]
			lz := g.VertexZ[v]
			ly := g.VertexY[v]
			fmt.Printf("  (lx=%d+%d lz=%d+%d h=%d)", lx/128, lx%128, lz/128, lz%128, -ly)
		}
		fmt.Println()
	}
}
