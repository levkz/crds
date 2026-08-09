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

	// SelectionSummary returns aggregate stats for the given deck/tag
	// selection. An empty selection means all decks.
	SelectionSummary(deckIDs, tags []string) (Summary, error)

	// SelectionHistory returns daily review aggregates for the given
	// deck/tag selection. An empty selection means all decks.
	SelectionHistory(deckIDs, tags []string) ([]DayPoint, error)

	// WordStats returns per-entry statistics for a single entry.
	WordStats(entryID string) (WordStats, error)

	// WordHistory returns daily review aggregates for a single entry.
	WordHistory(entryID string) ([]DayPoint, error)
}
