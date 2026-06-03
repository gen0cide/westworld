package orsc

import (
	"fmt"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

// dump_faces.go is the orsc-backed retarget of the old render/dump_faces.go. It
// turns an rscdump/1 L1 dump into the engine-comparable []render.BuiltFace set
// the structural diff consumes (RENDER_DIFF_DESIGN.md §5), walking the SAME orsc
// Scene the pixel render rasterizes (terrain + walls/doors/roofs + scenery) and
// reading each model's BAKED transform (resetTransformCache -> vertXTransform/
// YTransform/ZTransform, faceIndices, faceTextureFront/Back, faceHead).
//
// render.BuiltFace / Key() / CentroidGrid stay in the leaf render package (the
// JAR oracle hardcodes them); this file is their PRODUCER. It lives in package
// orsc because the face walk needs the unexported Model internals AND orsc
// already imports render (so render cannot import orsc — the entrypoints can
// only live on this side of the dependency).
//
// The dump-shared scene assembly (buildDumpScene), the synthetic facts, the
// Bundle, and the dump→render.View projection (dumpView) live here too so the
// pixel leg (dump.go) and the face leg share one byte-identical scene build.

// wallObjectHeight is the full-wall height (deob World.java getWallObjectHeight /
// the deleted render.roof.go constant: 192) used by the synthetic boundary def so
// a hand-authored fixture renders a wall/door without loading DoorDef.xml.
const wallObjectHeight = 192

// WallColourWood is the flat WOOD door-leaf colour (Scene.method305(120,90,55) =
// -1 - (120/8)*1024 - (90/8)*32 - 55/8 = -15719), the same fill the synthetic
// boundary + diagonal-object door-leaf carries. Exported so the fidelity tests
// (which assert the leaf colour) reference the one canonical value.
const WallColourWood = int32(-15719)

// Bundle holds the decoded asset archives the dump renderer consumes. The classic
// render.Bundle (a ModelCache) is gone; the orsc renderer reads scenery geometry
// straight out of the OpenRSC models.orsc *assets.Archive (placeSceneryModels
// decodes .ob3 on demand), so the Bundle just carries that archive. Sprites
// (textures) are opened lazily only for the pixel leg — the face leg needs no
// texels (faces carry fill ids, not pixels), so a nil Bundle still builds full
// geometry.
type Bundle struct {
	Models *assets.Archive
	// Textures is the authentic rev-235 "Textures" content archive (content11): the
	// source of the 3D texture texel banks (water/wall/planks/lava/…) the textured
	// scenery-face + textured-overlay fill paths sample. Opened by the mesh harness
	// (cmd/meshrender) so a TEXTURED object face (e.g. the well's 18 wall/planks
	// faces) renders with its real texels instead of the magenta-transparent skip.
	// nil => textures degrade to the flat-skip path (unchanged terrain-only render).
	Textures *assets.Archive
}

// OpenBundle opens the OpenRSC models.orsc archive (the scenery geometry source
// the dump render places). Mirrors the old render.OpenBundle entry so callers
// (the windmill fidelity test, real-map dumps) open a model source the same way.
func OpenBundle(modelsPath string) (*Bundle, error) {
	arc, err := assets.OpenArchive(modelsPath)
	if err != nil {
		return nil, err
	}
	return &Bundle{Models: arc}, nil
}

// RenderDumpFaces builds the SAME Scene RenderDump rasterizes — terrain,
// boundaries, roofs, scenery — from an L1 dump, then exports its built face set
// as []render.BuiltFace for the structural diff. It uses syntheticFacts so a
// hand-authored fixture with no GameData still builds its wall/door (matching
// RenderDump); pass real facts via RenderDumpFacesWith for a real-map dump.
func RenderDumpFaces(d *rscdump.Dump) ([]render.BuiltFace, error) {
	return RenderDumpFacesWith(d, syntheticFacts(d), nil)
}

// RenderDumpFacesWith is RenderDumpFaces with caller-supplied facts + bundle, the
// structural-diff twin of RenderDumpWith. f=nil skips boundaries + scenery (bare
// terrain); b=nil skips scenery MODELS (boundaries still render from f).
func RenderDumpFacesWith(d *rscdump.Dump, f *facts.Facts, b *Bundle) ([]render.BuiltFace, error) {
	sc, err := buildDumpScene(d, f, b, false)
	if err != nil {
		return nil, err
	}
	return sceneFaces(sc), nil
}

// buildDumpScene assembles the orsc Scene from an L1 dump exactly as
// view.go:RenderView assembles the live spectator scene — terrain + textures +
// scenery + stacked walls/doors/roofs + entity billboards — but driven from the
// dump's own landscape, view, facts and scenery list rather than the live world.
// withTextures controls whether the per-texture texel banks are loaded (true for
// the pixel leg; the face leg needs none). It is the SINGLE scene build the pixel
// leg (dump.go) and the face leg share, so a face and a pixel always describe the
// same geometry. Returns an error for a nil / invalid / non-L1 / terrain-less dump.
func buildDumpScene(d *rscdump.Dump, f *facts.Facts, b *Bundle, withTextures bool) (*Scene, error) {
	if d == nil {
		return nil, fmt.Errorf("orsc: nil dump")
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if d.Level != rscdump.LevelL1 {
		return nil, fmt.Errorf("orsc: RenderDump supports level %q, got %q", rscdump.LevelL1, d.Level)
	}
	land := d.Landscape()
	if land == nil {
		return nil, fmt.Errorf("orsc: dump has no terrain")
	}
	v := dumpView(d)

	// The orsc terrain ambience is a deterministic coord-hash (world.go
	// terrainAmbience), so the render is reproducible WITHOUT pinning the JAR's
	// Math.random sequence — the old render.withDumpTerrainSeed RNG-pin is
	// unnecessary here (RENDER_RECONCILIATION.md plan item 6). d.Terrain.TerrainSeed
	// only affects per-vertex SHADE, never geometry, so the face leg is seed-
	// independent and the pixel leg is already byte-reproducible.

	w, h := v.W, v.H
	if w == 0 {
		w = 512
	}
	if h == 0 {
		h = 334
	}
	baseX := v.X - windowCentreTile
	baseY := v.Y - windowCentreTile
	plane := v.Plane

	surf := NewSurface(w, h)
	scene := NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)

	scene.AddModel(BuildTerrain(land, baseX, baseY, plane))

	// Merge the dump's placed scenery (d.Scenery) into the facts as SceneryLocs so
	// orsc's placeSceneryModels (which reads f.SceneryLocs, NOT the live View's
	// DynamicScenery) places them with their real .ob3 models — the path the
	// windmill-sail fidelity test exercises. Bare fixtures carry no scenery, so the
	// merge is a no-op there.
	f = withDumpScenery(f, d)

	var arc *assets.Archive
	if b != nil {
		arc = b.Models
	}
	if withTextures {
		// Textures are only needed for the pixel leg; the face leg skips them.
		// Prefer the AUTHENTIC rev-235 "Textures" content archive (content11, on the
		// Bundle) so TEXTURED scenery faces + terrain overlays render with the real
		// texel banks the DEOB/JAR oracles also load — byte-identical because
		// drawWorld re-quantises the composited RGB (see textures_content11.go). When
		// the Bundle carries no content11, fall back to the OpenRSC Authentic_Sprites
		// path (the live RenderBridge has no content11 Bundle), and ultimately to the
		// flat-skip degrade (LoadTexturesFrom* no-op on a nil source).
		if b != nil && b.Textures != nil {
			scene.LoadTexturesFromContent11(b.Textures)
		} else {
			scene.LoadTexturesFromArchive(bundleSprites(b), textureSlots)
		}
	}

	// Entity layer (NPC/player 2D-sprite billboards). When the entity gate is
	// engaged (RSC_MESH_NPC set OR a fixture npc entity present), Phase 0 places a
	// SINGLE solid-debug-colour billboard at the host centre tile to validate the
	// projection (the placement-sanity gate); the real composited-sprite path
	// (addViewEntities) is suppressed so ONLY the spec entity renders, matching the
	// Java legs (which place exactly the spec billboard). When the gate is NOT
	// engaged, the entity layer is the prior behaviour (addViewEntities), so
	// terrain/scenery fixtures render byte-identically to before.
	hasFixtureNPC := false
	for _, e := range d.Entities {
		if e.Kind == "npc" {
			hasFixtureNPC = true
			break
		}
	}
	playerGate := playerGateEngaged()
	gateOn := entityGateEngaged(hasFixtureNPC) || playerGate
	// Phase 2 (Milestone B): when RSC_NPC_PHASE2 is set, the SPEC rat is placed via
	// the real composited-sprite path (addViewEntities), not the Phase-0 solid
	// debug billboard — exercising the content1 decode + composite + recolour
	// on-screen (the scaler residual is closed in Phase 4). The Phase-0 debug
	// billboard remains the default gate behaviour so placement-sanity reruns are
	// unaffected. The PLAYER gate (RSC_MESH_PLAYER) always uses the real per-layer
	// path (there is no placement-sanity stage for it).
	phase2 := gateOn && phase2RatEnabled()
	if f != nil {
		placeSceneryModels(scene, arc, land, f, baseX, baseY, plane)
		// Stacked walls + doors + roofs for every visible story, threaded through one
		// accumulating elevation cache (BuildStories — incl. the under-roof cull and
		// the id-74 sail lift inside PlaceScenery).
		for _, m := range BuildStories(land, f, baseX, baseY, plane) {
			scene.AddModel(m)
		}
		// Real composited-sprite entities only when the gate is OFF (prior behaviour).
		// When the gate is ON the Phase-0 debug billboard below is the sole entity,
		// UNLESS Phase 2 replaces it with the composited rat via addViewEntities.
		if !gateOn {
			addViewEntities(scene, land, f, v, baseX, baseY, plane)
		}
	}
	if phase2 {
		// Place the SPEC rat at the host centre tile via addViewEntities (the SAME
		// path the live spectator uses). We override facts + view so the SOLE entity
		// is the synthesized rat (suppressing the dump's own demo entities), reading
		// the SAME content1 layers the on-screen 16.16 blit consumes.
		//
		// PARITY (Phase 4): addViewEntities resolves the on-screen facing as
		// (Heading + camTerm) & 7 where camTerm = (v.Rotation+16)/32 — the authentic
		// camera-relative facing. The DEOB/JAR oracle, by contrast, blits the SPEC dir
		// DIRECTLY (frame = sf[dir*3], no camera term — its rect-replicator picks the
		// raw spec frame). To compare the SAME pose we pre-subtract camTerm from the
		// Heading so the final facing equals exactly phase2RatDir() (the spec dir), and
		// the orsc rat renders the IDENTICAL frame the oracle blits.
		camTerm := (v.Rotation + 16) / 32
		ratView := v
		ratView.Entities = []render.Entity{{
			X:       v.X,
			Y:       v.Y,
			Kind:    render.EntityNPC,
			NpcID:   ratServerID,
			Heading: phase2RatDir() - camTerm,
		}}
		ratView.NoSelf = true
		addViewEntities(scene, land, ratFacts(), ratView, baseX, baseY, plane)
	} else if playerGate {
		// PLAYER on-screen path (Phase 4 extension): place the default-human player at
		// the host centre tile via the SAME per-layer 16.16 blit as the rat, with the
		// content0 serverId-1 appearance (head1/body1/legs1) + the DEOB-resolved dye/skin
		// colours. Like the rat, pre-subtract camTerm so the final facing equals the spec
		// dir (the DEOB/JAR player path blits the spec frame directly, no camera term).
		camTerm := (v.Rotation + 16) / 32
		pv := v
		pv.NoSelf = false
		pv.Entities = nil
		pv.X, pv.Y = v.X, v.Y
		pv.SelfHasEquip = true
		pv.SelfEquipSprites = playerParityEquip
		pv.SelfHairColour = playerParityHair
		pv.SelfTopColour = playerParityTop
		pv.SelfTrouserColour = playerParityTrouser
		pv.SelfSkinColour = playerParitySkin
		pv.SelfHeading = playerGateDir() - camTerm
		pv.SelfOffX, pv.SelfOffZ = 0, 0
		addViewEntities(scene, land, &facts.Facts{}, pv, baseX, baseY, plane)
	} else if gateOn {
		// Phase-0 debug billboard runs UNCONDITIONALLY of facts (it needs none — it is
		// a solid rect, like the diagonal-object pass below): a bare fixture (f==nil,
		// no wall defs) must still place it so placement sanity works on bare terrain.
		if spec := phase0EntitySpec(hasFixtureNPC); spec != nil {
			placeStaticEntity(scene, land, spec, baseX, baseY, plane)
		}
	}

	// Diagonal scenery objects / diagonal doors: the 48001..59999 DiagonalWalls band
	// the wall + scenery passes drop (World.addModels, diagobj.go). Driven by the
	// landscape diag grid, NOT facts, so it runs even for a bare fixture with no
	// boundary/scenery defs (f==nil) — a hand-authored diagonal door then builds a
	// synthetic wood door-leaf (the old render.buildScene called this unconditionally,
	// not gated on facts). Without this a diagonal door renders as ZERO geometry.
	for _, m := range BuildDiagonalObjects(arc, land, f, baseX, baseY, plane) {
		scene.AddModel(m)
	}

	// Camera: dumpView already inverted the dump's 0..1023 angles back to render.View
	// units (yaw/4, distance/2); RenderView's own SetCamera re-derives them (yaw*4,
	// distance*2), so the eye matches the dump verbatim.
	cx := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffX)
	cz := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffZ)
	elev := elevationOf(land, baseX, baseY, plane, cx, cz)
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	// AUTHENTIC rev-235 clip/fog so the per-vertex shade fog ramp + far-clip match
	// the DEOB/JAR oracle (see the auth* constants in harness.go). OpenRSC's
	// defaults (fogSmoothingStartDistance=10, fogZFalloff=20, fogLandscape=10000)
	// would over-brighten + over-extend distant terrain vs the vanilla client.
	scene.fogSmoothingStartDistance = authFogZDistance // deob fogZDistance/G = 2300
	scene.fogZFalloff = authFogZFalloff                // deob fogZFalloff/P  = 1
	scene.fogLandscapeDistance = authClipFar           // deob clipFar3d      = 2400
	scene.fogEntityDistance = authClipFar
	scene.SetCamera(cx, -elev, cz, cameraPitch, int32(v.Rotation)*4, 0, int32(zoom)*2)
	return scene, nil
}

// bundleSprites lazily opens the texture sprite archive for the pixel leg. The
// Bundle only carries the models archive (geometry); textures come from the
// canonical sprites path. Returns nil (textures skipped) if it can't be opened —
// a bare/synthetic fixture renders its flat-coloured walls without texels anyway.
func bundleSprites(b *Bundle) *assets.SpriteArchive {
	if b == nil {
		return nil
	}
	sa, err := assets.OpenSprites(spritesPath)
	if err != nil {
		return nil
	}
	return sa
}

// withDumpScenery returns f augmented with one facts.SceneryLoc per dump-placed
// scenery object (d.Scenery), so orsc's SceneryLocs-driven placeSceneryModels
// places them. f is not mutated — a shallow copy carries the extended SceneryLocs
// (the def maps are shared, read-only). Returns f unchanged when the dump places
// no scenery, and a fresh facts holding only the locs when f is nil but scenery
// exists (so the loc has somewhere to live; without a def it is skipped harmlessly).
func withDumpScenery(f *facts.Facts, d *rscdump.Dump) *facts.Facts {
	if len(d.Scenery) == 0 {
		return f
	}
	out := &facts.Facts{}
	if f != nil {
		*out = *f // shallow copy: share the def maps + other locs, replace SceneryLocs
	}
	locs := make([]facts.SceneryLoc, 0, len(out.SceneryLocs)+len(d.Scenery))
	locs = append(locs, out.SceneryLocs...)
	for _, s := range d.Scenery {
		locs = append(locs, facts.SceneryLoc{DefID: s.ID, X: s.X, Y: s.Y, Direction: s.Dir})
	}
	out.SceneryLocs = locs
	// orsc's placeSceneryModels iterates SceneryLocs directly (it does not consult
	// the by-tile spatial index), so no reindex is needed for placement.
	return out
}

// sceneFaces walks every model in the scene and emits one render.BuiltFace per
// face, computing each face's centroid from the model's TRANSFORMED vertices
// (resetTransformCache bakes the queued rotate/scale/translate into
// vertXTransform/YTransform/ZTransform — the same transform the camera then
// projects). Degenerate faces (<3 verts) are skipped, matching the rasterizer's
// own guard. The centroid is rounded to render.CentroidGrid via render.RoundToGrid
// so the diff key is byte-identical to the old in-package walker.
func sceneFaces(sc *Scene) []render.BuiltFace {
	if sc == nil {
		return nil
	}
	var out []render.BuiltFace
	for mi := 0; mi < sc.modelCount; mi++ {
		m := sc.models[mi]
		if m == nil || m.faceHead == 0 {
			continue
		}
		m.resetTransformCache() // bake transform into vertX/Y/ZTransform (idempotent once clean)
		for fi := 0; fi < m.faceHead; fi++ {
			verts := m.faceIndices[fi]
			if len(verts) < 3 {
				continue
			}
			var sx, sy, sz int64
			for _, vtx := range verts {
				sx += int64(m.vertXTransform[vtx])
				sy += int64(m.vertYTransform[vtx])
				sz += int64(m.vertZTransform[vtx])
			}
			n := int64(len(verts))
			out = append(out, render.BuiltFace{
				Model: mi,
				Centroid: [3]int32{
					render.RoundToGrid(int32(sx/n), render.CentroidGrid),
					render.RoundToGrid(int32(sy/n), render.CentroidGrid),
					render.RoundToGrid(int32(sz/n), render.CentroidGrid),
				},
				NumVerts:  len(verts),
				FillFront: m.faceTextureFront[fi],
				FillBack:  m.faceTextureBack[fi],
			})
		}
	}
	return out
}

// dumpView projects an L1 dump onto the render.View the orsc scene build consumes.
// The dump camera angles are 0..1023 (pre-inversion); the scene build re-derives
// the orsc camera from View.Rotation (0..255, yaw = Rotation*4) + View.Zoom
// (distance = Zoom*2), so we convert back: Rotation = Yaw/4, Zoom = Distance/2.
// Pitch/roll in the fixture match the orsc fixed pitch=912/roll=0, so they need no
// carry here.
func dumpView(d *rscdump.Dump) render.View {
	v := render.View{
		X:         d.Window.BaseX + d.Terrain.Size/2,
		Y:         d.Window.BaseY + d.Terrain.Size/2,
		Plane:     d.Window.Plane,
		Rotation:  int(d.Camera.Yaw) / 4,
		Zoom:      int(d.Camera.Distance) / 2,
		W:         d.Camera.ScreenW,
		H:         d.Camera.ScreenH,
		AnimFrame: 0,
		NoSelf:    true,
	}
	// Self pose: a dump may centre the view on the local player's tile and carry his
	// appearance; honour it when present (the bare-terrain fixtures set NoSelf).
	if d.Self != nil {
		v.X = d.Self.X
		v.Y = d.Self.Y
		v.SelfHeading = d.Self.Heading
		v.NoSelf = d.Self.NoSelf
		v.SelfHasEquip = d.Self.HasEquip
		v.SelfEquipSprites = d.Self.Equip
		v.SelfHairColour = d.Self.Hair
		v.SelfTopColour = d.Self.Top
		v.SelfTrouserColour = d.Self.Trouser
		v.SelfSkinColour = d.Self.Skin
	}

	for _, e := range d.Entities {
		kind := render.EntityPlayer
		if e.Kind == "npc" {
			kind = render.EntityNPC
		}
		v.Entities = append(v.Entities, render.Entity{
			X: e.X, Y: e.Y, Kind: kind, NpcID: e.ID, Heading: e.Heading,
			HasEquip: e.HasEquip, EquipSprites: e.Equip,
			HairColour: e.HairCol, TopColour: e.TopCol,
			TrouserColour: e.TrouserCol, SkinColour: e.SkinCol,
		})
	}
	for _, gi := range d.GroundItems {
		v.GroundItems = append(v.GroundItems, render.GroundItemMarker{X: gi.X, Y: gi.Y, ItemID: gi.ItemID})
	}
	if len(d.RemovedBoundaries) > 0 {
		set := make(map[[3]int]bool, len(d.RemovedBoundaries))
		for _, e := range d.RemovedBoundaries {
			set[[3]int{e.X, e.Y, e.Dir}] = true
		}
		v.BoundaryRemoved = func(x, y, dir int) bool { return set[[3]int{x, y, dir}] }
	}
	if len(d.RemovedScenery) > 0 {
		set := make(map[[2]int]bool, len(d.RemovedScenery))
		for _, t := range d.RemovedScenery {
			set[[2]int{t.X, t.Y}] = true
		}
		v.SceneryRemoved = func(x, y int) bool { return set[[2]int{x, y}] }
	}
	return v
}

// syntheticFacts builds a minimal facts.Facts that gives the orsc wall/door pass a
// def for every wall/door id the dump's terrain grids reference, so a hand-authored
// fixture renders a wall/door WITHOUT loading external GameData (§4 rule 5). The
// def is a generic openable boundary (wood door leaf): Unknown=1 marks it openable
// (so orsc's doorPass — not the static wall pass — builds it), Height=192 full
// wall, and FrontDeco/BackDeco carry the flat WOOD colour directly (a negative
// fill is a flat 5:5:5 colour, so no texture archive is needed). Returns nil if the
// dump references no walls (orsc skips the boundary pass cleanly).
func syntheticFacts(d *rscdump.Dump) *facts.Facts {
	if d.Terrain == nil {
		return nil
	}
	ids := map[int]bool{}
	collect := func(g []byte, oneBased bool) {
		for _, bb := range g {
			if bb == 0 {
				continue
			}
			if oneBased {
				ids[int(bb)-1] = true
			} else {
				ids[int(bb)] = true
			}
		}
	}
	collect(d.Terrain.WallH, true)
	collect(d.Terrain.WallV, true)
	for _, dv := range d.Terrain.WallDiag {
		if dv <= 0 || dv >= 24000 {
			continue
		}
		if dv < 12000 {
			ids[int(dv)-1] = true
		} else {
			ids[int(dv-12001)] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	defs := make(map[int]*facts.BoundaryDef, len(ids))
	for id := range ids {
		defs[id] = &facts.BoundaryDef{
			ID:        id,
			Name:      "Door",
			Height:    wallObjectHeight,
			FrontDeco: int(WallColourWood),
			BackDeco:  int(WallColourWood),
			DoorType:  1,
			// Unknown=0 -> built by orsc's STATIC wallPass (World.java method422,
			// wall light 122), NOT doorPass (door-object light -95). The authentic
			// deob/JAR render an injected boundary as a standing wall via method422
			// (no live server door-object instance exists to swap in the -95 leaf),
			// so the dump leg must shade it with the SAME wall light to match 1:1.
			Unknown: 0,
		}
	}
	return &facts.Facts{BoundaryDefs: defs}
}
