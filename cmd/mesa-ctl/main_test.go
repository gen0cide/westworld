package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCollectDir: a directory of <host_id>.json yields one upsert per file, with
// host_id from the filename and the raw bytes as the payload. Non-json is skipped.
func TestCollectDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("drone1.json", `{"a":1}`)
	write("drone2.json", `{"b":2}`)
	write("notes.txt", "ignore me") // non-json, must be skipped

	items, err := collectDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, it := range items {
		got[it.HostId] = string(it.PersonaJson)
	}
	if len(got) != 2 || got["drone1"] != `{"a":1}` || got["drone2"] != `{"b":2}` {
		t.Fatalf("collectDir = %v, want drone1/drone2 from the .json files only", got)
	}
}

// TestParseNDJSON: each line carries host_id + an inline persona object.
func TestParseNDJSON(t *testing.T) {
	in := strings.Join([]string{
		`{"host_id":"drone1","persona":{"a":1}}`,
		``, // blank lines tolerated
		`{"host_id":"drone2","persona":{"b":2}}`,
	}, "\n")
	items, err := parseNDJSON(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].HostId != "drone1" || string(items[0].PersonaJson) != `{"a":1}` {
		t.Fatalf("item0 = %s / %s", items[0].HostId, items[0].PersonaJson)
	}
	if items[1].HostId != "drone2" || string(items[1].PersonaJson) != `{"b":2}` {
		t.Fatalf("item1 = %s / %s", items[1].HostId, items[1].PersonaJson)
	}
}

func TestParseNDJSONBadLine(t *testing.T) {
	if _, err := parseNDJSON(strings.NewReader(`{not json`)); err == nil {
		t.Fatal("expected an error on malformed NDJSON")
	}
}
