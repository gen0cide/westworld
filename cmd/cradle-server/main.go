// Command cradle-server is the host control-plane daemon. It loads a *.hostcfg
// configuration (a single file or a directory of them), then runs and supervises
// the whole fleet in ONE process over shared static resources (facts + landscape
// loaded once, shared by pointer; one mesa connection per host).
//
// This is the daemon core: lifecycle + supervision. The HTTP/JSON control API and
// web UI mount on top of the same registry in a following step.
//
//	cradle-server -config cradle/hostcfg/examples/swarm.hostcfg
//	cradle-server -config ./fleet/        # a directory of *.hostcfg files
//
// Passwords are never in the config — each host names an env var (default
// WESTWORLD_PASSWORD) resolved at launch.
package main

import (
	"context"
	"flag"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gen0cide/westworld/cradle"
	"github.com/gen0cide/westworld/cradle/hostcfg"
)

// build is stamped by scripts/ship.sh (-ldflags "-X main.build=<sha>") so the
// running binary can always be matched against origin/main.
var build = "dev"

func main() {
	var (
		configPath string
		factsRoot  string
		listenAddr string
		ramp       time.Duration
		logLevel   string
	)
	flag.StringVar(&configPath, "config", "", "path to a *.hostcfg file or a directory of them (required)")
	flag.StringVar(&factsRoot, "facts", "/Users/flint/Code/openrsc", "OpenRSC source root for shared static facts; empty disables")
	flag.StringVar(&listenAddr, "listen", "localhost:8099", "HTTP control-plane API (+ web UI) address")
	flag.DurationVar(&ramp, "ramp", 200*time.Millisecond, "delay between host logins (plus small jitter) to spread the login burst")
	flag.StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")
	flag.Parse()

	log := newLogger(logLevel)

	if configPath == "" {
		log.Error("cradle-server: -config is required (a *.hostcfg file or a directory of them)")
		os.Exit(2)
	}

	hosts, err := hostcfg.Load(configPath)
	if err != nil {
		log.Error("cradle-server: failed to load config", "config", configPath, "err", err)
		os.Exit(1)
	}
	if err := hostcfg.ValidateSet(hosts); err != nil {
		log.Error("cradle-server: invalid config", "err", err)
		os.Exit(1)
	}
	log.Info("loaded host configuration", "hosts", len(hosts), "config", configPath)

	deps, closeDeps, err := cradle.BuildSharedDeps(factsRoot, log)
	if err != nil {
		log.Error("cradle-server: failed to build shared deps", "err", err)
		os.Exit(1)
	}
	defer closeDeps()

	ctx, cancel := signalContext()
	defer cancel()

	reg := cradle.NewRegistry(deps, log)
	api := cradle.NewAPI(reg, log) // registers lifecycle hooks BEFORE hosts go live

	// HTTP control plane (JSON API; the web UI mounts here next).
	httpSrv := &http.Server{Addr: listenAddr, Handler: api.Handler()}
	go func() {
		<-ctx.Done()
		sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer scancel()
		_ = httpSrv.Shutdown(sctx)
	}()
	go func() {
		log.Info("cradle-server control plane listening", "addr", listenAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("cradle-server: HTTP server failed", "err", err)
			cancel() // trigger graceful shutdown of the fleet
		}
	}()

	// Start hosts with a ramp + jitter so a big fleet doesn't hit the login server
	// all at once. Runs on its own goroutine so shutdown can interrupt a long ramp.
	go startFleet(ctx, log, reg, hosts, ramp)

	log.Info("cradle-server up", "build", build, "fleet", len(hosts), "ramp", ramp, "listen", listenAddr)

	<-ctx.Done()
	log.Info("cradle-server: shutting down — stopping all hosts")
	reg.StopAll()
	reg.Wait()
	log.Info("cradle-server: all hosts stopped")
}

func startFleet(ctx context.Context, log *slog.Logger, reg *cradle.Registry, hosts []hostcfg.Host, ramp time.Duration) {
	for i, h := range hosts {
		if ctx.Err() != nil {
			return
		}
		if err := reg.Start(ctx, h); err != nil {
			log.Error("cradle-server: failed to start host", "host", h.Name, "err", err)
			continue
		}
		if i < len(hosts)-1 && ramp > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(ramp + jitter(ramp)):
			}
		}
	}
	log.Info("cradle-server: fleet launch complete")
}

// jitter returns a random [0, base/2) padding so logins don't land in lockstep.
func jitter(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(base/2) + 1))
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
