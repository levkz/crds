package mapping

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestMappingApply(t *testing.T) {
	m := New(map[string]string{
		"e/": "é",
		"e^": "ê",
		"a/": "á",
		"ss": "ß",
	})
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no trigger", "hello", "hello"},
		{"simple trigger at end", "mange/", "mangé"},
		{"longest suffix wins", "e/e/", "e/é"},
		{"suffix at word end", "see/", "seé"},
		{"trigger mid-string untouched", "e/x", "e/x"},
		{"unicode already composed", "café", "café"},
		{"replacement can retrigger nothing", "é", "é"},
		{"single rune trigger", "a/", "á"},
		{"no accidental middle match", "shoe", "shoe"},
		{"multi-rune trigger", "kuss", "kuß"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Apply(tt.input); got != tt.want {
				t.Errorf("Apply(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMappingApplyAt(t *testing.T) {
	m := New(map[string]string{
		"e/": "é",
		"e^": "ê",
		"ss": "ß",
	})
	tests := []struct {
		name     string
		input    string
		end      int
		wantStr  string
		wantEnd  int
	}{
		{"empty", "", 0, "", 0},
		{"no match", "abc", 2, "abc", 2},
		{"match at end", "see/", 4, "seé", 3},
		{"mid-string match", "e/x", 2, "éx", 1},
		{"match collapses cursor", "x e/", 4, "x é", 3},
		{"old text before cursor untouched", "elle/il", 7, "elle/il", 7},
		{"end clamped", "ab", 10, "ab", 2},
		{"cursor at start", "é", 0, "é", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotEnd := m.ApplyAt(tt.input, tt.end)
			if gotStr != tt.wantStr || gotEnd != tt.wantEnd {
				t.Errorf("ApplyAt(%q, %d) = (%q, %d), want (%q, %d)",
					tt.input, tt.end, gotStr, gotEnd, tt.wantStr, tt.wantEnd)
			}
		})
	}
}

func TestMappingPairs(t *testing.T) {
	m := New(map[string]string{
		"e^": "ê",
		"e/": "é",
		"ss": "ß",
	})
	pairs := m.Pairs()
	want := []Pair{
		{"e/", "é"},
		{"e^", "ê"},
		{"ss", "ß"},
	}
	if !reflect.DeepEqual(pairs, want) {
		t.Errorf("Pairs() = %v, want %v", pairs, want)
	}
	if got := New(nil).Pairs(); got != nil {
		t.Errorf("New(nil).Pairs() = %v, want nil", got)
	}
	var nilM *Mapping
	if got := nilM.Pairs(); got != nil {
		t.Errorf("nil Mapping.Pairs() = %v, want nil", got)
	}
}

func TestMappingNewFilters(t *testing.T) {
	m := New(map[string]string{
		"":    "x", // empty key dropped
		"e/":  "",  // empty replacement dropped
		"aa":  "aa", // self-identity dropped
		"good": "ok",
	})
	if m.Len() != 1 {
		t.Fatalf("expected 1 kept pair, got %d", m.Len())
	}
	if got := m.Apply("say good"); got != "say ok" {
		t.Errorf("Apply = %q, want %q", got, "say ok")
	}
}

func TestMappingLongestSuffixDeterministic(t *testing.T) {
	m := New(map[string]string{
		"e/":  "é",
		"ee/": "éé",
	})
	if got := m.Apply("see/"); got != "séé" {
		t.Errorf("longest match failed: Apply = %q, want %q", got, "séé")
	}
}

func TestStoreResolvePrecedence(t *testing.T) {
	s := NewStore()
	s.user["fr"] = map[string]string{
		"e/": "é",
		"q/": "q!",
	}

	m := s.Resolve("fr", map[string]string{"e/": "e/", "x:": "×"})

	// User's q/ override built-in-less addition is present.
	if got := m.Apply("q/"); got != "q!" {
		t.Errorf("user mapping not applied: got %q, want %q", got, "q!")
	}
	// Deck override of e/ wins over user + builtin.
	if got := m.Apply("e/"); got != "e/" {
		t.Errorf("deck mapping should win: got %q, want %q", got, "e/")
	}
	// Built-in (no user override) still present.
	if got := m.Apply("c,"); got != "ç" {
		t.Errorf("builtin mapping not applied: got %q, want %q", got, "ç")
	}
	// Deck-only mapping present.
	if got := m.Apply("x:"); got != "×" {
		t.Errorf("deck-only mapping not applied: got %q, want %q", got, "×")
	}
}

func TestStoreResolveUnknownLanguage(t *testing.T) {
	s := NewStore()
	s.user["de"] = map[string]string{"ss": "ß"}
	m := s.Resolve("de", nil)
	if got := m.Apply("straße"); got != "straße" {
		t.Errorf("expected unchanged, got %q", got)
	}
	if got := m.Apply("kuss"); got != "kuß" {
		t.Errorf("expected kuß, got %q", got)
	}
	m2 := s.Resolve("xx", nil)
	if m2.Len() != 0 {
		t.Errorf("unknown language should have no mappings, got %d", m2.Len())
	}
}

func TestStoreResolveNil(t *testing.T) {
	var s *Store
	m := s.Resolve("fr", nil)
	if m.Len() == 0 {
		t.Errorf("nil store should still resolve builtin defaults")
	}
}

func TestBuiltinFrenchOELigature(t *testing.T) {
	m := NewStore().Resolve("fr", nil)
	if got := m.Apply("oe"); got != "œ" {
		t.Errorf("builtin fr: Apply(oe) = %q, want %q", got, "œ")
	}
	if got := m.Apply("coe"); got != "cœ" {
		t.Errorf("builtin fr: Apply(coe) = %q, want %q", got, "cœ")
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "fr.yaml"), "'e/': 'é'\n'e\"': 'è'\n")
	writeFile(t, filepath.Join(dir, "de.yaml"), "'ss': 'ß'\n")
	writeFile(t, filepath.Join(dir, "ignore.txt"), "not yaml\n")

	s, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}

	languages := s.Languages()
	sort.Strings(languages)
	want := []string{"de", "fr"}
	if !reflect.DeepEqual(languages, want) {
		t.Errorf("Languages() = %v, want %v", languages, want)
	}

	m := s.Resolve("fr", nil)
	if got := m.Apply("e/"); got != "é" {
		t.Errorf("fr user mapping not loaded: got %q", got)
	}
	if got := m.Apply(`e"`); got != "è" {
		t.Errorf("fr quoted user mapping not loaded: got %q", got)
	}
}

func TestLoadDirMissing(t *testing.T) {
	s, err := LoadDir(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("LoadDir(missing) = %v", err)
	}
	if len(s.Languages()) != 1 {
		t.Errorf("expected only builtin fr, got %v", s.Languages())
	}
}

func TestLoadDirMalformed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "fr.yaml"), "e/: [not, a, map]\n")
	if _, err := LoadDir(dir); err == nil {
		t.Errorf("expected error for malformed mapping file")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}
