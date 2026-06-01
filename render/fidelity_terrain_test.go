package render

import (
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/internal/rscdump"
)

// fidelity_terrain_test guards the TIER-3 terrain render-fidelity fixes against
// the deob ground truth (World.java buildSection / buildOverlayTriangles). Each
// test loads a hunt/t3_*.json fixture, builds the GO face set, and asserts the
// CORRECTED geometry the deob algorithm produces. See
// reference/rsc-client/docs/build/RENDER_FIDELITY_FINDINGS.md.

func loadT3Faces(t *testing.T, name string) []BuiltFace {
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

// TestFidelity_WaterFlattenByTileType (terrain-water-flatten-overlay-id /
// bridge_flatten_overlay2): overlay 2 is tileType 3, NOT 4, so the deob does NOT
// flatten it to y=0 — it renders as a textured type-3 overlay floor at real
// elevation (World.java:891-895). The fixture sits at uniform elevation 40 =>
// world height 120 => centroid Y -128 (16-grid). Before the fix every overlay-2
// face was a fixed-water quad at centroid Y 0 (sunk into the river trench).
func TestFidelity_WaterFlattenByTileType(t *testing.T) {
	faces := loadT3Faces(t, "t3_water_ov2.json")
	// No overlay-2 face may sit at the flattened water plane (centroid Y 0) inside
	// the patch; the patch must render at its real elevation (Y -128).
	sunkAtZero := 0
	atElevation := 0
	for _, f := range faces {
		// the overlay-2 patch renders with water texture id 1 in the back slot.
		if f.FillBack != 1 && f.FillFront != 1 {
			continue
		}
		switch f.Centroid[1] {
		case 0:
			sunkAtZero++
		case -128:
			atElevation++
		}
	}
	if sunkAtZero != 0 {
		t.Errorf("overlay-2 (tileType 3) wrongly flattened to y=0: %d water faces at centroid Y 0 (want 0)", sunkAtZero)
	}
	if atElevation == 0 {
		t.Errorf("overlay-2 patch not rendered at its real elevation (want texture-1 faces at centroid Y -128, got %d)", atElevation)
	}
}

// TestFidelity_Type4StillFlattens guards the OTHER half of the same fix: a
// GENUINE type-4 tile (ids 4/12/20/21) MUST still flatten to y=0 and emit a deck
// quad at raw elevation (the Lumbridge bridge double-render, World.java:704-723).
func TestFidelity_Type4StillFlattens(t *testing.T) {
	faces := loadT3Faces(t, "t3_bridge_ov4.json")
	waterAtZero := 0 // flattened water-shortcut quads
	deckAtElevation := 0
	for _, f := range faces {
		if f.FillFront == 1 && f.Centroid[1] == 0 {
			waterAtZero++
		}
		// id-4 deck colour is texture 3 (planks) in the back slot, at raw height.
		if f.FillBack == 3 && f.Centroid[1] != 0 {
			deckAtElevation++
		}
	}
	if waterAtZero == 0 {
		t.Errorf("type-4 band lost its flattened water plane (want fixed-water faces at centroid Y 0)")
	}
	if deckAtElevation == 0 {
		t.Errorf("type-4 band lost its raised deck quads (want planks faces above y=0)")
	}
}

// TestFidelity_IndoorFloorNotChamfered (terrain-type2-floor-split): a tileType-2
// indoor floor with no interior diagonal wall is NOT colour-split (World.java:932
// `decoType != 2`), so its corners stay square — no grass-revert (underlay)
// triangle inside the floor patch. The fixture is a 7x7 block of planks (id 3,
// tileType 2) in grass; before the fix the 4 extreme corners chamfered into
// grass wedges.
func TestFidelity_IndoorFloorNotChamfered(t *testing.T) {
	faces := loadT3Faces(t, "t3_indoor_split.json")
	const grassUnderlay = int32(-2625)
	// planks fill span ~ 9792..10560 world units; look for grass-revert triangles
	// inside that bbox.
	chamfers := 0
	planksQuads := 0
	for _, f := range faces {
		inBBox := f.Centroid[0] > 9700 && f.Centroid[0] < 10600 &&
			f.Centroid[2] > 9700 && f.Centroid[2] < 10600
		if !inBBox {
			continue
		}
		if f.NumVerts == 3 && f.FillBack == grassUnderlay {
			chamfers++
		}
		// planks = texture id 3 in the back slot, full quad.
		if f.NumVerts == 4 && f.FillBack == 3 {
			planksQuads++
		}
	}
	if chamfers != 0 {
		t.Errorf("indoor floor chamfered: %d grass-revert triangles inside the floor patch (want 0)", chamfers)
	}
	if planksQuads == 0 {
		t.Errorf("indoor floor not rendered as solid quads (want planks quads, got 0)")
	}
}

// TestFidelity_SameClassCornerNotSplit (terrain-split-id-vs-class): two DIFFERENT
// overlays of the SAME tileType class (id 7 and id 19, both tileType 3) meeting
// at a concave corner must NOT split — the deob compares the getTileType CLASS
// (0 == 0), not the raw id (World.java:933-942). The fixture's concave corner is
// tile (12,12), model corner ~ (10240,10240). Before the fix GO cut a grass
// wedge there.
func TestFidelity_SameClassCornerNotSplit(t *testing.T) {
	faces := loadT3Faces(t, "t3_class_corner.json")
	const grassUnderlay = int32(-2625)
	wedges := 0
	for _, f := range faces {
		if f.NumVerts != 3 || f.FillBack != grassUnderlay {
			continue
		}
		if f.Centroid[0] > 10180 && f.Centroid[0] < 10320 &&
			f.Centroid[2] > 10180 && f.Centroid[2] < 10320 {
			wedges++
		}
	}
	if wedges != 0 {
		t.Errorf("same-tileType-class concave corner wrongly split: %d grass wedges near tile(12,12) (want 0)", wedges)
	}
}

// TestFidelity_Type4NeighbourSpread (terrain-overlay-no-neighbour-spread): an
// isolated type-4 tile emits FIVE overlay quads (the tile + its 4 cardinal
// neighbours), per buildOverlayTriangles (World.java:1117-1139). The fixture is a
// single id-4 (tileType 4) tile in grass; id-4 deck colour is texture 3 (planks)
// in the back slot. Before the fix GO emitted only 1.
func TestFidelity_Type4NeighbourSpread(t *testing.T) {
	faces := loadT3Faces(t, "t3_type4_neighbour.json")
	overlayQuads := 0
	for _, f := range faces {
		if f.NumVerts == 4 && f.FillBack == 3 {
			overlayQuads++
		}
	}
	if overlayQuads != 5 {
		t.Errorf("type-4 overlay neighbour-spread wrong: got %d type-4 overlay quads, want 5 (tile + 4 neighbours)", overlayQuads)
	}
}
