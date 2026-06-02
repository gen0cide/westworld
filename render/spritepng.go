package render

import (
	"bytes"
	"image"
	"image/png"

	"github.com/gen0cide/westworld/facts"
)

// ItemSpritePNG returns the item's authentic inventory icon as a transparent
// PNG, or (nil,false) if the sprite archives are unavailable or the id has no
// decodable icon. The image is the sprite's full-frame canvas (classically
// 48x32) with a real alpha channel: decoded-transparent pixels are alpha 0.
//
// It is concurrency-safe (compositeItem guards its caches) and resolves the
// item-id -> sprite mapping from the item def's AppearanceID (carried by f) into
// Authentic_Sprites.orsc. Used by cradle's GET /sprite route so the browser UI
// can render real RSC item icons in inventory/bank/shop cells. Returns
// (nil,false) when f is nil (the item-def table is the appearance-id source).
func ItemSpritePNG(f *facts.Facts, itemID int) ([]byte, bool) {
	cs := compositeItem(f, itemID)
	if cs == nil || cs.W <= 0 || cs.H <= 0 || len(cs.Pix) < cs.W*cs.H {
		return nil, false
	}
	return compositeSpritePNG(cs), true
}

// compositeSpritePNG encodes a CompositeSprite as a straight-alpha NRGBA PNG.
// cs.Pix holds 0x00RRGGBB; cs.Opaque marks drawn (alpha 255) vs transparent
// (alpha 0) pixels. A set cs.Flip mirrors the icon horizontally.
func compositeSpritePNG(cs *CompositeSprite) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, cs.W, cs.H))
	for y := 0; y < cs.H; y++ {
		for x := 0; x < cs.W; x++ {
			src := y*cs.W + x
			dx := x
			if cs.Flip {
				dx = cs.W - 1 - x
			}
			di := img.PixOffset(dx, y)
			rgb := cs.Pix[src]
			a := uint8(0)
			if src < len(cs.Opaque) && cs.Opaque[src] {
				a = 255
			}
			img.Pix[di+0] = uint8(rgb >> 16)
			img.Pix[di+1] = uint8(rgb >> 8)
			img.Pix[di+2] = uint8(rgb)
			img.Pix[di+3] = a
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
