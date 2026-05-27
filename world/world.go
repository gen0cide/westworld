package world

import "github.com/gen0cide/westworld/event"

// World is the composition of all per-host state mirrors. It's the
// single object higher layers consult to ask "what does my agent
// believe about the world right now?"
type World struct {
	Self      *Self
	Inventory *Inventory
}

// NewWorld returns a freshly-initialized World with all sub-mirrors
// ready (but empty until inbound packets populate them).
func NewWorld() *World {
	return &World{
		Self:      NewSelf(),
		Inventory: NewInventory(),
	}
}

// Apply updates the world state based on an inbound event. Returns
// true if the world was meaningfully changed (caller may use this for
// metrics or change-detection).
//
// Called by the runtime layer for every decoded inbound packet, BEFORE
// the event is published to subscribers. Subscribers therefore see a
// world state already consistent with the event they're being told
// about.
func (w *World) Apply(ev event.Event) bool {
	switch e := ev.(type) {
	case event.StatUpdate:
		w.Self.SetSkill(int(e.Skill), e.Current, e.Max, e.Experience)
		return true
	case event.StatsSnapshot:
		w.Self.SetAllSkills(snapshotSkills(e.Current), snapshotSkills(e.Max), snapshotSkills(e.Experience), e.QuestPoints)
		return true
	case event.ExperienceGain:
		// Only XP changed; preserve current/max.
		s := int(e.Skill)
		w.Self.SetSkill(s, w.Self.SkillLevel(s), w.Self.SkillMax(s), e.XP)
		return true
	case event.FatigueUpdate:
		w.Self.SetFatigue(e.Value)
		return true
	case event.InventorySnapshot:
		slots := make([]InvSlot, len(e.Items))
		for i, it := range e.Items {
			slots[i] = InvSlot{ItemID: it.ItemID, Amount: it.Amount, Wielded: it.Wielded}
		}
		w.Inventory.Replace(slots)
		return true
	case event.InventorySlotUpdate:
		if e.Item == nil {
			w.Inventory.Set(e.Slot, nil)
		} else {
			w.Inventory.Set(e.Slot, &InvSlot{
				ItemID:  e.Item.ItemID,
				Amount:  e.Item.Amount,
				Wielded: e.Item.Wielded,
			})
		}
		return true
	case event.OwnPositionUpdate:
		w.Self.SetPosition(Coord{X: e.X, Y: e.Y})
		return true
	}
	return false
}

// snapshotSkills converts a slice-like [18]int from event types to
// the same shape stored on Self.
func snapshotSkills(in [18]int) [NumSkills]int {
	var out [NumSkills]int
	for i := 0; i < NumSkills && i < len(in); i++ {
		out[i] = in[i]
	}
	return out
}
