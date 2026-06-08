package action

import (
	"testing"

	"github.com/gen0cide/westworld/proto/v235"
)

// TestSanitizeChatFoldsLLMUnicode: the characters LLMs habitually emit fold
// to clean RSC-charset ASCII instead of '?'.
func TestSanitizeChatFoldsLLMUnicode(t *testing.T) {
	got := sanitizeChat("I’ll meet you — bring “gold”…")
	want := "I'll meet you - bring \"gold\"..."
	if got != want {
		t.Fatalf("sanitizeChat:\n got %q\nwant %q", got, want)
	}
}

// TestChatCountIsRuneCountNotByteCount locks the crash fix. The server decodes
// exactly the smart-length character count; EncipherRSCString encodes one code
// per RUNE. Sending len(message) (Go bytes) for a multi-byte rune told the
// server to decode MORE characters than were encoded, overrunning its Huffman
// decoder (StringEncryption.decryptString ArrayIndexOutOfBounds). The count
// must be the rune count.
func TestChatCountIsRuneCountNotByteCount(t *testing.T) {
	// "café stuff": é is 2 UTF-8 bytes / 1 rune, and has no sanitize mapping,
	// so the message stays multi-byte after sanitization.
	msg := sanitizeChat("café stuff")
	nbytes, nchars := len(msg), len([]rune(msg))
	if nbytes == nchars {
		t.Fatalf("test precondition: wanted a multi-byte message, got bytes=%d chars=%d", nbytes, nchars)
	}
	// The rune count round-trips through the codec; the byte count (the old
	// bug) would have over-declared the length and overrun the decoder.
	enc := v235.EncipherRSCString(msg)
	if got := v235.DecipherRSCString(enc, nchars); len([]rune(got)) != nchars {
		t.Fatalf("rune-count round-trip: decoded %d runes, want %d", len([]rune(got)), nchars)
	}
}
