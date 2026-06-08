package goalgraph

import "testing"

func TestUpsertAndStatus(t *testing.T) {
	g := New()
	g.Upsert("mine_ore", KindGoal, "Mine iron ore", StatusActive)
	n, ok := g.Get("MINE_ORE") // normalized
	if !ok || n.Kind != KindGoal || n.Status != StatusActive {
		t.Fatalf("node = %+v ok=%v", n, ok)
	}
	g.SetProgress("mine_ore", 0.5)
	g.SetStatus("mine_ore", StatusBlocked)
	n, _ = g.Get("mine_ore")
	if n.Progress != 0.5 || n.Status != StatusBlocked {
		t.Fatalf("after updates: %+v", n)
	}
}

func TestLinkDedupAndTraversal(t *testing.T) {
	g := New()
	// ore --serves--> smithing (downstream value chain)
	g.Link("iron_ore", "smithing", RelServes)
	g.Link("iron_ore", "smithing", RelServes) // duplicate ignored
	g.Upsert("smithing", KindGoal, "Train smithing", "")
	if got := len(g.Edges()); got != 1 {
		t.Fatalf("edges=%d, want 1 (dedup)", got)
	}
	// Endpoints auto-created; "iron_ore" exists as a state node.
	if !g.Has("iron_ore") {
		t.Fatal("Link should auto-create the 'from' endpoint")
	}
	out := g.Out("iron_ore", RelServes)
	if len(out) != 1 || out[0].To != "smithing" {
		t.Fatalf("Out(iron_ore, serves) = %+v", out)
	}
	in := g.In("smithing", RelServes)
	if len(in) != 1 || in[0].From != "iron_ore" {
		t.Fatalf("In(smithing, serves) = %+v", in)
	}
}

func TestBlockersAndFailureToEnablingGoal(t *testing.T) {
	g := New()
	// The bear example: mine is blocked by the bear; combat enables surviving.
	g.Upsert("mine_at_site", KindGoal, "Mine at the site", StatusActive)
	g.Link("mine_at_site", "bear", RelBlockedBy)
	blockers := g.Blockers("mine_at_site")
	if len(blockers) != 1 || blockers[0].ID != "bear" {
		t.Fatalf("blockers = %+v, want [bear]", blockers)
	}
	// Spawn the enabling sub-goal the way a later phase would.
	g.Upsert("train_combat", KindSubgoal, "Train combat to survive the bear", StatusOpen)
	g.Link("train_combat", "survive_at_site", RelEnables)
	if got := g.Out("train_combat", RelEnables); len(got) != 1 || got[0].To != "survive_at_site" {
		t.Fatalf("enables edge missing: %+v", got)
	}
}

func TestOpenQuestions(t *testing.T) {
	g := New()
	g.Upsert("where_pickaxe", KindOpenQuestion, "Where do I buy a pickaxe?", StatusOpen)
	g.Upsert("where_anvil", KindOpenQuestion, "Where is an anvil?", StatusOpen)
	g.Upsert("mine_ore", KindGoal, "Mine ore", StatusActive)
	qs := g.OpenQuestions()
	if len(qs) != 2 {
		t.Fatalf("open questions = %d, want 2", len(qs))
	}
	// Resolving one drops it from the open set.
	g.SetStatus("where_pickaxe", StatusDone)
	if got := len(g.OpenQuestions()); got != 1 {
		t.Fatalf("after resolving one, open=%d, want 1", got)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	g := New()
	g.Upsert("smithing", KindGoal, "Train smithing", StatusActive)
	g.SetProgress("smithing", 0.3)
	g.Tag("smithing", "core")
	g.Link("iron_ore", "smithing", RelServes)
	snap := g.Export()
	if len(snap.Nodes) != 2 || len(snap.Edges) != 1 {
		t.Fatalf("snapshot nodes=%d edges=%d, want 2/1", len(snap.Nodes), len(snap.Edges))
	}
	g2 := New()
	if g2.Has("smithing") {
		t.Fatal("fresh graph should be empty before import")
	}
	g2.Import(snap)
	n, ok := g2.Get("smithing")
	if !ok || n.Progress != 0.3 || len(n.Tags) != 1 {
		t.Fatalf("round-trip lost node data: %+v", n)
	}
	if len(g2.Out("iron_ore", RelServes)) != 1 {
		t.Fatal("round-trip lost edge")
	}
}
