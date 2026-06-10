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
		// Without this, every interpreter construction (every turn, detour,
		// /script, ANALYSIS command) leaves a dead channel registered on the
		// bus forever — the dominant runtime leak (audit 2026-06-10).
		defer h.bus.Unsubscribe("*", sub)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-sub:
				if !ok {
					return
				}
				// A single bus event may surface as MORE THAN ONE DSL
				// event (e.g. SystemMessage drives both server_message
				// and the #119 ring-backed `message`). translateEvents
				// returns all of them in order; each is forwarded
				// independently so a full queue drops only the laggard.
				for _, pe := range translateEvents(h, ev) {
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
		}
	}()
}

// translateEvents maps one typed bus event onto the full set of DSL
// PendingEvents it should surface (zero, one, or several). It wraps the
// legacy single-event translateEvent (kept for back-compat) and adds
// the #119 events that either fan out from a shared source (message,
// alongside server_message) or have a dedicated synthetic source
// (xp_gain, target_died + its npc_killed alias).
//
// Order matters for fan-out: the legacy event is delivered first so
// existing `on server_message` handlers keep their relative timing.
func translateEvents(h *Host, ev event.Event) []interp.PendingEvent {
	var out []interp.PendingEvent
	if pe, has := translateEvent(h, ev); has {
		out = append(out, pe)
	}
	switch e := ev.(type) {
	case event.SystemMessage:
		// #119: `on message(text)` — fires per new server message,
		// fed by the same SystemMessage that drives server_message.
		// Backed by the world.messages ring (world.Apply ran before
		// publish, so the ring already holds this entry). Routines
		// filter with `if text.contains(...)` or read world.messages.
		out = append(out, interp.PendingEvent{
			Name: "message",
			Args: []interp.Value{interp.String(e.Message)},
		})
	case event.XPGain:
		// #119: `on xp_gain(skill, amount)` — parameterized by skill.
		// The handler binds the lowercase skill name + the positive
		// xp delta and filters on the name (mirrors level_up).
		out = append(out, interp.PendingEvent{
			Name: "xp_gain",
			Args: []interp.Value{
				interp.String(event.SkillName(e.Skill)),
				interp.Int(int64(e.Amount)),
			},
		})
	case event.TargetDied:
		// #119: the engaged combat target died. Resolve the NPC view
		// live from world.Npcs by index (Null if it already left view)
		// and surface it under BOTH `target_died` and the `npc_killed`
		// alias, since the contract names both for the same edge.
		var target interp.Value = interp.Null{}
		if rec, ok := h.world.Npcs.Get(e.NpcIndex); ok {
			target = &npcView{record: rec, facts: h.facts, host: h}
		}
		out = append(out,
			interp.PendingEvent{Name: "target_died", Args: []interp.Value{target}},
			interp.PendingEvent{Name: "npc_killed", Args: []interp.Value{target}},
		)
	case event.LevelUp:
		// `on level_up(skill, new_level)` — our own base level rose.
		out = append(out, interp.PendingEvent{
			Name: "level_up",
			Args: []interp.Value{
				interp.String(event.SkillName(e.Skill)),
				interp.Int(int64(e.NewLevel)),
			},
		})
	case event.EquipmentChanged:
		// `on equipment_changed(slot, item)` — our own worn set changed in
		// `slot`; resolve the new item there from our inventory.
		var item interp.Value = interp.Null{}
		if slots, ok := equipSlotGroups[e.Slot]; ok {
			item = &wornItemView{w: h.selfWornGroup(slots)}
		}
		out = append(out, interp.PendingEvent{
			Name: "equipment_changed",
			Args: []interp.Value{interp.String(e.Slot), item},
		})
	case event.PlayerEquipmentChanged:
		// `on player_equipment_changed(player, slot, item)` — resolve the
		// live player view and the new worn item in the changed slot, so
		// the handler gets exactly what changed (not just "something did").
		var p interp.Value = interp.Null{}
		var item interp.Value = interp.Null{}
		if rec, ok := h.world.Players.Get(e.PlayerIndex); ok {
			p = &playerView{record: rec, host: h}
			if slots, ok := equipSlotGroups[e.Slot]; ok {
				item = &wornItemView{w: h.wornGroupFromAppearance(rec.EquipBySlot, slots)}
			}
		}
		out = append(out, interp.PendingEvent{
			Name: "player_equipment_changed",
			Args: []interp.Value{p, interp.String(e.Slot), item},
		})
	case event.PlayerLevelChanged:
		// `on player_level_changed(player, new_level)`.
		var p interp.Value = interp.Null{}
		if rec, ok := h.world.Players.Get(e.PlayerIndex); ok {
			p = &playerView{record: rec, host: h}
		}
		out = append(out, interp.PendingEvent{
			Name: "player_level_changed",
			Args: []interp.Value{p, interp.Int(int64(e.NewLevel))},
		})
	}
	return out
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
	case event.Death:
		return interp.PendingEvent{
			Name: "death",
			Args: nil,
		}, true
	case event.OtherPlayerDamage:
		// The server publishes damage events for every visible player.
		// Routine handlers want `damage_taken` to mean "I was hit" —
		// so we only translate when PlayerIndex == 0 (the local
		// player's own slot in their world view). Damage to other
		// players is observable via the typed event bus for routines
		// that want it, but is not surfaced through the DSL handler
		// table (yet) because no scenario has reached for it.
		//
		// Source is the empty string for admin-issued damage; in
		// combat it'd be the attacker's name, but the v235 protocol
		// puts no attacker info on the damage packet — so we leave
		// it blank and let routines branch on last_attacked_npc /
		// last_attacked_player if they need attribution.
		if e.PlayerIndex != h.world.Players.SelfIndex() {
			return interp.PendingEvent{}, false
		}
		return interp.PendingEvent{
			Name: "damage_taken",
			Args: []interp.Value{interp.Int(int64(e.Damage)), interp.String("")},
		}, true
	case event.BankOpened:
		return interp.PendingEvent{
			Name: "bank_opened",
			Args: []interp.Value{interp.Int(int64(e.MaxSize))},
		}, true
	case event.BankSlotUpdate:
		return interp.PendingEvent{
			Name: "bank_slot_update",
			Args: []interp.Value{
				interp.Int(int64(e.Slot)),
				interp.Int(int64(e.ItemID)),
				interp.Int(int64(e.Amount)),
			},
		}, true
	case event.BankClosed:
		return interp.PendingEvent{
			Name: "bank_closed",
			Args: nil,
		}, true
	case event.BoundaryUpdates:
		// Surface as one event per delta so routines don't need to
		// iterate. The world.Boundaries state is already updated
		// before this fires.
		// (For now we emit only the FIRST delta to keep arity
		// stable; future: a list-shaped boundary_updates event.)
		if len(e.Updates) == 0 {
			return interp.PendingEvent{}, false
		}
		if h == nil || h.world == nil {
			return interp.PendingEvent{}, false
		}
		pos := h.world.Self.Position()
		d := e.Updates[0]
		ax := pos.X + d.OffsetX
		ay := pos.Y + d.OffsetY
		return interp.PendingEvent{
			Name: "boundary_changed",
			Args: []interp.Value{
				interp.Int(int64(ax)),
				interp.Int(int64(ay)),
				interp.Int(int64(d.Dir)),
				interp.Int(int64(d.ID)),
			},
		}, true
	case event.GroundItemEvent:
		// Convert relative offsets to absolute coords using the
		// player position at packet arrival (same calc as
		// world.Apply does). Routines see absolute coords directly.
		if h == nil || h.world == nil {
			return interp.PendingEvent{}, false
		}
		pos := h.world.Self.Position()
		x := pos.X + e.OffsetX
		y := pos.Y + e.OffsetY
		name := "item_appeared"
		if e.Disappear {
			name = "item_disappeared"
		}
		return interp.PendingEvent{
			Name: name,
			Args: []interp.Value{
				interp.Int(int64(e.ItemID)),
				interp.Int(int64(x)),
				interp.Int(int64(y)),
			},
		}, true
	case event.GroundItemUpdates:
		// The opcode-99 batch. Fire one representative DSL event (preferring
		// the first ADD, mirroring boundary_changed's single-delta dispatch);
		// the FULL set reaches perception via world.GroundItems in world.Apply,
		// so look_around / scan see every item even though only one fires here.
		if h == nil || h.world == nil || len(e.Updates) == 0 {
			return interp.PendingEvent{}, false
		}
		pos := h.world.Self.Position()
		d := e.Updates[0]
		for _, u := range e.Updates {
			if !u.Disappear {
				d = u
				break
			}
		}
		x := pos.X + d.OffsetX
		y := pos.Y + d.OffsetY
		name := "item_appeared"
		if d.Disappear {
			name = "item_disappeared"
		}
		return interp.PendingEvent{
			Name: name,
			Args: []interp.Value{
				interp.Int(int64(d.ItemID)),
				interp.Int(int64(x)),
				interp.Int(int64(y)),
			},
		}, true
	case event.TradeRequestReceived:
		// OpenRSC's notification only carries the requester's name
		// (not their server-index). Pass the name string; the
		// `accept_trade(name)` builtin resolves it back via
		// world.Players for the outbound packet.
		return interp.PendingEvent{
			Name: "trade_request",
			Args: []interp.Value{interp.String(e.FromPlayerName)},
		}, true
	case event.DuelRequestReceived:
		return interp.PendingEvent{
			Name: "duel_request_incoming",
			Args: []interp.Value{interp.String(e.FromPlayerName)},
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
		// world.Apply runs before publish — read the trade record's
		// terminal phase rather than e.Completed (the protocol's
		// close packet has no completion bit; the inference lives
		// in world/world.go).
		completed := e.Completed
		if h != nil && h.world != nil {
			if rec := h.world.Trade.Trade(); rec != nil && rec.Phase == "completed" {
				completed = true
			}
		}
		return interp.PendingEvent{
			Name: "trade_closed",
			Args: []interp.Value{interp.Bool(completed)},
		}, true
	case event.TradeConfirmShown:
		return interp.PendingEvent{
			Name: "trade_confirm_shown",
			Args: nil,
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
		// Same caveat as trade_closed — protocol close packet has no
		// completion bit; read from the world record after Apply.
		completed := e.Completed
		if h != nil && h.world != nil {
			if rec := h.world.Duel.Duel(); rec != nil && rec.Phase == "completed" {
				completed = true
			}
		}
		return interp.PendingEvent{
			Name: "duel_closed",
			Args: []interp.Value{interp.Bool(completed)},
		}, true
	}
	return interp.PendingEvent{}, false
}
