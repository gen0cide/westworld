package runtime

import "sync"

// liveGoalState carries an operator goal pushed at runtime — a mesa
// Provision.Subscribe GOAL_REVISION, ultimately from `mesa-ctl goal push` — to
// the director. It is a SOFT override of the goal the director was constructed
// with (genesis/persona): the director prefers it when set, and falls back to
// its construction-time goal when empty. Latest-wins, NOT consume-once — every
// turn reads the current value, so a single push steers all subsequent planning
// until another push replaces it. Guarded by its own mutex (cheap, contended
// only by the subscribe goroutine writing and the director reading per turn).
type liveGoalState struct {
	mu   sync.Mutex
	goal string
}

func (l *liveGoalState) set(goal string) {
	l.mu.Lock()
	l.goal = goal
	l.mu.Unlock()
}

func (l *liveGoalState) get() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.goal
}
