package runtime

import (
	"context"

	"github.com/gen0cide/westworld/action"
)

// PrivateMessage sends a /tell-style private message to a named
// player. Recipient must be online AND have us on their friends
// list; the server silently drops the packet otherwise.
func (h *Host) PrivateMessage(ctx context.Context, recipient, message string) error {
	return action.PrivateMessage(ctx, h.conn, recipient, message)
}

// AddFriend adds a player to the bot's friends list. Required before
// PM exchange will work.
func (h *Host) AddFriend(ctx context.Context, name string) error {
	return action.AddFriend(ctx, h.conn, name)
}

// RemoveFriend removes a player from the bot's friends list.
func (h *Host) RemoveFriend(ctx context.Context, name string) error {
	return action.RemoveFriend(ctx, h.conn, name)
}

// AddIgnore adds a player to the bot's ignore list. The server replies with a
// full ignore-list (109) packet, which World.Apply mirrors into world.Social.
func (h *Host) AddIgnore(ctx context.Context, name string) error {
	return action.AddIgnore(ctx, h.conn, name)
}

// RemoveIgnore removes a player from the bot's ignore list.
func (h *Host) RemoveIgnore(ctx context.Context, name string) error {
	return action.RemoveIgnore(ctx, h.conn, name)
}
