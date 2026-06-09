package runtime

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// Views for the `world` faculty — the root worldView and everything it
// reaches: recent-event records, dialog menu state, the per-instance
// entity views (player / npc / ground item / boundary), and the
// facts-backed location queries (locs / placements). The trade / duel
// / bank sub-views live in views_trade.go / views_duel.go /
// views_bank.go but are reached through worldView.Get below.
//
// The view-root wiring (it.Reserved["world"] = &worldView{...}) lives
// centrally in dsl_bridge.go.

// ---------- world ----------

type worldView struct{ host *Host }

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
		return &tradeView{host: w.host}, true
	case "duel":
		return &duelView{host: w.host}, true
	case "bank":
		return &bankView{host: w.host}, true
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

// ---------- recent-events views ----------

// chatRecordView wraps a world.ChatRecord so DSL routines can
// read .speaker / .message / .at.
type chatRecordView struct{ r *world.ChatRecord }

func (v *chatRecordView) Kind() string    { return "chat_record" }
func (v *chatRecordView) Display() string { return v.r.Speaker + ": " + v.r.Message }
func (v *chatRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "speaker":
		return interp.String(v.r.Speaker), true
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

type pmRecordView struct{ r *world.PMRecord }

func (v *pmRecordView) Kind() string    { return "pm_record" }
func (v *pmRecordView) Display() string { return "PM from " + v.r.Sender + ": " + v.r.Message }
func (v *pmRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "sender":
		return interp.String(v.r.Sender), true
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

type damageRecordView struct{ r *world.DamageRecord }

func (v *damageRecordView) Kind() string    { return "damage_record" }
func (v *damageRecordView) Display() string { return "took " + intDisp(v.r.Amount) + " damage" }
func (v *damageRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "amount":
		return interp.Int(int64(v.r.Amount)), true
	case "source":
		return interp.String(v.r.Source), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	}
	return nil, false
}

type serverMsgRecordView struct{ r *world.ServerMsgRecord }

func (v *serverMsgRecordView) Kind() string    { return "server_msg_record" }
func (v *serverMsgRecordView) Display() string { return v.r.Message }
func (v *serverMsgRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

// messageView is a Message INSTANCE (api.md §2/§8): one entry in the
// world.messages log. Fields: .text (String), .kind (String), .at
// (Time, formatted). Method: .contains(needle) — case-insensitive
// substring match. The Message type unifies what was the single
// last_server_message record into a list element.
type messageView struct {
	text string
	kind string
	at   time.Time
}

func (m *messageView) Kind() string    { return "message" }
func (m *messageView) Display() string { return m.text }
func (m *messageView) Get(field string) (interp.Value, bool) {
	switch field {
	case "text", "message":
		return interp.String(m.text), true
	case "kind":
		return interp.String(m.kind), true
	case "at":
		return interp.String(m.at.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: m.text}, true
	}
	return nil, false
}

type dialogTextRecordView struct{ r *world.DialogTextRecord }

func (v *dialogTextRecordView) Kind() string    { return "dialog_text_record" }
func (v *dialogTextRecordView) Display() string { return v.r.Text }
func (v *dialogTextRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "text":
		return interp.String(v.r.Text), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Text}, true
	}
	return nil, false
}

// substringCallable backs the `.contains(needle)` method on
// message record views. Case-insensitive by default — routines
// branch on prose without caring about server capitalization.
// Returns Bool. Used as: world.last_server_message.contains("locked").
type substringCallable struct{ haystack string }

func (c substringCallable) Kind() string    { return "builtin" }
func (c substringCallable) Display() string { return "<contains>" }
func (c substringCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("contains takes 1 arg (needle), got %d", len(args))
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, fmt.Errorf("contains: needle must be String, got %s", args[0].Kind())
	}
	return interp.Bool(strings.Contains(strings.ToLower(c.haystack), strings.ToLower(string(needle)))), nil
}

// ---------- entity views ----------

type playerView struct {
	record world.PlayerRecord
	host   *Host
}

func (p *playerView) Kind() string    { return "player" }
func (p *playerView) Display() string { return p.record.Name }
func (p *playerView) Get(field string) (interp.Value, bool) {
	// Worn-equipment perception (#equipment): `player.equipment` for the
	// whole set, or a direct slot accessor (player.helmet / .weapon / ...).
	// Resolved from the player's appearance packet via the runtime-level
	// Host.WornItemFromAppearance — see runtime/equipment.go.
	if field == "equipment" {
		return &equipmentView{host: p.host, eq: p.record.EquipBySlot}, true
	}
	if slots, ok := equipSlotGroups[field]; ok {
		return &wornItemView{w: p.host.wornGroupFromAppearance(p.record.EquipBySlot, slots)}, true
	}
	switch field {
	case "index":
		return interp.Int(int64(p.record.Index)), true
	case "name":
		return interp.String(p.record.Name), true
	case "x":
		return interp.Int(int64(p.record.X)), true
	case "y":
		return interp.Int(int64(p.record.Y)), true
	case "position":
		return &positionView{X: p.record.X, Y: p.record.Y}, true

	// ----- combat level + relative threat -----
	// combat_level comes from the appearance packet's trailing bytes
	// (null until decoded). relative_level / threat / threat_colour
	// express how dangerous this player is RELATIVE to us — the cue the
	// client paints as the level's colour (Formulae.getLvlDiffColour);
	// a host can't read a UI colour, so we surface the concept. See
	// runtime/equipment.go::threatBand.
	case "combat_level":
		if p.record.HasAppearanceCombat {
			return interp.Int(int64(p.record.CombatLevel)), true
		}
		return interp.Null{}, true
	case "relative_level":
		if p.record.HasAppearanceCombat {
			return interp.Int(int64(p.record.CombatLevel - p.host.world.Self.CombatLevel())), true
		}
		return interp.Null{}, true
	case "threat", "threat_colour":
		if !p.record.HasAppearanceCombat {
			return interp.Null{}, true
		}
		label, colour := threatBand(p.host.world.Self.CombatLevel(), p.record.CombatLevel)
		if field == "threat_colour" {
			return interp.String(colour), true
		}
		return interp.String(label), true
	case "is_skulled":
		return interp.Bool(p.record.SkullType != 0), true

	// ----- combat (#117) -----
	// hp_fraction / health read the cur/max hitpoints from the
	// opcode-234 type-2 damage update (PlayerRecord.CurHits/MaxHits,
	// gated by HasHits). In PVP the engaged opponent's health bar
	// arrives this way; an unfought player reports null. Mirrors the
	// Npc shape so combat.target.hp_fraction works for either side.
	case "hp_fraction":
		if p.record.HasHits && p.record.MaxHits > 0 {
			return interp.Float(float64(p.record.CurHits) / float64(p.record.MaxHits)), true
		}
		return interp.Null{}, true
	case "health":
		if p.record.HasHits {
			return interp.Int(int64(p.record.CurHits)), true
		}
		return interp.Null{}, true

	// Stubs until the host tracks friends list + combat targets.
	// is_friend will derive from a friends-list mirror once we
	// decode the friend-list packets (currently we send AddFriend
	// outbound but don't mirror server-side state). in_combat_with
	// requires per-player combat target tracking — both stubbed.
	case "is_friend":
		return interp.Bool(false), true
	case "in_combat_with":
		return interp.Null{}, true
	}
	return nil, false
}

type npcView struct {
	record world.NpcRecord
	facts  *facts.Facts
	host   *Host
}

func (n *npcView) Kind() string { return "npc" }
func (n *npcView) Display() string {
	if n.facts != nil {
		if def := n.facts.NpcDef(n.record.TypeID); def != nil {
			return def.Name
		}
	}
	return "npc#" + intDisp(n.record.Index)
}

// def returns the facts-side NPC definition, or nil if facts
// aren't loaded or the type isn't known. Helper for the Get
// switch arms.
func (n *npcView) def() *facts.NpcDef {
	if n.facts == nil {
		return nil
	}
	return n.facts.NpcDef(n.record.TypeID)
}

func (n *npcView) Get(field string) (interp.Value, bool) {
	switch field {
	case "index":
		return interp.Int(int64(n.record.Index)), true
	case "type_id":
		return interp.Int(int64(n.record.TypeID)), true
	case "x":
		return interp.Int(int64(n.record.X)), true
	case "y":
		return interp.Int(int64(n.record.Y)), true
	case "position":
		return &positionView{X: n.record.X, Y: n.record.Y}, true
	case "def":
		// The static catalog entry (NpcDef), api.md §2. Carries
		// id/name/combat_level/max_hp/attackable/aggressive/
		// command1/command2. Null if facts not loaded / type unknown.
		if def := n.def(); def != nil {
			return &npcDefView{def: def}, true
		}
		return interp.Null{}, true
	case "name":
		// Always return a String so routines can write
		// `n => n.name.lower == "cook"` without null-guards. NPCs
		// whose def hasn't loaded yet get "" — find/filter lambdas
		// see them as non-matching, which is the right behavior.
		if def := n.def(); def != nil {
			return interp.String(def.Name), true
		}
		return interp.String(""), true

	// Facts-derived combat / interaction fields.
	case "combat_level":
		if def := n.def(); def != nil {
			// Standard RSC combat-level approximation: (atk+str+def)/4 + hits/4
			cl := (def.Attack+def.Strength+def.Defense)/4 + def.Hits/4
			return interp.Int(int64(cl)), true
		}
		return interp.Null{}, true
	case "relative_level", "threat", "threat_colour":
		// How dangerous this NPC is RELATIVE to us — the cue the client
		// paints as the level number's colour (the tutorial's "darker red
		// = more dangerous"). See runtime/equipment.go::threatBand.
		def := n.def()
		if def == nil || n.host == nil {
			return interp.Null{}, true
		}
		theirs := (def.Attack+def.Strength+def.Defense)/4 + def.Hits/4
		mine := n.host.world.Self.CombatLevel()
		if field == "relative_level" {
			return interp.Int(int64(theirs - mine)), true
		}
		label, colour := threatBand(mine, theirs)
		if field == "threat_colour" {
			return interp.String(colour), true
		}
		return interp.String(label), true
	case "max_hp":
		if def := n.def(); def != nil {
			return interp.Int(int64(def.Hits)), true
		}
		return interp.Null{}, true
	case "is_attackable":
		if def := n.def(); def != nil {
			return interp.Bool(def.Attackable), true
		}
		return interp.Bool(false), true
	case "is_aggressive":
		if def := n.def(); def != nil {
			return interp.Bool(def.Aggressive), true
		}
		return interp.Bool(false), true

	// ----- combat (#117) -----
	// Live combat-state fields. hp_fraction reads the cur/max
	// hitpoints landed by the opcode-104 SEND_UPDATE_NPC type-2
	// decoder (NpcRecord.CurHits/MaxHits, gated by HasHits). Only the
	// engaged target gets a health update on the wire, so an NPC we've
	// never fought reports null (HasHits=false) — routines branch on
	// that. `health` is the integer current hitpoints (same gate).
	case "hp_fraction":
		if n.record.HasHits && n.record.MaxHits > 0 {
			return interp.Float(float64(n.record.CurHits) / float64(n.record.MaxHits)), true
		}
		return interp.Null{}, true
	case "health":
		if n.record.HasHits {
			return interp.Int(int64(n.record.CurHits)), true
		}
		return interp.Null{}, true
	case "in_combat_with":
		return interp.Null{}, true
	}
	return nil, false
}

// npcDefView is an NpcDef DEFINITION (api.md §2): the static catalog
// entry for an NPC species, immutable, from the facts registry.
// Reached via Npc.def. Static attributes only — no live state.
type npcDefView struct{ def *facts.NpcDef }

func (d *npcDefView) Kind() string    { return "npc_def" }
func (d *npcDefView) Display() string { return d.def.Name }

func (d *npcDefView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		return interp.Int(int64(d.def.ID)), true
	case "name":
		return interp.String(d.def.Name), true
	case "description":
		return interp.String(d.def.Description), true
	case "combat_level":
		cl := (d.def.Attack+d.def.Strength+d.def.Defense)/4 + d.def.Hits/4
		return interp.Int(int64(cl)), true
	case "max_hp":
		return interp.Int(int64(d.def.Hits)), true
	case "attackable", "is_attackable":
		return interp.Bool(d.def.Attackable), true
	case "aggressive", "is_aggressive":
		return interp.Bool(d.def.Aggressive), true
	case "command1":
		return interp.String(d.def.Command1), true
	case "command2":
		return interp.String(d.def.Command2), true
	}
	return nil, false
}

// groundItemsView is the entry point for ground-item queries:
//
//	world.ground_items                   — list of all visible ground items
//	world.ground_items.by_id(N)          — nearest visible ground item with id N, or Null
//	world.ground_items.by_id(N, radius=R) — same, restricted to within R tiles
//	world.ground_items.nearest           — closest ground item to self (any id), or Null
//	world.ground_items.nearest(pos)      — closest ground item to an explicit position (#117)
//	world.ground_items.most_valuable     — highest-base-value visible item, or Null (#117)
//
// Implements Iterable so `for gi in world.ground_items { ... }` works
// the same as the previous list-returning shape.
type groundItemsView struct{ host *Host }

func (g *groundItemsView) Kind() string    { return "ground_items" }
func (g *groundItemsView) Display() string { return "<ground_items>" }

func (g *groundItemsView) Iter() []interp.Value {
	records := g.host.world.GroundItems.All()
	out := make([]interp.Value, 0, len(records))
	for _, r := range records {
		out = append(out, &groundItemView{record: r, facts: g.host.facts})
	}
	return out
}

func (g *groundItemsView) Get(field string) (interp.Value, bool) {
	switch field {
	case "by_id":
		return &groundItemsByIDCallable{host: g.host}, true
	case "nearest":
		// Dual-mode (#117): `.nearest` (bare field) keeps the legacy
		// shape — the closest item to self.position, or Null when no
		// items are visible. `.nearest(pos)` recenters on an explicit
		// position. When items exist we return a wrapper that both
		// delegates field reads to the by-self nearest item AND is
		// Callable; when none are visible we return Null directly so
		// the common `world.ground_items.nearest == null` guard still
		// reads true.
		records := g.host.world.GroundItems.All()
		if len(records) == 0 {
			return interp.Null{}, true
		}
		base := g.nearest(records).(*groundItemView)
		return &groundItemsNearestValue{host: g.host, base: base}, true
	case "most_valuable":
		// Value-sorted selector (#117): the visible ground item with
		// the highest ItemDef base value, or Null when none visible /
		// none resolvable. Enables loot-most-valuable without an
		// author-side fold over world.ground_items.
		return g.mostValuable(g.host.world.GroundItems.All()), true
	case "all":
		return &interp.List{Items: g.Iter()}, true
	case "length":
		return interp.Int(int64(len(g.Iter()))), true
	}
	return nil, false
}

// Index supports `world.ground_items[N]` for backwards compatibility
// with the previous list-returning shape.
func (g *groundItemsView) Index(idx interp.Value) (interp.Value, bool) {
	i, ok := interp.AsInt(idx)
	if !ok {
		return nil, false
	}
	items := g.Iter()
	if int(i) < 0 || int(i) >= len(items) {
		return nil, false
	}
	return items[i], true
}

// nearest returns the closest ground item to self.position, or Null
// if records is empty.
func (g *groundItemsView) nearest(records []world.GroundItemRecord) interp.Value {
	if len(records) == 0 {
		return interp.Null{}
	}
	pos := g.host.world.Self.Position()
	bestIdx := 0
	bestDist := chebyshev(pos.X, pos.Y, records[0].X, records[0].Y)
	for i := 1; i < len(records); i++ {
		d := chebyshev(pos.X, pos.Y, records[i].X, records[i].Y)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return &groundItemView{record: records[bestIdx], facts: g.host.facts}
}

// mostValuable returns the visible ground item with the highest
// ItemDef base value (facts.ItemDef.BasePrice), or Null when records
// is empty or facts can't resolve any of them. Ties resolve to the
// first encountered. Items whose def isn't loaded score 0 — they only
// win if nothing else resolves (and then BasePrice 0 still beats the
// initial sentinel only when at least one record is present).
func (g *groundItemsView) mostValuable(records []world.GroundItemRecord) interp.Value {
	if len(records) == 0 {
		return interp.Null{}
	}
	bestIdx := -1
	bestVal := -1
	for i, r := range records {
		val := 0
		if g.host.facts != nil {
			if def := g.host.facts.ItemDef(r.ItemID); def != nil {
				val = def.BasePrice
			}
		}
		if val > bestVal {
			bestVal = val
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return interp.Null{}
	}
	return &groundItemView{record: records[bestIdx], facts: g.host.facts}
}

// groundItemsNearestValue backs `world.ground_items.nearest`. It is
// both a Getter (delegating to the by-self nearest GroundItem so
// `.nearest.id` / `.nearest.position` keep working) and a Callable
// (so `.nearest(pos)` recenters on an explicit position). See the
// dual-mode note in groundItemsView.Get.
type groundItemsNearestValue struct {
	host *Host
	base *groundItemView // the by-self nearest item (never nil here)
}

func (n *groundItemsNearestValue) Kind() string    { return n.base.Kind() }
func (n *groundItemsNearestValue) Display() string { return n.base.Display() }

func (n *groundItemsNearestValue) Get(field string) (interp.Value, bool) {
	return n.base.Get(field)
}

func (n *groundItemsNearestValue) Index(idx interp.Value) (interp.Value, bool) {
	return nil, false
}

// Call recenters the nearest-item search on the supplied position
// argument. With no args it falls back to self.position (identical to
// the bare-field read). Accepts a single position-like value
// (anything carrying .x/.y, e.g. self.position) or two Int args (x, y).
func (n *groundItemsNearestValue) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	all := n.host.world.GroundItems.All()
	if len(args) == 0 && len(named) == 0 {
		// Same as the bare field.
		return (&groundItemsView{host: n.host}).nearest(all), nil
	}
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("world.ground_items.nearest: %v", err)
	}
	if len(all) == 0 {
		return interp.Null{}, nil
	}
	bestIdx := 0
	bestDist := chebyshev(x, y, all[0].X, all[0].Y)
	for i := 1; i < len(all); i++ {
		d := chebyshev(x, y, all[i].X, all[i].Y)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return &groundItemView{record: all[bestIdx], facts: n.host.facts}, nil
}

// groundItemsByIDCallable backs `world.ground_items.by_id(N, radius=R?)`.
type groundItemsByIDCallable struct{ host *Host }

func (c *groundItemsByIDCallable) Kind() string    { return "callable" }
func (c *groundItemsByIDCallable) Display() string { return "<world.ground_items.by_id>" }
func (c *groundItemsByIDCallable) Yields() bool    { return false }

func (c *groundItemsByIDCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 {
		return nil, errf("world.ground_items.by_id needs an item id")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("world.ground_items.by_id: expected Int item_id")
	}
	radius := -1
	if v, ok := named["radius"]; ok {
		if r, ok := interp.AsInt(v); ok {
			radius = int(r)
		}
	}
	pos := c.host.world.Self.Position()
	all := c.host.world.GroundItems.All()
	var matches []world.GroundItemRecord
	for _, r := range all {
		if r.ItemID != int(id) {
			continue
		}
		if radius >= 0 && chebyshev(pos.X, pos.Y, r.X, r.Y) > radius {
			continue
		}
		matches = append(matches, r)
	}
	if len(matches) == 0 {
		return interp.Null{}, nil
	}
	v := (&groundItemsView{host: c.host}).nearest(matches)
	return v, nil
}

type groundItemView struct {
	record world.GroundItemRecord
	facts  *facts.Facts
}

func (g *groundItemView) Kind() string { return "ground_item" }
func (g *groundItemView) Display() string {
	return itemName(g.facts, g.record.ItemID) + "@(" + intDisp(g.record.X) + "," + intDisp(g.record.Y) + ")"
}
func (g *groundItemView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id", "item_id":
		return interp.Int(int64(g.record.ItemID)), true
	case "x":
		return interp.Int(int64(g.record.X)), true
	case "y":
		return interp.Int(int64(g.record.Y)), true
	case "position":
		return &positionView{X: g.record.X, Y: g.record.Y}, true
	case "def":
		// The static catalog entry (ItemDef), api.md §2. Same ItemDef
		// an InvSlot of this item would carry. Null if facts not
		// loaded / id unknown.
		if g.facts != nil {
			if def := g.facts.ItemDef(g.record.ItemID); def != nil {
				return &itemDefView{def: def}, true
			}
		}
		return interp.Null{}, true
	case "name":
		return interp.String(itemName(g.facts, g.record.ItemID)), true

	// Stub until the host tracks 3-min ground-item ownership
	// windows. Routines that pick up "their" loot should check
	// is_mine — currently always false (safe default — routines
	// will fall through to "try and see" with pick_up()'s
	// SERVER_REJECTED error code).
	case "is_mine":
		return interp.Bool(false), true
	}
	return nil, false
}

// ---------- dynamic scenery (world.scenery) ----------

// sceneryView is the entry point for dynamic scenery (GameObject)
// queries — the in-view objects the server streams via
// SEND_SCENERY_HANDLER (opcode 48), which is the ONLY place
// runtime-spawned scenery surfaces. A fire lit by firemaking (def 97)
// is registered server-side at runtime and never appears in the
// static facts.SceneryLocs map — so world.locs.search("fire") finds
// only PRE-PLACED fires, never one you just lit. world.scenery is the
// live mirror; use it to find a fire/range you just created:
//
//	world.scenery                       — all visible dynamic scenery
//	world.scenery.by_id(97)             — nearest visible scenery def 97 (a fire)
//	world.scenery.by_id(97, radius=2)   — same, within 2 tiles
//	world.scenery.nearest               — closest scenery to self
//	world.scenery.nearest(pos)          — closest to an explicit position
//
// Results are placement views (kind="scenery") so they drop straight
// into use(item, scenery) → USE_ITEM_ON_SCENERY (opcode 115), e.g.
// use(raw_beef, world.scenery.by_id(97)).
type sceneryView struct{ host *Host }

func (s *sceneryView) Kind() string    { return "scenery" }
func (s *sceneryView) Display() string { return "<scenery>" }

func (s *sceneryView) Iter() []interp.Value {
	records := s.host.world.Scenery.All()
	out := make([]interp.Value, 0, len(records))
	for _, r := range records {
		out = append(out, s.placement(r))
	}
	return out
}

// placement builds a placementView from a dynamic scenery record,
// resolving the def name from facts so .name reads correctly and the
// kind="scenery" routes use() to opcode 115.
func (s *sceneryView) placement(r world.SceneryRecord) *placementView {
	name := "scenery"
	if s.host.facts != nil {
		if def := s.host.facts.SceneryDef(r.ID); def != nil {
			name = def.Name
		}
	}
	return &placementView{p: facts.Placement{
		Kind:  "scenery",
		DefID: r.ID,
		Name:  name,
		X:     r.X,
		Y:     r.Y,
	}}
}

func (s *sceneryView) Get(field string) (interp.Value, bool) {
	switch field {
	case "by_id":
		return &sceneryByIDCallable{host: s.host}, true
	case "nearest":
		records := s.host.world.Scenery.All()
		if len(records) == 0 {
			return interp.Null{}, true
		}
		return s.nearest(records, nil), true
	case "all":
		return &interp.List{Items: s.Iter()}, true
	case "length":
		return interp.Int(int64(len(s.host.world.Scenery.All()))), true
	}
	return nil, false
}

func (s *sceneryView) Index(idx interp.Value) (interp.Value, bool) {
	i, ok := interp.AsInt(idx)
	if !ok {
		return nil, false
	}
	items := s.Iter()
	if int(i) < 0 || int(i) >= len(items) {
		return nil, false
	}
	return items[i], true
}

// nearest returns the scenery record closest to `center` (or
// self.position when center is nil), as a placementView. Returns Null
// when records is empty.
func (s *sceneryView) nearest(records []world.SceneryRecord, center *positionView) interp.Value {
	if len(records) == 0 {
		return interp.Null{}
	}
	cx, cy := 0, 0
	if center != nil {
		cx, cy = center.X, center.Y
	} else {
		pos := s.host.world.Self.Position()
		cx, cy = pos.X, pos.Y
	}
	bestIdx := 0
	bestDist := chebyshev(cx, cy, records[0].X, records[0].Y)
	for i := 1; i < len(records); i++ {
		d := chebyshev(cx, cy, records[i].X, records[i].Y)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return s.placement(records[bestIdx])
}

// sceneryByIDCallable backs `world.scenery.by_id(N, radius=R?)`.
type sceneryByIDCallable struct{ host *Host }

func (c *sceneryByIDCallable) Kind() string    { return "callable" }
func (c *sceneryByIDCallable) Display() string { return "<world.scenery.by_id>" }
func (c *sceneryByIDCallable) Yields() bool    { return false }

func (c *sceneryByIDCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 {
		return nil, errf("world.scenery.by_id needs a scenery def id")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("world.scenery.by_id: expected Int def_id")
	}
	radius := -1
	if v, ok := named["radius"]; ok {
		if r, ok := interp.AsInt(v); ok {
			radius = int(r)
		}
	}
	pos := c.host.world.Self.Position()
	all := c.host.world.Scenery.All()
	var matches []world.SceneryRecord
	for _, r := range all {
		if r.ID != int(id) {
			continue
		}
		if radius >= 0 && chebyshev(pos.X, pos.Y, r.X, r.Y) > radius {
			continue
		}
		matches = append(matches, r)
	}
	if len(matches) == 0 {
		return interp.Null{}, nil
	}
	sv := &sceneryView{host: c.host}
	return sv.nearest(matches, nil), nil
}

// ---------- boundaries ----------

// boundariesView is the entry point for boundary queries:
// `world.boundaries.at(x=103, y=532, dir=0)` returns a boundary
// view that can be passed to use(), open_boundary(), etc.
//
// Today it only supports `.at(x, y, dir)` for direct lookup of
// boundary loc data via facts.Facts. Future: `.near(pos, radius)`
// returning a list of nearby openable boundaries, `.find(predicate)`
// with lambda filtering.
type boundariesView struct{ host *Host }

func (b *boundariesView) Kind() string    { return "boundaries" }
func (b *boundariesView) Display() string { return "<boundaries>" }

func (b *boundariesView) Get(field string) (interp.Value, bool) {
	switch field {
	case "at":
		return &boundaryAtCallable{host: b.host}, true
	case "near":
		// world.boundaries.near(radius) → nearest-first list of openable
		// boundary views (doors/gates) from facts, each with its real
		// direction, ready to pass to open_boundary(). Default radius 8.
		return &boundaryNearCallable{host: b.host}, true
	case "is_open":
		// world.boundaries.is_open(x, y, dir) — true iff we've seen
		// a SEND_BOUNDARY_HANDLER mark this tile/dir as removed
		// (door opened, web cut). Returns false for unknown tiles
		// (caller falls back to walk_to or pathfinder rejection).
		return &boundaryIsOpenCallable{host: b.host}, true
	case "dynamic":
		// Snapshot of all dynamic boundary overrides as a list of
		// {x, y, dir, id}. Useful for debugging / persistence.
		all := b.host.world.Boundaries.All()
		items := make([]interp.Value, 0, len(all))
		for k, id := range all {
			items = append(items, &interp.List{Items: []interp.Value{
				interp.Int(int64(k.X)),
				interp.Int(int64(k.Y)),
				interp.Int(int64(k.Dir)),
				interp.Int(int64(id)),
			}})
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// boundaryIsOpenCallable backs world.boundaries.is_open(x, y, dir).
type boundaryIsOpenCallable struct{ host *Host }

func (c *boundaryIsOpenCallable) Kind() string    { return "callable" }
func (c *boundaryIsOpenCallable) Display() string { return "<world.boundaries.is_open>" }
func (c *boundaryIsOpenCallable) Yields() bool    { return false }

func (c *boundaryIsOpenCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, dir, err := resolveBoundaryAt(args, named)
	if err != nil {
		return nil, err
	}
	return interp.Bool(c.host.world.Boundaries.IsRemoved(x, y, dir)), nil
}

// boundaryAtCallable backs `world.boundaries.at(x=X, y=Y, dir=D)`.
// Returns a boundaryView, or Null if no boundary def is known at
// that (x, y, dir) — the facts index drives the lookup; if the
// landscape encodes a wall via the .orsc tile bytes but the
// BoundaryLocs JSON doesn't list it, we still synthesize a view
// from the coords so use(key, door) can fire regardless.
type boundaryAtCallable struct{ host *Host }

func (c *boundaryAtCallable) Kind() string    { return "callable" }
func (c *boundaryAtCallable) Display() string { return "<world.boundaries.at>" }
func (c *boundaryAtCallable) Yields() bool    { return false }

func (c *boundaryAtCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, dir, err := resolveBoundaryAt(args, named)
	if err != nil {
		return nil, err
	}
	// Even if facts doesn't know about this exact (x, y, dir),
	// return a synthetic view — callers may know about a door
	// the static facts data missed (and use() works either way
	// since the server validates the click).
	return &boundaryView{x: x, y: y, direction: dir, host: c.host}, nil
}

// boundaryNearCallable backs world.boundaries.near(radius) — a nearest-first
// list of nearby openable boundary views (doors/gates) drawn from facts, each
// carrying its real direction so open_boundary() fires the correct click.
type boundaryNearCallable struct{ host *Host }

func (c *boundaryNearCallable) Kind() string    { return "callable" }
func (c *boundaryNearCallable) Display() string { return "<world.boundaries.near>" }
func (c *boundaryNearCallable) Yields() bool    { return false }

func (c *boundaryNearCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	radius := 8
	if len(args) >= 1 {
		if r, ok := args[0].(interp.Int); ok {
			radius = int(r)
		}
	}
	if c.host.facts == nil || c.host.world == nil || c.host.world.Self == nil {
		return &interp.List{}, nil
	}
	pos := c.host.world.Self.Position()
	type bd struct {
		v    *boundaryView
		dist int
	}
	var bds []bd
	for _, p := range c.host.facts.Near(pos.X, pos.Y, radius) {
		if p.Kind != "boundary" {
			continue
		}
		// Only OPENABLE boundaries (doors/gates): Unknown==1 marks them; plain
		// walls/fences are Unknown==0 and must not be returned (opening a wall is
		// a no-op that strands the caller).
		def := c.host.facts.BoundaryDef(p.DefID)
		if def == nil || def.Unknown != 1 {
			continue
		}
		bds = append(bds, bd{
			v:    &boundaryView{x: p.X, y: p.Y, direction: p.Direction, host: c.host},
			dist: absInt(p.X-pos.X) + absInt(p.Y-pos.Y),
		})
	}
	sort.Slice(bds, func(i, j int) bool { return bds[i].dist < bds[j].dist })
	items := make([]interp.Value, 0, len(bds))
	for _, b := range bds {
		items = append(items, b.v)
	}
	return &interp.List{Items: items}, nil
}

// resolveBoundaryAt parses (x, y, dir) from positional or named
// args. Accepts:
//
//	at(x, y, dir)             // positional
//	at(x=X, y=Y, dir=D)       // named
//
// `dir` defaults to 0 if omitted.
func resolveBoundaryAt(args []interp.Value, named map[string]interp.Value) (int, int, int, error) {
	var x, y, dir int
	if len(args) >= 2 {
		if xi, ok := args[0].(interp.Int); ok {
			x = int(xi)
		} else {
			return 0, 0, 0, fmt.Errorf("world.boundaries.at: positional x must be Int")
		}
		if yi, ok := args[1].(interp.Int); ok {
			y = int(yi)
		} else {
			return 0, 0, 0, fmt.Errorf("world.boundaries.at: positional y must be Int")
		}
		if len(args) >= 3 {
			if di, ok := args[2].(interp.Int); ok {
				dir = int(di)
			}
		}
	}
	if v, ok := named["x"]; ok {
		if i, ok := v.(interp.Int); ok {
			x = int(i)
		}
	}
	if v, ok := named["y"]; ok {
		if i, ok := v.(interp.Int); ok {
			y = int(i)
		}
	}
	if v, ok := named["dir"]; ok {
		if i, ok := v.(interp.Int); ok {
			dir = int(i)
		}
	}
	return x, y, dir, nil
}

// boundaryView is a DSL-visible reference to a boundary tile +
// direction. Carries enough info for use() / open_boundary() to
// dispatch the right packet. Construct via `world.boundaries.at()`.
type boundaryView struct {
	x, y, direction int
	host            *Host
}

func (b *boundaryView) Kind() string { return "boundary" }
func (b *boundaryView) Display() string {
	return fmt.Sprintf("<boundary (%d, %d) dir=%d>", b.x, b.y, b.direction)
}

func (b *boundaryView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(int64(b.x)), true
	case "y":
		return interp.Int(int64(b.y)), true
	case "direction", "dir":
		return interp.Int(int64(b.direction)), true
	case "position":
		return &positionView{X: b.x, Y: b.y}, true
	case "name":
		// If facts knows a def at this tile, surface its name.
		if b.host != nil && b.host.facts != nil {
			for _, p := range b.host.facts.At(b.x, b.y) {
				if p.Kind == "boundary" && p.Direction == b.direction {
					return interp.String(p.Name), true
				}
			}
		}
		return interp.String("door"), true
	case "door_type":
		if d := b.def(); d != nil {
			return interp.Int(int64(d.DoorType)), true
		}
		return interp.Int(0), true
	case "is_openable":
		// True for doors/doorframes the bot can open. When the def is
		// unknown, assume openable (conservative for a thing named "door").
		if d := b.def(); d != nil {
			return interp.Bool(d.IsOpenable()), true
		}
		return interp.Bool(true), true
	case "blocks_when_closed":
		// True for closeable doors (DoorType==1) that block while closed —
		// pathfind to them and open before crossing. Unknown def -> false.
		if d := b.def(); d != nil {
			return interp.Bool(d.BlocksWhenClosed()), true
		}
		return interp.Bool(false), true
	}
	return nil, false
}

// def resolves the BoundaryDef known at this boundary's tile + direction,
// or nil if facts has none. Powers the name / door-state field accessors.
func (b *boundaryView) def() *facts.BoundaryDef {
	if b.host == nil || b.host.facts == nil {
		return nil
	}
	for _, p := range b.host.facts.At(b.x, b.y) {
		if p.Kind == "boundary" && p.Direction == b.direction {
			return b.host.facts.BoundaryDef(p.DefID)
		}
	}
	return nil
}

// ---------- locs (facts-backed location queries) ----------

// locsView exposes `world.locs.banks`, `world.locs.fishing_spots`,
// etc. as searchable lists.
//
// Implementation strategy: the DSL field name (e.g. "banks") maps to
// a set of facts-known categories. Each returns a *locListView that
// supports `.nearest(point)` and iteration.
type locsView struct{ host *Host }

func (l *locsView) Kind() string    { return "locs" }
func (l *locsView) Display() string { return "<locs>" }

func (l *locsView) Get(field string) (interp.Value, bool) {
	// Some categories are well-known names. Others are queried by
	// substring match in scenery/boundary/NPC names. We start with
	// the common ones; new categories are easy to add.
	switch field {
	case "banks":
		return l.searchByName("bank"), true
	case "altars":
		return l.searchByName("altar"), true
	case "furnaces":
		return l.searchByName("furnace"), true
	case "anvils":
		return l.searchByName("anvil"), true
	case "fishing_spots":
		return l.searchByName("fishing"), true
	case "trees":
		return l.searchByName("tree"), true
	case "rocks":
		return l.searchByName("rock"), true
	case "doors":
		return l.searchBoundaryByName("door"), true
	case "ladders":
		// Ladders are SCENERY (GameObjectDef "Ladder", e.g.
		// LADDER_GENERIC_UP=5 / _DOWN=6), not boundaries — they are
		// climbed via the scenery Command1 "Climb-up"/"Climb-down"
		// (interact_at / ObjectCommand opcode 136), not opened like a
		// door. searchBoundaryByName never matched anything (DoorDef
		// has no ladders), so this returned an empty set everywhere.
		return l.searchByName("ladder"), true
	case "shops":
		return l.searchByName("shop"), true
	case "scenery":
		// Every scenery def — useful with .within(radius) to find
		// all nearby objects regardless of kind.
		return l.allOfKind("scenery"), true
	case "spawn_points":
		// Every NPC spawn point. Distinct from world.npcs (which is
		// currently-visible NPCs); spawn_points is "where NPCs
		// originate" — useful for "walk to the goblin spawn area."
		return l.allOfKind("npc_spawn"), true
	case "search":
		// Generic substring search for scenery / boundary / spawn
		// defs by name. Returns a locListView (chain .nearest /
		// .within / .names like the named categories).
		return locsSearchCallable{l: l}, true
	}
	return nil, false
}

// locsSearchCallable: `world.locs.search("furnace").nearest(self.position)`.
type locsSearchCallable struct{ l *locsView }

func (c locsSearchCallable) Kind() string    { return "builtin" }
func (c locsSearchCallable) Display() string { return "<locs.search>" }
func (c locsSearchCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("locs.search takes 1 arg (needle string)")
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("locs.search: needle must be String, got %s", args[0].Kind())
	}
	return c.l.searchByName(string(needle)), nil
}

// allOfKind populates a locListView with every def of the given
// facts kind ("scenery" / "npc_spawn" / "boundary"). Used by
// world.locs.scenery and world.locs.spawn_points — categories
// that don't filter by name but by kind alone.
func (l *locsView) allOfKind(kind string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	switch kind {
	case "scenery":
		for _, def := range l.host.facts.SceneryDefs {
			if def != nil {
				out.kinds = append(out.kinds, "scenery")
				out.names = append(out.names, def.Name)
			}
		}
	case "npc_spawn":
		for _, def := range l.host.facts.NpcDefs {
			if def != nil {
				out.kinds = append(out.kinds, "npc_spawn")
				out.names = append(out.names, def.Name)
			}
		}
	case "boundary":
		for _, def := range l.host.facts.BoundaryDefs {
			if def != nil {
				out.kinds = append(out.kinds, "boundary")
				out.names = append(out.names, def.Name)
			}
		}
	}
	return out
}

// searchByName collects scenery placements whose name contains
// the given substring (case-insensitive). If no scenery matches,
// falls back to NPC defs (e.g. world.locs.banks will hit the
// "Banker" NPC if no bank-booth scenery is defined). This
// preference matters because routines almost always want the
// physical destination (the bank counter), not the NPC standing
// next to it.
//
// Discovered live 2026-05-28: a substring match on "bank" hit
// both the Bank booth scenery AND the Banker NPC; routines that
// then called walk_to ended up at the NPC's spawn instead of
// the counter. Preferring scenery resolves that.
func (l *locsView) searchByName(needle string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	needle = strings.ToLower(needle)
	for i := range l.host.facts.SceneryDefs {
		def := l.host.facts.SceneryDefs[i]
		if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
			out.kinds = append(out.kinds, "scenery")
			out.names = append(out.names, def.Name)
		}
	}
	// Fall back to NPC defs only when scenery yielded nothing.
	if len(out.names) == 0 {
		for i := range l.host.facts.NpcDefs {
			def := l.host.facts.NpcDefs[i]
			if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
				out.kinds = append(out.kinds, "npc_spawn")
				out.names = append(out.names, def.Name)
			}
		}
	}
	return out
}

func (l *locsView) searchBoundaryByName(needle string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	needle = strings.ToLower(needle)
	for i := range l.host.facts.BoundaryDefs {
		def := l.host.facts.BoundaryDefs[i]
		if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
			out.kinds = append(out.kinds, "boundary")
			out.names = append(out.names, def.Name)
		}
	}
	return out
}

// locListView wraps a name+kind list and resolves to actual
// placements lazily via facts.NearestByName when `.nearest(point)`
// is called. We don't materialize every placement up front because
// some categories (e.g. trees) have thousands of instances.
type locListView struct {
	host  *Host
	kinds []string
	names []string
}

func (v *locListView) Kind() string    { return "loc_list" }
func (v *locListView) Display() string { return "<loc_list:" + intDisp(len(v.names)) + ">" }

func (v *locListView) Get(field string) (interp.Value, bool) {
	switch field {
	case "length":
		// Approximate — we don't know the actual placement count
		// without scanning, so return the number of *kinds* that
		// match. Routines should prefer .nearest() over .length.
		return interp.Int(int64(len(v.names))), true
	case "nearest":
		return locNearestCallable{owner: v}, true
	case "within":
		return locWithinCallable{owner: v}, true
	case "names":
		items := make([]interp.Value, 0, len(v.names))
		for _, n := range v.names {
			items = append(items, interp.String(n))
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// locNearestCallable: `locs.banks.nearest(self.position)` -> placement
type locNearestCallable struct{ owner *locListView }

func (c locNearestCallable) Kind() string    { return "builtin" }
func (c locNearestCallable) Display() string { return "<locs.nearest>" }
func (c locNearestCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, err
	}
	if c.owner.host.facts == nil {
		return interp.Null{}, nil
	}
	const radius = 64
	var best *facts.Placement
	bestD := -1
	for i, kind := range c.owner.kinds {
		p := c.owner.host.facts.NearestByName(c.owner.names[i], x, y, radius, kind)
		if p == nil {
			continue
		}
		d := chebyshev(p.X, p.Y, x, y)
		if best == nil || d < bestD {
			best = p
			bestD = d
		}
	}
	if best == nil {
		return interp.Null{}, nil
	}
	return &placementView{p: *best}, nil
}

// locWithinCallable: `world.locs.banks.within(radius)` →
// list<placement> of placements of the matching def(s) within
// the Chebyshev radius of self.position. Useful for "find every
// fishing spot within 20 tiles."
//
// Signature variants accepted:
//
//	.within(radius)              — center = self.position
//	.within(radius, position)    — center = supplied position
//
// Returns an empty list when facts aren't loaded.
type locWithinCallable struct{ owner *locListView }

func (c locWithinCallable) Kind() string    { return "builtin" }
func (c locWithinCallable) Display() string { return "<locs.within>" }
func (c locWithinCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) == 0 {
		return nil, errf("locs.within requires at least a radius")
	}
	r, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("locs.within: first arg must be int radius, got %s", args[0].Kind())
	}
	radius := int(r)
	if radius < 0 {
		return nil, errf("locs.within: radius must be >= 0, got %d", radius)
	}
	// Center defaults to host's current position.
	cx, cy := 0, 0
	if c.owner.host != nil {
		p := c.owner.host.world.Self.Position()
		cx, cy = p.X, p.Y
	}
	// Optional position override.
	if len(args) >= 2 {
		if g, ok := args[1].(interp.Getter); ok {
			if xv, ok := g.Get("x"); ok {
				if x, ok := interp.AsInt(xv); ok {
					cx = int(x)
				}
			}
			if yv, ok := g.Get("y"); ok {
				if y, ok := interp.AsInt(yv); ok {
					cy = int(y)
				}
			}
		}
	}
	if c.owner.host == nil || c.owner.host.facts == nil {
		return &interp.List{}, nil
	}
	// Build set of (kind, name) we're matching, then scan
	// facts.Near for matches.
	type matcher struct{ kind, name string }
	wants := make(map[matcher]struct{}, len(c.owner.names))
	for i, name := range c.owner.names {
		wants[matcher{c.owner.kinds[i], name}] = struct{}{}
	}
	hits := c.owner.host.facts.Near(cx, cy, radius)
	out := make([]interp.Value, 0, len(hits))
	for _, p := range hits {
		if _, want := wants[matcher{p.Kind, p.Name}]; want {
			placement := p
			out = append(out, &placementView{p: placement})
		}
	}
	return &interp.List{Items: out}, nil
}

// placementView wraps a facts.Placement so DSL code can read
// `.x`, `.y`, `.name`, etc.
type placementView struct{ p facts.Placement }

func (v *placementView) Kind() string    { return "placement" }
func (v *placementView) Display() string { return v.p.Name }
func (v *placementView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(int64(v.p.X)), true
	case "y":
		return interp.Int(int64(v.p.Y)), true
	case "name":
		return interp.String(v.p.Name), true
	case "kind":
		return interp.String(v.p.Kind), true
	case "def_id":
		return interp.Int(int64(v.p.DefID)), true
	case "direction":
		return interp.Int(int64(v.p.Direction)), true
	case "position":
		return &positionView{X: v.p.X, Y: v.p.Y}, true
	}
	return nil, false
}
