// Package cradle is the host control-plane daemon: it loads host configurations,
// runs many full mesa-connected hosts in one process over shared static
// resources, and supervises their lifecycle (start/stop/pause/resume/restart).
package cradle

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/worldmap"
)

// BuildSharedDeps loads the process-wide world singletons ONCE (facts + landscape,
// shared by pointer across every host) and wires the per-host mesa dialer. The
// returned closer releases the Landscape at daemon shutdown — RunHost never
// closes it.
//
// Facts come from the checked-in generated tables (cmd/defsgen) and cannot
// fail to load. A missing landscape only disables pathfinding (a warning),
// matching cmd/host.
func BuildSharedDeps(factsRoot string, log *slog.Logger) (runtime.SharedDeps, func(), error) {
	var (
		loadedFacts     *facts.Facts
		loadedLandscape *pathfind.Landscape
	)
	// Defs/locs are checked-in generated data (cmd/defsgen) — zero file I/O,
	// no openrsc checkout needed. factsRoot now locates only the landscape
	// collision archive, which is still file-loaded.
	loadedFacts = facts.LoadStatic()
	log.Info("loaded world facts (static)", "summary", loadedFacts.Summary())
	if factsRoot == "" {
		log.Warn("no facts root configured; landscape collision unavailable (pathfinding disabled)")
	} else {
		lpath := filepath.Join(factsRoot, "server", "conf", "server", "data", "Authentic_Landscape.orsc")
		if l, err := pathfind.OpenLandscape(lpath); err != nil {
			log.Warn("landscape load failed; pathfinding disabled", "path", lpath, "err", err)
		} else {
			loadedLandscape = l
			log.Info("loaded landscape archive (shared)", "path", lpath)
		}
	}

	// Precompute the WorldOracle once (shared by pointer across every host like
	// Facts/Landscape). It is pure CPU+memory — the caller keeps owning the
	// landscape fd — so closeDeps need not free it. A failure degrades to a nil
	// oracle (the search_map / reachable / survey_map verbs report "no map data
	// loaded"); it is never fatal, matching the landscape-load degrade above.
	var oracle *worldmap.Oracle
	if loadedFacts != nil && loadedLandscape != nil {
		start := time.Now()
		o, err := worldmap.Precompute(loadedFacts, loadedLandscape)
		if err != nil {
			log.Warn("world oracle precompute failed; map perception disabled", "err", err)
		} else {
			oracle = o
			log.Info("precomputed world oracle (shared)",
				"took", time.Since(start).String(),
				"components", o.NumComponents(),
				"transport_edges", o.EdgesLoaded(),
				"transport_skipped", o.EdgesSkipped(),
			)
		}
	}

	deps := runtime.SharedDeps{
		Facts:       loadedFacts,
		Landscape:   loadedLandscape,
		Mesa:        DialMesa,
		Logger:      log,
		WorldOracle: oracle,
	}
	closeDeps := func() {
		if loadedLandscape != nil {
			loadedLandscape.Close()
		}
	}
	return deps, closeDeps, nil
}

// DialMesa is the cradle's mesa dialer: one gRPC connection per host (no pooling),
// authenticated with the host's derived key. It returns the client plus its Close
// so RunHost owns the per-host connection lifecycle. Hosts may target different
// mesa instances — the address comes from each host's own HostConfig.Mesa.
func DialMesa(addr, hostID string) (mesaclient.Client, func() error, error) {
	c, err := mesaclient.NewGRPCClient(addr, mesaclient.HostKey(hostID))
	if err != nil {
		return nil, nil, err
	}
	return c, c.Close, nil
}
