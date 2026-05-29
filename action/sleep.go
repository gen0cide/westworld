package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// outSleepWord is SLEEPWORD_ENTERED (Payload235Parser.java:214-215,
// case 45). Sent in response to SEND_SLEEPSCREEN to type the sleep
// word. The server parses it as [unsigned byte sleepDelay] followed by
// [zero-padded string sleepWord] (Payload235Parser.java:697-702).
//
// (The no-target "use/op" that starts sleeping — ITEM_COMMAND, opcode
// 90 — already lives in action/interaction.go as ItemCommand; the
// sleeping bag routes through it. See dslUse's 1-arg path.)
const outSleepWord byte = 45

// SendSleepWord fires SLEEPWORD_ENTERED (opcode 45) to answer the sleep
// screen captcha. On this OpenRSC server the correct word is hardcoded
// to "asleep" (CaptchaGenerator.generateRSCLCaptcha sets
// player.setSleepword("asleep") when prerendered sleepwords aren't
// loaded — CaptchaGenerator.java:79-80), so the runtime auto-answers
// with "asleep" the moment SEND_SLEEPSCREEN arrives — no OCR needed.
//
// Payload per Payload235Parser.java SLEEPWORD_ENTERED case
// (lines 697-702):
//
//	[unsigned byte]        sleepDelay  (we send 0 — only used by the 233
//	                                    "compatible" client; SleepHandler
//	                                    treats it as a TODO no-op)
//	[zero-padded string]   sleepWord   (the typed word, e.g. "asleep")
func SendSleepWord(ctx context.Context, conn *session.Conn, word string) error {
	if word == "" {
		return fmt.Errorf("action: empty sleep word")
	}
	buf := v235.NewBuffer(len(word) + 3)
	buf.WriteByte(0) // sleepDelay — unused by the v235 server path
	buf.WriteZeroPaddedString(word)
	return conn.Send(outSleepWord, buf.Bytes())
}
