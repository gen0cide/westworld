package render

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// diagObjectBand is the DiagonalWalls value range RSC uses to encode a
// diagonally-placed SCENERY object (NOT a wall). The wall builder (boundary.go)
// drops this band; World.addModels realizes it as placed object models.
//
//	getWallDiagonal(x,y) in (48000, 60000)  =>  objectId = value - 48001
//
// (World.java:796-797; OpenRSC World.addLoginScreenModels uses ">48000 && <60000".)
const (
	diagObjectLo = 48000 // exclusive lower bound (48001 is the first valid)
	diagObjectHi = 60000 // exclusive upper bound
	diagObjectID = 48001 // objectId = diagonal value - diagObjectID
)

// BuildDiagonalObjects ports World.addModels (World.java:792-820) — the pass the
// Go render lib never had. RSC encodes a diagonally-PLACED scenery object,
// INCLUDING a diagonal door, as an id in the 48001..59999 band of the per-tile
// DiagonalWalls grid (a marker, not a wall). The boundary builder correctly drops
// that band (boundary.go: `d < 24000`), so without this pass a diagonal door /
// diagonal object renders as ZERO geometry — "a door the cradle can't perceive".
//
// For every tile in the window whose DiagonalWalls value lands in the object
// band, this builds the object's placed model and adds it to the scene, exactly
// like the static scenery pass:
//
//   - objectId = diag - 48001                     (World.java:797)
//   - dir      = getTileDirection(x,y)             (World.java:798) — the new
//     pathfind.Tile.TileDirection grid, threaded through the rscdump schema.
//   - footprint (w,hgt) from the object def's Width/Height, SWAPPED for dir 0/4
//     (World.java:800-801): exactly PlaceScenery's footprint+swap rule.
//   - the model is cloned, translated to the footprint CENTRE
//     cx = 128*(w+2x)/2, cz = 128*(hgt+2y)/2 at -getElevation(cx,cz)
//     (World.java:805-807), and oriented by dir*32 about the vertical axis
//     (World.java:808 `model.g(0,-999999,0,dir*32)` — a roll, not a yaw).
//   - it is lit (World.java:810 `setLight(48,48,-10,..,-50,-50)`).
//
// All of footprint-centre, the dir 0/4 width/height swap, the dir*32 ROLL
// orientation and the bilinear -elevation ground-snap are REUSED verbatim from
// the verified-faithful scenery pass: the real-def path calls PlaceScenery
// directly (passing the object id + tile-direction as a synthesized SceneryLoc),
// so a diagonal object is placed by the identical, already-audited helper. Only
// when no real model is available (a hand-authored fixture with no archive) does
// it fall back to a synthetic wood door-leaf, placed by the SAME math.
//
// The footprint-clear over multi-tile objects (World.java:811-816) is applied to
// a local copy of the diagonal grid so a 2x2+ object is built once (its other
// covered tiles are skipped), without mutating the shared landscape.
//
// f/b mirror the scenery pass's inputs: f carries the object defs, b.Models the
// model archive. With b==nil (the fixture path) the synthetic door-leaf is used.
// Returns the built models to Add to the scene; nil if the window holds no
// diagonal objects.
func BuildDiagonalObjects(mc *ModelCache, f *facts.Facts, land *pathfind.Landscape,
	baseX, baseY, plane int, animFrame int) []*GameModel {

	n := terrainSize

	// A local snapshot of the diagonal-object band over the window, so the
	// footprint-clear (World.java:811-816) can zero a multi-tile object's other
	// covered cells WITHOUT mutating the shared pathfind.Landscape (which the
	// boundary/terrain passes also read). Indexed [lx][ly] in window-local coords;
	// only the object band is captured (walls 1..23999 are irrelevant here).
	diag := make([][]int32, n)
	for lx := 0; lx < n; lx++ {
		diag[lx] = make([]int32, n)
		for ly := 0; ly < n; ly++ {
			d := land.Tile(baseX+lx, baseY+ly, plane).DiagonalWalls
			if d > diagObjectLo && d < diagObjectHi {
				diag[lx][ly] = d
			}
		}
	}

	var out []*GameModel
	// World.addModels scans x (outer) then y (inner); preserve that order so the
	// footprint-clear of a multi-tile object reaches its +x/+y cells before they
	// are visited (World.java:793-794).
	for lx := 0; lx < n; lx++ {
		for ly := 0; ly < n; ly++ {
			d := diag[lx][ly]
			if d <= diagObjectLo || d >= diagObjectHi { // World.java:796 (already filtered, but explicit)
				continue
			}
			objectID := int(d) - diagObjectID // World.java:797

			ax, ay := baseX+lx, baseY+ly
			dir := int(land.Tile(ax, ay, plane).TileDirection) // World.java:798 getTileDirection

			// Footprint (w,hgt). PlaceScenery applies the SAME swap rule
			// (Direction != 0 && != 4 => swap w/hgt), which is exactly the deob's
			// `if (dir != 0 && dir != 4) { w=Width; hgt=Height } else { swap }`
			// (World.java:800-801). So we resolve the un-swapped (Width,Height)
			// footprint from the def and let PlaceScenery do the swap, keeping a
			// single source of truth for the orientation/footprint geometry.
			w, hgt := 1, 1
			if def := defOrNil(f, objectID); def != nil {
				if def.Width > 0 {
					w = def.Width
				}
				if def.Height > 0 {
					hgt = def.Height
				}
			}

			// Place via the verified scenery helper (footprint centre + dir*32 roll
			// + bilinear -elevation ground-snap), routed by a synthesized loc — the
			// object id joins the scenery def space; dir is the per-tile direction.
			// PlaceScenery dereferences f.SceneryDef, so only call it when a real
			// def source exists; with no facts (a hand-authored fixture / no
			// archive) go straight to the synthetic wood door-leaf.
			var g *GameModel
			if f != nil {
				// Real def source: place exactly like the static scenery pass. A
				// missing def/model yields nil and the object is skipped (no
				// synthetic stand-in for a real-data render — matches the static
				// scenery loop, which silently drops an unknown id).
				loc := facts.SceneryLoc{DefID: objectID, X: ax, Y: ay, Direction: dir}
				g = PlaceScenery(mc, f, land, baseX, baseY, plane, loc, animFrame)
			} else {
				// No facts at all (a hand-authored fixture with no GameData/archive):
				// build a generic wood door-leaf with the SAME placement math so the
				// diagonal door renders as *something* instead of nothing — mirroring
				// how syntheticFacts supplies a generic wood boundary def.
				g = placeSyntheticDiagObject(land, baseX, baseY, plane, ax, ay, dir, w, hgt)
			}
			if g != nil {
				out = append(out, g)
			}

			// Footprint-clear (World.java:811-816): for a multi-tile object, zero
			// the diagonal id over the rest of its WxH footprint so the inner tiles
			// don't re-build the same object. dir 0/4 swap the footprint extent the
			// same way PlaceScenery does, so clear over the swapped extent.
			cw, ch := w, hgt
			if dir != 0 && dir != 4 {
				cw, ch = hgt, w // mirror PlaceScenery's w,hgt = hgt,w
			}
			if cw <= 1 && ch <= 1 {
				continue
			}
			for gx := lx; gx < lx+cw && gx < n; gx++ {
				for gy := ly; gy < ly+ch && gy < n; gy++ {
					if (gx > lx || gy > ly) && diag[gx][gy] == d {
						diag[gx][gy] = 0
					}
				}
			}
		}
	}
	return out
}

// defOrNil returns the scenery def for id, or nil if facts is nil / the id is
// unknown.
func defOrNil(f *facts.Facts, id int) *facts.SceneryDef {
	if f == nil {
		return nil
	}
	return f.SceneryDef(id)
}

// placeSyntheticDiagObject builds a generic wood door-leaf GameModel for a
// diagonal object that has no real def/model (a hand-authored fixture with no
// archive), placed with the IDENTICAL footprint-centre + dir*32 roll +
// -elevation ground-snap PlaceScenery uses for a real model. It mirrors how
// syntheticFacts(dump.go) provides a generic wood boundary def so a wall/door
// fixture renders without GameData: here the object renders without models.orsc.
//
// The leaf is a single 1-tile-wide upright quad (a door panel) coloured with the
// flat WOOD colour, the same colour the synthetic boundary door-leaf carries.
func placeSyntheticDiagObject(land *pathfind.Landscape, baseX, baseY, plane, ax, ay, dir, w, hgt int) *GameModel {
	// Footprint width/height swap for any direction other than 0/4 — the same
	// rule PlaceScenery applies (World.java:800-801).
	fw, fh := int32(1), int32(1)
	if w > 0 {
		fw = int32(w)
	}
	if hgt > 0 {
		fh = int32(hgt)
	}
	if dir != 0 && dir != 4 {
		fw, fh = fh, fw
	}

	lx := int32(ax - baseX)
	ly := int32(ay - baseY)
	// Footprint centre (mudclient/World.addModels: cx=128*(w+2x)/2, cz=128*(hgt+2y)/2).
	cx := (lx*2 + fw) * 128 / 2
	cz := (ly*2 + fh) * 128 / 2
	cy := -elevationAt(land, baseX, baseY, cx, cz, plane)

	// A single upright door-leaf quad spanning one tile, centred on the origin so
	// the queued translate places its base at the footprint centre and the dir*32
	// roll spins it about the vertical axis (matching the real model's local
	// origin + setRot256(0, dir*32, 0)).
	const half = 64    // half a tile (128 units) wide
	const height = 192 // a full wall/door leaf tall
	g := NewGameModel(4, 1)
	v0 := g.AddVertex(-half, 0, 0)
	v1 := g.AddVertex(-half, -height, 0)
	v2 := g.AddVertex(half, -height, 0)
	v3 := g.AddVertex(half, 0, 0)
	g.AddFace([]int{v0, v1, v2, v3}, wallColourWood, wallColourWood, 0)

	g.SetLight(48, 48, -50, -10, -50)  // PlaceScenery's scenery light
	g.Rotate(0, 0, int32(dir)*32&0xff) // dir -> ROLL about the vertical axis
	g.Translate(cx, cy, cz)
	return g
}
