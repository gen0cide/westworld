package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/session"
)

// resetSession is a minimal game-server session whose only job is to make an
// account ONLINE so the admin API can address it: the API resolves players
// through the live world (AdminRouter.playerByName), so an offline player is
// unreachable until someone logs it in. The session deliberately keeps no
// world-state mirror — a client mirror can never see the bank without opening
// a banker anyway — and all verification reads go through the admin API,
// which is authoritative for skills, inventory, AND bank.
type resetSession struct {
	conn *session.Conn
}

// openSession dials the game server and performs the v235 login handshake as
// username. The password is passed by value and is never logged or echoed —
// login failures surface only the server's 1-byte response code. Inbound
// frames are drained in the background so the server's pushes (stat updates,
// inventory packets from the admin mutations) don't back up the read loop.
func openSession(ctx context.Context, server, username, password string, log *slog.Logger) (*resetSession, error) {
	conn, err := session.Dial(ctx, server, session.Options{Logger: log})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", server, err)
	}
	if _, err := conn.Login(ctx, session.LoginParams{Username: username, Password: password}); err != nil {
		_ = conn.Close()
		return nil, err // carries only the response code, never credentials
	}
	conn.Start()
	go func() {
		for range conn.Recv() {
		}
	}()
	return &resetSession{conn: conn}, nil
}

// close sends a clean logout and tears the connection down. The brief wait
// gives the server time to process the logout request (it saves the player on
// unregister either way; a hard close would also be saved, just less tidily).
func (s *resetSession) close(ctx context.Context) {
	if s == nil || s.conn == nil {
		return
	}
	_ = action.Logout(ctx, s.conn)
	select {
	case <-time.After(500 * time.Millisecond):
	case <-s.conn.Done():
	case <-ctx.Done():
	}
	_ = s.conn.Close()
}
