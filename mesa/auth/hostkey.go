// Package auth is the tiny shared leaf for host<->mesa identity derivation —
// imported by BOTH the client and the server so neither imports the other
// (mesad importing mesa/client for one hash was the module's dirtiest edge).
package auth

import (
	"crypto/sha512"
	"encoding/hex"
)

// HostKey derives a host's mesa bearer token from its username. For now this is
// just SHA-512(username) so a host can derive its own key with no out-of-band
// issuance, and mesa can derive the expected key per registered host.
//
// NOTE: this is a PLACEHOLDER, not a secret — the username is not confidential,
// so anyone who knows it can derive the key. It exists to wire identity binding
// through the stack now; replace with a real per-host secret (or mTLS client
// identity) before this is exposed beyond a trusted/local link.
func HostKey(username string) string {
	sum := sha512.Sum512([]byte(username))
	return hex.EncodeToString(sum[:])
}
