package runtime

import (
	"context"
	"time"

	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// metricsReportInterval is how often the host reports telemetry to mesa.
const metricsReportInterval = 60 * time.Second

// runMetrics is the host's telemetry path: a bus-subscriber goroutine (sibling
// to runLimbic/runMemory) that tallies counters from the bus (pearl vetoes,
// agent turns) and, on a cadence, reports them plus state snapshots (memory
// hit/miss, ledger/journal sizes, uptime) to mesa's structured metrics table —
// observability + cron aggregation inputs. No-op without a mesa seam. Started by
// Run; exits when ctx is cancelled.
func (h *Host) runMetrics(ctx context.Context) {
	ch := h.bus.Subscribe("*", 256)
	t := time.NewTicker(metricsReportInterval)
	defer t.Stop()
	start := time.Now()
	var vetoes, turns int64
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch ev.(type) {
			case event.PolicyVeto:
				vetoes++
			case event.AgentThought:
				turns++
			}
		case <-t.C:
			h.reportMetrics(ctx, vetoes, turns, time.Since(start))
		}
	}
}

// reportMetrics assembles the current telemetry batch and ships it to mesa.
func (h *Host) reportMetrics(ctx context.Context, vetoes, turns int64, uptime time.Duration) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() {
		return
	}
	metrics := []mesaclient.Metric{
		{Name: "host.uptime_seconds", Value: uptime.Seconds()},
		{Name: "pearl.vetoes", Value: float64(vetoes)},
		{Name: "agent.turns", Value: float64(turns)},
		{Name: "journal.episodes", Value: float64(h.journalLen())},
	}
	if h.ledger != nil {
		metrics = append(metrics, mesaclient.Metric{Name: "ledger.relationships", Value: float64(len(h.ledger.All()))})
	}
	if h.Memory != nil {
		snap := h.Memory.Metrics().Snapshot()
		metrics = append(metrics,
			mesaclient.Metric{Name: "memory.hits", Value: float64(sumCounts(snap.Hits))},
			mesaclient.Metric{Name: "memory.misses", Value: float64(sumCounts(snap.Misses))},
			mesaclient.Metric{Name: "memory.journal_depth", Value: float64(snap.JournalDepth)},
		)
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := h.mesaMem.ReportMetrics(cctx, h.opts.Username, metrics); err != nil {
		h.log.Debug("metrics: report failed", "err", err)
		return
	}
	h.log.Info("metrics reported to mesa", "samples", len(metrics), "vetoes", vetoes, "turns", turns)
}

// sumCounts totals a per-namespace/tier counter map.
func sumCounts(m map[string]int64) int64 {
	var n int64
	for _, v := range m {
		n += v
	}
	return n
}
