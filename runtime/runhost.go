package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/worldmap"
)

// HostConfig is the per-host knob surface — the cmd/host config struct minus the
// shared/process knobs (factsRoot, log level → SharedDeps) and with the secret
// resolved to a value by the caller. The mesa instance address is per-host (Mesa)
// so different hosts can target different mesa instances.
//
// SECURITY: Password is the resolved OpenRSC test secret. It is never printed,
// logged, marshalled, or echoed by RunHost. Callers must keep it out of logs too.
type HostConfig struct {
	Server   string // OpenRSC host:port (Options.Server)
	Username string // identity; data-dir/db/aliases/host-key leaf
	Password string // RESOLVED secret value (caller read env); never logged

	Goal     string       // -goal autonomous north-star
	Director DirectorSpec // routine/routines/loop/repl when not autonomous

	DataDir     string // "" => ~/.westworld/hosts/<Username>
	Fresh       bool   // ephemeral: in-memory store/aliases, no mesa mirror
	PersonaPath string // offline persona JSON; ignored when Mesa enabled

	Mesa     string // per-host mesa instance address (host:port); "" => offline
	Genesis  bool   // run login Genesis when mesa linked (default true)
	Headless bool   // no stdin: never drop to the REPL; error out instead (daemon hosts)

	TurnTimeout time.Duration // ConductorOptions.TurnTimeout (default 2m)
	Settle      time.Duration // ConductorOptions.Settle (default 500ms)

	Logger *slog.Logger // optional per-host child logger; falls back to deps.Logger
	Debug  *DebugSpec   // optional: nil = no per-host debug surface

	// OnReady, when set, is called once with the live host + conductor just before
	// the turn loop starts. The cradle uses it to register the running host for
	// control (pause/resume) and live introspection. nil in cmd/host.
	OnReady func(*HostHandle)
}

// HostHandle is the live control surface for a running host, handed to the caller
// via HostConfig.OnReady once the host is connected and its conductor is built.
// It exposes the live Host (world mirror, event bus), the Conductor (Pause /
// Resume / Store / Scratch), and the host's lifecycle context (cancelled when the
// host stops — bind per-host observers to it so they tear down cleanly).
type HostHandle struct {
	Host      *Host
	Conductor *Conductor
	Ctx       context.Context
}

// DirectorSpec carries the fixed-routine director knobs used when the host is
// not running the autonomous mesa Act planner.
type DirectorSpec struct {
	Routine  string
	Routines []string
	Loop     bool
	REPL     bool
}

// DebugSpec configures a per-host debug control plane. Addr "" => mount-only
// (a later step); a set Addr => own listening port (cmd/host's path).
//
// New constructs the debug surface over the live host once it exists. It is a
// factory hook so package runtime need not import debughttp (which imports
// runtime) — the caller (cmd/host) supplies the constructor. A nil New means no
// debug surface is started even when DebugSpec is non-nil.
type DebugSpec struct {
	Addr    string
	LogPath string
	New     func(host *Host) DebugSurface
}

// DebugSurface is the minimal lifecycle RunHost drives on a debug control plane:
// start recording bus events early (to capture login/initial snapshot), then
// serve until ctx is cancelled. *debughttp.Server satisfies it.
type DebugSurface interface {
	StartRecorder(ctx context.Context)
	Serve(ctx context.Context) error
}

// SharedDeps are the process-wide resources loaded ONCE by the parent and shared
// by pointer across every RunHost goroutine.
type SharedDeps struct {
	Facts     *facts.Facts        // load-once, share-by-pointer
	Landscape *pathfind.Landscape // load-once, share-by-pointer; PARENT owns Close()
	Mesa      MesaDialer          // nil => mesa unavailable; else dials one conn per host
	Logger    *slog.Logger        // process logger; RunHost derives child .With("host", Username)

	// WorldOracle is the precomputed static-geography engine (plane-0
	// walkability + capability-gated transport). Load-once, share-by-pointer:
	// it is immutable after worldmap.Precompute and every query passes the
	// host's Capability per-call, so a single instance is safe across all hosts
	// with no locking. Holds no fd, so closeDeps need not free it. nil =>
	// search_map / reachable / survey_map degrade to "no map data loaded".
	WorldOracle *worldmap.Oracle
}

// MesaDialer dials a mesa client for ONE host: its own gRPC connection to the
// mesa instance at addr, authenticated as hostID. There is no pooling — one
// connection per host. It returns the client plus a closer that the caller
// (RunHost) invokes when the host exits, so the connection lifecycle is per-host.
type MesaDialer func(addr, hostID string) (client mesaclient.Client, closeConn func() error, err error)
