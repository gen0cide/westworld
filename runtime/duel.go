package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/world"
)

// InitDuelRequest sends a duel request to the player at serverIndex.
// Same adjacency requirement as trade — we pathfind first.
// Updates world.Duel to "request_sent" so routines can poll phase.
func (h *Host) InitDuelRequest(ctx context.Context, serverIndex int) error {
	name := ""
	if rec, ok := h.world.Players.Get(serverIndex); ok {
		name = rec.Name
	}
	h.world.Duel.BeginRequest(serverIndex, name)
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.InitDuelRequest(ctx, h.conn, serverIndex)
	}
	h.log.Info("InitDuelRequest: pathfinding",
		"server_index", serverIndex,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.InitDuelRequest(ctx, h.conn, serverIndex)
	})
}

// AcceptIncomingDuel accepts a pending duel request. RSC duels use
// the same accept path as the initial trade-style request — the
// client sends the same outbound that the initiator did
// (PLAYER_DUEL) but with the originator's index. From the
// responder's point of view the handshake is symmetric.
//
// In practice, receiving a duel request opens a confirm dialog on
// the client and an "accept" closes it; we model it by re-sending
// the request packet pointing at the requester. World state is
// seeded by the same BeginRequest path the initiator uses.
func (h *Host) AcceptIncomingDuel(ctx context.Context, fromIndex int) error {
	name := ""
	if rec, ok := h.world.Players.Get(fromIndex); ok {
		name = rec.Name
	}
	h.world.Duel.BeginRequest(fromIndex, name)
	return action.InitDuelRequest(ctx, h.conn, fromIndex)
}

// DeclineDuel cancels the duel at any phase. Marks the duel closed
// immediately so polls reflect the cancellation.
func (h *Host) DeclineDuel(ctx context.Context) error {
	h.world.Duel.MarkClosed(false)
	return action.DeclineDuel(ctx, h.conn)
}

// OfferDuelItems sets our staked items. Converts world.TradeItem to
// action.TradeItem (same shape as trade) and updates world state.
func (h *Host) OfferDuelItems(ctx context.Context, items []world.TradeItem) error {
	actItems := make([]action.TradeItem, len(items))
	for i, it := range items {
		actItems[i] = action.TradeItem{CatalogID: it.ItemID, Amount: it.Amount}
	}
	if err := action.OfferDuelItems(ctx, h.conn, actItems); err != nil {
		return err
	}
	h.world.Duel.SetMyOffer(items)
	return nil
}

// SetDuelRules sends the 4 rule toggles. The server unifies with
// the opponent's last-set rules and broadcasts the result. Resets
// both first-accept flags (server rule).
func (h *Host) SetDuelRules(ctx context.Context, r world.DuelRules) error {
	if err := action.SetDuelRules(ctx, h.conn, action.DuelRules(r)); err != nil {
		return err
	}
	h.world.Duel.SetRules(r)
	return nil
}

// AcceptDuelOffer is the first accept-click (offer screen). Sends
// DUEL_FIRST_ACCEPTED and updates world state. If the opponent has
// already first-accepted, the world transitions to "confirm" phase
// (server will follow up with SEND_DUEL_CONFIRMWINDOW).
func (h *Host) AcceptDuelOffer(ctx context.Context) error {
	d := h.world.Duel.Duel()
	if d == nil {
		return fmt.Errorf("AcceptDuelOffer: no active duel")
	}
	if err := action.AcceptDuelFirst(ctx, h.conn); err != nil {
		return err
	}
	h.world.Duel.MarkMyFirstAccepted()
	return nil
}

// AcceptDuelConfirm is the second accept-click (confirm screen).
// Sends DUEL_SECOND_ACCEPTED and updates world state.
func (h *Host) AcceptDuelConfirm(ctx context.Context) error {
	d := h.world.Duel.Duel()
	if d == nil {
		return fmt.Errorf("AcceptDuelConfirm: no active duel")
	}
	if err := action.AcceptDuelSecond(ctx, h.conn); err != nil {
		return err
	}
	h.world.Duel.MarkMySecondAccepted()
	return nil
}
