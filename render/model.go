package render

import (
	"math"

	"github.com/gen0cide/westworld/assets"
)

// magic mirrors GameModel.magic (0xbc614e).
const magic = assets.Magic

// GameModel is the renderer's transformable/projectable model (GameModel.java).
// It owns its own vertex/face arrays (so scenery instances can be placed
// independently) and a single queued transform (yaw/translate) like the static
// world models the cradle builds. All math is int32 with deliberate overflow.
type GameModel struct {
	NumVertices int
	NumFaces    int

	VertexX []int32
	VertexY []int32
	VertexZ []int32

	FaceNumVertices []int
	FaceVertices    [][]int
	FaceFillFront   []int32
	FaceFillBack    []int32
	FaceIntensity   []int32
	// faceFixed marks faces whose intensity is a fixed flat value that
	// relight()/light() must NOT recompute from the face normal. Used for
	// water/shore quads: a near-vertical shore face would otherwise be lit to a
	// near-zero (black) intensity, producing the "dark wedge" coastline artifact.
	faceFixed []bool

	// transformed / camera / view space (filled by project)
	tX, tY, tZ       []int32
	camX, camY, camZ []int32
	viewX, viewY     []int32
	vertexIntensity  []int32
	// vertexAmbience is a per-vertex shade offset added into the gouraud
	// intensity (Scene.java:489-491). Terrain seeds it with a deterministic
	// +/-5 speckle (World.java:696); non-terrain models leave it nil (= 0), so
	// walls/scenery are never perturbed.
	vertexAmbience     []int32
	faceNormalX        []int32
	faceNormalY        []int32
	faceNormalZ        []int32
	faceCameraNormalSc []int32

	// queued transform
	baseX, baseY, baseZ                int32
	orientYaw, orientPitch, orientRoll int32
	transformKind                      int

	// lighting params (GameModel ctor defaults)
	lightDirX, lightDirY, lightDirZ          int32
	lightDirMag, lightDiffuse, lightAmbience int32

	// bounds
	x1, y1, z1, x2, y2, z2 int32
	diameter               int32

	visible bool
}

// FromAsset wraps a decoded .ob3 Model as a fresh, transformable GameModel.
func FromAsset(a *assets.Model) *GameModel {
	g := &GameModel{
		NumVertices:     a.NumVertices,
		NumFaces:        a.NumFaces,
		VertexX:         append([]int32(nil), a.VertexX...),
		VertexY:         append([]int32(nil), a.VertexY...),
		VertexZ:         append([]int32(nil), a.VertexZ...),
		FaceNumVertices: a.FaceNumVertices,
		FaceVertices:    a.FaceVertices,
		FaceFillFront:   append([]int32(nil), a.FaceFillFront...),
		FaceFillBack:    append([]int32(nil), a.FaceFillBack...),
		FaceIntensity:   append([]int32(nil), a.FaceIntensity...),
	}
	g.initDefaults()
	return g
}

// NewGameModel builds an empty model (used for the terrain mesh).
func NewGameModel(nV, nF int) *GameModel {
	g := &GameModel{
		NumVertices:     0,
		NumFaces:        0,
		VertexX:         make([]int32, nV),
		VertexY:         make([]int32, nV),
		VertexZ:         make([]int32, nV),
		FaceNumVertices: make([]int, nF),
		FaceVertices:    make([][]int, nF),
		FaceFillFront:   make([]int32, nF),
		FaceFillBack:    make([]int32, nF),
		FaceIntensity:   make([]int32, nF),
		faceFixed:       make([]bool, nF),
	}
	g.initDefaults()
	return g
}

// AddFixedFace appends a flat-coloured face whose intensity is FIXED (never
// recomputed by relight). Use for water/shore quads so a near-vertical face
// cannot be lit to black. intensity is the constant flat shade index.
func (g *GameModel) AddFixedFace(verts []int, fillFront, fillBack, intensity int32) int {
	i := g.AddFace(verts, fillFront, fillBack, intensity)
	if g.faceFixed == nil {
		g.faceFixed = make([]bool, len(g.FaceIntensity))
	}
	g.faceFixed[i] = true
	return i
}

func (g *GameModel) initDefaults() {
	g.lightDirX = 180
	g.lightDirY = 155
	g.lightDirZ = 95
	g.lightDirMag = 256
	g.lightDiffuse = 512
	g.lightAmbience = 32
	g.diameter = magic
}

// AddVertex appends a vertex, returns its index (World mesh build helper).
func (g *GameModel) AddVertex(x, y, z int32) int {
	i := g.NumVertices
	g.VertexX[i] = x
	g.VertexY[i] = y
	g.VertexZ[i] = z
	g.NumVertices++
	return i
}

// SetVertexAmbience records a per-vertex shade offset (GameModel.setVertexAmbience).
// Lazily allocates the slice; only the terrain mesh uses it.
func (g *GameModel) SetVertexAmbience(v int, ambience int32) {
	if g.vertexAmbience == nil {
		g.vertexAmbience = make([]int32, len(g.VertexX))
	}
	g.vertexAmbience[v] = ambience
}

// AddFace appends a face with the given vertex indices and front/back fill.
func (g *GameModel) AddFace(verts []int, fillFront, fillBack, intensity int32) int {
	i := g.NumFaces
	g.FaceNumVertices[i] = len(verts)
	g.FaceVertices[i] = verts
	g.FaceFillFront[i] = fillFront
	g.FaceFillBack[i] = fillBack
	g.FaceIntensity[i] = intensity
	g.NumFaces++
	return i
}

// SetLight sets the directional light, ported from GameModel.setLight(int
// ambience, int diffuse, int x, int y, int z) (GameModel.java:510). The args
// are the RAW values the World code passes (e.g. setLight(48, 48, -50,-10,-50));
// they are remapped exactly as the client does:
//
//	lightAmbience = 256 - ambience*4
//	lightDiffuse  = (64 - diffuse)*16 + 128
//
// The MVP stored the raw args directly, which left lightDiffuse far too small
// so the lighting divisor blew up and every face clamped to intensity 0 (pure
// black) or 255. This remap is what makes faces shade correctly.
func (g *GameModel) SetLight(ambience, diffuse, dx, dy, dz int32) {
	g.lightAmbience = 256 - ambience*4
	g.lightDiffuse = (64-diffuse)*16 + 128
	g.lightDirX = dx
	g.lightDirY = dy
	g.lightDirZ = dz
	mag := int32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	if mag == 0 {
		mag = 1
	}
	g.lightDirMag = mag
	// (lighting is (re)computed in project -> apply -> relight -> light, which
	// runs after the transform is baked, so we don't relight here.)
}

// Rotate queues a yaw/pitch/roll rotation (256-step space).
func (g *GameModel) Rotate(yaw, pitch, roll int32) {
	g.orientPitch = (g.orientPitch + pitch) & 0xff
	g.orientYaw = (g.orientYaw + yaw) & 0xff
	g.orientRoll = (g.orientRoll + roll) & 0xff
	g.computeTransformKind()
}

// Translate queues a translation.
func (g *GameModel) Translate(dx, dy, dz int32) {
	g.baseX += dx
	g.baseY += dy
	g.baseZ += dz
	g.computeTransformKind()
}

func (g *GameModel) computeTransformKind() {
	if g.orientYaw != 0 || g.orientPitch != 0 || g.orientRoll != 0 {
		if g.transformKind < 2 {
			g.transformKind = 2
		}
	} else if g.baseX != 0 || g.baseY != 0 || g.baseZ != 0 {
		if g.transformKind < 1 {
			g.transformKind = 1
		}
	}
}

func (g *GameModel) ensureScratch() {
	if g.tX == nil {
		g.tX = make([]int32, g.NumVertices)
		g.tY = make([]int32, g.NumVertices)
		g.tZ = make([]int32, g.NumVertices)
		g.camX = make([]int32, g.NumVertices)
		g.camY = make([]int32, g.NumVertices)
		g.camZ = make([]int32, g.NumVertices)
		g.viewX = make([]int32, g.NumVertices)
		g.viewY = make([]int32, g.NumVertices)
		g.vertexIntensity = make([]int32, g.NumVertices)
		g.faceNormalX = make([]int32, g.NumFaces)
		g.faceNormalY = make([]int32, g.NumFaces)
		g.faceNormalZ = make([]int32, g.NumFaces)
		g.faceCameraNormalSc = make([]int32, g.NumFaces)
	}
}

// apply bakes the queued transform into transformed-space and relights.
func (g *GameModel) apply() {
	g.ensureScratch()
	for v := 0; v < g.NumVertices; v++ {
		g.tX[v] = g.VertexX[v]
		g.tY[v] = g.VertexY[v]
		g.tZ[v] = g.VertexZ[v]
	}
	if g.transformKind >= 2 {
		g.applyRotation(g.orientYaw, g.orientPitch, g.orientRoll)
	}
	if g.transformKind >= 1 {
		for v := 0; v < g.NumVertices; v++ {
			g.tX[v] += g.baseX
			g.tY[v] += g.baseY
			g.tZ[v] += g.baseZ
		}
	}
	g.computeBounds()
	g.relight()
}

func (g *GameModel) applyRotation(yaw, pitch, roll int32) {
	for v := 0; v < g.NumVertices; v++ {
		if pitch != 0 {
			sin := sine9[pitch]
			cos := sine9[pitch+256]
			x := (g.tY[v]*sin + g.tX[v]*cos) >> 15
			g.tY[v] = (g.tY[v]*cos - g.tX[v]*sin) >> 15
			g.tX[v] = x
		}
		if yaw != 0 {
			sin := sine9[yaw]
			cos := sine9[yaw+256]
			y := (g.tY[v]*cos - g.tZ[v]*sin) >> 15
			g.tZ[v] = (g.tY[v]*sin + g.tZ[v]*cos) >> 15
			g.tY[v] = y
		}
		if roll != 0 {
			sin := sine9[roll]
			cos := sine9[roll+256]
			x := (g.tZ[v]*sin + g.tX[v]*cos) >> 15
			g.tZ[v] = (g.tZ[v]*cos - g.tX[v]*sin) >> 15
			g.tX[v] = x
		}
	}
}

func (g *GameModel) computeBounds() {
	g.x1 = 999999
	g.y1 = 999999
	g.z1 = 999999
	g.x2 = -999999
	g.y2 = -999999
	g.z2 = -999999
	for v := 0; v < g.NumVertices; v++ {
		if g.tX[v] < g.x1 {
			g.x1 = g.tX[v]
		}
		if g.tX[v] > g.x2 {
			g.x2 = g.tX[v]
		}
		if g.tY[v] < g.y1 {
			g.y1 = g.tY[v]
		}
		if g.tY[v] > g.y2 {
			g.y2 = g.tY[v]
		}
		if g.tZ[v] < g.z1 {
			g.z1 = g.tZ[v]
		}
		if g.tZ[v] > g.z2 {
			g.z2 = g.tZ[v]
		}
	}
}

// relight recomputes face normals and per-face/per-vertex intensities
// (GameModel.relight + light, GameModel.java:704-773). The cross products
// intentionally overflow int32; do NOT widen.
func (g *GameModel) relight() {
	for face := 0; face < g.NumFaces; face++ {
		verts := g.FaceVertices[face]
		if len(verts) < 3 {
			g.faceNormalX[face] = 0
			g.faceNormalY[face] = 0
			g.faceNormalZ[face] = 0
			g.faceCameraNormalSc[face] = -1
			continue
		}
		aX := g.tX[verts[0]]
		aY := g.tY[verts[0]]
		aZ := g.tZ[verts[0]]
		bX := g.tX[verts[1]] - aX
		bY := g.tY[verts[1]] - aY
		bZ := g.tZ[verts[1]] - aZ
		cX := g.tX[verts[2]] - aX
		cY := g.tY[verts[2]] - aY
		cZ := g.tZ[verts[2]] - aZ
		normX := bY*cZ - cY*bZ
		normY := bZ*cX - cZ*bX
		normZ := bX*cY - cX*bY
		for normX > 8192 || normY > 8192 || normZ > 8192 || normX < -8192 || normY < -8192 || normZ < -8192 {
			normX >>= 1
			normY >>= 1
			normZ >>= 1
		}
		normMag := int32(256.0 * math.Sqrt(float64(normX*normX+normY*normY+normZ*normZ)))
		if normMag <= 0 {
			normMag = 1
		}
		g.faceNormalX[face] = (normX * 0x10000) / normMag
		g.faceNormalY[face] = (normY * 0x10000) / normMag
		g.faceNormalZ[face] = (normZ * 65535) / normMag
		g.faceCameraNormalSc[face] = -1
	}
	g.light()
}

func (g *GameModel) light() {
	divisor := g.lightDiffuse * g.lightDirMag >> 8
	if divisor == 0 {
		divisor = 1
	}
	for face := 0; face < g.NumFaces; face++ {
		if face < len(g.faceFixed) && g.faceFixed[face] {
			continue // fixed-intensity face (water/shore): keep its flat shade
		}
		if g.FaceIntensity[face] != magic {
			g.FaceIntensity[face] = (g.faceNormalX[face]*g.lightDirX +
				g.faceNormalY[face]*g.lightDirY + g.faceNormalZ[face]*g.lightDirZ) / divisor
		}
	}
	normalX := make([]int32, g.NumVertices)
	normalY := make([]int32, g.NumVertices)
	normalZ := make([]int32, g.NumVertices)
	normalMag := make([]int32, g.NumVertices)
	for face := 0; face < g.NumFaces; face++ {
		if g.FaceIntensity[face] == magic {
			for v := 0; v < g.FaceNumVertices[face]; v++ {
				k1 := g.FaceVertices[face][v]
				normalX[k1] += g.faceNormalX[face]
				normalY[k1] += g.faceNormalY[face]
				normalZ[k1] += g.faceNormalZ[face]
				normalMag[k1]++
			}
		}
	}
	for v := 0; v < g.NumVertices; v++ {
		if normalMag[v] > 0 {
			g.vertexIntensity[v] = (normalX[v]*g.lightDirX +
				normalY[v]*g.lightDirY + normalZ[v]*g.lightDirZ) / (divisor * normalMag[v])
		}
	}
}

// project transforms transformed-space into camera + view space
// (GameModel.project, :809). viewDist=9 by default; clipNear=5.
func (g *GameModel) project(cam Camera, viewDist, clipNear int32) {
	g.apply()
	g.visible = true
	var yawSin, yawCos, pitchSin, pitchCos, rollSin, rollCos int32
	if cam.CameraYaw != 0 {
		yawSin = sine11[cam.CameraYaw]
		yawCos = sine11[cam.CameraYaw+1024]
	}
	if cam.CameraRoll != 0 {
		rollSin = sine11[cam.CameraRoll]
		rollCos = sine11[cam.CameraRoll+1024]
	}
	if cam.CameraPitch != 0 {
		pitchSin = sine11[cam.CameraPitch]
		pitchCos = sine11[cam.CameraPitch+1024]
	}
	for v := 0; v < g.NumVertices; v++ {
		x := g.tX[v] - cam.CameraX
		y := g.tY[v] - cam.CameraY
		z := g.tZ[v] - cam.CameraZ
		if cam.CameraYaw != 0 {
			X := (y*yawSin + x*yawCos) >> 15
			y = (y*yawCos - x*yawSin) >> 15
			x = X
		}
		if cam.CameraRoll != 0 {
			X := (z*rollSin + x*rollCos) >> 15
			z = (z*rollCos - x*rollSin) >> 15
			x = X
		}
		if cam.CameraPitch != 0 {
			Y := (y*pitchCos - z*pitchSin) >> 15
			z = (y*pitchSin + z*pitchCos) >> 15
			y = Y
		}
		if z >= clipNear {
			g.viewX[v] = (x << uint(viewDist)) / z
			g.viewY[v] = (y << uint(viewDist)) / z
		} else {
			g.viewX[v] = x << uint(viewDist)
			g.viewY[v] = y << uint(viewDist)
		}
		g.camX[v] = x
		g.camY[v] = y
		g.camZ[v] = z
	}
	// reset the per-face camera-normal scale; recomputed lazily per face.
	for f := 0; f < g.NumFaces; f++ {
		g.faceCameraNormalSc[f] = -1
	}
}

// cameraNormalSign ports Scene.method293's anInt365 (Scene.java:2326): the dot
// product of vertex0's CAMERA-space position with the CAMERA-space face normal
// (cross product of the first two camera-space edges). Its sign decides
// front/back facing. A per-face scale is computed once (the >>1 normalisation
// loop) so the cross product never overflows; subsequent calls reuse it. This
// replaces the MVP's inverted 2D-winding test, which produced solid-black faces.
func (g *GameModel) cameraNormalSign(face int) int32 {
	verts := g.FaceVertices[face]
	j1 := g.camX[verts[0]]
	k1 := g.camY[verts[0]]
	l1 := g.camZ[verts[0]]
	i2 := g.camX[verts[1]] - j1
	j2 := g.camY[verts[1]] - k1
	k2 := g.camZ[verts[1]] - l1
	l2 := g.camX[verts[2]] - j1
	i3 := g.camY[verts[2]] - k1
	j3 := g.camZ[verts[2]] - l1
	k3 := j2*j3 - i3*k2
	l3 := k2*l2 - j3*i2
	i4 := i2*i3 - l2*j2
	l := g.faceCameraNormalSc[face]
	if l == -1 {
		l = 0
		for k3 > 25000 || l3 > 25000 || i4 > 25000 || k3 < -25000 || l3 < -25000 || i4 < -25000 {
			l++
			k3 >>= 1
			l3 >>= 1
			i4 >>= 1
		}
		g.faceCameraNormalSc[face] = l
	} else {
		k3 >>= uint(l)
		l3 >>= uint(l)
		i4 >>= uint(l)
	}
	return j1*k3 + k1*l3 + l1*i4
}
