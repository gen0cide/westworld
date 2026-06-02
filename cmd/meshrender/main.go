// Command meshrender renders an rscdump/1 L1 fixture through the faithful orsc
// engine, optionally placing a REAL scenery object (decoded from the authentic
// rev-235 "3d models" content archive) on the fixture's diagonal-object band.
//
// It is the orsc leg of the 3-engine scenery-mesh parity check (orsc vs the
// rev-235 JAR via rscplus/DumpRenderer vs the readable DEOB via client.DumpRender):
// all three load the SAME content9 .ob3 bytes and place the object by the SAME
// addModels math, so the render compares each engine's decode->build->place->
// light->project->raster pipeline on byte-identical input.
//
// usage:
//
//	ORSC_FLAT_AMBIENCE=1 meshrender -fixture F.json -out out.png \
//	    [-model ladder -objid 0 -w 1 -h 1] [-cache /tmp/rsc-run/cache]
//
// With -model unset it renders bare terrain+synthetic (RenderDump). With -model
// set it synthesizes facts.SceneryDef(objid)={model,w,h} and loads the model from
// the content "3d models" archive, exercising the real diagonal-object path.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render/orsc"
)

func main() {
	fixture := flag.String("fixture", "", "rscdump/1 L1 fixture JSON")
	out := flag.String("out", "out.png", "output PNG path")
	model := flag.String("model", "", "content9 object model name (e.g. ladder); empty = no object")
	objid := flag.Int("objid", 0, "object id for the diagonal band (diag value = 48001+objid)")
	w := flag.Int("w", 1, "object footprint width (tiles)")
	h := flag.Int("h", 1, "object footprint height (tiles)")
	dir := flag.Int("dir", 0, "tile direction 0-7 (object heading; rotates dir*32 about Y)")
	cache := flag.String("cache", "/tmp/rsc-run/cache", "RSC content-pack cache dir (holds content9_*)")
	flag.Parse()

	if *fixture == "" {
		fmt.Fprintln(os.Stderr, "meshrender: -fixture required")
		os.Exit(2)
	}
	d, err := rscdump.Load(*fixture)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load fixture:", err)
		os.Exit(1)
	}

	var f *facts.Facts
	var b *orsc.Bundle
	if *model != "" {
		// content9 is the "3d models" archive (readDataFile("3d models",60,9,84)).
		// Resolve by glob so a regenerated cache (changed CRC suffix) still matches.
		matches, _ := filepath.Glob(filepath.Join(*cache, "content9_*"))
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "no content9_* in %s\n", *cache)
			os.Exit(1)
		}
		arc, err := assets.OpenArchive(matches[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "open models archive:", err)
			os.Exit(1)
		}
		f = &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
			*objid: {Model: *model, Width: *w, Height: *h},
		}}
		b = &orsc.Bundle{Models: arc}
		// content11 = the "Textures" archive (readDataFile("Textures",50,11,111)): the
		// texel-bank source for TEXTURED object faces (e.g. the well's wall/planks
		// faces). Load it onto the Bundle so the pixel leg renders those faces with
		// real texels instead of skipping them. Resolve by glob (CRC suffix varies).
		if texMatches, _ := filepath.Glob(filepath.Join(*cache, "content11_*")); len(texMatches) > 0 {
			if texArc, terr := assets.OpenArchive(texMatches[0]); terr == nil {
				b.Textures = texArc
			} else {
				fmt.Fprintln(os.Stderr, "warn: open textures archive:", terr)
			}
		}
		// Drive the object's heading uniformly via the dump's tileDirection grid
		// (World.getTileDirection): BuildDiagonalObjects reads it per object tile and
		// rotates dir*32 about Y. Filling all tiles (dir 0 == the absent-grid default)
		// keeps the DEOB/JAR legs in sync (they fill their tileDirection grid the same).
		if d.Terrain != nil {
			n := d.Terrain.Size * d.Terrain.Size
			td := make([]byte, n)
			for i := range td {
				td[i] = byte(*dir)
			}
			d.Terrain.TileDirection = td
		}
	}

	var png []byte
	var raw []int32
	if *model == "" {
		png, raw, err = orsc.RenderDump(d) // syntheticFacts: bare terrain + wall/door leaves
	} else {
		png, raw, err = orsc.RenderDumpWith(d, f, b)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
	nz := 0
	for _, p := range raw {
		if p&0xFFFFFF != 0 {
			nz++
		}
	}
	if err := os.WriteFile(*out, png, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s  nonzero=%d/%d  object=%q(id %d, %dx%d)\n", *out, nz, len(raw), *model, *objid, *w, *h)
}
