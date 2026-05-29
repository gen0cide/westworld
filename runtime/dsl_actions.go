package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
)

// dsl_actions.go is the ACTION REGISTRY HUB. It holds:
//
//   - the actionHandler signature every wrapper conforms to,
//   - the central actionHandlers table — the single name->handler
//     map for every action builtin, grouped by namespace so it reads
//     like a table of contents mirroring docs/lang/api.md,
//   - the actionCallable dispatch shape the bridge binds per-Host,
//   - the shared error/result plumbing (errf, wrapServerErr, makeStub).
//
// The handler FUNCTION BODIES live in the per-namespace
// actions_*.go files (actions_trade.go, actions_bank.go,
// actions_combat.go, actions_magic.go, actions_prayer.go,
// actions_duel.go, actions_inventory.go, actions_ambient.go). Shared
// argument resolvers + tiny helpers live in dsl_helpers.go. View
// structs + view-root wiring live in views_*.go and dsl_bridge.go.
//
// Each action returns a typed *interp.CallResult — `interp.Ok(value)`
// on success, `interp.Fail(code, reason)` on failure with a typed
// ErrorCode. See docs/lang/actions.md for the error taxonomy and the
// bang convention; the bridge auto-registers `<name>!` BangCallables
// for every Result-returning action (see dsl_bridge.go).

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
//
// The grouping below mirrors the namespace order in
// docs/lang/api.md so the entire action surface is visible at a
// glance. Each handler body lives in the matching actions_*.go file.
var actionHandlers = map[string]actionHandler{
	// ---- ambient: movement & navigation (actions_ambient.go) ----
	"walk_to":       dslWalkTo,
	"walk_path":     dslWalkPath,
	"is_reachable":  dslIsReachable,
	"follow":        dslFollow,
	"open_boundary": dslOpenBoundary,

	// ---- ambient: NPC / player interaction (actions_ambient.go) ----
	// pickpocket is the canonical NPC-command verb (§10 drops
	// npc_command as a second name).
	"talk_to":               dslTalkTo,
	"pickpocket":            dslNpcCommand,
	"answer":                dslAnswer,
	"interact_at":           dslInteractAt,
	"use":                   dslUse,
	"use_inventory_default": dslUseInventoryDefault,

	// ---- inventory / items (actions_inventory.go) ----
	"drop":    dslDrop,
	"pick_up": dslPickUp,
	"eat":     dslEat,
	"equip":   dslEquip,
	"unequip": dslUnequip,

	// ---- combat (actions_combat.go) ----
	// `attack` is the sanctioned §9 alias for combat.attack; the
	// namespaced combat.attack / combat.set_style verbs dispatch
	// through combatView (see views_combat.go + combatVerbs).
	"attack": dslAttack,

	// ---- magic (actions_magic.go) ----
	// `cast` is the sanctioned §9 alias for magic.cast (polymorphic);
	// the namespaced magic.cast dispatches through magicView.
	"cast": dslMagicCast,

	// The trade.* / duel.* / bank.* / magic.* / prayer.* / combat.*
	// verbs are namespaced view-dispatched callables, NOT bare
	// builtins — see views_*.go + the per-namespace verb tables in
	// this file. Their handler bodies live in actions_*.go.

	// ---- ambient: social & chat (actions_ambient.go) ----
	"say":         dslSay,
	"whisper":     dslWhisper,
	"add_friend":  dslAddFriend,
	"find_option": dslFindOption,

	// ---- ambient: session & admin (actions_ambient.go) ----
	"command": dslCommand,
	"logout":  dslLogout,

	// ---- ambient: spatial utilities + bounds constructors (actions_ambient.go) ----
	"distance_to":    dslDistanceTo,
	"distance_to_xy": dslDistanceToXY,
	"in_region":      dslInRegion,
	"box":            dslBox,
	"circle":         dslCircle,
	"near":           dslNear,

	// ---- primitives — flow / timing / introspection (actions_ambient.go) ----
	"wait":            dslWait,
	"wait_until":      dslWaitUntil,
	"wait_for_dialog": dslWaitForDialog,
	"note":            dslNote,

	// ---- control plane: recognition / fuzzy resolution (actions_resolve.go) ----
	// Fenced, non-GUI primitives (api.md §5). Routed through the host's
	// recognition faculty (Host.Resolver: learned-alias → fuzzy →
	// brain). No packets, no bang variant — like note().
	"resolve":     dslResolve,
	"resolve_one": dslResolveOne,

	// ---- cognition bridge: LLM stdlib (actions_ambient.go) ----
	// Routed through Host.Strategist (brain.Strategist). Stub
	// strategist returns deterministic canned decisions; the Phase 4
	// LLM impl drops in by swapping Host.Strategist.
	"contemplate_reality": dslContemplateReality,
	"evaluate":            dslEvaluate,
	"decide":              dslDecide,

	// ---- cognition bridge: memory stdlib (actions_ambient.go) ----
	// Routed through Host.Retriever (cognition.Client). Stub
	// retriever returns canned bundles; the Phase 3 mesa impl drops
	// in by swapping Host.Retriever.
	"recall":        dslRecall,
	"relation_with": dslRelationWith,
}

// ---- namespaced verb tables (view-dispatched; api.md §6/§10) ----
//
// These map a namespace verb (no root, no bang) to its handler body.
// The namespace views (views_trade.go, views_bank.go, …) consult
// these via Host.namespaceAction so the frozen surface reads as
// trade.request(p) / bank.deposit(item, n) / magic.cast(spell, t?)
// etc. The bodies are shared with the §9 flat aliases (attack, cast).

// tradeVerbs — frozen trade.* actions (§10: request absorbs the old
// trade_request+accept_trade; accept<-confirm_trade; confirm<-
// finalize_trade).
var tradeVerbs = map[string]actionHandler{
	"request": dslTradeRequest,
	"offer":   dslOfferTrade,
	"accept":  dslConfirmTrade,  // offer-screen accept (screen 1)
	"confirm": dslFinalizeTrade, // confirm-screen accept (screen 2)
	"decline": dslDeclineTrade,
}

// duelVerbs — frozen duel.* actions (§10: request absorbs duel_request
// +accept_duel; set_rules; stake<-offer_duel; accept<-accept_duel_offer;
// confirm<-accept_duel_confirm).
var duelVerbs = map[string]actionHandler{
	"request":   dslDuelRequest,
	"set_rules": dslSetDuelRules,
	"stake":     dslOfferDuel,
	"accept":    dslAcceptDuelOffer,   // offer-screen accept (screen 1)
	"confirm":   dslAcceptDuelConfirm, // confirm-screen accept (screen 2)
	"decline":   dslDeclineDuel,
}

// bankVerbs — frozen bank.* actions (§10) + bulk verbs (#117/#118).
var bankVerbs = map[string]actionHandler{
	"open":     dslOpenBank,
	"deposit":  dslDeposit,
	"withdraw": dslWithdraw,
	"close":    dslCloseBank,
	// bulk verbs (#117/#118): deposit_all(keep?), withdraw_all(item),
	// withdraw_x(item, amount).
	"deposit_all":  dslDepositAll,
	"withdraw_all": dslWithdrawAll,
	"withdraw_x":   dslWithdrawX,
}

// magicVerbs — frozen magic.* actions (§10: one polymorphic cast).
var magicVerbs = map[string]actionHandler{
	"cast": dslMagicCast,
}

// prayerVerbs — frozen prayer.* actions (§10).
var prayerVerbs = map[string]actionHandler{
	"activate":   dslActivatePrayer,
	"deactivate": dslDeactivatePrayer,
}

// combatVerbs — frozen combat.* actions (§10: attack kept as alias).
var combatVerbs = map[string]actionHandler{
	"attack":    dslAttack,
	"set_style": dslSetCombatStyle,
	"retreat":   dslRetreat, // #117: break melee by walking one tile away
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

// makeStub returns a handler that logs + fails with NOT_IMPLEMENTED.
// Used by the bridge for spec.Actions entries that are declared but
// not yet wired (NotYetImplemented) or have no handler entry above.
func makeStub(name string) func(context.Context, *Host, []interp.Value, map[string]interp.Value) (interp.Value, error) {
	return func(_ context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		h.log.Warn("dsl action not yet implemented", "action", name)
		return interp.Fail(interp.NOT_IMPLEMENTED, name), nil
	}
}
