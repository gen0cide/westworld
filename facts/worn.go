package facts

import (
	"sort"
	"strings"
)

// Worn-equipment perception. The opcode-234 player-appearance packet
// carries, per worn slot, the LOW BYTE of the equipped item's
// AppearanceID (itemDef.getAppearanceId() & 0xFF) — never an item id.
// To let a host perceive *what another player is wearing*, we build a
// reverse index once at load: [wearSlot][appearanceID & 0xFF] → the
// item definitions that render that way.
//
// Most slots resolve to a single item. Same-metal melee weapons share a
// worn appearance (e.g. every bronze sword/dagger), so a weapon slot can
// resolve to several candidates — which is faithful: a human can't tell a
// bronze short sword from a bronze long sword slung on someone's back
// either. Helmets and armour are unique.

// buildWornIndex populates wornBySlot and itemByName from the loaded item
// defs. Idempotent.
func (f *Facts) buildWornIndex() {
	f.wornBySlot = map[int]map[int][]*ItemDef{}
	f.itemByName = make(map[string]*ItemDef, len(f.ItemDefs))
	for _, d := range f.ItemDefs {
		if d == nil {
			continue
		}
		// Reverse name index (lowest id wins on duplicate names).
		if name := strings.ToLower(strings.TrimSpace(d.Name)); name != "" {
			if prev, ok := f.itemByName[name]; !ok || d.ID < prev.ID {
				f.itemByName[name] = d
			}
		}
		if !d.IsWearable {
			continue
		}
		key := d.AppearanceID & 0xFF
		bySlot := f.wornBySlot[d.WearSlot]
		if bySlot == nil {
			bySlot = map[int][]*ItemDef{}
			f.wornBySlot[d.WearSlot] = bySlot
		}
		bySlot[key] = append(bySlot[key], d)
	}
	// Deterministic candidate order (lowest id first).
	for _, bySlot := range f.wornBySlot {
		for _, cands := range bySlot {
			sort.Slice(cands, func(i, j int) bool { return cands[i].ID < cands[j].ID })
		}
	}
}

// ItemDefByName is the reverse of ItemDef(id): a case-insensitive exact
// name lookup (lowest id on duplicate names). Returns nil if unknown.
// Lets cognition/brain code resolve a name back to a def/id so a host
// can reason in item names, not raw ids.
func (f *Facts) ItemDefByName(name string) *ItemDef {
	if f == nil || f.itemByName == nil {
		return nil
	}
	return f.itemByName[strings.ToLower(strings.TrimSpace(name))]
}

// ResolveWorn returns the candidate items matching a worn-equipment
// appearance value in the given wear slot. `appearanceLowByte` is the per
// slot value the appearance packet carries (PlayerRecord.EquipBySlot[slot]).
// Returns nil for an empty slot (value 0), an unknown slot, or no match.
// Usually one candidate; same-metal melee weapons return several.
func (f *Facts) ResolveWorn(slot, appearanceLowByte int) []*ItemDef {
	if f == nil || appearanceLowByte == 0 {
		return nil
	}
	if bySlot, ok := f.wornBySlot[slot]; ok {
		return bySlot[appearanceLowByte&0xFF]
	}
	return nil
}
