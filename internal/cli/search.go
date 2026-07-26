package cli

import (
	"fmt"
	"strings"

	"crds/internal/app"
	"crds/internal/ui"
)

type SearchCmd struct {
	Deck  string `arg:"" required:"" help:"Deck to search." completion-predictor:"deck"`
	Query string `arg:"" required:"" help:"Search query."`
}

func (c *SearchCmd) Run(a *app.App) error {
	query := strings.ToLower(c.Query)

	deck, err := a.Store.LoadDeck(c.Deck)
	if err != nil {
		return fmt.Errorf("load deck %q: %w", c.Deck, err)
	}

	var results []ui.CardData
	for _, card := range deck.Cards {
		if matches(card, query) {
			results = append(results, card)
		}
	}

	if len(results) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	fmt.Printf("%d match(es) in %q:\n\n", len(results), c.Deck)
	for _, r := range results {
		fmt.Printf("  %s\n", r.Front)
		for _, b := range r.Back {
			fmt.Printf("         %s\n", b)
		}
		if r.Notes != "" {
			fmt.Printf("         notes: %s\n", r.Notes)
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
