package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// View for the `duel` faculty — world.duel.* state. Reached via
// worldView.Get("duel") (see views_world.go).

// duelView surfaces world.Duel.* to routines. Mirrors tradeView
// with additional fields for the rule toggles:
//
//	world.duel.is_active             → bool
//	world.duel.phase                 → string (request_sent / open / confirm / completed / cancelled / none)
//	world.duel.with                  → opponent name or Null
//	world.duel.with_index            → opponent server-index Int
//	world.duel.my_offer / their_offer → list of [item_id, amount]
//	world.duel.my_first_accepted / their_first_accepted / etc.
//	world.duel.disallow_retreat / _magic / _prayer / _weapons → bool
type duelView struct{ host *Host }

func (d *duelView) Kind() string    { return "duel" }
func (d *duelView) Display() string { return "<duel>" }

func (d *duelView) Get(field string) (interp.Value, bool) {
	// Action verbs (request/set_rules/stake/accept/confirm/decline +
	// bang) first.
	if v, ok := d.host.namespaceAction("duel", field, duelVerbs); ok {
		return v, true
	}
	rec := d.host.world.Duel.Duel()
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
		return duelOfferList(rec, true), true
	case "their_offer":
		return duelOfferList(rec, false), true
	case "accepted", "my_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted), true
	case "they_accepted", "their_first_accepted":
		return interp.Bool(rec != nil && rec.TheirFirstAccepted), true
	case "both_accepted", "both_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted && rec.TheirFirstAccepted), true
	case "confirmed", "my_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted), true
	case "they_confirmed", "their_second_accepted":
		return interp.Bool(rec != nil && rec.TheirSecondAccepted), true
	case "both_confirmed", "both_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted && rec.TheirSecondAccepted), true
	case "disallow_retreat":
		return interp.Bool(rec != nil && rec.Rules.DisallowRetreat), true
	case "disallow_magic":
		return interp.Bool(rec != nil && rec.Rules.DisallowMagic), true
	case "disallow_prayer":
		return interp.Bool(rec != nil && rec.Rules.DisallowPrayer), true
	case "disallow_weapons":
		return interp.Bool(rec != nil && rec.Rules.DisallowWeapons), true
	}
	return nil, false
}

// duelOfferList returns one side's staked items, same format as
// tradeOfferList.
func duelOfferList(rec *world.DuelRecord, mine bool) interp.Value {
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
