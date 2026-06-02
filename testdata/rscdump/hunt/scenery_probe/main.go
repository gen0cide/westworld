// scenery_probe drives render.PlaceScenery (via RenderDumpFacesWith with a real
// facts + model bundle) for a multi-tile, asymmetric scenery object (Longtable,
// w=4 h=1) at each of the 8 directions, and prints the placed model's footprint
// centre (mean of all its face centroids in transformed world space) so we can
// compare GO-actual against the deob addModels / Mudclient-rebase formula:
//
//	deob (correct obf-table identities, World.java:800-806 & Mudclient:6125-6133):
//	  dir 0/4 : Xextent = objectHeight(def.Height), Zextent = objectWidth(def.Width)
//	  dir else: Xextent = objectWidth(def.Width),   Zextent = objectHeight(def.Height)
//	  cx = 128*(Xextent + 2*tileX)/2 ; cz = 128*(Zextent + 2*tileY)/2
//
// The terrain is flat (all-zero tiles) so elevation is 0 and the model's
// vertical anchor is 0; only X/Z matter for the footprint-centre test.
package main

import (
	"fmt"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render/orsc"
)

const modelsPath = "/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/models.orsc"

func mkDump(tileX, tileY, defID, dir, size int) *rscdump.Dump {
	n := size * size
	return &rscdump.Dump{
		Schema: rscdump.SchemaID,
		Level:  rscdump.LevelL1,
		Source: rscdump.SourceHandAuthored,
		Camera: rscdump.Camera{
			Yaw: 0, Distance: 1500, ScreenW: 512, ScreenH: 334,
			ViewDist: 9, ClipNear: 5, ClipFar: 7000,
		},
		Window: rscdump.Window{BaseX: 0, BaseY: 0, Plane: 0, Size: size},
		Terrain: &rscdump.Terrain{
			Size:        size,
			Elevation:   make([]byte, n),
			TerrainSeed: 0,
		},
		Scenery: []rscdump.Scenery{{X: tileX, Y: tileY, ID: defID, Dir: dir}},
	}
}

func main() {
	b, err := orsc.OpenBundle(modelsPath)
	if err != nil {
		panic(err)
	}

	// One asymmetric multi-tile def: Longtable, width=4, height=1, model "longtable".
	// (mirrors GameObjectDef id 9). We also test a 2x3 (cart) and 1x2 (range).
	type tc struct {
		name  string
		defID int
		model string
		w, h  int
	}
	cases := []tc{
		{"Longtable", 9, "longtable", 4, 1},
		{"Cart", 54, "cart", 2, 3},
		{"Range", 11, "range", 1, 2},
	}

	const size = 16
	const tileX, tileY = 6, 6
	// buildScene re-bases the window: baseX = v.X - terrainSize/2, where
	// v.X = dump.BaseX + size/2. terrainSize is the render lib's window size (160).
	// We don't know it here, so we recover the window-local tile from the faces by
	// subtracting a constant; instead we compare GO-formula vs deob-formula in the
	// SAME (window-local) frame by deriving the window-local tile from the observed
	// 1x1 symmetric placement. Simpler: replicate BOTH formulas in window-local
	// coords using the SAME local tile index lx/ly, since the rebase is identical
	// for both engines. The rebase offset is whatever GO uses; we measure it once.

	for _, c := range cases {
		f := &facts.Facts{
			SceneryDefs: map[int]*facts.SceneryDef{
				c.defID: {ID: c.defID, Name: c.name, Model: c.model, Width: c.w, Height: c.h},
			},
		}
		fmt.Printf("\n=== %s (defID=%d width=%d height=%d) ===\n", c.name, c.defID, c.w, c.h)
		fmt.Printf("Comparing the footprint-CENTRE GO actually translates to (recovered from\n")
		fmt.Printf("face centroids) vs GO-formula vs deob-formula, all in window-local world units.\n")
		fmt.Printf("dir | GO-actual cX,cZ | GO-formula cX,cZ | deob-formula cX,cZ | axis-swap?\n")

		for dir := 0; dir < 8; dir++ {
			d := mkDump(tileX, tileY, c.defID, dir, size)
			faces, err := orsc.RenderDumpFacesWith(d, f, b)
			if err != nil {
				panic(err)
			}
			maxModel := 0
			for _, ff := range faces {
				if ff.Model > maxModel {
					maxModel = ff.Model
				}
			}
			var sx, sz, cnt int64
			for _, ff := range faces {
				if ff.Model == maxModel && maxModel > 0 {
					sx += int64(ff.Centroid[0])
					sz += int64(ff.Centroid[2])
					cnt++
				}
			}
			var goActualX, goActualZ int64
			if cnt > 0 {
				goActualX = sx / cnt
				goActualZ = sz / cnt
			}

			// Recover window-local tile lx/ly from GO-actual for dir 0 by inverting
			// the GO formula (1x1 reference would give lx*128+64). We instead derive
			// lx,ly from the fact that GO translates to:
			//   cx = (2*lx + wExt)*128/2 ; cz = (2*ly + hExt)*128/2  (GO PlaceScenery)
			// GO PlaceScenery extents:
			var goWExt, goHExt int
			goWExt, goHExt = c.w, c.h
			if dir != 0 && dir != 4 {
				goWExt, goHExt = goHExt, goWExt
			}
			// deob extents — COLLISION-CONSISTENT identity (ub.g=Height, f.f=Width per
			// Mudclient comments + BOTH OpenRSC oracle paths agree): dir 0/4 -> X=Width,
			// Z=Height (no swap); else -> X=Height, Z=Width (swap).
			var deWExt, deHExt int
			if dir == 0 || dir == 4 {
				deWExt, deHExt = c.w, c.h
			} else {
				deWExt, deHExt = c.h, c.w
			}

			// We need the window-local lx,ly. Derive it from the GO-actual centre of
			// the dir=0 placement and the GO formula (model-local centroid offset is
			// small relative to a tile and rotation-symmetric; we round to tiles).
			// lx = (2*goActualX/128 - wExt)/2 (approx). Use it consistently.
			lx := (2*goActualX/128 - int64(goWExt)) / 2
			ly := (2*goActualZ/128 - int64(goHExt)) / 2
			goCX := (2*lx + int64(goWExt)) * 128 / 2
			goCZ := (2*ly + int64(goHExt)) * 128 / 2
			deCX := (2*lx + int64(deWExt)) * 128 / 2
			deCZ := (2*ly + int64(deHExt)) * 128 / 2

			swap := "—"
			if goWExt != deWExt || goHExt != deHExt {
				swap = fmt.Sprintf("SWAPPED (GO Xext=%d Zext=%d ; deob Xext=%d Zext=%d)", goWExt, goHExt, deWExt, deHExt)
			}
			fmt.Printf(" %d  | %6d,%6d | %6d,%6d | %6d,%6d | %s\n",
				dir, goActualX, goActualZ, goCX, goCZ, deCX, deCZ, swap)
		}
	}
}
