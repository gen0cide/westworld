package v235

import (
	"bytes"
	"testing"
)

// TestEncodeFrame_OpcodeOnly verifies the length==1 case:
// just [length=1] [encOpcode]. From RSCProtocolEncoderMain.java:93-97.
func TestEncodeFrame_OpcodeOnly(t *testing.T) {
	got := EncodeFrame(OutHeartbeat, nil)
	want := []byte{1, OutHeartbeat}
	if !bytes.Equal(got, want) {
		t.Errorf("opcode-only frame: got %v, want %v", got, want)
	}
}

// TestEncodeFrame_ShortPacket verifies the tail-byte reordering for
// packets with length < 160. From RSCProtocolEncoderMain.java:75-92.
//
// For walk-to-point with payload [0, 100, 0, 200] (x=100, y=200):
//   length = 5, payload last byte = 200 (0xC8)
//   Wire: [5] [200] [encOpcode] [0, 100, 0]
func TestEncodeFrame_ShortPacket(t *testing.T) {
	payload := []byte{0, 100, 0, 200}
	got := EncodeFrame(OutWalkToPoint, payload)
	want := []byte{5, 200, OutWalkToPoint, 0, 100, 0}
	if !bytes.Equal(got, want) {
		t.Errorf("short packet: got %v, want %v", got, want)
	}
}

// TestEncodeFrame_LargePacket verifies the 2-byte-length case for
// length >= 160. From RSCProtocolEncoderMain.java:60-69.
func TestEncodeFrame_LargePacket(t *testing.T) {
	payload := make([]byte, 200) // length = 201 > 160
	for i := range payload {
		payload[i] = byte(i)
	}
	got := EncodeFrame(0xAB, payload)

	// length = 201; high = 201/256 + 160 = 0 + 160 = 160
	//               low = 201 & 0xFF = 0xC9
	if got[0] != 160 || got[1] != 0xC9 {
		t.Fatalf("length header: got [%d, %d], want [160, 201]", got[0], got[1])
	}
	if got[2] != 0xAB {
		t.Fatalf("opcode: got 0x%x, want 0xAB", got[2])
	}
	if !bytes.Equal(got[3:], payload) {
		t.Fatalf("payload mismatch (large packet should NOT use tail-byte reordering)")
	}
}

// TestRoundTrip verifies decode(encode(x)) == x for various sizes.
func TestRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		opcode  byte
		payload []byte
	}{
		{"opcode-only", OutHeartbeat, nil},
		{"small payload", OutWalkToPoint, []byte{0, 100, 0, 200}},
		{"medium payload", 42, makePayload(50)},
		{"159-byte length (boundary, still 1-byte length)", 99, makePayload(158)},
		{"160-byte length (boundary, now 2-byte length)", 99, makePayload(159)},
		{"large payload", 50, makePayload(500)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire := EncodeFrame(tc.opcode, tc.payload)
			dec := NewFrameDecoder(PlainDecode)
			dec.Feed(wire)
			frame, err := dec.Next()
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if frame.Opcode != tc.opcode {
				t.Errorf("opcode: got 0x%x, want 0x%x", frame.Opcode, tc.opcode)
			}
			if !bytes.Equal(frame.Payload, tc.payload) {
				t.Errorf("payload mismatch\ngot:  %v\nwant: %v", frame.Payload, tc.payload)
			}
		})
	}
}

// TestDecoder_PartialFeed verifies that the decoder correctly handles
// bytes arriving across multiple Feed calls.
func TestDecoder_PartialFeed(t *testing.T) {
	wire := EncodeFrame(OutWalkToPoint, []byte{0, 100, 0, 200})
	dec := NewFrameDecoder(PlainDecode)

	// Feed bytes one at a time. Should not produce a frame until
	// all bytes are present.
	for i := 0; i < len(wire)-1; i++ {
		dec.Feed(wire[i : i+1])
		_, err := dec.Next()
		if err != ErrShortRead {
			t.Fatalf("after %d bytes fed: got err=%v, want ErrShortRead", i+1, err)
		}
	}

	// Feed the last byte.
	dec.Feed(wire[len(wire)-1:])
	frame, err := dec.Next()
	if err != nil {
		t.Fatalf("final feed: %v", err)
	}
	if frame.Opcode != OutWalkToPoint {
		t.Errorf("got opcode 0x%x, want 0x%x", frame.Opcode, OutWalkToPoint)
	}
}

// TestDecoder_ConcatenatedFrames verifies that multiple frames in one
// feed are decoded sequentially.
func TestDecoder_ConcatenatedFrames(t *testing.T) {
	w1 := EncodeFrame(OutHeartbeat, nil)
	w2 := EncodeFrame(OutWalkToPoint, []byte{0, 100, 0, 200})
	w3 := EncodeFrame(OutLogout, nil)
	combined := append(append(append([]byte(nil), w1...), w2...), w3...)

	dec := NewFrameDecoder(PlainDecode)
	dec.Feed(combined)
	frames, err := dec.ReadAllFrames()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if len(frames) != 3 {
		t.Fatalf("got %d frames, want 3", len(frames))
	}
	if frames[0].Opcode != OutHeartbeat || frames[1].Opcode != OutWalkToPoint || frames[2].Opcode != OutLogout {
		t.Errorf("frame opcodes: %v", []byte{frames[0].Opcode, frames[1].Opcode, frames[2].Opcode})
	}
}

func makePayload(n int) []byte {
	p := make([]byte, n)
	for i := range p {
		p[i] = byte(i)
	}
	return p
}
