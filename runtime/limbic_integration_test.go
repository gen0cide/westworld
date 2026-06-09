package runtime

import (
	"context"
	"sync"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/world"
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

// TestDeathDuringDuelFightNoGrievance is the C2 regression: DuelClosed{Completed:true}
// fires at fight START, so the duel-UI phase is already terminal (inDuel()==false)
// while the fight is live. A skulled wilderness-duel death landing DURING that window
// must NOT be mis-recorded as a PK gank — the sporting-fight window covers it. The
// pre-fix gate (only !inDuel()) would have attributed a 3.0 grievance + "ganked-me".
func TestDeathDuringDuelFightNoGrievance(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(8, "spar")
	h.world.Players.SetAppearanceCombat(8, 60, 1) // skulled — would otherwise attribute a gank
	h.world.Players.SetEngagement(0, 0, -1, 8)    // Self.EngagedPlayerIndex = 8 (would feed pkKiller)
	// Fight just started: world.Apply flipped the record terminal BEFORE the event.
	h.world.Duel.BeginRequest(8, "spar")
	h.world.Duel.MarkOpened(8, "spar")
	h.world.Duel.MarkClosed(true) // Phase="completed" ⇒ inDuel() is now FALSE
	if h.inDuel() {
		t.Fatal("precondition: the duel UI phase should already be terminal at fight start")
	}
	h.limbicHandle(event.DuelClosed{Completed: true}) // opens the sporting-fight window
	if !h.inSportingFight() {
		t.Fatal("DuelClosed{Completed:true} should open the sporting-fight window")
	}

	h.limbicHandle(event.Death{}) // death DURING the fight
	if r := h.ledger.Rel("spar"); r.Grievance != 0 {
		t.Fatalf("a death during a duel fight must not add grievance, got %v", r.Grievance)
	}
	if r := h.ledger.Rel("spar"); contains(r.Tags, "ganked-me") {
		t.Fatalf("a duel-fight death must not tag ganked-me, tags=%v", r.Tags)
	}

	// After respawn (the full StatsSnapshot) the window closes, so a LATER genuine
	// gank is attributed again — the window does not permanently mute grievance.
	h.limbicHandle(event.StatsSnapshot{})
	if h.inSportingFight() {
		t.Fatal("respawn StatsSnapshot should close the sporting-fight window")
	}
	h.limbicHandle(event.Death{})
	if r := h.ledger.Rel("spar"); r.Grievance <= 0 {
		t.Fatalf("a post-respawn PK death should attribute grievance again, got %v", r.Grievance)
	}
}

// TestPvEDeathDoesNotBlameSameIndexPlayer is the M10 regression: combatRoundTarget
// holds an NPC scene index on a PvE death; the old fallback looked that index up in
// the separate PLAYER store and blamed whatever skulled player happened to occupy
// the same number. With the fallback dropped (player-scoped attribution only), a
// pure-PvE death with no engaged/last-attacked PLAYER attributes NOTHING.
func TestPvEDeathDoesNotBlameSameIndexPlayer(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	// A skulled bystander player occupying numeric index 7 (the SAME number as the
	// NPC we were fighting). On a PvE death the old code would blame this player.
	h.world.Players.SetName(7, "bystander")
	h.world.Players.SetAppearanceCombat(7, 80, 1) // skulled
	// We were fighting NPC index 7 (a bear), not a player.
	h.world.Players.SetEngagement(0, 0, 7, -1) // EngagedNpcIndex=7, no engaged player
	h.combatRoundTarget = 7                    // the (now-irrelevant) NPC-index fallback

	h.limbicHandle(event.Death{})
	if r := h.ledger.Rel("bystander"); r.Grievance != 0 || contains(r.Tags, "ganked-me") {
		t.Fatalf("a PvE death must not blame a player at the same numeric index: %+v", r)
	}
}

// TestOwnDeathWritesBlockedByAndEnabler is the H19 regression: the host's OWN death
// while pursuing an active goal must become GENERATIVE — write
// <active-goal> --blocked_by--> <cause> and spawn an enabling sub-goal — instead of
// touching only affect + grievance. (Cause attributed from the PK killer here.)
func TestOwnDeathWritesBlockedByAndEnabler(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(5, "ganker")
	h.world.Players.SetAppearanceCombat(5, 80, 1) // skulled => attributable cause
	h.world.Players.SetEngagement(0, 0, -1, 5)

	// An active goal the host is pursuing when it dies.
	h.goalGraph.Upsert("mine-coal", goalgraph.KindGoal, "Mine coal in the wilderness", goalgraph.StatusActive)

	h.limbicHandle(event.Death{})

	g, ok := h.goalGraph.Get("mine-coal")
	if !ok || g.Status != goalgraph.StatusBlocked {
		t.Fatalf("death should block the active goal, got %+v ok=%v", g, ok)
	}
	cid := "cause:ganker"
	if len(h.goalGraph.Out("mine-coal", goalgraph.RelBlockedBy)) != 1 {
		t.Fatalf("death should write a blocked_by edge to the cause, edges=%v", h.goalGraph.Edges())
	}
	if _, ok := h.goalGraph.Get(cid); !ok {
		t.Fatalf("death should create the cause node %q", cid)
	}
	sid := "survive:ganker"
	sub, ok := h.goalGraph.Get(sid)
	if !ok || sub.Kind != goalgraph.KindSubgoal {
		t.Fatalf("death should spawn an enabling sub-goal %q, got %+v ok=%v", sid, sub, ok)
	}
	if len(h.goalGraph.Out(sid, goalgraph.RelEnables)) != 1 {
		t.Fatalf("the sub-goal should enable the blocked goal")
	}

	// Idempotent: a second identical death reinforces the SAME nodes/edges (no growth).
	before := len(h.goalGraph.Nodes())
	h.limbicHandle(event.Death{})
	if after := len(h.goalGraph.Nodes()); after != before {
		t.Fatalf("repeated death should not grow the graph: %d -> %d", before, after)
	}
}

// TestOwnDeathWithNoActiveGoalNoWrite proves the death→goalgraph writer is a no-op
// when there is no active goal (better silent than a placeholder edge — H19).
func TestOwnDeathWithNoActiveGoalNoWrite(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	h.world.Players.SetName(5, "ganker")
	h.world.Players.SetAppearanceCombat(5, 80, 1)
	h.world.Players.SetEngagement(0, 0, -1, 5)
	// No active goal node in the graph.
	h.limbicHandle(event.Death{})
	if !h.goalGraph.Empty() {
		t.Fatalf("a death with no active goal must not write to the goal graph: %v", h.goalGraph.Nodes())
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

// TestCompletedTradeAccumulatesValueTraded proves a COMPLETED trade folds the
// magnitude of goods exchanged (total item quantity across both offers) into the
// counterparty's monotone value_traded volume, attributed to the confirm-screen
// opponent. A cancelled trade contributes NOTHING. Regression for the latent trap
// where value_traded was always 0 (limbic.Entry had no field + no accumulator).
func TestCompletedTradeAccumulatesValueTraded(t *testing.T) {
	t.Run("completed accumulates the exchange magnitude", func(t *testing.T) {
		h := newTestHost()
		h.world.Trade.BeginRequest(9, "merchant")
		h.world.Trade.MarkOpened(9, "merchant")
		h.world.Trade.SetMyOffer([]world.TradeItem{{ItemID: 10, Amount: 3}})
		h.world.Trade.SetTheirOffer([]world.TradeItem{{ItemID: 20, Amount: 5}, {ItemID: 21, Amount: 2}})
		// The confirm screen captures the named counterparty (the attribution net).
		h.limbicHandle(event.TradeConfirmShown{OpponentName: "merchant"})
		h.world.Trade.MarkClosed(true) // world.Apply ran first → Phase="completed"
		h.limbicHandle(event.TradeClosed{Completed: true})

		if got := h.ledger.Export(); len(got) == 0 {
			t.Fatal("no relationship row after a completed trade")
		}
		if r := relValueTraded(h, "merchant"); r != 10 { // 3 + 5 + 2
			t.Fatalf("value_traded=%v, want 10 (total quantity exchanged)", r)
		}
		// A SECOND completed trade compounds the volume (monotone total). Reset both
		// sides of the record so the second exchange's magnitude is unambiguous.
		h.world.Trade.SetMyOffer([]world.TradeItem{{ItemID: 10, Amount: 4}})
		h.world.Trade.SetTheirOffer(nil)
		h.limbicHandle(event.TradeClosed{Completed: true})
		if r := relValueTraded(h, "merchant"); r != 14 { // 10 + 4 (second trade adds 4)
			t.Fatalf("value_traded after second trade=%v, want 14 (compounded)", r)
		}
	})

	t.Run("cancelled trade contributes nothing", func(t *testing.T) {
		h := newTestHost()
		h.world.Trade.BeginRequest(9, "merchant")
		h.world.Trade.MarkOpened(9, "merchant")
		h.world.Trade.SetMyOffer([]world.TradeItem{{ItemID: 10, Amount: 3}})
		h.limbicHandle(event.TradeConfirmShown{OpponentName: "merchant"})
		h.world.Trade.MarkClosed(false) // cancelled
		h.limbicHandle(event.TradeClosed{Completed: false})
		if r := relValueTraded(h, "merchant"); r != 0 {
			t.Fatalf("a cancelled trade must not add value_traded, got %v", r)
		}
	})
}

// TestValueTradedWireRoundTrip proves the value_traded accumulator survives the
// mesa converters in BOTH directions (latent-trap close: ledgerToRelationships
// no longer always writes 0).
func TestValueTradedWireRoundTrip(t *testing.T) {
	l := newTestHost().ledger
	l.ObserveValueTraded("x", 240)

	wire := ledgerToRelationships(l.Export())
	var onWire float64
	for _, r := range wire {
		if r.Name == "x" {
			onWire = r.ValueTraded
		}
	}
	if onWire != 240 {
		t.Fatalf("ledgerToRelationships dropped value_traded: got %v, want 240", onWire)
	}
	back := relationshipsToLedger(wire)
	var got float64
	for _, e := range back {
		if e.Name == "x" {
			got = e.ValueTraded
		}
	}
	if got != 240 {
		t.Fatalf("value_traded lost on the reverse converter: got %v, want 240", got)
	}
}

// relValueTraded reads the raw value_traded accumulator off the ledger snapshot for
// name (Rel doesn't surface it — it's a wire/storage field, not a decision input yet).
func relValueTraded(h *Host, name string) float64 {
	for _, e := range h.ledger.Export() {
		if e.Name == name {
			return e.ValueTraded
		}
	}
	return 0
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

// TestLimbicConcurrencyRaceFree exercises the three host-side concurrency fixes
// under -race: C1 (SetAffectBaseline mutates affect IN PLACE rather than swapping
// the pointer the limbic/pearl/director readers hold), H11 (the combat-tracking
// fields are mutex-guarded across the frame-pump + limbic readers), and H12
// (keywordLadder is an atomic.Pointer so the bootstrap write doesn't tear the
// reactive reader's range). Pre-fix this fails with WARNING: DATA RACE; post-fix it
// passes. Deterministic-result-free: it asserts no race, not a value.
func TestLimbicConcurrencyRaceFree(t *testing.T) {
	h := newTestHost()
	h.world.Players.SetName(0, "test")
	h.world.Players.SetSelfName("test")
	const iters = 500
	var wg sync.WaitGroup

	// C1: re-baseline the affect vector while readers snapshot it.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.SetAffectBaseline(0.1, 0.6, 0.2)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_, _, _ = h.affect.Snapshot()
			h.affect.OnXPGain(10)
		}
	}()

	// H11: write the combat counters (frame pump) while a reader reads them (limbic).
	h.combatRoundTarget = 7
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.noteCombatRound(event.NpcDamage{NpcIndex: 7})
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = h.inSportingFight() // reads duelFightUntil under combatMu
			h.limbicHandle(event.DuelClosed{Completed: true})
		}
	}()

	// H12: install the keyword ladder while the reactive reader ranges it.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.SetKeywordLadder([]mesaclient.KeywordRung{{Keyword: "pickaxe", Tier: "TRADE_INTEREST"}})
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = h.triggerHit("npc", "merchant", "i sell a pickaxe", 0.9)
		}
	}()

	// value_traded: completed trades fold volume into the ledger (ObserveValueTraded)
	// while a reader snapshots the ledger — the new accumulator must be race-free too.
	h.world.Trade.BeginRequest(9, "merchant")
	h.world.Trade.MarkOpened(9, "merchant")
	h.world.Trade.SetMyOffer([]world.TradeItem{{ItemID: 10, Amount: 2}})
	h.world.Trade.MarkClosed(true)
	h.lastTradeOpponent = "merchant"
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.limbicHandle(event.TradeClosed{Completed: true})
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = h.ledger.Export()
		}
	}()

	wg.Wait()
}
