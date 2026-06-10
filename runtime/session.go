package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/action"
)

// Logout is a high-level convenience that proxies to action.Logout.
// It sends a single logout request and does NOT wait for the server
// to acknowledge — use LogoutGraceful when you need the session
// actually released (e.g., before a fast re-login of the same
// account). The OpenRSC server refuses logout while the player is
// in combat / dueling / busy, or within ~10s of combat
// (Player.canLogout), so a fire-and-forget Logout is frequently
// ignored by the server.
func (h *Host) Logout(ctx context.Context) error {
	return action.Logout(ctx, h.conn)
}

// LogoutGraceful sends a logout request and waits for the server to
// actually release the session, retrying periodically until released
// or maxWait elapses.
//
// Why this exists: the server's Player.canLogout() returns false
// while the host is inCombat()/dueling/busy and for ~10s after the
// combat timer. A scenario that ends mid-combat therefore cannot log
// out immediately; if the cradle just hard-closes the TCP socket,
// the server holds the session until its connection reaper fires,
// and a same-account re-login in that window is rejected with login
// code 4 ("already logged in"). Riding out the cooldown and getting
// the server to release the session frees it immediately.
//
// Signal: the server does NOT send a logout-confirm packet on a clean
// logout — when it accepts (canLogout()==true), it unregisters the
// player and closes the socket. The session read loop turns that EOF
// into conn.Done() closing. So "logout accepted" == conn.Done() fired.
// When logout is REFUSED (in combat), the server silently keeps the
// connection open; we just retry on the next tick until the cooldown
// passes and a resend is accepted.
//
// Returns nil once the connection is released. Returns an error if
// maxWait passes with the connection still open (caller should
// hard-close anyway — the reaper will eventually release it). Honors
// ctx cancellation.
func (h *Host) LogoutGraceful(ctx context.Context, maxWait time.Duration) error {
	if h.conn == nil {
		return nil
	}
	deadline := time.Now().Add(maxWait)
	const retryEvery = 1200 * time.Millisecond
	for {
		// (Re)send the logout request. A transient write error is
		// ignored — either the connection just closed (we'll catch it
		// on the next select via conn.Done) or we retry next tick.
		_ = action.Logout(ctx, h.conn)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-h.conn.Done():
			// Server released the session and closed the socket.
			return nil
		case <-time.After(retryEvery):
			if !time.Now().Before(deadline) {
				return fmt.Errorf("logout not accepted within %s (connection still open — host likely in combat / busy)", maxWait)
			}
		}
	}
}

// Command sends an admin command (without the leading "::").
func (h *Host) Command(ctx context.Context, cmd string) error {
	return action.Command(ctx, h.conn, cmd)
}

// Say sends a public chat message (RSC-compressed under the hood). On a
// successful send it captures the line into the reactive tier's latched windows
// (the single self-line seam — all callers including socialReflex funnel here),
// so a Q&A pairs up when the host replies to a latched speaker.
func (h *Host) Say(ctx context.Context, message string) error {
	if err := action.Say(ctx, h.conn, message); err != nil {
		return err
	}
	h.reactiveObserveSelf(message)
	return nil
}

// Close shuts down the host: first a BEST-EFFORT clean RSC logout (so the server
// saves + releases the session instead of timing out a dropped socket — which
// otherwise blocks a same-account re-login with "already logged in"), then the
// socket and the event bus. LogoutGraceful rides out the combat-logout cooldown
// up to its bound, and is a no-op when the connection is already gone.
func (h *Host) Close() error {
	if h.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		_ = h.LogoutGraceful(ctx, 12*time.Second)
		cancel()
		h.conn.Close()
	}
	h.bus.Close()
	return nil
}
