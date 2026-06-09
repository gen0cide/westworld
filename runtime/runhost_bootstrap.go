package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/persona"
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
	hostDone := make(chan error, 1)
	go func() { hostDone <- host.Run(ctx) }()

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

// socialReflex answers players who speak to the host on a cheap, reactive path
// (Game.Chat on the Haiku tier), independent of the Act planning loop. To give
// RICH replies (so "what are you doing?" gets a real answer), it tracks her
// latest published thought (reasoning + perception) off the bus and passes that
// as context — the same self-knowledge the Act loop has, without being the Act
// loop. Ignores her own echoed chat + rate-limits.
func socialReflex(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, username, goal string) {
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
				// The AgentThought tick is a host-owned proactive clock: try the
				// intent-driven ASK drive here (its own AgentThought is filtered out
				// below — we ignore our own ask echo to avoid re-entrant asking).
				if e.Trigger != "ask" && e.Trigger != "forage" {
					tryAsk(ctx, log, host, mc, username)
					// FORAGE drive (5b): runs AFTER tryAsk and shares its global floor,
					// so exactly one of {ask, forage} actuates per gap — forage fires
					// only when no local interlocutor could be asked this tick.
					tryForage(ctx, log, host, mc, username)
				}
			case event.OtherPlayerChat:
				from := reflexPlayerName(host, e.PlayerIndex)
				if strings.EqualFold(from, username) {
					continue // her own chat echoed back — never reply to herself
				}
				// OPERATOR OVERRIDE trigger: "!<username> ANALYSIS" / "!<username>
				// RESUME" from the configured operator toggles analysis mode; while
				// active, every operator line is routed to the directive interpreter.
				// Auth is mandatory (a host-takeover vector otherwise): only an EXACT,
				// non-empty sender name == the configured operator is honored — the
				// "a player" placeholder for an unresolved index NEVER authenticates.
				if handled := reflexAnalysis(ctx, log, host, from, e.MessageText, username); handled {
					continue
				}
				// While analysis mode is active, the host is OFFLINE TO THE WORLD:
				// non-operator chat gets no in-character reply.
				if host.AnalysisActive() {
					continue
				}
				if time.Since(last) < 3*time.Second {
					continue // light rate-limit so rapid lines don't spam replies
				}
				// Knowledge-grounded reply (Deliverable 2): answer from what the host
				// actually KNOWS (hedged when low-confidence), say so honestly when it
				// does NOT, and optionally volunteer a high-confidence belief (the
				// host↔host propagation seed). Honesty is a host-supplied FACT, not a
				// hope about the LLM.
				rctx := socialContext(host, goal, doing, perception)
				rctx = append(rctx, host.groundReply(from, e.MessageText, time.Now())...)
				text, speak, err := mc.Chat(ctx, username, from, e.MessageText, rctx)
				if err != nil || !speak || text == "" {
					continue
				}
				// This is a DIRECTED reply to `from` — route the host's own line into
				// ONLY the addressee's conversation window so the Q→A pairs there and
				// does not broadcast into every latched conversation (L7). Genuinely-
				// public DSL chat (actions_ambient say()) keeps the untargeted fan.
				host.reactive.directSelfTo(normalizeSpeaker(from))
				if err := host.Say(ctx, text); err != nil {
					host.reactive.clearDirectSelf() // send failed → drop the unconsumed routing hint
					// Never swallow this again — a silently-dropped reply (e.g. over the
					// RSC 80-char limit) looks exactly like "she isn't talking".
					log.Warn("social: reply failed to send", "to", from, "said", text, "err", err)
				} else {
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
func socialContext(host *Host, goal, doing, perception string) []string {
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

// reflexAnalysis handles the in-game operator-override channel for one inbound
// chat line. It returns true when the line was consumed by analysis mode (a
// trigger, or an operator directive while active) and the caller must NOT fall
// through to the normal social reply.
//
// AUTH (mandatory — a host-takeover vector otherwise): the line is only honored
// when `from` is an EXACT, case-sensitive, non-empty match for the configured
// operator. An unconfigured operator ("") disables the in-game channel entirely;
// reflexPlayerName's "a player" placeholder for an unresolved index can never
// match a real operator name. RSC names ride an untrusted channel, so this is
// the only gate between a passerby and a full host takeover.
func reflexAnalysis(ctx context.Context, log *slog.Logger, host *Host, from, message, username string) bool {
	op := host.AnalysisOperator()
	if op == "" || from == "" || from != op {
		return false // no operator configured, or sender is not the exact operator
	}
	trimmed := strings.TrimSpace(message)
	prefix := "!" + username + " "
	if cmd, ok := cutPrefixFold(trimmed, prefix); ok {
		switch {
		case cmd == "ANALYSIS": // UPPERCASE keyword per spec
			res := host.EnterAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "enter")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		case cmd == "RESUME":
			res := host.ExitAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "exit")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		}
		// "!<username> <directive>" while active: route the rest as a directive.
		if host.AnalysisActive() {
			res := host.Analyze(ctx, cmd)
			log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", cmd)
			_ = host.Say(ctx, analysisSayLine(res))
			return true
		}
		return false
	}
	// No "!<username>" addressing: while active, any bare operator line is a
	// directive (the operator is already in a takeover session with this host).
	if host.AnalysisActive() {
		res := host.Analyze(ctx, trimmed)
		log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", trimmed)
		_ = host.Say(ctx, analysisSayLine(res))
		return true
	}
	return false
}

// cutPrefixFold strips prefix (case-insensitively) from s, reporting whether it
// matched. The host-username addressing "!Delores " should match regardless of
// the operator's capitalization of the name, while the ANALYSIS/RESUME keyword
// itself stays case-sensitive (checked by the caller).
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(s[len(prefix):]), true
}

// analysisSayLine renders an analysis verdict as a single flat in-game line back
// to the operator (capped to the RSC chat limit by the action layer). Flat
// affect: terse + literal, prefixed so the operator can tell it from in-persona
// speech.
func analysisSayLine(res AnalysisResult) string {
	body := res.Text
	if res.Err != "" && body == "" {
		body = "error: " + res.Err
	}
	return "[analysis:" + string(res.Kind) + "] " + body
}

func reflexPlayerName(host *Host, idx int) string {
	if w := host.World(); w != nil && w.Players != nil {
		for _, p := range w.Players.All() {
			if p.Index == idx && p.Name != "" {
				return p.Name
			}
		}
	}
	return "a player"
}

// applyPersona compiles a persona's policy and wires it onto the host: the pearl
// table (gate + decide seam), the decision floor, and the affect baseline. This
// is the deterministic local half of the compiler — shared by the offline
// file path and the mesa-provisioned path.
func applyPersona(log *slog.Logger, host *Host, p *persona.Persona) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}
	cp := persona.CompilePolicy(p)
	host.Pearl = pearl.New(cp.Table, cp.DecisionFloor)
	m := p.Trajectory.Mood
	host.SetAffectBaseline(m.Stress, m.Confidence, m.Valence)
	// Capture the explore<->exploit dial vector for decision-time curiosity
	// weighting (the persona is otherwise discarded after this). Sole chokepoint
	// for both the offline (loadPersona) and mesa (provisionPersona) paths.
	host.curiosity = p.Cornerstone.Prefs.Curiosity
	// Persona north-star: the advancement fallback when the active goal closes and
	// no graph open_goal successor is queued (same chokepoint as curiosity capture).
	host.northStar = strings.TrimSpace(p.Cornerstone.Identity.NorthStar.Statement)
	// One-line "who I am" grounding card for the reactive extractor (the persona is
	// otherwise discarded after this — same chokepoint as the curiosity capture).
	host.personaSnippet = strings.TrimSpace(fmt.Sprintf("%s — %s",
		strings.TrimSpace(p.Cornerstone.Identity.Name), strings.TrimSpace(p.Cornerstone.Identity.ArchetypeTag)))
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
func loadPersona(log *slog.Logger, host *Host, path string) error {
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
func provisionPersona(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, hostID string) ([]string, error) {
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
	go subscribeDirectives(ctx, log, host, mc, hostID)
	return prov.Goals, nil
}

// subscribeDirectives consumes the mesa→host push stream (Provision.Subscribe)
// and applies what it can live. GOAL_REVISION installs an operator goal override
// on the host (read by the director each turn); applying PEARL_REFRESH /
// PERSONA_REVISION (recompile) lands later — those are still logged.
func subscribeDirectives(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, hostID string) {
	ch, err := mc.Subscribe(ctx, hostID)
	if err != nil {
		log.Warn("mesa subscribe failed", "err", err)
		return
	}
	for d := range ch {
		switch d.Kind {
		case mesaclient.DirectiveGoalRevision:
			var goals []string
			if err := json.Unmarshal(d.Payload, &goals); err != nil {
				log.Warn("mesa directive: bad goal_revision payload", "id", d.ID, "err", err)
				continue
			}
			if len(goals) == 0 {
				continue
			}
			host.SetLiveGoal(goals[0])
			log.Info("mesa directive: live goal applied", "id", d.ID, "goal", goals[0])
		default:
			log.Info("mesa directive", "id", d.ID, "kind", string(d.Kind), "bytes", len(d.Payload))
		}
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
