package runtime

import (
	"context"

	"github.com/gen0cide/westworld/action"
)

// ActivatePrayer turns on a prayer slot. Server rejects (and emits
// no SEND_PRAYERS_ACTIVE update) if our prayer level is below the
// requirement or prayer points are 0.
func (h *Host) ActivatePrayer(ctx context.Context, prayerID int) error {
	return action.ActivatePrayer(ctx, h.conn, prayerID)
}

// DeactivatePrayer turns off a prayer slot.
func (h *Host) DeactivatePrayer(ctx context.Context, prayerID int) error {
	return action.DeactivatePrayer(ctx, h.conn, prayerID)
}
