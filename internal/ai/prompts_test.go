package ai

import (
	"strings"
	"testing"

	"crds/internal/model"
)

func TestInterpretMessages_IncludesRawAndLanguages(t *testing.T) {
	system, user := InterpretMessages("manger\nboire", LanguageContext{Language: "fr", TranslationLanguage: "en"})

	if !strings.Contains(user, "manger") || !strings.Contains(user, "boire") {
		t.Errorf("user message should include raw input")
	}
	if !strings.Contains(user, "Source language: fr") {
		t.Errorf("user message should include source language")
	}
	if !strings.Contains(user, "Target language for translations: en") {
		t.Errorf("user message should include target language")
	}
	if !strings.Contains(system, "YAML") {
		t.Errorf("system message should mention YAML")
	}
}

func TestInterpretMessages_NoLanguages(t *testing.T) {
	_, user := InterpretMessages("just a word", LanguageContext{})
	if strings.Contains(user, "Source language") {
		t.Errorf("without a deck, languages should be omitted")
	}
}

func TestFillMessages_IncludesAllowedTags(t *testing.T) {
	entries := []model.Entry{
		{Term: "baguette", Translations: []model.Translation{{Text: "baguette bread"}}},
	}
	system, user := FillMessages(entries, DeckContext{
		Language:            "fr",
		TranslationLanguage: "en",
		AllowedTags:         []string{"food", "A1"},
	})

	if !strings.Contains(user, "allowed tags: food, A1") {
		t.Errorf("user message should list allowed tags, got:\n%s", user)
	}
	if !strings.Contains(user, "baguette") {
		t.Errorf("user message should include the input entries")
	}
	if !strings.Contains(system, "(une/la)") {
		t.Errorf("system message should describe convention rules")
	}
	if !strings.Contains(system, "(to)") {
		t.Errorf("system message should describe the verb convention")
	}
}

func TestFillMessages_NoTags(t *testing.T) {
	_, user := FillMessages(nil, DeckContext{AllowedTags: nil})
	if !strings.Contains(user, "do not add tags") {
		t.Errorf("empty tag allowlist should instruct no tags, got:\n%s", user)
	}
}

func TestFillMessages_IncludesSamples(t *testing.T) {
	samples := []model.Entry{
		{Term: "(un/le) chat", Translations: []model.Translation{{Text: "(to) cat"}}},
	}
	_, user := FillMessages(nil, DeckContext{AllowedTags: []string{"animal"}, Samples: samples})

	if !strings.Contains(user, "(un/le) chat") {
		t.Errorf("user message should include sample entries")
	}
}

func TestFillMessages_KeepsInputOrder(t *testing.T) {
	entries := []model.Entry{
		{Term: "un", Translations: []model.Translation{{Text: "one"}}},
		{Term: "deux", Translations: []model.Translation{{Text: "two"}}},
	}
	_, user := FillMessages(entries, DeckContext{})
	idx1 := strings.Index(user, "un\n")
	idx2 := strings.Index(user, "deux\n")
	if idx1 == -1 || idx2 == -1 || idx1 > idx2 {
		t.Errorf("input entries should appear in order in the message")
	}
}