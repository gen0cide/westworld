package assets

import (
	"fmt"
	"os"
)

// Archive is a decoded JAG / OpenRSC .orsc container. After Open, Body holds
// the (outer-decompressed) catalog+payload buffer; offset 0 is the u16 entry
// count exactly as Utility.getDataFileOffset expects.
type Archive struct {
	Body    []byte
	entries []jagEntry
}

type jagEntry struct {
	nameHash int32
	uncomp   int // uncompressed length
	comp     int // compressed (stored) length
	offset   int // byte offset into Body where this entry's payload begins
}

func readU24(b []byte, o int) int {
	return int(b[o]&0xff)<<16 | int(b[o+1]&0xff)<<8 | int(b[o+2]&0xff)
}

func readU16(b []byte, o int) int {
	return int(b[o]&0xff)<<8 | int(b[o+1]&0xff)
}

// nameHash replicates Utility.getDataFileOffset's hash: h = h*61 + upper(c) - 32
// over uppercased characters, with Java int32 wraparound.
func nameHash(s string) int32 {
	var h int32
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			c -= 32 // toUpperCase for ASCII
		}
		h = h*61 + c - 32
	}
	return h
}

// OpenArchive reads a JAG/.orsc file from disk and parses its catalog.
//
// Outer container (GameApplet.readDataFile): 6-byte header
//   [u24 expectedTotal][u24 total]; if total != expectedTotal the body is a
// single bzip2 stream of expectedTotal bytes; else stored.
//
// Inner JAG (Utility.unpackData): the body starts with [u16 entryCount] then
//   entryCount * { u32 nameHash, u24 uncompLen, u24 compLen }; payload is
// concatenated after the directory at offset 2 + count*10. Per entry, if
// uncomp != comp the entry is individually bzip2-compressed.
func OpenArchive(path string) (*Archive, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < 6 {
		return nil, fmt.Errorf("archive %s too small", path)
	}
	expectedTotal := readU24(raw, 0)
	total := readU24(raw, 3)
	body := raw[6:]
	if total != expectedTotal {
		body, err = Decompress(body[:total], expectedTotal)
		if err != nil {
			return nil, fmt.Errorf("outer decompress %s: %w", path, err)
		}
	}

	a := &Archive{Body: body}
	count := readU16(body, 0)
	off := 2 + count*10
	for i := 0; i < count; i++ {
		base := i*10 + 2
		nh := int32(uint32(body[base])<<24 | uint32(body[base+1])<<16 |
			uint32(body[base+2])<<8 | uint32(body[base+3]))
		uncomp := readU24(body, base+4)
		comp := readU24(body, base+7)
		a.entries = append(a.entries, jagEntry{
			nameHash: nh, uncomp: uncomp, comp: comp, offset: off,
		})
		off += comp
	}
	return a, nil
}

// EntryCount is the number of catalog entries.
func (a *Archive) EntryCount() int { return len(a.entries) }

// Get returns the (decompressed) bytes for the named entry, or nil if absent.
func (a *Archive) Get(name string) ([]byte, error) {
	h := nameHash(name)
	for _, e := range a.entries {
		if e.nameHash == h {
			return a.read(e)
		}
	}
	return nil, fmt.Errorf("entry %q not found", name)
}

// Has reports whether the named entry exists.
func (a *Archive) Has(name string) bool {
	h := nameHash(name)
	for _, e := range a.entries {
		if e.nameHash == h {
			return true
		}
	}
	return false
}

func (a *Archive) read(e jagEntry) ([]byte, error) {
	if e.offset+e.comp > len(a.Body) {
		return nil, fmt.Errorf("entry payload out of range: off=%d comp=%d len=%d", e.offset, e.comp, len(a.Body))
	}
	payload := a.Body[e.offset : e.offset+e.comp]
	if e.uncomp == e.comp {
		out := make([]byte, e.uncomp)
		copy(out, payload)
		return out, nil
	}
	return Decompress(payload, e.uncomp)
}
