package render

import "sort"

// Frustum / view constants (Scene.java init + setMidpoints). viewDist=9 for the
// MVP (the plan's spec); clipNear=5; clipFar matches the client 3d far plane
// (mudclient sets scene.clipFar_3d/_2d = 2400 per frame, drawScene
// mudclient.java:3748-3751; also at init :3215-3218).
//
// Distance fog is NOT the client's index-darkening term (that fades to black);
// it is a depth-based blend toward skyColour done in the span fillers — see
// fogStart in render.go and fogBlend in rasterize.go.
const (
	viewDist = 9
	clipNear = 5
	clipFar  = 2400
)

// Scene holds the camera and the list of GameModels to render. RenderTo
// projects every model, collects visible faces (selecting front/back fill via
// the real camera-space normal sign), painter-sorts by camera depth, and
// rasterizes back-to-front with the ported Scene scanline filler.
type Scene struct {
	Cam    Camera
	Models []*GameModel
}

func (s *Scene) Add(m *GameModel) { s.Models = append(s.Models, m) }

type collectedFace struct {
	model *GameModel
	face  int
	depth int32
	fill  int32
	front bool // anInt365 < 0

	// camera-space face plane (Scene.method293): scaled normal (anInt362/363/364),
	// the magnitude (faceCameraNormalMagnitude / anInt366 used by the plane tests),
	// and the signed offset against vertex0 (anInt365).
	nx, ny, nz int32
	normMag    int32
	dotV0      int32

	// screen-space AABB (anInt353/355 = min/max view-X, anInt354/356 = min/max
	// view-Y) and camera-Z extent (anInt357/358 = min/max camZ). Used by the
	// method277 window cull, method295's AABB / camZ-range early-outs.
	minVX, maxVX int32
	minVY, maxVY int32
	minCZ, maxCZ int32

	// method277/method278 reinsertion bookkeeping (Polygon.anInt368/369 +
	// aBoolean367).
	anInt368 int32
	anInt369 int32
	visited  bool
}

// RenderTo projects + rasterizes all models into the surface. cx/cy are the
// pixel-space screen centre (anInt399/anInt400).
func (s *Scene) RenderTo(surf *Surface, cx, cy int) {
	// Axis swap at the call site: Scene.endscene calls project with
	// (cameraPitch, cameraYaw, cameraRoll) bound to project's
	// (cameraPitch, cameraRoll, cameraYaw). Replicate by swapping yaw/roll.
	cam := s.Cam
	cam.CameraYaw, cam.CameraRoll = s.Cam.CameraRoll, s.Cam.CameraYaw

	// horizontal/vertical half-extents for the frustum bounds cull (anInt397/398)
	anInt397 := int32(cx)
	anInt398 := int32(cy)

	var faces []collectedFace
	for _, m := range s.Models {
		m.project(cam, viewDist, clipNear)
		if !m.visible {
			continue
		}
		for f := 0; f < m.NumFaces; f++ {
			verts := m.FaceVertices[f]
			if len(verts) < 3 {
				continue
			}
			// require at least one vertex within the near/far z band
			ok := false
			for _, v := range verts {
				z := m.camZ[v]
				if z > clipNear && z < clipFar {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
			// Frustum X/Y bounds cull (Scene.endscene :366-390): the projected
			// face must span the horizontal view (some vertex right of the left
			// edge AND some left of the right edge) and likewise vertically.
			// Without this, far off-screen terrain quads keep their runaway
			// view-X (millions of units) and rasterize as full-width streaks.
			l1 := 0
			for _, v := range verts {
				vx := m.viewX[v]
				if vx > -anInt397 {
					l1 |= 1
				}
				if vx < anInt397 {
					l1 |= 2
				}
				if l1 == 3 {
					break
				}
			}
			if l1 != 3 {
				continue
			}
			i2 := 0
			for _, v := range verts {
				vy := m.viewY[v]
				if vy > -anInt398 {
					i2 |= 1
				}
				if vy < anInt398 {
					i2 |= 2
				}
				if i2 == 3 {
					break
				}
			}
			if i2 != 3 {
				continue
			}
			// Front/back facing via the real camera-space normal sign
			// (Scene.method293 anInt365 = vertex0 . cameraNormal). The MVP's
			// 2D projected-winding test was inverted and produced black faces.
			front := m.cameraNormalSign(f) < 0
			var fill int32
			if front {
				fill = m.FaceFillFront[f]
			} else {
				fill = m.FaceFillBack[f]
			}
			if fill == magic {
				continue
			}
			// Camera-Z sum (for the average-depth seed) and the screen/camZ AABB
			// the overlap pass needs. The AABB tracking mirrors Scene.method293
			// (Scene.java:2364-2393): seed min=max from vertex0, then fold in the
			// rest (the client's `if > max else if < min` form).
			v0 := verts[0]
			minCZ, maxCZ := m.camZ[v0], m.camZ[v0]
			minVX, maxVX := m.viewX[v0], m.viewX[v0]
			minVY, maxVY := m.viewY[v0], m.viewY[v0]
			sum := m.camZ[v0]
			for k := 1; k < len(verts); k++ {
				v := verts[k]
				cz := m.camZ[v]
				if cz > maxCZ {
					maxCZ = cz
				} else if cz < minCZ {
					minCZ = cz
				}
				vx := m.viewX[v]
				if vx > maxVX {
					maxVX = vx
				} else if vx < minVX {
					minVX = vx
				}
				vy := m.viewY[v]
				if vy > maxVY {
					maxVY = vy
				} else if vy < minVY {
					minVY = vy
				}
				sum += cz
			}
			depth := sum/int32(len(verts)) + m.Depth
			nx, ny, nz, dotV0, normMag := m.cameraNormal(f)
			faces = append(faces, collectedFace{
				model: m, face: f, depth: depth, fill: fill, front: front,
				nx: nx, ny: ny, nz: nz, normMag: normMag, dotV0: dotV0,
				minVX: minVX, maxVX: maxVX, minVY: minVY, maxVY: maxVY,
				minCZ: minCZ, maxCZ: maxCZ,
			})
		}
	}

	// painter's sort: far (large depth) first, near last. This is the faithful
	// coarse seed (Scene.endscene:444 qsort) BEFORE method277.
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})

	// Authentic overlap-resolution pass (Scene.method277): reorder
	// mutually-overlapping faces by true 3D plane tests so wall corners and
	// perpendicular junctions paint in the correct order.
	resolveOverlaps(faces)

	r := newRaster(surf, cx, cy)
	// scratch arrays for the (possibly clip-split) projected polygon
	var ax, ay, ai [16]int32
	// scratch arrays for the UNCLIPPED camera-space verts of the original face,
	// in original vertex order (method282's planar projection reads [0],[1],[last]).
	var cX, cY, cZ [16]int32
	for _, cf := range faces {
		s.rasterFace(r, cf, ax[:], ay[:], ai[:], cX[:], cY[:], cZ[:])
	}
}

// rasterFace clips one face against the near plane (vertex-splitting, exactly
// like the method281 caller in Scene.endscene :482-528), builds the per-vertex
// view-X/view-Y/intensity arrays, then runs method281 + method291.
func (s *Scene) rasterFace(r *raster, cf collectedFace, ax, ay, ai, cX, cY, cZ []int32) {
	m := cf.model
	verts := m.FaceVertices[cf.face]
	l10 := len(verts)

	// Per-pixel depth the span fillers write (FIX A): the face's average
	// camera-Z, computed in RenderTo (cf.depth). Coarse per-face depth is
	// enough for the 2D sprite-occlusion pass — RSC face spans are small.
	r.faceDepth = cf.depth
	// Per-face distance-fog factor: distant faces blend toward skyColour in the
	// span fillers (fogBlend) so terrain/water/scenery fades into the blue sky.
	r.setFaceFog(cf.depth)

	// flat per-face intensity base (j10), used when the face is NOT gouraud.
	amb := m.lightAmbience
	gouraud := m.FaceIntensity[cf.face] == magic
	fixed := cf.face < len(m.faceFixed) && m.faceFixed[cf.face]
	var j10 int32
	if fixed {
		// Water/shore: the stored intensity IS the final flat shade index,
		// independent of front/back facing — so a near-vertical shore face
		// can never be lit to a black wedge.
		j10 = m.FaceIntensity[cf.face]
	} else if !gouraud {
		if cf.front {
			j10 = amb - m.FaceIntensity[cf.face]
		} else {
			j10 = amb + m.FaceIntensity[cf.face]
		}
	}

	// Texture routing. A negative fill is an encoded flat 5:5:5 colour. A
	// non-negative fill is a TEXTURE id: prefer the REAL perspective-correct
	// textured span fill (method282) when a shaded texel buffer was built for
	// that id; otherwise fall back EXACTLY to the prior flat sampled-colour path
	// (gouraudRamp(textureFill(id))), so the worst case is never worse than the
	// committed renderer.
	var texBuf *textureBuf
	textured := false
	if cf.fill >= 0 {
		if b := textureBuffer(cf.fill); b != nil {
			texBuf = b
			textured = true
		}
	}

	// Shade ramp for the FLAT path (flat colours, or the textured fallback).
	var ramp [256]int32
	if cf.fill < 0 {
		ramp = gouraudRamp(cf.fill)
	} else {
		ramp = gouraudRamp(textureFill(cf.fill))
	}

	// k8 = count of emitted (clip) vertices.
	k8 := 0
	for k11 := 0; k11 < l10; k11++ {
		k2 := verts[k11]
		// UNCLIPPED camera-space verts in ORIGINAL order (method282 reads
		// [0],[1],[l10-1]); captured for ALL faces so the textured path has them.
		cX[k11] = m.camX[k2]
		cY[k11] = m.camY[k2]
		cZ[k11] = m.camZ[k2]
		// per-vertex intensity for gouraud faces (vertexAmbience is 0 here)
		jj := j10
		if gouraud {
			if cf.front {
				jj = amb - m.vertexIntensity[k2]
			} else {
				jj = amb + m.vertexIntensity[k2]
			}
			// authentic terrain speckle (Scene.java:489-491); nil for
			// walls/scenery so they're never perturbed.
			if m.vertexAmbience != nil {
				jj += m.vertexAmbience[k2]
			}
		}
		if m.camZ[k2] >= clipNear {
			ax[k8] = m.viewX[k2]
			ay[k8] = m.viewY[k2]
			// NOTE: distance fog is NOT applied to the shade INDEX here. The
			// old index-darkening term (jj += (camZ-fogZDistance)/fogZFalloff)
			// pushed distant faces toward ramp-black, so distant terrain faded
			// to BLACK voids against the sky. Fog is now done in the span fillers
			// (rasterize.go fogBlend): the rendered pixel is blended toward
			// skyColour by depth, so distant terrain fades into the blue sky
			// instead of darkening to black.
			ai[k8] = jj
			k8++
		} else {
			// near-clip vertex SPLIT: emit the intersection with each
			// neighbouring in-front edge (Scene.java:499-527).
			var prev int
			if k11 == 0 {
				prev = verts[l10-1]
			} else {
				prev = verts[k11-1]
			}
			if m.camZ[prev] >= clipNear {
				k7 := m.camZ[k2] - m.camZ[prev]
				i5 := m.camX[k2] - ((m.camX[k2]-m.camX[prev])*(m.camZ[k2]-clipNear))/k7
				j6 := m.camY[k2] - ((m.camY[k2]-m.camY[prev])*(m.camZ[k2]-clipNear))/k7
				ax[k8] = (i5 << viewDist) / clipNear
				ay[k8] = (j6 << viewDist) / clipNear
				ai[k8] = jj
				k8++
			}
			var next int
			if k11 == l10-1 {
				next = verts[0]
			} else {
				next = verts[k11+1]
			}
			if m.camZ[next] >= clipNear {
				l7 := m.camZ[k2] - m.camZ[next]
				j5 := m.camX[k2] - ((m.camX[k2]-m.camX[next])*(m.camZ[k2]-clipNear))/l7
				k6 := m.camY[k2] - ((m.camY[k2]-m.camY[next])*(m.camZ[k2]-clipNear))/l7
				ax[k8] = (j5 << viewDist) / clipNear
				ay[k8] = (k6 << viewDist) / clipNear
				ai[k8] = jj
				k8++
			}
		}
	}
	if k8 < 3 {
		return
	}

	// clamp intensities to [0,255] (Scene.java:530-534), then apply the textured
	// intensity shift: <<9 for a 128px (class1) texture, <<6 for a 64px (class0)
	// texture (Scene.java:535-539). Flat faces keep the unshifted intensity for
	// method291's ramp lookup.
	for i := 0; i < k8; i++ {
		if ai[i] < 0 {
			ai[i] = 0
		} else if ai[i] > 255 {
			ai[i] = 255
		}
		if textured {
			if texBuf.size == 128 {
				ai[i] <<= 9
			} else {
				ai[i] <<= 6
			}
		}
	}

	r.method281(int32(k8), ax[:k8], ay[:k8], ai[:k8])
	if r.anInt440 <= r.anInt439 {
		return
	}
	if textured {
		r.method282(texBuf, cX[:l10], cY[:l10], cZ[:l10], l10)
	} else {
		r.method291(&ramp)
	}
}
