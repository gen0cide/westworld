package mesad

import (
	"context"
	"encoding/json"
	"io"
	"sort"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// admin.go implements the operator-only Admin service: runtime persona CRUD for
// mesa-ctl. Handlers reuse Server.Register (validate → render → persist to
// Postgres → register live), so changes take effect with NO mesad restart. Auth
// is the admin token (auth.go); these methods are never reachable with a host
// bearer token.

// PutPersonas bulk-upserts a streamed set of personas, validating + registering
// each. One bad persona does NOT fail the batch — its error rides back in the
// per-item report so a 199-good / 1-bad import still lands the 199.
func (s *Server) PutPersonas(stream grpc.ClientStreamingServer[mesapb.PersonaUpsert, mesapb.BatchResult]) error {
	res := &mesapb.BatchResult{}
	for {
		up, err := stream.Recv()
		if err == io.EOF {
			s.log.Info("admin: bulk persona upsert", "ok", res.Ok, "failed", res.Failed)
			return stream.SendAndClose(res)
		}
		if err != nil {
			return err
		}
		item := &mesapb.ItemResult{HostId: up.GetHostId()}
		if err := s.putOnePersona(up.GetHostId(), up.GetPersonaJson()); err != nil {
			item.Ok, item.Error = false, err.Error()
			res.Failed++
		} else {
			item.Ok = true
			res.Ok++
		}
		res.Items = append(res.Items, item)
	}
}

// putOnePersona unmarshals + registers a single persona — the shared body behind
// both the bulk stream and a single `mesa-ctl persona put`.
func (s *Server) putOnePersona(hostID string, personaJSON []byte) error {
	if hostID == "" {
		return status.Error(codes.InvalidArgument, "empty host_id")
	}
	var p persona.Persona
	if err := json.Unmarshal(personaJSON, &p); err != nil {
		return status.Errorf(codes.InvalidArgument, "persona json: %v", err)
	}
	// Register validates, renders prose/system, registers live, and persists.
	if err := s.Register(hostID, p); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return nil
}

// GetPersona returns one registered persona, including its JSON.
func (s *Server) GetPersona(ctx context.Context, ref *mesapb.HostRef) (*mesapb.PersonaRecord, error) {
	id := ref.GetHostId()
	e, ok := s.lookup(id)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no persona registered for host_id %q", id)
	}
	pj, err := json.Marshal(e.persona)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal persona: %v", err)
	}
	rec := &mesapb.PersonaRecord{HostId: id, Name: e.persona.Cornerstone.Identity.Name, PersonaJson: pj, Prose: e.prose}
	if t, ok := s.personaUpdatedAt(ctx)[id]; ok {
		rec.UpdatedAt = t.UTC().Format(time.RFC3339)
	}
	return rec, nil
}

// ListPersonas returns every registered persona, sorted by host_id. Metadata
// only (host_id, name, updated_at) unless with_json is set.
func (s *Server) ListPersonas(ctx context.Context, req *mesapb.ListPersonasRequest) (*mesapb.PersonaList, error) {
	times := s.personaUpdatedAt(ctx)
	s.mu.RLock()
	recs := make([]*mesapb.PersonaRecord, 0, len(s.reg))
	for id, e := range s.reg {
		rec := &mesapb.PersonaRecord{HostId: id, Name: e.persona.Cornerstone.Identity.Name}
		if req.GetWithJson() {
			if pj, err := json.Marshal(e.persona); err == nil {
				rec.PersonaJson = pj
			}
		}
		if t, ok := times[id]; ok {
			rec.UpdatedAt = t.UTC().Format(time.RFC3339)
		}
		recs = append(recs, rec)
	}
	s.mu.RUnlock()
	sort.Slice(recs, func(i, j int) bool { return recs[i].HostId < recs[j].HostId })
	return &mesapb.PersonaList{Personas: recs}, nil
}

// DeletePersona removes a persona from the live registry + durable store and
// revokes the host's derived auth token. Idempotent (ok even if absent).
func (s *Server) DeletePersona(ctx context.Context, ref *mesapb.HostRef) (*mesapb.AdminAck, error) {
	id := ref.GetHostId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "empty host_id")
	}
	s.mu.Lock()
	delete(s.reg, id)
	delete(s.tokens, mesaclient.HostKey(id))
	s.mu.Unlock()
	if s.ltm != nil {
		if err := s.ltm.DeletePersona(ctx, id); err != nil {
			return nil, status.Errorf(codes.Internal, "delete persona: %v", err)
		}
	}
	s.log.Info("admin: deleted persona", "host_id", id)
	return &mesapb.AdminAck{Ok: true}, nil
}

// personaUpdatedAt returns host_id→updated_at, best-effort: empty when there's no
// LTM or the query fails (updated_at is display metadata; it never blocks a
// response or the registry-as-source-of-truth set).
func (s *Server) personaUpdatedAt(ctx context.Context) map[string]time.Time {
	if s.ltm == nil {
		return nil
	}
	m, err := s.ltm.PersonaTimes(ctx)
	if err != nil {
		s.log.Warn("admin: persona times query failed", "err", err)
		return nil
	}
	return m
}
