package cli

import (
	"fmt"

	"crds/internal/app"
)

type SyncCmd struct {
	Path string `arg:"" optional:"" default:"decks"`

	Write bool `short:"w" help:"Write generated IDs back to YAML files."`
}

func (s *SyncCmd) Run(a *app.App) error {
	if err := a.Store.SyncDecks(a.DataDir); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	fmt.Println("Decks synchronized.")

	if s.Write {
		// TODO: write auto-generated IDs back to YAML files.
		// Requires round-trip YAML (preserving comments) or re-marshalling
		// after `assignIDs()` fills empty IDs.
		fmt.Println("Write flag set — ID backfill not yet implemented.")
	}

	return nil
}
