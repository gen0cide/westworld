package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Trade outbound opcodes per Payload235Parser.java.
//
// RSC trades use a two-screen accept flow like duels:
//   1. After offer items are placed, both sides click "Accept" on
//      the OFFER screen — opcode 55 PLAYER_ACCEPTED_INIT_TRADE_REQUEST.
//   2. Server transitions both to the FINAL CONFIRM screen.
//   3. Both sides click "Accept" again — opcode 104
//      PLAYER_ACCEPTED_TRADE.
//
// The original `accept_trade` for an incoming request is NOT one
// of these — it's a re-send of opcode 142 PLAYER_INIT_TRADE_REQUEST
// pointing at the original requester (symmetric handshake).
const (
	outInitTradeRequest  byte = 142 // [short serverIndex]
	outAcceptTradeOffer  byte = 55  // (empty) — first accept on offer screen
	outDeclineTrade      byte = 230 // (empty)
	outAddTradeItems     byte = 46  // [byte count] then per-item [short id, int amount]
	outAcceptTradeConfirm byte = 104 // (empty) — second accept on confirm screen
)

// InitTradeRequest sends a trade request to the player at serverIndex.
// The other player must accept (or decline) via their client.
func InitTradeRequest(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outInitTradeRequest, serverIndex, "InitTradeRequest")
}

// AcceptTradeOffer is the FIRST accept-click — on the offer screen.
// Both sides must do this before the confirm screen appears.
func AcceptTradeOffer(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outAcceptTradeOffer, nil)
}

// DeclineTrade declines a pending trade (either before or during the
// item-offering phase).
func DeclineTrade(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outDeclineTrade, nil)
}

// TradeItem is one item being put up for trade.
type TradeItem struct {
	CatalogID int // item type id (joins to facts.ItemDef.ID)
	Amount    int
}

// OfferTradeItems sets the items we're offering in the current trade.
// Replaces any prior offer (the server uses the latest list).
//
// Source: Payload235Parser.java:530-540 (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER):
//   [byte count]
//   for each: [short catalogID] [int amount]
func OfferTradeItems(ctx context.Context, conn *session.Conn, items []TradeItem) error {
	if len(items) > 0xFF {
		return fmt.Errorf("action: trade item count %d exceeds 255", len(items))
	}
	buf := v235.NewBuffer(1 + len(items)*6)
	buf.WriteByte(byte(len(items)))
	for _, it := range items {
		if it.CatalogID < 0 || it.CatalogID > 0xFFFF {
			return fmt.Errorf("action: trade item catalogID %d out of range", it.CatalogID)
		}
		buf.WriteUint16(uint16(it.CatalogID))
		buf.WriteUint32(uint32(it.Amount))
	}
	return conn.Send(outAddTradeItems, buf.Bytes())
}

// AcceptTradeConfirm is the SECOND accept-click — on the final
// confirm screen. Both sides must do this for the trade to complete.
func AcceptTradeConfirm(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outAcceptTradeConfirm, nil)
}
