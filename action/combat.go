package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Combat + NPC interaction outbound opcodes.
//
// Source: Payload235Parser.java (verified case → OpcodeIn mapping).
const (
	outNpcTalkTo    byte = 153 // [short serverIndex]
	outNpcCommand   byte = 202 // right-click NPC (depends on default cmd: talk/attack/use)
	outNpcAttack    byte = 190 // [short serverIndex]
	outPlayerAttack byte = 171 // [short serverIndex] — PVP
	outCombatStyle  byte = 29  // [byte style] — 0=controlled, 1=aggressive, 2=accurate, 3=defensive
)

// CombatStyle is the active melee xp-split mode. Values match the
// OpenRSC CombatStyle enum + the RSC client's combat-mode toggle.
type CombatStyle byte

const (
	CombatStyleControlled CombatStyle = 0 // even split: 1/3 each attack/strength/defense
	CombatStyleAggressive CombatStyle = 1 // all strength
	CombatStyleAccurate   CombatStyle = 2 // all attack
	CombatStyleDefensive  CombatStyle = 3 // all defense
)

// String returns the canonical lowercase style name used by the DSL
// combat.style read-side view (the inverse of the name→CombatStyle
// mapping in runtime/actions_combat.go::dslSetCombatStyle). Unknown
// values render as "controlled" (the server default).
func (s CombatStyle) String() string {
	switch s {
	case CombatStyleAggressive:
		return "aggressive"
	case CombatStyleAccurate:
		return "accurate"
	case CombatStyleDefensive:
		return "defensive"
	default:
		return "controlled"
	}
}

// SetCombatStyle sends opcode 29 to change the player's melee
// combat style. Takes effect on the next attack tick.
func SetCombatStyle(ctx context.Context, conn *session.Conn, style CombatStyle) error {
	if style > CombatStyleDefensive {
		return fmt.Errorf("action: SetCombatStyle style %d out of range [0..3]", style)
	}
	buf := v235.NewBuffer(1)
	buf.WriteByte(byte(style))
	return conn.Send(outCombatStyle, buf.Bytes())
}

// AttackNpc initiates combat with an NPC at the given server index.
// Server resolves whether attack is legal (NPC must be attackable, in
// range, not already in combat, etc.).
func AttackNpc(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outNpcAttack, serverIndex, "AttackNpc")
}

// AttackPlayer initiates PVP with a player at the given server index.
// Only legal in PVP zones (wilderness above level 1, duels, certain
// minigames).
func AttackPlayer(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outPlayerAttack, serverIndex, "AttackPlayer")
}

// TalkToNpc opens dialog with an NPC. Server responds with
// NPC dialog text + optional menu of dialog options.
func TalkToNpc(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outNpcTalkTo, serverIndex, "TalkToNpc")
}

// NpcCommand right-clicks an NPC (default command — typically "Talk-to"
// for friendly NPCs, "Attack" for hostile ones).
func NpcCommand(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outNpcCommand, serverIndex, "NpcCommand")
}

func sendShortPacket(ctx context.Context, conn *session.Conn, opcode byte, serverIndex int, name string) error {
	if serverIndex < 0 || serverIndex > 0xFFFF {
		return fmt.Errorf("action: %s serverIndex %d out of range", name, serverIndex)
	}
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(serverIndex))
	return conn.Send(opcode, buf.Bytes())
}
