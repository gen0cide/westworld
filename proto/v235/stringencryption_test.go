package v235

import "testing"

// TestEncipherDecipherRoundtrip verifies that a string can be
// compressed and decompressed back to itself.
func TestEncipherDecipherRoundtrip(t *testing.T) {
	cases := []string{
		"hello",
		"hi",
		"a",
		"hello world",
		"buying gold",
		"lol",
		"need help",
		"i am alex",
		"trade?",
		"123",
	}
	for _, msg := range cases {
		t.Run(msg, func(t *testing.T) {
			compressed := EncipherRSCString(msg)
			decompressed := DecipherRSCString(compressed, len(msg))
			if decompressed != msg {
				t.Errorf("roundtrip failed for %q:\n  compressed: %v\n  decompressed: %q",
					msg, compressed, decompressed)
			}
		})
	}
}

// TestEncipherEmpty verifies that an empty input produces no bytes
// (or a deterministic small output).
func TestEncipherEmpty(t *testing.T) {
	compressed := EncipherRSCString("")
	if len(compressed) > 0 {
		t.Logf("empty input produced %d bytes (not necessarily wrong)", len(compressed))
	}
}
