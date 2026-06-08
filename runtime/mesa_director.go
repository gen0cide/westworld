package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// MesaDirector is the autonomous agent loop: each turn it snapshots the host's
// game state into a mesaclient.Situation, asks mesa.Act "what do I do now?", and
// turns the returned Move (DSL) into a conductor Intent. mesa owns the planning
// (LLM); the host owns execution (interpreter + pearl gate). This is the seam
// the PROTOCOL.md "Act — the agent step" describes.
type MesaDirector struct {
	client mesaclient.Client
	hostID string
	goal   string
	log    *slog.Logger
	turn   int

	// sub + transcript give the planner MEMORY across turns: a rolling log of
	// narrative events (NPC speech, system messages, items/levels gained) drained
	// from the bus each turn. Without this, what an NPC said during converse() is
	// gone by the next turn and the host re-greets in a loop.
	sub        <-chan event.Event
	transcript []string

	// stuck detection: if she hasn't moved for several turns, her approach isn't
	// working — widen perception and tell her to explore instead of repeating.
	prevX, prevY int
	hasPrev      bool
	stuckTurns   int

	// lastPlayerMsg pins the most recent thing a real player said to her for a
	// few turns and surfaces it to the PLANNER (not just the chat reflex), so a
	// player's directions ("go through the north door") actually steer her.
	// lastPlayerName is who said it, so the planner can weigh the trust ledger
	// (a known friend's "follow me" lands differently than a stranger's).
	lastPlayerMsg  string
	lastPlayerName string
	playerMsgAge   int

	// visited tiles she has stood on — used to flag doors she has ALREADY gone
	// through (a door adjacent to a visited tile leads BACK) so she stops
	// backtracking through the same door in this linear, forward-only tutorial.
	visited map[[2]int]bool

	// keywordLadder is the session-genesis attention ladder: words/people that
	// should catch her attention in others' chat, each with a tier + reflex.
	// Surfaced to the Act planner so she orients to her name, friends, trade
	// words, and goal topics. Set once at login from the genesis compile.
	keywordLadder []mesaclient.KeywordRung
}

// SetKeywordLadder installs the session-genesis attention ladder (called once at
// login, before the conductor starts).
func (d *MesaDirector) SetKeywordLadder(ladder []mesaclient.KeywordRung) {
	d.keywordLadder = ladder
}

const transcriptCap = 80 // narrative lines retained; the last ~18 feed each turn

// NewMesaDirector builds a director that drives the host toward goal via mesa.Act.
func NewMesaDirector(client mesaclient.Client, hostID, goal string, log *slog.Logger) *MesaDirector {
	if log == nil {
		log = slog.Default()
	}
	return &MesaDirector{client: client, hostID: hostID, goal: goal, log: log}
}

// Next implements Director: build the situation, call Act, convert the Move.
// It logs the full thought stream — what Delores perceives, what she decides,
// and the exact DSL she authors — so a run is fully observable.
func (d *MesaDirector) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	d.ensureSub(h)
	d.drainTranscript(h) // fold this turn's narrative events into the rolling memory
	sit := d.situation(h, last)

	// --- perception: what she sees this turn ---
	d.log.Info("act ① perceives",
		"trigger", sit.Trigger,
		"pos", fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		"hp", fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		"inv", orNone(sit.World.InvSummary),
		"npcs", orNone(sit.World.NearbyNpcs),
		"dialog_open", sit.Hints["dialog_options"] != "",
		"dialog_options", sit.Hints["dialog_options"],
		"recalled_episodes", h.journalLen(),
		"nearby_players", sit.Hints["nearby_players"],
	)
	for _, m := range sit.Recent {
		d.log.Info("act ① recent message", "text", m)
	}
	if last.Err != nil {
		d.log.Info("act ① last error", "err", last.Err.Msg)
	}

	move, err := d.client.Act(ctx, sit)
	if err != nil {
		// Keep the loop alive on a planner error — pause and retry next turn.
		d.log.Warn("act ✗ planner failed; idling", "err", err)
		return d.idle(3), true
	}

	// --- decision: her reasoning + the DSL she wrote ---
	d.log.Info("act ② decides", "kind", moveKindName(move.Kind), "reasoning", move.Reasoning)
	if move.DSLSource != "" {
		d.log.Info("act ② authored DSL:\n" + move.DSLSource)
	}
	if move.Verb != "" {
		d.log.Info("act ② direct action", "verb", move.Verb, "args", strings.Join(move.ActionArgs, ", "))
	}

	// Publish the full thought to the bus so the debug control plane streams it
	// live (/ws, the browser dashboard) and records it (/events, JSONL).
	d.turn++
	h.bus.Publish(event.AgentThought{
		Turn: d.turn, Trigger: sit.Trigger, Goal: d.goal,
		Pos:        fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		HP:         fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		Perception: strings.Join(sit.Recent, " | "),
		Reasoning:  move.Reasoning, MoveKind: moveKindName(move.Kind), DSL: move.DSLSource,
	})
	return d.moveToIntent(move), true
}

func orNone(xs []string) string {
	if len(xs) == 0 {
		return "(none)"
	}
	return strings.Join(xs, ", ")
}

// situation snapshots the live world + affect + recent on-screen messages.
func (d *MesaDirector) situation(h *Host, last Outcome) *mesaclient.Situation {
	w := h.world
	pos := w.Self.Position()

	// Stuck detection: if she hasn't moved (within 1 tile) across turns, her
	// approach isn't working — widen perception and push her to explore.
	if d.hasPrev && absInt(pos.X-d.prevX) <= 1 && absInt(pos.Y-d.prevY) <= 1 {
		d.stuckTurns++
	} else {
		d.stuckTurns = 0
	}
	d.prevX, d.prevY, d.hasPrev = pos.X, pos.Y, true
	stuck := d.stuckTurns >= 3

	inv := make([]string, 0, 8)
	for _, sl := range w.Inventory.Slots() {
		name := d.itemName(h, sl.ItemID)
		if sl.Amount > 1 {
			name = fmt.Sprintf("%d %s", sl.Amount, name)
		}
		if sl.Wielded {
			name += " (worn)"
		}
		inv = append(inv, name)
	}

	// Nearby NPCs by name, de-duplicated (the scene often has many copies).
	npcSeen := map[string]bool{}
	npcs := make([]string, 0, 8)
	for _, n := range w.Npcs.All() {
		name := d.npcName(h, n.TypeID)
		if name == "" || npcSeen[name] {
			continue
		}
		npcSeen[name] = true
		npcs = append(npcs, name)
	}

	players := make([]string, 0, 4)
	for _, p := range w.Players.All() {
		if p.Name != "" {
			players = append(players, p.Name)
		}
	}

	// Recent narrative — the cross-turn transcript (NPC speech, system messages,
	// items/levels gained) the director accumulates from the bus. This is how the
	// host remembers what an NPC just told it.
	recent := lastN(d.transcript, 18)

	stress, confidence, valence := h.affect.Snapshot()

	trigger := triggerFor(last)
	if stuck {
		trigger = fmt.Sprintf("STUCK — you have not moved in %d turns; your current approach is NOT working, change it", d.stuckTurns)
	}

	sit := &mesaclient.Situation{
		HostID:  d.hostID,
		Goal:    d.goal,
		Trigger: trigger,
		World: mesaclient.WorldSnapshot{
			X: pos.X, Y: pos.Y,
			HPCur: w.Self.HP(), HPMax: w.Self.MaxHP(),
			CombatLevel: w.Self.CombatLevel(),
			Fatigue:     float64(w.Self.Fatigue()) / 750.0,
			InvFree:     w.Inventory.FreeSlots(),
			InvSummary:  inv, NearbyNpcs: npcs, NearbyPlayers: players,
		},
		Recent: recent,
		Affect: mesaclient.Affect{Stress: stress, Confidence: confidence, Valence: valence},
	}
	hints := map[string]string{}
	// Scene perception: NPCs, doors/boundaries, and notable scenery WITH
	// coordinates — so she can see (and walk to / open) the door an instruction
	// refers to. Widen the radius when stuck so an out-of-view exit appears.
	radius := 16
	if stuck {
		radius = 32
		hints["explore"] = "You are STUCK. Stop repeating the same action/coordinates. Re-read the latest instruction's DIRECTION (e.g. \"northeast\"). In RSC coordinates north = smaller y and EAST = SMALLER x (x increases to the west) — so northeast = smaller x AND smaller y. Look through 'what you see around you' for a door/exit in that direction (compare its coordinates to yours) and walk_to it. If none is listed, EXPLORE: walk ~10 tiles that way (e.g. northeast = your_x-10, your_y-10) and look again. Do NOT reuse a door you already came through."
	}
	hints["scene"] = d.describeArea(h, radius)
	// The single most-recent server message — often a blocking/prerequisite
	// notice ("Speak to the controls guide before going through this door"). It
	// must not get buried in the transcript history.
	if sm := w.Recent.ServerMessage(); sm != nil {
		if t := strings.TrimSpace(stripColors(sm.Message)); t != "" {
			hints["latest_message"] = t
		}
	}
	if opts := w.Recent.DialogOptions(); opts != nil && len(opts.Options) > 0 {
		numbered := make([]string, len(opts.Options))
		for i, o := range opts.Options {
			numbered[i] = fmt.Sprintf("%d. %s", i+1, o)
		}
		hints["dialog_options"] = strings.Join(numbered, "   ")
	}
	// A real player's recent words, surfaced to the PLANNER for a few turns: if
	// it's a direction/instruction, she should act on it. (Replying is handled
	// separately by the social reflex.) Annotate the speaker with what she knows
	// of them (trust ledger) so a friend's request and a stranger's read
	// differently — judged by the RELATIONSHIP, not by keywords.
	if d.lastPlayerMsg != "" && d.playerMsgAge < 3 {
		pd := d.lastPlayerMsg
		if note := d.trustNote(h, d.lastPlayerName); note != "" {
			pd += "  [what you know of " + d.lastPlayerName + ": " + note + "]"
		}
		hints["player_directive"] = pd
	}
	d.playerMsgAge++
	// Self-context: what she did last turn + how it ended, so she doesn't repeat
	// a step that already succeeded.
	if last.Intent.Label != "" {
		hints["last_action"] = last.Intent.Label
		hints["last_result"] = last.Kind.String()
		if last.Err != nil {
			hints["last_result"] = last.Kind.String() + ": " + last.Err.Msg
		}
		if last.Intent.Source != "" {
			hints["last_dsl"] = last.Intent.Source
		}
	}
	// Durable EPISODIC memory: record the standing objective (survives a restart
	// → feeds a future session-genesis) and recall the salient things she's done
	// this life into the Situation, so the planner builds on its own history
	// instead of re-deriving the world each tick.
	if h.journal != nil {
		h.journal.SetObjective(d.goal)
		if mem := d.memoryHint(h); mem != "" {
			hints["memory"] = mem
		}
	}
	// Proactive social presence: surface nearby players as people she MAY choose
	// to engage — greet, trade, help, follow — not only when chatted at. Each is
	// annotated with what she knows of them (trust ledger). Trade is first-class:
	// an offer is judged by the relationship + the actual deal, never by a word.
	if np := d.nearbyPlayersHint(h, pos.X, pos.Y); np != "" {
		hints["nearby_players"] = np
	}
	// Session-genesis attention ladder: the words/people she compiled at login as
	// worth noticing in chat (her name, friends, trade words, goal topics).
	if att := d.attentionHint(); att != "" {
		hints["attention"] = att
	}
	if len(hints) > 0 {
		sit.Hints = hints
	}
	return sit
}

// memoryHint renders the salient durable episodes for the Situation — what she
// has done this life (across restarts), so she doesn't redo finished steps and
// the planner has continuity the transcript ring (which resets) cannot give.
func (d *MesaDirector) memoryHint(h *Host) string {
	eps := h.journal.Salient(8)
	if len(eps) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Things you remember doing (durable across this life — don't redo what's already done):")
	for _, e := range eps {
		b.WriteString("\n- " + e.Text)
	}
	return b.String()
}

// nearbyPlayersHint lists players within reach as optional social opportunities,
// each tagged with her standing relationship (trust grade, familiarity, notes).
// px,py is her current position. Empty when no one is around.
func (d *MesaDirector) nearbyPlayersHint(h *Host, px, py int) string {
	const radius = 16
	parts := make([]string, 0, 4)
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-px)+absInt(p.Y-py) > radius {
			continue
		}
		who := d.trustNote(h, p.Name)
		if who == "" {
			who = "a stranger so far"
		}
		parts = append(parts, fmt.Sprintf("%s (to your %s @ %d,%d — %s)",
			p.Name, bearingFrom(px, py, p.X, p.Y), p.X, p.Y, who))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// attentionHint renders the genesis keyword ladder for the Situation: the
// words/people she should notice in chat, grouped with their default reflex.
func (d *MesaDirector) attentionHint() string {
	if len(d.keywordLadder) == 0 {
		return ""
	}
	parts := make([]string, 0, len(d.keywordLadder))
	for _, r := range d.keywordLadder {
		if r.Keyword == "" {
			continue
		}
		s := "\"" + r.Keyword + "\""
		if r.Action != "" {
			s += " → " + r.Action
		} else if r.Tier != "" {
			s += " (" + r.Tier + ")"
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// trustNote summarizes what the host knows of a player from the trust ledger —
// grade, familiarity, and any tags — or "" for someone she's never met. This is
// the substrate for judging an offer by the RELATIONSHIP rather than a keyword.
func (d *MesaDirector) trustNote(h *Host, name string) string {
	if h.ledger == nil || name == "" || !h.ledger.Known(name) {
		return ""
	}
	rel := h.ledger.Rel(name)
	parts := make([]string, 0, 3)
	parts = append(parts, rel.Grade.String())
	if rel.Familiar > 0 {
		parts = append(parts, fmt.Sprintf("met %d×", rel.Familiar))
	}
	parts = append(parts, rel.Tags...)
	return strings.Join(parts, ", ")
}

// ensureSub lazily subscribes to the bus (once) so the director can build a
// rolling narrative transcript across turns.
func (d *MesaDirector) ensureSub(h *Host) {
	if d.sub == nil {
		d.sub = h.bus.Subscribe("*", 4096)
	}
}

// drainTranscript pulls all currently-buffered bus events and appends the
// narrative-salient ones (NPC speech, system messages, items/levels) to the
// transcript, de-duplicating adjacent repeats.
func (d *MesaDirector) drainTranscript(h *Host) {
	for {
		select {
		case ev, ok := <-d.sub:
			if !ok {
				d.sub = nil
				return
			}
			if pc, isChat := ev.(event.OtherPlayerChat); isChat {
				if name := d.playerName(h, pc.PlayerIndex); !strings.EqualFold(name, d.hostID) {
					d.lastPlayerMsg = name + ": " + strings.TrimSpace(stripColors(pc.MessageText))
					d.lastPlayerName = name
					d.playerMsgAge = 0
				}
			}
			line := d.narrativeLine(h, ev)
			if line == "" {
				continue
			}
			if n := len(d.transcript); n > 0 && d.transcript[n-1] == line {
				continue
			}
			d.transcript = append(d.transcript, line)
			if len(d.transcript) > transcriptCap {
				d.transcript = d.transcript[len(d.transcript)-transcriptCap:]
			}
		default:
			return
		}
	}
}

// narrativeLine formats a bus event as a transcript line, or "" to skip it.
func (d *MesaDirector) narrativeLine(h *Host, ev event.Event) string {
	switch e := ev.(type) {
	case event.SystemMessage:
		return strings.TrimSpace(stripColors(e.Message))
	case event.NpcDialogText:
		return "NPC: " + strings.TrimSpace(stripColors(e.Text))
	case event.NpcChat:
		return "NPC: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.OtherPlayerChat:
		name := d.playerName(h, e.PlayerIndex)
		if strings.EqualFold(name, d.hostID) {
			return "" // her OWN chat echoed back as local chat — never treat as someone talking to her
		}
		return name + " says to you: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.ChatReceived:
		if e.Speaker != "" {
			return e.Speaker + ": " + strings.TrimSpace(stripColors(e.Message))
		}
	case event.ItemGained:
		return fmt.Sprintf("(received %s ×%d)", d.itemName(h, e.ItemID), e.Count)
	case event.LevelUp:
		return fmt.Sprintf("(leveled up — skill %d is now %d)", int(e.Skill), e.NewLevel)
	case event.PolicyVeto:
		return fmt.Sprintf("⛔ YOUR OWN NATURE blocked '%s': %s — this action will NEVER work for you; do something else.", e.Action, e.Reason)
	}
	return ""
}

// moveToIntent turns a mesa Move into a conductor Intent.
func (d *MesaDirector) moveToIntent(m *mesaclient.Move) Intent {
	switch m.Kind {
	case mesaclient.MoveWriteRoutine:
		name := m.RoutineName
		if name == "" {
			name = "act_step"
		}
		return Intent{Label: "act:" + name, Name: name, Source: m.DSLSource}
	case mesaclient.MoveRunRoutine:
		return Intent{Label: "act:" + m.RoutinePath, RoutinePath: m.RoutinePath, Args: toValues(m.Args)}
	case mesaclient.MoveDirectAction:
		// Guard: an empty/whitespace verb must NOT become `act_direct() { () }` —
		// that parses, "completes" in 0s doing nothing, and (pre-OneShot) used to be
		// cached + replayed forever. Treat a verbless direct action as a brief idle.
		if strings.TrimSpace(m.Verb) == "" {
			d.log.Warn("act ✗ direct action with empty verb — idling instead of authoring a no-op")
			return d.idle(2)
		}
		// Wrap a single verb in a one-shot routine so it runs through the gate.
		src := fmt.Sprintf("runtime \"1.0\"\nroutine act_direct() {\n    %s(%s)\n}", m.Verb, dslArgList(m.ActionArgs))
		return Intent{Label: "act:" + m.Verb, Name: "act_direct", Source: src, OneShot: true}
	default: // MoveIdle
		secs := m.IdleSeconds
		if secs <= 0 {
			secs = 2
		}
		return d.idle(secs)
	}
}

func (d *MesaDirector) idle(secs int) Intent {
	src := fmt.Sprintf("runtime \"1.0\"\nroutine act_idle() {\n    wait(%d)\n}", secs)
	return Intent{Label: "act:idle", Name: "act_idle", Source: src, OneShot: true}
}

func (d *MesaDirector) npcName(h *Host, typeID int) string {
	if h.facts != nil {
		if def := h.facts.NpcDef(typeID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return ""
}

// describeArea gives the planner FULL area context: every nearby NPC, object,
// scenery, and door with NAME and COORDINATES (nearest-first), unfiltered — so
// she can see the cooking range, fishing spot, rock, etc. and act on them
// (walk_to or use items on them by coordinate). The old notable-only filter hid
// interactables like the range; this does not.
func (d *MesaDirector) describeArea(h *Host, radius int) string {
	pos := h.world.Self.Position()
	// Record where she is so doors she's passed can be flagged later.
	if d.visited == nil {
		d.visited = map[[2]int]bool{}
	}
	d.visited[[2]int{pos.X, pos.Y}] = true

	type obj struct {
		label string
		dist  int
	}
	var objs []obj
	seen := map[string]bool{}
	add := func(label string, x, y int) {
		if label == "" {
			return
		}
		full := fmt.Sprintf("%s @ (%d,%d), to your %s", label, x, y, bearingFrom(pos.X, pos.Y, x, y))
		if seen[full] { // dedup by name+coords, so multiple doors all show
			return
		}
		seen[full] = true
		objs = append(objs, obj{label: full, dist: absInt(x-pos.X) + absInt(y-pos.Y)})
	}
	// NPCs (with coords, for talk_to / use targeting).
	for _, n := range h.world.Npcs.All() {
		if absInt(n.X-pos.X)+absInt(n.Y-pos.Y) > radius {
			continue
		}
		add(d.npcName(h, n.TypeID), n.X, n.Y)
	}
	// Players (with coords + bearing) — so she knows which way a real player is
	// and can answer/follow "I'm to your east" correctly instead of guessing.
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-pos.X)+absInt(p.Y-pos.Y) > radius {
			continue
		}
		add("player "+p.Name, p.X, p.Y)
	}
	// Static scenery + boundaries from facts (UNFILTERED — includes the range).
	if h.facts != nil {
		for _, p := range h.facts.Near(pos.X, pos.Y, radius) {
			switch p.Kind {
			case "scenery":
				add(p.Name, p.X, p.Y)
			case "boundary":
				label := p.Name
				if d.doorUsed(p.X, p.Y) {
					label += " (you've been through this one before)"
				}
				add(label, p.X, p.Y)
			}
		}
	}
	// Dynamic scenery (GameObjects: fires, etc.), names resolved via facts.
	if h.world.Scenery != nil {
		for _, s := range h.world.Scenery.Near(pos.X, pos.Y, radius) {
			name := ""
			if h.facts != nil {
				if def := h.facts.SceneryDef(s.ID); def != nil {
					name = def.Name
				}
			}
			add(name, s.X, s.Y)
		}
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].dist < objs[j].dist })
	if len(objs) > 28 {
		objs = objs[:28]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "You are at (%d,%d). Nearby (name @ x,y — walk_to them, or use items on scenery via use(item, x=, y=)):\n", pos.X, pos.Y)
	for _, o := range objs {
		fmt.Fprintf(&b, "- %s\n", o.label)
	}
	return b.String()
}

// bearingFrom returns a compass direction from (px,py) to (x,y) in RSC
// coordinates (north = smaller y, EAST = smaller x — the x-axis increases west).
// So a player saying "the north door" maps to the door with the smaller y.
func bearingFrom(px, py, x, y int) string {
	dx := x - px // east is negative dx
	dy := y - py // north is negative dy
	if dx == 0 && dy == 0 {
		return "right here"
	}
	var ns, ew string
	if dy < 0 {
		ns = "north"
	} else if dy > 0 {
		ns = "south"
	}
	if dx < 0 {
		ew = "east"
	} else if dx > 0 {
		ew = "west"
	}
	switch {
	case ns != "" && ew != "" && absInt(dy) > 2*absInt(dx):
		return ns
	case ns != "" && ew != "" && absInt(dx) > 2*absInt(dy):
		return ew
	case ns != "" && ew != "":
		return ns + ew
	case ns != "":
		return ns
	default:
		return ew
	}
}

// doorUsed reports whether she has already stood on or beside a door tile (so
// it leads back the way she came). Used to flag doors as do-not-reuse.
func (d *MesaDirector) doorUsed(x, y int) bool {
	for _, off := range [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		if d.visited[[2]int{x + off[0], y + off[1]}] {
			return true
		}
	}
	return false
}

func (d *MesaDirector) playerName(h *Host, idx int) string {
	for _, p := range h.world.Players.All() {
		if p.Index == idx && p.Name != "" {
			return p.Name
		}
	}
	return "a player"
}

func (d *MesaDirector) itemName(h *Host, itemID int) string {
	if h.facts != nil {
		if def := h.facts.ItemDef(itemID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return fmt.Sprintf("item#%d", itemID)
}

// --- helpers ----------------------------------------------------------------

func triggerFor(last Outcome) string {
	switch {
	case last.Intent.Label == "": // zero outcome
		return "start"
	case last.Err != nil:
		return "recover"
	default:
		return "continue"
	}
}

func lastN(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return xs[len(xs)-n:]
}

func toValues(args []string) []interp.Value {
	if len(args) == 0 {
		return nil
	}
	vs := make([]interp.Value, 0, len(args))
	for _, a := range args {
		vs = append(vs, interp.String(a))
	}
	return vs
}

// dslArgList renders direct-action args as DSL literals: numbers bare, else quoted.
func dslArgList(args []string) string {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if isNumeric(a) {
			parts = append(parts, a)
		} else {
			parts = append(parts, "\""+strings.ReplaceAll(a, "\"", "\\\"")+"\"")
		}
	}
	return strings.Join(parts, ", ")
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if (r == '-' || r == '+') && i == 0 {
			continue
		}
		if r == '.' {
			continue
		}
		return false
	}
	return true
}

// stripColors removes RSC "@xxx@" colour codes (5 chars: @ + 3 + @) and turns
// "%" line breaks into spaces, so message text reads cleanly to the model.
func stripColors(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '@' && i+4 < len(s) && s[i+4] == '@' {
			i += 5
			continue
		}
		if s[i] == '%' {
			b.WriteByte(' ')
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func moveKindName(k mesaclient.MoveKind) string { return k.String() }
