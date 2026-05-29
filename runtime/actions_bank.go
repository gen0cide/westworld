package runtime

import (
	"context"
	"strings"
	"time"

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

// bankAccessNeedles are the substrings (case-insensitive) we scan the
// banker's dialog menu for to find the "open my bank" option. RSC
// banker scripts present "I'd like to access my bank account please"
// as the first option; we match on the stable middle phrase first,
// then degrade to broader fallbacks so a localized / lightly-reworded
// menu still resolves. Order matters: most-specific first.
var bankAccessNeedles = []string{
	"access my bank",
	"bank account",
	"access my account",
}

// dialog menu / bank-open polling cadence + budgets. The menu (opcode
// 245) lands ~2 server ticks after the talk packet; the bank window
// (opcode 42) lands ~1 tick after we answer. 8s budgets absorb a
// double-tick plus the walk-adjacent settle on a busy server.
const (
	bankDialogPoll     = 200 * time.Millisecond
	bankDialogTimeout  = 8 * time.Second
	bankOpenPollBudget = 8 * time.Second
)

// isFail reports whether a handler return value is a CallResult
// carrying a typed error (the .err path). Used to short-circuit a
// multi-step handler when an inner step failed.
func isFail(v interp.Value) bool {
	r, ok := v.(*interp.CallResult)
	return ok && r.Err != nil
}

// dslOpenBank drives the banker conversation all the way to the open
// bank screen. RSC has no dedicated "open bank" packet — the bank UI
// is the bottom of a dialog tree: TalkTo the banker, wait for the
// options menu, pick "I'd like to access my bank account please", and
// the server replies with SEND_BANK_OPEN (opcode 42) which flips
// world.bank.is_open true.
//
// This used to be a bare alias for talk_to, forcing every routine to
// hand-roll wait_for_dialog + answer(find_option(...)) + wait_until.
// Now bank.open(banker) does the whole handshake and returns Ok only
// once the bank window is actually up (Fail otherwise), so a routine
// can write `bank.open(banker)` then `bank.deposit(...)` directly.
//
// Idempotent: if the bank is already open we return Ok without
// re-talking.
func dslOpenBank(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if h.world.Bank.IsOpen() {
		return interp.Ok(interp.Null{}), nil
	}
	// Invalidate any stale menu cached from a prior interaction so the
	// pollDialogOptions below blocks for the banker's OWN menu rather
	// than matching leftover options.
	h.world.Recent.ClearDialogOptions()
	// Walk to + talk to the banker (same target resolution as talk_to:
	// npc view or raw server index).
	if res, err := dslTalkTo(ctx, h, args, nil); err != nil || isFail(res) {
		// Propagate a hard Go error, or surface the talk_to failure
		// (e.g. NPC not in view) verbatim.
		return res, err
	}

	// Wait for the banker's options menu (opcode 245) to surface.
	if !pollDialogOptions(ctx, h, bankDialogTimeout) {
		return interp.Fail(interp.DIALOG_NOT_OPEN,
			"bank.open: banker dialog menu never appeared after talk_to"), nil
	}

	// Find the "access my bank account" option and answer it.
	idx := findBankAccessOption(h)
	if idx < 0 {
		opts := "<none>"
		if rec := h.world.Recent.DialogOptions(); rec != nil {
			opts = strings.Join(rec.Options, " | ")
		}
		return interp.Fail(interp.SERVER_REJECTED,
			"bank.open: no bank-access option in banker menu; options were: "+opts), nil
	}
	if err := h.ChooseDialogOption(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	// The menu is consumed; clear it so any follow-up dialog waits
	// block on the NEXT menu rather than re-reading this one.
	h.world.Recent.ClearDialogOptions()

	// Wait for the bank window (opcode 42 → world.Bank.Open).
	if !pollBankOpen(ctx, h, bankOpenPollBudget) {
		return interp.Fail(interp.BANK_NOT_OPEN,
			"bank.open: bank window never opened after selecting the access option"), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// findBankAccessOption returns the 0-based index of the first banker
// dialog option matching any bankAccessNeedles substring, or -1 if no
// menu / no match. Case-insensitive.
func findBankAccessOption(h *Host) int {
	rec := h.world.Recent.DialogOptions()
	if rec == nil {
		return -1
	}
	for _, needle := range bankAccessNeedles {
		for i, opt := range rec.Options {
			if strings.Contains(strings.ToLower(opt), needle) {
				return i
			}
		}
	}
	return -1
}

// pollDialogOptions blocks until a dialog options menu is present or
// the timeout elapses. Returns true iff a menu surfaced. Mirrors
// dslWaitForDialog's cadence.
func pollDialogOptions(ctx context.Context, h *Host, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if h.world.Recent.DialogOptions() != nil {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(bankDialogPoll):
		}
	}
}

// pollBankOpen blocks until world.Bank.IsOpen() or the timeout
// elapses. Returns true iff the bank opened.
func pollBankOpen(ctx context.Context, h *Host, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if h.world.Bank.IsOpen() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(bankDialogPoll):
		}
	}
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
