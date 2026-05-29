package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Duel action handler bodies — parallel structure to trade. Registered
// in the central actionHandlers table in dsl_actions.go. Player-index
// resolution lives in dsl_helpers.go (resolvePlayerIndex).

// dslDuelRequest initiates a duel with a player or server-index.
// Walks adjacent first (server requires adjacency).
func dslDuelRequest(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("duel_request takes 1 arg (player), got %d", len(args))
	}
	var idx int
	switch v := args[0].(type) {
	case *playerView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(args[0]); ok {
			idx = int(i)
		} else {
			return nil, errf("duel_request: target must be a player view or Int index, got %s", args[0].Kind())
		}
	}
	if err := h.InitDuelRequest(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAcceptDuel accepts an incoming duel request by re-sending the
// duel-request packet to the original requester (symmetric handshake
// like trade). Takes a player view, server-index Int, or player-name
// String. The duel_request event delivers the requester's name, so
// the typical call is `accept_duel(name)`.
func dslAcceptDuel(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("accept_duel takes 1 arg (requester player/name/index), got %d", len(args))
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	if err := h.AcceptIncomingDuel(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslDeclineDuel(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.DeclineDuel(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOfferDuel takes a list of [item_id, amount] pairs and stakes
// them. Same shape as offer_trade.
func dslOfferDuel(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("offer_duel takes 1 arg (list of [item_id, amount]), got %d", len(args))
	}
	list, ok := args[0].(*interp.List)
	if !ok {
		return nil, errf("offer_duel: arg must be a list, got %s", args[0].Kind())
	}
	items := make([]world.TradeItem, 0, len(list.Items))
	for i, el := range list.Items {
		pair, ok := el.(*interp.List)
		if !ok || len(pair.Items) != 2 {
			return nil, errf("offer_duel: element %d must be [item_id, amount], got %s", i, el.Kind())
		}
		id, idok := pair.Items[0].(interp.Int)
		amt, amtok := pair.Items[1].(interp.Int)
		if !idok || !amtok {
			return nil, errf("offer_duel: element %d fields must be Int", i)
		}
		items = append(items, world.TradeItem{ItemID: int(id), Amount: int(amt)})
	}
	if err := h.OfferDuelItems(ctx, items); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslSetDuelRules takes a 4-element list of bools or kwargs and
// sends the rule toggles. Accepts either:
//   set_duel_rules([true, false, false, true])    # retreat, magic, prayer, weapons
//   set_duel_rules(retreat=true, weapons=true)
func dslSetDuelRules(ctx context.Context, h *Host, args []interp.Value, kwargs map[string]interp.Value) (interp.Value, error) {
	var r world.DuelRules
	if len(args) == 1 {
		list, ok := args[0].(*interp.List)
		if !ok || len(list.Items) != 4 {
			return nil, errf("set_duel_rules: positional arg must be a 4-element list [retreat, magic, prayer, weapons]")
		}
		flags := make([]bool, 4)
		for i, el := range list.Items {
			b, bok := el.(interp.Bool)
			if !bok {
				return nil, errf("set_duel_rules: list element %d must be Bool", i)
			}
			flags[i] = bool(b)
		}
		r = world.DuelRules{
			DisallowRetreat: flags[0],
			DisallowMagic:   flags[1],
			DisallowPrayer:  flags[2],
			DisallowWeapons: flags[3],
		}
	} else if len(args) == 0 {
		// kwargs mode
		readFlag := func(name string) bool {
			v, ok := kwargs[name]
			if !ok {
				return false
			}
			b, _ := v.(interp.Bool)
			return bool(b)
		}
		r = world.DuelRules{
			DisallowRetreat: readFlag("retreat"),
			DisallowMagic:   readFlag("magic"),
			DisallowPrayer:  readFlag("prayer"),
			DisallowWeapons: readFlag("weapons"),
		}
	} else {
		return nil, errf("set_duel_rules takes either a 4-element list or named kwargs (retreat/magic/prayer/weapons)")
	}
	if err := h.SetDuelRules(ctx, r); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslAcceptDuelOffer(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.AcceptDuelOffer(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslAcceptDuelConfirm(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.AcceptDuelConfirm(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
