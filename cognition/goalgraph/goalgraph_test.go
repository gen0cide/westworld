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

// TestOpenGoals proves OpenGoals returns only live KindOpenGoal nodes,
// newest-first, and drops a node once it's done/abandoned/promoted-away — the
// successor-candidate reader the director's advancement consumes.
func TestOpenGoals(t *testing.T) {
	g := New()
	// Non-candidates: a KindGoal (the active goal) and an open question.
	g.Upsert("mine ore", KindGoal, "Mine ore", StatusActive)
	g.Upsert("where_anvil", KindOpenQuestion, "Where is an anvil?", StatusOpen)
	// Candidates: two open_goal nodes. Force ordering via the injected clock so the
	// newest-first sort is deterministic (g.now resolution is whole seconds).
	tick := int64(100)
	g.now = func() int64 { tick++; return tick }
	g.Upsert("learn fishing", KindOpenGoal, "Learn fishing", StatusOpen)
	g.Upsert("see varrock", KindOpenGoal, "See Varrock", StatusOpen) // newer

	og := g.OpenGoals()
	if len(og) != 2 {
		t.Fatalf("open goals = %d, want 2 (only KindOpenGoal nodes)", len(og))
	}
	if og[0].ID != "see varrock" {
		t.Fatalf("newest-first: got %q first, want %q", og[0].ID, "see varrock")
	}
	// Closing a candidate (done/abandoned) drops it from the set.
	g.SetStatus("see varrock", StatusDone)
	g.SetStatus("learn fishing", StatusAbandoned)
	if got := len(g.OpenGoals()); got != 0 {
		t.Fatalf("after closing both, open goals = %d, want 0", got)
	}
}

// TestUntag proves Untag is the inverse of Tag, is idempotent, and never
// creates a node (H5 — the director clears a transient "spinning" tag on
// recovery so it does not pollute the goal's planner prose forever).
func TestUntag(t *testing.T) {
	g := New()
	g.Upsert("mine_ore", KindGoal, "Mine ore", StatusActive)
	g.Tag("mine_ore", "spinning")
	g.Tag("mine_ore", "core")
	g.Untag("mine_ore", "spinning")
	n, _ := g.Get("mine_ore")
	if len(n.Tags) != 1 || n.Tags[0] != "core" {
		t.Fatalf("after untag spinning: tags=%v, want [core]", n.Tags)
	}
	// Untagging an absent tag / absent node is a no-op (no panic, no creation).
	g.Untag("mine_ore", "not-there")
	g.Untag("nonexistent", "spinning")
	if g.Has("nonexistent") {
		t.Fatal("Untag must not create a node")
	}
	if got, _ := g.Get("mine_ore"); len(got.Tags) != 1 {
		t.Fatalf("no-op untag changed tags: %v", got.Tags)
	}
}

// TestPruneTerminal proves PruneTerminal bounds the node count by evicting the
// oldest terminal nodes first, preserves all non-terminal nodes even past the
// cap, and removes edges orphaned by an eviction (H10 — host-light backstop).
func TestPruneTerminal(t *testing.T) {
	g := New()
	tick := int64(100)
	g.now = func() int64 { tick++; return tick }
	// Three terminal nodes of ascending age, plus two live nodes.
	g.Upsert("done_old", KindGoal, "Old done goal", StatusDone)           // at 101 (oldest terminal)
	g.Upsert("done_mid", KindGoal, "Mid abandoned goal", StatusAbandoned) // at 102
	g.Upsert("done_new", KindGoal, "New done goal", StatusDone)           // at 103
	g.Upsert("live_goal", KindGoal, "Active goal", StatusActive)          // at 104 (live)
	g.Upsert("live_q", KindOpenQuestion, "Open question", StatusOpen)     // at 105 (live)
	// An edge that will be orphaned when done_old is evicted.
	g.Link("done_old", "live_goal", RelEnables)
	// An edge between two survivors that must remain.
	g.Link("live_q", "live_goal", RelServes)

	if got := len(g.Nodes()); got != 5 {
		t.Fatalf("setup nodes=%d, want 5", got)
	}
	// Cap at 4 → exactly one (the oldest terminal) must go.
	if evicted := g.PruneTerminal(4); evicted != 1 {
		t.Fatalf("PruneTerminal(4) evicted=%d, want 1", evicted)
	}
	if g.Has("done_old") {
		t.Fatal("oldest terminal node should have been evicted")
	}
	if !g.Has("done_mid") || !g.Has("done_new") || !g.Has("live_goal") || !g.Has("live_q") {
		t.Fatal("PruneTerminal evicted the wrong node(s)")
	}
	// The done_old->live_goal edge is now orphaned and must be gone; the
	// live_q->live_goal edge must survive.
	if got := len(g.Edges()); got != 1 {
		t.Fatalf("after prune edges=%d, want 1 (orphan dropped)", got)
	}
	if out := g.Out("live_q", RelServes); len(out) != 1 {
		t.Fatalf("survivor edge dropped: %+v", out)
	}

	// Cap below the live-node count: terminals are exhausted before touching
	// the two non-terminal nodes — they are NEVER evicted.
	evicted := g.PruneTerminal(1)
	if g.Has("done_mid") || g.Has("done_new") {
		t.Fatal("both terminal nodes should be evicted when capping hard")
	}
	if !g.Has("live_goal") || !g.Has("live_q") {
		t.Fatal("non-terminal nodes must survive even below the cap")
	}
	if evicted != 2 {
		t.Fatalf("hard cap evicted=%d, want 2 (only the terminals)", evicted)
	}
	// A no-op prune when already under the cap.
	if g.PruneTerminal(DefaultGraphCap) != 0 {
		t.Fatal("prune under cap should evict nothing")
	}
}

// TestUpsertNoKindDowngrade proves a re-adopt keyed on a colliding normalized id
// cannot demote a goal to a less-specific kind, but can still promote (L2).
func TestUpsertNoKindDowngrade(t *testing.T) {
	g := New()
	g.Upsert("Mine Ore", KindGoal, "Mine ore", StatusActive)
	// A re-adopt that normalizes to the same key with a lower-specificity kind
	// must NOT downgrade the existing goal.
	g.Upsert("mine ore", KindSubgoal, "Mine ore (restated)", "")
	n, _ := g.Get("mine ore")
	if n.Kind != KindGoal {
		t.Fatalf("kind downgraded to %q, want %q (no demote)", n.Kind, KindGoal)
	}
	if n.Label != "Mine ore (restated)" {
		t.Fatalf("label should still LWW-update: %q", n.Label)
	}
	// Promotion is allowed: an open_goal can become a goal.
	g.Upsert("see varrock", KindOpenGoal, "See Varrock", StatusOpen)
	g.Upsert("see varrock", KindGoal, "", "")
	if n, _ := g.Get("see varrock"); n.Kind != KindGoal {
		t.Fatalf("open_goal should promote to goal, got %q", n.Kind)
	}
}

// TestLinkCanonicalEndpoints proves Link stores the node's canonical ID, not the
// caller's case/spacing, so Edge.From/To always match the matching Node.ID (L3).
func TestLinkCanonicalEndpoints(t *testing.T) {
	g := New()
	g.Upsert("Iron Ore", KindState, "Iron ore", StatusOpen) // canonical id "Iron Ore"
	g.Upsert("Smithing", KindGoal, "Train smithing", StatusActive)
	// Link with DIFFERENT case/spacing — the edge must store the canonical IDs.
	g.Link("  iron ore  ", "smithing", RelServes)
	edges := g.Edges()
	if len(edges) != 1 {
		t.Fatalf("edges=%d, want 1", len(edges))
	}
	if edges[0].From != "Iron Ore" || edges[0].To != "Smithing" {
		t.Fatalf("edge endpoints not canonical: from=%q to=%q, want %q/%q",
			edges[0].From, edges[0].To, "Iron Ore", "Smithing")
	}
	// Dedup still works across case variants.
	g.Link("IRON ORE", "Smithing", RelServes)
	if got := len(g.Edges()); got != 1 {
		t.Fatalf("case-variant relink should dedup: edges=%d, want 1", got)
	}
}

// TestImportDeepCopiesTags proves Import does not alias the caller's Tags backing
// array (M3) — mutating the snapshot afterward cannot corrupt the stored node.
func TestImportDeepCopiesTags(t *testing.T) {
	snap := Snapshot{
		Nodes: []Node{{ID: "smithing", Kind: KindGoal, Label: "Train smithing", Status: StatusActive, Tags: []string{"core"}}},
	}
	g := New()
	g.Import(snap)
	// Mutate the caller's snapshot after import.
	snap.Nodes[0].Tags[0] = "MUTATED"
	n, _ := g.Get("smithing")
	if len(n.Tags) != 1 || n.Tags[0] != "core" {
		t.Fatalf("Import aliased Tags: stored=%v after caller mutation", n.Tags)
	}
	// Two graphs importing the same snapshot must not share backing arrays.
	snap2 := Snapshot{Nodes: []Node{{ID: "x", Tags: []string{"a"}}}}
	g1, g2 := New(), New()
	g1.Import(snap2)
	g2.Import(snap2)
	g1.Tag("x", "b")
	if n2, _ := g2.Get("x"); len(n2.Tags) != 1 {
		t.Fatalf("two graphs share Tags backing array: g2=%v", n2.Tags)
	}
}

// TestImportMergeAdditive guards H17 (host side): ImportMerge folds a snapshot
// into a NON-EMPTY live graph additively — it adds cron-only nodes/edges, keeps the
// host's live Status/Progress for a node present in both (does NOT adopt the
// snapshot's status), unions cron-added tags, never demotes a kind, and never
// deletes a host node the snapshot omits.
func TestImportMergeAdditive(t *testing.T) {
	g := New()
	// The live host working set the cron must not clobber.
	g.Upsert("mine_ore", KindGoal, "Mine ore", StatusActive)
	g.SetProgress("mine_ore", 0.7)
	g.Upsert("host_only", KindSubgoal, "host private step", StatusActive)

	// The server-reconciled snapshot: re-states the host's active goal as DONE (a
	// stale snapshot view) + adds a cron tag, plus a brand-new chain node + an edge.
	snap := Snapshot{
		Nodes: []Node{
			{ID: "mine_ore", Kind: KindOpenGoal, Label: "Mine ore (cron)", Status: StatusDone, Progress: 0.0, Tags: []string{"cron-chain"}},
			{ID: "rune_plate", Kind: KindOpenGoal, Label: "obtain rune platebody", Status: StatusOpen},
		},
		Edges: []Edge{{From: "mine_ore", To: "rune_plate", Rel: RelServes}},
	}
	g.ImportMerge(snap)

	// The host's live status/progress/kind WIN over the stale snapshot (authority
	// partition: host owns goal/status/progress).
	n, _ := g.Get("mine_ore")
	if n.Status != StatusActive {
		t.Fatalf("ImportMerge clobbered live status: %q, want active (host owns it)", n.Status)
	}
	if n.Progress != 0.7 {
		t.Fatalf("ImportMerge clobbered live progress: %v, want 0.7", n.Progress)
	}
	if n.Kind != KindGoal {
		t.Fatalf("ImportMerge demoted kind to %q, want goal (no demote)", n.Kind)
	}
	// But the cron-added TAG is adopted (union).
	if !sliceHas(n.Tags, "cron-chain") {
		t.Fatalf("ImportMerge dropped the cron-added tag: %v", n.Tags)
	}
	// The cron-only node + its edge land in the host's graph (survive the next flush).
	if !g.Has("rune_plate") {
		t.Fatal("ImportMerge did not add the cron-only chain node")
	}
	if out := g.Out("mine_ore", RelServes); len(out) != 1 || out[0].To != "rune_plate" {
		t.Fatalf("ImportMerge did not add the cron chain edge: %+v", out)
	}
	// The host-only node the snapshot omits is NEVER deleted.
	if !g.Has("host_only") {
		t.Fatal("ImportMerge deleted a host node absent from the snapshot")
	}
	// Re-merging the SAME snapshot is idempotent (dedup) — no edge/tag duplication.
	g.ImportMerge(snap)
	if got := len(g.Edges()); got != 1 {
		t.Fatalf("re-merge duplicated edges: %d, want 1", got)
	}
	if n, _ := g.Get("mine_ore"); len(n.Tags) != 1 {
		t.Fatalf("re-merge duplicated tags: %v", n.Tags)
	}
}

// TestImportMergeStaysBounded proves ImportMerge re-bounds via PruneTerminal so a
// re-import can never push the host past DefaultGraphCap (host-light invariant).
func TestImportMergeStaysBounded(t *testing.T) {
	g := New()
	tick := int64(100)
	g.now = func() int64 { tick++; return tick }
	g.Upsert("live", KindGoal, "live goal", StatusActive) // 1 non-terminal survivor
	// A snapshot of many TERMINAL cron nodes well over the cap.
	var nodes []Node
	for i := 0; i < DefaultGraphCap+30; i++ {
		nodes = append(nodes, Node{ID: "cron-" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			Kind: KindGoal, Label: "done", Status: StatusDone, At: int64(i)})
	}
	g.ImportMerge(Snapshot{Nodes: nodes})
	if got := len(g.Nodes()); got > DefaultGraphCap {
		t.Fatalf("ImportMerge left the graph unbounded: %d nodes, want <= %d", got, DefaultGraphCap)
	}
	if !g.Has("live") {
		t.Fatal("ImportMerge prune evicted the non-terminal live node")
	}
}

func sliceHas(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
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

// TestAspirations proves the aspiration accessor: only live KindAspiration
// nodes, in SEED order (oldest-first — genesis emits in priority order), with a
// retired aspiration dropped. Also proves the kind outranks goal (an aspiration
// can never be demoted by a re-upsert collision or a cron merge).
func TestAspirations(t *testing.T) {
	g := New()
	tick := int64(100)
	g.now = func() int64 { tick++; return tick }
	g.Upsert("aspiration:master-a-craft", KindAspiration, "master a craft", StatusActive)
	g.Upsert("aspiration:know-the-map", KindAspiration, "come to know the whole map", StatusActive)
	g.Upsert("mine ore", KindGoal, "Mine ore", StatusActive) // not an aspiration

	asps := g.Aspirations()
	if len(asps) != 2 {
		t.Fatalf("aspirations = %d, want 2", len(asps))
	}
	if asps[0].ID != "aspiration:master-a-craft" {
		t.Fatalf("seed order: got %q first, want master-a-craft", asps[0].ID)
	}
	// Kind never demotes: a goal-kind upsert on the same id keeps aspiration.
	g.Upsert("aspiration:master-a-craft", KindGoal, "", "")
	if n, _ := g.Get("aspiration:master-a-craft"); n.Kind != KindAspiration {
		t.Fatalf("aspiration demoted to %q by a weaker upsert", n.Kind)
	}
	// A retired aspiration leaves the live set.
	g.SetStatus("aspiration:know-the-map", StatusAbandoned)
	if got := len(g.Aspirations()); got != 1 {
		t.Fatalf("after retiring one, aspirations = %d, want 1", got)
	}
}

// TestRollup proves the aspiration rollup math: serving goal-kind nodes counted
// by status, state/question servers ignored, and LastTouched moving only on
// WORKED nodes (a queued open_goal does not touch the aspiration — adopting is
// not progress) while completing a goal does.
func TestRollup(t *testing.T) {
	g := New()
	tick := int64(100)
	g.now = func() int64 { tick++; return tick }
	g.Upsert("aspiration:master-a-craft", KindAspiration, "master a craft", StatusActive)
	aspAt := tick

	g.Upsert("smelt bars", KindGoal, "", StatusDone)
	g.Upsert("mine ore", KindGoal, "", StatusActive)
	g.Upsert("buy a hammer", KindOpenGoal, "", StatusOpen)
	g.Upsert("iron_ore", KindState, "", StatusOpen) // a dependency, not pursued work
	for _, from := range []string{"smelt bars", "mine ore", "buy a hammer", "iron_ore"} {
		g.Link(from, "aspiration:master-a-craft", RelServes)
	}

	r := g.Rollup("aspiration:master-a-craft")
	if r.Done != 1 || r.Working != 1 || r.Open != 1 {
		t.Fatalf("rollup = %+v, want done=1 working=1 open=1", r)
	}
	// LastTouched: the worked nodes were upserted AFTER the aspiration, so it
	// must have advanced past the aspiration's own At...
	if r.LastTouched <= aspAt {
		t.Fatalf("LastTouched %d should exceed the aspiration's At %d (worked servers touch it)", r.LastTouched, aspAt)
	}
	// ...but a LATER queued open_goal must NOT advance it further.
	before := r.LastTouched
	g.Upsert("learn fletching", KindOpenGoal, "", StatusOpen)
	g.Link("learn fletching", "aspiration:master-a-craft", RelServes)
	if got := g.Rollup("aspiration:master-a-craft").LastTouched; got != before {
		t.Fatalf("queuing an open_goal moved LastTouched %d → %d; adopting is not progress", before, got)
	}
	// Completing a serving goal DOES touch it.
	g.SetStatus("mine ore", StatusDone)
	r = g.Rollup("aspiration:master-a-craft")
	if r.Done != 2 || r.LastTouched <= before {
		t.Fatalf("after completion rollup = %+v (lastTouched before=%d), want done=2 and a fresher touch", r, before)
	}
	// An absent node rolls up to zero.
	if z := g.Rollup("aspiration:nope"); z != (ServesRollup{}) {
		t.Fatalf("absent aspiration rollup = %+v, want zero", z)
	}
}
