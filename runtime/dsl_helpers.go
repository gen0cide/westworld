package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/cognition/corpus"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// This file holds the shared argument resolvers and tiny utility
// helpers used across the per-namespace action handlers. They carry
// no namespace of their own — they are the plumbing every builtin
// reaches for when turning DSL values into the ints / ids / slots the
// Host methods want.
//
// The dispatch core (actionCallable, the actionHandlers registry,
// errf, wrapServerErr, makeStub) lives in dsl_actions.go; this file
// is the resolver toolbox.

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
	// A view/position as args[0] resolves to its (x, y). Accept this even
	// when a trailing positional follows (e.g. interact_at(scenery, 2) —
	// the second arg is the option), so option-2 scenery clicks work
	// with a view argument and not only named x=/y=.
	if len(args) >= 1 {
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

// resolveSpellID accepts either an Int (raw spell ID) or a String
// (canonical spell name like "Wind Strike", "Varrock Teleport") and
// returns the underlying SpellDef ID. The string lookup is
// case-insensitive via facts.SpellByName, which indexes the catalog
// at init.
func resolveSpellID(v interp.Value) (int, error) {
	if i, ok := interp.AsInt(v); ok {
		return int(i), nil
	}
	if s, ok := v.(interp.String); ok {
		def := facts.SpellByName(string(s))
		if def == nil {
			return 0, errf("unknown spell %q (see facts.SpellDef.xml)", string(s))
		}
		return def.ID, nil
	}
	return 0, errf("spell must be Int id or String name, got %s", v.Kind())
}

// resolvePrayerID accepts an Int (raw prayer slot 0..13) or a String
// (canonical prayer name like "Burst of Strength", "Protect from Melee")
// and returns the prayer ID. String lookup is case-insensitive via
// facts.PrayerByName.
func resolvePrayerID(v interp.Value) (int, error) {
	if i, ok := interp.AsInt(v); ok {
		return int(i), nil
	}
	if s, ok := v.(interp.String); ok {
		def := facts.PrayerByName(string(s))
		if def == nil {
			return 0, errf("unknown prayer %q (see facts.PrayerDef.xml)", string(s))
		}
		return def.ID, nil
	}
	return 0, errf("prayer must be Int slot or String name, got %s", v.Kind())
}

// ---------- tiny helpers ----------

// boolNamed reads an optional boolean named arg, returning def when the
// key is absent. A present-but-non-bool value is a programmer error
// (returns a Go error so the action wrapper surfaces it as a
// RuntimeError, matching the walk_to attempt_open_doors convention).
func boolNamed(verb string, named map[string]interp.Value, key string, def bool) (bool, error) {
	v, ok := named[key]
	if !ok {
		return def, nil
	}
	b, ok := v.(interp.Bool)
	if !ok {
		return false, errf("%s: %s must be a bool, got %s", verb, key, v.Kind())
	}
	return bool(b), nil
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

// formatChunkForRecall renders a corpus.Chunk into the single string
// routines see. Includes provenance so a routine that branches on
// content can also log where it came from.
func formatChunkForRecall(c corpus.Chunk) string {
	header := c.PageTitle
	if c.SectionTitle != "" {
		header = header + " § " + c.SectionTitle
	}
	return fmt.Sprintf("[%s § %s] %s", c.Source, header, c.Text)
}

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
