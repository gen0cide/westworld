package v235

import (
	"bytes"
	"testing"
)

func TestBuildWalkToPoint(t *testing.T) {
	got := BuildWalkToPoint(120, 504)
	// 120 = 0x0078, big-endian = [0x00, 0x78]
	// 504 = 0x01F8, big-endian = [0x01, 0xF8]
	want := []byte{0x00, 0x78, 0x01, 0xF8}
	if !bytes.Equal(got, want) {
		t.Errorf("walk payload: got %v, want %v", got, want)
	}
}

func TestBuildHeartbeatEmpty(t *testing.T) {
	if got := BuildHeartbeat(); len(got) != 0 {
		t.Errorf("heartbeat should have empty payload, got %v", got)
	}
}

func TestBuildLogoutEmpty(t *testing.T) {
	if got := BuildLogout(); len(got) != 0 {
		t.Errorf("logout should have empty payload, got %v", got)
	}
}
