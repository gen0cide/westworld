package orsc

import "github.com/gen0cide/westworld/assets"

// textures_content11.go loads the AUTHENTIC rev-235 "Textures" content archive
// (content11) into the Scene's resource database — the data-plumbing that makes
// TEXTURED scenery faces (e.g. the well's 18 wall/planks faces) and textured
// terrain overlays render with their real texels instead of the
// 64-magenta-transparent stub that silently skipped every textured face.
//
// This is a faithful Go port of Mudclient.loadTextures (Mudclient.java:2658-2716)
// + Surface.parseSprite (Surface.java:838-891) + Surface.drawWorld (the
// frequency-histogram 256-colour quantiser, Surface.java:948-1006). Each texture
// id 0..54 maps to a `<base>.dat` entry in content11 (resolved by the EntityHandler
// order, authenticTextureNames below) sharing one `index.dat`. The DEOB pipeline:
//
//	parseSprite(scratch, base.dat, dummy=88, index)   // decode palette + index grid
//	drawBox(magenta, 128x128)                          // magenta back-transparent box
//	drawSprite(-1, scratch, 0, 0)                      // blit the base sprite over it
//	[if subname: parseSprite(scratch, sub.dat, 109); drawSprite(-1, scratch, 0, 0)]
//	capture size×size into TEX                          // size = spriteWidthFull[scratch]
//	chroma-key: green 0x00FF00 -> magenta 0xFF00FF      // in the captured RGB
//	drawWorld(TEX)                                      // re-quantise to 256 colours
//	scene.defineTexture(id, palette, size/64-1, indices)
//
// We reconstruct the composited RGB box headlessly (no AWT Surface needed): the
// magenta box + base/sub sprite blit is a pure pixel composite, and drawWorld's
// quantiser is the SAME histogram quantiser as orsc.quantizeTexture (texture.go
// proves the texel banks come out byte-identical across orsc/DEOB/JAR when fed the
// same source RGB, because the DEOB re-quantises the composited RGB rather than
// passing through the .dat's original palette). So feeding quantizeTexture the
// reconstructed RGB box yields the identical (palette, indices, sizeClass).

// authenticTextureNames is the rev-235 texture id -> (base, sub) binding, VERBATIM
// from OpenRSC EntityHandler.loadTextureDefinitions (EntityHandler.java:252-306).
// Index == texture id. base names a `<base>.dat` entry in content11; a non-empty
// sub names a `<sub>.dat` overlay sprite blended over the base (the wall/door,
// planks/window, … variants). ALL THREE engines bake this SAME ordered list so
// the texel banks bind to the same texture ids.
var authenticTextureNames = [][2]string{
	{"wall", "door"}, {"water", ""}, {"wall", ""}, {"planks", ""},
	{"wall", "doorway"}, {"wall", "window"}, {"roof", ""}, {"wall", "arrowslit"},
	{"leafytree", ""}, {"treestump", ""}, {"fence", ""}, {"mossy", ""},
	{"railings", ""}, {"painting1", ""}, {"painting2", ""}, {"marble", ""},
	{"deadtree", ""}, {"fountain", ""}, {"wall", "stainedglass"}, {"target", ""},
	{"books", ""}, {"timbered", ""}, {"timbered", "timberwindow"}, {"mossybricks", ""},
	{"growingwheat", ""}, {"gungywater", ""}, {"web", ""}, {"wall", "desertwindow"},
	{"wall", "crumbled"}, {"cavern", ""}, {"cavern2", ""}, {"lava", ""},
	{"pentagram", ""}, {"mapletree", ""}, {"yewtree", ""}, {"helmet", ""},
	{"canvas", "tentbottom"}, {"Chainmail2", ""}, {"mummy", ""}, {"jungleleaf", ""},
	{"jungleleaf3", ""}, {"jungleleaf4", ""}, {"jungleleaf5", ""}, {"jungleleaf6", ""},
	{"mossybricks", "arrowslit"}, {"planks", "window"}, {"planks", "junglewindow"}, {"cargonet", ""},
	{"bark", ""}, {"canvas", ""}, {"canvas", "tentdoor"}, {"wall", "lowcrumbled"},
	{"cavern", "crumbled"}, {"cavern2", "crumbled"}, {"lava", "flames"},
}

// parsedSprite holds the decoded output of Surface.parseSprite for one frame: the
// 256-colour-or-less palette (index 0 == magenta transparent key), the per-pixel
// palette index grid (spriteWidth × spriteHeight), the trim/translate offsets, the
// frame width/height, and the full (untrimmed) box width/height.
type parsedSprite struct {
	palette                []int32
	indices                []int8
	transX, transY         int
	width, height          int
	fullWidth, fullHeight  int
}

// parseContent11Sprite is the Go port of Surface.parseSprite (Surface.java:838-891)
// for the FIRST frame (frameCount=1, the only frame used by loadTextures). data is
// the `<base>.dat` entry; index is the shared `index.dat`. dummy selects the frame
// header offset inside index (88 for the base sprite, 109 for the subname overlay)
// — but the layout is the SAME; the dummy is an anti-tamper no-op in the deob, so
// we ignore it. Returns nil for a malformed entry.
func parseContent11Sprite(data, index []byte) *parsedSprite {
	if len(data) < 2 || len(index) < 5 {
		return nil
	}
	dataOff := getUShort(data, 0) // offset of pixel data inside data
	if dataOff+5 > len(index) {
		return nil
	}
	fullWidth := getUShort(index, dataOff)
	dataOff += 2
	fullHeight := getUShort(index, dataOff)
	dataOff += 2
	paletteSize := int(index[dataOff]) & 0xff
	dataOff++
	palette := make([]int32, paletteSize)
	if paletteSize > 0 {
		palette[0] = 0xff00ff // index 0 is transparent magenta
	}
	for i := 0; i < paletteSize-1; i++ {
		if dataOff+3 > len(index) {
			return nil
		}
		palette[i+1] = int32((int(index[dataOff])&0xff)<<16 +
			(int(index[dataOff+1])&0xff)<<8 +
			(int(index[dataOff+2]) & 0xff))
		dataOff += 3
	}

	// Frame header (Surface.java:854-862), pixelOff starts at 2 (skip the 2-byte
	// data offset). frameCount==1 so a single frame is read. The header is 7 bytes:
	// transX(1) transY(1) w(2) h(2) layout(1) — the LAST byte read is index[dataOff+6].
	// (flames.dat's frame header ends exactly at the index.dat tail, so the bound
	// must be +7, not +8, or the lava/flames overlay parse spuriously fails and the
	// id-54 texel bank diverges from the DEOB/JAR oracle.)
	pixelOff := 2
	if dataOff+7 > len(index) {
		return nil
	}
	transX := int(index[dataOff]) & 0xff
	dataOff++
	transY := int(index[dataOff]) & 0xff
	dataOff++
	width := getUShort(index, dataOff)
	dataOff += 2
	height := getUShort(index, dataOff)
	dataOff += 2
	layout := int(index[dataOff]) & 0xff
	dataOff++

	area := width * height
	indices := make([]int8, area)
	if layout == 0 {
		for i := 0; i < area; i++ {
			if pixelOff >= len(data) {
				return nil
			}
			indices[i] = int8(data[pixelOff])
			pixelOff++
		}
	} else if layout == 1 {
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				if pixelOff >= len(data) {
					return nil
				}
				indices[x+y*width] = int8(data[pixelOff])
				pixelOff++
			}
		}
	}

	return &parsedSprite{
		palette: palette, indices: indices,
		transX: transX, transY: transY,
		width: width, height: height,
		fullWidth: fullWidth, fullHeight: fullHeight,
	}
}

// compositeOnMagenta replicates the loadTextures pixel composite: a fullWidth ×
// fullWidth magenta box (Surface.drawBox 0xFF00FF), with the base sprite (and
// optional subname overlay) blitted over it at its trim offset (Surface.drawSprite
// -> blitSpriteIndexed: palette-index 0 is the per-sprite transparent pixel and is
// skipped). The captured square is fullWidth × fullWidth (the texture is square;
// loadTextures captures size×size with size = spriteWidthFull). Returns the RGB
// box (0x00RRGGBB row-major) and its edge size, or (nil,0) on failure.
func compositeOnMagenta(base, sub *parsedSprite) ([]int32, int) {
	if base == nil {
		return nil, 0
	}
	size := base.fullWidth
	if size <= 0 {
		return nil, 0
	}
	box := make([]int32, size*size)
	for i := range box {
		box[i] = 0xff00ff // drawBox(0xFF00FF, size, size)
	}
	blit := func(sp *parsedSprite) {
		if sp == nil {
			return
		}
		// blitSpriteIndexed: dst at (transX,transY), palette-index 0 is transparent.
		for sy := 0; sy < sp.height; sy++ {
			dy := sp.transY + sy
			if dy < 0 || dy >= size {
				continue
			}
			for sx := 0; sx < sp.width; sx++ {
				dx := sp.transX + sx
				if dx < 0 || dx >= size {
					continue
				}
				idx := int(sp.indices[sx+sy*sp.width]) & 0xff
				if idx == 0 {
					continue // transparent texel
				}
				if idx < len(sp.palette) {
					box[dx+dy*size] = sp.palette[idx]
				}
			}
		}
	}
	blit(base)
	blit(sub)
	// Chroma-key: green 0x00FF00 -> magenta 0xFF00FF (Mudclient.java:2701-2705),
	// so the green key colour collapses onto the same transparency key as the box.
	for i, c := range box {
		if c&0xffffff == 0x00ff00 {
			box[i] = 0xff00ff
		}
	}
	return box, size
}

// LoadTexturesFromContent11 loads the authentic texture set from the content11
// "Textures" archive into the Scene, for each of the (up to textureCount) names in
// authenticTextureNames. Mirrors Mudclient.loadTextures: parse base (+ optional
// subname) sprite from content11, composite onto a magenta box, chroma-key, then
// re-quantise the RGB box (drawWorld == quantizeTexture) and register via
// LoadTexture. A missing/odd entry is skipped (degrades to the flat-skip path), so
// it never panics on a partial archive.
func (s *Scene) LoadTexturesFromContent11(arc *assets.Archive) {
	if arc == nil {
		return
	}
	index, err := arc.Get("index.dat")
	if err != nil || index == nil {
		return
	}
	count := len(authenticTextureNames)
	if count > s.textureCount {
		count = s.textureCount
	}
	for id := 0; id < count; id++ {
		baseName := authenticTextureNames[id][0]
		subName := authenticTextureNames[id][1]
		baseData, err := arc.Get(baseName + ".dat")
		if err != nil || baseData == nil {
			continue
		}
		base := parseContent11Sprite(baseData, index)
		if base == nil {
			continue
		}
		var sub *parsedSprite
		if subName != "" {
			if subData, err := arc.Get(subName + ".dat"); err == nil && subData != nil {
				sub = parseContent11Sprite(subData, index)
			}
		}
		box, size := compositeOnMagenta(base, sub)
		if box == nil || (size != textureSizeClass0 && size != textureSizeClass1) {
			continue
		}
		palette, indices := quantizeDrawWorld(box)
		s.LoadTexture(id, palette, size/64-1, indices)
	}
}

// quantizeDrawWorld is the EXACT Go port of Surface.drawWorld (Surface.java:948-1006):
// the frequency-histogram 256-colour quantiser the content11 texture pipeline uses.
//
// It is DELIBERATELY NOT quantizeRGB: quantizeRGB ports loadTexturesAuthentic
// (mudclient.java:14658), which maps every pure-black 0x000000 pixel to the magenta
// key 0xff00ff BEFORE the index pass. drawWorld has NO such black->magenta step — a
// black pixel is indexed to the nearest palette entry by RGB distance, which for a
// composited texture (palette holds both 0xff00ff at slot 0 AND a real near-black
// at some slot) is the near-black slot (dist 0), NOT slot 0. The well's textures
// have no black texels so the two agree there, but the wall+arrowslit (id 7) and
// mossybricks+arrowslit (id 44) overlays DO carry black texels: using quantizeRGB
// there diverged from the DEOB/JAR (708 indices wrong, mapped to magenta-transparent
// instead of the real black). This faithful drawWorld port fixes that.
func quantizeDrawWorld(src []int32) (palette []int32, indices []int8) {
	area := len(src)
	histogram := make([]int, 32768)
	for i := 0; i < area; i++ {
		c := int(src[i])
		histogram[((c>>9)&31744)+((c&0xf800)>>6)+((c>>3)&31)]++
	}

	palette = make([]int32, 256)
	palette[0] = 0xff00ff
	freq := make([]int, 256)
	for bucket := 0; bucket < 32768; bucket++ {
		count := histogram[bucket]
		if count > freq[255] {
			for slot := 1; slot < 256; slot++ {
				if freq[slot] < count {
					for k := 255; k > slot; k-- {
						palette[k] = palette[k-1]
						freq[k] = freq[k-1]
					}
					palette[slot] = int32(((bucket&31)<<3) + ((bucket << 6) & 0xf800) + ((bucket << 9) & 0xf80000) + 0x040404)
					freq[slot] = count
					break
				}
			}
		}
		histogram[bucket] = -1 // reuse as bucket->paletteIndex cache
	}

	indices = make([]int8, area)
	for i := 0; i < area; i++ {
		c := int(src[i])
		bucket := ((c >> 3) & 31) + ((c & 0xf800) >> 6) + ((c & 0xf80000) >> 9)
		pi := histogram[bucket]
		if pi == -1 {
			best := 999999999
			r, g, b := (c>>16)&0xff, (c>>8)&0xff, c&0xff
			for p := 0; p < 256; p++ {
				pc := int(palette[p])
				pr, pg, pb := (pc>>16)&0xff, (pc>>8)&0xff, pc&0xff
				if dist := (b-pb)*(b-pb) + (r-pr)*(r-pr) + (g-pg)*(g-pg); dist < best {
					pi = p
					best = dist
				}
			}
			histogram[bucket] = pi
		}
		indices[i] = int8(pi)
	}
	return palette, indices
}

// getUShort reads a big-endian u16 (Surface.getUShort): ((b[o]&0xff)<<8)+(b[o+1]&0xff).
func getUShort(b []byte, o int) int {
	return (int(b[o])&0xff)<<8 + (int(b[o+1]) & 0xff)
}
