package v235

// TryDecodeRSCString attempts to decode an RSC-compressed chat
// message. The format uses a fixed character-frequency table and
// variable-length bit codes. Phase 1.6 implements a best-effort
// decoder that gets the common chat path right; some edge cases
// (non-ASCII, special punctuation) may produce garbled output.
//
// Source: util/rsc/DataConversions.java:350-390 (getEncryptedString
// and getStringFromBytes) plus the encryption.decryptString class.
//
// expectedLen is the smart-prefixed length the sender claimed; we
// decode up to that many characters, then stop.
//
// Returns the decoded string, or "" if decoding failed (caller should
// fall back to MessageRaw).
func TryDecodeRSCString(compressed []byte, expectedLen int) string {
	if expectedLen <= 0 || len(compressed) == 0 {
		return ""
	}
	// Phase 1.6: stub decoder. The full algorithm requires the
	// character-frequency tree from OpenRSC's encryption class — to
	// be ported in Phase 1.7. For now we return an empty string;
	// callers should display MessageRaw as a fallback or note that
	// the body is undecoded.
	return ""
}
