package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// fakeExtractClient is a mesa Client that returns a canned extraction so the
// host-side reactive path (window → RPC → writeback → interrupt) can be driven
// deterministically offline. It embeds StubClient to satisfy the full interface
// and overrides only ExtractDialog + Healthy.
type fakeExtractClient struct {
	mesaclient.StubClient
	result *mesaclient.DialogExtraction
	calls  int
	gotWin []string
}

func (f *fakeExtractClient) Healthy() bool { return true }

func (f *fakeExtractClient) ExtractDialog(_ context.Context, _, _, _ string, window []string, _, _ string, _ []string) (*mesaclient.DialogExtraction, error) {
	f.calls++
	f.gotWin = window
	if f.result != nil {
		return f.result, nil
	}
	return &mesaclient.DialogExtraction{Intent: mesaclient.DialogIntent{Kind: "statement", Urgency: "low"}}, nil
}

// --- window: bounded buffer, prune, LRU evict --------------------------------

func TestReactivePushLineCapsAndPrunes(t *testing.T) {
	r := newReactiveState()
	now := time.Now()
	// Push more than the buffer cap; only the newest reactiveBufLines survive.
	for i := 0; i < reactiveBufLines+5; i++ {
		r.pushLine("alice", dialogLine{at: now.Add(time.Duration(i) * time.Second), role: "player", speaker: "Alice", text: "line"})
	}
	snap := r.snapshot("alice")
	if len(snap) != reactiveBufLines {
		t.Fatalf("buffer not capped: got %d lines, want %d", len(snap), reactiveBufLines)
	}

	// A line older than the lookback is pruned on the next push.
	r2 := newReactiveState()
	r2.pushLine("bob", dialogLine{at: now.Add(-2 * reactiveLookback), role: "player", speaker: "Bob", text: "stale"})
	r2.pushLine("bob", dialogLine{at: now, role: "player", speaker: "Bob", text: "fresh"})
	snap = r2.snapshot("bob")
	if len(snap) != 1 || snap[0] != "Bob: fresh" {
		t.Fatalf("stale line not pruned: %v", snap)
	}
}

func TestReactiveSpeakerCapEvictsNonLatched(t *testing.T) {
	r := newReactiveState()
	now := time.Now()
	// Fill to the speaker cap; the first speaker is latched (must NOT be evicted).
	r.pushLine("keep", dialogLine{at: now, role: "player", speaker: "Keep", text: "hi"})
	r.tryLatch("keep", now)
	for i := 0; i < reactiveMaxSpeakers+3; i++ {
		k := time.Now().Format("150405.000000000") + string(rune('a'+i%26)) + string(rune('0'+i))
		r.pushLine(k, dialogLine{at: now.Add(time.Duration(i) * time.Millisecond), role: "player", speaker: k, text: "x"})
	}
	r.mu.Lock()
	_, keepPresent := r.windows["keep"]
	total := len(r.windows)
	r.mu.Unlock()
	if !keepPresent {
		t.Fatal("a latched speaker was evicted under the speaker cap — must be protected")
	}
	if total > reactiveMaxSpeakers {
		t.Fatalf("speaker windows not bounded: have %d, cap %d", total, reactiveMaxSpeakers)
	}
}

// --- latch: cap, refresh, decay ----------------------------------------------

func TestReactiveLatchCapAndDecay(t *testing.T) {
	r := newReactiveState()
	now := time.Now()
	// Latch up to the concurrent cap.
	for i := 0; i < reactiveMaxLatches; i++ {
		k := "spk" + string(rune('0'+i))
		r.pushLine(k, dialogLine{at: now, role: "player", speaker: k, text: "hi"})
		if !r.tryLatch(k, now) {
			t.Fatalf("speaker %s should latch (under cap)", k)
		}
	}
	// One more must be refused (at the cap).
	r.pushLine("overflow", dialogLine{at: now, role: "player", speaker: "overflow", text: "hi"})
	if r.tryLatch("overflow", now) {
		t.Fatal("latch cap not enforced: a speaker latched beyond reactiveMaxLatches")
	}
	if got := r.latchCount(); got != reactiveMaxLatches {
		t.Fatalf("latch count = %d, want %d", got, reactiveMaxLatches)
	}

	// gcLatches after the TTL decays every latch.
	r.gcLatches(now.Add(reactiveLatchTTL + time.Second))
	if got := r.latchCount(); got != 0 {
		t.Fatalf("latches did not decay: count = %d, want 0", got)
	}

	// A fresh latch can be acquired again, and a new line refreshes its TTL.
	later := now.Add(reactiveLatchTTL + 2*time.Second)
	r.pushLine("spk0", dialogLine{at: later, role: "player", speaker: "spk0", text: "back"})
	if !r.tryLatch("spk0", later) {
		t.Fatal("should be able to re-latch after decay freed a slot")
	}
	// A subsequent line on the latched speaker returns latched=true (bypass) and
	// refreshes the TTL.
	if !r.pushLine("spk0", dialogLine{at: later.Add(time.Second), role: "player", speaker: "spk0", text: "more"}) {
		t.Fatal("a line on a latched speaker must report latched=true (bypass the gate)")
	}
}

// tryLatch must CREATE the window for a cold target (H6): the ASK path latches a
// shop NPC it has never heard, and the window must exist (and be latched) so the
// host's own question can be fanned in before the answer arrives.
func TestTryLatchCreatesColdWindow(t *testing.T) {
	r := newReactiveState()
	now := time.Now()
	if !r.tryLatch("nurmof", now) {
		t.Fatal("tryLatch should create + latch a cold (never-heard) target window")
	}
	r.mu.Lock()
	w, ok := r.windows["nurmof"]
	r.mu.Unlock()
	if !ok || w == nil || !w.latched {
		t.Fatal("a cold tryLatch must leave a latched window in place")
	}
	if r.latchCount() != 1 {
		t.Fatalf("latch count = %d, want 1", r.latchCount())
	}
	// unlatch undoes it (the ASK send-failure path) without deleting the window.
	r.unlatch("nurmof")
	if r.latchCount() != 0 {
		t.Fatalf("unlatch should drop the latch: count = %d", r.latchCount())
	}
}

// A directed self-line (directSelfTo) fans into ONLY the armed target's window,
// never the broadcast — even when other conversations are latched (L7).
func TestDirectedSelfLineFansIntoTargetOnly(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	now := time.Now()
	// Two latched conversations: the ASK target (A) and an unrelated one (B).
	h.reactive.pushLine("a", dialogLine{at: now, role: "npc", speaker: "A", text: "hello"})
	h.reactive.tryLatch("a", now)
	h.reactive.pushLine("b", dialogLine{at: now, role: "player", speaker: "B", text: "what's up"})
	h.reactive.tryLatch("b", now)

	// Arm the directed routing at A, then emit a self-line (as host.Say would).
	h.reactive.directSelfTo("a")
	h.reactiveObserveSelf("where do you sell pickaxes?")

	if snap := h.reactive.snapshot("a"); len(snap) != 2 || snap[1] != "Me: where do you sell pickaxes?" {
		t.Fatalf("directed self-line did not land in the target window: %v", snap)
	}
	if snap := h.reactive.snapshot("b"); len(snap) != 1 {
		t.Fatalf("directed self-line leaked into an unrelated latched window (L7): %v", snap)
	}
}

// TestSocialReplyFansToAddresseeOnly is the L7 socialReflex regression: a DIRECTED
// reply (the host answering a specific player) must pair only with that player's
// conversation window, not broadcast into every latched conversation. socialReflex
// now arms directSelfTo(normalizeSpeaker(from)) right before host.Say, whose
// reactiveObserveSelf consumes the hint — this exercises that exact success and
// send-failure sequence against the reactive primitives Say funnels through.
func TestSocialReplyFansToAddresseeOnly(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	now := time.Now()

	// Two live conversations: the player she is replying to (mixed-case "Arthur",
	// so normalizeSpeaker matters) and an unrelated latched bystander.
	const from = "Arthur"
	h.reactive.pushLine(normalizeSpeaker(from), dialogLine{at: now, role: "player", speaker: from, text: "what are you doing?"})
	h.reactive.tryLatch(normalizeSpeaker(from), now)
	h.reactive.pushLine("bystander", dialogLine{at: now, role: "player", speaker: "Bystander", text: "hey all"})
	h.reactive.tryLatch("bystander", now)

	// SUCCESS path: socialReflex arms the addressee, then host.Say (on a successful
	// send) runs reactiveObserveSelf, which consumes the hint and fans the reply.
	h.reactive.directSelfTo(normalizeSpeaker(from))
	h.reactiveObserveSelf("I'm mining some copper") // what a successful Say does

	if snap := h.reactive.snapshot(normalizeSpeaker(from)); len(snap) != 2 || snap[1] != "Me: I'm mining some copper" {
		t.Fatalf("the reply did not pair into the addressee's window: %v", snap)
	}
	if snap := h.reactive.snapshot("bystander"); len(snap) != 1 {
		t.Fatalf("the directed reply leaked into an unrelated latched window (L7): %v", snap)
	}

	// FAILURE path: socialReflex arms the addressee, the send fails, so it calls
	// clearDirectSelf — the unconsumed hint must NOT leak into a LATER public line,
	// which then broadcasts to every latched window as a genuine public utterance.
	h.reactive.directSelfTo(normalizeSpeaker(from))
	h.reactive.clearDirectSelf() // the host.Say send-failure branch
	h.reactiveObserveSelf("anyone selling logs?")

	if snap := h.reactive.snapshot("bystander"); len(snap) != 2 || snap[1] != "Me: anyone selling logs?" {
		t.Fatalf("after a cleared hint, a public line should broadcast to all latched windows: %v", snap)
	}
	if snap := h.reactive.snapshot(normalizeSpeaker(from)); len(snap) != 3 {
		t.Fatalf("the public broadcast should also reach the addressee window: %v", snap)
	}
}

// An UNDIRECTED self-line still broadcasts into every latched window (a public
// utterance heard by everyone in local-chat range).
func TestUndirectedSelfLineBroadcasts(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	now := time.Now()
	h.reactive.pushLine("a", dialogLine{at: now, role: "npc", speaker: "A", text: "hi"})
	h.reactive.tryLatch("a", now)
	h.reactive.pushLine("b", dialogLine{at: now, role: "player", speaker: "B", text: "yo"})
	h.reactive.tryLatch("b", now)

	h.reactiveObserveSelf("anyone selling logs?") // no directSelfTo armed → broadcast
	if snap := h.reactive.snapshot("a"); len(snap) != 2 {
		t.Fatalf("undirected self-line should reach latched window A: %v", snap)
	}
	if snap := h.reactive.snapshot("b"); len(snap) != 2 {
		t.Fatalf("undirected self-line should reach latched window B: %v", snap)
	}
}

// --- trigger detector: fires on ladder/goal/directed, not on ambient ---------

func TestTriggerHitOnLadderGoalDirected(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	h.SetKeywordLadder([]mesaclient.KeywordRung{{Keyword: "pickaxe", Tier: "TRADE_INTEREST"}})
	h.goalGraph.Upsert("q1", "open_question", "where to buy a pickaxe", "open")

	// Plain ambient chatter: no ladder word, no goal-touch, low salience → no trigger.
	if h.triggerHit("player", "Stranger", "nice weather today huh", 0.5) {
		t.Fatal("ambient chatter must NOT trigger")
	}
	// Ladder hit.
	if !h.triggerHit("player", "Smith", "i can sell you a pickaxe cheap", 0.5) {
		t.Fatal("a ladder keyword must trigger")
	}
	// Goal-touch: the line shares a significant word ("pickaxe") with the open
	// question "where to buy a pickaxe". (Use a low salience + a non-ladder
	// phrasing so ONLY the goal-touch axis can be responsible — but note "pickaxe"
	// is also a ladder word here, so verify goal-touch in isolation below.)
	hNoLadder := newTestHost()
	hNoLadder.goalGraph.Upsert("q1", "open_question", "where to buy a pickaxe", "open")
	if !hNoLadder.triggerHit("player", "Smith", "the pickaxe you want is up north", 0.3) {
		t.Fatal("a goal/open-question touch (shared significant word) must trigger")
	}
	// A line that shares NO significant word with the goal/questions stays ambient.
	if hNoLadder.triggerHit("player", "Smith", "the weather is lovely outside", 0.3) {
		t.Fatal("a line with no goal-touch and no ladder hit must NOT trigger")
	}
	// Directed: a salient PM/whisper.
	if !h.triggerHit("player", "Friend", "hey come here", reactiveDirectedSal) {
		t.Fatal("a salient directed line must trigger")
	}
	// Directed: my name appears in the line.
	if !h.triggerHit("player", "Friend", "watch out Delores theres a scammer", 0.5) {
		t.Fatal("a line naming the host must trigger")
	}
	// A low-salience server line below the threshold does not trigger by itself.
	if h.triggerHit("server", "server", "Welcome to RuneScape.", 0.1) {
		t.Fatal("a low-salience server line must NOT trigger on its own")
	}
}

// --- self-line fan-in: only into latched windows -----------------------------

func TestReactiveSelfLineFansIntoLatchedOnly(t *testing.T) {
	r := newReactiveState()
	now := time.Now()
	r.pushLine("latched", dialogLine{at: now, role: "player", speaker: "Latched", text: "where is the bank?"})
	r.tryLatch("latched", now)
	r.pushLine("ambient", dialogLine{at: now, role: "player", speaker: "Ambient", text: "lol"})

	r.appendToLatched(dialogLine{at: now.Add(time.Second), role: "self", speaker: "Me", text: "go north"})

	if snap := r.snapshot("latched"); len(snap) != 2 || snap[1] != "Me: go north" {
		t.Fatalf("self line did not fan into the latched window: %v", snap)
	}
	if snap := r.snapshot("ambient"); len(snap) != 1 {
		t.Fatalf("self line leaked into a non-latched window: %v", snap)
	}
}

// --- writeback: provenance from role, idempotent -----------------------------

func TestWritebackClaimsProvenanceAndIdempotency(t *testing.T) {
	h := newTestHost()

	// Player-told → Hearsay.
	h.writebackClaims("Smith", "player", []mesaclient.DialogClaim{
		{Subject: "Nurmof", Kind: "npc", Claim: "sells pickaxes", Confidence: 0.7, Provenance: "system" /* the LLM's self-report is IGNORED */},
	})
	f := h.knowledge.Get("Nurmof")
	if len(f.Beliefs) != 1 || f.Beliefs[0].Provenance != knowledge.ProvHearsay {
		t.Fatalf("player-told claim should be Hearsay regardless of LLM self-report: %+v", f.Beliefs)
	}

	// NPC/server → System (game-authoritative).
	h.writebackClaims("server", "server", []mesaclient.DialogClaim{
		{Subject: "tutorial", Kind: "quest", Claim: "talk to the guide", Confidence: 0},
	})
	if f := h.knowledge.Get("tutorial"); len(f.Beliefs) != 1 || f.Beliefs[0].Provenance != knowledge.ProvSystem {
		t.Fatalf("server claim should be System: %+v", f.Beliefs)
	}

	// Idempotent: an identical player claim reinforces (one belief, not two).
	a0 := h.knowledge.Get("Nurmof").Beliefs[0].Alpha
	h.writebackClaims("Smith", "player", []mesaclient.DialogClaim{
		{Subject: "Nurmof", Kind: "npc", Claim: "sells pickaxes", Confidence: 0.7},
	})
	f = h.knowledge.Get("Nurmof")
	if len(f.Beliefs) != 1 {
		t.Fatalf("a restated claim duplicated instead of reinforcing: %d beliefs", len(f.Beliefs))
	}
	if f.Beliefs[0].Alpha <= a0 {
		t.Fatalf("a restated claim did not reinforce: alpha %v -> %v", a0, f.Beliefs[0].Alpha)
	}
}

// --- integration: a latched exchange drives writeback + an urgent interrupt --

func TestReactiveSpawnExtractWritebackAndInterrupt(t *testing.T) {
	h := newTestHost()
	h.opts.Username = "Delores"
	fake := &fakeExtractClient{result: &mesaclient.DialogExtraction{
		Claims: []mesaclient.DialogClaim{
			{Subject: "Bob's Axes", Kind: "shop", Claim: "sells iron pickaxe", Confidence: 0.8},
		},
		Intent: mesaclient.DialogIntent{Kind: "warning", Urgency: "high", Gist: "a scammer is nearby"},
	}}

	// Wire a conductor (so the interrupt has somewhere to land) + the fake client.
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})
	h.configureAnalysis("Operator", fake, c)

	// Seed the speaker window so snapshot() yields an exchange, then extract.
	r := h.reactive
	r.pushLine(normalizeSpeaker("Bob"), dialogLine{at: time.Now(), role: "player", speaker: "Bob", text: "buy at Bob's Axes, but watch out for the scammer"})
	h.spawnExtract("Bob", "player")

	// Extraction runs in a goroutine; wait for the writeback to land (<10s budget).
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if h.knowledge.Known("Bob's Axes") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !h.knowledge.Known("Bob's Axes") {
		t.Fatal("reactive extraction did not write the claim into the knowledge ledger")
	}
	f := h.knowledge.Get("Bob's Axes")
	if len(f.Beliefs) != 1 || f.Beliefs[0].Provenance != knowledge.ProvHearsay {
		t.Fatalf("writeback belief wrong: %+v", f.Beliefs)
	}

	// The high-urgency intent must have raised a reactive conductor interrupt.
	select {
	case req := <-c.interrupts:
		if req.tier != "reactive" {
			t.Fatalf("interrupt tier = %q, want reactive", req.tier)
		}
		if req.abort {
			t.Fatal("reactive interrupt must be abort:false (park-and-resume)")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("an urgent intent did not raise a conductor interrupt")
	}

	if fake.calls != 1 {
		t.Fatalf("ExtractDialog calls = %d, want 1", fake.calls)
	}
}

// A normal-urgency intent writes claims but must NOT interrupt.
func TestReactiveNormalUrgencyDoesNotInterrupt(t *testing.T) {
	h := newTestHost()
	fake := &fakeExtractClient{result: &mesaclient.DialogExtraction{
		Claims: []mesaclient.DialogClaim{{Subject: "thing", Kind: "concept", Claim: "exists", Confidence: 0.6}},
		Intent: mesaclient.DialogIntent{Kind: "statement", Urgency: "normal"},
	}}
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})
	h.configureAnalysis("Operator", fake, c)

	r := h.reactive
	r.pushLine(normalizeSpeaker("Tom"), dialogLine{at: time.Now(), role: "player", speaker: "Tom", text: "by the way that thing exists"})
	h.spawnExtract("Tom", "player")

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if h.knowledge.Known("thing") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !h.knowledge.Known("thing") {
		t.Fatal("normal-urgency claim should still be written to the ledger")
	}
	select {
	case <-c.interrupts:
		t.Fatal("a normal-urgency intent must NOT raise an interrupt")
	case <-time.After(300 * time.Millisecond):
		// good — no interrupt
	}
}

// Analysis-mode freezes the reactive path entirely (no extraction).
func TestReactiveFrozenUnderAnalysis(t *testing.T) {
	h := newTestHost()
	fake := &fakeExtractClient{}
	h.configureAnalysis("Operator", fake, nil)
	h.EnterAnalysis()

	h.reactiveObserve("player_chat", "Eve", "pickaxe pickaxe pickaxe", 0.9)
	h.spawnExtract("Eve", "player")
	time.Sleep(100 * time.Millisecond)
	if fake.calls != 0 {
		t.Fatalf("analysis mode must freeze reactive extraction; got %d calls", fake.calls)
	}
}
