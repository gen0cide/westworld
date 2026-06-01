package pathfind

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// SectorSize is the side length of one landscape sector. The archive
// stores each sector as 48*48 = 2304 tiles.
const SectorSize = 48

// Tile is one tile of decoded landscape data. The fields match the
// 10-byte on-disk record produced by
// /Users/flint/Code/openrsc/Client_Base/src/com/openrsc/client/model/Tile.java.
//
// HorizontalWall / VerticalWall hold the door-def index for any wall
// on this tile's edge (0 = none); DiagonalWalls is the door-def index
// for any diagonal wall (0 = none, 1..11999 = SW-NE, 12001..23999 =
// NW-SE per the server's loadSection convention).
type Tile struct {
	GroundElevation byte
	GroundTexture   byte
	GroundOverlay   byte
	RoofTexture     byte
	HorizontalWall  byte
	VerticalWall    byte
	DiagonalWalls   int32

	// TileDirection is the per-tile object/heading direction (World 'mb' grid,
	// World.getTileDirection). It orients a diagonally-placed scenery object
	// (incl. diagonal doors, the 48000+ DiagonalWalls band) in World.addModels:
	// the object model is rotated by dir*32 and its footprint width/height swap
	// for dir 0/4. The on-disk 10-byte .orsc sector record carries no direction
	// byte, so decodeSector leaves this 0 (dir 0) — additive + backward-compatible;
	// the render-diff harness (internal/rscdump) populates it from the dump.
	TileDirection byte
}

// Sector is a 48x48 grid of tiles. Indexing is x*SectorSize+y, where
// (x, y) are local within-sector coords (0..47).
type Sector struct {
	Tiles [SectorSize * SectorSize]Tile
}

// At returns the tile at (sectorLocalX, sectorLocalY). Coords are
// 0..47 within the sector. Out-of-range coords panic.
func (s *Sector) At(x, y int) Tile {
	return s.Tiles[x*SectorSize+y]
}

// World coordinate → archive sector index. The original RSC client
// uses worldX/48 within a region, but the archive entries are offset
// by `wildX = 2304` and `wildY = 1776 - (plane*944)`. So plane-0
// world (0, 0) lives in archive sector h0x48y37, and Lumbridge at
// roughly (130, 645) lives in h0x50y50.
//
// Reverse-engineered from
// /Users/flint/Code/openrsc/server/src/com/openrsc/server/io/WorldLoader.java
// loadWorld() iteration over sectors.
const (
	archiveOffsetX = 48 // = 2304 / 48 = wildX / sectorSize
	archiveOffsetY = 37 // = 1776 / 48 = wildY (plane 0) / sectorSize
)

// SectorKey identifies a sector in the archive by plane and archive
// coordinates.
type SectorKey struct {
	Plane int
	SX    int
	SY    int
}

// SectorForWorld returns the archive sector key that contains
// (worldX, worldY) at the given plane. In the .orsc archive the upper
// floors are stored as separate entries keyed ONLY by the h{plane}
// prefix at the SAME (sx, sy) archive coords as plane 0 (verified
// against Authentic_Landscape.orsc: h0x50y50, h1x50y50, h2x50y50,
// h3x50y50 all decode the Lumbridge-castle column). The server-side
// wildY = 1776 - plane*944 shift is a world-Y offset, NOT an archive
// key offset, so a tile's archive coords are plane-independent here and
// we use the same formula for every plane. This lets the renderer read
// an upper story's RoofTexture/walls (multi-story building roofs) while
// pathfind stays on the ground floor.
func SectorForWorld(worldX, worldY, plane int) SectorKey {
	return SectorKey{
		Plane: plane,
		SX:    worldX/SectorSize + archiveOffsetX,
		SY:    worldY/SectorSize + archiveOffsetY,
	}
}

// TileLocalInSector returns the tile coordinates within a sector for
// world (worldX, worldY).
func TileLocalInSector(worldX, worldY int) (int, int) {
	lx := worldX % SectorSize
	if lx < 0 {
		lx += SectorSize
	}
	ly := worldY % SectorSize
	if ly < 0 {
		ly += SectorSize
	}
	return lx, ly
}

// Landscape is a lazy-loading reader for the .orsc archive. It caches
// decoded sectors so we don't pay decompression cost on every pathfind
// run.
type Landscape struct {
	zr      *zip.ReadCloser
	entries map[string]*zip.File

	mu    sync.RWMutex
	cache map[SectorKey]*Sector
}

// OpenLandscape opens an .orsc ZIP archive. Caller owns the returned
// Landscape and must Close it when done. Returns the empty-but-usable
// archive if the file is missing — pathfind will degrade to
// no-landscape-walls mode (scenery + boundary defs only) rather than
// crashing.
func OpenLandscape(path string) (*Landscape, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("pathfind: open landscape archive %q: %w", path, err)
	}
	l := &Landscape{
		zr:      zr,
		entries: make(map[string]*zip.File, len(zr.File)),
		cache:   make(map[SectorKey]*Sector),
	}
	for _, f := range zr.File {
		l.entries[f.Name] = f
	}
	return l, nil
}

// NewMemoryLandscape builds a Landscape backed entirely by an in-memory set of
// pre-decoded sectors, with NO archive file behind it. Tile() reads straight
// out of the supplied cache (a missing key returns the zero Tile, exactly like
// a void sector in the on-disk archive). This is the seam the render-diff
// harness (internal/rscdump) uses to feed a hand-authored or dumped terrain
// grid through the unchanged render.RenderView path: a dump carries the per-tile
// grids explicitly (so all three engines render the same bytes rather than
// re-decoding map files — RENDER_DIFF_DESIGN.md determinism rule 1), and this
// constructor wraps them in a *Landscape the renderer already knows how to
// consume. The map is taken by reference; the caller must not mutate it after.
func NewMemoryLandscape(sectors map[SectorKey]*Sector) *Landscape {
	if sectors == nil {
		sectors = make(map[SectorKey]*Sector)
	}
	return &Landscape{cache: sectors}
}

// Close releases the underlying archive file.
func (l *Landscape) Close() error {
	if l == nil || l.zr == nil {
		return nil
	}
	return l.zr.Close()
}

// Sector returns the sector at the given archive coordinates. Returns
// nil (with no error) if the sector doesn't exist — many sectors are
// pure-blank "void" areas not stored in the archive.
func (l *Landscape) Sector(key SectorKey) (*Sector, error) {
	l.mu.RLock()
	if s, ok := l.cache[key]; ok {
		l.mu.RUnlock()
		return s, nil
	}
	l.mu.RUnlock()

	name := fmt.Sprintf("h%dx%dy%d", key.Plane, key.SX, key.SY)
	f, ok := l.entries[name]
	if !ok {
		// Mark missing so we don't re-look up.
		l.mu.Lock()
		l.cache[key] = nil
		l.mu.Unlock()
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("pathfind: open sector %q: %w", name, err)
	}
	defer rc.Close()
	buf, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("pathfind: read sector %q: %w", name, err)
	}
	if len(buf) != SectorSize*SectorSize*10 {
		return nil, fmt.Errorf("pathfind: sector %q wrong size: got %d bytes, want %d", name, len(buf), SectorSize*SectorSize*10)
	}
	s := decodeSector(buf)
	l.mu.Lock()
	l.cache[key] = s
	l.mu.Unlock()
	return s, nil
}

// Tile returns the tile at (worldX, worldY, plane). Returns the zero
// tile if the sector isn't in the archive (void area).
func (l *Landscape) Tile(worldX, worldY, plane int) Tile {
	key := SectorForWorld(worldX, worldY, plane)
	s, err := l.Sector(key)
	if err != nil || s == nil {
		return Tile{}
	}
	lx, ly := TileLocalInSector(worldX, worldY)
	return s.At(lx, ly)
}

func decodeSector(buf []byte) *Sector {
	s := &Sector{}
	for i := 0; i < SectorSize*SectorSize; i++ {
		off := i * 10
		s.Tiles[i] = Tile{
			GroundElevation: buf[off+0],
			GroundTexture:   buf[off+1],
			GroundOverlay:   buf[off+2],
			RoofTexture:     buf[off+3],
			HorizontalWall:  buf[off+4],
			VerticalWall:    buf[off+5],
			DiagonalWalls:   int32(binary.BigEndian.Uint32(buf[off+6 : off+10])),
		}
	}
	return s
}
