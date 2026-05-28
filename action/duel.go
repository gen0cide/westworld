package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Duel outbound opcodes per Payload235Parser.java.
const (
	outInitDuelRequest      byte = 103 // [short serverIndex]
	outDuelFirstAccepted    byte = 176 // (empty) — first accept on offer screen
	outDuelDeclined         byte = 197 // (empty) — abort the duel at any phase
	outDuelOfferItem        byte = 33  // [byte count] then per-item [short id, int amount]
	outDuelSecondAccepted   byte = 77  // (empty) — final accept on confirm screen
	outDuelSettingsChanged  byte = 8   // [4 × byte] disallow retreat/magic/prayer/weapons
)

// InitDuelRequest sends a duel request to the player at serverIndex.
// The other player accepts (or declines) via their client; on accept
// both sides see SEND_DUEL_WINDOW.
func InitDuelRequest(ctx context.Context, conn *session.Conn, serverIndex int) error {
	return sendShortPacket(ctx, conn, outInitDuelRequest, serverIndex, "InitDuelRequest")
}

// AcceptDuelFirst is the first accept-click — on the offer screen
// (stake items + rules visible). Both sides must do this before the
// confirm screen appears.
func AcceptDuelFirst(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outDuelFirstAccepted, nil)
}

// AcceptDuelSecond is the final accept-click — on the confirm
// screen. Both sides must do this for the fight to begin.
func AcceptDuelSecond(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outDuelSecondAccepted, nil)
}

// DeclineDuel cancels the duel at any phase (request, offer,
// confirm). Closes the window for both sides.
func DeclineDuel(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outDuelDeclined, nil)
}

// DuelRules are the four pre-fight toggles. true = disallow.
type DuelRules struct {
	DisallowRetreat bool
	DisallowMagic   bool
	DisallowPrayer  bool
	DisallowWeapons bool
}

// SetDuelRules sends DUEL_FIRST_SETTINGS_CHANGED with the four
// disallow flags. Either side can change rules from the offer
// screen; the server unifies the result and broadcasts it back via
// SEND_DUEL_SETTINGS. Changing rules resets both first-accept flags.
func SetDuelRules(ctx context.Context, conn *session.Conn, r DuelRules) error {
	buf := v235.NewBuffer(4)
	buf.WriteByte(boolByte(r.DisallowRetreat))
	buf.WriteByte(boolByte(r.DisallowMagic))
	buf.WriteByte(boolByte(r.DisallowPrayer))
	buf.WriteByte(boolByte(r.DisallowWeapons))
	return conn.Send(outDuelSettingsChanged, buf.Bytes())
}

// OfferDuelItems sets the items WE are staking in the current duel.
// Replaces any prior offer.
//
// Source: Payload235Parser.java:503-512 (DUEL_OFFER_ITEM):
//   [byte count]
//   for each: [short catalogID] [int amount]
func OfferDuelItems(ctx context.Context, conn *session.Conn, items []TradeItem) error {
	if len(items) > 0xFF {
		return fmt.Errorf("action: duel item count %d exceeds 255", len(items))
	}
	buf := v235.NewBuffer(1 + len(items)*6)
	buf.WriteByte(byte(len(items)))
	for _, it := range items {
		if it.CatalogID < 0 || it.CatalogID > 0xFFFF {
			return fmt.Errorf("action: duel item catalogID %d out of range", it.CatalogID)
		}
		buf.WriteUint16(uint16(it.CatalogID))
		buf.WriteUint32(uint32(it.Amount))
	}
	return conn.Send(outDuelOfferItem, buf.Bytes())
}

func boolByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
