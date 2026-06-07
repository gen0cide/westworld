package memory

import (
	"sort"
	"sync"
	"time"
)

// Episode is one durable thing that happened to the host — a level-up, a kill,
// a death, an objective milestone. It is the host's EPISODIC memory (the "what
// happened" of docs/cognition-and-autonomy.md §5), distinct from the per-turn
// transcript ring (which is ephemeral and resets per process). Episodes survive
// across restarts and re-genesis because the Journal is persisted through the
// tiered memory.Manager, and they are the history a session-genesis compile
// reasons over.
type Episode struct {
	Seq        int64   `json:"seq"`              // monotonic id within a journal
	Kind       string  `json:"kind"`             // milestone | kill | death | objective | discovery | social
	Text       string  `json:"text"`             // human-readable, already colour-stripped
	Importance float64 `json:"importance"`       // 0..1 — how salient (deaths high, routine kills low)
	At         int64   `json:"at"`               // unix seconds
	Entity     string  `json:"entity,omitempty"` // optional player/npc the episode is about
}

// JournalSnapshot is the serializable form of a Journal — what gets written to
// and read from the durable memory tier. It mirrors limbic.Ledger's Export/Import
// shape so the runtime persistence seam (loadJournal/flushJournal) is symmetric.
type JournalSnapshot struct {
	Objective   string    `json:"objective,omitempty"`
	ObjectiveAt int64     `json:"objective_at,omitempty"`
	Seq         int64     `json:"seq"`
	Episodes    []Episode `json:"episodes"`
}

// Journal is a bounded, importance-ranked episodic log plus the host's standing
// objective. It is a pure in-RAM data structure (like limbic.Ledger): it knows
// nothing about the Manager or the wire. The runtime owns persistence — it calls
// Export() to flush the snapshot through memory.Manager and Import() to restore
// it on start. Safe for concurrent use: a single capture goroutine appends while
// the director reads recall on its own goroutine.
type Journal struct {
	mu          sync.Mutex
	cap         int
	seq         int64
	objective   string
	objectiveAt int64
	episodes    []Episode
	now         func() time.Time
}

// DefaultJournalCap bounds the retained episode count. Recall surfaces a small
// salient subset; the cap keeps the persisted blob and memory footprint bounded
// while still spanning a long session of milestones.
const DefaultJournalCap = 64

// NewJournal builds an empty journal. cap <= 0 uses DefaultJournalCap.
func NewJournal(cap int) *Journal {
	if cap <= 0 {
		cap = DefaultJournalCap
	}
	return &Journal{cap: cap, now: time.Now}
}

// Append records an episode and returns it (with its assigned Seq). Importance
// is clamped to [0,1]. As a cheap de-dup, if the most recent episode has the
// same Kind+Text, its timestamp is refreshed instead of adding a duplicate —
// this collapses repeats like grinding the same low-value kill without losing
// the fact that it is still happening (it stays recency-fresh).
func (j *Journal) Append(kind, text string, importance float64, entity string) Episode {
	if importance < 0 {
		importance = 0
	} else if importance > 1 {
		importance = 1
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	at := j.now().Unix()
	if n := len(j.episodes); n > 0 {
		if last := &j.episodes[n-1]; last.Kind == kind && last.Text == text {
			last.At = at
			if importance > last.Importance {
				last.Importance = importance
			}
			return *last
		}
	}
	j.seq++
	ep := Episode{Seq: j.seq, Kind: kind, Text: text, Importance: importance, At: at, Entity: entity}
	j.episodes = append(j.episodes, ep)
	if len(j.episodes) > j.cap {
		j.episodes = j.episodes[len(j.episodes)-j.cap:]
	}
	return ep
}

// SetObjective records the host's standing goal so it survives a restart and is
// available to a future session-genesis compile. Returns true when the objective
// actually changed (so the caller can avoid churning the timestamp every turn).
func (j *Journal) SetObjective(goal string) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if goal == "" || goal == j.objective {
		return false
	}
	j.objective = goal
	j.objectiveAt = j.now().Unix()
	return true
}

// Objective returns the standing goal and the unix time it was set (0 if unset).
func (j *Journal) Objective() (string, int64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.objective, j.objectiveAt
}

// Recent returns up to n most-recent episodes in chronological order (oldest
// first), like the transcript ring but durable.
func (j *Journal) Recent(n int) []Episode {
	j.mu.Lock()
	defer j.mu.Unlock()
	if n <= 0 || len(j.episodes) == 0 {
		return nil
	}
	start := 0
	if len(j.episodes) > n {
		start = len(j.episodes) - n
	}
	return append([]Episode(nil), j.episodes[start:]...)
}

// Salient returns up to n episodes blending IMPORTANCE and RECENCY, in
// chronological order. A death from early in the session stays salient; so does
// whatever just happened. This is what feeds the per-turn Situation so the
// planner reasons over "what I've done / what mattered" instead of re-deriving
// the world each tick.
func (j *Journal) Salient(n int) []Episode {
	j.mu.Lock()
	defer j.mu.Unlock()
	if n <= 0 || len(j.episodes) == 0 {
		return nil
	}
	if len(j.episodes) <= n {
		return append([]Episode(nil), j.episodes...)
	}
	type scored struct {
		ep    Episode
		score float64
	}
	last := len(j.episodes) - 1
	ranked := make([]scored, len(j.episodes))
	for i, e := range j.episodes {
		recency := float64(i) / float64(last) // 0 (oldest) .. 1 (newest)
		ranked[i] = scored{ep: e, score: e.Importance + 0.6*recency}
	}
	sort.SliceStable(ranked, func(a, b int) bool { return ranked[a].score > ranked[b].score })
	picked := ranked[:n]
	out := make([]Episode, len(picked))
	for i, s := range picked {
		out[i] = s.ep
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Seq < out[b].Seq })
	return out
}

// Len reports the number of retained episodes.
func (j *Journal) Len() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.episodes)
}

// Export snapshots the journal for persistence (deep-copies the episode slice).
func (j *Journal) Export() JournalSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	return JournalSnapshot{
		Objective:   j.objective,
		ObjectiveAt: j.objectiveAt,
		Seq:         j.seq,
		Episodes:    append([]Episode(nil), j.episodes...),
	}
}

// Import restores a journal from a snapshot, replacing all state. The retained
// set is trimmed to cap (keeping the most recent), and seq is advanced past any
// imported episode so newly-appended ids never collide with restored ones.
func (j *Journal) Import(s JournalSnapshot) {
	j.mu.Lock()
	defer j.mu.Unlock()
	eps := append([]Episode(nil), s.Episodes...)
	if len(eps) > j.cap {
		eps = eps[len(eps)-j.cap:]
	}
	j.episodes = eps
	j.objective = s.Objective
	j.objectiveAt = s.ObjectiveAt
	j.seq = s.Seq
	for _, e := range eps {
		if e.Seq > j.seq {
			j.seq = e.Seq
		}
	}
}
