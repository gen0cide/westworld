package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Movement / navigation action handler bodies (walk_to / go_to /
// walk_path / is_reachable / follow). Registered in the central
// actionHandlers table in dsl_actions.go.

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
		// Gazetteer places are plane-0 footprint coords; use the
		// plane-local Y so an upstairs host still names the right area.
		if lm, _, ok := h.facts.Gazetteer().NearestPlace(stuck.X, stuck.Y%world.PlaneHeight); ok {
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
	// A wrapped DoorLockedError carries the REAL diagnosis (which barrier,
	// the server's own words, what would unblock it) — surface that as
	// DOOR_LOCKED instead of the canned guess. errors.As because GoTo
	// wraps with %w.
	var doorErr *DoorLockedError
	if errors.As(err, &doorErr) {
		detail := doorErr.ServerMessage
		if detail == "" {
			detail = "the barrier did not open"
		}
		hint := doorErr.Precondition
		if hint == "" {
			hint = "find another way around, or satisfy whatever locks it"
		}
		return interp.Fail(interp.DOOR_LOCKED, fmt.Sprintf(
			"go_to %s blocked by a locked barrier at (%d, %d): %s — %s (you are at (%d, %d)%s)",
			target, doorErr.DoorX, doorErr.DoorY, detail, hint, stuck.X, stuck.Y, near))
	}
	return interp.Fail(interp.PATH_BLOCKED, fmt.Sprintf(
		"go_to %s blocked: stuck at (%d, %d)%s — a gate may need payment or be quest-locked; pay it if you have coins, pick another destination of this type, or route around (%v)",
		target, stuck.X, stuck.Y, near, err))
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
