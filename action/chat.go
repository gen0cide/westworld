package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// outChatMessage is the client→server public chat opcode.
// Source: Payload235Parser.java:193-194 (case 216 → CHAT_MESSAGE).
const outChatMessage byte = 216

// MaxChatLength is the conventional max message length in RSC.
const MaxChatLength = 80

// Say sends a public chat message. The body is RSC-compressed
// (Huffman-like per OpenRSC's StringEncryption) before sending.
//
// Server reads via Packet.getEncryptedString (DataConversions.java:350)
// which expects: [smart_len chars] [N compressed bytes].
//
// Errors if message is empty or exceeds MaxChatLength.
func Say(ctx context.Context, conn *session.Conn, message string) error {
	if message == "" {
		return fmt.Errorf("action: empty chat message")
	}
	if len(message) > MaxChatLength {
		return fmt.Errorf("action: chat message too long (%d > %d)", len(message), MaxChatLength)
	}
	buf := v235.NewBuffer(len(message) + 4)
	buf.WriteSmart08_16(len(message))
	compressed := v235.EncipherRSCString(message)
	buf.WriteBytes(compressed)
	return conn.Send(outChatMessage, buf.Bytes())
}
