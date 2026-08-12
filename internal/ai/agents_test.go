package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"crds/internal/model"
)

func TestInterpret_UsesClient(t *testing.T) {
	fake := &fakeClient{}
	fake.reply = "- term: bonjour\n  translations:\n    - text: hello\n"

	entries, err := Interpret(context.Background(), fake, "bonjour = hello", LanguageContext{}, "")
	if err != nil {
		t.Fatalf("Interpret: %v", err)
	}
	if len(entries) != 1 || entries[0].Term != "bonjour" {
		t.Fatalf("entries = %+v", entries)
	}
	if !strings.Contains(fake.user, "bonjour") {
		t.Errorf("client should receive the raw input")
	}
}

func TestInterpret_MsgPassthrough(t *testing.T) {
	fake := &fakeClient{}
	fake.reply = "- term: bonjour\n  translations:\n    - text: hello\n"

	if _, err := Interpret(context.Background(), fake, "bonjour", LanguageContext{}, "use formal register"); err != nil {
		t.Fatalf("Interpret: %v", err)
	}
	if !strings.Contains(fake.user, "use formal register") {
		t.Errorf("client should receive the extra instruction, got:\n%s", fake.user)
	}
}

func TestInterpretFull_UsesClient(t *testing.T) {
	fake := &fakeClient{}
	fake.reply = "- term: hola\n  translations:\n    - text: hello\n  examples:\n    - text: Hola, ¿cómo estás?\n      translation: Hello, how are you?\n"

	entries, err := InterpretFull(context.Background(), fake, "hola", DeckContext{}, "")
	if err != nil {
		t.Fatalf("InterpretFull: %v", err)
	}
	if len(entries) != 1 || entries[0].Term != "hola" {
		t.Fatalf("entries = %+v", entries)
	}
	if !strings.Contains(fake.sys, "at least 4") {
		t.Errorf("full system message should demand 4 examples, got:\n%s", fake.sys)
	}
}

func TestFill_UsesClientAndTagContext(t *testing.T) {
	fake := &fakeClient{}
	fake.reply = "- term: (un/le) chat\n  translations:\n    - text: (to) cat\n"

	entries, err := Fill(context.Background(), fake,
		[]model.Entry{{Term: "chat", Translations: []model.Translation{{Text: "cat"}}}},
		DeckContext{AllowedTags: []string{"animal"}}, "")
	if err != nil {
		t.Fatalf("Fill: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %+v", entries)
	}
	if !strings.Contains(fake.sys, "(un/le)") {
		t.Errorf("system message should carry conventions")
	}
}

func TestInterpret_ClientError(t *testing.T) {
	fake := &fakeClient{err: errors.New("boom")}
	if _, err := Interpret(context.Background(), fake, "x", LanguageContext{}, ""); err == nil {
		t.Fatal("expected error propagation")
	}
}

func TestSuggestDeck_ReturnsMatch(t *testing.T) {
	fake := &fakeClient{reply: `{"deck": "spanish"}`}
	decks := []DeckInfo{
		{ID: "spanish", Name: "Spanish Basics", Language: "es", TranslationLanguage: "en"},
		{ID: "french_a1", Name: "French A1", Language: "fr", TranslationLanguage: "en"},
	}

	res, err := SuggestDeck(context.Background(), fake, decks, "hola\ncómo estás", "")
	if err != nil {
		t.Fatalf("SuggestDeck: %v", err)
	}
	if res.Deck != "spanish" || res.Proposed != nil {
		t.Fatalf("result = %+v, want deck spanish", res)
	}
	if !strings.Contains(fake.user, "spanish") || !strings.Contains(fake.user, "es") {
		t.Errorf("client should receive the deck list, got:\n%s", fake.user)
	}
	if !strings.Contains(fake.user, "hola") {
		t.Errorf("client should receive the raw input")
	}
}

func TestSuggestDeck_UnknownIDIgnored(t *testing.T) {
	fake := &fakeClient{reply: `{"deck": "hallucinated"}`}
	decks := []DeckInfo{{ID: "spanish", Name: "Spanish", Language: "es", TranslationLanguage: "en"}}

	res, err := SuggestDeck(context.Background(), fake, decks, "hola", "")
	if err != nil {
		t.Fatalf("SuggestDeck: %v", err)
	}
	if res.Deck != "" {
		t.Fatalf("result = %+v, want no deck", res)
	}
}

func TestSuggestDeck_ReturnsProposal(t *testing.T) {
	fake := &fakeClient{reply: `{"deck": null, "proposed": {"name": "French Basics", "from": "fr", "to": "en"}}`}

	res, err := SuggestDeck(context.Background(), fake, nil, "bonjour", "")
	if err != nil {
		t.Fatalf("SuggestDeck: %v", err)
	}
	if res.Deck != "" || res.Proposed == nil {
		t.Fatalf("result = %+v, want a proposal", res)
	}
	if res.Proposed.Name != "French Basics" || res.Proposed.Language != "fr" || res.Proposed.TranslationLanguage != "en" {
		t.Fatalf("proposal = %+v", res.Proposed)
	}
}

func TestSuggestDeck_GarbageDegradesToNoMatch(t *testing.T) {
	fake := &fakeClient{reply: "definitely not json"}
	decks := []DeckInfo{{ID: "spanish", Name: "Spanish", Language: "es", TranslationLanguage: "en"}}

	res, err := SuggestDeck(context.Background(), fake, decks, "hola", "")
	if err != nil {
		t.Fatalf("SuggestDeck: %v", err)
	}
	if res.Deck != "" || res.Proposed != nil {
		t.Fatalf("result = %+v, want empty result", res)
	}
}

func TestSuggestDeck_ClientError(t *testing.T) {
	fake := &fakeClient{err: errors.New("boom")}
	if _, err := SuggestDeck(context.Background(), fake, nil, "hola", ""); err == nil {
		t.Fatal("expected error propagation")
	}
}

func TestInterpret_ParseError(t *testing.T) {
	fake := &fakeClient{reply: "not yaml at all"}
	if _, err := Interpret(context.Background(), fake, "x", LanguageContext{}, ""); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestIsStructuredInput(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"- term: bonjour\n  translations:\n", true},
		{"- id: e1\n  term: bonjour\n", true},
		{"term: bonjour\n", true},
		{"id: e1\n", true},
		{"entries:\n  - term: bonjour\n", true},
		{"bonjour\nmanger\nla maison", false},
		{"", false},
		{"  \n  ", false},
	}
	for _, c := range cases {
		if got := IsStructuredInput(c.input); got != c.want {
			t.Errorf("IsStructuredInput(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}