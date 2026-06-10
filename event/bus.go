package event

import (
	"sync"
)

// Bus is a typed pub/sub for game events. Subscribers receive events
// by Kind() string. A subscriber can also subscribe to "*" to receive
// every event.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
	closed      bool
}

// NewBus returns a new empty Bus.
func NewBus() *Bus {
	return &Bus{subscribers: make(map[string][]chan Event)}
}

// Subscribe registers a channel for the given event kind. The
// returned channel has the given buffer size; callers should drain it
// promptly. Use "*" to receive every event.
//
// Unsubscribe by closing the channel returned by Subscribe — actually,
// just stop reading and let the channel be GC'd; the next slow-send
// from the bus will drop messages destined for it.
//
// For Phase 1 we use a simple "drop if subscriber is slow" semantic to
// avoid head-of-line blocking. Phase 2 may add stricter guarantees.
func (b *Bus) Subscribe(kind string, buffer int) <-chan Event {
	ch := make(chan Event, buffer)
	b.mu.Lock()
	if b.closed {
		// Subscribing to a closed bus (e.g. a WebSocket landing in the
		// restart window after bus.Close) must not panic on the nil map;
		// return a pre-closed channel — callers already handle `!ok`.
		b.mu.Unlock()
		close(ch)
		return ch
	}
	b.subscribers[kind] = append(b.subscribers[kind], ch)
	b.mu.Unlock()
	return ch
}

// SubscriberCounts reports the live subscriber-channel count per event kind —
// the leak gauge for dead fan-out registrations: a healthy host's "*" count
// plateaus at its fixed session set; growth ~1/turn means a subscriber is not
// unsubscribing.
func (b *Bus) SubscriberCounts() map[string]int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(map[string]int, len(b.subscribers))
	for k, chs := range b.subscribers {
		out[k] = len(chs)
	}
	return out
}

// Unsubscribe removes a channel previously returned by Subscribe for the same
// kind, so a finished subscriber (e.g. a closed WebSocket) stops receiving events
// and — critically — stops adding per-event work + retained memory to Publish.
// Without this, a long-lived host watched through repeatedly-reopened WebSockets
// accumulates dead subscribers unboundedly. The channel is NOT closed (the owning
// goroutine may have already returned); it is simply dropped from the fan-out.
// Idempotent; a no-op if the channel isn't registered or the bus is closed.
func (b *Bus) Unsubscribe(kind string, ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	subs := b.subscribers[kind]
	for i, c := range subs {
		if (<-chan Event)(c) == ch {
			b.subscribers[kind] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

// Publish delivers the given event to subscribers. Async, non-blocking:
// if a subscriber's channel is full, the event is dropped for that
// subscriber.
//
// Use sync.Cond if strict delivery is required (not yet needed for
// Phase 1).
func (b *Bus) Publish(ev Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	for _, ch := range b.subscribers[ev.Kind()] {
		select {
		case ch <- ev:
		default:
			// subscriber too slow; drop. Phase 2 may add a metric.
		}
	}
	for _, ch := range b.subscribers["*"] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// Close stops the bus and closes all subscriber channels. After Close,
// Publish is a no-op.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for _, chs := range b.subscribers {
		for _, ch := range chs {
			close(ch)
		}
	}
	b.subscribers = nil
}
