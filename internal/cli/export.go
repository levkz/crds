package cli

import (
	"fmt"

	"crds/internal/app"
)

type ExportCmd struct {
	Deck   string `arg:"" required:"" help:"Deck to export." completion-predictor:"deck"`
	Output string `short:"o" help:"Destination path (default: <deck>.yaml)."`
}

func (c *ExportCmd) Run(a *app.App) error {
	dst := c.Output
	if dst == "" {
		dst = c.Deck + ".yaml"
	}
	if err := a.Store.ExportDeck(c.Deck, dst, a.DataDir); err != nil {
		return fmt.Errorf("export: %w", err)
	}
	fmt.Printf("Exported %q to %q.\n", c.Deck, dst)
	return nil
}
