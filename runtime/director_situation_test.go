package runtime

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
	"github.com/gen0cide/westworld/worldmap"
)

// --- Sell affordance (TODO C-cluster; retro cause #6) ------------------------
//
// shop.sell was wired for the project's whole life and NEVER invoked; the hint
// these tests pin down is the prose that makes the affordance visible. The
// contract under test: fires for broke+sellable, silent when rich, silent when
// nothing a shop would buy, item list capped, coins never counted as stock.

// sellTestHost builds the standard test host with the REAL static item
// catalogue (live BasePrice/IsUntradable flags + the real Coins id) and the
// given inventory. Real defs keep the sellability rules honest — fixture defs
// could drift from the catalogue the live hint prices against.
//
// Static-catalogue items used below (id, name, base gp):
//
//	10 Coins(1) · 0 Iron Mace(63) · 1 Iron Short Sword(91)
//	2 Iron Kite Shield(238) · 3 Iron Square Shield(168) · 4 Wooden Shield(20)
//	23 Flour(2, UNTRADABLE) · 24 Amulet of GhostSpeak(35, UNTRADABLE)
func sellTestHost(slots ...world.InvSlot) *Host {
	h := newTestHost()
	h.facts = facts.LoadStaticCatalogs()
	h.world.Inventory.Replace(slots)
	return h
}

// TestSellAffordanceHintBrokeAndSellable proves the core affordance: a broke
// host (12gp) carrying shop-buyable items gets ONE compact line that names the
// items most-valuable-first, groups duplicate unstackable slots, prices the
// lot, teaches shop.sell — and that the line actually reaches the channel the
// model reads (mesad's actPrompt renders a fixed key set, so it rides "scene").
func TestSellAffordanceHintBrokeAndSellable(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 10, Amount: 12}, // 12gp — broke
		world.InvSlot{ItemID: 2, Amount: 1},   // Iron Kite Shield, 238gp
		world.InvSlot{ItemID: 0, Amount: 1},   // Iron Mace ×3, one per slot
		world.InvSlot{ItemID: 0, Amount: 1},
		world.InvSlot{ItemID: 0, Amount: 1},
	)
	d := quietDirector()
	hint := d.sellAffordanceHint(h, 120, 504)
	for _, want := range []string{
		"You hold sellable items",
		"Iron Kite Shield", // most valuable, named first
		"3 Iron Mace",      // three slots grouped into one stack
		"~427gp",           // 238 + 3×63; the coins themselves not counted
		"shop.sell(item)",
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	if strings.Contains(hint, "Coins") {
		t.Fatalf("coins must never be listed as sellable stock:\n%s", hint)
	}
	if strings.Index(hint, "Iron Kite Shield") > strings.Index(hint, "Iron Mace") {
		t.Fatalf("most valuable item should lead the list:\n%s", hint)
	}
	// End-to-end: the situation the planner sees must carry the line.
	sit := d.situation(h, Outcome{})
	if !strings.Contains(sit.Hints["scene"], "shop.sell(item)") {
		t.Fatalf("sell affordance did not reach the scene hint:\n%s", sit.Hints["scene"])
	}
}

// TestSellAffordanceHintAbsentWhenRich proves scarcity: at or above the coin
// floor the hint must not render at all — hints stay scarce to stay load-bearing.
func TestSellAffordanceHintAbsentWhenRich(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 10, Amount: 500}, // rich
		world.InvSlot{ItemID: 2, Amount: 1},    // sellable, but irrelevant now
	)
	d := quietDirector()
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("rich host must get no sell hint, got:\n%s", hint)
	}
	// Boundary: exactly the floor counts as "not broke".
	h.world.Inventory.Replace([]world.InvSlot{
		{ItemID: 10, Amount: sellHintCoinFloor},
		{ItemID: 2, Amount: 1},
	})
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("at the coin floor the hint must stay silent, got:\n%s", hint)
	}
	sit := d.situation(h, Outcome{})
	if strings.Contains(sit.Hints["scene"], "shop.sell") {
		t.Fatalf("scene must not carry a sell hint when rich:\n%s", sit.Hints["scene"])
	}
}

// TestSellAffordanceHintAbsentWhenNothingSellable proves the other scarcity
// edge: untradables (even with a positive BasePrice — the tradeability filter,
// not the price one), an empty inventory, and a catalogue-less host all stay
// silent.
func TestSellAffordanceHintAbsentWhenNothingSellable(t *testing.T) {
	d := quietDirector()
	h := sellTestHost(
		world.InvSlot{ItemID: 23, Amount: 1}, // Flour — untradable, BasePrice 2
		world.InvSlot{ItemID: 24, Amount: 1}, // Amulet of GhostSpeak — untradable
	)
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("untradables must not trigger the hint, got:\n%s", hint)
	}
	h.world.Inventory.Replace(nil)
	if hint := d.sellAffordanceHint(h, 120, 504); hint != "" {
		t.Fatalf("an empty inventory must not trigger the hint, got:\n%s", hint)
	}
	bare := newTestHost() // h.facts == nil: nothing can be priced
	if hint := d.sellAffordanceHint(bare, 120, 504); hint != "" {
		t.Fatalf("nil facts must not trigger the hint, got:\n%s", hint)
	}
}

// TestSellAffordanceHintItemCap proves the 3-item cap: with five sellable
// kinds only the three most valuable are named, the cheapest two fold into
// "and 2 more", and the gp figure still counts everything.
func TestSellAffordanceHintItemCap(t *testing.T) {
	h := sellTestHost(
		world.InvSlot{ItemID: 0, Amount: 1}, // Iron Mace 63 — folded
		world.InvSlot{ItemID: 1, Amount: 1}, // Iron Short Sword 91
		world.InvSlot{ItemID: 2, Amount: 1}, // Iron Kite Shield 238
		world.InvSlot{ItemID: 3, Amount: 1}, // Iron Square Shield 168
		world.InvSlot{ItemID: 4, Amount: 1}, // Wooden Shield 20 — folded
	)
	d := quietDirector()
	hint := d.sellAffordanceHint(h, 120, 504)
	for _, want := range []string{
		"Iron Kite Shield", "Iron Square Shield", "Iron Short Sword",
		"and 2 more",
		"~580gp", // 63+91+238+168+20 — the fold still counts toward the total
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint missing %q:\n%s", want, hint)
		}
	}
	for _, banned := range []string{"Iron Mace", "Wooden Shield"} {
		if strings.Contains(hint, banned) {
			t.Fatalf("cap exceeded — %q should fold into 'and 2 more':\n%s", banned, hint)
		}
	}
}

// TestSellAffordanceHintNearestShopRider proves the proximity rider uses only
// the embedded gazetteer already in RAM: (120,504) sits ~10 tiles from the
// (130,512) general shop so the rider names it; from (700,300) the nearest is
// hundreds of tiles away — no rider (a far shop is search_map's job).
func TestSellAffordanceHintNearestShopRider(t *testing.T) {
	h := sellTestHost(world.InvSlot{ItemID: 0, Amount: 1}) // broke, one mace
	d := quietDirector()
	if hint := d.sellAffordanceHint(h, 120, 504); !strings.Contains(hint, "Nearest: general shop ~10 tiles") {
		t.Fatalf("expected the nearest-shop rider at (120,504):\n%s", hint)
	}
	if hint := d.sellAffordanceHint(h, 700, 300); strings.Contains(hint, "Nearest:") {
		t.Fatalf("no rider expected far from any shop:\n%s", hint)
	}
}

// TestIgnoranceHintsRender pins the three self-aware-ignorance lines: fog
// coverage (weighted, with frontier directions), never-tried skills, and the
// aspiration portfolio — all riding the scene hint (fixed-key constraint).
func TestIgnoranceHintsRender(t *testing.T) {
	h := newTestHost()
	fake := &fakeFogOracle{dim: 96, dimY: 96, comps: map[[2]int]int32{
		{10, 10}: 0, {60, 10}: 0,
	}}
	h.fog.oracle = fake
	h.fogObservePosition(10, 10)

	d := NewMesaDirector(&fakeAskClient{healthy: true}, "Delores", "g", slog.Default())
	hint := d.explorationHint(h, 10, 10)
	if !strings.Contains(hint, "EXPLORATION") || !strings.Contains(hint, "%") {
		t.Fatalf("exploration hint missing coverage: %q", hint)
	}
	if !strings.Contains(hint, "Unknown lands") {
		t.Fatalf("exploration hint missing frontier: %q", hint)
	}

	sk := d.skillIgnoranceHint(h)
	if sk != "" && !strings.Contains(sk, "never tried") {
		t.Fatalf("skill hint malformed: %q", sk)
	}

	// Aspirations: dark without a portfolio, renders with one.
	if asp := d.aspirationHint(h); asp != "" {
		t.Fatalf("aspiration hint should be dark with no portfolio: %q", asp)
	}
}

// --- C-2: confidence-scaled routine commitment --------------------------------

// TestPlanConfidenceDerivation pins the threshold geometry: zero evidence is
// full confidence; the soft-fail nudge's territory (2 fails = exactly the
// floor) is NOT yet LOW; one more failure, a near-threshold stall, or mixed
// evidence IS; spin alone one-shots before reaching LOW; the value clamps at
// 0 under pile-ups and 1 at the top; and recorded goal PROGRESS (#33's
// positive-evidence term) offsets transient negatives without ever masking a
// hard pile-up.
func TestPlanConfidenceDerivation(t *testing.T) {
	cases := []struct {
		name               string
		fails, stall, spin int
		progress           float64
		want               float64
		low                bool
	}{
		{"clean", 0, 0, 0, 0, 1.0, false},
		{"soft-fail boundary stays the nudge's job", antiStuckSoftFails, 0, 0, 0, 0.5, false},
		{"three fails", 3, 0, 0, 0, 0.25, true},
		{"stall at threshold", 0, antiStuckWorldStall, 0, 0, 0.25, true},
		{"stall just before threshold", 0, antiStuckWorldStall - 1, 0, 0, 0.4, true},
		{"spin alone never trips LOW (one-shots at 3)", 0, 0, antiStuckSpinTurns, 0, 0.55, false},
		{"mixed evidence", 1, 2, 0, 0, 0.45, true},
		{"clamped at zero", 10, 10, 10, 0, 0, true},
		{"progress offsets transient fails", 3, 0, 0, 1.0, 0.5, false},
		{"partial progress softens but stays LOW", 3, 0, 0, 0.4, 0.35, true},
		{"progress clamps at the 1.0 ceiling", 0, 0, 0, 1.0, 1.0, false},
		{"progress cannot mask a hard pile-up", 10, 10, 10, 1.0, 0, true},
	}
	for _, tc := range cases {
		d := quietDirector()
		d.failStreak, d.worldStall, d.spinCount = tc.fails, tc.stall, tc.spin
		got := d.planConfidence(tc.progress)
		if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("%s: planConfidence = %.3f, want %.3f", tc.name, got, tc.want)
		}
		if (got < lowConfidenceFloor) != tc.low {
			t.Errorf("%s: LOW = %v at %.3f, want %v", tc.name, got < lowConfidenceFloor, got, tc.low)
		}
	}
}

// TestLowConfidenceChangesMarchingOrders proves the Act-visible contract: LOW
// confidence appends the one-small-reversible-step orders to the trigger
// (riding the existing channel — mesad renders the trigger verbatim), and a
// clean host gets no rider at all.
func TestLowConfidenceChangesMarchingOrders(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	d.failStreak = 3 // 0.25 — LOW; Outcome{} (no Label) leaves the preset intact
	sit := d.situation(h, Outcome{})
	if !strings.Contains(sit.Trigger, "LOW CONFIDENCE (0.25)") {
		t.Fatalf("LOW confidence must change the marching orders, got: %q", sit.Trigger)
	}
	if !strings.Contains(sit.Trigger, "ONE small, reversible step") {
		t.Fatalf("LOW orders must demand one small reversible step, got: %q", sit.Trigger)
	}

	h2 := newTestHost()
	d2 := quietDirector()
	sit2 := d2.situation(h2, Outcome{})
	if strings.Contains(sit2.Trigger, "LOW CONFIDENCE") {
		t.Fatalf("a clean host must get no confidence rider, got: %q", sit2.Trigger)
	}
}

// TestConfidenceCalibrationRecord proves calibration honesty: every live Act
// escalation publishes one parseable "confidence" decision record (the bus →
// decisions.jsonl seam), and a frozen analysis dry-run publishes none (M16).
func TestConfidenceCalibrationRecord(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	d.failStreak = 3
	ch := h.bus.Subscribe(event.AgentThought{}.Kind(), 16)
	d.situation(h, Outcome{})
	rec := drainConfidenceRecord(ch)
	if rec == nil {
		t.Fatal("no confidence calibration record published")
	}
	for _, want := range []string{"confidence=0.25", "fails=3", "stall=0", "spin=0", "progress=0.00", "last=none"} {
		if !strings.Contains(rec.Reasoning, want) {
			t.Fatalf("calibration record missing %q: %q", want, rec.Reasoning)
		}
	}

	// Freeze: an analysis dry-run must not pollute the calibration stream.
	h.analysis.activeBit.Store(true)
	d.situation(h, Outcome{})
	if rec := drainConfidenceRecord(ch); rec != nil {
		t.Fatalf("freeze must suppress the calibration record, got: %q", rec.Reasoning)
	}
}

// drainConfidenceRecord non-blockingly drains ch and returns the last
// Trigger=="confidence" thought, or nil.
func drainConfidenceRecord(ch <-chan event.Event) *event.AgentThought {
	var found *event.AgentThought
	for {
		select {
		case ev := <-ch:
			if t, ok := ev.(event.AgentThought); ok && t.Trigger == "confidence" {
				found = &t
			}
		default:
			return found
		}
	}
}

// --- C-3: stuck-breaker with distilled evidence --------------------------------

// stallTestRig wires the production shape: a MesaDirector wrapped by a
// HybridDirector, bound to the host through a conductor (hybridWrapper's read
// path), plus a blocked goal with forage breadcrumbs on its question node.
func stallTestRig(t *testing.T) (*Host, *MesaDirector, *HybridDirector) {
	t.Helper()
	h := newTestHost()
	d := quietDirector() // goal "do a thing"
	h.goalGraph.Upsert("do a thing", goalgraph.KindGoal, "do a thing", goalgraph.StatusActive)
	h.goalGraph.Upsert("where-to-buy:pickaxe", goalgraph.KindOpenQuestion, "Where do I buy a pickaxe?", goalgraph.StatusOpen)
	h.goalGraph.Link("do a thing", "where-to-buy:pickaxe", goalgraph.RelBlockedBy)
	h.goalGraph.Tag("where-to-buy:pickaxe", "source-tried:place:Lumbridge general-shop")
	h.goalGraph.Tag("where-to-buy:pickaxe", "source-spent:place:Falador general-shop")
	h.goalGraph.Tag("where-to-buy:pickaxe", "forage-exhausted")

	hd := NewHybridDirector(d, NewRoutineLibrary(nil), "do a thing", nil)
	c := NewConductor(h, ConductorOptions{Director: hd})
	h.configureAnalysis("", nil, c)
	return h, d, hd
}

// TestStallBreakerTriggerEvidence proves the distilled-context assembly: the
// STALLED trigger names the failed routines with their errors (newest first,
// repeats folded), the planner's recent reasonings, the blocking open question
// with its forage state, and the abandon contract — all riding the trigger
// string (no new prompt channel).
func TestStallBreakerTriggerEvidence(t *testing.T) {
	h, d, hd := stallTestRig(t)

	// Evidence rings: two identical failures fold; a newer distinct one leads.
	hd.harvestDecisions(h) // subscribe BEFORE publishing (production order: turn 1 precedes any Act thought)
	hd.noteFailure(failure(authoredIntent()))
	hd.noteFailure(failure(authoredIntent()))
	hd.noteFailure(Outcome{Intent: Intent{Label: "act:go_to"}, Kind: interp.ResultErrored, Err: failure(authoredIntent()).Err})
	h.bus.Publish(event.AgentThought{Turn: 7, Reasoning: "I will mine here", MoveKind: "WRITE_ROUTINE"})
	hd.harvestDecisions(h)

	d.worldStall = antiStuckWorldStall
	trig := d.stallBreakerTrigger(h)

	for _, want := range []string{
		"STALLED — 5 turns",
		"FAILED act:go_to — boom",
		"FAILED act:mine — boom (×2)",
		"you reasoned: I will mine here",
		"Where do I buy a pickaxe?",
		"already tried: Lumbridge general-shop",
		"confirmed nothing there: Falador general-shop",
		"EVERY known source exhausted",
		`goal_op:"abandoned"`,
	} {
		if !strings.Contains(trig, want) {
			t.Fatalf("breaker missing %q:\n%s", want, trig)
		}
	}
	// Newest-first: the go_to failure happened after the mine failures.
	if strings.Index(trig, "act:go_to") > strings.Index(trig, "act:mine") {
		t.Fatalf("failures must render newest first:\n%s", trig)
	}
	// No world oracle on a test host: no map verdicts, and no panic.
	if strings.Contains(trig, "map:") {
		t.Fatalf("no oracle ⇒ no map evidence:\n%s", trig)
	}
}

// TestStallBreakerFallsBackPlain proves the no-evidence degrade: a bare host
// (no wrapper bound, empty graph) still gets the original plain STALLED
// contract rather than an empty evidence list.
func TestStallBreakerFallsBackPlain(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	d.worldStall = antiStuckWorldStall
	trig := d.stallBreakerTrigger(h)
	if !strings.Contains(trig, "STALLED") || !strings.Contains(trig, "FUNDAMENTALLY different") {
		t.Fatalf("plain STALLED fallback malformed: %q", trig)
	}
	if strings.Contains(trig, "EVIDENCE") {
		t.Fatalf("no evidence ⇒ no evidence header: %q", trig)
	}
}

// TestStalledTriggerFiresThroughSituation pins the firing seam end-to-end: the
// old `trigger == ""` guard was dead (triggerFor never returns ""), so a
// stalled host only ever saw the bare STUCK text. Now worldStall ≥ threshold
// must produce the STALLED breaker — and keep it even when a hard fail-streak
// would otherwise overwrite it with the evidence-free BLOCKED text.
func TestStalledTriggerFiresThroughSituation(t *testing.T) {
	h, d, _ := stallTestRig(t)
	d.worldStall = antiStuckWorldStall
	sit := d.situation(h, Outcome{})
	if !strings.Contains(sit.Trigger, "STALLED") {
		t.Fatalf("worldStall ≥ threshold must fire STALLED, got: %q", sit.Trigger)
	}
	if strings.Contains(sit.Trigger, "STUCK — you have not moved") {
		t.Fatalf("STALLED must replace the bare STUCK text, got: %q", sit.Trigger)
	}

	// Hard fail-streak: the graph write still lands, but the breaker keeps the
	// trigger (its evidence already names the failures).
	h2, d2, _ := stallTestRig(t)
	d2.worldStall = antiStuckWorldStall
	d2.failStreak = antiStuckHardFails
	sit2 := d2.situation(h2, Outcome{})
	if !strings.Contains(sit2.Trigger, "STALLED") || strings.Contains(sit2.Trigger, "BLOCKED — your last") {
		t.Fatalf("breaker must outrank the bare BLOCKED text, got: %q", sit2.Trigger)
	}
}

// TestStallBreakerPayloadBounded proves the hard caps: oversized reasonings are
// clipped per line and the whole evidence list stops at the byte budget, so a
// stalled host costs hundreds of tokens, not thousands.
func TestStallBreakerPayloadBounded(t *testing.T) {
	h, d, hd := stallTestRig(t)
	hd.harvestDecisions(h)
	long := strings.Repeat("the same long thought over and over ", 30) // ~1100 chars
	for i := 0; i < stallDecTrailCap+2; i++ {
		h.bus.Publish(event.AgentThought{Turn: i + 1, Reasoning: fmt.Sprintf("%d %s", i, long)})
	}
	hd.harvestDecisions(h)
	for i := 0; i < stallFailLogCap+2; i++ {
		hd.noteFailure(Outcome{Intent: Intent{Label: fmt.Sprintf("act:try%d", i)}, Kind: interp.ResultErrored, Err: failure(authoredIntent()).Err})
	}
	d.worldStall = antiStuckWorldStall
	trig := d.stallBreakerTrigger(h)
	const slack = 700 // header + closing demand + bullets
	if len(trig) > stallEvBudget+slack {
		t.Fatalf("breaker payload unbounded: %d bytes", len(trig))
	}
	if !strings.Contains(trig, "…") {
		t.Fatalf("oversized reasonings must be clipped per line:\n%s", trig)
	}
}

// TestStallEvidenceRings pins the HybridDirector ring semantics: neutral
// outcomes (no prior action, budget-expired) record nothing; identical
// consecutive failures fold into a count; the ring keeps only the newest
// stallFailLogCap distinct failures; and the decision trail drops Turn-0
// sub-act records and empty reasonings, keeping the newest stallDecTrailCap.
func TestStallEvidenceRings(t *testing.T) {
	h := newTestHost()
	hd := NewHybridDirector(quietDirector(), NewRoutineLibrary(nil), "g", nil)

	hd.noteFailure(Outcome{})                       // no prior action
	hd.noteFailure(budgetExpired(authoredIntent())) // neutral, not a failure
	hd.noteFailure(success(authoredIntent()))       // success
	if len(hd.failLog) != 0 {
		t.Fatalf("neutral outcomes must not record failures: %+v", hd.failLog)
	}

	for i := 0; i < 3; i++ {
		hd.noteFailure(failure(authoredIntent()))
	}
	if len(hd.failLog) != 1 || hd.failLog[0].count != 3 {
		t.Fatalf("identical failures must fold: %+v", hd.failLog)
	}
	for i := 0; i < stallFailLogCap+1; i++ {
		hd.noteFailure(Outcome{Intent: Intent{Label: fmt.Sprintf("act:t%d", i)}, Kind: interp.ResultErrored, Err: failure(authoredIntent()).Err})
	}
	if len(hd.failLog) != stallFailLogCap {
		t.Fatalf("failLog must cap at %d, got %d", stallFailLogCap, len(hd.failLog))
	}
	if got := hd.recentFailures()[0].label; got != fmt.Sprintf("act:t%d", stallFailLogCap) {
		t.Fatalf("recentFailures must be newest-first, got %q first", got)
	}

	hd.harvestDecisions(h) // subscribe first
	h.bus.Publish(event.AgentThought{Turn: 0, Reasoning: "sub-act record (publishDecision)"})
	h.bus.Publish(event.AgentThought{Turn: 1, Reasoning: "   "})
	for i := 0; i < stallDecTrailCap+2; i++ {
		h.bus.Publish(event.AgentThought{Turn: i + 1, Reasoning: fmt.Sprintf("thought %d", i)})
	}
	hd.harvestDecisions(h)
	trail := hd.recentDecisions()
	if len(trail) != stallDecTrailCap {
		t.Fatalf("decTrail must cap at %d, got %d (%v)", stallDecTrailCap, len(trail), trail)
	}
	if trail[0] != fmt.Sprintf("thought %d", stallDecTrailCap+1) {
		t.Fatalf("recentDecisions must be newest-first, got %q first", trail[0])
	}
	for _, banned := range []string{"sub-act record", "   "} {
		for _, got := range trail {
			if got == banned {
				t.Fatalf("trail must filter %q", banned)
			}
		}
	}
}

// TestStallReachLineWording pins the verdict prose for each oracle outcome —
// the conclusion lines that turn "go somewhere else" into a grounded move (or
// an explicit cannot-from-here).
func TestStallReachLineWording(t *testing.T) {
	open := searchMapHit{label: "Lumbridge general-shop", x: 130, y: 512, dist: 12,
		info: worldmap.ReachInfo{Reach: worldmap.ReachOpen}}
	if l := stallReachLine(120, 504, "general-shop", open); !strings.Contains(l, "IS reachable") {
		t.Fatalf("open verdict malformed: %q", l)
	}
	gated := searchMapHit{label: "Al Kharid mine", x: 70, y: 580, dist: 40,
		info: worldmap.ReachInfo{Reach: worldmap.ReachGated, Gate: "Toll gate (south)", Needs: "10 coins", YouHave: 15}}
	if l := stallReachLine(120, 504, "mining-site", gated); !strings.Contains(l, "GATED by Toll gate") ||
		!strings.Contains(l, "needs 10 coins, you have 15 (you CAN pay)") {
		t.Fatalf("gated verdict malformed: %q", l)
	}
	walled := searchMapHit{label: "upstairs room", x: 120, y: 1448, dist: 2,
		info: worldmap.ReachInfo{Reach: worldmap.ReachBlocked}}
	if l := stallReachLine(120, 504, "bank", walled); !strings.Contains(l, "UNREACHABLE") ||
		!strings.Contains(l, "cannot be done from HERE") {
		t.Fatalf("no-route verdict malformed: %q", l)
	}
}
