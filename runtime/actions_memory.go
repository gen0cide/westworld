package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gen0cide/westworld/cognition"
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

// ---------- Memory stdlib (cognition.Client) ----------

// dslRecall routes `recall(query, top?)` to the host's knowledge
// surfaces and returns a List<String> on .val.
//
// Priority order:
//
//  1. If h.Corpus is wired (Phase 2.6+), query it directly. Returns
//     formatted chunk strings: "[source § page § section] text".
//     This is the path real wiki/AutoRune content flows through.
//  2. Otherwise, fall back to h.Retriever (Phase 2.5 stub behavior)
//     and return its Bundle.Reflections list.
//
// Routines do not see which path was used — both return List<String>.
// The Corpus path is preferred because chunks carry provenance.
func dslRecall(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("recall takes 1 or 2 args (query, top?), got %d", len(args))
	}
	query := stringOf(args[0])
	maxItems := 3
	if len(args) == 2 {
		if i, ok := args[1].(interp.Int); ok {
			maxItems = int(i)
		}
	}
	if h.Corpus != nil {
		chunks, err := h.Corpus.Recall(ctx, query, maxItems)
		if err != nil {
			return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("recall: %v", err)), nil
		}
		items := make([]interp.Value, 0, len(chunks))
		for _, c := range chunks {
			items = append(items, interp.String(formatChunkForRecall(c)))
		}
		return interp.Ok(&interp.List{Items: items}), nil
	}
	if h.Retriever == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "recall: no retriever or corpus wired"), nil
	}
	hostName := ""
	if h.opts.Username != "" {
		hostName = h.opts.Username
	}
	bundle, err := h.Retriever.Retrieve(ctx, cognition.Retrieval{
		Goal:     query,
		HostName: hostName,
		MaxItems: maxItems,
	})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("recall: %v", err)), nil
	}
	items := make([]interp.Value, 0, len(bundle.Reflections))
	for _, r := range bundle.Reflections {
		items = append(items, interp.String(r))
	}
	return interp.Ok(&interp.List{Items: items}), nil
}

// dslRelationWith routes `relation_with(name)` → retriever with
// goal = "relation with NAME". Returns a string describing the
// relationship from bundle.Persona["relation:NAME"] if present,
// else from the first reflection.
func dslRelationWith(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("relation_with takes 1 argument (name), got %d", len(args))
	}
	name := stringOf(args[0])
	// Prefer the host's own trust ledger (System-1, learned from interactions):
	// if we've met this party, return our felt trust grade.
	if h.ledger != nil && h.ledger.Known(name) {
		return interp.Ok(interp.String(h.ledger.Rel(name).Grade.String())), nil
	}
	if h.Retriever == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "relation_with: no retriever wired"), nil
	}
	bundle, err := h.Retriever.Retrieve(ctx, cognition.Retrieval{
		Goal:     "relation with " + name,
		HostName: h.opts.Username,
		MaxItems: 1,
	})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("relation_with: %v", err)), nil
	}
	rel := ""
	if v, ok := bundle.Persona["relation:"+name]; ok {
		rel = v
	} else if len(bundle.Reflections) > 0 {
		rel = bundle.Reflections[0]
	} else {
		rel = "unknown"
	}
	return interp.Ok(interp.String(rel)), nil
}
