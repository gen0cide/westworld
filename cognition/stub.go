package cognition

import (
	"context"
	"strings"
	"time"
)

// StubClient is a deterministic, zero-I/O implementation of Client
// used for tests and early integration before the real mesa-backed
// client lands in Phase 3.
//
// The stub inspects Retrieval.Goal with case-insensitive substring
// matching and synthesizes a Bundle. The matching heuristics are
// stable + documented (see Retrieve) so routine + integration tests
// can rely on specific reflections/episodes showing up for specific
// goals.
//
// StubClient holds no state and is safe for concurrent use.
//
// The optional Now field overrides the time source (for tests that
// want a fixed timestamp). If nil, time.Now is used.
type StubClient struct {
	// Now is the time source for Bundle.Timestamp. If nil,
	// time.Now is used. Override in tests to assert exact values.
	Now func() time.Time
}

// NewStubClient returns a StubClient using time.Now as its clock.
func NewStubClient() *StubClient {
	return &StubClient{}
}

// defaultMaxItems is what Retrieve uses when r.MaxItems == 0.
const defaultMaxItems = 3

// Retrieve returns a canned Bundle keyed off Retrieval.Goal. The
// heuristics, in priority order:
//
//   - Goal contains "combat" or "fight" or "attack" → reflections
//     about aggressive style + episodic memory of a past goblin kill.
//   - Goal contains "bank" → reflections about banking habits +
//     episodic memory of a past deposit.
//   - Goal contains "chat" or "greet" or "speak" → reflections
//     about social style + episodic memory of a recent conversation.
//   - Goal contains "fish" or "cook" or "mine" → reflections about
//     skilling preferences + episodic memory of a past skilling run.
//   - Any other goal → a generic exploration reflection + episodic.
//
// The Bundle is always populated with at least one reflection and
// one episodic entry (truncated to MaxItems if that's smaller).
// Persona is a small fixed map. WorldSnapshot is populated only if
// Retrieval.IncludeWorld is true.
//
// Retrieve never returns an error in the stub; the signature is
// preserved for parity with the real client.
func (c *StubClient) Retrieve(ctx context.Context, r Retrieval) (*Bundle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	max := r.MaxItems
	if max <= 0 {
		max = defaultMaxItems
	}

	reflections, episodic := cannedFor(r.Goal)
	reflections = truncate(reflections, max)
	episodic = truncate(episodic, max)

	b := &Bundle{
		Goal:        r.Goal,
		Reflections: reflections,
		Episodic:    episodic,
		Persona:     cannedPersona(r.HostName),
		Timestamp:   c.now(),
	}
	if r.IncludeWorld {
		b.WorldSnapshot = cannedWorld(r.HostName)
	}
	return b, nil
}

func (c *StubClient) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// cannedFor returns (reflections, episodic) for a goal string. The
// heuristics are documented on Retrieve.
func cannedFor(goal string) ([]string, []string) {
	g := strings.ToLower(goal)
	switch {
	case containsAny(g, "combat", "fight", "attack"):
		return []string{
				"I tend to favor an aggressive style.",
				"I retreat once my hp drops below a third.",
				"I prefer fighting one foe at a time.",
			}, []string{
				"Last week I killed a goblin near Lumbridge for 12 xp.",
				"Yesterday I was nearly killed by a moss giant.",
				"I once fled from a bear in Varrock forest.",
			}
	case containsAny(g, "bank"):
		return []string{
				"I bank as soon as my inventory is nearly full.",
				"I keep ranged ammo stocked at Falador.",
				"I never carry more than 1k coins out of the bank.",
			}, []string{
				"This morning I deposited 28 lobsters at Edgeville.",
				"Last session I forgot to withdraw a pickaxe.",
				"I once lost a load to a pkr because I banked too late.",
			}
	case containsAny(g, "chat", "greet", "speak", "say"):
		return []string{
				"I'm friendly but rarely the one to start a conversation.",
				"I avoid sharing my location with strangers.",
				"I respond to greetings within a few seconds.",
			}, []string{
				"Earlier today someone said 'hi' near the Lumbridge well.",
				"A player asked me for a free fish; I declined politely.",
				"I traded jokes with a friend at the Falador fountain.",
			}
	case containsAny(g, "fish", "cook", "mine", "smith", "skill"):
		return []string{
				"I enjoy long skilling sessions to clear my mind.",
				"I take short breaks every ten inventories to reduce fatigue.",
				"I prefer skilling at quieter spots away from PvP zones.",
			}, []string{
				"Last week I caught 200 trout at Barbarian Village.",
				"I burned half a stack of cooked salmon last session.",
				"I once found a clue scroll while mining iron.",
			}
	default:
		return []string{
				"I am curious and tend to wander when I have no goal.",
				"I prefer to plan a route before walking far.",
			}, []string{
				"Earlier I walked from Lumbridge to Draynor without incident.",
				"I noticed a player following me near the Varrock gate.",
			}
	}
}

func cannedPersona(host string) map[string]string {
	// Keyed by host name so tests can distinguish multiple hosts if
	// they want to, but the values themselves are stable canned
	// labels.
	name := host
	if name == "" {
		name = "anon"
	}
	return map[string]string{
		"name":       name,
		"mood":       "calm",
		"style":      "cautious",
		"north_star": "explore the world and stay alive",
	}
}

func cannedWorld(host string) string {
	// Deterministic single-line world summary. Real client builds
	// this from world.World.Snapshot().
	if host == "" {
		return "Standing at Lumbridge spawn (120, 648). Daytime. No other players nearby."
	}
	return "Host " + host + " standing at Lumbridge spawn (120, 648). Daytime. No other players nearby."
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func truncate(xs []string, max int) []string {
	if max <= 0 || len(xs) <= max {
		return xs
	}
	out := make([]string, max)
	copy(out, xs[:max])
	return out
}
