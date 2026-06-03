package orsc

// entityspec.go — the SHARED static-entity spec for the NPC/player 2D-sprite
// parity build (docs/build/NPC_SPRITE_PARITY_PLAN.md tool #3). It is the one
// canonical description of "which entity, where, how big" that ALL THREE legs
// (orsc / DEOB / rev-235 JAR) read so the spec is identical with no env drift.
//
// Two sources feed the spec, deliberately redundant so the legs can be driven
// either way and cross-checked:
//   1. the rscdump Entities[] fixture field (orsc reads it natively via
//      addViewEntities; the Java legs parse the same JSON), and
//   2. the RSC_MESH_NPC / RSC_MESH_PLAYER env gate (mirroring RSC_MESH_REALDEFS),
//      so a parity run can place an entity WITHOUT editing the fixture.
//
// PHASE 0 (placement sanity): the spec drives a SOLID DEBUG-COLOUR billboard
// (NOT the real composited sprite) at the host centre tile, so a PNG diff of the
// screen rect isolates the PROJECTION (billboard world-size -> screen w/h/x/y)
// from sprite decode/composite. The first entity is the rat (content0 serverId
// 19), whose billboard is 346x136 (entityIndexTableC[19]/legacyMaskTable[19] ==
// OpenRSC NpcDefs id 19 camera1/camera2). Later phases reuse the SAME spec
// placement and swap the solid fill for the real sprite stack.

import (
	"os"
	"strconv"
	"strings"

	"github.com/gen0cide/westworld/pathfind"
)

// debugBillboardColour is the SOLID Phase-0 placement-sanity fill (0x00RRGGBB):
// a fully-saturated cyan, distinct from every terrain/scenery/door colour so the
// rect is trivially isolable in a diff. ALL THREE legs fill the projected
// billboard rect with EXACTLY this value; Phase 1 replaces it with the decoded
// sprite. (Not the 0xFF00FF transparent sentinel — that is the sprite-pipeline
// transparent key; a debug fill must be opaque everywhere.)
const debugBillboardColour = int32(0x0000FFFF)

// ratBillboardW / ratBillboardH are the rat's (content0 serverId 19) world-space
// billboard size, fixed by the plan (entityIndexTableC[19]/legacyMaskTable[19]).
// Phase 0 hardcodes them so placement sanity has zero dependency on def loading
// (Phase 1 loads real per-id sizes). The Java legs hardcode the SAME 346x136.
const (
	ratServerID    = 19
	ratBillboardW  = 346
	ratBillboardH  = 136
)

// staticEntitySpec is one entity to place: its host-window-LOCAL tile, its
// world-space billboard size, and (Phase 0) the solid debug fill. tileLocalX/Y
// are window-local (0..worldWindowTiles); windowCentreTile (=48) is the host
// centre tile, where the camera sits.
type staticEntitySpec struct {
	tileLocalX int
	tileLocalY int
	billboardW int
	billboardH int
	debug      bool  // Phase 0: fill the projected rect with debugColour, not a sprite
	debugColour int32
	serverID   int
	dir        int
	step       int
}

// phase0EntitySpec returns the Phase-0 debug billboard spec when the entity gate
// is engaged, or nil when it is not (so terrain/scenery fixtures render
// unchanged). The gate is engaged by either:
//   - RSC_MESH_NPC=<serverId>[:<dir>:<step>]  (env, mirrors RSC_MESH_REALDEFS), or
//   - a fixture Entities[] entry of kind "npc" (the rscdump field),
// with the env gate taking precedence (it can override a fixture's entity).
// PHASE 0 places exactly ONE billboard, at the host centre tile (grid-local 48),
// of the rat's 346x136 size, filled solid — proving placement before any decode.
func phase0EntitySpec(hasFixtureNPC bool) *staticEntitySpec {
	gate := strings.TrimSpace(os.Getenv("RSC_MESH_NPC"))
	if gate == "" && !hasFixtureNPC {
		return nil
	}
	spec := &staticEntitySpec{
		tileLocalX:  windowCentreTile,
		tileLocalY:  windowCentreTile,
		billboardW:  ratBillboardW,
		billboardH:  ratBillboardH,
		debug:       true,
		debugColour: debugBillboardColour,
		serverID:    ratServerID,
	}
	if gate != "" {
		parts := strings.Split(gate, ":")
		if v, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			spec.serverID = v
		}
		if len(parts) > 1 {
			if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				spec.dir = v
			}
		}
		if len(parts) > 2 {
			if v, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
				spec.step = v
			}
		}
	}
	return spec
}

// entityGateEngaged reports whether the entity layer is active for this render
// (env gate set OR a fixture npc entity present). When false the dump path must
// render terrain/scenery EXACTLY as before (the Phase-0 gate invariant).
func entityGateEngaged(hasFixtureNPC bool) bool {
	return strings.TrimSpace(os.Getenv("RSC_MESH_NPC")) != "" || hasFixtureNPC
}

// debugBillboard is one registered Phase-0 placement-sanity billboard: the mT
// face index Render projected it as, plus its world-space size + solid fill. The
// face index lets fillDebugBillboards read back the PROJECTED vertex after Render.
type debugBillboard struct {
	faceIndex  int
	billboardW int32
	billboardH int32
	colour     int32
}

// placeStaticEntity registers the spec's billboard with the Scene's mT model at
// its host-window-local tile centre, EXACTLY where addViewEntities/AddEntity put
// a real entity's foot (tile centre, ground elevation, top billboardH above) —
// so Render PROJECTS it with the same camera as terrain/scenery. It records the
// mT face index (not the entity-sprite path) so fillDebugBillboards can read back
// the projected vertex after Render and fill the SOLID rect on top of the 3D
// framebuffer (pure projection, never depth-occluded — Phase-0 isolation).
func placeStaticEntity(scene *Scene, land *pathfind.Landscape, spec *staticEntitySpec, baseX, baseY, plane int) {
	if spec == nil || spec.billboardW <= 0 || spec.billboardH <= 0 {
		return
	}
	lx, lz := spec.tileLocalX, spec.tileLocalY
	if lx < 0 || lx >= worldWindowTiles || lz < 0 || lz >= worldWindowTiles {
		return
	}
	wx := int32(lx*tileWorldUnits + tileWorldUnits/2)
	wz := int32(lz*tileWorldUnits + tileWorldUnits/2)
	wy := -elevationOf(land, baseX, baseY, plane, wx, wz)

	colour := spec.debugColour
	if colour == 0 {
		colour = debugBillboardColour
	}

	// Two billboard verts: foot at (wx,wy,wz), top billboardH above (same as
	// AddEntity / Scene.addSprite). The mT face index == the count BEFORE insert.
	v0 := scene.mT.insertVertex(wx, wy, wz)
	v1 := scene.mT.insertVertex(wx, wy-int32(spec.billboardH), wz)
	if v0 < 0 || v1 < 0 {
		return
	}
	faceIndex := scene.mT.faceHead
	if scene.mT.insertFace(2, []int{v0, v1}, 0, 0) < 0 {
		return
	}
	scene.debugBillboards = append(scene.debugBillboards, debugBillboard{
		faceIndex:  faceIndex,
		billboardW: int32(spec.billboardW),
		billboardH: int32(spec.billboardH),
		colour:     colour,
	})
}

// fillDebugBillboards is called AFTER Scene.Render projects the mT model. For each
// registered debug billboard it reflection-free reads the PROJECTED vertex of the
// billboard's mT foot vertex (vertexParam6 = vx, vertexParam2 = vy, vertZRot = vz)
// and the private projection fields (rot1024VpSrc = viewDistance shift, mZb =
// baseX, mNb = baseY) and recomputes the screen rect with the canonical 5-line
// formula — the SAME w/h/x/y the Java legs' rect replicator recomputes:
//
//	w = (billboardW << viewDist) / vz
//	h = (billboardH << viewDist) / vz
//	x = (vx - w/2) + baseX
//	y = baseY + vy - h
//
// then fills [x,x+w) x [y,y+h) (clipped to the surface) SOLID with the debug
// colour. This is the orsc analogue of the Java drawBox-at-recomputed-rect, so
// all three legs paint the IDENTICAL rect, isolating the projection.
func (s *Scene) fillDebugBillboards() {
	for _, bb := range s.debugBillboards {
		if bb.faceIndex < 0 || bb.faceIndex >= s.mT.faceHead {
			continue
		}
		index := s.mT.faceIndices[bb.faceIndex]
		if len(index) < 1 {
			continue
		}
		v0 := index[0]
		vx := s.mT.vertexParam6[v0]
		vy := s.mT.vertexParam2[v0]
		vz := s.mT.vertZRot[v0]
		if vz <= 0 {
			continue
		}
		w := (bb.billboardW << uint(s.rot1024VpSrc)) / vz
		h := (bb.billboardH << uint(s.rot1024VpSrc)) / vz
		x := (vx - w/2) + s.mZb
		y := s.mNb + vy - h
		fillSolidRect(s.pixelData, int(s.mVb), s.graphics.Width, s.graphics.Height, int(x), int(y), int(w), int(h), bb.colour)
	}
}

// fillSolidRect fills the half-open rect [x,x+w) x [y,y+h), clipped to the surface
// bounds, with the solid colour (0x00RRGGBB) directly into the framebuffer. Same
// clip + write semantics as the Java legs' drawBox-at-recomputed-rect, so the
// rendered rect is byte-identical across legs given the identical (x,y,w,h).
func fillSolidRect(pixels []int32, stride, surfW, surfH, x, y, w, h int, colour int32) {
	for dy := 0; dy < h; dy++ {
		py := y + dy
		if py < 0 || py >= surfH {
			continue
		}
		rowBase := py * stride
		for dx := 0; dx < w; dx++ {
			px := x + dx
			if px < 0 || px >= surfW {
				continue
			}
			pixels[rowBase+px] = colour
		}
	}
}
