package orsc

import "github.com/gen0cide/westworld/assets"

// textureload.go wires OpenRSC's Authentic_Sprites.orsc textures into the Scene's
// resource database — the data-plumbing the harness needs that the pure three/
// port leaves to the client's loadTexturesAuthentic (mudclient.java:14650).
//
// Each terrain/model face whose fill is >= 0 names a texture id; world.go emits
// the classic ids (water=1, planks=3, gungywater=25, lava=31, …). We load the
// authentic texture set [0, textureSlots) — sprite spriteTextureBase+i — quantise
// each to a 256-colour palette + index grid exactly as loadTexturesAuthentic, and
// hand it to Scene.LoadTexture (which builds the shade-mip resource buffer).

const (
	// textureSlots sizes the Scene's texture arrays; larger than any texture id a
	// terrain face fill carries, so no LRU eviction occurs for a static render.
	textureSlots = 64
	// spriteTextureBase is mudclient.spriteTexture (mudclient.java:65): texture id
	// i is sprite (3225 + i) in Authentic_Sprites.orsc.
	spriteTextureBase = 3225
)

// LoadTexturesFromArchive loads texture ids [0,count) from the OpenRSC sprite
// archive into the Scene. Missing/odd-sized sprites are skipped (the fill path
// degrades a missing texture to flat rather than crashing). Never panics.
func (s *Scene) LoadTexturesFromArchive(sa *assets.SpriteArchive, count int) {
	if sa == nil {
		return
	}
	if count > s.textureCount {
		count = s.textureCount
	}
	for i := 0; i < count; i++ {
		sp, err := sa.Sprite(spriteTextureBase + i)
		if err != nil || sp == nil || sp.Width <= 0 || sp.Width != sp.Height {
			continue
		}
		if sp.Width != textureSizeClass0 && sp.Width != textureSizeClass1 {
			continue
		}
		palette, indices, sizeClass := quantizeTexture(sp)
		s.LoadTexture(i, palette, sizeClass, indices)
	}
}

// quantizeTexture ports mudclient.loadTexturesAuthentic's texture quantiser
// (mudclient.java:14658-14716): build a 5:5:5-bucket histogram, map black ->
// magenta transparency key, pick the 256 most-frequent colours into a palette
// (palette[0]=magenta key) with the +0x40404 5:5:5 reconstruction bump, and
// re-index every pixel to its nearest palette colour. Returns the palette (256
// packed 0x00RRGGBB), the per-texel index bytes, and the size class (0=64px,
// 1=128px) = Something1/64-1. The Scene's buildTextureBuffer then masks 0xf8f8ff
// + builds the shade-mip; palette[0] survives the mask as the transparency key.
func quantizeTexture(sp *assets.Sprite) (palette []int32, indices []int8, sizeClass int) {
	px := sp.Pixels
	n := len(px)
	bucket := func(p int) int {
		return ((p & 0xf80000) >> 9) + ((p & 0xf800) >> 6) + ((p & 0xf8) >> 3)
	}
	hist := make([]int, 32768)
	for _, p := range px {
		hist[bucket(int(p)&0xffffff)]++
	}
	// black -> magenta key (AFTER the histogram, matching the client's order).
	pix := make([]int, n)
	for i, p := range px {
		c := int(p) & 0xffffff
		if c == 0x000000 {
			c = 0xff00ff
		}
		pix[i] = c
	}
	dict := make([]int32, 256)
	dict[0] = 0xff00ff
	counts := make([]int, 256)
	for i1 := 0; i1 < len(hist); i1++ {
		if j1 := hist[i1]; j1 > counts[255] {
			for k1 := 1; k1 < 256; k1++ {
				if j1 <= counts[k1] {
					continue
				}
				for i2 := 255; i2 > k1; i2-- {
					dict[i2] = dict[i2-1]
					counts[i2] = counts[i2-1]
				}
				dict[k1] = int32(((i1&0x7c00)<<9) + ((i1&0x3e0)<<6) + ((i1&0x1f)<<3) + 0x40404)
				counts[k1] = j1
				break
			}
		}
		hist[i1] = -1 // repurpose hist as a bucket->dict-index cache (-1 = unmapped)
	}
	indices = make([]int8, n)
	for l1 := 0; l1 < n; l1++ {
		j2 := pix[l1]
		k2 := bucket(j2)
		l2 := hist[k2]
		if l2 == -1 {
			best := 0x3b9ac9ff
			r, g, b := (j2>>16)&0xff, (j2>>8)&0xff, j2&0xff
			for i4 := 0; i4 < 256; i4++ {
				d := int(dict[i4])
				dr, dg, db := r-((d>>16)&0xff), g-((d>>8)&0xff), b-(d&0xff)
				if dist := dr*dr + dg*dg + db*db; dist < best {
					best, l2 = dist, i4
				}
			}
			hist[k2] = l2
		}
		indices[l1] = int8(l2)
	}
	sizeClass = sp.Width/64 - 1
	return dict, indices, sizeClass
}
