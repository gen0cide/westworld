package resolve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// aliasEntry is one learned mapping in the store's on-disk form.
//
// Text is the loose, player-spoken phrase the host learned ("r2h").
// Canonical is the exact facts name it resolves to ("Rune 2-handed Sword").
// Kind is the catalog the canonical name lives in ("item", "npc", "loc",
// "spell", "prayer"); it lets the same text disambiguate across catalogs and
// lets a kind-filtered resolve() ignore mappings from other catalogs.
type aliasEntry struct {
	Text      string `json:"text"`
	Canonical string `json:"canonical"`
	Kind      string `json:"kind"`
}

// aliasFile is the JSON document persisted to disk.
type aliasFile struct {
	// Version lets us migrate the format later without guessing.
	Version int          `json:"version"`
	Aliases []aliasEntry `json:"aliases"`
}

const aliasFileVersion = 1

// aliasKey is the normalized lookup key. We normalize both the learned text
// and the query the same way so "R2H", "r2h" and " r 2 h " collapse to one
// bucket. Kind is part of the key so the same text can map to different
// canonical names in different catalogs without collision.
type aliasKey struct {
	text string // normalized
	kind string // "" means "any kind"
}

// AliasStore is the host's persisted learned-alias table (text→canonical).
//
// It is the first stage of the resolve() pipeline and the write-back target
// of the brain fallback. It is safe for concurrent use across a host's
// routines. Construct one with NewAliasStore (in-memory) or LoadAliasStore
// (JSON-backed); seed it with Seed; grow it with Learn.
//
// Persistence is opportunistic: Learn writes through to disk when a path is
// configured, but a write error never fails the learn — the in-memory table
// is the source of truth for the running host, and a failed flush is reported
// (so callers can log it) without dropping the learned mapping.
type AliasStore struct {
	mu sync.RWMutex
	// byKey maps (normalized-text, kind) → canonical name. We keep a
	// kind-specific bucket and a kind-agnostic ("") bucket so a resolve with
	// no kind filter still gets a fast-path hit.
	byKey map[aliasKey]string
	// order preserves insertion order for stable, deterministic persistence
	// and iteration (tests + diffable JSON).
	order []aliasEntry

	// path is the JSON file backing the store; empty means in-memory only.
	path string
}

// NewAliasStore returns an empty, in-memory alias store (no persistence).
func NewAliasStore() *AliasStore {
	return &AliasStore{byKey: map[aliasKey]string{}}
}

// LoadAliasStore opens (or creates) a JSON-backed per-host alias store at
// path. A missing file is not an error — it yields an empty store that will
// be created on the first successful flush. A present-but-corrupt file IS an
// error, so a host never silently discards its learned lingo.
func LoadAliasStore(path string) (*AliasStore, error) {
	s := &AliasStore{byKey: map[aliasKey]string{}, path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	var doc aliasFile
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	for _, e := range doc.Aliases {
		s.insert(e.Text, e.Canonical, e.Kind)
	}
	return s, nil
}

// Seed bulk-loads mappings (e.g. a persona's starting vocabulary) without
// touching disk. It is additive; later seeds and Learns override earlier
// entries for the same (text, kind). Returns the receiver for chaining.
func (s *AliasStore) Seed(entries map[string]SeedAlias) *AliasStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Iterate in sorted text order so seeding is deterministic.
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, text := range keys {
		v := entries[text]
		s.insert(text, v.Canonical, v.Kind)
	}
	return s
}

// SeedAlias is one entry for Seed: the canonical name and which catalog kind
// it belongs to.
type SeedAlias struct {
	Canonical string
	Kind      string
}

// Lookup returns the canonical name learned for text within kind, best-effort.
// kind == "" matches any catalog. The bool reports whether a mapping was
// found. Lookup never mutates the store.
func (s *AliasStore) Lookup(text, kind string) (string, bool) {
	n := normalize(text)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if kind != "" {
		if c, ok := s.byKey[aliasKey{text: n, kind: kind}]; ok {
			return c, true
		}
		return "", false
	}
	// No kind filter: prefer the kind-agnostic bucket, then fall back to any
	// kind-specific bucket for this text (deterministic by insertion order).
	if c, ok := s.byKey[aliasKey{text: n, kind: ""}]; ok {
		return c, true
	}
	for _, e := range s.order {
		if normalize(e.Text) == n {
			return e.Canonical, true
		}
	}
	return "", false
}

// Learn records text→canonical for the given kind and, if the store is
// JSON-backed, flushes to disk. The returned error is only ever a persistence
// (flush) error; the in-memory mapping is always recorded regardless, so a
// host that cannot write its file still benefits from the learning for the
// rest of the session.
func (s *AliasStore) Learn(text, canonical, kind string) error {
	s.mu.Lock()
	s.insert(text, canonical, kind)
	path := s.path
	doc := s.snapshotLocked()
	s.mu.Unlock()
	if path == "" {
		return nil
	}
	return flush(path, doc)
}

// Len reports the number of distinct (text, kind) mappings stored.
func (s *AliasStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.order)
}

// insert records/overwrites a mapping. Caller holds the write lock. It keeps
// both a kind-specific and a kind-agnostic bucket pointed at the canonical so
// that an unfiltered resolve still fast-paths. order is updated in place when
// the same (text, kind) is re-learned, preserving deterministic output.
func (s *AliasStore) insert(text, canonical, kind string) {
	n := normalize(text)
	s.byKey[aliasKey{text: n, kind: kind}] = canonical
	// Maintain the kind-agnostic bucket: last writer for this text wins,
	// which matches "the most recently learned meaning is the default."
	s.byKey[aliasKey{text: n, kind: ""}] = canonical
	for i := range s.order {
		if normalize(s.order[i].Text) == n && s.order[i].Kind == kind {
			s.order[i].Canonical = canonical
			s.order[i].Text = text
			return
		}
	}
	s.order = append(s.order, aliasEntry{Text: text, Canonical: canonical, Kind: kind})
}

// snapshotLocked builds the persistable document. Caller holds a lock.
func (s *AliasStore) snapshotLocked() aliasFile {
	out := aliasFile{Version: aliasFileVersion, Aliases: make([]aliasEntry, len(s.order))}
	copy(out.Aliases, s.order)
	return out
}

// flush atomically writes doc to path (temp file + rename), creating parent
// directories as needed.
func flush(path string, doc aliasFile) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// normalize folds a phrase to its canonical comparison form: lower-cased,
// punctuation→space, runs of whitespace collapsed to single spaces, trimmed.
// "Rune 2-handed Sword" and "rune 2 handed sword" normalize identically.
func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSpace = false
		default:
			// Treat every non-alphanumeric rune (space, '-', '_', etc.) as a
			// separator so token boundaries are uniform.
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// tokens splits a normalized phrase into its words.
func tokens(s string) []string {
	n := normalize(s)
	if n == "" {
		return nil
	}
	return strings.Fields(n)
}
