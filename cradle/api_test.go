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
