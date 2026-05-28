package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/event"
)

// Follow sends the server-side PLAYER_FOLLOW packet (opcode 165) to
// stick adjacent to the named player. After the server accepts, IT
// handles all the walking — we just have to keep the session alive.
//
// Blocks until ctx is cancelled OR until startupTimeout expires
// while waiting for the target to first become visible.
//
// We learn the target's serverIndex from inbound OtherPlayerAppearance
// events. Once we've sent the follow packet, the server takes over;
// we periodically re-send the follow packet to handle the case where
// the target moves out of view and back into it.
func (h *Host) Follow(ctx context.Context, targetName string, startupTimeout time.Duration) error {
	if targetName == "" {
		return errors.New("Follow: empty target name")
	}
	target := strings.ToLower(targetName)
	sub := h.bus.Subscribe("*", 256)

	// Wait for the target to first become visible (so we learn their
	// serverIndex).
	waitCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	var serverIndex int = -1
	for serverIndex == -1 {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("Follow: target %q never became visible: %w", targetName, waitCtx.Err())
		case ev, ok := <-sub:
			if !ok {
				return errors.New("Follow: bus closed")
			}
			if ap, isAp := ev.(event.OtherPlayerAppearance); isAp {
				if strings.EqualFold(ap.Name, target) {
					serverIndex = ap.PlayerIndex
				}
			}
		}
	}
	h.log.Info("follow: target acquired",
		"name", targetName,
		"server_index", serverIndex,
	)

	// Send the follow request.
	if err := action.FollowPlayer(ctx, h.conn, serverIndex); err != nil {
		return fmt.Errorf("Follow: send: %w", err)
	}

	// Refresh the follow request every 10s in case the target moved
	// out of view momentarily and the server stopped following.
	refresh := time.NewTicker(10 * time.Second)
	defer refresh.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-sub:
			if !ok {
				return nil
			}
			// Re-acquire serverIndex from new appearance events (the
			// index can change between sessions).
			if ap, isAp := ev.(event.OtherPlayerAppearance); isAp && strings.EqualFold(ap.Name, target) {
				if ap.PlayerIndex != serverIndex {
					h.log.Info("follow: target index changed",
						"old", serverIndex,
						"new", ap.PlayerIndex,
					)
					serverIndex = ap.PlayerIndex
					_ = action.FollowPlayer(ctx, h.conn, serverIndex)
				}
			}
		case <-refresh.C:
			if err := action.FollowPlayer(ctx, h.conn, serverIndex); err != nil {
				h.log.Warn("follow: refresh failed", "err", err)
			}
		}
	}
}
