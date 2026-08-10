package app

import (
	"testing"

	cfg "crds/internal/config"
	"crds/internal/fuzzy"
	"crds/internal/ui"
)

func TestConfigDefaultMatchingMode(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MatchingMode != fuzzy.Approximate {
		t.Errorf("default MatchingMode = %v, want approximate", cfg.MatchingMode)
	}
	if cfg.Mappings == nil {
		t.Error("default Mappings store should not be nil")
	}
}

func TestConfigApplyYAML(t *testing.T) {
	base := DefaultConfig()

	t.Run("matching_mode strict", func(t *testing.T) {
		got := base.ApplyYAML(&cfg.ConfigYAML{MatchingMode: "strict"})
		if got.MatchingMode != fuzzy.Strict {
			t.Errorf("MatchingMode = %v, want strict", got.MatchingMode)
		}
	})

	t.Run("matching_mode approximate", func(t *testing.T) {
		got := base.ApplyYAML(&cfg.ConfigYAML{MatchingMode: "approximate"})
		if got.MatchingMode != fuzzy.Approximate {
			t.Errorf("MatchingMode = %v, want approximate", got.MatchingMode)
		}
	})

	t.Run("empty keeps default", func(t *testing.T) {
		got := base.ApplyYAML(&cfg.ConfigYAML{})
		if got.MatchingMode != fuzzy.Approximate {
			t.Errorf("MatchingMode = %v, want default approximate", got.MatchingMode)
		}
	})

	t.Run("nil keeps default", func(t *testing.T) {
		got := base.ApplyYAML(nil)
		if got.MatchingMode != fuzzy.Approximate {
			t.Errorf("MatchingMode = %v, want default approximate", got.MatchingMode)
		}
	})

	t.Run("quiz mode still applied", func(t *testing.T) {
		got := base.ApplyYAML(&cfg.ConfigYAML{QuizMode: "due"})
		if got.QuizMode != ui.ParseQuizMode("due") {
			t.Errorf("QuizMode = %v, want due", got.QuizMode)
		}
	})
}

func TestMergeDecks(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		merged := mergeDecks(nil)
		if merged.Name != "" || len(merged.Cards) != 0 {
			t.Errorf("expected empty merge, got %+v", merged)
		}
	})

	t.Run("single deck passthrough", func(t *testing.T) {
		d := ui.DeckData{Name: "A", Language: "fr", Cards: []ui.CardData{{ID: "x"}}}
		merged := mergeDecks([]ui.DeckData{d})
		if merged.Name != "A" || merged.Language != "fr" || len(merged.Cards) != 1 {
			t.Errorf("expected passthrough, got %+v", merged)
		}
	})

	t.Run("merges cards names and mappings", func(t *testing.T) {
		d1 := ui.DeckData{
			Name:     "French A",
			Language: "fr",
			InputMappings: map[string]string{"e/": "é", "a/": "á"},
			Cards:    []ui.CardData{{ID: "a"}},
		}
		d2 := ui.DeckData{
			Name:     "French B",
			InputMappings: map[string]string{"a/": "à", "o/": "ó"},
			Cards:    []ui.CardData{{ID: "b"}},
		}
		merged := mergeDecks([]ui.DeckData{d1, d2})

		if merged.Name != "French A + French B" {
			t.Errorf("Name = %q", merged.Name)
		}
		if merged.Language != "fr" {
			t.Errorf("Language should come from first deck, got %q", merged.Language)
		}
		if len(merged.Cards) != 2 {
			t.Errorf("expected 2 cards, got %d", len(merged.Cards))
		}
		if merged.InputMappings["e/"] != "é" {
			t.Errorf("e/ mapping lost from first deck")
		}
		if merged.InputMappings["a/"] != "à" {
			t.Errorf("later deck should override a/ mapping")
		}
		if merged.InputMappings["o/"] != "ó" {
			t.Errorf("o/ mapping from second deck missing")
		}
	})

	t.Run("second deck with no mappings keeps first", func(t *testing.T) {
		d1 := ui.DeckData{Name: "A", InputMappings: map[string]string{"e/": "é"}}
		d2 := ui.DeckData{Name: "B"}
		merged := mergeDecks([]ui.DeckData{d1, d2})
		if merged.InputMappings["e/"] != "é" {
			t.Errorf("first deck mappings should be preserved, got %v", merged.InputMappings)
		}
	})
}
