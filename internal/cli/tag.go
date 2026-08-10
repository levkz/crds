package cli

import (
	"fmt"
	"sort"

	"crds/internal/app"
)

type TagCmd struct {
	Add  TagAddCmd  `cmd:"" help:"Add tags to a term."`
	Rm   TagRmCmd   `cmd:"" help:"Remove tags from a term."`
	List TagListCmd `cmd:"" help:"List tags on a term."`
}

type TagAddCmd struct {
	Deck   string   `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string   `arg:"" required:"" help:"Term ID." completion-predictor:"term"`
	Tags   []string `arg:"" optional:"" help:"Tags to add."`
}

func (c *TagAddCmd) Run(a *app.App) error {
	err := a.Store.AddTagsToEntry(c.Deck, c.TermID, c.Tags, a.DataDir)
	if err != nil {
		return err
	}
	fmt.Printf("Added %d tag(s) to entry %q in deck %q.\n", len(c.Tags), c.TermID, c.Deck)
	return nil
}

type TagRmCmd struct {
	Deck   string   `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string   `arg:"" required:"" help:"Term ID." completion-predictor:"term"`
	Tags   []string `arg:"" optional:"" help:"Tags to remove."`
}

func (c *TagRmCmd) Run(a *app.App) error {
	err := a.Store.RemoveTagsFromEntry(c.Deck, c.TermID, c.Tags, a.DataDir)
	if err != nil {
		return err
	}
	fmt.Printf("Removed %d tag(s) from entry %q in deck %q.\n", len(c.Tags), c.TermID, c.Deck)
	return nil
}

type TagListCmd struct {
	Deck   string `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string `arg:"" optional:"" help:"Term ID (omitted to list all tags in the deck)." completion-predictor:"term"`
}

func (c *TagListCmd) Run(a *app.App) error {
	var tags []string
	var err error

	if c.TermID == "" {
		tags, err = a.Store.ListDeckTags(c.Deck)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}
	} else {
		tags, err = a.Store.GetTagsByEntry(c.TermID)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}
	}

	if len(tags) == 0 {
		if c.TermID == "" {
			fmt.Printf("No tags in deck %q.\n", c.Deck)
		} else {
			fmt.Printf("No tags on entry %q in deck %q.\n", c.TermID, c.Deck)
		}
		return nil
	}
	sort.Strings(tags)
	for _, t := range tags {
		fmt.Println(t)
	}
	return nil
}
