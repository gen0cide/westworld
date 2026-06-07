package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// act runs the agent step: ground the model in the cached DSL manual + the
// host's persona, hand it the live situation, and parse its DSL Move. Uses the
// Act-tier model (Sonnet/Haiku) — high-volume, so the big manual is cached.
func (s *Server) act(ctx context.Context, hostID string, sit *mesapb.Situation) (*mesapb.Move, error) {
	persona := "(no persona)"
	if e, ok := s.lookup(hostID); ok {
		persona = e.prose
	}
	blocks := []llm.SystemBlock{
		{Text: dslManual, Cache: true}, // shared, static → prompt-cached prefix
		{Text: "# YOUR CHARACTER\n" + persona, Cache: false},
	}
	raw, err := s.actLLM.CompleteSystem(ctx, blocks, actPrompt(sit), 1500)
	if err != nil {
		return nil, err
	}
	move := parseMove(raw)
	s.log.Info("act", "host_id", hostID, "kind", move.GetKind().String(),
		"routine", move.GetRoutineName(), "verb", move.GetVerb(), "reasoning", move.GetReasoning())
	if src := move.GetDslSource(); src != "" {
		s.log.Info("act authored DSL [" + hostID + "]:\n" + src)
	}
	return move, nil
}

// Chat is the fast social reply path — a cheap (Haiku) call, off the Act loop,
// so a host can answer players without a full routine rewrite. Returns silence
// rather than an error so the reflex never wedges.
func (s *Server) Chat(ctx context.Context, t *mesapb.ChatTurn) (*mesapb.ChatReply, error) {
	if s.decideLLM == nil {
		return &mesapb.ChatReply{Speak: false}, nil
	}
	hostID := hostIDFromContext(ctx)
	persona := "(no persona)"
	if e, ok := s.lookup(hostID); ok {
		persona = e.prose
	}
	system := strings.TrimSpace(persona) + "\n\nYou are this character in RuneScape Classic. Another player just spoke to you in local chat. Reply in ONE short, in-character line. If it isn't worth a reply (spam, not aimed at you, your own words echoed back), stay silent. Respond ONLY as JSON: {\"text\":\"<one-line reply>\",\"speak\":true} or {\"speak\":false}."
	user := fmt.Sprintf("%s said: %q", t.GetFrom(), t.GetMessage())
	if r := t.GetRecent(); len(r) > 0 {
		user += "\nRecent chat:\n- " + strings.Join(r, "\n- ")
	}
	raw, err := s.decideLLM.Complete(ctx, system, user, 150)
	if err != nil {
		s.log.Warn("chat llm error", "host_id", hostID, "err", err)
		return &mesapb.ChatReply{Speak: false}, nil
	}
	var v struct {
		Text  string `json:"text"`
		Speak bool   `json:"speak"`
	}
	if js := extractJSON(raw); js != "" {
		_ = json.Unmarshal([]byte(js), &v)
	}
	reply := strings.TrimSpace(v.Text)
	speak := v.Speak && reply != ""
	s.log.Info("chat", "host_id", hostID, "from", t.GetFrom(), "msg", t.GetMessage(), "reply", reply, "speak", speak)
	return &mesapb.ChatReply{Text: reply, Speak: speak}, nil
}

// actPrompt renders the live situation the model reasons over.
func actPrompt(sit *mesapb.Situation) string {
	var b strings.Builder
	w := sit.GetWorld()
	if g := strings.TrimSpace(sit.GetGoal()); g != "" {
		fmt.Fprintf(&b, "GOAL: %s\n\n", g)
	}
	if lm := sit.GetHints()["latest_message"]; lm != "" {
		fmt.Fprintf(&b, "⚠ LATEST GAME FEEDBACK (heed this FIRST — it is often a prerequisite or a reason your last action was blocked): %q\n\n", lm)
	}
	if ex := sit.GetHints()["explore"]; ex != "" {
		fmt.Fprintf(&b, "🧭 %s\n\n", ex)
	}
	if pd := sit.GetHints()["player_directive"]; pd != "" {
		fmt.Fprintf(&b, "📍 A PLAYER IS TALKING TO YOU — %q. If this is a direction or instruction (e.g. which door to use, where to go), FOLLOW it: match it to the named objects/doors in 'what you see around you' (each shows its bearing from you) and act.\n\n", pd)
	}
	if opts := sit.GetHints()["dialog_options"]; opts != "" {
		fmt.Fprintf(&b, "🗨 A DIALOG MENU IS OPEN — you MUST answer it THIS turn, before anything else, or you stay stuck. Numbered options: %s\nReply with answer(N) (1-based). Pick the option that ADVANCES or ENDS the conversation (e.g. \"yes\"/\"ok\"/\"thank you\"/\"continue\"/\"I'm ready\") — do NOT pick options that just ask more questions and loop you back.\n\n", opts)
	}
	b.WriteString("SITUATION:\n")
	fmt.Fprintf(&b, "- Position: (%d, %d)", w.GetX(), w.GetY())
	if r := w.GetRegion(); r != "" {
		fmt.Fprintf(&b, " in %s", r)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "- HP: %d/%d   Combat level: %d   Fatigue: %.0f%%\n",
		w.GetHpCur(), w.GetHpMax(), w.GetCombatLevel(), w.GetFatigue()*100)
	fmt.Fprintf(&b, "- Inventory (%d free): %s\n", w.GetInvFree(), joinOr(w.GetInvSummary(), "empty"))
	fmt.Fprintf(&b, "- Nearby NPCs: %s\n", joinOr(w.GetNearbyNpcs(), "none"))
	if p := w.GetNearbyPlayers(); len(p) > 0 {
		fmt.Fprintf(&b, "- Nearby players: %s\n", strings.Join(p, ", "))
	}
	if scene := sit.GetHints()["scene"]; scene != "" {
		fmt.Fprintf(&b, "\nWHAT YOU SEE AROUND YOU (NPCs, doors/boundaries, scenery — with coordinates you can walk_to or open):\n%s\n", scene)
	}
	if rec := sit.GetRecent(); len(rec) > 0 {
		b.WriteString("\nRECENT IN-GAME MESSAGES & NPC SPEECH (oldest→newest — these tell you the current step):\n")
		for _, m := range rec {
			fmt.Fprintf(&b, "- %s\n", m)
		}
	}
	if la := sit.GetHints()["last_action"]; la != "" {
		fmt.Fprintf(&b, "\nWHAT YOU JUST DID (previous turn): %s → %s\n", la, sit.GetHints()["last_result"])
		if dsl := sit.GetHints()["last_dsl"]; dsl != "" {
			fmt.Fprintf(&b, "The routine you ran:\n%s\n", dsl)
		}
		b.WriteString("If that step already succeeded, DO NOT repeat it — advance to the NEXT step (e.g. the next instructor or task).\n")
	}
	if t := strings.TrimSpace(sit.GetTrigger()); t != "" {
		fmt.Fprintf(&b, "\n(Trigger: %s)\n", t)
	}
	b.WriteString("\nDecide the single next step toward the goal and return the JSON Move.")
	return b.String()
}

func joinOr(xs []string, empty string) string {
	if len(xs) == 0 {
		return empty
	}
	return strings.Join(xs, ", ")
}

// parseMove extracts the model's JSON Move and maps it to the protobuf form.
// Falls back to a short Idle on any parse failure so the loop never wedges.
func parseMove(raw string) *mesapb.Move {
	var v struct {
		Kind        string   `json:"kind"`
		RoutineName string   `json:"routine_name"`
		DSLSource   string   `json:"dsl_source"`
		Verb        string   `json:"verb"`
		Args        []string `json:"args"`
		IdleSeconds int32    `json:"idle_seconds"`
		Reasoning   string   `json:"reasoning"`
	}
	js := extractJSON(raw)
	if js == "" || json.Unmarshal([]byte(js), &v) != nil {
		return &mesapb.Move{Kind: mesapb.MoveKind_IDLE, IdleSeconds: 3, Reasoning: "could not parse a move; pausing"}
	}
	switch strings.ToLower(strings.TrimSpace(v.Kind)) {
	case "write_routine":
		return &mesapb.Move{
			Kind: mesapb.MoveKind_WRITE_ROUTINE, RoutineName: v.RoutineName,
			DslSource: v.DSLSource, Quarantined: true, Reasoning: v.Reasoning,
		}
	case "run_routine":
		return &mesapb.Move{
			Kind: mesapb.MoveKind_RUN_ROUTINE, RoutinePath: v.RoutineName,
			Args: v.Args, Reasoning: v.Reasoning,
		}
	case "direct_action":
		return &mesapb.Move{
			Kind: mesapb.MoveKind_DIRECT_ACTION, Verb: v.Verb,
			ActionArgs: v.Args, Reasoning: v.Reasoning,
		}
	default:
		secs := v.IdleSeconds
		if secs <= 0 {
			secs = 3
		}
		return &mesapb.Move{Kind: mesapb.MoveKind_IDLE, IdleSeconds: secs, Reasoning: v.Reasoning}
	}
}
