package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Trade action handler bodies — request / accept / offer / confirm /
// finalize / decline. Registered in the central actionHandlers table
// in dsl_actions.go. Player-index resolution lives in dsl_helpers.go
// (resolvePlayerIndex).

// dslTradeRequest sends a trade request to a player. Backs the frozen
// trade.request(p) (§10), which ABSORBS the old trade_request +
// accept_trade: clicking a player to trade *is* both initiating and
// accepting an incoming request (same opcode 142), and mutual requests
// open the window. Accepts a player-view, server-index Int, or player-
// name String. Walks adjacent first when the player is visible.
func dslTradeRequest(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("trade.request takes 1 arg (player), got %d", len(args))
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	if err := h.InitTradeRequest(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslDeclineTrade(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.DeclineTrade(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOfferTrade takes a list of [item_id, amount] pairs (Int, Int)
// and sends a trade offer. Replaces any prior offer.
func dslOfferTrade(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("offer_trade takes 1 arg (list of [item_id, amount]), got %d", len(args))
	}
	list, ok := args[0].(*interp.List)
	if !ok {
		return nil, errf("offer_trade: arg must be a list, got %s", args[0].Kind())
	}
	items := make([]world.TradeItem, 0, len(list.Items))
	for i, el := range list.Items {
		pair, ok := el.(*interp.List)
		if !ok || len(pair.Items) != 2 {
			return nil, errf("offer_trade: element %d must be [item_id, amount], got %s", i, el.Kind())
		}
		id, idok := pair.Items[0].(interp.Int)
		amt, amtok := pair.Items[1].(interp.Int)
		if !idok || !amtok {
			return nil, errf("offer_trade: element %d fields must be Int", i)
		}
		items = append(items, world.TradeItem{ItemID: int(id), Amount: int(amt)})
	}
	if err := h.OfferTradeItems(ctx, items); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslConfirmTrade(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.ConfirmTrade(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslFinalizeTrade clicks "Confirm" on the second trade screen. Pair with
// confirm_trade() (first screen): confirm_trade() then finalize_trade().
func dslFinalizeTrade(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.FinalizeTrade(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
