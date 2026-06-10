package spec

// EventSpec describes one bus event a routine can `on`-handle.
// The validator uses Params to enforce handler arity. The runtime
// bridge (runtime/dsl_events.go) translates typed Go event.X values
// into PendingEvents using the same Name.
type EventSpec struct {
	// Name is the DSL identifier (lowercase snake_case).
	Name string

	// Params lists the handler's positional parameter names in
	// declaration order, matching the typed event's field
	// extraction in runtime/dsl_events.go::translateEvent.
	Params []string

	// DocSummary describes when the event fires.
	DocSummary string

	// NotYetImplemented marks an event the spec defines but the
	// runtime translator doesn't yet emit. Listed here so routine
	// authors can pre-write handlers; the consistency test
	// distinguishes spec-only from runtime-only.
	NotYetImplemented bool
}

// Events is the v1 transient-event table. These are events the
// world produces that can't be polled as state — they happen and
// they're gone. State changes that COULD be polled live as
// `when expr` watchers instead (see docs/lang/events.md "Category B").
//
// The list is intentionally small. If a routine wants to react to
// "this state is now true" rather than "this thing happened," use
// `when`/`select` over the query layer, not a bus event.
var Events = []EventSpec{
	// Currently implemented in runtime/dsl_events.go.
	{Name: "chat_received", Params: []string{"speaker", "message"},
		DocSummary: "Another player said something publicly within view."},
	{Name: "private_message", Params: []string{"speaker", "message"},
		DocSummary: "An incoming PM from a friend."},
	{Name: "server_message", Params: []string{"text"},
		DocSummary: "System message from the server."},
	{Name: "coords_changed", Params: []string{"x", "y"},
		DocSummary: "Own position updated (every coord packet)."},
	{Name: "trade_request", Params: []string{"from"},
		DocSummary: "Another player initiated a trade with us. `from` is the requester's NAME (String) — OpenRSC's notification packet does not carry an index. Pass it directly to `trade.request(from)` (which reciprocates the request — the server opens the trade once both sides have requested) — it resolves the name to an index via world.Players."},
	{Name: "duel_request_incoming", Params: []string{"from"},
		DocSummary: "Another player initiated a duel with us. `from` is the requester's NAME. Pass to `duel.request(from)` to reciprocate (the server opens the duel once both sides have requested)."},
	{Name: "trade_opened", Params: []string{"other_index"},
		DocSummary: "Both players have accepted the trade request; the offer screen is open. Use world.trade.* to read state."},
	{Name: "trade_other_offer", Params: []string{"items"},
		DocSummary: "Opponent updated their offered items. `items` is a list of [item_id, amount] pairs. Mirrors world.trade.their_offer."},
	{Name: "trade_other_accepted", Params: nil,
		DocSummary: "Opponent clicked Accept on the current trade screen (offer or confirm)."},
	{Name: "trade_closed", Params: []string{"completed"},
		DocSummary: "Trade ended. `completed` is true if both sides confirmed, false on cancel/decline."},
	{Name: "trade_confirm_shown", Params: nil,
		DocSummary: "Server pushed the final confirm-window — both sides have first-accepted and we're on the final review screen."},
	{Name: "duel_opened", Params: []string{"other_index"},
		DocSummary: "Duel request was accepted; the offer screen is open. Use world.duel.* to read state."},
	{Name: "duel_other_offer", Params: []string{"items"},
		DocSummary: "Opponent updated their staked items. `items` is a list of [item_id, amount] pairs. Mirrors world.duel.their_offer."},
	{Name: "duel_settings_update", Params: []string{"disallow_retreat", "disallow_magic", "disallow_prayer", "disallow_weapons"},
		DocSummary: "Duel rule toggles changed. All four bools delivered together (server pushes the full set on any change)."},
	{Name: "duel_other_accepted", Params: nil,
		DocSummary: "Opponent clicked Accept on the current duel screen (offer or confirm)."},
	{Name: "duel_confirm_shown", Params: nil,
		DocSummary: "Server pushed the final confirm-window — both sides have first-accepted and we're on the final review screen."},
	{Name: "duel_closed", Params: []string{"completed"},
		DocSummary: "Duel ended. `completed` is true if both sides confirmed twice (fight starts); false on cancel/decline."},
	{Name: "item_gained", Params: []string{"item_id", "count"},
		DocSummary: "Inventory net-gained N of item_id. Fires per added item (not per stack-increment). Synthesized from inventory snapshot/slot-update diffs."},

	// ===== #119 events (runtime/dsl_events.go::translateEvents) =====
	{Name: "message", Params: []string{"text"},
		DocSummary: "A new server (system) message arrived. `text` is the message string; filter with text.contains(needle) or read the bounded log via world.messages. Ring-backed sibling of server_message."},
	{Name: "xp_gain", Params: []string{"skill", "amount"},
		DocSummary: "A skill's experience increased. `skill` is the lowercase skill name (attack, fishing, …); `amount` is the positive xp delta. Filter in-body on `skill` (e.g. `if skill == \"fishing\"`). Synthesized by diffing the per-skill xp mirror across stat/xp packets."},
	{Name: "target_died", Params: []string{"target"},
		DocSummary: "The engaged combat target (combat.target / combat.last_npc) died — its opcode-104 current-hitpoints reading went from >0 to 0. `target` is the dead Npc view (or null if it already left view). Fires once per kill."},
	{Name: "npc_killed", Params: []string{"target"},
		DocSummary: "Alias of target_died: the engaged NPC combat target died. `target` is the dead Npc view (or null). Fires alongside target_died on the same death edge."},

	// Spec'd per docs/lang/events.md "Category A" but not yet
	// emitted by the translator. Routines can author handlers for
	// these; the runtime will start firing them as we wire the
	// inbound side.
	{Name: "damage_taken", Params: []string{"amount", "source"},
		DocSummary: "We were hit. Distinct from `self.hp` polling — fires once per hit.",
		NotYetImplemented: true},
	{Name: "level_up", Params: []string{"skill", "new_level"},
		DocSummary: "One of our OWN skills just gained a base level. skill is the lowercase skill name; new_level is the new base level. Synthesized by diffing per-skill base levels across stat updates."},
	{Name: "equipment_changed", Params: []string{"slot", "item"},
		DocSummary: "Our OWN worn/wielded set changed. Fires once per changed slot: slot is the human slot name (helmet/body/legs/weapon/...), item is the new worn item there (item.is_empty if removed). Also read self.equipped / self.equipped.bonuses for the full new state."},
	{Name: "player_equipment_changed", Params: []string{"player", "slot", "item"},
		DocSummary: "A visible player's worn equipment changed. Fires once per changed slot: player is the player view, slot is the human slot name (helmet/body/legs/weapon/shield/gloves/boots/amulet/cape), item is the new worn item in that slot (item.is_empty if removed). Fires only on an actual change, not the periodic appearance re-send."},
	{Name: "player_level_changed", Params: []string{"player", "new_level"},
		DocSummary: "A visible player's combat level changed. player is the player view; new_level is their new combat level."},
	{Name: "death", Params: nil,
		DocSummary: "We just died. Fires once when the server confirms HP=0 (before respawn). The runtime resets self.hp to max immediately after respawn, so the only reliable way to react to a kill is `on death { ... }`."},
	{Name: "item_appeared", Params: []string{"item_id", "x", "y"},
		DocSummary: "A ground item entered visibility. Coords are absolute (resolved from the packet's player-relative offsets at delivery time). Use world.ground_items to query the full visible set."},
	{Name: "item_disappeared", Params: []string{"item_id", "x", "y"},
		DocSummary: "A ground item left visibility (picked up by someone, or we walked too far away)."},
	{Name: "bank_opened", Params: []string{"max_size"},
		DocSummary: "The bank window opened (right-click banker → bank, or talk_to + dialog choice). max_size is the slot capacity. world.bank.slots has the contents."},
	{Name: "bank_slot_update", Params: []string{"slot", "item_id", "amount"},
		DocSummary: "A bank slot was updated. amount=0 means the slot was emptied. world.bank reflects the change before the handler fires."},
	{Name: "bank_closed", Params: nil,
		DocSummary: "The bank window closed. world.bank.is_open is false from this point."},
	{Name: "boundary_changed", Params: []string{"x", "y", "dir", "id"},
		DocSummary: "A dynamic boundary update arrived (door opened, web cut, etc.). id=-1 means the boundary was removed. world.boundaries.is_open(x, y, dir) reflects the change before the handler fires."},
}

// eventByName is the lookup map. Built once at init.
var eventByName map[string]*EventSpec

func init() {
	eventByName = make(map[string]*EventSpec, len(Events))
	for i := range Events {
		eventByName[Events[i].Name] = &Events[i]
	}
}

// EventByName returns the EventSpec for `name` and ok=true if
// found.
func EventByName(name string) (*EventSpec, bool) {
	s, ok := eventByName[name]
	return s, ok
}
