package orsc

// texture.go — faithful port of the Scene texture-buffer machinery from
// OpenRSC's three/Scene.java: the public loadTexture entry point
// (Scene.loadTexture, Scene.java:2858), the lazy resource-database allocator
// (Scene.b(int,boolean), Scene.java:2356), and the shade-mip builder that turns
// a paletted texture into the four-tier shaded texel buffer the perspective
// span filler samples (Scene.setFrustum(int,byte), Scene.java:434).
//
// Java->Go gotchas honoured here:
//   - Java int is 32-bit; the texel buffers + LUTs are []int32.
//   - textureSource (Scene.m_g) is byte[][] in Java; signed. Every index read is
//     masked &0xff (Java `this.m_g[var1][..] & 255`, Scene.java:451).
//   - The 5:5:5 colour quantise is the caller's job (the palette m_L already
//     holds packed 0x00RRGGBB colours); setFrustum just masks 0xf8f8ff
//     (16316671) and builds the darker tiers.
//   - >>> (logical shift) on the int32 RGB value: Go's >> on int32 is
//     arithmetic, but these RGB values are always >=0 (top byte clear after the
//     0xf8f8ff mask), so >> and >>> coincide. We keep them as plain >>.

// textureSizeClass1 / textureSizeClass0 are the two on-disk texture edge sizes.
// 128px (class 1, Scene.m_Hb[id]!=0) needs a 65536-int buffer; 64px (class 0)
// needs 16384. The shade-mip quadruples that (4 brightness tiers), so the live
// buffer is 4× the base size. (Scene.java:439-442, 2372/2396.)
const (
	textureSizeClass1 = 128 // Scene.setFrustum var3=128 when m_Hb[id]!=0
	textureSizeClass0 = 64  // Scene.setFrustum var3=64  otherwise

	textureBufLen1 = 65536 // Scene.b: new int[65536] for class-1 (128² * 4)
	textureBufLen0 = 16384 // Scene.b: new int[16384] for class-0 (64²  * 4)
)

// texMask is the 5:5:5 colour mask 0xf8f8ff (16316671) applied to every texel
// (Scene.java:452 `var8 &= 16316671;` and the shade-mip masks 468-470). It zeroes
// the low 3 bits of R and G (keeping 5:5 there) but leaves the full blue byte —
// matching the classic engine's asymmetric quantise.
const texMask = 0xf8f8ff // 16316671

// texTransparentKey is the colour-keyed "transparent" texel after masking:
// 0xF800FF (16253183), i.e. the magenta transparency key (full red + full blue,
// zero green) preserved through the 0xF8F8FF mask. When a texel resolves to this
// after masking, the texture is flagged as having transparency (mS[id]=true) and
// the texel is forced to 0 (skip). Scene.java:454 `if (var8 == 16253183)`.
const texTransparentKey = 0xF800FF // 16253183

// textureClock mirrors the Java global MiscFunctions.world_s_e (Scene.java:2363):
// a monotonically increasing tick stamped into mD[id] on every texture touch,
// used by ensureTexture's LRU eviction. It is package-level (not per-Scene)
// exactly as the Java counter is static/global; mD values are only ever compared
// to each other, never to wall-clock, so a shared counter is faithful.
var textureClock int64

// LoadTexture registers texture id with its colour palette, index bytes, size
// class and (re)builds its resource-database entry. Faithful port of
// Scene.loadTexture (Scene.java:2858-2873):
//
//	this.m_g[var1] = var5;            // index bytes  (textureSource)
//	this.m_L[var1] = var3;            // palette      (texturePalette)
//	this.m_Hb[var1] = var4;           // size class   (mHb: 0=64px,1=128px)
//	this.m_D[var1] = 0L;              // LRU stamp    (mD)
//	this.m_S[var1] = false;           // has-alpha    (mS)
//	this.resourceDatabase[var1] = null;
//	this.b(var1, true);              // lazily allocate + fill
//
// palette is the per-texture colour LUT (256 packed 0x00RRGGBB ints). source is
// the per-texel palette-index byte grid (sizeClass²). sizeClass is 0 (64px) or 1
// (128px) — exactly Scene.m_Hb.
func (s *Scene) LoadTexture(id int, palette []int32, sizeClass int, source []int8) {
	s.textureSource[id] = source
	s.texturePalette[id] = palette
	s.mHb[id] = int32(sizeClass)
	s.mD[id] = 0
	s.mS[id] = false
	s.resourceDatabase[id] = nil
	s.ensureTexture(id, true)
}

// ensureTexture is Scene.b(int,boolean) (Scene.java:2356-2424): the lazy
// allocator + LRU evictor for the resource database. On a miss it grabs a free
// pool buffer of the right size class (pool128 for class 1, pool64 for class 0),
// or — if the pool is exhausted — steals the least-recently-used buffer of the
// same class from another texture (the var8/var5 LRU scan), then fills it via
// buildTextureBuffer.
//
// The `keep` argument mirrors Scene.b's var2: when false the Java code clears
// m_K (a pick-buffer-valid flag) — there is no pick buffer in this port, so the
// flag is dropped, but the parameter is preserved so call sites read identically
// to Scene.java (loadTexture passes true; the span filler passes true).
func (s *Scene) ensureTexture(id int, keep bool) {
	_ = keep // Scene.b var2: drives m_K (pick-buffer flag) in Java; no pick buffer here.
	if id < 0 {
		return
	}

	// Scene.java:2363 `this.m_D[var1] = (long)(MiscFunctions.world_s_e++);`
	// LRU timestamp; we use a monotonically increasing scene clock.
	textureClock++
	s.mD[id] = textureClock

	if s.resourceDatabase[id] != nil {
		return // already resident
	}

	if s.mHb[id] != 0 {
		// ---- class-1 (128px / 65536-int) path (Scene.java:2369-2392) ----
		for i := range s.pool128 {
			if s.pool128[i] == nil {
				s.pool128[i] = make([]int32, textureBufLen1)
				s.resourceDatabase[id] = s.pool128[i]
				s.buildTextureBuffer(id)
				return
			}
		}
		// Pool exhausted: steal the LRU class-1 buffer. (Scene.java:2379-2391)
		var oldest int64 = 1 << 30 // 1073741824L
		victim := 0
		for o := 0; o < s.textureCount; o++ {
			if o != id && s.mHb[o] == 1 && s.resourceDatabase[o] != nil && s.mD[o] < oldest {
				oldest = s.mD[o]
				victim = o
			}
		}
		s.resourceDatabase[id] = s.resourceDatabase[victim]
		s.resourceDatabase[victim] = nil
		s.buildTextureBuffer(id)
		return
	}

	// ---- class-0 (64px / 16384-int) path (Scene.java:2393-2416) ----
	for i := range s.pool64 {
		if s.pool64[i] == nil {
			s.pool64[i] = make([]int32, textureBufLen0)
			s.resourceDatabase[id] = s.pool64[i]
			s.buildTextureBuffer(id)
			return
		}
	}
	var oldest int64 = 1 << 30
	victim := 0
	for o := 0; o < s.textureCount; o++ {
		if o != id && s.mHb[o] == 0 && s.resourceDatabase[o] != nil && s.mD[o] < oldest {
			oldest = s.mD[o]
			victim = o
		}
	}
	s.resourceDatabase[id] = s.resourceDatabase[victim]
	s.resourceDatabase[victim] = nil
	s.buildTextureBuffer(id)
}

// buildTextureBuffer is Scene.setFrustum(int,byte) (Scene.java:434-476): it fills
// resourceDatabase[id] with the four shade-mip tiers.
//
// Stage 1 (Scene.java:449-464): for each texel, look up the palette colour via
// the signed index byte (&0xff), mask 0xf8f8ff, and:
//   - colour 0   -> store 1   (a near-black sentinel so the texel is non-zero,
//     because 0 means "transparent/skip" in the textured Shader paths);
//   - colour == texTransparentKey -> mark mS[id]=true and store 0 (real skip);
//   - else store the masked colour.
//
// The base tier is laid out row-major (var5 runs 0..size²-1).
//
// Stage 2 (Scene.java:466-470): for every base texel build three darker copies
// at offsets size², 2·size², 3·size²:
//
//	tier1 = (c - (c>>3)) & 0xf8f8ff   //  -1/8  (≈87.5% bright)
//	tier2 = (c - (c>>2)) & 0xf8f8ff   //  -1/4  (≈75%   bright)
//	tier3 = (c - (c>>3) - (c>>2)) & 0xf8f8ff   // -3/8 (≈62.5% bright)
//
// The Shader selects a tier per-quad from the high bits of the interpolated
// shade value (var6>>23 / var5>>20), so each block boundary multiplies the index
// by size² to jump tiers.
func (s *Scene) buildTextureBuffer(id int) {
	var size int
	if s.mHb[id] != 0 {
		size = textureSizeClass1 // 128
	} else {
		size = textureSizeClass0 // 64
	}

	buf := s.resourceDatabase[id]
	src := s.textureSource[id]
	pal := s.texturePalette[id]

	// Unpopulated texture id (no archive loaded for this slot): degrade to a flat
	// transparent texel bank instead of dereferencing the nil source. The NewScene
	// ctor documents this contract ("an unpopulated id (textureSource[id]==nil) is
	// treated as flat by the fill path so a missing texture degrades instead of
	// crashing"); this realizes it. mS[id]=true marks the bank transparent so the
	// textured Shader skips its texels (a missing-archive dump renders its
	// flat-coloured geometry; the textured terrain overlay is simply not sampled),
	// rather than panicking on src[...] (the roof_cull_* hunt fixtures hit this when
	// rendered through the bundle-less RenderDump path).
	if src == nil || pal == nil {
		for i := range buf {
			buf[i] = 0
		}
		s.mS[id] = true
		return
	}

	// ---- Stage 1: base tier (Scene.java:449-464) ----
	dst := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			// Java: this.m_L[var1][this.m_g[var1][var7 + var6*var3] & 255]
			idx := int(src[col+row*size]) & 0xff
			c := pal[idx] & texMask
			if c != 0 {
				if c == texTransparentKey {
					s.mS[id] = true
					c = 0
				}
			} else {
				c = 1
			}
			buf[dst] = c
			dst++
		}
	}

	// ---- Stage 2: three darker shade-mip tiers (Scene.java:466-470) ----
	// dst now equals size² (the tier stride, Java's var5).
	stride := dst
	for i := 0; i < stride; i++ {
		c := buf[i]
		buf[stride+i] = (c - (c >> 3)) & texMask
		buf[i+stride*2] = (c - (c >> 2)) & texMask
		buf[i+stride*3] = (c - (c >> 3) - (c >> 2)) & texMask
	}
}
