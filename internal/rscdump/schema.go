// Package rscdump defines the canonical rscdump/1 render-diff state schema
// (RENDER_DIFF_DESIGN.md §4) as Go structs with JSON round-trip + file Load/Save.
//
// A dump is a serialized, deterministic snapshot of "the world state at a tick"
// (level L1) or "the built geometry at a tick" (level L2). The point of the
// schema is that all three render engines — the Go render/ lib, the
// deobfuscated Java client, and the vanilla rev-235 jar via rscplus — load the
// IDENTICAL blob and render it, so a pixel/structural diff localizes a
// faithfulness bug to a specific builder rather than to RNG/network drift. See
// RENDER_DIFF_DESIGN.md §0-§1 for the framing and §4 for this schema.
//
// Phase 0 (the current milestone) uses L1 only: render.RenderDump consumes an
// L1 dump through the existing RenderView path. The L2 model list is defined
// here so the schema is complete and stable, even though no engine emits/loads
// it yet.
package rscdump

import (
	"encoding/json"
	"fmt"
	"os"
)

// SchemaID is the schema tag every dump carries; consumers reject mismatches.
const SchemaID = "rscdump/1"

// Level is the dump granularity (RENDER_DIFF_DESIGN.md §1).
const (
	// LevelL1 is high-level decoded WORLD state, BEFORE mesh building. Each
	// engine builds its own scene from it, so an L1 diff exercises the
	// scene-builder (BuildTerrain/BuildBoundaries/... ↔ World.buildSection ↔ the
	// jar mesher). This is the primary target — the door/bridge/roof bugs all
	// live in World→GameModel building.
	LevelL1 = "L1"
	// LevelL2 is the built GameModel[] list + camera, AFTER mesh building, so an
	// L2 diff isolates the rasterizer/projector. Defined but unused in Phase 0.
	LevelL2 = "L2"
)

// Source identifies who produced a dump (provenance), so a diff can label which
// engine each PNG came from.
const (
	SourceJarReplay = "jar-replay"
	SourceDeob      = "deob"
	SourceGo        = "go"
	// SourceHandAuthored marks a fixture written by hand (the Phase-0 single-tile
	// dump), as opposed to one captured from a running engine.
	SourceHandAuthored = "hand-authored"
)

// Dump is the top-level rscdump/1 document (RENDER_DIFF_DESIGN.md §4 schema).
type Dump struct {
	Schema string `json:"schema"` // always SchemaID
	Level  string `json:"level"`  // LevelL1 | LevelL2
	Source string `json:"source"` // who produced it (provenance)
	Tick   int    `json:"tick"`   // replay frame / sim tick

	Camera Camera  `json:"camera"`
	Window Window  `json:"window"`
	Terrain *Terrain `json:"terrain,omitempty"` // L1 only

	Scenery           []Scenery     `json:"scenery,omitempty"`
	Entities          []Entity      `json:"entities,omitempty"`
	GroundItems       []GroundItem  `json:"groundItems,omitempty"`
	Self              *Self         `json:"self,omitempty"`
	RemovedBoundaries []EdgeRef     `json:"removedBoundaries,omitempty"`
	RemovedScenery    []TileRef     `json:"removedScenery,omitempty"`

	Models []Model `json:"models,omitempty"` // L2 only
}

// Camera is the eye/projection state (RENDER_DIFF_DESIGN.md §4 + the per-engine
// mapping table). Angles are 0..1023 (pre-inversion); the frustum constants are
// pinned so the diff fails loudly if an engine uses a different value (rule 4).
type Camera struct {
	X int32 `json:"x"` // eye position, world units (GO Camera.CameraX after SetCamera)
	Y int32 `json:"y"`
	Z int32 `json:"z"`

	Pitch int32 `json:"pitch"` // 0..1023 camera angle (pre-inversion)
	Yaw   int32 `json:"yaw"`
	Roll  int32 `json:"roll"`

	Distance int32 `json:"distance"` // zoom*2 (GO View.Zoom -> SetCamera distance)

	ViewDist int32 `json:"viewDist"` // GO scene.go viewDist (=9)
	ClipNear int32 `json:"clipNear"` // GO scene.go clipNear (=5)
	ClipFar  int32 `json:"clipFar"`  // GO scene.go clipFar

	ScreenW int `json:"screenW"` // output image width
	ScreenH int `json:"screenH"` // output image height
}

// Window is the terrain window placement: which plane, the SW corner, and the
// per-side tile size of the per-tile grids in Terrain (RENDER_DIFF_DESIGN.md §4
// "window"). RenderView centres its own window on the host tile, so for the GO
// engine BaseX/BaseY are informational; Size must match Terrain.Size.
type Window struct {
	BaseX int `json:"baseX"` // SW corner world-tile X of the terrain grid
	BaseY int `json:"baseY"` // SW corner world-tile Y (plane-local)
	Plane int `json:"plane"`
	Size  int `json:"size"` // grid side length (tiles)
}

// Terrain carries the per-tile decoded grids explicitly so every engine renders
// the same bytes rather than re-decoding map files (rule 1). Every grid is
// row-major Size*Size, indexed [x*Size + y] with x the column (world-X − BaseX)
// and y the row (world-Y − BaseY), matching pathfind.Sector layout. The fields
// mirror the schema-field → per-engine mapping table in RENDER_DIFF_DESIGN.md §4.
type Terrain struct {
	Size int `json:"size"`

	Elevation    []byte  `json:"elevation"`    // GroundElevation (×3 = world height)
	GroundColour []byte  `json:"groundColour"` // GroundTexture (palette index)
	Overlay      []byte  `json:"overlay"`      // GroundOverlay / decoration id (pre-remap)
	Roof         []byte  `json:"roof"`         // RoofTexture / wall-roof id
	WallH        []byte  `json:"wallH"`        // HorizontalWall (door-def index, 1-based)
	WallV        []byte  `json:"wallV"`        // VerticalWall (door-def index, 1-based)
	WallDiag     []int32 `json:"wallDiag"`     // DiagonalWalls (int; 1..23999 walls, 48000+ scenery)

	// TileDirection is the per-tile object/heading direction (obf World 'mb' grid,
	// World.getTileDirection). It orients a diagonally-placed scenery object
	// (incl. diagonal doors, the 48000+ WallDiag band): World.addModels rotates the
	// object model by dir*32 and swaps its footprint width/height for dir 0/4.
	// OPTIONAL + backward-compatible: an omitted grid defaults to all-zero (dir 0),
	// the prior behaviour. Same row-major Size*Size layout as the other grids.
	TileDirection []byte `json:"tileDirection,omitempty"`

	// TerrainSeed pins the per-vertex ambience speckle (RNG kill, rule 2). 0 =
	// flat zero speckle (the simplest fully-reproducible choice); any other value
	// folds deterministically into the per-tile hash. See render.terrainAmbience.
	TerrainSeed int32 `json:"terrainSeed"`
}

// idx returns the row-major grid index for window-local (x, y).
func (t *Terrain) idx(x, y int) int { return x*t.Size + y }

// Scenery is one placed scenery object (RENDER_DIFF_DESIGN.md §4 scenery[]). Id
// joins SceneryDef.ID; Dir is the object heading.
type Scenery struct {
	X   int `json:"x"`
	Y   int `json:"y"`
	ID  int `json:"id"`
	Dir int `json:"dir"`
}

// Entity is one actor (npc or player) the host perceives (§4 entities[]).
type Entity struct {
	X          int      `json:"x"`
	Y          int      `json:"y"`
	Kind       string   `json:"kind"` // "npc" | "player"
	ID         int      `json:"id"`   // npc sprite id (config85) for kind=="npc"
	Heading    int      `json:"heading"`
	Equip      [12]int  `json:"equip"`      // worn-equipment sprite array (players)
	HairCol    int      `json:"hairCol"`
	TopCol     int      `json:"topCol"`
	TrouserCol int      `json:"trouserCol"`
	SkinCol    int      `json:"skinCol"`
	HasEquip   bool     `json:"hasEquip"`
}

// GroundItem is one dropped item the host perceives (§4 groundItems[]).
type GroundItem struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	ItemID int `json:"itemId"`
}

// Self is the local player's pose + appearance (§4 self).
type Self struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Heading  int     `json:"heading"`
	HasEquip bool    `json:"hasEquip"`
	Equip    [12]int `json:"equip"`
	Hair     int     `json:"hair"`
	Top      int     `json:"top"`
	Trouser  int     `json:"trouser"`
	Skin     int     `json:"skin"`
	// NoSelf, when true, suppresses drawing the local player (the bare-world
	// fixtures set it so the scene is purely terrain + boundaries).
	NoSelf bool `json:"noSelf,omitempty"`
}

// EdgeRef names a wall/door edge for the removed-boundary (door-open) override
// (§4 removedBoundaries[]). Dir uses the createModel convention (render.View
// .BoundaryRemoved): 0=east-west, 1=north-south, 2='\', 3='/'.
type EdgeRef struct {
	X   int `json:"x"`
	Y   int `json:"y"`
	Dir int `json:"dir"`
}

// TileRef names a tile for the removed-scenery override (§4 removedScenery[]).
type TileRef struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// --- L2 (built geometry) — defined for schema completeness; unused in Phase 0 ---

// Model is one built GameModel (§4 models[]): base verts + faces + lighting,
// the rasterizer-isolation payload an L2 diff feeds identically to all engines.
type Model struct {
	ID          int     `json:"id"`
	Transparent bool    `json:"transparent"`
	Depth       int32   `json:"depth"`
	LightAmbience int32 `json:"lightAmbience"`
	LightDiffuse  int32 `json:"lightDiffuse"`
	LightDir    [3]int32 `json:"lightDir"`

	Verts          [][3]int32 `json:"verts"`          // base vertexX/Y/Z
	VertexAmbience []int32    `json:"vertexAmbience"` // per-vertex speckle (RNG-pinned)
	Faces          []Face     `json:"faces"`
}

// Face is one model face (§4 models[].faces[]).
type Face struct {
	V         []int `json:"v"`         // vertex indices
	FillFront int32 `json:"fillFront"` // <0 flat 5:5:5 colour, >=0 texture id
	FillBack  int32 `json:"fillBack"`
	Intensity int32 `json:"intensity"`
	Gouraud   bool  `json:"gouraud"`
	Tag       int32 `json:"tag"`
}

// Validate checks the structural invariants a renderer relies on: the schema
// tag, a known level, and (for L1) self-consistent grid sizes. It does not
// enforce per-engine constant agreement — that is the diff's job (rule 4).
func (d *Dump) Validate() error {
	if d.Schema != SchemaID {
		return fmt.Errorf("rscdump: schema %q != %q", d.Schema, SchemaID)
	}
	switch d.Level {
	case LevelL1:
		if d.Terrain == nil {
			return fmt.Errorf("rscdump: L1 dump has no terrain")
		}
		if err := d.Terrain.validate(); err != nil {
			return err
		}
		if d.Window.Size != 0 && d.Window.Size != d.Terrain.Size {
			return fmt.Errorf("rscdump: window.size %d != terrain.size %d", d.Window.Size, d.Terrain.Size)
		}
	case LevelL2:
		// L2 has no terrain invariants; models are validated when loaded.
	default:
		return fmt.Errorf("rscdump: unknown level %q", d.Level)
	}
	return nil
}

func (t *Terrain) validate() error {
	n := t.Size * t.Size
	if t.Size <= 0 {
		return fmt.Errorf("rscdump: terrain.size must be > 0, got %d", t.Size)
	}
	grids := []struct {
		name string
		got  int
	}{
		{"elevation", len(t.Elevation)},
		{"groundColour", len(t.GroundColour)},
		{"overlay", len(t.Overlay)},
		{"roof", len(t.Roof)},
		{"wallH", len(t.WallH)},
		{"wallV", len(t.WallV)},
		{"wallDiag", len(t.WallDiag)},
		{"tileDirection", len(t.TileDirection)},
	}
	for _, g := range grids {
		// A grid may be omitted entirely (treated as all-zero); if present it
		// MUST be exactly Size*Size so an engine never indexes out of bounds.
		if g.got != 0 && g.got != n {
			return fmt.Errorf("rscdump: terrain.%s len %d != size*size %d", g.name, g.got, n)
		}
	}
	return nil
}

// Load reads + parses + validates an rscdump JSON document from path.
func Load(path string) (*Dump, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("rscdump: load %q: %w", path, err)
	}
	var d Dump
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("rscdump: parse %q: %w", path, err)
	}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

// Save validates + writes the dump as indented JSON to path.
func (d *Dump) Save(path string) error {
	if d.Schema == "" {
		d.Schema = SchemaID
	}
	if err := d.Validate(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("rscdump: marshal: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("rscdump: save %q: %w", path, err)
	}
	return nil
}
