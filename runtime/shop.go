package runtime

import (
	"context"

	"github.com/gen0cide/westworld/action"
)

// Host boundary methods for the shop faculty — mirror runtime/bank.go.
// These thread the outbound action.Shop* packet builders through the
// Host's connection. The shop must already be open server-side (the
// bot opened it via NPC interaction); world.Shop mirrors that state
// from the inbound SEND_SHOP_OPEN packet.
//
// RSC has no client-driven shop-update packet: the server re-sends the
// full catalogue (SEND_SHOP_OPEN) after every buy/sell, so the world
// mirror refreshes itself — these methods don't mutate world.Shop
// directly (unlike the trade/bank optimistic-update pattern).

// ShopBuy buys `amount` of catalogID from the open shop. The current
// believed stock is read from world.Shop and sent as the server's
// price-sync hint.
func (h *Host) ShopBuy(ctx context.Context, catalogID, amount int) error {
	stock := h.world.Shop.Stock(catalogID)
	return action.ShopBuy(ctx, h.conn, catalogID, stock, amount)
}

// ShopSell sells `amount` of catalogID to the open shop.
func (h *Host) ShopSell(ctx context.Context, catalogID, amount int) error {
	stock := h.world.Shop.Stock(catalogID)
	return action.ShopSell(ctx, h.conn, catalogID, stock, amount)
}

// ShopClose closes the shop window and clears the local mirror
// (the server also sends SEND_SHOP_CLOSE, but we clear eagerly so a
// routine polling world.shop.is_open sees the close immediately).
func (h *Host) ShopClose(ctx context.Context) error {
	if err := action.ShopClose(ctx, h.conn); err != nil {
		return err
	}
	h.world.Shop.Close()
	return nil
}
