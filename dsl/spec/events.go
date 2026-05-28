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
		DocSummary: "Another player initiated a trade with us. `from` is the requester's NAME (String) — OpenRSC's notification packet does not carry an index. Pass it directly to `accept_trade(from)` which resolves to an index via world.Players."},
	{Name: "duel_request_incoming", Params: []string{"from"},
		DocSummary: "Another player initiated a duel with us. `from` is the requester's NAME. Pass to `accept_duel(from)` to reciprocate."},
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

	// Spec'd per docs/lang/events.md "Category A" but not yet
	// emitted by the translator. Routines can author handlers for
	// these; the runtime will start firing them as we wire the
	// inbound side.
	{Name: "damage_taken", Params: []string{"amount", "source"},
		DocSummary: "We were hit. Distinct from `self.hp` polling — fires once per hit.",
		NotYetImplemented: true},
	{Name: "level_up", Params: []string{"skill", "new_level"},
		DocSummary: "Skill level just increased.",
		NotYetImplemented: true},
	{Name: "death", Params: nil,
		DocSummary: "We just died. Fires once when the server confirms HP=0 (before respawn). The runtime resets self.hp to max immediately after respawn, so the only reliable way to react to a kill is `on death { ... }`."},
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
