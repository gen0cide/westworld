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
	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
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
	if cfg.fresh {
		log.Info("fresh mode: ephemeral state — no persistence, no memory mirror")
		dataDir = "" // degrade store + aliases to in-memory; nothing is written
	}
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
	// mesa link (optional): the host's gateway to LLM cognition, RAG, long-term
	// memory, and persona provisioning — everything not compute-local-feasible.
	// When unset, the host runs fully self-contained on its local tiers + pearl.
	var mc *mesaclient.GRPCClient
	var remote memory.Remote
	if cfg.mesa != "" {
		var err error
		// The host authenticates with its derived host key (SHA-512 of its
		// username, for now). mesa binds the resolved host_id into every call.
		mc, err = mesaclient.NewGRPCClient(cfg.mesa, mesaclient.HostKey(cfg.username))
		if err != nil {
			return fmt.Errorf("mesa: %w", err)
		}
		defer mc.Close()
		if !cfg.fresh {
			remote = mesaclient.AsRemote(mc, cfg.username) // memory mirror off in fresh mode
		}
	}

	// Tiered memory: remote is mesa when linked (cross-host recall + Mirror),
	// else offline (NopRemote) — everything served from the local tiers.
	host.Memory = memory.New(memory.Options{Scratch: scratch, Local: local, Remote: remote})

	if mc != nil {
		host.Strategist = mesaclient.AsStrategist(mc, cfg.username)
		host.Retriever = mesaclient.AsRetriever(mc, cfg.username)
		host.SetMesaMemory(mc) // two-way episodic LTM: mirror up + cold-start bootstrap down
		log.Info("mesa connected", "addr", cfg.mesa, "healthy", mc.Healthy())
		// Social reflex: answer players who speak to her on a cheap, reactive
		// path (Game.Chat), off the Act loop, so chatting costs no routine rewrite.
		go socialReflex(rootCtx, log, host, mc, cfg.username, cfg.goal)
	}

	// Persona: when linked, mesa is authoritative — provision it down and compile
	// the pearl policy locally (the table is func-valued, so it can't cross the
	// wire). Offline, fall back to a local persona file. Either way, compiling
	// disposition → the pearl gate + decide seam is what makes the host BEHAVE
	// per-persona.
	var personaGoals []string
	switch {
	case mc != nil:
		g, err := provisionPersona(rootCtx, log, host, mc, cfg.username)
		if err != nil {
			log.Warn("mesa provision failed; host runs without a persona", "err", err)
		} else {
			personaGoals = g
		}
	case cfg.personaPath != "":
		if err := loadPersona(log, host, cfg.personaPath); err != nil {
			return fmt.Errorf("persona: %w", err)
		}
	}

	// Debug control plane: start the recorder before connecting so it captures
	// the login + initial-snapshot events, then serve the HTTP/WS dashboard on
	// its own goroutine alongside the conductor.
	if cfg.debugAddr != "" {
		dbg := debughttp.New(host, debughttp.Config{Username: cfg.username, Addr: cfg.debugAddr}, log)
		dbg.StartRecorder(rootCtx)
		go func() {
			if err := dbg.Serve(rootCtx); err != nil {
				log.Warn("debug server exited", "err", err)
			}
		}()
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

	// Director: the mesa Act planner when there's a goal (autonomous play), else
	// the fixed routine sequence/loop from flags. The goal is the operator's
	// -goal flag if given; otherwise it falls back to the host's OWN persona
	// north-star (provisioned from mesa) so she keeps living/acting once any
	// explicit task is done, instead of idling because "the goal is complete".
	goal := cfg.goal
	// Session genesis: one heavy Opus-at-login compile (mesa-side) reads the
	// host's full history — persona + episodes + relationships + standing goal —
	// and produces THIS session's goal, mood baseline, and attention keyword
	// ladder. The host runs cheap on the compiled output. An explicit -goal
	// overrides the compiled goal; a failure falls back to the persona north-star.
	var genesisLadder []mesaclient.KeywordRung
	if mc != nil {
		pos := host.World().Self.Position()
		ws := fmt.Sprintf("You just woke at map position (%d, %d).", pos.X, pos.Y)
		gctx, gcancel := context.WithTimeout(rootCtx, 90*time.Second)
		gr, gerr := mc.Genesis(gctx, cfg.username, "login", ws)
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
	var director runtime.Director
	if goal != "" && mc != nil {
		md := runtime.NewMesaDirector(mc, cfg.username, goal, log)
		md.SetKeywordLadder(genesisLadder)
		director = md
		log.Info("autonomous mode: mesa Act planner", "goal", goal)
	} else {
		director = buildDirector(cfg)
	}
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

// socialReflex answers players who speak to the host on a cheap, reactive path
// (Game.Chat on the Haiku tier), independent of the Act planning loop. To give
// RICH replies (so "what are you doing?" gets a real answer), it tracks her
// latest published thought (reasoning + perception) off the bus and passes that
// as context — the same self-knowledge the Act loop has, without being the Act
// loop. Ignores her own echoed chat + rate-limits.
func socialReflex(ctx context.Context, log *slog.Logger, host *runtime.Host, mc *mesaclient.GRPCClient, username, goal string) {
	ch := host.Bus().Subscribe("*", 256)
	var last time.Time
	var doing, perception string // her most recent reasoning + perceived context
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch e := ev.(type) {
			case event.AgentThought:
				doing, perception = e.Reasoning, e.Perception // what she's currently up to
			case event.OtherPlayerChat:
				from := reflexPlayerName(host, e.PlayerIndex)
				if strings.EqualFold(from, username) {
					continue // her own chat echoed back — never reply to herself
				}
				if time.Since(last) < 3*time.Second {
					continue // light rate-limit so rapid lines don't spam replies
				}
				text, speak, err := mc.Chat(ctx, username, from, e.MessageText, socialContext(host, goal, doing, perception))
				if err != nil || !speak || text == "" {
					continue
				}
				if err := host.Say(ctx, text); err == nil {
					last = time.Now()
					log.Info("social: replied", "to", from, "heard", e.MessageText, "said", text)
				}
			}
		}
	}
}

// socialContext gathers RICH context for a chat reply: her goal, what she's
// doing right now (latest reasoning), what she recently perceived, and the live
// game messages — so she can actually answer questions like "what are you doing?"
func socialContext(host *runtime.Host, goal, doing, perception string) []string {
	out := make([]string, 0, 8)
	if goal != "" {
		out = append(out, "Your overall goal: "+goal)
	}
	if doing != "" {
		out = append(out, "What you are doing RIGHT NOW: "+doing)
	}
	if perception != "" {
		out = append(out, "What you've recently seen/heard: "+perception)
	}
	if w := host.World(); w != nil && w.Self != nil {
		pos := w.Self.Position()
		out = append(out, fmt.Sprintf("You are at (%d,%d), HP %d/%d.", pos.X, pos.Y, w.Self.HP(), w.Self.MaxHP()))
	}
	return out
}

func reflexPlayerName(host *runtime.Host, idx int) string {
	if w := host.World(); w != nil && w.Players != nil {
		for _, p := range w.Players.All() {
			if p.Index == idx && p.Name != "" {
				return p.Name
			}
		}
	}
	return "a player"
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

// applyPersona compiles a persona's policy and wires it onto the host: the pearl
// table (gate + decide seam), the decision floor, and the affect baseline. This
// is the deterministic local half of the compiler — shared by the offline
// file path and the mesa-provisioned path.
func applyPersona(log *slog.Logger, host *runtime.Host, p *persona.Persona) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}
	cp := persona.CompilePolicy(p)
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

// loadPersona reads + validates a local persona file and applies it (the offline
// path, when there is no mesa to provision from).
func loadPersona(log *slog.Logger, host *runtime.Host, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var p persona.Persona
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return applyPersona(log, host, &p)
}

// provisionPersona pulls the host's authoritative persona from mesa (unary
// Provision.Fetch), applies it, then opens the live push stream so later
// revisions can be applied without a restart.
func provisionPersona(ctx context.Context, log *slog.Logger, host *runtime.Host, mc *mesaclient.GRPCClient, hostID string) ([]string, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	prov, err := mc.Provision(fetchCtx, hostID)
	if err != nil {
		return nil, err
	}
	if err := applyPersona(log, host, &prov.Persona); err != nil {
		return nil, err
	}
	log.Info("provisioned persona from mesa",
		"name", prov.Persona.Cornerstone.Identity.Name,
		"goals", len(prov.Goals), "prose_chars", len(prov.Prose))
	go subscribeDirectives(ctx, log, mc, hostID)
	return prov.Goals, nil
}

// subscribeDirectives consumes the mesa→host push stream (Provision.Subscribe).
// For now it logs each directive; applying PEARL_REFRESH/PERSONA_REVISION
// (recompile) and GOAL_REVISION live lands next.
func subscribeDirectives(ctx context.Context, log *slog.Logger, mc *mesaclient.GRPCClient, hostID string) {
	ch, err := mc.Subscribe(ctx, hostID)
	if err != nil {
		log.Warn("mesa subscribe failed", "err", err)
		return
	}
	for d := range ch {
		log.Info("mesa directive", "id", d.ID, "kind", string(d.Kind), "bytes", len(d.Payload))
	}
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
