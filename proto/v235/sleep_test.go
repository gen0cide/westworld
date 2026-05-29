package v235

import (
	"bytes"
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDecodeSleepScreen covers the inbound SEND_SLEEPSCREEN (opcode 117).
// Payload is the raw captcha image bytes; the decoder surfaces only the
// size (we don't OCR — the server's word is hardcoded "asleep").
func TestDecodeSleepScreen(t *testing.T) {
	img := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	ev, err := DecodeInbound(Frame{Opcode: InSleepScreen, Payload: img}, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	ss, ok := ev.(event.SleepScreenAppeared)
	if !ok {
		t.Fatalf("got %T, want event.SleepScreenAppeared", ev)
	}
	if ss.ImageBytes != len(img) {
		t.Errorf("ImageBytes: got %d, want %d", ss.ImageBytes, len(img))
	}
}

// TestDecodeStopSleep covers the inbound SEND_STOPSLEEP (opcode 84): the
// wake packet, no payload.
func TestDecodeStopSleep(t *testing.T) {
	ev, err := DecodeInbound(Frame{Opcode: InStopSleep, Payload: nil}, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := ev.(event.SleepEnded); !ok {
		t.Fatalf("got %T, want event.SleepEnded", ev)
	}
}

// TestBuildSleepWord covers the outbound SLEEPWORD_ENTERED (opcode 45)
// payload encoding: [byte sleepDelay=0] then a zero-padded string
// (leading 0x00, content, trailing 0x00 — Packet.readZeroPaddedString).
func TestBuildSleepWord(t *testing.T) {
	buf := NewBuffer(len("asleep") + 3)
	buf.WriteByte(0)
	buf.WriteZeroPaddedString("asleep")
	got := buf.Bytes()

	// sleepDelay=0x00, then 0x00 'a' 's' 'l' 'e' 'e' 'p' 0x00.
	want := []byte{0x00, 0x00, 'a', 's', 'l', 'e', 'e', 'p', 0x00}
	if !bytes.Equal(got, want) {
		t.Errorf("sleepword payload: got %v, want %v", got, want)
	}
}
