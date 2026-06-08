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
//   - Provision — pull this host's compiled persona/goals bundle (on connect).
//   - Act      — the core call: "what do I do now?" → a Move (run/write/act/idle).
//   - Decide   — the narrow in-routine choice: pick one author-supplied option.
//   - Recall   — game knowledge (procedural / entity / episodic / social).
//   - Remember — mirror a game episode + social deltas up (async).
//   - Subscribe — Provision push: compiled routine library + policy + goals.
//
// Each isolated host owns its own Client + connection and authenticates as its
// own host_id. Implementations must be concurrency-safe and honor ctx deadlines.
type Client interface {
	Provision(ctx context.Context, hostID string) (*Provisioning, error)
	// Genesis runs the heavy Opus-at-login session compile (history-aware goal +
	// mood baseline + keyword ladder). Called once at startup; the host applies it.
	Genesis(ctx context.Context, hostID, trigger, worldSummary string) (*GenesisResult, error)
	Act(ctx context.Context, s *Situation) (*Move, error)
	Decide(ctx context.Context, c *Choice) (*Decision, error)
	// Chat is the fast social reply path: a player's utterance in, a short
	// spoken reply out (speak=false ⇒ stay silent). Cheap + off the Act loop.
	Chat(ctx context.Context, hostID, from, message string, recent []string) (text string, speak bool, err error)
	// AnalysisInterpret classifies an operator-override directive (flat affect,
	// not in-persona) into a command/answer/hypothetical verdict, grounded in the
	// supplied flat host-state facts. Off the Act loop, cheap tier.
	AnalysisInterpret(ctx context.Context, directive string, state []string) (*AnalysisVerdict, error)
	// ExtractDialog is the reactive tier (speed-2): given a windowed exchange the
	// host latched onto + light context, return structured claims (for the
	// knowledge ledger) and one classified intent (which the host uses
	// deterministically to decide whether to interrupt). Cheap tier (Haiku /
	// Sonnet on nuance), off the Act loop. Never errors the host's reactive path.
	ExtractDialog(ctx context.Context, hostID, speaker, role string, window []string, personaSnippet, activeGoal string, openQuestions []string) (*DialogExtraction, error)
	Recall(ctx context.Context, q *Query) (*Knowledge, error)
	Remember(ctx context.Context, e *Episode) error
	// RecordObservation streams one raw, salience-gated perception up to mesa
	// (the firehose; cron fodder, distinct from Remember/episodes). Fire-and-forget.
	RecordObservation(ctx context.Context, o *Observation) error
	// SyncRelationships pushes the host's full trust-ledger snapshot up (AuthLocal
	// mirror). FetchRelationships pulls it back for a cold-start bootstrap.
	SyncRelationships(ctx context.Context, hostID string, rels []Relationship) error
	FetchRelationships(ctx context.Context, hostID string) ([]Relationship, error)
	// SyncGoal/FetchGoal mirror the host's standing objective + progress (structured).
	SyncGoal(ctx context.Context, hostID string, g Goal) error
	FetchGoal(ctx context.Context, hostID string) (Goal, bool, error)
	// ReportMetrics writes a host telemetry batch (observability + cron inputs).
	ReportMetrics(ctx context.Context, hostID string, metrics []Metric) error
	// KV is the GENERIC opaque-state transport — the substrate under the
	// memory.Manager remote tier (AsRemote), NOT a host-facing surface. Purposeful
	// state uses a semantic op above (Remember / SyncRelationships / SyncGoal).
	PutKV(ctx context.Context, hostID, key string, value []byte) error
	GetKV(ctx context.Context, hostID, key string) ([]byte, bool, error)
	DeleteKV(ctx context.Context, hostID, key string) error
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

// --- AnalysisInterpret: operator-override directive verdict ------------------

// AnalysisVerdict is mesa's flat classification of an operator-override
// directive (≙ the mesapb.AnalysisVerdict). Kind is "command" | "answer" |
// "hypothetical": for a command, DSL carries the one statement the host runs
// ungated; for an answer, Text carries the terse literal reply.
type AnalysisVerdict struct {
	Kind string
	DSL  string
	Text string
}

// --- ExtractDialog: reactive-tier dialog extraction -------------------------

// DialogExtraction is the reactive tier's result (≙ mesapb.ExtractedDialogSet):
// the claims to write into the host's knowledge ledger + the speaker's classified
// intent (the host decides the interrupt from Intent.Urgency).
type DialogExtraction struct {
	Claims []DialogClaim
	Intent DialogIntent
}

// DialogClaim is one extracted fact. Provenance is the LLM's advisory view; the
// host overrides it from the speaker role on writeback (a player can't claim
// system authority).
type DialogClaim struct {
	Subject    string
	Kind       string
	Claim      string
	Confidence float64
	Provenance string
}

// DialogIntent is the speaker's classified intent toward the host. Urgency
// (immediate|high|normal|low) drives the host's deterministic interrupt decision.
type DialogIntent struct {
	Kind    string
	Urgency string
	Gist    string
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

// Observation is one raw, salience-gated perception streamed up to mesa as cron
// fodder (≙ mesapb.Observation). Distinct from Episode: observations are the
// firehose the distillation crons chew on, not milestones recalled to the planner.
type Observation struct {
	HostID         string
	IdempotencyKey string
	Kind           string // entity_sighting | claim_heard | transaction | outcome
	Subject        string
	Text           string
	Salience       float64
	OccurredAtUnix int64
	Tags           map[string]string
}

// Relationship is the host's ABSOLUTE felt-trust state toward one counterparty
// (Beta alpha/beta), the snapshot form ≙ limbic.Entry. Relationships are
// AuthLocal: the host pushes a full snapshot up and mesa mirrors it (replace),
// then serves it back for a cold-start bootstrap.
type Relationship struct {
	Name        string
	Alpha       float64
	Beta        float64
	Encounters  int
	Tags        []string
	ValueTraded float64
}

// Goal is the host's MUTABLE standing objective + progress markers (distinct
// from the immutable persona north-star). Mirrored to mesa's structured goals
// table so a fresh host can resume the plan and crons/analytics can query it.
type Goal struct {
	Objective string
	Progress  []string
	UpdatedAt int64
}

// GenesisResult is the compiled session apparatus from a session-genesis call:
// a history-aware goal, a mood baseline, and the keyword→tier→action ladder.
type GenesisResult struct {
	Goal          string
	Mood          Affect
	KeywordLadder []KeywordRung
	Reasoning     string
}

// KeywordRung is one entry of the attention ladder: a word/pattern, its interrupt
// tier, and the default reflex.
type KeywordRung struct {
	Keyword string
	Tier    string
	Action  string
}

// Metric is one named telemetry sample the host reports to mesa.
type Metric struct {
	Name  string
	Value float64
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

func (StubClient) Provision(context.Context, string) (*Provisioning, error) {
	return nil, ErrOffline
}
func (StubClient) Genesis(context.Context, string, string, string) (*GenesisResult, error) {
	return nil, ErrOffline
}
func (StubClient) Chat(context.Context, string, string, string, []string) (string, bool, error) {
	return "", false, nil
}
func (StubClient) AnalysisInterpret(context.Context, string, []string) (*AnalysisVerdict, error) {
	return nil, ErrOffline
}
func (StubClient) ExtractDialog(context.Context, string, string, string, []string, string, string, []string) (*DialogExtraction, error) {
	// Offline = a safe no-op: no claims, a low-urgency statement (never interrupts).
	return &DialogExtraction{Intent: DialogIntent{Kind: "statement", Urgency: "low"}}, nil
}
func (StubClient) Act(context.Context, *Situation) (*Move, error)        { return nil, ErrOffline }
func (StubClient) Decide(context.Context, *Choice) (*Decision, error)    { return nil, ErrOffline }
func (StubClient) Recall(context.Context, *Query) (*Knowledge, error)    { return &Knowledge{}, nil }
func (StubClient) Remember(context.Context, *Episode) error              { return nil }
func (StubClient) RecordObservation(context.Context, *Observation) error { return nil }
func (StubClient) SyncRelationships(context.Context, string, []Relationship) error {
	return nil
}
func (StubClient) FetchRelationships(context.Context, string) ([]Relationship, error) {
	return nil, nil
}
func (StubClient) SyncGoal(context.Context, string, Goal) error { return nil }
func (StubClient) FetchGoal(context.Context, string) (Goal, bool, error) {
	return Goal{}, false, nil
}
func (StubClient) ReportMetrics(context.Context, string, []Metric) error { return nil }
func (StubClient) PutKV(context.Context, string, string, []byte) error   { return nil }
func (StubClient) GetKV(context.Context, string, string) ([]byte, bool, error) {
	return nil, false, nil
}
func (StubClient) DeleteKV(context.Context, string, string) error { return nil }
func (StubClient) Subscribe(context.Context, string) (<-chan Directive, error) {
	ch := make(chan Directive)
	close(ch)
	return ch, nil
}
func (StubClient) Healthy() bool { return false }
