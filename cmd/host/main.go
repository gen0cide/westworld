// Command host is a standalone runner for a single Westworld host.
//
// It launches one host outside the cradle: connect to an OpenRSC server, pump
// the world mirror, and either drive an autonomous turn loop (the conductor) or
// drop into an interactive REPL. This is the one-off, "just bring up a host"
// entry point — the cradle remains the path for running a configuration of many
// hosts at once.
//
// Examples:
//
//	# autonomous: run a fixed sequence of routines, then exit
//	host -username stubbs -routines tutorial/start.rt,tutorial/mine.rt
//
//	# autonomous: loop one routine until Ctrl-C
//	host -username stubbs -routine routines/wander.rt -loop
//
//	# interactive: REPL (default when no routine is given)
//	host -username stubbs
//
// The password comes from -password or the WESTWORLD_PASSWORD environment
// variable.
//
// cmd/host is a thin wrapper: it parses flags, resolves the password, owns
// process signals + world-data loading, and delegates the per-host lifecycle to
// runtime.RunHost. The reusable bring-up logic lives in package runtime.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/runtime"
)

type config struct {
	server   string
	username string
	password string

	factsRoot string
	dataDir   string

	mesa        string
	goal        string
	debugAddr   string
	fresh       bool
	personaPath string
	routine     string
	routines    string
	loop        bool
	turnTimeout time.Duration
	settle      time.Duration
	repl        bool

	logLevel string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.server, "server", "localhost:43594", "OpenRSC server host:port")
	flag.StringVar(&cfg.username, "username", "stubbs", "RSC account username")
	flag.StringVar(&cfg.password, "password", "", "RSC account password (or set WESTWORLD_PASSWORD)")
	flag.StringVar(&cfg.factsRoot, "facts", "/Users/flint/Code/openrsc", "OpenRSC source root for static facts; empty disables")
	flag.StringVar(&cfg.dataDir, "data-dir", "", "per-host writable data directory; defaults to ~/.westworld/hosts/<username>")
	flag.StringVar(&cfg.mesa, "mesa", "", "mesa service address (host:port); enables LLM cognition + persona sync over gRPC")
	flag.StringVar(&cfg.goal, "goal", "", "autonomous goal; with -mesa, runs the mesa Act planner (situation→DSL) toward this goal instead of a fixed -routine")
	flag.StringVar(&cfg.debugAddr, "debug-addr", "", "if set (e.g. localhost:8090), serve the debug control plane + live dashboard (GET / in a browser; /ws thought stream; /state; /events)")
	flag.BoolVar(&cfg.fresh, "fresh", false, "ephemeral mode: in-memory state only — no durable store, no learned aliases, no mesa memory mirror. Nothing persists between runs (use while debugging so memory isn't polluted).")
	flag.StringVar(&cfg.personaPath, "persona", "", "local persona JSON to compile (offline path; ignored when -mesa is set, which provisions the persona from mesa)")
	flag.StringVar(&cfg.routine, "routine", "", "a single routine file to run (looped with -loop)")
	flag.StringVar(&cfg.routines, "routines", "", "comma-separated routine files to run in sequence, then exit")
	flag.BoolVar(&cfg.loop, "loop", false, "with -routine, run it repeatedly until interrupted")
	flag.DurationVar(&cfg.turnTimeout, "turn-timeout", 2*time.Minute, "per-routine execution ceiling")
	flag.DurationVar(&cfg.settle, "settle", 500*time.Millisecond, "pause between conductor turns")
	flag.BoolVar(&cfg.repl, "repl", false, "force interactive REPL instead of the conductor")
	flag.StringVar(&cfg.logLevel, "log-level", "info", "log level: debug|info|warn|error")
	flag.Parse()

	log := newLogger(cfg.logLevel) // process logger

	if cfg.password == "" {
		cfg.password = os.Getenv("WESTWORLD_PASSWORD")
	}
	if cfg.password == "" {
		fmt.Fprintln(os.Stderr, "host: missing password — pass -password or set WESTWORLD_PASSWORD")
		os.Exit(2)
	}

	ctx, cancel := signalContext() // CLI owns signals + the single root ctx
	defer cancel()

	deps, closeDeps := buildSharedDeps(log, cfg) // loadWorld + the mesa dialer
	defer closeDeps()                            // closes the single shared Landscape (the parent owns it)

	hc := runtime.HostConfig{
		Server:   cfg.server,
		Username: cfg.username,
		Password: cfg.password,
		Goal:     cfg.goal,
		Director: runtime.DirectorSpec{
			Routine:  cfg.routine,
			Routines: splitCSV(cfg.routines),
			Loop:     cfg.loop,
			REPL:     cfg.repl,
		},
		DataDir:     cfg.dataDir,
		Fresh:       cfg.fresh,
		PersonaPath: cfg.personaPath,
		Mesa:        cfg.mesa,
		Genesis:     true,
		TurnTimeout: cfg.turnTimeout,
		Settle:      cfg.settle,
		Logger:      log,
	}
	if cfg.debugAddr != "" {
		hc.Debug = &runtime.DebugSpec{
			Addr: cfg.debugAddr,
			// cmd/host owns its own listening port via debughttp.Serve. The
			// factory keeps package runtime free of a debughttp import (debughttp
			// imports runtime). No debughttp mountability refactor here.
			New: func(host *runtime.Host) runtime.DebugSurface {
				return debughttp.New(host, debughttp.Config{Username: cfg.username, Addr: cfg.debugAddr}, log)
			},
		}
	}

	if err := runtime.RunHost(ctx, hc, deps); err != nil {
		log.Error("host exited with error", "err", err)
		os.Exit(1)
	}
}

// buildSharedDeps loads the process-wide world singletons and provides the mesa
// dialer RunHost uses to open each host's own connection. The returned closer
// releases the shared Landscape (the parent owns it; RunHost never closes it).
func buildSharedDeps(log *slog.Logger, cfg config) (runtime.SharedDeps, func()) {
	loadedFacts, loadedLandscape := loadWorld(log, cfg)
	deps := runtime.SharedDeps{
		Facts:     loadedFacts,
		Landscape: loadedLandscape,
		Mesa:      dialMesa,
		Logger:    log,
	}
	closeDeps := func() {
		if loadedLandscape != nil {
			loadedLandscape.Close()
		}
	}
	return deps, closeDeps
}

// dialMesa is the default mesa dialer: one gRPC connection per host, authenticated
// with the host's derived key (SHA-512 of its username, for now; mesa binds the
// resolved host_id into every call). It returns the client plus its Close so
// RunHost owns the per-host connection lifecycle.
func dialMesa(addr, hostID string) (mesaclient.Client, func() error, error) {
	c, err := mesaclient.NewGRPCClient(addr, mesaclient.HostKey(hostID))
	if err != nil {
		return nil, nil, err
	}
	return c, c.Close, nil
}

// splitCSV splits a comma-separated routine list into trimmed, non-empty paths.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// stripColorCodes removes RSC "@xxx@" colour codes and "%" line breaks.
func stripColorCodes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '@' && i+4 < len(s) && s[i+4] == '@' {
			i += 5
			continue
		}
		if s[i] == '%' {
			b.WriteByte(' ')
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// loadWorld loads the static facts catalog and landscape archive, mirroring the
// cradle. Both are optional — a load failure logs a warning and the host runs
// with reduced capability rather than failing to start.
func loadWorld(log *slog.Logger, cfg config) (*facts.Facts, *pathfind.Landscape) {
	if cfg.factsRoot == "" {
		return nil, nil
	}
	loadedFacts, err := facts.Load(facts.DefaultSources(cfg.factsRoot))
	if err != nil {
		log.Warn("facts load failed; continuing without world knowledge", "err", err)
		loadedFacts = nil
	} else {
		log.Info("loaded world facts", "summary", loadedFacts.Summary())
	}
	landscapePath := filepath.Join(cfg.factsRoot, "server", "conf", "server", "data", "Authentic_Landscape.orsc")
	loadedLandscape, err := pathfind.OpenLandscape(landscapePath)
	if err != nil {
		log.Warn("landscape load failed; pathfinding disabled", "err", err)
		return loadedFacts, nil
	}
	log.Info("loaded landscape archive", "path", landscapePath)
	return loadedFacts, loadedLandscape
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
	return ctx, cancel
}
