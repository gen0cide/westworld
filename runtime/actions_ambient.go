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
	if n, ok := args[0].(*npcView); ok {
		if err := h.TalkToNpc(ctx, n.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	if i, ok := interp.AsInt(args[0]); ok {
		if err := h.TalkToNpc(ctx, int(i)); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	return nil, errf("talk_to: target must be npc or int, got %s", args[0].Kind())
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
	h.log.Info("routine note", "text", args[0].Display())
	return interp.Null{}, nil
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
//   is_reachable(x, y)
//   is_reachable(position)
//   is_reachable(view)  — any view with .x/.y
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
//   interact_at(x=X, y=Y)
//   interact_at(x=X, y=Y, option=2)
//   interact_at(position)              — any view with .x/.y
//   interact_at(scenery_view)          — placement from world.locs
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
func dslUse(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
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
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "decide: no strategist wired"), nil
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question, Options: options})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("decide: %v", err)), nil
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
