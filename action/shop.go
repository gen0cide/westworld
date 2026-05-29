package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Shop outbound opcodes per Payload235Parser.java (case→OpcodeIn map)
// and the authentic client send sites in mudclient.java.
//
// The shop UI must already be open server-side (player.getShop() must
// be non-null). That state is set when the player opens a shop via NPC
// interaction. Without it, the InterfaceShopHandler flags the packet as
// suspicious and resets the shop.
const (
	outShopBuy   byte = 236 // [short catalogID][short shopStock][short amount]
	outShopSell  byte = 221 // [short catalogID][short shopStock][short amount]
	outShopClose byte = 166 // (empty)
)

// ShopBuy buys `amount` of catalogID from the open shop. shopStock is
// the quantity the client believes the shop currently holds for this
// item — the server uses it as a price-sync sanity hint (see the TODO
// in InterfaceShopHandler.process). Pass the value from
// world.Shop.Stock(catalogID); 0 is acceptable (the server clamps the
// buy to its real stock regardless).
//
// Wire format (mudclient.java#SHOP_BUY, opcode 236):
//
//	[short catalogID]
//	[short shopStock]
//	[short amount]
func ShopBuy(ctx context.Context, conn *session.Conn, catalogID, shopStock, amount int) error {
	return sendShopTrade(conn, outShopBuy, "ShopBuy", catalogID, shopStock, amount)
}

// ShopSell sells `amount` of catalogID to the open shop. Same wire
// shape as ShopBuy (mudclient.java#SHOP_SELL, opcode 221). The item
// must be in inventory and the shop must accept it (general store, or
// a specialty shop that stocks it).
func ShopSell(ctx context.Context, conn *session.Conn, catalogID, shopStock, amount int) error {
	return sendShopTrade(conn, outShopSell, "ShopSell", catalogID, shopStock, amount)
}

// ShopClose tells the server the bot is done shopping. Releases the
// server-side shop reference (player.resetShop()).
func ShopClose(ctx context.Context, conn *session.Conn) error {
	return conn.Send(outShopClose, nil)
}

// sendShopTrade builds the shared buy/sell packet shape.
func sendShopTrade(conn *session.Conn, opcode byte, name string, catalogID, shopStock, amount int) error {
	if catalogID < 0 || catalogID > 0xFFFF {
		return fmt.Errorf("action: %s catalogID %d out of uint16 range", name, catalogID)
	}
	if amount <= 0 {
		return fmt.Errorf("action: %s amount must be > 0 (got %d)", name, amount)
	}
	// shopStock is a client sync hint; clamp into range rather than
	// erroring (a stale 0 / over-large value is harmless server-side).
	if shopStock < 0 {
		shopStock = 0
	}
	if shopStock > 0xFFFF {
		shopStock = 0xFFFF
	}
	if amount > 0xFFFF {
		amount = 0xFFFF
	}
	buf := v235.NewBuffer(6)
	buf.WriteUint16(uint16(catalogID))
	buf.WriteUint16(uint16(shopStock))
	buf.WriteUint16(uint16(amount))
	return conn.Send(opcode, buf.Bytes())
}
