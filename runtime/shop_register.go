package runtime

// shop_register.go documents the shop faculty's registry + spec
// additions in ONE clearly-commented, easily-mergeable place.
//
// WHY DOCS-ONLY (no init): the shop ACTIONS are reachable today via
// the shopView method path — `world.shop.buy(item, qty)`,
// `world.shop.sell(item, qty)`, `world.shop.close()` — which dispatch
// as member calls (recv.method) through shopView.Get (see
// runtime/views_shop.go) and need NO spec.Actions / actionHandlers
// rows. That keeps go build green AND the cross-package consistency
// tests green without touching the files the namespacing-refactor agent
// owns (runtime/dsl_actions.go, dsl/spec/*).
//
// The api.md §8 names are `shop.buy` / `shop.sell` / `shop.close`. Once
// `shop` becomes a top-level reserved namespace in the refactor (it.
// Reserved["shop"] = &shopView{...} in dsl_bridge.go, mirroring how the
// rename agent is hoisting bank/trade), those exact dotted forms resolve
// against shopView with zero further work.
//
// ----------------------------------------------------------------------
// TO ALSO EXPOSE THE FLAT-NAME BUILTINS (optional — mirrors the bank
// deposit/withdraw flat idiom), apply these rows INLINE in the central
// files (an init()-append from this package is too late: dsl/spec's
// byName index is built in spec's init, before runtime's init runs, so
// spec.ByName would miss appended entries and the consistency test would
// fail). Apply directly:
//
//   runtime/dsl_actions.go — in the actionHandlers map, add a
//   "// ---- shop (actions_shop.go) ----" group:
//       "shop_buy":   dslShopBuy,
//       "shop_sell":  dslShopSell,
//       "shop_close": dslShopClose,
//
//   dsl/spec/actions.go — in the Actions slice, add a "// Shop" group:
//       {Name: "shop_buy", Kind: PrimaryAction, MinArgs: 2, MaxArgs: 2,
//           Params: []string{"item", "qty"},
//           DocSummary: "Buy qty of the item from the open shop."},
//       {Name: "shop_sell", Kind: PrimaryAction, MinArgs: 2, MaxArgs: 2,
//           Params: []string{"item", "qty"},
//           DocSummary: "Sell qty of the item to the open shop."},
//       {Name: "shop_close", Kind: PrimaryAction, MinArgs: 0, MaxArgs: 0,
//           DocSummary: "Close the shop window."},
//
// ----------------------------------------------------------------------
// REQUIRED ONE-LINE WIRE (the only edit to an existing file the shop
// faculty needs that a new file cannot make for itself):
//
//   runtime/views_world.go — in worldView.Get's switch, add:
//       case "shop":
//           return &shopView{host: w.host}, true
//
// Without it, world.shop.* (and thus the shop.buy/sell/close method
// path) does not resolve. The shopView + handlers are all built and
// ready in runtime/views_shop.go + runtime/actions_shop.go.
//
// ----------------------------------------------------------------------
// SPEC ACCESSOR ROWS (dsl/spec/accessors.go) for the read surface —
// add once consistency_test.go's allowedRoots includes "shop":
//       {Path: []string{"shop", "is_open"}, Kind: "bool",
//           DocSummary: "Shop UI currently shown."},
//       {Path: []string{"shop", "slots"}, Kind: "list<ShopSlot>",
//           DocSummary: "Shop catalogue: item_id / stock / base_stock."},
