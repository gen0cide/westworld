package mesad

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ctxHostIDKey is the context key under which the AUTHENTICATED host_id is bound
// after the auth interceptor runs. Handlers read it via hostIDFromContext and
// must never trust a host_id from the request body.
type ctxHostIDKey struct{}

// hostIDFromContext returns the authenticated host_id bound by the interceptor.
func hostIDFromContext(ctx context.Context) string {
	s, _ := ctx.Value(ctxHostIDKey{}).(string)
	return s
}

// Authorize registers a bearer token → host_id mapping. Issued out-of-band (one
// per host); a host presents it on connect to authenticate as that host_id.
func (s *Server) Authorize(hostID, token string) {
	if token == "" {
		return
	}
	s.mu.Lock()
	s.tokens[token] = hostID
	s.mu.Unlock()
}

// authenticate resolves the bearer token from request metadata to a host_id and
// binds it into the context. It rejects missing/invalid tokens with
// Unauthenticated, so every RPC is tied to a known host.
func (s *Server) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "missing metadata")
	}
	token := bearer(md.Get("authorization"))
	if token == "" {
		return ctx, status.Error(codes.Unauthenticated, "missing bearer token")
	}
	s.mu.RLock()
	hostID, ok := s.tokens[token]
	s.mu.RUnlock()
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "invalid token")
	}
	return context.WithValue(ctx, ctxHostIDKey{}, hostID), nil
}

func bearer(vals []string) string {
	if len(vals) == 0 {
		return ""
	}
	v := strings.TrimSpace(vals[0])
	if len(v) >= 7 && strings.EqualFold(v[:7], "Bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return v
}

// UnaryAuth is the unary server interceptor: authenticate, then bind the
// host_id into the handler's context.
func (s *Server) UnaryAuth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	authed, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}
	return handler(authed, req)
}

// authStream wraps a ServerStream to expose the authenticated context.
type authStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (a authStream) Context() context.Context { return a.ctx }

// StreamAuth is the streaming server interceptor: authenticate, then run the
// handler with a stream whose Context carries the authenticated host_id.
func (s *Server) StreamAuth(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	authed, err := s.authenticate(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, authStream{ServerStream: ss, ctx: authed})
}
