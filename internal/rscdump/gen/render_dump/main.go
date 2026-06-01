// Command render_dump renders an rscdump/1 L1 fixture through the GO engine
// (render.RenderDump) to a PNG, and prints sanity stats (dimensions + count of
// non-sky pixels). This is the Phase-0 "dump → PNG" smoke test
// (RENDER_DIFF_DESIGN.md §6).
//
//	go run ./internal/rscdump/gen/render_dump <dump.json> <out.png>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

// skyColour mirrors render.skyColour (unexported): the flat background fill a
// pixel keeps when no geometry covers it. Used to count non-background pixels.
const skyColour = int32(0x6080a0)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <dump.json> <out.png>", os.Args[0])
	}
	inPath, outPath := os.Args[1], os.Args[2]

	d, err := rscdump.Load(inPath)
	if err != nil {
		log.Fatal(err)
	}
	png, raw, err := render.RenderDump(d)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outPath, png, 0o644); err != nil {
		log.Fatal(err)
	}

	w, h := d.Camera.ScreenW, d.Camera.ScreenH
	nonSky := 0
	for _, p := range raw {
		if p != skyColour {
			nonSky++
		}
	}
	pct := 100.0 * float64(nonSky) / float64(len(raw))
	fmt.Printf("rendered %s -> %s\n", inPath, outPath)
	fmt.Printf("  dimensions: %dx%d (%d pixels, %d bytes PNG)\n", w, h, len(raw), len(png))
	fmt.Printf("  non-sky pixels: %d (%.1f%%)\n", nonSky, pct)
}
