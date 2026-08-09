package stats

// Confidence computes a 0.0-1.0 score from correct/incorrect counts.
// Returns 0.5 for unseen cards (0 correct, 0 incorrect).
func Confidence(correct, incorrect int) float64 {
	total := correct + incorrect
	if total == 0 {
		return 0.5
	}
	return float64(correct) / float64(total)
}

// EntryProgress holds per-entry learning data from the progress table.
// This aggregates across both forward/reverse directions.
type EntryProgress struct {
	Correct   int
	Incorrect int
}

// Confidence returns the confidence score for this entry.
func (p EntryProgress) Confidence() float64 {
	return Confidence(p.Correct, p.Incorrect)
}

// Summary aggregates stats across all decks.
type Summary struct {
	ReviewedToday int
	Accuracy      float64
	TotalCards    int
	Mastered      int
	DueToday      int
}

// DeckStats holds per-deck statistics.
type DeckStats struct {
	TotalEntries  int
	ReviewedToday int
	Accuracy      float64
	AvgConfidence float64
}
