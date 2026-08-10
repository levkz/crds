package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"crds/internal/fuzzy"
	"crds/internal/mapping"
	"crds/internal/stats"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
)

func TestTypingQuizInputMapping(t *testing.T) {
	store := mapping.NewStore()
	store.Resolve("fr", nil) // ensure builtin present

	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "manger",
				Front:    "manger",
				Back:     []string{"to eat"},
				Variants: []string{"to eat"},
			}},
		},
		DeckProgress: map[string]stats.EntryProgress{},
		Due:          []string{},
	})

	for _, r := range "mange/" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}

	if m.input != "mangé" {
		t.Errorf("expected input %q after typing mange/, got %q", "mangé", m.input)
	}
}

func TestTypingQuizInputMappingMidString(t *testing.T) {
	store := mapping.NewStore()
	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "x",
				Front:    "x",
				Back:     []string{"y"},
				Variants: []string{"y"},
			}},
		},
	})

	// Type "e/" normally -> mapping fires; the cursor sits on the replacement.
	for _, r := range "e/" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.input != "é" {
		t.Fatalf("expected %q, got %q", "é", m.input)
	}
	if m.cursor != 1 {
		t.Fatalf("cursor should sit on the replacement, got %d", m.cursor)
	}

	// Insert a "/" mid-string (at the start) — no suffix mapping should fire.
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyLeft}))
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'/'}}))
	if m.input != "/é" {
		t.Errorf("expected mid-string insertion untouched, got %q", m.input)
	}
}

func TestTypingQuizParseToggle(t *testing.T) {
	store := mapping.NewStore()
	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "elle",
				Front:    "she",
				Back:     []string{"elle/il"},
				Variants: []string{"elle/il"},
			}},
		},
	})

	// Parse off: the literal trigger "e/" inside "elle/il" must be preserved.
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlP}))
	if m.parseOn {
		t.Fatal("expected parse mode off after toggle")
	}
	for _, r := range "elle/il" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.input != "elle/il" {
		t.Fatalf("parse off: expected literal %q, got %q", "elle/il", m.input)
	}

	// Parse on: newly typed text parses even when an old literal trigger precedes it.
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlP}))
	if !m.parseOn {
		t.Fatal("expected parse mode on after toggle")
	}
	for _, r := range " de/" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.input != "elle/il dé" {
		t.Fatalf("parse on: expected %q, got %q", "elle/il dé", m.input)
	}
}

func TestTypingQuizParseOnMidString(t *testing.T) {
	store := mapping.NewStore()
	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "x",
				Front:    "x",
				Back:     []string{"y"},
				Variants: []string{"y"},
			}},
		},
	})

	for _, r := range "ab" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyLeft}))
	for _, r := range "e/" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.input != "aéb" {
		t.Fatalf("expected mid-string parse %q, got %q", "aéb", m.input)
	}
}

func TestTypingQuizMappingLegend(t *testing.T) {
	newModel := func() *TypingQuizModel {
		store := mapping.NewStore()
		m := NewTypingQuiz(fuzzy.Approximate, store)
		m.SetSize(120, 40)
		m.SyncState(ui.AppState{
			Deck: &ui.DeckData{
				Name:     "French",
				Language: "fr",
				Cards: []ui.CardData{{
					ID:       "x",
					Front:    "x",
					Back:     []string{"y"},
					Variants: []string{"y"},
				}},
			},
		})
		return m
	}

	m := newModel()
	// Parse on + not revealed: legend shows the resolved French triggers.
	legend := m.renderMappingLegend()
	if legend == "" {
		t.Fatal("expected a mapping legend while parsing is on")
	}
	if !strings.Contains(legend, "e/→é") {
		t.Errorf("legend should mention e/→é, got %q", legend)
	}

	// Revealed: legend hidden.
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'y'}}))
	m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	if !m.revealed {
		t.Fatal("expected revealed after submit")
	}
	if got := m.renderMappingLegend(); got != "" {
		t.Errorf("legend should be hidden after submit, got %q", got)
	}

	// Parse off + not revealed: legend hidden.
	m2 := newModel()
	m2.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlP}))
	if m2.parseOn {
		t.Fatal("expected parse toggled off")
	}
	if got := m2.renderMappingLegend(); got != "" {
		t.Errorf("legend should be hidden when parsing is off, got %q", got)
	}
}

func TestTypingQuizMappingLegendGap(t *testing.T) {
	store := mapping.NewStore()
	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SetSize(120, 40)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "x",
				Front:    "x",
				Back:     []string{"y"},
				Variants: []string{"y"},
			}},
		},
	})

	lines := strings.Split(renderer.StripANSI(m.View()), "\n")
	gap := ui.Theme.Spacing.Xxs
	footerIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "card 1/1") {
			footerIdx = i
			break
		}
	}
	if footerIdx == -1 {
		t.Fatal("footer not found in view")
	}
	if footerIdx < gap {
		t.Fatalf("not enough lines above footer: %d", footerIdx)
	}
	for i := footerIdx - 1; i >= footerIdx-gap; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			t.Errorf("expected blank line at index %d, got %q", i, lines[i])
		}
	}
	hasLegend := false
	for i := 0; i < footerIdx-gap; i++ {
		if strings.Contains(lines[i], "→") {
			hasLegend = true
		}
	}
	if !hasLegend {
		t.Error("expected legend above the gap")
	}
}

func TestTypingQuizOELigature(t *testing.T) {
	store := mapping.NewStore()
	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "coeur",
				Front:    "heart",
				Back:     []string{"cœur"},
				Variants: []string{"cœur"},
			}},
		},
	})

	for _, r := range "coeur" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if m.input != "cœur" {
		t.Errorf("expected %q after typing coeur, got %q", "cœur", m.input)
	}
}

func TestTypingQuizApproximateGrading(t *testing.T) {
	store := mapping.NewStore()

	m := NewTypingQuiz(fuzzy.Approximate, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "cafe",
				Front:    "coffee",
				Back:     []string{"café"},
				Variants: []string{"café"},
			}},
		},
	})

	// Type "cafe" (no accent) for the accented answer "café" — approximate mode grades Good.
	for _, r := range "cafe" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	if !m.gradeInput() {
		t.Fatal("gradeInput returned false")
	}
	if m.grade != fuzzy.Good {
		t.Errorf("approximate mode: expected Good, got %d", m.grade)
	}
}

func TestTypingQuizStrictGrading(t *testing.T) {
	store := mapping.NewStore()

	m := NewTypingQuiz(fuzzy.Strict, store)
	m.SyncState(ui.AppState{
		Deck: &ui.DeckData{
			Name:     "French",
			Language: "fr",
			Cards: []ui.CardData{{
				ID:       "cafe",
				Front:    "coffee",
				Back:     []string{"café"},
				Variants: []string{"café"},
			}},
		},
	})

	// "cafe" for "café" in strict mode is not an exact match.
	for _, r := range "cafe" {
		m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	}
	m.gradeInput()
	if m.grade == fuzzy.Good {
		t.Errorf("strict mode: expected not Good for missing accent, got %d", m.grade)
	}
}
