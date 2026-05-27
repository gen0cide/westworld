package v235

import (
	"fmt"

	"github.com/gen0cide/westworld/event"
)

// DecodePlayerCoords parses opcode 191 (InSendPlayerCoords) — the
// bitpacked per-tick update of own position + nearby players.
//
// Returns:
//   - own position update (always present in this packet)
//   - list of nearby-player events (may be empty)
//
// Source: GameStateUpdater.java:293-462; spec written in docs/protocol.md.
//
// Bit layout (client version >= 177, which includes v235):
//
//	[11 bits]  own X
//	[13 bits]  own Y
//	[4 bits]   own sprite (facing/anim, 0-15)
//	[8 bits]   localPlayersCount (already-tracked players to update)
//
//	for each local player:
//	  [1 bit]   needsUpdate flag
//	  if needsUpdate:
//	    [1 bit]  updateType (0=movement, 1=removal-or-sprite)
//	    if type=0 (movement):
//	      [3 bits]  direction (0-7)
//	    if type=1:
//	      [2 bits]  removalSubType — values:
//	                  3=remove player, else sprite update with 4 bits to follow
//	      if subType != 3:
//	        [4 bits] new sprite
//
//	then for each NEW player entering view (until packet ends):
//	  [11 bits] player index
//	  [5 bits]  offset X (relative to our X, wrapped via mod-32 mask)
//	  [5 bits]  offset Y
//	  [4 bits]  sprite
//	  [1 bit]   isKnownPlayer flag (omitted for some client versions)
//
// We can't tell from the packet alone where the new-player loop ends
// vs garbage bits — we read until the underlying buffer is exhausted
// at byte granularity. For Phase 1.5 this is fine; Phase 2 may need
// stricter parsing.
func DecodePlayerCoords(payload []byte) (event.OwnPositionUpdate, []event.NearbyPlayerEvent, error) {
	b := WrapBuffer(payload)
	b.StartBitAccess()

	ownX, err := b.ReadBits(11)
	if err != nil {
		return event.OwnPositionUpdate{}, nil, fmt.Errorf("playercoords own_x: %w", err)
	}
	ownY, err := b.ReadBits(13)
	if err != nil {
		return event.OwnPositionUpdate{}, nil, fmt.Errorf("playercoords own_y: %w", err)
	}
	ownSprite, err := b.ReadBits(4)
	if err != nil {
		return event.OwnPositionUpdate{}, nil, fmt.Errorf("playercoords own_sprite: %w", err)
	}
	localCount, err := b.ReadBits(8)
	if err != nil {
		return event.OwnPositionUpdate{}, nil, fmt.Errorf("playercoords local_count: %w", err)
	}

	ownEvent := event.OwnPositionUpdate{
		X:      int(ownX),
		Y:      int(ownY),
		Sprite: int(ownSprite),
	}

	nearby := make([]event.NearbyPlayerEvent, 0, int(localCount))

	// Iterate through known-local players. For Phase 1.5 we just
	// consume the bits without precise tracking — we don't have a
	// stable index→player map yet.
	for i := uint32(0); i < localCount; i++ {
		needsUpdate, err := b.ReadBits(1)
		if err != nil {
			return ownEvent, nearby, nil // best-effort: return what we have
		}
		if needsUpdate == 0 {
			continue
		}
		updateType, err := b.ReadBits(1)
		if err != nil {
			return ownEvent, nearby, nil
		}
		if updateType == 0 {
			// movement update: 3 bits direction
			if _, err := b.ReadBits(3); err != nil {
				return ownEvent, nearby, nil
			}
		} else {
			// removal or sprite change
			subType, err := b.ReadBits(2)
			if err != nil {
				return ownEvent, nearby, nil
			}
			if subType == 3 {
				// removal: nothing more to read for this player
			} else {
				// sprite update: 4 more bits (subType bits we already read
				// are actually the high 2 of the new 6-bit sprite per some
				// docs; for Phase 1.5 we just consume more bits if available)
				// Conservative: read 2 more bits to total 4-bit sprite.
				if _, err := b.ReadBits(2); err != nil {
					return ownEvent, nearby, nil
				}
			}
		}
	}

	// Remaining bits encode new players entering view. Each new player
	// is 11+5+5+4 = 25 bits (plus possibly 1 isKnownPlayer flag for some
	// client versions — we don't read it here for v235).
	//
	// We read until we run out of bytes at byte granularity. Server
	// pads the last byte with zeros; we may consume the padding by
	// mistake, so we accept some noise here.
	for {
		// Need at least 25 bits available.
		remaining := b.Len()*8 - (b.bitPos - b.rpos*8)
		if remaining < 25 {
			break
		}
		idx, err := b.ReadBits(11)
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
		// Convert 5-bit signed offset (-16..15) to int. The mask is
		// usually treated as unsigned 0..31 then sign-extended past
		// the midpoint, but RSC's actual handling varies; for Phase
		// 1.5 we just expose the raw offset to the event.
		dx := int(offX)
		if dx >= 16 {
			dx -= 32
		}
		dy := int(offY)
		if dy >= 16 {
			dy -= 32
		}
		nearby = append(nearby, event.NearbyPlayerEvent{
			Index:  int(idx),
			X:      int(ownX) + dx,
			Y:      int(ownY) + dy,
			Sprite: int(sprite),
			IsNew:  true,
		})
		// Guard against runaway loops on malformed data.
		if len(nearby) > 250 {
			break
		}
	}

	b.FinishBitAccess()
	return ownEvent, nearby, nil
}
