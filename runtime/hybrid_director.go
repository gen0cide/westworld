package runtime

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/event"
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

// Stall-evidence rings (C-3): bounded history the STALLED stuck-breaker
// (director_situation.go) distills into the Act trigger. The wrapper is the one
// component that sees EVERY turn's Outcome — replayed, authored, and detoured
// alike — and outlives the per-turn flow, so it owns the rings; situation()
// reads them back through the conductor handle (hybridWrapper).
const (
	stallFailLogCap  = 4   // distinct recent failures retained (consecutive repeats fold into a count)
	stallDecTrailCap = 4   // recent Act reasonings retained
	stallReasonLen   = 160 // runes kept per recorded reasoning
)

// outcomeFailure is one remembered routine failure: which intent, what error.
type outcomeFailure struct {
	label  string
	errMsg string
	count  int // consecutive identical failures folded together
}

// noProgressEvictAfter is the cached-routine strike-out (#30a / soak-retro fix 14):
// a library routine whose replay COMPLETES while the world-progress key never moves
// this many turns in a row loses its slot. "Ran without crashing while changing
// nothing" is the cheap-loop disease — it must not keep a routine resident. This
// fires well before the stallEscalateTurns backstop, which remains for loops the
// strike counter can't see (e.g. re-authored-every-turn churn).
const noProgressEvictAfter = 2

// HybridDirector is the CHEAP LOCAL LOOP (#16): the L3 layer above mesa.Act. Each
// turn it computes a coarse situation signature and REPLAYS a learned library
// routine for it with NO LLM call; only on a cache miss (a novel situation, or a
// periodic re-validation) does it escalate to the wrapped MesaDirector (Act). A
// successful authored routine that VISIBLY MOVED THE WORLD (the progressKey
// changed during its run — #30a) is PROMOTED into the library, so the LLM becomes
// the ESCALATION target rather than the per-turn driver — the mechanic that makes
// running 200 hosts viable. The cache self-heals: a replayed routine that fails
// is evicted, one that completes without progress strikes out after
// noProgressEvictAfter consecutive replays, and a stable signature is re-checked
// by the LLM at least every maxConsecutiveReuse turns.
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

	// Promotion-on-progress gate (#30a). runProgress carries "this run already
	// visibly moved the world" across the budget-resume seam: a grind that
	// progressed during an earlier budget-expired stretch keeps its promotion
	// credit even if its final, completing turn shows no new delta. noProgSig/
	// noProgRun count CONSECUTIVE completed-but-zero-progress replays of one
	// cached routine toward strike-out eviction (noProgressEvictAfter).
	runProgress bool
	noProgSig   string
	noProgRun   int

	// C-3 stall evidence: failLog remembers recent FAILED outcomes (label +
	// error, consecutive repeats folded into a count); decTrail remembers the
	// planner's recent Act reasonings, harvested off the bus (AgentThought with
	// Turn>0 — the sub-act publishDecision records carry Turn 0 and are
	// skipped). Both are read back newest-first by the STALLED stuck-breaker
	// (director_situation.go) via hybridWrapper. Written only on the conductor
	// goroutine (Next); read inside the wrapped planner's situation() on that
	// same goroutine (and from the analysis path only while the conductor is
	// frozen), so no locking — matching the rest of this struct.
	failLog  []outcomeFailure
	decSub   <-chan event.Event
	decTrail []string
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

	// C-3 stall evidence: remember what just failed (and why) and the planner's
	// recent reasonings, so a STALLED escalation can show the LLM the evidence
	// instead of a bare "do something different" (assembled in situation()).
	d.noteFailure(last)
	d.harvestDecisions(h)

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
		// Aspiration portfolio: ensure it exists even on cheap-loop replay turns
		// that never reach the wrapped planner's Next (the seeding must not wait
		// for the first LLM escalation).
		m.ensureAspirations(h)
	}

	// Learn from the previous turn: promote a working authored GRIND — but NEVER a
	// one-shot (a single say / direct action / idle), which must not become a
	// cached routine that replays forever. Evict a replayed routine that failed, so
	// we re-author next time. Recovery beyond that is universal (not tied to any
	// one activity): the re-validation cap re-consults the planner every
	// maxConsecutiveReuse turns regardless of what she's doing.
	switch d.lastKind {
	case "lib":
		// A budget-expired replay is NOT a failing routine — the grind simply
		// outlived the turn budget mid-work; evicting it would discard a learned
		// working routine (soak retro #5).
		if !last.OK() && !last.BudgetExpired {
			d.lib.Evict(d.lastSig)
			d.log.Info("cheap-loop: evicted failing library routine", "size", d.lib.Len())
			h.publishDecision("cheap-loop", "evict", "evicted a failing learned routine — will re-author next turn")
			d.noProgSig, d.noProgRun = "", 0
		} else if d.stallRun > 0 {
			// Promotion-on-progress gate (#30a), replay side: the replay ran
			// fine mechanically but the world-progress key never moved — that
			// earns a strike, not credit. noProgressEvictAfter strikes in a row
			// and the routine loses its slot, so a learned-then-stale grind
			// can't squat in the cache being replayed/evicted/re-authored
			// forever (the drone9 509-routines loop).
			if d.lastSig == d.noProgSig {
				d.noProgRun++
			} else {
				d.noProgSig, d.noProgRun = d.lastSig, 1
			}
			if d.noProgRun >= noProgressEvictAfter {
				d.lib.Evict(d.lastSig)
				d.log.Info("cheap-loop: evicted no-progress library routine", "strikes", d.noProgRun, "size", d.lib.Len())
				h.publishDecision("cheap-loop", "evict", fmt.Sprintf("learned routine replayed %d times without moving the world — it no longer works here; evicted, will re-plan", d.noProgRun))
				d.noProgSig, d.noProgRun = "", 0
			}
		} else {
			d.noProgSig, d.noProgRun = "", 0
		}
	case "authored":
		// Promotion-on-progress gate (#30a): "completed without crashing" is not
		// the bar — the routine must have VISIBLY MOVED THE WORLD (the stall
		// detector's progressKey changed: a step, an item, fatigue, hp, xp)
		// during its run to earn a library slot. stallRun==0 means the key
		// flipped this turn; runProgress means it flipped during an earlier
		// budget-resumed stretch of the same run.
		if last.OK() && d.lastSig != "" && last.Intent.Source != "" && !last.Intent.OneShot {
			if d.stallRun == 0 || d.runProgress {
				d.lib.Promote(d.lastSig, last.Intent.Name, last.Intent.Source)
				d.log.Info("cheap-loop: promoted routine to library", "name", last.Intent.Name, "size", d.lib.Len())
				h.publishDecision("cheap-loop", "promote", "learned + cached '"+last.Intent.Name+"' — future turns in this situation replay it with no LLM call")
			} else {
				d.log.Info("cheap-loop: routine completed without world progress — not promoted", "name", last.Intent.Name)
				h.publishDecision("cheap-loop", "no-promote", "'"+last.Intent.Name+"' completed but changed nothing — not cached; a routine must visibly move the world to be learned")
			}
		}
		d.noProgSig, d.noProgRun = "", 0 // a non-replay turn breaks the strike streak
	default:
		d.noProgSig, d.noProgRun = "", 0
	}

	// Reset the replay counter whenever the situation changes.
	if sig != d.reuseSig {
		d.reuseSig, d.reuseRun = sig, 0
	}

	// A pending operator whisper must reach the model: every cheap path below
	// returns WITHOUT a real director turn, so a host in a stable replay
	// groove would not surface the thought for minutes. Skip the cheap paths
	// while whispers wait (review finding, 2026-06-11).
	whisperWait := h.PendingWhispers() > 0
	if whisperWait {
		d.log.Info("cheap-loop: bypassed — operator whisper pending")
	}

	// Budget-expired but PROGRESSING (soak retro #5): the routine outlived the
	// turn budget WHILE the world kept changing (position/fatigue/hp/inventory/xp
	// — the progressKey flipped, so stallRun is 0). That is a WORKING grind cut
	// off by the scheduler, not a decision point: RESUME the same intent with no
	// LLM call. Self-limiting — the first expired turn that changes nothing falls
	// through to the normal stall-aware path below. One-shots can't meaningfully
	// resume; an idle/empty intent has nothing to re-run.
	if !whisperWait && last.BudgetExpired && d.stallRun == 0 && !last.Intent.OneShot &&
		(last.Intent.Source != "" || last.Intent.RoutinePath != "") {
		d.lastSig = sig
		if last.Intent.Source != "" {
			d.lastKind = "authored"
		} else {
			d.lastKind = "other"
		}
		// This run has already moved the world: carry that promotion credit
		// (#30a) so a grind that does its work in budget-resumed stretches is
		// still promoted when its final, completing turn shows no new delta.
		d.runProgress = true
		d.log.Info("cheap-loop: resuming budget-expired grind (world progressed; no LLM)", "intent", last.Intent.Label)
		h.publishDecision("cheap-loop", "resume", "the turn budget expired mid-grind but the world is progressing — resuming '"+last.Intent.Label+"' (no LLM)")
		return last.Intent, true
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
	} else if !whisperWait && d.reuseRun < maxConsecutiveReuse {
		if e, ok := d.lib.Lookup(sig); ok {
			d.lastSig, d.lastKind = sig, "lib"
			d.runProgress = false // fresh run — no carried promotion credit
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
	d.runProgress = false // fresh run — no carried promotion credit
	d.reuseRun = 0
	return intent, ok
}

// noteFailure records a real failed turn into the stall-evidence ring. A
// budget-expired turn is NEUTRAL (a grind outliving its turn is not a failure
// — soak retro #5) and an empty Label means no prior action. The same failure
// repeating folds into a count instead of flooding the ring, so the breaker
// can still show DISTINCT failures alongside "×N".
func (d *HybridDirector) noteFailure(last Outcome) {
	if last.Intent.Label == "" || last.OK() || last.BudgetExpired {
		return
	}
	msg := last.Kind.String()
	if last.Err != nil {
		msg = last.Err.Msg
	}
	if n := len(d.failLog); n > 0 && d.failLog[n-1].label == last.Intent.Label && d.failLog[n-1].errMsg == msg {
		d.failLog[n-1].count++
		return
	}
	d.failLog = append(d.failLog, outcomeFailure{label: last.Intent.Label, errMsg: msg, count: 1})
	if len(d.failLog) > stallFailLogCap {
		d.failLog = d.failLog[len(d.failLog)-stallFailLogCap:]
	}
}

// harvestDecisions drains the bus subscription for Act thoughts and keeps the
// last few reasonings — the in-RAM tail of decisions.jsonl. Lazy subscribe +
// non-blocking drain, the same pattern as MesaDirector.drainTranscript.
func (d *HybridDirector) harvestDecisions(h *Host) {
	if d.decSub == nil {
		if h == nil || h.bus == nil {
			return
		}
		d.decSub = h.bus.Subscribe(event.AgentThought{}.Kind(), 64)
	}
	for {
		select {
		case ev, ok := <-d.decSub:
			if !ok {
				d.decSub = nil
				return
			}
			t, isThought := ev.(event.AgentThought)
			if !isThought || t.Turn == 0 || strings.TrimSpace(t.Reasoning) == "" {
				continue // Turn 0 = sub-act layer records (publishDecision), not Act reasonings
			}
			d.decTrail = append(d.decTrail, clipRunes(t.Reasoning, stallReasonLen))
			if len(d.decTrail) > stallDecTrailCap {
				d.decTrail = d.decTrail[len(d.decTrail)-stallDecTrailCap:]
			}
		default:
			return
		}
	}
}

// recentFailures returns the failed-routine evidence newest-first (a copy).
func (d *HybridDirector) recentFailures() []outcomeFailure {
	out := make([]outcomeFailure, 0, len(d.failLog))
	for i := len(d.failLog) - 1; i >= 0; i-- {
		out = append(out, d.failLog[i])
	}
	return out
}

// recentDecisions returns the recent Act reasonings newest-first (a copy).
func (d *HybridDirector) recentDecisions() []string {
	out := make([]string, 0, len(d.decTrail))
	for i := len(d.decTrail) - 1; i >= 0; i-- {
		out = append(out, d.decTrail[i])
	}
	return out
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
