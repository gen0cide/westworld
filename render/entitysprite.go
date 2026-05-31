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
// We render the STANDING pose for an 8-way facing dir (0..7) computed from the
// entity heading + camera rotation: l1 = (animationCurrent + (cameraRotation+16)
// /32) & 7. drawNpc maps dir 5/6/7 -> i2 3/2/1 with a horizontal flip (W/SW/NW
// are the mirror images of E/SE/NE), picks the per-direction back-to-front layer
// order npcAnimationArray[l1], and within each <name>.dat block draws standing
// frame j2 = i2*3 (+15 when the layer is flipped and animationHasF is set). Each
// composite is cached per (id, dir) — 8 variants per id, bounding memory.
// Everything here never panics and falls back gracefully to nil so the 3D-cross
// billboards (BuildEntities) can stand in.

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

// npcAnimationArray is the 8-way facing layer-draw-order table, transcribed
// VERBATIM from mudclient.java:7161-7186. Index [dir][k] gives the body-part
// layer drawn k-th (back-to-front) for facing dir 0..7, where each value indexes
// npcSprite[npc][layer]. drawNpc/drawPlayer pick the row by
// l1 = (animationCurrent + (cameraRotation+16)/32) & 7, then iterate all 12
// entries. The W/SW/NW rows (5/6/7) are the mirror images of E/SE/NE (3/2/1)
// flagged for the horizontal flip in spriteClipping.
var npcAnimationArray = [8][12]int{
	{11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
	{11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3, 4},
	{11, 3, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4},
	{3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
	{3, 4, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
	{4, 3, 2, 9, 7, 1, 6, 10, 8, 11, 0, 5},
	{11, 4, 2, 9, 7, 1, 6, 10, 0, 5, 8, 3},
	{11, 2, 9, 7, 1, 6, 10, 0, 5, 8, 4, 3},
}

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

	// parsed config85.jag (GameData) — the item picture table plus the NPC +
	// animation tables.
	itemCount   int
	itemPicture []int // GameData.itemPicture: item id -> flat objects*.dat sprite index

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
	animationHasA    []int          // GameData.animationHasA (1 = has an A-frame)
	animationHasF    []int          // GameData.animationHasF (1 = has the +15 flipped F-frame)
	animByName       map[string]int // lowercased animation name -> id
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

	// --- items (capture itemPicture; skip the rest — only the offsets matter) ---
	itemCount := g.us()
	if g.bad || itemCount < 0 || itemCount > 100000 {
		return false
	}
	for i := 0; i < itemCount; i++ {
		g.gs() // name
	}
	for i := 0; i < itemCount; i++ {
		g.gs() // description
	}
	for i := 0; i < itemCount; i++ {
		g.gs() // command
	}
	// itemPicture (GameData.java:71): the flat sprite index into the objects*.dat
	// run, the one item field we need to draw a ground item's real icon.
	itemPicture := make([]int, itemCount)
	for i := 0; i < itemCount; i++ {
		itemPicture[i] = g.us()
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
	// animationSomething / animationHasA / animationHasF are three ub() loops in
	// THIS order (GameData.java:209-219). We don't use animationSomething, but it
	// MUST be consumed so the substream offset stays aligned (skipping or
	// misordering these corrupts every later config read). animationNumber
	// (GameData.java:218) is not needed (we decode <name>.dat blocks by name, not
	// by a flat sprite-id offset) and the config tail past it is unused, so we
	// stop here.
	animationSomething := make([]int, animCount)
	for i := range animationSomething {
		animationSomething[i] = g.ub()
	}
	animationHasA := make([]int, animCount)
	for i := range animationHasA {
		animationHasA[i] = g.ub()
	}
	animationHasF := make([]int, animCount)
	for i := range animationHasF {
		animationHasF[i] = g.ub()
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

	animByName := make(map[string]int, animCount)
	for i := 0; i < animCount; i++ {
		key := strings.ToLower(animationName[i])
		if _, ok := animByName[key]; !ok {
			animByName[key] = i
		}
	}

	ea.itemCount = itemCount
	ea.itemPicture = itemPicture
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
	ea.animationHasA = animationHasA
	ea.animationHasF = animationHasF
	ea.animByName = animByName
	return true
}

// animFrame is one decoded body-part sprite (a single frame of a <name>.dat
// block), positioned by translateX/translateY inside the shared fullWidth x
// fullHeight figure canvas. pix holds 0x00RRGGBB or -1 for transparent.
type animFrame struct {
	w, h, fullW, fullH, tx, ty int
	pix                        []int
}

// decodeAnimFrame decodes the requested frame of the named animation's
// <name>.dat sprite block, exactly as Surface.loadSprite (Surface.java:371-423)
// reads its frames: ONE shared palette + a run of 6-byte per-frame headers from
// index.dat, with the pixel payload starting at spriteData offset 2 and advancing
// w*h bytes per frame. To reach frame N we walk N+1 headers (advancing the
// index-data cursor each time) while accumulating the payload offset by each
// preceding frame's w*h, then decode the Nth frame's pixels from there. Palette
// index 0 and the 0xff00ff magenta key are transparent; header flag==1 means
// column-major pixel order. Returns nil on any failure (missing entry, malformed
// header, frame out of range). Recovers from panics.
func (ea *entityArchive) decodeAnimFrame(name string, frame int) (f *animFrame) {
	return decodeSpriteFrame(ea.sprites, ea.indexDat, name+".dat", frame)
}

// decodeSpriteFrame decodes one frame of the named .dat sprite block from the
// given sprites archive + its index.dat, using the EXACT Surface.loadSprite
// layout (one shared palette + a run of 6-byte per-frame headers from index.dat,
// pixel payload starting at .dat offset 2 and advancing w*h per frame). It is
// the shared decode used both by entity24.jag body-part frames (decodeAnimFrame)
// and media58.jag item-icon frames (objects*.dat). The full entry name (with the
// ".dat" suffix) is passed so item sprites can name objectsN.dat directly.
// Returns nil on any failure and recovers from panics.
func decodeSpriteFrame(sprites *assets.Archive, idx []byte, entryName string, frame int) (f *animFrame) {
	defer func() {
		if recover() != nil {
			f = nil
		}
	}()
	if frame < 0 || sprites == nil {
		return nil
	}

	spriteData, err := sprites.Get(entryName)
	if err != nil || len(spriteData) < 2 {
		return nil
	}
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

	// Walk the per-frame headers up to and including the requested frame, tracking
	// the cumulative pixel-payload offset. Each header is translateX, translateY
	// (u8), width, height (u16), flag (u8); the payload (spriteOff, starting at 2)
	// advances by w*h after each frame.
	spriteOff := 2
	var tx, ty, w, h, flag int
	for fr := 0; fr <= frame; fr++ {
		if fr > 0 {
			spriteOff += w * h // advance past the previous frame's pixels
		}
		if io+6 > len(idx) {
			return nil
		}
		tx = int(idx[io] & 0xff)
		io++
		ty = int(idx[io] & 0xff)
		io++
		w = entU16(idx, io)
		io += 2
		h = entU16(idx, io)
		io += 2
		flag = int(idx[io] & 0xff)
		io++
	}

	size := w * h
	if size <= 0 || fullW <= 0 || fullH <= 0 {
		return nil
	}
	if spriteOff < 0 || spriteOff+size > len(spriteData) {
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
//
// Flip records that this composite is a W/SW/NW facing built from the mirrored
// E/SE/NE sprites: the canvas pixels are NOT pre-mirrored, so the blit must
// sample columns right-to-left (BlitSpriteScaled's flip arg) to show the correct
// side. This matches drawNpc/drawPlayer passing the flag to spriteClipping.
type CompositeSprite struct {
	W, H   int
	Pix    []int32
	Opaque []bool
	Flip   bool
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
//
// rawColours distinguishes the two authentic colour spaces:
//   - PLAYERS pass palette INDICES (mudclient.java:1133/3007 index
//     characterTopBottomColours[]/characterHairColours[]) -> resolveClothingColour.
//   - NPCs pass RAW 24-bit dye colours (mudclient.java:2144-2151 use
//     npcColour*[id] DIRECTLY, never indexed) -> used as-is. Routing a raw NPC
//     colour through the palette index path mis-mapped any value < len(table)
//     (a small/dark colour, e.g. the black-clothed Man) onto a player palette
//     entry -> the wrong green/yellow kit.
func dyeForLayer(charColour, hair, top, bottom, skin int, rawColours bool) (dye, skinOut int) {
	resolve := func(field int, table []int) int {
		if rawColours {
			return field // NPC: already a raw 24-bit dye colour, never an index
		}
		return resolveClothingColour(field, table)
	}
	switch charColour {
	case 1:
		return resolve(hair, characterHairColours), skin
	case 2:
		return resolve(top, characterTopBottomColours), skin
	case 3:
		return resolve(bottom, characterTopBottomColours), skin
	default:
		return charColour, 0
	}
}

// layerSpec is one resolved body-part layer ready to composite, including the
// per-layer frame index to decode (j2 + the +15 F-frame adjustment).
type layerSpec struct {
	animName   string
	charColour int
	frame      int
}

// composite decodes + recolours + stacks the given layers (already in
// back-to-front draw order) into one CompositeSprite. hair/top/bottom/skin are
// the character's colours; flip marks a mirrored (W/SW/NW) facing and is stored
// on the result for the blit. Returns nil if nothing decoded.
func (ea *entityArchive) composite(layers []layerSpec, hair, top, bottom, skin int, flip, rawColours bool) *CompositeSprite {
	var fullW, fullH int
	type decoded struct {
		f          *animFrame
		dye, skinC int
	}
	var ds []decoded
	for _, l := range layers {
		f := ea.decodeAnimFrame(l.animName, l.frame)
		if f == nil {
			continue
		}
		if f.fullW > fullW {
			fullW = f.fullW
		}
		if f.fullH > fullH {
			fullH = f.fullH
		}
		dye, skinC := dyeForLayer(l.charColour, hair, top, bottom, skin, rawColours)
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
		Flip:   flip,
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

// facingPose maps an 8-way facing dir (l1) to the standing-pose params drawNpc /
// drawPlayer derive at mudclient.java:2101-2112 / 2936-2947: i2 (the 0..4 pose
// column, with 5/6/7 collapsed onto 3/2/1) and whether that pose is horizontally
// mirrored (flag). The standing frame is j2 = i2*3 (stepCount 0 -> npcWalkModel
// [0] == 0). dir is taken mod 8 so callers can pass a raw heading sum.
func facingPose(dir int) (i2 int, flip bool) {
	dir &= 7
	i2 = dir
	switch dir {
	case 5:
		i2, flip = 3, true
	case 6:
		i2, flip = 2, true
	case 7:
		i2, flip = 1, true
	}
	return i2, flip
}

// npcLayers builds the back-to-front layer specs for an NPC facing dir (0..7).
// The layer draw order is npcAnimationArray[dir]; each non-empty layer's sprite
// id is mapped to its animation name + charColour, and its frame is j2 = i2*3
// plus the +15 F-frame adjustment when the pose is flipped (i2 in 1..3) and the
// layer's animation has an F-frame (drawNpc:2126-2134).
func (ea *entityArchive) npcLayers(npcID, dir, stepFrame int) []layerSpec {
	if npcID < 0 || npcID >= ea.npcCount {
		return nil
	}
	i2, flip := facingPose(dir)
	// j2 = i2*3 is the pose's first frame; the 3 consecutive frames are the
	// walk cycle, selected by stepFrame = npcWalkModel[phase] ∈ {0,1,2} (drawNpc
	// mudclient.java:2118 j2 = i2*3 + npcWalkModel[...]). stepFrame 0 == standing.
	j2 := i2*3 + stepFrame
	var layers []layerSpec
	for _, layer := range npcAnimationArray[dir&7] {
		spriteID := ea.npcSprite[npcID][layer]
		if spriteID < 0 || spriteID >= ea.animCount {
			continue
		}
		frame := j2
		if flip && i2 >= 1 && i2 <= 3 && ea.animationHasF[spriteID] == 1 {
			frame += 15
		}
		layers = append(layers, layerSpec{
			animName:   ea.animationName[spriteID],
			charColour: ea.animationCharCol[spriteID],
			frame:      frame,
		})
	}
	return layers
}

// npcCompositeKey identifies a cached composite by NPC id + 8-way facing dir, so
// each id holds at most 8 variants (bounding memory).
type npcCompositeKey struct {
	id   int
	dir  int
	step int // walk-cycle frame offset (npcWalkModel value); 0 = standing
}

var (
	npcCompositeMu    sync.Mutex
	npcCompositeCache = map[npcCompositeKey]*CompositeSprite{}
	npcCompositeMiss  = map[npcCompositeKey]bool{} // (id,dir) that failed (don't retry)
)

// compositeNPC returns the cached standing-frame billboard for an NPC id facing
// dir (0..7), or nil if the archives are unavailable or the NPC has no valid
// sprite layers for that facing (caller then falls back to the 3D-cross
// billboard). Memoised per (id, dir).
func compositeNPC(npcID, dir, stepFrame int) *CompositeSprite {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return nil
	}
	dir &= 7
	key := npcCompositeKey{npcID, dir, stepFrame}
	npcCompositeMu.Lock()
	defer npcCompositeMu.Unlock()
	if cs, ok := npcCompositeCache[key]; ok {
		return cs
	}
	if npcCompositeMiss[key] {
		return nil
	}
	layers := entityArc.npcLayers(npcID, dir, stepFrame)
	if len(layers) == 0 {
		npcCompositeMiss[key] = true
		return nil
	}
	_, flip := facingPose(dir)
	cs := entityArc.composite(
		layers,
		entityArc.npcColourHair[npcID],
		entityArc.npcColourTop[npcID],
		entityArc.npcColourBtm[npcID],
		entityArc.npcColourSkin[npcID],
		flip,
		true, // NPC colours are RAW 24-bit dye values (mudclient drawNpc :2144-2151)
	)
	if cs == nil {
		npcCompositeMiss[key] = true
		return nil
	}
	npcCompositeCache[key] = cs
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

// playerStandKey identifies a default-human composite by facing dir + walk-cycle
// frame offset (0 = standing).
type playerStandKey struct{ dir, step int }

var (
	playerCompositeMu    sync.Mutex
	playerCompositeCache = map[playerStandKey]*CompositeSprite{}
	playerCompositeMiss  = map[playerStandKey]bool{}
)

// playerLayer is one fixed-appearance body part for the local player: the
// animation name + its charColour marker. They are listed back-to-front (legs ->
// body -> head), matching the appearance preview (mudclient.java:1133-1141).
var playerLayers = []struct {
	name       string
	charColour int
}{
	{"legs1", 3},
	{"body1", 2},
	{"head1", 1},
}

// compositePlayer returns the cached default-human billboard for the local
// player facing dir (0..7). Returns nil if the archives are unavailable (caller
// falls back to the 3D-cross billboard). The layers are legs1/body1/head1
// (charColour markers 3/2/1 resolve to bottom/top/hair); skin is direct. The
// per-layer frame is the facing standing frame j2 = i2*3, plus the +15 F-frame
// when the pose is flipped (W/SW/NW) and the layer's animation has an F-frame —
// so the player shows the correct side as the camera pans. Memoised per dir.
func compositePlayer(dir, stepFrame int) *CompositeSprite {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return nil
	}
	dir &= 7
	key := playerStandKey{dir, stepFrame}
	playerCompositeMu.Lock()
	defer playerCompositeMu.Unlock()
	if cs, ok := playerCompositeCache[key]; ok {
		return cs
	}
	if playerCompositeMiss[key] {
		return nil
	}
	i2, flip := facingPose(dir)
	j2 := i2*3 + stepFrame // walk-cycle frame (stepFrame 0 = standing)
	var layers []layerSpec
	for _, pl := range playerLayers {
		frame := j2
		if flip && i2 >= 1 && i2 <= 3 {
			if id, ok := entityArc.animByName[pl.name]; ok && entityArc.animationHasF[id] == 1 {
				frame += 15
			}
		}
		layers = append(layers, layerSpec{animName: pl.name, charColour: pl.charColour, frame: frame})
	}
	cs := entityArc.composite(
		layers,
		playerHairColIdx,   // hair index
		playerTopColIdx,    // top index
		playerBottomColIdx, // bottom index
		playerSkinColour,   // skin direct
		flip,
		false, // player colours are palette INDICES (resolveClothingColour)
	)
	if cs == nil {
		playerCompositeMiss[key] = true
		return nil
	}
	playerCompositeCache[key] = cs
	return cs
}

// playerAppearanceKey identifies a real-appearance composite by the full
// worn-equipment sprite array + the four colour indices + the 8-way facing
// dir. Two players with the same outfit + colours + facing share a cached
// composite (bernard + a drone in the same kit collapse to one entry).
type playerAppearanceKey struct {
	equip                  [12]int
	hair, top, trouser, sk int
	dir                    int
	step                   int // walk-cycle frame offset; 0 = standing
}

var (
	playerAppMu    sync.Mutex
	playerAppCache = map[playerAppearanceKey]*CompositeSprite{}
	playerAppMiss  = map[playerAppearanceKey]bool{}
)

// compositePlayerAppearance returns the cached standing-frame billboard for a
// player wearing the given per-slot equipment sprites, dyed by the four
// appearance colour indices (hair/top/trouser/skin), facing dir (0..7). It
// uses the SAME per-direction machinery as compositePlayer(dir): the layer
// draw order is npcAnimationArray[dir], and within each layer the worn
// animation id is equip[layer]-1 (mudclient.java:2963 equippedItem[layer]-1),
// recoloured per the layer's animationCharCol marker (1=hair, 2=top, 3=bottom)
// — exactly the drawPlayer loop. Layers with equip[layer]==0 (nothing worn /
// no sprite) are skipped, so an empty outfit yields nil and the caller falls
// back to the default-human compositePlayer. NEVER panics; returns nil on any
// failure (archives missing, no decodable layer). Memoised per
// (equip + colours + dir).
func compositePlayerAppearance(equip [12]int, hair, top, trouser, skin, dir, stepFrame int) (cs *CompositeSprite) {
	defer func() {
		if recover() != nil {
			cs = nil
		}
	}()
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		return nil
	}
	dir &= 7
	key := playerAppearanceKey{equip: equip, hair: hair, top: top, trouser: trouser, sk: skin, dir: dir, step: stepFrame}

	playerAppMu.Lock()
	defer playerAppMu.Unlock()
	if c, ok := playerAppCache[key]; ok {
		return c
	}
	if playerAppMiss[key] {
		return nil
	}

	i2, flip := facingPose(dir)
	j2 := i2*3 + stepFrame // walk-cycle frame (stepFrame 0 = standing)
	var layers []layerSpec
	for _, layer := range npcAnimationArray[dir&7] {
		if layer < 0 || layer >= len(equip) {
			continue
		}
		animID := equip[layer] - 1 // equippedItem[layer] - 1 (mudclient.java:2963)
		if animID < 0 || animID >= entityArc.animCount {
			continue // nothing worn in that layer / out-of-range sprite
		}
		frame := j2
		if flip && i2 >= 1 && i2 <= 3 && entityArc.animationHasF[animID] == 1 {
			frame += 15
		}
		layers = append(layers, layerSpec{
			animName:   entityArc.animationName[animID],
			charColour: entityArc.animationCharCol[animID],
			frame:      frame,
		})
	}
	if len(layers) == 0 {
		playerAppMiss[key] = true
		return nil
	}
	// hair/top/trouser are clothing-table INDICES (resolveClothingColour maps
	// them); skin is likewise an index into characterSkinColours. The composite
	// machinery resolves the clothing dye from the index per layer; pass the
	// resolved skin colour directly so the skin-recolour rule fires.
	skinColour := skin
	if skin >= 0 && skin < len(characterSkinColours) {
		skinColour = characterSkinColours[skin]
	}
	cs = entityArc.composite(layers, hair, top, trouser, skinColour, flip, false) // player colours are palette indices
	if cs == nil {
		playerAppMiss[key] = true
		return nil
	}
	playerAppCache[key] = cs
	return cs
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
