package worldmap

import "fmt"

// Reach-EXPLANATION: the cognition-first layer the host queries. Reachable /
// CanReach answer only "can I get there?" (a bool / a component set). They do
// NOT say WHY a destination is or isn't reachable — which gate sits on the
// route, what it costs, and whether the host can pay it. ExplainReach adds
// exactly that, so the WorldOracle can INFORM the host instead of silently
// deciding for it: the host SEES the toll gate + its 10-coin requirement +
// what it currently carries, then the brain chooses (pay, pick a free
// alternative, or go earn coins).
//
// This is the ONE permitted extension to the engine. It reuses the existing
// per-query flood (reachableTiles) verbatim for the host's real capability,
// and re-runs the SAME flood with synthetic gate-passability overrides only to
// pin down WHICH gate is binding — it never duplicates the walking predicate.

// ReachStatus is the per-destination verdict.
type ReachStatus string

const (
	// ReachOpen: the destination is reachable on foot/teleport with no
	// capability gate anywhere on the route — a free walk.
	ReachOpen ReachStatus = "open"
	// ReachGated: a gate lies on the only route, but the host's Capability
	// MEETS its Requirement (Payable=true), so the host can pass it (e.g. a
	// 15-coin host at the 10-coin toll gate).
	ReachGated ReachStatus = "gated"
	// ReachBlocked: the only route is across a gate whose Requirement the
	// host does NOT meet (Payable=false) — the host cannot reach the
	// destination as it stands.
	ReachBlocked ReachStatus = "blocked"
)

// ReachInfo is the explained verdict for ONE destination tile from ONE start
// tile under ONE Capability. It is the shape the DSL search_map / reachable
// handlers surface to the host.
type ReachInfo struct {
	// Reach is the open|gated|blocked verdict.
	Reach ReachStatus

	// Gate is the human name of the binding gate (empty when Reach==open).
	Gate string

	// Needs is the human-readable requirement of the binding gate, e.g.
	// "10 coins", "Fishing 68", "quest Prince Ali Rescue", "members". Empty
	// when Reach==open.
	Needs string

	// YouHave is what the host currently has against the binding gate's
	// dimension — its coin count for a coin gate, its skill level for a skill
	// gate, 1/0 for item/quest/members gates. 0 and meaningful only when a
	// gate is named.
	YouHave int

	// Payable reports whether the host's Capability satisfies the binding
	// gate (true for gated, false for blocked). Always false when open.
	Payable bool

	// req is the binding gate's raw Requirement (unexported; callers read the
	// formatted Needs / YouHave fields).
	req Requirement

	// snapOK reports whether the destination tile snapped to a standable
	// tile at all; false means the destination is off-map / void and the
	// other fields are zero.
	snapOK bool
}

// SnapOK reports whether the destination resolved to a standable tile. A
// false verdict means the coords are off-map or void; the host should treat it
// as "no such place" rather than a real reach result.
func (r ReachInfo) SnapOK() bool { return r.snapOK }

// ExplainReach returns the explained reach verdict for the destination tile
// (toX,toY) from the start tile (fromX,fromY) under cap. It:
//
//  1. snaps the destination to its nearest standable tile (like CanReach);
//  2. floods with the host's REAL capability — if the destination is in the
//     reachable set, the host can get there;
//  3. distinguishes open from gated by re-flooding with ALL gates treated as
//     impassable: if the destination is still reachable then no gate was on
//     the route (open); otherwise a gate the host satisfies was crossed
//     (gated), and the binding gate is whichever single gate, made passable on
//     its own, restores reachability;
//  4. for a blocked destination, the binding gate is whichever single
//     currently-unsatisfied gate, made passable on its own, would unlock it.
//
// The gate's Name + a formatted Needs string + YouHave (host's standing on the
// gate's dimension) + Payable (cap.Satisfies(req)) are returned so the host can
// reason about it without re-guessing.
func (o *Oracle) ExplainReach(fromX, fromY, toX, toY int, cap Capability) ReachInfo {
	_, tx, ty, ok := o.CompNear(toX, toY)
	if !ok {
		return ReachInfo{Reach: ReachBlocked, snapOK: false}
	}
	target := tx*o.dim + ty

	// (2) The host's real reachability.
	realSeen := o.reachableTilesGated(fromX, fromY, func(e *TransportEdge) bool {
		return cap.Satisfies(e.Req)
	})

	if realSeen[target] {
		// Reachable. Open vs gated: would the host still reach it with every
		// gate slammed shut? If yes, no gate was on the route → open.
		closedSeen := o.reachableTilesGated(fromX, fromY, func(e *TransportEdge) bool {
			return false // every gate impassable
		})
		if closedSeen[target] {
			return ReachInfo{Reach: ReachOpen, snapOK: true}
		}
		// A gate the host SATISFIES was crossed. Find which single gate, made
		// passable alone (all others shut), restores reachability — that's the
		// binding gate on this route.
		if e := o.bindingGate(fromX, fromY, target); e != nil {
			info := o.gateInfo(e, cap)
			info.Reach = ReachGated
			info.Payable = true // it's on the host's actual reachable route
			info.snapOK = true
			return info
		}
		// Couldn't isolate a single binding gate (multiple gates in series).
		// Still reachable for real, so report gated without a specific gate.
		return ReachInfo{Reach: ReachGated, Payable: true, snapOK: true}
	}

	// (4) Not reachable. Find the gate that, made passable alone, would unlock
	// it — the binding blocked gate.
	if e := o.bindingGate(fromX, fromY, target); e != nil {
		info := o.gateInfo(e, cap)
		info.Reach = ReachBlocked
		info.Payable = cap.Satisfies(e.Req) // false here (else it'd be reachable)
		info.snapOK = true
		return info
	}
	// No gate isolated — genuinely unreachable (water / no path / island with
	// no ferry the host can board). Report blocked with no nameable gate.
	return ReachInfo{Reach: ReachBlocked, snapOK: true}
}

// bindingGate returns the single barrier-bearing gate edge that, made passable
// on its own (every OTHER gate shut), makes `target` reachable from
// (fromX,fromY). It is how ExplainReach pins the SPECIFIC gate sitting on the
// boundary between the host's reachable region and the destination. Returns nil
// if no single gate isolates the target (the destination is unreachable even
// with one gate opened, or reachable already with all gates shut).
func (o *Oracle) bindingGate(fromX, fromY, target int) *TransportEdge {
	for i := range o.edges {
		e := &o.edges[i]
		if e.Kind != "gate" || e.Barrier == nil {
			continue // only barrier gates cut the flood in v1
		}
		open := e // capture
		seen := o.reachableTilesGated(fromX, fromY, func(c *TransportEdge) bool {
			return c == open // only THIS gate is passable
		})
		if seen[target] {
			return open
		}
	}
	return nil
}

// gateInfo builds the human-readable ReachInfo fields (Gate name, Needs,
// YouHave) for one gate edge under cap. Reach/Payable/snapOK are filled by the
// caller.
func (o *Oracle) gateInfo(e *TransportEdge, cap Capability) ReachInfo {
	return ReachInfo{
		Gate:    e.Name,
		Needs:   requirementText(e.Req),
		YouHave: youHave(e.Req, cap),
		req:     e.Req,
	}
}

// requirementText renders a Requirement as the short human phrase the host
// reads in `needs`, e.g. "10 coins", "Fishing 68", "quest Prince Ali Rescue",
// "members". The toll gate carries Coins plus a QuestFree free-pass clause; we
// surface the coin cost (the actionable cost) and mention the free-pass quest.
func requirementText(r Requirement) string {
	switch {
	case r.Coins > 0:
		if r.QuestFree != "" {
			return fmt.Sprintf("%d coins (or complete quest %s)", r.Coins, r.QuestFree)
		}
		return fmt.Sprintf("%d coins", r.Coins)
	case r.Item != "":
		return r.Item
	case r.SkillName != "":
		return fmt.Sprintf("%s %d", r.SkillName, r.SkillLevel)
	case r.QuestDone != "":
		return "quest " + r.QuestDone
	case r.Members:
		return "members"
	case r.None:
		return ""
	default:
		return ""
	}
}

// youHave returns the host's current standing on the gate's binding dimension:
// the coin count for a coin gate, the (best-matching) skill level for a skill
// gate, and 1/0 for item/quest/members gates. Lets the host compare what it has
// to what the gate needs without re-deriving it.
func youHave(r Requirement, cap Capability) int {
	switch {
	case r.Coins > 0:
		return cap.Coins
	case r.Item != "":
		if cap.hasItem(r.Item) {
			return 1
		}
		return 0
	case r.SkillName != "":
		return cap.skillLevel(r.SkillName)
	case r.QuestDone != "":
		if cap.questDone(r.QuestDone) {
			return 1
		}
		return 0
	case r.Members:
		if cap.Members {
			return 1
		}
		return 0
	default:
		return 0
	}
}
