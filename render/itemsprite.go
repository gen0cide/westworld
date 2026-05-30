package render

import (
	"strconv"
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// Ground items in RSC are drawn as their real 2D inventory icon, NOT a 3D model.
// Each icon is one frame of an objects<N>.dat sprite block in media58.jag
// (Version.MEDIA = 58), the same archive that holds the inventory/UI sprites.
// The item-id -> sprite mapping is GameData.itemPicture[id] (config85.jag): a
// flat sprite index `p` into the objects*.dat run, where loadMedia loads up to
// 30 frames per objects<j>.dat block (mudclient.java:3085-3092,
// spriteItem + (j-1)*30). So item id -> icon resolves as:
//
//	p     = entityArc.itemPicture[id]
//	block = objects<(p/30)+1>.dat
//	frame = p % 30
//
// and the frame decodes with the IDENTICAL Surface.loadSprite machinery the
// entity body-part frames use (decodeSpriteFrame: shared palette + 6-byte
// per-frame headers from index.dat + the 0xff00ff transparency key). Item icons
// share a 48x32 inventory canvas; the per-frame translateX/translateY position
// the (smaller) drawn pixels inside it (mudclient draws them with spriteClipping
// at 48x32). We composite each frame onto its full 48x32 canvas as a
// CompositeSprite so the renderer blits it depth-scaled like a character
// billboard. Everything here never panics and falls back to nil so the existing
// red ground-item marker can stand in.

// mediaJagSearch lists the classic media58.jag locations. WESTWORLD_MEDIA_JAG
// overrides everything (openArchiveCandidates).
var mediaJagSearch = []string{
	"/Users/flint/Code/rscdump.com-runescape-classic-dump/eggsampler-rsc-204-d223fc6b77db/eggsampler-rsc-204-d223fc6b77db/data/media58.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/media58.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/video/media58.jag",
	"/Users/flint/Code/openrsc/media58.jag",
}

// itemArchive holds the opened media58.jag sprite archive + its index.dat. The
// id->sprite mapping (itemPicture) lives on the shared entityArc (parsed from
// config85.jag), so the item loader piggybacks on the already-loaded config.
type itemArchive struct {
	sprites  *assets.Archive
	indexDat []byte
}

var (
	itemArchiveOnce sync.Once
	itemArc         *itemArchive // nil if media58.jag could not load
)

// loadItemArchive opens media58.jag + its index.dat. It NEVER panics and leaves
// itemArc nil on any failure so callers fall back to the red marker. Memoised.
func loadItemArchive() {
	defer func() {
		if recover() != nil {
			itemArc = nil
		}
	}()
	sprites := openArchiveCandidates("WESTWORLD_MEDIA_JAG", mediaJagSearch)
	if sprites == nil {
		return
	}
	indexDat, err := sprites.Get("index.dat")
	if err != nil || len(indexDat) < 5 {
		return
	}
	itemArc = &itemArchive{sprites: sprites, indexDat: indexDat}
}

var (
	itemSpriteMu    sync.Mutex
	itemSpriteCache = map[int]*CompositeSprite{}
	itemSpriteMiss  = map[int]bool{} // item ids that failed (don't retry)
)

// compositeItem returns the cached inventory-icon billboard for an item id, or
// nil if media58.jag / config85.jag are unavailable or the id has no decodable
// icon (caller then falls back to the red ground marker). The icon is the
// item's objects*.dat frame composited onto its full 48x32 inventory canvas (so
// the icon's intra-canvas placement matches the inventory rendering). Memoised
// per item id. Never panics.
func compositeItem(itemID int) (cs *CompositeSprite) {
	defer func() {
		if recover() != nil {
			cs = nil
		}
	}()
	entityArchiveOnce.Do(loadEntityArchive)
	itemArchiveOnce.Do(loadItemArchive)
	if entityArc == nil || itemArc == nil {
		return nil
	}
	if itemID < 0 || itemID >= entityArc.itemCount {
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

	p := entityArc.itemPicture[itemID]
	if p < 0 {
		itemSpriteMiss[itemID] = true
		return nil
	}
	block := "objects" + strconv.Itoa(p/30+1) + ".dat"
	frame := p % 30
	f := decodeSpriteFrame(itemArc.sprites, itemArc.indexDat, block, frame)
	if f == nil || f.fullW <= 0 || f.fullH <= 0 {
		itemSpriteMiss[itemID] = true
		return nil
	}

	cs = &CompositeSprite{
		W:      f.fullW,
		H:      f.fullH,
		Pix:    make([]int32, f.fullW*f.fullH),
		Opaque: make([]bool, f.fullW*f.fullH),
	}
	for yy := 0; yy < f.h; yy++ {
		dy := f.ty + yy
		if dy < 0 || dy >= f.fullH {
			continue
		}
		for xx := 0; xx < f.w; xx++ {
			c := f.pix[xx+yy*f.w]
			if c < 0 {
				continue
			}
			dx := f.tx + xx
			if dx < 0 || dx >= f.fullW {
				continue
			}
			idx := dy*f.fullW + dx
			cs.Pix[idx] = int32(c & 0xffffff)
			cs.Opaque[idx] = true
		}
	}
	itemSpriteCache[itemID] = cs
	return cs
}
