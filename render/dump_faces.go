package render

import (
	"fmt"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
)

// BuiltFace is one face of a built GameModel, described in a camera-independent,
// engine-comparable way: by WHERE it sits in the world (rounded transformed
// centroid), HOW MANY vertices it has, and WHAT fill/back-fill it carries. This
// is the structural-diff key (RENDER_DIFF_DESIGN.md §5 diff 2): "is this
// face/wall present?". A missing door quad, a bridge deck face with the wrong
// fill, or a roof face that should be culled but isn't, all surface as a
// BuiltFace that is present in one engine's set and absent in another's —
// independent of any rasterization quirk.
//
// Centroid is the mean of the face's TRANSFORMED (world-space) vertices, AFTER
// each model's queued baseX/Y/Z + orientation transform but BEFORE the camera
// transform, so two engines that build the same geometry produce the same key
// regardless of camera. It is rounded to CentroidGrid units to absorb sub-unit
// integer-math drift between engines while staying far finer than a tile (128
// world units), so two distinct wall quads on the same tile never collide.
type BuiltFace struct {
	Model     int      `json:"model"`     // index of the source model in Scene.Models
	Centroid  [3]int32 `json:"centroid"`  // rounded transformed-space centroid (X,Y,Z)
	NumVerts  int      `json:"numVerts"`  // vertex count of the face
	FillFront int32    `json:"fillFront"` // FaceFillFront (<0 flat 5:5:5 colour, >=0 texture id)
	FillBack  int32    `json:"fillBack"`  // FaceFillBack
}

// CentroidGrid is the rounding granularity (world units) for a BuiltFace
// centroid. RSC tiles are 128 world units; 16 units is 1/8 of a tile — fine
// enough that two walls on one tile stay distinct, coarse enough that an
// engine's last-bit integer drift in a vertex coordinate doesn't move a face
// into a different bucket.
const CentroidGrid = int32(16)

// Key is the structural identity of a face: everything BUT the source-model
// index. Two faces with the same Key are "the same wall/quad" for the purpose
// of the present-in-A-but-not-B diff (the model index is build-order noise and
// must not be part of the identity).
func (bf BuiltFace) Key() string {
	return fmt.Sprintf("c=%d,%d,%d|n=%d|f=%d|b=%d",
		bf.Centroid[0], bf.Centroid[1], bf.Centroid[2], bf.NumVerts, bf.FillFront, bf.FillBack)
}

// RenderDumpFaces builds the SAME Scene render.RenderDump rasterizes — terrain,
// boundaries, roofs, scenery — from an L1 dump, then exports its built face set
// as []BuiltFace for the structural diff (RENDER_DIFF_DESIGN.md §5). It is the
// GO engine's "emit the built face/model list" entrypoint: an L2-style export
// derived from the L1 build, so renderdiff can structurally compare two dumps'
// geometry WITHOUT the other engines yet (the self-test renders with-door vs
// without-door and shows the door face present in one set, absent in the other).
//
// It uses syntheticFacts so a hand-authored fixture with no GameData still
// builds its wall/door (matching RenderDump); pass real facts via
// RenderDumpFacesWith for a real-map dump. The terrain seed is honoured for
// parity with RenderDump, though it only affects vertex SHADE, not geometry.
func RenderDumpFaces(d *rscdump.Dump) ([]BuiltFace, error) {
	return RenderDumpFacesWith(d, syntheticFacts(d), nil)
}

// RenderDumpFacesWith is RenderDumpFaces with caller-supplied facts + bundle,
// the structural-diff twin of RenderDumpWith.
func RenderDumpFacesWith(d *rscdump.Dump, f *facts.Facts, b *Bundle) ([]BuiltFace, error) {
	if d == nil {
		return nil, fmt.Errorf("render: nil dump")
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if d.Level != rscdump.LevelL1 {
		return nil, fmt.Errorf("render: RenderDumpFaces supports level %q, got %q", rscdump.LevelL1, d.Level)
	}
	land := d.Landscape()
	if land == nil {
		return nil, fmt.Errorf("render: dump has no terrain")
	}
	v := dumpView(d)

	var sc *Scene
	withDumpTerrainSeed(d.Terrain.TerrainSeed, func() {
		sc, _ = buildScene(land, f, b, v)
	})
	return sceneFaces(sc), nil
}

// sceneFaces walks every model in the scene and emits one BuiltFace per face,
// computing each face's centroid from the model's TRANSFORMED vertices (apply()
// bakes the queued baseX/Y/Z + orientation into tX/tY/tZ — the same transform
// the camera then projects). Degenerate faces (<3 verts) are skipped, matching
// the rasterizer's own guard.
func sceneFaces(sc *Scene) []BuiltFace {
	if sc == nil {
		return nil
	}
	var out []BuiltFace
	for mi, m := range sc.Models {
		if m == nil || m.NumFaces == 0 {
			continue
		}
		m.apply() // bake transform into tX/tY/tZ (idempotent re-run is fine)
		for fi := 0; fi < m.NumFaces; fi++ {
			verts := m.FaceVertices[fi]
			if len(verts) < 3 {
				continue
			}
			var sx, sy, sz int64
			for _, vtx := range verts {
				sx += int64(m.tX[vtx])
				sy += int64(m.tY[vtx])
				sz += int64(m.tZ[vtx])
			}
			n := int64(len(verts))
			out = append(out, BuiltFace{
				Model: mi,
				Centroid: [3]int32{
					roundTo(int32(sx/n), CentroidGrid),
					roundTo(int32(sy/n), CentroidGrid),
					roundTo(int32(sz/n), CentroidGrid),
				},
				NumVerts:  len(verts),
				FillFront: m.FaceFillFront[fi],
				FillBack:  m.FaceFillBack[fi],
			})
		}
	}
	return out
}

// roundTo rounds v to the nearest multiple of grid (grid>0), with correct
// rounding for negative values so a face just south of the origin doesn't snap
// the wrong way.
func roundTo(v, grid int32) int32 {
	if grid <= 1 {
		return v
	}
	if v >= 0 {
		return ((v + grid/2) / grid) * grid
	}
	return -(((-v + grid/2) / grid) * grid)
}
