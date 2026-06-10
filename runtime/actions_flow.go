package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/pearl"
)

// Flow / timing / introspection primitives (wait / wait_until / note /
// look_around) and the LLM-stdlib cognition bridge (contemplate_reality /
// evaluate / decide, routed through Host.Strategist). Registered in the
// central actionHandlers table in dsl_actions.go.

// ---------- primitives (no Result wrap; can't fail in the typed sense) ----------

// dslWait sleeps for the given duration (seconds). Returns Null{}
// directly on success — wait is a primitive, not an action; no bang
// variant. Cancellation flows through ctx; a canceled wait returns
// the Go ctx error which becomes a RuntimeError.
func dslWait(ctx context.Context, _ *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("wait takes 1 argument (seconds), got %d", len(args))
	}
	secs, ok := interp.AsFloat(args[0])
	if !ok {
		return nil, errf("wait: expected number, got %s", args[0].Kind())
	}
	if secs <= 0 {
		return interp.Null{}, nil
	}
	t := time.NewTimer(time.Duration(secs * float64(time.Second)))
	defer t.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.C:
		return interp.Null{}, nil
	}
}

// dslWaitUntil blocks until the predicate lambda evaluates truthy
// (or the optional timeout fires). The predicate is a single-arg
// lambda whose argument is ignored — RSC routines write
// `wait_until(_ => self.hp > 1)` and `wait_until(_ => world.bank.is_open, 10)`.
//
// Why a lambda instead of a bare expression: the DSL evaluates
// expressions eagerly at the call site. A bare predicate like
// `wait_until(self.hp > 1)` would compute true/false once and pass
// a Bool — the predicate's actual re-evaluation logic lives here in
// Go. Wrapping in a lambda is the existing "delay this expression"
// convention (cf. filter/map/find).
//
// Poll interval is fixed at 200ms (matches wait_for_dialog).
//
// Returns Bool(true) on satisfied, Bool(false) on timeout. Errors
// from predicate evaluation propagate as RuntimeError.
func dslWaitUntil(ctx context.Context, _ *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("wait_until takes 1 or 2 args (predicate_lambda, timeout?), got %d", len(args))
	}
	pred, ok := args[0].(interp.Callable)
	if !ok {
		return nil, errf("wait_until: first arg must be a lambda (e.g. `_ => self.hp > 1`), got %s", args[0].Kind())
	}
	var deadline time.Time
	if len(args) == 2 {
		secs, ok := interp.AsFloat(args[1])
		if !ok {
			return nil, errf("wait_until: timeout must be a number of seconds, got %s", args[1].Kind())
		}
		if secs > 0 {
			deadline = time.Now().Add(time.Duration(secs * float64(time.Second)))
		}
	}
	const poll = 200 * time.Millisecond
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Lambdas are 1-arg; we pass Null as the ignored param.
		v, err := pred.Call([]interp.Value{interp.Null{}}, nil)
		if err != nil {
			return nil, errf("wait_until predicate: %v", err)
		}
		if interp.Truthy(v) {
			return interp.Bool(true), nil
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return interp.Bool(false), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(poll):
		}
	}
}

// dslNote writes to the host's logger as an info entry. The full
// journal persistence is Phase 3. Primitive: returns Null directly,
// no bang variant.
func dslNote(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("note takes 1 argument (text), got %d", len(args))
	}
	text := args[0].Display()
	h.log.Info("routine note", "text", text)
	// Publish the narration onto the bus so the cradle UI (and any other
	// subscriber) can stream it as a live in-character feed.
	h.bus.Publish(event.RoutineNote{Text: text})
	return interp.Null{}, nil
}

// dslLookAround returns a brain-ready text summary of the scene around
// the host (self vitals, nearby NPCs/players with combat level + threat +
// gear, ground items, notable scenery) — one call a host hands to the LLM
// instead of stitching a dozen accessors. Optional radius (default 10).
func dslLookAround(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	radius := 10
	if len(args) >= 1 {
		if r, ok := interp.AsInt(args[0]); ok {
			radius = int(r)
		}
	}
	return interp.String(h.DescribeSurroundings(radius)), nil
}

// ---------- LLM stdlib (brain.Strategist) ----------

// dslContemplateReality dispatches `contemplate_reality(question)`
// through Host.Strategist. Returns the brain's Choice as a String
// wrapped in CallResult.val on success; brain errors become
// CallResult.err with code SERVER_REJECTED (the strategist is
// conceptually a remote service even for the stub).
func dslContemplateReality(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	question := ""
	if len(args) > 0 {
		question = stringOf(args[0])
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "contemplate_reality: no strategist wired"), nil
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("contemplate_reality: %v", err)), nil
	}
	return interp.Ok(interp.String(decision.Choice)), nil
}

// dslEvaluate routes `evaluate(situation)` → strategist with
// Options=[] (free-form). Returns Confidence as Float on .val.
// The 0-1 numeric assessment in the spec maps to the strategist's
// confidence score.
func dslEvaluate(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("evaluate takes 1 argument (situation), got %d", len(args))
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "evaluate: no strategist wired"), nil
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: stringOf(args[0])})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("evaluate: %v", err)), nil
	}
	return interp.Ok(interp.Float(decision.Confidence)), nil
}

// dslDecide routes `decide(options, context?)` → strategist with
// Options bound from the list arg. Returns Choice as String on .val.
// Forwards the optional context string into Situation.Question so
// the strategist sees it.
func dslDecide(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("decide takes 1 or 2 args (options, context?), got %d", len(args))
	}
	options := []string{}
	if list, ok := args[0].(*interp.List); ok {
		for _, item := range list.Items {
			options = append(options, stringOf(item))
		}
	} else {
		return nil, errf("decide: first arg must be a list of options")
	}
	question := ""
	if len(args) == 2 {
		question = stringOf(args[1])
	}
	// Pearl fast path: the host's compiled policy may answer locally (no LLM)
	// or, on a miss, hand back a persona-biased option ordering for the LLM.
	if h.Pearl != nil {
		f := h.pearlFacts(pearl.EventCtx{Action: "decide", Question: question})
		if d, biased, hit := h.Pearl.TryDecide(f, options); hit {
			return interp.Ok(interp.String(d.Choice)), nil
		} else if len(biased) > 0 {
			options = biased
		}
	}
	if h.Strategist == nil {
		return interp.Fail(interp.NOT_IMPLEMENTED, "decide: no strategist wired"), nil
	}
	// Decision cache (#16): a repeated pearl-MISS decision in materially-the-same
	// state reuses the prior verdict, skipping the (Haiku) Strategist call. Pearl
	// hits above are never cached — they are already free + authoritative.
	key := h.decisionCacheKey(question, options)
	if h.decisionCache != nil {
		if cached, ok := h.decisionCache.Get(key); ok {
			if choice, ok := cached.(string); ok {
				return interp.Ok(interp.String(choice)), nil
			}
		}
	}
	decision, err := h.Strategist.Decide(ctx, brain.Situation{Question: question, Options: options})
	if err != nil {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf("decide: %v", err)), nil
	}
	if h.decisionCache != nil {
		h.decisionCache.Set(key, decision.Choice, decisionCacheTTL)
	}
	return interp.Ok(interp.String(decision.Choice)), nil
}
