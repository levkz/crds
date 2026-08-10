package screens

import (
	"strings"
	"testing"

	"crds/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestSearch() *SearchModel {
	m := NewSearch()
	m.SetSize(80, 24)
	return m
}

func TestFilterCardsPreservesTagsAndExamples(t *testing.T) {
	m := newTestSearch()
	m.cards = []ui.CardData{
		{ID: "bonjour", Front: "bonjour", Back: []string{"hello"}, Tags: []string{"greeting"}, Examples: []ui.ExampleData{{Text: "Bonjour madame"}}},
	}
	m.query = "bon"
	m.filterResults()

	if len(m.results) != 1 {
		t.Fatalf("results = %d, want 1", len(m.results))
	}
	r := m.results[0]
	if len(r.tags) != 1 || r.tags[0] != "greeting" {
		t.Errorf("tags = %v, want [greeting]", r.tags)
	}
	if len(r.examples) != 1 || r.examples[0].Text != "Bonjour madame" {
		t.Errorf("examples = %v, want [Bonjour madame]", r.examples)
	}
}

func TestSearchSelectWordCarriesTagsAndExamples(t *testing.T) {
	m := newTestSearch()
	m.cards = []ui.CardData{
		{
			ID:    "bonjour",
			Front: "bonjour",
			Back:  []string{"hello"},
			Tags:  []string{"greeting", "formal"},
			Examples: []ui.ExampleData{
				{Text: "Bonjour madame", Translation: "Hello ma'am"},
			},
		},
	}
	m.query = "bonjour"
	m.filterResults()
	m.mode = searchResults

	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	if cmd == nil {
		t.Fatal("expected a command from result selection")
	}
	msg := cmd()
	nav, ok := msg.(ui.NavigateToDetailMsg)
	if !ok {
		t.Fatalf("expected NavigateToDetailMsg, got %T", msg)
	}
	if nav.Entry.ID != "bonjour" {
		t.Errorf("Entry.ID = %q, want bonjour", nav.Entry.ID)
	}
	if len(nav.Entry.Tags) != 2 || nav.Entry.Tags[0] != "greeting" {
		t.Errorf("Entry.Tags = %v, want [greeting formal]", nav.Entry.Tags)
	}
	if len(nav.Entry.Examples) != 1 || nav.Entry.Examples[0].Text != "Bonjour madame" {
		t.Errorf("Entry.Examples = %v, want [Bonjour madame]", nav.Entry.Examples)
	}
}

func TestDetailRendersTagsAndExamplesLikeQuiz(t *testing.T) {
	m := NewDetail()
	m.SetSize(80, 24)
	m.SetEntry(ui.CardData{
		ID:    "bonjour",
		Front: "bonjour",
		Back:  []string{"hello"},
		Tags:  []string{"greeting"},
		Examples: []ui.ExampleData{
			{Text: "Bonjour madame", Translation: "Hello ma'am"},
		},
	})

	out := m.View()
	for _, want := range []string{"greeting", "Bonjour madame"} {
		if !strings.Contains(out, want) {
			t.Errorf("detail view missing %q\n%s", want, out)
		}
	}
}
