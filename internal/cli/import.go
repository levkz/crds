package cli

import (
	"fmt"

	"crds/internal/app"
)

type ImportCmd struct {
	Src string `arg:"" required:"" help:"Path to the YAML file to import."`
}

func (c *ImportCmd) Run(a *app.App) error {
	if err := a.Store.ImportDeck(c.Src, a.DataDir); err != nil {
		return fmt.Errorf("import: %w", err)
	}
	fmt.Printf("Imported %q.\n", c.Src)
	return nil
}
