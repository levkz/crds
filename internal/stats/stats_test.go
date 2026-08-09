package stats

import (
	"testing"
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
