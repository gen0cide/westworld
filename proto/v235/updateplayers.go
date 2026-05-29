package v235

import (
	"fmt"

	"github.com/gen0cide/westworld/event"
)

// DecodeUpdatePlayers parses opcode 234 (InSendUpdatePlayers), which
// carries appearance, chat, damage, and projectile updates for
// players visible to us.
//
// Source: GameStateUpdater.java:584-1050+ (the appearance-update
// emitter); Payload235Generator.java:625-643.
//
// Top-level layout:
//
//	[short] updateCount (big-endian)
//
//	for i = 0..updateCount-1:
//	    [short] playerIndex (server's stable id for the player)
//	    [byte]  updateType:
//	            0 = action bubble
//	            1 = public chat
//	            2 = damage taken
//	            3 = projectile fired at NPC
//	            4 = projectile fired at player
//	            5 = appearance / identity
//	            6 = quest chat
//	            7 = muted chat / tutorial chat
//	    [type-specific fields]
//
// Phase 1.6 decodes the outer structure and the four most-impactful
// types (chat, damage, appearance, action-bubble). Projectiles are
// stubbed for now.
//
// Returns a slice of events; the caller's host loop applies them
// individually to the world state and publishes them.
func DecodeUpdatePlayers(payload []byte) ([]event.Event, error) {
	b := WrapBuffer(payload)
	count, err := b.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("updateplayers count: %w", err)
	}

	out := make([]event.Event, 0, count)
	for i := 0; i < int(count); i++ {
		if b.Len() < 3 {
			break // ran out — payload didn't match count (or we mis-parsed earlier)
		}
		idx, err := b.ReadUint16()
		if err != nil {
			break
		}
		typ, err := b.ReadByte()
		if err != nil {
			break
		}
		switch typ {
		case 0:
			// Action bubble.
			bubbleID, _ := b.ReadUint16()
			out = append(out, event.PlayerActionBubble{
				PlayerIndex: int(idx),
				BubbleID:    int(bubbleID),
			})
		case 1, 6, 7:
			// Public / quest / muted chat. Body is RSC-compressed
			// (RSCString). Format: [byte icon] [smart_len chars]
			// [N compressed bytes]. N is determined by the compressed
			// data — we decode it to find out how many bytes we
			// consumed.
			icon, _ := b.ReadByte()
			rscLen, _ := b.ReadSmart08_16()
			// We don't know how many bytes the compressed body is
			// up-front; peek at the remaining bytes, decompress, and
			// then advance by however many we consumed.
			remaining := b.Bytes()
			decoded, consumed := DecipherRSCStringWithLen(remaining, rscLen)
			// Manually advance the reader past the consumed bytes.
			_, _ = b.ReadBytes(consumed)
			out = append(out, event.OtherPlayerChat{
				PlayerIndex: int(idx),
				Icon:        int(icon),
				ChatKind:    chatKindName(typ),
				MessageText: decoded,
				MessageRaw:  remaining[:consumed],
			})
		case 2:
			// Damage taken.
			dmg, _ := b.ReadByte()
			curHits, _ := b.ReadByte()
			maxHits, _ := b.ReadByte()
			out = append(out, event.OtherPlayerDamage{
				PlayerIndex: int(idx),
				Damage:      int(dmg),
				CurHits:     int(curHits),
				MaxHits:     int(maxHits),
			})
		case 3:
			// Projectile to NPC: [short projectileType] [short victimNpcIndex].
			projID, _ := b.ReadUint16()
			victimNpc, _ := b.ReadUint16()
			out = append(out, event.OtherPlayerProjectile{
				CasterIndex:    int(idx),
				ProjectileID:   int(projID),
				VictimNpcIndex: int(victimNpc),
				VictimIsNpc:    true,
			})
		case 4:
			// Projectile to player: [short projectileType] [short victimPlayerIndex].
			projID, _ := b.ReadUint16()
			victimPlayer, _ := b.ReadUint16()
			out = append(out, event.OtherPlayerProjectile{
				CasterIndex:       int(idx),
				ProjectileID:      int(projID),
				VictimPlayerIndex: int(victimPlayer),
				VictimIsNpc:       false,
			})
		case 5:
			// Appearance / identity update. Variable-length, complex
			// (different sub-layouts per client version). We decode the
			// identity (name + appearance id) AND the two trailing
			// combat-state bytes the v235 wire carries (combatLevel +
			// skullType), consuming the equipment/colour blocks in
			// between.
			//
			// v235 (isUsing233CompatibleClient) layout, per
			// GameStateUpdater.issuePlayerAppearanceUpdatePacket
			// (GameStateUpdater.java:845-1011):
			//
			//	[short]    appearanceID (the VIEWER's id — a server quirk,
			//	           but always 2 bytes on the wire)
			//	[zqstring] playerName
			//	[zqstring] playerName (full) OR 1-char abbreviation when
			//	           the packet carries >= 65 appearance updates
			//	[byte]     equipmentCount N
			//	[byte × N] worn-item sprites (writeAppearanceByte → 1 byte each)
			//	[byte]     hairColour
			//	[byte]     topColour
			//	[byte]     trouserColour
			//	[byte]     skinColour
			//	[byte]     combatLevel   <- NEW: decoded for combat perception
			//	[byte]     skullType     <- NEW: 0=none, 1=skulled/PK-flagged
			//
			// Heuristic: try to read the fixed prefix; if anything fails,
			// abort this update record and continue. The two combat
			// bytes are only emitted (HasCombat=true) when both were
			// successfully read — a truncated record yields HasCombat
			// false so consumers don't mistake a missing field for a
			// combat-level of 0.
			appearanceID, errA := b.ReadUint16()
			if errA != nil {
				// We can't safely advance; truncate.
				break
			}
			name, errN := b.ReadZeroQuotedString()
			if errN != nil {
				break
			}
			// The second name (or 1-char abbreviation) — try to skip it.
			_, _ = b.ReadZeroQuotedString()
			// Equipment block: [byte count] then count bytes.
			eqCount, _ := b.ReadByte()
			// Skip the equipment bytes (each is 1 byte in v235).
			if int(eqCount) > 0 && b.Len() >= int(eqCount) {
				_, _ = b.ReadBytes(int(eqCount))
			}
			// Colours block: 4 bytes (hair, top, trouser, skin).
			if b.Len() >= 4 {
				_, _ = b.ReadBytes(4)
			}
			// Combat-state block: 2 bytes (combatLevel, skullType).
			// Only mark HasCombat when both are present.
			ap := event.OtherPlayerAppearance{
				PlayerIndex:  int(idx),
				Name:         name,
				AppearanceID: int(appearanceID),
			}
			if b.Len() >= 2 {
				combatLevel, errC := b.ReadByte()
				skullType, errS := b.ReadByte()
				if errC == nil && errS == nil {
					ap.CombatLevel = int(combatLevel)
					ap.SkullType = int(skullType)
					ap.HasCombat = true
				}
			}
			out = append(out, ap)
		default:
			// Unknown update type — best-effort: stop processing
			// further records to avoid mis-aligned reads.
			break
		}
	}
	return out, nil
}

func chatKindName(typeByte byte) string {
	switch typeByte {
	case 1:
		return "public"
	case 6:
		return "quest"
	case 7:
		return "muted"
	}
	return "unknown"
}
