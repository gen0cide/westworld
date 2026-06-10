package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// Views for the `magic` faculty — the top-level magic.* root (action
// verb magic.cast + the promoted spell catalog, §10: self.spells.* ->
// magic.*) and its spell-def accessors. The spellsView under
// self.spells is kept for back-compat and reused by magicView for the
// catalog reads.

// magicView is the top-level `magic` namespace root. It dispatches the
// magic.cast action verb, then exposes the spell catalog (promoted
// from self.spells per §10). Reached via it.Reserved["magic"].
type magicView struct {
	host *Host
	bind *routineBinding
}

func (m *magicView) Kind() string    { return "magic" }
func (m *magicView) Display() string { return "<magic>" }

func (m *magicView) Get(field string) (interp.Value, bool) {
	// Action verb (cast + bang) first.
	if v, ok := m.host.namespaceAction(m.bind, "magic", field, magicVerbs); ok {
		return v, true
	}
	// ----- magic reads (#117) -----
	// Top-level magic.* reads computed CLIENT-SIDE (no packet): they
	// gate on the *current/boosted* Magic level (skill index 6,
	// SkillLevel = effective level) per api.md §621/§625, in contrast
	// to the promoted self.spells.* catalog under spellsView which
	// keys off the base (max) level. These are added at the magic
	// namespace root WITHOUT touching spellsView, so magic.spells.* /
	// self.spells.* retain their existing behavior.
	switch field {
	case "level":
		// Current (boostable) Magic level. §621.
		return interp.Int(int64(m.host.magicLevel())), true
	case "max_level":
		// Base (unboostable) Magic level. §623.
		return interp.Int(int64(m.host.world.Self.SkillMax(magicSkillID))), true
	case "known":
		// Spells whose req_level <= current/boosted magic level. §629
		// at the root keys off the effective level (vs spellsView.known
		// which uses the base level).
		myMagic := m.host.magicLevel()
		items := make([]interp.Value, 0)
		for i := range facts.Spells {
			if facts.Spells[i].ReqLevel <= myMagic {
				items = append(items, &spellDefView{def: &facts.Spells[i]})
			}
		}
		return &interp.List{Items: items}, true
	case "can_cast":
		return &spellCanCastCallable{host: m.host}, true
	case "has_runes_for":
		// Root has_runes_for honors equipped elemental staves (an
		// air-staff satisfies the air rune, etc.); the promoted
		// spellsView.has_runes_for is a plain inventory-only check.
		return &spellHasRunesStaffCallable{host: m.host}, true
	}
	// Spell catalog (book/by_id/by_name/count + the inventory-only
	// known/has_runes_for), promoted from self.spells.
	return (&spellsView{host: m.host}).Get(field)
}

// magicSkillID is the Magic skill's stat index (RSC skill ordering).
const magicSkillID = 6

// magicLevel returns the host's current/boosted Magic level (skill
// index 6, effective level) — the figure the client uses to decide
// which spells are castable.
func (h *Host) magicLevel() int { return h.world.Self.SkillLevel(magicSkillID) }

// elementalStaffItems maps an elemental rune item id to the staff item
// ids that supply that element for free (so the rune need not be held).
// Fire(31) → 197/615/682, Water(32) → 102/616/683, Air(33) →
// 101/617/684, Earth(34) → 103/618/685 (#117).
var elementalStaffItems = map[int][]int{
	31: {197, 615, 682}, // fire
	32: {102, 616, 683}, // water
	33: {101, 617, 684}, // air
	34: {103, 618, 685}, // earth
}

// spellsView surfaces self.spells.* — the magic spellbook catalog.
//
//	self.spells.book               → list of spell-defs
//	self.spells.known              → list of spell-defs we have the magic level for
//	self.spells.by_id(N)           → spell-def or null
//	self.spells.by_name("Heal")    → spell-def or null
//	self.spells.has_runes_for(N)   → bool (only meaningful if .by_id(N) != null)
//	self.spells.count              → total catalog size
type spellsView struct{ host *Host }

func (s *spellsView) Kind() string    { return "spells" }
func (s *spellsView) Display() string { return "<spells>" }

func (s *spellsView) Get(field string) (interp.Value, bool) {
	switch field {
	case "count":
		return interp.Int(int64(len(facts.Spells))), true
	case "book":
		items := make([]interp.Value, 0, len(facts.Spells))
		for i := range facts.Spells {
			items = append(items, &spellDefView{def: &facts.Spells[i]})
		}
		return &interp.List{Items: items}, true
	case "known":
		// Spells whose req_level we've reached. Magic skill = index 6.
		myMagic := s.host.world.Self.SkillMax(6)
		items := make([]interp.Value, 0)
		for i := range facts.Spells {
			if facts.Spells[i].ReqLevel <= myMagic {
				items = append(items, &spellDefView{def: &facts.Spells[i]})
			}
		}
		return &interp.List{Items: items}, true
	case "by_id":
		return &spellByIDCallable{}, true
	case "by_name":
		return &spellByNameCallable{}, true
	case "has_runes_for":
		return &spellHasRunesCallable{host: s.host}, true
	}
	return nil, false
}

type spellDefView struct{ def *facts.SpellDef }

func (s *spellDefView) Kind() string    { return "spell_def" }
func (s *spellDefView) Display() string { return s.def.Name }

func (s *spellDefView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		return interp.Int(int64(s.def.ID)), true
	case "name":
		return interp.String(s.def.Name), true
	case "req_level":
		return interp.Int(int64(s.def.ReqLevel)), true
	case "type":
		return interp.Int(int64(s.def.Type)), true
	case "exp":
		return interp.Int(int64(s.def.ExpReward)), true
	case "description":
		return interp.String(s.def.Description), true
	case "members":
		return interp.Bool(s.def.Members), true
	case "evil":
		return interp.Bool(s.def.Evil), true
	case "runes":
		// list of [item_id, count] pairs
		items := make([]interp.Value, 0, len(s.def.Runes))
		for _, r := range s.def.Runes {
			items = append(items, &interp.List{Items: []interp.Value{
				interp.Int(int64(r.ItemID)),
				interp.Int(int64(r.Count)),
			}})
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

type spellByIDCallable struct{}

func (c *spellByIDCallable) Kind() string    { return "callable" }
func (c *spellByIDCallable) Display() string { return "<spells.by_id>" }
func (c *spellByIDCallable) Yields() bool    { return false }
func (c *spellByIDCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("spells.by_id takes 1 arg (spell id)")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("spells.by_id: id must be Int")
	}
	d := facts.SpellByID(int(id))
	if d == nil {
		return interp.Null{}, nil
	}
	return &spellDefView{def: d}, nil
}

type spellByNameCallable struct{}

func (c *spellByNameCallable) Kind() string    { return "callable" }
func (c *spellByNameCallable) Display() string { return "<spells.by_name>" }
func (c *spellByNameCallable) Yields() bool    { return false }
func (c *spellByNameCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("spells.by_name takes 1 arg (name)")
	}
	name, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("spells.by_name: name must be String")
	}
	d := facts.SpellByName(string(name))
	if d == nil {
		return interp.Null{}, nil
	}
	return &spellDefView{def: d}, nil
}

// spellHasRunesCallable checks whether the host has enough runes in
// inventory to cast a given spell. Takes spell id or name; returns
// false if the spell is unknown.
type spellHasRunesCallable struct{ host *Host }

func (c *spellHasRunesCallable) Kind() string    { return "callable" }
func (c *spellHasRunesCallable) Display() string { return "<spells.has_runes_for>" }
func (c *spellHasRunesCallable) Yields() bool    { return false }
func (c *spellHasRunesCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("spells.has_runes_for takes 1 arg (spell id or name)")
	}
	var def *facts.SpellDef
	if id, ok := interp.AsInt(args[0]); ok {
		def = facts.SpellByID(int(id))
	} else if name, ok := args[0].(interp.String); ok {
		def = facts.SpellByName(string(name))
	} else {
		return nil, errf("spells.has_runes_for: arg must be Int id or String name")
	}
	if def == nil {
		return interp.Bool(false), nil
	}
	// Sum inventory holdings per item id, then verify each rune.
	counts := map[int]int{}
	for _, sl := range c.host.world.Inventory.Slots() {
		counts[sl.ItemID] += sl.Amount
	}
	for _, r := range def.Runes {
		if counts[r.ItemID] < r.Count {
			return interp.Bool(false), nil
		}
	}
	return interp.Bool(true), nil
}

// ----- magic reads (#117) -----

// resolveSpellRef resolves a SpellRef arg — Int(spellbook id) |
// String(name, case-insensitive) | SpellDef view — to its SpellDef,
// returning nil when the spell is unknown. Shared by the magic.*
// can_cast / has_runes_for reads so the SpellRef contract (api.md
// §625/§632) is identical across them.
func resolveSpellRef(v interp.Value) *facts.SpellDef {
	switch a := v.(type) {
	case *spellDefView:
		return a.def
	case interp.String:
		return facts.SpellByName(string(a))
	default:
		if id, ok := interp.AsInt(v); ok {
			return facts.SpellByID(int(id))
		}
	}
	return nil
}

// spellCanCastCallable backs magic.can_cast(spell) — a pure, client-
// side LEVEL-ONLY gate: true iff the spell's req_level is within the
// host's current/boosted Magic level (api.md §625). It is null-safe:
// an unknown spell returns Bool(false), never Null. Rune/gear checks
// are deliberately NOT part of this (use magic.has_runes_for).
type spellCanCastCallable struct{ host *Host }

func (c *spellCanCastCallable) Kind() string    { return "callable" }
func (c *spellCanCastCallable) Display() string { return "<magic.can_cast>" }
func (c *spellCanCastCallable) Yields() bool    { return false }
func (c *spellCanCastCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("magic.can_cast takes 1 arg (spell id, name, or SpellDef)")
	}
	def := resolveSpellRef(args[0])
	if def == nil {
		return interp.Bool(false), nil
	}
	return interp.Bool(def.ReqLevel <= c.host.magicLevel()), nil
}

// spellHasRunesStaffCallable backs magic.has_runes_for(spell): for each
// required rune, inventory count >= count — but an *equipped* elemental
// staff (a wielded inventory slot holding one of elementalStaffItems)
// satisfies that element's rune outright. Null-safe: an unknown spell
// returns Bool(false), never Null (#117).
type spellHasRunesStaffCallable struct{ host *Host }

func (c *spellHasRunesStaffCallable) Kind() string    { return "callable" }
func (c *spellHasRunesStaffCallable) Display() string { return "<magic.has_runes_for>" }
func (c *spellHasRunesStaffCallable) Yields() bool    { return false }
func (c *spellHasRunesStaffCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("magic.has_runes_for takes 1 arg (spell id, name, or SpellDef)")
	}
	def := resolveSpellRef(args[0])
	if def == nil {
		return interp.Bool(false), nil
	}
	// Sum inventory holdings per item id; separately note which
	// elemental rune ids are covered by an equipped (wielded) staff.
	counts := map[int]int{}
	staffCovered := map[int]bool{}
	for _, sl := range c.host.world.Inventory.Slots() {
		counts[sl.ItemID] += sl.Amount
		if sl.Wielded {
			for rune, staves := range elementalStaffItems {
				for _, st := range staves {
					if sl.ItemID == st {
						staffCovered[rune] = true
					}
				}
			}
		}
	}
	for _, r := range def.Runes {
		if staffCovered[r.ItemID] {
			continue // equipped staff supplies this element
		}
		if counts[r.ItemID] < r.Count {
			return interp.Bool(false), nil
		}
	}
	return interp.Bool(true), nil
}
