package runtime

import (
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
	case "position":
		p := self.Position()
		return &positionView{X: p.X, Y: p.Y}, true
	case "hp":
		return interp.Int(self.HP()), true
	case "max_hp":
		return interp.Int(self.MaxHP()), true
	case "prayer":
		return interp.Int(self.Prayer()), true
	case "max_prayer":
		return interp.Int(self.MaxPrayer()), true
	case "fatigue":
		return interp.Int(self.Fatigue()), true
	case "combat_level":
		return interp.Int(self.CombatLevel()), true
	case "skills":
		return &skillsView{host: s.host}, true
	case "wielded":
		for _, slot := range inv.Slots() {
			if slot.Wielded {
				return s.itemView(slot.ItemID, slot.Amount), true
			}
		}
		return interp.Null{}, true
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

// skillsView resolves `self.skills.cooking` etc. via the standard
// 18-skill ordering used by RSC (0=Attack, 3=Hits, 5=Prayer, etc).
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
	return interp.Int(s.host.world.Self.SkillLevel(id)), true
}

// ---------- inventory ----------

type inventoryView struct{ host *Host }

func (v *inventoryView) Kind() string    { return "inventory" }
func (v *inventoryView) Display() string { return "<inventory>" }

func (v *inventoryView) Get(field string) (interp.Value, bool) {
	inv := v.host.world.Inventory
	switch field {
	case "free":
		return interp.Int(inv.FreeSlots()), true
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
	}
	return nil, false
}

// ---------- world ----------

type worldView struct{ host *Host }

func (w *worldView) Kind() string    { return "world" }
func (w *worldView) Display() string { return "<world>" }

func (w *worldView) Get(field string) (interp.Value, bool) {
	switch field {
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
		records := w.host.world.GroundItems.All()
		items := make([]interp.Value, 0, len(records))
		for _, g := range records {
			items = append(items, &groundItemView{record: g, facts: w.host.facts})
		}
		return &interp.List{Items: items}, true
	case "locs":
		return &locsView{host: w.host}, true
	}
	return nil, false
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
		if n.facts != nil {
			if def := n.facts.NpcDef(n.record.TypeID); def != nil {
				return interp.String(def.Name), true
			}
		}
		return interp.Null{}, true
	}
	return nil, false
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
	}
	return nil, false
}

// itemViewVal carries item ID + amount + name. Returned by
// inventory.slots and self.wielded.
type itemViewVal struct {
	ID     int
	Amount int
	Name   string
}

func newItemView(f *facts.Facts, id, amount int) *itemViewVal {
	return &itemViewVal{ID: id, Amount: amount, Name: itemName(f, id)}
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
	}
	return nil, false
}

// searchByName collects scenery/NPC placements whose name contains
// the given substring. Case-insensitive.
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
	for i := range l.host.facts.NpcDefs {
		def := l.host.facts.NpcDefs[i]
		if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
			out.kinds = append(out.kinds, "npc_spawn")
			out.names = append(out.names, def.Name)
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
