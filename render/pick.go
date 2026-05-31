package render

import "github.com/gen0cide/westworld/pathfind"

// PickTile maps a screen-space click (px,py in the rendered W×H frame) to the
// world TILE the host would walk to. It avoids any fragile inverse-camera math:
// it rebuilds the EXACT camera RenderView used for this View, forward-projects
// every window tile's centre to screen with that camera (the same transform the
// renderer uses), and returns the tile whose projected centre is nearest the
// click (among tiles in front of the camera). Returns plane-LOCAL world coords
// (the caller adds the plane offset) and ok=false if nothing projects near it.
//
// Brute force over the window (~160² tiles) is fine — a click is rare and this
// is far simpler + more robust than unprojecting a ray through non-flat terrain.
func PickTile(land *pathfind.Landscape, v View, px, py int) (worldX, worldY int, ok bool) {
	if land == nil {
		return 0, 0, false
	}
	n := terrainSize
	baseX := v.X - n/2
	baseY := v.Y - n/2

	// Same camera RenderView builds (render.go): host at window centre, pitch
	// 912, yaw rotation*4, distance zoom*2. Camera height = -(host elevation*3).
	localX := int32(v.X-baseX)*128 + 64
	localZ := int32(v.Y-baseY)*128 + 64
	elev := int32(land.Tile(v.X, v.Y, v.Plane).GroundElevation) * 3
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	cam := SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(zoom)*2)
	// RenderTo swaps yaw/roll before project(); replicate so our projection
	// matches the rendered frame pixel-for-pixel.
	cam.CameraYaw, cam.CameraRoll = cam.CameraRoll, cam.CameraYaw

	w, h := v.W, v.H
	if w <= 0 {
		w = 512
	}
	if h <= 0 {
		h = 336
	}
	cx, cy := int32(w/2), int32(h/2)

	bestD := int64(1) << 62
	bestX, bestY := 0, 0
	found := false
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			wx := int32(i)*128 + 64
			wz := int32(j)*128 + 64
			wy := -int32(land.Tile(baseX+i, baseY+j, v.Plane).GroundElevation) * 3
			sx, sy, camZ, vis := projectPoint(cam, wx, wy, wz, cx, cy)
			if !vis || camZ <= clipNear || camZ >= clipFar {
				continue
			}
			dx := int64(int(sx) - px)
			dy := int64(int(sy) - py)
			d := dx*dx + dy*dy
			if d < bestD {
				bestD = d
				bestX, bestY = baseX+i, baseY+j
				found = true
			}
		}
	}
	if !found {
		return 0, 0, false
	}
	return bestX, bestY, true
}

// projectPoint projects one world point (X,Y,Z; Y is height, -up) to screen
// pixels using the (already yaw/roll-swapped) camera, mirroring GameModel.project
// + the screen-centre offset. Returns screen (sx,sy), camera-space depth camZ,
// and vis=false if the point is behind the near plane.
func projectPoint(cam Camera, X, Y, Z, cx, cy int32) (sx, sy, camZ int32, vis bool) {
	x := X - cam.CameraX
	y := Y - cam.CameraY
	z := Z - cam.CameraZ
	if cam.CameraYaw != 0 {
		s := sine11[cam.CameraYaw]
		c := sine11[cam.CameraYaw+1024]
		nx := (y*s + x*c) >> 15
		y = (y*c - x*s) >> 15
		x = nx
	}
	if cam.CameraRoll != 0 {
		s := sine11[cam.CameraRoll]
		c := sine11[cam.CameraRoll+1024]
		nx := (z*s + x*c) >> 15
		z = (z*c - x*s) >> 15
		x = nx
	}
	if cam.CameraPitch != 0 {
		s := sine11[cam.CameraPitch]
		c := sine11[cam.CameraPitch+1024]
		ny := (y*c - z*s) >> 15
		z = (y*s + z*c) >> 15
		y = ny
	}
	if z < clipNear {
		return 0, 0, z, false
	}
	vx := (x << uint(viewDist)) / z
	vy := (y << uint(viewDist)) / z
	// Screen: the renderer plots at centre + viewX, centre + viewY (project's y
	// already carries the world's -up sign, so +vy is correct).
	return cx + vx, cy + vy, z, true
}
