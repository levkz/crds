package app

import (
	"time"

	"crds/internal/stats"
	"crds/internal/storage"
	"crds/internal/ui"
)

// DeckProvider is the interface the UI uses to list and load decks.
// The concrete implementation lives outside the UI package.
type DeckProvider interface {
	ListDecks() ([]string, error)
	LoadDeck(name string) (ui.DeckData, error)
}

// TagProvider is the interface for tag-related queries used by the deck selection screen.
type TagProvider interface {
	ListAllTags() ([]string, error)
	ListDeckTags(deckID string) ([]string, error)
	FilterDecksByTags(tags []string) ([]string, error)
	FilterTagsByDecks(deckIDs []string) ([]string, error)
}

// ProgressRecorder persists learning progress from quiz sessions.
type ProgressRecorder interface {
	RecordAnswer(deckID, cardID string, grade int, reverse bool) error
}

// SessionManager creates and manages quiz sessions.
type SessionManager interface {
	EnsureSession() (int64, error)
	ResetSession() error
}

// TypingRecorder persists typing-specific answer details.
type TypingRecorder interface {
	RecordAnswerFull(sessionID int64, deckID, entryID string, grade int, reverse bool, userInput, correctAnswer string, similarity float64) (int64, error)
}

// DueProvider reports which cards belong in the review queue for a
// deck/tag selection (unseen cards first, then due cards).
type DueProvider interface {
	DueForSelection(deckIDs, tags []string, now time.Time) ([]string, error)
}

// Dependencies bundles all external services the UI needs.
type Dependencies struct {
	Decks    DeckProvider
	Progress ProgressRecorder
	Stats    stats.Provider
	State    *storage.StateStore
	Tags     TagProvider
	Due      DueProvider
}
