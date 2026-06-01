package render

import "github.com/gen0cide/westworld/pathfind"

// hittest.go holds the SINGLE copy of the billboard foot-projection + on-screen
// AABB math that BOTH the renderer (DrawEntitySprites) and the screen-space
// picker (Pick) depend on. The two must agree pixel-for-pixel — a hit box is
// only correct if it equals the rectangle the sprite was actually blitted into —
// so the math lives here once rather than being duplicated and silently drifting
// the moment someone tweaks one side.
//
// Both helpers are pure integer math (no archives, no I/O beyond land.Tile's
// elevation read) and carry no rendering side effects, so they are equally usable
// from a unit test, from DrawEntitySprites' blit loop, and from Pick's hit loop.

// billboardCamera reconstructs the (yaw/roll-swapped) camera a View renders
// with, applying the SAME zoom/W/H defaults buildScene applies, and returns it
// alongside the window origin (baseX, baseY), the screen centre (cx, cy), and the
// flattened terrain-height grid the foot points anchor to.
//
// It is the shared origin both DrawEntitySprites and Pick build their projection
// on, so an under-specified View (Zoom/W/H == 0) projects to the EXACT place it
// rendered. Crucially it applies buildScene's defaults (Zoom 750 / W 512 / H 334
// — render.go:175-183), NOT PickTile's looser 600/512/336: tile picking is robust
// to a small zoom error but a billboard AABB is tight, so the picker must use the
// renderer's own defaults to land on the same pixels.
//
// The yaw/roll axis swap Scene.RenderTo performs before project()
// (render.go:404-406) is applied here EXACTLY ONCE — callers hand the returned
// camera straight to projectBillboard without swapping again. v is taken by
// pointer only to read it (and to default a freshly-copied caller's V); it is not
// mutated for the caller — callers that need the defaulted W/H read them back from
// cx*2 / cy*2 or default their own copy identically.
func billboardCamera(land *pathfind.Landscape, v View) (cam Camera, baseX, baseY, cx, cy int, heights [][]int32) {
	// buildScene defaults (render.go:175-183) — NOT PickTile's 600/336. The picker
	// AABB is tight, so it must reconstruct the camera the renderer actually used.
	if v.Zoom == 0 {
		v.Zoom = 750
	}
	if v.W == 0 {
		v.W = 512
	}
	if v.H == 0 {
		v.H = 334
	}

	n := terrainSize
	baseX = v.X - n/2
	baseY = v.Y - n/2

	// camera at the host tile, identical to buildScene (render.go:373-378): the
	// SelfOffX/Z glide that scrolls the world is folded into the look-at point.
	localX := int32(v.X-baseX)*128 + 64 + int32(v.SelfOffX)
	localZ := int32(v.Y-baseY)*128 + 64 + int32(v.SelfOffZ)
	elev := int32(land.Tile(v.X, v.Y, v.Plane).GroundElevation) * 3
	// setCamera(x, z=height, y, pitch=912, yaw=rotation*4, roll=0, distance)
	cam = SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(v.Zoom)*2)

	// Axis swap at the call site (Scene.RenderTo / DrawEntitySprites entry):
	// project is called with yaw/roll bound swapped. Replicate ONCE so screen
	// placement matches the 3D pass. Callers must NOT swap again.
	cam.CameraYaw, cam.CameraRoll = cam.CameraRoll, cam.CameraYaw

	cx = v.W / 2
	cy = v.H / 2
	heights = TerrainHeights(land, baseX, baseY, v.Plane)
	return cam, baseX, baseY, cx, cy, heights
}

// projectBillboard projects one actor/item foot point and derives its on-screen
// AABB. It is the LITERAL math previously inlined in DrawEntitySprites — the foot
// projection closure (render.go:415-447) followed by the screenW/screenH rect
// block (render.go:539-554) — moved here verbatim so the renderer's blit
// rectangle and the picker's hit box are the same integer expression.
//
// Inputs:
//   - cam        : the yaw/roll-swapped camera from billboardCamera (already swapped).
//   - cx, cy     : the screen centre (v.W/2, v.H/2).
//   - heights    : the flattened terrain grid from billboardCamera.
//   - lx, ly     : the foot tile in WINDOW-LOCAL coords (e.X-baseX, e.Y-baseY).
//   - ox, oz     : sub-tile WORLD-unit glide offsets (Entity.OffX/OffZ; 0 for items).
//   - worldW/H   : the billboard's world-space size (npcBillboardSize / 145×220 /
//     cs.W*groundItemPixelToWorld …).
//
// Returns the screen AABB as [minX, minY, maxX, maxY] in frame pixels — exactly
// the blit destination rectangle (top-left (sx-screenW/2, feetY-screenH), size
// (screenW, screenH)) — plus the foot pixel (sx, feetY) and the UN-biased camera
// depth camZ used for painter/pick ordering. ok=false on exactly the two rejects
// the renderer applies (z < clipNear, or a non-positive screen size), so a box
// exists iff the renderer would have blitted the sprite. The blit's
// spriteDepthBias is a depth-test concern and is deliberately NOT applied here.
func projectBillboard(cam Camera, cx, cy int, heights [][]int32, lx, ly int, ox, oz int32, worldW, worldH int) (rect [4]int, sx, feetY, camZ int32, ok bool) {
	n := terrainSize
	if lx < 0 || lx >= n || ly < 0 || ly >= n {
		return rect, 0, 0, 0, false
	}
	// Foot world point (render.go:419-421): the actor's tile centre at the terrain
	// height there, shifted by the sub-tile glide offset.
	x := (int32(lx)*128 + 64 + ox) - cam.CameraX
	y := -heights[lx][ly] - cam.CameraY
	z := (int32(ly)*128 + 64 + oz) - cam.CameraZ
	// yaw → roll → pitch rotations (render.go:422-442). The camera is already
	// yaw/roll-swapped, so this mirrors GameModel.project verbatim.
	if cam.CameraYaw != 0 {
		ys := sine11[cam.CameraYaw]
		yc := sine11[cam.CameraYaw+1024]
		X := (y*ys + x*yc) >> 15
		y = (y*yc - x*ys) >> 15
		x = X
	}
	if cam.CameraRoll != 0 {
		rs := sine11[cam.CameraRoll]
		rc := sine11[cam.CameraRoll+1024]
		X := (z*rs + x*rc) >> 15
		z = (z*rc - x*rs) >> 15
		x = X
	}
	if cam.CameraPitch != 0 {
		ps := sine11[cam.CameraPitch]
		pc := sine11[cam.CameraPitch+1024]
		Y := (y*pc - z*ps) >> 15
		z = (y*ps + z*pc) >> 15
		y = Y
	}
	if z < clipNear {
		return rect, 0, 0, 0, false
	}
	// screen foot (render.go:446)
	sx = int32(cx) + (x<<uint(viewDist))/z
	feetY = int32(cy) + (y<<uint(viewDist))/z
	camZ = z

	// screen size + rectangle (render.go:539-554). The renderer rejects a
	// non-positive screen size; mirror that so no degenerate box is produced.
	screenW := (int32(worldW) << uint(viewDist)) / camZ
	screenH := (int32(worldH) << uint(viewDist)) / camZ
	if screenW <= 0 || screenH <= 0 {
		return rect, sx, feetY, camZ, false
	}
	left := int(sx) - int(screenW)/2
	top := int(feetY) - int(screenH)
	// right/bottom equal the previous (sx-screenW/2)+screenW and feetY by
	// construction, so the blit's top-left (rect[0], rect[1]) and size
	// (rect[2]-rect[0], rect[3]-rect[1]) reproduce the pre-refactor numbers exactly.
	right := left + int(screenW)
	bottom := int(feetY)
	rect = [4]int{left, top, right, bottom}
	return rect, sx, feetY, camZ, true
}
