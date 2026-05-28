package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
	"github.com/gen0cide/westworld/world"
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

	Logger            *slog.Logger
	HeartbeatInterval time.Duration
	EventBufferSize   int
}

// Host is one bot's entire runtime: connection, world state, event
// bus, and the main event loop that ties them together.
type Host struct {
	opts Options

	conn      *session.Conn
	world     *world.World
	bus       *event.Bus
	facts     *facts.Facts
	landscape *pathfind.Landscape
	log       *slog.Logger

	// Strategist + Retriever are the cognition+brain layer hooks
	// used by the routine builtins contemplate_reality/decide/
	// evaluate/recall/etc. Defaults: stub implementations from
	// brain/ and cognition/ that return deterministic canned
	// values. Phase 3+ replaces these with real implementations
	// (mesa retrieval + Anthropic LLM call). Hosts share interfaces
	// safely across goroutines — one instance per process is fine.
	Strategist brain.Strategist
	Retriever  cognition.Client

	loggedIn bool
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
	return &Host{
		opts:      opts,
		world:     world.NewWorld(),
		bus:       event.NewBus(),
		facts:     opts.Facts,
		landscape: opts.Landscape,
		log:       opts.Logger,
		// Stub strategist + retriever by default. Production
		// wiring overrides these with real implementations
		// after Phase 3/4 land.
		Strategist: &brain.StubStrategist{},
		Retriever:  &cognition.StubClient{},
	}
}

// Facts returns the host's shared knowledge base (may be nil if no
// Facts were passed in opts).
func (h *Host) Facts() *facts.Facts { return h.facts }

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
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InNpcCoords:
		pos := h.world.Self.Position()
		events, err := v235.DecodeNpcCoords(f.Payload, pos.X, pos.Y)
		if err != nil {
			h.log.Warn("decode npccoords", "err", err)
			return
		}
		for _, ev := range events {
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	}

	// Single-event opcodes.
	ev, err := v235.DecodeInbound(f)
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
	changed := h.world.Apply(ev)
	if changed {
		h.log.Debug("world updated", "by", ev.Kind())
	}
	h.bus.Publish(ev)
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
func (h *Host) WalkToOpts(ctx context.Context, x, y int, opts WalkOptions) error {
	const (
		pollInterval = 200 * time.Millisecond
		stallTimeout = 5 * time.Second
		arriveRadius = 1
		// maxDoorAttempts caps re-tries on the same door to avoid
		// infinite loops when the door is locked or the open
		// interaction silently fails. Two attempts is enough to
		// recover from the rare race where the door re-closed
		// between our open and our re-walk.
		maxDoorAttempts = 2
	)
	// Track door-open attempts keyed by boundary tile so a single
	// locked door can't burn cycles forever.
	doorAttempts := map[[2]int]int{}
	// Outer loop replans when the previous WalkPath finishes
	// short (server-side path truncation, e.g. blocked by a
	// closed door we haven't opened). Each iteration pathfinds
	// fresh from the current position so a moving obstacle or
	// dynamic boundary state is picked up.
	for {
	replan:
		if err := ctx.Err(); err != nil {
			return err
		}
		pos := h.world.Self.Position()
		dx := x - pos.X
		dy := y - pos.Y
		if absVal(dx) <= arriveRadius && absVal(dy) <= arriveRadius {
			h.log.Debug("walkto arrived", "pos", fmt.Sprintf("(%d, %d)", pos.X, pos.Y))
			return nil
		}

		// Pathfind through the static landscape + facts grid.
		// reachBorder=false because the caller asked for a tile,
		// not "stand-adjacent-to" semantics (which is for
		// attack/talk_to/open). On no-path, fail fast with a
		// clear error rather than sending a doomed straight walk.
		corners, pathErr := h.pathToTile(x, y, false)
		if pathErr != nil {
			if errors.Is(pathErr, ErrNoPath) {
				return fmt.Errorf("walkto: %w (no route from (%d, %d) to (%d, %d))", pathErr, pos.X, pos.Y, x, y)
			}
			return fmt.Errorf("walkto: pathfind: %w", pathErr)
		}
		if len(corners) == 0 {
			// Already at target.
			return nil
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

		// Wait for arrival OR stall (no position change for
		// stallTimeout). Each tick polls Self.Position which is
		// updated by the inbound packet handler.
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
				// Progress made — extend the stall deadline and
				// keep waiting for either arrival or a fresh
				// stall (the path is still in flight server-side
				// since the server interpolates between corners).
				startPos = cur
				stallDeadline = time.Now().Add(stallTimeout)
				continue
			}
			if time.Now().After(stallDeadline) {
				// Stalled. If attempt-open-doors is enabled and
				// there's an openable boundary near our current
				// position, try opening it once (or twice — see
				// maxDoorAttempts) and let the outer loop replan
				// and re-walk. This handles both the static case
				// (closed door on our path from the start) and
				// the dynamic case (a door someone closed in
				// front of us while we were walking).
				if opts.AttemptOpenDoors {
					if door := h.findOpenableNear(cur.X, cur.Y); door != nil {
						key := [2]int{door.X, door.Y}
						if doorAttempts[key] < maxDoorAttempts {
							doorAttempts[key]++
							h.log.Info("walkto: stalled at door, attempting to open",
								"door", fmt.Sprintf("(%d, %d, dir=%d)", door.X, door.Y, door.Direction),
								"attempt", doorAttempts[key],
							)
							if err := h.InteractWithBoundary(ctx, door.X, door.Y, door.Direction); err != nil {
								h.log.Warn("walkto: open door failed", "err", err)
								// Fall through to PATH_BLOCKED.
							} else {
								// Give the server a beat to apply the
								// open (it's a single tick at ~640ms;
								// 800ms is conservative).
								select {
								case <-ctx.Done():
									return ctx.Err()
								case <-time.After(800 * time.Millisecond):
								}
								// Break inner stall loop; outer loop
								// replans from new position (the
								// open packet itself may have walked
								// us to the door).
								goto replan
							}
						}
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
		// Loop to replan from new position (server truncated path
		// short of target — usually a closed boundary or blocked
		// scenery the grid didn't account for).
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
func (h *Host) Logout(ctx context.Context) error {
	return action.Logout(ctx, h.conn)
}

// Command sends an admin command (without the leading "::").
func (h *Host) Command(ctx context.Context, cmd string) error {
	return action.Command(ctx, h.conn, cmd)
}

// Say sends a public chat message (RSC-compressed under the hood).
func (h *Host) Say(ctx context.Context, message string) error {
	return action.Say(ctx, h.conn, message)
}

// Close shuts down the underlying session and event bus.
func (h *Host) Close() error {
	if h.conn != nil {
		h.conn.Close()
	}
	h.bus.Close()
	return nil
}
