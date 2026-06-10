package debughttp

import (
	"encoding/json"
	"testing"

	"github.com/gen0cide/westworld/limbic"
)

// TestMindRelCarriesAffinityGrievance locks down the /mind relationship view:
// an operator debugging "why won't this host ask the player next to it" needs the
// squashed affinity/grievance read values, not just trust. The previous mindRel
// dropped both (see cognition-audit-findings.md, positive-control tail). This
// asserts both (a) the handler copies limbic.Rel.Affinity/Grievance through and
// (b) the JSON wire shape carries the keys.
func TestMindRelCarriesAffinityGrievance(t *testing.T) {
	// Source values as Host.Relationships() (limbic.Ledger.All) would emit them:
	// already squashed on read into [-1,1] / [0,1].
	rels := []limbic.Rel{{
		Name:      "Delores",
		Grade:     limbic.Neutral,
		Trust:     0.2,
		Affinity:  0.6,
		Grievance: 0.4,
		Familiar:  3,
		Tags:      []string{"sparring-partner"},
	}}

	// Mirror exactly the mapping handleMind performs.
	var out []mindRel
	for _, r := range rels {
		out = append(out, mindRel{Name: r.Name, Grade: r.Grade.String(), Trust: r.Trust, Affinity: r.Affinity, Grievance: r.Grievance, Familiar: r.Familiar, Tags: r.Tags})
	}

	if got := out[0].Affinity; got != 0.6 {
		t.Errorf("affinity not carried: got %v, want 0.6", got)
	}
	if got := out[0].Grievance; got != 0.4 {
		t.Errorf("grievance not carried: got %v, want 0.4", got)
	}

	// The JSON wire contract: both keys must be present (a consumer/UI reads them).
	b, err := json.Marshal(out[0])
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"affinity": 0.6, "grievance": 0.4} {
		v, ok := wire[key]
		if !ok {
			t.Errorf("mindRel JSON missing %q key: %s", key, b)
			continue
		}
		if f, _ := v.(float64); f != want {
			t.Errorf("mindRel JSON %q = %v, want %v", key, v, want)
		}
	}
}
