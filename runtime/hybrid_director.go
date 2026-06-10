package runtime

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/world"
)

// maxConsecutiveReuse caps how many turns in a row a stable situation replays its
// cached routine before the LLM is consulted again to re-validate. It bounds
// runaway loops (a one-shot routine that keeps re-firing) and lets the host
// notice change, at the cost of one Act call every maxConsecutiveReuse turns on a
// long grind. Tunable.
const maxConsecutiveReuse = 8

// stallEscalateTurns is the world-progress stall threshold: this many consecutive
// turns with NO change to the coarse world state (position/fatigue/hp/inventory/xp)
// means the current routine — replayed or freshly authored — is accomplishing
// nothing. The cheap replay is then SUPPRESSED (and the cached routine evicted) so
// a "completed but pointless" routine can't loop forever, and the planner is
// consulted with a STALLED signal (NoteStall). Matches director antiStuckWorldStall.
const stallEscalateTurns = 5

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

	lastProgress string // coarse world-state key of the previous turn (progressKey)
	stallRun     int    // consecutive turns the world state has NOT changed
}

// NewHybridDirector wraps a fallback director (the MesaDirector) with the local
// routine cache.
func NewHybridDirector(mesa Director, lib *RoutineLibrary, goal string, log *slog.Logger) *HybridDirector {
	if log == nil {
		log = slog.Default()
	}
	return &HybridDirector{mesa: mesa, lib: lib, goal: goal, log: log}
}

// Unwrap exposes the wrapped planner so callers needing the concrete
// *MesaDirector (ask-prioritization, forage targeting, topical questions) can
// reach it through the production wrapping. Without this, a bare
// `c.director.(*MesaDirector)` assertion silently fails under the wrapper and
// every drive keyed on it goes inert — exactly what shipped: the Phase-5b
// forage drive was dead in production while every test (wired with a bare
// MesaDirector) passed.
func (d *HybridDirector) Unwrap() Director { return d.mesa }

// Next implements Director.
func (d *HybridDirector) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	sig := d.signature(h)

	// World-progress stall detector: the coarse signature + the director's
	// plan-fingerprint spin both MISS a loop where the host keeps "succeeding" at a
	// routine that changes nothing (e.g. replaying "go to a bank" that never cures
	// 100% fatigue — the live loop). Track a finer world-state key; if it hasn't
	// changed for stallEscalateTurns turns the approach is futile — suppress the
	// cheap replay below and tell the planner (NoteStall) so it re-plans differently.
	pk := d.progressKey(h)
	if pk == d.lastProgress {
		d.stallRun++
	} else {
		d.stallRun, d.lastProgress = 0, pk
	}
	stalled := d.stallRun >= stallEscalateTurns
	if m, ok := d.mesa.(*MesaDirector); ok {
		m.NoteStall(d.stallRun)
	}

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
			h.publishDecision("cheap-loop", "evict", "evicted a failing learned routine — will re-author next turn")
		}
	case "authored":
		if last.OK() && d.lastSig != "" && last.Intent.Source != "" && !last.Intent.OneShot {
			d.lib.Promote(d.lastSig, last.Intent.Name, last.Intent.Source)
			d.log.Info("cheap-loop: promoted routine to library", "name", last.Intent.Name, "size", d.lib.Len())
			h.publishDecision("cheap-loop", "promote", "learned + cached '"+last.Intent.Name+"' — future turns in this situation replay it with no LLM call")
		}
	}

	// Reset the replay counter whenever the situation changes.
	if sig != d.reuseSig {
		d.reuseSig, d.reuseRun = sig, 0
	}

	// Replay a cached routine with NO LLM — unless we're due a re-validation, OR the
	// world has STALLED. A no-progress routine must not loop forever: evict it (so it
	// stops being the cached answer for this signature) and fall through to a fresh
	// plan, where the STALLED trigger steers the planner to do something different.
	if stalled {
		if _, ok := d.lib.Lookup(sig); ok {
			d.lib.Evict(sig)
			d.log.Info("cheap-loop: evicted stalled routine (no world progress)", "stall", d.stallRun, "size", d.lib.Len())
			h.publishDecision("stall", "evict", fmt.Sprintf("no world progress for %d turns — the current approach is futile; evicted the cached routine and forcing a fresh plan", d.stallRun))
		}
	} else if d.reuseRun < maxConsecutiveReuse {
		if e, ok := d.lib.Lookup(sig); ok {
			d.lastSig, d.lastKind = sig, "lib"
			d.reuseRun++
			d.log.Info("cheap-loop: replay library routine (no LLM)", "name", e.Name, "reuse", d.reuseRun)
			h.publishDecision("cheap-loop", "replay", fmt.Sprintf("replaying learned routine '%s' — no LLM (reuse %d/%d)", e.Name, d.reuseRun, maxConsecutiveReuse))
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
	// Key on the EFFECTIVE goal so a live operator push (mesa GOAL_REVISION)
	// changes the signature → the cached routine for the old goal misses and we
	// re-author against the new one, instead of replaying the stale grind.
	goal := d.goal
	if lg := h.LiveGoal(); lg != "" {
		goal = lg
	}
	b.WriteString(shortHash(goal))
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

// progressKey is a coarse fingerprint of the host's WORLD STATE used to detect a
// no-progress stall: exact position, fatigue, hp, inventory free-slots, and total
// skill XP. ANY meaningful change (a step, a gained/lost item, fatigue moving,
// damage, an XP tick) flips the key and resets the stall counter. Unchanged across
// turns ⇒ whatever she's doing — replaying a cached routine or re-authoring — is
// accomplishing nothing. Cheaper + finer than the situation signature (which is
// deliberately coarse so a stationary grind replays).
func (d *HybridDirector) progressKey(h *Host) string {
	s := h.world.Self
	pos := s.Position()
	xp := 0
	for i := 0; i < world.NumSkills; i++ {
		xp += s.SkillXP(i)
	}
	return fmt.Sprintf("%d,%d|f%d|h%d|i%d|x%d",
		pos.X, pos.Y, s.Fatigue(), s.HP(), h.world.Inventory.FreeSlots(), xp)
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
