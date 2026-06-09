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

// Upsert adds or updates a node. A new node defaults to StatusOpen; an existing
// node keeps its status/kind/label unless a non-empty replacement is given.
func (g *Graph) Upsert(id, kind, label, status string) Node {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.nodeLocked(id, kind, label)
	if kind != "" {
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
// nodes so callers can link toward a not-yet-explicit node.
func (g *Graph) Link(from, to, rel string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodeLocked(from, "", "")
	g.nodeLocked(to, "", "")
	nf, nt := norm(from), norm(to)
	for _, e := range g.edges {
		if norm(e.From) == nf && norm(e.To) == nt && e.Rel == rel {
			return // already linked
		}
	}
	g.edges = append(g.edges, Edge{From: strings.TrimSpace(from), To: strings.TrimSpace(to), Rel: rel})
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

// Import replaces the graph from a snapshot.
func (g *Graph) Import(s Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes = make(map[string]*Node, len(s.Nodes))
	for i := range s.Nodes {
		n := s.Nodes[i]
		g.nodes[norm(n.ID)] = &n
	}
	g.edges = append([]Edge(nil), s.Edges...)
}

// Empty reports whether the graph has no nodes (used to gate persistence).
func (g *Graph) Empty() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.nodes) == 0
}
