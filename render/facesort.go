package render

// This file ports the authentic RuneScape Classic scene face overlap-resolution
// pass from Scene.java (deob-204): method277 (the sliding-window driver),
// method278 (the recursive stable reinsertion), the pairwise overlap tests
// method295/method296, and the 2D polygon-overlap test method309 (+ helpers
// method306/307/308).
//
// The faithful coarse painter sort (qsort by average camera-Z, far-first) is
// applied in scene.go BEFORE this pass; resolveOverlaps then reorders
// mutually-overlapping faces by true 3D plane tests, fixing fragmented walls at
// L-corners and perpendicular junctions where a pure average-depth sort inverts.
//
// Port notes (DO NOT clean up the int math; the sign conventions are
// load-bearing):
//   - Polygon fields map onto collectedFace:
//       anInt353/355 -> minVX/maxVX   (screen view-X bbox)
//       anInt354/356 -> minVY/maxVY   (screen view-Y bbox)
//       anInt357/358 -> minCZ/maxCZ   (camera-Z extent)
//       anInt362/363/364 -> nx/ny/nz  (scaled camera-space normal)
//       anInt365     -> dotV0         (signed plane offset against vertex0)
//       faceCameraNormalMagnitude[j] -> normMag   (== 4*|n|, anInt402==4)
//       aBoolean367  -> visited
//       anInt368/369 -> anInt368/369  (reinsertion bookkeeping)
//   - method278's anInt454/anInt455 out-params live on faceSorter as parallel
//     state, carried EXACTLY as the client does (a naive "swap if behind"
//     mis-terminates / loops).
//   - All vertex/view/camera reads go through the model (m.viewX/viewY/camX/
//     camY/camZ) at test time, mirroring the client's per-Polygon model lookup.

// faceSorter owns the pointer slice being reordered plus method278's
// anInt454/anInt455 out-parameters (instance fields on Scene in the client).
// We reorder a []*collectedFace rather than the value slice so that swapping a
// slot mirrors the client swapping object REFERENCES in aclass7[] — the
// captured `polygon`/`polygon_1` references and their anInt368/anInt369 fields
// follow the element wherever method278 relocates it (the bookkeeping is
// object-identity based in the client, NOT index based).
type faceSorter struct {
	faces    []*collectedFace
	anInt454 int
	anInt455 int
}

// resolveOverlaps is the entry point: Scene.method277(100, aPolygonArray414,
// anInt413). i is the sliding-window span (100); the slice IS aclass7; j is the
// element count (anInt413). The client treats aclass7[j] as a valid (already
// projected, never-touched) terminator slot whose aBoolean367 stays false, so
// the `while (aclass7[l].aBoolean367) l++` walk halts at l==j and method277
// returns. We append one sentinel element to reproduce that terminator.
func resolveOverlaps(faces []collectedFace) {
	j := len(faces)
	if j < 2 {
		return
	}
	// pointer view + a sentinel terminator at index j (visited stays false).
	ptrs := make([]*collectedFace, j+1)
	for i := range faces {
		ptrs[i] = &faces[i]
	}
	ptrs[j] = &collectedFace{}
	s := &faceSorter{faces: ptrs}
	s.method277(100, j)
	// write the reordered elements back into the original slice.
	out := make([]collectedFace, j)
	for i := 0; i < j; i++ {
		out[i] = *ptrs[i]
	}
	copy(faces, out)
}

// method277 (Scene.java:203-234): the driver. Seeds the per-face bookkeeping,
// then walks a sliding window of width i, reinserting any closer face that
// overlaps and is in front of the current face. The l==j return is the
// termination: l reaches the terminator index (its visited flag never set).
func (s *faceSorter) method277(i, j int) {
	a := s.faces
	for k := 0; k <= j; k++ {
		a[k].visited = false
		a[k].anInt368 = int32(k)
		a[k].anInt369 = -1
	}

	l := 0
	for {
		for a[l].visited {
			l++
		}
		if l == j {
			return
		}
		polygon := a[l] // captured reference; follows the element if relocated
		polygon.visited = true
		i1 := l
		j1 := l + i
		if j1 >= j {
			j1 = j - 1
		}
		for k1 := j1; k1 >= i1+1; k1-- {
			polygon1 := a[k1]
			if polygon.minVX < polygon1.maxVX && polygon1.minVX < polygon.maxVX &&
				polygon.minVY < polygon1.maxVY && polygon1.minVY < polygon.maxVY &&
				polygon.anInt368 != polygon1.anInt369 &&
				!s.method295(polygon, polygon1) && s.method296(polygon1, polygon) {
				s.method278(i1, k1)
				if a[k1] != polygon1 {
					k1++
				}
				i1 = s.anInt454
				polygon1.anInt369 = polygon.anInt368
			}
		}
	}
}

// method278 (Scene.java:236-279): recursive stable reinsertion. Pushes element i
// forward (and element j backward) past every element it is NOT separated from
// (method295 == false), narrowing the [i,j] window until it collapses, recursing
// on the inner window. anInt454/anInt455 carry the resulting window bounds back
// to the caller EXACTLY as the client's instance fields do. Returns whether a
// terminal swap reached i==j on this level.
func (s *faceSorter) method278(i, j int) bool {
	a := s.faces
	for {
		polygon := a[i]
		// Bubble polygon forward past following elements it does not separate from.
		for k := i + 1; k <= j; k++ {
			polygon1 := a[k]
			if !s.method295(polygon1, polygon) {
				break
			}
			a[i] = polygon1
			a[k] = polygon
			i = k
			if i == j {
				s.anInt454 = i
				s.anInt455 = i - 1
				return true
			}
		}

		polygon2 := a[j]
		// Bubble polygon2 backward past preceding elements it does not separate from.
		for l := j - 1; l >= i; l-- {
			polygon3 := a[l]
			if !s.method295(polygon2, polygon3) {
				break
			}
			a[j] = polygon3
			a[l] = polygon2
			j = l
			if i == j {
				s.anInt454 = j + 1
				s.anInt455 = j
				return true
			}
		}

		if i+1 >= j {
			s.anInt454 = i
			s.anInt455 = j
			return false
		}
		if !s.method278(i+1, j) {
			s.anInt454 = i
			return false
		}
		j = s.anInt455
	}
}

// method295 (Scene.java:2446-2554): the pairwise overlap test. Returns true when
// the two faces are PROVABLY separated (do NOT overlap / no reorder needed) and
// false when they may overlap. Order: screen-AABB early-out, camZ-range
// early-out, then the two separating-plane vertex-dot loops, then the method309
// 2D-overlap fallback (mandatory).
func (s *faceSorter) method295(polygon, polygon1 *collectedFace) bool {
	if polygon.minVX >= polygon1.maxVX {
		return true
	}
	if polygon1.minVX >= polygon.maxVX {
		return true
	}
	if polygon.minVY >= polygon1.maxVY {
		return true
	}
	if polygon1.minVY >= polygon.maxVY {
		return true
	}
	if polygon.minCZ >= polygon1.maxCZ {
		return true
	}
	if polygon1.minCZ > polygon.maxCZ {
		return false
	}

	gameModel := polygon.model
	gameModel1 := polygon1.model
	i := polygon.face
	j := polygon1.face
	ai := gameModel.FaceVertices[i]
	ai1 := gameModel1.FaceVertices[j]
	k := gameModel.FaceNumVertices[i]
	l := gameModel1.FaceNumVertices[j]

	// Is every vertex of `polygon` on one side of `polygon1`'s plane?
	k2 := gameModel1.camX[ai1[0]]
	l2 := gameModel1.camY[ai1[0]]
	i3 := gameModel1.camZ[ai1[0]]
	j3 := polygon1.nx
	k3 := polygon1.ny
	l3 := polygon1.nz
	i4 := polygon1.normMag
	j4 := polygon1.dotV0
	flag := false
	for k4 := 0; k4 < k; k4++ {
		i1 := ai[k4]
		i2 := (k2-gameModel.camX[i1])*j3 + (l2-gameModel.camY[i1])*k3 + (i3-gameModel.camZ[i1])*l3
		if (i2 >= -i4 || j4 >= 0) && (i2 <= i4 || j4 <= 0) {
			continue
		}
		flag = true
		break
	}
	if !flag {
		return true
	}

	// Is every vertex of `polygon1` on one side of `polygon`'s plane?
	k2 = gameModel.camX[ai[0]]
	l2 = gameModel.camY[ai[0]]
	i3 = gameModel.camZ[ai[0]]
	j3 = polygon.nx
	k3 = polygon.ny
	l3 = polygon.nz
	i4 = polygon.normMag
	j4 = polygon.dotV0
	flag = false
	for l4 := 0; l4 < l; l4++ {
		j1 := ai1[l4]
		j2 := (k2-gameModel1.camX[j1])*j3 + (l2-gameModel1.camY[j1])*k3 + (i3-gameModel1.camZ[j1])*l3
		if (j2 >= -i4 || j4 <= 0) && (j2 <= i4 || j4 >= 0) {
			continue
		}
		flag = true
		break
	}
	if !flag {
		return true
	}

	// 2D screen-polygon overlap fallback (method309). The 2-vertex "line" faces
	// (RSC walls projected edge-on) are widened into a ±20px quad first, exactly
	// like the client (Scene.java:2509-2552).
	var ai2, ai3 []int32
	if k == 2 {
		ai2 = make([]int32, 4)
		ai3 = make([]int32, 4)
		i5 := ai[0]
		k1 := ai[1]
		ai2[0] = gameModel.viewX[i5] - 20
		ai2[1] = gameModel.viewX[k1] - 20
		ai2[2] = gameModel.viewX[k1] + 20
		ai2[3] = gameModel.viewX[i5] + 20
		ai3[0] = gameModel.viewY[i5]
		ai3[3] = gameModel.viewY[i5]
		ai3[1] = gameModel.viewY[k1]
		ai3[2] = gameModel.viewY[k1]
	} else {
		ai2 = make([]int32, k)
		ai3 = make([]int32, k)
		for j5 := 0; j5 < k; j5++ {
			i6 := ai[j5]
			ai2[j5] = gameModel.viewX[i6]
			ai3[j5] = gameModel.viewY[i6]
		}
	}
	var ai4, ai5 []int32
	if l == 2 {
		ai4 = make([]int32, 4)
		ai5 = make([]int32, 4)
		k5 := ai1[0]
		l1 := ai1[1]
		ai4[0] = gameModel1.viewX[k5] - 20
		ai4[1] = gameModel1.viewX[l1] - 20
		ai4[2] = gameModel1.viewX[l1] + 20
		ai4[3] = gameModel1.viewX[k5] + 20
		ai5[0] = gameModel1.viewY[k5]
		ai5[3] = gameModel1.viewY[k5]
		ai5[1] = gameModel1.viewY[l1]
		ai5[2] = gameModel1.viewY[l1]
	} else {
		ai4 = make([]int32, l)
		ai5 = make([]int32, l)
		for l5 := 0; l5 < l; l5++ {
			j6 := ai1[l5]
			ai4[l5] = gameModel1.viewX[j6]
			ai5[l5] = gameModel1.viewY[j6]
		}
	}
	return !method309(ai2, ai3, ai4, ai5)
}

// method296 (Scene.java:2556-2604): the one-way plane test, NO 2D fallback. Used
// by method277 to decide direction (is `polygon` in front of `polygon1`?).
func (s *faceSorter) method296(polygon, polygon1 *collectedFace) bool {
	gameModel := polygon.model
	gameModel1 := polygon1.model
	i := polygon.face
	j := polygon1.face
	ai := gameModel.FaceVertices[i]
	ai1 := gameModel1.FaceVertices[j]
	k := gameModel.FaceNumVertices[i]
	l := gameModel1.FaceNumVertices[j]

	i2 := gameModel1.camX[ai1[0]]
	j2 := gameModel1.camY[ai1[0]]
	k2 := gameModel1.camZ[ai1[0]]
	l2 := polygon1.nx
	i3 := polygon1.ny
	j3 := polygon1.nz
	k3 := polygon1.normMag
	l3 := polygon1.dotV0
	flag := false
	for i4 := 0; i4 < k; i4++ {
		i1 := ai[i4]
		k1 := (i2-gameModel.camX[i1])*l2 + (j2-gameModel.camY[i1])*i3 + (k2-gameModel.camZ[i1])*j3
		if (k1 >= -k3 || l3 >= 0) && (k1 <= k3 || l3 <= 0) {
			continue
		}
		flag = true
		break
	}
	if !flag {
		return true
	}

	i2 = gameModel.camX[ai[0]]
	j2 = gameModel.camY[ai[0]]
	k2 = gameModel.camZ[ai[0]]
	l2 = polygon.nx
	i3 = polygon.ny
	j3 = polygon.nz
	k3 = polygon.normMag
	l3 = polygon.dotV0
	flag = false
	for j4 := 0; j4 < l; j4++ {
		j1 := ai1[j4]
		l1 := (i2-gameModel1.camX[j1])*l2 + (j2-gameModel1.camY[j1])*i3 + (k2-gameModel1.camZ[j1])*j3
		if (l1 >= -k3 || l3 <= 0) && (l1 <= k3 || l3 >= 0) {
			continue
		}
		flag = true
		break
	}
	return !flag
}

// method306 (Scene.java:2772): linear interpolation helper. Given the segment
// (i,j)->(k,l) returns the x at scanline-coordinate i1.
func method306(i, j, k, l, i1 int32) int32 {
	if l == j {
		return i
	}
	return i + ((k-i)*(i1-j))/(l-j)
}

// method307 (Scene.java:2779): the 4-value side test used inside method309.
func method307(i, j, k, l int32, flag bool) bool {
	if flag && i <= k || i < k {
		if i > l {
			return true
		}
		if j > k {
			return true
		}
		if j > l {
			return true
		}
		return !flag
	}
	if i < l {
		return true
	}
	if j < k {
		return true
	}
	if j < l {
		return true
	}
	return flag
}

// method308 (Scene.java:2799): the 3-value side test used inside method309.
func method308(i, j, k int32, flag bool) bool {
	if flag && i <= k || i < k {
		if j > k {
			return true
		}
		return !flag
	}
	if j < k {
		return true
	}
	return flag
}

// method309 (Scene.java:2811-3043): the 2D convex-polygon overlap test. ai/ai1
// are polygon A's (x,y); ai2/ai3 are polygon B's (x,y). Returns true if the two
// screen polygons overlap. Ported in FULL, including the byte0 state machine.
func method309(ai, ai1, ai2, ai3 []int32) bool {
	i := len(ai)
	j := len(ai2)
	var byte0 int32 // 0

	i20 := ai1[0]
	k20 := i20
	k := 0
	j20 := ai3[0]
	l20 := j20
	i1 := 0
	for i21 := 1; i21 < i; i21++ {
		if ai1[i21] < i20 {
			i20 = ai1[i21]
			k = i21
		} else if ai1[i21] > k20 {
			k20 = ai1[i21]
		}
	}
	for j21 := 1; j21 < j; j21++ {
		if ai3[j21] < j20 {
			j20 = ai3[j21]
			i1 = j21
		} else if ai3[j21] > l20 {
			l20 = ai3[j21]
		}
	}

	if j20 >= k20 {
		return false
	}
	if i20 >= l20 {
		return false
	}
	var l, j1 int
	var flag bool
	if ai1[k] < ai3[i1] {
		for l = k; ai1[l] < ai3[i1]; l = (l + 1) % i {
		}
		for ; ai1[k] < ai3[i1]; k = ((k - 1) + i) % i {
		}
		k1 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
		k6 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
		l10 := ai2[i1]
		flag = (k1 < l10) || (k6 < l10)
		if method308(k1, k6, l10, flag) {
			return true
		}
		j1 = (i1 + 1) % j
		i1 = ((i1 - 1) + j) % j
		if k == l {
			byte0 = 1
		}
	} else {
		for j1 = i1; ai3[j1] < ai1[k]; j1 = (j1 + 1) % j {
		}
		for ; ai3[i1] < ai1[k]; i1 = ((i1 - 1) + j) % j {
		}
		l1 := ai[k]
		i11 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[k])
		l15 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[k])
		flag = (l1 < i11) || (l1 < l15)
		if method308(i11, l15, l1, !flag) {
			return true
		}
		l = (k + 1) % i
		k = ((k - 1) + i) % i
		if i1 == j1 {
			byte0 = 2
		}
	}

	for byte0 == 0 {
		if ai1[k] < ai1[l] {
			if ai1[k] < ai3[i1] {
				if ai1[k] < ai3[j1] {
					i2 := ai[k]
					l6 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai1[k])
					j11 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[k])
					i16 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[k])
					if method307(i2, l6, j11, i16, flag) {
						return true
					}
					k = ((k - 1) + i) % i
					if k == l {
						byte0 = 1
					}
				} else {
					j2 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
					i7 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
					k11 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
					j16 := ai2[j1]
					if method307(j2, i7, k11, j16, flag) {
						return true
					}
					j1 = (j1 + 1) % j
					if i1 == j1 {
						byte0 = 2
					}
				}
			} else if ai3[i1] < ai3[j1] {
				k2 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
				j7 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
				l11 := ai2[i1]
				k16 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai3[i1])
				if method307(k2, j7, l11, k16, flag) {
					return true
				}
				i1 = ((i1 - 1) + j) % j
				if i1 == j1 {
					byte0 = 2
				}
			} else {
				l2 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
				k7 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
				i12 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
				l16 := ai2[j1]
				if method307(l2, k7, i12, l16, flag) {
					return true
				}
				j1 = (j1 + 1) % j
				if i1 == j1 {
					byte0 = 2
				}
			}
		} else if ai1[l] < ai3[i1] {
			if ai1[l] < ai3[j1] {
				i3 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai1[l])
				l7 := ai[l]
				j12 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[l])
				i17 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[l])
				if method307(i3, l7, j12, i17, flag) {
					return true
				}
				l = (l + 1) % i
				if k == l {
					byte0 = 1
				}
			} else {
				j3 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
				i8 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
				k12 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
				j17 := ai2[j1]
				if method307(j3, i8, k12, j17, flag) {
					return true
				}
				j1 = (j1 + 1) % j
				if i1 == j1 {
					byte0 = 2
				}
			}
		} else if ai3[i1] < ai3[j1] {
			k3 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
			j8 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
			l12 := ai2[i1]
			k17 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai3[i1])
			if method307(k3, j8, l12, k17, flag) {
				return true
			}
			i1 = ((i1 - 1) + j) % j
			if i1 == j1 {
				byte0 = 2
			}
		} else {
			l3 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
			k8 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
			i13 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
			l17 := ai2[j1]
			if method307(l3, k8, i13, l17, flag) {
				return true
			}
			j1 = (j1 + 1) % j
			if i1 == j1 {
				byte0 = 2
			}
		}
	}

	for byte0 == 1 {
		if ai1[k] < ai3[i1] {
			if ai1[k] < ai3[j1] {
				i4 := ai[k]
				j13 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[k])
				i18 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[k])
				return method308(j13, i18, i4, !flag)
			}
			j4 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
			l8 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
			k13 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
			j18 := ai2[j1]
			if method307(j4, l8, k13, j18, flag) {
				return true
			}
			j1 = (j1 + 1) % j
			if i1 == j1 {
				byte0 = 0
			}
		} else if ai3[i1] < ai3[j1] {
			k4 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
			i9 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
			l13 := ai2[i1]
			k18 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai3[i1])
			if method307(k4, i9, l13, k18, flag) {
				return true
			}
			i1 = ((i1 - 1) + j) % j
			if i1 == j1 {
				byte0 = 0
			}
		} else {
			l4 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[j1])
			j9 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[j1])
			i14 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai3[j1])
			l18 := ai2[j1]
			if method307(l4, j9, i14, l18, flag) {
				return true
			}
			j1 = (j1 + 1) % j
			if i1 == j1 {
				byte0 = 0
			}
		}
	}

	for byte0 == 2 {
		if ai3[i1] < ai1[k] {
			if ai3[i1] < ai1[l] {
				i5 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
				k9 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
				j14 := ai2[i1]
				return method308(i5, k9, j14, flag)
			}
			j5 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai1[l])
			l9 := ai[l]
			k14 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[l])
			i19 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[l])
			if method307(j5, l9, k14, i19, flag) {
				return true
			}
			l = (l + 1) % i
			if k == l {
				byte0 = 0
			}
		} else if ai1[k] < ai1[l] {
			k5 := ai[k]
			i10 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai1[k])
			l14 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[k])
			j19 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[k])
			if method307(k5, i10, l14, j19, flag) {
				return true
			}
			k = ((k - 1) + i) % i
			if k == l {
				byte0 = 0
			}
		} else {
			l5 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai1[l])
			j10 := ai[l]
			i15 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[l])
			k19 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[l])
			if method307(l5, j10, i15, k19, flag) {
				return true
			}
			l = (l + 1) % i
			if k == l {
				byte0 = 0
			}
		}
	}

	if ai1[k] < ai3[i1] {
		i6 := ai[k]
		j15 := method306(ai2[(i1+1)%j], ai3[(i1+1)%j], ai2[i1], ai3[i1], ai1[k])
		l19 := method306(ai2[((j1-1)+j)%j], ai3[((j1-1)+j)%j], ai2[j1], ai3[j1], ai1[k])
		return method308(j15, l19, i6, !flag)
	}
	j6 := method306(ai[(k+1)%i], ai1[(k+1)%i], ai[k], ai1[k], ai3[i1])
	k10 := method306(ai[((l-1)+i)%i], ai1[((l-1)+i)%i], ai[l], ai1[l], ai3[i1])
	k15 := ai2[i1]
	return method308(j6, k10, k15, flag)
}
