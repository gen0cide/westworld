package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Surface is the int[] software framebuffer (Surface.java). Pix holds 0x00RRGGBB
// pixels in row-major order; the renderer writes directly into it and we encode
// the result as a PNG (replacing the AWT ImageProducer handoff).
type Surface struct {
	Pix    []int32
	Width  int
	Height int
}

func NewSurface(w, h int) *Surface {
	return &Surface{Pix: make([]int32, w*h), Width: w, Height: h}
}

// Clear fills the whole framebuffer with a single colour.
func (s *Surface) Clear(rgb int32) {
	for i := range s.Pix {
		s.Pix[i] = rgb
	}
}

// SetPixel writes one clipped pixel.
func (s *Surface) SetPixel(x, y int, rgb int32) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.Pix[y*s.Width+x] = rgb
}

// BlitSpriteScaled draws a CompositeSprite scaled to dstW x dstH with its
// top-left at (dstX, dstY), nearest-neighbour sampled, skipping transparent
// pixels. Pixels outside the surface are clipped by SetPixel. This is the 2D
// character-billboard blit (the depth-scaled equivalent of Surface.spriteClipping
// after the perspective divide has already produced screen-space dst extents).
//
// When flip is true the source is sampled mirrored left-to-right (source column
// cs.W-1-srcX), so the W/SW/NW facings — which RSC draws by horizontally
// mirroring the E/SE/NE poses — render as the correct side
// (Surface.spriteClipping's mirror branch, mudclient drawNpc flag arg).
func (s *Surface) BlitSpriteScaled(cs *CompositeSprite, dstX, dstY, dstW, dstH int, flip bool) {
	if cs == nil || dstW <= 0 || dstH <= 0 || cs.W <= 0 || cs.H <= 0 {
		return
	}
	for dy := 0; dy < dstH; dy++ {
		py := dstY + dy
		if py < 0 || py >= s.Height {
			continue
		}
		sy := dy * cs.H / dstH
		if sy >= cs.H {
			sy = cs.H - 1
		}
		row := sy * cs.W
		for dx := 0; dx < dstW; dx++ {
			px := dstX + dx
			if px < 0 || px >= s.Width {
				continue
			}
			sx := dx * cs.W / dstW
			if sx >= cs.W {
				sx = cs.W - 1
			}
			if flip {
				sx = cs.W - 1 - sx
			}
			idx := row + sx
			if !cs.Opaque[idx] {
				continue
			}
			s.Pix[py*s.Width+px] = cs.Pix[idx]
		}
	}
}

// PNG encodes the framebuffer as a PNG byte slice.
func (s *Surface) PNG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, s.Width, s.Height))
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			p := s.Pix[y*s.Width+x]
			img.Set(x, y, color.RGBA{
				R: uint8(p >> 16 & 0xff),
				G: uint8(p >> 8 & 0xff),
				B: uint8(p & 0xff),
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
