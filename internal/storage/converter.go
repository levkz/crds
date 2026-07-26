package storage

import (
	"crds/internal/model"
	"crds/internal/ui"
)

func expandVariants(translations []model.Translation) []string {
	var variants []string
	for _, t := range translations {
		variants = append(variants, model.ExpandText(t.Text)...)
	}
	return variants
}

func entryToCardData(entry model.Entry) ui.CardData {
	back := make([]string, len(entry.Translations))
	var variants []string
	for i, t := range entry.Translations {
		back[i] = t.Text
	}
	variants = expandVariants(entry.Translations)

	if len(variants) == 0 {
		variants = back
	}

	return ui.CardData{
		ID:       entry.ID,
		Front:    entry.Term,
		Back:     back,
		Variants: variants,
		Notes:    entry.Notes,
	}
}
