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
