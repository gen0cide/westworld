// Command single_tile_nodoor authors the PERTURBED render-diff fixture: it is
// byte-for-byte the Stage-1 single_tile_door fixture with the ONE wall/door
// boundary REMOVED (the VerticalWall byte cleared back to 0). Everything else —
// terrain grids, camera, window, seed — is identical, so a diff between the two
// fixtures isolates exactly the door's wall quad.
//
//	go run ./internal/rscdump/gen/single_tile_nodoor
//
// This is the renderdiff SELF-TEST input (RENDER_DIFF_DESIGN.md §5): rendering
// with-door vs without-door must produce a non-zero pixel-diff region localized
// to the wall AND a structural diff showing the door face present in one set and
// absent in the other — demonstrating the tool detects the "unwalkable door /
// unperceived wall" bug class WITHOUT the Java engines.
//
// It is authored by loading the existing single_tile_door.json and zeroing its
// walls, so the two fixtures can NEVER drift apart in any field but the door.
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gen0cide/westworld/internal/rscdump"
)

func main() {
	root := func(rel ...string) string {
		p, err := filepath.Abs(filepath.Join(append([]string{"testdata", "rscdump"}, rel...)...))
		if err != nil {
			log.Fatal(err)
		}
		return p
	}

	// Load the canonical with-door fixture; the perturbation is the ONLY change.
	d, err := rscdump.Load(root("single_tile_door.json"))
	if err != nil {
		log.Fatalf("load with-door fixture (run gen/single_tile_door first?): %v", err)
	}

	// Remove every wall/door boundary: clear the H/V/diagonal grids to all-zero.
	// The with-door fixture's only boundary is one VerticalWall byte; zeroing all
	// three grids guarantees no boundary survives regardless of which one carried
	// the door, while leaving terrain (grass/elevation) untouched.
	removed := 0
	for i := range d.Terrain.WallV {
		if d.Terrain.WallV[i] != 0 {
			removed++
		}
		d.Terrain.WallV[i] = 0
	}
	for i := range d.Terrain.WallH {
		if d.Terrain.WallH[i] != 0 {
			removed++
		}
		d.Terrain.WallH[i] = 0
	}
	for i := range d.Terrain.WallDiag {
		if d.Terrain.WallDiag[i] != 0 {
			removed++
		}
		d.Terrain.WallDiag[i] = 0
	}

	out := root("single_tile_NOdoor.json")
	if err := d.Save(out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (removed %d wall byte(s); flat-grass control with NO boundary)\n", out, removed)
}
