package render

import (
	"sync"

	"github.com/gen0cide/westworld/facts"
)

// Ground items in RSC are drawn as their real 2D inventory icon, NOT a 3D model.
// The icon for item id N is a single sprite in OpenRSC's Authentic_Sprites.orsc
// at id (spriteItem + itemIcons[N].pic) — the same lookup the client uses
// (mudclient.loadMediaAuthentic loads item sprites sequentially from spriteItem,
// so flat picture index p resolves to sprite spriteItem+p, and p = the item's
// authentic picture index = OpenRSC ItemDef.spriteID = oracle GameData.itemPicture).
// itemIcons (render/itempicture_data.go, generated from OpenRSC's EntityHandler)
// also carries the per-item pictureMask/blueMask used to recolour the grayscale
// base sprite per tier (recolorItemPixel). NOTE: pic is NOT ItemDef.AppearanceID —
// that is the WORN appearance and resolves to the wrong icon (it was the cause of
// the "every non-wieldable item shows the same wrong sprite" bug).
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
// sprite (spriteItem + itemIcons[id].pic), tier-recoloured by its pictureMask
// (recolorItemPixel), composited onto the 48x32 inventory canvas at its
// XShift/YShift. Memoised per item id. Never panics.
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
	// (icon.pic / oracle GameData.itemPicture), NOT ItemDef.AppearanceID:
	// appearanceID is the WORN/equipment appearance, which points at the wrong
	// sprite for inventory icons and is 0 for every non-wieldable item (which
	// collapsed them all onto a single icon). See render/itempicture_data.go.
	icon, ok := itemIcons[itemID]
	if !ok {
		// No authentic numeric picture for this id (unknown id, or a custom
		// named-pack item like bones) — fall back to the red ground marker.
		itemSpriteMiss[itemID] = true
		return nil
	}
	sp, err := sa.Sprite(spriteItem + icon.pic)
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
			// Tier recolour: tint the grayscale base-sprite pixels by the
			// item's pictureMask (and blueMask). Many tiers share one base
			// sprite and differ only by mask (e.g. all scimitars are sprite 83;
			// bronze→orange, mithril→blue, rune→teal). Coloured pixels (the
			// hilt) are left untouched. See recolorItemPixel.
			p = uint32(recolorItemPixel(int(p), icon.mask, icon.blue))
			if p == 0 {
				continue // recolour collapsed it to the transparency key
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

// recolorItemPixel applies OpenRSC's authentic inventory-icon recolour to a
// single 0x00RRGGBB sprite pixel — a faithful port of the masked branch of
// GraphicsController.plot_trans_scale_with_2_masks (the path item draws take via
// drawSpriteClipping(sprite, .., pictureMask, 0, blueMask, ..)). The rules:
//   - a GRAY pixel (R==G==B) is multiplied by the pictureMask (mask1) — this is
//     the tier tint that recolours the grayscale blade/body;
//   - a white-ish pixel (R==255, G==B) is multiplied by mask2, which is always 0
//     for item draws (so such pixels go black/transparent), matching the client;
//   - the blueMask special case (R==G, B!=G) scales by R*B>>16 when blueMask is set;
//   - any other (already-coloured) pixel — e.g. a gold hilt — is left untouched.
// mask==0 means "no recolour": the base sprite is drawn as-is (this guards the
// many items whose stored sprite is already final-coloured from being zeroed).
func recolorItemPixel(p, mask1, blue int) int {
	if mask1 == 0 {
		return p
	}
	r := (p >> 16) & 0xff
	g := (p >> 8) & 0xff
	b := p & 0xff
	if blue == 0 {
		blue = 0xFFFFFF
	}
	switch {
	case r == g && g == b: // grayscale → tint by pictureMask (mask1)
		r = (r * ((mask1 >> 16) & 0xff)) >> 8
		g = (g * ((mask1 >> 8) & 0xff)) >> 8
		b = (b * (mask1 & 0xff)) >> 8
	case r == 255 && g == b: // white-ish → mask2 (== 0 for item draws → black)
		r, g, b = 0, 0, 0
	case blue != 0xFFFFFF && r == g && b != g: // blueMask special case
		shifter := r * b
		r = (((blue >> 16) & 0xff) * shifter) >> 16
		g = (((blue >> 8) & 0xff) * shifter) >> 16
		b = ((blue & 0xff) * shifter) >> 16
	}
	return (r << 16) | (g << 8) | b
}
