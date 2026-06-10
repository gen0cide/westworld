package runtime

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// MesaDirector is the autonomous agent loop: each turn it snapshots the host's
// game state into a mesaclient.Situation, asks mesa.Act "what do I do now?", and
// turns the returned Move (DSL) into a conductor Intent. mesa owns the planning
// (LLM); the host owns execution (interpreter + pearl gate). This is the seam
// the PROTOCOL.md "Act — the agent step" describes.
type MesaDirector struct {
	client mesaclient.Client
	hostID string
	// goalMu guards the per-session lifecycle goal (d.goal). The director
	// goroutine ADVANCES d.goal in resolveGoal(mutate=true) when the active goal
	// closes; the SPEECH goroutine READS it via effectiveGoalView → resolveGoal(
	// mutate=false) (effectiveGoalForSpeech, so the ASK tier slices blockers off
	// the SAME goal the director plans against). Those two goroutines touched
	// d.goal unguarded — a data race (H8 regression). An RWMutex keeps the read
	// path (the common case, many per turn) cheap while serialising the rare
	// advance write.
	goalMu sync.RWMutex
	goal   string
	log    *slog.Logger
	turn   int

	// sub + transcript give the planner MEMORY across turns: a rolling log of
	// narrative events (NPC speech, system messages, items/levels gained) drained
	// from the bus each turn. Without this, what an NPC said during converse() is
	// gone by the next turn and the host re-greets in a loop.
	sub        <-chan event.Event
	transcript []string

	// stuck detection: if she hasn't moved for several turns, her approach isn't
	// working — widen perception and tell her to explore instead of repeating.
	prevX, prevY int
	hasPrev      bool
	stuckTurns   int
	// failStreak counts consecutive FAILED outcomes (anti-stuck v0): a host can
	// loop while still moving, repeating an action that keeps failing. A
	// successful turn resets it, so a working grind is never flagged. See
	// situation(); thresholds are antiStuckSoftFails / antiStuckHardFails.
	failStreak int

	// Spin detector (Phase-5a): the failStreak path only catches FAILING loops; a
	// host can also SUCCEED every turn while re-deriving substantially the same
	// plan forever (the live pickaxe bug — micro-actions succeed, the goal never
	// closes). lastPlanFP is the content fingerprint of the last EXECUTED plan;
	// spinCount counts successive OK turns producing the SAME fingerprint. At
	// antiStuckSpinTurns it fires the BLOCKED graph-write + a SPINNING trigger
	// nudging a done/abandon declaration or a different tack, then re-arms.
	lastPlanFP uint64
	spinCount  int

	// worldStall is set by the HybridDirector's world-progress detector (NoteStall):
	// the number of consecutive turns the host's coarse world state (position,
	// fatigue, hp, inventory, total xp) has NOT changed. It catches a loop the
	// position-only stuckTurns and the plan-fingerprint spinCount both miss — a host
	// that keeps "succeeding" at a routine that accomplishes nothing (the live
	// 100%-fatigue-at-the-bank loop). >= antiStuckWorldStall surfaces the STUCK
	// hint + a STALLED trigger so the planner re-plans differently or names the
	// unknown blocking it. Set on the conductor goroutine; read in situation().
	worldStall int

	// lastPlayerMsg pins the most recent thing a real player said to her for a
	// few turns and surfaces it to the PLANNER (not just the chat reflex), so a
	// player's directions ("go through the north door") actually steer her.
	// lastPlayerName is who said it, so the planner can weigh the trust ledger
	// (a known friend's "follow me" lands differently than a stranger's).
	lastPlayerMsg  string
	lastPlayerName string
	playerMsgAge   int

	// visited is a BOUNDED recency set of the last visitedCap tiles she has stood
	// on — used to flag doors she has ALREADY gone through (a door adjacent to a
	// recently-visited tile leads BACK) so she stops backtracking through the same
	// door. A bounded ring (not the old unbounded lifetime map) keeps the per-turn
	// director path host-LIGHT: an open-ended wanderer otherwise accreted tens of
	// thousands of tiles for its whole life (M14). doorUsed only needs a small
	// recency window, so the cap costs nothing.
	visited *tileRing

	// keywordLadder is the session-genesis attention ladder: words/people that
	// should catch her attention in others' chat, each with a tier + reflex.
	// Surfaced to the Act planner so she orients to her name, friends, trade
	// words, and goal topics. Set once at login from the genesis compile.
	keywordLadder []mesaclient.KeywordRung

	// Aspiration ladder (aspirations.go): pendingAspirations/pendingServes carry
	// the genesis-emitted month-horizon aspiration labels (and which one the
	// opening goal serves) until ensureAspirations seeds them into the goal graph
	// on the turn path — after the limbic warm-start, never at construction.
	// aspirationsReady latches once the portfolio is CONFIRMED present (or
	// confirmed underivable) so the per-turn probe costs one bool.
	pendingAspirations []string
	pendingServes      string
	aspirationsReady   bool
}

// SetKeywordLadder installs the session-genesis attention ladder (called once at
// login, before the conductor starts).
func (d *MesaDirector) SetKeywordLadder(ladder []mesaclient.KeywordRung) {
	d.keywordLadder = ladder
}

// idleIntentLabel marks an idle/no-op intent (d.idle). The spin detector must
// ignore these: an idle is the ABSENCE of a plan, not a re-derived one, so
// legitimate idling — including the director's OWN error/empty-verb idle
// fallbacks — must not accrue spinCount and fire SPINNING (H4).
const idleIntentLabel = "act:idle"

// planFingerprint is a content hash of an EXECUTED plan (the Intent the conductor
// just ran), used by the spin detector to tell "the same plan again" from "a new
// plan". FNV-1a over the Label, the sorted Args, and the whitespace-normalized
// Source — so cosmetic reformatting doesn't read as a different plan, but a real
// change does. Returns 0 for an empty OR idle intent (first turn / no-op / a
// legit wait) so those never false-trip the detector (fp==0 resets spinCount).
func planFingerprint(in Intent) uint64 {
	if (in.Label == "" && in.Source == "") || in.Label == idleIntentLabel {
		return 0
	}
	args := make([]string, len(in.Args))
	for i, a := range in.Args {
		if a != nil {
			args[i] = a.Display()
		}
	}
	sort.Strings(args)
	hh := fnv.New64a()
	hh.Write([]byte(in.Label))
	hh.Write([]byte{0})
	hh.Write([]byte(in.RoutinePath))
	hh.Write([]byte{0})
	for _, a := range args {
		hh.Write([]byte(a))
		hh.Write([]byte{0})
	}
	hh.Write([]byte(strings.Join(strings.Fields(in.Source), " ")))
	return hh.Sum64()
}

const transcriptCap = 80 // narrative lines retained; the last ~18 feed each turn

// Anti-stuck v0 failure-streak thresholds (consecutive failed outcomes).
const (
	antiStuckSoftFails = 2 // nudge: "reconsider the approach"
	antiStuckHardFails = 4 // override: "abandon this approach"
)

// antiStuckSpinTurns is the spin threshold: the same executed plan this many
// successive OK turns means the host is re-deriving one plan and getting
// nowhere (a SUCCEEDING loop the fail-streak path misses). 3 is the smallest N
// that distinguishes deliberate iteration from a stuck loop.
const antiStuckSpinTurns = 3

// antiStuckWorldStall is the world-progress stall threshold: this many consecutive
// turns with NO change to the coarse world state (position/fatigue/hp/inventory/xp)
// means the host is making zero progress regardless of what plan it runs — the
// loop the cheap-loop replay + plan-fingerprint spin both miss. Surfaces the STUCK
// hint + a STALLED trigger. Matches the HybridDirector's stallEscalateTurns.
const antiStuckWorldStall = 5

// NoteStall records the HybridDirector's world-progress stall count (consecutive
// no-change turns). Called each turn before the director plans, on the conductor
// goroutine. situation() turns a high count into a STUCK hint + STALLED trigger.
func (d *MesaDirector) NoteStall(turns int) { d.worldStall = turns }

// enablerProgressFloor nudges a freshly-unblocked goal off a frozen 0 so the
// inspector + planner read an un-block as movement, not a stall.
const enablerProgressFloor = 0.25

// NewMesaDirector builds a director that drives the host toward goal via mesa.Act.
func NewMesaDirector(client mesaclient.Client, hostID, goal string, log *slog.Logger) *MesaDirector {
	if log == nil {
		log = slog.Default()
	}
	return &MesaDirector{client: client, hostID: hostID, goal: goal, log: log}
}

// Next implements Director: build the situation, call Act, convert the Move.
// It logs the full thought stream — what Delores perceives, what she decides,
// and the exact DSL she authors — so a run is fully observable.
func (d *MesaDirector) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	d.ensureSub(h)
	d.drainTranscript(h)   // fold this turn's narrative events into the rolling memory
	d.ensureAspirations(h) // seed/confirm the aspiration portfolio (aspirations.go)
	sit := d.situation(h, last)

	// --- perception: what she sees this turn ---
	d.log.Info("act ① perceives",
		"trigger", sit.Trigger,
		"pos", fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		"hp", fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		"inv", orNone(sit.World.InvSummary),
		"npcs", orNone(sit.World.NearbyNpcs),
		"dialog_open", sit.Hints["dialog_options"] != "",
		"dialog_options", sit.Hints["dialog_options"],
		"recalled_episodes", h.journalLen(),
		"nearby_players", sit.Hints["nearby_players"],
	)
	for _, m := range sit.Recent {
		d.log.Info("act ① recent message", "text", m)
	}
	if last.Err != nil {
		d.log.Info("act ① last error", "err", last.Err.Msg)
	}

	move, err := d.client.Act(ctx, sit)
	if err != nil {
		// Keep the loop alive on a planner error — pause and retry next turn.
		d.log.Warn("act ✗ planner failed; idling", "err", err)
		return d.idle(3), true
	}

	// Phase-5a goal lifecycle: apply the planner's DECLARATION about the active
	// goal (done / abandoned / adopt / progress). This is the ONLY place both the
	// Move and d.effectiveGoal(h) are in hand. The host writes the lifecycle state
	// the planner NAMED — it does not judge satisfaction itself (memory, not a
	// solver). The judgment becomes a durable graph write, fixing the live bug
	// where "the goal is met" was reasoned but never recorded.
	d.applyGoalOp(h, move)

	// --- decision: her reasoning + the DSL she wrote ---
	d.log.Info("act ② decides", "kind", moveKindName(move.Kind), "reasoning", move.Reasoning)
	if move.DSLSource != "" {
		d.log.Info("act ② authored DSL:\n" + move.DSLSource)
	}
	if move.Verb != "" {
		d.log.Info("act ② direct action", "verb", move.Verb, "args", strings.Join(move.ActionArgs, ", "))
	}

	// Publish the full thought to the bus so the debug control plane streams it
	// live (/ws, the browser dashboard) and records it (/events, JSONL).
	d.turn++
	h.bus.Publish(event.AgentThought{
		Turn: d.turn, Trigger: sit.Trigger, Goal: sit.Goal,
		Pos:        fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		HP:         fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		Perception: strings.Join(sit.Recent, " | "),
		Reasoning:  move.Reasoning, MoveKind: moveKindName(move.Kind), DSL: move.DSLSource,
	})
	return d.moveToIntent(move), true
}

func orNone(xs []string) string {
	if len(xs) == 0 {
		return "(none)"
	}
	return strings.Join(xs, ", ")
}

// memoryHint renders the salient durable episodes for the Situation — what she
// has done this life (across restarts), so she doesn't redo finished steps and
// the planner has continuity the transcript ring (which resets) cannot give.
func (d *MesaDirector) memoryHint(h *Host) string {
	eps := h.journal.Salient(8)
	if len(eps) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Things you remember doing (durable across this life — don't redo what's already done):")
	for _, e := range eps {
		b.WriteString("\n- " + e.Text)
	}
	return b.String()
}

// nearbyPlayersHint lists players within reach as optional social opportunities,
// each tagged with her standing relationship (trust grade, familiarity, notes).
// px,py is her current position. Empty when no one is around.
func (d *MesaDirector) nearbyPlayersHint(h *Host, px, py int) string {
	const radius = 16
	parts := make([]string, 0, 4)
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-px)+absInt(p.Y-py) > radius {
			continue
		}
		who := d.trustNote(h, p.Name)
		if who == "" {
			who = "a stranger so far"
		}
		parts = append(parts, fmt.Sprintf("%s (to your %s @ %d,%d — %s)",
			p.Name, bearingFrom(px, py, p.X, p.Y), p.X, p.Y, who))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// attentionHint renders the genesis keyword ladder for the Situation: the
// words/people she should notice in chat, grouped with their default reflex.
func (d *MesaDirector) attentionHint() string {
	if len(d.keywordLadder) == 0 {
		return ""
	}
	parts := make([]string, 0, len(d.keywordLadder))
	for _, r := range d.keywordLadder {
		if r.Keyword == "" {
			continue
		}
		s := "\"" + r.Keyword + "\""
		if r.Action != "" {
			s += " → " + r.Action
		} else if r.Tier != "" {
			s += " (" + r.Tier + ")"
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// trustNote summarizes what the host knows of a player from the trust ledger —
// grade, familiarity, and any tags — or "" for someone she's never met. This is
// the substrate for judging an offer by the RELATIONSHIP rather than a keyword.
func (d *MesaDirector) trustNote(h *Host, name string) string {
	if h.ledger == nil || name == "" || !h.ledger.Known(name) {
		return ""
	}
	rel := h.ledger.Rel(name)
	parts := make([]string, 0, 5)
	parts = append(parts, rel.Grade.String())
	if rel.Familiar > 0 {
		parts = append(parts, fmt.Sprintf("met %d×", rel.Familiar))
	}
	// Multi-axis colour, only when an axis is meaningfully off neutral (terse).
	if rel.Affinity >= 0.3 {
		parts = append(parts, "warm")
	} else if rel.Affinity <= -0.3 {
		parts = append(parts, "cold")
	}
	if rel.Grievance >= 0.3 {
		parts = append(parts, "you hold a grudge")
	}
	parts = append(parts, rel.Tags...)
	return strings.Join(parts, ", ")
}

// ensureSub lazily subscribes to the bus (once) so the director can build a
// rolling narrative transcript across turns.
func (d *MesaDirector) ensureSub(h *Host) {
	if d.sub == nil {
		d.sub = h.bus.Subscribe("*", 4096)
	}
}

// drainTranscript pulls all currently-buffered bus events and appends the
// narrative-salient ones (NPC speech, system messages, items/levels) to the
// transcript, de-duplicating adjacent repeats.
func (d *MesaDirector) drainTranscript(h *Host) {
	for {
		select {
		case ev, ok := <-d.sub:
			if !ok {
				d.sub = nil
				return
			}
			if pc, isChat := ev.(event.OtherPlayerChat); isChat {
				if name := d.playerName(h, pc.PlayerIndex); !strings.EqualFold(name, d.hostID) {
					d.lastPlayerMsg = name + ": " + strings.TrimSpace(stripColors(pc.MessageText))
					d.lastPlayerName = name
					d.playerMsgAge = 0
				}
			}
			line := d.narrativeLine(h, ev)
			if line == "" {
				continue
			}
			if n := len(d.transcript); n > 0 && d.transcript[n-1] == line {
				continue
			}
			d.transcript = append(d.transcript, line)
			if len(d.transcript) > transcriptCap {
				d.transcript = d.transcript[len(d.transcript)-transcriptCap:]
			}
		default:
			return
		}
	}
}

// narrativeLine formats a bus event as a transcript line, or "" to skip it.
func (d *MesaDirector) narrativeLine(h *Host, ev event.Event) string {
	switch e := ev.(type) {
	case event.SystemMessage:
		return strings.TrimSpace(stripColors(e.Message))
	case event.NpcDialogText:
		return "NPC: " + strings.TrimSpace(stripColors(e.Text))
	case event.NpcChat:
		return "NPC: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.OtherPlayerChat:
		name := d.playerName(h, e.PlayerIndex)
		if strings.EqualFold(name, d.hostID) {
			return "" // her OWN chat echoed back as local chat — never treat as someone talking to her
		}
		return name + " says to you: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.ChatReceived:
		if e.Speaker != "" {
			return e.Speaker + ": " + strings.TrimSpace(stripColors(e.Message))
		}
	case event.ItemGained:
		return fmt.Sprintf("(received %s ×%d)", d.itemName(h, e.ItemID), e.Count)
	case event.LevelUp:
		return fmt.Sprintf("(leveled up — skill %d is now %d)", int(e.Skill), e.NewLevel)
	case event.PolicyVeto:
		return fmt.Sprintf("⛔ YOUR OWN NATURE blocked '%s': %s — this action will NEVER work for you; do something else.", e.Action, e.Reason)
	}
	return ""
}

// moveToIntent turns a mesa Move into a conductor Intent.
func (d *MesaDirector) moveToIntent(m *mesaclient.Move) Intent {
	switch m.Kind {
	case mesaclient.MoveWriteRoutine:
		name := m.RoutineName
		if name == "" {
			name = "act_step"
		}
		return Intent{Label: "act:" + name, Name: name, Source: m.DSLSource}
	case mesaclient.MoveRunRoutine:
		return Intent{Label: "act:" + m.RoutinePath, RoutinePath: m.RoutinePath, Args: toValues(m.Args)}
	case mesaclient.MoveDirectAction:
		// Guard: an empty/whitespace verb must NOT become `act_direct() { () }` —
		// that parses, "completes" in 0s doing nothing, and (pre-OneShot) used to be
		// cached + replayed forever. Treat a verbless direct action as a brief idle.
		if strings.TrimSpace(m.Verb) == "" {
			d.log.Warn("act ✗ direct action with empty verb — idling instead of authoring a no-op")
			return d.idle(2)
		}
		// Wrap a single verb in a one-shot routine so it runs through the gate.
		src := fmt.Sprintf("runtime \"1.0\"\nroutine act_direct() {\n    %s(%s)\n}", m.Verb, dslArgList(m.ActionArgs))
		return Intent{Label: "act:" + m.Verb, Name: "act_direct", Source: src, OneShot: true}
	default: // MoveIdle
		secs := m.IdleSeconds
		if secs <= 0 {
			secs = 2
		}
		return d.idle(secs)
	}
}

func (d *MesaDirector) idle(secs int) Intent {
	src := fmt.Sprintf("runtime \"1.0\"\nroutine act_idle() {\n    wait(%d)\n}", secs)
	return Intent{Label: idleIntentLabel, Name: "act_idle", Source: src, OneShot: true}
}

func (d *MesaDirector) npcName(h *Host, typeID int) string {
	if h.facts != nil {
		if def := h.facts.NpcDef(typeID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return ""
}

func (d *MesaDirector) playerName(h *Host, idx int) string {
	for _, p := range h.world.Players.All() {
		if p.Index == idx && p.Name != "" {
			return p.Name
		}
	}
	return "a player"
}

func (d *MesaDirector) itemName(h *Host, itemID int) string {
	if h.facts != nil {
		if def := h.facts.ItemDef(itemID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return fmt.Sprintf("item#%d", itemID)
}

// --- helpers ----------------------------------------------------------------

func triggerFor(last Outcome) string {
	switch {
	case last.Intent.Label == "": // zero outcome
		return "start"
	case last.BudgetExpired:
		// The turn budget expired mid-work — nothing to recover from; carry on.
		return "continue"
	case last.Err != nil:
		return "recover"
	default:
		return "continue"
	}
}

func lastN(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return xs[len(xs)-n:]
}

func toValues(args []string) []interp.Value {
	if len(args) == 0 {
		return nil
	}
	vs := make([]interp.Value, 0, len(args))
	for _, a := range args {
		vs = append(vs, interp.String(a))
	}
	return vs
}

// dslArgList renders direct-action args as DSL literals: numbers bare, else quoted.
func dslArgList(args []string) string {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if isNumeric(a) {
			parts = append(parts, a)
		} else {
			parts = append(parts, "\""+strings.ReplaceAll(a, "\"", "\\\"")+"\"")
		}
	}
	return strings.Join(parts, ", ")
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if (r == '-' || r == '+') && i == 0 {
			continue
		}
		if r == '.' {
			continue
		}
		return false
	}
	return true
}

// stripColors removes RSC "@xxx@" colour codes (5 chars: @ + 3 + @) and turns
// "%" line breaks into spaces, so message text reads cleanly to the model.
func stripColors(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '@' && i+4 < len(s) && s[i+4] == '@' {
			i += 5
			continue
		}
		if s[i] == '%' {
			b.WriteByte(' ')
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func moveKindName(k mesaclient.MoveKind) string { return k.String() }
