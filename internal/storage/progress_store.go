package storage

import (
	"sync"
	"time"

	"crds/internal/model"
	"crds/internal/ui"
)

type ProgressStore struct {
	mu      sync.Mutex
	reviews []model.Review
}

func NewProgressStore() *ProgressStore {
	return &ProgressStore{}
}

func (s *ProgressStore) RecordAnswer(deckID, cardID string, grade int, reverse bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reviews = append(s.reviews, model.Review{
		DeckID:     deckID,
		EntryID:    cardID,
		ReviewedAt: time.Now(),
		Grade:      grade,
		Reverse:    reverse,
	})
	return nil
}

func (s *ProgressStore) Stats() ui.Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour)

	var reviewedToday, correctToday int
	seen := make(map[string]bool)

	for _, r := range s.reviews {
		if r.ReviewedAt.After(todayStart) {
			reviewedToday++
			if r.Grade >= int(ui.GradeGood) {
				correctToday++
			}
		}
		seen[r.EntryID] = true
	}

	var accuracy float64
	if reviewedToday > 0 {
		accuracy = float64(correctToday) / float64(reviewedToday) * 100
	}

	return ui.Stats{
		ReviewedToday: reviewedToday,
		Accuracy:      accuracy,
		TotalCards:    len(seen),
	}
}
