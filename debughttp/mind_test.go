package debughttp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
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

	// The exact mapping handleMind performs.
	var out []mindRel
	for _, r := range rels {
		out = append(out, toMindRel(r))
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

// TestMindRelCarriesInteractionsAndLastAt locks down the G-16 recency columns:
// the relationships table reads `interactions` (the limbic encounter counter —
// the ledger's one interaction counter) and `last_at` (unix seconds — the v2
// client computes ages/sorts numerically). A row from before LastAt stamping
// (zero time) must OMIT the key, not emit a 1970 epoch.
func TestMindRelCarriesInteractionsAndLastAt(t *testing.T) {
	at := time.Date(2026, 6, 10, 14, 2, 0, 0, time.UTC)
	got := toMindRel(limbic.Rel{Name: "Gregor", Grade: limbic.Friendly, Familiar: 47, LastAt: at})
	if got.Interactions != 47 {
		t.Errorf("interactions not carried from Familiar: got %d, want 47", got.Interactions)
	}
	if got.LastAt != at.Unix() {
		t.Errorf("last_at: got %d, want unix seconds %d", got.LastAt, at.Unix())
	}

	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	if v, ok := wire["interactions"].(float64); !ok || v != 47 {
		t.Errorf("mindRel JSON interactions = %v, want 47: %s", wire["interactions"], b)
	}
	if v, ok := wire["last_at"].(float64); !ok || v != float64(at.Unix()) {
		t.Errorf("mindRel JSON last_at = %v, want %d: %s", wire["last_at"], at.Unix(), b)
	}

	// Pre-stamping row: last_at omitted entirely.
	b, err = json.Marshal(toMindRel(limbic.Rel{Name: "old", Grade: limbic.Neutral}))
	if err != nil {
		t.Fatal(err)
	}
	wire = map[string]any{}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	if _, present := wire["last_at"]; present {
		t.Errorf("zero LastAt must omit the last_at key: %s", b)
	}
}

// TestMindFactFullBeliefs locks down the G-3 contract: without full, `beliefs`
// is the COUNT (the cheap always-on poll shape); with ?full=1 the same key
// becomes the complete belief list {claim,provenance,alpha,beta,at} the
// knowledge table's row-expand renders.
func TestMindFactFullBeliefs(t *testing.T) {
	fact := knowledge.Fact{
		Subject:    "furnace-key",
		Kind:       "item",
		Confidence: 0.55,
		Familiar:   2,
		Beliefs: []knowledge.Belief{
			{Claim: "the smith keeps it", Provenance: "hearsay", Alpha: 1.1, Beta: 0.9, At: 1750000000},
			{Claim: "found near the anvil", Provenance: "deduced", Alpha: 0.6, Beta: 1.0}, // no timestamp
		},
	}

	// Default shape: count.
	b, err := json.Marshal(toMindFact(fact, false))
	if err != nil {
		t.Fatal(err)
	}
	var slim map[string]any
	if err := json.Unmarshal(b, &slim); err != nil {
		t.Fatal(err)
	}
	if v, ok := slim["beliefs"].(float64); !ok || v != 2 {
		t.Fatalf("default beliefs should be the count 2, got %v: %s", slim["beliefs"], b)
	}

	// full=1 shape: the array, every belief, with α/β + provenance + unix-second
	// at (knowledge.Belief.At verbatim — the v2 client ages/sorts numerically).
	b, err = json.Marshal(toMindFact(fact, true))
	if err != nil {
		t.Fatal(err)
	}
	var full struct {
		TopClaim string `json:"top_claim"`
		Beliefs  []struct {
			Claim      string  `json:"claim"`
			Provenance string  `json:"provenance"`
			Alpha      float64 `json:"alpha"`
			Beta       float64 `json:"beta"`
			At         int64   `json:"at"`
		} `json:"beliefs"`
	}
	if err := json.Unmarshal(b, &full); err != nil {
		t.Fatalf("full beliefs did not decode as a list: %v: %s", err, b)
	}
	if full.TopClaim != "the smith keeps it" {
		t.Errorf("top_claim = %q, want the first belief", full.TopClaim)
	}
	if len(full.Beliefs) != 2 {
		t.Fatalf("full beliefs: got %d rows, want 2: %s", len(full.Beliefs), b)
	}
	first := full.Beliefs[0]
	if first.Claim != "the smith keeps it" || first.Provenance != "hearsay" || first.Alpha != 1.1 || first.Beta != 0.9 {
		t.Errorf("belief[0] fields not carried: %+v", first)
	}
	if first.At != 1750000000 {
		t.Errorf("belief[0].at = %d, want unix seconds 1750000000", first.At)
	}
	var raw struct {
		Beliefs []map[string]any `json:"beliefs"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw.Beliefs[1]["at"]; present {
		t.Errorf("belief without timestamp must omit at: %s", b)
	}
}
