package v235

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestDecodeSceneryHandler covers the inbound SEND_SCENERY_HANDLER
// (opcode 48). Each record is [short id, byte offsetX, byte offsetY];
// id == 60000 (SceneryRemoveSentinel) marks a removal. Offsets are
// signed bytes relative to the player at delivery.
//
// The first record uses the exact payload observed live when a fire
// (def 97) was lit at the player's own tile: [0x00 0x61 0x00 0x00].
func TestDecodeSceneryHandler(t *testing.T) {
	payload := []byte{
		0x00, 0x61, 0x00, 0x00, // id=97 (0x0061), offset (0, 0) — a fire on our tile
		0xEA, 0x60, 0xFF, 0x02, // id=60000 removal, offset (-1, +2)
		0x00, 0x0B, 0x05, 0xFB, // id=11 (Range), offset (+5, -5)
	}
	ev, err := DecodeInbound(Frame{Opcode: InSceneryHandler, Payload: payload}, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	su, ok := ev.(event.SceneryUpdates)
	if !ok {
		t.Fatalf("got %T, want event.SceneryUpdates", ev)
	}
	if len(su.Updates) != 3 {
		t.Fatalf("len(Updates): got %d, want 3", len(su.Updates))
	}

	if d := su.Updates[0]; d.ID != 97 || d.OffsetX != 0 || d.OffsetY != 0 {
		t.Errorf("update[0] (fire): got %+v, want {ID:97 OffsetX:0 OffsetY:0}", d)
	}
	if d := su.Updates[1]; d.ID != 60000 || d.OffsetX != -1 || d.OffsetY != 2 {
		t.Errorf("update[1] (removal): got %+v, want {ID:60000 OffsetX:-1 OffsetY:2}", d)
	}
	if d := su.Updates[2]; d.ID != 11 || d.OffsetX != 5 || d.OffsetY != -5 {
		t.Errorf("update[2] (range): got %+v, want {ID:11 OffsetX:5 OffsetY:-5}", d)
	}
}

// TestDecodeSceneryHandlerEmpty: a zero-length payload yields an empty
// (non-nil) SceneryUpdates rather than an error — matches the boundary
// handler's tolerant decode.
func TestDecodeSceneryHandlerEmpty(t *testing.T) {
	ev, err := DecodeInbound(Frame{Opcode: InSceneryHandler, Payload: nil}, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	su, ok := ev.(event.SceneryUpdates)
	if !ok {
		t.Fatalf("got %T, want event.SceneryUpdates", ev)
	}
	if len(su.Updates) != 0 {
		t.Errorf("len(Updates): got %d, want 0", len(su.Updates))
	}
}
