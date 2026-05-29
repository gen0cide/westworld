package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Shop action handler bodies (shop.buy / shop.sell / shop.close).
// Parallel structure to actions_bank.go. These are reached today via
// the shopView method path (world.shop.buy/sell/close — see
// runtime/views_shop.go); the exact flat-builtin registry/spec rows to
// fold into the central tables post-refactor are documented in
// runtime/shop_register.go (kept out of dsl_actions.go + dsl/spec/* to
// avoid colliding with the concurrent namespacing refactor).
//
// Item+amount resolution reuses resolveItemAmount (dsl_helpers.go), the
// same helper bank deposit/withdraw use.

// dslShopBuy implements shop.buy(item, qty) → Result<Null>.
// Errors: SHOP_NOT_OPEN (no shop window); SERVER_REJECTED (out of
// stock, members-only item for a free player, insufficient coins —
// the server signals these with prose, not a typed code).
func dslShopBuy(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if !h.world.Shop.IsOpen() {
		return interp.Fail(interp.SHOP_NOT_OPEN, "no shop window open"), nil
	}
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.ShopBuy(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslShopSell implements shop.sell(item, qty) → Result<Null>.
// Errors: SHOP_NOT_OPEN; NO_SUCH_ITEM if inventory lacks the item;
// SERVER_REJECTED (shop won't buy it).
func dslShopSell(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if !h.world.Shop.IsOpen() {
		return interp.Fail(interp.SHOP_NOT_OPEN, "no shop window open"), nil
	}
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if h.world.Inventory.Count(itemID) <= 0 {
		return interp.Fail(interp.NO_SUCH_ITEM, "item not in inventory"), nil
	}
	if err := h.ShopSell(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslShopClose implements shop.close() → Result<Null>.
func dslShopClose(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.ShopClose(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
