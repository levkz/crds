package storage

import (
	"os"
	"path/filepath"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestStateStore_Load_FileNotExists(t *testing.T) {
	s := NewStateStore(t.TempDir())

	state, err := s.Load()
	if err != nil {
		t.Fatalf("Load on nonexistent file: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil State")
	}
	if len(state.SelectedDecks) != 0 {
		t.Errorf("expected empty SelectedDecks, got %v", state.SelectedDecks)
	}
}

func TestStateStore_Load_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.yaml")
	content := []byte("selected_decks:\n  - french_a1\n  - japanese_n5\n")
	if err := os.WriteFile(statePath, content, 0644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	s := NewStateStore(dir)
	state, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if state == nil {
		t.Fatal("expected non-nil State")
	}
	want := []string{"french_a1", "japanese_n5"}
	if len(state.SelectedDecks) != len(want) {
		t.Fatalf("expected %d decks, got %d", len(want), len(state.SelectedDecks))
	}
	for i, d := range state.SelectedDecks {
		if d != want[i] {
			t.Errorf("deck[%d]: expected %q, got %q", i, want[i], d)
		}
	}
}

func TestStateStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewStateStore(dir)

	original := &State{
		SelectedDecks: []string{"spanish_a1", "german_a2"},
	}
	if err := s.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil State")
	}
	if len(loaded.SelectedDecks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(loaded.SelectedDecks))
	}
	if loaded.SelectedDecks[0] != "spanish_a1" || loaded.SelectedDecks[1] != "german_a2" {
		t.Errorf("got %v, want [spanish_a1 german_a2]", loaded.SelectedDecks)
	}
}

func TestStateStore_Save_CreatesDir(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "nonexistent", "subdir")
	s := NewStateStore(sub)

	state := &State{SelectedDecks: []string{"test_deck"}}
	if err := s.Save(state); err != nil {
		t.Fatalf("Save creating directories: %v", err)
	}

	if _, err := os.Stat(filepath.Join(sub, "state.yaml")); os.IsNotExist(err) {
		t.Fatal("state.yaml was not created")
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load after save with created dir: %v", err)
	}
	if len(loaded.SelectedDecks) != 1 || loaded.SelectedDecks[0] != "test_deck" {
		t.Errorf("got %v, want [test_deck]", loaded.SelectedDecks)
	}
}

func TestStateStore_Load_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("invalid: [yaml: \n"), 0644); err != nil {
		t.Fatalf("write invalid state: %v", err)
	}

	s := NewStateStore(dir)
	_, err := s.Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestStateStore_Save_WritesCorrectYAML(t *testing.T) {
	dir := t.TempDir()
	s := NewStateStore(dir)

	state := &State{
		SelectedDecks: []string{"deck_a", "deck_b"},
	}
	if err := s.Save(state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		t.Fatalf("read state.yaml: %v", err)
	}

	var parsed State
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal saved yaml: %v", err)
	}
	if len(parsed.SelectedDecks) != 2 {
		t.Fatalf("expected 2 decks, got %d", len(parsed.SelectedDecks))
	}
	if parsed.SelectedDecks[0] != "deck_a" || parsed.SelectedDecks[1] != "deck_b" {
		t.Errorf("got %v, want [deck_a deck_b]", parsed.SelectedDecks)
	}
}
