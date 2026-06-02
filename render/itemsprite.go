package render

import (
	"sync"

	"github.com/gen0cide/westworld/facts"
)

// Ground items in RSC are drawn as their real 2D inventory icon, NOT a 3D model.
// The icon for item id N is a single sprite in OpenRSC's Authentic_Sprites.orsc
// at id (spriteItem + ItemDef.AppearanceID) — the same lookup the client uses
// (mudclient.loadMediaAuthentic loads item sprites sequentially from spriteItem
// in blocks of 30, so flat picture index p resolves to sprite spriteItem+p, and
// p = the item's AppearanceID/spriteID). Verified empirically: item 0 (Iron
// Mace) AppearanceID=117 -> sprite 2267 = the mace icon.
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
// sprite (spriteItem + AppearanceID) composited onto the 48x32 inventory canvas
// at its XShift/YShift. Memoised per item id. Never panics.
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

	def := f.ItemDefs[itemID]
	if def == nil {
		itemSpriteMiss[itemID] = true
		return nil
	}
	sp, err := sa.Sprite(spriteItem + def.AppearanceID)
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
