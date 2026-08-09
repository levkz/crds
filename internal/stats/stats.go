package stats

import "time"

// Confidence computes a 0.0-1.0 score from correct/incorrect counts.
// Returns 0.5 for unseen cards (0 correct, 0 incorrect).
func Confidence(correct, incorrect int) float64 {
	total := correct + incorrect
	if total == 0 {
		return 0.5
	}
	return float64(correct) / float64(total)
}

// DayPoint aggregates review outcomes for a single day.
type DayPoint struct {
	Day       string // date in YYYY-MM-DD form
	Correct   int
	Incorrect int
}

// Confidence returns the confidence score for this day.
func (p DayPoint) Confidence() float64 {
	return Confidence(p.Correct, p.Incorrect)
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
	Streak        int
}

// DeckStats holds per-deck statistics.
type DeckStats struct {
	TotalEntries  int
	ReviewedToday int
	Accuracy      float64
	AvgConfidence float64
}

// WordStats holds per-entry learning statistics.
type WordStats struct {
	TotalReviews  int
	ReviewedToday int
	Correct       int
	Incorrect     int
	LastReviewed  *time.Time
}

// Accuracy returns the review accuracy percentage (0-100) or 0 when unseen.
func (w WordStats) Accuracy() float64 {
	if w.TotalReviews == 0 {
		return 0
	}
	return float64(w.Correct) / float64(w.TotalReviews) * 100
}

// Confidence returns the confidence score for this word.
func (w WordStats) Confidence() float64 {
	return Confidence(w.Correct, w.Incorrect)
}

// Mastered reports whether the word has reached the mastery threshold.
func (w WordStats) Mastered() bool {
	return w.Confidence() >= 0.8
}

// Streak returns the number of consecutive days (ending today or yesterday)
// with at least one review. Day ordering and duplicates do not matter; each
// time.Time is truncated to its calendar date.
func Streak(days []time.Time) int {
	if len(days) == 0 {
		return 0
	}

	set := make(map[time.Time]bool, len(days))
	for _, d := range days {
		trunc := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		set[trunc] = true
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// If today has no review, streak may still be alive through yesterday.
	cur := today
	if !set[today] {
		yesterday := today.AddDate(0, 0, -1)
		if !set[yesterday] {
			return 0
		}
		cur = yesterday
	}

	streak := 0
	for set[cur] {
		streak++
		cur = cur.AddDate(0, 0, -1)
	}
	return streak
}
