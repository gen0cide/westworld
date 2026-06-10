// health.go is mesad's self-observability + hang/OOM insurance:
//
//   - ensureDeadline gives every LLM-backed handler a server-side ceiling even
//     when the client sent no deadline (legacy hosts) — a handler goroutine can
//     never outlive its budget, so a wedged upstream can't accumulate stuck
//     handlers (the drone7 34-minute-hang class, server half).
//   - StartMemGauge logs heap/goroutine/host gauges on a ticker so the NEXT
//     exit-137 (suspected OOM, 2× in 36h) leaves a memory trail in
//     /tmp/ww-mesad.log instead of dying silently.
package mesad

import (
	"context"
	"runtime"
	"runtime/debug"
	"time"
)

// Server-side handler deadline ceilings. The CLIENT ceilings (mesa/client
// Timeouts) are the primary defense; these are the backstop for clients that
// arrive without a deadline. Each is deliberately a little above its client
// twin so the client deadline fires first (cleaner error attribution), and
// each comfortably bounds the handler's worst case (act: up to 3 LLM attempts).
const (
	actDeadline     = 100 * time.Second
	decideDeadline  = 45 * time.Second
	chatDeadline    = 45 * time.Second // Chat / Ask / AnalysisInterpret / ExtractDialog
	genesisDeadline = 150 * time.Second
)

// ensureDeadline returns ctx unchanged when it already carries a deadline
// (the client's cap propagates over gRPC and wins); otherwise it caps ctx at d.
// The returned cancel must always be called.
func ensureDeadline(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok || d <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}

// StartMemGauge launches a background loop that logs process memory + load
// gauges every `every` (<=0 → 60s) until ctx is done: heap in use, total heap
// allocated, OS-reported memory, goroutine count, registered host count, and
// the configured soft memory limit. Cheap (one ReadMemStats per tick) and
// purely observational — the breadcrumb trail for the next suspected-OOM kill.
func (s *Server) StartMemGauge(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 60 * time.Second
	}
	s.logMemGauge() // baseline line at startup
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.logMemGauge()
			}
		}
	}()
}

// logMemGauge emits one mem-gauge log line (also called once at startup so the
// baseline is on record).
func (s *Server) logMemGauge() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.mu.RLock()
	hosts := len(s.reg)
	s.mu.RUnlock()
	s.log.Info("mem gauge",
		"heap_inuse_mb", m.HeapInuse>>20,
		"heap_alloc_mb", m.HeapAlloc>>20,
		"sys_mb", m.Sys>>20,
		"num_gc", m.NumGC,
		"goroutines", runtime.NumGoroutine(),
		"hosts", hosts,
		"mem_limit_mb", debug.SetMemoryLimit(-1)>>20, // -1 reads without changing
	)
}
