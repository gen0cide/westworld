package mesad

import (
	"context"
	"testing"
	"time"
)

// TestEnsureDeadlineAddsCeiling proves the server-side backstop: a deadline-less
// inbound ctx (a legacy client) gets capped, so an LLM-backed handler goroutine
// can never out-live its budget on a wedged upstream.
func TestEnsureDeadlineAddsCeiling(t *testing.T) {
	ctx, cancel := ensureDeadline(context.Background(), 5*time.Second)
	defer cancel()
	dl, ok := ctx.Deadline()
	if !ok {
		t.Fatal("ensureDeadline left a deadline-less ctx unbounded")
	}
	if until := time.Until(dl); until > 5*time.Second || until <= 0 {
		t.Fatalf("ceiling deadline %v from now; want within (0, 5s]", until)
	}
}

// TestEnsureDeadlineRespectsClient proves a client-supplied deadline wins: the
// backstop must not extend (or replace) a propagated gRPC deadline.
func TestEnsureDeadlineRespectsClient(t *testing.T) {
	parent, pcancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer pcancel()
	ctx, cancel := ensureDeadline(parent, 5*time.Second)
	defer cancel()
	dl, ok := ctx.Deadline()
	if !ok {
		t.Fatal("deadline lost")
	}
	if time.Until(dl) > 200*time.Millisecond {
		t.Fatalf("backstop extended the client's 100ms deadline to %v", time.Until(dl))
	}
}

// TestEnsureDeadlineDisabled: d <= 0 is a no-op (tests/dev escape hatch).
func TestEnsureDeadlineDisabled(t *testing.T) {
	ctx, cancel := ensureDeadline(context.Background(), 0)
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("d=0 should leave ctx unbounded")
	}
}
