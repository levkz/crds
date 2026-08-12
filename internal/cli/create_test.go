package cli

import (
	"os"
	"path/filepath"
	"testing"

	"crds/internal/parser"
)

func TestCreateCmd_Run(t *testing.T) {
	a := newTestApp(t)

	c := &CreateCmd{Deck: "french_a1", From: "fr", To: "en"}
	if err := c.Run(a); err != nil {
		t.Fatalf("CreateCmd.Run: %v", err)
	}

	path := filepath.Join(a.DataDir, "french_a1.yaml")
	deck, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("parse created deck: %v", err)
	}
	if deck.ID != "french_a1" || deck.Name != "french_a1" {
		t.Fatalf("expected id/name french_a1, got id=%q name=%q", deck.ID, deck.Name)
	}
	if deck.Language != "fr" || deck.TranslationLanguage != "en" {
		t.Fatalf("expected fr→en, got %q→%q", deck.Language, deck.TranslationLanguage)
	}
	if len(deck.Entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(deck.Entries))
	}

	if _, err := a.Store.LoadDeck("french_a1"); err != nil {
		t.Fatalf("deck not synced to cache: %v", err)
	}
}

func TestCreateCmd_Run_AlreadyExists(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "french_a1")
	syncDecks(t, a)

	path := filepath.Join(a.DataDir, "french_a1.yaml")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	c := &CreateCmd{Deck: "french_a1", From: "de", To: "fr"}
	if err := c.Run(a); err == nil {
		t.Fatal("expected error when deck already exists")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("existing deck file was modified")
	}
}