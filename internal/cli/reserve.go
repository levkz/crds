package cli

import (
	"fmt"

	"crds/internal/app"
	"crds/internal/config"
)

type ReserveCmd struct {
	Output string `short:"o" help:"Output directory (default: ~/.local/share/crds/reserve-copies/)."`
	Name   string `short:"n" help:"Archive name (.tar.gz auto-appended)."`
}

func (c *ReserveCmd) Run(a *app.App) error {
	configDir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("reserve: config dir: %w", err)
	}

	path, err := a.Store.CreateReserveTo(a.SharedDir, configDir, c.Output, c.Name)
	if err != nil {
		return fmt.Errorf("reserve: %w", err)
	}
	fmt.Printf("Reserve copy created at %q.\n", path)
	return nil
}
