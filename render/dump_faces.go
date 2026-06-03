package render

import "fmt"

// dump_faces.go now holds ONLY the structural-diff SCHEMA — BuiltFace + its Key +
// CentroidGrid + RoundToGrid. These are the JAR-oracle-coupled types: the rev-235
// jar harness (rscplus/dumprender/DumpRenderer.java:249,282-284) hardcodes
// CentroidGrid=16 and the BuiltFace json field names, so they MUST stay
// byte-stable and live in this leaf package (no renderer imports).
//
// The face PRODUCER moved to render/orsc (orsc.RenderDumpFaces /
// RenderDumpFacesWith): the orsc faithful renderer imports this package for
// render.View / render.BuiltFace, so the dump→faces entrypoint cannot live here
// (render → render/orsc would be an import cycle). orsc walks its own
// Scene.models and emits []render.BuiltFace using this exact type, so the diff
// key, the JAR sidecar schema, and cmd/renderdiff are all unchanged — only the
// engine behind RenderDumpFaces is now orsc instead of the deleted classic Scene.

// BuiltFace is one face of a built model, described in a camera-independent,
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
// into a different bucket. The JAR oracle (DumpRenderer.java:249) hardcodes 16.
const CentroidGrid = int32(16)

// Key is the structural identity of a face: everything BUT the source-model
// index. Two faces with the same Key are "the same wall/quad" for the purpose
// of the present-in-A-but-not-B diff (the model index is build-order noise and
// must not be part of the identity).
func (bf BuiltFace) Key() string {
	return fmt.Sprintf("c=%d,%d,%d|n=%d|f=%d|b=%d",
		bf.Centroid[0], bf.Centroid[1], bf.Centroid[2], bf.NumVerts, bf.FillFront, bf.FillBack)
}

// RoundToGrid rounds v to the nearest multiple of grid (grid>0), with correct
// rounding for negative values so a face just south of the origin doesn't snap
// the wrong way. Exported so the face producer (render/orsc) rounds centroids
// IDENTICALLY to the old in-package sceneFaces — the diff key stays byte-stable.
func RoundToGrid(v, grid int32) int32 {
	if grid <= 1 {
		return v
	}
	if v >= 0 {
		return ((v + grid/2) / grid) * grid
	}
	return -(((-v + grid/2) / grid) * grid)
}
