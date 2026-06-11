package mesad

import (
	"testing"
	"time"
)

func TestGenesisRevFingerprintsAllInputs(t *testing.T) {
	base := genesisRev("prose", 5, 3, "mine tin")
	if base != genesisRev("prose", 5, 3, "mine tin") {
		t.Fatal("rev must be deterministic")
	}
	for name, rev := range map[string]string{
		"prose":      genesisRev("prose2", 5, 3, "mine tin"),
		"episodes":   genesisRev("prose", 6, 3, "mine tin"),
		"rels":       genesisRev("prose", 5, 4, "mine tin"),
		"objective":  genesisRev("prose", 5, 3, "smelt bronze"),
	} {
		if rev == base {
			t.Errorf("changing %s must change the rev", name)
		}
	}
}

func TestGenesisCacheFresh(t *testing.T) {
	now := time.Now()
	c := cachedGenesis{Rev: "r1", At: now.Add(-30 * time.Minute).Unix()}
	if !genesisCacheFresh(c, "r1", now) {
		t.Fatal("30-minute-old same-rev entry must be fresh")
	}
	if genesisCacheFresh(c, "r2", now) {
		t.Fatal("rev mismatch must miss")
	}
	stale := cachedGenesis{Rev: "r1", At: now.Add(-2 * time.Hour).Unix()}
	if genesisCacheFresh(stale, "r1", now) {
		t.Fatal("entry past TTL must miss")
	}
}
