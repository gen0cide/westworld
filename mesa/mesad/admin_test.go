package mesad

import (
	"context"
	"encoding/json"
	"github.com/gen0cide/westworld/mesa/auth"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

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

// newAdminTestServer stands up an in-process mesad (no LTM, registry-only) over
// bufconn with the real auth interceptors, and returns the server + a dialed
// client conn (so a test can build Admin AND Provision clients on it).
func newAdminTestServer(t *testing.T, adminToken string) (*Server, *grpc.ClientConn) {
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
	return srv, conn
}

// newAdminTestClient returns just an Admin client over a fresh test server.
func newAdminTestClient(t *testing.T, adminToken string) mesapb.AdminClient {
	t.Helper()
	_, conn := newAdminTestServer(t, adminToken)
	return mesapb.NewAdminClient(conn)
}

// waitForSub blocks until host_id has an active Provision.Subscribe channel in
// the registry (the subscribe-vs-publish race: the client stream returns before
// the server handler runs registerSub).
func waitForSub(t *testing.T, srv *Server, hostID string) {
	t.Helper()
	for i := 0; i < 200; i++ {
		srv.subsMu.Lock()
		_, ok := srv.subs[hostID]
		srv.subsMu.Unlock()
		if ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("subscriber %q never registered", hostID)
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

// recvGoal reads one directive off a Subscribe stream and asserts it is a
// GOAL_REVISION carrying exactly the expected goal.
func recvGoal(t *testing.T, sub grpc.ServerStreamingClient[mesapb.Directive], want string) {
	t.Helper()
	d, err := sub.Recv()
	if err != nil {
		t.Fatalf("Subscribe recv: %v", err)
	}
	if d.Kind != mesapb.DirectiveKind_GOAL_REVISION {
		t.Fatalf("directive kind = %v, want GOAL_REVISION", d.Kind)
	}
	var goals []string
	if err := json.Unmarshal(d.Payload, &goals); err != nil {
		t.Fatalf("goal payload: %v", err)
	}
	if len(goals) != 1 || goals[0] != want {
		t.Fatalf("goal payload = %v, want [%q]", goals, want)
	}
}

func TestAdminPushGoal(t *testing.T) {
	const tok = "s3cr3t"
	srv, conn := newAdminTestServer(t, tok)
	admin := mesapb.NewAdminClient(conn)
	prov := mesapb.NewProvisionClient(conn)
	adminCtx := withToken(tok)

	// Register three hosts. On Register the server authorizes each with the
	// deterministic HostKey, so a host subscribes with that as its bearer token.
	putPersonas(t, admin, adminCtx, []*mesapb.PersonaUpsert{
		{HostId: "drone1", PersonaJson: validPersonaJSON(t, "drone1")},
		{HostId: "drone2", PersonaJson: validPersonaJSON(t, "drone2")},
		{HostId: "lonely", PersonaJson: validPersonaJSON(t, "lonely")},
	})

	// No baseline on connect: a never-pushed host's fresh Subscribe must NOT
	// receive a GOAL_REVISION (the genesis/persona goal it already holds wins).
	lonelyCtx, cancelLonely := context.WithCancel(withToken(auth.HostKey("lonely")))
	lonelySub, err := prov.Subscribe(lonelyCtx, &mesapb.SubscribeRequest{})
	if err != nil {
		t.Fatalf("lonely Subscribe: %v", err)
	}
	waitForSub(t, srv, "lonely")
	gotc := make(chan *mesapb.Directive, 1)
	errc := make(chan error, 1)
	go func() {
		if d, err := lonelySub.Recv(); err != nil {
			errc <- err
		} else {
			gotc <- d
		}
	}()
	select {
	case d := <-gotc:
		t.Fatalf("unexpected baseline directive to never-pushed host: %+v", d)
	case err := <-errc:
		t.Fatalf("lonely recv error: %v", err)
	case <-time.After(300 * time.Millisecond): // good — nothing sent
	}
	cancelLonely()

	// drone1 subscribes; drone2 does not.
	droneCtx, cancelDrone := context.WithCancel(withToken(auth.HostKey("drone1")))
	defer cancelDrone()
	sub, err := prov.Subscribe(droneCtx, &mesapb.SubscribeRequest{})
	if err != nil {
		t.Fatalf("drone1 Subscribe: %v", err)
	}
	waitForSub(t, srv, "drone1")

	// Push to drone* : sets the goal on drone1+drone2 (lonely excluded by the
	// glob), but only drone1 is live so only it is delivered to.
	res, err := admin.PushGoal(adminCtx, &mesapb.PushGoalRequest{Goal: "roam Varrock", Match: "drone*"})
	if err != nil {
		t.Fatalf("PushGoal: %v", err)
	}
	if res.Matched != 2 {
		t.Fatalf("matched=%d, want 2 (drone1,drone2)", res.Matched)
	}
	if res.Pushed != 1 || len(res.HostIds) != 1 || res.HostIds[0] != "drone1" {
		t.Fatalf("pushed=%d ids=%v, want 1 [drone1]", res.Pushed, res.HostIds)
	}
	recvGoal(t, sub, "roam Varrock")

	// Reconnect stickiness: a pushed goal IS re-sent on a fresh connect, so a
	// host that drops and reconnects re-adopts the override.
	reCtx, cancelRe := context.WithCancel(withToken(auth.HostKey("drone1")))
	defer cancelRe()
	reSub, err := prov.Subscribe(reCtx, &mesapb.SubscribeRequest{})
	if err != nil {
		t.Fatalf("drone1 re-Subscribe: %v", err)
	}
	recvGoal(t, reSub, "roam Varrock")
}

// TestShutdownDrainsSubscribe is the H14 regression: a live Subscribe stream
// must NOT block a graceful gRPC drain. Pre-fix, Subscribe's select returned only
// on client disconnect, so GracefulStop (which waits for every in-flight stream)
// hung forever. With Server.Shutdown() closing s.shutdown, the server ends its own
// stream and GracefulStop completes.
func TestShutdownDrainsSubscribe(t *testing.T) {
	srv := New(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	srv.SetAdminToken("s3cr3t")
	if err := srv.registerLocal("drone1", mustPersona(t, "drone1")); err != nil {
		t.Fatalf("registerLocal: %v", err)
	}

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer(grpc.UnaryInterceptor(srv.UnaryAuth), grpc.StreamInterceptor(srv.StreamAuth))
	srv.Attach(gs)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(func() { gs.Stop(); lis.Close() })

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	prov := mesapb.NewProvisionClient(conn)
	// A long-lived Subscribe stream the host holds open (never disconnects here).
	hostCtx := withToken(auth.HostKey("drone1"))
	if _, err := prov.Subscribe(hostCtx, &mesapb.SubscribeRequest{}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitForSub(t, srv, "drone1")

	// Server-driven drain: signal Subscribe to exit, then GracefulStop must return.
	srv.Shutdown()
	done := make(chan struct{})
	go func() { gs.GracefulStop(); close(done) }()
	select {
	case <-done: // good — the open stream did not block the drain
	case <-time.After(5 * time.Second):
		t.Fatal("GracefulStop hung with an open Subscribe stream (H14 not fixed)")
	}

	// Shutdown is idempotent (closes once).
	srv.Shutdown()
}

// mustPersona builds a registered-ready persona for the registry-only tests.
func mustPersona(t *testing.T, name string) persona.Persona {
	t.Helper()
	var p persona.Persona
	if err := json.Unmarshal(validPersonaJSON(t, name), &p); err != nil {
		t.Fatalf("unmarshal persona: %v", err)
	}
	return p
}

// TestPushGoalConcurrentWithFetch is the H13 race regression: PushGoal mutates a
// host's entry while Fetch/lookup readers concurrently read that same entry's
// fields (goals/goalPushed/persona/prose). Pre-fix, lookup() returned a bare
// *entry shared across goroutines and PushGoal mutated it in place — a torn
// slice-header read (run under `go test -race`). With copy-on-write PushGoal it's
// race-free. Run: `go test -race -run Push ./mesa/mesad/`.
func TestPushGoalConcurrentWithFetch(t *testing.T) {
	const tok = "s3cr3t"
	srv, conn := newAdminTestServer(t, tok)
	admin := mesapb.NewAdminClient(conn)
	prov := mesapb.NewProvisionClient(conn)
	adminCtx := withToken(tok)

	putPersonas(t, admin, adminCtx, []*mesapb.PersonaUpsert{
		{HostId: "drone1", PersonaJson: validPersonaJSON(t, "drone1")},
	})
	hostCtx := withToken(auth.HostKey("drone1"))

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Reader goroutines: Fetch reads e.goals/e.persona/e.prose through lookup().
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				if _, err := prov.Fetch(hostCtx, &mesapb.HostRef{}); err != nil {
					return
				}
			}
		}()
	}

	// Writer goroutine: PushGoal mutates the matched entry repeatedly.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			if _, err := admin.PushGoal(adminCtx, &mesapb.PushGoalRequest{
				Goal: "goal-" + string(rune('a'+i%26)), Match: "drone*",
			}); err != nil {
				return
			}
		}
	}()

	// Let the writer finish, then stop the readers and join.
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()

	// Sanity: the entry survived and carries an operator-pushed goal.
	e, ok := srv.lookup("drone1")
	if !ok || !e.goalPushed || len(e.goals) != 1 {
		t.Fatalf("after push storm: ok=%v entry=%+v", ok, e)
	}
}
