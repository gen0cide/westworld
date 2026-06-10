package interp_test

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// callableNull is a minimal NullLike + Callable value — the shape the runtime
// uses for an empty-but-still-invocable selector read (e.g. the
// world.ground_items.nearest wrapper when every visible item is reach-gated
// out, whose called form .nearest(reachable=false) must stay invocable).
type callableNull struct{}

func (callableNull) Kind() string    { return "null" }
func (callableNull) Display() string { return "null" }
func (callableNull) NullLike()       {}
func (callableNull) Call([]interp.Value, map[string]interp.Value) (interp.Value, error) {
	return interp.Bool(true), nil
}

// TestNullLikeReadsAsNull pins the NullLike contract: a marked value is
// falsey, compares == null (in both operand positions, and against another
// null-like), compares != to any non-null value, and reports IsNullish — so
// routine guards like `x == null` / `if x { }` cannot tell it from Null.
func TestNullLikeReadsAsNull(t *testing.T) {
	cn := callableNull{}
	if interp.Truthy(cn) {
		t.Fatal("a NullLike value must be falsey")
	}
	if !interp.IsNullish(cn) || !interp.IsNullish(interp.Null{}) {
		t.Fatal("IsNullish must accept both Null and NullLike")
	}
	if interp.IsNullish(interp.Int(0)) {
		t.Fatal("IsNullish must reject non-null values")
	}
	if !interp.Equal(cn, interp.Null{}) || !interp.Equal(interp.Null{}, cn) {
		t.Fatal("a NullLike value must == null in either operand position")
	}
	if !interp.Equal(cn, callableNull{}) {
		t.Fatal("two null-like values must compare equal")
	}
	if interp.Equal(cn, interp.Int(1)) || interp.Equal(interp.String("x"), cn) {
		t.Fatal("a NullLike value must != any non-null value")
	}
	// And it stays Callable — the whole point of the marker.
	if _, ok := interp.Value(cn).(interp.Callable); !ok {
		t.Fatal("the test double must remain Callable")
	}
}
