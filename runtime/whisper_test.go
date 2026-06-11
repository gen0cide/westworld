package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
)

// --- render: drain-once into the scene hint, first-person voice ---------------

// TestWhisperRenderDrainsOnce proves the situation-render contract: queued
// whispers surface in the scene hint EXACTLY once (oldest first, first-person,
// urgency-shaded voice) and the queue is cleared by the render — the next turn
// carries no echo.
func TestWhisperRenderDrainsOnce(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	h.QueueWhisper("the bank is north of you", "normal")
	h.QueueWhisper("just musing", "low")
	h.QueueWhisper("a scammer approaches", "high")

	sit := d.situation(h, Outcome{})
	scene := sit.Hints["scene"]
	for _, want := range []string{
		"💭",
		"A thought surfaces: the bank is north of you",
		"A passing thought: just musing",
		"A thought surfaces, urgent and impossible to ignore: a scammer approaches",
	} {
		if !strings.Contains(scene, want) {
			t.Fatalf("scene missing %q:\n%s", want, scene)
		}
	}
	// Oldest first: the normal whisper was queued before the low one.
	if strings.Index(scene, "the bank is north") > strings.Index(scene, "just musing") {
		t.Fatalf("whispers must render oldest first:\n%s", scene)
	}

	// Rendered once, then gone: the next turn is whisper-free.
	if h.PendingWhispers() != 0 {
		t.Fatalf("render must drain the queue, %d still pending", h.PendingWhispers())
	}
	sit2 := d.situation(h, Outcome{})
	if strings.Contains(sit2.Hints["scene"], "A thought surfaces") || strings.Contains(sit2.Hints["scene"], "💭") {
		t.Fatalf("a drained whisper must not re-render:\n%s", sit2.Hints["scene"])
	}
}

// TestWhisperSurvivesAnalysisDryRun proves the freeze gate: an analysis-mode
// dry-run (situation under freeze) must neither RENDER nor EAT a queued
// whisper — it stays queued and surfaces on the next real director turn
// (POST /whisper works regardless of analysis mode; render waits).
func TestWhisperSurvivesAnalysisDryRun(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	h.QueueWhisper("remember the toll gate", "normal")

	h.analysis.activeBit.Store(true) // freeze (the M16 dry-run path)
	sit := d.situation(h, Outcome{})
	if strings.Contains(sit.Hints["scene"], "A thought surfaces") {
		t.Fatalf("a frozen dry-run must not voice whispers:\n%s", sit.Hints["scene"])
	}
	if h.PendingWhispers() != 1 {
		t.Fatalf("a frozen dry-run must not consume the queue: %d pending, want 1", h.PendingWhispers())
	}

	h.analysis.activeBit.Store(false) // the director runs for real again
	sit2 := d.situation(h, Outcome{})
	if !strings.Contains(sit2.Hints["scene"], "A thought surfaces: remember the toll gate") {
		t.Fatalf("the queued whisper must render on the next real turn:\n%s", sit2.Hints["scene"])
	}
}

// --- queue: clip, cap (oldest dropped), bus publish ---------------------------

// TestQueueWhisperClipCapAndPublish pins the queue mechanics: text clips to
// ~300 runes, the queue caps at whisperQueueCap dropping the OLDEST, an
// unknown urgency normalizes to "normal", empties are ignored, and every
// accepted whisper is published on the bus as kind "whisper".
func TestQueueWhisperClipCapAndPublish(t *testing.T) {
	h := newTestHost()
	ch := h.bus.Subscribe("whisper", whisperQueueCap*2+4)

	h.QueueWhisper("   ", "normal") // blank: ignored entirely
	if h.PendingWhispers() != 0 {
		t.Fatal("a blank whisper must not queue")
	}

	long := strings.Repeat("я", whisperTextCap+50) // multibyte: rune-safe clip
	h.QueueWhisper(long, "shouting")               // unknown urgency → normal
	ws := h.TakeWhispers()
	if len(ws) != 1 || ws[0].Urgency != "normal" {
		t.Fatalf("unknown urgency must normalize: %+v", ws)
	}
	if r := []rune(ws[0].Text); len(r) != whisperTextCap || r[len(r)-1] != '…' {
		t.Fatalf("text must clip to %d runes with an ellipsis, got %d", whisperTextCap, len(r))
	}

	for i := 0; i < whisperQueueCap+2; i++ {
		h.QueueWhisper(string(rune('a'+i)), "low")
	}
	ws = h.TakeWhispers()
	if len(ws) != whisperQueueCap {
		t.Fatalf("queue must cap at %d, got %d", whisperQueueCap, len(ws))
	}
	if ws[0].Text != "c" { // "a","b" (the oldest two) dropped
		t.Fatalf("cap must drop the OLDEST first, queue starts with %q", ws[0].Text)
	}

	// Every accepted whisper hit the bus (1 clip + cap+2 = 11), blanks never.
	want := 1 + whisperQueueCap + 2
	got := 0
	for {
		select {
		case ev := <-ch:
			if _, ok := ev.(event.WhisperReceived); ok {
				got++
			}
		case <-time.After(100 * time.Millisecond):
			if got != want {
				t.Fatalf("bus publishes = %d, want %d", got, want)
			}
			return
		}
	}
}

// --- urgency: high interrupts like a high-urgency extraction ------------------

// TestWhisperHighUrgencyInterrupts proves the reactive-tier seam: urgency=high
// raises a conductor interrupt exactly like a high-urgency dialog extraction
// (tier "reactive", abort:false — park, note, resume), while normal/low queue
// silently for the next turn. No conductor (REPL/test default) must be a no-op.
func TestWhisperHighUrgencyInterrupts(t *testing.T) {
	// No conductor bound: high urgency must not panic, just queue.
	bare := newTestHost()
	bare.QueueWhisper("hello?", "high")
	if bare.PendingWhispers() != 1 {
		t.Fatal("a high whisper must queue even with no conductor")
	}

	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})
	h.configureAnalysis("Operator", nil, c)

	h.QueueWhisper("drop everything and check your inventory", "high")
	select {
	case req := <-c.interrupts:
		if req.tier != "reactive" {
			t.Fatalf("interrupt tier = %q, want reactive", req.tier)
		}
		if req.abort {
			t.Fatal("a whisper interrupt must be abort:false (park-and-resume)")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a high-urgency whisper did not raise a conductor interrupt")
	}
	// The payload rides the pending-reactive map (the coalescing channel item
	// is a wakeup marker); it must carry the whisper note intent.
	batch := c.takePendingReactive()
	if len(batch) != 1 || !strings.Contains(batch[0].reason, "whisper") {
		t.Fatalf("pending reactive batch = %+v, want one whisper detour", batch)
	}

	// Normal urgency: queued, no interrupt.
	h2 := newTestHost()
	c2 := NewConductor(h2, ConductorOptions{Detours: true, Director: nil})
	h2.configureAnalysis("Operator", nil, c2)
	h2.QueueWhisper("the bank is north", "normal")
	select {
	case <-c2.interrupts:
		t.Fatal("a normal-urgency whisper must NOT interrupt")
	case <-time.After(200 * time.Millisecond):
	}
	if h2.PendingWhispers() != 1 {
		t.Fatal("the normal whisper must still be queued for the next turn")
	}
}
