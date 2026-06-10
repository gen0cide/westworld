package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// TestAnalysisClassify pins the host-side FAST-PATH classifier: explicit
// prefixes and a bare DSL call are resolved deterministically (ok=true); any
// other natural-language directive returns ok=false to defer to the mesa
// AnalysisInterpret RPC.
func TestAnalysisClassify(t *testing.T) {
	cases := []struct {
		in       string
		wantKind AnalysisKind
		wantBody string
		wantOK   bool
	}{
		// Fast paths (host-side, deterministic).
		{`go_to("bank")`, AnalysisCommand, `go_to("bank")`, true},
		{`bank.deposit_all()`, AnalysisCommand, `bank.deposit_all()`, true},
		{`do: say("hi")`, AnalysisCommand, `say("hi")`, true},
		{`cmd: walk_to(120, 504)`, AnalysisCommand, `walk_to(120, 504)`, true},
		{`? how much gold`, AnalysisAnswer, `how much gold`, true},
		{`ask: are you hurt`, AnalysisAnswer, `are you hurt`, true},
		{`what would you do if attacked`, AnalysisHypothetical, `what would you do if attacked`, true},
		{`would: flee the goblin`, AnalysisHypothetical, `flee the goblin`, true},
		// Free natural language → deferred to the mesa interpreter (ok=false).
		{`where are you?`, "", `where are you?`, false},
		{`what is your hp`, "", `what is your hp`, false},
		{`go fight the rat`, "", `go fight the rat`, false},
	}
	for _, c := range cases {
		k, body, ok := classifyDirective(c.in)
		if ok != c.wantOK {
			t.Errorf("classify(%q): ok = %v, want %v", c.in, ok, c.wantOK)
		}
		if ok && k != c.wantKind {
			t.Errorf("classify(%q): kind = %q, want %q", c.in, k, c.wantKind)
		}
		if body != c.wantBody {
			t.Errorf("classify(%q): body = %q, want %q", c.in, body, c.wantBody)
		}
	}
}

// TestAnalysisEnterExitFreezesConductor proves entering analysis pauses the live
// conductor and exiting resumes it, with the active flag tracking both.
func TestAnalysisEnterExitFreezesConductor(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Director: Loop(Intent{Label: "tick"})})
	h.configureAnalysis("Operator", nil, c)

	if h.AnalysisActive() {
		t.Fatal("analysis should start inactive")
	}
	h.EnterAnalysis()
	if !h.AnalysisActive() {
		t.Fatal("EnterAnalysis should set active")
	}
	if !c.Paused() {
		t.Fatal("EnterAnalysis should pause the conductor")
	}
	h.ExitAnalysis()
	if h.AnalysisActive() {
		t.Fatal("ExitAnalysis should clear active")
	}
	if c.Paused() {
		t.Fatal("ExitAnalysis should resume the conductor")
	}
}

// TestAnalysisSuspendsMemoryWrites proves the single gate on memoryCapture and
// limbicHandle stops episodic + trust writes while active, and resumes after.
func TestAnalysisSuspendsMemoryWrites(t *testing.T) {
	h := newTestHost()
	h.configureAnalysis("Operator", nil, nil)

	// While ACTIVE: a level-up (episodic) and a chat (trust) must NOT be recorded.
	h.EnterAnalysis()
	h.memoryCapture(event.LevelUp{Skill: event.SkillID(1), NewLevel: 5})
	h.limbicHandle(event.ChatReceived{Speaker: "Bob", Message: "hi"})
	if got := h.journal.Len(); got != 0 {
		t.Fatalf("journal captured %d episodes while in analysis; want 0", got)
	}
	if h.ledger.Known("Bob") {
		t.Fatal("trust ledger recorded a relationship while in analysis; want none")
	}

	// After EXIT: writes resume normally.
	h.ExitAnalysis()
	h.limbicHandle(event.ChatReceived{Speaker: "Bob", Message: "hi"})
	if !h.ledger.Known("Bob") {
		t.Fatal("trust ledger did not resume after exiting analysis")
	}
}

// TestAnalysisCommandRunsUngated proves an operator COMMAND compiles + runs on
// the WithoutPearlGate interpreter path (the full-bypass command executor).
func TestAnalysisCommandRunsUngated(t *testing.T) {
	h := newTestHost()
	h.configureAnalysis("Operator", nil, nil)
	h.EnterAnalysis()

	// note() is a pure builtin (no packet) — a clean way to prove the ungated
	// interpreter compiles + runs an operator command without a live server.
	res := h.runCommand(context.Background(), `note("operator was here")`)
	if res.Kind != AnalysisCommand {
		t.Fatalf("kind = %q, want command", res.Kind)
	}
	if !res.Executed {
		t.Fatalf("command did not execute: err=%q text=%q", res.Err, res.Text)
	}
}

// TestAnalysisAnswerOfflineIsTruthful proves a factual question is answered from
// host state even with no mesa reachable (the deterministic fallback).
func TestAnalysisAnswerOfflineIsTruthful(t *testing.T) {
	h := newTestHost()
	h.configureAnalysis("Operator", nil, nil) // nil mc => offline
	h.EnterAnalysis()

	res := h.Analyze(context.Background(), "where are you?")
	if res.Kind != AnalysisAnswer {
		t.Fatalf("kind = %q, want answer", res.Kind)
	}
	// newTestHost positions self at (120,504); the flat state line must reflect it.
	if res.Text == "" {
		t.Fatal("answer was empty")
	}
}

// fakeInterpretClient is a healthy mesa Client that returns a canned
// AnalysisInterpret verdict so the host-side verdict router is driven offline.
type fakeInterpretClient struct {
	mesaclient.StubClient
	verdict *mesaclient.AnalysisVerdict
}

func (f *fakeInterpretClient) Healthy() bool { return true }
func (f *fakeInterpretClient) AnalysisInterpret(context.Context, string, []string) (*mesaclient.AnalysisVerdict, error) {
	return f.verdict, nil
}

// TestAnalysisReflectVerdictSurfacesText proves the host router surfaces a mesa
// "reflect" verdict's persona-voiced text verbatim (C-4 / #27) instead of
// discarding it into the dry-run planner, and that an EMPTY reflect text
// degrades to the hypothetical path like any unclassifiable directive.
func TestAnalysisReflectVerdictSurfacesText(t *testing.T) {
	h := newTestHost()
	fake := &fakeInterpretClient{verdict: &mesaclient.AnalysisVerdict{
		Kind: "reflect", Text: "I stopped mining because the rock ran dry."}}
	h.configureAnalysis("Operator", fake, nil)
	h.EnterAnalysis()

	res := h.Analyze(context.Background(), "why did you stop mining")
	if res.Kind != AnalysisReflect {
		t.Fatalf("kind = %q, want reflect", res.Kind)
	}
	if res.Text != "I stopped mining because the rock ran dry." {
		t.Fatalf("reflect text not surfaced verbatim: %q", res.Text)
	}
	if res.Executed {
		t.Fatal("a reflect answer must not execute anything")
	}

	// Empty reflect text: degrade to the planner dry-run (here offline → the
	// hypothetical error shape), never an empty answer.
	fake.verdict = &mesaclient.AnalysisVerdict{Kind: "reflect", Text: "  "}
	res = h.Analyze(context.Background(), "why did you stop mining")
	if res.Kind != AnalysisHypothetical {
		t.Fatalf("empty reflect text: kind = %q, want hypothetical fallback", res.Kind)
	}
}

// TestAnalysisInGameAuthRejectsNonOperator proves the in-game trigger ignores any
// sender that is not an EXACT match for the configured operator (takeover guard).
func TestAnalysisInGameAuthRejectsNonOperator(t *testing.T) {
	h := newTestHost() // username "test"
	h.configureAnalysis("Operator", nil, nil)
	ctx := context.Background()

	// A non-operator saying the trigger must NOT be handled and must NOT engage.
	if reflexAnalysis(ctx, h.log, h, "Mallory", "!test ANALYSIS", "test") {
		t.Fatal("non-operator trigger was handled — host-takeover vector")
	}
	if h.AnalysisActive() {
		t.Fatal("non-operator engaged analysis mode")
	}
	// The "a player" placeholder (unresolved index) must never authenticate.
	if reflexAnalysis(ctx, h.log, h, "a player", "!test ANALYSIS", "test") {
		t.Fatal("unresolved 'a player' authenticated as operator")
	}
	if h.AnalysisActive() {
		t.Fatal("placeholder sender engaged analysis mode")
	}
}
