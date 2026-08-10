// Package mapping resolves and applies input mappings: trigger sequences that
// expand into special characters while typing answers, e.g. "e/" → "é".
//
// Mappings come from three layers, later layers winning:
//
//  1. Built-in defaults per language (embedded in the binary)
//  2. User files in ~/.config/crds/mappings/<lang>.yaml
//  3. A deck's input_mappings field
package mapping

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// builtin holds the shipped defaults keyed by language code.
var builtin = map[string]map[string]string{
	"fr": {
		"e/": "é",
		"e`": "è",
		"e^": "ê",
		`e"`: "ë",
		"a`": "à",
		"a^": "â",
		`a"`: "ä",
		"c,": "ç",
		"i^": "î",
		`i"`: "ï",
		"o^": "ô",
		`o"`: "ö",
		"oe": "œ",
		"u`": "ù",
		"u^": "û",
		`u"`: "ü",
		`y"`: "ÿ",
	},
}

// Mapping is an ordered set of trigger → replacement pairs. Keys are matched
// as the longest suffix of the input, so "e/" inside "mange/" expands to "é".
type Mapping struct {
	keys []string
	reps map[string]string
}

// New builds a Mapping from pairs, ordering keys longest-first so the longest
// matching suffix wins. Empty keys and keys equal to their replacement are
// dropped.
func New(pairs map[string]string) *Mapping {
	keys := make([]string, 0, len(pairs))
	reps := make(map[string]string, len(pairs))
	for k, v := range pairs {
		if k == "" || v == "" || k == v {
			continue
		}
		reps[k] = v
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		li, lj := len([]rune(keys[i])), len([]rune(keys[j]))
		if li != lj {
			return li > lj
		}
		return keys[i] < keys[j]
	})
	return &Mapping{keys: keys, reps: reps}
}

// Len returns the number of trigger pairs in the mapping.
func (m *Mapping) Len() int {
	if m == nil {
		return 0
	}
	return len(m.keys)
}

// Pair is a single trigger → replacement entry in a Mapping.
type Pair struct {
	Trigger     string
	Replacement string
}

// Pairs returns the mapping's trigger pairs sorted by trigger. A nil or
// empty mapping yields nil.
func (m *Mapping) Pairs() []Pair {
	if m == nil || len(m.keys) == 0 {
		return nil
	}
	out := make([]Pair, 0, len(m.keys))
	for _, k := range m.keys {
		out = append(out, Pair{Trigger: k, Replacement: m.reps[k]})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Trigger < out[j].Trigger
	})
	return out
}

// Apply replaces the longest key that is a suffix of input with its
// replacement, single pass. It returns input unchanged when no key matches.
func (m *Mapping) Apply(input string) string {
	out, _ := m.ApplyAt(input, len([]rune(input)))
	return out
}

// ApplyAt replaces the longest key that is a suffix of input[:end] (a rune
// index) with its replacement, single pass. It returns input unchanged and
// the same end when no key matches. The returned index is the rune position
// where the replacement ends, so callers can keep the cursor on the inserted
// text even when a multi-character trigger collapsed (e.g. "e/" → "é").
func (m *Mapping) ApplyAt(input string, end int) (string, int) {
	if m == nil || len(m.keys) == 0 {
		return input, end
	}
	runes := []rune(input)
	if end > len(runes) {
		end = len(runes)
	}
	for _, k := range m.keys {
		kr := []rune(k)
		if len(kr) > end {
			continue
		}
		start := end - len(kr)
		if equalRunes(runes[start:end], kr) {
			var b strings.Builder
			b.WriteString(string(runes[:start]))
			b.WriteString(m.reps[k])
			b.WriteString(string(runes[end:]))
			return b.String(), start + len([]rune(m.reps[k]))
		}
	}
	return input, end
}

func equalRunes(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Store holds built-in and user mappings keyed by language code.
type Store struct {
	user map[string]map[string]string
}

// NewStore returns a Store with only built-in defaults.
func NewStore() *Store {
	return &Store{user: map[string]map[string]string{}}
}

// LoadDir reads every *.yaml file in dir as a user mapping for the language
// named by its filename (e.g. fr.yaml → "fr"). A missing directory yields an
// empty store without error.
func LoadDir(dir string) (*Store, error) {
	s := NewStore()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read mappings dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		pairs, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		lang := strings.TrimSuffix(e.Name(), ".yaml")
		s.user[lang] = pairs
	}
	return s, nil
}

func loadFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var pairs map[string]string
	if err := yaml.Unmarshal(data, &pairs); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return pairs, nil
}

// Resolve returns the effective mapping for a language: built-in defaults
// overlaid by user mappings, then by deck mappings. An empty language falls
// back to the deck mappings alone.
func (s *Store) Resolve(lang string, deck map[string]string) *Mapping {
	pairs := map[string]string{}
	for k, v := range builtin[lang] {
		pairs[k] = v
	}
	if s != nil {
		for k, v := range s.user[lang] {
			pairs[k] = v
		}
	}
	for k, v := range deck {
		pairs[k] = v
	}
	return New(pairs)
}

// Languages returns the sorted set of language codes with built-in or user
// mappings.
func (s *Store) Languages() []string {
	seen := map[string]bool{}
	for lang := range builtin {
		seen[lang] = true
	}
	if s != nil {
		for lang := range s.user {
			seen[lang] = true
		}
	}
	out := make([]string, 0, len(seen))
	for lang := range seen {
		out = append(out, lang)
	}
	sort.Strings(out)
	return out
}
