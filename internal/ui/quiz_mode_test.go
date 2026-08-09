package ui

import "testing"

func TestQuizModeDueString(t *testing.T) {
	if QuizModeDue.String() != "due" {
		t.Errorf("QuizModeDue.String() = %q, want \"due\"", QuizModeDue.String())
	}
}

func TestQuizModeParseDue(t *testing.T) {
	if got := ParseQuizMode("due"); got != QuizModeDue {
		t.Errorf("ParseQuizMode(due) = %v, want QuizModeDue", got)
	}
}

func TestQuizModeNextCycle(t *testing.T) {
	seq := make([]QuizMode, 0, NumQuizModes+1)
	m := QuizModeNormal
	for i := 0; i <= int(NumQuizModes); i++ {
		seq = append(seq, m)
		m = m.Next()
	}
	for i := 0; i < len(seq); i++ {
		if seq[i] != QuizMode(i%int(NumQuizModes)) {
			t.Fatalf("mode cycle off at %d: %v", i, seq[i])
		}
	}
	if seq[0] != seq[len(seq)-1] {
		t.Errorf("mode cycle should wrap around: %v vs %v", seq[0], seq[len(seq)-1])
	}
}