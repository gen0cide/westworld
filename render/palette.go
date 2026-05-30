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

// textureFill maps an RSC texture id to a representative flat 5:5:5 colour
// (method305-encoded) so a textured face renders with the right hue while the
// full textured-span fill + sprite archive are still unported (priority 4
// proper). The colours track the canonical RSC texture set ordering
// (GameData.textureName in the 204 client): stone/brick, woods, foliage,
// water, thatch, cloth, metal, etc. Ids past the table fall back to a neutral
// stone grey — better than the previous uniform near-black grey because the
// gouraud ramp now scales a mid-grey rather than a fixed 96/255 grey.
//
// This is an APPROXIMATION (a single dominant colour per texture, not the real
// sampled texels) — honest stand-in geometry colour, not real texturing.
var textureRGB = [][3]int32{
	{120, 100, 80},  // 0  mossy stone / brick wall
	{96, 152, 255},  // 1  water (animated scroll, single frame) — real textures17.jag
	{60, 120, 40},   // 2  leaves / hedge (foliage)
	{40, 80, 150},   // 3  water
	{200, 180, 120}, // 4  thatch / straw roof
	{110, 110, 115}, // 5  stone slab / cobble
	{150, 40, 30},   // 6  red roof tile
	{90, 70, 50},    // 7  bark / log
	{160, 150, 130}, // 8  sandstone
	{70, 130, 50},   // 9  grass / bush
	{130, 90, 60},   // 10 plank floor
	{180, 175, 165}, // 11 light stone / marble
	{50, 100, 35},   // 12 dark foliage
	{145, 120, 85},  // 13 timber wall
	{60, 60, 65},    // 14 dark metal / iron bars
	{170, 140, 90},  // 15 wattle / daub
	{55, 110, 60},   // 16 mossy
	{80, 136, 247},  // 17 fountain water (animated scroll, single frame) — real textures17.jag
	{200, 130, 70},  // 18 fire / lava glow
	{100, 140, 170}, // 19 glass / ice
}

// textureFill returns the method305-encoded flat colour for a texture id.
func textureFill(id int32) int32 {
	var rgb [3]int32
	if id >= 0 && int(id) < len(textureRGB) {
		rgb = textureRGB[id]
	} else {
		rgb = [3]int32{125, 120, 115} // neutral stone grey
	}
	return method305(rgb[0], rgb[1], rgb[2])
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
