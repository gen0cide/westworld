package runtime

import (
	"fmt"
	"sort"
	"time"
)

// decisionCacheTTL bounds how long a memoized decide() verdict stays valid. The
// key already folds in coarse mood/HP, so a meaningful change re-decides before
// this; the TTL just caps staleness for slow drift on an otherwise-stable state.
const decisionCacheTTL = 5 * time.Minute

// decisionCacheKey derives a stable key for a decide() call: the question + the
// option SET (order-independent) + COARSE host state (HP tier, fatigue tier,
// mood buckets). Volatile state (exact position, inventory) is omitted so a
// repeated decision in materially-the-same state reuses the cached verdict,
// while meaningful change (HP drop, mood shift) changes the key → a fresh ask.
func (h *Host) decisionCacheKey(question string, options []string) string {
	opts := append([]string(nil), options...)
	sort.Strings(opts) // option order must not split the key

	self := h.world.Self
	hpTier := 0
	if mx := self.MaxHP(); mx > 0 {
		hpTier = self.HP() * 5 / mx // 0..5 HP quintile
	}
	fatTier := self.Fatigue() / 150 // 0..5 (server fatigue is 0..750)

	var stress, conf, val float64
	if h.affect != nil {
		stress, conf, val = h.affect.Snapshot()
	}
	raw := fmt.Sprintf("q=%s|o=%v|hp%d|ft%d|a%d,%d,%d",
		question, opts, hpTier, fatTier,
		int(stress*4), int(conf*4), int((val+1)*4)) // coarse mood buckets
	return shortHash(raw)
}
