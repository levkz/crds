package cli

import (
	"fmt"
	"strings"

	"crds/internal/app"
	"crds/internal/ui"
)

type SearchCmd struct {
	Query string `arg:"" required:"" help:"Search query."`
}

func (c *SearchCmd) Run(a *app.App) error {
	deckIDs, err := a.Store.ListDecks()
	if err != nil {
		return fmt.Errorf("list decks: %w", err)
	}

	query := strings.ToLower(c.Query)
	var results []struct {
		deck string
		card ui.CardData
	}

	for _, id := range deckIDs {
		deck, err := a.Store.LoadDeck(id)
		if err != nil {
			continue
		}
		for _, card := range deck.Cards {
			if matches(card, query) {
				results = append(results, struct {
					deck string
					card ui.CardData
				}{deck: deck.Name, card: card})
			}
		}
	}

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	fmt.Printf("%d match(es):\n\n", len(results))
	for _, r := range results {
		fmt.Printf("  [%s] %s\n", r.deck, r.card.Front)
		for _, b := range r.card.Back {
			fmt.Printf("         %s\n", b)
		}
		if r.card.Notes != "" {
			fmt.Printf("         notes: %s\n", r.card.Notes)
		}
		fmt.Println()
	}

	return nil
}

func matches(card ui.CardData, query string) bool {
	if strings.Contains(strings.ToLower(card.Front), query) {
		return true
	}
	if strings.Contains(strings.ToLower(card.Notes), query) {
		return true
	}
	for _, b := range card.Back {
		if strings.Contains(strings.ToLower(b), query) {
			return true
		}
	}
	for _, v := range card.Variants {
		if strings.Contains(strings.ToLower(v), query) {
			return true
		}
	}
	return false
}
