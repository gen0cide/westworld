package mesaclient

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// slowGame is a Game server whose Act/Decide wedge until the inbound ctx dies —
// the wedged-mesad shape that hung drone7's conductor for 34 minutes. It proves
// the client's per-RPC ceiling fires even when the server never answers.
type slowGame struct {
	mesapb.UnimplementedGameServer
}

func (slowGame) Act(ctx context.Context, _ *mesapb.Situation) (*mesapb.Move, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second): // far past any test deadline — never reached
		return &mesapb.Move{Kind: mesapb.MoveKind_IDLE}, nil
	}
}

func (slowGame) Decide(ctx context.Context, _ *mesapb.Choice) (*mesapb.Decision, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return &mesapb.Decision{Choice: "a"}, nil
	}
}

// startSlowServer serves slowGame on a loopback port and returns its address.
func startSlowServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	gs := grpc.NewServer()
	mesapb.RegisterGameServer(gs, slowGame{})
	go gs.Serve(lis) //nolint:errcheck
	t.Cleanup(gs.Stop)
	return lis.Addr().String()
}

// TestActDeadlineFires proves the structural hang-proofing: a deadline-FREE
// caller context (exactly what runtime/conductor passes to director.Next) +
// a server that never answers must yield DeadlineExceeded promptly — not a
// 34-minute wedge.
func TestActDeadlineFires(t *testing.T) {
	addr := startSlowServer(t)
	c, err := NewGRPCClientWithTimeouts(addr, "tok", Timeouts{Act: 200 * time.Millisecond})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer c.Close()

	start := time.Now()
	_, err = c.Act(context.Background(), &Situation{HostID: "h", Goal: "g"})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Act against a wedged server returned nil error; want DeadlineExceeded")
	}
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("Act err = %v (code %s), want DeadlineExceeded", err, status.Code(err))
	}
	if elapsed > 3*time.Second {
		t.Fatalf("Act took %v to fail; the deadline did not fire promptly", elapsed)
	}
}

// TestDecideDeadlineFires is the same proof for the Decide ceiling.
func TestDecideDeadlineFires(t *testing.T) {
	addr := startSlowServer(t)
	c, err := NewGRPCClientWithTimeouts(addr, "tok", Timeouts{Decide: 200 * time.Millisecond})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer c.Close()

	start := time.Now()
	_, err = c.Decide(context.Background(), &Choice{HostID: "h", Question: "q", Options: []string{"a"}})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("Decide err = %v (code %s), want DeadlineExceeded", err, status.Code(err))
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("Decide took %v to fail; the deadline did not fire promptly", elapsed)
	}
}

// TestCallerDeadlineStillWins proves the ceiling only tightens: a caller whose
// own deadline is EARLIER than the configured ceiling keeps it.
func TestCallerDeadlineStillWins(t *testing.T) {
	addr := startSlowServer(t)
	c, err := NewGRPCClientWithTimeouts(addr, "tok", Timeouts{Act: 10 * time.Second})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = c.Act(ctx, &Situation{HostID: "h"})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("Act err = %v (code %s), want DeadlineExceeded", err, status.Code(err))
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("caller's 150ms deadline was extended to %v by the 10s ceiling", elapsed)
	}
}

// TestDefaultTimeoutsAllSet guards the production defaults: every RPC family
// must carry a positive ceiling (a zero would silently disable hang-proofing),
// and Act must fit inside the conductor's 2-minute turn budget.
func TestDefaultTimeoutsAllSet(t *testing.T) {
	d := DefaultTimeouts()
	for name, v := range map[string]time.Duration{
		"Act": d.Act, "Decide": d.Decide, "Genesis": d.Genesis, "Chat": d.Chat, "Default": d.Default,
	} {
		if v <= 0 {
			t.Errorf("DefaultTimeouts().%s = %v; every family needs a positive ceiling", name, v)
		}
	}
	if d.Act >= 2*time.Minute {
		t.Errorf("DefaultTimeouts().Act = %v; must fit inside the conductor's 2-minute turn budget", d.Act)
	}
}

// TestTimeoutsFromEnv proves the operator env seam (no runtime/ knob needed).
func TestTimeoutsFromEnv(t *testing.T) {
	t.Setenv("MESA_TIMEOUT_ACT", "42s")
	t.Setenv("MESA_TIMEOUT_DECIDE", "bogus") // malformed → default stands
	got := TimeoutsFromEnv(DefaultTimeouts())
	if got.Act != 42*time.Second {
		t.Errorf("MESA_TIMEOUT_ACT override: Act = %v, want 42s", got.Act)
	}
	if got.Decide != DefaultTimeouts().Decide {
		t.Errorf("malformed MESA_TIMEOUT_DECIDE changed Decide to %v; want the default %v", got.Decide, DefaultTimeouts().Decide)
	}
}
