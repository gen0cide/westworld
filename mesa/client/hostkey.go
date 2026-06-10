package mesaclient

import "github.com/gen0cide/westworld/mesa/auth"

// HostKey derives a host's mesa bearer token from its username. Thin alias for
// mesa/auth.HostKey — the derivation lives in the shared leaf so the SERVER
// can derive expected keys without importing this client package.
func HostKey(username string) string { return auth.HostKey(username) }
