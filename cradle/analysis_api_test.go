package cradle

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/runtime"
)

// TestAnalysisUIServed confirms the operator-console UI is actually shipped in
// the served page — the toggle, directive box, verdict feed, and their handlers.
func TestAnalysisUIServed(t *testing.T) {
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	for _, want := range []string{
		`id="d-analysis-btn"`, `id="consolebox"`, `id="feed-console"`, `id="d-analysis-banner"`,
		"toggleAnalysis", "sendDirective", "renderVerdict", "applyAnalysisState",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("served UI is missing analysis-console piece %q", want)
		}
	}
}

// TestAnalysisRoutesRegistered confirms all four analysis routes are wired: for a
// non-existent host each returns 400 (the handler ran, host-not-running) — NOT a
// 404, which would mean the route itself is missing.
func TestAnalysisRoutesRegistered(t *testing.T) {
	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	cases := []struct{ method, path, body string }{
		{"POST", "/api/hosts/ghost/analysis/enter", ""},
		{"POST", "/api/hosts/ghost/analysis/exit", ""},
		{"POST", "/api/hosts/ghost/analysis/directive", "go to the bank"},
		{"GET", "/api/hosts/ghost/analysis", ""},
	}
	for _, tc := range cases {
		req, _ := http.NewRequest(tc.method, ts.URL+tc.path, strings.NewReader(tc.body))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", tc.method, tc.path, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("%s %s: 404 — route is NOT registered", tc.method, tc.path)
		} else if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s %s: got %d, want 400 (host not running)", tc.method, tc.path, resp.StatusCode)
		}
	}

	// An empty directive is rejected up front (before the host lookup).
	req, _ := http.NewRequest("POST", ts.URL+"/api/hosts/ghost/analysis/directive", nil)
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty directive: got %d, want 400", resp.StatusCode)
	}
}

// TestAnalysisRoundTripJSON drives a REAL host through the HTTP control plane and
// asserts the exact JSON the UI consumes: enter → {kind:"control", active:true},
// the 2s status poll (HostStatus.analysis) flips to true, GET → {active, last:{…}},
// exit → active:false and the poll flips back. This is the contract renderVerdict
// / applyAnalysisState / seedAnalysisFeed parse.
func TestAnalysisRoundTripJSON(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	host := runtime.New(runtime.Options{Username: "op"})
	t.Cleanup(func() { host.Close() })
	conductor := runtime.NewConductor(host, runtime.ConductorOptions{Director: runtime.Sequence()})

	reg := NewRegistry(runtime.SharedDeps{}, slog.Default())
	reg.run = blockingRun(&runtime.HostHandle{Host: host, Conductor: conductor})
	api := NewAPI(reg, slog.Default())
	ts := httptest.NewServer(api.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := reg.Start(ctx, testSpec("alpha")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, reg, "alpha", StatusRunning, 2*time.Second)

	post := func(path string) map[string]any {
		t.Helper()
		req, _ := http.NewRequest("POST", ts.URL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("POST %s: %d %s", path, resp.StatusCode, b)
		}
		var m map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		return m
	}

	// enter → control verdict, active:true
	v := post("/api/hosts/alpha/analysis/enter")
	if v["kind"] != "control" {
		t.Errorf("enter kind: got %v, want \"control\"", v["kind"])
	}
	if v["active"] != true {
		t.Errorf("enter active: got %v, want true", v["active"])
	}

	// The 2s status poll the UI uses for sync must now report analysis:true.
	var afterEnter HostStatus
	getJSON(t, ts.URL+"/api/hosts/alpha", &afterEnter)
	if !afterEnter.Analysis {
		t.Errorf("HostStatus.Analysis: got false, want true after enter (UI toggle sync)")
	}

	// GET state → {active:true, last:{kind:"control", ...}} (seedAnalysisFeed path)
	var state map[string]any
	getJSON(t, ts.URL+"/api/hosts/alpha/analysis", &state)
	if state["active"] != true {
		t.Errorf("GET analysis active: got %v, want true", state["active"])
	}
	if last, ok := state["last"].(map[string]any); !ok || last["kind"] != "control" {
		t.Errorf("GET analysis last: got %v, want {kind:control,...}", state["last"])
	}

	// exit → active:false, and the poll flips back.
	v = post("/api/hosts/alpha/analysis/exit")
	if v["active"] != false {
		t.Errorf("exit active: got %v, want false", v["active"])
	}
	var afterExit HostStatus
	getJSON(t, ts.URL+"/api/hosts/alpha", &afterExit)
	if afterExit.Analysis {
		t.Errorf("HostStatus.Analysis: got true, want false after exit")
	}
}
