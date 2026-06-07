package runtime

import (
	"fmt"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
)

// Equipment perception — the runtime/VM-level surface for answering
// "what is this entity wearing / wielding". Both the DSL views
// (self.equipped.<slot>, world.players[i].equipment.<slot>) and any
// future cognition code go through these Host methods, so the resolution
// logic lives in exactly one place.
//
// Two data sources, picked by who we're asking about:
//   - SELF: the inventory's Wielded items, matched to a slot by the item
//     def's WearSlot. EXACT item ids, and immune to the not-yet-tracked
//     self appearance index (SelfPlayerIndex is hardcoded; see world.go).
//   - OTHER PLAYERS: their appearance packet's per-slot value
//     (PlayerRecord.EquipBySlot[slot] = item AppearanceID & 0xFF),
//     reverse-mapped via facts.ResolveWorn. Same-metal melee weapons are
//     visually identical worn, so a weapon slot can resolve to several
//     candidates — faithful to what a human could actually tell.

// WornItem is the resolved contents of one worn-equipment slot.
type WornItem struct {
	Slot     int              // event.EquipSlot*
	SlotName string           // "helmet", "weapon", "shield", ...
	Items    []*facts.ItemDef // candidates; empty slot => nil
}

// Empty reports nothing worn in the slot.
func (w WornItem) Empty() bool { return len(w.Items) == 0 }

// Exact reports a single, unambiguous item.
func (w WornItem) Exact() bool { return len(w.Items) == 1 }

// Label is a readable name: the item name, or "<name> (or similar)" when
// several visually-identical items match (e.g. bronze melee weapons).
func (w WornItem) Label() string {
	switch len(w.Items) {
	case 0:
		return ""
	case 1:
		return w.Items[0].Name
	default:
		return w.Items[0].Name + " (or similar)"
	}
}

// SelfWornItem resolves what the host itself wears in `slot`, from its
// own inventory (Wielded items matched by facts WearSlot). Exact.
func (h *Host) SelfWornItem(slot int) WornItem {
	w := WornItem{Slot: slot, SlotName: event.EquipSlotName(slot)}
	if h.facts == nil || h.world == nil || h.world.Inventory == nil {
		return w
	}
	for _, s := range h.world.Inventory.Slots() {
		if !s.Wielded {
			continue
		}
		if def := h.facts.ItemDef(s.ItemID); def != nil && def.WearSlot == slot {
			w.Items = append(w.Items, def)
		}
	}
	return w
}

// WornItemFromAppearance resolves one slot of a player's appearance
// (PlayerRecord.EquipBySlot[slot]) back to candidate items. Used for
// other players, whose exact inventory we can't see.
func (h *Host) WornItemFromAppearance(slot, appearanceVal int) WornItem {
	w := WornItem{Slot: slot, SlotName: event.EquipSlotName(slot)}
	if h.facts != nil {
		w.Items = h.facts.ResolveWorn(slot, appearanceVal)
	}
	return w
}

// canonicalEquipSlots is the ordered set of human equip slots used for
// diffing a player's appearance (dual-layer head/body/legs collapsed to
// one). Names match equipSlotGroups.
var canonicalEquipSlots = []struct {
	Name   string
	Layers []int
}{
	{"helmet", []int{event.EquipSlotHat, event.EquipSlotHead}},
	{"body", []int{event.EquipSlotBody, event.EquipSlotShirt}},
	{"legs", []int{event.EquipSlotPants, event.EquipSlotLegs}},
	{"shield", []int{event.EquipSlotShield}},
	{"weapon", []int{event.EquipSlotWeapon}},
	{"gloves", []int{event.EquipSlotGloves}},
	{"boots", []int{event.EquipSlotBoots}},
	{"amulet", []int{event.EquipSlotAmulet}},
	{"cape", []int{event.EquipSlotCape}},
}

// wornDiffers reports whether two resolved worn items represent a real
// change (equipped/removed, or a different item).
func wornDiffers(a, b WornItem) bool {
	if a.Empty() != b.Empty() {
		return true
	}
	if a.Empty() {
		return false
	}
	return a.Items[0].ID != b.Items[0].ID
}

// diffEquip returns the human slot names whose worn item differs between
// two appearance equipment arrays — drives the per-slot
// player_equipment_changed events.
func (h *Host) diffEquip(prev, cur [event.NumEquipSlots]int) []string {
	var changed []string
	for _, g := range canonicalEquipSlots {
		if wornDiffers(h.wornGroupFromAppearance(prev, g.Layers), h.wornGroupFromAppearance(cur, g.Layers)) {
			changed = append(changed, g.Name)
		}
	}
	return changed
}

// selfEquipSnapshot maps each canonical human slot to the item id the
// host has worn there (0 = empty). Used to diff own-equipment changes
// and fire per-slot equipment_changed events. Self resolution is exact
// (inventory Wielded), so each slot is a single id.
func (h *Host) selfEquipSnapshot() map[string]int {
	m := make(map[string]int, len(canonicalEquipSlots))
	for _, g := range canonicalEquipSlots {
		if w := h.selfWornGroup(g.Layers); w.Exact() {
			m[g.Name] = w.Items[0].ID
		} else {
			m[g.Name] = 0
		}
	}
	return m
}

// selfWornGroup returns the first non-empty worn item across the given
// body-animation layers (head/body/legs span two — see equipSlotGroups).
func (h *Host) selfWornGroup(slots []int) WornItem {
	for _, s := range slots {
		if w := h.SelfWornItem(s); !w.Empty() {
			return w
		}
	}
	return WornItem{Slot: slots[0], SlotName: event.EquipSlotName(slots[0])}
}

// wornGroupFromAppearance is selfWornGroup's other-player counterpart,
// resolving across layers from a player's appearance equipment array.
func (h *Host) wornGroupFromAppearance(eq [event.NumEquipSlots]int, slots []int) WornItem {
	for _, s := range slots {
		if w := h.WornItemFromAppearance(s, eq[s]); !w.Empty() {
			return w
		}
	}
	return WornItem{Slot: slots[0], SlotName: event.EquipSlotName(slots[0])}
}

// SelfEquipment returns every non-empty worn slot for the host itself.
func (h *Host) SelfEquipment() []WornItem {
	out := make([]WornItem, 0, event.NumEquipSlots)
	for slot := 0; slot < event.NumEquipSlots; slot++ {
		if w := h.SelfWornItem(slot); !w.Empty() {
			out = append(out, w)
		}
	}
	return out
}

// PlayerEquipment returns every non-empty worn slot for a player given
// their appearance equipment array (PlayerRecord.EquipBySlot).
func (h *Host) PlayerEquipment(eq [event.NumEquipSlots]int) []WornItem {
	out := make([]WornItem, 0, event.NumEquipSlots)
	for slot := 0; slot < event.NumEquipSlots; slot++ {
		if w := h.WornItemFromAppearance(slot, eq[slot]); !w.Empty() {
			out = append(out, w)
		}
	}
	return out
}

// ---- DSL views (thin wrappers over the runtime API above) ----

// wornItemView exposes one WornItem to the DSL. Backs the per-slot
// accessors on world.players[i].equipment and (for self) the resolved
// fields of self.equipped.<slot>.
type wornItemView struct{ w WornItem }

func (v *wornItemView) Kind() string { return "worn_item" }
func (v *wornItemView) Display() string {
	if v.w.Empty() {
		return "<empty " + v.w.SlotName + ">"
	}
	return v.w.Label()
}

func (v *wornItemView) Get(field string) (interp.Value, bool) {
	switch field {
	case "slot":
		return interp.Int(int64(v.w.Slot)), true
	case "slot_name":
		return interp.String(v.w.SlotName), true
	case "is_empty":
		return interp.Bool(v.w.Empty()), true
	case "ambiguous":
		return interp.Bool(len(v.w.Items) > 1), true
	case "name":
		if v.w.Empty() {
			return interp.Null{}, true
		}
		return interp.String(v.w.Label()), true
	case "id":
		if v.w.Exact() {
			return interp.Int(int64(v.w.Items[0].ID)), true
		}
		return interp.Null{}, true
	case "def":
		if v.w.Exact() {
			return &itemDefView{def: v.w.Items[0]}, true
		}
		return interp.Null{}, true
	case "candidates":
		items := make([]interp.Value, 0, len(v.w.Items))
		for _, d := range v.w.Items {
			items = append(items, interp.String(d.Name))
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// equipmentView backs world.players[i].equipment. Per-slot accessors
// (.helmet / .weapon / .shield / .body / .legs / .gloves / .boots /
// .amulet / .cape / .head) plus .all and .length.
type equipmentView struct {
	host *Host
	eq   [event.NumEquipSlots]int
}

func (e *equipmentView) Kind() string    { return "equipment" }
func (e *equipmentView) Display() string { return "<equipment>" }

func (e *equipmentView) Get(field string) (interp.Value, bool) {
	if slots, ok := equipSlotGroups[field]; ok {
		return &wornItemView{w: e.host.wornGroupFromAppearance(e.eq, slots)}, true
	}
	switch field {
	case "all":
		ws := e.host.PlayerEquipment(e.eq)
		items := make([]interp.Value, 0, len(ws))
		for _, w := range ws {
			items = append(items, &wornItemView{w: w})
		}
		return &interp.List{Items: items}, true
	case "length":
		return interp.Int(int64(len(e.host.PlayerEquipment(e.eq)))), true
	}
	return nil, false
}

// ---- equipment bonuses ("equipment status") ----

// EquipmentBonuses is the summed combat bonus the worn gear provides —
// the same five totals the in-game equipment screen shows. Recomputed on
// demand from the currently-wielded items, so it always reflects what's
// equipped right now (the world mirror updates Wielded on every
// equip/unequip, so there is nothing to persist).
type EquipmentBonuses struct {
	Armour      int
	WeaponAim   int
	WeaponPower int
	Magic       int
	Prayer      int
}

// SelfEquipmentBonuses totals the host's worn-item bonuses.
func (h *Host) SelfEquipmentBonuses() EquipmentBonuses {
	var b EquipmentBonuses
	if h.facts == nil || h.world == nil || h.world.Inventory == nil {
		return b
	}
	for _, s := range h.world.Inventory.Slots() {
		if !s.Wielded {
			continue
		}
		if def := h.facts.ItemDef(s.ItemID); def != nil {
			b.Armour += def.ArmourBonus
			b.WeaponAim += def.WeaponAimBonus
			b.WeaponPower += def.WeaponPowerBonus
			b.Magic += def.MagicBonus
			b.Prayer += def.PrayerBonus
		}
	}
	return b
}

// equipmentBonusesView backs self.equipped.bonuses.
type equipmentBonusesView struct{ b EquipmentBonuses }

func (v *equipmentBonusesView) Kind() string { return "equipment_bonuses" }
func (v *equipmentBonusesView) Display() string {
	return fmt.Sprintf("armour=%d aim=%d power=%d magic=%d prayer=%d",
		v.b.Armour, v.b.WeaponAim, v.b.WeaponPower, v.b.Magic, v.b.Prayer)
}
func (v *equipmentBonusesView) Get(field string) (interp.Value, bool) {
	switch field {
	case "armour":
		return interp.Int(int64(v.b.Armour)), true
	case "weapon_aim", "aim":
		return interp.Int(int64(v.b.WeaponAim)), true
	case "weapon_power", "power":
		return interp.Int(int64(v.b.WeaponPower)), true
	case "magic":
		return interp.Int(int64(v.b.Magic)), true
	case "prayer":
		return interp.Int(int64(v.b.Prayer)), true
	}
	return nil, false
}

// ---- relative threat ("how dangerous is X to me") ----

// threatBand maps a combat-level gap to the authentic RSC danger label +
// colour the client paints over an entity's level (server-side
// Formulae.getLvlDiffColour). `d = myLevel - theirLevel`: very negative
// (they tower over me) is deadly/red, very positive (I tower over them)
// is trivial/green, zero is an even/white match. This surfaces the
// "darker red = more dangerous" cue a host can't read off a UI it never sees.
func threatBand(myLevel, theirLevel int) (label, colour string) {
	switch d := myLevel - theirLevel; {
	case d < -9:
		return "deadly", "@red@"
	case d < -6:
		return "very dangerous", "@or3@"
	case d < -3:
		return "dangerous", "@or2@"
	case d < 0:
		return "risky", "@or1@"
	case d > 9:
		return "trivial", "@gre@"
	case d > 6:
		return "very easy", "@gr3@"
	case d > 3:
		return "easy", "@gr2@"
	case d > 0:
		return "favourable", "@gr1@"
	default:
		return "even", "@whi@"
	}
}
