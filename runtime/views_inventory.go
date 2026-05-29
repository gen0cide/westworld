package runtime

import (
	"strings"

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
		for idx, s := range inv.Slots() {
			items = append(items, newItemViewAt(v.host.facts, idx, s.ItemID, s.Amount))
		}
		return &interp.List{Items: items}, true
	case "has":
		return invHasCallable{host: v.host}, true
	case "count":
		return invCountCallable{host: v.host}, true
	case "find":
		return invFindCallable{host: v.host}, true
	case "find_all":
		return invFindAllCallable{host: v.host}, true
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
	for idx, s := range c.host.world.Inventory.Slots() {
		if s.ItemID == id {
			return newItemViewAt(c.host.facts, idx, s.ItemID, s.Amount), nil
		}
	}
	return interp.Null{}, nil
}

// invFindAllCallable implements `inventory.find_all(item)` — returns
// every matching slot as an InvSlot instance (one per slot; non-
// stackables occupy several). Empty list if no match. ItemRef
// resolution matches has/find/count.
type invFindAllCallable struct{ host *Host }

func (c invFindAllCallable) Kind() string    { return "builtin" }
func (c invFindAllCallable) Display() string { return "<inventory.find_all>" }
func (c invFindAllCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("inventory.find_all takes 1 argument, got %d", len(args))
	}
	id, err := resolveItemID(c.host.facts, args[0])
	if err != nil {
		return nil, err
	}
	items := make([]interp.Value, 0)
	for idx, s := range c.host.world.Inventory.Slots() {
		if s.ItemID == id {
			items = append(items, newItemViewAt(c.host.facts, idx, s.ItemID, s.Amount))
		}
	}
	return &interp.List{Items: items}, nil
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

// itemViewVal is an InvSlot INSTANCE (api.md §2): a live occurrence of
// an item in an inventory slot. Returned by inventory.slots,
// inventory.find(), inventory.find_all(), and self.wielded.
//
// Frozen instance shape: InvSlot{idx:Int, def:ItemDef, quantity:Int}.
// Static attributes (name/stackable/wearable/members/…) live under
// `.def`; live state (`.idx`, `.quantity`) lives directly. The flat
// `.id`/`.amount`/`.name`/`.is_*`/`.command` fields are retained as a
// back-compat convenience (they proxy the def), so existing routines
// keep working while authoring migrates to the .def form.
//
// Lookups against facts populate the descriptive fields lazily on
// first .Get() — keeps the struct small for hot list traversal. Idx
// is -1 when the slot index is unknown (e.g. self.wielded, where the
// equipment-by-slot packet isn't decoded yet).
type itemViewVal struct {
	ID     int
	Amount int
	Idx    int
	Name   string
	facts  *facts.Facts
}

func newItemView(f *facts.Facts, id, amount int) *itemViewVal {
	return &itemViewVal{ID: id, Amount: amount, Idx: -1, Name: itemName(f, id), facts: f}
}

// newItemViewAt is newItemView with a known slot index — used by
// inventory.slots / find / find_all where the slot position is known.
func newItemViewAt(f *facts.Facts, idx, id, amount int) *itemViewVal {
	return &itemViewVal{ID: id, Amount: amount, Idx: idx, Name: itemName(f, id), facts: f}
}

func (i *itemViewVal) Kind() string    { return "item" }
func (i *itemViewVal) Display() string { return i.Name }
func (i *itemViewVal) Get(field string) (interp.Value, bool) {
	switch field {
	// ---- Instance fields (frozen InvSlot shape) ----
	case "idx":
		if i.Idx < 0 {
			return interp.Null{}, true
		}
		return interp.Int(int64(i.Idx)), true
	case "quantity":
		return interp.Int(int64(i.Amount)), true
	case "def":
		// The static catalog entry (ItemDef). Carries name/stackable/
		// wearable/members/edible/command. Null if facts not loaded.
		if i.facts != nil {
			if def := i.facts.ItemDef(i.ID); def != nil {
				return &itemDefView{def: def}, true
			}
		}
		return interp.Null{}, true
	// ---- Back-compat flat fields ----
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

// itemDefView is an ItemDef DEFINITION (api.md §2): the static catalog
// entry for an item id, immutable, from the facts registry. Reached
// via InvSlot.def / GroundItem.def. Carries only static attributes —
// no live state.
type itemDefView struct{ def *facts.ItemDef }

func (d *itemDefView) Kind() string    { return "item_def" }
func (d *itemDefView) Display() string { return d.def.Name }

func (d *itemDefView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		return interp.Int(int64(d.def.ID)), true
	case "name":
		return interp.String(d.def.Name), true
	case "description":
		return interp.String(d.def.Description), true
	case "command":
		return interp.String(d.def.Command), true
	case "stackable", "is_stackable":
		return interp.Bool(d.def.IsStackable), true
	case "wearable", "is_wearable":
		return interp.Bool(d.def.IsWearable), true
	case "members", "is_members_only":
		return interp.Bool(d.def.IsMembersOnly), true
	case "tradable", "is_tradable":
		return interp.Bool(!d.def.IsUntradable), true
	case "edible", "is_edible":
		// RSC edibility isn't a dedicated def flag; the default
		// command ("Eat"/"Drink") is the closest GUI-visible signal.
		cmd := strings.ToLower(d.def.Command)
		return interp.Bool(cmd == "eat" || cmd == "drink"), true
	}
	return nil, false
}
