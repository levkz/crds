package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crds/internal/model"
	"crds/internal/parser"
	"crds/internal/storage/db"

	"go.yaml.in/yaml/v3"
)

const testDeckYAML = `id: test_deck
name: Test Deck
language: fr
translation_language: en

entries:
  - id: entry_1
    term: bonjour
    translations:
      - text: hello
      - text: good morning
    tags:
      - greeting

  - id: entry_2
    term: au revoir
    translations:
      - text: goodbye
    examples:
      - text: Au revoir, Marie.
        translation: Goodbye, Marie.
`

func writeTestDeck(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "test_deck.yaml")
	if err := os.WriteFile(path, []byte(testDeckYAML), 0644); err != nil {
		t.Fatalf("write test deck: %v", err)
	}
	return path
}

func setupSyncedDeck(t *testing.T, store *Store) string {
	t.Helper()
	deckDir := t.TempDir()
	writeTestDeck(t, deckDir)
	if err := store.SyncDecks(deckDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}
	return deckDir
}

// --- ImportDeck ---

func TestImportDeck(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	srcDir := t.TempDir()
	deckDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "source.yaml")
	if err := os.WriteFile(srcPath, []byte(testDeckYAML), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := store.ImportDeck(srcPath, deckDir); err != nil {
		t.Fatalf("ImportDeck: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "test_deck")
	if err != nil {
		t.Errorf("deck should exist in DB: %v", err)
	}

	dstPath := filepath.Join(deckDir, "test_deck.yaml")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("file %q should exist after import", dstPath)
	}
}

func TestImportDeck_DuplicateID(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "source.yaml")
	if err := os.WriteFile(srcPath, []byte(testDeckYAML), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	err := store.ImportDeck(srcPath, deckDir)
	if err == nil {
		t.Fatal("expected error for duplicate import")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestImportDeck_InvalidYAML(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "bad.yaml")
	if err := os.WriteFile(srcPath, []byte("invalid: [yaml: broken"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	err := store.ImportDeck(srcPath, t.TempDir())
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// --- ExportDeck ---

func TestExportDeck(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	dstPath := filepath.Join(t.TempDir(), "exported.yaml")
	if err := store.ExportDeck("test_deck", dstPath, deckDir); err != nil {
		t.Fatalf("ExportDeck: %v", err)
	}

	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if !strings.Contains(string(data), "id: test_deck") {
		t.Errorf("exported file missing deck ID")
	}
	if !strings.Contains(string(data), "bonjour") {
		t.Errorf("exported file missing entry term")
	}
}

func TestExportDeck_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	err := store.ExportDeck("nonexistent", "/tmp/out.yaml", t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// --- ExportDeckFromCache ---

func TestExportDeckFromCache(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	dstPath := filepath.Join(t.TempDir(), "cached.yaml")
	if err := store.ExportDeckFromCache("test_deck", dstPath); err != nil {
		t.Fatalf("ExportDeckFromCache: %v", err)
	}

	exported, err := parser.ParseFile(dstPath)
	if err != nil {
		t.Fatalf("parse exported file: %v", err)
	}

	if exported.ID != "test_deck" {
		t.Errorf("expected id 'test_deck', got %q", exported.ID)
	}
	if exported.Name != "Test Deck" {
		t.Errorf("expected name 'Test Deck', got %q", exported.Name)
	}
	if len(exported.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(exported.Entries))
	}

	// Remove the source file and verify cache still works
	os.Remove(filepath.Join(deckDir, "test_deck.yaml"))
	dstPath2 := filepath.Join(t.TempDir(), "cached2.yaml")
	if err := store.ExportDeckFromCache("test_deck", dstPath2); err != nil {
		t.Fatalf("ExportDeckFromCache after source deleted: %v", err)
	}
}

func TestExportDeckFromCache_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	err := store.ExportDeckFromCache("nonexistent", "/tmp/out.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

// --- RenameDeck ---

func TestRenameDeck(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	if err := store.RenameDeck("test_deck", "Renamed Deck", deckDir); err != nil {
		t.Fatalf("RenameDeck: %v", err)
	}

	d, err := store.queries.GetDeck(context.Background(), "test_deck")
	if err != nil {
		t.Fatalf("GetDeck: %v", err)
	}
	if d.Name != "Renamed Deck" {
		t.Errorf("expected name 'Renamed Deck', got %q", d.Name)
	}

	yamlPath := filepath.Join(deckDir, "test_deck.yaml")
	parsed, err := parser.ParseFile(yamlPath)
	if err != nil {
		t.Fatalf("parse YAML after rename: %v", err)
	}
	if parsed.Name != "Renamed Deck" {
		t.Errorf("expected YAML name 'Renamed Deck', got %q", parsed.Name)
	}
}

func TestRenameDeck_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := t.TempDir()
	err := store.RenameDeck("nonexistent", "New Name", deckDir)
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

// --- ChangeDeckID ---

func TestChangeDeckID(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	if err := store.ChangeDeckID("test_deck", "renamed_deck", deckDir); err != nil {
		t.Fatalf("ChangeDeckID: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "renamed_deck")
	if err != nil {
		t.Errorf("deck should exist with new ID: %v", err)
	}

	_, err = store.queries.GetDeck(context.Background(), "test_deck")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("old deck should be gone: %v", err)
	}

	newPath := filepath.Join(deckDir, "renamed_deck.yaml")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("new YAML file should exist")
	}

	oldPath := filepath.Join(deckDir, "test_deck.yaml")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("old YAML file should be removed")
	}

	entries, err := store.queries.ListEntriesByDeck(context.Background(), "renamed_deck")
	if err != nil {
		t.Fatalf("ListEntriesByDeck: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.DeckID != "renamed_deck" {
			t.Errorf("entry %q has deck_id %q, expected %q", e.ID, e.DeckID, "renamed_deck")
		}
	}
}

func TestChangeDeckID_Duplicate(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	// Create a second deck with a different ID
	deck2YAML := `id: existing_deck
name: Existing
language: fr
translation_language: en
entries:
  - id: e1
    term: hello
    translations:
      - text: bonjour
`
	path2 := filepath.Join(deckDir, "existing_deck.yaml")
	if err := os.WriteFile(path2, []byte(deck2YAML), 0644); err != nil {
		t.Fatalf("write deck2: %v", err)
	}
	if err := store.SyncDecks(deckDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	err := store.ChangeDeckID("test_deck", "existing_deck", deckDir)
	if err == nil {
		t.Fatal("expected error for duplicate new ID")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestChangeDeckID_WithProgress(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)
	ctx := context.Background()

	due := timeOnly("2026-07-26 12:00:00")

	// Insert progress for both entries
	for _, entryID := range []string{"entry_1", "entry_2"} {
		err := store.queries.UpsertProgress(ctx, db.UpsertProgressParams{
			DeckID:   "test_deck",
			EntryID:  entryID,
			Reverse:  0,
			Ease:     2.5,
			Interval: 1,
			Due:      &due,
			Correct:  3,
			Incorrect: 0,
		})
		if err != nil {
			t.Fatalf("UpsertProgress(%q): %v", entryID, err)
		}
	}

	// Record reviews
	if err := store.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_2", 2, true); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if err := store.ChangeDeckID("test_deck", "renamed_deck", deckDir); err != nil {
		t.Fatalf("ChangeDeckID: %v", err)
	}

	// Verify progress now points to new ID
	for _, entryID := range []string{"entry_1", "entry_2"} {
		p, err := store.queries.GetProgress(ctx, db.GetProgressParams{
			DeckID:  "renamed_deck",
			EntryID: entryID,
			Reverse: 0,
		})
		if err != nil {
			t.Errorf("GetProgress(%q) after rename: %v", entryID, err)
		} else if p.Correct != 3 {
			t.Errorf("expected correct=3 for %q, got %d", entryID, p.Correct)
		}
	}

	// Verify reviews now point to new ID
	reviews1, err := store.GetReviewsByEntry("entry_1", 10)
	if err != nil {
		t.Fatalf("GetReviewsByEntry: %v", err)
	}
	for _, r := range reviews1 {
		if r.DeckID != "renamed_deck" {
			t.Errorf("review deck_id %q, expected %q", r.DeckID, "renamed_deck")
		}
	}
}

// TestExportDeckFromCache_RoundTrip verifies that a cached export can be imported
// into a fresh store.
func TestExportDeckFromCache_RoundTrip(t *testing.T) {
	store1 := newTestStore(t)
	defer store1.Close()

	setupSyncedDeck(t, store1)

	dstPath := filepath.Join(t.TempDir(), "roundtrip.yaml")
	if err := store1.ExportDeckFromCache("test_deck", dstPath); err != nil {
		t.Fatalf("ExportDeckFromCache: %v", err)
	}

	store2 := newTestStore(t)
	defer store2.Close()

	importDir := t.TempDir()
	if err := store2.ImportDeck(dstPath, importDir); err != nil {
		t.Fatalf("ImportDeck into fresh store: %v", err)
	}

	decks, err := store2.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks) != 1 || decks[0] != "test_deck" {
		t.Errorf("expected [test_deck] in fresh store, got %v", decks)
	}
}

// TestRenameDeck_Resync verifies that SyncDecks after rename is a no-op.
func TestRenameDeck_Resync(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	if err := store.RenameDeck("test_deck", "Renamed Deck", deckDir); err != nil {
		t.Fatalf("RenameDeck: %v", err)
	}

	// Sync should be a no-op since mtime hasn't changed (file was just written)
	if err := store.SyncDecks(deckDir); err != nil {
		t.Fatalf("SyncDecks after rename: %v", err)
	}

	d, err := store.queries.GetDeck(context.Background(), "test_deck")
	if err != nil {
		t.Fatalf("GetDeck after resync: %v", err)
	}
	if d.Name != "Renamed Deck" {
		t.Errorf("expected name 'Renamed Deck' after resync, got %q", d.Name)
	}
}

// TestChangeDeckID_Resync verifies that SyncDecks after ID change is a no-op.
func TestChangeDeckID_Resync(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	if err := store.ChangeDeckID("test_deck", "renamed_deck", deckDir); err != nil {
		t.Fatalf("ChangeDeckID: %v", err)
	}

	if err := store.SyncDecks(deckDir); err != nil {
		t.Fatalf("SyncDecks after ID change: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "renamed_deck")
	if err != nil {
		t.Errorf("deck should still exist after resync: %v", err)
	}
}

// TestImportDeck_ImportToEmptyDir verifies importing without prior SyncDecks.
func TestImportDeck_ImportToEmptyDir(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "source.yaml")
	if err := os.WriteFile(srcPath, []byte(testDeckYAML), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	deckDir := t.TempDir()
	if err := store.ImportDeck(srcPath, deckDir); err != nil {
		t.Fatalf("ImportDeck to empty dir: %v", err)
	}

	// ListDecks should include the imported deck
	decks, err := store.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks) != 1 || decks[0] != "test_deck" {
		t.Errorf("expected [test_deck], got %v", decks)
	}
}

// TestExportDeckFromCache_EmptyDeck verifies exporting a deck with no entries.
func TestExportDeckFromCache_EmptyDeck(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := t.TempDir()
	emptyYAML := `id: empty_deck
name: Empty
language: fr
translation_language: en`

	path := filepath.Join(deckDir, "empty_deck.yaml")
	if err := os.WriteFile(path, []byte(emptyYAML), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := store.SyncDecks(deckDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	dstPath := filepath.Join(t.TempDir(), "empty.yaml")
	if err := store.ExportDeckFromCache("empty_deck", dstPath); err != nil {
		t.Fatalf("ExportDeckFromCache empty deck: %v", err)
	}

	parsed, err := parser.ParseFile(dstPath)
	if err != nil {
		t.Fatalf("parse empty export: %v", err)
	}
	if len(parsed.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(parsed.Entries))
	}
}

// TestExportDeckFromCache_RendersYamlCorrectly verifies the YAML output structure.
func TestExportDeckFromCache_RendersYamlCorrectly(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	setupSyncedDeck(t, store)
	dstPath := filepath.Join(t.TempDir(), "out.yaml")

	if err := store.ExportDeckFromCache("test_deck", dstPath); err != nil {
		t.Fatalf("ExportDeckFromCache: %v", err)
	}

	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var parsed model.Deck
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.ID != "test_deck" {
		t.Errorf("id: got %q", parsed.ID)
	}
	if parsed.Name != "Test Deck" {
		t.Errorf("name: got %q", parsed.Name)
	}
	if len(parsed.Entries) != 2 {
		t.Fatalf("entries: got %d", len(parsed.Entries))
	}

	e0 := parsed.Entries[0]
	if e0.Term != "bonjour" {
		t.Errorf("entry 0 term: got %q", e0.Term)
	}
	if len(e0.Translations) != 2 {
		t.Errorf("entry 0 translations: got %d", len(e0.Translations))
	}
	if len(e0.Tags) != 1 || e0.Tags[0] != "greeting" {
		t.Errorf("entry 0 tags: got %v", e0.Tags)
	}

	e1 := parsed.Entries[1]
	if len(e1.Examples) != 1 {
		t.Errorf("entry 1 examples: got %d", len(e1.Examples))
	}
}

// --- DeleteDeck ---

func TestDeleteDeck(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	if err := store.DeleteDeck("test_deck", deckDir); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "test_deck")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("deck should be gone from DB: %v", err)
	}

	yamlPath := filepath.Join(deckDir, "test_deck.yaml")
	if _, err := os.Stat(yamlPath); !os.IsNotExist(err) {
		t.Errorf("YAML file should be removed")
	}

	entries, err := store.queries.ListEntriesByDeck(context.Background(), "test_deck")
	if err != nil {
		t.Fatalf("ListEntriesByDeck: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestDeleteDeck_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	err := store.DeleteDeck("nonexistent", t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestDeleteDeck_WithProgress(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	due := timeOnly("2026-07-26 12:00:00")
	if err := store.queries.UpsertProgress(context.Background(), db.UpsertProgressParams{
		DeckID:    "test_deck",
		EntryID:   "entry_1",
		Reverse:   0,
		Ease:      2.5,
		Interval:  1,
		Due:       &due,
		Correct:   5,
		Incorrect: 1,
	}); err != nil {
		t.Fatalf("UpsertProgress: %v", err)
	}

	if err := store.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_2", 1, true); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if err := store.DeleteDeck("test_deck", deckDir); err != nil {
		t.Fatalf("DeleteDeck: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "test_deck")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("deck should be gone")
	}
}

func TestDeleteDeck_FileMissing(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	// Remove the YAML file before deleting (simulates partial state)
	yamlPath := filepath.Join(deckDir, "test_deck.yaml")
	if err := os.Remove(yamlPath); err != nil {
		t.Fatalf("remove yaml: %v", err)
	}

	if err := store.DeleteDeck("test_deck", deckDir); err != nil {
		t.Fatalf("DeleteDeck with missing file: %v", err)
	}

	_, err := store.queries.GetDeck(context.Background(), "test_deck")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("deck should still be deleted from DB: %v", err)
	}
}

// --- Entry operations ---

func TestAddEntry(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	newEntry := model.Entry{
		ID:   "entry_3",
		Term: "merci",
		Translations: []model.Translation{
			{Text: "thank you"},
		},
	}

	if err := store.AddEntry("test_deck", newEntry, deckDir); err != nil {
		t.Fatalf("AddEntry: %v", err)
	}

	entries, err := store.queries.ListEntriesByDeck(context.Background(), "test_deck")
	if err != nil {
		t.Fatalf("ListEntriesByDeck: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries in DB, got %d", len(entries))
	}

	found := false
	for _, e := range entries {
		if e.ID == "entry_3" && e.Term == "merci" {
			found = true
			break
		}
	}
	if !found {
		t.Error("entry_3 not found in DB after AddEntry")
	}

	// Verify YAML has the new entry
	parsed, err := parser.ParseFile(filepath.Join(deckDir, "test_deck.yaml"))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	if len(parsed.Entries) != 3 {
		t.Fatalf("expected 3 entries in YAML, got %d", len(parsed.Entries))
	}
}

func TestAddEntry_DuplicateID(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	dup := model.Entry{
		ID:   "entry_1",
		Term: "duplicate",
		Translations: []model.Translation{
			{Text: "dup"},
		},
	}

	err := store.AddEntry("test_deck", dup, deckDir)
	if err == nil {
		t.Fatal("expected error for duplicate entry ID")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestUpdateEntry(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	updated := model.Entry{
		Term: "bonjour (updated)",
		Translations: []model.Translation{
			{Text: "hello"},
			{Text: "good day"},
		},
		Tags: []string{"greeting", "updated"},
	}

	if err := store.UpdateEntry("test_deck", "entry_1", updated, deckDir); err != nil {
		t.Fatalf("UpdateEntry: %v", err)
	}

	parsed, err := parser.ParseFile(filepath.Join(deckDir, "test_deck.yaml"))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}

	var found *model.Entry
	for i := range parsed.Entries {
		if parsed.Entries[i].ID == "entry_1" {
			found = &parsed.Entries[i]
			break
		}
	}
	if found == nil {
		t.Fatal("entry_1 not found after update")
	}
	if found.Term != "bonjour (updated)" {
		t.Errorf("expected term 'bonjour (updated)', got %q", found.Term)
	}
	if len(found.Translations) != 2 || found.Translations[1].Text != "good day" {
		t.Errorf("translations not updated: %v", found.Translations)
	}
	if len(found.Tags) != 2 || found.Tags[1] != "updated" {
		t.Errorf("tags not updated: %v", found.Tags)
	}
}

func TestUpdateEntry_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	err := store.UpdateEntry("test_deck", "nonexistent", model.Entry{
		Term: "nope",
		Translations: []model.Translation{{Text: "no"}},
	}, deckDir)
	if err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}

func TestReplaceEntryID(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	// Create some progress for entry_1
	due := timeOnly("2026-07-26 12:00:00")
	if err := store.queries.UpsertProgress(context.Background(), db.UpsertProgressParams{
		DeckID:    "test_deck",
		EntryID:   "entry_1",
		Reverse:   0,
		Ease:      2.5,
		Interval:  1,
		Due:       &due,
		Correct:   3,
		Incorrect: 0,
	}); err != nil {
		t.Fatalf("UpsertProgress: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if err := store.ReplaceEntryID("test_deck", "entry_1", "entry_1_new", deckDir); err != nil {
		t.Fatalf("ReplaceEntryID: %v", err)
	}

	// Verify old ID no longer exists
	_, err := store.queries.GetEntry(context.Background(), "entry_1")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("old entry should be gone: %v", err)
	}

	// Verify new ID exists
	got, err := store.queries.GetEntry(context.Background(), "entry_1_new")
	if err != nil {
		t.Fatalf("GetEntry new: %v", err)
	}
	if got.ID != "entry_1_new" {
		t.Errorf("expected entry_1_new, got %q", got.ID)
	}

	// Verify progress migrated to new ID
	p, err := store.queries.GetProgress(context.Background(), db.GetProgressParams{
		DeckID:  "test_deck",
		EntryID: "entry_1_new",
		Reverse: 0,
	})
	if err != nil {
		t.Fatalf("GetProgress after replace: %v", err)
	}
	if p.Correct != 3 {
		t.Errorf("expected progress.correct=3, got %d", p.Correct)
	}

	// Verify reviews migrated to new ID
	reviews, err := store.GetReviewsByEntry("entry_1_new", 10)
	if err != nil {
		t.Fatalf("GetReviewsByEntry: %v", err)
	}
	if len(reviews) != 1 {
		t.Errorf("expected 1 review for new ID, got %d", len(reviews))
	}
}

func TestReplaceEntryID_Duplicate(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	err := store.ReplaceEntryID("test_deck", "entry_1", "entry_2", deckDir)
	if err == nil {
		t.Fatal("expected error for duplicate target ID")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestRemoveEntry(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	// Create progress for entry_2
	due := timeOnly("2026-07-26 12:00:00")
	if err := store.queries.UpsertProgress(context.Background(), db.UpsertProgressParams{
		DeckID:    "test_deck",
		EntryID:   "entry_2",
		Reverse:   0,
		Ease:      2.5,
		Interval:  1,
		Due:       &due,
		Correct:   1,
		Incorrect: 0,
	}); err != nil {
		t.Fatalf("UpsertProgress: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_2", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if err := store.RemoveEntry("test_deck", "entry_2", deckDir); err != nil {
		t.Fatalf("RemoveEntry: %v", err)
	}

	// Verify removed from YAML
	parsed, err := parser.ParseFile(filepath.Join(deckDir, "test_deck.yaml"))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	if len(parsed.Entries) != 1 {
		t.Fatalf("expected 1 entry after remove, got %d", len(parsed.Entries))
	}
	if parsed.Entries[0].ID == "entry_2" {
		t.Error("entry_2 should not be in YAML after remove")
	}

	// Verify removed from DB
	entries, err := store.queries.ListEntriesByDeck(context.Background(), "test_deck")
	if err != nil {
		t.Fatalf("ListEntriesByDeck: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in DB, got %d", len(entries))
	}

	// Verify progress cleaned up
	_, err = store.queries.GetProgress(context.Background(), db.GetProgressParams{
		DeckID:  "test_deck",
		EntryID: "entry_2",
		Reverse: 0,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("progress should be deleted: %v", err)
	}

	// Verify reviews cleaned up
	reviews, err := store.GetReviewsByEntry("entry_2", 10)
	if err != nil {
		t.Fatalf("GetReviewsByEntry: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("expected 0 reviews after remove, got %d", len(reviews))
	}
}

func TestRemoveEntry_NotFound(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	deckDir := setupSyncedDeck(t, store)

	err := store.RemoveEntry("test_deck", "nonexistent", deckDir)
	if err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}
