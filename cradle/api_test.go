package cradle

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/runtime"
)

func getJSON(t *testing.T, url string, v any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}

func getCode(t *testing.T, url string) int {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func postCode(t *testing.T, url string) int {
	t.Helper()
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestAPIServesIndex(t *testing.T) {
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /: status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct == "" || ct[:9] != "text/html" {
		t.Fatalf("GET /: content-type %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "CRADLE") {
		t.Fatal("index HTML did not render the app shell")
	}
}

func TestAPIServesV2(t *testing.T) {
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v2/: status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("GET /v2/: content-type %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	want, err := v2FS.ReadFile("web/v2/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, want) {
		t.Fatal("GET /v2/ did not serve the embedded web/v2/index.html")
	}

	// Sub-paths resolve through the same embedded tree (StripPrefix works)...
	if code := getCode(t, ts.URL+"/v2/index.html"); code != http.StatusOK {
		t.Fatalf("GET /v2/index.html: status %d", code)
	}
	// ...and a missing asset is a real 404, not an index fallback.
	if code := getCode(t, ts.URL+"/v2/no-such-asset.js"); code != http.StatusNotFound {
		t.Fatalf("GET /v2/no-such-asset.js: want 404, got %d", code)
	}
}

// TestDurableEventKinds pins the JSONL allowlist to REAL event kinds: the
// chat_message typo silently dropped chat from the durable record, so every
// family is cross-checked against the event package's Kind() strings.
func TestDurableEventKinds(t *testing.T) {
	kinds := make(map[string]bool, len(durableEventKinds))
	for _, k := range durableEventKinds {
		kinds[k] = true
	}
	if kinds["chat_message"] {
		t.Fatal(`allowlist carries the "chat_message" typo (the real kind is chat_received)`)
	}
	for _, ev := range []event.Event{
		event.ChatReceived{},
		event.TradeRequestReceived{}, event.TradeOpened{}, event.TradeOtherOffer{},
		event.TradeOtherAccepted{}, event.TradeConfirmShown{}, event.TradeClosed{},
		event.BankOpened{}, event.BankSlotUpdate{}, event.BankClosed{},
		event.ShopOpened{}, event.ShopClosed{},
		event.LevelUp{}, event.Death{}, event.TargetDied{}, event.FatigueUpdate{},
	} {
		if !kinds[ev.Kind()] {
			t.Errorf("durable allowlist missing kind %q", ev.Kind())
		}
	}
}

func TestAPIPersonaProxy(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	t.Setenv(adminTokenEnv, "") // no admin creds: every persona fetch must 501
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	reg.run = blockingRun(&runtime.HostHandle{})
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := reg.Start(ctx, testSpec("alpha")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, reg, "alpha", StatusRunning, 2*time.Second)

	// Unknown host -> 404 (registry idiom, same as GET /api/hosts/{name}).
	if code := getCode(t, ts.URL+"/api/hosts/nope/persona"); code != http.StatusNotFound {
		t.Fatalf("unknown host: want 404, got %d", code)
	}

	get501 := func(url string) {
		t.Helper()
		resp, err := http.Get(url)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Fatalf("GET %s: want 501, got %d", url, resp.StatusCode)
		}
		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode %s: %v", url, err)
		}
		if body["error"] == "" {
			t.Fatalf("GET %s: 501 body missing error field", url)
		}
	}

	// No admin credentials -> 501 with an error body (the UI hides the pane).
	get501(ts.URL + "/api/hosts/alpha/persona")

	// Credentialed but mesa unreachable -> still 501 (the RPC fails at dial).
	t.Setenv(adminTokenEnv, "secret")
	spec := testSpec("beta")
	spec.Mesa = "127.0.0.1:1" // nothing listens; connection refused, not a hang
	if err := reg.Start(ctx, spec); err != nil {
		t.Fatal(err)
	}
	get501(ts.URL + "/api/hosts/beta/persona")

	// Host with no mesa address at all -> 501 before any dial.
	noMesa := testSpec("gamma")
	noMesa.Mesa = ""
	noMesa.Director.Routine = "r.dsl" // keep the spec valid without mesa
	noMesa.Goal = ""
	if err := reg.Start(ctx, noMesa); err != nil {
		t.Fatal(err)
	}
	get501(ts.URL + "/api/hosts/gamma/persona")
}

func TestAPIControlEndpoints(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	conductor := runtime.NewConductor(nil, runtime.ConductorOptions{Director: runtime.Sequence()})
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	reg.run = blockingRun(&runtime.HostHandle{Conductor: conductor}) // Host nil => not live for debug proxy
	api := NewAPI(reg, slog.Default())                               // registers hooks

	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := reg.Start(ctx, testSpec("alpha")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, reg, "alpha", StatusRunning, 2*time.Second)

	// GET /api/hosts -> list
	var list []HostStatus
	getJSON(t, ts.URL+"/api/hosts", &list)
	if len(list) != 1 || list[0].Name != "alpha" {
		t.Fatalf("list wrong: %+v", list)
	}

	// GET /api/hosts/{name} -> status
	var st HostStatus
	getJSON(t, ts.URL+"/api/hosts/alpha", &st)
	if st.Name != "alpha" || st.Status != StatusRunning {
		t.Fatalf("get wrong: %+v", st)
	}

	// Unknown host -> 404
	if code := getCode(t, ts.URL+"/api/hosts/nope"); code != http.StatusNotFound {
		t.Fatalf("unknown host: want 404, got %d", code)
	}

	// Pause / resume drive the conductor.
	if code := postCode(t, ts.URL+"/api/hosts/alpha/pause"); code != http.StatusOK {
		t.Fatalf("pause: %d", code)
	}
	if !conductor.Paused() {
		t.Fatal("conductor should be paused after POST /pause")
	}
	if code := postCode(t, ts.URL+"/api/hosts/alpha/resume"); code != http.StatusOK {
		t.Fatalf("resume: %d", code)
	}
	if conductor.Paused() {
		t.Fatal("conductor should be running after POST /resume")
	}

	// Debug proxy: fake handle has no live Host, so the surface is not live -> 404.
	if code := getCode(t, ts.URL+"/api/hosts/alpha/debug/state"); code != http.StatusNotFound {
		t.Fatalf("debug proxy (not live): want 404, got %d", code)
	}

	// Stop.
	if code := postCode(t, ts.URL+"/api/hosts/alpha/stop"); code != http.StatusOK {
		t.Fatalf("stop: %d", code)
	}
	waitStatus(t, reg, "alpha", StatusStopped, 2*time.Second)
	reg.Wait()
}
