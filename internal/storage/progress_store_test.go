package storage

import "testing"

func TestProgressStore_RecordAnswer(t *testing.T) {
	s := NewProgressStore()

	if err := s.RecordAnswer("card1", 3); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := s.RecordAnswer("card2", 3); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
}

func TestProgressStore_Stats_Empty(t *testing.T) {
	s := NewProgressStore()

	stats := s.Stats()
	if stats.ReviewedToday != 0 {
		t.Errorf("expected 0 reviewed, got %d", stats.ReviewedToday)
	}
	if stats.Accuracy != 0 {
		t.Errorf("expected 0 accuracy, got %f", stats.Accuracy)
	}
	if stats.TotalCards != 0 {
		t.Errorf("expected 0 total cards, got %d", stats.TotalCards)
	}
}

func TestProgressStore_Stats_WithReviews(t *testing.T) {
	s := NewProgressStore()

	_ = s.RecordAnswer("card1", 4)
	_ = s.RecordAnswer("card2", 3)
	_ = s.RecordAnswer("card3", 1)

	stats := s.Stats()

	if stats.ReviewedToday != 3 {
		t.Errorf("expected 3 reviewed today, got %d", stats.ReviewedToday)
	}
	if stats.Accuracy != 66.66666666666666 {
		t.Errorf("expected ~66.67 accuracy, got %f", stats.Accuracy)
	}
	if stats.TotalCards != 3 {
		t.Errorf("expected 3 total cards, got %d", stats.TotalCards)
	}
}

func TestProgressStore_Stats_GradeBoundary(t *testing.T) {
	s := NewProgressStore()

	_ = s.RecordAnswer("card1", 3)
	_ = s.RecordAnswer("card2", 2)

	stats := s.Stats()

	if stats.Accuracy != 50 {
		t.Errorf("expected 50 accuracy, got %f", stats.Accuracy)
	}
}
