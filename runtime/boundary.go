package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/facts"
)

// InteractWithBoundary fires the default click on a boundary (door,
// ladder, fence) — e.g., to open a closed door.
//
// OpenRSC's GameObjectWallAction handler only sets a one-shot
// WalkToAction; it does NOT initiate walking itself. We pathfind a
// route to a tile inside the boundary's atObject rectangle, send the
// multi-corner walk packet (WALK_TO_POINT, since the target is an
// object not a mob), then the interact packet.
func (h *Host) InteractWithBoundary(ctx context.Context, x, y, direction int) error {
	h.log.Info("InteractWithBoundary: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.InteractWithBoundary(ctx, h.conn, x, y, direction)
	})
}

// findOpenableNear looks for an openable boundary (door / doorframe;
// i.e. BoundaryDef.Unknown == 1) at or directly adjacent (Chebyshev
// distance 1) to (x, y). Returns the first match, or nil if none
// found. Used by WalkTo's open_doors=true path: when a walk stalls,
// we infer "the next tile is gated by a door" by looking up the
// boundary the server's collision check would have blocked us at.
//
// We don't try to figure out which side of the player the door is
// on — the InteractWithBoundary packet carries the boundary's own
// (x, y, direction), and the server handles the geometry. Returning
// any openable boundary in immediate range is enough.
func (h *Host) findOpenableNear(x, y int) *facts.BoundaryLoc {
	if h.facts == nil {
		return nil
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			for _, p := range h.facts.At(x+dx, y+dy) {
				if p.Kind != "boundary" {
					continue
				}
				def := h.facts.BoundaryDef(p.DefID)
				if def == nil {
					continue
				}
				// Unknown=1 marks openable boundaries (doors,
				// doorframes). Plain walls/fences have Unknown=0
				// and BlocksMovement=true.
				if def.Unknown != 1 {
					continue
				}
				return &facts.BoundaryLoc{
					DefID:     p.DefID,
					X:         p.X,
					Y:         p.Y,
					Direction: p.Direction,
				}
			}
		}
	}
	return nil
}
