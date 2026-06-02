package orsc

import (
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

// pick.go — screen-click -> world-tile for the orsc view. The cradle's /walk maps
// a click to the tile the host walks to; that must use the SAME camera + projection
// the displayed frame used, or clicks land on a neighbouring tile (worse toward the
// frame edges, where the three/ and classic projections diverge). render.PickTile
// uses the classic render/ camera; this is its orsc twin, so clicks on an orsc
// frame resolve exactly.

// projectPoint replicates model.project's per-vertex camera transform + perspective
// divide (model.go:771-803) for a single WORLD point. The scene models carry no
// model translate (vertices are absolute window coords), so vertXTransform == the
// world coord and we skip the bake. Returns the projection-centre-relative screen
// x/y (add mZb / mNb for pixels) and the camera-space Z (depth; > rot1024ZTop means
// in front of the camera).
func (s *Scene) projectPoint(wx, wy, wz int32) (sx, sy, camZ int32) {
	x0 := wx - s.cam.rot1024OffX
	yO := wy - s.cam.rot1024OffY
	zO := wz - s.cam.rot1024OffZ
	rotZ, rotY, rotX := s.cam.cameraProjZ, s.cam.cameraProjY, s.cam.cameraProjX

	if rotZ != 0 { // xy rotation (model.go:777-781)
		xyXy := trigTable1024[rotZ]
		xyYy := trigTable1024[1024+rotZ]
		tmp := (yO*xyXy + xyYy*x0) >> 15
		yO = (yO*xyYy - x0*xyXy) >> 15
		x0 = tmp
	}
	if rotY != 0 { // xz rotation (model.go:782-786)
		xzXz := trigTable1024[rotY]
		xzXx := trigTable1024[rotY+1024]
		tmp := (xzXx*x0 + zO*xzXz) >> 15
		zO = (xzXx*zO - x0*xzXz) >> 15
		x0 = tmp
	}
	if rotX != 0 { // yz rotation (model.go:787-791)
		yzZy := trigTable1024[rotX]
		yzZz := trigTable1024[rotX+1024]
		tmp := (yO*yzZz - yzZy*zO) >> 15
		zO = (yzZy*yO + yzZz*zO) >> 15
		yO = tmp
	}

	vp := uint(s.rot1024VpSrc)
	if zO < rot1024ZTop { // model.go:793-803
		return x0 << vp, yO << vp, zO
	}
	return (x0 << vp) / zO, (yO << vp) / zO, zO
}

// PickTile maps a screen click (px,py) in the orsc W×H frame to the world tile the
// host would walk to — the orsc analogue of render.PickTile. It rebuilds the EXACT
// camera RenderView used (host at the window centre, pitch cameraPitch, yaw
// Rotation*4, distance Zoom*2), forward-projects every window tile centre with the
// orsc projection, and returns the tile whose projected centre is nearest the click
// (among tiles in front of the camera). Coords follow render.PickTile's convention:
// v.Y is plane-LOCAL and the returned worldY is plane-local too (the caller adds the
// plane offset). ok=false if nothing projects near the click.
func PickTile(land *pathfind.Landscape, v render.View, px, py int) (worldX, worldY int, ok bool) {
	if land == nil {
		return 0, 0, false
	}
	w, h := v.W, v.H
	if w <= 0 {
		w = 512
	}
	if h <= 0 {
		h = 334
	}
	baseX := v.X - windowCentreTile
	baseY := v.Y - windowCentreTile
	plane := v.Plane

	// Same camera RenderView builds (view.go RenderView): centred on the host tile,
	// height = -ground elevation there, the spectator pitch, yaw = Rotation*4,
	// orbit distance = Zoom*2.
	surf := NewSurface(w, h)
	scene := NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)
	cx := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2)
	cz := cx
	elev := elevationOf(land, baseX, baseY, plane, cx, cz)
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	scene.fogLandscapeDistance = fogLandscape // far-Z cutoff for the in-front test
	scene.fogEntityDistance = fogLandscape
	scene.SetCamera(cx, -elev, cz, cameraPitch, int32(v.Rotation)*4, 0, int32(zoom)*2)

	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}
	bestD := int64(1) << 62
	bx, by := 0, 0
	found := false
	for i := 0; i < worldWindowTiles; i++ {
		for j := 0; j < worldWindowTiles; j++ {
			wx := int32(i*tileWorldUnits + tileWorldUnits/2)
			wz := int32(j*tileWorldUnits + tileWorldUnits/2)
			wy := -b.getElevation(wx, wz) // bilinear ground at the tile centre (the camera/terrain surface)
			sx, sy, camZ := scene.projectPoint(wx, wy, wz)
			if camZ <= rot1024ZTop || camZ >= scene.fogLandscapeDistance {
				continue
			}
			scrX := int(sx + scene.mZb)
			scrY := int(sy + scene.mNb)
			dx := int64(scrX - px)
			dy := int64(scrY - py)
			d := dx*dx + dy*dy
			if d < bestD {
				bestD = d
				bx = baseX + i
				by = baseY + j
				found = true
			}
		}
	}
	if !found {
		return 0, 0, false
	}
	return bx, by, true
}
