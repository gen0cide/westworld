package interp

import (
	"context"
	"time"

	"github.com/gen0cide/westworld/dsl/ast"
)

// Watchers + select live here because they share infrastructure:
// both observe predicate transitions in the routine's running env,
// both fire bodies on rising/falling/changes edges, and both need
// to integrate with the event bus for `on`-case delivery.
//
// Design notes (docs/lang/events.md):
//   - `when expr { body }` registers a block-scoped watcher; the
//     body fires on every predicate transition matching the
//     qualifier. Active for the lifetime of the enclosing block.
//   - `select { case... }` blocks until one case becomes ready,
//     runs that case's body, then exits. Deterministic
//     first-declared-wins on simultaneous readiness.
//
// Re-eval cadence: predicates re-evaluate inside dispatchPendingEvents
// (after each yielding action, plus on each select-poll tick).
// Predicates are pure-by-validator so re-eval is cheap.

// watcher is one live `when expr { body }` registration.
type watcher struct {
	predicate ast.Expr
	qualifier ast.WhenQualifier
	body      *ast.Block
	env       *Env  // closure env at registration time
	last      Value // most-recent predicate value (used for edge detection)
}

// fireIf evaluates the watcher's predicate in its captured env and
// returns true if the transition matches its qualifier. Updates
// the watcher's `last` to the current value either way (so
// successive ticks compute deltas correctly).
func (w *watcher) fireIf(ctx context.Context, it *Interpreter) bool {
	cur := it.eval(ctx, w.predicate, w.env)
	fire := false
	switch w.qualifier {
	case ast.WhenBecomesTrue:
		fire = !Truthy(w.last) && Truthy(cur)
	case ast.WhenBecomesFalse:
		fire = Truthy(w.last) && !Truthy(cur)
	case ast.WhenChanges:
		fire = !valueEqual(w.last, cur)
	}
	w.last = cur
	return fire
}

// watcherFrame is the set of watchers registered in one lexical
// scope. A new frame is pushed on block entry, popped on exit.
type watcherFrame []*watcher

// pushWatcherFrame opens a new scope-level frame. Returns a
// cleanup function the caller must defer to pop the frame. Pop
// is unconditional — runs on any exit path including panic.
func (it *Interpreter) pushWatcherFrame() func() {
	it.watcherMu.Lock()
	it.watchers = append(it.watchers, watcherFrame{})
	it.watcherMu.Unlock()
	return func() {
		it.watcherMu.Lock()
		if n := len(it.watchers); n > 0 {
			it.watchers = it.watchers[:n-1]
		}
		it.watcherMu.Unlock()
	}
}

// registerWatcher adds a watcher to the innermost active frame.
// If no frame exists (called outside any block), the watcher is
// silently dropped — the validator should have prevented this.
func (it *Interpreter) registerWatcher(w *watcher) {
	it.watcherMu.Lock()
	defer it.watcherMu.Unlock()
	if len(it.watchers) == 0 {
		return
	}
	top := len(it.watchers) - 1
	it.watchers[top] = append(it.watchers[top], w)
}

// activeWatchers returns a flat snapshot of every watcher in
// every frame, in declaration order (outermost-first). The
// snapshot is safe to iterate without holding the lock.
func (it *Interpreter) activeWatchers() []*watcher {
	it.watcherMu.Lock()
	defer it.watcherMu.Unlock()
	out := []*watcher{}
	for _, f := range it.watchers {
		out = append(out, f...)
	}
	return out
}

// execWhen handles `when expr [becomes ...|changes] { body }` at a
// statement position. Registers a watcher in the current frame,
// then re-evaluates immediately so already-matching predicates
// fire on the same execution beat.
//
// Seeding rules (docs/lang/events.md "Truth semantics"):
//   - WhenBecomesTrue: last=false → fires immediately if predicate
//     currently true (rising edge, register-as-rising-edge)
//   - WhenBecomesFalse: last=true → fires immediately if predicate
//     currently false (symmetric)
//   - WhenChanges: last=current → first re-eval that observes a
//     different value fires; immediate eval does not fire
func (it *Interpreter) execWhen(ctx context.Context, n *ast.WhenStmt, env *Env) {
	w := &watcher{
		predicate: n.Predicate,
		qualifier: n.Qualifier,
		body:      n.Body,
		env:       env,
	}
	switch n.Qualifier {
	case ast.WhenBecomesTrue:
		w.last = Bool(false)
	case ast.WhenBecomesFalse:
		w.last = Bool(true)
	case ast.WhenChanges:
		w.last = it.eval(ctx, n.Predicate, env)
	}
	it.registerWatcher(w)
	it.evalWatchersOnce(ctx)
}

// evalWatchersOnce sweeps every active watcher and fires bodies
// for those whose predicate transitioned per qualifier. Called
// from dispatchPendingEvents (after each yielding action) and
// from select's poll loop. Errors / aborts inside a body
// propagate (caller is the routine; abortSignal unwinds to the
// nearest try/recover or to execBody).
//
// Firing order is outermost-to-innermost, declaration order.
func (it *Interpreter) evalWatchersOnce(ctx context.Context) {
	ws := it.activeWatchers()
	for _, w := range ws {
		if err := ctx.Err(); err != nil {
			return
		}
		if w.fireIf(ctx, it) {
			it.execBlock(ctx, w.body, w.env.Child())
		}
	}
}

// valueEqual is a light-weight equality test used for `changes`
// detection. Defined on Bool/Int/Float/String/Null; reference
// types fall back to pointer identity, which over-fires (different
// view instances for the same entity look "changed"). That's the
// conservative direction — a refinement test on view-IDs can
// land later.
func valueEqual(a, b Value) bool {
	if a == nil || b == nil {
		return a == b
	}
	switch av := a.(type) {
	case Bool:
		bv, ok := b.(Bool)
		return ok && av == bv
	case Int:
		bv, ok := b.(Int)
		return ok && av == bv
	case Float:
		bv, ok := b.(Float)
		return ok && av == bv
	case String:
		bv, ok := b.(String)
		return ok && av == bv
	case Null:
		_, ok := b.(Null)
		return ok
	}
	return a == b
}

// ----- event listeners (used by select-on cases) -----

// eventListener is a Go-side subscription to a bus event. Used by
// select-on cases to capture the event's args into a side channel
// without going through the regular on-handler path.
type eventListener struct {
	name string
	fn   func(args []Value)
}

// addEventListener registers fn for events of the given name.
// Returns a remove function to undo it.
func (it *Interpreter) addEventListener(name string, fn func(args []Value)) func() {
	it.listenerMu.Lock()
	defer it.listenerMu.Unlock()
	l := &eventListener{name: name, fn: fn}
	it.listeners = append(it.listeners, l)
	return func() {
		it.listenerMu.Lock()
		defer it.listenerMu.Unlock()
		for i, x := range it.listeners {
			if x == l {
				it.listeners = append(it.listeners[:i], it.listeners[i+1:]...)
				return
			}
		}
	}
}

// fireEventListeners is called by dispatchPendingEvents alongside
// the AST-handler path to deliver an event to any registered
// Go-side listeners. Listener fns must not panic and should return
// quickly (they run on the interpreter goroutine).
func (it *Interpreter) fireEventListeners(name string, args []Value) {
	it.listenerMu.Lock()
	ls := make([]*eventListener, 0, len(it.listeners))
	for _, l := range it.listeners {
		if l.name == name {
			ls = append(ls, l)
		}
	}
	it.listenerMu.Unlock()
	for _, l := range ls {
		l.fn(args)
	}
}

// ----- select{} runtime -----

// whenCaseState tracks one when-case's polling state inside
// execSelect. Mirrors the watcher struct but is local to the
// select's lifetime and doesn't participate in the watcher
// registry — select fires once and exits, watchers fire repeatedly.
type whenCaseState struct {
	idx  int
	expr ast.Expr
	qual ast.WhenQualifier
	body *ast.Block
	last Value
}

// fireIf — same shape as watcher.fireIf but for a select case.
func (s *whenCaseState) fireIf(ctx context.Context, it *Interpreter, env *Env) bool {
	cur := it.eval(ctx, s.expr, env)
	fire := false
	switch s.qual {
	case ast.WhenBecomesTrue:
		fire = !Truthy(s.last) && Truthy(cur)
	case ast.WhenBecomesFalse:
		fire = Truthy(s.last) && !Truthy(cur)
	case ast.WhenChanges:
		fire = !valueEqual(s.last, cur)
	}
	s.last = cur
	return fire
}

// onFired is a delivered on-case event waiting to be matched
// against the select's cases.
type onFired struct {
	caseIdx int
	args    []Value
	params  []string
}

// timeoutCase is one timeout-case's poll state.
type timeoutCase struct {
	caseIdx int
	dur     time.Duration
	body    *ast.Block
}

// selectTickInterval is how often the poll loop re-evaluates
// when-case predicates. Chosen short enough to feel responsive
// (~10 ticks/sec) but long enough that watcher predicates with
// many accessors don't dominate CPU. Override per-test via
// Interpreter.SelectTickOverride.
var selectTickInterval = 100 * time.Millisecond

// execSelect runs a `select { ... }` block. The semantics
// (docs/lang/events.md "select"):
//   - Evaluate each when-case at entry; if any is already in its
//     target state, the first one declaration-order fires
//     immediately.
//   - Otherwise: subscribe to on-case events, start the timeout
//     timer, enter the poll loop. Each tick: re-eval when-cases,
//     check for events, check timeouts. First-declared wins.
//   - break/continue from a case body propagate to the enclosing
//     loop (we don't trap them).
//
// Cancellation: ctx.Done() unwinds via the eval's normal ctx
// check; if ctx fires while we're blocked in the tick select,
// we panic a RuntimeError so the routine ends cleanly.
func (it *Interpreter) execSelect(ctx context.Context, n *ast.SelectStmt, env *Env) {
	whenStates := []whenCaseState{}
	for i := range n.Cases {
		c := &n.Cases[i]
		if c.Kind != ast.SelectWhenCase {
			continue
		}
		s := whenCaseState{idx: i, expr: c.Predicate, qual: c.Qualifier, body: c.Body}
		switch c.Qualifier {
		case ast.WhenBecomesTrue:
			s.last = Bool(false)
		case ast.WhenBecomesFalse:
			s.last = Bool(true)
		case ast.WhenChanges:
			s.last = it.eval(ctx, c.Predicate, env)
		}
		whenStates = append(whenStates, s)
	}

	// Pre-check: any when-case already in target state fires now.
	for j := range whenStates {
		s := &whenStates[j]
		if s.fireIf(ctx, it, env) {
			it.execBlock(ctx, s.body, env.Child())
			return
		}
	}

	// Set up on-case listeners. Each writes (caseIndex, args) to
	// a buffered side-channel. Buffer is one slot per case so a
	// burst doesn't block the publisher.
	onCh := make(chan onFired, max(1, len(n.Cases)))
	removers := []func(){}
	for i := range n.Cases {
		c := &n.Cases[i]
		if c.Kind != ast.SelectOnCase {
			continue
		}
		caseIdx := i
		params := c.EventParams
		remove := it.addEventListener(c.EventName, func(args []Value) {
			cp := make([]Value, len(args))
			copy(cp, args)
			select {
			case onCh <- onFired{caseIdx: caseIdx, args: cp, params: params}:
			default:
				// Channel full — drop. Select fires once, so
				// queueing more than one event per case is unneeded.
			}
		})
		removers = append(removers, remove)
	}
	defer func() {
		for _, r := range removers {
			r()
		}
	}()

	// Index timeout cases.
	timeouts := []timeoutCase{}
	for i := range n.Cases {
		c := &n.Cases[i]
		if c.Kind == ast.SelectTimeoutCase {
			timeouts = append(timeouts, timeoutCase{
				caseIdx: i,
				dur:     time.Duration(c.TimeoutMillis) * time.Millisecond,
				body:    c.Body,
			})
		}
	}

	tickD := selectTickInterval
	if it.SelectTickOverride > 0 {
		tickD = it.SelectTickOverride
	}
	ticker := time.NewTicker(tickD)
	defer ticker.Stop()

	start := time.Now()

	for {
		if err := ctx.Err(); err != nil {
			panic(newError(n.Position, "select cancelled: %v", err))
		}

		// Drain pending on-fires; lowest caseIdx wins among on-cases.
		var bestOn *onFired
		drainLoop:
		for {
			select {
			case f := <-onCh:
				if bestOn == nil || f.caseIdx < bestOn.caseIdx {
					tmp := f
					bestOn = &tmp
				}
			default:
				break drainLoop
			}
		}

		// Re-eval when-cases (lowest caseIdx among fires wins).
		bestWhenIdx := -1
		var bestWhenBody *ast.Block
		for j := range whenStates {
			s := &whenStates[j]
			if s.fireIf(ctx, it, env) {
				if bestWhenIdx == -1 || s.idx < bestWhenIdx {
					bestWhenIdx = s.idx
					bestWhenBody = s.body
				}
			}
		}

		// Check timeouts (lowest caseIdx among elapsed wins).
		bestTimeoutIdx := -1
		var bestTimeoutBody *ast.Block
		elapsed := time.Since(start)
		for _, tc := range timeouts {
			if elapsed >= tc.dur {
				if bestTimeoutIdx == -1 || tc.caseIdx < bestTimeoutIdx {
					bestTimeoutIdx = tc.caseIdx
					bestTimeoutBody = tc.body
				}
			}
		}

		// First-declared wins across all three ready buckets.
		winnerIdx := -1
		var winnerBody *ast.Block
		var winnerParams []string
		var winnerArgs []Value
		consider := func(idx int, body *ast.Block, params []string, args []Value) {
			if idx == -1 {
				return
			}
			if winnerIdx == -1 || idx < winnerIdx {
				winnerIdx = idx
				winnerBody = body
				winnerParams = params
				winnerArgs = args
			}
		}
		consider(bestWhenIdx, bestWhenBody, nil, nil)
		if bestOn != nil {
			consider(bestOn.caseIdx, n.Cases[bestOn.caseIdx].Body, bestOn.params, bestOn.args)
		}
		consider(bestTimeoutIdx, bestTimeoutBody, nil, nil)

		if winnerIdx != -1 {
			caseEnv := env.Child()
			for i, p := range winnerParams {
				if i < len(winnerArgs) {
					caseEnv.Define(p, winnerArgs[i])
				} else {
					caseEnv.Define(p, Null{})
				}
			}
			it.execBlock(ctx, winnerBody, caseEnv)
			return
		}

		// No case ready — drain bus events so handlers + listeners
		// can run, then wait for the next tick OR an immediate
		// on-case fire.
		if it.routineEnv != nil {
			it.dispatchPendingEvents(ctx, it.routineEnv)
		}

		select {
		case <-ctx.Done():
			panic(newError(n.Position, "select cancelled: %v", ctx.Err()))
		case <-ticker.C:
		case f := <-onCh:
			// Re-enqueue at head so the next iteration's drain
			// picks it up. Non-blocking: buffer always has room
			// after our own drain at top of loop.
			select {
			case onCh <- f:
			default:
			}
		}
	}
}
