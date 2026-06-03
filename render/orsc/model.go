package orsc

import "math"

// model.go — faithful port of RSModel.java
// (/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/RSModel.java).
//
// Method bodies replace the stubs declared in types.go. The struct itself is
// defined in types.go and is NOT redefined here.
//
// Java->Go conventions used throughout:
//   - Java int is 32-bit; the whole fixed-point geometry/projection pipeline is
//     int32 with deliberate two's-complement overflow (>> is arithmetic).
//   - Java signed byte[] -> []int8; every read masks &0xff via int(b)&0xff,
//     mirroring FastMath.bitwiseAnd(255, b) / `255 & data[..]`.
//   - 32767 fill marker -> TRANSPARENT (= m_Vb = 12345678), RSModel.java:125-137.

// =============================================================================
// Frustum globals — mirror of orsc.MiscFunctions.frustum{NearZ,MinY,MinX,FarZ,
// MaxY,MaxX} (MiscFunctions.java:12,13,18,20,21,22). In OpenRSC these are static
// fields on MiscFunctions that Scene.endScene zeroes + grows from the frustum
// corners (Scene.java:2560-2580) and RSModel.rotate1024 reads to frustum-test a
// model (RSModel.java:1018-1020). They are render-pass scratch shared between
// Scene and RSModel, so they live at package scope here. The Scene port writes
// them; project() reads them.
// =============================================================================
var (
	frustumNearZ int32 // MiscFunctions.frustumNearZ
	frustumFarZ  int32 // MiscFunctions.frustumFarZ
	frustumMinX  int32 // MiscFunctions.frustumMinX
	frustumMaxX  int32 // MiscFunctions.frustumMaxX
	frustumMinY  int32 // MiscFunctions.frustumMinY
	frustumMaxY  int32 // MiscFunctions.frustumMaxY
)

// =============================================================================
// Construction
// =============================================================================

// NewModel mirrors the empty-capacity ctor RSModel(int vertexCount, int
// faceCount) (RSModel.java:171-183): it runs setFaceVertexCount to allocate the
// arrays and seeds faceParam1[face][0]=face. Geometry is then filled by
// insertVertex/insertFace (or by the FromAssets adapter below for .ob3 data).
func NewModel(vertexCount, faceCount int) *Model {
	m := &Model{}
	// RSModel ctor defaults (RSModel.java:21,30,43-48 etc.) that the Java field
	// initialisers set before setFaceVertexCount runs.
	m.diffuseParam1 = 32         // RSModel.java:21
	m.diffuseDirX = 180          // RSModel.java:43
	m.diffuseDirY = 155          // RSModel.java:44
	m.diffuseDirZ = 95           // RSModel.java:45
	m.diffuseMag = 256           // RSModel.java:46
	m.diffuseParam2 = 512        // RSModel.java:47
	m.dontComputeDiffuse = false // RSModel.java:48
	m.mDc = true                 // RSModel.m_dc=true (RSModel.java:27)

	m.setFaceVertexCount(faceCount, vertexCount, 69)

	// faceParam1 is RSModel.java:174-178 — a [faceCount][1] table seeded face->face.
	// We track it only as facePickIndex (the single useful column); the full
	// faceParam1 matrix is World-generation bookkeeping not needed by the renderer.
	for face := 0; face < faceCount; face++ {
		if !m.mDb {
			m.facePickIndex[face] = int32(face)
		}
	}
	return m
}

// FromAssets adapts a decoded .ob3 *assets.Model into the RSModel layout — the
// RSModel(byte[],offset,var3) ctor (RSModel.java:92-169) fed pre-decoded
// geometry. The caller passes the asset fields directly so this package needn't
// import assets.
//
// modelData arrives with the 32767 fill marker already mapped to assets.Magic
// (== TRANSPARENT == 12345678) by DecodeModel, so the RSModel.java:125-126/135-137
// remap is a no-op here; faceIntensity arrives as 0 (flat) or Magic (per-vertex),
// matching RSModel.java:141-148. Both sentinels equal TRANSPARENT.
//
// Field correspondence (assets.Model -> RSModel):
//
//	VertexX/Y/Z      -> vertX/vertY/vertZ          (RSModel.java:103-115)
//	FaceVertices     -> faceIndices                (RSModel.java:150-161)
//	FaceNumVertices  -> faceIndexCount             (RSModel.java:119-121)
//	FaceFillFront    -> faceTextureFront           (RSModel.java:123-130)
//	FaceFillBack     -> faceTextureBack            (RSModel.java:132-138)
//	FaceIntensity    -> faceDiffuseLight           (RSModel.java:141-148)
func FromAssets(
	vertexCount, faceCount int,
	vertX, vertY, vertZ []int32,
	faceIndices [][]int,
	faceNumVertices []int,
	faceFillFront, faceFillBack, faceIntensity []int32,
) *Model {
	m := NewModel(vertexCount, faceCount)

	// Copy geometry (RSModel.java:102-117).
	for j := 0; j < vertexCount; j++ {
		m.vertX[j] = vertX[j]
		m.vertY[j] = vertY[j]
		m.vertZ[j] = vertZ[j]
	}
	m.vertHead = vertexCount

	// faceIndexCount (RSModel.java:119-121).
	for j := 0; j < faceCount; j++ {
		m.faceIndexCount[j] = faceNumVertices[j]
	}

	// faceTextureFront / faceTextureBack (RSModel.java:123-138). 32767 -> m_Vb is
	// already applied upstream (== TRANSPARENT). Mirror it defensively anyway.
	for j := 0; j < faceCount; j++ {
		f := faceFillFront[j]
		if f == 32767 {
			f = TRANSPARENT
		}
		m.faceTextureFront[j] = f
	}
	for j := 0; j < faceCount; j++ {
		f := faceFillBack[j]
		if f == 32767 {
			f = TRANSPARENT
		}
		m.faceTextureBack[j] = f
	}

	// faceDiffuseLight (RSModel.java:141-148): the .ob3 "intensity" byte is 0 =>
	// flat (0) or non-zero => per-vertex gouraud sentinel (m_Vb). Upstream already
	// collapsed it to 0 or Magic; treat anything != 0 as the sentinel.
	for j := 0; j < faceCount; j++ {
		if faceIntensity[j] != 0 {
			m.faceDiffuseLight[j] = TRANSPARENT
		} else {
			m.faceDiffuseLight[j] = 0
		}
	}

	// faceIndices (RSModel.java:150-161). Copy so callers may reuse their slices.
	for j := 0; j < faceCount; j++ {
		src := faceIndices[j]
		dst := make([]int, m.faceIndexCount[j])
		copy(dst, src)
		m.faceIndices[j] = dst
	}

	m.faceHead = faceCount
	m.mYb = 1 // RSModel.java:164: full rebuild on first project.
	return m
}

// setFaceVertexCount allocates all per-vertex / per-face arrays and resets the
// transform to identity (RSModel.java:1225-1305). The Java `var3` gate at
// :1296-1298 nulls faceIndexCount when var3<=68; we always need it, and every
// caller in this port passes 69 or 97 (>68), so it is always allocated.
func (m *Model) setFaceVertexCount(faceCount, vertexCount, var3 int) {
	if !m.mDb { // RSModel.java:1227-1230
		m.mZb = make([]int8, faceCount)
		m.facePickIndex = make([]int32, faceCount)
	}

	m.scenePolyNormalMagnitude = make([]int32, faceCount) // :1232
	m.vertY = make([]int32, vertexCount)                  // :1233
	m.vertLightOther = make([]int8, vertexCount)          // :1234
	m.vertX = make([]int32, vertexCount)                  // :1235
	m.scenePolyNormalShift = make([]int32, faceCount)     // :1236
	m.vertZ = make([]int32, vertexCount)                  // :1237
	m.faceTextureBack = make([]int32, faceCount)          // :1238
	m.vertDiffuseLight = make([]int32, vertexCount)       // :1239
	m.faceTextureFront = make([]int32, faceCount)         // :1240
	m.faceIndices = make([][]int, faceCount)              // :1241
	m.faceIndexCount = make([]int, faceCount)             // :1242

	if !m.mB { // RSModel.java:1243-1249
		m.vertexParam2 = make([]int32, vertexCount)
		m.vertYRot = make([]int32, vertexCount)
		m.vertXRot = make([]int32, vertexCount)
		m.vertZRot = make([]int32, vertexCount)
		m.vertexParam6 = make([]int32, vertexCount)
	}

	m.faceDiffuseLight = make([]int32, faceCount) // :1252

	// Identity rotation-matrix terms (RSModel.java:1253-1257,1273,1290-1291).
	m.rotMzToX = 256
	m.rot256Y = 0
	m.rotMxToZ = 256
	m.rotMyToX = 256
	m.scaleZ = 256

	if !m.mC { // RSModel.java:1258-1265
		m.faceMaxX = make([]int32, faceCount)
		m.faceMinY = make([]int32, faceCount)
		m.faceMinZ = make([]int32, faceCount)
		m.faceMaxY = make([]int32, faceCount)
		m.faceMaxZ = make([]int32, faceCount)
		m.faceMinX = make([]int32, faceCount)
	}

	if !m.dontComputeDiffuse || !m.mC { // RSModel.java:1267-1271
		m.faceNormY = make([]int32, faceCount)
		m.faceNormZ = make([]int32, faceCount)
		m.faceNormX = make([]int32, faceCount)
	}

	m.rotMyToZ = 256 // :1273
	m.faceHead = 0   // :1274
	m.rot256X = 0    // :1275
	m.translateZ = 0 // :1276
	m.scaleX = 256   // :1277

	if !m.mV { // RSModel.java:1278-1286
		m.vertXTransform = make([]int32, vertexCount)
		m.vertZTransform = make([]int32, vertexCount)
		m.vertYTransform = make([]int32, vertexCount)
	} else {
		// Alias the transform working copy onto the model-local arrays.
		m.vertXTransform = m.vertX
		m.vertZTransform = m.vertZ
		m.vertYTransform = m.vertY
	}

	m.translateX = 0             // :1288
	m.vertexCount2 = vertexCount // :1289
	m.rotMxToY = 256             // :1290
	m.rotMzToY = 256             // :1291
	m.translateY = 0             // :1292
	m.appliedTransform = 0       // :1293
	m.vertHead = 0               // :1294
	m.scaleY = 256               // :1295

	// RSModel.java:1296-1298: faceIndexCount nulled when var3<=68. Never hit here.
	if var3 <= 68 {
		m.faceIndexCount = nil
	}

	m.faceCount = faceCount // :1300
	m.rot256Z = 0           // :1301
}

// insertVertex mirrors RSModel.insertVertex (RSModel.java:883-904): dedup an
// existing matching vertex, else append. Returns the vertex id or -1 if full.
func (m *Model) insertVertex(x, y, z int32) int {
	for i := 0; i < m.vertHead; i++ {
		if x == m.vertX[i] && y == m.vertY[i] && z == m.vertZ[i] {
			return i
		}
	}
	if m.vertHead < m.vertexCount2 {
		m.vertX[m.vertHead] = x
		m.vertY[m.vertHead] = y
		m.vertZ[m.vertHead] = z
		m.vertHead++
		return m.vertHead - 1
	}
	return -1
}

// insertFace mirrors RSModel.insertFace (RSModel.java:860-881). The var5
// "addRotation" side effect (:863-865) is never used by the renderer paths, so
// it is omitted; callers always pass var5=false.
func (m *Model) insertFace(indexCount int, indices []int, texFront, texBack int32) int {
	if m.faceCount > m.faceHead {
		m.faceIndexCount[m.faceHead] = indexCount
		m.faceIndices[m.faceHead] = indices
		m.faceTextureFront[m.faceHead] = texFront
		m.faceTextureBack[m.faceHead] = texBack
		m.mYb = 1
		m.faceHead++
		return m.faceHead - 1
	}
	return -1
}

// =============================================================================
// Transform setters (public API + computeAppliedTransform)
// =============================================================================

// Translate is RSModel.setTranslate (RSModel.java:1320-1331): set the absolute
// translate, recompute the applied-transform tier, mark the cache dirty.
func (m *Model) Translate(tx, ty, tz int32) {
	m.translateY = ty
	m.translateX = tx
	m.translateZ = tz
	m.computeAppliedTransform()
	m.mYb = 1
}

// Orient is RSModel.setRot256 (RSModel.java:1307-1318): set the absolute 256-step
// rotation (masked &255), recompute the applied-transform tier, mark dirty.
func (m *Model) Orient(rx, ry, rz int32) {
	m.rot256X = rx & 255
	m.rot256Y = ry & 255
	m.rot256Z = rz & 255
	m.computeAppliedTransform()
	m.mYb = 1
}

// addRotation is RSModel.addRotation (RSModel.java:383-394): relative rotate.
func (m *Model) addRotation(rotX, rotY, rotZ int32) {
	m.rot256X = (rotX + m.rot256X) & 255
	m.rot256Y = (rotY + m.rot256Y) & 255
	m.rot256Z = (m.rot256Z + rotZ) & 255
	m.computeAppliedTransform()
	m.mYb = 1
}

// translate2 is RSModel.translate2 (RSModel.java:1357-1368): relative translate.
func (m *Model) translate2(tx, ty, tz int32) {
	m.translateY += ty
	m.translateZ += tz
	m.translateX += tx
	m.computeAppliedTransform()
	m.mYb = 1
}

// computeAppliedTransform is RSModel.computeAppliedTransform (RSModel.java:591-615):
// pick the smallest transform tier (0..4) that the current rot/scale/translate
// state needs. resetTransformCache applies stages >= each tier (RSModel.java:987-1002).
//
//	4: a rotation-matrix term differs from identity (256)
//	3: a scale differs from 256
//	2: a rot256 angle is non-zero
//	1: a translate is non-zero
//	0: identity
func (m *Model) computeAppliedTransform() {
	if m.rotMyToX == 256 && m.rotMyToZ == 256 && m.rotMzToX == 256 && m.rotMzToY == 256 &&
		m.rotMxToZ == 256 && m.rotMxToY == 256 {
		if m.scaleX == 256 && m.scaleY == 256 && m.scaleZ == 256 {
			if m.rot256X == 0 && m.rot256Y == 0 && m.rot256Z == 0 {
				if m.translateX == 0 && m.translateY == 0 && m.translateZ == 0 {
					m.appliedTransform = 0
				} else {
					m.appliedTransform = 1
				}
			} else {
				m.appliedTransform = 2
			}
		} else {
			m.appliedTransform = 3
		}
	} else {
		m.appliedTransform = 4
	}
}

// =============================================================================
// Transform bake — resetTransformCache + the rotate256/scale/applyRotMatrix/
// translate stages it drives.
// =============================================================================

// resetTransformCache is RSModel.resetTransformCache (RSModel.java:954-1011): the
// lazy bake. m_Yb selects the tier:
//
//	2 -> verbatim copy of vertX/Y/Z into the transform arrays + "infinite" bounds
//	     (the m_T sprite/billboard model, which is laid out pre-transformed).
//	1 -> full rebuild: copy, then rotate256/scale/applyRotMatrix/translate per the
//	     appliedTransform tier, then calculateBoundingBoxes + computeNormals.
//	0 -> clean, nothing to do.
//
// The Java `var1` parameter is a verification arg (always 7972 on the live path;
// :956-958 dead branch is skipped). The `var1-7972` passed to translate (:1001)
// becomes the vertex-offset 0.
func (m *Model) resetTransformCache() {
	if m.mYb == 2 {
		m.mYb = 0
		for i := 0; i < m.vertHead; i++ {
			m.vertXTransform[i] = m.vertX[i]
			m.vertYTransform[i] = m.vertY[i]
			m.vertZTransform[i] = m.vertZ[i]
		}
		// RSModel.java:971-977 — bounds set so the frustum test always passes.
		m.maxX = 9999999
		m.maxY = 9999999
		m.minY = -9999999
		m.maxZ = 9999999
		m.minZ = -9999999
		m.maxFaceDimension = 9999999
		m.minX = -9999999
	} else if m.mYb == 1 {
		m.mYb = 0
		for i := 0; i < m.vertHead; i++ {
			m.vertXTransform[i] = m.vertX[i]
			m.vertYTransform[i] = m.vertY[i]
			m.vertZTransform[i] = m.vertZ[i]
		}

		if m.appliedTransform >= 2 { // RSModel.java:987-989
			m.rotate256(m.rot256X, m.rot256Z, m.rot256Y)
		}
		if m.appliedTransform >= 3 { // RSModel.java:991-993
			m.scale(m.scaleX, m.scaleZ, m.scaleY)
		}
		if m.appliedTransform >= 4 { // RSModel.java:995-998
			m.applyRotMatrix(m.rotMxToZ, m.rotMzToY, m.rotMyToX, m.rotMyToZ, m.rotMzToX, m.rotMxToY)
		}
		if m.appliedTransform >= 1 { // RSModel.java:1000-1002 (vertex offset 0)
			m.translate(0, m.translateY, m.translateZ, m.translateX)
		}

		m.calculateBoundingBoxes(999999) // :1004
		m.computeNormals()               // :1005
	}
}

// rotate256 is RSModel.rotate256 (RSModel.java:1097-1134): rotate the baked verts
// in place by the rot256X/Z/Y angles using the 512-entry trigTable256
// (sin[0..255], cos[256..511]). Java arg order is (var1, rotX, rotZ, rotY); the
// body applies Z, then X, then Y. The :1100-1102 faceNormY-null side effect is a
// no-op for us (computeNormals reallocates / overwrites the normals afterwards).
func (m *Model) rotate256(rotX, rotZ, rotY int32) {
	for v := 0; v < m.vertHead; v++ {
		var tmp int32
		if rotZ != 0 {
			xx := trigTable256[rotZ+256]
			xy := trigTable256[rotZ]
			tmp = (m.vertXTransform[v]*xx + m.vertYTransform[v]*xy) >> 15
			m.vertYTransform[v] = (m.vertYTransform[v]*xx - xy*m.vertXTransform[v]) >> 15
			m.vertXTransform[v] = tmp
		}
		if rotX != 0 {
			yz := trigTable256[rotX]
			yy := trigTable256[256+rotX]
			tmp = (yy*m.vertYTransform[v] - yz*m.vertZTransform[v]) >> 15
			m.vertZTransform[v] = (yz*m.vertYTransform[v] + yy*m.vertZTransform[v]) >> 15
			m.vertYTransform[v] = tmp
		}
		if rotY != 0 {
			xz := trigTable256[rotY]
			xx := trigTable256[256+rotY]
			tmp = (xz*m.vertZTransform[v] + m.vertXTransform[v]*xx) >> 15
			m.vertZTransform[v] = (m.vertZTransform[v]*xx - m.vertXTransform[v]*xz) >> 15
			m.vertXTransform[v] = tmp
		}
	}
}

// scale is RSModel.scale (RSModel.java:1139-1155): 8.8 fixed-point scale of the
// baked verts. Java arg order (xScale, var2, zScale, yScale); var2 is a verify arg.
func (m *Model) scale(xScale, zScale, yScale int32) {
	for i := 0; i < m.vertHead; i++ {
		m.vertXTransform[i] = (m.vertXTransform[i] * xScale) >> 8
		m.vertYTransform[i] = (m.vertYTransform[i] * yScale) >> 8
		m.vertZTransform[i] = (zScale * m.vertZTransform[i]) >> 8
	}
}

// applyRotMatrix is RSModel.applyRotMatrix (RSModel.java:399-428): apply the six
// shear-style cross terms (each 8.8) to the baked verts. The order is exactly the
// Java order: yToX, yToZ, zToX, zToY, xToZ, xToY. Note the cross terms read the
// values mutated by earlier terms within the same vertex (matching Java).
func (m *Model) applyRotMatrix(xToZ, zToY, yToX, yToZ, zToX, xToY int32) {
	for i := 0; i < m.vertHead; i++ {
		if yToX != 0 {
			m.vertXTransform[i] += (m.vertYTransform[i] * yToX) >> 8
		}
		if yToZ != 0 {
			m.vertZTransform[i] += (yToZ * m.vertYTransform[i]) >> 8
		}
		if zToX != 0 {
			m.vertXTransform[i] += (zToX * m.vertZTransform[i]) >> 8
		}
		if zToY != 0 {
			m.vertYTransform[i] += (zToY * m.vertZTransform[i]) >> 8
		}
		if xToZ != 0 {
			m.vertZTransform[i] += (xToZ * m.vertXTransform[i]) >> 8
		}
		if xToY != 0 {
			m.vertYTransform[i] += (m.vertXTransform[i] * xToY) >> 8
		}
	}
}

// translate is RSModel.translate (RSModel.java:1342-1355): add the translate to
// the baked verts starting at vertex vOff (always 0 on the live path).
func (m *Model) translate(vOff int, yt, zt, xt int32) {
	for i := vOff; i < m.vertHead; i++ {
		m.vertXTransform[i] += xt
		m.vertYTransform[i] += yt
		m.vertZTransform[i] += zt
	}
}

// =============================================================================
// Bounds + normals + lighting
// =============================================================================

// computeBounds (exported entry per the contract) is calculateBoundingBoxes with
// the live var1=999999 (RSModel.java:430-526 / :1004). It is normally driven by
// resetTransformCache; exposed here so callers may force a recompute.
func (m *Model) computeBounds() { m.calculateBoundingBoxes(999999) }

// calculateBoundingBoxes is RSModel.calculateBoundingBoxes (RSModel.java:430-526):
// per-face (when !m_c) and per-model min/max over the BAKED vertXTransform/Y/Z,
// plus maxFaceDimension (the largest per-face X/Y/Z extent). var1 seeds minY
// (always 999999 live).
func (m *Model) calculateBoundingBoxes(var1 int32) {
	m.minX = 999999
	m.minZ = 999999
	m.maxZ = -999999
	m.minY = var1
	m.maxX = -999999
	m.maxY = -999999
	m.maxFaceDimension = -999999

	for face := 0; face < m.faceHead; face++ {
		fIndex := m.faceIndices[face]
		fIndexCount := m.faceIndexCount[face]
		vID := fIndex[0]
		minZ := m.vertZTransform[vID]
		maxZ := minZ
		minY := m.vertYTransform[vID]
		maxY := minY
		minX := m.vertXTransform[vID]
		maxX := minX

		for vert := 0; vert < fIndexCount; vert++ {
			vID = fIndex[vert]
			if m.vertZTransform[vID] >= minZ {
				if m.vertZTransform[vID] > maxZ {
					maxZ = m.vertZTransform[vID]
				}
			} else {
				minZ = m.vertZTransform[vID]
			}

			if m.vertYTransform[vID] < minY {
				minY = m.vertYTransform[vID]
			} else if m.vertYTransform[vID] > maxY {
				maxY = m.vertYTransform[vID]
			}

			if minX <= m.vertXTransform[vID] {
				if m.vertXTransform[vID] > maxX {
					maxX = m.vertXTransform[vID]
				}
			} else {
				minX = m.vertXTransform[vID]
			}
		}

		if !m.mC {
			m.faceMinX[face] = minX
			m.faceMaxX[face] = maxX
			m.faceMinY[face] = minY
			m.faceMaxY[face] = maxY
			m.faceMinZ[face] = minZ
			m.faceMaxZ[face] = maxZ
		}

		if maxX-minX > m.maxFaceDimension {
			m.maxFaceDimension = maxX - minX
		}
		if maxY-minY > m.maxFaceDimension {
			m.maxFaceDimension = maxY - minY
		}
		if m.maxX < maxX {
			m.maxX = maxX
		}
		if maxZ > m.maxZ {
			m.maxZ = maxZ
		}
		if maxZ-minZ > m.maxFaceDimension {
			m.maxFaceDimension = maxZ - minZ
		}
		if m.maxY < maxY {
			m.maxY = maxY
		}
		if minX < m.minX {
			m.minX = minX
		}
		if minY < m.minY {
			m.minY = minY
		}
		if m.minZ > minZ {
			m.minZ = minZ
		}
	}
}

// light (exported entry per the contract) runs computeNormals — which itself
// chains computeDiffuse (RSModel.java:672-718 -> :712). It is normally driven by
// resetTransformCache; exposed so callers can force a relight after changing the
// diffuse direction.
func (m *Model) light() { m.computeNormals() }

// computeNormals is RSModel.computeNormals (RSModel.java:672-718). The Java `var1`
// arg gates the whole body on var1==14 (only the live :1005 call passes 14), so
// the gate is unconditionally true here. It computes a per-face normal from the
// first three baked verts (halving until it fits +/-8192), normalises to 16.16,
// marks scenePolyNormalShift[face]=-1 (force recompute in Scene.computePolygon),
// then chains computeDiffuse(var1 ^ -85) == computeDiffuse(14 ^ -85) = -91.
func (m *Model) computeNormals() {
	if m.dontComputeDiffuse && m.mC {
		// RSModel.java:675 — `if (!dontComputeDiffuse || !m_c)`: when both hold,
		// the normals arrays were never allocated and the body is skipped.
		return
	}

	for face := 0; face < m.faceHead; face++ {
		idx := m.faceIndices[face]
		xp := m.vertXTransform[idx[0]]
		yp := m.vertYTransform[idx[0]]
		zp := m.vertZTransform[idx[0]]

		x21 := m.vertXTransform[idx[1]] - xp
		y21 := m.vertYTransform[idx[1]] - yp
		z21 := m.vertZTransform[idx[1]] - zp

		x31 := m.vertXTransform[idx[2]] - xp
		y31 := m.vertYTransform[idx[2]] - yp
		z31 := m.vertZTransform[idx[2]] - zp

		xN := z31*y21 - z21*y31
		yN := z21*x31 - x21*z31
		zN := x21*y31 - x31*y21

		for xN > 8192 || yN > 8192 || zN > 8192 || xN < -8192 || yN < -8192 || zN < -8192 {
			yN >>= 1
			zN >>= 1
			xN >>= 1
		}

		mag := int32(math.Sqrt(float64(yN*yN+xN*xN+zN*zN)) * 256.0)
		if mag <= 0 {
			mag = 1
		}

		// NOTE: faceNormZ uses 65535 (not 65536) exactly as RSModel.java:708.
		m.faceNormX[face] = xN * 65536 / mag
		m.faceNormY[face] = yN * 65536 / mag
		m.faceNormZ[face] = zN * 65535 / mag
		m.scenePolyNormalShift[face] = -1
	}

	m.computeDiffuse(-91) // RSModel.java:712: computeDiffuse(14 ^ -85) == -91 (arg unused — var1 is dead in computeDiffuse, matching Java where it only fed a discarded term).
}

// computeDiffuse is RSModel.computeDiffuse (RSModel.java:617-670): flat per-face
// shade for faces whose faceDiffuseLight != m_Vb, and accumulated per-vertex
// gouraud shade for faces whose faceDiffuseLight == m_Vb (TRANSPARENT sentinel).
//
//	diffuseDivide = diffuseParam2 * diffuseMag >> 8
//	flat:  light = (Nx*dirX + Ny*dirY + Nz*dirZ) / diffuseDivide
//	gouraud: per vertex, sum the normals of its m_Vb faces, then
//	         light = (sumN . dir) / (diffuseDivide * faceCount)
//
// var7 (= -16 / ((var1+55)/32)) is computed exactly as Java but is dead (the
// Java code computes it and never uses it; preserved for fidelity, marked _).
func (m *Model) computeDiffuse(var1 int32) {
	if m.dontComputeDiffuse {
		return
	}

	diffuseDivide := (m.diffuseParam2 * m.diffuseMag) >> 8

	for f := 0; f < m.faceHead; f++ {
		if m.faceDiffuseLight[f] != TRANSPARENT {
			m.faceDiffuseLight[f] = (m.faceNormY[f]*m.diffuseDirY +
				m.faceNormX[f]*m.diffuseDirX + m.diffuseDirZ*m.faceNormZ[f]) / diffuseDivide
		}
	}

	tmpXNorm := make([]int32, m.vertHead)
	tmpYNorm := make([]int32, m.vertHead)
	tmpZNorm := make([]int32, m.vertHead)
	faceCount := make([]int32, m.vertHead)
	// Go zero-initialises slices, matching the explicit :637-642 reset loop.

	// RSModel.java:644 computes `-16 / ((var1+55)/32)` and DISCARDS the result.
	// Porting it verbatim divides by zero (both Go and Java truncate (var1+55)/32
	// toward zero, which is 0 for the var1 reached during a terrain build). Since
	// the value is unused, we omit the dead computation rather than panic.

	for i := 0; i < m.faceHead; i++ {
		if m.faceDiffuseLight[i] == TRANSPARENT {
			for fi := 0; fi < m.faceIndexCount[i]; fi++ {
				fVert := m.faceIndices[i][fi]
				tmpXNorm[fVert] += m.faceNormX[i]
				tmpYNorm[fVert] += m.faceNormY[i]
				tmpZNorm[fVert] += m.faceNormZ[i]
				faceCount[fVert]++
			}
		}
	}

	for i := 0; i < m.vertHead; i++ {
		if faceCount[i] > 0 {
			m.vertDiffuseLight[i] = (tmpZNorm[i]*m.diffuseDirZ + tmpXNorm[i]*m.diffuseDirX +
				tmpYNorm[i]*m.diffuseDirY) / (diffuseDivide * faceCount[i])
		}
	}
}

// setDiffuseDir is RSModel.setDiffuseDir (RSModel.java:1157-1174): set the light
// direction, derive its magnitude, and relight. var1 (set m_Yb=71) is dropped —
// the renderer paths pass false, and computeDiffuse does not need a re-bake.
func (m *Model) setDiffuseDir(dirX, dirY, dirZ int32) {
	if m.dontComputeDiffuse {
		return
	}
	m.diffuseDirZ = dirZ
	m.diffuseDirY = dirY
	m.diffuseDirX = dirX
	m.diffuseMag = int32(math.Sqrt(float64(dirZ*dirZ + dirY*dirY + dirX*dirX)))
	m.computeDiffuse(52)
}

// =============================================================================
// Projection — rotate1024
// =============================================================================

// project is RSModel.rotate1024 (RSModel.java:1013-1095). It first bakes the
// transform (resetTransformCache), then frustum-tests the model against the
// package-level frustum* globals (the MiscFunctions.frustum* statics the Scene
// port maintains). If the model is culled, m_dc is cleared and the call returns.
// Otherwise it allocates the projection scratch (clearRotDataAndParams26),
// rotates each baked vertex into camera space (vertXRot/YRot/ZRot) by the
// (rotZ, rotY, rotX) 1024-step angles via trigTable1024, and projects through the
// vpSrc (=rot1024_vp_src=9) shift + perspective divide into vertexParam6 (screen
// X) / vertexParam2 (screen Y). Verts nearer than zTop use the un-divided value.
//
// Argument binding (Scene.endScene, Scene.java:2586-2591):
//
//	yOffset = rot1024_off_y, vParamSrc = rot1024_vp_src, xOffset = rot1024_off_x,
//	zOffset = rot1024_off_z, rotY = cameraProjY, rotZ = cameraProjZ,
//	rotX = cameraProjX, zTop = rot1024_zTop.
//
// The public signature is project(cam, vpSrc, zTop): the offsets/angles come from
// cam, vpSrc is rot1024_vp_src (9), zTop is rot1024_zTop (5).
func (m *Model) project(cam Camera, vpSrc, zTop int32) {
	yOffset := cam.rot1024OffY
	xOffset := cam.rot1024OffX
	zOffset := cam.rot1024OffZ
	rotY := cam.cameraProjY
	rotZ := cam.cameraProjZ
	rotX := cam.cameraProjX

	m.resetTransformCache()

	// Frustum test (RSModel.java:1018-1020). Note the axis swap: model minZ/maxZ
	// vs frustumMinX/MaxX, minX/maxX vs frustumMinY/MaxY, minY/maxY vs
	// frustumNearZ/FarZ — exactly as Java.
	if m.minZ <= frustumMinX && m.maxZ >= frustumMaxX &&
		m.minX <= frustumMinY && m.maxX >= frustumMaxY &&
		m.minY <= frustumNearZ && m.maxY >= frustumFarZ {
		m.mDc = true

		// allocProjectionScratch (re)allocates the rot/param scratch sized to
		// vertHead. This is the projection-arrays half of OpenRSC's
		// clearRotDataAndParams26; the diffuse-light RESEED that OpenRSC's
		// RSModel does here is DELIBERATELY DROPPED — it is an OpenRSC
		// infidelity vs the rev-235 GameModel oracle (see the function doc).
		m.allocProjectionScratch()

		var xyXy, xyYy int32
		var yzZy, yzZz int32
		var xzXz, xzXx int32

		if rotZ != 0 { // RSModel.java:1031-1034
			xyYy = trigTable1024[1024+rotZ]
			xyXy = trigTable1024[rotZ]
		}
		if rotX != 0 { // RSModel.java:1037-1040
			yzZy = trigTable1024[rotX]
			yzZz = trigTable1024[rotX+1024]
		}
		if rotY != 0 { // RSModel.java:1042-1045
			xzXz = trigTable1024[rotY]
			xzXx = trigTable1024[rotY+1024]
		}

		for v := 0; v < m.vertHead; v++ {
			x0 := m.vertXTransform[v] - xOffset
			yO := m.vertYTransform[v] - yOffset
			zO := m.vertZTransform[v] - zOffset

			var tmp int32
			if rotZ != 0 { // RSModel.java:1053-1057
				tmp = (yO*xyXy + xyYy*x0) >> 15
				yO = (yO*xyYy - x0*xyXy) >> 15
				x0 = tmp
			}
			if rotY != 0 { // RSModel.java:1059-1063
				tmp = (xzXx*x0 + zO*xzXz) >> 15
				zO = (xzXx*zO - x0*xzXz) >> 15
				x0 = tmp
			}
			if rotX != 0 { // RSModel.java:1065-1069
				tmp = (yO*yzZz - yzZy*zO) >> 15
				zO = (yzZy*yO + yzZz*zO) >> 15
				yO = tmp
			}

			if zO < zTop { // RSModel.java:1071-1075
				m.vertexParam6[v] = x0 << uint(vpSrc)
			} else {
				m.vertexParam6[v] = (x0 << uint(vpSrc)) / zO
			}

			if zO < zTop { // RSModel.java:1077-1081
				m.vertexParam2[v] = yO << uint(vpSrc)
			} else {
				m.vertexParam2[v] = (yO << uint(vpSrc)) / zO
			}

			m.vertXRot[v] = x0
			m.vertYRot[v] = yO
			m.vertZRot[v] = zO
		}
	} else {
		m.mDc = false
	}
}

// allocProjectionScratch (re)allocates the projection scratch arrays sized to
// the live vertHead. It is the SAFE half of OpenRSC's RSModel.clearRotDataAndParams26
// (RSModel.java:528-543).
//
// OpenRSC's clearRotDataAndParams26 ALSO unconditionally re-seeds the diffuse
// light to setDiffuseLight(40,102,104,108,-20,-89) on every first project (fired
// from rotate1024:1024, byte -103 < 49). That reseed is an OpenRSC RSModel
// INFIDELITY vs the rev-235 J++ client (= the obfuscated JAR oracle, = the deob
// in reference/rsc-client/). The authentic rev-235 GameModel.project()
// (GameModel.java:1225-1281) does NO project-time relight: it calls apply(7972)
// (GameModel.java:1169-1213), which only re-lights on transformState==1 using the
// MODEL'S OWN setLight params (lightAmbience/lightDiffuse/lightDirection set at
// build, e.g. terrain via World.java:1025 terrain.setLight(-50,40,-10,-50,...) =>
// lightAmbience=96, lightDiffuse=384), then projects vertices with no light touch.
//
// Re-seeding clobbered every model's authentic per-build light: terrain's
// diffuseParam1 became 256-102*4 = -152 (vs the correct 96), driving the flat-grass
// gouraud shadeBase negative => clamped to 0 => the brightest ramp entry (green
// 0x8e) instead of the JAR's 0x48. Dropping the reseed restores every model
// (terrain via world.go:520 setDiffuseLightAndColor(-50,-10,-50,40,48,...),
// walls 60/24, roofs 50/50, scenery) to the light it was BUILT with — exactly
// what the deob GameModel does. computeDiffuse already ran at build time and again
// in resetTransformCache->computeNormals (the apply(7972) analog) using those own
// params, so no relight is needed or wanted here.
func (m *Model) allocProjectionScratch() {
	m.vertexParam2 = make([]int32, m.vertHead)
	m.vertXRot = make([]int32, m.vertHead)
	m.vertYRot = make([]int32, m.vertHead)
	m.vertZRot = make([]int32, m.vertHead)
	m.vertexParam6 = make([]int32, m.vertHead)
}

// setDiffuseLight is RSModel.setDiffuseLight (RSModel.java:1176-1197): derive
// diffuseParam1/2 from the two scalar inputs, set the light direction + magnitude,
// and relight. Java arg order: (var1, var2, diffuseDirY, var4, diffuseDirX,
// diffuseDirZ); var4 is a verify arg whose only effect (:1182 set diffuseDirY=-67)
// is immediately overwritten when !dontComputeDiffuse, so it is dropped.
func (m *Model) setDiffuseLight(var1, var2, diffuseDirY, var4, diffuseDirX, diffuseDirZ int32) {
	m.diffuseParam1 = 256 - var2*4
	m.diffuseParam2 = (64-var1)*16 + 128
	_ = var4

	if !m.dontComputeDiffuse {
		m.diffuseDirX = diffuseDirX
		m.diffuseDirZ = diffuseDirZ
		m.diffuseDirY = diffuseDirY
		m.diffuseMag = int32(math.Sqrt(float64(diffuseDirZ*diffuseDirZ +
			diffuseDirY*diffuseDirY + diffuseDirX*diffuseDirX)))
		m.computeDiffuse(-102)
	}
}
