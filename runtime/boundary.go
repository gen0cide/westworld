package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
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
