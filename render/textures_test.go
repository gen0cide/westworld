package render

import "testing"

// TestTextureColourSampling derives the flat texture-fill colours from OpenRSC's
// Authentic_Sprites.orsc and checks the dominant-colour sampler against three
// landmark textures. It SKIPS (rather than fails) when no sprite archive is
// found, so it stays green on machines without the OpenRSC cache present.
func TestTextureColourSampling(t *testing.T) {
	textureColourOnce.Do(loadTextureColours)
	if len(textureColour) == 0 {
		t.Skip("Authentic_Sprites.orsc not found; sampler falls back to approximations")
	}

	rgb := func(fill int32) (r, g, b int32) {
		c := -1 - fill
		return c >> 10 & 0x1f * 8, c >> 5 & 0x1f * 8, c & 0x1f * 8
	}

	cases := []struct {
		id         int32
		name       string
		wr, wg, wb int32
	}{
		// Expected values are post-method305 5:5:5 quantisation (each component
		// floored to a multiple of 8): sampled dominant water=(96,152,255),
		// fountain=(80,136,247), lava=(248,104,12).
		{1, "water", 96, 152, 248},
		{17, "fountain", 80, 136, 240},
		{31, "lava", 248, 104, 8},
	}
	for _, c := range cases {
		fill, ok := textureColour[c.id]
		if !ok {
			t.Errorf("texture %d (%s): not sampled", c.id, c.name)
			continue
		}
		r, g, b := rgb(fill)
		t.Logf("texture %2d %-10s sampled=(%d,%d,%d)", c.id, c.name, r, g, b)
		if r != c.wr || g != c.wg || b != c.wb {
			t.Errorf("texture %d (%s): got (%d,%d,%d) want (%d,%d,%d)",
				c.id, c.name, r, g, b, c.wr, c.wg, c.wb)
		}
	}
}
