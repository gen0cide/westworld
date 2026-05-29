package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Prayer action handler bodies (activate_prayer / deactivate_prayer).
// Registered in the central actionHandlers table in dsl_actions.go.
// Prayer-id resolution lives in dsl_helpers.go (resolvePrayerID).

// dslActivatePrayer turns on a prayer slot.
// activate_prayer(N) where N is 0..13. Server may silently reject
// (low prayer level or zero prayer points) — routines should check
// world.prayer.active(N) after.
func dslActivatePrayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("activate_prayer takes 1 arg (prayer name or slot)")
	}
	id, err := resolvePrayerID(args[0])
	if err != nil {
		return nil, errf("activate_prayer: %v", err)
	}
	if err := h.ActivatePrayer(ctx, id); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslDeactivatePrayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("deactivate_prayer takes 1 arg (prayer name or slot)")
	}
	id, err := resolvePrayerID(args[0])
	if err != nil {
		return nil, errf("deactivate_prayer: %v", err)
	}
	if err := h.DeactivatePrayer(ctx, int(id)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
