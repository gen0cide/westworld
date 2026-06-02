package assets

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

// Sprite is one decoded image from an OpenRSC sprite archive
// (Authentic_Sprites.orsc). Pixels are 0xAARRGGBB, one per (x,y) row-major.
// OpenRSC textures/icons store alpha=0 and use a COLOUR KEY for transparency
// (green 0x00ff00 for 3D textures, black 0x000000 for 2D icons), so callers
// must key on RGB, not the alpha byte — see render's texture/sprite builders.
//
// Mirrors com/openrsc/client/model/Sprite (the fields RequiresShift/XShift/
// YShift/Something1/Something2 are decoded though most consumers ignore them;
// Something1 is the texture size-class hint: Something1/64-1 → 0=64px, 1=128px).
type Sprite struct {
	Pixels        []uint32
	Width, Height int
	RequiresShift bool
	XShift, YShift int
	Something1, Something2 int
}

// SpriteArchive is an OpenRSC sprite container (Authentic_Sprites.orsc): a plain
// ZIP whose entries are named by their decimal sprite id ("0", "1", … "3225").
// This replaces the classic .jag sprite sources (textures17/media58/entity24);
// the OpenRSC client itself loads every sprite the same way — by numeric id,
// ignoring any package name (GraphicsController.loadSprite, line 3049:
// spriteArchive.getEntry(String.valueOf(id))).
type SpriteArchive struct {
	rc    *zip.ReadCloser
	byID  map[int]*zip.File
}

// OpenSprites opens an OpenRSC sprite ZIP archive and indexes its entries by
// the integer id encoded in each entry name.
func OpenSprites(path string) (*SpriteArchive, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	a := &SpriteArchive{rc: rc, byID: make(map[int]*zip.File, len(rc.File))}
	for _, f := range rc.File {
		if id, err := strconv.Atoi(f.Name); err == nil {
			a.byID[id] = f
		}
	}
	return a, nil
}

// Close releases the underlying archive.
func (a *SpriteArchive) Close() error { return a.rc.Close() }

// Has reports whether a sprite with the given id exists.
func (a *SpriteArchive) Has(id int) bool { _, ok := a.byID[id]; return ok }

// Count is the number of decimal-id sprite entries.
func (a *SpriteArchive) Count() int { return len(a.byID) }

// Sprite reads and unpacks the sprite with the given id. Returns (nil, nil) when
// the id is absent — mirroring the client's getUnknownSprite fallback so callers
// degrade gracefully (a missing icon/texture skips rather than crashes).
func (a *SpriteArchive) Sprite(id int) (*Sprite, error) {
	f, ok := a.byID[id]
	if !ok {
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open sprite %d: %w", id, err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read sprite %d: %w", id, err)
	}
	return unpackSprite(b)
}

// unpackSprite decodes the OpenRSC Sprite.unpack format (Sprite.java:48-76): a
// 25-byte big-endian header [i32 width, i32 height, u8 requiresShift, i32 xShift,
// i32 yShift, i32 something1, i32 something2] followed by width*height big-endian
// ARGB int32 pixels.
func unpackSprite(b []byte) (*Sprite, error) {
	const hdr = 25
	if len(b) < hdr {
		return nil, fmt.Errorf("sprite header truncated: %d bytes", len(b))
	}
	be := binary.BigEndian
	s := &Sprite{
		Width:         int(int32(be.Uint32(b[0:]))),
		Height:        int(int32(be.Uint32(b[4:]))),
		RequiresShift: b[8] == 1,
		XShift:        int(int32(be.Uint32(b[9:]))),
		YShift:        int(int32(be.Uint32(b[13:]))),
		Something1:    int(int32(be.Uint32(b[17:]))),
		Something2:    int(int32(be.Uint32(b[21:]))),
	}
	if s.Width < 0 || s.Height < 0 || s.Width > 4096 || s.Height > 4096 {
		return nil, fmt.Errorf("sprite dims out of range: %dx%d", s.Width, s.Height)
	}
	n := s.Width * s.Height
	if hdr+n*4 > len(b) {
		return nil, fmt.Errorf("sprite pixel underrun: need %d have %d (%dx%d)", hdr+n*4, len(b), s.Width, s.Height)
	}
	s.Pixels = make([]uint32, n)
	for i := range n {
		s.Pixels[i] = be.Uint32(b[hdr+i*4:])
	}
	return s, nil
}
