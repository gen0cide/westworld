package runtime

import (
	"context"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// authoredIntent is what a fake mesa planner returns: a WRITE_ROUTINE move (a
// repeatable grind — NOT a one-shot).
func authoredIntent() Intent {
	return Intent{Label: "act:mine", Name: "mine", Source: "runtime \"1.0\"\nroutine mine() { wait(1) }"}
}

// oneShotIntent is a single action (e.g. a say or a direct action) — it must never
// be cached + replayed.
func oneShotIntent() Intent {
	return Intent{Label: "act:say", Name: "act_direct", Source: "runtime \"1.0\"\nroutine act_direct() { wait(0) }", OneShot: true}
}

func success(in Intent) Outcome { return Outcome{Intent: in, Kind: interp.ResultCompleted} }
func failure(in Intent) Outcome {
	return Outcome{Intent: in, Kind: interp.ResultErrored, Err: &interp.RuntimeError{Msg: "boom"}}
}

// budgetExpired is what the conductor reports when the per-turn budget cuts a
// routine off mid-work: an errored Kind, but classified BudgetExpired.
func budgetExpired(in Intent) Outcome {
	return Outcome{Intent: in, Kind: interp.ResultErrored,
		Err: &interp.RuntimeError{Msg: "wait: context deadline exceeded"}, BudgetExpired: true}
}

// TestHybridDirectorPromotesAndReplays proves the cheap loop: a novel situation
// pays one LLM (Act) call; once that authored routine succeeds it is promoted,
// and the same situation then replays it from the library WITHOUT calling Act.
func TestHybridDirectorPromotesAndReplays(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	// Turn 1: cache miss → escalates to the (fake) planner.
	i1, ok := d.Next(ctx, h, Outcome{})
	if !ok || i1.Source == "" {
		t.Fatalf("turn 1 should author a routine: ok=%v src=%q", ok, i1.Source)
	}
	if calls != 1 {
		t.Fatalf("turn 1 calls=%d, want 1", calls)
	}

	// Turn 2: the routine's run visibly moved the world (mining xp climbed) and it
	// completed → promote, then REPLAY from the library. Progress is REQUIRED for
	// promotion (#30a) — see TestHybridDirectorNoPromotionWithoutProgress.
	h.world.Self.SetSkill(8, 1, 99, 10)
	i2, ok := d.Next(ctx, h, success(i1))
	if !ok {
		t.Fatal("turn 2 should return an intent")
	}
	if calls != 1 {
		t.Fatalf("turn 2 hit the planner (calls=%d); it should replay from the library", calls)
	}
	if !strings.HasPrefix(i2.Label, "lib:") || i2.Source == "" {
		t.Fatalf("turn 2 should replay a library routine, got label=%q", i2.Label)
	}
	if lib.Len() != 1 {
		t.Fatalf("library size = %d, want 1", lib.Len())
	}
}

// TestHybridDirectorDoesNotCacheOneShot proves the no-op-spin fix: a one-shot
// action (a single say / direct action / idle) is NEVER promoted, so it can't
// become a cached routine that replays forever doing nothing — the host re-decides
// each turn instead.
func TestHybridDirectorDoesNotCacheOneShot(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return oneShotIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "socialize", nil)

	i1, _ := d.Next(ctx, h, Outcome{})   // escalate → one-shot
	i2, _ := d.Next(ctx, h, success(i1)) // report success — must NOT promote it
	if lib.Len() != 0 {
		t.Fatalf("a one-shot action must never be cached; library size = %d", lib.Len())
	}
	if strings.HasPrefix(i2.Label, "lib:") {
		t.Fatalf("turn 2 replayed a cached one-shot (%q) — it should re-decide via the planner", i2.Label)
	}
	if calls != 2 {
		t.Fatalf("a one-shot should re-escalate each turn (no replay); calls=%d", calls)
	}
}

// TestHybridDirectorEvictsFailingReplay proves self-healing: a replayed routine
// that fails is evicted, so the next turn re-escalates to the planner.
func TestHybridDirectorEvictsFailingReplay(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{})   // author (calls=1)
	h.world.Self.SetSkill(8, 1, 99, 10)  // world progress → earns promotion (#30a)
	i2, _ := d.Next(ctx, h, success(i1)) // promote + replay (calls=1)
	if calls != 1 || lib.Len() != 1 {
		t.Fatalf("precondition: calls=%d libsize=%d", calls, lib.Len())
	}

	// Turn 3: report the REPLAY as failed → evict → miss → re-author.
	i3, ok := d.Next(ctx, h, failure(i2))
	if !ok {
		t.Fatal("turn 3 should return")
	}
	if lib.Len() != 0 {
		t.Fatalf("failing replay should be evicted, library size = %d", lib.Len())
	}
	if calls != 2 || i3.Source == "" {
		t.Fatalf("turn 3 should re-author after eviction: calls=%d", calls)
	}
}

// TestHybridDirectorRevalidates proves the re-validation cap: a stable situation
// replays cheaply but consults the planner again at least every
// maxConsecutiveReuse turns, so it can't loop a stale routine forever.
func TestHybridDirectorRevalidates(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	last := Outcome{}
	replays := 0
	// Run enough turns to force at least one re-validation after the first promote.
	for turn := 0; turn < maxConsecutiveReuse+4; turn++ {
		// Simulate a real grind MAKING progress (mining xp climbing) so the
		// world-progress stall detector doesn't trip: xp is in progressKey but NOT
		// in the situation signature, so the signature stays stable (cheap loop
		// replays) — exactly the revalidation case under test, not a stuck no-op.
		h.world.Self.SetSkill(8, 1, 99, (turn+1)*10)
		in, ok := d.Next(ctx, h, last)
		if !ok {
			t.Fatal("director stopped unexpectedly")
		}
		if strings.HasPrefix(in.Label, "lib:") {
			replays++
		}
		last = success(in)
	}
	// Over maxConsecutiveReuse+4 turns the planner should be hit only a couple of
	// times (initial author + re-validation), the rest replayed locally.
	if calls < 2 {
		t.Fatalf("expected a re-validation Act call after the reuse cap; calls=%d", calls)
	}
	if calls >= maxConsecutiveReuse {
		t.Fatalf("planner called too often (calls=%d over %d turns); cheap loop not saving LLM calls", calls, maxConsecutiveReuse+4)
	}
	if replays < maxConsecutiveReuse-1 {
		t.Fatalf("expected ~%d local replays, got %d", maxConsecutiveReuse, replays)
	}
}

// TestHybridDirectorResumesBudgetExpiredProgressingGrind is the soak-retro #5
// regression at the cheap-loop layer: a routine that exceeds the turn budget
// WHILE the world-progress key keeps changing (xp climbing) is a WORKING grind —
// the director resumes the SAME intent with no LLM call instead of re-planning,
// and nothing escalates toward BLOCKED.
func TestHybridDirectorResumesBudgetExpiredProgressingGrind(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{}) // turn 1: cache miss → author (calls=1)
	if calls != 1 {
		t.Fatalf("precondition: calls=%d, want 1", calls)
	}

	// The grind outlives the turn budget for several consecutive turns, but the
	// world-progress key keeps changing (mining xp climbs): each turn must RESUME
	// the same intent with no planner call.
	last := budgetExpired(i1)
	for turn := 0; turn < 6; turn++ {
		h.world.Self.SetSkill(8, 1, 99, (turn+1)*10) // world progress: xp climbing
		in, ok := d.Next(ctx, h, last)
		if !ok {
			t.Fatal("director stopped unexpectedly")
		}
		if in.Label != i1.Label || in.Source != i1.Source {
			t.Fatalf("turn %d: expected the same grind resumed, got label=%q", turn, in.Label)
		}
		if calls != 1 {
			t.Fatalf("turn %d: budget expiry on a progressing grind consulted the planner (calls=%d, want 1)", turn, calls)
		}
		last = budgetExpired(in)
	}
}

// TestHybridDirectorBudgetExpiryWithoutProgressEscalates pins the boundary of
// the resume path: a budget-expired routine in a FROZEN world (progressKey
// unchanged) is not resumed blindly — it falls through to the normal planner
// path, where the stall machinery can do its job.
func TestHybridDirectorBudgetExpiryWithoutProgressEscalates(t *testing.T) {
	h := newTestHost() // frozen world — nothing changes turn to turn
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{}) // author (calls=1)
	_, ok := d.Next(ctx, h, budgetExpired(i1))
	if !ok {
		t.Fatal("director stopped unexpectedly")
	}
	if calls != 2 {
		t.Fatalf("a no-progress budget expiry should re-consult the planner; calls=%d, want 2", calls)
	}
}

// TestHybridDirectorDoesNotEvictBudgetExpiredReplay proves a learned library
// routine is NOT evicted when a replay of it merely runs out of turn time while
// progressing — evicting would discard a working grind (soak retro #5).
func TestHybridDirectorDoesNotEvictBudgetExpiredReplay(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{})   // author (calls=1)
	h.world.Self.SetSkill(8, 1, 99, 10)  // world progress → earns promotion (#30a)
	i2, _ := d.Next(ctx, h, success(i1)) // promote + replay from the library
	if calls != 1 || lib.Len() != 1 {
		t.Fatalf("precondition: calls=%d libsize=%d", calls, lib.Len())
	}

	h.world.Self.SetSkill(8, 1, 99, 50) // world progress since the last turn
	i3, _ := d.Next(ctx, h, budgetExpired(i2))
	if lib.Len() != 1 {
		t.Fatalf("budget-expired replay was evicted from the library (size=%d, want 1)", lib.Len())
	}
	if i3.Label != i2.Label {
		t.Fatalf("expected the budget-expired replay resumed, got label=%q", i3.Label)
	}
	if calls != 1 {
		t.Fatalf("budget-expired replay consulted the planner (calls=%d, want 1)", calls)
	}
}

// TestHybridDirectorBreaksNoProgressLoop is the regression for the live 100%-fatigue
// loop: a routine that "completes" every turn but changes NOTHING in the world must
// not be replayed forever by the cheap loop. With a frozen world (progressKey never
// changes), once the stall threshold is crossed the cached routine is evicted and
// the planner is re-consulted every turn — ZERO cheap replays past the threshold.
func TestHybridDirectorBreaksNoProgressLoop(t *testing.T) {
	h := newTestHost() // frozen world — nothing changes turn to turn
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "recover", nil)

	last := Outcome{}
	replaysAfterStall := 0
	const turns = stallEscalateTurns + 5
	for turn := 0; turn < turns; turn++ {
		in, ok := d.Next(ctx, h, last)
		if !ok {
			t.Fatal("director stopped unexpectedly")
		}
		if turn >= stallEscalateTurns && strings.HasPrefix(in.Label, "lib:") {
			replaysAfterStall++
		}
		last = success(in) // "succeeds" every turn, but the world never advances
	}
	if replaysAfterStall > 0 {
		t.Fatalf("no-progress loop not broken: %d cheap replays past the stall threshold (want 0 — should evict + escalate)", replaysAfterStall)
	}
	if calls < 2 {
		t.Fatalf("planner never re-consulted on stall (calls=%d)", calls)
	}
}

// TestHybridDirectorNoPromotionWithoutProgress is the promotion-on-progress gate
// (#30a / soak-retro fix 14): a routine that completes successfully while the
// world-progress key never moves earned NOTHING — "ran without crashing while
// changing nothing" must not become a cached routine. It is not promoted, and the
// next turn re-consults the planner instead of replaying a do-nothing routine.
func TestHybridDirectorNoPromotionWithoutProgress(t *testing.T) {
	h := newTestHost() // frozen world — nothing changes turn to turn
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{}) // author (calls=1)
	// Turn 2: the routine "succeeded" but the world is byte-identical → no slot.
	i2, ok := d.Next(ctx, h, success(i1))
	if !ok {
		t.Fatal("director stopped unexpectedly")
	}
	if lib.Len() != 0 {
		t.Fatalf("a zero-progress routine was promoted (library size = %d, want 0)", lib.Len())
	}
	if strings.HasPrefix(i2.Label, "lib:") {
		t.Fatalf("turn 2 replayed a zero-progress routine (%q) — it should re-consult the planner", i2.Label)
	}
	if calls != 2 {
		t.Fatalf("turn 2 should re-escalate (nothing was cached); calls=%d, want 2", calls)
	}
}

// TestHybridDirectorEvictsNoProgressReplays wires the eviction condition for an
// ALREADY-LEARNED routine that has gone stale: it keeps completing on replay but
// moves nothing, so after noProgressEvictAfter (=2) consecutive no-progress
// replays it loses its slot — well before the stallEscalateTurns backstop — and
// the planner is re-consulted. This is the drone9 cheap-loop disease: cached
// non-progressing routines replaying/evicting/re-authoring forever.
func TestHybridDirectorEvictsNoProgressReplays(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{})   // author (calls=1)
	h.world.Self.SetSkill(8, 1, 99, 10)  // real progress → routine earns its slot
	i2, _ := d.Next(ctx, h, success(i1)) // promote + replay 1
	if calls != 1 || lib.Len() != 1 {
		t.Fatalf("precondition: calls=%d libsize=%d", calls, lib.Len())
	}

	// The world freezes: the cached routine keeps "succeeding" while changing nothing.
	i3, _ := d.Next(ctx, h, success(i2)) // strike 1 → still cached, replays once more
	if lib.Len() != 1 {
		t.Fatalf("evicted after a single no-progress replay (size=%d) — want %d strikes", lib.Len(), noProgressEvictAfter)
	}
	if !strings.HasPrefix(i3.Label, "lib:") {
		t.Fatalf("strike 1 should still replay (benefit of the doubt), got label=%q", i3.Label)
	}
	i4, ok := d.Next(ctx, h, success(i3)) // strike 2 → evict → miss → re-plan
	if !ok {
		t.Fatal("director stopped unexpectedly")
	}
	if lib.Len() != 0 {
		t.Fatalf("routine survived %d consecutive no-progress replays (size=%d, want 0)", noProgressEvictAfter, lib.Len())
	}
	if strings.HasPrefix(i4.Label, "lib:") {
		t.Fatalf("turn after strike-out still replayed from the library (%q)", i4.Label)
	}
	if calls != 2 {
		t.Fatalf("strike-out should re-consult the planner; calls=%d, want 2", calls)
	}
}

// TestHybridDirectorPromotesResumedGrindOnCompletion pins the composition of the
// promotion gate with the budget-resume path (8d362ba): a grind that did its
// visible work during a budget-expired-then-resumed stretch is STILL promoted
// when it finally completes, even if the completing turn itself shows no new
// world delta — the progress credit spans the whole run, not just the last tick.
func TestHybridDirectorPromotesResumedGrindOnCompletion(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	calls := 0
	fake := DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		calls++
		return authoredIntent(), true
	})
	lib := NewRoutineLibrary(nil)
	d := NewHybridDirector(fake, lib, "mine tin", nil)

	i1, _ := d.Next(ctx, h, Outcome{})  // author (calls=1)
	h.world.Self.SetSkill(8, 1, 99, 10) // progress during the first stretch
	i2, _ := d.Next(ctx, h, budgetExpired(i1))
	if i2.Label != i1.Label || calls != 1 {
		t.Fatalf("precondition: progressing budget-expired grind should resume (label=%q calls=%d)", i2.Label, calls)
	}

	// The final stretch completes with NO further world delta — promotion must
	// still happen on the credit earned mid-run.
	d.Next(ctx, h, success(i2))
	if lib.Len() != 1 {
		t.Fatalf("resumed grind that progressed mid-run was not promoted on completion (size=%d, want 1)", lib.Len())
	}
	if calls != 1 {
		t.Fatalf("completion of a learned grind consulted the planner (calls=%d, want 1)", calls)
	}
}
