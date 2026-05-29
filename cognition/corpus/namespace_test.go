package corpus

import (
	"context"
	"testing"
)

// gameplaySource + devSource are tiny fixtures used by the
// federation tests. Each holds one easily-identifiable chunk.
func gameplaySource() Source {
	return Source{
		Name:      "rscwiki",
		Namespace: Gameplay,
		Corpus: LoadFromChunks([]Chunk{{
			Source: "rscwiki", PageTitle: "Cooking",
			Text: "Use a raw lobster on a fire to cook it.",
		}}),
	}
}

func devSource() Source {
	return Source{
		Name:      "openrsc-source",
		Namespace: Dev,
		Corpus: LoadFromChunks([]Chunk{{
			Source: "openrsc", PageTitle: "PlayerHandler.java",
			Text: "Inside the cookLobster handler, the server checks the lobster id is 13 then awards 90 xp.",
		}}),
	}
}

func TestFederationDevModeIncludesDev(t *testing.T) {
	f := NewFederation(
		[]Source{gameplaySource(), devSource()},
		[]Namespace{Gameplay, Dev},
	)
	if f.Len() != 2 {
		t.Errorf("Len: got %d, want 2", f.Len())
	}
	out, err := f.Recall(context.Background(), "lobster", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("hits: got %d, want 2 (both sources match \"lobster\")", len(out))
	}
}

func TestFederationProductionModeBlocksDev(t *testing.T) {
	f := NewFederation(
		[]Source{gameplaySource(), devSource()},
		[]Namespace{Gameplay}, // production: gameplay only
	)
	if f.Len() != 1 {
		t.Errorf("Len: got %d, want 1 (dev source must be filtered out)", f.Len())
	}
	out, err := f.Recall(context.Background(), "lobster", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("hits: got %d, want 1", len(out))
	}
	if out[0].Source != "rscwiki" {
		t.Errorf("source: got %q, want \"rscwiki\" — dev content leaked into production!", out[0].Source)
	}
}

func TestFederationDefaultAllowedIsGameplayOnly(t *testing.T) {
	// Nil allowed → defaults to Gameplay. This is the safe default
	// so a forgotten config can't accidentally enable dev sources.
	f := NewFederation(
		[]Source{gameplaySource(), devSource()},
		nil,
	)
	if f.Len() != 1 {
		t.Errorf("Len: got %d, want 1 (nil allowed must default to gameplay-only)", f.Len())
	}
}

func TestFederationNilCorpusSourceIsSkipped(t *testing.T) {
	// A Source with a nil Corpus is silently dropped — keeps wiring
	// code simple when an ingest hasn't been built yet.
	f := NewFederation(
		[]Source{
			gameplaySource(),
			{Name: "autorune", Namespace: Dev, Corpus: nil},
		},
		[]Namespace{Gameplay, Dev},
	)
	if f.Len() != 1 {
		t.Errorf("Len: got %d, want 1 (nil-corpus source must be skipped)", f.Len())
	}
}

func TestFederationSourcesReportsPostFilterSet(t *testing.T) {
	f := NewFederation(
		[]Source{gameplaySource(), devSource()},
		[]Namespace{Gameplay},
	)
	infos := f.Sources()
	if len(infos) != 1 {
		t.Fatalf("Sources: got %d, want 1", len(infos))
	}
	if infos[0].Name != "rscwiki" || infos[0].Namespace != Gameplay {
		t.Errorf("info: got %+v, want {rscwiki, gameplay}", infos[0])
	}
}

func TestFederationEmptyFederationReturnsNoHits(t *testing.T) {
	f := NewFederation(nil, []Namespace{Gameplay})
	if f.Len() != 0 {
		t.Errorf("Len: got %d, want 0", f.Len())
	}
	out, err := f.Recall(context.Background(), "anything", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if out != nil {
		t.Errorf("hits: got %v, want nil", out)
	}
}
