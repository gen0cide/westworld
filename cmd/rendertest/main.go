package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

const (
	modelsPath    = "/Users/flint/Code/openrsc/Client_Base/Cache/video/models.orsc"
	landscapePath = "/Users/flint/Code/openrsc/Client_Base/Cache/video/Authentic_Landscape.orsc"
	openRSCRoot   = "/Users/flint/Code/openrsc"
	outDir        = "/tmp/render_out"
	outPNG        = "/tmp/render_out/iter1.png"
)

func main() {
	_ = os.MkdirAll(outDir, 0o755)

	// ---- Phase 0: prove the archive + .ob3 decode ----
	arc, err := assets.OpenArchive(modelsPath)
	if err != nil {
		fmt.Println("FATAL open models.orsc:", err)
		os.Exit(1)
	}
	fmt.Printf("models.orsc: %d catalog entries\n", arc.EntryCount())

	probeNames := []string{"tree", "well", "tree2", "table", "chair", "ladder", "altar", "range"}
	var decoded int
	for _, name := range probeNames {
		for _, cand := range []string{name, name + ".ob3", name + "1"} {
			if !arc.Has(cand) {
				continue
			}
			b, err := arc.Get(cand)
			if err != nil {
				fmt.Printf("  %-12s present but read err: %v\n", cand, err)
				break
			}
			m := assets.DecodeModel(b, 0)
			fmt.Printf("  %-12s OK  bytes=%-6d verts=%-4d faces=%-4d\n", cand, len(b), m.NumVertices, m.NumFaces)
			decoded++
			break
		}
	}
	fmt.Printf("decoded %d/%d probe models\n", decoded, len(probeNames))

	// ---- Phase 1: terrain + scenery render ----
	land, err := pathfind.OpenLandscape(landscapePath)
	if err != nil {
		fmt.Println("WARN open landscape:", err)
	}

	var f *facts.Facts
	if land != nil {
		fc, ferr := facts.Load(facts.DefaultSources(openRSCRoot))
		if ferr != nil {
			fmt.Println("WARN load facts:", ferr)
		} else {
			f = fc
			fmt.Printf("facts: %d scenery defs, %d scenery locs | %d boundary defs, %d boundary locs\n",
				len(f.SceneryDefs), len(f.SceneryLocs), len(f.BoundaryDefs), len(f.BoundaryLocs))
			// how many boundary locs fall in bernard's (120,649) window [72..166, 601..695]?
			bw := 0
			for _, b := range f.BoundaryLocs {
				if b.X >= 72 && b.X <= 166 && b.Y >= 601 && b.Y <= 695 {
					bw++
					if bw <= 6 {
						d := f.BoundaryDef(b.DefID)
						nm := "?"
						if d != nil {
							nm = d.Name
						}
						fmt.Printf("  bnd loc (%d,%d) dir=%d def=%d %q\n", b.X, b.Y, b.Direction, b.DefID, nm)
					}
				}
			}
			fmt.Printf("boundary locs in (120,649) window: %d\n", bw)
		}
	}

	bundle := &render.Bundle{Models: render.NewModelCache(arc)}

	if land == nil {
		fmt.Println("no landscape; emitting Phase-0 framebuffer only")
		emitProbeFramebuffer(arc)
		return
	}

	// Offline entities: use NPC spawn tiles (StartX/StartY) as standing-frame
	// billboards near the camera. The live render-view path snapshots the real
	// perceived NPCs/players instead; offline we have only the static spawns.
	entitiesNear := func(cx, cy, plane, radius int) []render.Entity {
		var ents []render.Entity
		if f == nil {
			return ents
		}
		for _, n := range f.NpcLocs {
			if abs(n.StartX-cx) <= radius && abs(n.StartY-cy) <= radius {
				// Resolve the config85.jag npc id (the sprite source) from the
				// NPC name — stable across data sets — and fall back to the
				// fact's DefID if the name lookup misses. NpcID drives the 2D
				// sprite composite in DrawEntitySprites.
				npcID := n.DefID
				if d := f.NpcDef(n.DefID); d != nil {
					if cid := render.NpcIDForName(d.Name); cid >= 0 {
						npcID = cid
					}
				}
				ents = append(ents, render.Entity{
					X: n.StartX, Y: n.StartY, Kind: render.EntityNPC, NpcID: npcID,
				})
			}
		}
		return ents
	}

	type shot struct {
		name string
		view render.View
		out  string
	}
	// Calibration: bernard's exact tile (120,649), 4 cardinal rotations at
	// native RSC viewport 512x346 (the user's RSCPlus is 2x integer-scaled
	// = 1024x692). Compared against /tmp/render_out/GT_bernard_rot*.png.
	shots := []shot{}
	for _, rot := range []int{0, 32, 64, 96, 128, 160, 192, 224} {
		shots = append(shots, shot{
			fmt.Sprintf("cal_rot%d", rot),
			render.View{X: 120, Y: 649, Plane: 0, Rotation: rot, Zoom: 750, W: 512, H: 336, Entities: entitiesNear(120, 649, 0, 20)},
			filepath.Join(outDir, fmt.Sprintf("cal_rot%d.png", rot)),
		})
	}
	// Gatehouse passage: host UNDER the Lumbridge gate arch (an overlay-3
	// under-roof tile). With the 0x80 under-roof cull the arch roof + upper-story
	// walls are hidden, so this should render a LIT, walkable passage rather than
	// a dark/opaque ceiling.
	for _, rot := range []int{0, 96} {
		shots = append(shots, shot{
			fmt.Sprintf("gate_rot%d", rot),
			render.View{X: 118, Y: 656, Plane: 0, Rotation: rot, Zoom: 750, W: 512, H: 336, Entities: entitiesNear(118, 656, 0, 20)},
			filepath.Join(outDir, fmt.Sprintf("gate_rot%d.png", rot)),
		})
	}
	// Castle entrance: alex's exact RSCPlus tile (116,659), the def-16 "blank"
	// archway at x=117 he walks through. RSCPlus shows an OPEN lit passage +
	// arrow-slit (def-13 "arrowslit") wall above; we used to paint a flat-grey
	// quad for the "blank" arch. Sweep rotations to match the user's view.
	for _, rot := range []int{0, 64, 128, 192} {
		shots = append(shots, shot{
			fmt.Sprintf("castle_rot%d", rot),
			render.View{X: 116, Y: 659, Plane: 0, Rotation: rot, Zoom: 750, W: 512, H: 336, Entities: entitiesNear(116, 659, 0, 20)},
			filepath.Join(outDir, fmt.Sprintf("castle_rot%d.png", rot)),
		})
	}
	for _, sh := range shots {
		png, err := render.RenderView(land, f, bundle, sh.view)
		if err != nil {
			fmt.Printf("FATAL RenderView %s: %v\n", sh.name, err)
			emitProbeFramebuffer(arc)
			os.Exit(1)
		}
		if err := os.WriteFile(sh.out, png, 0o644); err != nil {
			fmt.Println("FATAL write png:", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s (%d bytes)\n", sh.out, len(png))
		assessPNG(sh.out)
	}
}

// emitProbeFramebuffer dumps a single decoded model silhouette so Phase-0
// always produces a visible PNG even if the scene path fails.
func emitProbeFramebuffer(arc *assets.Archive) {
	mc := render.NewModelCache(arc)
	var am *assets.Model
	for _, n := range []string{"tree", "well", "table"} {
		if am = mc.Get(n); am != nil {
			break
		}
	}
	if am == nil {
		return
	}
	g := render.FromAsset(am)
	g.SetLight(48, 48, -50, -10, -50)
	sc := &render.Scene{}
	sc.Add(g)
	sc.Cam = render.SetCamera(0, -200, -600, 912, 0, 0, 0)
	surf := render.NewSurface(512, 334)
	surf.Clear(0x404040)
	sc.RenderTo(surf, 256, 167)
	b, _ := surf.PNG()
	_ = os.WriteFile(filepath.Join(outDir, "iter1_probe.png"), b, 0o644)
	fmt.Println("wrote probe framebuffer iter1_probe.png")
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func assessPNG(path string) {
	fr, err := os.Open(path)
	if err != nil {
		return
	}
	defer fr.Close()
	img, _, err := image.Decode(fr)
	if err != nil {
		fmt.Println("decode err:", err)
		return
	}
	b := img.Bounds()
	colourCount := map[uint32]int{}
	var nonSky int
	const sky = 0x6080a0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, gg, bb, _ := img.At(x, y).RGBA()
			key := (r>>8)<<16 | (gg>>8)<<8 | (bb >> 8)
			colourCount[key]++
			if key != sky {
				nonSky++
			}
		}
	}
	total := (b.Max.X - b.Min.X) * (b.Max.Y - b.Min.Y)
	fmt.Printf("PNG %dx%d: %d distinct colours, %d/%d non-sky pixels (%.1f%%)\n",
		b.Dx(), b.Dy(), len(colourCount), nonSky, total, 100*float64(nonSky)/float64(total))
}
