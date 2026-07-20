package storage

import (
	"crds/internal/model"
	"crds/internal/ui"
)

func entryToCardData(entry model.Entry) ui.CardData {
	back := make([]string, len(entry.Translations))
	for i, t := range entry.Translations {
		back[i] = t.Text
	}

	return ui.CardData{
		ID:    entry.ID,
		Front: entry.Term,
		Back:  back,
		Notes: entry.Notes,
	}
}
