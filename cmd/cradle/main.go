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

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/runtime"
)

type config struct {
	server, username, password   string
	walkArg, walkToArg           string
	command, sayArg, followArg   string
	dropSlot                     int
	pickupArg                    string
	attackNpc, attackPlayer      int
	talkToNpc, tradeInit         int
	tradeAccept, tradeDecline    bool
	openBoundary                 string // X,Y,DIR
	itemCommandSlot              int    // -1 = unset
	pmTo, pmMsg                  string
	addFriend                    string
	combatTarget                 string
	combatKills                  int
	lookAround                   bool
	lookRadius                   int
	dwell                        time.Duration
	watch, look                  bool
	factsRoot                    string
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
	flag.BoolVar(&cfg.tradeAccept, "trade-accept", false, "after login, accept any pending trade request")
	flag.BoolVar(&cfg.tradeDecline, "trade-decline", false, "after login, decline any pending or active trade")
	flag.StringVar(&cfg.openBoundary, "open-boundary", "", "after login, interact with boundary at X,Y,DIR (e.g., '132,641,2' to open the Lumbridge shop door)")
	flag.IntVar(&cfg.itemCommandSlot, "bury", -1, "after login, fire default item action (Bury/Eat/etc) on this inventory slot")
	flag.StringVar(&cfg.pmTo, "pm-to", "", "after login, send private message TO this player (use with -pm-msg)")
	flag.StringVar(&cfg.pmMsg, "pm-msg", "", "private message body (use with -pm-to)")
	flag.StringVar(&cfg.addFriend, "add-friend", "", "after login, add this player to friends list (required before PMs can be sent/received)")
	flag.StringVar(&cfg.combatTarget, "combat-loop", "", "run kill→loot→bury reactor targeting this NPC name (e.g., 'Goblin', 'Man') until -dwell expires or kill cap reached")
	flag.IntVar(&cfg.combatKills, "combat-kills", 5, "max kills the combat-loop will perform before stopping (0 = unlimited)")
	flag.BoolVar(&cfg.lookAround, "look-around", false, "after login, print an LLM-style observation report of the bot's surroundings")
	flag.IntVar(&cfg.lookRadius, "look-radius", 10, "radius (tiles) for -look-around")
	flag.DurationVar(&cfg.dwell, "dwell", 5*time.Second, "how long to stay logged in after the optional walk/command")
	flag.BoolVar(&cfg.watch, "watch", false, "log all events received from the server during dwell")
	flag.BoolVar(&cfg.look, "look", false, "after login, log scenery/NPCs known to be near our position (facts-derived)")
	flag.StringVar(&cfg.factsRoot, "facts", "/Users/flint/Code/openrsc", "OpenRSC source root for static facts; empty disables")
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

	level := slog.LevelInfo
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

	host := runtime.New(runtime.Options{
		Server:    cfg.server,
		Username:  cfg.username,
		Password:  cfg.password,
		Facts:     loadedFacts,
		Landscape: loadedLandscape,
		Logger:    log,
	})
	defer host.Close()

	log.Info("connecting", "server", cfg.server)
	if err := host.Connect(rootCtx); err != nil {
		return err
	}

	// Subscribe to all events for the watch UI / debug logging.
	watchCh := host.Bus().Subscribe("*", 256)
	go watchEvents(log, watchCh, cfg.watch)

	// Run the host's main loop in a goroutine; the rest of this
	// function drives the script.
	hostDone := make(chan error, 1)
	go func() {
		hostDone <- host.Run(rootCtx)
	}()

	// Give the server a moment to send initial state, including the
	// first tick of NPC/player position data the world-state mirror
	// needs for the walk-hint logic in Host.AttackNpc/TalkToNpc.
	time.Sleep(3 * time.Second)

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

	if cfg.tradeAccept {
		log.Info("accepting pending trade request")
		if err := host.AcceptIncomingTrade(rootCtx); err != nil {
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

	if cfg.combatTarget != "" {
		opts := runtime.DefaultCombatLoopOptions()
		opts.MaxKills = cfg.combatKills
		// Look up the requested type by name.
		opts.TargetTypeIDs = nil
		if host.Facts() != nil {
			wantName := strings.ToLower(cfg.combatTarget)
			for id, def := range host.Facts().NpcDefs {
				if def == nil {
					continue
				}
				if strings.EqualFold(def.Name, cfg.combatTarget) ||
					strings.Contains(strings.ToLower(def.Name), wantName) {
					if def.Attackable {
						opts.TargetTypeIDs = append(opts.TargetTypeIDs, id)
					}
				}
			}
		}
		if len(opts.TargetTypeIDs) == 0 {
			return fmt.Errorf("combat-loop: no attackable NPC matches %q", cfg.combatTarget)
		}
		log.Info("starting combat loop",
			"target", cfg.combatTarget,
			"resolved_type_ids", opts.TargetTypeIDs,
			"max_kills", opts.MaxKills,
		)
		combatCtx, cancel := context.WithTimeout(rootCtx, cfg.dwell)
		err := host.CombatLoop(combatCtx, opts)
		cancel()
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			log.Warn("combat loop ended", "err", err)
		} else {
			log.Info("combat loop ended")
		}
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

	log.Info("logging out")
	if err := host.Logout(rootCtx); err != nil {
		log.Warn("logout send failed", "err", err)
	}

	select {
	case <-rootCtx.Done():
	case <-time.After(500 * time.Millisecond):
	case err := <-hostDone:
		if err != nil && err != context.Canceled {
			return err
		}
	}

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
