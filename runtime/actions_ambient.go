package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pearl"
)

// Ambient + control-plane action handler bodies: movement, NPC/player
// interaction verbs, social chat, session/admin, the polymorphic
// use()/interact_at() dispatchers, spatial constructors, the timing
// primitives, and the cognition bridge (LLM + memory stdlib).
//
// Registration for all of these lives in the central actionHandlers
// table in dsl_actions.go.

// ---------- movement ----------

func dslWalkTo(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, err
	}
	opts := DefaultWalkOptions()
	// `attempt_open_doors` named arg overrides the default (which is
	// true). Pass `attempt_open_doors=false` to opt out for routines
	// that want strict "fail on any obstacle" semantics — e.g. a
	// scout that should report locked doors rather than barge in.
	if v, ok := named["attempt_open_doors"]; ok {
		if b, ok := v.(interp.Bool); ok {
			opts.AttemptOpenDoors = bool(b)
		} else {
			return nil, errf("walk_to: attempt_open_doors must be a bool, got %s", v.Kind())
		}
	}
	if err := h.WalkToOpts(ctx, x, y, opts); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

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

// ---------- social ----------

func dslSay(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("say takes 1 argument (message), got %d", len(args))
	}
	msg := args[0].Display()
	if err := h.Say(ctx, msg); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslCommand sends a server admin command via the dedicated command
// opcode (38), NOT public chat. The DSL passes the command WITHOUT
// the leading "::" — that prefix is the in-game UI convention; on
// the wire it's a distinct opcode. Common commands: tele <x> <y>,
// summon <name>, blink, invisible. Requires admin permissions on
// the host's account; non-admins get rejected by the server.
func dslCommand(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("command takes 1 argument (cmd), got %d", len(args))
	}
	cmd := args[0].Display()
	if err := h.Command(ctx, cmd); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslWhisper(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var to, msg string
	if v, ok := named["to"]; ok {
		to = v.Display()
	}
	if v, ok := named["message"]; ok {
		msg = v.Display()
	}
	// Positional fallback: whisper(target, message).
	if to == "" && len(args) >= 1 {
		to = args[0].Display()
	}
	if msg == "" && len(args) >= 2 {
		msg = args[1].Display()
	}
	if to == "" || msg == "" {
		return nil, errf("whisper requires target and message")
	}
	if err := h.PrivateMessage(ctx, to, msg); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAddFriend wraps Host.AddFriend — adds a player to the friend
// list so PMs can flow. Returns Result{Null} on success.
func dslAddFriend(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("add_friend takes 1 arg (name), got %d", len(args))
	}
	name, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("add_friend: name must be String, got %s", args[0].Kind())
	}
	if err := h.AddFriend(ctx, string(name)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslFollow wraps Host.Follow — server-side follow of a player view.
// Takes a player view (or string name); the Host method handles
// the lookup. Bang-eligible.
func dslFollow(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("follow takes 1 arg (player view or name), got %d", len(args))
	}
	var name string
	switch v := args[0].(type) {
	case *playerView:
		name = v.record.Name
	case interp.String:
		name = string(v)
	default:
		return nil, errf("follow: target must be a player view or String name, got %s", args[0].Kind())
	}
	if err := h.Follow(ctx, name, 5*time.Second); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslLogout(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.Logout(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslFindOption returns the 1-based index of the first dialog option
// whose text contains `needle` (case-insensitive substring), or 0
// if no match. Returns 0 (not -1) so routines can write
// `answer(find_option("Yes"))` and have the server respond
// gracefully when nothing matches.
//
// Dialog options must have been surfaced via world.dialog.options
// (set by the NPC dialog menu handler); call wait_for_dialog first.
func dslFindOption(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("find_option takes 1 arg (needle text), got %d", len(args))
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("find_option: needle must be String, got %s", args[0].Kind())
	}
	rec := h.world.Recent.DialogOptions()
	if rec == nil {
		return interp.Int(0), nil
	}
	lower := strings.ToLower(string(needle))
	for i, opt := range rec.Options {
		if strings.Contains(strings.ToLower(opt), lower) {
			return interp.Int(int64(i + 1)), nil
		}
	}
	return interp.Int(0), nil
}

// ---------- primitives (no Result wrap; can't fail in the typed sense) ----------

// dslWait sleeps for the given duration (seconds). Returns Null{}
// directly on success — wait is a primitive, not an action; no bang
// variant. Cancellation flows through ctx; a canceled wait returns
// the Go ctx error which becomes a RuntimeError.
func dslWait(ctx context.Context, _ *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("wait takes 1 argument (seconds), got %d", len(args))
	}
	secs, ok := interp.AsFloat(args[0])
	if !ok {
		return nil, errf("wait: expected number, got %s", args[0].Kind())
	}
	if secs <= 0 {
		return interp.Null{}, nil
	}
	t := time.NewTimer(time.Duration(secs * float64(time.Second)))
	defer t.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.C:
		return interp.Null{}, nil
	}
}

// dslWaitUntil blocks until the predicate lambda evaluates truthy
// (or the optional timeout fires). The predicate is a single-arg
// lambda whose argument is ignored — RSC routines write
// `wait_until(_ => self.hp > 1)` and `wait_until(_ => world.bank.is_open, 10)`.
//
// Why a lambda instead of a bare expression: the DSL evaluates
// expressions eagerly at the call site. A bare predicate like
// `wait_until(self.hp > 1)` would compute true/false once and pass
// a Bool — the predicate's actual re-evaluation logic lives here in
// Go. Wrapping in a lambda is the existing "delay this expression"
// convention (cf. filter/map/find).
//
// Poll interval is fixed at 200ms (matches wait_for_dialog).
//
// Returns Bool(true) on satisfied, Bool(false) on timeout. Errors
// from predicate evaluation propagate as RuntimeError.
func dslWaitUntil(ctx context.Context, _ *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("wait_until takes 1 or 2 args (predicate_lambda, timeout?), got %d", len(args))
	}
	pred, ok := args[0].(interp.Callable)
	if !ok {
		return nil, errf("wait_until: first arg must be a lambda (e.g. `_ => self.hp > 1`), got %s", args[0].Kind())
	}
	var deadline time.Time
	if len(args) == 2 {
		secs, ok := interp.AsFloat(args[1])
		if !ok {
			return nil, errf("wait_until: timeout must be a number of seconds, got %s", args[1].Kind())
		}
		if secs > 0 {
			deadline = time.Now().Add(time.Duration(secs * float64(time.Second)))
		}
	}
	const poll = 200 * time.Millisecond
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Lambdas are 1-arg; we pass Null as the ignored param.
		v, err := pred.Call([]interp.Value{interp.Null{}}, nil)
		if err != nil {
			return nil, errf("wait_until predicate: %v", err)
		}
		if interp.Truthy(v) {
			return interp.Bool(true), nil
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return interp.Bool(false), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(poll):
		}
	}
}

// dslNote writes to the host's logger as an info entry. The full
// journal persistence is Phase 3. Primitive: returns Null directly,
// no bang variant.
func dslNote(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("note takes 1 argument (text), got %d", len(args))
	}
	text := args[0].Display()
	h.log.Info("routine note", "text", text)
	// Publish the narration onto the bus so the cradle UI (and any other
	// subscriber) can stream it as a live in-character feed.
	h.bus.Publish(event.RoutineNote{Text: text})
	return interp.Null{}, nil
}

// dslLookAround returns a brain-ready text summary of the scene around
// the host (self vitals, nearby NPCs/players with combat level + threat +
// gear, ground items, notable scenery) — one call a host hands to the LLM
// instead of stitching a dozen accessors. Optional radius (default 10).
func dslLookAround(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	radius := 10
	if len(args) >= 1 {
		if r, ok := interp.AsInt(args[0]); ok {
			radius = int(r)
		}
	}
	return interp.String(h.DescribeSurroundings(radius)), nil
}

// dslGoTo travels the host to a destination anywhere in the world —
// across regions, beyond the local pathfinder window. The destination is
// coords (two ints / a position) or a known TOWN / landmark name
// ("Lumbridge", "Varrock"). It deliberately does NOT resolve a POI TYPE
// ("bank", "mining-site", ...) anymore — that auto-route masked ignorance
// and could walk a host straight into a gate it couldn't pass. To reach a
// type, search_map it (which explains reach per destination) and go_to the
// coords you choose. Backed by Host.GoTo (iterative WalkTo with doors).
func dslGoTo(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var tx, ty int
	if len(args) == 1 {
		if s, ok := args[0].(interp.String); ok {
			if h.facts == nil {
				return interp.Fail(interp.NO_SUCH_ITEM, "go_to: no map data loaded"), nil
			}
			name := string(s)
			pos := h.world.Self.Position()
			g := h.facts.Gazetteer()
			// Only a known TOWN / landmark name resolves. A POI TYPE is no longer
			// auto-routed — steer the host to search_map + go_to(coords) instead.
			p, ok := g.PlaceByName(name, pos.X, pos.Y)
			if !ok {
				return interp.Fail(interp.NO_SUCH_ITEM,
					"go_to: \""+name+"\" is not a known town/place. go_to does NOT take a POI type — call search_map(\""+name+"\") to list reachable destinations, then go_to(their x, y); or go_to(x, y) coords you remember from visiting."), nil
			}
			tx, ty = p.X, p.Y
			if err := h.GoTo(ctx, tx, ty); err != nil {
				return blockedTravelFail(h, name, tx, ty, err), nil
			}
			return interp.Ok(interp.Null{}), nil
		}
	}
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("go_to: %v", err)
	}
	if err := h.GoTo(ctx, x, y); err != nil {
		return blockedTravelFail(h, "", x, y, err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// blockedTravelFail builds a LOUD, re-plannable PATH_BLOCKED failure for a
// go_to that GoTo could not complete. Cognition-first: rather than silently
// auto-rerouting (GoTo opens ordinary doors but cannot pay tolls or pass
// quest gates), it names the resolved target (name + coords), the host's
// CURRENT stuck position, the nearest gazetteer landmark to that stuck tile
// (e.g. "Toll gate"), and a short actionable hint — so the brain perceives
// the block and re-plans. PATH_BLOCKED is set explicitly (not via
// wrapServerErr's fragile substring match) so the LLM can branch on
// r.err.code == "PATH_BLOCKED".
func blockedTravelFail(h *Host, name string, tx, ty int, err error) interp.Value {
	stuck := h.world.Self.Position()
	landmark := ""
	if h.facts != nil {
		if lm, _, ok := h.facts.Gazetteer().NearestPlace(stuck.X, stuck.Y); ok {
			landmark = lm.Name
		}
	}
	target := fmt.Sprintf("(%d, %d)", tx, ty)
	if name != "" {
		target = fmt.Sprintf("%q (%d, %d)", name, tx, ty)
	}
	near := ""
	if landmark != "" {
		near = fmt.Sprintf(" near %s", landmark)
	}
	return interp.Fail(interp.PATH_BLOCKED, fmt.Sprintf(
		"go_to %s blocked: stuck at (%d, %d)%s — a gate may need payment or be quest-locked; pay it if you have coins, pick another destination of this type, or route around (%v)",
		target, stuck.X, stuck.Y, near, err))
}

// dslWhereAmI returns a readable summary of where the host is in the
// world — nearest named area + notable POIs with bearing/distance — so a
// host can reason about its place on the map, not raw coords.
func dslWhereAmI(_ context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	return interp.String(h.LocationSummary()), nil
}

// dslBearingTo returns the 8-point compass direction from the host to a
// target tile (x, y, or a position-like value). "here" if coincident.
func dslBearingTo(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("bearing_to: %v", err)
	}
	pos := h.world.Self.Position()
	if b := facts.Bearing(pos.X, pos.Y, x, y); b != "" {
		return interp.String(b), nil
	}
	return interp.String("here"), nil
}

// dslWhereIs locates a named place ("Lumbridge") or a POI type ("bank",
// "altar", "furnace") relative to the host: distance + bearing + coords.
func dslWhereIs(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("where_is takes 1 arg (place name or POI type), got %d", len(args))
	}
	if h.facts == nil {
		return interp.Null{}, nil
	}
	name := args[0].Display()
	pos := h.world.Self.Position()
	g := h.facts.Gazetteer()
	if p, ok := g.PlaceByName(name, pos.X, pos.Y); ok {
		return interp.String(fmt.Sprintf("%s is %d tiles %s, at (%d, %d)",
			p.Name, chebyshev(pos.X, pos.Y, p.X, p.Y), facts.Bearing(pos.X, pos.Y, p.X, p.Y), p.X, p.Y)), nil
	}
	if p, d, ok := g.NearestPOI(name, pos.X, pos.Y); ok {
		return interp.String(fmt.Sprintf("nearest %s is %d tiles %s, at (%d, %d)",
			name, d, facts.Bearing(pos.X, pos.Y, p.X, p.Y), p.X, p.Y)), nil
	}
	return interp.String("unknown place: " + name), nil
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

// ---------- walk_path / is_reachable — explicit pathfinding ----------

// dslWalkPath dispatches a routine-supplied multi-corner walk.
// Used when the routine has computed its own route (e.g. a
// quest sequence with known corners) rather than asking walk_to
// to pathfind for it. Single packet send; max 25 corners per
// the RSC walk packet (action.MaxWalkCorners). Longer routes
// chunk via repeated walk_path calls.
//
// Accepts:
//
//	walk_path([[103, 532], [105, 525], [110, 522]])
//
// Each element is a 2-element list [x, y]. Returns ErrorCode if
// the packet send fails.
func dslWalkPath(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("walk_path takes 1 arg (list of [x,y] pairs), got %d", len(args))
	}
	list, ok := args[0].(*interp.List)
	if !ok {
		return nil, errf("walk_path: arg must be a list, got %s", args[0].Kind())
	}
	if len(list.Items) == 0 {
		return interp.Ok(interp.Null{}), nil
	}
	corners := make([][2]int, 0, len(list.Items))
	for i, el := range list.Items {
		pair, ok := el.(*interp.List)
		if !ok || len(pair.Items) != 2 {
			return nil, errf("walk_path: element %d must be [x, y], got %s", i, el.Kind())
		}
		x, xok := pair.Items[0].(interp.Int)
		y, yok := pair.Items[1].(interp.Int)
		if !xok || !yok {
			return nil, errf("walk_path: element %d coords must be Int", i)
		}
		corners = append(corners, [2]int{int(x), int(y)})
	}
	if err := action.WalkPath(ctx, h.conn, corners); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslIsReachable runs the local BFS pathfinder from self.position
// to (x, y) and returns true iff a path exists. Pure — no packet
// sent. Bounded by the 96×96 FOV grid; routines that need to
// reason about cross-region routes should sequence is_reachable
// checks along a planned chain.
//
// Accepts:
//
//	is_reachable(x, y)
//	is_reachable(position)
//	is_reachable(view)  — any view with .x/.y
func dslIsReachable(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("is_reachable: %v", err)
	}
	corners, perr := h.pathToTile(x, y, false)
	if perr != nil || len(corners) == 0 {
		return interp.Bool(false), nil
	}
	return interp.Bool(true), nil
}

// ---------- spatial utilities (pure, no opcodes) ----------

// dslDistanceTo returns the Chebyshev distance from self.position
// to the target. Chebyshev (max(|dx|, |dy|)) matches RSC's walk
// cost — one diagonal step = one tile.
//
// Accepts any view with .x/.y (positionView, playerView, npcView,
// groundItemView, placementView, boundaryView), or named (x=X, y=Y).
func dslDistanceTo(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 && (len(named) == 0) {
		return nil, errf("distance_to takes 1 target argument (view or {x,y}), got %d", len(args))
	}
	tx, ty, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("distance_to: %v", err)
	}
	pos := h.world.Self.Position()
	dx := pos.X - tx
	dy := pos.Y - ty
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	d := dx
	if dy > dx {
		d = dy
	}
	return interp.Int(int64(d)), nil
}

// dslDistanceToXY is a positional convenience over distance_to —
// `distance_to_xy(304, 542)` is shorter than `distance_to(x=304, y=542)`
// when the target is a literal tile rather than a view. Chebyshev
// distance (max(|dx|, |dy|)) like the underlying.
func dslDistanceToXY(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("distance_to_xy takes 2 args (x, y), got %d", len(args))
	}
	tx, ok1 := interp.AsInt(args[0])
	ty, ok2 := interp.AsInt(args[1])
	if !ok1 || !ok2 {
		return nil, errf("distance_to_xy: both args must be Int, got %s/%s", args[0].Kind(), args[1].Kind())
	}
	pos := h.world.Self.Position()
	dx := pos.X - int(tx)
	dy := pos.Y - int(ty)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dy > dx {
		return interp.Int(int64(dy)), nil
	}
	return interp.Int(int64(dx)), nil
}

// dslNearestNpc returns the NPC closest to self.position (Chebyshev
// distance), optionally filtered by a 1-arg predicate lambda. It exists
// because world.npcs.find returns the FIRST roster match, not the
// closest — fatal when several NPCs of the same type are in view and
// some are far away or already engaged by another bot. nearest_npc lets
// a combat routine reliably attack the goblin it just spawned adjacent
// to itself rather than a contested wanderer across the field.
//
//	nearest_npc()                    -> closest visible NPC (any), or Null
//	nearest_npc(n => n.type_id == 4) -> closest visible Goblin, or Null
//
// Ties (equal distance) resolve to the first in roster order, which is
// stable across calls within a tick. Returns Null when no NPC matches.
func dslNearestNpc(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) > 1 {
		return nil, errf("nearest_npc takes 0 or 1 args (optional predicate lambda), got %d", len(args))
	}
	var pred interp.Callable
	if len(args) == 1 {
		p, ok := args[0].(interp.Callable)
		if !ok {
			return nil, errf("nearest_npc: arg must be a lambda (e.g. `n => n.type_id == 4`), got %s", args[0].Kind())
		}
		pred = p
	}
	pos := h.world.Self.Position()
	var best *npcView
	bestDist := int(^uint(0) >> 1) // max int
	for _, rec := range h.world.Npcs.All() {
		nv := &npcView{record: rec, facts: h.facts, host: h}
		if pred != nil {
			v, err := pred.Call([]interp.Value{nv}, nil)
			if err != nil {
				return nil, errf("nearest_npc predicate: %v", err)
			}
			if !interp.Truthy(v) {
				continue
			}
		}
		dx := rec.X - pos.X
		if dx < 0 {
			dx = -dx
		}
		dy := rec.Y - pos.Y
		if dy < 0 {
			dy = -dy
		}
		d := dx
		if dy > dx {
			d = dy
		}
		if d < bestDist {
			bestDist = d
			best = nv
		}
	}
	if best == nil {
		return interp.Null{}, nil
	}
	return best, nil
}

// dslInRegion returns true iff self.position is inside the
// rectangle (x1,y1)..(x2,y2) inclusive. Geometry helper. Arg
// order doesn't matter — we normalize min/max.
func dslInRegion(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 4 {
		return nil, errf("in_region takes 4 args (x1, y1, x2, y2), got %d", len(args))
	}
	x1, x2, y1, y2 := intArg(args[0]), intArg(args[2]), intArg(args[1]), intArg(args[3])
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	pos := h.world.Self.Position()
	in := pos.X >= x1 && pos.X <= x2 && pos.Y >= y1 && pos.Y <= y2
	return interp.Bool(in), nil
}

// ---------- bounds shape constructors: box, circle, near ----------
//
// These are pure constructors — they return an interp.RegionPredicate
// value that the bounds-block registration machinery uses as a
// location filter. No server I/O.

// dslBox builds an axis-aligned rectangle predicate. Positional:
// box(x1, y1, x2, y2). Named: box(x1=..., y1=..., x2=..., y2=...).
// Inclusive on all four edges; argument order doesn't matter
// (x1/x2 and y1/y2 are normalized).
func dslBox(_ context.Context, _ *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var x1, y1, x2, y2 int
	switch len(args) {
	case 0:
		x1 = intArg(named["x1"])
		y1 = intArg(named["y1"])
		x2 = intArg(named["x2"])
		y2 = intArg(named["y2"])
	case 4:
		x1, y1, x2, y2 = intArg(args[0]), intArg(args[1]), intArg(args[2]), intArg(args[3])
	default:
		return nil, errf("box(x1, y1, x2, y2): expected 4 positional or 4 named args, got %d positional", len(args))
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	pred := func(x, y int) bool {
		return x >= x1 && x <= x2 && y >= y1 && y <= y2
	}
	name := fmt.Sprintf("box(%d,%d,%d,%d)", x1, y1, x2, y2)
	return interp.NewRegionPredicate(name, pred), nil
}

// dslCircle builds a Chebyshev-distance disk predicate.
// circle(cx, cy, radius) or circle(cx=..., cy=..., radius=...).
// Uses Chebyshev (max of |dx|, |dy|) since RSC movement is grid-8.
func dslCircle(_ context.Context, _ *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var cx, cy, r int
	switch len(args) {
	case 0:
		cx = intArg(named["cx"])
		cy = intArg(named["cy"])
		r = intArg(named["radius"])
	case 3:
		cx, cy, r = intArg(args[0]), intArg(args[1]), intArg(args[2])
	default:
		return nil, errf("circle(cx, cy, radius): expected 3 positional or named args, got %d positional", len(args))
	}
	if r < 0 {
		r = 0
	}
	pred := func(x, y int) bool {
		dx := x - cx
		if dx < 0 {
			dx = -dx
		}
		dy := y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx > dy {
			return dx <= r
		}
		return dy <= r
	}
	name := fmt.Sprintf("circle(%d,%d,r=%d)", cx, cy, r)
	return interp.NewRegionPredicate(name, pred), nil
}

// dslNear builds a disk predicate centered on self.position at
// routine-start time. near(radius) or near(radius=N). Useful for
// "react to events within N tiles of where I started" without
// hard-coding coords.
func dslNear(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var r int
	if v, ok := named["radius"]; ok {
		r = intArg(v)
	} else if len(args) == 1 {
		r = intArg(args[0])
	} else {
		return nil, errf("near(radius): expected 1 positional or named arg")
	}
	if r < 0 {
		r = 0
	}
	pos := h.world.Self.Position()
	cx, cy := pos.X, pos.Y
	pred := func(x, y int) bool {
		dx := x - cx
		if dx < 0 {
			dx = -dx
		}
		dy := y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx > dy {
			return dx <= r
		}
		return dy <= r
	}
	name := fmt.Sprintf("near(%d,%d,r=%d)", cx, cy, r)
	return interp.NewRegionPredicate(name, pred), nil
}

// ---------- interact_at(target, option?) — far-range scenery click ----------

// dslInteractAt fires the primary (option=1, default) or secondary
// (option=2) click on a scenery tile. Generic verb — the actual
// verb dispatched server-side depends on the scenery def's
// Command1 / Command2 fields ("Chop", "Mine", "Climb-Up", etc.).
//
// Accepts:
//
//	interact_at(x=X, y=Y)
//	interact_at(x=X, y=Y, option=2)
//	interact_at(position)              — any view with .x/.y
//	interact_at(scenery_view)          — placement from world.locs
func dslInteractAt(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("interact_at: %v", err)
	}
	option := 1
	if v, ok := named["option"]; ok {
		if i, ok := v.(interp.Int); ok {
			option = int(i)
		}
	} else if len(args) >= 2 {
		if i, ok := args[1].(interp.Int); ok {
			option = int(i)
		}
	}
	if option != 1 && option != 2 {
		return nil, errf("interact_at: option must be 1 or 2, got %d", option)
	}
	if err := h.InteractAt(ctx, x, y, option); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOpenBoundary fires the default open/cross click on a boundary
// (door, gate, fence, web). Takes a boundary view from world.locs.
// The host pathfinds adjacent before sending — same as the existing
// walk-then-act flow.
func dslOpenBoundary(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("open_boundary takes 1 arg (boundary view), got %d", len(args))
	}
	bv, ok := args[0].(*boundaryView)
	if !ok {
		return nil, errf("open_boundary: expected boundary view, got %s", args[0].Kind())
	}
	if err := h.InteractWithBoundary(ctx, bv.x, bv.y, bv.direction); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- use(item, target) — polymorphic interaction ----------

// dslUse dispatches to the right server opcode based on the
// target's view type. Single DSL verb, multiple wire formats:
//
//	use(key, door)        — UseItemOnBoundary (opcode 161)
//	use(needle, cloth)    — UseItemOnItem      (opcode 91)
//	use(log, fire)        — UseItemOnScenery   (opcode 115)  [TODO]
//
// The item arg can be:
//   - itemViewVal (from inventory.find / inventory.slots)
//   - Int (raw item id; we look up the slot)
//   - String (item name; we look up the id then slot)
//
// Resolving the inventory slot here means routines never have to
// pass slot numbers explicitly — the bot finds the item itself.
// If the item isn't in inventory we return NO_SUCH_ITEM.
func dslUse(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("use takes 1 (item) or 2 (item, target) arguments, got %d", len(args))
	}
	itemID, err := resolveItemID(h.facts, args[0])
	if err != nil {
		return nil, errf("use: bad item arg: %v", err)
	}
	slot := -1
	for i, s := range h.world.Inventory.Slots() {
		if s.ItemID == itemID {
			slot = i
			break
		}
	}
	if slot < 0 {
		return interp.Fail(interp.NO_SUCH_ITEM,
			fmt.Sprintf("use: item id %d not in inventory", itemID)), nil
	}
	// Coordinate target: use(item, x=X, y=Y) uses the item on the scenery at that
	// tile — the bridge from the "Object @ (x,y)" the agent perceives to an
	// action (e.g. cooking: use(raw_food, x=RANGEX, y=RANGEY)). cook() is not yet
	// implemented, so this is how skilling-on-scenery is done.
	if xv, okx := named["x"]; okx {
		if yv, oky := named["y"]; oky {
			tx, _ := interp.AsInt(xv)
			ty, _ := interp.AsInt(yv)
			if err := h.UseItemOnScenery(ctx, int(tx), int(ty), slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		}
	}
	// No-target form: use(item) fires the item's own inventory command
	// (ITEM_COMMAND, opcode 90). This is how the sleeping bag (item
	// 1263) starts the fatigue/sleep flow — the server's OpInv trigger
	// runs the item's command and replies with SEND_SLEEPSCREEN. The
	// cradle then auto-answers the sleep word. (Beds are scenery: a bed
	// is used via interact_at(...) / use(item, scenery) on the bed tile,
	// whose "rest"/"sleep"/"lie in" command routes to the same
	// sendEnterSleep — see Sleeping.onOpLoc.)
	if len(args) == 1 {
		if err := h.ItemCommand(ctx, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	// Target dispatch.
	switch t := args[1].(type) {
	case *boundaryView:
		if err := h.UseItemOnBoundary(ctx, t.x, t.y, t.direction, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *itemViewVal:
		// Inventory-on-inventory: find the target's slot too.
		otherSlot := -1
		for i, s := range h.world.Inventory.Slots() {
			if s.ItemID == t.ID && i != slot {
				otherSlot = i
				break
			}
		}
		if otherSlot < 0 {
			return interp.Fail(interp.NO_SUCH_ITEM,
				fmt.Sprintf("use: target item id %d not in a different inventory slot", t.ID)), nil
		}
		if err := h.UseItemOnItem(ctx, slot, otherSlot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *placementView:
		// Result of world.locs.X.nearest() — kind="scenery" /
		// "boundary" / "npc_spawn" decides the opcode. The DSL
		// caller wrote `use(item, world.locs.fires.nearest())`
		// and the dispatch shape is hidden behind the verb.
		switch t.p.Kind {
		case "scenery":
			if err := h.UseItemOnScenery(ctx, t.p.X, t.p.Y, slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		case "boundary":
			if err := h.UseItemOnBoundary(ctx, t.p.X, t.p.Y, t.p.Direction, slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		case "npc_spawn":
			return nil, errf("use: cannot use(item, npc_spawn) — pass the live NPC view from world.npcs, not the spawn placement")
		default:
			return nil, errf("use: unsupported placement kind %q", t.p.Kind)
		}
	case *groundItemView:
		// Inv item on a ground-item. Server needs the ground item
		// type id too (multiple stacks can pile on one tile).
		if err := h.UseItemOnGroundItem(ctx, t.record.X, t.record.Y, t.record.ItemID, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *npcView:
		// Inv item on an NPC (thieving, item-give, quest hand-in).
		if err := h.UseItemOnNpc(ctx, t.record.Index, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *playerView:
		// Inv item on another player (trade-init via "use" gesture,
		// gift-give). Server confirms via TradeRequestReceived event
		// on the target's side.
		if err := h.UseItemOnPlayer(ctx, t.record.Index, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	default:
		return nil, errf("use: unsupported target type %s — expected boundary, item, scenery, ground_item, npc, or player", args[1].Kind())
	}
}

// ---------- LLM stdlib (brain.Strategist) ----------

// dslContemplateReality dispatches `contemplate_reality(question)`
// through Host.Strategist. Returns the brain's Choice as a String
// wrapped in CallResult.val on success; brain errors become
// CallResult.err with code SERVER_REJECTED (the strategist is
// conceptually a remote service even for the stub).
func dslContemplateReality(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	question := ""
	if len(args) > 0 {
		question = stringOf(args[0])
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "contemplate_reality: no strategist wired"), nil
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("contemplate_reality: %v", err)), nil
	}
	return interp.Ok(interp.String(decision.Choice)), nil
}

// dslEvaluate routes `evaluate(situation)` → strategist with
// Options=[] (free-form). Returns Confidence as Float on .val.
// The 0-1 numeric assessment in the spec maps to the strategist's
// confidence score.
func dslEvaluate(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("evaluate takes 1 argument (situation), got %d", len(args))
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "evaluate: no strategist wired"), nil
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: stringOf(args[0])})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("evaluate: %v", err)), nil
	}
	return interp.Ok(interp.Float(decision.Confidence)), nil
}

// dslDecide routes `decide(options, context?)` → strategist with
// Options bound from the list arg. Returns Choice as String on .val.
// Forwards the optional context string into Situation.Question so
// the strategist sees it.
func dslDecide(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("decide takes 1 or 2 args (options, context?), got %d", len(args))
	}
	options := []string{}
	if list, ok := args[0].(*interp.List); ok {
		for _, item := range list.Items {
			options = append(options, stringOf(item))
		}
	} else {
		return nil, errf("decide: first arg must be a list of options")
	}
	question := ""
	if len(args) == 2 {
		question = stringOf(args[1])
	}
	// Pearl fast path: the host's compiled policy may answer locally (no LLM)
	// or, on a miss, hand back a persona-biased option ordering for the LLM.
	if h.Pearl != nil {
		f := h.pearlFacts(pearl.EventCtx{Action: "decide", Question: question})
		if d, biased, hit := h.Pearl.TryDecide(f, options); hit {
			return interp.Ok(interp.String(d.Choice)), nil
		} else if len(biased) > 0 {
			options = biased
		}
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "decide: no strategist wired"), nil
	}
	// Decision cache (#16): a repeated pearl-MISS decision in materially-the-same
	// state reuses the prior verdict, skipping the (Haiku) Strategist call. Pearl
	// hits above are never cached — they are already free + authoritative.
	key := h.decisionCacheKey(question, options)
	if h.decisionCache != nil {
		if cached, ok := h.decisionCache.Get(key); ok {
			if choice, ok := cached.(string); ok {
				return interp.Ok(interp.String(choice)), nil
			}
		}
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question, Options: options})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("decide: %v", err)), nil
	}
	if h.decisionCache != nil {
		h.decisionCache.Set(key, decision.Choice, decisionCacheTTL)
	}
	return interp.Ok(interp.String(decision.Choice)), nil
}

// ---------- Memory stdlib (cognition.Client) ----------

// dslRecall routes `recall(query, top?)` to the host's knowledge
// surfaces and returns a List<String> on .val.
//
// Priority order:
//
//  1. If h.Corpus is wired (Phase 2.6+), query it directly. Returns
//     formatted chunk strings: "[source § page § section] text".
//     This is the path real wiki/AutoRune content flows through.
//  2. Otherwise, fall back to h.Retriever (Phase 2.5 stub behavior)
//     and return its Bundle.Reflections list.
//
// Routines do not see which path was used — both return List<String>.
// The Corpus path is preferred because chunks carry provenance.
func dslRecall(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("recall takes 1 or 2 args (query, top?), got %d", len(args))
	}
	query := stringOf(args[0])
	maxItems := 3
	if len(args) == 2 {
		if i, ok := args[1].(interp.Int); ok {
			maxItems = int(i)
		}
	}
	if h.Corpus != nil {
		chunks, err := h.Corpus.Recall(ctx, query, maxItems)
		if err != nil {
			return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("recall: %v", err)), nil
		}
		items := make([]interp.Value, 0, len(chunks))
		for _, c := range chunks {
			items = append(items, interp.String(formatChunkForRecall(c)))
		}
		return interp.Ok(&interp.List{Items: items}), nil
	}
	if h.Retriever == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "recall: no retriever or corpus wired"), nil
	}
	hostName := ""
	if h.opts.Username != "" {
		hostName = h.opts.Username
	}
	bundle, err := h.Retriever.Retrieve(ctx, cognition.Retrieval{
		Goal:     query,
		HostName: hostName,
		MaxItems: maxItems,
	})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("recall: %v", err)), nil
	}
	items := make([]interp.Value, 0, len(bundle.Reflections))
	for _, r := range bundle.Reflections {
		items = append(items, interp.String(r))
	}
	return interp.Ok(&interp.List{Items: items}), nil
}

// dslRelationWith routes `relation_with(name)` → retriever with
// goal = "relation with NAME". Returns a string describing the
// relationship from bundle.Persona["relation:NAME"] if present,
// else from the first reflection.
func dslRelationWith(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("relation_with takes 1 argument (name), got %d", len(args))
	}
	name := stringOf(args[0])
	// Prefer the host's own trust ledger (System-1, learned from interactions):
	// if we've met this party, return our felt trust grade.
	if h.ledger != nil && h.ledger.Known(name) {
		return interp.Ok(interp.String(h.ledger.Rel(name).Grade.String())), nil
	}
	if h.Retriever == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "relation_with: no retriever wired"), nil
	}
	bundle, err := h.Retriever.Retrieve(ctx, cognition.Retrieval{
		Goal:     "relation with " + name,
		HostName: h.opts.Username,
		MaxItems: 1,
	})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("relation_with: %v", err)), nil
	}
	rel := ""
	if v, ok := bundle.Persona["relation:"+name]; ok {
		rel = v
	} else if len(bundle.Reflections) > 0 {
		rel = bundle.Reflections[0]
	} else {
		rel = "unknown"
	}
	return interp.Ok(interp.String(rel)), nil
}
