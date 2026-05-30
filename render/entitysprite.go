package render

import (
	"os"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/assets"
)

// RSC characters (players + NPCs) are 2D billboard sprites, NOT 3D models. Each
// is composited from up to 12 body-part "animation" sprites (head/body/legs/…)
// decoded from entity24.jag, stacked into a shared fullWidth x fullHeight canvas
// via per-frame translateX/translateY, and recoloured per the dye/skin rule
// (Surface.transparentSpritePlot, Surface.java:1643). The layer set + per-layer
// animation name + clothing/skin colours come from config85.jag (GameData), the
// classic config that matches entity24.jag's sprite ordering (the OpenRSC
// NpcDefs.json sprite ids are renumbered and do NOT line up with entity24.jag).
//
// We render FRAME 0 of every layer (the front-facing standing/idle pose at
// camera yaw 0: drawNpc computes l1 = (animationCurrent + (0+16)/32) & 7 = 0,
// i2 = 0, j2 = 0, so the drawn frame within each <name>.dat is frame 0). The
// back-to-front layer draw order is npcAnimationArray[0]
// (mudclient.java:7161). Everything here never panics and falls back gracefully
// to nil so the 3D-cross billboards (BuildEntities) can stand in.

// entitySprite search paths for the classic entity24.jag (Version.ENTITY = 24).
// WESTWORLD_ENTITY_JAG overrides everything.
var entityJagSearch = []string{
	"/Users/flint/Code/rscdump.com-runescape-classic-dump/eggsampler-rsc-204-d223fc6b77db/eggsampler-rsc-204-d223fc6b77db/data/entity24.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/entity24.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/video/entity24.jag",
	"/Users/flint/Code/openrsc/entity24.jag",
}

// config85.jag search paths (Version.CONFIG = 85). WESTWORLD_CONFIG_JAG overrides.
var configJagSearch = []string{
	"/Users/flint/Code/rscdump.com-runescape-classic-dump/eggsampler-rsc-204-d223fc6b77db/eggsampler-rsc-204-d223fc6b77db/data/config85.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/config85.jag",
	"/Users/flint/Code/openrsc/Client_Base/Cache/video/config85.jag",
	"/Users/flint/Code/openrsc/config85.jag",
}

// npcAnimationArray[0]: the front-facing (l1 == 0) back-to-front body-part draw
// order. Each value indexes npcSprite[npc][layer] (mudclient.java:7161).
var npcLayerOrder = [12]int{11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4}

// characterSkinColours / characterTopBottomColours / characterHairColours are
// the authentic palettes (mudclient.java:7228 / 7246 / 7299). A character's
// clothing colour field holds an INDEX into one of these tables (when small);
// the skin field is a direct 24-bit colour. The charColour marker on each
// animation (1=hair, 2=top, 3=bottom) selects which clothing table/index drives
// that layer's dye.
var characterSkinColours = []int{0xecded0, 0xccb366, 0xb38c40, 0x997326, 0x906020}

var characterTopBottomColours = []int{
	0xff0000, 0xff8000, 0xffe000, 0xa0e000, 57344, 32768, 41088, 45311, 33023, 12528,
	0xe000e0, 0x303030, 0x604000, 0x805000, 0xffffff,
}

var characterHairColours = []int{
	0xffc030, 0xffa040, 0x805030, 0x604020, 0x303030, 0xff6020, 0xff4000, 0xffffff, 65280, 65535,
}

// entityArchive holds the opened sprite + config archives plus the parsed
// GameData tables we need. Loaded once, never re-parsed.
type entityArchive struct {
	sprites  *assets.Archive
	indexDat []byte

	// parsed config85.jag (GameData) — only the NPC + animation tables.
	npcCount      int
	npcSprite     [][12]int // -1 = no layer
	npcColourHair []int
	npcColourTop  []int
	npcColourBtm  []int
	npcColourSkin []int
	npcWidth      []int          // npcSomething_1 (camera1)
	npcHeight     []int          // npcSomething_2 (camera2)
	npcByName     map[string]int // lowercased name -> first config85 npc id

	animCount        int
	animationName    []string
	animationCharCol []int
}

var (
	entityArchiveOnce sync.Once
	entityArc         *entityArchive // nil if either archive could not load
)

// loadEntityArchive opens entity24.jag + config85.jag and parses the GameData
// NPC/animation tables. It NEVER panics and leaves entityArc nil on any failure
// so callers fall back to the 3D-cross billboards. Memoised via the Once.
func loadEntityArchive() {
	defer func() {
		if recover() != nil {
			entityArc = nil
		}
	}()

	sprites := openArchiveCandidates("WESTWORLD_ENTITY_JAG", entityJagSearch)
	if sprites == nil {
		return
	}
	indexDat, err := sprites.Get("index.dat")
	if err != nil || len(indexDat) < 5 {
		return
	}

	cfg := openArchiveCandidates("WESTWORLD_CONFIG_JAG", configJagSearch)
	if cfg == nil {
		return
	}
	ea := &entityArchive{sprites: sprites, indexDat: indexDat}
	if !ea.parseConfig(cfg) {
		return
	}
	entityArc = ea
}

// openArchiveCandidates returns the first openable archive among the env
// override (if set) followed by the search list, or nil.
func openArchiveCandidates(envVar string, search []string) *assets.Archive {
	candidates := search
	if p := os.Getenv(envVar); p != "" {
		candidates = append([]string{p}, search...)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if a, err := assets.OpenArchive(p); err == nil {
			return a
		}
	}
	return nil
}

// gameDataReader walks the GameData string.dat / integer.dat substreams exactly
// as GameData.loadData does (GameData.java:14-219).
type gameDataReader struct {
	str, intg []byte
	so, io    int
	bad       bool
}

func (g *gameDataReader) ub() int {
	if g.io >= len(g.intg) {
		g.bad = true
		return 0
	}
	v := int(g.intg[g.io] & 0xff)
	g.io++
	return v
}

func (g *gameDataReader) us() int {
	if g.io+1 >= len(g.intg) {
		g.bad = true
		return 0
	}
	v := int(g.intg[g.io]&0xff)<<8 | int(g.intg[g.io+1]&0xff)
	g.io += 2
	return v
}

func (g *gameDataReader) ui() int {
	if g.io+3 >= len(g.intg) {
		g.bad = true
		return 0
	}
	v := int(g.intg[g.io]&0xff)<<24 | int(g.intg[g.io+1]&0xff)<<16 |
		int(g.intg[g.io+2]&0xff)<<8 | int(g.intg[g.io+3]&0xff)
	g.io += 4
	if v > 0x5f5e0ff {
		v = 0x5f5e0ff - v
	}
	return v
}

func (g *gameDataReader) gs() string {
	var sb strings.Builder
	for g.so < len(g.str) && g.str[g.so] != 0 {
		sb.WriteByte(g.str[g.so])
		g.so++
	}
	g.so++ // skip the terminator
	return sb.String()
}

// parseConfig parses the item/npc/texture/animation prefix of GameData.loadData
// — only as far as the animation table (we don't need objects/wall/etc.). It
// stops early on any out-of-range read (g.bad) and returns false so the caller
// discards the partially-parsed archive. Mirrors GameData.java:41-219 field for
// field so the substream offsets stay aligned.
func (ea *entityArchive) parseConfig(cfg *assets.Archive) bool {
	sd, err := cfg.Get("string.dat")
	if err != nil {
		return false
	}
	id, err := cfg.Get("integer.dat")
	if err != nil {
		return false
	}
	g := &gameDataReader{str: sd, intg: id}

	// --- items (skip everything; only the offsets matter) ---
	itemCount := g.us()
	for i := 0; i < itemCount; i++ {
		g.gs()
	}
	for i := 0; i < itemCount; i++ {
		g.gs()
	}
	for i := 0; i < itemCount; i++ {
		g.gs()
	}
	for i := 0; i < itemCount; i++ {
		g.us()
	}
	for i := 0; i < itemCount; i++ {
		g.ui()
	}
	for i := 0; i < itemCount; i++ {
		g.ub()
	}
	for i := 0; i < itemCount; i++ {
		g.ub()
	}
	for i := 0; i < itemCount; i++ {
		g.us()
	}
	for i := 0; i < itemCount; i++ {
		g.ui()
	}
	for i := 0; i < itemCount; i++ {
		g.ub()
	}
	for i := 0; i < itemCount; i++ {
		g.ub()
	}

	// --- npcs ---
	npcCount := g.us()
	if g.bad || npcCount < 0 || npcCount > 100000 {
		return false
	}
	npcName := make([]string, npcCount)
	for i := 0; i < npcCount; i++ {
		npcName[i] = g.gs()
	}
	for i := 0; i < npcCount; i++ {
		g.gs() // description
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // attack
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // strength
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // hits
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // defense
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // attackable
	}
	npcSprite := make([][12]int, npcCount)
	for i := 0; i < npcCount; i++ {
		for s := 0; s < 12; s++ {
			v := g.ub()
			if v == 255 {
				v = -1
			}
			npcSprite[i][s] = v
		}
	}
	npcColourHair := make([]int, npcCount)
	for i := range npcColourHair {
		npcColourHair[i] = g.ui()
	}
	npcColourTop := make([]int, npcCount)
	for i := range npcColourTop {
		npcColourTop[i] = g.ui()
	}
	npcColourBtm := make([]int, npcCount)
	for i := range npcColourBtm {
		npcColourBtm[i] = g.ui()
	}
	npcColourSkin := make([]int, npcCount)
	for i := range npcColourSkin {
		npcColourSkin[i] = g.ui()
	}
	npcWidth := make([]int, npcCount)
	for i := range npcWidth {
		npcWidth[i] = g.us() // npcSomething_1 (camera1 / billboard width)
	}
	npcHeight := make([]int, npcCount)
	for i := range npcHeight {
		npcHeight[i] = g.us() // npcSomething_2 (camera2 / billboard height)
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // walkModel
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // combatModel
	}
	for i := 0; i < npcCount; i++ {
		g.ub() // combatAnimation
	}
	for i := 0; i < npcCount; i++ {
		g.gs() // command
	}

	// --- textures (names only; skip) ---
	textureCount := g.us()
	for i := 0; i < textureCount; i++ {
		g.gs()
	}
	for i := 0; i < textureCount; i++ {
		g.gs()
	}

	// --- animations (name + charColour are all we need) ---
	animCount := g.us()
	if g.bad || animCount < 0 || animCount > 100000 {
		return false
	}
	animationName := make([]string, animCount)
	for i := range animationName {
		animationName[i] = g.gs()
	}
	animationCharCol := make([]int, animCount)
	for i := range animationCharCol {
		animationCharCol[i] = g.ui()
	}
	if g.bad {
		return false
	}

	npcByName := make(map[string]int, npcCount)
	for i := 0; i < npcCount; i++ {
		key := strings.ToLower(npcName[i])
		if _, ok := npcByName[key]; !ok {
			npcByName[key] = i // first id wins (the canonical base NPC)
		}
	}

	ea.npcCount = npcCount
	ea.npcSprite = npcSprite
	ea.npcColourHair = npcColourHair
	ea.npcColourTop = npcColourTop
	ea.npcColourBtm = npcColourBtm
	ea.npcColourSkin = npcColourSkin
	ea.npcWidth = npcWidth
	ea.npcHeight = npcHeight
	ea.npcByName = npcByName
	ea.animCount = animCount
	ea.animationName = animationName
	ea.animationCharCol = animationCharCol
	return true
}

// animFrame is one decoded body-part sprite (frame 0 of a <name>.dat block),
// positioned by translateX/translateY inside the shared fullWidth x fullHeight
// figure canvas. pix holds 0x00RRGGBB or -1 for transparent.
type animFrame struct {
	w, h, fullW, fullH, tx, ty int
	pix                        []int
}

// decodeAnimFrame0 decodes frame 0 of the named animation's <name>.dat sprite
// block, exactly as Surface.loadSprite (Surface.java:371-423) reads its first
// frame: shared palette + 15 per-frame headers from index.dat, pixel bytes from
// the payload at offset 2. Palette index 0 and the 0xff00ff magenta key are
// transparent. flag==1 means column-major pixel order. Returns nil on any
// failure (missing entry, malformed header). Recovers from panics.
func (ea *entityArchive) decodeAnimFrame0(name string) (f *animFrame) {
	defer func() {
		if recover() != nil {
			f = nil
		}
	}()

	spriteData, err := ea.sprites.Get(name + ".dat")
	if err != nil || len(spriteData) < 2 {
		return nil
	}
	idx := ea.indexDat
	io := entU16(spriteData, 0)
	if io+5 > len(idx) {
		return nil
	}
	fullW := entU16(idx, io)
	io += 2
	fullH := entU16(idx, io)
	io += 2
	colourCount := int(idx[io] & 0xff)
	io++
	if colourCount < 1 || io+3*(colourCount-1) > len(idx) {
		return nil
	}
	colours := make([]int, colourCount)
	colours[0] = 0xff00ff // transparency key
	for i := 0; i < colourCount-1; i++ {
		colours[i+1] = int(idx[io]&0xff)<<16 | int(idx[io+1]&0xff)<<8 | int(idx[io+2]&0xff)
		io += 3
	}

	// frame 0 header: translateX, translateY (u8), width, height (u16), flag (u8)
	if io+6 > len(idx) {
		return nil
	}
	tx := int(idx[io] & 0xff)
	io++
	ty := int(idx[io] & 0xff)
	io++
	w := entU16(idx, io)
	io += 2
	h := entU16(idx, io)
	io += 2
	flag := int(idx[io] & 0xff)
	io++

	size := w * h
	if size <= 0 || fullW <= 0 || fullH <= 0 {
		return nil
	}
	const spriteOff = 2
	if spriteOff+size > len(spriteData) {
		return nil
	}

	pix := make([]int, size)
	for i := range pix {
		pix[i] = -1
	}
	off := spriteOff
	if flag == 0 {
		for p := 0; p < size; p++ {
			ci := int(spriteData[off] & 0xff)
			off++
			if ci != 0 && ci < colourCount {
				if c := colours[ci]; c != 0xff00ff {
					pix[p] = c
				}
			}
		}
	} else {
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				ci := int(spriteData[off] & 0xff)
				off++
				if ci != 0 && ci < colourCount {
					if c := colours[ci]; c != 0xff00ff {
						pix[x+y*w] = c
					}
				}
			}
		}
	}
	return &animFrame{w: w, h: h, fullW: fullW, fullH: fullH, tx: tx, ty: ty, pix: pix}
}

// recolourTexel applies the Surface.transparentSpritePlot dye/skin rule
// (Surface.java:1677-1682): a grey palette colour (R==G==B) is multiplied by the
// dye (clothing) colour; an R==255 && G==B colour is multiplied by the skin
// colour; everything else is used unchanged. A zero dye/skin means "no recolour"
// (treated as 0xffffff, matching spriteClipping's j1==0 / k1==0 guard).
func recolourTexel(c, dye, skin int) int {
	if dye == 0 {
		dye = 0xffffff
	}
	if skin == 0 {
		skin = 0xffffff
	}
	r := c >> 16 & 0xff
	g := c >> 8 & 0xff
	b := c & 0xff
	if r == g && g == b {
		return ((r * (dye >> 16 & 0xff) >> 8) << 16) |
			((g * (dye >> 8 & 0xff) >> 8) << 8) |
			(b * (dye & 0xff) >> 8)
	}
	if r == 255 && g == b {
		return ((r * (skin >> 16 & 0xff) >> 8) << 16) |
			((g * (skin >> 8 & 0xff) >> 8) << 8) |
			(b * (skin & 0xff) >> 8)
	}
	return c
}

// CompositeSprite is a fully decoded, recoloured standing character billboard:
// a fullW x fullH RGB canvas with a per-pixel transparency mask. The renderer
// blits it depth-scaled. pix is row-major 0x00RRGGBB; opaque[i] reports whether
// pixel i is drawn (transparent pixels are skipped during the blit).
type CompositeSprite struct {
	W, H   int
	Pix    []int32
	Opaque []bool
}

// resolveClothingColour maps a character clothing colour field (hair/top/bottom)
// to a 24-bit RGB. Small values index the authentic colour table; out-of-range
// values are used as a direct 24-bit colour (the authentic client passes the
// field straight through to spriteClipping's dye).
func resolveClothingColour(field int, table []int) int {
	if field >= 0 && field < len(table) {
		return table[field]
	}
	return field
}

// dyeForLayer resolves the dye + skin colour for one layer given its animation's
// charColour marker and the character's clothing/skin colours. Marker 1 = hair,
// 2 = top, 3 = bottom (each with skin); any other marker is a literal colour used
// directly with no skin recolour (Surface.transparentSpritePlot, mudclient
// drawNpc :2141-2152).
func dyeForLayer(charColour, hair, top, bottom, skin int) (dye, skinOut int) {
	switch charColour {
	case 1:
		return resolveClothingColour(hair, characterHairColours), skin
	case 2:
		return resolveClothingColour(top, characterTopBottomColours), skin
	case 3:
		return resolveClothingColour(bottom, characterTopBottomColours), skin
	default:
		return charColour, 0
	}
}

// layerSpec is one resolved body-part layer ready to composite.
type layerSpec struct {
	animName   string
	charColour int
}

// composite decodes + recolours + stacks the given layers (already in
// back-to-front draw order) into one CompositeSprite. hair/top/bottom/skin are
// the character's colours. Returns nil if nothing decoded.
func (ea *entityArchive) composite(layers []layerSpec, hair, top, bottom, skin int) *CompositeSprite {
	var fullW, fullH int
	type decoded struct {
		f          *animFrame
		dye, skinC int
	}
	var ds []decoded
	for _, l := range layers {
		f := ea.decodeAnimFrame0(l.animName)
		if f == nil {
			continue
		}
		if f.fullW > fullW {
			fullW = f.fullW
		}
		if f.fullH > fullH {
			fullH = f.fullH
		}
		dye, skinC := dyeForLayer(l.charColour, hair, top, bottom, skin)
		ds = append(ds, decoded{f, dye, skinC})
	}
	if fullW <= 0 || fullH <= 0 || len(ds) == 0 {
		return nil
	}
	cs := &CompositeSprite{
		W:      fullW,
		H:      fullH,
		Pix:    make([]int32, fullW*fullH),
		Opaque: make([]bool, fullW*fullH),
	}
	for _, d := range ds {
		f := d.f
		for yy := 0; yy < f.h; yy++ {
			dy := f.ty + yy
			if dy < 0 || dy >= fullH {
				continue
			}
			for xx := 0; xx < f.w; xx++ {
				c := f.pix[xx+yy*f.w]
				if c < 0 {
					continue
				}
				dx := f.tx + xx
				if dx < 0 || dx >= fullW {
					continue
				}
				rc := recolourTexel(c, d.dye, d.skinC)
				idx := dy*fullW + dx
				cs.Pix[idx] = int32(rc & 0xffffff)
				cs.Opaque[idx] = true
			}
		}
	}
	return cs
}

// npcLayers builds the back-to-front layer specs for an NPC from npcSprite +
// npcAnimationArray[0]. Each non-empty layer's sprite id is mapped to its
// animation name + charColour via the GameData tables.
func (ea *entityArchive) npcLayers(npcID int) []layerSpec {
	if npcID < 0 || npcID >= ea.npcCount {
		return nil
	}
	var layers []layerSpec
	for _, layer := range npcLayerOrder {
		spriteID := ea.npcSprite[npcID][layer]
		if spriteID < 0 || spriteID >= ea.animCount {
			continue
		}
		layers = append(layers, layerSpec{
			animName:   ea.animationName[spriteID],
			charColour: ea.animationCharCol[spriteID],
		})
	}
	return layers
}

var (
	npcCompositeMu    sync.Mutex
	npcCompositeCache = map[int]*CompositeSprite{}
	npcCompositeMiss  = map[int]bool{} // ids that failed to composite (don't retry)
)

// compositeNPC returns the cached standing-frame billboard for an NPC id, or nil
// if the archives are unavailable or the NPC has no valid sprite layers (caller
// then falls back to the 3D-cross billboard). Memoised per id.
func compositeNPC(npcID int) *CompositeSprite {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return nil
	}
	npcCompositeMu.Lock()
	defer npcCompositeMu.Unlock()
	if cs, ok := npcCompositeCache[npcID]; ok {
		return cs
	}
	if npcCompositeMiss[npcID] {
		return nil
	}
	layers := entityArc.npcLayers(npcID)
	if len(layers) == 0 {
		npcCompositeMiss[npcID] = true
		return nil
	}
	cs := entityArc.composite(
		layers,
		entityArc.npcColourHair[npcID],
		entityArc.npcColourTop[npcID],
		entityArc.npcColourBtm[npcID],
		entityArc.npcColourSkin[npcID],
	)
	if cs == nil {
		npcCompositeMiss[npcID] = true
		return nil
	}
	npcCompositeCache[npcID] = cs
	return cs
}

// npcBillboardSize returns the world-space billboard width/height for an NPC
// (npcSomething_1 / _2; the values drawSprite passes). Falls back to the default
// human 145x220 when unknown.
func npcBillboardSize(npcID int) (w, h int) {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil || npcID < 0 || npcID >= entityArc.npcCount {
		return playerBillboardW, playerBillboardH
	}
	w = entityArc.npcWidth[npcID]
	h = entityArc.npcHeight[npcID]
	if w <= 0 {
		w = playerBillboardW
	}
	if h <= 0 {
		h = playerBillboardH
	}
	return w, h
}

// Default local-player ("bernard") appearance — the standing human shown at the
// centre of his own view: legs1 + body1 + head1, default clothing colours
// (mudclient appearance defaults). Billboard is a fixed 145x220 (drawSprite at
// mudclient.java:3647).
const (
	playerBillboardW = 145
	playerBillboardH = 220

	playerSkinColour   = 0xecded0 // characterSkinColours[0]
	playerHairColIdx   = 2        // characterHairColours[2] = 0x805030
	playerTopColIdx    = 8        // characterTopBottomColours[8] = 0x80ff (blue)
	playerBottomColIdx = 14       // characterTopBottomColours[14] = 0xffffff (white)
)

var (
	playerCompositeOnce sync.Once
	playerComposite     *CompositeSprite
)

// compositePlayer returns the cached default-human billboard for the local
// player. Returns nil if the archives are unavailable (caller falls back to the
// 3D-cross billboard). The layers are legs1/body1/head1 (charColour markers
// 3/2/1 resolve to bottom/top/hair); skin is direct.
func compositePlayer() *CompositeSprite {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return nil
	}
	playerCompositeOnce.Do(func() {
		// Back-to-front: legs, then body, then head (head1 is layer 0 in
		// npcAnimationArray[0]'s tail; for the fixed human appearance the visible
		// order is legs -> body -> head, matching the appearance preview at
		// mudclient.java:1133-1141 and the layer order's relative positions).
		layers := []layerSpec{
			{animName: "legs1", charColour: 3},
			{animName: "body1", charColour: 2},
			{animName: "head1", charColour: 1},
		}
		playerComposite = entityArc.composite(
			layers,
			playerHairColIdx,   // hair index
			playerTopColIdx,    // top index
			playerBottomColIdx, // bottom index
			playerSkinColour,   // skin direct
		)
	})
	return playerComposite
}

// NpcIDForName resolves an NPC name to its config85.jag npc id (the id that
// drives sprite compositing). Matching is case-insensitive; the first id with
// that name wins. Returns -1 if the archives are unavailable or the name is
// unknown. This is the bridge from a fact's NPC name (which IS stable across the
// OpenRSC + classic data) to the classic sprite id, sidestepping the renumbered
// OpenRSC NpcDefs.json sprite ids.
func NpcIDForName(name string) int {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return -1
	}
	if id, ok := entityArc.npcByName[strings.ToLower(name)]; ok {
		return id
	}
	return -1
}

func entU16(b []byte, o int) int {
	if o < 0 || o+1 >= len(b) {
		return 0
	}
	return int(b[o]&0xff)<<8 | int(b[o+1]&0xff)
}
