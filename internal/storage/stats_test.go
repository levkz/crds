package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"crds/internal/storage/db"
)

func upsertDeck(t *testing.T, store *Store, id string) {
	t.Helper()
	if err := store.queries.UpsertDeck(context.Background(), db.UpsertDeckParams{
		ID:                  id,
		Name:                id,
		Language:            "fr",
		TranslationLanguage: "en",
	}); err != nil {
		t.Fatalf("UpsertDeck(%q): %v", id, err)
	}
}

func upsertEntry(t *testing.T, store *Store, id, deckID string) {
	t.Helper()
	if err := store.queries.UpsertEntry(context.Background(), db.UpsertEntryParams{
		ID:     id,
		DeckID: deckID,
		Term:   id,
		Notes:  "",
	}); err != nil {
		t.Fatalf("UpsertEntry(%q): %v", id, err)
	}
}

// insertReviewRaw inserts a review with an explicit reviewed_at timestamp.
func insertReviewRaw(t *testing.T, store *Store, deckID, entryID string, grade int, reviewedAt string) {
	t.Helper()
	ctx := context.Background()
	sess, err := store.EnsureSession()
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}
	if _, err := store.conn.ExecContext(ctx,
		"INSERT INTO reviews (session_id, deck_id, entry_id, grade, reviewed_at) VALUES (?, ?, ?, ?, ?)",
		sess, deckID, entryID, grade, reviewedAt); err != nil {
		t.Fatalf("insert review: %v", err)
	}
}

func todayAt(hhmmss string) string {
	now := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("%s %s", now, hhmmss)
}

func TestStoreSelectionSummary(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	upsertDeck(t, store, "d1")
	upsertDeck(t, store, "d2")
	upsertEntry(t, store, "e1", "d1")
	upsertEntry(t, store, "e2", "d2")

	// e1 correct today, e2 incorrect today.
	insertReviewRaw(t, store, "d1", "e1", 3, todayAt("10:00:00"))
	insertReviewRaw(t, store, "d2", "e2", 1, todayAt("11:00:00"))

	t.Run("single deck", func(t *testing.T) {
		s, err := store.SelectionSummary([]string{"d1"}, nil)
		if err != nil {
			t.Fatalf("SelectionSummary: %v", err)
		}
		if s.ReviewedToday != 1 {
			t.Errorf("ReviewedToday = %d, want 1", s.ReviewedToday)
		}
		if s.Accuracy != 100 {
			t.Errorf("Accuracy = %f, want 100", s.Accuracy)
		}
		if s.TotalCards != 1 {
			t.Errorf("TotalCards = %d, want 1", s.TotalCards)
		}
		if s.Streak != 1 {
			t.Errorf("Streak = %d, want 1", s.Streak)
		}
	})

	t.Run("both decks", func(t *testing.T) {
		s, err := store.SelectionSummary([]string{"d1", "d2"}, nil)
		if err != nil {
			t.Fatalf("SelectionSummary: %v", err)
		}
		if s.ReviewedToday != 2 {
			t.Errorf("ReviewedToday = %d, want 2", s.ReviewedToday)
		}
		if s.Accuracy != 50 {
			t.Errorf("Accuracy = %f, want 50", s.Accuracy)
		}
		if s.TotalCards != 2 {
			t.Errorf("TotalCards = %d, want 2", s.TotalCards)
		}
	})

	t.Run("empty selection is all decks", func(t *testing.T) {
		s, err := store.SelectionSummary(nil, nil)
		if err != nil {
			t.Fatalf("SelectionSummary: %v", err)
		}
		if s.ReviewedToday != 2 || s.TotalCards != 2 {
			t.Errorf("got ReviewedToday=%d TotalCards=%d, want 2/2", s.ReviewedToday, s.TotalCards)
		}
	})
}

func TestStoreSelectionSummaryByTag(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	upsertDeck(t, store, "d1")
	upsertDeck(t, store, "d2")
	upsertEntry(t, store, "e1", "d1")
	upsertEntry(t, store, "e2", "d2")

	if err := store.queries.InsertTag(context.Background(), "greeting"); err != nil {
		t.Fatalf("InsertTag: %v", err)
	}
	if err := store.queries.InsertDeckTag(context.Background(), db.InsertDeckTagParams{
		DeckID: "d1",
		Tag:    "greeting",
	}); err != nil {
		t.Fatalf("InsertDeckTag: %v", err)
	}

	insertReviewRaw(t, store, "d1", "e1", 3, todayAt("10:00:00"))
	insertReviewRaw(t, store, "d2", "e2", 1, todayAt("11:00:00"))

	s, err := store.SelectionSummary(nil, []string{"greeting"})
	if err != nil {
		t.Fatalf("SelectionSummary: %v", err)
	}
	if s.ReviewedToday != 1 {
		t.Errorf("ReviewedToday = %d, want 1 (only d1 tagged)", s.ReviewedToday)
	}
	if s.TotalCards != 1 {
		t.Errorf("TotalCards = %d, want 1", s.TotalCards)
	}
}

func TestStoreSelectionHistory(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	upsertDeck(t, store, "d1")
	upsertDeck(t, store, "d2")
	upsertEntry(t, store, "e1", "d1")
	upsertEntry(t, store, "e2", "d2")

	insertReviewRaw(t, store, "d1", "e1", 3, "2026-08-01 10:00:00")
	insertReviewRaw(t, store, "d1", "e1", 1, "2026-08-01 15:00:00")
	insertReviewRaw(t, store, "d2", "e2", 3, "2026-08-02 10:00:00")

	t.Run("single deck", func(t *testing.T) {
		points, err := store.SelectionHistory([]string{"d1"}, nil)
		if err != nil {
			t.Fatalf("SelectionHistory: %v", err)
		}
		if len(points) != 1 {
			t.Fatalf("len = %d, want 1", len(points))
		}
		if points[0].Day != "2026-08-01" {
			t.Errorf("Day = %q, want 2026-08-01", points[0].Day)
		}
		if points[0].Correct != 1 || points[0].Incorrect != 1 {
			t.Errorf("got correct=%d incorrect=%d, want 1/1", points[0].Correct, points[0].Incorrect)
		}
	})

	t.Run("all decks", func(t *testing.T) {
		points, err := store.SelectionHistory(nil, nil)
		if err != nil {
			t.Fatalf("SelectionHistory: %v", err)
		}
		if len(points) != 2 {
			t.Fatalf("len = %d, want 2", len(points))
		}
	})

	t.Run("no reviews", func(t *testing.T) {
		points, err := store.SelectionHistory([]string{"d2"}, nil)
		if err != nil {
			t.Fatalf("SelectionHistory: %v", err)
		}
		if len(points) != 1 {
			t.Fatalf("len = %d, want 1", len(points))
		}
	})
}

func TestStoreWordStats(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	upsertDeck(t, store, "d1")
	upsertEntry(t, store, "e1", "d1")

	insertReviewRaw(t, store, "d1", "e1", 3, "2026-08-01 10:00:00")
	insertReviewRaw(t, store, "d1", "e1", 3, "2026-08-01 15:00:00")
	insertReviewRaw(t, store, "d1", "e1", 1, todayAt("09:00:00"))

	ws, err := store.WordStats("e1")
	if err != nil {
		t.Fatalf("WordStats: %v", err)
	}
	if ws.TotalReviews != 3 {
		t.Errorf("TotalReviews = %d, want 3", ws.TotalReviews)
	}
	if ws.ReviewedToday != 1 {
		t.Errorf("ReviewedToday = %d, want 1", ws.ReviewedToday)
	}
	if ws.Correct != 2 || ws.Incorrect != 1 {
		t.Errorf("got correct=%d incorrect=%d, want 2/1", ws.Correct, ws.Incorrect)
	}
	if ws.LastReviewed == nil {
		t.Fatal("LastReviewed should be set")
	}
	if ws.LastReviewed.Format("2006-01-02") != time.Now().UTC().Format("2006-01-02") {
		t.Errorf("LastReviewed date = %v, want today", ws.LastReviewed)
	}
}

func TestStoreWordHistory(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	upsertDeck(t, store, "d1")
	upsertEntry(t, store, "e1", "d1")

	insertReviewRaw(t, store, "d1", "e1", 3, "2026-08-01 10:00:00")
	insertReviewRaw(t, store, "d1", "e1", 1, "2026-08-03 10:00:00")

	points, err := store.WordHistory("e1")
	if err != nil {
		t.Fatalf("WordHistory: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("len = %d, want 2", len(points))
	}
	if points[0].Day != "2026-08-01" || points[1].Day != "2026-08-03" {
		t.Errorf("days = [%q %q], want [2026-08-01 2026-08-03]", points[0].Day, points[1].Day)
	}
}
