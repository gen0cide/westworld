// Command cradle is a single-bot RSC client.
//
// Phase 0 form: connects to an OpenRSC server, logs in, walks to a
// target coordinate, waits briefly, logs out. End-to-end validation of
// the wire protocol against the live server.
//
// Usage:
//
//	cradle -server localhost:43596 -username alex -password REDACTED -walk 120,504
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

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
	"github.com/gen0cide/westworld/world"
)

func main() {
	var (
		server   = flag.String("server", "localhost:43596", "OpenRSC server host:port")
		username = flag.String("username", "alex", "RSC account username")
		password = flag.String("password", "REDACTED", "RSC account password")
		walkArg  = flag.String("walk", "120,504", "destination coords as X,Y (default Varrock center)")
		dwell    = flag.Duration("dwell", 5*time.Second, "how long to stay logged in after walking")
		verbose  = flag.Bool("v", false, "verbose logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	x, y, err := parseCoord(*walkArg)
	if err != nil {
		log.Error("invalid -walk argument", "err", err)
		os.Exit(2)
	}

	if err := run(log, *server, *username, *password, x, y, *dwell); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, server, username, password string, x, y int, dwell time.Duration) error {
	rootCtx, cancel := signalContext()
	defer cancel()

	log.Info("connecting", "server", server)
	conn, err := session.Dial(rootCtx, server, session.Options{Logger: log})
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	log.Info("logging in", "username", username)
	res, err := conn.Login(rootCtx, session.LoginParams{
		Username:      username,
		Password:      password,
		ClientVersion: 235,
		RSAPublicKey:  v235.DefaultServerRSA(),
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Info("login successful", "response_code", res.ResponseCode)

	conn.Start()

	// Spawn a goroutine to log inbound frames for observability.
	self := world.NewSelf()
	go logInbound(log, conn, self)

	// Heartbeat ticker.
	heartCtx, stopHeart := context.WithCancel(rootCtx)
	defer stopHeart()
	go heartbeatLoop(heartCtx, log, conn)

	// Brief settle before walking — give the server a chance to send
	// initial state.
	time.Sleep(1 * time.Second)

	log.Info("walking", "to", fmt.Sprintf("(%d, %d)", x, y))
	if err := action.Walk(rootCtx, conn, x, y); err != nil {
		return fmt.Errorf("walk: %w", err)
	}

	log.Info("dwelling", "for", dwell)
	select {
	case <-rootCtx.Done():
		log.Info("context cancelled during dwell")
	case <-time.After(dwell):
	}

	log.Info("logging out")
	if err := action.Logout(rootCtx, conn); err != nil {
		log.Warn("logout send failed", "err", err)
	}

	// Brief wait for server's logout ack.
	select {
	case <-rootCtx.Done():
	case <-time.After(500 * time.Millisecond):
	}

	if cerr := conn.Err(); cerr != nil {
		return fmt.Errorf("connection terminated with error: %w", cerr)
	}
	return nil
}

func logInbound(log *slog.Logger, conn *session.Conn, self *world.Self) {
	for frame := range conn.Recv() {
		log.Debug("inbound frame",
			"opcode", fmt.Sprintf("0x%02x (%d)", frame.Opcode, frame.Opcode),
			"payload_bytes", len(frame.Payload),
		)
		switch frame.Opcode {
		case v235.InSendLogout:
			log.Info("server confirmed logout")
		case v235.InSendPlayerCoords:
			// Bitpacked; for Phase 0 we just acknowledge.
			log.Debug("received player-coords update", "bytes", len(frame.Payload))
		}
	}
}

func heartbeatLoop(ctx context.Context, log *slog.Logger, conn *session.Conn) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := action.Heartbeat(ctx, conn); err != nil {
				log.Warn("heartbeat send failed", "err", err)
				return
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
