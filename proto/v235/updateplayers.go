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
			// Public (1) / quest (6) / muted (7) chat. Body is RSC-compressed
			// (RSCString). Format: [byte icon (type 1 ONLY)] [smart_len chars]
			// [N compressed bytes]. The OpenRSC server writes the leading icon
			// byte only when updateType != 6 (GameStateUpdater.java:791-793), so
			// quest chat (type 6) has NO icon — reading it unconditionally
			// over-consumes one byte and shifts the length prefix + Huffman body,
			// garbling the start of the message (it self-resyncs at the tail).
			var icon byte
			if typ == 1 || typ == 7 {
				icon, _ = b.ReadByte()
			}
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
			// Equipment block: [byte count] then count worn-sprite bytes,
			// one per equip slot in AppearanceId.SLOT_* order (head,
			// shirt, pants, shield, weapon, hat, body, legs, gloves,
			// boots, amulet, cape — int[12]). Each byte is the worn
			// SPRITE id (itemDef.getAppearanceId() & 0xFF), NOT a
			// catalogue item id. Authentic v235 normally sends all 12.
			ap := event.OtherPlayerAppearance{
				PlayerIndex:  int(idx),
				Name:         name,
				AppearanceID: int(appearanceID),
			}
			eqCount, _ := b.ReadByte()
			if int(eqCount) > 0 && b.Len() >= int(eqCount) {
				worn, _ := b.ReadBytes(int(eqCount))
				ap.WornCount = len(worn)
				for i := 0; i < len(worn) && i < event.NumEquipSlots; i++ {
					ap.WornSprites[i] = int(worn[i])
				}
				ap.HasWorn = true
			}
			// Colours block: 4 bytes (hair, top, trouser/bottom, skin) — the
			// exact order the server writes them (GameStateUpdater appearance
			// writer; client decode at mudclient.java:5429-5432:
			// colourHair, colourTop, colourBottom, colourSkin), and they
			// precede the 2 combat bytes already kept. These are clothing-table
			// INDICES (hair/top/trouser) + a skin index; the renderer dyes the
			// composite layers with them.
			if b.Len() >= 4 {
				cols, _ := b.ReadBytes(4)
				if len(cols) == 4 {
					ap.HairColour = int(cols[0])
					ap.TopColour = int(cols[1])
					ap.TrouserColour = int(cols[2])
					ap.SkinColour = int(cols[3])
					ap.HasColours = true
				}
			}
			// Combat-state block: 2 bytes (combatLevel, skullType).
			// Only mark HasCombat when both are present.
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
