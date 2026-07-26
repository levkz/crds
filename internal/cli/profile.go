package cli

import (
	"fmt"

	"crds/internal/app"
	"crds/internal/config"
)

type ProfileCmd struct {
	Export ProfileExportCmd `cmd:"" help:"Export full profile for device migration."`
	Import ProfileImportCmd `cmd:"" help:"Import a profile from another device."`
}

type ProfileExportCmd struct {
	Output string `short:"o" help:"Output directory (default: current directory)."`
	Name   string `short:"n" help:"Archive name (.tar.gz auto-appended)."`
}

func (c *ProfileExportCmd) Run(a *app.App) error {
	configDir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("profile export: config dir: %w", err)
	}

	outputDir := c.Output
	if outputDir == "" {
		outputDir = "."
	}

	path, err := a.Store.CreateProfile(a.SharedDir, configDir, outputDir, c.Name)
	if err != nil {
		return fmt.Errorf("profile export: %w", err)
	}
	fmt.Printf("Profile exported to %q.\n", path)
	return nil
}

type ProfileImportCmd struct {
	File string `arg:"" required:"" help:"Path to profile archive."`
}

func (c *ProfileImportCmd) Run(a *app.App) error {
	configDir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("profile import: config dir: %w", err)
	}

	if err := a.Store.ImportProfile(a.SharedDir, configDir, c.File); err != nil {
		return fmt.Errorf("profile import: %w", err)
	}
	fmt.Printf("Profile imported from %q.\n", c.File)
	return nil
}
