package ai

import (
	"fmt"
	"strings"

	"crds/internal/model"

	"go.yaml.in/yaml/v3"
)

// ParseEntries converts a model reply into validated entries. It tolerates a
// single surrounding markdown fence and accepts a bare entry or a deck-shaped
// document as well as a plain list.
func ParseEntries(output string) ([]model.Entry, error) {
	yamlText := extractYAML(output)
	if strings.TrimSpace(yamlText) == "" {
		return nil, fmt.Errorf("ai: empty model reply")
	}

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
	return nil
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