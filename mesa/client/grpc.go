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
