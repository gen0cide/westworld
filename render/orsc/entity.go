package orsc

import "github.com/gen0cide/westworld/render"

// entity.go — the 2D entity-billboard layer over the 3D scene: players + NPCs.
//
// Faithful to OpenRSC's model: an entity is a SCALED 2D sprite billboard added to
// the Scene (Scene.drawSprite, Scene.java:2532 — registers the m_T model's verts
// + the sprite arrays), depth-sorted WITH the 3D geometry (so an NPC behind a
// wall is occluded by the painter sort), and blitted by drawEntity
// (GraphicsController.drawEntity, GraphicsController.java:524 -> drawSprite). The
// caller composites the entity's pixels (the body-part stack + recolour) and
// hands them in via AddEntity; the Scene positions + blits them.
//
// PHASE 4 (Milestone C — on-screen 1:1): the integer-NN drawEntity below diverges
// from the authentic rev-235 client under heavy downscale (~3x for the rat) because
// the client blits each body-part layer SEPARATELY through Surface.spriteClipping
// (10-arg), applying each layer's trim in 16.16 fixed-point AT BLIT TIME with
// fractional carry + per-row skew + flip + in-blit dye/skin recolour, whereas this
// path pre-composites all layers at INTEGER trim then whole-canvas integer-NN scales
// (losing the sub-pixel trim + skew). So the parity path now registers the RAW layers
// (AddEntityLayers) and blits each via spriteClippingLayer (a faithful Go port of
// Surface.spriteClipping + transparentSpritePlot). The integer-NN path stays for
// non-parity callers (an entity with no raw layers).

// entitySprite is one billboard's pixel data. Two representations:
//   - layers != nil: the RAW per-layer stack for the faithful 16.16 spriteClipping
//     blit (Phase 4). Each layer carries its own trimmed pixels + full-canvas trim +
//     dye/skin; the on-screen rect is sized from fullW/fullH (the shared figure canvas).
//   - layers == nil: the legacy pre-composited canvas (pix/opaque, w×h) blitted by the
//     whole-canvas integer-NN drawEntity (back-compat for callers that hand a flattened
//     CompositeSprite via AddEntity).
type entitySprite struct {
	pix    []int32
	opaque []bool
	w, h   int
	flip   bool

	// Raw-layer (Phase 4) fields. layers back-to-front; fullW/fullH = the shared
	// figure-canvas dims the billboard rect is sized from (max over layers).
	layers       []render.EntityLayer
	fullW, fullH int
}

// AddEntity registers a billboard entity at world position (worldX, worldY,
// worldZ) with on-screen world size billboardW×billboardH (the m_ob/m_Eb the
// projection scales by 1/z), and its composited pixels (srcW×srcH). Mirrors
// Scene.drawSprite (Scene.java:2538-2548): set the sprite arrays + insert the
// two-vertex m_T billboard face so endScene projects + depth-sorts it. spriteHead
// doubles as both the sprite-array index and the m_T face index (Java m_n), so
// each call adds exactly one face. No-op if the sprite pool is full.
func (s *Scene) AddEntity(worldX, worldY, worldZ, billboardW, billboardH int32, pix []int32, opaque []bool, srcW, srcH int, flip bool) {
	if srcW <= 0 || srcH <= 0 || len(pix) < srcW*srcH {
		return
	}
	s.addBillboard(worldX, worldY, worldZ, billboardW, billboardH, &entitySprite{pix: pix, opaque: opaque, w: srcW, h: srcH, flip: flip})
}

// AddEntityLayers registers a billboard entity at world position (worldX, worldY,
// worldZ) carrying the RAW body-part layers (NOT a pre-composited canvas), so the
// on-screen blit applies each layer's trim in 16.16 fixed-point — the authentic
// rev-235 per-layer path (Phase 4 / Milestone C). billboardW×billboardH is the
// world-space billboard size the projection scales by 1/z (the SAME size the
// integer path uses). The figure-canvas dims (fullW/fullH = max over layers) drive
// the per-layer rect sizing inside spriteClippingLayer. No-op if es is empty.
func (s *Scene) AddEntityLayers(worldX, worldY, worldZ, billboardW, billboardH int32, es *render.EntitySprite) {
	if es == nil || len(es.Layers) == 0 {
		return
	}
	var fullW, fullH int
	for _, l := range es.Layers {
		if l.FullW > fullW {
			fullW = l.FullW
		}
		if l.FullH > fullH {
			fullH = l.FullH
		}
	}
	if fullW <= 0 || fullH <= 0 {
		return
	}
	s.addBillboard(worldX, worldY, worldZ, billboardW, billboardH, &entitySprite{
		layers: es.Layers, fullW: fullW, fullH: fullH, flip: es.Flip,
	})
}

// addBillboard is the shared registration: insert the two-vertex m_T billboard face
// (Scene.drawSprite, Scene.java:2538-2548) + the sprite arrays, and append the pixel
// data. Foot at (x,y,z), top billboardH ABOVE it (off the VERTICAL Y axis, not depth).
func (s *Scene) addBillboard(worldX, worldY, worldZ, billboardW, billboardH int32, es *entitySprite) {
	if s.spriteHead >= len(s.spriteID) {
		return
	}
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
	s.entitySprites = append(s.entitySprites, es)
	s.spriteTexID[h] = 0
	s.spriteY[h] = worldZ
	s.spriteX[h] = worldX
	s.spriteW[h] = billboardW
	s.spriteH[h] = billboardH
	s.spriteXOff[h] = 0
	s.spriteHead++
}

// drawEntity blits the billboard spriteID at the projected rect. For a raw-layer
// entity (Phase 4) it dispatches to the faithful per-layer 16.16 blit
// (drawEntityLayers); for a legacy pre-composited entity it does the whole-canvas
// integer-NN scale (back-compat). It is called in painter order (endScene's
// back-to-front fill loop), so nearer 3D geometry drawn afterward correctly
// occludes it. x already includes m_Zb; the pixel index matches the scanline fill's
// stride (mVb) + absolute x. skew is the per-row horizontal shear (the projected
// billboard's vertex-X delta), passed through to the layer blit.
func (s *Scene) drawEntity(spriteID, x, y, w, h, scale, skew int32) {
	_ = scale
	if spriteID < 0 || int(spriteID) >= len(s.entitySprites) || w <= 0 || h <= 0 {
		return
	}
	es := s.entitySprites[spriteID]
	if es == nil {
		return
	}
	if es.layers != nil {
		s.drawEntityLayers(es, x, y, w, h, skew)
		return
	}
	if es.w <= 0 || es.h <= 0 {
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
