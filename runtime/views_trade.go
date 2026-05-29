package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// View for the `trade` faculty — world.trade.* handshake state.
// Reached via worldView.Get("trade") (see views_world.go).

// tradeView surfaces world.Trade.* to routines:
//   world.trade.is_active             → bool
//   world.trade.phase                 → "request_sent" / "open" / "confirm" / "completed" / "cancelled"
//   world.trade.with                  → opponent name (String) or Null
//   world.trade.my_offer              → list of {item_id, amount}
//   world.trade.their_offer           → list
//   world.trade.my_first_accepted     → bool (we clicked first Accept)
//   world.trade.their_first_accepted  → bool
//   world.trade.my_second_accepted    → bool
//   world.trade.their_second_accepted → bool
//   world.trade.both_first_accepted   → both clicked first Accept (ready for confirm screen)
//   world.trade.both_second_accepted  → both clicked second Accept (trade about to complete)
type tradeView struct{ host *Host }

func (t *tradeView) Kind() string    { return "trade" }
func (t *tradeView) Display() string { return "<trade>" }

func (t *tradeView) Get(field string) (interp.Value, bool) {
	// Action verbs (request/offer/accept/confirm/decline + bang) first.
	if v, ok := t.host.namespaceAction("trade", field, tradeVerbs); ok {
		return v, true
	}
	rec := t.host.world.Trade.Trade()
	switch field {
	case "is_active":
		return interp.Bool(rec != nil && rec.Phase != "completed" && rec.Phase != "cancelled"), true
	case "phase":
		if rec == nil {
			return interp.String("none"), true
		}
		return interp.String(rec.Phase), true
	case "with":
		if rec == nil {
			return interp.Null{}, true
		}
		return interp.String(rec.WithName), true
	case "with_index":
		if rec == nil {
			return interp.Int(0), true
		}
		return interp.Int(int64(rec.WithIndex)), true
	case "my_offer":
		return tradeOfferList(rec, true), true
	case "their_offer":
		return tradeOfferList(rec, false), true
	// Offer-screen accepts. Frozen names (§7/§10): accepted /
	// they_accepted / both_accepted. The my/their/both_first_accepted
	// names are kept for back-compat with world.trade.*.
	case "accepted", "my_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted), true
	case "they_accepted", "their_first_accepted":
		return interp.Bool(rec != nil && rec.TheirFirstAccepted), true
	case "both_accepted", "both_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted && rec.TheirFirstAccepted), true
	// Confirm-screen accepts. Frozen names: confirmed / they_confirmed
	// / both_confirmed.
	case "confirmed", "my_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted), true
	case "they_confirmed", "their_second_accepted":
		return interp.Bool(rec != nil && rec.TheirSecondAccepted), true
	case "both_confirmed", "both_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted && rec.TheirSecondAccepted), true
	}
	return nil, false
}

// tradeOfferList returns one side's offer as a List of {item_id, amount}.
// Empty list if no trade or that side hasn't offered yet.
func tradeOfferList(rec *world.TradeRecord, mine bool) interp.Value {
	if rec == nil {
		return &interp.List{Items: []interp.Value{}}
	}
	src := rec.TheirOffer
	if mine {
		src = rec.MyOffer
	}
	items := make([]interp.Value, 0, len(src))
	for _, it := range src {
		pair := &interp.List{Items: []interp.Value{
			interp.Int(int64(it.ItemID)),
			interp.Int(int64(it.Amount)),
		}}
		items = append(items, pair)
	}
	return &interp.List{Items: items}
}
