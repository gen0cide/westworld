package render

import "sort"

// Frustum / view constants (Scene.java init + setMidpoints). viewDist=9 for the
// MVP (the plan's spec); clipNear=5; clipFar=2400ish.
const (
	viewDist = 9
	clipNear = 5
	clipFar  = 2400
)

// Scene holds the camera and the list of GameModels to render. RenderTo
// projects every model, collects visible faces, painter-sorts by average
// camera-Z, and rasterizes back-to-front (Scene.endscene, :345).
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
}

// RenderTo projects + rasterizes all models into the surface. cx/cy are the
// pixel-space screen centre (anInt399/anInt400).
func (s *Scene) RenderTo(surf *Surface, cx, cy int) {
	// NOTE the axis swap at the call site: Scene calls project with
	// (cameraPitch, cameraYaw, cameraRoll) bound to project's
	// (cameraPitch, cameraRoll, cameraYaw). We replicate by swapping yaw/roll.
	cam := s.Cam
	cam.CameraYaw, cam.CameraRoll = s.Cam.CameraRoll, s.Cam.CameraYaw

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
			// front/back via 2D signed area of the projected face
			// (winding). Positive area -> one facing, negative the other.
			a := m.viewX[verts[0]]
			b := m.viewY[verts[0]]
			c := m.viewX[verts[1]]
			d := m.viewY[verts[1]]
			e := m.viewX[verts[2]]
			ff := m.viewY[verts[2]]
			area := (c-a)*(ff-b) - (e-a)*(d-b)
			front := area < 0
			var fill int32
			if front {
				fill = m.FaceFillFront[f]
			} else {
				fill = m.FaceFillBack[f]
			}
			if fill == magic {
				continue
			}
			var sum int32
			for _, v := range verts {
				sum += m.camZ[v]
			}
			depth := sum / int32(len(verts))
			faces = append(faces, collectedFace{
				model: m, face: f, depth: depth, fill: fill, front: front,
			})
		}
	}

	// painter's sort: far (large depth) first, near last.
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})

	for _, cf := range faces {
		s.rasterFace(surf, cx, cy, cf)
	}
}

func (s *Scene) rasterFace(surf *Surface, cx, cy int, cf collectedFace) {
	m := cf.model
	verts := m.FaceVertices[cf.face]

	// Only render flat-colour fills (fill < 0). Texture ids (>= 0) become a
	// mid-grey placeholder for the MVP.
	var ramp [256]int32
	if cf.fill < 0 {
		ramp = gouraudRamp(cf.fill)
	} else {
		// grey placeholder ramp
		for i := int32(0); i < 256; i++ {
			v := (i * i) / 256 / 2
			ramp[255-i] = (v << 16) | (v << 8) | v
		}
	}

	amb := m.lightAmbience
	gouraud := m.FaceIntensity[cf.face] == magic

	// build per-vertex screen coords + intensity index
	type pv struct{ x, y, in int }
	pvs := make([]pv, 0, len(verts))
	for _, v := range verts {
		if m.camZ[v] < clipNear {
			// crude near-clip: skip whole face if any vertex behind near plane
			return
		}
		sx := int(m.viewX[v]) + cx
		sy := int(m.viewY[v]) + cy
		var in int32
		if gouraud {
			if cf.front {
				in = amb + m.vertexIntensity[v]
			} else {
				in = amb - m.vertexIntensity[v]
			}
		} else {
			if cf.front {
				in = amb - m.FaceIntensity[cf.face]
			} else {
				in = amb + m.FaceIntensity[cf.face]
			}
		}
		if in < 0 {
			in = 0
		}
		if in > 255 {
			in = 255
		}
		pvs = append(pvs, pv{sx, sy, int(in)})
	}

	// triangle-fan the polygon
	for i := 1; i+1 < len(pvs); i++ {
		fillTriangleGouraud(surf,
			pvs[0].x, pvs[0].y, pvs[0].in,
			pvs[i].x, pvs[i].y, pvs[i].in,
			pvs[i+1].x, pvs[i+1].y, pvs[i+1].in,
			&ramp)
	}
}
