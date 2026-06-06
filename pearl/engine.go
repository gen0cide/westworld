package pearl

import (
	"slices"
	"sort"

	"github.com/gen0cide/westworld/brain"
)

// Default tuning for New.
const (
	defaultFloor        = 0.6  // min rule confidence for a TryDecide hit
	defaultDecideMargin = 0.25 // min weight gap between top and runner-up
)

// Engine is a host's compiled policy engine: one rule table evaluated by two
// tiny scanners. Construct it with New from a compiled Table. It holds no
// model weights and does no I/O; every method is pure over the passed Facts.
type Engine struct {
	rules        []Rule  // sorted by Salience desc, ID asc (deterministic)
	floor        float64 // confidence floor for a TryDecide hit
	decideMargin float64 // min top-vs-runnerup weight gap for a TryDecide hit
	version      string
}

// New builds an Engine from a compiled table. floor is the minimum rule
// confidence for TryDecide to answer locally; pass <= 0 for the default.
func New(t Table, floor float64) *Engine {
	rules := make([]Rule, len(t.Rules))
	copy(rules, t.Rules)
	// Stable, deterministic order: higher salience first, ties by ID.
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Salience != rules[j].Salience {
			return rules[i].Salience > rules[j].Salience
		}
		return rules[i].ID < rules[j].ID
	})
	e := &Engine{rules: rules, floor: floor, decideMargin: defaultDecideMargin, version: t.Version}
	if e.floor <= 0 {
		e.floor = defaultFloor
	}
	return e
}

// Version returns the compiled table's version stamp.
func (e *Engine) Version() string { return e.version }

// TryDecide is the decision fast path. Given the assembled Facts and the
// option set from a decide() call, it returns:
//
//   - (decision, _, true)  on a confident local hit — NO LLM call should follow;
//   - (nil, biased, false) on a miss — the caller escalates to the LLM, passing
//     `biased`, the option list reordered by persona bias so even a miss is
//     persona-shaped.
//
// Evaluation: a firing Substitute whose Action is among the options forces that
// choice; firing Vetoes remove their named options; remaining options are
// scored by accumulated Bias weights, and the top option wins iff its margin
// over the runner-up clears decideMargin AND its backing confidence clears the
// floor.
func (e *Engine) TryDecide(f *Facts, options []string) (*brain.Decision, []string, bool) {
	if f.Event.Action == "" {
		f.Event.Action = "decide"
	}
	weights := make(map[string]float64, len(options))
	conf := make(map[string]float64, len(options))
	vetoed := make(map[string]bool)
	inOpts := func(s string) bool { return slices.Contains(options, s) }

	var sub *Rule
	for i := range e.rules {
		r := &e.rules[i]
		if r.When == nil || !r.When(f) {
			continue
		}
		switch r.Then.Kind {
		case EffectSubstitute:
			if sub == nil && inOpts(r.Then.Action) {
				sub = r
			}
		case EffectVeto:
			for _, o := range r.Then.Options {
				vetoed[o] = true
			}
		case EffectBias:
			for opt, w := range r.Then.Bias {
				weights[opt] += w
				if r.Confidence > conf[opt] {
					conf[opt] = r.Confidence
				}
			}
		}
	}

	// A substitute is a forced choice (the persona's "always do X here").
	if sub != nil {
		return &brain.Decision{Choice: sub.Then.Action, Confidence: sub.Confidence, Reasoning: sub.Then.Reason}, nil, true
	}

	// Candidates = options minus any vetoed by a firing rule.
	candidates := make([]string, 0, len(options))
	for _, o := range options {
		if !vetoed[o] {
			candidates = append(candidates, o)
		}
	}
	switch len(candidates) {
	case 0:
		// Everything vetoed — can't decide locally; escalate with the original set.
		return nil, options, false
	case 1:
		c := candidates[0]
		cf := conf[c]
		if cf < e.floor {
			cf = e.floor // a forced-by-elimination choice is a confident local answer
		}
		return &brain.Decision{Choice: c, Confidence: cf, Reasoning: "only remaining option after veto"}, nil, true
	}

	top, second := rankTwo(candidates, weights)
	if weights[top]-weights[second] >= e.decideMargin && conf[top] >= e.floor {
		return &brain.Decision{Choice: top, Confidence: conf[top], Reasoning: "policy bias"}, nil, true
	}
	// Miss: hand back the candidates reordered by weight (persona-shaped).
	return nil, reorderByWeight(candidates, weights), false
}

// Gate is the action-seam veto. Called synchronously before an action executes
// (author-written or brain-authored). It returns the first decisive verdict in
// salience order: a veto denies, a substitute replaces. Default is allow.
func (e *Engine) Gate(f *Facts, action string, args []string) Verdict {
	f.Event.Action = action
	for i := range e.rules {
		r := &e.rules[i]
		if r.Then.Kind != EffectVeto && r.Then.Kind != EffectSubstitute {
			continue
		}
		if r.When == nil || !r.When(f) {
			continue
		}
		if r.Then.Kind == EffectSubstitute {
			eff := r.Then
			return Verdict{Allow: false, Substitute: &eff, Reason: r.Then.Reason, RuleID: r.ID}
		}
		return Verdict{Allow: false, Reason: r.Then.Reason, RuleID: r.ID}
	}
	return Verdict{Allow: true}
}

// Injections returns the firing EffectInject intents (quirk/reverie tics) for
// the current Facts, in salience order. The caller is responsible for
// refractory-gating and picking at most one to emit.
func (e *Engine) Injections(f *Facts) []Effect {
	var out []Effect
	for i := range e.rules {
		r := &e.rules[i]
		if r.Then.Kind != EffectInject {
			continue
		}
		if r.When == nil || !r.When(f) {
			continue
		}
		out = append(out, r.Then)
	}
	return out
}

// rankTwo returns the highest- and second-highest-weighted candidates. Ties
// (and absent weights, which default to 0) break by slice order for
// determinism.
func rankTwo(candidates []string, weights map[string]float64) (top, second string) {
	top = candidates[0]
	second = candidates[1]
	if weights[second] > weights[top] {
		top, second = second, top
	}
	for _, c := range candidates[2:] {
		w := weights[c]
		switch {
		case w > weights[top]:
			second = top
			top = c
		case w > weights[second]:
			second = c
		}
	}
	return top, second
}

// reorderByWeight returns a copy of candidates sorted by descending weight,
// stable for equal weights.
func reorderByWeight(candidates []string, weights map[string]float64) []string {
	out := make([]string, len(candidates))
	copy(out, candidates)
	sort.SliceStable(out, func(i, j int) bool { return weights[out[i]] > weights[out[j]] })
	return out
}
