package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// Views for the `prayer` faculty — the self.prayers.* active-state and
// prayer-book catalog accessors.

// prayersView surfaces self.prayers.* to routines:
//   self.prayers.active        → list of int slots that are on
//   self.prayers.is_active(N)  → bool for a specific slot
//   self.prayers.count         → number of currently-active prayers
//
// Slot indices follow the RSC prayer book order:
//   0 Thick Skin / 1 Burst of Strength / 2 Clarity of Thought
//   3 Rock Skin / 4 Superhuman Strength / 5 Improved Reflexes
//   6 Rapid Restore / 7 Rapid Heal / 8 Protect Item
//   9 Steel Skin / 10 Ultimate Strength / 11 Incredible Reflexes
//   12 Protect from Magic / 13 Protect from Missiles
//   14 Protect from Melee (members)
type prayersView struct{ host *Host }

func (p *prayersView) Kind() string    { return "prayers" }
func (p *prayersView) Display() string { return "<prayers>" }

func (p *prayersView) Get(field string) (interp.Value, bool) {
	active := p.host.world.Self.ActivePrayers()
	switch field {
	case "active":
		items := make([]interp.Value, 0)
		for i, on := range active {
			if on {
				items = append(items, interp.Int(int64(i)))
			}
		}
		return &interp.List{Items: items}, true
	case "count":
		n := 0
		for _, on := range active {
			if on {
				n++
			}
		}
		return interp.Int(int64(n)), true
	case "is_active":
		return &prayerIsActiveCallable{host: p.host}, true
	case "book":
		// Static catalog as a list of prayer-defs.
		items := make([]interp.Value, 0, len(facts.Prayers))
		for i := range facts.Prayers {
			items = append(items, &prayerDefView{def: &facts.Prayers[i]})
		}
		return &interp.List{Items: items}, true
	case "by_id":
		return &prayerByIDCallable{}, true
	case "by_name":
		return &prayerByNameCallable{}, true
	}
	return nil, false
}

// prayerDefView surfaces one prayer's static attributes:
//   .id / .name / .req_level / .drain_rate / .description
type prayerDefView struct{ def *facts.PrayerDef }

func (p *prayerDefView) Kind() string    { return "prayer_def" }
func (p *prayerDefView) Display() string { return p.def.Name }

func (p *prayerDefView) Get(field string) (interp.Value, bool) {
	switch field {
	case "id":
		return interp.Int(int64(p.def.ID)), true
	case "name":
		return interp.String(p.def.Name), true
	case "req_level":
		return interp.Int(int64(p.def.ReqLevel)), true
	case "drain_rate":
		return interp.Int(int64(p.def.DrainRate)), true
	case "description":
		return interp.String(p.def.Description), true
	}
	return nil, false
}

type prayerByIDCallable struct{}

func (c *prayerByIDCallable) Kind() string    { return "callable" }
func (c *prayerByIDCallable) Display() string { return "<prayers.by_id>" }
func (c *prayerByIDCallable) Yields() bool    { return false }
func (c *prayerByIDCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("prayers.by_id takes 1 arg (slot index)")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("prayers.by_id: id must be Int")
	}
	d := facts.PrayerByID(int(id))
	if d == nil {
		return interp.Null{}, nil
	}
	return &prayerDefView{def: d}, nil
}

type prayerByNameCallable struct{}

func (c *prayerByNameCallable) Kind() string    { return "callable" }
func (c *prayerByNameCallable) Display() string { return "<prayers.by_name>" }
func (c *prayerByNameCallable) Yields() bool    { return false }
func (c *prayerByNameCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("prayers.by_name takes 1 arg (name string)")
	}
	name, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("prayers.by_name: name must be String")
	}
	d := facts.PrayerByName(string(name))
	if d == nil {
		return interp.Null{}, nil
	}
	return &prayerDefView{def: d}, nil
}

type prayerIsActiveCallable struct{ host *Host }

func (c *prayerIsActiveCallable) Kind() string    { return "callable" }
func (c *prayerIsActiveCallable) Display() string { return "<prayers.is_active>" }
func (c *prayerIsActiveCallable) Yields() bool    { return false }

func (c *prayerIsActiveCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("prayers.is_active takes 1 arg (slot index)")
	}
	idx, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("prayers.is_active: slot must be Int")
	}
	return interp.Bool(c.host.world.Self.PrayerActive(int(idx))), nil
}
