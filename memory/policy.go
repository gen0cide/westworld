package memory

import (
	"strings"
	"time"
)

// ReadMode selects how a Get traverses the tiers.
type ReadMode uint8

const (
	// CascadeCached probes scratch → local → remote and accepts the first hit;
	// a remote hit is promoted back down. The default for most data.
	CascadeCached ReadMode = iota
	// CascadeFresh bypasses the cache tiers for the authority tier so the
	// caller always sees the source of truth (then promotes it down).
	CascadeFresh
	// LocalOnly probes scratch → local and never touches the network.
	LocalOnly
	// RemoteOnly goes straight to the remote tier (capability-exclusive data).
	RemoteOnly
)

// WriteMode selects how a Put fans out across the tiers.
type WriteMode uint8

const (
	// WriteAround writes scratch only — hot, ephemeral state, never persisted.
	WriteAround WriteMode = iota
	// WriteBack writes local now and queues the remote write in the journal
	// (drained later / when remote is healthy). Caches in scratch too.
	WriteBack
	// WriteThrough writes local AND remote synchronously (must-not-lose).
	WriteThrough
	// RemoteAuthoritative writes remote (or journals it when remote is down)
	// and keeps a local+scratch cached projection.
	RemoteAuthoritative
)

// Authority names which tier wins on conflict / counts as "fresh".
type Authority uint8

const (
	AuthLocal  Authority = iota // this host's private truth (relationship, goal)
	AuthRemote                  // cross-host / of-record truth (reputation, episodic)
)

// Class is the policy for one namespace: how its keys are read, written, who
// owns them, and the cache/remote timing.
type Class struct {
	Read           ReadMode
	Write          WriteMode
	Authority      Authority
	TTL            time.Duration // promotion + negative-cache lifetime in scratch
	RemoteDeadline time.Duration // base sync deadline for a remote read (scaled by maturity)
}

// Hint is an optional per-call override. It is the highest-priority input in
// policy resolution, ahead of the namespace table and the global default.
type Hint struct {
	Fresh     bool // force a fresh read past the cache tiers
	Ephemeral bool // force a write-around (scratch only)
	Durable   bool // force at least a write-back (local) for an otherwise-ephemeral class
}

// Policy is the per-namespace table plus a conservative default.
type Policy struct {
	Classes map[string]Class
	Default Class
}

// DefaultPolicy returns the starting policy table. Newborn hosts get generous
// remote deadlines (they phone home a lot and wait patiently); the maturity
// dial tightens these as local tiers warm up.
func DefaultPolicy() Policy {
	const patient = 5 * time.Second
	return Policy{
		Default: Class{Read: CascadeCached, Write: WriteBack, Authority: AuthLocal, TTL: 2 * time.Minute, RemoteDeadline: patient},
		Classes: map[string]Class{
			// Hot ephemeral working memory — never leaves RAM.
			"scratch": {Read: LocalOnly, Write: WriteAround, Authority: AuthLocal},
			// This host's felt trust — local truth, pushed up lazily.
			"relationship": {Read: CascadeCached, Write: WriteBack, Authority: AuthLocal, TTL: 5 * time.Minute, RemoteDeadline: patient},
			// Goal/progress — local truth.
			"goal": {Read: CascadeCached, Write: WriteBack, Authority: AuthLocal, TTL: 2 * time.Minute, RemoteDeadline: patient},
			// Episodic JOURNAL — the host's durable log of what it did this life
			// (level-ups, kills, deaths, objective milestones) + its standing
			// objective. Local truth, read fast from disk; write-back queues the
			// mirror-up to mesa's LTM (the per-event semantic-recall corpus lands
			// with the "episodic" namespace + mesa Knowledge.Recall, task #13).
			"journal": {Read: CascadeCached, Write: WriteBack, Authority: AuthLocal, TTL: 2 * time.Minute, RemoteDeadline: patient},
			// Semantic WORLD-KNOWLEDGE ledger — graded, provenance-tagged beliefs
			// about things (npcs/places/shops/items/mechanics). Local truth;
			// write-back queues the mirror to mesa. See cognition/knowledge.
			"knowledge": {Read: CascadeCached, Write: WriteBack, Authority: AuthLocal, TTL: 5 * time.Minute, RemoteDeadline: patient},
			// Global reputation — remote truth, pulled down fresh, short cache.
			"reputation": {Read: CascadeFresh, Write: RemoteAuthoritative, Authority: AuthRemote, TTL: 1 * time.Minute, RemoteDeadline: patient},
			// Episodic events / reflections — remote of record.
			"episodic": {Read: RemoteOnly, Write: RemoteAuthoritative, Authority: AuthRemote, RemoteDeadline: patient},
		},
	}
}

// namespaceOf returns the portion of key before the first ':' (or the whole
// key when there is no ':'). "relationship:alex" → "relationship".
func namespaceOf(key string) string {
	ns, _, _ := strings.Cut(key, ":")
	return ns
}

// classFor returns the Class registered for key's namespace, or the default.
func (p Policy) classFor(key string) Class {
	if c, ok := p.Classes[namespaceOf(key)]; ok {
		return c
	}
	return p.Default
}

// resolve applies a hint (if any) over the namespace class. Resolution order:
// hint > namespace table > default.
func (p Policy) resolve(key string, hints ...Hint) Class {
	c := p.classFor(key)
	if len(hints) == 0 {
		return c
	}
	h := hints[0]
	if h.Fresh && c.Read != RemoteOnly {
		c.Read = CascadeFresh
	}
	if h.Ephemeral {
		c.Write = WriteAround
	} else if h.Durable && c.Write == WriteAround {
		c.Write = WriteBack
	}
	return c
}
