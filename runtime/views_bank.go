package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
)

// View for the `bank` faculty — world.bank.* state. Reached via
// worldView.Get("bank") (see views_world.go).

// bankView surfaces world.Bank.* to routines:
//
//	world.bank.is_open    → Bool
//	world.bank.max_size   → Int (slot capacity)
//	world.bank.used       → Int (occupied slots)
//	world.bank.free       → Int (max_size - used)
//	world.bank.slots      → list of {item_id, amount}
//	world.bank.has(id)    → Int (total of that item id, 0 if absent)
//	world.bank.count(id)  → alias for .has(id)
type bankView struct{ host *Host }

func (b *bankView) Kind() string    { return "bank" }
func (b *bankView) Display() string { return "<bank>" }

func (b *bankView) Get(field string) (interp.Value, bool) {
	// Action verbs (open/deposit/withdraw/close + bang) first.
	if v, ok := b.host.namespaceAction("bank", field, bankVerbs); ok {
		return v, true
	}
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
