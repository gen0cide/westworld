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
// the DSL can dispatch by name. Each action follows the same
// pattern:
//
//   1. Convert args from interp.Value to Go types.
//   2. Call the Host method with the bound context.
//   3. Map the returned error to a result Value.
//
// Result convention: actions return interp.Null on success, or an
// interp.String error code on a routine-recoverable failure (e.g.
// "out_of_range"). Unexpected errors propagate as Go errors.

// actionCallable is the standard shape for an action wrapper. Bound
// to a Host and a function that takes the resolved positional / named
// args and returns a result.
type actionCallable struct {
	name string
	host *Host
	ctx  context.Context
	fn   func(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error)
}

func (a *actionCallable) Kind() string    { return "action" }
func (a *actionCallable) Display() string { return "<action " + a.name + ">" }
func (a *actionCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.fn(ctx, a.host, args, named)
}

// errf is a small helper for returning formatted errors from
// Callable.Call().
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// ---------- movement ----------

func dslWalkTo(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, err
	}
	if err := h.WalkTo(ctx, x, y); err != nil {
		return interp.String("walk_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
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
			return interp.String("attack_failed: " + err.Error()), nil
		}
	case *playerView:
		if err := h.AttackPlayer(ctx, v.record.Index); err != nil {
			return interp.String("attack_failed: " + err.Error()), nil
		}
	default:
		// Fall back: if it's an int, treat as a server index (NPC).
		if i, ok := interp.AsInt(target); ok {
			if err := h.AttackNpc(ctx, int(i)); err != nil {
				return interp.String("attack_failed: " + err.Error()), nil
			}
			return interp.Null{}, nil
		}
		return nil, errf("attack: target must be npc, player, or int, got %s", target.Kind())
	}
	return interp.Null{}, nil
}

func dslTalkTo(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("talk_to takes 1 argument (npc), got %d", len(args))
	}
	if n, ok := args[0].(*npcView); ok {
		if err := h.TalkToNpc(ctx, n.record.Index); err != nil {
			return interp.String("talk_failed: " + err.Error()), nil
		}
		return interp.Null{}, nil
	}
	if i, ok := interp.AsInt(args[0]); ok {
		if err := h.TalkToNpc(ctx, int(i)); err != nil {
			return interp.String("talk_failed: " + err.Error()), nil
		}
		return interp.Null{}, nil
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
		return interp.String("answer_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

// ---------- items ----------

func dslDrop(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	// drop(item) or drop(slot=N) — we accept either.
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.DropItem(ctx, slot); err != nil {
		return interp.String("drop_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
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
		return interp.String("pick_up_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

func dslEat(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	// eat(item) or eat(slot=N). Routes through ItemCommand which
	// handles Eat/Drink/Bury depending on the item def.
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.ItemCommand(ctx, slot); err != nil {
		return interp.String("eat_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

// ---------- banking ----------

func dslDeposit(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankDeposit(ctx, itemID, amount); err != nil {
		return interp.String("deposit_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

func dslWithdraw(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	itemID, amount, err := resolveItemAmount(h, args, named)
	if err != nil {
		return nil, err
	}
	if err := h.BankWithdraw(ctx, itemID, amount); err != nil {
		return interp.String("withdraw_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

func dslCloseBank(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.BankClose(ctx); err != nil {
		return interp.String("close_bank_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
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
		return interp.String("say_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
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
		return interp.String("whisper_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

func dslLogout(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.Logout(ctx); err != nil {
		return interp.String("logout_failed: " + err.Error()), nil
	}
	return interp.Null{}, nil
}

// ---------- primitives ----------

// dslWait sleeps for the given duration (seconds). The interpreter
// passes a single Float arg (the resolved duration, with range jitter
// already applied).
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

// dslNote writes to the host's logger as a debug entry. The full
// "journal" persistence is step 8.
func dslNote(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("note takes 1 argument (text), got %d", len(args))
	}
	h.log.Info("routine note", "text", args[0].Display())
	return interp.Null{}, nil
}

// ---------- stubs for skills not yet implemented ----------

func makeStub(name string) func(context.Context, *Host, []interp.Value, map[string]interp.Value) (interp.Value, error) {
	return func(_ context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		h.log.Warn("dsl action not yet implemented", "action", name)
		return interp.String(name + ": not_implemented"), nil
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

// resolveSlot accepts: drop(item) where item carries .amount/.id, or
// drop(slot=N) as an explicit slot index.
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
