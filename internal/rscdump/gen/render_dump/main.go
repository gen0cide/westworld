// Command render_dump renders an rscdump/1 L1 fixture through the GO engine
// (orsc.RenderDump — the faithful three/ renderer) to a PNG, and prints sanity
// stats (dimensions + count of non-background pixels). This is the Phase-0
// "dump → PNG" smoke test (RENDER_DIFF_DESIGN.md §6).
//
//	go run ./internal/rscdump/gen/render_dump <dump.json> <out.png>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render/orsc"
)

// backgroundColour is the flat fill an orsc-rendered pixel keeps when no geometry
// covers it. The orsc dump surface is not sky-cleared (NewSurface zero-fills), so
// uncovered pixels are 0x000000; counting non-zero pixels gauges scene coverage.
const backgroundColour = int32(0x000000)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <dump.json> <out.png>", os.Args[0])
	}
	inPath, outPath := os.Args[1], os.Args[2]

	d, err := rscdump.Load(inPath)
	if err != nil {
		log.Fatal(err)
	}
	png, raw, err := orsc.RenderDump(d)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outPath, png, 0o644); err != nil {
		log.Fatal(err)
	}

	w, h := d.Camera.ScreenW, d.Camera.ScreenH
	nonBg := 0
	for _, p := range raw {
		if p != backgroundColour {
			nonBg++
		}
	}
	pct := 100.0 * float64(nonBg) / float64(len(raw))
	fmt.Printf("rendered %s -> %s\n", inPath, outPath)
	fmt.Printf("  dimensions: %dx%d (%d pixels, %d bytes PNG)\n", w, h, len(raw), len(png))
	fmt.Printf("  non-background pixels: %d (%.1f%%)\n", nonBg, pct)
}
