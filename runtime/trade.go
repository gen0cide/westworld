package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
)

// InitTradeRequest sends a trade request to the player at serverIndex.
// Trade requires standing adjacent to the target, so we pathfind first.
func (h *Host) InitTradeRequest(ctx context.Context, serverIndex int) error {
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

// DeclineTrade declines a pending or active trade.
func (h *Host) DeclineTrade(ctx context.Context) error {
	return action.DeclineTrade(ctx, h.conn)
}

// OfferTradeItems sets our offered items in the current trade.
func (h *Host) OfferTradeItems(ctx context.Context, items []action.TradeItem) error {
	return action.OfferTradeItems(ctx, h.conn, items)
}

// ConfirmTrade clicks Accept on the trade window.
func (h *Host) ConfirmTrade(ctx context.Context) error {
	return action.ConfirmTrade(ctx, h.conn)
}
