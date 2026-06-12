package mesad

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// Cross-host extraction dedup: at a co-located knot, every hearer latches the
// same speaker and ships a near-identical window — 20 hosts overhearing one
// line cost 20 Haiku calls for one set of claims (measured ~1,000 calls/min
// at pickup-spot convergence, 2026-06-11). Identical windows now share one
// extraction: the result is cached briefly and replayed per hearer, so each
// host still writes the claims into its OWN ledger — only the LLM call is
// shared. Exact-match by design: window text differs the moment a hearer's
// own replies fan in, which is precisely when its perspective differs enough
// to deserve its own extraction.
//
// extractDedupTTL is a var so tests can shrink it.
var extractDedupTTL = 90 * time.Second

const extractDedupCap = 512

type extractDedupEntry struct {
	set *mesapb.ExtractedDialogSet // treated as immutable after parse
	at  time.Time
}

type extractDedup struct {
	mu sync.Mutex
	m  map[string]extractDedupEntry
}

func newExtractDedup() *extractDedup { return &extractDedup{m: make(map[string]extractDedupEntry)} }

// key fingerprints the exchange content — speaker, role, and the window lines —
// deliberately NOT the hearer (that is the whole point).
func (d *extractDedup) key(w *mesapb.DialogWindow) string {
	h := sha256.Sum256([]byte(w.GetSpeaker() + "\x00" + w.GetSpeakerRole() + "\x00" + strings.Join(w.GetWindow(), "\n")))
	return hex.EncodeToString(h[:16])
}

func (d *extractDedup) get(key string, now time.Time) (*mesapb.ExtractedDialogSet, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.m[key]
	if !ok || now.Sub(e.at) > extractDedupTTL {
		return nil, false
	}
	return e.set, true
}

func (d *extractDedup) put(key string, set *mesapb.ExtractedDialogSet, now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.m) >= extractDedupCap {
		// Cheap pressure valve: drop expired entries; if none expired, drop
		// arbitrary ones (map order) until under cap. Correctness is unaffected
		// — a dropped entry just costs one extra LLM call.
		for k, e := range d.m {
			if now.Sub(e.at) > extractDedupTTL || len(d.m) >= extractDedupCap {
				delete(d.m, k)
			}
		}
	}
	d.m[key] = extractDedupEntry{set: set, at: now}
}
