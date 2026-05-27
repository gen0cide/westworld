package v235

import "testing"

// TestIsaacSymmetric verifies that encoding then decoding with two
// instances seeded identically produces the original byte.
//
// Both encoder and decoder advance their internal state in lockstep,
// so this is the canonical sanity check.
func TestIsaacSymmetric(t *testing.T) {
	keys := []uint32{0x12345678, 0x9abcdef0, 0xdeadbeef, 0xcafebabe}

	enc := NewIsaac()
	enc.SetKeys(keys)
	dec := NewIsaac()
	dec.SetKeys(keys)

	for op := 0; op < 256; op++ {
		encoded := enc.EncodeOpcode(byte(op))
		decoded := dec.DecodeOpcode(encoded)
		if decoded != byte(op) {
			t.Fatalf("opcode %d: enc=%d decoded=%d (want %d)", op, encoded, decoded, op)
		}
	}
}

// TestIsaacDeterministic verifies the keystream is deterministic
// given the same seed. Two cipher instances with the same seed should
// produce the same sequence of NextValue results.
func TestIsaacDeterministic(t *testing.T) {
	keys := []uint32{1, 2, 3, 4}

	a := NewIsaac()
	a.SetKeys(keys)
	b := NewIsaac()
	b.SetKeys(keys)

	for i := 0; i < 1000; i++ {
		av := a.NextValue()
		bv := b.NextValue()
		if av != bv {
			t.Fatalf("step %d: %d vs %d", i, av, bv)
		}
	}
}

// TestIsaacDifferentSeeds verifies different seeds produce different
// keystreams (basic sanity — not a cryptographic test).
func TestIsaacDifferentSeeds(t *testing.T) {
	a := NewIsaac()
	a.SetKeys([]uint32{1, 2, 3, 4})
	b := NewIsaac()
	b.SetKeys([]uint32{1, 2, 3, 5})

	// Within the first 256 values, at least one should differ.
	differ := false
	for i := 0; i < 256; i++ {
		if a.NextValue() != b.NextValue() {
			differ = true
			break
		}
	}
	if !differ {
		t.Fatalf("two different seeds produced identical first-256 keystream")
	}
}

// TestFrameDecoderWithIsaac verifies that a frame encoded with one
// Isaac instance can be decoded by a matching Isaac instance via the
// FrameDecoder.SetDecode plumbing.
func TestFrameDecoderWithIsaac(t *testing.T) {
	keys := []uint32{0x12345678, 0x9abcdef0, 0xdeadbeef, 0xcafebabe}

	encCipher := NewIsaac()
	encCipher.SetKeys(keys)
	decCipher := NewIsaac()
	decCipher.SetKeys(keys)

	// Encode a walk-to-point packet with ISAAC.
	payload := []byte{0, 120, 1, 248} // x=120, y=504
	wire := EncodeFrame(encCipher.EncodeOpcode(OutWalkToPoint), payload)

	dec := NewFrameDecoder(decCipher.DecodeOpcode)
	dec.Feed(wire)
	frame, err := dec.Next()
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if frame.Opcode != OutWalkToPoint {
		t.Errorf("opcode: got 0x%x, want 0x%x", frame.Opcode, OutWalkToPoint)
	}
	if string(frame.Payload) != string(payload) {
		t.Errorf("payload: got %v, want %v", frame.Payload, payload)
	}
}
