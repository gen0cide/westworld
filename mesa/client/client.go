package mesaclient

import (
	"context"
	"errors"
)

// ErrOffline is returned by mesa calls when no mesa is reachable. Callers treat
// it as "fall back to local behavior", never as a hard failure.
var ErrOffline = errors.New("mesa: offline")

// Client is the host's game-native gateway to mesa. The currency is DSL + game
// state: the host ships game SITUATIONS up and receives game MOVES (DSL) down.
//
//   - Act      — the core call: "what do I do now?" → a Move (run/write/act/idle).
//   - Decide   — the narrow in-routine choice: pick one author-supplied option.
//   - Recall   — game knowledge (procedural / entity / episodic / social).
//   - Remember — mirror a game episode + social deltas up (async).
//   - Subscribe — Provision push: compiled routine library + policy + goals.
//
// Each isolated host owns its own Client + connection and authenticates as its
// own host_id. Implementations must be concurrency-safe and honor ctx deadlines.
type Client interface {
	Act(ctx context.Context, s *Situation) (*Move, error)
	Decide(ctx context.Context, c *Choice) (*Decision, error)
	Recall(ctx context.Context, q *Query) (*Knowledge, error)
	Remember(ctx context.Context, e *Episode) error
	Subscribe(ctx context.Context, hostID string) (<-chan Directive, error)
	Healthy() bool
}

// --- Act: the core "what do I do now?" call ---------------------------------

// Situation is the game state the host ships up to ask for its next move. It is
// what the host already knows locally — mesa adds RAG/memory/persona itself.
type Situation struct {
	HostID  string
	Goal    string            // the host's current goal (may be empty)
	Trigger string            // why we're asking: "routine_exhausted" | "novel" | "chat" | "low_hp" | ...
	World   WorldSnapshot     // the live game state the host can see
	Recent  []string          // recent salient events (chat lines, combat, level-ups, ...)
	Affect  Affect            // current mood (modulates model tier/temperature)
	Hints   map[string]string // extra local context (counterparty, last outcome, ...)
}

// WorldSnapshot is a compact view of the host's current game state.
type WorldSnapshot struct {
	X, Y          int
	Region        string // gazetteer place name, if known
	HPCur, HPMax  int
	CombatLevel   int
	Fatigue       float64
	InvFree       int
	InvSummary    []string // e.g. ["bronze pickaxe", "5 lobster"]
	NearbyNpcs    []string
	NearbyPlayers []string
}

// MoveKind tags the game move mesa chose.
type MoveKind uint8

const (
	MoveRunRoutine   MoveKind = iota // run an existing library routine
	MoveWriteRoutine                 // author a fresh DSL routine and run it
	MoveDirectAction                 // perform a single game action
	MoveIdle                         // do nothing for a bit
)

func (k MoveKind) String() string {
	switch k {
	case MoveRunRoutine:
		return "run_routine"
	case MoveWriteRoutine:
		return "write_routine"
	case MoveDirectAction:
		return "direct_action"
	case MoveIdle:
		return "idle"
	default:
		return "unknown"
	}
}

// Move is mesa's game-native answer to Act — a DSL/game directive, not a choice
// string. It is a tagged union: Kind selects which fields are meaningful.
type Move struct {
	Kind      MoveKind
	Reasoning string

	// MoveRunRoutine: a routine already in the host's library.
	RoutinePath string

	// MoveWriteRoutine: freshly authored DSL the host parses, validates, runs
	// under the pearl Gate, and logs. Quarantined ⇒ run once/gated, do NOT
	// promote to the durable library (promotion is a separate reviewed step).
	RoutineName string
	DSLSource   string
	Quarantined bool

	// MoveRunRoutine | MoveWriteRoutine: positional args to the routine.
	Args []string

	// MoveDirectAction: one game verb (a DSL action name) + its args.
	Verb       string
	ActionArgs []string

	// MoveIdle.
	IdleSeconds int
}

// --- Decide: the narrow in-routine option choice ----------------------------

// Choice is an in-routine decision: pick one of the options the running routine
// supplied (≙ brain.Situation). Pearl absorbs the repeats; only misses reach here.
type Choice struct {
	HostID   string
	Question string
	Options  []string
	Affect   Affect
}

// Decision is the chosen option (≙ brain.Decision) plus optional cache hints.
type Decision struct {
	Choice          string
	Reasoning       string
	Confidence      float64
	CacheKey        string // optional: fold into the host's local decision cache
	CacheTTLSeconds int64  // 0 = do not cache
}

// --- Recall: game knowledge -------------------------------------------------

// QueryKind narrows a recall to a kind of game knowledge.
type QueryKind uint8

const (
	KnowAny        QueryKind = iota
	KnowProcedural           // how-to (wiki/AutoRune) — may come back as a DSL snippet
	KnowEntity               // facts about an NPC / item / place
	KnowEpisodic             // the host's own past ("last time I fought X")
	KnowSocial               // relationship with a named player
)

// Query is a game-knowledge request.
type Query struct {
	HostID string
	Text   string
	Kind   QueryKind
	TopK   int
}

// Knowledge is the recall result.
type Knowledge struct {
	Items []KnowledgeItem
}

// KnowledgeItem is one recalled piece. For KnowProcedural, DSL may carry a
// runnable routine/snippet; otherwise Text holds the knowledge.
type KnowledgeItem struct {
	Kind       QueryKind
	Text       string
	DSL        string // non-empty for procedural how-to that compiles to a routine
	Provenance string
	Score      float64
}

// --- Remember: game episodes + social ---------------------------------------

// Episode is a game event worth remembering, mirrored up async.
type Episode struct {
	HostID         string
	IdempotencyKey string // host_id + event-hash + occurred_at (UNIQUE; at-least-once)
	Kind           string // "kill" | "death" | "trade" | "scam" | "quest_step" | "discovery" | "social"
	Text           string
	Importance     float64 // cheap local hint; mesa re-scores
	OccurredAtUnix int64
	Relationship   *RelationDelta // optional social/trust delta
	Tags           map[string]string
}

// RelationDelta is a lossless trust-ledger increment (≙ limbic.Entry delta).
type RelationDelta struct {
	Name             string
	DAlpha           float64
	DBeta            float64
	DEncounters      int
	TotalValueTraded float64
	AddTags          []string
}

// Affect is the wire form of the host's mood vector.
type Affect struct {
	Stress     float64
	Confidence float64
	Valence    float64
}

// --- Provision: compiled game behavior pushed down --------------------------

// DirectiveKind enumerates what mesa pushes down — all of it compiled game
// behavior: routines, the selection policy, goals, decay.
type DirectiveKind string

const (
	DirectiveRoutineUpsert     DirectiveKind = "routine_upsert"     // a DSL routine for the library
	DirectivePearlRefresh      DirectiveKind = "pearl_refresh"      // compiled selection policy (pearl.Table)
	DirectivePersonaRevision   DirectiveKind = "persona_revision"   // compiled Cornerstone
	DirectiveGoalRevision      DirectiveKind = "goal_revision"      //
	DirectiveTrustDecay        DirectiveKind = "trust_decay"        // cron output
	DirectiveReverieRebaseline DirectiveKind = "reverie_rebaseline" // cron output
)

// Directive is one mesa→host push. ID is monotonic per host; the host applies
// strictly increasing IDs and ignores stale ones (idempotent, no CRDT).
type Directive struct {
	HostID  string
	ID      int64
	Kind    DirectiveKind
	Payload []byte
}

// --- StubClient: the offline default ----------------------------------------

// StubClient is an always-offline Client. Act/Decide error with ErrOffline,
// Recall returns empty, Remember is a no-op, Subscribe yields a closed channel.
// It lets a host run fully self-contained on its local routines + pearl until a
// real mesa is wired.
type StubClient struct{}

func (StubClient) Act(context.Context, *Situation) (*Move, error)     { return nil, ErrOffline }
func (StubClient) Decide(context.Context, *Choice) (*Decision, error) { return nil, ErrOffline }
func (StubClient) Recall(context.Context, *Query) (*Knowledge, error) { return &Knowledge{}, nil }
func (StubClient) Remember(context.Context, *Episode) error           { return nil }
func (StubClient) Subscribe(context.Context, string) (<-chan Directive, error) {
	ch := make(chan Directive)
	close(ch)
	return ch, nil
}
func (StubClient) Healthy() bool { return false }
