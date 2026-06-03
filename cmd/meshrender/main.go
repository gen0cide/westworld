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
	npcCanvas := flag.String("npc-canvas", "", "when set with RSC_MESH_NPC, write the orsc UNSCALED rat CompositeSprite canvas PNG to this path (Phase 2 / Milestone B byte-compare); skips the scene render")
	npcDir := flag.Int("npc-dir", 0, "facing dir (0..7) for -npc-canvas")
	npcStep := flag.Int("npc-step", 0, "walk step for -npc-canvas")
	realDefs := flag.String("realdefs", "", "OpenRSC server root (e.g. /home/free/code/rsc-hacking/openrsc); when set, the object def (model+W/H) is loaded from the AUTHENTIC GameObjectDef table keyed by -objid instead of the synthesized -model/-w/-h. orsc reads OpenRSC's GameObjectDef.xml, verified 1:1 by name+footprint vs the rev-235 content0 table the DEOB/JAR legs parse.")
	flag.Parse()

	// Phase 2 (Milestone B): dump the orsc UNSCALED rat composite canvas and exit.
	// This is the decode+composite+recolour the parity diff byte-compares against
	// the DEOB-reconstructed canvas; it needs no fixture/scene (the composite is the
	// figure canvas, not a scene render). content1 must be present at the cache dir.
	if *npcCanvas != "" {
		png, cw, ch, opaque, cerr := orsc.DumpRatCompositeCanvas(*npcDir, *npcStep)
		if cerr != nil {
			fmt.Fprintln(os.Stderr, "npc-canvas:", cerr)
			os.Exit(1)
		}
		if werr := os.WriteFile(*npcCanvas, png, 0o644); werr != nil {
			fmt.Fprintln(os.Stderr, "write npc-canvas:", werr)
			os.Exit(1)
		}
		fmt.Printf("wrote %s  canvas=%dx%d opaque=%d dir=%d step=%d\n", *npcCanvas, cw, ch, opaque, *npcDir, *npcStep)
		return
	}

	if *fixture == "" {
		fmt.Fprintln(os.Stderr, "meshrender: -fixture required")
		os.Exit(2)
	}
	d, err := rscdump.Load(*fixture)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load fixture:", err)
		os.Exit(1)
	}

	// Entity-parity renders (Phase 4 / Milestone C): the canonical fixture
	// door_diag_obj.json carries a 48001 diagonal-object band. orsc builds a SYNTHETIC
	// wood door-leaf stand-in there (diagobj.go), but the DEOB/JAR oracle legs build
	// NOTHING (objid 0 has no GameData def, so addModels drops it — the documented
	// door_diag_obj 9025/9026 caveat in ORSC_JAR_PARITY.md). That orsc-only door would
	// occlude the entity's lower body and contaminate the 3-engine diff with ~2.7k px
	// that are NOT an entity-blit divergence. So for an entity render with NO real
	// scenery object (placeObject false), clear the diagonal-object band so orsc builds
	// the SAME geometry the oracle does (terrain + the entity only) — an apples-to-apples
	// scene. This is a HARNESS fixture tweak (not a renderer change): terrain/scenery
	// (ladder/well, -model/-realdefs) renders are untouched, and the entity itself is
	// unaffected. Gated on RSC_MESH_NPC / RSC_MESH_PLAYER (the entity gate).
	entityGate := os.Getenv("RSC_MESH_NPC") != "" || os.Getenv("RSC_MESH_PLAYER") != ""
	if entityGate && *model == "" && *realDefs == "" && d.Terrain != nil {
		for i, v := range d.Terrain.WallDiag {
			if v > 48000 && v < 60000 {
				d.Terrain.WallDiag[i] = 0
			}
		}
	}

	var f *facts.Facts
	var b *orsc.Bundle
	placeObject := *model != "" || *realDefs != ""
	if placeObject {
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
		if *realDefs != "" {
			// Load the AUTHENTIC object def table from OpenRSC (GameObjectDef.xml, keyed by
			// list index = object id; verified 1:1 by name+footprint vs the rev-235 content0
			// string.dat/integer.dat table the DEOB/JAR legs parse). The orsc diagonal-object
			// path already reads f.SceneryDef(objectID).Model/Width/Height, so wiring the real
			// facts drives the object by its TRUE map id.
			f, err = facts.Load(facts.DefaultSources(*realDefs))
			if err != nil {
				fmt.Fprintln(os.Stderr, "load real defs:", err)
				os.Exit(1)
			}
			def := f.SceneryDef(*objid)
			if def == nil {
				fmt.Fprintf(os.Stderr, "no authentic def for objid %d\n", *objid)
				os.Exit(1)
			}
			// Keep only the DEFS (the diag-band lookup needs SceneryDefs); drop ALL world
			// placements (SceneryLocs/BoundaryLocs/NpcLocs/GroundItems). The DEOB/JAR legs
			// render a "fresh scene" — just the injected fixture grids + the single diag-band
			// object — so the orsc dump must NOT place OpenRSC's whole live world here (that
			// adds 19k+ divergent px from off-fixture scenery/NPCs). Mirrors the synthesized
			// facts path, which carries defs with no locs.
			f.SceneryLocs = nil
			f.BoundaryLocs = nil
			f.NpcLocs = nil
			f.GroundItemLocs = nil
			fmt.Printf("realdefs: id=%d -> model=%q W=%d H=%d type=%d\n", *objid, def.Model, def.Width, def.Height, def.Type)
		} else {
			f = &facts.Facts{SceneryDefs: map[int]*facts.SceneryDef{
				*objid: {Model: *model, Width: *w, Height: *h},
			}}
		}
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
		// The fixture's diagonal band carries objid 0 (value 48001); BuildDiagonalObjects
		// reads objectID = bandValue-48001. To drive by a TRUE id, remap any in-band tile to
		// 48001+objid so the lookup hits f.SceneryDef(objid). The DEOB/JAR legs remap the same.
		if *objid != 0 && d.Terrain != nil {
			for i, v := range d.Terrain.WallDiag {
				if v > 48000 && v < 60000 {
					d.Terrain.WallDiag[i] = int32(48001 + *objid)
				}
			}
		}
	}

	var png []byte
	var raw []int32
	if !placeObject {
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
