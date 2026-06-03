package main

import (
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
	"github.com/gen0cide/westworld/render/orsc"
)

// repoRel resolves a path relative to the module root (this test lives at
// cmd/renderdiff, so the root is two dirs up).
func repoRel(t *testing.T, rel ...string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(append([]string{"..", ".."}, rel...)...))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func renderFixture(t *testing.T, name string) ([]int32, []render.BuiltFace) {
	t.Helper()
	d, err := rscdump.Load(repoRel(t, "testdata", "rscdump", name))
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	_, raw, err := orsc.RenderDump(d)
	if err != nil {
		t.Fatalf("render %s: %v", name, err)
	}
	faces, err := orsc.RenderDumpFaces(d)
	if err != nil {
		t.Fatalf("faces %s: %v", name, err)
	}
	return raw, faces
}

// TestSelfTest_DoorDetection is the renderdiff self-test (RENDER_DIFF_DESIGN.md
// §5): rendering the single-tile fixture WITH the door vs WITHOUT it must show a
// non-zero pixel-diff region AND a structural diff with exactly the door face
// present in one set and absent in the other — i.e. the tool detects the
// "unwalkable door / unperceived wall" bug class without the Java engines.
func TestSelfTest_DoorDetection(t *testing.T) {
	doorPix, doorFaces := renderFixture(t, "single_tile_door.json")
	noDoorPix, noDoorFaces := renderFixture(t, "single_tile_NOdoor.json")

	w, h := 512, 334

	// Pixel diff: must be non-zero and localized (a bounded region, not the
	// whole frame — the change is one wall quad).
	pd, err := ComparePixels(doorPix, noDoorPix, w, h, 1)
	if err != nil {
		t.Fatal(err)
	}
	if pd.Differing == 0 {
		t.Fatalf("door detection FAILED: pixel diff is 0 (the removed wall must change pixels)")
	}
	if pd.PctDiffering() > 10 {
		t.Errorf("pixel diff not localized: %.2f%% of the frame differs (expected a small wall region)", pd.PctDiffering())
	}
	regionW := pd.MaxX - pd.MinX + 1
	regionH := pd.MaxY - pd.MinY + 1
	if regionW <= 0 || regionH <= 0 || regionW > w/2 || regionH > h/2 {
		t.Errorf("pixel diff region [%d..%d]x[%d..%d] is not a localized wall region", pd.MinX, pd.MaxX, pd.MinY, pd.MaxY)
	}
	t.Logf("pixel diff: %d/%d px (%.3f%%) maxΔ=%d region=[x %d..%d, y %d..%d]",
		pd.Differing, pd.Total, pd.PctDiffering(), pd.MaxDelta, pd.MinX, pd.MaxX, pd.MinY, pd.MaxY)

	// Structural diff: exactly the door face is present in with-door, absent in
	// without-door, and no spurious extras in the other direction.
	sd := CompareFaces(doorFaces, noDoorFaces)
	if len(sd.OnlyInA) != 1 {
		t.Fatalf("door detection FAILED: expected exactly 1 face only-in-with-door, got %d: %+v", len(sd.OnlyInA), sd.OnlyInA)
	}
	if len(sd.OnlyInB) != 0 {
		t.Errorf("unexpected %d face(s) only-in-without-door (removing a wall should not ADD faces): %+v", len(sd.OnlyInB), sd.OnlyInB)
	}
	door := sd.OnlyInA[0]
	if door.NumVerts != 4 {
		t.Errorf("the door face should be a quad (4 verts), got %d", door.NumVerts)
	}
	t.Logf("structural diff: door face present-in-door/absent-in-nodoor = %+v (%d shared)", door, sd.Shared)
}

// TestSelfTest_ZeroDiffControl is the false-positive control: diffing a render
// against ITSELF must report zero pixel diff and zero structural diff. This
// proves the determinism rules (§4) hold — two independent renders of the same
// dump are byte-identical — and that the tool does not flag spurious diffs.
func TestSelfTest_ZeroDiffControl(t *testing.T) {
	a, aFaces := renderFixture(t, "single_tile_door.json")
	b, bFaces := renderFixture(t, "single_tile_door.json")

	pd, err := ComparePixels(a, b, 512, 334, 1)
	if err != nil {
		t.Fatal(err)
	}
	if pd.Differing != 0 || pd.MaxDelta != 0 {
		t.Fatalf("false positive: identical renders differ (%d px, maxΔ=%d) — determinism broken", pd.Differing, pd.MaxDelta)
	}

	sd := CompareFaces(aFaces, bFaces)
	if len(sd.OnlyInA) != 0 || len(sd.OnlyInB) != 0 {
		t.Fatalf("false positive: identical face sets differ (%d only-A, %d only-B)", len(sd.OnlyInA), len(sd.OnlyInB))
	}
	t.Logf("zero-diff control: 0 pixel diff, %d faces both, %d shared", sd.CountA, sd.Shared)
}
