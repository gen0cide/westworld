package render

import "math"

// groundColour is the 256-entry ground-colour palette anIntArray578
// (World.<init>), each entry an encoded flat 5:5:5 fill from method305.
var groundColour [256]int32

// overlayColour maps a GroundOverlay (getTileDecoration) id to its flat
// method305 fill, ported from OpenRSC EntityHandler.loadTileDefinitions
// (TileDef.colour, method305-encoded). The authentic client OVERWRITES a
// tile's underlay colour with the overlay's whenever an overlay is present
// (World.java:714-726). Overlay id 1 is the grey ROAD/PATH that dominates
// town scenes; 23 is the brown dirt of flower planters / farm plots. Water
// overlays (2/11/13) are handled by the water-flatten path in buildTerrain,
// not here. Ids absent from this table fall back to the grass underlay.
var overlayColour = map[byte]int32{
	1:  method305(128, 128, 128), // road / gravel path (grey)
	3:  method305(122, 122, 122), // textured path -> grey approx
	6:  method305(216, 8, 32),    // red floor
	9:  method305(200, 200, 200), // light stone slab
	16: method305(24, 24, 24),    // dark stone / near-black
	23: method305(136, 96, 0),    // dirt / farm plot (planters)
	24: method305(112, 64, 8),    // dark dirt
	25: method305(110, 110, 110), // gravel texture -> grey approx
}

// method305 encodes an (r,g,b) triple as a flat fill: -1 - (r/8)*1024 -
// (g/8)*32 - b/8 (Scene.method305).
func method305(r, g, b int32) int32 {
	return -1 - (r/8)*1024 - (g/8)*32 - b/8
}

// gouraudRamp builds the 256-entry shade ramp for a flat 5:5:5 colour, exactly
// as Scene.method281's colour-miss branch does (Scene.java:1326). fill is the
// encoded flat colour from a face (fill = -1 - colour). ramp[255-j] scales each
// component by j^2/65536, so index 255 is brightest, index 0 is black.
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

// textureFallbackRGB is the LAST-resort approximation used only when the real
// texture bitmap could not be sampled from textures17.jag (archive missing or an
// entry failed to decode). The auto-sampler in textures.go (loadTextureColours)
// is the source of truth: it opens the authentic textures17.jag and computes the
// dominant non-transparent texel colour for every texture id. This small table
// keeps a handful of hand-tuned hues so the renderer still looks plausible with
// no archive present; ids absent from it fall back to neutral stone grey.
var textureFallbackRGB = map[int32][3]int32{
	0:  {120, 100, 80},  // wall (brick)
	1:  {96, 152, 255},  // water
	5:  {110, 110, 115}, // wall (window) -> stone
	17: {80, 136, 247},  // fountain water
	21: {145, 120, 85},  // timbered wall
	25: {64, 96, 117},   // gungywater
	31: {248, 104, 12},  // lava (orange glow)
}

// textureFill returns the method305-encoded flat colour for a texture id. It
// first consults the auto-sampled colours decoded from the authentic texture
// bitmaps (loadTextureColours, memoised); on a miss it uses the small hardcoded
// fallback, and failing that a neutral stone grey. The gouraud ramp then scales
// this flat colour per face (Scene.method281 colour-miss branch).
func textureFill(id int32) int32 {
	textureColourOnce.Do(loadTextureColours)
	if fill, ok := textureColour[id]; ok {
		return fill
	}
	if rgb, ok := textureFallbackRGB[id]; ok {
		return method305(rgb[0], rgb[1], rgb[2])
	}
	return method305(125, 120, 115) // neutral stone grey
}

func init() {
	for i := int32(0); i < 64; i++ {
		groundColour[i] = method305(255-i*4, 255-int32(float64(i)*1.75), 255-i*4)
	}
	for j := int32(0); j < 64; j++ {
		groundColour[j+64] = method305(j*3, 144, 0)
	}
	for k := int32(0); k < 64; k++ {
		groundColour[k+128] = method305(192-int32(float64(k)*1.5), 144-int32(float64(k)*1.5), 0)
	}
	for l := int32(0); l < 64; l++ {
		groundColour[l+192] = method305(96-int32(float64(l)*1.5), 48+int32(float64(l)*1.5), 0)
	}
	_ = math.Pi
}
