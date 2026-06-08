package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/validator"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// maxActAttempts bounds the author→validate→re-prompt self-correction loop, so a
// model that keeps emitting invalid DSL can't spin forever. After the last
// attempt we fall back to a brief idle rather than shipping a broken move.
const maxActAttempts = 3

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

	// Author → validate → re-prompt. mesa parses + sanity-checks the move BEFORE it
	// round-trips to the host, so a syntax error or an empty/unknown verb is caught
	// here and the model self-corrects — instead of the host wasting a turn running
	// (or no-op'ing) broken DSL.
	prompt := actPrompt(sit)
	var lastErr error
	for attempt := 1; attempt <= maxActAttempts; attempt++ {
		raw, err := s.actLLM.CompleteSystem(ctx, blocks, prompt, 1500)
		if err != nil {
			return nil, err
		}
		move := parseMove(raw)
		if verr := validateMove(move, s.catalog); verr != nil {
			lastErr = verr
			s.log.Warn("act move rejected; re-prompting", "host_id", hostID,
				"attempt", attempt, "kind", move.GetKind().String(), "err", verr)
			prompt = actPrompt(sit) + fmt.Sprintf(
				"\n\n⛔ YOUR PREVIOUS MOVE WAS REJECTED before it could run: %s\nReturn a corrected JSON Move that fixes exactly this problem.", verr)
			continue
		}
		s.log.Info("act", "host_id", hostID, "kind", move.GetKind().String(),
			"routine", move.GetRoutineName(), "verb", move.GetVerb(), "attempt", attempt,
			"reasoning", move.GetReasoning())
		if src := move.GetDslSource(); src != "" {
			s.log.Info("act authored DSL [" + hostID + "]:\n" + src)
		}
		return move, nil
	}

	// Exhausted retries — idle rather than ship a broken move.
	s.log.Warn("act: no valid move after retries; idling", "host_id", hostID, "err", lastErr)
	return &mesapb.Move{Kind: mesapb.MoveKind_IDLE, IdleSeconds: 3,
		Reasoning: "could not author a valid move"}, nil
}

// validateMove rejects degenerate moves before they reach the host: a write_routine
// whose DSL won't parse or does nothing, and a direct_action with an empty or
// unknown verb. It also rejects HALLUCINATED literal args — go_to("mining-site")
// for a missing POI, eat("typo-item") — by checking every compile-time string
// literal of a catalog-typed param against the world name-sets (cat), so a bad
// value is caught here and re-prompted instead of Fail-ing on the host. cat may
// be nil/unloaded, in which case the arg check is skipped (graceful degradation).
// Run/idle moves are always fine.
func validateMove(m *mesapb.Move, cat *argCatalog) error {
	switch m.GetKind() {
	case mesapb.MoveKind_WRITE_ROUTINE:
		src := strings.TrimSpace(m.GetDslSource())
		if src == "" {
			return fmt.Errorf("write_routine has an empty dsl_source")
		}
		file, err := parser.Parse("<act-move>", src)
		if err != nil {
			return fmt.Errorf("dsl_source does not parse: %v", err)
		}
		if !hasGameAction(src) {
			return fmt.Errorf("the routine takes no real game action (only notes/waits/reads) — author one that actually acts toward the goal")
		}
		if cat.loaded() {
			if errs := validator.CheckArgLiterals(file, cat); len(errs) > 0 {
				return fmt.Errorf("%s", joinErrs(errs))
			}
		}
	case mesapb.MoveKind_DIRECT_ACTION:
		verb := strings.TrimSpace(m.GetVerb())
		if verb == "" {
			return fmt.Errorf("direct_action has an empty verb")
		}
		a, ok := spec.ByName(verb)
		if !ok {
			return fmt.Errorf("unknown verb %q — use one of the documented actions", verb)
		}
		if a.Kind != spec.PrimaryAction {
			return fmt.Errorf("verb %q is not a game action (it's a %s); a direct_action must be a real action, or use write_routine", verb, a.Kind)
		}
		if cat.loaded() {
			if err := validateDirectArgs(verb, a, m.GetActionArgs(), cat); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateDirectArgs checks a direct_action's bare string args positionally
// against the verb's catalog-typed params — the direct_action counterpart of
// the AST literal check (its args arrive as already-bare strings, not an AST).
// Only catalogued params are checked; npc params are soft (skipped).
func validateDirectArgs(verb string, a *spec.ActionSpec, args []string, cat *argCatalog) error {
	if a.ParamKinds == nil {
		return nil
	}
	for i, raw := range args {
		kind := a.ParamKind(i)
		if kind == spec.CatalogNone || kind == spec.CatalogNPC {
			continue
		}
		arg := strings.TrimSpace(raw)
		if arg == "" {
			continue
		}
		ok := true
		switch kind {
		case spec.CatalogPlaceOrPOI:
			ok = cat.KnownPlaceOrPOI(arg)
		case spec.CatalogItem:
			ok = cat.KnownItem(arg)
		}
		if ok {
			continue
		}
		param := "arg"
		if i < len(a.Params) {
			param = a.Params[i]
		}
		ae := &validator.ArgError{Verb: verb, Param: param, Kind: kind, Value: arg, Samples: cat.Examples(kind)}
		return fmt.Errorf("%s", ae.Error())
	}
	return nil
}

// joinErrs renders one or more arg errors as a single re-prompt string.
func joinErrs(errs []error) string {
	if len(errs) == 1 {
		return errs[0].Error()
	}
	parts := make([]string, len(errs))
	for i, e := range errs {
		parts[i] = "• " + e.Error()
	}
	return strings.Join(parts, "\n")
}

// hasGameAction reports whether a parsed routine contains at least one call to a
// real game action (PrimaryAction), as opposed to only local primitives (note,
// wait, look_around) — i.e. it actually does something in the world.
func hasGameAction(src string) bool {
	for _, a := range spec.Actions {
		if a.Kind != spec.PrimaryAction {
			continue
		}
		if containsCall(src, a.Name) {
			return true
		}
	}
	return false
}

// containsCall is a cheap check for `name(` appearing as a call in src, ignoring
// matches that are part of a longer identifier (e.g. "walk" in "walk_to").
func containsCall(src, name string) bool {
	for i := 0; ; {
		j := strings.Index(src[i:], name+"(")
		if j < 0 {
			return false
		}
		j += i
		if j == 0 || !isIdentRune(src[j-1]) {
			return true
		}
		i = j + len(name)
	}
}

func isIdentRune(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
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
	system := strings.TrimSpace(persona) + "\n\nYou are this character in RuneScape Classic. Another player just spoke to you in local chat. Reply in ONE short, in-character line of AT MOST 80 CHARACTERS — the game silently drops anything longer, so keep it to a single brief sentence. If it isn't worth a reply (spam, not aimed at you, your own words echoed back), stay silent. Respond ONLY as JSON: {\"text\":\"<one-line reply, <=80 chars>\",\"speak\":true} or {\"speak\":false}."
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
	reply := capChat(strings.TrimSpace(v.Text)) // RSC drops messages > 80 chars; never ship one that would be silently rejected
	speak := v.Speak && reply != ""
	s.log.Info("chat", "host_id", hostID, "from", t.GetFrom(), "msg", t.GetMessage(), "reply", reply, "speak", speak)
	return &mesapb.ChatReply{Text: reply, Speak: speak}, nil
}

// analysisSystem is the FLAT operator-override prompt: terse, literal, NOT in
// persona. The model classifies an operator DIRECTIVE (given the host STATE)
// into one of three kinds and returns JSON {kind,dsl,text}. Unlike Chat there
// is no 80-char cap and no character voice — this is an operator console, not a
// world reply.
const analysisSystem = `You are a FLAT, literal command interpreter for an operator who has taken over a game bot in RuneScape Classic. You are NOT the character. Do not roleplay, do not use a persona voice. Be terse and literal.

Classify the operator DIRECTIVE, given the host STATE, into exactly one kind and reply ONLY as JSON:

1. COMMAND — an instruction to DO something NOW. Reply {"kind":"command","dsl":"<one statement>"}.
   Translate the intent into ONE statement using these bot verbs (QUOTE every string argument):
     go_to("bank") | go_to("Varrock") | go_to(x, y)   walk to a named place/POI or coordinates
     walk_to(x, y)                                     step to exact coordinates
     say("...")                                        speak a line in local chat
     drop("item")                                      drop an inventory item by name
     equip("item")                                     wear/wield an item by name
     attack(nearest_npc())                             attack the closest NPC
     talk_to("name")                                   talk to a named NPC
     bank.deposit_all()                                deposit everything at a bank
     pick_up(world.ground_items.nearest)               pick up the nearest ground item
   Pick the single closest verb. Output exactly one statement, no trailing commentary.

2. ANSWER — a FACTUAL QUESTION about the host's current state. Reply {"kind":"answer","text":"<terse literal answer drawn from STATE>"}.
   Answer ONLY from the STATE facts. If the STATE does not contain it, say so plainly.

3. HYPOTHETICAL — "what would you do", deliberation, or planning ("should I...", "what's the best..."). Reply {"kind":"hypothetical"}.

Reply with JSON only.`

// AnalysisInterpret is the operator-override directive interpreter (Game.Analysis-
// Interpret). It classifies a flat operator DIRECTIVE — given the host STATE
// facts — into a command (one DSL statement to run), a factual answer, or a
// hypothetical (deliberation routed back to the real Act planner host-side). It
// reuses the cheap decide tier + extractJSON, exactly like Chat, but with a flat
// (non-persona) prompt and NO length cap. host_id rides the gRPC context.
func (s *Server) AnalysisInterpret(ctx context.Context, d *mesapb.AnalysisDirective) (*mesapb.AnalysisVerdict, error) {
	if s.decideLLM == nil {
		return &mesapb.AnalysisVerdict{Kind: "hypothetical"}, nil
	}
	hostID := hostIDFromContext(ctx)
	var b strings.Builder
	b.WriteString("STATE:\n")
	if facts := d.GetState(); len(facts) > 0 {
		for _, f := range facts {
			fmt.Fprintf(&b, "- %s\n", f)
		}
	} else {
		b.WriteString("- (no live state available)\n")
	}
	fmt.Fprintf(&b, "\nDIRECTIVE: %s", strings.TrimSpace(d.GetDirective()))

	raw, err := s.decideLLM.Complete(ctx, analysisSystem, b.String(), 300)
	if err != nil {
		s.log.Warn("analysis interpret llm error", "host_id", hostID, "err", err)
		return nil, err
	}
	var v struct {
		Kind string `json:"kind"`
		DSL  string `json:"dsl"`
		Text string `json:"text"`
	}
	if js := extractJSON(raw); js != "" {
		_ = json.Unmarshal([]byte(js), &v)
	}
	kind := strings.ToLower(strings.TrimSpace(v.Kind))
	switch kind {
	case "command", "answer", "hypothetical":
	default:
		kind = "hypothetical" // unclassifiable → defer to the real planner host-side
	}
	verdict := &mesapb.AnalysisVerdict{Kind: kind, Dsl: strings.TrimSpace(v.DSL), Text: strings.TrimSpace(v.Text)}
	s.log.Info("analysis interpret", "host_id", hostID, "directive", d.GetDirective(),
		"kind", verdict.Kind, "dsl", verdict.Dsl, "text", verdict.Text)
	return verdict, nil
}

// extractDialogSystem is the FLAT reactive-tier prompt: not the persona voice, a
// narrow analyzer. It turns a windowed exchange into structured claims + one
// classified intent. The §3.6 Tier-1 reactive job — cheap, fast, JSON-only.
const extractDialogSystem = `You analyze a short overheard/directed exchange from RuneScape Classic and extract structured knowledge. You are NOT the character — do not roleplay. Be literal.

The window is chronological (oldest first); lines tagged "Me:" are the host's OWN words. Return ONLY JSON:
{"claims":[{"subject":"<what it's about>","kind":"<item|shop|npc|place|quest|concept|recipe|...>","claim":"<a single concrete fact stated/implied>","confidence":0.0-1.0,"provenance":"player-told|npc-told|server-msg|implied"}],
 "intent":{"kind":"query|offer|warning|instruction|statement|greeting|...","urgency":"immediate|high|normal|low","gist":"<one short line: what the speaker wants from the host RIGHT NOW>"}}

Rules:
- Extract only DURABLE facts worth remembering (where to buy/find things, prices, dangers, who does what, how to do something). Skip small talk, insults, and pure noise — return an empty claims list for those.
- confidence reflects how firmly the fact was asserted (a hedge "maybe" is low; a flat statement is high).
- urgency is how time-sensitive the speaker's intent is to the HOST: a warning of imminent danger or a direct instruction/offer needing a reply NOW is "immediate"/"high"; ambient chatter is "low".
- If nothing is asserted and nothing is wanted, return {"claims":[],"intent":{"kind":"statement","urgency":"low","gist":""}}.

Reply with JSON only.`

// extractDialogPrompt renders the windowed exchange + light grounding context.
func extractDialogPrompt(w *mesapb.DialogWindow) string {
	var b strings.Builder
	role := strings.TrimSpace(w.GetSpeakerRole())
	if role == "" {
		role = "player"
	}
	fmt.Fprintf(&b, "SPEAKER: %s (%s)\n", strings.TrimSpace(w.GetSpeaker()), role)
	if p := strings.TrimSpace(w.GetPersonaSnippet()); p != "" {
		fmt.Fprintf(&b, "YOU ARE: %s\n", p)
	}
	if g := strings.TrimSpace(w.GetActiveGoal()); g != "" {
		fmt.Fprintf(&b, "YOUR ACTIVE GOAL: %s\n", g)
	}
	if q := w.GetOpenQuestions(); len(q) > 0 {
		b.WriteString("YOUR OPEN QUESTIONS (a claim that answers one of these is valuable):\n")
		for _, oq := range q {
			fmt.Fprintf(&b, "- %s\n", oq)
		}
	}
	b.WriteString("\nEXCHANGE (chronological, oldest first; the LAST line is NEWEST):\n")
	win := w.GetWindow()
	for i, line := range win {
		tag := ""
		if i == len(win)-1 {
			tag = "   ⟵ NEWEST"
		}
		fmt.Fprintf(&b, "%d. %s%s\n", i+1, line, tag)
	}
	b.WriteString("\nExtract the claims + the speaker's intent. Return JSON only.")
	return b.String()
}

// ExtractDialog is the reactive tier (Game.ExtractDialog): given a windowed
// exchange the host latched onto, return the claims to write into its knowledge
// ledger + one classified intent the host uses (deterministically) to decide
// whether to interrupt. Model tier is HAIKU by default (decideLLM), escalating
// to the Sonnet-class Act tier (actLLM) for nuance (a longer window or a goal in
// play) — NEVER Opus. Parse failure degrades to an empty set + a low-urgency
// statement so the host's reactive path never errors out. host_id rides the
// gRPC context.
func (s *Server) ExtractDialog(ctx context.Context, w *mesapb.DialogWindow) (*mesapb.ExtractedDialogSet, error) {
	safe := func() *mesapb.ExtractedDialogSet {
		return &mesapb.ExtractedDialogSet{Intent: &mesapb.DialogIntent{Kind: "statement", Urgency: "low"}}
	}
	if s.decideLLM == nil {
		return safe(), nil
	}
	hostID := hostIDFromContext(ctx)

	// Tier select (deterministic, server-side): Haiku base, escalate to the
	// Sonnet-class Act tier for nuance — a longer exchange or one that touches an
	// active goal. Falls back to decideLLM when actLLM isn't wired. Never Opus.
	tier := s.decideLLM
	if extractWantsNuance(w) && s.actLLM != nil {
		tier = s.actLLM
	}

	raw, err := tier.Complete(ctx, extractDialogSystem, extractDialogPrompt(w), 700)
	if err != nil {
		s.log.Warn("extract dialog llm error", "host_id", hostID, "speaker", w.GetSpeaker(), "err", err)
		return safe(), nil
	}
	set := parseExtractedDialog(raw)
	s.log.Info("extract dialog", "host_id", hostID, "speaker", w.GetSpeaker(),
		"role", w.GetSpeakerRole(), "claims", len(set.GetClaims()),
		"intent", set.GetIntent().GetKind(), "urgency", set.GetIntent().GetUrgency())
	return set, nil
}

// extractWantsNuance is the deterministic Sonnet-escalation predicate for the
// reactive tier: escalate to the nuance pass only for a genuinely long multi-turn
// exchange (≥5 lines), which is where Haiku's single-pass extraction tends to
// miss cross-line context. A merely goal-relevant one-liner stays on Haiku —
// "Haiku for the bulk, scale up only when more depth is needed" (cost stance);
// the active-goal signal is already supplied to Haiku as prompt context. Pure +
// unit-tested. (Dial: re-add a goal/contradiction clause once Haiku quality is
// measured in live play.)
func extractWantsNuance(w *mesapb.DialogWindow) bool {
	return len(w.GetWindow()) >= 5
}

// parseExtractedDialog maps the model's JSON to the protobuf ExtractedDialogSet.
// On any parse failure it returns an empty claims set + a low-urgency statement
// intent (never nil), so the host's reactive path is always safe to read.
func parseExtractedDialog(raw string) *mesapb.ExtractedDialogSet {
	out := &mesapb.ExtractedDialogSet{Intent: &mesapb.DialogIntent{Kind: "statement", Urgency: "low"}}
	js := extractJSON(raw)
	if js == "" {
		return out
	}
	var v struct {
		Claims []struct {
			Subject    string  `json:"subject"`
			Kind       string  `json:"kind"`
			Claim      string  `json:"claim"`
			Confidence float64 `json:"confidence"`
			Provenance string  `json:"provenance"`
		} `json:"claims"`
		Intent struct {
			Kind    string `json:"kind"`
			Urgency string `json:"urgency"`
			Gist    string `json:"gist"`
		} `json:"intent"`
	}
	if json.Unmarshal([]byte(js), &v) != nil {
		return out
	}
	for _, c := range v.Claims {
		claim := strings.TrimSpace(c.Claim)
		if claim == "" {
			continue // a claim with no content is noise
		}
		out.Claims = append(out.Claims, &mesapb.DialogClaim{
			Subject:    strings.TrimSpace(c.Subject),
			Kind:       strings.TrimSpace(c.Kind),
			Claim:      claim,
			Confidence: c.Confidence,
			Provenance: strings.TrimSpace(c.Provenance),
		})
	}
	if k := strings.TrimSpace(v.Intent.Kind); k != "" {
		out.Intent.Kind = k
	}
	if u := strings.TrimSpace(v.Intent.Urgency); u != "" {
		out.Intent.Urgency = strings.ToLower(u)
	}
	out.Intent.Gist = strings.TrimSpace(v.Intent.Gist)
	return out
}

// actPrompt renders the live situation the model reasons over.
func actPrompt(sit *mesapb.Situation) string {
	var b strings.Builder
	w := sit.GetWorld()
	if g := strings.TrimSpace(sit.GetGoal()); g != "" {
		fmt.Fprintf(&b, "GOAL: %s\n\n", g)
	}
	// Phase-2a reasoning layer: frame everything with the INTENTION STRUCTURE (the
	// goal's blockers / sub-goals / downstream value) and the curiosity-gated
	// LEARNING PRIORITY, so the model reasons over the intention graph rather than
	// a flat goal line. Both are omitted when the host had nothing to surface.
	if it := sit.GetHints()["intention"]; it != "" {
		fmt.Fprintf(&b, "📊 YOUR INTENTION STRUCTURE (your goal and what blocks/enables it — reason over THIS, not just the goal line):\n%s\n\n", it)
	}
	if lp := sit.GetHints()["learning_priority"]; lp != "" {
		fmt.Fprintf(&b, "🎯 LEARNING PRIORITY: %s\n\n", lp)
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
	if np := sit.GetHints()["nearby_players"]; np != "" {
		fmt.Fprintf(&b, "\n👥 NEARBY PLAYERS — real people you MAY choose to approach (greet, trade, help, follow) if it fits who you are and what you want; you can also just keep to your goal. Trading is normal and useful — judge any offer by your relationship with them and the actual deal, never by a single word like \"free\": %s\n", np)
	}
	if mem := sit.GetHints()["memory"]; mem != "" {
		fmt.Fprintf(&b, "\n🧠 %s\n", mem)
	}
	// Phase-2a: what the host KNOWS (graded + provenanced) relevant to its goal,
	// then what it does NOT reliably know — the explicit-unknowns block is the
	// direct antidote to confidently acting from ignorance (go_to("mining-site")).
	if kn := sit.GetHints()["known"]; kn != "" {
		// 📚 (learned facts about the world), distinct from 🧠 memory above (things
		// the host has done) — different hint types must not share an icon.
		fmt.Fprintf(&b, "\n📚 WHAT YOU KNOW (relevant to your goal — confidence + how you learned it):%s\n", kn)
	}
	if un := sit.GetHints()["unknowns"]; un != "" {
		fmt.Fprintf(&b, "\n❓ %s\n", un)
	}
	if att := sit.GetHints()["attention"]; att != "" {
		fmt.Fprintf(&b, "\n👂 WHAT CATCHES YOUR ATTENTION (if you see these in chat, orient — your name, friends, trade words, your topics): %s\n", att)
	}
	if rec := sit.GetRecent(); len(rec) > 0 {
		b.WriteString("\nRECENT IN-GAME MESSAGES & NPC SPEECH — CHRONOLOGICAL: line 1 is OLDEST, the LAST line is NEWEST. The newest line is what just happened and is usually your current step; read to the bottom:\n")
		for i, m := range rec {
			tag := ""
			if i == len(rec)-1 {
				tag = "   ⟵ NEWEST / most recent"
			}
			fmt.Fprintf(&b, "%d. %s%s\n", i+1, m, tag)
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

// capChat trims a chat reply to the RSC 80-byte limit on a rune boundary, so a
// too-long (or multibyte) reply is shortened rather than rejected outright by the
// wire layer. The Chat prompt already asks for <=80; this is the safety net.
func capChat(s string) string {
	const max = 80 // action.MaxChatLength
	if len(s) <= max {
		return s
	}
	var n int
	for i, r := range s {
		if i+len(string(r)) > max {
			return s[:n]
		}
		n = i + len(string(r))
	}
	return s[:n]
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
