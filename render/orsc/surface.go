package orsc

// surface.go — the raster target's behaviour. The Surface struct, its fields,
// the depthFar sentinel, and NewSurface are declared in types.go (the
// foundation contract). This file adds the operations the Shader/Scene fill
// into and the GraphicsController->Image handoff.
//
// Surface.Pixels IS GraphicsController.pixelData (the int[] the Java Shader
// writes), 0x00RRGGBB row-major (GraphicsController.java:43,145-147). The AWT
// ImageProducer step (GraphicsController -> java.awt.Image via MemoryImageSource)
// is replaced here by toImage / ToImage / PNG, which materialise the same int[]
// as a Go image.Image (and PNG bytes).

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Clear fills the whole framebuffer with one colour and resets the depth buffer
// to the far sentinel. This stands in for the per-frame background fill the
// mudclient does before drawScene (the sky colour) plus the implicit depth
// reset; Scene.endScene assumes a cleared target.
//
// rgb is 0x00RRGGBB to match Surface.Pixels / GraphicsController.pixelData.
func (s *Surface) Clear(rgb int32) {
	for i := range s.Pixels {
		s.Pixels[i] = rgb
	}
	for i := range s.Depth {
		s.Depth[i] = depthFar
	}
}

// SetPixel writes one clipped pixel into Pixels (0x00RRGGBB). Out-of-bounds
// writes are dropped — the Java code clips its spans before writing, so this is
// only a guard for callers outside the scanline fill.
func (s *Surface) SetPixel(x, y int, rgb int32) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.Pixels[y*s.Width+x] = rgb
}

// SetDepth writes the coarse camera-Z for one clipped pixel into the Depth
// buffer (depthFar = nothing there yet). The Java renderer painter-sorts sprite
// faces into the same polygon list (Scene.java:430-435) and so needs no depth
// buffer; this Go port keeps a per-pixel coarse depth so a later 2D sprite pass
// (the harness's billboards) can be occluded by nearer scene geometry. A smaller
// stored value is nearer the camera.
func (s *Surface) SetDepth(x, y int, z int32) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.Depth[y*s.Width+x] = z
}

// image materialises Pixels (0x00RRGGBB, fully opaque) into an *image.RGBA. This
// is the replacement for GraphicsController.drawGraphics' MemoryImageSource /
// ImageProducer handoff: the same int[] becomes an in-memory raster.
func (s *Surface) image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, s.Width, s.Height))
	for y := 0; y < s.Height; y++ {
		row := y * s.Width
		for x := 0; x < s.Width; x++ {
			p := s.Pixels[row+x]
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(p >> 16 & 0xff),
				G: uint8(p >> 8 & 0xff),
				B: uint8(p & 0xff),
				A: 0xff,
			})
		}
	}
	return img
}

// ToImage returns the framebuffer as a Go image.Image (0x00RRGGBB -> opaque
// RGBA), the high-level handoff replacing the AWT Image.
func (s *Surface) ToImage() image.Image { return s.image() }

// toImage PNG-encodes Pixels (0x00RRGGBB) and returns the bytes — the exact
// signature the foundation contract pins for the AWT ImageProducer replacement
// (types.go: `(*Surface).toImage() ([]byte, error)`).
func (s *Surface) toImage() ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, s.image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PNG is the exported alias the render harness calls to get PNG bytes.
func (s *Surface) PNG() ([]byte, error) { return s.toImage() }
