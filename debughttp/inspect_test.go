package debughttp_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// newInspectServer is the shared harness for the Tablet read endpoints: a bare
// (offline) host behind the debug mux, exactly as the cradle proxies it.
func newInspectServer(t *testing.T) (*runtime.Host, *httptest.Server) {
	t.Helper()
	host := runtime.New(runtime.Options{Username: "x"})
	t.Cleanup(func() { host.Close() })
	d := debughttp.New(host, debughttp.Config{Username: "x"}, slog.Default())
	ts := httptest.NewServer(d.Handler())
	t.Cleanup(ts.Close)
	return host, ts
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

// TestAspirationsEndpoint verifies the /aspirations wire contract: a host with
// a dark ladder yields {"aspirations":[]} (an array, never null) — and the
// AspirationStatus json tags produce the snake_case keys the Tablet's
// cognition tree reads, with a never-touched rung omitting last_touched.
func TestAspirationsEndpoint(t *testing.T) {
	_, ts := newInspectServer(t)

	var out struct {
		Aspirations []map[string]any `json:"aspirations"`
	}
	body := map[string]json.RawMessage{}
	getJSON(t, ts.URL+"/aspirations", &body)
	raw, ok := body["aspirations"]
	if !ok || strings.TrimSpace(string(raw)) == "null" {
		t.Fatalf("aspirations must be an array, got %s", raw)
	}
	if err := json.Unmarshal(raw, &out.Aspirations); err != nil {
		t.Fatal(err)
	}
	if len(out.Aspirations) != 0 {
		t.Fatalf("dark ladder should be empty, got %v", out.Aspirations)
	}

	// The wire shape of one rung (the struct's json tags, not the handler).
	b, err := json.Marshal(runtime.AspirationStatus{
		ID: "aspiration:master-a-craft", Label: "master a craft",
		GoalsDone: 2, GoalsActive: 1, GoalsOpen: 3,
		LastTouched: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC), Neglected: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "label", "goals_done", "goals_active", "goals_open", "last_touched", "neglected"} {
		if _, ok := wire[key]; !ok {
			t.Errorf("AspirationStatus JSON missing %q: %s", key, b)
		}
	}
	if b, err = json.Marshal(runtime.AspirationStatus{ID: "a", Label: "a"}); err != nil {
		t.Fatal(err)
	}
	wire = map[string]any{}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	if _, present := wire["last_touched"]; present {
		t.Errorf("never-touched rung must omit last_touched: %s", b)
	}
}

// TestFogEndpointShape verifies /fog's numeric-only contract on an oracle-less
// host: every key present with zero/empty values (the degraded read), arrays
// never null, and NO mask keys — the canvas-veil data stays shelved.
func TestFogEndpointShape(t *testing.T) {
	host, ts := newInspectServer(t)
	host.World().Self.SetPosition(world.Coord{X: 120, Y: 504})

	var out struct {
		Coverage struct {
			Frac  float64 `json:"frac"`
			Seen  int     `json:"seen"`
			Total int     `json:"total"`
		} `json:"coverage"`
		KnownPOIs *int               `json:"known_pois"`
		Frontiers []map[string]any   `json:"frontiers"`
		Here      map[string]float64 `json:"here"`
		Adjacent  []map[string]any   `json:"adjacent"`
	}
	getJSON(t, ts.URL+"/fog", &out)
	if out.KnownPOIs == nil {
		t.Error("known_pois key missing")
	}
	if out.Frontiers == nil || out.Adjacent == nil {
		t.Errorf("frontiers/adjacent must be arrays, got %v / %v", out.Frontiers, out.Adjacent)
	}
	for _, key := range []string{"terrain", "contents"} {
		if _, ok := out.Here[key]; !ok {
			t.Errorf("here missing %q: %v", key, out.Here)
		}
	}
	if out.Coverage.Total != 0 || out.Coverage.Seen != 0 || out.Coverage.Frac != 0 {
		t.Errorf("oracle-less coverage should be zeros, got %+v", out.Coverage)
	}

	// No cell/sector masks ever (numeric-only contract).
	var raw map[string]json.RawMessage
	getJSON(t, ts.URL+"/fog", &raw)
	for key := range raw {
		if strings.Contains(key, "mask") || strings.Contains(key, "cells") || strings.Contains(key, "sectors") {
			t.Errorf("fog snapshot leaked a mask-ish key %q", key)
		}
	}
}

// TestDecisionsEndpoint drives /decisions over the real published-path seam
// (SetDecisionLogPath — what RunHost wires): tail truncation, chronological
// order, the RFC3339 before-cursor, and the bad-cursor 400.
func TestDecisionsEndpoint(t *testing.T) {
	host, ts := newInspectServer(t)

	path := filepath.Join(t.TempDir(), "decisions.jsonl")
	base := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	var lines strings.Builder
	for i := 0; i < 5; i++ {
		b, _ := json.Marshal(map[string]any{
			"at": base.Add(time.Duration(i) * time.Minute), "trigger": "tick",
			"kind": "act", "reasoning": "r", "goal": "g",
		})
		lines.Write(b)
		lines.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(lines.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	host.SetDecisionLogPath(path)

	var out struct {
		Decisions []struct {
			At      time.Time `json:"at"`
			Trigger string    `json:"trigger"`
			Kind    string    `json:"kind"`
		} `json:"decisions"`
	}
	getJSON(t, ts.URL+"/decisions?tail=3", &out)
	if len(out.Decisions) != 3 {
		t.Fatalf("tail=3: got %d records", len(out.Decisions))
	}
	if !out.Decisions[0].At.Equal(base.Add(2*time.Minute)) || !out.Decisions[2].At.Equal(base.Add(4*time.Minute)) {
		t.Errorf("want the newest 3 in chronological order, got %v..%v", out.Decisions[0].At, out.Decisions[2].At)
	}
	if out.Decisions[0].Trigger != "tick" || out.Decisions[0].Kind != "act" {
		t.Errorf("record fields not carried: %+v", out.Decisions[0])
	}

	// before= excludes records at/after the cursor.
	getJSON(t, ts.URL+"/decisions?tail=10&before="+base.Add(2*time.Minute).Format(time.RFC3339), &out)
	if len(out.Decisions) != 2 {
		t.Fatalf("before-cursor: got %d records, want 2", len(out.Decisions))
	}

	resp, err := http.Get(ts.URL + "/decisions?before=lunchtime")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad before= should 400, got %d", resp.StatusCode)
	}
}

// TestDecisionsEndpointNoStream: a host that never published a decisions path
// (fresh/ephemeral) serves an empty array, not an error.
func TestDecisionsEndpointNoStream(t *testing.T) {
	_, ts := newInspectServer(t)
	var body map[string]json.RawMessage
	getJSON(t, ts.URL+"/decisions", &body)
	if got := strings.TrimSpace(string(body["decisions"])); got != "[]" {
		t.Errorf("decisions = %s, want []", got)
	}
}

// TestWhereEndpoint verifies the position-as-text view: floor derives from the
// 944-tile Y band (the boxed-upstairs triage signal), and the narration +
// surroundings come back even with no facts loaded.
func TestWhereEndpoint(t *testing.T) {
	host, ts := newInspectServer(t)
	host.World().Self.SetPosition(world.Coord{X: 120, Y: 944 + 10}) // floor 1

	var out struct {
		Text         string `json:"text"`
		Floor        int    `json:"floor"`
		X            int    `json:"x"`
		Y            int    `json:"y"`
		Surroundings string `json:"surroundings"`
	}
	getJSON(t, ts.URL+"/where", &out)
	if out.X != 120 || out.Y != 954 {
		t.Errorf("coords: got (%d,%d), want (120,954)", out.X, out.Y)
	}
	if out.Floor != 1 {
		t.Errorf("floor: got %d, want 1 (y/944)", out.Floor)
	}
	if !strings.Contains(out.Text, "120") {
		t.Errorf("text should narrate the position, got %q", out.Text)
	}
	if !strings.Contains(out.Surroundings, "Self:") {
		t.Errorf("surroundings should carry the DescribeSurroundings report, got %q", out.Surroundings)
	}
}
