package cradle

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// adminTokenEnv is the operator credential mesad's Admin API is gated on (the
// same env var mesa-ctl reads). Distinct from per-host bearer tokens — the
// persona registry is operator-only.
const adminTokenEnv = "ADMIN_TOKEN"

// personaRPCTimeout bounds one Admin.GetPersona call so a hung mesa degrades
// to the 501 hide-the-pane path instead of pinning the UI's fetch.
const personaRPCTimeout = 5 * time.Second

var errNoAdminToken = errors.New("no admin credentials ($ADMIN_TOKEN unset); persona proxy unavailable")

// personaResponse is the persona proxy body: the Admin.GetPersona record with
// the persona document inlined raw. The full document carries everything the
// Tablet header renders — HEXACO mu+bands (cornerstone.hexaco), curiosity
// flavors + λ/patience/aggression/decisiveness (cornerstone.prefs), and the
// north star (cornerstone.identity.north_star). Mood (Trajectory.Affect) is
// deliberately absent: no admin RPC exposes it today.
type personaResponse struct {
	HostID    string          `json:"host_id"`
	Name      string          `json:"name,omitempty"`
	UpdatedAt string          `json:"updated_at,omitempty"`
	Prose     string          `json:"prose,omitempty"`
	Persona   json.RawMessage `json:"persona"`
}

// handlePersona proxies mesa Admin.GetPersona for one managed host (gap G-7).
// Unknown host is 404 (registry idiom); ANY proxy failure — no admin token, no
// mesa address, mesa down — is 501 with an error body, the UI's signal to hide
// the persona pane.
//
// TRUST BOUNDARY: this re-serves operator-only mesa data ($ADMIN_TOKEN-gated
// at mesad) on the cradle's HTTP API, which is deliberately unauthenticated —
// the same posture as POST /eval and the lifecycle endpoints. The cradle
// listener is assumed bound to an operator-trusted interface; gate the
// listener (or grow a shared operator credential for the whole API), not this
// handler, if that ever changes.
func (a *API) handlePersona(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	st, err := a.reg.Get(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	if st.Mesa == "" {
		writeJSON(w, http.StatusNotImplemented, errBody(errors.New("host has no mesa address; persona proxy unavailable")))
		return
	}
	if os.Getenv(adminTokenEnv) == "" {
		writeJSON(w, http.StatusNotImplemented, errBody(errNoAdminToken))
		return
	}
	c, err := a.adminClient(st.Mesa)
	if err != nil {
		writeJSON(w, http.StatusNotImplemented, errBody(err))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), personaRPCTimeout)
	defer cancel()
	rec, err := c.GetPersona(ctx, &mesapb.HostRef{HostId: name})
	if err != nil {
		writeJSON(w, http.StatusNotImplemented, errBody(err))
		return
	}
	doc := json.RawMessage(rec.PersonaJson)
	if len(doc) == 0 {
		doc = json.RawMessage("null") // keep the body valid JSON if mesa sent no document
	}
	writeJSON(w, http.StatusOK, personaResponse{
		HostID:    rec.HostId,
		Name:      rec.Name,
		UpdatedAt: rec.UpdatedAt,
		Prose:     rec.Prose,
		Persona:   doc,
	})
}

// adminClient returns the cached admin client for a mesa address, dialing
// lazily on the first persona request. grpc.NewClient never connects eagerly —
// a down mesa surfaces at RPC time — so a cached client cannot go stale.
// Connections live for the daemon's lifetime (one per distinct mesa address;
// hosts may target different mesa instances).
func (a *API) adminClient(addr string) (mesapb.AdminClient, error) {
	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	if c, ok := a.admins[addr]; ok {
		return c, nil
	}
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(adminTokenCreds{}))
	if err != nil {
		return nil, err
	}
	c := mesapb.NewAdminClient(conn)
	a.admins[addr] = c
	return c, nil
}

// adminTokenCreds attaches the operator bearer token per RPC (mirrors
// mesa-ctl's adminCreds; insecure transport for the local/trusted link).
// Reading the env per call keeps a token rotation effective without a redial.
type adminTokenCreds struct{}

func (adminTokenCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	tok := os.Getenv(adminTokenEnv)
	if tok == "" {
		return nil, nil
	}
	return map[string]string{"authorization": "Bearer " + tok}, nil
}

func (adminTokenCreds) RequireTransportSecurity() bool { return false }
