package render

import (
	"fmt"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/internal/rscdump"
)

// RenderDump renders an rscdump/1 LEVEL-L1 dump through the unchanged RenderView
// pipeline and returns the encoded PNG, the raw int32 framebuffer (for a
// byte-exact render diff against the other engines), and any error.
//
// This is the GO engine's "render a dump → PNG/int[]" entrypoint
// (RENDER_DIFF_DESIGN.md §5). It is deterministic by construction:
//   - The terrain grids are read straight out of the dump (Dump.Landscape), so
//     no map file is decoded (§4 rule 1).
//   - The per-vertex ambience speckle is pinned to dump.Terrain.TerrainSeed for
//     the duration of the render (withDumpTerrainSeed), so the RNG that the
//     vanilla client injects per vertex is replaced by a fixed, reproducible
//     value (§4 rule 2). seed 0 = flat zero speckle ⇒ byte-identical output.
//   - Camera/window/animation are taken verbatim from the dump (§4 rules 3-4).
//
// Colours: for a hand-authored fixture with no external GameData (§4 rule 5),
// callers leave the Bundle nil (no scenery models) and rely on the ground
// palette + a synthesized door/wall def. RenderDump builds a minimal
// facts.Facts carrying a generic openable boundary def for every wall id the
// terrain grids reference, so a wall/door renders without loading DoorDef.xml.
// A caller that has real GameData can pass it instead via RenderDumpWith.
func RenderDump(d *rscdump.Dump) (pngBytes []byte, rawPix []int32, err error) {
	return RenderDumpWith(d, syntheticFacts(d), nil)
}

// RenderDumpWith renders a dump with caller-supplied facts (boundary/scenery
// defs) and asset bundle (scenery models). Pass f=nil to skip the boundary +
// scenery passes entirely (bare terrain); pass b=nil to skip scenery models
// (boundaries still render from f). This is the seam Phase 3+ uses to render a
// real-map dump against the same GameData the live client uses.
func RenderDumpWith(d *rscdump.Dump, f *facts.Facts, b *Bundle) (pngBytes []byte, rawPix []int32, err error) {
	if d == nil {
		return nil, nil, fmt.Errorf("render: nil dump")
	}
	if err := d.Validate(); err != nil {
		return nil, nil, err
	}
	if d.Level != rscdump.LevelL1 {
		return nil, nil, fmt.Errorf("render: RenderDump supports level %q, got %q", rscdump.LevelL1, d.Level)
	}
	land := d.Landscape()
	if land == nil {
		return nil, nil, fmt.Errorf("render: dump has no terrain")
	}
	v := dumpView(d)

	var surf *Surface
	withDumpTerrainSeed(d.Terrain.TerrainSeed, func() {
		surf, err = renderViewToSurface(land, f, b, v)
	})
	if err != nil {
		return nil, nil, err
	}
	png, err := surf.PNG()
	if err != nil {
		return nil, nil, err
	}
	// Copy the framebuffer so the caller owns it independently of the Surface.
	raw := make([]int32, len(surf.Pix))
	copy(raw, surf.Pix)
	return png, raw, nil
}

// dumpView projects an L1 dump onto the render.View RenderView consumes. The
// camera angles are 0..1023 (pre-inversion); RenderView re-derives the GO camera
// from View.Rotation (0..255, yaw = rotation*4) + View.Zoom (distance = zoom*2),
// so we convert back: rotation = yaw/4, zoom = distance/2. Pitch/roll in the
// fixture match RenderView's fixed pitch=912/roll=0, so they need no carry here
// (a later phase that needs arbitrary pitch will widen View).
func dumpView(d *rscdump.Dump) View {
	v := View{
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
	// Self pose: a dump may centre the view on the local player's tile and carry
	// his appearance; honour it when present (the bare-terrain fixtures set
	// NoSelf so the scene is purely world geometry).
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
		kind := EntityPlayer
		if e.Kind == "npc" {
			kind = EntityNPC
		}
		ent := Entity{
			X: e.X, Y: e.Y, Kind: kind, NpcID: e.ID, Heading: e.Heading,
			HasEquip: e.HasEquip, EquipSprites: e.Equip,
			HairColour: e.HairCol, TopColour: e.TopCol,
			TrouserColour: e.TrouserCol, SkinColour: e.SkinCol,
		}
		v.Entities = append(v.Entities, ent)
	}
	for _, s := range d.Scenery {
		v.DynamicScenery = append(v.DynamicScenery, DynamicSceneryItem{X: s.X, Y: s.Y, ID: s.ID, Direction: s.Dir})
	}
	for _, gi := range d.GroundItems {
		v.GroundItems = append(v.GroundItems, GroundItemMarker{X: gi.X, Y: gi.Y, ItemID: gi.ItemID})
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

// syntheticFacts builds a minimal facts.Facts that gives BuildBoundaries a def
// for every wall/door id the dump's terrain grids reference, so a hand-authored
// fixture renders a wall/door WITHOUT loading external GameData (§4 rule 5). The
// def is a generic openable boundary (wood door leaf): Unknown=1 marks it
// openable, Height=192 full wall, and FrontDeco/BackDeco carry the flat WOOD
// colour directly (method422 now paints the def's modelVar2/3 verbatim, so a
// negative fill is the flat door-leaf colour — no texture archive needed).
// Returns nil if the dump references no walls (RenderView skips the boundary
// pass cleanly).
func syntheticFacts(d *rscdump.Dump) *facts.Facts {
	if d.Terrain == nil {
		return nil
	}
	ids := map[int]bool{}
	collect := func(g []byte, oneBased bool) {
		for _, b := range g {
			if b == 0 {
				continue
			}
			if oneBased {
				ids[int(b)-1] = true
			} else {
				ids[int(b)] = true
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
			ID:   id,
			Name: "Door",
			// FrontDeco/BackDeco carry the flat WOOD door-leaf colour directly
			// (method422 paints modelVar2/3 verbatim; a <0 value is a flat 5:5:5
			// colour, so no texture archive is needed). The same -15719 the
			// renderdiff door self-test asserts.
			Height:    wallObjectHeight,
			FrontDeco: int(wallColourWood),
			BackDeco:  int(wallColourWood),
			DoorType:  1,
			Unknown:   1, // openable
		}
	}
	return &facts.Facts{BoundaryDefs: defs}
}
