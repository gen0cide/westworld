package remoteclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/runtime"
)

// Fixed def IDs the dispatch tests reference; fakeFacts() returns a *facts.Facts
// whose def maps hold exactly these, so a Dispatcher built with it re-derives the
// SAME option list the test's BuildMenu lookup uses (the /pick<->/act agreement).
const (
	monsterNpc             = 10 // Command1 "Attack", Attackable
	talkNpc                = 11 // Command1 "Talk-to"
	thievableNpc           = 12 // Command1 "Pickpocket", Attackable
	command2Npc            = 13 // Command1 "Talk-to", Command2 "Pickpocket" (command2 verb-route)
	sceneryWithCommand     = 20 // Command1 "Mine"
	sceneryWithTwoCommands = 21 // Command1 "Search", Command2 "Open"
	boundaryDoor           = 30 // Command1 "Open", Command2 "Examine"
	boundaryGate           = 31 // Command1 "Open", Command2 "Close" (real command2)
	wearableItem           = 40 // Command "Wield", IsWearable
	plainItem              = 41 // no command, not wearable
	foodItem               = 42 // Command "Eat"
)

// fakeFacts builds an in-memory *facts.Facts with the fixed defs the dispatch
// tests need, exercising Dispatch's real defsFor lookup path (rather than feeding
// MenuDefs directly) so the test proves Dispatch re-derives the option list from
// the same facts a /pick would have used.
func fakeFacts() *facts.Facts {
	return &facts.Facts{
		NpcDefs: map[int]*facts.NpcDef{
			monsterNpc:   {Name: "Goblin", Command1: "Attack", Attackable: true},
			talkNpc:      {Name: "Shopkeeper", Command1: "Talk-to"},
			thievableNpc: {Name: "Man", Command1: "Pickpocket", Attackable: true},
			command2Npc:  {Name: "Master Thief", Command1: "Talk-to", Command2: "Pickpocket"},
		},
		SceneryDefs: map[int]*facts.SceneryDef{
			sceneryWithCommand:     {Name: "Rock", Command1: "Mine", Command2: "Examine"},
			sceneryWithTwoCommands: {Name: "Chest", Command1: "Search", Command2: "Open"},
		},
		BoundaryDefs: map[int]*facts.BoundaryDef{
			boundaryDoor: {Name: "Door", Command1: "Open", Command2: "Examine"},
			boundaryGate: {Name: "Gate", Command1: "Open", Command2: "Close"},
		},
		ItemDefs: map[int]*facts.ItemDef{
			wearableItem: {Name: "Bronze sword", Command: "Wield", IsWearable: true},
			plainItem:    {Name: "Coins", IsStackable: true},
			foodItem:     {Name: "Lobster", Command: "Eat"},
		},
	}
}

// dispatch_test.go is the deterministic unit suite for the Layer 2 Dispatcher: it
// proves that a {MenuTarget, optionId} pair routes to the EXACT ActionHost method
// with the EXACT arguments and returns the right execution Lane, with no live Host
// and no packets. It leans on the §7 testability seam — Dispatcher accepts the
// narrow ActionHost / ExamineHost / WorldView interfaces, so a recording mock can
// stand in for *runtime.Host.
//
// The §9 coverage matrix: NPC verb-route branches (Attack / Talk-to / generic
// NpcCommand), the single 1-based InteractAt (scenery), 0-based inventory slots +
// stale re-validation, the boundary same-packet collapse, examine-no-packet, and
// the Lane returned for walk / follow / action / sync.

// mockHost records the last ActionHost call so a test can assert which method
// fired and with what args. Every method appends a call record; field names map
// to the wire arg meaning so assertions read clearly.
type mockHost struct {
	calls []hostCall
	err   error // optional error to return from the next call
}

type hostCall struct {
	method string
	args   []int
	name   string // for Follow
}

func (m *mockHost) record(method string, name string, args ...int) error {
	m.calls = append(m.calls, hostCall{method: method, args: args, name: name})
	return m.err
}
func (m *mockHost) last() (hostCall, bool) {
	if len(m.calls) == 0 {
		return hostCall{}, false
	}
	return m.calls[len(m.calls)-1], true
}

func (m *mockHost) Walk(_ context.Context, x, y int) error { return m.record("Walk", "", x, y) }
func (m *mockHost) AttackNpc(_ context.Context, i int) error {
	return m.record("AttackNpc", "", i)
}
func (m *mockHost) TalkToNpc(_ context.Context, i int) error {
	return m.record("TalkToNpc", "", i)
}
func (m *mockHost) NpcCommand(_ context.Context, i int) error {
	return m.record("NpcCommand", "", i)
}
func (m *mockHost) InitTradeRequest(_ context.Context, i int) error {
	return m.record("InitTradeRequest", "", i)
}
func (m *mockHost) InitDuelRequest(_ context.Context, i int) error {
	return m.record("InitDuelRequest", "", i)
}
func (m *mockHost) AttackPlayer(_ context.Context, i int) error {
	return m.record("AttackPlayer", "", i)
}
func (m *mockHost) Follow(_ context.Context, name string, _ time.Duration) error {
	return m.record("Follow", name)
}
func (m *mockHost) PickUpItem(_ context.Context, x, y, id int) error {
	return m.record("PickUpItem", "", x, y, id)
}
func (m *mockHost) InteractAt(_ context.Context, x, y, opt int) error {
	return m.record("InteractAt", "", x, y, opt)
}
func (m *mockHost) InteractWithBoundary(_ context.Context, x, y, dir int) error {
	return m.record("InteractWithBoundary", "", x, y, dir)
}
func (m *mockHost) ItemCommand(_ context.Context, slot int) error {
	return m.record("ItemCommand", "", slot)
}
func (m *mockHost) DropItem(_ context.Context, slot int) error {
	return m.record("DropItem", "", slot)
}
func (m *mockHost) EquipItem(_ context.Context, slot int) error {
	return m.record("EquipItem", "", slot)
}
func (m *mockHost) UnequipItem(_ context.Context, slot int) error {
	return m.record("UnequipItem", "", slot)
}
func (m *mockHost) UseItemOnItem(_ context.Context, a, b int) error {
	return m.record("UseItemOnItem", "", a, b)
}
func (m *mockHost) UseItemOnScenery(_ context.Context, x, y, slot int) error {
	return m.record("UseItemOnScenery", "", x, y, slot)
}
func (m *mockHost) UseItemOnBoundary(_ context.Context, x, y, dir, slot int) error {
	return m.record("UseItemOnBoundary", "", x, y, dir, slot)
}
func (m *mockHost) UseItemOnGroundItem(_ context.Context, x, y, id, slot int) error {
	return m.record("UseItemOnGroundItem", "", x, y, id, slot)
}
func (m *mockHost) UseItemOnNpc(_ context.Context, i, slot int) error {
	return m.record("UseItemOnNpc", "", i, slot)
}
func (m *mockHost) UseItemOnPlayer(_ context.Context, i, slot int) error {
	return m.record("UseItemOnPlayer", "", i, slot)
}

// mockWorld is a recording WorldView that reports a fixed visibility / slot state.
type mockWorld struct {
	npcVisible    map[int]bool
	playerVisible map[int]bool
	slots         map[int]struct {
		id      int
		wielded bool
	}
}

func (w *mockWorld) NpcVisible(i int) bool    { return w.npcVisible[i] }
func (w *mockWorld) PlayerVisible(i int) bool { return w.playerVisible[i] }
func (w *mockWorld) SlotItem(slot int) (int, bool, bool) {
	s, ok := w.slots[slot]
	if !ok {
		return 0, false, false
	}
	return s.id, s.wielded, true
}

// mockExamine returns a fixed Examination for any kind so the sync-lane test can
// assert the message is the examine String() without a live Host.
type mockExamine struct{ ex runtime.Examination }

func (m mockExamine) ExamineNpc(int) runtime.Examination           { return m.ex }
func (m mockExamine) ExaminePlayer(int) runtime.Examination        { return m.ex }
func (m mockExamine) ExamineSelf() runtime.Examination             { return m.ex }
func (m mockExamine) ExamineGroundItem(int, int) runtime.Examination { return m.ex }
func (m mockExamine) ExamineScenery(int, int) runtime.Examination  { return m.ex }
func (m mockExamine) ExamineBoundary(int, int) runtime.Examination { return m.ex }
func (m mockExamine) ExamineInventorySlot(int) runtime.Examination { return m.ex }
func (m mockExamine) ExamineItem(int) runtime.Examination          { return m.ex }

// optID looks up the wire optionId of a given OptionID in the menu BuildMenu would
// produce for (kind, defs). It is how the test references options stably — by
// meaning, not by hard-coded index — exactly as the handler does at dispatch time.
func optID(t *testing.T, kind TargetKind, defs MenuDefs, want OptionID) int {
	t.Helper()
	_, ids := BuildMenu(kind, defs)
	for i, id := range ids {
		if id == want {
			return i
		}
	}
	t.Fatalf("option %q not present in menu for kind %q (ids=%v)", want, kind, ids)
	return -1
}

func TestDispatchRouting(t *testing.T) {
	bg := context.Background()

	t.Run("scenery Command1 -> InteractAt 1-based option=1", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), nil)
		ref := MenuTarget{Kind: KindScenery, X: 222, Y: 333, ID: sceneryWithCommand}
		defs := MenuDefs{Scenery: fakeFacts().SceneryDef(sceneryWithCommand)}
		oid := optID(t, KindScenery, defs, OptCommand1)
		msg, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneAction {
			t.Errorf("lane = %v, want LaneAction", lane)
		}
		c, _ := mh.last()
		if c.method != "InteractAt" {
			t.Fatalf("method = %q, want InteractAt", c.method)
		}
		// 1-based option ordinal: the single 1-based int in the whole table.
		if got := c.args; len(got) != 3 || got[0] != 222 || got[1] != 333 || got[2] != 1 {
			t.Errorf("InteractAt args = %v, want [222 333 1]", got)
		}
		_ = msg
	})

	t.Run("scenery Command2 -> InteractAt option=2", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), nil)
		ref := MenuTarget{Kind: KindScenery, X: 10, Y: 20, ID: sceneryWithTwoCommands}
		defs := MenuDefs{Scenery: fakeFacts().SceneryDef(sceneryWithTwoCommands)}
		oid := optID(t, KindScenery, defs, OptCommand2)
		_, _, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		c, _ := mh.last()
		if c.method != "InteractAt" || len(c.args) != 3 || c.args[2] != 2 {
			t.Errorf("got %q %v, want InteractAt [10 20 2]", c.method, c.args)
		}
	})

	t.Run("boundary Command1 -> InteractWithBoundary same packet (no ordinal)", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), nil)
		ref := MenuTarget{Kind: KindBoundary, X: 5, Y: 6, Dir: 3, ID: boundaryDoor}
		defs := MenuDefs{Boundary: fakeFacts().BoundaryDef(boundaryDoor)}
		oid := optID(t, KindBoundary, defs, OptCommand1)
		_, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneAction {
			t.Errorf("lane = %v, want LaneAction", lane)
		}
		c, _ := mh.last()
		if c.method != "InteractWithBoundary" || len(c.args) != 3 ||
			c.args[0] != 5 || c.args[1] != 6 || c.args[2] != 3 {
			t.Errorf("got %q %v, want InteractWithBoundary [5 6 3]", c.method, c.args)
		}
	})

	t.Run("ground item Pick up -> PickUpItem(x,y,itemID)", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, nil, nil)
		ref := MenuTarget{Kind: KindGroundItem, X: 7, Y: 8, ID: 42}
		oid := optID(t, KindGroundItem, MenuDefs{}, OptPickup)
		_, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneAction {
			t.Errorf("lane = %v, want LaneAction", lane)
		}
		c, _ := mh.last()
		if c.method != "PickUpItem" || len(c.args) != 3 ||
			c.args[0] != 7 || c.args[1] != 8 || c.args[2] != 42 {
			t.Errorf("got %q %v, want PickUpItem [7 8 42]", c.method, c.args)
		}
	})

	t.Run("terrain Walk here -> Walk + LaneWalk", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, nil, nil)
		ref := MenuTarget{Kind: KindTerrain, X: 100, Y: 200}
		oid := optID(t, KindTerrain, MenuDefs{}, OptWalkHere)
		_, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneWalk {
			t.Errorf("lane = %v, want LaneWalk", lane)
		}
		c, _ := mh.last()
		if c.method != "Walk" || len(c.args) != 2 || c.args[0] != 100 || c.args[1] != 200 {
			t.Errorf("got %q %v, want Walk [100 200]", c.method, c.args)
		}
	})

	t.Run("player Trade/Duel/Attack route + visible", func(t *testing.T) {
		mw := &mockWorld{playerVisible: map[int]bool{9: true}}
		cases := []struct {
			opt    OptionID
			method string
		}{
			{OptTrade, "InitTradeRequest"},
			{OptDuel, "InitDuelRequest"},
			{OptAttack, "AttackPlayer"},
		}
		for _, tc := range cases {
			mh := &mockHost{}
			d := NewDispatcher(mh, nil, nil, mw)
			ref := MenuTarget{Kind: KindPlayer, Index: 9, Name: "Bob"}
			oid := optID(t, KindPlayer, MenuDefs{}, tc.opt)
			_, lane, err := d.Dispatch(bg, ref, oid)
			if err != nil {
				t.Fatalf("%s Dispatch err: %v", tc.method, err)
			}
			if lane != LaneAction {
				t.Errorf("%s lane = %v, want LaneAction", tc.method, lane)
			}
			c, _ := mh.last()
			if c.method != tc.method || len(c.args) != 1 || c.args[0] != 9 {
				t.Errorf("got %q %v, want %s [9]", c.method, c.args, tc.method)
			}
		}
	})

	t.Run("player Follow -> LaneFollow keyed by name", func(t *testing.T) {
		mw := &mockWorld{playerVisible: map[int]bool{9: true}}
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, nil, mw)
		ref := MenuTarget{Kind: KindPlayer, Index: 9, Name: "Bob"}
		oid := optID(t, KindPlayer, MenuDefs{}, OptFollow)
		_, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneFollow {
			t.Errorf("lane = %v, want LaneFollow", lane)
		}
		c, _ := mh.last()
		if c.method != "Follow" || c.name != "Bob" {
			t.Errorf("got %q name=%q, want Follow name=Bob", c.method, c.name)
		}
	})

	t.Run("inventory slots are 0-based: Wield/Remove/Drop route by slot", func(t *testing.T) {
		// worn item -> Remove (UnequipItem); un-worn wearable -> Wield (EquipItem).
		cases := []struct {
			name    string
			wielded bool
			opt     OptionID
			method  string
		}{
			{"wield un-worn", false, OptWield, "EquipItem"},
			{"remove worn", true, OptRemove, "UnequipItem"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				mw := &mockWorld{slots: map[int]struct {
					id      int
					wielded bool
				}{3: {id: wearableItem, wielded: tc.wielded}}}
				mh := &mockHost{}
				d := NewDispatcher(mh, nil, fakeFacts(), mw)
				ref := MenuTarget{Kind: KindInventoryItem, Slot: 3, ID: wearableItem}
				// The menu shape depends on the live wielded state (re-read in defsFor).
				defs := MenuDefs{Item: fakeFacts().ItemDef(wearableItem)}
				defs.InvSlot.Wielded = tc.wielded
				oid := optID(t, KindInventoryItem, defs, tc.opt)
				_, lane, err := d.Dispatch(bg, ref, oid)
				if err != nil {
					t.Fatalf("Dispatch err: %v", err)
				}
				if lane != LaneAction {
					t.Errorf("lane = %v, want LaneAction", lane)
				}
				c, _ := mh.last()
				if c.method != tc.method || len(c.args) != 1 || c.args[0] != 3 {
					t.Errorf("got %q %v, want %s [3] (0-based slot)", c.method, c.args, tc.method)
				}
			})
		}
	})

	t.Run("inventory Drop -> DropItem(slot)", func(t *testing.T) {
		mw := &mockWorld{slots: map[int]struct {
			id      int
			wielded bool
		}{0: {id: 5}}}
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindInventoryItem, Slot: 0, ID: 5}
		oid := optID(t, KindInventoryItem, MenuDefs{Item: fakeFacts().ItemDef(plainItem)}, OptDrop)
		_, _, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		c, _ := mh.last()
		if c.method != "DropItem" || len(c.args) != 1 || c.args[0] != 0 {
			t.Errorf("got %q %v, want DropItem [0]", c.method, c.args)
		}
	})

	t.Run("inventory Command (Eat) -> ItemCommand(slot)", func(t *testing.T) {
		mw := &mockWorld{slots: map[int]struct {
			id      int
			wielded bool
		}{4: {id: foodItem}}}
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindInventoryItem, Slot: 4, ID: foodItem}
		oid := optID(t, KindInventoryItem, MenuDefs{Item: fakeFacts().ItemDef(foodItem)}, OptCommand)
		msg, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneAction {
			t.Errorf("lane = %v, want LaneAction", lane)
		}
		// The message reports the authentic def verb (Eat), not a generic "Use".
		if msg != "Eat" {
			t.Errorf("msg = %q, want %q", msg, "Eat")
		}
		c, _ := mh.last()
		if c.method != "ItemCommand" || len(c.args) != 1 || c.args[0] != 4 {
			t.Errorf("got %q %v, want ItemCommand [4]", c.method, c.args)
		}
	})

	t.Run("inventory Use is an M2 no-op: no packet, LaneSync", func(t *testing.T) {
		mw := &mockWorld{slots: map[int]struct {
			id      int
			wielded bool
		}{4: {id: foodItem}}}
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		// Use is not in the M1 menu for any kind, so dispatch it by constructing the
		// option directly (the dispatcher must still handle the reserved verb safely).
		// Build a target whose option list we know, then call the inventory path with
		// an explicit OptUse via the private dispatch helper is not accessible; instead
		// assert through the public surface using a def whose menu would never include
		// Use — so we cover the dispatchInventory OptUse arm via a crafted call.
		// Since OptUse is never emitted by BuildMenu, we verify the arm directly.
		_, lane, err := d.dispatchInventory(bg, MenuTarget{Kind: KindInventoryItem, Slot: 4, ID: foodItem}, OptUse, fakeFacts().ItemDef(foodItem))
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneSync {
			t.Errorf("use lane = %v, want LaneSync", lane)
		}
		if len(mh.calls) != 0 {
			t.Errorf("use fired %d host calls, want 0 (M1 no-op)", len(mh.calls))
		}
	})

	t.Run("scenery Command1 message is the authentic def verb", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), nil)
		ref := MenuTarget{Kind: KindScenery, X: 1, Y: 2, ID: sceneryWithCommand}
		defs := MenuDefs{Scenery: fakeFacts().SceneryDef(sceneryWithCommand)}
		oid := optID(t, KindScenery, defs, OptCommand1)
		msg, _, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if msg != "Mine" {
			t.Errorf("msg = %q, want %q", msg, "Mine")
		}
	})

	t.Run("boundary Command2 -> same InteractWithBoundary packet, message Close", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), nil)
		ref := MenuTarget{Kind: KindBoundary, X: 5, Y: 6, Dir: 2, ID: boundaryGate}
		defs := MenuDefs{Boundary: fakeFacts().BoundaryDef(boundaryGate)}
		oid := optID(t, KindBoundary, defs, OptCommand2)
		msg, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("Dispatch err: %v", err)
		}
		if lane != LaneAction {
			t.Errorf("lane = %v, want LaneAction", lane)
		}
		if msg != "Close" {
			t.Errorf("msg = %q, want %q", msg, "Close")
		}
		c, _ := mh.last()
		// command2 must fire the SAME InteractWithBoundary packet as command1 (server toggles).
		if c.method != "InteractWithBoundary" || len(c.args) != 3 ||
			c.args[0] != 5 || c.args[1] != 6 || c.args[2] != 2 {
			t.Errorf("got %q %v, want InteractWithBoundary [5 6 2]", c.method, c.args)
		}
	})

	t.Run("npc plain attack/talk_to OptionIDs route directly", func(t *testing.T) {
		mw := &mockWorld{npcVisible: map[int]bool{1: true}}
		cases := []struct {
			opt    OptionID
			method string
		}{
			{OptAttack, "AttackNpc"},
			{OptTalkTo, "TalkToNpc"},
		}
		for _, tc := range cases {
			mh := &mockHost{}
			d := NewDispatcher(mh, nil, fakeFacts(), mw)
			// dispatchNPC handles OptAttack/OptTalkTo arms directly; these are not in the
			// thievable menu's command list but are valid dispatch keys, so exercise the
			// arm via the private helper (the thievable NPC has commands, so BuildMenu
			// emits OptCommand1/OptCommand2; the bare talk_to/attack keys are tested
			// here via the private helper rather than through BuildMenu's output).
			ref := MenuTarget{Kind: KindNPC, Index: 1, ID: thievableNpc}
			_, lane, err := d.dispatchNPC(bg, ref, tc.opt, fakeFacts().NpcDef(thievableNpc))
			if err != nil {
				t.Fatalf("%s Dispatch err: %v", tc.method, err)
			}
			if lane != LaneAction {
				t.Errorf("%s lane = %v, want LaneAction", tc.method, lane)
			}
			c, _ := mh.last()
			if c.method != tc.method || len(c.args) != 1 || c.args[0] != 1 {
				t.Errorf("got %q %v, want %s [1]", c.method, c.args, tc.method)
			}
		}
	})
}

// TestDispatchNpcCommand2VerbRoute proves command2 verb-routes exactly like
// command1 — a def whose Command2 is "Pickpocket" routes command2 -> NpcCommand,
// while its Command1 "Talk-to" routes -> TalkToNpc.
func TestDispatchNpcCommand2VerbRoute(t *testing.T) {
	bg := context.Background()
	mw := &mockWorld{npcVisible: map[int]bool{1: true}}
	defs := MenuDefs{Npc: fakeFacts().NpcDef(command2Npc)}

	cases := []struct {
		opt    OptionID
		method string
	}{
		{OptCommand1, "TalkToNpc"},  // Command1 "Talk-to"
		{OptCommand2, "NpcCommand"}, // Command2 "Pickpocket"
	}
	for _, tc := range cases {
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindNPC, Index: 1, ID: command2Npc}
		oid := optID(t, KindNPC, defs, tc.opt)
		_, lane, err := d.Dispatch(bg, ref, oid)
		if err != nil {
			t.Fatalf("%s Dispatch err: %v", tc.method, err)
		}
		if lane != LaneAction {
			t.Errorf("%s lane = %v, want LaneAction", tc.method, lane)
		}
		c, _ := mh.last()
		if c.method != tc.method || len(c.args) != 1 || c.args[0] != 1 {
			t.Errorf("got %q %v, want %s [1]", c.method, c.args, tc.method)
		}
	}
}

// TestDispatchExamineAllKinds proves every examine arm resolves on LaneSync,
// fires NO host packet, and returns the Examination.String() — across npc,
// player, self, ground_item, scenery, boundary, and inventory_item.
func TestDispatchExamineAllKinds(t *testing.T) {
	bg := context.Background()
	ex := mockExamine{ex: runtime.Examination{Kind: "thing", Name: "A Thing", Description: "An examinable thing."}}

	cases := []struct {
		name string
		kind TargetKind
		ref  MenuTarget
		defs MenuDefs
	}{
		{"npc", KindNPC, MenuTarget{Kind: KindNPC, Index: 4, ID: monsterNpc}, MenuDefs{Npc: fakeFacts().NpcDef(monsterNpc)}},
		{"player", KindPlayer, MenuTarget{Kind: KindPlayer, Index: 7, Name: "Bob"}, MenuDefs{}},
		{"self", KindSelf, MenuTarget{Kind: KindSelf}, MenuDefs{}},
		{"ground_item", KindGroundItem, MenuTarget{Kind: KindGroundItem, X: 3, Y: 4, ID: 9}, MenuDefs{}},
		{"scenery", KindScenery, MenuTarget{Kind: KindScenery, X: 3, Y: 4, ID: sceneryWithCommand}, MenuDefs{Scenery: fakeFacts().SceneryDef(sceneryWithCommand)}},
		{"boundary", KindBoundary, MenuTarget{Kind: KindBoundary, X: 3, Y: 4, Dir: 1, ID: boundaryDoor}, MenuDefs{Boundary: fakeFacts().BoundaryDef(boundaryDoor)}},
		{"inventory_item", KindInventoryItem, MenuTarget{Kind: KindInventoryItem, Slot: 2, ID: foodItem}, MenuDefs{Item: fakeFacts().ItemDef(foodItem)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mh := &mockHost{}
			// World wired so visibility/slot re-validation passes for the kinds that check.
			mw := &mockWorld{
				npcVisible:    map[int]bool{4: true},
				playerVisible: map[int]bool{7: true},
				slots: map[int]struct {
					id      int
					wielded bool
				}{2: {id: foodItem}},
			}
			d := NewDispatcher(mh, ex, fakeFacts(), mw)
			oid := optID(t, tc.kind, tc.defs, OptExamine)
			msg, lane, err := d.Dispatch(bg, tc.ref, oid)
			if err != nil {
				t.Fatalf("Dispatch err: %v", err)
			}
			if lane != LaneSync {
				t.Errorf("examine lane = %v, want LaneSync", lane)
			}
			if len(mh.calls) != 0 {
				t.Errorf("examine fired %d host calls, want 0", len(mh.calls))
			}
			if msg != ex.ex.String() {
				t.Errorf("examine msg = %q, want %q", msg, ex.ex.String())
			}
		})
	}
}

func TestDispatchExamineNoPacket(t *testing.T) {
	bg := context.Background()
	ex := mockExamine{ex: runtime.Examination{Kind: "npc", Name: "Goblin", Description: "An ugly green creature."}}
	mh := &mockHost{}
	mw := &mockWorld{npcVisible: map[int]bool{4: true}}
	d := NewDispatcher(mh, ex, fakeFacts(), mw)

	ref := MenuTarget{Kind: KindNPC, Index: 4, ID: monsterNpc}
	defs := MenuDefs{Npc: fakeFacts().NpcDef(monsterNpc)}
	oid := optID(t, KindNPC, defs, OptExamine)

	msg, lane, err := d.Dispatch(bg, ref, oid)
	if err != nil {
		t.Fatalf("Dispatch err: %v", err)
	}
	if lane != LaneSync {
		t.Errorf("examine lane = %v, want LaneSync", lane)
	}
	if len(mh.calls) != 0 {
		t.Errorf("examine fired %d host calls, want 0 (examine sends no packet)", len(mh.calls))
	}
	if msg != ex.ex.String() {
		t.Errorf("examine msg = %q, want %q", msg, ex.ex.String())
	}
}

func TestDispatchNpcVerbRoute(t *testing.T) {
	bg := context.Background()
	mw := &mockWorld{npcVisible: map[int]bool{1: true}}

	// Each NPC def shape drives a different Command1 verb; assert it routes to the
	// right wire primitive: Attack->AttackNpc, Talk-to->TalkToNpc, else->NpcCommand.
	cases := []struct {
		name   string
		defID  int
		method string
	}{
		{"Attack verb -> AttackNpc", monsterNpc, "AttackNpc"},
		{"Talk-to verb -> TalkToNpc", talkNpc, "TalkToNpc"},
		{"Pickpocket verb -> NpcCommand", thievableNpc, "NpcCommand"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mh := &mockHost{}
			d := NewDispatcher(mh, nil, fakeFacts(), mw)
			ref := MenuTarget{Kind: KindNPC, Index: 1, ID: tc.defID}
			defs := MenuDefs{Npc: fakeFacts().NpcDef(tc.defID)}
			oid := optID(t, KindNPC, defs, OptCommand1)
			_, lane, err := d.Dispatch(bg, ref, oid)
			if err != nil {
				t.Fatalf("Dispatch err: %v", err)
			}
			if lane != LaneAction {
				t.Errorf("lane = %v, want LaneAction", lane)
			}
			c, _ := mh.last()
			if c.method != tc.method || len(c.args) != 1 || c.args[0] != 1 {
				t.Errorf("got %q %v, want %s [1]", c.method, c.args, tc.method)
			}
		})
	}
}

func TestDispatchStaleTarget(t *testing.T) {
	bg := context.Background()

	t.Run("npc no longer visible -> ErrStaleTarget, no packet", func(t *testing.T) {
		mw := &mockWorld{npcVisible: map[int]bool{}} // index 4 NOT visible
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		// thievableNpc has a synthetic Attack (Command1 is "Pickpocket", not "Attack").
		ref := MenuTarget{Kind: KindNPC, Index: 4, ID: thievableNpc}
		defs := MenuDefs{Npc: fakeFacts().NpcDef(thievableNpc)}
		oid := optID(t, KindNPC, defs, OptAttack)
		_, _, err := d.Dispatch(bg, ref, oid)
		if !errors.Is(err, ErrStaleTarget) {
			t.Errorf("err = %v, want ErrStaleTarget", err)
		}
		if len(mh.calls) != 0 {
			t.Errorf("fired %d calls on stale target, want 0", len(mh.calls))
		}
	})

	t.Run("inventory slot shifted to a different item -> ErrStaleTarget", func(t *testing.T) {
		mw := &mockWorld{slots: map[int]struct {
			id      int
			wielded bool
		}{2: {id: 999}}} // slot now holds 999, not the expected 5
		mh := &mockHost{}
		d := NewDispatcher(mh, nil, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindInventoryItem, Slot: 2, ID: 5}
		oid := optID(t, KindInventoryItem, MenuDefs{Item: fakeFacts().ItemDef(plainItem)}, OptDrop)
		_, _, err := d.Dispatch(bg, ref, oid)
		if !errors.Is(err, ErrStaleTarget) {
			t.Errorf("err = %v, want ErrStaleTarget", err)
		}
		if len(mh.calls) != 0 {
			t.Errorf("fired %d calls on stale slot, want 0", len(mh.calls))
		}
	})
}

// failExamine is an ExamineHost that fails the test if ANY accessor is called.
// Used by the ResolveLane purity test: ResolveLane must read no examine text
// (and fire no packet) — it only routes.
type failExamine struct{ t *testing.T }

func (e failExamine) fail(m string) runtime.Examination {
	e.t.Helper()
	e.t.Fatalf("ResolveLane called ExamineHost.%s — it must be side-effect-free", m)
	return runtime.Examination{}
}
func (e failExamine) ExamineNpc(int) runtime.Examination           { return e.fail("ExamineNpc") }
func (e failExamine) ExaminePlayer(int) runtime.Examination        { return e.fail("ExaminePlayer") }
func (e failExamine) ExamineSelf() runtime.Examination             { return e.fail("ExamineSelf") }
func (e failExamine) ExamineGroundItem(int, int) runtime.Examination { return e.fail("ExamineGroundItem") }
func (e failExamine) ExamineScenery(int, int) runtime.Examination  { return e.fail("ExamineScenery") }
func (e failExamine) ExamineBoundary(int, int) runtime.Examination { return e.fail("ExamineBoundary") }
func (e failExamine) ExamineInventorySlot(int) runtime.Examination { return e.fail("ExamineInventorySlot") }
func (e failExamine) ExamineItem(int) runtime.Examination          { return e.fail("ExamineItem") }

// TestResolveLanePure proves ResolveLane is the pure routing path the /act
// handler relies on: across EVERY kind+option it must fire ZERO ActionHost calls
// and ZERO ExamineHost calls while returning the correct Lane. A mock ActionHost
// that records calls (asserted empty) and an ExamineHost that t.Fatals on any
// call together guarantee no packet and no side effect. This is the regression
// guard for the double-dispatch bug: the handler routes via ResolveLane, never
// via Dispatch (which executes as it resolves).
func TestResolveLanePure(t *testing.T) {
	mw := &mockWorld{
		npcVisible:    map[int]bool{1: true},
		playerVisible: map[int]bool{9: true},
		slots: map[int]struct {
			id      int
			wielded bool
		}{3: {id: wearableItem, wielded: false}, 4: {id: foodItem}},
	}

	cases := []struct {
		name string
		ref  MenuTarget
		defs MenuDefs
		opt  OptionID
		want Lane
	}{
		{"npc attack -> action", MenuTarget{Kind: KindNPC, Index: 1, ID: monsterNpc}, MenuDefs{Npc: fakeFacts().NpcDef(monsterNpc)}, OptCommand1, LaneAction},
		{"npc examine -> sync", MenuTarget{Kind: KindNPC, Index: 1, ID: monsterNpc}, MenuDefs{Npc: fakeFacts().NpcDef(monsterNpc)}, OptExamine, LaneSync},
		{"player trade -> action", MenuTarget{Kind: KindPlayer, Index: 9, Name: "Bob"}, MenuDefs{}, OptTrade, LaneAction},
		{"player follow -> follow", MenuTarget{Kind: KindPlayer, Index: 9, Name: "Bob"}, MenuDefs{}, OptFollow, LaneFollow},
		{"player examine -> sync", MenuTarget{Kind: KindPlayer, Index: 9, Name: "Bob"}, MenuDefs{}, OptExamine, LaneSync},
		{"self examine -> sync", MenuTarget{Kind: KindSelf}, MenuDefs{}, OptExamine, LaneSync},
		{"ground pickup -> action", MenuTarget{Kind: KindGroundItem, X: 7, Y: 8, ID: 42}, MenuDefs{}, OptPickup, LaneAction},
		{"scenery cmd1 -> action", MenuTarget{Kind: KindScenery, X: 1, Y: 2, ID: sceneryWithCommand}, MenuDefs{Scenery: fakeFacts().SceneryDef(sceneryWithCommand)}, OptCommand1, LaneAction},
		{"boundary cmd1 -> action", MenuTarget{Kind: KindBoundary, X: 5, Y: 6, Dir: 3, ID: boundaryDoor}, MenuDefs{Boundary: fakeFacts().BoundaryDef(boundaryDoor)}, OptCommand1, LaneAction},
		{"terrain walk -> walk", MenuTarget{Kind: KindTerrain, X: 100, Y: 200}, MenuDefs{}, OptWalkHere, LaneWalk},
		{"inv wield -> action", MenuTarget{Kind: KindInventoryItem, Slot: 3, ID: wearableItem}, MenuDefs{Item: fakeFacts().ItemDef(wearableItem)}, OptWield, LaneAction},
		{"inv examine -> sync", MenuTarget{Kind: KindInventoryItem, Slot: 4, ID: foodItem}, MenuDefs{Item: fakeFacts().ItemDef(foodItem)}, OptExamine, LaneSync},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mh := &mockHost{}
			d := NewDispatcher(mh, failExamine{t}, fakeFacts(), mw)
			oid := optID(t, tc.ref.Kind, tc.defs, tc.opt)
			lane, err := d.ResolveLane(tc.ref, oid)
			if err != nil {
				t.Fatalf("ResolveLane err: %v", err)
			}
			if lane != tc.want {
				t.Errorf("lane = %v, want %v", lane, tc.want)
			}
			if len(mh.calls) != 0 {
				t.Fatalf("ResolveLane fired %d ActionHost calls (%v), want 0 — it must be side-effect-free", len(mh.calls), mh.calls)
			}
		})
	}
}

// TestResolveLaneValidation proves ResolveLane reports the same malformed/stale
// errors Dispatch would, WITHOUT a packet, so the handler can route + reject in
// one pure call.
func TestResolveLaneValidation(t *testing.T) {
	t.Run("out-of-range option -> ErrUnknownOption", func(t *testing.T) {
		mh := &mockHost{}
		d := NewDispatcher(mh, failExamine{t}, nil, nil)
		_, err := d.ResolveLane(MenuTarget{Kind: KindTerrain, X: 1, Y: 1}, 5)
		if !errors.Is(err, ErrUnknownOption) {
			t.Errorf("err = %v, want ErrUnknownOption", err)
		}
		if len(mh.calls) != 0 {
			t.Errorf("fired %d calls, want 0", len(mh.calls))
		}
	})

	t.Run("npc no longer visible -> ErrStaleTarget", func(t *testing.T) {
		mw := &mockWorld{npcVisible: map[int]bool{}} // index 4 NOT visible
		mh := &mockHost{}
		d := NewDispatcher(mh, failExamine{t}, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindNPC, Index: 4, ID: thievableNpc}
		defs := MenuDefs{Npc: fakeFacts().NpcDef(thievableNpc)}
		oid := optID(t, KindNPC, defs, OptAttack)
		_, err := d.ResolveLane(ref, oid)
		if !errors.Is(err, ErrStaleTarget) {
			t.Errorf("err = %v, want ErrStaleTarget", err)
		}
		if len(mh.calls) != 0 {
			t.Errorf("fired %d calls, want 0", len(mh.calls))
		}
	})

	t.Run("inventory slot shifted -> ErrStaleTarget", func(t *testing.T) {
		mw := &mockWorld{slots: map[int]struct {
			id      int
			wielded bool
		}{2: {id: 999}}} // slot now holds 999, not the expected 5
		mh := &mockHost{}
		d := NewDispatcher(mh, failExamine{t}, fakeFacts(), mw)
		ref := MenuTarget{Kind: KindInventoryItem, Slot: 2, ID: 5}
		oid := optID(t, KindInventoryItem, MenuDefs{Item: fakeFacts().ItemDef(plainItem)}, OptDrop)
		_, err := d.ResolveLane(ref, oid)
		if !errors.Is(err, ErrStaleTarget) {
			t.Errorf("err = %v, want ErrStaleTarget", err)
		}
		if len(mh.calls) != 0 {
			t.Errorf("fired %d calls, want 0", len(mh.calls))
		}
	})
}

func TestDispatchUnknownOption(t *testing.T) {
	bg := context.Background()
	mh := &mockHost{}
	d := NewDispatcher(mh, nil, nil, nil)
	ref := MenuTarget{Kind: KindTerrain, X: 1, Y: 1}
	// Terrain menu has exactly one option (index 0); index 5 is out of range.
	_, _, err := d.Dispatch(bg, ref, 5)
	if !errors.Is(err, ErrUnknownOption) {
		t.Errorf("err = %v, want ErrUnknownOption", err)
	}
	if len(mh.calls) != 0 {
		t.Errorf("fired %d calls on bad optionId, want 0", len(mh.calls))
	}
}

// ensure *mockHost satisfies ActionHost and the mocks satisfy their interfaces at
// compile time — a guard against silent interface drift breaking the test.
var (
	_ ActionHost  = (*mockHost)(nil)
	_ WorldView   = (*mockWorld)(nil)
	_ ExamineHost = mockExamine{}
)
