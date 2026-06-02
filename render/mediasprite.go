package render

import (
	"bytes"
	"image"
	"image/png"
)

// MediaSpritePNG returns the UI "media" sprite (spriteMedia + id) as a
// transparent PNG, or (nil,false) when the sprite archive is unavailable or the
// id has no decodable sprite. Unlike ItemSpritePNG these are flat 2D interface
// graphics (tab strips, the compass, button chrome) that need no item-def table
// or recolour/compositing — so the signature takes no *facts.Facts. The raw
// sprite's black colour-key (0x000000, per assets.Sprite's doc) is keyed out to
// alpha 0; every other pixel is emitted opaque. Used by cradle's GET
// /sprite?kind=media route so the browser UI can draw authentic RSC tab icons.
func MediaSpritePNG(id int) ([]byte, bool) {
	a := sprites()
	if a == nil {
		return nil, false
	}
	s, _ := a.Sprite(spriteMedia + id)
	if s == nil || s.Width <= 0 || s.Height <= 0 || len(s.Pixels) < s.Width*s.Height {
		return nil, false
	}
	// Same straight-alpha NRGBA encoding shape as compositeSpritePNG, but the
	// source is a raw media sprite (0x00RRGGBB, black = colour-key transparent).
	img := image.NewNRGBA(image.Rect(0, 0, s.Width, s.Height))
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			src := y*s.Width + x
			di := img.PixOffset(x, y)
			rgb := s.Pixels[src] & 0xFFFFFF
			a := uint8(255)
			if rgb == 0x000000 {
				a = 0
			}
			img.Pix[di+0] = uint8(rgb >> 16)
			img.Pix[di+1] = uint8(rgb >> 8)
			img.Pix[di+2] = uint8(rgb)
			img.Pix[di+3] = a
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes(), true
}
