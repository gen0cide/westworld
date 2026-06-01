package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"

	"github.com/gen0cide/westworld/render"
)

// ---------------------------------------------------------------------------
// Pixel diff (RENDER_DIFF_DESIGN.md §5 diff 1)
// ---------------------------------------------------------------------------

// PixelDiff is the result of comparing two equal-size framebuffers.
type PixelDiff struct {
	Width, Height int
	Total         int // total pixels
	Differing     int // pixels whose max channel |Δ| > Tolerance
	MaxDelta      int // largest single-channel |Δ| seen anywhere
	Tolerance     int // per-channel LSB tolerance applied (Δ<=tol is "same")

	// BBox is the axis-aligned bounding box (minX,minY,maxX,maxY inclusive) of
	// the differing pixels — the localized region a structural change shows up
	// in. Zeroed (all -1) when nothing differs.
	MinX, MinY, MaxX, MaxY int
}

// PctDiffering is the share of pixels that differ, in percent.
func (d PixelDiff) PctDiffering() float64 {
	if d.Total == 0 {
		return 0
	}
	return 100 * float64(d.Differing) / float64(d.Total)
}

// channels splits a 0x00RRGGBB pixel into its three 8-bit channels.
func channels(p int32) (r, g, b int) {
	return int(p>>16) & 0xff, int(p>>8) & 0xff, int(p) & 0xff
}

// absDelta is the max per-channel absolute difference between two pixels.
func absDelta(a, b int32) int {
	ar, ag, ab := channels(a)
	br, bg, bb := channels(b)
	m := abs(ar - br)
	if d := abs(ag - bg); d > m {
		m = d
	}
	if d := abs(ab - bb); d > m {
		m = d
	}
	return m
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ComparePixels diffs two equal-size int32 framebuffers. A pixel "differs" when
// its max per-channel |Δ| exceeds tol (so a 1-LSB rounding wobble is tolerated,
// per the design's ≤1 LSB tolerance). It also records the max delta and the
// bounding box of the differing region.
func ComparePixels(a, b []int32, w, h, tol int) (PixelDiff, error) {
	if len(a) != len(b) {
		return PixelDiff{}, fmt.Errorf("renderdiff: pixel buffers differ in size (%d vs %d)", len(a), len(b))
	}
	if w*h != len(a) {
		return PixelDiff{}, fmt.Errorf("renderdiff: %dx%d != buffer len %d", w, h, len(a))
	}
	d := PixelDiff{
		Width: w, Height: h, Total: len(a), Tolerance: tol,
		MinX: -1, MinY: -1, MaxX: -1, MaxY: -1,
	}
	for i := range a {
		delta := absDelta(a[i], b[i])
		if delta > d.MaxDelta {
			d.MaxDelta = delta
		}
		if delta > tol {
			d.Differing++
			x, y := i%w, i/w
			if d.MinX < 0 || x < d.MinX {
				d.MinX = x
			}
			if d.MaxX < 0 || x > d.MaxX {
				d.MaxX = x
			}
			if d.MinY < 0 || y < d.MinY {
				d.MinY = y
			}
			if d.MaxY < 0 || y > d.MaxY {
				d.MaxY = y
			}
		}
	}
	return d, nil
}

// Heatmap renders a diff PNG: pixels within tolerance are drawn as a dimmed
// grayscale of A (so you can see the scene as context), and pixels that differ
// by more than tol are painted bright magenta scaled by delta magnitude (the
// "highlight |Δ|>threshold" of §5). Returns the encoded PNG bytes.
func Heatmap(a, b []int32, w, h, tol int) ([]byte, error) {
	if len(a) != len(b) || w*h != len(a) {
		return nil, fmt.Errorf("renderdiff: heatmap size mismatch")
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range a {
		x, y := i%w, i/w
		delta := absDelta(a[i], b[i])
		if delta <= tol {
			// context: dim grayscale of A's luminance
			r, g, bb := channels(a[i])
			lum := uint8((r*30 + g*59 + bb*11) / 100 / 3)
			img.Set(x, y, color.RGBA{R: lum, G: lum, B: lum, A: 255})
			continue
		}
		// highlight: magenta, intensity scales with delta (clamped)
		mag := delta
		if mag > 255 {
			mag = 255
		}
		img.Set(x, y, color.RGBA{R: uint8(mag), G: 0, B: uint8(mag), A: 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodePNG reads a PNG file and returns its pixels as a 0x00RRGGBB int32
// framebuffer plus dimensions — the inverse of render.Surface.PNG, so a
// pre-rendered PNG (e.g. a future jar/deob capture) can be diffed against a GO
// render without the raw buffer.
func DecodePNG(path string) ([]int32, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("renderdiff: decode %q: %w", path, err)
	}
	bnd := img.Bounds()
	w, h := bnd.Dx(), bnd.Dy()
	pix := make([]int32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(bnd.Min.X+x, bnd.Min.Y+y).RGBA()
			pix[y*w+x] = int32(r>>8)<<16 | int32(g>>8)<<8 | int32(b>>8)
		}
	}
	return pix, w, h, nil
}

// ---------------------------------------------------------------------------
// Structural diff (RENDER_DIFF_DESIGN.md §5 diff 2)
// ---------------------------------------------------------------------------

// StructDiff is the result of comparing two built face sets by structural key.
type StructDiff struct {
	CountA  int                // faces in A
	CountB  int                // faces in B
	OnlyInA []render.BuiltFace // present in A, absent in B (e.g. door wall that exists in A)
	OnlyInB []render.BuiltFace // present in B, absent in A
	Shared  int                // faces with a matching key in both
}

// CompareFaces diffs two face sets by BuiltFace.Key(). A face counts as
// "shared" when its key appears in both sets (multiplicity-aware: a key present
// twice in A and once in B leaves one A-only). The result lists the
// asymmetric-difference faces — exactly the missing/extra walls a perception
// bug produces.
func CompareFaces(a, b []render.BuiltFace) StructDiff {
	bCount := map[string]int{}
	bFace := map[string]render.BuiltFace{}
	for _, f := range b {
		bCount[f.Key()]++
		bFace[f.Key()] = f
	}
	aCount := map[string]int{}
	aFace := map[string]render.BuiltFace{}
	for _, f := range a {
		aCount[f.Key()]++
		aFace[f.Key()] = f
	}

	d := StructDiff{CountA: len(a), CountB: len(b)}
	// A-only: each occurrence in A beyond B's count.
	for k, ac := range aCount {
		bc := bCount[k]
		if ac > bc {
			for i := 0; i < ac-bc; i++ {
				d.OnlyInA = append(d.OnlyInA, aFace[k])
			}
		}
		if bc < ac {
			d.Shared += bc
		} else {
			d.Shared += ac
		}
	}
	// B-only: each occurrence in B beyond A's count.
	for k, bc := range bCount {
		ac := aCount[k]
		if bc > ac {
			for i := 0; i < bc-ac; i++ {
				d.OnlyInB = append(d.OnlyInB, bFace[k])
			}
		}
	}
	sortFaces(d.OnlyInA)
	sortFaces(d.OnlyInB)
	return d
}

func sortFaces(fs []render.BuiltFace) {
	sort.Slice(fs, func(i, j int) bool { return fs[i].Key() < fs[j].Key() })
}
