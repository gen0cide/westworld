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
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	mesapb.UnimplementedAdminServer

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

	// catalog holds the world name-sets (item/npc/place/POI) for static
	// validation of hallucinated literal args in authored DSL. Nil/unloaded
	// when -facts is unset or load failed — validation then degrades to a
	// no-op (never blocks). Wired by the mesad binary via SetCatalog.
	catalog *argCatalog

	// adminToken gates the operator-only Admin service (persona CRUD via
	// mesa-ctl), kept separate from per-host bearer tokens. Empty => the Admin
	// API is DISABLED (every Admin call is rejected). Wired from $ADMIN_TOKEN.
	adminToken string

	// subs is the live Provision.Subscribe push registry: host_id → the buffered
	// channel its open Subscribe stream is draining. Admin.PushGoal fans a
	// GOAL_REVISION out across it with no restart. Guarded by subsMu (NOT s.mu —
	// a Subscribe call blocks for the whole connection lifetime, so its registry
	// touch must not hold the persona lock). dirSeq hands out monotonic directive
	// ids so a host's last_applied bookkeeping stays ordered.
	subsMu sync.Mutex
	subs   map[string]chan *mesapb.Directive
	dirSeq atomic.Int64

	// Phase-4 cron lifecycle. cronCancel stops the background distillation
	// loops; cronWG waits for in-flight folds to finish on shutdown (no torn
	// writes); cronSem bounds the number of in-flight per-host LLM jobs so the
	// crons never starve the host-facing RPCs (Act/Chat/Decide). All nil/zero
	// until StartCrons; StopCrons is safe to call when crons never started.
	cronCancel context.CancelFunc
	cronWG     sync.WaitGroup
	cronSem    chan struct{}

	// consolidateLLMOverride, when non-nil, replaces decideLLM as the Tier-1
	// consolidation seam — TEST-ONLY (inject a fake completer; no real API calls).
	// Production leaves it nil so the bulk cron unconditionally uses Haiku.
	consolidateLLMOverride completer
}

// SetCatalog attaches the world name-set catalog used to statically reject
// hallucinated literal args in Act moves. Call once at startup. A nil or
// empty catalog disables the extra validation (graceful degradation).
func (s *Server) SetCatalog(c *argCatalog) { s.catalog = c }

// SetLTM attaches the durable long-term-memory store. Call once at startup.
func (s *Server) SetLTM(l *LTM) { s.ltm = l }

// SetAdminToken enables the operator-only Admin service (persona CRUD) and sets
// the credential mesa-ctl must present. Call once at startup. An empty token
// leaves the Admin API disabled (every Admin call returns Unauthenticated).
func (s *Server) SetAdminToken(token string) {
	s.mu.Lock()
	s.adminToken = token
	s.mu.Unlock()
}

// entry is a registered host's compiled identity, held mesa-side.
type entry struct {
	persona persona.Persona
	prose   string // rendered identity card (cook output / Render floor)
	system  string // the decide/act system prompt derived from prose
	goals   []string
	// goalPushed marks goals as an OPERATOR override (Admin.PushGoal) rather than
	// the persona baseline. Only a pushed goal is re-sent over a (re)connecting
	// Subscribe stream, so the connect-time send never clobbers a genesis goal the
	// host already holds — the host adopts a live goal only when an operator means it.
	goalPushed bool
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
		reg:        map[string]*entry{},
		mem:        map[string][]*mesapb.Episode{},
		tokens:     map[string]string{},
		subs:       map[string]chan *mesapb.Directive{},
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
		// Persist the derived prose card alongside the JSON so it's a first-class
		// column (no re-derivation on read); cooked tracks whether it came from the
		// LLM cook vs the Render() floor.
		prose := persona.Render(&p)
		if err := s.ltm.UpsertPersona(context.Background(), hostID, data, prose, p.Cornerstone.Gen.LLMMaterialized); err != nil {
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
	mesapb.RegisterAdminServer(gs, s)
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

// RecordObservations ingests the client-streamed perception firehose. Stored in
// the observations table when LTM is wired (cron fodder, distinct from episode
// recall); otherwise counted and dropped (not retained in the in-mem scaffold).
func (s *Server) RecordObservations(stream grpc.ClientStreamingServer[mesapb.Observation, mesapb.RememberAck]) error {
	hostID := hostIDFromContext(stream.Context())
	var accepted, deduped int64
	for {
		o, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&mesapb.RememberAck{Accepted: accepted, Deduped: deduped})
		}
		if err != nil {
			return err
		}
		if s.ltm != nil {
			dup, aerr := s.ltm.AddObservation(stream.Context(), hostID, o)
			if aerr != nil {
				s.log.Warn("observation: ltm add failed", "host_id", hostID, "err", aerr)
				return status.Errorf(codes.Internal, "ltm add observation: %v", aerr)
			}
			if dup {
				deduped++
			} else {
				accepted++
			}
			continue
		}
		accepted++ // no durable store: count + drop (firehose not retained in-mem)
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

// SyncKnowledge mirrors the host's world-knowledge snapshot into Postgres
// (host→mesa). Shares the knowledge table with the consolidation cron — both
// upsert by (host_id,subject), last-writer-wins per subject. Best-effort
// durability; graceful empty when no LTM is wired (matches the trust mirror).
func (s *Server) SyncKnowledge(ctx context.Context, led *mesapb.KnowledgeLedger) (*mesapb.SyncAck, error) {
	hostID := hostIDFromContext(ctx) // authenticated identity, never the request body
	if s.ltm == nil {
		return &mesapb.SyncAck{}, nil
	}
	n, err := s.ltm.SyncKnowledge(ctx, hostID, led.GetEntries())
	if err != nil {
		s.log.Warn("sync knowledge failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "ltm sync knowledge: %v", err)
	}
	s.log.Info("knowledge synced", "host_id", hostID, "count", n)
	return &mesapb.SyncAck{Stored: int64(n)}, nil
}

// FetchKnowledge returns the host's distilled world-knowledge ledger for a
// cold-start bootstrap (mesa→host) — the missing knowledge analogue of
// FetchRelationships, so a restarted/cold host warm-starts beliefs the cron
// distilled that it never explicitly wrote. Empty when no LTM is wired.
func (s *Server) FetchKnowledge(ctx context.Context, _ *mesapb.HostRef) (*mesapb.KnowledgeLedger, error) {
	hostID := hostIDFromContext(ctx)
	if s.ltm == nil {
		return &mesapb.KnowledgeLedger{}, nil
	}
	entries, err := s.ltm.Knowledge(ctx, hostID)
	if err != nil {
		s.log.Warn("fetch knowledge failed", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "ltm knowledge: %v", err)
	}
	s.log.Info("knowledge fetched", "host_id", hostID, "count", len(entries))
	return &mesapb.KnowledgeLedger{Entries: entries}, nil
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

	ch := make(chan *mesapb.Directive, 8)
	s.registerSub(hostID, ch)
	defer s.unregisterSub(hostID, ch)

	// Re-send a standing OPERATOR goal so the override survives a reconnect. The
	// persona baseline is NOT re-sent (the host already provisioned it and may hold
	// a richer genesis goal) — only an Admin.PushGoal makes goalPushed true.
	if e, ok := s.lookup(hostID); ok && e.goalPushed && len(e.goals) > 0 {
		if payload, err := json.Marshal(e.goals); err == nil {
			_ = stream.Send(&mesapb.Directive{Id: s.nextDirID(), Kind: mesapb.DirectiveKind_GOAL_REVISION, Payload: payload})
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case d := <-ch:
			if err := stream.Send(d); err != nil {
				return err
			}
		}
	}
}

// nextDirID hands out the next monotonic directive id (kept ordered so a host's
// last_applied bookkeeping is meaningful across pushes within a mesad lifetime).
func (s *Server) nextDirID() int64 { return s.dirSeq.Add(1) }

// registerSub records a host's live push channel, replacing any prior one (a
// reconnect supersedes the stale stream; the old goroutine exits on its own
// ctx.Done and its deferred unregister no-ops via the identity check below).
func (s *Server) registerSub(hostID string, ch chan *mesapb.Directive) {
	s.subsMu.Lock()
	s.subs[hostID] = ch
	s.subsMu.Unlock()
}

// unregisterSub removes a host's push channel only if it is still THIS channel,
// so a superseded older stream tearing down doesn't evict the newer one.
func (s *Server) unregisterSub(hostID string, ch chan *mesapb.Directive) {
	s.subsMu.Lock()
	if s.subs[hostID] == ch {
		delete(s.subs, hostID)
	}
	s.subsMu.Unlock()
}

// PushGoal sets a soft runtime goal override on every registered host matching
// the glob (empty = all), marking it operator-pushed so it sticks across a
// reconnect, then delivers a GOAL_REVISION to those currently subscribed. A slow
// subscriber whose buffer is full is logged + skipped rather than blocking the
// fan-out. Returns how many matched (goal set) vs. were pushed to live.
func (s *Server) PushGoal(ctx context.Context, req *mesapb.PushGoalRequest) (*mesapb.PushGoalResult, error) {
	goal := strings.TrimSpace(req.GetGoal())
	if goal == "" {
		return nil, status.Error(codes.InvalidArgument, "empty goal")
	}
	match := req.GetMatch()
	if match != "" {
		if _, err := filepath.Match(match, ""); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "bad match pattern %q: %v", match, err)
		}
	}
	payload, err := json.Marshal([]string{goal})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal goal: %v", err)
	}

	// Set the goal on the registry (sticky for reconnects + future provisions).
	s.mu.Lock()
	matched := make([]string, 0, len(s.reg))
	for id, e := range s.reg {
		if match != "" {
			if ok, _ := filepath.Match(match, id); !ok {
				continue
			}
		}
		e.goals = []string{goal}
		e.goalPushed = true
		matched = append(matched, id)
	}
	s.mu.Unlock()

	// Deliver live to those with an open Subscribe stream.
	var pushed []string
	for _, id := range matched {
		s.subsMu.Lock()
		ch, ok := s.subs[id]
		s.subsMu.Unlock()
		if !ok {
			continue
		}
		d := &mesapb.Directive{Id: s.nextDirID(), Kind: mesapb.DirectiveKind_GOAL_REVISION, Payload: payload}
		select {
		case ch <- d:
			pushed = append(pushed, id)
		default:
			s.log.Warn("admin: goal push dropped (slow subscriber)", "host_id", id)
		}
	}
	sort.Strings(pushed)
	s.log.Info("admin: pushed goal", "match", match, "matched", len(matched), "pushed", len(pushed))
	return &mesapb.PushGoalResult{Pushed: int32(len(pushed)), HostIds: pushed, Matched: int32(len(matched))}, nil
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
