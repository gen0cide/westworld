// Command varrock-tiles inspects the OpenRSC landscape around a center point and
// reports, per tile, whether it is WALKABLE (worldmap collision) and whether it
// carries the town PATH/road ground overlay (id 1 — the grey "cement" texture).
//
// It exists to answer one operational question: when we teleport drones into
// Varrock, are the target tiles actually on the paved paths (walkable, not inside
// an object, not grass)? It prints an ASCII map and, with -paths, the list of
// validated path tiles inside the bbox that are reachable from the center — the
// safe dispersed teleport set.
//
//	varrock-tiles                       # map around Varrock center (122,509)
//	varrock-tiles -probe 122,509        # report one tile's overlay + walkable
//	varrock-tiles -paths -n 200         # emit up to 200 dispersed path tiles
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/worldmap"
)

const pathOverlay = 1 // RSC ground overlay id for road/path ("cement") — TileDef.xml

func main() {
	root := flag.String("facts", "/Users/flint/Code/openrsc", "OpenRSC source root")
	cx := flag.Int("cx", 122, "center X (Varrock town point)")
	cy := flag.Int("cy", 509, "center Y")
	minX := flag.Int("minx", 95, "bbox min X (inside-gates region)")
	maxX := flag.Int("maxx", 152, "bbox max X")
	minY := flag.Int("miny", 482, "bbox min Y (north)")
	maxY := flag.Int("maxy", 535, "bbox max Y (south)")
	probe := flag.String("probe", "", "report a single x,y tile and exit")
	emitPaths := flag.Bool("paths", false, "emit the validated dispersed path-tile set")
	n := flag.Int("n", 200, "max path tiles to emit (spread evenly)")
	out := flag.String("o", "", "write emitted path tiles to this file (x y per line)")
	flag.Parse()

	f, err := facts.Load(facts.DefaultSources(*root))
	must("load facts", err)
	lpath := filepath.Join(*root, "server", "conf", "server", "data", "Authentic_Landscape.orsc")
	ls, err := pathfind.OpenLandscape(lpath)
	must("open landscape", err)
	defer ls.Close()
	oracle, err := worldmap.Precompute(f, ls)
	must("precompute oracle", err)

	overlayAt := func(x, y int) byte {
		key := pathfind.SectorForWorld(x, y, 0)
		sec, err := ls.Sector(key)
		if err != nil || sec == nil {
			return 255
		}
		lx, ly := pathfind.TileLocalInSector(x, y)
		return sec.At(lx, ly).GroundOverlay
	}
	walkable := func(x, y int) bool { return oracle.CompAt(x, y) >= 0 }
	isPath := func(x, y int) bool { return walkable(x, y) && overlayAt(x, y) == pathOverlay }

	if *probe != "" {
		var x, y int
		if _, err := fmt.Sscanf(*probe, "%d,%d", &x, &y); err != nil {
			fmt.Fprintln(os.Stderr, "bad -probe (want x,y):", err)
			os.Exit(2)
		}
		fmt.Printf("tile (%d,%d): overlay=%d walkable=%v comp=%d path=%v\n",
			x, y, overlayAt(x, y), walkable(x, y), oracle.CompAt(x, y), isPath(x, y))
		return
	}

	// Center sanity: the teleport target everyone currently lands on.
	fmt.Printf("CENTER (%d,%d): overlay=%d walkable=%v path=%v  (overlay 1 == cement path)\n\n",
		*cx, *cy, overlayAt(*cx, *cy), walkable(*cx, *cy), isPath(*cx, *cy))

	centerComp := oracle.CompAt(*cx, *cy)

	// ASCII map: north (minY) at top. @ center, P path, , walkable-nonpath, # blocked.
	fmt.Printf("map x[%d..%d] y[%d..%d]  (@=center P=path ,=walkable #=blocked)\n", *minX, *maxX, *minY, *maxY)
	fmt.Print("      ")
	for x := *minX; x <= *maxX; x++ {
		fmt.Print(x % 10)
	}
	fmt.Println()
	var pathTiles [][2]int
	for y := *minY; y <= *maxY; y++ {
		fmt.Printf("%5d ", y)
		for x := *minX; x <= *maxX; x++ {
			switch {
			case x == *cx && y == *cy:
				fmt.Print("@")
			case isPath(x, y) && oracle.CompAt(x, y) == centerComp:
				fmt.Print("P")
				pathTiles = append(pathTiles, [2]int{x, y})
			case isPath(x, y):
				fmt.Print("p") // path tile but NOT in center's component (boxed off)
			case walkable(x, y):
				fmt.Print(",")
			default:
				fmt.Print("#")
			}
		}
		fmt.Println()
	}
	fmt.Printf("\nvalidated path tiles (overlay 1 + walkable + reachable from center): %d\n", len(pathTiles))

	if !*emitPaths {
		return
	}
	picked := spread(pathTiles, *n)
	fmt.Printf("emitting %d spread path tiles:\n", len(picked))
	var lines string
	for _, t := range picked {
		lines += fmt.Sprintf("%d %d\n", t[0], t[1])
	}
	if *out != "" {
		must("write tiles", os.WriteFile(*out, []byte(lines), 0o644))
		fmt.Printf("wrote %s\n", *out)
	} else {
		fmt.Print(lines)
	}
}

// spread picks up to n tiles spaced as evenly as possible across the set, so the
// drones disperse rather than cluster. Sorted by a space-filling-ish key then
// strided.
func spread(tiles [][2]int, n int) [][2]int {
	if n <= 0 || len(tiles) <= n {
		return tiles
	}
	sort.Slice(tiles, func(i, j int) bool {
		if tiles[i][0] != tiles[j][0] {
			return tiles[i][0] < tiles[j][0]
		}
		return tiles[i][1] < tiles[j][1]
	})
	out := make([][2]int, 0, n)
	step := float64(len(tiles)) / float64(n)
	for i := 0; i < n; i++ {
		out = append(out, tiles[int(float64(i)*step)])
	}
	return out
}

func must(what string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "varrock-tiles: %s: %v\n", what, err)
		os.Exit(1)
	}
}
