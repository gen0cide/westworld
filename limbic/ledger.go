package limbic

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
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
// Relationships are multi-axis (§3.4): TRUST (honesty in deals), AFFINITY
// (warmth — sparring, friendly chat), and GRIEVANCE (wrongs — ganked, scammed)
// are INDEPENDENT. "I enjoy dueling you / I don't trust you / you annoy me" are
// not one number. Familiarity is how well-known the party is. Grade is a coarse
// blend of all three for quick branching.
type Rel struct {
	Name      string
	Trust     float64 // Beta mean mapped to [-1,1]
	Affinity  float64 // warmth, squashed to [-1,1]
	Grievance float64 // accumulated harm, squashed to [0,1] (monotone)
	Grade     TrustGrade
	Familiar  int       // encounter count
	LastAt    time.Time // last evidence-bearing interaction; zero = unknown (row predates stamping)
	Tags      []string
}

// Entry is the stored relationship state. Exported fields so it JSON-round-trips
// for persistence via the memory layer. The axis backing is deliberately mixed:
// TRUST is a Beta(α,β) hit-rate (will this trade be honest?), while AFFINITY and
// GRIEVANCE are signed/monotone sentiment ACCUMULATORS (drop sentiment / accrue
// grievance) — squashed on read. The two *Sum fields are additive (omitempty):
// an old alpha/beta-only row unmarshals them as 0 = neutral, so persistence stays
// backward-compatible with no migration.
type Entry struct {
	Name         string   `json:"name"`
	Alpha        float64  `json:"alpha"`
	Beta         float64  `json:"beta"`
	Encounters   int      `json:"encounters"`
	Tags         []string `json:"tags,omitempty"`
	AffinitySum  float64  `json:"affinity_sum,omitempty"`  // signed warmth accumulator
	GrievanceSum float64  `json:"grievance_sum,omitempty"` // monotone harm accumulator (>=0)
	ValueTraded  float64  `json:"value_traded,omitempty"`  // monotone total goods exchanged (>=0)
	LastAt       int64    `json:"last_at,omitempty"`       // unix secs of the last evidence-bearing mutation; 0 = pre-stamping row
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
	e.LastAt = nowUnix()
}

// ObserveAffinity moves the AFFINITY (warmth) axis. weight is signed: positive
// for warm signals (a friendly spar, a smooth trade), negative for cold ones
// (friction). Independent of TRUST and GRIEVANCE.
func (l *Ledger) ObserveAffinity(name string, weight float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	e.AffinitySum += weight
	e.LastAt = nowUnix()
}

// ObserveGrievance accrues GRIEVANCE (a wrong: ganked, scammed, repeated item
// loss). weight must be > 0 (grievance is monotone — it only grows in the
// deterministic spine; forgiveness/decay is the Phase-4 LLM overlay). The sum is
// clamped to >= 0 defensively.
func (l *Ledger) ObserveGrievance(name string, weight float64) {
	if weight <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	e.GrievanceSum += weight
	if e.GrievanceSum < 0 {
		e.GrievanceSum = 0
	}
	e.LastAt = nowUnix()
}

// ObserveValueTraded accrues the total goods VALUE exchanged with a party over a
// completed trade. weight must be > 0 (the magnitude of a settled exchange, e.g.
// the quantity of items that changed hands); it is monotone — a relationship's
// trade volume only grows. Independent of TRUST/AFFINITY/GRIEVANCE. The sum is
// clamped to >= 0 defensively. This volume mirrors up as the relationship's
// value_traded so mesa can weight a long-standing trade partner.
func (l *Ledger) ObserveValueTraded(name string, weight float64) {
	if weight <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	e.ValueTraded += weight
	if e.ValueTraded < 0 {
		e.ValueTraded = 0
	}
	e.LastAt = nowUnix()
}

// Met records an encounter (familiarity), without moving trust.
func (l *Ledger) Met(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.rowLocked(name)
	e.Encounters++
	e.LastAt = nowUnix()
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
	aff := squash(e.AffinitySum, affinityCap)
	gri := squash(e.GrievanceSum, grievanceCap) // GrievanceSum>=0 ⇒ result ∈ [0,1]
	r := Rel{
		Name: e.Name, Trust: t, Affinity: aff, Grievance: gri,
		Grade: gradeOfMultiAxis(t, aff, gri), Familiar: e.Encounters,
		Tags: append([]string(nil), e.Tags...),
	}
	if e.LastAt > 0 {
		r.LastAt = time.Unix(e.LastAt, 0)
	}
	return r
}

// nowUnix stamps Entry.LastAt on every evidence-bearing mutation (Observe*/Met
// — NOT Tag, which attaches a label without an interaction). Stamped in unix
// seconds, the same additive-omitempty persistence pattern as the axis sums.
func nowUnix() int64 { return time.Now().Unix() }

// trustFromBeta maps the Beta mean α/(α+β) ∈ [0,1] to a signed trust in [-1,1].
func trustFromBeta(a, b float64) float64 {
	if a+b <= 0 {
		return 0
	}
	return 2*(a/(a+b)) - 1
}

// affinityCap/grievanceCap set how many accumulated events saturate an axis: a
// handful of clear-signal events takes the axis to its bound. Tuning policy, not
// fact — retunable without migration because the raw sums (not the squashed
// reads) are what persist + travel on the wire.
const affinityCap, grievanceCap = 6.0, 6.0

// squash maps a signed accumulator into [-cap..cap]/cap, clamped to [-1,1].
func squash(sum, cap float64) float64 {
	switch {
	case sum > cap:
		return 1
	case sum < -cap:
		return -1
	default:
		return sum / cap
	}
}

// gradeOfMultiAxis folds the three axes into a single coarse grade. Trust leads;
// affinity nudges it up; a grievance drags it down hard. A standing grudge
// (grievance >= 0.5) caps the grade at Wary no matter how warm/trusted — you can
// re-enjoy sparring someone who once ganked you, but the relationship is never
// "friendly/trusted" while the grudge stands.
func gradeOfMultiAxis(trust, affinity, grievance float64) TrustGrade {
	adj := trust + 0.3*affinity - 0.8*grievance
	g := gradeOf(adj)
	if grievance >= 0.5 && g > Wary {
		g = Wary
	}
	return g
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
