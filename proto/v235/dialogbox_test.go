package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// zeroQuoted builds an OpenRSC zero-quoted string payload (leading
// 0x00, UTF-8 body, trailing 0x00) — the shape SEND_BOX / SEND_BOX2
// write for dialog text.
func zeroQuoted(s string) []byte {
	out := make([]byte, 0, len(s)+2)
	out = append(out, 0)
	out = append(out, []byte(s)...)
	out = append(out, 0)
	return out
}

// SEND_BOX2 (opcode 89) carries the same single zero-quoted string as
// SEND_BOX (222). Before this fix, 89 fell through to UnknownPacket
// and its player-facing text was discarded. Both must now decode to
// NpcDialogText so the text reaches world.messages + the cradle log.
func TestDecodeBoxOpcodes(t *testing.T) {
	const text = "You can't reach that."
	for _, op := range []byte{InNpcDialogText, InNpcDialogBox} {
		ev, err := DecodeInbound(Frame{Opcode: op, Payload: zeroQuoted(text)}, nil)
		if err != nil {
			t.Fatalf("opcode %d: decode err: %v", op, err)
		}
		dt, ok := ev.(event.NpcDialogText)
		if !ok {
			t.Fatalf("opcode %d: got %T, want event.NpcDialogText", op, ev)
		}
		if dt.Text != text {
			t.Errorf("opcode %d: text = %q, want %q", op, dt.Text, text)
		}
	}
}
