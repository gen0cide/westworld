package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// ===== #119: world.messages bounded ring =====

// The ring keeps multiple server messages (oldest-first), not just
// the latest, and world.messages exposes them as a List<Message>.
func TestWorldMessagesRingKeepsHistory(t *testing.T) {
	h := newTestHost()
	h.world.Recent.SetServerMessage("first")
	h.world.Recent.SetServerMessage("second")
	h.world.Recent.SetServerMessage("third")

	res := runRoutine(t, h, `routine r() { return world.messages.length }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 3 {
		t.Fatalf("world.messages.length: got %v, want Int(3)", res.Value)
	}
	// oldest-first ordering: index 0 is the first message, last is newest.
	res = runRoutine(t, h, `routine r() { return world.messages.first.text }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "first" {
		t.Errorf("world.messages.first.text: got %v, want \"first\"", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return world.messages.last.text }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "third" {
		t.Errorf("world.messages.last.text: got %v, want \"third\"", res.Value)
	}
	// last_server_message still mirrors the newest entry.
	res = runRoutine(t, h, `routine r() { return world.last_server_message.message }`)
	if s, ok := res.Value.(interp.String); !ok || string(s) != "third" {
		t.Errorf("world.last_server_message.message: got %v, want \"third\"", res.Value)
	}
}

// The ring is bounded: once more than ServerMsgRingCap messages
// arrive, the oldest fall off the front.
func TestWorldMessagesRingIsBounded(t *testing.T) {
	h := newTestHost()
	total := world.ServerMsgRingCap + 10
	for i := 0; i < total; i++ {
		h.world.Recent.SetServerMessage(intDisp(i))
	}
	msgs := h.world.Recent.ServerMessages()
	if len(msgs) != world.ServerMsgRingCap {
		t.Fatalf("ring length: got %d, want %d (capped)", len(msgs), world.ServerMsgRingCap)
	}
	// Oldest survivor is message #10 (the first `total-cap` were dropped).
	wantOldest := intDisp(total - world.ServerMsgRingCap)
	if msgs[0].Message != wantOldest {
		t.Errorf("oldest survivor: got %q, want %q", msgs[0].Message, wantOldest)
	}
	if msgs[len(msgs)-1].Message != intDisp(total-1) {
		t.Errorf("newest: got %q, want %q", msgs[len(msgs)-1].Message, intDisp(total-1))
	}
}

// ===== #119: translateEvents fan-out =====

// A SystemMessage produces BOTH server_message (legacy) and the new
// ring-backed message event, in that order.
func TestTranslateEventsSystemMessageFansOut(t *testing.T) {
	h := newTestHost()
	pes := translateEvents(h, event.SystemMessage{Message: "You can't go through this door."})
	if len(pes) != 2 {
		t.Fatalf("SystemMessage fan-out: got %d events, want 2", len(pes))
	}
	if pes[0].Name != "server_message" {
		t.Errorf("first event: got %q, want server_message", pes[0].Name)
	}
	if pes[1].Name != "message" {
		t.Errorf("second event: got %q, want message", pes[1].Name)
	}
	if s, ok := pes[1].Args[0].(interp.String); !ok || string(s) != "You can't go through this door." {
		t.Errorf("message arg: got %v", pes[1].Args[0])
	}
}

// XPGain surfaces as xp_gain(skill_name, amount).
func TestTranslateEventsXPGain(t *testing.T) {
	h := newTestHost()
	pes := translateEvents(h, event.XPGain{Skill: event.SkillFishing, Amount: 40, Total: 1234})
	if len(pes) != 1 || pes[0].Name != "xp_gain" {
		t.Fatalf("XPGain: got %+v, want one xp_gain", pes)
	}
	if s, ok := pes[0].Args[0].(interp.String); !ok || string(s) != "fishing" {
		t.Errorf("xp_gain skill arg: got %v, want \"fishing\"", pes[0].Args[0])
	}
	if a, ok := pes[0].Args[1].(interp.Int); !ok || int64(a) != 40 {
		t.Errorf("xp_gain amount arg: got %v, want 40", pes[0].Args[1])
	}
}

// TargetDied surfaces under both target_died and npc_killed, carrying
// the dead NPC view resolved live from world.Npcs.
func TestTranslateEventsTargetDiedAndNpcKilled(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 7, X: 100, Y: 500, TypeID: 42})
	pes := translateEvents(h, event.TargetDied{NpcIndex: 7, TypeID: 42})
	if len(pes) != 2 {
		t.Fatalf("TargetDied fan-out: got %d events, want 2", len(pes))
	}
	if pes[0].Name != "target_died" || pes[1].Name != "npc_killed" {
		t.Errorf("event names: got %q,%q want target_died,npc_killed", pes[0].Name, pes[1].Name)
	}
	nv, ok := pes[0].Args[0].(*npcView)
	if !ok {
		t.Fatalf("target arg: got %T, want *npcView", pes[0].Args[0])
	}
	if nv.record.Index != 7 {
		t.Errorf("target index: got %d, want 7", nv.record.Index)
	}
}

// When the dead NPC has already left view, the target arg is Null
// (not a crash).
func TestTranslateEventsTargetDiedGoneFromView(t *testing.T) {
	h := newTestHost()
	pes := translateEvents(h, event.TargetDied{NpcIndex: 99, TypeID: 1})
	if len(pes) != 2 {
		t.Fatalf("got %d events, want 2", len(pes))
	}
	if _, ok := pes[0].Args[0].(interp.Null); !ok {
		t.Errorf("target arg when NPC gone: got %T, want Null", pes[0].Args[0])
	}
}

// ===== #119: synthetic XP-delta + death-edge synthesis (host) =====

// emitXPGainDeltas fires only on net-positive xp changes and carries
// the delta, not the new total.
func TestEmitXPGainDeltas(t *testing.T) {
	h := newTestHost()
	sub := h.bus.Subscribe("*", 16)

	// Fishing starts at xp 0 in newTestHost. Snapshot, bump, emit.
	prev := h.skillXPSnapshot()
	h.world.Self.SetSkill(int(event.SkillFishing), 26, 26, 83) // +83 xp
	h.emitXPGainDeltas(prev)

	select {
	case ev := <-sub:
		xp, ok := ev.(event.XPGain)
		if !ok {
			t.Fatalf("got %T, want event.XPGain", ev)
		}
		if xp.Skill != event.SkillFishing {
			t.Errorf("skill: got %v, want fishing", xp.Skill)
		}
		if xp.Amount != 83 {
			t.Errorf("amount: got %d, want 83 (delta, not total)", xp.Amount)
		}
		if xp.Total != 83 {
			t.Errorf("total: got %d, want 83", xp.Total)
		}
	default:
		t.Fatal("no XPGain published for a +83 xp change")
	}
}

// No xp change -> no XPGain event.
func TestEmitXPGainDeltasNoChange(t *testing.T) {
	h := newTestHost()
	sub := h.bus.Subscribe("*", 16)
	prev := h.skillXPSnapshot()
	h.emitXPGainDeltas(prev) // nothing changed
	select {
	case ev := <-sub:
		t.Fatalf("unexpected event with no xp change: %T", ev)
	default:
	}
}

// emitTargetDeathEdge fires once on the alive->dead transition for the
// engaged target, and not on a repeated 0-hits reading.
func TestEmitTargetDeathEdge(t *testing.T) {
	h := newTestHost()
	h.lastAttackedNpcIndex.Store(5)
	h.world.Npcs.Set(world.NpcRecord{Index: 5, TypeID: 42})
	sub := h.bus.Subscribe("target_died", 8)

	// Land the killing blow: curHits 0 after the hit. wasAlive=true.
	h.world.Npcs.SetHits(5, 3, 0, 10)
	h.emitTargetDeathEdge(5, true)

	select {
	case ev := <-sub:
		td, ok := ev.(event.TargetDied)
		if !ok || td.NpcIndex != 5 || td.TypeID != 42 {
			t.Fatalf("got %#v, want TargetDied{NpcIndex:5,TypeID:42}", ev)
		}
	default:
		t.Fatal("no TargetDied published on the death edge")
	}

	// A repeated 0-hits reading (wasAlive=false) must NOT re-fire.
	h.emitTargetDeathEdge(5, false)
	select {
	case ev := <-sub:
		t.Fatalf("double-fire on repeated 0-hits: %T", ev)
	default:
	}
}

// engagedTargetAlive treats an unknown (no-reading) NPC as alive so a
// first 0-hits packet still registers as a death edge, and a known
// 0-hits NPC as dead.
func TestEngagedTargetAlive(t *testing.T) {
	h := newTestHost()
	if h.engagedTargetAlive() {
		t.Error("no engaged target -> should be not-alive")
	}
	h.lastAttackedNpcIndex.Store(3)
	h.world.Npcs.Set(world.NpcRecord{Index: 3, TypeID: 1}) // HasHits=false
	if !h.engagedTargetAlive() {
		t.Error("engaged target with no health reading -> should be alive")
	}
	h.world.Npcs.SetHits(3, 5, 0, 10) // dead
	if h.engagedTargetAlive() {
		t.Error("engaged target at 0 hits -> should be not-alive")
	}
}

// ===== #119: combat.target live resolution =====

// combat.target resolves the engaged NPC while alive, and clears to
// null once it dies.
func TestCombatTargetLiveAndClearsOnDeath(t *testing.T) {
	h := newTestHost()
	// No target yet.
	res := runRoutine(t, h, `routine r() { return combat.target }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("combat.target with no engagement: got %v, want Null", res.Value)
	}

	h.lastAttackedNpcIndex.Store(11)
	h.world.Npcs.Set(world.NpcRecord{Index: 11, X: 100, Y: 500, TypeID: 42})
	res = runRoutine(t, h, `routine r() { return combat.target.index }`)
	if i, ok := res.Value.(interp.Int); !ok || int64(i) != 11 {
		t.Errorf("combat.target.index (alive): got %v, want Int(11)", res.Value)
	}

	// Kill it -> combat.target clears.
	h.world.Npcs.SetHits(11, 4, 0, 10)
	res = runRoutine(t, h, `routine r() { return combat.target }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("combat.target after death: got %v, want Null", res.Value)
	}
}
