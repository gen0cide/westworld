package limbic

import (
	"slices"
	"sort"
	"strings"
	"sync"
)

// Default Beta prior: α=β=1 ⇒ uniform ⇒ trust 0 (neutral) for a never-met party.
const (
	defaultPriorA = 1.0
	defaultPriorB = 1.0
)

// TrustGrade is a coarse bucket over the [-1,1] trust scale, for quick branching.
type TrustGrade uint8

const (
	Hostile  TrustGrade = iota // trust < -0.4
	Wary                       // -0.4 .. -0.1
	Neutral                    // -0.1 .. 0.1
	Friendly                   // 0.1 .. 0.4
	Trusted                    // >= 0.4
)

func (g TrustGrade) String() string {
	switch g {
	case Hostile:
		return "hostile"
	case Wary:
		return "wary"
	case Neutral:
		return "neutral"
	case Friendly:
		return "friendly"
	case Trusted:
		return "trusted"
	default:
		return "unknown"
	}
}

func gradeOf(trust float64) TrustGrade {
	switch {
	case trust < -0.4:
		return Hostile
	case trust < -0.1:
		return Wary
	case trust < 0.1:
		return Neutral
	case trust < 0.4:
		return Friendly
	default:
		return Trusted
	}
}

// Rel is a read-only view of the host's relationship with one counterparty.
type Rel struct {
	Name     string
	Trust    float64 // Beta mean mapped to [-1,1]
	Grade    TrustGrade
	Familiar int // encounter count
	Tags     []string
}

// Entry is the stored Beta state. Exported fields so it JSON-round-trips for
// persistence via the memory layer.
type Entry struct {
	Name       string   `json:"name"`
	Alpha      float64  `json:"alpha"`
	Beta       float64  `json:"beta"`
	Encounters int      `json:"encounters"`
	Tags       []string `json:"tags,omitempty"`
}

// Ledger is the host's Beta(α,β) trust ledger. Each counterparty accrues
// positive (α) and negative (β) evidence; trust is the Beta mean mapped to
// [-1,1]. Keyed by normalized (lower-cased) name. Safe for concurrent use.
type Ledger struct {
	mu             sync.Mutex
	rows           map[string]*Entry
	priorA, priorB float64
}

// NewLedger returns an empty ledger with the default neutral prior.
func NewLedger() *Ledger {
	return &Ledger{rows: map[string]*Entry{}, priorA: defaultPriorA, priorB: defaultPriorB}
}

func (l *Ledger) rowLocked(name string) *Entry {
	key := normalizeName(name)
	e, ok := l.rows[key]
	if !ok {
		e = &Entry{Name: name, Alpha: l.priorA, Beta: l.priorB}
		l.rows[key] = e
	}
	return e
}

// Observe records an interaction outcome: good adds α (positive evidence), else
// β (negative). weight scales the evidence (default to 1 with weight<=0).
func (l *Ledger) Observe(name string, good bool, weight float64) {
	if weight <= 0 {
		weight = 1
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	if good {
		e.Alpha += weight
	} else {
		e.Beta += weight
	}
}

// Met records an encounter (familiarity), without moving trust.
func (l *Ledger) Met(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rowLocked(name).Encounters++
}

// Tag attaches a label (e.g. "scammer", "trusted_partner") if not already present.
func (l *Ledger) Tag(name, tag string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	if !slices.Contains(e.Tags, tag) {
		e.Tags = append(e.Tags, tag)
	}
}

// Rel returns the relationship view for name. A never-met party yields a
// neutral row (trust 0, familiar 0).
func (l *Ledger) Rel(name string) Rel {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := normalizeName(name)
	if e, ok := l.rows[key]; ok {
		return e.view()
	}
	return Rel{Name: name, Trust: 0, Grade: Neutral, Familiar: 0}
}

// Known reports whether the ledger has a row for name.
func (l *Ledger) Known(name string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.rows[normalizeName(name)]
	return ok
}

// All returns every relationship view, sorted by descending familiarity.
func (l *Ledger) All() []Rel {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Rel, 0, len(l.rows))
	for _, e := range l.rows {
		out = append(out, e.view())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Familiar > out[j].Familiar })
	return out
}

func (e *Entry) view() Rel {
	t := trustFromBeta(e.Alpha, e.Beta)
	return Rel{Name: e.Name, Trust: t, Grade: gradeOf(t), Familiar: e.Encounters, Tags: append([]string(nil), e.Tags...)}
}

// trustFromBeta maps the Beta mean α/(α+β) ∈ [0,1] to a signed trust in [-1,1].
func trustFromBeta(a, b float64) float64 {
	if a+b <= 0 {
		return 0
	}
	return 2*(a/(a+b)) - 1
}

// --- persistence (snapshot for the memory layer) ---------------------------

// Export returns a JSON-serializable snapshot of every row.
func (l *Ledger) Export() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, 0, len(l.rows))
	for _, e := range l.rows {
		cp := *e
		cp.Tags = append([]string(nil), e.Tags...)
		out = append(out, cp)
	}
	return out
}

// Import loads rows from a snapshot, replacing any existing state.
func (l *Ledger) Import(rows []Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rows = make(map[string]*Entry, len(rows))
	for i := range rows {
		e := rows[i]
		l.rows[normalizeName(e.Name)] = &e
	}
}

func normalizeName(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
