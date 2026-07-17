package cli

import (
	"fmt"

	"crds/internal/app"
)

type SyncCmd struct {
	Path string `arg:"" optional:"" default:"decks"`

	Write bool `short:"w" help:"Write generated IDs."`
}

func (s *SyncCmd) Run(a *app.App) error {
	fmt.Printf("Sync path=%q write=%v\n",
		s.Path,
		s.Write,
	)

	return nil
}
