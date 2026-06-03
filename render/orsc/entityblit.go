package orsc

import "github.com/gen0cide/westworld/render"

// entityblit.go — the faithful 16.16 per-layer entity-sprite blit (NPC/player
// 2D-sprite parity, Phase 4 / Milestone C, docs/build/NPC_SPRITE_PARITY_PLAN.md
// tool #2). A direct, line-for-line Go port of the authentic rev-235
// Surface.spriteClipping (10-arg, the per-body-part scaled+skewed+two-tint blit) and
// Surface.transparentSpritePlot (the inner per-row transparent plot with in-blit
// dye/skin recolour), from reference/rsc-client/src/client/scene/Surface.java:1825
// (spriteClipping) / :1935-2125 (the four transparentSpritePlot variants).
//
// Why a per-layer blit (not the pre-composited integer-NN drawEntity): under the
// fixture camera the rat billboard downscales ~3x (346->~118px). The authentic
// client blits each body-part layer SEPARATELY, applying that layer's trim
// (spriteTranslate) in 16.16 fixed-point AT BLIT TIME with fractional carry (u0/v0),
// so each layer samples its own sub-pixel source columns; the pre-composite-then-
// scale path collapses that to integer offsets and loses it. This port reproduces
// the exact source-column sampling, the per-row skew, the flip, and the dye/skin
// recolour rule — so the orsc on-screen render matches the DEOB/JAR byte-for-byte.

// drawEntityLayers blits each raw layer of a billboard back-to-front through the
// faithful 16.16 spriteClipping port. (x, y) is the projected rect origin (already
// including mZb), w/h the projected billboard dims, skew the per-row horizontal shear
// (the projected billboard's vertex-X delta). The clip rectangle is the full surface
// [0,W)x[0,H) — the SAME as the DEOB Surface ctor's boundsTop/BottomX/Y (the entity
// blit there is unclipped beyond the surface).
func (s *Scene) drawEntityLayers(es *entitySprite, x, y, w, h, skew int32) {
	flip := es.flip
	for i := range es.layers {
		s.spriteClippingLayer(&es.layers[i], int(x), int(y), int(w), int(h), flip, int(skew))
	}
}

// spriteClippingLayer is the Go port of Surface.spriteClipping(x,y,w,flip,h,sprite,
// colour1,colour2,skew,dummy) (Surface.java:1825) for ONE entity body-part layer.
// colour1/colour2 are the layer's Dye/Skin (0 => 0xffffff identity). It computes the
// 16.16 trim (spriteTranslate is always set for entity sprites: a full-canvas figure
// with a trimmed sub-rect) then dispatches to the inner per-row plot. The clip
// rectangle is the full surface (boundsTop=0, boundsBottom=W/H).
func (s *Scene) spriteClippingLayer(l *render.EntityLayer, x, y, w, h int, flip bool, skew int) {
	colour1 := l.Dye
	colour2 := l.Skin
	if colour1 == 0 {
		colour1 = 0xffffff
	}
	if colour2 == 0 {
		colour2 = 0xffffff
	}
	sw := l.W
	sh := l.H
	if sw <= 0 || sh <= 0 || w <= 0 || h <= 0 || len(l.Pix) < sw*sh {
		return
	}
	width := s.graphics.Width  // == Surface.width (row stride)
	height := s.graphics.Height
	boundsTopX, boundsTopY := 0, 0
	boundsBottomX, boundsBottomY := width, height

	u0, v0 := 0, 0
	skewAcc := skew << 16
	stepU := (sw << 16) / w
	stepV := (sh << 16) / h
	skewStep := -(skew << 16) / h

	// spriteTranslate branch (Surface.java:1841-1866): the entity sprite is a trimmed
	// sub-rect of a full FullW x FullH figure canvas placed at (TX, TY). Apply the trim
	// in 16.16 with the fractional carry u0/v0 that recovers the sub-pixel start column.
	fw := l.FullW
	fh := l.FullH
	if fw == 0 || fh == 0 {
		return // clean base guards the full-size divisions
	}
	stepU = (fw << 16) / w
	stepV = (fh << 16) / h
	tx := l.TX
	ty := l.TY
	if flip {
		tx = fw - l.W - tx
	}
	x += (tx*w + fw - 1) / fw
	tyOff := (ty*h + fh - 1) / fh
	y += tyOff
	skewAcc += tyOff * skewStep
	if (tx*w)%fw != 0 {
		u0 = ((fw - (tx*w)%fw) << 16) / w
	}
	if (ty*h)%fh != 0 {
		v0 = ((fh - (ty*h)%fh) << 16) / h
	}
	w = (((l.W << 16) - u0) + stepU - 1) / stepU
	h = (((l.H << 16) - v0) + stepV - 1) / stepV
	if w <= 0 || h <= 0 {
		return
	}

	dst := y * width
	skewAcc += x << 16
	if y < boundsTopY {
		cut := boundsTopY - y
		h -= cut
		y = boundsTopY
		dst += cut * width
		v0 += stepV * cut
		skewAcc += skewStep * cut
	}
	if y+h >= boundsBottomY {
		h -= (y + h) - boundsBottomY + 1
	}
	if h <= 0 {
		return
	}
	parity := (dst / width) & 1
	if !s.interlace {
		parity = 2 // disables the interlace skip
	}

	// Dispatch (Surface.java:1884-1910). Entity layers always carry direct RGB pixels
	// (the content1 decode -> int32 0x00RRGGBB), so the spritePixels (not palette-index)
	// variant is the path. flip negates the U step and starts U at the mirrored column.
	uStart := u0
	step := stepU
	if flip {
		uStart = (l.W << 16) - u0 - 1
		step = -stepU
	}
	if colour2 == 0xffffff {
		s.transparentSpritePlot1(l.Pix, uStart, v0, dst, w, h, step, stepV, sw, colour1, skewAcc, skewStep, parity,
			width, boundsTopX, boundsBottomX)
	} else {
		s.transparentSpritePlot2(l.Pix, uStart, v0, dst, w, h, step, stepV, sw, colour1, colour2, skewAcc, skewStep, parity,
			width, boundsTopX, boundsBottomX)
	}
}

// transparentSpritePlot1 is the Go port of the single-tint direct-pixel
// transparentSpritePlot (Surface.java:1935): for each row re-clip the skewed span
// against [boundsTopX, boundsBottomX), then per column sample src[(u>>16)+rowBase],
// skip the transparent sentinel, recolour a grey texel (r==g==b) by tint, else copy.
// src is the layer's 0x00RRGGBB pixels with -1 = transparent (the per-sprite palette
// index-0 key) — the != 0 transparency test in the Java reads the BLACK-keyed direct
// pixels; here the decode already keyed transparency to -1, so we skip < 0 (and 0
// stays a legitimate black texel, matching the content1 decode where index 0 is the
// ONLY transparent key).
func (s *Scene) transparentSpritePlot1(src []int32, u, v, dstRow, w, h, stepU, stepV, spriteWidth, tint int,
	skew, skewStep, parity, width, boundsTopX, boundsBottomX int) {
	tr := (tint >> 16) & 0xff
	tg := (tint >> 8) & 0xff
	tb := tint & 0xff
	savedU := u
	pix := s.pixelData
	for row := -h; row < 0; row++ {
		rowBase := (v >> 16) * spriteWidth
		xs := skew >> 16
		span := w
		if xs < boundsTopX {
			cut := boundsTopX - xs
			span -= cut
			xs = boundsTopX
			u += stepU * cut
		}
		if xs+span >= boundsBottomX {
			span -= (xs + span) - boundsBottomX
		}
		parity = 1 - parity
		if parity != 0 {
			for col := xs; col < xs+span; col++ {
				si := (u >> 16) + rowBase
				if si >= 0 && si < len(src) {
					c := src[si]
					if c >= 0 { // -1 = transparent
						r := int(c>>16) & 0xff
						g := int(c>>8) & 0xff
						b := int(c) & 0xff
						di := col + dstRow
						if di >= 0 && di < len(pix) {
							if r == g && g == b {
								pix[di] = int32(((r*tr>>8)<<16) + ((g*tg>>8)<<8) + (b * tb >> 8))
							} else {
								pix[di] = c
							}
						}
					}
				}
				u += stepU
			}
		}
		v += stepV
		u = savedU
		dstRow += width
		skew += skewStep
	}
}

// transparentSpritePlot2 is the Go port of the two-tint direct-pixel
// transparentSpritePlot (Surface.java:1982): adds the second recolour case
// (r==255 && g==b -> tint2, the skin recolour) over the single-tint variant.
func (s *Scene) transparentSpritePlot2(src []int32, u, v, dstRow, w, h, stepU, stepV, spriteWidth, tint1, tint2 int,
	skew, skewStep, parity, width, boundsTopX, boundsBottomX int) {
	t1r := (tint1 >> 16) & 0xff
	t1g := (tint1 >> 8) & 0xff
	t1b := tint1 & 0xff
	t2r := (tint2 >> 16) & 0xff
	t2g := (tint2 >> 8) & 0xff
	t2b := tint2 & 0xff
	savedU := u
	pix := s.pixelData
	for row := -h; row < 0; row++ {
		rowBase := (v >> 16) * spriteWidth
		xs := skew >> 16
		span := w
		if xs < boundsTopX {
			cut := boundsTopX - xs
			span -= cut
			xs = boundsTopX
			u += stepU * cut
		}
		if xs+span >= boundsBottomX {
			span -= (xs + span) - boundsBottomX
		}
		parity = 1 - parity
		if parity != 0 {
			for col := xs; col < xs+span; col++ {
				si := (u >> 16) + rowBase
				if si >= 0 && si < len(src) {
					c := src[si]
					if c >= 0 { // -1 = transparent
						r := int(c>>16) & 0xff
						g := int(c>>8) & 0xff
						b := int(c) & 0xff
						di := col + dstRow
						if di >= 0 && di < len(pix) {
							if r == g && g == b {
								pix[di] = int32(((r*t1r>>8)<<16) + ((g*t1g>>8)<<8) + (b * t1b >> 8))
							} else if r == 255 && g == b {
								pix[di] = int32(((r*t2r>>8)<<16) + ((g*t2g>>8)<<8) + (b * t2b >> 8))
							} else {
								pix[di] = c
							}
						}
					}
				}
				u += stepU
			}
		}
		v += stepV
		u = savedU
		dstRow += width
		skew += skewStep
	}
}
