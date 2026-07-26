package cli

import (
	"fmt"

	"crds/internal/app"
)

type StatsCmd struct {
	Deck string `help:"Show stats for a single deck." completion-predictor:"deck"`
	Tag  string `help:"Show stats filtered by tag."`
}

func (c *StatsCmd) Run(a *app.App) error {
	if c.Deck != "" {
		ds, err := a.Store.DeckStats(c.Deck)
		if err != nil {
			return fmt.Errorf("stats: %w", err)
		}
		fmt.Printf("Statistics for deck %q:\n", c.Deck)
		fmt.Printf("  Total entries:   %d\n", ds.TotalEntries)
		fmt.Printf("  Reviewed today:  %d\n", ds.ReviewedToday)
		fmt.Printf("  Accuracy:        %.1f%%\n", ds.Accuracy)
		return nil
	}

	stats := a.Store.Stats()
	fmt.Println("Statistics")
	fmt.Printf("  Reviewed today:  %d\n", stats.ReviewedToday)
	fmt.Printf("  Accuracy:        %.1f%%\n", stats.Accuracy)
	fmt.Printf("  Total cards:     %d\n", stats.TotalCards)

	return nil
}
