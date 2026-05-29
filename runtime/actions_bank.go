package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Banking action handler bodies (open_bank / deposit / withdraw /
// close_bank). Registered in the central actionHandlers table in
// dsl_actions.go. Item+amount resolution lives in dsl_helpers.go
// (resolveItemAmount).

func dslDeposit(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankDeposit(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslWithdraw(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankWithdraw(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCloseBank(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.BankClose(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslOpenBank(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	// open_bank(banker) = talk_to(banker); the DSL author then uses
	// answer(N) to pick the "I'd like to access my bank" option.
	// We delegate to talk_to since RSC has no dedicated open-bank
	// packet — the bank UI is the bottom of a dialog tree.
	return dslTalkTo(ctx, h, args, nil)
}

// ----- bank bulk verbs (#117/#118) -----

// dslDepositAll wires Host.DepositAll: bank.deposit_all(keep?: List<Int>)
// deposits every inventory item except those whose ItemID appears in
// the optional keep-list. The keep-list may be raw item IDs (Int),
// item names (String), or item views (Getter exposing .id) — each is
// resolved through the facts registry. Requires an open bank.
func dslDepositAll(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if !h.world.Bank.IsOpen() {
		return interp.Fail(interp.BANK_NOT_OPEN, "bank.deposit_all: bank is not open"), nil
	}
	keep, err := resolveKeepIDs(h, args, named)
	if err != nil {
		return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
	}
	if err := h.DepositAll(ctx, keep); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslWithdrawAll wires Host.WithdrawAll: bank.withdraw_all(item)
// withdraws the entire banked quantity of one item. Requires an open
// bank.
func dslWithdrawAll(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if !h.world.Bank.IsOpen() {
		return interp.Fail(interp.BANK_NOT_OPEN, "bank.withdraw_all: bank is not open"), nil
	}
	itemID, err := resolveBankItem(h, args, named)
	if err != nil {
		return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
	}
	if err := h.WithdrawAll(ctx, itemID); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslWithdrawX wires Host.WithdrawX: bank.withdraw_x(item, amount)
// withdraws a preset amount, clamped to the banked quantity. Requires
// an open bank.
func dslWithdrawX(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if !h.world.Bank.IsOpen() {
		return interp.Fail(interp.BANK_NOT_OPEN, "bank.withdraw_x: bank is not open"), nil
	}
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
	}
	if err := h.WithdrawX(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// resolveBankItem extracts a single item ID for withdraw_all from
// either a positional arg or the `item` named arg.
func resolveBankItem(h *Host, args []interp.Value, named map[string]interp.Value) (int, error) {
	var itemVal interp.Value
	if v, ok := named["item"]; ok {
		itemVal = v
	} else if len(args) >= 1 {
		itemVal = args[0]
	}
	if itemVal == nil {
		return 0, errf("bank.withdraw_all needs an item (id, name, or item view)")
	}
	return resolveItemID(h.facts, itemVal)
}

// resolveKeepIDs turns the optional deposit_all keep argument into a
// set of item IDs. Accepts a positional List or the `keep` named arg;
// missing => empty set (deposit everything). List elements may be Int
// IDs, String names, or item views — each resolved via facts.
func resolveKeepIDs(h *Host, args []interp.Value, named map[string]interp.Value) (map[int]struct{}, error) {
	keep := map[int]struct{}{}
	var listVal interp.Value
	if v, ok := named["keep"]; ok {
		listVal = v
	} else if len(args) >= 1 {
		listVal = args[0]
	}
	if listVal == nil {
		return keep, nil
	}
	lst, ok := listVal.(*interp.List)
	if !ok {
		return nil, errf("bank.deposit_all: keep must be a list of item ids/names")
	}
	for _, el := range lst.Items {
		id, err := resolveItemID(h.facts, el)
		if err != nil {
			return nil, err
		}
		keep[id] = struct{}{}
	}
	return keep, nil
}
