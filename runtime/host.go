package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/cognition/corpus"
	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/cognition/resolve"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/limbic"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/persona"
	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
	"github.com/gen0cide/westworld/world"
	"github.com/gen0cide/westworld/worldmap"
)

// Options for creating a Host.
type Options struct {
	Server   string
	Username string
	Password string

	ClientVersion uint16
	RSAPublicKey  *v235.RSAPublicKey

	// Facts is the host's read-only knowledge of static world data.
	// In a swarm, one *facts.Facts is loaded once per process and
	// shared by pointer across all hosts; do not allocate per-host.
	// Optional — if nil, the host has no general world knowledge
	// (still works for protocol/walk/state, but the brain can't
	// answer "where's the nearest bank" questions).
	Facts *facts.Facts

	// Landscape is the binary .orsc landscape archive used by the
	// client-side pathfinder. Like Facts, it's safe to share one
	// *Landscape pointer across every host in a swarm. Optional —
	// if nil, action methods fall back to sending a single-tile walk
	// hint instead of a full BFS-routed path.
	Landscape *pathfind.Landscape

	// WorldOracle is the precomputed static-geography engine backing the
	// search_map / reachable / survey_map perception verbs (global walkability
	// + capability-gated transport). Like Facts/Landscape it is loaded once per
	// process and shared by pointer across all hosts — immutable after
	// Precompute, every query passes the host's Capability per-call, so it
	// needs no locking. Optional — if nil, those verbs report "no map data
	// loaded".
	WorldOracle *worldmap.Oracle

	Logger            *slog.Logger
	HeartbeatInterval time.Duration
	EventBufferSize   int
}

// Host is one bot's entire runtime: connection, world state, event
// bus, and the main event loop that ties them together.
type Host struct {
	opts Options

	conn        *session.Conn
	world       *world.World
	bus         *event.Bus
	facts       *facts.Facts
	landscape   *pathfind.Landscape
	worldOracle *worldmap.Oracle
	log         *slog.Logger

	// Strategist + Retriever are the cognition+brain layer hooks
	// used by the routine builtins contemplate_reality/decide/
	// evaluate/recall/etc. Defaults: stub implementations from
	// brain/ and cognition/ that return deterministic canned
	// values. Phase 3+ replaces these with real implementations
	// (mesa retrieval + Anthropic LLM call). Hosts share interfaces
	// safely across goroutines — one instance per process is fine.
	Strategist brain.Strategist
	Retriever  cognition.Client

	// Pearl is the host's onboard policy/quirk engine — the deterministic
	// fast path that gates and shapes decisions/actions WITHOUT an LLM. When
	// non-nil it is consulted before decide()'s Strategist (TryDecide) and
	// wraps every state-mutating action (Gate). Nil-default: no engine means
	// no gating and no fast path — behavior is identical to today. Production
	// wiring sets it from a persona-compiled table, exactly like Strategist.
	// Safe to share read-only across a host's routines.
	Pearl *pearl.Engine

	// Memory is the host's tiered memory manager — the location-blind backend
	// for the remember/recollect/forget DSL verbs. It routes each operation
	// across scratch → local → mesa per a namespace policy (cache cascade,
	// write-back journal, negative caching, maturity-dialed remote reads). Nil
	// means those verbs report NOT_IMPLEMENTED; production wiring builds one
	// from the host's hostkv store + scratch (see cmd/host).
	Memory *memory.Manager

	// affect + ledger are the host's System-1 limbic state, maintained by the
	// runLimbic bus-subscriber goroutine: affect is the mood vector (feeds the
	// pearl engine's affect predicates), ledger is the Beta(α,β) trust ledger
	// (backs relation_with and the relationship predicates). Always non-nil
	// after New; deterministic, no LLM.
	affect *limbic.Affect
	ledger *limbic.Ledger

	// knowledge is the host's SEMANTIC world-knowledge ledger (cognition/knowledge):
	// graded, provenance-tagged beliefs about THINGS (npcs/places/shops/items/
	// mechanics) — the third leg of the mind beside the trust ledger (who) and the
	// journal (what happened). Always non-nil after New; persisted under the
	// "knowledge:" namespace via loadKnowledge/flushKnowledge (runtime/knowledge.go).
	// Phase 1 = the structure + persistence; writers/distillation land later.
	knowledge *knowledge.Ledger

	// goalGraph is the host's INTENTION graph (cognition/goalgraph): goals,
	// sub-goals, open-goals and open questions as nodes with typed edges
	// (requires/produces/enables/blocked_by/serves) — the structure that makes
	// actions purposeful and failures generative, and the anti-stuck backbone. A
	// lightweight accreting memory graph (read/traversed by the host, grown by
	// crons), NOT a planner. Persisted under "goalgraph:" (runtime/goalgraph.go).
	goalGraph *goalgraph.Graph

	// curiosity is the persona's explore<->exploit dial vector, captured at
	// bootstrap (the persona is otherwise discarded after applyPersona). Read at
	// decision time (mesa_director) to bias optional curiosity detours when an
	// unknown does NOT block the active goal. Zero value = neutral (no bias).
	curiosity persona.Curiosity

	// northStar is the persona's standing north-star statement, captured at
	// bootstrap (same chokepoint as curiosity — the persona is otherwise discarded
	// after applyPersona). The director's selectNextGoal falls back to it when the
	// active goal closes and no graph open_goal successor is queued, so a host that
	// finishes its genesis goal advances toward its identity instead of idling
	// forever. Empty for personaless hosts (REPL/test).
	northStar string

	// keywordLadder is a read-only session snapshot of the genesis attention
	// ladder (the words/people that catch the host's attention), pushed onto the
	// host at bootstrap. The reactive trigger detector (reactive.go) reads it to
	// decide trigger-vs-ambient — deterministically, no LLM. Empty for hosts
	// without a genesis ladder (REPL/test).
	//
	// Held in an atomic.Pointer because the bootstrap SetKeywordLadder write happens
	// AFTER `go host.Run` has started the limbic goroutine (the genesis ladder is only
	// known post-Run, once the world has filled), which ranges the ladder for every
	// dialog line in triggerHit — a plain slice-header write there raced the range
	// (H12, torn header → OOB). The pointer load/store is atomic; readers go through
	// keywordRungs(). nil pointer (never Set) reads as an empty ladder.
	keywordLadder atomic.Pointer[[]mesaclient.KeywordRung]

	// personaSnippet is a one-line "who I am" grounding card, captured at
	// applyPersona (the persona is otherwise discarded). Shipped to mesa with a
	// reactive ExtractDialog call so the cheap-tier extractor has minimal persona
	// grounding. Empty when no persona was applied.
	personaSnippet string

	// reactive is the host's speed-2 reactive-tier state: bounded per-speaker
	// conversation windows + latches (RAM-only, capped). nil for REPL/test hosts
	// (every reactive method no-ops on nil). See runtime/reactive.go.
	reactive *reactiveState

	// obsInflight caps the concurrent fire-and-forget observation-emit goroutines
	// (the speed-3 firehose, observation.go emitObservation). Mirrors
	// reactive.inflight / reactiveMaxInflight: a non-blocking select-acquire before
	// spawning, a DROP when full (an ambient observation is not worth queueing), and
	// a release in the goroutine's defer — so a salience burst can't spawn an
	// unbounded number of RecordObservation goroutines (M15).
	obsInflight chan struct{}

	// speech is the intent-driven speech gate's anti-spam state: per-question and
	// per-target cooldowns for the proactive ASK drive + the volunteer-TEACH limit
	// (RAM-only, mutex-guarded, bounded by the open-question count). Kept separate
	// from reactiveState so the reactive mutex stays uncontended. nil for REPL/test
	// hosts that never run socialReflex. See runtime/speech.go.
	speech *speechGate
	forage *forageGate // 5b: directed-foraging drive state (RAM cache + inflight TTL)
	// blocked is the learned-impassable ledger (locked doors, toll gates):
	// the traversal flow writes it on refusal and skips ledgered obstacles
	// (TTL-bounded), so a locked door is tried twice, REMEMBERED, and
	// routed around instead of re-attempted on every replan.
	blocked *blockedEdges

	// emitSay is the chat-emission seam the proactive ASK drive uses. nil in
	// production ⇒ the real Host.Say (network send + reactive self-line fan-in);
	// a test overrides it to capture the line without a live socket. The reflex
	// reply path always uses Host.Say directly — only the off-loop ASK path reads
	// this so its deterministic gate is unit-testable. See runtime/speech.go.
	emitSay func(context.Context, string) error

	// perceive is the perception writers' tiny deterministic cursor: cross-event
	// context (last-seen named NPC for shop attribution) + dedup state (last area
	// keyed for familiarity, per-shop stock snapshot) so the handler stays O(1)
	// and never re-writes unchanged perception. RAM-only; not persisted. The zero
	// value is valid (maps lazy-init in the handler). See runtime/perception.go.
	perceive perceptionState

	// journal is the host's durable EPISODIC memory — a bounded, importance-
	// ranked log of what it did this life (level-ups, kills, deaths, objective
	// milestones) plus its standing objective. Maintained by the runMemory
	// bus-subscriber goroutine (capture) and recalled by the director into the
	// per-turn Situation (so the planner reasons over what it has done instead
	// of re-deriving the world each tick). Always non-nil after New; persisted
	// through Memory under the "journal:" namespace. See docs/cognition-and-
	// autonomy.md §5.
	journal *memory.Journal

	// mesaMem is the two-way mesa state seam (Postgres): it mirrors local state
	// UP (episodes, the trust ledger) and, on a cold start with no local store,
	// pulls it back DOWN to reconstitute. bbolt is the fast local warm-start;
	// mesa is the authoritative source that can bootstrap a fresh / in-memory
	// host from nothing. Wired by cmd/host from the mesa client; nil when offline.
	mesaMem MesaMemory

	// decisionCache memoizes Strategist (Haiku) decide() verdicts so a repeated
	// pearl-MISS decision in materially-the-same state skips the LLM call — the
	// decision half of the cheap loop (#16). Bounded LRU+TTL; keyed by
	// question+options+coarse state. Pearl hits are never cached (already free).
	decisionCache *hostkv.Scratch

	// memoryWarmupUntil suppresses LevelUp capture during the post-login window.
	// On login the server sends the full stats snapshot, which the edge detector
	// turns into a LevelUp for every skill (baseline 0 → current) — a burst of
	// phantom "achievements" that are really just the initial sync. Genuine
	// level-ups only happen well into play (the first Act decision alone takes
	// ~10s), so the journal ignores LevelUp until this deadline. Touched only by
	// the runMemory goroutine.
	memoryWarmupUntil time.Time

	// Corpus is the shared-knowledge retrieval surface (rsc.wiki +
	// AutoRune script archive). When non-nil, the `recall()` DSL
	// builtin queries it directly and returns real chunks; when
	// nil, recall falls back to the per-host Retriever's reflection
	// list (Phase 2.5 stub behavior). Multiple hosts can share one
	// Corpus instance — chunks are read-only after load.
	//
	// Phase 2.6 Slice 1: an in-memory MemoryCorpus loaded from the
	// rsc.wiki HTML dump. Slice 2 swaps to a mesa-backed corpus
	// without changing this field's type.
	Corpus corpus.Corpus

	// Resolver is the host's recognition faculty — the engine behind the
	// fenced control-plane `resolve()` / `resolve_one()` primitives
	// (api.md §5). It maps loose player text ("r2h") to canonical facts
	// definitions through the learned-alias → fuzzy → brain pipeline.
	//
	// When nil, the resolve builtins lazily construct an in-memory
	// resolver from the host's Facts (learning works for the session but
	// is not persisted, and the brain stage is skipped). Production
	// wiring (cmd/cradle) sets this to a resolver backed by a per-host
	// JSON alias store so learned lingo survives restarts. A single
	// *facts.Facts catalog is safe to share; each host keeps its own
	// alias store.
	Resolver *resolve.Resolver

	// Last-attacked entity indices for the combat.last_npc /
	// combat.last_player accessors. Set on attack() dispatch;
	// stays set across the entity leaving view so routines can
	// flee/respawn-loot even after the target despawns. Zero
	// means "no attack this session" (server indices are 1+).
	// Written on the conductor goroutine (attack dispatch) but ALSO read on the
	// frame-pump (Apply death-edge), limbic (pkKiller / engagedNpcName), and views
	// goroutines — so each is its own atomic.Int32 rather than a plain int (the old
	// "same goroutine" comment was wrong; the cross-goroutine reads raced). Server
	// indices are 1+, so 0 still means "no attack this session".
	lastAttackedNpcIndex    atomic.Int32
	lastAttackedPlayerIndex atomic.Int32

	// combatStyle mirrors the most-recently-applied melee xp-split
	// mode for the read-side combat.style accessor (#117). RSC sends
	// no confirmation packet for opcode 29, so this is a write-through
	// mirror set on combat.set_style dispatch — it reflects our intent,
	// not a server echo. Its zero value is CombatStyleControlled (0),
	// the server's start state, so the view reports "controlled" before
	// any set_style call without extra bookkeeping.
	combatStyle action.CombatStyle

	// combatMu guards the combat-tracking block below (combatStartedAt,
	// combatRounds, combatRoundTarget, outgoingHits) — and the duel-fight window
	// (duelFightUntil). These are written by the frame-pump goroutine
	// (noteCombatRound / emitTargetDeathEdge), written+read by the conductor/routine
	// goroutine (beginCombatRoundTracking / confirmEngaged / the retreat gate), and
	// read by the limbic goroutine (pkKiller / the own-death gate). Three goroutines,
	// so the plain-int fields raced (H11, confirmed -race). A single mutex keeps the
	// host light. (lastAttacked*Index are NOT under this mutex — they are their own
	// atomic.Int32 fields, see the decls above, because they are read on more
	// goroutines than this block and a single mutex would over-couple the hot read path.)
	combatMu sync.Mutex

	// Combat-round tracking for the anti-kite retreat gate (#r3-retreat).
	// RSC forbids retreating until the opponent has made >= 3 hits — the
	// "first 3 rounds of combat" (WalkRequest.java checks
	// opponent.getHitsMade() >= 3). We approximate the opponent's
	// hits-made count by tallying combat-round damage exchanges observed
	// while engaged with the current target:
	//   - combatStartedAt is stamped when we attack() a fresh target and
	//     zeroed when combat ends (target dies / we disengage), so a
	//     wall-clock estimate (one round per ~combatRoundTick) is
	//     available as a fallback when no damage has been observed yet.
	//   - combatRounds counts each damage-bearing combat tick we see on
	//     this engagement (our hit on the target OR the target's hit on
	//     us). It is the engine's clean round-count signal; the retreat
	//     verb can wait until it reaches retreatRoundGate before sending.
	// Both reset on a new attack() against a different target. The
	// server's rejection message remains the authoritative gate; this
	// counter only lets retreat() wait pre-emptively instead of blindly
	// poking the server every tick.
	combatStartedAt   time.Time
	combatRounds      int
	combatRoundTarget int // npc/player index this round count belongs to

	// lastDuelOpponent / lastTradeOpponent capture the counterparty NAME at the
	// moment a duel/trade is offered, as a belt-and-suspenders attribution net for
	// the limbic relationship updates (DuelClosed / trade outcome) regardless of
	// bus ordering or world-state terminal phase. In-RAM, deterministic; written +
	// read only on the runLimbic subscriber goroutine.
	lastDuelOpponent  string
	lastTradeOpponent string

	// outgoingHits counts damage WE dealt to the engaged NPC
	// (combatRoundTarget) — each NpcDamage on that index is a hit we
	// landed (an NPC never damages itself). Unlike combatRounds (which
	// also counts the NPC hitting US), this is a one-directional "are we
	// actually swinging" signal: AttackNpc.confirmEngaged uses it to tell
	// a live fight (don't re-attack — would reset combat) apart from the
	// passive-victim stall (bot adjacent taking hits but dealing none —
	// MUST re-attack). Reset alongside the round counter on a new/cleared
	// engagement.
	outgoingHits int

	// duelFightUntil is the "we are in a sporting fight" window (C2). A duel's
	// DuelClosed{Completed:true} fires when the FIGHT STARTS (not when it ends), so
	// the world Duel record is already terminal (IsActive()==false) by the time a
	// Death/projectile lands DURING the fight — leaving the grievance gate OPEN and
	// mis-recording a skulled wilderness-duel death as a PK gank. We stamp this on
	// fight-start and treat any death/hostile projectile before it as a duel event,
	// not a wrong. Cleared on respawn (the post-death StatsSnapshot). Guarded by
	// combatMu (written on the limbic goroutine, read there too). Zero = not fighting.
	duelFightUntil time.Time

	// routineCtx is the context bound by the active routine interpreter
	// (set in NewRoutineInterpreter). Namespace-dispatched action
	// callables (trade.request, bank.deposit, magic.cast, …) carry no
	// ctx of their own — they read it from here so cancellation /
	// deadline propagation matches the flat builtins. nil outside a
	// routine run (falls back to context.Background()).
	routineCtx context.Context

	// OnStmt, when set, is installed as the interpreter's per-statement
	// hook by NewRoutineInterpreter — it fires with the source line about
	// to execute, once per statement, on the routine goroutine. The
	// conductor sets it to track the current line for the cradle's live
	// Routine panel (cheap line tracking, not a per-statement bus event).
	// Must be O(1) and non-blocking. Nil = no per-statement observation.
	// Set once at wiring time and read by every routine run.
	OnStmt func(line int)

	// analysis is the host's OPERATOR-OVERRIDE state — Westworld "analysis mode".
	// While active, the autonomous conductor is frozen, in-character replies to
	// the world are suppressed, ALL memory writes are suspended (episodic journal,
	// LTM mirror, limbic trust/affect — except an explicit remember command), and
	// the operator's directives are interpreted with a full bypass of
	// pearl/persona/cognition. Both the in-game "!<username> ANALYSIS" trigger and
	// the cradle control-plane converge on this single shared object. Guarded by
	// its own mutex; the bus-goroutine memory gates read .Active() lock-free via
	// the atomic mirror. See analysis.go.
	analysis analysisState

	// displacement carries the most recent unexpected position jump from the
	// displacementArbiter (a conductor goroutine) to the director, so a re-plan
	// after a displacement abort tells the planner the host was MOVED rather
	// than that its action failed. Consume-once; see detour.go.
	displacement displacementState

	// liveGoal holds an operator goal pushed at runtime (mesa Provision.Subscribe
	// GOAL_REVISION, from `mesa-ctl goal push`) — a SOFT override of the director's
	// genesis/persona goal. Empty until a push arrives; latest-wins. Read by the
	// director each turn. See subscribeDirectives and livegoal.go.
	liveGoal liveGoalState

	loggedIn bool
}

// SetLiveGoal installs an operator goal pushed at runtime (mesa GOAL_REVISION),
// a soft override of the director's construction-time goal. Thread-safe; called
// from the subscribe goroutine.
func (h *Host) SetLiveGoal(goal string) { h.liveGoal.set(goal) }

// LiveGoal returns the current runtime goal override, or "" if none has been
// pushed. The director prefers it over its genesis/persona goal.
func (h *Host) LiveGoal() string { return h.liveGoal.get() }

// SetKeywordLadder pushes the genesis attention ladder onto the host as a
// read-only session snapshot — the substrate for the reactive trigger detector
// (reactive.go). Called once at bootstrap (mirrors MesaDirector.SetKeywordLadder),
// AFTER the limbic reader goroutine is live, so the write is done atomically to
// avoid a torn slice header racing triggerHit's range (H12).
func (h *Host) SetKeywordLadder(l []mesaclient.KeywordRung) { h.keywordLadder.Store(&l) }

// keywordRungs returns the current genesis attention ladder snapshot (the atomic
// read paired with SetKeywordLadder). Returns nil — a valid empty range — when no
// ladder was ever set. The reactive trigger detector ranges this.
func (h *Host) keywordRungs() []mesaclient.KeywordRung {
	if p := h.keywordLadder.Load(); p != nil {
		return *p
	}
	return nil
}

// conductorHandle returns the live conductor bound at bootstrap (configure-
// Analysis), or nil for REPL/test hosts. Reuses the analysis handle — there is
// exactly one conductor per host and analysisState already holds it, so the
// reactive interrupt path borrows it rather than duplicating the field.
func (h *Host) conductorHandle() *Conductor {
	h.analysis.mu.Lock()
	defer h.analysis.mu.Unlock()
	return h.analysis.conductor
}

// New constructs a Host (no I/O yet). Call Connect to dial+login,
// then Run to drive the main loop.
func New(opts Options) *Host {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = 5 * time.Second
	}
	if opts.EventBufferSize == 0 {
		opts.EventBufferSize = 64
	}
	if opts.ClientVersion == 0 {
		opts.ClientVersion = 235
	}
	h := &Host{
		opts:        opts,
		world:       world.NewWorld(),
		bus:         event.NewBus(),
		facts:       opts.Facts,
		landscape:   opts.Landscape,
		worldOracle: opts.WorldOracle,
		log:         opts.Logger,
		// Stub strategist + retriever by default. Production
		// wiring overrides these with real implementations
		// after Phase 3/4 land.
		Strategist: &brain.StubStrategist{},
		Retriever:  &cognition.StubClient{},
		// System-1 limbic state: neutral mood baseline + an empty trust ledger.
		// Driven by runLimbic once Run starts; safe to read before then.
		affect:    limbic.NewAffect(0, 0.5, 0, 0),
		ledger:    limbic.NewLedger(),
		knowledge: knowledge.NewLedger(),
		goalGraph: goalgraph.New(),
		// Reactive tier (speed-2): bounded per-speaker windows + latches. Driven by
		// the perception bus handler + the Say send seam once Run starts; safe to
		// read before then. RAM-only, never persisted.
		reactive: newReactiveState(),
		// Speed-3 firehose emit semaphore: bounds concurrent observation goroutines
		// (M15), mirroring reactive's inflight cap.
		obsInflight: make(chan struct{}, observationMaxInflight),
		// Intent-driven speech gate (ask/answer/teach anti-spam). Driven by
		// socialReflex once Run starts; RAM-only, never persisted.
		speech:  newSpeechGate(),
		forage:  newForageGate(),
		blocked: newBlockedEdges(),
		// Episodic memory: an empty journal. Driven by runMemory once Run starts
		// (restored from durable storage there); safe to read before then.
		journal: memory.NewJournal(0),
		// Decision cache: bounded LRU memoizing decide() Strategist verdicts.
		decisionCache: hostkv.NewScratch(256),
	}
	// Tell the world mirror our username so it can identify our own
	// server player index from appearance updates (we are NOT always
	// index 0). Drives correct self appearance/combat attribution.
	h.world.Players.SetSelfName(opts.Username)
	return h
}

// Facts returns the host's shared knowledge base (may be nil if no
// Facts were passed in opts).
func (h *Host) Facts() *facts.Facts { return h.facts }

// WorldOracle returns the host's shared static-geography engine (may be nil if
// no oracle was precomputed / passed in opts). Backs the search_map /
// reachable / survey_map perception verbs.
func (h *Host) WorldOracle() *worldmap.Oracle { return h.worldOracle }

// resolver returns the host's recognition faculty, lazily building an
// in-memory one over the host's Facts if none was wired. The lazy
// default has no persistence (alias learning lasts the session only)
// and skips the brain stage (nil brain → pipeline is alias + fuzzy).
// Production callers set h.Resolver explicitly (see cmd/cradle) to get
// a persisted per-host alias store; this fallback keeps resolve()
// useful in tests and the REPL without extra setup.
//
// The lazily-built resolver is cached on h.Resolver so a host shares
// one alias store across all resolve() calls in a session.
func (h *Host) resolver() *resolve.Resolver {
	if h.Resolver == nil {
		h.Resolver = resolve.New(h.facts, nil, nil)
	}
	return h.Resolver
}

// isStackableItem is the callback handed to v235.DecodeInbound so
// the inventory packet decoder knows when to read the stackable-
// amount field. Treats every item as unstackable when facts isn't
// loaded — safe under-read; over-read would corrupt later slots
// (see decodeInventory comment).
func (h *Host) isStackableItem(itemID int) bool {
	if h.facts == nil {
		return false
	}
	d := h.facts.ItemDef(itemID)
	if d == nil {
		return false
	}
	return d.IsStackable
}

// World returns the world state mirror. Read-only via the returned
// pointer's accessors (which are themselves rwlock-safe).
func (h *Host) World() *world.World { return h.world }

// Bus returns the host's event bus for subscribing to typed events.
func (h *Host) Bus() *event.Bus { return h.bus }

// Conn returns the underlying session connection. Use sparingly —
// most callers should go through Action* methods.
func (h *Host) Conn() *session.Conn { return h.conn }

// Connect dials the server and completes the login handshake.
func (h *Host) Connect(ctx context.Context) error {
	conn, err := session.Dial(ctx, h.opts.Server, session.Options{Logger: h.log})
	if err != nil {
		return fmt.Errorf("runtime: dial: %w", err)
	}
	h.conn = conn

	res, err := conn.Login(ctx, session.LoginParams{
		Username:      h.opts.Username,
		Password:      h.opts.Password,
		ClientVersion: h.opts.ClientVersion,
		RSAPublicKey:  h.opts.RSAPublicKey,
	})
	if err != nil {
		return fmt.Errorf("runtime: login: %w", err)
	}
	h.log.Info("host logged in",
		"username", h.opts.Username,
		"response", res.ResponseCode,
	)
	h.loggedIn = true
	conn.Start()
	return nil
}

// Run drives the main host loop until ctx is cancelled or the
// connection terminates. Returns ctx.Err() on normal cancellation,
// or the connection's terminal error otherwise.
func (h *Host) Run(ctx context.Context) error {
	if !h.loggedIn {
		return fmt.Errorf("runtime: Run called before Connect")
	}

	heartCtx, stopHeart := context.WithCancel(ctx)
	defer stopHeart()
	go h.heartbeatLoop(heartCtx)
	// System-1 limbic path: a second bus-subscriber goroutine (no tick) that
	// folds game events into affect + the trust ledger. Deterministic, no LLM.
	go h.runLimbic(heartCtx)
	// Episodic-memory path: a third bus-subscriber goroutine (no tick) that
	// folds salient events into the durable Journal. Deterministic, no LLM.
	go h.runMemory(heartCtx)
	// Telemetry path: a fourth bus-subscriber goroutine that tallies counters
	// and reports host metrics to mesa on a cadence. Best-effort, no LLM.
	go h.runMetrics(heartCtx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case frame, ok := <-h.conn.Recv():
			if !ok {
				// Channel closed by session — connection terminated.
				if err := h.conn.Err(); err != nil {
					return fmt.Errorf("runtime: connection error: %w", err)
				}
				return nil
			}
			h.handleFrame(frame)
		}
	}
}

// heartbeatLoop sends a keepalive every HeartbeatInterval.
func (h *Host) heartbeatLoop(ctx context.Context) {
	t := time.NewTicker(h.opts.HeartbeatInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := action.Heartbeat(ctx, h.conn); err != nil {
				h.log.Warn("heartbeat failed", "err", err)
				return
			}
		}
	}
}
