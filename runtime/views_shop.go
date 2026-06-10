package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
)

// View for the `shop` faculty — world.shop.* state. Reached via
// worldView.Get("shop") (mirrors world.bank → bankView). The wiring
// line in runtime/views_world.go is:
//
//	case "shop":
//	    return &shopView{host: w.host}, true
//
// (See the registration note — that one line is the only edit needed in
// views_world.go to surface this view.)
//
// shopView surfaces world.Shop.* to routines:
//
//	world.shop.is_open        → Bool, true while a shop window is open
//	world.shop.stock(item)    → Int, quantity in stock (0 if absent/closed)
//	world.shop.price(item)    → Int, unit BUY price in gp (0 if absent/closed)
//	world.shop.slots          → list of {item_id, stock, base_stock}
//	world.shop.is_general     → Bool, true for a general store
type shopView struct {
	host *Host
	bind *routineBinding
}

func (s *shopView) Kind() string    { return "shop" }
func (s *shopView) Display() string { return "<shop>" }

func (s *shopView) Get(field string) (interp.Value, bool) {
	rec := s.host.world.Shop.Shop()
	switch field {
	case "is_open":
		return interp.Bool(rec != nil), true
	case "is_general":
		return interp.Bool(rec != nil && rec.IsGeneral), true
	case "slots":
		if rec == nil {
			return &interp.List{Items: []interp.Value{}}, true
		}
		items := make([]interp.Value, 0, len(rec.Slots))
		for _, sl := range rec.Slots {
			triple := &interp.List{Items: []interp.Value{
				interp.Int(int64(sl.ItemID)),
				interp.Int(int64(sl.Stock)),
				interp.Int(int64(sl.BaseStock)),
			}}
			items = append(items, triple)
		}
		return &interp.List{Items: items}, true
	case "stock":
		return &shopStockCallable{host: s.host}, true
	case "price":
		return &shopPriceCallable{host: s.host}, true

	// Action methods on the view, so the api.md §8 dotted forms
	// `shop.buy(...)` / `shop.sell(...)` / `shop.close()` work as
	// member calls (recv.method). The same handlers are ALSO registered
	// as flat builtins shop_buy/shop_sell/shop_close in shop_register.go
	// for the pre-rename flat-name idiom (mirrors bank deposit/withdraw).
	case "buy":
		return s.actionCallable("buy", dslShopBuy, true), true
	case "sell":
		return s.actionCallable("sell", dslShopSell, true), true
	case "close":
		return s.actionCallable("close", dslShopClose, false), true
	}
	return nil, false
}

// actionCallable wraps a shop handler body as a yielding action
// callable bound to this view's Host and interpreter routine ctx,
// reusing the runtime's standard actionCallable shape.
func (s *shopView) actionCallable(name string, fn actionHandler, _ bool) interp.Value {
	return &actionCallable{name: "shop." + name, host: s.host, ctx: s.bind.context(), fn: fn}
}

// shopStockCallable backs world.shop.stock(item) → Int. Accepts an
// ItemRef (id, name, or item-view) like the action verbs.
type shopStockCallable struct{ host *Host }

func (c *shopStockCallable) Kind() string    { return "callable" }
func (c *shopStockCallable) Display() string { return "<world.shop.stock>" }
func (c *shopStockCallable) Yields() bool    { return false }

func (c *shopStockCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("world.shop.stock takes 1 argument (item)")
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	return interp.Int(int64(c.host.world.Shop.Stock(id))), nil
}

// shopPriceCallable backs world.shop.price(item) → Int (unit buy price,
// gp). The gp price needs the item's catalogue base price, which the
// shop packet does not carry — we resolve it from facts.ItemDef and
// apply the authentic client price formula in world.ShopState.BuyPrice.
// Returns 0 if the shop is closed, the item isn't stocked, or its def
// is unknown.
type shopPriceCallable struct{ host *Host }

func (c *shopPriceCallable) Kind() string    { return "callable" }
func (c *shopPriceCallable) Display() string { return "<world.shop.price>" }
func (c *shopPriceCallable) Yields() bool    { return false }

func (c *shopPriceCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("world.shop.price takes 1 argument (item)")
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	basePrice := 0
	if c.host.facts != nil {
		if def := c.host.facts.ItemDef(id); def != nil {
			basePrice = def.BasePrice
		}
	}
	return interp.Int(int64(c.host.world.Shop.BuyPrice(id, basePrice))), nil
}
