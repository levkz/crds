package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"crds/internal/model"

	"go.yaml.in/yaml/v3"
)

// ParseEntries converts a model reply into validated entries. It tolerates a
// single surrounding markdown fence and accepts a bare entry or a deck-shaped
// document as well as a plain list. Every parse or validation failure is
// wrapped with a snippet of the model's reply so the error is actionable.
func ParseEntries(output string) ([]model.Entry, error) {
	yamlText := extractYAML(output)
	if strings.TrimSpace(yamlText) == "" {
		return nil, fmt.Errorf("ai: empty model reply")
	}

	entries, err := parseEntries(yamlText)
	if err != nil {
		return nil, fmt.Errorf("%s — model returned: %s", err, snippet(yamlText))
	}
	return entries, nil
}

func parseEntries(yamlText string) ([]model.Entry, error) {
	var entries []model.Entry
	if err := yaml.Unmarshal([]byte(yamlText), &entries); err != nil {
		// Fall back to a single entry or a deck document.
		var single model.Entry
		if err2 := yaml.Unmarshal([]byte(yamlText), &single); err2 == nil && strings.TrimSpace(single.Term) != "" {
			entries = []model.Entry{single}
		} else {
			var deck model.Deck
			if err3 := yaml.Unmarshal([]byte(yamlText), &deck); err3 == nil && len(deck.Entries) > 0 {
				entries = deck.Entries
			} else {
				return nil, fmt.Errorf("ai: parse model output: %w", err)
			}
		}
	}

	for i := range entries {
		entry := &entries[i]
		if err := validateEntry(entry); err != nil {
			return nil, err
		}
	}

	return entries, nil
}

// snippet renders a short single-line preview of a model reply for error text.
func snippet(s string) string {
	const max = 140
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max] + "…"
	}
	return fmt.Sprintf("%q", s)
}

func validateEntry(entry *model.Entry) error {
	entry.Term = strings.TrimSpace(entry.Term)
	entry.Notes = strings.TrimSpace(entry.Notes)
	for j := range entry.Translations {
		entry.Translations[j].Text = strings.TrimSpace(entry.Translations[j].Text)
	}
	for j := range entry.Examples {
		entry.Examples[j].Text = strings.TrimSpace(entry.Examples[j].Text)
		entry.Examples[j].Translation = strings.TrimSpace(entry.Examples[j].Translation)
	}
	for j := range entry.Tags {
		entry.Tags[j] = strings.TrimSpace(entry.Tags[j])
	}
	if len(entry.Examples) == 0 {
		entry.Examples = nil
	}
	if len(entry.Tags) == 0 {
		entry.Tags = nil
	}

	if entry.Term == "" {
		return fmt.Errorf("ai: entry has an empty term")
	}
	if len(entry.Translations) == 0 {
		return fmt.Errorf("ai: entry %q has no translations", entry.Term)
	}
	for _, t := range entry.Translations {
		if t.Text == "" {
			return fmt.Errorf("ai: entry %q has an empty translation", entry.Term)
		}
	}
	for _, ex := range entry.Examples {
		if ex.Text == "" {
			return fmt.Errorf("ai: entry %q has an example with an empty source sentence", entry.Term)
		}
		if ex.Translation == "" {
			return fmt.Errorf("ai: entry %q has an example with an empty translation", entry.Term)
		}
	}
	return nil
}

// SuggestResult is the parsed deck-suggestion reply. Deck holds a matched
// existing deck id; when empty, Proposed carries a new-deck proposal (nil when
// the model had nothing sensible to propose).
type SuggestResult struct {
	Deck     string
	Proposed *DeckProposal
}

// DeckProposal is a model-proposed new deck (name + language pair).
type DeckProposal struct {
	Name                string `json:"name"`
	Language            string `json:"from"`
	TranslationLanguage string `json:"to"`
}

// suggestionReply mirrors the strict JSON contract demanded by the prompt.
type suggestionReply struct {
	Deck     string       `json:"deck"`
	Proposed *DeckProposal `json:"proposed"`
}

// ParseSuggestion converts a model reply into a SuggestResult. A deck id that
// is not among knownIDs is treated as no match (the model must not invent
// ids). A malformed reply yields an empty result, never an error: a wrong
// deck guess is a UX prompt, not a data-write risk.
func ParseSuggestion(output string, knownIDs []string) SuggestResult {
	var reply suggestionReply
	if err := json.Unmarshal([]byte(extractYAML(output)), &reply); err != nil {
		return SuggestResult{}
	}
	if reply.Deck != "" {
		for _, id := range knownIDs {
			if id == reply.Deck {
				return SuggestResult{Deck: reply.Deck}
			}
		}
		return SuggestResult{}
	}
	if reply.Proposed != nil {
		reply.Proposed.Name = strings.TrimSpace(reply.Proposed.Name)
		reply.Proposed.Language = strings.TrimSpace(reply.Proposed.Language)
		reply.Proposed.TranslationLanguage = strings.TrimSpace(reply.Proposed.TranslationLanguage)
		if reply.Proposed.Name == "" {
			return SuggestResult{}
		}
		return SuggestResult{Proposed: reply.Proposed}
	}
	return SuggestResult{}
}

// extractYAML strips a ``` ... ``` fence if the reply contains one.
func extractYAML(output string) string {
	out := strings.TrimSpace(output)

	open := strings.Index(out, "```")
	if open == -1 {
		return out
	}
	rest := out[open+3:]
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[nl+1:]
	} else {
		return ""
	}
	if close := strings.LastIndex(rest, "```"); close >= 0 {
		rest = rest[:close]
	}
	return strings.TrimSpace(rest)
}