package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/world"
)

// InitTradeRequest sends a trade request to the player at serverIndex.
// Trade requires standing adjacent to the target, so we pathfind first.
// Updates world.Trade to "request_sent" so routines can poll phase.
func (h *Host) InitTradeRequest(ctx context.Context, serverIndex int) error {
	name := ""
	if rec, ok := h.world.Players.Get(serverIndex); ok {
		name = rec.Name
	}
	h.world.Trade.BeginRequest(serverIndex, name)
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.InitTradeRequest(ctx, h.conn, serverIndex)
	}
	h.log.Info("InitTradeRequest: pathfinding",
		"server_index", serverIndex,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.InitTradeRequest(ctx, h.conn, serverIndex)
	})
}

// AcceptIncomingTrade accepts a pending trade request.
func (h *Host) AcceptIncomingTrade(ctx context.Context) error {
	return action.AcceptIncomingTrade(ctx, h.conn)
}

// DeclineTrade declines a pending or active trade. Marks the
// trade closed (not-completed) immediately so routine polls
// reflect the cancellation.
func (h *Host) DeclineTrade(ctx context.Context) error {
	h.world.Trade.MarkClosed(false)
	return action.DeclineTrade(ctx, h.conn)
}

// OfferTradeItems sets our offered items in the current trade.
// Takes world.TradeItem and converts to action.TradeItem for the
// packet send. Updates world state on successful dispatch.
func (h *Host) OfferTradeItems(ctx context.Context, items []world.TradeItem) error {
	actItems := make([]action.TradeItem, len(items))
	for i, it := range items {
		actItems[i] = action.TradeItem{CatalogID: it.ItemID, Amount: it.Amount}
	}
	if err := action.OfferTradeItems(ctx, h.conn, actItems); err != nil {
		return err
	}
	h.world.Trade.SetMyOffer(items)
	return nil
}

// ConfirmTrade clicks Accept on the trade window. First call is
// the offer-screen accept; second call is the final-confirm-screen
// accept. World state tracks which step we're on.
func (h *Host) ConfirmTrade(ctx context.Context) error {
	t := h.world.Trade.Trade()
	if t == nil {
		return fmt.Errorf("ConfirmTrade: no active trade")
	}
	if !t.MyFirstAccepted {
		h.world.Trade.MarkMyFirstAccepted()
	} else if !t.MySecondAccepted {
		h.world.Trade.MarkMySecondAccepted()
	}
	return action.ConfirmTrade(ctx, h.conn)
}
