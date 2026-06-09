package mesaclient

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// GRPCClient is the host's Client over gRPC — the real transport (pure Go, no
// CGO). It translates the host-facing mesaclient Go types to/from the generated
// protobuf messages so the rest of the host stays transport-agnostic. One
// GRPCClient per host, authenticating as its host_id in every request.
type GRPCClient struct {
	conn *grpc.ClientConn
	game mesapb.GameClient
	know mesapb.KnowledgeClient
	jrnl mesapb.JournalClient
	prov mesapb.ProvisionClient
	kv   mesapb.KVClient
}

// compile-time proof the gRPC client satisfies the full Client surface.
var _ Client = (*GRPCClient)(nil)

// NewGRPCClient dials a mesa service at addr ("host:port") authenticating with
// the host's bearer token (sent on every RPC, including streams). The connection
// is lazy (grpc.NewClient connects on first RPC); transport is insecure (the
// host ↔ mesa link is trusted/local for now — TLS/mTLS lands with deployment).
func NewGRPCClient(addr, token string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(tokenCreds{token: token}),
	)
	if err != nil {
		return nil, fmt.Errorf("mesa dial %s: %w", addr, err)
	}
	return &GRPCClient{
		conn: conn,
		game: mesapb.NewGameClient(conn),
		know: mesapb.NewKnowledgeClient(conn),
		jrnl: mesapb.NewJournalClient(conn),
		prov: mesapb.NewProvisionClient(conn),
		kv:   mesapb.NewKVClient(conn),
	}, nil
}

// Close shuts the underlying connection.
func (c *GRPCClient) Close() error { return c.conn.Close() }

func (c *GRPCClient) ref(hostID string) *mesapb.HostRef { return &mesapb.HostRef{HostId: hostID} }

// Provision pulls the host's compiled persona/goals (unary Provision.Fetch).
func (c *GRPCClient) Provision(ctx context.Context, hostID string) (*Provisioning, error) {
	pb, err := c.prov.Fetch(ctx, c.ref(hostID))
	if err != nil {
		return nil, err
	}
	var p persona.Persona
	if len(pb.GetPersonaJson()) > 0 {
		if err := json.Unmarshal(pb.GetPersonaJson(), &p); err != nil {
			return nil, fmt.Errorf("mesa provision: decode persona: %w", err)
		}
	}
	return &Provisioning{Persona: p, Prose: pb.GetProse(), Goals: pb.GetGoals()}, nil
}

// Act ships a situation up (Game.Act) and gets a Move back.
func (c *GRPCClient) Act(ctx context.Context, s *Situation) (*Move, error) {
	pb, err := c.game.Act(ctx, situationToPB(s))
	if err != nil {
		return nil, err
	}
	return moveFromPB(pb), nil
}

// Decide resolves one in-routine option choice via mesa's LLM (Game.Decide).
func (c *GRPCClient) Decide(ctx context.Context, ch *Choice) (*Decision, error) {
	pb, err := c.game.Decide(ctx, &mesapb.Choice{
		Host:     c.ref(ch.HostID),
		Question: ch.Question,
		Options:  ch.Options,
		Affect:   affectToPB(ch.Affect),
	})
	if err != nil {
		return nil, err
	}
	return &Decision{
		Choice:          pb.GetChoice(),
		Reasoning:       pb.GetReasoning(),
		Confidence:      pb.GetConfidence(),
		CacheKey:        pb.GetCacheKey(),
		CacheTTLSeconds: pb.GetCacheTtlSeconds(),
	}, nil
}

// Chat resolves a fast social reply (Game.Chat) on the cheap tier.
func (c *GRPCClient) Chat(ctx context.Context, hostID, from, message string, recent []string) (string, bool, error) {
	r, err := c.game.Chat(ctx, &mesapb.ChatTurn{Host: c.ref(hostID), From: from, Message: message, Recent: recent})
	if err != nil {
		return "", false, err
	}
	return r.GetText(), r.GetSpeak(), nil
}

// Ask composes a PROACTIVE in-character question (Game.Chat, mode="ask") on the
// cheap tier — the intent-driven twin of Chat. The host has already decided WHEN
// + WHO deterministically; mesa only supplies the WORDS.
func (c *GRPCClient) Ask(ctx context.Context, hostID, target, question string, recent []string) (string, bool, error) {
	r, err := c.game.Chat(ctx, &mesapb.ChatTurn{
		Host: c.ref(hostID), From: target, Message: question, Topic: question, Mode: "ask", Recent: recent,
	})
	if err != nil {
		return "", false, err
	}
	return r.GetText(), r.GetSpeak(), nil
}

// AnalysisInterpret classifies an operator-override directive (Game.Analysis-
// Interpret). host_id rides the request context (auth), not the message — like
// every other Game RPC.
func (c *GRPCClient) AnalysisInterpret(ctx context.Context, directive string, state []string) (*AnalysisVerdict, error) {
	v, err := c.game.AnalysisInterpret(ctx, &mesapb.AnalysisDirective{Directive: directive, State: state})
	if err != nil {
		return nil, err
	}
	return &AnalysisVerdict{Kind: v.GetKind(), DSL: v.GetDsl(), Text: v.GetText()}, nil
}

// ExtractDialog runs the reactive-tier dialog extraction (Game.ExtractDialog).
// host_id rides the request context (auth), not the message — like every other
// Game RPC.
func (c *GRPCClient) ExtractDialog(ctx context.Context, hostID, speaker, role string, window []string, personaSnippet, activeGoal string, openQuestions []string) (*DialogExtraction, error) {
	pb, err := c.game.ExtractDialog(ctx, &mesapb.DialogWindow{
		Host:           c.ref(hostID),
		Speaker:        speaker,
		SpeakerRole:    role,
		Window:         window,
		PersonaSnippet: personaSnippet,
		ActiveGoal:     activeGoal,
		OpenQuestions:  openQuestions,
	})
	if err != nil {
		return nil, err
	}
	out := &DialogExtraction{}
	for _, c := range pb.GetClaims() {
		out.Claims = append(out.Claims, DialogClaim{
			Subject:    c.GetSubject(),
			Kind:       c.GetKind(),
			Claim:      c.GetClaim(),
			Confidence: c.GetConfidence(),
			Provenance: c.GetProvenance(),
		})
	}
	if in := pb.GetIntent(); in != nil {
		out.Intent = DialogIntent{Kind: in.GetKind(), Urgency: in.GetUrgency(), Gist: in.GetGist()}
	}
	return out, nil
}

// SyncRelationships pushes the host's full trust-ledger snapshot up (AuthLocal).
func (c *GRPCClient) SyncRelationships(ctx context.Context, hostID string, rels []Relationship) error {
	set := &mesapb.RelationshipSet{Host: c.ref(hostID)}
	for _, r := range rels {
		set.Relationships = append(set.Relationships, &mesapb.Relationship{
			Name:        r.Name,
			Alpha:       r.Alpha,
			Beta:        r.Beta,
			Encounters:  int32(r.Encounters),
			Tags:        r.Tags,
			ValueTraded: r.ValueTraded,
			Affinity:    r.Affinity,  // multi-axis (Phase 3b): carry the raw sums over the mirror
			Grievance:   r.Grievance, // else a mesa-bootstrap would drop the new axes
		})
	}
	_, err := c.jrnl.SyncRelationships(ctx, set)
	return err
}

// FetchRelationships pulls the host's stored trust ledger for cold-start bootstrap.
func (c *GRPCClient) FetchRelationships(ctx context.Context, hostID string) ([]Relationship, error) {
	pb, err := c.know.FetchRelationships(ctx, c.ref(hostID))
	if err != nil {
		return nil, err
	}
	var out []Relationship
	for _, r := range pb.GetRelationships() {
		out = append(out, Relationship{
			Name:        r.GetName(),
			Alpha:       r.GetAlpha(),
			Beta:        r.GetBeta(),
			Encounters:  int(r.GetEncounters()),
			Tags:        r.GetTags(),
			ValueTraded: r.GetValueTraded(),
			Affinity:    r.GetAffinity(),
			Grievance:   r.GetGrievance(),
		})
	}
	return out, nil
}

// SyncGoal mirrors the host's standing objective + progress up (structured goals).
func (c *GRPCClient) SyncGoal(ctx context.Context, hostID string, g Goal) error {
	_, err := c.jrnl.SyncGoal(ctx, &mesapb.Goal{
		Host:          c.ref(hostID),
		Objective:     g.Objective,
		Progress:      g.Progress,
		UpdatedAtUnix: g.UpdatedAt,
	})
	return err
}

// FetchGoal pulls the host's stored objective + progress for cold-start resume.
func (c *GRPCClient) FetchGoal(ctx context.Context, hostID string) (Goal, bool, error) {
	pb, err := c.know.FetchGoal(ctx, c.ref(hostID))
	if err != nil {
		return Goal{}, false, err
	}
	return Goal{
		Objective: pb.GetObjective(),
		Progress:  pb.GetProgress(),
		UpdatedAt: pb.GetUpdatedAtUnix(),
	}, pb.GetFound(), nil
}

// SyncKnowledge pushes the host's full world-knowledge snapshot up (Journal.
// SyncKnowledge). Shares the per-subject upsert with the consolidation cron.
func (c *GRPCClient) SyncKnowledge(ctx context.Context, hostID string, entries []KnowledgeEntry) error {
	led := &mesapb.KnowledgeLedger{Host: c.ref(hostID)}
	for _, e := range entries {
		pe := &mesapb.KnowledgeEntry{
			Subject:      e.Subject,
			Kind:         e.Kind,
			Encounters:   int32(e.Encounters),
			LastSeenUnix: e.LastSeenUnix,
			Tags:         e.Tags,
		}
		for _, b := range e.Beliefs {
			pe.Beliefs = append(pe.Beliefs, &mesapb.KnowledgeBelief{
				Claim: b.Claim, Provenance: b.Provenance, Alpha: b.Alpha, Beta: b.Beta, AtUnix: b.AtUnix,
			})
		}
		led.Entries = append(led.Entries, pe)
	}
	_, err := c.jrnl.SyncKnowledge(ctx, led)
	return err
}

// FetchKnowledge pulls the host's distilled world-knowledge ledger for a
// cold-start bootstrap (Knowledge.FetchKnowledge).
func (c *GRPCClient) FetchKnowledge(ctx context.Context, hostID string) ([]KnowledgeEntry, error) {
	pb, err := c.know.FetchKnowledge(ctx, c.ref(hostID))
	if err != nil {
		return nil, err
	}
	var out []KnowledgeEntry
	for _, e := range pb.GetEntries() {
		ke := KnowledgeEntry{
			Subject:      e.GetSubject(),
			Kind:         e.GetKind(),
			Encounters:   int(e.GetEncounters()),
			LastSeenUnix: e.GetLastSeenUnix(),
			Tags:         e.GetTags(),
		}
		for _, b := range e.GetBeliefs() {
			ke.Beliefs = append(ke.Beliefs, KnowledgeBelief{
				Claim: b.GetClaim(), Provenance: b.GetProvenance(),
				Alpha: b.GetAlpha(), Beta: b.GetBeta(), AtUnix: b.GetAtUnix(),
			})
		}
		out = append(out, ke)
	}
	return out, nil
}

// SyncGoalGraph pushes the host's full intention-graph snapshot up (Journal.
// SyncGoalGraph). Shares the per-host upsert with the insight cron.
func (c *GRPCClient) SyncGoalGraph(ctx context.Context, hostID string, snap GoalGraphSnapshot) error {
	pb := &mesapb.GoalGraphSnapshot{Host: c.ref(hostID)}
	for _, n := range snap.Nodes {
		pb.Nodes = append(pb.Nodes, &mesapb.GoalGraphNode{
			Id: n.ID, Kind: n.Kind, Label: n.Label, Status: n.Status,
			Progress: n.Progress, Tags: n.Tags, AtUnix: n.AtUnix,
		})
	}
	for _, e := range snap.Edges {
		pb.Edges = append(pb.Edges, &mesapb.GoalGraphEdge{From: e.From, To: e.To, Rel: e.Rel})
	}
	_, err := c.jrnl.SyncGoalGraph(ctx, pb)
	return err
}

// FetchGoalGraph pulls the host's distilled intention graph for a cold-start
// bootstrap (Knowledge.FetchGoalGraph).
func (c *GRPCClient) FetchGoalGraph(ctx context.Context, hostID string) (GoalGraphSnapshot, error) {
	pb, err := c.know.FetchGoalGraph(ctx, c.ref(hostID))
	if err != nil {
		return GoalGraphSnapshot{}, err
	}
	var out GoalGraphSnapshot
	for _, n := range pb.GetNodes() {
		out.Nodes = append(out.Nodes, GoalGraphNode{
			ID: n.GetId(), Kind: n.GetKind(), Label: n.GetLabel(), Status: n.GetStatus(),
			Progress: n.GetProgress(), Tags: n.GetTags(), AtUnix: n.GetAtUnix(),
		})
	}
	for _, e := range pb.GetEdges() {
		out.Edges = append(out.Edges, GoalGraphEdge{From: e.GetFrom(), To: e.GetTo(), Rel: e.GetRel()})
	}
	return out, nil
}

// Genesis runs the session-genesis compile (Provision.Genesis).
func (c *GRPCClient) Genesis(ctx context.Context, hostID, trigger, worldSummary string) (*GenesisResult, error) {
	pb, err := c.prov.Genesis(ctx, &mesapb.GenesisRequest{
		Host:         c.ref(hostID),
		Trigger:      trigger,
		WorldSummary: worldSummary,
	})
	if err != nil {
		return nil, err
	}
	res := &GenesisResult{Goal: pb.GetGoal(), Reasoning: pb.GetReasoning()}
	if m := pb.GetMood(); m != nil {
		res.Mood = Affect{Stress: m.GetStress(), Confidence: m.GetConfidence(), Valence: m.GetValence()}
	}
	for _, r := range pb.GetKeywordLadder() {
		res.KeywordLadder = append(res.KeywordLadder, KeywordRung{
			Keyword: r.GetKeyword(), Tier: r.GetTier(), Action: r.GetAction(),
		})
	}
	return res, nil
}

// ReportMetrics writes a host telemetry batch (Journal.ReportMetrics).
func (c *GRPCClient) ReportMetrics(ctx context.Context, hostID string, metrics []Metric) error {
	rep := &mesapb.MetricsReport{Host: c.ref(hostID)}
	for _, m := range metrics {
		rep.Metrics = append(rep.Metrics, &mesapb.Metric{Name: m.Name, Value: m.Value})
	}
	_, err := c.jrnl.ReportMetrics(ctx, rep)
	return err
}

// PutKV mirrors a host-namespaced blob to mesa (memory.Manager remote write).
func (c *GRPCClient) PutKV(ctx context.Context, hostID, key string, value []byte) error {
	_, err := c.kv.Put(ctx, &mesapb.KVPut{Host: c.ref(hostID), Key: key, Value: value})
	return err
}

// GetKV reads a host-namespaced blob (found=false on miss).
func (c *GRPCClient) GetKV(ctx context.Context, hostID, key string) ([]byte, bool, error) {
	v, err := c.kv.Get(ctx, &mesapb.KVKey{Host: c.ref(hostID), Key: key})
	if err != nil {
		return nil, false, err
	}
	return v.GetValue(), v.GetFound(), nil
}

// DeleteKV removes a host-namespaced blob.
func (c *GRPCClient) DeleteKV(ctx context.Context, hostID, key string) error {
	_, err := c.kv.Delete(ctx, &mesapb.KVKey{Host: c.ref(hostID), Key: key})
	return err
}

// Recall pulls game knowledge (Knowledge.Recall).
func (c *GRPCClient) Recall(ctx context.Context, q *Query) (*Knowledge, error) {
	pb, err := c.know.Recall(ctx, &mesapb.Query{
		Host: c.ref(q.HostID),
		Text: q.Text,
		Kind: mesapb.QueryKind(q.Kind),
		TopK: int32(q.TopK),
	})
	if err != nil {
		return nil, err
	}
	k := &Knowledge{}
	for _, it := range pb.GetItems() {
		k.Items = append(k.Items, KnowledgeItem{
			Kind:       QueryKind(it.GetKind()),
			Text:       it.GetText(),
			DSL:        it.GetDsl(),
			Provenance: it.GetProvenance(),
			Score:      it.GetScore(),
		})
	}
	return k, nil
}

// Remember mirrors one episode up (Journal.Remember client-stream: send one,
// close, await the ack).
func (c *GRPCClient) Remember(ctx context.Context, e *Episode) error {
	stream, err := c.jrnl.Remember(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(episodeToPB(e)); err != nil {
		return err
	}
	_, err = stream.CloseAndRecv()
	return err
}

// RecordObservation streams one raw perception up to mesa (the firehose).
func (c *GRPCClient) RecordObservation(ctx context.Context, o *Observation) error {
	stream, err := c.jrnl.RecordObservations(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(observationToPB(o)); err != nil {
		return err
	}
	_, err = stream.CloseAndRecv()
	return err
}

func observationToPB(o *Observation) *mesapb.Observation {
	return &mesapb.Observation{
		Host:           &mesapb.HostRef{HostId: o.HostID},
		IdempotencyKey: o.IdempotencyKey,
		Kind:           o.Kind,
		Subject:        o.Subject,
		Text:           o.Text,
		Salience:       o.Salience,
		OccurredAtUnix: o.OccurredAtUnix,
		Tags:           o.Tags,
	}
}

// Subscribe opens the Provision push stream (Provision.Subscribe) and fans
// directives onto a channel until the stream ends or ctx is cancelled.
func (c *GRPCClient) Subscribe(ctx context.Context, hostID string) (<-chan Directive, error) {
	stream, err := c.prov.Subscribe(ctx, &mesapb.SubscribeRequest{Host: c.ref(hostID)})
	if err != nil {
		return nil, err
	}
	ch := make(chan Directive, 8)
	go func() {
		defer close(ch)
		for {
			d, err := stream.Recv()
			if err != nil {
				return
			}
			select {
			case ch <- Directive{HostID: hostID, ID: d.GetId(), Kind: directiveKindFromPB(d.GetKind()), Payload: d.GetPayload()}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// Healthy reports whether the connection is usable. grpc.NewClient is lazy, so
// Idle (not yet dialed) counts as healthy-pending; only a shut/failed conn is unhealthy.
func (c *GRPCClient) Healthy() bool {
	switch c.conn.GetState() {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return true
	default:
		return false
	}
}

// --- translators ------------------------------------------------------------

func affectToPB(a Affect) *mesapb.Affect {
	return &mesapb.Affect{Stress: a.Stress, Confidence: a.Confidence, Valence: a.Valence}
}

func situationToPB(s *Situation) *mesapb.Situation {
	return &mesapb.Situation{
		Host:    &mesapb.HostRef{HostId: s.HostID},
		Goal:    s.Goal,
		Trigger: s.Trigger,
		Recent:  s.Recent,
		Affect:  affectToPB(s.Affect),
		Hints:   s.Hints,
		World: &mesapb.World{
			X: int32(s.World.X), Y: int32(s.World.Y), Region: s.World.Region,
			HpCur: int32(s.World.HPCur), HpMax: int32(s.World.HPMax),
			CombatLevel: int32(s.World.CombatLevel), Fatigue: s.World.Fatigue,
			InvFree: int32(s.World.InvFree), InvSummary: s.World.InvSummary,
			NearbyNpcs: s.World.NearbyNpcs, NearbyPlayers: s.World.NearbyPlayers,
		},
	}
}

func moveFromPB(m *mesapb.Move) *Move {
	return &Move{
		Kind:        MoveKind(m.GetKind()),
		Reasoning:   m.GetReasoning(),
		RoutinePath: m.GetRoutinePath(),
		RoutineName: m.GetRoutineName(),
		DSLSource:   m.GetDslSource(),
		Quarantined: m.GetQuarantined(),
		Args:        m.GetArgs(),
		Verb:        m.GetVerb(),
		ActionArgs:  m.GetActionArgs(),
		IdleSeconds: int(m.GetIdleSeconds()),

		GoalOp:       m.GetGoalOp(),
		GoalText:     m.GetGoalText(),
		GoalProgress: m.GetGoalProgress(),
	}
}

func episodeToPB(e *Episode) *mesapb.Episode {
	pb := &mesapb.Episode{
		Host:           &mesapb.HostRef{HostId: e.HostID},
		IdempotencyKey: e.IdempotencyKey,
		Kind:           e.Kind,
		Text:           e.Text,
		Importance:     e.Importance,
		OccurredAtUnix: e.OccurredAtUnix,
		Tags:           e.Tags,
	}
	if e.Relationship != nil {
		pb.Relation = &mesapb.RelationDelta{
			Name:             e.Relationship.Name,
			DAlpha:           e.Relationship.DAlpha,
			DBeta:            e.Relationship.DBeta,
			DEncounters:      int32(e.Relationship.DEncounters),
			TotalValueTraded: e.Relationship.TotalValueTraded,
			AddTags:          e.Relationship.AddTags,
		}
	}
	return pb
}

var pbDirectiveKind = map[mesapb.DirectiveKind]DirectiveKind{
	mesapb.DirectiveKind_ROUTINE_UPSERT:     DirectiveRoutineUpsert,
	mesapb.DirectiveKind_PEARL_REFRESH:      DirectivePearlRefresh,
	mesapb.DirectiveKind_PERSONA_REVISION:   DirectivePersonaRevision,
	mesapb.DirectiveKind_GOAL_REVISION:      DirectiveGoalRevision,
	mesapb.DirectiveKind_TRUST_DECAY:        DirectiveTrustDecay,
	mesapb.DirectiveKind_REVERIE_REBASELINE: DirectiveReverieRebaseline,
}

func directiveKindFromPB(k mesapb.DirectiveKind) DirectiveKind {
	if d, ok := pbDirectiveKind[k]; ok {
		return d
	}
	return DirectiveKind("unspecified")
}
