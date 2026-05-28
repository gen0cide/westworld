package interp

// Env is a chained name → Value map used as the lexical environment
// for an executing routine. Each block/loop introduces a child Env;
// lookups walk up the chain.
type Env struct {
	parent *Env
	vars   map[string]Value
}

// NewEnv returns a fresh root environment.
func NewEnv() *Env { return &Env{vars: map[string]Value{}} }

// Child returns a new environment chained off this one.
func (e *Env) Child() *Env { return &Env{parent: e, vars: map[string]Value{}} }

// Define unconditionally binds name → val in the current scope.
// Used for parameter binding and `for x in ...` loop variables.
func (e *Env) Define(name string, val Value) { e.vars[name] = val }

// Set assigns to an existing binding if found in this scope or an
// ancestor, otherwise creates the binding in the current scope.
// This matches the DSL's lack of explicit `let` — assignment creates
// or updates.
func (e *Env) Set(name string, val Value) {
	for c := e; c != nil; c = c.parent {
		if _, ok := c.vars[name]; ok {
			c.vars[name] = val
			return
		}
	}
	e.vars[name] = val
}

// Get returns the binding for name, walking ancestors.
func (e *Env) Get(name string) (Value, bool) {
	for c := e; c != nil; c = c.parent {
		if v, ok := c.vars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Names returns every binding name in this env's chain. Used by
// tooling (REPL `.env`, debug dumps); not for hot paths. Order is
// "innermost scope first" within a level; level order is arbitrary
// since callers typically sort.
func (e *Env) Names() []string {
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for c := e; c != nil; c = c.parent {
		for name := range c.vars {
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}
