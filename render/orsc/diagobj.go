package orsc

import (
	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// diagobj.go ports World.addModels (deob World.java:826-854; clean k.java:
// 1930-2010; OpenRSC three/ World.addLoginScreenModels World.java:162-216) — the
// diagonal SCENERY-object / diagonal-DOOR pass the orsc World build never had.
//
// RSC encodes a diagonally-PLACED scenery object (INCLUDING a diagonal door) as
// an id in the 48001..59999 band of the per-tile DiagonalWalls grid (a marker,
// NOT a wall). The wall passes (walls.go wallPass/doorPass) correctly consume only
// the <24000 wall bands (World.java:858-875), and scenery_place.go drives static
// scenery off SceneryLocs.json — so the 48000+ band is dropped by BOTH and a
// diagonal door renders as ZERO geometry. This file realizes that band as placed
// object models, driven off the landscape diag grid exactly like World.addModels:
//
//	for each tile (x,z) with  48000 < getWallDiagonal(x,z) < 60000:
//	    objectId = getWallDiagonal(x,z) - 48001          (World.java/k.java)
//	    dir      = getTileDirection(x,z)                  (TileDirection grid)
//	    place the object's model at the footprint CENTRE, oriented by dir*32 about
//	    the VERTICAL axis, snapped to -getElevation; clear the diag id over the
//	    object's WxH footprint so a multi-tile object builds once.
//
// Placement is REUSED verbatim from the verified scenery pass: PlaceScenery /
// SceneryCentre (world.go:632,659) apply the SAME footprint-centre
// ((xSize+2x)*128/2), the SAME dir 0/4 width/height swap (World.java:173-179), the
// SAME -getElevation bilinear ground-snap, and the SAME orientation. On the
// orientation: addModels orients by dir*32 about the VERTICAL (Y) axis —
//
//	deob:    model.orient(yaw=0, guard=-999999, pitch=0, roll=dir*32)
//	         where applyRotation's "roll" term mixes X<->Z, i.e. it spins the model
//	         about the vertical Y axis (GameModel.java:835-945).
//	OpenRSC: copy.setRot256(0, dir*32, 0)  -> rot256Y = dir*32
//
// orsc's rotate256 applies rot256Y via the term that mixes Z<->X (model.go:427-433)
// — the identical vertical-axis spin. PlaceScenery's Orient(0, dir*32, 0)
// (world.go:661) is therefore byte-for-byte the addModels roll; no axis change is
// needed, the diagonal object is oriented by the SAME, already-audited helper the
// static scenery uses.
//
// The footprint-clear (World.java:847-852) is applied to a LOCAL snapshot of the
// diag band so a 2x2+ object builds once (its other covered tiles are skipped)
// without mutating the shared pathfind.Landscape (the terrain/wall passes also read
// it). The clear extent uses the SAME dir 0/4 swap PlaceScenery applies, so it
// matches the footprint the placement used.
//
// Inputs mirror placeSceneryModels (scenery_place.go): arc is the models archive
// (the object's .ob3 geometry), f carries the object/scenery defs. With arc==nil
// (a hand-authored fixture with no GameData/archive) a synthetic wood door-leaf is
// built with the IDENTICAL placement math, mirroring how render/'s fixture path
// renders a diagonal door as *something* rather than nothing. Returns the built
// models for the caller to Scene.AddModel; nil if the window holds no diagonal
// objects.

// diagObjectBand is the DiagonalWalls value range RSC uses to encode a
// diagonally-PLACED scenery object (not a wall): getWallDiagonal in (48000,60000)
// => objectId = value - 48001 (World.java:830-832; OpenRSC World.java:168-169).
const (
	diagObjectLo = 48000 // exclusive lower bound (48001 is the first valid id)
	diagObjectHi = 60000 // exclusive upper bound
	diagObjectID = 48001 // objectId = diagonal value - diagObjectID
)

// BuildDiagonalObjects builds the placed models for every diagonal scenery-object
// / diagonal-door in the [baseX,baseX+96)x[baseY,baseY+96) window at the given
// plane (World.addModels). See the file comment for the algorithm + ground truth.
func BuildDiagonalObjects(arc *assets.Archive, land *pathfind.Landscape, f *facts.Facts, baseX, baseY, plane int) []*Model {
	if land == nil {
		return nil
	}
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}

	// Local snapshot of the diagonal-OBJECT band over the window, so the
	// footprint-clear (World.java:847-852) can zero a multi-tile object's other
	// covered cells WITHOUT mutating the shared landscape. Indexed [x][z] in
	// window-local coords; only the object band is captured (walls 1..23999 are
	// irrelevant here).
	var diag [windowTiles][windowTiles]int32
	for x := 0; x < windowTiles; x++ {
		for z := 0; z < windowTiles; z++ {
			d := b.getWallDiagonal(x, z)
			if d > diagObjectLo && d < diagObjectHi {
				diag[x][z] = d
			}
		}
	}

	// Per-decoded-model cache keyed by model name, mirroring placeSceneryModels.
	modelCache := map[string]*assets.Model{}
	getOb3 := func(name string) *assets.Model {
		if m, ok := modelCache[name]; ok {
			return m
		}
		m := decodeOb3(arc, name)
		modelCache[name] = m
		return m
	}

	var out []*Model
	// addModels scans x (outer) then z (inner) — preserve that order so the
	// footprint-clear of a multi-tile object reaches its +x/+z cells before they
	// are visited (World.java:827-828). The deob/OpenRSC loop bound is 94.
	for x := 0; x < windowTiles-2; x++ {
		for z := 0; z < windowTiles-2; z++ {
			d := diag[x][z]
			if d <= diagObjectLo || d >= diagObjectHi { // already filtered; explicit (World.java:830)
				continue
			}
			objectID := int(d) - diagObjectID      // World.java:831
			dir := int(b.tile(x, z).TileDirection) // World.java:832 getTileDirection

			// Footprint (un-swapped Width,Height) from the def; PlaceScenery /
			// SceneryCentre apply the dir 0/4 swap (World.java:173-179), so a single
			// source of truth handles the orientation/footprint geometry. Width/Height
			// default to 1 when unknown (matches SceneryCentre's <=0 guard).
			w, hgt := 1, 1
			if f != nil {
				if def := f.SceneryDef(objectID); def != nil {
					if def.Width > 0 {
						w = def.Width
					}
					if def.Height > 0 {
						hgt = def.Height
					}
				}
			}

			// Build + place the model exactly like the static scenery pass: a real
			// def+model is placed via PlaceScenery; with no archive a hand-authored
			// fixture gets a synthetic wood door-leaf placed by the SAME math. A
			// missing def/model on the real path yields nil and the object is silently
			// skipped (matches placeSceneryModels / the deob, which drops unknown ids).
			if g := b.buildDiagObject(arc, f, getOb3, objectID, x, z, dir, w, hgt); g != nil {
				out = append(out, g)
			}

			// Footprint-clear (World.java:847-852): zero the diag id over the rest of
			// the object's WxH footprint so the inner tiles don't re-build it. The
			// extent swaps for dir 0/4 the SAME way PlaceScenery does.
			cw, ch := w, hgt
			if dir != 0 && dir != 4 {
				cw, ch = hgt, w
			}
			if cw <= 1 && ch <= 1 {
				continue
			}
			for gx := x; gx < x+cw && gx < windowTiles; gx++ {
				for gz := z; gz < z+ch && gz < windowTiles; gz++ {
					if (gx > x || gz > z) && diag[gx][gz] == d {
						diag[gx][gz] = 0
					}
				}
			}
		}
	}
	return out
}

// buildDiagObject builds the placed *Model for one diagonal object at window-local
// tile (x,z): the real def+model placed via PlaceScenery, or — when no archive /
// def / model is available (a hand-authored fixture) — a synthetic wood door-leaf
// placed with the IDENTICAL footprint-centre + dir*32 orient + -elevation snap.
// Returns nil when a real-data object's def/model is missing (the deob drops it).
func (b *terrainBuilder) buildDiagObject(arc *assets.Archive, f *facts.Facts,
	getOb3 func(string) *assets.Model, objectID, x, z, dir, w, hgt int) *Model {

	// Real-data path: an object def with a model name + decodable .ob3 geometry is
	// placed by the verified scenery helper (footprint centre + dir*32 vertical-axis
	// orient + bilinear -elevation snap).
	if f != nil && arc != nil {
		if def := f.SceneryDef(objectID); def != nil && def.Model != "" {
			if am := getOb3(def.Model); am != nil && am.NumVertices > 0 {
				m := FromAssets(am.NumVertices, am.NumFaces, am.VertexX, am.VertexY, am.VertexZ,
					am.FaceVertices, am.FaceNumVertices, am.FaceFillFront, am.FaceFillBack, am.FaceIntensity)
				// addModels ALWAYS re-lights the placed model (deob World.java:865
				// model.setLight(48,48,-10,magic^9,-50,-50); the live magic=-113 => arg4
				// = -113^9 = -122 ; OpenRSC copy.setDiffuseLight(48,48,-10,-122,-50,-50)).
				// PlaceScenery only orients+translates, so the real path MUST apply the
				// same diffuse light the synthetic leaf does (placeSyntheticDiagObject)
				// or the shading diverges from the DEOB/JAR oracle. Applied to the local
				// geometry before the transform, exactly as the synthetic path does.
				m.setDiffuseLight(48, 48, -10, -122, -50, -50)
				PlaceScenery(m, b.land, b.baseX, b.baseY, b.plane, x, z, def.Width, def.Height, dir, objectID)
				return m
			}
		}
		// Real data source present but the def/model is unknown: skip (no synthetic
		// stand-in for a real-data render — matches placeSceneryModels' silent drop).
		return nil
	}

	// No archive (a hand-authored fixture with no GameData): build a generic wood
	// door-leaf so the diagonal door renders as *something* instead of nothing,
	// placed with the SAME footprint-centre + orient + ground-snap a real model uses.
	return b.placeSyntheticDiagObject(objectID, x, z, dir, w, hgt)
}

// placeSyntheticDiagObject builds a generic wood door-leaf *Model for a diagonal
// object with no real def/model (the fixture path), placed by the IDENTICAL
// PlaceScenery math. It mirrors how render/'s fixture path renders a diagonal door
// without an archive: the leaf is a single upright 1-tile-wide quad coloured with
// the flat WOOD colour (genColorToResource(120,90,55) == the old render package's
// method305(120,90,55) wallColourWood). Centred on the model-local origin so the
// PlaceScenery Orient(0,dir*32,0) spins it about the vertical axis and the
// Translate plants its base at the footprint centre.
func (b *terrainBuilder) placeSyntheticDiagObject(objectID, x, z, dir, w, hgt int) *Model {
	const (
		half       = 64  // half a tile (128 units) wide
		leafHeight = 192 // a full wall/door leaf tall
	)
	woodColour := genColorToResource(120, 90, 55) // == the old render package's method305(120,90,55)

	// A single upright door-leaf quad spanning one tile, centred on the local origin
	// (the leaf extends in X, rises in -Y), so PlaceScenery's vertical-axis orient
	// spins it to face dir and the translate plants its base at the footprint centre.
	m := NewModel(4, 1)
	v0 := m.insertVertex(-half, 0, 0)
	v1 := m.insertVertex(-half, -leafHeight, 0)
	v2 := m.insertVertex(half, -leafHeight, 0)
	v3 := m.insertVertex(half, 0, 0)
	m.insertFace(4, []int{v0, v1, v2, v3}, woodColour, woodColour)

	// The scenery diffuse light addModels applies (deob setLight(48,48,-10,..,-50,-50);
	// OpenRSC copy.setDiffuseLight(48,48,-10,-122,-50,-50)). orsc's setDiffuseLight
	// is the exact analogue: (ambient, diffuse, dirY, guard, dirX, dirZ). The real
	// per-face/vertex shade is (re)computed at project time by resetTransformCache.
	m.setDiffuseLight(48, 48, -10, -122, -50, -50)

	// PlaceScenery's footprint-centre + dir*32 vertical-axis orient + -elevation snap
	// (world.go:659), exactly as for a real model.
	PlaceScenery(m, b.land, b.baseX, b.baseY, b.plane, x, z, w, hgt, dir, objectID)
	return m
}
