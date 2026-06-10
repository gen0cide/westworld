package runtime

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// dyingSubClient's Subscribe stream dies instantly — the shape a mesad restart
// produces on every connected host.
type dyingSubClient struct {
	mesaclient.StubClient
	calls atomic.Int32
}

func (c *dyingSubClient) Subscribe(context.Context, string) (<-chan mesaclient.Directive, error) {
	c.calls.Add(1)
	ch := make(chan mesaclient.Directive)
	close(ch)
	return ch, nil
}

func TestSubscribeDirectivesResubscribesAfterStreamDeath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &dyingSubClient{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	done := make(chan struct{})
	go func() { subscribeDirectives(ctx, log, &Host{}, c, "t"); close(done) }()

	deadline := time.After(3 * time.Second)
	for c.calls.Load() < 2 {
		select {
		case <-deadline:
			t.Fatalf("no re-subscribe after stream death: %d subscribe call(s)", c.calls.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("subscribeDirectives did not exit on ctx cancel")
	}
}
