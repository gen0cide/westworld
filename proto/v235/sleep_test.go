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

// TestDecodeSleepFatigue covers the inbound SEND_SLEEP_FATIGUE (opcode
// 244): the provisional fatigue value draining per tick while asleep.
// Same [short scaled 0..750] body as SEND_FATIGUE, but it must decode
// to the DISTINCT SleepFatigueUpdate event — applying it as a regular
// FatigueUpdate would let a provisional drain value (which only commits
// on a successful wake) overwrite real fatigue.
func TestDecodeSleepFatigue(t *testing.T) {
	for _, tc := range []struct {
		payload []byte
		want    int
	}{
		{[]byte{0x01, 0x5E}, 350}, // mid-drain
		{[]byte{0x00, 0x00}, 0},   // drain complete — the answer trigger
	} {
		ev, err := DecodeInbound(Frame{Opcode: InSleepFatigue, Payload: tc.payload}, nil)
		if err != nil {
			t.Fatalf("decode %v: %v", tc.payload, err)
		}
		sf, ok := ev.(event.SleepFatigueUpdate)
		if !ok {
			t.Fatalf("got %T, want event.SleepFatigueUpdate", ev)
		}
		if sf.Value != tc.want {
			t.Errorf("Value: got %d, want %d", sf.Value, tc.want)
		}
	}
}

// TestDecodeSleepwordIncorrect covers the inbound SEND_SLEEPWORD_INCORRECT
// (opcode 194, no payload): the server rejected our captcha answer. It was
// UNDECODED (landed as UnknownPacket) while sleep was a behavioral no-op
// fleet-wide; decoding it lets the runtime retry on the re-sent sleep
// screen instead of trapping asleep.
func TestDecodeSleepwordIncorrect(t *testing.T) {
	ev, err := DecodeInbound(Frame{Opcode: InSleepwordIncorrect, Payload: nil}, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := ev.(event.SleepwordIncorrect); !ok {
		t.Fatalf("got %T, want event.SleepwordIncorrect", ev)
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
