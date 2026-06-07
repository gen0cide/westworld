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
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/gen0cide/westworld/mesa/llm"
	"github.com/gen0cide/westworld/mesa/mesad"
	"github.com/gen0cide/westworld/persona"
)

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
	hosts := hostMap{}
	flag.Var(hosts, "host", "host_id=persona.json (repeatable)")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var actLLM, decideLLM *llm.Client
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		actLLM = llm.New(key, *actModel)
		decideLLM = llm.New(key, *decideModel)
		log.Info("llm enabled", "act_model", actLLM.Model(), "decide_model", decideLLM.Model())
	} else {
		log.Warn("ANTHROPIC_API_KEY unset; LLM seams (Decide/Act) will be Unavailable")
	}

	srv := mesad.New(actLLM, decideLLM, log)
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
