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
// Returns events for new NPCs entering view; existing-NPC updates
// (movement, sprite change, removal) are consumed but only the
// movement/removal events are emitted (sprite-only is suppressed
// for noise reduction).
func DecodeNpcCoords(payload []byte, ownX, ownY int) ([]event.Event, error) {
	b := WrapBuffer(payload)
	b.StartBitAccess()

	localCount, err := b.ReadBits(8)
	if err != nil {
		return nil, fmt.Errorf("npccoords local_count: %w", err)
	}

	var events []event.Event

	// Iterate through existing-NPC update records. For Phase 1.6 we
	// don't have a per-host NPC tracking table yet, so we can't
	// produce "NPC X moved to..." events from these. We just consume
	// the bits correctly so the new-NPCs section parses.
	for i := uint32(0); i < localCount; i++ {
		needsUpdate, err := b.ReadBits(1)
		if err != nil {
			return events, nil
		}
		if needsUpdate == 0 {
			continue
		}
		discrim, err := b.ReadBits(1)
		if err != nil {
			return events, nil
		}
		if discrim == 0 {
			// Movement: 3 bits direction
			if _, err := b.ReadBits(3); err != nil {
				return events, nil
			}
		} else {
			// Not moving: 2 bits removal-vs-sprite
			subType, err := b.ReadBits(2)
			if err != nil {
				return events, nil
			}
			if subType == 3 {
				// Removal — nothing more to read for this NPC.
			} else {
				// First 2 bits of a 4-bit sprite; read 2 more.
				if _, err := b.ReadBits(2); err != nil {
					return events, nil
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
			Index:    int(idx),
			X:        ownX + dx,
			Y:        ownY + dy,
			Sprite:   int(sprite),
			TypeID:   int(typeID),
			IsNew:    true,
		})

		if len(events) > 250 {
			break
		}
	}

	b.FinishBitAccess()
	return events, nil
}
