package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// outCommand is the admin-command opcode per Payload235Parser.java:196-198.
const outCommand byte = 38

// Command sends an admin command (the kind that starts with "::" in
// the RSC client — e.g., "heal", "teleport 120 504", "spawnnpc 184").
// The leading "::" must NOT be included.
//
// Payload format per Payload235Parser.java:345-348 (COMMAND case):
//
//	[zero-padded string] command_text
//
// Requires the account to have a staffmodlevel sufficient for the
// specific command (e.g., teleport requires admin).
func Command(ctx context.Context, conn *session.Conn, command string) error {
	if command == "" {
		return fmt.Errorf("action: empty command")
	}
	// Strip leading "::" if user accidentally included it.
	command = strings.TrimPrefix(command, "::")

	buf := v235.NewBuffer(len(command) + 1)
	buf.WriteZeroPaddedString(command)
	return conn.Send(outCommand, buf.Bytes())
}
