package remoteclient

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/render"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// MenuDefs carries the static facts definitions a single target needs to build
// its authentic option list. Exactly one of the def pointers is non-nil for any
// real target (which one depends on the kind); the rest are nil and ignored. A
// nil-def target still produces a valid menu — it just collapses to the
// synthetic verbs (e.g. a scenery with no SceneryDef becomes [Examine], a
// boundary becomes [Examine], an NPC becomes [Examine]).
//
// InvSlot carries the inventory-slot facts (Wielded state) for KindInventoryItem
// so the builder can decide Wield vs Remove and fold a Wield/Wear Command. It is
// the zero value for non-inventory kinds.
type MenuDefs struct {
	Scenery  *facts.SceneryDef
	Boundary *facts.BoundaryDef
	Npc      *facts.NpcDef
	Item     *facts.ItemDef
	InvSlot  world.InvSlot
}

// BuildMenu is the PURE, DETERMINISTIC verb-list builder at the heart of Layer
// 2. Given a target kind and its facts defs, it returns the authentic, ordered
// RSC right-click menu for that target, exactly as the mudclient menu* arrays
// would order it.
//
// It returns FOUR parallel pieces:
//   - opts: the wire []MenuOption (id = index, verb = authentic label) the
//     browser renders. opts[0] is the left-click default.
//   - ids: the internal []OptionID, index-aligned with opts, that the dispatcher
//     indexes by the wire optionId to pick a Host method. The browser never sees
//     this; it is the stable dispatch contract.
//
// Because /pick (which sends opts to the browser) and /act (which re-derives ids
// to dispatch) both call this same pure function with the same defs, the verb
// the user clicks and the packet that fires are guaranteed to agree. The
// ordering must therefore be a deterministic function of (kind, defs) only — it
// reads no live state and has no randomness.
//
// De-duplication and folding rules (per 50-impl-spec §3):
//   - Examine is always the LAST entry, synthetic; a def Command2 that is
//     literally "Examine" (case-insensitive) is dropped so Examine never doubles.
//   - A boundary Command1 of "WalkTo" (a plain wall's no-op) is dropped.
//   - For npc, Attack is only added if the def did not already name it via a
//     Command, so a monster whose Command1 is "Attack" gets one Attack entry.
//   - For inventory, a Command that is itself "Wield"/"Wear" is folded into the
//     single Wield entry (never both).
func BuildMenu(kind TargetKind, d MenuDefs) (opts []MenuOption, ids []OptionID) {
	var verbs []string

	add := func(id OptionID, verb string) {
		ids = append(ids, id)
		verbs = append(verbs, verb)
	}

	switch kind {
	case KindNPC:
		named := map[string]bool{} // lowercased verbs already emitted via a Command
		if d.Npc != nil {
			if c := strings.TrimSpace(d.Npc.Command1); c != "" {
				add(OptCommand1, c)
				named[strings.ToLower(c)] = true
			}
			if c := strings.TrimSpace(d.Npc.Command2); c != "" {
				add(OptCommand2, c)
				named[strings.ToLower(c)] = true
			}
		}
		// Attack as a fallback only when the NPC is attackable AND a Command did
		// not already name it (a monster's Command1 is usually "Attack").
		if d.Npc != nil && d.Npc.Attackable && !named["attack"] {
			add(OptAttack, "Attack")
		}
		add(OptExamine, "Examine")

	case KindPlayer:
		// Fixed standard player menu (no per-def verbs for players). Attack is
		// still offered even outside PvP zones — the server rejects it there.
		add(OptTrade, "Trade with")
		add(OptFollow, "Follow")
		add(OptDuel, "Duel with")
		add(OptAttack, "Attack")
		add(OptExamine, "Examine")

	case KindSelf:
		// Self rarely surfaces its own menu (clicks fall through to terrain);
		// when it does it is just Examine — Walk-here is the terrain target's job.
		add(OptExamine, "Examine")

	case KindGroundItem:
		add(OptPickup, "Pick up")
		add(OptExamine, "Examine")

	case KindScenery:
		if d.Scenery != nil {
			if c := strings.TrimSpace(d.Scenery.Command1); c != "" {
				add(OptCommand1, c)
			}
			// Command2 is often literally "Examine" in the def data — drop it and
			// rely on the synthetic Examine so the menu never shows it twice.
			if c := strings.TrimSpace(d.Scenery.Command2); c != "" && !strings.EqualFold(c, "Examine") {
				add(OptCommand2, c)
			}
		}
		add(OptExamine, "Examine")

	case KindBoundary:
		if d.Boundary != nil {
			// A plain wall's Command1 is "WalkTo" — not a real interaction; drop
			// it so the wall collapses to [Examine].
			if c := strings.TrimSpace(d.Boundary.Command1); c != "" && !strings.EqualFold(c, "WalkTo") {
				add(OptCommand1, c)
			}
			if c := strings.TrimSpace(d.Boundary.Command2); c != "" && !strings.EqualFold(c, "Examine") {
				add(OptCommand2, c)
			}
		}
		add(OptExamine, "Examine")

	case KindTerrain:
		add(OptWalkHere, "Walk here")

	case KindInventoryItem:
		wielded := d.InvSlot.Wielded
		wearable := d.Item != nil && d.Item.IsWearable
		cmd := ""
		if d.Item != nil {
			cmd = strings.TrimSpace(d.Item.Command)
		}
		cmdIsWield := strings.EqualFold(cmd, "Wield") || strings.EqualFold(cmd, "Wear")

		// Default Command (Eat/Drink/Bury) — but NOT when it is itself the wield
		// verb (that folds into the Wield entry below).
		if cmd != "" && !cmdIsWield {
			add(OptCommand, cmd)
		}
		// Wield/Wear for an un-worn wearable. If the def Command WAS the wield
		// verb, prefer that exact label (e.g. "Wear" for armour) over a generic
		// "Wield"; otherwise default to "Wield".
		if wearable && !wielded {
			verb := "Wield"
			if cmdIsWield {
				verb = cmd
			}
			add(OptWield, verb)
		}
		// Remove when the slot is already worn — lets you unequip from the
		// inventory view too.
		if wielded {
			add(OptRemove, "Remove")
		}
		add(OptDrop, "Drop")
		add(OptExamine, "Examine")

	default:
		// Unknown kind: degrade to the universal Walk-here so a click is never a
		// dead end. Should not happen for a well-formed candidate.
		add(OptWalkHere, "Walk here")
	}

	opts = make([]MenuOption, len(verbs))
	for i, v := range verbs {
		opts[i] = MenuOption{ID: i, Verb: v}
	}
	return opts, ids
}

// ExamineHost is the narrow read-only interface BuildCandidates needs to
// pre-resolve examine text + names. It is satisfied by *runtime.Host (every
// method sends NO packet — it reads facts + world mirrors). Keeping it an
// interface lets the menu tests run without a live Host.
type ExamineHost interface {
	ExamineNpc(serverIndex int) runtime.Examination
	ExaminePlayer(serverIndex int) runtime.Examination
	ExamineSelf() runtime.Examination
	ExamineGroundItem(x, y int) runtime.Examination
	ExamineScenery(x, y int) runtime.Examination
	ExamineBoundary(x, y int) runtime.Examination
	ExamineInventorySlot(slot int) runtime.Examination
	ExamineItem(itemID int) runtime.Examination
}

// BuildCandidates maps the depth-ordered render.Pick candidates into the
// per-candidate wire Candidates the browser renders, preserving order
// (candidates[0] is the topmost / nearest thing; the terrain "Walk here"
// candidate is last). It is the one place the plane-LOCAL render coordinate
// space is folded into the ABSOLUTE world coords the wire (and every Host
// action) uses: ref.Y = candidate.Y + plane*world.PlaneHeight, exactly as the
// /walk handler does.
//
// For each candidate it:
//   - builds the wire MenuTarget (collapsing NpcID/DefID/ItemID into ref.ID),
//   - looks up the facts defs and runs BuildMenu for the authentic options,
//   - pre-resolves Label / Examine / Detail via the matching ex.Examine* call
//     (free — no packet) so the browser needs no /examine round-trip, and
//   - computes Dist (Chebyshev self->target) as a UI ordering aid.
//
// ex may be nil (then examine text is empty and Label falls back to a synthetic
// name); f may be nil (then no defs, so menus collapse to their synthetic verbs).
func BuildCandidates(ex ExamineHost, f *facts.Facts, cands []render.PickCandidate, selfX, selfY, plane int) []Candidate {
	out := make([]Candidate, 0, len(cands))
	for _, c := range cands {
		absY := c.Y + plane*world.PlaneHeight

		ref := MenuTarget{
			Kind: kindToWire(c.Kind),
			X:    c.X,
			Y:    absY,
			Slot: -1, // world targets carry no inventory slot
		}

		var d MenuDefs
		var label, examine, detail string

		switch ref.Kind {
		case KindNPC:
			ref.Index = c.Index
			ref.ID = c.NpcID
			if f != nil {
				d.Npc = f.NpcDef(c.NpcID)
			}
			label, examine, detail = resolveExamine(ex, "NPC", func(h ExamineHost) runtime.Examination {
				return h.ExamineNpc(c.Index)
			})
			ref.Name = label

		case KindPlayer:
			ref.Index = c.Index
			label, examine, detail = resolveExamine(ex, "Player", func(h ExamineHost) runtime.Examination {
				return h.ExaminePlayer(c.Index)
			})
			ref.Name = label

		case KindSelf:
			label, examine, detail = resolveExamine(ex, "Me", func(h ExamineHost) runtime.Examination {
				return h.ExamineSelf()
			})

		case KindGroundItem:
			ref.ID = c.ItemID
			if f != nil {
				d.Item = f.ItemDef(c.ItemID)
			}
			label, examine, detail = resolveExamine(ex, defItemName(d.Item, c.ItemID), func(h ExamineHost) runtime.Examination {
				return h.ExamineGroundItem(c.X, absY)
			})
			ref.Name = label

		case KindScenery:
			ref.ID = c.DefID
			ref.Dir = c.Direction
			if f != nil {
				d.Scenery = f.SceneryDef(c.DefID)
			}
			label, examine, detail = resolveExamine(ex, sceneryName(d.Scenery), func(h ExamineHost) runtime.Examination {
				return h.ExamineScenery(c.X, absY)
			})
			ref.Name = label

		case KindBoundary:
			ref.ID = c.DefID
			ref.Dir = c.Direction
			if f != nil {
				d.Boundary = f.BoundaryDef(c.DefID)
			}
			label, examine, detail = resolveExamine(ex, boundaryName(d.Boundary), func(h ExamineHost) runtime.Examination {
				return h.ExamineBoundary(c.X, absY)
			})
			ref.Name = label

		case KindTerrain:
			label = "Walk here"
		}

		opts, _ := BuildMenu(ref.Kind, d)
		out = append(out, Candidate{
			Ref:     ref,
			Kind:    ref.Kind,
			Label:   label,
			Examine: examine,
			Detail:  detail,
			Dist:    chebyshev(selfX, selfY, c.X, c.Y),
			Options: opts,
		})
	}
	return out
}

// InventoryMenu builds the Candidate for one occupied inventory slot, used by
// /state so the inventory right-click menu is built by the SAME BuildMenu as a
// world pick (no server round-trip needed for an inventory menu, and a
// world-picked item and an inventory-panel item of the same type offer
// identical verbs). slot is the 0-based inventory slot; s is its contents; def
// is the item's facts def (may be nil). ex may be nil.
func InventoryMenu(ex ExamineHost, slot int, s world.InvSlot, def *facts.ItemDef) Candidate {
	ref := MenuTarget{
		Kind: KindInventoryItem,
		ID:   s.ItemID,
		Slot: slot,
	}

	d := MenuDefs{Item: def, InvSlot: s}
	opts, _ := BuildMenu(KindInventoryItem, d)

	label := defItemName(def, s.ItemID)
	ref.Name = label
	var examine, detail string
	if ex != nil {
		e := ex.ExamineInventorySlot(slot)
		if e.Name != "" {
			label = e.Name
			ref.Name = label
		}
		examine = e.Description
		detail = e.Detail
	}

	return Candidate{
		Ref:     ref,
		Kind:    KindInventoryItem,
		Label:   label,
		Examine: examine,
		Detail:  detail,
		Dist:    0,
		Options: opts,
	}
}

// resolveExamine runs one ex.Examine* accessor (via the lookup closure) and
// folds the result into (label, examine, detail). It is nil-ExamineHost-safe:
// when ex is nil it returns the fallback label and empty examine text. The
// fallback label is used whenever no examine name is available, so a Candidate
// always has a human-readable Label.
func resolveExamine(ex ExamineHost, fallbackLabel string, lookup func(ExamineHost) runtime.Examination) (label, examine, detail string) {
	if ex == nil {
		return fallbackLabel, "", ""
	}
	e := lookup(ex)
	label = e.Name
	if label == "" {
		label = fallbackLabel
	}
	return label, e.Description, e.Detail
}

func defItemName(def *facts.ItemDef, itemID int) string {
	if def != nil && def.Name != "" {
		return def.Name
	}
	return fmt.Sprintf("Item %d", itemID)
}
func sceneryName(def *facts.SceneryDef) string {
	if def != nil && def.Name != "" {
		return def.Name
	}
	return "Object"
}
func boundaryName(def *facts.BoundaryDef) string {
	if def != nil && def.Name != "" {
		return def.Name
	}
	return "Object"
}

func chebyshev(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}
