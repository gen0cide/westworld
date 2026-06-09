package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

// sleepWord is the hardcoded answer to the fatigue sleep-screen captcha
// on this OpenRSC server. The server's CaptchaGenerator falls back to a
// prerendered image of the word "asleep" and sets
// player.setSleepword("asleep") when no prerendered sleepword set is
// loaded (CaptchaGenerator.generateRSCLCaptcha — CaptchaGenerator.java:
// 79-80). SleepHandler.process then accepts our typed word iff it
// equalsIgnoreCase the stored word. So we always answer "asleep" — no
// OCR of the captcha bitmap required.
const sleepWord = "asleep"

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
		speech: newSpeechGate(),
		forage: newForageGate(),
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

// handleFrame decodes a frame, applies it to world state, publishes
// the event(s).
//
// Some opcodes produce MULTIPLE events from one packet (UpdatePlayers,
// PlayerCoords with nearby players, NpcCoords). Those are special-cased
// here. Single-event opcodes flow through DecodeInbound.
func (h *Host) handleFrame(f v235.Frame) {
	switch f.Opcode {
	case v235.InSendPlayerCoords:
		own, nearby, err := v235.DecodePlayerCoords(f.Payload)
		if err != nil {
			h.log.Warn("decode playercoords", "err", err)
			return
		}
		h.world.Apply(own)
		h.bus.Publish(own)
		for _, np := range nearby {
			// Apply to world.Players so position queries
			// (world.players.find(...).position) resolve to the
			// most-recent observed coords. Previously only
			// published — world.Players never got the update, so
			// `target.position` returned (0, 0) and walk_to to a
			// remote player walked to the world origin.
			h.world.Apply(np)
			h.bus.Publish(np)
		}
		return
	case v235.InSendUpdatePlayers:
		events, err := v235.DecodeUpdatePlayers(f.Payload)
		if err != nil {
			h.log.Warn("decode updateplayers", "err", err)
			return
		}
		for _, ev := range events {
			// For another player's appearance, diff combat level + worn
			// equipment against what we last saw and fire edge events
			// (player_level_changed / player_equipment_changed) — the raw
			// appearance packet re-sends periodically, so only an actual
			// change should wake a routine. Snapshot BEFORE Apply.
			if ap, ok := ev.(event.OtherPlayerAppearance); ok {
				prev, had := h.world.Players.Get(ap.PlayerIndex)
				h.world.Apply(ev)
				h.bus.Publish(ev)
				if had && ap.HasCombat && prev.HasAppearanceCombat && prev.CombatLevel != ap.CombatLevel {
					h.bus.Publish(event.PlayerLevelChanged{
						PlayerIndex: ap.PlayerIndex, Name: ap.Name,
						OldLevel: prev.CombatLevel, NewLevel: ap.CombatLevel,
					})
				}
				if had && ap.HasWorn && prev.HasEquip {
					// One event per human slot that actually changed, each
					// naming the slot so the handler gets good information.
					for _, slot := range h.diffEquip(prev.EquipBySlot, ap.WornSprites) {
						h.bus.Publish(event.PlayerEquipmentChanged{PlayerIndex: ap.PlayerIndex, Name: ap.Name, Slot: slot})
					}
				}
				continue
			}
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InUpdateNpc:
		events, err := v235.DecodeUpdateNpcs(f.Payload)
		if err != nil {
			h.log.Warn("decode updatenpcs", "err", err)
			return
		}
		for _, ev := range events {
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InNpcCoords:
		pos := h.world.Self.Position()
		// Snapshot the ordered local-NPC list BEFORE applying events:
		// the opcode-79 update records are positional against the
		// server's localNpcs list, which this mirror tracks. Apply()
		// then mutates the order (prune on REMOVE, append on new).
		order := h.world.Npcs.Order()
		events, localCount, err := v235.DecodeNpcCoords(f.Payload, pos.X, pos.Y, order)
		if err != nil {
			h.log.Warn("decode npccoords", "err", err)
			return
		}
		// Tier-1 anomaly assertion (docs/lang/protocol-debug-strategy.md):
		// the opcode-79 existing-NPC records are positional against the
		// server's localNpcs list, which `order` mirrors. If their lengths
		// diverge, our slot->index map has desynced and every positional
		// update from here is misattributed (NPCs silently vanish from
		// world.npcs). Log loudly so the desync is caught at its source
		// instead of surfacing as a mysterious perception gap.
		if localCount != len(order) {
			h.log.Warn("npccoords order desync: localNpcs count != our order list — NPC tracking corrupted",
				"server_local_count", localCount,
				"our_order_len", len(order),
			)
		}
		for _, ev := range events {
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InSleepScreen:
		// SEND_SLEEPSCREEN (opcode 117): the fatigue sleep-screen
		// captcha is up. We apply+publish the SleepScreenAppeared event
		// (sets self.is_sleeping = true) and then AUTO-RESPOND with the
		// sleep word. On this OpenRSC server the word is hardcoded to
		// "asleep" (CaptchaGenerator.java:79-80) — no OCR/image-solving
		// needed — so we immediately send SLEEPWORD_ENTERED("asleep") to
		// wake + reset fatigue. The server replies with SEND_STOPSLEEP
		// (handled below) on a correct word.
		//
		// (Small, clearly-commented case mirroring the InUpdateNpc case
		// added in the wave-2 stitch commit.)
		ev := event.SleepScreenAppeared{ImageBytes: len(f.Payload)}
		h.world.Apply(ev)
		h.bus.Publish(ev)
		if err := action.SendSleepWord(context.Background(), h.conn, sleepWord); err != nil {
			h.log.Warn("auto-answer sleep word", "err", err)
		} else {
			h.log.Debug("auto-answered sleep word", "word", sleepWord)
		}
		return
	case v235.InStopSleep:
		// SEND_STOPSLEEP (opcode 84, no payload): the server woke us —
		// the sleep word was correct (or sleep otherwise ended). Clear
		// the sleep state (self.is_sleeping = false). The reset fatigue
		// value arrives separately as a FatigueUpdate packet.
		ev := event.SleepEnded{}
		h.world.Apply(ev)
		h.bus.Publish(ev)
		return
	}

	// Single-event opcodes.
	ev, err := v235.DecodeInbound(f, h.isStackableItem)
	if err != nil {
		h.log.Warn("decode error",
			"opcode", fmt.Sprintf("0x%02x (%d)", f.Opcode, f.Opcode),
			"err", err,
		)
		return
	}
	if ev == nil {
		return
	}
	// Log every player-facing server message at INFO so each one is
	// visible in cradle stdout — the server's response to whatever a
	// routine just did becomes observable instead of being silently
	// discarded. This is the single chokepoint: all message-bearing
	// opcodes (131 server/chat, 120 PM, 222/89 dialog box, 245 menu)
	// arrive through this single-event path. One tidy line per message;
	// the runner sweep greps these "server msg" lines to explain a
	// non-PASS scenario. See logServerMessage.
	h.logServerMessage(ev)
	// Snapshot inventory counts before Apply so we can emit synthetic
	// ItemGained events afterward. Cheap (map iteration) and only
	// runs when the event might change inventory. Skipping for
	// non-inventory events keeps the hot path clean.
	var preCounts map[int]int
	// preEquip snapshots own worn gear per human slot before an inventory
	// Apply, so we fire one equipment_changed(slot) per slot that changed.
	var preEquip map[string]int
	switch ev.(type) {
	case event.InventorySnapshot, event.InventorySlotUpdate:
		preCounts = h.inventoryCounts()
		preEquip = h.selfEquipSnapshot()
	}
	// Snapshot per-skill xp before Apply for xp-bearing events so we
	// can emit synthetic XPGain deltas afterward (#119). The raw
	// packets carry the NEW TOTAL, not a delta — only a diff yields
	// the gain. Same edge-detection pattern as ItemGained.
	var preXP, preMax map[int]int
	switch ev.(type) {
	case event.StatUpdate, event.StatsSnapshot, event.ExperienceGain:
		preXP = h.skillXPSnapshot()
		preMax = h.skillMaxSnapshot()
	}
	// Snapshot the engaged target's pre-Apply health so we can detect
	// the death edge after NpcDamage lands (#119 target_died / npc_killed).
	preTargetAlive := false
	if d, ok := ev.(event.NpcDamage); ok && d.NpcIndex == int(h.lastAttackedNpcIndex.Load()) {
		preTargetAlive = h.engagedTargetAlive()
	}
	changed := h.world.Apply(ev)
	if changed {
		h.log.Debug("world updated", "by", ev.Kind())
	}
	if preCounts != nil {
		h.emitItemGainedDeltas(preCounts)
	}
	if preXP != nil {
		h.emitXPGainDeltas(preXP)
	}
	if preMax != nil {
		h.emitLevelUpDeltas(preMax)
	}
	if preEquip != nil {
		for name, id := range h.selfEquipSnapshot() {
			if preEquip[name] != id {
				h.bus.Publish(event.EquipmentChanged{Slot: name})
			}
		}
	}
	if d, ok := ev.(event.NpcDamage); ok && d.NpcIndex == int(h.lastAttackedNpcIndex.Load()) {
		h.emitTargetDeathEdge(d.NpcIndex, preTargetAlive)
	}
	// Tally combat rounds for the anti-kite retreat gate (#r3-retreat).
	// Each combat tick produces a type-2 damage update on a participant:
	// the target's hp drops when we hit it (NpcDamage on the engaged
	// index) and our own hp drops when it hits us (OtherPlayerDamage on
	// self). Either is one round elapsed. We count both — the server's
	// gate is the OPPONENT's hits-made, but in melee the exchange is
	// near-simultaneous each tick, so counting any damage event on the
	// engagement gives a faithful round estimate. See noteCombatRound.
	h.noteCombatRound(ev)
	h.bus.Publish(ev)
}

// noteCombatRound advances the engaged round counter on each combat
// damage tick while we have an active engagement, so retreat() can
// wait until the authentic 3-round anti-kite window has elapsed
// instead of blindly poking the server. Only damage involving our
// current engagement target counts; damage from/to other entities in
// view is ignored.
func (h *Host) noteCombatRound(ev event.Event) {
	// Resolve self index OUTSIDE the combat lock (it takes the world lock) to keep
	// the two locks from nesting.
	selfIdx := h.world.Players.SelfIndex()
	h.combatMu.Lock()
	defer h.combatMu.Unlock()
	idx := h.combatRoundTarget
	if idx == 0 {
		return // no engagement we're counting rounds for
	}
	switch e := ev.(type) {
	case event.NpcDamage:
		// Our hit landed on the engaged NPC (its hp changed).
		if e.NpcIndex == idx {
			h.combatRounds++
			h.outgoingHits++ // one-directional "we are swinging" signal
		}
	case event.OtherPlayerDamage:
		// our own index == we took a hit this round (the engaged entity
		// hit us). Closest analogue to the server's opponent.getHitsMade().
		if e.PlayerIndex == selfIdx {
			h.combatRounds++
		}
	}
}

// logServerMessage emits one INFO line for any inbound event that
// carries player-facing text, so every message the server sends is
// visible in cradle stdout (and greppable by the scenario sweep).
//
// RSC funnels almost all action feedback through opcode 131
// (SEND_SERVER_MESSAGE) — "You can't retreat for 3 rounds.", "The
// door is locked.", "Nothing interesting happens." — which decodes to
// SystemMessage (no sender) or ChatReceived (with sender). Dialog
// boxes (222/89 -> NpcDialogText), menus (245 -> NpcDialog) and PMs
// (120 -> PrivateMessage) round out the player-facing set. We tag each
// with its kind so a reader can tell server text from NPC speech.
//
// Kept deliberately uniform ("server msg" + kind + text) so the
// runner's failure dump can grep one stable token.
func (h *Host) logServerMessage(ev event.Event) {
	switch e := ev.(type) {
	case event.SystemMessage:
		h.log.Info("server msg", "kind", "server", "text", e.Message)
	case event.ChatReceived:
		// Opcode-131 with a sender: a public chat line the server
		// relayed (or, on this server, system text attributed to a
		// name). Include the sender for context.
		h.log.Info("server msg", "kind", "chat", "from", e.Speaker, "text", e.Message)
	case event.PrivateMessage:
		h.log.Info("server msg", "kind", "pm", "from", e.Sender, "text", e.Message)
	case event.NpcDialogText:
		h.log.Info("server msg", "kind", "dialog", "text", e.Text)
	case event.NpcDialog:
		h.log.Info("server msg", "kind", "menu", "options", strings.Join(e.Options, " | "))
	}
}

// skillXPSnapshot captures {skill_id -> total_xp} for all 18 skills.
// Used to diff xp before/after an xp-bearing packet so emitXPGainDeltas
// can publish synthetic XPGain events with the positive delta.
func (h *Host) skillXPSnapshot() map[int]int {
	snap := make(map[int]int, world.NumSkills)
	for id := 0; id < world.NumSkills; id++ {
		snap[id] = h.world.Self.SkillXP(id)
	}
	return snap
}

// emitXPGainDeltas compares the pre-Apply xp snapshot to the current
// per-skill xp and publishes one event.XPGain per net-positive delta.
// Routines subscribe via `on xp_gain(skill, amount) { ... }` and
// filter on the skill name. Mirrors emitItemGainedDeltas.
func (h *Host) emitXPGainDeltas(prev map[int]int) {
	for id := 0; id < world.NumSkills; id++ {
		total := h.world.Self.SkillXP(id)
		delta := total - prev[id]
		if delta > 0 {
			h.bus.Publish(event.XPGain{Skill: event.SkillID(id), Amount: delta, Total: total})
		}
	}
}

// skillMaxSnapshot captures the per-skill BASE (max) level before an
// Apply so emitLevelUpDeltas can detect a level increase afterward.
func (h *Host) skillMaxSnapshot() map[int]int {
	snap := make(map[int]int, world.NumSkills)
	for id := 0; id < world.NumSkills; id++ {
		snap[id] = h.world.Self.SkillMax(id)
	}
	return snap
}

// emitLevelUpDeltas publishes a LevelUp for each of our own skills whose
// base level rose since the snapshot — the edge that powers
// `on level_up(skill, new_level)`. Mirrors emitXPGainDeltas.
func (h *Host) emitLevelUpDeltas(prev map[int]int) {
	for id := 0; id < world.NumSkills; id++ {
		now := h.world.Self.SkillMax(id)
		if now > prev[id] {
			h.bus.Publish(event.LevelUp{Skill: event.SkillID(id), NewLevel: now})
		}
	}
}

// engagedTargetAlive reports whether the currently-engaged NPC target
// (lastAttackedNpcIndex) is known to have hitpoints remaining. True
// when we have a health reading (HasHits) and CurHits > 0, or when we
// have no reading yet (unknown defaults to "alive" so a first 0-hits
// reading still counts as a death edge). Used by the target_died
// edge detector to avoid double-firing once the NPC is already dead.
func (h *Host) engagedTargetAlive() bool {
	idx := int(h.lastAttackedNpcIndex.Load())
	if idx == 0 {
		return false
	}
	rec, ok := h.world.Npcs.Get(idx)
	if !ok {
		return false
	}
	if !rec.HasHits {
		return true // no reading yet — treat as alive
	}
	return rec.CurHits > 0
}

// emitTargetDeathEdge publishes the synthetic TargetDied event when
// our engaged target's health transitions from alive (>0 or unknown)
// to dead (CurHits == 0). Called after Apply landed the NpcDamage, so
// world.Npcs holds the post-hit reading. `wasAlive` is the pre-Apply
// liveness — the edge only fires on alive→dead, never on a repeated
// 0-hits packet, so `on target_died` / `on npc_killed` fire once.
func (h *Host) emitTargetDeathEdge(npcIndex int, wasAlive bool) {
	if !wasAlive {
		return
	}
	rec, ok := h.world.Npcs.Get(npcIndex)
	if !ok {
		return
	}
	if !rec.HasHits || rec.CurHits != 0 {
		return
	}
	// The engaged target died: combat is over, so clear the anti-kite
	// round tracking (#r3-retreat). A subsequent attack() on a new
	// target re-arms it; leaving stale state would make the next
	// engagement think rounds had already elapsed.
	h.combatMu.Lock()
	if h.combatRoundTarget == npcIndex {
		h.combatRoundTarget = 0
		h.combatRounds = 0
		h.outgoingHits = 0
		h.combatStartedAt = time.Time{}
	}
	h.combatMu.Unlock()
	h.bus.Publish(event.TargetDied{NpcIndex: npcIndex, TypeID: rec.TypeID})
}

// inventoryCounts builds a snapshot of {item_id -> total count}
// from the current world.Inventory state. Used to compute item_gained
// deltas across an inventory packet apply.
func (h *Host) inventoryCounts() map[int]int {
	counts := map[int]int{}
	for _, s := range h.world.Inventory.Slots() {
		counts[s.ItemID] += s.Amount
		if s.Amount == 0 {
			// Unstackable slots have Amount=1 baked in by the
			// decoder; defensive in case decoder reports 0.
			counts[s.ItemID] += 1
		}
	}
	return counts
}

// emitItemGainedDeltas compares the pre-Apply inventory snapshot
// to the current state and publishes one event.ItemGained for
// each net-positive delta. Routines subscribe via
// `on item_gained(item_id, count) { ... }`.
//
// Net-negative deltas don't fire — eat/drop/deposit each have
// their own surface. We could add ItemLost as a follow-up if a
// routine needs it.
func (h *Host) emitItemGainedDeltas(prev map[int]int) {
	cur := h.inventoryCounts()
	for id, curCount := range cur {
		delta := curCount - prev[id]
		if delta > 0 {
			h.bus.Publish(event.ItemGained{ItemID: id, Count: delta})
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

// Walk sends one walk-to-coord packet, with FOV validation against
// the bot's current position. Errors with action.ErrOutOfRange if the
// target is more than MaxClickRange tiles away.
//
// For long journeys, use WalkTo (which chunks into multiple in-FOV
// segments).
func (h *Host) Walk(ctx context.Context, x, y int) error {
	pos := h.world.Self.Position()
	return action.Walk(ctx, h.conn, pos.X, pos.Y, x, y)
}

// WalkTo walks toward (x, y) by sending one or more in-FOV walk
// packets, waiting for progress between them. Returns when the bot
// reaches (or stops within 1 tile of) the destination, when ctx is
// cancelled, or when progress stalls beyond stallTimeout.
//
// The "wait for progress" loop polls position; in Phase 2 this should
// be event-driven via Bus subscription, but polling is simpler for
// Phase 1.5.
// DoorLockedError is returned by WalkTo when an openable boundary
// on the path was contacted, the open interaction was attempted,
// but the host couldn't pass after the retry budget. Distinct
// from a plain stall (which means "no door, just terrain") so the
// DSL layer can map it to ErrorCode DOOR_LOCKED instead of the
// generic PATH_BLOCKED.
//
// ServerMessage captures whatever
// world.Recent.LastServerMessage held immediately after the
// failing open attempt — typically the server's "you need a key"
// / "for members only" prose, so routines can branch on the text.
// Empty if no relevant message was observed in the window.
type DoorLockedError struct {
	DoorX, DoorY, DoorDir int
	Attempts              int
	ServerMessage         string
}

func (e *DoorLockedError) Error() string {
	if e.ServerMessage != "" {
		return fmt.Sprintf("door locked at (%d, %d) dir=%d after %d attempt(s): %s",
			e.DoorX, e.DoorY, e.DoorDir, e.Attempts, e.ServerMessage)
	}
	return fmt.Sprintf("door locked at (%d, %d) dir=%d after %d attempt(s)",
		e.DoorX, e.DoorY, e.DoorDir, e.Attempts)
}

// WalkOptions tunes WalkTo behavior. Zero value = sensible defaults
// (currently: attempt to open any closed door blocking the path).
// Construct via DefaultWalkOptions() and override fields rather than
// initializing directly, so future field additions don't break callers.
type WalkOptions struct {
	// AttemptOpenDoors, when true, makes WalkTo try to open an
	// adjacent openable boundary (door / doorframe) on stall and
	// retry the walk. Mirrors the Java RSC client's auto-door
	// behavior. Default: true. Set to false for routines that
	// want strict "stop at any obstacle" semantics (e.g. quest
	// checks that need to detect locked doors).
	//
	// If the door is truly locked (e.g. quest gate), the open
	// interaction succeeds at the packet layer but the host
	// can't pass; on a second stall at the same tile WalkTo
	// stops trying and returns PATH_BLOCKED with the stall pos
	// so the script can react.
	AttemptOpenDoors bool
}

// DefaultWalkOptions returns a WalkOptions with defaults applied:
// attempt-open-doors enabled. Callers that want non-default behavior
// should start here and tweak.
func DefaultWalkOptions() WalkOptions {
	return WalkOptions{AttemptOpenDoors: true}
}

// WalkTo navigates to (x, y) using the BFS pathfinder. Wraps
// WalkToOpts with default options.
func (h *Host) WalkTo(ctx context.Context, x, y int) error {
	return h.WalkToOpts(ctx, x, y, DefaultWalkOptions())
}

// WalkToOpts is WalkTo with explicit options. The DSL `walk_to`
// builtin routes through here, exposing the options as named args.
//
// The flow mirrors the reference clients over a grid that models CLOSED
// doors and gates as walls (the server's own collision rule):
//
//	plan (BFS) → walk → on no-path or stall: CLASSIFY the adjacent
//	obstacle (wall-door vs scenery-gate, straddle-ordered toward the
//	goal) → open via the obstacle's own channel → CONFIRM BY STATE
//	(def swap in the live store / position advance / plane change) →
//	replan from the new world.
//
// There is deliberately NO blind direct-walk fallback: a no-path with no
// openable obstacle is a real PATH_BLOCKED the caller must replan around
// (the old raw far-target Walk straight-lined into walls — the server
// stops at the first wall and never detours).
func (h *Host) WalkToOpts(ctx context.Context, x, y int, opts WalkOptions) error {
	const (
		pollInterval = 200 * time.Millisecond
		stallTimeout = 5 * time.Second
		arriveRadius = 1
	)
	// Per-obstacle open attempts + captured server prose, keyed by the
	// obstacle's identifying tile+dir. Caps re-tries on a locked door so
	// the walk fails loudly (DoorLockedError) instead of spinning.
	attempts := map[[3]int]int{}
	messages := map[[3]int]string{}
	for {
	replan:
		if err := ctx.Err(); err != nil {
			return err
		}
		pos := h.world.Self.Position()
		if absVal(x-pos.X) <= arriveRadius && absVal(y-pos.Y) <= arriveRadius {
			h.log.Debug("walkto arrived", "pos", fmt.Sprintf("(%d, %d)", pos.X, pos.Y))
			return nil
		}

		corners, pathErr := h.pathToTile(x, y, false)
		if pathErr != nil {
			if errors.Is(pathErr, ErrNoPath) {
				// The grid models closed doors/gates as walls, so a
				// no-path very often means a traversable obstacle sits
				// between us and the goal. Classify and open it, then
				// replan over the updated world.
				if opts.AttemptOpenDoors {
					opened, terr := h.tryTraverse(ctx, x, y, attempts, messages)
					if terr != nil {
						return terr
					}
					if opened {
						goto replan
					}
				}
				return fmt.Errorf("walkto: %w (no route from (%d, %d) to (%d, %d))", pathErr, pos.X, pos.Y, x, y)
			}
			return fmt.Errorf("walkto: pathfind: %w", pathErr)
		}
		if len(corners) == 0 {
			return nil // already at target
		}
		h.log.Debug("walkto path",
			"from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
			"target", fmt.Sprintf("(%d, %d)", x, y),
			"corners", len(corners),
			"final_corner", fmt.Sprintf("(%d, %d)", corners[len(corners)-1][0], corners[len(corners)-1][1]),
		)
		if err := action.WalkPath(ctx, h.conn, corners); err != nil {
			return fmt.Errorf("walkto: send path: %w", err)
		}

		// Wait for arrival OR stall (no position change for stallTimeout).
		startPos := pos
		stallDeadline := time.Now().Add(stallTimeout)
		arrived := false
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
			}
			cur := h.world.Self.Position()
			if absVal(cur.X-x) <= arriveRadius && absVal(cur.Y-y) <= arriveRadius {
				arrived = true
				break
			}
			if cur.X != startPos.X || cur.Y != startPos.Y {
				startPos = cur
				stallDeadline = time.Now().Add(stallTimeout)
				continue
			}
			if time.Now().After(stallDeadline) {
				// Stalled mid-walk: the server refused a step the grid
				// thought open (a door someone closed under us, or live
				// state we haven't seen). Same recovery as no-path.
				if opts.AttemptOpenDoors {
					opened, terr := h.tryTraverse(ctx, x, y, attempts, messages)
					if terr != nil {
						return terr
					}
					if opened {
						goto replan
					}
				}
				h.log.Warn("walkto stalled",
					"at", fmt.Sprintf("(%d, %d)", cur.X, cur.Y),
					"target", fmt.Sprintf("(%d, %d)", x, y),
				)
				return fmt.Errorf("walkto: stalled at (%d, %d) targeting (%d, %d)", cur.X, cur.Y, x, y)
			}
		}
		if arrived {
			return nil
		}
		// Loop to replan from the new position (server truncated the
		// path short of the target — live state will explain why).
	}
}

// tryTraverse classifies the actionable obstacle nearest the host's intended
// step toward (x, y), opens it via its own channel (boundary interaction for
// wall doors, scenery interaction for gates), and confirms BY STATE. Returns
// opened=true when an obstacle was confirmed traversed (caller replans), a
// DoorLockedError when an obstacle exhausted its attempts (locked / toll /
// quest-gated — the server prose is attached), or (false, nil) when nothing
// actionable is adjacent.
func (h *Host) tryTraverse(ctx context.Context, x, y int, attempts map[[3]int]int, messages map[[3]int]string) (bool, error) {
	const (
		maxOpenAttempts = 2
		// totalOpenBudget bounds opens across one WalkToOpts call so a
		// pathological corridor of doors can't spin forever.
		totalOpenBudget = 6
	)
	total := 0
	for _, n := range attempts {
		total += n
	}
	if total >= totalOpenBudget {
		return false, nil
	}
	cur := h.world.Self.Position()
	// Adjacent obstacles first (stall recovery, straddle-ordered); when
	// nothing is within reach — the plan-time no-path case, where the
	// blocking door may be far from BOTH host and goal — scan the grid
	// window for openable barriers ON the route.
	cands := h.findTraversableNear(cur.X, cur.Y, x, y)
	if len(cands) == 0 {
		cands = h.findTraversableToward(x, y)
	}
	for _, t := range cands {
		if t.Kind == TraversableLadder {
			continue // plane transitions are routed deliberately, never auto-climbed
		}
		key := [3]int{t.X, t.Y, t.Dir}
		if attempts[key] >= maxOpenAttempts {
			// Exhausted: locked / toll / quest gate. Fail loudly with the
			// captured server prose so cognition can plan (pay, fetch a
			// key, route around) instead of spinning.
			h.log.Warn("walkto: obstacle locked",
				"kind", t.Kind, "name", t.Name,
				"at", fmt.Sprintf("(%d, %d, dir=%d)", t.X, t.Y, t.Dir),
				"attempts", attempts[key],
				"server_msg", messages[key],
			)
			return false, &DoorLockedError{
				DoorX:         t.X,
				DoorY:         t.Y,
				DoorDir:       t.Dir,
				Attempts:      attempts[key],
				ServerMessage: messages[key],
			}
		}
		attempts[key]++

		// Snapshot the latest server message BEFORE the open so a
		// rejection ("the door is locked", "you must pay a toll of 10
		// gold coins") arriving in the confirm window is attributable.
		var preMsgAt time.Time
		if prev := h.world.Recent.ServerMessage(); prev != nil {
			preMsgAt = prev.At
		}
		h.log.Info("walkto: opening obstacle in path",
			"kind", t.Kind, "name", t.Name,
			"at", fmt.Sprintf("(%d, %d, dir=%d)", t.X, t.Y, t.Dir),
			"attempt", attempts[key],
		)
		if err := t.Open(ctx); err != nil {
			h.log.Warn("walkto: open dispatch failed", "err", err)
			continue
		}
		h.awaitTraversalConfirm(ctx, t, cur)
		if msg := h.world.Recent.ServerMessage(); msg != nil && msg.At.After(preMsgAt) {
			messages[key] = msg.Message
		}
		// Replan REGARDLESS of confirmation: the BFS re-reads the live
		// stores, so a confirmed swap routes through immediately, and an
		// unconfirmed (locked) obstacle hits no-path again — one retry,
		// then the attempts cap surfaces DoorLockedError with the prose.
		return true, nil
	}
	return false, nil
}

// awaitTraversalConfirm polls (~2.5s) for STATE evidence that an opened
// obstacle is traversed/traversable — never "the packet was sent":
//
//   - wall door:    the live boundary def at (X,Y,Dir) swapped to a
//     non-blocking def (the authentic wire NEVER removes in-range
//     boundaries — open == ID overwrite to e.g. doorframe 11);
//   - scenery gate: the live scenery def at the loc CHANGED from the
//     closed id (gate 60 → open 59) or the object was removed;
//   - any:          the host's position advanced (doDoor walks-then-
//     teleports THROUGH; mandatory signal — doDoor(replaceID=-1)
//     streams no def swap at all), or the plane band changed.
func (h *Host) awaitTraversalConfirm(ctx context.Context, t Traversable, from world.Coord) bool {
	const (
		confirmWindow = 2500 * time.Millisecond
		poll          = 200 * time.Millisecond
	)
	deadline := time.Now().Add(confirmWindow)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(poll):
		}
		cur := h.world.Self.Position()
		// Position-advance only confirms when the obstacle was ADJACENT
		// at open time: for a far obstacle the Open's own approach-walk
		// moves the host, which proves nothing about the obstacle.
		wasAdjacent := absVal(from.X-t.X) <= 1 && absVal(from.Y-t.Y) <= 1
		if wasAdjacent && (cur.X != from.X || cur.Y != from.Y) {
			return true // moved — the open walked/teleported us through
		}
		if cur.Y/world.PlaneHeight != from.Y/world.PlaneHeight {
			return true // plane band changed (ladder/stairs)
		}
		switch t.Kind {
		case TraversableWallDoor:
			if h.world.Boundaries != nil {
				if id, ok := h.world.Boundaries.Get(t.X, t.Y, t.Dir); ok {
					if id < 0 {
						return true // removed (defensive)
					}
					if d := h.facts.BoundaryDef(id); d != nil && !d.BlocksWhenClosed() && !d.BlocksMovement() {
						return true // swapped to the open def
					}
				}
			}
		case TraversableSceneryGate:
			if h.world.Scenery != nil {
				if rec, ok := h.world.Scenery.At(t.X, t.Y); ok && rec.ID != t.ClosedID {
					return true // def swap (closed gate -> open gate)
				}
				if h.world.Scenery.IsRemoved(t.X, t.Y) {
					return true
				}
			}
		}
		if time.Now().After(deadline) {
			return false
		}
	}
}

func absVal(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Logout is a high-level convenience that proxies to action.Logout.
// It sends a single logout request and does NOT wait for the server
// to acknowledge — use LogoutGraceful when you need the session
// actually released (e.g., before a fast re-login of the same
// account). The OpenRSC server refuses logout while the player is
// in combat / dueling / busy, or within ~10s of combat
// (Player.canLogout), so a fire-and-forget Logout is frequently
// ignored by the server.
func (h *Host) Logout(ctx context.Context) error {
	return action.Logout(ctx, h.conn)
}

// LogoutGraceful sends a logout request and waits for the server to
// actually release the session, retrying periodically until released
// or maxWait elapses.
//
// Why this exists: the server's Player.canLogout() returns false
// while the host is inCombat()/dueling/busy and for ~10s after the
// combat timer. A scenario that ends mid-combat therefore cannot log
// out immediately; if the cradle just hard-closes the TCP socket,
// the server holds the session until its connection reaper fires,
// and a same-account re-login in that window is rejected with login
// code 4 ("already logged in"). Riding out the cooldown and getting
// the server to release the session frees it immediately.
//
// Signal: the server does NOT send a logout-confirm packet on a clean
// logout — when it accepts (canLogout()==true), it unregisters the
// player and closes the socket. The session read loop turns that EOF
// into conn.Done() closing. So "logout accepted" == conn.Done() fired.
// When logout is REFUSED (in combat), the server silently keeps the
// connection open; we just retry on the next tick until the cooldown
// passes and a resend is accepted.
//
// Returns nil once the connection is released. Returns an error if
// maxWait passes with the connection still open (caller should
// hard-close anyway — the reaper will eventually release it). Honors
// ctx cancellation.
func (h *Host) LogoutGraceful(ctx context.Context, maxWait time.Duration) error {
	if h.conn == nil {
		return nil
	}
	deadline := time.Now().Add(maxWait)
	const retryEvery = 1200 * time.Millisecond
	for {
		// (Re)send the logout request. A transient write error is
		// ignored — either the connection just closed (we'll catch it
		// on the next select via conn.Done) or we retry next tick.
		_ = action.Logout(ctx, h.conn)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-h.conn.Done():
			// Server released the session and closed the socket.
			return nil
		case <-time.After(retryEvery):
			if !time.Now().Before(deadline) {
				return fmt.Errorf("logout not accepted within %s (connection still open — host likely in combat / busy)", maxWait)
			}
		}
	}
}

// Command sends an admin command (without the leading "::").
func (h *Host) Command(ctx context.Context, cmd string) error {
	return action.Command(ctx, h.conn, cmd)
}

// Say sends a public chat message (RSC-compressed under the hood). On a
// successful send it captures the line into the reactive tier's latched windows
// (the single self-line seam — all callers including socialReflex funnel here),
// so a Q&A pairs up when the host replies to a latched speaker.
func (h *Host) Say(ctx context.Context, message string) error {
	if err := action.Say(ctx, h.conn, message); err != nil {
		return err
	}
	h.reactiveObserveSelf(message)
	return nil
}

// Close shuts down the host: first a BEST-EFFORT clean RSC logout (so the server
// saves + releases the session instead of timing out a dropped socket — which
// otherwise blocks a same-account re-login with "already logged in"), then the
// socket and the event bus. LogoutGraceful rides out the combat-logout cooldown
// up to its bound, and is a no-op when the connection is already gone.
func (h *Host) Close() error {
	if h.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		_ = h.LogoutGraceful(ctx, 12*time.Second)
		cancel()
		h.conn.Close()
	}
	h.bus.Close()
	return nil
}
