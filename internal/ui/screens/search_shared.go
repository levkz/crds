package screens

import (
	"strings"

	"crds/internal/ui"
)

// filterCards returns search entries whose front or any translation contains
// the query (case-insensitive substring). Used by SearchModel and the
// Statistics per-word tab.
func filterCards(cards []ui.CardData, query string) []searchEntry {
	if query == "" || len(cards) == 0 {
		return nil
	}
	q := strings.ToLower(query)
	var results []searchEntry
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card.Front), q) {
			results = append(results, searchEntry{
				ID:       card.ID,
				front:    card.Front,
				back:     card.Back,
				notes:    card.Notes,
				tags:     card.Tags,
				examples: card.Examples,
			})
			continue
		}
		for _, t := range card.Back {
			if strings.Contains(strings.ToLower(t), q) {
				results = append(results, searchEntry{
					ID:       card.ID,
					front:    card.Front,
					back:     card.Back,
					notes:    card.Notes,
					tags:     card.Tags,
					examples: card.Examples,
				})
				break
			}
		}
	}
	return results
}
