package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"crds/internal/app"
)

type ExportCmd struct {
	Deck   string `arg:"" optional:"" help:"Deck to export." completion-predictor:"deck"`
	All    bool   `help:"Export all decks."`
	Output string `short:"o" help:"Destination path (default: <deck>.yaml) or directory (with --all)."`
}

func (c *ExportCmd) Run(a *app.App) error {
	if c.All {
		return c.exportAll(a)
	}
	if c.Deck == "" {
		return fmt.Errorf("export: specify a deck or use --all")
	}

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

func (c *ExportCmd) exportAll(a *app.App) error {
	decks, err := a.Store.ListDecksWithStats()
	if err != nil {
		return fmt.Errorf("export: list decks: %w", err)
	}

	outDir := c.Output
	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("export: create dir %s: %w", outDir, err)
	}

	var exported int
	for _, d := range decks {
		dst := filepath.Join(outDir, d.ID+".yaml")
		if err := a.Store.ExportDeck(d.ID, dst, a.DataDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: export %s: %v\n", d.ID, err)
			continue
		}
		exported++
	}
	fmt.Printf("Exported %d deck(s) to %q.\n", exported, outDir)
	return nil
}
