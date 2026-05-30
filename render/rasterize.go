package render

// gouraudRamp builds the 256-entry shade ramp for a flat 5:5:5 colour, exactly
// as Scene.method281 does on a colour miss (Scene.java:1326). fill is the
// encoded flat colour from a face (fill = -1 - colour). ramp[255-j] scales each
// component by j^2/65536, so index 255 is brightest, 0 is black.
func gouraudRamp(fill int32) [256]int32 {
	c := -1 - fill
	r := (c >> 10 & 0x1f) * 8
	gg := (c >> 5 & 0x1f) * 8
	b := (c & 0x1f) * 8
	var ramp [256]int32
	for j := int32(0); j < 256; j++ {
		j6 := j * j
		k7 := (r * j6) / 0x10000
		l8 := (gg * j6) / 0x10000
		j10 := (b * j6) / 0x10000
		ramp[255-j] = (k7 << 16) + (l8 << 8) + j10
	}
	return ramp
}

// fillTriangleGouraud rasterizes a single triangle into the surface using a
// per-vertex intensity ramp index (0..255). This is a faithful-enough port of
// the method281 edge table for the MVP: scanline fill with linear interpolation
// of X and intensity along edges, clamped to the surface. Screen coordinates
// are pixel-space (already centred). ramp maps intensity index -> RGB.
func fillTriangleGouraud(s *Surface, x0, y0, i0, x1, y1, i1, x2, y2, i2 int, ramp *[256]int32) {
	// order by y ascending
	if y1 < y0 {
		x0, y0, i0, x1, y1, i1 = x1, y1, i1, x0, y0, i0
	}
	if y2 < y0 {
		x0, y0, i0, x2, y2, i2 = x2, y2, i2, x0, y0, i0
	}
	if y2 < y1 {
		x1, y1, i1, x2, y2, i2 = x2, y2, i2, x1, y1, i1
	}
	if y2 == y0 {
		return
	}
	total := y2 - y0
	for y := y0; y <= y2; y++ {
		if y < 0 || y >= s.Height {
			continue
		}
		// long edge 0->2
		var ax, ai int
		{
			t := y - y0
			ax = x0 + (x2-x0)*t/total
			ai = i0 + (i2-i0)*t/total
		}
		// short edge: 0->1 (upper) or 1->2 (lower)
		var bx, bi int
		if y < y1 {
			seg := y1 - y0
			if seg == 0 {
				bx, bi = x0, i0
			} else {
				t := y - y0
				bx = x0 + (x1-x0)*t/seg
				bi = i0 + (i1-i0)*t/seg
			}
		} else {
			seg := y2 - y1
			if seg == 0 {
				bx, bi = x1, i1
			} else {
				t := y - y1
				bx = x1 + (x2-x1)*t/seg
				bi = i1 + (i2-i1)*t/seg
			}
		}
		lx, li, rx, ri := ax, ai, bx, bi
		if rx < lx {
			lx, li, rx, ri = rx, ri, lx, li
		}
		span := rx - lx
		row := y * s.Width
		for x := lx; x <= rx; x++ {
			if x < 0 || x >= s.Width {
				continue
			}
			ii := li
			if span > 0 {
				ii = li + (ri-li)*(x-lx)/span
			}
			if ii < 0 {
				ii = 0
			}
			if ii > 255 {
				ii = 255
			}
			s.Pix[row+x] = ramp[ii]
		}
	}
}
