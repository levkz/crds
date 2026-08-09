package components

import (
	"strings"
	"testing"

	"crds/internal/stats"
)

func TestGraphEmpty(t *testing.T) {
	if got := Graph(nil, 60); got != "No reviews yet" {
		t.Errorf("Graph(nil) = %q, want %q", got, "No reviews yet")
	}
}

func TestGraphZeroWidth(t *testing.T) {
	pts := []GraphPoint{{Day: "04-08", Correct: 1, Incorrect: 0}}
	if got := Graph(pts, 0); got != "" {
		t.Errorf("Graph(_, 0) = %q, want empty", got)
	}
}

func TestGraphRendersRows(t *testing.T) {
	pts := []GraphPoint{
		{Day: "04-08", Correct: 9, Incorrect: 1},
		{Day: "04-09", Correct: 0, Incorrect: 3},
	}
	out := Graph(pts, 40)

	for _, want := range []string{"04-08", "04-09", "90%", "(9/10)", "(0/3)"} {
		if !strings.Contains(out, want) {
			t.Errorf("Graph output missing %q:\n%s", want, out)
		}
	}
}

func TestGraphUnseenDay(t *testing.T) {
	pts := []GraphPoint{{Day: "04-08", Correct: 0, Incorrect: 0}}
	out := Graph(pts, 40)
	if !strings.Contains(out, "(0/0)") {
		t.Errorf("Graph output missing (0/0):\n%s", out)
	}
}

func TestToGraphPoints(t *testing.T) {
	days := []stats.DayPoint{
		{Day: "2026-04-08", Correct: 1, Incorrect: 0},
	}
	pts := ToGraphPoints(days)
	if len(pts) != 1 {
		t.Fatalf("ToGraphPoints length = %d, want 1", len(pts))
	}
	if pts[0].Day != "04-08" {
		t.Errorf("label = %q, want %q", pts[0].Day, "04-08")
	}
	if pts[0].Correct != 1 || pts[0].Incorrect != 0 {
		t.Errorf("got %+v, want correct=1 incorrect=0", pts[0])
	}
}
