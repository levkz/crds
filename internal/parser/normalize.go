package parser

import (
	"strings"

	"crds/internal/model"
)

func Normalize(deck *model.Deck) {
	deck.ID = strings.TrimSpace(deck.ID)
	deck.Name = strings.TrimSpace(deck.Name)
	deck.Language = strings.TrimSpace(deck.Language)
	deck.TranslationLanguage = strings.TrimSpace(deck.TranslationLanguage)

	for i := range deck.Entries {
		entry := &deck.Entries[i]

		entry.ID = strings.TrimSpace(entry.ID)
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
	}
}
