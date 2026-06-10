package cradle

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cradle/hostcfg"
	"github.com/gen0cide/westworld/runtime"
)

const testPWEnv = "WW_TEST_PW_REGISTRY"

// blockingRun is a fake runFunc that reports ready then blocks until ctx cancel.
func blockingRun(handle *runtime.HostHandle) runFunc {
	return func(ctx context.Context, cfg runtime.HostConfig, _ runtime.SharedDeps) error {
		if cfg.OnReady != nil {
			cfg.OnReady(handle)
		}
		<-ctx.Done()
		return ctx.Err()
	}
}

func waitStatus(t *testing.T, r *Registry, name string, want Status, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s, err := r.Get(name); err == nil && s.Status == want {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	s, _ := r.Get(name)
	t.Fatalf("host %q: want status %q, got %q", name, want, s.Status)
}

func testSpec(name string) hostcfg.Host {
	return hostcfg.Host{Name: name, Goal: "g", Mesa: "m", PasswordEnv: testPWEnv}
}

func TestRegistryStartStop(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	r.run = blockingRun(&runtime.HostHandle{})

	if err := r.Start(context.Background(), testSpec("a")); err != nil {
		t.Fatalf("start: %v", err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)

	if got := r.List(); len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("list wrong: %+v", got)
	}
	if err := r.Stop("a"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	waitStatus(t, r, "a", StatusStopped, 2*time.Second)
	r.Wait()
}

func TestRegistryDuplicate(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	r.run = blockingRun(&runtime.HostHandle{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx, testSpec("a")); err != nil {
		t.Fatal(err)
	}
	if err := r.Start(ctx, testSpec("a")); err == nil {
		t.Fatal("expected duplicate-name error")
	}
	cancel()
	r.Wait()
}

func TestRegistryStartFailsFast(t *testing.T) {
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	r.run = blockingRun(&runtime.HostHandle{})

	// Missing password env -> fail before any goroutine spins.
	if err := r.Start(context.Background(), hostcfg.Host{Name: "a", Goal: "g", Mesa: "m", PasswordEnv: "WW_DEFINITELY_UNSET"}); err == nil {
		t.Fatal("expected missing-password error")
	}
	// Invalid spec (goal without mesa).
	t.Setenv(testPWEnv, "x")
	if err := r.Start(context.Background(), hostcfg.Host{Name: "b", Goal: "g", PasswordEnv: testPWEnv}); err == nil {
		t.Fatal("expected validation error")
	}
	if len(r.List()) != 0 {
		t.Fatal("failed starts must not register a host")
	}
}

func TestRegistrySupervisionRestart(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	r.backoffBase = 3 * time.Millisecond
	var calls int32
	r.run = func(_ context.Context, _ runtime.HostConfig, _ runtime.SharedDeps) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("boom") // crash immediately
	}
	spec := testSpec("a")
	spec.Supervision = hostcfg.SuperviseRestart
	if err := r.Start(context.Background(), spec); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&calls) < 3 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	if c := atomic.LoadInt32(&calls); c < 3 {
		t.Fatalf("expected >=3 run attempts under restart policy, got %d", c)
	}
	if s, _ := r.Get("a"); s.Restarts < 2 {
		t.Fatalf("restart count not tracked: %d", s.Restarts)
	}
	if err := r.Stop("a"); err != nil {
		t.Fatal(err)
	}
	r.Wait()
}

func TestRegistrySupervisionHold(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	var calls int32
	r.run = func(_ context.Context, _ runtime.HostConfig, _ runtime.SharedDeps) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("boom")
	}
	spec := testSpec("a")
	spec.Supervision = hostcfg.SuperviseHold
	if err := r.Start(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusCrashed, 2*time.Second)
	time.Sleep(40 * time.Millisecond)
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("hold policy must not restart; got %d attempts", c)
	}
	r.Wait()
}

func TestRegistryPauseResume(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	conductor := runtime.NewConductor(nil, runtime.ConductorOptions{Director: runtime.Sequence()})
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	r.run = blockingRun(&runtime.HostHandle{Conductor: conductor})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx, testSpec("a")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)

	if err := r.Pause("a"); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if !conductor.Paused() {
		t.Fatal("conductor should be paused after registry Pause")
	}
	if s, _ := r.Get("a"); s.Status != StatusPaused {
		t.Fatalf("status after pause: %q", s.Status)
	}
	if err := r.Resume("a"); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if conductor.Paused() {
		t.Fatal("conductor should be running after registry Resume")
	}
	cancel()
	r.Wait()
}

// countingRun is blockingRun plus an invocation counter, for restart tests.
func countingRun(handle *runtime.HostHandle, n *atomic.Int32) runFunc {
	return func(ctx context.Context, cfg runtime.HostConfig, _ runtime.SharedDeps) error {
		n.Add(1)
		if cfg.OnReady != nil {
			cfg.OnReady(handle)
		}
		<-ctx.Done()
		return ctx.Err()
	}
}

// TestRegistryRestartRunning: bouncing a live host relaunches it immediately —
// no crash-backoff, no restart-count tick — and it comes back running.
func TestRegistryRestartRunning(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	var runs atomic.Int32
	r.run = countingRun(&runtime.HostHandle{}, &runs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx, testSpec("a")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)

	if err := r.Restart("a"); err != nil {
		t.Fatalf("restart: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for runs.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	if got := runs.Load(); got != 2 {
		t.Fatalf("run invocations = %d, want 2 (bounce relaunches)", got)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)
	if s, _ := r.Get("a"); s.Restarts != 0 {
		t.Fatalf("restarts = %d, want 0 (operator bounce is not a crash)", s.Restarts)
	}
	cancel()
	r.Wait()
}

// TestRegistryRestartStopped: Restart doubles as start — a stopped host (dead
// supervisor) gets a fresh one and runs again.
func TestRegistryRestartStopped(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	var runs atomic.Int32
	r.run = countingRun(&runtime.HostHandle{}, &runs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx, testSpec("a")); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)
	if err := r.Stop("a"); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusStopped, 2*time.Second)

	if err := r.Restart("a"); err != nil {
		t.Fatalf("restart stopped host: %v", err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)
	if got := runs.Load(); got != 2 {
		t.Fatalf("run invocations = %d, want 2", got)
	}
	cancel()
	r.Wait()
}

// TestRegistryRestartCrashedHold: a supervision=hold host that crashed is held;
// Restart relaunches it anyway (the operator override beats the policy).
func TestRegistryRestartCrashedHold(t *testing.T) {
	t.Setenv(testPWEnv, "x")
	r := NewRegistry(runtime.SharedDeps{}, slog.Default())
	var runs atomic.Int32
	boom := errors.New("boom")
	r.run = func(ctx context.Context, cfg runtime.HostConfig, _ runtime.SharedDeps) error {
		if runs.Add(1) == 1 {
			return boom // first run crashes immediately
		}
		if cfg.OnReady != nil {
			cfg.OnReady(&runtime.HostHandle{})
		}
		<-ctx.Done()
		return ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spec := testSpec("a")
	spec.Supervision = hostcfg.SuperviseHold
	if err := r.Start(ctx, spec); err != nil {
		t.Fatal(err)
	}
	waitStatus(t, r, "a", StatusCrashed, 2*time.Second)

	if err := r.Restart("a"); err != nil {
		t.Fatalf("restart crashed host: %v", err)
	}
	waitStatus(t, r, "a", StatusRunning, 2*time.Second)
	cancel()
	r.Wait()
}

// TestToHostConfigMapping carries the mapping assertions that lived in
// hostcfg_test.go before the runtime.HostConfig mapping moved here (CP-1:
// hostcfg is a yaml+stdlib leaf now; the mapping is cradle's).
func TestToHostConfigMapping(t *testing.T) {
	tt := 90 * time.Second
	h := hostcfg.Host{
		Name: "stubbs", Mesa: "localhost:7077", Goal: "Finish tutorial island",
		State: hostcfg.StateFile, TurnTimeout: hostcfg.Duration(tt),
	}
	hc := toHostConfig(h, "secret")
	if hc.Username != "stubbs" || hc.Mesa != "localhost:7077" || hc.Goal != "Finish tutorial island" {
		t.Fatalf("mapping wrong: %+v", hc)
	}
	if hc.Fresh {
		t.Fatal("state:file should not be Fresh")
	}
	if !hc.Genesis {
		t.Fatal("Genesis should default true")
	}
	if hc.TurnTimeout != tt {
		t.Fatalf("TurnTimeout not mapped: %v", hc.TurnTimeout)
	}
	if hc.Server != hostcfg.DefaultServer {
		t.Fatalf("server default not applied: %q", hc.Server)
	}

	mem := hostcfg.Host{Name: "bernard", Mesa: "localhost:7077", State: hostcfg.StateMemory}
	if mc := toHostConfig(mem, "secret"); !mc.Fresh {
		t.Fatal("state:memory should map to Fresh")
	}
}
