package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"

	"crds/internal/app"
	"crds/internal/storage"
	"crds/internal/ui"
)

func init() {
	goose.SetLogger(goose.NopLogger())
}

func newTestApp(t *testing.T) *app.App {
	t.Helper()

	sharedDir := t.TempDir()
	dataDir := filepath.Join(sharedDir, "decks")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	store, err := storage.NewStore(filepath.Join(sharedDir, "crds.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	stateStore := storage.NewStateStore(sharedDir)

	return &app.App{
		Store:     store,
		State:     stateStore,
		SharedDir: sharedDir,
		DataDir:   dataDir,
	}
}

func writeTestDeck(t *testing.T, dataDir, id string) {
	t.Helper()
	content := `id: ` + id + `
name: Test ` + id + `
language: en
translation_language: fr
entries:
  - id: e1
    term: hello
    translations:
      - text: bonjour
  - id: e2
    term: goodbye
    translations:
      - text: au revoir
    tags:
      - common
`
	path := filepath.Join(dataDir, id+".yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func syncDecks(t *testing.T, a *app.App) {
	t.Helper()
	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		t.Fatal(err)
	}
}

func TestQuizCmd_PreSelectDeck_Valid(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	q := &QuizCmd{Deck: "spanish"}
	if err := q.preSelectDeck(a); err != nil {
		t.Fatalf("preSelectDeck: %v", err)
	}

	state, err := a.State.Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(state.SelectedDecks) != 1 || state.SelectedDecks[0] != "spanish" {
		t.Fatalf("expected [spanish], got %v", state.SelectedDecks)
	}
}

func TestQuizCmd_PreSelectDeck_NotFound(t *testing.T) {
	a := newTestApp(t)
	q := &QuizCmd{Deck: "nonexistent"}
	if err := q.preSelectDeck(a); err == nil {
		t.Fatal("expected error for nonexistent deck")
	}
}

func TestSyncCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "french")

	s := &SyncCmd{}
	if err := s.Run(a); err != nil {
		t.Fatalf("SyncCmd.Run: %v", err)
	}
}

func TestStatsCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "german")
	syncDecks(t, a)

	s := &StatsCmd{}
	if err := s.Run(a); err != nil {
		t.Fatalf("StatsCmd.Run: %v", err)
	}
}

func TestSearchCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "italian")
	syncDecks(t, a)

	s := &SearchCmd{Deck: "italian", Query: "hello"}
	if err := s.Run(a); err != nil {
		t.Fatalf("SearchCmd.Run: %v", err)
	}
}

func TestSearchCmd_Run_NoMatch(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "latin")
	syncDecks(t, a)

	s := &SearchCmd{Deck: "latin", Query: "zzzznonexistent"}
	if err := s.Run(a); err != nil {
		t.Fatalf("SearchCmd.Run: %v", err)
	}
}

func TestSearchCmd_Matches(t *testing.T) {
	card := ui.CardData{
		Front:    "hello",
		Back:     []string{"bonjour", "salut"},
		Variants: []string{"hi", "hey"},
		Notes:    "common greeting",
	}

	tests := []struct {
		query string
		want  bool
	}{
		{"hello", true},
		{"HELLO", true},
		{"bonjour", true},
		{"salut", true},
		{"greeting", true},
		{"hi", true},
		{"zzzz", false},
	}

	for _, tt := range tests {
		got := matches(card, strings.ToLower(tt.query))
		if got != tt.want {
			t.Errorf("matches(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestExportCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "dutch")
	syncDecks(t, a)

	dst := filepath.Join(t.TempDir(), "exported.yaml")
	e := &ExportCmd{Deck: "dutch", Output: dst}
	if err := e.Run(a); err != nil {
		t.Fatalf("ExportCmd.Run: %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Fatal("exported file does not exist")
	}
}

func TestReserveCmd_Run(t *testing.T) {
	a := newTestApp(t)

	r := &ReserveCmd{}
	if err := r.Run(a); err != nil {
		t.Fatalf("ReserveCmd.Run: %v", err)
	}
}

func TestTermRmCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "swedish")
	syncDecks(t, a)

	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		t.Fatal(err)
	}

	r := &TermRmCmd{Deck: "swedish", TermID: "e1"}
	if err := r.Run(a); err != nil {
		t.Fatalf("TermRmCmd.Run: %v", err)
	}

	deck, err := a.Store.LoadDeck("swedish")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	for _, card := range deck.Cards {
		if card.ID == "e1" {
			t.Fatal("entry e1 was not removed")
		}
	}
}
