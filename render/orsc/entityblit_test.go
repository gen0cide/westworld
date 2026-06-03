package orsc

import (
	"image/png"
	"os"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/render"
)

// TestRatLayerBlit exercises the faithful per-layer 16.16 spriteClippingLayer (the
// Phase 4 / Milestone C entity blit) in ISOLATION: it decodes the NPC-rat layer from
// content1, blits it into a bare surface at the projected fixture rect (197,121,118,46,
// skew=0 — the rect all three legs' rect-replicators compute), and (when a DEOB oracle
// PNG is supplied via DEOB_PNG) byte-compares every pixel the blit wrote against the
// DEOB render. The DEOB blits the rat on top of terrain at the END of its render, so a
// pixel the orsc blit writes must match the DEOB output exactly — this isolates the
// blit from any scene-occlusion/painter-order concerns. Skips when content1 is absent
// (cache-dependent). With DEOB_PNG present the gate is 0 diff px.
func TestRatLayerBlit(t *testing.T) {
	if os.Getenv("RSC_MESH_CACHE") == "" {
		os.Setenv("RSC_MESH_CACHE", "/tmp/rsc-run/cache")
	}
	f := &facts.Facts{NpcDefs: map[int]*facts.NpcDef{19: {Sprites: [12]int{123, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, Camera1: 346, Camera2: 136}}}
	es := render.NPCEntityLayers(f, 19, 0, 0)
	if es == nil || len(es.Layers) == 0 {
		t.Skip("content1 unavailable (set RSC_MESH_CACHE to the rev-235 cache dir)")
	}
	const W, H = 512, 334
	surf := NewSurface(W, H)
	for i := range surf.Pixels {
		surf.Pixels[i] = -1 // sentinel: untouched stays -1 so we see exactly what the blit wrote
	}
	sc := NewScene(surf, 4, 16, 4)
	sc.spriteClippingLayer(&es.Layers[0], 197, 121, 118, 46, false, 0)

	wrote := 0
	for _, p := range surf.Pixels {
		if p != -1 {
			wrote++
		}
	}
	if wrote == 0 {
		t.Fatalf("blit wrote no pixels")
	}
	t.Logf("blit wrote %d px", wrote)

	dp := os.Getenv("DEOB_PNG")
	if dp == "" {
		t.Skip("set DEOB_PNG to the DEOB rat render to assert 0-diff vs the oracle")
	}
	fh, err := os.Open(dp)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	dimg, err := png.Decode(fh)
	if err != nil {
		t.Fatal(err)
	}
	diff := 0
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			oc := surf.Pixels[y*W+x]
			if oc == -1 {
				continue // untouched by the blit; the DEOB shows terrain here
			}
			r, g, b, _ := dimg.At(x, y).RGBA()
			dc := int32((r>>8&0xff)<<16 | (g>>8&0xff)<<8 | (b >> 8 & 0xff))
			if oc != dc {
				diff++
				if diff <= 8 {
					t.Logf("diff x=%d y=%d orsc=%06x deob=%06x", x, y, oc&0xffffff, dc&0xffffff)
				}
			}
		}
	}
	if diff != 0 {
		t.Fatalf("isolated rat blit differs from DEOB oracle by %d px (expected 0)", diff)
	}
	t.Logf("isolated rat blit matches DEOB oracle: 0 diff px")
}
