package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// depthFar is the far sentinel every Depth cell is reset to: nothing is "in
// front of" it, so an untouched (sky) pixel never occludes a sprite. Chosen
// well beyond clipFar so a real face depth always compares as nearer.
const depthFar = int32(1 << 30)

// Surface is the int[] software framebuffer (Surface.java). Pix holds 0x00RRGGBB
// pixels in row-major order; the renderer writes directly into it and we encode
// the result as a PNG (replacing the AWT ImageProducer handoff).
//
// Depth is a per-pixel coarse camera-Z buffer the scene rasterizer writes
// alongside Pix (one entry per pixel, same row-major layout). RSC's real
// renderer painter-sorts character sprites INTO the same face list (Scene.java
// :430-435), so they occlude correctly; this renderer blits sprites in a 2D
// pass after the 3D scene, so it instead consults Depth to skip sprite pixels
// that sit behind nearer scene geometry. Depth stores the per-FACE average
// camera-Z (coarse is fine — RSC face spans are small).
type Surface struct {
	Pix    []int32
	Depth  []int32
	Width  int
	Height int
}

func NewSurface(w, h int) *Surface {
	s := &Surface{
		Pix:    make([]int32, w*h),
		Depth:  make([]int32, w*h),
		Width:  w,
		Height: h,
	}
	for i := range s.Depth {
		s.Depth[i] = depthFar
	}
	return s
}

// Clear fills the whole framebuffer with a single colour and resets Depth to
// the far sentinel.
func (s *Surface) Clear(rgb int32) {
	for i := range s.Pix {
		s.Pix[i] = rgb
	}
	for i := range s.Depth {
		s.Depth[i] = depthFar
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
// pixels. Pixels outside the surface are clipped. This is the 2D
// character-billboard blit (the depth-scaled equivalent of Surface.spriteClipping
// after the perspective divide has already produced screen-space dst extents).
//
// camZ is the sprite's camera-space depth (the project() foot depth, already
// biased toward the camera by the caller). Each destination pixel is skipped
// when camZ is GREATER than the scene Depth stored there — i.e. nearer scene
// geometry occludes the sprite (a mob behind a wall is hidden). The sprite does
// NOT write Depth: it is the last pass, so sprites never occlude each other via
// the buffer (their mutual order is the caller's far-first painter sort).
//
// When flip is true the source is sampled mirrored left-to-right (source column
// cs.W-1-srcX), so the W/SW/NW facings — which RSC draws by horizontally
// mirroring the E/SE/NE poses — render as the correct side
// (Surface.spriteClipping's mirror branch, mudclient drawNpc flag arg).
func (s *Surface) BlitSpriteScaled(cs *CompositeSprite, dstX, dstY, dstW, dstH int, flip bool, camZ int32) {
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
			dst := py*s.Width + px
			// Occlusion: skip when scene geometry at this pixel is nearer than
			// the sprite (sprite is farther => behind the wall/building).
			if camZ > s.Depth[dst] {
				continue
			}
			s.Pix[dst] = cs.Pix[idx]
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
