package event

// AgentThought is one turn of the mesa Act planner's reasoning, published to the
// bus so the debug control plane records it (GET /events?kind=agent_thought) and
// the host's full decision stream is observable in the browser UI + replayable
// from the JSONL log. It is emitted by the autonomous director each turn, after
// it perceives the situation and the planner returns a move.
type AgentThought struct {
	base
	Turn       int    `json:"turn"`
	Trigger    string `json:"trigger"`
	Goal       string `json:"goal,omitempty"`
	Pos        string `json:"pos"`
	HP         string `json:"hp"`
	Perception string `json:"perception"` // recent on-screen messages / dialog she saw
	Reasoning  string `json:"reasoning"`  // her first-person rationale
	MoveKind   string `json:"move_kind"`
	DSL        string `json:"dsl,omitempty"` // the routine she authored, if any
}

// Kind implements Event.
func (AgentThought) Kind() string { return "agent_thought" }

// PolicyVeto is emitted when the pearl gate blocks an action (a persona/quirk
// rule firing). Publishing it makes the veto visible to the host's own cognition
// (it lands in the Act transcript) and to the debug stream — otherwise a vetoed
// action looks like a silent no-op and the planner retries it forever.
type PolicyVeto struct {
	base
	Action string `json:"action"`
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

// Kind implements Event.
func (PolicyVeto) Kind() string { return "policy_veto" }
