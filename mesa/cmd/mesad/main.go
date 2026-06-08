// Command mesad runs the mesa service: the off-host gRPC server that owns LLM
// inference, the persona lifecycle, RAG, and long-term memory. Hosts dial it as
// their own host_id. It holds the Anthropic key (the host never does).
//
//	mesad -addr :7077 -host Delores=docs/personas/dolores.json
//
// ANTHROPIC_API_KEY enables the LLM seams (Decide/Act); without it, persona
// provisioning still works and the host degrades to local behavior.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/mesa/embed"
	"github.com/gen0cide/westworld/mesa/llm"
	"github.com/gen0cide/westworld/mesa/mesad"
	"github.com/gen0cide/westworld/persona"
)

// defaultDSN is the local-dev Postgres connection used when neither -db nor
// DATABASE_URL is set. mesad keeps ALL its durable state in Postgres.
const defaultDSN = "postgres://localhost:5432/westworld?sslmode=disable"

// hostMap collects repeatable -host host_id=persona.json flags.
type hostMap map[string]string

func (m hostMap) String() string { return fmt.Sprintf("%v", map[string]string(m)) }
func (m hostMap) Set(v string) error {
	id, path, ok := strings.Cut(v, "=")
	if !ok || id == "" || path == "" {
		return fmt.Errorf("want host_id=path, got %q", v)
	}
	m[id] = path
	return nil
}

func main() {
	addr := flag.String("addr", ":7077", "gRPC listen address")
	actModel := flag.String("act-model", "claude-sonnet-4-6", "model for Act/DSL authoring (high volume; manual is cached)")
	decideModel := flag.String("decide-model", "claude-haiku-4-5-20251001", "model for narrow Decide option-picks")
	genesisModel := flag.String("genesis-model", "claude-opus-4-8", "model for session-genesis (rare, history-rich login compile)")
	dbDSN := flag.String("db", "", "PostgreSQL DSN for durable storage (default $DATABASE_URL or "+defaultDSN+")")
	factsRoot := flag.String("facts", "/Users/flint/Code/openrsc", "OpenRSC source root for world name-catalogs (static arg validation); empty disables")
	hosts := hostMap{}
	flag.Var(hosts, "host", "host_id=persona.json (repeatable)")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var actLLM, decideLLM, genesisLLM *llm.Client
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		actLLM = llm.New(key, *actModel)
		decideLLM = llm.New(key, *decideModel)
		genesisLLM = llm.New(key, *genesisModel)
		log.Info("llm enabled", "act_model", actLLM.Model(), "decide_model", decideLLM.Model(), "genesis_model", genesisLLM.Model())
	} else {
		log.Warn("ANTHROPIC_API_KEY unset; LLM seams (Act/Decide/Genesis) will be Unavailable")
	}

	srv := mesad.New(actLLM, decideLLM, genesisLLM, log)

	// World name-catalogs for static arg validation: load the lightweight
	// defs-only name sets (item/npc names + gazetteer places/POI types) so a
	// hallucinated literal arg (go_to("mining-site"), eat("typo-item")) is
	// rejected + re-prompted before it round-trips to the host. If -facts is
	// empty or the load fails, catalog validation is skipped (never blocks).
	if *factsRoot != "" {
		if f, err := facts.LoadCatalogs(facts.DefaultSources(*factsRoot)); err != nil {
			log.Warn("facts catalogs load failed; static arg validation disabled", "root", *factsRoot, "err", err)
		} else {
			cat := mesad.NewArgCatalog(f)
			srv.SetCatalog(cat)
			log.Info("world name-catalogs loaded for static arg validation",
				"items", len(f.ItemDefs), "npcs", len(f.NpcDefs),
				"places", len(f.Gazetteer().Places), "pois", len(f.Gazetteer().POIs))
		}
	} else {
		log.Info("-facts empty; static arg validation disabled")
	}

	// Durable long-term memory in Postgres (mesa's system of record). Resolution:
	// -db flag > $DATABASE_URL > local default. Fatal if unreachable — mesad is
	// not meant to silently run without its storage.
	dsn := *dbDSN
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = defaultDSN
	}
	// Semantic recall: embed episodes + queries with Voyage when a key is set;
	// without it, Recall degrades to Postgres full-text + recency.
	var embedder mesad.Embedder
	if key := os.Getenv("VOYAGE_AI_KEY"); key != "" {
		embedder = embed.NewVoyage(key, embed.DefaultModel)
		log.Info("embeddings enabled (voyage)", "model", embed.DefaultModel)
	} else {
		log.Warn("VOYAGE_AI_KEY unset; recall degrades to full-text + recency")
	}
	ltm, err := mesad.OpenLTM(context.Background(), dsn, embedder)
	if err != nil {
		log.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer ltm.Close()
	srv.SetLTM(ltm)
	log.Info("ltm connected (postgres)")

	// Restore personas already persisted in Postgres, so a host's identity
	// survives a restart without re-specifying its -host file.
	if err := srv.LoadPersonas(context.Background()); err != nil {
		log.Warn("load personas from postgres failed", "err", err)
	}

	for id, path := range hosts {
		p, err := loadPersona(path)
		if err != nil {
			log.Error("load persona failed", "host_id", id, "path", path, "err", err)
			os.Exit(1)
		}
		if err := srv.Register(id, *p); err != nil {
			log.Error("register persona failed", "host_id", id, "err", err)
			os.Exit(1)
		}
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Error("listen failed", "addr", *addr, "err", err)
		os.Exit(1)
	}
	gs := grpc.NewServer(
		grpc.UnaryInterceptor(srv.UnaryAuth),
		grpc.StreamInterceptor(srv.StreamAuth),
	)
	srv.Attach(gs)
	log.Info("mesa listening", "addr", *addr, "hosts", len(hosts))
	if err := gs.Serve(lis); err != nil {
		log.Error("serve failed", "err", err)
		os.Exit(1)
	}
}

func loadPersona(path string) (*persona.Persona, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p persona.Persona
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
