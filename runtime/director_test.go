package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

func setHPMax(h *Host, cur, max int) {
	var c, m, x [world.NumSkills]int
	for i := range c {
		c[i], m[i] = 30, 30
	}
	c[3], m[3] = cur, max
	h.world.Self.SetAllSkills(c, m, x, 0)
}

func TestGoalDirectorPicksHighestPriorityEligible(t *testing.T) {
	h := newTestHost()
	setHPMax(h, 2, 10) // 20% HP

	idle := Intent{Label: "wander", RoutinePath: "wander.rt"}
	d := NewGoalDirector(&idle,
		RoutineDrive("mine", 10, nil, "mine.rt"),               // always eligible, low priority
		RoutineDrive("heal", 100, WhenHPBelow(0.3), "heal.rt"), // eligible (hurt), high priority
	)
	got, ok := d.Next(context.Background(), h, Outcome{})
	if !ok || got.Label != "heal" {
		t.Fatalf("Next = %+v ok=%v, want heal (highest-priority eligible)", got, ok)
	}
}

func TestGoalDirectorFallsThroughIneligible(t *testing.T) {
	h := newTestHost()
	setHPMax(h, 10, 10) // healthy ⇒ heal not eligible

	idle := Intent{Label: "wander", RoutinePath: "wander.rt"}
	d := NewGoalDirector(&idle,
		RoutineDrive("heal", 100, WhenHPBelow(0.3), "heal.rt"),
		RoutineDrive("mine", 10, nil, "mine.rt"),
	)
	got, _ := d.Next(context.Background(), h, Outcome{})
	if got.Label != "mine" {
		t.Fatalf("Next = %q, want mine (heal ineligible at full HP)", got.Label)
	}
}

func TestGoalDirectorIdleFallback(t *testing.T) {
	h := newTestHost()
	idle := Intent{Label: "wander", RoutinePath: "wander.rt"}
	d := NewGoalDirector(&idle,
		RoutineDrive("heal", 100, WhenHPBelow(0.01), "heal.rt"), // never (HP>1%)
	)
	got, ok := d.Next(context.Background(), h, Outcome{})
	if !ok || got.Label != "wander" {
		t.Fatalf("Next = %+v, want idle wander", got)
	}
}

func TestGoalDirectorNilIdleStops(t *testing.T) {
	h := newTestHost()
	setHPMax(h, 10, 10)
	d := NewGoalDirector(nil,
		RoutineDrive("heal", 100, WhenHPBelow(0.3), "heal.rt"),
	)
	if _, ok := d.Next(context.Background(), h, Outcome{}); ok {
		t.Fatal("Next should stop (ok=false) when nothing eligible and no idle")
	}
}

func TestGoalDirectorBuildDeclineFallsThrough(t *testing.T) {
	h := newTestHost()
	idle := Intent{Label: "idle", RoutinePath: "idle.rt"}
	d := NewGoalDirector(&idle,
		Drive{Name: "declines", Priority: 100, Build: func(*Host) (Intent, bool) { return Intent{}, false }},
		RoutineDrive("backup", 10, nil, "backup.rt"),
	)
	got, _ := d.Next(context.Background(), h, Outcome{})
	if got.Label != "backup" {
		t.Fatalf("Next = %q, want backup (high-priority drive declined)", got.Label)
	}
}

// End-to-end through the Conductor: the executor mutates HP, and the director
// switches from mining to healing on the next turn — proving state-driven
// behavior selection across turns.
func TestGoalDirectorDrivesConductor(t *testing.T) {
	h := newTestHost()
	setHPMax(h, 10, 10) // start healthy

	director := NewGoalDirector(nil,
		RoutineDrive("heal", 100, WhenHPBelow(0.3), "heal.rt"),
		RoutineDrive("mine", 10, nil, "mine.rt"), // always eligible ⇒ default behavior
	)

	ctx, cancel := context.WithCancel(context.Background())
	var ran []string
	// Bind the conductor to the REAL host so the director's guards can read its
	// live state (newTestConductor uses a nil host).
	c := NewConductor(h, ConductorOptions{Director: director, Settle: -1})
	c.execute = func(_ context.Context, in Intent) Outcome {
		ran = append(ran, in.Label)
		switch len(ran) {
		case 1:
			setHPMax(h, 2, 10) // turn 1 (mine) "hurt" us ⇒ turn 2 should heal
		case 2:
			setHPMax(h, 10, 10) // healed ⇒ turn 3 mines again
		case 3:
			cancel() // stop after 3 turns
		}
		return Outcome{Intent: in, Kind: interp.ResultCompleted}
	}
	_ = c.Run(ctx)
	if len(ran) < 3 || ran[0] != "mine" || ran[1] != "heal" || ran[2] != "mine" {
		t.Fatalf("conductor drive sequence = %v, want [mine heal mine]", ran)
	}
}
