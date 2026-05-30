package render

// This file ports the real RSC software scanline rasterizer from
// Scene.java (deob-204): the per-scanline edge table (Class8), the
// method281 edge-table builder, and the method291 flat-shaded gouraud
// span fill. The MVP triangle-fan filler is gone.
//
// Coordinate convention matches the client exactly:
//   - The edge table is indexed by ABSOLUTE screen ROW y.
//   - method281 receives view-space X (anIntArray441) and view-space Y
//     (anIntArray442) plus per-vertex intensity (anIntArray443). It adds
//     anInt400 (vertical centre) to every Y to get the row index, and
//     leaves X in 8.8 fixed point (value << 8).
//   - method291 walks rows anInt439..anInt440 and writes spans, adding
//     anInt399 (horizontal centre) to the x offset.

// class8 is the per-scanline edge record (Class8.java):
//
//	anInt370 = left X edge   (8.8 fixed)
//	anInt371 = right X edge  (8.8 fixed)
//	anInt372 = left  intensity along that span
//	anInt373 = right intensity along that span
type class8 struct {
	anInt370 int32
	anInt371 int32
	anInt372 int32
	anInt373 int32
}

// raster holds the scanline scratch the client keeps on Scene: the edge
// table, the framebuffer, and the screen-centre / half-extent constants
// (anInt397..anInt400). One raster is built per Surface render pass.
type raster struct {
	pix    []int32
	stride int32 // anInt396 (== width)

	anInt397 int32 // half width  (horizontal clip extent)
	anInt398 int32 // half height (vertical clip extent)
	anInt399 int32 // horizontal centre
	anInt400 int32 // vertical centre

	edges []class8 // aClass8Array438, indexed by absolute row y

	anInt439 int32 // min row touched by current poly
	anInt440 int32 // max row touched by current poly
}

func newRaster(s *Surface, cx, cy int) *raster {
	r := &raster{
		pix:      s.Pix,
		stride:   int32(s.Width),
		anInt397: int32(cx),
		anInt398: int32(cy),
		anInt399: int32(cx),
		anInt400: int32(cy),
		edges:    make([]class8, s.Height),
	}
	// half extents: clip to the actual surface so edge rows never index OOB.
	if int32(s.Width)-int32(cx) < r.anInt397 {
		// keep symmetric clip the same as client (uses anInt397 both sides)
	}
	return r
}

// method281 builds the per-scanline left/right edge table for one polygon.
// i1 is the vertex count; ai/ai1/ai2 are view-X / view-Y / intensity.
// This is a faithful port of Scene.method281 (Scene.java:551). The triangle
// and quad fast paths and the general-polygon path are all reproduced; the
// int32 fixed-point overflow is preserved exactly.
func (r *raster) method281(i1 int32, ai, ai1, ai2 []int32) {
	if i1 == 3 {
		k1 := ai1[0] + r.anInt400
		k2 := ai1[1] + r.anInt400
		k3 := ai1[2] + r.anInt400
		k4 := ai[0]
		l5 := ai[1]
		j7 := ai[2]
		l8 := ai2[0]
		j10 := ai2[1]
		j11 := ai2[2]
		j12 := (r.anInt400 + r.anInt398) - 1
		var l12, j13, l13, j14 int32
		l14 := int32(0xbc614e)
		j15 := int32(-0xbc614e) // 0xff439eb2
		if k3 != k1 {
			j13 = ((j7 - k4) << 8) / (k3 - k1)
			j14 = ((j11 - l8) << 8) / (k3 - k1)
			if k1 < k3 {
				l12 = k4 << 8
				l13 = l8 << 8
				l14 = k1
				j15 = k3
			} else {
				l12 = j7 << 8
				l13 = j11 << 8
				l14 = k3
				j15 = k1
			}
			if l14 < 0 {
				l12 -= j13 * l14
				l13 -= j14 * l14
				l14 = 0
			}
			if j15 > j12 {
				j15 = j12
			}
		}
		var l15, j16, l16, j17 int32
		l17 := int32(0xbc614e)
		j18 := int32(-0xbc614e)
		if k2 != k1 {
			j16 = ((l5 - k4) << 8) / (k2 - k1)
			j17 = ((j10 - l8) << 8) / (k2 - k1)
			if k1 < k2 {
				l15 = k4 << 8
				l16 = l8 << 8
				l17 = k1
				j18 = k2
			} else {
				l15 = l5 << 8
				l16 = j10 << 8
				l17 = k2
				j18 = k1
			}
			if l17 < 0 {
				l15 -= j16 * l17
				l16 -= j17 * l17
				l17 = 0
			}
			if j18 > j12 {
				j18 = j12
			}
		}
		var l18, j19, l19, j20 int32
		l20 := int32(0xbc614e)
		j21 := int32(-0xbc614e)
		if k3 != k2 {
			j19 = ((j7 - l5) << 8) / (k3 - k2)
			j20 = ((j11 - j10) << 8) / (k3 - k2)
			if k2 < k3 {
				l18 = l5 << 8
				l19 = j10 << 8
				l20 = k2
				j21 = k3
			} else {
				l18 = j7 << 8
				l19 = j11 << 8
				l20 = k3
				j21 = k2
			}
			if l20 < 0 {
				l18 -= j19 * l20
				l19 -= j20 * l20
				l20 = 0
			}
			if j21 > j12 {
				j21 = j12
			}
		}
		r.anInt439 = l14
		if l17 < r.anInt439 {
			r.anInt439 = l17
		}
		if l20 < r.anInt439 {
			r.anInt439 = l20
		}
		r.anInt440 = j15
		if j18 > r.anInt440 {
			r.anInt440 = j18
		}
		if j21 > r.anInt440 {
			r.anInt440 = j21
		}
		var l21 int32
		var i, j, l int32
		for k := r.anInt439; k < r.anInt440; k++ {
			if k >= l14 && k < j15 {
				i = l12
				j = l12
				l = l13
				l21 = l13
				l12 += j13
				l13 += j14
			} else {
				i = 0xa0000
				j = -0xa0000 // 0xfff60000
			}
			if k >= l17 && k < j18 {
				if l15 < i {
					i = l15
					l = l16
				}
				if l15 > j {
					j = l15
					l21 = l16
				}
				l15 += j16
				l16 += j17
			}
			if k >= l20 && k < j21 {
				if l18 < i {
					i = l18
					l = l19
				}
				if l18 > j {
					j = l18
					l21 = l19
				}
				l18 += j19
				l19 += j20
			}
			if k >= 0 && k < int32(len(r.edges)) {
				e := &r.edges[k]
				e.anInt370 = i
				e.anInt371 = j
				e.anInt372 = l
				e.anInt373 = l21
			}
		}
		if r.anInt439 < r.anInt400-r.anInt398 {
			r.anInt439 = r.anInt400 - r.anInt398
		}
		return
	}

	// general polygon path (also handles i1 == 4); Scene.java:898-1024
	r.anInt439 = ai1[0] + r.anInt400
	ai1[0] = r.anInt439
	r.anInt440 = r.anInt439
	for k := int32(1); k < i1; k++ {
		i2 := ai1[k] + r.anInt400
		ai1[k] = i2
		if i2 < r.anInt439 {
			r.anInt439 = i2
		} else if i2 > r.anInt440 {
			r.anInt440 = i2
		}
	}
	if r.anInt439 < r.anInt400-r.anInt398 {
		r.anInt439 = r.anInt400 - r.anInt398
	}
	if r.anInt440 >= r.anInt400+r.anInt398 {
		r.anInt440 = (r.anInt400 + r.anInt398) - 1
	}
	if r.anInt439 >= r.anInt440 {
		return
	}
	for k := r.anInt439; k < r.anInt440; k++ {
		if k >= 0 && k < int32(len(r.edges)) {
			r.edges[k].anInt370 = 0xa0000
			r.edges[k].anInt371 = -0xa0000
		}
	}
	j2 := i1 - 1
	i3 := ai1[0]
	i4 := ai1[j2]
	if i3 < i4 {
		i5 := ai[0] << 8
		j6 := ((ai[j2] - ai[0]) << 8) / (i4 - i3)
		l7 := ai2[0] << 8
		j9 := ((ai2[j2] - ai2[0]) << 8) / (i4 - i3)
		if i3 < 0 {
			i5 -= j6 * i3
			l7 -= j9 * i3
			i3 = 0
		}
		if i4 > r.anInt440 {
			i4 = r.anInt440
		}
		for k := i3; k <= i4; k++ {
			if k >= 0 && k < int32(len(r.edges)) {
				e := &r.edges[k]
				e.anInt370 = i5
				e.anInt371 = i5
				e.anInt372 = l7
				e.anInt373 = l7
			}
			i5 += j6
			l7 += j9
		}
	} else if i3 > i4 {
		j5 := ai[j2] << 8
		k6 := ((ai[0] - ai[j2]) << 8) / (i3 - i4)
		i8 := ai2[j2] << 8
		k9 := ((ai2[0] - ai2[j2]) << 8) / (i3 - i4)
		if i4 < 0 {
			j5 -= k6 * i4
			i8 -= k9 * i4
			i4 = 0
		}
		if i3 > r.anInt440 {
			i3 = r.anInt440
		}
		for k := i4; k <= i3; k++ {
			if k >= 0 && k < int32(len(r.edges)) {
				e := &r.edges[k]
				e.anInt370 = j5
				e.anInt371 = j5
				e.anInt372 = i8
				e.anInt373 = i8
			}
			j5 += k6
			i8 += k9
		}
	}
	for k := int32(0); k < j2; k++ {
		k5 := k + 1
		j3 := ai1[k]
		j4 := ai1[k5]
		if j3 < j4 {
			l6 := ai[k] << 8
			j8 := ((ai[k5] - ai[k]) << 8) / (j4 - j3)
			l9 := ai2[k] << 8
			l10 := ((ai2[k5] - ai2[k]) << 8) / (j4 - j3)
			if j3 < 0 {
				l6 -= j8 * j3
				l9 -= l10 * j3
				j3 = 0
			}
			if j4 > r.anInt440 {
				j4 = r.anInt440
			}
			for l11 := j3; l11 <= j4; l11++ {
				if l11 < 0 || l11 >= int32(len(r.edges)) {
					l6 += j8
					l9 += l10
					continue
				}
				e := &r.edges[l11]
				if l6 < e.anInt370 {
					e.anInt370 = l6
					e.anInt372 = l9
				}
				if l6 > e.anInt371 {
					e.anInt371 = l6
					e.anInt373 = l9
				}
				l6 += j8
				l9 += l10
			}
		} else if j3 > j4 {
			i7 := ai[k5] << 8
			k8 := ((ai[k] - ai[k5]) << 8) / (j3 - j4)
			i10 := ai2[k5] << 8
			i11 := ((ai2[k] - ai2[k5]) << 8) / (j3 - j4)
			if j4 < 0 {
				i7 -= k8 * j4
				i10 -= i11 * j4
				j4 = 0
			}
			if j3 > r.anInt440 {
				j3 = r.anInt440
			}
			for i12 := j4; i12 <= j3; i12++ {
				if i12 < 0 || i12 >= int32(len(r.edges)) {
					i7 += k8
					i10 += i11
					continue
				}
				e := &r.edges[i12]
				if i7 < e.anInt370 {
					e.anInt370 = i7
					e.anInt372 = i10
				}
				if i7 > e.anInt371 {
					e.anInt371 = i7
					e.anInt373 = i10
				}
				i7 += k8
				i10 += i11
			}
		}
	}
	if r.anInt439 < r.anInt400-r.anInt398 {
		r.anInt439 = r.anInt400 - r.anInt398
	}
}

// method291 is the default (opaque, non-textured) flat-shaded gouraud span
// fill (Scene.java:1406-1428 driver + method291 :2245). It walks the rows
// the edge table just filled and emits horizontal spans whose colour is
// looked up in the shade ramp by the interpolated intensity. ramp is the
// 256-entry shade table for the face's flat colour.
func (r *raster) method291(ramp *[256]int32) {
	stride := r.stride
	base := r.anInt399 + r.anInt439*stride
	for i := r.anInt439; i < r.anInt440; i++ {
		if i < 0 || i >= int32(len(r.edges)) {
			base += stride
			continue
		}
		e := r.edges[i]
		j := e.anInt370 >> 8
		k5 := e.anInt371 >> 8
		i7 := k5 - j
		if i7 <= 0 {
			base += stride
			continue
		}
		j8 := e.anInt372             // left intensity
		k9 := (e.anInt373 - j8) / i7 // intensity step per pixel
		if j < -r.anInt397 {
			j8 += (-r.anInt397 - j) * k9
			j = -r.anInt397
			i7 = k5 - j
		}
		if k5 > r.anInt397 {
			i7 = r.anInt397 - j
		}
		r.spanFill(-i7, base+j, ramp, j8, k9)
		base += stride
	}
}

// texClass holds the per-size-class constants method282/283/285 hard-code for a
// 64px (class0) or 128px (class1) texture. They are derived once from the
// resolved viewDist=9 shift expressions (NOT transcribed from the Java
// `x << 5-viewDist+...` source, which Go would mis-parenthesise).
type texClass struct {
	// per-row projection setup shifts (method282)
	shA  uint // l9/k11 base   (class1 <<12, class0 <<11)
	shM  uint // k10/i12 (U/V numerator per-row delta)  (class1 <<7, class0 <<6)
	shN  uint // i11/k12 (denominator per-row delta)     (class1 <<3, class0 <<2)
	shD  uint // i13     base   (always <<5)
	shDM uint // k13     (intensity-mip per-row delta)    (class1 <<0, class0 <<0)
	shDN uint // i14     (>> viewDist-5 == >>4)
	// inner-loop texel constants (method283/286)
	proj      uint  // perspective fixed-point shift (class1 7, class0 6)
	vmask     int32 // row select mask     (class1 0x3f80, class0 0x0fc0)
	umask     int32 // U wrap mask          (class1 0x3fff, class0 0x0fff)
	mipmask   int32 // shade-mip block bits (class1 0x600000, class0 0x0c0000)
	shadeSh   uint  // intensity -> >>> shift selector (class1 23, class0 20)
	uClampMax int32 // U clamp ceiling      (class1 16256, class0 4032)
}

var (
	texClass1 = texClass{
		shA: 12, shM: 7, shN: 3, shD: 5, shDM: 0, shDN: 4,
		proj: 7, vmask: 0x3f80, umask: 0x3fff, mipmask: 0x600000,
		shadeSh: 23, uClampMax: 16256,
	}
	texClass0 = texClass{
		shA: 11, shM: 6, shN: 2, shD: 5, shDM: 0, shDN: 4,
		proj: 6, vmask: 0x0fc0, umask: 0x0fff, mipmask: 0x0c0000,
		shadeSh: 20, uClampMax: 4032,
	}
)

// method282 is the perspective-correct textured-span driver (Scene.java:1036).
// buf is the shaded texel buffer for the face's texture; camX/camY/camZ are the
// CAMERA-space coords of the face's first/second/last ORIGINAL vertices (indexed
// in clip-emit order; only [0], [1] and [k-1] are read), and k is the original
// face vertex count (l10). It sets up the planar P/M/N projection from those 3
// vertices, then walks the edge-table rows method281 just filled, emitting one
// textured span per row via texSpan. The interlace branch is dropped (false).
func (r *raster) method282(buf *textureBuf, camX, camY, camZ []int32, k int) {
	if buf == nil || k < 3 {
		return
	}
	var tc texClass
	if buf.size == 128 {
		tc = texClass1
	} else {
		tc = texClass0
	}

	i1 := camX[0]
	k1 := camY[0]
	j2 := camZ[0]
	i3 := i1 - camX[1]
	k3 := k1 - camY[1]
	i4 := j2 - camZ[1]
	last := k - 1
	i6 := camX[last] - i1
	j7 := camY[last] - k1
	k8 := camZ[last] - j2

	// planar texture-projection constants (resolved viewDist=9 shifts).
	l9 := (i6*k1 - j7*i1) << tc.shA
	k10 := (j7*j2 - k8*k1) << tc.shM
	i11 := (k8*i1 - i6*j2) << tc.shN
	k11 := (i3*k1 - k3*i1) << tc.shA
	i12 := (k3*j2 - i4*k1) << tc.shM
	k12 := (i4*i1 - i3*j2) << tc.shN
	i13 := (k3*i6 - i3*j7) << tc.shD
	k13 := (i4*j7 - k3*k8) << tc.shDM
	i14 := (i3*k8 - i4*i6) >> tc.shDN

	// per-pixel deltas used to slide the row start to the clipped left x.
	k14 := k10 >> 4
	i15 := i12 >> 4
	k15 := k13 >> 4

	i16 := r.anInt439 - r.anInt400
	k16 := r.stride
	i17 := r.anInt399 + r.anInt439*k16

	// advance the row accumulators to the first scanned row.
	l9 += i11 * i16
	k11 += k12 * i16
	i13 += i14 * i16

	for i := r.anInt439; i < r.anInt440; i++ {
		if i < 0 || i >= int32(len(r.edges)) {
			l9 += i11
			k11 += k12
			i13 += i14
			i17 += k16
			continue
		}
		e := r.edges[i]
		j := e.anInt370 >> 8
		rightX := e.anInt371 >> 8
		span := rightX - j
		if span <= 0 {
			l9 += i11
			k11 += k12
			i13 += i14
			i17 += k16
			continue
		}
		i22 := e.anInt372
		k23 := (e.anInt373 - i22) / span
		if j < -r.anInt397 {
			i22 += (-r.anInt397 - j) * k23
			j = -r.anInt397
			span = rightX - j
		}
		if rightX > r.anInt397 {
			span = r.anInt397 - j
		}
		if span > 0 {
			r.texSpan(buf, &tc,
				l9+k14*j, k11+i15*j, i13+k15*j,
				k10, i12, k13,
				span, i17+j, i22, k23<<2)
		}
		l9 += i11
		k11 += k12
		i13 += i14
		i17 += k16
	}
}

// texSpan emits one perspective-correct textured scanline (the readable
// per-pixel port of method283 opaque / method285 transparent). Args mirror
// method283: uNum/vNum/denom are the planar U-numerator, V-numerator and
// denominator at the span's left pixel; dU/dV/dDen are their per-16px deltas;
// n is the span length; off the framebuffer offset; shade the intensity
// accumulator and dShade its (already <<2) per-step delta. The perspective
// divide is recomputed every 16 pixels, affine-interpolated between (the client
// scheme). When buf.hasAlpha, texel value 0 (the transparency key) is skipped
// (method285); otherwise every pixel is written (method283).
func (r *raster) texSpan(buf *textureBuf, tc *texClass,
	uNum, vNum, denom, dU, dV, dDen int32,
	n, off, shade, dShade int32) {
	if n <= 0 {
		return
	}
	pix := r.pix
	plen := int32(len(pix))
	tex := buf.texels
	tlen := int32(len(tex))
	skipAlpha := buf.hasAlpha

	// perspective divide at the span start.
	u, v := projDiv(uNum, vNum, denom, tc)
	pos := int32(0)
	for pos < n {
		// next 16-pixel group boundary
		grp := n - pos
		if grp > 16 {
			grp = 16
		}
		// advance the planar numerators/denominator one group and divide at the
		// group end, then affine-interpolate the 16 pixels between.
		uNum += dU
		vNum += dV
		denom += dDen
		uEnd, vEnd := projDiv(uNum, vNum, denom, tc)
		stepU := (uEnd - u) >> 4
		stepV := (vEnd - v) >> 4

		for g := int32(0); g < grp; g++ {
			if (g & 3) == 0 {
				// refresh the shade right-shift + re-add the mip-block bits every
				// 4 pixels (method283: i4 = k2>>shadeSh; i = (i&umask)+(k2&mipmask)).
				u = (u & tc.umask) + (shade & tc.mipmask)
				shade += dShade
			}
			i4 := uint(shade>>tc.shadeSh) & 31
			ti := (v & tc.vmask) + (u >> tc.proj)
			if ti >= 0 && ti < tlen {
				c := int32(uint32(tex[ti]) >> i4)
				if !(skipAlpha && c == 0) {
					if off >= 0 && off < plen {
						pix[off] = c
					}
				}
			}
			off++
			u += stepU
			v += stepV
		}
		pos += grp
		// continue affine accumulators from the divided group-end values.
		u = uEnd
		v = vEnd
	}
}

// projDiv performs the per-group perspective divide (method283:1439-1446):
// u = clamp(uNum/denom) << proj, v = (vNum/denom) << proj. denom==0 keeps the
// previous (here: zero) value, matching the client's guarded divide.
func projDiv(uNum, vNum, denom int32, tc *texClass) (u, v int32) {
	if denom != 0 {
		u = (uNum / denom) << tc.proj
		v = (vNum / denom) << tc.proj
	}
	if u < 0 {
		u = 0
	} else if u > tc.uClampMax {
		u = tc.uClampMax
	}
	return u, v
}

// spanFill is method291's inner loop (Scene.java:2245). off is the pixel
// position into pix; n is NEGATIVE span length (count = -n). l is the
// 8.x-fixed intensity accumulator, i1 the step. The ramp index is
// (l>>8)&0xff. The <<2 on the step and the 4-pixel-per-sample unrolling are
// preserved (the client samples the ramp once per 4 pixels).
func (r *raster) spanFill(n, off int32, ramp *[256]int32, l, i1 int32) {
	if n >= 0 {
		return
	}
	pix := r.pix
	plen := int32(len(pix))
	i1 <<= 2
	k := ramp[(l>>8)&0xff]
	l += i1
	put := func(v int32) {
		if off >= 0 && off < plen {
			pix[off] = v
		}
		off++
	}
	j1 := n / 16
	for k1 := j1; k1 < 0; k1++ {
		put(k)
		put(k)
		put(k)
		put(k)
		k = ramp[(l>>8)&0xff]
		l += i1
		put(k)
		put(k)
		put(k)
		put(k)
		k = ramp[(l>>8)&0xff]
		l += i1
		put(k)
		put(k)
		put(k)
		put(k)
		k = ramp[(l>>8)&0xff]
		l += i1
		put(k)
		put(k)
		put(k)
		put(k)
		k = ramp[(l>>8)&0xff]
		l += i1
	}
	j1 = -(n % 16)
	for l1 := int32(0); l1 < j1; l1++ {
		put(k)
		if (l1 & 3) == 3 {
			k = ramp[(l>>8)&0xff]
			l += i1
		}
	}
}
