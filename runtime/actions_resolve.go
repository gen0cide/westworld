package runtime

// actions_resolve.go — the control-plane recognition primitives
// resolve()/resolve_one() (api.md §5). These are fenced, non-GUI
// primitives (like note()): they reach the host's recognition faculty
// (learned-alias store → conservative fuzzy match → brain fallback),
// not the game. They send no packets and have no bang variant.
//
//   resolve(text, kind?)     -> List<Match>   (ranked, best-first; [] if none)
//   resolve_one(text, kind?) -> Match | Null  (best match, or Null if none)
//
// `kind`, when supplied, restricts the search to one catalog: one of
// "item", "npc", "loc", "spell", "prayer" (resolve.Kind* constants).
// Definitions and ids come from the facts catalog — never invented by
// the brain (see cognition/resolve).

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// dslResolve implements `resolve(text, kind?) -> List<Match>`. It
// routes through the host's recognition faculty and wraps each
// resolve.Match in a matchView. Returns an empty list when nothing
// resolves (never Null — the contract return type is a List).
func dslResolve(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("resolve takes 1 or 2 args (text, kind?), got %d", len(args))
	}
	text := stringOf(args[0])
	kind := ""
	if len(args) == 2 {
		kind = stringOf(args[1])
	}
	matches := h.resolver().ResolveCtx(ctx, text, kind)
	items := make([]interp.Value, 0, len(matches))
	for i := range matches {
		items = append(items, &matchView{m: matches[i]})
	}
	return &interp.List{Items: items}, nil
}

// dslResolveOne implements `resolve_one(text, kind?) -> Match | Null`.
// It is the common case sugar over resolve(): take the single best
// (first, highest-scored) candidate, or Null when nothing resolves.
func dslResolveOne(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("resolve_one takes 1 or 2 args (text, kind?), got %d", len(args))
	}
	text := stringOf(args[0])
	kind := ""
	if len(args) == 2 {
		kind = stringOf(args[1])
	}
	matches := h.resolver().ResolveCtx(ctx, text, kind)
	if len(matches) == 0 {
		return interp.Null{}, nil
	}
	return &matchView{m: matches[0]}, nil
}
