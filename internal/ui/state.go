package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/stats"
)

// AppState is the application-wide data snapshot owned by the root model.
// Screens that need global data implement StateSyncer and receive a copy of
// this snapshot on entry (when they become visible) and whenever it changes.
type AppState struct {
	Deck          *DeckData
	DeckProgress  map[string]stats.EntryProgress
	AllDecks      []string
	SelectedDecks []string
	AllTags       []string
	SelectedTags  []string
	AllDeckTags   map[string][]string
	QuizMode      QuizMode
	Stats         *stats.Summary
}

// StateSyncer is implemented by screens that read from AppState.
// SyncState must be idempotent: it recomputes derived state only when the
// incoming snapshot differs from what the screen already holds.
type StateSyncer interface {
	SyncState(AppState) tea.Cmd
}

// StateChangedMsg is emitted by the root whenever AppState mutates. The root
// forwards it to the active screen via StateSyncer.SyncState.
type StateChangedMsg struct {
	State AppState
}

// SetQuizModeMsg requests a global quiz mode change (persisted in AppState so
// both quiz screens stay in sync).
type SetQuizModeMsg struct {
	Mode QuizMode
}

// RefreshStatsMsg is emitted by the Statistics screen on entry to request a
// fresh stats summary from the root.
type RefreshStatsMsg struct{}
