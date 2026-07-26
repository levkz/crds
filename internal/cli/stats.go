package cli

import (
	"fmt"

	"crds/internal/app"
)

type StatsCmd struct {
	Deck string `arg:"" optional:"" help:"Show stats for a single deck." completion-predictor:"deck"`
}

func (c *StatsCmd) Run(a *app.App) error {
	stats := a.Store.Stats()

	fmt.Println("Statistics")
	fmt.Printf("  Reviewed today:  %d\n", stats.ReviewedToday)
	fmt.Printf("  Accuracy:        %.1f%%\n", stats.Accuracy*100)
	fmt.Printf("  Total cards:     %d\n", stats.TotalCards)

	return nil
}
