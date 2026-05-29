package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// Views for the `inventory` faculty plus the itemViewVal value type
// that inventory.slots / inventory.find / self.wielded return. The
// view-root wiring (it.Reserved["inventory"] = &inventoryView{...})
// lives centrally in dsl_bridge.go.

// ---------- inventory ----------

type inventoryView struct{ host *Host }

func (v *inventoryView) Kind() string    { return "inventory" }
func (v *inventoryView) Display() string { return "<inventory>" }

func (v *inventoryView) Get(field string) (interp.Value, bool) {
	inv := v.host.world.Inventory
	switch field {
	case "free":
		return interp.Int(inv.FreeSlots()), true
	case "used":
		// 30-slot inventory is the RSC convention. used = capacity -
		// free; if we ever support variable-capacity inventories
		// (members vs free, etc.), this becomes a real accessor.
		return interp.Int(30 - inv.FreeSlots()), true
	case "capacity":
		return interp.Int(30), true
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
