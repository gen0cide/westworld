package debughttp_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/runtime"
)

// whisperServer builds the standard Handler()-only test rig (the state_test
// idiom: no Activate — /whisper never touches the interpreter session).
func whisperServer(t *testing.T) (*runtime.Host, *httptest.Server) {
	t.Helper()
	host := runtime.New(runtime.Options{Username: "x"})
	t.Cleanup(func() { host.Close() })
	d := debughttp.New(host, debughttp.Config{Username: "x"}, slog.Default())
	ts := httptest.NewServer(d.Handler())
	t.Cleanup(ts.Close)
	return host, ts
}

func postWhisper(t *testing.T, ts *httptest.Server, body string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Post(ts.URL+"/whisper", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, out
}

// TestWhisperEndpointQueuesCapsAndDrains proves the queue contract end-to-end:
// each POST queues (queued gauge grows), the queue caps at 8 with the OLDEST
// dropped, the bus carries each whisper as kind "whisper", and the drain seam
// the situation render uses (TakeWhispers) clears the queue — a whisper is
// delivered once, then gone.
func TestWhisperEndpointQueuesCapsAndDrains(t *testing.T) {
	host, ts := whisperServer(t)
	ch := host.Bus().Subscribe("whisper", 32)

	// First whisper: queued=1, default urgency normal.
	code, out := postWhisper(t, ts, `{"text":"the bank is north of you"}`)
	if code != http.StatusOK || out["ok"] != true {
		t.Fatalf("POST /whisper = %d %v, want 200 ok", code, out)
	}
	if out["queued"] != float64(1) || out["urgency"] != "normal" {
		t.Fatalf("first whisper: queued=%v urgency=%v, want 1/normal", out["queued"], out["urgency"])
	}
	if out["ledger_written"] != false {
		t.Fatal("no subject+claim ⇒ ledger_written must be false")
	}

	// The bus event lands synchronously (QueueWhisper publishes before returning).
	select {
	case ev := <-ch:
		w, ok := ev.(event.WhisperReceived)
		if !ok || ev.Kind() != "whisper" {
			t.Fatalf("bus event = %T kind %q, want WhisperReceived/whisper", ev, ev.Kind())
		}
		if w.Text != "the bank is north of you" || w.Urgency != "normal" {
			t.Fatalf("bus payload wrong: %+v", w)
		}
	case <-time.After(time.Second):
		t.Fatal("no whisper event published on the bus")
	}

	// Overfill: 9 more (10 total) — the cap is 8, oldest dropped.
	for i := 0; i < 9; i++ {
		code, out = postWhisper(t, ts, fmt.Sprintf(`{"text":"w%d","urgency":"low"}`, i))
		if code != http.StatusOK {
			t.Fatalf("whisper %d failed: %v", i, out)
		}
	}
	if out["queued"] != float64(8) {
		t.Fatalf("queue must cap at 8, gauge says %v", out["queued"])
	}

	// Drain (the situation render's consume-once read): 8 entries, oldest
	// first, the first two whispers ("the bank…", "w0") evicted.
	ws := host.TakeWhispers()
	if len(ws) != 8 {
		t.Fatalf("TakeWhispers: got %d, want 8", len(ws))
	}
	if ws[0].Text != "w1" || ws[7].Text != "w8" {
		t.Fatalf("cap must drop the OLDEST: got first=%q last=%q, want w1/w8", ws[0].Text, ws[7].Text)
	}
	// Consume-once: a second drain is empty, and the gauge agrees.
	if again := host.TakeWhispers(); len(again) != 0 {
		t.Fatalf("drain must clear the queue, second take returned %d", len(again))
	}
	if host.PendingWhispers() != 0 {
		t.Fatalf("gauge after drain = %d, want 0", host.PendingWhispers())
	}
}

// TestWhisperEndpointValidation pins the 4xx contract: empty/missing text, an
// unknown urgency (rejected loudly — a typo must not silently downgrade an
// urgent whisper), a half-given subject/claim pair, and a non-POST method.
func TestWhisperEndpointValidation(t *testing.T) {
	host, ts := whisperServer(t)

	for name, body := range map[string]string{
		"empty text":            `{"text":"  "}`,
		"missing text":          `{"urgency":"high"}`,
		"bad urgency":           `{"text":"hi","urgency":"urgent"}`,
		"subject without claim": `{"text":"hi","subject":"Nurmof"}`,
		"claim without subject": `{"text":"hi","claim":"sells pickaxes"}`,
		"bad json":              `{"text":`,
	} {
		if code, out := postWhisper(t, ts, body); code != http.StatusBadRequest || out["error"] == "" {
			t.Errorf("%s: got %d %v, want 400 with error", name, code, out)
		}
	}
	if host.PendingWhispers() != 0 {
		t.Fatalf("rejected whispers must not queue: %d pending", host.PendingWhispers())
	}

	resp, err := http.Get(ts.URL + "/whisper")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET /whisper = %d, want 405", resp.StatusCode)
	}
}

// TestWhisperEndpointLedgerWrite proves the subject+claim teach: the claim
// lands in the knowledge ledger operator-grade — game-authoritative provenance
// (the reactive writeback's role-derived ProvSystem, exactly the mentor boost)
// at the 0.85 authoritative default when no confidence is given — and the
// whisper itself still queues.
func TestWhisperEndpointLedgerWrite(t *testing.T) {
	host, ts := whisperServer(t)

	code, out := postWhisper(t, ts,
		`{"text":"Nurmof sells pickaxes, go see him","subject":"Nurmof","claim":"sells pickaxes"}`)
	if code != http.StatusOK || out["ok"] != true || out["ledger_written"] != true {
		t.Fatalf("ledger whisper failed: %d %v", code, out)
	}
	if out["queued"] != float64(1) {
		t.Fatalf("the thought must queue alongside the teach: queued=%v", out["queued"])
	}

	var fact *knowledge.Fact
	for _, f := range host.KnowledgeFacts() {
		if f.Subject == "Nurmof" {
			fact = &f
			break
		}
	}
	if fact == nil || len(fact.Beliefs) == 0 {
		t.Fatalf("claim did not reach the knowledge ledger: %+v", host.KnowledgeFacts())
	}
	b := fact.Beliefs[0]
	if b.Claim != "sells pickaxes" || b.Provenance != knowledge.ProvSystem {
		t.Fatalf("operator claim must be game-authoritative: %+v", b)
	}
	if fact.Confidence < 0.8 {
		t.Fatalf("omitted confidence must default operator-grade (~0.85), got %v", fact.Confidence)
	}
}
