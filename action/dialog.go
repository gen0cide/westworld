package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// ChooseDialogOption picks one of the options the server presented in
// a SendNpcDialogOptions packet (opcode 245). The option index is
// 0-based and must match a slot in the inbound NpcDialog.Options
// slice we just received.
//
// Source: Payload235Parser.java case 116 → QUESTION_DIALOG_ANSWER;
//         MenuOptionStruct payload = [byte option].
const outDialogChoice byte = 116

func ChooseDialogOption(ctx context.Context, conn *session.Conn, optionIndex int) error {
	if optionIndex < 0 || optionIndex > 0xFF {
		return fmt.Errorf("action: dialog option %d out of byte range", optionIndex)
	}
	buf := v235.NewBuffer(1)
	buf.WriteByte(byte(optionIndex))
	return conn.Send(outDialogChoice, buf.Bytes())
}
