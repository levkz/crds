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
		if err := a.Store.WriteBackIDs(a.DataDir); err != nil {
			return fmt.Errorf("write-back IDs: %w", err)
		}
		fmt.Println("Auto-generated IDs written to YAML files.")
	}

	return nil
}
