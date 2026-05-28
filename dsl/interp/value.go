// Package interp is the AST-walking interpreter for .routine files.
// This file defines the runtime Value type and its concrete kinds.
//
// Design notes:
//
//   - Values are immutable views from the runtime's perspective:
//     mutations happen only through Actions (which go through the Host
//     bridge in step 6) — DSL code never mutates a value in place.
//   - Truthiness follows dsl.md: false, null, 0, 0.0, "", []
//     are falsey; everything else is truthy.
//   - Equality is structural / value-equal — `==` compares values, not
//     references. List and Map equality is element-wise.
//   - The Member/Index protocols (Getter/Indexer) let host-supplied
//     entities (self, world, inventory, etc.) participate in attribute
//     access without the interpreter knowing about every field.
package interp

import (
	"fmt"
	"strconv"
	"strings"
)

// Value is the runtime representation of any DSL value.
type Value interface {
	// Kind returns a short type-name suitable for error messages.
	Kind() string
	// Display returns the string form used by f-string
	// interpolation, note(), say(), etc.
	Display() string
}

// Truthy implements dsl.md's "Python-style many-falsey" rule.
func Truthy(v Value) bool {
	switch x := v.(type) {
	case nil:
		return false
	case Null:
		return false
	case Bool:
		return bool(x)
	case Int:
		return int64(x) != 0
	case Float:
		return float64(x) != 0
	case String:
		return string(x) != ""
	case *List:
		return len(x.Items) != 0
	case *Map:
		return len(x.Items) != 0
	}
	// Custom entities (Getter / Indexer) are truthy if non-nil.
	return v != nil
}

// Equal implements `==` / `!=`. Cross-numeric-kind comparisons
// promote Int → Float, like Python.
func Equal(a, b Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if _, ok := a.(Null); ok {
		_, isNull := b.(Null)
		return isNull
	}
	if _, ok := b.(Null); ok {
		return false
	}
	// Numeric cross-kind: promote Int → Float.
	if ai, ok := a.(Int); ok {
		if bf, ok := b.(Float); ok {
			return float64(ai) == float64(bf)
		}
	}
	if af, ok := a.(Float); ok {
		if bi, ok := b.(Int); ok {
			return float64(af) == float64(bi)
		}
	}
	switch x := a.(type) {
	case Bool:
		y, ok := b.(Bool)
		return ok && bool(x) == bool(y)
	case Int:
		y, ok := b.(Int)
		return ok && int64(x) == int64(y)
	case Float:
		y, ok := b.(Float)
		return ok && float64(x) == float64(y)
	case String:
		y, ok := b.(String)
		return ok && string(x) == string(y)
	case *List:
		y, ok := b.(*List)
		if !ok || len(x.Items) != len(y.Items) {
			return false
		}
		for i := range x.Items {
			if !Equal(x.Items[i], y.Items[i]) {
				return false
			}
		}
		return true
	case *Map:
		y, ok := b.(*Map)
		if !ok || len(x.Items) != len(y.Items) {
			return false
		}
		for k, av := range x.Items {
			bv, has := y.Items[k]
			if !has || !Equal(av, bv) {
				return false
			}
		}
		return true
	}
	// Reference equality for everything else (entities).
	return a == b
}

// ----- concrete value kinds -----

// Null is the singleton `null` value.
type Null struct{}

func (Null) Kind() string    { return "null" }
func (Null) Display() string { return "null" }

// Bool wraps a Go bool.
type Bool bool

func (b Bool) Kind() string { return "bool" }
func (b Bool) Display() string {
	if b {
		return "true"
	}
	return "false"
}

// Int wraps a Go int64.
type Int int64

func (i Int) Kind() string    { return "int" }
func (i Int) Display() string { return strconv.FormatInt(int64(i), 10) }

// Float wraps a Go float64.
type Float float64

func (f Float) Kind() string { return "float" }
func (f Float) Display() string {
	return strconv.FormatFloat(float64(f), 'g', -1, 64)
}

// String wraps a Go string.
type String string

func (s String) Kind() string    { return "string" }
func (s String) Display() string { return string(s) }

// Range is `low..high` — produced by RangeLit. Mostly fed to
// wait(); also iterable by for-in.
type Range struct{ Low, High Value }

func (r *Range) Kind() string { return "range" }
func (r *Range) Display() string {
	return fmt.Sprintf("%s..%s", r.Low.Display(), r.High.Display())
}

// List is a sequence. Mutation in v1 happens only via stdlib /
// builtins — DSL syntax can't push/pop.
type List struct{ Items []Value }

func (l *List) Kind() string { return "list" }
func (l *List) Display() string {
	parts := make([]string, len(l.Items))
	for i, it := range l.Items {
		parts[i] = it.Display()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Map is an unordered key->value collection. Display is a
// best-effort summary.
type Map struct{ Items map[string]Value }

func (m *Map) Kind() string { return "map" }
func (m *Map) Display() string {
	parts := make([]string, 0, len(m.Items))
	for k, v := range m.Items {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v.Display()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// ----- entity protocols -----

// Getter is implemented by host-supplied entities (self, world,
// inventory, combat, entities returned from world.npcs etc.) so the
// interpreter can resolve `recv.field`. Returns (value, true) when
// the field exists; (nil, false) when it does not.
type Getter interface {
	Get(field string) (Value, bool)
}

// Indexer is implemented by entities that support `recv[index]`.
type Indexer interface {
	Index(idx Value) (Value, bool)
}

// Callable is implemented by callable values — bound methods,
// builtin functions, and stdlib functions. Args are pre-evaluated by
// the interpreter; named args are collected and passed via NamedArgs.
type Callable interface {
	Call(args []Value, namedArgs map[string]Value) (Value, error)
}

// Length is implemented by collections that have a .length property.
type Lengther interface {
	Length() int
}

// ----- helpers -----

// AsInt extracts a Go int64 from an Int. Returns (0,false) for
// any other type, including Float (caller must promote explicitly).
func AsInt(v Value) (int64, bool) {
	if i, ok := v.(Int); ok {
		return int64(i), true
	}
	return 0, false
}

// AsFloat returns the value as a float64 if numeric. Promotes Int.
func AsFloat(v Value) (float64, bool) {
	switch x := v.(type) {
	case Int:
		return float64(x), true
	case Float:
		return float64(x), true
	}
	return 0, false
}

// AsString returns the value as a Go string if it's a String.
func AsString(v Value) (string, bool) {
	if s, ok := v.(String); ok {
		return string(s), true
	}
	return "", false
}
