package render

import (
	"os"
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

// textureSubtype is the per-texture SUBTYPE overlay sprite name, ported verbatim
// from the second TextureDef arg (OpenRSC EntityHandler.loadTextureDefinitions
// :252-306, corroborated by deob GameData.textureSubtypeName). The authentic
// client composites this sprite OVER the base bitmap before building the texel
// buffer: deob mudclient.java:4524-4544 drawBox(magenta)->drawSprite(base)->
// (if subtype) loadSprite(subtype)+drawSprite(subtype). Surface.drawSprite
// (Surface.java:937-939) copies a subtype texel ONLY where its palette INDEX != 0,
// so transparent-index texels leave the base showing — this is what paints the
// arrowslit / stained-glass / window PATTERN onto a plain stone/timber wall. An
// empty string means "no overlay" (base sprite only, unchanged behaviour). Must
// stay index-aligned with textureName (55 entries).
var textureSubtype = []string{
	"door",         // 0
	"",             // 1
	"",             // 2
	"",             // 3
	"doorway",      // 4
	"window",       // 5
	"",             // 6
	"arrowslit",    // 7
	"",             // 8
	"",             // 9
	"",             // 10
	"",             // 11
	"",             // 12
	"",             // 13
	"",             // 14
	"",             // 15
	"",             // 16
	"",             // 17
	"stainedglass", // 18
	"",             // 19
	"",             // 20
	"",             // 21
	"timberwindow", // 22
	"",             // 23
	"",             // 24
	"",             // 25
	"",             // 26
	"desertwindow", // 27
	"",             // 28
	"",             // 29
	"",             // 30
	"",             // 31
	"",             // 32
	"",             // 33
	"",             // 34
	"",             // 35
	"tentbottom",   // 36
	"",             // 37
	"",             // 38
	"",             // 39
	"",             // 40
	"",             // 41
	"",             // 42
	"",             // 43
	"arrowslit",    // 44
	"window",       // 45
	"junglewindow", // 46
	"",             // 47
	"",             // 48
	"",             // 49
	"tentdoor",     // 50
	"lowcrumbled",  // 51
	"crumbled",     // 52
	"crumbled",     // 53
	"flames",       // 54
}

// textureJagSearch is the ordered list of candidate filesystem locations for
// the authentic classic textures17.jag (Version.TEXTURES = 17). OpenRSC repacks
// textures into its .orsc sprite archives under opaque sequential names, so the
// only per-name "<texture>.dat" source is the classic JAG shipped with the deob
// reference clients. The WESTWORLD_TEXTURES_JAG env var overrides everything.
var textureJagSearch = []string{
	"/Users/flint/Code/rscdump.com-runescape-classic-dump/eggsampler-rsc-204-d223fc6b77db/eggsampler-rsc-204-d223fc6b77db/data/textures17.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/textures17.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/video/textures17.jag",
	"/Users/flint/Code/openrsc/textures17.jag",
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

// loadTextureColours opens textures17.jag (first candidate that exists, or the
// WESTWORLD_TEXTURES_JAG override) and, for every texture id, decodes its sprite
// texels and records the dominant non-transparent colour as a method305 fill.
// It NEVER panics and NEVER fails the renderer: any error leaves the entry unset
// so textureFill uses its fallback. Result is memoised via textureColourOnce.
func loadTextureColours() {
	textureColour = map[int32]int32{}

	var arc *assets.Archive
	candidates := textureJagSearch
	if p := os.Getenv("WESTWORLD_TEXTURES_JAG"); p != "" {
		candidates = append([]string{p}, candidates...)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		a, err := assets.OpenArchive(p)
		if err != nil {
			continue
		}
		arc = a
		break
	}
	if arc == nil {
		return // no archive -> all ids fall back to approximation
	}

	indexData, err := arc.Get("index.dat")
	if err != nil || len(indexData) < 5 {
		return
	}

	for id, name := range textureName {
		spriteData, err := arc.Get(name + ".dat")
		if err != nil {
			continue
		}
		r, g, b, ok := dominantTexel(spriteData, indexData)
		if !ok {
			continue
		}
		textureColour[int32(id)] = method305(int32(r), int32(g), int32(b))
	}
}

// decodedSprite is one fully-decoded texture sprite (frame 0): the RGB palette
// (entry 0 = 0xff00ff magenta transparency key), the per-texel palette INDEX
// bytes laid out row-major (x + y*fullW) within the FULL fullW x fullH canvas
// (the frame is placed at its translate offset, the rest left as index 0 =
// transparent — replicating drawBox(magenta)+drawSprite), and the full size.
type decodedSprite struct {
	palette []int  // RGB per palette index; [0] is the magenta key
	idx     []byte // fullW*fullH palette indices, row-major
	fullW   int
	fullH   int
}

// decodeTextureSprite parses one texture sprite exactly as Surface.loadSprite +
// the drawBox/drawSprite compositing the client does before method300: it reads
// the palette and frame header from index.dat, then lays the frame's texels into
// a fullW x fullH index canvas (respecting the column-major flag==1 storage
// order, the known trap) at the frame's translate offset, with the uncovered
// border left transparent. Returns ok=false on any malformed/short entry; never
// panics.
func decodeTextureSprite(spriteData, indexData []byte) (ds decodedSprite, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	if len(spriteData) < 2 {
		return ds, false
	}
	// The index offset into index.dat lives in the first two bytes of the sprite
	// payload (Surface.loadSprite: indexOff = getUnsignedShort(spriteData, 0)).
	io := texU16(spriteData, 0)
	if io+5 > len(indexData) {
		return ds, false
	}
	fullW := texU16(indexData, io)
	io += 2
	fullH := texU16(indexData, io)
	io += 2
	colourCount := int(indexData[io] & 0xff)
	io++
	if colourCount < 1 || io+3*(colourCount-1) > len(indexData) {
		return ds, false
	}
	colours := make([]int, colourCount)
	colours[0] = 0xff00ff // transparency key
	for i := 0; i < colourCount-1; i++ {
		colours[i+1] = (int(indexData[io]&0xff) << 16) |
			(int(indexData[io+1]&0xff) << 8) |
			int(indexData[io+2]&0xff)
		io += 3
	}

	// frame 0 header: translateX, translateY (u8), width, height (u16), flag (u8)
	if io+7 > len(indexData) {
		return ds, false
	}
	tx := int(indexData[io] & 0xff)
	ty := int(indexData[io+1] & 0xff)
	io += 2
	w := texU16(indexData, io)
	io += 2
	h := texU16(indexData, io)
	io += 2
	flag := int(indexData[io] & 0xff)
	frameSize := w * h
	if fullW <= 0 || fullH <= 0 || frameSize <= 0 {
		return ds, false
	}

	spriteOff := 2
	if spriteOff+frameSize > len(spriteData) {
		return ds, false
	}

	// Full-canvas index buffer, all transparent (index 0) by default — this is
	// the drawBox(0,0,W,H,magenta) the client paints first. Then the frame is
	// composited at (tx,ty) like drawSprite(0,0,id).
	canvas := make([]byte, fullW*fullH)
	put := func(x, y int, v byte) {
		cx, cy := x+tx, y+ty
		if cx < 0 || cy < 0 || cx >= fullW || cy >= fullH {
			return
		}
		canvas[cx+cy*fullW] = v
	}
	if flag == 0 {
		// row-major source: sequential bytes are pixel (x + y*w).
		p := spriteOff
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				put(x, y, spriteData[p])
				p++
			}
		}
	} else {
		// column-major source (flag==1, the wall.dat trap): bytes arrive
		// column-by-column (for x { for y }), stored at row-major (x + y*w).
		p := spriteOff
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				put(x, y, spriteData[p])
				p++
			}
		}
	}

	return decodedSprite{palette: colours, idx: canvas, fullW: fullW, fullH: fullH}, true
}

// dominantTexel decodes a single texture sprite (frame 0) and returns the
// most-frequent visible (non-transparent) texel colour, masked with 0xf8f8ff to
// match the client's 5:5:5 quantisation. ok is false when nothing decodes / no
// visible texel.
func dominantTexel(spriteData, indexData []byte) (r, g, b int, ok bool) {
	ds, dok := decodeTextureSprite(spriteData, indexData)
	if !dok {
		return 0, 0, 0, false
	}
	counts := map[int]int{}
	bestColour, bestCount := -1, 0
	for _, ci := range ds.idx {
		c := ds.palette[int(ci)&0xff]
		if c == 0xff00ff {
			continue // transparent border / key
		}
		c &= 0xf8f8ff // 5:5:5 quantise, matches client texture decode
		counts[c]++
		if counts[c] > bestCount {
			bestCount = counts[c]
			bestColour = c
		}
	}
	if bestColour < 0 {
		return 0, 0, 0, false
	}
	return bestColour >> 16 & 0xff, bestColour >> 8 & 0xff, bestColour & 0xff, true
}

// loadTextureBuffers builds the shaded texel buffer for every texture id from
// the authentic textures17.jag, porting Scene.method300's per-texel decode + the
// 4-level shade-mip expansion. It NEVER panics and NEVER fails the renderer: any
// archive/decode error leaves the id absent so the caller falls back to the flat
// sampled colour. Result is memoised via textureBufOnce.
//
// Buffer layout per id (matching method300): level0 = base RGB texels, then
// level1 = base-(base>>3), level2 = base-(base>>2), level3 = base-(base>>2)-
// (base>>3); each level is size*size ints in row-major order, masked 0xf8f8ff.
// Texel value 0x000000 is bumped to 1 (so a genuine black texel isn't read as
// "transparent"); the masked magenta key 0xf800ff becomes 0 and sets hasAlpha.
func loadTextureBuffers() {
	textureBufs = map[int32]*textureBuf{}

	var arc *assets.Archive
	candidates := textureJagSearch
	if p := os.Getenv("WESTWORLD_TEXTURES_JAG"); p != "" {
		candidates = append([]string{p}, candidates...)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		a, err := assets.OpenArchive(p)
		if err != nil {
			continue
		}
		arc = a
		break
	}
	if arc == nil {
		return // no archive -> every id falls back to the flat colour
	}
	indexData, err := arc.Get("index.dat")
	if err != nil || len(indexData) < 5 {
		return
	}

	for id, name := range textureName {
		spriteData, err := arc.Get(name + ".dat")
		if err != nil {
			continue
		}
		// Composite the window/arrowslit/stained-glass etc. SUBTYPE over the base
		// (deob loadTextures). A missing subtype .dat just leaves the base buffer
		// (never worse than before).
		var subData []byte
		if id < len(textureSubtype) {
			if sub := textureSubtype[id]; sub != "" {
				if sd, e := arc.Get(sub + ".dat"); e == nil {
					subData = sd
				}
			}
		}
		if buf := buildTextureBuf(spriteData, subData, indexData); buf != nil {
			textureBufs[int32(id)] = buf
		}
	}
}

// buildTextureBuf decodes one texture sprite and expands it into the 4-level
// shaded texel buffer (Scene.method300). The size class is fullWidth/64 - 1
// (64 -> class0 size 64, 128 -> class1 size 128), matching mudclient's
// loadTexture(... wh/64 - 1). Returns nil on any decode failure or unsupported
// size so the caller falls back to the flat colour. Never panics.
func buildTextureBuf(spriteData, subtypeData, indexData []byte) (buf *textureBuf) {
	defer func() {
		if recover() != nil {
			buf = nil
		}
	}()

	ds, ok := decodeTextureSprite(spriteData, indexData)
	if !ok {
		return nil
	}
	// size class from FULL width (mudclient: wh/64 - 1); only 64 and 128 exist
	// in the authentic set.
	var size int32
	switch ds.fullW {
	case 64:
		size = 64
	case 128:
		size = 128
	default:
		return nil
	}
	// The texel canvas must be exactly size x size (the client renders the
	// sprite into a wh x wh box). Guard so a mismatched entry can't index OOB.
	if ds.fullH != int(size) || len(ds.idx) < int(size*size) {
		return nil
	}

	n := size * size
	// Resolve the base sprite to an RGB canvas, then overlay the subtype where
	// its palette INDEX is non-transparent (Surface.drawSprite: index 0 = skip).
	// deob captures the size class from the BASE; if the subtype's full size
	// differs (e.g. planks=64 base vs window=128 subtype, ids 45) we skip the
	// overlay and keep the plain base — exactly the client's behaviour.
	rgb := make([]int, n)
	for p := int32(0); p < n; p++ {
		rgb[p] = ds.palette[int(ds.idx[p])&0xff]
	}
	if subtypeData != nil {
		if sub, sok := decodeTextureSprite(subtypeData, indexData); sok &&
			sub.fullW == ds.fullW && sub.fullH == ds.fullH && len(sub.idx) >= int(n) {
			for p := int32(0); p < n; p++ {
				if si := int(sub.idx[p]) & 0xff; si != 0 {
					rgb[p] = sub.palette[si]
				}
			}
		}
	}

	texels := make([]int32, n*4)
	hasAlpha := false
	for p := int32(0); p < n; p++ {
		c := rgb[p] & 0xf8f8ff
		if c == 0 {
			c = 1
		} else if c == 0xf800ff {
			c = 0
			hasAlpha = true
		}
		texels[p] = int32(c)
	}
	// 3 darker shade-mip levels (method300:2701-2705). Use uint32 for the >>>
	// (logical) shifts, exactly like the Java >>> on positive RGB values.
	for p := int32(0); p < n; p++ {
		k := uint32(texels[p])
		texels[n+p] = int32((k - (k >> 3)) & 0xf8f8ff)
		texels[n*2+p] = int32((k - (k >> 2)) & 0xf8f8ff)
		texels[n*3+p] = int32((k - (k >> 2) - (k >> 3)) & 0xf8f8ff)
	}
	return &textureBuf{texels: texels, size: size, hasAlpha: hasAlpha}
}

// textureBuffer returns the shaded texel buffer for texture id, or nil when the
// archive/decode was unavailable (memoised). A nil result tells the rasteriser
// to use the existing flat-colour fallback path for that face.
func textureBuffer(id int32) *textureBuf {
	textureBufOnce.Do(loadTextureBuffers)
	return textureBufs[id]
}

func texU16(b []byte, o int) int {
	if o < 0 || o+1 >= len(b) {
		return 0
	}
	return int(b[o]&0xff)<<8 | int(b[o+1]&0xff)
}
