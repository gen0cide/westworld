// Package orsc is a faithful, line-for-line Go port of OpenRSC's three/
// software renderer. It is the SECOND, independent renderer in this repo: the
// existing github.com/gen0cide/westworld/render package ports the SAME
// algorithm from the classic mudclient, and this package exists so the two can
// render the same scene and be diffed pixel-for-pixel.
//
// # Source of truth
//
// Every type, field name, constant and method in this package mirrors the Java
// classes under, with line citations in the code:
//
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/RSModel.java
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/Scene.java
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/Shader.java
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/World.java
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/Polygon.java
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/Scanline.java
//
// The trig tables mirror orsc/util/FastMath.java (NOT the existing
// render/camera.go tables, which differ): trigTable256 is 512 entries
// (sin[0..255], cos[256..511]) and trigTable1024/trigTable_1024 are 2048
// entries (sin[0..1023], cos[1024..2047]). RSModel.rotate256 indexes
// trigTable256; RSModel.rotate1024 indexes trigTable1024; Scene.setCamera and
// Scene.setFrustum (the frustum-corner rotate) index trigTable_1024 — which is
// numerically identical to trigTable1024 but is a distinct table in Java, so we
// keep both names. See types.go.
//
// Where this package CANONICALLY diverges from the Java (and why):
//   - Java int is 32-bit. The fixed-point pipeline (>>, *, the perspective
//     divide) is therefore int32. Java's >>> (logical right shift) becomes a
//     uint32 shift in the texture/shade-mip code (Shader). Signed bytes from the
//     .ob3 / landscape streams are masked &0xff. See the Java->Go notes on each
//     field.
//   - The AWT ImageProducer handoff (GraphicsController.pixelData -> Image) is
//     replaced by Surface, whose Pixels []int32 buffer IS the
//     GraphicsController.pixelData int[] the Java renderer writes into; toImage
//     encodes it to a Go image.Image / PNG.
//   - Asset loading uses the existing westworld decoders rather than the Java
//     ZipFile/DataInputStream paths: .ob3 models via
//     assets.OpenArchive(models.orsc).Get(name) + assets.DecodeModel; the
//     landscape via pathfind.OpenLandscape(Authentic_Landscape.orsc).Tile; and
//     textures/sprites via assets.OpenSprites(Authentic_Sprites.orsc).Sprite.
//     NewModel adapts an *assets.Model into the RSModel layout below.
//
// # Pipeline (mirrors mudclient.drawScene -> Scene.endScene)
//
//  1. Build geometry. World.generateLandscapeModel emits terrain/wall/roof faces
//     into accumulator RSModels (NewModel for scenery/entities from .ob3). Each
//     RSModel carries model-local vertX/vertY/vertZ plus an orientation
//     (rot256X/Y/Z), scale and translate that resetTransformCache bakes into the
//     vertXTransform/Y/Z working arrays, after which calculateBoundingBoxes and
//     computeNormals/computeDiffuse run (Translate/Orient/computeBounds/light).
//  2. Add to the Scene. Scene.AddModel appends to the model list (capacity m_u).
//  3. Camera. Scene.SetCamera computes the rot1024 eye offset + the cameraProjX/
//     Y/Z inverse angles fed to RSModel.rotate1024.
//  4. Render (Scene.endScene). For every model rotate1024 transforms the baked
//     verts into camera space (vertXRot/YRot/ZRot) and projects them through the
//     rot1024_vp_src (=9) shift / perspective divide into vertexParam6 (screen X)
//     and vertexParam2 (screen Y). Visible faces become Polygons; computePolygon
//     derives each face's camera-space plane (normalX/Y/Z, orientation,
//     scenePolyNormalMagnitude) and screen/cam-Z AABB. The Polygons are
//     coarse-sorted by m_t (average camera-Z) then re-ordered by the exact 3D
//     plane-overlap resolution (the polygonHit1/polygonHit2/booleanCombinatoric
//     family) so wall corners paint in the right order. Each face is then scanned
//     (setFrustum builds the per-row Scanline m_d/m_k left/right spans) and filled
//     by Shader.shadeScanline — flat 5:5:5 colour, or a 5:5:5-quantised,
//     0xf8f8ff-masked, shade-mipped texture (the resourceDatabase[id] built by
//     Scene.b/setFrustum), with TRANSPARENT (=12345678) texels skipped.
//
// project() in this port is RSModel.rotate1024's per-vertex body; light() is
// computeNormals+computeDiffuse; computeBounds is calculateBoundingBoxes;
// Translate/Orient set translate*/rot256* and mark the transform dirty (m_Yb).
//
// This file and types.go compile on their own (the method bodies are stubs);
// the parallel port files supply the bodies against the contract documented in
// types.go.
package orsc
