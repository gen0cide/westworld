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
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/persona"
	"github.com/gen0cide/westworld/runtime"
)

type config struct {
	server   string
	username string
	password string

	factsRoot string
	dataDir   string

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
	flag.StringVar(&cfg.personaPath, "persona", "", "persona JSON to compile into the host's policy (pearl table + affect baseline)")
	flag.StringVar(&cfg.routine, "routine", "", "a single routine file to run (looped with -loop)")
	flag.StringVar(&cfg.routines, "routines", "", "comma-separated routine files to run in sequence, then exit")
	flag.BoolVar(&cfg.loop, "loop", false, "with -routine, run it repeatedly until interrupted")
	flag.DurationVar(&cfg.turnTimeout, "turn-timeout", 2*time.Minute, "per-routine execution ceiling")
	flag.DurationVar(&cfg.settle, "settle", 500*time.Millisecond, "pause between conductor turns")
	flag.BoolVar(&cfg.repl, "repl", false, "force interactive REPL instead of the conductor")
	flag.StringVar(&cfg.logLevel, "log-level", "info", "log level: debug|info|warn|error")
	flag.Parse()

	log := newLogger(cfg.logLevel)

	if cfg.password == "" {
		cfg.password = os.Getenv("WESTWORLD_PASSWORD")
	}
	if cfg.password == "" {
		fmt.Fprintln(os.Stderr, "host: missing password — pass -password or set WESTWORLD_PASSWORD")
		os.Exit(2)
	}

	if err := run(log, cfg); err != nil {
		log.Error("host exited with error", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, cfg config) error {
	rootCtx, cancel := signalContext()
	defer cancel()

	loadedFacts, loadedLandscape := loadWorld(log, cfg)
	if loadedLandscape != nil {
		defer loadedLandscape.Close()
	}

	host := runtime.New(runtime.Options{
		Server:    cfg.server,
		Username:  cfg.username,
		Password:  cfg.password,
		Facts:     loadedFacts,
		Landscape: loadedLandscape,
		Logger:    log,
	})
	defer host.Close()

	// Per-host writable state: the learned-alias store (recognition) and the
	// durable hostkv store (conductor progress, future trust ledger).
	dataDir := resolveDataDir(log, cfg)
	host.Resolver = resolve.New(loadedFacts, loadAliasStore(log, dataDir), nil)
	store := openLocalStore(log, dataDir, cfg.username)
	if store != nil {
		defer store.Close()
	}

	// Local tiers shared by the conductor and the memory manager: one scratch
	// (ephemeral) + one durable store. If no data dir, fall back to in-memory so
	// the memory verbs still work (just not persisted).
	scratch := hostkv.NewScratch(256)
	local := store
	if local == nil {
		local = hostkv.NewMemory()
	}
	// Tiered memory: remote (mesa) is deferred, so it runs offline (NopRemote) —
	// everything served from / kept in the local tiers, remote writes journaled.
	host.Memory = memory.New(memory.Options{Scratch: scratch, Local: local})

	// Persona: compile disposition → the pearl policy (the deterministic half of
	// the compiler; mesa would do this, but it's pure Go so the host can too) and
	// set the affect baseline. This is what makes the host BEHAVE per-persona —
	// the pearl gate + decide seam now fire its rules.
	if cfg.personaPath != "" {
		if err := loadPersona(log, host, cfg.personaPath); err != nil {
			return fmt.Errorf("persona: %w", err)
		}
	}

	log.Info("connecting", "server", cfg.server, "username", cfg.username)
	if err := host.Connect(rootCtx); err != nil {
		return err
	}

	// Frame pump on its own goroutine — keeps the world mirror live. Must be
	// running before the conductor drives routines that wait on world changes.
	hostDone := make(chan error, 1)
	go func() { hostDone <- host.Run(rootCtx) }()

	// Give the mirror a beat to fill from the initial login burst.
	select {
	case <-rootCtx.Done():
		return waitHost(rootCtx, hostDone)
	case <-time.After(time.Second):
	}

	director := buildDirector(cfg)
	if cfg.repl || director == nil {
		log.Info("entering REPL")
		r := host.NewREPL(rootCtx, os.Stdin, os.Stdout)
		if err := r.Run(); err != nil {
			log.Warn("repl exited", "err", err)
		}
		cancel()
		return waitHost(rootCtx, hostDone)
	}

	conductor := runtime.NewConductor(host, runtime.ConductorOptions{
		Director:    director,
		TurnTimeout: cfg.turnTimeout,
		Settle:      cfg.settle,
		Store:       local,
		Scratch:     scratch,
		Logger:      log,
	})
	log.Info("starting conductor")
	cerr := conductor.Run(rootCtx)
	cancel() // conductor finished or was cancelled — stop the frame pump
	herr := waitHost(rootCtx, hostDone)
	if cerr != nil && cerr != context.Canceled {
		return cerr
	}
	return herr
}

// buildDirector turns the routine flags into a Director, or returns nil when no
// routines were specified (caller falls back to the REPL).
func buildDirector(cfg config) runtime.Director {
	if cfg.routines != "" {
		var intents []runtime.Intent
		for _, p := range strings.Split(cfg.routines, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			intents = append(intents, runtime.Intent{Label: filepath.Base(p), RoutinePath: p})
		}
		if len(intents) > 0 {
			return runtime.Sequence(intents...)
		}
	}
	if cfg.routine != "" {
		in := runtime.Intent{Label: filepath.Base(cfg.routine), RoutinePath: cfg.routine}
		if cfg.loop {
			return runtime.Loop(in)
		}
		return runtime.Sequence(in)
	}
	return nil
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

// loadPersona reads + validates a persona, compiles its policy, and wires it
// onto the host: the pearl table (gate + decide seam), the decision floor, and
// the affect baseline. Compilation is deterministic (mesa would normally do it).
func loadPersona(log *slog.Logger, host *runtime.Host, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var p persona.Persona
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid persona %s: %w", path, err)
	}
	cp := persona.CompilePolicy(&p)
	host.Pearl = pearl.New(cp.Table, cp.DecisionFloor)
	m := p.Trajectory.Mood
	host.SetAffectBaseline(m.Stress, m.Confidence, m.Valence)
	log.Info("loaded persona",
		"name", p.Cornerstone.Identity.Name,
		"archetype", p.Cornerstone.Identity.ArchetypeTag,
		"pearl_rules", len(cp.Table.Rules),
		"decision_floor", cp.DecisionFloor,
		"fairness", cp.Trade.FairnessThreshold,
		"scam_propensity", cp.Trade.ScamPropensity,
		"risk_aversion", cp.Trade.RiskAversion,
	)
	for _, r := range cp.Table.Rules {
		log.Info("  pearl rule", "id", r.ID, "origin", r.Origin)
	}
	return nil
}

// loadAliasStore opens the per-host JSON alias store, or returns nil (in-memory
// recognition) when no data dir or on load error.
func loadAliasStore(log *slog.Logger, dataDir string) *resolve.AliasStore {
	if dataDir == "" {
		return nil
	}
	path := filepath.Join(dataDir, "aliases.json")
	store, err := resolve.LoadAliasStore(path)
	if err != nil {
		log.Warn("alias store load failed; recognition will not persist", "path", path, "err", err)
		return nil
	}
	log.Info("loaded learned-alias store", "path", path, "aliases", store.Len())
	return store
}

// openLocalStore opens the durable hostkv store under the data dir, named after
// the host so the file is self-identifying (e.g. stubbs.db). Returns nil (the
// conductor falls back to in-memory) when no data dir or on error.
func openLocalStore(log *slog.Logger, dataDir, username string) *hostkv.Store {
	if dataDir == "" {
		return nil
	}
	name := username
	if name == "" {
		name = "host"
	}
	path := filepath.Join(dataDir, name+".db")
	store, err := hostkv.Open(path)
	if err != nil {
		log.Warn("local store open failed; host state will not persist", "path", path, "err", err)
		return nil
	}
	log.Info("opened local store", "path", path, "entries", store.Len())
	return store
}

// waitHost waits for the frame-pump goroutine to finish, mapping clean
// cancellation to nil.
func waitHost(ctx context.Context, done <-chan error) error {
	err := <-done
	if err == nil || err == context.Canceled || err == ctx.Err() {
		return nil
	}
	return err
}

// resolveDataDir mirrors the cradle: -data-dir, else ~/.westworld/hosts/<user>.
func resolveDataDir(log *slog.Logger, cfg config) string {
	if cfg.dataDir != "" {
		return cfg.dataDir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		log.Warn("no -data-dir and home dir unavailable; host state will not persist", "err", err)
		return ""
	}
	name := cfg.username
	if name == "" {
		name = "anon"
	}
	return filepath.Join(home, ".westworld", "hosts", name)
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
