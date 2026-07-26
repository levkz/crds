package parser

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		wantError bool
	}{
		{
			name:      "valid deck",
			file:      "valid.yaml",
			wantError: false,
		},
		{
			name:      "minimal deck",
			file:      "minimal.yaml",
			wantError: false,
		},
		{
			name:      "duplicate ids",
			file:      "duplicate_ids.yaml",
			wantError: true,
		},
		{
			name:      "duplicate terms",
			file:      "duplicate_terms.yaml",
			wantError: true,
		},
		{
			name:      "missing term",
			file:      "missing_term.yaml",
			wantError: true,
		},
		{
			name:      "missing translation",
			file:      "missing_translation.yaml",
			wantError: true,
		},
		{
			name:      "malformed yaml",
			file:      "malformed.yaml",
			wantError: true,
		},
		{
			name:      "unicode",
			file:      "unicode.yaml",
			wantError: false,
		},
		{
			name:      "comments",
			file:      "comments.yaml",
			wantError: false,
		},
		{
			name:      "auto ids",
			file:      "auto_ids.yaml",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)

			deck, err := ParseFile(path)

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if deck == nil {
				t.Fatal("expected deck, got nil")
			}
		})
	}
}

func TestValidDeckContents(t *testing.T) {
	deck, err := ParseFile(filepath.Join("testdata", "valid.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deck.ID != "french_a1" {
		t.Errorf("deck.ID = %q, want %q", deck.ID, "french_a1")
	}

	if deck.Name != "French A1" {
		t.Errorf("deck.Name = %q", deck.Name)
	}

	if deck.Language != "fr" {
		t.Errorf("deck.Language = %q", deck.Language)
	}

	if len(deck.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(deck.Entries))
	}

	first := deck.Entries[0]

	if first.ID != "fr_bonjour" {
		t.Errorf("entry.ID = %q", first.ID)
	}

	if first.Term != "bonjour" {
		t.Errorf("entry.Term = %q", first.Term)
	}

	if len(first.Translations) != 2 {
		t.Fatalf("expected 2 translations")
	}

	if first.Translations[0].Text != "hello" {
		t.Errorf("unexpected translation")
	}

	if len(first.Examples) != 1 {
		t.Fatalf("expected 1 example")
	}

	if first.Examples[0].Translation != "Hello, Marie." {
		t.Errorf("unexpected example translation")
	}

	if len(first.Tags) != 2 {
		t.Errorf("expected 2 tags")
	}
}

func TestAutoIDs(t *testing.T) {
	deck, err := ParseFile(filepath.Join("testdata", "auto_ids.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deck.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(deck.Entries))
	}

	tests := []struct {
		index   int
		wantID  string
		wantErr bool
	}{
		{index: 0, wantID: "bonjour"},
		{index: 1, wantID: "mange"},
		{index: 2, wantID: "s_il_vous_plaît"},
	}

	for _, tt := range tests {
		got := deck.Entries[tt.index].ID
		if got != tt.wantID {
			t.Errorf("entry %d: id = %q, want %q", tt.index, got, tt.wantID)
		}
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	deck, err := ParseFile(filepath.Join("testdata", "whitespace.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deck fields
	if got := deck.ID; got != "french" {
		t.Errorf("deck.ID = %q, want %q", got, "french")
	}

	if got := deck.Name; got != "French" {
		t.Errorf("deck.Name = %q, want %q", got, "French")
	}

	if got := deck.Language; got != "fr" {
		t.Errorf("deck.Language = %q, want %q", got, "fr")
	}

	if got := deck.TranslationLanguage; got != "en" {
		t.Errorf("deck.TranslationLanguage = %q, want %q", got, "en")
	}

	if len(deck.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(deck.Entries))
	}

	entry := deck.Entries[0]

	// Entry fields
	if got := entry.ID; got != "fr_bonjour" {
		t.Errorf("entry.ID = %q, want %q", got, "fr_bonjour")
	}

	if got := entry.Term; got != "bonjour" {
		t.Errorf("entry.Term = %q, want %q", got, "bonjour")
	}

	if got := entry.Notes; got != "Common greeting." {
		t.Errorf("entry.Notes = %q, want %q", got, "Common greeting.")
	}

	// Translations
	if len(entry.Translations) != 1 {
		t.Fatalf("expected 1 translation, got %d", len(entry.Translations))
	}

	if got := entry.Translations[0].Text; got != "hello" {
		t.Errorf("translation.Text = %q, want %q", got, "hello")
	}

	// Examples
	if len(entry.Examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(entry.Examples))
	}

	example := entry.Examples[0]

	if got := example.Text; got != "Bonjour, Marie." {
		t.Errorf("example.Text = %q, want %q", got, "Bonjour, Marie.")
	}

	if got := example.Translation; got != "Hello, Marie." {
		t.Errorf("example.Translation = %q, want %q", got, "Hello, Marie.")
	}

	// Tags
	if len(entry.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(entry.Tags))
	}

	if got := entry.Tags[0]; got != "greeting" {
		t.Errorf("tag[0] = %q, want %q", got, "greeting")
	}

	if got := entry.Tags[1]; got != "A1" {
		t.Errorf("tag[1] = %q, want %q", got, "A1")
	}
}
