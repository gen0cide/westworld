package runtime

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

// maxConsecutiveReuse caps how many turns in a row a stable situation replays its
// cached routine before the LLM is consulted again to re-validate. It bounds
// runaway loops (a one-shot routine that keeps re-firing) and lets the host
// notice change, at the cost of one Act call every maxConsecutiveReuse turns on a
// long grind. Tunable.
const maxConsecutiveReuse = 8

// HybridDirector is the CHEAP LOCAL LOOP (#16): the L3 layer above mesa.Act. Each
// turn it computes a coarse situation signature and REPLAYS a learned library
// routine for it with NO LLM call; only on a cache miss (a novel situation, or a
// periodic re-validation) does it escalate to the wrapped MesaDirector (Act). A
// successful authored routine is PROMOTED into the library, so the LLM becomes
// the ESCALATION target rather than the per-turn driver — the mechanic that makes
// running 200 hosts viable. The cache self-heals: a replayed routine that fails
// is evicted, and a stable signature is re-checked by the LLM at least every
// maxConsecutiveReuse turns.
type HybridDirector struct {
	mesa Director // the LLM planner (mesa.Act) — escalation only
	lib  *RoutineLibrary
	goal string
	log  *slog.Logger

	lastSig  string // signature that produced the previous turn's intent
	lastKind string // "lib" | "authored" | "other"
	reuseSig string // signature the current replay streak is counting against
	reuseRun int    // consecutive replays of reuseSig
}

// NewHybridDirector wraps a fallback director (the MesaDirector) with the local
// routine cache.
func NewHybridDirector(mesa Director, lib *RoutineLibrary, goal string, log *slog.Logger) *HybridDirector {
	if log == nil {
		log = slog.Default()
	}
	return &HybridDirector{mesa: mesa, lib: lib, goal: goal, log: log}
}

// Next implements Director.
func (d *HybridDirector) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	sig := d.signature(h)

	// Learn from the previous turn: promote a working authored GRIND — but NEVER a
	// one-shot (a single say / direct action / idle), which must not become a
	// cached routine that replays forever. Evict a replayed routine that failed, so
	// we re-author next time. Recovery beyond that is universal (not tied to any
	// one activity): the re-validation cap re-consults the planner every
	// maxConsecutiveReuse turns regardless of what she's doing.
	switch d.lastKind {
	case "lib":
		if !last.OK() {
			d.lib.Evict(d.lastSig)
			d.log.Info("cheap-loop: evicted failing library routine", "size", d.lib.Len())
		}
	case "authored":
		if last.OK() && d.lastSig != "" && last.Intent.Source != "" && !last.Intent.OneShot {
			d.lib.Promote(d.lastSig, last.Intent.Name, last.Intent.Source)
			d.log.Info("cheap-loop: promoted routine to library", "name", last.Intent.Name, "size", d.lib.Len())
		}
	}

	// Reset the replay counter whenever the situation changes.
	if sig != d.reuseSig {
		d.reuseSig, d.reuseRun = sig, 0
	}

	// Replay a cached routine with NO LLM — unless we're due a re-validation.
	if d.reuseRun < maxConsecutiveReuse {
		if e, ok := d.lib.Lookup(sig); ok {
			d.lastSig, d.lastKind = sig, "lib"
			d.reuseRun++
			d.log.Info("cheap-loop: replay library routine (no LLM)", "name", e.Name, "reuse", d.reuseRun)
			return Intent{Label: "lib:" + e.Name, Name: e.Name, Source: e.Source}, true
		}
	}

	// Cache miss (or re-validation due) → escalate to the LLM planner.
	intent, ok := d.mesa.Next(ctx, h, last)
	d.lastSig = sig
	if intent.Source != "" {
		d.lastKind = "authored"
	} else {
		d.lastKind = "other"
	}
	d.reuseRun = 0
	return intent, ok
}

// signature is a COARSE key for the current situation: goal + coarse position +
// nearby NPC types + inventory-fullness bucket + dialog-open flag. Coarse enough
// that a stationary grind (mining the same rock) replays its routine; specific
// enough that moving, a new scene, or an open dialog is a fresh decision. The
// self-healing cache (evict-on-failure + periodic re-validation) absorbs the
// imprecision. Tunable.
func (d *HybridDirector) signature(h *Host) string {
	pos := h.world.Self.Position()
	var b strings.Builder
	b.WriteString(shortHash(d.goal))
	fmt.Fprintf(&b, "|p%d,%d", pos.X/3, pos.Y/3)

	// nearby NPC type ids, sorted + deduped
	seen := map[int]bool{}
	ids := make([]int, 0, 8)
	for _, n := range h.world.Npcs.All() {
		if !seen[n.TypeID] {
			seen[n.TypeID] = true
			ids = append(ids, n.TypeID)
		}
	}
	sort.Ints(ids)
	b.WriteString("|n")
	for _, id := range ids {
		fmt.Fprintf(&b, ",%d", id)
	}

	// inventory fullness bucket + dialog state
	fmt.Fprintf(&b, "|i%d", fullnessBucket(h.world.Inventory.FreeSlots()))
	if opts := h.world.Recent.DialogOptions(); opts != nil && len(opts.Options) > 0 {
		b.WriteString("|dlg")
	}
	return b.String()
}

func fullnessBucket(free int) int {
	switch {
	case free <= 0:
		return 0
	case free <= 3:
		return 1
	case free <= 10:
		return 2
	default:
		return 3
	}
}

func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:6])
}
