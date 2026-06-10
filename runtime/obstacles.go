package runtime

import (
	"strings"
	"sync"
	"time"
)

// RefusalKind classifies the server's prose answer to a traversal attempt —
// the strings are VERIFIED against the OpenRSC plugin sources (DoorAction,
// BorderGuard, quest/door plugins).
type RefusalKind int

const (
	RefusalUnknown RefusalKind = iota
	// RefusalLocked: "the door is locked" — needs a key or is quest-stage
	// gated; re-clicking will never open it.
	RefusalLocked
	// RefusalToll: the Al-Kharid family — passage is bought by TALKING to
	// the guard NPC ("You must pay a toll of 10 gold coins to pass" /
	// "You need to talk to the border guard"); the gate itself refuses
	// direct interaction.
	RefusalToll
	// RefusalRequirement: level / quest-points / members gating ("you need
	// a mining level of 60", "members only", quest prose).
	RefusalRequirement
	// RefusalTutorial: tutorial-island stage doors ("speak to the X before
	// going through this door").
	RefusalTutorial
	// RefusalWrongChannel: "Nothing interesting happens" — the interaction
	// hit an unhandled command/channel (e.g. object-channel click on
	// something the server dispatches as a boundary, or vice versa).
	RefusalWrongChannel
)

// classifyRefusal maps a server message to a RefusalKind plus the
// precondition the host could satisfy (human-readable; fed to cognition).
func classifyRefusal(msg string) (RefusalKind, string) {
	m := strings.ToLower(msg)
	switch {
	case m == "":
		return RefusalUnknown, ""
	case strings.Contains(m, "door is locked") || strings.Contains(m, "is locked"):
		return RefusalLocked, "needs a key or a quest stage"
	case strings.Contains(m, "pay a toll") || strings.Contains(m, "talk to the border guard"):
		return RefusalToll, "talk to the border guard and pay the toll (10 coins)"
	case strings.Contains(m, "members only") || strings.Contains(m, "member's") ||
		strings.Contains(m, "you need a") || strings.Contains(m, "quest point") ||
		strings.Contains(m, "level of"):
		return RefusalRequirement, msg
	case strings.Contains(m, "before going through this door") || strings.Contains(m, "speak to the"):
		return RefusalTutorial, msg
	case strings.Contains(m, "nothing interesting happens"):
		return RefusalWrongChannel, "wrong interaction channel — retry once via the other channel"
	default:
		return RefusalUnknown, msg
	}
}

// BlockedEdge is one obstacle the host has LEARNED it cannot currently pass:
// a locked door, a toll gate, a quest-gated barrier. Remembering these stops
// the traversal flow from re-trying the same locked door on every replan and
// lets route planning cost gated paths honestly.
type BlockedEdge struct {
	X, Y, Dir    int
	Kind         RefusalKind
	Prose        string // the server's own words
	Precondition string // what would unblock it, per classifyRefusal
	At           time.Time
}

// blockedEdgeTTL bounds how long a learned block suppresses retries. Long
// enough to stop spinning, short enough that a state change (quest done,
// coins acquired, someone opened the door) gets re-discovered.
const blockedEdgeTTL = 5 * time.Minute

// blockedEdges is the host's ledger of learned-impassable obstacles.
type blockedEdges struct {
	mu sync.Mutex
	m  map[[3]int]BlockedEdge
}

func newBlockedEdges() *blockedEdges {
	return &blockedEdges{m: map[[3]int]BlockedEdge{}}
}

// Note records (or refreshes) a learned block.
func (b *blockedEdges) Note(e BlockedEdge) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e.At = time.Now()
	b.m[[3]int{e.X, e.Y, e.Dir}] = e
}

// Blocked reports whether the obstacle at (x, y, dir) is currently ledgered
// (within TTL); expired entries are dropped lazily.
func (b *blockedEdges) Blocked(x, y, dir int) (BlockedEdge, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	k := [3]int{x, y, dir}
	e, ok := b.m[k]
	if !ok {
		return BlockedEdge{}, false
	}
	if time.Since(e.At) > blockedEdgeTTL {
		delete(b.m, k)
		return BlockedEdge{}, false
	}
	return e, true
}

// All returns the live (unexpired) ledger entries.
func (b *blockedEdges) All() []BlockedEdge {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]BlockedEdge, 0, len(b.m))
	for k, e := range b.m {
		if time.Since(e.At) > blockedEdgeTTL {
			delete(b.m, k)
			continue
		}
		out = append(out, e)
	}
	return out
}
