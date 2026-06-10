package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/world"
)

// Walk sends one walk-to-coord packet, with FOV validation against
// the bot's current position. Errors with action.ErrOutOfRange if the
// target is more than MaxClickRange tiles away.
//
// For long journeys, use WalkTo (which chunks into multiple in-FOV
// segments).
func (h *Host) Walk(ctx context.Context, x, y int) error {
	pos := h.world.Self.Position()
	return action.Walk(ctx, h.conn, pos.X, pos.Y, x, y)
}

// DoorLockedError is returned by WalkTo when an openable boundary
// on the path was contacted, the open interaction was attempted,
// but the host couldn't pass after the retry budget. Distinct
// from a plain stall (which means "no door, just terrain") so the
// DSL layer can map it to ErrorCode DOOR_LOCKED instead of the
// generic PATH_BLOCKED.
//
// ServerMessage captures whatever
// world.Recent.LastServerMessage held immediately after the
// failing open attempt — typically the server's "you need a key"
// / "for members only" prose, so routines can branch on the text.
// Empty if no relevant message was observed in the window.
type DoorLockedError struct {
	DoorX, DoorY, DoorDir int
	Attempts              int
	ServerMessage         string
	// Kind/Precondition classify the refusal (classifyRefusal): locked vs
	// toll vs level/quest requirement — and what would satisfy it. Carried
	// through go_to/walk_to so cognition plans (pay / fetch key / route
	// around) instead of guessing from a canned hint.
	Kind         RefusalKind
	Precondition string
}

func (e *DoorLockedError) Error() string {
	if e.ServerMessage != "" {
		return fmt.Sprintf("door locked at (%d, %d) dir=%d after %d attempt(s): %s",
			e.DoorX, e.DoorY, e.DoorDir, e.Attempts, e.ServerMessage)
	}
	return fmt.Sprintf("door locked at (%d, %d) dir=%d after %d attempt(s)",
		e.DoorX, e.DoorY, e.DoorDir, e.Attempts)
}

// WalkOptions tunes WalkTo behavior. Zero value = sensible defaults
// (currently: attempt to open any closed door blocking the path).
// Construct via DefaultWalkOptions() and override fields rather than
// initializing directly, so future field additions don't break callers.
type WalkOptions struct {
	// AttemptOpenDoors, when true, makes WalkTo try to open an
	// adjacent openable boundary (door / doorframe) on stall and
	// retry the walk. Mirrors the Java RSC client's auto-door
	// behavior. Default: true. Set to false for routines that
	// want strict "stop at any obstacle" semantics (e.g. quest
	// checks that need to detect locked doors).
	//
	// If the door is truly locked (e.g. quest gate), the open
	// interaction succeeds at the packet layer but the host
	// can't pass; on a second stall at the same tile WalkTo
	// stops trying and returns PATH_BLOCKED with the stall pos
	// so the script can react.
	AttemptOpenDoors bool
}

// DefaultWalkOptions returns a WalkOptions with defaults applied:
// attempt-open-doors enabled. Callers that want non-default behavior
// should start here and tweak.
func DefaultWalkOptions() WalkOptions {
	return WalkOptions{AttemptOpenDoors: true}
}

// WalkTo navigates to (x, y) using the BFS pathfinder. Wraps
// WalkToOpts with default options.
func (h *Host) WalkTo(ctx context.Context, x, y int) error {
	return h.WalkToOpts(ctx, x, y, DefaultWalkOptions())
}

// WalkToOpts is WalkTo with explicit options. The DSL `walk_to`
// builtin routes through here, exposing the options as named args.
//
// The flow mirrors the reference clients over a grid that models CLOSED
// doors and gates as walls (the server's own collision rule):
//
//	plan (BFS) → walk → on no-path or stall: CLASSIFY the adjacent
//	obstacle (wall-door vs scenery-gate, straddle-ordered toward the
//	goal) → open via the obstacle's own channel → CONFIRM BY STATE
//	(def swap in the live store / position advance / plane change) →
//	replan from the new world.
//
// There is deliberately NO blind direct-walk fallback: a no-path with no
// openable obstacle is a real PATH_BLOCKED the caller must replan around
// (the old raw far-target Walk straight-lined into walls — the server
// stops at the first wall and never detours).
func (h *Host) WalkToOpts(ctx context.Context, x, y int, opts WalkOptions) error {
	const (
		pollInterval = 200 * time.Millisecond
		stallTimeout = 5 * time.Second
		arriveRadius = 1
	)
	// Per-obstacle open attempts + captured server prose, keyed by the
	// obstacle's identifying tile+dir. Caps re-tries on a locked door so
	// the walk fails loudly (DoorLockedError) instead of spinning.
	attempts := map[[3]int]int{}
	messages := map[[3]int]string{}
	for {
	replan:
		if err := ctx.Err(); err != nil {
			return err
		}
		pos := h.world.Self.Position()
		if absVal(x-pos.X) <= arriveRadius && absVal(y-pos.Y) <= arriveRadius {
			h.log.Debug("walkto arrived", "pos", fmt.Sprintf("(%d, %d)", pos.X, pos.Y))
			return nil
		}

		corners, pathErr := h.pathToTile(x, y, false)
		if pathErr != nil {
			if errors.Is(pathErr, ErrNoPath) {
				// The grid models closed doors/gates as walls, so a
				// no-path very often means a traversable obstacle sits
				// between us and the goal. Classify and open it, then
				// replan over the updated world.
				if opts.AttemptOpenDoors {
					opened, terr := h.tryTraverse(ctx, x, y, attempts, messages)
					if terr != nil {
						return terr
					}
					if opened {
						goto replan
					}
				}
				return fmt.Errorf("walkto: %w (no route from (%d, %d) to (%d, %d))", pathErr, pos.X, pos.Y, x, y)
			}
			return fmt.Errorf("walkto: pathfind: %w", pathErr)
		}
		if len(corners) == 0 {
			return nil // already at target
		}
		h.log.Debug("walkto path",
			"from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
			"target", fmt.Sprintf("(%d, %d)", x, y),
			"corners", len(corners),
			"final_corner", fmt.Sprintf("(%d, %d)", corners[len(corners)-1][0], corners[len(corners)-1][1]),
		)
		if err := action.WalkPath(ctx, h.conn, corners); err != nil {
			return fmt.Errorf("walkto: send path: %w", err)
		}

		// Wait for arrival OR stall (no position change for stallTimeout).
		startPos := pos
		stallDeadline := time.Now().Add(stallTimeout)
		arrived := false
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
			}
			cur := h.world.Self.Position()
			if absVal(cur.X-x) <= arriveRadius && absVal(cur.Y-y) <= arriveRadius {
				arrived = true
				break
			}
			if cur.X != startPos.X || cur.Y != startPos.Y {
				startPos = cur
				stallDeadline = time.Now().Add(stallTimeout)
				continue
			}
			if time.Now().After(stallDeadline) {
				// Stalled mid-walk: the server refused a step the grid
				// thought open (a door someone closed under us, or live
				// state we haven't seen). Same recovery as no-path.
				if opts.AttemptOpenDoors {
					opened, terr := h.tryTraverse(ctx, x, y, attempts, messages)
					if terr != nil {
						return terr
					}
					if opened {
						goto replan
					}
				}
				h.log.Warn("walkto stalled",
					"at", fmt.Sprintf("(%d, %d)", cur.X, cur.Y),
					"target", fmt.Sprintf("(%d, %d)", x, y),
				)
				return fmt.Errorf("walkto: stalled at (%d, %d) targeting (%d, %d)", cur.X, cur.Y, x, y)
			}
		}
		if arrived {
			return nil
		}
		// Loop to replan from the new position (server truncated the
		// path short of the target — live state will explain why).
	}
}

// tryTraverse classifies the actionable obstacle nearest the host's intended
// step toward (x, y), opens it via its own channel (boundary interaction for
// wall doors, scenery interaction for gates), and confirms BY STATE. Returns
// opened=true when an obstacle was confirmed traversed (caller replans), a
// DoorLockedError when an obstacle exhausted its attempts (locked / toll /
// quest-gated — the server prose is attached), or (false, nil) when nothing
// actionable is adjacent.
func (h *Host) tryTraverse(ctx context.Context, x, y int, attempts map[[3]int]int, messages map[[3]int]string) (bool, error) {
	const (
		maxOpenAttempts = 2
		// totalOpenBudget bounds opens across one WalkToOpts call so a
		// pathological corridor of doors can't spin forever.
		totalOpenBudget = 6
	)
	total := 0
	for _, n := range attempts {
		total += n
	}
	if total >= totalOpenBudget {
		return false, nil
	}
	cur := h.world.Self.Position()
	// Adjacent obstacles first (stall recovery, straddle-ordered); when
	// nothing is within reach — the plan-time no-path case, where the
	// blocking door may be far from BOTH host and goal — scan the grid
	// window for openable barriers ON the route.
	cands := h.findTraversableNear(cur.X, cur.Y, x, y)
	if len(cands) == 0 {
		cands = h.findTraversableToward(x, y)
	}
	for _, t := range cands {
		if t.Kind == TraversableLadder {
			continue // plane transitions are routed deliberately, never auto-climbed
		}
		key := [3]int{t.X, t.Y, t.Dir}
		if attempts[key] >= maxOpenAttempts {
			// Exhausted: locked / toll / quest gate. LEDGER it (TTL-bounded)
			// so subsequent replans skip it, and fail loudly with the
			// classified server prose so cognition can plan (pay, fetch a
			// key, route around) instead of spinning.
			kind, precond := classifyRefusal(messages[key])
			if h.blocked != nil {
				h.blocked.Note(BlockedEdge{
					X: t.X, Y: t.Y, Dir: t.Dir,
					Kind: kind, Prose: messages[key], Precondition: precond,
				})
			}
			h.log.Warn("walkto: obstacle locked",
				"kind", t.Kind, "name", t.Name,
				"at", fmt.Sprintf("(%d, %d, dir=%d)", t.X, t.Y, t.Dir),
				"attempts", attempts[key],
				"server_msg", messages[key],
				"refusal", kind,
			)
			return false, &DoorLockedError{
				DoorX:         t.X,
				DoorY:         t.Y,
				DoorDir:       t.Dir,
				Attempts:      attempts[key],
				ServerMessage: messages[key],
				Kind:          kind,
				Precondition:  precond,
			}
		}
		attempts[key]++

		// Snapshot the latest server message BEFORE the open so a
		// rejection ("the door is locked", "you must pay a toll of 10
		// gold coins") arriving in the confirm window is attributable.
		var preMsgAt time.Time
		if prev := h.world.Recent.ServerMessage(); prev != nil {
			preMsgAt = prev.At
		}
		h.log.Info("walkto: opening obstacle in path",
			"kind", t.Kind, "name", t.Name,
			"at", fmt.Sprintf("(%d, %d, dir=%d)", t.X, t.Y, t.Dir),
			"attempt", attempts[key],
		)
		if err := t.Open(ctx); err != nil {
			h.log.Warn("walkto: open dispatch failed", "err", err)
			continue
		}
		h.awaitTraversalConfirm(ctx, t, cur)
		if msg := h.world.Recent.ServerMessage(); msg != nil && msg.At.After(preMsgAt) {
			messages[key] = msg.Message
		}
		// Replan REGARDLESS of confirmation: the BFS re-reads the live
		// stores, so a confirmed swap routes through immediately, and an
		// unconfirmed (locked) obstacle hits no-path again — one retry,
		// then the attempts cap surfaces DoorLockedError with the prose.
		return true, nil
	}
	return false, nil
}

// awaitTraversalConfirm polls (~2.5s) for STATE evidence that an opened
// obstacle is traversed/traversable — never "the packet was sent":
//
//   - wall door:    the live boundary def at (X,Y,Dir) swapped to a
//     non-blocking def (the authentic wire NEVER removes in-range
//     boundaries — open == ID overwrite to e.g. doorframe 11);
//   - scenery gate: the live scenery def at the loc CHANGED from the
//     closed id (gate 60 → open 59) or the object was removed;
//   - any:          the host's position advanced (doDoor walks-then-
//     teleports THROUGH; mandatory signal — doDoor(replaceID=-1)
//     streams no def swap at all), or the plane band changed.
func (h *Host) awaitTraversalConfirm(ctx context.Context, t Traversable, from world.Coord) bool {
	const (
		confirmWindow = 2500 * time.Millisecond
		poll          = 200 * time.Millisecond
	)
	deadline := time.Now().Add(confirmWindow)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(poll):
		}
		cur := h.world.Self.Position()
		// Position-advance only confirms when the obstacle was ADJACENT
		// at open time: for a far obstacle the Open's own approach-walk
		// moves the host, which proves nothing about the obstacle.
		wasAdjacent := absVal(from.X-t.X) <= 1 && absVal(from.Y-t.Y) <= 1
		if wasAdjacent && (cur.X != from.X || cur.Y != from.Y) {
			return true // moved — the open walked/teleported us through
		}
		if cur.Y/world.PlaneHeight != from.Y/world.PlaneHeight {
			return true // plane band changed (ladder/stairs)
		}
		switch t.Kind {
		case TraversableWallDoor:
			if h.world.Boundaries != nil {
				if id, ok := h.world.Boundaries.Get(t.X, t.Y, t.Dir); ok {
					if id < 0 {
						return true // removed (defensive)
					}
					if d := h.facts.BoundaryDef(id); d != nil && !d.BlocksWhenClosed() && !d.BlocksMovement() {
						return true // swapped to the open def
					}
				}
			}
		case TraversableSceneryGate:
			if h.world.Scenery != nil {
				if rec, ok := h.world.Scenery.At(t.X, t.Y); ok && rec.ID != t.ClosedID {
					return true // def swap (closed gate -> open gate)
				}
				if h.world.Scenery.IsRemoved(t.X, t.Y) {
					return true
				}
			}
		}
		if time.Now().After(deadline) {
			return false
		}
	}
}

func absVal(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
