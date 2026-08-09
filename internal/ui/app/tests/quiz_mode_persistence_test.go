package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/storage"
	"crds/internal/ui"
	"crds/internal/ui/app"
)

// drainBatch runs a tea.Cmd and flattens nested BatchMsg results into a
// flat slice of leaf messages.
func drainBatch(cmd tea.Cmd) []tea.Msg {
	var out []tea.Msg
	var walk func(tea.Msg)
	walk = func(msg tea.Msg) {
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					walk(c())
				}
			}
			return
		}
		if msg != nil {
			out = append(out, msg)
		}
	}
	walk(cmd())
	return out
}

func TestQuizModePersistedOnChange(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStateStore(dir)
	m := app.Model{
		Config: app.DefaultConfig(),
		State:  ui.AppState{SelectedDecks: []string{"deck_a"}},
		Dispatcher: &app.Dispatcher{
			State: store,
		},
	}

	updated, cmd := m.Update(ui.SetQuizModeMsg{Mode: ui.QuizModeDue})
	m2 := updated.(app.Model)
	if m2.State.QuizMode != ui.QuizModeDue {
		t.Fatalf("State.QuizMode = %v, want %v", m2.State.QuizMode, ui.QuizModeDue)
	}
	msgs := drainBatch(cmd)
	if !containsSavedState(msgs) {
		t.Fatalf("expected a SavedMsg, got %v", msgs)
	}

	data, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		t.Fatalf("read state.yaml: %v", err)
	}
	if !strings.Contains(string(data), "quiz_mode: due") {
		t.Errorf("state.yaml missing quiz_mode: due\n%s", data)
	}
}

func containsSavedState(msgs []tea.Msg) bool {
	for _, msg := range msgs {
		if saved, ok := msg.(app.SavedMsg); ok && saved.Kind == app.MsgKindState {
			return true
		}
	}
	return false
}

func TestQuizModeRestoredFromState(t *testing.T) {
	dir := t.TempDir()
	content := "selected_decks:\n  - deck_a\nquiz_mode: due\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}
	store := storage.NewStateStore(dir)
	m := app.Model{
		Config:     app.DefaultConfig(),
		State:      ui.AppState{},
		Dispatcher: &app.Dispatcher{State: store},
	}

	updated, _ := m.Update(app.DataLoadedMsg{Kind: app.MsgKindDeckList, Data: []string{"deck_a", "deck_b"}})
	m2 := updated.(app.Model)

	if m2.State.QuizMode != ui.QuizModeDue {
		t.Errorf("State.QuizMode = %v, want %v", m2.State.QuizMode, ui.QuizModeDue)
	}
	if len(m2.State.SelectedDecks) != 1 || m2.State.SelectedDecks[0] != "deck_a" {
		t.Errorf("State.SelectedDecks = %v, want [deck_a]", m2.State.SelectedDecks)
	}
}

func TestQuizModeSaveWithNilDispatcher(t *testing.T) {
	m := app.Model{
		Config: app.DefaultConfig(),
		State:  ui.AppState{},
	}

	updated, cmd := m.Update(ui.SetQuizModeMsg{Mode: ui.QuizModeRandom})
	m2 := updated.(app.Model)

	if m2.State.QuizMode != ui.QuizModeRandom {
		t.Fatalf("State.QuizMode = %v, want %v", m2.State.QuizMode, ui.QuizModeRandom)
	}
	if cmd == nil {
		t.Fatal("expected a command from SetQuizModeMsg")
	}
	msgs := drainBatch(cmd)
	for _, msg := range msgs {
		if _, isSave := msg.(app.SavedMsg); isSave {
			t.Fatalf("did not expect a save with nil Dispatcher, got %v", msg)
		}
	}
}