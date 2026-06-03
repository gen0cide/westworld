package render

import (
	"fmt"
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// content8.go — the AUTHENTIC ground/inventory ITEM-PICTURE source (ground-item
// parity, Task B2). It decodes an item's inventory-icon sprite from the rev-235
// cache content8 archive ("2d graphics"), the SAME bytes the DEOB/JAR B1 legs
// decode — so the orsc raw item layer is byte-identical to the DEOB-reconstructed
// item-sprite canvas (content8 pixels == OpenRSC Authentic_Sprites.orsc[spriteItem+
// pic] pixels; cross-checked in content8_test.go).
//
// content8 stores the item pictures as 30-per-sheet "objects{N}.dat" entries
// (Mudclient.java:2499-2503: sheet N decoded at base spriteBaseNpcs + 30*(N-1),
// 30 frames, flag 109) plus a shared "index.dat". A flat picture index `pic`
// resolves to sheet (pic/30 + 1), frame (pic%30) — the SAME slot math the on-screen
// ground draw uses (spriteIndex = unusedIntsBb[type] + sg; sg=spriteItem=2150). We
// decode the ONE sheet that holds `pic` via the shared multi-frame parser
// (parseContent1Frames — the exact Surface.parseSprite port content1 uses) and
// return the requested frame.
//
// SCOPING: like content1, the content8 source is consulted ONLY when RSC_MESH_CACHE
// is set (the parity harness sets it; the live cradle does not). When the env is
// unset or the pack is absent, the item decode returns nil and the caller falls
// back (the live cradle keeps its Authentic_Sprites.orsc itemsprite.go path
// verbatim — this changes nothing outside a parity run). Never panics.

// itemSheetFrameCount is the per-sheet frame stride of the content8 "objects{N}.dat"
// sheets (Mudclient.java:2499-2503 readMedia parses 30 frames per sheet). pic/30 +1
// selects the sheet, pic%30 the frame within it.
const itemSheetFrameCount = 30

var (
	content8Once sync.Once
	content8Arc  *assets.Archive
	content8Idx  []byte
)

// content8Archive lazily opens the content8 archive + its shared index.dat, or
// (nil,nil) when RSC_MESH_CACHE is unset / the pack is absent / unreadable (caller
// falls back). Memoised process-wide. Reuses content1's cache-dir resolver
// (content1CacheDir) so both entity + item sources read the SAME parity cache.
func content8Archive() (*assets.Archive, []byte) {
	content8Once.Do(func() {
		dir := content1CacheDir()
		if dir == "" {
			return // RSC_MESH_CACHE unset: live cradle uses the Authentic_Sprites path
		}
		arc, idx := openContentArchive(dir, "content8_")
		if arc == nil || idx == nil {
			return
		}
		content8Arc = arc
		content8Idx = idx
	})
	return content8Arc, content8Idx
}

// content8SheetCache memoises the decoded 30-frame sheet blocks per sheet name
// ("objects1".."objects4"), so a second item from the same sheet re-uses the parse.
var (
	content8SheetMu    sync.Mutex
	content8SheetCache = map[string][]*parsedFrame{}
	content8SheetMiss  = map[string]bool{}
)

// content8ItemFrame decodes the inventory-icon frame for a flat picture index `pic`
// from the content8 archive and returns it as an animFrame (the SAME shape
// decodeEntitySprite returns), or nil when content8 is unavailable / the sheet is
// missing / the frame is empty. Transparency keys on palette index 0 (exactly as
// Surface.transparentSpritePlot skips it). fullW/fullH are the shared 48x32 item
// canvas; the frame's trim (transX/transY) places the trimmed bitmap within it.
// Never panics.
func content8ItemFrame(pic int) (f *animFrame) {
	defer func() {
		if recover() != nil {
			f = nil
		}
	}()
	if pic < 0 {
		return nil
	}
	arc, index := content8Archive()
	if arc == nil || index == nil {
		return nil
	}
	sheet := pic/itemSheetFrameCount + 1
	frame := pic % itemSheetFrameCount
	name := fmt.Sprintf("objects%d", sheet)

	content8SheetMu.Lock()
	frames, ok := content8SheetCache[name]
	miss := content8SheetMiss[name]
	content8SheetMu.Unlock()
	if !ok {
		if miss {
			return nil
		}
		data, err := arc.Get(name + ".dat")
		if err != nil || data == nil {
			content8SheetMu.Lock()
			content8SheetMiss[name] = true
			content8SheetMu.Unlock()
			return nil
		}
		frames = parseContent1Frames(data, index, itemSheetFrameCount)
		content8SheetMu.Lock()
		if frames == nil {
			content8SheetMiss[name] = true
		} else {
			content8SheetCache[name] = frames
		}
		content8SheetMu.Unlock()
		if frames == nil {
			return nil
		}
	}
	if frame < 0 || frame >= len(frames) {
		return nil
	}
	pf := frames[frame]
	if pf == nil || pf.width <= 0 || pf.height <= 0 {
		return nil
	}
	fullW, fullH := pf.fullW, pf.fullH
	if fullW <= 0 {
		fullW = pf.width
	}
	if fullH <= 0 {
		fullH = pf.height
	}
	pix := make([]int, pf.width*pf.height)
	for i := range pf.indices {
		idx := int(pf.indices[i]) & 0xff
		if idx == 0 {
			pix[i] = -1 // palette index 0 = transparency key
			continue
		}
		if idx < len(pf.palette) {
			pix[i] = int(pf.palette[idx]) & 0xffffff
		} else {
			pix[i] = -1
		}
	}
	return &animFrame{w: pf.width, h: pf.height, fullW: fullW, fullH: fullH, tx: pf.transX, ty: pf.transY, pix: pix}
}
