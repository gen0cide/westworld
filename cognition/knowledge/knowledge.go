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

// Ledger is the host's world-knowledge ledger, keyed by normalized subject.
type Ledger struct {
	mu   sync.Mutex
	rows map[string]*Entry
	now  func() int64 // injectable clock for tests
}

// NewLedger returns an empty knowledge ledger.
func NewLedger() *Ledger {
	return &Ledger{rows: map[string]*Entry{}, now: func() int64 { return time.Now().Unix() }}
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
			// Same claim restated: add evidence in the direction of `confidence`.
			if confidence >= 0.5 {
				e.Beliefs[i].Alpha++
			} else {
				e.Beliefs[i].Beta++
			}
			if provenanceRank(provenance) > provenanceRank(e.Beliefs[i].Provenance) {
				e.Beliefs[i].Provenance = provenance // upgrade to the stronger source
			}
			e.Beliefs[i].At = now
			return e.Beliefs[i]
		}
	}
	b := Belief{Claim: claim, Provenance: provenance, Alpha: confidence, Beta: 1 - confidence, At: now}
	e.Beliefs = append(e.Beliefs, b)
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

// Import loads entries from a snapshot, replacing existing state.
func (l *Ledger) Import(rows []Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rows = make(map[string]*Entry, len(rows))
	for i := range rows {
		e := rows[i]
		l.rows[normalize(e.Subject)] = &e
	}
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
