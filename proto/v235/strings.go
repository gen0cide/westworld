package v235

import "fmt"

// ReadZeroQuotedString reads a string in OpenRSC's "zero-quoted"
// format: a leading 0x00 byte, UTF-8 bytes of the string, and a
// trailing 0x00 byte. Used for chat sender names, NPC dialog text,
// and most server-emitted text.
//
// Mirrors PacketBuilder.java:309-313 writeZeroQuotedString.
func (b *Buffer) ReadZeroQuotedString() (string, error) {
	// Discard leading 0x00 if present (defensive: a few packets seem
	// to omit it).
	if b.Len() > 0 && b.data[b.rpos] == 0 {
		b.rpos++
	}
	s, err := b.ReadStringTerm(0)
	if err != nil {
		return "", fmt.Errorf("read zero-quoted string: %w", err)
	}
	return s, nil
}

// ReadZeroPaddedString reads until a 0 byte (or end of buffer)
// without consuming a leading byte. Used for admin commands.
func (b *Buffer) ReadZeroPaddedString() (string, error) {
	for i := b.rpos; i < b.wpos; i++ {
		if b.data[i] == 0 {
			s := string(b.data[b.rpos:i])
			b.rpos = i + 1
			return s, nil
		}
	}
	// No terminator: return remaining bytes.
	s := string(b.data[b.rpos:b.wpos])
	b.rpos = b.wpos
	return s, nil
}

// WriteZeroPaddedString writes a string in the format expected by
// OpenRSC's Packet.readZeroPaddedString (Packet.java:173-182): a
// leading 0x00 byte, UTF-8 content, and a trailing 0x00 byte. Used
// for admin commands.
//
// Note: despite the name "padded" vs "quoted" in OpenRSC's API, the
// reader requires the leading 0x00 byte — the same shape as
// WriteZeroQuotedString. The naming is historical.
func (b *Buffer) WriteZeroPaddedString(s string) {
	b.WriteByte(0)
	b.WriteBytes([]byte(s))
	b.WriteByte(0)
}

// ReadSmart08_16 reads a "smart" value: 1 byte if < 128, otherwise 2
// bytes where (value - 32768) is the actual int. Used for some payload
// lengths.
//
// Mirrors Packet.java:getSmart08_16.
func (b *Buffer) ReadSmart08_16() (int, error) {
	if b.Len() < 1 {
		return 0, fmt.Errorf("smart08_16: empty")
	}
	first := int(b.data[b.rpos])
	if first < 128 {
		b.rpos++
		return first, nil
	}
	// 2-byte form: first byte's value is offset by 128, real value
	// is ((first - 128) << 8) | second. But OpenRSC's exact formula
	// is `(getUnsignedShort() & 0xFFFF) + 32768`, so we read 2 bytes
	// and subtract 32768.
	v, err := b.ReadUint16()
	if err != nil {
		return 0, err
	}
	return int(v) - 32768, nil
}
