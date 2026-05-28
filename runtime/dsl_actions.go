package runtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// This file wraps each Host action method as an interp.Callable so
// the DSL can dispatch by name. Each action returns a typed
// *interp.CallResult — `interp.Ok(value)` on success,
// `interp.Fail(code, reason)` on failure with a typed ErrorCode.
//
// See docs/lang/actions.md for the error taxonomy and the bang
// convention. The bridge auto-registers `<name>!` BangCallables
// for every Result-returning action; see dsl_bridge.go.

// actionHandler is the Go signature every action wrapper conforms to.
type actionHandler func(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error)

// actionHandlers maps every spec.Actions entry's name to its Go
// implementation. The bridge iterates spec.Actions and looks up the
// handler here. Adding a builtin = add a row to spec.Actions AND
// add a handler entry here; the consistency test in dsl/spec/
// catches any mismatch.
//
// For entries marked NotYetImplemented in spec.Actions, the bridge
// uses a NOT_IMPLEMENTED stub instead of looking up here.
var actionHandlers = map[string]actionHandler{
	// Primary actions
	"walk_to":    dslWalkTo,
	"attack":     dslAttack,
	"talk_to":    dslTalkTo,
	"answer":     dslAnswer,
	"drop":       dslDrop,
	"pick_up":    dslPickUp,
	"eat":        dslEat,
	"open_bank":  dslOpenBank,
	"deposit":    dslDeposit,
	"withdraw":   dslWithdraw,
	"close_bank": dslCloseBank,
	"say":        dslSay,
	"whisper":    dslWhisper,
	"command":    dslCommand,
	"use":        dslUse,
	"logout":     dslLogout,

	// Primitives
	"wait": dslWait,
	"note": dslNote,

	// LLM stdlib — routed through Host.Strategist (brain.Strategist).
	// Stub strategist returns deterministic canned decisions; the
	// Phase 4 LLM impl drops in by swapping Host.Strategist.
	"contemplate_reality": dslContemplateReality,
	"evaluate":            dslEvaluate,
	"decide":              dslDecide,

	// Memory stdlib — routed through Host.Retriever (cognition.Client).
	// Stub retriever returns canned bundles; the Phase 3 mesa impl
	// drops in by swapping Host.Retriever.
	"recall":        dslRecall,
	"relation_with": dslRelationWith,
}

// actionCallable is the standard shape for an action wrapper. Bound
// to a Host and a function that takes the resolved positional / named
// args and returns a *CallResult-shaped Value.
type actionCallable struct {
	name string
	host *Host
	ctx  context.Context
	fn   func(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error)
}

func (a *actionCallable) Kind() string    { return "action" }
func (a *actionCallable) Display() string { return "<action " + a.name + ">" }

// Yields implements interp.Yielder — primary actions yield control
// to the interpreter's event dispatcher before they run, so any
// `on` handler can interleave between actions.
func (a *actionCallable) Yields() bool { return true }

func (a *actionCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.fn(ctx, a.host, args, named)
}

// errf is a small helper for returning formatted Go-level errors from
// Callable.Call. These surface as RuntimeError (routine-ends-
// abnormally) — for typed routine-recoverable failures use
// interp.Fail directly.
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// wrapServerErr maps a Go error returned from a Host method into a
// typed *interp.CallResult failure value so the DSL routine can
// branch on `.err.code`. Classification is by message-substring
// matching for now; in Phase 2.6+ Host methods will return typed
// errors directly and this helper can shrink.
func wrapServerErr(err error) interp.Value {
	// Typed sentinels first — these carry richer info (e.g. door
	// coords, server prose) than the string-match classifier
	// below can recover. Add cases here as Host methods migrate
	// from formatted-string errors to typed ones.
	if doorErr, ok := err.(*DoorLockedError); ok {
		return interp.Fail(interp.DOOR_LOCKED, doorErr.Error())
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	var code interp.ErrorCode
	switch {
	case strings.Contains(lower, "out of range"),
		strings.Contains(lower, "outofrange"),
		strings.Contains(lower, "too far"):
		code = interp.OUT_OF_RANGE
	case strings.Contains(lower, "stalled"),
		strings.Contains(lower, "path"):
		code = interp.PATH_BLOCKED
	case strings.Contains(lower, "inventory full"),
		strings.Contains(lower, "inv full"):
		code = interp.INVENTORY_FULL
	case strings.Contains(lower, "canceled"),
		strings.Contains(lower, "deadline exceeded"):
		code = interp.INTERRUPTED
	case strings.Contains(lower, "timeout"):
		code = interp.ACTION_TIMEOUT
	default:
		code = interp.SERVER_REJECTED
	}
	return interp.Fail(code, msg)
}

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

// ---------- combat ----------

func dslAttack(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("attack takes 1 argument (target), got %d", len(args))
	}
	target := args[0]
	// Two valid shapes: an NPC view (.index) or a player view (.index).
	switch v := target.(type) {
	case *npcView:
		if err := h.AttackNpc(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	case *playerView:
		if err := h.AttackPlayer(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	default:
		// Fall back: if it's an int, treat as a server index (NPC).
		if i, ok := interp.AsInt(target); ok {
			if err := h.AttackNpc(ctx, int(i)); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		}
		return nil, errf("attack: target must be npc, player, or int, got %s", target.Kind())
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

func dslAnswer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("answer takes 1 argument (option index), got %d", len(args))
	}
	idx, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("answer: option index must be int, got %s", args[0].Kind())
	}
	if err := h.ChooseDialogOption(ctx, int(idx)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- items ----------

func dslDrop(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		// "item N not in inventory" is routine-recoverable (NO_SUCH_ITEM);
		// other resolution failures are programmer bugs (return Go err).
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.DropItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslPickUp(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("pick_up takes 1 argument (ground_item), got %d", len(args))
	}
	gi, ok := args[0].(*groundItemView)
	if !ok {
		return nil, errf("pick_up: expected ground_item, got %s", args[0].Kind())
	}
	if err := h.PickUpItem(ctx, gi.record.X, gi.record.Y, gi.record.ItemID); err != nil {
		return wrapServerErr(err), nil
	}
	// On success, return the picked-up item-view so DSL can do:
	//     got = pick_up!(item) ; say(got.name)
	picked := newItemView(h.facts, gi.record.ItemID, 1)
	return interp.Ok(picked), nil
}

func dslEat(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.ItemCommand(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- banking ----------

func dslDeposit(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankDeposit(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslWithdraw(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankWithdraw(ctx, itemID, amount); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCloseBank(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.BankClose(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslOpenBank(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	// open_bank(banker) = talk_to(banker); the DSL author then uses
	// answer(N) to pick the "I'd like to access my bank" option.
	// We delegate to talk_to since RSC has no dedicated open-bank
	// packet — the bank UI is the bottom of a dialog tree.
	return dslTalkTo(ctx, h, args, nil)
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

func dslLogout(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.Logout(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
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
	if len(args) != 2 {
		return nil, errf("use takes 2 arguments (item, target), got %d", len(args))
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
	default:
		return nil, errf("use: unsupported target type %s — expected boundary, item (scenery + ground_item + npc/player coming)", args[1].Kind())
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

// dslRecall routes `recall(query, top?)` → retriever's Retrieve
// with the query as Goal. Returns the bundle's Reflections list
// as a List<String> on .val, trimmed to `top` if specified.
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
	if h.Retriever == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "recall: no retriever wired"), nil
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

// stringOf coerces any DSL Value to its Display() string. Used
// for the LLM/memory routing where we need free-text passes —
// not a substitute for type coercion in user code.
func stringOf(v interp.Value) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(interp.String); ok {
		return string(s)
	}
	return v.Display()
}

// ---------- stubs for actions not yet implemented ----------

func makeStub(name string) func(context.Context, *Host, []interp.Value, map[string]interp.Value) (interp.Value, error) {
	return func(_ context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		h.log.Warn("dsl action not yet implemented", "action", name)
		return interp.Fail(interp.NOT_IMPLEMENTED, name), nil
	}
}

// ---------- argument resolvers ----------

// resolvePoint extracts (x, y) from args. Accepts:
//   - positional: walk_to(x, y) — two ints
//   - named: walk_to(x=..., y=...)
//   - single value with .x/.y (e.g. self.position)
func resolvePoint(args []interp.Value, named map[string]interp.Value) (int, int, error) {
	if vx, ok := named["x"]; ok {
		if vy, ok := named["y"]; ok {
			x, xok := interp.AsInt(vx)
			y, yok := interp.AsInt(vy)
			if !xok || !yok {
				return 0, 0, errf("x and y must be ints")
			}
			return int(x), int(y), nil
		}
	}
	if len(args) == 2 {
		x, xok := interp.AsInt(args[0])
		y, yok := interp.AsInt(args[1])
		if xok && yok {
			return int(x), int(y), nil
		}
	}
	if len(args) == 1 {
		if g, ok := args[0].(interp.Getter); ok {
			xv, hasX := g.Get("x")
			yv, hasY := g.Get("y")
			if hasX && hasY {
				x, xok := interp.AsInt(xv)
				y, yok := interp.AsInt(yv)
				if xok && yok {
					return int(x), int(y), nil
				}
			}
		}
	}
	return 0, 0, errf("could not resolve (x, y) from arguments")
}

// resolveSlot accepts: drop(item) where item carries .id, or
// drop(slot=N) as an explicit slot index. Returns "item N not in
// inventory" error when the item isn't found — callers should
// convert that into a NO_SUCH_ITEM CallResult.
func resolveSlot(h *Host, args []interp.Value, named map[string]interp.Value) (int, error) {
	if v, ok := named["slot"]; ok {
		if i, ok := interp.AsInt(v); ok {
			return int(i), nil
		}
	}
	if len(args) == 1 {
		if i, ok := interp.AsInt(args[0]); ok {
			return int(i), nil
		}
		if g, ok := args[0].(interp.Getter); ok {
			if v, ok := g.Get("id"); ok {
				if id, ok := interp.AsInt(v); ok {
					// Find first slot with that item ID.
					for i, s := range h.world.Inventory.Slots() {
						if s.ItemID == int(id) {
							return i, nil
						}
					}
					return 0, errf("item %d not in inventory", id)
				}
			}
		}
	}
	return 0, errf("could not resolve inventory slot")
}

// resolveItemAmount extracts (item ID, amount) for deposit/withdraw.
// Accepts: deposit(item, count) or deposit(item=ID, count=N).
func resolveItemAmount(h *Host, args []interp.Value, named map[string]interp.Value) (int, int, error) {
	var itemVal interp.Value
	var countVal interp.Value
	if v, ok := named["item"]; ok {
		itemVal = v
	} else if len(args) >= 1 {
		itemVal = args[0]
	}
	if v, ok := named["count"]; ok {
		countVal = v
	} else if v, ok := named["amount"]; ok {
		countVal = v
	} else if len(args) >= 2 {
		countVal = args[1]
	}
	if itemVal == nil || countVal == nil {
		return 0, 0, errf("deposit/withdraw need item + count")
	}
	id, err := resolveItemID(h.facts, itemVal)
	if err != nil {
		return 0, 0, err
	}
	count, ok := interp.AsInt(countVal)
	if !ok {
		return 0, 0, errf("count must be int")
	}
	return id, int(count), nil
}

// resolveItemID accepts either a literal item ID (Int), an item view
// (Getter exposing .id), or a string item name (looked up via facts).
func resolveItemID(f *facts.Facts, v interp.Value) (int, error) {
	if i, ok := interp.AsInt(v); ok {
		return int(i), nil
	}
	if g, ok := v.(interp.Getter); ok {
		if iv, ok := g.Get("id"); ok {
			if i, ok := interp.AsInt(iv); ok {
				return int(i), nil
			}
		}
	}
	if s, ok := interp.AsString(v); ok {
		if f == nil {
			return 0, errf("cannot resolve item name %q without facts", s)
		}
		// Linear scan; facts.ItemDefs is small (<1000 entries).
		needle := strings.ToLower(s)
		for id, def := range f.ItemDefs {
			if def != nil && strings.EqualFold(def.Name, s) {
				return id, nil
			}
		}
		// Substring fallback.
		for id, def := range f.ItemDefs {
			if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
				return id, nil
			}
		}
		return 0, errf("item name %q not found in facts", s)
	}
	return 0, errf("cannot resolve item ID from %s", v.Kind())
}

// ---------- tiny helpers ----------

// itemName looks up the friendly name of an item ID from facts, or
// falls back to "item#N" if facts isn't loaded.
func itemName(f *facts.Facts, id int) string {
	if f != nil {
		if def := f.ItemDef(id); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return "item#" + intDisp(id)
}

func intDisp(i int) string { return strconv.Itoa(i) }

func chebyshev(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// _ = action.MaxClickRange — silence import (used elsewhere in this
// package; keep the package referenced so the import line is real).
var _ = action.MaxClickRange
