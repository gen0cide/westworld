package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// flagSetFixture is a tiny FlagSet wrapper used to exercise parseFlags directly.
type flagSetFixture struct {
	fs *flag.FlagSet
	n  *int
}

func newFlagSetFixture() *flagSetFixture {
	fs := flag.NewFlagSet("fixture", flag.ContinueOnError)
	return &flagSetFixture{fs: fs, n: fs.Int("n", 0, "a number")}
}

// captured holds what the fake admin server saw for one request.
type captured struct {
	method string
	path   string
	body   map[string]any
}

// runCmd looks the command up in the registry, points a Client at a fake admin
// server, runs the command's handler, and returns the request the server saw
// (or the handler's error before any request was made). It exercises the real
// flag-parsing + body-construction path end to end.
func runCmd(t *testing.T, group, name string, args ...string) (captured, error) {
	t.Helper()

	// printResult dereferences this package-level flag; give it a backing value.
	no := false
	jsonOut = &no

	cmd := findCmd(group, name)
	if cmd == nil {
		t.Fatalf("command %q %q not registered", group, name)
	}

	var got captured
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &got.body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":null}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", 5*time.Second)
	err := cmd.run(c, args)
	return got, err
}

// numField reads a JSON number out of the decoded body as an int64. json.Unmarshal
// into map[string]any decodes numbers as float64.
func numField(t *testing.T, body map[string]any, key string) (int64, bool) {
	t.Helper()
	v, ok := body[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("field %q is %T, want number", key, v)
	}
	return int64(f), true
}

// TestFlagAfterPositional proves the H20 fix: flags written AFTER the positional
// are parsed, not silently dropped at their defaults.
func TestFlagAfterPositional(t *testing.T) {
	t.Run("mute -minutes 5 after username carries minutes=5", func(t *testing.T) {
		got, err := runCmd(t, "player", "mute", "Delores", "-minutes", "5", "-shadow")
		if err != nil {
			t.Fatalf("mute returned error: %v", err)
		}
		if !strings.HasSuffix(got.path, "/players/Delores/mute") {
			t.Fatalf("path = %q", got.path)
		}
		if m, ok := numField(t, got.body, "minutes"); !ok || m != 5 {
			t.Fatalf("minutes = %v (ok=%v), want 5", m, ok)
		}
		if got.body["shadow"] != true {
			t.Fatalf("shadow = %v, want true", got.body["shadow"])
		}
	})

	t.Run("mute -minutes 5 BEFORE username also works", func(t *testing.T) {
		got, err := runCmd(t, "player", "mute", "-minutes", "5", "Delores")
		if err != nil {
			t.Fatalf("mute returned error: %v", err)
		}
		if !strings.HasSuffix(got.path, "/players/Delores/mute") {
			t.Fatalf("path = %q", got.path)
		}
		if m, ok := numField(t, got.body, "minutes"); !ok || m != 5 {
			t.Fatalf("minutes = %v (ok=%v), want 5", m, ok)
		}
	})

	t.Run("kick -reason after username carries the reason", func(t *testing.T) {
		got, err := runCmd(t, "player", "kick", "Delores", "-reason", "spamming")
		if err != nil {
			t.Fatalf("kick returned error: %v", err)
		}
		if got.body["reason"] != "spamming" {
			t.Fatalf("reason = %v, want spamming", got.body["reason"])
		}
	})

	t.Run("restore -all-skills=false after username is honored", func(t *testing.T) {
		got, err := runCmd(t, "player", "restore", "Delores", "-all-skills=false")
		if err != nil {
			t.Fatalf("restore returned error: %v", err)
		}
		if got.body["allSkills"] != false {
			t.Fatalf("allSkills = %v, want false", got.body["allSkills"])
		}
	})

	t.Run("teleport -x -y after username does not error and is sent", func(t *testing.T) {
		got, err := runCmd(t, "player", "teleport", "Delores", "-x", "122", "-y", "503")
		if err != nil {
			t.Fatalf("teleport returned error: %v", err)
		}
		x, _ := numField(t, got.body, "x")
		y, _ := numField(t, got.body, "y")
		if x != 122 || y != 503 {
			t.Fatalf("x,y = %d,%d want 122,503", x, y)
		}
	})

	t.Run("inv-add -catalog-id after username is honored", func(t *testing.T) {
		got, err := runCmd(t, "player", "inv-add", "Delores", "-catalog-id", "10", "-amount", "3")
		if err != nil {
			t.Fatalf("inv-add returned error: %v", err)
		}
		if cid, ok := numField(t, got.body, "catalogId"); !ok || cid != 10 {
			t.Fatalf("catalogId = %v (ok=%v), want 10", cid, ok)
		}
		if amt, ok := numField(t, got.body, "amount"); !ok || amt != 3 {
			t.Fatalf("amount = %v (ok=%v), want 3", amt, ok)
		}
	})
}

// TestSummonPartialCoordinate proves the L16 fix: a lone -x (or lone -y) is an
// error, not a silent y=0 teleport; both together still work.
func TestSummonPartialCoordinate(t *testing.T) {
	t.Run("summon -x only errors", func(t *testing.T) {
		_, err := runCmd(t, "player", "summon", "Delores", "-x", "50")
		if err == nil {
			t.Fatal("summon -x 50 (no -y) should error, got nil")
		}
	})

	t.Run("summon -y only errors", func(t *testing.T) {
		_, err := runCmd(t, "player", "summon", "Delores", "-y", "50")
		if err == nil {
			t.Fatal("summon -y 50 (no -x) should error, got nil")
		}
	})

	t.Run("summon -x -y both works", func(t *testing.T) {
		got, err := runCmd(t, "player", "summon", "Delores", "-x", "50", "-y", "60")
		if err != nil {
			t.Fatalf("summon -x -y returned error: %v", err)
		}
		x, _ := numField(t, got.body, "x")
		y, _ := numField(t, got.body, "y")
		if x != 50 || y != 60 {
			t.Fatalf("x,y = %d,%d want 50,60", x, y)
		}
	})

	t.Run("summon-all -x only errors", func(t *testing.T) {
		_, err := runCmd(t, "player", "summon-all", "-x", "50")
		if err == nil {
			t.Fatal("summon-all -x 50 (no -y) should error, got nil")
		}
	})

	t.Run("summon-all -x -y both works", func(t *testing.T) {
		got, err := runCmd(t, "player", "summon-all", "-x", "50", "-y", "60")
		if err != nil {
			t.Fatalf("summon-all -x -y returned error: %v", err)
		}
		x, _ := numField(t, got.body, "x")
		y, _ := numField(t, got.body, "y")
		if x != 50 || y != 60 {
			t.Fatalf("x,y = %d,%d want 50,60", x, y)
		}
	})
}

// TestParseFlagsInterspersed exercises the helper directly, including the "--"
// terminator and multiple positionals.
func TestParseFlagsInterspersed(t *testing.T) {
	newFS := func() (*flagSetFixture, func() []string) {
		f := newFlagSetFixture()
		return f, f.fs.Args
	}

	t.Run("flag after positional", func(t *testing.T) {
		f, args := newFS()
		if err := parseFlags(f.fs, []string{"Delores", "-n", "5"}); err != nil {
			t.Fatal(err)
		}
		if *f.n != 5 {
			t.Fatalf("n = %d want 5", *f.n)
		}
		if got := args(); len(got) != 1 || got[0] != "Delores" {
			t.Fatalf("positionals = %v want [Delores]", got)
		}
	})

	t.Run("flag before positional", func(t *testing.T) {
		f, args := newFS()
		if err := parseFlags(f.fs, []string{"-n", "5", "Delores"}); err != nil {
			t.Fatal(err)
		}
		if *f.n != 5 || args()[0] != "Delores" {
			t.Fatalf("n=%d args=%v", *f.n, args())
		}
	})

	t.Run("two positionals straddling a flag", func(t *testing.T) {
		f, args := newFS()
		if err := parseFlags(f.fs, []string{"Delores", "-n", "5", "questId"}); err != nil {
			t.Fatal(err)
		}
		got := args()
		if len(got) != 2 || got[0] != "Delores" || got[1] != "questId" {
			t.Fatalf("positionals = %v want [Delores questId]", got)
		}
		if *f.n != 5 {
			t.Fatalf("n = %d want 5", *f.n)
		}
	})

	t.Run("double dash makes the rest positional", func(t *testing.T) {
		f, args := newFS()
		if err := parseFlags(f.fs, []string{"--", "-not-a-flag"}); err != nil {
			t.Fatal(err)
		}
		got := args()
		if len(got) != 1 || got[0] != "-not-a-flag" {
			t.Fatalf("positionals = %v want [-not-a-flag]", got)
		}
	})
}
