package render

import "github.com/gen0cide/westworld/pathfind"

// terrainSize is the side length of the heightmap window (96x96 tiles, 2x2
// sectors), matching World.loadSection.
const terrainSize = 160

// waterOverlay is the GroundOverlay (getTileDecoration) value for water tiles
// in the OpenRSC landscape. Water is the dominant elev-0 overlay in the data
// (435/450 overlay-2 tiles sit at elevation 0). The reference client flattens
// water-category tiles (GameData.anIntArray98==4) to the water plane.
const waterOverlay = 2

// waterOverlay2 is the second water decoration id (11) the pathfind grid also
// treats as water (pathfind/grid.go). Both flatten + recolour to the water
// plane so neither forms a black gouraud cliff.
const waterOverlay2 = 11

// bridgeOverlay (250) is the on-disk GroundOverlay marker for BRIDGE tiles.
// Both authentic clients run a pre-pass that REMAPS 250 before building the
// terrain model (OpenRSC World.java:1632 setTileDecorationOnBridge, deob106
// p.java:805 so(), eggsampler World.java:433 setTiles): a chunk-seam tile whose
// +x/+y neighbour is neither bridge(250) nor water(2) becomes overlay 9 (a flat
// brown plank edge), every interior deck tile becomes overlay 2 — so the deck
// renders with the SAME water texture as the river it spans (the wooden bridge
// structure itself is separate scenery). The server (WorldLoader.java:380)
// remaps 250->2 for collision. We never remapped, so a 250 deck fell through to
// the grass underlay AND the shoreline-diagonal code bled water onto it — the
// "all messed up" bridge. The remap is plane-0 only (the client gates it on
// plane==0, World.java:1484).
const bridgeOverlay = 250

// bridgeOrWater reports whether an overlay id is a bridge marker (250) or water
// (2) — the client's chunk-seam test treats both as "still bridge/water" so the
// plank edge only forms where the deck abuts something else (dry land).
func bridgeOrWater(o byte) bool { return o == bridgeOverlay || o == waterOverlay }

// rawOverlayAt reads a tile's RAW (un-remapped) GroundOverlay at world coords,
// for the bridge chunk-seam neighbour test (the neighbour may be outside the
// render window). land.Tile is bounds-safe.
func rawOverlayAt(land *pathfind.Landscape, x, y, plane int) byte {
	return land.Tile(x, y, plane).GroundOverlay
}

// waterColour is the flat fill for water tiles (method305-encoded blue). The
// reference uses the config tile colour (anIntArray97) but absent the config
// table we paint a plausible RSC water blue so water reads as water, not void.
var waterColour = method305(40, 70, 140)

// waterTextureID is the authentic water texture (textures17.jag index 1). RSC
// textures the river surface with the animated water sprite; we render its
// first frame so the river reads as rippled water rather than a flat blue
// smear. A non-negative fill routes the face through the perspective
// textured-span path (method282); textureBuffer(1) is always present.
const waterTextureID int32 = 1

// waterShade is the FIXED flat shade index every water / shore quad is lit at.
// The shade ramp is built as ramp[255-j] = colour*(j*j)/65536, so a LOW index
// is BRIGHT and a high index is dark (index 0 = full colour, 255 = black). We
// pick a low index so water reads as bright water-blue. Because it is fixed
// (AddFixedFace), relight() never recomputes it from the steep shore normal, so
// the coastline can never gouraud-darken to the black "dark wedge".
const waterShade = 40

// BuildTerrain assembles the heightmap mesh GameModel for the 96x96 window
// whose SW corner is (baseX, baseY) in world-tile coords (use the same
// midRegion centring as pathfind.BuildGrid). Vertices are placed at
// (x*128, -height, y*128) with height = elevation*3 (World.getTerrainHeight),
// and quads are flat-coloured from the ground palette (no overlays/textures).
func BuildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) *GameModel {
	g, _ := buildTerrain(land, baseX, baseY, plane)
	return g
}

// TerrainHeights returns the flattened terrain-height grid for the window —
// the SAME heights BuildTerrain places its vertices at (water-flattened). The
// boundary builder anchors wall quads to these so walls sit flush on the
// ground (RSC method422 reads terrainHeightLocal, the post-flatten cache).
func TerrainHeights(land *pathfind.Landscape, baseX, baseY, plane int) [][]int32 {
	_, h := buildTerrain(land, baseX, baseY, plane)
	return h
}

func buildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) (*GameModel, [][]int32) {
	n := terrainSize
	// elevation + colour + water grids
	h := make([][]int32, n)
	col := make([][]int32, n)
	ovl := make([][]byte, n)
	water := make([][]bool, n)
	rawH := make([][]int32, n) // un-flattened elevation, for the raised bridge deck
	deck := make([][]bool, n)  // tileType-4 (water-class) tile = a raised plank bridge deck
	deckCount := 0             // # of deck tiles, for sizing the extra deck-quad verts/faces
	for i := 0; i < n; i++ {
		h[i] = make([]int32, n)
		col[i] = make([]int32, n)
		ovl[i] = make([]byte, n)
		water[i] = make([]bool, n)
		rawH[i] = make([]int32, n)
		deck[i] = make([]bool, n)
		for j := 0; j < n; j++ {
			t := land.Tile(baseX+i, baseY+j, plane)
			h[i][j] = int32(t.GroundElevation) * 3
			rawH[i][j] = h[i][j]
			col[i][j] = int32(t.GroundTexture)
			raw := t.GroundOverlay
			ovl[i][j] = raw
			// BRIDGE remap (plane 0 only): the on-disk overlay 250 is the bridge
			// marker; the client rewrites it BEFORE building terrain (see
			// bridgeOverlay). A chunk-seam tile (sector coord %48==47 whose +x then
			// +y neighbour is neither bridge nor water) becomes overlay 9 (a brown
			// plank edge); every interior deck tile becomes overlay 2 — so the deck
			// flattens + renders with the SAME water texture as the river it spans,
			// EXACTLY as the client does (overlay 2 -> water category). This is what
			// removes the grass + half-water mess on a 250 deck.
			if raw == bridgeOverlay && plane == 0 {
				wx, wy := baseX+i, baseY+j
				switch {
				case wx%48 == 47 && !bridgeOrWater(rawOverlayAt(land, wx+1, wy, plane)):
					ovl[i][j] = 9
				case wy%48 == 47 && !bridgeOrWater(rawOverlayAt(land, wx, wy+1, plane)):
					ovl[i][j] = 9
				default:
					ovl[i][j] = waterOverlay
				}
			}
			// FLATTEN + DECK are the SAME deob condition (World.java:891-895): a
			// terrain vertex is forced to y=0 IFF a touching tile's
			// getTileTypeOnPlane == 4. The genuine type-4 ("water-class") ids are
			// 4/12/20/21 (palette.go tileDefs). The earlier port keyed the flatten on
			// the overlay ID 2/11 instead — but overlay 2 (the river AND the
			// remapped-250 bridge interior) and overlay 11 (lava) are tileType 3, NOT
			// 4, so the deob does NOT flatten them; they stay at real elevation and
			// render as a textured type-3 overlay floor (water tex 1). Flattening
			// them by id sank the whole bridge deck into the river trench
			// (bridge_flatten_overlay2 / terrain-water-flatten-overlay-id). Keying on
			// tileType==4 keeps the deck flush with the banks.
			//
			// A type-4 tile is ALSO a RAISED BRIDGE DECK (the Lumbridge->Varrock road
			// bridge): the client renders it TWICE — the flattened water quad at river
			// level (this flatten + the water recolour below) AND a SECOND quad at the
			// tile's REAL elevation painted with the deck texture (deckPass below,
			// World.java:704-723). So the flatten flag and the deck flag coincide.
			if def, ok := overlayDef(ovl[i][j]); ok && def.tileType == 4 {
				water[i][j] = true // flatten the 4 shared vertices to y=0
				deck[i][j] = true
				deckCount++
			}
		}
	}

	// Flatten any vertex touched by a water tile to height 0 (the reference's
	// 4-corner check: a vertex is shared by up to 4 tiles; if ANY is water,
	// the vertex sinks to the water plane).
	flat := make([][]bool, n)
	for i := 0; i < n; i++ {
		flat[i] = make([]bool, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if !water[i][j] && !deck[i][j] {
				continue // deck (type-4) flattens too: its under-water seats at y=0
			}
			for di := 0; di <= 1; di++ {
				for dj := 0; dj <= 1; dj++ {
					vi, vj := i+di, j+dj
					if vi < n && vj < n {
						flat[vi][vj] = true
					}
				}
			}
		}
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if flat[i][j] {
				h[i][j] = 0
			}
		}
	}

	// is4 reports whether the tile at window-local (a,b) is a type-4 overlay tile
	// (the deob getTileDecoration/tileType==4 probe), false outside the window.
	is4 := func(a, b int) (int32, bool) {
		if a < 0 || b < 0 || a >= n || b >= n {
			return 0, false
		}
		if def, ok := overlayDef(ovl[a][b]); ok && def.tileType == 4 {
			return def.colour, true
		}
		return 0, false
	}
	// OVERLAY NEIGHBOUR-SPREAD count (World.java:1123-1135): buildOverlayTriangles
	// also paints a type-4 overlay quad onto each NON-type-3 tile that is adjacent
	// to a type-4 tile, so a water body has no grass-triangle gap at its border.
	// Count those extra quads up front so the model is sized to hold them (each is
	// 4 verts + 1 face), then emit them in the deck pass below
	// (terrain-overlay-no-neighbour-spread). The interior range is 1..n-2 to match
	// the deob's 1..94 bounds (it probes all four neighbours).
	spreadCount := 0
	for i := 1; i < n-1; i++ {
		for j := 1; j < n-1; j++ {
			if def, ok := overlayDef(ovl[i][j]); ok && (def.tileType == 3 || def.tileType == 4) {
				continue // type-3 emits nothing here; type-4 already drawn by the deck pass
			}
			for _, nb := range [4][2]int{{i, j + 1}, {i, j - 1}, {i + 1, j}, {i - 1, j}} {
				if _, ok := is4(nb[0], nb[1]); ok {
					spreadCount++
				}
			}
		}
	}

	// n*n terrain verts + 4 per raised bridge-deck quad + 4 per neighbour-spread
	// overlay quad; 2 faces/tile (split) + 1 deck face per deck tile + 1 spread
	// face. Pre-size so AddVertex/AddFace never index OOB and the vertexAmbience
	// slice covers every vertex.
	nV := n*n + 4*deckCount + 4*spreadCount
	nF := (n-1)*(n-1)*2 + deckCount + spreadCount
	g := NewGameModel(nV, nF)

	idx := func(i, j int) int { return i*n + j }
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			vi := g.AddVertex(int32(i)*128, -h[i][j], int32(j)*128)
			// Authentic per-vertex ambience speckle (World.java:696-697:
			// (int)(Math.random()*10)-5). Kept deterministic for render
			// caching by hashing the WORLD-tile coord instead of time-seeding.
			g.SetVertexAmbience(vi, terrainAmbience(baseX+i, baseY+j))
		}
	}

	// ovlClassAt is the tile's overlay id used as the surface "class" for the
	// diagonal colour split (the client's method420); 0 (grass underlay) outside
	// the window.
	ovlClassAt := func(a, b int) byte {
		if a < 0 || b < 0 || a >= n || b >= n {
			return 0
		}
		return ovl[a][b]
	}
	// tileTypeClass is the deob getTileType (World.java:288-292): -1 if the tile
	// has NO decoration, 1 if its overlay tileType == 2 (indoor floor), else 0.
	// The colour-split neighbour compare must use THIS class, not the raw overlay
	// id (World.java:933-942) — otherwise two DIFFERENT overlays of the SAME
	// tileType class (e.g. gungywater id 7 vs outdoor-special id 19, both type 3)
	// compare as different and the GO wrongly cuts a grass wedge at their concave
	// junction (terrain-split-id-vs-class). Folding through getTileType collapses
	// all type-0/1/3 decorated overlays into class 0, so a same-class junction is
	// "equal" and stays a solid quad, while a grass<->overlay or floor<->non-floor
	// boundary still splits.
	tileTypeClass := func(a, b int) int {
		ov := ovlClassAt(a, b)
		if ov == 0 {
			return -1
		}
		def, ok := overlayDef(ov)
		if !ok {
			return -1
		}
		if def.tileType == 2 {
			return 1
		}
		return 0
	}
	isWaterAt := func(a, b int) bool {
		return a >= 0 && a < n && b >= 0 && b < n && water[a][b]
	}
	// emitHalf draws one terrain triangle: a WATER half is back-face fixed-
	// textured (fillBack = water id, fixed bright shade so it never gouraud-
	// darkens); a land half is back-face gouraud at flat colour `fill`.
	emitHalf := func(tri []int, fill int32, isW bool) {
		if isW {
			g.AddFixedFace(tri, magic, waterTextureID, waterShade)
		} else {
			g.AddFace(tri, magic, fill, magic)
		}
	}

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			// quad corners (i,j)-(i+1,j)-(i+1,j+1)-(i,j+1)
			v0 := idx(i, j)
			v1 := idx(i+1, j)
			v2 := idx(i+1, j+1)
			v3 := idx(i, j+1)
			if water[i][j] {
				// Water tile: authentic water TEXTURE (id 1) via the textured-span
				// path, FIXED bright shade (the tile is flattened to the horizontal
				// y=0 plane so it never gouraud-darkens). Water is the client's
				// separate type-4 pass (World.java:820); the land tiles below handle
				// their own colour boundaries.
				g.AddFixedFace([]int{v0, v1, v2, v3}, waterTextureID, waterTextureID, waterShade)
				continue
			}
			// Resolve UNDERLAY (grass) colour + the tile's final colour (an overlay
			// floor — road/dirt/slab — OVERWRITES the underlay, World.java:714-726).
			underlay := groundColour[col[i][j]&0xff]
			c := underlay
			hasOverlay := false
			if def, ok := overlayDef(ovl[i][j]); ok {
				switch def.tileType {
				case 4:
					// Water-class tile (ids 4/12/20/21): the client FORCES these to
					// the water texture (deob World.java:719-726, OpenRSC
					// World.java:574-581) — id 12 to lava-ish tex 31, the rest to
					// water tex 1. Without this they'd fall through to def.colour
					// (e.g. id 4's colour 3 = planks) and wrongly render as a wood
					// floor in open water.
					c = waterTextureID
					if ovl[i][j] == 12 {
						c = 31
					}
					hasOverlay = true
				case 5:
					// Bridge sentinel (colour 12345678): out of scope — keep the
					// grass underlay so the sentinel never gets used as a texture id.
				default:
					// colour >= 0 is a TEXTURE id (chapel planks=3, marble=17->15,
					// pentagram=14->32, outdoor textured ids); < 0 is a method305 flat
					// colour. Either way scene.go routes the fill correctly: a >=0
					// fill textures the floor, a <0 fill paints flat.
					c = def.colour
					hasOverlay = true
				}
			}

			// COLOUR SPLIT (World.java:704-815): an overlay (path/road) tile reverts
			// ONE triangle half to the underlay (grass) on the side facing tiles of a
			// DIFFERENT surface, and picks which diagonal (l14) so the overlay edge
			// cuts DIAGONALLY across the tile instead of stair-stepping along the
			// grid — this is what makes paths read as smooth diagonals. k7 is the
			// first triangle's colour, i10 the second's; equal => no colour split.
			// k7/i10 = the two triangle colours; w0/w1 mark a half as WATER. l14
			// picks which diagonal carries the seam (World.java:766-803).
			k7, i10, l14 := c, c, 0
			w0, w1 := false, false
			// World.java:932 gate: a tileType-2 (indoor floor) tile is NOT
			// colour-split unless it ALSO carries an interior diagonal wall
			// (1..23999). Without this guard GO chamfers indoor-floor corners with
			// spurious grass wedges (terrain-type2-floor-split): a chapel/bank/house
			// floor reads as "rounded" instead of square. decoType-4 is handled by
			// the water shortcut above; decoType-5 (bridge sentinel) keeps grass.
			splitEligible := hasOverlay
			if def, ok := overlayDef(ovl[i][j]); ok && def.tileType == 2 {
				diagWall := int(land.Tile(baseX+i, baseY+j, plane).DiagonalWalls)
				splitEligible = diagWall > 0 && diagWall < 24000
			}
			if splitEligible {
				me := tileTypeClass(i, j)
				switch { // mirrors World.java:933-942 (getTileType neighbour compares)
				case tileTypeClass(i-1, j) != me && tileTypeClass(i, j-1) != me:
					k7, l14 = underlay, 0
				case tileTypeClass(i+1, j) != me && tileTypeClass(i, j+1) != me:
					i10, l14 = underlay, 0
				case tileTypeClass(i+1, j) != me && tileTypeClass(i, j-1) != me:
					i10, l14 = underlay, 1
				case tileTypeClass(i-1, j) != me && tileTypeClass(i, j+1) != me:
					k7, l14 = underlay, 1
				}
			} else if !hasOverlay {
				// SHORELINE diagonal (enhancement): where a grass tile meets water on
				// two ADJACENT edges (a corner of the water body), make the corner-
				// facing triangle WATER so the shore cuts diagonally instead of as a
				// blocky tile step — same diagonal-selection shape as the overlay split.
				switch {
				case isWaterAt(i-1, j) && isWaterAt(i, j-1):
					w0, l14 = true, 0
				case isWaterAt(i+1, j) && isWaterAt(i, j+1):
					w1, l14 = true, 0
				case isWaterAt(i+1, j) && isWaterAt(i, j-1):
					w1, l14 = true, 1
				case isWaterAt(i-1, j) && isWaterAt(i, j+1):
					w0, l14 = true, 1
				}
			}

			// Height twist: a non-planar tile must ALSO split (round-5 relief fix),
			// so the gouraud normal is exact per planar triangle (no faceting).
			twist := (h[i+1][j+1] - h[i+1][j]) + h[i][j+1] - h[i][j]

			if k7 != i10 || w0 != w1 || twist != 0 {
				if l14 == 0 {
					emitHalf([]int{v1, v0, v3}, k7, w0)
					emitHalf([]int{v3, v2, v1}, i10, w1)
				} else {
					emitHalf([]int{v3, v2, v0}, k7, w0)
					emitHalf([]int{v1, v0, v2}, i10, w1)
				}
			} else {
				g.AddFace([]int{v1, v0, v3, v2}, magic, c, magic)
			}
		}
	}

	// BRIDGE-DECK second pass (World.java:704-723 / eggsampler World.java:820-835):
	// a tileType-4 "water-class" tile is drawn a SECOND time as a quad at its REAL
	// (un-flattened) elevation, painted with the tile's deck texture (overlay-4
	// colour 3 = planks). The first pass above already drew the FLATTENED water
	// under it (the river); this raised quad is the brown plank deck the player
	// walks on. Same winding/back-slot fill as a textured terrain overlay, built
	// into the same model so it depth-sorts in FRONT of the lower water (different
	// elevation -> no z-fight). This is the Lumbridge->Varrock road bridge.
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			if !deck[i][j] {
				continue
			}
			def, ok := overlayDef(ovl[i][j])
			if !ok {
				continue
			}
			v0 := g.AddVertex(int32(i)*128, -rawH[i][j], int32(j)*128)
			v1 := g.AddVertex(int32(i+1)*128, -rawH[i+1][j], int32(j)*128)
			v2 := g.AddVertex(int32(i+1)*128, -rawH[i+1][j+1], int32(j+1)*128)
			v3 := g.AddVertex(int32(i)*128, -rawH[i][j+1], int32(j+1)*128)
			for _, v := range []int{v0, v1, v2, v3} {
				g.SetVertexAmbience(v, terrainAmbience(baseX+i, baseY+j))
			}
			g.AddFace([]int{v1, v0, v3, v2}, magic, def.colour, magic)
		}
	}

	// OVERLAY NEIGHBOUR-SPREAD pass (World.java:1123-1135 buildOverlayTriangles):
	// for each NON-type-3, NON-type-4 tile adjacent to a type-4 tile, emit a
	// type-4 overlay quad AT THIS tile at its raw elevation, painted with the
	// neighbouring type-4 tile's colour. This fills the grass-triangle gap a water
	// body would otherwise leave at its border (terrain-overlay-no-neighbour-
	// spread) — the deob emits 5 overlay quads for an isolated type-4 tile (the
	// tile + its 4 neighbours), where GO previously emitted only 1.
	for i := 1; i < n-1; i++ {
		for j := 1; j < n-1; j++ {
			if def, ok := overlayDef(ovl[i][j]); ok && (def.tileType == 3 || def.tileType == 4) {
				continue
			}
			for _, nb := range [4][2]int{{i, j + 1}, {i, j - 1}, {i + 1, j}, {i - 1, j}} {
				col, ok := is4(nb[0], nb[1])
				if !ok {
					continue
				}
				v0 := g.AddVertex(int32(i)*128, -rawH[i][j], int32(j)*128)
				v1 := g.AddVertex(int32(i+1)*128, -rawH[i+1][j], int32(j)*128)
				v2 := g.AddVertex(int32(i+1)*128, -rawH[i+1][j+1], int32(j+1)*128)
				v3 := g.AddVertex(int32(i)*128, -rawH[i][j+1], int32(j+1)*128)
				for _, v := range []int{v0, v1, v2, v3} {
					g.SetVertexAmbience(v, 0) // overlay quads carry no random speckle (deob terrain.e)
				}
				g.AddFace([]int{v1, v0, v3, v2}, magic, col, magic)
			}
		}
	}

	// Match the real World terrain light: setLight(true, 40, 48, -50,-10,-50)
	// -> ambience 96, diffuse 384, gouraud, light dir (-50,-10,-50).
	g.SetLight(40, 48, -50, -10, -50)
	return g, h
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// dumpTerrainSeed, when set to a NON-NEGATIVE value, overrides the per-vertex
// terrain-ambience speckle for a render-diff dump render (RENDER_DIFF_DESIGN.md
// §4 determinism rule 2). The vanilla client injects (int)(Math.random()*10)-5
// per terrain vertex, which is non-reproducible; for a byte-identical dump
// render every engine must instead derive the speckle from a fixed schema field
// (terrainSeed). A seed of 0 produces a flat ZERO speckle (no perturbation — the
// simplest fully-reproducible choice); a positive seed mixes into the per-tile
// hash so a dump can still carry textured relief deterministically. A negative
// value (the default, -1) means "no dump override" — live renders keep the
// existing coord-hash speckle unchanged. render.RenderDump sets this around its
// render and restores it; renders are single-threaded per call.
var dumpTerrainSeed int32 = -1

// withDumpTerrainSeed runs fn with the dump ambience seed pinned to seed, then
// restores the previous value. seed >= 0 enables the deterministic dump speckle
// (0 = flat zero); the caller passes the schema's terrainSeed straight through.
func withDumpTerrainSeed(seed int32, fn func()) {
	prev := dumpTerrainSeed
	dumpTerrainSeed = seed
	defer func() { dumpTerrainSeed = prev }()
	fn()
}

// terrainAmbience returns a stable pseudo-random ambience offset in [-5,4] for
// a world-tile coord, mirroring World.java's (int)(Math.random()*10)-5 but
// derived from a cheap integer hash of (x,y) so renders stay deterministic
// (cacheable) instead of time-seeded.
//
// During a dump render (dumpTerrainSeed >= 0) the speckle is instead derived
// purely from the dump's terrainSeed so the same dump renders byte-identical
// pixels across runs and engines: seed 0 yields a flat 0 (no speckle), any
// other seed folds into the same coord hash. This is the RNG kill of §4 rule 2.
func terrainAmbience(x, y int) int32 {
	if s := dumpTerrainSeed; s >= 0 {
		if s == 0 {
			return 0 // flat: fully reproducible, no per-vertex perturbation
		}
		h := uint32(x)*0x9e3779b1 + uint32(y)*0x85ebca77 + uint32(s)*0x27d4eb2f
		h ^= h >> 15
		h *= 0x2c1b3c6d
		h ^= h >> 12
		return int32(h%10) - 5
	}
	h := uint32(x)*0x9e3779b1 + uint32(y)*0x85ebca77
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	return int32(h%10) - 5
}
