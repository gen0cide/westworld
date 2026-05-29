package world

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// The messages ring is the complete player-facing-text log: opcode-131
// server messages AND dialog/NPC-speech text all land in it, each
// tagged with its kind. SetServerMessage defaults to "server";
// SetMessage carries the explicit kind.
func TestMessageRingCarriesKind(t *testing.T) {
	r := NewRecentEvents()
	r.SetServerMessage("the door is locked")
	r.SetMessage("dialog", "Greetings, adventurer!")

	msgs := r.ServerMessages()
	if len(msgs) != 2 {
		t.Fatalf("ring length: got %d, want 2", len(msgs))
	}
	if msgs[0].Kind != "server" || msgs[0].Message != "the door is locked" {
		t.Errorf("entry 0: got kind=%q msg=%q, want server/door-locked", msgs[0].Kind, msgs[0].Message)
	}
	if msgs[1].Kind != "dialog" || msgs[1].Message != "Greetings, adventurer!" {
		t.Errorf("entry 1: got kind=%q msg=%q, want dialog/greetings", msgs[1].Kind, msgs[1].Message)
	}
	// last_server_message still mirrors the newest ring entry.
	if last := r.ServerMessage(); last == nil || last.Message != "Greetings, adventurer!" {
		t.Errorf("ServerMessage(): got %+v, want newest entry", last)
	}
}

// world.Apply(NpcDialogText) must both set the dedicated dialog slot
// and append to the messages ring (kind=dialog), so quest/dialog prose
// is observable in the unified message log.
func TestApplyDialogTextFeedsRing(t *testing.T) {
	w := NewWorld()
	w.Apply(event.NpcDialogText{Text: "Speak to the cook in Lumbridge castle."})

	if dt := w.Recent.DialogText(); dt == nil || dt.Text != "Speak to the cook in Lumbridge castle." {
		t.Fatalf("dialog-text slot not set: %+v", dt)
	}
	msgs := w.Recent.ServerMessages()
	if len(msgs) != 1 || msgs[0].Kind != "dialog" {
		t.Fatalf("ring: got %+v, want one dialog entry", msgs)
	}
	if msgs[0].Message != "Speak to the cook in Lumbridge castle." {
		t.Errorf("ring text: got %q", msgs[0].Message)
	}
}
