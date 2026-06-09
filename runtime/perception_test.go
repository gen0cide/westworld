package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

// perceptionFacts builds a minimal Facts carrying the named item + npc defs the
// perception writers resolve against, so a shop catalogue / kill / pickup turns
// into a NAMED knowledge row instead of an unresolved item#N (which is skipped).
func perceptionFacts(items []*facts.ItemDef, npcs []*facts.NpcDef) *facts.Facts {
	im := make(map[int]*facts.ItemDef, len(items))
	for _, d := range items {
		im[d.ID] = d
	}
	nm := make(map[int]*facts.NpcDef, len(npcs))
	for _, d := range npcs {
		nm[d.ID] = d
	}
	return &facts.Facts{ItemDefs: im, NpcDefs: nm}
}

// --- shop ---

// TestPerceiveShopWritesKnowledge proves the flagship writer: an opened shop
// records "sells <item>" beliefs keyed on the best-effort keeper, with high
// confidence for in-stock items (the "where to buy a pickaxe" signal).
func TestPerceiveShopWritesKnowledge(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)
	h.perceive.lastNpc = "Nurmof"

	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 5}}})

	if !h.knowledge.Known("Nurmof") {
		t.Fatal("shop open should record knowledge keyed on the keeper")
	}
	f := h.knowledge.Get("Nurmof")
	if len(f.Beliefs) == 0 {
		t.Fatal("expected at least one belief about the shop")
	}
	top := f.Beliefs[0]
	if !strings.Contains(top.Claim, "sells") || !strings.Contains(top.Claim, "rune pickaxe") {
		t.Fatalf("top belief = %q, want it to mention 'sells rune pickaxe'", top.Claim)
	}
	if top.Confidence() <= 0.5 {
		t.Fatalf("in-stock belief confidence = %v, want > 0.5", top.Confidence())
	}
	hasShopTag := false
	for _, tg := range f.Tags {
		if tg == "shop" {
			hasShopTag = true
		}
	}
	if !hasShopTag {
		t.Fatalf("shop subject should be tagged 'shop'; tags=%v", f.Tags)
	}
}

// TestPerceiveShopStockDedup proves an identical re-sent catalogue (RSC's stock-
// update path) is a no-op, but a changed stock IS recorded.
func TestPerceiveShopStockDedup(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)
	h.perceive.lastNpc = "Nurmof"

	open := event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 5}}}
	h.perceptionHandle(open)
	first := h.knowledge.Get("Nurmof").Beliefs[0]
	a1 := first.Alpha

	// Identical re-send: gated, so alpha must not move.
	h.perceptionHandle(open)
	a2 := h.knowledge.Get("Nurmof").Beliefs[0].Alpha
	if a2 != a1 {
		t.Fatalf("identical re-send should be deduped: alpha %v -> %v", a1, a2)
	}

	// Changed stock: recorded (alpha grows).
	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 9}}})
	a3 := h.knowledge.Get("Nurmof").Beliefs[0].Alpha
	if a3 <= a2 {
		t.Fatalf("changed stock should record more evidence: alpha %v -> %v", a2, a3)
	}
}

// TestPerceiveShopAbsence proves an out-of-stock item is recorded as evidence
// AGAINST the "sells X" claim, so its confidence stays low.
func TestPerceiveShopAbsence(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)
	h.perceive.lastNpc = "Nurmof"

	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 0}}})
	f := h.knowledge.Get("Nurmof")
	if len(f.Beliefs) == 0 {
		t.Fatal("expected a belief even for an out-of-stock item")
	}
	if c := f.Beliefs[0].Confidence(); c >= 0.5 {
		t.Fatalf("out-of-stock 'sells' confidence = %v, want low (< 0.5)", c)
	}
}

// TestPerceiveShopFallbackSubject proves that with no known keeper the shop is
// attributed to a generic descriptor (general store / shop), not dropped.
func TestPerceiveShopFallbackSubject(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)

	h.perceptionHandle(event.ShopOpened{IsGeneral: true, Items: []event.ShopItem{{ItemID: 156, Stock: 5}}})
	if !h.knowledge.Known("general store") {
		t.Fatalf("no-keeper general store should fall back to 'general store'; rows=%v", h.knowledge.All())
	}
}

// TestPerceiveShopSkipsUnresolvedItems proves unresolved item ids (item#N) never
// create a knowledge belief — junk ids must not pollute the ledger.
func TestPerceiveShopSkipsUnresolvedItems(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts(nil, nil) // nothing resolves
	h.perceive.lastNpc = "Nurmof"

	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 999, Stock: 5}}})
	f := h.knowledge.Get("Nurmof")
	for _, b := range f.Beliefs {
		if strings.Contains(b.Claim, "item#") {
			t.Fatalf("unresolved id leaked into the ledger as a claim: %q", b.Claim)
		}
	}
	if len(f.Beliefs) != 0 {
		t.Fatalf("unresolved-only shop should write no 'sells' beliefs; got %v", f.Beliefs)
	}
}

// --- area ---

// TestPerceiveAreaFamiliarity proves arriving AT a named place bumps location
// familiarity once per session (re-arrival is deduped).
func TestPerceiveAreaFamiliarity(t *testing.T) {
	h := newTestHost()
	h.facts = &facts.Facts{} // real embedded gazetteer
	g := h.facts.Gazetteer()
	if len(g.Places) == 0 {
		t.Skip("no gazetteer places embedded")
	}
	p := g.Places[0]
	resolved, _, _ := g.NearestPlace(p.X, p.Y)

	h.perceptionHandle(event.OwnPositionUpdate{X: p.X, Y: p.Y})
	if fam := h.knowledge.Get(resolved.Name).Familiar; fam < 1 {
		t.Fatalf("arriving at %q should bump familiarity; got %d", resolved.Name, fam)
	}

	// Re-arrival at the same place is deduped this session: no further bump.
	h.perceptionHandle(event.OwnPositionUpdate{X: p.X, Y: p.Y})
	if fam := h.knowledge.Get(resolved.Name).Familiar; fam != 1 {
		t.Fatalf("re-arrival should be session-deduped; familiarity = %d, want 1", fam)
	}
}

// --- kill / item ---

// TestPerceiveKillWritesKnowledge proves a target death records a "killable by
// me" belief about the NPC type.
func TestPerceiveKillWritesKnowledge(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts(nil, []*facts.NpcDef{{ID: 3, Name: "Giant rat"}})

	h.perceptionHandle(event.TargetDied{NpcIndex: 1, TypeID: 3})
	f := h.knowledge.Get("Giant rat")
	found := false
	for _, b := range f.Beliefs {
		if b.Claim == "killable by me" && b.Alpha > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("kill should record 'killable by me' with alpha>0; beliefs=%v", f.Beliefs)
	}
}

// TestPerceiveItemGained proves a pickup records an "obtainable by me" belief.
func TestPerceiveItemGained(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 20, Name: "bronze pickaxe"}}, nil)

	h.perceptionHandle(event.ItemGained{ItemID: 20, Count: 1})
	f := h.knowledge.Get("bronze pickaxe")
	found := false
	for _, b := range f.Beliefs {
		if b.Claim == "obtainable by me" && b.Alpha > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("pickup should record 'obtainable by me' with alpha>0; beliefs=%v", f.Beliefs)
	}
}

// --- dialog attribution ---

// TestPerceiveDialogSetsLastNpc proves dialog text attributes the conversation
// to the nearest named NPC, so a shop opened mid-talk is keyed on the keeper.
func TestPerceiveDialogSetsLastNpc(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts(
		[]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}},
		[]*facts.NpcDef{{ID: 90, Name: "Nurmof"}},
	)
	h.world.Npcs.Set(world.NpcRecord{Index: 1, TypeID: 90, X: 120, Y: 504})

	h.perceptionHandle(event.NpcDialogText{Text: "Want to buy a pickaxe?"})
	if h.perceive.lastNpc != "Nurmof" {
		t.Fatalf("dialog should set lastNpc to the named keeper; got %q", h.perceive.lastNpc)
	}

	// A shop opened now is attributed to that keeper.
	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 5}}})
	if !h.knowledge.Known("Nurmof") {
		t.Fatal("shop should be keyed on the keeper resolved from dialog")
	}
}

// TestNpcNearbyResolvesByIndexNotType proves a movement NpcNearby (TypeID==0,
// IsNew=false) credits the NPC that actually MOVED — resolved by scene index
// through the world model — and never the phantom NPC type 0 ("Unicorn", id 0).
func TestNpcNearbyResolvesByIndexNotType(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts(nil, []*facts.NpcDef{
		{ID: 0, Name: "Unicorn"},
		{ID: 3, Name: "Giant rat"},
	})
	// A Giant rat (scene index 7, type 3) is already in the world model.
	h.world.Npcs.Set(world.NpcRecord{Index: 7, TypeID: 3, X: 121, Y: 504})

	// A movement update carries TypeID==0 (only the new-NPC section carries a
	// real type). The OLD code did npcNameByType(0)=="Unicorn"; the fix resolves
	// by Index instead.
	h.perceptionHandle(event.NpcNearby{Index: 7, DX: 1, IsNew: false, TypeID: 0})

	if h.knowledge.Known("Unicorn") {
		t.Fatal("movement update must not credit the phantom NPC type 0 (Unicorn)")
	}
	if !h.knowledge.Known("Giant rat") {
		t.Fatalf("movement update should credit the NPC at the moved index; rows=%v", h.knowledge.All())
	}
}

// TestNpcNearbyUnresolvedIndexIsNoOp proves a movement update for an index the
// world model does not (yet) know about names no NPC — no phantom credit.
func TestNpcNearbyUnresolvedIndexIsNoOp(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts(nil, []*facts.NpcDef{{ID: 0, Name: "Unicorn"}})

	h.perceptionHandle(event.NpcNearby{Index: 42, DX: 1, IsNew: false, TypeID: 0})
	if len(h.knowledge.All()) != 0 {
		t.Fatalf("unresolved-index movement must credit nothing; rows=%v", h.knowledge.All())
	}
}

// TestNearestNamedNpcIsDeterministicNearest proves nearestNamedNpc returns the
// TRUE nearest named NPC by Chebyshev distance (not a random map element), so
// shop / dialog attribution is deterministic across calls.
func TestNearestNamedNpcIsDeterministicNearest(t *testing.T) {
	h := newTestHost() // self at (120, 504)
	h.facts = perceptionFacts(nil, []*facts.NpcDef{
		{ID: 1, Name: "Far Guard"},
		{ID: 2, Name: "Near Keeper"},
	})
	// Far Guard at distance 20; Near Keeper at distance 1 (adjacent).
	h.world.Npcs.Set(world.NpcRecord{Index: 5, TypeID: 1, X: 140, Y: 504})
	h.world.Npcs.Set(world.NpcRecord{Index: 9, TypeID: 2, X: 121, Y: 504})

	// Many calls must all return the nearest, regardless of map iteration order.
	for i := 0; i < 50; i++ {
		if got := h.nearestNamedNpc(); got != "Near Keeper" {
			t.Fatalf("call %d: nearestNamedNpc = %q, want the nearest 'Near Keeper'", i, got)
		}
	}
}

// TestNearestNamedNpcTieBreakIsStable proves an equidistant tie breaks on the
// lowest scene index, so the result is stable across calls.
func TestNearestNamedNpcTieBreakIsStable(t *testing.T) {
	h := newTestHost() // self at (120, 504)
	h.facts = perceptionFacts(nil, []*facts.NpcDef{
		{ID: 1, Name: "Index Three"},
		{ID: 2, Name: "Index Seven"},
	})
	// Both at Chebyshev distance 2 from self; index 3 must win deterministically.
	h.world.Npcs.Set(world.NpcRecord{Index: 7, TypeID: 2, X: 122, Y: 504})
	h.world.Npcs.Set(world.NpcRecord{Index: 3, TypeID: 1, X: 118, Y: 504})

	for i := 0; i < 50; i++ {
		if got := h.nearestNamedNpc(); got != "Index Three" {
			t.Fatalf("call %d: equidistant tie should break to lowest index ('Index Three'); got %q", i, got)
		}
	}
}

// --- analysis freeze ---

// TestPerceptionFreezesUnderAnalysis proves the operator override freezes ALL
// perception writes (no knowledge rows created while in analysis mode).
func TestPerceptionFreezesUnderAnalysis(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)
	h.perceive.lastNpc = "Nurmof"
	h.EnterAnalysis()

	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 5}}})
	if len(h.knowledge.All()) != 0 {
		t.Fatalf("perception must not write under analysis mode; rows=%v", h.knowledge.All())
	}
}

// TestShopClosedResetsSlateUnderAnalysis proves the ShopClosed state-RESET runs
// even during an analysis freeze: a ShopClosed arriving while frozen must clear
// the per-item dedup slate so a post-freeze ShopOpened with the same subject is
// not suppressed (the L4 stale-slate bug).
func TestShopClosedResetsSlateUnderAnalysis(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{{ID: 156, Name: "rune pickaxe"}}, nil)
	h.perceive.lastNpc = "Nurmof"

	// Shop open BEFORE the freeze populates the dedup slate.
	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{{ItemID: 156, Stock: 5}}})
	if h.perceive.shopStock == nil || h.perceive.shopSubject != "Nurmof" {
		t.Fatalf("pre-freeze shop open should set the slate; stock=%v subject=%q", h.perceive.shopStock, h.perceive.shopSubject)
	}

	// Enter analysis, then the shop closes DURING the freeze.
	h.EnterAnalysis()
	h.perceptionHandle(event.ShopClosed{})
	if h.perceive.shopStock != nil || h.perceive.shopSubject != "" {
		t.Fatalf("ShopClosed must clear the slate even under analysis; stock=%v subject=%q", h.perceive.shopStock, h.perceive.shopSubject)
	}
}

// --- observation firehose (sparse) ---

// TestPerceiveShopEmitsOneObservation proves a shop open emits EXACTLY ONE
// observation (sparse — not per item) of kind "transaction", and that an item
// pickup emits ZERO (stays local, would otherwise flood mesa).
func TestPerceiveShopEmitsOneObservation(t *testing.T) {
	h := newTestHost()
	h.facts = perceptionFacts([]*facts.ItemDef{
		{ID: 156, Name: "rune pickaxe"},
		{ID: 157, Name: "steel pickaxe"},
		{ID: 158, Name: "iron pickaxe"},
	}, nil)
	h.perceive.lastNpc = "Nurmof"
	sink := &fakeSink{healthy: true, obs: make(chan *mesaclient.Observation, 8)}
	h.SetMesaMemory(sink)

	h.perceptionHandle(event.ShopOpened{Items: []event.ShopItem{
		{ItemID: 156, Stock: 5}, {ItemID: 157, Stock: 3}, {ItemID: 158, Stock: 7},
	}})

	var first *mesaclient.Observation
	select {
	case first = <-sink.obs:
	case <-time.After(2 * time.Second):
		t.Fatal("shop open should emit one observation")
	}
	if first.Kind != "transaction" {
		t.Fatalf("observation kind = %q, want transaction", first.Kind)
	}
	// No second observation: the catalogue is one sparse signal, not per item.
	select {
	case o := <-sink.obs:
		t.Fatalf("shop open should emit ONE observation, got a second: %+v", o)
	case <-time.After(200 * time.Millisecond):
	}

	// An item pickup emits ZERO observations (frequent/low-salience; stays local).
	h.perceptionHandle(event.ItemGained{ItemID: 156, Count: 1})
	select {
	case o := <-sink.obs:
		t.Fatalf("item pickup should not emit an observation, got %+v", o)
	case <-time.After(200 * time.Millisecond):
	}
}

// --- dialog / chat / server capture (the reactive-tier substrate) ---

// TestPerceiveDialogCapturesSignals proves the firehose CAPTURE of directed
// language: NPC dialog, player chat, and server messages each emit a tagged
// observation (kind routes by source, subject = speaker) — the substrate the
// Phase-2 reactive tier consumes. Empty text and the host's own echoed chat are
// skipped (this is capture only — no parsing; that's mesa's job).
func TestPerceiveDialogCapturesSignals(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	h.world.Players.SetName(2, "Alex")
	h.world.Players.SetName(0, "Delores") // self
	sink := &fakeSink{healthy: true, obs: make(chan *mesaclient.Observation, 16)}
	h.SetMesaMemory(sink)

	h.perceive.lastNpc = "Hans"
	h.perceptionHandle(event.NpcDialogText{Text: "You need a key for that door."})
	h.perceptionHandle(event.OtherPlayerChat{PlayerIndex: 2, MessageText: "want to trade?"})
	h.perceptionHandle(event.SystemMessage{Message: "Welcome to the tutorial."})

	got := map[string]string{} // kind -> subject
	for i := 0; i < 3; i++ {
		select {
		case o := <-sink.obs:
			got[o.Kind] = o.Subject
		case <-time.After(2 * time.Second):
			t.Fatalf("expected 3 dialog observations, got %d: %v", len(got), got)
		}
	}
	if got["npc_dialog"] != "Hans" {
		t.Errorf("npc_dialog subject = %q, want Hans", got["npc_dialog"])
	}
	if got["player_chat"] != "Alex" {
		t.Errorf("player_chat subject = %q, want Alex", got["player_chat"])
	}
	if _, ok := got["server_msg"]; !ok {
		t.Errorf("expected a server_msg observation; got %v", got)
	}

	// Empty text, the host's own echoed public chat (index 0 == self), and its own
	// named echo must all be skipped — no further observations.
	h.perceptionHandle(event.NpcDialogText{Text: "   "})
	h.perceptionHandle(event.OtherPlayerChat{PlayerIndex: 0, MessageText: "my own echo"})
	h.perceptionHandle(event.ChatReceived{Speaker: "Delores", Message: "also me"})
	select {
	case o := <-sink.obs:
		t.Fatalf("empty text / self echo must not emit; got %+v", o)
	case <-time.After(200 * time.Millisecond):
	}
}
