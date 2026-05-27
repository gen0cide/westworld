// Package event provides a typed pub/sub bus for game events.
//
// Events are derived from inbound packets by decoders in the runtime
// layer. Subscribers receive events by type. The bus is in-memory,
// goroutine-safe, and lock-free for the happy path.
//
// Two delivery modes:
//
//   - Async (default): published events are buffered and delivered on
//     subscriber goroutines. A slow subscriber doesn't block publishers.
//   - Sync: published events are delivered before Publish returns.
//     Used internally by state-mirror updaters that need to apply
//     changes before downstream consumers see derived state.
//
// Phase 1 vocabulary:
//
//   - ChatReceived: another player said something
//   - PrivateMessage: someone whispered us
//   - SystemMessage: the server told us something
//   - StatUpdate: a single skill level changed
//   - StatsSnapshot: full stats dump (after login)
//   - InventorySnapshot: full inventory after login
//   - InventorySlotUpdate: one inventory slot changed
//   - FatigueUpdate: fatigue value changed
//   - ExperienceGain: a single skill gained XP
//   - WelcomeInfo: last-login info shown after login
//   - Death: we died
//   - LogoutConfirm: server agreed to our logout request
//   - GroundItemAppear / GroundItemDisappear: an item entered/left visibility
//   - NpcDialog: NPC asked us to pick from options
//   - NpcDialogText: NPC said something at us
//
// Phase 2 and beyond will add: damage taken, combat started, attacked by,
// etc.
package event
