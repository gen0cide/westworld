package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/persona"
)

// Persona wiring: the deterministic local half of the persona compile
// (pearl table + decision floor + affect baseline) plus the mesa
// provisioning path and the live-directive subscription. Called from
// RunHost (runhost.go).

// applyPersona compiles a persona's policy and wires it onto the host: the pearl
// table (gate + decide seam), the decision floor, and the affect baseline. This
// is the deterministic local half of the compiler — shared by the offline
// file path and the mesa-provisioned path.
func applyPersona(log *slog.Logger, host *Host, p *persona.Persona) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid persona: %w", err)
	}
	cp := persona.CompilePolicy(p)
	host.Pearl = pearl.New(cp.Table, cp.DecisionFloor)
	m := p.Trajectory.Mood
	host.SetAffectBaseline(m.Stress, m.Confidence, m.Valence)
	// Capture the explore<->exploit dial vector for decision-time curiosity
	// weighting (the persona is otherwise discarded after this). Sole chokepoint
	// for both the offline (loadPersona) and mesa (provisionPersona) paths.
	host.curiosity = p.Cornerstone.Prefs.Curiosity
	// Persona north-star: the advancement fallback when the active goal closes and
	// no graph open_goal successor is queued (same chokepoint as curiosity capture).
	host.northStar = strings.TrimSpace(p.Cornerstone.Identity.NorthStar.Statement)
	// One-line "who I am" grounding card for the reactive extractor (the persona is
	// otherwise discarded after this — same chokepoint as the curiosity capture).
	host.personaSnippet = strings.TrimSpace(fmt.Sprintf("%s — %s",
		strings.TrimSpace(p.Cornerstone.Identity.Name), strings.TrimSpace(p.Cornerstone.Identity.ArchetypeTag)))
	log.Info("loaded persona",
		"name", p.Cornerstone.Identity.Name,
		"archetype", p.Cornerstone.Identity.ArchetypeTag,
		"pearl_rules", len(cp.Table.Rules),
		"decision_floor", cp.DecisionFloor,
		"fairness", cp.Trade.FairnessThreshold,
		"scam_propensity", cp.Trade.ScamPropensity,
		"risk_aversion", cp.Trade.RiskAversion,
	)
	for _, r := range cp.Table.Rules {
		log.Info("  pearl rule", "id", r.ID, "origin", r.Origin)
	}
	return nil
}

// loadPersona reads + validates a local persona file and applies it (the offline
// path, when there is no mesa to provision from).
func loadPersona(log *slog.Logger, host *Host, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var p persona.Persona
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return applyPersona(log, host, &p)
}

// provisionPersona pulls the host's authoritative persona from mesa (unary
// Provision.Fetch), applies it, then opens the live push stream so later
// revisions can be applied without a restart.
func provisionPersona(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, hostID string) ([]string, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	prov, err := mc.Provision(fetchCtx, hostID)
	if err != nil {
		return nil, err
	}
	if err := applyPersona(log, host, &prov.Persona); err != nil {
		return nil, err
	}
	log.Info("provisioned persona from mesa",
		"name", prov.Persona.Cornerstone.Identity.Name,
		"goals", len(prov.Goals), "prose_chars", len(prov.Prose))
	go subscribeDirectives(ctx, log, host, mc, hostID)
	return prov.Goals, nil
}

// subscribeDirectives consumes the mesa→host push stream (Provision.Subscribe)
// and applies what it can live. GOAL_REVISION installs an operator goal override
// on the host (read by the director each turn); applying PEARL_REFRESH /
// PERSONA_REVISION (recompile) lands later — those are still logged.
func subscribeDirectives(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, hostID string) {
	// The stream dies with mesad (the server side holds it), so a one-shot
	// subscribe silently severs the operator push channel at the first mesad
	// restart — re-subscribe forever with capped backoff instead.
	backoff := time.Second
	for {
		ch, err := mc.Subscribe(ctx, hostID)
		if err != nil {
			log.Warn("mesa subscribe failed; will retry", "err", err, "backoff", backoff)
		} else {
			backoff = time.Second
			consumeDirectives(log, host, ch)
			log.Warn("mesa directive stream closed; re-subscribing")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(2*backoff, 30*time.Second)
	}
}

func consumeDirectives(log *slog.Logger, host *Host, ch <-chan mesaclient.Directive) {
	for d := range ch {
		switch d.Kind {
		case mesaclient.DirectiveGoalRevision:
			var goals []string
			if err := json.Unmarshal(d.Payload, &goals); err != nil {
				log.Warn("mesa directive: bad goal_revision payload", "id", d.ID, "err", err)
				continue
			}
			if len(goals) == 0 {
				continue
			}
			host.SetLiveGoal(goals[0])
			log.Info("mesa directive: live goal applied", "id", d.ID, "goal", goals[0])
		default:
			log.Info("mesa directive", "id", d.ID, "kind", string(d.Kind), "bytes", len(d.Payload))
		}
	}
}

// loadAliasStore opens the per-host JSON alias store, or returns nil (in-memory
