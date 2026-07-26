package cli

import (
	"fmt"

	"crds/internal/app"
)

type DeleteCmd struct {
	Deck  string `arg:"" required:"" help:"Deck to delete." completion-predictor:"deck"`
	Force bool   `short:"f" help:"Skip confirmation."`
}

func (c *DeleteCmd) Run(a *app.App) error {
	if !c.Force {
		fmt.Printf("Delete deck %q? [y/N] ", c.Deck)
		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := a.Store.DeleteDeck(c.Deck, a.DataDir); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	fmt.Printf("Deleted deck %q.\n", c.Deck)
	return nil
}
