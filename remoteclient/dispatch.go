package remoteclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// ActionHost is the narrow interface of mutating runtime.Host methods the
// dispatcher routes to. Listing the exact verified method set (rather than
// accepting a concrete *runtime.Host) keeps Layer 2 decoupled from runtime's
// internals AND lets dispatch_test.go assert routing with a mock — the whole
// point of the §4/§7 testability contract. *runtime.Host satisfies this
// interface (every signature verified against runtime/*.go on feat/remote-client).
type ActionHost interface {
	// movement
	Walk(ctx context.Context, x, y int) error
	// npc
	AttackNpc(ctx context.Context, serverIndex int) error
	TalkToNpc(ctx context.Context, serverIndex int) error
	NpcCommand(ctx context.Context, npcServerIndex int) error
	// player
	InitTradeRequest(ctx context.Context, serverIndex int) error
	InitDuelRequest(ctx context.Context, serverIndex int) error
	AttackPlayer(ctx context.Context, serverIndex int) error
	Follow(ctx context.Context, targetName string, startupTimeout time.Duration) error
	// ground items
	PickUpItem(ctx context.Context, x, y, itemID int) error
	// scenery / boundary
	InteractAt(ctx context.Context, x, y, option int) error
	InteractWithBoundary(ctx context.Context, x, y, direction int) error
	// inventory / equipment
	ItemCommand(ctx context.Context, slot int) error
	DropItem(ctx context.Context, slot int) error
	EquipItem(ctx context.Context, slotIndex int) error
	UnequipItem(ctx context.Context, slotIndex int) error
	// M2 use-on (verbs reserved; not wired into M1 menus, but kept on the
	// interface so the dispatch table can grow without an interface change)
	UseItemOnItem(ctx context.Context, slot1, slot2 int) error
	UseItemOnScenery(ctx context.Context, x, y, slot int) error
	UseItemOnBoundary(ctx context.Context, x, y, direction, slot int) error
	UseItemOnGroundItem(ctx context.Context, x, y, groundItemID, slot int) error
	UseItemOnNpc(ctx context.Context, npcServerIndex, slot int) error
	UseItemOnPlayer(ctx context.Context, playerServerIndex, slot int) error
}

// Compile-time proof that the real *runtime.Host satisfies BOTH narrow
// interfaces the dispatcher routes to. The dispatch_test.go mocks only prove the
// MOCKS satisfy the interfaces; these assertions catch the case the audit cares
// about — a signature drift in runtime that would silently stop *runtime.Host
// from being a valid ActionHost/ExamineHost (which the handler relies on, §6.3/§7).
var (
	_ ActionHost  = (*runtime.Host)(nil)
	_ ExamineHost = (*runtime.Host)(nil)
)

// WorldView is the OPTIONAL read-only mirror the dispatcher uses to re-validate
// volatile identity before firing a packet — an NPC/player must still be visible
// at the given Index, an inventory slot must still hold the expected ItemID. It
// is satisfied by *world.World. It is optional: a nil WorldView skips
// re-validation (so the dispatch routing can be unit-tested with only a mock
// ActionHost, per §7), which is exactly what dispatch_test.go relies on.
type WorldView interface {
	NpcVisible(serverIndex int) bool
	PlayerVisible(serverIndex int) bool
	// SlotItem returns the ItemID and Wielded state of inventory slot (0-based)
	// and whether the slot exists. The ItemID is used to reject (or re-resolve) a
	// stale inventory action when the inventory shifted between /pick and /act
	// (RemoveSlot compacts holes); the Wielded state is needed so the dispatcher
	// re-derives the SAME option list /state showed (Wield for an un-worn item vs
	// Remove for a worn one), keeping optionId aligned.
	SlotItem(slot int) (itemID int, wielded bool, ok bool)
}

// Lane tells the HTTP handler which execution lane an action belongs to, so it
// routes the call to the right worker without re-deriving the option. RSC
// actions override each other, so most go through ONE serialized worker
// (LaneAction); walk uses a coalescing retarget lane; Follow blocks and needs
// its own cancellable goroutine; examine sends no packet and is resolved
// synchronously.
type Lane int

const (
	LaneAction Lane = iota // serialized single-action worker (the default)
	LaneWalk               // coalescing walk-retarget lane (fire-and-forget)
	LaneFollow             // own cancellable goroutine (Follow blocks until ctx cancel)
	LaneSync               // resolved synchronously, no worker, no packet (examine)
)

func (l Lane) String() string {
	switch l {
	case LaneAction:
		return "action"
	case LaneWalk:
		return "walk"
	case LaneFollow:
		return "follow"
	case LaneSync:
		return "sync"
	default:
		return "unknown"
	}
}

// ErrStaleTarget is returned (wrapped) when re-validation against the live world
// fails: the actor moved out of view, or the inventory slot no longer holds the
// expected item. The handler maps this to {ok:false} with the message — NOT an
// HTTP error — so the browser can simply re-pick rather than treat it as a
// server fault.
var ErrStaleTarget = errors.New("stale target")

// ErrUnknownOption is returned when optionId is out of range for the target's
// re-derived option list (a malformed or out-of-date /act request).
var ErrUnknownOption = errors.New("unknown option")

// followStartupTimeout bounds how long Follow waits for the target to first
// become visible before giving up (it then loops until ctx cancel). Mirrors a
// reasonable interactive default; the handler's own ctx still governs overall
// lifetime.
const followStartupTimeout = 10 * time.Second

// Dispatcher maps a {MenuTarget, optionId} pair to the exact runtime.Host call.
// It is stateless between /pick and /act: Dispatch re-derives the option list
// for the target's kind (from facts via target.ID) with the SAME BuildMenu that
// /pick used, then indexes it by optionId to recover the stable OptionID, then
// routes per the §4 table. host is required; ex (examine reads) and facts and
// world (re-validation) are optional but recommended.
type Dispatcher struct {
	host  ActionHost
	ex    ExamineHost
	facts *facts.Facts
	world WorldView
}

// NewDispatcher builds a Dispatcher. host is required (it owns the wire). ex
// resolves examine text for the sync lane (may be nil — then examine returns a
// minimal message). f supplies the defs to re-derive the option list (may be nil
// — then menus collapse to synthetic verbs, identical to BuildCandidates with a
// nil facts). w re-validates volatile identity (may be nil — then validation is
// skipped, which is what the unit tests use).
func NewDispatcher(host ActionHost, ex ExamineHost, f *facts.Facts, w WorldView) *Dispatcher {
	return &Dispatcher{host: host, ex: ex, facts: f, world: w}
}

// Dispatch resolves and fires (or routes) one action. It returns:
//   - message: a human-readable result line for the browser (e.g. the examine
//     text, or "Attack" for a queued interaction). May be empty.
//   - lane: which execution lane the handler should run this on (LaneWalk /
//     LaneFollow are routed by the handler; LaneAction is the serialized worker;
//     LaneSync was already fully resolved here and needs no worker).
//   - err: ErrStaleTarget (wrapped) if re-validation failed; ErrUnknownOption if
//     optionId is out of range; or a Host method error. The handler maps
//     ErrStaleTarget/ErrUnknownOption to {ok:false}, not an HTTP error.
//
// For LaneSync (examine) the action is fully performed here (it sends no packet)
// and message holds the examine text. For LaneAction the Host method is invoked
// synchronously with ctx — the CALLER (handler) is responsible for having put
// the call on the serialized worker by passing a ctx scoped to that worker; the
// dispatcher does not own the worker. For LaneWalk / LaneFollow the dispatcher
// still performs the call (Walk / Follow) with the given ctx, but signals the
// lane so the handler can manage coalescing / cancellation around it.
func (d *Dispatcher) Dispatch(ctx context.Context, t MenuTarget, optionID int) (message string, lane Lane, err error) {
	// Re-derive the SAME option list /pick produced, from the target's kind +
	// its defs (looked up via t.ID). This is what makes /pick and /act agree.
	defs := d.defsFor(t)
	_, ids := BuildMenu(t.Kind, defs)
	if optionID < 0 || optionID >= len(ids) {
		return "", LaneAction, fmt.Errorf("%w: option %d of %d for kind %q", ErrUnknownOption, optionID, len(ids), t.Kind)
	}
	opt := ids[optionID]

	switch t.Kind {
	case KindNPC:
		return d.dispatchNPC(ctx, t, opt, defs.Npc)
	case KindPlayer:
		return d.dispatchPlayer(ctx, t, opt)
	case KindSelf:
		return d.dispatchSelf(t, opt)
	case KindGroundItem:
		return d.dispatchGroundItem(ctx, t, opt)
	case KindScenery:
		return d.dispatchScenery(ctx, t, opt, defs.Scenery)
	case KindBoundary:
		return d.dispatchBoundary(ctx, t, opt, defs.Boundary)
	case KindTerrain:
		return d.dispatchTerrain(ctx, t, opt)
	case KindInventoryItem:
		return d.dispatchInventory(ctx, t, opt, defs.Item)
	default:
		return "", LaneAction, fmt.Errorf("dispatch: unknown kind %q", t.Kind)
	}
}

// ResolveLane is the PURE, SIDE-EFFECT-FREE counterpart of Dispatch: given the
// same {MenuTarget, optionId} pair, it re-derives which execution Lane the action
// belongs to and validates it — WITHOUT firing any Host call. The HTTP handler
// calls this FIRST to route the request to the correct worker lane, then calls
// Dispatch EXACTLY ONCE inside that lane. This separation fixes the
// double-dispatch hazard: Dispatch executes the Host method as it resolves, so
// using it for routing (and again inside the worker) would fire every packet
// twice — the first on an unscoped ctx outside the serialized worker.
//
// It performs the identical validation Dispatch's per-kind helpers do, in the
// same order, so a request that ResolveLane accepts is one Dispatch will also
// accept (and vice-versa):
//   - re-derives the option list via BuildMenu (the same pure function /pick used)
//     and rejects an out-of-range optionId with ErrUnknownOption;
//   - re-validates volatile identity (requireNpcVisible / requirePlayerVisible /
//     requireSlotItem) and rejects a vanished actor or shifted slot with
//     ErrStaleTarget — so the handler emits {ok:false} without ever touching the
//     wire;
//   - maps the resolved OptionID -> Lane (OptExamine / OptUse -> LaneSync,
//     OptWalkHere -> LaneWalk, OptFollow -> LaneFollow, everything else ->
//     LaneAction).
//
// It reads only static defs + the read-only WorldView; it NEVER calls an
// ActionHost or ExamineHost method, so a mock ActionHost that fails the test on
// any invocation proves ResolveLane is pure (see dispatch_test.go).
func (d *Dispatcher) ResolveLane(t MenuTarget, optionID int) (Lane, error) {
	defs := d.defsFor(t)
	_, ids := BuildMenu(t.Kind, defs)
	if optionID < 0 || optionID >= len(ids) {
		return LaneAction, fmt.Errorf("%w: option %d of %d for kind %q", ErrUnknownOption, optionID, len(ids), t.Kind)
	}
	opt := ids[optionID]

	// Re-validate volatile identity for the kinds Dispatch validates, BEFORE
	// committing a lane, so a stale target is rejected here (the handler maps it
	// to {ok:false}) rather than firing a doomed packet inside the worker.
	switch t.Kind {
	case KindNPC:
		if err := d.requireNpcVisible(t); err != nil {
			return LaneAction, err
		}
	case KindPlayer:
		if err := d.requirePlayerVisible(t); err != nil {
			return LaneAction, err
		}
		// Follow additionally needs a name to key the Host method; reject early so
		// the handler does not start a useless follow goroutine.
		if opt == OptFollow && t.Name == "" {
			return LaneFollow, fmt.Errorf("%w: follow needs a player name", ErrStaleTarget)
		}
	case KindInventoryItem:
		if err := d.requireSlotItem(t); err != nil {
			return LaneAction, err
		}
	}

	switch opt {
	case OptExamine, OptUse:
		// OptExamine sends no packet; OptUse is the reserved M1 no-op. Both are
		// fully handled by a single synchronous Dispatch call (no worker).
		return LaneSync, nil
	case OptWalkHere:
		return LaneWalk, nil
	case OptFollow:
		return LaneFollow, nil
	default:
		return LaneAction, nil
	}
}

// defsFor looks up the facts defs a target's option list depends on, keyed by
// the polymorphic t.ID. Mirrors BuildCandidates' per-kind def selection so the
// re-derived option list is byte-identical to the one /pick built.
func (d *Dispatcher) defsFor(t MenuTarget) MenuDefs {
	var defs MenuDefs
	if d.facts == nil {
		return defs
	}
	switch t.Kind {
	case KindNPC:
		defs.Npc = d.facts.NpcDef(t.ID)
	case KindScenery:
		defs.Scenery = d.facts.SceneryDef(t.ID)
	case KindBoundary:
		defs.Boundary = d.facts.BoundaryDef(t.ID)
	case KindGroundItem, KindInventoryItem:
		defs.Item = d.facts.ItemDef(t.ID)
	}
	// For inventory we also need the current Wielded state so the re-derived
	// list (Wield for an un-worn item vs Remove for a worn one) matches exactly
	// what /state showed — otherwise optionId would index a differently shaped
	// list and mis-route. Re-read it from the live world; when no world is wired
	// (unit tests), Wielded defaults to false and the list is still deterministic.
	if t.Kind == KindInventoryItem && d.world != nil {
		if id, wielded, ok := d.world.SlotItem(t.Slot); ok {
			defs.InvSlot = world.InvSlot{ItemID: id, Wielded: wielded}
		}
	}
	return defs
}

func (d *Dispatcher) dispatchNPC(ctx context.Context, t MenuTarget, opt OptionID, def *facts.NpcDef) (string, Lane, error) {
	if err := d.requireNpcVisible(t); err != nil {
		return "", LaneAction, err
	}
	switch opt {
	case OptCommand1:
		return d.npcVerb(ctx, t, verbOf(def, 1))
	case OptCommand2:
		return d.npcVerb(ctx, t, verbOf(def, 2))
	case OptAttack:
		return "Attack", LaneAction, d.host.AttackNpc(ctx, t.Index)
	case OptTalkTo:
		return "Talk-to", LaneAction, d.host.TalkToNpc(ctx, t.Index)
	case OptExamine:
		return d.examine(KindNPC, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch npc: unhandled option %q", opt)
}

// npcVerb routes an NPC command verb to the right wire primitive: the three the
// protocol distinguishes are Attack (AttackNpc), Talk-to (TalkToNpc), and the
// generic "primary command" (NpcCommand — opcode NPC_COMMAND, used for
// Pickpocket and any other def command). Any verb that is not Attack/Talk-to
// routes to NpcCommand.
func (d *Dispatcher) npcVerb(ctx context.Context, t MenuTarget, verb string) (string, Lane, error) {
	switch {
	case equalVerb(verb, "Attack"):
		return verb, LaneAction, d.host.AttackNpc(ctx, t.Index)
	case equalVerb(verb, "Talk-to"):
		return verb, LaneAction, d.host.TalkToNpc(ctx, t.Index)
	default:
		return verb, LaneAction, d.host.NpcCommand(ctx, t.Index)
	}
}

func (d *Dispatcher) dispatchPlayer(ctx context.Context, t MenuTarget, opt OptionID) (string, Lane, error) {
	if err := d.requirePlayerVisible(t); err != nil {
		return "", LaneAction, err
	}
	switch opt {
	case OptTrade:
		return "Trade with", LaneAction, d.host.InitTradeRequest(ctx, t.Index)
	case OptDuel:
		return "Duel with", LaneAction, d.host.InitDuelRequest(ctx, t.Index)
	case OptAttack:
		return "Attack", LaneAction, d.host.AttackPlayer(ctx, t.Index)
	case OptFollow:
		// Follow blocks until ctx cancel — it MUST run on its own cancellable
		// goroutine, never the shared serialized worker (it would wedge the
		// queue). Signal LaneFollow so the handler runs it there; we still issue
		// the call with the (follow-scoped) ctx the handler provides. It is keyed
		// by name, the one name-keyed Host method.
		if t.Name == "" {
			return "", LaneFollow, fmt.Errorf("%w: follow needs a player name", ErrStaleTarget)
		}
		return "Follow", LaneFollow, d.host.Follow(ctx, t.Name, followStartupTimeout)
	case OptExamine:
		return d.examine(KindPlayer, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch player: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchSelf(t MenuTarget, opt OptionID) (string, Lane, error) {
	switch opt {
	case OptExamine:
		return d.examine(KindSelf, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch self: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchGroundItem(ctx context.Context, t MenuTarget, opt OptionID) (string, Lane, error) {
	switch opt {
	case OptPickup:
		return "Pick up", LaneAction, d.host.PickUpItem(ctx, t.X, t.Y, t.ID)
	case OptExamine:
		return d.examine(KindGroundItem, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch ground_item: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchScenery(ctx context.Context, t MenuTarget, opt OptionID, def *facts.SceneryDef) (string, Lane, error) {
	switch opt {
	case OptCommand1:
		// InteractAt's option arg is 1-BASED: 1 = primary (the single 1-based
		// ordinal in the whole table). Report the authentic def verb (Mine/Chop/…).
		verb := "Command1"
		if def != nil && strings.TrimSpace(def.Command1) != "" {
			verb = strings.TrimSpace(def.Command1)
		}
		return verb, LaneAction, d.host.InteractAt(ctx, t.X, t.Y, 1)
	case OptCommand2:
		verb := "Command2"
		if def != nil && strings.TrimSpace(def.Command2) != "" {
			verb = strings.TrimSpace(def.Command2)
		}
		return verb, LaneAction, d.host.InteractAt(ctx, t.X, t.Y, 2)
	case OptExamine:
		return d.examine(KindScenery, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch scenery: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchBoundary(ctx context.Context, t MenuTarget, opt OptionID, def *facts.BoundaryDef) (string, Lane, error) {
	switch opt {
	case OptCommand1, OptCommand2:
		// Boundary interaction carries NO option ordinal on the wire — there is one
		// InteractWithBoundary(x,y,direction) packet; the server toggles open/close.
		// Both command labels map to the same call; t.Dir is passed through unchanged.
		// Report the authentic def verb (Open/Close) so the browser echoes what the
		// user clicked, even though command1 and command2 fire the identical packet.
		verb := "Use"
		if def != nil {
			if opt == OptCommand1 && strings.TrimSpace(def.Command1) != "" {
				verb = strings.TrimSpace(def.Command1)
			} else if opt == OptCommand2 && strings.TrimSpace(def.Command2) != "" {
				verb = strings.TrimSpace(def.Command2)
			}
		}
		return verb, LaneAction, d.host.InteractWithBoundary(ctx, t.X, t.Y, t.Dir)
	case OptExamine:
		return d.examine(KindBoundary, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch boundary: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchTerrain(ctx context.Context, t MenuTarget, opt OptionID) (string, Lane, error) {
	switch opt {
	case OptWalkHere:
		// Walk goes through the coalescing retarget lane (spam-click safe).
		return "Walk here", LaneWalk, d.host.Walk(ctx, t.X, t.Y)
	}
	return "", LaneAction, fmt.Errorf("dispatch terrain: unhandled option %q", opt)
}

func (d *Dispatcher) dispatchInventory(ctx context.Context, t MenuTarget, opt OptionID, def *facts.ItemDef) (string, Lane, error) {
	// Re-validate the slot against the live inventory: RemoveSlot compacts holes,
	// so a slot may have shifted to a different item since the pick. If the slot
	// no longer holds the expected ItemID, reject as stale rather than firing a
	// mis-aimed packet at whatever is in the slot now.
	if err := d.requireSlotItem(t); err != nil {
		return "", LaneAction, err
	}
	switch opt {
	case OptCommand:
		// Slot is 0-based end to end — no conversion. Report the authentic def
		// verb (Eat/Drink/Bury) when known so the browser echoes the real action.
		verb := "Use"
		if def != nil && strings.TrimSpace(def.Command) != "" {
			verb = strings.TrimSpace(def.Command)
		}
		return verb, LaneAction, d.host.ItemCommand(ctx, t.Slot)
	case OptWield:
		return "Wield", LaneAction, d.host.EquipItem(ctx, t.Slot)
	case OptRemove:
		return "Remove", LaneAction, d.host.UnequipItem(ctx, t.Slot)
	case OptDrop:
		return "Drop", LaneAction, d.host.DropItem(ctx, t.Slot)
	case OptUse:
		// M2 placeholder: use-on needs a second (dragged) target. No packet in M1.
		return "Use-on is M2", LaneSync, nil
	case OptExamine:
		return d.examine(KindInventoryItem, t), LaneSync, nil
	}
	return "", LaneAction, fmt.Errorf("dispatch inventory_item: unhandled option %q", opt)
}

// --- re-validation against the live world (no-op when world is nil) ---

func (d *Dispatcher) requireNpcVisible(t MenuTarget) error {
	if d.world == nil {
		return nil
	}
	if !d.world.NpcVisible(t.Index) {
		return fmt.Errorf("%w: npc %d (%s) no longer visible", ErrStaleTarget, t.Index, t.Name)
	}
	return nil
}

func (d *Dispatcher) requirePlayerVisible(t MenuTarget) error {
	if d.world == nil {
		return nil
	}
	if !d.world.PlayerVisible(t.Index) {
		return fmt.Errorf("%w: player %d (%s) no longer visible", ErrStaleTarget, t.Index, t.Name)
	}
	return nil
}

func (d *Dispatcher) requireSlotItem(t MenuTarget) error {
	if d.world == nil {
		return nil
	}
	id, _, ok := d.world.SlotItem(t.Slot)
	if !ok {
		return fmt.Errorf("%w: inventory slot %d empty", ErrStaleTarget, t.Slot)
	}
	if t.ID != 0 && id != t.ID {
		return fmt.Errorf("%w: inventory slot %d now holds item %d, expected %d", ErrStaleTarget, t.Slot, id, t.ID)
	}
	return nil
}

// examine resolves the examine text for a target via the matching ExamineHost
// accessor (no packet). Returns the Examination's String() form, or a minimal
// fallback when no ExamineHost is wired.
func (d *Dispatcher) examine(kind TargetKind, t MenuTarget) string {
	if d.ex == nil {
		if t.Name != "" {
			return t.Name
		}
		return string(kind)
	}
	var e runtime.Examination
	switch kind {
	case KindNPC:
		e = d.ex.ExamineNpc(t.Index)
	case KindPlayer:
		e = d.ex.ExaminePlayer(t.Index)
	case KindSelf:
		e = d.ex.ExamineSelf()
	case KindGroundItem:
		e = d.ex.ExamineGroundItem(t.X, t.Y)
	case KindScenery:
		e = d.ex.ExamineScenery(t.X, t.Y)
	case KindBoundary:
		e = d.ex.ExamineBoundary(t.X, t.Y)
	case KindInventoryItem:
		e = d.ex.ExamineInventorySlot(t.Slot)
	}
	return e.String()
}

// verbOf returns NpcDef.Command1 (n==1) or Command2 (n==2), or "" when the def
// is nil. Used by the NPC command verb-router.
func verbOf(def *facts.NpcDef, n int) string {
	if def == nil {
		return ""
	}
	if n == 1 {
		return def.Command1
	}
	return def.Command2
}

// equalVerb compares menu verbs case-insensitively after trimming surrounding
// space, so "Talk-to", "talk-to", and " Talk-to " all match.
func equalVerb(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
