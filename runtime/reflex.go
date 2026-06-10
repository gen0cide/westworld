package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// The social-reflex drive: the speed-2 actuation clock. Started by
// RunHost for mesa-linked hosts, it runs for the life of the host — a
// permanent runtime drive, not bootstrap. The reactive/speech/forage
// gates (reactive.go / speech.go / forage.go) decide; socialReflex
// actuates.

// socialReflex answers players who speak to the host on a cheap, reactive path
// (Game.Chat on the Haiku tier), independent of the Act planning loop. To give
// RICH replies (so "what are you doing?" gets a real answer), it tracks her
// latest published thought (reasoning + perception) off the bus and passes that
// as context — the same self-knowledge the Act loop has, without being the Act
// loop. Ignores her own echoed chat + rate-limits.
func socialReflex(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, username, goal string) {
	ch := host.Bus().Subscribe("*", 256)
	var last time.Time
	var doing, perception string // her most recent reasoning + perceived context
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch e := ev.(type) {
			case event.AgentThought:
				doing, perception = e.Reasoning, e.Perception // what she's currently up to
				// The AgentThought tick is a host-owned proactive clock: try the
				// intent-driven ASK drive here (its own AgentThought is filtered out
				// below — we ignore our own ask echo to avoid re-entrant asking).
				if e.Trigger != "ask" && e.Trigger != "forage" {
					tryAsk(ctx, log, host, mc, username)
					// FORAGE drive (5b): runs AFTER tryAsk and shares its global floor,
					// so exactly one of {ask, forage} actuates per gap — forage fires
					// only when no local interlocutor could be asked this tick.
					tryForage(ctx, log, host, mc, username)
				}
			case event.OtherPlayerChat:
				from := reflexPlayerName(host, e.PlayerIndex)
				if strings.EqualFold(from, username) {
					continue // her own chat echoed back — never reply to herself
				}
				// OPERATOR OVERRIDE trigger: "!<username> ANALYSIS" / "!<username>
				// RESUME" from the configured operator toggles analysis mode; while
				// active, every operator line is routed to the directive interpreter.
				// Auth is mandatory (a host-takeover vector otherwise): only an EXACT,
				// non-empty sender name == the configured operator is honored — the
				// "a player" placeholder for an unresolved index NEVER authenticates.
				if handled := reflexAnalysis(ctx, log, host, from, e.MessageText, username); handled {
					continue
				}
				// While analysis mode is active, the host is OFFLINE TO THE WORLD:
				// non-operator chat gets no in-character reply.
				if host.AnalysisActive() {
					continue
				}
				if time.Since(last) < 3*time.Second {
					continue // light rate-limit so rapid lines don't spam replies
				}
				// Knowledge-grounded reply (Deliverable 2): answer from what the host
				// actually KNOWS (hedged when low-confidence), say so honestly when it
				// does NOT, and optionally volunteer a high-confidence belief (the
				// host↔host propagation seed). Honesty is a host-supplied FACT, not a
				// hope about the LLM.
				rctx := socialContext(host, goal, doing, perception)
				rctx = append(rctx, host.groundReply(from, e.MessageText, time.Now())...)
				text, speak, err := mc.Chat(ctx, username, from, e.MessageText, rctx)
				if err != nil || !speak || text == "" {
					continue
				}
				// This is a DIRECTED reply to `from` — route the host's own line into
				// ONLY the addressee's conversation window so the Q→A pairs there and
				// does not broadcast into every latched conversation (L7). Genuinely-
				// public DSL chat (actions_social say()) keeps the untargeted fan.
				host.reactive.directSelfTo(normalizeSpeaker(from))
				if err := host.Say(ctx, text); err != nil {
					host.reactive.clearDirectSelf() // send failed → drop the unconsumed routing hint
					// Never swallow this again — a silently-dropped reply (e.g. over the
					// RSC 80-char limit) looks exactly like "she isn't talking".
					log.Warn("social: reply failed to send", "to", from, "said", text, "err", err)
				} else {
					last = time.Now()
					log.Info("social: replied", "to", from, "heard", e.MessageText, "said", text)
				}
			}
		}
	}
}

// socialContext gathers RICH context for a chat reply: her goal, what she's
// doing right now (latest reasoning), what she recently perceived, and the live
// game messages — so she can actually answer questions like "what are you doing?"
func socialContext(host *Host, goal, doing, perception string) []string {
	out := make([]string, 0, 8)
	if goal != "" {
		out = append(out, "Your overall goal: "+goal)
	}
	if doing != "" {
		out = append(out, "What you are doing RIGHT NOW: "+doing)
	}
	if perception != "" {
		out = append(out, "What you've recently seen/heard: "+perception)
	}
	if w := host.World(); w != nil && w.Self != nil {
		pos := w.Self.Position()
		out = append(out, fmt.Sprintf("You are at (%d,%d), HP %d/%d.", pos.X, pos.Y, w.Self.HP(), w.Self.MaxHP()))
	}
	return out
}

// reflexAnalysis handles the in-game operator-override channel for one inbound
// chat line. It returns true when the line was consumed by analysis mode (a
// trigger, or an operator directive while active) and the caller must NOT fall
// through to the normal social reply.
//
// AUTH (mandatory — a host-takeover vector otherwise): the line is only honored
// when `from` is an EXACT, case-sensitive, non-empty match for the configured
// operator. An unconfigured operator ("") disables the in-game channel entirely;
// reflexPlayerName's "a player" placeholder for an unresolved index can never
// match a real operator name. RSC names ride an untrusted channel, so this is
// the only gate between a passerby and a full host takeover.
func reflexAnalysis(ctx context.Context, log *slog.Logger, host *Host, from, message, username string) bool {
	op := host.AnalysisOperator()
	if op == "" || from == "" || from != op {
		return false // no operator configured, or sender is not the exact operator
	}
	trimmed := strings.TrimSpace(message)
	prefix := "!" + username + " "
	if cmd, ok := cutPrefixFold(trimmed, prefix); ok {
		switch {
		case cmd == "ANALYSIS": // UPPERCASE keyword per spec
			res := host.EnterAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "enter")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		case cmd == "RESUME":
			res := host.ExitAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "exit")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		}
		// "!<username> <directive>" while active: route the rest as a directive.
		if host.AnalysisActive() {
			res := host.Analyze(ctx, cmd)
			log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", cmd)
			_ = host.Say(ctx, analysisSayLine(res))
			return true
		}
		return false
	}
	// No "!<username>" addressing: while active, any bare operator line is a
	// directive (the operator is already in a takeover session with this host).
	if host.AnalysisActive() {
		res := host.Analyze(ctx, trimmed)
		log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", trimmed)
		_ = host.Say(ctx, analysisSayLine(res))
		return true
	}
	return false
}

// cutPrefixFold strips prefix (case-insensitively) from s, reporting whether it
// matched. The host-username addressing "!Delores " should match regardless of
// the operator's capitalization of the name, while the ANALYSIS/RESUME keyword
// itself stays case-sensitive (checked by the caller).
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(s[len(prefix):]), true
}

// analysisSayLine renders an analysis verdict as a single flat in-game line back
// to the operator (capped to the RSC chat limit by the action layer). Flat
// affect: terse + literal, prefixed so the operator can tell it from in-persona
// speech.
func analysisSayLine(res AnalysisResult) string {
	body := res.Text
	if res.Err != "" && body == "" {
		body = "error: " + res.Err
	}
	return "[analysis:" + string(res.Kind) + "] " + body
}

func reflexPlayerName(host *Host, idx int) string {
	if w := host.World(); w != nil && w.Players != nil {
		for _, p := range w.Players.All() {
			if p.Index == idx && p.Name != "" {
				return p.Name
			}
		}
	}
	return "a player"
}
