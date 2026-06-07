package mesaclient

import "github.com/gen0cide/westworld/persona"

// Provisioning is the compiled-behavior bundle a host pulls from mesa on connect
// (and, later, receives as live pushes over Subscribe). Mesa OWNS the persona
// lifecycle — authoring, the LLM prose cook, revisions — and serves the
// authoritative persona down; the host compiles it the rest of the way locally
// (persona.CompilePolicy → the pearl table), because the compiled table is
// func-valued and cannot cross the wire. So mesa is the source of truth for WHO
// a host is; the host is the source of truth for the in-RAM machinery.
type Provisioning struct {
	// Persona is the authoritative persona for this host. The host runs
	// persona.CompilePolicy on it to (re)build its pearl table + affect baseline.
	Persona persona.Persona `json:"persona"`
	// Prose is mesa's rendered identity card (the cook output, or the
	// deterministic Render floor) — used as system-prompt grounding for the LLM
	// seams. The host keeps it for display/debug; mesa also holds it for Decide.
	Prose string `json:"prose,omitempty"`
	// Goals is the host's current goal stack (north-star first), pushed by mesa.
	Goals []string `json:"goals,omitempty"`
}
