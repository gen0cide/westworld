package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// Analysis mode is the operator-override — a Westworld "full-bypass" takeover of
// a single host. When the configured operator says "!<username> ANALYSIS" in
// local chat (or toggles the cradle control-plane Analysis switch), the host:
//
//   - FREEZES the autonomous conductor (Conductor.Pause): the current routine
//     finishes, no new turn starts.
//   - goes OFFLINE TO THE WORLD: non-operator chat gets no in-character reply.
//   - SUSPENDS all memory writes behind one gate (episodic journal, the LTM
//     mirror downstream of it, and the entire limbic trust/affect ledger) —
//     EXCEPT an explicit operator `remember(...)` command, which is still run.
//
// While active, an operator directive is interpreted with FLAT AFFECT (terse,
// literal, not in-persona) and classified into one of three kinds:
//
//   - COMMAND      — execute the action(s) directly. The directive's DSL runs on
//     a fresh UNGATED interpreter (WithoutPearlGate), bypassing pearl/persona and
//     the conductor turn-loop entirely.
//   - ANSWER       — a factual question about host state, answered truthfully from
//     the live world mirror.
//   - HYPOTHETICAL — "what would you do?" runs her REAL cognition as a DRY RUN: it
//     requests a Move from the EXISTING mesa Act RPC, reports the would-be
//     reasoning + DSL, but DOES NOT execute it.
//
// CLASSIFICATION: explicit operator prefixes (do:/cmd:/ask:/?/would:/"what
// would") and a bare DSL-form verb(...) are resolved HOST-SIDE (deterministic,
// zero-latency). Any other natural-language directive is classified by the mesa
// AnalysisInterpret RPC — a flat-affect (non-persona) interpreter that returns a
// {command,answer,hypothetical} verdict; "what would you do" hypotheticals run
// the REAL Act RPC as a dry run. Offline, the interpreter degrades to the
// deterministic state summary as a best-effort flat answer.

// AnalysisKind classifies an operator directive.
type AnalysisKind string

const (
	AnalysisCommand      AnalysisKind = "command"      // run the DSL directly (full bypass)
	AnalysisAnswer       AnalysisKind = "answer"       // factual question → truthful host-state answer
	AnalysisHypothetical AnalysisKind = "hypothetical" // "what would you do?" → dry-run Act, no execution
	AnalysisControl      AnalysisKind = "control"      // enter/exit/status meta-directive
)

// AnalysisResult is the structured verdict the interpreter returns for one
// directive. It is the wire contract the control-plane endpoint serialises and
// the UI feed renders. Exactly the fields a flat-affect operator console needs:
// what kind of directive it was, the DSL that ran (command) or would run
// (hypothetical), the textual answer/reasoning, and whether anything executed.
type AnalysisResult struct {
	Kind     AnalysisKind `json:"kind"`
	Text     string       `json:"text"`          // the flat answer / reasoning / outcome line
	DSL      string       `json:"dsl,omitempty"` // command: what ran; hypothetical: what would run
	Executed bool         `json:"executed"`      // true ONLY for an executed COMMAND
	Active   bool         `json:"active"`        // analysis-mode state AFTER this directive
	Err      string       `json:"error,omitempty"`
}

// analysisState is the per-host analysis-mode state. It carries the configured
// operator identity (for in-game sender auth), the live mesa client + conductor
// handle (captured at bootstrap so the interpreter can reach Chat/Act and the
// freeze), and the active flag. The flag is mirrored into an atomic.Bool so the
// memory-write bus goroutines (runMemory/runLimbic) can consult it lock-free on
// their hot path while the slower control paths take the mutex.
type analysisState struct {
	mu        sync.Mutex
	active    bool
	activeBit atomic.Bool // lock-free mirror of active, read by memory gates

	operator  string            // in-game username authorized to trigger (from hostcfg); "" => in-game trigger disabled
	mc        mesaclient.Client // mesa transport for Chat/Act; nil => offline (commands still work)
	conductor *Conductor        // the turn loop to freeze/thaw; may be nil (REPL hosts)
	last      *AnalysisResult   // most-recent verdict (for GET state)
}

// configureAnalysis is called once at bootstrap to bind the operator identity,
// the mesa client, and the conductor handle onto the host so both the in-game
// reflex and the control-plane can drive analysis mode against ONE shared state.
func (h *Host) configureAnalysis(operator string, mc mesaclient.Client, c *Conductor) {
	h.analysis.mu.Lock()
	h.analysis.operator = strings.TrimSpace(operator)
	h.analysis.mc = mc
	h.analysis.conductor = c
	h.analysis.mu.Unlock()
}

// AnalysisActive reports whether analysis mode is on. Lock-free (atomic mirror)
// so the memory-write bus goroutines can gate cheaply on their hot path.
func (h *Host) AnalysisActive() bool { return h.analysis.activeBit.Load() }

// AnalysisOperator returns the configured in-game operator identity ("" when no
// operator is configured, i.e. the in-game trigger is disabled for this host).
func (h *Host) AnalysisOperator() string {
	h.analysis.mu.Lock()
	defer h.analysis.mu.Unlock()
	return h.analysis.operator
}

// EnterAnalysis turns analysis mode ON and FREEZES the conductor. Idempotent.
// Returns the resulting state so the control plane can echo it. The conductor
// pause takes effect at the next turn boundary (the current routine finishes).
func (h *Host) EnterAnalysis() AnalysisResult {
	h.analysis.mu.Lock()
	already := h.analysis.active
	h.analysis.active = true
	h.analysis.activeBit.Store(true)
	c := h.analysis.conductor
	h.analysis.mu.Unlock()
	if c != nil {
		c.Pause() // freeze the autonomous turn loop
	}
	text := "analysis mode ENGAGED — conductor frozen, world replies suppressed, memory writes suspended"
	if already {
		text = "analysis mode already engaged"
	}
	h.log.Info("analysis: entered", "already", already)
	res := AnalysisResult{Kind: AnalysisControl, Text: text, Active: true}
	h.analysis.recordLast(res)
	return res
}

// ExitAnalysis turns analysis mode OFF and RESUMES the conductor. Idempotent.
func (h *Host) ExitAnalysis() AnalysisResult {
	h.analysis.mu.Lock()
	already := !h.analysis.active
	h.analysis.active = false
	h.analysis.activeBit.Store(false)
	c := h.analysis.conductor
	h.analysis.mu.Unlock()
	if c != nil {
		c.Resume() // thaw the autonomous turn loop
	}
	text := "analysis mode DISENGAGED — autonomy resumed"
	if already {
		text = "analysis mode was not engaged"
	}
	h.log.Info("analysis: exited", "already", already)
	res := AnalysisResult{Kind: AnalysisControl, Text: text, Active: false}
	h.analysis.recordLast(res)
	return res
}

func (s *analysisState) recordLast(r AnalysisResult) {
	s.mu.Lock()
	cp := r
	s.last = &cp
	s.mu.Unlock()
}

// AnalysisState returns the most-recent directive verdict (for a GET state poll)
// and whether analysis mode is currently active.
func (h *Host) AnalysisState() (last *AnalysisResult, active bool) {
	h.analysis.mu.Lock()
	defer h.analysis.mu.Unlock()
	return h.analysis.last, h.analysis.active
}

// Analyze interprets ONE operator directive while in analysis mode and returns a
// structured verdict. This is the single host-side interpreter both trigger
// paths (in-game reflex + control-plane) route to. It auto-enters analysis mode
// if a directive arrives while off (the control-plane "directive" endpoint and
// cradle-ctl tell are control-plane-authoritative; the in-game path enters
// explicitly on "ANALYSIS" before any directive).
//
// Classification is layered:
//   - FAST PATHS (host-side, deterministic, zero-latency): an explicit prefix —
//     "do:"/"cmd:"/"run:" => command, "ask:"/"?" => answer,
//     "would:"/"hypothetical:"/"what would" => hypothetical — and a bare DSL-form
//     "<verb>(" call => command. These never touch the network.
//   - OTHERWISE the natural-language directive is sent to the mesa
//     AnalysisInterpret RPC (flat affect), and its verdict routed:
//     command => runCommand(dsl), answer => flat AnalysisResult, hypothetical
//     => runHypothetical (the REAL Act dry run). Offline/error => the
//     deterministic state summary as a best-effort flat answer.
func (h *Host) Analyze(ctx context.Context, directive string) AnalysisResult {
	directive = strings.TrimSpace(directive)
	if directive == "" {
		return AnalysisResult{Kind: AnalysisControl, Text: "empty directive", Active: h.AnalysisActive(), Err: "empty directive"}
	}
	if !h.AnalysisActive() {
		h.EnterAnalysis() // a directive implies engagement (control-plane authoritative)
	}

	var res AnalysisResult
	if kind, body, ok := classifyDirective(directive); ok {
		// Fast path: an explicit prefix or a bare DSL call — resolved host-side.
		switch kind {
		case AnalysisCommand:
			res = h.runCommand(ctx, body)
		case AnalysisHypothetical:
			res = h.runHypothetical(ctx, body)
		default: // AnalysisAnswer
			res = h.answerFromState(body)
		}
	} else {
		// Natural-language directive: classify via the mesa interpreter RPC.
		res = h.interpret(ctx, directive)
	}
	res.Active = h.AnalysisActive()
	h.analysis.recordLast(res)
	return res
}

// interpret sends a natural-language operator directive to the mesa Analysis-
// Interpret RPC and routes the flat verdict. Offline / empty verdict / error
// degrade to the deterministic state summary as a best-effort flat answer (the
// operator console never goes dark).
func (h *Host) interpret(ctx context.Context, directive string) AnalysisResult {
	h.analysis.mu.Lock()
	mc := h.analysis.mc
	h.analysis.mu.Unlock()
	if mc == nil || !mc.Healthy() {
		return h.answerFromState(directive) // offline best-effort
	}
	v, err := mc.AnalysisInterpret(ctx, directive, factLines(h.stateSummary()))
	if err != nil || v == nil {
		return h.answerFromState(directive)
	}
	switch strings.ToLower(strings.TrimSpace(v.Kind)) {
	case "command":
		if strings.TrimSpace(v.DSL) == "" {
			return h.answerFromState(directive)
		}
		return h.runCommand(ctx, v.DSL)
	case "answer":
		text := strings.TrimSpace(v.Text)
		if text == "" {
			text = h.stateSummary()
		}
		return AnalysisResult{Kind: AnalysisAnswer, Text: text}
	default: // hypothetical (or unclassifiable → defer to the real planner)
		return h.runHypothetical(ctx, directive)
	}
}

// classifyDirective resolves an operator directive HOST-SIDE when an explicit
// prefix or a bare DSL call makes the kind unambiguous, returning ok=true with
// the kind + the body (prefix stripped). It returns ok=false for any other
// natural-language directive, which the caller then sends to the mesa
// interpreter. Pure + deterministic.
func classifyDirective(s string) (kind AnalysisKind, body string, ok bool) {
	t := strings.TrimSpace(s)
	low := strings.ToLower(t)
	switch {
	case strings.HasPrefix(low, "do:"), strings.HasPrefix(low, "cmd:"), strings.HasPrefix(low, "run:"):
		return AnalysisCommand, strings.TrimSpace(t[strings.IndexByte(t, ':')+1:]), true
	case strings.HasPrefix(low, "ask:"):
		return AnalysisAnswer, strings.TrimSpace(t[strings.IndexByte(t, ':')+1:]), true
	case strings.HasPrefix(t, "?"):
		return AnalysisAnswer, strings.TrimSpace(strings.TrimPrefix(t, "?")), true
	case strings.HasPrefix(low, "would:"), strings.HasPrefix(low, "hypothetical:"):
		return AnalysisHypothetical, strings.TrimSpace(t[strings.IndexByte(t, ':')+1:]), true
	case strings.HasPrefix(low, "what would"):
		return AnalysisHypothetical, t, true
	}
	// No prefix: a bare DSL call (verb(...)) is an unambiguous command.
	if looksLikeDSL(t) {
		return AnalysisCommand, t, true
	}
	// Anything else is free natural language → defer to the mesa interpreter.
	return "", t, false
}

// looksLikeDSL reports whether the text reads as a DSL call: an identifier
// immediately followed by '(' (e.g. go_to("bank"), say("hi"), bank.deposit_all()).
func looksLikeDSL(s string) bool {
	open := strings.IndexByte(s, '(')
	if open <= 0 {
		return false
	}
	head := s[:open]
	for _, r := range head {
		if r == '_' || r == '.' || r == '!' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

// runCommand executes an operator COMMAND directly on the live host via a fresh
// UNGATED interpreter (WithoutPearlGate), fully bypassing pearl/persona and the
// conductor turn-loop. The directive body must be a DSL snippet (a single call
// or a routine body); a bare call is wrapped into a one-shot routine.
func (h *Host) runCommand(ctx context.Context, body string) AnalysisResult {
	src := wrapAnalysisCommand(body)
	res, err := h.RunRoutineSource(ctx, "analysis/cmd", src, nil, WithoutPearlGate())
	out := AnalysisResult{Kind: AnalysisCommand, DSL: body, Executed: true}
	if err != nil {
		out.Executed = false
		out.Err = err.Error()
		out.Text = "command failed to parse/compile: " + err.Error()
		return out
	}
	if res.Err != nil {
		out.Err = res.Err.Error()
		out.Text = "command ran with error: " + res.Err.Error()
		return out
	}
	out.Text = "command executed"
	if res.Value != nil {
		out.Text = "command executed → " + res.Value.Display()
	}
	return out
}

// wrapAnalysisCommand turns an operator directive body into a runnable routine
// source. If the body is already a full routine (declares `routine`), it is run
// as-is; a bare statement / call is wrapped in a one-shot analysis routine with
// the runtime header the parser requires.
func wrapAnalysisCommand(body string) string {
	t := strings.TrimSpace(body)
	if strings.Contains(t, "routine ") && strings.Contains(t, "runtime ") {
		return t
	}
	if strings.Contains(t, "routine ") {
		return "runtime \"1.0\"\n" + t
	}
	return "runtime \"1.0\"\nroutine analysis_cmd() {\n    " + t + "\n}"
}

// runHypothetical runs the host's REAL cognition as a DRY RUN: it builds the same
// Situation the autonomous director would, calls the EXISTING mesa Act RPC, and
// reports the would-be reasoning + DSL WITHOUT executing it (no moveToIntent, no
// bus publish, no turn increment — purely a question answered by the planner).
func (h *Host) runHypothetical(ctx context.Context, body string) AnalysisResult {
	out := AnalysisResult{Kind: AnalysisHypothetical}
	h.analysis.mu.Lock()
	mc := h.analysis.mc
	h.analysis.mu.Unlock()
	if mc == nil || !mc.Healthy() {
		out.Err = "offline"
		out.Text = "hypothetical unavailable: no mesa planner reachable (host is offline)"
		return out
	}
	sit := h.analysisSituation(body)
	move, err := mc.Act(ctx, sit)
	if err != nil {
		out.Err = err.Error()
		out.Text = "planner error: " + err.Error()
		return out
	}
	out.DSL = move.DSLSource
	if out.DSL == "" && move.Verb != "" {
		out.DSL = move.Verb + "(" + strings.Join(move.ActionArgs, ", ") + ")"
	}
	reason := strings.TrimSpace(move.Reasoning)
	if reason == "" {
		reason = "(planner returned no reasoning)"
	}
	out.Text = "WOULD: " + reason + " [DRY RUN — not executed]"
	return out
}

// analysisSituation builds the mesa Situation for a dry-run Act using a throwaway
// MesaDirector (so the snapshot logic — world, affect, scene, memory hints —
// stays in ONE place, director_situation.go). The operator's "what would you do about
// X" body is surfaced as the trigger so the planner reasons about it specifically.
func (h *Host) analysisSituation(body string) *mesaclient.Situation {
	h.analysis.mu.Lock()
	mc := h.analysis.mc
	h.analysis.mu.Unlock()
	d := NewMesaDirector(mc, h.opts.Username, h.analysisGoal(), h.log)
	sit := d.situation(h, Outcome{})
	if t := strings.TrimSpace(body); t != "" {
		sit.Trigger = "OPERATOR DRY-RUN QUERY (report what you WOULD do; this will NOT be executed): " + t
	}
	return sit
}

// analysisGoal returns the host's standing objective for a dry-run situation,
// falling back to a neutral prompt so Act still has something to reason toward.
func (h *Host) analysisGoal() string {
	if h.journal != nil {
		if g, _ := h.journal.Objective(); strings.TrimSpace(g) != "" {
			return strings.TrimSpace(g)
		}
	}
	return "Continue pursuing your current purpose in the world."
}

// answerFromState answers a factual operator directive truthfully from the live
// host state — FLAT affect, no persona, fully deterministic. It is the best-
// effort answer for an offline host and the fallback when the mesa interpreter
// is unreachable or returns nothing. The persona Chat path (which leaked an
// in-character voice) is deliberately gone: an operator console reports facts,
// not roleplay.
func (h *Host) answerFromState(_ string) AnalysisResult {
	return AnalysisResult{Kind: AnalysisAnswer, Text: h.stateSummary()}
}

// stateSummary renders the host's current ground truth as a single flat line:
// position, vitals, fatigue %, coins, a short inventory headline, the standing
// goal, and what she's doing. Truthful, deterministic, persona-free — the
// substrate for a flat factual answer and the facts the interpreter is grounded in.
func (h *Host) stateSummary() string {
	w := h.World()
	if w == nil || w.Self == nil {
		return "no live world state"
	}
	pos := w.Self.Position()
	parts := []string{
		fmt.Sprintf("pos (%d,%d)", pos.X, pos.Y),
		fmt.Sprintf("HP %d/%d", w.Self.HP(), w.Self.MaxHP()),
		fmt.Sprintf("combat lvl %d", w.Self.CombatLevel()),
		fmt.Sprintf("fatigue %d%%", w.Self.FatiguePercent()),
		fmt.Sprintf("coins %d", h.coinCount()),
		fmt.Sprintf("inv %d free", w.Inventory.FreeSlots()),
		"carrying: " + h.inventoryHeadline(),
	}
	if g := strings.TrimSpace(h.analysisGoal()); g != "" {
		parts = append(parts, "goal: "+g)
	}
	if h.analysis.conductor != nil {
		if cur := h.analysis.conductor.CurrentIntent(); cur != "" {
			parts = append(parts, "last routine: "+cur)
		}
	}
	return strings.Join(parts, "; ")
}

// coinCount sums the coin slots in the inventory, resolving the coin item id by
// name via the live facts catalog (the id differs between fixtures and live RSC).
// Returns 0 when facts are unloaded or the host carries no coins.
func (h *Host) coinCount() int {
	w := h.World()
	if w == nil || w.Inventory == nil || h.facts == nil {
		return 0
	}
	def := h.facts.ItemDefByName("Coins")
	if def == nil {
		return 0
	}
	return w.Inventory.Count(def.ID)
}

// inventoryHeadline renders a short, friendly-name summary of the held
// inventory (up to the first few distinct stacks), so an operator answer reads
// "carrying: bronze pickaxe, 5 lobster, ..." rather than raw item ids.
func (h *Host) inventoryHeadline() string {
	w := h.World()
	if w == nil || w.Inventory == nil {
		return "nothing"
	}
	const maxItems = 6
	var items []string
	more := 0
	for _, sl := range w.Inventory.Slots() {
		if sl.ItemID <= 0 {
			continue
		}
		if len(items) >= maxItems {
			more++
			continue
		}
		name := fmt.Sprintf("item#%d", sl.ItemID)
		if h.facts != nil {
			if def := h.facts.ItemDef(sl.ItemID); def != nil && def.Name != "" {
				name = def.Name
			}
		}
		if sl.Amount > 1 {
			name = fmt.Sprintf("%d %s", sl.Amount, name)
		}
		items = append(items, name)
	}
	if len(items) == 0 {
		return "nothing"
	}
	out := strings.Join(items, ", ")
	if more > 0 {
		out += fmt.Sprintf(" (+%d more)", more)
	}
	return out
}

// factLines splits the flat state summary into discrete lines (each fact on its
// own line) for the mesa interpreter's STATE grounding.
func factLines(summary string) []string {
	raw := strings.Split(summary, ";")
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		if t := strings.TrimSpace(r); t != "" {
			out = append(out, t)
		}
	}
	return out
}
