package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gen0cide/westworld/cradle/hostcfg"
)

// loadGenerated writes genFleet output to a temp .hostcfg and loads it back
// through the real cradle loader — proving the generator's output is valid and
// the defaults/names land as expected.
func loadGenerated(t *testing.T, out []byte) []hostcfg.Host {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fleet.hostcfg")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	hosts, err := hostcfg.Load(path)
	if err != nil {
		t.Fatalf("generated hostcfg failed to load: %v\n--- yaml ---\n%s", err, out)
	}
	if err := hostcfg.ValidateSet(hosts); err != nil {
		t.Fatalf("generated hostcfg failed validation: %v", err)
	}
	return hosts
}

// TestGenFleetSharedGoal: names == host_ids, defaults applied, shared goal on all.
func TestGenFleetSharedGoal(t *testing.T) {
	ids := []string{"drone1", "drone2", "drone3"}
	out, err := genFleet(ids, fleetOpts{
		Server: "localhost:43594", Mesa: "localhost:7077",
		PasswordEnv: "WESTWORLD_PASSWORD", State: "memory", Goal: "mine tin to 30",
	})
	if err != nil {
		t.Fatal(err)
	}
	hosts := loadGenerated(t, out)
	if len(hosts) != 3 {
		t.Fatalf("got %d hosts, want 3", len(hosts))
	}
	byName := map[string]hostcfg.Host{}
	for _, h := range hosts {
		byName[h.Name] = h
	}
	for _, id := range ids {
		h, ok := byName[id]
		if !ok {
			t.Fatalf("missing host %q", id)
		}
		if h.Server != "localhost:43594" || h.Mesa != "localhost:7077" {
			t.Errorf("%s: defaults not applied: server=%q mesa=%q", id, h.Server, h.Mesa)
		}
		if h.PasswordEnvName() != "WESTWORLD_PASSWORD" {
			t.Errorf("%s: password env = %q", id, h.PasswordEnvName())
		}
		if h.Goal != "mine tin to 30" {
			t.Errorf("%s: goal = %q, want shared goal", id, h.Goal)
		}
	}
}

// TestGenFleetPerHostGoals: a per-host goal map lands on each host entry.
func TestGenFleetPerHostGoals(t *testing.T) {
	ids := []string{"alpha", "beta"}
	out, err := genFleet(ids, fleetOpts{
		Server: "s:1", Mesa: "m:2", PasswordEnv: "PW",
		Goals: map[string]string{"alpha": "goal-A", "beta": "goal-B"},
	})
	if err != nil {
		t.Fatal(err)
	}
	hosts := loadGenerated(t, out)
	byName := map[string]hostcfg.Host{}
	for _, h := range hosts {
		byName[h.Name] = h
	}
	if byName["alpha"].Goal != "goal-A" || byName["beta"].Goal != "goal-B" {
		t.Fatalf("per-host goals wrong: alpha=%q beta=%q", byName["alpha"].Goal, byName["beta"].Goal)
	}
}

func TestGenFleetEmpty(t *testing.T) {
	if _, err := genFleet(nil, fleetOpts{Server: "s", Mesa: "m"}); err == nil {
		t.Fatal("expected an error generating a fleet with no host_ids")
	}
}
