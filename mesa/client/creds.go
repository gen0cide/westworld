package mesaclient

import "context"

// tokenCreds attaches a host's bearer token to every RPC (unary + stream) as
// gRPC per-RPC credentials. mesa's auth interceptor resolves the token to a
// host_id and binds it into the request context, so the host's identity is
// authenticated at the connection — not self-asserted in the request body.
//
// RequireTransportSecurity is false because the host↔mesa link is insecure for
// now (trusted/local). The production path is TLS/mTLS: flip this to true (and,
// with mTLS, the client cert can replace the token as the identity).
type tokenCreds struct{ token string }

func (t tokenCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	if t.token == "" {
		return nil, nil
	}
	return map[string]string{"authorization": "Bearer " + t.token}, nil
}

func (tokenCreds) RequireTransportSecurity() bool { return false }
