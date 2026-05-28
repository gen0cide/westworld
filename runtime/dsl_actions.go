package runtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
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
	if err := h.WalkTo(ctx, x, y); err != nil {
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
