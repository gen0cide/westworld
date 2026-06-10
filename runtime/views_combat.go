package runtime

import (
	"github.com/gen0cide/westworld/dsl/interp"
)

// Views for the `combat` faculty. The view-root wiring
// (it.Reserved["combat"] = &combatView{...}) lives centrally in
// dsl_bridge.go.

// ---------- combat ----------

type combatView struct {
	host *Host
	bind *routineBinding
}

func (c *combatView) Kind() string    { return "combat" }
func (c *combatView) Display() string { return "<combat>" }

// Combat perception (#117) reads the engagement / health fields the
// opcode-104 / opcode-234 decoders land on the local player's own
// record (world.Players index 0, world.SelfPlayerIndex) and on the
// NPC / player roster. melee in authentic v235 surfaces only as type-2
// damage on both participants — no projectile — so combat.engaged also
// falls back to our own last-attack intent (host.lastAttacked*Index)
// and to "someone is firing at me".
func (c *combatView) Get(field string) (interp.Value, bool) {
	// Action verbs (attack/set_style/retreat + bang) first. `attack`
	// is also the sanctioned §9 flat alias.
	if v, ok := c.host.namespaceAction(c.bind, "combat", field, combatVerbs); ok {
		return v, true
	}
	switch field {
	case "engaged":
		return interp.Bool(c.engaged()), true
	case "target":
		if t, ok := c.target(); ok {
			return t, true
		}
		return interp.Null{}, true
	case "style":
		// Read-side mirror of the most-recently-applied melee
		// xp-split (#117). Write-through from combat.set_style; RSC
		// sends no echo, so this reflects intent. Defaults to the
		// server start state ("controlled") before any set_style.
		return interp.String(c.host.combatStyle.String()), true
	case "last_npc":
		// Most-recently-attacked NPC view (resolved live from
		// world.npcs by stored index). Null if never attacked
		// OR the NPC has since left view / died. Routines flee/
		// retarget on null.
		idx := int(c.host.lastAttackedNpcIndex.Load())
		if idx == 0 {
			return interp.Null{}, true
		}
		for _, n := range c.host.world.Npcs.All() {
			if n.Index == idx {
				return &npcView{record: n, facts: c.host.facts, host: c.host}, true
			}
		}
		return interp.Null{}, true
	case "last_player":
		// Same shape for the last-attacked player. Common in
		// duels — track who we engaged for fleeing/loot.
		idx := int(c.host.lastAttackedPlayerIndex.Load())
		if idx == 0 {
			return interp.Null{}, true
		}
		if rec, ok := c.host.world.Players.Get(idx); ok {
			return &playerView{record: rec, host: c.host}, true
		}
		return interp.Null{}, true
	}
	return nil, false
}

// target resolves our current engagement target into an Npc or Player
// view, preferring the wire-observed engagement on our own record
// (world.Players index 0) and falling back to our last-attack intent
// (host.lastAttacked*Index) when authentic melee carried no projectile
// to populate the engagement fields. Returns (nil, false) when no
// target can be resolved (caller maps to Null).
func (c *combatView) target() (interp.Value, bool) {
	// 1. Wire-observed engagement on our own slot (ranged/magic, or
	//    any sub-update that set our EngagedNpc/PlayerIndex). Gate on
	//    EngagedAt: a never-engaged record has both indices at the
	//    zero value, which is indistinguishable from "engaging server
	//    index 0" without the timestamp. SetEngagement parks the
	//    unused side at -1, so only the >= 0 side is authoritative.
	if self, ok := c.host.world.Players.Self(); ok && !self.EngagedAt.IsZero() {
		if self.EngagedNpcIndex >= 0 {
			if v, ok := c.resolveNpc(self.EngagedNpcIndex); ok {
				return v, true
			}
		}
		if self.EngagedPlayerIndex >= 0 {
			if v, ok := c.resolvePlayer(self.EngagedPlayerIndex); ok {
				return v, true
			}
		}
	}
	// 2. Melee fallback: the entity we last issued attack() against,
	//    while it is still visible. (Mirrors combat.last_npc/last_player
	//    resolution but only counts as a live target if still in view.)
	if idx := int(c.host.lastAttackedNpcIndex.Load()); idx != 0 {
		if v, ok := c.resolveNpc(idx); ok {
			return v, true
		}
	}
	if idx := int(c.host.lastAttackedPlayerIndex.Load()); idx != 0 {
		if v, ok := c.resolvePlayer(idx); ok {
			return v, true
		}
	}
	return nil, false
}

// engaged reports whether self is in active combat: either we have a
// resolvable target, or another player is firing projectiles at us
// (their EngagedPlayerIndex == SelfPlayerIndex). Authentic melee has
// no projectile signal for "being attacked", so a defender with no
// outgoing attack relies on the target() / last-attack side.
func (c *combatView) engaged() bool {
	if _, ok := c.target(); ok {
		return true
	}
	if _, ok := c.host.world.Players.AttackerOfSelf(); ok {
		return true
	}
	return false
}

// resolveNpc looks up an NPC by server index in the live roster.
func (c *combatView) resolveNpc(index int) (interp.Value, bool) {
	for _, n := range c.host.world.Npcs.All() {
		if n.Index == index {
			// A target observed at 0 hits has died — combat.target
			// clears to null on death (the on target_died/npc_killed
			// edge fires off this same transition). #119.
			if n.HasHits && n.CurHits == 0 {
				return nil, false
			}
			return &npcView{record: n, facts: c.host.facts, host: c.host}, true
		}
	}
	return nil, false
}

// resolvePlayer looks up a player by server index in the live roster.
// Our own index (ourselves) is never a valid combat target.
func (c *combatView) resolvePlayer(index int) (interp.Value, bool) {
	if index == c.host.world.Players.SelfIndex() {
		return nil, false
	}
	if rec, ok := c.host.world.Players.Get(index); ok {
		return &playerView{record: rec, host: c.host}, true
	}
	return nil, false
}
