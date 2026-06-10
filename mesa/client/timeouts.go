package mesaclient

import (
	"context"
	"os"
	"time"
)

// Timeouts are the hard per-RPC deadlines the gRPC client imposes on every
// unary call (and the short client-streams Remember/RecordObservation). They
// exist because the conductor calls the planner with a deadline-free context
// (runtime/conductor.go passes the OUTER ctx to director.Next — the 2-minute
// turn budget only wraps executeRoutine), so a mesad that stops responding
// while the TCP connection stays up wedged a host's conductor for 34 minutes
// (drone7, soak 2026-06-10). With these in place an Act/Decide/Genesis RPC is
// STRUCTURALLY incapable of out-living its budget: the client cancels and the
// host falls back to local behavior.
//
// A caller-supplied earlier deadline still wins (context.WithTimeout never
// extends a parent deadline); these are ceilings, not floors. Zero/negative
// disables that ceiling (NOT recommended outside tests).
type Timeouts struct {
	// Act bounds the core "what do I do now?" planner call. It must fit inside
	// the conductor's 2-minute turn budget with room for the host to act on it.
	Act time.Duration
	// Decide bounds the narrow in-routine option pick (cheap tier, fast).
	Decide time.Duration
	// Genesis bounds the heavy Opus login compile (rare, allowed to be slow).
	Genesis time.Duration
	// Chat bounds the cheap social/reactive calls (Chat/Ask/AnalysisInterpret/
	// ExtractDialog).
	Chat time.Duration
	// Default bounds everything else (Provision, Recall, Remember, observations,
	// Sync*/Fetch*, metrics, KV) — plain storage round-trips.
	Default time.Duration
}

// DefaultTimeouts is the production ceiling set: Act 90s (inside the 2-min turn
// budget), Decide 30s, Genesis 120s, Chat 30s, everything else 30s.
func DefaultTimeouts() Timeouts {
	return Timeouts{
		Act:     90 * time.Second,
		Decide:  30 * time.Second,
		Genesis: 120 * time.Second,
		Chat:    30 * time.Second,
		Default: 30 * time.Second,
	}
}

// TimeoutsFromEnv returns base with any MESA_TIMEOUT_* environment override
// applied (MESA_TIMEOUT_ACT, MESA_TIMEOUT_DECIDE, MESA_TIMEOUT_GENESIS,
// MESA_TIMEOUT_CHAT, MESA_TIMEOUT_DEFAULT — Go duration strings, e.g. "45s").
// The env seam exists so an operator can retune a deployed host binary without
// a runtime/ code change (the runtime constructs the client; mesa/client owns
// the knob). A malformed value is ignored (the base default stands).
func TimeoutsFromEnv(base Timeouts) Timeouts {
	if d, ok := envDuration("MESA_TIMEOUT_ACT"); ok {
		base.Act = d
	}
	if d, ok := envDuration("MESA_TIMEOUT_DECIDE"); ok {
		base.Decide = d
	}
	if d, ok := envDuration("MESA_TIMEOUT_GENESIS"); ok {
		base.Genesis = d
	}
	if d, ok := envDuration("MESA_TIMEOUT_CHAT"); ok {
		base.Chat = d
	}
	if d, ok := envDuration("MESA_TIMEOUT_DEFAULT"); ok {
		base.Default = d
	}
	return base
}

func envDuration(key string) (time.Duration, bool) {
	v := os.Getenv(key)
	if v == "" {
		return 0, false
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, false
	}
	return d, true
}

// withTimeout caps ctx at d (a no-op when d <= 0). The returned cancel must
// always be called. A parent deadline earlier than now+d is preserved —
// context.WithTimeout can only tighten, never extend.
func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}
