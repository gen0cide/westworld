package v235

import (
	"fmt"

	"github.com/gen0cide/westworld/event"
)

// DecodeNpcCoords parses opcode 79 (InNpcCoords): bitpacked per-tick
// NPC movement/spawn updates within view range.
//
// Ported from GameStateUpdater.updateNpcs (GameStateUpdater.java:148-291).
//
// Layout (v235 client, non-retro):
//
//	[8 bits]  localNpcCount (number of already-tracked NPCs to update)
//
//	for each local NPC (in same order the server has them):
//	    [1 bit]   needsUpdate
//	    if needsUpdate:
//	        [1 bit] discriminator (0 = movement, 1 = not-moving)
//	        if 0 (movement):
//	            [3 bits] sprite (direction 0-7)
//	        if 1 (not moving):
//	            Peek next 2 bits:
//	              - if == 3 (binary 11): REMOVE_NPC (consume 2 bits)
//	              - else: this is the high 2 bits of a 4-bit sprite;
//	                consume 2 more for the low 2 bits (sprite_changed)
//
//	then new NPCs entering view (until packet ends):
//	    [12 bits] npc_index
//	    [5 bits]  offset_x (signed; mod-64 wrapped)
//	    [5 bits]  offset_y (signed; mod-64 wrapped)
//	    [4 bits]  sprite
//	    [10 bits] npc_type_id (joins to NpcDef.ID)
//
// `order` is the client's mirror of the server's per-player localNpcs
// list (same iteration order). The existing-NPC update records are
// positional: record i refers to the NPC at order[i]. We use it to emit
// movement (re-stamped position) and — critically — REMOVAL events so a
// despawned/dead NPC is pruned from the world mirror. Pass the result of
// world.Npcs.Order(); a short/empty slice degrades gracefully (we still
// consume the bits correctly so the new-NPCs section parses).
//
// Returns events for movement, removal, and new NPCs entering view.
// Sprite-only changes are consumed but suppressed (display noise).
//
// Also returns localCount (the number of already-tracked NPCs the server
// expects us to update positionally). The caller compares it to its own
// order-list length: any mismatch means our slot->index mirror has
// desynced from the server's localNpcs, after which every positional
// update is misattributed — a Tier-1 anomaly worth logging loudly.
func DecodeNpcCoords(payload []byte, ownX, ownY int, order []int) ([]event.Event, int, error) {
	b := WrapBuffer(payload)
	b.StartBitAccess()

	localCount, err := b.ReadBits(8)
	if err != nil {
		return nil, 0, fmt.Errorf("npccoords local_count: %w", err)
	}

	var events []event.Event

	// indexAt maps a positional slot to the NPC index our mirror has at
	// that slot. Out-of-range (our order desynced or is empty on first
	// packet) yields -1 — we still parse the bits, just can't attribute
	// the update to an index.
	indexAt := func(slot uint32) int {
		if int(slot) < len(order) {
			return order[slot]
		}
		return -1
	}

	// Existing-NPC update records, in server localNpcs order.
	for i := uint32(0); i < localCount; i++ {
		needsUpdate, err := b.ReadBits(1)
		if err != nil {
			return events, int(localCount), nil
		}
		if needsUpdate == 0 {
			continue
		}
		discrim, err := b.ReadBits(1)
		if err != nil {
			return events, int(localCount), nil
		}
		if discrim == 0 {
			// Movement: 3 bits direction (0-7, RSC compass order).
			dir, err := b.ReadBits(3)
			if err != nil {
				return events, int(localCount), nil
			}
			idx := indexAt(i)
			if idx >= 0 {
				// RELATIVE move: the NPC stepped one tile in `dir` from its OWN
				// last position. Emit the delta (DX/DY) — the world mirror applies
				// it to the stored position. The old code stamped ownX+dx (the
				// HOST's tile ± 1), which teleported every moving NPC onto the host
				// and produced the phantom crowd standing on bernard.
				dx, dy := npcMoveOffset(int(dir))
				events = append(events, event.NpcNearby{
					Index:  idx,
					DX:     dx,
					DY:     dy,
					Sprite: int(dir), // movement direction == facing (0-7)
					IsNew:  false,
				})
			}
		} else {
			// Not moving: 2 bits removal-vs-sprite.
			subType, err := b.ReadBits(2)
			if err != nil {
				return events, int(localCount), nil
			}
			if subType == 3 {
				// REMOVE_NPC: the server dropped this NPC from view
				// (despawn / death / left range). Emit a removal so the
				// world mirror prunes it and combat.target clears.
				idx := indexAt(i)
				if idx >= 0 {
					events = append(events, event.NpcNearby{Index: idx, Removed: true})
				}
			} else {
				// First 2 bits of a 4-bit sprite; read 2 more. Sprite-only
				// change (e.g. facing / combat animation) — no position or
				// liveness change to surface.
				if _, err := b.ReadBits(2); err != nil {
					return events, int(localCount), nil
				}
			}
		}
	}

	// New NPCs section: 36 bits each.
	for {
		// We need 36 bits. Compute remaining bits cheaply.
		consumedBits := b.bitPos - b.rpos*8
		totalBits := (b.wpos - b.rpos) * 8
		if totalBits-consumedBits < 36 {
			break
		}
		idx, err := b.ReadBits(12)
		if err != nil {
			break
		}
		offX, err := b.ReadBits(5)
		if err != nil {
			break
		}
		offY, err := b.ReadBits(5)
		if err != nil {
			break
		}
		sprite, err := b.ReadBits(4)
		if err != nil {
			break
		}
		typeID, err := b.ReadBits(10)
		if err != nil {
			break
		}

		// Convert 5-bit signed offset (-16..15).
		dx := int(offX)
		if dx >= 16 {
			dx -= 32
		}
		dy := int(offY)
		if dy >= 16 {
			dy -= 32
		}

		events = append(events, event.NpcNearby{
			Index:  int(idx),
			X:      ownX + dx,
			Y:      ownY + dy,
			Sprite: int(sprite),
			TypeID: int(typeID),
			IsNew:  true,
		})

		if len(events) > 250 {
			break
		}
	}

	b.FinishBitAccess()

	// Collapse the REMOVE_NPC + re-add-as-NEW churn the server emits for
	// an NPC that's IN COMBAT with us: each tick the server removes it
	// from the positional list (its combat sprite is incompatible with a
	// movement update) and immediately re-adds it as a "new" NPC with the
	// fresh combat sprite (GameStateUpdater.updateNpcs, the inCombat()
	// remove + re-add). That is NOT a despawn — if we honoured the
	// REMOVE we'd prune the engaged NPC (and its hits) every tick. So
	// when the same index appears as BOTH removed and new in one packet,
	// drop the removal: only the NEW (position/type refresh) survives,
	// and world.Npcs.Set carries the accumulated combat state forward. A
	// REMOVE with NO matching re-add is a true despawn/death and is kept.
	return dropChurnedRemovals(events), int(localCount), nil
}

// dropChurnedRemovals filters out REMOVE events for any NPC index that
// is also (re-)added as a new NPC in the same packet — the in-combat
// remove+readd churn. Order is otherwise preserved.
func dropChurnedRemovals(events []event.Event) []event.Event {
	readded := make(map[int]bool)
	for _, ev := range events {
		if n, ok := ev.(event.NpcNearby); ok && n.IsNew {
			readded[n.Index] = true
		}
	}
	if len(readded) == 0 {
		return events
	}
	out := events[:0]
	for _, ev := range events {
		if n, ok := ev.(event.NpcNearby); ok && n.Removed && readded[n.Index] {
			continue
		}
		out = append(out, ev)
	}
	return out
}

// npcMoveOffset maps a 3-bit RSC movement sprite/direction (0-7) to a
// one-tile (dx, dy) delta. RSC compass order, matching the client's
// movement decode: 0=N, 1=NE, 2=E, 3=SE, 4=S, 5=SW, 6=W, 7=NW. In RSC
// wire space north is -Y. Used to re-stamp a moving NPC's position from
// a movement update so world.npcs positions stay roughly current.
func npcMoveOffset(dir int) (int, int) {
	switch dir {
	case 0: // N
		return 0, -1
	case 1: // NE
		return 1, -1
	case 2: // E
		return 1, 0
	case 3: // SE
		return 1, 1
	case 4: // S
		return 0, 1
	case 5: // SW
		return -1, 1
	case 6: // W
		return -1, 0
	case 7: // NW
		return -1, -1
	}
	return 0, 0
}
