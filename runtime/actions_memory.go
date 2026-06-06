package runtime

import (
	"context"
	"encoding/json"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Location-blind storage verbs. These route through the host's tiered memory
// Manager (scratch → local → mesa), which decides where each key actually lives
// per its namespace policy. Routines never see local-vs-remote — they just
// remember/recollect/forget by key. Keys are namespaced by convention with a
// "namespace:id" form ("relationship:alex", "goal:current"); the namespace
// drives the storage policy.
//
// v1 stores string values (the value is coerced to its display string and
// JSON-encoded). Structured value round-tripping is a later enhancement.

// dslRemember writes a value under a key: remember("relationship:alex", "trusted").
func dslRemember(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("remember takes 2 args (key, value), got %d", len(args))
	}
	if h.Memory == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "remember: no memory manager wired"), nil
	}
	key := stringOf(args[0])
	if key == "" {
		return nil, errf("remember: key must be a non-empty string")
	}
	val, err := json.Marshal(stringOf(args[1]))
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, "remember: encode value: "+err.Error()), nil
	}
	if err := h.Memory.Put(ctx, key, val); err != nil {
		return interp.Fail(interp.SERVER_REJECTED, "remember: "+err.Error()), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslRecollect reads a value by key: recollect("relationship:alex"). Returns the
// stored value on a hit, or Null on a miss (a routine branches on `== null`).
func dslRecollect(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("recollect takes 1 arg (key), got %d", len(args))
	}
	if h.Memory == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "recollect: no memory manager wired"), nil
	}
	rec, ok, err := h.Memory.Get(ctx, stringOf(args[0]))
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, "recollect: "+err.Error()), nil
	}
	if !ok {
		return interp.Ok(interp.Null{}), nil
	}
	var s string
	if err := json.Unmarshal(rec.Value, &s); err != nil {
		// Stored as something other than a plain string — hand back the raw JSON.
		return interp.Ok(interp.String(string(rec.Value))), nil
	}
	return interp.Ok(interp.String(s)), nil
}

// dslForget deletes a key from every tier: forget("goal:current").
func dslForget(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("forget takes 1 arg (key), got %d", len(args))
	}
	if h.Memory == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "forget: no memory manager wired"), nil
	}
	if err := h.Memory.Delete(ctx, stringOf(args[0])); err != nil {
		return interp.Fail(interp.SERVER_REJECTED, "forget: "+err.Error()), nil
	}
	return interp.Ok(interp.Null{}), nil
}
