package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRevertReserve verifies a full backup→modify→revert cycle.
func TestRevertReserve(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)

	// Set up initial state: state.yaml + a deck
	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("theme: dark\nselected_decks:\n  - test_deck\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	decksDir := filepath.Join(sharedDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		t.Fatalf("mkdir decks: %v", err)
	}
	writeTestDeck(t, decksDir)
	if err := store.SyncDecks(decksDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	// A config mapping file that must survive the revert.
	writeMappings(t, configDir)

	// Record some progress
	if err := store.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_1", 4, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	// Create a backup
	if err := store.CreateReserve(sharedDir, configDir); err != nil {
		t.Fatalf("CreateReserve: %v", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, _ := os.ReadDir(reserveDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(entries))
	}
	reservePath := filepath.Join(reserveDir, entries[0].Name())

	// Now modify the state: add more reviews, change state.yaml, overwrite the mapping
	if err := store.RecordAnswer("test_deck", "entry_2", 1, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_1", 2, true); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	stateBefore := store.Stats()
	if stateBefore.ReviewedToday != 4 {
		t.Fatalf("expected 4 reviews before revert, got %d", stateBefore.ReviewedToday)
	}

	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("theme: light\nselected_decks: []\n"), 0644); err != nil {
		t.Fatalf("rewrite state.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "mappings", "fr.yaml"), []byte("oe: œ\n"), 0644); err != nil {
		t.Fatalf("overwrite mapping: %v", err)
	}

	// Revert
	if err := store.RevertReserve(sharedDir, configDir, reservePath); err != nil {
		t.Fatalf("RevertReserve: %v", err)
	}

	// Verify state.yaml is restored
	stateData, err := os.ReadFile(filepath.Join(sharedDir, "state.yaml"))
	if err != nil {
		t.Fatalf("read state.yaml: %v", err)
	}
	if !strings.Contains(string(stateData), "theme: dark") {
		t.Errorf("expected theme: dark after revert, got: %s", stateData)
	}

	// Verify the deck YAML still exists
	if _, err := os.Stat(filepath.Join(decksDir, "test_deck.yaml")); os.IsNotExist(err) {
		t.Errorf("deck YAML missing after revert")
	}

	// Verify the mapping file was restored to the backup's content
	mappingData, err := os.ReadFile(filepath.Join(configDir, "mappings", "fr.yaml"))
	if err != nil {
		t.Fatalf("read restored mapping: %v", err)
	}
	if string(mappingData) != "e/: é\n" {
		t.Errorf("mapping content after revert = %q, want %q", mappingData, "e/: é\n")
	}

	// Verify the DB was restored (stats should be at 2 reviews, not 4)
	stateAfter := store.Stats()
	if stateAfter.ReviewedToday != 2 {
		t.Errorf("expected 2 reviews after revert, got %d", stateAfter.ReviewedToday)
	}

	// Verify a pre-revert backup was created (it should have 3 reviews)
	preEntries, _ := os.ReadDir(reserveDir)
	if len(preEntries) != 2 {
		t.Errorf("expected 2 backup files (original + pre-revert), got %d", len(preEntries))
	}
}

// TestRevertReserve_InvalidArchive verifies rejection of archives without crds.db.
func TestRevertReserve_InvalidArchive(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)

	// Create a dummy tar.gz without crds.db
	badPath := filepath.Join(t.TempDir(), "bad.tar.gz")
	badData := []byte("this is not a valid gzip file")
	if err := os.WriteFile(badPath, badData, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := store.RevertReserve(sharedDir, configDir, badPath)
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

// TestRevertReserve_StoreUsableAfterRevert verifies the store works after revert.
func TestRevertReserve_StoreUsableAfterRevert(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)

	// Set up minimal state
	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	if err := store.CreateReserve(sharedDir, configDir); err != nil {
		t.Fatalf("CreateReserve: %v", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, _ := os.ReadDir(reserveDir)
	reservePath := filepath.Join(reserveDir, entries[0].Name())

	if err := store.RevertReserve(sharedDir, configDir, reservePath); err != nil {
		t.Fatalf("RevertReserve: %v", err)
	}

	if err := store.RecordAnswer("test", "entry", 3, false); err != nil {
		t.Fatalf("RecordAnswer after revert: %v", err)
	}
	stats := store.Stats()
	if stats.ReviewedToday != 1 {
		t.Errorf("expected 1 review after revert+record, got %d", stats.ReviewedToday)
	}
}

// TestRevertReserve_MissingArchive verifies clean error for nonexistent file.
func TestRevertReserve_MissingArchive(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)

	err := store.RevertReserve(sharedDir, configDir, "/nonexistent/path.tar.gz")
	if err == nil {
		t.Fatal("expected error for nonexistent archive")
	}
}
