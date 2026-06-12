package mesad

import "testing"

func TestPeerMuted(t *testing.T) {
	s := &Server{}
	s.SetChatPolicy("alex", false)
	for _, tc := range []struct {
		speaker, role string
		want          bool
	}{
		{"Drone63", "player", true},   // peer → muted
		{"alex", "player", false},     // operator → heard
		{"ALEX", "player", false},     // case-insensitive
		{" alex ", "player", false},   // whitespace tolerated
		{"Shopkeeper", "npc", false},  // NPCs always processed
		{"server", "server", false},   // server always processed
		{"Drone63", "", true},         // empty role defaults to player semantics
	} {
		if got := s.peerMuted(tc.speaker, tc.role); got != tc.want {
			t.Errorf("peerMuted(%q,%q)=%v want %v", tc.speaker, tc.role, got, tc.want)
		}
	}
	s.SetChatPolicy("alex", true)
	if s.peerMuted("Drone63", "player") {
		t.Error("peerChat=true must not mute anyone")
	}
}
