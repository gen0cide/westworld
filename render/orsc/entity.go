package orsc

// entity.go — the 2D entity-billboard layer over the 3D scene: players + NPCs.
//
// Faithful to OpenRSC's model: an entity is a SCALED 2D sprite billboard added to
// the Scene (Scene.drawSprite, Scene.java:2532 — registers the m_T model's verts
// + the sprite arrays), depth-sorted WITH the 3D geometry (so an NPC behind a
// wall is occluded by the painter sort), and blitted by drawEntity
// (GraphicsController.drawEntity, GraphicsController.java:524 -> drawSprite). The
// caller composites the entity's pixels (the body-part stack + recolour) and
// hands them in via AddEntity; the Scene positions + blits them.

// entitySprite is one composited billboard's pixels: w×h row-major 0x00RRGGBB,
// with a parallel opaque mask (transparent texels skipped). flip mirrors the
// source horizontally (a W/SW/NW facing built from the mirrored E/SE/NE sprites).
type entitySprite struct {
	pix    []int32
	opaque []bool
	w, h   int
	flip   bool
}

// AddEntity registers a billboard entity at world position (worldX, worldY,
// worldZ) with on-screen world size billboardW×billboardH (the m_ob/m_Eb the
// projection scales by 1/z), and its composited pixels (srcW×srcH). Mirrors
// Scene.drawSprite (Scene.java:2538-2548): set the sprite arrays + insert the
// two-vertex m_T billboard face so endScene projects + depth-sorts it. spriteHead
// doubles as both the sprite-array index and the m_T face index (Java m_n), so
// each call adds exactly one face. No-op if the sprite pool is full.
func (s *Scene) AddEntity(worldX, worldY, worldZ, billboardW, billboardH int32, pix []int32, opaque []bool, srcW, srcH int, flip bool) {
	if s.spriteHead >= len(s.spriteID) || srcW <= 0 || srcH <= 0 || len(pix) < srcW*srcH {
		return
	}
	// Two billboard verts: foot at (x, y, z), top billboardH ABOVE it. Java's
	// drawSprite does insertVertex2(false, z, x, y) then (false, z, x, y - height)
	// — i.e. the height comes off the VERTICAL (Y) axis, not depth (Z). vertY here
	// is -elevation (the terrain's vertical axis), so subtract billboardH from
	// worldY. (Putting it on worldZ — depth — corrupted poly.mT, the sprite's
	// depth-sort key, so terrain/walls sorted over the sprite's lower half on some
	// frames: the "terrain at their thighs" clipping.)
	v0 := s.mT.insertVertex(worldX, worldY, worldZ)
	v1 := s.mT.insertVertex(worldX, worldY-billboardH, worldZ)
	if v0 < 0 || v1 < 0 {
		return
	}
	if s.mT.insertFace(2, []int{v0, v1}, 0, 0) < 0 {
		return
	}
	h := s.spriteHead
	s.spriteID[h] = int32(len(s.entitySprites))
	s.entitySprites = append(s.entitySprites, &entitySprite{pix: pix, opaque: opaque, w: srcW, h: srcH, flip: flip})
	s.spriteTexID[h] = 0
	s.spriteY[h] = worldZ
	s.spriteX[h] = worldX
	s.spriteW[h] = billboardW
	s.spriteH[h] = billboardH
	s.spriteXOff[h] = 0
	s.spriteHead++
}

// drawEntity blits the composited billboard spriteID scaled to w×h at absolute
// screen (x, y) onto the surface, skipping transparent texels and mirroring when
// the composite's flip is set. It is called in painter order (endScene's
// back-to-front fill loop), so nearer 3D geometry drawn afterward correctly
// occludes it — no depth buffer needed. (Replaces the GraphicsController
// drawSprite blit; nearest-neighbour scale.) x already includes m_Zb; the pixel
// index matches the scanline fill's stride (mVb) + absolute x.
func (s *Scene) drawEntity(spriteID, x, y, w, h, scale, flip int32) {
	_ = scale
	_ = flip
	if spriteID < 0 || int(spriteID) >= len(s.entitySprites) || w <= 0 || h <= 0 {
		return
	}
	es := s.entitySprites[spriteID]
	if es == nil || es.w <= 0 || es.h <= 0 {
		return
	}
	surfW, surfH := int32(s.graphics.Width), int32(s.graphics.Height)
	for dy := int32(0); dy < h; dy++ {
		py := y + dy
		if py < 0 || py >= surfH {
			continue
		}
		sy := int(dy) * es.h / int(h)
		rowDst := py * s.mVb
		rowSrc := sy * es.w
		for dx := int32(0); dx < w; dx++ {
			px := x + dx
			if px < 0 || px >= surfW {
				continue
			}
			sx := int(dx) * es.w / int(w)
			if es.flip {
				sx = es.w - 1 - sx
			}
			si := rowSrc + sx
			if si < 0 || si >= len(es.opaque) || !es.opaque[si] {
				continue
			}
			s.pixelData[rowDst+px] = es.pix[si]
		}
	}
}
