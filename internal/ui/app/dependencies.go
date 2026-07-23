package app

import (
	"crds/internal/storage"
	"crds/internal/ui"
)

// DeckProvider is the interface the UI uses to list and load decks.
// The concrete implementation lives outside the UI package.
type DeckProvider interface {
	ListDecks() ([]string, error)
	LoadDeck(name string) (ui.DeckData, error)
}

// ProgressRecorder persists learning progress from quiz sessions.
type ProgressRecorder interface {
	RecordAnswer(cardID string, grade int) error
}

// SessionManager creates and manages quiz sessions.
type SessionManager interface {
	EnsureSession() (int64, error)
	ResetSession() error
}

// TypingRecorder persists typing-specific answer details.
type TypingRecorder interface {
	RecordAnswerFull(sessionID int64, deckID, entryID string, grade int, userInput, correctAnswer string, similarity float64) (int64, error)
}

// StatsProvider provides learning statistics.
type StatsProvider interface {
	Stats() ui.Stats
}

// Dependencies bundles all external services the UI needs.
type Dependencies struct {
	Decks    DeckProvider
	Progress ProgressRecorder
	Stats    StatsProvider
	State    *storage.StateStore
}
