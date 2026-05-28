package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// SOCIAL_SEND_PRIVATE_MESSAGE outbound opcode and structure:
//   [zero-padded string recipient_username]
//   [RSC-compressed bytes encrypted_message]
//
// Source: Payload235Parser.java case 218 → SOCIAL_SEND_PRIVATE_MESSAGE;
// FriendStruct decoder reads recipient with readZeroPaddedString and
// message body with getEncryptedString (the RSC string-compression
// codec).
const (
	outPrivateMessage byte = 218 // SOCIAL_SEND_PRIVATE_MESSAGE
	outAddFriend      byte = 195 // SOCIAL_ADD_FRIEND
	outRemoveFriend   byte = 167 // SOCIAL_REMOVE_FRIEND
)

// PrivateMessage sends a "/tell"-style private message to the named
// player. The recipient must be online AND must have us on their
// friends list (mutual friending); otherwise the server silently
// drops the packet without notice.
//
// Wire format (per mudclient.putStringPair + RSBufferUtils.putEncryptedString):
//   [zero-padded string recipient]
//   [smart08_16 message-char-count] [RSC-compressed bytes]
//
// The smart-count prefix is critical — without it, the server's
// getEncryptedString reads the first byte of the ciphertext as the
// length and decompresses N bogus characters from the rest.
func PrivateMessage(ctx context.Context, conn *session.Conn, recipient, message string) error {
	if recipient == "" {
		return fmt.Errorf("action: PM recipient empty")
	}
	if message == "" {
		return fmt.Errorf("action: PM message empty")
	}
	body := v235.EncipherRSCString(message)
	buf := v235.NewBuffer(4 + len(recipient) + len(body))
	buf.WriteZeroPaddedString(recipient)
	buf.WriteSmart08_16(len(message))
	buf.WriteBytes(body)
	return conn.Send(outPrivateMessage, buf.Bytes())
}

// AddFriend adds a player to the bot's friend list. Required before
// the bot can send or receive PMs from that player.
//
// Source: Payload235Parser.java case 195 → SOCIAL_ADD_FRIEND;
//         FriendStruct payload = [zero-padded string player_name].
func AddFriend(ctx context.Context, conn *session.Conn, name string) error {
	if name == "" {
		return fmt.Errorf("action: AddFriend name empty")
	}
	buf := v235.NewBuffer(2 + len(name))
	buf.WriteZeroPaddedString(name)
	return conn.Send(outAddFriend, buf.Bytes())
}

// RemoveFriend removes a player from the bot's friend list.
func RemoveFriend(ctx context.Context, conn *session.Conn, name string) error {
	if name == "" {
		return fmt.Errorf("action: RemoveFriend name empty")
	}
	buf := v235.NewBuffer(2 + len(name))
	buf.WriteZeroPaddedString(name)
	return conn.Send(outRemoveFriend, buf.Bytes())
}
