package storage

import (
	"context"
	"reflect"
	"testing"
	"time"

	"crds/internal/scheduler"
	"crds/internal/storage/db"
)

// TestRecordAnswerPersistsProgress verifies that answering a brand-new card
// writes a scheduling state (ease, interval, due, counters) via the SM-2 rule.
func TestRecordAnswerPersistsProgress(t *testing.T) {
	store := newTestStore(t)

	if err := store.RecordAnswer("deck1", "entry1", scheduler.GradeGood, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	p, err := store.queries.GetProgress(context.Background(), db.GetProgressParams{DeckID: "deck1", EntryID: "entry1", Reverse: 0})
	if err != nil {
		t.Fatalf("GetProgress: %v", err)
	}
	if p.Ease != 2.5 {
		t.Errorf("ease = %v, want 2.5", p.Ease)
	}
	if p.Interval != 1 {
		t.Errorf("interval = %d, want 1", p.Interval)
	}
	if p.Correct != 1 || p.Incorrect != 0 {
		t.Errorf("counters = %d/%d, want 1/0", p.Correct, p.Incorrect)
	}
	if p.Due == nil {
		t.Fatal("due is nil")
	}
}

// TestRecordAnswerTracksReverseSeparately verifies forward and reverse
// answers land in separate progress rows.
func TestRecordAnswerTracksReverseSeparately(t *testing.T) {
	store := newTestStore(t)

	if err := store.RecordAnswer("deck1", "entry1", scheduler.GradeAgain, true); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if _, err := store.queries.GetProgress(context.Background(), db.GetProgressParams{DeckID: "deck1", EntryID: "entry1", Reverse: 0}); err == nil {
		t.Error("forward row should not exist")
	}
	rev, err := store.queries.GetProgress(context.Background(), db.GetProgressParams{DeckID: "deck1", EntryID: "entry1", Reverse: 1})
	if err != nil {
		t.Fatalf("reverse row missing: %v", err)
	}
	if rev.Incorrect != 1 {
		t.Errorf("reverse row incorrect = %d, want 1", rev.Incorrect)
	}
}

// TestDueForSelection verifies the review queue: unseen cards first, then
// due cards; non-due cards are dropped.
func TestDueForSelection(t *testing.T) {
	store := newTestStore(t)
	setupSyncedDeck(t, store)

	// Both entries are unseen -> both in the queue, deck order.
	queue, err := store.DueForSelection([]string{"test_deck"}, nil, time.Now())
	if err != nil {
		t.Fatalf("DueForSelection: %v", err)
	}
	if want := []string{"entry_1", "entry_2"}; !reflect.DeepEqual(queue, want) {
		t.Errorf("queue = %v, want %v", queue, want)
	}

	// entry_1 answered well -> scheduled for tomorrow, leaves the queue.
	if err := store.RecordAnswer("test_deck", "entry_1", scheduler.GradeGood, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	queue, err = store.DueForSelection([]string{"test_deck"}, nil, time.Now())
	if err != nil {
		t.Fatalf("DueForSelection: %v", err)
	}
	if want := []string{"entry_2"}; !reflect.DeepEqual(queue, want) {
		t.Errorf("queue = %v, want %v", queue, want)
	}

	// entry_2 failed -> due immediately, stays in the queue.
	if err := store.RecordAnswer("test_deck", "entry_2", scheduler.GradeAgain, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	queue, err = store.DueForSelection([]string{"test_deck"}, nil, time.Now())
	if err != nil {
		t.Fatalf("DueForSelection: %v", err)
	}
	if want := []string{"entry_2"}; !reflect.DeepEqual(queue, want) {
		t.Errorf("queue = %v, want %v", queue, want)
	}
}

// TestSummaryDueToday verifies the Due Today metric counts only cards whose
// scheduling state is already past due, not unseen cards.
func TestSummaryDueToday(t *testing.T) {
	store := newTestStore(t)
	setupSyncedDeck(t, store)

	before, err := store.Summary()
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if before.DueToday != 0 {
		t.Errorf("DueToday before any reviews = %d, want 0", before.DueToday)
	}

	// Good answer schedules the card for tomorrow -> still not due today.
	if err := store.RecordAnswer("test_deck", "entry_1", scheduler.GradeGood, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	mid, err := store.Summary()
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if mid.DueToday != 0 {
		t.Errorf("DueToday after Good answer = %d, want 0", mid.DueToday)
	}

	// Again answer makes the card due now.
	if err := store.RecordAnswer("test_deck", "entry_2", scheduler.GradeAgain, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	after, err := store.Summary()
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if after.DueToday != 1 {
		t.Errorf("DueToday after Again answer = %d, want 1", after.DueToday)
	}

	sel, err := store.SelectionSummary([]string{"test_deck"}, nil)
	if err != nil {
		t.Fatalf("SelectionSummary: %v", err)
	}
	if sel.DueToday != 1 {
		t.Errorf("Selection DueToday = %d, want 1", sel.DueToday)
	}
}