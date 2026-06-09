package runtime

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/event"
)

// perception.go is the LIVE-WRITER half of the host's semantic mind: it folds
// raw game events into the world-knowledge ledger (cognition/knowledge) + the
// salience-gated observation firehose (observation.go). It is the sibling of
// limbic.go — where limbicHandle folds events into affect (mood) and the trust
// ledger (who), perceptionHandle folds them into beliefs about THINGS (npcs,
// places, shops, items) and streams the rich signals up to mesa.
//
// Phase 1 of docs/world-knowledge-and-learning.md: WIRE THE WRITERS, no
// reasoning. The host stays LIGHT — deterministic, in-RAM, cheap, O(1) per
// event, NO LLM, NO blocking. The heavy growth (authoritative shop identity,
// claim parsing, means-ends sub-goals) is mesa's job; the host only perceives,
// records cheap deterministic facts, and emits. Frozen under analysis-mode like
// limbicHandle (the operator override freezes all learning I/O).

// perceptionState is the perception handler's RAM-only cursor. lastNpc is the
// most recently named NPC the host interacted/conversed with (best-effort shop
// attribution); shopStock dedups per-item stock so a re-sent catalogue does not
// re-write knowledge; areas dedups area-familiarity bumps within a session.
type perceptionState struct {
	lastNpc     string
	shopSubject string          // identity of the shop shopStock belongs to
	shopStock   map[int]int     // itemID -> last-noted stock for the open shop
	areas       map[string]bool // gazetteer place name -> already-noted this session
}

// perceptionHandle folds one game event into the SEMANTIC ledgers (knowledge +
// goal graph) and the salience-gated observation firehose. It is the live-writer
// sibling of limbicHandle (who/affect): deterministic, in-RAM, no LLM, no
// blocking. Called from runLimbic on the shared "*" subscription, after the
// limbic fold. Frozen under operator analysis-mode, like limbicHandle.
func (h *Host) perceptionHandle(ev event.Event) {
	// ShopClosed is pure state-RESET (no learning write), so it must run even
	// during an analysis freeze: otherwise a ShopClosed arriving while frozen is
	// dropped, and the next post-freeze ShopOpened with the same subject inherits
	// the previous shop's per-item dedup slate, suppressing legitimate "sells X"
	// writes. Handle it before the AnalysisActive early-return.
	if _, ok := ev.(event.ShopClosed); ok {
		h.perceive.shopStock = nil // forget the open shop's snapshot
		h.perceive.shopSubject = ""
		return
	}
	if h.AnalysisActive() {
		return
	}
	switch e := ev.(type) {

	// --- shop catalogue: "where can I buy X" (the flagship writer) ---
	case event.ShopOpened:
		h.perceiveShop(e)

	// --- NPC attribution context (free, for shop/claim subjects) ---
	case event.NpcNearby:
		if !e.Removed {
			// Resolve the moving NPC by SCENE INDEX through the already-applied
			// world model — only the new-NPC wire section carries a real TypeID;
			// movement updates (the bulk of NpcNearby traffic) arrive with
			// TypeID==0, which would credit a phantom NPC type 0 ("Unicorn") on
			// every step. world.Apply runs before bus.Publish and MoveBy
			// preserves the stored TypeID, so index resolution names the moving
			// NPC on both spawn and movement.
			if n := h.npcNameByIndex(e.Index); n != "" {
				h.knowledge.Seen(n, "npc")
			}
		}

	// --- NPC dialog: CAPTURE the speech for the firehose AND set the
	// shop-keeper attribution cursor. The text is the substrate the reactive
	// tier (Phase 2) and the slow crons (Phase 4) extract claims from; the host
	// only CAPTURES here — no parsing (that's an LLM, i.e. mesa's job). The wire
	// gives no keyed NPC on a speech bubble, so attribute to the nearest named
	// NPC in the scene; NpcChat carries a scene index we can resolve directly. ---
	case event.NpcDialogText:
		if n := h.nearestNamedNpc(); n != "" {
			h.perceive.lastNpc = n
		}
		h.perceiveDialog("npc_dialog", h.npcDialogSubject(-1), e.Text, 0.6)
	case event.NpcChat:
		if n := h.npcNameByIndex(e.NpcIndex); n != "" {
			h.perceive.lastNpc = n
		}
		h.perceiveDialog("npc_dialog", h.npcDialogSubject(e.NpcIndex), e.MessageText, 0.6)
	case event.NpcDialog:
		if n := h.nearestNamedNpc(); n != "" {
			h.perceive.lastNpc = n
		}

	// --- player chat: overheard public chat + whispers. This is the AMBIENT
	// firehose (speed 3); the reactive trigger + per-speaker window (speed 2)
	// lands in Phase 2. Resolve the speaker by name and skip our own echo. ---
	case event.OtherPlayerChat:
		if name := h.playerNameByIndex(e.PlayerIndex); name != "" {
			h.perceiveDialog("player_chat", name, e.MessageText, 0.5)
		}
	case event.PrivateMessage:
		h.perceiveDialog("player_chat", e.Sender, e.Message, 0.6) // directed → a touch more salient

	// --- server / system messages: "the game told me something" (tutorial
	// prerequisites, quest hints — the canonical 'server talks to you' channel). ---
	case event.SystemMessage:
		h.perceiveDialog("server_msg", "server", e.Message, 0.55)
	case event.ChatReceived:
		if e.Speaker == "" {
			h.perceiveDialog("server_msg", "server", e.Message, 0.55)
		} else if !strings.EqualFold(e.Speaker, h.opts.Username) {
			h.perceiveDialog("player_chat", e.Speaker, e.Message, 0.5)
		}

	// --- area familiarity: "I have been here" ---
	case event.OwnPositionUpdate:
		h.perceiveArea(e.X, e.Y)

	// --- self outcomes: acquisition + combat resolution ---
	case event.ItemGained:
		h.perceiveItemGained(e)
	case event.TargetDied:
		h.perceiveKill(e)
	}
}

// perceiveShop is the flagship knowledge writer: it records what a shop sells
// (the "where do I buy a pickaxe" signal) as graded, provenance-tagged beliefs,
// and emits ONE rich observation (full catalogue) up to mesa, which reconciles
// authoritative shop identity from its transcript. Stock-deduped so a re-sent
// catalogue (RSC's stock-update path) does not double-count.
func (h *Host) perceiveShop(e event.ShopOpened) {
	// Best-effort subject: the keeper we were just talking to. Mesa reconciles
	// the authoritative identity from the transcript; the host only attributes.
	subject := h.perceive.lastNpc
	if subject == "" {
		subject = "general store"
		if !e.IsGeneral {
			subject = "shop"
		}
	}
	// Reset the per-item dedup slate when the SHOP changes — a ShopOpened with no
	// preceding ShopClosed must not inherit the previous shop's item ids
	// (cross-shop collision) — but PRESERVE it for a re-sent catalogue of the same
	// shop (RSC's stock-update path re-sends the catalogue; we must not re-count).
	if h.perceive.shopStock == nil || h.perceive.shopSubject != subject {
		h.perceive.shopStock = make(map[int]int, len(e.Items))
		h.perceive.shopSubject = subject
	}
	h.knowledge.Tag(subject, "shop") // once per open; Tag dedups internally

	names := make([]string, 0, 6)
	for _, it := range e.Items {
		name := itemName(h.facts, it.ItemID)
		// Unresolved ids (no facts / unknown) add no semantic value to the
		// ledger; still counted for the observation summary below.
		resolved := name != "" && !strings.HasPrefix(name, "item#")
		if resolved && len(names) < 6 {
			names = append(names, name)
		}
		if !resolved {
			continue
		}
		// Dedup gate: an identical re-send of this item's stock is a no-op.
		if prev, ok := h.perceive.shopStock[it.ItemID]; ok && prev == it.Stock {
			continue
		}
		h.perceive.shopStock[it.ItemID] = it.Stock

		claim := "sells " + name
		if it.Stock > 0 {
			// Saw it on the shelf myself — high confidence, direct observation.
			h.knowledge.Note(subject, "shop", claim, knowledge.ProvObserved, 0.9)
		} else {
			// Out of stock: explicit evidence AGAINST the claim, so confidence
			// falls if an item is perpetually unavailable.
			h.knowledge.Observe(subject, claim, false, 1.0)
		}
	}

	// ONE salience-gated observation per shop open (sparse — not per item). The
	// rich catalogue is exactly the "where to buy X" signal mesa distills.
	summary := strings.Join(names, ", ")
	text := fmt.Sprintf("shop %q stocks %d items: %s", subject, len(e.Items), summary)
	h.emitObservation("transaction", subject, text, 0.8)
}

// perceiveArea credits being AT a named place (location familiarity), deduped
// per session so a re-visited place never re-bumps or re-emits. Cheap Seen +
// one modest first-visit observation (a genuinely new area is mildly salient).
func (h *Host) perceiveArea(x, y int) {
	if h.facts == nil {
		return
	}
	place, dist, ok := h.facts.Gazetteer().NearestPlace(x, y)
	if !ok || dist > 3 { // only credit being AT a place (mirrors examine.go)
		return
	}
	if h.perceive.areas == nil {
		h.perceive.areas = map[string]bool{}
	}
	if h.perceive.areas[place.Name] {
		return // already noted this session — no re-bump, no spam
	}
	h.perceive.areas[place.Name] = true

	// Location analogue of ledger.Met: bump familiarity, assert no claim.
	h.knowledge.Seen(place.Name, "location")
	h.emitObservation("location", place.Name, "arrived at "+place.Name, 0.5)
}

// perceiveItemGained records that an item is obtainable by me (did-it-myself
// evidence). No observation — pickups are frequent/low-salience and would flood
// mesa; they stay local.
func (h *Host) perceiveItemGained(e event.ItemGained) {
	name := itemName(h.facts, e.ItemID)
	if name == "" || strings.HasPrefix(name, "item#") {
		return // unresolved id: no semantic value
	}
	h.knowledge.Observe(name, "obtainable by me", true, 1.0)
}

// perceiveKill records that an NPC type is killable by me, and emits a salient
// observation (a kill is a meaningful, relatively rare outcome worth streaming).
func (h *Host) perceiveKill(e event.TargetDied) {
	name := h.npcNameByType(e.TypeID)
	if name == "" {
		return
	}
	h.knowledge.Observe(name, "killable by me", true, 1.0)
	h.emitObservation("outcome", name, "defeated "+name, 0.7)
}

// perceiveDialog CAPTURES one line of dialog/chat/server text into the
// salience-gated firehose — capture only, no parsing (the reactive tier + crons
// extract claims). subject is the speaker (per-speaker tagging for the Phase-2
// reactive window); kind routes by source (npc_dialog / player_chat /
// server_msg). Skips empty text (the RSC decoder can fail to recover a body) and
// empty subjects so we never stream noise.
func (h *Host) perceiveDialog(kind, subject, text string, salience float64) {
	text = strings.TrimSpace(text)
	subject = strings.TrimSpace(subject) // a whitespace-only speaker would normalize to "" and conflate windows
	if text == "" || subject == "" {
		return
	}
	h.emitObservation(kind, subject, text, salience) // speed-3 ambient firehose (UNCHANGED)
	h.reactiveObserve(kind, subject, text, salience) // speed-2 window + trigger (reactive.go)
}

// npcDialogSubject names the NPC for a dialog observation: the indexed speaker
// when resolvable, else the best-effort conversation cursor (lastNpc), else a
// generic "npc". idx < 0 means the event carried no scene index (a speech bubble).
func (h *Host) npcDialogSubject(idx int) string {
	if name := h.npcNameByIndex(idx); name != "" {
		return name
	}
	if h.perceive.lastNpc != "" {
		return h.perceive.lastNpc
	}
	return "npc"
}

// npcNameByIndex resolves an NPC SCENE index (not a type id) to its friendly name
// via the world model + facts. Empty when out of view / unresolved / idx < 0.
func (h *Host) npcNameByIndex(idx int) string {
	if h.world == nil || idx < 0 {
		return ""
	}
	if rec, ok := h.world.Npcs.Get(idx); ok {
		return h.npcNameByType(rec.TypeID)
	}
	return ""
}

// playerNameByIndex resolves a player SCENE index to a name (mirrors limbic's
// OtherPlayerChat path) and skips the host's own echoed chat. Empty otherwise.
func (h *Host) playerNameByIndex(idx int) string {
	if h.world == nil {
		return ""
	}
	if p, ok := h.world.Players.Get(idx); ok && p.Name != "" && !strings.EqualFold(p.Name, h.opts.Username) {
		return p.Name
	}
	return ""
}

// npcNameByType resolves an NPC type id to its friendly name via facts (mirror
// of MesaDirector.npcName, but on *Host so perception can call it standalone).
// Empty when facts are unwired or the type is unknown.
func (h *Host) npcNameByType(typeID int) string {
	if h.facts != nil {
		if def := h.facts.NpcDef(typeID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return ""
}

// nearestNamedNpc returns the visible NPC with a known name CLOSEST to the host
// (best-effort shop-keeper attribution). O(n) over the small visible set.
// Iterating Npcs.All() ranges a Go map in randomized order, so the result must
// be made deterministic: pick the true nearest by Chebyshev distance, with a
// stable lowest-scene-index tie-break. The host-LIGHT invariant requires this to
// be deterministic — it feeds shop attribution + the npc_dialog firehose subject
// + the reactive window key. Empty when nothing named is in view.
func (h *Host) nearestNamedNpc() string {
	if h.world == nil || h.facts == nil {
		return ""
	}
	pos := h.world.Self.Position()
	best := ""
	bestDist := 1 << 30
	bestIdx := 1 << 30
	for _, n := range h.world.Npcs.All() {
		name := h.npcNameByType(n.TypeID)
		if name == "" {
			continue
		}
		d := chebyshev(pos.X, pos.Y, n.X, n.Y)
		if d < bestDist || (d == bestDist && n.Index < bestIdx) {
			best, bestDist, bestIdx = name, d, n.Index
		}
	}
	return best
}
