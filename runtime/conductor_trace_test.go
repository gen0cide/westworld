package runtime

import "testing"

// TestConductorLineTrace verifies the per-statement line trace the cradle UI
// replays: every executed line is recorded in order, the sequence is monotonic,
// and the ring keeps only the last maxLineTrace lines (still contiguous + in
// order) once it overflows.
func TestConductorLineTrace(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Director: Sequence()})

	// The OnStmt hook installed by NewConductor feeds the trace.
	for i := 1; i <= 5; i++ {
		h.OnStmt(i)
	}
	trace, seq := c.LineTrace()
	if seq != 5 {
		t.Fatalf("seq=%d, want 5", seq)
	}
	if len(trace) != 5 {
		t.Fatalf("trace len=%d, want 5: %v", len(trace), trace)
	}
	for i, v := range trace {
		if v != i+1 {
			t.Fatalf("trace[%d]=%d, want %d (%v)", i, v, i+1, trace)
		}
	}

	// Overflow the ring: only the last maxLineTrace survive, in order; seq keeps
	// counting every statement.
	last := 5
	for i := 6; i <= 5+maxLineTrace+10; i++ {
		h.OnStmt(i)
		last = i
	}
	trace, seq = c.LineTrace()
	if seq != last {
		t.Fatalf("seq=%d, want %d", seq, last)
	}
	if len(trace) != maxLineTrace {
		t.Fatalf("trace len=%d, want %d (ring cap)", len(trace), maxLineTrace)
	}
	if trace[len(trace)-1] != last {
		t.Fatalf("trace tail=%d, want %d (latest line)", trace[len(trace)-1], last)
	}
	for i := 1; i < len(trace); i++ {
		if trace[i] != trace[i-1]+1 {
			t.Fatalf("trace not contiguous at %d: %v", i, trace)
		}
	}
}
