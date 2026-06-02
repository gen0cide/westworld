package orsc

import "math"

// =============================================================================
// scene.go — faithful port of Scene.java's SCENE / PROJECTION half.
//
// Source of truth:
//   /Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/Scene.java
//
// This file ports:
//   - the Scene constructor                       (Scene.java:81-141)
//   - addModel                                    (Scene.java:2314-2328)
//   - setCamera                                   (Scene.java:2930-2976)
//   - endScene (here exported as Render)          (Scene.java:2555-2856)
//   - computePolygon                              (Scene.java:478-564)
//   - the painter qsort                           (Scene.java:1345-1381)
//   - the 3D plane overlap-resolution family:
//        polygonHit1 / polygonHit2               (Scene.java:143-203 / 245-381)
//        booleanCombinatoric / 2 / 3             (Scene.java:214-243 / 415-432)
//        reinsertion walk (setFrustum int,Polygon[])   (Scene.java:663-715)
//        reinsertion split (setFrustum int,Polygon[],int) (Scene.java:1383-1441)
//        overlap test (setFrustum int[]...)       (Scene.java:1443-1841)
//   - the scanline span builder
//        (setFrustum int,int,int[],...,int)       (Scene.java:717-1343)
//
// The per-pixel fill (Shader.shadeScanline + the flat-colour copyBlock paths)
// lives in raster.go (another port file); scene.go drives it through the
// rasterizer hooks documented at the bottom of this file (fillTexturedSpan /
// fillFlatSpan / fillGouraudColumn / fillGouraudColumnAlt).
//
// Java->Go: Java int is 32-bit. The whole fixed-point pipeline (>>, *, the
// perspective divide, the scanline interpolants) is therefore int32. Java's
// >>> (logical shift) is handled in raster.go via uint32. faceIndices etc. are
// []int (Java int[][]) so they index directly.
// =============================================================================

// The MiscFunctions frustum bounds (frustumNearZ/FarZ/MinX/MaxX/MinY/MaxY,
// MiscFunctions.java:12-22) are package-level vars declared in model.go (they
// are written here by frustumCorner and read by model.project's frustum test).
// =============================================================================
// Scene constructor — Scene(GraphicsController, maxModels, maxPolygons,
// maxSpriteFaces) (Scene.java:81-141).
//
// Scene.java seeds m_A=256, m_wb=192, m_Nb=256, m_vb=512, m_Zb=256 then
// overrides m_A=width/2, m_wb=height/2 from the GraphicsController. The pixel
// row stride m_vb is the surface Width (Scene writes pixelData[var33 = m_vb*row
// + m_Zb + x]); we mirror mudclient.setMidpoints' final layout: m_Zb=Width/2,
// m_Nb=Height/2, m_vb=Width, rot1024_vp_src=9 (the m_qd override,
// mudclient.java:516,11505 — see types.go rot1024VpSrc).
// =============================================================================
func NewScene(surf *Surface, maxModels, maxPolygons, maxSpriteFaces int) *Scene {
	s := &Scene{}

	// Scene.java:84,95 — fog seeds.
	s.fogSmoothingStartDistance = 10 // Scene.java:27
	s.fogZFalloff = 20               // Scene.java:28
	s.fogLandscapeDistance = 1000    // Scene.java:84
	s.fogEntityDistance = 1000       // Scene.java:95

	// Scene.java:103-107 — bind the framebuffer.
	s.graphics = surf
	s.pixelData = surf.Pixels
	s.modelCount = 0
	s.modelMax = maxModels // m_u

	// Scene.java:105,107 — projection centres from the surface size.
	s.mA = int32(surf.Width / 2)   // m_A: half view width / horizontal clip extent
	s.mWb = int32(surf.Height / 2) // m_wb: half view height

	// mudclient.setMidpoints layout (mudclient.java:11505): m_Nb=height/2,
	// m_Zb=width/2, m_vb=width (row stride), rot1024_vp_src=m_qd=9.
	s.mNb = int32(surf.Height / 2) // raster Y origin (scanline array base)
	s.mZb = int32(surf.Width / 2)  // raster X origin
	s.mVb = int32(surf.Width)      // pixel row stride
	s.rot1024VpSrc = rot1024VpSrc  // =9 (types.go)

	// Scene.java:108-116 — model + polygon pools.
	s.models = make([]*Model, s.modelMax+1) // +1: endScene appends m_T at [modelCount]
	s.polyHead = 0
	s.polygons = make([]*Polygon, maxPolygons)
	for i := 0; i < maxPolygons; i++ {
		s.polygons[i] = &Polygon{}
	}

	// Scene.java:117-141 — the synthetic sprite/billboard model and its scratch.
	s.mT = NewModel(maxSpriteFaces*2, maxSpriteFaces)
	s.spriteW = make([]int32, maxSpriteFaces)     // m_ob
	s.spriteH = make([]int32, maxSpriteFaces)     // m_Eb
	s.spriteTexID = make([]int32, maxSpriteFaces) // m_Fb
	s.spriteX = make([]int32, maxSpriteFaces)     // m_Ob
	s.spriteID = make([]int32, maxSpriteFaces)    // m_gb
	s.spriteY = make([]int32, maxSpriteFaces)     // m_a
	s.spriteXOff = make([]int32, maxSpriteFaces)  // m_Q
	s.spriteHead = 0

	// The per-row Scanline array (mudclient.setMidpoints, Scene.java:3038-3045):
	// length m_wb + m_Nb, all rows allocated up front.
	rows := int(s.mWb + s.mNb)
	s.scanlines = make([]*Scanline, rows)
	for i := 0; i < rows; i++ {
		s.scanlines[i] = &Scanline{}
	}

	// gradient cache for untextured walls (Scene.java:82,92): m_ib=50 ramps of
	// 256 entries, keyed by colour in m_v.
	s.gradientCache = make([][]int32, 50)
	for i := range s.gradientCache {
		s.gradientCache[i] = make([]int32, 256)
	}
	s.gradientKeys = make([]int32, 50)

	// Texture resource database + LRU pools (Scene.java ctor texture-array alloc).
	// Sized to hold the full authentic texture set resident (textureSlots > the
	// max texture id a terrain face fill can carry), so no LRU eviction is needed
	// for a static render. LoadTexturesFromArchive populates these from
	// Authentic_Sprites.orsc; an unpopulated id (textureSource[id]==nil) is treated
	// as flat by the fill path so a missing texture degrades instead of crashing.
	s.textureCount = textureSlots
	s.resourceDatabase = make([][]int32, textureSlots)
	s.texturePalette = make([][]int32, textureSlots)
	s.textureSource = make([][]int8, textureSlots)
	s.mHb = make([]int32, textureSlots)
	s.mS = make([]bool, textureSlots)
	s.mD = make([]int64, textureSlots)
	s.pool128 = make([][]int32, textureSlots)
	s.pool64 = make([][]int32, textureSlots)

	return s
}

// =============================================================================
// addModel — Scene.addModel (Scene.java:2314-2328).
// =============================================================================
func (s *Scene) AddModel(m *Model) {
	if m == nil {
		return // Scene.java:2317 just warns
	}
	if s.modelCount < s.modelMax {
		s.models[s.modelCount] = m
		s.modelCount++
	}
}

// =============================================================================
// setCamera — Scene.setCamera(centerX, centerY, centerZ, xRot, yRot, zRot,
// offset) (Scene.java:2930-2976).
//
// Computes the inverse projection angles cameraProjX/Y/Z = (1024 - rot) & 1023
// and the orbit eye offset rot1024_off_x/y/z = center - rotated(offset). Uses
// trigTable_1024 (FastMath.java, the SEPARATE copy used by setCamera — types.go).
//
// AXIS / ARG ORDER (contract note 1): do NOT swap yaw/roll like render/camera.go.
// endScene binds rotate1024's (rotY,rotZ,rotX) <- (cameraProjY,cameraProjZ,
// cameraProjX) exactly (Scene.java:2586-2591); setCamera computes them straight.
// =============================================================================
func (s *Scene) SetCamera(centerX, centerY, centerZ, xRot, yRot, zRot, offset int32) {
	zRot &= 1023 // Scene.java:2932-2934
	xRot &= 1023
	yRot &= 1023

	s.cam.cameraProjZ = (1024 - zRot) & 1023 // Scene.java:2936-2938
	s.cam.cameraProjX = (1024 - xRot) & 1023
	s.cam.cameraProjY = (1024 - yRot) & 1023

	var offX int32 // Scene.java:2939-2941
	var offY int32
	offZ := offset

	// NOTE: Java `a*b - c*d >> 15` means `(a*b - c*d) >> 15` (shift binds LOOSER
	// than +/-). In Go `>>` binds TIGHTER than +/-, so the parens are REQUIRED to
	// match Java — without them only the last term is shifted, corrupting the
	// camera offset. (Same Java/Go trap recurs anywhere `±…>>n`/`±…<<n` appears.)
	var sin, cos, tmp int32
	if xRot != 0 { // Scene.java:2945-2951
		sin = trigTable_1024[xRot]
		cos = trigTable_1024[xRot+1024]
		tmp = (cos*offY - sin*offset) >> 15
		offZ = (sin*offY + offset*cos) >> 15
		offY = tmp
	}
	if yRot != 0 { // Scene.java:2953-2959
		sin = trigTable_1024[yRot]
		cos = trigTable_1024[yRot+1024]
		tmp = (offX*cos + offZ*sin) >> 15
		offZ = (cos*offZ - sin*offX) >> 15
		offX = tmp
	}
	if zRot != 0 { // Scene.java:2961-2967
		cos = trigTable_1024[zRot+1024]
		sin = trigTable_1024[zRot]
		tmp = (offX*cos + sin*offY) >> 15
		offY = (offY*cos - sin*offX) >> 15
		offX = tmp
	}

	s.cam.rot1024OffZ = centerZ - offZ // Scene.java:2969-2971
	s.cam.rot1024OffY = centerY - offY
	s.cam.rot1024OffX = centerX - offX
}

// =============================================================================
// frustumCorner — Scene.setFrustum(int x,int y,int z,boolean) (Scene.java:566).
// Rotates a frustum corner by the inverse camera angles (trigTable_1024) and
// folds it into the package frustum min/max accumulators. Mirrors setCamera's
// (1024-cameraProj*)&1023 unwinding so the corners land in model space.
// =============================================================================
func (s *Scene) frustumCorner(x, y, z int32) {
	projX := (1024 - s.cam.cameraProjX) & 1023 // Scene.java:569-571
	projY := (1024 - s.cam.cameraProjY) & 1023
	projZ := (1024 - s.cam.cameraProjZ) & 1023

	var t0, t1, t2 int32
	if projZ != 0 { // Scene.java:575-581
		t0 = trigTable_1024[projZ]
		t1 = trigTable_1024[1024+projZ]
		t2 = (t1*y + t0*z) >> 15
		z = (z*t1 - t0*y) >> 15
		y = t2
	}
	if projX != 0 { // Scene.java:583-589
		t1 = trigTable_1024[1024+projX]
		t0 = trigTable_1024[projX]
		t2 = (z*t1 - t0*x) >> 15
		x = (t1*x + t0*z) >> 15
		z = t2
	}
	if projY != 0 { // Scene.java:591-597
		t0 = trigTable_1024[projY]
		t1 = trigTable_1024[1024+projY]
		t2 = (t0*x + y*t1) >> 15
		x = (t1*x - t0*y) >> 15
		y = t2
	}

	if x > frustumMinX { // Scene.java:599-621
		frustumMinX = x
	}
	if z < frustumFarZ {
		frustumFarZ = z
	}
	if y > frustumMinY {
		frustumMinY = y
	}
	if z > frustumNearZ {
		frustumNearZ = z
	}
	if x < frustumMaxX {
		frustumMaxX = x
	}
	if y < frustumMaxY {
		frustumMaxY = y
	}
}

// =============================================================================
// computePolygon — Scene.computePolygon(int polyID) (Scene.java:478-564).
// Derives the camera-space face plane (normalX/Y/Z, orientation = N.v0,
// scenePolyNormalMagnitude) and the projected/cam-Z AABB for one polygon.
// =============================================================================
func (s *Scene) computePolygon(polyID int) {
	poly := s.polygons[polyID]
	model := poly.model
	face := poly.faceID
	index := model.faceIndices[face]
	indexCount := model.faceIndexCount[face]
	fParam4 := model.scenePolyNormalShift[face]

	v0X := model.vertXRot[index[0]] // Scene.java:487-495
	v0Y := model.vertYRot[index[0]]
	v0Z := model.vertZRot[index[0]]
	v1DX := model.vertXRot[index[1]] - v0X
	v1DY := model.vertYRot[index[1]] - v0Y
	v1DZ := model.vertZRot[index[1]] - v0Z
	v2DX := model.vertXRot[index[2]] - v0X
	v2DY := model.vertYRot[index[2]] - v0Y
	v2DZ := model.vertZRot[index[2]] - v0Z

	normX := v2DZ*v1DY - v1DZ*v2DY // Scene.java:496-498
	normY := v2DX*v1DZ - v1DX*v2DZ
	normZ := v1DX*v2DY - v2DX*v1DY

	if fParam4 != -1 { // Scene.java:499-516
		normZ >>= uint(fParam4)
		normX >>= uint(fParam4)
		normY >>= uint(fParam4)
	} else {
		fParam4 = 0
		for normX > 25000 || normY > 25000 || normZ > 25000 ||
			normX < -25000 || normY < -25000 || normZ < -25000 {
			normX >>= 1
			normY >>= 1
			normZ >>= 1
			fParam4++
		}
		model.scenePolyNormalShift[face] = fParam4
		model.scenePolyNormalMagnitude[face] = int32(float64(polyNormalScale) *
			math.Sqrt(float64(normZ*normZ+normY*normY+normX*normX)))
	}

	poly.normalX = normX // Scene.java:518-521
	poly.normalY = normY
	poly.normalZ = normZ
	poly.orientation = (normX * v0X) + (normY * v0Y) + (normZ * v0Z)

	minZ := model.vertZRot[index[0]] // Scene.java:523-528
	maxZ := minZ
	minP6 := model.vertexParam6[index[0]]
	maxP6 := minP6
	minP2 := model.vertexParam2[index[0]]
	maxP2 := minP2

	for v := 1; v < indexCount; v++ { // Scene.java:530-553
		vvt := model.vertZRot[index[v]]
		if vvt <= maxZ {
			if vvt < minZ {
				minZ = vvt
			}
		} else {
			maxZ = vvt
		}
		vvt = model.vertexParam6[index[v]]
		if vvt > maxP6 {
			maxP6 = vvt
		} else if vvt < minP6 {
			minP6 = vvt
		}
		vvt = model.vertexParam2[index[v]]
		if vvt > maxP2 {
			maxP2 = vvt
		} else if minP2 > vvt {
			minP2 = vvt
		}
	}

	poly.minP6 = minP6 // Scene.java:555-560
	poly.maxP6 = maxP6
	poly.maxP2 = maxP2
	poly.maxZ = maxZ
	poly.minP2 = minP2
	poly.minZ = minZ
}

// =============================================================================
// computeSpritePolygon — Scene.b(int var1,int var2) (Scene.java:2426-2491).
// The m_T billboard-face equivalent of computePolygon: a degenerate 2-vertex
// "face" gets a Z-facing normal and a +/-20 padded screen AABB.
// =============================================================================
func (s *Scene) computeSpritePolygon(polyID int) {
	poly := s.polygons[polyID]
	model := poly.model
	face := poly.faceID
	index := model.faceIndices[face]

	var nX, nY, nZ int32 = 0, 0, 1 // Scene.java:2433-2435 (var9/var10/var11)
	v0X := model.vertXRot[index[0]]
	v0Y := model.vertYRot[index[0]]
	v0Z := model.vertZRot[index[0]]
	model.scenePolyNormalMagnitude[face] = 1 // Scene.java:2439-2440
	model.scenePolyNormalShift[face] = 0
	poly.normalY = nY // Scene.java:2441-2444
	poly.normalX = nX
	poly.normalZ = nZ
	poly.orientation = v0Z*nZ + v0X*nX + v0Y*nY

	minZ := model.vertZRot[index[0]] // Scene.java:2445-2456
	maxZ := minZ
	minP6 := model.vertexParam6[index[0]]
	maxP6 := minP6
	if model.vertexParam6[index[1]] >= minP6 {
		maxP6 = model.vertexParam6[index[1]]
	} else {
		minP6 = model.vertexParam6[index[1]]
	}
	maxP2 := model.vertexParam2[index[1]]
	minP2 := model.vertexParam2[index[0]]

	z1 := model.vertZRot[index[1]] // Scene.java:2457-2464
	if z1 <= minZ {
		if minZ > z1 {
			minZ = z1
		}
	} else {
		maxZ = z1
	}

	p6 := model.vertexParam6[index[1]] // Scene.java:2466-2473
	if maxP6 >= p6 {
		if minP6 > p6 {
			minP6 = p6
		}
	} else {
		maxP6 = p6
	}

	p2 := model.vertexParam2[index[1]] // Scene.java:2475-2487
	poly.maxP6 = maxP6 + 20
	poly.minP6 = minP6 - 20
	if minP2 < p2 {
		minP2 = p2
	} else if p2 < maxP2 {
		maxP2 = p2
	}
	poly.maxZ = maxZ
	poly.minZ = minZ
	poly.maxP2 = minP2
	poly.minP2 = maxP2
}

// =============================================================================
// Render — Scene.endScene(int) (Scene.java:2555-2856).
//
// Drives RSModel.rotate1024 (model.project) on every model, collects visible,
// in-frustum, in-window faces into Polygons (computePolygon), coarse-sorts them
// by m_t (average camera-Z) via the qsort, resolves mutual 3D overlap with the
// polygonHit / booleanCombinatoric family so wall corners paint correctly, then
// scans + fills each face.
//
// NewScene already bound the framebuffer and SetCamera the camera, so Render
// takes no args (mirroring endScene's effectively-arg-less body).
// =============================================================================
func (s *Scene) Render() {
	// Scene.java:2558 `this.m_f = this.graphics.interlace;` — GraphicsController's
	// interlace flag defaults false (full-resolution); Surface has no such field
	// in this port, so progressive scan is used. raster.go reads s.interlace.
	s.interlace = false

	// Scene.java:2559-2580 — build the view frustum bounds in model space. The
	// far corners are at z=fogLandscapeDistance scaled to the half-view extents;
	// the near corners are the screen edges at z=0.
	var7 := (s.mA * s.fogLandscapeDistance) >> uint(s.rot1024VpSrc) // Scene.java:2559
	frustumFarZ = 0                                                 // Scene.java:2560-2566
	frustumNearZ = 0
	frustumMaxX = 0
	frustumMinX = 0
	var8 := (s.fogLandscapeDistance * s.mWb) >> uint(s.rot1024VpSrc) // Scene.java:2564
	frustumMinY = 0
	frustumMaxY = 0
	s.frustumCorner(s.fogLandscapeDistance, -var7, -var8) // Scene.java:2567-2574
	s.frustumCorner(s.fogLandscapeDistance, -var7, var8)
	s.frustumCorner(s.fogLandscapeDistance, var7, -var8)
	s.frustumCorner(s.fogLandscapeDistance, var7, var8)
	s.frustumCorner(0, -s.mA, -s.mWb)
	s.frustumCorner(0, -s.mA, s.mWb)
	s.frustumCorner(0, s.mA, -s.mWb)
	s.frustumCorner(0, s.mA, s.mWb)
	frustumNearZ += s.cam.rot1024OffY // Scene.java:2575-2580
	frustumMinX += s.cam.rot1024OffZ
	frustumFarZ += s.cam.rot1024OffY
	frustumMaxY += s.cam.rot1024OffX
	frustumMaxX += s.cam.rot1024OffZ
	frustumMinY += s.cam.rot1024OffX

	// Scene.java:2581-2591 — append the m_T sprite model and project everything.
	// endScene binds rotate1024(rotY,rotZ,rotX) <- (cameraProjY,cameraProjZ,
	// cameraProjX); model.project reads cam.cameraProj* + cam.rot1024Off* and
	// uses s.rot1024VpSrc / rot1024ZTop, so the binding lives inside project().
	s.models[s.modelCount] = s.mT
	s.mT.mYb = 2 // verbatim-copy transform tier (resetTransformCache m_Yb==2)

	for i := 0; i < s.modelCount; i++ { // Scene.java:2585-2588
		s.models[i].project(s.cam, s.rot1024VpSrc, rot1024ZTop)
	}
	s.models[s.modelCount].project(s.cam, s.rot1024VpSrc, rot1024ZTop) // Scene.java:2590-2591

	s.polyHead = 0 // Scene.java:2592 (m_zb)

	// Scene.java:2602-2682 — collect visible faces from the geometry models.
	for mi := 0; mi < s.modelCount; mi++ {
		model := s.models[mi]
		if !model.mDc {
			continue
		}
		for f := 0; f < model.faceHead; f++ {
			indexCount := model.faceIndexCount[f]
			index := model.faceIndices[f]

			// near/far cam-Z band test (Scene.java:2611-2617)
			inBand := false
			for v := 0; v < indexCount; v++ {
				z := model.vertZRot[index[v]]
				if rot1024ZTop < z && s.fogLandscapeDistance > z {
					inBand = true
					break
				}
			}
			if !inBand {
				continue
			}

			// horizontal window test (Scene.java:2620-2638)
			win := int32(0)
			for v := 0; v < indexCount; v++ {
				vx := model.vertexParam6[index[v]]
				if -s.mA < vx {
					win |= 1
				}
				if s.mA > vx {
					win |= 2
				}
				if win == 3 {
					break
				}
			}
			if win != 3 {
				continue
			}

			// vertical window test (Scene.java:2638-2654)
			win = 0
			for v := 0; v < indexCount; v++ {
				vy := model.vertexParam2[index[v]]
				if -s.mWb < vy {
					win |= 1
				}
				if s.mWb > vy {
					win |= 2
				}
				if win == 3 {
					break
				}
			}
			if win != 3 {
				continue
			}

			// build the polygon (Scene.java:2655-2677)
			poly := s.polygons[s.polyHead]
			poly.model = model
			poly.faceID = f
			s.computePolygon(s.polyHead)

			var fill int32 // pick front/back fill by orientation sign
			if poly.orientation < 0 {
				fill = model.faceTextureFront[f]
			} else {
				fill = model.faceTextureBack[f]
			}
			if fill != TRANSPARENT {
				var zsum int32
				for v := 0; v < indexCount; v++ {
					zsum += model.vertZRot[index[v]]
				}
				poly.mT = model.mHc + zsum/int32(indexCount) // m_t = depth bias + avg camZ
				s.polyHead++
				poly.mB = fill
			}
		}
	}

	// Scene.java:2688-2712 — collect the m_T billboard faces (entity sprites).
	mt := s.mT
	if mt.mDc {
		for f := 0; f < mt.faceHead; f++ {
			index := mt.faceIndices[f]
			v0 := index[0]
			p6 := mt.vertexParam6[v0]
			p2 := mt.vertexParam2[v0]
			z := mt.vertZRot[v0]
			if rot1024ZTop < z && z < s.fogEntityDistance {
				w := (s.spriteW[f] << uint(s.rot1024VpSrc)) / z
				h := (s.spriteH[f] << uint(s.rot1024VpSrc)) / z
				if s.mA >= p6-w/2 && -s.mA <= p6+w/2 && p2-h <= s.mWb && p2 >= -s.mWb {
					poly := s.polygons[s.polyHead]
					poly.faceID = f
					poly.model = mt
					s.computeSpritePolygon(s.polyHead)
					poly.mT = (mt.vertZRot[index[1]] + z) / 2
					s.polyHead++
				}
			}
		}
	}

	if s.polyHead == 0 { // Scene.java:2714
		return
	}

	// Scene.java:2715-2716 — coarse painter sort (far first) then overlap pass.
	s.qsortPolygons(0, s.polyHead-1)
	s.resolveOverlaps(s.polyHead, 100)

	// Scene.java:2718-2849 — scan + fill each polygon back-to-front.
	for pi := 0; pi < s.polyHead; pi++ {
		poly := s.polygons[pi]
		f := poly.faceID
		model := poly.model

		if model == mt {
			// entity-sprite billboard (Scene.java:2725-2748). The 2D blit is a
			// raster.go concern; drive it through drawEntity.
			index := model.faceIndices[f]
			v0 := index[0]
			p6 := model.vertexParam6[v0]
			p2 := model.vertexParam2[v0]
			z := model.vertZRot[v0]
			w := (s.spriteW[f] << uint(s.rot1024VpSrc)) / z
			h := (s.spriteH[f] << uint(s.rot1024VpSrc)) / z
			flip := model.vertexParam6[index[1]] - p6
			x0 := p6 - w/2
			y0 := s.mNb - (h - p2)
			s.drawEntity(s.spriteID[f], x0+s.mZb, y0, w, h,
				(256<<uint(s.rot1024VpSrc))/z, flip)
			continue
		}

		// Geometry face (Scene.java:2749-2847).
		indexCount := model.faceIndexCount[f]

		// flat per-face shade base (Scene.java:2753-2759).
		var shadeBase int32
		if model.faceDiffuseLight[f] != TRANSPARENT {
			if poly.orientation < 0 {
				shadeBase = model.diffuseParam1 - model.faceDiffuseLight[f]
			} else {
				shadeBase = model.diffuseParam1 + model.faceDiffuseLight[f]
			}
		}

		index := model.faceIndices[f]
		// near-clip-split projected polygon + per-vertex shade
		// (Scene.java:2761-2824). m_yb=screenX, m_B=screenY, m_r=shade.
		var screenX [40]int32 // m_yb
		var screenY [40]int32 // m_B
		var shade [40]int32   // m_r
		// camera-space verts in ORIGINAL order for the textured planar projection
		// (m_J=camZ, m_Qb=camX, m_Vb=camY; Scene.java:2765-2767).
		var camX [40]int32 // m_Qb
		var camY [40]int32 // m_Vb
		var camZ [40]int32 // m_J
		emitted := 0

		for v := 0; v < indexCount; v++ {
			vid := index[v]
			camX[v] = model.vertXRot[vid]
			camY[v] = model.vertYRot[vid]
			camZ[v] = model.vertZRot[vid]

			if model.faceDiffuseLight[f] == TRANSPARENT { // gouraud per-vertex
				if poly.orientation < 0 {
					shadeBase = model.diffuseParam1 + int32(model.vertLightOther[vid]) -
						model.vertDiffuseLight[vid]
				} else {
					shadeBase = int32(model.vertLightOther[vid]) + model.diffuseParam1 +
						model.vertDiffuseLight[vid]
				}
			}

			if model.vertZRot[vid] >= rot1024ZTop { // in front of the near plane
				screenX[emitted] = model.vertexParam6[vid]
				screenY[emitted] = model.vertexParam2[vid]
				shade[emitted] = shadeBase
				if model.vertZRot[vid] > s.fogSmoothingStartDistance {
					shade[emitted] += (model.vertZRot[vid] - s.fogSmoothingStartDistance) / s.fogZFalloff
				}
				emitted++
			} else {
				// near-clip split against both neighbours (Scene.java:2787-2822)
				var prev int
				if v != 0 {
					prev = index[v-1]
				} else {
					prev = index[indexCount-1]
				}
				if model.vertZRot[prev] >= rot1024ZTop {
					dz := model.vertZRot[vid] - model.vertZRot[prev]
					cy := model.vertYRot[vid] - (model.vertZRot[vid]-rot1024ZTop)*
						(model.vertYRot[vid]-model.vertYRot[prev])/dz
					cx := model.vertXRot[vid] - (model.vertXRot[vid]-model.vertXRot[prev])*
						(model.vertZRot[vid]-rot1024ZTop)/dz
					screenX[emitted] = (cx << uint(s.rot1024VpSrc)) / rot1024ZTop
					screenY[emitted] = (cy << uint(s.rot1024VpSrc)) / rot1024ZTop
					shade[emitted] = shadeBase
					emitted++
				}
				var next int
				if indexCount-1 == v {
					next = index[0]
				} else {
					next = index[v+1]
				}
				if model.vertZRot[next] >= rot1024ZTop {
					dz := model.vertZRot[vid] - model.vertZRot[next]
					cy := model.vertYRot[vid] - (model.vertZRot[vid]-rot1024ZTop)*
						(model.vertYRot[vid]-model.vertYRot[next])/dz
					cx := model.vertXRot[vid] - (model.vertXRot[vid]-model.vertXRot[next])*
						(model.vertZRot[vid]-rot1024ZTop)/dz
					screenX[emitted] = (cx << uint(s.rot1024VpSrc)) / rot1024ZTop
					screenY[emitted] = (cy << uint(s.rot1024VpSrc)) / rot1024ZTop
					shade[emitted] = shadeBase
					emitted++
				}
			}
		}

		// clamp shade to [0,255] then apply the textured intensity shift
		// (Scene.java:2826-2842): <<9 for a 128px texture, <<6 for 64px. Java bounds
		// this by var17 = the ORIGINAL faceIndexCount (not the clipped/emitted count
		// var14), so near-plane-split faces leave their added clip vertices unshifted
		// — match that with indexCount, not emitted.
		for v := 0; v < int(indexCount); v++ {
			if shade[v] >= 0 {
				if shade[v] > 255 {
					shade[v] = 255
				}
			} else {
				shade[v] = 0
			}
			if poly.mB >= 0 {
				if s.mHb[poly.mB] != 1 {
					shade[v] <<= 6
				} else {
					shade[v] <<= 9
				}
			}
		}

		// build the scanline spans then fill. Scene.java:2844 calls the scanline
		// builder setFrustum(0, faceID, m_B/*screenY*/, 0, 0, model, m_yb/*screenX*/,
		// m_r/*shade*/, 0, 5960, vertCount) -> raster.scanEdges; Scene.java:2846
		// fills setFrustum(m_Vb/*camY*/, model, 1, indexCount, m_b/*fill*/,
		// m_J/*camZ*/, m_Qb/*camX*/, 0, 0) -> raster.fillFace.
		s.scanEdges(f, screenY[:], model, screenX[:], shade[:], 5960, int32(emitted))
		if s.rasterMinY < s.rasterMaxY {
			s.fillFace(camY[:], model, 1, int32(indexCount), poly.mB, camZ[:], camX[:])
		}
	}
}

// drawEntity (the entity-billboard blit, Scene.endScene -> GraphicsController.
// drawEntity, Scene.java:2738) is implemented in entity.go.

// =============================================================================
// qsortPolygons — Scene.setFrustum(int var1,int,Polygon[],int var4)
// (Scene.java:1345-1381). Median-of-three-ish quicksort on poly.mT, FAR first
// (descending m_t: var3[var5].m_t > pivot advances). Operates on s.polygons.
// =============================================================================
func (s *Scene) qsortPolygons(lo, hi int) {
	if hi <= lo {
		return
	}
	left := lo - 1
	right := hi + 1
	mid := (hi + lo) / 2
	tmp := s.polygons[mid] // Scene.java:1352-1354 swap mid<->lo
	s.polygons[mid] = s.polygons[lo]
	s.polygons[lo] = tmp
	pivot := tmp.mT

	for right > left { // Scene.java:1357-1371
		for {
			left++
			if s.polygons[left].mT <= pivot {
				break
			}
		}
		for {
			right--
			if pivot <= s.polygons[right].mT {
				break
			}
		}
		if right > left {
			t := s.polygons[left]
			s.polygons[left] = s.polygons[right]
			s.polygons[right] = t
		}
	}
	s.qsortPolygons(lo, right)   // Scene.java:1373
	s.qsortPolygons(right+1, hi) // Scene.java:1374
}

// =============================================================================
// resolveOverlaps — Scene.setFrustum(int var1,int var2,int var3,Polygon[])
// (Scene.java:663-715). The reinsertion walk: for each visited polygon, look
// back through the window [i+1 .. i+span] for an overlapping polygon that must
// be painted earlier, and reinsert it before this one (via splitOverlap).
//
//	count = s.polyHead, span = 100 (Scene.java:2716 passes 100).
//
// =============================================================================
func (s *Scene) resolveOverlaps(count, span int) {
	last := count - 1
	for i := 0; i <= last; i++ { // Scene.java:670-674 init
		p := s.polygons[i]
		p.mC = false
		p.mF = int32(i)
		p.mP = -1
	}

	i := 0
	for { // Scene.java:679-710
		for !s.polygons[i].mC {
			if i == last {
				return
			}
			cur := s.polygons[i]
			cur.mC = true
			lowBound := i
			highBound := i + span
			if highBound >= last {
				highBound = last - 1
			}
			for j := highBound; j >= 1+lowBound; j-- {
				other := s.polygons[j]
				if other.maxP6 > cur.minP6 && other.minP6 < cur.maxP6 &&
					other.maxP2 > cur.minP2 && other.minP2 < cur.maxP2 &&
					cur.mF != other.mP &&
					!s.polygonHit2(other, cur) && s.polygonHit1(other, cur) {
					s.splitOverlap(lowBound, j)
					lowBound = int(s.mE)
					if s.polygons[j] != other {
						j++
					}
					other.mP = cur.mF
				}
			}
		}
		i++
	}
}

// =============================================================================
// splitOverlap — Scene.setFrustum(int var1,Polygon[],int var3,byte)
// (Scene.java:1383-1441). Reinserts the overlapping polygon at index var3
// before the run starting at var1, bubbling it past every polygon it must
// occlude (polygonHit2). Writes s.mE / s.mEb.
// =============================================================================
func (s *Scene) splitOverlap(var1, var3 int) bool {
	for { // Scene.java:1387-1436
		cur := s.polygons[var1]
		for j := var1 + 1; var3 >= j; j++ {
			other := s.polygons[j]
			if !s.polygonHit2(cur, other) {
				break
			}
			s.polygons[var1] = other
			var1 = j
			s.polygons[j] = cur
			if var3 == j {
				s.mEb = int32(j - 1)
				s.mE = int32(j)
				return true
			}
		}

		end := s.polygons[var3]
		for j := var3 - 1; j >= var1; j-- {
			other := s.polygons[j]
			if !s.polygonHit2(other, end) {
				break
			}
			s.polygons[var3] = other
			s.polygons[j] = end
			var3 = j
			if j == var1 {
				s.mEb = int32(j)
				s.mE = int32(j + 1)
				return true
			}
		}

		if var1+1 >= var3 {
			s.mEb = int32(var3)
			s.mE = int32(var1)
			return false
		}

		if !s.splitOverlap(var1+1, var3) {
			s.mE = int32(var1)
			return false
		}
		var3 = int(s.mEb)
	}
}

// =============================================================================
// polygonHit1 — Scene.polygonHit1(Polygon,Polygon) (Scene.java:143-203).
// Plane-side test of polyA's verts against polyB's plane, then B against A.
// Returns true when the two faces do NOT pierce (so order is already correct).
// =============================================================================
func (s *Scene) polygonHit1(polyA, polyB *Polygon) bool {
	modelA := polyA.model
	modelB := polyB.model
	faceA := polyA.faceID
	faceB := polyB.faceID
	indexA := modelA.faceIndices[faceA]
	indexB := modelB.faceIndices[faceB]
	indexCountA := modelA.faceIndexCount[faceA]
	indexCountB := modelB.faceIndexCount[faceB]

	bv0X := modelB.vertXRot[indexB[0]] // Scene.java:154-162
	bv0Y := modelB.vertYRot[indexB[0]]
	bv0Z := modelB.vertZRot[indexB[0]]
	bnX := polyB.normalX
	bnY := polyB.normalY
	bnZ := polyB.normalZ
	bfNormMag := modelB.scenePolyNormalMagnitude[faceB]
	hit := false
	orientation := polyB.orientation

	for v := 0; indexCountA > v; v++ { // Scene.java:164-172
		vID := indexA[v]
		dot := bnY*(bv0Y-modelA.vertYRot[vID]) + (bv0X-modelA.vertXRot[vID])*bnX +
			(bv0Z-modelA.vertZRot[vID])*bnZ
		if (-bfNormMag > dot && orientation < 0) || (dot > bfNormMag && orientation > 0) {
			hit = true
			break
		}
	}

	if !hit { // Scene.java:174-198
		return true
	}
	bv0X = modelA.vertXRot[indexA[0]]
	bnZ = polyA.normalZ
	bfNormMag = modelA.scenePolyNormalMagnitude[faceA]
	bv0Z = modelA.vertZRot[indexA[0]]
	bv0Y = modelA.vertYRot[indexA[0]]
	hit = false
	bnX = polyA.normalX
	bnY = polyA.normalY
	orientation = polyA.orientation

	for v := 0; v < indexCountB; v++ {
		vID := indexB[v]
		dot := bnX*(bv0X-modelB.vertXRot[vID]) -
			(-(bnY * (bv0Y - modelB.vertYRot[vID])) - (bv0Z-modelB.vertZRot[vID])*bnZ)
		if (dot < -bfNormMag && orientation > 0) || (bfNormMag < dot && orientation < 0) {
			hit = true
			break
		}
	}
	return !hit
}

// =============================================================================
// polygonHit2 — Scene.polygonHit2(byte,Polygon,Polygon) (Scene.java:245-381).
// AABB + plane-side early-outs, then the full screen-space overlap test
// (overlapTest). Returns true when var2 does NOT need to be drawn before var3.
// (the leading byte arg in Java is a dead obfuscation parameter; dropped here).
// =============================================================================
func (s *Scene) polygonHit2(var2, var3 *Polygon) bool {
	if var3.minP6 >= var2.maxP6 { // Scene.java:248-264
		return true
	}
	if var2.minP6 >= var3.maxP6 {
		return true
	}
	if var2.maxP2 <= var3.minP2 {
		return true
	}
	if var3.maxP2 <= var2.minP2 {
		return true
	}
	if var2.maxZ <= var3.minZ {
		return true
	}
	if var3.maxZ < var2.minZ {
		return false
	}

	modelB := var3.model // Scene.java:266-281
	modelA := var2.model
	faceB := var3.faceID
	faceA := var2.faceID
	indexB := modelB.faceIndices[faceB]
	indexA := modelA.faceIndices[faceA]
	indexCountB := modelB.faceIndexCount[faceB]
	indexCountA := modelA.faceIndexCount[faceA]
	av0X := modelA.vertXRot[indexA[0]]
	av0Y := modelA.vertYRot[indexA[0]]
	av0Z := modelA.vertZRot[indexA[0]]
	anX := var2.normalX
	anY := var2.normalY
	anZ := var2.normalZ
	aMag := modelA.scenePolyNormalMagnitude[faceA]
	aOri := var2.orientation
	hit := false

	for k := 0; k < indexCountB; k++ { // Scene.java:287-295 (B verts vs A plane)
		vID := indexB[k]
		dot := (av0Z-modelB.vertZRot[vID])*anZ + (av0Y-modelB.vertYRot[vID])*anY +
			anX*(av0X-modelB.vertXRot[vID])
		if (dot < -aMag && aOri < 0) || (dot > aMag && aOri > 0) {
			hit = true
			break
		}
	}
	if !hit { // Scene.java:297-298
		return true
	}

	hit = false // Scene.java:300-318 (A verts vs B plane)
	bOri := var3.orientation
	bv0Y := modelB.vertYRot[indexB[0]]
	bv0X := modelB.vertXRot[indexB[0]]
	bMag := modelB.scenePolyNormalMagnitude[faceB]
	bv0Z := modelB.vertZRot[indexB[0]]
	bnY := var3.normalY
	bnZ := var3.normalZ
	bnX := var3.normalX

	for k := 0; indexCountA > k; k++ {
		vID := indexA[k]
		dot := (bv0Z-modelA.vertZRot[vID])*bnZ + (bv0X-modelA.vertXRot[vID])*bnX +
			(bv0Y-modelA.vertYRot[vID])*bnY
		if (-bMag > dot && bOri > 0) || (bMag < dot && bOri < 0) {
			hit = true
			break
		}
	}
	if !hit { // Scene.java:320-321
		return true
	}

	// screen-space polygon overlap (Scene.java:323-373). The 2-vertex billboard
	// faces are padded into a +/-20 quad before the test.
	var bx, by []int32
	if indexCountB != 2 { // var30/var24 from modelB face B
		bx = make([]int32, indexCountB)
		by = make([]int32, indexCountB)
		for k := 0; k < indexCountB; k++ {
			vid := indexB[k]
			bx[k] = modelB.vertexParam6[vid]
			by[k] = modelB.vertexParam2[vid]
		}
	} else {
		bx = make([]int32, 4)
		by = make([]int32, 4)
		v1 := indexB[1]
		v0 := indexB[0]
		bx[0] = modelB.vertexParam6[v0] - 20
		bx[1] = modelB.vertexParam6[v1] - 20
		bx[2] = 20 + modelB.vertexParam6[v1]
		bx[3] = modelB.vertexParam6[v0] + 20
		by[0] = modelB.vertexParam2[v0]
		by[3] = modelB.vertexParam2[v0]
		by[1] = modelB.vertexParam2[v1]
		by[2] = modelB.vertexParam2[v1]
	}

	var ax, ay []int32
	if indexCountA != 2 { // var25/var26 from modelA face A
		ax = make([]int32, indexCountA)
		ay = make([]int32, indexCountA)
		for k := 0; k < indexCountA; k++ {
			vid := indexA[k]
			ax[k] = modelA.vertexParam6[vid]
			ay[k] = modelA.vertexParam2[vid]
		}
	} else {
		ay = make([]int32, 4)
		ax = make([]int32, 4)
		v0 := indexA[0]
		v1 := indexA[1]
		ax[0] = modelA.vertexParam6[v0] - 20
		ax[1] = modelA.vertexParam6[v1] - 20
		ax[2] = modelA.vertexParam6[v1] + 20
		ax[3] = modelA.vertexParam6[v0] + 20
		ay[0] = modelA.vertexParam2[v0]
		ay[3] = modelA.vertexParam2[v0]
		ay[1] = modelA.vertexParam2[v1]
		ay[2] = modelA.vertexParam2[v1]
	}

	// Scene.java:373 calls setFrustum(var25, var24, var30, var26, 1) where
	// var25=modelA.X (vertexParam6), var24=modelB.Y (vertexParam2), var30=modelB.X,
	// var26=modelA.Y — i.e. (A.X, B.Y, B.X, A.Y). With ax=A.X, ay=A.Y, bx=B.X,
	// by=B.Y that is overlapTest(ax, by, bx, ay, 1). The Y lists are interleaved
	// (B.Y paired with A.X's list, A.Y with B.X's) so A and B may differ in vertex
	// count (quad vs tri). (A prior pass had this as (bx,ay,ax,by) — wrong order.)
	return !s.overlapTest(ax, by, bx, ay, 1)
}

// booleanCombinatoric — Scene.booleanCombinatoric (Scene.java:214-243).
func (s *Scene) booleanCombinatoric(var2 bool, var3, var4, var5, var6 int32) bool {
	if (!var2 || var5 > var6) && var5 >= var6 {
		if var5 < var4 {
			return true
		}
		if var3 < var6 {
			return true
		}
		if var3 < var4 {
			return true
		}
		return var2
	}
	if var5 > var4 {
		return true
	}
	if var3 > var6 {
		return true
	}
	if var3 > var4 {
		return true
	}
	return !var2
}

// booleanCombinatoric2 — Scene.booleanCombinatoric2 (Scene.java:415-422).
func (s *Scene) booleanCombinatoric2(var1 int32, var2 bool, var3, var5 int32) bool {
	if (!var2 || var3 > var1) && var3 >= var1 {
		if var5 < var1 {
			return true
		}
		return var2
	}
	if var1 >= var5 {
		return !var2
	}
	return true
}

// booleanCombinatoric3 — Scene.booleanCombinatoric3 (Scene.java:424-432).
func (s *Scene) booleanCombinatoric3(var1 int32, var3, var4, var5, var6 int32) int32 {
	if var4 == var1 {
		return var6
	}
	return (var5-var6)*(var3-var1)/(var4-var1) + var6
}

// =============================================================================
// overlapTest — Scene.setFrustum(int[] var1,int[] var2,int[] var3,int[] var4,
// int var5) (Scene.java:1443-1841). The exact screen-space polygon/polygon
// boundary-walk overlap test. var1/var2 are one polygon's X/Y, var3/var4 the
// other's; returns true when they overlap. Ported verbatim (int32 throughout).
//
//	In the caller (polygonHit2) it is invoked as overlapTest(ax,ay,bx,by,1):
//	var1=ax, var2=ay, var3=bx, var4=by, var5=1.
//
// =============================================================================
func (s *Scene) overlapTest(var1, var2, var3, var4 []int32, var5 int) bool {
	var6 := len(var3) // Scene.java:1446-1447
	var7 := len(var1)
	var16 := 0

	var8 := 0 // Scene.java:1449-1464 — min/max of var2, index of min in var8
	var18 := var2[0]
	var20 := var18
	for var22 := 1; var6 > var22; var22++ {
		if var2[var22] >= var18 {
			if var2[var22] > var20 {
				var20 = var2[var22]
			}
		} else {
			var8 = var22
			var18 = var2[var22]
		}
	}

	var10 := 0 // Scene.java:1466-1478 — min/max of var4 from var5, index in var10
	var19 := var4[0]
	var21 := var19
	for var22 := var5; var7 > var22; var22++ {
		if var4[var22] >= var19 {
			if var21 < var4[var22] {
				var21 = var4[var22]
			}
		} else {
			var19 = var4[var22]
			var10 = var22
		}
	}

	if !(var19 < var20) { // Scene.java:1480 / 1834
		return false
	}
	if var21 <= var18 { // Scene.java:1481-1482
		return false
	}

	var var9, var11 int // array indices (Java int)
	var var12, var13, var14, var15 int32
	var17 := false

	if var4[var10] > var2[var8] { // Scene.java:1491-1514
		for var9 = var8; var2[var8] < var4[var10]; var8 = (var8 - (1 - var6)) % var6 {
		}
		for var2[var9] < var4[var10] {
			var9 = (1 + var9) % var6
		}
		var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var10], var2[var8],
			var3[var8], var3[(1+var8)%var6])
		var13 = s.booleanCombinatoric3(var2[(var6+(var9-1))%var6], var4[var10],
			var2[var9], var3[var9], var3[(var6-1+var9)%var6])
		var14 = var1[var10]
		var17 = var12 < var14 || var13 < var14
		if s.booleanCombinatoric2(var14, var17, var12, var13) {
			return true
		}
		var11 = (var10 + 1) % var7
		var10 = (var10 + var7 - 1) % var7
		if var8 == var9 {
			var16 = 1
		}
	} else { // Scene.java:1515-1538
		for var11 = var10; var2[var8] > var4[var10]; var10 = (var10 + var7 - 1) % var7 {
		}
		for var12 = var3[var8]; var2[var8] > var4[var11]; var11 = (var11 + 1) % var7 {
		}
		var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var2[var8], var4[var10],
			var1[var10], var1[(var10+1)%var7])
		var15 = s.booleanCombinatoric3(var4[(var7+(var11-1))%var7], var2[var8],
			var4[var11], var1[var11], var1[(var11-1+var7)%var7])
		var17 = var12 < var14 || var12 < var15
		if s.booleanCombinatoric2(var12, !var17, var14, var15) {
			return true
		}
		var9 = (1 + var8) % var6
		var8 = (var6 + (var8 - 1)) % var6
		if var10 == var11 {
			var16 = 2
		}
	}

	for var16 == 0 { // Scene.java:1540-1676
		if var2[var8] >= var2[var9] {
			if var2[var9] >= var4[var10] {
				if var4[var10] >= var4[var11] {
					var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var11],
						var2[var8], var3[var8], var3[(1+var8)%var6])
					var13 = s.booleanCombinatoric3(var2[(var9-1+var6)%var6],
						var4[var11], var2[var9], var3[var9], var3[(var6+(var9-1))%var6])
					var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var4[var11],
						var4[var10], var1[var10], var1[(var10+1)%var7])
					var15 = var1[var11]
					if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
						return true
					}
					var11 = (1 + var11) % var7
					if var11 == var10 {
						var16 = 2
					}
				} else {
					var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var10],
						var2[var8], var3[var8], var3[(1+var8)%var6])
					var13 = s.booleanCombinatoric3(var2[(var9+var6-1)%var6],
						var4[var10], var2[var9], var3[var9], var3[(var6-1+var9)%var6])
					var14 = var1[var10]
					var15 = s.booleanCombinatoric3(var4[(var7+(var11-1))%var7],
						var4[var10], var4[var11], var1[var11], var1[(var11-1+var7)%var7])
					if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
						return true
					}
					var10 = (var10 - 1 + var7) % var7
					if var11 == var10 {
						var16 = 2
					}
				}
			} else if var2[var9] < var4[var11] {
				var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var2[var9],
					var2[var8], var3[var8], var3[(1+var8)%var6])
				var13 = var3[var9]
				var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var2[var9],
					var4[var10], var1[var10], var1[(1+var10)%var7])
				var15 = s.booleanCombinatoric3(var4[(var11-1+var7)%var7], var2[var9],
					var4[var11], var1[var11], var1[(var11-(1-var7))%var7])
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var9 = (var9 + 1) % var6
				if var8 == var9 {
					var16 = 1
				}
			} else {
				var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var11],
					var2[var8], var3[var8], var3[(var8+1)%var6])
				var13 = s.booleanCombinatoric3(var2[(var6+var9-1)%var6], var4[var11],
					var2[var9], var3[var9], var3[(var6-1+var9)%var6])
				var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var4[var11],
					var4[var10], var1[var10], var1[(var10+1)%var7])
				var15 = var1[var11]
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var11 = (var11 + 1) % var7
				if var10 == var11 {
					var16 = 2
				}
			}
		} else if var4[var10] > var2[var8] {
			if var2[var8] >= var4[var11] {
				var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var11],
					var2[var8], var3[var8], var3[(1+var8)%var6])
				var13 = s.booleanCombinatoric3(var2[(var6+(var9-1))%var6], var4[var11],
					var2[var9], var3[var9], var3[(var6+(var9-1))%var6])
				var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var4[var11],
					var4[var10], var1[var10], var1[(1+var10)%var7])
				var15 = var1[var11]
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var11 = (1 + var11) % var7
				if var10 == var11 {
					var16 = 2
				}
			} else {
				var12 = var3[var8]
				var13 = s.booleanCombinatoric3(var2[(var9+(var6-1))%var6], var2[var8],
					var2[var9], var3[var9], var3[(var9+var6-1)%var6])
				var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var2[var8],
					var4[var10], var1[var10], var1[(1+var10)%var7])
				var15 = s.booleanCombinatoric3(var4[(var7-1+var11)%var7], var2[var8],
					var4[var11], var1[var11], var1[(var7+(var11-1))%var7])
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var8 = (var6 + (var8 - 1)) % var6
				if var8 == var9 {
					var16 = 1
				}
			}
		} else if var4[var10] < var4[var11] {
			var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var10], var2[var8],
				var3[var8], var3[(1+var8)%var6])
			var13 = s.booleanCombinatoric3(var2[(var6+(var9-1))%var6], var4[var10],
				var2[var9], var3[var9], var3[(var6-1+var9)%var6])
			var14 = var1[var10]
			var15 = s.booleanCombinatoric3(var4[(var7+(var11-1))%var7], var4[var10],
				var4[var11], var1[var11], var1[(var7+var11-1)%var7])
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var10 = (var7 + var10 - 1) % var7
			if var11 == var10 {
				var16 = 2
			}
		} else {
			var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var11], var2[var8],
				var3[var8], var3[(var8+1)%var6])
			var13 = s.booleanCombinatoric3(var2[(var9+var6-1)%var6], var4[var11],
				var2[var9], var3[var9], var3[(var6+(var9-1))%var6])
			var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var4[var11], var4[var10],
				var1[var10], var1[(1+var10)%var7])
			var15 = var1[var11]
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var11 = (var11 + 1) % var7
			if var10 == var11 {
				var16 = 2
			}
		}
	}

	for var16 == 1 { // Scene.java:1678-1743
		if ^var2[var8] <= ^var4[var10] {
			if var4[var10] < var4[var11] {
				var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var10], var2[var8],
					var3[var8], var3[(var8+1)%var6])
				var13 = s.booleanCombinatoric3(var2[(var9-1+var6)%var6], var4[var10],
					var2[var9], var3[var9], var3[(var9-1+var6)%var6])
				var14 = var1[var10]
				var15 = s.booleanCombinatoric3(var4[(var11-1+var7)%var7], var4[var10],
					var4[var11], var1[var11], var1[(var11+(var7-1))%var7])
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var10 = (var10 + (var7 - 1)) % var7
				if var10 == var11 {
					var16 = 0
				}
			} else {
				var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var11], var2[var8],
					var3[var8], var3[(1+var8)%var6])
				var13 = s.booleanCombinatoric3(var2[(var6+var9-1)%var6], var4[var11],
					var2[var9], var3[var9], var3[(var9+(var6-1))%var6])
				var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var4[var11],
					var4[var10], var1[var10], var1[(1+var10)%var7])
				var15 = var1[var11]
				if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
					return true
				}
				var11 = (1 + var11) % var7
				if var10 == var11 {
					var16 = 0
				}
			}
		} else {
			if var4[var11] > var2[var8] {
				var12 = var3[var8]
				var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var2[var8], var4[var10],
					var1[var10], var1[(1+var10)%var7])
				var15 = s.booleanCombinatoric3(var4[(var11-1+var7)%var7], var2[var8],
					var4[var11], var1[var11], var1[(var7+(var11-1))%var7])
				if !s.booleanCombinatoric2(var12, !var17, var14, var15) {
					return false
				}
				return true
			}
			var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var11], var2[var8],
				var3[var8], var3[(1+var8)%var6])
			var13 = s.booleanCombinatoric3(var2[(var9+var6-1)%var6], var4[var11],
				var2[var9], var3[var9], var3[(var9+var6-1)%var6])
			var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var4[var11], var4[var10],
				var1[var10], var1[(1+var10)%var7])
			var15 = var1[var11]
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var11 = (1 + var11) % var7
			if var10 == var11 {
				var16 = 0
			}
		}
	}

	for var16 == 2 { // Scene.java:1745-1808
		if var4[var10] < var2[var8] {
			if var4[var10] < var2[var9] {
				var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var4[var10], var2[var8],
					var3[var8], var3[(var8+1)%var6])
				var13 = s.booleanCombinatoric3(var2[(var9-1+var6)%var6], var4[var10],
					var2[var9], var3[var9], var3[(var6-1+var9)%var6])
				var14 = var1[var10]
				if !s.booleanCombinatoric2(var14, var17, var12, var13) {
					return false
				}
				return true
			}
			var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var2[var9], var2[var8],
				var3[var8], var3[(1+var8)%var6])
			var13 = var3[var9]
			var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var2[var9], var4[var10],
				var1[var10], var1[(var10+1)%var7])
			var15 = s.booleanCombinatoric3(var4[(var11-1+var7)%var7], var2[var9],
				var4[var11], var1[var11], var1[(var11+var7-1)%var7])
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var9 = (1 + var9) % var6
			if var8 == var9 {
				var16 = 0
			}
		} else if var2[var8] >= var2[var9] {
			var12 = s.booleanCombinatoric3(var2[(var8+1)%var6], var2[var9], var2[var8],
				var3[var8], var3[(1+var8)%var6])
			var13 = var3[var9]
			var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var2[var9], var4[var10],
				var1[var10], var1[(1+var10)%var7])
			var15 = s.booleanCombinatoric3(var4[(var11-(1-var7))%var7], var2[var9],
				var4[var11], var1[var11], var1[(var11+var7-1)%var7])
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var9 = (1 + var9) % var6
			if var8 == var9 {
				var16 = 0
			}
		} else {
			var12 = var3[var8]
			var13 = s.booleanCombinatoric3(var2[(var9+var6-1)%var6], var2[var8],
				var2[var9], var3[var9], var3[(var6+var9-1)%var6])
			var14 = s.booleanCombinatoric3(var4[(var10+1)%var7], var2[var8], var4[var10],
				var1[var10], var1[(var10+1)%var7])
			var15 = s.booleanCombinatoric3(var4[(var11+(var7-1))%var7], var2[var8],
				var4[var11], var1[var11], var1[(var7+var11-1)%var7])
			if s.booleanCombinatoric(var17, var13, var15, var12, var14) {
				return true
			}
			var8 = (var6 + var8 - 1) % var6
			if var9 == var8 {
				var16 = 0
			}
		}
	}

	if var4[var10] <= var2[var8] { // Scene.java:1810-1832
		var12 = s.booleanCombinatoric3(var2[(1+var8)%var6], var4[var10], var2[var8],
			var3[var8], var3[(var8+1)%var6])
		var13 = s.booleanCombinatoric3(var2[(var9-1+var6)%var6], var4[var10],
			var2[var9], var3[var9], var3[(var6+(var9-1))%var6])
		var14 = var1[var10]
		return s.booleanCombinatoric2(var14, var17, var12, var13)
	}
	var12 = var3[var8]
	var14 = s.booleanCombinatoric3(var4[(1+var10)%var7], var2[var8], var4[var10],
		var1[var10], var1[(var10+1)%var7])
	var15 = s.booleanCombinatoric3(var4[(var7+(var11-1))%var7], var2[var8],
		var4[var11], var1[var11], var1[(var7-1+var11)%var7])
	return s.booleanCombinatoric2(var12, !var17, var14, var15)
}
