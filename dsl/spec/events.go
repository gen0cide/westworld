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
		DocSummary: "Another player initiated a trade with us."},
	{Name: "trade_opened", Params: []string{"other_index"},
		DocSummary: "Both players have accepted the trade request; the offer screen is open. Use world.trade.* to read state."},
	{Name: "trade_other_offer", Params: []string{"items"},
		DocSummary: "Opponent updated their offered items. `items` is a list of [item_id, amount] pairs. Mirrors world.trade.their_offer."},
	{Name: "trade_other_accepted", Params: nil,
		DocSummary: "Opponent clicked Accept on the current trade screen (offer or confirm)."},
	{Name: "trade_closed", Params: []string{"completed"},
		DocSummary: "Trade ended. `completed` is true if both sides confirmed, false on cancel/decline."},
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
		DocSummary: "We just died.",
		NotYetImplemented: true},
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
