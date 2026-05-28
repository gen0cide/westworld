package v235

// TryDecodeRSCString decompresses an RSC-encoded chat message of the
// given expected character count. Delegates to DecipherRSCString
// (see stringencryption.go).
func TryDecodeRSCString(compressed []byte, expectedLen int) string {
	return DecipherRSCString(compressed, expectedLen)
}
