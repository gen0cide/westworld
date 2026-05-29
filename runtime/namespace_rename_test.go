package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// These tests lock in the api.md §6/§10 flat->namespaced rename + the
// §2 Def/Instance value types. They assert runtime DISPATCH, not just
// parsing: that the promoted top-level roots resolve their verbs to
// action callables (and bang variants), that the frozen trade/duel
// view aliases read, that InvSlot carries .idx/.def/.quantity, that
// inventory.find_all returns slots, and that world.messages is a list.

// Each namespaced action verb must resolve to a Callable through its
// view's Get — proving trade.request / bank.deposit / duel.stake /
// magic.cast / prayer.activate / combat.attack dispatch.
func TestNamespacedActionVerbsResolveToCallables(t *testing.T) {
	h := newTestHost()
	// Bind a routine ctx the way NewRoutineInterpreter does.
	h.NewRoutineInterpreter(t.Context())

	// viewVal is implemented by every namespace view (Value + Getter).
	type viewVal interface {
		interp.Value
		Get(field string) (interp.Value, bool)
	}
	cases := []struct {
		view  viewVal
		field string
	}{
		{&tradeView{host: h}, "request"},
		{&tradeView{host: h}, "offer"},
		{&tradeView{host: h}, "accept"},
		{&tradeView{host: h}, "confirm"},
		{&tradeView{host: h}, "decline"},
		{&duelView{host: h}, "request"},
		{&duelView{host: h}, "set_rules"},
		{&duelView{host: h}, "stake"},
		{&duelView{host: h}, "accept"},
		{&duelView{host: h}, "confirm"},
		{&duelView{host: h}, "decline"},
		{&bankView{host: h}, "open"},
		{&bankView{host: h}, "deposit"},
		{&bankView{host: h}, "withdraw"},
		{&bankView{host: h}, "close"},
		{&magicView{host: h}, "cast"},
		{&prayerView{host: h}, "activate"},
		{&prayerView{host: h}, "deactivate"},
		{&combatView{host: h}, "attack"},
		{&combatView{host: h}, "set_style"},
	}
	for _, c := range cases {
		v, ok := c.view.Get(c.field)
		if !ok {
			t.Errorf("%s.%s: not resolved", c.view.Kind(), c.field)
			continue
		}
		if _, ok := v.(interp.Callable); !ok {
			t.Errorf("%s.%s: resolved to %s, want Callable", c.view.Kind(), c.field, v.Kind())
		}
		// Bang variant must also resolve to a (bang) Callable.
		vb, ok := c.view.Get(c.field + "!")
		if !ok {
			t.Errorf("%s.%s!: not resolved", c.view.Kind(), c.field)
			continue
		}
		if _, ok := vb.(interp.Callable); !ok {
			t.Errorf("%s.%s!: resolved to %s, want Callable", c.view.Kind(), c.field, vb.Kind())
		}
	}
}

// The promoted catalog reads must still work through the new roots:
// magic.* exposes the spellbook, prayer.* exposes the prayer state.
func TestPromotedCatalogReadsThroughNewRoots(t *testing.T) {
	h := newTestHost()
	m := &magicView{host: h}
	if v, ok := m.Get("book"); !ok {
		t.Error("magic.book: not resolved")
	} else if _, ok := v.(*interp.List); !ok {
		t.Errorf("magic.book: got %s, want List", v.Kind())
	}
	p := &prayerView{host: h}
	if v, ok := p.Get("active"); !ok {
		t.Error("prayer.active: not resolved")
	} else if _, ok := v.(*interp.List); !ok {
		t.Errorf("prayer.active: got %s, want List", v.Kind())
	}
}

// The frozen trade view aliases (accepted/they_accepted/both_accepted +
// confirmed/...) must read as Bool — they map onto the existing
// first/second-accept fields (§10).
func TestFrozenTradeViewAliases(t *testing.T) {
	h := newTestHost()
	tv := &tradeView{host: h}
	for _, f := range []string{
		"accepted", "they_accepted", "both_accepted",
		"confirmed", "they_confirmed", "both_confirmed",
	} {
		v, ok := tv.Get(f)
		if !ok {
			t.Errorf("trade.%s: not resolved", f)
			continue
		}
		if _, ok := v.(interp.Bool); !ok {
			t.Errorf("trade.%s: got %s, want Bool", f, v.Kind())
		}
	}
}

// InvSlot is the §2 Instance: idx + quantity live, def carries the
// static catalog entry. The flat back-compat fields stay.
func TestInvSlotInstanceShape(t *testing.T) {
	h := newTestHost()
	// inventory.slots[1] is the 5-lobster slot (id 373) at idx 1.
	res := runRoutine(t, h, `routine r() { return inventory.slots[1].idx }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("InvSlot.idx: got %v, want Int(1)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return inventory.slots[1].quantity }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 5 {
		t.Errorf("InvSlot.quantity: got %v, want Int(5)", res.Value)
	}
	// No facts loaded on the test host, so .def is Null (clean
	// degradation) — assert it resolves rather than erroring.
	res = runRoutine(t, h, `routine r() { return inventory.slots[1].def }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("InvSlot.def (no facts): got %v, want Null", res.Value)
	}
}

// inventory.find_all returns a List of every matching slot.
func TestInventoryFindAll(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return inventory.find_all(373).length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("inventory.find_all(373).length: got %v, want Int(1)", res.Value)
	}
}

// world.messages is a List<Message> (§10 rename of last_server_message).
func TestWorldMessagesIsList(t *testing.T) {
	h := newTestHost()
	res := runRoutine(t, h, `routine r() { return world.messages.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 0 {
		t.Errorf("world.messages.length (none yet): got %v, want Int(0)", res.Value)
	}
	h.world.Recent.SetServerMessage("Welcome to RuneScape.")
	res = runRoutine(t, h, `routine r() { return world.messages.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 1 {
		t.Errorf("world.messages.length (one): got %v, want Int(1)", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return world.messages.last.text }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "Welcome to RuneScape." {
		t.Errorf("world.messages.last.text: got %v", res.Value)
	}
}

// End-to-end through the interpreter: a routine that calls a
// namespaced verb behind a guard so no packet is sent (bank not open)
// still parses, validates, and dispatches — the verb is a real
// callable, not a parse-time accident.
func TestNamespacedVerbDispatchEndToEnd(t *testing.T) {
	h := newTestHost()
	// world.bank closed → bank.is_open false; we just confirm the
	// view + verb both resolve and the routine runs to completion.
	res := runRoutine(t, h, `routine r() {
		if bank.is_open {
			bank.close()
		}
		return bank.is_open
	}`)
	if res.Err != nil {
		t.Fatalf("routine errored: %v", res.Err)
	}
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("bank.is_open: got %v, want false", res.Value)
	}
}
