package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
)

// Views for the `combat` faculty. The view-root wiring
// (it.Reserved["combat"] = &combatView{...}) lives centrally in
// dsl_bridge.go.

// ---------- combat ----------

type combatView struct{ host *Host }

func (c *combatView) Kind() string    { return "combat" }
func (c *combatView) Display() string { return "<combat>" }

// Combat state isn't tracked yet — runtime/combat_loop.go owns it
// transiently. For now: engaged is false, target is null. Future work
// (step 8) wires the persistent state.
func (c *combatView) Get(field string) (interp.Value, bool) {
	switch field {
	case "engaged":
		return interp.Bool(false), true
	case "target":
		return interp.Null{}, true
	case "last_npc":
		// Most-recently-attacked NPC view (resolved live from
		// world.npcs by stored index). Null if never attacked
		// OR the NPC has since left view / died. Routines flee/
		// retarget on null.
		idx := c.host.lastAttackedNpcIndex
		if idx == 0 {
			return interp.Null{}, true
		}
		for _, n := range c.host.world.Npcs.All() {
			if n.Index == idx {
				return &npcView{record: n, facts: c.host.facts}, true
			}
		}
		return interp.Null{}, true
	case "last_player":
		// Same shape for the last-attacked player. Common in
		// duels — track who we engaged for fleeing/loot.
		idx := c.host.lastAttackedPlayerIndex
		if idx == 0 {
			return interp.Null{}, true
		}
		if rec, ok := c.host.world.Players.Get(idx); ok {
			return &playerView{record: rec}, true
		}
		return interp.Null{}, true
	}
	return nil, false
}
