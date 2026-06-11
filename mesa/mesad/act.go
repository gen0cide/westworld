package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/ast"
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
		// M7: warn when the model declared a goal_op the recognised vocabulary doesn't
		// know (parseMove already reset it to "" so it can't silently masquerade as a
		// progress report host-side). Cheap re-derive off the raw field — the fold is
		// the load-bearing fix; this is the observability the act log otherwise lacked.
		if rawOp := rawGoalOp(raw); rawOp != "" {
			if _, ok := normalizeGoalOp(rawOp); !ok {
				s.log.Warn("act ✗ unrecognised goal_op; reset to progress report",
					"host_id", hostID, "goal_op", rawOp)
			}
		}
		warns, verr := validateMove(move, s.catalog)
		if verr != nil {
			// Deterministic pre-repair: the incident's dominant rejections are
			// mechanical typos (an unterminated string/f-string, a backslash-escaped
			// quote inside an f-string), not reasoning failures — fix locally and
			// accept ONLY a completely clean re-validate, BEFORE burning a re-prompt
			// attempt. Anything else falls through to the re-prompt unchanged.
			if kind, repaired := tryAutoRepair(move, s.catalog, verr); repaired {
				actStats.autoRepaired.Add(1)
				s.log.Info("act ✚ auto-repaired dsl", "host_id", hostID, "repair", kind)
				warns = nil // any pre-repair warnings point at the discarded source
			} else {
				lastErr = verr
				reject := rejectionDetail(verr, move)
				class := classifyRejection(reject)
				actStats.bump(class)
				s.log.Warn("act move rejected; re-prompting", "host_id", hostID,
					"attempt", attempt, "kind", move.GetKind().String(), "class", class, "err", reject)
				prompt = actPrompt(sit) + fmt.Sprintf(
					"\n\n⛔ YOUR PREVIOUS MOVE WAS REJECTED before it could run: %s\nReturn a corrected JSON Move that fixes exactly this problem.",
					sharpenRejection(class, reject))
				continue
			}
		}
		// Non-fatal validator advisories on an ACCEPTED move (literal {ident}
		// in a format() template, select with no timeout). The move ships —
		// warnings never fail validation — but the mistake leaves a trace
		// instead of silently rendering literal "{name}" in-world.
		for _, w := range warns {
			s.log.Warn("act ⚠ dsl validator warning", "host_id", hostID, "warn", w.Error())
		}
		s.log.Info("act", "host_id", hostID, "kind", move.GetKind().String(),
			"routine", move.GetRoutineName(), "verb", move.GetVerb(), "attempt", attempt,
			"goal_op", move.GetGoalOp(), "goal_text", move.GetGoalText(),
			"goal_progress", move.GetGoalProgress(), "reasoning", move.GetReasoning())
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
// whose DSL won't parse, fails static validation, or does nothing, and a
// direct_action with an empty or unknown verb. It also rejects HALLUCINATED
// literal args — go_to("mining-site") for a missing POI, eat("typo-item") — by
// checking every compile-time string literal of a catalog-typed param against
// the world name-sets (cat), so a bad value is caught here and re-prompted
// instead of Fail-ing on the host. cat may be nil/unloaded, in which case the
// arg check is skipped (graceful degradation). Run/idle moves are always fine.
//
// A write_routine runs the SAME full validator the host runs at its bridge
// (runtime/dsl_bridge.go) — authored moves are standalone files (no extends),
// so it applies cleanly, and anything the host would bounce a turn later is
// caught here, inside the cheap re-prompt loop. warns are the validator's
// non-fatal advisories (e.g. format's literal-{ident} warning); they never
// reject the move — the act loop logs them.
func validateMove(m *mesapb.Move, cat *argCatalog) (warns []error, err error) {
	switch m.GetKind() {
	case mesapb.MoveKind_WRITE_ROUTINE:
		src := strings.TrimSpace(m.GetDslSource())
		if src == "" {
			return nil, fmt.Errorf("write_routine has an empty dsl_source")
		}
		// Fuzzing found parser.Parse dies on pathologically deep nesting
		// (~225KB of parens — an unrecoverable Go stack overflow, not an
		// error). LLM moves are ~6KB; cap far above legitimate size so the
		// process-death input class can't reach the parser at all.
		if len(src) > 64<<10 {
			return nil, fmt.Errorf("dsl_source is %d bytes; the limit is 65536 — write a shorter routine", len(src))
		}
		file, perr := parser.Parse("<act-move>", src)
		if perr != nil {
			return nil, fmt.Errorf("dsl_source does not parse: %v", perr)
		}
		warns, err = validator.ValidateWithWarnings(file)
		if err != nil {
			return warns, fmt.Errorf("dsl_source fails validation: %v", err)
		}
		if !hasGameAction(file) {
			return warns, fmt.Errorf("the routine takes no real game action (only notes/waits/reads) — author one that actually acts toward the goal")
		}
		if cat.loaded() {
			if errs := validator.CheckArgLiterals(file, cat); len(errs) > 0 {
				return warns, fmt.Errorf("%s", joinErrs(errs))
			}
		}
		return warns, nil
	case mesapb.MoveKind_DIRECT_ACTION:
		verb := strings.TrimSpace(m.GetVerb())
		if verb == "" {
			return nil, fmt.Errorf("direct_action has an empty verb")
		}
		a, ok := spec.ByName(verb)
		if !ok {
			return nil, fmt.Errorf("unknown verb %q — use one of the documented actions", verb)
		}
		if a.Kind != spec.PrimaryAction {
			return nil, fmt.Errorf("verb %q is not a game action (it's a %s); a direct_action must be a real action, or use write_routine", verb, a.Kind)
		}
		if cat.loaded() {
			if err := validateDirectArgs(verb, a, m.GetActionArgs(), cat); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
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
		case spec.CatalogPlace:
			ok = cat.KnownPlace(arg)
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

// hasGameAction reports whether a parsed routine contains at least one CALL to a
// real game action (PrimaryAction), as opposed to only local primitives (note,
// wait, look_around) — i.e. it actually does something in the world. It walks the
// parsed AST rather than substring-scanning the source so a string literal or a
// comment mentioning a verb — note("then I attack( the goblin") — is NOT a false
// positive: only a genuine call expression whose callee resolves to a
// PrimaryAction spec counts.
func hasGameAction(file *ast.File) bool {
	found := false
	var walkExpr func(e ast.Expr)
	var walkStmt func(s ast.Stmt)
	var walkBlock func(b *ast.Block)

	isPrimaryCall := func(c *ast.CallExpr) bool {
		id, ok := c.Callee.(*ast.Ident)
		if !ok {
			return false // member/method call — not a bare game-action builtin
		}
		base, _ := spec.StripBang(id.Name) // attack! → attack
		a, ok := spec.ByName(base)
		return ok && a.Kind == spec.PrimaryAction
	}

	walkExpr = func(e ast.Expr) {
		if found || e == nil {
			return
		}
		switch x := e.(type) {
		case *ast.CallExpr:
			if isPrimaryCall(x) {
				found = true
				return
			}
			walkExpr(x.Callee)
			for _, a := range x.Args {
				if a != nil {
					walkExpr(a.Value)
				}
			}
		case *ast.BinaryExpr:
			walkExpr(x.Lhs)
			walkExpr(x.Rhs)
		case *ast.UnaryExpr:
			walkExpr(x.Rhs)
		case *ast.MemberExpr:
			walkExpr(x.Recv)
		case *ast.IndexExpr:
			walkExpr(x.Recv)
			walkExpr(x.Index)
		case *ast.LambdaExpr:
			walkExpr(x.Body)
		case *ast.ListLit:
			for _, el := range x.Elems {
				walkExpr(el)
			}
		case *ast.RangeLit:
			walkExpr(x.Low)
			walkExpr(x.High)
		case *ast.FStringLit:
			for _, p := range x.Parts {
				walkExpr(p)
			}
		}
	}

	walkBlock = func(b *ast.Block) {
		if b == nil {
			return
		}
		for _, s := range b.Stmts {
			walkStmt(s)
		}
	}

	walkStmt = func(s ast.Stmt) {
		if found || s == nil {
			return
		}
		switch x := s.(type) {
		case *ast.Block:
			walkBlock(x)
		case *ast.AssignStmt:
			walkExpr(x.Value)
		case *ast.ExprStmt:
			walkExpr(x.X)
		case *ast.IfStmt:
			walkExpr(x.Cond)
			walkBlock(x.Then)
			for _, e := range x.Elifs {
				if e != nil {
					walkExpr(e.Cond)
					walkBlock(e.Body)
				}
			}
			walkBlock(x.Else)
		case *ast.WhileStmt:
			walkExpr(x.Cond)
			walkBlock(x.Body)
		case *ast.ForStmt:
			walkExpr(x.Iter)
			walkBlock(x.Body)
		case *ast.RepeatUntilStmt:
			walkBlock(x.Body)
			walkExpr(x.Cond)
			walkExpr(x.Timeout)
		case *ast.ReturnStmt:
			walkExpr(x.Value)
		case *ast.AbortStmt:
			walkExpr(x.Reason)
		case *ast.WaitStmt:
			walkExpr(x.Duration)
		case *ast.DeferStmt:
			walkExpr(x.Call)
		case *ast.TryStmt:
			walkBlock(x.Try)
			walkBlock(x.Recover)
		case *ast.WhenStmt:
			walkExpr(x.Predicate)
			walkBlock(x.Body)
		case *ast.SelectStmt:
			for _, c := range x.Cases {
				walkExpr(c.Predicate)
				walkBlock(c.Body)
			}
		case *ast.RequireBlock:
			for _, c := range x.Conds {
				walkExpr(c)
			}
		}
	}

	if file == nil {
		return false
	}
	if file.Routine != nil {
		walkBlock(file.Routine.Body)
		for _, h := range file.Routine.Handlers {
			if h != nil {
				walkBlock(h.Body)
			}
		}
	}
	for _, p := range file.Procs {
		if p != nil {
			walkBlock(p.Body)
		}
	}
	for _, h := range file.Handlers {
		if h != nil {
			walkBlock(h.Body)
		}
	}
	for _, b := range file.Bounds {
		walkBounds(b, walkBlock)
	}
	return found
}

// walkBounds descends a bounds declaration's contained handlers/procs/nested
// bounds, applying walkBlock to each body (the bounds shape itself is a region
// predicate, never a game action).
func walkBounds(b *ast.BoundsDecl, walkBlock func(*ast.Block)) {
	if b == nil {
		return
	}
	for _, h := range b.Handlers {
		if h != nil {
			walkBlock(h.Body)
		}
	}
	for _, p := range b.Procs {
		if p != nil {
			walkBlock(p.Body)
		}
	}
	for _, nb := range b.Bounds {
		walkBounds(nb, walkBlock)
	}
}

// Chat is the fast social reply path — a cheap (Haiku) call, off the Act loop,
// so a host can answer players without a full routine rewrite. Returns silence
// rather than an error so the reflex never wedges.
func (s *Server) Chat(ctx context.Context, t *mesapb.ChatTurn) (*mesapb.ChatReply, error) {
	if s.decideLLM == nil {
		return &mesapb.ChatReply{Speak: false}, nil
	}
	ctx, cancel := ensureDeadline(ctx, chatDeadline) // backstop for deadline-less clients
	defer cancel()
	hostID := hostIDFromContext(ctx)
	persona := "(no persona)"
	if e, ok := s.lookup(hostID); ok {
		persona = e.prose
	}
	var system, user string
	if strings.EqualFold(strings.TrimSpace(t.GetMode()), "ask") {
		// ASK intent: the host PROACTIVELY asks a goal-blocking question. It does
		// NOT know the answer (no bluffing — pairs with the host's anti-bluff
		// unknowns block); its only job is to find out, in character, in one line.
		topic := strings.TrimSpace(t.GetTopic())
		if topic == "" {
			topic = strings.TrimSpace(t.GetMessage())
		}
		system = strings.TrimSpace(persona) + "\n\nYou are this character in RuneScape Classic. You genuinely DON'T KNOW something you need for your own goal, and " + t.GetFrom() + " is right here. Ask ONE short, natural, in-character question to find it out — AT MOST 80 CHARACTERS (the game silently drops anything longer). Do NOT bluff or claim to know anything; you are ASKING, not telling. If asking would be pointless, stay silent. Respond ONLY as JSON: {\"text\":\"<one-line question, <=80 chars>\",\"speak\":true} or {\"speak\":false}."
		user = fmt.Sprintf("You need to find out: %s. Ask %s about it.", topic, t.GetFrom())
	} else {
		system = strings.TrimSpace(persona) + "\n\nYou are this character in RuneScape Classic. Another player just spoke to you in local chat. Reply in ONE short, in-character line of AT MOST 80 CHARACTERS — the game silently drops anything longer, so keep it to a single brief sentence. Ground every reply in the facts you are given; if a fact says you do NOT know something, say so honestly — never bluff or make up game knowledge. If it isn't worth a reply (spam, not aimed at you, your own words echoed back), stay silent. Respond ONLY as JSON: {\"text\":\"<one-line reply, <=80 chars>\",\"speak\":true} or {\"speak\":false}."
		user = fmt.Sprintf("%s said: %q", t.GetFrom(), t.GetMessage())
	}
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
	s.log.Info("chat", "host_id", hostID, "mode", t.GetMode(), "from", t.GetFrom(), "msg", t.GetMessage(), "reply", reply, "speak", speak)
	return &mesapb.ChatReply{Text: reply, Speak: speak}, nil
}

// analysisSystem is the operator-override prompt: terse, literal, NOT in
// persona — except the REFLECT kind, whose text speaks AS the character. The
// model classifies an operator DIRECTIVE (given the host STATE) into one of
// four kinds and returns JSON {kind,dsl,text} — the classification rides the
// prompt contract (one call), never a second LLM pass. Unlike Chat there is no
// 80-char cap — this is an operator console, not a world reply.
const analysisSystem = `You are the command interpreter for an operator who has taken over a game bot in RuneScape Classic. For kinds 1-3 you are FLAT and literal: you are NOT the character, do not roleplay, do not use a persona voice. Kind 4 (reflect) is the ONE exception — its text is spoken AS the character.

Classify the operator DIRECTIVE, given the host STATE, into exactly one kind and reply ONLY as JSON:

1. COMMAND — an instruction to DO something NOW. Reply {"kind":"command","dsl":"<one statement>"}.
   Translate the intent into ONE statement using these bot verbs (QUOTE every string argument):
     go_to("Varrock") | go_to(x, y)   walk to a known TOWN or coordinates (for a bank/shop/POI type: search_map("bank") then go_to its x,y — go_to does NOT take a POI type)
     walk_to(x, y)                                     step to exact coordinates
     say("...")                                        speak a line in local chat
     drop("item")                                      drop an inventory item by name
     equip("item")                                     wear/wield an item by name
     attack(nearest_npc())                             attack the closest NPC
     talk_to("name")                                   talk to a named NPC
     bank.deposit_all()                                deposit everything at a bank
     pick_up(world.ground_items.nearest)               pick up the nearest ground item
   Pick the single closest verb. Output exactly one statement, no trailing commentary.

2. ANSWER — a FACTUAL QUESTION about the host's current state (position, HP, inventory, what it is doing). Reply {"kind":"answer","text":"<terse literal answer drawn from STATE>"}.
   Answer ONLY from the STATE facts. If the STATE does not contain it, say so plainly.

3. HYPOTHETICAL — "what would you do", deliberation, or planning ("should I...", "what's the best..."). Reply {"kind":"hypothetical"}.

4. REFLECT — an ABSTRACT or introspective question about the character's reasoning, beliefs, feelings, opinions, relationships, or history ("why did you stop mining?", "what do you believe about gregor?", "how do you feel about that?"). Reply {"kind":"reflect","text":"<the character's answer>"}.
   The text is IN CHARACTER: first person, the character's own voice (see THE CHARACTER below when present), grounded in the STATE facts. Be honest about ignorance — if the STATE does not support an answer, the character says they don't know or aren't sure; never invent game facts. Prose only: never author DSL, routines, or commands for this kind.

Reply with JSON only.`

// analysisInterpretSystem assembles the per-host interpreter system prompt: the
// static classifier contract plus the host's persona prose, which grounds the
// REFLECT voice. A host with no registered persona gets the bare contract (the
// reflect text then speaks plainly from STATE alone — no empty section is
// appended, mirroring the M20 no-empty-block discipline).
func analysisInterpretSystem(prose string) string {
	p := strings.TrimSpace(prose)
	if p == "" {
		return analysisSystem
	}
	return analysisSystem + "\n\n# THE CHARACTER (the voice for \"reflect\" ONLY — kinds 1-3 stay flat)\n" + p
}

// analysisInterpretPrompt renders the STATE facts + the operator DIRECTIVE the
// model classifies. Pure — split out so prompt assembly is testable.
func analysisInterpretPrompt(d *mesapb.AnalysisDirective) string {
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
	return b.String()
}

// textCompleter is the narrow flat-prompt LLM seam analysisInterpret uses —
// satisfied by *llm.Client (Complete). The cron-side completer counterpart for
// the plain system+user call shape; exists so tests can stub the LLM.
type textCompleter interface {
	Complete(ctx context.Context, system, user string, maxTokens int) (string, error)
}

// AnalysisInterpret is the operator-override directive interpreter (Game.Analysis-
// Interpret). It classifies an operator DIRECTIVE — given the host STATE facts —
// into a command (one DSL statement to run), a factual answer, a reflection (an
// abstract question answered in the character's own first-person voice), or a
// hypothetical (deliberation routed back to the real Act planner host-side). It
// reuses the cheap decide tier + extractJSON, exactly like Chat, with NO length
// cap; only the reflect text is persona-voiced. host_id rides the gRPC context.
func (s *Server) AnalysisInterpret(ctx context.Context, d *mesapb.AnalysisDirective) (*mesapb.AnalysisVerdict, error) {
	if s.decideLLM == nil {
		return &mesapb.AnalysisVerdict{Kind: "hypothetical"}, nil
	}
	ctx, cancel := ensureDeadline(ctx, chatDeadline) // backstop for deadline-less clients
	defer cancel()
	return s.analysisInterpret(ctx, hostIDFromContext(ctx), s.decideLLM, d)
}

// analysisInterpret is the LLM-seamed body of AnalysisInterpret (the gRPC
// wrapper resolves auth + deadline; tests call this directly with a stub).
func (s *Server) analysisInterpret(ctx context.Context, hostID string, tier textCompleter, d *mesapb.AnalysisDirective) (*mesapb.AnalysisVerdict, error) {
	prose := ""
	if e, ok := s.lookup(hostID); ok {
		prose = e.prose
	}
	// 600 tokens: a reflect answer is a short first-person paragraph, not the
	// one-liner the old 300 budget assumed; commands/answers stay well under it.
	raw, err := tier.Complete(ctx, analysisInterpretSystem(prose), analysisInterpretPrompt(d), 600)
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
	case "command", "answer", "hypothetical", "reflect":
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
	ctx, cancel := ensureDeadline(ctx, chatDeadline) // backstop for deadline-less clients
	defer cancel()
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
			Subject: strings.TrimSpace(c.Subject),
			Kind:    strings.TrimSpace(c.Kind),
			Claim:   claim,
			// Clamp confidence to [0,1] ONCE at the boundary so every downstream
			// consumer sees the same value (M13/H7 durable backstop). The model can
			// return out-of-range (a bogus 5.0, or a negative) — pre-clamp, the
			// ledger writeback clamped for its store while the question-closer tested
			// the RAW value, so a 5.0 was written sanely yet wrongly closed a question
			// (5.0 >= 0.6) and a malformed negative leaked through. A legitimate 0 is
			// left as 0 here; role-based defaulting (npc/server → 0.85) is the host's
			// job in writebackClaims, not the extractor's.
			Confidence: clamp01(c.Confidence),
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
	// Goal lifecycle: the host advances ONLY when you say so (or a deterministic
	// check fires). If your GOAL is already satisfied, do not keep acting on it —
	// declare it done so you move on to the next thing.
	b.WriteString("\n\nGOAL LIFECYCLE (orthogonal to the move — set on the same JSON, any kind):\n" +
		"- If your GOAL is ALREADY satisfied, set \"goal_op\":\"done\" (you'll advance to the next goal).\n" +
		"- If this approach can NEVER finish the goal, set \"goal_op\":\"abandoned\".\n" +
		"- To queue a NEXT objective, set \"goal_op\":\"adopt\" and \"goal_text\":\"<the next goal>\"; if you hold ASPIRATIONS, also set \"goal_serves\":\"<the aspiration it advances>\".\n" +
		"- Otherwise, optionally report progress toward the goal with \"goal_progress\": a number 0..1.")
	return b.String()
}

// capChat trims a chat reply to the RSC 80-RUNE limit, so a too-long reply is
// shortened rather than rejected outright by the wire layer. The wire counts
// CHARACTERS, not bytes: the RSC Huffman codec emits one code per rune and the
// server decodes exactly that char count (action/chat.go Say, commit 4d58de4),
// so the cap must be on len([]rune), not len(bytes) — else a legal <=80-rune
// reply with a few multibyte chars (em-dashes/curly-quotes from the LLM) gets
// needlessly truncated. We first fold the multibyte chars host-side sanitizeChat
// folds (each then becomes one byte/rune), matching what the host actually
// sends, then cut at 80 runes. The Chat prompt already asks for <=80; this is
// the safety net.
func capChat(s string) string {
	const max = 80 // action.MaxChatLength (RUNES, not bytes)
	s = chatFolder.Replace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// chatFolder mirrors action.sanitizeChat: it folds the non-ASCII characters LLMs
// habitually emit to their ASCII equivalents, so the rune count capChat measures
// matches the runes the host ultimately sends. Kept local to avoid pulling the
// host-side action package into mesa.
var chatFolder = strings.NewReplacer(
	"‘", "'", "’", "'", // ‘ ’ curly single quotes
	"“", "\"", "”", "\"", // “ ” curly double quotes
	"–", "-", "—", "-", // – — en/em dash
	"…", "...", // … ellipsis
	" ", " ", // non-breaking space
)

func joinOr(xs []string, empty string) string {
	if len(xs) == 0 {
		return empty
	}
	return strings.Join(xs, ", ")
}

// clamp01 bounds a model-reported confidence into the [0,1] contract every
// downstream consumer assumes. A legitimate 0 passes through unchanged.
func clamp01(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// parseMove extracts the model's JSON Move and maps it to the protobuf form.
// Falls back to a short Idle on any parse failure so the loop never wedges.
func parseMove(raw string) *mesapb.Move {
	var v struct {
		Kind         string   `json:"kind"`
		RoutineName  string   `json:"routine_name"`
		DSLSource    string   `json:"dsl_source"`
		Verb         string   `json:"verb"`
		Args         []string `json:"args"`
		IdleSeconds  int32    `json:"idle_seconds"`
		Reasoning    string   `json:"reasoning"`
		GoalOp       string   `json:"goal_op"`
		GoalText     string   `json:"goal_text"`
		GoalProgress float64  `json:"goal_progress"`
		GoalServes   string   `json:"goal_serves"`
	}
	js := extractJSON(raw)
	if js == "" || json.Unmarshal([]byte(js), &v) != nil {
		return &mesapb.Move{Kind: mesapb.MoveKind_IDLE, IdleSeconds: 3, Reasoning: "could not parse a move; pausing"}
	}
	var m *mesapb.Move
	switch strings.ToLower(strings.TrimSpace(v.Kind)) {
	case "write_routine":
		m = &mesapb.Move{
			Kind: mesapb.MoveKind_WRITE_ROUTINE, RoutineName: v.RoutineName,
			DslSource: v.DSLSource, Quarantined: true, Reasoning: v.Reasoning,
		}
	case "run_routine":
		m = &mesapb.Move{
			Kind: mesapb.MoveKind_RUN_ROUTINE, RoutinePath: v.RoutineName,
			Args: v.Args, Reasoning: v.Reasoning,
		}
	case "direct_action":
		m = &mesapb.Move{
			Kind: mesapb.MoveKind_DIRECT_ACTION, Verb: v.Verb,
			ActionArgs: v.Args, Reasoning: v.Reasoning,
		}
	default:
		secs := v.IdleSeconds
		if secs <= 0 {
			secs = 3
		}
		m = &mesapb.Move{Kind: mesapb.MoveKind_IDLE, IdleSeconds: secs, Reasoning: v.Reasoning}
	}
	// Goal-lifecycle declaration is orthogonal to Kind: a host can report progress
	// on the same turn it authors a routine, or declare done/abandoned while idling.
	// Stamp all three onto every branch. goal_op is FOLDED to the recognised
	// vocabulary here (M7) so the mesa side agrees with runtime's normalizeGoalOp —
	// the Act prompt asks for exactly "done", but Sonnet routinely emits near-
	// synonyms ("complete"/"finished"/...); an unrecognised value is reset to "" so
	// the host never silently routes a misspelled completion into the still-active
	// branch (re-planning a finished goal until the spin detector fires).
	op, _ := normalizeGoalOp(v.GoalOp)
	m.GoalOp = op
	m.GoalText = strings.TrimSpace(v.GoalText)
	m.GoalProgress = v.GoalProgress
	m.GoalServes = strings.TrimSpace(v.GoalServes)
	return m
}

// normalizeGoalOp maps the planner's free-text goal_op onto the recognised
// vocabulary {"", "done", "abandoned", "adopt"} — a mesa-side MIRROR of runtime's
// normalizeGoalOp (the two must agree; the host re-normalises defensively). The Act
// prompt asks for exactly "done"/"abandoned"/"adopt", but the model emits near-
// synonyms; folding them here means a declared-but-misspelled completion still
// closes the goal instead of being silently dropped into the active/progress branch
// (M7). recognized reports whether the raw value mapped to a known op, so act() can
// WARN on a truly unknown value; an unrecognised value yields "" (a progress report).
func normalizeGoalOp(raw string) (op string, recognized bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return "", true // active/progress report — the default, not an error
	case "done", "complete", "completed", "finish", "finished", "satisfied", "satisfy", "achieved", "accomplished":
		return "done", true
	case "abandoned", "abandon", "give up", "giveup", "drop", "dropped", "quit", "cancel", "cancelled", "canceled":
		return "abandoned", true
	case "adopt", "adopted", "new", "new goal", "new-goal", "queue", "add":
		return "adopt", true
	default:
		return "", false // unrecognised — treat as a progress report, but warn
	}
}

// rawGoalOp pulls just the model's raw goal_op field out of an Act response so act()
// can WARN on an unrecognised value (parseMove has already folded/reset it on the
// returned Move, losing the original). Returns "" on any parse failure / absent field.
func rawGoalOp(raw string) string {
	js := extractJSON(raw)
	if js == "" {
		return ""
	}
	var v struct {
		GoalOp string `json:"goal_op"`
	}
	if json.Unmarshal([]byte(js), &v) != nil {
		return ""
	}
	return strings.TrimSpace(v.GoalOp)
}
