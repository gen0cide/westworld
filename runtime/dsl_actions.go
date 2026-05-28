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
	"github.com/gen0cide/westworld/world"
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
	"command":     dslCommand,
	"use":         dslUse,
	"interact_at":  dslInteractAt,
	"distance_to":  dslDistanceTo,
	"in_region":    dslInRegion,
	"box":          dslBox,
	"circle":       dslCircle,
	"near":         dslNear,
	"activate_prayer":   dslActivatePrayer,
	"deactivate_prayer": dslDeactivatePrayer,
	"equip":             dslEquip,
	"unequip":           dslUnequip,
	"cast_on_self":      dslCastOnSelf,
	"cast_on_npc":       dslCastOnNpc,
	"cast_on_player":    dslCastOnPlayer,
	"cast_on_land":      dslCastOnLand,
	"cast_on_item":      dslCastOnInventory,
	"walk_path":       dslWalkPath,
	"is_reachable":    dslIsReachable,
	"wait_for_dialog": dslWaitForDialog,

	// Trade
	"trade_request":      dslTradeRequest,
	"accept_trade":       dslAcceptTrade,
	"decline_trade":      dslDeclineTrade,
	"offer_trade":        dslOfferTrade,
	"confirm_trade":      dslConfirmTrade,
	"duel_request":       dslDuelRequest,
	"accept_duel":        dslAcceptDuel,
	"decline_duel":       dslDeclineDuel,
	"offer_duel":         dslOfferDuel,
	"set_duel_rules":     dslSetDuelRules,
	"accept_duel_offer":  dslAcceptDuelOffer,
	"accept_duel_confirm": dslAcceptDuelConfirm,
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
		h.lastAttackedNpcIndex = v.record.Index
		if err := h.AttackNpc(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	case *playerView:
		h.lastAttackedPlayerIndex = v.record.Index
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

// ---------- trade — request / accept / offer / confirm / decline ----------

// dslTradeRequest sends a trade request to a player. Accepts a
// player-view or a server-index Int. Walks adjacent first
// (server requires adjacency).
func dslTradeRequest(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("trade_request takes 1 arg (player), got %d", len(args))
	}
	var idx int
	switch v := args[0].(type) {
	case *playerView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(args[0]); ok {
			idx = int(i)
		} else {
			return nil, errf("trade_request: target must be a player view or Int index, got %s", args[0].Kind())
		}
	}
	if err := h.InitTradeRequest(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAcceptTrade accepts an incoming trade request by re-sending the
// trade-request packet to the original requester (OpenRSC's symmetric
// handshake — both sides must request each other for the window to
// open). Takes a player view, server-index Int, or player-name String
// (string lookups resolve via world.Players).
func dslAcceptTrade(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("accept_trade takes 1 arg (requester player/name/index), got %d", len(args))
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	if err := h.AcceptIncomingTrade(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// resolvePlayerIndex turns a DSL value into a server-side player
// index. Accepts:
//   - *playerView (from world.players.find / by_index)
//   - interp.String (looked up via world.Players.FindByName)
//   - interp.Int (raw index, used as-is)
func resolvePlayerIndex(h *Host, v interp.Value) (int, error) {
	switch x := v.(type) {
	case *playerView:
		return x.record.Index, nil
	case interp.String:
		rec, ok := h.world.Players.FindByName(string(x))
		if !ok {
			return 0, errf("player %q not visible", string(x))
		}
		return rec.Index, nil
	default:
		if i, ok := interp.AsInt(v); ok {
			return int(i), nil
		}
	}
	return 0, errf("expected player view, name string, or Int index; got %s", v.Kind())
}

func dslDeclineTrade(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.DeclineTrade(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOfferTrade takes a list of [item_id, amount] pairs (Int, Int)
// and sends a trade offer. Replaces any prior offer.
func dslOfferTrade(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("offer_trade takes 1 arg (list of [item_id, amount]), got %d", len(args))
	}
	list, ok := args[0].(*interp.List)
	if !ok {
		return nil, errf("offer_trade: arg must be a list, got %s", args[0].Kind())
	}
	items := make([]world.TradeItem, 0, len(list.Items))
	for i, el := range list.Items {
		pair, ok := el.(*interp.List)
		if !ok || len(pair.Items) != 2 {
			return nil, errf("offer_trade: element %d must be [item_id, amount], got %s", i, el.Kind())
		}
		id, idok := pair.Items[0].(interp.Int)
		amt, amtok := pair.Items[1].(interp.Int)
		if !idok || !amtok {
			return nil, errf("offer_trade: element %d fields must be Int", i)
		}
		items = append(items, world.TradeItem{ItemID: int(id), Amount: int(amt)})
	}
	if err := h.OfferTradeItems(ctx, items); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslConfirmTrade(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.ConfirmTrade(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- duel builtins (parallel structure to trade) ----------

// dslDuelRequest initiates a duel with a player or server-index.
// Walks adjacent first (server requires adjacency).
func dslDuelRequest(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("duel_request takes 1 arg (player), got %d", len(args))
	}
	var idx int
	switch v := args[0].(type) {
	case *playerView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(args[0]); ok {
			idx = int(i)
		} else {
			return nil, errf("duel_request: target must be a player view or Int index, got %s", args[0].Kind())
		}
	}
	if err := h.InitDuelRequest(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAcceptDuel accepts an incoming duel request by re-sending the
// duel-request packet to the original requester (symmetric handshake
// like trade). Takes a player view, server-index Int, or player-name
// String. The duel_request event delivers the requester's name, so
// the typical call is `accept_duel(name)`.
func dslAcceptDuel(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("accept_duel takes 1 arg (requester player/name/index), got %d", len(args))
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	if err := h.AcceptIncomingDuel(ctx, idx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslDeclineDuel(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.DeclineDuel(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOfferDuel takes a list of [item_id, amount] pairs and stakes
// them. Same shape as offer_trade.
func dslOfferDuel(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("offer_duel takes 1 arg (list of [item_id, amount]), got %d", len(args))
	}
	list, ok := args[0].(*interp.List)
	if !ok {
		return nil, errf("offer_duel: arg must be a list, got %s", args[0].Kind())
	}
	items := make([]world.TradeItem, 0, len(list.Items))
	for i, el := range list.Items {
		pair, ok := el.(*interp.List)
		if !ok || len(pair.Items) != 2 {
			return nil, errf("offer_duel: element %d must be [item_id, amount], got %s", i, el.Kind())
		}
		id, idok := pair.Items[0].(interp.Int)
		amt, amtok := pair.Items[1].(interp.Int)
		if !idok || !amtok {
			return nil, errf("offer_duel: element %d fields must be Int", i)
		}
		items = append(items, world.TradeItem{ItemID: int(id), Amount: int(amt)})
	}
	if err := h.OfferDuelItems(ctx, items); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslSetDuelRules takes a 4-element list of bools or kwargs and
// sends the rule toggles. Accepts either:
//   set_duel_rules([true, false, false, true])    # retreat, magic, prayer, weapons
//   set_duel_rules(retreat=true, weapons=true)
func dslSetDuelRules(ctx context.Context, h *Host, args []interp.Value, kwargs map[string]interp.Value) (interp.Value, error) {
	var r world.DuelRules
	if len(args) == 1 {
		list, ok := args[0].(*interp.List)
		if !ok || len(list.Items) != 4 {
			return nil, errf("set_duel_rules: positional arg must be a 4-element list [retreat, magic, prayer, weapons]")
		}
		flags := make([]bool, 4)
		for i, el := range list.Items {
			b, bok := el.(interp.Bool)
			if !bok {
				return nil, errf("set_duel_rules: list element %d must be Bool", i)
			}
			flags[i] = bool(b)
		}
		r = world.DuelRules{
			DisallowRetreat: flags[0],
			DisallowMagic:   flags[1],
			DisallowPrayer:  flags[2],
			DisallowWeapons: flags[3],
		}
	} else if len(args) == 0 {
		// kwargs mode
		readFlag := func(name string) bool {
			v, ok := kwargs[name]
			if !ok {
				return false
			}
			b, _ := v.(interp.Bool)
			return bool(b)
		}
		r = world.DuelRules{
			DisallowRetreat: readFlag("retreat"),
			DisallowMagic:   readFlag("magic"),
			DisallowPrayer:  readFlag("prayer"),
			DisallowWeapons: readFlag("weapons"),
		}
	} else {
		return nil, errf("set_duel_rules takes either a 4-element list or named kwargs (retreat/magic/prayer/weapons)")
	}
	if err := h.SetDuelRules(ctx, r); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslAcceptDuelOffer(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.AcceptDuelOffer(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslAcceptDuelConfirm(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.AcceptDuelConfirm(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
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

// intArg coerces an interp.Value to int. Used by simple builtins
// that expect Int params — returns 0 for non-Int values, callers
// validate args before reaching this.
func intArg(v interp.Value) int {
	if i, ok := v.(interp.Int); ok {
		return int(i)
	}
	return 0
}

// ---------- equip / unequip ----------

// dslEquip moves an inventory item into its worn slot.
// equip(item_view) or equip(slot=N).
func dslEquip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.EquipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslUnequip returns a wielded item to the inventory.
func dslUnequip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.UnequipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- magic cast ----------

func dslCastOnSelf(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("cast_on_self takes 1 arg (spell_id)")
	}
	sp, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("cast_on_self: spell_id must be Int")
	}
	if err := h.CastOnSelf(ctx, int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnNpc(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("cast_on_npc takes 2 args (npc, spell_id)")
	}
	var idx int
	switch v := args[0].(type) {
	case *npcView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(args[0]); ok {
			idx = int(i)
		} else {
			return nil, errf("cast_on_npc: npc arg must be npc view or Int index")
		}
	}
	sp, ok := interp.AsInt(args[1])
	if !ok {
		return nil, errf("cast_on_npc: spell_id must be Int")
	}
	if err := h.CastOnNpc(ctx, idx, int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnPlayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("cast_on_player takes 2 args (player, spell_id)")
	}
	idx, err := resolvePlayerIndex(h, args[0])
	if err != nil {
		return nil, err
	}
	sp, ok := interp.AsInt(args[1])
	if !ok {
		return nil, errf("cast_on_player: spell_id must be Int")
	}
	if err := h.CastOnPlayer(ctx, idx, int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnLand(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 3 {
		return nil, errf("cast_on_land takes 3 args (x, y, spell_id)")
	}
	x, _ := interp.AsInt(args[0])
	y, _ := interp.AsInt(args[1])
	sp, _ := interp.AsInt(args[2])
	if err := h.CastOnLand(ctx, int(x), int(y), int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnInventory(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("cast_on_item takes (item, spell_id)")
	}
	slot, err := resolveSlot(h, args[:1], named)
	if err != nil {
		return nil, err
	}
	var sp int64
	if len(args) == 2 {
		if i, ok := interp.AsInt(args[1]); ok {
			sp = i
		}
	}
	if v, ok := named["spell_id"]; ok {
		if i, okk := interp.AsInt(v); okk {
			sp = i
		}
	}
	if err := h.CastOnInventory(ctx, slot, int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- prayer activate/deactivate ----------

// dslActivatePrayer turns on a prayer slot.
// activate_prayer(N) where N is 0..13. Server may silently reject
// (low prayer level or zero prayer points) — routines should check
// world.prayer.active(N) after.
func dslActivatePrayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("activate_prayer takes 1 arg (prayer index)")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("activate_prayer: index must be Int")
	}
	if err := h.ActivatePrayer(ctx, int(id)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslDeactivatePrayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("deactivate_prayer takes 1 arg (prayer index)")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("deactivate_prayer: index must be Int")
	}
	if err := h.DeactivatePrayer(ctx, int(id)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
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
