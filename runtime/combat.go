package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/pathfind"
)

// AttackNpc initiates combat with an NPC at the given server index.
// Pathfinds from the bot's current position to a tile adjacent to the
// NPC, sends the multi-corner walk packet, then the attack opcode —
// matching mudclient.walkToArea(..., walkToEntity=true) per click.
//
// Falls back to sending a bare attack packet if the NPC isn't in
// the world-state mirror yet (e.g., the bot just logged in and hasn't
// received an NpcCoords update).
func (h *Host) AttackNpc(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("AttackNpc: NPC not in world state, sending bare action", "server_index", serverIndex)
		return action.AttackNpc(ctx, h.conn, serverIndex)
	}
	// If we're standing on the NPC's exact tile (a "stacked" target, e.g.
	// an NPC that walked onto us or one spawned on our tile), the server
	// cannot melee it — there's no adjacent attack square — and pathToTile
	// reports "already there" so walkAndAct would skip the walk and fire a
	// bare, dropped attack. The authentic client steps to a neighbouring
	// tile and then attacks. Mirror that: step off first, then fall
	// through to the normal walk-then-attack so we approach from adjacent.
	if pos := h.world.Self.Position(); pos.X == rec.X && pos.Y == rec.Y {
		h.stepOffNpcTile(ctx, rec.X, rec.Y)
		// Re-read the NPC tile: it may have moved while we stepped.
		if r2, ok2 := h.world.Npcs.Get(serverIndex); ok2 {
			rec = r2
		}
	}
	h.log.Info("AttackNpc: pathfinding",
		"server_index", serverIndex,
		"npc_type_id", rec.TypeID,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	// Engage-with-retry: the FIRST attack packet after a fresh approach is
	// often dropped server-side — the target may still be settling (a just-
	// spawned NPC is briefly attack-gated, WalkToMobAction can lose the
	// race), so the bot just stands adjacent taking hits without engaging.
	// The authentic client re-clicks until combat starts. Mirror that: send
	// the walk+attack, then confirm engagement; if unconfirmed, re-send (up
	// to a small cap). We re-send ONLY while NOT engaged, so an active fight
	// is never interrupted (re-attacking mid-combat just earns a harmless
	// "you are already busy fighting" and a re-pathfind that resets the
	// combat event — the very stall we're avoiding).
	const maxEngageAttempts = 4
	var lastErr error
	for attempt := 0; attempt < maxEngageAttempts; attempt++ {
		lastErr = h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
			return action.AttackNpc(ctx, h.conn, serverIndex)
		})
		if lastErr != nil {
			return lastErr
		}
		if h.confirmEngaged(ctx, serverIndex) {
			return nil
		}
		// Not engaged — re-read the target tile (it may have drifted) and
		// retry the approach+attack.
		if r2, ok2 := h.world.Npcs.Get(serverIndex); ok2 {
			rec = r2
		} else {
			// NPC left our view entirely (killed by us between sends, or
			// despawned). Nothing to re-attack.
			return nil
		}
	}
	return lastErr
}

// confirmEngaged polls for evidence that WE are actively attacking
// serverIndex (not merely being attacked by it). The only reliable
// signals that our attack landed are: the target took damage (its
// HasHits became true — a hit we dealt populates its health bar), or the
// target is gone (we killed/removed it). We deliberately DON'T treat
// "our own HP dropped" as engagement: a goblin can reach us and attack
// FIRST, making us a passive victim that never swings back, while our
// own attack packet was dropped ("you are already busy fighting"). In
// that state we take damage but deal none — exactly the stall this retry
// exists to break — so counting incoming damage as "engaged" would
// wrongly suppress the re-attack and leave us a punching bag. Returns
// false if no outgoing-damage / kill signal appears within the window,
// signalling the attack should be re-sent. Bounded so it never hangs.
func (h *Host) confirmEngaged(ctx context.Context, serverIndex int) bool {
	const (
		poll   = 250 * time.Millisecond
		window = 5 * time.Second
	)
	// Require having SEEN the target in-roster at least once before a
	// later disappearance counts as "killed". A freshly-spawned NPC can
	// briefly drop from the local roster (REMOVE+re-add as it settles /
	// moves); treating that transient gap as a kill would falsely confirm
	// engagement and suppress the re-attack. Only a seen→gone transition
	// is a real kill signal.
	seen := false
	deadline := time.Now().Add(window)
	for {
		// An ALREADY-ACTIVE fight must NEVER be re-attacked: re-pathfinding
		// mid-combat resets the server-side combat event (the bot walks out
		// of melee and re-approaches), so the fight never resolves — the
		// regression this guard prevents.
		//
		// The authentic signal that combat is live is the anti-kite round
		// counter advancing on THIS target. In RSC melee both combatants
		// auto-swing every round once a combat event is paired, so taking
		// the target's hits at round cadence (OtherPlayerDamage on self)
		// OR landing our own (NpcDamage on its index) equally proves we are
		// in a live, self-resolving fight — we will swing each round even
		// if we keep MISSING (zero outgoing damage). Requiring an outgoing
		// HIT here is wrong: a real fight where we miss several rounds would
		// be falsely re-attacked (resetting the combat event) and stall. A
		// genuinely-idle bot (attack packet dropped, no pairing) takes no
		// hits at all, so combatRounds stays 0 and we correctly re-send.
		if h.combatRoundTarget == serverIndex && h.combatRounds > 0 {
			return true
		}
		rec, ok := h.world.Npcs.Get(serverIndex)
		if ok {
			seen = true
			if rec.HasHits && rec.MaxHits > 0 && rec.CurHits < rec.MaxHits {
				// The target has taken real damage (below full) — our hit
				// landed, so we ARE the aggressor. A full-health bar can show
				// merely because the NPC is in combat (attacking us), which is
				// NOT proof we're swinging — so require strictly-below-max.
				return true
			}
		} else if seen {
			return true // target was here and is now gone — killed
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(poll):
		}
	}
}

// stepOffNpcTile moves the bot one tile off (npcX, npcY) — the tile it is
// currently stacked on — to an adjacent standable square, so the
// subsequent attack can approach the NPC from a legal adjacent tile.
// Best-effort: it picks the first walkable cardinal neighbour (falling
// back to diagonals), sends a single in-FOV walk, and waits briefly for
// arrival. If no neighbour is standable or the landscape isn't loaded it
// returns without moving — the caller still attempts the attack.
func (h *Host) stepOffNpcTile(ctx context.Context, npcX, npcY int) {
	pos := h.world.Self.Position()
	// Cardinal first (cleaner approach square), then diagonals.
	cand := [][2]int{
		{pos.X + 1, pos.Y}, {pos.X - 1, pos.Y},
		{pos.X, pos.Y + 1}, {pos.X, pos.Y - 1},
		{pos.X + 1, pos.Y + 1}, {pos.X - 1, pos.Y - 1},
		{pos.X + 1, pos.Y - 1}, {pos.X - 1, pos.Y + 1},
	}
	var grid *pathfind.Grid
	if h.landscape != nil {
		if g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0); err == nil {
			grid = g
		}
	}
	for _, c := range cand {
		if c[0] == npcX && c[1] == npcY {
			continue // don't step onto the NPC
		}
		if grid != nil && !grid.TileStandable(c[0], c[1]) {
			continue
		}
		if err := h.Walk(ctx, c[0], c[1]); err != nil {
			continue
		}
		h.awaitArrival(ctx, c[0], c[1])
		return
	}
}

// AttackPlayer initiates PVP with a player. Pathfinds to the
// player's last-known position then sends the attack packet.
func (h *Host) AttackPlayer(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	}
	h.log.Info("AttackPlayer: pathfinding",
		"server_index", serverIndex,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	})
}

// ChooseDialogOption replies to the server's most recent
// SendNpcDialogOptions prompt by picking the option at `index`. The
// 0-based index must match the order of Options in the
// event.NpcDialog the bot received.
func (h *Host) ChooseDialogOption(ctx context.Context, index int) error {
	return action.ChooseDialogOption(ctx, h.conn, index)
}

// SetCombatStyle changes the melee xp-split. Takes effect on the
// next attack tick — RSC doesn't acknowledge the change with a
// dedicated packet, so observe self.skills.<style>.xp deltas if
// you need to confirm it was applied.
func (h *Host) SetCombatStyle(ctx context.Context, style action.CombatStyle) error {
	return action.SetCombatStyle(ctx, h.conn, style)
}

// retreatRoundGate is the authentic anti-kite threshold: RSC forbids
// retreating until the opponent has made >= 3 hits — the "first 3
// rounds of combat" (OpenRSC WalkRequest.java checks
// opponent.getHitsMade() >= 3).
const retreatRoundGate = 3

// combatRoundTick is the wall-clock duration of one RSC combat round.
// Used only as a fallback estimate when we want to wait out the gate
// but have observed no damage events yet (e.g. both sides missing). One
// game tick is ~640ms; combat rounds resolve one tick apart.
const combatRoundTick = 640 * time.Millisecond

// retreatRejectedMarker is the substring of the server's anti-kite
// rejection ("You can't retreat during the first 3 rounds of combat",
// WalkRequest.java). We match a lowercased substring so a server with a
// slightly reworded message ("...for another N rounds.") still
// classifies as RETREAT_TOO_EARLY.
const retreatRejectedMarker = "retreat"

// RetreatTooEarlyError is the typed error Host.Retreat / RetreatTo
// returns when the server rejected the flee for the anti-kite rule. It
// carries the exact server prose so a routine can surface it verbatim.
type RetreatTooEarlyError struct {
	ServerMessage string
	RoundsSeen    int
}

func (e *RetreatTooEarlyError) Error() string {
	if e.ServerMessage != "" {
		return e.ServerMessage
	}
	return fmt.Sprintf("can't retreat yet (only %d of %d combat rounds elapsed)",
		e.RoundsSeen, retreatRoundGate)
}

// Retreat breaks melee combat by walking one tile away — the only
// mechanic the authentic v235 protocol (and the GUI) offers for
// disengaging. Verified against OpenRSC WalkRequest.java: there is no
// dedicated "stop fighting" opcode; sending a WALK_TO_POINT while
// inCombat() triggers player.resetCombatEvent() server-side — BUT ONLY
// after the opponent has made >= 3 hits. An earlier WALK_TO_POINT is
// rejected and the server emits "You can't retreat during the first 3
// rounds of combat".
//
// Round-awareness (#r3-retreat): waitGate controls whether we
// pre-emptively wait out the 3-round gate before sending. With
// waitGate=true (the DSL default / production path) we wait for the
// rounds and, if the server still rejects, wait-and-retry until it
// allows the flee — so retreat reliably succeeds once 3 rounds elapse.
// With waitGate=false we make one raw attempt and return a typed
// RetreatTooEarlyError carrying the server prose, so a routine can
// poll-and-branch instead of believing it fled.
//
// Direction: step directly away from the engaged target when we can
// resolve its tile (so we actually open distance); otherwise step one
// tile south as a deterministic fallback. We send a single in-FOV walk
// (Host.Walk = WALK_TO_POINT) rather than a full BFS route — retreat is
// "take one step out of melee range now".
func (h *Host) Retreat(ctx context.Context, waitGate bool) error {
	pos := h.world.Self.Position()
	dx, dy := 0, 1 // fallback: one tile south

	// Prefer stepping away from the current engagement target so we
	// open real distance. Resolve the target's tile from the same
	// sources combat.target uses (own-slot engagement, then last
	// attacked entity).
	if tx, ty, ok := h.engagedTargetTile(); ok {
		sx := sign(pos.X - tx)
		sy := sign(pos.Y - ty)
		if sx == 0 && sy == 0 {
			// Stacked on the target (shouldn't happen in melee) —
			// keep the south fallback.
		} else {
			dx, dy = sx, sy
		}
	}
	tx, ty := pos.X+dx, pos.Y+dy
	return h.retreatWalk(ctx, waitGate, func(ctx context.Context) error {
		// Re-read position at send time — waiting for the round gate
		// may have let us drift, and Host.Walk encodes the delta from
		// our current tile. Recompute the one-step-away target.
		cur := h.world.Self.Position()
		return h.Walk(ctx, cur.X+(tx-pos.X), cur.Y+(ty-pos.Y))
	})
}

// RetreatTo flees toward a specific safe tile once retreat is allowed
// (#r3-retreat). It shares the 3-round anti-kite gate and the typed
// RetreatTooEarlyError rejection with Retreat: it waits out the gate,
// sends the breaking WALK_TO_POINT (the first leg of the route to the
// safe tile), and — only if that wasn't rejected — pathfinds the rest
// of the way with WalkTo. Effectively a combat-aware walk_to.
func (h *Host) RetreatTo(ctx context.Context, x, y int, waitGate bool) error {
	if err := h.retreatWalk(ctx, waitGate, func(ctx context.Context) error {
		// The breaking step: a single WALK_TO_POINT toward the safe
		// tile. The server breaks combat on this packet (once the gate
		// is satisfied); Host.Walk caps to the in-FOV step the wire
		// allows, and the WalkTo below covers the remaining distance.
		return h.Walk(ctx, x, y)
	}); err != nil {
		return err
	}
	// Combat is broken — now route the rest of the way out. If we're
	// already at/adjacent to the target this is a cheap no-op.
	if pos := h.world.Self.Position(); pos.X == x && pos.Y == y {
		return nil
	}
	return h.WalkTo(ctx, x, y)
}

// maxRetreatRetries bounds the wait-and-retry loop used by the default
// (waitGate=true) retreat path. Each retry waits ~one combat round; the
// authentic gate is 3 rounds, so a handful of retries comfortably covers
// the window plus jitter. Past this cap we give up and surface the
// rejection so the caller isn't hung indefinitely.
const maxRetreatRetries = 6

// retreatWalk performs a round-aware breaking walk. When waitGate is
// true (the default DSL path) it first waits out the engine's round
// estimate, then sends; if the server still rejects with the anti-kite
// message it waits ~one round and retries, up to maxRetreatRetries —
// the SERVER'S message is the authoritative gate, so this guarantees
// the default retreat actually succeeds once the 3 rounds elapse rather
// than depending on the round-counter being perfectly accurate. When
// waitGate is false it makes exactly one raw attempt and returns the
// typed RetreatTooEarlyError on rejection (poll-and-branch path).
//
// A non-anti-kite send error (out-of-range / connection) is returned
// unchanged for the DSL layer to map.
func (h *Host) retreatWalk(ctx context.Context, waitGate bool, send func(context.Context) error) error {
	if !waitGate {
		return h.retreatSendOnce(ctx, send)
	}

	// Pre-emptively wait out the engine's round estimate so the common
	// path doesn't poke the server with a doomed walk.
	if h.combatRoundTarget != 0 {
		if err := h.waitForRetreatGate(ctx); err != nil {
			return err
		}
	}

	// Attempt, and if the server still says "too early" (round estimate
	// undershot), wait one round and retry. The server message is the
	// source of truth for the gate.
	var last error
	for attempt := 0; attempt < maxRetreatRetries; attempt++ {
		err := h.retreatSendOnce(ctx, send)
		if _, tooEarly := err.(*RetreatTooEarlyError); !tooEarly {
			return err // success (nil) or a real send error
		}
		last = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(combatRoundTick):
		}
	}
	return last
}

// retreatSendOnce makes a single breaking-walk attempt: snapshot the
// latest server-message timestamp, send, wait a beat, then classify any
// message that arrived in the window. A retreat rejection -> typed
// RetreatTooEarlyError; otherwise nil (success) or the send's own error.
func (h *Host) retreatSendOnce(ctx context.Context, send func(context.Context) error) error {
	var preMsgAt time.Time
	if prev := h.world.Recent.ServerMessage(); prev != nil {
		preMsgAt = prev.At
	}

	if err := send(ctx); err != nil {
		return err
	}

	// Give the server a beat to either break combat or emit the
	// rejection prose. One tick is ~640ms; 800ms is conservative.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(800 * time.Millisecond):
	}

	if cur := h.world.Recent.ServerMessage(); cur != nil && cur.At.After(preMsgAt) {
		if strings.Contains(strings.ToLower(cur.Message), retreatRejectedMarker) {
			return &RetreatTooEarlyError{ServerMessage: cur.Message, RoundsSeen: h.combatRounds}
		}
	}
	return nil
}

// waitForRetreatGate blocks until the engine has observed enough combat
// rounds (combatRounds >= retreatRoundGate) on the current engagement,
// the engagement ends, or a conservative timeout elapses. This lets
// retreat() flee on the FIRST allowed tick instead of polling the
// server with doomed walks. The server's rejection remains the
// authoritative check (retreatWalk reads it), so an over-eager wake is
// harmless — we just attempt and may still see RETREAT_TOO_EARLY.
func (h *Host) waitForRetreatGate(ctx context.Context) error {
	// Cap the wait so a stuck/observerless engagement can't hang the
	// routine: gate rounds at one tick each, plus generous slack for
	// tick jitter and the first attack landing.
	deadline := time.Now().Add(time.Duration(retreatRoundGate+3) * combatRoundTick)
	for {
		// Enough rounds seen -> allowed. Also bail if the engagement we
		// were counting has cleared (combat already over).
		if h.combatRounds >= retreatRoundGate || h.combatRoundTarget == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			// Timed out estimating rounds — proceed anyway and let the
			// server's message be the source of truth (retreatWalk maps
			// a rejection to RETREAT_TOO_EARLY).
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(80 * time.Millisecond):
		}
	}
}

// beginCombatRoundTracking (re)starts the anti-kite round counter for a
// freshly attacked target. Resets the count only when the target
// changes, so re-issuing attack() on the SAME target mid-fight (e.g.
// after it walked) doesn't zero progress toward the 3-round gate.
func (h *Host) beginCombatRoundTracking(targetIndex int) {
	if h.combatRoundTarget == targetIndex && !h.combatStartedAt.IsZero() {
		return // same engagement; keep the running count
	}
	h.combatRoundTarget = targetIndex
	h.combatRounds = 0
	h.outgoingHits = 0
	h.combatStartedAt = time.Now()
}

// engagedTargetTile returns the (x, y) of our current combat target if
// one can be resolved, mirroring combat.target's resolution order.
func (h *Host) engagedTargetTile() (int, int, bool) {
	if self, ok := h.world.Players.Self(); ok && !self.EngagedAt.IsZero() {
		if self.EngagedNpcIndex >= 0 {
			if rec, ok := h.world.Npcs.Get(self.EngagedNpcIndex); ok {
				return rec.X, rec.Y, true
			}
		}
		if self.EngagedPlayerIndex >= 1 {
			if rec, ok := h.world.Players.Get(self.EngagedPlayerIndex); ok {
				return rec.X, rec.Y, true
			}
		}
	}
	if idx := h.lastAttackedNpcIndex; idx != 0 {
		if rec, ok := h.world.Npcs.Get(idx); ok {
			return rec.X, rec.Y, true
		}
	}
	if idx := h.lastAttackedPlayerIndex; idx != 0 {
		if rec, ok := h.world.Players.Get(idx); ok {
			return rec.X, rec.Y, true
		}
	}
	return 0, 0, false
}

// sign returns -1, 0, or +1 for the sign of n.
func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

// TalkToNpc opens dialog with an NPC. Walks adjacent (reachBorder
// mode, mirroring "click NPC across a counter") then sends the
// talk-to packet.
func (h *Host) TalkToNpc(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("TalkToNpc: NPC not in world state, sending bare action", "server_index", serverIndex)
		return action.TalkToNpc(ctx, h.conn, serverIndex)
	}
	h.log.Info("TalkToNpc: pathfinding",
		"server_index", serverIndex,
		"npc_type_id", rec.TypeID,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.TalkToNpc(ctx, h.conn, serverIndex)
	})
}
