package orsc

import "math"

// =============================================================================
// Constants (Scene.java:11, RSModel.java:13, RSModel.java:64)
// =============================================================================

// TRANSPARENT is the fill sentinel meaning "no face / fully transparent".
// Scene.java:11 `static final int TRANSPARENT = 12345678;`. In RSModel it is the
// private field m_Vb (RSModel.java:13 `private final int m_Vb = 12345678;`) and
// is substituted for the 32767 fill marker read from the .ob3 stream
// (RSModel.java:125-126,135-137) and for faceDiffuseLight that must be
// per-vertex gouraud (RSModel.java:144,648).
const TRANSPARENT = 12345678

// maxFaceDimensionInit is RSModel.maxFaceDimension's declared value
// (RSModel.java:64). It also doubles as the "no max yet" seed flipped negative
// in calculateBoundingBoxes (RSModel.java:439).
const maxFaceDimensionInit = 12345678

// polyNormalScale is Scene.polyNormalScale (Scene.java:15, set =4 at
// Scene.java:90): the multiplier applied to a face's normal magnitude in
// computePolygon (Scene.java:514).
const polyNormalScale = 4

// rot1024ZTop / rot1024VpSrc are Scene.rot1024_zTop (=5, Scene.java:26) and the
// projection viewport-source shift rot1024_vp_src (=9 in practice; the Scene
// ctor seeds 8 at Scene.java:97 but mudclient.setMidpoints passes m_qd=9,
// mudclient.java:516,11505). project() does `(x << rot1024_vp_src) / z`.
const (
	rot1024ZTop  = 5
	rot1024VpSrc = 9
)

// =============================================================================
// Trig tables — mirror orsc/util/FastMath.java:7-31 (NOT render/camera.go).
//
// Java->Go: Java int is 32-bit; entries are (int)(32768*sin/cos), so int32.
//   trigTable256   [512]  : sin[0..255],  cos[256..511]   (step 0.02454369)
//   trigTable1024  [2048] : sin[0..1023], cos[1024..2047] (step 0.00613592315)
//   trigTable_1024 [2048] : numerically identical to trigTable1024 but a SEPARATE
//                           Java table (FastMath.java:9,29). Scene.setCamera and
//                           the frustum-corner rotate use trigTable_1024
//                           (Scene.java:576-597,2946-2963); RSModel.rotate1024
//                           uses trigTable1024 (RSModel.java:1032-1044).
//   trigTable_256  [512]  : private sin/cos twin of trigTable256 (FastMath:9,24);
//                           kept for completeness, unused by three/ rendering.
// =============================================================================

var (
	trigTable256  [512]int32
	trigTable1024 [2048]int32

	trigTable_256  [512]int32
	trigTable_1024 [2048]int32
)

func init() {
	for i := 0; i < 256; i++ {
		trigTable256[i] = int32(32768.0 * math.Sin(0.02454369*float64(i)))
		trigTable256[256+i] = int32(32768.0 * math.Cos(float64(i)*0.02454369))
		trigTable_256[i] = int32(math.Sin(0.02454369*float64(i)) * 32768.0)
		trigTable_256[256+i] = int32(math.Cos(float64(i)*0.02454369) * 32768.0)
	}
	for i := 0; i < 1024; i++ {
		trigTable1024[i] = int32(math.Sin(float64(i)*0.00613592315) * 32768.0)
		trigTable1024[i+1024] = int32(math.Cos(float64(i)*0.00613592315) * 32768.0)
		trigTable_1024[i] = int32(math.Sin(float64(i)*0.00613592315) * 32768.0)
		trigTable_1024[1024+i] = int32(math.Cos(float64(i)*0.00613592315) * 32768.0)
	}
}

// =============================================================================
// Model — mirror of RSModel (RSModel.java:12-90).
//
// Field names follow RSModel exactly. Java int[] -> []int32 for the fixed-point
// geometry/projection arrays; Java byte[] -> []int8 with &0xff masking at use
// sites where the Java code reads signed bytes; Java int[][] -> [][]int.
//
// The vertX/vertY/vertZ arrays are MODEL-LOCAL coords. resetTransformCache bakes
// orientation+scale+translate into vertXTransform/Y/Z (the working copy;
// RSModel.java:954-1011). rotate1024 then rotates those into the camera-space
// vertXRot/vertYRot/vertZRot and projects into vertexParam6 (screen X) and
// vertexParam2 (screen Y) (RSModel.java:1013-1095).
// =============================================================================
type Model struct {
	// ---- counts / heads (RSModel.java:36,49,86,16,15,62 etc.) ----
	vertHead     int // RSModel.vertHead: live vertex count
	faceHead     int // RSModel.faceHead: live face count
	faceCount    int // RSModel.faceCount: face capacity
	vertexCount2 int // RSModel.vertexCount2: vertex capacity

	// ---- model-local geometry (RSModel.java:37,88,40) ----
	vertX []int32 // RSModel.vertX
	vertY []int32 // RSModel.vertY
	vertZ []int32 // RSModel.vertZ

	// ---- baked transform working copy (RSModel.java:87,89,90) ----
	vertXTransform []int32 // RSModel.vertXTransform
	vertYTransform []int32 // RSModel.vertYTransform
	vertZTransform []int32 // RSModel.vertZTransform

	// ---- camera-space + projection scratch (RSModel.java:38,39,41,32,35,33) --
	vertXRot         []int32 // RSModel.vertXRot
	vertYRot         []int32 // RSModel.vertYRot
	vertZRot         []int32 // RSModel.vertZRot
	vertexParam6     []int32 // RSModel.vertexParam6: projected screen X
	vertexParam2     []int32 // RSModel.vertexParam2: projected screen Y
	vertDiffuseLight []int32 // RSModel.vertDiffuseLight: per-vertex gouraud light
	vertLightOther   []int8  // RSModel.vertLightOther (signed byte; mask &0xff)

	// ---- faces (RSModel.java:17,16,25,24,14,20,31) ----
	faceIndices      [][]int // RSModel.faceIndices
	faceIndexCount   []int   // RSModel.faceIndexCount
	faceTextureFront []int32 // RSModel.faceTextureFront (texture id, flat colour, or TRANSPARENT)
	faceTextureBack  []int32 // RSModel.faceTextureBack
	faceDiffuseLight []int32 // RSModel.faceDiffuseLight (flat shade, or TRANSPARENT => per-vertex)
	facePickIndex    []int32 // RSModel.facePickIndex
	mZb              []int8  // RSModel.m_zb: per-face "skip pick" flag

	// ---- per-face normals + bounds (RSModel.java:56-58,50-55) ----
	faceNormX []int32 // RSModel.faceNormX
	faceNormY []int32 // RSModel.faceNormY
	faceNormZ []int32 // RSModel.faceNormZ
	faceMinX  []int32 // RSModel.faceMinX
	faceMaxX  []int32 // RSModel.faceMaxX
	faceMinY  []int32 // RSModel.faceMinY
	faceMaxY  []int32 // RSModel.faceMaxY
	faceMinZ  []int32 // RSModel.faceMinZ
	faceMaxZ  []int32 // RSModel.faceMaxZ

	// ---- scene polygon normal cache (RSModel.java:22,23) ----
	scenePolyNormalShift     []int32 // RSModel.scenePolyNormalShift (-1 => recompute)
	scenePolyNormalMagnitude []int32 // RSModel.scenePolyNormalMagnitude

	// ---- model-level bounds (RSModel.java:65-70,64) ----
	minX, maxX       int32 // RSModel.minX / maxX
	minY, maxY       int32 // RSModel.minY / maxY
	minZ, maxZ       int32 // RSModel.minZ / maxZ
	maxFaceDimension int32 // RSModel.maxFaceDimension

	// ---- orientation / scale / translate (RSModel.java:71-85) ----
	rot256X, rot256Y, rot256Z          int32 // RSModel.rot256X/Y/Z
	scaleX, scaleY, scaleZ             int32 // RSModel.scaleX/Y/Z
	translateX, translateY, translateZ int32 // RSModel.translateX/Y/Z
	// rotation-matrix terms (RSModel.java:74-79); applyRotMatrix
	rotMxToY, rotMxToZ int32 // RSModel.rotM_xToY / rotM_xToZ
	rotMyToX, rotMyToZ int32 // RSModel.rotM_yToX / rotM_yToZ
	rotMzToX, rotMzToY int32 // RSModel.rotM_zToX / rotM_zToY
	appliedTransform   int32 // RSModel.appliedTransform: which transform stages apply (0..4)

	// ---- diffuse-light params (RSModel.java:43-48,21) ----
	diffuseDirX, diffuseDirY, diffuseDirZ int32 // RSModel.diffuseDirX/Y/Z
	diffuseMag                            int32 // RSModel.diffuseMag
	diffuseParam1                         int32 // RSModel.diffuseParam1 (=32)
	diffuseParam2                         int32 // RSModel.diffuseParam2 (=512)
	dontComputeDiffuse                    bool  // RSModel.dontComputeDiffuse

	// ---- flags (RSModel.java:19,26-30,60-63) ----
	mYb int32 // RSModel.m_Yb: transform-dirty tier (0 clean, 1 full rebuild, 2 verbatim copy)
	mHc int32 // RSModel.m_hc: depth/sort bias added to face m_t
	mDc bool  // RSModel.m_dc: model passed the frustum test in rotate1024
	mDb bool  // RSModel.m_db: no pick-buffer
	mCb bool  // RSModel.m_cb: alternate (subtractive) sprite-blend path
	mKb bool  // RSModel.m_Kb: wall-style blend path
	mB  bool  // RSModel.m_b: skip projection scratch alloc
	mC  bool  // RSModel.m_c: skip per-face bounds alloc
	mV  bool  // RSModel.m_v: alias transform arrays onto vertX/Y/Z
}

// =============================================================================
// Polygon — mirror of Polygon (Polygon.java:3-24). One per visible face in a
// render pass; carries the camera-space plane + screen/cam-Z AABB.
// =============================================================================
type Polygon struct {
	model        *Model // Polygon.model
	faceID       int    // Polygon.faceID
	normalX      int32  // Polygon.normalX
	normalY      int32  // Polygon.normalY
	normalZ      int32  // Polygon.normalZ
	orientation  int32  // Polygon.orientation: normal . vertex0 (sign => front/back)
	minP6, maxP6 int32  // Polygon.minP6 / maxP6: projected-X span
	minP2, maxP2 int32  // Polygon.minP2 / maxP2: projected-Y span
	minZ, maxZ   int32  // Polygon.minZ / maxZ: camera-Z span
	mB           int32  // Polygon.m_b: resolved fill (texture id / colour) for this pass
	mT           int32  // Polygon.m_t: average camera-Z sort key
	mF           int32  // Polygon.m_f: original index (overlap-resolution bookkeeping)
	mP           int32  // Polygon.m_p: "already pushed past" marker (= -1 initially)
	mC           bool   // Polygon.m_c: visited flag in the reinsertion walk
}

// =============================================================================
// Scanline — mirror of Scanline (Scanline.java:3-8). One per raster row; holds
// the left/right span endpoints and their interpolant in 8.8 fixed point.
// =============================================================================
type Scanline struct {
	mD int32 // Scanline.m_d: left edge X (<<8)
	mK int32 // Scanline.m_k: right edge X (<<8)
	mE int32 // Scanline.m_e: left interpolant
	mL int32 // Scanline.m_l: right interpolant
}

// =============================================================================
// Surface — the raster target. Pixels IS GraphicsController.pixelData (the int[]
// the Java Shader writes), 0x00RRGGBB, row-major (GraphicsController.java:43,
// 145-147). Depth is a per-pixel coarse camera-Z buffer (a Go addition for 2D
// sprite occlusion; the Java renderer painter-sorts sprites into the face list
// instead). toImage replaces the AWT ImageProducer handoff.
// =============================================================================
type Surface struct {
	Pixels []int32 // == GraphicsController.pixelData
	Depth  []int32 // per-pixel coarse camera-Z (depthFar = untouched)
	Width  int     // GraphicsController.width2
	Height int     // GraphicsController.height2
}

// depthFar is the far sentinel Depth resets to (nothing is in front of it).
const depthFar = int32(1 << 30)

// =============================================================================
// Camera — the orbit state Scene.setCamera computes (Scene.java:2930-2976).
// rot1024OffX/Y/Z is the eye offset baked into each model's rotate1024; the
// cameraProj* are the inverse 1024-step angles fed to rotate1024 as
// (cameraProjY, cameraProjZ, cameraProjX) => (rotY, rotZ, rotX).
// =============================================================================
type Camera struct {
	rot1024OffX int32 // Scene.rot1024_off_x
	rot1024OffY int32 // Scene.rot1024_off_y
	rot1024OffZ int32 // Scene.rot1024_off_z
	cameraProjX int32 // Scene.cameraProjX = (1024 - xRot) & 1023
	cameraProjY int32 // Scene.cameraProjY = (1024 - yRot) & 1023
	cameraProjZ int32 // Scene.cameraProjZ = (1024 - zRot) & 1023
}

// =============================================================================
// Scene — mirror of Scene (Scene.java:10-79). Holds the model list, projection
// midpoints/centres, camera, the per-row Scanline array, the Polygon pool, and
// the texture resourceDatabase. Field names follow Scene's (cryptic) Java names
// with a doc gloss; the obfuscated single-letter names are renamed to readable
// equivalents and cited.
// =============================================================================
type Scene struct {
	graphics  *Surface // Scene.graphics (the GraphicsController) -> Surface
	pixelData []int32  // Scene.pixelData == graphics.Pixels

	models     []*Model // Scene.models (capacity m_u)
	modelCount int      // Scene.modelCount
	modelMax   int      // Scene.m_u

	mT *Model // Scene.m_T: the synthetic sprite/entity-billboard model

	cam Camera // (rot1024_off_* + cameraProj*) from setCamera

	// ---- projection midpoints / centres (Scene.java:33,64,82,70,63,67) ----
	mA           int32 // Scene.m_A: half view width (centre X / horizontal clip extent)
	mWb          int32 // Scene.m_wb: half view height (centre Y)
	mNb          int32 // Scene.m_Nb: raster Y origin (scanline array base)
	mVb          int32 // Scene.m_vb: pixel row stride
	mZb          int32 // Scene.m_Zb: raster X origin
	rot1024VpSrc int32 // Scene.rot1024_vp_src: projection shift (=9)

	// ---- scanline / polygon working set (Scene.java:66,68,69,67,33,35,40,65) -
	scanlines  []*Scanline // Scene.m_x: per-row spans (len = m_wb+m_Nb)
	polygons   []*Polygon  // Scene.polygons (capacity maxPolygonCount)
	polyHead   int         // Scene.m_zb: live polygon count this pass
	rasterMinY int32       // Scene.m_Xb: current face min raster row
	rasterMaxY int32       // Scene.m_Cb: current face max raster row
	mE         int32       // Scene.m_e: reinsertion split pivot result
	mEb        int32       // Scene.m_eb: reinsertion split pivot result

	interlace bool // Scene.m_f: interlaced (skip odd rows)

	// ---- texture resource database (Scene.java:53,45,46,61,48,42,49,37,32) ---
	// resourceDatabase[id] is the 64KB/16KB shade-mipped texel buffer built lazily
	// by Scene.b()/setFrustum(id,..) from textureSource (the sprite pixels) +
	// texturePalette. mHb[id]!=0 marks a 128px (class1) texture (65536 entries)
	// vs 64px (16384 entries). mS[id] marks a texture that contains TRANSPARENT.
	resourceDatabase [][]int32 // Scene.resourceDatabase
	texturePalette   [][]int32 // Scene.m_L: per-texture colour LUT
	textureSource    [][]int8  // Scene.m_g: per-texture index bytes (signed; &0xff)
	mHb              []int32   // Scene.m_Hb: per-texture size class (0=64px,1=128px)
	mS               []bool    // Scene.m_S: per-texture has-transparency flag
	mD               []int64   // Scene.m_D: per-texture LRU timestamp
	textureCount     int       // Scene.m_cb: number of textures
	pool64           [][]int32 // Scene.m_i: free 16384-int buffers (64px)
	pool128          [][]int32 // Scene.m_ec: free 65536-int buffers (128px)

	// ---- distance fog (Scene.java:27,28,29,30) ----
	fogSmoothingStartDistance int32 // Scene.fogSmoothingStartDistance (=10)
	fogZFalloff               int32 // Scene.fogZFalloff (=20)
	fogLandscapeDistance      int32 // Scene.fogLandscapeDistance (=1000; set per-frame)
	fogEntityDistance         int32 // Scene.fogEntityDistance (=1000)

	// ---- sprite/billboard scratch (Scene.java:46,41,44,57,58,60,33) ----
	spriteID    []int32 // Scene.m_gb
	spriteX     []int32 // Scene.m_Ob
	spriteW     []int32 // Scene.m_ob
	spriteH     []int32 // Scene.m_Eb
	spriteTexID []int32 // Scene.m_Fb
	spriteY     []int32 // Scene.m_a
	spriteXOff  []int32 // Scene.m_Q
	spriteHead  int     // Scene.m_n

	// entitySprites is the per-billboard composited pixel data (the Go analogue
	// of GraphicsController.sprites[m_gb[f]]); drawEntity blits it. spriteID[f]
	// indexes into this slice. Not in the Java Scene (sprites live on the
	// GraphicsController there); kept on the Scene here so the port stays
	// self-contained — the caller hands raw pixels via AddEntity.
	entitySprites []*entitySprite

	// debugBillboards holds Phase-0 placement-sanity billboards (the NPC/player
	// 2D-sprite parity build, docs/build/NPC_SPRITE_PARITY_PLAN.md). Each is
	// registered as an mT billboard face so Render PROJECTS it with the same
	// camera as terrain/scenery; after Render, fillDebugBillboards reads back the
	// projected vertex + the private projection fields (mZb/mNb/rot1024VpSrc),
	// recomputes the screen rect (the SAME w/h/x/y formula the Java legs' rect
	// replicator uses), and fills it SOLID on top of the 3D framebuffer — pure
	// projection, never depth-occluded — so the rect diff isolates the projection.
	debugBillboards []debugBillboard

	// ---- gradient cache for untextured walls (Scene.java:17,23) ----
	gradientCache [][]int32 // Scene.m_Ib: lazily-filled fog gradient ramps
	gradientKeys  []int32   // Scene.m_v: colour key per gradientCache slot
}

// =============================================================================
// Constructors + key entry-point STUBS. Bodies are filled by the parallel port
// files; these establish the package shape + contract. Empty bodies / panics.
// =============================================================================

// NewModel, FromAssets, and the (*Model) methods Translate, Orient,
// computeBounds, project and light are implemented in model.go (the faithful
// port of RSModel.java's ctor / setRot256 / setTranslate /
// calculateBoundingBoxes / rotate1024 / computeNormals+computeDiffuse).

// NewScene, AddModel, SetCamera and Render are implemented in scene.go (the
// faithful port of Scene.java's ctor / addModel / setCamera / endScene).

// NewSurface allocates a w*h framebuffer (GraphicsController image alloc,
// GraphicsController.java:68-70/145-147) with Depth reset to depthFar.
func NewSurface(w, h int) *Surface {
	s := &Surface{
		Pixels: make([]int32, w*h),
		Depth:  make([]int32, w*h),
		Width:  w,
		Height: h,
	}
	for i := range s.Depth {
		s.Depth[i] = depthFar
	}
	return s
}

// Surface.toImage / ToImage / PNG and the Surface.Clear/SetPixel/SetDepth
// helpers are implemented in surface.go (the GraphicsController -> Image handoff
// replacing the AWT ImageProducer).
