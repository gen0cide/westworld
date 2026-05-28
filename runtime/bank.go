package runtime

import (
	"context"
	"time"

	"github.com/gen0cide/westworld/action"
)

// BankDeposit moves the given catalog item into the bank. The bot
// must have already opened the bank (via TalkToNpc on a banker).
func (h *Host) BankDeposit(ctx context.Context, catalogID, amount int) error {
	return action.BankDeposit(ctx, h.conn, catalogID, amount)
}

// BankWithdraw is the dual of BankDeposit.
func (h *Host) BankWithdraw(ctx context.Context, catalogID, amount int) error {
	return action.BankWithdraw(ctx, h.conn, catalogID, amount)
}

// BankClose closes the bank window.
func (h *Host) BankClose(ctx context.Context) error {
	return action.BankClose(ctx, h.conn)
}

// DepositAll deposits every inventory item whose ItemID isn't in
// `keepIDs`. Useful at end of a kill run. Sleeps a tick between
// deposits so the inventory mirror catches up before we read it
// again on the next iteration.
func (h *Host) DepositAll(ctx context.Context, keepIDs map[int]struct{}) error {
	slots := h.world.Inventory.Slots()
	for i, s := range slots {
		if s.ItemID == 0 {
			continue
		}
		if _, keep := keepIDs[s.ItemID]; keep {
			continue
		}
		amount := s.Amount
		if amount <= 0 {
			amount = 1
		}
		h.log.Info("DepositAll: depositing",
			"slot", i,
			"item_id", s.ItemID,
			"amount", amount,
		)
		if err := h.BankDeposit(ctx, s.ItemID, amount); err != nil {
			h.log.Warn("DepositAll: BankDeposit failed", "item_id", s.ItemID, "err", err)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(700 * time.Millisecond):
		}
	}
	return nil
}
