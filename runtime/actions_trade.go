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

// dslTradeRequest sends a trade request to a player. Accepts a
// player-view or a server-index Int. Walks adjacent first
// (server requires adjacency).
func dslTradeRequest(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("trade_request takes 1 arg (player), got %d", len(args))
	}
	var idx int
	switch v := args[0].(type) {
	case *playerView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(args[0]); ok {
			idx = int(i)
		} else {
			return nil, errf("trade_request: target must be a player view or Int index, got %s", args[0].Kind())
		}
	}
	if err := h.InitTradeRequest(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAcceptTrade accepts an incoming trade request by re-sending the
// trade-request packet to the original requester (OpenRSC's symmetric
// handshake — both sides must request each other for the window to
// open). Takes a player view, server-index Int, or player-name String
// (string lookups resolve via world.Players).
func dslAcceptTrade(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("accept_trade takes 1 arg (requester player/name/index), got %d", len(args))
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	if err := h.AcceptIncomingTrade(ctx, idx); err != nil {
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
