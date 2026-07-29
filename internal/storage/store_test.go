package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"crds/internal/storage/db"

	"github.com/pressly/goose/v3"
)

// newTestStore creates a Store backed by an in-memory SQLite database
// for use in tests. Callers must close the returned store.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}

	// Run embedded migrations
	runMigrations(t, conn)

	return &Store{
		queries: db.New(conn),
		conn:    conn,
	}
}

func runMigrations(t *testing.T, conn *sql.DB) {
	t.Helper()
	goose.SetBaseFS(migrationsFS)
	goose.SetDialect("sqlite")
	if err := goose.Up(conn, "migrations"); err != nil {
		t.Fatalf("migrations: %v", err)
	}
}

func TestStoreCreateSession(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	id, err := store.EnsureSession()
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero session ID")
	}
}

func TestStoreRecordAnswer(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	if err := store.RecordAnswer("", "test_entry", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	stats := store.Stats()
	if stats.ReviewedToday != 1 {
		t.Errorf("expected 1 review today, got %d", stats.ReviewedToday)
	}
	if stats.Accuracy != 100 {
		t.Errorf("expected 100%% accuracy, got %.0f%%", stats.Accuracy)
	}
}

func TestStoreRecordAnswerFull(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	sessionID, err := store.EnsureSession()
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}

	reviewID, err := store.RecordAnswerFull(
		sessionID, "test_deck", "test_entry", 2, false,
		"useur input", "user input", 0.85,
	)
	if err != nil {
		t.Fatalf("RecordAnswerFull: %v", err)
	}

	// Verify the typing detail was stored
	detail, err := store.queries.GetTypingDetailByReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("GetTypingDetailByReview: %v", err)
	}

	if detail.UserInput != "useur input" {
		t.Errorf("expected UserInput 'useur input', got %q", detail.UserInput)
	}
	if detail.CorrectAnswer != "user input" {
		t.Errorf("expected CorrectAnswer 'user input', got %q", detail.CorrectAnswer)
	}
	if detail.Similarity != 0.85 {
		t.Errorf("expected Similarity 0.85, got %f", detail.Similarity)
	}
}

func TestStoreAnswerStats(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	if err := store.RecordAnswer("", "entry_a", 3, false); err != nil {
		t.Fatalf("RecordAnswer grade 3: %v", err)
	}
	if err := store.RecordAnswer("", "entry_b", 2, false); err != nil {
		t.Fatalf("RecordAnswer grade 2: %v", err)
	}
	if err := store.RecordAnswer("", "entry_c", 1, false); err != nil {
		t.Fatalf("RecordAnswer grade 1: %v", err)
	}
	if err := store.RecordAnswer("", "entry_d", 4, false); err != nil {
		t.Fatalf("RecordAnswer grade 4: %v", err)
	}

	stats := store.Stats()
	if stats.ReviewedToday != 4 {
		t.Errorf("expected 4 reviews today, got %d", stats.ReviewedToday)
	}
	if stats.Accuracy != 50 {
		t.Errorf("expected 50%% accuracy (2 out of 4), got %.0f%%", stats.Accuracy)
	}
}

func TestStoreSessionReset(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	// Record in the first session
	if err := store.RecordAnswer("", "entry1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	session1 := store.currentSession
	if session1 == 0 {
		t.Fatal("expected a session to be created")
	}

	// Reset session
	if err := store.ResetSession(); err != nil {
		t.Fatalf("ResetSession: %v", err)
	}

	// Record in the second session
	if err := store.RecordAnswer("", "entry2", 1, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	session2 := store.currentSession

	if session1 == session2 {
		t.Error("expected different session IDs after reset")
	}
}

func TestStoreWeakTypingEntries(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	sessionID, err := store.EnsureSession()
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}

	// Low similarity -> weak entry
	_, err = store.RecordAnswerFull(sessionID, "deck1", "entry1", 1, false, "wrong", "correct", 0.3)
	if err != nil {
		t.Fatalf("RecordAnswerFull weak: %v", err)
	}

	// High similarity -> not weak
	_, err = store.RecordAnswerFull(sessionID, "deck1", "entry2", 3, false, "correct", "correct", 1.0)
	if err != nil {
		t.Fatalf("RecordAnswerFull good: %v", err)
	}

	weak, err := store.GetWeakTypingEntries("deck1", 10)
	if err != nil {
		t.Fatalf("GetWeakTypingEntries: %v", err)
	}

	if len(weak) != 1 {
		t.Fatalf("expected 1 weak entry, got %d", len(weak))
	}
	if weak[0].EntryID != "entry1" {
		t.Errorf("expected entry1, got %q", weak[0].EntryID)
	}
}

func TestStoreProgress(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// Upsert progress
	due := timeOnly("2026-07-22 12:00:00")
	err := store.queries.UpsertProgress(ctx, db.UpsertProgressParams{
		DeckID:    "french_a1",
		EntryID:   "fr_bonjour",
		Reverse:   0,
		Ease:      2.5,
		Interval:  1,
		Due:       &due,
		Correct:   3,
		Incorrect: 0,
	})
	if err != nil {
		t.Fatalf("UpsertProgress: %v", err)
	}

	// Read it back
	p, err := store.queries.GetProgress(ctx, db.GetProgressParams{
		DeckID:  "french_a1",
		EntryID: "fr_bonjour",
		Reverse: 0,
	})
	if err != nil {
		t.Fatalf("GetProgress: %v", err)
	}

	if p.Ease != 2.5 {
		t.Errorf("expected ease 2.5, got %f", p.Ease)
	}
	if p.Correct != 3 {
		t.Errorf("expected correct 3, got %d", p.Correct)
	}
}

func TestStoreGetReviewsByEntry(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	store.RecordAnswer("", "entry_x", 3, false)
	store.RecordAnswer("", "entry_y", 2, false)
	store.RecordAnswer("", "entry_x", 4, false)

	reviews, err := store.GetReviewsByEntry("entry_x", 5)
	if err != nil {
		t.Fatalf("GetReviewsByEntry: %v", err)
	}

	if len(reviews) != 2 {
		t.Fatalf("expected 2 reviews for entry_x, got %d", len(reviews))
	}
	grades := map[int64]bool{}
	for _, r := range reviews {
		grades[r.Grade] = true
	}
	if !grades[3] || !grades[4] {
		t.Errorf("expected grades 3 and 4 for entry_x, got %v", reviews)
	}
}

func TestStoreListAllTags(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// Insert tags directly
	for _, tag := range []string{"greeting", "verb", "noun", "A1"} {
		if err := store.queries.InsertTag(ctx, tag); err != nil {
			t.Fatalf("InsertTag: %v", err)
		}
	}

	tags, err := store.ListAllTags()
	if err != nil {
		t.Fatalf("ListAllTags: %v", err)
	}

	expected := []string{"A1", "greeting", "noun", "verb"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(tags), tags)
	}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("tags[%d] = %q, want %q", i, tag, expected[i])
		}
	}
}

func TestStoreListDeckTags(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// Insert decks
	for _, id := range []string{"french_a1", "spanish_b1"} {
		if err := store.queries.UpsertDeck(ctx, db.UpsertDeckParams{
			ID:                  id,
			Name:                id,
			Language:            "fr",
			TranslationLanguage: "en",
		}); err != nil {
			t.Fatalf("UpsertDeck: %v", err)
		}
	}

	// Insert tags
	for _, tag := range []string{"greeting", "verb", "noun", "A1", "B1"} {
		if err := store.queries.InsertTag(ctx, tag); err != nil {
			t.Fatalf("InsertTag: %v", err)
		}
	}

	// Insert deck_tags
	for _, dt := range []struct{ deck, tag string }{
		{"french_a1", "greeting"},
		{"french_a1", "noun"},
		{"french_a1", "A1"},
		{"spanish_b1", "greeting"},
		{"spanish_b1", "verb"},
		{"spanish_b1", "B1"},
	} {
		if err := store.queries.InsertDeckTag(ctx, db.InsertDeckTagParams{
			DeckID: dt.deck,
			Tag:    dt.tag,
		}); err != nil {
			t.Fatalf("InsertDeckTag: %v", err)
		}
	}

	// Test french_a1 tags
	tags, err := store.ListDeckTags("french_a1")
	if err != nil {
		t.Fatalf("ListDeckTags: %v", err)
	}
	expected := []string{"A1", "greeting", "noun"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags for french_a1, got %d: %v", len(expected), len(tags), tags)
	}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("tags[%d] = %q, want %q", i, tag, expected[i])
		}
	}

	// Test spanish_b1 tags
	tags, err = store.ListDeckTags("spanish_b1")
	if err != nil {
		t.Fatalf("ListDeckTags: %v", err)
	}
	expected = []string{"B1", "greeting", "verb"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags for spanish_b1, got %d: %v", len(expected), len(tags), tags)
	}
}

func TestStoreListDecksByTag(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	for _, id := range []string{"french_a1", "spanish_b1", "japanese_n5"} {
		if err := store.queries.UpsertDeck(ctx, db.UpsertDeckParams{
			ID:                  id,
			Name:                id,
			Language:            "xx",
			TranslationLanguage: "en",
		}); err != nil {
			t.Fatalf("UpsertDeck: %v", err)
		}
	}

	for _, tag := range []string{"greeting", "verb"} {
		if err := store.queries.InsertTag(ctx, tag); err != nil {
			t.Fatalf("InsertTag: %v", err)
		}
	}

	for _, dt := range []struct{ deck, tag string }{
		{"french_a1", "greeting"},
		{"spanish_b1", "greeting"},
		{"spanish_b1", "verb"},
		{"japanese_n5", "verb"},
	} {
		if err := store.queries.InsertDeckTag(ctx, db.InsertDeckTagParams{
			DeckID: dt.deck,
			Tag:    dt.tag,
		}); err != nil {
			t.Fatalf("InsertDeckTag: %v", err)
		}
	}

	// Decks with "greeting"
	decks, err := store.ListDecksByTag("greeting")
	if err != nil {
		t.Fatalf("ListDecksByTag: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("expected 2 decks with 'greeting', got %d: %v", len(decks), decks)
	}
	if decks[0] != "french_a1" || decks[1] != "spanish_b1" {
		t.Errorf("expected [french_a1, spanish_b1], got %v", decks)
	}
}

func TestStoreFilterDecksByTags(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	for _, id := range []string{"french_a1", "spanish_b1", "japanese_n5"} {
		if err := store.queries.UpsertDeck(ctx, db.UpsertDeckParams{
			ID:                  id,
			Name:                id,
			Language:            "xx",
			TranslationLanguage: "en",
		}); err != nil {
			t.Fatalf("UpsertDeck: %v", err)
		}
	}

	for _, tag := range []string{"greeting", "verb", "noun"} {
		if err := store.queries.InsertTag(ctx, tag); err != nil {
			t.Fatalf("InsertTag: %v", err)
		}
	}

	for _, dt := range []struct{ deck, tag string }{
		{"french_a1", "greeting"},
		{"french_a1", "verb"},
		{"spanish_b1", "greeting"},
		{"japanese_n5", "verb"},
		{"japanese_n5", "noun"},
	} {
		if err := store.queries.InsertDeckTag(ctx, db.InsertDeckTagParams{
			DeckID: dt.deck,
			Tag:    dt.tag,
		}); err != nil {
			t.Fatalf("InsertDeckTag: %v", err)
		}
	}

	// AND: decks that have BOTH "greeting" AND "verb"
	decks, err := store.FilterDecksByTags([]string{"greeting", "verb"})
	if err != nil {
		t.Fatalf("FilterDecksByTags: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("expected 1 deck with greeting+verb, got %d: %v", len(decks), decks)
	}
	if decks[0] != "french_a1" {
		t.Errorf("expected french_a1, got %q", decks[0])
	}

	// Empty filter returns all decks
	decks, err = store.FilterDecksByTags([]string{})
	if err != nil {
		t.Fatalf("FilterDecksByTags empty: %v", err)
	}
	if len(decks) != 3 {
		t.Errorf("expected 3 decks with empty filter, got %d", len(decks))
	}
}

func TestStoreFilterTagsByDecks(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	ctx := context.Background()

	for _, id := range []string{"french_a1", "spanish_b1", "japanese_n5"} {
		if err := store.queries.UpsertDeck(ctx, db.UpsertDeckParams{
			ID:                  id,
			Name:                id,
			Language:            "xx",
			TranslationLanguage: "en",
		}); err != nil {
			t.Fatalf("UpsertDeck: %v", err)
		}
	}

	for _, tag := range []string{"greeting", "verb", "noun", "A1"} {
		if err := store.queries.InsertTag(ctx, tag); err != nil {
			t.Fatalf("InsertTag: %v", err)
		}
	}

	for _, dt := range []struct{ deck, tag string }{
		{"french_a1", "greeting"},
		{"french_a1", "verb"},
		{"french_a1", "A1"},
		{"spanish_b1", "greeting"},
		{"spanish_b1", "verb"},
		{"japanese_n5", "verb"},
		{"japanese_n5", "noun"},
	} {
		if err := store.queries.InsertDeckTag(ctx, db.InsertDeckTagParams{
			DeckID: dt.deck,
			Tag:    dt.tag,
		}); err != nil {
			t.Fatalf("InsertDeckTag: %v", err)
		}
	}

	// Intersection: tags common to french_a1 AND spanish_b1
	tags, err := store.FilterTagsByDecks([]string{"french_a1", "spanish_b1"})
	if err != nil {
		t.Fatalf("FilterTagsByDecks: %v", err)
	}
	expected := []string{"greeting", "verb"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(tags), tags)
	}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("tags[%d] = %q, want %q", i, tag, expected[i])
		}
	}

	// Empty filter returns all tags
	tags, err = store.FilterTagsByDecks([]string{})
	if err != nil {
		t.Fatalf("FilterTagsByDecks empty: %v", err)
	}
	if len(tags) != 4 {
		t.Errorf("expected 4 tags with empty filter, got %d", len(tags))
	}
}

// timeOnly is a helper that parses "2006-01-02 15:04:05" format.
func timeOnly(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
