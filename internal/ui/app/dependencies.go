package app

// DeckProvider is the interface the UI uses to list and load decks.
// The concrete implementation lives outside the UI package.
type DeckProvider interface {
	ListDecks() ([]string, error)
	LoadDeck(name string) (DeckData, error)
}

// DeckData is a lightweight transfer type passed into the UI.
type DeckData struct {
	Name  string
	Cards []CardData
}

// CardData represents a single vocabulary card for display.
type CardData struct {
	ID    string
	Front string
	Back  []string
	Notes string
}

// ProgressRecorder persists learning progress from quiz sessions.
type ProgressRecorder interface {
	RecordAnswer(cardID string, grade int) error
}

// Dependencies bundles all external services the UI needs.
type Dependencies struct {
	Decks    DeckProvider
	Progress ProgressRecorder
}
