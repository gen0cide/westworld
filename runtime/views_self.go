package runtime

import (
	"math"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Views for the `self` faculty plus the shared positionView. Each view
// is an interp.Value implementing interp.Getter so member access in
// the DSL translates to method calls on the host's world mirror.
//
// Views never cache — they reflect the latest world-state snapshot at
// the moment of access. The view-root wiring (it.Reserved["self"] =
// &selfView{...}) lives centrally in dsl_bridge.go.

// ---------- self ----------

// selfView exposes `self.*` to the DSL.
type selfView struct{ host *Host }

func (s *selfView) Kind() string    { return "self" }
func (s *selfView) Display() string { return "<self>" }

func (s *selfView) Get(field string) (interp.Value, bool) {
	self := s.host.world.Self
	inv := s.host.world.Inventory
	switch field {
	// Identity
	case "name", "username":
		// The runtime stores the host's username on opts.Username
		// (set from the cradle -username flag). Routines need this
		// to address admin commands like `damage <name> N` at
		// themselves.
		return interp.String(s.host.opts.Username), true

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
	case "spells":
		return &spellsView{host: s.host}, true
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

// positionView is `{x, y}` for self.position. Shared across every
// view that exposes a `.position` field (self, players, npcs, ground
// items, placements, boundaries).
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
