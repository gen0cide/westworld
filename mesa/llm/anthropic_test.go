package llm

import (
	"testing"
	"time"
)

// TestNewHasHTTPTimeout proves the default client carries the hard per-request
// HTTP deadline — the property that makes a wedged Anthropic upstream a timeout
// error instead of a stuck handler goroutine (the 34-minute-hang class).
func TestNewHasHTTPTimeout(t *testing.T) {
	c := New("key", "")
	if c.Timeout() != DefaultTimeout {
		t.Fatalf("New() timeout = %v, want %v", c.Timeout(), DefaultTimeout)
	}
	if c.Timeout() <= 0 {
		t.Fatal("client constructed with no HTTP timeout — unbounded requests are forbidden")
	}
}

// TestNewWithTimeout proves the timeout is operator-tunable but can never be
// disabled: zero/negative falls back to the default.
func TestNewWithTimeout(t *testing.T) {
	if got := NewWithTimeout("key", "", 5*time.Second).Timeout(); got != 5*time.Second {
		t.Fatalf("NewWithTimeout(5s) timeout = %v, want 5s", got)
	}
	if got := NewWithTimeout("key", "", 0).Timeout(); got != DefaultTimeout {
		t.Fatalf("NewWithTimeout(0) timeout = %v, want the %v default (never unbounded)", got, DefaultTimeout)
	}
	if got := NewWithTimeout("key", "", -1).Timeout(); got != DefaultTimeout {
		t.Fatalf("NewWithTimeout(-1) timeout = %v, want the %v default (never unbounded)", got, DefaultTimeout)
	}
}

// TestNewWithTimeoutModelFallback keeps the model fallback intact on the new
// constructor path.
func TestNewWithTimeoutModelFallback(t *testing.T) {
	if got := NewWithTimeout("key", "", time.Minute).Model(); got != DefaultModel {
		t.Fatalf("empty model = %q, want DefaultModel %q", got, DefaultModel)
	}
	if got := NewWithTimeout("key", "claude-haiku-4-5", time.Minute).Model(); got != "claude-haiku-4-5" {
		t.Fatalf("model = %q, want the configured id", got)
	}
}
