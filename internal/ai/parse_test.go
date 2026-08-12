package ai

import (
	"testing"
)

func TestParseEntries_PlainList(t *testing.T) {
	out := `- term: bonjour
  translations:
    - text: hello
- term: matin
  translations:
    - text: morning
`
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].Term != "bonjour" || entries[0].Translations[0].Text != "hello" {
		t.Errorf("entry 0 = %+v", entries[0])
	}
}

func TestParseEntries_FencedYAML(t *testing.T) {
	out := "```yaml\n- term: bonjour\n  translations:\n    - text: hello\n```\n"
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 1 || entries[0].Term != "bonjour" {
		t.Fatalf("expected 1 entry, got %+v", entries)
	}
}

func TestParseEntries_WithProseAfterFence(t *testing.T) {
	out := "Here are the entries:\n\n```\n- term: bonjour\n  translations:\n    - text: hello\n```\n\nLet me know if you need more."
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestParseEntries_SingleEntry(t *testing.T) {
	out := "term: bonjour\ntranslations:\n  - text: hello\n"
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 1 || entries[0].Term != "bonjour" {
		t.Fatalf("expected single entry, got %+v", entries)
	}
}

func TestParseEntries_DeckShaped(t *testing.T) {
	out := "id: deck\nname: Deck\nlanguage: fr\nentries:\n  - term: bonjour\n    translations:\n      - text: hello\n"
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 1 || entries[0].Term != "bonjour" {
		t.Fatalf("expected entry from deck doc, got %+v", entries)
	}
}

func TestParseEntries_Garbage(t *testing.T) {
	if _, err := ParseEntries("this is not yaml: [ : ]"); err == nil {
		t.Fatal("expected error for garbage input")
	}
}

func TestParseEntries_EmptyReply(t *testing.T) {
	if _, err := ParseEntries("   "); err == nil {
		t.Fatal("expected error for empty reply")
	}
}

func TestParseEntries_MissingTerm(t *testing.T) {
	out := "- translations:\n    - text: hello\n"
	if _, err := ParseEntries(out); err == nil {
		t.Fatal("expected error for missing term")
	}
}

func TestParseEntries_MissingTranslation(t *testing.T) {
	out := "- term: bonjour\n"
	if _, err := ParseEntries(out); err == nil {
		t.Fatal("expected error for missing translation")
	}
}

func TestParseEntries_TrimsWhitespace(t *testing.T) {
	out := "- term:  bonjour  \n  translations:\n    - text:  hello \n  notes:  a note \n"
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if entries[0].Term != "bonjour" {
		t.Errorf("term = %q, want trimmed", entries[0].Term)
	}
	if entries[0].Translations[0].Text != "hello" {
		t.Errorf("translation = %q, want trimmed", entries[0].Translations[0].Text)
	}
	if entries[0].Notes != "a note" {
		t.Errorf("notes = %q, want trimmed", entries[0].Notes)
	}
}

func TestParseEntries_ExampleStruct(t *testing.T) {
	out := "- term: bonjour\n  translations:\n    - text: hello\n  examples:\n    - text: Bonjour Marie.\n      translation: Hello Marie.\n"
	entries, err := ParseEntries(out)
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries[0].Examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(entries[0].Examples))
	}
	if entries[0].Examples[0].Text != "Bonjour Marie." {
		t.Errorf("example text = %q", entries[0].Examples[0].Text)
	}
	if entries[0].Examples[0].Translation != "Hello Marie." {
		t.Errorf("example translation = %q", entries[0].Examples[0].Translation)
	}
}

func TestParseEntries_ExampleEmptyText(t *testing.T) {
	out := "- term: hola\n  translations:\n    - text: hello\n  examples:\n    - text: \"\"\n      translation: Hello.\n"
	if _, err := ParseEntries(out); err == nil {
		t.Fatal("expected error for example with empty source sentence")
	}
}

func TestParseEntries_ExampleEmptyTranslation(t *testing.T) {
	out := "- term: hola\n  translations:\n    - text: hello\n  examples:\n    - text: Hola.\n      translation: \"\"\n"
	if _, err := ParseEntries(out); err == nil {
		t.Fatal("expected error for example with empty translation")
	}
}

func TestParseSuggestion_Match(t *testing.T) {
	res := ParseSuggestion(`{"deck": "spanish"}`, []string{"spanish", "french_a1"})
	if res.Deck != "spanish" || res.Proposed != nil {
		t.Fatalf("result = %+v, want deck spanish", res)
	}
}

func TestParseSuggestion_UnknownID(t *testing.T) {
	res := ParseSuggestion(`{"deck": "not_a_deck"}`, []string{"spanish"})
	if res.Deck != "" || res.Proposed != nil {
		t.Fatalf("result = %+v, want no match", res)
	}
}

func TestParseSuggestion_WrappedJSON(t *testing.T) {
	res := ParseSuggestion("```json\n{\"deck\": \"spanish\"}\n```", []string{"spanish"})
	if res.Deck != "spanish" {
		t.Fatalf("result = %+v, want wrapped json tolerated", res)
	}
}

func TestParseSuggestion_Proposal(t *testing.T) {
	res := ParseSuggestion(`{"deck": null, "proposed": {"name": "French Basics", "from": "fr", "to": "en"}}`, nil)
	if res.Deck != "" || res.Proposed == nil {
		t.Fatalf("result = %+v, want a proposal", res)
	}
	if res.Proposed.Name != "French Basics" || res.Proposed.Language != "fr" || res.Proposed.TranslationLanguage != "en" {
		t.Fatalf("proposal = %+v", res.Proposed)
	}
}

func TestParseSuggestion_ProposalEmptyName(t *testing.T) {
	res := ParseSuggestion(`{"deck": null, "proposed": {"name": "", "from": "fr", "to": "en"}}`, nil)
	if res.Deck != "" || res.Proposed != nil {
		t.Fatalf("result = %+v, want no proposal with empty name", res)
	}
}

func TestParseSuggestion_EmptyProposal(t *testing.T) {
	res := ParseSuggestion(`{"deck": null, "proposed": null}`, nil)
	if res.Deck != "" || res.Proposed != nil {
		t.Fatalf("result = %+v, want empty result", res)
	}
}

func TestParseSuggestion_Garbage(t *testing.T) {
	for _, out := range []string{"", "not json", "I think the spanish deck fits best."} {
		res := ParseSuggestion(out, []string{"spanish"})
		if res.Deck != "" || res.Proposed != nil {
			t.Fatalf("garbage %q: result = %+v, want empty result", out, res)
		}
	}
}

func TestParseEntries_KindListFromEmptyModel(t *testing.T) {
	// A model that echoes headers but nothing else must still error cleanly.
	_, err := ParseEntries("entries:")
	if err == nil {
		t.Fatal("expected error for empty entries key")
	}
}