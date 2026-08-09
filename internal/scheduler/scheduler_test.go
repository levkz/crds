package scheduler

import (
	"testing"
	"time"

	"crds/internal/model"
)

var fixedNow = time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)

func days(n int) time.Time {
	return fixedNow.AddDate(0, 0, n)
}

func TestUpdateNewCard(t *testing.T) {
	tests := []struct {
		name  string
		grade int
		want  struct {
			ease     float64
			interval int
			due      time.Time
		}
		correct   int
		incorrect int
	}{
		{"again", GradeAgain, struct {
			ease     float64
			interval int
			due      time.Time
		}{initialEase - lapsePenalty, 0, fixedNow}, 0, 1},
		{"hard", GradeHard, struct {
			ease     float64
			interval int
			due      time.Time
		}{initialEase, 1, days(1)}, 1, 0},
		{"good", GradeGood, struct {
			ease     float64
			interval int
			due      time.Time
		}{initialEase, 1, days(1)}, 1, 0},
		{"easy", GradeEasy, struct {
			ease     float64
			interval int
			due      time.Time
		}{initialEase, 4, days(4)}, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Update(model.Progress{}, tt.grade, fixedNow)
			if got.Ease != tt.want.ease {
				t.Errorf("ease = %v, want %v", got.Ease, tt.want.ease)
			}
			if got.Interval != tt.want.interval {
				t.Errorf("interval = %d, want %d", got.Interval, tt.want.interval)
			}
			if !got.Due.Equal(tt.want.due) {
				t.Errorf("due = %v, want %v", got.Due, tt.want.due)
			}
			if got.Correct != tt.correct || got.Incorrect != tt.incorrect {
				t.Errorf("counters = correct:%d incorrect:%d, want %d/%d", got.Correct, got.Incorrect, tt.correct, tt.incorrect)
			}
		})
	}
}

func TestUpdateExistingCard(t *testing.T) {
	base := model.Progress{
		DeckID:   "fr_a1",
		EntryID:  "bonjour",
		Reverse:  false,
		Ease:     2.5,
		Interval: 5,
		Due:      fixedNow.AddDate(0, 0, -1),
		Correct:  2,
		Incorrect: 0,
	}

	t.Run("good grows by ease factor", func(t *testing.T) {
		got := Update(base, GradeGood, fixedNow)
		if got.Interval != 13 { // round(5 * 2.5) = 13
			t.Errorf("interval = %d, want 13", got.Interval)
		}
		if got.Ease != 2.5 {
			t.Errorf("ease = %v, want 2.5", got.Ease)
		}
		if !got.Due.Equal(days(13)) {
			t.Errorf("due = %v, want %v", got.Due, days(13))
		}
		if got.Correct != 3 || got.Incorrect != 0 {
			t.Errorf("counters = %d/%d, want 3/0", got.Correct, got.Incorrect)
		}
	})

	t.Run("hard grows slowly and lowers ease", func(t *testing.T) {
		got := Update(base, GradeHard, fixedNow)
		if got.Interval != 6 { // round(5 * 1.2) = 6
			t.Errorf("interval = %d, want 6", got.Interval)
		}
		if got.Ease != 2.5-easeStep {
			t.Errorf("ease = %v, want %v", got.Ease, 2.5-easeStep)
		}
	})

	t.Run("easy grows fastest and raises ease", func(t *testing.T) {
		got := Update(base, GradeEasy, fixedNow)
		// ease raised to 2.65 first, then round(5 * 2.65 * 1.3) = 17
		if got.Interval != 17 {
			t.Errorf("interval = %d, want 17", got.Interval)
		}
		if got.Ease != 2.5+easeStep {
			t.Errorf("ease = %v, want %v", got.Ease, 2.5+easeStep)
		}
	})

	t.Run("again lapses", func(t *testing.T) {
		got := Update(base, GradeAgain, fixedNow)
		if got.Interval != 0 {
			t.Errorf("interval = %d, want 0", got.Interval)
		}
		if !got.Due.Equal(fixedNow) {
			t.Errorf("due = %v, want %v", got.Due, fixedNow)
		}
		if got.Ease != 2.5-lapsePenalty {
			t.Errorf("ease = %v, want %v", got.Ease, 2.5-lapsePenalty)
		}
		if got.Correct != 2 || got.Incorrect != 1 {
			t.Errorf("counters = %d/%d, want 2/1", got.Correct, got.Incorrect)
		}
	})
}

func TestUpdateKeepDeckEntryReverse(t *testing.T) {
	prev := model.Progress{DeckID: "d1", EntryID: "e1", Reverse: true}
	got := Update(prev, GradeGood, fixedNow)
	if got.DeckID != "d1" || got.EntryID != "e1" || !got.Reverse {
		t.Errorf("identity lost: %+v", got)
	}
}

func TestEaseFloor(t *testing.T) {
	prev := model.Progress{
		Ease:     1.3,
		Interval: 2,
		Due:      fixedNow.AddDate(0, 0, -1),
	}
	got := Update(prev, GradeAgain, fixedNow)
	if got.Ease != minEase {
		t.Errorf("ease = %v, want floor %v", got.Ease, minEase)
	}
}

func TestIsCorrect(t *testing.T) {
	if IsCorrect(GradeGood) != true || IsCorrect(GradeEasy) != true {
		t.Error("Good/Easy should count as correct")
	}
	if IsCorrect(GradeAgain) != false || IsCorrect(GradeHard) != false {
		t.Error("Again/Hard should not count as correct")
	}
}

func TestUpdateHardFirstReviewKeepsEase(t *testing.T) {
	got := Update(model.Progress{}, GradeHard, fixedNow)
	if got.Ease != initialEase {
		t.Errorf("first Hard review should keep initial ease, got %v", got.Ease)
	}
}

func TestUpdateEasyFirstReviewKeepsEase(t *testing.T) {
	got := Update(model.Progress{}, GradeEasy, fixedNow)
	if got.Ease != initialEase {
		t.Errorf("first Easy review should keep initial ease, got %v", got.Ease)
	}
}