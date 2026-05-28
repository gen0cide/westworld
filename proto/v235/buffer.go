package v235

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Buffer is a growable byte buffer with helpers for the RSC wire
// encoding: big-endian integers, null- or newline-terminated strings,
// and bit-level access for the bitpacked update packets.
//
// Read and Write positions are independent; this matches the OpenRSC
// PacketBuilder / Packet pair, where construction writes from index 0
// and parsing reads from index 0.
type Buffer struct {
	data    []byte
	rpos    int
	wpos    int
	bitMode bool
	bitPos  int // current bit position when bitMode is true
}

// NewBuffer returns an empty buffer with the given initial capacity.
func NewBuffer(initialCap int) *Buffer {
	return &Buffer{data: make([]byte, 0, initialCap)}
}

// WrapBuffer wraps an existing byte slice as a Buffer for reading. The
// slice is not copied; do not mutate it after wrapping.
func WrapBuffer(b []byte) *Buffer {
	return &Buffer{data: b, wpos: len(b)}
}

// Bytes returns the underlying byte slice from the read position to the
// write position. Useful for sending the buffer's contents over the wire.
func (b *Buffer) Bytes() []byte { return b.data[b.rpos:b.wpos] }

// Len returns the number of unread bytes.
func (b *Buffer) Len() int { return b.wpos - b.rpos }

// Cap returns the underlying capacity.
func (b *Buffer) Cap() int { return cap(b.data) }

// Reset clears the buffer for reuse without freeing memory.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
	b.rpos = 0
	b.wpos = 0
	b.bitMode = false
	b.bitPos = 0
}

// ----- Write helpers -----

func (b *Buffer) ensure(n int) {
	need := b.wpos + n
	if need <= cap(b.data) {
		b.data = b.data[:need]
		return
	}
	// Grow with doubling.
	newCap := cap(b.data) * 2
	if newCap < need {
		newCap = need
	}
	grown := make([]byte, need, newCap)
	copy(grown, b.data[:b.wpos])
	b.data = grown
}

// WriteByte writes a single byte.
func (b *Buffer) WriteByte(v byte) {
	b.ensure(1)
	b.data[b.wpos] = v
	b.wpos++
}

// WriteBytes writes a raw byte slice.
func (b *Buffer) WriteBytes(src []byte) {
	b.ensure(len(src))
	copy(b.data[b.wpos:], src)
	b.wpos += len(src)
}

// WriteUint16 writes a big-endian uint16.
func (b *Buffer) WriteUint16(v uint16) {
	b.ensure(2)
	binary.BigEndian.PutUint16(b.data[b.wpos:], v)
	b.wpos += 2
}

// WriteUint32 writes a big-endian uint32.
func (b *Buffer) WriteUint32(v uint32) {
	b.ensure(4)
	binary.BigEndian.PutUint32(b.data[b.wpos:], v)
	b.wpos += 4
}

// WriteUint64 writes a big-endian uint64.
func (b *Buffer) WriteUint64(v uint64) {
	b.ensure(8)
	binary.BigEndian.PutUint64(b.data[b.wpos:], v)
	b.wpos += 8
}

// WriteInt16 / WriteInt32: signed convenience wrappers.
func (b *Buffer) WriteInt16(v int16) { b.WriteUint16(uint16(v)) }
func (b *Buffer) WriteInt32(v int32) { b.WriteUint32(uint32(v)) }

// WriteStringTerm writes a string terminated by the given byte. RSC
// uses '\n' (0x0A) for some strings and '\0' for others; specify which
// per call site. The terminator byte is included in the output.
func (b *Buffer) WriteStringTerm(s string, term byte) {
	b.WriteBytes([]byte(s))
	b.WriteByte(term)
}

// ----- Read helpers -----

// ReadByte reads a single byte. Returns io.EOF if no bytes remain.
func (b *Buffer) ReadByte() (byte, error) {
	if b.rpos >= b.wpos {
		return 0, io.EOF
	}
	v := b.data[b.rpos]
	b.rpos++
	return v, nil
}

// ReadBytes reads n bytes into a fresh slice. Errors if fewer than n
// bytes remain.
func (b *Buffer) ReadBytes(n int) ([]byte, error) {
	if b.rpos+n > b.wpos {
		return nil, fmt.Errorf("buffer: short read (need %d, have %d)", n, b.Len())
	}
	out := make([]byte, n)
	copy(out, b.data[b.rpos:b.rpos+n])
	b.rpos += n
	return out, nil
}

// ReadUint16 reads a big-endian uint16.
func (b *Buffer) ReadUint16() (uint16, error) {
	if b.rpos+2 > b.wpos {
		return 0, fmt.Errorf("buffer: short read (need 2, have %d)", b.Len())
	}
	v := binary.BigEndian.Uint16(b.data[b.rpos:])
	b.rpos += 2
	return v, nil
}

// ReadUint32 reads a big-endian uint32.
func (b *Buffer) ReadUint32() (uint32, error) {
	if b.rpos+4 > b.wpos {
		return 0, fmt.Errorf("buffer: short read (need 4, have %d)", b.Len())
	}
	v := binary.BigEndian.Uint32(b.data[b.rpos:])
	b.rpos += 4
	return v, nil
}

// ReadUint64 reads a big-endian uint64.
func (b *Buffer) ReadUint64() (uint64, error) {
	if b.rpos+8 > b.wpos {
		return 0, fmt.Errorf("buffer: short read (need 8, have %d)", b.Len())
	}
	v := binary.BigEndian.Uint64(b.data[b.rpos:])
	b.rpos += 8
	return v, nil
}

// ReadInt16 / ReadInt32: signed wrappers.
func (b *Buffer) ReadInt16() (int16, error) {
	v, err := b.ReadUint16()
	return int16(v), err
}
func (b *Buffer) ReadInt32() (int32, error) {
	v, err := b.ReadUint32()
	return int32(v), err
}

// WriteSmart08_16 writes a "smart" length: 1 byte if value < 128,
// otherwise 2 bytes where the high bit of the first byte is set.
// Mirrors PacketBuilder.writeSmart08_16 (Java side).
func (b *Buffer) WriteSmart08_16(v int) {
	if v < 128 {
		b.WriteByte(byte(v))
	} else {
		b.WriteUint16(uint16(v + 32768))
	}
}

// RemainingBytes returns the bytes from the current read position to
// the end of the buffer without advancing the cursor. Useful for
// payloads whose tail is a variable-length blob (e.g., RSC-compressed
// chat bodies).
func (b *Buffer) RemainingBytes() []byte {
	return b.data[b.rpos:b.wpos]
}

// ReadStringTerm reads bytes up to (and including) the given terminator
// and returns the string without the terminator. Errors if EOF before
// terminator.
func (b *Buffer) ReadStringTerm(term byte) (string, error) {
	for i := b.rpos; i < b.wpos; i++ {
		if b.data[i] == term {
			s := string(b.data[b.rpos:i])
			b.rpos = i + 1
			return s, nil
		}
	}
	return "", fmt.Errorf("buffer: unterminated string (terminator 0x%02x not found)", term)
}

// ----- Bit-level access (for inbound update packets) -----
//
// RSC's update packets pack flags into individual bits, most-significant-bit
// first within each byte. StartBitAccess / FinishBitAccess delimit a
// section where bit-level reads/writes are valid.

// StartBitAccess switches the buffer into bit-read/bit-write mode.
func (b *Buffer) StartBitAccess() {
	b.bitMode = true
	b.bitPos = b.rpos * 8
}

// FinishBitAccess returns the buffer to byte mode. The read position is
// advanced to the next byte boundary.
func (b *Buffer) FinishBitAccess() {
	if !b.bitMode {
		return
	}
	b.rpos = (b.bitPos + 7) / 8
	b.bitMode = false
}

// ReadBits reads n bits (1..32) as an unsigned integer (MSB first).
func (b *Buffer) ReadBits(n int) (uint32, error) {
	if !b.bitMode {
		return 0, fmt.Errorf("buffer: bit access not started")
	}
	if n < 1 || n > 32 {
		return 0, fmt.Errorf("buffer: ReadBits(%d) out of range", n)
	}
	bytePos := b.bitPos >> 3
	bitOffset := 8 - (b.bitPos & 7)
	var value uint32
	for n > 0 {
		if bytePos >= b.wpos {
			return 0, fmt.Errorf("buffer: bit-read past end")
		}
		take := bitOffset
		if take > n {
			take = n
		}
		shift := bitOffset - take
		mask := uint32((1 << take) - 1)
		value = (value << take) | ((uint32(b.data[bytePos]) >> shift) & mask)
		b.bitPos += take
		n -= take
		bitOffset -= take
		if bitOffset == 0 {
			bitOffset = 8
			bytePos++
		}
	}
	return value, nil
}
