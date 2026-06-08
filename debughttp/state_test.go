package debughttp_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// TestStateIncludesSkills verifies the /state snapshot carries the full 18-skill
// level/xp table the UI's Skills panel renders, including a drained current level.
func TestStateIncludesSkills(t *testing.T) {
	host := runtime.New(runtime.Options{Username: "x"})
	t.Cleanup(func() { host.Close() })

	var cur, mx, xp [world.NumSkills]int
	for i := range cur {
		cur[i], mx[i] = 1, 1
	}
	mx[14], cur[14], xp[14] = 30, 30, 13363 // mining: level 30
	mx[3], cur[3], xp[3] = 10, 8, 1154      // hits: base 10, drained to 8
	host.World().Self.SetAllSkills(cur, mx, xp, 0)

	d := debughttp.New(host, debughttp.Config{Username: "x"}, slog.Default())
	ts := httptest.NewServer(d.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var snap struct {
		Skills []struct {
			Name  string `json:"name"`
			Level int    `json:"level"`
			Cur   int    `json:"cur"`
			XP    int    `json:"xp"`
		} `json:"skills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}

	if len(snap.Skills) != world.NumSkills {
		t.Fatalf("skills: got %d, want %d", len(snap.Skills), world.NumSkills)
	}
	type sk struct{ Level, Cur, XP int }
	byName := map[string]sk{}
	for _, k := range snap.Skills {
		byName[k.Name] = sk{k.Level, k.Cur, k.XP}
	}
	if m := byName["mining"]; m.Level != 30 || m.XP != 13363 {
		t.Errorf("mining: got level=%d xp=%d, want 30 / 13363", m.Level, m.XP)
	}
	if h := byName["hits"]; h.Level != 10 || h.Cur != 8 {
		t.Errorf("hits: got level=%d cur=%d, want base 10 / drained 8", h.Level, h.Cur)
	}
}
