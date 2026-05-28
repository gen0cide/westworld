package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/session"
)

// Prayer outbound opcodes per Payload235Parser.java.
const (
	outPrayerActivated   byte = 60
	outPrayerDeactivated byte = 254
)

// ActivatePrayer turns on the given prayer slot. prayerID is the
// RSC prayer index 0..13 (Thick Skin, Burst of Strength, ...).
// Server rejects if the player's prayer level is below the
// requirement OR if prayer points are 0.
func ActivatePrayer(ctx context.Context, conn *session.Conn, prayerID int) error {
	if prayerID < 0 || prayerID > 0xFF {
		return fmt.Errorf("action: ActivatePrayer prayerID %d out of byte range", prayerID)
	}
	return conn.Send(outPrayerActivated, []byte{byte(prayerID)})
}

// DeactivatePrayer turns off the given prayer slot.
func DeactivatePrayer(ctx context.Context, conn *session.Conn, prayerID int) error {
	if prayerID < 0 || prayerID > 0xFF {
		return fmt.Errorf("action: DeactivatePrayer prayerID %d out of byte range", prayerID)
	}
	return conn.Send(outPrayerDeactivated, []byte{byte(prayerID)})
}
