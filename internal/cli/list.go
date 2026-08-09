package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"crds/internal/app"
)

type ListCmd struct {
}

func (c *ListCmd) Run(a *app.App) error {
	decks, err := a.Store.ListDecksWithStats()
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	if len(decks) == 0 {
		fmt.Println("No decks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "DECK ID\tNAME\tENTRIES\tLANGUAGE\tTRANSLATION LANGUAGE"); err != nil {
		return err
	}
	for _, d := range decks {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			d.ID, d.Name, d.EntryCount, d.Language, d.TranslationLanguage); err != nil {
			return err
		}
	}
	return w.Flush()
}
