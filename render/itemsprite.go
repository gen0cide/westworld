package render

import (
	"sync"

	"github.com/gen0cide/westworld/facts"
)

// Ground items in RSC are drawn as their real 2D inventory icon, NOT a 3D model.
// The icon for item id N is a single sprite in OpenRSC's Authentic_Sprites.orsc
// at id (spriteItem + itemPictureIndex[N]) — the same lookup the client uses
// (mudclient.loadMediaAuthentic loads item sprites sequentially from spriteItem,
// so flat picture index p resolves to sprite spriteItem+p, and p = the item's
// authentic picture index = OpenRSC ItemDef.spriteID = oracle GameData.itemPicture).
// itemPictureIndex (render/itempicture_data.go) is generated from OpenRSC's
// EntityHandler. NOTE: this is NOT ItemDef.AppearanceID — that is the WORN
// appearance and resolves to the wrong icon (it was the cause of the "every
// non-wieldable item shows the same wrong sprite" bug).
//
// 2D icons key transparency on BLACK (0x000000), NOT the magenta key the 3D
// textures use. Each icon is composited onto the standard 48x32 inventory canvas
// at the sprite's XShift/YShift (when RequiresShift), matching the client's
// drawSpriteClipping(48,32). Everything here never panics and falls back to nil
// so the existing red ground-item marker can stand in.

const itemIconW, itemIconH = 48, 32

var (
	itemSpriteMu    sync.Mutex
	itemSpriteCache = map[int]*CompositeSprite{}
	itemSpriteMiss  = map[int]bool{} // item ids that failed (don't retry)
)

// compositeItem returns the cached inventory-icon billboard for an item id, or
// nil if Authentic_Sprites.orsc is unavailable or the id has no decodable icon
// (caller then falls back to the red ground marker). The icon is the item's
// sprite (spriteItem + itemPictureIndex[id]) composited onto the 48x32 inventory
// canvas at its XShift/YShift. Memoised per item id. Never panics.
func compositeItem(f *facts.Facts, itemID int) (cs *CompositeSprite) {
	defer func() {
		if recover() != nil {
			cs = nil
		}
	}()
	if f == nil || itemID < 0 {
		return nil
	}
	sa := sprites()
	if sa == nil {
		return nil
	}

	itemSpriteMu.Lock()
	defer itemSpriteMu.Unlock()
	if c, ok := itemSpriteCache[itemID]; ok {
		return c
	}
	if itemSpriteMiss[itemID] {
		return nil
	}

	// The inventory-icon sprite index is the item's authentic picture index
	// (itemPictureIndex / oracle GameData.itemPicture), NOT ItemDef.AppearanceID:
	// appearanceID is the WORN/equipment appearance, which points at the wrong
	// sprite for inventory icons and is 0 for every non-wieldable item (which
	// collapsed them all onto a single icon). See render/itempicture_data.go.
	pic, ok := itemPictureIndex[itemID]
	if !ok {
		// No authentic numeric picture for this id (unknown id, or a custom
		// named-pack item like bones) — fall back to the red ground marker.
		itemSpriteMiss[itemID] = true
		return nil
	}
	sp, err := sa.Sprite(spriteItem + pic)
	if err != nil || sp == nil || sp.Width <= 0 || sp.Height <= 0 {
		itemSpriteMiss[itemID] = true
		return nil
	}

	ox, oy := 0, 0
	if sp.RequiresShift {
		ox, oy = sp.XShift, sp.YShift
	}
	cs = &CompositeSprite{
		W:      itemIconW,
		H:      itemIconH,
		Pix:    make([]int32, itemIconW*itemIconH),
		Opaque: make([]bool, itemIconW*itemIconH),
	}
	for y := 0; y < sp.Height; y++ {
		dy := oy + y
		if dy < 0 || dy >= itemIconH {
			continue
		}
		for x := 0; x < sp.Width; x++ {
			p := sp.Pixels[y*sp.Width+x] & 0xffffff
			if p == 0 {
				continue // BLACK is the transparency key for 2D icons
			}
			dx := ox + x
			if dx < 0 || dx >= itemIconW {
				continue
			}
			idx := dy*itemIconW + dx
			cs.Pix[idx] = int32(p)
			cs.Opaque[idx] = true
		}
	}
	itemSpriteCache[itemID] = cs
	return cs
}
