package debughttp

// whisper.go — POST /whisper: operator thought injection (the mentor channel
// without walking an avatar over). The cradle proxies it per host at
// /api/hosts/{name}/debug/whisper, exactly like /eval and /state.
//
// The body is JSON: text (required; the host clips to ~300 runes) + urgency
// (low|normal|high, default normal) queue a first-person thought the director
// voices into its next situation render ("A thought surfaces: …"), published
// on the bus as kind "whisper" so feeds show it; urgency=high additionally
// raises a reactive-tier conductor interrupt so the host considers it within
// seconds. subject+claim (+confidence) — all optional, but only together —
// also write an operator-grade entry into the knowledge ledger through the
// same writeback path reactive extraction uses (game-authoritative provenance,
// 0.85 default confidence, may close a standing open question).
//
// Unlike /eval this never takes evalMu: the queue is its own mutex on the
// Host, so a whisper lands even while a routine is mid-run — and it queues
// under analysis mode too (rendered when the director next runs for real).

import (
	"encoding/json"
	"net/http"
	"strings"
)

// whisperRequest is the POST /whisper body.
type whisperRequest struct {
	Text    string `json:"text"`
	Urgency string `json:"urgency,omitempty"` // low|normal|high; "" = normal
	// Subject+Claim (required together when used) write a ledger claim
	// alongside the queued thought; Confidence is optional (0 or out-of-range
	// ⇒ the operator-grade 0.85 authoritative default).
	Subject    string  `json:"subject,omitempty"`
	Claim      string  `json:"claim,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type whisperResponse struct {
	OK      bool   `json:"ok"`
	Queued  int    `json:"queued"`         // pending whispers after this one (cap 8, oldest dropped)
	Urgency string `json:"urgency"`        // the normalized urgency that was queued
	Ledger  bool   `json:"ledger_written"` // subject+claim landed in the knowledge ledger
	Error   string `json:"error,omitempty"`
}

func (d *Server) handleWhisper(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, whisperResponse{Error: "POST only"})
		return
	}
	var req whisperRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, whisperResponse{Error: "bad JSON body: " + err.Error()})
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		writeJSON(w, http.StatusBadRequest, whisperResponse{Error: `"text" is required`})
		return
	}
	switch req.Urgency {
	case "":
		req.Urgency = "normal"
	case "low", "normal", "high":
	default:
		// An operator typo silently downgraded to normal could bury an urgent
		// whisper — reject loudly instead.
		writeJSON(w, http.StatusBadRequest, whisperResponse{Error: "urgency must be low|normal|high"})
		return
	}
	subject, claim := strings.TrimSpace(req.Subject), strings.TrimSpace(req.Claim)
	if (subject == "") != (claim == "") {
		writeJSON(w, http.StatusBadRequest, whisperResponse{Error: "subject and claim must be given together"})
		return
	}
	d.host.QueueWhisper(req.Text, req.Urgency)
	if subject != "" {
		d.host.WhisperClaim(subject, claim, req.Confidence)
	}
	writeJSON(w, http.StatusOK, whisperResponse{
		OK: true, Queued: d.host.PendingWhispers(), Urgency: req.Urgency, Ledger: subject != "",
	})
}
