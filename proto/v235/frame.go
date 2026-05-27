package v235

import (
	"errors"
	"fmt"
	"io"
)

// Frame is one decoded packet: opcode plus payload bytes (payload does
// NOT include the opcode byte).
type Frame struct {
	Opcode  byte
	Payload []byte
}

// ErrShortRead indicates not enough bytes are available yet to decode
// a complete frame. Callers should buffer more bytes and try again.
var ErrShortRead = errors.New("v235: short read")

// EncodeFrame encodes an opcode + payload into wire bytes, applying
// OpenRSC's tail-byte-reordering and length-encoding quirks.
//
// Ported from RSCProtocolEncoderMain.java:55-100.
//
// Behaviour:
//
//   - length = len(payload) + 1
//   - if length == 1: [length=1] [encOpcode]
//   - if length < 160: [length] [payload[N-1]] [encOpcode] [payload[0..N-2]]
//   - if length >= 160: [(length/256)+160] [length&0xFF] [encOpcode] [payload]
//
// encOpcode is the ISAAC-encoded opcode for this packet (caller is
// responsible for computing it via Isaac.EncodeOpcode); pass the plain
// opcode here only for pre-ISAAC packets like the login request.
func EncodeFrame(encOpcode byte, payload []byte) []byte {
	n := len(payload)
	length := n + 1

	if length >= 160 {
		// 2-byte length, opcode, full payload.
		out := make([]byte, 0, 3+n)
		out = append(out, byte(length/256+160))
		out = append(out, byte(length&0xFF))
		out = append(out, encOpcode)
		out = append(out, payload...)
		return out
	}

	if length == 1 {
		// Opcode-only packet, no tail byte.
		return []byte{1, encOpcode}
	}

	// length < 160 and n > 0: tail-byte reordering. The LAST byte of
	// the payload is moved between the length byte and the opcode.
	out := make([]byte, 0, 2+n)
	out = append(out, byte(length))
	out = append(out, payload[n-1])
	out = append(out, encOpcode)
	out = append(out, payload[:n-1]...)
	return out
}

// FrameDecoder is a stateful decoder that consumes bytes from a
// connection and emits Frames. It owns no goroutines — callers feed
// bytes via Feed and pull frames via Next.
//
// Ported (semantically) from RSCProtocolDecoder.java:147-238.
type FrameDecoder struct {
	// buf holds bytes received from the network that have not yet
	// been consumed by a frame.
	buf []byte

	// decode is the opcode-decryption function. Before ISAAC is
	// active, this should be the identity transform. After ISAAC
	// activation, it should call Isaac.DecodeOpcode.
	decode func(enc byte) byte
}

// NewFrameDecoder returns a decoder with the given opcode-decode
// function. Pass v235.PlainDecode before ISAAC is active.
func NewFrameDecoder(decode func(enc byte) byte) *FrameDecoder {
	if decode == nil {
		decode = PlainDecode
	}
	return &FrameDecoder{decode: decode}
}

// PlainDecode is the identity function for the opcode byte — used
// before ISAAC is initialized.
func PlainDecode(b byte) byte { return b }

// SetDecode replaces the opcode-decode function. Call this after the
// login response is accepted and ISAAC has been seeded.
func (d *FrameDecoder) SetDecode(fn func(enc byte) byte) {
	d.decode = fn
}

// Feed appends incoming bytes to the decoder's buffer.
func (d *FrameDecoder) Feed(b []byte) {
	d.buf = append(d.buf, b...)
}

// Next attempts to decode one frame from the buffered bytes. If not
// enough bytes are available, ErrShortRead is returned and the
// caller should Feed more bytes.
func (d *FrameDecoder) Next() (Frame, error) {
	if len(d.buf) < 1 {
		return Frame{}, ErrShortRead
	}

	first := int(d.buf[0])
	var length, lengthLen int

	if first >= 160 {
		if len(d.buf) < 2 {
			return Frame{}, ErrShortRead
		}
		second := int(d.buf[1])
		length = (first-160)*256 + second
		lengthLen = 2
	} else {
		length = first
		lengthLen = 1
	}

	if length < 1 {
		// Drop the bad length byte and try again. Should not happen.
		d.buf = d.buf[lengthLen:]
		return Frame{}, fmt.Errorf("v235: invalid frame length %d", length)
	}

	totalBytes := lengthLen + length
	if len(d.buf) < totalBytes {
		return Frame{}, ErrShortRead
	}

	body := d.buf[lengthLen:totalBytes]
	d.buf = d.buf[totalBytes:]

	if lengthLen == 1 && length > 1 {
		// Tail-byte reordering inverse.
		//
		// On the wire, body = [tail_byte] [encOpcode] [payload[0..N-2]]
		// After inversion, packet = [opcode] [payload[0..N-2]] [tail_byte]
		// So payload (without opcode) = [payload[0..N-2]] [tail_byte]
		tail := body[0]
		encOp := body[1]
		mid := body[2:]
		op := d.decode(encOp)

		payload := make([]byte, 0, length-1)
		payload = append(payload, mid...)
		payload = append(payload, tail)

		return Frame{Opcode: op, Payload: payload}, nil
	}

	// length == 1 (opcode only) or length >= 160 (no tail-byte).
	encOp := body[0]
	payload := append([]byte(nil), body[1:]...)
	return Frame{Opcode: d.decode(encOp), Payload: payload}, nil
}

// ReadAllFrames is a convenience to drain a buffer of all complete
// frames. It returns errors other than ErrShortRead unchanged.
func (d *FrameDecoder) ReadAllFrames() ([]Frame, error) {
	var out []Frame
	for {
		f, err := d.Next()
		if err == ErrShortRead {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		out = append(out, f)
	}
}

// WriteFrame is a convenience for writing one frame to an io.Writer.
// Returns the number of bytes written.
func WriteFrame(w io.Writer, encOpcode byte, payload []byte) (int, error) {
	return w.Write(EncodeFrame(encOpcode, payload))
}
