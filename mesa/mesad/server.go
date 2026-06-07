// Package mesad is the mesa service: the off-host home for everything that is
// not compute-local-feasible. It owns the persona lifecycle (storage, prose
// cook, revisions) and the LLM seams (Decide/Act), and will grow RAG (Recall)
// and the long-term-memory crons. Hosts connect over gRPC as their own host_id,
// provision their compiled persona down, then escalate decisions/knowledge up.
//
// This is mesa-side code (gitignored): it holds the Anthropic key and makes the
// external calls the host is forbidden from making.
package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// Server implements the mesa gRPC services (Game/Knowledge/Journal/Provision).
// It is concurrency-safe.
type Server struct {
	mesapb.UnimplementedGameServer
	mesapb.UnimplementedKnowledgeServer
	mesapb.UnimplementedJournalServer
	mesapb.UnimplementedProvisionServer

	actLLM    *llm.Client // Act/DSL authoring tier (Sonnet/Haiku) — high volume, cached prefix
	decideLLM *llm.Client // narrow option-pick tier (Haiku)
	log       *slog.Logger

	mu     sync.RWMutex
	reg    map[string]*entry            // host_id → its persona/prose/system prompt
	mem    map[string][]*mesapb.Episode // host_id → mirrored episodes (scaffold LTM)
	tokens map[string]string            // bearer token → host_id (auth)
}

// entry is a registered host's compiled identity, held mesa-side.
type entry struct {
	persona persona.Persona
	prose   string // rendered identity card (cook output / Render floor)
	system  string // the decide/act system prompt derived from prose
	goals   []string
}

// New builds a mesa server with per-tier LLM clients: actLLM authors DSL moves
// (Act — Sonnet/Haiku, high volume), decideLLM resolves narrow option picks
// (Decide — Haiku). Either may be nil (that seam then returns Unavailable → the
// host degrades to local behavior); persona provisioning works without either.
func New(actLLM, decideLLM *llm.Client, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		actLLM:    actLLM,
		decideLLM: decideLLM,
		log:       log,
		reg:       map[string]*entry{},
		mem:       map[string][]*mesapb.Episode{},
		tokens:    map[string]string{},
	}
}

// Register stores a persona for a host_id, rendering its prose card + system
// prompt once. It validates the persona first.
func (s *Server) Register(hostID string, p persona.Persona) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("register %s: %w", hostID, err)
	}
	prose := persona.Render(&p)
	s.mu.Lock()
	s.reg[hostID] = &entry{persona: p, prose: prose, system: decideSystem(prose), goals: goalsOf(p)}
	s.mu.Unlock()
	// Auto-authorize with the derived host key (placeholder: SHA-512(host_id) —
	// the host derives the same on its side). Replace with a real secret later.
	s.Authorize(hostID, mesaclient.HostKey(hostID))
	s.log.Info("registered host persona", "host_id", hostID,
		"name", p.Cornerstone.Identity.Name, "prose_chars", len(prose))
	return nil
}

// Register wires the server onto a gRPC server as all four services.
func (s *Server) Attach(gs *grpc.Server) {
	mesapb.RegisterGameServer(gs, s)
	mesapb.RegisterKnowledgeServer(gs, s)
	mesapb.RegisterJournalServer(gs, s)
	mesapb.RegisterProvisionServer(gs, s)
}

func (s *Server) lookup(hostID string) (*entry, bool) {
	s.mu.RLock()
	e, ok := s.reg[hostID]
	s.mu.RUnlock()
	return e, ok
}

// --- Game ------------------------------------------------------------------

// Decide resolves an in-routine option choice via the LLM, grounded in the
// host's persona.
func (s *Server) Decide(ctx context.Context, c *mesapb.Choice) (*mesapb.Decision, error) {
	if s.decideLLM == nil {
		return nil, status.Error(codes.Unavailable, "llm not configured")
	}
	hostID := hostIDFromContext(ctx) // authenticated identity, not the request body
	system := genericDecideSystem
	if e, ok := s.lookup(hostID); ok {
		system = e.system
	}
	raw, err := s.decideLLM.Complete(ctx, system, decidePrompt(c.GetQuestion(), c.GetOptions()), 400)
	if err != nil {
		s.log.Warn("decide llm error", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "llm: %v", err)
	}
	choice, reasoning, conf := parseDecision(raw, c.GetOptions())
	s.log.Info("decide", "host_id", hostID, "question", c.GetQuestion(),
		"choice", choice, "confidence", conf, "reasoning", reasoning)
	return &mesapb.Decision{Choice: choice, Reasoning: reasoning, Confidence: conf}, nil
}

// Act is the agent step (situation → DSL move), LLM-backed and persona-grounded.
func (s *Server) Act(ctx context.Context, sit *mesapb.Situation) (*mesapb.Move, error) {
	if s.actLLM == nil {
		return nil, status.Error(codes.Unavailable, "llm not configured")
	}
	return s.act(ctx, hostIDFromContext(ctx), sit)
}

// --- Knowledge -------------------------------------------------------------

// Recall is RAG over host memory + the curated wiki. Scaffold: empty for now.
func (s *Server) Recall(ctx context.Context, q *mesapb.Query) (*mesapb.KnowledgeSet, error) {
	s.log.Info("recall (stub)", "host_id", hostIDFromContext(ctx), "text", q.GetText(), "kind", q.GetKind())
	return &mesapb.KnowledgeSet{}, nil
}

// --- Journal ---------------------------------------------------------------

// Remember ingests the client-streamed episodes (Mirror). Scaffold: stored
// in-memory + logged; durable LTM + re-scoring lands with the crons.
func (s *Server) Remember(stream grpc.ClientStreamingServer[mesapb.Episode, mesapb.RememberAck]) error {
	hostID := hostIDFromContext(stream.Context()) // authenticated; all episodes attributed to it
	var accepted int64
	for {
		ep, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&mesapb.RememberAck{Accepted: accepted})
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.mem[hostID] = append(s.mem[hostID], ep)
		n := len(s.mem[hostID])
		s.mu.Unlock()
		accepted++
		s.log.Info("remember", "host_id", hostID, "kind", ep.GetKind(), "total", n)
	}
}

// --- Provision -------------------------------------------------------------

// Fetch returns the host's compiled persona/goals (unary initial sync). The
// persona crosses the wire as JSON; the host compiles the pearl table locally.
func (s *Server) Fetch(ctx context.Context, _ *mesapb.HostRef) (*mesapb.Provisioning, error) {
	hostID := hostIDFromContext(ctx) // authenticated; the request body is ignored
	e, ok := s.lookup(hostID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no persona registered for host_id %q", hostID)
	}
	pj, err := json.Marshal(e.persona)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal persona: %v", err)
	}
	s.log.Info("provision fetch", "host_id", hostID, "name", e.persona.Cornerstone.Identity.Name)
	return &mesapb.Provisioning{PersonaJson: pj, Prose: e.prose, Goals: e.goals}, nil
}

// Subscribe is the live push stream. Scaffold: it sends the current goals once
// (so the path is exercised end-to-end), then holds the stream open until the
// host disconnects. The revision crons will drive real pushes here.
func (s *Server) Subscribe(req *mesapb.SubscribeRequest, stream grpc.ServerStreamingServer[mesapb.Directive]) error {
	hostID := hostIDFromContext(stream.Context()) // authenticated identity
	s.log.Info("provision subscribe", "host_id", hostID, "last_applied", req.GetLastAppliedId())
	if e, ok := s.lookup(hostID); ok && len(e.goals) > 0 {
		if payload, err := json.Marshal(e.goals); err == nil {
			_ = stream.Send(&mesapb.Directive{Id: 1, Kind: mesapb.DirectiveKind_GOAL_REVISION, Payload: payload})
		}
	}
	<-stream.Context().Done()
	return nil
}

// --- prompts ----------------------------------------------------------------

const decideInstruction = "\n\nYou are roleplaying this character inside RuneScape Classic. " +
	"You will be given a decision and a list of options. Choose what THIS character would do. " +
	"Respond with ONLY a single-line JSON object and nothing else: " +
	`{"choice":"<one option copied verbatim>","reasoning":"<one short first-person sentence in your voice>","confidence":<number 0-1>}.`

const genericDecideSystem = "You are a RuneScape Classic player deciding your next action." + decideInstruction

func decideSystem(prose string) string { return strings.TrimSpace(prose) + decideInstruction }

func decidePrompt(question string, options []string) string {
	var b strings.Builder
	if q := strings.TrimSpace(question); q != "" {
		fmt.Fprintf(&b, "Decision: %s\n", q)
	} else {
		b.WriteString("Decision: what do you do?\n")
	}
	b.WriteString("Options:\n")
	for i, o := range options {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, o)
	}
	return b.String()
}

func goalsOf(p persona.Persona) []string {
	if s := strings.TrimSpace(p.Cornerstone.Identity.NorthStar.Statement); s != "" {
		return []string{s}
	}
	return nil
}

// parseDecision extracts the model's JSON verdict and snaps the choice to one of
// the offered options. Returns (choice, reasoning, confidence).
func parseDecision(raw string, options []string) (string, string, float64) {
	var v struct {
		Choice     string  `json:"choice"`
		Reasoning  string  `json:"reasoning"`
		Confidence float64 `json:"confidence"`
	}
	if js := extractJSON(raw); js != "" {
		_ = json.Unmarshal([]byte(js), &v)
	}
	choice := matchOption(v.Choice, options)
	conf := v.Confidence
	if conf <= 0 || conf > 1 {
		conf = 0.7
	}
	return choice, strings.TrimSpace(v.Reasoning), conf
}

// matchOption snaps the model's choice to one of the offered options: exact,
// then case-insensitive/substring; failing all, the first option.
func matchOption(choice string, options []string) string {
	if len(options) == 0 {
		return choice
	}
	c := strings.TrimSpace(choice)
	for _, o := range options {
		if o == c {
			return o
		}
	}
	lc := strings.ToLower(c)
	for _, o := range options {
		lo := strings.ToLower(o)
		if lo == lc || (lc != "" && (strings.Contains(lo, lc) || strings.Contains(lc, lo))) {
			return o
		}
	}
	return options[0]
}

// extractJSON returns the first {...} span in s (LLMs sometimes wrap it).
func extractJSON(s string) string {
	i := strings.IndexByte(s, '{')
	j := strings.LastIndexByte(s, '}')
	if i < 0 || j < i {
		return ""
	}
	return s[i : j+1]
}
