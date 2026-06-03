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
// DEFAULT (real sprite): a bare RSC_MESH_NPC=<id> / RSC_MESH_PLAYER=<gate> now
// places the REAL composited sprite at the host centre tile via addViewEntities
// (the SAME path the live spectator uses), 0/0/0 against the DEOB/JAR oracles.
//
// PLACEMENT-SANITY OPT-IN (RSC_NPC_DEBUG_BILLBOARD): the spec can instead drive a
// SOLID DEBUG-COLOUR billboard (NOT the real composited sprite) at the host centre
// tile, so a PNG diff of the screen rect isolates the PROJECTION (billboard
// world-size -> screen w/h/x/y) from sprite decode/composite. This is a RETAINED
// diagnostic, fired ONLY when the NPC gate is engaged AND RSC_NPC_DEBUG_BILLBOARD
// is set. The first entity is the rat (content0 serverId 19), whose billboard is
// 346x136 (entityIndexTableC[19]/legacyMaskTable[19] == OpenRSC NpcDefs id 19
// camera1/camera2). The debug spec reuses the SAME placement the real sprite path
// uses; it just swaps the real sprite stack for a solid fill.

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

// ratNpcDef synthesizes the NPC-Rat (content0 serverId 19) NpcDef the entity
// composite path consumes: ONE non-empty body-part layer (Sprites[0] = animID 123,
// "rat"), all other layers -1 (skipped). The rat's charColour (4805259) is a RAW
// 24-bit dye value carried by authenticAnimDefs[123], NOT a 1/2/3 marker, so the
// dye path is identity (NPCs use rawColours=true); HairColour/TopColour/...
// /SkinColour are 0 (no recolour). Camera1/Camera2 are the world-space billboard
// 346x136 (entityIndexTableC[19]/legacyMaskTable[19]). Driving the spec off this
// def (instead of the full OpenRSC NpcDefs.json) keeps the Phase-2 composite path
// self-contained — the SAME spec the DEOB/JAR legs encode.
func ratNpcDef() *facts.NpcDef {
	return &facts.NpcDef{
		Sprites:      [12]int{ratServerNPCAnimID, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		HairColour:   0,
		TopColour:    0,
		BottomColour: 0,
		SkinColour:   0,
		Camera1:      ratBillboardW,
		Camera2:      ratBillboardH,
	}
}

// ratServerNPCAnimID is the rat's body-part animation id (authenticAnimDefs index,
// "rat" => animationNumbers()[123] = 837). animID+1 is the synthesized equippedItem
// the DEOB/JAR NPC path reads (the -1 recovers 123); the orsc NPC path seeds
// NpcDef.Sprites directly with the animID (no +1), so we store 123 here.
const ratServerNPCAnimID = 123

// ratFacts returns a facts.Facts carrying ONLY the synthesized rat NpcDef, so
// addViewEntities (which reads f.NpcDefs[id] via render.CompositeNPCSprite) can
// composite + place the rat without loading OpenRSC's NpcDefs.json.
func ratFacts() *facts.Facts {
	return &facts.Facts{NpcDefs: map[int]*facts.NpcDef{ratServerID: ratNpcDef()}}
}

// DumpRatCompositeCanvas builds the orsc UNSCALED CompositeSprite canvas for the
// rat (the SAME render.CompositeSprite addViewEntities places) and encodes it as a
// straight-alpha NRGBA PNG — the orsc side of the Phase-2 (Milestone B) byte
// compare against the DEOB-reconstructed rat-layer canvas. dir/step select the
// pose (0/0 = standing south, frame 0). Returns the PNG bytes + the canvas W/H +
// opaque-pixel count, or an error if the composite is unavailable (e.g. content1
// absent). This isolates the decode+composite+recolour (which Phase 2 gates to
// DIFFS=0) from the on-screen scaler (the remaining orsc gap, Phase 4).
func DumpRatCompositeCanvas(dir, step int) (png []byte, w, h, opaque int, err error) {
	cs := render.CompositeNPCSprite(ratFacts(), ratServerID, dir, step)
	if cs == nil || cs.W <= 0 || cs.H <= 0 {
		return nil, 0, 0, 0, fmt.Errorf("orsc: rat composite unavailable (content1 missing?)")
	}
	for _, o := range cs.Opaque {
		if o {
			opaque++
		}
	}
	return render.CompositeSpritePNG(cs), cs.W, cs.H, opaque, nil
}

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
	ratServerID   = 19
	ratBillboardW = 346
	ratBillboardH = 136
)

// staticEntitySpec is one entity to place: its host-window-LOCAL tile, its
// world-space billboard size, and (Phase 0) the solid debug fill. tileLocalX/Y
// are window-local (0..worldWindowTiles); windowCentreTile (=48) is the host
// centre tile, where the camera sits.
type staticEntitySpec struct {
	tileLocalX  int
	tileLocalY  int
	billboardW  int
	billboardH  int
	debug       bool // Phase 0: fill the projected rect with debugColour, not a sprite
	debugColour int32
	serverID    int
	dir         int
	step        int
}

// phase0EntitySpec returns the placement-sanity debug billboard spec when the
// entity gate is engaged, or nil when it is not (so terrain/scenery fixtures render
// unchanged). The gate is engaged by either:
//   - RSC_MESH_NPC=<serverId>[:<dir>:<step>]  (env, mirrors RSC_MESH_REALDEFS), or
//   - a fixture Entities[] entry of kind "npc" (the rscdump field),
//
// with the env gate taking precedence (it can override a fixture's entity). This
// spec is consumed ONLY behind the RSC_NPC_DEBUG_BILLBOARD opt-in (the default gate
// behaviour is the real composited sprite). It places exactly ONE billboard, at the
// host centre tile (grid-local 48), of the rat's 346x136 size, filled solid —
// proving placement before any decode.
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

// debugBillboardEnabled reports whether the placement-sanity SOLID debug billboard
// is explicitly opted in (RSC_NPC_DEBUG_BILLBOARD set, any non-empty value). The
// debug billboard is a retained DIAGNOSTIC tool (it isolates the projection from
// sprite decode/composite), NOT the default gate behaviour: a bare RSC_MESH_NPC=<id>
// now draws the REAL composited sprite. Only when this opt-in is set AND the NPC
// gate is engaged does the entity layer fall back to the solid debug rect.
func debugBillboardEnabled() bool {
	return strings.TrimSpace(os.Getenv("RSC_NPC_DEBUG_BILLBOARD")) != ""
}

// npcGateDir parses the facing dir from RSC_MESH_NPC=<id>[:<dir>[:<step>]]
// (default 0 = south, the standing frame-0 pose), so the on-screen composited rat
// faces the spec direction.
func npcGateDir() int {
	gate := strings.TrimSpace(os.Getenv("RSC_MESH_NPC"))
	parts := strings.Split(gate, ":")
	if len(parts) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			return v & 7
		}
	}
	return 0
}

// entityGateEngaged reports whether the entity layer is active for this render
// (env gate set OR a fixture npc entity present). When false the dump path must
// render terrain/scenery EXACTLY as before (the Phase-0 gate invariant).
func entityGateEngaged(hasFixtureNPC bool) bool {
	return strings.TrimSpace(os.Getenv("RSC_MESH_NPC")) != "" || hasFixtureNPC
}

// playerGateEngaged reports whether the on-screen PLAYER entity path is active
// (RSC_MESH_PLAYER set), mirroring the DEOB/JAR playerGateEngaged. It drives the
// default-human player (head1/body1/legs1, content0 serverId-1 colours) through the
// SAME faithful per-layer 16.16 blit as the rat (Phase 4 / Milestone C extension).
func playerGateEngaged() bool {
	return strings.TrimSpace(os.Getenv("RSC_MESH_PLAYER")) != ""
}

// itemGateId parses the gated ground-item id from RSC_MESH_ITEM, or -1 when the env
// is unset / not an int. The ground-item entity arm (Task B2) is active iff this is
// >= 0. MUTUALLY EXCLUSIVE with the NPC/player gates (one entity arm per run) — the
// DEOB/JAR B1 legs read the SAME env (DumpRender.itemGateId).
func itemGateId() int {
	g := strings.TrimSpace(os.Getenv("RSC_MESH_ITEM"))
	if g == "" {
		return -1
	}
	if v, err := strconv.Atoi(g); err == nil && v >= 0 {
		return v
	}
	return -1
}

// itemGateEngaged reports whether the ground-item entity arm is active (RSC_MESH_ITEM
// set to a valid item id). When false the dump path renders terrain/scenery exactly
// as before.
func itemGateEngaged() bool { return itemGateId() >= 0 }

// playerGateDir parses the facing dir from RSC_MESH_PLAYER=<gate>[:<dir>[:<step>]],
// defaulting to dir 0 (south, the standing frame-0 pose). The LEADING field is the
// enable token (any non-empty value engages the gate, e.g. RSC_MESH_PLAYER=1) and is
// NOT a direction — this mirrors the DEOB/JAR oracle, whose PLAYER_FRAME is fixed at
// dir 0 / step 0 and which treats RSC_MESH_PLAYER purely as an on/off gate
// (DumpRender.playerGateEngaged). Reading the dir from parts[0] (the old behaviour)
// made RSC_MESH_PLAYER=1 render the dir-1 turned pose while the oracle rendered dir 0,
// causing an 880px per-frame mismatch. The optional :<dir> field (parts[1]) overrides
// the dir, matching the NPC gate's RSC_MESH_NPC=<id>[:<dir>[:<step>]] layout.
func playerGateDir() int {
	gate := strings.TrimSpace(os.Getenv("RSC_MESH_PLAYER"))
	parts := strings.Split(gate, ":")
	if len(parts) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			return v & 7
		}
	}
	return 0
}

// Default-human player appearance for the on-screen parity render (content0 serverId
// 1, the SAME the DEOB/JAR player path resolves): the appearance grid head1/body1/
// legs1 (animIDs 0/1/2) and the content0-default dye/skin colours (RAW 24-bit, so the
// player path's resolveClothingColour uses them verbatim — out-of-table values pass
// through). equip[layer]-1 is the animation id (head slot 0 -> equip 1, body slot 1 ->
// equip 2, legs slot 2 -> equip 3); the per-layer marker (1/2/3 on head1/body1/legs1)
// selects hair/top/bottom for the in-blit dye, skin for the r==255 recolour — exactly
// the DEOB phase3 blit (dye hair=0x303030 top/bottom=0xff0000 skin=0x906020).
var (
	playerParityEquip   = [12]int{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	playerParityHair    = 0x303030 // 3158064  (marker-1 dye)
	playerParityTop     = 0xff0000 // 16711680 (marker-2 dye)
	playerParityTrouser = 0xff0000 // 16711680 (marker-3 dye)
	playerParitySkin    = 0x906020 // 9461792  (skin recolour)
)

// itemBillboardW / itemBillboardH are the ground item's LITERAL world-space billboard
// size, fixed by the authentic queue (Mudclient.java:6562 addSprite(40000+itemId, …,
// 96, 64, 109)). NOT icon-canvas-derived: every dropped item, regardless of icon size,
// is a 96x64 world billboard. The DEOB/JAR B1 legs hardcode the SAME 96x64.
const (
	itemBillboardW = 96
	itemBillboardH = 64
)

// placeGroundItem synthesizes ONE ground item at the host centre tile and registers it
// via the RAW 16.16 path (scene.AddEntityLayers — critique #1) at the literal 96x64 world
// billboard size (critique: itemBillboardW/H). The 1-layer EntitySprite's item-picture
// pixels are decoded from content8 and recoloured in-blit via transparentSpritePlot
// (Dye=pictureMask/Skin=0 — critique #2), exactly the DEOB/JAR B1 spriteClipping. Foot at
// the centre-tile centre, ground elevation, NO glide offset (a fresh drop). No-op when the
// item has no in-scope picture / content8 is unavailable (es==nil).
func placeGroundItem(scene *Scene, land *pathfind.Landscape, itemID, baseX, baseY, plane int) {
	es := render.ItemEntityLayers(itemID)
	if es == nil || len(es.Layers) == 0 {
		return
	}
	wx := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2)
	wz := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2)
	wy := -elevationOf(land, baseX, baseY, plane, wx, wz)
	scene.AddEntityLayers(wx, wy, wz, int32(itemBillboardW), int32(itemBillboardH), es)
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
