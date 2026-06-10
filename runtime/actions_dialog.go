package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
)

// NPC dialog / conversation action handler bodies (talk_to / pickpocket
// via npc_command / answer / converse / wait_for_dialog) plus their
// target-resolution and dialog-result helpers. Registered in the central
// actionHandlers table in dsl_actions.go.

func dslTalkTo(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("talk_to takes 1 argument (npc), got %d", len(args))
	}
	idx, ok := h.npcTargetArg(args[0])
	if !ok {
		if _, isStr := args[0].(interp.String); isStr {
			return interp.Fail(interp.TARGET_OUT_OF_VIEW, "talk_to: no visible NPC named "+args[0].Display()), nil
		}
		return nil, errf("talk_to: target must be an npc, an int index, or a name string, got %s", args[0].Kind())
	}
	if err := h.TalkToNpc(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslNpcCommand fires an NPC's primary action command (command1) — e.g.
// "pickpocket" on a Man. Registered as both npc_command and pickpocket.
// The action repeats per call (one attempt each), so loop for several.
func dslNpcCommand(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("npc_command takes 1 argument (npc), got %d", len(args))
	}
	if n, ok := args[0].(*npcView); ok {
		if err := h.NpcCommand(ctx, n.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	if i, ok := interp.AsInt(args[0]); ok {
		if err := h.NpcCommand(ctx, int(i)); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	return nil, errf("npc_command: target must be npc or int, got %s", args[0].Kind())
}

func dslAnswer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("answer takes 1 argument (option index), got %d", len(args))
	}
	idx, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("answer: option index must be int, got %s", args[0].Kind())
	}
	// The DSL dialog index is 1-based (answer(1) = first option), matching
	// find_option's 1-based return. The wire protocol is 0-based, so convert
	// here. 0 is find_option's "no match" sentinel — answering it would
	// silently select the wrong option, so reject it.
	if idx < 1 {
		return interp.Fail(interp.NO_SUCH_ITEM,
			fmt.Sprintf("answer: option index is 1-based (1=first); got %d — did find_option find no match?", idx)), nil
	}
	if err := h.ChooseDialogOption(ctx, int(idx)-1); err != nil {
		return wrapServerErr(err), nil
	}
	// Clear the current menu so a follow-up wait_for_dialog blocks until the
	// NEXT menu actually arrives (smithing/crafting present a chain of menus).
	// If this was the last menu, nothing repopulates and that's fine.
	h.world.Recent.ClearDialogOptions()
	return interp.Ok(interp.Null{}), nil
}

// npcIndexFromArg extracts a server NPC index from a DSL value (an Npc
// view or a raw Int index).
func npcIndexFromArg(v interp.Value) (int, bool) {
	if n, ok := v.(*npcView); ok {
		return n.record.Index, true
	}
	if i, ok := interp.AsInt(v); ok {
		return int(i), true
	}
	return 0, false
}

// nearestVisibleNpcByName returns the server index of the nearest visible
// NPC whose def name matches `name` (case-insensitive; exact match wins
// over substring, then nearest by Chebyshev distance). Lets talk_to /
// converse take a plain name — `converse("banker")` — instead of forcing
// the caller to hand-write a find/nearest over world.npcs.
func (h *Host) nearestVisibleNpcByName(name string) (int, bool) {
	if h.facts == nil {
		return 0, false
	}
	want := strings.ToLower(strings.TrimSpace(name))
	if want == "" {
		return 0, false
	}
	pos := h.world.Self.Position()
	bestIdx := -1
	bestDist := 1 << 30
	bestExact := false
	for _, n := range h.world.Npcs.All() {
		def := h.facts.NpcDef(n.TypeID)
		if def == nil {
			continue
		}
		nm := strings.ToLower(def.Name)
		exact := nm == want
		if !exact && !strings.Contains(nm, want) {
			continue
		}
		d := chebyshev(pos.X, pos.Y, n.X, n.Y)
		if bestIdx == -1 || (exact && !bestExact) || (exact == bestExact && d < bestDist) {
			bestIdx, bestDist, bestExact = n.Index, d, exact
		}
	}
	return bestIdx, bestIdx != -1
}

// npcTargetArg resolves a talk/converse NPC argument that may be an Npc
// view, an Int server index, OR a name string (→ nearest visible NPC of
// that name). Returns the index and whether it resolved.
func (h *Host) npcTargetArg(v interp.Value) (int, bool) {
	if idx, ok := npcIndexFromArg(v); ok {
		return idx, true
	}
	if s, ok := v.(interp.String); ok {
		return h.nearestVisibleNpcByName(string(s))
	}
	return 0, false
}

// dslConverse LISTENS to an NPC: it opens the dialog (talk_to), takes in
// and AGGREGATES everything the NPC says, and advances — but auto-picks
// ONLY choices code can resolve (a lone "continue" prompt, an all-exit
// menu, a banker's bank-access option; see resolveKnownDialogChoice). At
// any REAL multi-option choice it STOPS, leaving the menu open, and returns
// so the host reads what was said and DECIDES with answer(n)/find_option —
// no oracle short-circuit, no blind auto-answer. Takes only the npc arg;
// there is no "pick a topic" argument (NPCs are not queryable — they only
// speak their pre-authored lines).
//
// Returns Ok({ said:[lines], options:[menu]|null, ended:bool, answered:int }):
// options!=null ⇒ a real choice is open and waiting for the host; ended==true
// ⇒ the conversation finished. Fails SERVER_REJECTED if the NPC is busy.
//
//	r = converse(npc)                                  // listen; stop at the first real choice
//	if r.val.options != null {                         // a decision is waiting
//	    answer(find_option("ready")); converse(npc)    // pick, then continue
//	}
//
// isDialogExitOption reports whether a dialog menu option looks like it ENDS or
// advances-past the conversation (vs. asking another question that re-opens the
// menu). Used by converse to avoid looping a question menu forever.
func isDialogExitOption(opt string) bool {
	lo := strings.ToLower(opt)
	for _, p := range []string{
		"goodbye", "bye", "no thank", "that's all", "thats all", "nothing else",
		"i'm ready", "im ready", "ready to go", "ready to leave", "let's go", "lets go",
		"yes please", "yes i'm", "yes im", "ok then", "okay then",
	} {
		if strings.Contains(lo, p) {
			return true
		}
	}
	return false
}

func dslConverse(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("converse takes (npc) only — it LISTENS and stops at each real choice; to pick an option, read what was said then answer(find_option(\"...\")). got %d args", len(args))
	}
	idx, ok := h.npcTargetArg(args[0])
	if !ok {
		if _, isStr := args[0].(interp.String); isStr {
			return interp.Fail(interp.TARGET_OUT_OF_VIEW, "converse: no visible NPC named "+args[0].Display()), nil
		}
		return nil, errf("converse: first arg must be an npc, an int index, or a name string, got %s", args[0].Kind())
	}

	// Remember the latest server message so we can tell a NEW "busy"
	// notice (NPC engaged with another player) from a stale one.
	var preMsgAt time.Time
	if m := h.world.Recent.ServerMessage(); m != nil {
		preMsgAt = m.At
	}

	// Only open the dialog if one isn't already in progress — talking to
	// an NPC while a menu is already up resets/disrupts it. If a prior
	// partial converse left a menu open, we skip straight to answering it.
	if h.world.Recent.DialogOptions() == nil {
		if err := h.TalkToNpc(ctx, idx); err != nil {
			return wrapServerErr(err), nil
		}
	}

	// Speech-aware drive: NPC speech bubbles stream in over several
	// seconds (each line freshens Recent.DialogText), then a menu (or
	// nothing) follows. We answer every menu as it appears and keep
	// waiting WHILE speech is arriving; we only conclude the dialogue is
	// over after a quiet gap with no new speech and no menu — so a slow
	// guide whose menu lands 10s into its monologue isn't missed.
	const (
		poll        = 200 * time.Millisecond
		quietWindow = 3 * time.Second  // no speech + no menu this long → done
		maxTotal    = 40 * time.Second // hard cap per converse
	)
	answered := 0
	var said []string // everything the NPC says this burst, aggregated in order
	seenSaid := map[string]bool{}
	appendSaid := func(t string) {
		if t = strings.TrimSpace(t); t != "" && !seenSaid[t] {
			seenSaid[t] = true
			said = append(said, t)
		}
	}
	start := time.Now()
	lastActivity := start
	var lastSpeechAt time.Time
	if d := h.world.Recent.DialogText(); d != nil {
		appendSaid(d.Text)
		lastSpeechAt = d.At
	}
	for time.Since(start) < maxTotal {
		if m := h.world.Recent.ServerMessage(); m != nil && m.At.After(preMsgAt) &&
			strings.Contains(strings.ToLower(m.Message), "busy") {
			return interp.Fail(interp.SERVER_REJECTED, "converse: npc busy: "+m.Message), nil
		}
		if r := h.world.Recent.DialogOptions(); r != nil {
			// Two-speed decision: auto-advance ONLY a choice code can
			// resolve (a lone "continue", an all-exit menu, a banker's
			// bank-access option). At any REAL choice, STOP and hand back
			// to the host — leave the menu OPEN so the next Act turn
			// decides it alongside the aggregated speech.
			choice, known := resolveKnownDialogChoice(h, r.Options)
			if !known {
				return interp.Ok(dialogResult(said, r.Options, false, answered)), nil
			}
			if err := h.ChooseDialogOption(ctx, choice); err != nil { // wire is 0-based
				return wrapServerErr(err), nil
			}
			h.world.Recent.ClearDialogOptions()
			answered++
			lastActivity = time.Now()
		} else if d := h.world.Recent.DialogText(); d != nil && d.At.After(lastSpeechAt) {
			// New speech bubble — aggregate it; dialogue still in progress.
			appendSaid(d.Text)
			lastSpeechAt = d.At
			lastActivity = time.Now()
		} else if time.Since(lastActivity) > quietWindow {
			break // no menu, no fresh speech — dialogue is finished
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(poll):
		}
	}
	return interp.Ok(dialogResult(said, nil, true, answered)), nil
}

// resolveKnownDialogChoice is the two-speed split applied to NPC dialog: it
// decides whether a menu can be advanced by CODE (no planner/LLM) or is a
// real decision the host must make. Returns the 0-based option index + true
// when it is safe to auto-pick — a lone option ("click to continue"), an
// all-exit menu (nothing to decide), or a known interaction like a banker's
// "access my bank account" (findBankAccessOption). Returns (-1,false) for a
// genuine multi-option choice, so converse stops and surfaces it.
func resolveKnownDialogChoice(h *Host, opts []string) (int, bool) {
	if len(opts) == 0 {
		return -1, false
	}
	if len(opts) == 1 {
		return 0, true // a lone option is "continue" — no real choice
	}
	if i := findBankAccessOption(h); i >= 0 {
		return i, true // banker → open the bank (code knows the right option)
	}
	allExit := true
	for _, o := range opts {
		if !isDialogExitOption(o) {
			allExit = false
			break
		}
	}
	if allExit {
		return 0, true // every option just ends the conversation — nothing to decide
	}
	return -1, false // a genuine multi-option choice — the host decides
}

// dialogResult builds the record converse returns: { said:[lines],
// options:[menu]|null, ended:bool, answered:int }. options!=null means a
// real choice is OPEN and waiting for the host's decision (read `said`,
// then answer(find_option(...))); ended==true means the conversation
// finished with nothing left to choose.
func dialogResult(said, options []string, ended bool, answered int) interp.Value {
	toList := func(ss []string) interp.Value {
		items := make([]interp.Value, len(ss))
		for i, s := range ss {
			items[i] = interp.String(s)
		}
		return &interp.List{Items: items}
	}
	items := map[string]interp.Value{
		"said":     toList(said),
		"ended":    interp.Bool(ended),
		"answered": interp.Int(int64(answered)),
	}
	if options == nil {
		items["options"] = interp.Null{}
	} else {
		items["options"] = toList(options)
	}
	return &interp.Map{Items: items}
}

// dslWaitForDialog polls world.dialog.is_open every 200ms until a
// menu lands or timeout elapses. Default timeout 5s — quest dialogs
// open within 2 server ticks (~1.3s) on average. Returns Bool.
//
// Used as: if wait_for_dialog(8) { answer(find_option("Yes")) }
// to replace the brittle `wait N; if world.dialog.is_open` pattern.
//
// Implementation polls because predicates aren't lazy-evaluated in
// our DSL yet (see deferred general wait_until/repeat_until). The
// poll interval matches the watcher-sweep cadence so dialog events
// surface within one tick of arriving.
func dslWaitForDialog(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	timeoutSec := 5
	if len(args) >= 1 {
		if i, ok := args[0].(interp.Int); ok {
			timeoutSec = int(i)
		}
	}
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for {
		if h.world.Recent.DialogOptions() != nil {
			return interp.Bool(true), nil
		}
		if time.Now().After(deadline) {
			return interp.Bool(false), nil
		}
		select {
		case <-ctx.Done():
			return interp.Bool(false), nil
		case <-time.After(200 * time.Millisecond):
		}
	}
}
