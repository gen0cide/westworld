package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/spec"
)

// REPL is the interactive shell that reads DSL fragments from
// `in`, evaluates them against a long-running interpreter Session
// bound to this Host, and prints results / errors to `out`.
//
// The Session inherits the Host's reserved entities (self, world,
// inventory, combat) and action callables — typing `self.hp` at
// the prompt queries the live host the same way a routine would.
//
// One line per input. The line is tried first as an expression
// (so `self.hp` prints the value); if expression parsing fails,
// it's retried as a statement (so `x = 5` works). Meta-commands
// prefixed with `.` (`.help`, `.state`, `.quit`, etc.) are handled
// specially and do not go through the interpreter.
//
// Per docs/lang/repl.md. The drop-into-shell-on-failure mode
// (`-repl-on-fail` with `.resume`) is planned but not yet wired —
// see task #54 for the work split.
type REPL struct {
	host *Host
	sess *interp.Session
	in   *bufio.Scanner
	out  io.Writer
	ctx  context.Context
}

// NewREPL constructs a REPL bound to the given Host and IO. The
// session is created immediately (so reserved entities + builtins
// are ready). Run() drives the read loop.
func (h *Host) NewREPL(ctx context.Context, in io.Reader, out io.Writer) *REPL {
	it := h.NewRoutineInterpreter(ctx)
	sess := it.NewSession(ctx, "<repl>")
	return &REPL{
		host: h,
		sess: sess,
		in:   bufio.NewScanner(in),
		out:  out,
		ctx:  ctx,
	}
}

// Run drives the read-eval-print loop until the user types `.quit`
// or the input ends (EOF on stdin). Returns nil on clean exit; a
// non-nil error only if Scanner itself fails (rare).
func (r *REPL) Run() error {
	r.printBanner()
	for {
		if err := r.ctx.Err(); err != nil {
			fmt.Fprintln(r.out, "[INFO] ctx canceled; exiting REPL")
			return nil
		}
		fmt.Fprint(r.out, ">>> ")
		if !r.in.Scan() {
			fmt.Fprintln(r.out)
			fmt.Fprintln(r.out, "[INFO] EOF on input; exiting REPL")
			return r.in.Err()
		}
		line := strings.TrimSpace(r.in.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ".") {
			if quit := r.handleMeta(line); quit {
				return nil
			}
			continue
		}
		r.evalAndPrint(line)
	}
}

func (r *REPL) printBanner() {
	fmt.Fprintln(r.out, "westworld REPL — type .help for commands, .quit to exit")
}

// evalAndPrint runs one DSL fragment and prints the result (for
// expressions) or just an "ok" marker (for statements).
func (r *REPL) evalAndPrint(line string) {
	res := r.sess.Eval(r.ctx, line)
	if res.Err != nil {
		fmt.Fprintf(r.out, "ERR: %v\n", res.Err)
		return
	}
	if res.IsExpression && res.Value != nil {
		fmt.Fprintf(r.out, "%s\n", res.Value.Display())
	}
}

// handleMeta dispatches dot-commands. Returns true iff the REPL
// should exit after this command.
func (r *REPL) handleMeta(line string) bool {
	fields := strings.Fields(line)
	cmd := fields[0]
	args := fields[1:]
	switch cmd {
	case ".quit", ".exit", ".q":
		fmt.Fprintln(r.out, "[INFO] exiting REPL")
		return true
	case ".help":
		r.metaHelp(args)
	case ".state":
		r.metaState()
	case ".env":
		r.metaEnv()
	case ".builtins":
		r.metaBuiltins()
	case ".accessors":
		r.metaAccessors()
	case ".events":
		// Two meanings of `.events`: with no args, dump the
		// recent-events buffer (the in-flight values delores has
		// observed). With "spec" arg, list the language's event
		// catalog (the on-handler signatures).
		if len(args) == 1 && args[0] == "spec" {
			r.metaEvents()
		} else {
			r.metaEventsRecent()
		}
	case ".load":
		r.metaLoad(args)
	case ".run":
		r.metaRun(args)
	default:
		fmt.Fprintf(r.out, "ERR: unknown meta command %q — try .help\n", cmd)
	}
	return false
}

func (r *REPL) metaHelp(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(r.out, "Meta commands:")
		fmt.Fprintln(r.out, "  .quit / .exit / .q     exit the REPL")
		fmt.Fprintln(r.out, "  .help [<name>]         this help, or docs for a specific builtin")
		fmt.Fprintln(r.out, "  .state                 print host vitals snapshot")
		fmt.Fprintln(r.out, "  .env                   list locals defined in this session")
		fmt.Fprintln(r.out, "  .builtins              list every registered DSL callable")
		fmt.Fprintln(r.out, "  .accessors             list known query-layer accessor paths")
		fmt.Fprintln(r.out, "  .events                dump the recent-events buffer (live)")
		fmt.Fprintln(r.out, "  .events spec           list bus events you can `on`-handle")
		fmt.Fprintln(r.out, "  .load <path>           load a .routine file's procs + handlers")
		fmt.Fprintln(r.out, "  .run <path>            load + invoke a .routine file's entry point")
		fmt.Fprintln(r.out)
		fmt.Fprintln(r.out, "Anything else is parsed as DSL — try `self.hp`, `inventory.free`,")
		fmt.Fprintln(r.out, "or `on chat_received(s, m) { note(s) }` to register a live handler.")
		return
	}
	name := args[0]
	base, _ := spec.StripBang(name)
	if a, ok := spec.ByName(base); ok {
		bang := ""
		if a.BangEligible() {
			bang = " (bang variant: " + a.Name + "!)"
		}
		fmt.Fprintf(r.out, "%s — %s%s\n", a.Name, a.Kind, bang)
		fmt.Fprintf(r.out, "  arity: %d..%d\n", a.MinArgs, a.MaxArgs)
		if len(a.Params) > 0 {
			fmt.Fprintf(r.out, "  params: %s\n", strings.Join(a.Params, ", "))
		}
		fmt.Fprintf(r.out, "  %s\n", a.DocSummary)
		if a.NotYetImplemented {
			fmt.Fprintln(r.out, "  (not yet implemented — calls return Fail(NOT_IMPLEMENTED))")
		}
		return
	}
	if e, ok := spec.EventByName(name); ok {
		fmt.Fprintf(r.out, "on %s — bus event\n", e.Name)
		if len(e.Params) > 0 {
			fmt.Fprintf(r.out, "  params: %s\n", strings.Join(e.Params, ", "))
		}
		fmt.Fprintf(r.out, "  %s\n", e.DocSummary)
		return
	}
	fmt.Fprintf(r.out, "ERR: no spec entry for %q (try .builtins / .events / .accessors)\n", name)
}

func (r *REPL) metaState() {
	w := r.host.World()
	if w == nil || w.Self == nil {
		fmt.Fprintln(r.out, "(no world state — host not connected?)")
		return
	}
	pos := w.Self.Position()
	fmt.Fprintf(r.out, "  position    %d, %d\n", pos.X, pos.Y)
	fmt.Fprintf(r.out, "  hp          %d / %d\n", w.Self.HP(), w.Self.MaxHP())
	fmt.Fprintf(r.out, "  prayer      %d / %d\n", w.Self.Prayer(), w.Self.MaxPrayer())
	fmt.Fprintf(r.out, "  fatigue     %d%%\n", w.Self.FatiguePercent())
	fmt.Fprintf(r.out, "  combat_lvl  %d\n", w.Self.CombatLevel())
	if inv := w.Inventory; inv != nil {
		fmt.Fprintf(r.out, "  inv         %d used / %d free\n",
			len(inv.Slots()), inv.FreeSlots())
	}
	if npcs := w.Npcs; npcs != nil {
		fmt.Fprintf(r.out, "  npcs visible %d\n", len(npcs.All()))
	}
}

func (r *REPL) metaEnv() {
	// Walk the session env and print every defined local. The
	// reserved entities (self/world/inventory/combat) are bound at
	// the root scope but they're the same across sessions; only
	// surface user-defined locals to keep the listing focused.
	env := r.sess.Env()
	names := []string{}
	// Env doesn't currently expose its names. Walk a small helper
	// here — works because we only need this for diagnostics.
	for _, name := range knownEnvNames(env) {
		if isReservedDSLName(name) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintln(r.out, "  (no user-defined locals yet)")
		return
	}
	for _, n := range names {
		v, _ := env.Get(n)
		if v == nil {
			fmt.Fprintf(r.out, "  %s = nil\n", n)
		} else {
			fmt.Fprintf(r.out, "  %s = %s\n", n, v.Display())
		}
	}
}

func (r *REPL) metaBuiltins() {
	byKind := map[spec.ActionKind][]spec.ActionSpec{}
	for _, a := range spec.Actions {
		byKind[a.Kind] = append(byKind[a.Kind], a)
	}
	kinds := []spec.ActionKind{
		spec.PrimaryAction, spec.Primitive,
		spec.LLMStdlib, spec.MemoryStdlib, spec.PersonaRead,
	}
	for _, k := range kinds {
		group := byKind[k]
		if len(group) == 0 {
			continue
		}
		fmt.Fprintf(r.out, "%s:\n", k)
		sort.Slice(group, func(i, j int) bool { return group[i].Name < group[j].Name })
		for _, a := range group {
			marker := ""
			if a.NotYetImplemented {
				marker = " (stub)"
			}
			bang := ""
			if a.BangEligible() {
				bang = "!"
			}
			fmt.Fprintf(r.out, "  %s%s%s\n", a.Name, bang, marker)
		}
	}
}

func (r *REPL) metaAccessors() {
	for _, a := range spec.Accessors {
		marker := ""
		if a.NotYetImplemented {
			marker = " (stub)"
		}
		fmt.Fprintf(r.out, "  %s : %s%s\n", strings.Join(a.Path, "."), a.Kind, marker)
	}
}

func (r *REPL) metaEvents() {
	for _, e := range spec.Events {
		marker := ""
		if e.NotYetImplemented {
			marker = " (not yet dispatched)"
		}
		params := ""
		if len(e.Params) > 0 {
			params = "(" + strings.Join(e.Params, ", ") + ")"
		} else {
			params = "()"
		}
		fmt.Fprintf(r.out, "  on %s%s%s\n      %s\n", e.Name, params, marker, e.DocSummary)
	}
}

// metaEventsRecent dumps the live recent-events buffer — what
// delores has observed lately. Null kinds are skipped.
func (r *REPL) metaEventsRecent() {
	rec := r.host.World().Recent
	if rec == nil {
		fmt.Fprintln(r.out, "(no recent-events buffer — host not connected?)")
		return
	}
	any := false
	if c := rec.Chat(); c != nil {
		fmt.Fprintf(r.out, "  [%s] chat        %s: %s\n", c.At.Format("15:04:05"), c.Speaker, c.Message)
		any = true
	}
	if p := rec.PM(); p != nil {
		fmt.Fprintf(r.out, "  [%s] pm          %s: %s\n", p.At.Format("15:04:05"), p.Sender, p.Message)
		any = true
	}
	if d := rec.Damage(); d != nil {
		fmt.Fprintf(r.out, "  [%s] damage      %d from %q\n", d.At.Format("15:04:05"), d.Amount, d.Source)
		any = true
	}
	if s := rec.ServerMessage(); s != nil {
		fmt.Fprintf(r.out, "  [%s] server msg  %s\n", s.At.Format("15:04:05"), s.Message)
		any = true
	}
	if dt := rec.DialogText(); dt != nil {
		fmt.Fprintf(r.out, "  [%s] npc dialog  %s\n", dt.At.Format("15:04:05"), dt.Text)
		any = true
	}
	if !any {
		fmt.Fprintln(r.out, "  (nothing yet — no transient events observed this session)")
	}
}

// metaLoad parses a .routine file and registers its procs +
// handlers into the session, without invoking the entry point.
// Useful for setting up a routine's helpers before exploring
// in the REPL.
func (r *REPL) metaLoad(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(r.out, "ERR: usage: .load <path>")
		return
	}
	rf, err := ParseRoutineFile(args[0])
	if err != nil {
		fmt.Fprintf(r.out, "ERR: %v\n", err)
		return
	}
	r.sess.LoadFile(rf.File)
	fmt.Fprintf(r.out, "loaded %s: %d procs, %d handlers, routine %q ready\n",
		rf.Path, len(rf.File.Procs), len(rf.File.Handlers), rf.File.Routine.Name)
}

// metaRun loads + immediately invokes the routine's entry point,
// printing the Result kind and value.
func (r *REPL) metaRun(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(r.out, "ERR: usage: .run <path>")
		return
	}
	rf, err := ParseRoutineFile(args[0])
	if err != nil {
		fmt.Fprintf(r.out, "ERR: %v\n", err)
		return
	}
	r.sess.LoadFile(rf.File)
	res := r.sess.Interpreter().RunRoutine(r.ctx, rf.File, nil)
	displayValue := "null"
	if res.Value != nil {
		displayValue = res.Value.Display()
	}
	fmt.Fprintf(r.out, "[%s] %s\n", res.Kind, displayValue)
	if res.Err != nil {
		fmt.Fprintf(r.out, "  err: %v\n", res.Err)
	}
}

// knownEnvNames extracts the visible bindings in the session env.
// Walks the chain and returns all names in scope. Used by .env.
func knownEnvNames(env *interp.Env) []string {
	return env.Names()
}

func isReservedDSLName(s string) bool {
	switch s {
	case "self", "host", "world", "inventory", "combat":
		return true
	}
	return false
}
