package render

import "github.com/gen0cide/westworld/facts"

// Exported wrappers so the render/orsc faithful-port (and other callers) can
// reuse this package's verified entity-sprite compositing to build the pixels
// for their own billboard layer, without re-porting the body-part stack +
// recolour. The returned CompositeSprite is the W×H billboard canvas (Pix
// 0x00RRGGBB row-major, Opaque mask, Flip = mirrored W/SW/NW facing).

// CompositePlayerSprite returns the default-human player billboard for an 8-way
// facing dir (0..7) and walk-cycle step (0 = standing), or nil if the sprite
// archive is unavailable.
func CompositePlayerSprite(dir, step int) *CompositeSprite { return compositePlayer(dir, step) }

// CompositeNPCSprite returns the billboard for an OpenRSC npc id facing dir, or
// nil if the def/sprites are unavailable.
func CompositeNPCSprite(f *facts.Facts, npcID, dir, step int) *CompositeSprite {
	return compositeNPC(f, npcID, dir, step)
}

// NPCBillboardSize returns an NPC's world-space billboard width/height
// (NpcDef.Camera1/Camera2), defaulting to the human 145×220.
func NPCBillboardSize(f *facts.Facts, npcID int) (w, h int) { return npcBillboardSize(f, npcID) }

// CompositePlayerAppearanceSprite returns the billboard for a player wearing the
// given per-slot equipment sprites, dyed by the colour indices, facing dir, or
// nil if the outfit is empty/undecodable (caller falls back to the default human).
func CompositePlayerAppearanceSprite(equip [12]int, hair, top, trouser, skin, dir, step int) *CompositeSprite {
	return compositePlayerAppearance(equip, hair, top, trouser, skin, dir, step)
}

// CompositeItemSprite returns the cached inventory-icon billboard for a ground
// item id (the item's sprite composited onto the 48×32 canvas), or nil when the
// sprite archive is unavailable. The icon index + recolour masks come from
// itemIcons (render/itempicture_data.go); f is retained for signature/back-compat. Used by
// the orsc picker's ground-item billboard hit-test so its AABB is derived from
// the SAME icon the renderer drew.
func CompositeItemSprite(f *facts.Facts, itemID int) *CompositeSprite {
	return compositeItem(f, itemID)
}
