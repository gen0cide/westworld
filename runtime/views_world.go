package runtime

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Views for the `world` faculty — the root worldView and the dialog
// menu state. Everything else worldView reaches lives in sibling
// files: recent-event record views in views_records.go, the
// per-instance entity views (player / npc / ground item) in
// views_entities.go, the scenery / boundary views and the
// facts-backed location queries (locs / placements) in
// views_scenery.go. The trade / duel / bank sub-views live in
// views_trade.go / views_duel.go / views_bank.go. All are reached
// through worldView.Get below.
//
// The view-root wiring (it.Reserved["world"] = &worldView{...}) lives
// centrally in dsl_bridge.go.

// ---------- world ----------

type worldView struct {
	host *Host
	bind *routineBinding
}

func (w *worldView) Kind() string    { return "world" }
func (w *worldView) Display() string { return "<world>" }

func (w *worldView) Get(field string) (interp.Value, bool) {
	switch field {
	// Recent-events buffer — the single most-recent of each kind.
	// Returns Null when no event of that kind has been observed
	// this session.
	case "last_chat":
		if r := w.host.world.Recent.Chat(); r != nil {
			return &chatRecordView{r: r}, true
		}
		return interp.Null{}, true
	case "last_pm":
		if r := w.host.world.Recent.PM(); r != nil {
			return &pmRecordView{r: r}, true
		}
		return interp.Null{}, true
	case "last_damage":
		if r := w.host.world.Recent.Damage(); r != nil {
			return &damageRecordView{r: r}, true
		}
		return interp.Null{}, true
	case "last_server_message":
		if r := w.host.world.Recent.ServerMessage(); r != nil {
			return &serverMsgRecordView{r: r}, true
		}
		return interp.Null{}, true
	case "messages":
		// world.messages: List<Message> — the bounded server-message
		// log (§10: world.last_server_message -> world.messages),
		// oldest-first. Backed by RecentEvents' bounded ring (#119);
		// the `on message` event fires per new entry. Each Message
		// carries .text / .kind / .at and supports .contains(needle).
		ring := w.host.world.Recent.ServerMessages()
		items := make([]interp.Value, 0, len(ring))
		for _, r := range ring {
			kind := r.Kind
			if kind == "" {
				kind = "server"
			}
			items = append(items, &messageView{text: r.Message, kind: kind, at: r.At})
		}
		return &interp.List{Items: items}, true
	case "last_dialog_text":
		if r := w.host.world.Recent.DialogText(); r != nil {
			return &dialogTextRecordView{r: r}, true
		}
		return interp.Null{}, true

	// Visible entities (lists).
	case "players":
		records := w.host.world.Players.All()
		items := make([]interp.Value, 0, len(records))
		for _, p := range records {
			items = append(items, &playerView{record: p, host: w.host})
		}
		return &interp.List{Items: items}, true
	case "npcs":
		records := w.host.world.Npcs.All()
		items := make([]interp.Value, 0, len(records))
		for _, n := range records {
			items = append(items, &npcView{record: n, facts: w.host.facts, host: w.host})
		}
		return &interp.List{Items: items}, true
	case "ground_items":
		return &groundItemsView{host: w.host}, true
	case "scenery":
		return &sceneryView{host: w.host}, true
	case "locs":
		return &locsView{host: w.host}, true
	case "boundaries":
		return &boundariesView{host: w.host}, true
	case "dialog":
		return &dialogView{host: w.host}, true
	case "trade":
		return &tradeView{host: w.host, bind: w.bind}, true
	case "duel":
		return &duelView{host: w.host, bind: w.bind}, true
	case "bank":
		return &bankView{host: w.host, bind: w.bind}, true
	}
	return nil, false
}

// dialogView surfaces NPC dialog menu state:
//
//	world.dialog.is_open       → bool, true if a menu is presented
//	world.dialog.options       → list<String>, the options (empty if none)
//	world.dialog.find_option(s) → int (1-based index, or 0 if absent; pairs with answer)
//	world.dialog.answer(s)     → looks up + answers in one call
//	world.dialog.clear()       → reset after answering
//
// Replaces blind answer(N) indexing — routines read the option
// text and pick by content. RSC servers don't always send options
// in stable order across encounters.
type dialogView struct{ host *Host }

func (d *dialogView) Kind() string    { return "dialog" }
func (d *dialogView) Display() string { return "<dialog>" }

func (d *dialogView) Get(field string) (interp.Value, bool) {
	rec := d.host.world.Recent.DialogOptions()
	switch field {
	case "is_open":
		return interp.Bool(rec != nil), true
	case "options":
		if rec == nil {
			return &interp.List{Items: []interp.Value{}}, true
		}
		items := make([]interp.Value, 0, len(rec.Options))
		for _, o := range rec.Options {
			items = append(items, interp.String(o))
		}
		return &interp.List{Items: items}, true
	case "find_option":
		return findOptionCallable{host: d.host}, true
	case "clear":
		return clearDialogCallable{host: d.host}, true
	}
	return nil, false
}

// findOptionCallable backs world.dialog.find_option(text). Returns the
// 1-BASED index of the FIRST option whose text contains (case-insensitive)
// the given substring, or 0 if no match — matching the top-level
// find_option builtin and answer() so `answer(world.dialog.find_option("yes"))`
// is correct (previously this was 0-based/-1, a footgun that silently
// picked the wrong option). Substring (not exact) matching is forgiving —
// quest dialog text often has trailing details ("Yes, I'd like to help.").
type findOptionCallable struct{ host *Host }

func (c findOptionCallable) Kind() string    { return "builtin" }
func (c findOptionCallable) Display() string { return "<dialog.find_option>" }
func (c findOptionCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("find_option takes 1 arg (substring), got %d", len(args))
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, fmt.Errorf("find_option: needle must be String, got %s", args[0].Kind())
	}
	rec := c.host.world.Recent.DialogOptions()
	if rec == nil {
		return interp.Int(0), nil
	}
	lower := strings.ToLower(string(needle))
	for i, opt := range rec.Options {
		if strings.Contains(strings.ToLower(opt), lower) {
			return interp.Int(int64(i + 1)), nil // 1-based, pairs with answer()
		}
	}
	return interp.Int(0), nil
}

// clearDialogCallable backs world.dialog.clear(). Used after a
// routine has resolved the menu (via answer + side effects) to
// invalidate the cached options. The server doesn't reliably
// signal menu close.
type clearDialogCallable struct{ host *Host }

func (c clearDialogCallable) Kind() string    { return "builtin" }
func (c clearDialogCallable) Display() string { return "<dialog.clear>" }
func (c clearDialogCallable) Call(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	c.host.world.Recent.ClearDialogOptions()
	return interp.Null{}, nil
}
