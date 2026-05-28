package runtime

import (
	"context"
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
func (h *Host) WalkTo(ctx context.Context, x, y int) error {
	const (
		segmentRange  = action.MaxClickRange - 1 // one tile of margin
		pollInterval  = 200 * time.Millisecond
		stallTimeout  = 5 * time.Second
		arriveRadius  = 1
	)
	for {
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

		// Choose an in-FOV sub-target in the direction of (x, y).
		stepX := pos.X + clamp(dx, -segmentRange, segmentRange)
		stepY := pos.Y + clamp(dy, -segmentRange, segmentRange)
		h.log.Debug("walkto segment",
			"from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
			"to_segment", fmt.Sprintf("(%d, %d)", stepX, stepY),
			"final_target", fmt.Sprintf("(%d, %d)", x, y),
		)
		if err := action.Walk(ctx, h.conn, pos.X, pos.Y, stepX, stepY); err != nil {
			return fmt.Errorf("walkto segment: %w", err)
		}

		// Wait for progress: position must change within stallTimeout.
		startPos := pos
		stallDeadline := time.Now().Add(stallTimeout)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
			}
			cur := h.world.Self.Position()
			if cur.X != startPos.X || cur.Y != startPos.Y {
				// Progress made; continue outer loop to plan next segment.
				break
			}
			if time.Now().After(stallDeadline) {
				h.log.Warn("walkto stalled",
					"at", fmt.Sprintf("(%d, %d)", cur.X, cur.Y),
					"target", fmt.Sprintf("(%d, %d)", x, y),
				)
				return fmt.Errorf("walkto: stalled at (%d, %d) targeting (%d, %d)", cur.X, cur.Y, x, y)
			}
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
