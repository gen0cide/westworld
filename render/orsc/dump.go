package orsc

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
)

// dump.go is the orsc-backed retarget of the old render/dump.go pixel leg. It
// renders an rscdump/1 L1 dump through the faithful orsc engine and returns the
// encoded PNG, the raw int32 framebuffer (Surface.Pixels — 0x00RRGGBB row-major,
// byte-identical to the old render.Surface.Pix layout), and any error. This is
// the GO engine's "render a dump -> PNG/int[]" entrypoint
// (RENDER_DIFF_DESIGN.md §5) the pixel diff / heatmap consume.
//
// Determinism: the terrain grids come straight out of the dump (Dump.Landscape),
// no map file is decoded; the per-vertex ambience speckle is orsc's deterministic
// coord-hash (world.go terrainAmbience), so two renders of the same dump are
// byte-identical without pinning the JAR's Math.random sequence; camera / window
// are taken verbatim from the dump (dumpView). The dump's TerrainSeed only affects
// per-vertex SHADE (never geometry) and orsc folds the coordinate, not the seed,
// so the render is reproducible by construction.
//
// Colours: a hand-authored fixture with no GameData leaves the Bundle nil (no
// scenery models); RenderDump synthesizes a generic openable boundary def for
// every wall id the terrain grids reference (syntheticFacts) so a wall/door
// renders without loading DoorDef.xml. A caller with real GameData passes it via
// RenderDumpWith.

// RenderDump renders an L1 dump and returns (PNG bytes, raw int32 framebuffer,
// error), using syntheticFacts so a bare fixture still builds its wall/door.
func RenderDump(d *rscdump.Dump) (pngBytes []byte, rawPix []int32, err error) {
	return RenderDumpWith(d, syntheticFacts(d), nil)
}

// RenderDumpWith renders a dump with caller-supplied facts (boundary/scenery defs)
// and asset bundle (scenery models). f=nil skips boundaries + scenery (bare
// terrain); b=nil skips scenery models (boundaries still render from f). This is
// the seam a real-map dump uses to render against the same GameData the live
// client uses.
func RenderDumpWith(d *rscdump.Dump, f *facts.Facts, b *Bundle) (pngBytes []byte, rawPix []int32, err error) {
	scene, err := buildDumpScene(d, f, b, true)
	if err != nil {
		return nil, nil, err
	}
	scene.Render()
	surf := scene.graphics
	png, err := surf.PNG()
	if err != nil {
		return nil, nil, err
	}
	// Copy the framebuffer so the caller owns it independently of the Surface
	// (Surface.Pixels IS the int[] the rasterizer wrote — 0x00RRGGBB row-major).
	raw := make([]int32, len(surf.Pixels))
	copy(raw, surf.Pixels)
	return png, raw, nil
}
