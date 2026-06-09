package runtime

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/facts"
)

// approachRect is the server-accepted stand area for interacting with an
// object — the EXACT rect Mob.atObject checks (GameObject.getObjectBoundary).
// strict means the player must stand INSIDE the rect (boundaries and type-2/3
// usable-direction scenery have NO adjacency slack: an out-of-rect interact is
// silently queued forever, never executed and never answered).
type approachRect struct {
	MinX, MinY, MaxX, MaxY int
	Strict                 bool
}

func (r approachRect) contains(x, y int) bool {
	return x >= r.MinX && x <= r.MaxX && y >= r.MinY && y <= r.MaxY
}

// boundaryApproachRect mirrors GameObject.getObjectBoundary's boundary branch
// (GameObject.java:85-110): dir 0 → the two tiles either side of the north
// edge; dir 1 → either side of the west edge; dir 2/3 (diagonal) → the 3x3;
// anything else → the boundary's own tile. Always strict (Mob.java:209-211).
func boundaryApproachRect(x, y, dir int) approachRect {
	switch dir {
	case 0:
		return approachRect{x, y - 1, x, y, true}
	case 1:
		return approachRect{x - 1, y, x, y, true}
	case 2, 3:
		return approachRect{x - 1, y - 1, x + 1, y + 1, true}
	default:
		return approachRect{x, y, x, y, true}
	}
}

// sceneryApproachRect mirrors GameObject.getObjectBoundary's scenery branch
// (GameObject.java:54-84): the dir-swapped footprint, expanded for type-2/3
// usable-direction objects exactly as the server computes it (dir 0 widens
// west+east, dir 2 extends south, dir 4 extends east, dir 6 extends
// north+south — the maxX/maxY are recomputed AFTER the expansion, which is
// the server's own quirk and therefore ours). Type 2/3 are strict (must
// stand inside); other types allow border adjacency (canReach).
func sceneryApproachRect(x, y, dir int, def *facts.SceneryDef) approachRect {
	w, h := def.Width, def.Height
	if dir != 0 && dir != 4 {
		w, h = h, w
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	r := approachRect{x, y, x + w - 1, y + h - 1, false}
	if def.Type == 2 || def.Type == 3 {
		r.Strict = true
		switch dir {
		case 0:
			r.MinX--
			r.MaxX = x + w
		case 2:
			r.MaxY = y + h
		case 4:
			r.MaxX = x + w
		case 6:
			r.MinY--
			r.MaxY = y + h
		}
	}
	return r
}

// approachAndAct walks the host INTO an interact rect (or adjacent, for
// non-strict rects) and then fires the action. If the host already stands in
// the accepted area the action fires immediately with NO walk — the exact
// case the generic border-walk got wrong (it would drag the host to some
// other "adjacent" tile, sometimes across the map). Rect tiles are tried
// nearest-first; an unreachable rect fails loudly rather than queuing a
// latent action the server will never run.
func (h *Host) approachAndAct(ctx context.Context, r approachRect, sendAction func(context.Context) error) error {
	pos := h.world.Self.Position()
	if r.contains(pos.X, pos.Y) {
		return sendAction(ctx)
	}
	if !r.Strict {
		// Border adjacency allowed (plain solid scenery): the generic
		// border walk is correct here.
		return h.walkAndAct(ctx, r.MinX, r.MinY, true, walkToPoint, sendAction)
	}
	// Strict: must stand INSIDE the rect. Try its tiles nearest-first.
	type cand struct{ x, y, d int }
	var cands []cand
	for ty := r.MinY; ty <= r.MaxY; ty++ {
		for tx := r.MinX; tx <= r.MaxX; tx++ {
			cands = append(cands, cand{tx, ty, chebyshev(pos.X, pos.Y, tx, ty)})
		}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].d < cands[j].d })
	for _, c := range cands {
		corners, err := h.pathToTile(c.x, c.y, false)
		if err != nil {
			continue
		}
		if len(corners) > 0 {
			if werr := action.WalkToEntityPath(ctx, h.conn, corners); werr != nil {
				return fmt.Errorf("approach: send walk: %w", werr)
			}
			h.awaitArrival(ctx, c.x, c.y)
		}
		return sendAction(ctx)
	}
	return fmt.Errorf("approach: no reachable stand tile in interact rect (%d,%d)-(%d,%d)", r.MinX, r.MinY, r.MaxX, r.MaxY)
}

// InteractWithBoundary fires the default click on a boundary (door,
// window, fence edge) — e.g., to open a closed door.
//
// OpenRSC's GameObjectWallAction handler only sets a one-shot WalkToAction
// gated per tick on Mob.atObject — the player must stand strictly INSIDE
// the boundary's accepted rect or the click silently never executes. We
// therefore walk into the server-exact rect first.
func (h *Host) InteractWithBoundary(ctx context.Context, x, y, direction int) error {
	h.log.Info("InteractWithBoundary",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
	)
	return h.approachAndAct(ctx, boundaryApproachRect(x, y, direction), func(ctx context.Context) error {
		return action.InteractWithBoundary(ctx, h.conn, x, y, direction)
	})
}

// UseItemOnBoundary uses an inventory slot on a boundary edge (key-on-door,
// knife-on-vine). Same server-exact approach rect as InteractWithBoundary.
func (h *Host) UseItemOnBoundary(ctx context.Context, x, y, direction, slot int) error {
	h.log.Info("UseItemOnBoundary",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
		"slot", slot,
	)
	return h.approachAndAct(ctx, boundaryApproachRect(x, y, direction), func(ctx context.Context) error {
		return action.UseItemOnBoundary(ctx, h.conn, x, y, direction, slot)
	})
}

// UseItemOnItem combines two inventory items (e.g. needle on
// cloth, chisel on gem). No pathfinding — both items are in the
// host's inventory, fire the packet immediately.
func (h *Host) UseItemOnItem(ctx context.Context, slot1, slot2 int) error {
	h.log.Info("UseItemOnItem", "slot1", slot1, "slot2", slot2)
	return action.UseItemOnItem(ctx, h.conn, slot1, slot2)
}

// sceneryRectAt resolves the server-exact interact rect for the scenery at
// (x, y): static loc for the Direction, live def swap for the CURRENT type/
// footprint. ok=false when no def is known (fall back to the generic border
// walk rather than refusing to act).
func (h *Host) sceneryRectAt(x, y int) (approachRect, bool) {
	if h.facts == nil {
		return approachRect{}, false
	}
	dir := 0
	defID := -1
	for _, p := range h.facts.At(x, y) {
		if p.Kind == "scenery" {
			dir = p.Direction
			defID = p.DefID
			break
		}
	}
	if h.world != nil && h.world.Scenery != nil {
		if rec, ok := h.world.Scenery.At(x, y); ok {
			defID = rec.ID
		}
	}
	def := h.facts.SceneryDef(defID)
	if def == nil {
		return approachRect{}, false
	}
	return sceneryApproachRect(x, y, dir, def), true
}

// UseItemOnScenery uses an inventory slot on a world object (cook on fire,
// smelt on furnace). Walks into the def's server-exact interact rect first —
// type-2/3 usable-direction objects silently drop out-of-rect clicks.
func (h *Host) UseItemOnScenery(ctx context.Context, x, y, slot int) error {
	h.log.Info("UseItemOnScenery",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"slot", slot,
	)
	send := func(ctx context.Context) error {
		return action.UseItemOnScenery(ctx, h.conn, x, y, slot)
	}
	if r, ok := h.sceneryRectAt(x, y); ok {
		return h.approachAndAct(ctx, r, send)
	}
	return h.walkAndAct(ctx, x, y, true, walkToPoint, send)
}

// InteractAt fires the primary or secondary click on a scenery tile
// (opcode 136 / 79). option=1 = primary ("Chop", "Mine", "Open"),
// option=2 = secondary. Walks into the def's server-exact interact rect
// first (GameObject.getObjectBoundary geometry): the old generic border
// walk could approach a usable-direction object from a side the server's
// atObject rejects, queuing a click that never executes and never errors.
func (h *Host) InteractAt(ctx context.Context, x, y, option int) error {
	h.log.Info("InteractAt",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"option", option,
	)
	send := func(ctx context.Context) error {
		return action.ObjectCommand(ctx, h.conn, x, y, option)
	}
	if r, ok := h.sceneryRectAt(x, y); ok {
		return h.approachAndAct(ctx, r, send)
	}
	return h.walkAndAct(ctx, x, y, true, walkToPoint, send)
}

// UseItemOnGroundItem fires opcode 53. Walks adjacent to the
// ground tile, then sends the use packet. groundItemID identifies
// which item-type to use on (multiple items can pile on one tile).
func (h *Host) UseItemOnGroundItem(ctx context.Context, x, y, groundItemID, slot int) error {
	h.log.Info("UseItemOnGroundItem: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"ground_item", groundItemID,
		"slot", slot,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.UseItemOnGroundItem(ctx, h.conn, x, y, groundItemID, slot)
	})
}

// UseItemOnNpc walks adjacent to the NPC's current position then
// fires the use-on-npc packet. The NPC's local server-index
// identifies which one (resolved from the npcView passed by the
// routine — index is stable for the NPC's lifetime in our view).
func (h *Host) UseItemOnNpc(ctx context.Context, npcServerIndex, slot int) error {
	pos := h.npcPos(npcServerIndex)
	h.log.Info("UseItemOnNpc: pathfinding",
		"npc_index", npcServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"slot", slot,
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.UseItemOnNpc(ctx, h.conn, npcServerIndex, slot)
	})
}

// NpcCommand walks adjacent to the NPC then fires its primary action
// command (opcode NPC_COMMAND → server command1). For thievable NPCs
// command1 is "pickpocket"; for others it's whatever the NpcDef lists.
// Distinct from talk_to and attack. The skill action (e.g. pickpocket
// loot/xp) repeats per click — call in a loop for multiple attempts.
func (h *Host) NpcCommand(ctx context.Context, npcServerIndex int) error {
	pos := h.npcPos(npcServerIndex)
	h.log.Info("NpcCommand: pathfinding",
		"npc_index", npcServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.NpcCommand(ctx, h.conn, npcServerIndex)
	})
}

// UseItemOnPlayer is the same shape as UseItemOnNpc but targets
// another player (trade-init, gift). Walks adjacent to the
// player's current position then fires opcode 113.
func (h *Host) UseItemOnPlayer(ctx context.Context, playerServerIndex, slot int) error {
	pos := h.playerPos(playerServerIndex)
	h.log.Info("UseItemOnPlayer: pathfinding",
		"player_index", playerServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"slot", slot,
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.UseItemOnPlayer(ctx, h.conn, playerServerIndex, slot)
	})
}

// npcPos / playerPos look up the current position of a tracked
// entity by server-index. Used by Use*OnNpc / Use*OnPlayer to
// pathfind to the entity before firing the packet. Returns zero
// Coord if not found (caller's walkAndAct surfaces the failure
// via the pathfind error path).
type coordXY struct{ X, Y int }

func (h *Host) npcPos(index int) coordXY {
	for _, n := range h.world.Npcs.All() {
		if n.Index == index {
			return coordXY{X: n.X, Y: n.Y}
		}
	}
	return coordXY{}
}

func (h *Host) playerPos(index int) coordXY {
	if rec, ok := h.world.Players.Get(index); ok {
		return coordXY{X: rec.X, Y: rec.Y}
	}
	return coordXY{}
}

// TraversableKind classifies an in-path obstacle the host can act on.
type TraversableKind int

const (
	// TraversableWallDoor is a wall BOUNDARY door (DoorDef, Unknown==1,
	// DoorType==1): opened via the boundary channel (InteractWithBoundary,
	// server OpBound → DoorAction.doDoor).
	TraversableWallDoor TraversableKind = iota
	// TraversableSceneryGate is a SCENERY door/gate (GameObject whose name
	// contains gate/door with an open/close command — the server's own
	// dispatch rule): opened via the scenery channel (InteractAt, server
	// OpLoc → DoorAction.handleGates → replaceGameObject def swap).
	TraversableSceneryGate
	// TraversableLadder is climbable scenery (Climb-up/-down etc.) that
	// teleports the host across a plane band; classified here for the
	// cross-plane router, never auto-opened by the walk flow.
	TraversableLadder
)

// Traversable is one actionable obstacle adjacent to a stalled/blocked walk:
// what it is, where it is, the def the CLOSED state carries (for
// confirm-by-state), and how to open it.
type Traversable struct {
	Kind     TraversableKind
	X, Y     int
	Dir      int // boundary edge dir (wall doors) / loc direction (scenery)
	ClosedID int // the def id observed while closed — confirm = it CHANGES
	Name     string
	Open     func(ctx context.Context) error
}

// ladderCommands are the scenery command1 values the server's Ladders plugin
// dispatches on (Ladders.java:26-32).
var ladderCommands = map[string]bool{
	"climb-up": true, "climb-down": true, "climb": true,
	"go up": true, "go down": true, "pull": true,
}

// findTraversableNear classifies the actionable obstacles within Chebyshev 1
// of (x, y), nearest-the-step-direction first. (towardX, towardY) is where
// the walk was heading — the straddle guard: an obstacle ON the segment from
// (x,y) toward the goal outranks one behind the host, so a stalled walk opens
// the door in its way, not the nearest door anywhere (the old wrong-door
// bug). Both the LIVE stores and static facts are consulted; an obstacle the
// live state already shows open is skipped.
func (h *Host) findTraversableNear(x, y, towardX, towardY int) []Traversable {
	if h.facts == nil {
		return nil
	}
	sx, sy := sign(towardX-x), sign(towardY-y)
	var out []Traversable

	// Wall-boundary doors: every edge of the 3x3 around the host. Live
	// overrides win (an opened door's def swap means SKIP — already open).
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx, ty := x+dx, y+dy
			for _, p := range h.facts.At(tx, ty) {
				if p.Kind != "boundary" {
					continue
				}
				id := p.DefID
				if h.world != nil && h.world.Boundaries != nil {
					if liveID, ok := h.world.Boundaries.Get(tx, ty, p.Direction); ok {
						if liveID < 0 {
							continue // removed (defensive) — passable
						}
						id = liveID
					}
				}
				def := h.facts.BoundaryDef(id)
				if def == nil || !def.IsOpenable() || !def.BlocksWhenClosed() {
					continue // fixed wall, open doorframe, or unknown
				}
				bx, by, bdir := tx, ty, p.Direction
				out = append(out, Traversable{
					Kind: TraversableWallDoor, X: bx, Y: by, Dir: bdir,
					ClosedID: id, Name: def.Name,
					Open: func(ctx context.Context) error {
						return h.InteractWithBoundary(ctx, bx, by, bdir)
					},
				})
			}
		}
	}

	// Scenery gates/doors + ladders: FOOTPRINT-AWARE — the loc tile of a
	// 1x2 gate may be the NEIGHBOR of the tile we're stalled against, so
	// scan placements within radius 2 and match on footprint coverage.
	for _, p := range h.facts.Near(x, y, 2) {
		if p.Kind != "scenery" {
			continue
		}
		id := p.DefID
		if h.world != nil && h.world.Scenery != nil {
			if rec, ok := h.world.Scenery.At(p.X, p.Y); ok {
				id = rec.ID // live def swap wins (open gate = open def)
			} else if h.world.Scenery.IsRemoved(p.X, p.Y) {
				continue // object gone
			}
		}
		def := h.facts.SceneryDef(id)
		if def == nil {
			continue
		}
		switch {
		case def.IsOpenableBarrier() && (def.Type == 1 || def.Type == 2):
			// A CLOSED gate/door (blocking type + open command). Must
			// actually touch the host's 3x3 with its footprint.
			if !footprintTouches(p.X, p.Y, p.Direction, def, x, y, 1) {
				continue
			}
			opt := 1
			if !isOpenCmd(def.Command1) && isOpenCmd(def.Command2) {
				opt = 2
			}
			gx, gy := p.X, p.Y
			out = append(out, Traversable{
				Kind: TraversableSceneryGate, X: gx, Y: gy, Dir: p.Direction,
				ClosedID: id, Name: def.Name,
				Open: func(ctx context.Context) error {
					return h.InteractAt(ctx, gx, gy, opt)
				},
			})
		case ladderCommands[strings.ToLower(def.Command1)]:
			if !footprintTouches(p.X, p.Y, p.Direction, def, x, y, 1) {
				continue
			}
			lx, ly := p.X, p.Y
			out = append(out, Traversable{
				Kind: TraversableLadder, X: lx, Y: ly, Dir: p.Direction,
				ClosedID: id, Name: def.Name,
				Open: func(ctx context.Context) error {
					return h.InteractAt(ctx, lx, ly, 1)
				},
			})
		}
	}

	// Straddle guard ordering: obstacles in the step direction first, then
	// on the host's own tile, then the rest — so the door IN THE WAY is
	// tried before the door behind us.
	sort.SliceStable(out, func(i, j int) bool {
		return straddleRank(out[i], x, y, sx, sy) < straddleRank(out[j], x, y, sx, sy)
	})
	return out
}

// findTraversableToward classifies openable barriers ON THE ROUTE toward a
// goal the BFS cannot reach — the plan-time analogue of findTraversableNear's
// stall recovery. A no-path very often means a closed door/gate somewhere
// BETWEEN the host and the goal (possibly far from both), so this scans the
// local grid window for openable barriers, keeps the ones the host can
// actually walk to (BFS-reachable border), and ranks them by the classic
// detour metric dist(host→barrier)+dist(barrier→goal) so the barrier most
// "on the way" is tried first. Capped to the best few — opening is expensive.
func (h *Host) findTraversableToward(goalX, goalY int) []Traversable {
	if h.facts == nil {
		return nil
	}
	pos := h.world.Self.Position()
	const scanRadius = 40 // the local grid covers ~±48 around the host
	var cands []Traversable
	seen := map[[3]int]bool{}

	add := func(t Traversable) {
		key := [3]int{t.X, t.Y, t.Dir}
		if seen[key] {
			return
		}
		seen[key] = true
		cands = append(cands, t)
	}

	for _, p := range h.facts.Near(pos.X, pos.Y, scanRadius) {
		switch p.Kind {
		case "boundary":
			id := p.DefID
			if h.world != nil && h.world.Boundaries != nil {
				if liveID, ok := h.world.Boundaries.Get(p.X, p.Y, p.Direction); ok {
					if liveID < 0 {
						continue
					}
					id = liveID
				}
			}
			def := h.facts.BoundaryDef(id)
			if def == nil || !def.IsOpenable() || !def.BlocksWhenClosed() {
				continue
			}
			bx, by, bdir := p.X, p.Y, p.Direction
			add(Traversable{
				Kind: TraversableWallDoor, X: bx, Y: by, Dir: bdir,
				ClosedID: id, Name: def.Name,
				Open: func(ctx context.Context) error {
					return h.InteractWithBoundary(ctx, bx, by, bdir)
				},
			})
		case "scenery":
			id := p.DefID
			if h.world != nil && h.world.Scenery != nil {
				if rec, ok := h.world.Scenery.At(p.X, p.Y); ok {
					id = rec.ID
				} else if h.world.Scenery.IsRemoved(p.X, p.Y) {
					continue
				}
			}
			def := h.facts.SceneryDef(id)
			if def == nil || !def.IsOpenableBarrier() || (def.Type != 1 && def.Type != 2) {
				continue
			}
			opt := 1
			if !isOpenCmd(def.Command1) && isOpenCmd(def.Command2) {
				opt = 2
			}
			gx, gy := p.X, p.Y
			add(Traversable{
				Kind: TraversableSceneryGate, X: gx, Y: gy, Dir: p.Direction,
				ClosedID: id, Name: def.Name,
				Open: func(ctx context.Context) error {
					return h.InteractAt(ctx, gx, gy, opt)
				},
			})
		}
	}

	// Most-on-the-way first: minimize the detour through the barrier.
	sort.SliceStable(cands, func(i, j int) bool {
		di := chebyshev(pos.X, pos.Y, cands[i].X, cands[i].Y) + chebyshev(cands[i].X, cands[i].Y, goalX, goalY)
		dj := chebyshev(pos.X, pos.Y, cands[j].X, cands[j].Y) + chebyshev(cands[j].X, cands[j].Y, goalX, goalY)
		return di < dj
	})

	// Keep only barriers the host can actually WALK TO (BFS-reachable
	// border) — opening something across a river helps nobody. The check
	// is a full BFS per candidate, so cap how many we vet and return.
	const maxCands = 3
	var out []Traversable
	for _, t := range cands {
		if len(out) >= maxCands {
			break
		}
		if corners, err := h.pathToTile(t.X, t.Y, true); err == nil || len(corners) > 0 {
			out = append(out, t)
		}
	}
	return out
}

// straddleRank orders traversables by how squarely they sit on the host's
// intended step: 0 = on the adjacent tile toward the goal, 1 = on the host's
// tile, 2 = elsewhere in the 3x3.
func straddleRank(t Traversable, x, y, sx, sy int) int {
	switch {
	case t.X == x+sx && t.Y == y+sy:
		return 0
	case t.X == x && t.Y == y:
		return 1
	default:
		return 2
	}
}

// footprintTouches reports whether the dir-swapped w×h footprint of a scenery
// def at (lx, ly) comes within Chebyshev `reach` of (x, y).
func footprintTouches(lx, ly, dir int, def *facts.SceneryDef, x, y, reach int) bool {
	w, h := def.Width, def.Height
	if dir != 0 && dir != 4 {
		w, h = h, w
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if absVal(lx+dx-x) <= reach && absVal(ly+dy-y) <= reach {
				return true
			}
		}
	}
	return false
}

func isOpenCmd(cmd string) bool {
	c := strings.ToLower(cmd)
	return c == "open" || c == "close"
}
