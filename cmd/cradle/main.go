// Command cradle is a single-bot RSC client built on the runtime.Host
// abstraction.
//
// Usage:
//
//	cradle -server localhost:43596 -username alex -password REDACTED -walk 120,504
//	cradle -username alex -password REDACTED -dwell 30s -watch
//	cradle -username alex -password REDACTED -command "heal"
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/runtime"
)

func main() {
	var (
		server   = flag.String("server", "localhost:43596", "OpenRSC server host:port")
		username = flag.String("username", "alex", "RSC account username")
		password = flag.String("password", "REDACTED", "RSC account password")
		walkArg  = flag.String("walk", "", "optional destination coords as X,Y (e.g., 120,504); single FOV-bounded click")
		walkToArg = flag.String("walkto", "", "like -walk but chunks long journeys into multiple in-FOV segments")
		command  = flag.String("command", "", "optional admin command to send after login (e.g., 'heal')")
		dwell    = flag.Duration("dwell", 5*time.Second, "how long to stay logged in after the optional walk/command")
		watch    = flag.Bool("watch", false, "log all events received from the server during dwell")
		verbose  = flag.Bool("v", false, "debug-level logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if err := run(log, *server, *username, *password, *walkArg, *walkToArg, *command, *dwell, *watch); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, server, username, password, walkArg, walkToArg, command string, dwell time.Duration, watch bool) error {
	rootCtx, cancel := signalContext()
	defer cancel()

	host := runtime.New(runtime.Options{
		Server:   server,
		Username: username,
		Password: password,
		Logger:   log,
	})
	defer host.Close()

	log.Info("connecting", "server", server)
	if err := host.Connect(rootCtx); err != nil {
		return err
	}

	// Subscribe to all events for the watch UI / debug logging.
	watchCh := host.Bus().Subscribe("*", 256)
	go watchEvents(log, watchCh, watch)

	// Run the host's main loop in a goroutine; the rest of this
	// function drives the script.
	hostDone := make(chan error, 1)
	go func() {
		hostDone <- host.Run(rootCtx)
	}()

	// Give the server a moment to send initial state.
	time.Sleep(1 * time.Second)

	if walkArg != "" {
		x, y, err := parseCoord(walkArg)
		if err != nil {
			return fmt.Errorf("parse -walk: %w", err)
		}
		log.Info("walking (single click)", "to", fmt.Sprintf("(%d, %d)", x, y))
		if err := host.Walk(rootCtx, x, y); err != nil {
			return fmt.Errorf("walk: %w", err)
		}
	}

	if walkToArg != "" {
		x, y, err := parseCoord(walkToArg)
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

	if command != "" {
		log.Info("sending admin command", "cmd", command)
		if err := host.Command(rootCtx, command); err != nil {
			return fmt.Errorf("command: %w", err)
		}
	}

	log.Info("dwelling", "for", dwell)
	select {
	case <-rootCtx.Done():
	case <-time.After(dwell):
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
		case event.LogoutConfirm:
			log.Info("server confirmed logout")
		case event.UnknownPacket:
			if watch {
				log.Debug("unknown packet",
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
