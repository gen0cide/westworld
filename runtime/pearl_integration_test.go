package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/world"
)

// setHP overwrites the test host's Hits skill (id 3) so HP-based predicates can
// be exercised. Leaves the other skills as newTestHost set them.
func setHP(h *Host, cur, max int) {
	var c, m, x [world.NumSkills]int
	for i := range c {
		c[i], m[i] = 30, 30
	}
	c[3], m[3] = cur, max
	h.world.Self.SetAllSkills(c, m, x, 0)
}

// TestPearlDecideFastPathBypassesStrategist proves that when the engine answers
// a decide() locally, the LLM Strategist is never consulted: we set Strategist
// to nil, so a miss would surface NOT_IMPLEMENTED. A "flee" result instead means
// the pearl fast path returned without escalating.
func TestPearlDecideFastPathBypassesStrategist(t *testing.T) {
	h := newTestHost()
	h.Strategist = nil // any escalation would fail loudly
	h.Pearl = pearl.New(pearl.DemoTable(), 0)
	setHP(h, 2, 10) // 20% HP ⇒ flee_when_hurt fires

	res := runRoutine(t, h, `runtime "1.0"
routine r() { return decide(["fight", "flee"]).val }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("routine kind = %v (err=%v), want returned", res.Kind, res.Err)
	}
	if got := res.Value.Display(); got != "flee" {
		t.Fatalf("decide returned %q, want flee (pearl fast path)", got)
	}
}

// TestPearlDecideMissEscalates confirms the seam falls through to the strategist
// on a miss (healthy host ⇒ no confident rule). The stub strategist returns the
// first option by default, so we just assert the call path didn't hit the nil
// guard and produced a value.
func TestPearlDecideMissEscalatesToStrategist(t *testing.T) {
	h := newTestHost()
	h.Pearl = pearl.New(pearl.DemoTable(), 0)
	setHP(h, 10, 10) // full HP ⇒ no flee bias ⇒ miss ⇒ strategist (stub) decides

	res := runRoutine(t, h, `runtime "1.0"
routine r() { return decide(["fight", "flee"]).val }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("routine kind = %v (err=%v), want returned", res.Kind, res.Err)
	}
	if got := res.Value.Display(); got != "fight" && got != "flee" {
		t.Fatalf("decide returned %q, want a real option from the strategist", got)
	}
}

// TestPearlGateVetoesAction proves the action seam: an engine whose table vetoes
// `say` causes say() to return a POLICY_VETO failure WITHOUT the inner handler
// running (which, with no connection, would otherwise fail trying to send).
func TestPearlGateVetoesAction(t *testing.T) {
	h := newTestHost()
	h.Pearl = pearl.New(pearl.Table{Rules: []pearl.Rule{{
		ID:         "no_chatter",
		When:       pearl.OnAction("say"),
		Then:       pearl.Effect{Kind: pearl.EffectVeto, Reason: "this host stays quiet"},
		Confidence: 1,
	}}}, 0)

	res := runRoutine(t, h, `runtime "1.0"
routine r() {
    result = say("hello")
    return result.err.code
}`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("routine kind = %v (err=%v), want returned", res.Kind, res.Err)
	}
	if got := res.Value.Display(); got != "POLICY_VETO" {
		t.Fatalf("say().err.code = %q, want POLICY_VETO", got)
	}
}

// TestPearlGateAllowsWhenNoRuleFires confirms a non-matching table is a pure
// passthrough. We veto a DIFFERENT action ("attack"), so say() is allowed
// through to its real handler — which, with no live connection, panics on the
// nil socket (contained as ResultErrored). That error is the proof the gate did
// NOT short-circuit say(); a veto would instead have returned cleanly.
func TestPearlGateAllowsWhenNoRuleFires(t *testing.T) {
	h := newTestHost()
	h.Pearl = pearl.New(pearl.Table{Rules: []pearl.Rule{{
		ID:         "no_attack",
		When:       pearl.OnAction("attack"),
		Then:       pearl.Effect{Kind: pearl.EffectVeto},
		Confidence: 1,
	}}}, 0)

	res := runRoutine(t, h, `runtime "1.0"
routine r() { return say("hello") }`)
	if res.Kind != interp.ResultErrored {
		t.Fatalf("routine kind = %v, want errored (say reached the real handler, proving passthrough)", res.Kind)
	}
}

// TestPearlNilIsNoop confirms the zero-config path: with no engine, decide and
// say behave exactly as before (no gating). decide falls to the stub strategist.
func TestPearlNilIsNoop(t *testing.T) {
	h := newTestHost() // h.Pearl is nil
	res := runRoutine(t, h, `runtime "1.0"
routine r() { return decide(["a", "b"]).val }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("routine kind = %v (err=%v), want returned", res.Kind, res.Err)
	}
}
