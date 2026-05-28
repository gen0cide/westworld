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
	outNpcTalkTo  byte = 153 // [short serverIndex]
	outNpcCommand byte = 202 // right-click NPC (depends on default cmd: talk/attack/use)
	outNpcAttack  byte = 190 // [short serverIndex]
	outPlayerAttack byte = 171 // [short serverIndex] — PVP
)

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
