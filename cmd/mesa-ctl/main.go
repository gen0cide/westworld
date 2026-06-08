// Command mesa-ctl is a thin gRPC client over mesad's operator-only Admin
// service: manage the persona registry at runtime — single or bulk — with no
// mesad restart. It authenticates with the admin token ($ADMIN_TOKEN or -token),
// distinct from per-host bearer tokens.
//
//	mesa-ctl persona put Delores dolores.json   # single (file or - for stdin)
//	mesa-ctl persona import ./personas/         # bulk: a dir of <host_id>.json
//	mesa-ctl persona import -                    # bulk: NDJSON {host_id,persona} on stdin
//	mesa-ctl persona ls [--json]
//	mesa-ctl persona get Delores
//	mesa-ctl persona rm Delores
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

func main() {
	addr := flag.String("addr", "localhost:7077", "mesad gRPC host:port")
	token := flag.String("token", os.Getenv("ADMIN_TOKEN"), "admin token (default $ADMIN_TOKEN)")
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(adminCreds{token: *token}))
	if err != nil {
		fail(err)
	}
	defer conn.Close()
	c := mesapb.NewAdminClient(conn)
	ctx := context.Background()

	switch args[0] {
	case "persona":
		err = personaCmd(ctx, c, args[1:])
	case "fleet":
		err = fleetCmd(ctx, c, args[1:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fail(err)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `mesa-ctl [-addr host:port] [-token <admin-token>] <command> [args]

  persona put <host_id> <file|->     register/replace one persona (file or stdin)
  persona import <dir|->              bulk: a directory of <host_id>.json, or NDJSON
                                      ({"host_id":..,"persona":{..}} per line) on stdin
  persona ls [--json]                list registered personas
  persona get <host_id> [--json]     show one persona (facets + dials + prose; --json for raw)
  persona set <host_id> <field> <v>  edit one dial/field (validated); e.g. set drone1 aggression low
  persona rm <host_id>               remove a persona
  fleet gen [flags]                  emit a cradle hostcfg for the registered
                                     personas (names == host_ids); see -h for flags

Auth: $ADMIN_TOKEN (or -token) must match mesad's ADMIN_TOKEN.
`)
}

func personaCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mesa-ctl persona put|import|ls|get|rm ...")
	}
	switch args[0] {
	case "put":
		return putCmd(ctx, c, args[1:])
	case "import":
		return importCmd(ctx, c, args[1:])
	case "ls", "list":
		return listCmd(ctx, c, args[1:])
	case "get":
		return getCmd(ctx, c, args[1:])
	case "set":
		return setCmd(ctx, c, args[1:])
	case "rm", "delete":
		return rmCmd(ctx, c, args[1:])
	default:
		return fmt.Errorf("unknown persona subcommand %q", args[0])
	}
}

func putCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: mesa-ctl persona put <host_id> <file|->")
	}
	raw, err := readSource(args[1])
	if err != nil {
		return err
	}
	res, err := stream(ctx, c, []*mesapb.PersonaUpsert{{HostId: args[0], PersonaJson: raw}})
	if err != nil {
		return err
	}
	return report(res)
}

func importCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mesa-ctl persona import <dir|->")
	}
	items, err := collectImport(args[0])
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("no personas found in %q", args[0])
	}
	res, err := stream(ctx, c, items)
	if err != nil {
		return err
	}
	return report(res)
}

// collectImport gathers PersonaUpserts from a directory of <host_id>.json files,
// or from NDJSON on stdin ({"host_id":..,"persona":{..}} per line) when src is "-".
func collectImport(src string) ([]*mesapb.PersonaUpsert, error) {
	if src == "-" {
		return parseNDJSON(os.Stdin)
	}
	info, err := os.Stat(src)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("import source must be a directory or '-' (got a file; use 'persona put' for one)")
	}
	return collectDir(src)
}

// parseNDJSON reads one {"host_id":..,"persona":{..}} object per line.
func parseNDJSON(r io.Reader) ([]*mesapb.PersonaUpsert, error) {
	var items []*mesapb.PersonaUpsert
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24) // allow large persona lines
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row struct {
			HostID  string          `json:"host_id"`
			Persona json.RawMessage `json:"persona"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("ndjson line: %w", err)
		}
		items = append(items, &mesapb.PersonaUpsert{HostId: row.HostID, PersonaJson: row.Persona})
	}
	return items, sc.Err()
}

// collectDir reads every <host_id>.json in dir (host_id = filename without .json).
func collectDir(dir string) ([]*mesapb.PersonaUpsert, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var items []*mesapb.PersonaUpsert
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		hostID := strings.TrimSuffix(e.Name(), ".json")
		items = append(items, &mesapb.PersonaUpsert{HostId: hostID, PersonaJson: raw})
	}
	return items, nil
}

func listCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	asJSON := false
	for _, a := range args {
		if a == "--json" {
			asJSON = true
		}
	}
	res, err := c.ListPersonas(ctx, &mesapb.ListPersonasRequest{WithJson: asJSON})
	if err != nil {
		return err
	}
	if asJSON {
		out, _ := json.MarshalIndent(res.Personas, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	fmt.Printf("%-24s %-24s %s\n", "HOST_ID", "NAME", "UPDATED")
	for _, p := range res.Personas {
		fmt.Printf("%-24s %-24s %s\n", p.HostId, p.Name, p.UpdatedAt)
	}
	fmt.Printf("(%d personas)\n", len(res.Personas))
	return nil
}

func getCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	host, jsonOut := "", false
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
		} else if host == "" {
			host = a
		}
	}
	if host == "" {
		return fmt.Errorf("usage: mesa-ctl persona get <host_id> [--json]")
	}
	rec, err := c.GetPersona(ctx, &mesapb.HostRef{HostId: host})
	if err != nil {
		return err
	}
	if jsonOut {
		var v any
		if json.Unmarshal(rec.PersonaJson, &v) == nil {
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Println(string(rec.PersonaJson))
		}
		return nil
	}

	// Readable view: identity facets + dials (same names as `persona set`) + prose.
	var m map[string]any
	if err := json.Unmarshal(rec.PersonaJson, &m); err != nil {
		return err
	}
	g := func(p ...string) string { return getPath(m, p...) }
	fmt.Printf("%s  [%s]  cohort=%s  updated=%s\n", rec.HostId,
		g("cornerstone", "generation_meta", "archetype"), g("cornerstone", "generation_meta", "cohort_id"), rec.UpdatedAt)
	fmt.Printf("north_star (%s): %s\n", g("cornerstone", "identity", "north_star", "theme"), g("cornerstone", "identity", "north_star", "statement"))
	fmt.Printf("voice: %s, typos=%s, register=%q\n", g("cornerstone", "identity", "voice", "formality"), g("cornerstone", "identity", "voice", "typo_feel"), g("cornerstone", "identity", "voice", "register"))
	fmt.Printf("values: %s / %s\n", g("cornerstone", "values", "north_star_value"), g("cornerstone", "values", "secondary_value"))
	hx := func(k string) string { return g("cornerstone", "hexaco", k, "band") }
	fmt.Printf("hexaco: H=%s E=%s X=%s A=%s C=%s O=%s\n", hx("H"), hx("E"), hx("X"), hx("A"), hx("C"), hx("O"))
	pb := func(k string) string { return g("cornerstone", "prefs", k, "band") }
	fmt.Printf("prefs: patience=%s aggression=%s decisiveness=%s tenacity=%s loss_aversion=%s bulk_apperception=%s coop=%s\n",
		pb("patience"), pb("aggression"), pb("decisiveness"), pb("tenacity"), pb("loss_aversion"), pb("bulk_apperception"), g("cornerstone", "prefs", "coop_type"))
	fmt.Printf("risk: economic=%s bodily=%s social=%s   attention=%s\n",
		g("cornerstone", "prefs", "risk", "economic"), g("cornerstone", "prefs", "risk", "bodily"), g("cornerstone", "prefs", "risk", "social"), g("cornerstone", "prefs", "attention", "level"))
	cu := func(k string) string { return g("cornerstone", "prefs", "curiosity", k) }
	fmt.Printf("curiosity: social=%s spatial=%s skill=%s economic=%s risk=%s\n", cu("social"), cu("spatial"), cu("skill"), cu("economic"), cu("risk"))
	if sp := pb("self_preservation"); sp != "" {
		fmt.Printf("self_preservation=%s\n", sp)
	}
	fmt.Printf("\nprose:\n%s\n", rec.Prose)
	return nil
}

func rmCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mesa-ctl persona rm <host_id>")
	}
	if _, err := c.DeletePersona(ctx, &mesapb.HostRef{HostId: args[0]}); err != nil {
		return err
	}
	fmt.Printf("%s: removed\n", args[0])
	return nil
}

// stream sends all upserts through one PutPersonas client-stream.
func stream(ctx context.Context, c mesapb.AdminClient, items []*mesapb.PersonaUpsert) (*mesapb.BatchResult, error) {
	s, err := c.PutPersonas(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if err := s.Send(it); err != nil {
			return nil, err
		}
	}
	return s.CloseAndRecv()
}

// report prints the per-item result and returns an error if any item failed (so
// the exit code is non-zero on a partial failure).
func report(res *mesapb.BatchResult) error {
	for _, it := range res.Items {
		if it.Ok {
			fmt.Printf("  ✓ %s\n", it.HostId)
		} else {
			fmt.Printf("  ✗ %s: %s\n", it.HostId, it.Error)
		}
	}
	fmt.Printf("%d ok, %d failed\n", res.Ok, res.Failed)
	if res.Failed > 0 {
		return fmt.Errorf("%d persona(s) failed", res.Failed)
	}
	return nil
}

// readSource reads a file, or stdin when path is "-".
func readSource(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "mesa-ctl:", err)
	os.Exit(1)
}

// adminCreds attaches the admin bearer token to every RPC (mirrors the host
// client's tokenCreds; insecure transport for the local/trusted link).
type adminCreds struct{ token string }

func (a adminCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	if a.token == "" {
		return nil, nil
	}
	return map[string]string{"authorization": "Bearer " + a.token}, nil
}
func (adminCreds) RequireTransportSecurity() bool { return false }
