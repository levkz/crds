package storage

import (
	"os"
	"path/filepath"
	"testing"

	"crds/internal/model"
)

func TestDeckStore_ListDecks(t *testing.T) {
	dir := t.TempDir()

	s := NewDeckStore(dir)

	names, err := s.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks on empty dir: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 decks, got %d", len(names))
	}

	if err := os.WriteFile(filepath.Join(dir, "test.yaml"), []byte("id: test\nname: Test\nlanguage: fr\nentries: []"), 0644); err != nil {
		t.Fatalf("write test deck: %v", err)
	}

	names, err = s.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(names) != 1 || names[0] != "test" {
		t.Fatalf("expected [test], got %v", names)
	}

	// Non-yaml files should be ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	names, err = s.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(names))
	}
}

func TestDeckStore_ListDecks_NonExistentDir(t *testing.T) {
	s := NewDeckStore("/tmp/crds-test-nonexistent-" + t.Name())

	names, err := s.ListDecks()
	if err != nil {
		t.Fatalf("ListDecks on nonexistent dir: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 decks, got %d", len(names))
	}
}

func TestDeckStore_LoadDeck_NotFound(t *testing.T) {
	s := NewDeckStore(t.TempDir())

	_, err := s.LoadDeck("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestDeckStore_LoadDeck_Valid(t *testing.T) {
	dir := t.TempDir()
	yamlContent := []byte(`id: french_a1
name: French A1
language: fr
translation_language: en

entries:
  - id: fr_bonjour
    term: bonjour
    translations:
      - text: hello
    examples:
      - text: Bonjour, Marie.
        translation: Hello, Marie.
    tags: [greeting]
    notes: Common greeting.
`)

	if err := os.WriteFile(filepath.Join(dir, "french_a1.yaml"), yamlContent, 0644); err != nil {
		t.Fatalf("write deck: %v", err)
	}

	s := NewDeckStore(dir)
	deck, err := s.LoadDeck("french_a1")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}

	if deck.Name != "French A1" {
		t.Errorf("expected name 'French A1', got %q", deck.Name)
	}

	if len(deck.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(deck.Cards))
	}

	card := deck.Cards[0]
	if card.ID != "fr_bonjour" {
		t.Errorf("expected id 'fr_bonjour', got %q", card.ID)
	}
	if card.Front != "bonjour" {
		t.Errorf("expected front 'bonjour', got %q", card.Front)
	}
	if len(card.Back) != 1 || card.Back[0] != "hello" {
		t.Errorf("expected back ['hello'], got %v", card.Back)
	}
	if card.Notes != "Common greeting." {
		t.Errorf("expected notes 'Common greeting.', got %q", card.Notes)
	}
}

func TestDeckStore_LoadDeck_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("invalid: [yaml: \n"), 0644); err != nil {
		t.Fatalf("write bad deck: %v", err)
	}

	s := NewDeckStore(dir)
	_, err := s.LoadDeck("bad")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestEntryToCardData(t *testing.T) {
	entry := model.Entry{
		ID:    "test_id",
		Term:  "test term",
		Notes: "test notes",
		Translations: []model.Translation{
			{Text: "trans1"},
			{Text: "trans2"},
		},
	}

	card := entryToCardData(entry)

	if card.ID != "test_id" {
		t.Errorf("expected id 'test_id', got %q", card.ID)
	}
	if card.Front != "test term" {
		t.Errorf("expected front 'test term', got %q", card.Front)
	}
	if len(card.Back) != 2 || card.Back[0] != "trans1" || card.Back[1] != "trans2" {
		t.Errorf("expected back ['trans1', 'trans2'], got %v", card.Back)
	}
	if card.Notes != "test notes" {
		t.Errorf("expected notes 'test notes', got %q", card.Notes)
	}
}
