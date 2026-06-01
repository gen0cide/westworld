// Package remoteclient is Layer 2 of the browser remote-client stack: the
// bridge between a screen-space pick (render.Pick) or a UI panel slot and a
// concrete runtime.Host action.
//
// It owns three things and only these three things:
//
//  1. The wire types (MenuTarget, MenuOption, Candidate, PickResponse) that
//     round-trip losslessly between the browser and the cradle HTTP layer.
//  2. The pure, deterministic verb-list builder (BuildMenu / BuildCandidates /
//     InventoryMenu) that turns a target + its facts defs into the authentic,
//     ordered RSC right-click menu — with the left-click default at index 0.
//  3. The dispatcher (Dispatch) that maps a {target, optionId} pair to the EXACT
//     runtime.Host method and arguments, honouring RSC's 0-based vs 1-based
//     conventions per the dispatch table.
//
// It is PURE DOMAIN LOGIC: no net/http (that is the cradle handler, Layer 3) and
// no rendering (that is the render package, Layer 1). It sends no packets of its
// own — every action goes through a runtime.Host method (which owns the wire) or
// a runtime.Host Examine* accessor (which reads state and sends nothing). The
// single most important invariant it preserves is that the verb the browser
// shows and the packet the dispatcher fires both index the SAME deterministic
// option list, so /pick and /act never disagree.
//
// See docs/remote-client/20-menu-dispatch.md and 50-impl-spec.md for the full
// design; this package implements the §2/§3/§4 frozen contracts of 50-impl-spec.
package remoteclient

import "github.com/gen0cide/westworld/render"

// TargetKind is the string kind on the wire — the JSON contract the browser
// holds opaquely between /pick and /act. It is deliberately distinct from
// render.TargetKind (an int enum internal to the picker): the two live in
// different packages and serve different layers. kindToWire maps one to the
// other by name.
type TargetKind string

const (
	KindNPC           TargetKind = "npc"
	KindPlayer        TargetKind = "player"
	KindSelf          TargetKind = "self"
	KindGroundItem    TargetKind = "ground_item"
	KindScenery       TargetKind = "scenery"
	KindBoundary      TargetKind = "boundary"
	KindTerrain       TargetKind = "terrain"
	KindInventoryItem TargetKind = "inventory_item"
)

// MenuTarget is the self-describing reference that /pick emits and /act consumes
// verbatim. It is flat, value-typed (no pointers, no live server handles), and
// round-trips losslessly through JSON so the browser can hold it opaquely
// between the two calls. The server is STATELESS between /pick and /act: it
// re-resolves volatile identity (Index, Slot) against the live world mirror at
// dispatch time rather than trusting a stale snapshot.
//
// Coordinate convention: X and Y are ABSOLUTE world coords on the wire (the
// plane offset is already folded into Y by BuildCandidates), because every
// runtime.Host action takes absolute coords. render.PickCandidate carries
// plane-LOCAL Y; the conversion happens exactly once, in BuildCandidates.
//
// Per-kind field population (only the listed fields are read at dispatch; the
// rest are advisory for display / staleness messages):
//
//	npc            -> Index            (+ X,Y,ID=TypeID,Name advisory)
//	player         -> Index, Name      (Name keys the Follow method)
//	self           -> X,Y
//	ground_item    -> X,Y,ID=ItemID    (+ Name advisory)
//	scenery        -> X,Y (+ ID selects the verb list)  (+ Dir,Name advisory)
//	boundary       -> X,Y,Dir          (+ ID,Name advisory)
//	terrain        -> X,Y
//	inventory_item -> Slot (re-validated vs ID=ItemID at dispatch)  (+ Name advisory)
type MenuTarget struct {
	Kind  TargetKind `json:"kind"`
	Index int        `json:"index,omitempty"` // server actor index (npc|player); 0 otherwise
	X     int        `json:"x"`               // ABSOLUTE world X
	Y     int        `json:"y"`               // ABSOLUTE world Y (plane folded in)
	Dir   int        `json:"dir,omitempty"`   // boundary edge dir 0..3 (boundary only)
	ID    int        `json:"id,omitempty"`    // scenery DefID | boundary DefID | ground/inv ItemID | npc TypeID
	Slot  int        `json:"slot"`            // inventory slot 0..29; -1 if N/A (always emitted — slot 0 is real)
	Name  string     `json:"name,omitempty"`  // display name; the Follow key; staleness messages
}

// MenuOption is one verb row in a context menu. ID is the index into THIS
// candidate's Options slice (dispatch-table-relative to the candidate's Kind,
// NOT a global enum), and Verb is the authentic RSC label shown to the user.
// Options[0] is ALWAYS the left-click default — there is no per-option
// isDefault flag; ordering carries the default structurally.
type MenuOption struct {
	ID   int    `json:"id"`
	Verb string `json:"verb"`
}

// Candidate is one entry in /pick's depth-ordered list: a thing under the
// cursor plus its authentic menu. Ref round-trips to /act unchanged. Kind is
// duplicated from Ref.Kind purely for the browser's switch convenience. Label /
// Examine / Detail are pre-resolved from the matching runtime.Host Examine*
// accessor (which sends no packet), so the menu needs no extra round-trip to
// show examine text. Dist is the Chebyshev distance from the host to the
// target, an ordering aid for the UI.
type Candidate struct {
	Ref     MenuTarget   `json:"ref"`
	Kind    TargetKind   `json:"kind"`
	Label   string       `json:"label"`
	Examine string       `json:"examine"`
	Detail  string       `json:"detail"`
	Dist    int          `json:"dist"`
	Options []MenuOption `json:"options"` // >=1; Options[0] is the left-click default
}

// PickResponse is the body /pick returns: the depth-ordered candidate list,
// nearest-camera-first, with the always-present terrain "Walk here" candidate
// last. Never null (an empty slice marshals to []).
type PickResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// OptionID is the stable, label-independent dispatch key used INTERNALLY by
// Layer 2. It is never on the wire — the wire MenuOption.ID is the int index
// into a candidate's Options slice. BuildMenu emits the wire Options alongside a
// parallel []OptionID; the dispatcher re-derives the identical []OptionID from
// the target + defs and indexes it by the wire optionId. Keying dispatch on
// OptionID (not the localised Verb label, which varies per target's def data) is
// what makes /pick and /act agree.
type OptionID string

const (
	OptCommand1 OptionID = "command1" // scenery/boundary/npc primary def verb
	OptCommand2 OptionID = "command2" // scenery/boundary/npc secondary def verb
	OptAttack   OptionID = "attack"
	OptTalkTo   OptionID = "talk_to"
	OptTrade    OptionID = "trade"
	OptFollow   OptionID = "follow"
	OptDuel     OptionID = "duel"
	OptPickup   OptionID = "pickup"
	OptCommand  OptionID = "command" // inventory default verb (Eat/Drink/Bury)
	OptWield    OptionID = "wield"
	OptRemove   OptionID = "remove"
	OptDrop     OptionID = "drop"
	OptUse      OptionID = "use" // M2 placeholder (no-op in M1; verb reserved)
	OptWalkHere OptionID = "walk_here"
	OptExamine  OptionID = "examine"
)

// kindToWire maps the render package's internal int TargetKind enum to this
// package's string wire kind, 1:1 by meaning. render.TargetSelf maps to
// KindSelf; KindInventoryItem has no render counterpart (it is built by the
// handler from world.Inventory, not from a screen pick). An unrecognised kind
// falls back to KindTerrain (the universal, always-safe "Walk here" target).
func kindToWire(k render.TargetKind) TargetKind {
	switch k {
	case render.TargetNPC:
		return KindNPC
	case render.TargetPlayer:
		return KindPlayer
	case render.TargetSelf:
		return KindSelf
	case render.TargetGroundItem:
		return KindGroundItem
	case render.TargetScenery:
		return KindScenery
	case render.TargetBoundary:
		return KindBoundary
	case render.TargetTerrain:
		return KindTerrain
	default:
		return KindTerrain
	}
}
