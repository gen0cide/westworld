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
	"time"

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
	mesapb.UnimplementedKVServer

	actLLM     *llm.Client // Act/DSL authoring tier (Sonnet/Haiku) — high volume, cached prefix
	decideLLM  *llm.Client // narrow option-pick tier (Haiku)
	genesisLLM *llm.Client // session-genesis tier (Opus) — rare, history-rich login compile
	log        *slog.Logger

	mu     sync.RWMutex
	reg    map[string]*entry            // host_id → its persona/prose/system prompt
	mem    map[string][]*mesapb.Episode // host_id → mirrored episodes (in-mem fallback when no LTM)
	tokens map[string]string            // bearer token → host_id (auth)

	// ltm is the durable long-term episodic memory store (bbolt). When set,
	// Remember persists+dedups into it and Recall retrieves from it; when nil,
	// Remember falls back to the volatile mem map and Recall is empty. Wired by
	// the mesad binary from a -data-dir.
	ltm *LTM
}

// SetLTM attaches the durable long-term-memory store. Call once at startup.
func (s *Server) SetLTM(l *LTM) { s.ltm = l }

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
func New(actLLM, decideLLM, genesisLLM *llm.Client, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		actLLM:     actLLM,
		decideLLM:  decideLLM,
		genesisLLM: genesisLLM,
		log:        log,
		reg:       map[string]*entry{},
		mem:       map[string][]*mesapb.Episode{},
		tokens:    map[string]string{},
	}
}

// Register stores a persona for a host_id, rendering its prose card + system
// prompt once. It validates the persona first.
func (s *Server) Register(hostID string, p persona.Persona) error {
	if err := s.registerLocal(hostID, p); err != nil {
		return err
	}
	// Persist to Postgres so the persona survives a mesad restart without the
	// -host file, and mesa can bootstrap a host's identity from nothing.
	if s.ltm != nil {
		data, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("register %s: marshal: %w", hostID, err)
		}
		if err := s.ltm.UpsertPersona(context.Background(), hostID, data); err != nil {
			return fmt.Errorf("register %s: persist: %w", hostID, err)
		}
	}
	return nil
}

// registerLocal validates + renders a persona into the in-memory registry (and
// authorizes the host), WITHOUT persisting — used both by Register (which then
// persists) and by LoadPersonas (restoring already-persisted personas).
func (s *Server) registerLocal(hostID string, p persona.Persona) error {
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

// LoadPersonas restores all personas from Postgres into the in-memory registry.
// Called at startup so a host's identity survives a mesad restart without the
// -host file. -host flags processed afterward add or override.
func (s *Server) LoadPersonas(ctx context.Context) error {
	if s.ltm == nil {
		return nil
	}
	rows, err := s.ltm.Personas(ctx)
	if err != nil {
		return err
	}
	n := 0
	for _, row := range rows {
		var p persona.Persona
		if err := json.Unmarshal(row.JSON, &p); err != nil {
			s.log.Warn("load persona: bad json", "host_id", row.HostID, "err", err)
			continue
		}
		if err := s.registerLocal(row.HostID, p); err != nil {
			s.log.Warn("load persona: register", "host_id", row.HostID, "err", err)
			continue
		}
		n++
	}
	if n > 0 {
		s.log.Info("loaded personas from postgres", "count", n)
	}
	return nil
}

// Register wires the server onto a gRPC server as all four services.
func (s *Server) Attach(gs *grpc.Server) {
	mesapb.RegisterGameServer(gs, s)
	mesapb.RegisterKnowledgeServer(gs, s)
	mesapb.RegisterJournalServer(gs, s)
	mesapb.RegisterProvisionServer(gs, s)
	mesapb.RegisterKVServer(gs, s)
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

// Recall is RAG over host memory (and, later, the curated wiki). It returns the
// host's own past episodes most relevant to the query, ranked by the LTM store.
// Empty (not an error) when no durable store is wired — the host degrades to its
// local journal.
func (s *Server) Recall(ctx context.Context, q *mesapb.Query) (*mesapb.KnowledgeSet, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		s.log.Info("recall (no ltm)", "host_id", hostID, "text", q.GetText(), "kind", q.GetKind())
		return &mesapb.KnowledgeSet{}, nil
	}
	eps, err := s.ltm.Recall(ctx, hostID, q.GetText(), int(q.GetTopK()))
	if err != nil {
		s.log.Warn("recall failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "ltm recall: %v", err)
	}
	items := make([]*mesapb.KnowledgeItem, 0, len(eps))
	for _, e := range eps {
		items = append(items, &mesapb.KnowledgeItem{
			Kind:       mesapb.QueryKind_EPISODIC,
			Text:       e.Text,
			Provenance: e.Kind + "@" + time.Unix(e.At, 0).UTC().Format(time.RFC3339),
			Score:      e.Score,
		})
	}
	s.log.Info("recall", "host_id", hostID, "text", q.GetText(), "hits", len(items))
	return &mesapb.KnowledgeSet{Items: items}, nil
}

// --- Journal ---------------------------------------------------------------

// Remember ingests the client-streamed episodes (Mirror). Scaffold: stored
// in-memory + logged; durable LTM + re-scoring lands with the crons.
func (s *Server) Remember(stream grpc.ClientStreamingServer[mesapb.Episode, mesapb.RememberAck]) error {
	hostID := hostIDFromContext(stream.Context()) // authenticated; all episodes attributed to it
	var accepted, deduped int64
	for {
		ep, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&mesapb.RememberAck{Accepted: accepted, Deduped: deduped})
		}
		if err != nil {
			return err
		}
		if s.ltm != nil {
			dup, aerr := s.ltm.Add(stream.Context(), hostID, ep)
			if aerr != nil {
				s.log.Warn("remember: ltm add failed", "host_id", hostID, "err", aerr)
				return status.Errorf(codes.Internal, "ltm add: %v", aerr)
			}
			if dup {
				deduped++
			} else {
				accepted++
			}
			s.log.Info("remember", "host_id", hostID, "kind", ep.GetKind(), "dup", dup)
			continue
		}
		// No durable store wired: volatile in-memory scaffold fallback.
		s.mu.Lock()
		s.mem[hostID] = append(s.mem[hostID], ep)
		n := len(s.mem[hostID])
		s.mu.Unlock()
		accepted++
		s.log.Info("remember (in-mem)", "host_id", hostID, "kind", ep.GetKind(), "total", n)
	}
}

// SyncRelationships mirrors the host's trust-ledger snapshot into Postgres
// (host→mesa, AuthLocal). Best-effort durability; the host stays authoritative.
func (s *Server) SyncRelationships(ctx context.Context, set *mesapb.RelationshipSet) (*mesapb.SyncAck, error) {
	hostID := hostIDFromContext(ctx) // authenticated identity, never the request body
	if s.ltm == nil {
		return &mesapb.SyncAck{}, nil
	}
	n, err := s.ltm.SyncRelationships(ctx, hostID, set.GetRelationships())
	if err != nil {
		s.log.Warn("sync relationships failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "ltm sync: %v", err)
	}
	s.log.Info("relationships synced", "host_id", hostID, "count", n)
	return &mesapb.SyncAck{Stored: int64(n)}, nil
}

// FetchRelationships returns the host's stored trust ledger for a cold-start
// bootstrap (mesa→host). Empty when no durable store is wired.
func (s *Server) FetchRelationships(ctx context.Context, _ *mesapb.HostRef) (*mesapb.RelationshipSet, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.RelationshipSet{}, nil
	}
	rels, err := s.ltm.Relationships(ctx, hostID)
	if err != nil {
		s.log.Warn("fetch relationships failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "ltm relationships: %v", err)
	}
	s.log.Info("relationships fetched", "host_id", hostID, "count", len(rels))
	return &mesapb.RelationshipSet{Relationships: rels}, nil
}

// SyncGoal mirrors the host's standing objective + progress into the structured
// goals table (host→mesa). AuthLocal — the host owns its current plan.
func (s *Server) SyncGoal(ctx context.Context, g *mesapb.Goal) (*mesapb.SyncAck, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.SyncAck{}, nil
	}
	if err := s.ltm.UpsertGoal(ctx, hostID, g.GetObjective(), g.GetProgress()); err != nil {
		s.log.Warn("sync goal failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "goal upsert: %v", err)
	}
	s.log.Info("goal synced", "host_id", hostID, "objective", g.GetObjective())
	return &mesapb.SyncAck{Stored: 1}, nil
}

// FetchGoal returns the host's stored objective + progress for cold-start resume.
func (s *Server) FetchGoal(ctx context.Context, _ *mesapb.HostRef) (*mesapb.Goal, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.Goal{}, nil
	}
	objective, progress, found, err := s.ltm.Goal(ctx, hostID)
	if err != nil {
		s.log.Warn("fetch goal failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "goal fetch: %v", err)
	}
	return &mesapb.Goal{Objective: objective, Progress: progress, Found: found}, nil
}

// ReportMetrics records a host's telemetry batch (host→mesa, append-only series).
func (s *Server) ReportMetrics(ctx context.Context, r *mesapb.MetricsReport) (*mesapb.SyncAck, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.SyncAck{}, nil
	}
	at := time.Now()
	if u := r.GetAtUnix(); u != 0 {
		at = time.Unix(u, 0)
	}
	n, err := s.ltm.RecordMetrics(ctx, hostID, r.GetMetrics(), at)
	if err != nil {
		s.log.Warn("report metrics failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "metrics: %v", err)
	}
	s.log.Info("metrics reported", "host_id", hostID, "samples", n)
	return &mesapb.SyncAck{Stored: int64(n)}, nil
}

// --- KV (generic host-state mirror) ----------------------------------------

// Put stores a host-namespaced KV blob (memory.Manager remote-tier write).
func (s *Server) Put(ctx context.Context, p *mesapb.KVPut) (*mesapb.KVAck, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.KVAck{}, nil
	}
	if err := s.ltm.PutKV(ctx, hostID, p.GetKey(), p.GetValue()); err != nil {
		s.log.Warn("kv put failed", "host_id", hostID, "key", p.GetKey(), "err", err)
		return nil, status.Errorf(codes.Internal, "kv put: %v", err)
	}
	return &mesapb.KVAck{}, nil
}

// Get reads a host-namespaced KV blob (found=false on miss, not an error).
func (s *Server) Get(ctx context.Context, k *mesapb.KVKey) (*mesapb.KVValue, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.KVValue{}, nil
	}
	v, found, err := s.ltm.GetKV(ctx, hostID, k.GetKey())
	if err != nil {
		s.log.Warn("kv get failed", "host_id", hostID, "key", k.GetKey(), "err", err)
		return nil, status.Errorf(codes.Internal, "kv get: %v", err)
	}
	return &mesapb.KVValue{Value: v, Found: found}, nil
}

// Delete removes a host-namespaced KV blob.
func (s *Server) Delete(ctx context.Context, k *mesapb.KVKey) (*mesapb.KVAck, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.KVAck{}, nil
	}
	if err := s.ltm.DeleteKV(ctx, hostID, k.GetKey()); err != nil {
		return nil, status.Errorf(codes.Internal, "kv delete: %v", err)
	}
	return &mesapb.KVAck{}, nil
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
