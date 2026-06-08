package action

import (
	"context"
	"fmt"
	"strings"

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
	message = sanitizeChat(message)
	if message == "" {
		return fmt.Errorf("action: empty chat message")
	}
	// The RSC Huffman codec emits one code per CHARACTER (rune), and the
	// server decodes exactly the smart-length char count. Send the RUNE
	// count — NOT len(message) (Go bytes). A multi-byte rune (a curly quote
	// or em-dash from the LLM) makes byte-count > char-count, so the server
	// would try to decode more characters than we encoded and its Huffman
	// decoder walks off the end (StringEncryption.decryptString throws
	// ArrayIndexOutOfBounds). EncipherRSCString encodes exactly len(runes)
	// chars (one byte per rune, non-ASCII folded to '?'), so len(chars) is
	// the correct, matching count.
	chars := []rune(message)
	if len(chars) > MaxChatLength {
		return fmt.Errorf("action: chat message too long (%d > %d chars)", len(chars), MaxChatLength)
	}
	buf := v235.NewBuffer(len(message) + 4)
	buf.WriteSmart08_16(len(chars))
	compressed := v235.EncipherRSCString(message)
	buf.WriteBytes(compressed)
	return conn.Send(outChatMessage, buf.Bytes())
}

// sanitizeChat folds the non-ASCII characters LLMs habitually emit (curly
// quotes, en/em dashes, ellipsis, non-breaking space) to their RSC-charset
// ASCII equivalents, so chat reads cleanly rather than as '?'. Any remaining
// non-ASCII rune is left for EncipherRSCString to fold to '?' — still one byte
// per rune, so the smart-length char count stays correct either way.
func sanitizeChat(s string) string { return chatSanitizer.Replace(s) }

var chatSanitizer = strings.NewReplacer(
	"‘", "'", "’", "'", // ‘ ’ curly single quotes
	"“", "\"", "”", "\"", // “ ” curly double quotes
	"–", "-", "—", "-", // – — en/em dash
	"…", "...", // … ellipsis
	" ", " ", // non-breaking space
)
