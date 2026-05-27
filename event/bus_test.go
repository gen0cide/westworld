package event

import (
	"testing"
	"time"
)

func TestBusBasicPublishSubscribe(t *testing.T) {
	b := NewBus()
	defer b.Close()
	sub := b.Subscribe("chat_received", 4)
	b.Publish(NewChatReceived(MessageChat, "alice", "hi", ""))
	select {
	case ev := <-sub:
		c, ok := ev.(ChatReceived)
		if !ok {
			t.Fatalf("unexpected type %T", ev)
		}
		if c.Speaker != "alice" || c.Message != "hi" {
			t.Errorf("event content: %+v", c)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive event within 1s")
	}
}

func TestBusWildcardSubscribe(t *testing.T) {
	b := NewBus()
	defer b.Close()
	all := b.Subscribe("*", 4)
	b.Publish(NewChatReceived(MessageChat, "bob", "hello", ""))
	select {
	case ev := <-all:
		if ev.Kind() != "chat_received" {
			t.Errorf("wildcard got kind=%q", ev.Kind())
		}
	case <-time.After(time.Second):
		t.Fatal("wildcard subscriber did not receive event")
	}
}

func TestBusDropsOnSlowSubscriber(t *testing.T) {
	b := NewBus()
	defer b.Close()
	sub := b.Subscribe("chat_received", 1) // tiny buffer
	for i := 0; i < 100; i++ {
		b.Publish(NewChatReceived(MessageChat, "spammer", "msg", ""))
	}
	// Drain whatever's in the buffer; we shouldn't block.
	count := 0
loop:
	for {
		select {
		case <-sub:
			count++
		case <-time.After(50 * time.Millisecond):
			break loop
		}
	}
	if count == 0 {
		t.Errorf("expected at least 1 delivery, got 0")
	}
	if count >= 100 {
		t.Errorf("expected drops (buffer=1, publish=100), but got %d deliveries", count)
	}
}
