package ui

import (
	"reflect"
	"testing"

	"crds/internal/stats"
)

func testCards() []CardData {
	return []CardData{
		{ID: "a", Front: "alpha", Back: []string{"a"}},
		{ID: "b", Front: "beta", Back: []string{"b"}},
		{ID: "c", Front: "gamma", Back: []string{"c"}},
	}
}

func TestSortCardsDueFiltersToDueSet(t *testing.T) {
	cards := testCards()
	got := SortCards(QuizModeDue, cards, nil, []string{"c", "a"})
	want := []CardData{cards[2], cards[0]}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("due sort = %+v, want %+v", got, want)
	}
}

func TestSortCardsDueEmpty(t *testing.T) {
	got := SortCards(QuizModeDue, testCards(), nil, nil)
	if len(got) != 0 {
		t.Errorf("expected empty due queue, got %d cards", len(got))
	}
}

func TestSortCardsDuePreservesOrder(t *testing.T) {
	cards := testCards()
	got := SortCards(QuizModeDue, cards, nil, []string{"a", "b", "c"})
	if !reflect.DeepEqual(got, cards) {
		t.Errorf("due sort should preserve deck order, got %+v", got)
	}
}

func TestSortCardsOtherModesIgnoreDueSet(t *testing.T) {
	for _, mode := range []QuizMode{QuizModeNormal, QuizModeRandom, QuizModeSmart, QuizModeKindaSmart} {
		got := SortCards(mode, testCards(), map[string]stats.EntryProgress{}, nil)
		if len(got) != 3 {
			t.Errorf("mode %v should keep all cards, got %d", mode, len(got))
		}
	}
}