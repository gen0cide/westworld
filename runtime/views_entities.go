package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

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
// The selectors (.nearest / .by_id / .most_valuable) are REACHABLE-ONLY
// by default (views_reachable.go); the callable forms take
// reachable=false to opt out, and the raw list (.all / iteration) is
// never filtered.
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
		// shape — the closest item to self.position, or null when no
		// items qualify. `.nearest(pos)` recenters on an explicit
		// position. When items exist we return a wrapper that both
		// delegates field reads to the by-self nearest item AND is
		// Callable; when none qualify we return a null-LIKE callable
		// (groundItemsNearestNull) so the common `== null` guard still
		// reads true while the CALLED forms stay invocable — the call
		// evaluates this member first, so plain Null here made the
		// documented opt-out error with "Null is not callable".
		//
		// Reachable-only by default (views_reachable.go): an item in
		// negative space is not a loot target. The bare field has no
		// call syntax for the opt-out — use .nearest(reachable=false)
		// or fold over the raw .all list.
		records := reachableGroundItems(g.host.world.GroundItems.All(), g.host.reachGate())
		if len(records) == 0 {
			return &groundItemsNearestNull{host: g.host}, true
		}
		base := g.nearest(records).(*groundItemView)
		return &groundItemsNearestValue{host: g.host, base: base}, true
	case "most_valuable":
		// Value-sorted selector (#117): the visible ground item with
		// the highest ItemDef base value, or Null when none visible /
		// none resolvable. Enables loot-most-valuable without an
		// author-side fold over world.ground_items. Reachable-only by
		// default; the raw .all list is the omniscient escape.
		return g.mostValuable(reachableGroundItems(g.host.world.GroundItems.All(), g.host.reachGate())), true
	case "all":
		return &interp.List{Items: g.Iter()}, true
	case "length":
		return interp.Int(int64(len(g.Iter()))), true
	}
	return nil, false
}

// reachableGroundItems applies a selector reach gate to ground-item
// records. gate == nil keeps everything (no oracle / opt-out).
func reachableGroundItems(records []world.GroundItemRecord, gate func(x, y int) bool) []world.GroundItemRecord {
	if gate == nil {
		return records
	}
	out := make([]world.GroundItemRecord, 0, len(records))
	for _, r := range records {
		if gate(r.X, r.Y) {
			out = append(out, r)
		}
	}
	return out
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

// groundItemRecord makes the nearest-wrapper a valid ground-item argument
// (groundItemish) — see the note on groundItemView.groundItemRecord.
func (n *groundItemsNearestValue) groundItemRecord() world.GroundItemRecord {
	return n.base.record
}

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
// reachable=false opts out of the default reachable-only filter
// (views_reachable.go) — the called form is the bare field's escape
// hatch too: world.ground_items.nearest(reachable=false).
func (n *groundItemsNearestValue) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	return groundItemsNearestCall(n.host, args, named)
}

// groundItemsNearestCall is the shared called-form body behind both
// .nearest values (the item wrapper and the empty-result callable
// null): it re-resolves the live list under the call's own gate, so
// reachable=false works regardless of what the bare field resolved to.
func groundItemsNearestCall(h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	all := reachableGroundItems(h.world.GroundItems.All(), h.selectorGate(named))
	if len(all) == 0 {
		return interp.Null{}, nil
	}
	if len(args) == 0 {
		if _, hasX := named["x"]; !hasX {
			// Self-centered (bare-field equivalent, possibly ungated).
			return (&groundItemsView{host: h}).nearest(all), nil
		}
	}
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("world.ground_items.nearest: %v", err)
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
	return &groundItemView{record: all[bestIdx], facts: h.facts}, nil
}

// groundItemsNearestNull backs `world.ground_items.nearest` when the gated
// selector matches nothing (no items visible, or every visible one is
// unreachable). It reads as null — falsey, == null, absorbing field access —
// so the idiomatic `world.ground_items.nearest == null` guard keeps working,
// while staying Callable so the called forms (.nearest(pos),
// .nearest(reachable=false)) re-resolve against the live list instead of
// erroring "Null is not callable" (the documented opt-out was unreachable
// exactly when it was needed). Deliberately NOT a groundItemish: passing it
// to pick_up fails with the same type error plain Null gets.
type groundItemsNearestNull struct{ host *Host }

func (n *groundItemsNearestNull) Kind() string    { return "null" }
func (n *groundItemsNearestNull) Display() string { return "null" }
func (n *groundItemsNearestNull) NullLike()       {}

// Get absorbs field access like Null (`null.anything == null`).
func (n *groundItemsNearestNull) Get(string) (interp.Value, bool) { return interp.Null{}, true }

func (n *groundItemsNearestNull) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	return groundItemsNearestCall(n.host, args, named)
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
	// Reachable-only by default; reachable=false opts out (views_reachable.go).
	gate := c.host.selectorGate(named)
	all := c.host.world.GroundItems.All()
	var matches []world.GroundItemRecord
	for _, r := range all {
		if r.ItemID != int(id) {
			continue
		}
		if radius >= 0 && chebyshev(pos.X, pos.Y, r.X, r.Y) > radius {
			continue
		}
		if gate != nil && !gate(r.X, r.Y) {
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

// groundItemRecord exposes the underlying world record. Together with
// the identical method on groundItemsNearestValue it forms the small
// unwrap surface (groundItemish, actions_inventory.go) that action
// handlers accept wherever a "ground item" argument is expected. The
// nearest-wrapper REPORTS Kind()=="ground_item" (it delegates to its
// base) but is NOT a *groundItemView — a bare type assertion on the
// concrete type made pick_up(world.ground_items.nearest) fail with
// "expected ground_item, got ground_item" (soak retro 2026-06-10 #3a).
func (g *groundItemView) groundItemRecord() world.GroundItemRecord { return g.record }

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
