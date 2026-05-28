package pathfind

// Corner is one waypoint on a path in absolute world coordinates. A
// Path is a sequence of corners describing direction changes — each
// consecutive pair represents a straight or diagonal segment.
type Corner struct {
	X, Y int
}

// FindPath runs a breadth-first search from (startX, startY) toward
// the rectangular goal area [xLow..xHigh] × [yLow..yHigh] (all
// inclusive, in absolute world coords). Returns the corner waypoints
// in WALK ORDER (start-side first, goal last) or nil if no path
// exists within the local grid.
//
// reachBorder mirrors the client's "talk across a counter" mode: when
// true, the search succeeds as soon as it can reach a tile adjacent
// to the goal rectangle through an open wall edge — used for clicking
// an NPC that stands behind a shop counter or wall.
//
// Direct port of /Users/flint/Code/openrsc/Client_Base/src/orsc/
// graphics/three/World.java:339 findPath().
func (g *Grid) FindPath(startX, startY, xLowW, xHighW, yLowW, yHighW int, reachBorder bool) []Corner {
	// Translate world coords to local grid coords.
	sx, sy, ok := g.World2Local(startX, startY)
	if !ok {
		return nil
	}
	xLow, xHigh, yLow, yHigh := xLowW-g.BaseX, xHighW-g.BaseX, yLowW-g.BaseY, yHighW-g.BaseY

	// Source-direction matrix (records which neighbor each cell was
	// reached from, for reconstruction).
	var src [GridSize][GridSize]int
	src[sx][sy] = visitedStart

	// BFS open list, ring buffer.
	const cap = GridSize * GridSize
	var qx, qy [cap]int
	qx[0] = sx
	qy[0] = sy
	read := 0
	write := 1
	curX, curY := sx, sy
	found := false

	for read != write {
		curX = qx[read]
		curY = qy[read]
		read = (read + 1) % cap

		// Goal hit?
		if curX >= xLow && curX <= xHigh && curY >= yLow && curY <= yHigh {
			found = true
			break
		}

		// reachBorder: succeed when we're orthogonally adjacent to the
		// goal rect AND the wall between us and the goal isn't set.
		// Mirrors World.java:371-394.
		if reachBorder {
			if curX > 0 && curX-1 >= xLow && curX-1 <= xHigh && curY >= yLow && curY <= yHigh &&
				g.Mask[curX-1][curY]&WallWest == 0 {
				found = true
				break
			}
			if curX < GridSize-1 && curX+1 >= xLow && curX+1 <= xHigh && curY >= yLow && curY <= yHigh &&
				g.Mask[curX+1][curY]&WallEast == 0 {
				found = true
				break
			}
			if curY > 0 && curY-1 >= yLow && curY-1 <= yHigh && curX >= xLow && curX <= xHigh &&
				g.Mask[curX][curY-1]&WallSouth == 0 {
				found = true
				break
			}
			if curY < GridSize-1 && curY+1 >= yLow && curY+1 <= yHigh && curX >= xLow && curX <= xHigh &&
				g.Mask[curX][curY+1]&WallNorth == 0 {
				found = true
				break
			}
		}

		// Cardinal expansion. The "*Blocked" constants are named after
		// the WALL_* bit they include (matching Java's CollisionFlag
		// naming), NOT the movement direction — so northward moves
		// check SouthBlocked (wall on the destination's south edge,
		// which is what we're crossing) and vice versa.
		if curX > 0 && src[curX-1][curY] == 0 &&
			g.Mask[curX-1][curY]&WestBlocked == 0 {
			qx[write] = curX - 1
			qy[write] = curY
			src[curX-1][curY] = sourceWest
			write = (write + 1) % cap
		}
		if curX < GridSize-1 && src[curX+1][curY] == 0 &&
			g.Mask[curX+1][curY]&EastBlocked == 0 {
			qx[write] = curX + 1
			qy[write] = curY
			src[curX+1][curY] = sourceEast
			write = (write + 1) % cap
		}
		if curY > 0 && src[curX][curY-1] == 0 &&
			g.Mask[curX][curY-1]&SouthBlocked == 0 {
			qx[write] = curX
			qy[write] = curY - 1
			// We walked north to reach (curX, curY-1). To unwind, we
			// step south back to parent — sourceSouth is the bit that
			// drives y++ in the reconstruction loop.
			src[curX][curY-1] = sourceSouth
			write = (write + 1) % cap
		}
		if curY < GridSize-1 && src[curX][curY+1] == 0 &&
			g.Mask[curX][curY+1]&NorthBlocked == 0 {
			qx[write] = curX
			qy[write] = curY + 1
			// Walked south; unwind goes north → sourceNorth (y--).
			src[curX][curY+1] = sourceNorth
			write = (write + 1) % cap
		}

		// Diagonal expansion: require BOTH cardinal moves to be open
		// AND the diagonal tile itself not to be full-or-corner-
		// blocked. The "*Blocked" names match Java's CollisionFlag
		// constants and refer to the wall edge bit they include — so
		// going *north* uses SouthBlocked (the destination tile's
		// south-edge wall is the one being crossed), and going north-
		// west uses SouthWestBlocked on the NW tile.
		if curX > 0 && curY > 0 &&
			g.Mask[curX][curY-1]&SouthBlocked == 0 &&
			g.Mask[curX-1][curY]&WestBlocked == 0 &&
			g.Mask[curX-1][curY-1]&SouthWestBlocked == 0 &&
			src[curX-1][curY-1] == 0 {
			qx[write] = curX - 1
			qy[write] = curY - 1
			src[curX-1][curY-1] = sourceSouthWest
			write = (write + 1) % cap
		}
		if curX < GridSize-1 && curY > 0 &&
			g.Mask[curX][curY-1]&SouthBlocked == 0 &&
			g.Mask[curX+1][curY]&EastBlocked == 0 &&
			g.Mask[curX+1][curY-1]&SouthEastBlocked == 0 &&
			src[curX+1][curY-1] == 0 {
			qx[write] = curX + 1
			qy[write] = curY - 1
			src[curX+1][curY-1] = sourceSouthEast
			write = (write + 1) % cap
		}
		if curX > 0 && curY < GridSize-1 &&
			g.Mask[curX][curY+1]&NorthBlocked == 0 &&
			g.Mask[curX-1][curY]&WestBlocked == 0 &&
			g.Mask[curX-1][curY+1]&NorthWestBlocked == 0 &&
			src[curX-1][curY+1] == 0 {
			qx[write] = curX - 1
			qy[write] = curY + 1
			src[curX-1][curY+1] = sourceNorthWest
			write = (write + 1) % cap
		}
		if curX < GridSize-1 && curY < GridSize-1 &&
			g.Mask[curX][curY+1]&NorthBlocked == 0 &&
			g.Mask[curX+1][curY]&EastBlocked == 0 &&
			g.Mask[curX+1][curY+1]&NorthEastBlocked == 0 &&
			src[curX+1][curY+1] == 0 {
			qx[write] = curX + 1
			qy[write] = curY + 1
			src[curX+1][curY+1] = sourceNorthEast
			write = (write + 1) % cap
		}
	}

	if !found {
		return nil
	}

	// Reconstruct backwards from goal to start, recording only the
	// tiles where the source direction changes.
	x, y := curX, curY
	prevSrc := src[x][y]
	source := prevSrc
	// reverseCorners holds corners goal-first; we flip at the end so
	// the caller receives them start-first.
	reverseCorners := make([]Corner, 0, 16)
	reverseCorners = append(reverseCorners, Corner{X: g.BaseX + x, Y: g.BaseY + y})
	// Bound the unwind to grid-size steps so a bad source flag can't
	// run an infinite loop. A correct reconstruction needs at most
	// 2*GridSize steps.
	maxSteps := 2 * GridSize
	for (x != sx || y != sy) && maxSteps > 0 {
		if prevSrc != source {
			prevSrc = source
			reverseCorners = append(reverseCorners, Corner{X: g.BaseX + x, Y: g.BaseY + y})
		}
		if source&sourceSouth != 0 {
			y++
		} else if source&sourceNorth != 0 {
			y--
		}
		if source&sourceWest != 0 {
			x++
		} else if source&sourceEast != 0 {
			x--
		}
		if x < 0 || x >= GridSize || y < 0 || y >= GridSize {
			return nil // off the grid — broken reconstruction
		}
		source = src[x][y]
		maxSteps--
	}
	if x != sx || y != sy {
		// Didn't actually reach start — broken source-flag chain.
		return nil
	}

	// Flip into walk order (start-side first).
	out := make([]Corner, len(reverseCorners))
	for i, c := range reverseCorners {
		out[len(reverseCorners)-1-i] = c
	}
	return out
}

// FindPathToTile is a convenience wrapper for a single-tile goal.
func (g *Grid) FindPathToTile(startX, startY, goalX, goalY int, reachBorder bool) []Corner {
	return g.FindPath(startX, startY, goalX, goalX, goalY, goalY, reachBorder)
}
