package runtime

import (
	"math"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
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
		// Human/LLM-facing: normalized 0..100% (raw is 0..750). The
		// manual teaches "if self.fatigue > 80" as a percentage.
		return interp.Int(self.FatiguePercent()), true
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
		// Delegate to the combat faculty's engaged check (#combat-prune):
		// true when we have a resolvable combat target (our last-attacked
		// NPC/player still alive + visible, or a wire-observed engagement)
		// or someone is firing at us. Clears the moment the engaged NPC is
		// pruned on death — the same transition combat.target == null
		// watches. Was hard-stubbed false.
		return interp.Bool((&combatView{host: s.host}).engaged()), true
	// is_sleeping: de-stubbed by the fatigue->sleep faculty
	// (wt/b2-fatigue-sleep). Reads the world SleepState mirror set on
	// SEND_SLEEPSCREEN (true) / SEND_STOPSLEEP (false). MERGE NOTE:
	// concurrent equipped/plane edits also touch this file — this is the
	// only line that branch changed in this switch.
	case "is_sleeping":
		return interp.Bool(self.IsSleeping()), true

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
		// equippedView is BOTH the wielded-item list (preserving the
		// frozen `self.equipped` list contract — .length / .first /
		// .all / .filter / [N]) AND the per-slot accessor surface
		// (.weapon / .shield / .head / ... reading Self.EquipBySlot).
		// See the equippedView doc for the two distinct data sources.
		return &equippedView{host: s.host}, true
	}
	return nil, false
}

// itemView builds an item value, looking up the human-readable name
// from facts (if available). Tiny helper used by self.wielded and
// inventory.slots.
func (s *selfView) itemView(itemID, amount int) interp.Value {
	return newItemView(s.host.facts, itemID, amount)
}

// positionView is `{x, y, plane}` for self.position. Shared across
// every view that exposes a `.position` field (self, players, npcs,
// ground items, placements, boundaries).
type positionView struct{ X, Y int }

func (p *positionView) Kind() string    { return "position" }
func (p *positionView) Display() string { return "(" + intDisp(p.X) + "," + intDisp(p.Y) + ")" }
func (p *positionView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(p.X), true
	case "y":
		return interp.Int(p.Y), true
	case "plane":
		// Floor index derived from the wire Y (RSC stacks floors in
		// Y-space at world.PlaneHeight intervals; there is no separate
		// floor field on any packet). Matches world.Self.Plane() for
		// self.position, and is well-defined for every other position
		// view too (player/npc/ground-item/placement/boundary tiles).
		return interp.Int(int64(world.PlaneOf(p.Y))), true
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

// ---------- self.equipped (#117) ----------

// equippedView backs `self.equipped`. It serves TWO distinct surfaces
// over TWO distinct data sources, kept here so the frozen list contract
// and the new per-slot contract co-exist on one value:
//
//  1. List surface (.length / .first / .last / .all / .filter / .map /
//     .find / [N]) — the currently-wielded inventory items (real item
//     ids, from inv.Slots() where Wielded). This preserves the frozen
//     `self.equipped -> List<InvSlot>` behaviour and the routines that
//     filter it themselves.
//  2. Per-slot surface (.weapon / .shield / .head / .hat / .body /
//     .legs / .gloves / .boots / .amulet / .cape / .shirt / .pants) —
//     reads world.Self.EquipBySlot, the per-slot worn-equipment SPRITE
//     ids from the opcode-234 type-5 appearance update.
//
// IMPORTANT DATA-SOURCE CAVEAT for the per-slot surface: the wire
// carries appearance SPRITE ids, not catalogue item ids, and (a) the
// runtime cannot yet identify which player index is ours, so self's
// appearance is not landed into world.Self.EquipBySlot — it reads 0 —
// and (b) facts.ItemDef has no AppearanceId field, so the
// sprite-id -> item-id reverse lookup is impractical. Each per-slot
// accessor therefore returns an equipSlotView exposing .sprite_id (the
// honest datum) plus best-effort .id/.name that stay null until both
// gaps close. See blockers.
type equippedView struct{ host *Host }

func (e *equippedView) Kind() string    { return "equipped" }
func (e *equippedView) Display() string { return "<equipped>" }

// wieldedItems is the list-surface data source: inventory slots flagged
// Wielded, as item-views (real item ids). Unchanged from the prior
// `self.equipped` list behaviour.
func (e *equippedView) wieldedItems() []interp.Value {
	items := make([]interp.Value, 0)
	for _, slot := range e.host.world.Inventory.Slots() {
		if slot.Wielded {
			items = append(items, newItemView(e.host.facts, slot.ItemID, slot.Amount))
		}
	}
	return items
}

// equipSlotGroups maps a per-slot accessor name to the RSC body-animation
// LAYER(s) that hold that gear. Most are one layer, but head/body/legs each
// span TWO (the metal type picks which — e.g. plate-mail legs land in layer
// 2, skirts in layer 7), so the accessor checks both and returns whichever
// is worn. The specific-layer names (large_helmet/platelegs/…) are also
// exposed for callers that need the exact layer. See event.EquipSlot* doc.
var equipSlotGroups = map[string][]int{
	"head":   {event.EquipSlotHat, event.EquipSlotHead},   // medium or large helm
	"hat":    {event.EquipSlotHat, event.EquipSlotHead},   //
	"helmet": {event.EquipSlotHat, event.EquipSlotHead},   //
	"body":   {event.EquipSlotBody, event.EquipSlotShirt}, // chain/leather or platebody
	"torso":  {event.EquipSlotBody, event.EquipSlotShirt}, //
	"legs":   {event.EquipSlotPants, event.EquipSlotLegs}, // platelegs or skirt
	"shield": {event.EquipSlotShield},
	"weapon": {event.EquipSlotWeapon},
	"gloves": {event.EquipSlotGloves},
	"hands":  {event.EquipSlotGloves},
	"boots":  {event.EquipSlotBoots},
	"feet":   {event.EquipSlotBoots},
	"amulet": {event.EquipSlotAmulet},
	"neck":   {event.EquipSlotAmulet},
	"cape":   {event.EquipSlotCape},
	"back":   {event.EquipSlotCape},
	// exact layers
	"large_helmet": {event.EquipSlotHead},
	"med_helmet":   {event.EquipSlotHat},
	"platebody":    {event.EquipSlotShirt},
	"chainbody":    {event.EquipSlotBody},
	"platelegs":    {event.EquipSlotPants},
	"skirt":        {event.EquipSlotLegs},
}

func (e *equippedView) Get(field string) (interp.Value, bool) {
	// Per-slot surface first (the #117 additions).
	if slots, ok := equipSlotGroups[field]; ok {
		return &equipSlotView{host: e.host, slots: slots}, true
	}
	// List surface (preserves the frozen list contract).
	switch field {
	case "all":
		return &interp.List{Items: e.wieldedItems()}, true
	case "length":
		return interp.Int(int64(len(e.wieldedItems()))), true
	case "first":
		items := e.wieldedItems()
		if len(items) == 0 {
			return interp.Null{}, true
		}
		return items[0], true
	case "last":
		items := e.wieldedItems()
		if len(items) == 0 {
			return interp.Null{}, true
		}
		return items[len(items)-1], true
	case "bonuses":
		// The summed combat bonus of all worn gear ("equipment status"):
		// .armour / .aim / .power / .magic / .prayer. Recomputed from the
		// currently-wielded items, so it always reflects what's on now.
		return &equipmentBonusesView{b: e.host.SelfEquipmentBonuses()}, true
	}
	// Note: .filter / .map / .find are list-method callables the
	// interpreter only synthesises for a concrete *List, so route those
	// through `self.equipped.all` (a real list): e.g.
	// `self.equipped.all.filter(i => i.is_wearable)`.
	return nil, false
}

// Index supports `self.equipped[N]` over the wielded-item list.
func (e *equippedView) Index(idx interp.Value) (interp.Value, bool) {
	i, ok := interp.AsInt(idx)
	if !ok {
		return nil, false
	}
	items := e.wieldedItems()
	if i < 0 || int(i) >= len(items) {
		return interp.Null{}, true
	}
	return items[int(i)], true
}

// Iter lets `for slot in self.equipped` walk the wielded-item list,
// matching the groundItemsView precedent.
func (e *equippedView) Iter() []interp.Value { return e.wieldedItems() }

// equipSlotView is one worn-equipment slot accessor returned from
// self.equipped.<slot>. `slots` is the body-animation layer(s) the slot
// can occupy — head/body/legs span two (the metal type picks which), so
// the view resolves whichever layer is actually worn. Items resolve from
// the host's OWN inventory (Wielded flag matched by facts WearSlot —
// exact, and immune to the not-yet-tracked self appearance index).
type equipSlotView struct {
	host  *Host
	slots []int
}

// worn returns the worn item across this accessor's layer group.
func (e *equipSlotView) worn() WornItem { return e.host.selfWornGroup(e.slots) }

func (e *equipSlotView) Kind() string { return "equip_slot" }
func (e *equipSlotView) Display() string {
	if w := e.worn(); !w.Empty() {
		return w.Label()
	}
	return "<empty " + event.EquipSlotName(e.slots[0]) + ">"
}

func (e *equipSlotView) Get(field string) (interp.Value, bool) {
	w := e.worn()
	switch field {
	case "slot":
		return interp.Int(int64(e.slots[0])), true
	case "slot_name":
		return interp.String(event.EquipSlotName(e.slots[0])), true
	case "sprite_id":
		// The honest wire datum: the appearance sprite id in the primary
		// layer (0 = nothing / not yet observed).
		return interp.Int(int64(e.host.world.Self.EquipSpriteAt(e.slots[0]))), true
	case "is_empty":
		return interp.Bool(w.Empty()), true
	case "id":
		if w.Exact() {
			return interp.Int(int64(w.Items[0].ID)), true
		}
		return interp.Null{}, true
	case "name":
		if w.Empty() {
			return interp.Null{}, true
		}
		return interp.String(w.Label()), true
	case "def":
		if w.Exact() {
			return &itemDefView{def: w.Items[0]}, true
		}
		return interp.Null{}, true
	}
	return nil, false
}
