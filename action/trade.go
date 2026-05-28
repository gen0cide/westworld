package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Trade outbound opcodes per Payload235Parser.java.
const (
	outInitTradeRequest    byte = 142 // [short serverIndex]
	outAcceptInitTrade     byte = 55  // (empty) — accept the trade request
	outDeclineTrade        byte = 230 // (empty)
	outAddTradeItems       byte = 46  // [byte count] then per-item [short id, int amount]
	outConfirmTrade        byte = 104 // (empty) — first or second confirmation
)

// InitTradeRequest sends a trade request to the player at serverIndex.
// The other player must accept (or decline) via their client.
func InitTradeRequest(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outInitTradeRequest, serverIndex, "InitTradeRequest")
}

// AcceptIncomingTrade accepts a trade request that another player
// initiated. Server then sends SEND_TRADE_OPEN_CONFIRM to both sides.
func AcceptIncomingTrade(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outAcceptInitTrade, nil)
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

// ConfirmTrade clicks the "Accept" button on the trade window. RSC
// trades require BOTH sides to click accept TWICE — once on the
// initial item-offer screen, then again on the confirmation screen.
func ConfirmTrade(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outConfirmTrade, nil)
}
