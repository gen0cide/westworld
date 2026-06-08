package mesad

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// validPersonaJSON builds a minimal valid persona (mirrors persona.validPersona,
// which is an unexported test helper) and marshals it for the Admin RPCs.
func validPersonaJSON(t *testing.T, name string) []byte {
	t.Helper()
	band := func(b persona.Band) persona.Trait { return persona.Trait{Band: b} }
	p := persona.Persona{
		SchemaVersion: persona.CurrentSchemaVersion,
		Cornerstone: persona.Cornerstone{
			Identity: persona.Identity{
				Name:         name,
				ArchetypeTag: "cautious_social_grinder",
				NorthStar:    persona.NorthStar{Theme: persona.ThemeSkillMastery, Statement: "99 fishing, never scammed", Horizon: "open"},
				Voice:        persona.Voice{Register: "swamp casual", Formality: persona.FormalityCasual, TypoFeel: persona.TypoOccasional},
			},
			Hexaco: map[string]persona.Trait{
				"H": band(persona.BandHigh), "E": band(persona.BandMidHigh), "X": band(persona.BandHigh),
				"A": band(persona.BandMidHigh), "C": band(persona.BandHigh), "O": band(persona.BandMid),
			},
			Values: persona.Values{NorthStarValue: persona.Achievement, SecondaryValue: persona.Benevolence},
			Prefs: persona.Prefs{
				Patience:         band(persona.BandVeryHigh),
				LossAversion:     persona.Trait{Mu: 2.6, Band: persona.BandHigh},
				CoopType:         persona.ConditionalCooperator,
				Risk:             persona.DomainRisk{Economic: persona.BandLow, Bodily: persona.BandLow, Social: persona.BandMid},
				Attention:        persona.AttentionAnchor{Anchor: 0.75, Level: persona.Focused},
				Curiosity:        persona.Curiosity{Skill: 0.6, Social: 0.25, Economic: 0.1, Spatial: 0.05},
				Aggression:       band(persona.BandLow),
				Decisiveness:     band(persona.BandMid),
				Tenacity:         band(persona.BandHigh),
				BulkApperception: band(persona.BandMid),
			},
			Gen: persona.GenerationMeta{CohortID: "lumbridge_regulars", Archetype: "cautious_social_grinder", SamplerVersion: "genpop-v1"},
		},
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal persona: %v", err)
	}
	return raw
}

// newAdminTestClient stands up an in-process mesad (no LTM, registry-only) over
// bufconn with the real auth interceptors, and returns an Admin client.
func newAdminTestClient(t *testing.T, adminToken string) mesapb.AdminClient {
	t.Helper()
	srv := New(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	srv.SetAdminToken(adminToken)

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer(grpc.UnaryInterceptor(srv.UnaryAuth), grpc.StreamInterceptor(srv.StreamAuth))
	srv.Attach(gs)
	go func() { _ = gs.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close(); gs.Stop(); lis.Close() })
	return mesapb.NewAdminClient(conn)
}

// withToken attaches a bearer token to the outgoing context.
func withToken(token string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
}

// putPersonas streams the given (host_id → json) personas through PutPersonas.
func putPersonas(t *testing.T, c mesapb.AdminClient, ctx context.Context, items []*mesapb.PersonaUpsert) *mesapb.BatchResult {
	t.Helper()
	stream, err := c.PutPersonas(ctx)
	if err != nil {
		t.Fatalf("PutPersonas open: %v", err)
	}
	for _, it := range items {
		if err := stream.Send(it); err != nil {
			t.Fatalf("PutPersonas send: %v", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("PutPersonas close: %v", err)
	}
	return res
}

func TestAdminPersonaCRUD(t *testing.T) {
	const tok = "s3cr3t"
	c := newAdminTestClient(t, tok)
	ctx := withToken(tok)

	// Bulk put: 2 valid + 1 invalid (bad JSON). The bad one must not fail the batch.
	res := putPersonas(t, c, ctx, []*mesapb.PersonaUpsert{
		{HostId: "drone1", PersonaJson: validPersonaJSON(t, "drone1")},
		{HostId: "drone2", PersonaJson: validPersonaJSON(t, "drone2")},
		{HostId: "broken", PersonaJson: []byte("{not json")},
	})
	if res.Ok != 2 || res.Failed != 1 {
		t.Fatalf("batch result ok=%d failed=%d, want 2/1 (items=%+v)", res.Ok, res.Failed, res.Items)
	}
	var brokenErr string
	for _, it := range res.Items {
		if it.HostId == "broken" {
			if it.Ok {
				t.Fatal("broken persona reported ok")
			}
			brokenErr = it.Error
		}
	}
	if brokenErr == "" {
		t.Fatal("broken persona has no error message")
	}

	// Get one back.
	rec, err := c.GetPersona(ctx, &mesapb.HostRef{HostId: "drone1"})
	if err != nil {
		t.Fatalf("GetPersona: %v", err)
	}
	if rec.Name != "drone1" || len(rec.PersonaJson) == 0 {
		t.Fatalf("GetPersona drone1: name=%q json=%dB", rec.Name, len(rec.PersonaJson))
	}

	// List: exactly the 2 that registered, sorted, metadata-only (no JSON).
	list, err := c.ListPersonas(ctx, &mesapb.ListPersonasRequest{})
	if err != nil {
		t.Fatalf("ListPersonas: %v", err)
	}
	if len(list.Personas) != 2 || list.Personas[0].HostId != "drone1" || list.Personas[1].HostId != "drone2" {
		t.Fatalf("list = %+v, want [drone1 drone2]", list.Personas)
	}
	if len(list.Personas[0].PersonaJson) != 0 {
		t.Fatal("list without with_json should omit persona_json")
	}

	// Delete one → gone; the other remains.
	if _, err := c.DeletePersona(ctx, &mesapb.HostRef{HostId: "drone1"}); err != nil {
		t.Fatalf("DeletePersona: %v", err)
	}
	if _, err := c.GetPersona(ctx, &mesapb.HostRef{HostId: "drone1"}); status.Code(err) != codes.NotFound {
		t.Fatalf("GetPersona after delete: code=%v, want NotFound", status.Code(err))
	}
	list, _ = c.ListPersonas(ctx, &mesapb.ListPersonasRequest{})
	if len(list.Personas) != 1 || list.Personas[0].HostId != "drone2" {
		t.Fatalf("after delete list=%+v, want [drone2]", list.Personas)
	}
}

func TestAdminAuth(t *testing.T) {
	const tok = "s3cr3t"
	c := newAdminTestClient(t, tok)

	// No token → Unauthenticated.
	if _, err := c.GetPersona(context.Background(), &mesapb.HostRef{HostId: "x"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("no token: code=%v, want Unauthenticated", status.Code(err))
	}
	// Wrong token → Unauthenticated.
	if _, err := c.GetPersona(withToken("nope"), &mesapb.HostRef{HostId: "x"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("wrong token: code=%v, want Unauthenticated", status.Code(err))
	}
	// Correct token → NotFound (auth passed, host just not registered).
	if _, err := c.GetPersona(withToken(tok), &mesapb.HostRef{HostId: "x"}); status.Code(err) != codes.NotFound {
		t.Fatalf("good token: code=%v, want NotFound", status.Code(err))
	}
}

func TestAdminDisabledWithoutToken(t *testing.T) {
	// adminToken unset → the Admin API is closed even with a token presented.
	c := newAdminTestClient(t, "")
	if _, err := c.GetPersona(withToken("anything"), &mesapb.HostRef{HostId: "x"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("admin disabled: code=%v, want Unauthenticated", status.Code(err))
	}
}
