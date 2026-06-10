package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// RunHost constructs, connects, and drives ONE host to completion using shared
// process resources. It owns the per-host lifecycle (host, store, mesa client,
// conductor, goroutines) but NOT signals or world-data loading — the caller owns
// ctx and supplies the shared deps. Returns when ctx is cancelled, the conductor
// finishes, or a fatal bring-up error occurs.
//
// The shared singletons in deps (Facts, Landscape) are owned by the caller:
// RunHost never closes the Landscape. The mesa connection, by contrast, is
// per-host — RunHost dials it (via deps.Mesa) and closes it on exit.
func RunHost(ctx context.Context, cfg HostConfig, deps SharedDeps) error {
	log := cfg.Logger
	if log == nil {
		log = deps.Logger
	}
	if log != nil {
		log = log.With("host", cfg.Username)
	} else {
		log = slog.Default()
	}

	// Per-host child ctx: all four per-host goroutines (socialReflex,
	// subscribeDirectives, debug Serve, frame-pump host.Run) derive from this, so
	// it stops them when the parent ctx is cancelled AND when this host finishes
	// (the conductor/REPL returns) without disturbing the parent or sibling hosts.
	// This reproduces cmd/host's old `cancel()`-after-conductor frame-pump stop.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	host := New(Options{
		Server:      cfg.Server,
		Username:    cfg.Username,
		Password:    cfg.Password,
		Facts:       deps.Facts,
		Landscape:   deps.Landscape,
		WorldOracle: deps.WorldOracle,
		Logger:      log,
	})
	defer host.Close()

	// Per-host writable state: the learned-alias store (recognition) and the
	// durable hostkv store (conductor progress, future trust ledger).
	dataDir := resolveDataDir(log, cfg.DataDir, cfg.Username)
	if cfg.Fresh {
		log.Info("fresh mode: ephemeral state — no persistence, no memory mirror")
		dataDir = "" // degrade store + aliases to in-memory; nothing is written
	}
	host.Resolver = resolve.New(deps.Facts, loadAliasStore(log, dataDir), nil)
	store := openLocalStore(log, dataDir, cfg.Username)
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
	// mesa link (optional): the host's gateway to LLM cognition, RAG, long-term
	// memory, and persona provisioning — everything not compute-local-feasible.
	// When unset, the host runs fully self-contained on its local tiers + pearl.
	var mc mesaclient.Client
	var remote memory.Remote
	if cfg.Mesa != "" && deps.Mesa != nil {
		// Each host dials its OWN connection to its configured mesa instance,
		// authenticated as its host_id, and owns it — RunHost closes it on exit.
		// (One conn per host; no pooling.)
		client, closeConn, err := deps.Mesa(cfg.Mesa, cfg.Username)
		if err != nil {
			return fmt.Errorf("mesa dial %s: %w", cfg.Mesa, err)
		}
		if closeConn != nil {
			defer closeConn()
		}
		mc = client
		if !cfg.Fresh {
			remote = mesaclient.AsRemote(mc, cfg.Username) // memory mirror off in fresh mode
		}
	}

	// Tiered memory: remote is mesa when linked (cross-host recall + Mirror),
	// else offline (NopRemote) — everything served from the local tiers.
	host.Memory = memory.New(memory.Options{Scratch: scratch, Local: local, Remote: remote})

	if mc != nil {
		host.Strategist = mesaclient.AsStrategist(mc, cfg.Username)
		host.Retriever = mesaclient.AsRetriever(mc, cfg.Username)
		host.SetMesaMemory(mc) // two-way episodic LTM: mirror up + cold-start bootstrap down
		log.Info("mesa connected", "addr", cfg.Mesa, "healthy", mc.Healthy())
		// Social reflex: answer players who speak to her on a cheap, reactive
		// path (Game.Chat), off the Act loop, so chatting costs no routine rewrite.
		go socialReflex(ctx, log, host, mc, cfg.Username, cfg.Goal)
	}

	// Persona: when linked, mesa is authoritative — provision it down and compile
	// the pearl policy locally (the table is func-valued, so it can't cross the
	// wire). Offline, fall back to a local persona file. Either way, compiling
	// disposition → the pearl gate + decide seam is what makes the host BEHAVE
	// per-persona.
	var personaGoals []string
	switch {
	case mc != nil:
		g, err := provisionPersona(ctx, log, host, mc, cfg.Username)
		if err != nil {
			log.Warn("mesa provision failed; host runs without a persona", "err", err)
		} else {
			personaGoals = g
		}
	case cfg.PersonaPath != "":
		if err := loadPersona(log, host, cfg.PersonaPath); err != nil {
			return fmt.Errorf("persona: %w", err)
		}
	}

	// Debug control plane: start the recorder before connecting so it captures
	// the login + initial-snapshot events, then serve the HTTP/WS dashboard on
	// its own goroutine alongside the conductor.
	if cfg.Debug != nil && cfg.Debug.Addr != "" && cfg.Debug.New != nil {
		dbg := cfg.Debug.New(host)
		dbg.StartRecorder(ctx)
		go func() {
			if err := dbg.Serve(ctx); err != nil {
				log.Warn("debug server exited", "err", err)
			}
		}()
	}

	log.Info("connecting", "server", cfg.Server, "username", cfg.Username)
	if err := host.Connect(ctx); err != nil {
		return err
	}

	// Frame pump on its own goroutine — keeps the world mirror live. Must be
	// running before the conductor drives routines that wait on world changes.
	// Decision-stream persistence: every AgentThought (director decisions,
	// detours, stalls, goal lifecycle) appends to decisions.jsonl in the
	// host's data dir + mirrors to mesa — the overnight-soak trace.
	if dataDir != "" {
		go host.runDecisionLog(ctx, filepath.Join(dataDir, "decisions.jsonl"))
	}

	hostDone := make(chan error, 1)
	go func() {
		err := host.Run(ctx)
		hostDone <- err
		// The frame pump exiting means the CONNECTION is gone (server
		// EOF, network drop, decode failure). Cancel the per-host ctx so
		// the conductor stops spinning no-op turns against a dead socket
		// and RunHost can return the connection error to the supervisor.
		cancel()
	}()

	// Give the mirror a beat to fill from the initial login burst.
	select {
	case <-ctx.Done():
		return waitHost(ctx, hostDone)
	case <-time.After(time.Second):
	}

	// Director: the mesa Act planner when there's a goal (autonomous play), else
	// the fixed routine sequence/loop from flags. The goal is the operator's
	// -goal flag if given; otherwise it falls back to the host's OWN persona
	// north-star (provisioned from mesa) so she keeps living/acting once any
	// explicit task is done, instead of idling because "the goal is complete".
	goal := cfg.Goal
	// Session genesis: one heavy Opus-at-login compile (mesa-side) reads the
	// host's full history — persona + episodes + relationships + standing goal —
	// and produces THIS session's goal, mood baseline, and attention keyword
	// ladder. The host runs cheap on the compiled output. An explicit -goal
	// overrides the compiled goal; a failure falls back to the persona north-star.
	var genesisLadder []mesaclient.KeywordRung
	if mc != nil && cfg.Genesis {
		pos := host.World().Self.Position()
		ws := fmt.Sprintf("You just woke at map position (%d, %d).", pos.X, pos.Y)
		gctx, gcancel := context.WithTimeout(ctx, 90*time.Second)
		gr, gerr := mc.Genesis(gctx, cfg.Username, "login", ws)
		gcancel()
		if gerr != nil {
			log.Warn("session genesis failed; falling back to persona north-star", "err", gerr)
		} else {
			log.Info("session genesis compiled", "goal", gr.Goal,
				"mood", fmt.Sprintf("stress=%.2f confidence=%.2f valence=%.2f", gr.Mood.Stress, gr.Mood.Confidence, gr.Mood.Valence),
				"keywords", len(gr.KeywordLadder), "reasoning", gr.Reasoning)
			host.SetAffectBaseline(gr.Mood.Stress, gr.Mood.Confidence, gr.Mood.Valence)
			if goal == "" && gr.Goal != "" {
				goal = gr.Goal
			}
			genesisLadder = gr.KeywordLadder
		}
	}
	if goal == "" && len(personaGoals) > 0 {
		goal = "You are living in the world of RuneScape with no fixed task right now. " +
			"Pursue your own purpose, in character: " + personaGoals[0] + " " +
			"Explore, train skills, earn money, talk to people, and make your way — this is an open-ended life, not a checklist to finish."
	}
	var director Director
	if goal != "" && mc != nil {
		md := NewMesaDirector(mc, cfg.Username, goal, log)
		md.SetKeywordLadder(genesisLadder)
		host.SetKeywordLadder(genesisLadder) // reactive trigger detector reads the same ladder (reactive.go)
		// Cheap local loop (#16): replay learned routines without an LLM call,
		// escalating to mesa.Act only on a novel situation. The library persists
		// in the host's local memory tier, so a warmed host runs mostly local.
		lib := NewRoutineLibrary(host.Memory)
		director = NewHybridDirector(md, lib, goal, log)
		log.Info("autonomous mode: cheap local loop + mesa Act escalation", "goal", goal, "library", lib.Len())
	} else {
		director = buildDirector(cfg.Director)
	}
	if cfg.Director.REPL || director == nil {
		if cfg.Headless {
			// Daemon host with no stdin: never block on a REPL. This is reached only
			// if a host has no goal, no routine, AND genesis/persona produced no goal
			// — a misconfiguration the operator should fix.
			cancel() // stop the frame pump
			_ = waitHost(ctx, hostDone)
			return fmt.Errorf("host %q: nothing to do — no goal, no routine, and none derived from genesis/persona", cfg.Username)
		}
		log.Info("entering REPL")
		r := host.NewREPL(ctx, os.Stdin, os.Stdout)
		if err := r.Run(); err != nil {
			log.Warn("repl exited", "err", err)
		}
		cancel() // stop the frame pump
		return waitHost(ctx, hostDone)
	}

	conductor := NewConductor(host, ConductorOptions{
		Director:    director,
		TurnTimeout: cfg.TurnTimeout,
		Settle:      cfg.Settle,
		Store:       local,
		Scratch:     scratch,
		Logger:      log,
		// Interrupt/detour stack for autonomous play (survival preemption): park
		// the grind to eat when HP goes critical, then resume. Off for fixed
		// scripted hosts (load drones don't need it).
		Detours: goal != "" && mc != nil,
	})
	// Bind analysis-mode state: the operator identity (in-game trigger auth), the
	// mesa client (directive interpreter Chat/Act transport), and the live
	// conductor (the turn loop to freeze/thaw). Both trigger paths — the in-game
	// social reflex and the cradle control-plane — drive this single shared state.
	host.configureAnalysis(cfg.Operator, mc, conductor)
	if cfg.OnReady != nil {
		// Publish the live host + conductor + lifecycle ctx to the caller (the
		// cradle registry) so it can pause/resume, introspect, and bind per-host
		// observers that tear down when this host stops.
		cfg.OnReady(&HostHandle{Host: host, Conductor: conductor, Ctx: ctx})
	}
	log.Info("starting conductor")
	cerr := conductor.Run(ctx)
	cancel() // conductor finished or was cancelled — stop the frame pump
	herr := waitHost(ctx, hostDone)
	if cerr != nil && cerr != context.Canceled {
		return cerr
	}
	return herr
}

// buildDirector turns the routine spec into a Director, or returns nil when no
// routines were specified (caller falls back to the REPL).
func buildDirector(spec DirectorSpec) Director {
	if len(spec.Routines) > 0 {
		var intents []Intent
		for _, p := range spec.Routines {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			intents = append(intents, Intent{Label: filepath.Base(p), RoutinePath: p})
		}
		if len(intents) > 0 {
			return Sequence(intents...)
		}
	}
	if spec.Routine != "" {
		in := Intent{Label: filepath.Base(spec.Routine), RoutinePath: spec.Routine}
		if spec.Loop {
			return Loop(in)
		}
		return Sequence(in)
	}
	return nil
}

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

// resolveDataDir mirrors the cradle: an explicit data dir, else
// ~/.westworld/hosts/<user>.
func resolveDataDir(log *slog.Logger, dataDir, username string) string {
	if dataDir != "" {
		return dataDir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		log.Warn("no -data-dir and home dir unavailable; host state will not persist", "err", err)
		return ""
	}
	name := username
	if name == "" {
		name = "anon"
	}
	return filepath.Join(home, ".westworld", "hosts", name)
}
