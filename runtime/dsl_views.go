package runtime

import (
	"fmt"
	"math"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// This file defines the entity views that `.routine` code sees behind
// the reserved names `self`, `world`, `inventory`, and `combat`.
//
// Each view is an interp.Value that implements interp.Getter (and
// sometimes interp.Indexer or interp.Callable) so member access in
// the DSL translates directly to method calls on the underlying
// runtime / world / facts state.
//
// Views never cache. They reflect the latest world-state snapshot at
// the moment of access — the bot's world mirror is updated by the
// inbound packet loop, so a DSL `while inventory.free > 0` loop sees
// fresh counts each iteration.

// ---------- self ----------

// selfView exposes `self.*` to the DSL.
type selfView struct{ host *Host }

func (s *selfView) Kind() string    { return "self" }
func (s *selfView) Display() string { return "<self>" }

func (s *selfView) Get(field string) (interp.Value, bool) {
	self := s.host.world.Self
	inv := s.host.world.Inventory
	switch field {
	// Position
	case "position":
		p := self.Position()
		return &positionView{X: p.X, Y: p.Y}, true

	// Vitals — current / max / derived
	case "hp":
		return interp.Int(self.HP()), true
	case "max_hp":
		return interp.Int(self.MaxHP()), true
	case "hp_fraction":
		max := self.MaxHP()
		if max <= 0 {
			return interp.Float(0), true
		}
		return interp.Float(float64(self.HP()) / float64(max)), true
	case "prayer":
		// Shorthand for the prayer skill's current level. Use
		// self.prayers for the active-bitmap accessor.
		return interp.Int(self.Prayer()), true
	case "max_prayer":
		return interp.Int(self.MaxPrayer()), true
	case "prayers":
		return &prayersView{host: s.host}, true
	case "fatigue":
		return interp.Int(self.Fatigue()), true
	case "combat_level":
		return interp.Int(self.CombatLevel()), true
	case "quest_points":
		return interp.Int(self.QuestPoints()), true

	// Death/respawn tracking. last_death_at is captured during
	// Apply(event.Death) BEFORE the respawn position packet
	// overwrites self.position — so reading it from an
	// `on death` handler gives the death tile, not the spawn tile.
	case "last_death_at":
		c := self.LastDeathAt()
		if c.X == 0 && c.Y == 0 {
			return interp.Null{}, true
		}
		return &positionView{X: c.X, Y: c.Y}, true
	case "death_count":
		return interp.Int(int64(self.DeathCount())), true

	// State booleans — wire to real state tracking as it lands.
	// Returning false is the safe default (routines branching on
	// these won't misfire). Tracked separately as Stage 2
	// follow-ups: requires a per-host "current action" registry
	// (is_busy), combat-target observation (is_in_combat), and
	// a SleepScreen state field (is_sleeping).
	case "is_busy":
		return interp.Bool(false), true
	case "is_in_combat":
		return interp.Bool(false), true
	case "is_sleeping":
		return interp.Bool(false), true

	// Skills + equipped
	case "skills":
		return &skillsView{host: s.host}, true
	case "wielded":
		for _, slot := range inv.Slots() {
			if slot.Wielded {
				return s.itemView(slot.ItemID, slot.Amount), true
			}
		}
		return interp.Null{}, true
	case "equipped":
		// All currently-wielded items as a list of item-views.
		// Per-slot accessors (.weapon / .shield / .head / etc.)
		// require decoding the equipment-by-slot packet (not yet
		// wired) — until then routines that want a specific
		// slot iterate / filter the list themselves.
		items := make([]interp.Value, 0)
		for _, slot := range inv.Slots() {
			if slot.Wielded {
				items = append(items, s.itemView(slot.ItemID, slot.Amount))
			}
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// itemView builds an item value, looking up the human-readable name
// from facts (if available). Tiny helper used by self.wielded and
// inventory.slots.
func (s *selfView) itemView(itemID, amount int) interp.Value {
	return newItemView(s.host.facts, itemID, amount)
}

// positionView is `{x, y}` for self.position.
type positionView struct{ X, Y int }

func (p *positionView) Kind() string    { return "position" }
func (p *positionView) Display() string { return "(" + intDisp(p.X) + "," + intDisp(p.Y) + ")" }
func (p *positionView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(p.X), true
	case "y":
		return interp.Int(p.Y), true
	}
	return nil, false
}

// skillsView is the parent for `self.skills.<name>` access. Each
// named skill resolves to a skillView which then exposes the
// 5 per-skill fields (level/max_level/xp/xp_to_next_level/
// percent_to_next_level) per docs/lang/state.md row 2.
type skillsView struct{ host *Host }

func (s *skillsView) Kind() string    { return "skills" }
func (s *skillsView) Display() string { return "<skills>" }

// skillIDs maps DSL skill names to their RSC catalog index.
var skillIDs = map[string]int{
	"attack":      0,
	"defense":     1,
	"strength":    2,
	"hits":        3,
	"hitpoints":   3, // alias
	"ranged":      4,
	"prayer":      5,
	"magic":       6,
	"cooking":     7,
	"woodcutting": 8,
	"fletching":   9,
	"fishing":     10,
	"firemaking":  11,
	"crafting":    12,
	"smithing":    13,
	"mining":      14,
	"herblaw":     15,
	"agility":     16,
	"thieving":    17,
}

func (s *skillsView) Get(field string) (interp.Value, bool) {
	id, ok := skillIDs[strings.ToLower(field)]
	if !ok {
		return nil, false
	}
	return &skillView{host: s.host, id: id, name: field}, true
}

// skillView is one skill's per-field accessor: level / max_level /
// xp / xp_to_next_level / percent_to_next_level. Lookups go through
// the host's world.Self mirror.
type skillView struct {
	host *Host
	id   int
	name string
}

func (s *skillView) Kind() string    { return "skill" }
func (s *skillView) Display() string { return s.name }
func (s *skillView) Get(field string) (interp.Value, bool) {
	self := s.host.world.Self
	level := self.SkillLevel(s.id)
	maxLevel := self.SkillMax(s.id)
	xp := self.SkillXP(s.id)
	switch field {
	case "level":
		return interp.Int(int64(level)), true
	case "max_level":
		return interp.Int(int64(maxLevel)), true
	case "xp":
		return interp.Int(int64(xp)), true
	case "xp_to_next_level":
		nextThreshold := xpThresholdForLevel(maxLevel + 1)
		if nextThreshold <= xp {
			return interp.Int(0), true
		}
		return interp.Int(int64(nextThreshold - xp)), true
	case "percent_to_next_level":
		thisThreshold := xpThresholdForLevel(maxLevel)
		nextThreshold := xpThresholdForLevel(maxLevel + 1)
		span := nextThreshold - thisThreshold
		if span <= 0 {
			return interp.Float(1.0), true
		}
		progress := xp - thisThreshold
		if progress < 0 {
			progress = 0
		}
		if progress > span {
			progress = span
		}
		return interp.Float(float64(progress) / float64(span)), true
	case "name":
		return interp.String(s.name), true
	case "id":
		return interp.Int(int64(s.id)), true
	}
	return nil, false
}

// xpThresholdForLevel returns the total experience required to
// REACH a given level. Standard RSC formula:
//
//	XP(L) = floor(sum_{i=1}^{L-1} floor(i + 300 * 2^(i/7))) / 4
//
// XP(1) is 0; XP(2) is 83; XP(99) is 13,034,431.
//
// Precomputed up to level 100 so per-call cost is O(1).
func xpThresholdForLevel(level int) int {
	if level < 1 {
		return 0
	}
	if level >= len(xpThresholds) {
		return xpThresholds[len(xpThresholds)-1]
	}
	return xpThresholds[level]
}

var xpThresholds = func() [100]int {
	// Canonical RSC XP table:
	//   XP(L) = floor( (1/4) * sum_{i=1}^{L-1} floor(i + 300 * 2^(i/7)) )
	// The INNER floor matters — it's applied per term, not once at
	// the end. Without it, level 26 comes out 8742 instead of 8740.
	var t [100]int
	t[1] = 0
	sum := 0.0
	for L := 2; L < 100; L++ {
		i := float64(L - 1)
		sum += math.Floor(i + 300.0*math.Pow(2.0, i/7.0))
		t[L] = int(math.Floor(sum / 4))
	}
	return t
}()

// ---------- inventory ----------

type inventoryView struct{ host *Host }

func (v *inventoryView) Kind() string    { return "inventory" }
func (v *inventoryView) Display() string { return "<inventory>" }

func (v *inventoryView) Get(field string) (interp.Value, bool) {
	inv := v.host.world.Inventory
	switch field {
	case "free":
		return interp.Int(inv.FreeSlots()), true
	case "is_full":
		return interp.Bool(inv.FreeSlots() == 0), true
	case "slots":
		items := make([]interp.Value, 0)
		for _, s := range inv.Slots() {
			items = append(items, newItemView(v.host.facts, s.ItemID, s.Amount))
		}
		return &interp.List{Items: items}, true
	case "has":
		return invHasCallable{host: v.host}, true
	case "count":
		return invCountCallable{host: v.host}, true
	case "find":
		return invFindCallable{host: v.host}, true
	case "slot_of":
		return invSlotOfCallable{host: v.host}, true
	}
	return nil, false
}

// invHasCallable implements `inventory.has(item_name_or_id)`.
type invHasCallable struct{ host *Host }

func (c invHasCallable) Kind() string    { return "builtin" }
func (c invHasCallable) Display() string { return "<inventory.has>" }
func (c invHasCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("inventory.has takes 1 argument, got %d", len(args))
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	return interp.Bool(c.host.world.Inventory.Has(id)), nil
}

type invCountCallable struct{ host *Host }

func (c invCountCallable) Kind() string    { return "builtin" }
func (c invCountCallable) Display() string { return "<inventory.count>" }
func (c invCountCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("inventory.count takes 1 argument, got %d", len(args))
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	return interp.Int(int64(c.host.world.Inventory.Count(id))), nil
}

// invFindCallable implements `inventory.find(item)` — returns the
// first matching inventory slot as an item-view, or null.
type invFindCallable struct{ host *Host }

func (c invFindCallable) Kind() string    { return "builtin" }
func (c invFindCallable) Display() string { return "<inventory.find>" }
func (c invFindCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("inventory.find takes 1 argument, got %d", len(args))
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	for _, s := range c.host.world.Inventory.Slots() {
		if s.ItemID == id {
			return newItemView(c.host.facts, s.ItemID, s.Amount), nil
		}
	}
	return interp.Null{}, nil
}

// invSlotOfCallable implements `inventory.slot_of(item)` — returns
// the slot index (Int) of the first matching slot, or null if not
// found.
type invSlotOfCallable struct{ host *Host }

func (c invSlotOfCallable) Kind() string    { return "builtin" }
func (c invSlotOfCallable) Display() string { return "<inventory.slot_of>" }
func (c invSlotOfCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("inventory.slot_of takes 1 argument, got %d", len(args))
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	for i, s := range c.host.world.Inventory.Slots() {
		if s.ItemID == id {
			return interp.Int(int64(i)), nil
		}
	}
	return interp.Null{}, nil
}

// ---------- combat ----------

type combatView struct{ host *Host }

func (c *combatView) Kind() string    { return "combat" }
func (c *combatView) Display() string { return "<combat>" }

// Combat state isn't tracked yet — runtime/combat_loop.go owns it
// transiently. For now: engaged is false, target is null. Future work
// (step 8) wires the persistent state.
func (c *combatView) Get(field string) (interp.Value, bool) {
	switch field {
	case "engaged":
		return interp.Bool(false), true
	case "target":
		return interp.Null{}, true
	case "last_npc":
		// Most-recently-attacked NPC view (resolved live from
		// world.npcs by stored index). Null if never attacked
		// OR the NPC has since left view / died. Routines flee/
		// retarget on null.
		idx := c.host.lastAttackedNpcIndex
		if idx == 0 {
			return interp.Null{}, true
		}
		for _, n := range c.host.world.Npcs.All() {
			if n.Index == idx {
				return &npcView{record: n, facts: c.host.facts}, true
			}
		}
		return interp.Null{}, true
	case "last_player":
		// Same shape for the last-attacked player. Common in
		// duels — track who we engaged for fleeing/loot.
		idx := c.host.lastAttackedPlayerIndex
		if idx == 0 {
			return interp.Null{}, true
		}
		if rec, ok := c.host.world.Players.Get(idx); ok {
			return &playerView{record: rec}, true
		}
		return interp.Null{}, true
	}
	return nil, false
}

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
			items = append(items, &playerView{record: p})
		}
		return &interp.List{Items: items}, true
	case "npcs":
		records := w.host.world.Npcs.All()
		items := make([]interp.Value, 0, len(records))
		for _, n := range records {
			items = append(items, &npcView{record: n, facts: w.host.facts})
		}
		return &interp.List{Items: items}, true
	case "ground_items":
		return &groundItemsView{host: w.host}, true
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
//   world.dialog.is_open       → bool, true if a menu is presented
//   world.dialog.options       → list<String>, the options (empty if none)
//   world.dialog.find_option(s) → int (0-based index, or -1 if absent)
//   world.dialog.answer(s)     → looks up + answers in one call
//   world.dialog.clear()       → reset after answering
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

// findOptionCallable backs world.dialog.find_option(text). Returns
// the 0-based index of the FIRST option whose text contains
// (case-insensitive) the given substring, or -1 if no match.
// Substring (not exact) matching is forgiving — quest dialog text
// often has trailing details ("Yes, I'd like to help.") that
// routines shouldn't need to spell out verbatim.
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
		return interp.Int(-1), nil
	}
	lower := strings.ToLower(string(needle))
	for i, opt := range rec.Options {
		if strings.Contains(strings.ToLower(opt), lower) {
			return interp.Int(int64(i)), nil
		}
	}
	return interp.Int(-1), nil
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
	case "my_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted), true
	case "their_first_accepted":
		return interp.Bool(rec != nil && rec.TheirFirstAccepted), true
	case "my_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted), true
	case "their_second_accepted":
		return interp.Bool(rec != nil && rec.TheirSecondAccepted), true
	case "both_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted && rec.TheirFirstAccepted), true
	case "both_second_accepted":
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

// duelView surfaces world.Duel.* to routines. Mirrors tradeView
// with additional fields for the rule toggles:
//   world.duel.is_active             → bool
//   world.duel.phase                 → string (request_sent / open / confirm / completed / cancelled / none)
//   world.duel.with                  → opponent name or Null
//   world.duel.with_index            → opponent server-index Int
//   world.duel.my_offer / their_offer → list of [item_id, amount]
//   world.duel.my_first_accepted / their_first_accepted / etc.
//   world.duel.disallow_retreat / _magic / _prayer / _weapons → bool
type duelView struct{ host *Host }

func (d *duelView) Kind() string    { return "duel" }
func (d *duelView) Display() string { return "<duel>" }

func (d *duelView) Get(field string) (interp.Value, bool) {
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
	case "my_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted), true
	case "their_first_accepted":
		return interp.Bool(rec != nil && rec.TheirFirstAccepted), true
	case "my_second_accepted":
		return interp.Bool(rec != nil && rec.MySecondAccepted), true
	case "their_second_accepted":
		return interp.Bool(rec != nil && rec.TheirSecondAccepted), true
	case "both_first_accepted":
		return interp.Bool(rec != nil && rec.MyFirstAccepted && rec.TheirFirstAccepted), true
	case "both_second_accepted":
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

// prayersView surfaces self.prayers.* to routines:
//   self.prayers.active        → list of int slots that are on
//   self.prayers.is_active(N)  → bool for a specific slot
//   self.prayers.count         → number of currently-active prayers
//
// Slot indices follow the RSC prayer book order:
//   0 Thick Skin / 1 Burst of Strength / 2 Clarity of Thought
//   3 Rock Skin / 4 Superhuman Strength / 5 Improved Reflexes
//   6 Rapid Restore / 7 Rapid Heal / 8 Protect Item
//   9 Steel Skin / 10 Ultimate Strength / 11 Incredible Reflexes
//   12 Protect from Magic / 13 Protect from Missiles
//   14 Protect from Melee (members)
type prayersView struct{ host *Host }

func (p *prayersView) Kind() string    { return "prayers" }
func (p *prayersView) Display() string { return "<prayers>" }

func (p *prayersView) Get(field string) (interp.Value, bool) {
	active := p.host.world.Self.ActivePrayers()
	switch field {
	case "active":
		items := make([]interp.Value, 0)
		for i, on := range active {
			if on {
				items = append(items, interp.Int(int64(i)))
			}
		}
		return &interp.List{Items: items}, true
	case "count":
		n := 0
		for _, on := range active {
			if on {
				n++
			}
		}
		return interp.Int(int64(n)), true
	case "is_active":
		return &prayerIsActiveCallable{host: p.host}, true
	}
	return nil, false
}

type prayerIsActiveCallable struct{ host *Host }

func (c *prayerIsActiveCallable) Kind() string    { return "callable" }
func (c *prayerIsActiveCallable) Display() string { return "<prayers.is_active>" }
func (c *prayerIsActiveCallable) Yields() bool    { return false }

func (c *prayerIsActiveCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("prayers.is_active takes 1 arg (slot index)")
	}
	idx, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("prayers.is_active: slot must be Int")
	}
	return interp.Bool(c.host.world.Self.PrayerActive(int(idx))), nil
}

// bankView surfaces world.Bank.* to routines:
//   world.bank.is_open    → Bool
//   world.bank.max_size   → Int (slot capacity)
//   world.bank.used       → Int (occupied slots)
//   world.bank.free       → Int (max_size - used)
//   world.bank.slots      → list of {item_id, amount}
//   world.bank.has(id)    → Int (total of that item id, 0 if absent)
//   world.bank.count(id)  → alias for .has(id)
type bankView struct{ host *Host }

func (b *bankView) Kind() string    { return "bank" }
func (b *bankView) Display() string { return "<bank>" }

func (b *bankView) Get(field string) (interp.Value, bool) {
	rec := b.host.world.Bank.Bank()
	switch field {
	case "is_open":
		return interp.Bool(rec != nil), true
	case "max_size":
		if rec == nil {
			return interp.Int(0), true
		}
		return interp.Int(int64(rec.MaxSize)), true
	case "used":
		if rec == nil {
			return interp.Int(0), true
		}
		return interp.Int(int64(len(rec.Slots))), true
	case "free":
		if rec == nil {
			return interp.Int(0), true
		}
		return interp.Int(int64(rec.MaxSize - len(rec.Slots))), true
	case "slots":
		if rec == nil {
			return &interp.List{Items: []interp.Value{}}, true
		}
		items := make([]interp.Value, 0, len(rec.Slots))
		for _, s := range rec.Slots {
			pair := &interp.List{Items: []interp.Value{
				interp.Int(int64(s.ItemID)),
				interp.Int(int64(s.Amount)),
			}}
			items = append(items, pair)
		}
		return &interp.List{Items: items}, true
	case "has", "count":
		return &bankCountCallable{host: b.host}, true
	}
	return nil, false
}

// bankCountCallable backs world.bank.has(item_id) / .count(item_id).
type bankCountCallable struct{ host *Host }

func (c *bankCountCallable) Kind() string    { return "callable" }
func (c *bankCountCallable) Display() string { return "<world.bank.has>" }
func (c *bankCountCallable) Yields() bool    { return false }

func (c *bankCountCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("world.bank.has takes 1 argument (item_id)")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("world.bank.has: item_id must be Int")
	}
	return interp.Int(int64(c.host.world.Bank.Has(int(id)))), nil
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
	}
	return nil, false
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

type playerView struct{ record world.PlayerRecord }

func (p *playerView) Kind() string    { return "player" }
func (p *playerView) Display() string { return p.record.Name }
func (p *playerView) Get(field string) (interp.Value, bool) {
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
	case "name":
		if def := n.def(); def != nil {
			return interp.String(def.Name), true
		}
		return interp.Null{}, true

	// Facts-derived combat / interaction fields.
	case "combat_level":
		if def := n.def(); def != nil {
			// Standard RSC combat-level approximation: (atk+str+def)/4 + hits/4
			cl := (def.Attack+def.Strength+def.Defense)/4 + def.Hits/4
			return interp.Int(int64(cl)), true
		}
		return interp.Null{}, true
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

	// Combat-state fields. We don't track per-NPC hp in
	// real-time yet, so hp_fraction is a stub returning null —
	// routines need to derive from observation or wait for the
	// combat-target view (task #65). in_combat_with same.
	case "hp_fraction":
		return interp.Null{}, true
	case "in_combat_with":
		return interp.Null{}, true
	}
	return nil, false
}

// groundItemsView is the entry point for ground-item queries:
//   world.ground_items                   — list of all visible ground items
//   world.ground_items.by_id(N)          — nearest visible ground item with id N, or Null
//   world.ground_items.by_id(N, radius=R) — same, restricted to within R tiles
//   world.ground_items.nearest           — closest ground item (any id), or Null
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
		return g.nearest(g.host.world.GroundItems.All()), true
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

// itemViewVal carries item ID + amount + name. Returned by
// inventory.slots, inventory.find(), and self.wielded. Lookups
// against facts populate the descriptive fields lazily on first
// .Get() — keeps the struct small for hot list traversal.
type itemViewVal struct {
	ID     int
	Amount int
	Name   string
	facts  *facts.Facts
}

func newItemView(f *facts.Facts, id, amount int) *itemViewVal {
	return &itemViewVal{ID: id, Amount: amount, Name: itemName(f, id), facts: f}
}
func (i *itemViewVal) Kind() string    { return "item" }
func (i *itemViewVal) Display() string { return i.Name }
func (i *itemViewVal) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		return interp.Int(int64(i.ID)), true
	case "amount":
		return interp.Int(int64(i.Amount)), true
	case "name":
		return interp.String(i.Name), true
	case "is_stackable":
		// Facts-derived. Without facts loaded, assume non-stackable
		// (safe default — extra deposit calls vs. one bulk deposit
		// is the trade-off).
		if i.facts != nil {
			if def := i.facts.ItemDef(i.ID); def != nil {
				return interp.Bool(def.IsStackable), true
			}
		}
		return interp.Bool(false), true
	case "is_wearable":
		if i.facts != nil {
			if def := i.facts.ItemDef(i.ID); def != nil {
				return interp.Bool(def.IsWearable), true
			}
		}
		return interp.Bool(false), true
	case "is_members_only":
		if i.facts != nil {
			if def := i.facts.ItemDef(i.ID); def != nil {
				return interp.Bool(def.IsMembersOnly), true
			}
		}
		return interp.Bool(false), true
	case "command":
		// The default right-click command (Eat/Drink/Bury/Wield/...).
		if i.facts != nil {
			if def := i.facts.ItemDef(i.ID); def != nil {
				return interp.String(def.Command), true
			}
		}
		return interp.String(""), true
	}
	return nil, false
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
		return l.searchBoundaryByName("ladder"), true
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
	}
	return nil, false
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
//   .within(radius)              — center = self.position
//   .within(radius, position)    — center = supplied position
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
