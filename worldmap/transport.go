package worldmap

// The transport layer overlays GAME-LOGIC passages (ferries, toll gates,
// guild doors, post-quest spirit trees, ...) on top of the pure-collision
// walking components. The static collision engine deliberately knows
// nothing about coins/quests/items; the transport table is where those
// host CAPABILITIES gate movement.
//
// The durable ground truth is transport.json (embedded below), normalized
// from the 431 OpenRSC server-plugin passage records by
// tools/normalize_transport.py. We load the STANDING subset: repeatable
// and post_quest_permanent passages, plus conditional_other passages whose
// requirement is a host capability (coins / item / skill / quest-complete /
// members). One-time-quest cutscene moves and lever/puzzle/mid-quest-stage
// conditionals are NOT navigation and are excluded. Records whose source
// tile resolves only to an unprintable loc/bound id (per the index
// coverage_gaps) are counted as skipped, never invented.

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// transportJSON is the embedded, version-controlled transport ground truth.
//
//go:embed transport.json
var transportJSON []byte

// Requirement is a parsed host-capability gate on a transport edge. Exactly
// one discriminant is the binding gate (others are zero/empty), except the
// toll gate which carries Coins plus a QuestFree free-pass clause.
type Requirement struct {
	// Coins is the toll/fare in gold coins (0 = no coin gate).
	Coins int `json:"coins,omitempty"`
	// Item is a single consumable/held item the host must carry (e.g.
	// "Shantay Desert Pass", "Ship ticket"). Empty = no item gate.
	Item string `json:"item,omitempty"`
	// SkillName / SkillLevel is a skill-level gate (e.g. Fishing >= 68).
	// SkillName empty = no skill gate.
	SkillName  string `json:"skill_name,omitempty"`
	SkillLevel int    `json:"skill_level,omitempty"`
	// QuestDone is a quest that must be COMPLETE for the edge to stand
	// (the post_quest_permanent predicate, getQuestStage==-1). Empty =
	// no quest gate.
	QuestDone string `json:"quest_done,omitempty"`
	// Members is true if the edge only exists on a members world.
	Members bool `json:"members,omitempty"`
	// None is true for an always-traversable standing edge (no capability).
	None bool `json:"none,omitempty"`

	// QuestFree names a quest whose COMPLETION waives the Coins toll (the
	// Al-Kharid toll gate is free once Prince Ali Rescue is done). Only the
	// toll gate uses this.
	QuestFree string `json:"quest_free,omitempty"`
}

// Barrier is an axis-aligned conditional cut the transport layer overlays
// on the collision flood. It exists ONLY for game-logic gates that have no
// collision wall of their own — the Al-Kharid toll gate being the canonical
// case: the border is fully walkable in the collision map, so the toll can
// only be modeled by overlaying a cut on the authentic border line.
//
// A Barrier with Axis "x" cuts every step that crosses between x<Line and
// x>=Line (Axis "y" symmetric), for destination rows/cols in [Lo,Hi]. The
// per-query capability-flood may cross it ONLY if the edge's Requirement is
// satisfied by the host's Capability.
type Barrier struct {
	Axis string `json:"axis"` // "x" or "y"
	Line int    `json:"line"` // the cut coordinate
	Lo   int    `json:"lo"`   // inclusive extent along the other axis
	Hi   int    `json:"hi"`   // inclusive extent along the other axis
}

// TransportEdge is one standing passage. Two shapes:
//   - Kind "teleport": boarding at From (if resolved) jumps the host to To
//     when Capability satisfies Req. From may be unset (board-from-mainland).
//   - Kind "gate": a conditional Barrier cut (only the toll gate carries one);
//     other "gate" rows are informational (an openable door with a skill/quest
//     requirement) and have no flood effect in v1.
type TransportEdge struct {
	Kind           string      `json:"kind"`     // "teleport" | "gate"
	Category       string      `json:"category"` // ferry / toll / gate / ladder / ...
	Name           string      `json:"name"`
	SourceFile     string      `json:"source_file"`
	From           []int       `json:"from,omitempty"` // [x,y] board/approach tile
	To             []int       `json:"to,omitempty"`   // [x,y] teleport target tile
	Barrier        *Barrier    `json:"barrier,omitempty"`
	Req            Requirement `json:"req"`
	Traversability string      `json:"traversability"`
}

// transportTable is the decoded transport.json document.
type transportTable struct {
	Edges          []TransportEdge `json:"edges"`
	Skipped        int             `json:"skipped"`
	SkippedReasons map[string]int  `json:"skipped_reasons"`
}

// Capability is the per-query host state the reachability flood honors. It
// is supplied PER QUERY and never stored on the shared Oracle, so the
// precomputed engine stays read-only and shareable by pointer.
type Capability struct {
	Coins      int
	Items      map[string]bool // item-name -> held
	Skills     map[string]int  // skill-name -> current level
	QuestsDone map[string]bool // quest-name -> complete
	Members    bool
}

// Satisfies reports whether this Capability meets the edge Requirement. The
// rules mirror the server game-logic the records were extracted from:
//   - None: always satisfied.
//   - Coins: host must carry at least Req.Coins, UNLESS QuestFree names a
//     completed quest (the toll gate's free-pass clause).
//   - Item / Skill / Quest / Members: the matching capability must be held.
//
// Multiple populated fields are ANDed (e.g. Shantay = Members AND the pass).
// Name matching for items/skills/quests is case-insensitive and substring-
// tolerant so the free-text record names line up with host capability keys.
func (c Capability) Satisfies(r Requirement) bool {
	if r.None {
		return true
	}
	// Quest-waived toll: Prince Ali Rescue complete makes the gate free.
	coinWaived := false
	if r.QuestFree != "" && c.questDone(r.QuestFree) {
		coinWaived = true
	}
	if r.Coins > 0 && !coinWaived {
		if c.Coins < r.Coins {
			return false
		}
	}
	if r.Members && !c.Members {
		return false
	}
	if r.Item != "" && !c.hasItem(r.Item) {
		return false
	}
	if r.SkillName != "" {
		if c.skillLevel(r.SkillName) < r.SkillLevel {
			return false
		}
	}
	if r.QuestDone != "" && !c.questDone(r.QuestDone) {
		return false
	}
	return true
}

func (c Capability) hasItem(name string) bool {
	for k, v := range c.Items {
		if v && containsFold(k, name) {
			return true
		}
	}
	return false
}

func (c Capability) skillLevel(name string) int {
	best := 0
	for k, lvl := range c.Skills {
		if containsFold(k, name) && lvl > best {
			best = lvl
		}
	}
	return best
}

func (c Capability) questDone(name string) bool {
	for k, v := range c.QuestsDone {
		if v && (containsFold(k, name) || containsFold(name, k)) {
			return true
		}
	}
	return false
}

// loadTransport decodes the embedded transport.json into the Oracle. It is
// called from Precompute after the collision engine is built. It records
// the loaded edges plus the skipped count for telemetry.
func (o *Oracle) loadTransport() error {
	var t transportTable
	if err := json.Unmarshal(transportJSON, &t); err != nil {
		return fmt.Errorf("worldmap: decode transport.json: %w", err)
	}
	o.edges = t.Edges
	o.edgesSkipped = t.Skipped
	return nil
}

// Edges returns the loaded standing transport edges. The returned slice must
// not be mutated.
func (o *Oracle) Edges() []TransportEdge { return o.edges }

// EdgesLoaded / EdgesSkipped report the transport-table accounting.
func (o *Oracle) EdgesLoaded() int  { return len(o.edges) }
func (o *Oracle) EdgesSkipped() int { return o.edgesSkipped }
