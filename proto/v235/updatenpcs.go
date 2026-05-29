package v235

import (
	"fmt"

	"github.com/gen0cide/westworld/event"
)

// DecodeUpdateNpcs parses opcode 104 (InUpdateNpc / SEND_UPDATE_NPC),
// the per-tick batch of per-NPC state changes for NPCs visible to us:
// chat, damage (the NPC's OWN current/max hitpoints), and — only for
// custom clients — projectiles, skulls, wields and bubbles.
//
// Source: GameStateUpdater.updateNpcAppearances
// (GameStateUpdater.java:464-577); shares the AppearanceUpdateStruct
// wire shape with SEND_UPDATE_PLAYERS (opcode 234) per
// Payload235Generator.java:625-644.
//
// Top-level layout (identical outer framing to opcode 234):
//
//	[short] updateCount (big-endian)
//
//	for i = 0..updateCount-1:
//	    [short] npcIndex (server's stable id for the NPC)
//	    [byte]  updateType
//	    [type-specific fields]
//
// Update types (authentic v235 = NOT a custom client, so only 1 and 2
// are ever sent; 3-7 are decoded for completeness / forward-compat):
//
//	1 = chat:
//	      [short]      recipientIndex (-1 = broadcast / none)
//	      [smart08_16] decipheredLength
//	      [N bytes]    RSC-compressed body
//	2 = damage / hits  (the NPC's OWN health):
//	      [byte] damage
//	      [byte] curHits
//	      [byte] maxHits
//	3 = projectile at NPC:    [short projectileType] [short victimNpcIndex]
//	4 = projectile at player: [short projectileType] [short victimPlayerIndex]
//	5 = skull:                [byte skullType]
//	6 = wield:                [byte wield] [byte wield2]
//	7 = action bubble:        [short bubbleId]
//
// Returns a slice of events; the caller's host loop applies them
// individually to the world state and publishes them. Decoding the
// type-2 records is what un-stubs Npc.health for ANY visible NPC,
// since opcode 234 only ever surfaces NPC health indirectly (as a
// projectile victim).
func DecodeUpdateNpcs(payload []byte) ([]event.Event, error) {
	b := WrapBuffer(payload)
	count, err := b.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("updatenpcs count: %w", err)
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
		case 1:
			// NPC chat. recipientIndex is signed on the wire (-1 = none);
			// the server writes it as a short.
			recip, err := b.ReadUint16()
			if err != nil {
				break
			}
			recipient := int(int16(recip))
			rscLen, _ := b.ReadSmart08_16()
			remaining := b.Bytes()
			decoded, consumed := DecipherRSCStringWithLen(remaining, rscLen)
			_, _ = b.ReadBytes(consumed)
			out = append(out, event.NpcChat{
				NpcIndex:       int(idx),
				RecipientIndex: recipient,
				MessageText:    decoded,
				MessageRaw:     remaining[:consumed],
			})
		case 2:
			// Damage / hits update — the NPC's OWN current/max
			// hitpoints. This is the field opcode 234 never carries.
			dmg, _ := b.ReadByte()
			curHits, _ := b.ReadByte()
			maxHits, _ := b.ReadByte()
			out = append(out, event.NpcDamage{
				NpcIndex: int(idx),
				Damage:   int(dmg),
				CurHits:  int(curHits),
				MaxHits:  int(maxHits),
			})
		case 3:
			// Projectile to NPC: [short projectileType] [short victimNpcIndex].
			projID, _ := b.ReadUint16()
			victimNpc, _ := b.ReadUint16()
			out = append(out, event.NpcProjectile{
				CasterNpcIndex: int(idx),
				ProjectileID:   int(projID),
				VictimNpcIndex: int(victimNpc),
				VictimIsNpc:    true,
			})
		case 4:
			// Projectile to player: [short projectileType] [short victimPlayerIndex].
			projID, _ := b.ReadUint16()
			victimPlayer, _ := b.ReadUint16()
			out = append(out, event.NpcProjectile{
				CasterNpcIndex:    int(idx),
				ProjectileID:      int(projID),
				VictimPlayerIndex: int(victimPlayer),
				VictimIsNpc:       false,
			})
		case 5:
			// Skull update: [byte skullType]. Consumed but not surfaced
			// (NPCs don't carry the PK-flag perception we expose for
			// players); reading it keeps subsequent records aligned.
			if _, err := b.ReadByte(); err != nil {
				return out, nil
			}
		case 6:
			// Wield update: [byte wield] [byte wield2]. Consumed only.
			if b.Len() < 2 {
				return out, nil
			}
			_, _ = b.ReadByte()
			_, _ = b.ReadByte()
		case 7:
			// Action bubble: [short bubbleId]. Consumed only.
			if _, err := b.ReadUint16(); err != nil {
				return out, nil
			}
		default:
			// Unknown update type — stop to avoid mis-aligned reads.
			return out, nil
		}
	}
	return out, nil
}
