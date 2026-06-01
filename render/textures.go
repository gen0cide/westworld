package render

import (
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// textureName is the authentic RSC texture id -> base sprite name ordering,
// ported verbatim from OpenRSC EntityHandler.loadTextureDefinitions() (the
// TextureDef dataName, the per-texture base bitmap; the animationName subtype
// is an overlay we ignore for dominant-colour sampling). Index 1 = water,
// 17 = fountain, 25 = gungywater, 31 = lava, etc. The classic textures17.jag
// stores each as "<name>.dat" (Surface.loadSprite). This is the 55-entry
// authentic set (custom textures 55+ are not loaded by the authentic client).
var textureName = []string{
	"wall",         // 0  (door)
	"water",        // 1
	"wall",         // 2
	"planks",       // 3
	"wall",         // 4  (doorway)
	"wall",         // 5  (window)
	"roof",         // 6
	"wall",         // 7  (arrowslit)
	"leafytree",    // 8
	"treestump",    // 9
	"fence",        // 10
	"mossy",        // 11
	"railings",     // 12
	"painting1",    // 13
	"painting2",    // 14
	"marble",       // 15
	"deadtree",     // 16
	"fountain",     // 17
	"wall",         // 18 (stainedglass)
	"target",       // 19
	"books",        // 20
	"timbered",     // 21
	"timbered",     // 22 (timberwindow)
	"mossybricks",  // 23
	"growingwheat", // 24
	"gungywater",   // 25
	"web",          // 26
	"wall",         // 27 (desertwindow)
	"wall",         // 28 (crumbled)
	"cavern",       // 29
	"cavern2",      // 30
	"lava",         // 31
	"pentagram",    // 32
	"mapletree",    // 33
	"yewtree",      // 34
	"helmet",       // 35
	"canvas",       // 36 (tentbottom)
	"Chainmail2",   // 37
	"mummy",        // 38
	"jungleleaf",   // 39
	"jungleleaf3",  // 40
	"jungleleaf4",  // 41
	"jungleleaf5",  // 42
	"jungleleaf6",  // 43
	"mossybricks",  // 44 (arrowslit)
	"planks",       // 45 (window)
	"planks",       // 46 (junglewindow)
	"cargonet",     // 47
	"bark",         // 48
	"canvas",       // 49
	"canvas",       // 50 (tentdoor)
	"wall",         // 51 (lowcrumbled)
	"cavern",       // 52 (crumbled)
	"cavern2",      // 53 (crumbled)
	"lava",         // 54 (flames)
}

var (
	textureColourOnce sync.Once
	// textureColour[id] is the method305-encoded flat fill sampled from the
	// authentic texture bitmap. A zero entry means "not sampled" (decode failed
	// or no archive) and textureFill falls back to the hardcoded approximation.
	textureColour map[int32]int32
)

// textureBuf is the per-texture-id shaded texel buffer the perspective span
// filler (method282/283/285) reads from — the Go analogue of the client's
// anIntArrayArray429[id] built by Scene.method300. texels holds FOUR concatenated
// shade-mip levels (base, then -1/8, -1/4, -1/4-1/8 darker), each size*size RGB
// ints in row-major (x + y*size) order. size is 64 (class0) or 128 (class1).
// hasAlpha is set when any texel decoded to the magenta transparency key (so the
// face uses the transparent-skip filler method285).
type textureBuf struct {
	texels   []int32
	size     int32
	hasAlpha bool
}

var (
	textureBufOnce sync.Once
	// textureBufs[id] is the shaded texel buffer for texture id, or nil/absent
	// when the archive is missing or the entry failed to decode (the caller then
	// falls back to the flat sampled colour — never worse than today).
	textureBufs map[int32]*textureBuf
)

// loadTextureColours records a flat method305 fill per texture id — the dominant
// non-transparent colour of each texture's base texel buffer (built from
// Authentic_Sprites.orsc). textureFill uses it for a face that references a
// texture id but isn't drawn through the full texel buffer. Never panics;
// memoised via textureColourOnce.
func loadTextureColours() {
	textureColour = map[int32]int32{}
	textureBufOnce.Do(loadTextureBuffers)
	for id, buf := range textureBufs {
		if buf == nil {
			continue
		}
		n := int(buf.size * buf.size)
		if n <= 0 || n > len(buf.texels) {
			continue
		}
		counts := map[int32]int{}
		var best int32 = -1
		bestN := 0
		for p := 0; p < n; p++ {
			c := buf.texels[p] & 0xf8f8ff
			if c == 0 {
				continue // transparency key
			}
			if counts[c]++; counts[c] > bestN {
				bestN, best = counts[c], c
			}
		}
		if best < 0 {
			continue
		}
		textureColour[id] = method305(best>>16&0xff, best>>8&0xff, best&0xff)
	}
}

// loadTextureBuffers builds the shaded texel buffer for every texture id from
// OpenRSC's Authentic_Sprites.orsc (sprite spriteTexture+i), via
// loadTextureBuffersORSC. Memoised via textureBufOnce; never panics.
func loadTextureBuffers() {
	textureBufs = map[int32]*textureBuf{}
	if sa := sprites(); sa != nil {
		loadTextureBuffersORSC(sa)
	}
}

// shadeMip turns a resolved RGB texel canvas into the 4-level shaded texel buffer
// the rasteriser samples (Scene.b/d, three/Scene.java:2519-2523): level0 = base
// masked 0xf8f8ff, levels 1-3 = base-(base>>3), base-(base>>2),
// base-(base>>2)-(base>>3) — matching OpenRSC's mask (16316671 == 0xf8f8ff) and
// >>> shifts exactly. A texel masking to 0 (genuine black) is bumped to 1 so it
// stays opaque; the masked magenta key 0xf800ff becomes 0 and sets hasAlpha.
//
// greenKey selects the TRANSPARENCY KEY: the classic .jag textures key on GREEN
// (0x00ff00 -> magenta), so greenKey=true converts it. The OpenRSC .orsc textures
// key on BLACK (already mapped to magenta by the quantiser, mudclient
// loadTexturesAuthentic:14664), so greenKey=false leaves green as an opaque
// colour — green is legitimate texture colour there, not a hole.
func shadeMip(rgb []int, n, size int32, greenKey bool) *textureBuf {
	texels := make([]int32, n*4)
	hasAlpha := false
	for p := int32(0); p < n; p++ {
		raw := rgb[p]
		if greenKey && raw == 0x00ff00 {
			raw = 0xff00ff
		}
		c := raw & 0xf8f8ff
		if c == 0 {
			c = 1
		} else if c == 0xf800ff {
			c = 0
			hasAlpha = true
		}
		texels[p] = int32(c)
	}
	// 3 darker shade-mip levels. uint32 for the >>> (logical) shifts, exactly like
	// the Java >>> on positive RGB values.
	for p := int32(0); p < n; p++ {
		k := uint32(texels[p])
		texels[n+p] = int32((k - (k >> 3)) & 0xf8f8ff)
		texels[n*2+p] = int32((k - (k >> 2)) & 0xf8f8ff)
		texels[n*3+p] = int32((k - (k >> 2) - (k >> 3)) & 0xf8f8ff)
	}
	return &textureBuf{texels: texels, size: size, hasAlpha: hasAlpha}
}

// loadTextureBuffersORSC builds textureBufs from OpenRSC's Authentic_Sprites.orsc:
// each texture i is sprite spriteTexture+i, quantised + shade-mipped exactly as
// mudclient.loadTexturesAuthentic (14650) feeding Scene.loadTexture. Returns true
// if at least one texture loaded. textureName supplies only the COUNT (the .orsc
// is keyed by numeric id, not name — no per-name .dat, no subtype compositing:
// OpenRSC bakes the final texture into one sprite). Never panics.
func loadTextureBuffersORSC(sa *assets.SpriteArchive) bool {
	loaded := false
	for i := range textureName {
		sp, err := sa.Sprite(spriteTexture + i)
		if err != nil || sp == nil {
			continue
		}
		size := int32(sp.Width)
		if sp.Height != sp.Width || (size != 64 && size != 128) {
			continue
		}
		n := size * size
		if int(n) != len(sp.Pixels) {
			continue
		}
		rgb := quantizeTextureSprite(sp.Pixels)
		// OpenRSC textures key on BLACK (already mapped to the magenta key by the
		// quantiser), NOT green — so greenKey=false.
		textureBufs[int32(i)] = shadeMip(rgb, n, size, false)
		loaded = true
	}
	return loaded
}

// quantizeTextureSprite ports mudclient.loadTexturesAuthentic's texture quantiser
// (mudclient.java:14658-14715): build a 5:5:5-bucket histogram, map black->magenta
// transparency key, pick the 256 most-frequent colours into a dictionary
// (dict[0]=magenta) with the +0x40404 5:5:5 reconstruction bump, and re-index
// every pixel to its nearest dictionary colour. Returns the resolved RGB canvas
// (dictionary[index] per pixel) — the EXACT colours OpenRSC's Scene samples, so we
// mirror the client's 256-colour texture reduction rather than using raw ARGB.
func quantizeTextureSprite(px []uint32) []int {
	n := len(px)
	bucket := func(p int) int {
		return ((p & 0xf80000) >> 9) + ((p & 0xf800) >> 6) + ((p & 0xf8) >> 3)
	}
	hist := make([]int, 32768)
	for _, p := range px {
		hist[bucket(int(p)&0xffffff)]++
	}
	// black -> magenta key (AFTER the histogram, matching the client's order)
	pix := make([]int, n)
	for i, p := range px {
		c := int(p) & 0xffffff
		if c == 0x000000 {
			c = 0xff00ff
		}
		pix[i] = c
	}
	dict := make([]int, 256)
	dict[0] = 0xff00ff
	counts := make([]int, 256)
	for i1 := 0; i1 < len(hist); i1++ {
		if j1 := hist[i1]; j1 > counts[255] {
			for k1 := 1; k1 < 256; k1++ {
				if j1 <= counts[k1] {
					continue
				}
				for i2 := 255; i2 > k1; i2-- {
					dict[i2] = dict[i2-1]
					counts[i2] = counts[i2-1]
				}
				dict[k1] = ((i1 & 0x7c00) << 9) + ((i1 & 0x3e0) << 6) + ((i1 & 0x1f) << 3) + 0x40404
				counts[k1] = j1
				break
			}
		}
		hist[i1] = -1 // repurpose hist as a bucket->dict-index cache (-1 = unmapped)
	}
	rgb := make([]int, n)
	for l1 := 0; l1 < n; l1++ {
		j2 := pix[l1]
		k2 := bucket(j2)
		l2 := hist[k2]
		if l2 == -1 {
			best := 0x3b9ac9ff
			r, g, b := (j2>>16)&0xff, (j2>>8)&0xff, j2&0xff
			for i4 := 0; i4 < 256; i4++ {
				d := dict[i4]
				dr, dg, db := r-((d>>16)&0xff), g-((d>>8)&0xff), b-(d&0xff)
				if dist := dr*dr + dg*dg + db*db; dist < best {
					best, l2 = dist, i4
				}
			}
			hist[k2] = l2
		}
		rgb[l1] = dict[l2]
	}
	return rgb
}

// textureBuffer returns the shaded texel buffer for texture id, or nil when the
// archive/decode was unavailable (memoised). A nil result tells the rasteriser
// to use the existing flat-colour fallback path for that face.
func textureBuffer(id int32) *textureBuf {
	textureBufOnce.Do(loadTextureBuffers)
	return textureBufs[id]
}
