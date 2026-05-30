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

// dominantTexel decodes a single texture sprite (frame 0) exactly as
// Surface.loadSprite does — palette from index.dat, index bytes from spriteData
// — and returns the most-frequent visible (non-transparent) texel colour. The
// palette entry 0 is the 0xff00ff magenta transparency key and is skipped; every
// decoded colour is masked with 0xf8f8ff to match the client's 5:5:5 quantisation
// (Scene texture-load path). ok is false when nothing decodes / no visible texel.
//
// It is fully bounds-checked and recovers from any panic so a malformed entry
// can never crash the renderer.
func dominantTexel(spriteData, indexData []byte) (r, g, b int, ok bool) {
	defer func() {
		if recover() != nil {
			r, g, b, ok = 0, 0, 0, false
		}
	}()

	if len(spriteData) < 2 {
		return 0, 0, 0, false
	}
	// The index offset into index.dat lives in the first two bytes of the
	// sprite payload (Surface.loadSprite: indexOff = getUnsignedShort(spriteData, 0)).
	io := texU16(spriteData, 0)
	if io+5 > len(indexData) {
		return 0, 0, 0, false
	}
	_ = texU16(indexData, io) // fullWidth (unused for sampling)
	io += 2
	_ = texU16(indexData, io) // fullHeight (unused)
	io += 2
	colourCount := int(indexData[io] & 0xff)
	io++
	if colourCount < 1 || io+3*(colourCount-1) > len(indexData) {
		return 0, 0, 0, false
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
		return 0, 0, 0, false
	}
	io += 2 // skip translateX, translateY
	w := texU16(indexData, io)
	io += 2
	h := texU16(indexData, io)
	io += 2
	// indexData[io] is the row/column-major flag; irrelevant for a frequency
	// histogram, so we read texels in storage order regardless.
	size := w * h
	if size <= 0 {
		return 0, 0, 0, false
	}

	// index bytes begin at spriteData[2]; row-major (flag 0) vs column-major
	// (flag 1). For frequency counting the traversal order is irrelevant, so we
	// just read size bytes sequentially.
	spriteOff := 2
	if spriteOff+size > len(spriteData) {
		size = len(spriteData) - spriteOff
		if size <= 0 {
			return 0, 0, 0, false
		}
	}

	counts := map[int]int{}
	bestColour, bestCount := -1, 0
	for p := 0; p < size; p++ {
		ci := int(spriteData[spriteOff+p] & 0xff)
		if ci >= len(colours) {
			continue
		}
		c := colours[ci]
		if c == 0xff00ff {
			continue // transparent
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

func texU16(b []byte, o int) int {
	if o < 0 || o+1 >= len(b) {
		return 0
	}
	return int(b[o]&0xff)<<8 | int(b[o+1]&0xff)
}
