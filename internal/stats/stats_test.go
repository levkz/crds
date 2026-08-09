package stats

import (
	"testing"
	"time"
)

func TestConfidence(t *testing.T) {
	tests := []struct {
		name      string
		correct   int
		incorrect int
		want      float64
	}{
		{
			name:      "unseen returns neutral",
			correct:   0,
			incorrect: 0,
			want:      0.5,
		},
		{
			name:      "all correct",
			correct:   5,
			incorrect: 0,
			want:      1.0,
		},
		{
			name:      "all wrong",
			correct:   0,
			incorrect: 5,
			want:      0.0,
		},
		{
			name:      "mixed 3/2",
			correct:   3,
			incorrect: 2,
			want:      0.6,
		},
		{
			name:      "single correct",
			correct:   1,
			incorrect: 0,
			want:      1.0,
		},
		{
			name:      "single wrong",
			correct:   0,
			incorrect: 1,
			want:      0.0,
		},
		{
			name:      "even split",
			correct:   10,
			incorrect: 10,
			want:      0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Confidence(tt.correct, tt.incorrect)
			if got != tt.want {
				t.Errorf("Confidence(%d, %d) = %f, want %f", tt.correct, tt.incorrect, got, tt.want)
			}
		})
	}
}

func TestEntryProgressConfidence(t *testing.T) {
	tests := []struct {
		name      string
		correct   int
		incorrect int
		want      float64
	}{
		{name: "unseen", correct: 0, incorrect: 0, want: 0.5},
		{name: "perfect", correct: 5, incorrect: 0, want: 1.0},
		{name: "zero", correct: 0, incorrect: 5, want: 0.0},
		{name: "mixed", correct: 3, incorrect: 2, want: 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := EntryProgress{Correct: tt.correct, Incorrect: tt.incorrect}
			got := p.Confidence()
			if got != tt.want {
				t.Errorf("EntryProgress{Correct: %d, Incorrect: %d}.Confidence() = %f, want %f",
					tt.correct, tt.incorrect, got, tt.want)
			}
		})
	}
}

func TestDayPointConfidence(t *testing.T) {
	tests := []struct {
		name string
		p    DayPoint
		want float64
	}{
		{"unseen", DayPoint{}, 0.5},
		{"perfect", DayPoint{Correct: 3, Incorrect: 0}, 1.0},
		{"none correct", DayPoint{Correct: 0, Incorrect: 4}, 0.0},
		{"mixed", DayPoint{Correct: 2, Incorrect: 2}, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Confidence(); got != tt.want {
				t.Errorf("Confidence() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestWordStats(t *testing.T) {
	ws := WordStats{TotalReviews: 10, ReviewedToday: 2, Correct: 8, Incorrect: 2}
	if got := ws.Accuracy(); got != 80 {
		t.Errorf("Accuracy() = %f, want 80", got)
	}
	if got := ws.Confidence(); got != 0.8 {
		t.Errorf("Confidence() = %f, want 0.8", got)
	}
	if !ws.Mastered() {
		t.Error("Mastered() should be true at confidence 0.8")
	}

	weak := WordStats{TotalReviews: 2, Correct: 1, Incorrect: 1}
	if weak.Mastered() {
		t.Error("Mastered() should be false at confidence 0.5")
	}

	unseen := WordStats{}
	if got := unseen.Accuracy(); got != 0 {
		t.Errorf("Accuracy() for unseen = %f, want 0", got)
	}
	if got := unseen.Confidence(); got != 0.5 {
		t.Errorf("Confidence() for unseen = %f, want 0.5", got)
	}
	if unseen.Mastered() {
		t.Error("Mastered() for unseen should be false")
	}
}

func TestStreak(t *testing.T) {
	day := func(offset int) time.Time {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, offset)
	}

	tests := []struct {
		name string
		days []time.Time
		want int
	}{
		{"empty", nil, 0},
		{"single today", []time.Time{day(0)}, 1},
		{"single yesterday", []time.Time{day(-1)}, 1},
		{"single older than yesterday", []time.Time{day(-3)}, 0},
		{"two consecutive ending today", []time.Time{day(0), day(-1)}, 2},
		{"three consecutive ending yesterday", []time.Time{day(-1), day(-2), day(-3)}, 3},
		{"gap breaks streak", []time.Time{day(0), day(-2)}, 1},
		{"no today no yesterday", []time.Time{day(-2), day(-3)}, 0},
		{"duplicates counted once", []time.Time{day(0), day(0), day(-1)}, 2},
		{"unsorted input", []time.Time{day(-2), day(0), day(-1)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Streak(tt.days); got != tt.want {
				t.Errorf("Streak() = %d, want %d", got, tt.want)
			}
		})
	}
}
