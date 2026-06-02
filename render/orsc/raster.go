package orsc

import "math/rand"

// raster.go — faithful port of the rasterizer half of OpenRSC's three/ renderer:
// Scanline.java (the per-row span record, already declared as the Scanline type
// in types.go), the Scene edge-table builder (Scene.setFrustum(int,int,int[],
// int,int,RSModel,int[],int[],int,int,int), Scene.java:717), the textured/
// flat-colour face filler (Scene.setFrustum(int[],RSModel,int,int,int,int[],
// int[],int,int), Scene.java:1843), and the whole Shader.java family of
// perspective-correct shaded span writers (Shader.java:8,203,346,511,649,775).
//
// =============================================================================
// Java -> Go fixed-point rules (apply throughout this file)
// =============================================================================
//   - Java `int` is 32-bit. All the projection/texture-coordinate math below
//     deliberately relies on 32-bit wraparound, so EVERY accumulator and every
//     intermediate is int32. (Widening to int would change overflow behaviour
//     and desync from the Java output.)
//   - `>>`  in Java on an int is an ARITHMETIC shift  -> Go `>>` on int32.
//   - `>>>` in Java is a LOGICAL shift                -> Go: convert to uint32,
//     shift, convert back: int32(uint32(x) >> n). Used for the texel-value
//     brightness select (`var14[..] >>> var21`) and the FastMath.bitwiseAnd
//     blend reads.
//   - FastMath.bitwiseAnd(a,b) is just `a & b`; reproduced inline as `a & b`.
//   - `<<` precedence: Java `a << b - c + d` parses as `a << (b - c + d)` because
//     additive binds tighter than shift in Java. The shift-count expressions in
//     the 1843 filler (e.g. `<< 4 - this.rot1024_vp_src + 5 + 7`) are reproduced
//     with the SAME grouping — see the inline notes at each one.
//   - Texel value 0 == transparent/skip in the m_S (has-alpha) Shader variants;
//     they guard each write with `if texel != 0`.
//
// The Scene methods are reproduced as *Scene methods with readable names; their
// Java overload signatures are noted so scene.go (the sibling port driving
// endScene) can call them at the exact dispatch points.

// =============================================================================
// Shade-mip brightness mask. The Shader blends already-shaded texels with the
// destination using a /2 (one bit) darken on the dest: `(dest>>1) & 0x7f7f7f`
// or `(dest & 0xfefeff) >> 1`. These are the two ways Java spells "halve every
// channel"; the 0xfefeff form clears the carry between channels first, the
// 0x7f7f7f form clears it after — both yield the same packed result.
// =============================================================================
const (
	half7f = 0x7f7f7f // (x >> 1) & 0x7f7f7f
	halfFe = 0xfefeff // (x & 0xfefeff) >> 1  (alt spelling; 16711423)
	halfFE = 0xfefefe // (x & 0xfefefe) >> 1  (16711422)
)

// lshr is the Java >>> (logical right shift) on a 32-bit value.
func lshr(x, n int32) int32 { return int32(uint32(x) >> uint(n)) }

// subtractiveBlend mirrors Scene.m_Ub (Scene.java:91 `this.m_Ub = false;`): the
// alternate copyBlock4 flat-fill path. It is hard-false in this client, so it is
// a package constant rather than a Scene field (types.go has no m_Ub). The
// copyBlock4 branch is kept for fidelity but is dead with this value.
const subtractiveBlend = false

// rngFloat returns a uniform [0,1) double, the Java Math.random() used to pick a
// gradient-cache eviction slot (Scene.java:2184). The slot choice is a
// non-deterministic eviction detail (any free-or-LRU slot yields identical
// pixels once the ramp is built), so the source of randomness is not
// load-bearing for output fidelity.
func rngFloat() float64 { return rngState.Float64() }

// rngState backs rngFloat; seeded deterministically so test runs are repeatable.
var rngState = rand.New(rand.NewSource(1))

// =============================================================================
// scanEdges — Scene.setFrustum(var1, var2, var3, var4, var5, model, var7, var8,
// var9, var10, var11) (Scene.java:717-1342): build the per-row left/right span
// edge table (s.scanlines, Scene.m_x) for one face.
//
//	rowsY (Java var3 == m_B): per-vertex projected screen Y (vertexParam2). m_Nb
//	                  is added to each to form the ABSOLUTE scanline row index.
//	colsX (Java var7 == m_yb): per-vertex projected screen X (vertexParam6); the
//	                  span endpoint stored in each Scanline's m_d/m_k.
//	shade (Java var8 == m_r): per-vertex shade value; stored in m_e/m_l and
//	                  interpolated along the span by the filler.
//	n     (Java var11): vertex count (3, 4, or general).
//	faceID (var2), model (var6), pickFlag (var10==5960) feed the pick probe at the
//	  tail; there is no pick buffer in this port so that block is a no-op but
//	  preserved structurally and cited.
//
// CALL-SITE BINDING (Scene.java:2844): endScene calls
//
//	setFrustum(0, faceID, m_B, 0, 0, model, m_yb, m_r, 0, 5960, vertCount)
//
// with m_B[i]=vertexParam2 (screenY), m_yb[i]=vertexParam6 (screenX),
// m_r[i]=shade (Scene.java:2779-2781). So scene.go must call
//
//	scanEdges(faceID, m_B, model, m_yb, m_r, 5960, vertCount).
//
// Sets s.rasterMinY (m_Xb) / s.rasterMaxY (m_Cb) to the touched row range.
// =============================================================================
func (s *Scene) scanEdges(faceID int, rowsY []int32, model *Model, colsX, shade []int32, pickFlag, n int32) {
	// Internal aliases keep the body 1:1 with Scene.setFrustum's var3/var7/var8.
	xs, us, vs := rowsY, colsX, shade

	mNb := s.mNb
	mwb := s.mWb

	switch n {
	case 3:
		// ---- triangle fast path (Scene.java:751-930) ----
		y0 := mNb + xs[0]
		y1 := xs[1] + mNb
		y2 := mNb + xs[2]
		u0, u1, u2 := us[0], us[1], us[2]
		v0, v1, v2 := vs[0], vs[1], vs[2]
		lastRow := mwb + (mNb - 1)

		// edge 0-2
		var eA_u, eA_uStep, eA_v, eA_vStep int32
		eA_min := int32(TRANSPARENT)
		eA_max := int32(-TRANSPARENT)
		if y0 != y2 {
			if y0 >= y2 {
				eA_v = v2 << 8
				eA_min = y2
				eA_max = y0
				eA_u = u2 << 8
			} else {
				eA_min = y0
				eA_max = y2
				eA_u = u0 << 8
				eA_v = v0 << 8
			}
			eA_vStep = ((v2 - v0) << 8) / (y2 - y0)
			eA_uStep = ((u2 - u0) << 8) / (y2 - y0)
			if eA_min < 0 {
				eA_u -= eA_min * eA_uStep
				eA_v -= eA_vStep * eA_min
				eA_min = 0
			}
			if eA_max > lastRow {
				eA_max = lastRow
			}
		}

		// edge 0-1
		var eB_u, eB_uStep, eB_v, eB_vStep int32
		eB_min := int32(TRANSPARENT)
		eB_max := int32(-TRANSPARENT)
		if y0 != y1 {
			eB_uStep = ((u1 - u0) << 8) / (y1 - y0)
			eB_vStep = ((v1 - v0) << 8) / (y1 - y0)
			if y1 > y0 {
				eB_v = v0 << 8
				eB_max = y1
				eB_u = u0 << 8
				eB_min = y0
			} else {
				eB_v = v1 << 8
				eB_min = y1
				eB_u = u1 << 8
				eB_max = y0
			}
			if eB_max > lastRow {
				eB_max = lastRow
			}
			if eB_min < 0 {
				eB_v -= eB_vStep * eB_min
				eB_u -= eB_min * eB_uStep
				eB_min = 0
			}
		}

		// edge 1-2
		var eC_u, eC_uStep, eC_v, eC_vStep int32
		eC_min := int32(TRANSPARENT)
		eC_max := int32(-TRANSPARENT)
		if y2 != y1 {
			if y2 > y1 {
				eC_u = u1 << 8
				eC_min = y1
				eC_v = v1 << 8
				eC_max = y2
			} else {
				eC_min = y2
				eC_v = v2 << 8
				eC_max = y1
				eC_u = u2 << 8
			}
			eC_vStep = ((v2 - v1) << 8) / (y2 - y1)
			eC_uStep = ((u2 - u1) << 8) / (y2 - y1)
			if eC_min < 0 {
				eC_v -= eC_min * eC_vStep
				eC_u -= eC_uStep * eC_min
				eC_min = 0
			}
			if lastRow < eC_max {
				eC_max = lastRow
			}
		}

		minY := eA_min
		if minY > eB_min {
			minY = eB_min
		}
		if minY > eC_min {
			minY = eC_min
		}
		maxY := eA_max
		if eB_max > maxY {
			maxY = eB_max
		}
		if maxY < eC_max {
			maxY = eC_max
		}
		s.rasterMinY = minY
		s.rasterMaxY = maxY

		var spanL, spanR, intL, intR int32
		for row := minY; maxY > row; row++ {
			if row >= eA_min && row < eA_max {
				spanR = eA_u
				spanL = eA_u
				intR = eA_v
				intL = eA_v
				eA_u += eA_uStep
				eA_v += eA_vStep
			} else {
				spanL = 655360
				spanR = -655360
			}
			if eB_min <= row && row < eB_max {
				if spanR < eB_u {
					spanR = eB_u
					intR = eB_v
				}
				if eB_u < spanL {
					spanL = eB_u
					intL = eB_v
				}
				eB_v += eB_vStep
				eB_u += eB_uStep
			}
			if row >= eC_min && eC_max > row {
				if eC_u > spanR {
					intR = eC_v
					spanR = eC_u
				}
				if eC_u < spanL {
					spanL = eC_u
					intL = eC_v
				}
				eC_v += eC_vStep
				eC_u += eC_uStep
			}
			sl := s.scanlines[row]
			sl.mE = intL
			sl.mL = intR
			sl.mD = spanL
			sl.mK = spanR
		}
		if s.rasterMinY < mNb-mwb {
			s.rasterMinY = mNb - mwb
		}

	case 4:
		// ---- quad fast path (Scene.java:1087-1324) ----
		s.scanEdgesQuad(xs, us, vs)

	default:
		// ---- general n-gon path (Scene.java:931-1086) ----
		s.scanEdgesPoly(xs, us, vs, n)
	}

	// ---- pick probe (Scene.java:1327-1338): pickFlag==5960. No pick buffer in
	// this port, so this is intentionally a no-op; preserved for fidelity.
	_ = pickFlag
	_ = faceID
	_ = model
}

// scanEdgesQuad is the var11==4 branch of Scene.setFrustum (Scene.java:1087-1324).
func (s *Scene) scanEdgesQuad(xs, us, vs []int32) {
	mNb := s.mNb
	mwb := s.mWb

	y0 := xs[0] + mNb
	y1 := mNb + xs[1]
	y2 := mNb + xs[2]
	y3 := mNb + xs[3]
	u0, u1, u2, u3 := us[0], us[1], us[2], us[3]
	v0, v1, v2, v3 := vs[0], vs[1], vs[2], vs[3]
	lastRow := mwb + mNb - 1

	// edge 3-0 (Scene.java:1107-1131)
	var e0_u, e0_uStep, e0_v, e0_vStep int32
	e0_min := int32(TRANSPARENT)
	e0_max := int32(-TRANSPARENT)
	if y3 != y0 {
		e0_uStep = ((u3 - u0) << 8) / (y3 - y0)
		e0_vStep = ((v3 - v0) << 8) / (y3 - y0)
		if y3 <= y0 {
			e0_min = y3
			e0_u = u3 << 8
			e0_v = v3 << 8
			e0_max = y0
		} else {
			e0_max = y3
			e0_u = u0 << 8
			e0_min = y0
			e0_v = v0 << 8
		}
		if e0_min < 0 {
			e0_v -= e0_vStep * e0_min
			e0_u -= e0_min * e0_uStep
			e0_min = 0
		}
		if lastRow < e0_max {
			e0_max = lastRow
		}
	}

	// edge 0-1 (Scene.java:1133-1163)
	var e1_u, e1_uStep, e1_v, e1_vStep int32
	e1_min := int32(TRANSPARENT)
	e1_max := int32(-TRANSPARENT)
	if y0 != y1 {
		e1_vStep = ((v1 - v0) << 8) / (y1 - y0)
		if y1 <= y0 {
			e1_min = y1
			e1_v = v1 << 8
			e1_max = y0
			e1_u = u1 << 8
		} else {
			e1_min = y0
			e1_max = y1
			e1_u = u0 << 8
			e1_v = v0 << 8
		}
		e1_uStep = ((u1 - u0) << 8) / (y1 - y0)
		if lastRow < e1_max {
			e1_max = lastRow
		}
		if e1_min < 0 {
			e1_u -= e1_min * e1_uStep
			e1_v -= e1_vStep * e1_min
			e1_min = 0
		}
	}

	// edge 1-2 (Scene.java:1165-1195)
	var e2_u, e2_uStep, e2_v, e2_vStep int32
	e2_min := int32(TRANSPARENT)
	e2_max := int32(-TRANSPARENT)
	if y2 != y1 {
		e2_vStep = ((v2 - v1) << 8) / (y2 - y1)
		if y2 <= y1 {
			e2_min = y2
			e2_v = v2 << 8
			e2_u = u2 << 8
			e2_max = y1
		} else {
			e2_min = y1
			e2_v = v1 << 8
			e2_max = y2
			e2_u = u1 << 8
		}
		e2_uStep = ((u2 - u1) << 8) / (y2 - y1)
		if e2_min < 0 {
			e2_v -= e2_min * e2_vStep
			e2_u -= e2_uStep * e2_min
			e2_min = 0
		}
		if lastRow < e2_max {
			e2_max = lastRow
		}
	}

	// edge 2-3 (Scene.java:1197-1227)
	var e3_u, e3_uStep, e3_v, e3_vStep int32
	e3_min := int32(TRANSPARENT)
	e3_max := int32(-TRANSPARENT)
	if y3 != y2 {
		e3_vStep = ((v3 - v2) << 8) / (y3 - y2)
		if y2 >= y3 {
			e3_max = y2
			e3_v = v3 << 8
			e3_u = u3 << 8
			e3_min = y3
		} else {
			e3_v = v2 << 8
			e3_max = y3
			e3_min = y2
			e3_u = u2 << 8
		}
		e3_uStep = ((u3 - u2) << 8) / (y3 - y2)
		if e3_min < 0 {
			e3_u -= e3_min * e3_uStep
			e3_v -= e3_min * e3_vStep
			e3_min = 0
		}
		if lastRow < e3_max {
			e3_max = lastRow
		}
	}

	minY := e0_min
	if minY > e1_min {
		minY = e1_min
	}
	if e2_min < minY {
		minY = e2_min
	}
	maxY := e0_max
	if minY > e3_min {
		minY = e3_min
	}
	if e1_max > maxY {
		maxY = e1_max
	}
	if e2_max > maxY {
		maxY = e2_max
	}
	if maxY < e3_max {
		maxY = e3_max
	}
	s.rasterMinY = minY
	s.rasterMaxY = maxY

	var spanL, spanR, intL, intR int32
	for row := minY; maxY > row; row++ {
		if row >= e0_min && e0_max > row {
			intR = e0_v
			intL = e0_v
			spanR = e0_u
			spanL = e0_u
			e0_v += e0_vStep
			e0_u += e0_uStep
		} else {
			spanR = -655360
			spanL = 655360
		}
		if e1_min <= row && e1_max > row {
			if e1_u < spanL {
				intL = e1_v
				spanL = e1_u
			}
			if spanR < e1_u {
				intR = e1_v
				spanR = e1_u
			}
			e1_u += e1_uStep
			e1_v += e1_vStep
		}
		if row >= e2_min && row < e2_max {
			if e2_u > spanR {
				spanR = e2_u
				intR = e2_v
			}
			if e2_u < spanL {
				spanL = e2_u
				intL = e2_v
			}
			e2_u += e2_uStep
			e2_v += e2_vStep
		}
		if e3_min <= row && e3_max > row {
			if e3_u > spanR {
				intR = e3_v
				spanR = e3_u
			}
			if e3_u < spanL {
				spanL = e3_u
				intL = e3_v
			}
			e3_v += e3_vStep
			e3_u += e3_uStep
		}
		sl := s.scanlines[row]
		sl.mE = intL
		sl.mD = spanL
		sl.mK = spanR
		sl.mL = intR
	}
	if mNb-mwb > s.rasterMinY {
		s.rasterMinY = mNb - mwb
	}
}

// scanEdgesPoly is the general var11>4 (and !=4) branch (Scene.java:931-1086):
// initialise every touched row to an empty span, lay the first->last closing
// edge, then expand each consecutive edge by min/max into the row spans.
func (s *Scene) scanEdgesPoly(xs, us, vs []int32, n int32) {
	mNb := s.mNb
	mwb := s.mWb

	// row-range pass (Scene.java:932-950); note xs[i] is mutated in place by the
	// `var3[var4] += this.m_Nb` Java idiom, so xs now holds ABSOLUTE rows.
	xs[0] += mNb
	s.rasterMinY = xs[0]
	s.rasterMaxY = xs[0]
	for i := int32(1); n > i; i++ {
		xs[i] += mNb
		v := xs[i]
		if v >= s.rasterMinY {
			if s.rasterMaxY < v {
				s.rasterMaxY = v
			}
		} else {
			s.rasterMinY = v
		}
	}
	if s.rasterMaxY >= mNb+mwb {
		s.rasterMaxY = mNb - 1 + mwb
	}
	if mNb-mwb > s.rasterMinY {
		s.rasterMinY = mNb - mwb
	}
	if s.rasterMinY >= s.rasterMaxY {
		return
	}

	// empty-span init (Scene.java:956-960)
	for row := s.rasterMinY; row < s.rasterMaxY; row++ {
		sl := s.scanlines[row]
		sl.mK = -655360
		sl.mD = 655360
	}

	// closing edge last->first (Scene.java:962-1012)
	last := n - 1
	yA := xs[0]
	yB := xs[last]
	if yA >= yB {
		if yB < yA {
			u := us[last] << 8
			uStep := ((us[0] - us[last]) << 8) / (yA - yB)
			v := vs[last] << 8
			vStep := ((vs[0] - vs[last]) << 8) / (yA - yB)
			if yA > s.rasterMaxY {
				yA = s.rasterMaxY
			}
			if yB < 0 {
				v -= vStep * yB
				u -= uStep * yB
				yB = 0
			}
			for row := yB; row <= yA; row++ {
				sl := s.scanlines[row]
				sl.mD = u
				sl.mK = u
				sl.mE = v
				sl.mL = v
				u += uStep
				v += vStep
			}
		}
	} else {
		u := us[0] << 8
		uStep := ((us[last] - us[0]) << 8) / (yB - yA)
		v := vs[0] << 8
		vStep := ((vs[last] - vs[0]) << 8) / (yB - yA)
		if yA < 0 {
			u -= yA * uStep
			v -= yA * vStep
			yA = 0
		}
		if yB > s.rasterMaxY {
			yB = s.rasterMaxY
		}
		for row := yA; row <= yB; row++ {
			sl := s.scanlines[row]
			sl.mE = v
			sl.mL = v
			sl.mD = u
			sl.mK = u
			u += uStep
			v += vStep
		}
	}

	// remaining edges i -> i+1, min/max-expanding each row (Scene.java:1014-1082)
	for i := int32(0); i < last; i++ {
		ya := xs[i]
		j := i + 1
		yb := xs[j]
		if yb <= ya {
			if ya > yb {
				u := us[j] << 8
				uStep := ((us[i] - us[j]) << 8) / (ya - yb)
				v := vs[j] << 8
				vStep := ((vs[i] - vs[j]) << 8) / (ya - yb)
				if yb < 0 {
					u -= uStep * yb
					v -= yb * vStep
					yb = 0
				}
				if ya > s.rasterMaxY {
					ya = s.rasterMaxY
				}
				for row := yb; ya >= row; row++ {
					sl := s.scanlines[row]
					if u < sl.mD {
						sl.mE = v
						sl.mD = u
					}
					if u > sl.mK {
						sl.mL = v
						sl.mK = u
					}
					v += vStep
					u += uStep
				}
			}
		} else {
			u := us[i] << 8
			uStep := ((us[j] - us[i]) << 8) / (yb - ya)
			v := vs[i] << 8
			vStep := ((vs[j] - vs[i]) << 8) / (yb - ya)
			if yb > s.rasterMaxY {
				yb = s.rasterMaxY
			}
			if ya < 0 {
				u -= ya * uStep
				v -= ya * vStep
				ya = 0
			}
			for row := ya; yb >= row; row++ {
				sl := s.scanlines[row]
				if u > sl.mK {
					sl.mK = u
					sl.mL = v
				}
				if u < sl.mD {
					sl.mD = u
					sl.mE = v
				}
				v += vStep
				u += uStep
			}
		}
	}
	if mNb-mwb > s.rasterMinY {
		s.rasterMinY = mNb - mwb
	}
}

// =============================================================================
// fillFace — Scene.setFrustum(int[] var1, RSModel model, int var3, int var4,
// int var5, int[] var6, int[] var7, int var8, int var9) (Scene.java:1843-2312):
// the per-face span filler. It is invoked AFTER scanEdges has populated
// s.scanlines[rasterMinY..rasterMaxY]. It computes the perspective texture
// gradients in camera space, then walks each row and dispatches to the right
// Shader variant (texture size class × has-alpha × model blend mode), or to the
// flat-colour gradient fill when fill < 0.
//
//	ys    (Java var1): per-vertex camera-space Y == m_Vb == vertYRot
//	model (the Scene's current model)
//	onlyTri (Java var3): 1 when this is the single textured triangle of the face
//	nVerts (Java var4): face vertex count (used as last-index = nVerts-1)
//	fill  (Java var5): texture id (>=0) or -(colour+1) flat-colour key (<0)
//	zs    (Java var6): per-vertex camera-space Z == m_J  == vertZRot
//	xs    (Java var7): per-vertex camera-space X == m_Qb == vertXRot
//
// CALL-SITE BINDING (Scene.java:2846): endScene calls
//
//	setFrustum(m_Vb, model, 1, faceVertCount, poly.m_b, m_J, m_Qb, 0, 0)
//
// where m_Qb[i]=vertXRot, m_Vb[i]=vertYRot, m_J[i]=vertZRot (Scene.java:2765-67).
// So scene.go must call fillFace(m_Vb, model, 1, vertCount, fill, m_J, m_Qb).
// var8/var9 are scratch the Java method reuses internally; not inputs.
// =============================================================================
func (s *Scene) fillFace(ys []int32, model *Model, onlyTri, nVerts, fill int32, zs, xs []int32) {
	if fill == -2 {
		return // Scene.java:1847 guard
	}

	if fill >= 0 {
		// =====================================================================
		// TEXTURED FACE (Scene.java:1858-2175)
		// =====================================================================
		if fill >= int32(s.textureCount) {
			fill = 0
		}
		s.ensureTexture(int(fill), true) // Scene.b(var5,true), Scene.java:1863

		x0 := xs[0]
		y0 := ys[0]
		z0 := zs[0]
		dx1 := x0 - xs[1]
		dy1 := y0 - ys[1]
		nVerts-- // var4-- (last vertex index)
		dz1 := z0 - zs[1]
		dx2 := xs[nVerts] - x0
		dy2 := ys[nVerts] - y0
		dz2 := zs[nVerts] - z0

		vp := s.rot1024VpSrc

		var (
			a, aStep, b, bStep, c, cStep int32 // var19/21, var22/24, var25/27
			aMip, bMip, cMip             int32 // var20, var23, var26 (per-quad mip steps)
			aMipS, bMipS, cMipS          int32 // var28, var29, var30 (>>4 mip steps)
			rowOff, stride, dst          int32 // var31, var32, var33
			step                         int32 // var34 (1 or 2 for interlace)
		)

		if s.mHb[fill] == 1 {
			// ---- 128px texture gradients (Scene.java:1895-1913) ----
			// Java shift-count grouping: additive binds tighter than shift, so
			// e.g. `<< 4 - vp + 5 + 7` is `<< (4 - vp + 5 + 7)`. See
			// tex128Gradients for the full operand binding.
			a, aMip, aStep, b, bMip, bStep, c, cMip, cStep =
				s.tex128Gradients(x0, y0, z0, dx1, dy1, dz1, dx2, dy2, dz2, vp)
			aMipS = aMip >> 4
			bMipS = bMip >> 4
			cMipS = cMip >> 4
			rowOff = s.rasterMinY - s.mNb
			stride = s.mVb
			dst = stride*s.rasterMinY + s.mZb
			b += rowOff * bStep
			step = 1
			c += cStep * rowOff
			a += aStep * rowOff
			if s.interlace {
				if (s.rasterMinY & 1) == 1 {
					b += bStep
					dst += stride
					a += aStep
					c += cStep
					s.rasterMinY++
				}
				cStep <<= 1
				bStep <<= 1
				step = 2
				aStep <<= 1
				stride <<= 1
			}
			s.fill128(model, fill, a, aStep, aMip, aMipS, b, bStep, bMip, bMipS, c, cStep, cMip, cMipS, dst, stride, step)
			return
		}

		// ---- 64px texture gradients (Scene.java:2037-2070) ----
		a, aMip, aStep, b, bMip, bStep, c, cMip, cStep =
			s.tex64Gradients(x0, y0, z0, dx1, dy1, dz1, dx2, dy2, dz2, vp)
		aMipS = aMip >> 4
		bMipS = bMip >> 4
		cMipS = cMip >> 4
		rowOff = s.rasterMinY - s.mNb
		stride = s.mVb
		dst = stride*s.rasterMinY + s.mZb
		b += rowOff * bStep
		step = 1
		a += rowOff * aStep
		c += cStep * rowOff
		if s.interlace {
			if (1 & s.rasterMinY) == 1 {
				b += bStep
				c += cStep
				a += aStep
				s.rasterMinY++
				dst += stride
			}
			aStep <<= 1
			stride <<= 1
			cStep <<= 1
			step = 2
			bStep <<= 1
		}
		s.fill64(model, fill, a, aStep, aMip, aMipS, b, bStep, bMip, bMipS, c, cStep, cMip, cMipS, dst, stride, step)
		return
	}

	// =====================================================================
	// FLAT-COLOUR FACE (Scene.java:2176-2298): gradient-ramp gouraud fill.
	// =====================================================================
	s.fillFlat(model, onlyTri, fill)
	_ = nVerts
}

// fillFlat is the untextured branch of Scene.setFrustum (Scene.java:2176-2298):
// a 256-entry brightness ramp is looked up (or built) for the flat-colour key
// `fill` (= -(colour)-1 ... actually fill<0 here and var5=-1-fill recovers a
// 5:5:5 packed colour), then each row span is gouraud-filled by walking the ramp
// with the per-row shade endpoints.
//
//	model.mC  (Java model.m_cb): additive sprite blend -> blendRamp (GC.a)
//	!subtractiveBlend (Java !this.m_Ub): normal -> copyBlock16
//	else (m_Ub true; never in this client): copyBlock4
//
// onlyTri (Java var3) is forwarded to copyBlock16/blendRamp as the trailing
// "var3-1" arg (an unused obfuscation token in those helpers).
func (s *Scene) fillFlat(model *Model, onlyTri, fill int32) {
	// ---- ramp lookup / build (Scene.java:2177-2201) ----
	ramp := s.gradientRamp(fill)

	stride := s.mVb
	rowBase := s.rasterMinY*stride + s.mZb
	step := int32(1)
	if s.interlace {
		if (s.rasterMinY & 1) == 1 {
			s.rasterMinY++
			rowBase += stride
		}
		stride <<= 1
		step = 2
	}

	mA := s.mA
	if model.mC {
		// additive blend (Scene.java:2217-2243) -> GraphicsController.a
		for row := s.rasterMinY; s.rasterMaxY > row; row += step {
			sl := s.scanlines[row]
			left := sl.mD >> 8
			right := sl.mK >> 8
			w := right - left
			if w > 0 {
				shadeL := sl.mE
				shadeStep := (sl.mL - shadeL) / w
				if left < -mA {
					shadeL += shadeStep * (-mA - left)
					left = -mA
					w = right - left
				}
				if mA < right {
					right = mA
					w = right - left
				}
				blendRamp(shadeL, ramp, -w, s.pixelData, 0, shadeStep, left+rowBase, onlyTri-1)
			}
			rowBase += stride
		}
		return
	}

	if !subtractiveBlend {
		// normal (Scene.java:2244-2270) -> copyBlock16
		for row := s.rasterMinY; s.rasterMaxY > row; row += step {
			sl := s.scanlines[row]
			left := sl.mD >> 8
			right := sl.mK >> 8
			w := right - left
			if w > 0 {
				shadeL := sl.mE
				shadeStep := (sl.mL - shadeL) / w
				if left < -mA {
					shadeL += (-left - mA) * shadeStep
					left = -mA
					w = right - left
				}
				if mA < right {
					right = mA
					w = right - left
				}
				copyBlock16(0, shadeStep, -w, s.pixelData, ramp, shadeL, rowBase+left, 418609192)
			}
			rowBase += stride
		}
		return
	}

	// subtractive (Scene.java:2271-2297) -> copyBlock4 (m_Ub; unused in client)
	for row := s.rasterMinY; row < s.rasterMaxY; row += step {
		sl := s.scanlines[row]
		left := sl.mD >> 8
		right := sl.mK >> 8
		w := right - left
		if w > 0 {
			shadeL := sl.mE
			shadeStep := (sl.mL - shadeL) / w
			if left < -mA {
				shadeL += (-mA - left) * shadeStep
				left = -mA
				w = right - left
			}
			if mA < right {
				right = mA
				w = right - left
			}
			copyBlock4(shadeStep, 0, ramp, shadeL, left+rowBase, s.pixelData, -w, 82)
		}
		rowBase += stride
	}
}

// gradientRamp resolves the 256-entry brightness ramp for the flat-colour key
// `fill` (<0). It scans gradientKeys (Scene.m_v) for a cached ramp; on a miss it
// evicts a random slot, decodes `-1 - fill` into a 5:5:5 colour, and builds the
// ramp `ramp[255-i] = colour scaled by (i*i)/65536` (Scene.java:2177-2200).
func (s *Scene) gradientRamp(fill int32) []int32 {
	count := int32(len(s.gradientCache)) // Scene.m_ib (=50)
	for i := int32(0); i < count; i++ {
		if s.gradientKeys[i] == fill {
			return s.gradientCache[i]
		}
		if i == count-1 {
			slot := int32(rngFloat() * float64(count)) // (int)(Math.random()*m_ib)
			s.gradientKeys[slot] = fill
			key := -1 - fill
			r := ((32025 & key) >> 10) * 8
			g := ((1019 & key) >> 5) * 8
			b := (31 & key) * 8
			ramp := s.gradientCache[slot]
			for j := int32(0); j < 256; j++ {
				jj := j * j
				rr := r * jj / 65536
				gg := jj * g / 65536
				bb := b * jj / 65536
				ramp[255-j] = bb + (gg << 8) + (rr << 16)
			}
			return ramp
		}
	}
	return s.gradientCache[0] // unreachable when count>0; mirrors Java fallthrough
}

// copyBlock16 — MiscFunctions.copyBlock16 (MiscFunctions.java:172-224): the
// normal flat span fill. Walks the 256-entry ramp `src` by `srcStride<<2` per 4
// pixels (sampling src[(srcHead>>8)&0xff]) and writes `negCount` (negated)
// pixels into dest from destHead. var7 (418609192) is an unused obfuscation
// token. The 16-pixel unroll re-samples the ramp every 4 writes.
func copyBlock16(val, srcStride, negCount int32, dest, src []int32, srcHead, destHead, var7 int32) {
	if negCount >= 0 {
		return
	}
	val = src[255&(srcHead>>8)]
	srcStride <<= 2
	srcHead += srcStride
	negCap := negCount / 16
	for i := negCap; i < 0; i++ {
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		val = src[(srcHead&0xFF00)>>8]
		srcHead += srcStride
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		val = src[(0xFF00&srcHead)>>8]
		dest[destHead] = val
		destHead++
		srcHead += srcStride
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		val = src[0xFF&(srcHead>>8)]
		srcHead += srcStride
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		dest[destHead] = val
		destHead++
		val = src[(0xFF00&srcHead)>>8]
		srcHead += srcStride
	}
	negCap = -(negCount % 16)
	for v9 := int32(0); negCap > v9; v9++ {
		dest[destHead] = val
		destHead++
		if (3 & v9) == 3 {
			val = src[255&(srcHead>>8)]
			srcHead += srcStride
		}
	}
	_ = var7
}

// copyBlock4 — MiscFunctions.copyBlock4 (MiscFunctions.java:88-134): the
// subtractive (m_Ub) flat span fill, doubled-up ramp sampling. Unused by this
// client (m_Ub is always false) but ported for completeness. var7 (82) is an
// unused obfuscation token.
func copyBlock4(srcStep, val int32, src []int32, srcI, destI int32, dest []int32, negatedCount, var7 int32) {
	if negatedCount >= 0 {
		return
	}
	val = src[(0xFF00&srcI)>>8]
	srcStep <<= 1
	srcI += srcStep
	negCount := negatedCount / 8
	for i := negCount; i < 0; i++ {
		dest[destI] = val
		destI++
		dest[destI] = val
		destI++
		val = src[(0xFF00&srcI)>>8]
		srcI += srcStep
		dest[destI] = val
		destI++
		dest[destI] = val
		destI++
		val = src[(srcI&0xFF00)>>8]
		dest[destI] = val
		destI++
		srcI += srcStep
		dest[destI] = val
		destI++
		val = src[(0xFF00&srcI)>>8]
		srcI += srcStep
		dest[destI] = val
		destI++
		dest[destI] = val
		destI++
		val = src[(srcI&0xFF00)>>8]
		srcI += srcStep
	}
	negCount = -(negatedCount % 8)
	for i := int32(0); negCount > i; i++ {
		dest[destI] = val
		destI++
		if (i & 1) == 1 {
			val = src[srcI>>8&0xFF]
			srcI += srcStep
		}
	}
	_ = var7
}

// blendRamp — GraphicsController.a (GraphicsController.java:85-140): the additive
// (model.m_cb) flat span fill. Same ramp walk as copyBlock16 but each write
// averages the ramp colour into the existing dest pixel: dest = texel +
// ((dest>>1) & 0x7f7f7f). Reproduces the dest[i++] write / dest[i] read
// post-increment subtlety (see shadeScanlineTexBlend128) — the RHS reads
// var3[var6] AFTER var6 is incremented.
//
// Java signature: a(int var0,int[] var1,int var2,int[] var3,int var4,int var5,
//
//	int var6,int var7)
func blendRamp(var0 int32, var1 []int32, var2 int32, var3 []int32, var4, var5, var6, var7 int32) {
	if var2 >= 0 {
		return
	}
	var4 = var1[(0xffa7&var0)>>8]
	var5 <<= 2
	var0 += var5
	var8 := var2 / 16
	var w int32
	var var9 int32
	for var9 = var8; var9 < 0; var9++ {
		w = var6
		var6++
		var3[w] = var4 + (0x7F7F7F & (var3[var6] >> 1))
		w = var6
		var6++
		var3[w] = var4 + (0x7F7F7F & (var3[var6] >> 1))
		w = var6
		var6++
		var3[w] = var4 + ((var3[var6] & 0xFEFEFF) >> 1)
		w = var6
		var6++
		var3[w] = var4 + ((0xFEFEFF & var3[var6]) >> 1)
		var4 = var1[255&(var0>>8)]
		var0 += var5
		w = var6
		var6++
		var3[w] = ((var3[var6] >> 1) & 0x7F7F7F) + var4
		w = var6
		var6++
		var3[w] = ((var3[var6] >> 1) & 0x7F7F7F) + var4
		w = var6
		var6++
		var3[w] = ((0xFEFEFF & var3[var6]) >> 1) + var4
		w = var6
		var6++
		var3[w] = ((var3[var6] >> 1) & 0x7F7F7F) + var4
		var4 = var1[var0>>8&255]
		w = var6
		var6++
		var3[w] = ((var3[var6] & 0xFEFEFE) >> 1) + var4
		var0 += var5
		w = var6
		var6++
		var3[w] = ((0xFEFEFF & var3[var6]) >> 1) + var4
		w = var6
		var6++
		var3[w] = var4 + ((var3[var6] >> 1) & 0x7F7F7F)
		w = var6
		var6++
		var3[w] = var4 + (0x7F7F7F & (var3[var6] >> 1))
		var4 = var1[(0xff16&var0)>>8]
		var0 += var5
		w = var6
		var6++
		var3[w] = ((var3[var6] & 0xFEFEFF) >> 1) + var4
		w = var6
		var6++
		var3[w] = var4 + ((var3[var6] >> 1) & 0x7F7F7F)
		w = var6
		var6++
		var3[w] = var4 + ((var3[var6] & 0xFEFEFF) >> 1)
		w = var6
		var6++
		var3[w] = var4 + ((0xFEFEFF & var3[var6]) >> 1)
		var4 = var1[var0>>8&255]
		var0 += var5
	}
	var8 = -(var2 % 16)
	for var9 = var7; var8 > var9; var9++ {
		w = var6
		var6++
		var3[w] = ((var3[var6] >> 1) & 0x7F7F7F) + var4
		if (3 & var9) == 3 {
			var4 = var1[(var0&0xff38)>>8]
			var0 += var5
			var0 += var5
		}
	}
}

// tex128Gradients computes the nine perspective gradient accumulators for a
// 128px (class-1) texture (Scene.java:1895-1903). Returned in order:
// a,aMip,aStep (var19,var20,var21), b,bMip,bStep (var22,var23,var24),
// c,cMip,cStep (var25,var26,var27).
//
// Exact Java (var10=z0, var11=x0, var12=y0; var13=dz1, var14=dx1, var15=dy1;
// var16=dz2, var17=dx2, var18=dy2):
//
//	var19 = var11*var16 - var17*var10 << 12
//	var20 = var17*var12 - var18*var11 << 4 - vp + 5 + 7
//	var21 = var18*var10 - var16*var12 << 7 - vp + 5
//	var22 = var11*var13 - var10*var14 << 12
//	var23 = var14*var12 - var15*var11 << 5 - vp + 11
//	var24 = var10*var15 - var12*var13 << 7 + (5 - vp)
//	var25 = var16*var14 - var17*var13 << 5
//	var26 = var15*var17 - var14*var18 << 4 + (5 - vp)
//	var27 = var13*var18 - var15*var16 >> vp - 5
//
// where var10..var18 in Scene map to: var10=zs[0]=z0, var11=xs[0]=x0,
// var12=ys[0]... wait — careful: in Scene var7=xs(camX), var1=ys(camY),
// var6=zs(camZ); var10=var7[0]=x0(camX), var11=var1[0]=y0(camY),
// var12=var6[0]=z0(camZ). We bind to that naming here: A=camX, B=camY, C=camZ.
func (s *Scene) tex128Gradients(camX, camY, camZ, dCamX1, dCamY1, dCamZ1, dCamX2, dCamY2, dCamZ2, vp int32) (a, aMip, aStep, b, bMip, bStep, c, cMip, cStep int32) {
	// Faithful operand binding (see RSModel.rotate1024 axis note): Scene's var10
	// is camX, var11 is camY, var12 is camZ; var13=dCamX1, var14=dCamY1,
	// var15=dCamZ1; var16=dCamX2, var17=dCamY2, var18=dCamZ2.
	v10, v11, v12 := camX, camY, camZ
	v13, v14, v15 := dCamX1, dCamY1, dCamZ1
	v16, v17, v18 := dCamX2, dCamY2, dCamZ2

	a = (v11*v16 - v17*v10) << 12
	aMip = (v17*v12 - v18*v11) << (4 - vp + 5 + 7)
	aStep = (v18*v10 - v16*v12) << (7 - vp + 5)
	b = (v11*v13 - v10*v14) << 12
	bMip = (v14*v12 - v15*v11) << (5 - vp + 11)
	bStep = (v10*v15 - v12*v13) << (7 + (5 - vp))
	c = (v16*v14 - v17*v13) << 5
	cMip = (v15*v17 - v14*v18) << (4 + (5 - vp))
	cStep = (v13*v18 - v15*v16) >> (vp - 5)
	return
}

// tex64Gradients computes the nine accumulators for a 64px (class-0) texture
// (Scene.java:2037-2045). Java:
//
//	var19 = var16*var11 - var10*var17 << 11
//	var20 = var12*var17 - var18*var11 << 4 + 6 + (5 - vp)
//	var21 = var18*var10 - var16*var12 << 11 - vp
//	var22 = var11*var13 - var14*var10 << 11
//	var23 = var12*var14 - var11*var15 << 4 - vp + 11
//	var24 = var15*var10 - var12*var13 << 11 - vp
//	var25 = var16*var14 - var17*var13 << 5
//	var26 = var17*var15 - var14*var18 << 4 + (5 - vp)
//	var27 = var18*var13 - var16*var15 >> vp - 5
func (s *Scene) tex64Gradients(camX, camY, camZ, dCamX1, dCamY1, dCamZ1, dCamX2, dCamY2, dCamZ2, vp int32) (a, aMip, aStep, b, bMip, bStep, c, cMip, cStep int32) {
	v10, v11, v12 := camX, camY, camZ
	v13, v14, v15 := dCamX1, dCamY1, dCamZ1
	v16, v17, v18 := dCamX2, dCamY2, dCamZ2

	a = (v16*v11 - v10*v17) << 11
	aMip = (v12*v17 - v18*v11) << (4 + 6 + (5 - vp))
	aStep = (v18*v10 - v16*v12) << (11 - vp)
	b = (v11*v13 - v14*v10) << 11
	bMip = (v12*v14 - v11*v15) << (4 - vp + 11)
	bStep = (v15*v10 - v12*v13) << (11 - vp)
	c = (v16*v14 - v17*v13) << 5
	cMip = (v17*v15 - v14*v18) << (4 + (5 - vp))
	cStep = (v18*v13 - v16*v15) >> (vp - 5)
	return
}

// fill128 walks the row spans for a 128px texture and dispatches to the right
// Shader variant (Scene.java:1930-2035). model.mKb selects the wall-blend path;
// otherwise mS[fill] selects the transparent-skip path vs the opaque path.
func (s *Scene) fill128(model *Model, fill, a, aStep, aMip, aMipS, b, bStep, bMip, bMipS, c, cStep, cMip, cMipS, dst, stride, step int32) {
	res := s.resourceDatabase[fill]
	mA := s.mA

	if !model.mKb {
		if s.mS[fill] {
			// transparent-skip 128px (Scene.java:1932-1964) -> Shader.java:8
			for row := s.rasterMinY; row < s.rasterMaxY; row += step {
				sl := s.scanlines[row]
				left := sl.mD >> 8
				right := sl.mK >> 8
				w := right - left
				if w <= 0 {
					b += bStep
					a += aStep
					dst += stride
					c += cStep
				} else {
					shadeL := sl.mE
					shadeStep := (sl.mL - shadeL) / w
					if -mA > left {
						shadeL += (-left - mA) * shadeStep
						left = -mA
						w = right - left
					}
					if right > mA {
						right = mA
						w = right - left
					}
					// @ Scene.java:1956 — exact Java arg order.
					shadeScanlineTexTransp128(bMip, 10, 0, 0, s.pixelData, c+left*cMipS, shadeL,
						left*aMipS+a, b+left*bMipS, left+dst, cMip, shadeStep, 0, aMip, res, w)
					dst += stride
					b += bStep
					c += cStep
					a += aStep
				}
			}
		} else {
			// opaque 128px (Scene.java:1965-1999) -> Shader.java:203 (byte 50)
			for row := s.rasterMinY; row < s.rasterMaxY; row += step {
				sl := s.scanlines[row]
				left := sl.mD >> 8
				right := sl.mK >> 8
				w := right - left
				if w <= 0 {
					c += cStep
					a += aStep
					dst += stride
					b += bStep
				} else {
					shadeL := sl.mE
					shadeStep := (sl.mL - shadeL) / w
					if left < -mA {
						shadeL += (-mA - left) * shadeStep
						left = -mA
						w = right - left
					}
					if mA < right {
						right = mA
						w = right - left
					}
					// @ Scene.java:1990 — exact Java arg order (var2 byte=50).
					shadeScanlineTexOpaque128(b+bMipS*left, aMip, 50, c+left*cMipS, shadeL, shadeStep<<2, res,
						left+dst, left*aMipS+a, cMip, 0, 0, s.pixelData, bMip, w)
					a += aStep
					c += cStep
					dst += stride
					b += bStep
				}
			}
		}
		return
	}

	// model.mKb wall-blend 128px (Scene.java:2001-2034) -> Shader.java:346 (byte 119)
	for row := s.rasterMinY; row < s.rasterMaxY; row += step {
		sl := s.scanlines[row]
		left := sl.mD >> 8
		right := sl.mK >> 8
		w := right - left
		if w > 0 {
			shadeL := sl.mE
			shadeStep := (sl.mL - shadeL) / w
			if -mA > left {
				shadeL += shadeStep * (-left - mA)
				left = -mA
				w = right - left
			}
			if right > mA {
				right = mA
				w = right - left
			}
			// @ Scene.java:2021 — exact Java arg order (var14 byte=119).
			shadeScanlineTexBlend128(dst+left, b+left*bMipS, a+left*aMipS, 0, shadeL, bMip, 0,
				c+left*cMipS, aMip, shadeStep<<2, res, w, cMip, s.pixelData, 119)
			dst += stride
			b += bStep
			a += aStep
			c += cStep
		} else {
			dst += stride
			c += cStep
			a += aStep
			b += bStep
		}
	}
}

// fill64 walks the row spans for a 64px texture and dispatches (Scene.java:
// 2072-2174). model.mKb selects the blend path; otherwise mS[fill] selects
// transparent-skip vs opaque.
func (s *Scene) fill64(model *Model, fill, a, aStep, aMip, aMipS, b, bStep, bMip, bMipS, c, cStep, cMip, cMipS, dst, stride, step int32) {
	res := s.resourceDatabase[fill]
	mA := s.mA

	if model.mKb {
		// blend 64px (Scene.java:2072-2105) -> Shader.java:511 (boolean false)
		for row := s.rasterMinY; s.rasterMaxY > row; row += step {
			sl := s.scanlines[row]
			left := sl.mD >> 8
			right := sl.mK >> 8
			w := right - left
			if w <= 0 {
				b += bStep
				c += cStep
				dst += stride
				a += aStep
			} else {
				shadeL := sl.mE
				shadeStep := (sl.mL - shadeL) / w
				if left < -mA {
					shadeL += shadeStep * (-mA - left)
					left = -mA
					w = right - left
				}
				if right > mA {
					right = mA
					w = right - left
				}
				// @ Scene.java:2097 — exact Java arg order (var11 boolean=false).
				shadeScanlineTexBlend64(s.pixelData, bMip, cMip, left*cMipS+c, shadeStep, shadeL,
					left+dst, w, aMipS*left+a, 0, res, false, aMip, left*bMipS+b, 0)
				dst += stride
				c += cStep
				a += aStep
				b += bStep
			}
		}
		return
	}

	if !s.mS[fill] {
		// opaque 64px (Scene.java:2106-2139) -> Shader.java:649 (int 1121159302)
		for row := s.rasterMinY; s.rasterMaxY > row; row += step {
			sl := s.scanlines[row]
			left := sl.mD >> 8
			right := sl.mK >> 8
			w := right - left
			if w > 0 {
				shadeL := sl.mE
				shadeStep := (sl.mL - shadeL) / w
				if -mA > left {
					shadeL += (-mA - left) * shadeStep
					left = -mA
					w = right - left
				}
				if mA < right {
					right = mA
					w = right - left
				}
				// @ Scene.java:2126 — exact Java arg order (var1=1121159302).
				shadeScanlineTexOpaque64(shadeStep, 1121159302, bMip, left*bMipS+b, aMip, res, shadeL,
					0, a+aMipS*left, 0, s.pixelData, dst+left, c+left*cMipS, cMip, w)
				dst += stride
				b += bStep
				a += aStep
				c += cStep
			} else {
				dst += stride
				c += cStep
				b += bStep
				a += aStep
			}
		}
		return
	}

	// transparent-skip 64px (Scene.java:2140-2173) -> Shader.java:775 (byte 25)
	for row := s.rasterMinY; s.rasterMaxY > row; row += step {
		sl := s.scanlines[row]
		left := sl.mD >> 8
		right := sl.mK >> 8
		w := right - left
		if w <= 0 {
			c += cStep
			dst += stride
			a += aStep
			b += bStep
		} else {
			shadeL := sl.mE
			shadeStep := (sl.mL - shadeL) / w
			if left < -mA {
				shadeL += shadeStep * (-mA - left)
				left = -mA
				w = right - left
			}
			if right > mA {
				right = mA
				w = right - left
			}
			// @ Scene.java:2165 — exact Java arg order (var3 byte=25).
			shadeScanlineTexTransp64(w, cMipS*left+c, 0, 25, 0, aMip, cMip, shadeStep, res, s.pixelData,
				left+dst, left*aMipS+a, 0, bMip, shadeL, bMipS*left+b)
			c += cStep
			b += bStep
			dst += stride
			a += aStep
		}
	}
}

// =============================================================================
// Shader.java family — perspective-correct textured span writers.
//
// Each function is a verbatim port of one Shader.shadeScanline overload, keeping
// the Java parameter list in EXACT order (named varN) so the body reads 1:1 with
// the source and the call sites in fill128/fill64 pass arguments in the same
// order as Scene.java. The 16-texel unrolled inner blocks do an affine walk
// inside each block, re-dividing the perspective numerators once per block; the
// texel is fetched from the shade-mip buffer and brightness-selected with a
// logical >>> (lshr). In the has-alpha (transp) variants a texel value of 0 is
// skipped. The Java try/catch + the unreachable obfuscation-guard recursion
// (`if (var1 != 10) shadeScanline(...)`) are dropped: they never fire and Go
// panics already carry a stack trace.
// =============================================================================

// shadeScanlineTexTransp128 — Shader.java:8 (the m_S 128px has-alpha path).
// @ Scene.java:1956. Texel 0 == skip.
func shadeScanlineTexTransp128(var0, var1, var2, var3 int32, var4 []int32, var5, var6,
	var7, var8, var9, var10, var11, var12, var13 int32, var14 []int32, var15 int32) {
	if var15 <= 0 {
		return
	}
	var16 := int32(0)
	var17 := int32(0)
	var11 <<= 2
	if var5 != 0 {
		var17 = var8 / var5 << 7
		var16 = var7 / var5 << 7
	}
	if var16 < 0 {
		var16 = 0
	} else if var16 > 0x3F80 {
		var16 = 0x3F80
	}
	for var20 := var15; var20 > 0; var20 -= 16 {
		var7 += var13
		var3 = var17
		var5 += var10
		var2 = var16
		var8 += var0
		if var5 != 0 {
			var16 = var7 / var5 << 7
			var17 = var8 / var5 << 7
		}
		if var16 >= 0 {
			if var16 > 0x3F80 {
				var16 = 0x3F80
			}
		} else {
			var16 = 0
		}
		var18 := (var16 - var2) >> 4
		var19 := (var17 - var3) >> 4
		var21 := var6 >> 23
		var2 += 6291456 & var6
		var6 += var11
		if var20 < 16 {
			for var22 := int32(0); var20 > var22; var22++ {
				if var12 = lshr(var14[(var3&0x3F80)+(var2>>7)], var21); var12 != 0 {
					var4[var9] = var12
				}
				var9++
				var2 += var18
				var3 += var19
				if (var22 & 3) == 3 {
					var2 = (var6 & 6291456) + (16383 & var2)
					var21 = var6 >> 23
					var6 += var11
				}
			}
		} else {
			if var12 = lshr(var14[(var3&0x3F80)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var2 += var18
			var9++
			var3 += var19
			if var12 = lshr(var14[(var2>>7)+(0x3F80&var3)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var9++
			var3 += var19
			var2 += var18
			if var12 = lshr(var14[(var2>>7)+(0x3F80&var3)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var9++
			var3 += var19
			var2 += var18
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var2 += var18
			var3 += var19
			var9++
			var21 = var6 >> 23
			var2 = (var6 & 6291456) + (16383 & var2)
			var6 += var11
			if var12 = lshr(var14[(var3&0x3F80)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var2 += var18
			var9++
			var3 += var19
			if var12 = lshr(var14[(0x3F80&var3)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var9++
			var2 += var18
			var3 += var19
			if var12 = lshr(var14[(var2>>7)+(0x3F80&var3)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var2 += var18
			var3 += var19
			var9++
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var3 += var19
			var9++
			var2 += var18
			var21 = var6 >> 23
			var2 = (var2 & 16383) + (6291456 & var6)
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var6 += var11
			var9++
			var2 += var18
			var3 += var19
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var3 += var19
			var9++
			var2 += var18
			if var12 = lshr(var14[(0x3F80&var3)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var3 += var19
			var2 += var18
			var9++
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var2 += var18
			var3 += var19
			var9++
			var2 = (var2 & 16383) + (var6 & 6291456)
			var21 = var6 >> 23
			if var12 = lshr(var14[(var3&0x3F80)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var6 += var11
			var3 += var19
			var9++
			var2 += var18
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var9++
			var2 += var18
			var3 += var19
			if var12 = lshr(var14[(var2>>7)+(var3&0x3F80)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var3 += var19
			var2 += var18
			var9++
			if var12 = lshr(var14[(0x3F80&var3)+(var2>>7)], var21); var12 != 0 {
				var4[var9] = var12
			}
			var9++
		}
	}
	_ = var0
	_ = var1
}

// shadeScanlineTexOpaque128 — Shader.java:203 (opaque 128px walls, var2==50).
// @ Scene.java:1990. No alpha test; every texel overwrites dest.
// Java signature: shadeScanline(int var0,int var1,byte var2,int var3,int val,
//
//	int valStep,int[] src,int dH,int var8,int var9,int high,int low,int[] dest,
//	int var13,int var14)
func shadeScanlineTexOpaque128(var0, var1, var2, var3, val, valStep int32, src []int32,
	dH, var8, var9, high, low int32, dest []int32, var13, var14 int32) {
	if var14 <= 0 {
		return
	}
	if var2 != 50 {
		return
	}
	var15 := int32(0)
	var16 := int32(0)
	if var3 != 0 {
		low = var8 / var3 << 7
		high = var0 / var3 << 7
	}
	shift := int32(0)
	if low < 0 {
		low = 0
	} else if low > 0x3F80 {
		low = 0x3F80
	}

	var3 += var9
	var0 += var13
	var8 += var1
	if var3 != 0 {
		var16 = var0 / var3 << 7
		var15 = var8 / var3 << 7
	}
	if var15 >= 0 {
		if var15 > 0x3F80 {
			var15 = 0x3F80
		}
	} else {
		var15 = 0
	}
	lowStep := (var15 - low) >> 4
	highStep := (var16 - high) >> 4

	var var20 int32
	for var20 = var14 >> 4; var20 > 0; var20-- {
		low += val & 6291456
		shift = val >> 23
		dest[dH] = lshr(src[(0x3F80&high)+(low>>7)], shift)
		dH++
		val += valStep
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		low = (6291456 & val) + (16383 & low)
		shift = val >> 23
		dest[dH] = lshr(src[(low>>7)+(high&0x3F80)], shift)
		dH++
		val += valStep
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(0x3F80&high)+(low>>7)], shift)
		dH++
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		dest[dH] = lshr(src[(0x3F80&high)+(low>>7)], shift)
		dH++
		high += highStep
		low += lowStep
		low = (val & 6291456) + (16383 & low)
		shift = val >> 23
		val += valStep
		dest[dH] = lshr(src[(high&0x3F80)+(low>>7)], shift)
		dH++
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(0x3F80&high)+(low>>7)], shift)
		dH++
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		dest[dH] = lshr(src[(low>>7)+(high&0x3F80)], shift)
		dH++
		low += lowStep
		high += highStep
		low = (16383 & low) + (6291456 & val)
		shift = val >> 23
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		val += valStep
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(low>>7)+(high&0x3F80)], shift)
		dH++
		low += lowStep
		high += highStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		high += highStep
		low += lowStep
		dest[dH] = lshr(src[(low>>7)+(0x3F80&high)], shift)
		dH++
		low = var15
		high = var16
		var0 += var13
		var3 += var9
		var8 += var1
		if var3 != 0 {
			var16 = var0 / var3 << 7
			var15 = var8 / var3 << 7
		}
		if var15 >= 0 {
			if var15 > 0x3F80 {
				var15 = 0x3F80
			}
		} else {
			var15 = 0
		}
		highStep = (var16 - high) >> 4
		lowStep = (var15 - low) >> 4
	}

	for var20 = 0; (15 & var14) > var20; var20++ {
		if (var20 & 3) == 0 {
			shift = val >> 23
			low = (val & 6291456) + (16383 & low)
			val += valStep
		}
		dest[dH] = lshr(src[(low>>7)+(high&0x3F80)], shift)
		dH++
		high += highStep
		low += lowStep
	}
}

// shadeScanlineTexOpaque64 — Shader.java:649 (opaque 64px floors,
// var1==1121159302). @ Scene.java:2126. dest[i] = src[texel] >>> mip; no dest
// read, so no post-increment subtlety. 64px: u>>6, v&4032(0xFC0), mip var6>>20.
// Java signature: shadeScanline(int var0,int var1,int var2,int var3,int var4,
//
//	int[] src,int var6,int var7,int var8,int var9,int[] dest,int var11,int var12,
//	int var13,int var14)
func shadeScanlineTexOpaque64(var0, var1, var2, var3, var4 int32, src []int32,
	var6, var7, var8, var9 int32, dest []int32, var11, var12, var13, var14 int32) {
	if var14 <= 0 {
		return
	}
	var15 := int32(0)
	var16 := int32(0)
	if var12 != 0 {
		var16 = var3 / var12 << 6
		var15 = var8 / var12 << 6
	}
	var0 <<= 2
	if var15 >= 0 {
		if var15 > 4032 {
			var15 = 4032
		}
	} else {
		var15 = 0
	}
	for var19 := var14; var19 > 0; var19 -= 16 {
		var12 += var13
		var8 += var4
		var3 += var2
		var9 = var15
		var7 = var16
		if var12 != 0 {
			var15 = var8 / var12 << 6
			var16 = var3 / var12 << 6
		}
		if var15 < 0 {
			var15 = 0
		} else if var15 > 4032 {
			var15 = 4032
		}
		var18 := (var16 - var7) >> 4
		var17 := (var15 - var9) >> 4
		var20 := var6 >> 20
		var9 += 786432 & var6
		var6 += var0
		if var19 >= 16 {
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(var7&4032)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(4032&var7)], var20)
			var11++
			var9 += var17
			var7 += var18
			dest[var11] = lshr(src[(var9>>6)+(4032&var7)], var20)
			var11++
			var9 += var17
			var7 += var18
			var20 = var6 >> 20
			var9 = (var6 & 786432) + (4095 & var9)
			var6 += var0
			dest[var11] = lshr(src[(4032&var7)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var9 += var17
			var7 += var18
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(4032&var7)], var20)
			var11++
			var9 += var17
			var7 += var18
			var20 = var6 >> 20
			var9 = (786432 & var6) + (4095 & var9)
			var6 += var0
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(var7&4032)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(4032&var7)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			var20 = var6 >> 20
			var9 = (4095 & var9) + (var6 & 786432)
			var6 += var0
			dest[var11] = lshr(src[(var7&4032)+(var9>>6)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(var7&4032)], var20)
			var11++
			var7 += var18
			var9 += var17
			dest[var11] = lshr(src[(var9>>6)+(var7&4032)], var20)
			var11++
			var9 += var17
			var7 += var18
			dest[var11] = lshr(src[(4032&var7)+(var9>>6)], var20)
			var11++
		} else {
			for var21 := int32(0); var21 < var19; var21++ {
				dest[var11] = lshr(src[(var9>>6)+(4032&var7)], var20)
				var11++
				var7 += var18
				var9 += var17
				if (3 & var21) == 3 {
					var20 = var6 >> 20
					var9 = (var6 & 786432) + (4095 & var9)
					var6 += var0
				}
			}
		}
	}
	_ = var1
}

// shadeScanlineTexTransp64 — Shader.java:775 (m_S has-alpha 64px, var3==25).
// @ Scene.java:2165. Texel 0 == skip. 64px: u>>6, v&4032, mip var14>>20.
// Java signature: shadeScanline(int var0,int var1,int var2,byte var3,int var4,
//
//	int var5,int var6,int var7,int[] var8,int[] var9,int var10,int var11,
//	int var12,int var13,int var14,int var15)
func shadeScanlineTexTransp64(var0, var1, var2, var3, var4, var5, var6, var7 int32,
	var8, var9 []int32, var10, var11, var12, var13, var14, var15 int32) {
	if var0 <= 0 {
		return
	}
	if var3 != 25 {
		return
	}
	var16 := int32(0)
	var17 := int32(0)
	if var1 != 0 {
		var16 = var11 / var1 << 6
		var17 = var15 / var1 << 6
	}
	var7 <<= 2
	if var16 >= 0 {
		if var16 > 4032 {
			var16 = 4032
		}
	} else {
		var16 = 0
	}
	for var20 := var0; var20 > 0; var20 -= 16 {
		var4 = var17
		var12 = var16
		var11 += var5
		var1 += var6
		var15 += var13
		if var1 != 0 {
			var17 = var15 / var1 << 6
			var16 = var11 / var1 << 6
		}
		if var16 < 0 {
			var16 = 0
		} else if var16 > 4032 {
			var16 = 4032
		}
		var19 := (var17 - var4) >> 4
		var18 := (var16 - var12) >> 4
		var12 += 786432 & var14
		var21 := var14 >> 20
		var14 += var7
		if var20 >= 16 {
			if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var4 += var19
			var12 += var18
			if var2 = lshr(var8[(var12>>6)+(var4&4032)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var10++
			var12 += var18
			if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var10++
			var12 += var18
			if var2 = lshr(var8[(4032&var4)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			var21 = var14 >> 20
			var12 = (786432 & var14) + (4095 & var12)
			var14 += var7
			if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var10++
			var12 += var18
			var12 = (var12 & 4095) + (var14 & 786432)
			var21 = var14 >> 20
			if var2 = lshr(var8[(var12>>6)+(var4&4032)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var14 += var7
			var10++
			var12 += var18
			var4 += var19
			if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var12 += var18
			var10++
			if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var12 += var18
			var4 += var19
			var10++
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			var21 = var14 >> 20
			var12 = (var14 & 786432) + (var12 & 4095)
			var14 += var7
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var12 += var18
			var10++
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
			var12 += var18
			var4 += var19
			if var2 = lshr(var8[(var4&4032)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var4 += var19
			var10++
			var12 += var18
			if var2 = lshr(var8[(4032&var4)+(var12>>6)], var21); var2 != 0 {
				var9[var10] = var2
			}
			var10++
		} else {
			for var22 := int32(0); var20 > var22; var22++ {
				if var2 = lshr(var8[(var12>>6)+(4032&var4)], var21); var2 != 0 {
					var9[var10] = var2
				}
				var10++
				var12 += var18
				var4 += var19
				if (3 & var22) == 3 {
					var21 = var14 >> 20
					var12 = (4095 & var12) + (var14 & 786432)
					var14 += var7
				}
			}
		}
	}
}

// shadeScanlineTexBlend128 — Shader.java:346 (model.m_Kb wall-blend 128px,
// var14==119). @ Scene.java:2021.
//
// CRITICAL Java evaluation-order subtlety (JLS 15.26.1): the source writes
// `var13[var0++] = bitwiseAnd(var13[var0] >> 1, ...) + texel`. The array
// subscript `var0++` is evaluated BEFORE the RHS, so the WRITE targets the old
// var0 while the READ of var13[var0] on the RHS sees the ALREADY-incremented
// var0 — i.e. it blends the texel into dest[i] using dest[i+1]'s colour. This is
// reproduced exactly below as: `w := var0; var0++; var13[w] = blend(var13[var0])
// + texel`. (Reading dest[i] instead would silently diverge from Java.)
//
// 0x7F7F7F (8355711) / 0xFEFEFF (16711423) / 0xFEFEFE (16711422) are the
// halve-each-channel masks; Go `&` binds tighter than `+`, matching Java.
//
// Java signature: shadeScanline(int var0,int var1,int var2,int var3,int var4,
//
//	int var5,int var6,int var7,int var8,int var9,int[] var10,int var11,int var12,
//	int[] var13,byte var14)
func shadeScanlineTexBlend128(var0, var1, var2, var3, var4, var5, var6, var7, var8, var9 int32,
	var10 []int32, var11, var12 int32, var13 []int32, var14 int32) {
	if var11 <= 0 {
		return
	}
	var15 := int32(0)
	var16 := int32(0)
	var19 := int32(0)
	if var7 != 0 {
		var3 = var1 / var7 << 7
		var6 = var2 / var7 << 7
	}
	var7 += var12
	if var6 >= 0 {
		if var6 > 0x3F80 {
			var6 = 0x3F80
		}
	} else {
		var6 = 0
	}
	var1 += var5
	var2 += var8
	if var7 != 0 {
		var15 = var2 / var7 << 7
		var16 = var1 / var7 << 7
	}
	if var15 >= 0 {
		if var15 > 0x3F80 {
			var15 = 0x3F80
		}
	} else {
		var15 = 0
	}
	var17 := (var15 - var6) >> 4
	var18 := (var16 - var3) >> 4

	var w int32
	var var20 int32
	for var20 = var11 >> 4; var20 > 0; var20-- {
		var19 = var4 >> 23
		var6 += var4 & 6291456
		var4 += var9
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(var6>>7)+(var3&0x3F80)], var19)
		var3 += var18
		var6 += var17
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(var6>>7)+(0x3F80&var3)], var19)
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = lshr(var10[(0x3F80&var3)+(var6>>7)], var19) + ((16711422 & var13[var0]) >> 1)
		var3 += var18
		var6 += var17
		w = var0
		var0++
		var13[w] = ((16711422 & var13[var0]) >> 1) + lshr(var10[(var3&0x3F80)+(var6>>7)], var19)
		var3 += var18
		var6 += var17
		var19 = var4 >> 23
		var6 = (var6 & 16383) + (var4 & 6291456)
		var4 += var9
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(0x3F80&var3)+(var6>>7)], var19)
		var3 += var18
		var6 += var17
		w = var0
		var0++
		var13[w] = lshr(var10[(var6>>7)+(var3&0x3F80)], var19) + (var13[var0]>>1)&8355711
		var3 += var18
		var6 += var17
		w = var0
		var0++
		var13[w] = ((var13[var0] & 16711423) >> 1) + lshr(var10[(var3&0x3F80)+(var6>>7)], var19)
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = lshr(var10[(var6>>7)+(0x3F80&var3)], var19) + ((var13[var0] & 16711423) >> 1)
		var6 += var17
		var3 += var18
		var6 = (16383 & var6) + (var4 & 6291456)
		var19 = var4 >> 23
		w = var0
		var0++
		var13[w] = ((16711423 & var13[var0]) >> 1) + lshr(var10[(var6>>7)+(var3&0x3F80)], var19)
		var4 += var9
		var3 += var18
		var6 += var17
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(var6>>7)+(0x3F80&var3)], var19)
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = lshr(var10[(var3&0x3F80)+(var6>>7)], var19) + (8355711 & (var13[var0] >> 1))
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = ((16711423 & var13[var0]) >> 1) + lshr(var10[(0x3F80&var3)+(var6>>7)], var19)
		var3 += var18
		var6 += var17
		var6 = (var6 & 16383) + (var4 & 6291456)
		var19 = var4 >> 23
		w = var0
		var0++
		var13[w] = (8355711 & (var13[var0] >> 1)) + lshr(var10[(var6>>7)+(var3&0x3F80)], var19)
		var4 += var9
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(var6>>7)+(0x3F80&var3)], var19)
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = lshr(var10[(var3&0x3F80)+(var6>>7)], var19) + ((var13[var0] >> 1) & 8355711)
		var6 += var17
		var3 += var18
		w = var0
		var0++
		var13[w] = ((var13[var0] >> 1) & 8355711) + lshr(var10[(var6>>7)+(0x3F80&var3)], var19)
		var7 += var12
		var1 += var5
		var2 += var8
		var3 = var16
		var6 = var15
		if var7 != 0 {
			var16 = var1 / var7 << 7
			var15 = var2 / var7 << 7
		}
		if var15 >= 0 {
			if var15 > 0x3F80 {
				var15 = 0x3F80
			}
		} else {
			var15 = 0
		}
		var18 = (var16 - var3) >> 4
		var17 = (var15 - var6) >> 4
	}

	for var20 = 0; (var11 & 15) > var20; var20++ {
		if (var20 & 3) == 0 {
			var6 = (var4 & 6291456) + (var6 & 16383)
			var19 = var4 >> 23
			var4 += var9
		}
		w = var0
		var0++
		var13[w] = lshr(var10[(var3&0x3F80)+(var6>>7)], var19) + ((var13[var0] & 16711422) >> 1)
		var6 += var17
		var3 += var18
	}
	_ = var14
}

// shadeScanlineTexBlend64 — Shader.java:511 (model.m_Kb wall-blend 64px,
// boolean var11). @ Scene.java:2097. 64px texture: u>>6, v&0xFC0, mip select
// var5>>20. Same dest[i] write / dest[i+1] read post-increment subtlety as the
// 128px blend (see shadeScanlineTexBlend128) — reproduced via `w := var6;
// var6++; var0[w] = blend(var0[var6]) + texel`.
// Java signature: shadeScanline(int[] var0,int var1,int var2,int var3,int var4,
//
//	int var5,int var6,int var7,int var8,int var9,int[] var10,boolean var11,
//	int var12,int var13,int var14)
func shadeScanlineTexBlend64(var0 []int32, var1, var2, var3, var4, var5, var6, var7, var8, var9 int32,
	var10 []int32, var11 bool, var12, var13, var14 int32) {
	if var7 <= 0 {
		return
	}
	var15 := int32(0)
	var16 := int32(0)
	if var3 != 0 {
		var16 = var13 / var3 << 6
		var15 = var8 / var3 << 6
	}
	var4 <<= 2
	if var15 < 0 {
		var15 = 0
	} else if var15 > 0xFC0 {
		var15 = 0xFC0
	}
	var w int32
	for var19 := var7; var19 > 0; var19 -= 16 {
		var3 += var2
		var14 = var15
		var8 += var12
		var9 = var16
		var13 += var1
		if var3 != 0 {
			var15 = var8 / var3 << 6
			var16 = var13 / var3 << 6
		}
		if var15 >= 0 {
			if var15 > 0xFC0 {
				var15 = 0xFC0
			}
		} else {
			var15 = 0
		}
		var18 := (var16 - var9) >> 4
		var17 := (var15 - var14) >> 4
		var20 := var5 >> 20
		var14 += var5 & 786432
		var5 += var4
		if var19 >= 16 {
			w = var6
			var6++
			var0[w] = ((var0[var6] >> 1) & 0x7f7f7f) + lshr(var10[(0xFC0&var9)+(var14>>6)], var20)
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = ((var0[var6] & 0xfefeff) >> 1) + lshr(var10[(0xFC0&var9)+(var14>>6)], var20)
			var9 += var18
			var14 += var17
			w = var6
			var6++
			var0[w] = ((0xfefeff & var0[var6]) >> 1) + lshr(var10[(var9&0xFC0)+(var14>>6)], var20)
			var9 += var18
			var14 += var17
			w = var6
			var6++
			var0[w] = ((var0[var6] & 0xfefeff) >> 1) + lshr(var10[(var14>>6)+(0xFC0&var9)], var20)
			var14 += var17
			var9 += var18
			var14 = (var5 & 786432) + (4095 & var14)
			var20 = var5 >> 20
			w = var6
			var6++
			var0[w] = lshr(var10[(var9&0xFC0)+(var14>>6)], var20) + ((var0[var6] & 0xFEFEFE) >> 1)
			var5 += var4
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = lshr(var10[(var14>>6)+(0xFC0&var9)], var20) + ((var0[var6] & 0xfefeff) >> 1)
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = lshr(var10[(0xFC0&var9)+(var14>>6)], var20) + ((var0[var6] & 0xfefeff) >> 1)
			var9 += var18
			var14 += var17
			w = var6
			var6++
			var0[w] = lshr(var10[(0xFC0&var9)+(var14>>6)], var20) + ((0xfefeff & var0[var6]) >> 1)
			var14 += var17
			var9 += var18
			var14 = (786432 & var5) + (4095 & var14)
			var20 = var5 >> 20
			w = var6
			var6++
			var0[w] = lshr(var10[(0xFC0&var9)+(var14>>6)], var20) + ((var0[var6] & 0xFEFEFE) >> 1)
			var5 += var4
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = lshr(var10[(var9&0xFC0)+(var14>>6)], var20) + ((var0[var6] >> 1) & 0x7f7f7f)
			var9 += var18
			var14 += var17
			w = var6
			var6++
			var0[w] = lshr(var10[(0xFC0&var9)+(var14>>6)], var20) + ((var0[var6] & 0xfefeff) >> 1)
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = ((0xFEFEFE & var0[var6]) >> 1) + lshr(var10[(var14>>6)+(var9&0xFC0)], var20)
			var14 += var17
			var9 += var18
			var14 = (var5 & 786432) + (var14 & 4095)
			var20 = var5 >> 20
			w = var6
			var6++
			var0[w] = lshr(var10[(var9&0xFC0)+(var14>>6)], var20) + ((0xFEFEFE & var0[var6]) >> 1)
			var5 += var4
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = ((var0[var6] & 0xfefeff) >> 1) + lshr(var10[(var9&0xFC0)+(var14>>6)], var20)
			var14 += var17
			var9 += var18
			w = var6
			var6++
			var0[w] = lshr(var10[(0xFC0&var9)+(var14>>6)], var20) + ((var0[var6] >> 1) & 0x7f7f7f)
			var9 += var18
			var14 += var17
			w = var6
			var6++
			var0[w] = lshr(var10[(var14>>6)+(0xFC0&var9)], var20) + ((var0[var6] & 0xfefeff) >> 1)
		} else {
			for var21 := int32(0); var19 > var21; var21++ {
				w = var6
				var6++
				var0[w] = lshr(var10[(var14>>6)+(var9&0xFC0)], var20) + ((0xFEFEFE & var0[var6]) >> 1)
				var9 += var18
				var14 += var17
				if (var21 & 3) == 3 {
					var20 = var5 >> 20
					var14 = (var14 & 4095) + (786432 & var5)
					var5 += var4
				}
			}
		}
	}
	_ = var11
}
