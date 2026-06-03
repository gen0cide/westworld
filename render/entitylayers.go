package render

import "github.com/gen0cide/westworld/facts"

// entitylayers.go — the RAW per-layer entity contract for the orsc on-screen 1:1
// blit (NPC/player 2D-sprite parity, Phase 4 / Milestone C,
// docs/build/NPC_SPRITE_PARITY_PLAN.md tool #2).
//
// The pre-compositing CompositeSprite path (composite() / CompositeNPCSprite) flattens
// every body-part layer into ONE figure canvas at INTEGER trim offsets, then the orsc
// drawEntity whole-canvas integer-NN scales it — which loses each layer's sub-pixel
// trim (spriteTranslate) and the per-row skew. The authentic rev-235 client instead
// blits each body-part layer SEPARATELY through Surface.spriteClipping (10-arg), which
// applies the layer's trim in 16.16 fixed-point AT BLIT TIME with fractional carry, plus
// the per-row skew + flip + the in-blit dye/skin recolour. To match it byte-for-byte the
// orsc engine must register the RAW layers (NOT the composited canvas) and run the same
// 16.16 blit per layer.
//
// EntityLayers returns exactly those raw layers (decoded, back-to-front, UN-recoloured —
// the recolour happens in the blit), each carrying its trimmed pixels + the full-canvas
// trim (fullW/fullH/tx/ty) the 16.16 spriteClipping needs, plus the layer's dye/skin.

// EntityLayer is ONE decoded body-part sprite ready for the faithful 16.16 blit: the
// trimmed pixel grid (W×H, 0x00RRGGBB, transparency keyed on Transparent==-1 sentinel
// in Pix — the per-sprite palette-index-0 key), the full-canvas trim (FullW/FullH the
// layer would occupy un-trimmed, with the trimmed grid placed at offset TX/TY), and the
// layer's recolour colours (Dye applied to grey r==g==b texels, Skin to r==255&&g==b).
// This mirrors a single Surface sprite slot's {spriteWidth/Height, spriteWidthFull/
// HeightFull, spriteTranslateX/Y, pixels} the 10-arg spriteClipping consumes.
type EntityLayer struct {
	Pix              []int32 // W*H row-major; -1 = transparent (palette idx 0 key)
	W, H             int     // trimmed sprite dims (spriteWidth/spriteHeight)
	FullW, FullH     int     // full-canvas dims (spriteWidthFull/spriteHeightFull)
	TX, TY           int     // trim offset of the trimmed grid in the full canvas
	Dye, Skin        int     // recolour colours (0 => 0xffffff identity), applied in-blit
}

// EntitySprite is the raw-layer billboard for the faithful on-screen blit: the layers
// back-to-front + the mirror flag. Flip marks a W/SW/NW facing (the blit samples columns
// right-to-left). FullW/FullH are the SHARED figure-canvas dims (the max over layers) —
// the projected billboard rect is sized from these, EXACTLY as the DEOB rect-replicator
// sizes from spriteWidthFull/HeightFull.
type EntitySprite struct {
	Layers []EntityLayer
	Flip   bool
}

// decodeLayers runs the shared decode/colour-resolve loop (the front half of composite())
// but stops BEFORE flattening: it returns each layer's raw trimmed pixels + trim + dye/skin
// in back-to-front order. Returns nil if nothing decoded. Never panics (decodeEntitySprite
// recovers internally).
func decodeLayers(layers []layerSpec, hair, top, bottom, skin int, flip, rawColours bool) *EntitySprite {
	nums := entityAnimNum()
	var out []EntityLayer
	for _, l := range layers {
		if l.animID < 0 || l.animID >= len(nums) {
			continue
		}
		f := decodeEntitySprite(nums[l.animID] + l.frame)
		if f == nil {
			continue
		}
		dye, skinC := dyeForLayer(l.charColour, hair, top, bottom, skin, rawColours)
		// Copy the decoded pixels into an int32 slice (decodeEntitySprite returns []int
		// with -1 transparency; keep that sentinel — the blit skips it like the
		// per-sprite palette-index-0 key). FullW/FullH default to the trimmed dims when
		// the source carries no full-canvas size.
		pix := make([]int32, len(f.pix))
		for i, c := range f.pix {
			pix[i] = int32(c)
		}
		fullW, fullH := f.fullW, f.fullH
		if fullW <= 0 {
			fullW = f.w
		}
		if fullH <= 0 {
			fullH = f.h
		}
		out = append(out, EntityLayer{
			Pix: pix, W: f.w, H: f.h,
			FullW: fullW, FullH: fullH, TX: f.tx, TY: f.ty,
			Dye: dye, Skin: skinC,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return &EntitySprite{Layers: out, Flip: flip}
}

// NPCEntityLayers returns the raw back-to-front body-part layers for an OpenRSC npc id
// facing dir (0..7) at walk step, or nil when the def/sprites are unavailable. NPC colours
// are RAW 24-bit dye values (NpcDef.*Colour). Mirrors compositeNPC's layer/colour
// resolution exactly (same npcAnimationArray order, same facingPose/layerFrame, same
// rawColours=true), only WITHOUT the canvas flatten.
func NPCEntityLayers(f *facts.Facts, npcID, dir, step int) *EntitySprite {
	if f == nil {
		return nil
	}
	def := f.NpcDefs[npcID]
	if def == nil {
		return nil
	}
	dir &= 7
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
			frame:      layerFrame(animID, i2, step, flip),
		})
	}
	return decodeLayers(layers, def.HairColour, def.TopColour, def.BottomColour, def.SkinColour, flip, true)
}

// ItemEntityLayers returns a SINGLE raw item-picture layer for a ground item id, or nil
// when the item has no in-scope authentic picture / content8 is unavailable. This is the
// RAW 16.16 path for ground items (Task B2, critique #1: a 1-layer EntitySprite registered
// via Scene.AddEntityLayers, NOT the 48x32 compositeItem/AddEntity pre-composite).
//
// The single layer carries the trimmed item-icon pixels decoded from content8 "2d graphics"
// (the SAME bytes the DEOB/JAR B1 legs decode), the full 48x32 item canvas (FullW/FullH), the
// sprite's trim offset (TX/TY = spriteTranslateX/Y), and Dye = pictureMask / Skin = 0. The
// recolour then happens IN the blit via transparentSpritePlot exactly like the rat
// (critique #2: NOT recolorItemPixel — the authentic ground draw is grey-tint(colour1)+
// red-tint(colour2=identity) only, no blueMask branch). Flip is always false (the on-screen
// ground item is never mirrored). Never panics.
func ItemEntityLayers(itemID int) (es *EntitySprite) {
	defer func() {
		if recover() != nil {
			es = nil
		}
	}()
	icon, ok := itemIcons[itemID]
	if !ok {
		return nil
	}
	f := content8ItemFrame(icon.pic)
	if f == nil {
		return nil
	}
	pix := make([]int32, len(f.pix))
	for i, c := range f.pix {
		pix[i] = int32(c)
	}
	fullW, fullH := f.fullW, f.fullH
	if fullW <= 0 {
		fullW = f.w
	}
	if fullH <= 0 {
		fullH = f.h
	}
	return &EntitySprite{
		Layers: []EntityLayer{{
			Pix: pix, W: f.w, H: f.h,
			FullW: fullW, FullH: fullH, TX: f.tx, TY: f.ty,
			// Dye = pictureMask (grey-tint colour1); Skin = 0 -> colour2 becomes 0xffffff
			// identity in the blit, so transparentSpritePlot1 (single-tint) is the path —
			// EXACTLY the authentic ground draw (NO blue branch). For item 14 mask 0 -> dye
			// 0xffffff identity (pure decode+project+blit).
			Dye: icon.mask, Skin: 0,
		}},
		Flip: false,
	}
}

// PlayerEntityLayers returns the raw back-to-front body-part layers for the default-human
// player facing dir (0..7) at walk step, or nil when the sprites are unavailable. Player
// colours are palette INDICES (resolveClothingColour). Mirrors compositePlayer.
func PlayerEntityLayers(dir, step int) *EntitySprite {
	dir &= 7
	i2, flip := facingPose(dir)
	var layers []layerSpec
	for _, animID := range playerLayers {
		layers = append(layers, layerSpec{
			animID:     animID,
			charColour: authenticAnimDefs[animID].charColour,
			frame:      layerFrame(animID, i2, step, flip),
		})
	}
	return decodeLayers(layers, playerHairColIdx, playerTopColIdx, playerBottomColIdx, playerSkinColour, flip, false)
}

// PlayerAppearanceEntityLayers returns the raw back-to-front body-part layers for a player
// wearing the given per-slot equipment (equip[layer]-1 is the animation id), dyed by the
// four colour indices, facing dir at walk step, or nil when the outfit is empty/undecodable.
// Mirrors compositePlayerAppearance.
func PlayerAppearanceEntityLayers(equip [12]int, hair, top, trouser, skin, dir, step int) (es *EntitySprite) {
	defer func() {
		if recover() != nil {
			es = nil
		}
	}()
	dir &= 7
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
			frame:      layerFrame(animID, i2, step, flip),
		})
	}
	if len(layers) == 0 {
		return nil
	}
	skinColour := skin
	if skin >= 0 && skin < len(characterSkinColours) {
		skinColour = characterSkinColours[skin]
	}
	return decodeLayers(layers, hair, top, trouser, skinColour, flip, false)
}
