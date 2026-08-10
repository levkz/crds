package ai

import (
	"strings"
	"testing"

	"crds/internal/model"
)

func TestInterpretMessages_IncludesRawAndLanguages(t *testing.T) {
	system, user := InterpretMessages("manger\nboire", LanguageContext{Language: "fr", TranslationLanguage: "en"}, "")

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
	_, user := InterpretMessages("just a word", LanguageContext{}, "")
	if strings.Contains(user, "Source language") {
		t.Errorf("without a deck, languages should be omitted")
	}
}

func TestInterpretMessages_IncludesMsg(t *testing.T) {
	_, user := InterpretMessages("hola", LanguageContext{}, "use formal register")
	if !strings.Contains(user, "use formal register") {
		t.Errorf("extra instruction should be appended to the user message, got:\n%s", user)
	}
}

func TestInterpretFullMessages_DemandsFourExamples(t *testing.T) {
	system, user := InterpretFullMessages("hola", DeckContext{Language: "es", TranslationLanguage: "en"}, "")
	if !strings.Contains(system, "at least 4") {
		t.Errorf("system message should demand at least 4 examples, got:\n%s", system)
	}
	if !strings.Contains(system, "text: <a complete, natural sentence in the SOURCE language>") {
		t.Errorf("system message should spell out the example schema, got:\n%s", system)
	}
	if !strings.Contains(system, "Both text and translation must be non-empty") {
		t.Errorf("system message should require both example fields, got:\n%s", system)
	}
	if !strings.Contains(system, "Hola, como estas?") {
		t.Errorf("system message should include a correctly formed example, got:\n%s", system)
	}
	if !strings.Contains(system, "(une/la)") || !strings.Contains(system, "(to)") {
		t.Errorf("system message should carry convention rules")
	}
	if !strings.Contains(user, "hola") {
		t.Errorf("user message should include the raw input")
	}
}

func TestInterpretFullMessages_TagsAllowlist(t *testing.T) {
	_, user := InterpretFullMessages("hola", DeckContext{AllowedTags: []string{"greetings", "A1"}}, "")
	if !strings.Contains(user, "allowed theme tags: greetings, A1") {
		t.Errorf("user message should list allowed theme tags, got:\n%s", user)
	}
}

func TestInterpretFullMessages_IncludesMsg(t *testing.T) {
	_, user := InterpretFullMessages("hola", DeckContext{}, "add IPA pronunciation")
	if !strings.Contains(user, "add IPA pronunciation") {
		t.Errorf("extra instruction should be appended, got:\n%s", user)
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
	}, "")

	if !strings.Contains(user, "allowed theme tags: food, A1") {
		t.Errorf("user message should list allowed theme tags, got:\n%s", user)
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
	if !strings.Contains(system, "at least 4") {
		t.Errorf("system message should demand at least 4 examples, got:\n%s", system)
	}
	if !strings.Contains(system, "Both text and translation must be non-empty") {
		t.Errorf("system message should require both example fields, got:\n%s", system)
	}
}

func TestFillMessages_NoTags(t *testing.T) {
	_, user := FillMessages(nil, DeckContext{AllowedTags: nil}, "")
	if !strings.Contains(user, "add structural tags and a concise theme tag") {
		t.Errorf("empty tag allowlist should instruct structural + theme tags, got:\n%s", user)
	}
}

func TestInterpretFullMessages_StructuralTagsAlwaysAllowed(t *testing.T) {
	system, _ := InterpretFullMessages("hola", DeckContext{}, "")
	for _, want := range []string{
		"Structural tags are ALWAYS allowed",
		"noun, verb, adjective, adverb",
		"masculin, feminin, and neutral",
		"if the language has grammatical gender",
	} {
		if !strings.Contains(system, want) {
			t.Errorf("system message should contain %q, got:\n%s", want, system)
		}
	}
	if !strings.Contains(system, "add a concise theme tag") {
		t.Errorf("system message should allow a concise theme tag, got:\n%s", system)
	}
}

func TestTagRules_ProficiencyLevel(t *testing.T) {
	fullSys, _ := InterpretFullMessages("hola", DeckContext{}, "")
	if !strings.Contains(fullSys, "add exactly one CEFR level tag") ||
		!strings.Contains(fullSys, "A1, A2, B1, B2, C1, C2") {
		t.Errorf("full prompt should require a proficiency tag, got:\n%s", fullSys)
	}

	fillSys, _ := FillMessages(nil, DeckContext{}, "")
	if !strings.Contains(fillSys, "add exactly one CEFR level tag") {
		t.Errorf("fill prompt should require a proficiency tag, got:\n%s", fillSys)
	}
}

func TestFillMessages_StructuralTagsAlwaysAllowed(t *testing.T) {
	system, _ := FillMessages(nil, DeckContext{}, "")
	if !strings.Contains(system, "Structural tags are ALWAYS allowed") {
		t.Errorf("system message should allow structural tags, got:\n%s", system)
	}
}

func TestFillMessages_IncludesSamples(t *testing.T) {
	samples := []model.Entry{
		{Term: "(un/le) chat", Translations: []model.Translation{{Text: "(to) cat"}}},
	}
	_, user := FillMessages(nil, DeckContext{AllowedTags: []string{"animal"}, Samples: samples}, "")

	if !strings.Contains(user, "(un/le) chat") {
		t.Errorf("user message should include sample entries")
	}
}

func TestFillMessages_KeepsInputOrder(t *testing.T) {
	entries := []model.Entry{
		{Term: "un", Translations: []model.Translation{{Text: "one"}}},
		{Term: "deux", Translations: []model.Translation{{Text: "two"}}},
	}
	_, user := FillMessages(entries, DeckContext{}, "")
	idx1 := strings.Index(user, "un\n")
	idx2 := strings.Index(user, "deux\n")
	if idx1 == -1 || idx2 == -1 || idx1 > idx2 {
		t.Errorf("input entries should appear in order in the message")
	}
}

func TestFillMessages_IncludesMsg(t *testing.T) {
	entries := []model.Entry{
		{Term: "baguette", Translations: []model.Translation{{Text: "bread"}}},
	}
	_, user := FillMessages(entries, DeckContext{}, "emphasize informal register")
	if !strings.Contains(user, "emphasize informal register") {
		t.Errorf("extra instruction should be appended, got:\n%s", user)
	}
}