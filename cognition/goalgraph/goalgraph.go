// Package goalgraph is the host's INTENTION graph — the third structural leg of
// the mind beside the knowledge ledger (beliefs about the world) and the trust
// ledger (people). It records the host's goals/sub-goals/open-goals and open
// questions as NODES, and how its tasks relate as typed EDGES (requires /
// produces / enables / blocked_by / serves) — so actions become purposeful
// ("keep the ore" follows from ore --serves--> smithing) and failures become
// generative ("died to the bear" → mine --blocked_by--> bear → spawn
// train-combat --enables--> survive-at-mine).
//
// It is deliberately a LIGHTWEIGHT, ACCRETING MEMORY GRAPH the host reads and
// traverses, and the distillation crons grow and prune — NOT a real-time
// planner the host must solve every tick. Memory, not search. Persisted under
// the "goalgraph:" memory namespace (runtime/goalgraph.go), same spine as the
// knowledge + trust ledgers. Safe for concurrent use.
//
// This is the DATA layer of docs/world-knowledge-and-learning.md §3.7. Phase 1
// builds the structure + persistence + read-only traversal; the writers (failure
// → enabling sub-goal, recipes → edges) and the goal-graph-aware director land
// in later phases.
package goalgraph

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// NodeKind classifies a node in the intention graph.
const (
	KindAspiration   = "aspiration"    // a month-horizon ambition ("master a craft") — a portfolio root day-scale goals serve
	KindGoal         = "goal"          // a top-level objective
	KindSubgoal      = "subgoal"       // a step toward a goal
	KindOpenGoal     = "open_goal"     // a wanted-but-unstarted goal
	KindOpenQuestion = "open_question" // an unresolved unknown ("where do pickaxes sell?")
	KindState        = "state"         // a resource/condition ("have a pickaxe", "smithing L20")
)

// EdgeRel is a typed dependency between two nodes.
const (
	RelRequires  = "requires"   // From requires To   (smith requires bars requires ore)
	RelProduces  = "produces"   // From produces To   (mining produces ore)
	RelEnables   = "enables"    // From enables To     (combat enables survive-at-mine)
	RelBlockedBy = "blocked_by" // From blocked_by To  (mine blocked_by bear)
	RelServes    = "serves"     // From serves To      (ore serves smithing)
)

// Status of a node's pursuit.
const (
	StatusOpen      = "open"
	StatusActive    = "active"
	StatusBlocked   = "blocked"
	StatusDone      = "done"
	StatusAbandoned = "abandoned"
)

// DefaultGraphCap bounds the node count the host keeps in RAM (and re-serializes
// to bbolt + mirrors to mesa every flush). Like memory.DefaultJournalCap (64),
// the host structure stays LIGHT — the distillation crons on mesa are what grow
// the durable graph; the host's in-RAM copy is a bounded working set. See
// PruneTerminal.
const DefaultGraphCap = 64

// Node is one goal / sub-goal / open-question / state in the graph.
type Node struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Label    string   `json:"label"`
	Status   string   `json:"status"`
	Progress float64  `json:"progress"` // 0..1
	Tags     []string `json:"tags,omitempty"`
	At       int64    `json:"at"`
}

// Edge is a typed directed dependency between two node IDs.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rel  string `json:"rel"`
}

// Snapshot is the JSON-serialisable persistence form.
type Snapshot struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Graph is the per-host directed intention graph.
type Graph struct {
	mu    sync.Mutex
	nodes map[string]*Node
	edges []Edge
	now   func() int64
}

// New returns an empty intention graph.
func New() *Graph {
	return &Graph{nodes: map[string]*Node{}, now: func() int64 { return time.Now().Unix() }}
}

func norm(id string) string { return strings.ToLower(strings.TrimSpace(id)) }

func (g *Graph) nodeLocked(id, kind, label string) *Node {
	key := norm(id)
	n, ok := g.nodes[key]
	if !ok {
		n = &Node{ID: strings.TrimSpace(id), Kind: kind, Label: label, Status: StatusOpen, At: g.now()}
		if n.Kind == "" {
			n.Kind = KindState
		}
		g.nodes[key] = n
	}
	return n
}

// kindRank orders node kinds by COMMITMENT so an Upsert collision (two distinct
// ids that normalize to the same key — e.g. adopt-goals keyed on raw LLM
// free-text, "Mine ore" vs "Mine Ore") can never DEMOTE a node's kind. A higher
// rank is a stronger commitment: a top-level goal outranks a sub-goal, which
// outranks a wanted-but-unstarted open_goal, which outranks an open_question,
// which outranks a bare state node. In particular a re-adopt of a goal as a
// subgoal must not demote it (goal > subgoal). An aspiration is the strongest
// commitment of all (the month-horizon portfolio root day-scale goals serve) —
// no later upsert/merge may demote it. Unknown kinds rank 0.
func kindRank(kind string) int {
	switch kind {
	case KindAspiration:
		return 5
	case KindGoal:
		return 4
	case KindSubgoal:
		return 3
	case KindOpenGoal:
		return 2
	case KindOpenQuestion:
		return 1
	default: // KindState and unknown
		return 0
	}
}

// Upsert adds or updates a node. A new node defaults to StatusOpen; an existing
// node keeps its status/kind/label unless a non-empty replacement is given.
// On a normalized-id collision the Kind is only ADOPTED when it is at least as
// specific as the existing kind (see kindRank) — a re-adopt can never silently
// demote a goal to a subgoal/open_goal/state. Label/Status still last-writer-win.
func (g *Graph) Upsert(id, kind, label, status string) Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.nodeLocked(id, kind, label)
	if kind != "" && kindRank(kind) >= kindRank(n.Kind) {
		n.Kind = kind
	}
	if label != "" {
		n.Label = label
	}
	if status != "" {
		n.Status = status
	}
	n.At = g.now()
	return cloneNode(n)
}

// Link adds a typed edge (deduped). Missing endpoints are auto-created as state
// nodes so callers can link toward a not-yet-explicit node. Endpoints are
// resolved through nodeLocked and the edge stores each node's CANONICAL ID (the
// first creator's id), so Edge.From/To can never drift from the matching Node.ID
// even if a later caller links with different case/spacing — keeping a future
// SQL/LTM string-JOIN of edges against nodes from orphaning edges.
func (g *Graph) Link(from, to, rel string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fromID := g.nodeLocked(from, "", "").ID
	toID := g.nodeLocked(to, "", "").ID
	nf, nt := norm(fromID), norm(toID)
	for _, e := range g.edges {
		if norm(e.From) == nf && norm(e.To) == nt && e.Rel == rel {
			return // already linked
		}
	}
	g.edges = append(g.edges, Edge{From: fromID, To: toID, Rel: rel})
}

// SetStatus / SetProgress update a node (creating it if absent).
func (g *Graph) SetStatus(id, status string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.nodeLocked(id, "", "")
	n.Status = status
	n.At = g.now()
}

func (g *Graph) SetProgress(id string, p float64) {
	if p < 0 {
		p = 0
	} else if p > 1 {
		p = 1
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.nodeLocked(id, "", "")
	n.Progress = p
	n.At = g.now()
}

// Tag attaches a label to a node if not already present.
func (g *Graph) Tag(id, tag string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.nodeLocked(id, "", "")
	if !slices.Contains(n.Tags, tag) {
		n.Tags = append(n.Tags, tag)
	}
}

// Untag removes a label from a node if present (the inverse of Tag). It is a
// no-op when the node or the tag is absent — it never creates a node. The
// director uses this to clear a transient "spinning" tag once a goal recovers,
// so the marker (and its planner-facing prose) does not persist for the host's
// whole life after the spin has passed.
func (g *Graph) Untag(id, tag string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	n, ok := g.nodes[norm(id)]
	if !ok {
		return
	}
	if i := slices.Index(n.Tags, tag); i >= 0 {
		n.Tags = slices.Delete(n.Tags, i, i+1)
	}
}

// HasTag reports whether node id carries tag. Lock-guarded read; returns false for
// an absent node (it never creates one, unlike Tag's nodeLocked). Lets callers test
// a single tag without cloning the whole node.
func (g *Graph) HasTag(id, tag string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	n, ok := g.nodes[norm(id)]
	return ok && slices.Contains(n.Tags, tag)
}

// TagsWithPrefix returns the tags on node id that start with prefix (a copy; never
// the live slice). Lock-guarded; returns nil for an absent node. The prefix-aware
// companion to HasTag — callers count/enumerate a tag FAMILY (e.g.
// "source-spent:place:") without cloning the whole node. Never creates a node.
func (g *Graph) TagsWithPrefix(id, prefix string) []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	n, ok := g.nodes[norm(id)]
	if !ok {
		return nil
	}
	var out []string
	for _, t := range n.Tags {
		if strings.HasPrefix(t, prefix) {
			out = append(out, t)
		}
	}
	return out
}

// Get returns a node by id.
func (g *Graph) Get(id string) (Node, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if n, ok := g.nodes[norm(id)]; ok {
		return cloneNode(n), true
	}
	return Node{}, false
}

// Has reports whether a node exists.
func (g *Graph) Has(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.nodes[norm(id)]
	return ok
}

// Out returns edges originating at id (optionally filtered by rel; "" = all).
func (g *Graph) Out(id, rel string) []Edge {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.adjacent(id, rel, true)
}

// In returns edges pointing at id (optionally filtered by rel; "" = all).
func (g *Graph) In(id, rel string) []Edge {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.adjacent(id, rel, false)
}

func (g *Graph) adjacent(id, rel string, outgoing bool) []Edge {
	key := norm(id)
	var out []Edge
	for _, e := range g.edges {
		end := e.From
		if !outgoing {
			end = e.To
		}
		if norm(end) == key && (rel == "" || e.Rel == rel) {
			out = append(out, e)
		}
	}
	return out
}

// Blockers returns the nodes blocking id (via id --blocked_by--> X).
func (g *Graph) Blockers(id string) []Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []Node
	for _, e := range g.adjacent(id, RelBlockedBy, true) {
		if n, ok := g.nodes[norm(e.To)]; ok {
			out = append(out, cloneNode(n))
		}
	}
	return out
}

// OpenQuestions returns the unresolved open-question nodes (not done/abandoned).
func (g *Graph) OpenQuestions() []Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []Node
	for _, n := range g.nodes {
		if n.Kind == KindOpenQuestion && n.Status != StatusDone && n.Status != StatusAbandoned {
			out = append(out, cloneNode(n))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At > out[j].At })
	return out
}

// OpenGoals returns the unstarted goal nodes (KindOpenGoal, not done/abandoned),
// newest-first — the successor candidates for advancement when the active goal
// closes. Pure, lock-guarded read (memory, not a solver).
func (g *Graph) OpenGoals() []Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []Node
	for _, n := range g.nodes {
		if n.Kind == KindOpenGoal && n.Status != StatusDone && n.Status != StatusAbandoned {
			out = append(out, cloneNode(n))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At > out[j].At })
	return out
}

// Aspirations returns the live month-horizon aspiration nodes (KindAspiration,
// not done/abandoned — an aspiration never closes in the day-scale sense, but an
// operator/cron may retire one). Ordered oldest-first (then by ID), i.e. SEED
// order: genesis emits aspirations in priority order in one burst, so creation
// time is the authored ranking. Pure, lock-guarded read.
func (g *Graph) Aspirations() []Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []Node
	for _, n := range g.nodes {
		if n.Kind == KindAspiration && n.Status != StatusDone && n.Status != StatusAbandoned {
			out = append(out, cloneNode(n))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].At != out[j].At {
			return out[i].At < out[j].At // oldest (seed order) first
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// ServesRollup is the rolled-up progress of an aspiration: how the day-scale
// goals serving it (via X --serves--> aspiration) are faring, plus recency.
// Deliberately simple — count + recency, not a planner metric.
type ServesRollup struct {
	Done    int // serving goals completed
	Working int // serving goals being pursued (active or blocked)
	Open    int // serving goals queued but unstarted
	// LastTouched is the unix time of the latest WORK on the aspiration: the
	// max At over the aspiration node itself and every serving node that has
	// actually been worked (active/blocked/done). A merely-QUEUED open_goal does
	// not touch it — adopting a goal is not progress, working it is. Zero when
	// the aspiration node is absent.
	LastTouched int64
}

// Rollup computes the ServesRollup for aspiration (or any node) id: a walk of
// the incoming serves-edges counting the goal-kind nodes by status. Only
// goal/subgoal/open_goal nodes count — a state node serving an aspiration
// ("ore serves smithing") is a dependency, not pursued work. Pure, lock-guarded.
func (g *Graph) Rollup(id string) ServesRollup {
	g.mu.Lock()
	defer g.mu.Unlock()
	var r ServesRollup
	if n, ok := g.nodes[norm(id)]; ok {
		r.LastTouched = n.At
	}
	for _, e := range g.adjacent(id, RelServes, false) { // X --serves--> id
		n, ok := g.nodes[norm(e.From)]
		if !ok {
			continue
		}
		switch n.Kind {
		case KindGoal, KindSubgoal, KindOpenGoal:
		default:
			continue // states/questions serving an aspiration are not pursued work
		}
		switch n.Status {
		case StatusDone:
			r.Done++
		case StatusActive, StatusBlocked:
			r.Working++
		case StatusOpen:
			r.Open++
		}
		// Recency: worked nodes only (see LastTouched doc).
		if n.Status != StatusOpen && n.Status != StatusAbandoned && n.At > r.LastTouched {
			r.LastTouched = n.At
		}
	}
	return r
}

// Nodes returns all nodes (sorted by id for determinism).
func (g *Graph) Nodes() []Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, cloneNode(n))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Edges returns all edges.
func (g *Graph) Edges() []Edge {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]Edge(nil), g.edges...)
}

// cloneNode returns a value copy of n with its slice fields deep-copied, so a
// caller that mutates the returned node (e.g. appending to Tags) cannot corrupt
// the stored node or race a concurrent reader sharing the same backing array.
func cloneNode(n *Node) Node {
	cp := *n
	cp.Tags = append([]string(nil), n.Tags...)
	return cp
}

// --- persistence -----------------------------------------------------------

// Export returns a snapshot of the graph.
func (g *Graph) Export() Snapshot {
	g.mu.Lock()
	defer g.mu.Unlock()
	nodes := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, cloneNode(n))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return Snapshot{Nodes: nodes, Edges: append([]Edge(nil), g.edges...)}
}

// Import replaces the graph from a snapshot. Each node's Tags slice is
// deep-copied (matching cloneNode) so the imported graph never aliases the
// caller's snapshot backing arrays — mutating snap.Nodes[i].Tags afterward, or
// importing the same snapshot into two graphs, cannot corrupt a stored node.
func (g *Graph) Import(s Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes = make(map[string]*Node, len(s.Nodes))
	for i := range s.Nodes {
		n := s.Nodes[i]
		n.Tags = append([]string(nil), n.Tags...)
		g.nodes[norm(n.ID)] = &n
	}
	g.edges = append([]Edge(nil), s.Edges...)
}

// ImportMerge folds a snapshot into the live graph NON-DESTRUCTIVELY (H17 host-
// side). Unlike Import (which REPLACES wholesale and is the cold-start path), this
// is the WARM-host re-import: it lets a running host pull the now server-reconciled
// graph (the insight cron's cross-entity chains + open-question closures) into its
// local copy so those nodes/edges SURVIVE the host's next wholesale flush — without
// clobbering the live working set.
//
// Authority is PARTITIONED: the host owns goal/subgoal/status/progress; the cron
// owns chain/closure nodes + edges. So for a node present in BOTH, the host KEEPS
// its own Status/Progress/Label (the live truth) and only ADOPTS cron-added Tags
// (union) and a stronger Kind (kindRank, never a demote) — it never adopts the
// snapshot's status (which would re-open a goal the host just closed). A node the
// host doesn't know is ADDED as-is (deep-copied Tags). A host node ABSENT from the
// snapshot is NEVER deleted (the cron doesn't own it). Edges are added if absent
// (same dedup as Link). Bounded after the merge by PruneTerminal(DefaultGraphCap)
// so a re-import can never push the host past its cap.
func (g *Graph) ImportMerge(s Snapshot) {
	g.mu.Lock()
	for i := range s.Nodes {
		sn := s.Nodes[i]
		key := norm(sn.ID)
		if existing, ok := g.nodes[key]; ok {
			// Present in both: keep the host's Status/Progress/Label (live truth);
			// adopt only the cron-added Tags (union) + a stronger Kind (no demote).
			if sn.Kind != "" && kindRank(sn.Kind) > kindRank(existing.Kind) {
				existing.Kind = sn.Kind
			}
			for _, t := range sn.Tags {
				if !slices.Contains(existing.Tags, t) {
					existing.Tags = append(existing.Tags, t)
				}
			}
			continue
		}
		// New (cron-added) node: take it as-is with a deep-copied Tags slice.
		n := sn
		n.Tags = append([]string(nil), sn.Tags...)
		if n.ID == "" {
			n.ID = strings.TrimSpace(sn.ID)
		}
		g.nodes[key] = &n
	}
	// Add cron edges that don't already exist (same (from,to,rel) dedup as Link),
	// resolving endpoints to their canonical stored IDs so an edge can't orphan.
	for _, se := range s.Edges {
		fromID := g.nodeLocked(se.From, "", "").ID
		toID := g.nodeLocked(se.To, "", "").ID
		nf, nt := norm(fromID), norm(toID)
		dup := false
		for _, e := range g.edges {
			if norm(e.From) == nf && norm(e.To) == nt && e.Rel == se.Rel {
				dup = true
				break
			}
		}
		if !dup {
			g.edges = append(g.edges, Edge{From: fromID, To: toID, Rel: se.Rel})
		}
	}
	g.mu.Unlock()
	// Re-bound after the additive merge (PruneTerminal takes its own lock).
	g.PruneTerminal(DefaultGraphCap)
}

// Empty reports whether the graph has no nodes (used to gate persistence).
func (g *Graph) Empty() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.nodes) == 0
}

// PruneTerminal bounds the graph to at most max nodes by dropping the OLDEST
// TERMINAL nodes (StatusDone / StatusAbandoned, ordered by their At timestamp)
// first, then removing any edges left dangling by the eviction. Non-terminal
// nodes (open/active/blocked goals, sub-goals, open questions, state) are NEVER
// dropped — the live working set is preserved even if it alone exceeds max.
// max <= 0 uses DefaultGraphCap. It returns the number of nodes evicted.
//
// This is the host-LIGHT backstop: the graph otherwise grows monotonically
// (status flips to done/abandoned leave nodes resident forever), violating the
// "grow AND prune" package invariant and re-serializing an ever-larger blob to
// bbolt + mesa every flush. Deterministic (ties on At broken by node ID) and
// lock-guarded; intended to be called from the host's flush tick alongside the
// other GCs. O(N log N) in the terminal-node count.
func (g *Graph) PruneTerminal(max int) int {
	if max <= 0 {
		max = DefaultGraphCap
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	over := len(g.nodes) - max
	if over <= 0 {
		return 0
	}
	// Collect the terminal (evictable) nodes, oldest-first.
	type term struct {
		key string
		at  int64
		id  string
	}
	var terminals []term
	for key, n := range g.nodes {
		if n.Status == StatusDone || n.Status == StatusAbandoned {
			terminals = append(terminals, term{key: key, at: n.At, id: n.ID})
		}
	}
	sort.Slice(terminals, func(i, j int) bool {
		if terminals[i].at != terminals[j].at {
			return terminals[i].at < terminals[j].at // oldest first
		}
		return terminals[i].id < terminals[j].id // deterministic tie-break
	})
	if over > len(terminals) {
		over = len(terminals) // never touch non-terminal nodes
	}
	if over <= 0 {
		return 0
	}
	dropped := make(map[string]bool, over)
	for i := 0; i < over; i++ {
		delete(g.nodes, terminals[i].key)
		dropped[terminals[i].key] = true
	}
	// Drop edges now orphaned by an evicted endpoint.
	kept := g.edges[:0]
	for _, e := range g.edges {
		if dropped[norm(e.From)] || dropped[norm(e.To)] {
			continue
		}
		kept = append(kept, e)
	}
	g.edges = kept
	return over
}
