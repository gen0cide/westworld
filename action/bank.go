package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Outbound bank opcodes per Payload235Parser.java.
//
// The bank UI must already be open server-side (player.accessingBank()
// must return true). That state is set when the player TalkTo's a
// banker NPC. Without it, the deposit/withdraw handlers reject the
// packet as "suspicious".
const (
	outBankClose    byte = 212
	outBankWithdraw byte = 22
	outBankDeposit  byte = 23
)

// BankDeposit moves `amount` of the given catalogID item from
// inventory to bank. catalogID is the item def ID; for stackable
// items like coins this is one packet for the entire amount.
//
// Wire format:
//   [short catalogID]
//   [int   amount]
//   [int   magicNumber] — server-acknowledged legacy field, ignored
func BankDeposit(ctx context.Context, conn *session.Conn, catalogID, amount int) error {
	if catalogID < 0 || catalogID > 0xFFFF {
		return fmt.Errorf("action: BankDeposit catalogID %d out of uint16 range", catalogID)
	}
	if amount <= 0 {
		return fmt.Errorf("action: BankDeposit amount must be > 0 (got %d)", amount)
	}
	buf := v235.NewBuffer(10)
	buf.WriteUint16(uint16(catalogID))
	buf.WriteUint32(uint32(amount))
	buf.WriteUint32(0) // magicNumber — ignored by server
	return conn.Send(outBankDeposit, buf.Bytes())
}

// BankWithdraw is the dual of BankDeposit.
func BankWithdraw(ctx context.Context, conn *session.Conn, catalogID, amount int) error {
	if catalogID < 0 || catalogID > 0xFFFF {
		return fmt.Errorf("action: BankWithdraw catalogID %d out of uint16 range", catalogID)
	}
	if amount <= 0 {
		return fmt.Errorf("action: BankWithdraw amount must be > 0 (got %d)", amount)
	}
	buf := v235.NewBuffer(10)
	buf.WriteUint16(uint16(catalogID))
	buf.WriteUint32(uint32(amount))
	buf.WriteUint32(0)
	return conn.Send(outBankWithdraw, buf.Bytes())
}

// BankClose tells the server the bot is done banking. Releases the
// accessingBank flag.
func BankClose(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outBankClose, nil)
}
