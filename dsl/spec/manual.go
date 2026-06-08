package spec

import (
	"sort"
	"strings"
)

// APIReference renders the COMPLETE DSL surface — every action, accessor, and
// event the engine supports — as a brain-readable reference for the Act planner's
// manual. It is generated from the canonical spec (Actions / Accessors / Events),
// so it can NEVER drift from the engine: add a callable to the spec and it appears
// here automatically. This is what stops the planner from "having" a capability
// (e.g. bank.deposit) it was never told about and improvising around it.
//
// NotYetImplemented entries are omitted (the wrapper is still a stub).
func APIReference() string {
	var b strings.Builder
	b.WriteString("# COMPLETE API REFERENCE\n")
	b.WriteString("This is the ENTIRE surface you can call. If something is listed here you can use it; if it is NOT listed, it does not exist — do not invent calls. Namespaced calls use dots (e.g. bank.deposit, trade.request, magic.cast). Items/NPCs are passed as views or names, never raw nulls.\n")

	// --- Actions, grouped by kind ---
	b.WriteString("\n## ACTIONS — change game state; each returns a Result (.err / .val)\n")
	for _, grp := range []struct {
		kind  ActionKind
		title string
	}{
		{PrimaryAction, "Game actions (move, interact, fight, bank, trade, …)"},
		{LLMStdlib, "Brain calls (LLM-backed — expensive, use sparingly)"},
		{MemoryStdlib, "Memory & external (recall, relations, observe)"},
		{Primitive, "Primitives — local, no game effect (waits, reads, geometry)"},
		{PersonaRead, "Persona / identity reads"},
	} {
		var lines []string
		for _, a := range Actions {
			if a.Kind != grp.kind || a.NotYetImplemented {
				continue
			}
			lines = append(lines, "- "+actionSig(a)+orDash(a.DocSummary))
		}
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		b.WriteString("\n### " + grp.title + "\n")
		b.WriteString(strings.Join(lines, "\n") + "\n")
	}

	// --- Accessors / namespaced calls, grouped by first path segment ---
	b.WriteString("\n## ACCESSORS & NAMESPACED CALLS — read state in expressions, or callable(...) where the type says so\n")
	groups := map[string][]string{}
	var order []string
	for _, ac := range Accessors {
		if ac.NotYetImplemented || len(ac.Path) == 0 {
			continue
		}
		ns := ac.Path[0]
		if _, seen := groups[ns]; !seen {
			order = append(order, ns)
		}
		line := "- " + strings.Join(ac.Path, ".")
		if ac.Kind != "" {
			line += "  [" + ac.Kind + "]"
		}
		groups[ns] = append(groups[ns], line+orDash(ac.DocSummary))
	}
	sort.Strings(order)
	for _, ns := range order {
		lines := groups[ns]
		sort.Strings(lines)
		b.WriteString("\n### " + ns + ".*\n" + strings.Join(lines, "\n") + "\n")
	}

	// --- Events ---
	b.WriteString("\n## EVENTS — register a live handler with `on <name>(params) { ... }`\n")
	var evs []string
	for _, e := range Events {
		if e.NotYetImplemented {
			continue
		}
		evs = append(evs, "- on "+e.Name+"("+strings.Join(e.Params, ", ")+")"+orDash(e.DocSummary))
	}
	sort.Strings(evs)
	b.WriteString(strings.Join(evs, "\n") + "\n")

	return b.String()
}

func actionSig(a ActionSpec) string {
	return a.Name + "(" + strings.Join(a.Params, ", ") + ")"
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return " — " + s
}
