package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
)

// This file bridges the host's event.Bus to the DSL interpreter's
// PendingEvent channel. The translation matches the v1 event table
// in docs/dsl.md "Event handlers":
//
//   | DSL event name   | Source event.Event           | DSL args            |
//   |------------------|------------------------------|---------------------|
//   | chat_received    | event.ChatReceived           | speaker, message    |
//   | private_message  | event.PrivateMessage         | speaker, message    |
//   | server_message   | event.SystemMessage          | text                |
//   | inventory_full   | event.InventorySlotUpdate    | (none)              |
//   | level_up         | event.StatUpdate (level↑)    | skill, new_level    |
//   | trade_request    | event.TradeRequestReceived   | other (index)       |
//   | item_appeared    | event.GroundItemEvent (add)  | item (ground item)  |
//   | npc_appeared     | event.NpcNearby (new=true)   | npc                 |
//   | npc_moved        | event.NpcNearby (new=false)  | npc                 |
//   | damage_taken     | event.OtherPlayerDamage      | amount, source      |
//   | coords_changed   | event.OwnPositionUpdate      | x, y                |
//
// Events not in this table are silently dropped (the interpreter
// won't have a handler for them). hp_below and fatigue_above use
// registration arguments rather than event params, so they're
// driven by the threshold-watcher in auto_eat.go (step 6 territory).

// startEventTranslator subscribes to the host's event bus and
// forwards translated events into the interpreter's Events channel
// until ctx is canceled. Runs in a single goroutine — events are
// delivered in the order published.
//
// Buffered subscriber + a select with default drops events on a
// full DSL queue; a slow routine can't back-pressure the bus.
func (h *Host) startEventTranslator(ctx context.Context, it *interp.Interpreter) {
	sub := h.bus.Subscribe("*", 64)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-sub:
				if !ok {
					return
				}
				pe, has := translateEvent(h, ev)
				if !has {
					continue
				}
				select {
				case it.Events <- pe:
				default:
					// Routine fell behind — drop the event rather
					// than block the publisher. Routines that need
					// guarantees should poll via wait_for_chat or
					// recall.
					h.log.Debug("dsl event dropped (queue full)", "event", pe.Name)
				}
			}
		}
	}()
}

// translateEvent maps a typed event.Event onto a PendingEvent with
// the args expected by the v1 handler table. Returns (zero, false)
// for events the DSL doesn't currently surface.
//
// Takes *Host so player-index → name lookups can resolve via
// world.Players (OtherPlayerChat carries an index, not a name —
// the appearance event seen earlier seeds the name in the world
// state and we read it back here).
func translateEvent(h *Host, ev event.Event) (interp.PendingEvent, bool) {
	switch e := ev.(type) {
	case event.ChatReceived:
		// Server-issued message with sender (rare path — most
		// player-to-player chat goes through OtherPlayerChat below).
		return interp.PendingEvent{
			Name: "chat_received",
			Args: []interp.Value{interp.String(e.Speaker), interp.String(e.Message)},
		}, true
	case event.OtherPlayerChat:
		// Public chat from another visible player. The packet
		// only carries a player-index; we resolve to a name via
		// world.Players, which was seeded by the appearance event
		// for this index. If the appearance hasn't been observed
		// yet (unusual — appearance precedes chat in normal play),
		// emit with an empty speaker so routines can still react
		// to the message text alone.
		speaker := ""
		if h != nil && h.world != nil {
			if rec, ok := h.world.Players.Get(e.PlayerIndex); ok {
				speaker = rec.Name
			}
		}
		return interp.PendingEvent{
			Name: "chat_received",
			Args: []interp.Value{interp.String(speaker), interp.String(e.MessageText)},
		}, true
	case event.ItemGained:
		return interp.PendingEvent{
			Name: "item_gained",
			Args: []interp.Value{interp.Int(int64(e.ItemID)), interp.Int(int64(e.Count))},
		}, true
	case event.PrivateMessage:
		return interp.PendingEvent{
			Name: "private_message",
			Args: []interp.Value{interp.String(e.Sender), interp.String(e.Message)},
		}, true
	case event.SystemMessage:
		return interp.PendingEvent{
			Name: "server_message",
			Args: []interp.Value{interp.String(e.Message)},
		}, true
	case event.OwnPositionUpdate:
		return interp.PendingEvent{
			Name: "coords_changed",
			Args: []interp.Value{interp.Int(int64(e.X)), interp.Int(int64(e.Y))},
		}, true
	case event.TradeRequestReceived:
		return interp.PendingEvent{
			Name: "trade_request",
			Args: []interp.Value{interp.Int(int64(e.FromPlayerIndex))},
		}, true
	case event.TradeOpened:
		return interp.PendingEvent{
			Name: "trade_opened",
			Args: []interp.Value{interp.Int(int64(e.OtherPlayerIndex))},
		}, true
	case event.TradeOtherOffer:
		// Surface as a list-of-[id, count] pairs so DSL can iterate.
		pairs := make([]interp.Value, 0, len(e.Items))
		for _, it := range e.Items {
			pair := &interp.List{Items: []interp.Value{
				interp.Int(int64(it.ItemID)),
				interp.Int(int64(it.Amount)),
			}}
			pairs = append(pairs, pair)
		}
		return interp.PendingEvent{
			Name: "trade_other_offer",
			Args: []interp.Value{&interp.List{Items: pairs}},
		}, true
	case event.TradeOtherAccepted:
		return interp.PendingEvent{
			Name: "trade_other_accepted",
			Args: nil,
		}, true
	case event.TradeClosed:
		return interp.PendingEvent{
			Name: "trade_closed",
			Args: []interp.Value{interp.Bool(e.Completed)},
		}, true
	case event.DuelOpened:
		return interp.PendingEvent{
			Name: "duel_opened",
			Args: []interp.Value{interp.Int(int64(e.OtherPlayerIndex))},
		}, true
	case event.DuelOtherOffer:
		pairs := make([]interp.Value, 0, len(e.Items))
		for _, it := range e.Items {
			pair := &interp.List{Items: []interp.Value{
				interp.Int(int64(it.ItemID)),
				interp.Int(int64(it.Amount)),
			}}
			pairs = append(pairs, pair)
		}
		return interp.PendingEvent{
			Name: "duel_other_offer",
			Args: []interp.Value{&interp.List{Items: pairs}},
		}, true
	case event.DuelSettingsUpdate:
		return interp.PendingEvent{
			Name: "duel_settings_update",
			Args: []interp.Value{
				interp.Bool(e.DisallowRetreat),
				interp.Bool(e.DisallowMagic),
				interp.Bool(e.DisallowPrayer),
				interp.Bool(e.DisallowWeapons),
			},
		}, true
	case event.DuelOtherAccepted:
		return interp.PendingEvent{
			Name: "duel_other_accepted",
			Args: nil,
		}, true
	case event.DuelConfirmShown:
		return interp.PendingEvent{
			Name: "duel_confirm_shown",
			Args: nil,
		}, true
	case event.DuelClosed:
		return interp.PendingEvent{
			Name: "duel_closed",
			Args: []interp.Value{interp.Bool(e.Completed)},
		}, true
	}
	return interp.PendingEvent{}, false
}
