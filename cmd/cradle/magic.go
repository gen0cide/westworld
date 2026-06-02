package main

// magic.go — GET /spells and POST /cast route handlers for the remote client.
// Wired into serveClient via registerMagicRoutes, called once inside serveClient
// directly after registerSpriteRoutes.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/runtime"
)

// magicStaffItems maps an elemental rune item id to the staff item ids that
// supply that element for free (so the rune need not be held). Duplicated from
// runtime/views_magic.go:79-84 — the spec forbids importing from runtime/.
// Fire(31) → 197/615/682, Water(32) → 102/616/683, Air(33) → 101/617/684,
// Earth(34) → 103/618/685.
var magicStaffItems = map[int][]int{
	31: {197, 615, 682}, // fire
	32: {102, 616, 683}, // water
	33: {101, 617, 684}, // air
	34: {103, 618, 685}, // earth
}

// registerMagicRoutes adds GET /spells and POST /cast to mux.
func registerMagicRoutes(
	mux *http.ServeMux,
	host *runtime.Host,
	f *facts.Facts,
	enqueueAction func(func(context.Context) (string, error)) (string, error),
) {
	// GET /spells — static catalog, served once. The response is fully
	// determined by the embedded XML (facts.Spells); no world state is read.
	// Cache-Control: immutable (facts never change per process lifetime).
	type runeEntry struct {
		ItemID int `json:"itemId"`
		Count  int `json:"count"`
	}
	type spellEntry struct {
		ID          int         `json:"id"`
		Name        string      `json:"name"`
		ReqLevel    int         `json:"reqLevel"`
		Type        int         `json:"type"`
		Exp         int         `json:"exp"`
		Description string      `json:"description"`
		Members     bool        `json:"members"`
		Evil        bool        `json:"evil"`
		Runes       []runeEntry `json:"runes"`
	}

	// Build once at registration time; response bytes are immutable.
	catalog := make([]spellEntry, len(facts.Spells))
	for i, sp := range facts.Spells {
		runes := make([]runeEntry, len(sp.Runes))
		for j, r := range sp.Runes {
			runes[j] = runeEntry{ItemID: r.ItemID, Count: r.Count}
		}
		catalog[i] = spellEntry{
			ID: sp.ID, Name: sp.Name, ReqLevel: sp.ReqLevel,
			Type: int(sp.Type), Exp: sp.ExpReward,
			Description: sp.Description, Members: sp.Members,
			Evil: sp.Evil, Runes: runes,
		}
	}
	catalogJSON, _ := json.Marshal(catalog)

	mux.HandleFunc("/spells", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
		_, _ = w.Write(catalogJSON)
	})

	// POST /cast — funnels through the serialised action worker.
	mux.HandleFunc("/cast", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req castRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		if facts.SpellByID(req.SpellID) == nil {
			http.Error(w, "unknown spellId", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")

		spellID := req.SpellID
		targetKind := req.TargetKind
		targetIndex := req.TargetIndex

		msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
			switch targetKind {
			case "npc":
				return fmt.Sprintf("Cast spell %d on NPC %d", spellID, targetIndex),
					host.CastOnNpc(wctx, targetIndex, spellID)
			case "player":
				return fmt.Sprintf("Cast spell %d on player %d", spellID, targetIndex),
					host.CastOnPlayer(wctx, targetIndex, spellID)
			default: // "" or "self"
				return fmt.Sprintf("Cast spell %d on self", spellID),
					host.CastOnSelf(wctx, spellID)
			}
		})
		if runErr != nil {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
	})
}
