package render

import (
	"os"
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// spritesSearch is the ordered list of candidate locations for OpenRSC's own
// sprite archive Authentic_Sprites.orsc. This is the SINGLE source for every 2D
// sprite the renderer needs — 3D textures (3225+i), item icons (2150+n), UI
// media (2000+), and NPC/player animation frames — exactly as the OpenRSC client
// loads them (GraphicsController.loadSprite keys purely by decimal id). It
// replaces the classic .jag sources (textures17/media58/entity24/config85) and
// pulls NOTHING from the eggsampler/rscdump tree. WESTWORLD_SPRITES_ORSC
// overrides everything.
var spritesSearch = []string{
	"/Users/flint/Code/openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc",
	"/Users/flint/Code/openrsc/Client_Base/Cache/Authentic_Sprites.orsc",
	"/Users/flint/Code/openrsc/Authentic_Sprites.orsc",
}

var (
	spriteArchiveOnce sync.Once
	spriteArchive     *assets.SpriteArchive
)

// sprites returns the lazily-opened, process-wide OpenRSC sprite archive, or nil
// if none of the candidate paths exist (callers then fall back to their legacy
// path / a placeholder, never panicking). Memoised. NOTE: this package-global is
// a stepping stone; Stage 6 threads the archive through render.Bundle explicitly
// and this accessor's search list becomes the Bundle default.
func sprites() *assets.SpriteArchive {
	spriteArchiveOnce.Do(func() {
		candidates := spritesSearch
		if p := os.Getenv("WESTWORLD_SPRITES_ORSC"); p != "" {
			candidates = append([]string{p}, candidates...)
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err != nil {
				continue
			}
			if a, err := assets.OpenSprites(p); err == nil {
				spriteArchive = a
				return
			}
		}
	})
	return spriteArchive
}

// OpenRSC sprite-id bases (mudclient.java:61-65). A sprite id is always
// base + offset; GraphicsController.loadSprite ignores any package name.
const (
	spriteMedia   = 2000 // UI / interface elements
	spriteItem    = 2150 // inventory item icons: 2150 + itemPicture
	spriteTexture = 3225 // 3D textures: 3225 + textureIndex
)
