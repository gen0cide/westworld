// Package knowledge is the host's SEMANTIC world-knowledge ledger — the third
// leg of the mind alongside the trust ledger (who, limbic.Ledger) and the
// journal (what happened, memory.Journal). Where the trust ledger tracks Beta
// evidence about PEOPLE, this tracks Beta-backed BELIEFS about THINGS — npcs,
// places, shops, items, mechanics, quests — each with how it was learned
// (provenance) so the host can tell knowing from guessing and reconcile
// contradictions (direct observation overrides hearsay).
//
// It deliberately mirrors limbic.Ledger's shape (rowLocked / Observe / Seen /
// Known / Export / Import) so the runtime persistence wiring is parallel
// (runtime/knowledge.go ↔ runtime/limbic.go) and the same memory-namespace spine
// (bbolt local + mesa mirror) applies. Safe for concurrent use.
//
// This is the DATA layer of docs/world-knowledge-and-learning.md. The writers
// (perception → observations) and the distillation crons that grow it land in
// later phases; this package is the structure they read and write.
//
// The host's in-RAM copy is kept HOST-LIGHT: the ledger is bounded by a
// per-ledger subject cap and a per-subject belief cap (see DefaultMaxSubjects /
// DefaultMaxBeliefsPerEntry, configurable via SetCaps). Belief caps are enforced
// on the write paths; the host calls Prune on its flush tick to evict the
// least-recently-updated / lowest-evidence subjects deterministically. The
// durable mesa copy is the one the distillation crons grow + reconcile.
package knowledge

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// Provenance records HOW a belief was acquired. It is metadata the reasoning
// layer + distillation crons use to weigh contradictions (saw-it > did > told >
// guessed); it does not by itself alter stored confidence (the caller passes a
// confidence appropriate to the source).
const (
	ProvSystem   = "system"   // straight from the game/system — authoritative
	ProvObserved = "observed" // saw or did it myself
	ProvDeduced  = "deduced"  // inferred from other beliefs
	ProvHearsay  = "hearsay"  // someone told me (a player/NPC)
)

// Belief is one Beta(α,β)-backed claim about a subject, with its provenance.
// Exported fields so it JSON-round-trips for persistence.
type Belief struct {
	Claim      string  `json:"claim"`      // e.g. "sells iron/steel/rune pickaxe", "is at (122,509)"
	Provenance string  `json:"provenance"` // one of the Prov* constants
	Alpha      float64 `json:"alpha"`      // positive evidence
	Beta       float64 `json:"beta"`       // negative evidence
	At         int64   `json:"at"`         // unix seconds, last update
}

// Confidence is the Beta mean α/(α+β) ∈ [0,1] — P(the claim holds).
func (b Belief) Confidence() float64 {
	if b.Alpha+b.Beta <= 0 {
		return 0
	}
	return b.Alpha / (b.Alpha + b.Beta)
}

// Entry is the stored knowledge state for one subject. JSON-round-trips.
type Entry struct {
	Subject    string   `json:"subject"`
	Kind       string   `json:"kind"` // npc | location | shop | item | mechanic | quest | ...
	Beliefs    []Belief `json:"beliefs"`
	Encounters int      `json:"encounters"` // familiarity: times the subject was seen/touched
	LastSeen   int64    `json:"last_seen"`
	Tags       []string `json:"tags,omitempty"`
}

// Fact is a read-only view of what the host knows about a subject (parallel to
// limbic.Rel). Beliefs are sorted best-first by confidence; Confidence is the
// strongest belief's confidence (0 when nothing is known).
type Fact struct {
	Subject    string
	Kind       string
	Beliefs    []Belief
	Familiar   int
	Confidence float64
	Tags       []string
}

// Default caps keep the host-side ledger LIGHT and bounded (the distillation
// crons grow + prune the durable mesa copy; the host's in-RAM copy must NOT grow
// without bound). 0 in either cap means "unbounded" (the pre-cap behaviour).
const (
	DefaultMaxSubjects        = 512 // most subjects the host keeps locally
	DefaultMaxBeliefsPerEntry = 24  // most beliefs kept per subject
)

// Ledger is the host's world-knowledge ledger, keyed by normalized subject.
type Ledger struct {
	mu          sync.Mutex
	rows        map[string]*Entry
	now         func() int64 // injectable clock for tests
	maxSubjects int          // per-ledger subject cap (0 = unbounded)
	maxBeliefs  int          // per-subject belief cap (0 = unbounded)
}

// NewLedger returns an empty knowledge ledger with the default caps applied.
func NewLedger() *Ledger {
	return &Ledger{
		rows:        map[string]*Entry{},
		now:         func() int64 { return time.Now().Unix() },
		maxSubjects: DefaultMaxSubjects,
		maxBeliefs:  DefaultMaxBeliefsPerEntry,
	}
}

// SetCaps configures the per-ledger subject cap and the per-subject belief cap.
// A value <= 0 disables that cap (unbounded). Caps are enforced by Prune and on
// the per-subject belief write paths; this also re-enforces them immediately.
func (l *Ledger) SetCaps(maxSubjects, maxBeliefsPerEntry int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxSubjects = maxSubjects
	l.maxBeliefs = maxBeliefsPerEntry
	for _, e := range l.rows {
		l.capBeliefsLocked(e)
	}
	l.capSubjectsLocked()
}

func (l *Ledger) rowLocked(subject, kind string) *Entry {
	key := normalize(subject)
	e, ok := l.rows[key]
	if !ok {
		e = &Entry{Subject: strings.TrimSpace(subject), Kind: kind}
		l.rows[key] = e
	}
	if kind != "" {
		e.Kind = kind // a later, more specific kind wins
	}
	return e
}

// Note records a NEW belief (claim) about a subject with an initial confidence
// in [0,1] and its provenance. If the same claim already exists it is reinforced
// via Observe instead of duplicated. Returns the resulting belief.
func (l *Ledger) Note(subject, kind, claim, provenance string, confidence float64) Belief {
	confidence = clamp01(confidence)
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(subject, kind)
	now := l.now()
	e.LastSeen = now
	for i := range e.Beliefs {
		if e.Beliefs[i].Claim == claim {
			// Same claim restated: add weighted evidence, matching how the
			// initial belief is seeded (α=confidence, β=1-confidence) so a
			// restate at confidence c accumulates exactly like the first note —
			// not a hard ±1 vote that discards the magnitude.
			e.Beliefs[i].Alpha += confidence
			e.Beliefs[i].Beta += 1 - confidence
			if provenanceRank(provenance) > provenanceRank(e.Beliefs[i].Provenance) {
				e.Beliefs[i].Provenance = provenance // upgrade to the stronger source
			}
			e.Beliefs[i].At = now
			return e.Beliefs[i]
		}
	}
	b := Belief{Claim: claim, Provenance: provenance, Alpha: confidence, Beta: 1 - confidence, At: now}
	e.Beliefs = append(e.Beliefs, b)
	l.capBeliefsLocked(e)
	return b
}

// Observe adds evidence for/against an existing claim about a subject (good ⇒ α,
// else β), creating the claim if absent. weight scales the evidence (<=0 ⇒ 1).
func (l *Ledger) Observe(subject, claim string, good bool, weight float64) {
	if weight <= 0 {
		weight = 1
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(subject, "")
	now := l.now()
	e.LastSeen = now
	for i := range e.Beliefs {
		if e.Beliefs[i].Claim == claim {
			if good {
				e.Beliefs[i].Alpha += weight
			} else {
				e.Beliefs[i].Beta += weight
			}
			e.Beliefs[i].At = now
			return
		}
	}
	b := Belief{Claim: claim, Provenance: ProvObserved, At: now}
	if good {
		b.Alpha = weight
	} else {
		b.Beta = weight
	}
	e.Beliefs = append(e.Beliefs, b)
	l.capBeliefsLocked(e)
}

// Seen bumps a subject's familiarity (encounter count) without asserting a claim
// — the knowledge-side analogue of limbic.Ledger.Met.
func (l *Ledger) Seen(subject, kind string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(subject, kind)
	e.Encounters++
	e.LastSeen = l.now()
}

// Tag attaches a label to a subject if not already present.
func (l *Ledger) Tag(subject, tag string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(subject, "")
	if !slices.Contains(e.Tags, tag) {
		e.Tags = append(e.Tags, tag)
	}
}

// Known reports whether the ledger has anything on subject.
func (l *Ledger) Known(subject string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.rows[normalize(subject)]
	return ok
}

// Get returns the read-only Fact view for subject (empty Fact if unknown).
func (l *Ledger) Get(subject string) Fact {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.rows[normalize(subject)]; ok {
		return e.view()
	}
	return Fact{Subject: subject}
}

// All returns every Fact, sorted by descending familiarity.
func (l *Ledger) All() []Fact {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Fact, 0, len(l.rows))
	for _, e := range l.rows {
		out = append(out, e.view())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Familiar > out[j].Familiar })
	return out
}

func (e *Entry) view() Fact {
	beliefs := append([]Belief(nil), e.Beliefs...)
	sort.SliceStable(beliefs, func(i, j int) bool { return beliefs[i].Confidence() > beliefs[j].Confidence() })
	conf := 0.0
	if len(beliefs) > 0 {
		conf = beliefs[0].Confidence()
	}
	return Fact{
		Subject:    e.Subject,
		Kind:       e.Kind,
		Beliefs:    beliefs,
		Familiar:   e.Encounters,
		Confidence: conf,
		Tags:       append([]string(nil), e.Tags...),
	}
}

// --- persistence (snapshot for the memory layer) ---------------------------

// Export returns a JSON-serializable snapshot of every entry.
func (l *Ledger) Export() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, 0, len(l.rows))
	for _, e := range l.rows {
		cp := *e
		cp.Beliefs = append([]Belief(nil), e.Beliefs...)
		cp.Tags = append([]string(nil), e.Tags...)
		out = append(out, cp)
	}
	return out
}

// Import loads entries from a snapshot, replacing existing state. Each row is
// DEEP-copied (its Beliefs/Tags slices are cloned) so a later mutation of the
// caller's snapshot — e.g. an Export() taken from a still-live ledger — cannot
// reach into the imported state. When two snapshot rows normalize to the SAME
// subject key (a merged/cross-host snapshot), they are MERGED rather than
// silently last-writer-wins: beliefs are folded (max α/β per claim), and the
// max Encounters / latest LastSeen / union of Tags is kept. Caps are enforced
// after load so an over-large snapshot still leaves the ledger bounded.
func (l *Ledger) Import(rows []Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rows = make(map[string]*Entry, len(rows))
	for i := range rows {
		key := normalize(rows[i].Subject)
		if existing, dup := l.rows[key]; dup {
			mergeEntry(existing, rows[i]) // collision: fold, don't clobber
			continue
		}
		e := rows[i]
		e.Beliefs = append([]Belief(nil), rows[i].Beliefs...)
		e.Tags = append([]string(nil), rows[i].Tags...)
		l.rows[key] = &e
	}
	for _, e := range l.rows {
		l.capBeliefsLocked(e)
	}
	l.capSubjectsLocked()
}

// ImportMerge folds a snapshot into the live ledger NON-DESTRUCTIVELY (M17 host-
// side). Unlike Import (which REPLACES the whole map and is the cold-start path),
// this is the WARM-host re-import: a running host pulls the now server-reconciled
// ledger (the consolidation/insight crons' β-bumps + reconciliations) into its
// local copy so those survive the host's next wholesale flush. Existing subjects
// are FOLDED via the same per-(claim) max-evidence collision logic Import uses
// (mergeEntry) — additively, taking the stronger evidence + provenance + the union
// of tags + max familiarity — so a cron belief the host never re-learned is adopted
// while a live row is never wiped. A subject the snapshot omits is left untouched
// (the cron doesn't own it). Caps are re-enforced after the merge so a re-import
// can never push the ledger past its bounds.
func (l *Ledger) ImportMerge(rows []Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range rows {
		key := normalize(rows[i].Subject)
		if existing, ok := l.rows[key]; ok {
			mergeEntry(existing, rows[i]) // fold cron evidence in, don't clobber
			continue
		}
		e := rows[i]
		e.Beliefs = append([]Belief(nil), rows[i].Beliefs...)
		e.Tags = append([]string(nil), rows[i].Tags...)
		l.rows[key] = &e
	}
	for _, e := range l.rows {
		l.capBeliefsLocked(e)
	}
	l.capSubjectsLocked()
}

// mergeEntry folds src into dst in place (same normalized subject). Beliefs are
// merged per Claim taking the MAX α and MAX β (the stronger evidence on each
// side, mirroring how a re-Note/Observe accumulates) and the stronger
// provenance; Encounters takes the max, LastSeen the latest, Tags the union.
func mergeEntry(dst *Entry, src Entry) {
	if src.Kind != "" {
		dst.Kind = src.Kind // a more-specific kind wins (parallel to rowLocked)
	}
	for _, sb := range src.Beliefs {
		found := false
		for i := range dst.Beliefs {
			if dst.Beliefs[i].Claim == sb.Claim {
				dst.Beliefs[i].Alpha = maxF(dst.Beliefs[i].Alpha, sb.Alpha)
				dst.Beliefs[i].Beta = maxF(dst.Beliefs[i].Beta, sb.Beta)
				if provenanceRank(sb.Provenance) > provenanceRank(dst.Beliefs[i].Provenance) {
					dst.Beliefs[i].Provenance = sb.Provenance
				}
				if sb.At > dst.Beliefs[i].At {
					dst.Beliefs[i].At = sb.At
				}
				found = true
				break
			}
		}
		if !found {
			dst.Beliefs = append(dst.Beliefs, sb)
		}
	}
	if src.Encounters > dst.Encounters {
		dst.Encounters = src.Encounters
	}
	if src.LastSeen > dst.LastSeen {
		dst.LastSeen = src.LastSeen
	}
	for _, t := range src.Tags {
		if !slices.Contains(dst.Tags, t) {
			dst.Tags = append(dst.Tags, t)
		}
	}
}

// Prune enforces the configured caps deterministically: each subject is trimmed
// to maxBeliefs (dropping the lowest-confidence, oldest-tie beliefs first), then
// the ledger is trimmed to maxSubjects (evicting the least-recently-updated,
// lowest-Encounters subjects first). Intended to be called by the host on its
// flush tick so the in-RAM ledger stays host-light. A no-op when both caps are
// <= 0. Returns the number of subjects evicted.
func (l *Ledger) Prune() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.rows {
		l.capBeliefsLocked(e)
	}
	return l.capSubjectsLocked()
}

// capBeliefsLocked trims e.Beliefs to l.maxBeliefs, dropping the weakest beliefs
// first (lowest Confidence; ties broken by oldest At, then lexicographic Claim
// for full determinism). Caller holds l.mu.
func (l *Ledger) capBeliefsLocked(e *Entry) {
	if l.maxBeliefs <= 0 || len(e.Beliefs) <= l.maxBeliefs {
		return
	}
	sort.SliceStable(e.Beliefs, func(i, j int) bool {
		ci, cj := e.Beliefs[i].Confidence(), e.Beliefs[j].Confidence()
		if ci != cj {
			return ci > cj // keep the strongest
		}
		if e.Beliefs[i].At != e.Beliefs[j].At {
			return e.Beliefs[i].At > e.Beliefs[j].At // keep the freshest
		}
		return e.Beliefs[i].Claim < e.Beliefs[j].Claim
	})
	e.Beliefs = e.Beliefs[:l.maxBeliefs]
}

// capSubjectsLocked trims l.rows to l.maxSubjects, evicting the least-valuable
// subjects first (oldest LastSeen; ties broken by lowest Encounters, then by
// normalized key for full determinism). Returns the count evicted. Caller holds
// l.mu.
func (l *Ledger) capSubjectsLocked() int {
	if l.maxSubjects <= 0 || len(l.rows) <= l.maxSubjects {
		return 0
	}
	type keyed struct {
		key string
		e   *Entry
	}
	all := make([]keyed, 0, len(l.rows))
	for k, e := range l.rows {
		all = append(all, keyed{k, e})
	}
	// Sort BEST-first (kept), worst at the tail (evicted).
	sort.Slice(all, func(i, j int) bool {
		if all[i].e.LastSeen != all[j].e.LastSeen {
			return all[i].e.LastSeen > all[j].e.LastSeen // keep most-recently-seen
		}
		if all[i].e.Encounters != all[j].e.Encounters {
			return all[i].e.Encounters > all[j].e.Encounters // keep most-familiar
		}
		return all[i].key < all[j].key
	})
	evicted := 0
	for _, k := range all[l.maxSubjects:] {
		delete(l.rows, k.key)
		evicted++
	}
	return evicted
}

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// provenanceRank orders sources by trustworthiness so a stronger source upgrades
// a restated claim's provenance.
func provenanceRank(p string) int {
	switch p {
	case ProvSystem:
		return 4
	case ProvObserved:
		return 3
	case ProvDeduced:
		return 2
	case ProvHearsay:
		return 1
	default:
		return 0
	}
}
