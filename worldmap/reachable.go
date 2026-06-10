package worldmap

// Capability-parameterized reachability: "given a host at tile (x,y) with
// this Capability, where can it get?" This is THE layer where game-logic
// gates act as CONDITIONAL barriers, exactly as designed:
//
//   - The pure-collision flood (floodFill) already collapsed free walking
//     connectivity into components — within a component, movement is free.
//   - GATE barriers (the Al-Kharid toll gate) overlay a conditional CUT on
//     the collision walkability: the flood may cross the barrier line ONLY
//     if the host's Capability satisfies the gate's Requirement. The toll
//     gate is the load-bearing case — the collision map has NO wall there
//     (Lumbridge and the Al-Kharid mine are one component), so a coinless
//     host must be blocked by this overlay, and a 10-coin host let through.
//   - TELEPORT edges (ferries, ladders, spirit trees) JUMP the host from a
//     board tile/area to a target tile when Capability satisfies — this is
//     how a host reaches a collision-isolated island (e.g. Entrana).
//
// The query is a per-host tile flood (metered upstream as in-world "study"
// time), so the shared Oracle stays read-only. It reuses the SAME movement
// predicate as floodFill / pathfind.FindPath so reachability never drifts
// from the walking engine.

// Reachable returns the set of component ids the host can reach from tile
// (fromX,fromY) under cap. It floods the collision walkability from the
// start tile, refusing to cross any gate Barrier whose Requirement cap does
// not satisfy, and following any teleport edge whose Requirement cap
// satisfies (boarding from a resolved From tile, or — for From-less ferries
// — from the mainland component). The start component is always included if
// the start tile is standable.
//
// Returns the reachable component-id set. Tiles in a void/non-standable
// start return an empty set.
func (o *Oracle) Reachable(fromX, fromY int, cap Capability) map[int32]bool {
	reached := map[int32]bool{}
	seen := o.reachableTiles(fromX, fromY, cap)
	for idx, hit := range seen {
		if hit {
			if lbl := o.labels[idx]; lbl >= 0 {
				reached[lbl] = true
			}
		}
	}
	return reached
}

// CanReach reports whether the host at (fromX,fromY) with cap can reach the
// TARGET TILE (toX,toY), snapping the target to its nearest standable tile
// like CompNear. Returns false if either endpoint cannot be placed on a
// standable tile.
//
// This is deliberately TILE-granular, not component-granular: a conditional
// gate barrier (the toll gate) cuts WITHIN a single collision component —
// Lumbridge and the Al-Kharid mine are both component 0, separated only by
// the overlaid toll cut. A component-set membership test would wrongly report
// the mine reachable the moment the east half of component 0 is reached, so
// reachability of a specific destination must consult the per-tile flood.
func (o *Oracle) CanReach(fromX, fromY, toX, toY int, cap Capability) bool {
	_, tx, ty, ok := o.CompNear(toX, toY)
	if !ok {
		return false
	}
	seen := o.reachableTiles(fromX, fromY, cap)
	return seen[tx*o.dimY+ty]
}

// reachableTiles is the core per-query flood. It returns a per-tile bitset
// (indexed x*dim+y) of the tiles the host can stand on. Gate barriers are
// conditional cuts; teleport edges add jump frontiers. A flat []bool (not a
// map) is used because the flood routinely covers the ~900k-tile mainland.
//
// It is a thin wrapper over reachableTilesGated: a gate/teleport edge is
// passable iff cap.Satisfies(e.Req). reachableTilesGated takes the
// per-edge passability as an explicit predicate so the reach-EXPLANATION
// layer (explain.go) can re-run the SAME flood with synthetic gate states to
// discover WHICH gate is binding — without duplicating the walking predicate.
func (o *Oracle) reachableTiles(fromX, fromY int, cap Capability) []bool {
	return o.reachableTilesGated(fromX, fromY, func(e *TransportEdge) bool {
		return cap.Satisfies(e.Req)
	})
}

// reachableTilesGated is reachableTiles parameterized by an explicit per-edge
// passability predicate `passable`: a gate barrier is a cut unless
// passable(edge) is true, and a teleport edge fires only when passable(edge)
// is true. Callers normally pass `cap.Satisfies(e.Req)` (see reachableTiles);
// explain.go passes synthetic predicates (all-gates-shut, single-gate-open) to
// isolate the binding gate.
func (o *Oracle) reachableTilesGated(fromX, fromY int, passable func(e *TransportEdge) bool) []bool {
	dim := o.dim
	dimY := o.dimY
	seen := make([]bool, dim*dimY)

	// Seed: snap the start to a standable tile (hosts often query from a POI
	// marker on a building footprint).
	sc, sx, sy, ok := o.CompNear(fromX, fromY)
	_ = sc
	if !ok {
		return seen
	}

	// Precompute the active gate barriers (only those NOT passable are cuts;
	// a passable gate is transparent). Each entry is the barrier plus whether
	// this host may cross it.
	type activeGate struct {
		b        *Barrier
		passable bool
	}
	var gates []activeGate
	for i := range o.edges {
		e := &o.edges[i]
		if e.Kind == "gate" && e.Barrier != nil {
			gates = append(gates, activeGate{b: e.Barrier, passable: passable(e)})
		}
	}

	// crossesBlockedGate reports whether stepping from (ax,ay) to (bx,by)
	// crosses a gate barrier this host may NOT pass.
	crossesBlockedGate := func(ax, ay, bx, by int) bool {
		for gi := range gates {
			if gates[gi].passable {
				continue
			}
			b := gates[gi].b
			switch b.Axis {
			case "x":
				// crossing between x<Line and x>=Line, dest col in [Lo,Hi]
				lo, hi := minInt(ax, bx), maxInt(ax, bx)
				if lo < b.Line && hi >= b.Line {
					if by >= b.Lo && by <= b.Hi {
						return true
					}
				}
			case "y":
				lo, hi := minInt(ay, by), maxInt(ay, by)
				if lo < b.Line && hi >= b.Line {
					if bx >= b.Lo && bx <= b.Hi {
						return true
					}
				}
			}
		}
		return false
	}

	// BFS frontier of standable tile indices.
	queue := make([]int, 0, 4096)
	push := func(x, y int) {
		idx := x*dimY + y
		if !seen[idx] {
			seen[idx] = true
			queue = append(queue, idx)
		}
	}
	push(sx, sy)

	// mainlandReached tracks whether the flood has touched the mainland
	// (largest) component, which is the board area for From-less ferries.
	mainland := o.mainlandComp()

	// teleport edges fire when their board tile/area is reached; we re-scan
	// edges whenever the frontier grows into a new component, but to keep it
	// simple we evaluate teleports lazily: each time we pop a tile, if it is
	// a board tile (or we are on mainland and the edge is From-less), and cap
	// satisfies, we seed the target. To avoid O(tiles*edges) we only check
	// teleports at component boundaries: track which components we've already
	// expanded teleports for.
	teleportExpanded := map[int32]bool{}

	expandTeleports := func(comp int32) {
		if teleportExpanded[comp] {
			return
		}
		teleportExpanded[comp] = true
		for i := range o.edges {
			e := &o.edges[i]
			if e.Kind != "teleport" || len(e.To) != 2 {
				continue
			}
			if !passable(e) {
				continue
			}
			// Resolve the target's component once (also the snapped tile).
			tc, tx, ty, ok := o.CompNear(e.To[0], e.To[1])
			if !ok {
				continue
			}

			boardable := false
			if len(e.From) == 2 {
				// Anchored edge: boardable from the component of its From tile.
				if fc := o.CompAt(e.From[0], e.From[1]); fc == comp {
					boardable = true
				} else if fc < 0 {
					if c, _, _, ok := o.CompNear(e.From[0], e.From[1]); ok && c == comp {
						boardable = true
					}
				}
			} else if comp == mainland && tc != mainland {
				// From-less ferry boardable from the mainland — but ONLY when it
				// bridges to a genuinely-isolated (non-mainland) component, e.g.
				// the Monk of Entrana ferry. We deliberately do NOT honor an
				// intra-mainland From-less teleport (its board point is unknown
				// and could sit behind a capability barrier — e.g. the Shantay
				// stone gate or a desert ladder west of the toll), which would
				// otherwise let a coinless host hop AROUND the toll gate.
				boardable = true
			}
			if !boardable {
				continue
			}
			// Jump to the target's snapped standable tile.
			push(tx, ty)
		}
	}

	for qi := 0; qi < len(queue); qi++ {
		idx := queue[qi]
		cx := idx / dimY
		cy := idx % dimY

		// When we enter a tile, expand any teleport edges boardable from its
		// component (once per component).
		if lbl := o.labels[idx]; lbl >= 0 {
			expandTeleports(lbl)
		}

		// Cardinal + diagonal expansion, identical predicate to floodFill,
		// plus the conditional gate-barrier check on the crossed step.
		// North (cy-1): cross dest south edge.
		if cy > 0 && cy%planeHeight != 0 {
			ni := cx*dimY + (cy - 1)
			if !seen[ni] && o.mask[ni]&southBlocked == 0 && !crossesBlockedGate(cx, cy, cx, cy-1) {
				push(cx, cy-1)
			}
		}
		if cy < dimY-1 && (cy+1)%planeHeight != 0 {
			ni := cx*dimY + (cy + 1)
			if !seen[ni] && o.mask[ni]&northBlocked == 0 && !crossesBlockedGate(cx, cy, cx, cy+1) {
				push(cx, cy+1)
			}
		}
		if cx > 0 {
			ni := (cx-1)*dimY + cy
			if !seen[ni] && o.mask[ni]&westBlocked == 0 && !crossesBlockedGate(cx, cy, cx-1, cy) {
				push(cx-1, cy)
			}
		}
		if cx < dim-1 {
			ni := (cx+1)*dimY + cy
			if !seen[ni] && o.mask[ni]&eastBlocked == 0 && !crossesBlockedGate(cx, cy, cx+1, cy) {
				push(cx+1, cy)
			}
		}
		// Diagonals (mirror floodFill): require both adjacent cardinals open
		// AND the diagonal corner clear AND neither crossed step blocked by a
		// gate.
		if cx > 0 && cy > 0 && cy%planeHeight != 0 &&
			o.mask[cx*dimY+(cy-1)]&southBlocked == 0 &&
			o.mask[(cx-1)*dimY+cy]&westBlocked == 0 {
			ni := (cx-1)*dimY + (cy - 1)
			if !seen[ni] && o.mask[ni]&southWestBlocked == 0 &&
				!crossesBlockedGate(cx, cy, cx-1, cy-1) {
				push(cx-1, cy-1)
			}
		}
		if cx < dim-1 && cy > 0 && cy%planeHeight != 0 &&
			o.mask[cx*dimY+(cy-1)]&southBlocked == 0 &&
			o.mask[(cx+1)*dimY+cy]&eastBlocked == 0 {
			ni := (cx+1)*dimY + (cy - 1)
			if !seen[ni] && o.mask[ni]&southEastBlocked == 0 &&
				!crossesBlockedGate(cx, cy, cx+1, cy-1) {
				push(cx+1, cy-1)
			}
		}
		if cx > 0 && cy < dimY-1 && (cy+1)%planeHeight != 0 &&
			o.mask[cx*dimY+(cy+1)]&northBlocked == 0 &&
			o.mask[(cx-1)*dimY+cy]&westBlocked == 0 {
			ni := (cx-1)*dimY + (cy + 1)
			if !seen[ni] && o.mask[ni]&northWestBlocked == 0 &&
				!crossesBlockedGate(cx, cy, cx-1, cy+1) {
				push(cx-1, cy+1)
			}
		}
		if cx < dim-1 && cy < dimY-1 && (cy+1)%planeHeight != 0 &&
			o.mask[cx*dimY+(cy+1)]&northBlocked == 0 &&
			o.mask[(cx+1)*dimY+cy]&eastBlocked == 0 {
			ni := (cx+1)*dimY + (cy + 1)
			if !seen[ni] && o.mask[ni]&northEastBlocked == 0 &&
				!crossesBlockedGate(cx, cy, cx+1, cy+1) {
				push(cx+1, cy+1)
			}
		}
	}
	return seen
}

// Route plans a coarse same-plane corridor from (fromX, fromY) toward
// (toX, toY): a parent-tracked BFS over the band-encoded grid using the
// SAME cardinal edge predicate as the component flood — so openable
// doors/gates are crossable (the walk layer opens them en route) and the
// route can never disagree with the oracle's own reachability. Diagonals
// are skipped (a corridor is waypoints, not exact steps; the local BFS
// walks each leg optimally). avoid marks tiles to treat as solid — the
// runtime passes its learned-impassable ledger (locked doors, toll gates)
// so re-routes cost gated paths honestly. Returns waypoints sampled every
// ~16 tiles plus the goal tile, or nil when no route exists.
func (o *Oracle) Route(fromX, fromY, toX, toY int, avoid map[[2]int]bool) [][2]int {
	if fromY/planeHeight != toY/planeHeight {
		return nil // cross-plane is the ladder router's job
	}
	_, sx, sy, ok := o.CompNear(fromX, fromY)
	if !ok {
		return nil
	}
	_, tx, ty, ok := o.CompNear(toX, toY)
	if !ok || ty/planeHeight != sy/planeHeight {
		return nil
	}
	dim, dimY := o.dim, o.dimY
	start := sx*dimY + sy
	goal := tx*dimY + ty
	if start == goal {
		return [][2]int{{tx, ty}}
	}

	parent := make([]int32, dim*dimY)
	for i := range parent {
		parent[i] = -2 // unvisited
	}
	parent[start] = -1
	queue := make([]int, 0, 8192)
	queue = append(queue, start)

	tryStep := func(from, nx, ny, destBlockedMask int) bool {
		if nx < 0 || nx >= dim || ny < 0 || ny >= dimY {
			return false
		}
		ni := nx*dimY + ny
		if parent[ni] != -2 || o.mask[ni]&destBlockedMask != 0 {
			return false
		}
		if avoid != nil && avoid[[2]int{nx, ny}] {
			return false
		}
		parent[ni] = int32(from)
		queue = append(queue, ni)
		return ni == goal
	}

	found := false
	for qi := 0; qi < len(queue) && !found; qi++ {
		idx := queue[qi]
		cx, cy := idx/dimY, idx%dimY
		// Same cardinal predicate as floodFill (the *Blocked masks include
		// fullBlock), with the floor-band guards.
		if cy > 0 && cy%planeHeight != 0 {
			found = tryStep(idx, cx, cy-1, southBlocked)
		}
		if !found && cy < dimY-1 && (cy+1)%planeHeight != 0 {
			found = tryStep(idx, cx, cy+1, northBlocked)
		}
		if !found && cx > 0 {
			found = tryStep(idx, cx-1, cy, westBlocked)
		}
		if !found && cx < dim-1 {
			found = tryStep(idx, cx+1, cy, eastBlocked)
		}
	}
	if !found {
		return nil
	}

	// Walk parents back from the goal, then downsample to waypoints.
	var rev [][2]int
	for idx := int32(goal); idx >= 0; idx = parent[idx] {
		rev = append(rev, [2]int{int(idx) / dimY, int(idx) % dimY})
	}
	const stride = 16
	var out [][2]int
	for i := len(rev) - 1; i >= 0; i-- {
		stepsFromStart := len(rev) - 1 - i
		if stepsFromStart != 0 && (stepsFromStart%stride == 0 || i == 0) {
			out = append(out, rev[i])
		}
	}
	return out
}

// mainlandComp returns the id of the largest component (the mainland), or -1
// if there are no components. From-less ferries board from here.
func (o *Oracle) mainlandComp() int32 {
	best := int32(-1)
	bestTiles := -1
	for _, c := range o.comps {
		// The mainland is a GROUND-FLOOR notion: the upper/underground
		// bands contain huge enclosed walkable voids (plane-3 rock) that
		// out-tile the real mainland but are reachable from nowhere.
		if c.RepY/planeHeight != 0 {
			continue
		}
		if c.Tiles > bestTiles {
			bestTiles = c.Tiles
			best = c.ID
		}
	}
	return best
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
