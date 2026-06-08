package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	"github.com/gen0cide/westworld/pearl"
)

// TestLimbicHandleUpdatesAffect routes mood events and checks the vector moves.
func TestLimbicHandleUpdatesAffect(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.LevelUp{NewLevel: 5})
	_, conf, val := h.affect.Snapshot()
	if conf <= 0.5 || val <= 0 {
		t.Fatalf("level-up should lift confidence/valence: conf=%v val=%v", conf, val)
	}

	h2 := newTestHost()
	h2.limbicHandle(event.Death{})
	s, _, v := h2.affect.Snapshot()
	if s <= 0 || v >= 0 {
		t.Fatalf("death should raise stress / drop valence: stress=%v val=%v", s, v)
	}
}

// TestLimbicHandleUpdatesLedger routes social events and checks familiarity.
func TestLimbicHandleUpdatesLedger(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.ChatReceived{Speaker: "alex", Message: "hi"})
	h.limbicHandle(event.ChatReceived{Speaker: "alex", Message: "again"})
	if !h.ledger.Known("alex") {
		t.Fatal("chat should register the speaker in the ledger")
	}
	if r := h.ledger.Rel("alex"); r.Familiar != 2 {
		t.Fatalf("familiar=%d, want 2", r.Familiar)
	}
}

// TestLimbicHandleOtherPlayerChatLedger proves nearby players' PUBLIC chat — the
// index-based OtherPlayerChat (opcode 234), which is how ordinary conversation
// actually arrives — feeds the trust ledger after resolving the index to a name.
// Regression for a host never coming to "know" someone it only ever talked to.
func TestLimbicHandleOtherPlayerChatLedger(t *testing.T) {
	h := newTestHost() // Username "test"
	h.world.Players.SetName(7, "alex")
	h.world.Players.SetName(0, "test") // our own player record (self)

	h.limbicHandle(event.OtherPlayerChat{PlayerIndex: 7, MessageText: "hi"})
	h.limbicHandle(event.OtherPlayerChat{PlayerIndex: 7, MessageText: "again"})
	if !h.ledger.Known("alex") {
		t.Fatal("public player chat should register the speaker in the ledger")
	}
	if r := h.ledger.Rel("alex"); r.Familiar != 2 {
		t.Fatalf("familiar=%d, want 2", r.Familiar)
	}

	// Our own chat echoed back as local chat must NOT create a self-relationship.
	h.limbicHandle(event.OtherPlayerChat{PlayerIndex: 0, MessageText: "me talking"})
	if h.ledger.Known("test") {
		t.Fatal("own echoed chat must not register a self-relationship")
	}
}

// TestLimbicHandleTradeDuelLedger proves attributed trade/duel interactions feed
// the ledger: a trade engagement is a positive trust signal, a duel is
// familiarity-only (adversarial), both attributed by the name on the wire.
func TestLimbicHandleTradeDuelLedger(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.TradeRequestReceived{FromPlayerName: "alex"})
	h.limbicHandle(event.TradeConfirmShown{OpponentName: "alex"})
	h.limbicHandle(event.DuelRequestReceived{FromPlayerName: "rival"})

	if !h.ledger.Known("alex") || !h.ledger.Known("rival") {
		t.Fatal("trade/duel should register their counterparties in the ledger")
	}
	if ra := h.ledger.Rel("alex"); ra.Familiar != 2 || ra.Trust <= 0 {
		t.Fatalf("alex = %+v, want familiar=2 trust>0 (trade engagement is positive)", ra)
	}
	if rr := h.ledger.Rel("rival"); rr.Familiar != 1 || rr.Trust != 0 {
		t.Fatalf("rival = %+v, want familiar=1 trust=0 (duel = familiarity only)", rr)
	}
}

// TestDuelRequestStillMetOnly proves the duel=Met floor (invariant d) is NOT
// regressed: a bare duel offer only bumps familiarity — every relationship axis
// (trust/affinity/grievance) stays at neutral.
func TestDuelRequestStillMetOnly(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.DuelRequestReceived{FromPlayerName: "rival"})
	r := h.ledger.Rel("rival")
	if r.Familiar != 1 {
		t.Fatalf("duel request should bump familiarity to 1, got %d", r.Familiar)
	}
	if r.Trust != 0 || r.Affinity != 0 || r.Grievance != 0 {
		t.Fatalf("duel request must be Met-only, got trust=%v aff=%v gri=%v", r.Trust, r.Affinity, r.Grievance)
	}
}

// TestDuelClosedAffinityNotTrust proves a COMPLETED consensual duel is a sport:
// familiarity + AFFINITY + a "sparring-partner" tag, but NEVER trust or grievance
// (even though it ends a fight). Attribution flows from the live duel record.
func TestDuelClosedAffinityNotTrust(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(8, "spar")
	h.limbicHandle(event.DuelRequestReceived{FromPlayerName: "spar"}) // Met + capture name
	h.world.Duel.BeginRequest(8, "spar")
	h.world.Duel.MarkOpened(8, "spar")
	h.limbicHandle(event.DuelOpened{OtherPlayerIndex: 8})
	h.world.Duel.MarkClosed(true) // world.Apply ran first; record kept (WithName readable)
	h.limbicHandle(event.DuelClosed{Completed: true})

	r := h.ledger.Rel("spar")
	if r.Affinity <= 0 {
		t.Fatalf("completed duel should add affinity, got %v", r.Affinity)
	}
	if !contains(r.Tags, "sparring-partner") {
		t.Fatalf("completed duel should tag sparring-partner, tags=%v", r.Tags)
	}
	if r.Trust != 0 {
		t.Fatalf("a duel must NEVER move trust, got %v", r.Trust)
	}
	if r.Grievance != 0 {
		t.Fatalf("a duel must NEVER add grievance, got %v", r.Grievance)
	}
}

// TestDeathInDuelNoGrievance proves a death DURING a consensual duel is a sport,
// not a wrong: no grievance is recorded (the wilderness-context gate).
func TestDeathInDuelNoGrievance(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(8, "spar")
	h.world.Players.SetAppearanceCombat(8, 50, 1) // skulled — would otherwise attribute
	h.world.Players.SetEngagement(8, 0, -1, h.world.Players.SelfIndex())
	h.world.Duel.BeginRequest(8, "spar")
	h.world.Duel.MarkOpened(8, "spar") // duel is now ACTIVE
	if !h.inDuel() {
		t.Fatal("precondition: duel should be active")
	}
	h.limbicHandle(event.Death{})
	if r := h.ledger.Rel("spar"); r.Grievance != 0 {
		t.Fatalf("a death in a duel must not add grievance, got %v", r.Grievance)
	}
}

// TestDeathUnattributableRecordsNothing proves the anti-misattribution constraint:
// a death with no engaged target / no skull context writes NOTHING to the ledger.
func TestDeathUnattributableRecordsNothing(t *testing.T) {
	h := newTestHost()
	before := len(h.ledger.All())
	h.limbicHandle(event.Death{})
	if got := len(h.ledger.All()); got != before {
		t.Fatalf("unattributable death must not touch the ledger; rows %d -> %d", before, got)
	}
}

// TestDeathPKAttributed proves a wilderness PK kill (engaged skulled target +
// Death, not in a duel) records a large grievance + the "ganked-me" tag.
func TestDeathPKAttributed(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")            // self
	h.world.Players.SetSelfName("test")           // pin self index
	h.world.Players.SetName(5, "ganker")          // the aggressor
	h.world.Players.SetAppearanceCombat(5, 80, 1) // skulled => PK context
	// Our own record names us via Self() at index 0; engage us against 5.
	h.world.Players.SetEngagement(0, 0, -1, 5) // Self.EngagedPlayerIndex = 5

	h.limbicHandle(event.Death{})
	r := h.ledger.Rel("ganker")
	if r.Grievance <= 0 {
		t.Fatalf("PK death should add grievance, got %v", r.Grievance)
	}
	if !contains(r.Tags, "ganked-me") {
		t.Fatalf("PK death should tag ganked-me, tags=%v", r.Tags)
	}
}

// TestProjectilePKGrievance proves a hostile ranged/magic projectile aimed at us
// (named caster, no duel) accrues grievance + an "attacked-me" tag.
func TestProjectilePKGrievance(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(6, "caster")
	self := h.world.Players.SelfIndex()

	h.limbicHandle(event.OtherPlayerProjectile{
		CasterIndex: 6, VictimPlayerIndex: self, VictimIsNpc: false,
	})
	r := h.ledger.Rel("caster")
	if r.Grievance <= 0 {
		t.Fatalf("projectile PK should add grievance, got %v", r.Grievance)
	}
	if !contains(r.Tags, "attacked-me") {
		t.Fatalf("projectile PK should tag attacked-me, tags=%v", r.Tags)
	}
	if r.Trust != 0 {
		t.Fatalf("a hostile projectile must not move trust, got %v", r.Trust)
	}
}

// TestTradeOutcomeSplitsAxes proves a confirmed trade splits across TRUST and
// AFFINITY (a fair completed exchange is both honest and warm).
func TestTradeOutcomeSplitsAxes(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.TradeConfirmShown{OpponentName: "merchant"})
	r := h.ledger.Rel("merchant")
	if r.Trust <= 0 {
		t.Fatalf("a confirmed trade should raise trust, got %v", r.Trust)
	}
	if r.Affinity <= 0 {
		t.Fatalf("a smooth trade should raise affinity, got %v", r.Affinity)
	}
	if r.Grievance != 0 {
		t.Fatalf("a fair trade must not add grievance, got %v", r.Grievance)
	}
}

// TestRepeatedItemLossAccruesGrievance proves grievance COMPOUNDS: repeated PK
// hits from the same party accumulate, eventually crossing the grudge threshold.
func TestRepeatedItemLossAccruesGrievance(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(6, "raider")
	self := h.world.Players.SelfIndex()

	prev := 0.0
	for i := 0; i < 8; i++ {
		h.limbicHandle(event.OtherPlayerProjectile{CasterIndex: 6, VictimPlayerIndex: self})
		g := h.ledger.Rel("raider").Grievance
		if g < prev {
			t.Fatalf("grievance must be monotone; hit %d dropped %v -> %v", i, prev, g)
		}
		prev = g
	}
	if prev < 0.5 {
		t.Fatalf("sustained fire should accrue past the grudge threshold, got %v", prev)
	}
}

// TestMesaWireAxesRoundTrip proves the two new axes survive the mesa converters
// (ledger snapshot -> wire -> ledger import) as RAW sums.
func TestMesaWireAxesRoundTrip(t *testing.T) {
	l := newTestHost().ledger
	l.ObserveAffinity("x", 2.5)
	l.ObserveGrievance("x", 1.5)

	wire := ledgerToRelationships(l.Export())
	back := relationshipsToLedger(wire)
	var got *struct{ a, g float64 }
	for _, e := range back {
		if e.Name == "x" {
			got = &struct{ a, g float64 }{e.AffinitySum, e.GrievanceSum}
		}
	}
	if got == nil {
		t.Fatal("row x lost in wire round-trip")
	}
	if got.a != 2.5 || got.g != 1.5 {
		t.Fatalf("wire round-trip axes = aff %v gri %v, want 2.5/1.5", got.a, got.g)
	}
}

// TestPearlFactsReadsAffect proves the affect vector flows into pearl Facts.
func TestPearlFactsReadsAffect(t *testing.T) {
	h := newTestHost()
	h.affect.OnDeath() // spikes stress
	f := h.pearlFacts(pearl.EventCtx{Action: "decide"})
	if f.Affect.Stress <= 0 {
		t.Fatalf("pearlFacts.Affect.Stress = %v, want > 0 after death", f.Affect.Stress)
	}
}

// TestPearlStressPredicateFiresFromLimbic ties it together: an affect spike from
// the limbic path makes a stress-gated pearl rule fire through the decide seam.
func TestPearlStressPredicateFiresFromLimbic(t *testing.T) {
	h := newTestHost()
	h.Strategist = nil // a miss would surface NOT_IMPLEMENTED
	h.Pearl = pearl.New(pearl.Table{Rules: []pearl.Rule{{
		ID:         "panic_when_stressed",
		When:       pearl.StressAbove(0.3),
		Then:       pearl.Effect{Kind: pearl.EffectBias, Bias: map[string]float64{"flee": 1.0}},
		Confidence: 0.9,
	}}}, 0)
	h.affect.OnDeath() // stress now well above 0.3

	res := runRoutine(t, h, `runtime "1.0"
routine r() { return decide(["fight", "flee"]).val }`)
	if res.Kind != interp.ResultReturned || res.Value.Display() != "flee" {
		t.Fatalf("expected pearl to bias 'flee' from limbic stress; got kind=%v val=%v", res.Kind, res.Value.Display())
	}
}

// TestLimbicLedgerPersists proves the trust ledger survives a restart: flush
// through the memory layer, build a fresh host sharing the same store, reload.
func TestLimbicLedgerPersists(t *testing.T) {
	store := hostkv.NewMemory()
	scratch := hostkv.NewScratch(16)

	h := newTestHost()
	h.Memory = memory.New(memory.Options{Scratch: scratch, Local: store})
	for range 6 {
		h.ledger.Observe("alex", true, 1)
	}
	h.ledger.Met("alex")
	h.flushLimbic(context.Background())

	// A fresh host (new ledger) sharing the same durable store reloads it.
	h2 := newTestHost()
	h2.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	if h2.ledger.Known("alex") {
		t.Fatal("precondition: fresh ledger should not know alex before load")
	}
	h2.loadLimbic(context.Background())
	r := h2.ledger.Rel("alex")
	if r.Familiar != 1 || r.Trust <= 0 {
		t.Fatalf("after reload alex = %+v, want familiar=1 trust>0", r)
	}
}

// TestRelationWithUsesLedger proves relation_with reports the learned trust grade.
func TestRelationWithUsesLedger(t *testing.T) {
	h := newTestHost()
	for range 12 {
		h.ledger.Observe("alex", true, 1)
	}
	h.ledger.Met("alex")
	res := runRoutine(t, h, `runtime "1.0"
routine r() { return relation_with("alex").val }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind=%v err=%v", res.Kind, res.Err)
	}
	if got := res.Value.Display(); got != "trusted" {
		t.Fatalf("relation_with(alex) = %q, want trusted", got)
	}
}
