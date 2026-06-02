package render

import (
	"sync"

	"github.com/gen0cide/westworld/facts"
)

// RSC characters (players + NPCs) are 2D billboard sprites, NOT 3D models. Each
// is composited from up to 12 body-part "animation" sprites (head/body/legs/…),
// stacked into a shared figure canvas via per-sprite XShift/YShift and recoloured
// per the dye/skin rule (Surface.transparentSpritePlot). Every sprite is read
// from OpenRSC's Authentic_Sprites.orsc by numeric id: the frame for animation
// `animID` at frame `frame` is sprite (animationNumbers()[animID] + frame), the
// scheme mudclient.loadEntitiesAuthentic assigns (15 base + 3 A + 9 F per 27-slot
// stride). The layer set + colours come from facts.NpcDef (OpenRSC NpcDefs.json,
// keyed by the same npc id the server sends) and the generated authenticAnimDefs
// table (EntityHandler order). NO eggsampler / .jag / config85 data is used.
//
// We render the STANDING pose for an 8-way facing dir (0..7) computed from the
// entity heading + camera rotation. drawNpc maps dir 5/6/7 -> i2 3/2/1 with a
// horizontal flip (W/SW/NW are mirror images of E/SE/NE), picks the per-direction
// back-to-front layer order npcAnimationArray[dir], and within each animation
// draws standing frame j2 = i2*3 (+15 when the layer is flipped and the animation
// has an F-frame). Each composite is cached per (id, dir, step). Everything here
// never panics and falls back to nil so the 3D-cross billboards can stand in.

// npcAnimationArray is the 8-way facing layer-draw-order table, transcribed
// VERBATIM from mudclient.java:7161-7186. Index [dir][k] gives the body-part
// layer drawn k-th (back-to-front) for facing dir 0..7, where each value indexes
// NpcDef.Sprites[layer] / equip[layer]. The W/SW/NW rows (5/6/7) are the mirror
// images of E/SE/NE (3/2/1) flagged for the horizontal flip in spriteClipping.
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
// the authentic palettes (mudclient.java:7228 / 7246 / 7299). A PLAYER's clothing
// colour field holds an INDEX into one of these tables; an NPC's colour field is
// a direct 24-bit colour. The charColour marker on each animation (1=hair, 2=top,
// 3=bottom) selects which clothing table/index drives that layer's dye.
var characterSkinColours = []int{0xecded0, 0xccb366, 0xb38c40, 0x997326, 0x906020}

var characterTopBottomColours = []int{
	0xff0000, 0xff8000, 0xffe000, 0xa0e000, 57344, 32768, 41088, 45311, 33023, 12528,
	0xe000e0, 0x303030, 0x604000, 0x805000, 0xffffff,
}

var characterHairColours = []int{
	0xffc030, 0xffa040, 0x805030, 0x604020, 0x303030, 0xff6020, 0xff4000, 0xffffff, 65280, 65535,
}

// entityAnimNum returns the per-animation sprite-block base table (animationNumbers),
// computed once.
var (
	animNumOnce sync.Once
	animNum     []int
)

func entityAnimNum() []int {
	animNumOnce.Do(func() { animNum = animationNumbers() })
	return animNum
}

// animFrame is one decoded body-part sprite, positioned by tx/ty (XShift/YShift)
// inside the shared fullW x fullH figure canvas (Something1/Something2). pix holds
// 0x00RRGGBB or -1 for transparent.
type animFrame struct {
	w, h, fullW, fullH, tx, ty int
	pix                        []int
}

// decodeEntitySprite reads the entity body-part sprite with the given absolute id
// (animationNumbers()[animID] + frame) from Authentic_Sprites.orsc and converts it
// to an animFrame. Transparency keys on BLACK (0x000000), the 2D-sprite key. The
// figure canvas is Something1 x Something2 (e.g. 64x102 for a human); XShift/YShift
// place the part within it. Returns nil on any failure. Never panics.
func decodeEntitySprite(spriteID int) (f *animFrame) {
	defer func() {
		if recover() != nil {
			f = nil
		}
	}()
	sa := sprites()
	if sa == nil {
		return nil
	}
	sp, err := sa.Sprite(spriteID)
	if err != nil || sp == nil || sp.Width <= 0 || sp.Height <= 0 {
		return nil
	}
	fullW, fullH := sp.Something1, sp.Something2
	if fullW <= 0 {
		fullW = sp.Width
	}
	if fullH <= 0 {
		fullH = sp.Height
	}
	tx, ty := 0, 0
	if sp.RequiresShift {
		tx, ty = sp.XShift, sp.YShift
	}
	pix := make([]int, sp.Width*sp.Height)
	for i, p := range sp.Pixels {
		c := int(p) & 0xffffff
		if c == 0 {
			pix[i] = -1 // BLACK = transparency key
		} else {
			pix[i] = c
		}
	}
	return &animFrame{w: sp.Width, h: sp.Height, fullW: fullW, fullH: fullH, tx: tx, ty: ty, pix: pix}
}

// recolourTexel applies the Surface.transparentSpritePlot dye/skin rule
// (Surface.java:1677-1682): a grey palette colour (R==G==B) is multiplied by the
// dye (clothing) colour; an R==255 && G==B colour is multiplied by the skin
// colour; everything else is used unchanged. A zero dye/skin means "no recolour".
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
// blits it depth-scaled. Flip records a W/SW/NW facing built from the mirrored
// E/SE/NE sprites (the blit samples columns right-to-left).
type CompositeSprite struct {
	W, H   int
	Pix    []int32
	Opaque []bool
	Flip   bool
}

// resolveClothingColour maps a PLAYER clothing colour index (hair/top/bottom) to
// a 24-bit RGB. Small values index the authentic table; out-of-range values are
// used as a direct 24-bit colour.
func resolveClothingColour(field int, table []int) int {
	if field >= 0 && field < len(table) {
		return table[field]
	}
	return field
}

// dyeForLayer resolves the dye + skin colour for one layer given its animation's
// charColour marker and the character's clothing/skin colours. Marker 1 = hair,
// 2 = top, 3 = bottom (each with skin); any other marker is a literal colour used
// directly with no skin recolour. rawColours distinguishes PLAYERS (palette
// indices -> resolveClothingColour) from NPCs (raw 24-bit dye colours, used as-is).
func dyeForLayer(charColour, hair, top, bottom, skin int, rawColours bool) (dye, skinOut int) {
	resolve := func(field int, table []int) int {
		if rawColours {
			return field
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

// layerSpec is one resolved body-part layer ready to composite: the animation id
// (into authenticAnimDefs / animationNumbers) + the frame within its 27-slot block.
type layerSpec struct {
	animID     int
	charColour int
	frame      int
}

// composite decodes + recolours + stacks the given layers (already back-to-front)
// into one CompositeSprite. hair/top/bottom/skin are the character's colours; flip
// marks a mirrored facing. Returns nil if nothing decoded.
func composite(layers []layerSpec, hair, top, bottom, skin int, flip, rawColours bool) *CompositeSprite {
	nums := entityAnimNum()
	var fullW, fullH int
	type decoded struct {
		f          *animFrame
		dye, skinC int
	}
	var ds []decoded
	for _, l := range layers {
		if l.animID < 0 || l.animID >= len(nums) {
			continue
		}
		f := decodeEntitySprite(nums[l.animID] + l.frame)
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

// facingPose maps an 8-way facing dir to the standing-pose column i2 (0..4, with
// 5/6/7 collapsed onto 3/2/1) and whether that pose is horizontally mirrored.
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

// layerFrame computes the per-layer frame within an animation's 27-slot block:
// j2 = i2*3 + step, plus the +15 F-frame offset when the pose is flipped (i2 in
// 1..3) and the animation has an F-frame.
func layerFrame(animID, i2, step int, flip bool) int {
	frame := i2*3 + step
	if flip && i2 >= 1 && i2 <= 3 && animID >= 0 && animID < len(authenticAnimDefs) && authenticAnimDefs[animID].hasF {
		frame += 15
	}
	return frame
}

// npcCompositeKey identifies a cached composite by NPC id + facing dir + walk step.
type npcCompositeKey struct {
	id, dir, step int
}

var (
	npcCompositeMu    sync.Mutex
	npcCompositeCache = map[npcCompositeKey]*CompositeSprite{}
	npcCompositeMiss  = map[npcCompositeKey]bool{}
)

// compositeNPC returns the cached standing-frame billboard for an OpenRSC npc id
// facing dir (0..7), or nil if the sprites/defs are unavailable or the NPC has no
// valid layers (caller falls back to the 3D-cross billboard). Memoised per
// (id, dir, step). NPC colours are RAW 24-bit dye values (NpcDef.*Colour).
func compositeNPC(f *facts.Facts, npcID, dir, stepFrame int) *CompositeSprite {
	if f == nil {
		return nil
	}
	def := f.NpcDefs[npcID]
	if def == nil {
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
	i2, flip := facingPose(dir)
	var layers []layerSpec
	for _, layer := range npcAnimationArray[dir] {
		animID := def.Sprites[layer]
		if animID < 0 || animID >= len(authenticAnimDefs) {
			continue
		}
		layers = append(layers, layerSpec{
			animID:     animID,
			charColour: authenticAnimDefs[animID].charColour,
			frame:      layerFrame(animID, i2, stepFrame, flip),
		})
	}
	cs := composite(layers, def.HairColour, def.TopColour, def.BottomColour, def.SkinColour, flip, true)
	if cs == nil {
		npcCompositeMiss[key] = true
		return nil
	}
	npcCompositeCache[key] = cs
	return cs
}

// npcBillboardSize returns the world-space billboard width/height for an NPC
// (NpcDef.Camera1/Camera2), defaulting to the human 145x220 when unknown.
func npcBillboardSize(f *facts.Facts, npcID int) (w, h int) {
	if f != nil {
		if def := f.NpcDefs[npcID]; def != nil {
			w, h = def.Camera1, def.Camera2
		}
	}
	if w <= 0 {
		w = playerBillboardW
	}
	if h <= 0 {
		h = playerBillboardH
	}
	return w, h
}

// Default local-player appearance — legs1 + body1 + head1 (animIDs 2/1/0), default
// clothing colours. Billboard is a fixed 145x220.
const (
	playerBillboardW = 145
	playerBillboardH = 220

	playerSkinColour   = 0xecded0 // characterSkinColours[0]
	playerHairColIdx   = 2        // characterHairColours[2]
	playerTopColIdx    = 8        // characterTopBottomColours[8] = blue
	playerBottomColIdx = 14       // characterTopBottomColours[14] = white
)

// playerLayers is the default-human body parts back-to-front (legs->body->head),
// by authentic animation id: legs1=2, body1=1, head1=0.
var playerLayers = []int{2, 1, 0}

type playerStandKey struct{ dir, step int }

var (
	playerCompositeMu    sync.Mutex
	playerCompositeCache = map[playerStandKey]*CompositeSprite{}
	playerCompositeMiss  = map[playerStandKey]bool{}
)

// compositePlayer returns the cached default-human billboard for the local player
// facing dir (0..7). Player colours are palette INDICES (resolveClothingColour).
func compositePlayer(dir, stepFrame int) *CompositeSprite {
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
	var layers []layerSpec
	for _, animID := range playerLayers {
		layers = append(layers, layerSpec{
			animID:     animID,
			charColour: authenticAnimDefs[animID].charColour,
			frame:      layerFrame(animID, i2, stepFrame, flip),
		})
	}
	cs := composite(layers, playerHairColIdx, playerTopColIdx, playerBottomColIdx, playerSkinColour, flip, false)
	if cs == nil {
		playerCompositeMiss[key] = true
		return nil
	}
	playerCompositeCache[key] = cs
	return cs
}

// playerAppearanceKey identifies a real-appearance composite by worn-equipment +
// colours + facing + walk step.
type playerAppearanceKey struct {
	equip                  [12]int
	hair, top, trouser, sk int
	dir, step              int
}

var (
	playerAppMu    sync.Mutex
	playerAppCache = map[playerAppearanceKey]*CompositeSprite{}
	playerAppMiss  = map[playerAppearanceKey]bool{}
)

// compositePlayerAppearance returns the cached standing billboard for a player
// wearing the given per-slot equipment sprites (equip[layer]-1 is the animation
// id, mudclient.java:2963), dyed by the four colour INDICES, facing dir. Layers
// with equip[layer]==0 are skipped; an empty outfit yields nil (caller falls back
// to compositePlayer). Memoised. Never panics.
func compositePlayerAppearance(equip [12]int, hair, top, trouser, skin, dir, stepFrame int) (cs *CompositeSprite) {
	defer func() {
		if recover() != nil {
			cs = nil
		}
	}()
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
	var layers []layerSpec
	for _, layer := range npcAnimationArray[dir] {
		if layer < 0 || layer >= len(equip) {
			continue
		}
		animID := equip[layer] - 1
		if animID < 0 || animID >= len(authenticAnimDefs) {
			continue
		}
		layers = append(layers, layerSpec{
			animID:     animID,
			charColour: authenticAnimDefs[animID].charColour,
			frame:      layerFrame(animID, i2, stepFrame, flip),
		})
	}
	if len(layers) == 0 {
		playerAppMiss[key] = true
		return nil
	}
	// hair/top/trouser are clothing-table indices; skin is an index into
	// characterSkinColours -> resolve to a direct colour for the skin-recolour rule.
	skinColour := skin
	if skin >= 0 && skin < len(characterSkinColours) {
		skinColour = characterSkinColours[skin]
	}
	cs = composite(layers, hair, top, trouser, skinColour, flip, false)
	if cs == nil {
		playerAppMiss[key] = true
		return nil
	}
	playerAppCache[key] = cs
	return cs
}
