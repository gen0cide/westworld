package mesad

import (
	"testing"
	"time"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

func TestExtractDedupKeyIsContentNotHearer(t *testing.T) {
	d := newExtractDedup()
	w1 := &mesapb.DialogWindow{Speaker: "Drone63", SpeakerRole: "player", Window: []string{"Drone63: rest before the north road"}}
	w2 := &mesapb.DialogWindow{Speaker: "Drone63", SpeakerRole: "player", Window: []string{"Drone63: rest before the north road"}}
	if d.key(w1) != d.key(w2) {
		t.Fatal("identical windows from different hearers must share a key")
	}
	w3 := &mesapb.DialogWindow{Speaker: "Drone63", SpeakerRole: "player", Window: []string{"Drone63: rest before the north road", "me: aye"}}
	if d.key(w1) == d.key(w3) {
		t.Fatal("a window containing the hearer's own reply must key differently")
	}
}

func TestExtractDedupTTLAndReplay(t *testing.T) {
	old := extractDedupTTL
	extractDedupTTL = 50 * time.Millisecond
	defer func() { extractDedupTTL = old }()

	d := newExtractDedup()
	w := &mesapb.DialogWindow{Speaker: "Bob", SpeakerRole: "player", Window: []string{"Bob: the anvil is south"}}
	set := &mesapb.ExtractedDialogSet{Intent: &mesapb.DialogIntent{Kind: "statement", Urgency: "low"}}
	k := d.key(w)
	now := time.Now()
	d.put(k, set, now)
	if got, ok := d.get(k, now.Add(10*time.Millisecond)); !ok || got != set {
		t.Fatal("fresh entry must replay the cached set")
	}
	if _, ok := d.get(k, now.Add(200*time.Millisecond)); ok {
		t.Fatal("expired entry must miss")
	}
}

func TestExtractDedupCapPressure(t *testing.T) {
	d := newExtractDedup()
	set := &mesapb.ExtractedDialogSet{}
	now := time.Now()
	for i := 0; i < extractDedupCap+50; i++ {
		w := &mesapb.DialogWindow{Speaker: "S", Window: []string{string(rune('a'+i%26)), string(rune(i))}}
		d.put(d.key(w), set, now)
	}
	d.mu.Lock()
	n := len(d.m)
	d.mu.Unlock()
	if n > extractDedupCap {
		t.Fatalf("cache exceeded cap: %d > %d", n, extractDedupCap)
	}
}
