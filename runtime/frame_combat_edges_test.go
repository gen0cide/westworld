package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/world"
)

// Regression tests for the synthetic combat death/disengage edges
// (soak-retro 2026-06-10 cause #3, kill-detection half). The original
// detectors lived only on handleFrame's single-event tail, but every
// combat-bearing event arrives on a MULTI-event opcode whose branch
// returns early — NpcDamage on 104, NPC removals on 79, our own
// OtherPlayerDamage on 234 — so `on target_died` never fired from live
// traffic and fighters' selects timed out on winnable fights. These
// tests drive the real handleFrame paths with wire-shaped payloads.

// updateNpcDamageFrame builds an opcode-104 (SEND_UPDATE_NPC) frame
// carrying a single type-2 damage record for npcIndex: one short count,
// short index, byte type, then damage/curHits/maxHits bytes.
func updateNpcDamageFrame(npcIndex, damage, curHits, maxHits int) v235.Frame {
	return v235.Frame{
		Opcode: v235.InUpdateNpc,
		Payload: []byte{
			0x00, 0x01, // updateCount = 1
			byte(npcIndex >> 8), byte(npcIndex), // npcIndex
			0x02,         // updateType 2 = damage
			byte(damage), // damage
			byte(curHits),
			byte(maxHits),
		},
	}
}

// npcRemoveFrame builds an opcode-79 (SEND_NPC_COORDS) frame whose
// single positional record is REMOVE_NPC for the first NPC in the
// host's order list: 8 bits localCount=1, then bits 1 (needsUpdate),
// 1 (not-moving), 11 (REMOVE) packed MSB-first into the second byte.
func npcRemoveFrame() v235.Frame {
	return v235.Frame{
		Opcode:  v235.InNpcCoords,
		Payload: []byte{0x01, 0xF0},
	}
}

// drainTargetDied counts the TargetDied events currently buffered on sub.
func drainTargetDied(sub <-chan event.Event) []event.TargetDied {
	var out []event.TargetDied
	for {
		select {
		case ev := <-sub:
			if td, ok := ev.(event.TargetDied); ok {
				out = append(out, td)
			}
		default:
			return out
		}
	}
}

// The killing-blow damage update (curHits -> 0) arriving on the REAL
// opcode-104 path fires target_died exactly once and tears the
// engagement down so combat.engaged flips false.
func TestTargetDiedFiresViaUpdateNpcOpcode(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 5, X: 121, Y: 504, TypeID: 42})
	h.world.Npcs.SetHits(5, 0, 8, 8) // full health, alive
	h.lastAttackedNpcIndex.Store(5)
	h.beginCombatRoundTracking(5)
	sub := h.bus.Subscribe("target_died", 8)

	// A mid-fight hit must NOT fire the edge (but must count a round).
	h.handleFrame(updateNpcDamageFrame(5, 3, 5, 8))
	if got := drainTargetDied(sub); len(got) != 0 {
		t.Fatalf("mid-fight damage fired target_died: %#v", got)
	}
	h.combatMu.Lock()
	rounds := h.combatRounds
	h.combatMu.Unlock()
	if rounds != 1 {
		t.Errorf("combat rounds after one opcode-104 hit: got %d, want 1 (noteCombatRound bypassed?)", rounds)
	}

	// The killing blow.
	h.handleFrame(updateNpcDamageFrame(5, 5, 0, 8))
	got := drainTargetDied(sub)
	if len(got) != 1 {
		t.Fatalf("killing blow via opcode 104: got %d TargetDied, want 1", len(got))
	}
	if got[0].NpcIndex != 5 || got[0].TypeID != 42 {
		t.Errorf("TargetDied: got %+v, want NpcIndex=5 TypeID=42", got[0])
	}
	// Engagement must be torn down so wait_until predicates complete.
	if idx := h.lastAttackedNpcIndex.Load(); idx != 0 {
		t.Errorf("lastAttackedNpcIndex after kill: got %d, want 0", idx)
	}
	h.combatMu.Lock()
	roundTarget := h.combatRoundTarget
	h.combatMu.Unlock()
	if roundTarget != 0 {
		t.Errorf("combatRoundTarget after kill: got %d, want 0", roundTarget)
	}
	res := runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("combat.engaged after kill: got %v, want false", res.Value)
	}

	// A repeated 0-hits reading must not double-fire.
	h.handleFrame(updateNpcDamageFrame(5, 0, 0, 8))
	if again := drainTargetDied(sub); len(again) != 0 {
		t.Errorf("repeated 0-hits reading re-fired target_died: %#v", again)
	}
}

// On this server a killed NPC's final damage update is never sent
// (OpenRSC prunes it from localNpcs before updateNpcAppearances runs),
// so the kill surfaces ONLY as an opcode-79 REMOVE while engaged. That
// removal must fire target_died and flip the combat views.
func TestTargetDiedFiresOnRemovalWhileEngaged(t *testing.T) {
	h := newTestHost()
	// Adjacent to self (120,504): a melee kill.
	h.world.Npcs.Set(world.NpcRecord{Index: 7, X: 121, Y: 504, TypeID: 42})
	h.lastAttackedNpcIndex.Store(7)
	h.beginCombatRoundTracking(7)
	sub := h.bus.Subscribe("target_died", 8)

	h.handleFrame(npcRemoveFrame())

	got := drainTargetDied(sub)
	if len(got) != 1 {
		t.Fatalf("removal-while-engaged: got %d TargetDied, want 1", len(got))
	}
	if got[0].NpcIndex != 7 || got[0].TypeID != 42 {
		t.Errorf("TargetDied: got %+v, want NpcIndex=7 TypeID=42", got[0])
	}
	if _, ok := h.world.Npcs.Get(7); ok {
		t.Error("NPC record still in roster after REMOVE")
	}
	if idx := h.lastAttackedNpcIndex.Load(); idx != 0 {
		t.Errorf("lastAttackedNpcIndex after removal-kill: got %d, want 0", idx)
	}
	res := runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("combat.engaged after removal-kill: got %v, want false", res.Value)
	}
	res = runRoutine(t, h, `routine r() { return combat.target }`)
	if _, ok := res.Value.(interp.Null); !ok {
		t.Errorf("combat.target after removal-kill: got %v, want Null", res.Value)
	}
}

// A REMOVE for the engaged target recorded FAR away means WE left
// (retreat/teleport — the server pruned an out-of-range NPC): the
// engagement must end (combat.engaged flips false) but NO kill credit
// is granted.
func TestEngagedRemovalFarAwayIsDisengageNotKill(t *testing.T) {
	h := newTestHost()
	// 20 tiles east of self (120,504) — beyond engagedRemovalKillRadius.
	h.world.Npcs.Set(world.NpcRecord{Index: 9, X: 140, Y: 504, TypeID: 42})
	h.lastAttackedNpcIndex.Store(9)
	h.beginCombatRoundTracking(9)
	sub := h.bus.Subscribe("target_died", 8)

	h.handleFrame(npcRemoveFrame())

	if got := drainTargetDied(sub); len(got) != 0 {
		t.Fatalf("far-away removal granted kill credit: %#v", got)
	}
	// ... but the engagement is still over: views flip, tracking clears.
	if idx := h.lastAttackedNpcIndex.Load(); idx != 0 {
		t.Errorf("lastAttackedNpcIndex after disengage: got %d, want 0", idx)
	}
	h.combatMu.Lock()
	roundTarget := h.combatRoundTarget
	h.combatMu.Unlock()
	if roundTarget != 0 {
		t.Errorf("combatRoundTarget after disengage: got %d, want 0", roundTarget)
	}
	res := runRoutine(t, h, `routine r() { return combat.engaged }`)
	if b, ok := res.Value.(interp.Bool); !ok || bool(b) {
		t.Errorf("combat.engaged after disengage: got %v, want false", res.Value)
	}
}

// When the rare damage-to-zero edge DOES arrive before the REMOVE, the
// two signals must yield exactly one target_died (the damage edge clears
// the engagement, so the removal can no longer match it).
func TestTargetDiedNoDoubleFireDamageThenRemoval(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 4, X: 121, Y: 504, TypeID: 42})
	h.world.Npcs.SetHits(4, 0, 8, 8)
	h.lastAttackedNpcIndex.Store(4)
	h.beginCombatRoundTracking(4)
	sub := h.bus.Subscribe("target_died", 8)

	h.handleFrame(updateNpcDamageFrame(4, 8, 0, 8)) // killing blow seen
	h.handleFrame(npcRemoveFrame())                 // then the corpse prune

	got := drainTargetDied(sub)
	if len(got) != 1 {
		t.Fatalf("damage-to-zero + removal: got %d TargetDied, want exactly 1", len(got))
	}
}

// A removal of some OTHER NPC (not our target) must not end our
// engagement or grant a kill.
func TestUnrelatedNpcRemovalLeavesEngagementAlone(t *testing.T) {
	h := newTestHost()
	// Order matters: the bystander must be FIRST in the roster order so
	// the single positional REMOVE record in npcRemoveFrame names it.
	h.world.Npcs.Set(world.NpcRecord{Index: 3, X: 122, Y: 504, TypeID: 11}) // bystander
	h.world.Npcs.Set(world.NpcRecord{Index: 8, X: 121, Y: 504, TypeID: 42}) // our target
	h.lastAttackedNpcIndex.Store(8)
	h.beginCombatRoundTracking(8)
	sub := h.bus.Subscribe("target_died", 8)

	// localCount=2: REMOVE for slot 0 (bystander), no-update for slot 1.
	// Bits: 1 (needsUpdate), 1 (not-moving), 11 (REMOVE), 0 (slot 1 no
	// update) -> 0b11110 000 = 0xF0.
	h.handleFrame(v235.Frame{Opcode: v235.InNpcCoords, Payload: []byte{0x02, 0xF0}})

	if got := drainTargetDied(sub); len(got) != 0 {
		t.Fatalf("bystander removal fired target_died: %#v", got)
	}
	if idx := h.lastAttackedNpcIndex.Load(); idx != 8 {
		t.Errorf("lastAttackedNpcIndex after bystander removal: got %d, want 8", idx)
	}
	if _, ok := h.world.Npcs.Get(8); !ok {
		t.Error("our target was pruned by the bystander's REMOVE")
	}
}

// Incoming damage on our own player record (opcode 234's
// OtherPlayerDamage) advances the anti-kite round tally through
// applyCombatAware — the "we took a hit" half of noteCombatRound that
// the multi-event branch used to skip.
func TestCombatRoundsCountIncomingHitsViaApply(t *testing.T) {
	h := newTestHost()
	// Pin our own server index so the damage event reads as "self".
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(2, "test")
	h.world.Npcs.Set(world.NpcRecord{Index: 6, X: 121, Y: 504, TypeID: 42})
	h.lastAttackedNpcIndex.Store(6)
	h.beginCombatRoundTracking(6)

	h.applyCombatAware(event.OtherPlayerDamage{PlayerIndex: 2, Damage: 2, CurHits: 48, MaxHits: 50})

	h.combatMu.Lock()
	rounds := h.combatRounds
	h.combatMu.Unlock()
	if rounds != 1 {
		t.Errorf("combat rounds after incoming hit: got %d, want 1", rounds)
	}
}
