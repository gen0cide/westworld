package v235

import (
	"fmt"

	"github.com/gen0cide/westworld/event"
)

// DecodeInbound dispatches a decoded frame to its packet-specific
// parser, returning a typed event. If the opcode is not yet decoded,
// returns an UnknownPacket event with the opcode and payload size.
//
// Parser implementations live below, one per opcode. Each cites the
// source it was derived from.
// DecodeInbound decodes one inbound frame. The IsStackable callback
// is consulted only by the inventory decoder; passing nil treats
// every item as unstackable (safe default — over-reads turn into
// later-slot corruption, the same bug the typed callback fixes).
//
// We keep this in proto/v235 without importing facts to avoid a
// cycle; the runtime callsite passes the lookup closure.
func DecodeInbound(f Frame, isStackable func(itemID int) bool) (event.Event, error) {
	switch f.Opcode {
	case InServerMessage:
		return decodeServerMessage(f.Payload)
	case InPrivateMessage:
		return decodePrivateMessage(f.Payload)
	case InStat:
		return decodeStat(f.Payload)
	case InStats:
		return decodeStats(f.Payload)
	case InExperience:
		return decodeExperience(f.Payload)
	case InFatigue:
		return decodeFatigue(f.Payload)
	case InInventory:
		return decodeInventory(f.Payload, isStackable)
	case InInventorySlotUpdate:
		return decodeInventorySlotUpdate(f.Payload)
	case InInventoryRemoveItem:
		// Server sends [byte slot] to signal "this slot is now empty".
		// Synthesize an InventorySlotUpdate with nil item so the
		// world mirror clears it.
		b := WrapBuffer(f.Payload)
		slot, _ := b.ReadByte()
		return event.InventorySlotUpdate{Slot: int(slot), Item: nil}, nil
	case InWelcomeInfo:
		return decodeWelcomeInfo(f.Payload)
	case InSendLogout:
		return event.LogoutConfirm{}, nil
	case InDeath:
		return event.Death{}, nil
	case InGroundItemHandler:
		return decodeGroundItem(f.Payload)
	case InNpcDialogText:
		return decodeNpcDialogText(f.Payload)
	case InNpcDialogOptions:
		return decodeNpcDialogOptions(f.Payload)
	case InSleepScreen:
		return event.SleepScreenAppeared{}, nil
	case InTradeWindow:
		// Server is asking us to open the trade window because another
		// player has just requested a trade with us. Payload: [short
		// serverIndex of the requester].
		b := WrapBuffer(f.Payload)
		idx, err := b.ReadUint16()
		if err != nil {
			return event.UnknownPacket{Opcode: f.Opcode, PayloadSize: len(f.Payload)}, nil
		}
		return event.TradeRequestReceived{FromPlayerIndex: int(idx)}, nil
	case InTradeClose:
		// Server closes the trade window without an explicit
		// "completed" signal — we can't tell decline-vs-success
		// from this packet alone. Routines branching on outcome
		// should diff inventory before/after (item_gained event)
		// to detect a successful exchange. Mark completed=false
		// here so the default state is "cancelled until proven
		// otherwise"; downstream handlers can flip if needed.
		return event.TradeClosed{Completed: false}, nil
	case InTradeOtherAccepted:
		return event.TradeOtherAccepted{}, nil
	case InTradeOpenConfirm:
		// Server is opening the trade window because both sides
		// accepted the initial request. Payload (from
		// ActionSender.sendTradeAcceptConfirm): [short otherPlayerIndex]
		b := WrapBuffer(f.Payload)
		idx, _ := b.ReadUint16()
		return event.TradeOpened{OtherPlayerIndex: int(idx)}, nil
	case InTradeOtherItems:
		// Opponent updated their trade offer. Payload:
		//   [byte count]
		//   for each: [short catalogID] [int amount]
		return decodeTradeOtherItems(f.Payload)
	case InSendPlayerCoords:
		// PlayerCoords is special: it produces one own-position event
		// AND zero or more nearby-player events. The single-return
		// shape of DecodeInbound only emits the own event here; the
		// runtime layer should call DecodePlayerCoords directly when
		// it wants the nearby list too.
		own, _, err := DecodePlayerCoords(f.Payload)
		if err != nil {
			return event.UnknownPacket{Opcode: f.Opcode, PayloadSize: len(f.Payload)}, nil
		}
		return own, nil
	case InSendUpdatePlayers:
		// UpdatePlayers is multi-record: typically multiple events
		// from one packet. DecodeInbound's single-event return is
		// inadequate; runtime calls DecodeUpdatePlayers directly to
		// get the full list. Here we return the first event (or nil).
		events, err := DecodeUpdatePlayers(f.Payload)
		if err != nil || len(events) == 0 {
			return event.UnknownPacket{Opcode: f.Opcode, PayloadSize: len(f.Payload)}, nil
		}
		return events[0], nil
	default:
		return event.UnknownPacket{Opcode: f.Opcode, PayloadSize: len(f.Payload)}, nil
	}
}

// decodeServerMessage parses opcode 131. Per ActionSender.java:1241-1262.
//
//	[byte] message_type
//	[byte] info_contained (bit 0: has sender, bit 1: has color)
//	[zero-quoted string] message_text
//	[if has_sender: zero-quoted string] sender_name
//	[if has_sender: zero-quoted string] sender_name (duplicate)
//	[if has_color:  zero-quoted string] color_string
//
// If sender is present, this is a public chat message from that
// player (we emit a ChatReceived). If not, it's a server message
// (we emit a SystemMessage).
func decodeServerMessage(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	msgType, err := b.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("server_message type: %w", err)
	}
	info, err := b.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("server_message info: %w", err)
	}
	text, err := b.ReadZeroQuotedString()
	if err != nil {
		return nil, fmt.Errorf("server_message text: %w", err)
	}
	hasSender := info&0x01 != 0
	hasColor := info&0x02 != 0
	var sender, color string
	if hasSender {
		sender, err = b.ReadZeroQuotedString()
		if err != nil {
			return nil, fmt.Errorf("server_message sender: %w", err)
		}
		// Duplicate sender — read and discard.
		if _, err := b.ReadZeroQuotedString(); err != nil {
			return nil, fmt.Errorf("server_message sender_dup: %w", err)
		}
	}
	if hasColor {
		color, err = b.ReadZeroQuotedString()
		if err != nil {
			return nil, fmt.Errorf("server_message color: %w", err)
		}
	}
	if sender != "" {
		return event.NewChatReceived(event.MessageType(msgType), sender, text, color), nil
	}
	return event.SystemMessage{Type: event.MessageType(msgType), Message: text, Color: color}, nil
}

// decodePrivateMessage parses opcode 120. Per ActionSender.java:1291-1319.
//
//	[zero-quoted string] sender_name
//	[zero-quoted string] former_name
//	[byte]   icon_sprite
//	[byte]   padding
//	[byte]   padding
//	[byte]   padding
//	[short]  world_number
//	[3 bytes] message counter (24-bit)
//	[smart08_16] message_char_count
//	[RSCString] compressed message body
func decodePrivateMessage(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	sender, err := b.ReadZeroQuotedString()
	if err != nil {
		return nil, fmt.Errorf("private_message sender: %w", err)
	}
	former, _ := b.ReadZeroQuotedString()
	icon, _ := b.ReadByte()
	_, _ = b.ReadByte()
	_, _ = b.ReadByte()
	_, _ = b.ReadByte()
	world, _ := b.ReadUint16()
	// 3 bytes counter — consume
	_, _ = b.ReadByte()
	_, _ = b.ReadByte()
	_, _ = b.ReadByte()
	// Body: smart08_16 char count, then RSC-compressed bytes.
	charCount, _ := b.ReadSmart08_16()
	rest := b.RemainingBytes()
	body := DecipherRSCString(rest, charCount)
	return event.PrivateMessage{
		Sender:     sender,
		FormerName: former,
		IconSprite: int(icon),
		World:      int(world),
		Message:    body,
	}, nil
}

// decodeStat parses opcode 159. Per ActionSender.java:1357-1363.
//
//	[byte] skill_id
//	[byte] current_level
//	[byte] max_level
//	[int]  experience
func decodeStat(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	skill, _ := b.ReadByte()
	cur, _ := b.ReadByte()
	mx, _ := b.ReadByte()
	xp, err := b.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("stat xp: %w", err)
	}
	return event.StatUpdate{
		Skill:      event.SkillID(skill),
		Current:    int(cur),
		Max:        int(mx),
		Experience: int(xp),
	}, nil
}

// decodeStats parses opcode 156 (full stats dump). Per ActionSender.java:1382-1427.
//
//	[18 bytes] current levels
//	[18 bytes] max levels
//	[18 ints]  experience values
//	[byte]     quest points
func decodeStats(payload []byte) (event.Event, error) {
	if len(payload) < 18*2+18*4+1 {
		return nil, fmt.Errorf("stats: short payload %d", len(payload))
	}
	var out event.StatsSnapshot
	b := WrapBuffer(payload)
	for i := 0; i < 18; i++ {
		v, _ := b.ReadByte()
		out.Current[i] = int(v)
	}
	for i := 0; i < 18; i++ {
		v, _ := b.ReadByte()
		out.Max[i] = int(v)
	}
	for i := 0; i < 18; i++ {
		v, _ := b.ReadUint32()
		out.Experience[i] = int(v)
	}
	qp, _ := b.ReadByte()
	out.QuestPoints = int(qp)
	return out, nil
}

// decodeExperience parses opcode 33. Per ActionSender.java:1366-1371.
//
//	[byte] skill_id
//	[int]  experience
func decodeExperience(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	skill, _ := b.ReadByte()
	xp, _ := b.ReadUint32()
	return event.ExperienceGain{Skill: event.SkillID(skill), XP: int(xp)}, nil
}

// decodeFatigue parses opcode 114. Per ActionSender.java:462-466.
//
//	[short] fatigue_value_scaled (0..750 roughly)
func decodeFatigue(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	v, _ := b.ReadUint16()
	return event.FatigueUpdate{Value: int(v)}, nil
}

// decodeInventory parses opcode 53 (full inventory). Per ActionSender.java:967-991.
//
//	[byte] size
//	[for each:
//	    [short] item_id_with_wield_flag (high bit = wielded)
//	    [if amount > 0: smart u16/u32] amount
//	]
//
// The "if amount > 0" semantics depend on whether the item is
// stackable; for Phase 1 we always read an amount short and treat
// 0xFFFF as a sentinel.
//
// NOTE: The exact size encoding may need adjustment when we
// integration-test against real inventory dumps. Phase 1 captures the
// item IDs; amounts may be approximate.
func decodeInventory(payload []byte, isStackable func(itemID int) bool) (event.Event, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("inventory: empty payload")
	}
	b := WrapBuffer(payload)
	size, _ := b.ReadByte()
	items := make([]event.InventoryItem, 0, size)
	for i := 0; i < int(size); i++ {
		slotStart := b.rpos
		if b.Len() < 2 {
			break
		}
		raw, _ := b.ReadUint16()
		wielded := raw&0x8000 != 0
		itemID := int(raw & 0x7FFF)
		amount := 1
		// Per Payload235Generator.SEND_INVENTORY (ActionSender.java:967):
		// the server writes the amount field ONLY when the item is
		// stackable or noted. Unstackable items emit just the 2-byte
		// id+wielded short with no amount field.
		//
		// Decoder needs the same stackability rule to know whether to
		// consume amount bytes. Reading amount unconditionally (the
		// previous bug) realigned every subsequent slot against the
		// next slot's id bytes — making slot[1]+ corrupt and
		// breaking use(item, target) since the wrong slot was passed.
		if isStackable != nil && isStackable(itemID) {
			amount = b.readUnsignedShortIntSmart()
		}
		// Anomaly checks — if either field is wildly out of range
		// we likely have a decoder offset bug. WARN with the bytes
		// that produced the values so the next debug session can
		// trace it instantly. Cheap; happens once per slot per
		// inventory snapshot (rare).
		flagAnomaly("inventory.item_id[slot]", itemID, plausibleMaxItemID, payload[slotStart:b.rpos])
		flagAnomaly("inventory.amount[slot]", amount, plausibleMaxAmount, payload[slotStart:b.rpos])
		items = append(items, event.InventoryItem{ItemID: itemID, Amount: amount, Wielded: wielded})
	}
	return event.InventorySnapshot{Items: items}, nil
}

// decodeInventorySlotUpdate parses opcode 90. Per ActionSender.java:1825-1846.
//
//	[byte]   slot_index
//	[short]  item_id_with_wield_flag (0 = slot cleared)
//	[short/int] amount (smart-encoded)
func decodeInventorySlotUpdate(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	slot, _ := b.ReadByte()
	raw, _ := b.ReadUint16()
	if raw == 0 {
		return event.InventorySlotUpdate{Slot: int(slot), Item: nil}, nil
	}
	wielded := raw&0x8000 != 0
	itemID := int(raw & 0x7FFF)
	amount := 1
	if b.Len() >= 4 {
		a, _ := b.ReadUint32()
		amount = int(a)
	} else if b.Len() >= 2 {
		a, _ := b.ReadUint16()
		amount = int(a)
	}
	return event.InventorySlotUpdate{
		Slot: int(slot),
		Item: &event.InventoryItem{ItemID: itemID, Amount: amount, Wielded: wielded},
	}, nil
}

// decodeWelcomeInfo parses opcode 182. Per ActionSender.java:1111-1120.
//
//	[4 bytes] last login IP (or 6,_,_,_ for IPv6 truncated)
//	[short]   days since last login
//	[byte]    recovery days ago (0-200; 200=not set; -1 from server encodes as 200)
//	[short]   (unread_messages + 1)
func decodeWelcomeInfo(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	if b.Len() < 4 {
		return nil, fmt.Errorf("welcome_info: short payload %d", b.Len())
	}
	ipBytes, _ := b.ReadBytes(4)
	days, _ := b.ReadUint16()
	rec, _ := b.ReadByte()
	unread, _ := b.ReadUint16()
	ipStr := fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
	return event.WelcomeInfo{
		LastLoginIP:     ipStr,
		DaysSinceLogin:  int(days),
		RecoveryDaysAgo: int(rec),
		UnreadMessages:  int(unread) - 1,
	}, nil
}

// decodeGroundItem parses opcode 99. Each entry has either id+offsets
// (item appeared) or id with high bit (item removed). Multiple entries
// may be packed in one packet.
//
// Per GameStateUpdater.java:1128-1175.
func decodeGroundItem(payload []byte) (event.Event, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("ground_item: empty payload")
	}
	// Phase 1: emit one GroundItemEvent for the first entry only. Phase 2
	// will return a slice and the host loop will fan them out.
	b := WrapBuffer(payload)
	first, _ := b.ReadByte()
	if first == 0xFF {
		// First entry is a removal sentinel. The actual id follows…
		// but this packet uses 0xFF + offsets format inconsistently.
		// For now treat as unknown.
		return event.UnknownPacket{Opcode: InGroundItemHandler, PayloadSize: len(payload)}, nil
	}
	// Most common case: read raw byte we already consumed as high byte
	// of a uint16 item ID.
	low, _ := b.ReadByte()
	itemID := (int(first) << 8) | int(low)
	disappear := itemID&0x8000 != 0
	if disappear {
		itemID &= 0x7FFF
	}
	dx, _ := b.ReadByte()
	dy, _ := b.ReadByte()
	return event.GroundItemEvent{
		ItemID:    itemID,
		OffsetX:   int(int8(dx)),
		OffsetY:   int(int8(dy)),
		Disappear: disappear,
	}, nil
}

// decodeNpcDialogText parses opcode 222 (BOX). Per ActionSender.java
// (search for sendBox).
//
//	[zero-quoted string] text
func decodeNpcDialogText(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	text, err := b.ReadZeroQuotedString()
	if err != nil {
		return nil, fmt.Errorf("npc_dialog_text: %w", err)
	}
	return event.NpcDialogText{Text: text}, nil
}

// decodeNpcDialogOptions parses opcode 245 (OPTIONS_MENU_OPEN). Per
// ActionSender.java (search for sendMenu).
//
//	[byte] num_options
//	[for each: zero-quoted string] option_text
// decodeTradeOtherItems parses opcode 97 (TRADE_OTHER_ITEMS).
// Per ActionSender / Payload235Generator: opponent's offer.
//
//	[byte] count
//	for each: [short catalogID] [int amount]
func decodeTradeOtherItems(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	count, err := b.ReadByte()
	if err != nil {
		return event.TradeOtherOffer{}, nil
	}
	items := make([]event.InventoryItem, 0, count)
	for i := 0; i < int(count); i++ {
		id, err := b.ReadUint16()
		if err != nil {
			break
		}
		amt, err := b.ReadUint32()
		if err != nil {
			break
		}
		items = append(items, event.InventoryItem{ItemID: int(id), Amount: int(amt)})
	}
	return event.TradeOtherOffer{Items: items}, nil
}

func decodeNpcDialogOptions(payload []byte) (event.Event, error) {
	b := WrapBuffer(payload)
	num, _ := b.ReadByte()
	opts := make([]string, 0, num)
	for i := 0; i < int(num); i++ {
		s, err := b.ReadZeroQuotedString()
		if err != nil {
			break
		}
		opts = append(opts, s)
	}
	return event.NpcDialog{Options: opts}, nil
}
