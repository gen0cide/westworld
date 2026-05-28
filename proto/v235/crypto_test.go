package v235

import (
	"bytes"
	"testing"
)

// TestXTEARoundTrip verifies encrypt/decrypt is symmetric.
func TestXTEARoundTrip(t *testing.T) {
	keys := [4]uint32{0xDEADBEEF, 0x12345678, 0xCAFEBABE, 0xABCDEF01}

	cases := []struct {
		name string
		data []byte
	}{
		{"single block (8 bytes)", []byte("hello123")},
		{"two blocks (16 bytes)", []byte("hello123world456")},
		{"four blocks (32 bytes)", bytes.Repeat([]byte("X"), 32)},
		{"unaligned would error", nil}, // covered in TestXTEAUnaligned
	}
	for _, tc := range cases {
		if tc.data == nil {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			original := append([]byte(nil), tc.data...)
			if err := XTEAEncrypt(tc.data, keys); err != nil {
				t.Fatalf("encrypt: %v", err)
			}
			if bytes.Equal(tc.data, original) {
				t.Fatalf("encrypt produced identical bytes")
			}
			if err := XTEADecrypt(tc.data, keys); err != nil {
				t.Fatalf("decrypt: %v", err)
			}
			if !bytes.Equal(tc.data, original) {
				t.Fatalf("round trip mismatch:\ngot:  %v\nwant: %v", tc.data, original)
			}
		})
	}
}

// TestXTEAUnaligned verifies that non-multiple-of-8 input errors cleanly.
func TestXTEAUnaligned(t *testing.T) {
	keys := [4]uint32{1, 2, 3, 4}
	if err := XTEAEncrypt(make([]byte, 7), keys); err == nil {
		t.Fatalf("expected error for 7-byte input")
	}
	if err := XTEAEncrypt(make([]byte, 9), keys); err == nil {
		t.Fatalf("expected error for 9-byte input")
	}
}

// TestRSAEncryptDeterminism verifies that the same plaintext produces
// the same ciphertext (raw RSA is deterministic, unlike padded RSA).
func TestRSAEncryptDeterminism(t *testing.T) {
	pub := DefaultServerRSA()
	plaintext := bytes.Repeat([]byte{0x42}, 32)
	a, err := pub.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt #1: %v", err)
	}
	b, err := pub.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt #2: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Errorf("raw RSA not deterministic")
	}
}

// TestRSAEncryptSize verifies the output is roughly the modulus size
// (64 or 65 bytes for a 512-bit key).
func TestRSAEncryptSize(t *testing.T) {
	pub := DefaultServerRSA()
	plaintext := make([]byte, 61)
	plaintext[0] = 10 // checksum
	for i := 1; i < 61; i++ {
		plaintext[i] = byte(i)
	}
	ct, err := pub.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(ct) < 60 || len(ct) > 66 {
		t.Errorf("ciphertext size unusual: %d bytes (expected ~64)", len(ct))
	}
}

// TestEncodePassword exercises the 20-byte password padding behavior.
func TestEncodePassword(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hunter2pass", "hunter2pass         "},
		{"short", "short               "},
		{"with space", "with_space          "},
		{"exact_at_twenty_chrs", "exact_at_twenty_chrs"},
		{"way_too_long_to_fit_at_twenty_chars_total", "way_too_long_to_fit_"},
	}
	for _, tc := range cases {
		got := encodePassword(tc.in)
		if string(got) != tc.want {
			t.Errorf("encodePassword(%q) = %q, want %q", tc.in, string(got), tc.want)
		}
		if len(got) != 20 {
			t.Errorf("encodePassword(%q) length = %d, want 20", tc.in, len(got))
		}
	}
}

// TestLoginPayloadEncodesAtPlausibleSize is a sanity check: the
// resulting on-wire frame should be ~107-115 bytes for a short
// username, well under 160 so the tail-byte path applies.
func TestLoginPayloadEncodesAtPlausibleSize(t *testing.T) {
	p := &LoginPayload{
		Username:      "alex",
		Password:      "hunter2pass",
		ClientVersion: 235,
	}
	pub := DefaultServerRSA()
	wire, err := p.EncodeLoginFrame(pub)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(wire) < 100 || len(wire) > 200 {
		t.Errorf("wire size unusual: %d bytes", len(wire))
	}
	// First byte: length. Second byte: tail. Third byte: opcode = 0.
	if wire[2] != 0 {
		t.Errorf("opcode position: got 0x%x, want 0x00 (plain LOGIN)", wire[2])
	}
}
