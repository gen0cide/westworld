// Command cradle is a single-bot RSC client built on the runtime.Host
// abstraction.
//
// Usage:
//
//	export WESTWORLD_PASSWORD=...
//	cradle -server localhost:43596 -username alex -walk 120,504
//	cradle -username alex -dwell 30s -watch
//	cradle -username alex -command "heal"
//
// Password sources, in priority: the -password flag, then the
// WESTWORLD_PASSWORD environment variable. The default is the empty
// string — never embed credentials in the binary or in `ps`-visible
// flags on shared hosts.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gen0cide/westworld/cognition/corpus"
	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	orsc "github.com/gen0cide/westworld/render/orsc"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

type config struct {
	server, username, password string
	walkArg, walkToArg         string
	command, sayArg, followArg string
	dropSlot                   int
	pickupArg                  string
	attackNpc, attackPlayer    int
	talkToNpc, tradeInit       int
	tradeAcceptFrom            int // -1 = disabled, else server-index of requester
	tradeDecline               bool
	openBoundary               string // X,Y,DIR
	itemCommandSlot            int    // -1 = unset
	pmTo, pmMsg                string
	addFriend                  string
	lookAround                 bool
	lookRadius                 int
	routinePath                string
	repl                       bool
	dwell                      time.Duration
	watch, look                bool
	factsRoot                  string
	resetOnExit                bool
	wikiDumpDir                string
	devMode                    bool
	dataDir                    string
	renderView                 bool   // after login, render the host's live view to a PNG
	renderOut                  string // output PNG path for -render-view
	renderRotation             int    // 0..255 camera yaw for -render-view (<0 = 8-way sweep)
	renderZoom                 int    // camera zoom for -render-view (750=1x viewport, 1500=2x)
	renderW, renderH           int    // output viewport pixel size (bigger = wider FOV, same detail)
	spectate                   bool   // after login, serve a live browser viewport that follows the host
	spectateAddr               string // host:port for the -spectate HTTP server
	debugHTTP                  bool   // after login, serve the scriptable HTTP debug control plane
	debugAddr                  string // host:port for the -debug-http server
	debugLog                   string // JSONL event-log path (default /tmp/cradle_debug/<username>_events.jsonl)
}

func main() {
	var cfg config
	flag.StringVar(&cfg.server, "server", "localhost:43596", "OpenRSC server host:port")
	flag.StringVar(&cfg.username, "username", "alex", "RSC account username")
	flag.StringVar(&cfg.password, "password", "", "RSC account password (or set WESTWORLD_PASSWORD env var)")
	flag.StringVar(&cfg.walkArg, "walk", "", "optional destination coords as X,Y (e.g., 120,504); single FOV-bounded click")
	flag.StringVar(&cfg.walkToArg, "walkto", "", "like -walk but chunks long journeys into multiple in-FOV segments")
	flag.StringVar(&cfg.command, "command", "", "optional admin command to send after login (e.g., 'heal')")
	flag.StringVar(&cfg.sayArg, "say", "", "optional public chat message to send after login")
	flag.StringVar(&cfg.followArg, "follow", "", "after login, follow the named player (server-side opcode 165)")
	flag.IntVar(&cfg.dropSlot, "drop", -1, "after login, drop the inventory item in this slot")
	flag.StringVar(&cfg.pickupArg, "pickup", "", "after login, pick up the ground item at X,Y,ID (e.g., '120,648,428')")
	flag.IntVar(&cfg.attackNpc, "attack-npc", -1, "after login, attack the NPC at this server index")
	flag.IntVar(&cfg.attackPlayer, "attack-player", -1, "after login, attack the player at this server index (PVP zones only)")
	flag.IntVar(&cfg.talkToNpc, "talkto", -1, "after login, talk to the NPC at this server index")
	flag.IntVar(&cfg.tradeInit, "trade-init", -1, "after login, send a trade request to the player at this server index")
	flag.IntVar(&cfg.tradeAcceptFrom, "trade-accept-from", -1, "after login, accept a pending trade request from this server-index (use the requester's index)")
	flag.BoolVar(&cfg.tradeDecline, "trade-decline", false, "after login, decline any pending or active trade")
	flag.StringVar(&cfg.openBoundary, "open-boundary", "", "after login, interact with boundary at X,Y,DIR (e.g., '132,641,2' to open the Lumbridge shop door)")
	flag.IntVar(&cfg.itemCommandSlot, "bury", -1, "after login, fire default item action (Bury/Eat/etc) on this inventory slot")
	flag.StringVar(&cfg.pmTo, "pm-to", "", "after login, send private message TO this player (use with -pm-msg)")
	flag.StringVar(&cfg.pmMsg, "pm-msg", "", "private message body (use with -pm-to)")
	flag.StringVar(&cfg.addFriend, "add-friend", "", "after login, add this player to friends list (required before PMs can be sent/received)")
	flag.BoolVar(&cfg.lookAround, "look-around", false, "after login, print an LLM-style observation report of the bot's surroundings")
	flag.IntVar(&cfg.lookRadius, "look-radius", 10, "radius (tiles) for -look-around")
	flag.StringVar(&cfg.routinePath, "routine", "", "after login, parse + run this .routine file against the live host (no -dwell needed)")
	flag.BoolVar(&cfg.repl, "repl", false, "after login, drop into an interactive REPL (see docs/lang/repl.md)")
	flag.DurationVar(&cfg.dwell, "dwell", 5*time.Second, "how long to stay logged in after the optional walk/command")
	flag.BoolVar(&cfg.watch, "watch", false, "log all events received from the server during dwell")
	flag.BoolVar(&cfg.look, "look", false, "after login, log scenery/NPCs known to be near our position (facts-derived)")
	flag.StringVar(&cfg.factsRoot, "facts", "/Users/flint/Code/openrsc", "OpenRSC source root for static facts; empty disables")
	flag.StringVar(&cfg.wikiDumpDir, "wiki-dump", "", "directory of rsc.wiki HTML pages to load as the gameplay corpus (e.g., ~/Code/westworld/local/rscwiki/pages); empty disables")
	flag.BoolVar(&cfg.devMode, "dev", false, "enable dev-namespace corpora (autorune, server source). NEVER pass this for production hosts — they are supposed to know only what a real player would.")
	flag.StringVar(&cfg.dataDir, "data-dir", "", "per-host writable data directory (learned-alias store etc.); defaults to ~/.westworld/hosts/<username>")
	flag.BoolVar(&cfg.resetOnExit, "reset-on-exit", false, "ADMIN/TEST ONLY: before logging out, wipe inventory + teleport to Lumbridge spawn so the next scenario on this drone starts clean. Requires an admin account; never pass for production hosts.")
	flag.BoolVar(&cfg.renderView, "render-view", false, "after login + world-state load, render the host's live in-game view to a PNG (the decoupled SnapshotFromCradle -> RenderView path)")
	flag.StringVar(&cfg.renderOut, "render-out", "/tmp/render_out/bernard_live.png", "output PNG path for -render-view")
	flag.IntVar(&cfg.renderRotation, "render-rotation", 64, "camera yaw (0..255) for -render-view; negative = render the full 8-way 45deg sweep from one snapshot")
	flag.IntVar(&cfg.renderZoom, "render-zoom", 1100, "camera zoom for -render-view (750 = 1x viewport, 1500 = 2x — see more world at the same resolution); pulled back a bit so multi-story buildings + roofs are framed")
	flag.IntVar(&cfg.renderW, "render-w", 512, "output viewport WIDTH in px for -render-view (larger = wider field of view at the same per-pixel detail, NOT a zoom-out)")
	flag.IntVar(&cfg.renderH, "render-h", 336, "output viewport HEIGHT in px for -render-view")
	flag.BoolVar(&cfg.spectate, "spectate", false, "after login, serve a LIVE browser viewport (http) that follows the host around; arrow keys rotate the camera, +/- zoom. No native window / CGo — the browser is the display.")
	flag.StringVar(&cfg.spectateAddr, "spectate-addr", "localhost:8089", "host:port for the -spectate HTTP viewport server")
	flag.BoolVar(&cfg.debugHTTP, "debug-http", false, "after login, serve a SCRIPTABLE HTTP debug control plane (POST /eval a DSL line, POST /script a routine, GET /state, GET /events) and record every event to a JSONL log. Blocks until Ctrl-C, keeping the host logged in. See cmd/cradle/debug.go.")
	flag.StringVar(&cfg.debugAddr, "debug-addr", "localhost:8090", "host:port for the -debug-http control plane")
	flag.StringVar(&cfg.debugLog, "debug-log", "", "JSONL event-log path for -debug-http (default /tmp/cradle_debug/<username>_events.jsonl)")
	verbose := flag.Bool("v", false, "debug-level logging")
	flag.Parse()

	// Fall back to env var if -password wasn't supplied. Never echo
	// the password back; just take it.
	if cfg.password == "" {
		cfg.password = os.Getenv("WESTWORLD_PASSWORD")
	}
	if cfg.password == "" {
		fmt.Fprintln(os.Stderr, "cradle: missing password — pass -password or set WESTWORLD_PASSWORD")
		os.Exit(2)
	}

	// In -repl mode, suppress INFO logs so the interactive prompt
	// stays clean. WARN+ still surfaces because routine authors
	// want to see real problems. `-v` overrides and brings DEBUG
	// back regardless of mode.
	level := slog.LevelInfo
	if cfg.repl && !*verbose {
		level = slog.LevelWarn
	}
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if err := run(log, cfg); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, cfg config) error {
	rootCtx, cancel := signalContext()
	defer cancel()

	// Load static world facts if a path was provided. For a swarm,
	// this would happen once at delos startup; for single-host dev
	// it's just per-invocation.
	var loadedFacts *facts.Facts
	var loadedLandscape *pathfind.Landscape
	if cfg.factsRoot != "" {
		var err error
		loadedFacts, err = facts.Load(facts.DefaultSources(cfg.factsRoot))
		if err != nil {
			log.Warn("facts load failed; continuing without world knowledge", "err", err)
		} else {
			log.Info("loaded world facts", "summary", loadedFacts.Summary())
		}
		landscapePath := filepath.Join(cfg.factsRoot, "server", "conf", "server", "data", "Authentic_Landscape.orsc")
		loadedLandscape, err = pathfind.OpenLandscape(landscapePath)
		if err != nil {
			log.Warn("landscape load failed; pathfinding disabled", "err", err)
		} else {
			log.Info("loaded landscape archive", "path", landscapePath)
			defer loadedLandscape.Close()
		}
	}

	// Knowledge corpus. Each host can be configured with a set of
	// retrieval sources; the namespace gate is load-time, so a
	// production cradle launched without -dev physically cannot
	// have dev content in memory. See docs/lang/knowledge.md.
	loadedCorpus := loadCorpus(log, cfg)

	host := runtime.New(runtime.Options{
		Server:    cfg.server,
		Username:  cfg.username,
		Password:  cfg.password,
		Facts:     loadedFacts,
		Landscape: loadedLandscape,
		Logger:    log,
	})
	host.Corpus = loadedCorpus

	// Recognition faculty (api.md §5). Build a per-host resolver over
	// the loaded facts catalog, backed by a JSON alias store under the
	// host's writable data dir so learned lingo ("r2h" → "Rune 2-handed
	// Sword") survives restarts. The brain stage is left as the resolve
	// package's stub default (nil → noBrain, i.e. alias + fuzzy only);
	// the real LLM-backed Brain drops in here in Phase 3/4 without
	// changing the wiring. A facts-load failure leaves loadedFacts nil,
	// in which case spells/prayers still resolve and the world catalogs
	// are simply empty.
	dataDir := resolveDataDir(log, cfg)
	var aliasStore *resolve.AliasStore
	if dataDir != "" {
		aliasPath := filepath.Join(dataDir, "aliases.json")
		var aerr error
		aliasStore, aerr = resolve.LoadAliasStore(aliasPath)
		if aerr != nil {
			log.Warn("alias store load failed; using in-memory recognition (no persistence)", "path", aliasPath, "err", aerr)
			aliasStore = nil
		} else {
			log.Info("loaded learned-alias store", "path", aliasPath, "aliases", aliasStore.Len())
		}
	}
	host.Resolver = resolve.New(loadedFacts, aliasStore, nil)

	defer host.Close()

	log.Info("connecting", "server", cfg.server)
	if err := host.Connect(rootCtx); err != nil {
		return err
	}

	// Subscribe to all events for the watch UI / debug logging.
	watchCh := host.Bus().Subscribe("*", 256)
	go watchEvents(log, watchCh, cfg.watch)

	// Debug control plane: start the event recorder immediately (so the
	// login welcome + initial inventory/stats/position snapshots land in
	// the JSONL log) — the HTTP server itself is started later, after the
	// initial world state settles.
	var dbg *debugServer
	if cfg.debugHTTP {
		dbg = newDebugServer(log, cfg, host)
		dbg.startRecorder(rootCtx)
	}

	// Run the host's main loop in a goroutine; the rest of this
	// function drives the script.
	hostDone := make(chan error, 1)
	go func() {
		hostDone <- host.Run(rootCtx)
	}()

	// Clean logout on EVERY exit path — including the early
	// `return fmt.Errorf("routine: ...")` when a routine errors. A
	// hard TCP close (defer host.Close) leaves the server holding the
	// session, so a fast re-login of this same account gets login
	// code 4. Deferred here (after host.Run is pumping inbound, so
	// the LogoutConfirm can be received) and registered AFTER
	// `defer host.Close()` so it runs first (LIFO). Uses a fresh
	// context so a Ctrl-C (rootCtx cancel) still lets the logout
	// complete; shortens the wait when we're already shutting down.
	defer func() {
		wait := 12 * time.Second
		if rootCtx.Err() != nil {
			wait = 2 * time.Second // signal-driven shutdown: don't dawdle
		}
		// Test/harness cleanup: leave the drone in a known clean
		// state (empty bag, parked at Lumbridge spawn) so the next
		// scenario on this account isn't contaminated. Admin-only;
		// runs before logout while the connection is still alive.
		// Skipped on signal-driven shutdown (don't fight a Ctrl-C).
		if cfg.resetOnExit && rootCtx.Err() == nil {
			rctx, rcancel := context.WithTimeout(context.Background(), 4*time.Second)
			// NB: ::wipeinv requires an explicit player name ("::wipeinv
			// <name>"); the self-targeting commands (heal/recharge/
			// teleport) do not.
			resetCmds := []string{
				"wipeinv " + cfg.username,
				"heal",
				"recharge",
				"teleport 120 649",
			}
			for _, c := range resetCmds {
				if err := host.Command(rctx, c); err != nil {
					log.Warn("reset-on-exit command failed", "cmd", c, "err", err)
				}
				time.Sleep(300 * time.Millisecond)
			}
			rcancel()
		}
		lctx, lcancel := context.WithTimeout(context.Background(), wait+time.Second)
		defer lcancel()
		log.Info("logging out (graceful)")
		if err := host.LogoutGraceful(lctx, wait); err != nil {
			log.Warn("graceful logout not confirmed; hard-closing", "err", err)
		}
	}()

	// Give the server a moment to send initial state, including the
	// first tick of NPC/player position data the world-state mirror
	// needs for the walk-hint logic in Host.AttackNpc/TalkToNpc.
	time.Sleep(3 * time.Second)

	// Debug control plane: serve the scriptable HTTP harness and block
	// until Ctrl-C. This is its own driving loop (like -spectate); the
	// one-shot action flags below are skipped. The deferred graceful
	// logout still fires on exit.
	if cfg.debugHTTP {
		log.Info("entering debug-http control plane", "addr", cfg.debugAddr)
		return dbg.serve(rootCtx)
	}

	if cfg.walkArg != "" {
		x, y, err := parseCoord(cfg.walkArg)
		if err != nil {
			return fmt.Errorf("parse -walk: %w", err)
		}
		log.Info("walking (single click)", "to", fmt.Sprintf("(%d, %d)", x, y))
		if err := host.Walk(rootCtx, x, y); err != nil {
			return fmt.Errorf("walk: %w", err)
		}
	}

	if cfg.walkToArg != "" {
		x, y, err := parseCoord(cfg.walkToArg)
		if err != nil {
			return fmt.Errorf("parse -walkto: %w", err)
		}
		log.Info("walking-to (multi-segment)", "target", fmt.Sprintf("(%d, %d)", x, y))
		if err := host.WalkTo(rootCtx, x, y); err != nil {
			log.Warn("walkto did not complete", "err", err)
		} else {
			log.Info("walkto complete")
		}
	}

	if cfg.command != "" {
		log.Info("sending admin command", "cmd", cfg.command)
		if err := host.Command(rootCtx, cfg.command); err != nil {
			return fmt.Errorf("command: %w", err)
		}
	}

	if cfg.sayArg != "" {
		log.Info("saying publicly", "msg", cfg.sayArg)
		if err := host.Say(rootCtx, cfg.sayArg); err != nil {
			return fmt.Errorf("say: %w", err)
		}
	}

	if cfg.dropSlot >= 0 {
		log.Info("dropping item", "slot", cfg.dropSlot)
		if err := host.DropItem(rootCtx, cfg.dropSlot); err != nil {
			return fmt.Errorf("drop: %w", err)
		}
	}

	if cfg.pickupArg != "" {
		parts := strings.SplitN(cfg.pickupArg, ",", 3)
		if len(parts) != 3 {
			return fmt.Errorf("parse -pickup: expected X,Y,ID, got %q", cfg.pickupArg)
		}
		px, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		py, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		pid, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		log.Info("picking up", "at", fmt.Sprintf("(%d, %d)", px, py), "item_id", pid)
		if err := host.PickUpItem(rootCtx, px, py, pid); err != nil {
			return fmt.Errorf("pickup: %w", err)
		}
	}

	if cfg.tradeAcceptFrom >= 0 {
		log.Info("accepting pending trade request", "from", cfg.tradeAcceptFrom)
		if err := host.AcceptIncomingTrade(rootCtx, cfg.tradeAcceptFrom); err != nil {
			return fmt.Errorf("trade accept: %w", err)
		}
	}

	if cfg.tradeDecline {
		log.Info("declining trade")
		if err := host.DeclineTrade(rootCtx); err != nil {
			return fmt.Errorf("trade decline: %w", err)
		}
	}

	if cfg.tradeInit >= 0 {
		log.Info("sending trade request", "server_index", cfg.tradeInit)
		if err := host.InitTradeRequest(rootCtx, cfg.tradeInit); err != nil {
			return fmt.Errorf("trade init: %w", err)
		}
	}

	if cfg.openBoundary != "" {
		parts := strings.SplitN(cfg.openBoundary, ",", 3)
		if len(parts) != 3 {
			return fmt.Errorf("parse -open-boundary: expected X,Y,DIR, got %q", cfg.openBoundary)
		}
		bx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		by, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		bd, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		log.Info("interacting with boundary", "at", fmt.Sprintf("(%d, %d)", bx, by), "dir", bd)
		if err := host.InteractWithBoundary(rootCtx, bx, by, bd); err != nil {
			return fmt.Errorf("open-boundary: %w", err)
		}
	}

	if cfg.itemCommandSlot >= 0 {
		log.Info("firing default item action", "slot", cfg.itemCommandSlot)
		if err := host.ItemCommand(rootCtx, cfg.itemCommandSlot); err != nil {
			return fmt.Errorf("item-command: %w", err)
		}
	}

	if cfg.addFriend != "" {
		log.Info("adding friend", "name", cfg.addFriend)
		if err := host.AddFriend(rootCtx, cfg.addFriend); err != nil {
			return fmt.Errorf("add-friend: %w", err)
		}
		// Give the server a tick to process the friend-add before we
		// try to send a PM through the new relationship.
		time.Sleep(800 * time.Millisecond)
	}

	if cfg.pmTo != "" && cfg.pmMsg != "" {
		log.Info("sending private message", "to", cfg.pmTo, "msg", cfg.pmMsg)
		if err := host.PrivateMessage(rootCtx, cfg.pmTo, cfg.pmMsg); err != nil {
			return fmt.Errorf("pm: %w", err)
		}
	}

	if cfg.talkToNpc >= 0 {
		log.Info("talking to NPC", "server_index", cfg.talkToNpc)
		if err := host.TalkToNpc(rootCtx, cfg.talkToNpc); err != nil {
			return fmt.Errorf("talkto: %w", err)
		}
	}

	if cfg.attackNpc >= 0 {
		log.Info("attacking NPC", "server_index", cfg.attackNpc)
		if err := host.AttackNpc(rootCtx, cfg.attackNpc); err != nil {
			return fmt.Errorf("attack-npc: %w", err)
		}
	}

	if cfg.attackPlayer >= 0 {
		log.Info("attacking player", "server_index", cfg.attackPlayer)
		if err := host.AttackPlayer(rootCtx, cfg.attackPlayer); err != nil {
			return fmt.Errorf("attack-player: %w", err)
		}
	}

	if cfg.followArg != "" {
		log.Info("starting follow", "target", cfg.followArg)
		followCtx, cancel := context.WithTimeout(rootCtx, cfg.dwell)
		err := host.Follow(followCtx, cfg.followArg, 30*time.Second)
		cancel()
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			log.Warn("follow ended", "err", err)
		} else {
			log.Info("follow ended (dwell expired)")
		}
	}

	if cfg.lookAround {
		report := host.DescribeSurroundings(cfg.lookRadius)
		log.Info("=== look-around report ===\n" + report)
	}

	if cfg.repl {
		log.Info("entering REPL")
		r := host.NewREPL(rootCtx, os.Stdin, os.Stdout)
		if err := r.Run(); err != nil {
			log.Warn("repl exited with error", "err", err)
		}
	} else if cfg.routinePath != "" {
		log.Info("running routine", "path", cfg.routinePath, "timeout", cfg.dwell)
		routineCtx, cancel := context.WithTimeout(rootCtx, cfg.dwell)
		res, err := host.RunRoutine(routineCtx, cfg.routinePath, nil)
		cancel()
		if err != nil {
			return fmt.Errorf("routine: %w", err)
		}
		errStr := ""
		if res.Err != nil {
			errStr = res.Err.Error()
		}
		log.Info("routine ended",
			"kind", res.Kind.String(),
			"value", routineValueString(res),
			"err", errStr,
		)
	} else {
		log.Info("dwelling", "for", cfg.dwell)
		select {
		case <-rootCtx.Done():
		case <-time.After(cfg.dwell):
		}
	}

	if cfg.look && host.Facts() != nil {
		pos := host.World().Self.Position()
		log.Info("looking around", "from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y))
		near := host.Facts().Near(pos.X, pos.Y, 8)
		scenery, boundary, npcs, items := 0, 0, 0, 0
		for _, p := range near {
			switch p.Kind {
			case "scenery":
				scenery++
				if scenery <= 5 {
					log.Info("known nearby",
						"kind", "scenery",
						"name", p.Name,
						"at", fmt.Sprintf("(%d, %d)", p.X, p.Y),
					)
				}
			case "boundary":
				boundary++
			case "npc_spawn":
				npcs++
				if npcs <= 5 {
					log.Info("known nearby",
						"kind", "npc_spawn",
						"name", p.Name,
						"at", fmt.Sprintf("(%d, %d)", p.X, p.Y),
					)
				}
			case "ground_item":
				items++
			}
		}
		log.Info("look summary",
			"scenery", scenery,
			"boundaries", boundary,
			"npc_spawns", npcs,
			"ground_items", items,
			"within_tiles", 8,
		)
	}

	if cfg.renderView {
		if err := renderLiveView(log, cfg, host, loadedLandscape, loadedFacts); err != nil {
			log.Warn("render-view failed", "err", err)
		}
	}

	// Live browser viewport: blocks (serving HTTP) until rootCtx is cancelled
	// (Ctrl-C), keeping the host logged in + walking-visible the whole time.
	if cfg.spectate {
		if err := spectate(rootCtx, log, cfg, host, loadedLandscape, loadedFacts); err != nil {
			log.Warn("spectate failed", "err", err)
		}
	}

	// Logout is handled by the deferred LogoutGraceful registered
	// right after host.Run launched — it fires on every exit path
	// (normal, abort, and the early routine-error return) and waits
	// for the server to actually release the session.

	// Final position read.
	pos := host.World().Self.Position()
	log.Info("final state",
		"position", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"hp", host.World().Self.HP(),
		"max_hp", host.World().Self.MaxHP(),
		"fatigue", host.World().Self.Fatigue(),
		"combat_level", host.World().Self.CombatLevel(),
		"inventory_used", 30-host.World().Inventory.FreeSlots(),
	)
	return nil
}

// renderLiveView captures the host's CURRENT perceived world state (its live
// tile position + the terrain sector and scenery it perceives there) and
// rasterizes the host's-eye view to a PNG. This is the decoupled
// SnapshotFromCradle -> RenderView path from docs/render-port-plan.md: the
// renderer is a read-only consumer of cradle state (pathfind.Landscape +
// facts), so it renders what THIS host actually sees standing where it stands.
//
// We wait (poll) for a non-spawn position to ensure the region/world-state has
// actually loaded before snapshotting — a freshly-connected host may briefly
// report (0,0) until the first position update lands.
func renderLiveView(log *slog.Logger, cfg config, host *runtime.Host, land *pathfind.Landscape, f *facts.Facts) error {
	if land == nil {
		return fmt.Errorf("render-view: no landscape loaded (need -facts pointing at OpenRSC root)")
	}

	// Wait for the world-state to load before snapshotting.
	pos, ok := waitForLivePosition(host)
	if !ok {
		return fmt.Errorf("render-view: host position never loaded (still 0,0)")
	}

	plane := pos.Plane()
	localY := pos.Y - plane*world.PlaneHeight
	log.Info("rendering live host view",
		"world", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"plane", plane,
		"local", fmt.Sprintf("(%d, %d)", pos.X, localY),
		"rotation", cfg.renderRotation,
	)

	// Snapshot the live perceived world into a View (shared with -spectate).
	v := buildLiveView(host, pos)
	v.Zoom = cfg.renderZoom
	v.W = cfg.renderW
	v.H = cfg.renderH
	log.Info("snapshotted perceived entities for render", "count", len(v.Entities))

	// Rotations to render. A single yaw (>=0), or — when -render-rotation is
	// negative — the full 8-way 45deg sweep from this ONE frozen snapshot, so
	// the angles are an apples-to-apples set (no world drift between separate
	// logins). Output gets a _rotN suffix in sweep mode.
	rots := []int{cfg.renderRotation}
	sweep := cfg.renderRotation < 0
	if sweep {
		rots = []int{0, 32, 64, 96, 128, 160, 192, 224}
	}
	if dir := filepath.Dir(cfg.renderOut); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	for i, rot := range rots {
		v.Rotation = rot
		png, err := orsc.RenderViewCached(land, f, v)
		if err != nil {
			return fmt.Errorf("render-view rot %d: %w", rot, err)
		}
		out := cfg.renderOut
		if sweep {
			ext := filepath.Ext(out)
			out = out[:len(out)-len(ext)] + fmt.Sprintf("_rot%d", i) + ext
		}
		if err := os.WriteFile(out, png, 0o644); err != nil {
			return fmt.Errorf("render-view: write %q: %w", out, err)
		}
		log.Info("wrote live host view PNG", "path", out, "rotation", rot, "bytes", len(png))
	}
	return nil
}

// watchEvents logs each event arriving from the bus. If watch=false,
// only logs interesting ones (not UnknownPacket noise).
func watchEvents(log *slog.Logger, ch <-chan event.Event, watch bool) {
	for ev := range ch {
		switch e := ev.(type) {
		case event.ChatReceived:
			log.Info("chat", "from", e.Speaker, "msg", e.Message)
		case event.PrivateMessage:
			log.Info("private message", "from", e.Sender, "msg", e.Message)
		case event.SystemMessage:
			log.Info("system message", "msg", e.Message)
		case event.WelcomeInfo:
			log.Info("welcome", "last_ip", e.LastLoginIP, "days_ago", e.DaysSinceLogin)
		case event.StatsSnapshot:
			log.Info("stats snapshot",
				"hp", fmt.Sprintf("%d/%d", e.Current[3], e.Max[3]),
				"combat", fmt.Sprintf("atk=%d str=%d def=%d", e.Max[0], e.Max[2], e.Max[1]),
				"quest_points", e.QuestPoints,
			)
		case event.StatUpdate:
			log.Info("stat changed",
				"skill", event.SkillName(e.Skill),
				"current", e.Current,
				"max", e.Max,
				"xp", e.Experience,
			)
		case event.FatigueUpdate:
			if watch {
				log.Info("fatigue", "value", e.Value)
			}
		case event.InventorySnapshot:
			log.Info("inventory snapshot", "slots_used", len(e.Items))
		case event.InventorySlotUpdate:
			if e.Item != nil {
				log.Info("inv slot update", "slot", e.Slot, "item", e.Item.ItemID, "amount", e.Item.Amount)
			} else {
				log.Info("inv slot cleared", "slot", e.Slot)
			}
		case event.GroundItemEvent:
			if watch {
				log.Info("ground item",
					"id", e.ItemID,
					"offset", fmt.Sprintf("(%d, %d)", e.OffsetX, e.OffsetY),
					"disappear", e.Disappear,
				)
			}
		case event.NpcDialogText:
			log.Info("npc said", "text", e.Text)
		case event.NpcDialog:
			log.Info("npc options", "choices", e.Options)
		case event.Death:
			log.Warn("YOU DIED")
		case event.OwnPositionUpdate:
			if watch {
				log.Info("position", "x", e.X, "y", e.Y, "sprite", e.Sprite)
			}
		case event.NearbyPlayerEvent:
			log.Info("nearby player",
				"index", e.Index,
				"at", fmt.Sprintf("(%d, %d)", e.X, e.Y),
				"sprite", e.Sprite,
			)
		case event.OtherPlayerChat:
			msg := e.MessageText
			if msg == "" {
				msg = fmt.Sprintf("<rsc-encoded %d bytes>", len(e.MessageRaw))
			}
			log.Info("other player chat",
				"player_index", e.PlayerIndex,
				"chat_kind", e.ChatKind,
				"icon", e.Icon,
				"msg", msg,
			)
		case event.OtherPlayerAppearance:
			log.Info("other player appearance",
				"player_index", e.PlayerIndex,
				"name", e.Name,
				"appearance_id", e.AppearanceID,
			)
		case event.OtherPlayerDamage:
			log.Info("other player damage",
				"player_index", e.PlayerIndex,
				"damage", e.Damage,
				"hp", fmt.Sprintf("%d/%d", e.CurHits, e.MaxHits),
			)
		case event.TradeRequestReceived:
			log.Info("TRADE REQUEST received", "from_player_index", e.FromPlayerIndex)
		case event.TradeOpened:
			log.Info("TRADE opened", "opponent", e.OpponentName, "my_items", len(e.MyItems), "their_items", len(e.OpponentItems))
		case event.TradeOtherAccepted:
			log.Info("trade: other side clicked accept")
		case event.TradeClosed:
			log.Info("trade closed/cancelled")
		case event.NpcNearby:
			if watch {
				log.Info("nearby NPC",
					"index", e.Index,
					"type_id", e.TypeID,
					"at", fmt.Sprintf("(%d, %d)", e.X, e.Y),
					"new", e.IsNew,
				)
			}
		case event.LogoutConfirm:
			log.Info("server confirmed logout")
		case event.UnknownPacket:
			if watch {
				log.Info("unknown packet (undecoded)",
					"opcode", fmt.Sprintf("0x%02x (%d)", e.Opcode, e.Opcode),
					"size", e.PayloadSize,
				)
			}
		}
	}
}

func parseCoord(s string) (int, int, error) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected X,Y, got %q", s)
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("parse x: %w", err)
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("parse y: %w", err)
	}
	return x, y, nil
}

// routineValueString stringifies a routine Result for logging,
// guarding against nil values.
func routineValueString(r interp.Result) string {
	if r.Value == nil {
		return ""
	}
	return r.Value.Display()
}

// loadCorpus builds the knowledge corpus federation according to
// cradle config. Namespace gating is enforced here:
//
//   - rsc.wiki (Gameplay) is loaded whenever -wiki-dump points at a
//     non-empty HTML page directory.
//   - Dev sources (AutoRune scripts, OpenRSC server source, etc.)
//     are ONLY loaded when -dev is passed. They have not been
//     ingested yet (Phase 2.6 slice 3+); when they do exist, this
//     is the single chokepoint that controls whether the running
//     cradle can see them.
//
// Production hosts should never pass -dev. The federation will
// refuse to enter dev sources into memory if -dev is absent — the
// safety property is enforced at construction time, not query time.
func loadCorpus(log *slog.Logger, cfg config) corpus.Corpus {
	var sources []corpus.Source

	if cfg.wikiDumpDir != "" {
		mc, stats, err := corpus.LoadWikiDump(cfg.wikiDumpDir)
		if err != nil {
			log.Warn("wiki dump load failed; gameplay corpus skipped", "err", err)
		} else {
			log.Info("loaded rsc.wiki",
				"chunks", stats.Chunks,
				"pages", stats.LoadedFiles,
				"empty", stats.EmptyPages,
				"failed", stats.FailedFiles)
			sources = append(sources, corpus.Source{
				Name:      "rscwiki",
				Namespace: corpus.Gameplay,
				Corpus:    mc,
			})
		}
	}

	// Dev sources go here when slices 3+ land. For each one, add a
	// corpus.Source{Name, Namespace: corpus.Dev, Corpus}. They will
	// only enter the federation if -dev is true.

	if len(sources) == 0 {
		return nil
	}

	allowed := []corpus.Namespace{corpus.Gameplay}
	if cfg.devMode {
		allowed = []corpus.Namespace{corpus.Gameplay, corpus.Dev}
	}
	fed := corpus.NewFederation(sources, allowed)

	// Log the access surface so operators can see exactly what
	// knowledge the running host has. This is the auditable record.
	for _, s := range fed.Sources() {
		log.Info("corpus source enabled",
			"name", s.Name,
			"namespace", s.Namespace,
			"chunks", s.ChunkCount)
	}
	if cfg.devMode {
		log.Warn("cradle started in -dev mode; dev-namespace corpora are available — do NOT use this flag for production hosts")
	}
	return fed
}

// resolveDataDir picks the per-host writable data directory used for
// the learned-alias store (and future per-host state). Precedence:
//
//  1. -data-dir flag, if set.
//  2. ~/.westworld/hosts/<username>, if the home dir is discoverable.
//
// Returns "" when no directory can be determined (e.g. home dir
// unavailable and no flag given); the caller then runs with an
// in-memory, non-persisted alias store. The directory itself is not
// created here — the alias store's atomic flush MkdirAll's it on the
// first Learn that writes back.
func resolveDataDir(log *slog.Logger, cfg config) string {
	if cfg.dataDir != "" {
		return cfg.dataDir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		log.Warn("no -data-dir and home dir unavailable; recognition will not persist learned aliases", "err", err)
		return ""
	}
	name := cfg.username
	if name == "" {
		name = "anon"
	}
	return filepath.Join(home, ".westworld", "hosts", name)
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
