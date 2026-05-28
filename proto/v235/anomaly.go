package v235

import (
	"encoding/hex"
	"log/slog"
)

// Inline anomaly assertions for decoders. Per
// docs/lang/protocol-debug-strategy.md Tier 1, the goal is to
// surface "obviously out of range" field values as WARN log
// lines with hex of the surrounding bytes, so cascading
// off-by-N decoder bugs (like the inventory amount bug) show
// up immediately instead of after an hour of debugging.
//
// The plausible-range tables here are NOT precise specs — they
// reflect "anything outside this range is almost certainly a
// decode error, not legitimate data". Adjust over time as we
// discover edge cases.

const (
	// Item IDs: facts table has ~1500 entries. Anything above
	// 5000 is almost certainly a decode error (no legit RSC
	// item id exceeds the def table). Above 32768 means we
	// read garbage where an id should have been.
	plausibleMaxItemID = 5000

	// NPC server indices: typically < 5000 in normal play.
	// Above 0xF000 means we likely read a wielded-bit field or
	// other high-bit-set short as an index.
	plausibleMaxNpcIndex = 0xF000

	// World tile coords: RSC map is ~1000x4000. Values above
	// 16384 mean we likely read a non-coord field.
	plausibleMaxCoord = 16384

	// Damage per hit: hardly ever > 100 even at high levels.
	// 1000+ means the field is not damage.
	plausibleMaxDamage = 1000

	// Inventory amount for stackables: 2^31-1 is the legit
	// hard ceiling (game caps at ~2B gp). But for unstackables
	// the amount is always 1, so seeing >1 there is a bug.
	plausibleMaxAmount = 2_000_000_000
)

// anomalyLogger is the slog default — packages don't get a
// configured logger via DI, so we just use the global. Callers
// that want to silence this in tests can swap slog.Default.
func anomalyLogger() *slog.Logger { return slog.Default() }

// flagAnomaly logs a WARN when a decoded field is wildly out of
// expected range. Callers pass a label, the value, the plausible
// max, and the raw bytes that produced it for cross-referencing.
// Cheap; runs once per decoded field. The WARN line is what saves
// hours when a cascading off-by-N bug starts producing garbage.
func flagAnomaly(field string, value, plausibleMax int, raw []byte) {
	if value <= plausibleMax {
		return
	}
	anomalyLogger().Warn("proto decode anomaly",
		"field", field,
		"value", value,
		"plausible_max", plausibleMax,
		"raw_bytes", hex.EncodeToString(raw),
		"hint", "field decoded out of plausible range — likely off-by-N or field misalignment upstream",
	)
}
