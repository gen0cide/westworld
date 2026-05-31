package render

import "math"

// groundColour is the 256-entry ground-colour palette anIntArray578
// (World.<init>), each entry an encoded flat 5:5:5 fill from method305.
var groundColour [256]int32

// tileDef is one GroundOverlay (getTileDecoration) definition, ported verbatim
// from OpenRSC EntityHandler.loadTileDefinitions / server TileDef.xml (verified
// byte-for-byte against both). colour follows the SAME face-fill convention the
// renderer already uses: a value >= 0 is a TEXTURE INDEX (routed to the
// perspective textured-span filler via scene.go), a value < 0 is a method305
// flat colour. tileType drives the authentic class behaviour: 2 = indoor floor,
// 3 = outdoor textured surface, 4 = water (forced to the water texture), 5 =
// bridge (sentinel colour 12345678, out of scope here).
type tileDef struct {
	colour   int32
	tileType byte
}

// tileDefs is the 25-entry authentic table. Decoration id N maps to
// tileDefs[N-1] (mirrors the client's getTileDef(id-1)); id 0 = no overlay
// (grass underlay shows). This REPLACES the old hand-rolled flat-colour map so
// indoor floors (chapel planks=3->tex3, marble=17->tex15, pentagram=14->tex32),
// outdoor textured ids, and the authentic road/dirt flat colours all render
// exactly as the client classifies them (World.java:714-726).
var tileDefs = []tileDef{
	{-16913, 1},   // 1  road / path (flat)
	{1, 3},        // 2  water-edge texture
	{3, 2},        // 3  WOOD planks floor (chapel) -> tex 3
	{3, 4},        // 4  water-class (planks colour but type 4)
	{-16913, 2},   // 5  indoor flat
	{-27685, 2},   // 6  indoor flat (reddish)
	{25, 3},       // 7  gungywater texture
	{12345678, 5}, // 8  bridge sentinel
	{-26426, 1},   // 9  flat
	{-1, 5},       // 10 bridge
	{31, 3},       // 11 lava texture
	{3, 4},        // 12 water-class
	{-4534, 2},    // 13 indoor flat
	{32, 2},       // 14 pentagram -> tex 32
	{-9225, 2},    // 15 indoor flat
	{-3172, 2},    // 16 indoor flat
	{15, 2},       // 17 MARBLE floor -> tex 15
	{-2, 2},       // 18 indoor flat
	{-1, 3},       // 19 outdoor flat special
	{-2, 4},       // 20 water-class
	{-2, 4},       // 21 water-class
	{-2, 0},       // 22 flat
	{-17793, 2},   // 23 dirt / planter (flat)
	{-14594, 1},   // 24 flat
	{1, 3},        // 25 water-edge texture
}

// overlayDef returns the tileDef for a GroundOverlay id (1..25) and ok=true, or
// the zero def + false for id 0 / out-of-range (grass underlay shows).
func overlayDef(id byte) (tileDef, bool) {
	if id >= 1 && int(id) <= len(tileDefs) {
		return tileDefs[id-1], true
	}
	return tileDef{}, false
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
