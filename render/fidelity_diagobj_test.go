package render

import (
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/internal/rscdump"
)

// fidelity_diagobj_test guards the door-diag-object-not-built fix against the
// deob ground truth (World.addModels, World.java:792-820): RSC encodes a
// diagonally-PLACED scenery object — INCLUDING a diagonal door — as an id in the
// 48001..59999 band of the per-tile DiagonalWalls grid. The wall builder
// correctly DROPS that band (it is not a wall), so before BuildDiagonalObjects
// a diagonal door rendered as ZERO geometry. See
// reference/rsc-client/docs/build/RENDER_FIDELITY_FINDINGS.md.

// loadHuntFaces builds the GO face set for a hunt/*.json fixture (no archive).
func loadHuntFaces(t *testing.T, name string) []BuiltFace {
	t.Helper()
	p := filepath.Join("..", "testdata", "rscdump", "hunt", name)
	d, err := rscdump.Load(p)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	faces, err := RenderDumpFaces(d)
	if err != nil {
		t.Fatalf("RenderDumpFaces %s: %v", name, err)
	}
	return faces
}

// diagDoorLeafFaces counts the placed diagonal-object leaf faces: the synthetic
// wood door-leaf BuildDiagonalObjects builds for an object in the 48001..59999
// band that has no real model/def (the hand-authored fixture path). It is a
// single upright quad coloured with the flat WOOD colour (wallColourWood), the
// same colour the synthetic boundary door-leaf carries.
func diagDoorLeafFaces(faces []BuiltFace) int {
	n := 0
	for _, f := range faces {
		if f.NumVerts == 4 && f.FillFront == wallColourWood && f.FillBack == wallColourWood {
			n++
		}
	}
	return n
}

// TestFidelity_DiagObjectBuilt (door-diag-object-not-built /
// scenery-diag-object-not-built): a DiagonalWalls value in the object band
// (48001..59999) must build a PLACED model (World.addModels), not vanish. The
// fixture diagobj_door.json places a single diagonal-door object (wallDiag 48001
// => objectId 0) at world tile (208,209) with tileDirection 0 on flat grass; the
// empty control door_none.json has no walls/objects. We assert the diagonal door
// adds exactly one upright wood door-leaf the control lacks. Before the fix the
// two face sets were IDENTICAL (the 48000+ band was dropped and never built).
func TestFidelity_DiagObjectBuilt(t *testing.T) {
	diag := loadHuntFaces(t, "diagobj_door.json")
	none := loadHuntFaces(t, "door_none.json")

	gotDiag := diagDoorLeafFaces(diag)
	gotNone := diagDoorLeafFaces(none)

	if gotNone != 0 {
		t.Errorf("empty control door_none.json built %d door-leaf faces (want 0)", gotNone)
	}
	if gotDiag < 1 {
		t.Errorf("diagonal-object band (48001..59999) built %d door-leaf faces (want >=1); "+
			"the 48000+ DiagonalWalls band rendered as ZERO geometry — World.addModels not ported", gotDiag)
	}
	// The diagonal object must add net geometry the empty control lacks: the
	// whole face set is larger by the placed leaf (mirrors the structural diff
	// `diagobj=25282 faces, none=25281` => +1).
	if added := len(diag) - len(none); added != gotDiag {
		t.Errorf("diagonal object net face delta %d != door-leaf faces %d "+
			"(the placed model should be the ONLY difference vs the empty control)", added, gotDiag)
	}
}

// TestFidelity_DiagObjectVsStraightDoor sanity-checks that the diagonal-door
// object builds the SAME kind of geometry as a known-good straight (WallV) door:
// door_straight.json builds +1 wood leaf the control lacks, and so does the
// diagonal object. Both leaves are 4-vert wood quads — the diagonal door is no
// longer "a door the cradle can't perceive".
func TestFidelity_DiagObjectVsStraightDoor(t *testing.T) {
	straight := loadHuntFaces(t, "door_straight.json")
	diag := loadHuntFaces(t, "diagobj_door.json")

	leafStraight := diagDoorLeafFaces(straight)
	leafDiag := diagDoorLeafFaces(diag)

	if leafStraight < 1 {
		t.Fatalf("straight-door control built %d door-leaf faces (want >=1) — fixture/control changed?", leafStraight)
	}
	if leafDiag != leafStraight {
		t.Errorf("diagonal door leaf count %d != straight door leaf count %d "+
			"(the diagonal object should build the same kind of door geometry as a straight door)",
			leafDiag, leafStraight)
	}
}
