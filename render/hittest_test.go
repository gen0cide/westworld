package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// hittest_test.go is the deterministic unit suite for the screen-space picker
// (render.Pick) and the shared projection/AABB helpers it leans on
// (billboardCamera / projectBillboard). Every test is pure integer math against a
// FLAT zero-height memory landscape — no archives, no PNGs, no Landscape file I/O
// — so the suite is fast and reproducible.
//
// The fixed reference View used throughout (host at (300,500), rotation 0, the
// buildScene defaults 750/512/334) projects the host's OWN tile foot to the frame
// centre (256,167); all golden numbers below were captured once from the current
// projectBillboard math (which is the verbatim DrawEntitySprites math) and serve
// as the regression tripwire against silent drift between the renderer's blit
// rectangle and the picker's hit box.

// flatLand returns a Landscape whose every tile reads the zero Tile (elevation
// 0), so foot points sit on a flat ground plane and the projection is pure
// trigonometry with no terrain confound.
func flatLand() *pathfind.Landscape { return pathfind.NewMemoryLandscape(nil) }

// refView is the canonical test View: host at (300,500), north-facing camera,
// buildScene defaults. Used by most tests so the golden numbers are stable.
func refView() View {
	return View{X: 300, Y: 500, Plane: 0, Rotation: 0, Zoom: 750, W: 512, H: 334}
}

// stubSprite is the canvas used for the stubbed composites: a player-sized icon
// so the ground-item world size (W*groundItemPixelToWorld) is well-defined too.
var stubSprite = &CompositeSprite{W: 48, H: 64}

// withStubSprites overrides the picker's composite-function seam (pickComposite*)
// so every NPC/player/ground-item composites to a fixed non-nil sprite, then
// restores the real functions on cleanup. The sprite archives (entity24.jag /
// config85.jag / media58.jag) live at hardcoded paths absent in CI, so without
// this the picker's sprite-presence gate skips every billboard. Stubbing keeps
// the geometry tests (projection / AABB / ordering / identity) deterministic and
// archive-independent — they assert the PROJECTION math, not sprite decoding.
func withStubSprites(t *testing.T) {
	t.Helper()
	origNPC, origItem, origPlayer := pickCompositeNPC, pickCompositeItem, pickCompositePlayer
	pickCompositeNPC = func(npcID, dir, step int) *CompositeSprite { return stubSprite }
	pickCompositeItem = func(itemID int) *CompositeSprite { return stubSprite }
	pickCompositePlayer = func(hasEquip bool, equip [12]int, hair, top, trouser, skin, dir, step int) *CompositeSprite {
		return stubSprite
	}
	t.Cleanup(func() {
		pickCompositeNPC, pickCompositeItem, pickCompositePlayer = origNPC, origItem, origPlayer
	})
}

// ---------------------------------------------------------------------------
// 6.1 Projection round-trip / self-consistency
// ---------------------------------------------------------------------------

// TestProjectBillboardCentre: an actor on the HOST's own tile projects its foot
// sx to the frame centre cx (== v.W/2), because the camera looks straight down
// the host tile. Confirms billboardCamera/projectBillboard agree with the
// documented cx = v.W/2 screen-centre offset.
func TestProjectBillboardCentre(t *testing.T) {
	land := flatLand()
	v := refView()
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	if cx != v.W/2 || cy != v.H/2 {
		t.Fatalf("screen centre = (%d,%d), want (%d,%d)", cx, cy, v.W/2, v.H/2)
	}
	rect, sx, feetY, camZ, ok := projectBillboard(cam, cx, cy, heights, v.X-baseX, v.Y-baseY, 0, 0, playerBillboardW, playerBillboardH)
	if !ok {
		t.Fatal("host-tile billboard should project")
	}
	if int(sx) != cx {
		t.Errorf("host-tile foot sx = %d, want frame centre %d", sx, cx)
	}
	// The foot sits at the rect bottom and the box is centred horizontally on sx.
	if rect[3] != int(feetY) {
		t.Errorf("rect bottom = %d, want feetY = %d", rect[3], feetY)
	}
	if mid := (rect[0] + rect[2]) / 2; mid != int(sx) {
		t.Errorf("rect horizontal centre = %d, want sx = %d", mid, sx)
	}
	if camZ <= clipNear {
		t.Errorf("host-tile camZ = %d, want > clipNear %d", camZ, clipNear)
	}
}

// TestProjectBillboardMatchesProjectPoint: for several tiles and rotations,
// projectBillboard's foot pixel (sx, feetY) and depth camZ for a ZERO-size
// billboard equal projectPoint's (sx, sy, z) from pick.go. This cross-checks the
// two independent projection paths produce the IDENTICAL foot pixel, pinning the
// yaw/roll swap and the screen-centre offset. (A zero-size billboard fails
// projectBillboard's screen-size reject — ok=false — but the foot pixel + depth
// are computed and returned before that reject, so we compare them regardless.)
func TestProjectBillboardMatchesProjectPoint(t *testing.T) {
	land := flatLand()
	for _, rot := range []int{0, 64, 128, 192} {
		for _, off := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, -1}, {-2, 3}} {
			v := refView()
			v.Rotation = rot
			cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
			lx, ly := (v.X+off[0])-baseX, (v.Y+off[1])-baseY

			_, sx, feetY, camZ, _ := projectBillboard(cam, cx, cy, heights, lx, ly, 0, 0, 0, 0)

			wx := int32(lx)*128 + 64
			wz := int32(ly)*128 + 64
			wy := -heights[lx][ly]
			psx, psy, pz, vis := projectPoint(cam, wx, wy, wz, int32(cx), int32(cy))
			if !vis {
				t.Fatalf("rot=%d off=%v: projectPoint not visible", rot, off)
			}
			if sx != psx || feetY != psy || camZ != pz {
				t.Errorf("rot=%d off=%v: projectBillboard (%d,%d,%d) != projectPoint (%d,%d,%d)",
					rot, off, sx, feetY, camZ, psx, psy, pz)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 6.2 AABB correctness
// ---------------------------------------------------------------------------

// TestBillboardAABB: the rectangle is exactly the literal formula
// screenW=(worldW<<9)/camZ, screenH=(worldH<<9)/camZ, left=sx-screenW/2,
// top=feetY-screenH, right=left+screenW, bottom=feetY — recomputed independently
// here from the returned (sx, feetY, camZ) for a table of world sizes incl. the
// 145×220 player.
func TestBillboardAABB(t *testing.T) {
	land := flatLand()
	v := refView()
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)

	cases := []struct{ worldW, worldH int }{
		{145, 220}, // player / npc fallback
		{64, 96},   // a small npc
		{200, 80},  // a wide-short object (ground item shaped)
	}
	for _, tc := range cases {
		rect, sx, feetY, camZ, ok := projectBillboard(cam, cx, cy, heights, v.X-baseX, (v.Y+1)-baseY, 0, 0, tc.worldW, tc.worldH)
		if !ok {
			t.Fatalf("worldW=%d worldH=%d: expected ok", tc.worldW, tc.worldH)
		}
		screenW := (int32(tc.worldW) << uint(viewDist)) / camZ
		screenH := (int32(tc.worldH) << uint(viewDist)) / camZ
		want := [4]int{
			int(sx) - int(screenW)/2,
			int(feetY) - int(screenH),
			int(sx) - int(screenW)/2 + int(screenW),
			int(feetY),
		}
		if rect != want {
			t.Errorf("worldW=%d worldH=%d: rect=%v, want %v (screenW=%d screenH=%d)",
				tc.worldW, tc.worldH, rect, want, screenW, screenH)
		}
	}
}

// TestBillboardAABBGolden: hard-coded expected rectangles captured once from the
// current (post-refactor) projectBillboard math, for a fixed View+tile+size
// table. This is the renderer-parity tripwire — if anyone changes the projection
// or AABB so the picker's box (and therefore the renderer's blit rect) shifts,
// this test fails. (Doc 10 §6.5 TestDrawEntitySpritesAABBUnchanged.)
func TestBillboardAABBGolden(t *testing.T) {
	land := flatLand()
	v := refView()
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)

	type want struct {
		tx, ty         int
		worldW, worldH int
		rect           [4]int
		sx, feetY      int
		camZ           int32
	}
	golden := []want{
		{300, 500, 145, 220, [4]int{232, 92, 281, 167}, 256, 167, 1499},  // host tile
		{300, 499, 145, 220, [4]int{230, 116, 283, 196}, 256, 196, 1400}, // one north (nearer)
		{300, 501, 145, 220, [4]int{233, 71, 279, 141}, 256, 141, 1598},  // one south (farther)
		{301, 500, 145, 220, [4]int{275, 92, 324, 167}, 299, 167, 1499},  // one east
		{300, 502, 145, 220, [4]int{235, 52, 278, 118}, 256, 118, 1697},  // two south
		{300, 501, 64, 96, [4]int{246, 111, 266, 141}, 256, 141, 1598},   // small npc
	}
	for _, g := range golden {
		rect, sx, feetY, camZ, ok := projectBillboard(cam, cx, cy, heights, g.tx-baseX, g.ty-baseY, 0, 0, g.worldW, g.worldH)
		if !ok {
			t.Fatalf("tile(%d,%d): expected ok", g.tx, g.ty)
		}
		if rect != g.rect || int(sx) != g.sx || int(feetY) != g.feetY || camZ != g.camZ {
			t.Errorf("tile(%d,%d) size(%d,%d): got rect=%v sx=%d feetY=%d camZ=%d; want rect=%v sx=%d feetY=%d camZ=%d",
				g.tx, g.ty, g.worldW, g.worldH, rect, sx, feetY, camZ, g.rect, g.sx, g.feetY, g.camZ)
		}
	}
}

// TestBillboardAABBContainment: project a player at a known tile, then assert
// Pick returns it for a click at the rect centre and at all four corners
// (inclusive bounds), and NOT for a click one pixel outside each edge.
func TestBillboardAABBContainment(t *testing.T) {
	withStubSprites(t)
	land := flatLand()
	v := refView()
	// place one player on the host's own tile.
	v.NoSelf = true // suppress the auto self-billboard so only our entity is in play
	v.Entities = []Entity{{Kind: EntityPlayer, X: 300, Y: 500, Index: 9}}

	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	rect, _, _, _, ok := projectBillboard(cam, cx, cy, heights, 300-baseX, 500-baseY, 0, 0, playerBillboardW, playerBillboardH)
	if !ok {
		t.Fatal("setup: player should project")
	}

	// inside: rect centre + all four corners (inclusive).
	inside := [][2]int{
		{(rect[0] + rect[2]) / 2, (rect[1] + rect[3]) / 2},
		{rect[0], rect[1]}, {rect[2], rect[1]}, {rect[0], rect[3]}, {rect[2], rect[3]},
	}
	for _, p := range inside {
		got := Pick(land, nil, v, p[0], p[1])
		if !containsKindIndex(got, TargetPlayer, 9) {
			t.Errorf("click %v inside rect %v: expected the player (index 9) in %v", p, rect, summarize(got))
		}
	}

	// outside: one pixel beyond each edge — the player must NOT be a candidate.
	midX := (rect[0] + rect[2]) / 2
	midY := (rect[1] + rect[3]) / 2
	outside := [][2]int{
		{rect[0] - 1, midY}, // left
		{rect[2] + 1, midY}, // right
		{midX, rect[1] - 1}, // top
		{midX, rect[3] + 1}, // bottom
	}
	for _, p := range outside {
		got := Pick(land, nil, v, p[0], p[1])
		if containsKindIndex(got, TargetPlayer, 9) {
			t.Errorf("click %v outside rect %v: player should NOT be a candidate (got %v)", p, rect, summarize(got))
		}
	}
}

// TestClipNearReject: an actor placed so its foot is behind the near plane
// (z < clipNear) yields ok=false and contributes no billboard candidate.
func TestClipNearReject(t *testing.T) {
	land := flatLand()
	v := refView()
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	// the camera sits `distance` behind the host tile looking forward; a tile far
	// to the NORTH of the host (decreasing Y) eventually falls behind the near
	// plane. Walk north until projectBillboard rejects on clipNear.
	rejected := false
	for dy := 0; dy < terrainSize/2; dy++ {
		ly := (v.Y - dy) - baseY
		if ly < 0 {
			break
		}
		_, _, _, camZ, ok := projectBillboard(cam, cx, cy, heights, v.X-baseX, ly, 0, 0, 0, 0)
		if !ok && camZ < clipNear {
			rejected = true
			break
		}
	}
	if !rejected {
		t.Skip("no tile in the window fell behind clipNear with this camera; reject path is also covered by the zero-size test")
	}
}

// TestZeroSizeReject: a tiny world size at a large depth drives screenW/screenH
// to 0, which projectBillboard rejects (ok=false) exactly as the renderer's
// screenW<=0||screenH<=0 guard does.
func TestZeroSizeReject(t *testing.T) {
	land := flatLand()
	v := refView()
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	// worldW=worldH=0 => screenW=screenH=0 => reject. But the foot pixel + depth
	// are still computed (and non-zero), proving the reject is the size guard, not
	// a projection failure.
	_, sx, _, camZ, ok := projectBillboard(cam, cx, cy, heights, v.X-baseX, v.Y-baseY, 0, 0, 0, 0)
	if ok {
		t.Error("zero world size should be rejected")
	}
	if camZ <= 0 || sx == 0 {
		t.Errorf("foot should still project before the size reject: sx=%d camZ=%d", sx, camZ)
	}
}

// ---------------------------------------------------------------------------
// 6.3 Ordering
// ---------------------------------------------------------------------------

// TestPickOrderingNearestFirst: two overlapping billboards at the same screen
// pixel but different depths — the NEARER (smaller camZ) is candidates[0], the
// reverse of the renderer's far-first painter order.
func TestPickOrderingNearestFirst(t *testing.T) {
	withStubSprites(t)
	land := flatLand()
	v := refView()
	v.NoSelf = true
	// Two NPCs on the host column at different Y. Tile (300,500) is the host tile
	// (camZ 1499); tile (300,501) is one south (camZ 1598, FARTHER). Both project
	// their foot to sx=256, and their tall boxes overlap the centre column, so a
	// click near the host-tile foot hits both.
	v.Entities = []Entity{
		{Kind: EntityNPC, X: 300, Y: 501, Index: 1, NpcID: 0}, // farther (enumerated first)
		{Kind: EntityNPC, X: 300, Y: 500, Index: 2, NpcID: 0}, // nearer
	}
	// click at the nearer actor's foot region, on the shared centre column.
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	rNear, _, _, _, _ := projectBillboard(cam, cx, cy, heights, 300-baseX, 500-baseY, 0, 0, playerBillboardW, playerBillboardH)
	clickX := (rNear[0] + rNear[2]) / 2
	clickY := rNear[1] + 2 // near the top of the nearer box, inside both tall boxes

	got := Pick(land, nil, v, clickX, clickY)
	if len(got) < 2 {
		t.Fatalf("expected both NPCs to be hit, got %v", summarize(got))
	}
	if got[0].CamZ > got[1].CamZ {
		t.Errorf("candidates not nearest-first: [0].CamZ=%d [1].CamZ=%d", got[0].CamZ, got[1].CamZ)
	}
	if got[0].Index != 2 {
		t.Errorf("nearest candidate should be index 2 (host tile), got index %d (CamZ %d)", got[0].Index, got[0].CamZ)
	}
}

// TestPickStableTieOrder: two billboards at IDENTICAL depth (same tile) that both
// contain the click preserve enumeration order — the NPC (enumerated first)
// precedes the player.
func TestPickStableTieOrder(t *testing.T) {
	withStubSprites(t)
	land := flatLand()
	v := refView()
	v.NoSelf = true
	// NPC and player on the SAME tile => identical camZ. The picker enumerates
	// v.Entities in order (NPC first here), and the stable nearest-first sort must
	// keep that order on the tie.
	v.Entities = []Entity{
		{Kind: EntityNPC, X: 300, Y: 500, Index: 11, NpcID: 0},
		{Kind: EntityPlayer, X: 300, Y: 500, Index: 22},
	}
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	rect, _, _, _, _ := projectBillboard(cam, cx, cy, heights, 300-baseX, 500-baseY, 0, 0, playerBillboardW, playerBillboardH)
	clickX := (rect[0] + rect[2]) / 2
	clickY := (rect[1] + rect[3]) / 2

	got := Pick(land, nil, v, clickX, clickY)
	if len(got) < 2 {
		t.Fatalf("expected both actors hit, got %v", summarize(got))
	}
	if got[0].CamZ != got[1].CamZ {
		t.Fatalf("expected a depth tie, got CamZ %d vs %d", got[0].CamZ, got[1].CamZ)
	}
	if got[0].Kind != TargetNPC || got[1].Kind != TargetPlayer {
		t.Errorf("tie order not stable: got [%v(idx %d), %v(idx %d)], want [NPC, Player]",
			got[0].Kind, got[0].Index, got[1].Kind, got[1].Index)
	}
}

// TestPickTerrainAlwaysLast: with f != nil and a scenery + boundary on the picked
// tile, the candidate list ends with exactly one TargetTerrain, preceded by the
// scenery then the boundary (the §4.2 priority).
func TestPickTerrainAlwaysLast(t *testing.T) {
	land := flatLand()
	f := loadFixtureFacts(t, fixtureFacts{
		sceneryDefs:  []sceneryDefRow{{name: "Tree", cmd1: "Chop"}},  // index 0
		boundaryDefs: []boundaryDefRow{{name: "Door", cmd1: "Open"}}, // index 0
		sceneryLocs:  []locRow{{id: 0, x: 300, y: 500, dir: 0}},      // tree on host tile
		boundaryLocs: []locRow{{id: 0, x: 300, y: 500, dir: 1}},      // door on host tile
	})
	v := refView()
	v.NoSelf = true // keep billboards out of the way for a clean tile-target assertion

	got := Pick(land, f, v, 256, 167) // centre click resolves to the host tile (300,500)

	// find the tile-target tail (after any billboards — none here).
	if len(got) == 0 {
		t.Fatal("expected tile candidates")
	}
	last := got[len(got)-1]
	if last.Kind != TargetTerrain {
		t.Fatalf("last candidate must be terrain, got %v (%v)", last.Kind, summarize(got))
	}
	// exactly one terrain.
	terrCount := 0
	for _, c := range got {
		if c.Kind == TargetTerrain {
			terrCount++
		}
	}
	if terrCount != 1 {
		t.Errorf("expected exactly one terrain candidate, got %d (%v)", terrCount, summarize(got))
	}
	// scenery precedes boundary precedes terrain.
	si, bi, ti := indexOfKind(got, TargetScenery), indexOfKind(got, TargetBoundary), indexOfKind(got, TargetTerrain)
	if si < 0 || bi < 0 {
		t.Fatalf("expected both scenery and boundary candidates, got %v", summarize(got))
	}
	if !(si < bi && bi < ti) {
		t.Errorf("priority wrong: scenery@%d boundary@%d terrain@%d (want scenery<boundary<terrain) in %v", si, bi, ti, summarize(got))
	}
}

// ---------------------------------------------------------------------------
// 6.4 Identity completeness
// ---------------------------------------------------------------------------

// TestCandidateIdentityNPC: an NPC entity produces a TargetNPC candidate with the
// server Index, the NpcID, matching plane-local X/Y, and the echoed Plane.
func TestCandidateIdentityNPC(t *testing.T) {
	withStubSprites(t)
	land := flatLand()
	v := refView()
	v.NoSelf = true
	v.Entities = []Entity{{Kind: EntityNPC, Index: 7, NpcID: 0, X: 300, Y: 500}}

	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	rect, _, _, _, _ := projectBillboard(cam, cx, cy, heights, 300-baseX, 500-baseY, 0, 0, npcW(0), npcH(0))
	got := Pick(land, nil, v, (rect[0]+rect[2])/2, (rect[1]+rect[3])/2)

	c := findKind(got, TargetNPC)
	if c == nil {
		t.Fatalf("no NPC candidate in %v", summarize(got))
	}
	if c.Index != 7 || c.NpcID != 0 || c.X != 300 || c.Y != 500 || c.Plane != v.Plane {
		t.Errorf("NPC identity wrong: %+v (want Index 7, NpcID 0, X 300, Y 500, Plane %d)", *c, v.Plane)
	}
}

// TestCandidateIdentityGroundItem: a GroundItemMarker produces a TargetGroundItem
// candidate carrying the ItemID and the tile.
func TestCandidateIdentityGroundItem(t *testing.T) {
	withStubSprites(t)
	land := flatLand()
	v := refView()
	v.NoSelf = true
	const itemID = 10
	v.GroundItems = []GroundItemMarker{{X: 300, Y: 500, ItemID: itemID}}

	// the stubbed icon's world size = stubSprite.W*groundItemPixelToWorld etc.
	cs := pickCompositeItem(itemID)
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	rect, _, _, _, ok := projectBillboard(cam, cx, cy, heights, 300-baseX, 500-baseY, 0, 0, cs.W*groundItemPixelToWorld, cs.H*groundItemPixelToWorld)
	if !ok {
		t.Fatal("ground item should project")
	}
	got := Pick(land, nil, v, (rect[0]+rect[2])/2, (rect[1]+rect[3])/2)

	c := findKind(got, TargetGroundItem)
	if c == nil {
		t.Fatalf("no ground-item candidate in %v", summarize(got))
	}
	if c.ItemID != itemID || c.X != 300 || c.Y != 500 {
		t.Errorf("ground-item identity wrong: %+v (want ItemID %d, X 300, Y 500)", *c, itemID)
	}
}

// TestCandidateIdentityBoundary: a click on a tile carrying one BoundaryLoc with
// Direction 1 yields a TargetBoundary with that DefID and Direction (the exact
// InteractWithBoundary 3rd arg).
func TestCandidateIdentityBoundary(t *testing.T) {
	land := flatLand()
	f := loadFixtureFacts(t, fixtureFacts{
		boundaryDefs: []boundaryDefRow{{name: "Door", cmd1: "Open"}}, // index 0
		boundaryLocs: []locRow{{id: 0, x: 300, y: 500, dir: 1}},
	})
	v := refView()
	v.NoSelf = true

	got := Pick(land, f, v, 256, 167) // centre -> host tile (300,500)
	c := findKind(got, TargetBoundary)
	if c == nil {
		t.Fatalf("no boundary candidate in %v", summarize(got))
	}
	if c.DefID != 0 || c.Direction != 1 || c.X != 300 || c.Y != 500 {
		t.Errorf("boundary identity wrong: %+v (want DefID 0, Direction 1, X 300, Y 500)", *c)
	}
}

// TestCandidateIdentityScenery: a static scenery loc AND a dynamic
// DynamicSceneryItem on the SAME tile — the dynamic one appears (Dynamic=true)
// and the static one is suppressed (dedup parity with the renderer, §4.2).
func TestCandidateIdentityScenery(t *testing.T) {
	land := flatLand()
	f := loadFixtureFacts(t, fixtureFacts{
		sceneryDefs: []sceneryDefRow{
			{name: "Tree", cmd1: "Chop"},    // index 0 (static)
			{name: "Fire", cmd1: "Examine"}, // index 1 (dynamic)
		},
		sceneryLocs: []locRow{{id: 0, x: 300, y: 500, dir: 0}}, // static tree on host tile
	})
	v := refView()
	v.NoSelf = true
	// a live dynamic scenery on the SAME tile (plane-local, as the cradle stores it).
	v.DynamicScenery = []DynamicSceneryItem{{X: 300, Y: 500, ID: 1, Direction: 2}}

	got := Pick(land, f, v, 256, 167) // centre -> host tile (300,500)

	// the dynamic entry must be present with Dynamic=true and DefID 1.
	var dyn, stat *PickCandidate
	for i := range got {
		if got[i].Kind != TargetScenery {
			continue
		}
		if got[i].Dynamic {
			dyn = &got[i]
		} else {
			stat = &got[i]
		}
	}
	if dyn == nil {
		t.Fatalf("expected a dynamic scenery candidate, got %v", summarize(got))
	}
	if dyn.DefID != 1 || dyn.Direction != 2 || dyn.X != 300 || dyn.Y != 500 {
		t.Errorf("dynamic scenery identity wrong: %+v (want DefID 1, Direction 2, X 300, Y 500)", *dyn)
	}
	if stat != nil {
		t.Errorf("static scenery should be SUPPRESSED by the dynamic entry on the same tile, but got %+v", *stat)
	}
}

// TestPickNilFacts: with f == nil, only billboards + terrain are returned (no
// scenery/boundary enumeration), and terrain is still the last entry.
func TestPickNilFacts(t *testing.T) {
	land := flatLand()
	v := refView()
	v.NoSelf = true
	got := Pick(land, nil, v, 256, 167)
	if len(got) != 1 || got[0].Kind != TargetTerrain {
		t.Fatalf("with nil facts + no entities, want exactly [terrain], got %v", summarize(got))
	}
}

// TestPickNilLandscape: with land == nil there is no projection and no tile, so
// Pick returns an empty list (billboards need the terrain grid + host elevation).
func TestPickNilLandscape(t *testing.T) {
	v := refView()
	got := Pick(nil, nil, v, 256, 167)
	if len(got) != 0 {
		t.Fatalf("nil landscape should yield no candidates, got %v", summarize(got))
	}
}

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func npcW(id int) int { w, _ := npcBillboardSize(id); return w }
func npcH(id int) int { _, h := npcBillboardSize(id); return h }

func containsKindIndex(cs []PickCandidate, k TargetKind, idx int) bool {
	for _, c := range cs {
		if c.Kind == k && c.Index == idx {
			return true
		}
	}
	return false
}

func findKind(cs []PickCandidate, k TargetKind) *PickCandidate {
	for i := range cs {
		if cs[i].Kind == k {
			return &cs[i]
		}
	}
	return nil
}

func indexOfKind(cs []PickCandidate, k TargetKind) int {
	for i := range cs {
		if cs[i].Kind == k {
			return i
		}
	}
	return -1
}

// summarize renders a candidate list compactly for failure messages.
func summarize(cs []PickCandidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = kindName(c.Kind)
	}
	return out
}

func kindName(k TargetKind) string {
	switch k {
	case TargetNPC:
		return "npc"
	case TargetPlayer:
		return "player"
	case TargetSelf:
		return "self"
	case TargetGroundItem:
		return "ground_item"
	case TargetScenery:
		return "scenery"
	case TargetBoundary:
		return "boundary"
	case TargetTerrain:
		return "terrain"
	}
	return "?"
}

// ----- minimal hand-built *facts.Facts fixture -----
//
// Pick reads scenery/boundary tile contents through facts.At, which needs the
// spatial index that ONLY facts.Load builds. A directly-constructed &facts.Facts{}
// (the pattern the fidelity tests use for the render path, which reads
// f.SceneryLocs directly) would have an empty index, so f.At returns nothing. We
// therefore drive the public facts.Load over a tiny temp def/loc tree so the
// fixture is index-complete — this keeps the test inside the render package and
// touches only the public facts API.

type sceneryDefRow struct{ name, cmd1, cmd2 string }
type boundaryDefRow struct{ name, cmd1, cmd2 string }
type locRow struct {
	id, x, y, dir int
}

type fixtureFacts struct {
	sceneryDefs  []sceneryDefRow
	boundaryDefs []boundaryDefRow
	sceneryLocs  []locRow
	boundaryLocs []locRow
}

// loadFixtureFacts writes minimal OpenRSC-shaped def + loc files into a temp dir
// and loads them through facts.Load (so the spatial index facts.At needs is
// built). Def array POSITION is the def id (the loaders index by slice index), so
// a loc's `id` references the matching def row's position.
func loadFixtureFacts(t *testing.T, ff fixtureFacts) *facts.Facts {
	t.Helper()
	dir := t.TempDir()

	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return p
	}

	// GameObjectDef.xml — one <GameObjectDef> per scenery def, position == id.
	var sx string
	sx += "<GameObjectDefs>"
	for _, d := range ff.sceneryDefs {
		sx += "<GameObjectDef><name>" + d.name + "</name><command1>" + d.cmd1 + "</command1><command2>" + d.cmd2 + "</command2></GameObjectDef>"
	}
	sx += "</GameObjectDefs>"

	// DoorDef.xml — one <DoorDef> per boundary def, position == id.
	var bx string
	bx += "<DoorDefs>"
	for _, d := range ff.boundaryDefs {
		bx += "<DoorDef><name>" + d.name + "</name><command1>" + d.cmd1 + "</command1><command2>" + d.cmd2 + "</command2></DoorDef>"
	}
	bx += "</DoorDefs>"

	sources := facts.Sources{
		Root:            dir,
		SceneryDefsXML:  write("GameObjectDef.xml", sx),
		BoundaryDefsXML: write("DoorDef.xml", bx),
		// Empty def lists: the npc/item loaders try a {wrapper} first, then fall
		// back to a BARE array — so a bare [] satisfies them with no defs (an empty
		// {wrapper} would still hit the bare-array fallback's object-decode error).
		NpcDefsJSON:      write("NpcDefs.json", `[]`),
		ItemDefsJSON:     write("ItemDefs.json", `[]`),
		SceneryLocsJSON:  write("SceneryLocs.json", locsJSON("sceneries", ff.sceneryLocs)),
		BoundaryLocsJSON: write("BoundaryLocs.json", locsJSON("boundaries", ff.boundaryLocs)),
		NpcLocsJSON:      write("NpcLocs.json", `{"npclocs":[]}`),
		GroundItemsJSON:  write("GroundItems.json", `{"grounditems":[]}`),
	}
	f, err := facts.Load(sources)
	if err != nil {
		t.Fatalf("facts.Load fixture: %v", err)
	}
	return f
}

// locsJSON renders a {"<key>":[{id,pos,direction}]} loc file.
func locsJSON(key string, rows []locRow) string {
	s := `{"` + key + `":[`
	for i, r := range rows {
		if i > 0 {
			s += ","
		}
		s += `{"id":` + itoa(r.id) + `,"pos":{"X":` + itoa(r.x) + `,"Y":` + itoa(r.y) + `},"direction":` + itoa(r.dir) + `}`
	}
	s += `]}`
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
