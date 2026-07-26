package cli

import (
	"fmt"

	"crds/internal/app"
	"crds/internal/storage"
)

type RevertCmd struct {
	Latest bool   `help:"Revert to the latest reserve copy in the default location."`
	File   string `short:"f" help:"Path to a reserve archive." completion-predictor:"reserve"`
}

func (c *RevertCmd) Run(a *app.App) error {
	var path string
	switch {
	case c.File != "":
		path = c.File
	case c.Latest:
		reserves, err := storage.ListReserves(a.SharedDir)
		if err != nil {
			return fmt.Errorf("list reserves: %w", err)
		}
		if len(reserves) == 0 {
			return fmt.Errorf("no reserve copies found in default location")
		}
		path = reserves[0]
	default:
		return fmt.Errorf("specify --file or --latest")
	}

	if err := a.Store.RevertReserve(a.SharedDir, path); err != nil {
		return fmt.Errorf("revert: %w", err)
	}
	fmt.Printf("Reverted from %q.\n", path)
	return nil
}
