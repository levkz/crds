package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"crds/internal/model"
	"crds/internal/scheduler"
	"crds/internal/storage/db"
)

// progressByDeck resolves a deck/tag scope (see selectionDecks) and returns
// every deck that must be searched, or an error.
func (s *Store) resolveDueDecks(deckIDs, tags []string) ([]string, error) {
	decks, err := s.selectionDecks(deckIDs, tags)
	if err != nil {
		return nil, err
	}
	if len(decks) == 0 {
		names, err := s.queries.ListDeckNames(context.Background())
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			decks = append(decks, n.ID)
		}
	}
	return decks, nil
}

// DueForSelection returns the entry IDs that belong in the review queue for a
// deck/tag selection: unseen cards first (in deck order), then due cards
// ordered by due date. Duplicate entries across decks appear once.
func (s *Store) DueForSelection(deckIDs, tags []string, now time.Time) ([]string, error) {
	now = now.UTC()
	decks, err := s.resolveDueDecks(deckIDs, tags)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var newEntries []string
	var due []db.Progress
	for _, d := range decks {
		ids, err := s.queries.ListNewEntriesByDeck(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("list new entries for %q: %w", d, err)
		}
		newEntries = append(newEntries, ids...)

		rows, err := s.queries.GetDueCards(ctx, db.GetDueCardsParams{DeckID: d, Due: &now})
		if err != nil {
			return nil, fmt.Errorf("get due cards for %q: %w", d, err)
		}
		due = append(due, rows...)
	}

	sort.Slice(due, func(i, j int) bool {
		return dueTime(&due[i]).Before(dueTime(&due[j]))
	})

	out := make([]string, 0, len(newEntries)+len(due))
	seen := make(map[string]bool, len(newEntries)+len(due))
	for _, id := range newEntries {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	for _, r := range due {
		if !seen[r.EntryID] {
			seen[r.EntryID] = true
			out = append(out, r.EntryID)
		}
	}
	return out, nil
}

// dueTodayCount counts distinct entries with a due date at or before now,
// restricted to the given decks (nil/empty = all decks).
func (s *Store) dueTodayCount(decks []string, now time.Time) (int, error) {
	now = now.UTC()
	ctx := context.Background()
	if len(decks) == 0 {
		names, err := s.queries.ListDeckNames(ctx)
		if err != nil {
			return 0, err
		}
		for _, n := range names {
			decks = append(decks, n.ID)
		}
	}

	seen := make(map[string]bool)
	for _, d := range decks {
		rows, err := s.queries.GetDueCards(ctx, db.GetDueCardsParams{DeckID: d, Due: &now})
		if err != nil {
			continue
		}
		for _, r := range rows {
			seen[r.EntryID] = true
		}
	}
	return len(seen), nil
}

// persistAnswer writes a review and the updated scheduling state for a single
// answer in one transaction.
func persistAnswer(ctx context.Context, q *db.Queries, deckID, entryID string, grade int, reverse bool, now time.Time) error {
	prev, err := loadProgress(ctx, q, deckID, entryID, reverse)
	if err != nil {
		return err
	}

	next := scheduler.Update(prev, grade, now)
	due := next.Due

	return q.UpsertProgress(ctx, db.UpsertProgressParams{
		DeckID:    deckID,
		EntryID:   entryID,
		Reverse:   boolInt(reverse),
		Ease:      next.Ease,
		Interval:  int64(next.Interval),
		Due:       &due,
		Correct:   int64(next.Correct),
		Incorrect: int64(next.Incorrect),
	})
}

// loadProgress fetches the scheduling row for a card, returning a zero
// progress record (marked new) when none exists yet.
func loadProgress(ctx context.Context, q *db.Queries, deckID, entryID string, reverse bool) (model.Progress, error) {
	row, err := q.GetProgress(ctx, db.GetProgressParams{
		DeckID:  deckID,
		EntryID: entryID,
		Reverse: boolInt(reverse),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Progress{DeckID: deckID, EntryID: entryID, Reverse: reverse}, nil
		}
		return model.Progress{}, err
	}
	return dbProgressToModel(row), nil
}

func dbProgressToModel(p db.Progress) model.Progress {
	m := model.Progress{
		DeckID:    p.DeckID,
		EntryID:   p.EntryID,
		Reverse:   p.Reverse != 0,
		Ease:      p.Ease,
		Interval:  int(p.Interval),
		Correct:   int(p.Correct),
		Incorrect: int(p.Incorrect),
	}
	if p.Due != nil {
		m.Due = *p.Due
	}
	return m
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func dueTime(p *db.Progress) time.Time {
	if p.Due == nil {
		return time.Time{}
	}
	return *p.Due
}