package stats

// Provider is the interface for fetching statistics data.
// The storage layer implements this.
type Provider interface {
	// Summary returns aggregate stats across all decks.
	Summary() (Summary, error)

	// DeckSummary returns per-deck stats.
	DeckSummary(deckID string) (DeckStats, error)

	// EntryProgress returns per-entry progress data for a deck,
	// keyed by entry ID. Used by the card sorter for smart/kinda-smart modes.
	EntryProgress(deckID string) (map[string]EntryProgress, error)
}
