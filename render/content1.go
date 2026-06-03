package render

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// content1.go — the AUTHENTIC entity-sprite source (NPC/player parity, Milestone B,
// docs/build/NPC_SPRITE_PARITY_PLAN.md). It re-routes decodeEntitySprite off the
// (dead) OpenRSC Authentic_Sprites.orsc path onto the rev-235 cache content1
// archive ("people and monsters"), the SAME bytes the DEOB/JAR legs decode — so
// the orsc CompositeSprite canvas is byte-identical to a DEOB-reconstructed canvas
// (single-source: content1 pixels == OpenRSC Authentic_Sprites.orsc pixels).
//
// content1 stores one `<name>.dat` entry per animation (keyed by name, not by a
// global sprite number) plus a shared `index.dat`. The live client decodes each
// 27-slot block with Surface.parseSprite(uc, 15, <name>.dat, 83, index)
// (Mudclient.loadEntitySprites, Mudclient.java:2527). So a GLOBAL sprite id
// (animationNumbers()[animID] + frame, e.g. the rat's 837+0) maps back to
// (animation name, frame within the block); we read that name's `.dat`, run the
// MULTI-FRAME parseSprite port, and return the requested frame's animFrame.
//
// SCOPING: the content1 source is consulted ONLY when RSC_MESH_CACHE is explicitly
// set (the parity harness always sets it; the DEOB/JAR legs read the SAME env). The
// live cradle never sets RSC_MESH_CACHE, so it keeps its existing Authentic_Sprites
// .orsc entity-sprite path verbatim — this re-route changes nothing outside a parity
// run. When the env is set but the pack is absent / unreadable, decodeEntitySprite
// still falls back to the Authentic_Sprites.orsc path (never panics).

// content1CacheDir returns the parity cache dir from RSC_MESH_CACHE, or "" when the
// env is unset (the live cradle) — in which case the content1 source is disabled
// and the legacy Authentic_Sprites.orsc path is used.
func content1CacheDir() string {
	return os.Getenv("RSC_MESH_CACHE")
}

var (
	content1Once sync.Once
	content1Arc  *assets.Archive
	content1Idx  []byte
)

// content1Archive lazily opens the content1 archive + its shared index.dat, or
// (nil,nil) when the cache pack is absent / unreadable (so the caller falls back).
// Memoised process-wide. OpenArchive already strips the 6-byte outer header +
// bzip-inflates the JAG body exactly like the DEOB's World.unpackData(128,false,…),
// so the decoded entry bytes match the Java legs.
func content1Archive() (*assets.Archive, []byte) {
	content1Once.Do(func() {
		dir := content1CacheDir()
		if dir == "" {
			return // RSC_MESH_CACHE unset: live cradle uses the Authentic_Sprites path
		}
		matches, _ := filepath.Glob(filepath.Join(dir, "content1_*"))
		if len(matches) == 0 {
			return
		}
		arc, err := assets.OpenArchive(matches[0])
		if err != nil {
			return
		}
		idx, err := arc.Get("index.dat")
		if err != nil || idx == nil {
			return
		}
		content1Arc = arc
		content1Idx = idx
	})
	return content1Arc, content1Idx
}

// parsedFrame is one decoded body-part frame from a multi-frame content1 `.dat`:
// the per-pixel palette index grid (trimmed width × height), the trim/translate
// offset, and the full (untrimmed) figure-canvas size — shared by every frame in
// the block. palette index 0 is the per-sprite transparent key.
type parsedFrame struct {
	palette        []int32
	indices        []int8
	transX, transY int
	width, height  int
	fullW, fullH   int
}

// parseContent1Frames is the faithful Go port of the MULTI-FRAME Surface.parseSprite
// (reference/rsc-client/src/client/scene/Surface.java:838-891): one palette + full
// canvas size (read once from index.dat), then frameCount per-frame headers
// (transX, transY, width, height, layout) followed by their pixel runs (read from
// the contiguous `<name>.dat` pixel stream starting at offset 2). It is the
// multi-frame extension the parity build needs over the texture path's single-frame
// parseContent11Sprite. The `dummy` arg the live client passes (83) is an
// anti-tamper no-op (the layout is identical), so we ignore it. Returns nil on a
// malformed entry.
func parseContent1Frames(data, index []byte, frameCount int) []*parsedFrame {
	if frameCount <= 0 || len(data) < 2 || len(index) < 5 {
		return nil
	}
	dataOff := int(data[0]&0xff)<<8 + int(data[1]&0xff) // offset of pixel data inside index.dat
	if dataOff+5 > len(index) {
		return nil
	}
	fullW := int(index[dataOff]&0xff)<<8 + int(index[dataOff+1]&0xff)
	dataOff += 2
	fullH := int(index[dataOff]&0xff)<<8 + int(index[dataOff+1]&0xff)
	dataOff += 2
	paletteSize := int(index[dataOff]) & 0xff
	dataOff++
	palette := make([]int32, paletteSize)
	if paletteSize > 0 {
		palette[0] = 0xff00ff // index 0 is transparent magenta
	}
	for i := 0; i < paletteSize-1; i++ {
		if dataOff+3 > len(index) {
			return nil
		}
		palette[i+1] = int32((int(index[dataOff])&0xff)<<16 +
			(int(index[dataOff+1])&0xff)<<8 +
			(int(index[dataOff+2]) & 0xff))
		dataOff += 3
	}

	pixelOff := 2 // pixel stream skips the 2-byte data offset
	frames := make([]*parsedFrame, 0, frameCount)
	for f := 0; f < frameCount; f++ {
		if dataOff+6 > len(index) {
			return nil
		}
		transX := int(index[dataOff]) & 0xff
		dataOff++
		transY := int(index[dataOff]) & 0xff
		dataOff++
		width := int(index[dataOff]&0xff)<<8 + int(index[dataOff+1]&0xff)
		dataOff += 2
		height := int(index[dataOff]&0xff)<<8 + int(index[dataOff+1]&0xff)
		dataOff += 2
		layout := int(index[dataOff]) & 0xff
		dataOff++

		area := width * height
		indices := make([]int8, area)
		if layout == 0 {
			for i := 0; i < area; i++ {
				if pixelOff >= len(data) {
					return nil
				}
				indices[i] = int8(data[pixelOff])
				pixelOff++
			}
		} else if layout == 1 {
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					if pixelOff >= len(data) {
						return nil
					}
					indices[x+y*width] = int8(data[pixelOff])
					pixelOff++
				}
			}
		}
		frames = append(frames, &parsedFrame{
			palette: palette, indices: indices,
			transX: transX, transY: transY,
			width: width, height: height,
			fullW: fullW, fullH: fullH,
		})
	}
	return frames
}

// content1FrameCache memoises the decoded frame blocks per animation name (each
// `<name>.dat` decodes to 15 body frames — the parity build exercises frame 0).
var (
	content1FrameMu    sync.Mutex
	content1FrameCache = map[string][]*parsedFrame{}
	content1FrameMiss  = map[string]bool{}
)

// content1Frames returns the decoded 15-frame body block for the named animation
// (the content1 `<name>.dat` entry), memoised, or nil when content1 is unavailable
// / the entry is missing / malformed. frameCount is the body-frame stride (15, the
// live loadEntitySprites value).
func content1Frames(name string) []*parsedFrame {
	arc, index := content1Archive()
	if arc == nil || index == nil {
		return nil
	}
	content1FrameMu.Lock()
	defer content1FrameMu.Unlock()
	if fr, ok := content1FrameCache[name]; ok {
		return fr
	}
	if content1FrameMiss[name] {
		return nil
	}
	data, err := arc.Get(name + ".dat")
	if err != nil || data == nil {
		content1FrameMiss[name] = true
		return nil
	}
	frames := parseContent1Frames(data, index, 15)
	if frames == nil {
		content1FrameMiss[name] = true
		return nil
	}
	content1FrameCache[name] = frames
	return frames
}

// spriteIDToNameFrame reverse-maps a GLOBAL entity sprite id (animationNumbers()
// [animID] + frame) back to (animation name, frame within the 27-slot block) so the
// content1 source can read the right `<name>.dat`. It walks animationNumbers() for
// the block base ≤ spriteID < base+27 whose def name is first-seen at that base
// (the dedup-by-name assignment), and frame = spriteID - base. Returns ("",-1) when
// no block contains the id.
func spriteIDToNameFrame(spriteID int) (string, int) {
	nums := entityAnimNum()
	for i, base := range nums {
		// Only the FIRST def at a given base names the block (duplicates alias it);
		// authenticAnimDefs[i].name is correct for the first occurrence, which is the
		// one whose nums[i] == base was freshly assigned.
		if base <= spriteID && spriteID < base+27 {
			// Confirm i is the first def with this base (the block owner).
			if i == 0 || nums[i-1] != base {
				return authenticAnimDefs[i].name, spriteID - base
			}
		}
	}
	return "", -1
}

// decodeEntitySpriteContent1 reads the entity body-part frame with the given GLOBAL
// sprite id from the content1 archive and converts it to an animFrame (the SAME
// animFrame shape the Authentic_Sprites.orsc path returns), or nil when content1 is
// unavailable / the id has no content1 block / the frame is empty. Transparency
// keys on palette index 0 (the per-sprite transparent texel), exactly as
// Surface.spriteClipping skips it. The figure canvas is fullW × fullH; the frame's
// trim offset (transX/transY) places it within. Never panics.
func decodeEntitySpriteContent1(spriteID int) (f *animFrame) {
	defer func() {
		if recover() != nil {
			f = nil
		}
	}()
	name, frame := spriteIDToNameFrame(spriteID)
	if name == "" || frame < 0 {
		return nil
	}
	frames := content1Frames(name)
	if frames == nil || frame >= len(frames) {
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
