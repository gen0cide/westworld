// Command renderdiff is the §5 render-diff harness for the RSC 3-way render
// diff (RENDER_DIFF_DESIGN.md). It compares two (or three) renders of the SAME
// dumped world state and reports where they disagree, in two complementary
// modes:
//
//	PIXEL DIFF      per-pixel max-channel delta, count + % of differing pixels
//	               (with a small LSB tolerance), and a diff HEATMAP PNG per pair.
//	STRUCTURAL DIFF the SET of built faces/walls each engine produced, keyed by
//	               (rounded centroid world-XYZ, vertex count, fill id). Reports
//	               faces present in A-but-not-B and vice-versa — this is what
//	               pinpoints a missing door quad / broken bridge / unculled roof.
//
// Each input is one of:
//   - an rscdump/1 L1 JSON  (.json) — rendered through the GO engine
//     (render.RenderDump) to a framebuffer AND a built-face set, so a single
//     dump drives both diff modes.
//   - a pre-rendered PNG    (.png)  — decoded for the pixel diff; an optional
//     sidecar "<name>.faces.json" supplies its face set for the structural diff.
//     This is the seam the DEOB and JAR engines plug into: each emits a PNG +
//     a faces.json in the BuiltFace schema and renderdiff diffs them unchanged.
//
// Usage:
//
//	renderdiff -out DIR [-tol N] LABEL=PATH LABEL=PATH [LABEL=PATH]
//
//	e.g.  renderdiff -out testdata/rscdump/out \
//		        go-door=testdata/rscdump/single_tile_door.json \
//		        go-nodoor=testdata/rscdump/single_tile_NOdoor.json
//
// It prints a one-line verdict per pair, writes a heatmap PNG per pair into
// -out, and exits non-zero if any pair shows a pixel OR structural diff (so it
// doubles as a CI gate once the golden fixtures are in).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

// input is one labelled render source, with its framebuffer + face set resolved.
type input struct {
	Label    string
	Path     string
	Pix      []int32
	W, H     int
	Faces    []render.BuiltFace
	HasFaces bool // whether a face set is available for the structural diff
}

func main() {
	out := flag.String("out", ".", "directory for diff heatmap PNGs + rendered artifacts")
	tol := flag.Int("tol", 1, "per-channel LSB tolerance: a pixel differs only when max |Δ| exceeds this")
	emitRender := flag.Bool("emit-render", true, "for .json inputs, also write the rendered PNG into -out")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 || len(args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: renderdiff -out DIR [-tol N] LABEL=PATH LABEL=PATH [LABEL=PATH]")
		os.Exit(2)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}

	var inputs []input
	for _, a := range args {
		eq := strings.IndexByte(a, '=')
		if eq <= 0 {
			fail(fmt.Errorf("input %q is not LABEL=PATH", a))
		}
		label, path := a[:eq], a[eq+1:]
		in, err := loadInput(label, path, *out, *emitRender)
		if err != nil {
			fail(err)
		}
		inputs = append(inputs, in)
	}

	// Compare every unordered pair.
	anyDiff := false
	fmt.Printf("renderdiff: %d inputs, tolerance=%d LSB\n", len(inputs), *tol)
	for i := 0; i < len(inputs); i++ {
		for j := i + 1; j < len(inputs); j++ {
			d, err := diffPair(inputs[i], inputs[j], *out, *tol)
			if err != nil {
				fail(err)
			}
			if d {
				anyDiff = true
			}
		}
	}
	if anyDiff {
		os.Exit(1)
	}
}

// loadInput resolves a LABEL=PATH input into a framebuffer + (when available) a
// built-face set. A .json path is rendered through the GO engine; a .png path
// is decoded, and its sidecar "<base>.faces.json" is loaded if present.
func loadInput(label, path, outDir string, emitRender bool) (input, error) {
	in := input{Label: label, Path: path}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		d, err := rscdump.Load(path)
		if err != nil {
			return in, err
		}
		pngBytes, raw, err := render.RenderDump(d)
		if err != nil {
			return in, fmt.Errorf("render %s: %w", path, err)
		}
		faces, err := render.RenderDumpFaces(d)
		if err != nil {
			return in, fmt.Errorf("faces %s: %w", path, err)
		}
		in.Pix = raw
		in.W, in.H = d.Camera.ScreenW, d.Camera.ScreenH
		in.Faces = faces
		in.HasFaces = true
		if emitRender {
			rp := filepath.Join(outDir, label+".png")
			if err := os.WriteFile(rp, pngBytes, 0o644); err != nil {
				return in, err
			}
			fp := filepath.Join(outDir, label+".faces.json")
			if err := saveFaces(fp, faces); err != nil {
				return in, err
			}
		}
	case ".png":
		pix, w, h, err := DecodePNG(path)
		if err != nil {
			return in, err
		}
		in.Pix, in.W, in.H = pix, w, h
		// Optional sidecar face set: same base name with .faces.json.
		base := strings.TrimSuffix(path, filepath.Ext(path))
		if faces, ok, err := loadFaces(base + ".faces.json"); err != nil {
			return in, err
		} else if ok {
			in.Faces = faces
			in.HasFaces = true
		}
	default:
		return in, fmt.Errorf("input %q: unsupported extension (want .json or .png)", path)
	}
	return in, nil
}

// diffPair runs both diff modes on one pair, prints the verdict, writes the
// heatmap, and returns whether the pair differs (pixel OR structural).
func diffPair(a, b input, outDir string, tol int) (bool, error) {
	pairName := a.Label + "_vs_" + b.Label
	fmt.Printf("\n=== %s ===\n", pairName)

	differs := false

	// --- pixel diff ---
	if a.W != b.W || a.H != b.H {
		fmt.Printf("  PIXEL: SIZE MISMATCH %s=%dx%d %s=%dx%d (cannot pixel-diff)\n",
			a.Label, a.W, a.H, b.Label, b.W, b.H)
		differs = true
	} else {
		pd, err := ComparePixels(a.Pix, b.Pix, a.W, a.H, tol)
		if err != nil {
			return false, err
		}
		hm, err := Heatmap(a.Pix, b.Pix, a.W, a.H, tol)
		if err != nil {
			return false, err
		}
		hmPath := filepath.Join(outDir, "heatmap_"+pairName+".png")
		if err := os.WriteFile(hmPath, hm, 0o644); err != nil {
			return false, err
		}
		if pd.Differing == 0 {
			fmt.Printf("  PIXEL: IDENTICAL (0 differing, maxΔ=%d) heatmap=%s\n", pd.MaxDelta, hmPath)
		} else {
			differs = true
			fmt.Printf("  PIXEL: DIFFER %d/%d px (%.3f%%) maxΔ=%d, region [x %d..%d, y %d..%d] heatmap=%s\n",
				pd.Differing, pd.Total, pd.PctDiffering(), pd.MaxDelta,
				pd.MinX, pd.MaxX, pd.MinY, pd.MaxY, hmPath)
		}
	}

	// --- structural diff ---
	if !a.HasFaces || !b.HasFaces {
		fmt.Printf("  STRUCT: skipped (no face set for %s)\n", missingFaceLabel(a, b))
	} else {
		sd := CompareFaces(a.Faces, b.Faces)
		if len(sd.OnlyInA) == 0 && len(sd.OnlyInB) == 0 {
			fmt.Printf("  STRUCT: IDENTICAL (%d faces both, %d shared)\n", sd.CountA, sd.Shared)
		} else {
			differs = true
			fmt.Printf("  STRUCT: DIFFER  %s=%d faces, %s=%d faces, %d shared\n",
				a.Label, sd.CountA, b.Label, sd.CountB, sd.Shared)
			printFaces(fmt.Sprintf("    only in %s (present here, ABSENT in %s)", a.Label, b.Label), sd.OnlyInA)
			printFaces(fmt.Sprintf("    only in %s (present here, ABSENT in %s)", b.Label, a.Label), sd.OnlyInB)
		}
	}

	return differs, nil
}

func printFaces(header string, fs []render.BuiltFace) {
	if len(fs) == 0 {
		return
	}
	fmt.Println(header + ":")
	for _, f := range fs {
		fmt.Printf("      face model#%d centroid=(%d,%d,%d) verts=%d fillFront=%d fillBack=%d\n",
			f.Model, f.Centroid[0], f.Centroid[1], f.Centroid[2], f.NumVerts, f.FillFront, f.FillBack)
	}
}

func missingFaceLabel(a, b input) string {
	switch {
	case !a.HasFaces && !b.HasFaces:
		return a.Label + " and " + b.Label
	case !a.HasFaces:
		return a.Label
	default:
		return b.Label
	}
}

// --- face-set sidecar JSON (the schema the DEOB/JAR engines also emit) ---

type facesFile struct {
	Schema string             `json:"schema"`
	Faces  []render.BuiltFace `json:"faces"`
}

const facesSchema = "rscdump-faces/1"

func saveFaces(path string, faces []render.BuiltFace) error {
	b, err := json.MarshalIndent(facesFile{Schema: facesSchema, Faces: faces}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func loadFaces(path string) ([]render.BuiltFace, bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var ff facesFile
	if err := json.Unmarshal(b, &ff); err != nil {
		return nil, false, fmt.Errorf("renderdiff: parse face set %q: %w", path, err)
	}
	return ff.Faces, true, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "renderdiff:", err)
	os.Exit(2)
}
